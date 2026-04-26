package rust

import (
	"github.com/LegacyCodeHQ/clarity/depgraph/moduleapi"
	"github.com/LegacyCodeHQ/clarity/vcs"
)

type Module struct{}

func (Module) Name() string {
	return "Rust"
}

func (Module) Extensions() []string {
	return []string{".rs"}
}

func (Module) Maturity() moduleapi.MaturityLevel {
	return moduleapi.MaturityBasicTests
}

func (Module) NewResolver(ctx *moduleapi.Context, contentReader vcs.ContentReader) moduleapi.Resolver {
	return resolver{
		ctx:             ctx,
		contentReader:   contentReader,
		projectResolver: NewProjectImportResolver(ctx.SuppliedFiles, contentReader),
	}
}

func (Module) IsTestFile(filePath string, contentReader vcs.ContentReader) bool {
	return IsTestFileWithContent(filePath, contentReader)
}

type resolver struct {
	ctx             *moduleapi.Context
	contentReader   vcs.ContentReader
	projectResolver *ProjectImportResolver
}

func (r resolver) ResolveProjectImports(absPath, filePath, ext string) ([]string, error) {
	if r.projectResolver != nil {
		return r.projectResolver.ResolveProjectImports(absPath, filePath)
	}
	return ResolveRustProjectImports(absPath, filePath, r.ctx.SuppliedFiles, r.contentReader)
}

func (resolver) FinalizeGraph(_ moduleapi.Graph) error {
	return nil
}
