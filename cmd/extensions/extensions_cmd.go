package extensions

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/LegacyCodeHQ/clarity/depgraph/registry"
	"github.com/spf13/cobra"
)

type extensionInfo struct {
	Extension string `json:"extension"`
	Language  string `json:"language"`
	Maturity  string `json:"maturity"`
}

type extensionsOutput struct {
	Extensions []extensionInfo `json:"extensions"`
}

type options struct {
	format string
}

// Cmd represents the extensions command.
var Cmd = NewCommand()

// NewCommand returns a new extensions command instance.
func NewCommand() *cobra.Command {
	opts := &options{
		format: "json",
	}

	cmd := &cobra.Command{
		Use:   "extensions",
		Short: "List supported file extensions in machine-readable format",
		Long: `List supported file extensions and their mapped language.

Examples:
  clarity extensions
  clarity extensions --format text`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runExtensions(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.format, "format", opts.format, "Output format (json, text)")
	return cmd
}

func runExtensions(cmd *cobra.Command, opts *options) error {
	entries := buildExtensionEntries()

	switch strings.ToLower(opts.format) {
	case "json":
		payload := extensionsOutput{Extensions: entries}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(payload)
	case "text":
		writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintln(writer, "EXTENSION\tLANGUAGE\tMATURITY"); err != nil {
			return err
		}
		for _, entry := range entries {
			if _, err := fmt.Fprintf(
				writer,
				"%s\t%s\t%s\n",
				entry.Extension,
				entry.Language,
				entry.Maturity); err != nil {
				return err
			}
		}
		return writer.Flush()
	default:
		return fmt.Errorf("unknown format: %s (valid options: json, text)", opts.format)
	}
}

func buildExtensionEntries() []extensionInfo {
	entries := make([]extensionInfo, 0, 16)
	for _, language := range registry.SupportedLanguages() {
		for _, ext := range language.Extensions {
			entries = append(entries, extensionInfo{
				Extension: ext,
				Language:  language.Name,
				Maturity:  language.Maturity.DisplayName(),
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Extension == entries[j].Extension {
			return entries[i].Language < entries[j].Language
		}
		return entries[i].Extension < entries[j].Extension
	})

	return entries
}
