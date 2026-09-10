package cli

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// SPEC-MOAI-MCP-SERVER-001 M2 (AC-MCP-007 / AC-MCP-008).
//
// These tests pin the codex_audit + codex_setup contract BEFORE the
// implementation exists (RED). They assert: the review-output schema shape
// (§G.4), native→review/start + adversarial→turn/start dispatch, fail-open
// VerdictInconclusive on missing / erroring / malformed codex, the codex_setup
// Go probe shape, and that the implementation file carries NO Node/.mjs bridge
// dependency (REQ-MCP-007). The codex binary is OPTIONAL + experimental (R1);
// tests NEVER require a real codex — they inject a fake command runner.

// fakeCodexRunner is the injectable command-execution double. It records the
// invocation and returns a canned stdout / err, so tests drive every codex
// response shape (valid review, non-zero exit, malformed output) without
// spawning a process and without PATH stubs (cross-platform safe).
type fakeCodexRunner struct {
	stdout      string
	err         error
	gotBinary   string
	gotArgs     []string
	gotStdin    string
	calls       int
	stdoutByCmd map[string]string // keyed by joined args, for per-command responses
}

func (f *fakeCodexRunner) run(_ context.Context, binaryPath string, args []string, stdin string) (string, error) {
	f.calls++
	f.gotBinary = binaryPath
	f.gotArgs = args
	f.gotStdin = stdin
	if f.stdoutByCmd != nil {
		key := strings.Join(args, " ")
		if out, ok := f.stdoutByCmd[key]; ok {
			return out, f.err
		}
	}
	return f.stdout, f.err
}

// withCodexRunner swaps the package-level codexRunner seam and restores it on
// cleanup, so each test is isolated.
func withCodexRunner(t *testing.T, r codexCommandRunner) {
	t.Helper()
	prev := codexRunner
	codexRunner = r
	t.Cleanup(func() { codexRunner = prev })
}

func withCodexLookPath(t *testing.T, fn func(string) (string, error)) {
	t.Helper()
	prev := codexLookPath
	codexLookPath = fn
	t.Cleanup(func() { codexLookPath = prev })
}

// --- session seam doubles (the review-gate JSON-RPC session path) ---

// fakeCodexSession is the session-spawning double: start returns a fakeConn that
// replays the canned NDJSON lines in order and records every sent request line.
type fakeCodexSession struct {
	lines    []string // canned NDJSON lines recv yields, in order
	startErr error    // if non-nil, start returns this error
	sent     []string // recorded sent request lines (across the single session)
}

func (f *fakeCodexSession) start(context.Context, string, []string) (codexConn, error) {
	if f.startErr != nil {
		return nil, f.startErr
	}
	return &fakeCodexConn{lines: f.lines, sent: &f.sent}, nil
}

// fakeCodexConn replays canned lines via recv and records sends.
type fakeCodexConn struct {
	lines []string
	idx   int
	sent  *[]string
}

func (c *fakeCodexConn) send(line string) error { *c.sent = append(*c.sent, line); return nil }
func (c *fakeCodexConn) recv() (string, bool) {
	if c.idx >= len(c.lines) {
		return "", false
	}
	l := c.lines[c.idx]
	c.idx++
	return l, true
}
func (c *fakeCodexConn) close() error { return nil }

// withCodexSession installs a fake session that replays `lines` and exposes the
// recorded sent request lines. The single-shot runner + LookPath seams are also
// pointed at a stub binary so the gate's binary-resolution succeeds.
func withCodexSession(t *testing.T, lines []string) *fakeCodexSession {
	t.Helper()
	prevRunner, prevLook, prevSess := codexRunner, codexLookPath, codexSession
	codexRunner = stubCodexRunner{}
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	sess := &fakeCodexSession{lines: lines}
	codexSession = sess
	t.Cleanup(func() { codexRunner, codexLookPath, codexSession = prevRunner, prevLook, prevSess })
	return sess
}

// codexSessionScript builds a canned NDJSON session transcript that drives the
// client through initialize → thread/start → review/start ack → item/completed
// (exitedReviewMode carrying reviewText) → turn/completed. The reviewText
// controls the synthesized verdict (severity-tagged finding bullets ⇒ fail).
func codexSessionScript(reviewText string) []string {
	return []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":"trn","status":"inProgress"}}}`,
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn","completedAtMs":1,"item":{"type":"exitedReviewMode","id":"e1","review":` + jsonString(reviewText) + `}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn","status":"completed"}}}`,
	}
}

// jsonString quotes s as a JSON string (handles escapes).
func jsonString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `""`
	}
	return string(b)
}

// --- AC-MCP-007: codex_audit unified modes + schema output ---

// TestReviewOutputSchemaShape pins the §G.4 locked review-output schema so the
// orchestrator-translation layer + the future GLM backend (M3) share one shape.
func TestReviewOutputSchemaShape(t *testing.T) {
	raw, err := json.Marshal(ReviewOutput{
		Verdict:   "pass",
		Summary:   "s",
		Findings:  []Finding{{Severity: "high", Title: "t", Body: "b", File: "f", Line: 9, Confidence: 0.5, Recommendation: "r"}},
		NextSteps: []string{"a", "b"},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"verdict", "summary", "findings", "next_steps"} {
		if _, ok := got[key]; !ok {
			t.Errorf("ReviewOutput JSON missing key %q (§G.4 schema)", key)
		}
	}
	fi, _ := got["findings"].([]any)
	if len(fi) != 1 {
		t.Fatalf("findings: want 1, got %d", len(fi))
	}
	finding, _ := fi[0].(map[string]any)
	for _, key := range []string{"severity", "title", "body", "file", "line", "confidence", "recommendation"} {
		if _, ok := finding[key]; !ok {
			t.Errorf("Finding JSON missing key %q (§G.4 schema)", key)
		}
	}
	if VerdictInconclusive != "inconclusive" {
		t.Errorf("VerdictInconclusive = %q, want %q", VerdictInconclusive, "inconclusive")
	}
}

// TestCodexAudit_NativeDispatchesReviewStart proves mode=native drives a codex
// app-server JSON-RPC session whose third request is the review/start method,
// and surfaces the synthesized pass verdict through the review-output schema.
func TestCodexAudit_NativeDispatchesReviewStart(t *testing.T) {
	sess := withCodexSession(t, codexSessionScript("clean change, no findings"))

	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"mode":   codexModeNative,
			"target": codexTargetUncommitted,
		}},
	})
	if err != nil {
		t.Fatalf("handleCodexAudit: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected IsError result; codex present must not error")
	}
	// The session must have sent initialize, thread/start, then review/start (3rd).
	if len(sess.sent) < 3 {
		t.Fatalf("expected ≥3 sent requests; got %d (%v)", len(sess.sent), sess.sent)
	}
	if !strContains(sess.sent[2], codexMethodReviewStart) {
		t.Errorf("3rd request must be %q; got %s", codexMethodReviewStart, sess.sent[2])
	}
	if strContains(sess.sent[2], codexMethodTurnStart) {
		t.Errorf("native mode must NOT dispatch %q", codexMethodTurnStart)
	}
	if !jsonResultHasVerict(res, "pass") {
		t.Errorf("result does not surface verdict pass")
	}
}

// TestCodexAudit_AdversarialDispatchesTurnStart proves mode=adversarial drives a
// session whose third request is turn/start (the adversarial-review prompt rides
// the UserInput input array) and surfaces the synthesized fail verdict.
func TestCodexAudit_AdversarialDispatchesTurnStart(t *testing.T) {
	// turn/start is not a review-mode turn, so the transcript carries the verdict
	// as a final agentMessage rather than exitedReviewMode.
	lines := []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":"trn","status":"inProgress"}}}`,
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn","completedAtMs":1,"item":{"type":"agentMessage","id":"m1","text":"- [P1] concurrency hazard found"}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn","status":"completed"}}}`,
	}
	sess := withCodexSession(t, lines)

	res, _ := handleCodexAudit(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"mode":   codexModeAdversarial,
			"target": codexTargetBaseBranch,
			"focus":  "concurrency",
		}},
	})
	if !strContains(sess.sent[2], codexMethodTurnStart) {
		t.Errorf("3rd request must be %q; got %s", codexMethodTurnStart, sess.sent[2])
	}
	// SPEC-CODEX-REVIEW-TARGET-001 AC-CRT-007. This was the ONLY codex test in
	// the repository holding the string `baseBranch`, and it asserted nothing
	// about it: turn/start carries no target at all, so the argument above was
	// inert — a `baseBranch` that was never on the wire could not fail however
	// malformed the review/start path was. Assert the fact that makes it inert,
	// so the argument stops looking like coverage it never provided. The native
	// baseBranch target shape is pinned in codex_review_target_test.go.
	turnReq := sentRequest(t, sess.sent[2])
	turnParams, _ := turnReq["params"].(map[string]any)
	if _, has := turnParams["target"]; has {
		t.Errorf("adversarial mode must not serialize a target (turn/start declares none); params=%v", turnParams)
	}
	if !jsonResultHasVerict(res, "fail") {
		t.Errorf("adversarial result must surface verdict fail")
	}
}

// TestCodexAudit_FailOpenOnMissingCodex proves the mandatory fail-open
// (REQ-MCP-012 preview): a missing codex binary yields VerdictInconclusive as a
// STRUCTURED result (never a Go error, never a hard crash).
func TestCodexAudit_FailOpenOnMissingCodex(t *testing.T) {
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })
	withCodexSession(t, nil) // session must not be started (binary missing)

	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{"mode": codexModeNative}},
	})
	if err != nil {
		t.Fatalf("fail-open must not return a Go error; got %v", err)
	}
	if res == nil || res.IsError {
		t.Fatalf("fail-open must return a non-error structured result")
	}
	if !jsonResultHasVerict(res, VerdictInconclusive) {
		t.Errorf("missing codex must yield verdict %q", VerdictInconclusive)
	}
}

// TestCodexAudit_FailOpenOnCodexError proves a codex session-start failure
// degrades to VerdictInconclusive rather than surfacing a hard error.
func TestCodexAudit_FailOpenOnCodexError(t *testing.T) {
	prevRunner, prevLook, prevSess := codexRunner, codexLookPath, codexSession
	codexRunner = stubCodexRunner{}
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	codexSession = &fakeCodexSession{startErr: errFakeCodexCrash}
	t.Cleanup(func() { codexRunner, codexLookPath, codexSession = prevRunner, prevLook, prevSess })

	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{"mode": codexModeNative}},
	})
	if err != nil {
		t.Fatalf("fail-open must not return a Go error; got %v", err)
	}
	if !jsonResultHasVerict(res, VerdictInconclusive) {
		t.Errorf("codex error must yield verdict %q", VerdictInconclusive)
	}
}

// TestCodexAudit_FailOpenOnMalformedResponse proves a stream of unparseable
// NDJSON lines degrades to VerdictInconclusive (no panic, no hard error): the
// client skips noise and hits EOF without a verdict.
func TestCodexAudit_FailOpenOnMalformedResponse(t *testing.T) {
	withCodexSession(t, []string{"this is not json-rpc", "still not json"})

	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{"mode": codexModeNative}},
	})
	if err != nil {
		t.Fatalf("fail-open must not return a Go error; got %v", err)
	}
	if !jsonResultHasVerict(res, VerdictInconclusive) {
		t.Errorf("malformed response must yield verdict %q", VerdictInconclusive)
	}
}

// TestCodexTools_Registered proves the M2 codex tools are declared in the
// tools/list surface alongside the M1 core tools (AC-MCP-007/008 registration).
func TestCodexTools_Registered(t *testing.T) {
	tools := newMoaiMCPServer().ListTools()
	for _, name := range []string{"codex_audit", "codex_setup"} {
		if _, ok := tools[name]; !ok {
			t.Errorf("tool %q not registered in tools/list", name)
		}
	}
}

func TestCodexSetup_GoProbeNoNodeBridge(t *testing.T) {
	// Isolate stage 1 (empty CODEX_HOME ⇒ no auth.json ⇒ descend) and feed the
	// status line through the login-status seam — without isolation the
	// developer's real auth.json decides auth_provider, so the assertion below
	// would track the machine's login state instead of the probe under test.
	t.Setenv("CODEX_HOME", t.TempDir())
	withCodexLookPath(t, func(string) (string, error) { return "/fake/codex", nil })
	withCodexRunner(t, &fakeCodexRunner{
		stdoutByCmd: map[string]string{
			"--version": "codex 1.2.3\n",
		},
	})
	withCodexLoginStatusRunner(t, func(_ context.Context, _ string) ([]byte, []byte, int, error) {
		return nil, []byte("Logged in using ChatGPT\n"), 0, nil // field-observed channel: stderr
	})

	res, err := handleCodexSetup(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{}},
	})
	if err != nil {
		t.Fatalf("handleCodexSetup: %v", err)
	}
	if res == nil || res.IsError {
		t.Fatalf("codex_setup must return a non-error structured result")
	}
	payload := resultJSON(t, res)
	if got, _ := payload["installed"].(bool); !got {
		t.Errorf("installed: want true when LookPath succeeds")
	}
	if got, _ := payload["version"].(string); !strings.Contains(got, "1.2.3") {
		t.Errorf("version: want 1.2.3 substring, got %q", got)
	}
	if got, _ := payload["auth_provider"].(string); got != codexAuthChatGPT {
		t.Errorf("auth_provider: want %q, got %q", codexAuthChatGPT, got)
	}
	// enable_review_gate toggle is exposed (read of workflow.codex.review_gate.enabled).
	if _, ok := payload["enable_review_gate"]; !ok {
		t.Errorf("codex_setup must expose the enable_review_gate toggle")
	}
}

func TestCodexSetup_NotInstalledReportsUnknown(t *testing.T) {
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })
	withCodexRunner(t, &fakeCodexRunner{})

	res, _ := handleCodexSetup(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{}},
	})
	payload := resultJSON(t, res)
	if got, _ := payload["installed"].(bool); got {
		t.Errorf("installed: want false when LookPath fails")
	}
	if got, _ := payload["auth_provider"].(string); got != codexAuthUnknown {
		t.Errorf("auth_provider when absent: want %q, got %q", codexAuthUnknown, got)
	}
}

// TestMCP_Codex_NoNodeBridge is the hard contract: the codex backend
// implementation file MUST NOT spawn a Node process or load a JavaScript
// bridge module (REQ-MCP-007). The check targets actual USAGE patterns
// (exec spawning "node", a quoted require/import of a JS module) so it does
// not flag prose comments that merely mention the absence of a bridge.
func TestMCP_Codex_NoNodeBridge(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(".", "mcp_codex.go"))
	if err != nil {
		t.Fatalf("read mcp_codex.go: %v", err)
	}
	s := string(src)
	// Spawning node as a subprocess: exec.Command("node", ...) / CommandContext(.., "node", ...)
	if strings.Contains(s, `"node"`) {
		t.Errorf(`codex backend must not spawn a Node process (found "node" literal)`)
	}
	// Loading a JS bridge module via require/import inside a Go string.
	for _, needle := range []string{`".mjs"`, `require(`, `import(`} {
		if strings.Contains(s, needle) {
			t.Errorf("codex backend must not reference a JS bridge module (found %q)", needle)
		}
	}
}

// --- helpers ---

var (
	errFakeLookPath   = errors.New("fake: codex binary not in PATH")
	errFakeCodexCrash = errors.New("fake: codex exited non-zero")
)

// jsonResultHasVerict inspects the CallToolResult for the expected verdict. It
// checks both the TextContent JSON and the StructuredContent (a typed `any` —
// marshaled to JSON before parsing), so it is robust to either result shape.
func jsonResultHasVerict(res *mcp.CallToolResult, want string) bool {
	for _, m := range resultMaps(res) {
		if v, _ := m["verdict"].(string); v == want {
			return true
		}
	}
	return false
}

// resultMaps returns every parseable JSON map carried by the result (text
// content + structured content), in order.
func resultMaps(res *mcp.CallToolResult) []map[string]any {
	var maps []map[string]any
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			var m map[string]any
			if err := json.Unmarshal([]byte(tc.Text), &m); err == nil {
				maps = append(maps, m)
			}
		}
	}
	if res.StructuredContent != nil {
		if b, err := json.Marshal(res.StructuredContent); err == nil {
			var m map[string]any
			if err := json.Unmarshal(b, &m); err == nil {
				maps = append(maps, m)
			}
		}
	}
	return maps
}

func resultJSON(t *testing.T, res *mcp.CallToolResult) map[string]any {
	t.Helper()
	if maps := resultMaps(res); len(maps) > 0 {
		return maps[0]
	}
	t.Fatalf("no parseable JSON in result")
	return nil
}

func strContains(haystack, needle string) bool { return strings.Contains(haystack, needle) }

// --- production-path coverage (exec + config readers) ---

// TestRealCodexRunner_HappyAndError covers the production exec wrapper
// honestly (the injected seam leaves it at 0%). `cat` copies stdin→stdout ⇒
// happy stdout carries the piped input; a non-existent binary ⇒ error.
func TestRealCodexRunner_HappyAndError(t *testing.T) {
	cat, err := exec.LookPath("cat")
	if err != nil {
		t.Skip("cat not available")
	}
	r := realCodexRunner{}
	out, runErr := r.run(context.Background(), cat, nil, "hello")
	if runErr != nil {
		t.Fatalf("cat run: %v", runErr)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("cat stdout: want 'hello', got %q", out)
	}
	if _, err := r.run(context.Background(), "/nonexistent/binary/xyz", nil, ""); err == nil {
		t.Errorf("nonexistent binary must error")
	}
}

// TestClassifyCodexAuth_Branches covers each auth-provider classification arm
// through the two-stage ladder (SPEC-CODEX-LAUNCHER-001 M1). CODEX_HOME is
// isolated to an empty temp dir so stage 1 finds no auth.json and every case
// reaches the stage-2 probe — without isolation the developer's real auth.json
// short-circuits the ladder and the result follows the machine, not the case.
func TestClassifyCodexAuth_Branches(t *testing.T) {
	t.Setenv("CODEX_HOME", t.TempDir())
	cases := map[string]string{
		// Whole-line grammar positives — the two classification arms (REQ-CL-009).
		"Logged in using ChatGPT": codexAuthChatGPT,
		"Logged in using API key": codexAuthAPIKey,
		// Substring-era lines the whole-line grammar deliberately rejects: a
		// line merely CONTAINING a provider term never classifies.
		"Logged in to ChatGPT":             codexAuthUnknown,
		"Auth mode: API key (sk-...)":      codexAuthUnknown,
		"Configured custom provider 'foo'": codexAuthUnknown,
		"":                                 codexAuthUnknown,
		"something unrecognized":           codexAuthUnknown,
	}
	for out, want := range cases {
		// drive classifyCodexAuth via the login-status seam returning the canned stdout.
		withCodexLoginStatusRunner(t, func(_ context.Context, _ string) ([]byte, []byte, int, error) {
			return []byte(out), nil, 0, nil // "" ⇒ no grammar line ⇒ unknown
		})
		if got := classifyCodexAuth(context.Background(), "/fake/codex"); got != want {
			t.Errorf("classifyCodexAuth(%q) = %q, want %q", out, got, want)
		}
	}
	// error arm: probe runner returns err ⇒ unknown (an unreadable probe is a gap).
	withCodexLoginStatusRunner(t, func(_ context.Context, _ string) ([]byte, []byte, int, error) {
		return nil, nil, -1, errFakeCodexCrash
	})
	if got := classifyCodexAuth(context.Background(), "/fake/codex"); got != codexAuthUnknown {
		t.Errorf("classifyCodexAuth on runner error = %q, want %q", got, codexAuthUnknown)
	}
}

// TestReadCodexReviewGateEnabled_ConfigBranches covers the fail-CLOSED truth
// table: missing file ⇒ false, explicit true ⇒ true, explicit false ⇒ false.
// The documents use the deployed NESTED shape (a `workflow:` root); the flat
// shape this test formerly used is pinned as NOT honoured by
// TestReviewGateReaders_HonourNestedWorkflowKeyPath.
func TestReadCodexReviewGateEnabled_ConfigBranches(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// missing file ⇒ false
	if readCodexReviewGateEnabled(dir) {
		t.Errorf("missing config: want false (fail-CLOSED default off)")
	}
	// explicit true ⇒ true
	if err := os.WriteFile(filepath.Join(cfgDir, "workflow.yaml"),
		[]byte("workflow:\n  codex:\n    review_gate:\n      enabled: true\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !readCodexReviewGateEnabled(dir) {
		t.Errorf("enabled: true ⇒ want true")
	}
	// explicit false ⇒ false
	if err := os.WriteFile(filepath.Join(cfgDir, "workflow.yaml"),
		[]byte("workflow:\n  codex:\n    review_gate:\n      enabled: false\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if readCodexReviewGateEnabled(dir) {
		t.Errorf("enabled: false ⇒ want false")
	}
	// empty projectDir ⇒ false
	if readCodexReviewGateEnabled("") {
		t.Errorf("empty projectDir ⇒ want false")
	}
}

// ─── SPEC-CODEX-TEST-GAPS-001 M2 (REQ-CTG-002 / AC-CTG-002) ───────────────

// TestCodexIDMatches is the 5-arm table over codexIDMatches, the JSON-RPC id
// discriminator between awaitCodexResponse and everything else on the wire:
// a notification (no id) must miss, an integer id must match exactly, a
// string id must match through the strconv arm, and input that unmarshals as
// neither number nor string must miss rather than panic or match.
func TestCodexIDMatches(t *testing.T) {
	cases := []struct {
		name string
		raw  json.RawMessage
		want int
		hit  bool
	}{
		{"empty raw is a miss (notification)", nil, 7, false},
		{"integer id match", json.RawMessage("7"), 7, true},
		{"integer id mismatch", json.RawMessage("8"), 7, false},
		{"string id match (strconv arm)", json.RawMessage(`"7"`), 7, true},
		{"malformed — neither number nor string", json.RawMessage(`[7]`), 7, false},
	}
	for _, tc := range cases {
		if got := codexIDMatches(tc.raw, tc.want); got != tc.hit {
			t.Errorf("%s: codexIDMatches(%s, %d) = %t, want %t", tc.name, tc.raw, tc.want, got, tc.hit)
		}
	}
}

// ─── SPEC-CODEX-TEST-GAPS-001 M3 (REQ-CTG-003 / AC-CTG-003) ───────────────

// TestAwaitCodexResponse pins the context-cancel return of the read loop: a
// canceled context must come back AS the context error — not as a match, not
// as the stdout-closed EOF error — even when the connection still has
// non-matching noise lines buffered.
func TestAwaitCodexResponse(t *testing.T) {
	sent := []string{}
	conn := &fakeCodexConn{
		lines: []string{
			`{"method":"turn/started","params":{"threadId":"tid-fake"}}`,
			`{"id":99,"result":{"other":"request"}}`,
		},
		sent: &sent,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	msg, err := awaitCodexResponse(conn, 1, ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("awaitCodexResponse with a canceled context = (msg %+v, err %v), want the context.Canceled error", msg, err)
	}
}

// ─── SPEC-CODEX-TEST-GAPS-001 M5 (REQ-CTG-007 / AC-CTG-007) ───────────────

// TestCodexSessionError pins the delegation contract of codexSessionError,
// constructed through the production fail-open shape (codexHandshakeFailure):
// Error() defers verbatim to the cause (the operator sees codex's words, not
// our summary), Unwrap reaches the cause, and both errors.Is and errors.As
// traverse the wrap. The summary is preserved on the struct for callers that
// need the fail-open framing.
func TestCodexSessionError(t *testing.T) {
	cause := errors.New("codex initialize write failed: broken pipe")
	err := codexHandshakeFailure(&fakeCodexConn{}, "codex initialize rejected", cause)
	if err == nil {
		t.Fatal("codexHandshakeFailure returned a nil error")
	}
	if err.Error() != cause.Error() {
		t.Errorf("Error() = %q, want the cause verbatim %q", err.Error(), cause.Error())
	}
	if unwrapped := errors.Unwrap(err); unwrapped != cause {
		t.Errorf("Unwrap() = %v, want the cause itself", unwrapped)
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is must reach the cause through the wrap")
	}
	var se *codexSessionError
	if !errors.As(err, &se) {
		t.Fatal("errors.As must reach *codexSessionError through the wrap")
	}
	if se.summary != "codex initialize rejected" {
		t.Errorf("summary = %q, want the fail-open framing preserved on the struct", se.summary)
	}
	if se.cause != cause {
		t.Errorf("cause field = %v, want the construction cause", se.cause)
	}
}

// ─── SPEC-CODEX-TEST-GAPS-001 M6 (REQ-CTG-012 / AC-CTG-012) ───────────────

// TestRealCodexConnPid covers the 3-branch pid read directly, same package,
// no subprocess: nil cmd → 0, cmd without a Process → 0, and a Process from
// os.FindProcess(os.Getpid()) → that pid. The nil arms are a disjunction —
// cmd == nil must short-circuit before cmd.Process is dereferenced
// (acceptance.md §D.1).
func TestRealCodexConnPid(t *testing.T) {
	if got := (&realCodexConn{}).pid(); got != 0 {
		t.Errorf("pid on a zero conn = %d, want 0", got)
	}
	if got := (&realCodexConn{cmd: &exec.Cmd{}}).pid(); got != 0 {
		t.Errorf("pid with cmd but no Process = %d, want 0", got)
	}
	self, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess(os.Getpid()): %v", err)
	}
	if got := (&realCodexConn{cmd: &exec.Cmd{Process: self}}).pid(); got != os.Getpid() {
		t.Errorf("pid with a live Process = %d, want %d", got, os.Getpid())
	}
}

// ─── SPEC-CODEX-COVER-RESIDUAL-001 M2 (REQ-CCR-005 / AC-CCR-006) ──────────

// TestCodexSessionHandlePid covers the 3-branch pid read on the session handle
// directly, same package, no subprocess: a nil receiver → 0, a handle with a
// nil conn → 0, and a handle whose conn is a fakeCodexConn → fakeCodexConnPID.
//
// It is the sibling of TestRealCodexConnPid above — the same 3-arm nil-guard
// shape on the other receiver — and the nil arms are likewise a disjunction:
// h == nil MUST short-circuit before h.conn is dereferenced, or the typed-nil
// call panics rather than returning 0.
//
// Constructing &fakeCodexConn{} with a nil sent field is safe here: pid() reads
// no field of the conn beyond the interface dispatch, and send is never called.
func TestCodexSessionHandlePid(t *testing.T) {
	if got := (*codexSessionHandle)(nil).pid(); got != 0 {
		t.Errorf("pid on a nil handle = %d, want 0", got)
	}
	if got := (&codexSessionHandle{}).pid(); got != 0 {
		t.Errorf("pid on a handle with no conn = %d, want 0", got)
	}
	if got := (&codexSessionHandle{conn: &fakeCodexConn{}}).pid(); got != fakeCodexConnPID {
		t.Errorf("pid with a process-bearing conn = %d, want %d", got, fakeCodexConnPID)
	}
}
