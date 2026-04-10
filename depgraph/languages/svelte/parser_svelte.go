package svelte

import (
	"context"
	"fmt"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/svelte"

	"github.com/LegacyCodeHQ/clarity/depgraph/languages/javascript"
)

var (
	svelteLanguage   = svelte.GetLanguage()
	svelteParserPool = sync.Pool{
		New: func() any {
			parser := sitter.NewParser()
			parser.SetLanguage(svelteLanguage)
			return parser
		},
	}
)

// ParseSvelteImports parses a Svelte file and extracts JavaScript imports
// from its <script> blocks.
func ParseSvelteImports(sourceCode []byte) ([]javascript.JavaScriptImport, error) {
	parser, _ := svelteParserPool.Get().(*sitter.Parser)
	if parser == nil {
		parser = sitter.NewParser()
		parser.SetLanguage(svelteLanguage)
	}

	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	svelteParserPool.Put(parser)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Svelte code: %w", err)
	}
	defer tree.Close()

	scriptContents := extractScriptContents(tree.RootNode(), sourceCode)

	var allImports []javascript.JavaScriptImport
	for _, script := range scriptContents {
		imports, err := javascript.ParseJavaScriptImports(script, false)
		if err != nil {
			continue
		}
		allImports = append(allImports, imports...)
	}

	return allImports, nil
}

// extractScriptContents walks the Svelte AST and returns the text content
// of each <script> element.
func extractScriptContents(rootNode *sitter.Node, sourceCode []byte) [][]byte {
	var scripts [][]byte

	var walk func(*sitter.Node)
	walk = func(n *sitter.Node) {
		if n == nil {
			return
		}

		if n.Type() == "script_element" {
			if content := extractRawText(n, sourceCode); content != nil {
				scripts = append(scripts, content)
			}
		}

		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}

	walk(rootNode)
	return scripts
}

// extractRawText finds the raw_text child of a script_element node.
func extractRawText(scriptNode *sitter.Node, sourceCode []byte) []byte {
	for i := 0; i < int(scriptNode.ChildCount()); i++ {
		child := scriptNode.Child(i)
		if child != nil && child.Type() == "raw_text" {
			return []byte(child.Content(sourceCode))
		}
	}
	return nil
}
