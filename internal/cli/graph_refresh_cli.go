package cli

import (
	"fmt"
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

// refreshEdgesArtifact brings the derived edges layer back in sync with its
// sources: refresh the mx-index first (it is both a source and the mx-spec
// edge extractor's input — and its inventory IS the described-source state
// the code-derived layers consume), then rebuild (doc + code layers) and
// re-stamp. Mechanical only — no LLM, no network (REQ-GF-007).
func refreshEdgesArtifact(projectRoot, edgesFile string) (edgesRefreshStats, error) {
	start := time.Now()

	if _, err := mx.RefreshIndex(filepath.Join(projectRoot, ".moai", "state"), projectRoot, nil); err != nil {
		return edgesRefreshStats{}, fmt.Errorf("refresh mx-index: %w", err)
	}
	edges, _, err := graph.BuildWithCodeLayers(projectRoot)
	if err != nil {
		return edgesRefreshStats{}, fmt.Errorf("rebuild edges: %w", err)
	}
	if err := graph.WriteJSONL(edgesFile, edges); err != nil {
		return edgesRefreshStats{}, fmt.Errorf("write edges: %w", err)
	}
	if err := graph.WriteEdgesMeta(filepath.Join(filepath.Dir(edgesFile), graph.MetaFileName),
		projectRoot, graph.SourceFingerprintsForEdges(projectRoot)); err != nil {
		return edgesRefreshStats{}, fmt.Errorf("write edges meta: %w", err)
	}
	return edgesRefreshStats{duration: time.Since(start)}, nil
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
