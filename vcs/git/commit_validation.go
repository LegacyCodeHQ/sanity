package git

// ValidateCommit validates that a commit reference resolves in the given repository.
func ValidateCommit(repoPath, commitID string) error {
	return validateCommit(repoPath, commitID)
}
