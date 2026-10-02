package cli

// SPEC-CODEX-REVIEW-OWNERSHIP-001 M3 — fixtures and rigs for the self-review
// tools (codex_review / glm_review). Every tree is a real git repository; every
// call goes through an in-process MCP client so the registered schema, the
// handler and the result shape are exercised together. The reviewer seams are
// the existing ones (withCodexSession, withGLMSeams) — nothing here reaches a
// live codex or a live z.ai.

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// selfReviewTestPrefixes restates the runtime-managed prefixes on purpose: the
// expected material in these tests is computed from this list, not from the
// production variable, so a drift in either is a visible difference.
var selfReviewTestPrefixes = []string{
	".moai/state/", ".moai/cache/", ".moai/reports/", ".moai/logs/", ".moai/harness/", ".claude/agent-memory/",
}

// selfReviewIndepDiff is the independently computed `git diff <base> -- .`
// with the runtime-managed prefixes excluded by pathspec — the material the
// acceptance criteria name.
func selfReviewIndepDiff(t *testing.T, dir, base string) string {
	t.Helper()
	args := []string{"-C", dir, "diff", base, "--", "."}
	for _, p := range selfReviewTestPrefixes {
		args = append(args, ":(exclude)"+p)
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return string(out)
}

// selfReviewFixture is the acceptance.md §B card fixture. File letters follow
// the acceptance text: A committed on the card, B tracked + unstaged, G tracked
// + staged only, C and C2 untracked non-runtime, D a TRACKED runtime-prefix
// path modified, R an untracked runtime-prefix path, F foreign WIP in the
// primary-role tree (a tracked modification, so a diff taken from the wrong
// tree would show it).
type selfReviewFixture struct {
	primary string // develop checkout carrying F
	card    string // linked worktree on WT-selfreview-card
	canon   string // filepath.EvalSymlinks(card)
	link    string // a symlinked spelling of card
	c0      string // root commit (a real ancestor that is NOT the merge base)
	base    string // git merge-base develop HEAD at fixture time (== c1)
	devTip  string // develop tip (!= base: develop advanced after the card forked)
}

func newSelfReviewFixture(t *testing.T) *selfReviewFixture {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	primary := t.TempDir()
	cardScopeGit(t, primary, "init", "-q", "-b", "develop")
	writeCardFile(t, primary, "go.mod", "module example.com/sr\n\ngo 1.22\n")
	cardScopeGit(t, primary, "add", "-A")
	cardScopeGit(t, primary, "commit", "-q", "-m", "c0")
	c0 := cardScopeGit(t, primary, "rev-parse", "HEAD")
	writeCardFile(t, primary, ".moai/state/fixture-tracked.json", "{\"v\":1}\n")
	// c1 also adds a NON-runtime file, so a diff taken from c0 (the sentinel) and
	// one taken from c1 (the real merge base) differ in the material GLM is shown.
	writeCardFile(t, primary, "c1_marker.go", "package main\n\n// added on the commit the card forks from\n")
	cardScopeGit(t, primary, "add", "-A")
	cardScopeGit(t, primary, "commit", "-q", "-m", "c1")

	card := t.TempDir()
	cardScopeGit(t, primary, "worktree", "add", "-q", "-b", "WT-selfreview-card", card)

	// develop advances after the card forked: tip != merge base.
	writeCardFile(t, primary, "develop_extra.go", "package main\n\n// develop moved on\n")
	cardScopeGit(t, primary, "add", "develop_extra.go")
	cardScopeGit(t, primary, "commit", "-q", "-m", "c2")
	// F: foreign WIP in the primary-role tree.
	writeCardFile(t, primary, "go.mod", "module example.com/sr\n\ngo 1.22\n\n// foreign WIP F\n")

	writeCardFile(t, card, "card_a.go", "package main\n\n// card commit file A\n")
	writeCardFile(t, card, "card_b.go", "package main\n\n// card file B v1\n")
	writeCardFile(t, card, "card_g.go", "package main\n\n// card file G v1\n")
	cardScopeGit(t, card, "add", "-A")
	cardScopeGit(t, card, "commit", "-q", "-m", "feat(card): A")
	writeCardFile(t, card, "card_b.go", "package main\n\n// card file B v2 (unstaged)\n")
	writeCardFile(t, card, "card_g.go", "package main\n\n// card file G v2 (staged only)\n")
	cardScopeGit(t, card, "add", "card_g.go")
	writeCardFile(t, card, ".moai/state/fixture-tracked.json", "{\"v\":2}\n") // D
	writeCardFile(t, card, "untracked_c.go", "package main\n\n// untracked C\n")
	writeCardFile(t, card, "untracked_c2.go", "package main\n\n// untracked C2\n")
	writeCardFile(t, card, ".moai/state/untracked-runtime.json", "{}\n") // R

	canon, err := filepath.EvalSymlinks(card)
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "card-link")
	if err := os.Symlink(card, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	return &selfReviewFixture{
		primary: primary,
		card:    card,
		canon:   canon,
		link:    link,
		c0:      c0,
		base:    cardScopeGit(t, card, "merge-base", "develop", "HEAD"),
		devTip:  cardScopeGit(t, primary, "rev-parse", "develop"),
	}
}

// newSelfReviewEmptyCard adds a second card worktree whose card diff is empty:
// it forks at the develop tip, so HEAD == merge base, and the tree is clean.
func (f *selfReviewFixture) newEmptyCard(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cardScopeGit(t, f.primary, "worktree", "add", "-q", "-b", "WT-selfreview-empty", dir)
	return dir
}

// newSelfReviewPlainTree builds a MoAI project root (it has a .moai directory)
// that is a git repository on the named branch. kind selects the non-card
// variants: "branch" (HEAD on branch), "detached", "nogit".
func newSelfReviewPlainTree(t *testing.T, branch string, dirty bool) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	cardScopeGit(t, dir, "init", "-q", "-b", branch)
	writeCardFile(t, dir, "a.go", "package a\n")
	writeCardFile(t, dir, ".moai/state/plain.json", "{}\n")
	cardScopeGit(t, dir, "add", "-A")
	cardScopeGit(t, dir, "commit", "-q", "-m", "init")
	if dirty {
		writeCardFile(t, dir, "a.go", "package a\n\n// uncommitted change\n")
	}
	return dir
}

func newSelfReviewDetachedTree(t *testing.T) string {
	t.Helper()
	dir := newSelfReviewPlainTree(t, "develop", true)
	cardScopeGit(t, dir, "checkout", "-q", "--detach")
	return dir
}

func newSelfReviewNonGitTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// selfReviewCall invokes one self-review tool through an in-process client and
// returns the raw result plus its structured content as a generic map. A
// transport-level error (for example "tool not found") fails the test.
func selfReviewCall(t *testing.T, tool string, args map[string]any) (*mcp.CallToolResult, map[string]any) {
	t.Helper()
	srv := newMoaiMCPServer()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	ctx := context.Background()
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = tool
	req.Params.Arguments = args
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("%s: transport error: %v", tool, err)
	}
	var m map[string]any
	if res.StructuredContent != nil {
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatalf("marshal structured content: %v", err)
		}
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("decode structured content %s: %v", raw, err)
		}
	}
	return res, m
}

// noteSession is a minimal MCP client session whose notification channel the
// test can count: the in-process client does not surface server notifications,
// so the heartbeat check drives the server's message handler directly.
type noteSession struct{ ch chan mcp.JSONRPCNotification }

func (s *noteSession) Initialize()                                         {}
func (s *noteSession) Initialized() bool                                   { return true }
func (s *noteSession) NotificationChannel() chan<- mcp.JSONRPCNotification { return s.ch }
func (s *noteSession) SessionID() string                                   { return "selfreview-notes" }

// selfReviewNotifications calls tool with a progress token over a session that
// counts the notifications the server sends back, and returns that count.
func selfReviewNotifications(t *testing.T, tool string, args map[string]any) int {
	t.Helper()
	srv := newMoaiMCPServer()
	sess := &noteSession{ch: make(chan mcp.JSONRPCNotification, 256)}
	ctx := srv.WithContext(context.Background(), sess)
	raw, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "tools/call",
		"params": map[string]any{"name": tool, "arguments": args, "_meta": map[string]any{"progressToken": "tok-1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := srv.HandleMessage(ctx, raw).(mcp.JSONRPCResponse); !ok {
		t.Fatalf("%s: the tools/call did not return a JSON-RPC response", tool)
	}
	return len(sess.ch)
}

// selfReviewListTools returns the registered tools as generic JSON objects
// keyed by name, so schemas compare by value rather than by Go type.
func selfReviewListTools(t *testing.T) map[string]map[string]any {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	srv := newMoaiMCPServer()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	ctx := context.Background()
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	out := make(map[string]map[string]any, len(res.Tools))
	for _, tool := range res.Tools {
		raw, err := json.Marshal(tool)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		out[tool.Name] = m
	}
	return out
}

// selfReviewGLMDoer counts and records the HTTP calls the GLM backend makes.
type selfReviewGLMDoer struct {
	stubGLMDoer
	calls int
}

func (d *selfReviewGLMDoer) Do(req *http.Request) (*http.Response, error) {
	d.calls++
	return d.stubGLMDoer.Do(req)
}

// sent decodes the last request body posted to z.ai.
func (d *selfReviewGLMDoer) sent(t *testing.T) glmMessagesRequest {
	t.Helper()
	var r glmMessagesRequest
	if err := json.Unmarshal([]byte(d.gotBody), &r); err != nil {
		t.Fatalf("decode the z.ai request body %q: %v", d.gotBody, err)
	}
	return r
}

// selfReviewRig wires one backend to a stub reviewer that returns the named
// verdict ("pass", "fail" or "inconclusive") and reports how many times the
// reviewer was reached. For the inconclusive variant the reviewer is
// unavailable (codex binary absent / GLM key missing) — the fail-open arm.
type selfReviewRig struct {
	tool    string
	backend string
	calls   func() int
	glm     *selfReviewGLMDoer
	codex   *fakeCodexSession
}

func newSelfReviewRig(t *testing.T, backend, verdict string) *selfReviewRig {
	t.Helper()
	rig := &selfReviewRig{backend: backend, tool: backend + "_review"}
	switch backend {
	case "codex":
		body := realCleanReview
		if verdict == "fail" {
			body = "- [P1] probe finding"
		}
		rig.codex = withCodexSession(t, codexSessionScript(body))
		if verdict == "inconclusive" {
			withCodexLookPath(t, func(string) (string, error) { return "", os.ErrNotExist })
		}
		rig.calls = func() int { return len(rig.codex.sent) }
	case "glm":
		out := ReviewOutput{Verdict: verdict, Summary: verdict + " from stub", Findings: []Finding{}, NextSteps: []string{}}
		rig.glm = &selfReviewGLMDoer{stubGLMDoer: stubGLMDoer{body: glmMessagesResp(t, out)}}
		key := "stub-key"
		if verdict == "inconclusive" {
			key = ""
		}
		withGLMSeams(t, key, rig.glm)
		rig.calls = func() int { return rig.glm.calls }
	default:
		t.Fatalf("unknown backend %q", backend)
	}
	return rig
}

// codexWire returns the thread/start params and the review/start target the
// fake codex session received.
func codexWire(t *testing.T, sess *fakeCodexSession) (thread map[string]any, target map[string]any) {
	t.Helper()
	if len(sess.sent) < 3 {
		t.Fatalf("expected >=3 requests on the codex wire, got %d (%v)", len(sess.sent), sess.sent)
	}
	thread, _ = sentRequest(t, sess.sent[1])["params"].(map[string]any)
	review, _ := sentRequest(t, sess.sent[2])["params"].(map[string]any)
	target, _ = review["target"].(map[string]any)
	if thread == nil || target == nil {
		t.Fatalf("thread/start or review/start target missing: %v", sess.sent)
	}
	return thread, target
}

// srReadFile reads a file or fails the test.
func srReadFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}
