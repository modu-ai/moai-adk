package web

// SPEC-WEB-CONSOLE-013 M3 — Model Policy view (GET /model-policy).
//
// The Model Policy view is an observational dashboard mirroring the board.go
// precedent (REQ-WC13-020, plan §A.4-2). It renders the active performance tier
// from llm.performance_tier and the workflow.model_routing_profiles map as a
// 3 perfTier × 12 cell (S/M/L × plan/run/sync/mx) table — both READ-ONLY. The
// routing display lives OUTSIDE the schema/persist pipeline (no FieldDef, no
// PersistTarget). The legacy flat workflow.model_routing block and
// workflow.workflow_agents are deliberately NOT rendered (REQ-WC13-022/023).
//
// SPEC-MODEL-TIER-PLANTYPE-001 M4 adds ONE sanctioned write path atop this view:
// a plan_type selector (POST /model-policy/plan-type) that persists exactly
// llm.plan_type (§B.4 — a deliberately-narrow, user-approved exception to the
// SPEC-WEB-CONSOLE-013 REQ-WC13-021 read-only doctrine). NO OTHER field becomes
// writable; the page route /model-policy itself stays GET-only (405 on non-GET).

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
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

// planPreviewCell is one tier column ({model, effort}) for an agent within a plan
// preview. Values are DERIVED from template.GetTierProfileEntry at render time
// (REQ-MTP-023) — never hardcoded in the web layer.
type planPreviewCell struct {
	Tier   string // max / medium / low
	Model  string
	Effort string
}

// planPreviewRow is one agent's three tier columns (max/medium/low) within a plan.
type planPreviewRow struct {
	Agent string
	Cells []planPreviewCell
}

// planPreview is a single plan's {model, effort} preview table (10 agents × 3 tiers).
type planPreview struct {
	PlanType string // api / subscription
	IsActive bool   // true when this plan equals the effective active plan
	Rows     []planPreviewRow
}

// modelPolicyView is the typed view-model for the Model Policy page. As of
// SPEC-MODEL-TIER-PLANTYPE-001 M4 the page carries a single sanctioned write path
// (the plan_type selector, §B.4) atop the otherwise read-only routing display.
type modelPolicyView struct {
	BindAddr string

	// ActivePlanType is the effective plan type (llm.plan_type resolved to the
	// subscription default when absent/empty); PlanTypeIsEmpty drives the
	// "(default: subscription)" label (REQ-MTP-019).
	ActivePlanType  string
	PlanTypeIsEmpty bool

	// PlanPreviews carries the two plan preview tables (api + subscription), each
	// 10 agents × 3 tiers, derived from the Go tier-profile structure
	// (REQ-MTP-020/023 — no matrix literal duplicated here).
	PlanPreviews []planPreview

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

// modelPolicyPlanOptions returns the plan-type selector options in a stable order
// (api first, subscription second) for the /model-policy selector.
func modelPolicyPlanOptions() []string {
	return []string{config.PlanTypeAPI, config.PlanTypeSubscription}
}

// buildPlanPreviews derives the two plan preview tables from the single Go
// tier-profile structure (REQ-MTP-020/023). Both plans render every retained agent
// across the {max, medium, low} tiers; the {model, effort} cells come straight from
// template.GetTierProfileEntry so no matrix literal is duplicated in the web layer.
func buildPlanPreviews(active string) []planPreview {
	plans := modelPolicyPlanOptions()
	tiers := template.ValidPerformanceTiers() // {max, medium, low}
	agents := template.TierProfileAgents()
	previews := make([]planPreview, 0, len(plans))
	for _, plan := range plans {
		pv := planPreview{PlanType: plan, IsActive: plan == active}
		for _, agent := range agents {
			row := planPreviewRow{Agent: agent}
			for _, tier := range tiers {
				entry, _ := template.GetTierProfileEntry(plan, agent, tier)
				row.Cells = append(row.Cells, planPreviewCell{
					Tier:   tier,
					Model:  entry.Model,
					Effort: entry.Effort,
				})
			}
			pv.Rows = append(pv.Rows, row)
		}
		previews = append(previews, pv)
	}
	return previews
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

// handleModelPolicyPlanType serves POST /model-policy/plan-type — the single
// sanctioned write path of the Model Policy surface (SPEC-MODEL-TIER-PLANTYPE-001
// §B.4 / REQ-MTP-021). It validates the posted value against the closed set
// {api, subscription}; an out-of-set value is rejected 4xx WITHOUT mutating
// llm.yaml, and a valid value is persisted via template.ApplyPlanType (which
// rewrites ONLY the plan_type line). The loopback Host + Sec-Fetch-Site gate for
// mutating methods is applied upstream by hostCheckMiddleware. On success it
// redirects (303) back to the read-only page so the re-render reflects the new
// active plan (Post/Redirect/Get — a refresh does not re-submit).
func (a *app) handleModelPolicyPlanType(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request: could not parse form", http.StatusBadRequest)
		return
	}
	planType := strings.TrimSpace(r.PostFormValue("plan_type"))
	if !config.IsValidPlanType(planType) {
		// Out-of-set value: reject 4xx and leave llm.yaml byte-identical (REQ-MTP-021).
		http.Error(w, "invalid plan_type: must be one of {api, subscription}", http.StatusBadRequest)
		return
	}
	if err := template.ApplyPlanType(a.cfg.ProjectRoot, planType); err != nil {
		a.renderError(w, http.StatusInternalServerError, "could not persist plan_type: "+err.Error())
		return
	}
	// Re-apply the {model, effort} tier profile to the shipped agent definition
	// files so the plan_type switch takes effect on agents (mirrors `moai update`
	// at internal/cli/update.go:420). The tier is resolved from the persisted
	// llm.performance_tier; a non-initialized project (no manifest) is a no-op.
	a.applyTierProfileToAgents(w, r)
}

// handleModelPolicyTier serves POST /model-policy/tier — the tier (performance_tier)
// write path of the Model Policy surface, unified on the same page as plan_type.
// It validates the posted value against the canonical {max, medium, low} set
// (the vocabulary template.ApplyTierProfile + template.ValidPerformanceTiers
// expect — NOT the legacy wizard {high, medium, low}). An out-of-set value is
// rejected 4xx WITHOUT mutating llm.yaml or agent files. The loopback Host +
// Sec-Fetch-Site gate for mutating methods is applied upstream by
// hostCheckMiddleware. On success it persists performance_tier AND re-applies the
// tier profile to the shipped agent .md frontmatter, then redirects (303) back to
// the read-only page (Post/Redirect/Get).
func (a *app) handleModelPolicyTier(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request: could not parse form", http.StatusBadRequest)
		return
	}
	tier := strings.TrimSpace(r.PostFormValue("performance_tier"))
	if !template.IsValidPerformanceTier(tier) {
		// Out-of-set value: reject 4xx and leave llm.yaml byte-identical.
		http.Error(w, "invalid performance_tier: must be one of {max, medium, low}", http.StatusBadRequest)
		return
	}
	if err := template.ApplyPerformanceTier(a.cfg.ProjectRoot, tier); err != nil {
		a.renderError(w, http.StatusInternalServerError, "could not persist performance_tier: "+err.Error())
		return
	}
	// Re-apply the {model, effort} tier profile to the shipped agent definition
	// files so the tier switch takes effect on agents (mirrors `moai update`).
	a.applyTierProfileToAgents(w, r)
}

// applyTierProfileToAgents resolves the effective {plan_type, tier} from the
// persisted llm.yaml and runs template.ApplyTierProfile over the shipped agent
// .md files, mirroring the internal/cli/update.go:417-420 reference path. The
// manifest.Manager is constructed inline (no app-struct injection needed): a
// non-initialized project (manifest.Load returns an empty manifest, or no agents
// directory) degrades to a graceful no-op inside ApplyTierProfile. On error it
// surfaces a 500 (the llm.yaml write already succeeded, but the agent re-apply
// is the user-observable half of the change).
//
// @MX:NOTE: [AUTO] applyTierProfileToAgents — web entry point for ApplyTierProfile; mirrors update.go resolution
func (a *app) applyTierProfileToAgents(w http.ResponseWriter, r *http.Request) {
	projectRoot := a.cfg.ProjectRoot
	mgr := manifest.NewManager()
	// Load tolerates an absent manifest.json (returns an empty manifest, nil).
	// The only error path is a corrupt manifest.json — which Load already
	// auto-recovers (renames the corrupt file + seeds an empty manifest), leaving
	// the manager usable. So the error is safe to ignore: ApplyTierProfile degrades
	// to a no-op when no agents directory exists (non-initialized project).
	_, _ = mgr.Load(projectRoot)
	planType := template.ResolveProjectPlanType(projectRoot)
	tier := template.ResolveProjectPerformanceTier(projectRoot)
	if err := template.ApplyTierProfile(projectRoot, planType, tier, mgr); err != nil {
		a.renderError(w, http.StatusInternalServerError, "could not re-apply tier profile to agents: "+err.Error())
		return
	}
	http.Redirect(w, r, "/model-policy", http.StatusSeeOther)
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

	activePlan := cfg.LLM.EffectivePlanType()
	view := modelPolicyView{
		BindAddr:        a.resolveBindAddr(),
		ActivePlanType:  activePlan,
		PlanTypeIsEmpty: strings.TrimSpace(cfg.LLM.PlanType) == "",
		PlanPreviews:    buildPlanPreviews(activePlan),
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
