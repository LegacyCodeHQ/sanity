package depgraph

type languageRegistryEntry struct {
	Module LanguageModule
}

// languageRegistry is the single source of truth for supported languages.
// Adding/removing a language should happen here.
var languageRegistry = []languageRegistryEntry{
	{Module: dartLanguageModule{}},
	{Module: goLanguageModule{}},
	{Module: javaLanguageModule{}},
	{Module: kotlinLanguageModule{}},
	{Module: typeScriptLanguageModule{}},
}

func moduleForExtension(ext string) (LanguageModule, bool) {
	for _, language := range languageRegistry {
		for _, languageExt := range language.Module.Extensions() {
			if languageExt == ext {
				return language.Module, true
			}
		}
	}

	return nil, false
}
