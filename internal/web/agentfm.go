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

	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
	"github.com/modu-ai/moai-adk/internal/settings/agentfm"
	"github.com/modu-ai/moai-adk/internal/template"
)

// agentFMAbsent는 effort "(absent)" 상태의 폼 와이어 값이다 (EC-7 — 키 삭제).
const agentFMAbsent = "__absent__"

// modelFable은 상위 Mythos급 모델 옵션의 폼 와이어 값이다. v4manifest에 아직
// 대응 상수가 없어 웹 레이어에서 로컬 상수로 정의한다 (agentfm.Patch는 model을
// 닫힌 집합으로 검증하지 않으므로 이 값으로 frontmatter 영속화가 가능하다).
const modelFable = "fable"

// agentFMModelValues / agentFMEffortValues는 v4manifest closed sets에서 파생한
// 폼 옵션 값이다 (재선언 금지 — exported 상수 재사용). fable은 v4manifest 상수가
// 없어 modelFable 로컬 상수를 tail에 덧붙인다.
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

// agentIsSuggestedModel reports whether m is the tier-suggested model for the
// named agent (design.md §D). Used to mark the suggested option in the model
// <select> when the agent's model is unset (empty=preserve sentinel).
func agentIsSuggestedModel(name, m string) bool {
	tier, ok := v4manifest.AgentTier(name)
	if !ok {
		return false
	}
	sm, _ := v4manifest.TierSuggestedModelEffort(tier)
	return sm == m
}

// agentIsSuggestedEffort reports whether e is the tier-suggested effort.
func agentIsSuggestedEffort(name, e string) bool {
	tier, ok := v4manifest.AgentTier(name)
	if !ok {
		return false
	}
	_, se := v4manifest.TierSuggestedModelEffort(tier)
	return se == e
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

// agentTierSuggestedModel returns the tier-suggested model for the agent name
// (design.md §D), or "" if the agent has no tier entry.
func agentTierSuggestedModel(name string) string {
	tier, ok := v4manifest.AgentTier(name)
	if !ok {
		return ""
	}
	sm, _ := v4manifest.TierSuggestedModelEffort(tier)
	return sm
}

// agentTierSuggestedEffort returns the tier-suggested effort.
func agentTierSuggestedEffort(name string) string {
	tier, ok := v4manifest.AgentTier(name)
	if !ok {
		return ""
	}
	_, se := v4manifest.TierSuggestedModelEffort(tier)
	return se
}

// agentSelectedModel returns the model the select should display as selected: the
// agent's explicit frontmatter model, or — if absent — the tier-suggested default
// (the "resolved effective value"). Used by the agentFMRow render.
func agentSelectedModel(info agentfm.AgentInfo) string {
	if info.Model != "" {
		return info.Model
	}
	return agentTierSuggestedModel(info.Name)
}

// agentSelectedEffort returns the effort the select should display as selected.
func agentSelectedEffort(info agentfm.AgentInfo) string {
	if info.EffortPresent {
		return info.Effort
	}
	return agentTierSuggestedEffort(info.Name)
}

// agentModelIsDefault reports whether the selected model is a derived (tier-
// suggested) default for an absent-model agent (vs an explicit frontmatter value).
func agentModelIsDefault(info agentfm.AgentInfo) bool {
	return info.Model == "" && agentTierSuggestedModel(info.Name) != ""
}

// agentEffortIsDefault reports whether the selected effort is a derived default.
func agentEffortIsDefault(info agentfm.AgentInfo) bool {
	return !info.EffortPresent && agentTierSuggestedEffort(info.Name) != ""
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

// agentFMEdit는 agent 1종의 frontmatter 편집 제출이다.
type agentFMEdit struct {
	Agent        string
	Model        string // "" = 유지
	Effort       string // "" = 유지
	DeleteEffort bool   // "(absent)" 제출 — effort 키 삭제
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

// parseAgentFMForm은 agentfm.<agent>.model / agentfm.<agent>.effort 제출을
// 파싱한다. 빈 값은 preserve(EC-1), 검증 위반은 per-field 오류로 수집되어
// atomic reject에 합류한다. agent 이름은 뷰에 실존하는 파일 목록이 아니라
// 폼 이름에서 오므로 경로 조작 문자를 거부한다 (path traversal 가드).
func parseAgentFMForm(r *http.Request, agents []agentfm.AgentInfo) ([]agentFMEdit, map[string]string) {
	var edits []agentFMEdit
	errs := map[string]string{}

	for _, a := range agents {
		if !a.ParseOK {
			continue // 파싱 실패 행은 편집 비활성 (design.md §C.1 견고성)
		}
		if strings.ContainsAny(a.Name, "/\\") || strings.Contains(a.Name, "..") {
			continue // 방어적 — List가 만든 이름이라 정상 경로에선 불가능
		}
		edit := agentFMEdit{Agent: a.Name}

		if v := r.PostFormValue("agentfm." + a.Name + ".model"); v != "" {
			if !inList(agentFMModelValues(), v) {
				errs["agentfm."+a.Name+".model"] = "invalid option"
			} else if v != a.Model {
				if a.Model != "" || v != agentTierSuggestedModel(a.Name) {
					edit.Model = v
				}
			}
		}
		if v := r.PostFormValue("agentfm." + a.Name + ".effort"); v != "" {
			switch {
			case v == agentFMAbsent:
				if a.EffortPresent {
					edit.DeleteEffort = true
				}
			case !inList(agentFMEffortValues(), v):
				errs["agentfm."+a.Name+".effort"] = "invalid option"
			case v != a.Effort || !a.EffortPresent:
				if a.EffortPresent || v != agentTierSuggestedEffort(a.Name) {
					edit.Effort = v
				}
			}
		}

		if edit.Model != "" || edit.Effort != "" || edit.DeleteEffort {
			edits = append(edits, edit)
		}
	}
	return edits, errs
}

// applyAgentFMEdits는 편집을 frontmatter patch 레이어로 영속화한다 (live 파일
// 전용, no-change 편집은 파서가 이미 걸렀다). M5-a B6 — 다중 디렉터리(moai/ +
// harness/)에서 이름 → 절대 경로를 해상한다. 폼은 이름만 전달하므로, 패치 직전
// 라이브 디렉터리를 재스캔해 경로를 특정한다.
func applyAgentFMEdits(projectRoot string, edits []agentFMEdit) error {
	pathByName := map[string]string{}
	for _, dir := range agentDirsFor(projectRoot) {
		agents, err := agentfm.List(dir)
		if err != nil {
			return fmt.Errorf("agentfm: list %s: %w", dir, err)
		}
		for _, info := range agents {
			pathByName[info.Name] = info.Path
		}
	}
	for _, e := range edits {
		path, ok := pathByName[e.Agent]
		if !ok {
			continue // 방어적 — 폼과 라이브 리스트 불일치 (이미 삭제된 에이전트 등)
		}
		if err := agentfm.Patch(path, e.Model, e.Effort, e.DeleteEffort); err != nil {
			return err
		}
	}
	return nil
}
