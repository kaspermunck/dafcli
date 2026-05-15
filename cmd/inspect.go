package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kaspermunck/dafcli/daf"
	"github.com/spf13/cobra"
)

var (
	inspectJSON     bool
	inspectEnvelope bool
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <address>",
	Short: "End-to-end DAWA → MAT → BBR chain for a Danish address",
	Long: `Resolves an address via DAWA, looks up its parcel in MAT (with BFE),
and lists the buildings on the access-address from BBR. The one-shot path for
"tell me what you can see about this address."`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		for _, a := range args[1:] {
			query += " " + a
		}

		hits, err := daf.DawaSearch(query, 5)
		if err != nil {
			return err
		}
		if len(hits) == 0 {
			return fmt.Errorf("no DAWA matches for %q", query)
		}
		full, err := daf.DawaAdgangsadresseDetails(hits[0].AdgangsadresseID)
		if err != nil {
			return err
		}

		client, err := daf.NewClientFromEnv()
		if err != nil {
			return err
		}

		ejerlavLokalId := strconv.Itoa(full.Jordstykke.Ejerlav.Kode)
		jordstykker, err := client.Jordstykker(full.Jordstykke.Matrikelnr, ejerlavLokalId, 5)
		if err != nil {
			return err
		}

		bygninger, err := client.BygningerByHusnummer(full.ID, 20)
		if err != nil {
			// Buildings query failure is non-fatal — log and continue.
			fmt.Fprintf(os.Stderr, "warning: BBR lookup failed: %v\n", err)
		}

		payload := map[string]any{
			"adresse":     full,
			"jordstykker": jordstykker,
			"bygninger":   bygninger,
		}
		if inspectEnvelope {
			return encodeJSON(daf.Wrap("AddressInspection", payload))
		}
		if inspectJSON {
			return encodeJSON(payload)
		}

		printInspect(full, jordstykker, bygninger)
		return nil
	},
}

func init() {
	inspectCmd.Flags().BoolVar(&inspectJSON, "json", false, "print full chain result as JSON")
	inspectCmd.Flags().BoolVar(&inspectEnvelope, "envelope", false, "emit the shared envelope (Kind=AddressInspection)")
	rootCmd.AddCommand(inspectCmd)
}

func printInspect(a *daf.DawaAdgangsadresse, js []daf.Jordstykke, bs []daf.Bygning) {
	w := os.Stdout
	fmt.Fprintf(w, "Adresse: %s\n", a.Adressebetegnelse)
	fmt.Fprintf(w, "  adgangsadresseUUID: %s\n", a.ID)
	fmt.Fprintf(w, "  matrikelnr / ejerlav: %s, %s (%d)\n", a.Jordstykke.Matrikelnr, a.Jordstykke.Ejerlav.Navn, a.Jordstykke.Ejerlav.Kode)

	if len(js) > 0 {
		fmt.Fprintln(w, "\nMAT_Jordstykke")
		for _, j := range js {
			fmt.Fprintf(w, "  - id_lokalId %s, %d m², status %s, BFE %s\n", j.IDLokalId, j.RegistreretAreal, j.Status, j.SamletFastEjendomLokalId)
		}
	}

	fmt.Fprintf(w, "\nBBR_Bygning — %d stk.\n", len(bs))
	for _, b := range bs {
		anvLabel := daf.BBRAnvendelseLabel(b.Anvendelse)
		anv := b.Anvendelse
		if anvLabel != "" {
			anv = fmt.Sprintf("%s (%s)", b.Anvendelse, anvLabel)
		}
		fmt.Fprintf(w, "  - bygning %d, anvendelse %s, status %s\n", b.Bygningsnummer, anv, b.Status)
	}

	fmt.Fprintln(w, "\nEjer + pant: ikke tilgængelig — EJF kræver godkendelse, Tinglysning kører separat.")
}
