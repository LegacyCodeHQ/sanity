package ruby

import (
	"path/filepath"
	"strings"
)

// IsTestFile reports whether the given Ruby path is a test/spec file.
func IsTestFile(filePath string) bool {
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	if ext != ".rb" {
		return false
	}

	if strings.HasSuffix(fileName, "_test.rb") || strings.HasSuffix(fileName, "_spec.rb") || strings.HasPrefix(fileName, "test_") {
		return true
	}

	path := filepath.ToSlash(filePath)
	return strings.Contains(path, "/test/") || strings.Contains(path, "/tests/") || strings.Contains(path, "/spec/")
}
