Sanity Onboarding

Add this minimal snippet to AGENTS.md (or create it):

--- BEGIN AGENTS.MD CONTENT ---
---
## Sanity Usage

This project uses **sanity** for the following:

1. Auditing AI-generated code
2. Viewing, identifying, and fixing software design

### Agent Workflow: Using the `graph` Subcommand

After generating code or making changes, agents **MUST** run `sanity graph` on uncommitted changes to audit the modifications.

#### Implementation Based on Agent Environment:

**For shell/terminal agents (macOS):**
```bash
open $(sanity graph -u)  # Or equivalent command on other shells/operating systems
```

This generates a visualization URL and opens it directly in the default browser.

**For IDE/desktop agents with mermaid rendering support:**
```bash
sanity graph -f mermaid
```

Render the mermaid diagram output directly within the IDE or desktop application interface.

#### When to Run:

- After completing any code generation task
- Before committing changes
- When requested to audit or review changes
---
--- END AGENTS.MD CONTENT ---
