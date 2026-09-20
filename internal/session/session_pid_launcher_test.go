package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestRegister_RecordsLivePID_LauncherPath is the launcher-path counterpart to
// TestRegister_RecordsLivePID: `moai cc` / `moai glm` / `moai cg` stamp
// MOAI_SESSION_PID into the environment they execve(2) into, so every hook
// subprocess the session later spawns inherits the session PID outright. The
// registry must record that PID rather than the walk's own answer — the walk
// is a fallback on this path, not a step.
//
// The synthetic ancestry below deliberately resolves to a DIFFERENT PID, so a
// resolver that took the walk's answer would fail this assertion rather than
// coincidentally agree with it.
//
// Card t958 moved the stamp INTO the chain (self -> 7799 -> 7710) without
// changing what the row asserts. The launcher execve(2)s into Claude Code, so
// the pid it stamped IS the session's, and the session IS an ancestor of every
// hook subprocess that later reads the stamp — the previous fixture, which
// placed the stamp nowhere in the chain, described a shape the launcher path
// does not produce. What it DOES describe is the t958 defect: a stamp
// inherited from an unrelated session. The walk still answers 7799 here, so
// "the stamp wins over the walk" is measured exactly as before.
func TestRegister_RecordsLivePID_LauncherPath(t *testing.T) {
	const (
		launcherStampedPID = 7710 // what the launcher exported: the live session
		ancestryPID        = 7799 // what a walk would have found instead
	)
	withFakeAncestry(t,
		map[int]fakeProcess{
			os.Getpid():        {ppid: ancestryPID, comm: "moai"},
			ancestryPID:        {ppid: launcherStampedPID, comm: "claude"},
			launcherStampedPID: {ppid: 1, comm: "login"},
		},
		map[int]bool{os.Getpid(): true, ancestryPID: true, launcherStampedPID: true},
	)
	t.Setenv(config.EnvMoaiSessionPID, strconv.Itoa(launcherStampedPID))

	path := filepath.Join(t.TempDir(), "active-sessions.json")
	reg := NewRegistry(path, nil)
	if err := reg.Register("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", SpecIDNone, PhaseNone); err != nil {
		t.Fatalf("Register: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	var entries []Entry
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatalf("unmarshal registry: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("registry holds %d entries, want 1", len(entries))
	}
	if entries[0].PID != launcherStampedPID {
		t.Errorf("recorded PID = %d, want the launcher-stamped session PID %d (the ancestry walk would have yielded %d)",
			entries[0].PID, launcherStampedPID, ancestryPID)
	}
	if !pidIsAlive(entries[0].PID) {
		t.Errorf("recorded PID %d is not alive immediately after registration", entries[0].PID)
	}
}

// TestResolveSessionPID_IgnoresDeadLauncherStamp guards the inherited-stamp
// hazard from the other direction: an environment carried over from a launcher
// whose session has already exited must NOT be recorded. The resolver falls
// through to the ancestry walk instead.
func TestResolveSessionPID_IgnoresDeadLauncherStamp(t *testing.T) {
	const (
		deadStamp  = 7810
		sessionPID = 7820
	)
	withFakeAncestry(t,
		map[int]fakeProcess{
			os.Getpid(): {ppid: sessionPID, comm: "moai"},
			sessionPID:  {ppid: 1, comm: "claude"},
		},
		map[int]bool{os.Getpid(): true, sessionPID: true}, // deadStamp deliberately absent
	)
	t.Setenv(config.EnvMoaiSessionPID, strconv.Itoa(deadStamp))

	if got := resolveSessionPID(); got != sessionPID {
		t.Errorf("resolveSessionPID = %d, want the ancestry PID %d (a dead stamp must not win)", got, sessionPID)
	}
}
