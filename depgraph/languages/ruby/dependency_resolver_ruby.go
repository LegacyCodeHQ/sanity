package ruby

import (
	"fmt"
	"path/filepath"

	"github.com/LegacyCodeHQ/clarity/vcs"
)

func ResolveRubyProjectImports(
	absPath string,
	filePath string,
	suppliedFiles map[string]bool,
	contentReader vcs.ContentReader,
) ([]string, error) {
	content, err := contentReader(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", absPath, err)
	}

	imports, parseErr := ParseRubyImports(content)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse imports in %s: %w", filePath, parseErr)
	}

	var projectImports []string
	seen := make(map[string]struct{})

	for _, imp := range imports {
		resolvedFiles := ResolveRubyImportPath(absPath, imp, suppliedFiles)
		for _, file := range resolvedFiles {
			if file == absPath {
				continue
			}
			if _, ok := seen[file]; ok {
				continue
			}
			seen[file] = struct{}{}
			projectImports = append(projectImports, file)
		}
	}

	constantRefs := ParseRubyConstantReferences(content)
	if len(constantRefs) > 0 {
		pathComponentsCache := make(map[string][]string, len(suppliedFiles))
		for fp, exists := range suppliedFiles {
			if exists && filepath.Ext(fp) == ".rb" {
				pathComponentsCache[fp] = rubyPathComponents(fp)
			}
		}

		for _, ref := range constantRefs {
			resolvedFiles := resolveRubyConstantReferencePathCached(ref, pathComponentsCache)
			for _, file := range resolvedFiles {
				if file == absPath {
					continue
				}
				if _, ok := seen[file]; ok {
					continue
				}
				seen[file] = struct{}{}
				projectImports = append(projectImports, file)
			}
		}
	}

	return projectImports, nil
}
