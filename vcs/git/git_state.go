package git

import (
	"fmt"
	"strings"
)

const unbornHeadSignature = "(unborn)"

// GetRepositoryStateSignature returns a compact signature of repository state.
// It includes HEAD and porcelain status so callers can detect commit/index/worktree transitions.
func GetRepositoryStateSignature(repoPath string) (string, error) {
	head, err := getHEADSignature(repoPath)
	if err != nil {
		return "", err
	}

	status, stderr, err := runGitCommand(repoPath, "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return "", gitCommandError(err, stderr)
	}

	return fmt.Sprintf("%s\n%s", head, strings.TrimSpace(string(status))), nil
}

func getHEADSignature(repoPath string) (string, error) {
	head, stderr, err := runGitCommand(repoPath, "rev-parse", "--verify", "HEAD")
	if err == nil {
		return strings.TrimSpace(string(head)), nil
	}

	if strings.Contains(strings.ToLower(stderr), "needed a single revision") {
		return unbornHeadSignature, nil
	}

	return "", gitCommandError(err, stderr)
}
