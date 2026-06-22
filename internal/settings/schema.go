// Package settings는 두 설정 표면(moai web 콘솔, moai profile setup TUI)이
// 공유하는 정규 필드 스키마(SSOT)를 정의한다. internal/cli 와 internal/web 어느
// 쪽도 import 하지 않는 중립 패키지로, 두 표면이 모두 이 패키지를 import 하여
// 동일한 6개 섹션 / 34개 필드의 옵션 목록·빈 옵션 라벨·검증 규칙·i18n 키·영속화
// 대상을 단일 원천에서 파생한다. (SPEC-WEB-CONSOLE-010)
//
// @MX:ANCHOR: [AUTO] settings 스키마는 TUI(huh)와 웹(templ) 두 표면의 단일 진실 공급원이다.
// @MX:REASON: [AUTO] fan_in >= 2 (internal/cli + internal/web 가 모두 의존); 두 표면의
// 필드 정의·옵션 목록·빈 라벨·i18n 키·영속화 경로가 이 데이터에서 파생되므로 불변 계약이다.
package settings

import (
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/statusline"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// SectionID는 6개 정규 섹션을 식별한다.
type SectionID string

// 6개 정규 섹션. USED-필드 합집합(34개 필드)을 덮는다.
const (
	SectionIdentity      SectionID = "identity"       // user_name (1)
	SectionLanguage      SectionID = "language"       // conversation/git_commit/code_comment/doc (4)
	SectionLaunch        SectionID = "launch"         // model/model_policy/effort_level/permission_mode (4)
	SectionStatusline    SectionID = "statusline"     // theme + 15 segments (16)
	SectionQuality       SectionID = "quality"        // development_mode + 3 nested (4)
	SectionGitConvention SectionID = "git_convention" // convention + 4 nested (5)
)

// AllSections는 정규 섹션을 렌더 순서대로 반환한다.
func AllSections() []SectionID {
	return []SectionID{
		SectionIdentity,
		SectionLanguage,
		SectionLaunch,
		SectionStatusline,
		SectionQuality,
		SectionGitConvention,
	}
}

// FieldType는 필드의 위젯/값 종류를 나타낸다.
type FieldType string

const (
	TypeSelect      FieldType = "select"       // 단일 선택 (huh.Select / <select>)
	TypeMultiSelect FieldType = "multi-select" // 다중 선택 (huh.MultiSelect / 세그먼트 체크박스 묶음)
	TypeText        FieldType = "text"         // 자유 텍스트 (huh.Input / <input type=text>)
	TypeInt         FieldType = "int"          // 정수 (<input type=number>)
	TypeFloat       FieldType = "float"        // 실수 (<input type=number step=0.01>)
	TypeBool        FieldType = "bool"         // 불리언 토글 (체크박스)
)

// PersistKind는 필드 값이 어디에 영속화되는지를 구분한다.
type PersistKind string

const (
	// PersistProfileStore: 프로필 스토어(preferences.yaml)에 저장되며 일부는
	// SyncToProjectConfig 를 통해 user/language/statusline.yaml 로 동기화된다.
	PersistProfileStore PersistKind = "profile-store"
	// PersistProjectConfig: config 매니저를 통해 quality.yaml /
	// git-convention.yaml 의 <section>.<key> 경로에 저장된다.
	PersistProjectConfig PersistKind = "project-config"
)

// PersistTarget은 필드 값의 영속화 대상을 선언한다. ProfileStore 필드의 경우
// Field 에 ProfilePreferences 의 논리 키(예: "model")를, ProjectConfig 필드의
// 경우 Section + Key 에 yaml 섹션/키 경로(예: "quality" / "test_coverage_target")를
// 담는다. Normalize 가 non-nil 이면 저장 직전 값 정규화를 수행한다(예:
// permission_mode 의 acceptEdits → "" 정규화, REQ-WC10-014).
type PersistTarget struct {
	Kind      PersistKind
	Field     string              // ProfileStore: 논리 필드명
	Section   string              // ProjectConfig: yaml 섹션명
	Key       string              // ProjectConfig: dot-path 키(예: "tdd_settings.min_coverage_per_commit")
	Normalize func(string) string // 저장 직전 값 정규화 (nil이면 그대로 저장)
}

// OptionDef는 select/multi-select 필드의 한 옵션이다. Value 는 정규 와이어 값,
// I18nKey 는 표면별 사람 친화 라벨을 해석하는 i18n 키다.
type OptionDef struct {
	Value   string
	I18nKey string
}

// FieldDef는 정규 필드 한 개를 기술하는 데이터 레코드다. 동작이 아닌 데이터다 —
// 두 표면은 이 레코드에서 위젯/검증/영속화 대상을 파생한다.
type FieldDef struct {
	Name          string            // 논리 필드명 (예: "model", "quality.test_coverage_target")
	Section       SectionID         // 6개 정규 섹션 중 하나
	Type          FieldType         // 위젯/값 종류
	Options       []OptionDef       // select/multi-select 전용. 그 외 nil
	EmptyLabel    string            // 정규 빈 옵션 라벨(4개 드리프트를 단일화)
	EmptyLabelKey string            // 빈 옵션 라벨의 i18n 키
	Validate      func(string) bool // 검증 술어 (nil이면 항상 유효)
	I18nKey       string            // 두 스토어가 해석하는 공유 i18n 키 prefix (예: "f.model")
	Persist       PersistTarget     // 값 영속화 대상
}

// statuslineSegmentKeys는 정규 15개 세그먼트 키를 렌더 순서대로 반환한다.
// internal/statusline.Segment* 상수에서 파생되며(단일 원천) SegmentRepo(16번째)는
// 의도적으로 제외된다(15-키 스키마 밖, sync.go:154).
func statuslineSegmentKeys() []string {
	return []string{
		statusline.SegmentClaudeVersion,
		statusline.SegmentContext,
		statusline.SegmentDirectory,
		statusline.SegmentEffortThinking,
		statusline.SegmentGitBranch,
		statusline.SegmentGitStatus,
		statusline.SegmentMoaiVersion,
		statusline.SegmentModel,
		statusline.SegmentOutputStyle,
		statusline.SegmentPR,
		statusline.SegmentSessionTime,
		statusline.SegmentTask,
		statusline.SegmentUsage5H,
		statusline.SegmentUsage7D,
		statusline.SegmentWorktree,
	}
}

// languageOptions는 4개 지원 언어(en/ko/ja/zh)를 옵션으로 반환한다. 정규 언어
// 목록을 스키마가 단일 소유하므로 internal/web 이 langOptions 를 재선언하지 않는다.
func languageOptions() []OptionDef {
	codes := []string{"en", "ko", "ja", "zh"}
	out := make([]OptionDef, 0, len(codes))
	for _, c := range codes {
		out = append(out, OptionDef{Value: c, I18nKey: "lang.opt." + c})
	}
	return out
}

// modelOptions는 정규 모델 옵션 목록을 반환한다. internal/web/validate.go 의
// 손수-미러된 modelCanonical 을 대체하는 단일 원천이다(REQ-WC10-004).
func modelOptions() []OptionDef {
	values := []string{"opus", "opus[1m]", "sonnet", "sonnet[1m]", "haiku", "opusplan"}
	return optionDefsFromValues("f.model.opt.", values)
}

// effortOptions는 정규 effort level 옵션 목록을 반환한다(REQ-WC10-004).
func effortOptions() []OptionDef {
	values := []string{"low", "medium", "high", "xhigh", "max"}
	return optionDefsFromValues("f.effort_level.opt.", values)
}

// modelPolicyOptions는 정규 model policy 옵션 목록을 template.ValidModelPolicies()
// 에서 파생하여 반환한다(중복 선언 금지, 기존 export 재사용).
func modelPolicyOptions() []OptionDef {
	return optionDefsFromValues("f.model_policy.opt.", template.ValidModelPolicies())
}

// permissionModeOptions는 정규 permission mode 옵션 목록을 반환한다. 빈 문자열은
// 빈 옵션 라벨이 별도로 처리하므로 옵션 목록에는 비-빈 값만 포함한다. 순서는
// TUI 위저드의 기존 순서(acceptEdits, auto, default, plan, bypass, dontAsk)를 보존한다.
func permissionModeOptions() []OptionDef {
	values := []string{"acceptEdits", "auto", "default", "plan", "bypassPermissions", "dontAsk"}
	return optionDefsFromValues("f.permission_mode.opt.", values)
}

// developmentModeOptions는 정규 development mode 옵션 목록을
// models.ValidDevelopmentModes() 에서 파생한다(REQ-WC10-004).
func developmentModeOptions() []OptionDef {
	modes := models.ValidDevelopmentModes()
	values := make([]string, 0, len(modes))
	for _, m := range modes {
		values = append(values, string(m))
	}
	return optionDefsFromValues("f.development_mode.opt.", values)
}

// conventionOptions는 정규 git convention 옵션 목록을 반환한다. 결정적 렌더 순서를
// 위해 순서를 고정한다(config.ValidConventions 는 map-order 라 비결정적).
func conventionOptions() []OptionDef {
	values := []string{"auto", "conventional-commits", "angular", "karma"}
	return optionDefsFromValues("f.git_convention.opt.", values)
}

// optionDefsFromValues는 값 슬라이스를 prefix+value 의 i18n 키를 가진 OptionDef
// 슬라이스로 변환한다.
func optionDefsFromValues(keyPrefix string, values []string) []OptionDef {
	out := make([]OptionDef, 0, len(values))
	for _, v := range values {
		out = append(out, OptionDef{Value: v, I18nKey: keyPrefix + v})
	}
	return out
}

// inOptionValues는 v 가 옵션 값 집합의 멤버인지 보고한다(멤버십 검증 술어).
func inOptionValues(opts []OptionDef, v string) bool {
	for _, o := range opts {
		if o.Value == v {
			return true
		}
	}
	return false
}

// normalizePermissionMode는 permission_mode 의 저장 직전 정규화다: 프로젝트 기본값
// acceptEdits 는 빈 문자열로 정규화하여 불필요한 override 를 기록하지 않는다
// (REQ-WC10-014, 기존 profile_setup.go:443 의미 보존).
func normalizePermissionMode(v string) string {
	if v == defaultPermissionMode {
		return ""
	}
	return v
}

// defaultPermissionMode는 프로젝트 기본 permission mode 다(profile_setup.go 와 일치).
const defaultPermissionMode = "acceptEdits"

// statuslineThemeOptions는 정규 statusline 테마 옵션을 반환한다(TUI 위저드와 일치).
func statuslineThemeOptions() []OptionDef {
	return []OptionDef{
		{Value: "catppuccin-mocha", I18nKey: "f.statusline_theme.opt.catppuccin-mocha"},
		{Value: "catppuccin-latte", I18nKey: "f.statusline_theme.opt.catppuccin-latte"},
	}
}

// 정규 빈 옵션 라벨. 4개 드리프트(lang/model/effort/git_convention)를 단일화하여
// 두 표면이 동일한 라벨을 렌더하도록 한다(REQ-WC10-013).
const (
	emptyLabelUnset          = "(unset)"
	emptyLabelProjectDefault = "(project default)"
	emptyLabelRuntimeDefault = "(runtime default)"
)

// allFields는 6개 섹션의 34개 정규 필드를 렌더 순서대로 구성하여 반환한다.
// 모든 select/multi-select 필드의 옵션은 위 헬퍼(단일 원천)에서 파생된다.
func allFields() []FieldDef {
	fields := make([]FieldDef, 0, 34)

	// ── Section 1: Identity (1) ──────────────────────────────────────────
	fields = append(fields, FieldDef{
		Name:    "user_name",
		Section: SectionIdentity,
		Type:    TypeText,
		I18nKey: "f.user_name",
		Persist: PersistTarget{Kind: PersistProfileStore, Field: "user_name"},
	})

	// ── Section 2: Language (4) ──────────────────────────────────────────
	langOpts := languageOptions()
	langValidate := func(v string) bool { return inOptionValues(langOpts, v) }
	for _, lf := range []struct {
		name, i18n, field string
	}{
		{"conversation_lang", "f.conversation_lang", "conversation_lang"},
		{"git_commit_lang", "f.git_commit_lang", "git_commit_lang"},
		{"code_comment_lang", "f.code_comment_lang", "code_comment_lang"},
		{"doc_lang", "f.doc_lang", "doc_lang"},
	} {
		fields = append(fields, FieldDef{
			Name:          lf.name,
			Section:       SectionLanguage,
			Type:          TypeSelect,
			Options:       langOpts,
			EmptyLabel:    emptyLabelUnset,
			EmptyLabelKey: "opt.unset",
			Validate:      langValidate,
			I18nKey:       lf.i18n,
			Persist:       PersistTarget{Kind: PersistProfileStore, Field: lf.field},
		})
	}

	// ── Section 3: Launch (4) ────────────────────────────────────────────
	modelOpts := modelOptions()
	effortOpts := effortOptions()
	policyOpts := modelPolicyOptions()
	permOpts := permissionModeOptions()
	fields = append(fields,
		FieldDef{
			Name:          "model",
			Section:       SectionLaunch,
			Type:          TypeSelect,
			Options:       modelOpts,
			EmptyLabel:    emptyLabelProjectDefault,
			EmptyLabelKey: "opt.project_default",
			Validate:      func(v string) bool { return inOptionValues(modelOpts, v) },
			I18nKey:       "f.model",
			Persist:       PersistTarget{Kind: PersistProfileStore, Field: "model"},
		},
		FieldDef{
			Name:          "model_policy",
			Section:       SectionLaunch,
			Type:          TypeSelect,
			Options:       policyOpts,
			EmptyLabel:    emptyLabelProjectDefault,
			EmptyLabelKey: "opt.project_default",
			Validate:      template.IsValidModelPolicy,
			I18nKey:       "f.model_policy",
			Persist:       PersistTarget{Kind: PersistProfileStore, Field: "model_policy"},
		},
		FieldDef{
			Name:          "effort_level",
			Section:       SectionLaunch,
			Type:          TypeSelect,
			Options:       effortOpts,
			EmptyLabel:    emptyLabelRuntimeDefault,
			EmptyLabelKey: "opt.runtime_default",
			Validate:      func(v string) bool { return inOptionValues(effortOpts, v) },
			I18nKey:       "f.effort_level",
			Persist:       PersistTarget{Kind: PersistProfileStore, Field: "effort_level"},
		},
		FieldDef{
			Name:          "permission_mode",
			Section:       SectionLaunch,
			Type:          TypeSelect,
			Options:       permOpts,
			EmptyLabel:    emptyLabelProjectDefault,
			EmptyLabelKey: "opt.project_default",
			Validate:      profile.IsValidPermissionMode,
			I18nKey:       "f.permission_mode",
			Persist: PersistTarget{
				Kind:      PersistProfileStore,
				Field:     "permission_mode",
				Normalize: normalizePermissionMode,
			},
		},
	)

	// ── Section 4: Statusline (16: theme + 15 segments) ──────────────────
	themeOpts := statuslineThemeOptions()
	fields = append(fields, FieldDef{
		Name:          "statusline_theme",
		Section:       SectionStatusline,
		Type:          TypeSelect,
		Options:       themeOpts,
		EmptyLabel:    "",
		EmptyLabelKey: "",
		Validate:      func(v string) bool { return inOptionValues(themeOpts, v) },
		I18nKey:       "f.statusline_theme",
		Persist:       PersistTarget{Kind: PersistProfileStore, Field: "statusline_theme"},
	})
	for _, seg := range statuslineSegmentKeys() {
		fields = append(fields, FieldDef{
			Name:    "statusline_segment." + seg,
			Section: SectionStatusline,
			Type:    TypeBool,
			I18nKey: "seg." + seg,
			Persist: PersistTarget{Kind: PersistProfileStore, Field: "statusline_segments." + seg},
		})
	}

	// ── Section 5: Quality (4: development_mode + 3 nested) ───────────────
	devOpts := developmentModeOptions()
	fields = append(fields,
		FieldDef{
			Name:          "development_mode",
			Section:       SectionQuality,
			Type:          TypeSelect,
			Options:       devOpts,
			EmptyLabel:    emptyLabelProjectDefault,
			EmptyLabelKey: "opt.project_default",
			Validate:      func(v string) bool { return models.DevelopmentMode(v).IsValid() },
			I18nKey:       "f.development_mode",
			Persist:       PersistTarget{Kind: PersistProjectConfig, Section: "quality", Key: "development_mode"},
		},
		FieldDef{
			Name:    "quality.test_coverage_target",
			Section: SectionQuality,
			Type:    TypeInt,
			I18nKey: "f.quality.test_coverage_target",
			Persist: PersistTarget{Kind: PersistProjectConfig, Section: "quality", Key: "test_coverage_target"},
		},
		FieldDef{
			Name:    "quality.enforce_quality",
			Section: SectionQuality,
			Type:    TypeBool,
			I18nKey: "f.quality.enforce_quality",
			Persist: PersistTarget{Kind: PersistProjectConfig, Section: "quality", Key: "enforce_quality"},
		},
		FieldDef{
			Name:    "quality.tdd_settings.min_coverage_per_commit",
			Section: SectionQuality,
			Type:    TypeInt,
			I18nKey: "f.quality.tdd_settings.min_coverage_per_commit",
			Persist: PersistTarget{Kind: PersistProjectConfig, Section: "quality", Key: "tdd_settings.min_coverage_per_commit"},
		},
	)

	// ── Section 6: Git Convention (5: convention + 4 nested) ──────────────
	convOpts := conventionOptions()
	fields = append(fields,
		FieldDef{
			Name:          "git_convention",
			Section:       SectionGitConvention,
			Type:          TypeSelect,
			Options:       convOpts,
			EmptyLabel:    emptyLabelProjectDefault,
			EmptyLabelKey: "opt.project_default",
			Validate:      config.IsValidConvention,
			I18nKey:       "f.git_convention",
			Persist:       PersistTarget{Kind: PersistProjectConfig, Section: "git_convention", Key: "convention"},
		},
		FieldDef{
			Name:    "git_convention.auto_detection.enabled",
			Section: SectionGitConvention,
			Type:    TypeBool,
			I18nKey: "f.git_convention.auto_detection.enabled",
			Persist: PersistTarget{Kind: PersistProjectConfig, Section: "git_convention", Key: "auto_detection.enabled"},
		},
		FieldDef{
			Name:    "git_convention.auto_detection.confidence_threshold",
			Section: SectionGitConvention,
			Type:    TypeFloat,
			I18nKey: "f.git_convention.auto_detection.confidence_threshold",
			Persist: PersistTarget{Kind: PersistProjectConfig, Section: "git_convention", Key: "auto_detection.confidence_threshold"},
		},
		FieldDef{
			Name:    "git_convention.auto_detection.sample_size",
			Section: SectionGitConvention,
			Type:    TypeInt,
			I18nKey: "f.git_convention.auto_detection.sample_size",
			Persist: PersistTarget{Kind: PersistProjectConfig, Section: "git_convention", Key: "auto_detection.sample_size"},
		},
		FieldDef{
			Name:    "git_convention.validation.enforce_on_push",
			Section: SectionGitConvention,
			Type:    TypeBool,
			I18nKey: "f.git_convention.validation.enforce_on_push",
			Persist: PersistTarget{Kind: PersistProjectConfig, Section: "git_convention", Key: "validation.enforce_on_push"},
		},
	)

	return fields
}
