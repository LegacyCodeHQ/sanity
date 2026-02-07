package cpp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/cpp"
)

// IncludeKind distinguishes between system and local includes.
type IncludeKind int

const (
	IncludeLocal IncludeKind = iota
	IncludeSystem
)

// Include represents a C++ include directive.
type Include struct {
	Path string
	Kind IncludeKind
}

// CppIncludes parses a C++ file and returns its includes.
func CppIncludes(filePath string) ([]Include, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseCppIncludes(sourceCode)
}

// ParseCppIncludes parses C++ source code and extracts includes.
func ParseCppIncludes(sourceCode []byte) ([]Include, error) {
	lang := cpp.GetLanguage()

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse C++ code: %w", err)
	}
	defer tree.Close()

	return extractIncludes(tree.RootNode(), sourceCode), nil
}

func extractIncludes(rootNode *sitter.Node, sourceCode []byte) []Include {
	var includes []Include

	var walk func(*sitter.Node)
	walk = func(n *sitter.Node) {
		if n == nil {
			return
		}

		if n.Type() == "preproc_include" {
			if inc := extractIncludeFromNode(n, sourceCode); inc.Path != "" {
				includes = append(includes, inc)
			}
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}

	walk(rootNode)
	return includes
}

func extractIncludeFromNode(node *sitter.Node, sourceCode []byte) Include {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Type() {
		case "string_literal":
			return Include{Path: cleanStringLiteral(child.Content(sourceCode)), Kind: IncludeLocal}
		case "system_lib_string":
			return Include{Path: cleanSystemInclude(child.Content(sourceCode)), Kind: IncludeSystem}
		}
	}

	return Include{}
}

func cleanStringLiteral(raw string) string {
	return strings.Trim(raw, "\"' ")
}

func cleanSystemInclude(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "<")
	trimmed = strings.TrimSuffix(trimmed, ">")
	return strings.TrimSpace(trimmed)
}

// ResolveCppIncludePath resolves a C++ include path to possible file paths.
func ResolveCppIncludePath(sourceFile, includePath string, suppliedFiles map[string]bool) []string {
	sourceDir := filepath.Dir(sourceFile)
	baseCandidates := includeBasePathCandidates(sourceDir, includePath)
	seen := make(map[string]bool)

	var resolvedPaths []string
	appendResolved := func(path string) {
		if !seen[path] && suppliedFiles[path] {
			seen[path] = true
			resolvedPaths = append(resolvedPaths, path)
		}
	}

	if filepath.Ext(includePath) != "" {
		for _, base := range baseCandidates {
			appendResolved(base)
		}
		sort.Strings(resolvedPaths)
		return resolvedPaths
	}

	extensions := []string{".h", ".hpp", ".hh", ".hxx"}
	for _, base := range baseCandidates {
		for _, ext := range extensions {
			appendResolved(base + ext)
		}
	}

	sort.Strings(resolvedPaths)
	return resolvedPaths
}

func includeBasePathCandidates(sourceDir, includePath string) []string {
	candidates := make(map[string]bool)
	add := func(p string) {
		candidates[filepath.Clean(p)] = true
	}

	// Always try source-relative first.
	add(filepath.Join(sourceDir, includePath))

	// Also try ancestor roots and common include roots.
	for dir := sourceDir; ; dir = filepath.Dir(dir) {
		add(filepath.Join(dir, includePath))
		add(filepath.Join(dir, "include", includePath))

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	out := make([]string, 0, len(candidates))
	for p := range candidates {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}
