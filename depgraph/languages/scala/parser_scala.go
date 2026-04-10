package scala

import (
	"context"
	"strings"
	"sync"
	"unicode"

	sitter "github.com/smacker/go-tree-sitter"
	tsscale "github.com/smacker/go-tree-sitter/scala"
)

var (
	scalaLanguage   = tsscale.GetLanguage()
	scalaParserPool = sync.Pool{
		New: func() any {
			parser := sitter.NewParser()
			parser.SetLanguage(scalaLanguage)
			return parser
		},
	}
)

// ScalaImport represents an import in Scala source code.
type ScalaImport interface {
	Path() string
	IsWildcard() bool
	Package() string
}

// StandardLibraryImport represents a Scala/JDK standard library import.
type StandardLibraryImport struct {
	path       string
	isWildcard bool
}

func (s StandardLibraryImport) Path() string {
	return s.path
}

func (s StandardLibraryImport) IsWildcard() bool {
	return s.isWildcard
}

func (s StandardLibraryImport) Package() string {
	return scalaImportPackage(s.path)
}

// ExternalImport represents a third-party import.
type ExternalImport struct {
	path       string
	isWildcard bool
}

func (e ExternalImport) Path() string {
	return e.path
}

func (e ExternalImport) IsWildcard() bool {
	return e.isWildcard
}

func (e ExternalImport) Package() string {
	return scalaImportPackage(e.path)
}

// InternalImport represents an internal project import.
type InternalImport struct {
	path       string
	isWildcard bool
}

func (i InternalImport) Path() string {
	return i.path
}

func (i InternalImport) IsWildcard() bool {
	return i.isWildcard
}

func (i InternalImport) Package() string {
	return scalaImportPackage(i.path)
}

var scalaTopLevelDeclarationTypes = map[string]bool{
	"class_definition":  true,
	"trait_definition":  true,
	"object_definition": true,
	"enum_definition":   true,
}

const packageObjectTypeName = "__scala_package_object__"

// ParsePackageDeclaration extracts the Scala package from source code.
func ParsePackageDeclaration(sourceCode []byte) string {
	tree, err := parseScala(sourceCode)
	if err != nil {
		return ""
	}
	defer tree.Close()

	root := tree.RootNode()
	parts := []string{}
	seenPackageClause := false
	packageObjectName := ""
	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		if child == nil {
			continue
		}

		if child.Type() == "comment" {
			// Allow leading/mid package doc comments.
			continue
		}

		// Scala allows split package declarations:
		//   package a
		//   package b
		// which should resolve to package a.b
		if child.Type() == "package_object" {
			if name := findFirstChildOfType(child, "identifier"); name != nil {
				packageObjectName = strings.TrimSpace(name.Content(sourceCode))
			}
			break
		}

		if child.Type() != "package_clause" {
			if !seenPackageClause {
				continue
			}
			break
		}
		seenPackageClause = true

		pkg := findFirstNodeOfType(child, "package_identifier")
		if pkg == nil {
			continue
		}
		content := strings.TrimSpace(pkg.Content(sourceCode))
		if content == "" {
			continue
		}
		parts = append(parts, content)
	}

	if packageObjectName != "" {
		if len(parts) > 0 {
			parts = append(parts, packageObjectName)
			return strings.Join(parts, ".")
		}
		return packageObjectName
	}

	if len(parts) > 0 {
		return strings.Join(parts, ".")
	}

	node := findFirstNodeOfType(root, "package_clause")
	if node == nil {
		return ""
	}

	pkg := findFirstNodeOfType(node, "package_identifier")
	if pkg == nil {
		return ""
	}

	return strings.TrimSpace(pkg.Content(sourceCode))
}

// IsPackageObject reports whether this source declares a Scala package object.
func IsPackageObject(sourceCode []byte) bool {
	tree, err := parseScala(sourceCode)
	if err != nil {
		return false
	}
	defer tree.Close()

	return findFirstNodeOfType(tree.RootNode(), "package_object") != nil
}

// ParseTopLevelTypeNames extracts declared top-level type names from Scala source code.
func ParseTopLevelTypeNames(sourceCode []byte) []string {
	tree, err := parseScala(sourceCode)
	if err != nil {
		return []string{}
	}
	defer tree.Close()

	seen := make(map[string]bool)
	result := []string{}

	for i := 0; i < int(tree.RootNode().NamedChildCount()); i++ {
		child := tree.RootNode().NamedChild(i)
		if child == nil || !scalaTopLevelDeclarationTypes[child.Type()] {
			continue
		}

		nameNode := findFirstChildOfType(child, "identifier")
		if nameNode == nil {
			continue
		}

		name := strings.TrimSpace(nameNode.Content(sourceCode))
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, name)
	}

	return result
}

// ParseScalaImports parses Scala source code and classifies imports.
func ParseScalaImports(sourceCode []byte, projectPackages map[string]bool) []ScalaImport {
	tree, err := parseScala(sourceCode)
	if err != nil {
		return []ScalaImport{}
	}
	defer tree.Close()

	importDecls := findNodesOfType(tree.RootNode(), "import_declaration")
	if len(importDecls) == 0 {
		return []ScalaImport{}
	}

	imports := []ScalaImport{}
	for _, node := range importDecls {
		paths := extractImportPaths(node, sourceCode)
		for _, path := range paths {
			if path == "" {
				continue
			}
			imports = append(imports, classifyScalaImport(path, projectPackages))
		}
	}

	return imports
}

func classifyScalaImport(importPath string, projectPackages map[string]bool) ScalaImport {
	isWildcard := strings.HasSuffix(importPath, "._")
	if isStandardLibraryImport(importPath) {
		return StandardLibraryImport{path: importPath, isWildcard: isWildcard}
	}

	if isInternalScalaImport(importPath, projectPackages) {
		return InternalImport{path: importPath, isWildcard: isWildcard}
	}

	return ExternalImport{path: importPath, isWildcard: isWildcard}
}

func isStandardLibraryImport(path string) bool {
	prefixes := []string{
		"scala.",
		"java.",
		"javax.",
		"sun.",
		"dotty.",
		"_root_.scala.",
		"_root_.java.",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func isInternalScalaImport(importPath string, projectPackages map[string]bool) bool {
	pkg := scalaImportPackage(importPath)
	if projectPackages[pkg] || projectPackages[importPath] {
		return true
	}

	for projectPkg := range projectPackages {
		if strings.HasPrefix(pkg, projectPkg+".") || pkg == projectPkg {
			return true
		}
		if strings.HasPrefix(importPath, projectPkg+".") {
			return true
		}
	}

	return false
}

func scalaImportPackage(path string) string {
	trimmed := strings.TrimSuffix(path, "._")
	parts := strings.Split(trimmed, ".")
	if len(parts) <= 1 {
		return trimmed
	}

	last := parts[len(parts)-1]
	if last != "" {
		r, _ := utf8DecodeRuneInString(last)
		if unicode.IsUpper(r) {
			return strings.Join(parts[:len(parts)-1], ".")
		}
	}

	return trimmed
}

func simpleTypeName(path string) string {
	trimmed := strings.TrimSuffix(path, "._")
	parts := strings.Split(trimmed, ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// ExtractTypeIdentifiers returns referenced type-like identifiers in Scala source.
func ExtractTypeIdentifiers(sourceCode []byte) []string {
	tree, err := parseScala(sourceCode)
	if err != nil {
		return []string{}
	}
	defer tree.Close()

	seen := make(map[string]bool)
	result := []string{}

	var walk func(*sitter.Node)
	walk = func(node *sitter.Node) {
		if node == nil {
			return
		}

		typeName := node.Type()
		if typeName == "type_identifier" || typeName == "identifier" {
			name := strings.TrimSpace(node.Content(sourceCode))
			if shouldRecordTypeIdentifier(node, name) && !seen[name] {
				seen[name] = true
				result = append(result, name)
			}
		}

		for i := 0; i < int(node.NamedChildCount()); i++ {
			walk(node.NamedChild(i))
		}
	}

	walk(tree.RootNode())

	for _, name := range ParseTopLevelTypeNames(sourceCode) {
		if name != "" && !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}

	return result
}

func shouldRecordTypeIdentifier(node *sitter.Node, name string) bool {
	if name == "" {
		return false
	}
	r, _ := utf8DecodeRuneInString(name)
	if !unicode.IsUpper(r) {
		return false
	}

	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		switch parent.Type() {
		case "import_declaration", "package_clause", "package_identifier", "namespace_selectors", "arrow_renamed_identifier":
			return false
		}
	}

	return true
}

func extractImportPaths(node *sitter.Node, sourceCode []byte) []string {
	if node == nil {
		return nil
	}

	prefix := []string{}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "identifier":
			segment := strings.TrimSpace(child.Content(sourceCode))
			if segment != "" {
				prefix = append(prefix, segment)
			}
		case "namespace_wildcard":
			if len(prefix) == 0 {
				return nil
			}
			return []string{strings.Join(prefix, ".") + "._"}
		case "namespace_selectors":
			return extractSelectorImports(prefix, child, sourceCode)
		}
	}

	if len(prefix) == 0 {
		return nil
	}
	return []string{strings.Join(prefix, ".")}
}

func extractSelectorImports(prefix []string, selectors *sitter.Node, sourceCode []byte) []string {
	if len(prefix) == 0 || selectors == nil {
		return nil
	}

	base := strings.Join(prefix, ".")
	imports := []string{}

	for i := 0; i < int(selectors.NamedChildCount()); i++ {
		child := selectors.NamedChild(i)
		if child == nil {
			continue
		}

		switch child.Type() {
		case "identifier":
			name := strings.TrimSpace(child.Content(sourceCode))
			if name != "" {
				imports = append(imports, base+"."+name)
			}
		case "namespace_wildcard":
			imports = append(imports, base+"._")
		case "arrow_renamed_identifier":
			original := findFirstChildOfType(child, "identifier")
			if original == nil {
				continue
			}
			name := strings.TrimSpace(original.Content(sourceCode))
			if name != "" {
				imports = append(imports, base+"."+name)
			}
		}
	}

	return imports
}

func parseScala(sourceCode []byte) (*sitter.Tree, error) {
	parser, _ := scalaParserPool.Get().(*sitter.Parser)
	if parser == nil {
		parser = sitter.NewParser()
		parser.SetLanguage(scalaLanguage)
	}
	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	scalaParserPool.Put(parser)
	return tree, err
}

func findFirstNodeOfType(node *sitter.Node, nodeType string) *sitter.Node {
	if node == nil {
		return nil
	}
	if node.Type() == nodeType {
		return node
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		if found := findFirstNodeOfType(node.NamedChild(i), nodeType); found != nil {
			return found
		}
	}

	return nil
}

func findNodesOfType(node *sitter.Node, nodeType string) []*sitter.Node {
	if node == nil {
		return nil
	}

	result := []*sitter.Node{}
	var walk func(*sitter.Node)
	walk = func(n *sitter.Node) {
		if n == nil {
			return
		}
		if n.Type() == nodeType {
			result = append(result, n)
		}
		for i := 0; i < int(n.NamedChildCount()); i++ {
			walk(n.NamedChild(i))
		}
	}
	walk(node)

	return result
}

func findFirstChildOfType(node *sitter.Node, nodeTypes ...string) *sitter.Node {
	if node == nil {
		return nil
	}

	target := make(map[string]bool, len(nodeTypes))
	for _, nodeType := range nodeTypes {
		target[nodeType] = true
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		if target[child.Type()] {
			return child
		}
	}
	return nil
}

func utf8DecodeRuneInString(s string) (rune, int) {
	for _, r := range s {
		return r, 1
	}
	return rune(0), 0
}
