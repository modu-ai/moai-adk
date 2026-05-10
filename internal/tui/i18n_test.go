// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestKoEnKeysParity verifies that ko.yaml and en.yaml have identical key sets.
func TestKoEnKeysParity(t *testing.T) {
	koKeys, err := tui.CatalogKeys("ko")
	if err != nil {
		t.Fatalf("CatalogKeys(ko): %v", err)
	}
	enKeys, err := tui.CatalogKeys("en")
	if err != nil {
		t.Fatalf("CatalogKeys(en): %v", err)
	}

	// Build sets.
	koSet := make(map[string]struct{}, len(koKeys))
	enSet := make(map[string]struct{}, len(enKeys))
	for _, k := range koKeys {
		koSet[k] = struct{}{}
	}
	for _, k := range enKeys {
		enSet[k] = struct{}{}
	}

	for _, k := range koKeys {
		if _, ok := enSet[k]; !ok {
			t.Errorf("ko key %q is missing from en catalog", k)
		}
	}
	for _, k := range enKeys {
		if _, ok := koSet[k]; !ok {
			t.Errorf("en key %q is missing from ko catalog", k)
		}
	}

	if len(koKeys) != len(enKeys) {
		t.Errorf("catalog key count mismatch: ko=%d en=%d", len(koKeys), len(enKeys))
	}
}

// TestTranslateLookup verifies canonical key lookups in ko and en catalogs.
func TestTranslateLookup(t *testing.T) {
	tests := []struct {
		key  string
		lang string
		want string
	}{
		// Korean strings
		{"doctor.summary.pass", "ko", "통과"},
		{"init.wizard.step", "ko", "단계"},
		{"error.generic", "ko", "오류가 발생했습니다"},
		{"loop.phase.plan", "ko", "계획"},
		{"loop.phase.impl", "ko", "구현"},
		{"loop.phase.sync", "ko", "동기화"},
		// English strings
		{"doctor.summary.pass", "en", "Pass"},
		{"init.wizard.step", "en", "Step"},
		{"error.generic", "en", "An error occurred"},
		{"loop.phase.plan", "en", "Plan"},
		{"loop.phase.impl", "en", "Impl"},
		{"loop.phase.sync", "en", "Sync"},
	}

	for _, tt := range tests {
		t.Run(tt.key+"@"+tt.lang, func(t *testing.T) {
			got := tui.Translate(tt.key, tt.lang)
			if got != tt.want {
				t.Errorf("Translate(%q, %q) = %q, want %q", tt.key, tt.lang, got, tt.want)
			}
		})
	}
}

// TestI18nFallbackToEn verifies that missing lang falls back to English.
func TestI18nFallbackToEn(t *testing.T) {
	// "zh" is an empty placeholder catalog (no keys); should fall back to "en".
	got := tui.Translate("doctor.summary.pass", "zh")
	if got == "" {
		t.Error("Translate with zh fallback returned empty string; want English value")
	}
	// "xx" is not a valid lang code; should also fall back.
	got2 := tui.Translate("init.wizard.step", "xx")
	if got2 == "" {
		t.Error("Translate with unknown lang returned empty string; want English value")
	}
}

// TestNoEmojiInProductionCatalog verifies ko.yaml and en.yaml contain no emoji codepoints.
func TestNoEmojiInProductionCatalog(t *testing.T) {
	for _, lang := range []string{"ko", "en"} {
		values, err := tui.CatalogValues(lang)
		if err != nil {
			t.Fatalf("CatalogValues(%s): %v", lang, err)
		}
		for key, val := range values {
			for _, r := range val {
				cp := uint32(r)
				if (cp >= 0x1F300 && cp <= 0x1FAFF) ||
					(cp >= 0x2600 && cp <= 0x26FF) ||
					(cp >= 0x2700 && cp <= 0x27BF) ||
					(cp >= 0x1F000 && cp <= 0x1F2FF) {
					t.Errorf("emoji U+%04X '%c' found in %s catalog key %q", cp, r, lang, key)
				}
			}
		}
	}
}
