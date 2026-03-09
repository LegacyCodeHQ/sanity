package rust

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/rust"
)

var (
	rustLanguage = rust.GetLanguage()
	rustParserPool = sync.Pool{
		New: func() any {
			parser := sitter.NewParser()
			parser.SetLanguage(rustLanguage)
			return parser
		},
	}
)

// RustImportKind describes the type of Rust import-like declaration.
type RustImportKind int

const (
	RustImportUse RustImportKind = iota
	RustImportExternCrate
	RustImportModDecl
)

// RustImport represents a Rust import statement or module declaration.
type RustImport struct {
	Path string
	Kind RustImportKind
}

// RustImports parses a Rust file and returns its imports.
func RustImports(filePath string) ([]RustImport, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseRustImports(sourceCode)
}

// ParseRustImports parses Rust source code and extracts imports.
func ParseRustImports(sourceCode []byte) ([]RustImport, error) {
	if os.Getenv("CLARITY_RUST_IMPORTS_PARSER") != "tree" {
		imports, _ := parseRustImportsFast(sourceCode)
		return imports, nil
	}

	parser, _ := rustParserPool.Get().(*sitter.Parser)
	if parser == nil {
		parser = sitter.NewParser()
		parser.SetLanguage(rustLanguage)
	}
	defer rustParserPool.Put(parser)

	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Rust code: %w", err)
	}
	defer tree.Close()

	return extractImports(tree.RootNode(), sourceCode), nil
}

func parseRustImportsFast(sourceCode []byte) ([]RustImport, bool) {
	imports := make([]RustImport, 0, 8)
	var stmt []byte

	depth := 0
	inLineComment := false
	inBlockComment := 0
	inString := false
	inChar := false
	escaped := false

	for i := 0; i < len(sourceCode); i++ {
		c := sourceCode[i]
		next := byte(0)
		if i+1 < len(sourceCode) {
			next = sourceCode[i+1]
		}

		if inLineComment {
			if c == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment > 0 {
			if c == '/' && next == '*' {
				inBlockComment++
				i++
				continue
			}
			if c == '*' && next == '/' {
				inBlockComment--
				i++
			}
			continue
		}
		if inString {
			if escaped {
				escaped = false
				continue
			}
			switch c {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}
			continue
		}
		if inChar {
			if escaped {
				escaped = false
				continue
			}
			switch c {
			case '\\':
				escaped = true
			case '\'':
				inChar = false
			}
			continue
		}

		if c == '/' && next == '/' {
			inLineComment = true
			i++
			continue
		}
		if c == '/' && next == '*' {
			inBlockComment = 1
			i++
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == '\'' {
			inChar = true
			continue
		}

		switch c {
		case '{':
			if depth == 0 && !isLikelyUsePrefix(stmt) {
				stmt = stmt[:0]
			}
			depth++
			continue
		case '}':
			if depth > 0 {
				depth--
			}
			continue
		}

		if depth > 0 {
			continue
		}

		if c == ';' {
			if imp, ok := parseTopLevelRustImportStatementBytes(stmt); ok {
				imports = append(imports, imp)
			}
			stmt = stmt[:0]
			continue
		}
		if len(stmt) == 0 && (c == ' ' || c == '\t' || c == '\n' || c == '\r') {
			continue
		}
		stmt = append(stmt, c)
	}

	if inString || inChar || inBlockComment > 0 {
		return nil, false
	}

	return imports, true
}

func parseTopLevelRustImportStatementBytes(stmt []byte) (RustImport, bool) {
	s := trimSpaceBytes(stmt)
	if len(s) == 0 {
		return RustImport{}, false
	}
	s = stripLeadingRustAttributesBytes(s)
	if len(s) == 0 {
		return RustImport{}, false
	}
	s = stripRustVisibilityPrefixBytes(s)

	switch {
	case bytes.HasPrefix(s, []byte("use ")):
		path := normalizeUsePathBytes(trimSpaceBytes(s[len("use "):]))
		if len(path) == 0 {
			return RustImport{}, false
		}
		return RustImport{Path: string(path), Kind: RustImportUse}, true
	case bytes.HasPrefix(s, []byte("extern crate ")):
		name := leadingRustIdentBytes(trimSpaceBytes(s[len("extern crate "):]))
		if len(name) == 0 {
			return RustImport{}, false
		}
		return RustImport{Path: string(name), Kind: RustImportExternCrate}, true
	case bytes.HasPrefix(s, []byte("mod ")):
		name := leadingRustIdentBytes(trimSpaceBytes(s[len("mod "):]))
		if len(name) == 0 {
			return RustImport{}, false
		}
		return RustImport{Path: string(name), Kind: RustImportModDecl}, true
	default:
		return RustImport{}, false
	}
}

func trimSpaceBytes(b []byte) []byte {
	return bytes.TrimSpace(b)
}

func stripLeadingRustAttributesBytes(s []byte) []byte {
	trimmed := trimSpaceBytes(s)
	for bytes.HasPrefix(trimmed, []byte("#[")) || bytes.HasPrefix(trimmed, []byte("#![")) {
		open := bytes.IndexByte(trimmed, '[')
		if open < 0 {
			return trimmed
		}
		level := 0
		end := -1
		for i := open; i < len(trimmed); i++ {
			switch trimmed[i] {
			case '[':
				level++
			case ']':
				level--
				if level == 0 {
					end = i
					break
				}
			}
		}
		if end < 0 {
			return trimmed
		}
		trimmed = trimSpaceBytes(trimmed[end+1:])
	}
	return trimmed
}

func stripRustVisibilityPrefixBytes(s []byte) []byte {
	trimmed := trimSpaceBytes(s)
	if bytes.HasPrefix(trimmed, []byte("pub ")) {
		return trimSpaceBytes(trimmed[len("pub "):])
	}
	if bytes.HasPrefix(trimmed, []byte("pub(")) {
		if idx := bytes.IndexByte(trimmed, ')'); idx >= 0 {
			return trimSpaceBytes(trimmed[idx+1:])
		}
	}
	return trimmed
}

func normalizeUsePathBytes(expr []byte) []byte {
	path := trimSpaceBytes(expr)
	if len(path) == 0 {
		return nil
	}
	if idx := bytes.Index(path, []byte(" as ")); idx >= 0 {
		path = trimSpaceBytes(path[:idx])
	}
	if idx := bytes.IndexByte(path, '{'); idx >= 0 {
		path = trimSpaceBytes(path[:idx])
	}
	for bytes.HasSuffix(path, []byte("::")) {
		path = trimSpaceBytes(path[:len(path)-2])
	}
	path = bytes.TrimPrefix(path, []byte("::"))
	path = trimSpaceBytes(path)
	if len(path) == 0 {
		return nil
	}
	return path
}

func leadingRustIdentBytes(s []byte) []byte {
	if len(s) == 0 {
		return nil
	}
	i := 0
	for i < len(s) {
		c := s[i]
		if c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (i > 0 && c >= '0' && c <= '9') {
			i++
			continue
		}
		break
	}
	if i == 0 {
		return nil
	}
	return s[:i]
}

func isLikelyUsePrefix(stmt []byte) bool {
	return bytes.Contains(stmt, []byte("use "))
}

func parseTopLevelRustImportStatement(stmt string) (RustImport, bool) {
	s := strings.TrimSpace(stmt)
	if s == "" {
		return RustImport{}, false
	}
	s = stripLeadingRustAttributes(s)
	if s == "" {
		return RustImport{}, false
	}

	s = stripRustVisibilityPrefix(s)
	switch {
	case strings.HasPrefix(s, "use "):
		path := normalizeUsePath(strings.TrimSpace(strings.TrimPrefix(s, "use ")))
		if path == "" {
			return RustImport{}, false
		}
		return RustImport{Path: path, Kind: RustImportUse}, true
	case strings.HasPrefix(s, "extern crate "):
		name := leadingRustIdent(strings.TrimSpace(strings.TrimPrefix(s, "extern crate ")))
		if name == "" {
			return RustImport{}, false
		}
		return RustImport{Path: name, Kind: RustImportExternCrate}, true
	case strings.HasPrefix(s, "mod "):
		name := leadingRustIdent(strings.TrimSpace(strings.TrimPrefix(s, "mod ")))
		if name == "" {
			return RustImport{}, false
		}
		return RustImport{Path: name, Kind: RustImportModDecl}, true
	default:
		return RustImport{}, false
	}
}

func stripLeadingRustAttributes(s string) string {
	trimmed := strings.TrimSpace(s)
	for strings.HasPrefix(trimmed, "#[") || strings.HasPrefix(trimmed, "#![") {
		open := strings.Index(trimmed, "[")
		if open < 0 {
			return trimmed
		}
		level := 0
		end := -1
		for i := open; i < len(trimmed); i++ {
			switch trimmed[i] {
			case '[':
				level++
			case ']':
				level--
				if level == 0 {
					end = i
					break
				}
			}
		}
		if end < 0 {
			return trimmed
		}
		trimmed = strings.TrimSpace(trimmed[end+1:])
	}
	return trimmed
}

func stripRustVisibilityPrefix(s string) string {
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "pub ") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "pub "))
	}
	if strings.HasPrefix(trimmed, "pub(") {
		if idx := strings.Index(trimmed, ")"); idx >= 0 {
			return strings.TrimSpace(trimmed[idx+1:])
		}
	}
	return trimmed
}

func normalizeUsePath(expr string) string {
	path := strings.TrimSpace(expr)
	if path == "" {
		return ""
	}
	if idx := strings.Index(path, " as "); idx >= 0 {
		path = strings.TrimSpace(path[:idx])
	}
	if idx := strings.Index(path, "{"); idx >= 0 {
		path = strings.TrimSpace(path[:idx])
	}
	for strings.HasSuffix(path, "::") {
		path = strings.TrimSuffix(path, "::")
		path = strings.TrimSpace(path)
	}
	path = strings.TrimPrefix(path, "::")
	return strings.TrimSpace(path)
}

func leadingRustIdent(s string) string {
	if s == "" {
		return ""
	}
	i := 0
	for i < len(s) {
		c := s[i]
		if c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (i > 0 && c >= '0' && c <= '9') {
			i++
			continue
		}
		break
	}
	if i == 0 {
		return ""
	}
	return s[:i]
}

func extractImports(rootNode *sitter.Node, sourceCode []byte) []RustImport {
	if rootNode == nil {
		return nil
	}

	// Rust imports/declarations that affect module dependencies live at file scope.
	// Restricting to top-level declarations avoids a full-tree walk and reduces cgo traversal overhead.
	childCount := int(rootNode.NamedChildCount())
	imports := make([]RustImport, 0, childCount)
	for i := 0; i < childCount; i++ {
		n := rootNode.NamedChild(i)
		if n == nil {
			continue
		}
		switch n.Type() {
		case "use_declaration":
			if path := extractUsePath(n, sourceCode); path != "" {
				imports = append(imports, RustImport{Path: path, Kind: RustImportUse})
			}
		case "extern_crate_declaration":
			if crate := extractExternCrate(n, sourceCode); crate != "" {
				imports = append(imports, RustImport{Path: crate, Kind: RustImportExternCrate})
			}
		case "mod_item":
			if modName := extractModDecl(n, sourceCode); modName != "" {
				imports = append(imports, RustImport{Path: modName, Kind: RustImportModDecl})
			}
		}
	}
	return imports
}

func extractUsePath(node *sitter.Node, sourceCode []byte) string {
	if node == nil {
		return ""
	}

	arg := node.ChildByFieldName("argument")
	if arg == nil {
		return ""
	}

	switch arg.Type() {
	case "use_as_clause", "scoped_use_list":
		if path := arg.ChildByFieldName("path"); path != nil {
			return path.Content(sourceCode)
		}
	case "scoped_identifier", "identifier", "crate", "self", "super":
		return arg.Content(sourceCode)
	}

	return arg.Content(sourceCode)
}

func namedChildCount(node *sitter.Node) int {
	if node == nil {
		return 0
	}
	return int(node.NamedChildCount())
}

func extractExternCrate(node *sitter.Node, sourceCode []byte) string {
	if node == nil {
		return ""
	}
	if nameNode := node.ChildByFieldName("name"); nameNode != nil {
		return nameNode.Content(sourceCode)
	}
	childCount := namedChildCount(node)
	for i := 0; i < childCount; i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		if child.Type() == "identifier" {
			return child.Content(sourceCode)
		}
	}
	return ""
}

func extractModDecl(node *sitter.Node, sourceCode []byte) string {
	if node == nil {
		return ""
	}
	if modItemHasBody(node) {
		return ""
	}

	if nameNode := node.ChildByFieldName("name"); nameNode != nil {
		return nameNode.Content(sourceCode)
	}
	childCount := namedChildCount(node)
	for i := 0; i < childCount; i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		if child.Type() == "identifier" {
			return child.Content(sourceCode)
		}
	}
	return ""
}

func modItemHasBody(node *sitter.Node) bool {
	if node == nil {
		return false
	}
	if body := node.ChildByFieldName("body"); body != nil {
		return true
	}
	childCount := namedChildCount(node)
	for i := 0; i < childCount; i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		switch child.Type() {
		case "block", "declaration_list":
			return true
		}
	}
	return false
}
