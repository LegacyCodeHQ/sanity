//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

var sanityBinary string

func TestMain(m *testing.M) {
	// Build binary before running tests
	tmpDir, err := os.MkdirTemp("", "sanity-test")
	if err != nil {
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	sanityBinary = filepath.Join(tmpDir, "sanity-test")
	cmd := exec.Command("go", "build", "-o", sanityBinary, ".")
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestGraphOutput(t *testing.T) {
	testCases := []struct {
		name    string
		fixture string
		format  string
		args    []string
	}{
		{"simple-go-dot", "simple-go", "dot", nil},
		{"simple-go-json", "simple-go", "json", nil},
		{"simple-go-mermaid", "simple-go", "mermaid", nil},
		{"simple-dart-dot", "simple-dart", "dot", nil},
		{"simple-dart-json", "simple-dart", "json", nil},
		{"simple-dart-mermaid", "simple-dart", "mermaid", nil},
		{"empty-deps-dot", "empty-deps", "dot", nil},
		{"empty-deps-json", "empty-deps", "json", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixturePath := filepath.Join("testdata", "integration", "fixtures", tc.fixture)
			goldenPath := filepath.Join("testdata", "integration", "golden", tc.name+".golden")

			// Get absolute path for the fixture (needed for consistent output)
			absFixturePath, err := filepath.Abs(fixturePath)
			if err != nil {
				t.Fatalf("failed to get absolute path: %v", err)
			}

			args := []string{"graph", "-i", absFixturePath, "-f", tc.format}
			args = append(args, tc.args...)

			cmd := exec.Command(sanityBinary, args...)
			output, err := cmd.Output()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					t.Fatalf("command failed: %v\nstderr: %s", err, exitErr.Stderr)
				}
				t.Fatalf("command failed: %v", err)
			}

			// Normalize output (remove dynamic content)
			normalized := normalizeOutput(output, absFixturePath)

			if *update {
				// Ensure golden directory exists
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("failed to create golden directory: %v", err)
				}
				if err := os.WriteFile(goldenPath, normalized, 0644); err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				return
			}

			golden, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v\n(run with -update to create)", goldenPath, err)
			}

			if !bytes.Equal(normalized, golden) {
				t.Errorf("output mismatch for %s\n\n--- got ---\n%s\n\n--- want ---\n%s", tc.name, normalized, golden)
			}
		})
	}
}

// normalizeOutput removes dynamic content that varies between runs
func normalizeOutput(output []byte, fixturePath string) []byte {
	result := string(output)

	// Replace absolute fixture path with placeholder
	result = strings.ReplaceAll(result, fixturePath, "{{FIXTURE_PATH}}")

	// Replace commit hashes (7-40 hex chars) with placeholder
	// Only replace in labels, not in entire output (to preserve structure)
	// Look for patterns like "label=" or "title:" followed by commit-like content
	labelPattern := regexp.MustCompile(`(label\s*=\s*"[^"]*•\s*)([a-f0-9]{7,40})(-dirty)?`)
	result = labelPattern.ReplaceAllString(result, "${1}{{COMMIT}}${3}")

	// For mermaid, look for title with commit hash
	mermaidTitlePattern := regexp.MustCompile(`(---\s*title:\s*[^•]*•\s*)([a-f0-9]{7,40})(-dirty)?`)
	result = mermaidTitlePattern.ReplaceAllString(result, "${1}{{COMMIT}}${3}")

	// Sort lines within sections to handle non-deterministic map iteration order
	result = sortOutputLines(result)

	return []byte(result)
}

// sortOutputLines sorts lines in the output to handle non-deterministic ordering
// from Go map iteration. It preserves structure while making output deterministic.
func sortOutputLines(output string) string {
	lines := strings.Split(output, "\n")

	// Detect format and sort accordingly
	if strings.HasPrefix(strings.TrimSpace(output), "digraph") {
		return sortDOTOutput(lines)
	} else if strings.HasPrefix(strings.TrimSpace(output), "---") {
		return sortMermaidOutput(lines)
	} else if strings.HasPrefix(strings.TrimSpace(output), "{") {
		return sortJSONOutput(lines)
	}

	return output
}

// sortDOTOutput sorts DOT format output
func sortDOTOutput(lines []string) string {
	var header []string
	var nodeDefinitions []string
	var edges []string
	var footer []string

	inHeader := true
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines, add to current section
		if trimmed == "" {
			continue
		}

		// Closing brace
		if trimmed == "}" {
			footer = append(footer, line)
			continue
		}

		// Header ends after fontname line or first node definition
		if inHeader && (strings.Contains(trimmed, "[label=") || (strings.HasPrefix(trimmed, "\"") && !strings.Contains(trimmed, "->"))) {
			inHeader = false
		}

		if inHeader {
			header = append(header, line)
		} else if strings.Contains(trimmed, "->") {
			edges = append(edges, line)
		} else if strings.HasPrefix(trimmed, "\"") {
			nodeDefinitions = append(nodeDefinitions, line)
		} else {
			edges = append(edges, line)
		}
	}

	// Sort node definitions and edges
	sort.Strings(nodeDefinitions)
	sort.Strings(edges)

	// Reconstruct output
	var result []string
	result = append(result, header...)
	if len(header) > 0 {
		result = append(result, "")
	}
	result = append(result, nodeDefinitions...)
	if len(nodeDefinitions) > 0 {
		result = append(result, "")
	}
	result = append(result, edges...)
	result = append(result, footer...)

	return strings.Join(result, "\n")
}

// sortMermaidOutput sorts Mermaid format output and normalizes node IDs
func sortMermaidOutput(lines []string) string {
	var header []string
	var nodeDefinitions []string
	var edges []string
	var classDefs []string
	var classAssignments []string

	// First pass: collect all lines and build node ID to label mapping
	nodeIDToLabel := make(map[string]string)
	nodePattern := regexp.MustCompile(`^(\s*)(n\d+)\["([^"]+)"\]`)

	inHeader := true
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Header ends after flowchart line
		if strings.HasPrefix(trimmed, "flowchart") {
			header = append(header, line)
			inHeader = false
			continue
		}

		if inHeader {
			header = append(header, line)
			continue
		}

		// Classify lines and extract node info
		if strings.HasPrefix(trimmed, "classDef") {
			classDefs = append(classDefs, line)
		} else if strings.HasPrefix(trimmed, "class ") {
			classAssignments = append(classAssignments, line)
		} else if strings.Contains(trimmed, "-->") {
			edges = append(edges, line)
		} else if matches := nodePattern.FindStringSubmatch(line); len(matches) > 0 {
			nodeID := matches[2]
			label := matches[3]
			nodeIDToLabel[nodeID] = label
			nodeDefinitions = append(nodeDefinitions, line)
		}
	}

	// Build sorted label list and create new ID mapping
	var labels []string
	for _, label := range nodeIDToLabel {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	// Create mapping from old node ID to new node ID (based on sorted labels)
	labelToNewID := make(map[string]string)
	for i, label := range labels {
		labelToNewID[label] = fmt.Sprintf("n%d", i)
	}

	oldIDToNewID := make(map[string]string)
	for oldID, label := range nodeIDToLabel {
		oldIDToNewID[oldID] = labelToNewID[label]
	}

	// Helper function to replace node IDs in a line
	replaceNodeIDs := func(line string) string {
		result := line
		// Sort old IDs by length descending to avoid partial replacements
		var oldIDs []string
		for oldID := range oldIDToNewID {
			oldIDs = append(oldIDs, oldID)
		}
		sort.Slice(oldIDs, func(i, j int) bool {
			return len(oldIDs[i]) > len(oldIDs[j])
		})

		// Replace with temporary placeholders first to avoid conflicts
		for i, oldID := range oldIDs {
			result = strings.ReplaceAll(result, oldID, fmt.Sprintf("__TEMP_%d__", i))
		}
		// Then replace placeholders with new IDs
		for i, oldID := range oldIDs {
			result = strings.ReplaceAll(result, fmt.Sprintf("__TEMP_%d__", i), oldIDToNewID[oldID])
		}
		return result
	}

	// Apply ID normalization to all relevant lines
	for i, line := range nodeDefinitions {
		nodeDefinitions[i] = replaceNodeIDs(line)
	}
	for i, line := range edges {
		edges[i] = replaceNodeIDs(line)
	}
	for i, line := range classAssignments {
		classAssignments[i] = replaceNodeIDs(line)
	}

	// Sort each section
	sort.Strings(nodeDefinitions)
	sort.Strings(edges)
	sort.Strings(classDefs)

	// Sort class assignments after normalizing node IDs within them
	for i, line := range classAssignments {
		classAssignments[i] = sortClassAssignment(line)
	}
	sort.Strings(classAssignments)

	// Reconstruct output
	var result []string
	result = append(result, header...)
	result = append(result, nodeDefinitions...)
	result = append(result, "")
	result = append(result, edges...)
	result = append(result, "")
	result = append(result, classDefs...)
	result = append(result, classAssignments...)
	result = append(result, "")

	return strings.Join(result, "\n")
}

// sortClassAssignment sorts the class names in a mermaid class assignment line
func sortClassAssignment(line string) string {
	// Line format: "    class n0,n1,n2 newFile"
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "class ") {
		return line
	}

	parts := strings.SplitN(trimmed[6:], " ", 2)
	if len(parts) != 2 {
		return line
	}

	nodeList := strings.Split(parts[0], ",")
	sort.Strings(nodeList)

	indent := strings.TrimSuffix(line, trimmed)
	return indent + "class " + strings.Join(nodeList, ",") + " " + parts[1]
}

// sortJSONOutput sorts JSON output by parsing and re-encoding with sorted keys
func sortJSONOutput(lines []string) string {
	jsonStr := strings.Join(lines, "\n")

	// Parse the JSON into a map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// If parsing fails, return original
		return jsonStr
	}

	// Get sorted keys
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build output with sorted keys
	var sb strings.Builder
	sb.WriteString("{\n")
	for i, key := range keys {
		value := data[key]
		valueBytes, _ := json.Marshal(value)
		sb.WriteString(fmt.Sprintf("  %q: %s", key, string(valueBytes)))
		if i < len(keys)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("}\n")

	return sb.String()
}
