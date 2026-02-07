package ruby

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRubyImports(t *testing.T) {
	source := `
require "json"
require_relative "../lib/core"
require('set')
# require "ignored"
`

	imports, err := ParseRubyImports([]byte(source))

	require.NoError(t, err)
	assert.Len(t, imports, 3)
	assert.Equal(t, "json", imports[0].Path())
	assert.False(t, imports[0].IsRelative())
	assert.Equal(t, "../lib/core", imports[1].Path())
	assert.True(t, imports[1].IsRelative())
	assert.Equal(t, "set", imports[2].Path())
}

func TestRubyImports_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "app.rb")

	err := os.WriteFile(tmpFile, []byte("require 'json'\n"), 0644)
	require.NoError(t, err)

	imports, err := RubyImports(tmpFile)

	require.NoError(t, err)
	assert.Len(t, imports, 1)
	assert.Equal(t, "json", imports[0].Path())
}

func TestResolveRubyImportPath(t *testing.T) {
	suppliedFiles := map[string]bool{
		"/project/lib/core.rb":        true,
		"/project/app/models/user.rb": true,
	}

	rel := ResolveRubyImportPath("/project/app/main.rb", RubyImport{path: "../lib/core", isRelative: true}, suppliedFiles)
	assert.Equal(t, []string{"/project/lib/core.rb"}, rel)

	abs := ResolveRubyImportPath("/project/app/main.rb", RubyImport{path: "app/models/user", isRelative: false}, suppliedFiles)
	assert.Equal(t, []string{"/project/app/models/user.rb"}, abs)
}

func TestIsTestFile(t *testing.T) {
	assert.True(t, IsTestFile("spec/user_spec.rb"))
	assert.True(t, IsTestFile("test/user_test.rb"))
	assert.False(t, IsTestFile("lib/user.rb"))
}
