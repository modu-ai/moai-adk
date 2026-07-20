package uikit_test

// M4d contract tests for the compact banner (SPEC-CLI-TUX-V3-004
// REQ-TUX4-006, AC-TUX4-007). Test names match the AC run patterns
// 'CompactBanner' / 'BannerPill'.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// pinBannerEnv pins the environment for deterministic banner output.
func pinBannerEnv(t *testing.T) {
	t.Helper()
	t.Setenv("NO_COLOR", "1")
	t.Setenv("CLAUDE_CODE_VERSION", "test-claude-99")
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.26.0")
}

// TestCompactBanner_TwoLineIdentity verifies the banner is a compact 1-2 line
// identity (REQ-TUX4-006): identity line + pill metadata row, nothing more.
func TestCompactBanner_TwoLineIdentity(t *testing.T) {
	pinBannerEnv(t)
	out, err := captureStdout(func() { uikit.PrintBanner("1.2.3") })
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	nonEmpty := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	if nonEmpty > 2 {
		t.Errorf("compact banner must be 1-2 non-empty lines, got %d:\n%s", nonEmpty, out)
	}
}

// TestCompactBanner_NoASCIILogo verifies the large ASCII-art logo is retired.
func TestCompactBanner_NoASCIILogo(t *testing.T) {
	pinBannerEnv(t)
	out, err := captureStdout(func() { uikit.PrintBanner("1.2.3") })
	if err != nil {
		t.Fatal(err)
	}
	for _, glyph := range []string{"██", "╔", "╗", "╚", "╝", "═"} {
		if strings.Contains(out, glyph) {
			t.Errorf("compact banner must not contain ASCII-logo glyph %q:\n%s", glyph, out)
		}
	}
}

// TestCompactBanner_GlyphWhitelist verifies the NO_COLOR banner uses only
// ASCII plus the MoAI glyph vocabulary (plan §D whitelist: ✓ ✗ ! ● ○ ◆).
func TestCompactBanner_GlyphWhitelist(t *testing.T) {
	pinBannerEnv(t)
	out, err := captureStdout(func() { uikit.PrintBanner("1.2.3") })
	if err != nil {
		t.Fatal(err)
	}
	whitelist := map[rune]bool{'✓': true, '✗': true, '!': true, '●': true, '○': true, '◆': true}
	for _, r := range out {
		if r == '\n' || (r >= 0x20 && r < 0x7f) || whitelist[r] {
			continue
		}
		t.Errorf("banner contains non-whitelisted glyph %q (U+%04X):\n%s", r, r, out)
	}
}

// TestCompactBanner_BrandTagline verifies the brand tagline survives the
// compaction (identity preservation, plan §D brand constraint).
func TestCompactBanner_BrandTagline(t *testing.T) {
	pinBannerEnv(t)
	out, err := captureStdout(func() { uikit.PrintBanner("1.2.3") })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "MoAI-ADK") {
		t.Errorf("banner must carry the MoAI-ADK identity, got:\n%s", out)
	}
	if !strings.Contains(out, "Agentic Development Kit") {
		t.Errorf("banner must carry the brand tagline, got:\n%s", out)
	}
}

// TestBannerPill_Metadata verifies the pill row carries the version / go /
// claude metadata triplet (REQ-TUX4-006) and the legacy "Version:" label line
// is gone.
func TestBannerPill_Metadata(t *testing.T) {
	pinBannerEnv(t)
	out, err := captureStdout(func() { uikit.PrintBanner("1.2.3") })
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"v1.2.3", "go 1.26.0", "test-claude-99"} {
		if !strings.Contains(out, want) {
			t.Errorf("banner pill row must carry %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Version:") {
		t.Errorf("legacy 'Version:' label line must be retired, got:\n%s", out)
	}
}

// TestBannerPill_ClaudeFallback verifies the claude pill degrades gracefully
// to the "claude" placeholder when CLAUDE_CODE_VERSION is unset
// (acceptance.md §C edge: no error / empty pill).
func TestBannerPill_ClaudeFallback(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("CLAUDE_CODE_VERSION", "")
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.26.0")
	out, err := captureStdout(func() { uikit.PrintBanner("1.2.3") })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "claude") {
		t.Errorf("claude pill must degrade to the 'claude' placeholder, got:\n%s", out)
	}
}
