package csharp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LegacyCodeHQ/sanity/vcs"
)

func BuildCSharpIndices(
	suppliedFiles map[string]bool,
	contentReader vcs.ContentReader,
) (map[string][]string, map[string]map[string][]string, map[string]string) {
	namespaceToFiles := make(map[string][]string)
	namespaceToTypes := make(map[string]map[string][]string)
	fileToNamespace := make(map[string]string)

	for filePath := range suppliedFiles {
		if filepath.Ext(filePath) != ".cs" {
			continue
		}

		content, err := contentReader(filePath)
		if err != nil {
			continue
		}

		source := string(content)
		namespace := ParseCSharpNamespace(source)
		fileToNamespace[filePath] = namespace
		namespaceToFiles[namespace] = append(namespaceToFiles[namespace], filePath)

		typeNames := ParseTopLevelCSharpTypeNames(source)
		if len(typeNames) == 0 {
			continue
		}

		typeMap, ok := namespaceToTypes[namespace]
		if !ok {
			typeMap = make(map[string][]string)
			namespaceToTypes[namespace] = typeMap
		}
		for _, typeName := range typeNames {
			typeMap[typeName] = append(typeMap[typeName], filePath)
		}
	}

	return namespaceToFiles, namespaceToTypes, fileToNamespace
}

func ResolveCSharpProjectImports(
	absPath string,
	_ string,
	namespaceToFiles map[string][]string,
	namespaceToTypes map[string]map[string][]string,
	fileToNamespace map[string]string,
	suppliedFiles map[string]bool,
	contentReader vcs.ContentReader,
) ([]string, error) {
	content, err := contentReader(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", absPath, err)
	}

	source := string(content)
	imports := ParseCSharpImports(source)
	referencedTypes := ExtractCSharpTypeIdentifiers(source)
	declaredTypes := make(map[string]bool)
	for _, name := range ParseTopLevelCSharpTypeNames(source) {
		declaredTypes[name] = true
	}

	resolved := make([]string, 0, len(imports))
	seen := make(map[string]bool)
	addDep := func(path string) {
		if path == absPath || !suppliedFiles[path] || seen[path] {
			return
		}
		seen[path] = true
		resolved = append(resolved, path)
	}

	importedTypeNames := make(map[string]bool)
	for _, imp := range imports {
		path := imp.Path
		if path == "" {
			continue
		}

		// "using A.B;" form imports a namespace.
		if typeMap, ok := namespaceToTypes[path]; ok {
			for _, ref := range referencedTypes {
				if declaredTypes[ref] {
					continue
				}
				files := typeMap[ref]
				if len(files) != 1 {
					continue
				}
				addDep(files[0])
			}
			continue
		}

		// "using A.B.TypeName;" can import a specific type.
		lastDot := strings.LastIndex(path, ".")
		if lastDot <= 0 || lastDot >= len(path)-1 {
			continue
		}
		pkg := path[:lastDot]
		typeName := path[lastDot+1:]
		importedTypeNames[typeName] = true
		if !containsString(referencedTypes, typeName) {
			continue
		}
		typeMap := namespaceToTypes[pkg]
		files := typeMap[typeName]
		if len(files) != 1 {
			continue
		}
		addDep(files[0])
	}

	// Same-namespace references do not require using directives in C#.
	if namespace, ok := fileToNamespace[absPath]; ok {
		if typeMap, ok := namespaceToTypes[namespace]; ok {
			for _, ref := range referencedTypes {
				if declaredTypes[ref] || importedTypeNames[ref] {
					continue
				}
				files := typeMap[ref]
				if len(files) != 1 {
					continue
				}
				addDep(files[0])
			}
		}
	}

	_ = namespaceToFiles // retained for future namespace-wide heuristics.
	return resolved, nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
