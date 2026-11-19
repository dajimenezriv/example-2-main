package cmdutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

const (
	OutputTable = "table"
	OutputJSON  = "json"

	tablePadding = 2
)

var ErrInvalidOutput = errors.New("invalid output")

func AddOutputFlag(cmd *cobra.Command, output *string) {
	cmd.Flags().StringVarP(output, "output", "o", OutputTable, "output format: table or json")
}

func CheckOutput(output string) error {
	// Commands call this before any work, so a typo fails fast instead of falling back to a table.
	if output != OutputTable && output != OutputJSON {
		return fmt.Errorf("%w %q: want %s or %s", ErrInvalidOutput, output, OutputTable, OutputJSON)
	}
	return nil
}

func MarkRequired(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		// Only fails for an undefined flag: a programming error, caught on first run.
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(err)
		}
	}
}

func RegisterFlagValues(cmd *cobra.Command, name string, values []string) {
	// Only fails for an undefined flag, like MarkRequired.
	if err := cmd.RegisterFlagCompletionFunc(name, cobra.FixedCompletions(values, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}
}

func PrintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("write json: %w", err)
	}
	return nil
}

func PrintTable(w io.Writer, header []string, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 0, 0, tablePadding, ' ', 0)
	for _, row := range append([][]string{header}, rows...) {
		if _, err := io.WriteString(tw, strings.Join(row, "\t")+"\n"); err != nil {
			return fmt.Errorf("write table: %w", err)
		}
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("write table: %w", err)
	}
	return nil
}
