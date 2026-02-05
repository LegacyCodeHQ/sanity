package parsers

import (
	"fmt"

	"github.com/LegacyCodeHQ/sanity/parsers/typescript"
	"github.com/LegacyCodeHQ/sanity/vcs"
)

func buildTypeScriptProjectImports(
	absPath string,
	filePath string,
	ext string,
	suppliedFiles map[string]bool,
	contentReader vcs.ContentReader,
) ([]string, error) {
	content, err := contentReader(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", absPath, err)
	}

	imports, parseErr := typescript.ParseTypeScriptImports(content, ext == ".tsx")
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse imports in %s: %w", filePath, parseErr)
	}

	var projectImports []string
	for _, imp := range imports {
		if internalImp, ok := imp.(typescript.InternalImport); ok {
			resolvedFiles := typescript.ResolveTypeScriptImportPath(absPath, internalImp.Path(), suppliedFiles)
			projectImports = append(projectImports, resolvedFiles...)
		}
	}

	return projectImports, nil
}
