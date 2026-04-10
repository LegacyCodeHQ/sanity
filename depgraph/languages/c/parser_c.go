package c

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
)

var (
	cLanguage   = c.GetLanguage()
	cParserPool = sync.Pool{
		New: func() any {
			parser := sitter.NewParser()
			parser.SetLanguage(cLanguage)
			return parser
		},
	}
)

// IncludeKind distinguishes between system and local includes.
type IncludeKind int

const (
	IncludeLocal IncludeKind = iota
	IncludeSystem
)

// Include represents a C include directive.
type Include struct {
	Path string
	Kind IncludeKind
}

// CIncludes parses a C file and returns its includes.
func CIncludes(filePath string) ([]Include, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseCIncludes(sourceCode)
}

// ParseCIncludes parses C source code and extracts includes.
func ParseCIncludes(sourceCode []byte) ([]Include, error) {
	if includes, ok := parseCIncludesFast(sourceCode); ok {
		return includes, nil
	}

	parser, _ := cParserPool.Get().(*sitter.Parser)
	if parser == nil {
		parser = sitter.NewParser()
		parser.SetLanguage(cLanguage)
	}
	defer cParserPool.Put(parser)

	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse C code: %w", err)
	}
	defer tree.Close()

	return extractIncludes(tree.RootNode(), sourceCode), nil
}

// parseCIncludesFast extracts #include directives without using tree-sitter.
// Returns (includes, true) on success, or (nil, false) if parsing is uncertain.
func parseCIncludesFast(src []byte) ([]Include, bool) {
	includes := make([]Include, 0, 8)
	i := 0
	n := len(src)
	atLineStart := true

	for i < n {
		// Block comments /* ... */
		if i+1 < n && src[i] == '/' && src[i+1] == '*' {
			i += 2
			closed := false
			for i+1 < n {
				if src[i] == '*' && src[i+1] == '/' {
					i += 2
					closed = true
					break
				}
				if src[i] == '\n' {
					atLineStart = true
				}
				i++
			}
			if !closed {
				return nil, false
			}
			continue
		}
		// Line comments // ...
		if i+1 < n && src[i] == '/' && src[i+1] == '/' {
			for i < n && src[i] != '\n' {
				i++
			}
			continue
		}
		// String literals
		if src[i] == '"' {
			i++
			for i < n {
				if src[i] == '\\' {
					i += 2
					continue
				}
				if src[i] == '"' {
					i++
					break
				}
				if src[i] == '\n' {
					atLineStart = true
				}
				i++
			}
			atLineStart = false
			continue
		}
		// Char literals
		if src[i] == '\'' {
			i++
			for i < n {
				if src[i] == '\\' {
					i += 2
					continue
				}
				if src[i] == '\'' {
					i++
					break
				}
				i++
			}
			atLineStart = false
			continue
		}
		if src[i] == '\n' {
			atLineStart = true
			i++
			continue
		}
		if atLineStart {
			// Skip horizontal whitespace
			for i < n && (src[i] == ' ' || src[i] == '\t') {
				i++
			}
			if i < n && src[i] == '#' {
				i++
				for i < n && (src[i] == ' ' || src[i] == '\t') {
					i++
				}
				const directive = "include"
				if i+len(directive) <= n && string(src[i:i+len(directive)]) == directive {
					j := i + len(directive)
					for j < n && (src[j] == ' ' || src[j] == '\t') {
						j++
					}
					if j < n {
						switch src[j] {
						case '"':
							end := j + 1
							for end < n && src[end] != '"' && src[end] != '\n' {
								end++
							}
							if end < n && src[end] == '"' {
								includes = append(includes, Include{Path: string(src[j+1 : end]), Kind: IncludeLocal})
							}
						case '<':
							end := j + 1
							for end < n && src[end] != '>' && src[end] != '\n' {
								end++
							}
							if end < n && src[end] == '>' {
								includes = append(includes, Include{Path: string(src[j+1 : end]), Kind: IncludeSystem})
							}
						}
					}
				}
			}
			atLineStart = false
			continue
		}
		i++
	}
	return includes, true
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

// ResolveCIncludePath resolves a C include path to possible file paths.
func ResolveCIncludePath(sourceFile, includePath string, suppliedFiles map[string]bool) []string {
	sourceDir := filepath.Dir(sourceFile)
	basePath := filepath.Join(sourceDir, includePath)
	basePath = filepath.Clean(basePath)

	var resolvedPaths []string

	if filepath.Ext(includePath) != "" {
		if suppliedFiles[basePath] {
			resolvedPaths = append(resolvedPaths, basePath)
		}
		return resolvedPaths
	}

	candidate := basePath + ".h"
	if suppliedFiles[candidate] {
		resolvedPaths = append(resolvedPaths, candidate)
	}

	return resolvedPaths
}
