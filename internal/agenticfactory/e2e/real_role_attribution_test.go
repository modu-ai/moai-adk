package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

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
	if rec.State != core.StateBlocked || rec.Reason != core.ReasonProducerNotAllowed {
		t.Fatalf("record = state %s reason %s; want blocked/producer-not-allowed", rec.State, rec.Reason)
	}
	if !strings.Contains(rec.Resume, core.AttributionLocked) {
		t.Errorf("the resume text %q does not name the lock (%s)", rec.Resume, core.AttributionLocked)
	}
	if rec.Return != core.StateVerifying {
		t.Errorf("return stage = %s, want verifying", rec.Return)
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
	if err != nil || rec.State != core.StateBlocked || rec.Reason != core.ReasonProducerNotAllowed {
		t.Errorf("self-described producers: %+v, %v; want blocked/producer-not-allowed", rec, err)
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
	if err != nil || rec.State != core.StateBlocked || rec.Reason != core.ReasonProducerNotAllowed {
		t.Fatalf("wrong roles: %+v, %v; want blocked/producer-not-allowed", rec, err)
	}
	if !strings.Contains(rec.Resume, "role") {
		t.Errorf("the resume text %q does not name the role mismatch", rec.Resume)
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
	if err != nil || rec.State != core.StateBlocked || rec.Reason != core.ReasonProducerNotAllowed {
		t.Fatalf("unattributed: %+v, %v", rec, err)
	}
	if !strings.Contains(rec.Resume, "no grant on record") {
		t.Errorf("the resume text %q lost the reason", rec.Resume)
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
			if err != nil || rec.State != core.StateBlocked || rec.Reason != core.ReasonProducerNotAllowed {
				t.Errorf("%s: %+v, %v; want blocked/producer-not-allowed", name, rec, err)
			}
		})
	}
}
