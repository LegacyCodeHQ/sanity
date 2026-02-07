package ruby

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RubyImport represents a require/require_relative in a Ruby file.
type RubyImport struct {
	path       string
	isRelative bool
}

func (i RubyImport) Path() string {
	return i.path
}

func (i RubyImport) IsRelative() bool {
	return i.isRelative
}

// RubyImports parses a Ruby file and returns its imports.
func RubyImports(filePath string) ([]RubyImport, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseRubyImports(sourceCode)
}

// ParseRubyImports parses Ruby source code and extracts require directives.
func ParseRubyImports(sourceCode []byte) ([]RubyImport, error) {
	var imports []RubyImport

	scanner := bufio.NewScanner(bytes.NewReader(sourceCode))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if imp, ok := parseRubyImportLine(line, "require_relative", true); ok {
			imports = append(imports, imp)
			continue
		}

		if imp, ok := parseRubyImportLine(line, "require", false); ok {
			imports = append(imports, imp)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed scanning Ruby source: %w", err)
	}

	return imports, nil
}

func parseRubyImportLine(line, keyword string, isRelative bool) (RubyImport, bool) {
	if !strings.HasPrefix(line, keyword) {
		return RubyImport{}, false
	}

	rest := strings.TrimSpace(strings.TrimPrefix(line, keyword))
	rest = strings.TrimPrefix(rest, "(")
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return RubyImport{}, false
	}

	quote := rest[0]
	if quote != '\'' && quote != '"' {
		return RubyImport{}, false
	}

	end := strings.IndexByte(rest[1:], quote)
	if end < 0 {
		return RubyImport{}, false
	}

	path := strings.TrimSpace(rest[1 : end+1])
	if path == "" {
		return RubyImport{}, false
	}

	return RubyImport{path: path, isRelative: isRelative}, true
}

// ResolveRubyImportPath resolves a Ruby require path to possible file paths.
func ResolveRubyImportPath(sourceFile string, imp RubyImport, suppliedFiles map[string]bool) []string {
	if imp.IsRelative() {
		return resolveRelativeRubyImportPath(sourceFile, imp.Path(), suppliedFiles)
	}
	return resolveAbsoluteRubyImportPath(imp.Path(), suppliedFiles)
}

func resolveRelativeRubyImportPath(sourceFile, importPath string, suppliedFiles map[string]bool) []string {
	sourceDir := filepath.Dir(sourceFile)
	basePath := filepath.Clean(filepath.Join(sourceDir, importPath))

	candidates := rubyImportCandidates(basePath)
	return existingPaths(candidates, suppliedFiles)
}

func resolveAbsoluteRubyImportPath(importPath string, suppliedFiles map[string]bool) []string {
	cleanPath := strings.TrimSpace(filepath.ToSlash(importPath))
	if cleanPath == "" {
		return []string{}
	}

	withoutExt := strings.TrimSuffix(cleanPath, filepath.Ext(cleanPath))
	if withoutExt == "" {
		withoutExt = cleanPath
	}

	var resolved []string
	for filePath, exists := range suppliedFiles {
		if !exists {
			continue
		}
		if filepath.Ext(filePath) != ".rb" {
			continue
		}

		normalizedFile := filepath.ToSlash(filePath)
		if strings.HasSuffix(normalizedFile, "/"+withoutExt+".rb") ||
			strings.HasSuffix(normalizedFile, "/"+cleanPath) {
			resolved = append(resolved, filePath)
		}
	}

	sort.Strings(resolved)
	return resolved
}

func rubyImportCandidates(basePath string) []string {
	if filepath.Ext(basePath) == ".rb" {
		return []string{basePath}
	}

	return []string{
		basePath + ".rb",
		filepath.Join(basePath, "init.rb"),
	}
}

func existingPaths(candidates []string, suppliedFiles map[string]bool) []string {
	resolved := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if suppliedFiles[candidate] {
			resolved = append(resolved, candidate)
		}
	}
	return resolved
}
