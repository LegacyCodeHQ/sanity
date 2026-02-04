package onboard

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed ONBOARD.md
var onboardingContent string

// Cmd represents the onboard command
var Cmd = &cobra.Command{
	Use:   "onboard",
	Short: "Show how to use sanity with your project",
	Long: `Show how to use sanity with your project.

Provides guidance on integrating sanity into your workflow and
a minimal snippet to add to AGENTS.md for AI assistant awareness.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(onboardingContent)
	},
}
