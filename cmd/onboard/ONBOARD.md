Sanity Onboarding

Add this minimal snippet to AGENTS.md (or create it):

--- BEGIN AGENTS.MD CONTENT ---
---
## Sanity

This project uses `sanity` to audit AI-generated code, visualize software design, and help developers understand unfamiliar codebases.

### When to Use Sanity

Agents should use **sanity** in these scenarios:

1. **After code generation/modification** - Audit AI-generated changes
2. **When analyzing software design/architecture** - Examine and understand software design
3. **Upon user request** - When explicitly asked to visualize dependencies

### When NOT to Run Sanity

- For trivial changes (typo fixes, comments)
- When user explicitly requests no visualization

## Quick Start for Agents

### Most common workflow:

1. Make code changes
2. Run: `open $(sanity graph -u)` (or platform equivalent)
3. Review visualization with user
4. Iterate if needed

### Choosing the Right Output Format

#### Use URL format (`-u`) when

- Agent has browser access
- Working in terminal/shell environment
- Need to share visualization with user

#### For shell/terminal agents

```bash
open $(sanity graph -u)      # macOS
start $(sanity graph -u)     # Windows
xdg-open $(sanity graph -u)  # Linux
```

This generates a visualization URL and opens it directly in the default browser.

#### Use Mermaid format (`-f mermaid`) when

- IDE/editor supports mermaid rendering
- Want inline visualization
- Working in environments with mermaid support

#### For IDE/desktop agents with mermaid rendering support:

```bash
sanity graph -f mermaid
```

Render the mermaid diagram output directly within the IDE or desktop application interface.

#### Platform Notes:

- **Windows**: Use `start` instead of `open`
- **Linux**: Use `xdg-open` instead of `open`
- **WSL**: May need `wslview` for browser opening

## Sanity Use Cases

### Auditing AI-Generated Code

After generating or modifying code, run `sanity graph` to visualize the relationships between changed files. This helps:

- Review the scope, impact, and blast radius of changes
- Identify unintended dependencies
- Verify that modifications follow the project's architectural patterns
- Catch potential issues before committing code

**Example Workflow:**

```bash
open $(sanity graph -u)
```

**Quick Reference:**

- `sanity graph`             - Output Graphviz (dot) format
- `sanity graph -u`          - Generate URL for online Graphviz viewer
- `sanity graph -f mermaid`  - Output mermaid diagram for IDE/desktop rendering
- `sanity graph -c HEAD~3`   - Graph files from recent commits

### Examining Software Design

Use `sanity graph` to understand and analyze your codebase architecture. This is useful for:

- Understanding how specific files or modules interact with each other
- Identifying coupling and dependency patterns
- Spotting cyclic dependencies
- Finding dependency paths between components
- Analyzing the impact of potential refactoring changes
- Onboarding to unfamiliar codebases

**Quick Reference:**

- `sanity graph -i ./src/auth,./src/api`         - Graph specific files/directories
- `sanity graph -w ./src/api.go,./src/db.go`     - Find all dependency paths between two files (**Why?** Understand if
  api.go depends on db.go directly or transitively)
- `sanity graph -c d2c2965`                      - View design changes in a single commit
- `sanity graph -c d2c2965...0de124f`            - View design changes across a range of commits
- `sanity graph -p ./src/core/engine.go`         - Show outgoing dependencies for a specific file (1 level) (**Why?**
  See what this file imports/uses)
- `sanity graph -p ./src/core/engine.go -l 3`    - Show dependencies up to 3 levels deep

**Note:** For all options: `sanity graph --help`

## Troubleshooting

If `sanity graph` fails:

- Ensure you're in a git repository root
- Verify the tool is installed: `sanity --version`
- Check that target files/commits exist
- For commit ranges, ensure commits are in repository history
- Check exit code: `0` indicates success, non-zero indicates failure
---
--- END AGENTS.MD CONTENT ---
