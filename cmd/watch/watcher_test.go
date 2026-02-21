package watch

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestAddWatchDirsIgnoresMissingPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated privileges on Windows")
	}

	root := t.TempDir()
	worktree := filepath.Join(root, ".claude", "worktrees", "beautiful-gauss")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}

	linkPath := filepath.Join(worktree, "clarity.project")
	if err := os.Symlink("clarity/AGENTS.md", linkPath); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("new watcher: %v", err)
	}
	defer watcher.Close()

	if err := addWatchDirs(watcher, root); err != nil {
		t.Fatalf("addWatchDirs: %v", err)
	}
}

func TestAddWatchDirsIgnoresMissingDirectoriesFromAdder(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "missing-dir")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.RemoveAll(target); err != nil {
		t.Fatalf("remove target: %v", err)
	}

	adder := func(path string) error {
		if path == target {
			return fs.ErrNotExist
		}
		return nil
	}

	if err := addWatchDirsWithAdder(root, adder); err != nil {
		t.Fatalf("addWatchDirsWithAdder: %v", err)
	}
}

func TestAddWatchDirsSkipsBrokenSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated privileges on Windows")
	}

	root := t.TempDir()
	worktree := filepath.Join(root, ".claude", "worktrees", "beautiful-gauss")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}

	linkPath := filepath.Join(worktree, "clarity.project")
	if err := os.Symlink("clarity/AGENTS.md", linkPath); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	var added []string
	adder := func(path string) error {
		added = append(added, path)
		return nil
	}

	if err := addWatchDirsWithAdder(root, adder); err != nil {
		t.Fatalf("addWatchDirsWithAdder: %v", err)
	}

	for _, path := range added {
		if path == linkPath {
			t.Fatalf("expected broken symlink to be skipped, but was added")
		}
	}
}
