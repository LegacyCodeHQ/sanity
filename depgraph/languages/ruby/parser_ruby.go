package ruby

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// RubyImport represents a require/require_relative in a Ruby file.
type RubyImport struct {
	path       string
	isRelative bool
}

func (i RubyImport) Path() string {
	return i.path
}

func (i RubyImport) IsRelative() bool {
	return i.isRelative
}

// RubyImports parses a Ruby file and returns its imports.
func RubyImports(filePath string) ([]RubyImport, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseRubyImports(sourceCode)
}

// ParseRubyImports parses Ruby source code and extracts require directives.
func ParseRubyImports(sourceCode []byte) ([]RubyImport, error) {
	var imports []RubyImport

	scanner := bufio.NewScanner(bytes.NewReader(sourceCode))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if imp, ok := parseRubyImportLine(line, "require_relative", true); ok {
			imports = append(imports, imp)
			continue
		}

		if imp, ok := parseRubyImportLine(line, "require", false); ok {
			imports = append(imports, imp)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed scanning Ruby source: %w", err)
	}

	return imports, nil
}

func parseRubyImportLine(line, keyword string, isRelative bool) (RubyImport, bool) {
	if !strings.HasPrefix(line, keyword) {
		return RubyImport{}, false
	}

	rest := strings.TrimSpace(strings.TrimPrefix(line, keyword))
	rest = strings.TrimPrefix(rest, "(")
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return RubyImport{}, false
	}

	quote := rest[0]
	if quote != '\'' && quote != '"' {
		return RubyImport{}, false
	}

	end := strings.IndexByte(rest[1:], quote)
	if end < 0 {
		return RubyImport{}, false
	}

	path := strings.TrimSpace(rest[1 : end+1])
	if path == "" {
		return RubyImport{}, false
	}

	return RubyImport{path: path, isRelative: isRelative}, true
}

// ResolveRubyImportPath resolves a Ruby require path to possible file paths.
func ResolveRubyImportPath(sourceFile string, imp RubyImport, suppliedFiles map[string]bool) []string {
	if imp.IsRelative() {
		return resolveRelativeRubyImportPath(sourceFile, imp.Path(), suppliedFiles)
	}
	return resolveAbsoluteRubyImportPath(imp.Path(), suppliedFiles)
}

func resolveRelativeRubyImportPath(sourceFile, importPath string, suppliedFiles map[string]bool) []string {
	sourceDir := filepath.Dir(sourceFile)
	basePath := filepath.Clean(filepath.Join(sourceDir, importPath))

	candidates := rubyImportCandidates(basePath)
	return existingPaths(candidates, suppliedFiles)
}

func resolveAbsoluteRubyImportPath(importPath string, suppliedFiles map[string]bool) []string {
	cleanPath := strings.TrimSpace(filepath.ToSlash(importPath))
	if cleanPath == "" {
		return []string{}
	}

	withoutExt := strings.TrimSuffix(cleanPath, filepath.Ext(cleanPath))
	if withoutExt == "" {
		withoutExt = cleanPath
	}

	var resolved []string
	for filePath, exists := range suppliedFiles {
		if !exists {
			continue
		}
		if filepath.Ext(filePath) != ".rb" {
			continue
		}

		normalizedFile := filepath.ToSlash(filePath)
		if strings.HasSuffix(normalizedFile, "/"+withoutExt+".rb") ||
			strings.HasSuffix(normalizedFile, "/"+cleanPath) {
			resolved = append(resolved, filePath)
		}
	}

	sort.Strings(resolved)
	return resolved
}

func rubyImportCandidates(basePath string) []string {
	if filepath.Ext(basePath) == ".rb" {
		return []string{basePath}
	}

	return []string{
		basePath + ".rb",
		filepath.Join(basePath, "init.rb"),
	}
}

func existingPaths(candidates []string, suppliedFiles map[string]bool) []string {
	resolved := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if suppliedFiles[candidate] {
			resolved = append(resolved, candidate)
		}
	}
	return resolved
}

var rubyConstantReferencePattern = regexp.MustCompile(`(?:^|[^A-Za-z0-9_:])(::)?([A-Z][A-Za-z0-9_]*(?:::[A-Z][A-Za-z0-9_]*)+)`)

// ParseRubyConstantReferences extracts qualified constant references from Ruby source.
// Examples: ActiveSupport::Cache::Coder, ::JSON::ParserError.
func ParseRubyConstantReferences(sourceCode []byte) []string {
	matches := rubyConstantReferencePattern.FindAllSubmatchIndex(sourceCode, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(matches))
	refs := make([]string, 0, len(matches))
	for _, m := range matches {
		// m[4]:m[5] are the start/end indices of capture group 2
		if m[4] < 0 {
			continue
		}
		ref := string(sourceCode[m[4]:m[5]])
		if ref == "" {
			continue
		}
		if _, ok := seen[ref]; ok {
			continue
		}
		seen[ref] = struct{}{}
		refs = append(refs, ref)
	}

	return refs
}

// ResolveRubyConstantReferencePath maps a qualified constant reference to a concrete file path.
// Returns no paths when resolution is ambiguous.
func ResolveRubyConstantReferencePath(ref string, suppliedFiles map[string]bool) []string {
	normalized := strings.TrimPrefix(strings.TrimSpace(ref), "::")
	if normalized == "" || !strings.Contains(normalized, "::") {
		return nil
	}

	segments := strings.Split(normalized, "::")
	if len(segments) < 2 {
		return nil
	}

	for i, segment := range segments {
		segments[i] = camelToSnake(segment)
	}

	for end := len(segments); end >= 2; end-- {
		if resolved := resolveRubyConstantSegments(segments[:end], suppliedFiles); len(resolved) == 1 {
			return resolved
		}
	}

	return nil
}

func resolveRubyConstantSegments(segments []string, suppliedFiles map[string]bool) []string {
	bestPath := ""
	bestGaps := 0
	bestTrailing := 0
	bestLeading := 0
	tie := false

	for filePath, exists := range suppliedFiles {
		if !exists || filepath.Ext(filePath) != ".rb" {
			continue
		}

		pathParts := rubyPathComponents(filePath)
		gaps, leading, trailing, ok := subsequenceMatch(pathParts, segments)
		if !ok {
			continue
		}

		if bestPath == "" || isBetterConstantPathMatch(gaps, trailing, leading, bestGaps, bestTrailing, bestLeading) {
			bestPath = filePath
			bestGaps = gaps
			bestTrailing = trailing
			bestLeading = leading
			tie = false
			continue
		}

		if gaps == bestGaps && trailing == bestTrailing && leading == bestLeading {
			tie = true
		}
	}

	if bestPath == "" || tie {
		return nil
	}

	return []string{bestPath}
}

func resolveRubyConstantSegmentsCached(segments []string, pathComponentsCache map[string][]string) []string {
	bestPath := ""
	bestGaps := 0
	bestTrailing := 0
	bestLeading := 0
	tie := false

	for filePath, pathParts := range pathComponentsCache {
		gaps, leading, trailing, ok := subsequenceMatch(pathParts, segments)
		if !ok {
			continue
		}

		if bestPath == "" || isBetterConstantPathMatch(gaps, trailing, leading, bestGaps, bestTrailing, bestLeading) {
			bestPath = filePath
			bestGaps = gaps
			bestTrailing = trailing
			bestLeading = leading
			tie = false
			continue
		}

		if gaps == bestGaps && trailing == bestTrailing && leading == bestLeading {
			tie = true
		}
	}

	if bestPath == "" || tie {
		return nil
	}

	return []string{bestPath}
}

func resolveRubyConstantReferencePathCached(ref string, pathComponentsCache map[string][]string) []string {
	normalized := strings.TrimPrefix(strings.TrimSpace(ref), "::")
	if normalized == "" || !strings.Contains(normalized, "::") {
		return nil
	}

	segments := strings.Split(normalized, "::")
	if len(segments) < 2 {
		return nil
	}

	for i, segment := range segments {
		segments[i] = camelToSnake(segment)
	}

	for end := len(segments); end >= 2; end-- {
		if resolved := resolveRubyConstantSegmentsCached(segments[:end], pathComponentsCache); len(resolved) == 1 {
			return resolved
		}
	}

	return nil
}

func camelToSnake(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 4)

	for i := 0; i < len(s); i++ {
		c := s[i]
		if i > 0 && c >= 'A' && c <= 'Z' {
			prev := s[i-1]
			nextLower := i+1 < len(s) && s[i+1] >= 'a' && s[i+1] <= 'z'
			if (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9') || ((prev >= 'A' && prev <= 'Z') && nextLower) {
				b.WriteByte('_')
			}
		}
		if c >= 'A' && c <= 'Z' {
			b.WriteByte(c + 32)
		} else {
			b.WriteByte(c)
		}
	}

	return b.String()
}

func rubyPathComponents(filePath string) []string {
	normalized := filepath.ToSlash(strings.TrimSuffix(filePath, ".rb"))
	parts := strings.Split(normalized, "/")
	out := parts[:0]
	for _, part := range parts {
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func subsequenceMatch(pathParts, targetParts []string) (gaps, leading, trailing int, ok bool) {
	if len(targetParts) == 0 {
		return 0, 0, 0, false
	}

	next := 0
	firstIdx := -1
	prevIdx := -1

	for i, part := range pathParts {
		if next >= len(targetParts) {
			break
		}
		if part != targetParts[next] {
			continue
		}
		if firstIdx == -1 {
			firstIdx = i
		} else {
			gaps += i - prevIdx - 1
		}
		prevIdx = i
		next++
	}

	if next != len(targetParts) {
		return 0, 0, 0, false
	}

	leading = firstIdx
	trailing = len(pathParts) - 1 - prevIdx
	return gaps, leading, trailing, true
}

func isBetterConstantPathMatch(gaps, trailing, leading, bestGaps, bestTrailing, bestLeading int) bool {
	if gaps != bestGaps {
		return gaps < bestGaps
	}
	if trailing != bestTrailing {
		return trailing < bestTrailing
	}
	if leading != bestLeading {
		return leading < bestLeading
	}
	return false
}
