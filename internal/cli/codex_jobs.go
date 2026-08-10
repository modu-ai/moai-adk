// Package cli — SPEC-CODEX-PHASE2-001 M2 codex job registry.
//
// codex_jobs.go owns the durable per-job record that the codex task surface
// (codex_task, M3) writes and the job-control tools (codex_job_status /
// codex_job_result / codex_job_cancel, M4) read. It follows the
// .moai/state/audit-multi/<session>.json precedent (mcp_convergence.go
// persistConvergenceResult) — per-entity JSON files under .moai/state/, no
// shared index, so two sessions on one checkout never contend for a single file
// (spec.md §H R4).
//
// Three properties are load-bearing:
//
//   - turnId (REQ-CX2-003). turn/interrupt requires {threadId, turnId} and the
//     turnId is obtainable ONLY from the turn/started notification's turn.id
//     (M0 probe, progress.md §E.2 (a)). The registry therefore exposes a
//     mid-flight recorder (turnIDRecorder) the job goroutine installs on its
//     session handle, so the id lands in the record while the turn is still
//     running and therefore still cancellable. A record whose turn was never
//     observed starting carries an empty turnId, and the cancel path must treat
//     that as "not addressable" rather than sending a malformed interrupt.
//
//   - in-process lifetime (plan.md §D M0 decision). A background job is a
//     goroutine inside the running server, so the recorded pid is always one
//     THIS process spawned in THIS lifetime. The record deliberately carries no
//     reattachment metadata: a record found in a non-terminal status after a
//     restart is stale by construction, not resumable.
//
//   - secret hygiene (REQ-CX2-005). The record has no environment field by
//     construction, and the request summary is redacted before it is truncated.
//
// Nothing here invokes AskUserQuestion: a failure is a structured error value
// the tool layer renders through toolErr (REQ-CX2-014 / C2).
//
// @MX:SPEC: SPEC-CODEX-PHASE2-001
package cli

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// codex job status enum (REQ-CX2-004). The set is closed: update refuses a
// value outside it rather than persisting an unrecognized lifecycle state.
const (
	codexJobStatusQueued    = "queued"
	codexJobStatusRunning   = "running"
	codexJobStatusCompleted = "completed"
	codexJobStatusFailed    = "failed"
	codexJobStatusCancelled = "cancelled"
)

const (
	// codexJobsSubdir is the .moai/state subdirectory holding one JSON file per
	// job. .moai/state/ is runtime state and already gitignored, so no record
	// ever reaches a distributed user's tree (spec.md § Out of Scope — no new
	// template or distributed config surface).
	codexJobsSubdir = "codex-jobs"

	// codexJobIDPrefix prefixes every generated job id. The timestamp component
	// makes a directory listing chronological; the random component is what
	// actually guarantees uniqueness between two jobs started in the same
	// second (spec.md §H R4 / acceptance.md §B two-concurrent-jobs).
	codexJobIDPrefix = "job-"

	// codexJobSummaryEllipsis marks a request summary truncated at the bound.
	codexJobSummaryEllipsis = "…"

	// codexJobRedacted replaces a credential-shaped value in a request summary.
	codexJobRedacted = "[redacted]"

	// codexJobFileMode is 0600: a job record names a thread, a turn, and a pid
	// of a process running on this machine — operator-scoped, not world-readable.
	codexJobFileMode = 0o600
)

// codexJobStatuses is the enum in declaration order, used by codexJobStatusValid.
var codexJobStatuses = []string{
	codexJobStatusQueued,
	codexJobStatusRunning,
	codexJobStatusCompleted,
	codexJobStatusFailed,
	codexJobStatusCancelled,
}

// codexJobStatusValid reports whether s is a member of the closed status enum.
func codexJobStatusValid(s string) bool {
	for _, known := range codexJobStatuses {
		if s == known {
			return true
		}
	}
	return false
}

// codexJobTerminal reports whether s is a status a job never leaves. It is what
// codex_job_result reads to decide whether an output exists yet, and what
// codex_job_cancel reads to decide there is nothing left to cancel (REQ-CX2-010,
// REQ-CX2-011).
func codexJobTerminal(s string) bool {
	switch s {
	case codexJobStatusCompleted, codexJobStatusFailed, codexJobStatusCancelled:
		return true
	}
	return false
}

// ─── the record ───

// CodexJobRecord is the durable per-job record (REQ-CX2-003). Field names are
// snake_case to match the sibling state artifacts in this package
// (ConvergenceResult's per_backend_verdicts / next_steps); the protocol's own
// camelCase names appear only on the wire, never here.
//
// The record carries NO environment block, NO argv, and NO binary path: a
// credential must not be able to reach it even by accident (REQ-CX2-005), and
// there is nothing a later server lifetime could use to adopt the job.
type CodexJobRecord struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// ThreadID and TurnID are the two arguments turn/interrupt requires. TurnID
	// is empty until the turn/started notification is observed.
	ThreadID string `json:"thread_id"`
	TurnID   string `json:"turn_id"`

	// PID is the codex process THIS server spawned for the job in the current
	// process lifetime, or 0 when the session has no backing process.
	PID int `json:"pid"`

	// Mode is the requested codex mode (codexModeNative / codexModeAdversarial).
	Mode string `json:"mode"`

	// RequestSummary is a redacted, bounded description of what was asked.
	RequestSummary string `json:"request_summary"`

	// Output is the completed task output; Error is the failure reason. Both are
	// empty until the job reaches a terminal status.
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// codexJobSpec is the caller-supplied half of a new record. The registry owns
// the id, the status, and the timestamps; the caller owns everything else.
type codexJobSpec struct {
	ThreadID       string
	TurnID         string
	PID            int
	Mode           string
	RequestSummary string
}

// ─── structured errors ───

// errCodexJobNotFound is the sentinel behind codexJobNotFound. It exists so
// codex_job_status (M4) can return a structured not-found result rather than an
// error that aborts the tool call (REQ-CX2-009).
var errCodexJobNotFound = errors.New("codex job not found")

// codexJobNotFound reports whether err means "no record with that id".
func codexJobNotFound(err error) bool { return errors.Is(err, errCodexJobNotFound) }

// codexJobStateError is the structured error REQ-CX2-004 requires when the
// state directory cannot be written (or read). It names the operation and the
// path so the tool layer can render an operator-actionable result through
// toolErr — never a panic, never a process abort.
type codexJobStateError struct {
	Op   string // "create" | "read" | "write" | "mkdir"
	Path string
	Err  error
}

func (e *codexJobStateError) Error() string {
	return fmt.Sprintf("codex job state %s failed at %s: %v", e.Op, e.Path, e.Err)
}

func (e *codexJobStateError) Unwrap() error { return e.Err }

// ─── the registry ───

// codexJobRegistry reads and writes the per-job records under one project's
// .moai/state/codex-jobs/. The mutex serializes this process's own
// read-modify-write cycles; cross-process safety comes from the atomic rename
// in writeFileAtomic, which is what actually guarantees a concurrent reader
// never observes a partial record (REQ-CX2-004).
type codexJobRegistry struct {
	dir string
	mu  sync.Mutex
}

// newCodexJobRegistry returns the registry rooted at projectDir. The directory
// is created lazily on first write, so constructing a registry never touches
// the disk and can never fail.
func newCodexJobRegistry(projectDir string) *codexJobRegistry {
	return &codexJobRegistry{dir: filepath.Join(projectDir, ".moai", "state", codexJobsSubdir)}
}

// pathFor returns the record file path for a job id.
func (r *codexJobRegistry) pathFor(id string) string {
	return filepath.Join(r.dir, id+".json")
}

// create writes a new record in the queued status and returns it.
func (r *codexJobRegistry) create(spec codexJobSpec) (CodexJobRecord, error) {
	id, err := newCodexJobID()
	if err != nil {
		return CodexJobRecord{}, &codexJobStateError{Op: "create", Path: r.dir, Err: err}
	}
	now := time.Now().UTC()
	rec := CodexJobRecord{
		ID:             id,
		Status:         codexJobStatusQueued,
		CreatedAt:      now,
		UpdatedAt:      now,
		ThreadID:       spec.ThreadID,
		TurnID:         spec.TurnID,
		PID:            spec.PID,
		Mode:           spec.Mode,
		RequestSummary: codexJobSummary(spec.RequestSummary),
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.write(rec); err != nil {
		return CodexJobRecord{}, err
	}
	return rec, nil
}

// update applies mutate to the stored record and persists the result
// atomically. The status is validated AFTER mutation, so a transition to an
// unrecognized state is refused with the on-disk record left untouched.
func (r *codexJobRegistry) update(id string, mutate func(*CodexJobRecord)) (CodexJobRecord, error) {
	return r.mutateRecord(id, mutate, false)
}

// updateUnlessCancelled is update, except that a record already in the
// cancelled status is returned untouched and NO write is performed.
//
// The distinction is the point rather than an optimization. A mutator can
// decline to CHANGE anything, but it cannot decline the write: update persists
// whatever the mutator leaves behind and refreshes UpdatedAt regardless. So a
// job's goroutine finishing after codex_job_cancel returned still rewrote the
// record — the status survived (the mutator's own guard saw it), but the file
// changed after the tool had reported the job cancelled, and a caller watching
// UpdatedAt would see the record move under it. Only the registry can skip a
// write, so the guard belongs here (SPEC-CODEX-PHASE2-001 REQ-CX2-004 /
// REQ-CX2-011).
func (r *codexJobRegistry) updateUnlessCancelled(id string, mutate func(*CodexJobRecord)) (CodexJobRecord, error) {
	return r.mutateRecord(id, mutate, true)
}

// mutateRecord is the locked read-mutate-write body shared by update and
// updateUnlessCancelled.
func (r *codexJobRegistry) mutateRecord(id string, mutate func(*CodexJobRecord), skipCancelled bool) (CodexJobRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, err := r.read(id)
	if err != nil {
		return CodexJobRecord{}, err
	}
	if skipCancelled && rec.Status == codexJobStatusCancelled {
		return rec, nil
	}
	mutate(&rec)
	if !codexJobStatusValid(rec.Status) {
		return CodexJobRecord{}, fmt.Errorf("codex job %s: refusing status %q — not one of %s",
			id, rec.Status, strings.Join(codexJobStatuses, "/"))
	}
	rec.ID = id // the id is registry-owned; a mutator cannot rename a record
	rec.RequestSummary = codexJobSummary(rec.RequestSummary)
	rec.UpdatedAt = time.Now().UTC()
	if err := r.write(rec); err != nil {
		return CodexJobRecord{}, err
	}
	return rec, nil
}

// load returns the stored record for id.
func (r *codexJobRegistry) load(id string) (CodexJobRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.read(id)
}

// turnIDRecorder returns the mid-flight observer a background job installs on
// its session handle (codexSessionHandle.onTurnStarted): the turnId is
// persisted the moment turn/started is observed, while the turn is still
// running and therefore still cancellable (REQ-CX2-003, plan.md §D M2).
//
// The write is best-effort by construction — it runs inside the session's
// notification loop, which has no caller to return an error to, and a failure
// there must not abort a turn that is otherwise proceeding. The consequence is
// stated rather than hidden: the record keeps an empty turn_id, and the M4
// cancel path treats an empty turn_id as "turn not addressable" instead of
// sending a malformed interrupt.
func (r *codexJobRegistry) turnIDRecorder(jobID string) func(string) {
	return func(turnID string) {
		if turnID == "" {
			return
		}
		_, _ = r.update(jobID, func(rec *CodexJobRecord) { rec.TurnID = turnID })
	}
}

// latestThreadID returns the threadId of the most recently updated record that
// carries one — "the most recently recorded threadId for the project" that
// resume_last reuses (REQ-CX2-008). The second return is false when no record
// carries a thread, which is the case the caller must report rather than paper
// over by silently opening a new thread without saying so.
//
// Records are the only place a thread is recorded, and REQ-CX2-003 creates them
// for BACKGROUND jobs, so this resumes the last background job's thread. A
// project that has only ever run foreground tasks has nothing to resume, and
// says so.
//
// An unreadable directory or an undecodable record is skipped rather than
// surfaced: failing to find a thread to resume is a reportable outcome, not an
// error — the caller opens a new thread either way.
func (r *codexJobRegistry) latestThreadID() (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return "", false
	}
	var (
		bestID string
		bestAt time.Time
	)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(r.dir, e.Name())) //nolint:gosec // registry-owned directory listing
		if err != nil {
			continue
		}
		var rec CodexJobRecord
		if err := json.Unmarshal(raw, &rec); err != nil {
			continue
		}
		if rec.ThreadID == "" {
			continue
		}
		if bestID == "" || rec.UpdatedAt.After(bestAt) {
			bestID, bestAt = rec.ThreadID, rec.UpdatedAt
		}
	}
	return bestID, bestID != ""
}

// read loads and decodes one record. The caller holds r.mu.
func (r *codexJobRegistry) read(id string) (CodexJobRecord, error) {
	path := r.pathFor(id)
	raw, err := os.ReadFile(path) //nolint:gosec // path is registry-owned: <state dir>/<generated id>.json
	if err != nil {
		if os.IsNotExist(err) {
			return CodexJobRecord{}, fmt.Errorf("%w: %s", errCodexJobNotFound, id)
		}
		return CodexJobRecord{}, &codexJobStateError{Op: "read", Path: path, Err: err}
	}
	var rec CodexJobRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return CodexJobRecord{}, &codexJobStateError{Op: "read", Path: path, Err: err}
	}
	return rec, nil
}

// write persists one record atomically (temp file + rename via the consolidated
// writeFileAtomic helper), so a concurrent reader observes either the previous
// record or the new one — never a partial write (REQ-CX2-004). The caller holds
// r.mu.
func (r *codexJobRegistry) write(rec CodexJobRecord) error {
	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return &codexJobStateError{Op: "mkdir", Path: r.dir, Err: err}
	}
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return &codexJobStateError{Op: "write", Path: r.pathFor(rec.ID), Err: err}
	}
	path := r.pathFor(rec.ID)
	if err := writeFileAtomic(path, append(b, '\n'), codexJobFileMode); err != nil {
		return &codexJobStateError{Op: "write", Path: path, Err: err}
	}
	return nil
}

// ─── id generation ───

// newCodexJobID returns a job id of the form job-<UTC compact timestamp>-<8 hex>.
// The random suffix is what guarantees uniqueness; the timestamp only makes a
// directory listing readable.
func newCodexJobID() (string, error) {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("codex job id: %w", err)
	}
	return codexJobIDPrefix + time.Now().UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(buf[:]), nil
}

// ─── secret hygiene (REQ-CX2-005) ───

// codexJobBearerToken matches an HTTP bearer credential. It runs BEFORE the
// key/value pattern below, because "Authorization: Bearer <token>" would
// otherwise redact only the literal word "Bearer" and leave the token.
var codexJobBearerToken = regexp.MustCompile(`(?i)\bbearer\s+\S+`)

// codexJobSecretKeyValue matches a credential-shaped key followed by its value
// in either `k=v` or `k: v` form. The key and separator are preserved so the
// summary still says WHAT was redacted.
var codexJobSecretKeyValue = regexp.MustCompile(
	`(?i)\b(api[_-]?key|apikey|access[_-]?token|auth[_-]?token|token|secret|password|passwd|authorization)\b(\s*[:=]\s*)\S+`)

// codexJobSecretToken matches credentials recognizable by their own prefix,
// which appear with no key beside them.
var codexJobSecretToken = regexp.MustCompile(
	`(?i)\b(sk|rk)-[A-Za-z0-9_\-]{8,}|\bgh[pousr]_[A-Za-z0-9]{16,}|\bxox[abprs]-[A-Za-z0-9\-]{8,}|\bAKIA[0-9A-Z]{16}\b`)

// codexJobSummary redacts credential-shaped content from a request description
// and then bounds its length. Order matters: truncating first could split a
// credential across the boundary and leave a usable prefix in the record.
func codexJobSummary(s string) string {
	s = strings.TrimSpace(s)
	s = codexJobBearerToken.ReplaceAllString(s, codexJobRedacted)
	s = codexJobSecretKeyValue.ReplaceAllString(s, "${1}${2}"+codexJobRedacted)
	s = codexJobSecretToken.ReplaceAllString(s, codexJobRedacted)
	if len(s) > config.DefaultCodexJobSummaryMaxLen {
		s = strings.TrimSpace(s[:config.DefaultCodexJobSummaryMaxLen]) + codexJobSummaryEllipsis
	}
	return s
}
