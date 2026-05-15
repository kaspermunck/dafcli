package cmd

import (
	"fmt"
	"os"

	"github.com/kaspermunck/dafcli/daf"
	"github.com/spf13/cobra"
)

var (
	adresseLimit    int
	adresseJSON     bool
	adresseRaw      bool
	adresseEnvelope bool
)

var adresseCmd = &cobra.Command{
	Use:   "adresse <text>",
	Short: "Look up Danish addresses via DAWA (no auth required)",
	Long: `Resolves a free-text address against DAWA (Danmarks Adressers Web API).
Returns the top match with its access-address UUID, matrikel, and ejerlav —
the join keys you need for follow-up queries against MAT and BBR.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		for _, a := range args[1:] {
			query += " " + a
		}

		hits, err := daf.DawaSearch(query, adresseLimit)
		if err != nil {
			return err
		}
		if len(hits) == 0 {
			return fmt.Errorf("no DAWA matches for %q", query)
		}

		// Always fetch full details for the top match (gives matrikel + ejerlav).
		full, err := daf.DawaAdgangsadresseDetails(hits[0].AdgangsadresseID)
		if err != nil {
			return err
		}

		if adresseEnvelope {
			return encodeJSON(daf.Wrap("Address", map[string]any{
				"top":          full,
				"alternatives": hits[1:],
			}))
		}
		if adresseRaw {
			return encodeJSON(full)
		}
		if adresseJSON {
			return encodeJSON(map[string]any{
				"top":         full,
				"alternatives": hits[1:],
			})
		}

		printDawa(full, hits[1:])
		return nil
	},
}

func init() {
	adresseCmd.Flags().IntVar(&adresseLimit, "limit", 5, "max DAWA candidates to fetch")
	adresseCmd.Flags().BoolVar(&adresseJSON, "json", false, "print parsed result as JSON")
	adresseCmd.Flags().BoolVar(&adresseRaw, "raw", false, "print only the full DAWA adgangsadresse record (JSON)")
	adresseCmd.Flags().BoolVar(&adresseEnvelope, "envelope", false, "emit the shared {source,kind,version,data,fetchedAt} envelope (Kind=Address)")
	rootCmd.AddCommand(adresseCmd)
}

func printDawa(a *daf.DawaAdgangsadresse, alts []daf.DawaSearchHit) {
	w := os.Stdout
	fmt.Fprintln(w, "Adresse")
	fmt.Fprintf(w, "  %s\n", a.Adressebetegnelse)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Identifiers")
	fmt.Fprintf(w, "  adgangsadresseUUID: %s\n", a.ID)
	fmt.Fprintf(w, "  matrikelnummer:     %s\n", a.Jordstykke.Matrikelnr)
	fmt.Fprintf(w, "  ejerlavskode:       %d\n", a.Jordstykke.Ejerlav.Kode)
	fmt.Fprintf(w, "  ejerlavsnavn:       %s\n", a.Jordstykke.Ejerlav.Navn)
	fmt.Fprintf(w, "  kommunekode:        %s\n", a.Kommune.Kode)
	if len(a.Adgangspunkt.Koordinater) >= 2 {
		fmt.Fprintf(w, "  ETRS89 (lon, lat):  %.6f, %.6f\n", a.Adgangspunkt.Koordinater[0], a.Adgangspunkt.Koordinater[1])
	}

	if len(alts) > 0 {
		fmt.Fprintln(w, "\nAlternative matches")
		for _, h := range alts {
			fmt.Fprintf(w, "  - %s\n", h.Tekst)
		}
	}

	fmt.Fprintln(w, "\nNext steps")
	fmt.Fprintf(w, "  dafcli jordstykke %s --ejerlav %d\n", a.Jordstykke.Matrikelnr, a.Jordstykke.Ejerlav.Kode)
	fmt.Fprintf(w, "  dafcli bygning --husnummer %s\n", a.ID)
}
