package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kaspermunck/dafcli/daf"
	"github.com/spf13/cobra"
)

var (
	probeRegister string
	probeEnvelope bool
)

var probeCmd = &cobra.Command{
	Use:   "probe <Type> <field> [field...]",
	Short: "Schema discovery — list which candidate fields exist on a Datafordeler GraphQL type",
	Long: `Datafordeler blocks GraphQL introspection (HC0046). The practical
discovery loop is to send a wide selection and read which fields the server
rejects. probe automates that — give it a type and candidate fields, and it
prints valid + invalid sets.

Specify the register with --register (MAT, BBR, DAR, DAGI; default MAT).
The Type is the GraphQL type without the register prefix (e.g. "Jordstykke",
not "MAT_Jordstykke").

Example: dafcli probe --register MAT Jordstykke matrikelnummer ejerlavskode bfeNummer status`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := daf.NewClientFromEnv()
		if err != nil {
			return err
		}
		typeName := args[0]
		fields := args[1:]

		ts := daf.NowTimestamp()
		query := fmt.Sprintf(`{
			%s_%s(registreringstid:%q, virkningstid:%q, first:1) {
				nodes { %s }
			}
		}`, probeRegister, typeName, ts, ts, strings.Join(fields, " "))

		raw, err := client.QueryRaw(probeRegister, query)
		if err != nil {
			return err
		}

		var env struct {
			Data   json.RawMessage     `json:"data"`
			Errors []daf.GraphQLError  `json:"errors,omitempty"`
		}
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		invalid := map[string]bool{}
		for _, e := range env.Errors {
			if f, ok := e.Extensions["field"].(string); ok {
				invalid[f] = true
			}
		}

		valid := []string{}
		invalidList := []string{}
		for _, f := range fields {
			if invalid[f] {
				invalidList = append(invalidList, f)
			} else {
				valid = append(valid, f)
			}
		}

		if probeEnvelope {
			return encodeJSON(daf.Wrap("FieldProbe", map[string]any{
				"register": probeRegister,
				"type":     typeName,
				"valid":    valid,
				"invalid":  invalidList,
			}))
		}

		w := os.Stdout
		fmt.Fprintf(w, "Probe %s_%s — %d valid, %d invalid\n", probeRegister, typeName, len(valid), len(invalidList))
		if len(valid) > 0 {
			fmt.Fprintln(w, "\nValid fields:")
			for _, f := range valid {
				fmt.Fprintf(w, "  ✓ %s\n", f)
			}
		}
		if len(invalidList) > 0 {
			fmt.Fprintln(w, "\nInvalid fields (rejected by server):")
			for _, f := range invalidList {
				fmt.Fprintf(w, "  ✗ %s\n", f)
			}
		}
		return nil
	},
}

func init() {
	probeCmd.Flags().StringVar(&probeRegister, "register", "MAT", "Datafordeler register: MAT, BBR, DAR, DAGI, EJF")
	probeCmd.Flags().BoolVar(&probeEnvelope, "envelope", false, "emit the shared envelope (Kind=FieldProbe)")
	rootCmd.AddCommand(probeCmd)
}
