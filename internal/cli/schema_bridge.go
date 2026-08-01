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

import (
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/settings"
)

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

// schemaOptionBridge maps a schema OptionDef.I18nKey (e.g. "f.model.opt.opus[1m]")
// to the profileSetupText accessor carrying that option's localized TUI label.
// It is the option-level counterpart of schemaFieldBridge above: the field bridge
// resolves widget titles/descriptions, this one resolves the individual option
// labels, so the TUI can render option lists that are DERIVED from
// settings.FieldOptionDefs rather than re-declared inline (the drift source that
// let the wizard keep writing canonical model ids the web console rejects).
//
// The key namespace mirrors internal/web/assets/i18n.js, so every f.<field>.opt.*
// key the web store carries has a TUI counterpart here — including alias values
// that the current picker does not offer (bare opus/sonnet/fable, opusplan). They
// stay mapped because ModelAliasTable still accepts them as stored values and
// template.ModelAliasPickerValues() may re-offer them.
//
// git_convention options are deliberately absent: the TUI never had localized
// labels for them (the wizard rendered the raw wire value), so optionLabelFor
// falls back to OptionDef.Value and the rendered surface is unchanged.
var schemaOptionBridge = map[string]func(t profileSetupText) string{
	// Model (f.model.opt.*)
	"f.model.opt.opus":       func(t profileSetupText) string { return t.ModelOpus },
	"f.model.opt.opus[1m]":   func(t profileSetupText) string { return t.ModelOpus1M },
	"f.model.opt.sonnet":     func(t profileSetupText) string { return t.ModelSonnet },
	"f.model.opt.sonnet[1m]": func(t profileSetupText) string { return t.ModelSonnet1M },
	"f.model.opt.fable":      func(t profileSetupText) string { return t.ModelFable },
	"f.model.opt.fable[1m]":  func(t profileSetupText) string { return t.ModelFable1M },
	"f.model.opt.haiku":      func(t profileSetupText) string { return t.ModelHaiku },
	"f.model.opt.opusplan":   func(t profileSetupText) string { return t.ModelOpusPlan },
	// Effort level (f.effort_level.opt.*)
	"f.effort_level.opt.low":    func(t profileSetupText) string { return t.EffortLevelLow },
	"f.effort_level.opt.medium": func(t profileSetupText) string { return t.EffortLevelMedium },
	"f.effort_level.opt.high":   func(t profileSetupText) string { return t.EffortLevelHigh },
	"f.effort_level.opt.xhigh":  func(t profileSetupText) string { return t.EffortLevelXHigh },
	"f.effort_level.opt.max":    func(t profileSetupText) string { return t.EffortLevelMax },
	// Permission mode (f.permission_mode.opt.*)
	"f.permission_mode.opt.acceptEdits":       func(t profileSetupText) string { return t.PermAcceptEdits },
	"f.permission_mode.opt.auto":              func(t profileSetupText) string { return t.PermAuto },
	"f.permission_mode.opt.default":           func(t profileSetupText) string { return t.PermDefault },
	"f.permission_mode.opt.plan":              func(t profileSetupText) string { return t.PermPlan },
	"f.permission_mode.opt.bypassPermissions": func(t profileSetupText) string { return t.PermBypass },
	"f.permission_mode.opt.dontAsk":           func(t profileSetupText) string { return t.PermDontAsk },
	// Development mode (f.development_mode.opt.*)
	"f.development_mode.opt.ddd": func(t profileSetupText) string { return t.DevelopmentModeDDD },
	"f.development_mode.opt.tdd": func(t profileSetupText) string { return t.DevelopmentModeTDD },
}

// optionLabelFor resolves one schema option to its localized TUI label. Options
// with no bridge entry (or an empty translation) fall back to the raw wire value,
// which is what the wizard already rendered for git_convention.
func optionLabelFor(t profileSetupText, o settings.OptionDef) string {
	if fn, ok := schemaOptionBridge[o.I18nKey]; ok {
		if label := fn(t); label != "" {
			return label
		}
	}
	return o.Value
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
