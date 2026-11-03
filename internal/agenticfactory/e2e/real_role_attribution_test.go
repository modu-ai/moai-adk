package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// producerGated is the record after a verify refused for the role lock: since
// lane-18 3dd74138d2 (behavior change 1) a non-passing verify gate keeps the
// stage at verifying and records a Gate, instead of pausing as blocked.
func producerGated(rec claim.Record) bool {
	return rec.State == core.StateVerifying && rec.Gate != nil && rec.Gate.Reason == core.ReasonProducerNotAllowed
}

// gateText is everything the gate says to a reader.
func gateText(rec claim.Record) string {
	if rec.Gate == nil {
		return ""
	}
	return rec.Gate.Detail + " " + rec.Gate.Resume
}

// AC11 (DEC-AF-01 / DEC-AF-02) on lane-18's runner: review, audit, QA and
// coordinator evidence counts only when an attribution lookup says it was
// produced in the required role; with no lookup (the identity SPEC is not
// complete) none of it counts. The writer handoff must come from the holder and
// epoch the attempt was issued to. Nothing the evidence says about itself is used.

// verifyingTicket is a runner WITHOUT an attributor, at VERIFYING.
func lockedRunner(t *testing.T) (*runner.Runner, runner.Ticket) {
	t.Helper()
	st, ref, run := realStore(t)
	r := runner.New(st) // no Attributor: the lock
	r.Facts = myIssuesFacts()
	tk, err := r.Claim(ref, "w1", "d-1", run)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.StartAttempt(tk, "att-1", core.DeliveryExecuted); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Submit(tk, headA, false); err != nil {
		t.Fatal(err)
	}
	return r, tk
}

// With no attribution source, a complete all-PASS same-HEAD evidence set is
// still blocked, with producer-not-allowed and the lock reason recorded.
func TestRoleLock_NoAttributor_BlocksACompletePassSet(t *testing.T) {
	r, tk := lockedRunner(t)
	rec, err := r.Verify(tk, passAll(tk, headA))
	if err != nil {
		t.Fatal(err)
	}
	// Behavior change 1: expectation updated from blocked/producer-not-allowed to verifying + gate.
	if !producerGated(rec) {
		t.Fatalf("record = state %s gate %+v; want verifying with a producer-not-allowed gate", rec.State, rec.Gate)
	}
	if !strings.Contains(gateText(rec), core.AttributionLocked) {
		t.Errorf("the gate text %q does not name the lock (%s)", gateText(rec), core.AttributionLocked)
	}
	if _, err := r.Land(tk, headA); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("land after a locked verify: %v, want ErrTransitionInvalid", err)
	}
}

// Control: the same set with an attributor that names the right roles passes.
// Only the attribution source differs, so "refuse everything" is caught.
func TestRoleLock_RightRoles_PassTheSameSet(t *testing.T) {
	r, tk := newRunner(t)
	for _, step := range []func() error{
		func() error { _, err := r.StartAttempt(tk, "att-1", core.DeliveryExecuted); return err },
		func() error { _, err := r.Submit(tk, headA, false); return err },
	} {
		if err := step(); err != nil {
			t.Fatal(err)
		}
	}
	rec, err := r.Verify(tk, passAll(tk, headA))
	if err != nil || rec.State != core.StateReady {
		t.Fatalf("with the right roles: %+v, %v; want ready", rec, err)
	}
}

// What the evidence claims about its own producer does not matter: the same
// lock holds when the file would say "I am the auditor".
func TestRoleLock_EvidenceCannotVouchForItself(t *testing.T) {
	r, tk := lockedRunner(t)
	v := passAll(tk, headA)
	for i := range v.Results {
		v.Results[i].Name = "auditor-signed-by-the-auditor"
		v.Results[i].Session = "producer-" + string(v.Results[i].Kind)
	}
	rec, err := r.Verify(tk, v)
	if err != nil || !producerGated(rec) {
		t.Errorf("self-described producers: state %s gate %+v, %v; want verifying + producer-not-allowed gate", rec.State, rec.Gate, err)
	}
}

// A result produced in the wrong role is refused: the auditor's slot filled by
// a reviewer does not count.
type swappedAttributor struct{}

func (swappedAttributor) Attribute(res runner.Result) runner.Attribution {
	return runner.Attribution{Attributed: true, Role: core.RoleReviewer, Session: "someone"} // reviewer for everything
}

func TestRoleLock_WrongRole_IsRefused(t *testing.T) {
	r, tk := lockedRunner(t)
	r.Attributor = swappedAttributor{}
	rec, err := r.Verify(tk, passAll(tk, headA))
	if err != nil || !producerGated(rec) {
		t.Fatalf("wrong roles: state %s gate %+v, %v; want verifying + producer-not-allowed gate", rec.State, rec.Gate, err)
	}
	if !strings.Contains(gateText(rec), "role") {
		t.Errorf("the gate text %q does not name the role mismatch", gateText(rec))
	}
}

// An unattributed answer is refused whatever its reason text says.
type unattributed struct{ reason string }

func (u unattributed) Attribute(runner.Result) runner.Attribution {
	return runner.Attribution{Attributed: false, Reason: u.reason}
}

func TestRoleLock_UnattributedAnswer_IsRefused(t *testing.T) {
	r, tk := lockedRunner(t)
	r.Attributor = unattributed{reason: "no grant on record"}
	rec, err := r.Verify(tk, passAll(tk, headA))
	if err != nil || !producerGated(rec) {
		t.Fatalf("unattributed: state %s gate %+v, %v", rec.State, rec.Gate, err)
	}
	if !strings.Contains(gateText(rec), "no grant on record") {
		t.Errorf("the gate text %q lost the reason", gateText(rec))
	}
}

// The writer handoff must come from the holder and epoch the attempt was issued
// to: another session, or the right session at another epoch, is refused.
func TestRoleLock_WriterHandoff_MustBeFromTheIssuedHolderAndEpoch(t *testing.T) {
	cases := map[string]func(*runner.Result){
		"other-session": func(r *runner.Result) { r.Session = "another-session" },
		"other-epoch":   func(r *runner.Result) { r.Epoch++ },
		"no-session":    func(r *runner.Result) { r.Session = "" },
	}
	for name, mod := range cases {
		t.Run(name, func(t *testing.T) {
			r, tk := newRunner(t)
			if _, err := r.StartAttempt(tk, "att-1", core.DeliveryExecuted); err != nil {
				t.Fatal(err)
			}
			if _, err := r.Submit(tk, headA, false); err != nil {
				t.Fatal(err)
			}
			v := passAll(tk, headA)
			for i := range v.Results {
				if v.Results[i].Kind == runner.KindHandoffWriter {
					mod(&v.Results[i])
				}
			}
			rec, err := r.Verify(tk, v)
			if err != nil || !producerGated(rec) {
				t.Errorf("%s: state %s gate %+v, %v; want verifying + producer-not-allowed gate", name, rec.State, rec.Gate, err)
			}
		})
	}
}
