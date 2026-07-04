package settings

// 이 파일은 SPEC-WEB-CONSOLE-011 M2b 10섹션 확장의 신규 FieldDef 정의를 담는다
// (REQ-WC11-010/011/012/016/017/019). 기존 34-필드(schema.go allFields)의 뒤에
// sectionExtraFields()가 append 된다.
//
// 영속화 라우팅 (design.md §A.3):
//   - git_strategy / llm / quality 확장 키 → PersistTypedSection (sectionapply.go
//     typed applier — git-strategy는 dirty-flag Save, REQ-WC11-010).
//   - 8개 seam 섹션(workflow, harness, ralph, research, feedback, observability,
//     security, db) → PersistSeam (WriteSectionViaSeam, REQ-WC11-017).
//
// read-only 지정 키(REQ-WC11-013/019 — llm.mode/team_mode, db system 5키)는
// 편집 FieldDef를 갖지 않는다 — ReadOnlyDisplayFields()가 표시 전용 메타를 제공.
// form UI에 맞지 않는 map/list 서브블록(REQ-WC11-062)은 RawViewBlocks()가 제공.

import "strings"

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

// gitStrategyProfileLeaves는 3개 mode profile(manual/personal/team)이 공유하는
// 스칼라 leaf 정의다: (key suffix, type).
var gitStrategyProfileLeaves = []struct {
	key string
	typ FieldType
}{
	{"workflow", TypeText},
	{"environment", TypeText},
	{"github_integration", TypeBool},
	{"push_to_remote", TypeBool},
	{"automation.auto_branch", TypeBool},
	{"automation.auto_commit", TypeBool},
	{"automation.auto_pr", TypeBool},
	{"automation.auto_push", TypeBool},
	{"branch_creation.auto_enabled", TypeBool},
	{"branch_creation.prompt_always", TypeBool},
	{"commit_style.format", TypeText},
	{"commit_style.scope_required", TypeBool},
	{"hooks.pre_commit", TypeText},
	{"hooks.pre_push", TypeText},
	{"hooks.commit_msg", TypeText},
}

func gitStrategyFields() []FieldDef {
	fields := []FieldDef{
		withSelect(typedField(SectionGitStrategy, "git_strategy", "mode", TypeSelect),
			"f.git_strategy.mode.opt.", []string{"manual", "personal", "team"}, "", ""),
		withSelect(typedField(SectionGitStrategy, "git_strategy", "provider", TypeSelect),
			"f.git_strategy.provider.opt.", []string{"github", "gitlab"}, "", ""),
		typedField(SectionGitStrategy, "git_strategy", "github_username", TypeText),
		typedField(SectionGitStrategy, "git_strategy", "gitlab.instance_url", TypeText),
	}
	for _, profile := range []string{"manual", "personal", "team"} {
		for _, leaf := range gitStrategyProfileLeaves {
			fields = append(fields, typedField(SectionGitStrategy, "git_strategy", profile+"."+leaf.key, leaf.typ))
		}
	}
	// mode-conditional 키 (ModeProfile 주석 계약과 일치).
	fields = append(fields,
		typedField(SectionGitStrategy, "git_strategy", "manual.auto_checkpoint", TypeText),
		typedField(SectionGitStrategy, "git_strategy", "personal.branch_prefix", TypeText),
		typedField(SectionGitStrategy, "git_strategy", "personal.main_branch", TypeText),
		typedField(SectionGitStrategy, "git_strategy", "team.branch_prefix", TypeText),
		typedField(SectionGitStrategy, "git_strategy", "team.main_branch", TypeText),
		typedField(SectionGitStrategy, "git_strategy", "team.draft_pr", TypeBool),
		typedField(SectionGitStrategy, "git_strategy", "team.required_reviews", TypeInt),
		typedField(SectionGitStrategy, "git_strategy", "team.branch_protection", TypeBool),
	)
	return fields
}

// ─── llm 안전 키 (REQ-WC11-012/013/014 — typed 경로) ─────────────────────────

func llmFields() []FieldDef {
	perf := withSelect(typedField(SectionLLM, "llm", "performance_tier", TypeSelect),
		"f.llm.performance_tier.opt.", []string{"high", "medium", "low"},
		emptyLabelRuntimeDefault, "opt.runtime_default")
	fields := []FieldDef{perf}
	// claude_models tiers — 빈 값은 "(runtime default)" 시맨틱 (EC-4: 빈 값 유지).
	for _, tier := range []string{"high", "medium", "low"} {
		f := typedField(SectionLLM, "llm", "claude_models."+tier, TypeText)
		f.EmptyLabel = emptyLabelRuntimeDefault
		f.EmptyLabelKey = "opt.runtime_default"
		fields = append(fields, f)
	}
	for _, tier := range []string{"high", "medium", "low", "opus", "sonnet", "haiku"} {
		fields = append(fields, typedField(SectionLLM, "llm", "glm.models."+tier, TypeText))
	}
	return fields
}

// ─── quality 잔여 키 (REQ-WC11-011 — 기존 typed 경로 확장) ───────────────────

func qualityExtraFields() []FieldDef {
	q := func(key string, typ FieldType) FieldDef {
		return typedField(SectionQualityExtras, "quality", key, typ)
	}
	return []FieldDef{
		q("coverage_threshold", TypeInt),
		q("ddd_settings.require_existing_tests", TypeBool),
		q("ddd_settings.characterization_tests", TypeBool),
		q("ddd_settings.behavior_snapshots", TypeBool),
		q("ddd_settings.max_transformation_size", TypeText),
		q("ddd_settings.preserve_before_improve", TypeBool),
		q("tdd_settings.red_green_refactor", TypeBool),
		q("tdd_settings.test_first_required", TypeBool),
		q("tdd_settings.mutation_testing_enabled", TypeBool),
		q("coverage_exemptions.enabled", TypeBool),
		q("coverage_exemptions.require_justification", TypeBool),
		q("coverage_exemptions.max_exempt_percentage", TypeInt),
		q("test_quality.specification_based", TypeBool),
		q("test_quality.meaningful_assertions", TypeBool),
		q("test_quality.avoid_implementation_coupling", TypeBool),
		q("test_quality.mutation_testing_enabled", TypeBool),
	}
}

// QualityExcludedKeyPrefixes는 웹 노출에서 명시적으로 제외한 quality.yaml 키
// prefix다 (AC-WC11-011의 "제외 목록 명시분"). cycle_type_routing은 문서화 블록,
// lsp_* / principles는 대형 중첩 정책 블록 — form UI 부적합.
func QualityExcludedKeyPrefixes() []string {
	return []string{
		"cycle_type_routing.",
		"lsp_quality_gates.",
		"lsp_integration.",
		"principles.",
	}
}

// ─── 8개 seam 섹션 스칼라 키 (REQ-WC11-016/017 — 2026-07-03 §C-7 실측 기반) ──

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

		// ralph.
		s(SectionRalph, "ralph", TypeBool, "ralph", "enabled"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "lint_as_instruction"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "warn_as_instruction"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "ast_grep", "enabled"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "ast_grep", "auto_fix"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "ast_grep", "quality_scan"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "ast_grep", "security_scan"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "loop", "auto_fix"),
		s(SectionRalph, "ralph", TypeInt, "ralph", "loop", "max_iterations"),
		s(SectionRalph, "ralph", TypeInt, "ralph", "loop", "cooldown_seconds"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "loop", "require_confirmation"),
		s(SectionRalph, "ralph", TypeInt, "ralph", "loop", "completion", "coverage_threshold"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "loop", "completion", "tests_pass"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "loop", "completion", "zero_errors"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "loop", "completion", "zero_warnings"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "lsp", "auto_start"),
		s(SectionRalph, "ralph", TypeBool, "ralph", "lsp", "graceful_degradation"),
		s(SectionRalph, "ralph", TypeInt, "ralph", "lsp", "poll_interval_ms"),
		s(SectionRalph, "ralph", TypeInt, "ralph", "lsp", "timeout_seconds"),

		// research.
		s(SectionResearch, "research", TypeBool, "research", "enabled"),
		s(SectionResearch, "research", TypeInt, "research", "active", "budget_cap_tokens"),
		s(SectionResearch, "research", TypeInt, "research", "active", "max_experiments"),
		s(SectionResearch, "research", TypeFloat, "research", "active", "pass_threshold"),
		s(SectionResearch, "research", TypeInt, "research", "active", "runs_per_experiment"),
		s(SectionResearch, "research", TypeFloat, "research", "active", "target_score"),
		s(SectionResearch, "research", TypeText, "research", "dashboard", "default_mode"),
		s(SectionResearch, "research", TypeBool, "research", "dashboard", "html_open_browser"),
		s(SectionResearch, "research", TypeBool, "research", "passive", "enabled"),
		s(SectionResearch, "research", TypeInt, "research", "passive", "correction_window_seconds"),
		s(SectionResearch, "research", TypeFloat, "research", "safety", "canary_regression_threshold"),
		s(SectionResearch, "research", TypeBool, "research", "safety", "worktree_isolation"),

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

		// db (인터뷰 입력 3키만 편집 — REQ-WC11-019; system 5키는 read-only 표시).
		s(SectionDB, "db", TypeText, "db", "orm"),
		s(SectionDB, "db", TypeText, "db", "multi_tenant"),
		s(SectionDB, "db", TypeText, "db", "migration_tool"),
	}
}

// sectionExtraFields는 M2b 확장 필드 전체를 렌더 순서(AllSections 순서와 일치)로
// 반환한다.
func sectionExtraFields() []FieldDef {
	var fields []FieldDef
	fields = append(fields, qualityExtraFields()...) // SectionQuality 렌더 그룹에 합류
	fields = append(fields, gitStrategyFields()...)
	fields = append(fields, llmFields()...)
	fields = append(fields, seamSectionFields()...)
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
// (runtime-managed 레이스, REQ-WC11-013) + db system 스칼라 3키 (REQ-WC11-019;
// 나머지 system 2키 auto_sync/migration_patterns는 블록이라 RawViewBlocks 소관).
func ReadOnlyDisplayFields() []ReadOnlyField {
	return []ReadOnlyField{
		{SectionLLM, "llm", "llm.mode", []string{"llm", "mode"}},
		{SectionLLM, "llm", "llm.team_mode", []string{"llm", "team_mode"}},
		{SectionDB, "db", "db.enabled", []string{"db", "enabled"}},
		{SectionDB, "db", "db.dir", []string{"db", "dir"}},
		{SectionDB, "db", "db.engine", []string{"db", "engine"}},
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
		{SectionDB, "db", "db.auto_sync", []string{"db", "auto_sync"}},
		{SectionDB, "db", "db.migration_patterns", []string{"db", "migration_patterns"}},
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
		SectionResearch,
		SectionFeedback,
		SectionObservability,
		SectionSecurity,
		SectionDB,
	}
}
