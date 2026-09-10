package graph

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/modu-ai/moai-adk/internal/graph/symbol"
	"github.com/modu-ai/moai-adk/internal/navigator/astx"
)

// Code-derived edge kinds (REQ-GF-014): additive to the five doc-derived
// kinds — the build never replaces, drops, or rewrites a doc edge.
const (
	// KindCodeCall is a caller-symbol → callee-name edge from AST extraction.
	KindCodeCall = "code-call"
	// KindCodeImport is a file → module edge from import extraction.
	KindCodeImport = "code-import"
)

// GradeMatrix re-exports the seam's matrix for consumers of this package.
func GradeMatrix() map[string]string { return symbol.GradeMatrix() }

// ValidateGradeMatrix re-exports the seam's defect check.
func ValidateGradeMatrix(matrix map[string]string) []string {
	return symbol.ValidateGradeMatrix(matrix)
}

// CodeEdges maps the symbol seam's extraction into persisted graph edges.
// The seam (internal/graph/symbol) holds the astx consumption and stays free
// of navigator-tier dependencies; this mapper owns the persisted shapes.
func CodeEdges(projectRoot string) ([]Edge, map[string]string, error) {
	calls, imports, decls, _, matrix, err := symbol.Extract(projectRoot)
	if err != nil {
		return nil, matrix, fmt.Errorf("graph: extract code edges: %w", err)
	}
	return mapSymbolEdges(calls, imports, decls), matrix, nil
}

// declIndex is the deterministic lookup structure behind the confidence
// tier pass: declared names per file, declaring directories per name, and
// each file's Go-module (Local) import set — all derived from the single
// Extract walk, all consumed by set lookup only (no map-iteration order
// reaches an emitted value).
type declIndex struct {
	namesByFile  map[string]map[string]bool // file → declared names
	dirsByName   map[string]map[string]bool // name → declaring directories
	localImports map[string]map[string]bool // file → module-prefixed import dirs
}

// buildDeclIndex assembles the tier-join index from the seam's walk output.
func buildDeclIndex(imports []symbol.ImportEdge, decls []symbol.FileDecls) *declIndex {
	idx := &declIndex{
		namesByFile:  map[string]map[string]bool{},
		dirsByName:   map[string]map[string]bool{},
		localImports: map[string]map[string]bool{},
	}
	for _, imp := range imports {
		if !imp.Local {
			continue
		}
		set := idx.localImports[imp.File]
		if set == nil {
			set = map[string]bool{}
			idx.localImports[imp.File] = set
		}
		set[imp.Module] = true
	}
	for _, d := range decls {
		set := idx.namesByFile[d.File]
		if set == nil {
			set = map[string]bool{}
			idx.namesByFile[d.File] = set
		}
		dir := filepath.ToSlash(filepath.Dir(d.File))
		for _, name := range d.Names {
			set[name] = true
			dirs := idx.dirsByName[name]
			if dirs == nil {
				dirs = map[string]bool{}
				idx.dirsByName[name] = dirs
			}
			dirs[dir] = true
		}
	}
	return idx
}

// resolveConfidence applies the promotion tiers to one code-call edge, first
// match wins (plan.md §C): T1 same-file declaration → extracted; T2 the
// callee declared in a Go-module-imported directory → extracted (Go-module
// imports ONLY — the seam's Local flag marks exactly the strippable
// specifiers); T3 declared in the caller's own directory → intra-package;
// T4 otherwise → inferred. Evidence semantics, not scope-aware proof — that
// remains the grade axis's unassigned full tier (REQ-GEC-007).
func (idx *declIndex) resolveConfidence(file, callee string) string {
	if idx.namesByFile[file][callee] {
		return ResolutionExtracted // T1 same-file
	}
	callerDir := filepath.ToSlash(filepath.Dir(file))
	for dir := range idx.dirsByName[callee] { // T2 import evidence (set membership only)
		if idx.localImports[file][dir] {
			return ResolutionExtracted
		}
	}
	if idx.dirsByName[callee][callerDir] {
		return ResolutionIntraPackage // T3 same-directory declaration — package proximity (a foo_test external test package shares the dir)
	}
	return ResolutionInferred // T4 name-only fallback
}

// mapSymbolEdges converts seam-typed extraction into persisted edges,
// stamping resolution confidence on code-call edges only (REQ-GEC-001,
// REQ-GEC-006: doc-derived and code-import edges stay field-free).
func mapSymbolEdges(calls []symbol.CallEdge, imports []symbol.ImportEdge, decls []symbol.FileDecls) []Edge {
	idx := buildDeclIndex(imports, decls)
	var edges []Edge
	for _, c := range calls {
		source := c.File
		if c.Caller != "" {
			source = c.File + ":" + c.Caller
		}
		resolution := idx.resolveConfidence(c.File, c.Callee)
		edges = append(edges, Edge{
			Kind:       KindCodeCall,
			Source:     source,
			Target:     c.Callee,
			Line:       c.Line,
			Grade:      c.Grade,
			Resolution: resolution,
			Confidence: ConfidenceFor(resolution),
		})
	}
	for _, imp := range imports {
		edges = append(edges, Edge{
			Kind:   KindCodeImport,
			Source: imp.File,
			Target: imp.Module,
			Line:   imp.Line,
			Grade:  imp.Grade,
		})
	}
	sort.Slice(edges, func(i, j int) bool { return EdgeLess(edges[i], edges[j]) })
	return edges
}

// BuildWithCodeLayers runs the doc-derived Build and ADDS the code-derived
// layers (REQ-GF-014). Additivity invariant: every doc edge is present with
// its relationship fields unchanged. Doc/code disagreement on the same
// relationship is EXPOSED via the disagrees_with marker — never a silent
// pick (REQ-GF-015).
//
// The second return is the scanned-file list — the files the extraction walk
// actually processed — which the shrink guard consumes as its scanned set
// (REQ-GR-008): captured at build time inside the walk, never a second scan,
// and never a fingerprint aggregate (those hash doc-side source SETS and do
// not cover the source tree).
func BuildWithCodeLayers(projectRoot string) ([]Edge, []string, map[string]string, error) {
	return BuildWithCodeLayersMode(projectRoot, DisagreementRefuteOnly)
}

// DisagreementMode selects which disagreement directions the build marks.
type DisagreementMode int

const (
	// DisagreementRefuteOnly marks doc-explicit claims the code layer refutes
	// (the default). Code-found/doc-silent is the curated summary's normal
	// state — suppressed, but RETRIEVABLE via DisagreementAll (a decided-not-
	// to-report signal must never harden into cannot-be-reported).
	DisagreementRefuteOnly DisagreementMode = iota
	// DisagreementAll ALSO marks the suppressed direction: local code-import
	// dependencies the doc layer does not record.
	DisagreementAll
)

// BuildWithCodeLayersMode is BuildWithCodeLayers with the disagreement mode
// selected (revival path for the suppressed code-found/doc-silent direction;
// wired to `moai graph build --all-disagreements`).
//
// The code extraction runs FIRST so the doc layer's tag edges can join
// body-anchored tags to their enclosing symbols via the retained ranges
// (REQ-MTE-002 — retention, not a second parse). An extraction failure fails
// open on the code layers AND on the range join: doc edges survive with the
// self-edge tag form (REQ-MTE-015's no-range-data case).
func BuildWithCodeLayersMode(projectRoot string, mode DisagreementMode) ([]Edge, []string, map[string]string, error) {
	calls, imports, decls, scanned, matrix, extractErr := symbol.Extract(projectRoot)
	if extractErr != nil {
		// Fail-open on the code layers ONLY: doc edges must survive — with an
		// EMPTY scanned set: no file was processed, so the shrink guard (the
		// caller's write path) sees every removed code edge as unexplained.
		docEdges, err := Build(projectRoot)
		if err != nil {
			return nil, nil, GradeMatrix(), fmt.Errorf("graph: build doc edges: %w", err)
		}
		return docEdges, nil, GradeMatrix(), nil
	}
	docEdges, err := buildDocLayers(projectRoot, rangesByFileFromDecls(decls))
	if err != nil {
		return nil, nil, matrix, fmt.Errorf("graph: build doc edges: %w", err)
	}
	codeEdges := mapSymbolEdges(calls, imports, decls)
	markImportDisagreements(docEdges, codeEdges, imports, mode)

	all := make([]Edge, 0, len(docEdges)+len(codeEdges))
	all = append(all, docEdges...)
	all = append(all, codeEdges...)
	sort.Slice(all, func(i, j int) bool { return EdgeLess(all[i], all[j]) })
	return all, scanned, matrix, nil
}

// rangesByFileFromDecls indexes the seam's retained declaration ranges by
// repo-relative file — the tag-edge join's lookup structure.
func rangesByFileFromDecls(decls []symbol.FileDecls) map[string][]astx.FuncRange {
	out := map[string][]astx.FuncRange{}
	for _, d := range decls {
		out[d.File] = d.Ranges
	}
	return out
}

// markImportDisagreements annotates doc import edges the code layer REFUTES.
//
// Disagreement is asymmetric by domain: the doc graph is a curated summary,
// so a dependency the code layer finds and the doc layer simply does not
// mention is the summary's normal state, not a contradiction — no marker in
// the default mode. A marker fires when the DOC layer explicitly claims a
// dependency and the code layer scanned that source package without finding
// it: two claims about the same relationship, in opposite directions. The
// DisagreementAll mode revives the suppressed direction on the code edges.
func markImportDisagreements(docEdges, codeEdges []Edge, imports []symbol.ImportEdge, mode DisagreementMode) {
	// scannedPkgs: packages whose files the code layer actually scanned.
	scannedPkgs := map[string]bool{}
	// codeImplied: package → local-package dependencies the code layer found.
	codeImplied := map[string]bool{}
	localDomain := map[string]bool{} // every local package name seen anywhere
	for _, imp := range imports {
		pkg := filepath.ToSlash(filepath.Dir(imp.File))
		scannedPkgs[pkg] = true
		localDomain[pkg] = true
		if imp.Local {
			codeImplied[pkg+"\x00"+imp.Module] = true
			localDomain[imp.Module] = true
		}
	}
	docImports := map[string]bool{}
	for _, e := range docEdges {
		if e.Kind == KindImport {
			docImports[e.Source+"\x00"+e.Target] = true
		}
	}

	for i := range docEdges {
		e := &docEdges[i]
		if e.Kind != KindImport {
			continue
		}
		// Comparable only when the code layer scanned the source package AND
		// the target is a local package in the code layer's domain; anything
		// else is outside the code layer's observability — silence there is
		// not a refutation.
		if !scannedPkgs[e.Source] || !localDomain[e.Target] {
			continue
		}
		if !codeImplied[e.Source+"\x00"+e.Target] {
			e.DisagreesWith = KindCodeImport + " (code layer scanned " + e.Source + " and found no import of " + e.Target + ")"
		}
	}

	if mode == DisagreementAll {
		for i := range codeEdges {
			e := &codeEdges[i]
			if e.Kind != KindCodeImport {
				continue
			}
			if !symbolImportLocal(imports, e.Source, e.Target) {
				continue
			}
			pkg := filepath.ToSlash(filepath.Dir(e.Source))
			if !docImports[pkg+"\x00"+e.Target] {
				e.DisagreesWith = KindImport + " [revived] (doc layer does not record this dependency)"
			}
		}
	}
}

// symbolImportLocal reports whether a local import edge (file→module) exists
// in the seam's extraction — the marker revival path's locality gate.
func symbolImportLocal(imports []symbol.ImportEdge, file, module string) bool {
	for _, imp := range imports {
		if imp.File == file && imp.Module == module && imp.Local {
			return true
		}
	}
	return false
}
