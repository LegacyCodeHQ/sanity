package depgraph

import (
	"github.com/LegacyCodeHQ/sanity/depgraph/dart"
	"github.com/LegacyCodeHQ/sanity/depgraph/golang"
	"github.com/LegacyCodeHQ/sanity/depgraph/java"
	"github.com/LegacyCodeHQ/sanity/depgraph/kotlin"
	"github.com/LegacyCodeHQ/sanity/depgraph/typescript"
	"github.com/LegacyCodeHQ/sanity/vcs"
)

// LanguageResolver resolves imports for one language and can finalize graph-wide state.
type LanguageResolver interface {
	ResolveProjectImports(absPath, filePath, ext string) ([]string, error)
	FinalizeGraph(graph DependencyGraph) error
}

// LanguageModule describes pluggable language support.
type LanguageModule interface {
	Name() string
	Extensions() []string
	NewResolver(ctx *dependencyGraphContext, contentReader vcs.ContentReader) LanguageResolver
	IsTestFile(filePath string) bool
}

type dartLanguageModule struct{}

func (dartLanguageModule) Name() string {
	return "Dart"
}

func (dartLanguageModule) Extensions() []string {
	return []string{".dart"}
}

func (dartLanguageModule) NewResolver(ctx *dependencyGraphContext, contentReader vcs.ContentReader) LanguageResolver {
	return dartLanguageResolver{ctx: ctx, contentReader: contentReader}
}

func (dartLanguageModule) IsTestFile(filePath string) bool {
	return dart.IsTestFile(filePath)
}

type goLanguageModule struct{}

func (goLanguageModule) Name() string {
	return "Go"
}

func (goLanguageModule) Extensions() []string {
	return []string{".go"}
}

func (goLanguageModule) NewResolver(ctx *dependencyGraphContext, contentReader vcs.ContentReader) LanguageResolver {
	return goLanguageResolver{
		ctx:             ctx,
		contentReader:   contentReader,
		projectResolver: golang.NewProjectImportResolver(ctx.dirToFiles, ctx.suppliedFiles, contentReader),
	}
}

func (goLanguageModule) IsTestFile(filePath string) bool {
	return golang.IsTestFile(filePath)
}

type javaLanguageModule struct{}

func (javaLanguageModule) Name() string {
	return "Java"
}

func (javaLanguageModule) Extensions() []string {
	return []string{".java"}
}

func (javaLanguageModule) NewResolver(ctx *dependencyGraphContext, contentReader vcs.ContentReader) LanguageResolver {
	packageIndex, packageTypes, filePackages := java.BuildJavaIndices(ctx.javaFiles, contentReader)
	return javaLanguageResolver{
		ctx:           ctx,
		contentReader: contentReader,
		packageIndex:  packageIndex,
		packageTypes:  packageTypes,
		filePackages:  filePackages,
	}
}

func (javaLanguageModule) IsTestFile(filePath string) bool {
	return java.IsTestFile(filePath)
}

type kotlinLanguageModule struct{}

func (kotlinLanguageModule) Name() string {
	return "Kotlin"
}

func (kotlinLanguageModule) Extensions() []string {
	return []string{".kt", ".kts"}
}

func (kotlinLanguageModule) NewResolver(ctx *dependencyGraphContext, contentReader vcs.ContentReader) LanguageResolver {
	packageIndex, packageTypes, filePackages := kotlin.BuildKotlinIndices(ctx.kotlinFiles, contentReader)
	return kotlinLanguageResolver{
		ctx:           ctx,
		contentReader: contentReader,
		packageIndex:  packageIndex,
		packageTypes:  packageTypes,
		filePackages:  filePackages,
	}
}

func (kotlinLanguageModule) IsTestFile(filePath string) bool {
	return false
}

type typeScriptLanguageModule struct{}

func (typeScriptLanguageModule) Name() string {
	return "TypeScript"
}

func (typeScriptLanguageModule) Extensions() []string {
	return []string{".ts", ".tsx"}
}

func (typeScriptLanguageModule) NewResolver(ctx *dependencyGraphContext, contentReader vcs.ContentReader) LanguageResolver {
	return typeScriptLanguageResolver{ctx: ctx, contentReader: contentReader}
}

func (typeScriptLanguageModule) IsTestFile(filePath string) bool {
	return typescript.IsTestFile(filePath)
}

type dartLanguageResolver struct {
	ctx           *dependencyGraphContext
	contentReader vcs.ContentReader
}

func (r dartLanguageResolver) ResolveProjectImports(absPath, filePath, ext string) ([]string, error) {
	return dart.ResolveDartProjectImports(absPath, filePath, ext, r.ctx.suppliedFiles, r.contentReader)
}

func (dartLanguageResolver) FinalizeGraph(_ DependencyGraph) error {
	return nil
}

type goLanguageResolver struct {
	ctx             *dependencyGraphContext
	contentReader   vcs.ContentReader
	projectResolver *golang.ProjectImportResolver
}

func (r goLanguageResolver) ResolveProjectImports(absPath, filePath, _ string) ([]string, error) {
	return r.projectResolver.ResolveProjectImports(absPath, filePath)
}

func (r goLanguageResolver) FinalizeGraph(graph DependencyGraph) error {
	return golang.AddGoIntraPackageDependencies(graph, r.ctx.goFiles, r.contentReader)
}

type javaLanguageResolver struct {
	ctx           *dependencyGraphContext
	contentReader vcs.ContentReader
	packageIndex  map[string][]string
	packageTypes  map[string]map[string][]string
	filePackages  map[string]string
}

func (r javaLanguageResolver) ResolveProjectImports(absPath, filePath, _ string) ([]string, error) {
	return java.ResolveJavaProjectImports(
		absPath,
		filePath,
		r.packageIndex,
		r.packageTypes,
		r.filePackages,
		r.ctx.suppliedFiles,
		r.contentReader)
}

func (javaLanguageResolver) FinalizeGraph(_ DependencyGraph) error {
	return nil
}

type kotlinLanguageResolver struct {
	ctx           *dependencyGraphContext
	contentReader vcs.ContentReader
	packageIndex  map[string][]string
	packageTypes  map[string]map[string][]string
	filePackages  map[string]string
}

func (r kotlinLanguageResolver) ResolveProjectImports(absPath, filePath, _ string) ([]string, error) {
	return kotlin.ResolveKotlinProjectImports(
		absPath,
		filePath,
		r.packageIndex,
		r.packageTypes,
		r.filePackages,
		r.ctx.suppliedFiles,
		r.contentReader)
}

func (kotlinLanguageResolver) FinalizeGraph(_ DependencyGraph) error {
	return nil
}

type typeScriptLanguageResolver struct {
	ctx           *dependencyGraphContext
	contentReader vcs.ContentReader
}

func (r typeScriptLanguageResolver) ResolveProjectImports(absPath, filePath, ext string) ([]string, error) {
	return typescript.ResolveTypeScriptProjectImports(absPath, filePath, ext, r.ctx.suppliedFiles, r.contentReader)
}

func (typeScriptLanguageResolver) FinalizeGraph(_ DependencyGraph) error {
	return nil
}
