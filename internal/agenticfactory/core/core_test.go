package core

import (
	"errors"
	"strings"
	"testing"
)

func TestStateValid(t *testing.T) {
	for _, s := range AllStates() {
		if !s.Valid() {
			t.Errorf("state %q should be valid", s)
		}
	}
	for _, bad := range []State{"", "Queued", "running", "done "} {
		if bad.Valid() {
			t.Errorf("state %q should be invalid", bad)
		}
	}
	if got := len(AllStates()); got != 9 {
		t.Errorf("want 9 states, got %d", got)
	}
}

func TestStateClassification(t *testing.T) {
	if StateBlocked.IsProgress() || StateOnHold.IsProgress() {
		t.Error("blocked/on-hold are not progress stages")
	}
	if !StateImplementing.IsProgress() {
		t.Error("implementing is a progress stage")
	}
	if !StateBlocked.IsPause() || !StateOnHold.IsPause() || StateDone.IsPause() {
		t.Error("pause classification wrong")
	}
	if !StateDone.IsTerminal() || StateLanding.IsTerminal() {
		t.Error("terminal classification wrong")
	}
}

func TestBlockReasonScope(t *testing.T) {
	if err := ValidatePause(StateBlocked, ReasonOperatorHold); !errors.Is(err, ErrReasonNotAllowed) {
		t.Errorf("operator-hold on blocked must be rejected, got %v", err)
	}
	if err := ValidatePause(StateOnHold, ReasonOperatorHold); err != nil {
		t.Errorf("operator-hold on on-hold must pass: %v", err)
	}
	if err := ValidatePause(StateBlocked, ReasonRateLimited); err != nil {
		t.Errorf("rate-limited on blocked must pass: %v", err)
	}
	if err := ValidatePause(StateBlocked, BlockReason("made-up")); !errors.Is(err, ErrReasonUnknown) {
		t.Errorf("unknown reason must be rejected, got %v", err)
	}
	if err := ValidatePause(StateQueued, ReasonRateLimited); !errors.Is(err, ErrNotPauseState) {
		t.Errorf("non-pause state must be rejected, got %v", err)
	}
}

func TestDeliveryOrder(t *testing.T) {
	order := []Delivery{DeliverySent, DeliveryReceived, DeliveryExecuted, DeliveryCompleted}
	for i, d := range order {
		if !d.Valid() {
			t.Errorf("%q invalid", d)
		}
		if i > 0 && !d.AtLeast(order[i-1]) {
			t.Errorf("%q should be at least %q", d, order[i-1])
		}
		if i > 0 && order[i-1].AtLeast(d) {
			t.Errorf("%q must not be at least %q", order[i-1], d)
		}
	}
	if Delivery("x").AtLeast(DeliverySent) {
		t.Error("invalid delivery must not satisfy AtLeast")
	}
}

func TestHeadSHA(t *testing.T) {
	good := strings.Repeat("a1", 20)
	if _, err := ParseHeadSHA(good); err != nil {
		t.Errorf("40 hex must parse: %v", err)
	}
	for _, bad := range []string{"", "cb30e05b3e", strings.Repeat("A1", 20), strings.Repeat("g", 40), good + "0"} {
		if _, err := ParseHeadSHA(bad); !errors.Is(err, ErrHeadInvalid) {
			t.Errorf("%q must be rejected with ErrHeadInvalid, got %v", bad, err)
		}
	}
}

func TestIdentifiers(t *testing.T) {
	for _, ok := range []string{"run-1", "a", "r1-attempt-02"} {
		if _, err := ParseRunID(ok); err != nil {
			t.Errorf("run id %q: %v", ok, err)
		}
		if _, err := ParseAttemptID(ok); err != nil {
			t.Errorf("attempt id %q: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "Run", "a/b", "../x", "a b", strings.Repeat("a", 65)} {
		if _, err := ParseRunID(bad); !errors.Is(err, ErrIDInvalid) {
			t.Errorf("run id %q must be rejected, got %v", bad, err)
		}
		if _, err := ParseAttemptID(bad); !errors.Is(err, ErrIDInvalid) {
			t.Errorf("attempt id %q must be rejected, got %v", bad, err)
		}
	}
}

func TestIssueRefAndKey(t *testing.T) {
	if _, err := NewIssueRef(0); !errors.Is(err, ErrIssueInvalid) {
		t.Errorf("issue 0 must be rejected, got %v", err)
	}
	ref, err := NewIssueRef(1948)
	if err != nil {
		t.Fatal(err)
	}
	if ref.Resource() != "issue-1948" {
		t.Errorf("resource = %q", ref.Resource())
	}
}

func TestErrorSplit(t *testing.T) {
	if errors.Is(ErrHeld, ErrBusy) || errors.Is(ErrBusy, ErrHeld) {
		t.Error("Held and Busy must stay distinct")
	}
	for _, e := range []error{ErrHeld, ErrBusy, ErrNotHeld, ErrForeign, ErrStaleHead, ErrDuplicateDispatch, ErrOwnershipUnverified} {
		if e == nil || e.Error() == "" {
			t.Errorf("sentinel missing text: %v", e)
		}
	}
}

func TestOwnershipUnverifiedMapsToBlockReason(t *testing.T) {
	r, ok := ReasonFor(ErrOwnershipUnverified)
	if !ok || r != ReasonOwnershipUnverified {
		t.Errorf("got %q ok=%v", r, ok)
	}
	if _, ok := ReasonFor(ErrBusy); ok {
		t.Error("Busy is a retry, not a block reason")
	}
}

func TestAgentIdentity(t *testing.T) {
	a := Agent{SessionID: "11111111-2222-3333-4444-555555555555", Number: 3}
	if a.DisplayName() != "agent-3" || a.DisplayName() == a.SessionID {
		t.Errorf("display name = %q", a.DisplayName())
	}
	if _, err := ParseSessionID(a.SessionID); err != nil {
		t.Errorf("uuid must parse: %v", err)
	}
	for _, bad := range []string{"", "short", "a/b/c/d/e/f/g/h", "has space 12345678", strings.Repeat("a", 65)} {
		if _, err := ParseSessionID(bad); !errors.Is(err, ErrIDInvalid) {
			t.Errorf("%q must be rejected, got %v", bad, err)
		}
	}
}

func TestBacklogFullMapsToCapacityShort(t *testing.T) {
	if r, ok := ReasonFor(ErrBacklogFull); !ok || r != ReasonCapacityShort {
		t.Errorf("got %q ok=%v", r, ok)
	}
	if r, ok := ReasonFor(ErrStoreUnavailable); !ok || r != ReasonStoreUnavailable {
		t.Errorf("got %q ok=%v", r, ok)
	}
}

func TestRedactMasksSecretsAndCounts(t *testing.T) {
	in := "use ghp_abcdefghijklmnopqrstuvwxyz0123 and Bearer abcdefgh12345678 with api_key=zzz999 ok"
	out, n := Redact(in)
	if n != 3 || strings.Contains(out, "ghp_") || strings.Contains(out, "zzz999") || strings.Contains(out, "abcdefgh12345678") {
		t.Errorf("redact: %q n=%d", out, n)
	}
	if clean, n := Redact("nothing to hide"); clean != "nothing to hide" || n != 0 {
		t.Errorf("clean text changed: %q %d", clean, n)
	}
}

func TestIntakeErrorsMapToCapacity(t *testing.T) {
	for _, e := range []error{ErrIntakeFull, ErrNoSlot, ErrBacklogFull} {
		if r, ok := ReasonFor(e); !ok || r != ReasonCapacityShort {
			t.Errorf("%v -> %q %v", e, r, ok)
		}
	}
}

func TestRedactMasksEmailsAndAbsoluteHomePaths(t *testing.T) {
	cases := []struct{ name, in string }{
		{"email", "ask dev.person+x@example.co.kr about it"},
		{"ssh remote", "git@github.com:org/repo.git"},
		{"mac home", "see /Users/goos/MoAI/moai-adk-go/internal/x.go for details"},
		{"linux home", "log at /home/runner/work/_temp/out.log"},
		{"windows home", `file C:\Users\Goos\AppData\x.txt here`},
		{"bare home dir", "cwd=/Users/goos"},
		{"quoted", `{"path":"/Users/goos/secret/file.txt"}`},
	}
	for _, c := range cases {
		out, n := Redact(c.in)
		if n == 0 {
			t.Errorf("%s: nothing masked in %q", c.name, c.in)
		}
		for _, leak := range []string{"example.co.kr", "dev.person", "github.com:org", "goos", "Goos", "runner", "AppData"} {
			if strings.Contains(out, leak) && strings.Contains(c.in, leak) {
				t.Errorf("%s: %q still leaks in %q", c.name, leak, out)
			}
		}
	}
}

func TestRedactLeavesOrdinaryTextAlone(t *testing.T) {
	for _, in := range []string{
		"internal/agenticfactory/claim/claim.go:42",
		"/usr/local/bin/moai and /tmp/work",
		"@MX:NOTE reviewed by the team",
		"user@host without a domain suffix",
		"state=verifying head=cb30e05b3e",
		".moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md",
	} {
		if out, n := Redact(in); n != 0 || out != in {
			t.Errorf("over-redacted %q -> %q (%d)", in, out, n)
		}
	}
}
