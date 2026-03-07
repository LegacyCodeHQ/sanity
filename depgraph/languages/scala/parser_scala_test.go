package scala

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePackageDeclaration(t *testing.T) {
	src := []byte(`package com.example.core

class App
`)
	assert.Equal(t, "com.example.core", ParsePackageDeclaration(src))
}

func TestParseScalaImports_ClassifiesInternalAndStandard(t *testing.T) {
	src := []byte(`package com.example

import com.example.util.Helper
import scala.collection.mutable.Map
import org.typelevel.cats.Eval
import com.example.model.{User, Account => Acc, _}
`)
	projectPackages := map[string]bool{
		"com.example":       true,
		"com.example.util":  true,
		"com.example.model": true,
	}

	imports := ParseScalaImports(src, projectPackages)
	require.Len(t, imports, 6)

	_, isInternal := imports[0].(InternalImport)
	_, isStandard := imports[1].(StandardLibraryImport)
	_, isExternal := imports[2].(ExternalImport)
	_, isInternalSelector := imports[3].(InternalImport)
	_, isInternalAlias := imports[4].(InternalImport)
	_, isInternalWildcard := imports[5].(InternalImport)

	assert.True(t, isInternal)
	assert.True(t, isStandard)
	assert.True(t, isExternal)
	assert.True(t, isInternalSelector)
	assert.True(t, isInternalAlias)
	assert.True(t, isInternalWildcard)
	assert.True(t, imports[5].IsWildcard())
}

func TestParseTopLevelTypeNames(t *testing.T) {
	src := []byte(`package com.example

class App
trait Service
object Helpers
enum Mode { case On, Off }
`)
	types := ParseTopLevelTypeNames(src)
	assert.ElementsMatch(t, []string{"App", "Service", "Helpers", "Mode"}, types)
}

func TestIsTestFile(t *testing.T) {
	assert.True(t, IsTestFile("/project/src/test/scala/com/example/AppTest.scala"))
	assert.True(t, IsTestFile("/project/module/test/com/example/AppTests.scala"))
	assert.False(t, IsTestFile("/project/src/main/scala/com/example/App.scala"))
}

func TestExtractTypeIdentifiers(t *testing.T) {
	src := []byte(`package com.example.model

import com.example.util.Helper

class Cart {
  val paymentMethod: PaymentMethod = PaymentMethod.default
}
`)

	identifiers := ExtractTypeIdentifiers(src)
	assert.Contains(t, identifiers, "Cart")
	assert.Contains(t, identifiers, "PaymentMethod")
	assert.NotContains(t, identifiers, "Helper")
}
