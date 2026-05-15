// Package daf is the client library for Datafordeler — Klimadatastyrelsen's
// distribution platform for Danish grunddata. It targets the modern GraphQL
// stack (in prod since November 2025) at graphql.datafordeler.dk and the
// unauthenticated DAWA fallback at api.dataforsyningen.dk.
package daf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const graphqlBase = "https://graphql.datafordeler.dk"

// Client holds the API key for authenticated GraphQL calls.
//
// Datafordeler accepts the API key only as a `?apiKey=` query parameter on the
// GraphQL endpoint — Authorization: apikey/Bearer headers return 401. Every
// query also requires bitemporal arguments (registreringstid, virkningstid)
// and root queries return Connection types ({ nodes {...} }).
type Client struct {
	http   *http.Client
	apiKey string
}

// NewClientFromEnv reads DAF_API_KEY from the environment, falling
// back to the macOS Keychain entry "dafcli" / "DAF_API_KEY"
// when the env var is empty (handy when shell rc isn't propagated to the
// caller's shell).
func NewClientFromEnv() (*Client, error) {
	key := os.Getenv("DAF_API_KEY")
	if key == "" {
		if k, err := keychainLookup("dafcli", "DAF_API_KEY"); err == nil {
			key = k
		}
	}
	if key == "" {
		return nil, fmt.Errorf("DAF_API_KEY must be set (env var or macOS keychain entry dafcli/DAF_API_KEY)")
	}
	return &Client{http: &http.Client{Timeout: 30 * time.Second}, apiKey: key}, nil
}

func keychainLookup(service, account string) (string, error) {
	out, err := exec.Command("security", "find-generic-password", "-s", service, "-a", account, "-w").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// NowTimestamp returns the current UTC time formatted as Datafordeler expects:
// ISO-8601 with milliseconds and `Z` (e.g. "2026-05-15T07:36:00.000Z").
func NowTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

type GraphQLError struct {
	Message    string         `json:"message"`
	Path       []string       `json:"path,omitempty"`
	Extensions map[string]any `json:"extensions,omitempty"`
}

// QueryRaw POSTs a GraphQL query against the named register
// (e.g. "MAT", "BBR", "DAR", "DAGI", "EJF") and returns the raw response body.
//
// The API key never appears in returned error messages — Go's net/http will
// include the full URL (including the apiKey query parameter) in transport
// errors, so we wrap the request and scrub error strings before propagating.
func (c *Client) QueryRaw(register, query string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/%s/v1?apiKey=%s", graphqlBase, register, c.apiKey)
	safeURL := fmt.Sprintf("%s/%s/v1", graphqlBase, register)
	body, _ := json.Marshal(map[string]any{"query": query})
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %s", safeURL, c.scrubError(err))
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response from %s: %s", safeURL, c.scrubError(err))
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return raw, fmt.Errorf("HTTP 401 from %s — API key rejected. Verify the IT-system in portal.datafordeler.dk has the right tjenester subscribed", safeURL)
	}
	if resp.StatusCode >= 500 {
		return raw, fmt.Errorf("HTTP %d from %s", resp.StatusCode, safeURL)
	}
	return raw, nil
}

// scrubError redacts the API key from any error string. net/http transport
// errors include the full request URL — and our URL has the apiKey query
// parameter — so they'd otherwise leak the secret into logs and stack traces.
func (c *Client) scrubError(err error) string {
	if err == nil {
		return ""
	}
	return strings.ReplaceAll(err.Error(), c.apiKey, "<REDACTED>")
}

// decodeGraphQL unmarshals the GraphQL envelope and returns errors as a Go
// error with the first message + extensions code if available.
func decodeGraphQL[T any](raw []byte) (T, error) {
	var env struct {
		Data   T              `json:"data"`
		Errors []GraphQLError `json:"errors,omitempty"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return env.Data, fmt.Errorf("parse GraphQL response: %w (raw: %s)", err, trunc(string(raw), 200))
	}
	if len(env.Errors) > 0 {
		e := env.Errors[0]
		code, _ := e.Extensions["code"].(string)
		if code != "" {
			return env.Data, fmt.Errorf("graphql error [%s]: %s", code, e.Message)
		}
		return env.Data, fmt.Errorf("graphql error: %s", e.Message)
	}
	return env.Data, nil
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
