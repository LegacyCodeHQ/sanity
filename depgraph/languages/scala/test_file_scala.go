package scala

import (
	"path/filepath"
	"strings"
)

// IsTestFile reports whether the given Scala file path is a test file.
func IsTestFile(filePath string) bool {
	base := filepath.Base(filePath)
	if filepath.Ext(base) != ".scala" {
		return false
	}

	if strings.HasSuffix(base, "Test.scala") || strings.HasSuffix(base, "Tests.scala") {
		return true
	}

	slashed := filepath.ToSlash(filePath)
	return strings.Contains(slashed, "/src/test/") || strings.Contains(slashed, "/test/")
}
