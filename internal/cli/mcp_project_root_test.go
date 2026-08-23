package cli

// SPEC-MCP-WORKTREE-ROOT-001 — the MCP tools whose subject is the caller's
// working tree accept an optional project_root, because resolveProjectDir()
// prefers CLAUDE_PROJECT_DIR and Claude Code sets that to the PRIMARY checkout
// even for a worktree session. Without the parameter, an audit issued from a
// worktree audits the primary and a worktree-only SPEC is silently absent.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// projectRootTools are the in-scope tools of REQ-1 that are registered on the
// MCP server surface. codex_audit and audit_multi are asserted separately (M2),
// because their assertion targets a params map rather than a catalogue result.
var projectRootSpecTools = []string{"spec_audit", "spec_drift", "spec_progress"}

// newProbeProject writes a minimal project tree carrying one SPEC, so a
// catalogue scan rooted there has something distinctive to find.
func newProbeProject(t *testing.T, specID string) string {
	t.Helper()
	root := t.TempDir()
	specDir := filepath.Join(root, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	body := "---\nid: " + specID + "\ntitle: \"probe\"\nversion: \"0.1.0\"\n" +
		"status: draft\ncreated: 2026-08-22\nupdated: 2026-08-22\nauthor: t\n" +
		"priority: P3\nphase: \"v3.1.3\"\nmodule: \"internal/cli\"\n" +
		"lifecycle: exploratory\ntags: \"probe\"\n---\n\n# " + specID + "\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Return the CANONICAL path. validateProjectRoot resolves symlinks, and
	// t.TempDir() hands back a symlinked path on macOS (/var → /private/var), so
	// a test comparing a handler's answer against this value would otherwise be
	// comparing two spellings of one directory and failing on the spelling.
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", root, err)
	}
	return canonical
}

// AC-REQ-1a (schema half): each in-scope SPEC tool declares the optional input.
func TestSpecTools_DeclareProjectRootInput(t *testing.T) {
	srv := newMoaiMCPServer()
	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	seen := map[string]bool{}
	for _, tool := range res.Tools {
		for _, want := range projectRootSpecTools {
			if tool.Name != want {
				continue
			}
			seen[want] = true
			b, _ := json.Marshal(tool.InputSchema)
			if !strings.Contains(string(b), "project_root") {
				t.Errorf("%s inputSchema does not declare project_root; schema=%s", want, b)
			}
		}
	}
	for _, want := range projectRootSpecTools {
		if !seen[want] {
			t.Errorf("tool %q not registered — cannot assert its schema", want)
		}
	}
}

// AC-REQ-1a (behaviour half, PRESENT direction): a catalogue scan rooted at a
// named tree finds a SPEC that exists only there.
func TestSpecAudit_ProjectRootRedirectsTheCatalogue(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBEONE-901")
	got := callSpecAuditJSON(t, map[string]any{"project_root": root})
	if !strings.Contains(got, "SPEC-PROBEONE-901") && !strings.Contains(got, "\"total_specs\":1") {
		t.Errorf("audit rooted at the named tree did not see its only SPEC; got=%s", got)
	}
}

// AC-REQ-1a (behaviour half, ABSENT direction). The same probe SPEC is NOT
// visible when the parameter is omitted, because resolution falls back to the
// default root. A test that only asserted the present direction would pass just
// as well if the parameter were ignored and the default tree happened to
// contain the SPEC — which is exactly how this class of defect hides.
func TestSpecAudit_WithoutProjectRootDoesNotSeeTheProbe(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBETWO-902")
	_ = root // deliberately NOT passed
	got := callSpecAuditJSON(t, map[string]any{})
	if strings.Contains(got, "SPEC-PROBETWO-902") {
		t.Errorf("audit without project_root saw a SPEC that exists only in the probe tree; got=%s", got)
	}
}

// AC-REQ-2: an absent or empty parameter resolves exactly as before.
func TestResolveToolProjectRoot_AbsentAndEmptyMatchTheDefault(t *testing.T) {
	want := resolveProjectDir()
	for _, args := range []map[string]any{{}, {"project_root": ""}, {"project_root": "   "}} {
		got, err := resolveToolProjectRoot(newToolRequest(args))
		if err != nil {
			t.Fatalf("args=%v: unexpected error: %v", args, err)
		}
		if got != want {
			t.Errorf("args=%v: root=%q, want the default %q", args, got, want)
		}
	}
}

// AC-REQ-3: a named root that cannot be a project is REJECTED, never silently
// replaced by the default. A fallback would return a caller who mistyped its
// own worktree path to auditing the primary — this SPEC's defect, arriving
// through the mechanism meant to fix it, and reporting success while it does.
func TestResolveToolProjectRoot_RejectsAnUnusableRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-dir")

	file := filepath.Join(t.TempDir(), "a-file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	noMoai := t.TempDir()

	for name, path := range map[string]string{
		"nonexistent":      missing,
		"file-not-dir":     file,
		"dir-without-moai": noMoai,
	} {
		got, err := resolveToolProjectRoot(newToolRequest(map[string]any{"project_root": path}))
		if err == nil {
			t.Errorf("%s: expected rejection, got root=%q", name, got)
			continue
		}
		if !strings.Contains(err.Error(), path) {
			t.Errorf("%s: error does not name the offending path %q: %v", name, path, err)
		}
		if got != "" {
			t.Errorf("%s: rejected root must be empty, got %q", name, got)
		}
	}
}

// A rejected root must surface as a tool error rather than an audit of the
// default tree — the same guarantee as above, asserted at the handler boundary
// where a caller actually observes it.
func TestSpecAudit_RejectedProjectRootDoesNotAuditTheDefault(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent")
	got := callSpecAuditJSON(t, map[string]any{"project_root": missing})
	if !strings.Contains(got, missing) {
		t.Errorf("handler did not surface the offending path; got=%s", got)
	}
	if strings.Contains(got, "total_specs") {
		t.Errorf("handler audited something despite an unusable project_root; got=%s", got)
	}
}

// The other two SPEC tools, exercised rather than schema-asserted.
//
// The sync audit found spec_progress and spec_drift resting on read-the-diff
// assurance: their declaration was checked, their parameter-PRESENT behaviour
// never was — not by a test and not by the live probe, which called only
// spec_audit. AC-1a names spec_audit, so this was not an AC miss; it was two of
// five in-scope tools nobody had watched work. The risk was low precisely
// because all three share resolveToolProjectRoot — but "it shares a covered
// helper" is an argument, not an observation.
func TestSpecProgress_ProjectRootRedirectsTheScanner(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBESCAN-905")

	withParam := callSpecToolJSON(t, handleSpecProgress, map[string]any{"project_root": root})
	if !strings.Contains(withParam, "SPEC-PROBESCAN-905") {
		t.Errorf("spec_progress rooted at the named tree did not list its only SPEC; got=%s", withParam)
	}

	// The absent direction, for the same reason AC-1a demands it: a
	// present-only assertion passes just as well if the parameter is ignored
	// and the default tree happens to hold the SPEC.
	without := callSpecToolJSON(t, handleSpecProgress, map[string]any{})
	if strings.Contains(without, "SPEC-PROBESCAN-905") {
		t.Errorf("spec_progress without project_root listed a SPEC that exists only in the probe tree; got=%s", without)
	}
}

func TestSpecDrift_ProjectRootRedirectsTheCatalogue(t *testing.T) {
	// Two SPECs rather than one, so the count that proves the redirect is a
	// number the ambient tree is unlikely to match by coincidence.
	root := newProbeProject(t, "SPEC-PROBEDRIFT-906")
	addProbeSpec(t, root, "SPEC-PROBEDRIFT-907")

	withParam := callSpecToolJSON(t, handleSpecDrift, map[string]any{"project_root": root})
	if !strings.Contains(withParam, `"total_specs":2`) {
		t.Errorf("spec_drift rooted at the named tree did not audit its two SPECs; got=%s", withParam)
	}

	without := callSpecToolJSON(t, handleSpecDrift, map[string]any{})
	if strings.Contains(without, `"total_specs":2`) {
		t.Errorf("spec_drift without project_root reported the probe tree's count, so the parameter was not what redirected it; got=%s", without)
	}
}

// --- helpers -------------------------------------------------------------

// addProbeSpec writes one more minimal SPEC into an existing probe tree.
func addProbeSpec(t *testing.T, root, specID string) {
	t.Helper()
	specDir := filepath.Join(root, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	body := "---\nid: " + specID + "\ntitle: \"probe\"\nversion: \"0.1.0\"\n" +
		"status: draft\ncreated: 2026-08-23\nupdated: 2026-08-23\nauthor: t\n" +
		"priority: P3\nphase: \"v3.1.3\"\nmodule: \"internal/cli\"\n" +
		"lifecycle: exploratory\ntags: \"probe\"\n---\n\n# " + specID + "\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// callSpecToolJSON invokes any SPEC-tool handler directly and returns its
// rendered content as one string.
func callSpecToolJSON(
	t *testing.T,
	handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error),
	args map[string]any,
) string {
	t.Helper()
	res, err := handler(context.Background(), newToolRequest(args))
	if err != nil {
		t.Fatalf("handler returned a hard error: %v", err)
	}
	b, _ := json.Marshal(res)
	return string(b)
}

func newToolRequest(args map[string]any) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Name = "spec_audit"
	req.Params.Arguments = args
	return req
}

// callSpecAuditJSON invokes the spec_audit handler directly and returns its
// rendered content as one string, so a test can assert on what a caller sees.
func callSpecAuditJSON(t *testing.T, args map[string]any) string {
	t.Helper()
	res, err := handleSpecAudit(context.Background(), newToolRequest(args))
	if err != nil {
		t.Fatalf("handleSpecAudit returned a hard error: %v", err)
	}
	b, _ := json.Marshal(res)
	return string(b)
}

// AC-t183 (boundary canonicalization): a project_root reached through a symlink
// resolves to the real directory, not to the symlink path.
//
// filepath.Abs does not resolve symlinks, so before this the helper returned the
// symlink path verbatim. Harmless while nothing compares the result against
// another path — and exactly the hole a future containment constraint (compare
// the resolved root against the repository's git common dir) would be walked
// through, because the two sides would be spelled differently for the same
// directory. Canonicalizing here is what keeps such a comparison meaningful
// rather than decorative.
func TestValidateProjectRoot_ResolvesSymlinkedRoot(t *testing.T) {
	real := newProbeProject(t, "SPEC-PROBELINK-906")
	realCanonical, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatalf("EvalSymlinks(real): %v", err)
	}

	link := filepath.Join(t.TempDir(), "linked-root")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable on this platform: %v", err)
	}

	got, err := resolveToolProjectRoot(newToolRequest(map[string]any{"project_root": link}))
	if err != nil {
		t.Fatalf("unexpected rejection of a symlinked root: %v", err)
	}
	if got != realCanonical {
		t.Errorf("root=%q, want the canonical target %q (symlink was %q)", got, realCanonical, link)
	}
}

// A symlink whose target does not exist is rejected, and the error names the
// path the caller actually supplied — not the resolved one it never typed.
func TestValidateProjectRoot_RejectsBrokenSymlink(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "dangling")
	if err := os.Symlink(filepath.Join(dir, "no-such-target"), link); err != nil {
		t.Skipf("symlinks unavailable on this platform: %v", err)
	}

	got, err := resolveToolProjectRoot(newToolRequest(map[string]any{"project_root": link}))
	if err == nil {
		t.Fatalf("expected rejection of a broken symlink, got root=%q", got)
	}
	if !strings.Contains(err.Error(), link) {
		t.Errorf("error does not name the supplied path %q: %v", link, err)
	}
	if got != "" {
		t.Errorf("rejected root must be empty, got %q", got)
	}
}
