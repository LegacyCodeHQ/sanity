package rust

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func importKey(path string, kind RustImportKind) RustImport {
	return RustImport{Path: path, Kind: kind}
}

func TestParseRustImports(t *testing.T) {
	source := `
use std::io;
use crate::utils::helper as h;
extern crate serde;
mod nested;
`
	imports, err := ParseRustImports([]byte(source))

	require.NoError(t, err)
	assert.Len(t, imports, 4)
	assert.Equal(t, "std::io", imports[0].Path)
	assert.Equal(t, RustImportUse, imports[0].Kind)
	assert.Equal(t, "crate::utils::helper", imports[1].Path)
	assert.Equal(t, RustImportUse, imports[1].Kind)
	assert.Equal(t, "serde", imports[2].Path)
	assert.Equal(t, RustImportExternCrate, imports[2].Kind)
	assert.Equal(t, "nested", imports[3].Path)
	assert.Equal(t, RustImportModDecl, imports[3].Kind)
}

func TestRustImports_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "lib.rs")

	content := `
use std::fmt;
`
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	imports, err := RustImports(tmpFile)

	require.NoError(t, err)
	assert.Len(t, imports, 1)
	assert.Equal(t, "std::fmt", imports[0].Path)
	assert.Equal(t, RustImportUse, imports[0].Kind)
}

func TestParseRustImports_FiltersNestedImports(t *testing.T) {
	source := `
use crate::top::level;

fn helper() {
  use crate::nested::only;
}

mod nested;
`
	imports, err := ParseRustImports([]byte(source))
	require.NoError(t, err)

	assert.Len(t, imports, 2)
	assert.Equal(t, "crate::top::level", imports[0].Path)
	assert.Equal(t, RustImportUse, imports[0].Kind)
	assert.Equal(t, "nested", imports[1].Path)
	assert.Equal(t, RustImportModDecl, imports[1].Kind)
}

func TestParseRustImports_VisibilityAndScopedUseList(t *testing.T) {
	source := `
#[cfg(feature = "x")]
pub(crate) use crate::alpha::{beta, gamma};
pub mod public_mod;
`
	imports, err := ParseRustImports([]byte(source))
	require.NoError(t, err)

	assert.Len(t, imports, 2)
	assert.Equal(t, "crate::alpha", imports[0].Path)
	assert.Equal(t, RustImportUse, imports[0].Kind)
	assert.Equal(t, "public_mod", imports[1].Path)
	assert.Equal(t, RustImportModDecl, imports[1].Kind)
}

func TestParseRustImports_CollectsQualifiedPathReferences(t *testing.T) {
	source := `
use crate::alpha::beta;

fn run() {
  crate_b::foo::run();
  crate::core::do_work();
  let x = "crate_b::ignored::in_string";
  // crate_b::ignored::in_comment
}
`
	imports, err := ParseRustImports([]byte(source))
	require.NoError(t, err)

	assert.Contains(t, imports, importKey("crate_b::foo::run", RustImportUse))
	assert.Contains(t, imports, importKey("crate::core::do_work", RustImportUse))
	assert.Contains(t, imports, importKey("crate::alpha::beta", RustImportUse))
}

func TestParseRustImports_CollectsQualifiedPathsWhenLifetimesPresent(t *testing.T) {
	source := `
const fn marker() -> &'static str { "ok" }

fn run() {
  s7e_parser::analyze();
  s7e_flow::build_flow_graph();
}
`
	imports, err := ParseRustImports([]byte(source))
	require.NoError(t, err)

	assert.Contains(t, imports, importKey("s7e_parser::analyze", RustImportUse))
	assert.Contains(t, imports, importKey("s7e_flow::build_flow_graph", RustImportUse))
}
