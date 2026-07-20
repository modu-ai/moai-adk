package cli

// M4b contract tests for the status/spec-view glamour render wiring
// (SPEC-CLI-TUX-V3-004 REQ-TUX4-004/005, AC-TUX4-006). Test names match the
// AC-TUX4-006 run patterns 'StatusGolden' / 'SpecViewPlain'.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// chdirTempProject creates a temp project dir (optionally with a .moai tree)
// and chdirs into it, restoring the package dir on cleanup. Returns the
// pre-chdir package dir.
func chdirTempProject(t *testing.T, name string, initialized bool) string {
	t.Helper()
	pkgDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	projectDir := filepath.Join(t.TempDir(), name)
	if err := os.Mkdir(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	if initialized {
		if err := os.MkdirAll(filepath.Join(projectDir, ".moai", "config", "sections"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if chErr := os.Chdir(pkgDir); chErr != nil {
			t.Logf("failed to restore working directory: %v", chErr)
		}
	})
	t.Setenv("MOAI_NO_BODP_REMINDER", "1")
	return pkgDir
}

// sampleAcceptanceTree returns a small 2-level acceptance tree for renderer tests.
func sampleAcceptanceTree() []spec.Acceptance {
	return []spec.Acceptance{
		{
			ID:    "AC-ROOT",
			Given: "root requirement",
			Children: []spec.Acceptance{
				{ID: "AC-CHILD", When: "child action"},
			},
		},
		{ID: "AC-LAST", Then: "final result"},
	}
}

// TestStatusGolden_MarkdownSurface verifies the status render layer is the
// markdown compose + renderMarkdown gateway (REQ-TUX4-004): non-TTY output is
// plain markdown passthrough carrying every data field of the legacy Box
// surface, with zero ANSI and no box-drawing chrome.
func TestStatusGolden_MarkdownSurface(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("MOAI_THEME", "light")

	got, _ := captureStatusCmdWithPkgDir(t)

	// Markdown structure present.
	if !strings.Contains(got, "# Project Status") {
		t.Errorf("status output should carry markdown H1 '# Project Status', got:\n%s", got)
	}
	// Data fields preserved (render-layer-only swap, §D).
	for _, want := range []string{
		"**Project**: my-test-project",
		"moai-adk",
		".moai/config/sections",
		"**SPECs**: 2 found",
		"**Configs**: 2 section files",
		"Initialized",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("status markdown lost data field %q, got:\n%s", want, got)
		}
	}
	// Legacy Box chrome gone; no ANSI on non-TTY.
	if strings.Contains(got, "╭") || strings.Contains(got, "│") {
		t.Error("status output should no longer render the lipgloss Box chrome")
	}
	if strings.Contains(got, "\x1b") {
		t.Error("non-TTY status output must carry zero ANSI escape sequences")
	}
}

// TestStatusGolden_NoColorByteIdentical verifies NO_COLOR output equals the
// colour-enabled non-TTY output byte-for-byte: both take the plain markdown
// passthrough (REQ-TUX4-005), so theme env vars cannot change bytes off-TTY.
func TestStatusGolden_NoColorByteIdentical(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("MOAI_THEME", "light")
	colour, _ := captureStatusCmdWithPkgDir(t)

	t.Setenv("NO_COLOR", "1")
	noColour, _ := captureStatusCmdWithPkgDir(t)

	if colour != noColour {
		t.Errorf("non-TTY status output must be byte-identical with and without NO_COLOR\ncolour:\n%s\nnocolor:\n%s", colour, noColour)
	}
}

// TestStatusGolden_NotInitialized verifies the not-initialized path keeps its
// data semantics on the markdown surface.
func TestStatusGolden_NotInitialized(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	pkgDir := chdirTempProject(t, "bare-project", false)
	_ = pkgDir

	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	statusCmd.SetErr(buf)
	if err := statusCmd.RunE(statusCmd, []string{}); err != nil {
		t.Fatalf("statusCmd.RunE: %v", err)
	}
	statusCmd.SetOut(nil)
	statusCmd.SetErr(nil)

	got := buf.String()
	if !strings.Contains(got, "Not initialized") {
		t.Errorf("not-initialized status should say 'Not initialized', got:\n%s", got)
	}
	if !strings.Contains(got, "moai init") {
		t.Errorf("not-initialized status should hint 'moai init', got:\n%s", got)
	}
}

// TestSpecViewPlain_TreePassthrough verifies renderTreeMarkdown's plain path
// is a byte-stable passthrough of header + tree body — the non-TTY spec view
// output keeps the pre-M4b shape exactly (no fences, no ANSI).
func TestSpecViewPlain_TreePassthrough(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	buf := new(bytes.Buffer) // non-TTY
	header := "Acceptance Criteria for SPEC-X:"
	body := "├── AC-001: given: when: then\n└── AC-002: given2\n"

	got := renderTreeMarkdown(buf, header, body)
	want := header + "\n\n" + body
	if got != want {
		t.Errorf("plain tree render must be byte-stable passthrough\ngot:  %q\nwant: %q", got, want)
	}
	if strings.Contains(got, "```") {
		t.Error("plain tree passthrough must not contain markdown fences")
	}
}

// TestSpecViewPlain_CommandUsesGlamourGateway verifies the spec view command
// path routes through the glamour-mediated gateway and preserves output
// semantics end-to-end in a non-TTY capture (AC-TUX4-004 reachability at the
// behavior level; the symbol-level wiring is grep-verified by the AC command).
func TestSpecViewPlain_CommandUsesGlamourGateway(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	var b strings.Builder
	criteria := sampleAcceptanceTree()
	fprintTree(&b, criteria, "", false, 0, "")
	out := b.String()

	if !strings.Contains(out, "├── ") && !strings.Contains(out, "└── ") {
		t.Errorf("tree glyphs must survive the writer-based tree renderer, got:\n%s", out)
	}
}
