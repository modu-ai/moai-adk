package hook

// post_tool_scope_test.go — card t1507, re-landing card t1499's reverted M4.
//
// The PostToolUse Quality Gate and Security Guardian fired for throwaway
// scripts that live outside the project (a session scratchpad under
// /private/tmp, or /tmp). Both now apply only to files inside the project.
// "Inside" is judged against the project root and the root of the worktree the
// session cwd belongs to, after symlink resolution on both sides, because
// macOS /tmp is a symlink to /private/tmp and a spelling must not decide
// scope. Card t1499's two card-review rounds each found a case where the
// first attempt skipped a scan it should have run; every such hole class has
// a named regression test here (WorktreeSiblingCountsAsInside,
// RelativeAliasChainStaysInside, CaseSpellingStaysInside).

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
)

// gitFileMarker stands in for the .git FILE a linked worktree carries, so the
// scope walk stops here deterministically instead of wandering into whatever
// repository $TMPDIR may sit under.
func markWorktreeRoot(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: /nonexistent\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// scopeFixture is a project directory (marked as a worktree root), an
// unrelated outside directory, and a symlink spelling for each.
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
	markWorktreeRoot(t, f.project)
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
	if postToolTargetOutsideProject(writeInput("/x", "relative.go", "package x\n")) {
		t.Errorf("relative path with no root to resolve against: must not exclude")
	}
}

// The scope basis is the root of the worktree the session cwd belongs to. A
// cwd under a worktree outside CLAUDE_PROJECT_DIR (an L2 tree under the home
// directory, for example) keeps every file of that worktree in scope —
// including siblings and parents of the cwd directory itself. This is the
// round-2 card-review hole that reverted card t1499's first attempt.
func TestPostToolTargetOutsideProject_WorktreeSiblingCountsAsInside(t *testing.T) {
	f := newScopeFixture(t)
	wt := filepath.Join(f.outside, "lane-worktree") // outside the project on purpose
	if err := os.MkdirAll(filepath.Join(wt, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	markWorktreeRoot(t, wt)
	if err := os.WriteFile(filepath.Join(wt, "other.go"), []byte("package wt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cwd := filepath.Join(wt, "sub")

	if postToolTargetOutsideProject(writeInput(cwd, filepath.Join(wt, "other.go"), "x")) {
		t.Errorf("sibling of the cwd inside the same worktree judged outside")
	}
	if postToolTargetOutsideProject(writeInput(cwd, filepath.Join(wt, "new.go"), "x")) {
		t.Errorf("not-yet-existing file inside the same worktree judged outside")
	}
	// A file in neither the project nor the cwd's worktree is still outside.
	if !postToolTargetOutsideProject(writeInput(cwd, filepath.Join(f.outside, "x.go"), "x")) {
		t.Errorf("file outside both the project and the cwd worktree judged inside")
	}
}

// A cwd that belongs to a worktree keeps that worktree in scope. A cwd that
// belongs to no git root at all (a scratchpad) contributes no root, so a file
// beside it stays outside — that is the noise this card removes.
func TestPostToolTargetOutsideProject_SessionCwdCountsAsInside(t *testing.T) {
	f := newScopeFixture(t)

	wt := filepath.Join(f.outside, "lane-worktree")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	markWorktreeRoot(t, wt)
	if err := os.WriteFile(filepath.Join(wt, "x.go"), []byte("package wt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if postToolTargetOutsideProject(writeInput(wt, filepath.Join(wt, "x.go"), "x")) {
		t.Errorf("a file inside the session cwd's worktree must stay in scope")
	}

	scratch := filepath.Join(f.outside, "scratchpad")
	if err := os.MkdirAll(scratch, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scratch, "tmp.sh"), []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !postToolTargetOutsideProject(writeInput(scratch, filepath.Join(scratch, "tmp.sh"), "x")) {
		t.Errorf("a scratchpad file beside a scratchpad cwd must stay out of scope")
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

// A relative path whose spelling climbs out of the project and back in through
// a symlink (alias at outside/alias pointing at project/sub) is project/chain.go:
// concatenating the spelling and resolving symlinks first sees the link, while
// a filepath.Join would Clean it away and judge the target outside. This is
// the second round-2 card-review hole.
func TestPostToolTargetOutsideProject_RelativeAliasChainStaysInside(t *testing.T) {
	f := newScopeFixture(t)
	sub := filepath.Join(f.project, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(f.outside, "alias")
	if err := os.Symlink(sub, alias); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	sep := string(filepath.Separator)
	rel := ".." + sep + ".." + sep + "outside" + sep + "alias" + sep + ".." + sep + "chain.go"
	if postToolTargetOutsideProject(writeInput(sub, rel, "x")) {
		t.Errorf("%s with cwd=project/sub resolves to project/chain.go, judged outside", rel)
	}
}

// The same directory spelled with different letter case (measured in this
// environment: the launcher prints /Users/goos/moai while the session reports
// /Users/goos/MoAI) must not decide scope on a case-insensitive filesystem.
// The fold only ever widens the inside set, so it can re-add a scan but never
// remove one.
func TestPostToolTargetOutsideProject_CaseSpellingStaysInside(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		t.Skip("case-insensitive path comparison applies to darwin/windows only")
	}
	f := newScopeFixture(t)
	upper := strings.ToUpper(filepath.Base(f.project))
	cased := filepath.Join(filepath.Dir(f.project), upper)

	if postToolTargetOutsideProject(writeInput(f.project, filepath.Join(cased, "x.go"), "x")) {
		t.Errorf("case-only spelling %q of the project judged outside", cased)
	}
	if !postToolTargetOutsideProject(writeInput(f.project, filepath.Join(filepath.Dir(f.project), "OUTSIDE", "x.go"), "x")) {
		t.Errorf("case-fold widening must not pull a genuinely outside directory inside")
	}
}

// The session cwd may itself be reached through a symlink (a scratchpad path
// pointing into the lane worktree): the worktree-root walk runs on the
// resolved cwd, so the worktree is still found and its files stay in scope.
func TestPostToolTargetOutsideProject_CwdSymlinkIntoWorktree(t *testing.T) {
	f := newScopeFixture(t)
	wt := filepath.Join(f.outside, "lane-worktree")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	markWorktreeRoot(t, wt)
	if err := os.WriteFile(filepath.Join(wt, "x.go"), []byte("package wt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(f.outside, "into-worktree")
	if err := os.Symlink(wt, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	// Absolute target beside the cwd, inside the same worktree.
	if postToolTargetOutsideProject(writeInput(link, filepath.Join(wt, "x.go"), "x")) {
		t.Errorf("cwd reached through a symlink into a worktree: sibling file judged outside")
	}
	// Relative target from the symlinked cwd climbs into the worktree root.
	if postToolTargetOutsideProject(writeInput(link, "../lane-worktree/x.go", "x")) {
		t.Errorf("relative path from a symlinked cwd judged outside")
	}
}

// A dangling symlink cannot prove where its target really is: a spelling that
// climbs out through an alias whose destination is gone is an ambiguous
// judgment, and the card's fail-closed condition keeps the scan on.
func TestPostToolTargetOutsideProject_DanglingAliasStaysInScope(t *testing.T) {
	f := newScopeFixture(t)
	dangling := filepath.Join(f.outside, "dangling")
	if err := os.Symlink(filepath.Join(f.outside, "gone"), dangling); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if postToolTargetOutsideProject(writeInput("", filepath.Join(dangling, "x.go"), "x")) {
		t.Errorf("dangling alias target is an ambiguous judgment, judged outside")
	}
}

func TestContainingWorktreeRoot(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(base, "repo") // .git as a directory
	if err := os.MkdirAll(filepath.Join(repo, "deep", "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	wt := filepath.Join(base, "wt") // .git as a file, the linked-worktree form
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	markWorktreeRoot(t, wt)
	plain := filepath.Join(base, "plain")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		dir           string
		want          string
		wantUncertain bool
	}{
		{"repo root itself", repo, repo, false},
		{"deep under the repo", filepath.Join(repo, "deep", "sub"), repo, false},
		{"worktree with a .git file", wt, wt, false},
		{"plain directory has no root", plain, "", false},
		{"empty input", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, uncertain := containingWorktreeRoot(tt.dir)
			if got != tt.want || uncertain != tt.wantUncertain {
				t.Errorf("containingWorktreeRoot(%q) = (%q, %v), want (%q, %v)", tt.dir, got, uncertain, tt.want, tt.wantUncertain)
			}
		})
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
			got := out != nil && strings.Contains(out.SystemMessage, "[Quality Gate]")
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
