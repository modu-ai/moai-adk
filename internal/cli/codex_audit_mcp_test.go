//go:build !windows

// codex_audit_mcp_test.go — the MCP route to the Codex audit launcher: the
// codex_role_audit tool family is registered, confines its worktree root and
// destination exactly as the launcher core does (same function), runs the
// audit as a background job observable through status and result, and the
// tool-catalogue rule states the registered count.
package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	mcpcat "github.com/modu-ai/moai-adk/internal/mcp"
)

// callRoleAuditTool calls one codex_role_audit* tool on a fresh in-process server.
func callRoleAuditTool(t *testing.T, c *client.Client, name string, args map[string]any) (*mcp.CallToolResult, map[string]any) {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	res, err := c.CallTool(context.Background(), req)
	if err != nil {
		t.Fatalf("%s: transport error %v", name, err)
	}
	var m map[string]any
	if res.StructuredContent != nil {
		b, _ := json.Marshal(res.StructuredContent)
		_ = json.Unmarshal(b, &m)
	}
	return res, m
}

func newRoleAuditClient(t *testing.T) *client.Client {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	c, err := client.NewInProcessClient(newMoaiMCPServer())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeInProcessClient(c) })
	if _, err := c.Initialize(context.Background(), mcp.InitializeRequest{}); err != nil {
		t.Fatal(err)
	}
	return c
}

func roleAuditResultText(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func TestCodexAuditMCPTool(t *testing.T) {
	// 1. Registration and annotations: start writes, status and result read.
	annotations := listToolsWithAnnotations(t)
	for name, wantRO := range map[string]bool{
		codexRoleAuditToolName:       false,
		codexRoleAuditStatusToolName: true,
		codexRoleAuditResultToolName: true,
	} {
		ro, ok := annotations[name]
		if !ok {
			t.Fatalf("%s is not registered", name)
		}
		if ro != wantRO {
			t.Errorf("%s read-only hint = %v, want %v", name, ro, wantRO)
		}
	}

	// 2. The tool-catalogue rule and its template mirror state the registered count.
	registered := len(annotations)
	if registered != len(mcpcat.MoaiMCPToolNames()) {
		t.Fatalf("registered %d tools, catalog declares %d", registered, len(mcpcat.MoaiMCPToolNames()))
	}
	reTotal := regexp.MustCompile(`(\d+) tools exposed by the self-hosted`)
	for _, p := range projectRootDocFiles {
		body, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		m := reTotal.FindStringSubmatch(string(body))
		if m == nil {
			t.Fatalf("%s carries no total-count sentence", p)
		}
		if n, _ := strconv.Atoi(m[1]); n != registered {
			t.Errorf("%s says %d tools; %d are registered", p, n, registered)
		}
		for _, name := range []string{codexRoleAuditToolName, codexRoleAuditStatusToolName, codexRoleAuditResultToolName} {
			if !strings.Contains(string(body), "`"+name+"`") {
				t.Errorf("%s does not name %s", p, name)
			}
		}
	}

	// 3. Confinement: the server's own worktree is the only acceptable root.
	repo := newAuditRepo(t)
	fake := installFakeCodex(t)
	orig := codexRoleAuditServerDir
	codexRoleAuditServerDir = func() (string, error) { return repo.a1, nil }
	t.Cleanup(func() { codexRoleAuditServerDir = orig })
	c := newRoleAuditClient(t)

	u := filepath.Join(repo.base, "U")
	b := filepath.Join(repo.base, "B")
	b1 := filepath.Join(repo.base, "B1")
	for _, d := range []string{u, b} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	auditGit(t, b, "init", "-q", "-b", "main")
	auditGit(t, b, "commit", "-q", "--allow-empty", "-m", "init")
	auditGit(t, b, "worktree", "add", "-q", "-b", "wt-b", b1)
	l := filepath.Join(repo.base, "L")
	if err := os.Symlink(b, l); err != nil {
		t.Fatal(err)
	}
	for _, r := range []string{u, b1} {
		installAuditRoles(t, r)
	}
	legalOut := ".moai/reports/mcp/v.md"
	rejected := map[string]map[string]any{
		"primary checkout":     {"worktree_root": repo.a, "out": legalOut},
		"sibling worktree":     {"worktree_root": repo.a2, "out": legalOut},
		"unregistered dir":     {"worktree_root": u, "out": legalOut},
		"other repository":     {"worktree_root": b1, "out": legalOut},
		"symlink to other":     {"worktree_root": l, "out": legalOut},
		"outside report tree":  {"worktree_root": repo.a1, "out": "AGENTS.md"},
		".git component":       {"worktree_root": repo.a1, "out": ".moai/reports/x/.git/v.md"},
		"record directory":     {"worktree_root": repo.a1, "out": ".moai/reports/codex-audit/v.md"},
		"escape by dot-dot":    {"worktree_root": repo.a1, "out": ".moai/reports/../../v.md"},
		"not a read-only role": {"worktree_root": repo.a1, "out": legalOut, "role": "manager-docs"},
	}
	for name, args := range rejected {
		t.Run("rejects/"+name, func(t *testing.T) {
			if _, ok := args["role"]; !ok {
				args["role"] = "plan-auditor"
			}
			args["task"] = "audit"
			before := len(fake.calls(t))
			snap := auditSnapshotTree(t, repo.base)
			res, m := callRoleAuditTool(t, c, codexRoleAuditToolName, args)
			if !res.IsError {
				t.Fatalf("accepted: %v", m)
			}
			if n := len(fake.calls(t)) - before; n != 0 {
				t.Fatalf("rejected call reached codex %d times", n)
			}
			if d := auditDiffSnapshots(snap, auditSnapshotTree(t, repo.base)); len(d) > 0 {
				t.Fatalf("rejected call changed files: %v", d)
			}
			if strings.Contains(roleAuditResultText(res), "LAUNCH_RECORD") {
				t.Fatalf("rejection reported a launch record: %s", roleAuditResultText(res))
			}
		})
	}

	// 4. Async shape: start returns a job id before the audit ends; status and
	// result observe it; the verdict lands with exactly the returned text.
	msg := "## 판정\n\nPASS\n"
	fake.setExec(msg, 0)
	fake.write(t, "delay", "1")
	start := time.Now()
	res, m := callRoleAuditTool(t, c, codexRoleAuditToolName, map[string]any{
		"role": "sync-auditor", "worktree_root": repo.a1, "task": "audit it", "out": legalOut,
	})
	if res.IsError {
		t.Fatalf("legal start rejected: %s", roleAuditResultText(res))
	}
	jobID, _ := m["job_id"].(string)
	if jobID == "" || m["state"] != "running" {
		t.Fatalf("start result = %v, want a job id in state running", m)
	}
	if time.Since(start) > 900*time.Millisecond {
		t.Errorf("start blocked for %s — it must return before the audit ends", time.Since(start))
	}
	_, running := callRoleAuditTool(t, c, codexRoleAuditResultToolName, map[string]any{"job_id": jobID})
	if running["state"] != "running" {
		t.Errorf("result during the run = %v, want state running without blocking", running)
	}
	deadline := time.Now().Add(20 * time.Second)
	var st map[string]any
	for time.Now().Before(deadline) {
		_, st = callRoleAuditTool(t, c, codexRoleAuditStatusToolName, map[string]any{"job_id": jobID})
		if st["state"] != "running" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if st["state"] != "completed" {
		t.Fatalf("job did not complete: %v", st)
	}
	_, done := callRoleAuditTool(t, c, codexRoleAuditResultToolName, map[string]any{"job_id": jobID})
	if done["exit_code"] != float64(0) || done["verdict_path"] != legalOut {
		t.Fatalf("result = %v", done)
	}
	rec, _ := done["record_path"].(string)
	if !strings.HasPrefix(rec, ".moai/reports/codex-audit/sync-auditor-") {
		t.Fatalf("record_path = %q", rec)
	}
	raw, err := os.ReadFile(filepath.Join(repo.a1, filepath.FromSlash(rec)))
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(raw, &record); err != nil || record["route"] != "mcp" {
		t.Fatalf("launch record route = %v (err %v)", record["route"], err)
	}
	got, err := os.ReadFile(filepath.Join(repo.a1, filepath.FromSlash(legalOut)))
	if err != nil || string(got) != msg {
		t.Fatalf("verdict = %q (%v), want the returned text", got, err)
	}
	if ex := fake.execCalls(t); len(ex) != 1 {
		t.Fatalf("exec calls = %d, want 1", len(ex))
	}

	// 5. Unknown job ids are reported as errors.
	if res, _ := callRoleAuditTool(t, c, codexRoleAuditStatusToolName, map[string]any{"job_id": "nope"}); !res.IsError {
		t.Error("unknown job id accepted by status")
	}
}
