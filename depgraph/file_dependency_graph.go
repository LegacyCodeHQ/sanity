package depgraph

import (
	"path/filepath"
	"sort"

	"github.com/LegacyCodeHQ/sanity/vcs"
)

// FileDependencyGraph wraps a dependency graph with file-level metadata.
type FileDependencyGraph struct {
	Graph DependencyGraph
	Meta  FileGraphMetadata
}

// FileGraphMetadata contains metadata keyed by file and edge.
type FileGraphMetadata struct {
	Files  map[string]FileMetadata
	Edges  map[FileEdge]EdgeMetadata
	Cycles []FileCycle
}

// FileMetadata holds metadata for a single file node.
type FileMetadata struct {
	Stats     *vcs.FileStats
	IsTest    bool
	Extension string
}

// FileEdge identifies a directed edge between two files.
type FileEdge struct {
	From string
	To   string
}

// EdgeMetadata holds metadata for a graph edge.
type EdgeMetadata struct {
	InCycle bool
}

// FileCycle describes a canonical cycle path.
type FileCycle struct {
	Path []string
}

// NewFileDependencyGraph creates a file-annotated graph from a dependency graph and optional file stats.
func NewFileDependencyGraph(g DependencyGraph, fileStats map[string]vcs.FileStats) (FileDependencyGraph, error) {
	adjacency, err := AdjacencyList(g)
	if err != nil {
		return FileDependencyGraph{}, err
	}

	files := make(map[string]FileMetadata, len(adjacency))
	edges := make(map[FileEdge]EdgeMetadata)

	nodes := make([]string, 0, len(adjacency))
	for node := range adjacency {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	for _, node := range nodes {
		md := FileMetadata{
			IsTest:    IsTestFile(node),
			Extension: filepath.Ext(filepath.Base(node)),
		}

		if fileStats != nil {
			if stats, ok := fileStats[node]; ok {
				statsCopy := stats
				md.Stats = &statsCopy
			}
		}

		files[node] = md

		for _, dep := range adjacency[node] {
			edges[FileEdge{From: node, To: dep}] = EdgeMetadata{}
		}
	}

	return FileDependencyGraph{
		Graph: g,
		Meta: FileGraphMetadata{
			Files:  files,
			Edges:  edges,
			Cycles: nil,
		},
	}, nil
}
