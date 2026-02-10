# Clarity (formerly Sanity)

[![Built with Clarity](https://raw.githubusercontent.com/LegacyCodeHQ/clarity/main/badges/built-with-clarity-sunrise.svg)](https://raw.githubusercontent.com/LegacyCodeHQ/clarity/main/badges/built-with-clarity-sunrise.svg)
[![License](https://img.shields.io/github/license/LegacyCodeHQ/clarity)](LICENSE)
[![Release](https://img.shields.io/github/v/release/LegacyCodeHQ/clarity)](https://github.com/LegacyCodeHQ/clarity/releases)
[![npm version](https://img.shields.io/npm/v/@legacycodehq/clarity)](https://www.npmjs.com/package/@legacycodehq/clarity)
[![Go Report Card](https://goreportcard.com/badge/github.com/LegacyCodeHQ/clarity)](https://goreportcard.com/report/github.com/LegacyCodeHQ/clarity)

Clarity is a software design tool for developers and coding agents.

> **Renamed from Sanity:** If you previously used `sanity`, this is the same project under a new name.

## Quick Start

**Step 1:** Install with npm (cross-platform):

```bash
npm install -g @legacycodehq/clarity
```

Or install on macOS/Linux using Homebrew:

```bash
brew install LegacyCodeHQ/tap/clarity
```

**Step 2:** Inside your project:

```bash
clarity setup # Adds usage instructions to AGENTS.md for your coding agent
```

More install options: [Installation Guide](docs/usage/installation.md).

## Supported Languages

- C
- C++
- C#
- Dart
- Go
- JavaScript
- Java
- Kotlin
- Python
- Ruby
- Rust
- Swift
- TypeScript

## Use Cases

- Build maintainable software
- Understand codebases
- [Audit AI-generated code](https://youtu.be/EqOwJnZSiQs)
- Stabilize and reclaim apps built with AI

## Manual Usage

If you run `clarity setup`, your coding agent will use Clarity automatically from `AGENTS.md`.

If you want to use Clarity manually in your terminal, use `clarity show` commands like the examples below.

### Common Commands

```bash
clarity show                      # Visualize uncommitted changes
clarity show -c HEAD              # Visualize the latest commit
clarity show -c HEAD~3...HEAD     # Visualize a commit range
clarity show -i src,tests         # Build graph from specific files/directories
clarity show -w a.go,b.go         # Show all paths between files
clarity show -f mermaid           # Mermaid output (default is dot)
clarity show -u                   # Generate a shareable visualization URL
clarity languages                 # List supported languages and extensions
```

> Note: For quick viewing and sharing, run `clarity show -u` to generate a visualization URL directly.

### Output Options

- `-f dot`: Graphviz DOT output (default)
- `-f mermaid`: Mermaid flowchart output
- `-u`: Generate a visualization URL for supported formats

### Tips

- Run `clarity show` after every non-trivial code change to review blast radius.
- Use `clarity show -c <commit>` for clean, reproducible review snapshots.

## Clarity in Action

Clarity works with Desktop and IDE coding agents. If you are using a CLI coding agent, the agent can open diagrams in your browser for review.

<p align="center">
  <img src="docs/images/clarity+codex-app.png" alt="Clarity graph in Codex app">
  <small>Clarity shows impacted files and highlights tests in green.</small>
</p>

---

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE).
