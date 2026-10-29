package effect

import (
	"errors"
	"testing"
)

func refusal(t *testing.T, method, path string) *RefusedError {
	t.Helper()
	err := CheckRequest(method, path, nil)
	var re *RefusedError
	if !errors.As(err, &re) {
		t.Fatalf("%s %s: err = %v, want a refusal", method, path, err)
	}
	if !errors.Is(err, ErrRefused) {
		t.Fatalf("%s %s: refusal does not match ErrRefused", method, path)
	}
	return re
}

func TestReadsAndTheWritesWeSendAreAllowed(t *testing.T) {
	for _, c := range [][2]string{
		{"GET", "/repos/o/r/branches/main/protection"}, // a read of a protected thing is a read
		{"GET", "/repos/o/r/rulesets"},
		{"HEAD", "/repos/o/r/hooks"},
		{"POST", "/repos/o/r/issues"},
		{"POST", "/repos/o/r/labels"},
		{"PATCH", "/repos/o/r/labels/p0"},
		{"PATCH", "/repos/o/r/labels/type%3Abug"},
		{"PATCH", "/repos/o/r/labels/P0"},
		{"PATCH", "/repos/o/r/labels/kind%2Fbug"},
		{"PATCH", "/repos/o/r/labels/area%2Fhooks"},
		{"POST", "/repos/o/r/labels?x=1"},
	} {
		if err := CheckRequest(c[0], c[1], nil); err != nil {
			t.Errorf("%s %s refused: %v", c[0], c[1], err)
		}
	}
}

func TestOperationsThatWeakenProtectionAreRefusedByMethodAndPath(t *testing.T) {
	for _, c := range [][2]string{
		{"PUT", "/repos/o/r/branches/main/protection"},
		{"DELETE", "/repos/o/r/branches/main/protection"},
		{"PATCH", "/repos/o/r/branches/main/protection/required_status_checks"},
		{"DELETE", "/repos/o/r/branches/main/protection/enforce_admins"},
		{"POST", "/repos/o/r/rulesets"},
		{"PUT", "/repos/o/r/rulesets/12"},
		{"DELETE", "/repos/o/r/rulesets/12"},
		{"POST", "/repos/o/r/hooks"},
		{"PATCH", "/repos/o/r/hooks/3"},
		{"PUT", "/repos/o/r/collaborators/someone"},
		{"PUT", "/repos/o/r/actions/permissions"},
		{"POST", "/repos/o/r/keys"},
		{"PUT", "/projects/1/protected_branches/main"},
		{"POST", "/projects/1/hooks"},
		{"POST", "/projects/1/approval_rules"},
	} {
		if re := refusal(t, c[0], c[1]); re.Reason != ReasonWeakensProtection {
			t.Errorf("%s %s: reason = %s, want %s", c[0], c[1], re.Reason, ReasonWeakensProtection)
		}
	}
}

func TestAnythingNotOnTheAllowlistIsRefusedEvenWhenItLooksHarmless(t *testing.T) {
	for _, c := range [][2]string{
		{"POST", "/graphql"}, // a mutation cannot be inspected: refuse the endpoint
		{"DELETE", "/repos/o/r/labels/p0"},
		{"DELETE", "/repos/o/r/issues/comments/9"},
		{"PUT", "/repos/o/r/issues/7/lock"},
		{"POST", "/repos/o/r/pulls/7/merge"},
		{"PATCH", "/repos/o/r"},
	} {
		if re := refusal(t, c[0], c[1]); re.Reason != ReasonNotAllowlisted && re.Reason != ReasonWeakensProtection {
			t.Errorf("%s %s: reason = %q", c[0], c[1], re.Reason)
		}
	}
}

func TestPathTricksDoNotSlipAWeakeningWriteThrough(t *testing.T) {
	for _, c := range [][2]string{
		{"POST", "/repos/o/r/issues/../branches/main/protection"},
		{"POST", "/repos/o/r/labels/../hooks"},
		{"POST", "/repos/o/r//hooks"},
		{"POST", "/repos/o/r/hooks/"},
		{"post", "/repos/o/r/hooks"},
		{"PUT", "/REPOS/O/R/BRANCHES/MAIN/PROTECTION"},
		{"POST", "/repos/o/r/%68ooks"},
		{"POST", "/repos/o/r/labels/%2e%2e/hooks"},
		{"POST", "/repos/o/r/labels%2f..%2fhooks"},
		{"POST", "/repos/o/r/issues\x00/x"},
		{"POST", "https://api.github.com/repos/o/r/hooks"},
		{"POST", ""},
	} {
		if err := CheckRequest(c[0], c[1], nil); err == nil {
			t.Errorf("%s %q slipped through", c[0], c[1])
		}
	}
}

// The match is structural: a label that happens to be called "hooks" is a label, and a hook is a hook.
func TestMatchingIsBySegmentNotBySubstring(t *testing.T) {
	if err := CheckRequest("PATCH", "/repos/o/r/labels/hooks", nil); err != nil {
		t.Errorf("a label named hooks was refused: %v", err)
	}
	if err := CheckRequest("PATCH", "/repos/o/r/labels/protection", nil); err != nil {
		t.Errorf("a label named protection was refused: %v", err)
	}
	refusal(t, "PATCH", "/repos/o/r/hooks")
}

func TestMethodOverrideHeadersAreRefused(t *testing.T) {
	for _, h := range []map[string]string{
		{"X-HTTP-Method-Override": "DELETE"},
		{"x-http-method-override": "PUT"},
		{"X-Method-Override": "PATCH"},
	} {
		if err := CheckRequest("POST", "/repos/o/r/labels", h); !errors.Is(err, ErrRefused) {
			t.Errorf("%v: err = %v", h, err)
		}
	}
}

func TestRefusalParksTheIntentWithoutLookupOrResend(t *testing.T) {
	refused := CheckRequest("PUT", "/repos/o/r/branches/main/protection", nil)
	s := &fakeSender{err: refused}
	lk := &fakeLookup{}
	ex := newExec(s, lk, NewMemIntentStore())
	in, _ := ex.Execute(t.Context(), req())
	if in.Status != StatusParked || lk.calls != 0 || s.calls != 1 {
		t.Fatalf("in=%+v lookups=%d sends=%d", in, lk.calls, s.calls)
	}
	if in.ParkReason != string(ReasonWeakensProtection) {
		t.Errorf("park reason = %q", in.ParkReason)
	}
	if again, _ := ex.Execute(t.Context(), req()); again.Status != StatusParked || s.calls != 1 {
		t.Errorf("a refused operation was sent again: calls=%d", s.calls)
	}
}
