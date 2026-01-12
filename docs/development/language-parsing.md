# Language Parsing

The project uses the [go-tree-sitter](https://github.com/smacker/go-tree-sitter) bindings to integrate [tree-sitter](https://tree-sitter.github.io/tree-sitter/) for parsing and extracting dependency information from source files.

## Available Languages

go-tree-sitter provides built-in support for parsing various programming languages, with each language binding in its own directory within the [project's source](https://github.com/smacker/go-tree-sitter).

## Unsupported Languages

### Dart

Dart support is not available in go-tree-sitter. To add support for Dart:

1. The [tree-sitter-dart](https://github.com/UserNobody14/tree-sitter-dart) repository was cloned and the parser was built using the [tree-sitter-cli](https://github.com/tree-sitter/tree-sitter/blob/master/crates/cli/README.md) tool.
2. The parser files (`parser.c`, `parser.h`, `scanner.c`) were copied into `tree_sitter_dart/`
3. Go bindings were created to interface with the C-based tree-sitter parser using CGo
4. The `tree_sitter_dart/` directory was marked as vendored code to exclude it from language statistics and code analysis

## Resources

- [go-tree-sitter](https://github.com/smacker/go-tree-sitter)
- [tree-sitter-dart](https://github.com/UserNobody14/tree-sitter-dart)
