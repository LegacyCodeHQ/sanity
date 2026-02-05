package parsers

import (
	"github.com/LegacyCodeHQ/sanity/parsers/dart"
	_go "github.com/LegacyCodeHQ/sanity/parsers/go"
	"github.com/LegacyCodeHQ/sanity/parsers/kotlin"
	"github.com/LegacyCodeHQ/sanity/parsers/typescript"
	"github.com/LegacyCodeHQ/sanity/vcs"
)

// DependencyBuilder builds project imports per file and can finalize graph-wide dependencies.
type DependencyBuilder interface {
	BuildProjectImports(absPath, filePath, ext string) ([]string, error)
	FinalizeGraph(graph DependencyGraph) error
}

type defaultDependencyBuilder struct {
	ctx                  *dependencyGraphContext
	contentReader        vcs.ContentReader
	goPackageExportIndex map[string]_go.GoPackageExportIndex
	kotlinPackageIndex   map[string][]string
	kotlinPackageTypes   map[string]map[string][]string
	kotlinFilePackages   map[string]string
}

// NewDefaultDependencyBuilder creates the built-in language-aware dependency builder.
func NewDefaultDependencyBuilder(ctx *dependencyGraphContext, contentReader vcs.ContentReader) DependencyBuilder {
	goPackageExportIndex := _go.BuildGoPackageExportIndices(ctx.dirToFiles, contentReader)
	kotlinPackageIndex, kotlinPackageTypes, kotlinFilePackages := kotlin.BuildKotlinIndices(ctx.kotlinFiles, contentReader)

	return &defaultDependencyBuilder{
		ctx:                  ctx,
		contentReader:        contentReader,
		goPackageExportIndex: goPackageExportIndex,
		kotlinPackageIndex:   kotlinPackageIndex,
		kotlinPackageTypes:   kotlinPackageTypes,
		kotlinFilePackages:   kotlinFilePackages,
	}
}

func (b *defaultDependencyBuilder) BuildProjectImports(absPath, filePath, ext string) ([]string, error) {
	switch ext {
	case ".dart":
		return dart.BuildDartProjectImports(absPath, filePath, ext, b.ctx.suppliedFiles, b.contentReader)
	case ".go":
		return _go.BuildGoProjectImports(
			absPath,
			filePath,
			b.ctx.dirToFiles,
			b.goPackageExportIndex,
			b.ctx.suppliedFiles,
			b.contentReader,
		)
	case ".kt":
		return kotlin.BuildKotlinProjectImports(
			absPath,
			filePath,
			b.kotlinPackageIndex,
			b.kotlinPackageTypes,
			b.kotlinFilePackages,
			b.ctx.suppliedFiles,
			b.contentReader,
		)
	case ".ts", ".tsx":
		return typescript.BuildTypeScriptProjectImports(absPath, filePath, ext, b.ctx.suppliedFiles, b.contentReader)
	default:
		return []string{}, nil
	}
}

func (b *defaultDependencyBuilder) FinalizeGraph(graph DependencyGraph) error {
	return _go.AddGoIntraPackageDependencies(graph, b.ctx.goFiles, b.contentReader)
}
