package rust

import (
	"context"
	"fmt"
	"os"
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
