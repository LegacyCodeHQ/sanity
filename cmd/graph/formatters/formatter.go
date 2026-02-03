package formatters

import "github.com/LegacyCodeHQ/sanity/parsers"

// Formatter is the interface that all graph formatters must implement.
type Formatter interface {
	// Format converts a dependency graph to a formatted string representation.
	Format(g parsers.DependencyGraph, opts FormatOptions) (string, error)
	// GenerateURL creates a shareable URL for the formatted output.
	// Returns the URL and true if supported, or ("", false) if not.
	GenerateURL(output string) (string, bool)
}
