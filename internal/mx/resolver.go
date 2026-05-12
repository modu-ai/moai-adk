package mx

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// Resolver provides AnchorID resolution services.
// This is a placeholder for SPC-004 which will implement full fan-in analysis.
type Resolver struct {
	manager *Manager
}

// NewResolver creates a new AnchorID resolver.
func NewResolver(manager *Manager) *Resolver {
	return &Resolver{
		manager: manager,
	}
}

// ResolveAnchor looks up an AnchorID and returns the corresponding tag.
// Returns error if AnchorID is not found.
//
// @MX:ANCHOR: [AUTO] ResolveAnchor — invariant contract of AnchorID single lookup API
// @MX:REASON: fan_in >= 3 — ResolveAll internal calls, CLI, code map generation tools usage
func (r *Resolver) ResolveAnchor(anchorID string) (Tag, error) {
	tags := r.manager.GetAllTags()

	for _, tag := range tags {
		if tag.Kind == MXAnchor && tag.AnchorID == anchorID {
			return tag, nil
		}
	}

	return Tag{}, fmt.Errorf("anchor ID not found: %s", anchorID)
}

// ResolveAll returns all tags for a given set of AnchorIDs.
func (r *Resolver) ResolveAll(anchorIDs []string) ([]Tag, error) {
	var result []Tag
	var missing []string

	for _, anchorID := range anchorIDs {
		tag, err := r.ResolveAnchor(anchorID)
		if err != nil {
			missing = append(missing, anchorID)
		} else {
			result = append(result, tag)
		}
	}

	if len(missing) > 0 {
		return result, fmt.Errorf("missing anchors: %v", missing)
	}

	return result, nil
}

// ListAnchors returns all ANCHOR tags in the project, sorted by file path.
func (r *Resolver) ListAnchors() []Tag {
	tags := r.manager.GetAllTags()

	var anchors []Tag
	for _, tag := range tags {
		if tag.Kind == MXAnchor {
			anchors = append(anchors, tag)
		}
	}

	// Sort by file path for consistent output
	sort.Slice(anchors, func(i, j int) bool {
		return anchors[i].File < anchors[j].File
	})

	return anchors
}

// ResolveAnchorCallsites returns the list of locations that reference the given anchorID.
// When lspClient is non-nil and available, it uses textDocument/references for precise results
// (Method="lsp"). Otherwise it falls back to a text-based walk (Method="textual").
// includeTests controls whether _test.go and testdata paths are included.
//
// Existing ResolveAnchor signature is unchanged (backward compat preserved).
//
// @MX:ANCHOR: [AUTO] ResolveAnchorCallsites — additive location-aware variant of ResolveAnchor
// @MX:REASON: fan_in >= 3 — CLI mx_query, codemaps, M5 sweep will all call this for callsite drill-down
func (r *Resolver) ResolveAnchorCallsites(
	ctx context.Context,
	anchorID string,
	projectRoot string,
	includeTests bool,
	lspClient LSPReferencesClient,
) ([]Callsite, error) {
	// Resolve anchor to get Tag (re-uses existing contract, no signature change)
	tag, err := r.ResolveAnchor(anchorID)
	if err != nil {
		return nil, err
	}

	excludeTests := !includeTests

	// LSP path
	if lspClient != nil && lspClient.IsAvailable() {
		pos := lsp.Position{
			Line:      tag.Line - 1, // Tag.Line is 1-based; LSP is 0-based
			Character: 0,
		}
		locations, lspErr := lspClient.FindReferences(ctx, tag.File, pos)
		if lspErr == nil {
			callsites := make([]Callsite, 0, len(locations))
			for _, loc := range locations {
				filePath := uriToPath(loc.URI)
				if excludeTests && isTestFile(filePath) {
					continue
				}
				callsites = append(callsites, Callsite{
					File:   filePath,
					Line:   loc.Range.Start.Line + 1, // 0-based → 1-based
					Column: loc.Range.Start.Character + 1,
					Method: "lsp",
				})
			}
			return callsites, nil
		}
		// LSP error → fall through to textual
	}

	// Textual fallback: walk projectRoot and grep for anchorID
	return r.resolveCallsitesTextual(tag, projectRoot, excludeTests)
}

// resolveCallsitesTextual walks projectRoot and returns Callsite entries for
// every line that contains tag.AnchorID, excluding the definition file itself.
func (r *Resolver) resolveCallsitesTextual(tag Tag, projectRoot string, excludeTests bool) ([]Callsite, error) {
	if projectRoot == "" {
		return nil, nil
	}

	var callsites []Callsite

	walkErr := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // ignore and continue
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "vendor" || base == "node_modules" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip the definition file itself
		if path == tag.File {
			return nil
		}
		if excludeTests && isTestFile(path) {
			return nil
		}

		hits := callerLinesInFile(path, tag.AnchorID)
		for _, lineNum := range hits {
			callsites = append(callsites, Callsite{
				File:   path,
				Line:   lineNum,
				Column: 0,
				Method: "textual",
			})
		}
		return nil
	})

	if walkErr != nil {
		return callsites, nil
	}
	return callsites, nil
}

// callerLinesInFile returns the 1-based line numbers within filePath that contain symbol.
func callerLinesInFile(filePath, symbol string) []int {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var lines []int
	lineNum := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineNum++
		if strings.Contains(scanner.Text(), symbol) {
			lines = append(lines, lineNum)
		}
	}
	return lines
}

// AuditLowFanIn returns ANCHOR tags with low fan-in (< 3 callers).
// This is a placeholder implementation - SPC-004 will implement actual fan-in counting.
func (r *Resolver) AuditLowFanIn() []Tag {
	anchors := r.ListAnchors()

	// Placeholder: Return all anchors as "low fan-in" until SPC-004 implementation
	// In SPC-004, this will use actual static analysis to count callers
	return anchors
}
