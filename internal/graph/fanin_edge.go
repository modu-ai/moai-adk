package graph

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// EdgeFanInSource is the graph-backed fan-in evidence source for the MX
// validator's P1 rule (REQ-MTE-009). It answers from the persisted
// edges.jsonl artifact — DISTINCT evidence-backed caller FILES over
// confidence-bearing code-call edges, excluding the declaring file
// (plan B.7) and test-pattern caller files (REQ-MTE-012, the
// REQ-SPC-004-040 fallback pattern set via mx.IsTestFileWithPatterns; a
// user's mx.yaml test_paths globs are deliberately NOT honored — accepted
// divergence, spec.md §E).
//
// The type is STRUCTURALLY compatible with the consumer-side interface
// (internal/hook/mx FanInEvidenceSource): the method signature uses only
// primitive types so this data-layer package never imports the hook layer.
// The wiring happens at the one construction site permitted to import
// internal/graph: internal/hook/session_end.go (REQ-MTE-010).
//
// @MX:NOTE: [AUTO] EdgeFanInSource — structural-typing seam: the consumer interface lives in hook/mx, the implementation lives with the data (SPEC-MX-TAG-EDGES-001 plan B.6)
type EdgeFanInSource struct {
	projectRoot string
	edgesFile   string
}

// NewEdgeFanInSource builds a source over projectRoot's default edges
// artifact (<root>/.moai/project/graph/edges.jsonl).
func NewEdgeFanInSource(projectRoot string) *EdgeFanInSource {
	return &EdgeFanInSource{
		projectRoot: projectRoot,
		edgesFile:   filepath.Join(projectRoot, ".moai", "project", "graph", "edges.jsonl"),
	}
}

// EvidenceBacked answers the fan-in of funcName as seen from currentFile
// (the file that declares it — absolute paths accepted). Returns the
// blocking (evidence-backed) count, the inferred-only count, and a label
// naming the artifact state the answer came from. An error means the
// artifact cannot serve an evidence-backed answer (absent, unreadable, or
// carries zero code-call edges — the CGO-off degrade, REQ-MTE-015); the
// caller falls back to the textual source and labels the verdict
// (REQ-MTE-011). A stale-but-present artifact still answers — with a stale
// label and a stderr note, never silently (REQ-MTE-011).
func (s *EdgeFanInSource) EvidenceBacked(ctx context.Context, funcName, currentFile string) (evidence, inferredOnly int, label string, err error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, "", fmt.Errorf("graph: fan-in source cancelled: %w", err)
	}
	edges, err := LoadJSONL(s.edgesFile)
	if err != nil {
		return 0, 0, "", fmt.Errorf("graph: load edges artifact: %w", err)
	}
	if !hasCodeCallEdges(edges) {
		// Zero code-call edges means the artifact carries no extraction
		// evidence at all (fresh artifact under CGO-off, or an artifact
		// built before the code layers existed). Answering 0 would read as
		// "no callers" when it actually means "no evidence" — degrade to
		// the caller's textual fallback instead (REQ-MTE-015).
		return 0, 0, "", errors.New("graph: artifact carries no code-call evidence")
	}

	label = "edges"
	if EdgesSourcesMovedFor(s.projectRoot, s.edgesFile) {
		label = "edges(stale)"
		slog.Warn("graph: fan-in answered from a stale edges artifact — run 'moai graph build'",
			"symbol", funcName,
			"artifact", s.edgesFile,
		)
	}

	res := SymbolFanIn(edges, funcName, repoRel(s.projectRoot, currentFile))
	// Hub exclusion (REQ-MTE-012) applies at the aggregation layer so it
	// does not depend on extractor scope: test-pattern caller files are
	// removed from BOTH lists before counting.
	evidenceFiles := res.EvidenceFiles[:0:0]
	for _, f := range res.EvidenceFiles {
		if !isTestCallerFile(f) {
			evidenceFiles = append(evidenceFiles, f)
		}
	}
	inferredFiles := res.InferredFiles[:0:0]
	for _, f := range res.InferredFiles {
		if !isTestCallerFile(f) {
			inferredFiles = append(inferredFiles, f)
		}
	}
	return len(evidenceFiles), len(inferredFiles), label, nil
}

// hasCodeCallEdges reports whether the artifact carries any code-call edge.
func hasCodeCallEdges(edges []Edge) bool {
	for _, e := range edges {
		if e.Kind == KindCodeCall {
			return true
		}
	}
	return false
}

// isTestCallerFile applies the REQ-SPC-004-040 hard-coded fallback pattern
// set (*_test.go suffix, tests/, fixtures/, testdata/ path components) via
// the single shared predicate in internal/mx — ONE definition governs both
// the query-side textual counter and this source (plan B.8).
func isTestCallerFile(repoRelPath string) bool {
	return mx.IsTestFileWithPatterns(repoRelPath, nil)
}
