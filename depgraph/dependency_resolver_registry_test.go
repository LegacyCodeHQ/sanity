package depgraph

import "testing"

func TestDefaultResolverSupportsAllRegisteredLanguageExtensions(t *testing.T) {
	resolver := NewDefaultDependencyResolver(&dependencyGraphContext{}, nil)

	for _, ext := range SupportedLanguageExtensions() {
		if !resolver.SupportsFileExtension(ext) {
			t.Fatalf("resolver does not support registered extension %q", ext)
		}
	}
}
