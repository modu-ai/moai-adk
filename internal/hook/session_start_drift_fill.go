package hook

// session_start_drift_fill.go — the out-of-band drift-cache fill
// (SPEC-DRIFT-CACHE-FILL-001, card t871).
//
// The defect this closes, measured rather than reasoned: SessionStart's
// deferred advisory step computes drift and caches it keyed on HEAD, but the
// save happens AFTER the compute, the compute is strictly longer than the
// 250 ms bounded join, and the hook is a short-lived CLI process that exits
// when Handle returns. The cache therefore has no writer on this path at all —
// 15 of 15 real hook runs left it absent — so every cold session paid the full
// join and discarded the result, and the drift advisory never returned after a
// HEAD change.
//
// The fix is a writer that OUTLIVES the hook process: on a miss the handler
// starts one detached, self-bounded child and returns without awaiting it. The
// next session gets a hit.
//
// Two coordination layers, and neither replaces the other:
//
//   - The suppression RECORD, claimed HERE by the PARENT before the spawn. It
//     is the cheap spawn-suppressor. The parent claims because it is the only
//     actor guaranteed to be alive at that instant: a child that dies before
//     executing anything must still suppress the next session's respawn.
//   - The fill LOCK, taken by the CHILD (internal/spec.WithDriftFillLock). It
//     is the authoritative compute-serialiser.
//
// The record can be raced; the lock cannot. A record alone admits a burst; a
// lock alone admits a process fan-out. Both are needed.

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/spec"
)

// driftFillRecordFilename is the suppression record, in the same
// .moai/state/ runtime-state directory as the cache it guards the fill of.
const driftFillRecordFilename = "drift-fill.json"

// driftFillLockSuffix names the record's COMPANION lock. The claim target is
// this companion, never the record itself, and that distinction is load-bearing
// twice over:
//
//   - Reclaim TOCTOU. Removing an expired record outside a lock can destroy
//     another handler's fresh claim: A reclaims and wins; B, judging on a read
//     taken before A, removes A's record and also wins. Containing every read,
//     judgement and removal inside the lock removes the window by construction.
//   - The empty-record window. atomicfile.Claim writes no content, so claiming
//     the record would leave it EXISTING BUT EMPTY between the create and the
//     content write. The lock file is supposed to be empty; the record never is.
//
// This is the composition internal/verify/claim_lock.go already uses (it claims
// SnapshotPath(...)+".lock" rather than the snapshot).
const driftFillLockSuffix = ".lock"

// driftFillRecord is the suppression record's payload, written whole, under the
// lock.
type driftFillRecord struct {
	HeadSHA   string    `json:"head_sha"`
	StartedAt time.Time `json:"started_at"`
}

func driftFillRecordPath(projectDir string) string {
	return filepath.Join(projectDir, ".moai", "state", driftFillRecordFilename)
}

func driftFillLockPath(projectDir string) string {
	return driftFillRecordPath(projectDir) + driftFillLockSuffix
}

// Fill seams. All are nil/production-wired by default; tests replace them.
//
// driftFillSpawnProbe's PLACEMENT is load-bearing, and for the same reason
// landedSpawnProbe's is: under `go test` os.Executable() is the test binary, so
// the basename guard always blocks the real exec. A counter at function entry
// could not tell "gated before spawning" from "spawn blocked by the guard". The
// probe therefore sits AFTER the suppression gates and BEFORE the
// self-invocation guard.
var (
	driftFillSpawnProbe   func(projectDir string)
	driftFillExecutableFn = os.Executable
	driftFillHeadFn       = spec.CacheHeadSHA
	driftFillNowFn        = time.Now

	// driftFillStartFn is the SINGLE point at which a fill child is actually
	// started. It is the observation method for "no process was started": that
	// claim is established by this counter reading zero, never by asserting
	// against the absence of a real process. Production is (*exec.Cmd).Start.
	driftFillStartFn = func(cmd *exec.Cmd) error { return cmd.Start() }

	// driftFillPostReadProbe fires inside the critical section, between the
	// record read and the removal that follows it, so a test can inject a
	// record replacement into exactly the window the reclaim TOCTOU lived in
	// and observe that the handler re-judges at removal time rather than acting
	// on the earlier read.
	driftFillPostReadProbe func(recordPath string)
)

// maybeFillDriftCache starts at most one detached fill child for this project,
// and returns immediately either way.
//
// Every failure path is FAIL-OPEN and silent at debug level: this runs on the
// session-start advisory path, where a fill problem must never fail — or even
// degrade — the hook's output.
//
// @MX:WARN: [AUTO] fire-and-forget detached child — never awaited, exit status never read.
// @MX:REASON: SPEC-DRIFT-CACHE-FILL-001 REQ-DCF-001/003 — the whole point is a
//
//	writer that outlives this process. The child bounds ITSELF via the deadline
//	carried on its invocation; there is no supervisor and no trailing kill,
//	because a trailing kill is a line the parent may never reach.
func maybeFillDriftCache(projectDir string) {
	if projectDir == "" {
		return
	}
	if !config.DriftCacheFillEnabledForRoot(projectDir) {
		return
	}

	head, err := driftFillHeadFn(projectDir)
	if err != nil || head == "" {
		slog.Debug("session start (deferred): drift fill skipped, HEAD unresolved",
			"project_dir", projectDir)
		return
	}

	if !claimDriftFill(projectDir, head) {
		return
	}

	if driftFillSpawnProbe != nil {
		driftFillSpawnProbe(projectDir)
	}

	self, err := driftFillExecutableFn()
	if err != nil || !isMoaiExecutable(self) {
		slog.Debug("session start (deferred): drift fill not started, executable is not a moai binary",
			"executable", self)
		return
	}

	cmd := buildDriftFillCommand(self, projectDir)
	if err := driftFillStartFn(cmd); err != nil {
		slog.Debug("session start (deferred): drift fill failed to start",
			"error", err.Error())
		return
	}
	if cmd.Process != nil {
		_ = cmd.Process.Release() // fire-and-forget; never Wait()
	}
}

// buildDriftFillCommand constructs the fill child. It is separated from the
// spawn so the DETACH CONTRACT can be asserted on the constructed command
// rather than on a live exec.
//
// Detached means the child inherits none of the parent's streams. A child
// holding the hook's stdout keeps that pipe open, which makes whatever is
// reading it wait for a process it never knew about.
func buildDriftFillCommand(self, projectDir string) *exec.Cmd {
	cmd := exec.Command(self, "spec", "drift",
		"--fill-cache",
		"--fill-timeout", config.DefaultDriftCacheFillTimeout.String(),
	)
	cmd.Dir = projectDir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	return cmd
}

// isMoaiExecutable is the self-invocation guard, matching
// internal/statusline's isSelfInvocable: only a moai binary is re-invoked, so a
// test binary (or anything else os.Executable() happens to resolve to) never
// becomes a child.
func isMoaiExecutable(path string) bool {
	base := filepath.Base(path)
	return base == "moai" || base == "moai.exe"
}

// claimDriftFill decides whether THIS handler may start a fill, and records the
// decision so no other handler starts one for the same HEAD within the TTL.
//
// The whole read-judge-remove-write sequence happens INSIDE the companion lock.
// A plain write is not enough and a read-then-write is not enough: N handlers
// released simultaneously all complete their READS before the first WRITE
// lands, so all of them spawn. The ordering that matters is not "first write
// before any child" but "first write before other reads", and only an exclusive
// create provides it.
//
// An in-process mutex is explicitly NOT the fix: eight goroutines are not eight
// processes.
func claimDriftFill(projectDir, head string) bool {
	recordPath := driftFillRecordPath(projectDir)
	if err := os.MkdirAll(filepath.Dir(recordPath), 0o755); err != nil {
		slog.Debug("session start (deferred): drift fill state dir unavailable",
			"error", err.Error())
		return false
	}

	lockPath := driftFillLockPath(projectDir)
	if !acquireDriftFillClaimLock(lockPath) {
		// Another handler is inside the critical section. This one starts
		// nothing — no queueing, no retry.
		return false
	}
	defer func() { _ = os.Remove(lockPath) }()

	now := driftFillNowFn()
	ttl := config.DefaultDriftCacheFillTTL

	// Read under the lock. An unreadable record — absent, empty, truncated or
	// unparseable — is ONE stated disposition: expired. Two divergent readings
	// would diverge into a double spawn or a contradiction.
	if rec, ok := readDriftFillRecord(recordPath); ok && driftFillRecordSuppresses(rec, head, now, ttl) {
		return false
	}

	// The expiry judgement is made at the MOMENT OF REMOVAL, never carried over
	// from the read above. The probe marks that window explicitly so a test can
	// inject into it.
	if driftFillPostReadProbe != nil {
		driftFillPostReadProbe(recordPath)
	}
	if cur, ok := readDriftFillRecord(recordPath); ok && driftFillRecordSuppresses(cur, head, now, ttl) {
		// Something fresher landed; do not destroy it.
		return false
	}

	_ = os.Remove(recordPath)
	// The spawn is authorised ONLY by a record this handler actually wrote: a
	// failed write means the next session would see no suppression, so starting
	// a child here would be the respawn loop REQ-DCF-010 exists to prevent.
	return writeDriftFillRecord(recordPath, driftFillRecord{HeadSHA: head, StartedAt: now})
}

// driftFillRecordSuppresses reports whether an existing record bars a new fill.
// All three tests must hold; failing any one makes the record reclaimable.
//
// The future-timestamp test is not defensive noise: a clock stepped backwards,
// or a clone carrying a foreign stamp, produces a record that reads as "younger
// than the TTL" forever and would suppress every fill until HEAD moves.
func driftFillRecordSuppresses(rec driftFillRecord, head string, now time.Time, ttl time.Duration) bool {
	if rec.HeadSHA == "" || rec.HeadSHA != head {
		return false
	}
	if rec.StartedAt.After(now) {
		return false
	}
	return now.Sub(rec.StartedAt) <= ttl
}

// readDriftFillRecord returns the record when it is present AND parseable.
// ok=false is the single "expired" disposition for every unreadable shape.
func readDriftFillRecord(path string) (driftFillRecord, bool) {
	raw, err := atomicfile.ReadFile(path)
	if err != nil {
		return driftFillRecord{}, false
	}
	var rec driftFillRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return driftFillRecord{}, false
	}
	if rec.HeadSHA == "" || rec.StartedAt.IsZero() {
		return driftFillRecord{}, false
	}
	return rec, true
}

// writeDriftFillRecord writes the record WHOLE: a temp file, then an atomic
// replace. The record is never observed half-written, which is what lets an
// unreadable record have one unambiguous disposition.
func writeDriftFillRecord(path string, rec driftFillRecord) bool {
	data, err := json.Marshal(rec)
	if err != nil {
		return false
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".drift-fill-*.tmp")
	if err != nil {
		return false
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return false
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return false
	}
	if err := atomicfile.Replace(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return false
	}
	return true
}

// acquireDriftFillClaimLock takes the companion lock by EXCLUSIVE CREATION —
// an operation at most one caller on the machine can win, which fails rather
// than overwrites when the path already exists, and which is NEVER retried
// (a retry would hand the same claim to a second caller and destroy the
// exclusion it exists to provide).
//
// RESIDUAL, stated rather than presented as closed: a lock older than its own
// staleness bound is reclaimable, and that reclaim is a stat followed by a
// remove. Judging staleness at the moment of removal NARROWS the window to the
// interval between the stat and the remove; it does not eliminate it, because
// no filesystem offers an atomic test-and-remove. A microsecond-wide race
// therefore survives, and its worst case is exactly one extra child — the same
// fail-open cost this path already accepts everywhere else, never data
// corruption. This is the identical residual the in-tree precedent accepts with
// its own reasoning written down (internal/verify/claim_lock.go, the
// "Fail-open policy" comment on acquireKeyLock).
func acquireDriftFillClaimLock(lockPath string) bool {
	if err := atomicfile.Claim(lockPath, 0o644); err == nil {
		return true
	}
	if !driftFillLockStale(lockPath) {
		return false
	}
	_ = os.Remove(lockPath)
	// One re-attempt, never a loop: a handler that loses the re-attempt starts
	// no fill, which costs at most one deferred advisory.
	return atomicfile.Claim(lockPath, 0o644) == nil
}

// driftFillLockStale reports whether a held lock is old enough to have been
// abandoned by a handler that died inside the critical section. A non-existent
// lock is NOT stale — the caller will exclusively create it.
func driftFillLockStale(lockPath string) bool {
	info, err := os.Stat(lockPath)
	if err != nil {
		return false
	}
	return driftFillNowFn().Sub(info.ModTime()) > config.DriftCacheFillLockStaleness
}
