package diff

import (
	"fmt"
	"strings"

	"github.com/LegacyCodeHQ/clarity/vcs/git"
	"github.com/spf13/cobra"
)

type diffMode string

const (
	diffModeWorkingTree diffMode = "working-tree"
	diffModeCommit      diffMode = "commit"
)

var snapshotSelectorFlags = []string{"staged", "unstaged", "untracked"}

type diffOptions struct {
	repoPath   string
	summary    bool
	commitSpec string
}

type commitComparison struct {
	baseRef   string
	targetRef string
	mode      diffMode
}

// Cmd represents the diff command.
var Cmd = NewCommand()

// NewCommand returns a new diff command instance.
func NewCommand() *cobra.Command {
	opts := &diffOptions{}

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show dependency-graph changes between snapshots.",
		Long:  "Show dependency-graph changes between snapshots.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.repoPath, "repo", "r", "", "Git repository path (default: current directory)")
	cmd.Flags().BoolVar(&opts.summary, "summary", false, "Print text summary only")
	cmd.Flags().StringVarP(&opts.commitSpec, "commit", "c", "", "Compare committed snapshots (<commit> or <A>,<B>)")

	// Reserved snapshot selectors for future working-tree controls.
	cmd.Flags().Bool("staged", false, "Include staged changes")
	_ = cmd.Flags().MarkHidden("staged")
	cmd.Flags().Bool("unstaged", false, "Include unstaged changes")
	_ = cmd.Flags().MarkHidden("unstaged")
	cmd.Flags().Bool("untracked", false, "Include untracked changes")
	_ = cmd.Flags().MarkHidden("untracked")

	return cmd
}

func runDiff(cmd *cobra.Command, opts *diffOptions) error {
	repoPath := opts.repoPath
	if repoPath == "" {
		repoPath = "."
	}

	comparison, err := resolveModeAndCommitComparison(cmd, repoPath, opts.commitSpec)
	if err != nil {
		return err
	}

	_ = comparison
	_ = opts.summary

	return nil
}

func resolveModeAndCommitComparison(cmd *cobra.Command, repoPath, commitSpec string) (commitComparison, error) {
	trimmedCommit := strings.TrimSpace(commitSpec)
	if trimmedCommit == "" {
		return commitComparison{mode: diffModeWorkingTree}, nil
	}

	if err := validateCommitModeConflicts(cmd); err != nil {
		return commitComparison{}, err
	}

	baseRef, targetRef, err := parseCommitSpec(trimmedCommit)
	if err != nil {
		return commitComparison{}, err
	}

	if baseRef == "" {
		if err := git.ValidateCommit(repoPath, targetRef); err != nil {
			return commitComparison{}, err
		}
		return commitComparison{baseRef: "", targetRef: targetRef, mode: diffModeCommit}, nil
	}

	if err := git.ValidateCommit(repoPath, baseRef); err != nil {
		return commitComparison{}, err
	}
	if err := git.ValidateCommit(repoPath, targetRef); err != nil {
		return commitComparison{}, err
	}

	return commitComparison{baseRef: baseRef, targetRef: targetRef, mode: diffModeCommit}, nil
}

func validateCommitModeConflicts(cmd *cobra.Command) error {
	for _, flagName := range snapshotSelectorFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag != nil && flag.Changed {
			return fmt.Errorf("--commit cannot be combined with --%s", flagName)
		}
	}
	return nil
}

func parseCommitSpec(commitSpec string) (baseRef string, targetRef string, err error) {
	if commitSpec == "" {
		return "", "", fmt.Errorf("--commit requires a value")
	}

	commaCount := strings.Count(commitSpec, ",")
	if commaCount == 0 {
		return "", commitSpec, nil
	}
	if commaCount != 1 {
		return "", "", fmt.Errorf("invalid --commit value %q: expected <commit> or <A>,<B>", commitSpec)
	}

	parts := strings.SplitN(commitSpec, ",", 2)
	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])
	if left == "" || right == "" {
		return "", "", fmt.Errorf("invalid --commit value %q: both refs are required in <A>,<B>", commitSpec)
	}

	return left, right, nil
}
