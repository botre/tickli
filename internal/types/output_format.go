package types

import (
	"fmt"
	"github.com/spf13/cobra"
)

type OutputFormat string

const (
	OutputSimple OutputFormat = "simple"
	OutputJSON   OutputFormat = "json"
	OutputQuiet  OutputFormat = "quiet"
)

var OutputFormatCompletion = []cobra.Completion{
	cobra.CompletionWithDesc("simple", "Simple output format"),
	cobra.CompletionWithDesc("json", "JSON output format"),
}

var OutputFormatCompletionFunc = cobra.FixedCompletions(OutputFormatCompletion, cobra.ShellCompDirectiveNoFileComp)

func (o *OutputFormat) Set(value string) error {
	switch OutputFormat(value) {
	case OutputSimple, OutputJSON, OutputQuiet:
		*o = OutputFormat(value)
	default:
		return fmt.Errorf("invalid output format: %s", value)
	}
	return nil
}

// ResolveOutput checks the --json and --quiet persistent flags
// and returns the effective output format.
func ResolveOutput(o OutputFormat, jsonFlag, quietFlag bool) OutputFormat {
	if jsonFlag {
		return OutputJSON
	}
	if quietFlag {
		return OutputQuiet
	}
	return o
}

func (o OutputFormat) String() string {
	return string(o)
}

func (o OutputFormat) Type() string {
	return "OutputFormat"
}
