package parsers

import (
	"fmt"
	"path/filepath"

	_go "github.com/LegacyCodeHQ/sanity/parsers/go"
	"github.com/LegacyCodeHQ/sanity/vcs"
)

// BuildDependencyGraph analyzes a list of files and builds a dependency graph
// containing only project imports (excluding package:/dart: imports for Dart,
// and standard library/external imports for Go).
// Only dependencies that are in the supplied file list are included in the graph.
// The contentReader function is used to read file contents (from filesystem, git commit, etc.)
func BuildDependencyGraph(filePaths []string, contentReader vcs.ContentReader) (DependencyGraph, error) {
	graph := make(DependencyGraph)

	ctx, err := buildDependencyGraphContext(filePaths, contentReader)
	if err != nil {
		return nil, err
	}

	goPackageExportIndices := buildGoPackageExportIndices(ctx.dirToFiles, contentReader)
	kotlinPackageIndex, kotlinPackageTypes, kotlinFilePackages := buildKotlinIndices(ctx.kotlinFiles, contentReader)

	// Second pass: build the dependency graph
	for _, filePath := range filePaths {
		// Get absolute path
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve path %s: %w", filePath, err)
		}

		ext := filepath.Ext(absPath)

		// Check if this is a supported file type
		if !isSupportedDependencyFileExt(ext) {
			// Unsupported files are included in the graph with no dependencies
			graph[absPath] = []string{}
			continue
		}

		projectImports, err := buildProjectImports(
			absPath,
			filePath,
			ext,
			ctx,
			goPackageExportIndices,
			kotlinPackageIndex,
			kotlinPackageTypes,
			kotlinFilePackages,
			contentReader,
		)
		if err != nil {
			return nil, err
		}

		if len(projectImports) > 0 {
			projectImports = deduplicatePaths(projectImports)
		}

		graph[absPath] = projectImports
	}

	// Third pass: add intra-package dependencies for languages that need it.
	if err := addGoIntraPackageDependencies(graph, ctx.goFiles, contentReader); err != nil {
		return graph, fmt.Errorf("failed to add intra-package dependencies: %w", err)
	}

	return graph, nil
}

func isSupportedDependencyFileExt(ext string) bool {
	switch ext {
	case ".dart", ".go", ".kt", ".ts", ".tsx":
		return true
	default:
		return false
	}
}

func buildProjectImports(
	absPath string,
	filePath string,
	ext string,
	ctx *dependencyGraphContext,
	goPackageExportIndices map[string]_go.GoPackageExportIndex,
	kotlinPackageIndex map[string][]string,
	kotlinPackageTypes map[string]map[string][]string,
	kotlinFilePackages map[string]string,
	contentReader vcs.ContentReader,
) ([]string, error) {
	switch ext {
	case ".dart":
		return buildDartProjectImports(absPath, filePath, ext, ctx.suppliedFiles, contentReader)
	case ".go":
		return buildGoProjectImports(
			absPath,
			filePath,
			ctx.dirToFiles,
			goPackageExportIndices,
			ctx.suppliedFiles,
			contentReader,
		)
	case ".kt":
		return buildKotlinProjectImports(
			absPath,
			filePath,
			kotlinPackageIndex,
			kotlinPackageTypes,
			kotlinFilePackages,
			ctx.suppliedFiles,
			contentReader,
		)
	case ".ts", ".tsx":
		return buildTypeScriptProjectImports(absPath, filePath, ext, ctx.suppliedFiles, contentReader)
	default:
		return []string{}, nil
	}
}

// deduplicatePaths removes duplicate entries while preserving insertion order
func deduplicatePaths(paths []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	return result
}
