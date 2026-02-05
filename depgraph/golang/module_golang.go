package golang

import (
	"github.com/LegacyCodeHQ/sanity/depgraph/langsupport"
	"github.com/LegacyCodeHQ/sanity/vcs"
)

type Module struct{}

func (Module) Name() string {
	return "Go"
}

func (Module) Extensions() []string {
	return []string{".go"}
}

func (Module) NewResolver(ctx *langsupport.Context, contentReader vcs.ContentReader) langsupport.Resolver {
	return resolver{
		ctx:             ctx,
		contentReader:   contentReader,
		projectResolver: NewProjectImportResolver(ctx.DirToFiles, ctx.SuppliedFiles, contentReader),
	}
}

func (Module) IsTestFile(filePath string) bool {
	return IsTestFile(filePath)
}

type resolver struct {
	ctx             *langsupport.Context
	contentReader   vcs.ContentReader
	projectResolver *ProjectImportResolver
}

func (r resolver) ResolveProjectImports(absPath, filePath, _ string) ([]string, error) {
	return r.projectResolver.ResolveProjectImports(absPath, filePath)
}

func (r resolver) FinalizeGraph(graph langsupport.Graph) error {
	return AddGoIntraPackageDependencies(graph, r.ctx.GoFiles, r.contentReader)
}
