# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Committing Changes

- Do **not** commit changes unless the user explicitly asks for a commit.
- This applies even at session completion.

## Quick Reference

```bash
bd ready                             # Find available work
bd show <id>                         # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>                        # Complete work
bd sync                              # Sync with git
```

## Session Completion

**When ending a work session**, you MUST complete ALL steps below.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds:
   ```bash
   make lint
   make test
   ```
3. **Update issue status** - Close finished work, update in-progress items
   ```bash
   bd sync
   ```
4. **Hand off** - Provide context for next session

## Commit Guidelines

Follow the established history patterns:

- Use `type: subject` in lowercase with a short, imperative subject line.
- Keep subjects concise (ideally under ~72 characters).
- One logical change per commit.
- Prefer these types (use only when accurate):
  - `feat`: user-visible feature
  - `fix`: bug fix or correctness issue
  - `lint`: lint-only fixes and style-rule compliance
  - `refactor`: behavior-preserving code restructuring
  - `test`: tests or fixtures only
  - `docs`: documentation only
  - `build`: build system or tooling changes
  - `ci`: CI pipeline changes
  - `chore`: non-functional maintenance (e.g., syncs)
  - `style`: formatting-only changes (non-lint)
  - `human`: human-led or non-mechanical changes (use sparingly)
