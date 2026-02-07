package csharp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LegacyCodeHQ/sanity/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCSharpProjectImports(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "App"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "Lib"), 0o755))

	programPath := filepath.Join(tmpDir, "App", "Program.cs")
	require.NoError(t, os.WriteFile(programPath, []byte(`using Lib.Core;

namespace App;

public class Program
{
    private Logger logger;
    public void Run()
    {
        var fileLogger = new FileLogger("log.txt");
    }
}
`), 0o644))

	loggerPath := filepath.Join(tmpDir, "Lib", "Logger.cs")
	require.NoError(t, os.WriteFile(loggerPath, []byte(`namespace Lib.Core;
public class Logger {}
`), 0o644))

	fileLoggerPath := filepath.Join(tmpDir, "Lib", "FileLogger.cs")
	require.NoError(t, os.WriteFile(fileLoggerPath, []byte(`namespace Lib.Core;
public class FileLogger {}
`), 0o644))

	helperPath := filepath.Join(tmpDir, "App", "Helper.cs")
	require.NoError(t, os.WriteFile(helperPath, []byte(`namespace App;
public class Helper {}
`), 0o644))

	supplied := map[string]bool{
		programPath:    true,
		loggerPath:     true,
		fileLoggerPath: true,
		helperPath:     true,
	}
	reader := vcs.FilesystemContentReader()
	namespaceToFiles, namespaceToTypes, fileToNamespace := BuildCSharpIndices(supplied, reader)

	imports, err := ResolveCSharpProjectImports(
		programPath,
		programPath,
		namespaceToFiles,
		namespaceToTypes,
		fileToNamespace,
		supplied,
		reader)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{loggerPath, fileLoggerPath}, imports)
}

func TestResolveCSharpProjectImports_SkipsAmbiguousType(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "A"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "B"), 0o755))

	programPath := filepath.Join(tmpDir, "Program.cs")
	require.NoError(t, os.WriteFile(programPath, []byte(`using Shared;
public class Program { private Service svc; }
`), 0o644))

	serviceAPath := filepath.Join(tmpDir, "A", "Service.cs")
	require.NoError(t, os.WriteFile(serviceAPath, []byte(`namespace Shared;
public class Service {}
`), 0o644))

	serviceBPath := filepath.Join(tmpDir, "B", "Service.cs")
	require.NoError(t, os.WriteFile(serviceBPath, []byte(`namespace Shared;
public class Service {}
`), 0o644))

	supplied := map[string]bool{
		programPath:  true,
		serviceAPath: true,
		serviceBPath: true,
	}
	reader := vcs.FilesystemContentReader()
	namespaceToFiles, namespaceToTypes, fileToNamespace := BuildCSharpIndices(supplied, reader)

	imports, err := ResolveCSharpProjectImports(
		programPath,
		programPath,
		namespaceToFiles,
		namespaceToTypes,
		fileToNamespace,
		supplied,
		reader)
	require.NoError(t, err)
	assert.Empty(t, imports)
}
