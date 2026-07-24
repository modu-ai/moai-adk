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

// TestCompactBanner_TwoLineIdentity verifies the compact band (bannerString) is
// a compact 1-2 line identity (REQ-TUX4-006): identity line + pill metadata row,
// nothing more. SPEC-CLI-TUX-INIT-UPDATE-001 M3 re-targets this from the composed
// PrintBanner surface to bannerString DIRECTLY: the restored logo stacks in
// PrintBanner, so the "compact band stays compact" intent is preserved at the
// bannerString layer (§A.1 L3 / §B R6 reversal-minimizing invariant).
func TestCompactBanner_TwoLineIdentity(t *testing.T) {
	pinBannerEnv(t)
	out := uikit.BannerString("1.2.3")
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	nonEmpty := 0
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty++
		}
	}
	if nonEmpty > 2 {
		t.Errorf("compact band (bannerString) must be 1-2 non-empty lines, got %d:\n%s", nonEmpty, out)
	}
}

// TestCompactBanner_NoASCIILogo verifies the large ASCII-art logo is absent from
// the compact band. M3 re-targets this to bannerString: the logo is restored,
// but it stacks ONLY in PrintBanner — the compact band (bannerString) stays
// logo-free (§A.1 L3 / §B R6). The logo's presence on the composed surface is
// asserted by TestPrintBanner_CarriesLogo.
func TestCompactBanner_NoASCIILogo(t *testing.T) {
	pinBannerEnv(t)
	out := uikit.BannerString("1.2.3")
	for _, glyph := range []string{"██", "╔", "╗", "╚", "╝", "═"} {
		if strings.Contains(out, glyph) {
			t.Errorf("compact band (bannerString) must not contain ASCII-logo glyph %q:\n%s", glyph, out)
		}
	}
}

// TestCompactBanner_GlyphWhitelist verifies the NO_COLOR compact band uses only
// ASCII plus the MoAI status-glyph vocabulary (plan §D whitelist: ✓ ✗ ! ● ○ ◆).
// M3 SCOPES this to bannerString (the compact band): the logo's decorative
// block/box-drawing runes (█ ╗ ╔ ╚ ╝ ═ ║) are a SEPARATE decorative category
// exempt from the status-glyph whitelist (REQ-TUXIU-056) and appear only in
// PrintBanner, never in bannerString — so asserting the whitelist against the
// band (not the composed surface) is the carve-out mechanism.
func TestCompactBanner_GlyphWhitelist(t *testing.T) {
	pinBannerEnv(t)
	out := uikit.BannerString("1.2.3")
	whitelist := map[rune]bool{'✓': true, '✗': true, '!': true, '●': true, '○': true, '◆': true}
	for _, r := range out {
		if r == '\n' || (r >= 0x20 && r < 0x7f) || whitelist[r] {
			continue
		}
		t.Errorf("compact band contains non-whitelisted glyph %q (U+%04X):\n%s", r, r, out)
	}
}

// TestPrintBanner_CarriesLogo verifies the headline reversal
// (SPEC-CLI-TUX-INIT-UPDATE-001 Group F, REQ-TUXIU-050/054, reversing
// SPEC-CLI-TUX-V3-004 REQ-TUX4-006): the composed PrintBanner surface DOES carry
// the restored 6-line MoAI-ADK logo stacked ABOVE the compact ◆ MoAI-ADK band.
// The logo's first-row signature (███╗   ███╗) is the greppable marker; the
// compact band is retained below it (both present).
func TestPrintBanner_CarriesLogo(t *testing.T) {
	pinBannerEnv(t)
	out, err := captureStdout(func() { uikit.PrintBanner("1.2.3") })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "███╗   ███╗") {
		t.Errorf("PrintBanner must carry the restored logo signature ███╗   ███╗, got:\n%s", out)
	}
	if !strings.Contains(out, "◆ MoAI-ADK") {
		t.Errorf("PrintBanner must retain the compact ◆ MoAI-ADK band below the logo, got:\n%s", out)
	}
}

// TestPrintLogo_EmitsLogoOnly verifies uikit.PrintLogo emits JUST the restored
// logo through the Printer/stdout gateway — no compact band. This is the surface
// the root-help predicate (cli/fang.go) prints before fang's help body, so fang
// renders its own header and the band must NOT be duplicated (REQ-TUXIU-055).
func TestPrintLogo_EmitsLogoOnly(t *testing.T) {
	pinBannerEnv(t)
	out, err := captureStdout(func() { uikit.PrintLogo() })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "███╗   ███╗") {
		t.Errorf("PrintLogo must emit the restored logo signature ███╗   ███╗, got:\n%s", out)
	}
	if strings.Contains(out, "◆ MoAI-ADK") {
		t.Errorf("PrintLogo must emit ONLY the logo, not the compact band, got:\n%s", out)
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
