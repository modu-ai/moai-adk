package cli

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// SPEC-CODEX-RESUME-SCOPE-001 — codex_task resume_last must not resume another
// work item's thread. Verifies AC-CRS-001 … AC-CRS-015 over the fixture
// transport (acceptance.md §A): no live codex call is made anywhere here.
//
// Two fixtures are used, as acceptance.md §A defines them:
//   - the STOCK fixture (codexTaskScript) answers every thread request with
//     tid-fake, so a test on it asserts only what was SENT;
//   - the ECHO fixture (codexTaskScriptEcho) answers with the requested id, so a
//     test on it may assert turn/start's threadId and resumed_thread as well.

// ─── fixtures ───

// codexTaskScriptEcho is codexTaskScript with every tid-fake replaced by
// threadID — the fixture in which codex returns the thread that was requested.
func codexTaskScriptEcho(threadID, turnID, output string) []string {
	lines := codexTaskScript(turnID, output)
	for i := range lines {
		lines[i] = strings.ReplaceAll(lines[i], "tid-fake", threadID)
	}
	return lines
}

// seedCodexJob records one background-job record carrying threadID, workKey,
// and summary. The short sleep keeps consecutive records' updated_at distinct,
// so "recorded later" is unambiguous in the tests that rely on order.
func seedCodexJob(t *testing.T, reg *codexJobRegistry, threadID, workKey, summary string) CodexJobRecord {
	t.Helper()
	rec, err := reg.create(codexJobSpec{ThreadID: threadID, Mode: codexTaskMode, RequestSummary: summary, WorkKey: workKey})
	if err != nil {
		t.Fatalf("seed record %s: %v", threadID, err)
	}
	time.Sleep(3 * time.Millisecond)
	return rec
}

// sentThreadResumeID returns the threadId of the thread/resume request the
// fixture recorded, failing when none was sent.
func sentThreadResumeID(t *testing.T, sent []string) string {
	t.Helper()
	for i, line := range sent {
		if m, _ := sentRequest(t, line)["method"].(string); m == codexMethodThreadResume {
			id, _ := sentParams(t, sent, i)["threadId"].(string)
			return id
		}
	}
	t.Fatalf("no thread/resume request was sent (sent %d lines)", len(sent))
	return ""
}

// sentTurnStartThreadID returns the threadId of the first turn/start request.
func sentTurnStartThreadID(t *testing.T, sent []string) string {
	t.Helper()
	for i, line := range sent {
		if m, _ := sentRequest(t, line)["method"].(string); m == codexMethodTurnStart {
			id, _ := sentParams(t, sent, i)["threadId"].(string)
			return id
		}
	}
	t.Fatalf("no turn/start request was sent (sent %d lines)", len(sent))
	return ""
}

// assertRefusal checks the REQ-CRS-010 refusal shape: a structured result with
// status failed and the given error_code, NOT marked as an MCP tool error.
func assertRefusal(t *testing.T, res *mcp.CallToolResult, code string) map[string]any {
	t.Helper()
	if res.IsError {
		t.Errorf("refusal %s must not set IsError (REQ-CRS-010)", code)
	}
	got := structuredMap(t, res)
	if got["status"] != codexJobStatusFailed {
		t.Errorf("status = %v, want %q", got["status"], codexJobStatusFailed)
	}
	if got["error_code"] != code {
		t.Errorf("error_code = %v, want %q", got["error_code"], code)
	}
	return got
}

// candidateThreadIDs lists the thread ids of a refusal's candidates, in order.
func candidateThreadIDs(t *testing.T, got map[string]any) []string {
	t.Helper()
	raw, ok := got["candidates"].([]any)
	if !ok {
		t.Fatalf("candidates missing or not a list: %v", got["candidates"])
	}
	ids := make([]string, 0, len(raw))
	for _, c := range raw {
		m, _ := c.(map[string]any)
		id, _ := m["thread_id"].(string)
		ids = append(ids, id)
	}
	return ids
}

func assertStrings(t *testing.T, what string, got, want []string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("%s = %v, want %v", what, got, want)
	}
}

// ─── AC-CRS-001 (REQ-CRS-002, 004) — the issue, closed by work_key ───

func TestCodexTaskResumeScope_WorkKeyResumesOwnCardThread(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)
	seedCodexJob(t, reg, "thr-5", "card-A", "card A")
	seedCodexJob(t, reg, "thr-6", "card-B", "card B")
	sess := withCodexSession(t, codexTaskScriptEcho("thr-5", "trn-a", "work"))

	res := callCodexTask(t, map[string]any{"prompt": "continue card A", "resume_last": true, "work_key": "card-A"})
	got := structuredMap(t, res)

	if id := sentThreadResumeID(t, sess.sent); id != "thr-5" {
		t.Errorf("thread/resume threadId = %q, want thr-5 (card-A's thread, not the newer card-B one)", id)
	}
	if n := countSentMethod(t, sess.sent, codexMethodThreadResume); n != 1 {
		t.Errorf("thread/resume sent %d times, want 1", n)
	}
	if n := countSentMethod(t, sess.sent, codexMethodThreadStart); n != 0 {
		t.Errorf("thread/start sent %d times, want 0", n)
	}
	if got["resume_thread_id"] != "thr-5" || got["resume_basis"] != "work_key" {
		t.Errorf("resume_thread_id/resume_basis = %v/%v, want thr-5/work_key", got["resume_thread_id"], got["resume_basis"])
	}
}

// ─── AC-CRS-002 (REQ-CRS-001) — the resume basis is reported ───

func TestCodexTaskResumeScope_SoleThreadReportsBasis(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	seedCodexJob(t, newCodexJobRegistry(root), "tid-fake", "", "prior")
	withCodexSession(t, codexTaskScript("trn-s", "work"))

	got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "continue", "resume_last": true}))

	if got["resume_thread_id"] != "tid-fake" || got["thread_id"] != "tid-fake" {
		t.Errorf("resume_thread_id/thread_id = %v/%v, want tid-fake/tid-fake", got["resume_thread_id"], got["thread_id"])
	}
	if got["resumed_thread"] != true {
		t.Errorf("resumed_thread = %v, want true", got["resumed_thread"])
	}
	if got["resume_basis"] != "sole_thread" {
		t.Errorf("resume_basis = %v, want sole_thread", got["resume_basis"])
	}
}

// ─── AC-CRS-003 (REQ-CRS-001, 003) — thread_id resumes exactly that thread ───

func TestCodexTaskResumeScope_ThreadIDResumesExactThread(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)
	seedCodexJob(t, reg, "thr-5", "", "card A")
	seedCodexJob(t, reg, "thr-6", "", "card B") // newer
	sess := withCodexSession(t, codexTaskScriptEcho("thr-5", "trn-t", "work"))

	got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p", "thread_id": "thr-5"}))

	if id := sentThreadResumeID(t, sess.sent); id != "thr-5" {
		t.Errorf("thread/resume threadId = %q, want thr-5", id)
	}
	if id := sentTurnStartThreadID(t, sess.sent); id != "thr-5" {
		t.Errorf("turn/start threadId = %q, want thr-5", id)
	}
	if n := countSentMethod(t, sess.sent, codexMethodThreadStart); n != 0 {
		t.Errorf("thread/start sent %d times, want 0", n)
	}
	if got["resume_thread_id"] != "thr-5" || got["resume_basis"] != "thread_id" {
		t.Errorf("resume_thread_id/resume_basis = %v/%v, want thr-5/thread_id", got["resume_thread_id"], got["resume_basis"])
	}
}

// ─── AC-CRS-004 (REQ-CRS-003, 010) — an unrecorded thread_id is refused ───

func TestCodexTaskResumeScope_UnrecordedThreadIDRefused(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	seedCodexJob(t, newCodexJobRegistry(root), "thr-5", "", "card A")
	sess := withCodexSession(t, codexTaskScript("trn-x", "work"))

	res := callCodexTask(t, map[string]any{"prompt": "p", "thread_id": "thr-999"})
	got := assertRefusal(t, res, "thread_not_recorded")

	raw, _ := json.Marshal(got)
	if !strings.Contains(string(raw), "thr-999") {
		t.Errorf("refusal must name the supplied thread_id thr-999; got %s", raw)
	}
	if len(sess.sent) != 0 {
		t.Errorf("fixture recorded %d sent messages, want 0 (no codex process may start)", len(sess.sent))
	}
}

// ─── AC-CRS-006 (REQ-CRS-002, 006, 010) — ambiguity is refused, not guessed ───

func TestCodexTaskResumeScope_AmbiguousResumeLastRefused(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)
	seedCodexJob(t, reg, "thr-5", "", "card A")
	seedCodexJob(t, reg, "thr-6", "", "card B")
	sess := withCodexSession(t, codexTaskScript("trn-x", "work"))

	res := callCodexTask(t, map[string]any{"prompt": "continue card A", "resume_last": true})

	if len(sess.sent) != 0 {
		t.Errorf("fixture recorded %d sent messages, want 0 (neither thread/resume nor thread/start)", len(sess.sent))
	}
	got := assertRefusal(t, res, "resume_ambiguous")
	if got["candidate_total"] != float64(2) {
		t.Errorf("candidate_total = %v, want 2", got["candidate_total"])
	}
	assertStrings(t, "candidate order", candidateThreadIDs(t, got), []string{"thr-6", "thr-5"})
	for _, c := range got["candidates"].([]any) {
		m := c.(map[string]any)
		for _, key := range []string{"thread_id", "request_summary", "updated_at"} {
			if _, ok := m[key]; !ok {
				t.Errorf("candidate %v lacks %q", m, key)
			}
		}
	}
}

// ─── AC-CRS-007 (REQ-CRS-001, 007) — records sharing a thread count once ───

func TestCodexTaskResumeScope_SharedThreadRecordsCountOnce(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)
	for i := 0; i < 3; i++ {
		seedCodexJob(t, reg, "thr-5", "", "same thread")
	}
	sess := withCodexSession(t, codexTaskScript("trn-s", "work"))

	got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p", "resume_last": true}))

	if got["error_code"] != nil {
		t.Errorf("error_code = %v, want none (one distinct thread is not ambiguous)", got["error_code"])
	}
	if id := sentThreadResumeID(t, sess.sent); id != "thr-5" {
		t.Errorf("thread/resume threadId = %q, want thr-5", id)
	}
	if got["resume_thread_id"] != "thr-5" || got["resume_basis"] != "sole_thread" {
		t.Errorf("resume_thread_id/resume_basis = %v/%v, want thr-5/sole_thread", got["resume_thread_id"], got["resume_basis"])
	}
}

// ─── AC-CRS-008 (REQ-CRS-004) — an unmatched work_key opens a new thread ───

func TestCodexTaskResumeScope_UnmatchedWorkKeyOpensNewThread(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	seedCodexJob(t, newCodexJobRegistry(root), "thr-6", "card-B", "card B")
	sess := withCodexSession(t, codexTaskScript("trn-n", "work"))

	got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p", "resume_last": true, "work_key": "card-A"}))

	if n := countSentMethod(t, sess.sent, codexMethodThreadResume); n != 0 {
		t.Errorf("thread/resume sent %d times, want 0", n)
	}
	if n := countSentMethod(t, sess.sent, codexMethodThreadStart); n != 1 {
		t.Errorf("thread/start sent %d times, want 1", n)
	}
	if _, ok := got["resume_thread_id"]; ok {
		t.Errorf("resume_thread_id present (%v); want absent when no thread/resume was sent", got["resume_thread_id"])
	}
	if _, ok := got["resume_basis"]; ok {
		t.Errorf("resume_basis present (%v); want absent when no thread/resume was sent", got["resume_basis"])
	}
	note, _ := got["note"].(string)
	if !strings.Contains(note, "card-A") || !strings.Contains(note, "no prior thread") {
		t.Errorf("note = %q, want a statement that no prior thread is recorded for card-A", note)
	}
}

// ─── AC-CRS-009 (REQ-CRS-005) — a background record carries the work_key ───

func TestCodexTaskResumeScope_BackgroundRecordsWorkKey(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	withCodexSession(t, codexTaskScript("trn-b", "bg work"))

	got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p", "background": true, "work_key": "card-A"}))
	jobID, _ := got["job_id"].(string)
	if jobID == "" {
		t.Fatalf("background call returned no job_id: %v", got)
	}
	reg := newCodexJobRegistry(root)
	awaitTerminalJob(t, reg, jobID)

	raw, err := os.ReadFile(reg.pathFor(jobID))
	if err != nil {
		t.Fatalf("read record file: %v", err)
	}
	var file map[string]any
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("decode record file: %v", err)
	}
	if file["work_key"] != "card-A" {
		t.Errorf("record file work_key = %v, want card-A", file["work_key"])
	}

	status, err := handleCodexJobStatus(t.Context(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{codexJobIDArg: jobID}},
	})
	if err != nil {
		t.Fatalf("codex_job_status: %v", err)
	}
	if st := structuredMap(t, status); st["work_key"] != "card-A" {
		t.Errorf("codex_job_status work_key = %v, want card-A", st["work_key"])
	}
}

// ─── AC-CRS-010 (REQ-CRS-005, 010) — a malformed work_key is refused first ───

func TestCodexTaskResumeScope_InvalidWorkKeyRefused(t *testing.T) {
	cases := map[string]string{
		"blank":   "   ",
		"129B":    strings.Repeat("k", 129),
		"control": "card\nA",
	}
	for name, key := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			withCodexProjectDir(t, root)
			sess := withCodexSession(t, codexTaskScript("trn-x", "work"))

			res := callCodexTask(t, map[string]any{"prompt": "p", "resume_last": true, "work_key": key})
			assertRefusal(t, res, "invalid_work_key")
			if len(sess.sent) != 0 {
				t.Errorf("fixture recorded %d sent messages, want 0", len(sess.sent))
			}
		})
	}
}

// ─── AC-CRS-011 (REQ-CRS-006) — the candidate list is bounded and stable ───

func TestCodexTaskResumeScope_CandidatesBoundedAndOrdered(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)

	// Twelve distinct threads, one record each, written directly so the
	// timestamps are fixed. job-t0a and job-t0b share the NEWEST updated_at.
	base := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	write := func(id, thread string, at time.Time) {
		t.Helper()
		if err := reg.write(CodexJobRecord{
			ID: id, Status: codexJobStatusCompleted, CreatedAt: at, UpdatedAt: at,
			ThreadID: thread, Mode: codexTaskMode, RequestSummary: thread,
		}); err != nil {
			t.Fatalf("write %s: %v", id, err)
		}
	}
	write("job-t0b", "thr-tie-b", base.Add(time.Hour))
	write("job-t0a", "thr-tie-a", base.Add(time.Hour))
	for i := 0; i < 10; i++ {
		write("job-t1"+string(rune('a'+i)), "thr-"+string(rune('a'+i)), base.Add(time.Duration(i)*time.Minute))
	}

	call := func() map[string]any {
		sess := withCodexSession(t, codexTaskScript("trn-x", "work"))
		res := callCodexTask(t, map[string]any{"prompt": "p", "resume_last": true})
		if len(sess.sent) != 0 {
			t.Errorf("fixture recorded %d sent messages, want 0", len(sess.sent))
		}
		return assertRefusal(t, res, "resume_ambiguous")
	}
	first, second := call(), call()

	if first["candidate_total"] != float64(12) {
		t.Errorf("candidate_total = %v, want 12", first["candidate_total"])
	}
	ids := candidateThreadIDs(t, first)
	want := []string{"thr-tie-a", "thr-tie-b", "thr-j", "thr-i", "thr-h", "thr-g", "thr-f", "thr-e", "thr-d", "thr-c"}
	assertStrings(t, "candidates (updated_at desc, tie → record id asc, first 10)", ids, want)
	assertStrings(t, "second call candidates", candidateThreadIDs(t, second), ids)
}

// ─── AC-CRS-012 (REQ-CRS-005, 008) — thread_id wins; unused selectors listed ───

func TestCodexTaskResumeScope_ThreadIDPrecedenceAndUnusedSelectors(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)
	seedCodexJob(t, reg, "thr-5", "card-A", "card A")
	seedCodexJob(t, reg, "thr-6", "card-B", "card B")
	args := func(background bool) map[string]any {
		return map[string]any{"prompt": "p", "thread_id": "thr-6", "resume_last": true, "work_key": "card-A", "background": background}
	}
	check := func(label string, got map[string]any, resumed string) {
		t.Helper()
		if resumed != "thr-6" {
			t.Errorf("%s: thread/resume threadId = %q, want thr-6", label, resumed)
		}
		if got["resume_basis"] != "thread_id" {
			t.Errorf("%s: resume_basis = %v, want thread_id", label, got["resume_basis"])
		}
		raw, _ := got["unused_selectors"].([]any)
		var unused []string
		for _, u := range raw {
			s, _ := u.(string)
			unused = append(unused, s)
		}
		assertStrings(t, label+": unused_selectors", unused, []string{"resume_last", "work_key"})
	}

	// Foreground.
	sess := withCodexSession(t, codexTaskScriptEcho("thr-6", "trn-f", "work"))
	fg := structuredMap(t, callCodexTask(t, args(false)))
	check("foreground", fg, sentThreadResumeID(t, sess.sent))

	// Background: the resume is sent synchronously in the handler; the turn
	// runs in a goroutine, so the sent slice is read only after the job ends.
	sess = withCodexSession(t, codexTaskScriptEcho("thr-6", "trn-g", "work"))
	bg := structuredMap(t, callCodexTask(t, args(true)))
	jobID, _ := bg["job_id"].(string)
	if jobID == "" {
		t.Fatalf("background call returned no job_id: %v", bg)
	}
	rec := awaitTerminalJob(t, reg, jobID)
	check("background", bg, sentThreadResumeID(t, sess.sent))
	if rec.WorkKey != "card-A" {
		t.Errorf("background record work_key = %q, want card-A (recorded even though unused for selection)", rec.WorkKey)
	}
}

// ─── AC-CRS-013 (REQ-CRS-009) — the schema states the scoping ───

func TestCodexTaskResumeScope_SchemaDeclaresScoping(t *testing.T) {
	tool, ok := newMoaiMCPServer().ListTools()[codexTaskToolName]
	if !ok || tool == nil {
		t.Fatalf("%s is not registered", codexTaskToolName)
	}
	props := tool.Tool.InputSchema.Properties
	desc := func(name string) string {
		t.Helper()
		p, ok := props[name].(map[string]any)
		if !ok {
			t.Fatalf("property %q is not declared", name)
		}
		if p["type"] != "string" && name != "resume_last" {
			t.Errorf("property %q type = %v, want string", name, p["type"])
		}
		d, _ := p["description"].(string)
		return d
	}
	if d := desc("thread_id"); !strings.Contains(d, "recorded in this project's") {
		t.Errorf("thread_id description does not limit it to threads recorded in this project's registry: %q", d)
	}
	if d := desc("work_key"); !strings.Contains(d, "resume_last") || !strings.Contains(d, "records carrying") {
		t.Errorf("work_key description does not state that it scopes resume_last to records carrying it: %q", d)
	}
	d := desc("resume_last")
	if strings.Contains(d, "most recently recorded codex thread for this project") {
		t.Errorf("resume_last description still carries the unqualified recency wording: %q", d)
	}
	if !strings.Contains(d, "refused") || !strings.Contains(d, "more than one") {
		t.Errorf("resume_last description does not state the refusal on more than one recorded thread: %q", d)
	}
}

// ─── AC-CRS-014 (REQ-CRS-004, 006) — representative record across records ───

func TestCodexTaskResumeScope_RepresentativeRecordAcrossRecords(t *testing.T) {
	root := t.TempDir()
	withCodexProjectDir(t, root)
	reg := newCodexJobRegistry(root)
	seedCodexJob(t, reg, "thr-5", "card-A", "A first")
	seedCodexJob(t, reg, "thr-6", "card-B", "B first")
	seedCodexJob(t, reg, "thr-5", "card-C", "A resumed")
	fourth := seedCodexJob(t, reg, "thr-5", "", "A again")

	// (1) No selector: ambiguous, candidates from representative records.
	sess := withCodexSession(t, codexTaskScript("trn-x", "work"))
	got := assertRefusal(t, callCodexTask(t, map[string]any{"prompt": "p", "resume_last": true}), "resume_ambiguous")
	if len(sess.sent) != 0 {
		t.Errorf("(1) fixture recorded %d sent messages, want 0", len(sess.sent))
	}
	if got["candidate_total"] != float64(2) {
		t.Errorf("(1) candidate_total = %v, want 2", got["candidate_total"])
	}
	assertStrings(t, "(1) candidate order", candidateThreadIDs(t, got), []string{"thr-5", "thr-6"})
	cands := got["candidates"].([]any)
	c5, c6 := cands[0].(map[string]any), cands[1].(map[string]any)
	if c5["request_summary"] != "A again" {
		t.Errorf("(1) thr-5 request_summary = %v, want \"A again\" (the representative record)", c5["request_summary"])
	}
	if at, _ := time.Parse(time.RFC3339Nano, c5["updated_at"].(string)); !at.Equal(fourth.UpdatedAt) {
		t.Errorf("(1) thr-5 updated_at = %v, want the fourth record's %v", c5["updated_at"], fourth.UpdatedAt)
	}
	if _, ok := c5["work_key"]; ok {
		t.Errorf("(1) thr-5 candidate carries work_key %v; its representative record has none", c5["work_key"])
	}
	if c6["work_key"] != "card-B" {
		t.Errorf("(1) thr-6 candidate work_key = %v, want card-B", c6["work_key"])
	}

	// (2) and (3): work_key selects per record, not per representative.
	for _, key := range []string{"card-A", "card-C"} {
		sess := withCodexSession(t, codexTaskScriptEcho("thr-5", "trn-w", "work"))
		got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p", "resume_last": true, "work_key": key}))
		if id := sentThreadResumeID(t, sess.sent); id != "thr-5" {
			t.Errorf("work_key %s: thread/resume threadId = %q, want thr-5", key, id)
		}
		if got["resume_basis"] != "work_key" {
			t.Errorf("work_key %s: resume_basis = %v, want work_key", key, got["resume_basis"])
		}
	}
}

// ─── AC-CRS-015 (REQ-CRS-001) — the sent id is reported even when it did not take ───

func TestCodexTaskResumeScope_ReportsSentIDOnMismatchAndAckError(t *testing.T) {
	seed := func(t *testing.T) {
		t.Helper()
		root := t.TempDir()
		withCodexProjectDir(t, root)
		seedCodexJob(t, newCodexJobRegistry(root), "thr-5", "", "card A")
	}

	t.Run("codex returned a different id", func(t *testing.T) {
		seed(t)
		withCodexSession(t, codexTaskScript("trn-m", "work"))
		got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p", "resume_last": true}))
		if got["resume_thread_id"] != "thr-5" || got["resume_basis"] != "sole_thread" {
			t.Errorf("resume_thread_id/resume_basis = %v/%v, want thr-5/sole_thread", got["resume_thread_id"], got["resume_basis"])
		}
		if got["thread_id"] != "tid-fake" || got["resumed_thread"] != false {
			t.Errorf("thread_id/resumed_thread = %v/%v, want tid-fake/false", got["thread_id"], got["resumed_thread"])
		}
	})

	t.Run("thread ack rejected", func(t *testing.T) {
		seed(t)
		lines := codexTaskScript("trn-e", "work")
		lines[1] = `{"id":2,"error":{"code":-32600,"message":"rejected by fake"}}`
		withCodexSession(t, lines)
		got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p", "resume_last": true}))
		if got["status"] != codexJobStatusFailed {
			t.Errorf("status = %v, want failed", got["status"])
		}
		if got["resume_thread_id"] != "thr-5" || got["resume_basis"] != "sole_thread" {
			t.Errorf("resume_thread_id/resume_basis = %v/%v, want thr-5/sole_thread", got["resume_thread_id"], got["resume_basis"])
		}
	})
}

// ─── Added regression (no AC; plan.md §D) — a failure BEFORE thread/resume is
// written carries neither resume field (REQ-CRS-001 third sentence). ───

func TestCodexTaskResumeScope_PreSendFailureCarriesNoResumeFields(t *testing.T) {
	cases := map[string]func(t *testing.T) *fakeCodexSession{
		"session start fails": func(t *testing.T) *fakeCodexSession {
			sess := withCodexSession(t, nil)
			sess.startErr = errors.New("spawn failed")
			return sess
		},
		"initialize rejected": func(t *testing.T) *fakeCodexSession {
			lines := codexTaskScript("trn-p", "work")
			lines[0] = `{"id":1,"error":{"code":-32600,"message":"init rejected"}}`
			return withCodexSession(t, lines)
		},
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			withCodexProjectDir(t, root)
			seedCodexJob(t, newCodexJobRegistry(root), "thr-5", "", "card A")
			sess := setup(t)

			got := structuredMap(t, callCodexTask(t, map[string]any{"prompt": "p", "resume_last": true}))

			if n := countSentMethod(t, sess.sent, codexMethodThreadResume); n != 0 {
				t.Fatalf("precondition: thread/resume sent %d times, want 0 on a pre-send failure", n)
			}
			if got["status"] != codexJobStatusFailed {
				t.Errorf("status = %v, want failed", got["status"])
			}
			for _, key := range []string{"resume_thread_id", "resume_basis"} {
				if _, ok := got[key]; ok {
					t.Errorf("%s present (%v); want absent when thread/resume was never sent", key, got[key])
				}
			}
		})
	}
}
