package golang

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/LegacyCodeHQ/clarity/vcs"
)

// ProjectImportResolver encapsulates Go-specific dependency resolution caches and logic.
type ProjectImportResolver struct {
	dirToFiles             map[string][]string
	goPackageExportIndices map[string]GoPackageExportIndex
	suppliedFiles          map[string]bool
	contentReader          vcs.ContentReader
	moduleRootCache        sync.Map // source dir -> module root (or "")
	moduleInfoCache        sync.Map // module root -> goModuleInfo
	importPathCache        sync.Map // source file + import path -> resolved package dir (or "")
	symbolInfoCache        sync.Map // absolute file path -> *GoSymbolInfo
}

type goModuleInfo struct {
	moduleName  string
	replacePaths map[string]string
}

// NewProjectImportResolver creates a Go dependency resolver with precomputed package export indices.
func NewProjectImportResolver(
	dirToFiles map[string][]string,
	suppliedFiles map[string]bool,
	contentReader vcs.ContentReader,
) *ProjectImportResolver {
	return &ProjectImportResolver{
		dirToFiles:             dirToFiles,
		goPackageExportIndices: BuildGoPackageExportIndices(dirToFiles, contentReader),
		suppliedFiles:          suppliedFiles,
		contentReader:          contentReader,
	}
}

// ResolveProjectImports resolves Go project imports for a single file using cached indices.
func (r *ProjectImportResolver) ResolveProjectImports(absPath, filePath string) ([]string, error) {
	return resolveGoProjectImports(
		absPath,
		filePath,
		r.dirToFiles,
		r.goPackageExportIndices,
		r.suppliedFiles,
		r.contentReader,
		r.resolveImportPath,
		r.storeSymbolInfo)
}

func BuildGoPackageExportIndices(dirToFiles map[string][]string, contentReader vcs.ContentReader) map[string]GoPackageExportIndex {
	goPackageExportIndices := make(map[string]GoPackageExportIndex) // packageDir -> export index
	for dir, files := range dirToFiles {
		// Check if this directory has Go files
		hasGoFiles := false
		var goFilesInDir []string
		for _, f := range files {
			if filepath.Ext(f) == ".go" {
				hasGoFiles = true
				goFilesInDir = append(goFilesInDir, f)
			}
		}
		if hasGoFiles {
			exportIndex, err := BuildPackageExportIndex(goFilesInDir, vcs.ContentReader(contentReader))
			if err != nil {
				continue
			}
			goPackageExportIndices[dir] = exportIndex
		}
	}

	return goPackageExportIndices
}

func ResolveGoProjectImports(
	absPath string,
	filePath string,
	dirToFiles map[string][]string,
	goPackageExportIndices map[string]GoPackageExportIndex,
	suppliedFiles map[string]bool,
	contentReader vcs.ContentReader,
) ([]string, error) {
	return resolveGoProjectImports(
		absPath,
		filePath,
		dirToFiles,
		goPackageExportIndices,
		suppliedFiles,
		contentReader,
		func(sourceFile, importPath string) string {
			return resolveGoImportPath(sourceFile, importPath, contentReader)
		},
		nil)
}

func resolveGoProjectImports(
	absPath string,
	filePath string,
	dirToFiles map[string][]string,
	goPackageExportIndices map[string]GoPackageExportIndex,
	suppliedFiles map[string]bool,
	contentReader vcs.ContentReader,
	importPathResolver func(sourceFile, importPath string) string,
	symbolInfoSink func(filePath string, info *GoSymbolInfo),
) ([]string, error) {
	sourceContent, err := contentReader(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", absPath, err)
	}

	imports, embeds, symbolInfo, exportInfo, err := AnalyzeGoFileFromContent(absPath, sourceContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse imports in %s: %w", filePath, err)
	}
	if symbolInfoSink != nil && symbolInfo != nil {
		symbolInfoSink(absPath, symbolInfo)
	}

	var projectImports []string

	// Parse //go:embed directives.
	for _, embed := range embeds {
		projectImports = append(projectImports, resolveGoEmbedPaths(absPath, embed.Pattern, suppliedFiles)...)
	}

	// Determine if this is a test file
	isTestFile := strings.HasSuffix(absPath, "_test.go")

	for _, imp := range imports {
		var importPath string

		// Check both InternalImport and ExternalImport types
		// resolveGoImportPath will determine if they're actually part of this module
		switch typedImp := imp.(type) {
		case InternalImport:
			importPath = typedImp.Path()
		case ExternalImport:
			importPath = typedImp.Path()
		default:
			continue
		}

		packageDir := importPathResolver(absPath, importPath)
		if packageDir == "" {
			continue
		}

		sourceDir := filepath.Dir(absPath)
		sameDir := sourceDir == packageDir
		exportIndex, hasExportIndex := goPackageExportIndices[packageDir]

		var usedSymbols map[string]bool
		if exportInfo != nil {
			usedSymbols = GetUsedSymbolsFromPackage(exportInfo, importPath)
		}

		if files, ok := dirToFiles[packageDir]; ok {
			for _, depFile := range files {
				if depFile == absPath {
					continue
				}

				if strings.HasSuffix(depFile, "_test.go") && !sameDir {
					continue
				}

				if filepath.Ext(depFile) != ".go" {
					continue
				}

				if (!sameDir || isTestFile) && hasExportIndex && usedSymbols != nil && len(usedSymbols) > 0 {
					if !fileDefinesAnyUsedSymbol(depFile, usedSymbols, exportIndex) {
						continue
					}
				}

				projectImports = append(projectImports, depFile)
			}
		}
	}

	return projectImports, nil
}

func (r *ProjectImportResolver) resolveImportPath(sourceFile, importPath string) string {
	cacheKey := sourceFile + "\x00" + importPath
	if cached, ok := r.importPathCache.Load(cacheKey); ok {
		return cached.(string)
	}

	sourceDir := filepath.Dir(sourceFile)
	moduleRoot := r.findModuleRootCached(sourceDir)
	if moduleRoot == "" {
		r.importPathCache.Store(cacheKey, "")
		return ""
	}

	moduleInfo := r.getModuleInfoCached(moduleRoot)
	if moduleInfo.moduleName == "" {
		r.importPathCache.Store(cacheKey, "")
		return ""
	}

	if strings.HasPrefix(importPath, moduleInfo.moduleName) {
		relativePath := strings.TrimPrefix(importPath, moduleInfo.moduleName+"/")
		absPath := filepath.Join(moduleRoot, relativePath)
		resolved := filepath.Clean(absPath)
		r.importPathCache.Store(cacheKey, resolved)
		return resolved
	}

	if replacedPath := resolveViaReplace(importPath, moduleInfo.replacePaths); replacedPath != "" {
		r.importPathCache.Store(cacheKey, replacedPath)
		return replacedPath
	}

	r.importPathCache.Store(cacheKey, "")
	return ""
}

func (r *ProjectImportResolver) findModuleRootCached(startDir string) string {
	if cached, ok := r.moduleRootCache.Load(startDir); ok {
		return cached.(string)
	}

	dir := startDir
	visited := make([]string, 0, 8)
	for {
		visited = append(visited, dir)
		if cached, ok := r.moduleRootCache.Load(dir); ok {
			root := cached.(string)
			for _, path := range visited {
				r.moduleRootCache.Store(path, root)
			}
			return root
		}

		goModPath := filepath.Join(dir, "go.mod")
		if _, err := r.contentReader(goModPath); err == nil {
			for _, path := range visited {
				r.moduleRootCache.Store(path, dir)
			}
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			for _, path := range visited {
				r.moduleRootCache.Store(path, "")
			}
			return ""
		}
		dir = parent
	}
}

func (r *ProjectImportResolver) getModuleInfoCached(moduleRoot string) goModuleInfo {
	if cached, ok := r.moduleInfoCache.Load(moduleRoot); ok {
		return cached.(goModuleInfo)
	}

	moduleName, replacePaths := getModuleInfo(moduleRoot, r.contentReader)
	info := goModuleInfo{
		moduleName:  moduleName,
		replacePaths: replacePaths,
	}
	r.moduleInfoCache.Store(moduleRoot, info)
	return info
}

func (r *ProjectImportResolver) storeSymbolInfo(filePath string, info *GoSymbolInfo) {
	if info == nil {
		return
	}
	r.symbolInfoCache.Store(filePath, info)
}

func (r *ProjectImportResolver) getSymbolInfo(filePath string) (*GoSymbolInfo, bool) {
	cached, ok := r.symbolInfoCache.Load(filePath)
	if !ok {
		return nil, false
	}
	info, ok := cached.(*GoSymbolInfo)
	if !ok || info == nil {
		return nil, false
	}
	return info, true
}

func fileDefinesAnyUsedSymbol(depFile string, usedSymbols map[string]bool, exportIndex GoPackageExportIndex) bool {
	for symbol := range usedSymbols {
		if definingFiles, ok := exportIndex[symbol]; ok {
			for _, defFile := range definingFiles {
				if defFile == depFile {
					return true
				}
			}
		}
	}

	return false
}

// resolveGoImportPath resolves a Go import path to an absolute file path
// The contentReader is used to read go.mod content
func resolveGoImportPath(sourceFile, importPath string, contentReader vcs.ContentReader) string {
	// For Go files, we need to find the module root and resolve the import
	// This is a simplified version that assumes the project follows standard Go module structure

	// Find the go.mod file by walking up from the source file
	moduleRoot := findModuleRootWithReader(filepath.Dir(sourceFile), contentReader)
	if moduleRoot == "" {
		// If no module root found, return empty string
		return ""
	}

	// Get module metadata from go.mod using the content reader
	moduleName, replacePaths := getModuleInfo(moduleRoot, contentReader)
	if moduleName == "" {
		return ""
	}

	// Check if the import path starts with the module name
	if strings.HasPrefix(importPath, moduleName) {
		// Remove module name prefix to get relative path
		relativePath := strings.TrimPrefix(importPath, moduleName+"/")

		// Construct absolute path
		absPath := filepath.Join(moduleRoot, relativePath)

		// For Go, we don't add .go extension here because imports refer to packages (directories)
		return filepath.Clean(absPath)
	}

	// Check go.mod replace directives for local replacement targets.
	replacedPath := resolveViaReplace(importPath, replacePaths)
	if replacedPath != "" {
		return replacedPath
	}

	// Not an internal import relative to this module.
	return ""
}

// findModuleRootWithReader walks up the directory tree to find go.mod using the provided content reader.
func findModuleRootWithReader(startDir string, contentReader vcs.ContentReader) string {
	dir := startDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := contentReader(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			return ""
		}
		dir = parent
	}
}

// getModuleInfo reads module metadata from go.mod using the content reader.
func getModuleInfo(moduleRoot string, contentReader vcs.ContentReader) (string, map[string]string) {
	goModPath := filepath.Join(moduleRoot, "go.mod")
	content, err := contentReader(goModPath)
	if err != nil {
		return "", make(map[string]string)
	}

	moduleName := ""
	replacePaths := make(map[string]string)
	inReplaceBlock := false

	// Parse module name and local replace directives from the content.
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			moduleName = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "module")), "\"")
			continue
		}

		if strings.HasPrefix(line, "replace ") && strings.HasSuffix(line, "(") {
			inReplaceBlock = true
			continue
		}
		if inReplaceBlock && line == ")" {
			inReplaceBlock = false
			continue
		}

		if strings.HasPrefix(line, "replace ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "replace "))
		}

		if inReplaceBlock || strings.Contains(line, "=>") {
			oldPath, newPath, ok := parseReplaceLine(line)
			if !ok {
				continue
			}
			if !isLocalGoReplaceTarget(newPath) {
				continue
			}
			if !filepath.IsAbs(newPath) {
				newPath = filepath.Join(moduleRoot, newPath)
			}
			replacePaths[oldPath] = filepath.Clean(newPath)
		}
	}

	return moduleName, replacePaths
}

func parseReplaceLine(line string) (string, string, bool) {
	parts := strings.Split(line, "=>")
	if len(parts) != 2 {
		return "", "", false
	}

	oldPart := strings.TrimSpace(parts[0])
	newPart := strings.TrimSpace(parts[1])
	if oldPart == "" || newPart == "" {
		return "", "", false
	}

	oldFields := strings.Fields(oldPart)
	newFields := strings.Fields(newPart)
	if len(oldFields) == 0 || len(newFields) == 0 {
		return "", "", false
	}

	return oldFields[0], strings.Trim(newFields[0], "\""), true
}

func isLocalGoReplaceTarget(target string) bool {
	return strings.HasPrefix(target, "./") ||
		strings.HasPrefix(target, "../") ||
		strings.HasPrefix(target, "/")
}

func resolveViaReplace(importPath string, replacePaths map[string]string) string {
	bestOldPath := ""
	bestNewPath := ""
	for oldPath, newPath := range replacePaths {
		if importPath == oldPath || strings.HasPrefix(importPath, oldPath+"/") {
			if len(oldPath) > len(bestOldPath) {
				bestOldPath = oldPath
				bestNewPath = newPath
			}
		}
	}
	if bestOldPath == "" {
		return ""
	}

	suffix := strings.TrimPrefix(importPath, bestOldPath)
	suffix = strings.TrimPrefix(suffix, "/")
	if suffix == "" {
		return filepath.Clean(bestNewPath)
	}
	return filepath.Clean(filepath.Join(bestNewPath, suffix))
}

// resolveGoEmbedPaths resolves a Go embed pattern to absolute file paths.
func resolveGoEmbedPaths(sourceFile, pattern string, suppliedFiles map[string]bool) []string {
	// Get directory of source file
	sourceDir := filepath.Dir(sourceFile)

	// For simple file patterns (no glob characters), just resolve directly
	if !strings.ContainsAny(pattern, "*?[") {
		absPath := filepath.Join(sourceDir, pattern)
		absPath = filepath.Clean(absPath)

		// Check if this file is in the supplied files
		if suppliedFiles[absPath] {
			return []string{absPath}
		}
		return nil
	}

	// For glob patterns, we need to match against supplied files
	// Create a glob pattern with the full path
	globPattern := filepath.Join(sourceDir, pattern)

	var matches []string
	for file := range suppliedFiles {
		matched, err := filepath.Match(globPattern, file)
		if err != nil {
			continue
		}
		if matched {
			matches = append(matches, file)
		}
	}

	return matches
}
