package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRepositoryStateSignature_ChangesAfterCommit(t *testing.T) {
	dir := t.TempDir()
	setupGitRepo(t, dir)

	createFile(t, dir, "main.go", "package main\n")
	gitAdd(t, dir, "main.go")
	gitCommit(t, dir, "initial")

	before, err := GetRepositoryStateSignature(dir)
	require.NoError(t, err)

	createFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	gitAdd(t, dir, "main.go")
	gitCommit(t, dir, "second")

	after, err := GetRepositoryStateSignature(dir)
	require.NoError(t, err)

	assert.NotEqual(t, before, after)
}

func TestGetRepositoryStateSignature_ChangesOnStagingTransition(t *testing.T) {
	dir := t.TempDir()
	setupGitRepo(t, dir)

	createFile(t, dir, "main.go", "package main\n")
	gitAdd(t, dir, "main.go")
	gitCommit(t, dir, "initial")

	clean, err := GetRepositoryStateSignature(dir)
	require.NoError(t, err)

	createFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	dirtyUnstaged, err := GetRepositoryStateSignature(dir)
	require.NoError(t, err)

	gitAdd(t, dir, "main.go")
	dirtyStaged, err := GetRepositoryStateSignature(dir)
	require.NoError(t, err)

	assert.NotEqual(t, clean, dirtyUnstaged)
	assert.NotEqual(t, dirtyUnstaged, dirtyStaged)
}
