package kanban

// slot_lease_test.go — single-process behaviour of the resource slot lease
// (SPEC-RESOURCE-SLOT-LEASE-001, card t607, M1 RED).
//
// These tests pin the decisions M1 fixes before any implementation exists:
// the on-disk record schema and path, the audit-log vocabulary, the held/busy
// distinction, the liveness and declared-bound semantics, release ownership,
// resource-name validation, the separation from the integration window, and
// the ONE root-resolution function the CLI and the guard share (plan-audit N1).
//
// Records under test are SEEDED by writing JSON straight to the pinned path,
// not through AcquireSlotLease, so each test exercises exactly one operation
// against a known on-disk state — and so the schema is pinned by the reader
// side as well as the writer side.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

const slotTestResource = "demo"

// scrubSlotLeaseEnv pins the environment the SPEC's isolation clause names.
// The lane runtime stamps a real session id and owner pid into this process;
// nothing under test may pick them up.
func scrubSlotLeaseEnv(t *testing.T, root string) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(root))
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	t.Setenv("MOAI_SESSION_PID", "")
}

// slotRecordPath is the pinned record location (plan.md §B2).
func slotRecordPath(root, resource string) string {
	return filepath.Join(root, ".moai", "state", "slot-leases", resource+".json")
}

// slotMutationLockPath is the pinned per-resource mutation-lock location.
func slotMutationLockPath(root, resource string) string {
	return filepath.Join(root, ".moai", "state", "slot-leases", resource+".mutation.lock")
}

// slotAuditPath is the pinned audit-log location.
func slotAuditPath(root string) string {
	return filepath.Join(root, ".moai", "logs", "slot-lease-audit.jsonl")
}

func rfc(t time.Time) string { return t.UTC().Format(time.RFC3339) }

// seedSlotLease writes a record in the pinned schema directly to disk.
func seedSlotLease(t *testing.T, root string, lease SlotLease) {
	t.Helper()
	path := slotRecordPath(root, lease.Resource)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("seed mkdir: %v", err)
	}
	data, err := json.MarshalIndent(&lease, "", "  ")
	if err != nil {
		t.Fatalf("seed marshal: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("seed write: %v", err)
	}
}

// liveLease returns a record held by session with a LIVE owner (this test
// process) whose declared bound has not elapsed.
func liveLease(session string) SlotLease {
	now := time.Now()
	return SlotLease{
		Resource:    slotTestResource,
		SessionID:   session,
		SessionName: "lane-" + session,
		PID:         os.Getpid(),
		PIDSource:   PIDSourceSessionOwner,
		Command:     "heavy-suite",
		AcquiredAt:  rfc(now.Add(-time.Minute)),
		MaxDuration: (time.Hour).String(),
		ExpiresAt:   rfc(now.Add(time.Hour)),
	}
}

func fileSHA(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// slotAuditEvents parses the audit log. An absent log is an empty slice.
func slotAuditEvents(t *testing.T, root string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(slotAuditPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("audit line is not JSON (%v): %q", err, line)
		}
		out = append(out, m)
	}
	return out
}

// lastAudit returns the last audit entry's event and reason, failing when the
// log is empty.
func lastAudit(t *testing.T, root string) (event, reason string) {
	t.Helper()
	events := slotAuditEvents(t, root)
	if len(events) == 0 {
		t.Fatalf("audit log %s has no entry", slotAuditPath(root))
	}
	last := events[len(events)-1]
	event, _ = last["event"].(string)
	reason, _ = last["reason"].(string)
	return event, reason
}

// deadChildPID returns the pid of a child process that has already exited.
func deadChildPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	if err := cmd.Run(); err != nil {
		t.Fatalf("running the dummy child: %v", err)
	}
	pid := cmd.Process.Pid
	if FactoryProcessAlive(pid) {
		t.Skipf("pid %d of an exited child reads live on this machine (reused); the stale path is not exercisable", pid)
	}
	return pid
}

// AC-RSL-003a/b (kanban half) — every field lands on disk, omitted fields are
// recorded as empty strings rather than dropped or invented.
func TestSlotLease_RecordsAllFields(t *testing.T) {
	root := t.TempDir()
	scrubSlotLeaseEnv(t, root)
	fixed := time.Date(2026, 9, 12, 1, 2, 3, 0, time.UTC)

	t.Run("all_fields", func(t *testing.T) {
		got, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource:    slotTestResource,
			SessionID:   "s-1",
			SessionName: "lane-a",
			PID:         os.Getpid(),
			Command:     "heavy-suite",
			MaxDuration: 5 * time.Minute,
			Now:         fixed,
		})
		if err != nil {
			t.Fatalf("acquire on a free resource: %v", err)
		}
		if got == nil || got.SessionID != "s-1" {
			t.Fatalf("returned lease = %+v, want holder s-1", got)
		}

		data, err := os.ReadFile(slotRecordPath(root, slotTestResource))
		if err != nil {
			t.Fatalf("record not at the pinned path: %v", err)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("record is not JSON: %v", err)
		}
		want := map[string]any{
			"resource":     slotTestResource,
			"session_id":   "s-1",
			"session_name": "lane-a",
			"pid_source":   PIDSourceSessionOwner,
			"command":      "heavy-suite",
			"acquired_at":  "2026-09-12T01:02:03Z",
			"expires_at":   "2026-09-12T01:07:03Z",
		}
		for key, v := range want {
			if raw[key] != v {
				t.Errorf("record[%q] = %v, want %v", key, raw[key], v)
			}
		}
		// The pid is the caller-resolved owner, recorded verbatim — never
		// substituted by the process that wrote the record.
		if pid, _ := raw["pid"].(float64); int(pid) != os.Getpid() {
			t.Errorf("record pid = %v, want the supplied owner pid %d", raw["pid"], os.Getpid())
		}
		bound, _ := raw["max_duration"].(string)
		if d, perr := time.ParseDuration(bound); perr != nil || d != 5*time.Minute {
			t.Errorf("record max_duration = %q, want a duration string equal to 5m", bound)
		}
	})

	t.Run("omitted_name_and_command_recorded_empty", func(t *testing.T) {
		root := t.TempDir()
		if _, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource:    slotTestResource,
			SessionID:   "s-1",
			PID:         os.Getpid(),
			MaxDuration: 5 * time.Minute,
			Now:         fixed,
		}); err != nil {
			t.Fatalf("acquire without --name/--command must succeed: %v", err)
		}
		data, err := os.ReadFile(slotRecordPath(root, slotTestResource))
		if err != nil {
			t.Fatalf("record not written: %v", err)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("record is not JSON: %v", err)
		}
		for _, key := range []string{"session_name", "command"} {
			v, present := raw[key]
			if !present {
				t.Errorf("record has no %q key; an omitted value is recorded as an empty string, not dropped", key)
				continue
			}
			if v != "" {
				t.Errorf("record[%q] = %v, want the empty string (never an invented value)", key, v)
			}
		}
	})
}

// AC-RSL-004 — a live, unexpired foreign holder refuses the acquire and the
// refusal changes nothing on disk.
func TestSlotLease_RefusesLiveForeignHolderWritesNothing(t *testing.T) {
	root := t.TempDir()
	scrubSlotLeaseEnv(t, root)
	seedSlotLease(t, root, liveLease("s-1"))
	before := fileSHA(t, slotRecordPath(root, slotTestResource))

	_, err := AcquireSlotLease(root, SlotLeaseRequest{
		Resource: slotTestResource, SessionID: "s-2", PID: os.Getpid(), MaxDuration: time.Hour,
	})
	if !IsSlotLeaseHeld(err) {
		t.Fatalf("acquire over a live foreign holder: err = %v, want the held sentinel", err)
	}
	if !strings.Contains(err.Error(), "lane-s-1") {
		t.Errorf("refusal does not name the holder: %v", err)
	}
	if after := fileSHA(t, slotRecordPath(root, slotTestResource)); after != before {
		t.Errorf("record changed on a refusal (sha %s -> %s)", before, after)
	}
	if event, _ := lastAudit(t, root); event != "refuse" {
		t.Errorf("last audit event = %q, want refuse", event)
	}
}

// AC-RSL-002 — busy is not held, in both directions. The test holds the
// per-resource mutation lock itself for longer than the wait budget.
func TestSlotLeaseBusy_IsNotHeld(t *testing.T) {
	root := t.TempDir()
	scrubSlotLeaseEnv(t, root)
	seedSlotLease(t, root, liveLease("s-1"))
	recordPath := slotRecordPath(root, slotTestResource)
	before := fileSHA(t, recordPath)

	impl, err := acquireBoardLockImpl(slotMutationLockPath(root, slotTestResource))
	if err != nil {
		t.Fatalf("taking the per-resource mutation lock at the pinned path: %v", err)
	}
	released := false
	t.Cleanup(func() {
		if !released {
			_ = impl.release()
		}
	})

	assertBusy := func(t *testing.T, op string, err error, elapsed time.Duration) {
		t.Helper()
		if !IsSlotLeaseBusy(err) {
			t.Fatalf("%s under a contended mutation lock: err = %v, want the busy sentinel", op, err)
		}
		if IsSlotLeaseHeld(err) {
			t.Fatalf("%s: busy error also answers the held predicate — a transient contention would be reported as another session holding the resource", op)
		}
		if IsBoardLockHeld(err) {
			t.Fatalf("%s: the board-lock sentinel leaked across the scope boundary", op)
		}
		if elapsed < boardLockWaitBudget {
			t.Errorf("%s returned busy after %s, before the %s wait budget was spent", op, elapsed, boardLockWaitBudget)
		}
		if after := fileSHA(t, recordPath); after != before {
			t.Fatalf("%s changed the record while busy", op)
		}
	}

	t.Run("acquire_busy", func(t *testing.T) {
		start := time.Now()
		_, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource: slotTestResource, SessionID: "s-2", PID: os.Getpid(), MaxDuration: time.Hour,
		})
		assertBusy(t, "acquire", err, time.Since(start))
	})

	t.Run("release_busy", func(t *testing.T) {
		start := time.Now()
		_, err := ReleaseSlotLease(root, slotTestResource, "s-1", false)
		assertBusy(t, "release", err, time.Since(start))
	})

	t.Run("held_is_not_busy", func(t *testing.T) {
		if err := impl.release(); err != nil {
			t.Fatalf("releasing the mutation lock: %v", err)
		}
		released = true
		_, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource: slotTestResource, SessionID: "s-2", PID: os.Getpid(), MaxDuration: time.Hour,
		})
		if !IsSlotLeaseHeld(err) {
			t.Fatalf("acquire over a live holder with a free mutation lock: err = %v, want the held sentinel", err)
		}
		if IsSlotLeaseBusy(err) {
			t.Fatalf("held error also answers the busy predicate — the discriminator does not hold in this direction")
		}
	})
}

// AC-RSL-005 — liveness: a dead owner is taken over (recorded as stale), a
// live owner and an unresolvable (pid 0) owner are refused.
func TestSlotLease_Liveness(t *testing.T) {
	t.Run("stale_takeover", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		dead := liveLease("s-1")
		dead.PID = deadChildPID(t)
		seedSlotLease(t, root, dead)

		got, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource: slotTestResource, SessionID: "s-2", PID: os.Getpid(), MaxDuration: time.Hour,
		})
		if err != nil {
			t.Fatalf("acquire over a dead owner without --force: %v", err)
		}
		if got == nil || got.Displaced == nil {
			t.Fatalf("takeover did not record the displaced holder: %+v", got)
		}
		if got.Displaced.SessionID != "s-1" || got.Displaced.Reason != SlotDisplacedStale {
			t.Errorf("displaced = %+v, want session s-1 with reason %q", got.Displaced, SlotDisplacedStale)
		}
		onDisk, err := ReadSlotLease(root, slotTestResource)
		if err != nil || onDisk.Displaced == nil || onDisk.Displaced.Reason != SlotDisplacedStale {
			t.Errorf("record on disk does not carry the stale displacement: %+v (err %v)", onDisk, err)
		}
		if event, reason := lastAudit(t, root); event != "takeover" || reason != SlotDisplacedStale {
			t.Errorf("last audit = (%q, %q), want (takeover, stale)", event, reason)
		}
	})

	t.Run("live_refused", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		seedSlotLease(t, root, liveLease("s-1"))
		_, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource: slotTestResource, SessionID: "s-2", PID: os.Getpid(), MaxDuration: time.Hour,
		})
		if !IsSlotLeaseHeld(err) {
			t.Fatalf("acquire over a live owner: err = %v, want the held sentinel", err)
		}
	})

	t.Run("pid_zero_live", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		// The caller could not resolve its owner: pid 0 arrives and must be
		// recorded as 0, never replaced by the acquiring process's own pid.
		if _, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource: slotTestResource, SessionID: "s-1", PID: 0, MaxDuration: time.Hour,
		}); err != nil {
			t.Fatalf("acquire with an unresolvable owner: %v", err)
		}
		rec, err := ReadSlotLease(root, slotTestResource)
		if err != nil {
			t.Fatalf("read back: %v", err)
		}
		if rec.PID != 0 {
			t.Errorf("record pid = %d, want 0 (the acquiring process's pid %d must not be substituted)", rec.PID, os.Getpid())
		}
		if rec.PIDSource != PIDSourceSessionOwner {
			t.Errorf("pid_source = %q, want %q", rec.PIDSource, PIDSourceSessionOwner)
		}
		if !rec.Held() || rec.Stale() {
			t.Errorf("pid-0 record: held=%v stale=%v, want held and not stale", rec.Held(), rec.Stale())
		}
		if _, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource: slotTestResource, SessionID: "s-2", PID: os.Getpid(), MaxDuration: time.Hour,
		}); !IsSlotLeaseHeld(err) {
			t.Errorf("acquire over a pid-0 holder: err = %v, want the held sentinel", err)
		}
	})
}

// AC-RSL-006 — the declared bound: past it a live owner is taken over (recorded
// as expired), before it the owner is refused, the holder's own re-acquire
// restarts it, and a non-positive or unparseable bound is refused.
func TestSlotLease_DeclaredBound(t *testing.T) {
	t.Run("expired_takeover", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		now := time.Now()
		expired := liveLease("s-1")
		expired.AcquiredAt = rfc(now.Add(-2 * time.Hour))
		expired.ExpiresAt = rfc(now.Add(-time.Hour))
		seedSlotLease(t, root, expired)

		before, err := ReadSlotLease(root, slotTestResource)
		if err != nil {
			t.Fatalf("read before acquire: %v", err)
		}
		if !before.Expired(now) {
			t.Errorf("status before the takeover does not report the elapsed bound")
		}
		got, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource: slotTestResource, SessionID: "s-2", PID: os.Getpid(), MaxDuration: time.Hour, Now: now,
		})
		if err != nil {
			t.Fatalf("acquire over an expired live owner without --force: %v", err)
		}
		if got == nil || got.Displaced == nil || got.Displaced.Reason != SlotDisplacedExpired {
			t.Fatalf("displaced = %+v, want reason %q", got, SlotDisplacedExpired)
		}
		if event, reason := lastAudit(t, root); event != "takeover" || reason != SlotDisplacedExpired {
			t.Errorf("last audit = (%q, %q), want (takeover, expired)", event, reason)
		}
	})

	t.Run("not_expired_refused", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		seedSlotLease(t, root, liveLease("s-1"))
		if _, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource: slotTestResource, SessionID: "s-2", PID: os.Getpid(), MaxDuration: time.Hour,
		}); !IsSlotLeaseHeld(err) {
			t.Fatalf("acquire over an unexpired live owner: err = %v, want the held sentinel", err)
		}
	})

	t.Run("holder_reacquire_restarts_bound", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		now := time.Now().UTC().Truncate(time.Second)
		held := liveLease("s-1")
		held.AcquiredAt = rfc(now.Add(-10 * time.Minute))
		held.MaxDuration = (30 * time.Minute).String()
		held.ExpiresAt = rfc(now.Add(20 * time.Minute))
		seedSlotLease(t, root, held)

		got, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource: slotTestResource, SessionID: "s-1", PID: os.Getpid(), MaxDuration: 30 * time.Minute, Now: now,
		})
		if err != nil {
			t.Fatalf("holder re-acquire: %v", err)
		}
		if got == nil || got.ExpiresAt != rfc(now.Add(30*time.Minute)) {
			t.Fatalf("expires_at after re-acquire = %+v, want %s (re-acquire time + bound)", got, rfc(now.Add(30*time.Minute)))
		}
		if got.Displaced != nil {
			t.Errorf("a holder's own re-acquire recorded a displacement: %+v", got.Displaced)
		}
	})

	t.Run("invalid_bound_rejected", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		for _, bound := range []time.Duration{0, -5 * time.Minute} {
			_, err := AcquireSlotLease(root, SlotLeaseRequest{
				Resource: slotTestResource, SessionID: "s-1", PID: os.Getpid(), MaxDuration: bound,
			})
			if !errors.Is(err, ErrSlotLeaseBoundInvalid) {
				t.Errorf("acquire with bound %s: err = %v, want ErrSlotLeaseBoundInvalid", bound, err)
			}
		}
		for _, text := range []string{"", "0", "-5m", "abc", "5"} {
			if _, err := ParseSlotLeaseMaxDuration(text); !errors.Is(err, ErrSlotLeaseBoundInvalid) {
				t.Errorf("ParseSlotLeaseMaxDuration(%q): err = %v, want ErrSlotLeaseBoundInvalid", text, err)
			}
		}
		// Positive control: a well-formed bound parses, so the refusals above
		// are about the values and not a parser that refuses everything.
		if d, err := ParseSlotLeaseMaxDuration("5m"); err != nil || d != 5*time.Minute {
			t.Errorf("ParseSlotLeaseMaxDuration(\"5m\") = (%s, %v), want (5m, nil)", d, err)
		}
		if _, err := os.Stat(slotRecordPath(root, slotTestResource)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("a refused bound left a record behind (stat err %v)", err)
		}
	})
}

// AC-RSL-007 — a forced takeover records the displaced holder in the record
// and in the audit log, not only on standard output.
func TestSlotLease_ForceRecordsDisplaced(t *testing.T) {
	root := t.TempDir()
	scrubSlotLeaseEnv(t, root)
	seedSlotLease(t, root, liveLease("s-1"))

	got, err := AcquireSlotLease(root, SlotLeaseRequest{
		Resource: slotTestResource, SessionID: "s-2", PID: os.Getpid(), MaxDuration: time.Hour, Force: true,
	})
	if err != nil {
		t.Fatalf("forced acquire over a live holder: %v", err)
	}
	onDisk, err := ReadSlotLease(root, slotTestResource)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	for label, lease := range map[string]*SlotLease{"returned": got, "on disk": onDisk} {
		d := lease.Displaced
		if d == nil {
			t.Errorf("%s lease has no displaced holder", label)
			continue
		}
		if d.SessionID != "s-1" || d.PID != os.Getpid() || d.Reason != SlotDisplacedForce || d.At == "" {
			t.Errorf("%s displaced = %+v, want session s-1, pid %d, reason force, a timestamp", label, d, os.Getpid())
		}
	}
	if event, reason := lastAudit(t, root); event != "takeover" || reason != SlotDisplacedForce {
		t.Errorf("last audit = (%q, %q), want (takeover, force)", event, reason)
	}
}

// AC-RSL-008 — only the holder releases; a foreign release is refused without
// touching the record; releasing nothing is an error.
func TestSlotLease_Release(t *testing.T) {
	root := t.TempDir()
	scrubSlotLeaseEnv(t, root)
	seedSlotLease(t, root, liveLease("s-1"))
	recordPath := slotRecordPath(root, slotTestResource)
	before := fileSHA(t, recordPath)

	t.Run("foreign_refused", func(t *testing.T) {
		_, err := ReleaseSlotLease(root, slotTestResource, "s-2", false)
		if !IsSlotLeaseForeign(err) {
			t.Fatalf("foreign release: err = %v, want the foreign sentinel", err)
		}
		if after := fileSHA(t, recordPath); after != before {
			t.Errorf("record changed on a refused foreign release")
		}
	})

	t.Run("holder_releases", func(t *testing.T) {
		got, err := ReleaseSlotLease(root, slotTestResource, "s-1", false)
		if err != nil {
			t.Fatalf("holder release: %v", err)
		}
		if got == nil || got.SessionID != "s-1" {
			t.Errorf("release did not report the freed holder: %+v", got)
		}
		if _, err := os.Stat(recordPath); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("record survived its release (stat err %v)", err)
		}
		if event, _ := lastAudit(t, root); event != "release" {
			t.Errorf("last audit event = %q, want release", event)
		}
	})

	t.Run("empty_release_is_error", func(t *testing.T) {
		if _, err := ReleaseSlotLease(root, slotTestResource, "s-1", false); !IsSlotLeaseNotHeld(err) {
			t.Errorf("releasing an unheld resource: err = %v, want the not-held sentinel", err)
		}
	})
}

// treeSnapshot lists every path under dir (relative), for the no-write check.
func treeSnapshot(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		out = append(out, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	sort.Strings(out)
	return out
}

// AC-RSL-009 — a resource name is a trust boundary: an invalid one is refused
// by every operation and no file is written anywhere, inside or outside the
// lease directory.
func TestSlotLease_RejectsInvalidResourceNames(t *testing.T) {
	outer := t.TempDir()
	root := filepath.Join(outer, "project")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	scrubSlotLeaseEnv(t, root)

	invalid := []string{
		"../escape",
		"a/b",
		`a\b`,
		"",
		strings.Repeat("a", 65),
		"Demo",
		"a b",
		"..",
	}
	before := treeSnapshot(t, outer)
	for _, name := range invalid {
		if err := ValidateSlotResourceName(name); !IsSlotResourceNameInvalid(err) {
			t.Errorf("ValidateSlotResourceName(%q): err = %v, want the invalid-name sentinel", name, err)
		}
		if _, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource: name, SessionID: "s-1", PID: os.Getpid(), MaxDuration: time.Hour,
		}); !IsSlotResourceNameInvalid(err) {
			t.Errorf("acquire(%q): err = %v, want the invalid-name sentinel", name, err)
		}
		if _, err := ReleaseSlotLease(root, name, "s-1", true); !IsSlotResourceNameInvalid(err) {
			t.Errorf("release(%q): err = %v, want the invalid-name sentinel", name, err)
		}
		if _, err := ReadSlotLease(root, name); !IsSlotResourceNameInvalid(err) {
			t.Errorf("read(%q): err = %v, want the invalid-name sentinel", name, err)
		}
	}
	if after := treeSnapshot(t, outer); strings.Join(after, "\n") != strings.Join(before, "\n") {
		t.Errorf("invalid names wrote files:\n before: %v\n after:  %v", before, after)
	}

	// Positive control: the boundary names ARE accepted, so the refusals above
	// are about the names and not a validator that refuses everything.
	for _, name := range []string{"demo", "heavy-test-1", strings.Repeat("a", 64)} {
		if err := ValidateSlotResourceName(name); err != nil {
			t.Errorf("ValidateSlotResourceName(%q) = %v, want nil", name, err)
		}
	}
}

// AC-RSL-013(a) — the slot lease and the integration window share no record,
// lock file, or directory, in either direction.
func TestSlotLease_SeparateFromIntegrationWindow(t *testing.T) {
	t.Run("slot_acquire_creates_no_integration_files", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		if _, err := AcquireSlotLease(root, SlotLeaseRequest{
			Resource: slotTestResource, SessionID: "s-1", PID: os.Getpid(), MaxDuration: time.Hour,
		}); err != nil {
			t.Fatalf("slot acquire: %v", err)
		}
		for _, name := range []string{IntegrationLockFileName, integrationMutationLockFileName} {
			if _, err := os.Stat(filepath.Join(root, ".moai", "state", name)); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("slot acquire created the integration window's %s", name)
			}
		}
	})

	t.Run("integration_acquire_creates_no_slot_dir", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		if _, err := AcquireIntegrationLock(root, IntegrationLock{SessionID: "s-1", PID: os.Getpid()}, false); err != nil {
			t.Fatalf("integration acquire: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, ".moai", "state", SlotLeaseDirName)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("integration acquire created the slot-lease directory")
		}
	})
}

// plan.md §B4 — the lease core treats an unreadable record as an ERROR, never
// as a free resource; only an explicit --force clears it. Plus the remaining
// input edges of the core (missing session id, empty start, empty audit root,
// unparseable expiry, the holder-label fallback).
func TestSlotLease_UnreadableRecordAndInputEdges(t *testing.T) {
	writeCorrupt := func(t *testing.T, root string) {
		t.Helper()
		path := slotRecordPath(root, slotTestResource)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte("{ not json"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	t.Run("corrupt_record_is_not_free", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		writeCorrupt(t, root)
		if _, err := ReadSlotLease(root, slotTestResource); err == nil {
			t.Fatal("a corrupt record read as a valid lease")
		}
		if _, err := AcquireSlotLease(root, SlotLeaseRequest{Resource: slotTestResource, SessionID: "s-2", MaxDuration: time.Hour}); err == nil {
			t.Fatal("acquire over a corrupt record succeeded without --force")
		}
		if _, err := ReleaseSlotLease(root, slotTestResource, "s-2", false); err == nil {
			t.Fatal("release over a corrupt record succeeded without --force")
		}
	})

	t.Run("force_clears_corrupt_record", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		writeCorrupt(t, root)
		if _, err := ReleaseSlotLease(root, slotTestResource, "s-2", true); err != nil {
			t.Fatalf("forced release of a corrupt record: %v", err)
		}
		writeCorrupt(t, root)
		got, err := AcquireSlotLease(root, SlotLeaseRequest{Resource: slotTestResource, SessionID: "s-2", MaxDuration: time.Hour, Force: true})
		if err != nil || got.SessionID != "s-2" {
			t.Fatalf("forced acquire over a corrupt record = (%+v, %v), want s-2", got, err)
		}
	})

	t.Run("forced_foreign_release_is_audited_as_force", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		seedSlotLease(t, root, liveLease("s-1"))
		if _, err := ReleaseSlotLease(root, slotTestResource, "s-2", true); err != nil {
			t.Fatalf("forced foreign release: %v", err)
		}
		if event, reason := lastAudit(t, root); event != "release" || reason != SlotDisplacedForce {
			t.Errorf("last audit = (%q, %q), want (release, force)", event, reason)
		}
	})

	t.Run("input_edges", func(t *testing.T) {
		root := t.TempDir()
		scrubSlotLeaseEnv(t, root)
		if _, err := AcquireSlotLease(root, SlotLeaseRequest{Resource: slotTestResource, MaxDuration: time.Hour}); err == nil {
			t.Error("acquire without a session id succeeded; a holder identity must never be invented")
		}
		if _, err := ResolveSlotLeaseRoot("  "); err == nil {
			t.Error("ResolveSlotLeaseRoot(empty) returned no error")
		}
		if err := AppendSlotLeaseAudit("", SlotLeaseAuditEntry{Event: "acquire"}); err == nil {
			t.Error("AppendSlotLeaseAudit with no root returned no error")
		}
		odd := SlotLease{Resource: slotTestResource, SessionID: "s-1", ExpiresAt: "not-a-time"}
		if odd.Expired(time.Now()) {
			t.Error("an unparseable expiry reads as expired; the conservative reading is not expired")
		}
		if (&SlotLease{}).Expired(time.Now()) {
			t.Error("an empty lease reads as expired")
		}
		for _, tc := range []struct {
			lease *SlotLease
			want  string
		}{
			{nil, "unknown"},
			{&SlotLease{}, "unknown"},
			{&SlotLease{SessionID: "s-9"}, "s-9"},
			{&SlotLease{SessionID: "s-9", SessionName: "lane-9"}, "lane-9"},
		} {
			if got := tc.lease.holderLabel(); got != tc.want {
				t.Errorf("holderLabel(%+v) = %q, want %q", tc.lease, got, tc.want)
			}
		}
	})
}

// slotGitFixture builds a primary repository with one commit and a linked
// worktree, under one temp parent pinned as the git ceiling.
func slotGitFixture(t *testing.T) (primary, worktree string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not in PATH: %v", err)
	}
	parent := t.TempDir()
	primary = filepath.Join(parent, "primary")
	worktree = filepath.Join(parent, "wt")
	if err := os.MkdirAll(filepath.Join(primary, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	git := func(args ...string) {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", primary}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "slot-lease-test@example.com")
	git("config", "user.name", "Slot Lease Test")
	git("config", "core.hooksPath", "/dev/null")
	git("commit", "-q", "--allow-empty", "-m", "seed")
	git("worktree", "add", "-q", "--detach", worktree)
	return primary, worktree
}

func realPath(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s): %v", p, err)
	}
	return r
}

// AC-RSL-016 (shared-resolver half) and plan-audit N1 — ONE function maps any
// start directory inside a repository, including a linked worktree, to the
// primary checkout, and refuses to answer outside a repository. The CLI and
// the PreToolUse guard both call it, so a root one of them writes is the root
// the other reads.
func TestResolveSlotLeaseRoot_NormalizesToPrimary(t *testing.T) {
	primary, worktree := slotGitFixture(t)
	scrubSlotLeaseEnv(t, primary)
	want := realPath(t, primary)

	for _, tc := range []struct{ name, start string }{
		{"primary", primary},
		{"primary_subdir", filepath.Join(primary, "sub")},
		{"linked_worktree", worktree},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveSlotLeaseRoot(tc.start)
			if err != nil {
				t.Fatalf("ResolveSlotLeaseRoot(%s): %v", tc.start, err)
			}
			if realPath(t, got) != want {
				t.Errorf("ResolveSlotLeaseRoot(%s) = %s, want the primary checkout %s", tc.start, got, want)
			}
		})
	}

	t.Run("not_a_repository", func(t *testing.T) {
		plain := t.TempDir()
		t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(plain))
		got, err := ResolveSlotLeaseRoot(plain)
		if err == nil {
			t.Fatalf("ResolveSlotLeaseRoot(non-repo) = %q with no error; an unresolvable root must be reported, never replaced by the unnormalized start", got)
		}
		if !strings.Contains(err.Error(), plain) {
			t.Errorf("ResolveSlotLeaseRoot(non-repo) error does not name the directory it could not resolve: %v", err)
		}
	})
}
