package vcs

import (
	"os"
	"path/filepath"
)

// ContentReader is a function that reads file content given a file path.
// This allows the caller to control how files are read (filesystem, git, etc.)
type ContentReader func(filePath string) ([]byte, error)

// FilesystemContentReader returns a ContentReader that reads from the filesystem.
func FilesystemContentReader() ContentReader {
	return func(absPath string) ([]byte, error) {
		return os.ReadFile(absPath)
	}
}

// GitCommitContentReader returns a ContentReader that reads file content from a specific git commit.
func GitCommitContentReader(repoPath, commitID string) ContentReader {
	return func(absPath string) ([]byte, error) {
		relPath := getRelativePath(absPath, repoPath)
		return GetFileContentFromCommit(repoPath, commitID, relPath)
	}
}

// GetRelativePath converts an absolute file path to a path relative to the repository root
func getRelativePath(absPath, repoPath string) string {
	// Get absolute repository path
	absRepoPath, err := filepath.Abs(repoPath)
	if err != nil {
		// If we can't get absolute path, try relative path as-is
		relPath, err := filepath.Rel(repoPath, absPath)
		if err != nil {
			// Fallback to using the absolute path
			return absPath
		}
		return relPath
	}

	// Get path relative to repository root
	relPath, err := filepath.Rel(absRepoPath, absPath)
	if err != nil {
		// Fallback to using the absolute path
		return absPath
	}

	return relPath
}
