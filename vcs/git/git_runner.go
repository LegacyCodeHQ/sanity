package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const gitCommandTimeout = 10 * time.Second

func runGitCommand(repoPath string, args ...string) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrText := strings.TrimSpace(stderr.String())
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, stderrText, fmt.Errorf("git command timed out after %s", gitCommandTimeout)
		}
		return nil, stderrText, err
	}

	return stdout.Bytes(), strings.TrimSpace(stderr.String()), nil
}

func gitCommandError(err error, stderr string) error {
	if err == nil {
		return nil
	}
	if stderr != "" {
		return fmt.Errorf("git command failed: %s", stderr)
	}
	return err
}
