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
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
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

// V4EffortValues / V4ModelValues는 웹 계층(agent frontmatter 편집 검증,
// REQ-WC11-029)이 소비하는 공개 접근자다.
func V4EffortValues() []string { return v4EffortValues() }
func V4ModelValues() []string  { return v4ModelValues() }

// ─── Sub-agent tier accessors (SPEC-WEBCONF-SIMPLIFY-001 M1) ──────────────────
//
// 티어(tier)는 서브에이전트의 4색 추론-역할 분류다. 각 에이전트의 effort
// frontmatter와는 독립적인 display-only 분류며 (design.md §B), canonical 테이블/
// 접근자는 internal/harness/v4manifest에 있다. 여기선 웹 계층이 effort/model
// closed-set 접근자(V4EffortValues/V4ModelValues)를 이미 임포트한 같은 패키지에서
// 티어 표면도 함께 노출한다.

// TierForAgent returns the display-only Tier for the named agent (keyed by
// agent file stem, matching agentfm.AgentInfo.Name). The second return is
// false when the name has no tier entry (a future agent not yet added to the
// table — design.md EC-6).
func TierForAgent(name string) (v4manifest.Tier, bool) {
	return v4manifest.AgentTier(name)
}

// TierSuggestedModelEffort returns the suggested (model, effort) pair for the
// tier (design.md §D). Applied only on explicit user action; writes via
// agentfm.Patch, NOT a new tier: frontmatter key (C-7).
func TierSuggestedModelEffort(t v4manifest.Tier) (model, effort string) {
	return v4manifest.TierSuggestedModelEffort(t)
}

// NOTE: workflow.yaml team.role_profiles 키 목록을 반환하던 접근자와 그 isolation
// closed-set 헬퍼는 Agent Teams 정적 레이어와 함께 제거되었다 — 웹 콘솔은 더 이상
// Agent Teams 설정을 렌더하지 않는다 (SPEC-AGENT-TEAM-RETIRE-001).

// NOTE: 7-purpose taxonomy 슬러그를 반환하던 zero-caller 접근자는
// SPEC-WEB-CONSOLE-012 M4(REQ-WC12-032)에서 제거되었다 — 전 리포 호출자 0 실측.
// taxonomy의 canonical SSOT는 dynamic-workflows.md 표 +
// config.Workflow.WorkflowAgents 맵 키다.

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

// withRadio는 필드를 닫힌 옵션 집합의 라디오 버튼 그룹으로 전환한다 (멤버십 검증 포함).
func withRadio(f FieldDef, keyPrefix string, values []string, emptyLabel, emptyKey string) FieldDef {
	opts := optionDefsFromValues(keyPrefix, values)
	f.Type = TypeRadio
	f.Options = opts
	f.EmptyLabel = emptyLabel
	f.EmptyLabelKey = emptyKey
	f.Validate = func(v string) bool { return inOptionValues(opts, v) }
	return f
}

// ─── git-strategy (REQ-WC11-010 — typed dirty-flag Save) ────────────────────

// gitStrategyFields는 웹 편집 노출 대상 git-strategy 필드를 반환한다: mode
// (ActiveModeProfile 선택키) + 3개 profile의 merge_method (SPEC-WEB-CONSOLE-014
// M3 — SPEC-MERGE-METHOD-CONFIG-001 REQ-MMC-007 agent-prose consumer: sync
// delivery/manager-git가 active mode의 값으로 gh pr merge 플래그 선택). 모두
// 라디오 버튼(withRadio)으로 렌더링한다. hooks.pre_push는 웹 편집면에서
// 제거되었다 — yaml 값과 hook_pre_push.go 런타임 소비자는 보존(기본값 사용).
// 나머지 ~53개 검증 전용 키는 M4 다이어트에서 제거되었다 (struct 멤버와
// yaml 로드는 보존 — backward compat).
func gitStrategyFields() []FieldDef {
	modeField := withRadio(typedField(SectionGitStrategy, "git_strategy", "mode", TypeRadio),
		"f.git_strategy.mode.opt.", []string{"manual", "personal", "team"}, "", "")
	modeField.Description = "fieldDesc.git_strategy.mode"
	fields := []FieldDef{modeField}
	// merge_method 옵션은 config.ValidMergeMethods() SSOT에서 정렬 파생한다
	// (REQ-WC14-011 — 리터럴 재선언 금지; B3 — map-range 비결정성 제거를 위해 정렬
	// 후 사용. 정렬은 파생이지 재선언이 아님). 3개 profile 공유.
	mergeMethods := append([]string{}, config.ValidMergeMethods()...)
	sort.Strings(mergeMethods)
	for _, profile := range []string{"manual", "personal", "team"} {
		mergeMethod := withRadio(
			typedField(SectionGitStrategy, "git_strategy", profile+".merge_method", TypeRadio),
			"f.git_strategy.merge_method.opt.", mergeMethods, emptyLabelProjectDefault, "opt.project_default")
		mergeMethod.Description = "fieldDesc.git_strategy.merge_method"
		fields = append(fields, mergeMethod)
	}
	return fields
}

// ─── llm 안전 키 (REQ-WC11-012/013/014 — typed 경로) ─────────────────────────

// llmFields는 실소비 GLM tier 매핑 4종(high/medium/low/fable)만 반환한다
// (glm.go setGLMEnv 런타임 소비자 — ANTHROPIC_DEFAULT_*_MODEL 환경변수;
// ANTHROPIC_DEFAULT_OPUS_MODEL조차 Models.High에서 온다). legacy alias
// opus/sonnet/haiku는 SPEC-WEB-CONSOLE-012 REQ-WC12-002에서 웹 편집면에서
// 제거되었다 — resolveGLMModels empty-fallback 체인과 GLMModels legacy struct
// 멤버는 무접촉 보존되어 legacy yaml 로드가 backward-compat를 유지한다
// (REQ-WC12-006). performance_tier와 claude_models.*는 M4 다이어트에서 제거
// (struct 멤버와 yaml 로드 보존).
func llmFields() []FieldDef {
	var fields []FieldDef
	for _, tier := range []string{"high", "medium", "low", "fable"} {
		f := typedField(SectionLLM, "llm", "glm.models."+tier, TypeText)
		f.Description = "fieldDesc.llm.glm.models." + tier
		fields = append(fields, f)
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
		// SPEC-WEBCONF-SIMPLIFY-001 M4 (REQ-WC-004 / AC-WC-004): single enable/disable
		// toggle for the quality-extras feature, rendered on the launch tab (hand-built
		// in fieldsetLaunch). The detailed DDD-gate fields below stay baked/hidden.
		q("quality_extras_enabled", TypeBool),
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
		// workflow.team.* 편집 필드는 Agent Teams 정적 레이어와 함께 제거되었다
		// (SPEC-AGENT-TEAM-RETIRE-001). workflow.yaml team 블록은 M3에서 제거된다.
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
		// SPEC-WEB-CONSOLE-014 M2 (REQ-WC14-001): learning.auto_apply 편집 필드 철거 →
		// ReadOnlyDisplayFields로 강등. 파이프라인 in-memory AutoApply:true는 디스크 값
		// (auto_apply:false)을 절대 변경하지 않는 FROZEN 불변식(governance)이라 편집
		// 노출은 오해를 유발한다 (spec.md F3).
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
		// SPEC-WEB-CONSOLE-014 M2 (REQ-WC14-040): hook_metrics.output_path 편집 필드
		// 철거 → ReadOnlyDisplayFields로 강등. dead config — non-schema reader 0건이고
		// 실제 기록 경로는 hookMetricsRelPath 상수 고정(post_tool_duration.go)이라 편집
		// 노출은 오해를 유발한다 (spec.md F9). slow_hook_threshold_ms는 reader
		// 실존(post_tool_duration.go 로컬 struct 디코드)이라 editable 잔류(회귀 핀).
		s(SectionObservability, "observability", TypeInt, "observability", "hook_metrics", "slow_hook_threshold_ms"),

		// security (스칼라만 — 패턴 리스트는 raw view, REQ-WC11-062).
		s(SectionSecurity, "security", TypeBool, "security", "permission", "strict_mode"),
		s(SectionSecurity, "security", TypeBool, "security", "sandbox", "required"),
		s(SectionSecurity, "security", TypeText, "security", "sandbox", "docker_image"),
	}
}

// ─── SPEC-WEB-CONSOLE-013 M2: handoff / cache 섹션 (seam 전용) ────────────────

// handoffModeValues는 handoff.mode의 닫힌 집합 {manual, auto}이다 —
// handoff_inject.go가 SessionStart마다 문자 그대로 소비하는 값과 일치한다
// (REQ-WC13-011). 대소문자 정규화는 도입하지 않는다 (acceptance.md §D.3 edge).
var handoffModeValues = []string{"manual", "auto"}

// handoffFields는 handoff 섹션의 편집 FieldDef를 반환한다: mode(select) +
// guide(bool). 두 필드 모두 PersistSeam(handoff.yaml) — REQ-WC13-010.
func handoffFields() []FieldDef {
	return []FieldDef{
		withSelect(seamField(SectionHandoff, "handoff", TypeSelect, "handoff", "mode"),
			"f.handoff.mode.opt.", handoffModeValues, "", ""),
		seamField(SectionHandoff, "handoff", TypeBool, "handoff", "guide"),
	}
}

// cacheFields는 cache 섹션의 편집 FieldDef를 반환한다: cacheStrategy.enabled(bool)
// + cacheStrategy.session_ttl(select). session_ttl 옵션 집합은 config.
// ValidSessionTTLs()를 그대로 소비한다 — 재선언 금지, 드리프트 구조적 차단
// (REQ-WC13-012/013). spec_ttl / min_cacheable_tokens는 신설 FieldDef 대상에서
// 제외한다 (REQ-WC13-015); seam의 미모델링 키 보존이 파일 값을 무손상 유지한다
// (REQ-WC13-006).
func cacheFields() []FieldDef {
	return []FieldDef{
		seamField(SectionCache, "cache", TypeBool, "cacheStrategy", "enabled"),
		withSelect(seamField(SectionCache, "cache", TypeSelect, "cacheStrategy", "session_ttl"),
			"f.cacheStrategy.session_ttl.opt.", config.ValidSessionTTLs(), "", ""),
	}
}

// reportFormatValues는 report.format의 닫힌 집합이다 (html+md / md).
// moai-domain-html-report skill이 읽어 출력 포맷을 결정한다.
var reportFormatValues = []string{"html+md", "md"}

// reportFields는 report 섹션의 편집 FieldDef를 반환한다: format(radio). 2-옵션
// 닫힌 집합(html+md / md)이라 select-minimization으로 라디오 버튼 그룹으로 렌더한다
// (withRadio). report tab에서 제네릭 schemaFieldWidget(→ schemaRadioRow)로 렌더되며,
// seam 경로(report.yaml)로 영속화된다. Description은 fieldDesc.report.format i18n
// 키로 field-level 설명 문단을 방출한다.
func reportFields() []FieldDef {
	f := withRadio(seamField(SectionReport, "report", TypeRadio, "report", "format"),
		"f.report.format.opt.", reportFormatValues, "", "")
	f.Description = "fieldDesc.report.format"
	return []FieldDef{f}
}

// NOTE: agent-settings 웹 렌더 표면(team.role_profiles — 7 profiles ×
// {model, effort, isolation, mode})은 Agent Teams 정적 레이어와 함께 제거되었다
// (SPEC-AGENT-TEAM-RETIRE-001). 웹 콘솔은 더 이상 Agent Teams 설정을 렌더하지
// 않는다. sub-agent frontmatter 편집(agentfm.*)은 별도 표면(agentfm.go)으로
// 유지된다 — Agent Teams와 무관하다.

// sectionExtraFields는 M2b/M3 확장 필드 전체를 렌더 순서(AllSections 순서와
// 일치)로 반환한다.
func sectionExtraFields() []FieldDef {
	var fields []FieldDef
	fields = append(fields, qualityExtraFields()...) // SectionQuality 렌더 그룹에 합류
	fields = append(fields, gitStrategyFields()...)
	fields = append(fields, llmFields()...)
	fields = append(fields, seamSectionFields()...)
	fields = append(fields, handoffFields()...) // SPEC-WEB-CONSOLE-013 M2
	fields = append(fields, cacheFields()...)    // SPEC-WEB-CONSOLE-013 M2
	fields = append(fields, reportFields()...)   // report.format (launch tab)
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
	// NoteKey는 read-only 표시의 설명 라벨 i18n 키다. 빈 값이면 제네릭 "ro.note"
	// (runtime-managed)를 렌더한다. SPEC-WEB-CONSOLE-014 M2: governance FROZEN /
	// dead-config 강등 키에 정직한 설명 라벨을 부여한다 (REQ-WC14-001/040).
	NoteKey string
}

// ReadOnlyDisplayFields는 read-only 표시 키를 반환한다: SPEC-WEB-CONSOLE-014
// 강등 키 2종 (learning.auto_apply governance FROZEN,
// observability.hook_metrics.output_path dead config — REQ-WC14-001/040).
//
// llm.mode / llm.team_mode(구 REQ-WC11-013 runtime-managed 표시)는 콘솔 UI에서
// 제거됐다: 두 키는 `moai glm` / `moai cg` / `moai cc` 가 런타임에 기록하는 값이라
// 콘솔에서 읽어 보여줄 이유가 없다. 지속 경로는 struct 기반(LLMConfig 전체 재
// marshal, omitempty 없음)이므로 표시 목록에서 빠져도 yaml 키는 보존된다.
func ReadOnlyDisplayFields() []ReadOnlyField {
	return []ReadOnlyField{
		// SPEC-WEB-CONSOLE-014 M2 (REQ-WC14-001): governance FROZEN false.
		{SectionHarness, "harness", "learning.auto_apply", []string{"learning", "auto_apply"}, "ro.note.governance"},
		// SPEC-WEB-CONSOLE-014 M2 (REQ-WC14-040): dead config (기록 경로 상수 고정).
		{SectionObservability, "observability", "observability.hook_metrics.output_path", []string{"observability", "hook_metrics", "output_path"}, "ro.note.dead_config"},
	}
}

// RawBlockRef는 form UI에 맞지 않아 collapsed raw view로만 렌더하는 서브블록이다
// (REQ-WC11-062 — map-of-structs / 패턴 리스트).
type RawBlockRef struct {
	Section SectionID
	File    string
	Name    string // 표시 식별자
	Path    []string
	// NoteKey는 raw view 요약 라벨 i18n 키다. 빈 값이면 제네릭 "raw.note"
	// (structured block, read-only)를 렌더한다. SPEC-WEB-CONSOLE-014: 정보성
	// 표시(런타임 미배선/컴파일 상수 enforcement) 키에 정직한 라벨을 부여한다
	// (REQ-WC14-003/020 — F2-style honest label).
	NoteKey string
}

// RawViewBlocks는 raw view 대상 서브블록을 반환한다. SPEC-WEB-CONSOLE-014로
// learning(M2) + security.sandbox(M4) + mx(M4) 리스트 키가 추가되었다.
func RawViewBlocks() []RawBlockRef {
	return []RawBlockRef{
		// workflow.team.patterns raw view는 Agent Teams 정적 레이어와 함께 제거되었다
		// (SPEC-AGENT-TEAM-RETIRE-001).
		{SectionHarness, "harness", "harness.levels", []string{"harness", "levels"}, ""},
		{SectionSecurity, "security", "security.extra_dangerous_bash_patterns", []string{"security", "extra_dangerous_bash_patterns"}, ""},
		// SPEC-WEB-CONSOLE-014 M2 — learning list 키 (REQ-WC14-002/003).
		// tier_thresholds: 행동적 reader 실존(hook tier 분류) → 제네릭 라벨.
		{SectionHarness, "harness", "learning.tier_thresholds", []string{"learning", "tier_thresholds"}, ""},
		// rate_limit: 표시 전용 — enforcement는 컴파일 상수 → 정보성 정직 라벨.
		{SectionHarness, "harness", "learning.rate_limit", []string{"learning", "rate_limit"}, "raw.note.informational"},
		// SPEC-WEB-CONSOLE-014 M4 — security.sandbox list 키 (REQ-WC14-020, F6
		// config-미배선 scaffold — config는 로드되나 sandbox 실행 계층으로의
		// 브리지가 미배선 → F2-style 정보성 정직 라벨. 배선은 별도 SPEC).
		{SectionSecurity, "security", "security.sandbox.network_allowlist", []string{"security", "sandbox", "network_allowlist"}, "raw.note.informational"},
		{SectionSecurity, "security", "security.sandbox.env_scrub_extra", []string{"security", "sandbox", "env_scrub_extra"}, "raw.note.informational"},
		// SPEC-WEB-CONSOLE-014 M4 — mx list 키 (REQ-WC14-030, 행동적 reader 실존:
		// LoadDangerConfig / mx query fan-in·danger 분류가 소비 → 제네릭 라벨).
		// mx는 RouteExcluded 유지 — raw view는 표시 전용, 쓰기 경로 없음.
		{SectionMx, "mx", "mx.danger_categories", []string{"mx", "danger_categories"}, ""},
		{SectionMx, "mx", "mx.test_paths", []string{"mx", "test_paths"}, ""},
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
		SectionHandoff,
		SectionCache,
	}
}
