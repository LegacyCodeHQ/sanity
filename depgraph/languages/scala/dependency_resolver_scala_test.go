package scala

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LegacyCodeHQ/clarity/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveScalaProjectImports(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src", "main", "scala", "com", "example")
	utilDir := filepath.Join(srcDir, "util")
	require.NoError(t, os.MkdirAll(utilDir, 0o755))

	appPath := filepath.Join(srcDir, "App.scala")
	require.NoError(t, os.WriteFile(appPath, []byte(`package com.example

import com.example.util.Helper
import scala.collection.immutable.List

class App
`), 0o644))

	helperPath := filepath.Join(utilDir, "Helper.scala")
	require.NoError(t, os.WriteFile(helperPath, []byte(`package com.example.util

object Helper
`), 0o644))

	reader := vcs.FilesystemContentReader()
	indicesFiles := []string{appPath, helperPath}
	pkgIndex, typeIndex, filePackages := BuildScalaIndices(indicesFiles, reader)
	supplied := map[string]bool{
		appPath:    true,
		helperPath: true,
	}

	imports, err := ResolveScalaProjectImports(appPath, appPath, pkgIndex, typeIndex, filePackages, supplied, reader)
	require.NoError(t, err)
	assert.Equal(t, []string{helperPath}, imports)
}

func TestResolveScalaProjectImports_SamePackageInference(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src", "main", "scala", "com", "example", "model")
	require.NoError(t, os.MkdirAll(srcDir, 0o755))

	cartPath := filepath.Join(srcDir, "Cart.scala")
	require.NoError(t, os.WriteFile(cartPath, []byte(`package com.example.model

class Cart {
  val paymentMethod: PaymentMethod = PaymentMethod.default
}
`), 0o644))

	paymentPath := filepath.Join(srcDir, "PaymentMethod.scala")
	require.NoError(t, os.WriteFile(paymentPath, []byte(`package com.example.model

object PaymentMethod {
  val default: PaymentMethod = new PaymentMethod
}
class PaymentMethod
`), 0o644))

	reader := vcs.FilesystemContentReader()
	files := []string{cartPath, paymentPath}
	pkgIndex, typeIndex, filePackages := BuildScalaIndices(files, reader)
	supplied := map[string]bool{
		cartPath:    true,
		paymentPath: true,
	}

	imports, err := ResolveScalaProjectImports(cartPath, cartPath, pkgIndex, typeIndex, filePackages, supplied, reader)
	require.NoError(t, err)
	assert.Contains(t, imports, paymentPath)
}
