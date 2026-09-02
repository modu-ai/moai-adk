package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/graph"
	"github.com/modu-ai/moai-adk/internal/mx"
)

// edgesRefreshStats carries the measured cost of one edges refresh.
type edgesRefreshStats struct {
	duration time.Duration
}

// newEdgesRefreshClock constructs a wall-clock measurer for one refresh:
// call the constructor at the refresh's start, the returned func at its end.
func newEdgesRefreshClock() func() time.Duration {
	start := time.Now()
	return func() time.Duration { return time.Since(start) }
}

// edgesRefreshClock is the duration seam for the edges refresh (CR round-2
// 3855149237 / REQ-GFR-004): production constructs the wall-clock measurer
// via newEdgesRefreshClock; tests may inject a deterministic duration so
// budget-overrun assertions never depend on real refresh timing. The
// default-construction contract (seam == wall-clock constructor) is pinned
// by TestEdgesRefreshClockDefaultIsWallClock.
//
// @MX:NOTE: [AUTO] test-injection seam for budget-overrun determinism (SPEC-V3R6-GRAPH-FRESHNESS-002 REQ-GFR-004)
// @MX:REASON: mutating package var by design — the only test seam in the graph refresh path
var edgesRefreshClock = newEdgesRefreshClock

// edgesRefreshNeeded is the query path's refresh decision (REQ-GF-007, CR
// round-2 3855149254): the edges half evaluates the SELECTED --edges
// artifact's own provenance — EdgesSourcesMovedFor reads edgesFile's meta
// sidecar, never the default artifact's — while the mx-index probe stays
// tree-anchored: the index is a project-level source whichever artifact
// consumes it. mxIndexChangedFiles is the caller's drift red line
// (DefaultThresholds().MXIndexChangedFiles is the gate-calibrated value).
func edgesRefreshNeeded(projectRoot, edgesFile string, mxIndexChangedFiles int) bool {
	return graph.EdgesSourcesMovedFor(projectRoot, edgesFile) ||
		graph.MXIndexNeedsRefresh(projectRoot, mxIndexChangedFiles)
}

// refreshEdgesArtifact brings the derived edges layer back in sync with its
// sources: refresh the mx-index first (it is both a source and the mx-spec
// edge extractor's input — and its inventory IS the described-source state
// the code-derived layers consume), then rebuild (doc + code layers) and
// re-stamp. Mechanical only — no LLM, no network (REQ-GF-007).
//
// REQ-GR-008: the shrink guard evaluates BETWEEN the build and the write —
// pre-write, so a refusal performs ZERO writes and the prior artifact (and
// its meta sidecar) stays byte-identical by construction. A typed
// *graph.ShrinkRefusalError carries the report; the callers apply the
// fail-safe shape (REQ-GR-009): answer from the existing artifact (query
// path) or skip the refresh (deferred path). An unreadable prior artifact
// (absent or corrupt) skips the guard — there is nothing to protect, and the
// rebuild self-heals the corruption.
func refreshEdgesArtifact(projectRoot, edgesFile string) (edgesRefreshStats, error) {
	elapsed := edgesRefreshClock()

	if _, err := mx.RefreshIndex(filepath.Join(projectRoot, ".moai", "state"), projectRoot, nil); err != nil {
		return edgesRefreshStats{}, fmt.Errorf("refresh mx-index: %w", err)
	}
	edges, scanned, _, err := graph.BuildWithCodeLayers(projectRoot)
	if err != nil {
		return edgesRefreshStats{}, fmt.Errorf("rebuild edges: %w", err)
	}
	if existing, lErr := graph.LoadJSONL(edgesFile); lErr == nil {
		scannedSet := make(map[string]bool, len(scanned))
		for _, f := range scanned {
			scannedSet[f] = true
		}
		if report := graph.DetectUnexplainedShrink(existing, edges, scannedSet, projectRoot); !report.Empty() {
			return edgesRefreshStats{}, &graph.ShrinkRefusalError{Report: report}
		}
	}
	if err := graph.WriteJSONL(edgesFile, edges); err != nil {
		return edgesRefreshStats{}, fmt.Errorf("write edges: %w", err)
	}
	if err := graph.WriteEdgesMeta(filepath.Join(filepath.Dir(edgesFile), graph.MetaFileName),
		projectRoot, graph.SourceFingerprintsForEdges(projectRoot), len(edges)); err != nil {
		return edgesRefreshStats{}, fmt.Errorf("write edges meta: %w", err)
	}
	return edgesRefreshStats{duration: elapsed()}, nil
}

// deferredEdgesRefresh is the cli-injected DeferredEdgesRefresh seam target
// (SPEC-GRAPH-REPORT-001 REQ-GR-010): a thin wrapper around
// refreshEdgesArtifact — the SINGLE rebuild path, wrapped never forked — that
// refreshes the DEFAULT edges artifact of projectDir and applies the same
// warning-only budget-overrun signal as the query-time refresh (REQ-GR-012).
// Fail-open by contract: the hook layer logs the returned error and never
// blocks session start (REQ-GR-011); the prior artifact stays intact.
func deferredEdgesRefresh(projectDir string) error {
	edgesFile := filepath.Join(projectDir, ".moai", "project", "graph", "edges.jsonl")
	stats, err := refreshEdgesArtifact(projectDir, edgesFile)
	if err != nil {
		return err
	}
	if over := graphRefreshOverrun(projectDir, stats.duration); over > 0 {
		_, _ = fmt.Fprintf(os.Stderr,
			"deferred graph refresh cost %s exceeded the %dms update budget by %.0fms (warning only)\n",
			stats.duration.Round(time.Millisecond), graphRefreshBudgetMS(projectDir), over.Seconds()*1000)
	}
	return nil
}

// graphRefreshBudget resolves the query-time update-cost budget (milliseconds)
// from gate.yaml graph_freshness.update_budget_ms, falling back to the
// config-package default when unset. The budget warns only — an overrun never
// blocks the answer (REQ-GF-009: a stale-but-labeled answer beats no answer).
func graphRefreshBudgetMS(projectRoot string) int {
	loader := config.NewLoader()
	cfg, err := loader.Load(filepath.Join(projectRoot, ".moai"))
	if err != nil {
		return config.DefaultGraphFreshnessUpdateBudgetMS
	}
	if cfg.Gate.GraphFreshness.UpdateBudgetMS > 0 {
		return cfg.Gate.GraphFreshness.UpdateBudgetMS
	}
	return config.DefaultGraphFreshnessUpdateBudgetMS
}

// graphRefreshOverrun returns how far a measured duration exceeded the budget
// (0 when within budget).
func graphRefreshOverrun(projectRoot string, measured time.Duration) time.Duration {
	budget := time.Duration(graphRefreshBudgetMS(projectRoot)) * time.Millisecond
	if measured <= budget {
		return 0
	}
	return measured - budget
}
