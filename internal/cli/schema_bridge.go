package cli

// 이 파일은 settings 스키마의 dotted i18n 키를 TUI 의 struct-필드-주소 번역 스토어
// (profileSetupText)로 연결하는 named bridge resolver 를 담는다(REQ-WC10-017a,
// design §F.2). 웹 스토어는 dotted 키를 직접 문자열 조회하지만, TUI 스토어는
// getProfileText(locale).<NamedField> 형태의 struct-필드 주소이므로 dotted 키를
// 직접 조회할 수 없다. 본 bridge 가 각 스키마 키를 해당 struct 필드 접근자로 매핑한다.
//
// @MX:NOTE: [AUTO] schemaKeyToTUIField 는 스키마 키 namespace 의 단일 원천을 유지하면서
// TUI 가 기존 struct-of-strings 번역 스토어를 그대로 쓰게 하는 어댑터다. 스키마 키 변경 시
// 본 맵도 함께 갱신해야 하며 AC-WC10-016(TestI18nKeySetParity)이 이를 검증한다.
//
// SPEC-CLI-UIKIT-KERNEL-001 b-ii split: the TuiLabel type + SchemaKeyToTUIField +
// FieldDefTUILabel helpers moved to internal/cli/uikit (leaf package). The two
// profileSetupText-coupled maps STAY here (package cli); uikit accesses them via
// the RegisterSchemaBridge callback registered in init() below. uikit has zero
// profileSetupText references (REQ-CUK-007 b-ii, verified by AC-CUK-007).

import "github.com/modu-ai/moai-adk/internal/cli/uikit"

// schemaFieldBridge는 스키마 필드의 i18n 키(예: "f.model")를 profileSetupText 의
// 해당 title/desc 접근자로 매핑한다. 키는 settings.FieldDef.I18nKey 와 일치한다.
// 모든 34개 필드(빈 라벨 제외)가 항목을 가진다.
var schemaFieldBridge = map[string]func(t profileSetupText) uikit.TuiLabel{
	// Identity
	"f.user_name": func(t profileSetupText) uikit.TuiLabel { return uikit.TuiLabel{Title: t.UserNameTitle, Desc: t.UserNameDesc} },
	// Language (conversation_lang 는 langForm 첫 단계에서 처리되나 스키마 parity 위해 매핑 유지)
	"f.conversation_lang": func(t profileSetupText) uikit.TuiLabel { return uikit.TuiLabel{Title: t.LangSelectTitle, Desc: t.LangSelectDesc} },
	"f.git_commit_lang": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.GitCommitLangTitle, Desc: t.GitCommitLangDesc}
	},
	"f.code_comment_lang": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.CodeCommentLangTitle, Desc: t.CodeCommentLangDesc}
	},
	"f.doc_lang": func(t profileSetupText) uikit.TuiLabel { return uikit.TuiLabel{Title: t.DocLangTitle, Desc: t.DocLangDesc} },
	// Launch
	"f.model": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.ModelOverrideTitle, Desc: t.ModelOverrideDesc}
	},
	"f.model_policy": func(t profileSetupText) uikit.TuiLabel { return uikit.TuiLabel{Title: t.ModelPolicyTitle, Desc: t.ModelPolicyDesc} },
	"f.effort_level": func(t profileSetupText) uikit.TuiLabel { return uikit.TuiLabel{Title: t.EffortLevelTitle, Desc: t.EffortLevelDesc} },
	"f.permission_mode": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.PermissionModeTitle, Desc: t.PermissionModeDesc}
	},
	// Statusline theme (16 segments handled by schemaSegmentBridge below)
	"f.statusline_theme": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.StatuslineThemeTitle, Desc: t.StatuslineThemeDesc}
	},
	// Quality
	"f.development_mode": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.DevelopmentModeTitle, Desc: t.DevelopmentModeDesc}
	},
	"f.quality.test_coverage_target": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.QualityCoverageTargetTitle, Desc: t.QualityCoverageTargetDesc}
	},
	"f.quality.enforce_quality": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.QualityEnforceQualityTitle, Desc: t.QualityEnforceQualityDesc}
	},
	"f.quality.tdd_settings.min_coverage_per_commit": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.QualityMinCoverageTitle, Desc: t.QualityMinCoverageDesc}
	},
	// Git Convention
	"f.git_convention": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.GitConventionTitle, Desc: t.GitConventionDesc}
	},
	"f.git_convention.auto_detection.enabled": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.GitAutoEnabledTitle, Desc: t.GitAutoEnabledDesc}
	},
	"f.git_convention.auto_detection.confidence_threshold": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.GitConfidenceTitle, Desc: t.GitConfidenceDesc}
	},
	"f.git_convention.auto_detection.sample_size": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.GitSampleSizeTitle, Desc: t.GitSampleSizeDesc}
	},
	"f.git_convention.validation.enforce_on_push": func(t profileSetupText) uikit.TuiLabel {
		return uikit.TuiLabel{Title: t.GitEnforceOnPushTitle, Desc: t.GitEnforceOnPushDesc}
	},
}

// schemaSegmentBridge는 16개 statusline 세그먼트의 스키마 키(예: "seg.git_branch")를
// profileSetupText 의 세그먼트 라벨 필드로 매핑한다. 세그먼트는 title 만 가진다(desc 없음).
var schemaSegmentBridge = map[string]func(t profileSetupText) string{
	"seg.cache_hit":       func(t profileSetupText) string { return t.SegmentCacheHit },
	"seg.claude_version":  func(t profileSetupText) string { return t.SegmentClaudeVersion },
	"seg.context":         func(t profileSetupText) string { return t.SegmentContext },
	"seg.directory":       func(t profileSetupText) string { return t.SegmentDirectory },
	"seg.effort_thinking": func(t profileSetupText) string { return t.SegmentEffortThinking },
	"seg.git_branch":      func(t profileSetupText) string { return t.SegmentGitBranch },
	"seg.git_status":      func(t profileSetupText) string { return t.SegmentGitStatus },
	"seg.moai_version":    func(t profileSetupText) string { return t.SegmentMoaiVersion },
	"seg.model":           func(t profileSetupText) string { return t.SegmentModel },
	"seg.output_style":    func(t profileSetupText) string { return t.SegmentOutputStyle },
	"seg.pr":              func(t profileSetupText) string { return t.SegmentPR },
	"seg.session_time":    func(t profileSetupText) string { return t.SegmentSessionTime },
	"seg.task":            func(t profileSetupText) string { return t.SegmentTask },
	"seg.usage_5h":        func(t profileSetupText) string { return t.SegmentUsage5h },
	"seg.usage_7d":        func(t profileSetupText) string { return t.SegmentUsage7d },
	"seg.worktree":        func(t profileSetupText) string { return t.SegmentWorktree },
}

// init registers the profileSetupText-aware schema bridge resolver with the
// uikit leaf package. The resolver closure captures the two maps above + the
// package-cli-internal getProfileText; uikit dispatches SchemaKeyToTUIField /
// FieldDefTUILabel through this callback (b-ii split, REQ-CUK-007).
func init() {
	uikit.RegisterSchemaBridge(func(schemaKey, locale string) (uikit.TuiLabel, bool) {
		t := getProfileText(locale)
		if fn, ok := schemaFieldBridge[schemaKey]; ok {
			return fn(t), true
		}
		if fn, ok := schemaSegmentBridge[schemaKey]; ok {
			return uikit.TuiLabel{Title: fn(t)}, true
		}
		return uikit.TuiLabel{}, false
	})
}
