# Sanity

Sanity is a CLI tool for analyzing and visualizing dependency graphs in your codebase, with support for Dart and Go files.

## Command Help System

Sanity uses [Cobra](https://github.com/spf13/cobra), a powerful CLI framework for Go that automatically generates comprehensive help documentation for all commands.

### Accessing Command Descriptions

You can view descriptions for all commands using the built-in help system:

#### View Root Command Help
```bash
sanity --help
# or
sanity -h
```

This displays:
- The root command description
- A list of all available commands with their short descriptions
- Available global flags

#### View Specific Command Help
```bash
sanity graph --help
# or
sanity help graph
```

This displays:
- The command's detailed long description
- Usage syntax
- All available flags with their descriptions
- Example usage

### How Command Descriptions Are Defined

Command descriptions are defined in the code using Cobra's command structure:

#### Command-Level Descriptions

Each command is defined with three description fields:

1. **`Use`**: Command usage syntax
   ```go
   Use: "graph [files...]"
   ```

2. **`Short`**: Brief description shown in command lists
   ```go
   Short: "Generate dependency graph for project imports"
   ```

3. **`Long`**: Detailed description shown in `--help` output
   ```go
   Long: `Analyzes files and generates a dependency graph...
   
   Supports three modes:
     1. Explicit files: Analyze specific files
     2. Uncommitted files: Analyze all uncommitted files (--repo)
     3. Commit analysis: Analyze files changed in a commit (--repo --commit)
   ...`,
   ```

#### Flag Descriptions

Flags are documented using description strings in the flag definition:

```go
graphCmd.Flags().StringVarP(&outputFormat, "format", "f", "list", 
    "Output format (list, json, dot)")
```

The last parameter is the description shown in the help output.

### Available Commands

#### `sanity graph`

Generate dependency graphs for project files. Analyzes Dart and Go files to show import relationships.

**Location**: `cmd/graph.go`

**Flags**:
- `--format, -f`: Output format (list, json, dot) - default: "list"
- `--repo, -r`: Git repository path to analyze uncommitted files
- `--commit, -c`: Git commit to analyze (requires --repo)

**Examples**:
```bash
# Analyze specific files
sanity graph file1.dart file2.dart file3.dart

# Analyze uncommitted files in current repository
sanity graph --repo .

# Analyze files changed in a specific commit
sanity graph --repo . --commit abc123

# Output in JSON format
sanity graph --repo . --commit HEAD~1 --format=json

# Output in Graphviz DOT format for visualization
sanity graph --repo /path/to/repo --commit main --format=dot
```

## Adding New Commands

To add a new command with descriptions:

1. Create a new file in the `cmd/` directory (e.g., `cmd/newcommand.go`)
2. Define the command with `Use`, `Short`, and `Long` fields
3. Register it in `cmd/root.go` using `rootCmd.AddCommand(newCommand)`

Example:
```go
var newCommand = &cobra.Command{
    Use:   "newcommand",
    Short: "Brief description",
    Long:  `Detailed description with examples and usage`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // Command logic
        return nil
    },
}
```

Cobra will automatically:
- Include it in the command list when running `sanity --help`
- Generate help text from the `Long` description when running `sanity newcommand --help`
- Provide tab completion support (via `sanity completion`)

## Getting Help

- **List all commands**: `sanity --help`
- **Command-specific help**: `sanity <command> --help`
- **Help command alias**: `sanity help <command>`

The help system is fully integrated into Cobra and requires no additional setup - all descriptions defined in the command structs are automatically rendered when users request help.
