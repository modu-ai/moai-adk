package template

import (
	"io/fs"
	"strings"
	"testing"
)

// statusLineTemplatePath is the embedded path of the statusline wrapper template,
// rendered to ".moai/status_line.sh" at init/update time.
const statusLineTemplatePath = ".moai/status_line.sh"

// renderStatusLine renders the real status_line.sh template from the embedded FS
// with the supplied context. It is the shared helper for the fallback-chain tests.
//
// SPEC-STATUSLINE-WINDOWS-FALLBACK-001 (issue modu-ai/moai-adk#1068).
func renderStatusLine(t *testing.T, ctx *TemplateContext) string {
	t.Helper()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}
	// The on-disk template carries the ".tmpl" suffix; the deployer strips it.
	if _, err := fs.ReadFile(fsys, statusLineTemplatePath+".tmpl"); err != nil {
		t.Fatalf("status_line template not found in embedded FS: %v", err)
	}

	r := NewRenderer(fsys)
	out, err := r.Render(statusLineTemplatePath+".tmpl", ctx)
	if err != nil {
		t.Fatalf("Render(%s) error: %v", statusLineTemplatePath, err)
	}
	return string(out)
}

// TestStatusLineNoBakedGoBinPath is the M1 reproduction test for issue #1068.
//
// REQ-SWF-004 / AC-SWF-001: the statusline template MUST NOT bake the
// .GoBinPath template variable as an absolute path. On a Windows installer
// machine `gobin.Detect` returns a phantom `C:\Users\{user}\go\bin` directory
// (go env GOPATH default, no os.Stat), so a baked branch resolves to a path
// that does not exist and the statusline silently renders empty.
//
// RED: this FAILS against the pre-fix template (which emits
// `{{posixPath .GoBinPath}}/moai`). GREEN after the M3 template rewrite.
func TestStatusLineNoBakedGoBinPath(t *testing.T) {
	t.Parallel()

	// A non-existent GoBinPath simulates the Windows-installer phantom path.
	ctx := NewTemplateContext(
		WithGoBinPath(`C:\Users\phantom\go\bin`),
		WithResolvedMoaiPath(`/opt/moai/bin/moai`),
	)
	rendered := renderStatusLine(t, ctx)

	// The rendered phantom Windows path (posix-converted) must NOT appear: the
	// baked-.GoBinPath branch is removed entirely (§14 compliance).
	if strings.Contains(rendered, "C:/Users/phantom/go/bin") {
		t.Errorf("status_line.sh contains baked phantom GoBinPath branch (§14 violation):\n%s", rendered)
	}
}

// TestStatusLineResolvedExecutableBranch verifies the guarded resolved-executable
// branch is present and positioned correctly.
//
// REQ-SWF-002 / REQ-SWF-003 / AC-SWF-002b / AC-SWF-005: when ResolvedMoaiPath is
// non-empty, the rendered script contains a `[ -f "<path>" ]`-guarded branch
// referencing that path, positioned after `command -v moai` and before the
// $HOME-relative branches.
//
// RED: FAILS against the pre-fix template (no resolved-executable branch exists).
func TestStatusLineResolvedExecutableBranch(t *testing.T) {
	t.Parallel()

	const resolved = "/opt/moai/bin/moai"
	ctx := NewTemplateContext(WithResolvedMoaiPath(resolved))
	rendered := renderStatusLine(t, ctx)

	// 1. A guarded branch referencing the resolved path must be present.
	guard := `[ -f "` + resolved + `" ]`
	if !strings.Contains(rendered, guard) {
		t.Errorf("status_line.sh missing guarded resolved-executable branch %q:\n%s", guard, rendered)
	}

	// 2. The exec line for the resolved path must be present and quoted.
	execLine := `exec "` + resolved + `" statusline`
	if !strings.Contains(rendered, execLine) {
		t.Errorf("status_line.sh missing resolved-executable exec line %q:\n%s", execLine, rendered)
	}

	// 3. Ordering: resolved-executable branch comes AFTER `command -v moai`
	//    and BEFORE the $HOME/go/bin branch.
	idxPATH := strings.Index(rendered, "command -v moai")
	idxResolved := strings.Index(rendered, resolved)
	idxHomeGoBin := strings.Index(rendered, "$HOME/go/bin/moai")
	if idxPATH < 0 {
		t.Fatalf("status_line.sh missing `command -v moai` PATH stage:\n%s", rendered)
	}
	if idxHomeGoBin < 0 {
		t.Fatalf("status_line.sh missing `$HOME/go/bin/moai` branch:\n%s", rendered)
	}
	if idxPATH >= idxResolved || idxResolved >= idxHomeGoBin {
		t.Errorf("resolved-executable branch out of order: command-v(%d) < resolved(%d) < $HOME/go/bin(%d) expected\n%s",
			idxPATH, idxResolved, idxHomeGoBin, rendered)
	}
}

// TestStatusLineEmptyResolvedPathOmitsBranch verifies graceful degradation.
//
// REQ-SWF-005 / AC-SWF-004: when ResolvedMoaiPath is empty (os.Executable()
// errored at init/update), the resolved-executable branch is OMITTED entirely
// (no empty-valued `exec` line), and the script falls through to the
// $HOME-relative guarded branches.
func TestStatusLineEmptyResolvedPathOmitsBranch(t *testing.T) {
	t.Parallel()

	ctx := NewTemplateContext() // ResolvedMoaiPath defaults to ""
	rendered := renderStatusLine(t, ctx)

	// No empty-valued guard / exec line: `[ -f "" ]` or `exec "" statusline`.
	if strings.Contains(rendered, `[ -f "" ]`) {
		t.Errorf("status_line.sh emits empty-valued resolved-executable guard:\n%s", rendered)
	}
	if strings.Contains(rendered, `exec "" statusline`) {
		t.Errorf("status_line.sh emits empty-valued resolved-executable exec line:\n%s", rendered)
	}

	// The $HOME-relative branches must still be present (fall-through path).
	if !strings.Contains(rendered, "$HOME/go/bin/moai") {
		t.Errorf("status_line.sh missing $HOME/go/bin/moai fall-through branch:\n%s", rendered)
	}
	if !strings.Contains(rendered, "$HOME/.local/bin/moai") {
		t.Errorf("status_line.sh missing $HOME/.local/bin/moai fall-through branch:\n%s", rendered)
	}
}

// TestStatusLineEveryExecGuarded verifies the no-phantom-exec invariant.
//
// REQ-SWF-001 / REQ-SWF-006 / AC-SWF-003: every `exec ... moai statusline` line
// is reached only after a `command -v moai` PATH match or a `[ -f ... ]`
// existence guard — so a non-existent candidate path never causes a phantom exec.
func TestStatusLineEveryExecGuarded(t *testing.T) {
	t.Parallel()

	ctx := NewTemplateContext(WithResolvedMoaiPath("/opt/moai/bin/moai"))
	rendered := renderStatusLine(t, ctx)

	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "exec ") || !strings.Contains(trimmed, "statusline") {
			continue
		}
		// The PATH exec (`exec moai statusline`) is guarded by the preceding
		// `command -v moai` test; every other exec must be inside a `[ -f ]` if-block.
		if trimmed == `exec moai statusline < "$temp_file"` {
			continue // guarded by `command -v moai &> /dev/null`
		}
		// Look back for the nearest guarding `[ -f "..." ]` within the if-block.
		guarded := false
		for j := i - 1; j >= 0 && j >= i-3; j-- {
			if strings.Contains(lines[j], `[ -f "`) {
				guarded = true
				break
			}
		}
		if !guarded {
			t.Errorf("unguarded exec line at %d: %q\n%s", i+1, trimmed, rendered)
		}
	}
}
