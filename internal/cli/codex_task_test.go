package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-CODEX-PHASE2-001 M3 — codex_task (REQ-CX2-006), the write opt-in gate
// (REQ-CX2-007), and resume_last thread reuse (REQ-CX2-008). Verifies
// AC-CX2-009, AC-CX2-010 (both arms plus the codex_setup arm), AC-CX2-011, and
// closes the tool-level half of AC-CX2-007 that M2 recorded as debt.
//
// The canned session runner from M1/M2 is reused verbatim (AP-1): these tests
// exercise the TASK layer over the existing transport, never a second client.

// ─── fixtures ───

// withCodexProjectDir points the project-root seam at root for the duration of
// the test. codex_task and codex_setup both resolve their config + job registry
// through it.
func withCodexProjectDir(t *testing.T, root string) {
	t.Helper()
	prev := projectDirResolver
	projectDirResolver = func() string { return root }
	t.Cleanup(func() { projectDirResolver = prev })
}

// writeCodexWorkflowYAML writes .moai/config/sections/workflow.yaml with the
// given body. An empty body writes no file at all (the absent-file state).
func writeCodexWorkflowYAML(t *testing.T, root, body string) {
	t.Helper()
	if body == "" {
		return
	}
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
}

// allowWriteYAML is the opt-in config body with allow_write set to v.
func allowWriteYAML(v string) string {
	return "workflow:\n  codex:\n    task:\n      allow_write: " + v + "\n"
}

// codexTaskScript is the canned transcript for ONE task turn: handshake ack,
// thread ack, turn/start ack, turn/started (the turnId source), the agent's
// output, then turn/completed.
func codexTaskScript(turnID, output string) []string {
	q := jsonString
	return []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":` + q(turnID) + `,"status":"inProgress"}}}`,
		`{"method":"turn/started","params":{"threadId":"tid-fake","turn":{"id":` + q(turnID) + `,"status":"inProgress"}}}`,
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":` + q(turnID) + `,"item":{"type":"agentMessage","id":"m1","text":` + q(output) + `}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":` + q(turnID) + `,"status":"completed"}}}`,
	}
}

// thisRepoRoot returns the toplevel of the repository containing the test's
// cwd — a registered worktree of the default (real) serving repository, so raw
// handler calls that bypass callCodexTask can pass the project_root gate.
func thisRepoRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	top, err := codexAuditGit(context.Background(), cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatalf("test cwd is not inside a git worktree: %v", err)
	}
	return top
}

// callCodexTask invokes the handler with the given arguments. When the caller
// supplies no project_root, this repository's own toplevel — a registered
// worktree of the test server's repository — is injected, so tests exercising
// behavior OTHER than the project_root gate keep exercising it. Gate tests
// needing an absent project_root must call the handler raw.
func callCodexTask(t *testing.T, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	if args == nil {
		args = map[string]any{}
	}
	if _, ok := args["project_root"]; !ok {
		args["project_root"] = thisRepoRoot(t)
	}
	res, err := handleCodexTask(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: args},
	})
	if err != nil {
		t.Fatalf("handleCodexTask returned a Go error (the tool must return structured results only): %v", err)
	}
	if res == nil {
		t.Fatal("handleCodexTask returned a nil result")
	}
	return res
}

// structuredMap decodes a tool result's StructuredContent into a map.
func structuredMap(t *testing.T, res *mcp.CallToolResult) map[string]any {
	t.Helper()
	if res.StructuredContent == nil {
		t.Fatalf("result carries no StructuredContent: %+v", res)
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode StructuredContent %s: %v", raw, err)
	}
	return m
}

// sandboxPolicyType returns the `type` discriminant of a transmitted
// sandboxPolicy value. SandboxPolicy is an internally-tagged UNION object
// ({"type":"readOnly"}) per the generated protocol schema (codex-cli 0.146.1),
// exactly like `target` — a bare string is rejected with JSON-RPC -32600. The
// AC names the variant; this reads the variant out of the envelope the protocol
// declares.
func sandboxPolicyType(t *testing.T, params map[string]any) (string, bool) {
	t.Helper()
	raw, ok := params["sandboxPolicy"]
	if !ok {
		return "", false
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("sandboxPolicy must be an internally-tagged object; got %T (%v)", raw, raw)
	}
	typ, _ := obj["type"].(string)
	return typ, true
}

// awaitTerminalJob polls the registry until the job reaches a terminal status.
// Only the record is read — never the shared fake session's sent slice, which
// the background goroutine is still appending to.
func awaitTerminalJob(t *testing.T, reg *codexJobRegistry, id string) CodexJobRecord {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rec, err := reg.load(id)
		if err == nil {
			switch rec.Status {
			case codexJobStatusCompleted, codexJobStatusFailed, codexJobStatusCancelled:
				return rec
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job %s never reached a terminal status within the deadline", id)
	return CodexJobRecord{}
}

// ─── AC-CX2-009 (REQ-CX2-006) ───

// Foreground: the result carries the task output.
func TestCodexTask_ForegroundReturnsOutput(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withCodexSession(t, codexTaskScript("trn-fg", "the refactor is complete"))

	res := callCodexTask(t, map[string]any{"prompt": "refactor the parser"})
	if res.IsError {
		t.Fatalf("unexpected IsError result: %+v", res)
	}
	got := structuredMap(t, res)
	if got["background"] != false {
		t.Errorf("background = %v, want false", got["background"])
	}
	if out, _ := got["output"].(string); out != "the refactor is complete" {
		t.Errorf("output = %q, want the completed task output", out)
	}
	if id, _ := got["job_id"].(string); id != "" {
		t.Errorf("foreground result must carry no job id; got %q", id)
	}
}

// Background: the result carries a job id that resolves to an existing record.
func TestCodexTask_BackgroundReturnsJobIDResolvingToRecord(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withCodexSession(t, codexTaskScript("trn-bg", "background work done"))

	res := callCodexTask(t, map[string]any{"prompt": "audit the module", "background": true})
	if res.IsError {
		t.Fatalf("unexpected IsError result: %+v", res)
	}
	got := structuredMap(t, res)
	if got["background"] != true {
		t.Errorf("background = %v, want true", got["background"])
	}
	jobID, _ := got["job_id"].(string)
	if jobID == "" {
		t.Fatal("background result must carry a job id")
	}
	if out, _ := got["output"].(string); out != "" {
		t.Errorf("background result must not block for the output; got %q", out)
	}

	reg := newCodexJobRegistry(root)
	if _, err := reg.load(jobID); err != nil {
		t.Fatalf("job id %q does not resolve to a record: %v", jobID, err)
	}
	rec := awaitTerminalJob(t, reg, jobID)
	if rec.Status != codexJobStatusCompleted {
		t.Errorf("final status = %q, want %q", rec.Status, codexJobStatusCompleted)
	}
	if rec.Output != "background work done" {
		t.Errorf("recorded output = %q, want the task output", rec.Output)
	}
	if rec.TurnID != "trn-bg" {
		t.Errorf("recorded turn_id = %q, want trn-bg (captured mid-flight)", rec.TurnID)
	}
}

// ─── AC-CX2-010 (REQ-CX2-007) — the write gate, fail-closed ───

func TestCodexTask_WriteGateFailClosed(t *testing.T) {
	cases := []struct {
		name       string
		configBody string
		wantGrant  bool
		wantPolicy string
	}{
		{"absent-file", "", false, codexSandboxReadOnly},
		{"absent-key", "workflow:\n  codex:\n    review_gate:\n      enabled: true\n", false, codexSandboxReadOnly},
		{"malformed-yaml", "workflow: [this is not: a mapping\n", false, codexSandboxReadOnly},
		{"explicit-false", allowWriteYAML("false"), false, codexSandboxReadOnly},
		{"explicit-true", allowWriteYAML("true"), true, codexSandboxWorkspaceWrite},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			withCodexProjectDir(t, root)
			writeCodexWorkflowYAML(t, root, tc.configBody)
			sess := withCodexSession(t, codexTaskScript("trn-w", "done"))

			res := callCodexTask(t, map[string]any{"prompt": "edit the file", "write": true})
			got := structuredMap(t, res)

			if got["write_requested"] != true {
				t.Errorf("write_requested = %v, want true", got["write_requested"])
			}
			if got["write_granted"] != tc.wantGrant {
				t.Errorf("write_granted = %v, want %v", got["write_granted"], tc.wantGrant)
			}
			if !tc.wantGrant {
				if note, _ := got["note"].(string); note == "" {
					t.Error("a refused write must be stated in the result (note is empty)")
				}
			}
			// The transmitted turn (3rd request) carries the policy explicitly.
			policy, present := sandboxPolicyType(t, sentParams(t, sess.sent, 2))
			if !present {
				t.Fatal("turn/start must carry sandboxPolicy on EVERY turn")
			}
			if policy != tc.wantPolicy {
				t.Errorf("sandboxPolicy.type = %q, want %q", policy, tc.wantPolicy)
			}
		})
	}
}

// AC-CX2-010 sticky-policy arm (M0-forced): sandboxPolicy applies "for this turn
// AND subsequent turns", so a write-enabled turn would leave its thread
// write-enabled for a later turn that did not opt in. Two turns are driven on
// ONE reused thread — turn 1 with write, turn 2 without — and turn 2 must
// transmit readOnly explicitly rather than omitting the field.
func TestCodexTask_SandboxPolicyResetOnReusedThread(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	writeCodexWorkflowYAML(t, root, allowWriteYAML("true"))

	// A prior background job recorded the thread both turns will resume.
	reg := newCodexJobRegistry(root)
	if _, err := reg.create(codexJobSpec{ThreadID: "tid-fake", Mode: codexTaskMode, RequestSummary: "prior job"}); err != nil {
		t.Fatalf("seed prior job: %v", err)
	}

	sess := withCodexSession(t, codexTaskScript("trn-x", "ok"))

	callCodexTask(t, map[string]any{"prompt": "turn one", "write": true, "resume_last": true})
	callCodexTask(t, map[string]any{"prompt": "turn two", "resume_last": true})

	if n := countSentMethod(t, sess.sent, codexMethodThreadStart); n != 0 {
		t.Errorf("thread/start sent %d times, want 0 (both turns resume the recorded thread)", n)
	}
	turns := []int{}
	for i, line := range sess.sent {
		if m, _ := sentRequest(t, line)["method"].(string); m == codexMethodTurnStart {
			turns = append(turns, i)
		}
	}
	if len(turns) != 2 {
		t.Fatalf("turn/start sent %d times, want 2 (%v)", len(turns), sess.sent)
	}

	for _, i := range turns {
		if got, _ := sentParams(t, sess.sent, i)["threadId"].(string); got != "tid-fake" {
			t.Errorf("turn at request %d addresses threadId %q, want the reused tid-fake", i, got)
		}
	}

	first, ok := sandboxPolicyType(t, sentParams(t, sess.sent, turns[0]))
	if !ok || first != codexSandboxWorkspaceWrite {
		t.Errorf("turn 1 sandboxPolicy = %q (present=%v), want %q", first, ok, codexSandboxWorkspaceWrite)
	}
	second, ok := sandboxPolicyType(t, sentParams(t, sess.sent, turns[1]))
	if !ok {
		t.Fatal("turn 2 OMITTED sandboxPolicy — the thread's inherited write-enabled policy would outlive the request that opted into it")
	}
	if second != codexSandboxReadOnly {
		t.Errorf("turn 2 sandboxPolicy = %q, want %q", second, codexSandboxReadOnly)
	}
}

// AC-CX2-010 codex_setup arm: the decoded result map carries an allow_write key
// whose boolean value is the literal expected value for each config state,
// alongside the pre-existing enable_review_gate key.
func TestCodexSetup_ReportsAllowWriteState(t *testing.T) {
	cases := []struct {
		name       string
		configBody string
		want       bool
	}{
		{"absent-key", "", false},
		{"malformed-yaml", "workflow: [this is not: a mapping\n", false},
		{"explicit-false", allowWriteYAML("false"), false},
		{"explicit-true", allowWriteYAML("true"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			withCodexProjectDir(t, root)
			writeCodexWorkflowYAML(t, root, tc.configBody)
			withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })

			res, err := handleCodexSetup(context.Background(), mcp.CallToolRequest{})
			if err != nil {
				t.Fatalf("handleCodexSetup: %v", err)
			}
			got := structuredMap(t, res)
			if _, ok := got["enable_review_gate"]; !ok {
				t.Error("codex_setup must keep reporting enable_review_gate")
			}
			raw, ok := got["allow_write"]
			if !ok {
				t.Fatal("codex_setup result carries no allow_write key")
			}
			if raw != tc.want {
				t.Errorf("allow_write = %v, want %v", raw, tc.want)
			}
		})
	}
}

// The hand-rolled allow_write reader must agree with the real config loader on
// every state. This is the parallel of TestReviewGateReaders_AgreeWithConfigLoader
// and it exists for a concrete reason: an earlier revision of the SIBLING gate
// read its key at the TOP level, so the toggle could never read true and nothing
// caught it. A reader that silently disagrees with the loader on the write gate
// would be the same bug with a working tree at stake.
func TestCodexTaskAllowWriteReader_AgreesWithConfigLoader(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"on", "workflow:\n    codex:\n        task:\n            allow_write: true\n"},
		{"off", "workflow:\n    codex:\n        task:\n            allow_write: false\n"},
		{"absent-block", "workflow:\n    codex:\n        review_gate:\n            enabled: true\n"},
		{"absent-file", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeGateWorkflowYAML(t, tc.body)
			cfg, err := config.NewLoader().Load(filepath.Join(dir, ".moai"))
			if err != nil {
				t.Fatalf("config load: %v", err)
			}
			if got, want := readCodexTaskAllowWrite(dir), cfg.Workflow.Codex.Task.AllowWrite; got != want {
				t.Errorf("allow_write reader = %v, config loader = %v (schema drift)", got, want)
			}
		})
	}
}

// The distributed default is opt-out: a project with no config at all must not
// be able to run a writing turn (plan.md §F AP-5).
func TestCodexTaskAllowWrite_DistributedDefaultIsFalse(t *testing.T) {
	if config.NewDefaultWorkflowConfig().Codex.Task.AllowWrite {
		t.Error("the distributed default for workflow.codex.task.allow_write must be false")
	}
}

// ─── AC-CX2-011 (REQ-CX2-008) — resume_last ───

func TestCodexTask_ResumeLastReusesRecordedThread(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)
	if _, err := reg.create(codexJobSpec{ThreadID: "tid-fake", Mode: codexTaskMode, RequestSummary: "prior"}); err != nil {
		t.Fatalf("seed prior job: %v", err)
	}
	sess := withCodexSession(t, codexTaskScript("trn-r", "resumed work"))

	res := callCodexTask(t, map[string]any{"prompt": "continue", "resume_last": true})
	got := structuredMap(t, res)
	if got["resumed_thread"] != true {
		t.Errorf("resumed_thread = %v, want true", got["resumed_thread"])
	}
	if tid, _ := got["thread_id"].(string); tid != "tid-fake" {
		t.Errorf("thread_id = %q, want the recorded tid-fake", tid)
	}
	if n := countSentMethod(t, sess.sent, codexMethodThreadStart); n != 0 {
		t.Errorf("thread/start sent %d times, want 0 when resuming", n)
	}
	if n := countSentMethod(t, sess.sent, codexMethodThreadResume); n != 1 {
		t.Errorf("thread/resume sent %d times, want 1", n)
	}
	if got, _ := sentParams(t, sess.sent, 1)["threadId"].(string); got != "tid-fake" {
		t.Errorf("thread/resume threadId = %q, want tid-fake", got)
	}
	if got, _ := sentParams(t, sess.sent, 2)["threadId"].(string); got != "tid-fake" {
		t.Errorf("turn/start threadId = %q, want the recorded tid-fake", got)
	}
}

func TestCodexTask_ResumeLastWithNoRecordedThread(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	sess := withCodexSession(t, codexTaskScript("trn-n", "fresh work"))

	res := callCodexTask(t, map[string]any{"prompt": "start", "resume_last": true})
	got := structuredMap(t, res)
	if got["resumed_thread"] != false {
		t.Errorf("resumed_thread = %v, want false when nothing is recorded", got["resumed_thread"])
	}
	if note, _ := got["note"].(string); note == "" {
		t.Error("the result must state that no prior thread was resumed")
	}
	if n := countSentMethod(t, sess.sent, codexMethodThreadStart); n != 1 {
		t.Errorf("thread/start sent %d times, want 1 (a new thread is opened)", n)
	}
}

// ─── AC-CX2-007 tool arm (M2 debt closure) ───

// With .moai/state occupied by a regular file the registry cannot write, the
// TOOL returns a structured error result — the clause M2 could not close
// because codex_task did not exist yet. The process does not panic and the
// server serves the next call.
func TestCodexTask_UnwritableStateDirStructuredError(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	moai := filepath.Join(root, ".moai")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	if err := os.WriteFile(filepath.Join(moai, "state"), []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("occupy state path: %v", err)
	}
	withCodexSession(t, codexTaskScript("trn-e", "unreachable"))

	res := callCodexTask(t, map[string]any{"prompt": "do work", "background": true})
	if !res.IsError {
		t.Fatalf("an unwritable state dir must yield a structured error result; got %+v", res)
	}

	// The server still serves the next call: a foreground task needs no record.
	withCodexSession(t, codexTaskScript("trn-e2", "still serving"))
	next := callCodexTask(t, map[string]any{"prompt": "do work"})
	if next.IsError {
		t.Errorf("the next tool call must still be served; got %+v", next)
	}
}

// ─── §B edge cases ───

func TestCodexTask_FailOpenOnMissingCodex(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withCodexLookPath(t, func(string) (string, error) { return "", errFakeLookPath })
	withCodexSession(t, nil)

	res := callCodexTask(t, map[string]any{"prompt": "anything"})
	if res.IsError {
		t.Fatalf("a missing codex must fail OPEN as a structured result, not an error: %+v", res)
	}
	got := structuredMap(t, res)
	if got["status"] != codexJobStatusFailed {
		t.Errorf("status = %v, want %q", got["status"], codexJobStatusFailed)
	}
	if e, _ := got["error"].(string); e == "" {
		t.Error("the result must name the failure")
	}
}

func TestCodexTask_MissingPromptIsStructuredError(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withCodexSession(t, nil)

	res := callCodexTask(t, map[string]any{})
	if !res.IsError {
		t.Fatalf("a missing prompt must be a structured error result; got %+v", res)
	}
}

// TestCodexTaskProjectRootGate pins the codex_task project_root contract
// (SPEC-CODEX-ROLE-AUDIT-ROOT-001 REQ-005): the argument is required — an
// absent or empty value is refused naming the argument, never defaulted to the
// primary checkout — and the presented root passes the same registered-
// same-repository verification as codex_role_audit. An accepted root becomes
// the tree the turn acts on: the thread's cwd is the canonical accepted root.
func TestCodexTaskProjectRootGate(t *testing.T) {
	repo := newAuditRepo(t)
	orig := codexServingDir
	codexServingDir = func() (string, error) { return repo.a1, nil }
	t.Cleanup(func() { codexServingDir = orig })
	withCodexProjectDir(t, repo.a1)

	// A registered worktree of a DIFFERENT repository (the gating arm observable).
	foreign := filepath.Join(repo.base, "FB")
	foreignWt := filepath.Join(repo.base, "FB1")
	for _, d := range []string{foreign, foreignWt} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	auditGit(t, foreign, "init", "-q", "-b", "main")
	auditGit(t, foreign, "commit", "-q", "--allow-empty", "-m", "init")
	auditGit(t, foreign, "worktree", "add", "-q", "-b", "wt-fb", foreignWt)
	// An unregistered directory inside the serving repository.
	unreg := filepath.Join(repo.a, "unregistered")
	if err := os.MkdirAll(unreg, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Run("absent project_root refused", func(t *testing.T) {
		withCodexSession(t, nil)
		res, herr := handleCodexTask(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{Arguments: map[string]any{"prompt": "task"}},
		})
		if herr != nil {
			t.Fatal(herr)
		}
		if !res.IsError {
			t.Fatalf("an absent project_root must be refused; got %+v", res)
		}
		if s := roleAuditResultText(res); !strings.Contains(s, "project_root") {
			t.Fatalf("refusal must name the missing argument; got %q", s)
		}
	})
	t.Run("empty project_root refused", func(t *testing.T) {
		withCodexSession(t, nil)
		res := callCodexTask(t, map[string]any{"prompt": "task", "project_root": "  "})
		if !res.IsError {
			t.Fatalf("an empty project_root must be refused; got %+v", res)
		}
	})
	t.Run("foreign repository refused", func(t *testing.T) {
		withCodexSession(t, nil)
		res := callCodexTask(t, map[string]any{"prompt": "task", "project_root": foreignWt})
		if !res.IsError {
			t.Fatalf("a foreign repository must be refused; got %+v", res)
		}
		if s := roleAuditResultText(res); !strings.Contains(s, "belongs to a different repository") {
			t.Fatalf("refusal must carry the same-repo fragment; got %q", s)
		}
	})
	t.Run("unregistered directory refused", func(t *testing.T) {
		withCodexSession(t, nil)
		res := callCodexTask(t, map[string]any{"prompt": "task", "project_root": unreg})
		if !res.IsError {
			t.Fatalf("an unregistered directory must be refused; got %+v", res)
		}
		if s := roleAuditResultText(res); !strings.Contains(s, "is not a registered worktree of this repository") {
			t.Fatalf("refusal must carry the registration fragment; got %q", s)
		}
	})
	t.Run("registered worktree accepted and acted on", func(t *testing.T) {
		withCodexTaskTimeout(t, 10*time.Second)
		sess := withCodexSession(t, codexTaskScript("trn-pr", "done on the presented tree"))

		res := callCodexTask(t, map[string]any{"prompt": "task", "project_root": repo.a2})
		if res.IsError {
			t.Fatalf("a registered sibling worktree must be accepted: %s", roleAuditResultText(res))
		}
		out := structuredMap(t, res)
		if out["status"] != codexJobStatusCompleted {
			t.Fatalf("the turn did not complete: %v", out)
		}
		// The thread must have been opened on the ACCEPTED root, not on the
		// serving checkout and not on a default.
		for _, line := range sess.sent {
			var m map[string]any
			if json.Unmarshal([]byte(line), &m) != nil {
				continue
			}
			if params, ok := m["params"].(map[string]any); ok {
				if cwd, ok := params["cwd"].(string); ok && cwd != "" {
					if cwd != repo.a2 {
						t.Fatalf("thread cwd = %q, want the accepted root %q", cwd, repo.a2)
					}
					return
				}
			}
		}
		t.Fatal("no cwd found in the sent thread requests")
	})
}
