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
	"github.com/modu-ai/moai-adk/internal/settings"
)

// consoleTab 은 콘솔 탭 nav 의 한 탭을 기술한다 (M5-b D1). LabelKey 는 기존
// sec.<section>.title i18n 키를 재사용한다 — 신규 sec.* 키 없음. ID 는 탭과
// 패널을 매칭하는 data-tab/data-panel 식별자다.
type consoleTab struct {
	ID        string
	LabelKey  string // 기존 sec.<section>.title 재사용 (0 NEW sec.* keys)
	Baseline  string // data-i18n 미실행 시 표시될 영문 fallback
}

// consoleTabs 는 탭 nav 의 정렬된 탭 목록을 반환한다. 각 탭의 label 은 기존
// sec.<section>.title i18n 키를 재사용한다 (REQ: 0 NEW sec.* keys).
// 순서는 fieldset 렌더 순서와 일치한다.
func consoleTabs() []consoleTab {
	return []consoleTab{
		{ID: "identity", LabelKey: "sec.identity.title", Baseline: "Identity"},
		{ID: "language", LabelKey: "sec.language.title", Baseline: "Language"},
		{ID: "launch", LabelKey: "sec.launch.title", Baseline: "LLM"},
		{ID: "llm", LabelKey: "sec.llm.title", Baseline: "3rd Party LLM"},
		// workflow restored (Issue 3): the worktree auto-create toggle lives here.
		// Original ordering placed it after llm (pre-cca120c70).
		{ID: "workflow", LabelKey: "sec.workflow.title", Baseline: "Workflow"},
		// git-worktree (M1): the git_strategy section had FieldDefs but no render
		// meta, so mode + the three per-profile merge_method controls had no UI at
		// all. The tab id is deliberately NOT the section id — this panel mixes
		// git_strategy and workflow.worktree fields, so it carries its own tab.*
		// i18n namespace instead of reusing sec.<section>.*.
		{ID: "git-worktree", LabelKey: "tab.git-worktree.title", Baseline: "Git & Worktree"},
		{ID: "agentfm", LabelKey: "sec.agentfm.title", Baseline: "Agents"},
		{ID: "report", LabelKey: "sec.report.title", Baseline: "Report"},
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
}

// schemaSectionMetas는 제네릭 렌더 대상 패널의 표시 메타를 렌더 순서대로 반환한다.
func schemaSectionMetas() []schemaSectionMeta {
	return []schemaSectionMeta{
		{
			ID: settings.SectionLLM, PanelID: "llm", Icon: "rocket",
			TitleKey: "sec.llm.title", DescKey: "sec.llm.desc",
			Title: "3rd Party LLM", Desc: "GLM backend model tier mappings (high/medium/low/fable).",
			Fields: settings.SectionFields(settings.SectionLLM),
		},
		// workflow restored (Issue 3): reverse the cca120c70 reclassification for
		// workflow ONLY — the worktree auto-create toggle renders via this fieldset.
		{
			ID: settings.SectionWorkflow, PanelID: "workflow", Icon: "panel-bottom",
			TitleKey: "sec.workflow.title", DescKey: "sec.workflow.desc",
			Title: "Workflow", Desc: "Workflow execution mode and loop-prevention settings.",
			Fields: settings.SectionFields(settings.SectionWorkflow),
		},
		{
			ID: settings.SectionGitStrategy, PanelID: "git-worktree", Icon: "folder-git",
			TitleKey: "tab.git-worktree.title", DescKey: "tab.git-worktree.desc",
			Title: "Git & Worktree", Desc: "Git strategy mode, per-profile merge method, and worktree automation.",
			Fields: settings.SectionFields(settings.SectionGitStrategy),
		},
		{
			ID: settings.SectionReport, PanelID: "report", Icon: "panel-bottom",
			TitleKey: "sec.report.title", DescKey: "sec.report.desc",
			Title: "Report", Desc: "Output format for the HTML report skill (report.format: html+md or md).",
			Fields: settings.SectionFields(settings.SectionReport),
		},
	}
}

// schemaEditableField는 이 필드가 M2b 제네릭 폼 경로의 편집 대상인지 보고한다.
func schemaEditableField(f settings.FieldDef) bool {
	return f.Persist.Kind == settings.PersistSeam || f.Persist.Kind == settings.PersistTypedSection
}

// parseSchemaForm은 제출 폼에서 확장 필드 값을 스키마 주도로 파싱한다.
// empty=preserve(EC-1): 미제출/빈 값은 edits에 포함하지 않는다. bool은 hidden
// companion(name+"__present") 패턴으로 "unchecked → false"와 "미제출 → preserve"
// 를 구분한다 (기존 nested-config 선례). 타입/옵션 위반은 per-field 오류로
// 수집되어 atomic reject(EC-2)에 합류한다.
func parseSchemaForm(r *http.Request) (map[string]string, map[string]string) {
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
			raw := r.PostFormValue(f.Name)
			if raw == "" {
				continue
			}
			if f.Validate != nil && !f.Validate(raw) {
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
