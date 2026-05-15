package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kaspermunck/dafcli/daf"
	"github.com/spf13/cobra"
)

var (
	jordstykkeEjerlav   string
	jordstykkeLimit     int
	jordstykkeJSON      bool
	jordstykkeEnvelope  bool
)

var jordstykkeCmd = &cobra.Command{
	Use:   "jordstykke <matrikelnummer>",
	Short: "Look up cadastral parcels (MAT_Jordstykke)",
	Long: `Queries Datafordeler MAT for parcels matching the given matrikelnummer.
Use --ejerlav to narrow when the same matrikelnummer occurs in multiple
ejerlav. The ejerlav value is the integer kode from DAWA (e.g. 2006351).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := daf.NewClientFromEnv()
		if err != nil {
			return err
		}
		nodes, err := client.Jordstykker(args[0], jordstykkeEjerlav, jordstykkeLimit)
		if err != nil {
			return err
		}
		if len(nodes) == 0 {
			return fmt.Errorf("no jordstykke matches for matrikelnummer %q", args[0])
		}
		if jordstykkeEnvelope {
			return encodeJSON(daf.Wrap("Jordstykke", nodes))
		}
		if jordstykkeJSON {
			return encodeJSON(nodes)
		}
		printJordstykker(nodes)
		return nil
	},
}

func init() {
	jordstykkeCmd.Flags().StringVar(&jordstykkeEjerlav, "ejerlav", "", "narrow by ejerlavLokalId (integer kode from DAWA, as string)")
	jordstykkeCmd.Flags().IntVar(&jordstykkeLimit, "limit", 10, "max results")
	jordstykkeCmd.Flags().BoolVar(&jordstykkeJSON, "json", false, "print parsed result as JSON")
	jordstykkeCmd.Flags().BoolVar(&jordstykkeEnvelope, "envelope", false, "emit the shared envelope (Kind=Jordstykke)")
	rootCmd.AddCommand(jordstykkeCmd)
}

func printJordstykker(nodes []daf.Jordstykke) {
	w := os.Stdout
	for i, j := range nodes {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "Jordstykke %s\n", j.Matrikelnummer)
		fmt.Fprintf(w, "  id_lokalId:               %s\n", j.IDLokalId)
		fmt.Fprintf(w, "  status:                   %s\n", j.Status)
		fmt.Fprintf(w, "  registreretAreal:         %s m²\n", strconv.Itoa(j.RegistreretAreal))
		fmt.Fprintf(w, "  ejerlavLokalId:           %s\n", j.EjerlavLokalId)
		if j.SamletFastEjendomLokalId != "" {
			fmt.Fprintf(w, "  BFE (via SFE):            %s\n", j.SamletFastEjendomLokalId)
		}
		if j.RegistreringFra != "" {
			fmt.Fprintf(w, "  registreringFra:          %s\n", j.RegistreringFra)
		}

		if j.SamletFastEjendomLokalId != "" {
			fmt.Fprintln(w, "\n  Next steps")
			fmt.Fprintf(w, "    dafcli sfe %s\n", j.SamletFastEjendomLokalId)
		}
	}
}
