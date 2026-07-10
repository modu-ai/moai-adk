package web

// SPEC-WEB-CONSOLE-013 M3 — READ-ONLY Model Policy view (GET /model-policy).
//
// The Model Policy view is a purely observational dashboard mirroring the
// board.go precedent (REQ-WC13-020, plan §A.4-2). It renders (a) the active
// performance tier from llm.performance_tier and (b) the
// workflow.model_routing_profiles map as a 3 perfTier × 12 cell
// (S/M/L × plan/run/sync/mx) table — BOTH as READ-ONLY displays. It is
// FieldDef-less: it lives OUTSIDE the schema/persist pipeline and has NO write
// path, NO PersistTarget, NO form control, and NO status transition
// (REQ-WC13-021). The legacy flat workflow.model_routing block and
// workflow.workflow_agents are deliberately NOT rendered (REQ-WC13-022/023).

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// modelPolicyPerfTiers is the No-Haiku 3-tier order for the routing table rows
// (max/medium/low — the closed set validated by config.RouteModelFor). It is the
// routing-profile selector axis, DISTINCT from the Launch section's model_policy
// (init-time agent frontmatter policy: high/medium/low) — REQ-WC13-025.
func modelPolicyPerfTiers() []string { return []string{"max", "medium", "low"} }

// modelPolicySpecTiers / modelPolicyPhases enumerate the 12-cell axis
// (S/M/L × plan/run/sync/mx) per performance tier.
func modelPolicySpecTiers() []string { return []string{"S", "M", "L"} }
func modelPolicyPhases() []string    { return []string{"plan", "run", "sync", "mx"} }

// mpCell is one (specTier, phase) routing cell within a performance-tier profile.
type mpCell struct {
	Tier     string // S / M / L
	Phase    string // plan / run / sync / mx
	Model    string
	Effort   string
	Fallback bool // true → the value is the documented default (inherit/medium)
}

// mpProfile groups the 12 cells for one performance tier.
type mpProfile struct {
	PerfTier string
	Cells    []mpCell
}

// modelPolicyView is the typed view-model for the READ-ONLY Model Policy page.
type modelPolicyView struct {
	BindAddr string

	// PerfTier is the raw llm.performance_tier value; PerfTierIsEmpty drives the
	// "(runtime default: medium)" empty-value label (REQ-WC13-024).
	PerfTier        string
	PerfTierIsEmpty bool

	// Profiles carries the 3 performance-tier profiles (each 12 cells).
	Profiles []mpProfile

	// BlockAbsent is true when workflow.model_routing_profiles is absent — every
	// lookup falls back to inherit/medium (defaultRoutingEntry), REQ-WC13-026.
	BlockAbsent bool

	Banner     string
	BannerKind string
}

// handleModelPolicy serves GET /model-policy — the READ-ONLY Model Policy view.
// Non-GET methods are rejected with 405 (there is no write path — REQ-WC13-021).
func (a *app) handleModelPolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	view, err := a.buildModelPolicyView()
	if err != nil {
		a.renderError(w, http.StatusInternalServerError, "could not read model policy: "+err.Error())
		return
	}
	a.renderModelPolicy(w, http.StatusOK, view)
}

// buildModelPolicyView loads the project config (LoadRaw — read-intent, no
// validation) and assembles the read-only view-model. An absent config dir
// yields empty values (LoadRaw default), which surface as the empty-tier label +
// the absent-block fallback state rather than an error.
func (a *app) buildModelPolicyView() (modelPolicyView, error) {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(a.cfg.ProjectRoot)
	if err != nil {
		return modelPolicyView{}, err
	}

	view := modelPolicyView{
		BindAddr:        a.resolveBindAddr(),
		PerfTier:        cfg.LLM.PerformanceTier,
		PerfTierIsEmpty: strings.TrimSpace(cfg.LLM.PerformanceTier) == "",
		BlockAbsent:     len(cfg.Workflow.ModelRoutingProfiles) == 0,
	}

	for _, pt := range modelPolicyPerfTiers() {
		prof := mpProfile{PerfTier: pt}
		for _, tier := range modelPolicySpecTiers() {
			for _, phase := range modelPolicyPhases() {
				entry, rerr := cfg.RouteModelFor(tier, phase, pt)
				if rerr != nil {
					// tier/phase/perfTier are drawn from the closed sets above, so an
					// error is not expected; degrade to the documented fallback rather
					// than failing the whole page (view is render-only — REQ-WC13-026).
					entry = config.ModelRoutingEntry{Model: "inherit", Effort: "medium", FallbackApplied: true}
				}
				prof.Cells = append(prof.Cells, mpCell{
					Tier:     tier,
					Phase:    phase,
					Model:    entry.Model,
					Effort:   entry.Effort,
					Fallback: entry.FallbackApplied,
				})
			}
		}
		view.Profiles = append(view.Profiles, prof)
	}
	return view, nil
}

// renderModelPolicy renders the Model Policy Templ component into a buffer first,
// so a render error surfaces as a readable inline error (REQ-WC-010 discipline)
// rather than a half-written 200.
func (a *app) renderModelPolicy(w http.ResponseWriter, status int, view modelPolicyView) {
	var buf bytes.Buffer
	if err := modelPolicyPage(view).Render(context.Background(), &buf); err != nil {
		a.renderError(w, http.StatusInternalServerError, "internal error: model policy render failed: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}
