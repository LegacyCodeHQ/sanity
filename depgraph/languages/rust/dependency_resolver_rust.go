package rust

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/LegacyCodeHQ/clarity/vcs"
)

type ProjectImportResolver struct {
	suppliedFiles map[string]bool
	contentReader vcs.ContentReader

	crateRootCache sync.Map // source file path -> crate root (or "")
	crateNameCache sync.Map // crate root -> map[string]bool
	modDepsCache   sync.Map // mod.rs path -> []string
}

func NewProjectImportResolver(suppliedFiles map[string]bool, contentReader vcs.ContentReader) *ProjectImportResolver {
	return &ProjectImportResolver{
		suppliedFiles: suppliedFiles,
		contentReader: contentReader,
	}
}

func (r *ProjectImportResolver) ResolveProjectImports(absPath string, filePath string) ([]string, error) {
	content, err := r.contentReader(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", absPath, err)
	}

	imports, parseErr := ParseRustImports(content)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse imports in %s: %w", filePath, parseErr)
	}

	projectImports := make([]string, 0, len(imports))
	for _, imp := range imports {
		switch imp.Kind {
		case RustImportUse:
			projectImports = append(projectImports, r.resolveRustUsePath(absPath, imp.Path)...)
		case RustImportModDecl:
			projectImports = append(projectImports, resolveRustModDecl(absPath, imp.Path, r.suppliedFiles)...)
		case RustImportExternCrate:
			// External crate imports do not map to local project files.
		}
	}

	projectImports = filterOutRustSelfDependency(projectImports, absPath)
	return projectImports, nil
}

func ResolveRustProjectImports(
	absPath string,
	filePath string,
	suppliedFiles map[string]bool,
	contentReader vcs.ContentReader,
) ([]string, error) {
	resolver := NewProjectImportResolver(suppliedFiles, contentReader)
	return resolver.ResolveProjectImports(absPath, filePath)
}

func resolveRustModDecl(sourceFile, moduleName string, suppliedFiles map[string]bool) []string {
	if moduleName == "" {
		return nil
	}

	sourceDir := filepath.Dir(sourceFile)
	candidates := []string{
		filepath.Join(sourceDir, moduleName+".rs"),
		filepath.Join(sourceDir, moduleName, "mod.rs"),
	}

	return filterSuppliedFiles(candidates, suppliedFiles)
}

func (r *ProjectImportResolver) resolveRustUsePath(sourceFile, importPath string) []string {
	path := strings.TrimSpace(importPath)
	if path == "" {
		return nil
	}

	parts := strings.Split(path, "::")
	baseDir := ""
	crateRoot := ""
	rootedInLocalCrate := false

	switch parts[0] {
	case "crate":
		root, ok := r.findRustCrateRoot(sourceFile)
		if !ok {
			return nil
		}
		crateRoot = root
		baseDir = filepath.Join(root, "src")
		rootedInLocalCrate = true
		parts = parts[1:]
	case "self", "super":
		baseDir = filepath.Dir(sourceFile)
		for len(parts) > 0 {
			switch parts[0] {
			case "self":
				parts = parts[1:]
			case "super":
				baseDir = filepath.Dir(baseDir)
				parts = parts[1:]
			default:
				goto resolved
			}
		}
	default:
		root, ok := r.findRustCrateRoot(sourceFile)
		if !ok || !r.isLocalRustCrateImport(parts[0], root) {
			return nil
		}
		crateRoot = root
		baseDir = filepath.Join(root, "src")
		rootedInLocalCrate = true
		parts = parts[1:]
	}

resolved:
	if len(parts) == 0 && rootedInLocalCrate {
		return resolveRustCrateRootCandidates(crateRoot, r.suppliedFiles)
	}
	if len(parts) == 0 {
		return nil
	}

	candidates := resolveRustModuleCandidates(baseDir, parts, r.suppliedFiles)
	if len(parts) > 1 && len(candidates) == 0 {
		candidates = append(candidates, resolveRustModuleCandidates(baseDir, parts[:len(parts)-1], r.suppliedFiles)...)
	}
	if rootedInLocalCrate && len(parts) == 1 && len(candidates) == 0 {
		candidates = append(candidates, resolveRustCrateRootCandidates(crateRoot, r.suppliedFiles)...)
	}

	candidates = deduplicateSuppliedFiles(candidates, r.suppliedFiles)
	candidates = r.expandRustModRsCandidates(candidates)
	return deduplicateSuppliedFiles(candidates, r.suppliedFiles)
}

func resolveRustCrateRootCandidates(crateRoot string, suppliedFiles map[string]bool) []string {
	if crateRoot == "" {
		return nil
	}
	return filterSuppliedFiles([]string{filepath.Join(crateRoot, "src", "lib.rs")}, suppliedFiles)
}

func resolveRustModuleCandidates(baseDir string, parts []string, suppliedFiles map[string]bool) []string {
	if baseDir == "" || len(parts) == 0 {
		return nil
	}

	modulePath := filepath.Join(append([]string{baseDir}, parts...)...)
	candidates := []string{
		modulePath + ".rs",
		filepath.Join(modulePath, "mod.rs"),
	}

	return filterSuppliedFiles(candidates, suppliedFiles)
}

func expandRustModRsDependencies(
	modRsPath string,
	suppliedFiles map[string]bool,
	contentReader vcs.ContentReader,
) []string {
	content, err := contentReader(modRsPath)
	if err != nil {
		return nil
	}

	imports, parseErr := ParseRustImports(content)
	if parseErr != nil {
		return nil
	}

	var resolved []string
	for _, imp := range imports {
		if imp.Kind != RustImportModDecl {
			continue
		}
		resolved = append(resolved, resolveRustModDecl(modRsPath, imp.Path, suppliedFiles)...)
	}

	return deduplicateSuppliedFiles(resolved, suppliedFiles)
}

func filterOutRustSelfDependency(imports []string, sourceFile string) []string {
	if len(imports) == 0 {
		return imports
	}
	filtered := imports[:0]
	for _, imp := range imports {
		if imp == sourceFile {
			continue
		}
		filtered = append(filtered, imp)
	}
	return filtered
}

func findRustCrateRoot(sourceFile string, suppliedFiles map[string]bool, contentReader vcs.ContentReader) (string, bool) {
	dir := filepath.Dir(sourceFile)
	for {
		candidate := filepath.Join(dir, "Cargo.toml")
		if suppliedFiles[candidate] {
			return dir, true
		}
		if contentReader != nil {
			if _, err := contentReader(candidate); err == nil {
				return dir, true
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

func (r *ProjectImportResolver) findRustCrateRoot(sourceFile string) (string, bool) {
	if cached, ok := r.crateRootCache.Load(sourceFile); ok {
		root := cached.(string)
		return root, root != ""
	}

	root, ok := findRustCrateRoot(sourceFile, r.suppliedFiles, r.contentReader)
	if ok {
		r.crateRootCache.Store(sourceFile, root)
		return root, true
	}
	r.crateRootCache.Store(sourceFile, "")
	return "", false
}

func (r *ProjectImportResolver) isLocalRustCrateImport(firstSegment, crateRoot string) bool {
	if firstSegment == "" || crateRoot == "" {
		return false
	}
	if cached, ok := r.crateNameCache.Load(crateRoot); ok {
		return cached.(map[string]bool)[firstSegment]
	}

	names := make(map[string]bool)
	cargoToml := filepath.Join(crateRoot, "Cargo.toml")
	if r.contentReader != nil {
		if content, err := r.contentReader(cargoToml); err == nil {
			names = parseRustCrateNamesFromCargoToml(string(content))
		}
	}
	r.crateNameCache.Store(crateRoot, names)
	return names[firstSegment]
}

func (r *ProjectImportResolver) expandRustModRsCandidates(candidates []string) []string {
	if len(candidates) == 0 || r.contentReader == nil {
		return candidates
	}

	expanded := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if filepath.Base(candidate) != "mod.rs" {
			expanded = append(expanded, candidate)
			continue
		}

		modChildren := r.expandRustModRsDependencies(candidate)
		if len(modChildren) == 0 {
			expanded = append(expanded, candidate)
			continue
		}
		expanded = append(expanded, modChildren...)
	}

	return expanded
}

func (r *ProjectImportResolver) expandRustModRsDependencies(modRsPath string) []string {
	if cached, ok := r.modDepsCache.Load(modRsPath); ok {
		return cached.([]string)
	}

	resolved := expandRustModRsDependencies(modRsPath, r.suppliedFiles, r.contentReader)
	r.modDepsCache.Store(modRsPath, resolved)
	return resolved
}

func parseRustCrateNamesFromCargoToml(content string) map[string]bool {
	names := make(map[string]bool)
	section := ""
	packageName := ""
	libName := ""

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(strings.Trim(line, "[]"))
			continue
		}

		if !strings.HasPrefix(line, "name") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"")
		if value == "" {
			continue
		}
		switch section {
		case "package":
			packageName = value
		case "lib":
			libName = value
		}
	}

	if libName != "" {
		names[libName] = true
	}
	if packageName != "" {
		names[normalizeCargoCrateName(packageName)] = true
	}
	return names
}

func normalizeCargoCrateName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func filterSuppliedFiles(paths []string, suppliedFiles map[string]bool) []string {
	if len(paths) == 0 {
		return nil
	}
	var filtered []string
	for _, path := range paths {
		if suppliedFiles[path] {
			filtered = append(filtered, path)
		}
	}
	return filtered
}

func deduplicateSuppliedFiles(paths []string, suppliedFiles map[string]bool) []string {
	if len(paths) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	var result []string
	for _, path := range paths {
		if !suppliedFiles[path] {
			continue
		}
		if !seen[path] {
			seen[path] = true
			result = append(result, path)
		}
	}
	return result
}
