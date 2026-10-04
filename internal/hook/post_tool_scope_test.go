package hook

// post_tool_scope_test.go — card t1499 M4.
//
// The PostToolUse Quality Gate and Security Guardian fired for throwaway
// scripts that live outside the project (a session scratchpad under
// /private/tmp, or /tmp). Both now apply only to files inside the project.
// "Inside" is judged after EvalSymlinks on both sides, because macOS /tmp is a
// symlink to /private/tmp and a spelling must not decide scope.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
)

// scopeFixture is a project directory, an unrelated outside directory, and a
// symlink spelling for each.
type scopeFixture struct {
	project, outside         string
	projectLink, outsideLink string
}

func newScopeFixture(t *testing.T) scopeFixture {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := scopeFixture{
		project:     filepath.Join(base, "project"),
		outside:     filepath.Join(base, "outside"),
		projectLink: filepath.Join(base, "project-link"),
		outsideLink: filepath.Join(base, "outside-link"),
	}
	for _, d := range []string{f.project, f.outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(f.project, f.projectLink); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink(f.outside, f.outsideLink); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	// Existing files, so EvalSymlinks resolves the full path.
	for _, p := range []string{f.project, f.outside} {
		if err := os.WriteFile(filepath.Join(p, "x.go"), []byte("package x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv(config.EnvClaudeProjectDir, f.project)
	return f
}

func writeInput(cwd, filePath, content string) *HookInput {
	raw, _ := json.Marshal(map[string]any{"file_path": filePath, "content": content})
	return &HookInput{
		SessionID:     "sess-scope",
		CWD:           cwd,
		HookEventName: "PostToolUse",
		ToolName:      "Write",
		ToolInput:     raw,
		ToolOutput:    json.RawMessage(`{"success": true}`),
	}
}

func TestPostToolTargetOutsideProject(t *testing.T) {
	f := newScopeFixture(t)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"file inside the project", filepath.Join(f.project, "x.go"), false},
		{"file outside the project", filepath.Join(f.outside, "x.go"), true},
		{"symlinked spelling of an inside file stays inside", filepath.Join(f.projectLink, "x.go"), false},
		{"symlinked spelling of an outside file stays outside", filepath.Join(f.outsideLink, "x.go"), true},
		{"not-yet-existing inside file", filepath.Join(f.project, "new", "y.go"), false},
		{"not-yet-existing outside file", filepath.Join(f.outside, "new", "y.go"), true},
		{"not-yet-existing file via inside symlink", filepath.Join(f.projectLink, "new", "y.go"), false},
		{"not-yet-existing file via outside symlink", filepath.Join(f.outsideLink, "new", "y.go"), true},
		{"relative path resolves against the project", "x.go", false},
		{"dot-dot escape out of the project", filepath.Join(f.project, "..", "outside", "x.go"), true},
		{"sibling directory sharing the project name prefix", f.project + "-sibling/x.go", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := postToolTargetOutsideProject(writeInput(f.project, tt.path, "package x\n"))
			if got != tt.want {
				t.Errorf("postToolTargetOutsideProject(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// With no usable root, or no usable path, the scope check never excludes: the
// previous behavior is kept rather than silently muting the scans.
func TestPostToolTargetOutsideProject_FailsOpen(t *testing.T) {
	t.Setenv(config.EnvClaudeProjectDir, "")
	outside := filepath.Join(t.TempDir(), "x.go")

	if postToolTargetOutsideProject(writeInput("", outside, "package x\n")) {
		t.Errorf("no project root known: must not exclude")
	}
	if postToolTargetOutsideProject(&HookInput{ToolName: "Write", ToolInput: json.RawMessage(`{not json`), CWD: "/x"}) {
		t.Errorf("unparseable tool input: must not exclude")
	}
	if postToolTargetOutsideProject(writeInput("/x", "", "package x\n")) {
		t.Errorf("empty file_path: must not exclude")
	}
}

// The session cwd also counts as inside: a lane works in a worktree whose root
// differs from CLAUDE_PROJECT_DIR (the primary checkout), and its files must
// stay scanned.
func TestPostToolTargetOutsideProject_SessionCwdCountsAsInside(t *testing.T) {
	f := newScopeFixture(t)
	if postToolTargetOutsideProject(writeInput(f.outside, filepath.Join(f.outside, "x.go"), "package x\n")) {
		t.Errorf("a file inside the session cwd must stay in scope")
	}
}

func errorCollector() *mockDiagnosticsCollector {
	return &mockDiagnosticsCollector{
		getDiagnosticsFunc: func(_ context.Context, _ string) ([]lsphook.Diagnostic, error) {
			return []lsphook.Diagnostic{{Message: "undefined: foo", Severity: lsphook.SeverityError}}, nil
		},
		getSeverityCountFunc: func(_ []lsphook.Diagnostic) lsphook.SeverityCounts {
			return lsphook.SeverityCounts{Errors: 1}
		},
	}
}

func TestPostToolQualityGateScopedToProject(t *testing.T) {
	f := newScopeFixture(t)

	tests := []struct {
		name      string
		path      string
		wantGated bool
	}{
		{"inside file keeps the Quality Gate", filepath.Join(f.project, "x.go"), true},
		{"symlinked inside file keeps the Quality Gate", filepath.Join(f.projectLink, "x.go"), true},
		{"outside file gets no Quality Gate", filepath.Join(f.outside, "x.go"), false},
		{"symlinked outside file gets no Quality Gate", filepath.Join(f.outsideLink, "x.go"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewPostToolHandlerWithDiagnostics(errorCollector())
			out, err := h.Handle(context.Background(), writeInput(f.project, tt.path, "package x\n"))
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			got := strings.Contains(out.SystemMessage, "[Quality Gate]")
			if got != tt.wantGated {
				t.Errorf("Quality Gate present = %v, want %v (systemMessage %q)", got, tt.wantGated, out.SystemMessage)
			}
		})
	}
}

func TestPostToolSecurityGuardianScopedToProject(t *testing.T) {
	f := newScopeFixture(t)
	// Assembled from two halves so this test file does not itself trip the
	// hardcoded-key scan it exists to exercise.
	secret := "api_key = \"sk-live-" + "abcdef0123456789\""

	tests := []struct {
		name        string
		path        string
		wantAdvisor bool
	}{
		{"inside file keeps the Security Guardian", filepath.Join(f.project, "x.py"), true},
		{"symlinked inside file keeps the Security Guardian", filepath.Join(f.projectLink, "x.py"), true},
		{"outside file gets no Security Guardian", filepath.Join(f.outside, "x.py"), false},
		{"symlinked outside file gets no Security Guardian", filepath.Join(f.outsideLink, "x.py"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := NewPostToolGuardianHandler().Handle(context.Background(), writeInput(f.project, tt.path, secret))
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			got := out != nil && out.HookSpecificOutput != nil &&
				strings.Contains(out.HookSpecificOutput.AdditionalContext, "[MoAI Security Guardian]")
			if got != tt.wantAdvisor {
				t.Errorf("Security Guardian present = %v, want %v", got, tt.wantAdvisor)
			}
		})
	}
}

// A `..` that follows a symlink component walks up from the link's real
// target, not from the link's parent: outside/alias/../x.go with alias pointing
// at project/sub is project/x.go. Cleaning the spelling before resolving links
// would judge it outside.
func TestPostToolTargetOutsideProject_DotDotAfterSymlink(t *testing.T) {
	f := newScopeFixture(t)
	sub := filepath.Join(f.project, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(f.outside, "alias")
	if err := os.Symlink(sub, alias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	for name, rel := range map[string]string{
		"existing file": "x.go",
		"new file":      "brand-new.go",
	} {
		t.Run(name, func(t *testing.T) {
			path := alias + string(filepath.Separator) + ".." + string(filepath.Separator) + rel
			if postToolTargetOutsideProject(writeInput("", path, "x")) {
				t.Errorf("%s resolves to %s and is inside the project, judged outside", path, filepath.Join(f.project, rel))
			}
		})
	}
}

// A relative file_path is relative to the hook input's cwd, not to the project
// root: with cwd=project/sub, "../x.go" is project/x.go.
func TestPostToolTargetOutsideProject_RelativePathUsesInputCwd(t *testing.T) {
	f := newScopeFixture(t)
	sub := filepath.Join(f.project, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if postToolTargetOutsideProject(writeInput(sub, "../x.go", "x")) {
		t.Error("../x.go with cwd=project/sub is project/x.go, judged outside")
	}
	// A relative path that climbs out of the project is still outside.
	if !postToolTargetOutsideProject(writeInput(sub, "../../outside/x.go", "x")) {
		t.Error("../../outside/x.go with cwd=project/sub is outside/x.go, judged inside")
	}
}
