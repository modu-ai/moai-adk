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
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
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
			PID:           os.Getpid(),
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

	lockPath := r.path + ".lock"
	lock := newRegistryLock()
	timeout := r.lockTimeout
	if timeout <= 0 {
		timeout = LockTimeout
	}
	deadline := time.Now().Add(timeout)
	for {
		err := lock.acquire(lockPath)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: %v", ErrLockTimeout, err)
		}
		time.Sleep(20 * time.Millisecond)
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

	if err := os.Rename(tmpPath, r.path); err != nil {
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
