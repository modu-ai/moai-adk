// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// --- KV ---

// TestKVBasic verifies KV produces a key-value row.
func TestKVBasic(t *testing.T) {
	th := tui.LightTheme()
	out := tui.KV("Status", "Running", tui.KVOpts{Theme: &th})
	if out == "" {
		t.Fatal("KV returned empty string")
	}
	if !strings.Contains(out, "Status") {
		t.Error("KV: key not in output")
	}
	if !strings.Contains(out, "Running") {
		t.Error("KV: value not in output")
	}
	checkGolden(t, "kv-light-basic", out)
}

// TestKVDark verifies KV in dark theme.
func TestKVDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.KV("Version", "v3.2.4", tui.KVOpts{Theme: &th})
	if out == "" {
		t.Fatal("KV dark returned empty string")
	}
	checkGolden(t, "kv-dark-basic", out)
}

// TestKVKorean verifies KV with Korean key.
func TestKVKorean(t *testing.T) {
	th := tui.LightTheme()
	out := tui.KV("상태", "실행 중", tui.KVOpts{Theme: &th})
	if out == "" {
		t.Fatal("KV korean returned empty string")
	}
	checkGolden(t, "kv-light-korean", out)
}

// --- CheckLine ---

// TestCheckLineOk verifies CheckLine with ok status.
func TestCheckLineOk(t *testing.T) {
	th := tui.LightTheme()
	out := tui.CheckLine("ok", "Go binary", "/usr/local/go/bin/go", "1.26", &th)
	if out == "" {
		t.Fatal("CheckLine ok returned empty string")
	}
	checkGolden(t, "checkline-light-ok", out)
}

// TestCheckLineDark verifies CheckLine in dark theme.
func TestCheckLineDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.CheckLine("ok", "Claude Code", "claude", "v1.0.18", &th)
	if out == "" {
		t.Fatal("CheckLine dark returned empty string")
	}
	checkGolden(t, "checkline-dark-ok", out)
}

// TestCheckLineWarn verifies CheckLine with warn status.
func TestCheckLineWarn(t *testing.T) {
	th := tui.LightTheme()
	out := tui.CheckLine("warn", "Glamour cache", "stale", "갱신 권장", &th)
	if out == "" {
		t.Fatal("CheckLine warn returned empty string")
	}
	checkGolden(t, "checkline-light-warn", out)
}

// TestCheckLineErr verifies CheckLine with err status.
func TestCheckLineErr(t *testing.T) {
	th := tui.LightTheme()
	out := tui.CheckLine("err", "GitHub CLI", "not found", "", &th)
	if out == "" {
		t.Fatal("CheckLine err returned empty string")
	}
	checkGolden(t, "checkline-light-err", out)
}

// --- Section ---

// TestSectionBasic verifies Section produces a titled section header.
func TestSectionBasic(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Section("시스템", tui.SectionOpts{Theme: &th})
	if out == "" {
		t.Fatal("Section returned empty string")
	}
	if !strings.Contains(out, "시스템") {
		t.Error("Section: title not in output")
	}
	checkGolden(t, "section-light-basic", out)
}

// TestSectionDark verifies Section in dark theme.
func TestSectionDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.Section("MoAI-ADK", tui.SectionOpts{Theme: &th})
	if out == "" {
		t.Fatal("Section dark returned empty string")
	}
	checkGolden(t, "section-dark-basic", out)
}

// --- 18 Mixed Ko-En Cases (AC-CLI-TUI-007) ---
//
// TestRenderMixedKoEn runs all 18 mixed Korean+English string cases through
// KV rendering and verifies golden snapshots for both light and dark themes.
// Ref: acceptance.md §AC-CLI-TUI-007 검증 케이스
func TestRenderMixedKoEn(t *testing.T) {
	// The 18 cases from acceptance.md verbatim.
	cases := []struct {
		id    int
		label string
		value string
	}{
		{1, "환경 감지", "전부 한글"},
		{2, "Label", "Layer 1"},
		{3, "환경 감지 Layer 1", "한글 + 영문"},
		{4, "환경 감지 · 환경 변수 우선", "가운뎃점 포함"},
		{5, "moai init my-app", "영문 + 코드"},
		{6, "v3.2.4 · 2026-04-28", "코드 + 숫자"},
		{7, "통과 9 · 주의 1 · 실패 0", "한글 + 숫자"},
		{8, "feat/SPEC-AUTH-007", "영문/슬래시/하이픈"},
		{9, "8 ahead · 0 behind origin/main", "영한 mixed"},
		{10, "/usr/local/bin/moai", "전부 영문 코드"},
		{11, "TS · React · Next", "3 token 한가운뎃점"},
		{12, "한 줄에 한자漢字 섞음", "한글 + 한자"},
		{13, "Pretendard · JetBrains Mono", "영영 font names"},
		{14, "MoAI-ADK · 모두의AI", "조합"},
		{15, "↑↓ 이동 · enter 선택", "특수 + 한 + 영"},
		{16, "yuna@air ~/work/my-app", "호스트 prompt"},
		{17, "✓ 통과 · ! 주의 · ✗ 실패", "기호 + 한글"},
		{18, "긴 한글 텍스트가 박스에 들어갈 때 줄바꿈", "한 줄 70+ chars"},
	}

	themes := []struct {
		name string
		fn   func() tui.Theme
	}{
		{"light", tui.LightTheme},
		{"dark", tui.DarkTheme},
	}

	for _, th := range themes {
		th := th
		for _, c := range cases {
			c := c
			t.Run(fmt.Sprintf("%s-%02d", th.name, c.id), func(t *testing.T) {
				theme := th.fn()
				out := tui.KV(c.label, c.value, tui.KVOpts{Theme: &theme})
				if out == "" {
					t.Fatalf("KV returned empty string for case %d (%s)", c.id, th.name)
				}
				name := fmt.Sprintf("mixed-%s-%02d", th.name, c.id)
				checkGolden(t, name, out)
			})
		}
	}
}
