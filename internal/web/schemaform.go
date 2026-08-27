package web

// 이 파일은 SPEC-WEB-CONSOLE-011 M2b 10섹션 확장 필드의 웹 계층 배선이다:
// 스키마 주도 폼 파싱(제네릭 — 필드별 hand-wiring 없음, FieldDef가 SSOT) +
// 뷰모델 시딩 + 영속화 seam 연결. 영속화 자체는 settings.ApplySchemaEdits가
// 담당한다 (seam 8섹션 = yamlpatch, git-strategy/llm/quality = typed 경로).
//
// read-only 키(llm.mode/team_mode — REQ-WC11-013)와 제외군 섹션(REQ-WC11-018)은
// 스키마에 편집 FieldDef가 없으므로 이 파서가 절대 집어 올리지 않는다 — 위조
// 제출은 무시되고 파일은 불변이다 (EC-8).

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// consoleTab 은 콘솔 탭 nav 의 한 탭을 기술한다 (M5-b D1). LabelKey 는 기존
// sec.<section>.title i18n 키를 재사용한다 — 신규 sec.* 키 없음. ID 는 탭과
// 패널을 매칭하는 data-tab/data-panel 식별자다.
type consoleTab struct {
	ID       string
	LabelKey string // 기존 sec.<section>.title 재사용 (0 NEW sec.* keys)
	Baseline string // data-i18n 미실행 시 표시될 영문 fallback
}

// consoleTabs 는 탭 nav 의 정렬된 탭 목록을 반환한다. 각 탭의 label 은 기존
// sec.<section>.title i18n 키를 재사용한다 (REQ: 0 NEW sec.* keys).
// 순서는 fieldset 렌더 순서와 일치한다.
func consoleTabs() []consoleTab {
	return []consoleTab{
		{ID: "identity", LabelKey: "sec.identity.title", Baseline: "Identity"},
		{ID: "language", LabelKey: "sec.language.title", Baseline: "Language"},
		{ID: "launch", LabelKey: "sec.launch.title", Baseline: "LLM"},
		{ID: "llm", LabelKey: "sec.llm.title", Baseline: "GLM Settings"},
		// workflow restored (Issue 3): the worktree auto-create toggle lives here.
		// Original ordering placed it after llm (pre-cca120c70).
		{ID: "workflow", LabelKey: "sec.workflow.title", Baseline: "Workflow"},
		// git-worktree (M1): the git_strategy section had FieldDefs but no render
		// meta, so mode + the three per-profile merge_method controls had no UI at
		// all. The tab id is deliberately NOT the section id — this panel mixes
		// git_strategy and workflow.worktree fields, so it carries its own tab.*
		// i18n namespace instead of reusing sec.<section>.*.
		{ID: "git-worktree", LabelKey: "tab.git-worktree.title", Baseline: "Git & Worktree"},
		// audit (M2): workflow.audit.* moved off the workflow tab onto its own.
		// The move is a RENDER placement only — the fields keep SectionWorkflow
		// and the workflow.yaml seam persist target (AP-4).
		{ID: "audit", LabelKey: "tab.audit.title", Baseline: "Audit"},
		{ID: "agentfm", LabelKey: "sec.agentfm.title", Baseline: "Agents"},
		{ID: "report", LabelKey: "sec.report.title", Baseline: "Report"},
		// SPEC-MCP-CONSOLE-001 M2: the per-tool MCP enablement panel. Each of the
		// 17 tools renders as an individually-toggleable bool; the 4 write-capable
		// tools carry a text-bearing distinction (REQ-C-3). Rendered by the
		// dedicated fieldsetMCP component (not the generic schemaSectionMeta path)
		// so the write-capable badge can be sourced from the shared catalog.
		{ID: "mcp", LabelKey: "sec.mcp.title", Baseline: "MCP"},
		// crosssession — the cross-session messaging posture panel (inbound /
		// isolate_machines / dialog_expiry). The launchers translate
		// crosssession.yaml into a session --settings injection; this panel
		// edits the same file through the yamlpatch seam.
		{ID: "crosssession", LabelKey: "sec.crosssession.title", Baseline: "Cross-Session"},
		// feedback — reopened by SPEC-FEEDBACK-AUTO-SUBMIT-001 M7, reversing the
		// SPEC-WEBCONF-SIMPLIFY-001 M3 reclassification for this section only.
		// The panel carries the target repository and the auto-submit consent
		// toggle; both persist through the yamlpatch seam into feedback.yaml.
		{ID: "feedback", LabelKey: "sec.feedback.title", Baseline: "Feedback"},
	}
}

// mcpToolNameFromField extracts the tool identifier from an MCP enablement
// field name of the form mcp.tools.<name>.enabled. It returns ("", false) for
// a field that does not match that shape, so a caller cannot mistake an
// unrelated field for an MCP tool (SPEC-MCP-CONSOLE-001 M2, REQ-C-3).
func mcpToolNameFromField(fieldName string) (string, bool) {
	const prefix = "mcp.tools."
	const suffix = ".enabled"
	if !strings.HasPrefix(fieldName, prefix) || !strings.HasSuffix(fieldName, suffix) {
		return "", false
	}
	name := strings.TrimSuffix(strings.TrimPrefix(fieldName, prefix), suffix)
	if name == "" {
		return "", false
	}
	return name, true
}

// mcpToolIsWriteCapable reports whether the MCP enablement field's tool is
// write-capable, by looking the tool name up in the shared catalog
// (internal/mcp MoaiMCPTools — the single declaration, AP-C-4). A field whose
// tool name is not in the catalog returns false; the schema/catalog parity is
// pinned separately by the settings-layer tests, so a stale field here degrades
// to read-only marking rather than panicking the render.
func mcpToolIsWriteCapable(fieldName string) bool {
	name, ok := mcpToolNameFromField(fieldName)
	if !ok {
		return false
	}
	for _, t := range mcpcat.MoaiMCPTools() {
		if t.Name == name {
			return t.WriteCapable
		}
	}
	return false
}

// codexAuthProviderLabel maps the auth-provider token the probe emitted to a
// display string. It is NOT a classification — the token was already produced
// by the probe's classifier (AC-C-006); this helper only chooses the display
// spelling.
func codexAuthProviderLabel(provider string) string {
	switch provider {
	case codexAuthChatGPT:
		return "ChatGPT"
	case codexAuthAPIKey:
		return "API key"
	case codexAuthProvider:
		return "Provider"
	default:
		return provider // "unknown" or any future token renders verbatim
	}
}

// schemaSectionMeta는 제네릭 fieldset의 패널 표시 메타다. Title/Desc는 영어
// baseline이고 TitleKey/DescKey의 data-i18n 키가 4-locale 렌더를 담당한다.
// 필드 자체의 라벨은 기술 식별자(key chip)로 렌더하며 번역하지 않는다 —
// i18n.js 헤더의 "Field identifiers stay in English" 계약과 동일.
//
// PanelID는 data-tab/data-panel 식별자이고 ID는 영속화 섹션이다. 두 값은 대개
// 같지만 한 패널이 여러 섹션의 필드를 섞으면 갈라진다 (git-worktree). Fields는
// 그 패널이 렌더할 필드 목록이며, 섹션 전체가 아니라 패널 단위로 명시된다 —
// 탭 배치는 렌더 관심사이고 섹션은 영속화 단위다 (섹션 재분류로 탭을 옮기면
// 영속화 경로가 바뀐다).
type schemaSectionMeta struct {
	ID       settings.SectionID
	PanelID  string
	Icon     string
	TitleKey string
	DescKey  string
	Title    string
	Desc     string
	Fields   []settings.FieldDef
	// Extras가 true인 패널만 ID 섹션의 read-only 표시 키와 raw view 블록을
	// 렌더한다. 한 섹션이 여러 패널로 갈라질 때(workflow → 워크플로우/감사) 그
	// 부수 표면이 중복 렌더되는 것을 막는 primary-panel 표식이다.
	Extras bool
	// NoteKey/Note는 패널 헤더에 1회 렌더되는 주의 문구다 (빈 값이면 미렌더).
	// 필드마다 반복되는 힌트를 헤더로 승격하는 기존 관례(agentfm-gridnote)와
	// 동일한 자리다 — 한 패널의 모든 필드에 공통으로 걸리는 사실은 필드마다
	// 되풀이하지 않는다.
	NoteKey string
	Note    string
}

// isWorktreeFieldName은 workflow 섹션 필드 중 Git·워크트리 탭으로 배치되는 것을
// 판정한다. 이름 기반 분류이지 섹션 재분류가 아니다 — 영속화 경로는 그대로다.
func isWorktreeFieldName(name string) bool {
	return strings.HasPrefix(name, "workflow.worktree.") || name == "workflow.branch_guard.enabled"
}

// isAuditFieldName은 workflow 섹션 필드 중 감사 탭으로 배치되는 것을 판정한다.
func isAuditFieldName(name string) bool {
	return strings.HasPrefix(name, "workflow.audit.")
}

// isCodexToggleFieldName은 workflow 섹션 필드 중 MCP 콘솔의 codex 인증 서피스로
// 배치되는 것을 판정한다 (SPEC-MCP-CONSOLE-001 M3). 이 필드들은 workflow 탭이
// 아닌 MCP 탭의 codexAuthBlock 에서 렌더되므로 workflow 파티션에서 제외한다 —
// 중복 렌더(입력 4개)를 방지한다. 영속화 경로는 그대로다 (SectionWorkflow seam).
func isCodexToggleFieldName(name string) bool {
	return name == "workflow.codex.review_gate.enabled" ||
		name == "workflow.codex.task.allow_write"
}

// partitionWorkflowFields는 workflow 섹션 필드를 3개 탭으로 가른다: 워크플로우
// 잔여 / Git·워크트리 / 감사. codex 토글 필드는 MCP 탭에서 렌더되므로 어느
// workflow 탭에도 배치하지 않는다. 섹션 필드 순서를 보존한다.
func partitionWorkflowFields() (rest, worktree, audit []settings.FieldDef) {
	for _, f := range settings.SectionFields(settings.SectionWorkflow) {
		if isCodexToggleFieldName(f.Name) {
			continue // MCP 탭의 codexAuthBlock 에서 렌더 — workflow 탭 제외
		}
		switch {
		case isWorktreeFieldName(f.Name):
			worktree = append(worktree, f)
		case isAuditFieldName(f.Name):
			audit = append(audit, f)
		default:
			rest = append(rest, f)
		}
	}
	return rest, worktree, audit
}

// schemaSectionMetas는 제네릭 렌더 대상 패널의 표시 메타를 렌더 순서대로 반환한다.
func schemaSectionMetas() []schemaSectionMeta {
	workflowRest, worktreeFields, auditFields := partitionWorkflowFields()
	return []schemaSectionMeta{
		{
			ID: settings.SectionLLM, PanelID: "llm", Icon: "rocket",
			TitleKey: "sec.llm.title", DescKey: "sec.llm.desc",
			Title: "GLM Settings", Desc: "GLM backend model tier mappings and per-tier reasoning effort.",
			Fields: settings.SectionFields(settings.SectionLLM), Extras: true,
			// REQ-WCR-033: the honesty note. The four per-tier effort values are
			// stored and never applied — the runtime reads one session-global
			// ANTHROPIC_REASONING_EFFORT derived from the session effort_level
			// preference. Rendered once at the panel header, not per field.
			NoteKey: "sec.llm.effortnote",
			Note:    "Reasoning effort applied at runtime comes from the session-wide effort_level preference on the LLM tab, not from these tiers. The four per-tier values below are stored only.",
		},
		// workflow restored (Issue 3): reverse the cca120c70 reclassification for
		// workflow ONLY — the worktree auto-create toggle renders via this fieldset.
		{
			ID: settings.SectionWorkflow, PanelID: "workflow", Icon: "panel-bottom",
			TitleKey: "sec.workflow.title", DescKey: "sec.workflow.desc",
			Title: "Workflow", Desc: "Workflow execution mode and loop-prevention settings.",
			Fields: workflowRest, Extras: true,
		},
		{
			ID: settings.SectionGitStrategy, PanelID: "git-worktree", Icon: "folder-git",
			TitleKey: "tab.git-worktree.title", DescKey: "tab.git-worktree.desc",
			Title: "Git & Worktree", Desc: "Git strategy mode, per-profile merge method, and worktree automation.",
			Fields: append(settings.SectionFields(settings.SectionGitStrategy), worktreeFields...), Extras: true,
		},
		{
			// The audit panel's persistence section stays SectionWorkflow: the
			// tab is a render placement, the section is the write route (AP-4).
			ID: settings.SectionWorkflow, PanelID: "audit", Icon: "check-circle",
			TitleKey: "tab.audit.title", DescKey: "tab.audit.desc",
			Title: "Audit", Desc: "Review backend that gates merges and the per-auditor gate strictness.",
			Fields: auditFields,
		},
		{
			ID: settings.SectionReport, PanelID: "report", Icon: "panel-bottom",
			TitleKey: "sec.report.title", DescKey: "sec.report.desc",
			Title: "Report", Desc: "Output format for the HTML report skill (report.format: html+md or md).",
			Fields: settings.SectionFields(settings.SectionReport), Extras: true,
		},
		{
			ID: settings.SectionCrossSession, PanelID: "crosssession", Icon: "messages-square",
			TitleKey: "sec.crosssession.title", DescKey: "sec.crosssession.desc",
			Title: "Cross-Session", Desc: "How sessions launched by moai treat messages from your other Claude Code sessions.",
			Fields: settings.SectionFields(settings.SectionCrossSession), Extras: true,
			// The honesty note: these settings take effect at the NEXT launch —
			// the launchers read crosssession.yaml when moai cc/glm/cg starts a
			// session and inject it as that session's --settings. A running
			// session keeps the posture it was launched with.
			NoteKey: "sec.crosssession.note",
			Note:    "Applied at the next moai cc/glm/cg launch — running sessions keep the posture they were launched with.",
		},
		{
			ID: settings.SectionFeedback, PanelID: "feedback", Icon: "messages-square",
			TitleKey: "sec.feedback.title", DescKey: "sec.feedback.desc",
			Title: "Feedback", Desc: "Target repository for the feedback workflow, and whether it may submit without asking each time.",
			Fields: settings.SectionFields(settings.SectionFeedback), Extras: true,
		},
	}
}

// schemaPanelMeta는 패널 ID로 렌더 메타를 조회한다. 탭 목록이 순서의 단일
// 원천이므로 root.templ은 탭을 순회하며 이 함수로 메타를 끌어온다. 매칭이 없는
// ID는 빈 메타를 반환한다 — 필드 0개의 빈 fieldset이 렌더되며, 그 자체가 탭↔패널
// 배선 누락을 눈에 보이게 만든다 (TestEveryTabHasAPanel이 이를 고정한다).
func schemaPanelMeta(panelID string) schemaSectionMeta {
	for _, m := range schemaSectionMetas() {
		if m.PanelID == panelID {
			return m
		}
	}
	return schemaSectionMeta{PanelID: panelID}
}

// schemaEditableField는 이 필드가 M2b 제네릭 폼 경로의 편집 대상인지 보고한다.
func schemaEditableField(f settings.FieldDef) bool {
	return f.Persist.Kind == settings.PersistSeam || f.Persist.Kind == settings.PersistTypedSection
}

// parseSchemaForm은 제출 폼에서 확장 필드 값을 스키마 주도로 파싱한다.
// empty=preserve(EC-1): 미제출/빈 값은 edits에 포함하지 않는다 — 단 EmptySubmits
// 옵트인 필드(crosssession.inbound/dialog_expiry)의 ""는 실제 제출 값으로 취급해
// 키를 중립 ""로 되돌린다. bool은 hidden
// companion(name+"__present") 패턴으로 "unchecked → false"와 "미제출 → preserve"
// 를 구분한다 (기존 nested-config 선례). 타입/옵션 위반은 per-field 오류로
// 수집되어 atomic reject(EC-2)에 합류한다.
//
// current는 필드별 디스크 현재값이다 (RC2 passthrough-preserve,
// glm-settings-persist): 닫힌 옵션 집합 검증은 "이전에 영속된 값과 동일한
// 제출"을 통과시킨다 — 렌더러가 옵션 밖 영속값을 synthetic option으로
// round-trip시키므로, 그 제출이 거부되면 콘솔은 그 필드를 다시 저장할 수
// 없다. 전혀 새로운 집합 밖 제출은 여전히 거부된다. nil이면 옵트인 없는
// 종전 동작(엄격한 닫힌 집합)이다.
func parseSchemaForm(r *http.Request, current map[string]string) (map[string]string, map[string]string) {
	edits := map[string]string{}
	errs := map[string]string{}

	for _, f := range settings.AllFields() {
		if !schemaEditableField(f) {
			continue
		}
		switch f.Type {
		case settings.TypeBool:
			if r.PostFormValue(f.Name+"__present") == "" {
				continue // 미제출 → preserve
			}
			if r.PostFormValue(f.Name) != "" {
				edits[f.Name] = "true"
			} else {
				edits[f.Name] = "false"
			}
		case settings.TypeInt:
			raw := r.PostFormValue(f.Name)
			if raw == "" {
				continue
			}
			if _, err := strconv.Atoi(raw); err != nil {
				errs[f.Name] = "must be an integer"
				continue
			}
			edits[f.Name] = raw
		case settings.TypeFloat:
			raw := r.PostFormValue(f.Name)
			if raw == "" {
				continue
			}
			if _, err := strconv.ParseFloat(raw, 64); err != nil {
				errs[f.Name] = "must be a number"
				continue
			}
			edits[f.Name] = raw
		default: // TypeText / TypeSelect
			// 미제출 → preserve (EC-1). 키 존재로 "미제출"과 "빈 값 제출"을
			// 구분한다 — 브라우저 폼은 select를 항상 제출하므로 제출된 ""는
			// 사용자의 선택이고, 키 없는 폼(비브라우저 게시)은 보존이다.
			vals, submitted := r.PostForm[f.Name]
			if !submitted {
				continue
			}
			raw := ""
			if len(vals) > 0 {
				raw = vals[0]
			}
			// empty=preserve(EC-1) — 단, EmptySubmits 옵트인 필드는 ""가 실제
			// 제출 값이다 (선택 해제 → 키를 중립 ""로 되돌린다).
			if raw == "" && !f.EmptySubmits {
				continue
			}
			if f.Validate != nil && !f.Validate(raw) && raw != current[f.Name] {
				// RC2 passthrough-preserve: out-of-set BUT equal to the previously
				// persisted value → accept (the synthetic-option round-trip submits
				// exactly this). Genuinely new out-of-set submissions still reject.
				errs[f.Name] = "invalid option"
				continue
			}
			edits[f.Name] = raw
		}
	}
	return edits, errs
}

// applySchemaCurrent는 확장 필드 + read-only 표시 키 + raw view 블록의 디스크
// 현재 값을 뷰모델에 시드한다.
func (a *app) applySchemaCurrent(view *pageView) error {
	values, err := a.schemaCurrentValues(a.cfg.ProjectRoot)
	if err != nil {
		return err
	}
	blocks, err := a.rawBlockValues(a.cfg.ProjectRoot)
	if err != nil {
		return err
	}
	view.SchemaValues = values
	view.RawBlocks = blocks

	// goal-to-test (non-SPEC): seed the profile selector (hosted as the
	// performance_tier wire field) rendered at the top of the agentfm panel.
	// llm.yaml is read directly — this field is deliberately NOT part of the
	// generic schema. The plan_type display was removed
	// (SPEC-MODEL-PROFILE-MATRIX-001 REQ-MPM-019); the save persists to
	// llm.profile only and no longer mutates agent frontmatter (REQ-MPM-040).
	cfg, err := config.NewConfigManager().LoadRaw(a.cfg.ProjectRoot)
	if err != nil {
		return err
	}
	// Prefer the new llm.profile; fall back to the legacy performance_tier alias.
	activeProfile := cfg.LLM.EffectiveProfile()
	view.PerfTier = activeProfile
	view.PerfTierIsEmpty = strings.TrimSpace(cfg.LLM.Profile) == "" && strings.TrimSpace(cfg.LLM.PerformanceTier) == ""

	// M3: sub-agent frontmatter 현재 상태 시딩 (REQ-WC11-020/025). 목록 실패는
	// 빈 목록으로 저하 — 페이지 전체 실패 금지 (design.md §C.1 견고성). 정렬은
	// profile-matrix-resolved model/effort를 기준으로 하므로 cfg.LLM 을 같이 넘긴다.
	if agents, err := a.listAllAgentFMs(a.cfg.ProjectRoot, cfg.LLM); err == nil {
		view.AgentFMs = agents
	}

	// G3-1/G3-4: seed the loaded LLM config so the agentfm rows resolve each
	// agent's model/effort through the profile matrix, and preselect the Custom
	// pseudo-tier when any per-agent override is present.
	view.LLM = cfg.LLM
	view.PerfTierCustom = len(cfg.LLM.AgentOverrides) > 0
	return nil
}

// applySchemaCurrentBestEffort는 POST 재렌더 경로용 — 읽기 실패 시 빈 맵으로
// 우아하게 저하한다 (저장 자체의 성패 메시지가 이미 배너로 표시됨).
func (a *app) applySchemaCurrentBestEffort(view *pageView) {
	if err := a.applySchemaCurrent(view); err != nil {
		view.SchemaValues = map[string]string{}
		view.RawBlocks = map[string]string{}
	}
}

// overlaySchemaEdits는 rejected POST 재렌더에서 제출값을 현재값 위에 에코한다
// (EC-2 — 제출값 유지 렌더).
func overlaySchemaEdits(view *pageView, edits map[string]string) {
	if view.SchemaValues == nil {
		view.SchemaValues = map[string]string{}
	}
	for name, v := range edits {
		view.SchemaValues[name] = v
	}
}
