package cmd

import (
	"fmt"
	"os"

	"github.com/kaspermunck/dafcli/daf"
	"github.com/spf13/cobra"
)

var (
	bygningHusnummer string
	bygningLimit     int
	bygningJSON      bool
	bygningEnvelope  bool
)

var bygningCmd = &cobra.Command{
	Use:   "bygning",
	Short: "Look up BBR buildings",
	Long: `Returns BBR_Bygning records. Currently supports lookup by --husnummer
(an access-address UUID, e.g. from DAWA's adgangsadresseid).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if bygningHusnummer == "" {
			return fmt.Errorf("--husnummer is required (DAR Husnummer.id_lokalId / DAWA adgangsadresseid)")
		}
		client, err := daf.NewClientFromEnv()
		if err != nil {
			return err
		}
		nodes, err := client.BygningerByHusnummer(bygningHusnummer, bygningLimit)
		if err != nil {
			return err
		}
		if len(nodes) == 0 {
			return fmt.Errorf("no buildings registered against husnummer %s", bygningHusnummer)
		}
		if bygningEnvelope {
			return encodeJSON(daf.Wrap("Bygning", nodes))
		}
		if bygningJSON {
			return encodeJSON(nodes)
		}
		printBygninger(nodes)
		return nil
	},
}

func init() {
	bygningCmd.Flags().StringVar(&bygningHusnummer, "husnummer", "", "access-address UUID (DAWA adgangsadresseid)")
	bygningCmd.Flags().IntVar(&bygningLimit, "limit", 20, "max buildings to return")
	bygningCmd.Flags().BoolVar(&bygningJSON, "json", false, "print parsed result as JSON")
	bygningCmd.Flags().BoolVar(&bygningEnvelope, "envelope", false, "emit the shared envelope (Kind=Bygning)")
	rootCmd.AddCommand(bygningCmd)
}

func printBygninger(nodes []daf.Bygning) {
	w := os.Stdout
	fmt.Fprintf(w, "Bygninger på husnummer %s — %d stk.\n", nodes[0].Husnummer, len(nodes))
	for _, b := range nodes {
		anvLabel := daf.BBRAnvendelseLabel(b.Anvendelse)
		anv := b.Anvendelse
		if anvLabel != "" {
			anv = fmt.Sprintf("%s (%s)", b.Anvendelse, anvLabel)
		}
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  bygningsnummer:    %d\n", b.Bygningsnummer)
		fmt.Fprintf(w, "  id_lokalId:        %s\n", b.IDLokalId)
		fmt.Fprintf(w, "  anvendelse:        %s\n", anv)
		fmt.Fprintf(w, "  status:            %s\n", b.Status)
	}
}
