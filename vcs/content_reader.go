package vcs

// ContentReader is a function that reads file content given a file path.
// This allows the caller to control how files are read (filesystem, git, etc.)
type ContentReader func(filePath string) ([]byte, error)
