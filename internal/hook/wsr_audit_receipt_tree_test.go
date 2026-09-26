package hook

// SPEC-WORKTREE-STATE-ROOT-001 — the auditor receipt guard on a config-orphaned
// linked worktree: its store is the primary checkout's, every record carries
// the worktree's tree identity, and records of different trees coexist.
// Non-parallel throughout (t.Setenv, package clock).

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

const wsrGateRequired = "workflow:\n  audit:\n    gates:\n      codex: required\n"

// wsrFixture is fixture F (and F2 via add): P keeps .moai untracked, W is a
// linked worktree without .moai.
type wsrFixture struct {
	env []string
	P   string
	W   string
}

func wsrGitEnv(t *testing.T) []string {
	t.Helper()
	empty := filepath.Join(t.TempDir(), "empty-gitconfig")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			env = append(env, kv)
		}
	}
	return append(env, "GIT_CONFIG_GLOBAL="+empty, "GIT_CONFIG_SYSTEM="+empty, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
}

func wsrGit(t *testing.T, env []string, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

func wsrWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func newWSRHookFixture(t *testing.T, workflowYAML string) wsrFixture {
	t.Helper()
	env := wsrGitEnv(t)
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(base, "P")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	wsrGit(t, env, p, "init", "-q", "-b", "main")
	wsrWrite(t, filepath.Join(p, ".gitignore"), ".moai/\n")
	wsrWrite(t, filepath.Join(p, "tracked.txt"), "base\n")
	wsrWrite(t, filepath.Join(p, ".moai", "config", "sections", "workflow.yaml"), workflowYAML)
	wsrGit(t, env, p, "add", ".gitignore", "tracked.txt")
	wsrGit(t, env, p, "commit", "-q", "-m", "init")
	w := filepath.Join(base, "W")
	wsrGit(t, env, p, "worktree", "add", "-q", "-b", "wt", w)
	return wsrFixture{env: env, P: p, W: w}
}

func (fx wsrFixture) addWorktree(t *testing.T, name string) string {
	t.Helper()
	w := filepath.Join(filepath.Dir(fx.P), name)
	wsrGit(t, fx.env, fx.P, "worktree", "add", "-q", "-b", strings.ToLower(name), w)
	return w
}

func wsrTick(t *testing.T) {
	t.Helper()
	base := time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)
	n := 0
	prev := auditreceipt.Now
	auditreceipt.Now = func() time.Time { n++; return base.Add(time.Duration(n) * time.Second) }
	t.Cleanup(func() { auditreceipt.Now = prev })
}

func wsrStart(t *testing.T, cwd, agentID string) {
	t.Helper()
	if _, err := NewSubagentStartHandler().Handle(context.Background(), &HookInput{
		CWD: cwd, AgentID: agentID, AgentType: auditreceipt.AgentPlanAuditor, SessionID: "s",
	}); err != nil {
		t.Fatalf("SubagentStart: %v", err)
	}
}

func wsrVerdict(spec, receipt string) string {
	if receipt == "" {
		receipt = "none"
	}
	return "report\nAUDIT-VERDICT: PASS spec=" + spec + " receipts=" + receipt
}

func wsrStop(t *testing.T, cwd, agentID, message string, reentry bool) *HookOutput {
	t.Helper()
	return runStop(t, stopInput(cwd, auditreceipt.AgentPlanAuditor, agentID, message, reentry))
}

// wsrSpawn runs the PreToolUse spawn path with the model-audit log pointed at a
// scratch directory, so the spawn never writes under the tree it checks.
func wsrSpawn(t *testing.T, cwd string) *HookOutput {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	out, err := NewPreToolHandler(nil, DefaultSecurityPolicy()).Handle(context.Background(),
		receiptSpawnInput(cwd, "manager-develop", "Agent"))
	if err != nil {
		t.Fatalf("PreToolUse: %v", err)
	}
	return out
}

func wsrRecords(t *testing.T, dir string) []map[string]any {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	var out []map[string]any
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		out = append(out, m)
	}
	return out
}

func wsrStoreDir(root, kind string) string {
	return filepath.Join(root, ".moai", "state", "audit-receipts", kind)
}

func wsrBlocked(out *HookOutput) bool { return out != nil && out.Decision == "block" }

// AC-WSR-008: the guard is active on a config-orphaned W; its marker and
// rejection live in P's store with tree identity W; a receipt created after
// the start marker is accepted and clears W's rejection.
func TestWSR008_GuardActiveOnConfigOrphanedWorktree(t *testing.T) {
	wsrTick(t)
	fx := newWSRHookFixture(t, wsrGateRequired)

	wsrStart(t, fx.W, "a8")
	out := wsrStop(t, fx.W, "a8", wsrVerdict("SPEC-X-001", ""), false)
	if !wsrBlocked(out) || !strings.Contains(out.Reason, auditReceiptViolation) {
		t.Errorf("uncited PASS on W must be refused with %s, got %+v", auditReceiptViolation, out)
	}
	starts := wsrRecords(t, wsrStoreDir(fx.P, "starts"))
	if len(starts) != 1 || starts[0]["tree_root"] != fx.W {
		t.Errorf("want one start marker in P's store with tree_root W, got %v", starts)
	}
	rejs := wsrRecords(t, wsrStoreDir(fx.P, "rejections"))
	if len(rejs) != 1 || rejs[0]["tree_root"] != fx.W {
		t.Errorf("want one rejection in P's store with tree_root W, got %v", rejs)
	}
	if _, err := os.Stat(filepath.Join(fx.W, ".moai")); !os.IsNotExist(err) {
		t.Errorf("W/.moai must not exist (stat err=%v)", err)
	}

	// The same sequence citing a receipt codex_audit filed for W after the start.
	wsrStart(t, fx.W, "a8b")
	id, err := auditreceipt.WriteReceipt(fx.P, &auditreceipt.Receipt{
		Tool: auditreceipt.ToolCodexAudit, TreeRoot: fx.W, RootSource: auditreceipt.RootSourceArgument, CodexVerdict: "pass",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out := wsrStop(t, fx.W, "a8b", wsrVerdict("SPEC-X-001", id), false); wsrBlocked(out) {
		t.Errorf("PASS citing a valid receipt must be accepted, got block: %s", out.Reason)
	}
	for _, r := range wsrRecords(t, wsrStoreDir(fx.P, "rejections")) {
		if r["tree_root"] == fx.W {
			t.Errorf("W's rejection must be cleared by the corroborated PASS, still have %v", r)
		}
	}
}

// AC-WSR-007: rejections of different trees coexist in one store; a PASS in P
// clears only P's; the spawn check reads only its own tree's rejections.
func TestWSR007_RejectionsArePerTree(t *testing.T) {
	wsrTick(t)
	fx := newWSRHookFixture(t, wsrGateRequired)
	w2 := fx.addWorktree(t, "W2")

	wsrStart(t, fx.W, "A1")
	a1 := wsrStop(t, fx.W, "A1", wsrVerdict("SPEC-X-001", ""), false)

	// A2 starts before the codex audit it cites, as an auditor does.
	wsrStart(t, fx.P, "A2")
	id, err := auditreceipt.WriteReceipt(fx.P, &auditreceipt.Receipt{
		Tool: auditreceipt.ToolCodexAudit, TreeRoot: fx.P, RootSource: auditreceipt.RootSourceArgument, CodexVerdict: "pass",
	})
	if err != nil {
		t.Fatal(err)
	}
	a2 := wsrStop(t, fx.P, "A2", wsrVerdict("SPEC-Y-001", id), false)

	wsrStart(t, w2, "A3")
	a3 := wsrStop(t, w2, "A3", wsrVerdict("SPEC-X-001", ""), false)

	if !wsrBlocked(a1) || !wsrBlocked(a3) {
		t.Errorf("A1 and A3 uncited PASSes must be refused: a1=%+v a3=%+v", a1, a3)
	}
	if wsrBlocked(a2) {
		t.Errorf("A2 corroborated PASS must be accepted, got %s", a2.Reason)
	}
	var trees []string
	for _, r := range wsrRecords(t, wsrStoreDir(fx.P, "rejections")) {
		if r["agent_type"] == auditreceipt.AgentPlanAuditor && r["spec_id"] == "SPEC-X-001" {
			tr, _ := r["tree_root"].(string)
			trees = append(trees, tr)
		}
	}
	sort.Strings(trees)
	want := []string{fx.W, w2}
	sort.Strings(want)
	if strings.Join(trees, "|") != strings.Join(want, "|") {
		t.Errorf("want two plan-auditor/SPEC-X-001 rejections with tree roots W and W2, got %v", trees)
	}
	if ok, reason := denied(wsrSpawn(t, fx.W)); !ok || !strings.Contains(reason, auditReceiptViolation) || !strings.Contains(reason, "SPEC-X-001") {
		t.Errorf("spawn with cwd=W must be denied naming SPEC-X-001, got denied=%v reason=%q", ok, reason)
	}
	if ok, reason := denied(wsrSpawn(t, fx.P)); ok {
		t.Errorf("spawn with cwd=P must be allowed, got denied: %s", reason)
	}
}
