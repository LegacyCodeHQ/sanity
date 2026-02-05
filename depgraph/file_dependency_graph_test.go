package depgraph_test

import (
	"testing"

	"github.com/LegacyCodeHQ/sanity/depgraph"
	"github.com/LegacyCodeHQ/sanity/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileDependencyGraph(t *testing.T) {
	graph := depgraph.MustDependencyGraph(map[string][]string{
		"/project/main.go":        {"/project/utils.go"},
		"/project/main_test.go":   {"/project/main.go"},
		"/project/utils.go":       {},
		"/project/helper_test.go": {},
	})

	stats := map[string]vcs.FileStats{
		"/project/main.go": {
			Additions: 3,
			Deletions: 1,
		},
	}

	fileGraph, err := depgraph.NewFileDependencyGraph(graph, stats)
	require.NoError(t, err)

	mainMeta, ok := fileGraph.Meta.Files["/project/main.go"]
	require.True(t, ok)
	require.NotNil(t, mainMeta.Stats)
	assert.Equal(t, 3, mainMeta.Stats.Additions)
	assert.Equal(t, ".go", mainMeta.Extension)
	assert.False(t, mainMeta.IsTest)

	testMeta, ok := fileGraph.Meta.Files["/project/main_test.go"]
	require.True(t, ok)
	assert.True(t, testMeta.IsTest)
	assert.Equal(t, ".go", testMeta.Extension)
	assert.Nil(t, testMeta.Stats)

	_, hasEdge := fileGraph.Meta.Edges[depgraph.FileEdge{From: "/project/main.go", To: "/project/utils.go"}]
	assert.True(t, hasEdge)
}
