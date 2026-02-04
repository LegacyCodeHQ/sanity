package prime

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed PRIME.md
var primeContent string

// Cmd represents the prime command
var Cmd = &cobra.Command{
	Use:   "prime",
	Short: "Output AI-optimized workflow context",
	Long: `Output essential Sanity workflow context in AI-optimized markdown format.

Designed for AI agents to get full workflow details and command reference.
Run this after context compaction, clear, or at the start of a new session.

How it works:
  - sanity prime provides full workflow context (~100 lines)
  - sanity onboard provides a minimal snippet to add to AGENTS.md
  - AGENTS.md only needs the minimal pointer, not full instructions`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(primeContent)
	},
}
