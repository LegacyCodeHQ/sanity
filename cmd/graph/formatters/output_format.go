package formatters

// OutputFormat represents an output format type
type OutputFormat int

const (
	OutputFormatDOT OutputFormat = iota
	OutputFormatJSON
	OutputFormatMermaid
)

// String returns the string representation of the format
func (f OutputFormat) String() string {
	switch f {
	case OutputFormatDOT:
		return "dot"
	case OutputFormatJSON:
		return "json"
	case OutputFormatMermaid:
		return "mermaid"
	default:
		return "unknown"
	}
}

// ParseOutputFormat converts a string to OutputFormat
func ParseOutputFormat(s string) (OutputFormat, bool) {
	switch s {
	case "dot":
		return OutputFormatDOT, true
	case "json":
		return OutputFormatJSON, true
	case "mermaid":
		return OutputFormatMermaid, true
	default:
		return OutputFormatDOT, false
	}
}
