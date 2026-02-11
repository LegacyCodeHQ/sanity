# `clarity diff` Notes

This document captures design notes for a proposed `clarity diff` subcommand.

## Command

```bash
clarity diff
```

```bash
clarity diff --summary
```

```bash
clarity diff --commit <commit>
```

```bash
clarity diff --commit <A>,<B>
```

## Meaning

Default behavior compares the current working tree against `HEAD` and reports dependency graph changes.

In scope:
- Staged changes
- Unstaged tracked-file changes
- Untracked files (when they are parseable source files)

Out of scope:
- Ignored files (for example, `.gitignore` matches)

Terminology:
- `node` means a file-level node (same as `clarity show`, not symbol-level nodes).

## Snapshot model

`clarity diff` compares two snapshots without changing the current checkout:
- Base snapshot: repository state at `HEAD` (read from git objects, not via `git checkout`)
- Current snapshot: on-disk working copy state (including staged/unstaged/untracked-in-scope files)

Current snapshot precedence:
- If a file has both staged and unstaged changes, use the working tree version for `clarity diff`.
- Staged-only deletions/new files are still reflected through file existence in the current snapshot.

## Expected behavior for `clarity diff`

1. Resolve `HEAD` as the base snapshot.
2. Build a dependency graph for base (`HEAD`) by reading content from git objects.
3. Build a dependency graph for current workspace state (index + working tree + relevant untracked files).
4. Compute graph delta:
   - Nodes added and removed
   - Edges added and removed
   - Optional semantic findings (new cycles, layer violations)
5. Render delta in selected format (`dot` by default, optional `mermaid` later).
6. Exit with:
   - `0` on success with or without differences
   - non-zero on invalid git state or graph construction failure

## Parser and formatter failure behavior

Follow `clarity show` behavior:
- Unsupported file extensions are still included as file nodes with no dependencies.
- Parser/dependency-resolution failures are hard errors (command returns non-zero).
- Formatter failures are hard errors (command returns non-zero).

## Output contract

Default (`clarity diff`):
- Emit only delta graph output (no human-readable summary prelude).
- Output is formatter-native content (for example DOT or Mermaid).

Summary mode (`clarity diff --summary`):
- Emit human-readable summary only (counts and key findings), intended for quick scans/CI logs.
- Example categories:
  - Nodes added/removed
  - Edges added/removed
  - Optional semantic findings (new cycles, layer violations)
- Format interaction:
  - `--summary` ignores graph renderer output (`dot`/`mermaid`).
  - Summary output is plain text only.

Deterministic ordering for `--summary`:
- Categories are always printed in this order:
  1. Nodes added
  2. Nodes removed
  3. Edges added
  4. Edges removed
  5. Semantic findings (if enabled/present)
- Within each category, entries are sorted lexicographically by stable canonical path/key.
- Count-only lines are always emitted in the same category order, even when counts are zero.

## User-facing intent

`clarity diff` should answer: "How does my current uncommitted work change the dependency graph compared with `HEAD`?"

It is the graph-aware equivalent of `git diff` for architecture/design impact.

## File lifecycle semantics (v1)

Deleted files:
- If a file exists in base (`HEAD`) but not in current snapshot, mark it as a removed node.
- Mark all edges touching that node as removed edges.

New files:
- If a file exists in current snapshot but not in base (`HEAD`), mark it as an added node.
- Mark all edges touching that node as added edges.

Renames:
- Rename detection is deferred.
- In v1, a rename is represented as `removed old path + added new path`.

## Commit mode (`--commit`)

Purpose:
- Compare committed snapshots instead of working tree state.

Examples:
- `clarity diff --commit HEAD`
- `clarity diff --commit <commit>`
- `clarity diff --commit <A>,<B>`

Comparison semantics:
- Single commit (`--commit <commit>`):
  - Base snapshot: `<commit>^`
  - Target snapshot: `<commit>`
- Two-commit compare (`--commit <A>,<B>`):
  - Base snapshot: `<A>`
  - Target snapshot: `<B>`

Output behavior:
- `clarity diff --commit ...` emits graph delta only.
- `clarity diff --commit ... --summary` emits summary only.

Validation:
- `--commit` requires a value.
- Error on invalid refs or malformed two-commit values.
- For two-commit mode, value must be exactly two refs separated by a single comma.
- For two-commit mode, both refs must be non-empty after trimming surrounding whitespace.
- Reject commit values containing extra commas, empty refs, or leading/trailing comma.
- Validate refs with git revision verification before graph work begins; fail fast on invalid refs.
- If the selected single commit has no parent (root commit), compare empty graph -> commit graph.
- Single-commit mode uses first-parent semantics (`<commit>^1`).
- `--commit` mode bypasses working-tree state.
- In `--commit` mode, untracked/staged/unstaged working-tree changes are ignored entirely.

Merge commit note:
- For single-commit mode on merge commits, v1 compares `<commit>^1` -> `<commit>`.
- Alternate parent-diff strategies are deferred for a follow-up design.

## `--commit` precedence and conflict rules

Rules:
- `--commit` is a mode switch.
- When `--commit` is present, working-tree mode is bypassed entirely.
- `--commit` is mutually exclusive with any present or future snapshot-selector flags.

| Scenario                          | Example                               | Mode         | Expected behavior                                      | Valid? |
|-----------------------------------|---------------------------------------|--------------|--------------------------------------------------------|--------|
| Default working tree diff         | `clarity diff`                        | Working tree | Compare `HEAD` vs current working copy                 | Yes    |
| Working tree summary              | `clarity diff --summary`              | Working tree | Same as above, but text summary only                   | Yes    |
| Single commit diff                | `clarity diff --commit HEAD`          | Commit       | Compare `HEAD^` vs `HEAD`                              | Yes    |
| Single commit by SHA              | `clarity diff --commit a1b2c3d`       | Commit       | Compare `<sha>^` vs `<sha>`                            | Yes    |
| Two commit compare                | `clarity diff --commit A,B`           | Commit       | Compare `A` vs `B`                                     | Yes    |
| Two commit summary                | `clarity diff --commit A,B --summary` | Commit       | Same compare, summary text only                        | Yes    |
| Working tree dirty + commit mode  | `clarity diff --commit A,B`           | Commit       | Ignore untracked/staged/unstaged; compare `A`/`B` only | Yes    |
| Malformed commit pair (too many)  | `clarity diff --commit A,B,C`         | Commit       | Reject; exactly two refs max                           | No     |
| Malformed commit pair (empty ref) | `clarity diff --commit A,`            | Commit       | Reject; both refs required                             | No     |
| Empty commit value                | `clarity diff --commit`               | N/A          | Cobra missing value error                              | No     |
| Invalid ref                       | `clarity diff --commit not-a-ref`     | Commit       | Reject with invalid ref error                          | No     |
| Root commit single mode           | `clarity diff --commit <root>`        | Commit       | Compare empty graph vs root commit graph               | Yes    |
| Single merge commit               | `clarity diff --commit <merge>`       | Commit       | Compare `<merge>^1` vs `<merge>`                       | Yes    |

## Open questions

- Deleted-file handling follow-up: should removed nodes always be rendered, or be suppressible in compact output while keeping removed-edge counts?
- New-file handling follow-up: should added nodes always be rendered, or be suppressible in compact output while keeping added-edge counts?
- Should untracked files be included by default or behind a flag?
- Should exit code support policy flags (example: fail when new cycles are introduced)?
