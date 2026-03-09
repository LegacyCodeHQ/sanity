package typescript

import (
	"bytes"
	"fmt"

	"github.com/LegacyCodeHQ/clarity/vcs"
)

func ResolveTypeScriptProjectImports(
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
	if !bytes.Contains(content, []byte("./")) &&
		!bytes.Contains(content, []byte("../")) &&
		!bytes.Contains(content, []byte("@/")) {
		return nil, nil
	}

	imports, parseErr := ParseTypeScriptImports(content, ext == ".tsx")
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse imports in %s: %w", filePath, parseErr)
	}

	var projectImports []string
	seen := make(map[string]bool, len(imports))
	for _, imp := range imports {
		if internalImp, ok := imp.(InternalImport); ok {
			resolvedFiles := ResolveTypeScriptImportPath(absPath, internalImp.Path(), suppliedFiles)
			for _, resolvedFile := range resolvedFiles {
				if seen[resolvedFile] {
					continue
				}
				seen[resolvedFile] = true
				projectImports = append(projectImports, resolvedFile)
			}
		}
	}

	return projectImports, nil
}
