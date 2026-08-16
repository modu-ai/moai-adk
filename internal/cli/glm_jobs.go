// Package cli — GLM (z.ai) task job registry.
//
// glm_jobs.go owns the durable per-job record for the GLM delegation family:
// glm_task (glm_task.go) writes it and the job-control tools (glm_job_status /
// glm_job_result / glm_job_cancel, glm_job_control.go) read it. It mirrors the
// codex job registry (codex_jobs.go) — the structural SSOT for the
// delegation-family job model — minus the fields the GLM HTTP backend has no
// counterpart for: a GLM task is ONE HTTPS request, not an interactive session
// over a spawned process, so there is no thread id, no turn id, and no pid.
//
// Three properties are load-bearing (mirrored from the codex registry):
//
//   - in-process lifetime. A background job is a goroutine inside the running
//     server, and the record carries no reattachment metadata: a record found
//     in a non-terminal status after a restart is stale by construction, not
//     resumable. Ownership is decided by glmLiveJobs membership, never by the
//     record's contents.
//
//   - secret hygiene. The record has no environment field by construction, and
//     the request summary is redacted before it is truncated (via the shared
//     codexJobSummary — the redaction+truncation discipline is
//     delegation-family-agnostic, and one function guarantees both families
//     get the same hygiene).
//
//   - atomic writes. Per-entity JSON files under .moai/state/glm-jobs/, one
//     file per job, persisted through writeFileAtomic so a concurrent reader
//     never observes a partial record.
//
// Nothing here invokes AskUserQuestion: a failure is a structured error value
// the tool layer renders through toolErr.
package cli

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// GLM job status enum. The set is closed: update refuses a value outside it
// rather than persisting an unrecognized lifecycle state. The values mirror
// the codex job status strings so a caller that learned one family's lifecycle
// reads the other's without translation.
const (
	glmJobStatusQueued    = "queued"
	glmJobStatusRunning   = "running"
	glmJobStatusCompleted = "completed"
	glmJobStatusFailed    = "failed"
	glmJobStatusCancelled = "cancelled"
)

const (
	// glmJobsSubdir is the .moai/state subdirectory holding one JSON file per
	// job — the GLM sibling of codexJobsSubdir. .moai/state/ is runtime state
	// and already gitignored, so no record reaches a distributed tree.
	glmJobsSubdir = "glm-jobs"

	// glmJobIDPrefix prefixes every generated job id. The format is identical
	// to the codex job id (job-<UTC compact timestamp>-<8 hex>): the timestamp
	// makes a directory listing chronological and the random suffix guarantees
	// uniqueness between two jobs started in the same second.
	glmJobIDPrefix = "job-"

	// glmJobFileMode is 0600: a job record carries a request summary and the
	// model's output — operator-scoped, not world-readable.
	glmJobFileMode = 0o600

	// glmJobDirMode is 0755 for the record directory: the operator's own tools
	// traverse it, and the records inside carry their own 0600 (above). Named
	// beside its file-mode sibling so the pair is read and changed together.
	glmJobDirMode = 0o755
)

// glmJobStatuses is the enum in declaration order, used by glmJobStatusValid.
var glmJobStatuses = []string{
	glmJobStatusQueued,
	glmJobStatusRunning,
	glmJobStatusCompleted,
	glmJobStatusFailed,
	glmJobStatusCancelled,
}

// glmJobStatusValid reports whether s is a member of the closed status enum.
func glmJobStatusValid(s string) bool {
	for _, known := range glmJobStatuses {
		if s == known {
			return true
		}
	}
	return false
}

// glmJobTerminal reports whether s is a status a job never leaves. It is what
// glm_job_result reads to decide whether an output exists yet, and what
// glm_job_cancel reads to decide there is nothing left to cancel.
func glmJobTerminal(s string) bool {
	switch s {
	case glmJobStatusCompleted, glmJobStatusFailed, glmJobStatusCancelled:
		return true
	}
	return false
}

// ─── the record ───

// GLMJobRecord is the durable per-job record. Field names are snake_case to
// match the sibling state artifacts in this package. It mirrors CodexJobRecord
// minus ThreadID / TurnID / PID (the GLM HTTP backend has none of them) plus
// Model (the resolved z.ai model id the task ran on — the record's only
// provenance for "which model produced this output").
//
// The record carries NO environment block, NO credential, and NO endpoint: a
// key must not be able to reach it even by accident, and there is nothing a
// later server lifetime could use to adopt the job.
type GLMJobRecord struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Model is the GLM model id the task was sent to (the caller's override or
	// the resolved default).
	Model string `json:"model"`

	// RequestSummary is a redacted, bounded description of what was asked.
	RequestSummary string `json:"request_summary"`

	// Output is the completed task output; Error is the failure reason. Both
	// are empty until the job reaches a terminal status.
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// glmJobSpec is the caller-supplied half of a new record. The registry owns
// the id, the status, and the timestamps; the caller owns everything else.
type glmJobSpec struct {
	Model          string
	RequestSummary string
}

// ─── structured errors ───

// errGLMJobNotFound is the sentinel behind glmJobNotFound. It exists so
// glm_job_status can return a structured not-found result rather than an
// error that aborts the tool call.
var errGLMJobNotFound = errors.New("glm job not found")

// glmJobNotFound reports whether err means "no record with that id".
func glmJobNotFound(err error) bool { return errors.Is(err, errGLMJobNotFound) }

// glmJobStateError is the structured error for a state directory that cannot
// be written (or read). It names the operation and the path so the tool layer
// can render an operator-actionable result through toolErr — never a panic,
// never a process abort.
type glmJobStateError struct {
	Op   string // "create" | "read" | "write" | "mkdir"
	Path string
	Err  error
}

func (e *glmJobStateError) Error() string {
	return fmt.Sprintf("glm job state %s failed at %s: %v", e.Op, e.Path, e.Err)
}

func (e *glmJobStateError) Unwrap() error { return e.Err }

// ─── the registry ───

// glmJobRegistry reads and writes the per-job records under one project's
// .moai/state/glm-jobs/. The mutex serializes this process's own
// read-modify-write cycles; cross-process safety comes from the atomic rename
// in writeFileAtomic, which is what actually guarantees a concurrent reader
// never observes a partial record.
type glmJobRegistry struct {
	dir string
	mu  sync.Mutex
}

// newGLMJobRegistry returns the registry rooted at projectDir. The directory
// is created lazily on first write, so constructing a registry never touches
// the disk and can never fail.
func newGLMJobRegistry(projectDir string) *glmJobRegistry {
	return &glmJobRegistry{dir: filepath.Join(projectDir, ".moai", "state", glmJobsSubdir)}
}

// pathFor returns the record file path for a job id.
func (r *glmJobRegistry) pathFor(id string) string {
	return filepath.Join(r.dir, id+".json")
}

// create writes a new record in the queued status and returns it.
func (r *glmJobRegistry) create(spec glmJobSpec) (GLMJobRecord, error) {
	id, err := newGLMJobID()
	if err != nil {
		return GLMJobRecord{}, &glmJobStateError{Op: "create", Path: r.dir, Err: err}
	}
	now := time.Now().UTC()
	rec := GLMJobRecord{
		ID:             id,
		Status:         glmJobStatusQueued,
		CreatedAt:      now,
		UpdatedAt:      now,
		Model:          spec.Model,
		RequestSummary: codexJobSummary(spec.RequestSummary), // shared redaction+truncation discipline
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.write(rec); err != nil {
		return GLMJobRecord{}, err
	}
	return rec, nil
}

// update applies mutate to the stored record and persists the result
// atomically. The status is validated AFTER mutation, so a transition to an
// unrecognized state is refused with the on-disk record left untouched.
func (r *glmJobRegistry) update(id string, mutate func(*GLMJobRecord)) (GLMJobRecord, error) {
	return r.mutateRecord(id, mutate, false)
}

// updateUnlessCancelled is update, except that a record already in the
// cancelled status is returned untouched and NO write is performed.
//
// The distinction is the point rather than an optimization (mirrored from the
// codex registry): a mutator can decline to CHANGE anything, but it cannot
// decline the write. So a job's goroutine finishing after glm_job_cancel
// returned would still rewrite the record — the status survived, but the file
// changed after the tool had reported the job cancelled, and a caller watching
// updated_at would see the record move under it. Only the registry can skip a
// write, so the guard belongs here.
func (r *glmJobRegistry) updateUnlessCancelled(id string, mutate func(*GLMJobRecord)) (GLMJobRecord, error) {
	return r.mutateRecord(id, mutate, true)
}

// mutateRecord is the locked read-mutate-write body shared by update and
// updateUnlessCancelled.
func (r *glmJobRegistry) mutateRecord(id string, mutate func(*GLMJobRecord), skipCancelled bool) (GLMJobRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec, err := r.read(id)
	if err != nil {
		return GLMJobRecord{}, err
	}
	if skipCancelled && rec.Status == glmJobStatusCancelled {
		return rec, nil
	}
	mutate(&rec)
	if !glmJobStatusValid(rec.Status) {
		return GLMJobRecord{}, fmt.Errorf("glm job %s: refusing status %q — not one of %s",
			id, rec.Status, strings.Join(glmJobStatuses, "/"))
	}
	rec.ID = id // the id is registry-owned; a mutator cannot rename a record
	rec.RequestSummary = codexJobSummary(rec.RequestSummary)
	rec.UpdatedAt = time.Now().UTC()
	if err := r.write(rec); err != nil {
		return GLMJobRecord{}, err
	}
	return rec, nil
}

// load returns the stored record for id.
func (r *glmJobRegistry) load(id string) (GLMJobRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.read(id)
}

// read loads and decodes one record. The caller holds r.mu.
func (r *glmJobRegistry) read(id string) (GLMJobRecord, error) {
	path := r.pathFor(id)
	raw, err := os.ReadFile(path) //nolint:gosec // path is registry-owned: <state dir>/<generated id>.json
	if err != nil {
		if os.IsNotExist(err) {
			return GLMJobRecord{}, fmt.Errorf("%w: %s", errGLMJobNotFound, id)
		}
		return GLMJobRecord{}, &glmJobStateError{Op: "read", Path: path, Err: err}
	}
	var rec GLMJobRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return GLMJobRecord{}, &glmJobStateError{Op: "read", Path: path, Err: err}
	}
	return rec, nil
}

// write persists one record atomically (temp file + rename via the consolidated
// writeFileAtomic helper), so a concurrent reader observes either the previous
// record or the new one — never a partial write. The caller holds r.mu.
func (r *glmJobRegistry) write(rec GLMJobRecord) error {
	if err := os.MkdirAll(r.dir, glmJobDirMode); err != nil {
		return &glmJobStateError{Op: "mkdir", Path: r.dir, Err: err}
	}
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return &glmJobStateError{Op: "write", Path: r.pathFor(rec.ID), Err: err}
	}
	path := r.pathFor(rec.ID)
	if err := writeFileAtomic(path, append(b, '\n'), glmJobFileMode); err != nil {
		return &glmJobStateError{Op: "write", Path: path, Err: err}
	}
	return nil
}

// ─── id generation ───

// newGLMJobID returns a job id of the form job-<UTC compact timestamp>-<8 hex>
// — the same format as the codex job id. The random suffix is what guarantees
// uniqueness; the timestamp only makes a directory listing readable.
func newGLMJobID() (string, error) {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("glm job id: %w", err)
	}
	return glmJobIDPrefix + time.Now().UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(buf[:]), nil
}
