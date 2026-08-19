// Package session provides session lifecycle primitives for MoAI-ADK.
//
// This file implements the multi-session coordination registry per
// SPEC-V3R6-MULTI-SESSION-COORD-001. It coexists with the existing
// file-first state checkpoint primitives (state.go, store.go, lock.go)
// from SPEC-V3R2-RT-004 — both subsystems share the same package but
// operate on distinct files under .moai/state/.
//
// The registry tracks active Claude Code sessions on a single host. It is
// advisory (not a strong mutual exclusion lock): the orchestrator queries
// the registry pre-spawn and decides whether to proceed, defer, or escalate
// to the user via AskUserQuestion. The registry file is per-machine,
// gitignored, and lives at .moai/state/active-sessions.json.
//
// @MX:ANCHOR: [AUTO] SPEC-V3R6-MULTI-SESSION-COORD-001 L1 registry primitive
// @MX:REASON: fan_in >= 5 (CLI register/heartbeat/deregister/list/purge + hook session_start 3-step protocol). Any schema or atomic-write change here ripples through L2/L3/L4.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
)

// DefaultRegistryPath is the canonical project-relative path for the
// multi-session coordination registry. It is gitignored via the .moai/state/
// blanket rule. Tests pass an explicit path via NewRegistry; package-level
// helpers default to this constant.
const DefaultRegistryPath = ".moai/state/active-sessions.json"

// CurrentSideChannelFile is the project-relative path of the side-channel
// file the SessionStart hook writes (SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001
// M3 additionalContext injection) so `moai session current` can read THIS
// orchestrator's own UUID back. The file is per-project (lives under the
// gitignored .moai/state/ tree) and is overwritten on every SessionStart.
//
// This constant lives in internal/session (not internal/cli) to avoid an
// import cycle: internal/hook needs it for the write path, internal/cli
// needs it for the read path, and internal/cli already imports internal/hook.
//
// SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 REQ-RDP-002, REQ-RDP-004.
const CurrentSideChannelFile = ".moai/state/current-session-id.txt"

// PhaseNone is the sentinel string for entries with no SPEC scope yet
// (e.g., a session registered at SessionStart hook before any /moai
// subcommand has been invoked).
const PhaseNone = "(none)"

// SpecIDNone is the sentinel string for entries with no SPEC scope.
const SpecIDNone = "(none)"

// DefaultStaleMinutes is the default heartbeat threshold for PurgeStale.
// Sessions whose last_heartbeat is older than this are considered zombie
// and removed on the next PurgeStale call.
const DefaultStaleMinutes = 30

// LockTimeout is the default maximum wait for advisory lock acquisition.
// Beyond this, registry operations return ErrLockTimeout. This avoids
// indefinite block in CLI/hook contexts (AP-MSC-005). Tests with high
// contention may override via Registry.lockTimeout.
const LockTimeout = 2 * time.Second

// ErrLockTimeout is returned when registry lock acquisition exceeds LockTimeout.
var ErrLockTimeout = errors.New("session registry: lock acquisition timed out")

// ErrEntryNotFound is returned when a sessionID is not present in the registry.
// Note: Heartbeat and DeregisterSession are idempotent — they MUST NOT return
// this error. It is exported for callers that wish to query existence
// explicitly via internal helpers.
var ErrEntryNotFound = errors.New("session registry: entry not found")

// Entry is a single row in the active-sessions registry.
//
// Schema is frozen per REQ-COORD-002 and REQ-COORD-024. Any modification
// requires a follow-up SPEC superseding REQ-COORD-024.
type Entry struct {
	SessionID     string    `json:"session_id"`
	SpecID        string    `json:"spec_id"`
	Phase         string    `json:"phase"`
	StartedAt     time.Time `json:"started_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	PID           int       `json:"pid"`
	Host          string    `json:"host"`
	CWD           string    `json:"cwd"`
}

// Clock is the time abstraction used by registry helpers so tests can
// substitute a FakeClock for deterministic time control (AP-MSC-004).
type Clock interface {
	Now() time.Time
}

// realClock returns time.Now() in UTC.
type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

// FakeClock is a deterministic clock for tests. The current value is
// returned by Now() until callers reassign Current.
type FakeClock struct {
	Current time.Time
}

// Now returns the current FakeClock value.
func (f *FakeClock) Now() time.Time { return f.Current }

// Registry is the in-package handle used to interact with a specific
// registry file. Public package functions (RegisterSession, Heartbeat,
// etc.) delegate to a Registry instance bound to DefaultRegistryPath.
// Tests construct their own Registry via NewRegistry with t.TempDir() and
// a FakeClock.
type Registry struct {
	path        string
	clock       Clock
	lockTimeout time.Duration
}

// NewRegistry constructs a Registry bound to the given file path. The path
// may be relative (resolved against CWD) or absolute. Clock controls
// timestamp generation; pass realClock{} or nil for default UTC time.
// LockTimeout defaults to LockTimeout constant.
func NewRegistry(path string, clock Clock) *Registry {
	if clock == nil {
		clock = realClock{}
	}
	return &Registry{path: path, clock: clock, lockTimeout: LockTimeout}
}

// WithLockTimeout returns a copy of r with the lock acquisition timeout
// overridden. Used by tests to tolerate high concurrent contention. The
// production default (2s) is appropriate for CLI/hook contexts where
// indefinite block would freeze user interactions.
func (r *Registry) WithLockTimeout(d time.Duration) *Registry {
	clone := *r
	clone.lockTimeout = d
	return &clone
}

// defaultRegistry returns a Registry bound to DefaultRegistryPath with the
// real clock. Used by package-level RegisterSession/Heartbeat/etc. helpers.
func defaultRegistry() *Registry {
	return NewRegistry(DefaultRegistryPath, realClock{})
}

// RegisterSession atomically appends a new entry with started_at and
// last_heartbeat set to the current time. If an entry with the same
// sessionID already exists, it is updated in place (idempotent on
// session_id collision per §F.5 mitigation).
//
// REQ-COORD-001, REQ-COORD-003, REQ-COORD-008.
func RegisterSession(sessionID, specID, phase string) error {
	return defaultRegistry().Register(sessionID, specID, phase)
}

// Register is the method form of RegisterSession bound to a specific Registry.
func (r *Registry) Register(sessionID, specID, phase string) error {
	if sessionID == "" {
		return errors.New("session registry: sessionID cannot be empty")
	}
	host, _ := os.Hostname()
	cwd, _ := os.Getwd()
	now := r.clock.Now().UTC()
	return r.withLock(func(entries []Entry) ([]Entry, error) {
		// Idempotent: update in place if sessionID exists; else append.
		for i := range entries {
			if entries[i].SessionID == sessionID {
				entries[i].SpecID = specID
				entries[i].Phase = phase
				entries[i].LastHeartbeat = now
				// Preserve original StartedAt + PID + Host (PID may differ
				// across reconnects but we treat first-seen as canonical).
				return entries, nil
			}
		}
		entries = append(entries, Entry{
			SessionID:     sessionID,
			SpecID:        specID,
			Phase:         phase,
			StartedAt:     now,
			LastHeartbeat: now,
			// The session PID, NOT os.Getpid(): Register runs inside a hook
			// subprocess that exits immediately, so its own PID would be dead
			// before any liveness probe reads the registry back. See
			// session_pid.go.
			PID:           resolveSessionPID(),
			Host:          host,
			CWD:           cwd,
		})
		return entries, nil
	})
}

// Heartbeat atomically updates the LastHeartbeat field of the matching entry.
// Idempotent on missing entry: returns nil error if no entry matches.
//
// REQ-COORD-004.
func Heartbeat(sessionID string) error {
	return defaultRegistry().Heartbeat(sessionID)
}

// Heartbeat is the method form.
func (r *Registry) Heartbeat(sessionID string) error {
	if sessionID == "" {
		return errors.New("session registry: sessionID cannot be empty")
	}
	now := r.clock.Now().UTC()
	return r.withLock(func(entries []Entry) ([]Entry, error) {
		for i := range entries {
			if entries[i].SessionID == sessionID {
				entries[i].LastHeartbeat = now
				return entries, nil
			}
		}
		// Idempotent on missing — REQ-COORD-004 only mutates when found.
		return entries, nil
	})
}

// DeregisterSession atomically removes the matching entry. Idempotent on
// missing entry: returns nil error if no entry matches.
//
// REQ-COORD-005.
func DeregisterSession(sessionID string) error {
	return defaultRegistry().Deregister(sessionID)
}

// Deregister is the method form.
func (r *Registry) Deregister(sessionID string) error {
	if sessionID == "" {
		return errors.New("session registry: sessionID cannot be empty")
	}
	return r.withLock(func(entries []Entry) ([]Entry, error) {
		filtered := entries[:0]
		for _, e := range entries {
			if e.SessionID != sessionID {
				filtered = append(filtered, e)
			}
		}
		return filtered, nil
	})
}

// QueryActiveWork returns a snapshot of registry entries. If optSpecID is
// non-empty, only entries with matching SpecID are returned. The returned
// slice is a copy; mutating it does not affect the registry.
//
// REQ-COORD-006.
func QueryActiveWork(optSpecID string) ([]Entry, error) {
	return defaultRegistry().Query(optSpecID)
}

// Query is the method form.
func (r *Registry) Query(optSpecID string) ([]Entry, error) {
	entries, err := r.readAll()
	if err != nil {
		return nil, err
	}
	if optSpecID == "" {
		return entries, nil
	}
	filtered := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if e.SpecID == optSpecID {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

// PurgeStale removes entries whose LastHeartbeat is older than
// thresholdMinutes. Returns the count of removed entries.
//
// REQ-COORD-007.
func PurgeStale(thresholdMinutes int) (int, error) {
	return defaultRegistry().Purge(thresholdMinutes)
}

// Purge is the method form.
func (r *Registry) Purge(thresholdMinutes int) (int, error) {
	if thresholdMinutes <= 0 {
		thresholdMinutes = DefaultStaleMinutes
	}
	cutoff := r.clock.Now().UTC().Add(-time.Duration(thresholdMinutes) * time.Minute)
	purged := 0
	err := r.withLock(func(entries []Entry) ([]Entry, error) {
		filtered := entries[:0]
		for _, e := range entries {
			if e.LastHeartbeat.Before(cutoff) {
				purged++
				continue
			}
			filtered = append(filtered, e)
		}
		return filtered, nil
	})
	if err != nil {
		return 0, err
	}
	return purged, nil
}

// Retry backoff for the non-blocking registry lock loop
// (SPEC-CI-FLAKY-STABILIZE-001, REQ-CFS-010 / REQ-CFS-011).
//
// lockBackoffCap is deliberately far below the LockTimeout default (2s) so a
// single sleep can never consume the whole acquisition budget.
const (
	lockBackoffBase     = 5 * time.Millisecond
	lockBackoffCap      = 50 * time.Millisecond
	lockBackoffMaxShift = 8 // ceiling saturates well before this; guards the shift
)

// lockRetryDelay returns how long to sleep before retry number attempt
// (0-based) of the acquisition loop: full jitter over an exponentially growing
// ceiling, clamped to the time remaining before the deadline.
//
// Why jitter rather than a fixed interval: acquisition is LOCK_EX|LOCK_NB, so
// contenders never queue in the kernel and there is no fairness guarantee. With
// a fixed sleep every contender wakes on the same cadence, and an unlucky
// goroutine can lose the race repeatedly. Randomizing wake times breaks that
// synchronization. This is a probabilistic mitigation, NOT structural fairness.
//
// SPEC-V3R6-CI-FLAKY-STABILIZE-003 added the structural fix: the path-keyed
// in-process mutex (acquired at the top of withLock) now serializes same-process
// contenders deterministically, eliminating the same-process starvation scenario
// this jitter could only probabilistically mitigate. Jitter remains the
// cross-process retry strategy — distinct moai processes still contend on the
// NB flock with no kernel fairness, so under cross-process contention a repeated
// loser remains possible (the registry's original coordination purpose,
// REQ-COORD-008; production cross-process contention is a handful of sessions,
// so the probability is low).
//
// The clamp keeps one sleep from overshooting the deadline and costing the
// caller its final retry.
func lockRetryDelay(attempt int, remaining time.Duration) time.Duration {
	if remaining <= 0 {
		return 0
	}
	shift := attempt
	if shift > lockBackoffMaxShift {
		shift = lockBackoffMaxShift
	}
	ceiling := lockBackoffBase << shift
	if ceiling > lockBackoffCap {
		ceiling = lockBackoffCap
	}
	delay := time.Duration(rand.Int64N(int64(ceiling)) + 1) // full jitter over (0, ceiling]
	if delay > remaining {
		delay = remaining
	}
	return delay
}

// envFalsifiabilityDisableMutex is the env var that disables the in-process
// mutex for the AC-CFS3-002 falsifiability sub-case (local verification only;
// CI never sets it). When set to "1", withLock skips the in-process mutex so
// the structural probe can observe concurrent residency >= 2, proving the AC
// is not vacuous. SPEC-V3R6-CI-FLAKY-STABILIZE-003.
const envFalsifiabilityDisableMutex = "MOAI_CFS3_FALSIFIABILITY"

// inProcessMutexDisabled reports whether the in-process mutex is deliberately
// disabled for the AC-CFS3-002 falsifiability sub-case.
func inProcessMutexDisabled() bool {
	return os.Getenv(envFalsifiabilityDisableMutex) == "1"
}

// inProcessMutexes is the package-level path-keyed map of in-process mutexes
// (SPEC-V3R6-CI-FLAKY-STABILIZE-003, REQ-CFS3-001/002). Each distinct absolute
// lock path gets its own *sync.Mutex, so ALL Registry instances sharing a path
// serialize against each other — including the per-call defaultRegistry()
// instances the package helpers (RegisterSession, Heartbeat, ...) create, which
// a per-Registry field would miss. The map is path-keyed (not per-Registry) by
// design decision OQ-1 (ADOPT A): defaultRegistry() returns a fresh Registry per
// call, so only a path-keyed map covers every in-process contention shape.
//
// Lock ordering: in-process mutex -> registryLock.mu (fd guard) -> flock. No
// cycle (the in-process mutex is always released before withLock returns, and
// flock is released within the same withLock call).
//
// The map grows monotonically with distinct paths over the process lifetime;
// for the registry this is effectively bounded (one path per project). Tests
// use t.TempDir() paths which are process-lifetime, so the growth is bounded
// by the test run.
//
// @MX:NOTE: [AUTO] path-keyed in-process mutex; double-locking with cross-process flock (in-process serialization + cross-process coordination)
var inProcessMutexes sync.Map // map[string]*sync.Mutex

// acquireInProcessMutex returns the *sync.Mutex for the given absolute lock
// path, allocating it on first use via LoadOrStore. Callers MUST pair the
// returned Lock with a defer Unlock.
func acquireInProcessMutex(absLockPath string) *sync.Mutex {
	v, _ := inProcessMutexes.LoadOrStore(absLockPath, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// withLockProbe is the structural probe for AC-CFS3-002
// (SPEC-V3R6-CI-FLAKY-STABILIZE-003). It counts how many goroutines are
// simultaneously resident inside a withLock critical section. The in-process
// mutex (added by this SPEC) must keep the observed maximum at <= 1; the
// falsifiability sub-case (MOAI_CFS3_FALSIFIABILITY=1, which disables the
// mutex) must observe >= 2 under sufficient contention, proving the AC is not
// vacuous. Pure atomic operations; safe under -race.
//
// The inc/dec bracket the ENTIRE withLock body (including the NB-flock retry
// loop), so the probe measures "goroutines inside withLock" rather than "flock
// holders" — the latter is <= 1 regardless of the mutex (the OS serializes
// flock) and would make the AC unfalsifiable (verification-claim-integrity
// §1.1 surface 3).
type withLockProbeType struct {
	inFlight    atomic.Int64
	maxInFlight atomic.Int64
}

var withLockProbe = &withLockProbeType{}

// enter increments the in-flight counter and refreshes the observed maximum.
func (p *withLockProbeType) enter() {
	n := p.inFlight.Add(1)
	for {
		m := p.maxInFlight.Load()
		if n <= m || p.maxInFlight.CompareAndSwap(m, n) {
			return
		}
	}
}

// exit decrements the in-flight counter.
func (p *withLockProbeType) exit() {
	p.inFlight.Add(-1)
}

// max returns the highest concurrent in-flight count observed since reset.
func (p *withLockProbeType) max() int64 {
	return p.maxInFlight.Load()
}

// reset clears the probe counters. Test-only; used between sub-cases.
func (p *withLockProbeType) reset() {
	p.inFlight.Store(0)
	p.maxInFlight.Store(0)
}

// withLock acquires the advisory lock, reads the registry, applies the
// mutation function, writes back atomically (temp + rename), and releases
// the lock. The lock is best-effort: on contention, it retries up to
// LockTimeout. On timeout, returns ErrLockTimeout.
//
// All mutation paths (Register / Heartbeat / Deregister / Purge) go through
// this helper. Reads (Query) bypass the lock — eventually consistent reads
// are acceptable per AP-MSC-002.
//
// @MX:NOTE: [AUTO] Every registry mutation flows through withLock. Reads
// bypass the lock per the design decision in AP-MSC-002. New mutation paths
// must use this helper exactly.
func (r *Registry) withLock(mutate func([]Entry) ([]Entry, error)) error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return fmt.Errorf("session registry: mkdir parent: %w", err)
	}

	// In-process mutex (SPEC-V3R6-CI-FLAKY-STABILIZE-003, REQ-CFS3-001).
	// NB-flock provides no kernel fairness queue, so without this mutex an
	// unlucky same-process goroutine can lose the flock race repeatedly and
	// exceed lockTimeout (starvation -> ErrLockTimeout). The mutex is
	// path-keyed so every Registry instance sharing a lock path serializes.
	// It brackets the ENTIRE body (including the NB-flock retry loop below),
	// which is where the starvation occurs. Cross-process contention still
	// relies on flock + jittered backoff (REQ-CFS3-005/007/008); the mutex
	// only removes same-process unfairness.
	//
	// Placed at the withLock level (NOT inside registryLock.acquire) so the
	// TestWithLockTimeoutContract blocker — which calls acquire() directly on
	// a separate registryLock — does NOT take this mutex and the
	// ErrLockTimeout contract is preserved (AP-CFS3-004).
	if !inProcessMutexDisabled() {
		absLockPath, _ := filepath.Abs(r.path + ".lock")
		ipm := acquireInProcessMutex(absLockPath)
		ipm.Lock()
		defer ipm.Unlock()
	}

	// Structural probe (AC-CFS3-002): brackets the entire body so the
	// in-process mutex's serialization effect is observable. See withLockProbe
	// doc. enter()/exit() are pure atomic ops; negligible hot-path cost.
	withLockProbe.enter()
	defer withLockProbe.exit()

	lockPath := r.path + ".lock"
	lock := newRegistryLock()
	timeout := r.lockTimeout
	if timeout <= 0 {
		timeout = LockTimeout
	}
	deadline := time.Now().Add(timeout)
	for attempt := 0; ; attempt++ {
		err := lock.acquire(lockPath)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: %v", ErrLockTimeout, err)
		}
		time.Sleep(lockRetryDelay(attempt, time.Until(deadline)))
	}
	defer func() { _ = lock.release() }()

	entries, err := r.readAllUnlocked()
	if err != nil {
		return err
	}

	newEntries, err := mutate(entries)
	if err != nil {
		return err
	}

	// Sort by StartedAt for deterministic on-disk output (helps git diff
	// in the unlikely case the file is committed, and aids golden-file
	// snapshot testing). Per §F.3.
	sort.Slice(newEntries, func(i, j int) bool {
		if newEntries[i].StartedAt.Equal(newEntries[j].StartedAt) {
			return newEntries[i].SessionID < newEntries[j].SessionID
		}
		return newEntries[i].StartedAt.Before(newEntries[j].StartedAt)
	})

	return r.writeAtomic(newEntries)
}

// readAll returns a copy of the registry contents without holding the lock.
// Used by Query (REQ-COORD-006 explicitly permits eventually consistent
// reads).
func (r *Registry) readAll() ([]Entry, error) {
	return r.readAllUnlocked()
}

// readAllUnlocked is the inner read used both by readAll and by withLock.
// It returns an empty slice (not error) when the registry file is absent.
func (r *Registry) readAllUnlocked() ([]Entry, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("session registry: read %s: %w", r.path, err)
	}
	if len(data) == 0 {
		return []Entry{}, nil
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("session registry: parse %s: %w", r.path, err)
	}
	return entries, nil
}

// writeAtomic writes the entries to the registry via temp + rename.
// On POSIX, os.Rename is atomic within the same filesystem. On Windows,
// os.Rename uses MoveFileEx with MOVEFILE_REPLACE_EXISTING semantics.
//
// REQ-COORD-008, REQ-COORD-022.
func (r *Registry) writeAtomic(entries []Entry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("session registry: marshal: %w", err)
	}
	// Ensure trailing newline for POSIX cleanliness.
	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	dir := filepath.Dir(r.path)
	tmp, err := os.CreateTemp(dir, ".active-sessions-*.json.tmp")
	if err != nil {
		return fmt.Errorf("session registry: create temp: %w", err)
	}
	tmpPath := tmp.Name()

	// Best-effort cleanup on error.
	cleanup := func() {
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("session registry: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("session registry: close temp: %w", err)
	}

	if err := atomicfile.Replace(tmpPath, r.path); err != nil {
		cleanup()
		return fmt.Errorf("session registry: rename temp -> %s: %w", r.path, err)
	}
	return nil
}

// FormatStderrReminder formats the QueryActiveWork result as a stderr
// system-reminder block. Used by SessionStart hook Step 3.
//
// REQ-COORD-015.
func FormatStderrReminder(currentSessionID string, entries []Entry, now time.Time) string {
	others := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if e.SessionID != currentSessionID {
			others = append(others, e)
		}
	}
	if len(others) == 0 {
		return ""
	}
	out := "<system-reminder>\n"
	out += fmt.Sprintf("Multi-Session Coordination: %d other active session(s) on this host\n", len(others))
	out += fmt.Sprintf("(this session %s):\n", shortID(currentSessionID))
	for _, e := range others {
		age := now.Sub(e.LastHeartbeat)
		ageMin := int(age.Minutes())
		if ageMin < 0 {
			ageMin = 0
		}
		out += fmt.Sprintf("  - %s spec=%s phase=%s age=%dm pid=%d cwd=%s\n",
			shortID(e.SessionID), e.SpecID, e.Phase, ageMin, e.PID, e.CWD)
	}
	out += "</system-reminder>\n"
	return out
}

// shortID returns the first 8 characters of a UUID for compact display.
func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// detectHost reports the current hostname, falling back to "unknown" if
// os.Hostname fails. Exported for test introspection only.
func detectHost() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		return "unknown"
	}
	return host
}

// platformTag returns the GOOS/GOARCH string for diagnostic logging.
// Currently unused by the registry itself but kept for future
// cross-platform telemetry hooks. Compile-tested across build matrix.
func platformTag() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}
