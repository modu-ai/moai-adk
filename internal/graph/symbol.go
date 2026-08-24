package graph

import (
	"path/filepath"
	"sort"

	"github.com/modu-ai/moai-adk/internal/graph/symbol"
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
	calls, imports, matrix, err := symbol.Extract(projectRoot)
	if err != nil {
		return nil, matrix, err
	}
	return mapSymbolEdges(calls, imports), matrix, nil
}

// mapSymbolEdges converts seam-typed extraction into persisted edges.
func mapSymbolEdges(calls []symbol.CallEdge, imports []symbol.ImportEdge) []Edge {
	var edges []Edge
	for _, c := range calls {
		source := c.File
		if c.Caller != "" {
			source = c.File + ":" + c.Caller
		}
		edges = append(edges, Edge{
			Kind:   KindCodeCall,
			Source: source,
			Target: c.Callee,
			Line:   c.Line,
			Grade:  c.Grade,
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
func BuildWithCodeLayers(projectRoot string) ([]Edge, map[string]string, error) {
	docEdges, err := Build(projectRoot)
	if err != nil {
		return nil, nil, err
	}
	calls, imports, matrix, err := symbol.Extract(projectRoot)
	if err != nil {
		// Fail-open on the code layers ONLY: doc edges must survive.
		return docEdges, GradeMatrix(), nil
	}
	codeEdges := mapSymbolEdges(calls, imports)
	markImportDisagreements(docEdges, imports)

	all := make([]Edge, 0, len(docEdges)+len(codeEdges))
	all = append(all, docEdges...)
	all = append(all, codeEdges...)
	sort.Slice(all, func(i, j int) bool { return EdgeLess(all[i], all[j]) })
	return all, matrix, nil
}

// markImportDisagreements annotates doc import edges the code layer REFUTES.
//
// Disagreement is asymmetric by domain: the doc graph is a curated summary,
// so a dependency the code layer finds and the doc layer simply does not
// mention is the summary's normal state, not a contradiction — no marker.
// A marker fires only when the DOC layer explicitly claims a dependency and
// the code layer scanned that source package without finding it: two claims
// about the same relationship, in opposite directions.
func markImportDisagreements(docEdges []Edge, imports []symbol.ImportEdge) {
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
}
