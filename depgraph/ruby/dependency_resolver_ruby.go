package ruby

import (
	"fmt"

	"github.com/LegacyCodeHQ/sanity/vcs"
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
	for _, imp := range imports {
		resolvedFiles := ResolveRubyImportPath(absPath, imp, suppliedFiles)
		projectImports = append(projectImports, resolvedFiles...)
	}

	return projectImports, nil
}
