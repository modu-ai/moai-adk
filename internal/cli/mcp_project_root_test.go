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
	return root
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
		"nonexistent":   missing,
		"file-not-dir":  file,
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

// --- helpers -------------------------------------------------------------

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
