package web

// agentfm.go — the restored console surface's parse/persist helpers
// (SPEC-WEB-AGENTFM-RESTORE-001 M3; re-port of the file SPEC-WEB-CONSOLE-011
// M3 originally built, deleted by 384eb3460). Re-port, not revert (plan
// §D.4): the persist path writes llm.yaml through the settings seam
// (settings.WriteLLMProfile / settings.WriteLLMAgentOverrides) — the old
// SetSection("llm")→Save() full re-marshal stays retired (the GitHub issue
// #1731 defect mechanism), the legacy llm.performance_tier alias is never
// written (plan §D.4-3), and agent frontmatter is never mutated (REQ-AFR-005
// — the settings/agentfm Patch layer stays dead).
//
// Resolution reuses template.ResolveAgentModelEffort — the single derivation
// site (REQ-AFR-010); this package must never re-derive a per-agent cell
// (the mcp_audit_surface guard enforces the definition ban).
//
// The M3 half above is the save path; the M4 half below is the render surface
// the fieldsets.templ sub-section consumes.

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
	"github.com/modu-ai/moai-adk/internal/settings"
	"github.com/modu-ai/moai-adk/internal/settings/agentfm"
	"github.com/modu-ai/moai-adk/internal/template"
)

// modelFable은 상위 모델 옵션의 폼 와이어 값이다. v4manifest에 대응 상수가
// 없어 웹 레이어에서 로컬 상수로 정의한다 (config에도 fable *상수*는 없다 —
// GLMModels.Fable 필드만 있다; plan REQ-AFR-006).
const modelFable = "fable"

// perfTierCustom is the client-only "Custom" pseudo-state wire value for the
// perf-tier radio group (G3-4). It is NOT a member of ValidPerformanceTiers —
// selecting it means "keep the current profile and its per-agent overrides".
const perfTierCustom = "custom"

// agentOverridesSectionMarker is the data-section slice marker the rendered
// sub-section carries, so a test can slice it the way panelHTML slices a
// panel (REQ-AFR-013).
const agentOverridesSectionMarker = "agent-overrides"

// agentFMPerfTierDefault is the column the empty profile resolves to — the
// "(runtime default)" hint's value (templ references it bare).
var agentFMPerfTierDefault = template.PerformanceTierMedium

// agentFMModelValues / agentFMEffortValues는 agentfm override select의 폼 옵션
// 값이다. per-agent 편집은 llm.agent_overrides 로 영속화되므로, model 옵션은
// override-valid 집합 {inherit, haiku, sonnet, opus, fable} 에 맞춘다
// (config.validOverrideModels 와 동일 집합 — REQ-AFR-006). haiku 는 명시적
// per-agent 경제성 선택지로 허용되며, 정렬은 저렴한 모델부터
// inherit → haiku → sonnet → opus → fable 순이다.
func agentFMModelValues() []string {
	return []string{v4manifest.ModelInherit, v4manifest.ModelHaiku, v4manifest.ModelSonnet, v4manifest.ModelOpus, modelFable}
}

func agentFMEffortValues() []string {
	return []string{v4manifest.EffortLow, v4manifest.EffortMedium, v4manifest.EffortHigh, v4manifest.EffortXhigh, v4manifest.EffortMax}
}

// parsePerfTierForm parses the profile selector (hosted as the performance_tier
// wire field) at the top of the agent-overrides sub-section. An empty
// submission preserves the current value (mirrors the agentfm empty=preserve
// convention); the "custom" pseudo-state preserves too (it only means the
// submitted per-agent overrides carry the Custom state); a non-empty
// out-of-set value is rejected as a per-field error joining the existing
// atomic-reject mechanism (no partial persistence). The selector value is one
// of {max, medium, low} — the active per-agent model+effort profile column.
func parsePerfTierForm(r *http.Request) (perfTier string, errs map[string]string) {
	errs = map[string]string{}
	perfTier = strings.TrimSpace(r.PostFormValue("performance_tier"))
	if perfTier == perfTierCustom {
		return "", errs
	}
	if perfTier != "" && !template.IsValidPerformanceTier(perfTier) {
		errs["performance_tier"] = "invalid option"
	}
	return perfTier, errs
}

// applyPerfTierEdits persists the selected profile to llm.profile when
// non-empty. Re-port correction (plan §D.4-3): the retired
// llm.performance_tier alias is NOT written — the old applyPerformanceTier
// leg had no surviving consumer. The persist goes through the shared settings
// seam with its no-op gate (an equal value writes nothing).
func applyPerfTierEdits(projectRoot, perfTier string) error {
	if perfTier == "" {
		return nil
	}
	if !template.IsValidPerformanceTier(perfTier) {
		return nil // defensive — parse already rejected out-of-set values
	}
	if err := settings.WriteLLMProfile(projectRoot, perfTier); err != nil {
		return fmt.Errorf("agentfm: apply profile: %w", err)
	}
	return nil
}

// agentDirsFor는 frontmatter 스캔 대상 디렉터리 목록이다 (live 파일 전용).
// moai agents는 .claude/agents/moai/ 에, harness specialists는
// .claude/agents/harness/ 에 위치한다 (namespace doctrine). 하니스 행은
// 스캔만 하고 렌더하지 않는다 (REQ-AFR-001).
func agentDirsFor(projectRoot string) []string {
	return []string{
		filepath.Join(projectRoot, ".claude", "agents", "moai"),
		filepath.Join(projectRoot, ".claude", "agents", "harness"),
	}
}

// agentGroupRank classifies an agent name into one of 5 display buckets for
// the agentfm row grouping (lower rank renders first). The named retained
// agents map to their catalog class; every other agent (harness specialists
// under .claude/agents/harness/, or anything unclassified) falls into the
// final "other" bucket. Classification is by name — the closed retained
// catalog is name-stable, so a path check is unnecessary at this layer (the
// harness/ directory carries hns-* names that are not in the map below and so
// default to the other bucket).
func agentGroupRank(name string) int {
	switch name {
	// core/manager
	case "manager-spec", "manager-develop", "manager-docs", "manager-git", "manager-design", "manager-todo":
		return 0
	// meta/evaluator
	case "plan-auditor", "sync-auditor", "super-advisor":
		return 1
	// builder
	case "builder-harness":
		return 2
	// specialist
	case "e2e-tester":
		return 3
	}
	return 4 // other (harness specialists, …)
}

// agentModelCostRank maps a resolved model to an ascending cost rank for
// cheap-first ordering within a group. fable / inherit / unknown default to a
// large rank so they sort last within their group.
func agentModelCostRank(model string) int {
	switch model {
	case v4manifest.ModelHaiku:
		return 0
	case v4manifest.ModelSonnet:
		return 1
	case v4manifest.ModelOpus:
		return 2
	}
	return 3 // fable / inherit / unknown — sort last within group
}

// agentEffortCostRank maps a resolved effort to an ascending cost rank for
// cheap-first ordering within a group: low < medium < high < max. An empty
// effort defaults to medium (the most common resolved value) so unset rows do
// not float to either extreme.
func agentEffortCostRank(effort string) int {
	switch effort {
	case v4manifest.EffortLow:
		return 1
	case v4manifest.EffortMedium:
		return 2
	case v4manifest.EffortHigh:
		return 3
	case v4manifest.EffortMax:
		return 4
	}
	return 2 // empty → medium tier
}

// listAllAgentFMs는 모든 스캔 대상 디렉터리(moai/ + harness/)의 agent
// frontmatter를 group → cheap-first model cost → cheap-first effort cost → name
// 순으로 병합한다. 정렬은 profile-matrix-resolved model/effort 기준 — agent
// frontmatter는 model: inherit 이라 frontmatter Model 로는 구분이 안 되므로
// resolved 값이 SSOT다.
func (a *app) listAllAgentFMs(projectRoot string, llm config.LLMConfig) ([]agentfm.AgentInfo, error) {
	var all []agentfm.AgentInfo
	for _, dir := range agentDirsFor(projectRoot) {
		agents, err := a.listAgentFMs(dir)
		if err != nil {
			return nil, err
		}
		all = append(all, agents...)
	}
	sort.Slice(all, func(i, j int) bool {
		ai, aj := all[i], all[j]
		gi, gj := agentGroupRank(ai.Name), agentGroupRank(aj.Name)
		if gi != gj {
			return gi < gj
		}
		mi, mj := agentModelCostRank(agentResolvedModel(llm, ai.Name)), agentModelCostRank(agentResolvedModel(llm, aj.Name))
		if mi != mj {
			return mi < mj
		}
		ei, ej := agentEffortCostRank(agentResolvedEffort(llm, ai.Name)), agentEffortCostRank(agentResolvedEffort(llm, aj.Name))
		if ei != ej {
			return ei < ej
		}
		return ai.Name < aj.Name
	})
	return all, nil
}

// agentResolvedModel returns the model for the named agent resolved through
// the runtime profile matrix (template.ResolveAgentModelEffort): an
// llm.agent_overrides entry wins, else the Go-default cell. This is the
// SINGLE SOURCE OF TRUTH the console reads from — NOT the agent frontmatter
// and NOT a second derivation.
func agentResolvedModel(llm config.LLMConfig, name string) string {
	me, _ := template.ResolveAgentModelEffort(llm, name)
	return me.Model
}

// agentResolvedEffort returns the effort resolved through the profile matrix.
func agentResolvedEffort(llm config.LLMConfig, name string) string {
	me, _ := template.ResolveAgentModelEffort(llm, name)
	return me.Effort
}

// parseAgentFMForm parses the per-agent model/effort submissions and computes
// the desired llm.agent_overrides state for the RENDERED, profile-matrix-member
// agents.
//
// For each submitted agent, the (model, effort) is compared against the
// profile default for the target tier: a value equal to the default is NOT
// pinned (its override is cleared — reset-to-named-tier), any other value is
// a pin. An agent with no submitted fields is preserved (empty=preserve).
// Non-matrix agents (no group membership) are skipped entirely —
// config.validateAgentOverrides only accepts retained-catalog names.
//
// Returns: the pin map (agents that differ from the profile default), the set
// of submitted matrix-member agent names (the clear scope for
// applyAgentOverrides), and per-field validation errors joining the
// atomic-reject flow. `targetTier` is the perf-tier the cells will resolve
// under after the save ("" = the client Custom pseudo-state or no tier change
// → the current effective profile). The form names carry a path-traversal
// guard (defensive — List builds these names).
func parseAgentFMForm(r *http.Request, agents []agentfm.AgentInfo, llm config.LLMConfig, targetTier string) (map[string]config.ModelEffort, []string, map[string]string) {
	pins := map[string]config.ModelEffort{}
	var submitted []string
	errs := map[string]string{}

	// Resolve the profile-default base under the tier that will be active
	// after this save. A Custom/empty tier submission keeps the current
	// profile.
	baseProfile := targetTier
	if baseProfile == "" {
		baseProfile = llm.EffectiveProfile()
	}
	base := config.LLMConfig{Profile: baseProfile}

	for _, a := range agents {
		if !a.ParseOK {
			continue // 파싱 실패 행은 편집 비활성 (design.md §C.1 견고성)
		}
		if strings.ContainsAny(a.Name, "/\\") || strings.Contains(a.Name, "..") {
			continue // 방어적 — List가 만든 이름이라 정상 경로에선 불가능
		}

		m := strings.TrimSpace(r.PostFormValue("agentfm." + a.Name + ".model"))
		e := strings.TrimSpace(r.PostFormValue("agentfm." + a.Name + ".effort"))
		if m == "" && e == "" {
			continue // 미제출 → preserve (empty=preserve)
		}
		// Validate the submitted values for ANY agent (out-of-set → atomic
		// reject), BEFORE the matrix-membership skip below, so garbage input
		// is rejected even for a non-matrix agent.
		if m != "" && !inList(agentFMModelValues(), m) {
			errs["agentfm."+a.Name+".model"] = "invalid option"
		}
		if e != "" && !inList(agentFMEffortValues(), e) {
			errs["agentfm."+a.Name+".effort"] = "invalid option"
		}
		if errs["agentfm."+a.Name+".model"] != "" || errs["agentfm."+a.Name+".effort"] != "" {
			continue
		}
		// Overrides are only valid for profile-matrix member agents
		// (config.validateAgentOverrides rejects non-retained names); a valid
		// but non-matrix submission is ignored (no override, no write).
		if _, ok := template.AgentGroup(a.Name); !ok {
			continue
		}

		submitted = append(submitted, a.Name)
		if m == "" {
			m = agentResolvedModel(llm, a.Name)
		}
		if e == "" {
			e = agentResolvedEffort(llm, a.Name)
		}
		def, _ := template.ResolveAgentModelEffort(base, a.Name)
		if m == def.Model && e == def.Effort {
			continue // equals the profile default → cleared, not pinned
		}
		pins[a.Name] = config.ModelEffort{Model: m, Effort: e}
	}
	return pins, submitted, errs
}

// applyAgentOverrides persists per-agent pins to llm.agent_overrides. The
// desired state (pins minus submitted agents that matched the profile
// default) merges over the CURRENT map read from llm.yaml, and the full
// result persists through the settings block-splice seam — every llm.yaml
// byte outside the agent_overrides block survives. When nothing actually
// changes the write is skipped entirely (llm.yaml byte-identity preserved for
// empty/no-op submissions). Overrides for agents outside the submitted set
// are preserved.
func applyAgentOverrides(projectRoot string, pins map[string]config.ModelEffort, submitted []string) error {
	if len(pins) == 0 && len(submitted) == 0 {
		return nil // nothing submitted → no write
	}
	cfg, err := config.NewConfigManager().LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("agentfm: load config: %w", err)
	}
	ov := cfg.LLM.AgentOverrides
	if ov == nil {
		ov = map[string]config.ModelEffort{}
	}
	changed := false
	// Clear submitted agents that are NOT pins (they matched the profile
	// default).
	for _, name := range submitted {
		if _, isPin := pins[name]; isPin {
			continue
		}
		if _, ok := ov[name]; ok {
			delete(ov, name)
			changed = true
		}
	}
	// Set the pins.
	for name, me := range pins {
		if cur, ok := ov[name]; !ok || cur != me {
			ov[name] = me
			changed = true
		}
	}
	if !changed {
		return nil
	}
	if err := settings.WriteLLMAgentOverrides(projectRoot, ov); err != nil {
		return fmt.Errorf("agentfm: write agent overrides: %w", err)
	}
	return nil
}

// ─── M4 render half (SPEC-WEB-AGENTFM-RESTORE-001) ──────────────────────────

// agentFMPerfTierOptions returns the closed-set performance-tier selector
// options ({max, medium, low} — validated by template.IsValidPerformanceTier).
// DISTINCT from the Launch section's model_policy (the session model axis).
func agentFMPerfTierOptions() []string {
	return template.ValidPerformanceTiers()
}

// agentfmBadgeGlyphs is the LOCAL model→glyph map (plan §A.1: the
// v4manifest.ModelColor display helper died with the deletion and is NOT
// re-ported — the glyphs are a display concern, restated here once).
var agentfmBadgeGlyphs = map[string]string{
	v4manifest.ModelOpus:   "🔴",
	v4manifest.ModelSonnet: "🟠",
	v4manifest.ModelHaiku:  "🔵",
}

// agentBadgeInfo carries the model-badge display data for one agent row.
type agentBadgeInfo struct {
	Glyph      string // emoji glyph derived from the model, or "custom"
	TooltipKey string // fieldDesc.agentfm.model.<model> / fieldDesc.agentfm.custom i18n key
	HasBadge   bool   // false only when the row carries no usable model state
	IsCustom   bool   // true when effort=max (neutral "custom" badge)
}

// agentTierBadge computes the display-only badge for an agent row. The badge
// is derived from the agent's resolved model (the profile-matrix SSOT via
// agentResolvedModel) — NOT a name→tier table — so the badge tracks the model
// the matrix actually assigns. When the agent's current effort is the
// override sentinel `max`, the badge is a neutral "custom" marker. The name
// parameter is retained so call sites keep passing the row's full state; it
// is not used now that the badge is model-derived.
func agentTierBadge(name, model, effort string) agentBadgeInfo {
	_ = name
	if effort == v4manifest.EffortMax {
		return agentBadgeInfo{Glyph: "custom", TooltipKey: "fieldDesc.agentfm.custom", HasBadge: true, IsCustom: true}
	}
	glyph, ok := agentfmBadgeGlyphs[model]
	if !ok {
		glyph = "🩵" // inherit / unknown
	}
	return agentBadgeInfo{
		Glyph:      glyph,
		TooltipKey: "fieldDesc.agentfm.model." + model,
		HasBadge:   true,
	}
}

// agentGroupLabel returns the human-facing group header label for the named
// agent's catalog class. Returns "" for the "other" bucket — those rows
// (harness specialists) are filtered out by agentIsMoaiCore and never reach
// the grid. The label is the server-rendered English baseline (a structural
// taxonomy heading, like the agent names themselves — not a data-i18n node).
func agentGroupLabel(name string) string {
	switch agentGroupRank(name) {
	case 0:
		return "CORE / MANAGER"
	case 1:
		return "META / EVALUATOR"
	case 2:
		return "BUILDER"
	case 3:
		return "SPECIALIST"
	}
	return ""
}

// agentFMGridRow is one render entry in the agentfm grid: the agent plus the
// group header label to emit BEFORE this row ("" = no group header — the row
// stays in the same group as the previous rendered row). Precomputing the
// group-change signal in Go keeps the templ grid loop state-free.
type agentFMGridRow struct {
	Agent     agentfm.AgentInfo
	GroupHead string // label to render as a divider above this row; "" = none
}

// agentIsMoaiCore reports whether the agent lives in .claude/agents/moai/
// (the retained agents) vs .claude/agents/harness/ (the specialists). Derived
// from the source directory path — harness rows are scanned but never
// rendered (REQ-AFR-001).
func agentIsMoaiCore(info agentfm.AgentInfo) bool {
	return strings.Contains(info.Path, string(filepath.Separator)+"moai"+string(filepath.Separator))
}

// agentFMGridRows walks the (already group-sorted) agent list, filters to the
// moai-core rows that actually render, and attaches a GroupHead to the first
// row of each catalog-class group.
func agentFMGridRows(agents []agentfm.AgentInfo) []agentFMGridRow {
	out := make([]agentFMGridRow, 0, len(agents))
	prev := ""
	for _, a := range agents {
		if !agentIsMoaiCore(a) {
			continue
		}
		head := ""
		if g := agentGroupLabel(a.Name); g != prev {
			head = g
			prev = g
		}
		out = append(out, agentFMGridRow{Agent: a, GroupHead: head})
	}
	return out
}

// agentFMRenderCount reports how many of the listed agents actually render in
// the agent-overrides sub-section. Only the .claude/agents/moai/ rows render,
// so the section count must report that subset rather than the full scanned
// catalog. The harness agents stay in the scan — an unrendered agent simply
// submits no form values, which parseAgentFMForm reads as "preserve".
func agentFMRenderCount(agents []agentfm.AgentInfo) int {
	n := 0
	for _, a := range agents {
		if agentIsMoaiCore(a) {
			n++
		}
	}
	return n
}

// agentSelectedModel returns the model the row's select should display as
// selected — the profile-matrix-resolved value.
func agentSelectedModel(llm config.LLMConfig, info agentfm.AgentInfo) string {
	return agentResolvedModel(llm, info.Name)
}

// agentSelectedEffort returns the effort the row's select should display as
// selected — the profile-matrix-resolved value.
func agentSelectedEffort(llm config.LLMConfig, info agentfm.AgentInfo) string {
	return agentResolvedEffort(llm, info.Name)
}

// agentModelIsHaiku reports whether the agent's profile-matrix-resolved model
// is haiku. Haiku does not honor reasoning effort, so the paired effort
// select is disabled in that case (fieldsets.templ) with a muted inline hint.
// The save path is unaffected: a disabled select does not submit, and
// parseAgentFMForm backfills an unsubmitted effort with the resolved value.
func agentModelIsHaiku(llm config.LLMConfig, info agentfm.AgentInfo) bool {
	return agentResolvedModel(llm, info.Name) == v4manifest.ModelHaiku
}

// agentFMIsGLMBackend reports whether the live llm.yaml marks a GLM session
// backend (team_mode glm/cg, or the dormant mode field). Thin reuse of the
// config-level intent signal so the panel gate and the CLI resolver read the
// same predicate.
func agentFMIsGLMBackend(llm config.LLMConfig) bool {
	return template.IsGLMBackend(llm)
}

// agentGLMReasoning returns the per-agent GLM reasoning state for the
// effort(reasoning) map exposure: under a GLM backend the sub-agent model
// is session-inherited, so the per-agent axis is effort — and its GLM reading
// is the z.ai reasoning state the effort collapses to. The value reuses the
// SAME overlay the session derivation uses
// (template.ResolveGLMReasoningForModel — no second derivation, REQ-AFR-010),
// keyed by the profile-matrix-resolved effort so an override row shows its
// overridden state. Returns "" under a Claude backend, where the GLM reading
// is noise.
func agentGLMReasoning(llm config.LLMConfig, name string) string {
	if !template.IsGLMBackend(llm) {
		return ""
	}
	me, _ := template.ResolveAgentModelEffort(llm, name)
	return template.ResolveGLMReasoningForModel(llm.GLM.Models.High, name, me.Effort).Name
}

// profileMatrixData returns the per-profile per-agent {model, effort} matrix
// for the client-side tier-repopulation handler (G3-3), emitted via
// templ.JSONScript. Shape: {"<tier>": {"<agent>": {"model": "...", "effort":
// "..."}}}. It is static (derived from the Go default matrix through the SAME
// resolver), so the same data is emitted on every render. Keys are the
// selector wire values ({max, medium, low}); resolution folds max→high.
func profileMatrixData() map[string]map[string]map[string]string {
	out := map[string]map[string]map[string]string{}
	for _, tier := range template.ValidPerformanceTiers() {
		base := config.LLMConfig{Profile: tier}
		cells := map[string]map[string]string{}
		for _, agent := range template.ProfileMatrixAgents() {
			me, _ := template.ResolveAgentModelEffort(base, agent)
			cells[agent] = map[string]string{"model": me.Model, "effort": me.Effort}
		}
		out[tier] = cells
	}
	return out
}

// agentDescriptionShort returns a compact one-line summary of the agent's
// frontmatter description for display in the agentfm row. It collapses
// whitespace, extracts the first sentence (up to the first ". " boundary),
// and truncates to ~140 characters with an ellipsis. Empty input returns ""
// (the templ skips rendering when empty — graceful handling of agents
// lacking a description).
func agentDescriptionShort(desc string) string {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return ""
	}
	// Collapse all whitespace (newlines from YAML block/folded scalars) to single spaces.
	desc = strings.Join(strings.Fields(desc), " ")
	// Extract the first sentence — the period-space boundary.
	if i := strings.Index(desc, ". "); i >= 0 {
		desc = desc[:i+1]
	}
	// Truncate to a compact length with an ellipsis.
	const maxLen = 140
	if len(desc) > maxLen {
		desc = desc[:maxLen-3] + "…"
	}
	return desc
}
