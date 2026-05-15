package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dafcli",
	Short: "CLI for Datafordeler — Danish grunddata via GraphQL",
	Long: `dafcli queries Klimadatastyrelsen's Datafordeler platform via the
modern GraphQL stack (graphql.datafordeler.dk).

Subcommands wrap MAT (Matriklen2), BBR (buildings), and the unauthenticated
DAWA address service. EJF, DAGI, and full DAR coverage to follow.

Auth: set DAF_API_KEY in the environment, or store it in macOS
Keychain as service "dafcli" / account "DAF_API_KEY".`,
}

// Execute is the entry point used by main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
