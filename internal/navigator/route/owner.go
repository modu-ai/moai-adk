package route

import (
	"path/filepath"
	"strings"

	navsync "github.com/modu-ai/moai-adk/internal/navigator/sync"
)

// Action directives (REQ-NS4-003c: non-generic, per source_kind). These are
// M2's own prose — the Route layer's value-add — NOT imported from an
// existing asset (plan.md §C.7 asset-reuse map row 7).
const (
	actionOrphan  = "link this SPEC to a design feature or document its design rationale"
	actionMissing = "create a SPEC for this design feature or link existing code"
	actionDetect  = "verify the affected doc rows still hold after this edit"
)

// resolveOrphanOwner resolves the owner for an audit-orphan entry
// (REQ-NS4-004, path: audit-orphan → implementation_path).
//
//   - implementation_path non-empty → that path, normalized to absolute
//     (high confidence — owner resolved from a direct path field).
//   - implementation_path empty → SPEC directory path (low confidence —
//     B4 fallback: no code path mapped; the SPEC itself is the owner).
func resolveOrphanOwner(entry OrphanEntry, projectRoot string) (string, Confidence) {
	if entry.ImplementationPath != "" {
		return absPath(entry.ImplementationPath, projectRoot), ConfidenceHigh
	}
	// B4 fallback: no code path mapped → the SPEC directory is the owner.
	specDir := filepath.Join(projectRoot, ".moai", "specs", entry.SpecID)
	return specDir + string(filepath.Separator), ConfidenceLow
}

// resolveMissingOwner resolves the owner for an audit-missing entry
// (REQ-NS4-004, path: audit-missing → @NAV:SYM symbol via graph, else
// design-doc source.file).
//
//   - The design doc at source.file references an @NAV:SYM token AND the
//     symbol has a code declaration in the M0 graph → the code path
//     (medium confidence — owner resolved via one graph hop).
//   - No symbol found, OR the symbol has no code declaration → design-doc
//     source.file (low confidence — the doc author's responsibility).
//
// B5 heading-range heuristic decision: the heading_anchor → line-range
// mapping is NOT applied. The audit's source.heading_path is a string
// anchor (e.g. "## Authentication > ### OAuth2"), not a line number, and
// re-parsing the design doc to map anchors → ranges adds I/O + fragility
// (the doc may have been edited since the graph was built). Instead, a
// whole-doc symbol lookup is used: any @NAV:SYM token in the design doc
// qualifies. This avoids the unreliable heading→line-range mapping while
// still resolving symbols via the graph — more precise than the full
// doc-fallback (which skips symbol resolution entirely), and the ≥70%
// accuracy floor survives under both paths per the dual-arithmetic in
// plan.md §E / acceptance.md AC-NS4-010.
func resolveMissingOwner(entry MissingEntry, graph *navsync.Graph, projectRoot string) (string, Confidence) {
	docPath := absPath(entry.Source.File, projectRoot)

	if graph != nil {
		if codePath := resolveSymbolDeclaration(graph, docPath); codePath != "" {
			return codePath, ConfidenceMedium
		}
	}

	// Fallback: design-doc source.file (low confidence — no code binding).
	return docPath, ConfidenceLow
}

// resolveDetectOwner resolves the owner for a detect record
// (REQ-NS4-004, path: detect → changed_path).
//
// The changed_path is the code file the engineer just touched; they own the
// follow-up: verify the affected doc rows still hold. High confidence —
// owner resolved from a direct path field, no graph traversal.
func resolveDetectOwner(record DetectRecord, projectRoot string) (string, Confidence) {
	return absPath(record.ChangedPath, projectRoot), ConfidenceHigh
}

// resolveSymbolDeclaration finds the code declaration path for a symbol
// referenced via @NAV:SYM in the design doc at docPath. Returns "" if no
// symbol is found or no code declaration exists for the referenced symbol.
//
// Mechanism (B5 whole-doc lookup — see resolveMissingOwner docstring):
//  1. Scan the M0 graph for sym-edges whose SourcePath matches the design
//     doc. These tell us which symbols the doc references.
//  2. For each referenced symbol, find a sym-edge pointing to it whose
//     SourcePath is a code file (not a .md design doc). That code file is
//     the symbol's declaration location.
//  3. Return the first code declaration path found (deterministic: edges
//     are iterated in graph order, which is sorted by M0's join step).
func resolveSymbolDeclaration(graph *navsync.Graph, docPath string) string {
	docClean := filepath.Clean(docPath)

	// Step 1: collect symbols referenced from the design doc.
	symbolsFromDoc := make(map[string]bool)
	for _, edge := range graph.Edges {
		if edge.EdgeType != navsync.EdgeSym {
			continue
		}
		if filepath.Clean(edge.SourcePath) == docClean {
			symbolsFromDoc[edge.TargetNode] = true
		}
	}
	if len(symbolsFromDoc) == 0 {
		return ""
	}

	// Step 2: find a code declaration for any referenced symbol.
	for _, edge := range graph.Edges {
		if edge.EdgeType != navsync.EdgeSym {
			continue
		}
		if !symbolsFromDoc[edge.TargetNode] {
			continue
		}
		if isCodePath(edge.SourcePath) {
			return filepath.Clean(edge.SourcePath)
		}
	}

	return ""
}

// isCodePath returns true if the path is a code file (not a design doc).
// Design docs are .md files under .moai/project/ or .moai/docs/; all other
// files are treated as code for the purposes of owner resolution.
func isCodePath(path string) bool {
	return !strings.HasSuffix(path, ".md")
}

// absPath normalizes a path to cleaned absolute form using projectRoot.
// If the path is already absolute, it is cleaned in place. If relative, it
// is joined with projectRoot and cleaned. This is the single path-resolution
// helper for the Route layer — all owner paths pass through it so the
// output is deterministic.
func absPath(path, projectRoot string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(projectRoot, path))
}
