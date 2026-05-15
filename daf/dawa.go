package daf

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const dawaBase = "https://api.dataforsyningen.dk"

// DawaSearchHit is the slim record returned by /adresser?struktur=mini.
type DawaSearchHit struct {
	ID               string  `json:"id"`
	AdgangsadresseID string  `json:"adgangsadresseid"`
	Tekst            string  `json:"tekst"`
	X                float64 `json:"x"`
	Y                float64 `json:"y"`
	Kommunekode      string  `json:"kommunekode"`
}

// DawaAdgangsadresse is the full access-address record from /adgangsadresser/{id}.
// Includes matrikel + ejerlav, which are the join keys into Datafordeler MAT.
type DawaAdgangsadresse struct {
	ID                string `json:"id"`
	Adressebetegnelse string `json:"adressebetegnelse"`
	Husnr             string `json:"husnr"`
	Vejstykke         struct {
		Navn string `json:"navn"`
	} `json:"vejstykke"`
	Postnummer struct {
		Nr string `json:"nr"`
	} `json:"postnummer"`
	Kommune struct {
		Kode string `json:"kode"`
	} `json:"kommune"`
	Adgangspunkt struct {
		Koordinater []float64 `json:"koordinater"`
	} `json:"adgangspunkt"`
	Jordstykke struct {
		Matrikelnr string `json:"matrikelnr"`
		Ejerlav    struct {
			Kode int    `json:"kode"`
			Navn string `json:"navn"`
		} `json:"ejerlav"`
	} `json:"jordstykke"`
}

// DawaSearch returns the top-N hits for a free-text address query (DAWA, no auth).
func DawaSearch(query string, limit int) ([]DawaSearchHit, error) {
	if limit <= 0 {
		limit = 5
	}
	u := fmt.Sprintf("%s/adresser?q=%s&struktur=mini&per_side=%d", dawaBase, url.QueryEscape(query), limit)
	body, err := dawaGet(u)
	if err != nil {
		return nil, err
	}
	var hits []DawaSearchHit
	if err := json.Unmarshal(body, &hits); err != nil {
		return nil, fmt.Errorf("parse DAWA search: %w", err)
	}
	return hits, nil
}

// DawaAdgangsadresseDetails returns the full access-address record by UUID.
func DawaAdgangsadresseDetails(adgangsadresseID string) (*DawaAdgangsadresse, error) {
	u := fmt.Sprintf("%s/adgangsadresser/%s", dawaBase, url.PathEscape(adgangsadresseID))
	body, err := dawaGet(u)
	if err != nil {
		return nil, err
	}
	var a DawaAdgangsadresse
	if err := json.Unmarshal(body, &a); err != nil {
		return nil, fmt.Errorf("parse DAWA adgangsadresse: %w", err)
	}
	return &a, nil
}

func dawaGet(url string) ([]byte, error) {
	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return nil, fmt.Errorf("DAWA request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	// DAWA's chunked-encoding terminator is sometimes malformed → io.ReadAll
	// returns io.ErrUnexpectedEOF even though the response body is complete
	// (verified: bodies end with the expected ']' or '}'). Tolerate the EOF
	// when we have data and let the JSON parser decide whether the payload is
	// valid.
	if err != nil && err != io.ErrUnexpectedEOF {
		return raw, err
	}
	if len(raw) == 0 {
		return raw, fmt.Errorf("DAWA returned empty body")
	}
	if resp.StatusCode != http.StatusOK {
		return raw, fmt.Errorf("DAWA HTTP %d: %s", resp.StatusCode, trunc(string(raw), 200))
	}
	return raw, nil
}
