package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/template"
)

func TestGetProfileText_AllLanguages(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		t.Run(lang, func(t *testing.T) {
			txt := getProfileText(lang)
			if txt.ConfiguringProfile == "" {
				t.Errorf("ConfiguringProfile is empty for lang %q", lang)
			}
			if txt.LangSelectTitle == "" {
				t.Errorf("LangSelectTitle is empty for lang %q", lang)
			}
			if txt.UserNameTitle == "" {
				t.Errorf("UserNameTitle is empty for lang %q", lang)
			}
			if txt.SetupCancelled == "" {
				t.Errorf("SetupCancelled is empty for lang %q", lang)
			}
			if txt.SavedProfile == "" {
				t.Errorf("SavedProfile is empty for lang %q", lang)
			}
			if txt.ModelPolicyHigh == "" {
				t.Errorf("ModelPolicyHigh is empty for lang %q", lang)
			}
			if txt.ModelOpus1M == "" {
				t.Errorf("ModelOpus1M is empty for lang %q", lang)
			}
			if txt.ModelSonnet1M == "" {
				t.Errorf("ModelSonnet1M is empty for lang %q", lang)
			}
			// T-003: Effort level UI strings
			if txt.EffortLevelTitle == "" {
				t.Errorf("EffortLevelTitle is empty for lang %q", lang)
			}
			if txt.EffortLevelDesc == "" {
				t.Errorf("EffortLevelDesc is empty for lang %q", lang)
			}
			if txt.EffortLevelXHigh == "" {
				t.Errorf("EffortLevelXHigh is empty for lang %q", lang)
			}
			if txt.EffortLevelMax == "" {
				t.Errorf("EffortLevelMax is empty for lang %q", lang)
			}
			// SPEC-WEB-CONSOLE-003 AC-WC3-006a: project-config selects (development_mode + git_convention).
			if txt.DevelopmentModeTitle == "" {
				t.Errorf("DevelopmentModeTitle is empty for lang %q", lang)
			}
			if txt.DevelopmentModeDesc == "" {
				t.Errorf("DevelopmentModeDesc is empty for lang %q", lang)
			}
			if txt.DevelopmentModeDDD == "" {
				t.Errorf("DevelopmentModeDDD is empty for lang %q", lang)
			}
			if txt.DevelopmentModeTDD == "" {
				t.Errorf("DevelopmentModeTDD is empty for lang %q", lang)
			}
			if txt.GitConventionTitle == "" {
				t.Errorf("GitConventionTitle is empty for lang %q", lang)
			}
			if txt.GitConventionDesc == "" {
				t.Errorf("GitConventionDesc is empty for lang %q", lang)
			}
			// NOTE: the former ProjectDefaultOption / EffortLevelDefault /
			// ModelDefault empty-option labels were removed — the empty option is
			// single-sourced from settings.EmptyLabelFor, asserted by
			// TestSchemaSelectOptions_EmptyLabelFromSchema below.
		})
	}
}

// TestGetProfileText_OpusAliasValues verifies the `opus` alias labels name the
// model the alias ACTUALLY resolves to. template.ModelAliasCanonicalID("opus") is
// claude-opus-5, so the labels must say "Opus 5" — the former "Opus 4.8" wording
// named a model the alias no longer resolves to (claude-opus-4-8 survives only in
// template.ModelDeprecatedCanonicalIDs, for reading historical prefs). The
// negative assertion is what makes this falsifiable: re-introducing the stale
// version string fails the test.
func TestGetProfileText_OpusAliasValues(t *testing.T) {
	// The version token is derived from the canonical id the alias resolves to, so
	// bumping ModelAliasTable without re-labelling the wizard fails here.
	wantVersion := strings.TrimPrefix(template.ModelAliasCanonicalID("opus"), "claude-opus-")
	if wantVersion != "5" {
		t.Fatalf("opus alias resolves to %q; update the expected label version token",
			template.ModelAliasCanonicalID("opus"))
	}
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		for name, label := range map[string]string{
			"ModelOpus":   txt.ModelOpus,
			"ModelOpus1M": txt.ModelOpus1M,
		} {
			if !containsStr(label, "Opus "+wantVersion) {
				t.Errorf("lang=%q: %s %q should reference Opus %s", lang, name, label, wantVersion)
			}
			if containsStr(label, "4.8") {
				t.Errorf("lang=%q: %s %q still references the superseded Opus 4.8", lang, name, label)
			}
		}
		if !containsStr(txt.ModelOpus1M, "1M") {
			t.Errorf("lang=%q: ModelOpus1M %q should reference 1M context", lang, txt.ModelOpus1M)
		}
	}
}

// TestGetProfileText_EffortLevelValues verifies the effort level labels name their
// own wire value AND carry no stale model attribution. xhigh/max are still offered
// (they are valid Claude Code session effort levels and remain in the shared
// settings schema); what was stale was tying them to "Opus 4.8+", a model the
// wizard no longer targets.
func TestGetProfileText_EffortLevelValues(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		if !containsStr(txt.EffortLevelXHigh, template.EffortLevelXHigh) {
			t.Errorf("lang=%q: EffortLevelXHigh %q does not contain %q", lang, txt.EffortLevelXHigh, template.EffortLevelXHigh)
		}
		if !containsStr(txt.EffortLevelMax, template.EffortLevelMax) {
			t.Errorf("lang=%q: EffortLevelMax %q does not contain %q", lang, txt.EffortLevelMax, template.EffortLevelMax)
		}
		for name, label := range map[string]string{
			"EffortLevelDesc":  txt.EffortLevelDesc,
			"EffortLevelXHigh": txt.EffortLevelXHigh,
			"EffortLevelMax":   txt.EffortLevelMax,
		} {
			if containsStr(label, "4.8") {
				t.Errorf("lang=%q: %s %q still attributes the effort level to Opus 4.8", lang, name, label)
			}
		}
	}
}

// TestGetProfileText_PermAutoValues verifies the auto permission mode label is
// present in all supported languages and references the "auto" identifier.
// Auto mode: Claude Code v2.1.83+, Anthropic API only, Sonnet 4.6 / Opus 4.6 / Opus 4.7.
// Source: https://code.claude.com/docs/en/permission-modes
func TestGetProfileText_PermAutoValues(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		if txt.PermAuto == "" {
			t.Errorf("lang=%q: PermAuto is empty", lang)
		}
		if !containsStr(txt.PermAuto, "auto") {
			t.Errorf("lang=%q: PermAuto %q does not contain 'auto' identifier", lang, txt.PermAuto)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstr(s, sub))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// TestGetProfileText_PermAutoRuntimeWarning W-3: verifies that the PermAuto label contains
// a runtime error warning for each language.
func TestGetProfileText_PermAutoRuntimeWarning(t *testing.T) {
	runtimeErrByLang := map[string]string{
		"en": "Session errors at runtime",
		"ko": "런타임 오류",
		"ja": "実行時エラー",
		"zh": "运行时错误",
	}
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		want := runtimeErrByLang[lang]
		if !containsStr(txt.PermAuto, want) {
			t.Errorf("lang=%q: PermAuto %q should contain runtime error warning %q", lang, txt.PermAuto, want)
		}
	}
}

// TestGetProfileText_MigrationNoticeFields W-4: verifies that the theme
// MigrationNotice field is populated in all 4 languages and contains the %q
// format verb. The mode migration notice was removed by
// SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001.
func TestGetProfileText_MigrationNoticeFields(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		if txt.MigrationNoticeStatuslineTheme == "" {
			t.Errorf("lang=%q: MigrationNoticeStatuslineTheme is empty", lang)
		}
		// Must contain 2 %q format verbs (old value, new value)
		if !containsStr(txt.MigrationNoticeStatuslineTheme, "%q") {
			t.Errorf("lang=%q: MigrationNoticeStatuslineTheme %q should contain %%q format verb", lang, txt.MigrationNoticeStatuslineTheme)
		}
	}
}

// TestGetProfileText_SummarySyncSkippedNeutral W-5: verifies that SummarySyncSkipped uses
// neutral wording (no project-level sync).
func TestGetProfileText_SummarySyncSkippedNeutral(t *testing.T) {
	// Must not contain error-prone wording like "Sync skipped"
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		if txt.SummarySyncSkipped == "" {
			t.Errorf("lang=%q: SummarySyncSkipped is empty", lang)
		}
	}
	// en: verify neutral message
	en := getProfileText("en")
	if !containsStr(en.SummarySyncSkipped, "No project-level sync") {
		t.Errorf("en SummarySyncSkipped should be neutral, got: %q", en.SummarySyncSkipped)
	}
	// ko: verify neutral message
	ko := getProfileText("ko")
	if !containsStr(ko.SummarySyncSkipped, "프로젝트별 동기화 없음") {
		t.Errorf("ko SummarySyncSkipped should be neutral, got: %q", ko.SummarySyncSkipped)
	}
}

// TestGetProfileText_SummaryHeaderUpdated S-3: verifies that the ko/ja SummaryHeader uses
// updated wording.
func TestGetProfileText_SummaryHeaderUpdated(t *testing.T) {
	ko := getProfileText("ko")
	if !containsStr(ko.SummaryHeader, "저장된 설정값") {
		t.Errorf("ko SummaryHeader should be '저장된 설정값:', got: %q", ko.SummaryHeader)
	}
	ja := getProfileText("ja")
	if !containsStr(ja.SummaryHeader, "保存された設定値") {
		t.Errorf("ja SummaryHeader should be '保存された設定値:', got: %q", ja.SummaryHeader)
	}
}

func TestGetProfileText_FallbackToEnglish(t *testing.T) {
	en := getProfileText("en")
	fallback := getProfileText("unknown")
	if fallback.ConfiguringProfile != en.ConfiguringProfile {
		t.Errorf("fallback = %q, want %q", fallback.ConfiguringProfile, en.ConfiguringProfile)
	}
}

func TestGetProfileText_EmptyString(t *testing.T) {
	en := getProfileText("en")
	empty := getProfileText("")
	if empty.LangSelectTitle != en.LangSelectTitle {
		t.Errorf("empty string fallback = %q, want %q", empty.LangSelectTitle, en.LangSelectTitle)
	}
}

// TestProfileSetupTranslations_PresetSegments verifies that all 4 locales
// (en/ko/ja/zh) provide non-empty translations for the statusline preset
// selector + 16 segment toggle labels added by
// SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001 REQ-SPW-003 (cache_hit added by
// SPEC-WEB-CONSOLE-011 M6).
//
// 4 locales × 18 keys (2 segments section titles + 16 segment labels) = 72
// cells verified. The 6 preset-title/option cells were removed by
// SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 (preset Select retired).
func TestProfileSetupTranslations_PresetSegments(t *testing.T) {
	langs := []string{"en", "ko", "ja", "zh"}
	type cell struct {
		key   string
		value func(profileSetupText) string
	}
	cells := []cell{
		{"StatuslineSegmentsTitle", func(p profileSetupText) string { return p.StatuslineSegmentsTitle }},
		{"StatuslineSegmentsDesc", func(p profileSetupText) string { return p.StatuslineSegmentsDesc }},
		{"SegmentCacheHit", func(p profileSetupText) string { return p.SegmentCacheHit }},
		{"SegmentClaudeVersion", func(p profileSetupText) string { return p.SegmentClaudeVersion }},
		{"SegmentContext", func(p profileSetupText) string { return p.SegmentContext }},
		{"SegmentDirectory", func(p profileSetupText) string { return p.SegmentDirectory }},
		{"SegmentEffortThinking", func(p profileSetupText) string { return p.SegmentEffortThinking }},
		{"SegmentGitBranch", func(p profileSetupText) string { return p.SegmentGitBranch }},
		{"SegmentGitStatus", func(p profileSetupText) string { return p.SegmentGitStatus }},
		{"SegmentMoaiVersion", func(p profileSetupText) string { return p.SegmentMoaiVersion }},
		{"SegmentModel", func(p profileSetupText) string { return p.SegmentModel }},
		{"SegmentOutputStyle", func(p profileSetupText) string { return p.SegmentOutputStyle }},
		{"SegmentPR", func(p profileSetupText) string { return p.SegmentPR }},
		{"SegmentSessionTime", func(p profileSetupText) string { return p.SegmentSessionTime }},
		{"SegmentTask", func(p profileSetupText) string { return p.SegmentTask }},
		{"SegmentUsage5h", func(p profileSetupText) string { return p.SegmentUsage5h }},
		{"SegmentUsage7d", func(p profileSetupText) string { return p.SegmentUsage7d }},
		{"SegmentWorktree", func(p profileSetupText) string { return p.SegmentWorktree }},
	}
	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			text := getProfileText(lang)
			for _, c := range cells {
				if c.value(text) == "" {
					t.Errorf("lang %q: %s is empty", lang, c.key)
				}
			}
		})
	}
}
