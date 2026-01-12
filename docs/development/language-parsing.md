# Language Parsing

The project uses the [go-tree-sitter](https://github.com/smacker/go-tree-sitter) Go bindings to integrate [tree-sitter](https://tree-sitter.github.io/tree-sitter/) for extracting dependency information.

## Supported Languages

**go-tree-sitter** comes with built-in support for various programming languages. Each supported language has its own directory, which can be seen on the [project page](https://github.com/smacker/go-tree-sitter).

### Dart

Dart support required additional setup because there's no pre-built Go binding available in the go-tree-sitter repository.

#### Steps to Add Dart Support

1. **Clone the tree-sitter-dart repository**:
```bash
   git clone https://github.com/UserNobody14/tree-sitter-dart
   cd tree-sitter-dart
```

2. **Build the parser**:
```bash
   # Install the tree-sitter CLI tool
   # Using cargo (recommended)
   cargo install --locked tree-sitter-cli

   # Or using npm
   npm install -g tree-sitter-cli

   # Or download pre-built binaries from:
   # https://github.com/tree-sitter/tree-sitter/releases/latest

   # Generate parser.c and other necessary files
   tree-sitter generate
```

3. **Copy parser files into the project**:

   The following files were copied into `tree_sitter_dart/`:
   - `parser.c` - The generated parser implementation
   - `parser.h` - Parser header file
   - `scanner.c` - Custom scanner for Dart-specific syntax

4. **Create Go bindings** (`tree_sitter_dart/binding.go`):
```go
   package tree_sitter_dart

   import "C"

   import (
       "unsafe"
       sitter "github.com/smacker/go-tree-sitter"
   )

   func GetLanguage() *sitter.Language {
       ptr := unsafe.Pointer(C.tree_sitter_dart())
       return sitter.NewLanguage(ptr)
   }
```

This binding uses CGo to interface with the C-based tree-sitter parser.

5. **Mark as vendored code**:

   Since the `tree_sitter_dart` directory contains generated C code from an external repository, it's marked as vendored to exclude it from language statistics and code analysis.

## Resources

- [go-tree-sitter](https://github.com/smacker/go-tree-sitter)
- [tree-sitter-dart](https://github.com/UserNobody14/tree-sitter-dart)

