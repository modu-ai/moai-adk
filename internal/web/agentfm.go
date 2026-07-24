package web

// SPEC-WEB-CONSOLE-011 M3: sub-agent frontmatter 편집 표면의 웹 배선
// (REQ-WC11-025/027..029). 영속화는 internal/settings/agentfm 레이어
// (frontmatter-only patch, body byte-무접촉, live 파일 전용 — template
// dual-write 없음)가 담당하고, 여기는 파싱/검증/뷰 시딩만 한다.
//
// 검증은 v4manifest closed sets를 직접 재사용한다 (REQ-WC11-024/029 —
// model ∈ {inherit, haiku, sonnet, opus} ∪ {fable}, effort ∈ {low, medium,
// high, xhigh, max} ∪ {absent}; 옵션 목록 재선언 금지). fable은 v4manifest에
// 아직 상수가 없어 웹 레이어 로컬 상수(modelFable)로 추가된 Mythos급 모델 옵션이다.
//
// M5-a B6: 편집 대상 디렉터리를 다중(moai/ + harness/)으로 확장했다 — 8
// retained agents는 .claude/agents/moai/ 에, harness specialists는
// .claude/agents/harness/ 에 위치한다 (namespace doctrine, CLAUDE.local.md §24).

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
	"github.com/modu-ai/moai-adk/internal/settings/agentfm"
	"github.com/modu-ai/moai-adk/internal/template"
)

// modelFable은 상위 Mythos급 모델 옵션의 폼 와이어 값이다. v4manifest에 아직
// 대응 상수가 없어 웹 레이어에서 로컬 상수로 정의한다.
const modelFable = "fable"

// perfTierCustom is the client-only "Custom" pseudo-state wire value for the
// perf-tier radio group (G3-4). It is NOT a member of ValidPerformanceTiers —
// selecting it means "keep the current profile and its per-agent overrides".
const perfTierCustom = "custom"

// agentFMModelValues / agentFMEffortValues는 agentfm override select의 폼 옵션
// 값이다. G3 repoint 이후 per-agent 편집은 llm.agent_overrides 로 영속화되므로,
// model 옵션은 override-valid 집합 {inherit, haiku, sonnet, opus, fable} 에 맞춘다
// (config.validOverrideModels 와 동일 집합). haiku 는 명시적 per-agent 경제성
// 선택지로 허용되며, 정렬은 저렴한 모델부터 inherit → haiku → sonnet → opus →
// fable 순이다.
func agentFMModelValues() []string {
	return []string{v4manifest.ModelInherit, v4manifest.ModelHaiku, v4manifest.ModelSonnet, v4manifest.ModelOpus, modelFable}
}

func agentFMEffortValues() []string {
	return []string{v4manifest.EffortLow, v4manifest.EffortMedium, v4manifest.EffortHigh, v4manifest.EffortXhigh, v4manifest.EffortMax}
}

// agentFMPerfTierOptions returns the closed-set performance-tier selector
// options (max/medium/low — the No-Haiku 3-tier vocabulary validated by
// template.IsValidPerformanceTier). DISTINCT from the Launch section's
// model_policy (init-time agent frontmatter policy, high/medium/low).
func agentFMPerfTierOptions() []string {
	return template.ValidPerformanceTiers()
}

// parsePerfTierForm parses the profile selector (hosted as the performance_tier
// wire field) at the top of the agentfm panel. An empty submission preserves the
// current value (mirrors the agentfm empty=preserve convention); a non-empty
// out-of-set value is rejected as a per-field error joining the existing
// atomic-reject mechanism (no partial persistence). The selector value is one of
// {max, medium, low} — the active per-agent model+effort profile column.
func parsePerfTierForm(r *http.Request) (perfTier string, errs map[string]string) {
	errs = map[string]string{}
	perfTier = strings.TrimSpace(r.PostFormValue("performance_tier"))
	// "custom" is the client-only pseudo-state (G3-4) — it is NOT a persisted tier
	// (ValidPerformanceTiers stays {max,medium,low}). Treat it as "preserve the
	// current profile" so the submitted per-agent overrides carry the Custom state.
	if perfTier == perfTierCustom {
		return "", errs
	}
	if perfTier != "" && !template.IsValidPerformanceTier(perfTier) {
		errs["performance_tier"] = "invalid option"
	}
	return perfTier, errs
}

// applyPerfTierEdits persists the selected profile (max/medium/low) to
// llm.profile — and the legacy llm.performance_tier alias — when non-empty.
//
// SPEC-MODEL-PROFILE-MATRIX-001 REQ-MPM-040: this path NO LONGER mutates agent
// `.md` frontmatter. The former tier-profile re-application (which rewrote each
// shipped agent's model:/effort: and re-introduced the [1m]-hazard concrete-model
// pin) is retired — the web save now persists to llm.yaml only, leaving agent
// frontmatter at `model: inherit`. The runtime resolver (`moai model profile`)
// reads the profile matrix, not a mutated frontmatter pin.
func applyPerfTierEdits(projectRoot, perfTier string) error {
	if perfTier == "" {
		return nil
	}
	if !template.IsValidPerformanceTier(perfTier) {
		return nil // defensive — parse already rejected out-of-set values
	}
	// Persist the legacy alias (kept as the Tier x Phase perfTier source) and the
	// new profile key. Both are no-op regex replaces when the value is unchanged.
	if err := template.ApplyPerformanceTier(projectRoot, perfTier); err != nil {
		return fmt.Errorf("agentfm: apply performance_tier: %w", err)
	}
	if err := template.ApplyProfile(projectRoot, perfTier); err != nil {
		return fmt.Errorf("agentfm: apply profile: %w", err)
	}
	return nil
}

// agentBadgeInfo carries the M5 tier-badge display data for one agent row
// (SPEC-WEBCONF-SIMPLIFY-001 M5, REQ-WC-006/007/008, design.md §E).
type agentBadgeInfo struct {
	Glyph      string // emoji glyph (🔴🟠🔵🩵) or "custom" for the override-sentinel badge
	TooltipKey string // fieldDesc.agentfm.<tier|custom> i18n key (M4 description mechanism)
	HasBadge   bool   // false for unmapped agents (EC-6 — no badge rendered)
	IsCustom   bool   // true when model=inherit or effort=max (EC-2 / AC-WC-018 neutral badge)
}

// agentTierBadge computes the display-only tier badge for an agent row. The badge
// comes from the name-keyed lookup table (design.md §C — Option A, NOT from the
// agent's effort file). When the agent's current effort is the override sentinel
// `max`, the badge is a neutral "custom" marker (EC-2, AC-WC-018).
//
// `model: inherit` is NOT an override signal: SPEC-MODEL-PROFILE-MATRIX-001
// REQ-MPM-040 stopped mutating agent frontmatter (see applyPerfTierEdits above),
// so `inherit` is the SHIPPED default for the core agents. Treating it as a
// manual override mislabeled those defaults CUSTOM and produced the inconsistent
// mix of CUSTOM pills and tier glyphs across the agent rows. The model parameter
// is retained (unused) so the call sites keep passing the row's full state.
func agentTierBadge(name, model, effort string) agentBadgeInfo {
	if effort == v4manifest.EffortMax {
		return agentBadgeInfo{Glyph: "custom", TooltipKey: "fieldDesc.agentfm.custom", HasBadge: true, IsCustom: true}
	}
	tier, ok := v4manifest.AgentTier(name)
	if !ok {
		return agentBadgeInfo{} // unmapped agent — no badge (EC-6)
	}
	return agentBadgeInfo{
		Glyph:      v4manifest.TierColor(tier),
		TooltipKey: "fieldDesc.agentfm." + string(tier),
		HasBadge:   true,
	}
}

// agentTierSortRank returns a sort rank for the agent's display-only tier
// (lower = more expensive = renders first within each sub-tab). Used to order
// the agentfm rows red-first (post-close polish round 3,
// SPEC-WEBCONF-SIMPLIFY-001). red=0, orange=1, blue=2, lightblue=3,
// unmapped=4 (agents without a tier table entry sort last).
func agentTierSortRank(name string) int {
	tier, ok := v4manifest.AgentTier(name)
	if !ok {
		return 4
	}
	switch tier {
	case v4manifest.TierRed:
		return 0
	case v4manifest.TierOrange:
		return 1
	case v4manifest.TierBlue:
		return 2
	case v4manifest.TierLightBlue:
		return 3
	default:
		return 4
	}
}

// agentIsMoaiCore reports whether the agent lives in .claude/agents/moai/ (the 10
// core agents) vs .claude/agents/harness/ (the 10 harness specialists). Derived
// from the source directory path.
func agentIsMoaiCore(info agentfm.AgentInfo) bool {
	return strings.Contains(info.Path, string(filepath.Separator)+"moai"+string(filepath.Separator))
}

// agentFMRenderCount reports how many of the listed agents actually render in the
// agentfm section. Only the .claude/agents/moai/ rows render (the harness sub-tab
// was removed), so the section count must report that subset rather than the full
// scanned catalog. The harness agents stay in the scan — an unrendered agent
// simply submits no form values, which parseAgentFMForm reads as "preserve".
func agentFMRenderCount(agents []agentfm.AgentInfo) int {
	n := 0
	for _, a := range agents {
		if agentIsMoaiCore(a) {
			n++
		}
	}
	return n
}

// agentResolvedModel returns the model for the named agent resolved through the
// runtime profile matrix (template.ResolveAgentModelEffort): an llm.agent_overrides
// entry wins, else the active profile's group cell, else the Go-default cell. This
// is the SINGLE SOURCE OF TRUTH the console now reads from — NOT the agent
// frontmatter and NOT the badge-tier suggestions (G3-1 repoint).
func agentResolvedModel(llm config.LLMConfig, name string) string {
	me, _ := template.ResolveAgentModelEffort(llm, name)
	return me.Model
}

// agentResolvedEffort returns the effort resolved through the profile matrix.
func agentResolvedEffort(llm config.LLMConfig, name string) string {
	me, _ := template.ResolveAgentModelEffort(llm, name)
	return me.Effort
}

// agentSelectedModel returns the model the row's select should display as
// selected — the profile-matrix-resolved value (G3-1).
func agentSelectedModel(llm config.LLMConfig, info agentfm.AgentInfo) string {
	return agentResolvedModel(llm, info.Name)
}

// agentSelectedEffort returns the effort the row's select should display as
// selected — the profile-matrix-resolved value (G3-1).
func agentSelectedEffort(llm config.LLMConfig, info agentfm.AgentInfo) string {
	return agentResolvedEffort(llm, info.Name)
}

// agentModelIsHaiku reports whether the agent's profile-matrix-resolved model is
// haiku. Haiku does not honor reasoning effort, so the paired effort select is
// disabled in that case (fieldsets.templ) with a muted inline hint. The save path
// is unaffected: a disabled select does not submit, and parseAgentFMForm backfills
// an unsubmitted effort with the resolved value (empty=preserve), so nothing is
// corrupted or dropped.
func agentModelIsHaiku(llm config.LLMConfig, info agentfm.AgentInfo) bool {
	return agentResolvedModel(llm, info.Name) == v4manifest.ModelHaiku
}

// NOTE: agentHasOverride / agentModelIsDefault / agentEffortIsDefault were
// removed with the per-row "(default)" caption — the profile-vs-override
// distinction is already carried by the tier badge, so the extra tag was noise.

// profileMatrixData returns the per-profile per-agent {model, effort} matrix for
// the client-side tier-repopulation handler (G3-3), emitted via templ.JSONScript.
// Shape: {"<tier>": {"<agent>": {"model": "...", "effort": "..."}}}. It is static
// (derived from the Go default matrix), so the same data is emitted on every render.
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
// frontmatter description for display in the agentfm row (post-close polish,
// SPEC-WEBCONF-SIMPLIFY-001). It collapses whitespace, extracts the first
// sentence (up to the first ". " boundary), and truncates to ~140 characters
// with an ellipsis. Empty/whitespace-only input returns "" (the templ skips
// rendering when empty — graceful handling of agents lacking a description).
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

// agentDirsFor는 frontmatter 편집 대상 디렉터리 목록이다 (live 파일 전용).
// 10 moai-custom agents는 .claude/agents/moai/ 에, 10 harness specialists는
// .claude/agents/harness/ 에 위치한다 (20 total; namespace doctrine, CLAUDE.local.md §24).
// M5-a B6 — 단일 디렉터리에서 다중 디렉터리 해상도로 확장.
func agentDirsFor(projectRoot string) []string {
	return []string{
		filepath.Join(projectRoot, ".claude", "agents", "moai"),
		filepath.Join(projectRoot, ".claude", "agents", "harness"),
	}
}

// listAllAgentFMs는 모든 편집 대상 디렉터리(moai/ + harness/)의 agent
// frontmatter를 tier-then-name 순으로 병합한다 (post-close polish round 3 —
// 비싼 티어가 먼저: 🔴 → 🟠 → 🔵 → 🩵). 개별 디렉터리는 이미 정렬되어 반환되지만
// 두 디렉터리를 병합한 후 전체를 다시 정렬한다. templ이 moai/harness 파티션을
// 걸면서 상대 순서를 보존하므로 각 서브탭 내에서도 tier 순이 유지된다.
func (a *app) listAllAgentFMs(projectRoot string) ([]agentfm.AgentInfo, error) {
	var all []agentfm.AgentInfo
	for _, dir := range agentDirsFor(projectRoot) {
		agents, err := a.listAgentFMs(dir)
		if err != nil {
			return nil, err
		}
		all = append(all, agents...)
	}
	sort.Slice(all, func(i, j int) bool {
		ri, rj := agentTierSortRank(all[i].Name), agentTierSortRank(all[j].Name)
		if ri != rj {
			return ri < rj
		}
		return all[i].Name < all[j].Name
	})
	return all, nil
}

// parseAgentFMForm parses the per-agent model/effort submissions and computes the
// desired llm.agent_overrides state for the RENDERED, profile-matrix-member agents
// (G3-2 repoint — persistence moved off agent frontmatter onto llm.agent_overrides).
//
// For each submitted agent, the (model, effort) is compared against the profile
// default for the target tier: a value equal to the default is NOT pinned (its
// override is cleared — reset-to-named-tier), any other value is a pin. An agent
// with no submitted fields is preserved (empty=preserve). Non-matrix agents (no
// group membership) are skipped entirely — config.validateAgentOverrides only
// accepts retained-catalog names.
//
// Returns: the pin map (agents that differ from the profile default), the set of
// submitted matrix-member agent names (the clear scope for applyAgentOverrides),
// and per-field validation errors joining the atomic-reject flow. `targetTier` is
// the perf-tier the cells will resolve under after the save ("" = the client
// Custom pseudo-state or no tier change → the current effective profile). The form
// names carry a path-traversal guard (defensive — List builds these names).
func parseAgentFMForm(r *http.Request, agents []agentfm.AgentInfo, llm config.LLMConfig, targetTier string) (map[string]config.ModelEffort, []string, map[string]string) {
	pins := map[string]config.ModelEffort{}
	var submitted []string
	errs := map[string]string{}

	// Resolve the profile-default base under the tier that will be active after
	// this save. A Custom/empty tier submission keeps the current profile.
	baseProfile := targetTier
	if baseProfile == "" {
		baseProfile = llm.EffectiveProfile()
	}
	base := config.LLMConfig{Profile: baseProfile, Profiles: llm.Profiles}

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
		// Validate the submitted values for ANY agent (out-of-set → atomic reject),
		// BEFORE the matrix-membership skip below, so garbage input is rejected 4xx
		// even for a non-matrix agent.
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
		// (config.validateAgentOverrides rejects non-retained names); a valid but
		// non-matrix submission is ignored (no override, no frontmatter write).
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

// applyAgentOverrides persists per-agent pins to llm.agent_overrides (G3-2 —
// replaces the retired frontmatter-patch path). It round-trips through the config
// manager (LoadRaw → mutate cfg.LLM.AgentOverrides → SetSection("llm") → Save), the
// same typed path the glm.models.* edits use, so every llm.yaml field is preserved
// by the struct re-marshal. Submitted agents equal to the profile default have
// their override CLEARED; pinned agents are set. When nothing actually changes the
// write is skipped entirely (llm.yaml byte-identity preserved for empty/no-op
// submissions). Overrides for agents outside the submitted set are preserved.
func applyAgentOverrides(projectRoot string, pins map[string]config.ModelEffort, submitted []string) error {
	if len(pins) == 0 && len(submitted) == 0 {
		return nil // nothing submitted → no write
	}
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("agentfm: load config: %w", err)
	}
	ov := cfg.LLM.AgentOverrides
	if ov == nil {
		ov = map[string]config.ModelEffort{}
	}
	changed := false
	// Clear submitted agents that are NOT pins (they matched the profile default).
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
	cfg.LLM.AgentOverrides = ov
	if err := mgr.SetSection("llm", cfg.LLM); err != nil {
		return fmt.Errorf("agentfm: set llm section: %w", err)
	}
	if err := mgr.Save(); err != nil {
		return fmt.Errorf("agentfm: save config: %w", err)
	}
	return nil
}
