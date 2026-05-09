// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// update controls golden snapshot regeneration. Set via -update flag.
var updateGolden = os.Getenv("UPDATE_GOLDEN") == "1"

// goldenPath returns the path to a golden snapshot file.
func goldenPath(name string) string {
	return filepath.Join("testdata", name+".golden")
}

// checkGolden compares got to a golden file, regenerating it if UPDATE_GOLDEN=1.
func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	path := goldenPath(name)
	if updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		t.Logf("updated golden: %s", path)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with UPDATE_GOLDEN=1 to generate)", path, err)
	}
	if got != string(want) {
		t.Errorf("box output mismatch for %s\ngot:\n%s\nwant:\n%s", name, got, want)
	}
}

// TestBoxBasic verifies Box() with a basic title and body in light theme.
func TestBoxBasic(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Box(tui.BoxOpts{
		Title: "Test Box",
		Body:  "Hello, World!",
		Width: 40,
		Theme: &th,
	})
	if out == "" {
		t.Fatal("Box() returned empty string")
	}
	checkGolden(t, "box-light-basic", out)
}

// TestBoxAccent verifies Box() with accent=true in light theme.
func TestBoxAccent(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Box(tui.BoxOpts{
		Title:  "Accent Box",
		Body:   "Accent content here",
		Width:  40,
		Theme:  &th,
		Accent: true,
	})
	if out == "" {
		t.Fatal("Box() accent returned empty string")
	}
	checkGolden(t, "box-light-accent", out)
}

// TestBoxDashed verifies Box() with dashed border in light theme.
func TestBoxDashed(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Box(tui.BoxOpts{
		Title:  "Dashed Box",
		Body:   "Dashed border content",
		Width:  40,
		Theme:  &th,
		Border: tui.BorderDashed,
	})
	if out == "" {
		t.Fatal("Box() dashed returned empty string")
	}
	checkGolden(t, "box-light-dashed", out)
}

// TestThickBoxAccent verifies ThickBox() with accent in light theme.
func TestThickBoxAccent(t *testing.T) {
	th := tui.LightTheme()
	out := tui.ThickBox(tui.BoxOpts{
		Title:  "Thick Accent",
		Body:   "Thick accent content",
		Width:  40,
		Theme:  &th,
		Accent: true,
	})
	if out == "" {
		t.Fatal("ThickBox() accent returned empty string")
	}
	checkGolden(t, "box-light-thick", out)
}

// TestBoxBasicDark verifies Box() basic in dark theme.
func TestBoxBasicDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.Box(tui.BoxOpts{
		Title: "Test Box",
		Body:  "Hello, World!",
		Width: 40,
		Theme: &th,
	})
	if out == "" {
		t.Fatal("Box() dark basic returned empty string")
	}
	checkGolden(t, "box-dark-basic", out)
}

// TestBoxAccentDark verifies Box() accent in dark theme.
func TestBoxAccentDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.Box(tui.BoxOpts{
		Title:  "Accent Box",
		Body:   "Accent content here",
		Width:  40,
		Theme:  &th,
		Accent: true,
	})
	if out == "" {
		t.Fatal("Box() dark accent returned empty string")
	}
	checkGolden(t, "box-dark-accent", out)
}

// TestBoxDashedDark verifies Box() dashed in dark theme.
func TestBoxDashedDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.Box(tui.BoxOpts{
		Title:  "Dashed Box",
		Body:   "Dashed border content",
		Width:  40,
		Theme:  &th,
		Border: tui.BorderDashed,
	})
	if out == "" {
		t.Fatal("Box() dark dashed returned empty string")
	}
	checkGolden(t, "box-dark-dashed", out)
}

// TestThickBoxAccentDark verifies ThickBox() accent in dark theme.
func TestThickBoxAccentDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.ThickBox(tui.BoxOpts{
		Title:  "Thick Accent",
		Body:   "Thick accent content",
		Width:  40,
		Theme:  &th,
		Accent: true,
	})
	if out == "" {
		t.Fatal("ThickBox() dark accent returned empty string")
	}
	checkGolden(t, "box-dark-thick", out)
}
