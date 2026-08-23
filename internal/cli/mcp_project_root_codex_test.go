package cli

// SPEC-MCP-WORKTREE-ROOT-001 M2 — project_root on the codex path and through
// audit_multi's fan-out.
//
// These tests assert on the codex PARAMS MAP, not on the fan-out seam. The
// distinction is the milestone's exit condition: a test that swaps backendCall
// and checks what the double received is satisfied while performCodexAudit —
// where the codex params are actually built — is never entered. The root could
// stop at the seam with every criterion green. Pinning the assertion to the
// params map closes that gap (AC-1b).
//
// Both codex paths are asserted, because they had diverged: the single-backend
// handler has always passed a `cwd`, while the fan-out passed none at all.

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// projectRootCodexTools are the M2 surfaces of REQ-1.
var projectRootCodexTools = []string{"codex_audit", "audit_multi"}

// capturedCodexParams records the params map every codex audit path hands the
// RPC layer, with no live backend behind it. Concurrent because audit_multi
// fans out through errgroup.
type capturedCodexParams struct {
	mu    sync.Mutex
	calls []map[string]any
}

func (c *capturedCodexParams) rpc(_ context.Context, _, _ string, params map[string]any) (ReviewOutput, error) {
	c.mu.Lock()
	// Copy: the caller owns the map and may reuse it.
	clone := map[string]any{}
	for k, v := range params {
		clone[k] = v
	}
	c.calls = append(c.calls, clone)
	c.mu.Unlock()
	return ReviewOutput{Verdict: "pass", Summary: "captured", Findings: []Finding{}, NextSteps: []string{}}, nil
}

func (c *capturedCodexParams) only(t *testing.T) map[string]any {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.calls) != 1 {
		t.Fatalf("codex RPC invoked %d times, want exactly 1", len(c.calls))
	}
	return c.calls[0]
}

func (c *capturedCodexParams) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.calls)
}

// withCapturedCodexParams swaps BOTH codex seams: the binary lookup (so no
// codex need exist) and the RPC (so nothing is executed and the params map is
// observable).
func withCapturedCodexParams(t *testing.T) *capturedCodexParams {
	t.Helper()
	probe := &capturedCodexParams{}
	prevLook, prevRPC := codexLookPath, codexReviewRPC
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	codexReviewRPC = probe.rpc
	t.Cleanup(func() { codexLookPath, codexReviewRPC = prevLook, prevRPC })
	return probe
}

// callToolCodexAudit invokes the single-backend handler directly.
func callToolCodexAudit(t *testing.T, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = "codex_audit"
	req.Params.Arguments = args
	res, err := handleCodexAudit(context.Background(), req)
	if err != nil {
		t.Fatalf("handleCodexAudit returned a hard error: %v", err)
	}
	return res
}

// codexOnlyGates runs the fan-out with codex required and GLM off, so the only
// backend invoked is the one whose params map is under assertion.
var codexOnlyGates = map[string]any{"claude": "required", "codex": "required", "glm": "off"}

// AC-1a (schema half, M2 surfaces): both codex-carrying tools declare the input.
func TestCodexTools_DeclareProjectRootInput(t *testing.T) {
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
		for _, want := range projectRootCodexTools {
			if tool.Name != want {
				continue
			}
			seen[want] = true
			b, _ := json.Marshal(tool.InputSchema)
			if !strings.Contains(string(b), projectRootArg) {
				t.Errorf("%s inputSchema does not declare project_root; schema=%s", want, b)
			}
		}
	}
	for _, want := range projectRootCodexTools {
		if !seen[want] {
			t.Errorf("tool %q not registered — cannot assert its schema", want)
		}
	}
}

// AC-1b, single-backend path: the named tree lands in the params map codex
// receives — asserted on the map itself, with no live backend.
func TestCodexAudit_ProjectRootLandsInTheParamsMap(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBECODEX-903")
	cap := withCapturedCodexParams(t)

	callToolCodexAudit(t, map[string]any{"project_root": root})

	got, _ := cap.only(t)["cwd"].(string)
	if got != root {
		t.Errorf("codex params cwd = %q, want the named tree %q", got, root)
	}
}

// AC-2, single-backend path: this handler ALREADY passed a cwd before the
// parameter existed, so an absent parameter must keep passing that same value.
func TestCodexAudit_WithoutProjectRootKeepsTodaysCwd(t *testing.T) {
	cap := withCapturedCodexParams(t)

	callToolCodexAudit(t, map[string]any{})

	got, _ := cap.only(t)["cwd"].(string)
	if want := resolveProjectDir(); got != want {
		t.Errorf("codex params cwd = %q, want the unchanged default %q", got, want)
	}
}

// AC-3, single-backend path: an unusable root is a tool error naming the path,
// and codex is never reached. A silent fallback here would review the primary
// checkout while reporting success — this SPEC's defect, arriving through the
// mechanism meant to fix it.
func TestCodexAudit_RejectsAnUnusableProjectRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-tree")
	cap := withCapturedCodexParams(t)

	res := callToolCodexAudit(t, map[string]any{"project_root": missing})

	if !res.IsError {
		t.Error("result is not an error; an unusable project_root must be rejected")
	}
	b, _ := json.Marshal(res)
	if !strings.Contains(string(b), missing) {
		t.Errorf("error does not name the offending path %q; got=%s", missing, b)
	}
	if n := cap.count(); n != 0 {
		t.Errorf("codex was invoked %d times despite a rejected project_root; want 0", n)
	}
}

// AC-1b, audit_multi fan-out path — the assertion the whole card came from.
// performCodexAudit carried NO cwd key at all before this milestone, so this
// checks that the key now exists AND carries the named tree.
func TestAuditMulti_ProjectRootLandsInTheCodexParamsMap(t *testing.T) {
	root := newProbeProject(t, "SPEC-PROBEMULTI-904")
	cap := withCapturedCodexParams(t)

	_, err := callToolAuditMulti(t,
		map[string]any{"verdict": "pass", "summary": "claude:pass"},
		map[string]any{"target": "uncommittedChanges", "gates": codexOnlyGates, "project_root": root},
	)
	if err != nil {
		t.Fatalf("handleAuditMulti returned a hard error: %v", err)
	}

	params := cap.only(t)
	cwd, ok := params["cwd"].(string)
	if !ok {
		t.Fatalf("fan-out codex params carry no cwd key at all; params=%v", params)
	}
	if cwd != root {
		t.Errorf("fan-out codex params cwd = %q, want the named tree %q", cwd, root)
	}
}

// AC-2, audit_multi fan-out path: this path passed NO cwd before, so an absent
// parameter must keep passing none. Substituting a default here would change
// what an existing caller's backend receives — the one thing REQ-2 forbids.
func TestAuditMulti_WithoutProjectRootAddsNoCwd(t *testing.T) {
	cap := withCapturedCodexParams(t)

	if _, err := callToolAuditMulti(t,
		map[string]any{"verdict": "pass", "summary": "claude:pass"},
		map[string]any{"target": "uncommittedChanges", "gates": codexOnlyGates},
	); err != nil {
		t.Fatalf("handleAuditMulti returned a hard error: %v", err)
	}

	if _, present := cap.only(t)["cwd"]; present {
		t.Error("fan-out codex params gained a cwd key with no project_root supplied; today's behaviour was to pass none")
	}
}

// AC-3, audit_multi fan-out path: rejection rather than a fan-out over the
// default tree. audit_multi is otherwise fail-open — fail-open covers an absent
// or broken BACKEND, not a caller input the caller can correct.
func TestAuditMulti_RejectsAnUnusableProjectRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-tree")
	cap := withCapturedCodexParams(t)

	res, err := callToolAuditMulti(t,
		map[string]any{"verdict": "pass", "summary": "claude:pass"},
		map[string]any{"target": "uncommittedChanges", "gates": codexOnlyGates, "project_root": missing},
	)
	if err != nil {
		t.Fatalf("handleAuditMulti returned a hard error: %v", err)
	}
	if !res.IsError {
		t.Error("result is not an error; an unusable project_root must be rejected")
	}
	b, _ := json.Marshal(res)
	if !strings.Contains(string(b), missing) {
		t.Errorf("error does not name the offending path %q; got=%s", missing, b)
	}
	if n := cap.count(); n != 0 {
		t.Errorf("a backend ran %d times despite a rejected project_root; want 0", n)
	}
}

// The seam itself carries the root. This is NOT a substitute for the params-map
// assertions above — it is the complement: those prove the root reaches codex,
// this proves the fan-out hands it to every backend it invokes, including ones
// whose own params map is out of this milestone's scope.
func TestRunMultiAudit_FanOutCarriesTheProjectRoot(t *testing.T) {
	rc := &recordingCaller{}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	const root = "/some/worktree"
	cfg := MultiAuditConfig{ProjectRoot: root}
	runMultiAudit(context.Background(), claudeReview("pass"), "uncommittedChanges", "", cfg, nil)

	rc.mu.Lock()
	defer rc.mu.Unlock()
	if len(rc.calls) == 0 {
		t.Fatal("no backend was invoked — nothing to assert the root on")
	}
	for _, c := range rc.calls {
		if c.ProjectRoot != root {
			t.Errorf("backend %s received project_root %q, want %q", c.Backend, c.ProjectRoot, root)
		}
	}
}

// The independence invariant, restated after the widening: the seam takes five
// strings and returns a verdict — it never RECEIVES one. This declaration is
// the assertion; it stops compiling if a future edit threads claude_verdict in.
var _ backendCallFn = func(_ context.Context, _, _, _, _ string) ReviewOutput {
	return ReviewOutput{}
}
