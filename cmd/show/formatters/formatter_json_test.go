package formatters

import (
	"testing"

	"github.com/LegacyCodeHQ/clarity/internal/testhelpers"
	"github.com/LegacyCodeHQ/clarity/vcs"
	"github.com/stretchr/testify/require"
)

func TestJSONFormatter_Format(t *testing.T) {
	graph := testFileGraph(t, map[string][]string{
		"/project/main.go":  {"/project/utils.go"},
		"/project/utils.go": {},
	}, map[string]vcs.FileStats{
		"/project/main.go": {
			IsNew:     true,
			Additions: 3,
			Deletions: 1,
		},
	})

	formatter := jsonFormatter{}
	output, err := formatter.Format(graph, RenderOptions{Label: "test-label"})
	require.NoError(t, err)

	g := testhelpers.JSONGoldie(t)
	g.Assert(t, t.Name(), []byte(output))
}

func TestJSONFormatter_GenerateURL_NotSupported(t *testing.T) {
	formatter := jsonFormatter{}
	url, ok := formatter.GenerateURL(`{"nodes":[]}`)
	require.False(t, ok)
	require.Empty(t, url)
}

func TestJSONFormatter_Format_TestFileAttribute(t *testing.T) {
	graph := testFileGraph(t, map[string][]string{
		"/project/main.go":      {},
		"/project/main_test.go": {"/project/main.go"},
	}, nil)

	formatter := jsonFormatter{}
	output, err := formatter.Format(graph, RenderOptions{Label: "test-label"})
	require.NoError(t, err)

	g := testhelpers.JSONGoldie(t)
	g.Assert(t, t.Name(), []byte(output))
}
