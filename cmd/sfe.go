package cmd

import (
	"fmt"
	"os"

	"github.com/kaspermunck/dafcli/daf"
	"github.com/spf13/cobra"
)

var (
	sfeJSON     bool
	sfeEnvelope bool
)

var sfeCmd = &cobra.Command{
	Use:   "sfe <bfe-nummer>",
	Short: "Look up a Samlet Fast Ejendom by BFE-nummer",
	Long: `Returns the SFE for a given BFE-nummer. The BFE-nummer is the SFE's
id_lokalId — i.e. the value you get back from a Jordstykke as
samletFastEjendomLokalId.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := daf.NewClientFromEnv()
		if err != nil {
			return err
		}
		sfe, err := client.SFEByBFE(args[0])
		if err != nil {
			return err
		}
		if sfeEnvelope {
			return encodeJSON(daf.Wrap("SamletFastEjendom", sfe))
		}
		if sfeJSON {
			return encodeJSON(sfe)
		}
		w := os.Stdout
		fmt.Fprintln(w, "Samlet Fast Ejendom")
		fmt.Fprintf(w, "  BFE-nummer (id_lokalId): %s\n", sfe.IDLokalId)
		fmt.Fprintf(w, "  status:                  %s\n", sfe.Status)
		if sfe.DatafordelerRowId != "" {
			fmt.Fprintf(w, "  datafordelerRowId:       %s\n", sfe.DatafordelerRowId)
		}
		return nil
	},
}

func init() {
	sfeCmd.Flags().BoolVar(&sfeJSON, "json", false, "print parsed result as JSON")
	sfeCmd.Flags().BoolVar(&sfeEnvelope, "envelope", false, "emit the shared envelope (Kind=SamletFastEjendom)")
	rootCmd.AddCommand(sfeCmd)
}
