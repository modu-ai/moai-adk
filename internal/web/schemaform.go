package web

// 이 파일은 SPEC-WEB-CONSOLE-011 M2b 10섹션 확장 필드의 웹 계층 배선이다:
// 스키마 주도 폼 파싱(제네릭 — 필드별 hand-wiring 없음, FieldDef가 SSOT) +
// 뷰모델 시딩 + 영속화 seam 연결. 영속화 자체는 settings.ApplySchemaEdits가
// 담당한다 (seam 8섹션 = yamlpatch, git-strategy/llm/quality = typed 경로).
//
// read-only 키(llm.mode/team_mode, db system 5키 — REQ-WC11-013/019)와 제외군
// 섹션(REQ-WC11-018)은 스키마에 편집 FieldDef가 없으므로 이 파서가 절대 집어
// 올리지 않는다 — 위조 제출은 무시되고 파일은 불변이다 (EC-8).

import (
	"net/http"
	"strconv"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// schemaSectionMeta는 제네릭 fieldset의 섹션 표시 메타다. Title/Desc는 영어
// baseline이고 data-i18n 키(sec.<id>.title/.desc)가 4-locale 렌더를 담당한다.
// 필드 자체의 라벨은 기술 식별자(key chip)로 렌더하며 번역하지 않는다 —
// i18n.js 헤더의 "Field identifiers stay in English" 계약과 동일.
type schemaSectionMeta struct {
	ID    settings.SectionID
	Icon  string
	Title string
	Desc  string
}

// schemaSectionMetas는 제네릭 렌더 대상 섹션의 표시 메타를 렌더 순서대로
// 반환한다 (settings.SchemaSectionIDs와 동순).
func schemaSectionMetas() []schemaSectionMeta {
	return []schemaSectionMeta{
		{settings.SectionQualityExtras, "check-circle", "Quality (advanced)", "Remaining quality.yaml keys — DDD/TDD settings, exemptions, and test quality."},
		{settings.SectionGitStrategy, "folder-git", "Git Strategy", "Git workflow strategy — mode, provider, and per-mode automation profiles."},
		{settings.SectionLLM, "rocket", "LLM", "Model tier mappings for Claude and GLM backends."},
		{settings.SectionWorkflow, "panel-bottom", "Workflow", "Workflow execution, auto-clear, team, token budget, and worktree settings."},
		{settings.SectionHarness, "check-circle", "Harness", "Quality harness levels, escalation, and learning subsystem."},
		{settings.SectionRalph, "rocket", "Ralph", "Ralph feedback loop — LSP, AST-grep, and loop completion settings."},
		{settings.SectionResearch, "languages", "Research", "Self-harness research subsystem settings."},
		{settings.SectionFeedback, "alert-circle", "Feedback", "Feedback workflow target repository."},
		{settings.SectionObservability, "panel-bottom", "Observability", "Trace, report, and hook metrics settings."},
		{settings.SectionSecurity, "check-circle", "Security", "Permission strictness and sandbox settings (pattern lists are read-only)."},
		{settings.SectionDB, "folder-git", "Database", "Database documentation settings — interview keys editable, system keys read-only."},
		{settings.SectionAgentSettings, "user-round", "Agent Settings", "Team role profiles and workflow-agent purpose defaults (workflow.yaml — comment-preserving writes)."},
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

	// M3: sub-agent frontmatter 현재 상태 시딩 (REQ-WC11-020/025). 목록 실패는
	// 빈 목록으로 저하 — 페이지 전체 실패 금지 (design.md §C.1 견고성).
	if agents, err := a.listAgentFMs(agentsDirFor(a.cfg.ProjectRoot)); err == nil {
		view.AgentFMs = agents
	}
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
