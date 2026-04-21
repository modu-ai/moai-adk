package cli

import "testing"

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
			if txt.EffortLevelDefault == "" {
				t.Errorf("EffortLevelDefault is empty for lang %q", lang)
			}
			if txt.EffortLevelXHigh == "" {
				t.Errorf("EffortLevelXHigh is empty for lang %q", lang)
			}
			if txt.EffortLevelMax == "" {
				t.Errorf("EffortLevelMax is empty for lang %q", lang)
			}
		})
	}
}

// TestGetProfileText_OpusAliasValues verifies the `opus` alias label in every
// supported language advertises Opus 4.7 (per the simplified wizard UX where
// explicit claude-opus-4-7 options are removed and `opus`/`opus[1m]` map to 4.7).
func TestGetProfileText_OpusAliasValues(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		if !containsStr(txt.ModelOpus, "4.7") {
			t.Errorf("lang=%q: ModelOpus %q should reference Opus 4.7", lang, txt.ModelOpus)
		}
		if !containsStr(txt.ModelOpus1M, "4.7") {
			t.Errorf("lang=%q: ModelOpus1M %q should reference Opus 4.7", lang, txt.ModelOpus1M)
		}
		if !containsStr(txt.ModelOpus1M, "1M") {
			t.Errorf("lang=%q: ModelOpus1M %q should reference 1M context", lang, txt.ModelOpus1M)
		}
	}
}

// TestGetProfileText_EffortLevelValues verifies effort level labels contain xhigh/max keywords.
func TestGetProfileText_EffortLevelValues(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		if !containsStr(txt.EffortLevelXHigh, "xhigh") {
			t.Errorf("lang=%q: EffortLevelXHigh %q does not contain 'xhigh'", lang, txt.EffortLevelXHigh)
		}
		if !containsStr(txt.EffortLevelMax, "max") {
			t.Errorf("lang=%q: EffortLevelMax %q does not contain 'max'", lang, txt.EffortLevelMax)
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

// TestGetProfileText_PermAutoRuntimeWarning W-3: PermAuto 라벨에 런타임 오류 경고가
// 포함되어 있는지 언어별로 확인한다.
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

// TestGetProfileText_MigrationNoticeFields W-4: MigrationNotice 필드가 4개 언어에서
// 모두 채워져 있고 %q 형식 동사가 포함되어 있는지 확인한다.
func TestGetProfileText_MigrationNoticeFields(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		if txt.MigrationNoticeStatuslineMode == "" {
			t.Errorf("lang=%q: MigrationNoticeStatuslineMode is empty", lang)
		}
		if txt.MigrationNoticeStatuslineTheme == "" {
			t.Errorf("lang=%q: MigrationNoticeStatuslineTheme is empty", lang)
		}
		// %q 형식 동사가 2개 포함되어 있어야 한다 (이전 값, 새 값)
		if !containsStr(txt.MigrationNoticeStatuslineMode, "%q") {
			t.Errorf("lang=%q: MigrationNoticeStatuslineMode %q should contain %%q format verb", lang, txt.MigrationNoticeStatuslineMode)
		}
	}
}

// TestGetProfileText_SummarySyncSkippedNeutral W-5: SummarySyncSkipped가
// 중립적인 표현(프로젝트별 동기화 없음)을 사용하는지 확인한다.
func TestGetProfileText_SummarySyncSkippedNeutral(t *testing.T) {
	// "Sync skipped" 같은 오류성 표현이 없어야 한다
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		if txt.SummarySyncSkipped == "" {
			t.Errorf("lang=%q: SummarySyncSkipped is empty", lang)
		}
	}
	// en: 중립 메시지 확인
	en := getProfileText("en")
	if !containsStr(en.SummarySyncSkipped, "No project-level sync") {
		t.Errorf("en SummarySyncSkipped should be neutral, got: %q", en.SummarySyncSkipped)
	}
	// ko: 중립 메시지 확인
	ko := getProfileText("ko")
	if !containsStr(ko.SummarySyncSkipped, "프로젝트별 동기화 없음") {
		t.Errorf("ko SummarySyncSkipped should be neutral, got: %q", ko.SummarySyncSkipped)
	}
}

// TestGetProfileText_SummaryHeaderUpdated S-3: ko/ja SummaryHeader가 업데이트된
// 표현을 사용하는지 확인한다.
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
