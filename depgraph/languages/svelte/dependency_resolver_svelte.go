package svelte

import (
	"fmt"
	"path/filepath"

	"github.com/LegacyCodeHQ/clarity/depgraph/languages/javascript"
	"github.com/LegacyCodeHQ/clarity/vcs"
)

func ResolveSvelteProjectImports(
	absPath string,
	filePath string,
	suppliedFiles map[string]bool,
	contentReader vcs.ContentReader,
) ([]string, error) {
	content, err := contentReader(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", absPath, err)
	}

	imports, parseErr := ParseSvelteImports(content)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse imports in %s: %w", filePath, parseErr)
	}

	var projectImports []string
	for _, imp := range imports {
		if internalImp, ok := imp.(javascript.InternalImport); ok {
			resolvedFiles := ResolveSvelteImportPath(absPath, internalImp.Path(), suppliedFiles)
			projectImports = append(projectImports, resolvedFiles...)
		}
	}

	return projectImports, nil
}

// ResolveSvelteImportPath resolves a Svelte import path to possible file paths.
// It tries JS/JSX extensions first (via the JavaScript resolver), then .svelte.
func ResolveSvelteImportPath(sourceFile, importPath string, suppliedFiles map[string]bool) []string {
	resolved := javascript.ResolveJavaScriptImportPath(sourceFile, importPath, suppliedFiles)

	sourceDir := filepath.Dir(sourceFile)
	basePath := filepath.Join(sourceDir, importPath)
	basePath = filepath.Clean(basePath)

	// Try .svelte extension
	candidate := basePath + ".svelte"
	if suppliedFiles[candidate] {
		resolved = append(resolved, candidate)
	}

	// Try index.svelte for directory imports
	indexCandidate := filepath.Join(basePath, "index.svelte")
	if suppliedFiles[indexCandidate] {
		resolved = append(resolved, indexCandidate)
	}

	// If import already ends with .svelte, try exact path
	if filepath.Ext(importPath) == ".svelte" {
		if suppliedFiles[basePath] {
			resolved = append(resolved, basePath)
		}
	}

	return resolved
}
