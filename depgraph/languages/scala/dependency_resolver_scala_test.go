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

func TestResolveScalaProjectImports_DoesNotLinkToSplitPackageTestFile(t *testing.T) {
	tmpDir := t.TempDir()
	mainDir := filepath.Join(tmpDir, "core", "src", "main", "scala", "cats", "instances")
	kernelMainDir := filepath.Join(tmpDir, "kernel", "src", "main", "scala", "cats", "kernel")
	kernelTestDir := filepath.Join(tmpDir, "kernel-laws", "shared", "src", "test", "scala", "cats", "kernel", "laws")
	require.NoError(t, os.MkdirAll(mainDir, 0o755))
	require.NoError(t, os.MkdirAll(kernelMainDir, 0o755))
	require.NoError(t, os.MkdirAll(kernelTestDir, 0o755))

	sortedMapPath := filepath.Join(mainDir, "sortedMap.scala")
	require.NoError(t, os.WriteFile(sortedMapPath, []byte(`package cats.instances

import cats.kernel.{CommutativeMonoid, CommutativeSemigroup}

trait SortedMapInstances
`), 0o644))

	// Intentionally no top-level CommutativeMonoid/Semigroup declarations so
	// resolver fallback is exercised for cats.kernel imports.
	kernelPackagePath := filepath.Join(kernelMainDir, "package.scala")
	require.NoError(t, os.WriteFile(kernelPackagePath, []byte(`package cats.kernel

object Placeholder
`), 0o644))

	lawTestsPath := filepath.Join(kernelTestDir, "LawTests.scala")
	require.NoError(t, os.WriteFile(lawTestsPath, []byte(`package cats.kernel
package laws

object KernelCheck
`), 0o644))

	reader := vcs.FilesystemContentReader()
	files := []string{sortedMapPath, kernelPackagePath, lawTestsPath}
	pkgIndex, typeIndex, filePackages := BuildScalaIndices(files, reader)
	supplied := map[string]bool{
		sortedMapPath:     true,
		kernelPackagePath: true,
		lawTestsPath:      true,
	}

	imports, err := ResolveScalaProjectImports(sortedMapPath, sortedMapPath, pkgIndex, typeIndex, filePackages, supplied, reader)
	require.NoError(t, err)
	assert.NotContains(t, imports, lawTestsPath, "production file should not depend on split-package test file")
}

func TestResolveScalaProjectImports_DoesNotFanOutToPackageOnUnresolvedImportSymbols(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "core", "src", "main", "scala", "cats", "data")
	instancesDir := filepath.Join(tmpDir, "core", "src", "main", "scala", "cats", "instances")
	require.NoError(t, os.MkdirAll(dataDir, 0o755))
	require.NoError(t, os.MkdirAll(instancesDir, 0o755))

	sortedMapPath := filepath.Join(instancesDir, "sortedMap.scala")
	require.NoError(t, os.WriteFile(sortedMapPath, []byte(`package cats.instances

import cats.data.{Chain, Ior}

trait SortedMapInstances
`), 0o644))

	nonEmptyListPath := filepath.Join(dataDir, "NonEmptyList.scala")
	require.NoError(t, os.WriteFile(nonEmptyListPath, []byte(`package cats.data

final case class NonEmptyList[A](head: A)
`), 0o644))

	nonEmptyVectorPath := filepath.Join(dataDir, "NonEmptyVector.scala")
	require.NoError(t, os.WriteFile(nonEmptyVectorPath, []byte(`package cats.data

final case class NonEmptyVector[A](value: Vector[A])
`), 0o644))

	nonEmptyLazyListPath := filepath.Join(dataDir, "NonEmptyLazyList.scala")
	require.NoError(t, os.WriteFile(nonEmptyLazyListPath, []byte(`package cats.data

final case class NonEmptyLazyList[A](value: LazyList[A])
`), 0o644))

	reader := vcs.FilesystemContentReader()
	files := []string{sortedMapPath, nonEmptyListPath, nonEmptyVectorPath, nonEmptyLazyListPath}
	pkgIndex, typeIndex, filePackages := BuildScalaIndices(files, reader)
	supplied := map[string]bool{
		sortedMapPath:        true,
		nonEmptyListPath:     true,
		nonEmptyVectorPath:   true,
		nonEmptyLazyListPath: true,
	}

	imports, err := ResolveScalaProjectImports(sortedMapPath, sortedMapPath, pkgIndex, typeIndex, filePackages, supplied, reader)
	require.NoError(t, err)
	assert.NotContains(t, imports, nonEmptyListPath, "should not link unrelated cats.data file on unresolved symbol")
	assert.NotContains(t, imports, nonEmptyVectorPath, "should not link unrelated cats.data file on unresolved symbol")
	assert.NotContains(t, imports, nonEmptyLazyListPath, "should not link unrelated cats.data file on unresolved symbol")
}

func TestResolveScalaProjectImports_DoesNotLinkToUnrelatedPackageFileWhenImportUnresolved(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "core", "src", "main", "scala", "cats", "data")
	instancesDir := filepath.Join(tmpDir, "core", "src", "main", "scala", "cats", "instances")
	require.NoError(t, os.MkdirAll(dataDir, 0o755))
	require.NoError(t, os.MkdirAll(instancesDir, 0o755))

	nonEmptyListPath := filepath.Join(dataDir, "NonEmptyList.scala")
	require.NoError(t, os.WriteFile(nonEmptyListPath, []byte(`package cats.data

import cats.instances.StaticMethods

final case class NonEmptyList[A](head: A)
`), 0o644))

	// No StaticMethods declaration is present in this input set.
	sortedMapPath := filepath.Join(instancesDir, "sortedMap.scala")
	require.NoError(t, os.WriteFile(sortedMapPath, []byte(`package cats.instances

trait SortedMapInstances
`), 0o644))

	reader := vcs.FilesystemContentReader()
	files := []string{nonEmptyListPath, sortedMapPath}
	pkgIndex, typeIndex, filePackages := BuildScalaIndices(files, reader)
	supplied := map[string]bool{
		nonEmptyListPath: true,
		sortedMapPath:    true,
	}

	imports, err := ResolveScalaProjectImports(nonEmptyListPath, nonEmptyListPath, pkgIndex, typeIndex, filePackages, supplied, reader)
	require.NoError(t, err)
	assert.NotContains(t, imports, sortedMapPath, "should not link arbitrary package peer when imported symbol cannot be resolved")
}

func TestResolveScalaProjectImports_WildcardPackageImportResolvesToPackageObject(t *testing.T) {
	tmpDir := t.TempDir()
	instancesDir := filepath.Join(tmpDir, "core", "src", "main", "scala", "cats", "instances")
	rootCatsDir := filepath.Join(tmpDir, "core", "src", "main", "scala", "cats")
	require.NoError(t, os.MkdirAll(instancesDir, 0o755))
	require.NoError(t, os.MkdirAll(rootCatsDir, 0o755))

	sortedMapPath := filepath.Join(instancesDir, "sortedMap.scala")
	require.NoError(t, os.WriteFile(sortedMapPath, []byte(`package cats.instances

import cats.*

trait SortedMapInstances
`), 0o644))

	packageObjectPath := filepath.Join(rootCatsDir, "package.scala")
	require.NoError(t, os.WriteFile(packageObjectPath, []byte(`package object cats {
  type Id[A] = A
}
`), 0o644))

	reader := vcs.FilesystemContentReader()
	files := []string{sortedMapPath, packageObjectPath}
	pkgIndex, typeIndex, filePackages := BuildScalaIndices(files, reader)
	supplied := map[string]bool{
		sortedMapPath:     true,
		packageObjectPath: true,
	}

	imports, err := ResolveScalaProjectImports(sortedMapPath, sortedMapPath, pkgIndex, typeIndex, filePackages, supplied, reader)
	require.NoError(t, err)
	assert.Contains(t, imports, packageObjectPath, "wildcard cats import should resolve to package object cats")
}
