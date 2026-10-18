package cli

import (
	"context"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
)

// The marker reaches only the codex child of an audit leg (t1788 brain ruling,
// condition a). Ordinary codex paths keep the inherited environment.
func TestCodexLegChildEnvMarksOnlyAuditLeg(t *testing.T) {
	if env := codexLegChildEnv(context.Background()); env != nil {
		t.Fatalf("ordinary codex child must inherit the environment (nil), got %d entries", len(env))
	}
	env := codexLegChildEnv(withCodexAuditLeg(context.Background()))
	if !slices.Contains(env, codexAuditLegEnv+"=1") {
		t.Fatalf("audit-leg child environment lacks %s=1", codexAuditLegEnv)
	}
}

// Building a leg child environment must not mark the spawning process (condition a).
func TestCodexLegMarkerDoesNotLeakToParent(t *testing.T) {
	t.Setenv(codexAuditLegEnv, "")
	_ = codexLegChildEnv(withCodexAuditLeg(context.Background()))
	if got := os.Getenv(codexAuditLegEnv); got != "" {
		t.Fatalf("building a leg child environment changed the parent: %s=%q", codexAuditLegEnv, got)
	}
}

// The codex child command carries the marker only when a leg spawns it.
func TestCodexChildCmdCarriesMarkerOnlyInLeg(t *testing.T) {
	ordinary := codexChildCmd(context.Background(), "codex", []string{"app-server"})
	// t1869: the ordinary child carries the parent environment minus the Orca
	// pane-owned keys, so Env is set; what stays true is that it has no marker.
	if slices.Contains(ordinary.Env, codexAuditLegEnv+"=1") {
		t.Fatalf("ordinary codex child must not carry %s=1", codexAuditLegEnv)
	}
	leg := codexChildCmd(withCodexAuditLeg(context.Background()), "codex", []string{"app-server"})
	if !slices.Contains(leg.Env, codexAuditLegEnv+"=1") {
		t.Fatalf("audit-leg codex child lacks %s=1", codexAuditLegEnv)
	}
}

func TestCodexAuditLegMarkerIsReadFromEnv(t *testing.T) {
	t.Setenv(codexAuditLegEnv, "")
	if inCodexAuditLeg() {
		t.Fatal("an unset marker must not count as an audit leg")
	}
	t.Setenv(codexAuditLegEnv, "1")
	if !inCodexAuditLeg() {
		t.Fatal("the marker set to 1 must count as an audit leg")
	}
}

// One subtest per verify-demanding member (t1788 brain ruling). Inside an audit
// leg each is skipped, and the skip is visible rather than silent (condition b).
func TestAuditLegSkipsEachVerifyMember(t *testing.T) {
	t.Setenv(codexAuditLegEnv, "1")
	for n, name := range codexAuditLegSkipped {
		t.Run(name, func(t *testing.T) {
			if !auditLegSkipsMember(n) {
				t.Fatalf("member %d (%s) must be skipped inside an audit leg", n, name)
			}
			o := auditLegSkipOutcome(n)
			if o.Decision != codexadapter.DecisionAllow {
				t.Fatalf("a skipped member must not block: decision=%v", o.Decision)
			}
			if o.Status != stopStatusSkippedAuditLeg {
				t.Fatalf("status = %q, want %q", o.Status, stopStatusSkippedAuditLeg)
			}
			if !strings.Contains(o.Advisory, "skipped: audit leg") {
				t.Fatalf("the skip must be visible in the Stop output: advisory=%q", o.Advisory)
			}
			if len(o.Discards) != 1 {
				t.Fatalf("a skip must leave one discard record, got %d", len(o.Discards))
			}
		})
	}
}

// Per member, through the real budgeted path: inside a leg the member never runs.
func TestAuditLegWiringSkipsBeforeMemberRuns(t *testing.T) {
	t.Setenv(codexAuditLegEnv, "1")
	for n, name := range codexAuditLegSkipped {
		t.Run(name, func(t *testing.T) {
			var called bool
			fn := func(context.Context) stopMemberOutcome {
				called = true
				return stopMemberOutcome{Decision: codexadapter.DecisionDeny, Status: "ran"}
			}
			cutOff := func(context.Context) stopMemberOutcome {
				return stopMemberOutcome{Decision: codexadapter.DecisionDeny, Status: "cut-off"}
			}
			c := &codexStopChain{}
			o := c.budgeted(context.Background(), n, fn, cutOff)
			if called {
				t.Fatalf("member %d (%s) ran inside an audit leg", n, name)
			}
			if o.Status != stopStatusSkippedAuditLeg || o.Decision != codexadapter.DecisionAllow {
				t.Fatalf("budgeted returned status=%q decision=%v, want the skip outcome", o.Status, o.Decision)
			}
		})
	}
}

// Outside a leg the same path runs the member, so no ordinary gate is skipped.
func TestAuditLegWiringRunsMemberOutsideLeg(t *testing.T) {
	t.Setenv(codexAuditLegEnv, "")
	for n, name := range codexAuditLegSkipped {
		t.Run(name, func(t *testing.T) {
			var called bool
			fn := func(context.Context) stopMemberOutcome {
				called = true
				return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: "ran"}
			}
			cutOff := func(context.Context) stopMemberOutcome {
				return stopMemberOutcome{Decision: codexadapter.DecisionAllow, Status: "cut-off"}
			}
			c := &codexStopChain{budgetFor: func(int) time.Duration { return 5 * time.Second }}
			o := c.budgeted(context.Background(), n, fn, cutOff)
			if !called {
				t.Fatalf("member %d (%s) must run outside an audit leg", n, name)
			}
			if o.Status != "ran" {
				t.Fatalf("budgeted returned status=%q, want the member's own outcome", o.Status)
			}
		})
	}
}

// Outside an audit leg nothing is skipped, so the marker does not change ordinary gates.
func TestAuditLegSkipsNothingOutsideLeg(t *testing.T) {
	t.Setenv(codexAuditLegEnv, "")
	for n, name := range codexAuditLegSkipped {
		if auditLegSkipsMember(n) {
			t.Fatalf("member %d (%s) skipped outside an audit leg", n, name)
		}
	}
}

// The set is defined by property and enumerated (t1788): sync-gate, goal, and
// codex-review ask the model for a verification command. The multi-review member
// reports only a residual-risk note, so it is not in the set.
func TestAuditLegSkipSetEnumeratesVerifyMembers(t *testing.T) {
	want := map[int]string{2: "sync-gate", 3: "goal", 6: "codex-review"}
	if len(codexAuditLegSkipped) != len(want) {
		t.Fatalf("skip set = %v, want %v", codexAuditLegSkipped, want)
	}
	for n, name := range want {
		if got := codexAuditLegSkipped[n]; got != name {
			t.Fatalf("skip set[%d] = %q, want %q", n, got, name)
		}
	}
}
