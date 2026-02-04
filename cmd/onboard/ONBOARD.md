Sanity Onboarding

Add this minimal snippet to AGENTS.md (or create it):

--- BEGIN AGENTS.MD CONTENT ---
## Dependency Visualization

This project uses **sanity** for dependency visualization and design analysis.
Run `sanity prime` for workflow context.

**Quick reference:**
- `sanity graph -u`          - Visualize and open in browser
- `sanity graph -f mermaid`  - Output mermaid diagram
- `sanity graph -c HEAD~3`   - Graph files from recent commits

For full workflow details: `sanity prime`
--- END AGENTS.MD CONTENT ---

For GitHub Copilot users:
Add the same content to .github/copilot-instructions.md

How it works:
  - sanity prime provides full workflow context (~100 lines)
  - AGENTS.md only needs this minimal pointer, not full instructions

This keeps AGENTS.md lean while sanity prime provides up-to-date workflow details.
