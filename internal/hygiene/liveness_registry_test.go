package hygiene

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestLivenessRealRegistry — the production (seam-nil) evaluator path over
// a fixture registry file: entry present with a reaped pid and an old
// heartbeat reads DEAD-eligible on the pid+heartbeat half; an entry miss
// reads unmeasured; a future heartbeat reads affirmative (clock skew errs
// toward keeping).
func TestLivenessRealRegistry(t *testing.T) {
	cmd := exec.Command("true")
	if err := cmd.Start(); err != nil {
		t.Fatalf("spawn probe child: %v", err)
	}
	deadPID := cmd.Process.Pid
	_, _ = cmd.Process.Wait()

	registryPath := filepath.Join(t.TempDir(), "active-sessions.json")
	write := func(entries []RegistryEntry) {
		t.Helper()
		blob, err := json.Marshal(entries)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if err := os.WriteFile(registryPath, blob, 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	old := fixtureNow.Add(-testStaleHb).Add(-time.Hour)
	write([]RegistryEntry{{SessionID: keyDead, PID: deadPID, LastHeartbeat: old}})
	l := &Liveness{
		TranscriptWindow: testActivityWindow,
		HeartbeatWindow:  testStaleHb,
		RegistryPath:     registryPath,
		now:              func() time.Time { return fixtureNow },
	}

	// Entry present: pid negative (reaped), heartbeat negative, transcript
	// unmeasured (no roots) ⇒ INDETERMINATE — and the unmeasured names
	// carry exactly the transcript signal.
	verdict, unmeasured, evidence := l.Evaluate(keyDead)
	if verdict != VerdictIndeterminate {
		t.Fatalf("verdict = %s, want indeterminate (transcript unmeasured)", verdict)
	}
	if len(unmeasured) != 1 || unmeasured[0] != "transcript" {
		t.Fatalf("unmeasured = %v, want [transcript]", unmeasured)
	}
	if evidence["pid"] != "negative" || evidence["heartbeat"] != "negative" {
		t.Fatalf("evidence = %v", evidence)
	}

	// Entry missing for the key: pid+heartbeat unmeasured (transcript
	// unmeasured too — no roots), never DEAD.
	write([]RegistryEntry{{SessionID: keyGoal, PID: deadPID, LastHeartbeat: old}})
	verdict, unmeasured, _ = l.Evaluate(keyDead)
	if verdict != VerdictIndeterminate || len(unmeasured) != 3 {
		t.Fatalf("entry-miss verdict = %s unmeasured = %v", verdict, unmeasured)
	}

	// Future heartbeat reads affirmative (clock skew errs toward keeping).
	write([]RegistryEntry{{SessionID: keyDead, PID: deadPID,
		LastHeartbeat: fixtureNow.Add(time.Hour)}})
	verdict, _, evidence = l.Evaluate(keyDead)
	if verdict != VerdictLive || evidence["heartbeat"] != "affirmative" {
		t.Fatalf("future heartbeat: verdict = %s evidence = %v", verdict, evidence)
	}

	// Unparseable registry reads as absent: pid+heartbeat unmeasured
	// (transcript unmeasured as well), never DEAD.
	if err := os.WriteFile(registryPath, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("break registry: %v", err)
	}
	verdict, unmeasured, _ = l.Evaluate(keyDead)
	if verdict != VerdictIndeterminate || len(unmeasured) != 3 {
		t.Fatalf("broken registry verdict = %s unmeasured = %v", verdict, unmeasured)
	}
}

// TestNilClockDefaults — the production clock accessors default to
// time.Now when no injected clock is installed.
func TestNilClockDefaults(t *testing.T) {
	r := &Rotator{}
	if r.pnow().IsZero() {
		t.Fatalf("rotator nil clock returned zero time")
	}
	g := &GC{}
	if g.gnow().IsZero() {
		t.Fatalf("gc nil clock returned zero time")
	}
	l := &Liveness{}
	if l.lnow().IsZero() {
		t.Fatalf("liveness nil clock returned zero time")
	}
}
