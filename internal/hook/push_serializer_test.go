package hook

// push_serializer_test.go — the push serializer (SPEC-AUTONOMY-PRECONDITION-001
// M1, REQ-AP-001 / REQ-AP-002 / REQ-AP-007; design.md §B).
//
// Four criteria, four tests:
//
//	AC-AP-001  TestPushSerializationDeniesSecondPush
//	AC-AP-002  TestPushSerializationInactiveWhenNotArmed
//	AC-AP-003  TestPushLeaseReleasedOnFailedPush
//	AC-AP-004  TestPushLeaseReclaimAndFailOpen
//
// The activation triple is read from the `moai contract show --json` document
// (spec.md §C.6 — A1 is the field's producer); every fixture here carries that
// JSON document and decodes it through the same DecodePushShowJSON entry point
// the guard uses, so a schema change reaches these tests through A1's
// projection rather than through a second parser.

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

const (
	pushSessionA = "s-A"
	pushSessionB = "s-B"
	pushCommand  = "git push origin develop"
	pushBound    = time.Hour
)

// pushShowDoc builds a `moai contract show --json` document. pushLease nil
// means the field is absent altogether (AC-AP-002 condition (d) — the tree
// state before A1 supplies the field).
func pushShowDoc(t *testing.T, mode string, actions []string, pushLease *bool) []byte {
	t.Helper()
	doc := map[string]any{
		"spec_id": "SPEC-X-001",
		"state":   "signed-valid",
		"mode":    mode,
		"actions": actions,
	}
	if pushLease != nil {
		doc["push_requires_lease"] = *pushLease
	}
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal show --json document: %v", err)
	}
	return data
}

// pushArmedShow decodes an armed show --json document through the guard's
// single decode entry point.
func pushArmedShow(t *testing.T) *PushShowJSON {
	t.Helper()
	truev := true
	show, err := DecodePushShowJSON(pushShowDoc(t, "contract", []string{"push-develop"}, &truev))
	if err != nil {
		t.Fatalf("decode armed show document: %v", err)
	}
	return show
}

// pushLeasePath is the push-develop record under root.
func pushLeasePath(root string) string {
	return slotGuardRecordPath(root, "push-develop")
}

// pushLeaseHeldBy reads the record and reports its holder ("" when unheld).
func pushLeaseHeldBy(t *testing.T, root string) string {
	t.Helper()
	lease, err := kanban.ReadSlotLease(root, "push-develop")
	if err != nil {
		t.Fatalf("read push-develop lease: %v", err)
	}
	if !lease.Held() {
		return ""
	}
	return lease.SessionID
}

// TestPushSerializationDeniesSecondPush pins AC-AP-001: with the activation
// triple armed, A's push of develop is admitted with its hook output unchanged
// and the lease names A; B's push while A holds within the bound is denied
// with the PUSH_SERIALIZATION_VIOLATION: sentinel naming A.
func TestPushSerializationDeniesSecondPush(t *testing.T) {
	root := slotGuardRepo(t)
	show := pushArmedShow(t)
	var advisory bytes.Buffer

	// A's push: admitted, decision empty (hook output unchanged), lease written.
	dec, reason := checkPushSerializer(slotGuardInput(t, pushSessionA, pushCommand), root, show, pushBound, &advisory)
	if dec != "" {
		t.Fatalf("A's push: decision = %q (%s), want admit with empty decision", dec, reason)
	}
	if got := pushLeaseHeldBy(t, root); got != pushSessionA {
		t.Fatalf("A's push: lease holder = %q, want %q", got, pushSessionA)
	}

	// B's push while A holds within the bound: denied with the sentinel naming A.
	dec, reason = checkPushSerializer(slotGuardInput(t, pushSessionB, pushCommand), root, show, pushBound, &advisory)
	if dec != DecisionDeny {
		t.Fatalf("B's push: decision = %q, want %q", dec, DecisionDeny)
	}
	if !strings.HasPrefix(reason, "PUSH_SERIALIZATION_VIOLATION:") {
		t.Fatalf("B's push: reason = %q, want prefix %q", reason, "PUSH_SERIALIZATION_VIOLATION:")
	}
	if !strings.Contains(reason, pushSessionA) {
		t.Fatalf("B's push: reason %q does not name holder %q", reason, pushSessionA)
	}
}

// TestPushSerializationInactiveWhenNotArmed pins AC-AP-002: in each of the
// four off conditions both pushes are allowed, no push-develop record is
// written, and no escalation record appears; in the armed condition (e) the
// same two pushes produce the AC-AP-001 outcome, which is what makes the
// criterion non-vacuous.
func TestPushSerializationInactiveWhenNotArmed(t *testing.T) {
	falsev, truev := false, true
	conditions := []struct {
		name      string
		doc       []byte
		armedWant bool
	}{
		{"mode guided", pushShowDoc(t, "guided", []string{"push-develop"}, &truev), false},
		{"action omitted", pushShowDoc(t, "contract", []string{"commit", "worktree"}, &truev), false},
		{"push_requires_lease false", pushShowDoc(t, "contract", []string{"push-develop"}, &falsev), false},
		{"field absent", pushShowDoc(t, "contract", []string{"push-develop"}, nil), false},
		{"armed", pushShowDoc(t, "contract", []string{"push-develop"}, &truev), true},
	}

	for _, tc := range conditions {
		t.Run(tc.name, func(t *testing.T) {
			root := slotGuardRepo(t)
			show, err := DecodePushShowJSON(tc.doc)
			if err != nil {
				t.Fatalf("decode show document: %v", err)
			}
			var advisory bytes.Buffer

			decA, reasonA := checkPushSerializer(slotGuardInput(t, pushSessionA, pushCommand), root, show, pushBound, &advisory)
			decB, reasonB := checkPushSerializer(slotGuardInput(t, pushSessionB, pushCommand), root, show, pushBound, &advisory)

			if tc.armedWant {
				// Condition (e): the positive control. B must have been denied
				// with the sentinel and the record must name A — a guard that
				// fails open everywhere satisfies (a)-(d) trivially and fails
				// here.
				if decA != "" {
					t.Fatalf("armed condition: A's push decision = %q (%s), want admit", decA, reasonA)
				}
				if decB != DecisionDeny || !strings.HasPrefix(reasonB, "PUSH_SERIALIZATION_VIOLATION:") {
					t.Fatalf("armed condition: B's push decision = %q (%q), want the deny sentinel", decB, reasonB)
				}
				if got := pushLeaseHeldBy(t, root); got != pushSessionA {
					t.Fatalf("armed condition: lease holder = %q, want %q", got, pushSessionA)
				}
				return
			}

			if decA != "" {
				t.Fatalf("A's push: decision = %q (%s), want allow", decA, reasonA)
			}
			if decB != "" {
				t.Fatalf("B's push: decision = %q (%s), want allow", decB, reasonB)
			}

			if _, err := os.Stat(pushLeasePath(root)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("inactive condition: push-develop record exists (%v), want none", err)
			}
			if entries := slotGuardAudit(t, root); entries != nil {
				t.Fatalf("inactive condition: guard wrote %d audit lines, want none", len(entries))
			}
		})
	}
}

// TestPushLeaseReleasedOnFailedPush pins AC-AP-003: A holds after an admitted
// push; the push's PostToolUse reports a non-zero exit; the record is
// released, and B's subsequent push is admitted with B recorded as holder.
func TestPushLeaseReleasedOnFailedPush(t *testing.T) {
	root := slotGuardRepo(t)
	show := pushArmedShow(t)
	var advisory bytes.Buffer

	if dec, reason := checkPushSerializer(slotGuardInput(t, pushSessionA, pushCommand), root, show, pushBound, &advisory); dec != "" {
		t.Fatalf("A's push: decision = %q (%s), want admit", dec, reason)
	}

	// The failed push's PostToolUse: a non-zero exit releases immediately.
	failed := &HookInput{
		SessionID:     pushSessionA,
		HookEventName: "PostToolUse",
		ToolName:      "Bash",
		ToolInput:     pushRawJSON(t, map[string]any{"command": pushCommand}),
		ToolResponse:  pushRawJSON(t, map[string]any{"exit": 1, "stdout": "", "stderr": "rejected"}),
	}
	releasePushLeaseOnFailure(failed, root, show, &advisory)

	if got := pushLeaseHeldBy(t, root); got != "" {
		t.Fatalf("after failed push: lease holder = %q, want released", got)
	}

	// B's subsequent push is admitted and records B as holder.
	if dec, reason := checkPushSerializer(slotGuardInput(t, pushSessionB, pushCommand), root, show, pushBound, &advisory); dec != "" {
		t.Fatalf("B's push after release: decision = %q (%s), want admit", dec, reason)
	}
	if got := pushLeaseHeldBy(t, root); got != pushSessionB {
		t.Fatalf("B's push after release: lease holder = %q, want %q", got, pushSessionB)
	}
}

// TestPushLeaseReclaimAndFailOpen pins AC-AP-004: (a) a lease whose owning
// session is gone is reclaimed with A recorded as displaced; (b) a lease past
// its declared bound is reclaimed the same way; (c) an undecodable record
// fails open with exactly one audit line and no lease written.
func TestPushLeaseReclaimAndFailOpen(t *testing.T) {
	t.Run("stale holder reclaimed", func(t *testing.T) {
		root := slotGuardRepo(t)
		show := pushArmedShow(t)
		dead := exitedChildPID(t)
		seed := liveForeignLease()
		seed.Resource = "push-develop"
		seed.SessionID = pushSessionA
		seed.PID = dead
		seedSlotGuardLease(t, root, seed)

		dec, reason := checkPushSerializer(slotGuardInput(t, pushSessionB, pushCommand), root, show, pushBound, &bytes.Buffer{})
		if dec != "" {
			t.Fatalf("B's push over a stale holder: decision = %q (%s), want admit", dec, reason)
		}
		lease, err := kanban.ReadSlotLease(root, "push-develop")
		if err != nil {
			t.Fatalf("read reclaimed record: %v", err)
		}
		if !lease.Held() || lease.SessionID != pushSessionB {
			t.Fatalf("reclaimed record holder = %q, want %q", lease.SessionID, pushSessionB)
		}
		if lease.Displaced == nil || lease.Displaced.SessionID != pushSessionA {
			t.Fatalf("reclaimed record displaced = %+v, want displaced %q", lease.Displaced, pushSessionA)
		}
	})

	t.Run("expired bound reclaimed", func(t *testing.T) {
		root := slotGuardRepo(t)
		show := pushArmedShow(t)
		now := time.Now().UTC()
		seed := liveForeignLease()
		seed.Resource = "push-develop"
		seed.SessionID = pushSessionA
		seed.PID = os.Getpid()
		seed.AcquiredAt = now.Add(-2 * time.Hour).Format(time.RFC3339)
		seed.ExpiresAt = now.Add(-time.Hour).Format(time.RFC3339)
		seedSlotGuardLease(t, root, seed)

		dec, reason := checkPushSerializer(slotGuardInput(t, pushSessionB, pushCommand), root, show, pushBound, &bytes.Buffer{})
		if dec != "" {
			t.Fatalf("B's push over an expired bound: decision = %q (%s), want admit", dec, reason)
		}
		lease, err := kanban.ReadSlotLease(root, "push-develop")
		if err != nil {
			t.Fatalf("read reclaimed record: %v", err)
		}
		if !lease.Held() || lease.SessionID != pushSessionB {
			t.Fatalf("reclaimed record holder = %q, want %q", lease.SessionID, pushSessionB)
		}
		if lease.Displaced == nil || lease.Displaced.SessionID != pushSessionA {
			t.Fatalf("reclaimed record displaced = %+v, want displaced %q", lease.Displaced, pushSessionA)
		}
	})

	t.Run("undecodable record fails open once", func(t *testing.T) {
		root := slotGuardRepo(t)
		show := pushArmedShow(t)
		path := pushLeasePath(root)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
			t.Fatalf("seed undecodable record: %v", err)
		}
		var advisory bytes.Buffer

		dec, reason := checkPushSerializer(slotGuardInput(t, pushSessionB, pushCommand), root, show, pushBound, &advisory)
		if dec != "" {
			t.Fatalf("B's push over an undecodable record: decision = %q (%s), want allow (fail open)", dec, reason)
		}

		entries := slotGuardAudit(t, root)
		if len(entries) != 1 {
			t.Fatalf("fail-open audit lines = %d, want exactly 1 (%+v)", len(entries), entries)
		}
		if event, _ := entries[0]["event"].(string); event != "fail-open" {
			t.Fatalf("fail-open audit event = %q, want fail-open", event)
		}

		// No lease written: the record is still the undecodable bytes.
		if _, err := kanban.ReadSlotLease(root, "push-develop"); err == nil {
			t.Fatalf("record after fail-open reads clean, want still undecodable (no lease written)")
		}
	})
}

// pushRawJSON marshals v for HookInput raw fields.
func pushRawJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return data
}
