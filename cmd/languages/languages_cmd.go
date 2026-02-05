package languages

import (
	"fmt"
	"strings"

	"github.com/LegacyCodeHQ/sanity/depgraph"
	"github.com/spf13/cobra"
)

// Cmd represents the languages command.
var Cmd = NewCommand()

// NewCommand returns a new languages command instance.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "languages",
		Short: "List all supported languages and file extensions",
		Long: `List all supported programming languages and their mapped file extensions.

Examples:
  sanity languages`,
		RunE: runLanguages,
	}

	return cmd
}

func runLanguages(cmd *cobra.Command, _ []string) error {
	languages := depgraph.SupportedLanguages()

	for _, language := range languages {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\n", language.Name, strings.Join(language.Extensions, ", ")); err != nil {
			return err
		}
	}

	return nil
}
