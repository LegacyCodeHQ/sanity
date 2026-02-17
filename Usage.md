# Clarity Usage

> Generated using [lens](../tiny-tools/rust/lens) — extracted from cobra command definitions in `cmd/`.

A software design tool for AI-native developers and coding agents.

```
clarity <COMMAND> [OPTIONS]
```

**Use cases:**
- Keep a live impact view while coding with `clarity watch`
- Generate focused change snapshots with `clarity show`
- Run repeatable design checks in developer and coding-agent workflows

## Global Flags

Inherited by all subcommands. Extracted from `cmd/root.go`.

| Flag | Short | Default | Description |
|---|---|---|---|
| `--verbose` | `-v` | `false` | Enable verbose/debug output |
| `--version` | `-V` | `false` | Print version information and exit |
## Commands

| Command | Description | Source |
|---|---|---|
| `diff` | Show dependency-graph changes between snapshots | `/Users/ragunath/legacy-code-hq-ecosystem/products/clarity/cmd/diff/diff_cmd.go` |
| `languages` | List all supported languages and file extensions | `/Users/ragunath/legacy-code-hq-ecosystem/products/clarity/cmd/languages/languages_cmd.go` |
| `setup` | Add clarity usage instructions to AGENTS.md | `/Users/ragunath/legacy-code-hq-ecosystem/products/clarity/cmd/setup/setup_cmd.go` |
| `show` | Show a scoped file-based dependency graph | `/Users/ragunath/legacy-code-hq-ecosystem/products/clarity/cmd/show/show_cmd.go` |
| `watch` | Watch for file changes and serve a live dependency graph | `/Users/ragunath/legacy-code-hq-ecosystem/products/clarity/cmd/watch/watch_cmd.go` |
| `why <from> <to>` | Show direct dependency direction(s) between two files | `/Users/ragunath/legacy-code-hq-ecosystem/products/clarity/cmd/why/why_cmd.go` |

---


## `clarity diff`

Show dependency-graph changes between snapshots

```
clarity diff [OPTIONS]
```

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--repo` | `-r` | string | `""` | Git repository path (default: current directory) |
| `--format` | `-f` | string | `opts.outputFmt` | fmt.Sprintf("Output format (%s)", formatters.SupportedFormats()) |
| `--commit` | `-c` | string | `""` | Compare committed snapshots (<commit> or <A>,<B>) |
| `--summary` | | bool | `false` | Print text summary only |

---


## `clarity languages`

List all supported languages and file extensions

```
clarity languages [OPTIONS]
```

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|

---


## `clarity setup`

Add clarity usage instructions to AGENTS.md

```
clarity setup [OPTIONS]
```

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|

---


## `clarity show`

Show a scoped file-based dependency graph

```
clarity show [OPTIONS]
```

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--format` | `-f` | string | `opts.outputFormat` | fmt.Sprintf("Output format (%s)", formatters.SupportedFormats()) |
| `--repo` | `-r` | string | `""` | Git repository path (default: current directory) |
| `--commit` | `-c` | string | `""` | Git commit or range to analyze (e.g., f0459ec, HEAD~3, f0459ec...be3d11a) |
| `--file` | `-p` | string | `""` | Show dependencies for a specific file |
| `--url` | `-u` | bool | `false` | Generate visualization URL (supported formats: dot, mermaid) |
| `--input` | `-i` | []string | `nil` | Build graph from specific files and/or directories (comma-separated) |
| `--between` | `-w` | []string | `nil` | Find all paths between specified files (comma-separated) |
| `--level` | `-l` | int | `opts.depthLevel` | Depth level for dependencies (used with --file) |
| `--include-ext` | | string | `""` | Include only files with these extensions (comma-separated, e.g. .go,.java) |
| `--exclude-ext` | | string | `""` | Exclude files with these extensions (comma-separated, e.g. .go,.java) |
| `--allow-outside-repo` | | bool | `false` | Allow input paths outside the repo root |
| `--exclude` | | []string | `nil` | Exclude specific files and/or directories from graph inputs (comma-separated) |

---


## `clarity watch`

Watch for file changes and serve a live dependency graph

```
clarity watch [OPTIONS]
```

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--repo` | `-r` | string | `""` | Git repository path (default: current directory) |
| `--input` | `-i` | []string | `nil` | Watch specific files and/or directories (comma-separated) |
| `--port` | `-P` | int | `opts.port` | HTTP server port |
| `--include-ext` | | string | `""` | Include only files with these extensions (comma-separated, e.g. .go,.java) |
| `--exclude-ext` | | string | `""` | Exclude files with these extensions (comma-separated, e.g. .go,.java) |
| `--exclude` | | []string | `nil` | Exclude specific files and/or directories (comma-separated) |

---


## `clarity why <from> <to>`

Show direct dependency direction(s) between two files

```
clarity why <from> <to> [OPTIONS]
```

| Flag | Short | Type | Default | Description |
|---|---|---|---|---|
| `--format` | `-f` | string | `opts.outputFormat` | fmt.Sprintf("Output format (%s)", supportedFormats()) |
| `--repo` | `-r` | string | `""` | Git repository path (default: current directory) |
| `--allow-outside-repo` | | bool | `false` | Allow input paths outside the repo root |

---

