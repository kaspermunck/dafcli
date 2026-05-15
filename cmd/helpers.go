package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// writePrettyJSON pretty-prints a raw JSON byte slice to stdout, falling back
// to the original bytes if indentation fails.
func writePrettyJSON(raw []byte) error {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		_, _ = os.Stdout.Write(raw)
		fmt.Fprintln(os.Stdout)
		return nil
	}
	_, _ = os.Stdout.Write(pretty.Bytes())
	fmt.Fprintln(os.Stdout)
	return nil
}

func encodeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
