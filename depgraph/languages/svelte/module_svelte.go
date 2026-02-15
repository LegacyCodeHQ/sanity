package svelte

import (
	"github.com/LegacyCodeHQ/clarity/depgraph/moduleapi"
	"github.com/LegacyCodeHQ/clarity/vcs"
)

type Module struct{}

func (Module) Name() string {
	return "Svelte"
}

func (Module) Extensions() []string {
	return []string{".svelte"}
}

func (Module) Maturity() moduleapi.MaturityLevel {
	return moduleapi.MaturityUntested
}

func (Module) NewResolver(ctx *moduleapi.Context, contentReader vcs.ContentReader) moduleapi.Resolver {
	return resolver{ctx: ctx, contentReader: contentReader}
}

func (Module) IsTestFile(filePath string, _ vcs.ContentReader) bool {
	return IsTestFile(filePath)
}

type resolver struct {
	ctx           *moduleapi.Context
	contentReader vcs.ContentReader
}

func (r resolver) ResolveProjectImports(absPath, filePath, ext string) ([]string, error) {
	return ResolveSvelteProjectImports(absPath, filePath, r.ctx.SuppliedFiles, r.contentReader)
}

func (resolver) FinalizeGraph(_ moduleapi.Graph) error {
	return nil
}
