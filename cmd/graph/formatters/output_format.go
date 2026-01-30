package formatters

// OutputFormat represents an output format type
type OutputFormat string

const (
	OutputFormatDOT     OutputFormat = "dot"
	OutputFormatJSON    OutputFormat = "json"
	OutputFormatMermaid OutputFormat = "mermaid"
)

// String returns the string representation of the format
func (f OutputFormat) String() string {
	return string(f)
}
