package settings

// 이 파일은 SPEC-WEB-CONSOLE-011 M2b 10섹션 확장의 신규 FieldDef 정의를 담는다
// (REQ-WC11-010/011/012/016/017/019). 기존 34-필드(schema.go allFields)의 뒤에
// sectionExtraFields()가 append 된다.
//
// 영속화 라우팅 (design.md §A.3):
//   - git_strategy / llm / quality 확장 키 → PersistTypedSection (sectionapply.go
//     typed applier — git-strategy는 dirty-flag Save, REQ-WC11-010).
//   - 7개 seam 섹션(workflow, harness, ralph, research, feedback, observability,
//     security) → PersistSeam (WriteSectionViaSeam, REQ-WC11-017).
//
// read-only 지정 키(REQ-WC11-013 — llm.mode/team_mode)는 편집 FieldDef를 갖지
// 않는다 — ReadOnlyDisplayFields()가 표시 전용 메타를 제공. form UI에 맞지 않는
// map/list 서브블록(REQ-WC11-062)은 RawViewBlocks()가 제공.

import (
	"strings"

	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
)

// ─── v4manifest closed sets 재사용 (REQ-WC11-024/029/072 — 재선언 금지) ──────

// v4EffortValues는 v4manifest의 5-effort closed set이다 (exported 상수 재사용 —
// 옵션 목록 리터럴 재선언 금지).
func v4EffortValues() []string {
	return []string{
		v4manifest.EffortLow, v4manifest.EffortMedium, v4manifest.EffortHigh,
		v4manifest.EffortXhigh, v4manifest.EffortMax,
	}
}

// v4ModelValues는 v4manifest의 4-model-tier closed set이다.
func v4ModelValues() []string {
	return []string{
		v4manifest.ModelInherit, v4manifest.ModelHaiku,
		v4manifest.ModelSonnet, v4manifest.ModelOpus,
	}
}

// v4IsolationValues는 v4manifest의 2-isolation closed set이다.
func v4IsolationValues() []string {
	return []string{v4manifest.IsolationNone, v4manifest.IsolationWorktree}
}

// V4EffortValues / V4ModelValues는 웹 계층(agent frontmatter 편집 검증,
// REQ-WC11-029)이 소비하는 공개 접근자다.
func V4EffortValues() []string { return v4EffortValues() }
func V4ModelValues() []string  { return v4ModelValues() }

// RoleProfileNames는 workflow.yaml team.role_profiles의 7개 profile 키다
// (실측 — workflow.yaml role_profiles map).
func RoleProfileNames() []string {
	return []string{"analyst", "architect", "designer", "implementer", "researcher", "reviewer", "tester"}
}

// WorkflowAgentPurposes는 dynamic-workflows.md 7-purpose taxonomy 슬러그다
// ("Purpose-driven model+effort selection" 표 실측 — REQ-WC11-070). M5-a B1 이후
// 웹 렌더에서는 사용하지 않지만(숨김), canonical taxonomy 참조로 유지한다 —
// config.Workflow.WorkflowAgents 맵의 키 집합과 일치한다.
func WorkflowAgentPurposes() []string {
	return []string{
		"read-only-extract", "mechanical-transform", "synthesize",
		"research", "verify-judge", "implement", "design-architecture",
	}
}

// ─── 컴팩트 생성자 ───────────────────────────────────────────────────────────

// seamField는 seam 섹션의 스칼라 필드를 생성한다. Name은 Path의 dot-join이다
// (Path[0]가 섹션 파일의 최상위 yaml 키이므로 전 섹션에서 유일하다).
func seamField(sec SectionID, file string, typ FieldType, path ...string) FieldDef {
	name := strings.Join(path, ".")
	return FieldDef{
		Name:    name,
		Section: sec,
		Type:    typ,
		I18nKey: "f." + name,
		Persist: PersistTarget{Kind: PersistSeam, Section: file, Path: path},
	}
}

// typedField는 typed 섹션(git_strategy/llm/quality)의 필드를 생성한다.
// Name = "<section>.<key>" (기존 quality.* 네이밍 선례와 일치).
func typedField(sec SectionID, section, key string, typ FieldType) FieldDef {
	name := section + "." + key
	return FieldDef{
		Name:    name,
		Section: sec,
		Type:    typ,
		I18nKey: "f." + name,
		Persist: PersistTarget{Kind: PersistTypedSection, Section: section, Key: key},
	}
}

// withSelect는 필드를 닫힌 옵션 집합의 select로 전환한다 (멤버십 검증 포함).
func withSelect(f FieldDef, keyPrefix string, values []string, emptyLabel, emptyKey string) FieldDef {
	opts := optionDefsFromValues(keyPrefix, values)
	f.Type = TypeSelect
	f.Options = opts
	f.EmptyLabel = emptyLabel
	f.EmptyLabelKey = emptyKey
	f.Validate = func(v string) bool { return inOptionValues(opts, v) }
	return f
}

// ─── git-strategy (REQ-WC11-010 — typed dirty-flag Save) ────────────────────

// gitStrategyFields는 M4 다이어트 후 웹 편집 노출 대상 git-strategy 필드만
// 반환한다: mode (ActiveModeProfile 선택키) + 3개 profile의 hooks.pre_push
// (hook_pre_push.go:72 런타임 소비자 — skip/warn/enforce). 나머지 ~53개 검증
// 전용 키는 M4에서 제거되었다 (struct 멤버와 yaml 로드는 보존 — backward compat).
func gitStrategyFields() []FieldDef {
	fields := []FieldDef{
		withSelect(typedField(SectionGitStrategy, "git_strategy", "mode", TypeSelect),
			"f.git_strategy.mode.opt.", []string{"manual", "personal", "team"}, "", ""),
	}
	// hooks.pre_push 만 profile별 노출 — 런타임 reader (hook_pre_push.go:72).
	for _, profile := range []string{"manual", "personal", "team"} {
		fields = append(fields, typedField(SectionGitStrategy, "git_strategy", profile+".hooks.pre_push", TypeText))
	}
	return fields
}

// ─── llm 안전 키 (REQ-WC11-012/013/014 — typed 경로) ─────────────────────────

// llmFields는 M4 다이어트 후 GLM tier 매핑만 반환한다 (glm.go:195-197 런타임
// 소비자 — ANTHROPIC_DEFAULT_*_MODEL 환경변수). performance_tier와
// claude_models.*는 런타임 reader 없이 제거되었다 (struct 멤버와 yaml 로드 보존).
func llmFields() []FieldDef {
	var fields []FieldDef
	for _, tier := range []string{"high", "medium", "low", "opus", "sonnet", "haiku"} {
		fields = append(fields, typedField(SectionLLM, "llm", "glm.models."+tier, TypeText))
	}
	return fields
}

// ─── quality 잔여 키 (REQ-WC11-011 — 기존 typed 경로 확장) ───────────────────

// qualityExtraFields는 M4 다이어트 후 런타임 소비자가 있는 DDD 게이트 필드만
// 반환한다 (trust.go:740/748/756). 나머지 13개 dead/redundant 키는 제거되었다
// (struct 멤버와 yaml 로드 보존 — TestQualityKeyPartition이 명시 제외 prefix로
// 검증).
func qualityExtraFields() []FieldDef {
	q := func(key string, typ FieldType) FieldDef {
		return typedField(SectionQualityExtras, "quality", key, typ)
	}
	return []FieldDef{
		q("ddd_settings.characterization_tests", TypeBool),
		q("ddd_settings.behavior_snapshots", TypeBool),
		q("ddd_settings.preserve_before_improve", TypeBool),
	}
}

// QualityExcludedKeyPrefixes는 웹 노출에서 명시적으로 제외한 quality.yaml 키
// prefix다 (AC-WC11-011의 "제외 목록 명시분"). cycle_type_routing은 문서화 블록,
// lsp_* / principles는 대형 중첩 정책 블록 — form UI 부적합. M4 다이어트로 제거된
// dead/redundant leaf 키(coverage_threshold alias, 사용되지 않는 ddd/tdd leaf,
// coverage_exemptions/test_quality 서브트리 전체)도 명시 제외 — struct 멤버와
// yaml 로드는 보존되어 backward compat을 유지한다.
func QualityExcludedKeyPrefixes() []string {
	return []string{
		"cycle_type_routing.",
		"lsp_quality_gates.",
		"lsp_integration.",
		"principles.",
		// M4 다이어트 — 런타임 reader 없는 개별 leaf.
		"coverage_threshold",
		"ddd_settings.require_existing_tests",
		"ddd_settings.max_transformation_size",
		"tdd_settings.red_green_refactor",
		"tdd_settings.test_first_required",
		"tdd_settings.mutation_testing_enabled",
		// M4 다이어트 — 서브트리 전체 제거.
		"coverage_exemptions.",
		"test_quality.",
	}
}

// ─── 7개 seam 섹션 스칼라 키 (REQ-WC11-016/017 — 2026-07-03 §C-7 실측 기반) ──

func seamSectionFields() []FieldDef {
	s := seamField
	return []FieldDef{
		// workflow (파일: workflow.yaml, 최상위 키 workflow).
		s(SectionWorkflow, "workflow", TypeText, "workflow", "default_mode"),
		s(SectionWorkflow, "workflow", TypeText, "workflow", "execution_mode"),
		s(SectionWorkflow, "workflow", TypeBool, "workflow", "auto_clear", "enabled"),
		s(SectionWorkflow, "workflow", TypeBool, "workflow", "auto_clear", "after_plan"),
		s(SectionWorkflow, "workflow", TypeBool, "workflow", "auto_clear", "after_run"),
		s(SectionWorkflow, "workflow", TypeInt, "workflow", "auto_clear", "token_threshold"),
		s(SectionWorkflow, "workflow", TypeInt, "workflow", "agentic_loop", "max_iterations"),
		s(SectionWorkflow, "workflow", TypeBool, "workflow", "loop_prevention", "failure_pattern_detection"),
		s(SectionWorkflow, "workflow", TypeInt, "workflow", "loop_prevention", "max_iterations"),
		s(SectionWorkflow, "workflow", TypeInt, "workflow", "loop_prevention", "max_retries_per_operation"),
		s(SectionWorkflow, "workflow", TypeBool, "workflow", "team", "enabled"),
		s(SectionWorkflow, "workflow", TypeBool, "workflow", "team", "delegate_mode"),
		s(SectionWorkflow, "workflow", TypeInt, "workflow", "team", "max_teammates"),
		s(SectionWorkflow, "workflow", TypeText, "workflow", "team", "default_model"),
		s(SectionWorkflow, "workflow", TypeBool, "workflow", "team", "require_plan_approval"),
		s(SectionWorkflow, "workflow", TypeInt, "workflow", "token_budget", "plan"),
		s(SectionWorkflow, "workflow", TypeInt, "workflow", "token_budget", "run"),
		s(SectionWorkflow, "workflow", TypeInt, "workflow", "token_budget", "sync"),
		s(SectionWorkflow, "workflow", TypeBool, "workflow", "worktree", "auto_cleanup"),
		s(SectionWorkflow, "workflow", TypeBool, "workflow", "worktree", "auto_create"),
		s(SectionWorkflow, "workflow", TypeBool, "workflow", "worktree", "auto_merge"),
		s(SectionWorkflow, "workflow", TypeBool, "workflow", "worktree", "tmux_preferred"),

		// harness (파일: harness.yaml, 최상위 키 harness + learning).
		s(SectionHarness, "harness", TypeText, "harness", "default_profile"),
		s(SectionHarness, "harness", TypeText, "harness", "effort_mapping", "minimal"),
		s(SectionHarness, "harness", TypeText, "harness", "effort_mapping", "standard"),
		s(SectionHarness, "harness", TypeText, "harness", "effort_mapping", "thorough"),
		s(SectionHarness, "harness", TypeBool, "harness", "auto_detection", "enabled"),
		s(SectionHarness, "harness", TypeBool, "harness", "escalation", "enabled"),
		s(SectionHarness, "harness", TypeInt, "harness", "escalation", "max_escalations"),
		s(SectionHarness, "harness", TypeText, "harness", "evaluator", "memory_scope"),
		s(SectionHarness, "harness", TypeText, "harness", "mode_defaults", "cg"),
		s(SectionHarness, "harness", TypeText, "harness", "mode_defaults", "solo"),
		s(SectionHarness, "harness", TypeText, "harness", "mode_defaults", "team"),
		s(SectionHarness, "harness", TypeBool, "learning", "enabled"),
		s(SectionHarness, "harness", TypeBool, "learning", "auto_apply"),
		s(SectionHarness, "harness", TypeInt, "learning", "log_retention_days"),

		// ralph — M4 다이어트: 런타임 reader가 있는 2개만 잔류
		// (post_tool.go:431/444, engine.go:239). 나머지 17개(enabled/ast_grep/
		// loop/lsp)는 문서화 전용이고 Go runtime이 binding하지 않는다.
		s(SectionRalph, "ralph", TypeBool, "ralph", "lint_as_instruction"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "warn_as_instruction"),

		// feedback.
		s(SectionFeedback, "feedback", TypeText, "feedback", "repository"),

		// observability.
		s(SectionObservability, "observability", TypeBool, "observability", "enabled"),
		s(SectionObservability, "observability", TypeInt, "observability", "max_file_size_mb"),
		s(SectionObservability, "observability", TypeText, "observability", "report_dir"),
		s(SectionObservability, "observability", TypeInt, "observability", "retention_days"),
		s(SectionObservability, "observability", TypeText, "observability", "trace_dir"),
		s(SectionObservability, "observability", TypeText, "observability", "hook_metrics", "output_path"),
		s(SectionObservability, "observability", TypeInt, "observability", "hook_metrics", "slow_hook_threshold_ms"),

		// security (스칼라만 — 패턴 리스트는 raw view, REQ-WC11-062).
		s(SectionSecurity, "security", TypeBool, "security", "permission", "strict_mode"),
		s(SectionSecurity, "security", TypeBool, "security", "sandbox", "required"),
		s(SectionSecurity, "security", TypeText, "security", "sandbox", "docker_image"),
	}
}

// ─── M3: agent-settings 필드 (REQ-WC11-020..024, 070..073) ───────────────────

// agentSettingsFields는 agent-settings의 웹 렌더 표면을 반환한다:
// (b) team.role_profiles — 7 profiles × {model, effort, isolation, mode}.
//
//	effort는 seam이 opaque node로 패치한다 — RoleProfileEntry에 Effort 필드를
//	추가하지 않는다 (REQ-WC11-022/023, REQ-WEM-006 유지). effort는 Go 런타임이
//	행동적으로 바인딩하지 않는 선언적 힌트다 (M5-a B5 — "(Go 미독)").
//	mode는 permission_mode 옵션 집합을 재사용한다 (M5-a B4 — 자유 텍스트에서
//	검증 select로 승격; 새 mode 값 발명 금지).
//
// (d) workflow_agents — 웹 렌더에서 숨김 (M5-a B1). struct 필드와 yaml 키는
//	유지된다 (config.Workflow.WorkflowAgents, types.go 참조). 웹 폼을 통한
//	읽기/쓰기는 더 이상 발생하지 않는다 — dynamic-workflow JS 스크립트가
//	yaml 파일을 직접 읽는다 (dynamic-workflows.md §Config surface).
//
// 옵션은 v4manifest closed sets에서 파생한다 (REQ-WC11-024/072 — 재선언 금지).
func agentSettingsFields() []FieldDef {
	modelSel := func(f FieldDef) FieldDef {
		return withSelect(f, "f.v4.model.opt.", v4ModelValues(), "", "")
	}
	effortSel := func(f FieldDef) FieldDef {
		return withSelect(f, "f.v4.effort.opt.", v4EffortValues(), "", "")
	}
	// modeSel은 role_profiles.mode를 permission_mode 옵션 집합으로 검증한다
	// (M5-a B4). permissionModeOptions() 재사용 — 새 mode 값 발명 금지.
	modeSel := func(f FieldDef) FieldDef {
		f.Type = TypeSelect
		f.Options = permissionModeOptions()
		return f
	}

	var fields []FieldDef
	for _, p := range RoleProfileNames() {
		base := []string{"workflow", "team", "role_profiles", p}
		fields = append(fields,
			modelSel(seamField(SectionAgentSettings, "workflow", TypeSelect, append(base, "model")...)),
			effortSel(seamField(SectionAgentSettings, "workflow", TypeSelect, append(base, "effort")...)),
			withSelect(seamField(SectionAgentSettings, "workflow", TypeSelect, append(base, "isolation")...),
				"f.v4.isolation.opt.", v4IsolationValues(), "", ""),
			modeSel(seamField(SectionAgentSettings, "workflow", TypeSelect, append(base, "mode")...)),
		)
	}
	return fields
}

// sectionExtraFields는 M2b/M3 확장 필드 전체를 렌더 순서(AllSections 순서와
// 일치)로 반환한다.
func sectionExtraFields() []FieldDef {
	var fields []FieldDef
	fields = append(fields, qualityExtraFields()...) // SectionQuality 렌더 그룹에 합류
	fields = append(fields, gitStrategyFields()...)
	fields = append(fields, llmFields()...)
	fields = append(fields, seamSectionFields()...)
	fields = append(fields, agentSettingsFields()...)
	return fields
}

// ─── read-only 표시 + raw view 메타 (REQ-WC11-013/019/062) ───────────────────

// ReadOnlyField는 편집 불가·표시 전용 키의 메타다. 값 표시는 sectionvalues.go의
// 제네릭 리더가 담당하고, 쓰기 경로는 존재하지 않는다 (FieldDef 미등록).
type ReadOnlyField struct {
	Section SectionID // 렌더 섹션
	File    string    // 섹션 파일 base name
	Name    string    // 표시 식별자 (dot-path)
	Path    []string  // yaml 경로
}

// ReadOnlyDisplayFields는 read-only 표시 키를 반환한다: llm.mode / llm.team_mode
// (runtime-managed 레이스, REQ-WC11-013).
func ReadOnlyDisplayFields() []ReadOnlyField {
	return []ReadOnlyField{
		{SectionLLM, "llm", "llm.mode", []string{"llm", "mode"}},
		{SectionLLM, "llm", "llm.team_mode", []string{"llm", "team_mode"}},
	}
}

// RawBlockRef는 form UI에 맞지 않아 collapsed raw view로만 렌더하는 서브블록이다
// (REQ-WC11-062 — map-of-structs / 패턴 리스트).
type RawBlockRef struct {
	Section SectionID
	File    string
	Name    string // 표시 식별자
	Path    []string
}

// RawViewBlocks는 raw view 대상 서브블록을 반환한다.
func RawViewBlocks() []RawBlockRef {
	return []RawBlockRef{
		{SectionWorkflow, "workflow", "workflow.team.patterns", []string{"workflow", "team", "patterns"}},
		{SectionHarness, "harness", "harness.levels", []string{"harness", "levels"}},
		{SectionSecurity, "security", "security.extra_dangerous_bash_patterns", []string{"security", "extra_dangerous_bash_patterns"}},
	}
}

// SchemaSectionIDs는 M2b 제네릭 fieldset으로 렌더하는 확장 섹션을 렌더 순서대로
// 반환한다. 기존 6섹션(수제 fieldset)은 포함하지 않는다. quality 잔여 키는
// 전용 SectionQualityExtras 그룹으로 렌더한다 (기존 Project fieldset의 9-필드
// 구성과 충돌하지 않도록 분리 — REQ-WC11-011).
func SchemaSectionIDs() []SectionID {
	return []SectionID{
		SectionQualityExtras,
		SectionGitStrategy,
		SectionLLM,
		SectionWorkflow,
		SectionHarness,
		SectionRalph,
		SectionFeedback,
		SectionObservability,
		SectionSecurity,
		SectionAgentSettings,
	}
}
