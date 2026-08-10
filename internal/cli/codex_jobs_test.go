package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-CODEX-PHASE2-001 M2 — job registry (REQ-CX2-003, REQ-CX2-004,
// REQ-CX2-005). Verifies AC-CX2-005 through AC-CX2-008 plus the §B edge cases
// that bind the registry (two concurrent jobs, malformed record file).
//
// The registry is the DURABLE half of the job surface; the codex_task /
// codex_job_* tools that drive it are M3/M4. These tests therefore exercise the
// registry and the session seams it consumes (turn-started capture, spawned
// pid) directly, against the canned session runner the M1 tests already use.

// fakeCodexConnPID is the sentinel pid the canned connection reports, standing
// in for the pid of a codex subprocess this server spawned. It is deliberately
// implausible as a live pid so a test asserting on it can never be satisfied by
// an accidentally-real process.
const fakeCodexConnPID = 424242

// pid makes the canned connection satisfy codexProcessConn, so the session
// handle can report a spawned-process pid without spawning a process.
func (c *fakeCodexConn) pid() int { return fakeCodexConnPID }

// codexTurnStartedScript is codexSessionScript plus the turn/started
// notification REQ-CX2-003 names as the ONLY source of the turnId that
// turn/interrupt requires (M0 probe, progress.md §E.2 (a)).
func codexTurnStartedScript(turnID, reviewText string) []string {
	q := func(s string) string { return jsonString(s) }
	return []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":` + q(turnID) + `,"status":"inProgress"}}}`,
		`{"method":"turn/started","params":{"threadId":"tid-fake","turn":{"id":` + q(turnID) + `,"status":"inProgress"}}}`,
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":` + q(turnID) + `,"completedAtMs":1,"item":{"type":"exitedReviewMode","id":"e1","review":` + q(reviewText) + `}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":` + q(turnID) + `,"status":"completed"}}}`,
	}
}

// readJobFiles returns the names of every file in the registry directory.
func readJobFiles(t *testing.T, reg *codexJobRegistry) []string {
	t.Helper()
	entries, err := os.ReadDir(reg.dir)
	if err != nil {
		t.Fatalf("read job dir %s: %v", reg.dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// AC-CX2-005 (REQ-CX2-003) — a started background job leaves exactly one record
// file whose JSON decodes with every required field, including a turnId equal
// to the turn.id carried by the canned session's turn/started notification and
// the pid of the codex process this server spawned for the job.
func TestCodexJobRegistry_RecordShapeAndTurnIDCapture(t *testing.T) {
	root := t.TempDir()
	reg := newCodexJobRegistry(root)
	withCodexSession(t, codexTurnStartedScript("trn-77", "clean, no findings"))

	handle, err := openCodexSession(context.Background(), "/fake/codex", map[string]any{"cwd": root})
	if err != nil {
		t.Fatalf("openCodexSession: %v", err)
	}
	defer func() { _ = handle.close() }()

	rec, err := reg.create(codexJobSpec{
		ThreadID:       handle.threadID,
		PID:            handle.pid(),
		Mode:           codexModeAdversarial,
		RequestSummary: "review the uncommitted changes",
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	// The mid-flight hook the job goroutine installs: turnId is persisted as
	// soon as turn/started is observed, BEFORE the turn completes.
	handle.onTurnStarted = reg.turnIDRecorder(rec.ID)
	if _, err := handle.runTurn(context.Background(), codexMethodReviewStart, map[string]any{
		"target": codexTargetUncommitted,
	}); err != nil {
		t.Fatalf("runTurn: %v", err)
	}

	if files := readJobFiles(t, reg); len(files) != 1 {
		t.Fatalf("job dir holds %d files (%v), want exactly 1", len(files), files)
	}

	raw, err := os.ReadFile(reg.pathFor(rec.ID))
	if err != nil {
		t.Fatalf("read record: %v", err)
	}
	var got CodexJobRecord
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("record does not decode as JSON: %v\n%s", err, raw)
	}

	if got.ID == "" {
		t.Error("record id is empty")
	}
	if got.Status != codexJobStatusQueued {
		t.Errorf("status = %q, want %q on creation", got.Status, codexJobStatusQueued)
	}
	if got.CreatedAt.IsZero() {
		t.Error("created_at is zero")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("updated_at is zero")
	}
	if got.ThreadID != "tid-fake" {
		t.Errorf("thread_id = %q, want tid-fake", got.ThreadID)
	}
	if got.TurnID != "trn-77" {
		t.Errorf("turn_id = %q, want trn-77 (the turn/started notification's turn.id)", got.TurnID)
	}
	if got.PID != fakeCodexConnPID {
		t.Errorf("pid = %d, want %d (the pid of the process this server spawned)", got.PID, fakeCodexConnPID)
	}
	if got.Mode != codexModeAdversarial {
		t.Errorf("mode = %q, want %q", got.Mode, codexModeAdversarial)
	}
	if got.RequestSummary == "" {
		t.Error("request_summary is empty")
	}
}

// The record carries NO reattachment metadata: background jobs are in-process
// (plan.md §D M0 decision), so a record found non-terminal after a restart is
// stale by construction rather than resumable. Asserted on the serialized form
// so a future field addition that reintroduces reattachment is caught here.
func TestCodexJobRecord_NoReattachmentMetadata(t *testing.T) {
	reg := newCodexJobRegistry(t.TempDir())
	rec, err := reg.create(codexJobSpec{ThreadID: "t", PID: 1, Mode: codexModeNative, RequestSummary: "x"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	raw, err := os.ReadFile(reg.pathFor(rec.ID))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, forbidden := range []string{"binary_path", "argv", "env", "server_pid", "resumable", "reattach"} {
		if _, present := fields[forbidden]; present {
			t.Errorf("record carries %q — background jobs are in-process and are never adopted by a later server lifetime", forbidden)
		}
	}
}

// AC-CX2-006 (REQ-CX2-004) — every transition writes atomically: a concurrent
// reader observes a complete record carrying a status from the enum at every
// read, never a truncated or partially-written file.
func TestCodexJobRegistry_TransitionsAreAtomic(t *testing.T) {
	reg := newCodexJobRegistry(t.TempDir())
	rec, err := reg.create(codexJobSpec{ThreadID: "tid", PID: 7, Mode: codexModeNative, RequestSummary: "seed"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	transitions := []string{
		codexJobStatusRunning,
		codexJobStatusRunning,
		codexJobStatusCompleted,
		codexJobStatusFailed,
		codexJobStatusCancelled,
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})
	reads := 0
	var readErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			// Yield between reads: the assertion needs many reads across the
			// rename window, not a hot spin that starves the rest of the
			// package's tests.
			time.Sleep(50 * time.Microsecond)
			raw, err := os.ReadFile(reg.pathFor(rec.ID))
			if err != nil {
				continue // the rename window is not a read failure the caller sees
			}
			var got CodexJobRecord
			if err := json.Unmarshal(raw, &got); err != nil {
				readErr = err // a partial write would land here
				return
			}
			if !codexJobStatusValid(got.Status) {
				readErr = fmt.Errorf("read a status outside the enum: %q", got.Status)
				return
			}
			if got.ID != rec.ID {
				readErr = fmt.Errorf("read id %q, want %q", got.ID, rec.ID)
				return
			}
			reads++
		}
	}()

	for _, status := range transitions {
		if _, err := reg.update(rec.ID, func(r *CodexJobRecord) { r.Status = status }); err != nil {
			t.Fatalf("update to %s: %v", status, err)
		}
	}
	close(stop)
	wg.Wait()

	if readErr != nil {
		t.Fatalf("a concurrent read observed a partial or invalid record: %v", readErr)
	}
	if reads == 0 {
		t.Fatal("the concurrent reader never completed a read — the assertion proved nothing")
	}

	final, err := reg.load(rec.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if final.Status != codexJobStatusCancelled {
		t.Errorf("final status = %q, want %q", final.Status, codexJobStatusCancelled)
	}
	if !final.UpdatedAt.After(final.CreatedAt) && final.UpdatedAt != final.CreatedAt {
		t.Error("updated_at moved backwards relative to created_at")
	}
}

// The status enum is closed: an update that would write a value outside
// {queued,running,completed,failed,cancelled} is refused rather than persisted.
func TestCodexJobRegistry_RejectsUnknownStatus(t *testing.T) {
	reg := newCodexJobRegistry(t.TempDir())
	rec, err := reg.create(codexJobSpec{ThreadID: "tid", PID: 7, Mode: codexModeNative, RequestSummary: "seed"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := reg.update(rec.ID, func(r *CodexJobRecord) { r.Status = "eaten-by-a-grue" }); err == nil {
		t.Fatal("update accepted a status outside the enum, want an error")
	}
	after, err := reg.load(rec.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if after.Status != codexJobStatusQueued {
		t.Errorf("a refused transition still mutated the record: status = %q", after.Status)
	}
}

// AC-CX2-007 (REQ-CX2-004) — an unwritable state directory yields a STRUCTURED
// error the caller can render as a tool result, not a panic and not a process
// abort; the registry remains usable for a subsequent call once the obstruction
// is gone. (The tool-level half of this AC — that codex_task returns that
// structured error as its result — lands with codex_task in M3.)
func TestCodexJobRegistry_UnwritableStateDirStructuredError(t *testing.T) {
	root := t.TempDir()
	// Occupy .moai/state with a regular FILE so the codex-jobs directory can
	// never be created. Portable across POSIX and Windows (no chmod, which is
	// a no-op on Windows and is bypassed when the test runs as root).
	moai := filepath.Join(root, ".moai")
	if err := os.MkdirAll(moai, 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	statePath := filepath.Join(moai, "state")
	if err := os.WriteFile(statePath, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write state file: %v", err)
	}

	reg := newCodexJobRegistry(root)
	_, err := reg.create(codexJobSpec{ThreadID: "tid", PID: 7, Mode: codexModeNative, RequestSummary: "seed"})
	if err == nil {
		t.Fatal("create succeeded against an unwritable state dir, want a structured error")
	}
	var stateErr *codexJobStateError
	if !errors.As(err, &stateErr) {
		t.Fatalf("error is %T, want *codexJobStateError so the tool layer can render it structurally", err)
	}
	if stateErr.Path == "" || stateErr.Op == "" {
		t.Errorf("structured error is missing operator-facing context: op=%q path=%q", stateErr.Op, stateErr.Path)
	}
	if stateErr.Unwrap() == nil {
		t.Error("structured error drops its cause")
	}

	// The server survives: remove the obstruction and the very same registry
	// serves the next call.
	if err := os.Remove(statePath); err != nil {
		t.Fatalf("remove obstruction: %v", err)
	}
	if _, err := reg.create(codexJobSpec{ThreadID: "tid", PID: 7, Mode: codexModeNative, RequestSummary: "seed"}); err != nil {
		t.Fatalf("registry unusable after a recovered state-dir failure: %v", err)
	}
}

// AC-CX2-008 (REQ-CX2-005) — no credential reaches the record. Sentinels are
// seeded into BOTH the process environment and the request text; neither
// appears in the file, and no serialized environment block is present.
func TestCodexJobRecord_CarriesNoSecrets(t *testing.T) {
	const (
		envSentinel     = "sk-live-ENVSENTINEL0000000000"
		requestSentinel = "sk-proj-REQSENTINEL1111111111"
		bearerSentinel  = "BEARERSENTINEL2222222222"
		passwordValue   = "PASSWORDSENTINEL333"
	)
	t.Setenv("CODEX_API_KEY", envSentinel)
	t.Setenv("OPENAI_API_KEY", envSentinel)

	reg := newCodexJobRegistry(t.TempDir())
	rec, err := reg.create(codexJobSpec{
		ThreadID: "tid",
		PID:      7,
		Mode:     codexModeAdversarial,
		RequestSummary: "audit the diff using api_key=" + requestSentinel +
			" and Authorization: Bearer " + bearerSentinel +
			" (password: " + passwordValue + ")",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	raw, err := os.ReadFile(reg.pathFor(rec.ID))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	body := string(raw)
	for _, sentinel := range []string{envSentinel, requestSentinel, bearerSentinel, passwordValue} {
		if strings.Contains(body, sentinel) {
			t.Errorf("record leaks the sentinel %q:\n%s", sentinel, body)
		}
	}
	// The surviving summary must still say something about the request.
	if !strings.Contains(rec.RequestSummary, "audit the diff") {
		t.Errorf("redaction destroyed the summary entirely: %q", rec.RequestSummary)
	}
}

// The request summary is bounded, so an enormous prompt cannot inflate the
// record. The ceiling is the single Go-defined threshold (REQ-CX2-015).
func TestCodexJobRecord_RequestSummaryBounded(t *testing.T) {
	reg := newCodexJobRegistry(t.TempDir())
	rec, err := reg.create(codexJobSpec{
		ThreadID:       "tid",
		PID:            7,
		Mode:           codexModeNative,
		RequestSummary: strings.Repeat("a", 10_000),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	bound := config.DefaultCodexJobSummaryMaxLen + len(codexJobSummaryEllipsis)
	if len(rec.RequestSummary) > bound {
		t.Errorf("summary length %d exceeds the bound %d", len(rec.RequestSummary), bound)
	}
}

// §B edge case — two concurrent background jobs in one project produce two
// distinct ids and two distinct record files, neither overwriting the other.
func TestCodexJobRegistry_ConcurrentJobsDoNotCollide(t *testing.T) {
	reg := newCodexJobRegistry(t.TempDir())
	const n = 8

	var wg sync.WaitGroup
	ids := make([]string, n)
	errs := make([]error, n)
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rec, err := reg.create(codexJobSpec{ThreadID: "tid", PID: 7, Mode: codexModeNative, RequestSummary: "job"})
			if err != nil {
				errs[i] = err
				return
			}
			ids[i] = rec.ID
		}(i)
	}
	wg.Wait()

	seen := map[string]bool{}
	for i, id := range ids {
		if errs[i] != nil {
			t.Fatalf("create %d: %v", i, errs[i])
		}
		if seen[id] {
			t.Fatalf("duplicate job id %q — two jobs would share one record file", id)
		}
		seen[id] = true
	}
	if files := readJobFiles(t, reg); len(files) != n {
		t.Errorf("job dir holds %d files, want %d (one per job)", len(files), n)
	}
}

// §B edge case — a present-but-malformed record file is reported as unreadable
// rather than crashing the caller.
func TestCodexJobRegistry_MalformedRecordIsReported(t *testing.T) {
	reg := newCodexJobRegistry(t.TempDir())
	rec, err := reg.create(codexJobSpec{ThreadID: "tid", PID: 7, Mode: codexModeNative, RequestSummary: "seed"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := os.WriteFile(reg.pathFor(rec.ID), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("corrupt record: %v", err)
	}
	if _, err := reg.load(rec.ID); err == nil {
		t.Fatal("load accepted a malformed record, want a structured error")
	}
}

// An unknown job id is a not-found condition, distinguishable from a decode
// failure so codex_job_status (M4) can report it as such.
func TestCodexJobRegistry_UnknownIDIsNotFound(t *testing.T) {
	reg := newCodexJobRegistry(t.TempDir())
	_, err := reg.load("job-does-not-exist")
	if err == nil {
		t.Fatal("load of an unknown id succeeded, want a not-found error")
	}
	if !codexJobNotFound(err) {
		t.Errorf("error %v is not classified as not-found", err)
	}
}

// codexConnPID reports 0 for a connection with no backing process, so a record
// written against a processless session never names someone else's pid.
func TestCodexConnPID_ZeroWithoutProcess(t *testing.T) {
	if got := codexConnPID(processlessCodexConn{}); got != 0 {
		t.Errorf("codexConnPID(processless) = %d, want 0", got)
	}
}

// processlessCodexConn is a codexConn that does NOT implement codexProcessConn.
type processlessCodexConn struct{}

func (processlessCodexConn) send(string) error    { return nil }
func (processlessCodexConn) recv() (string, bool) { return "", false }
func (processlessCodexConn) close() error         { return nil }
