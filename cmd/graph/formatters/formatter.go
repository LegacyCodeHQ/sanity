package formatters

import (
	"fmt"

	"github.com/LegacyCodeHQ/sanity/git"
	"github.com/LegacyCodeHQ/sanity/parsers"
)

// FormatOptions contains optional parameters for formatting dependency graphs.
type FormatOptions struct {
	// Label is an optional title or label for the graph
	Label string
	// FileStats contains file statistics (additions/deletions) for display in nodes
	FileStats map[string]git.FileStats
}

// Formatter is the interface that all graph formatters must implement.
type Formatter interface {
	// Format converts a dependency graph to a formatted string representation.
	Format(g parsers.DependencyGraph, opts FormatOptions) (string, error)
}

// NewFormatter creates a Formatter for the specified format type.
// Supported formats: "json", "dot", "mermaid"
func NewFormatter(format string) (Formatter, error) {
	switch format {
	case "json":
		return &JSONFormatter{}, nil
	case "dot":
		return &DOTFormatter{}, nil
	case "mermaid":
		return &MermaidFormatter{}, nil
	default:
		return nil, fmt.Errorf("unknown format: %s (valid options: dot, json, mermaid)", format)
	}
}
