package e2e

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

// AC08 on lane-20's weakening-refusal guard (REQ-AF-019): every write to
// protection settings is refused by method and path before anything is sent.
// The guard never looks at a body, a command wording or an error message, so
// "turn a setting on" and "turn it off" are the same request to it and are both
// refused. No request here leaves the process: CheckRequest only decides.

func refusal(t *testing.T, err error) *effect.RefusedError {
	t.Helper()
	if !errors.Is(err, effect.ErrRefused) {
		t.Fatalf("want a refusal, got %v", err)
	}
	var re *effect.RefusedError
	if !errors.As(err, &re) {
		t.Fatalf("refusal %v carries no RefusedError", err)
	}
	return re
}

// The brief's seven weakening operations, as the REST requests that perform them.
func TestGuard_SevenWeakeningOperations_EachRefusedWithItsReason(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		reason effect.RefusalReason
	}{
		{"delete required status checks", "DELETE", "/repos/o/r/branches/main/protection/required_status_checks", effect.ReasonWeakensProtection},
		{"turn enforce_admins off", "DELETE", "/repos/o/r/branches/main/protection/enforce_admins", effect.ReasonWeakensProtection},
		{"allow force pushes (protection update)", "PUT", "/repos/o/r/branches/main/protection", effect.ReasonWeakensProtection},
		{"allow deletions (protection update)", "PUT", "/repos/o/r/branches/main/protection", effect.ReasonWeakensProtection},
		{"change required_linear_history to true", "PUT", "/repos/o/r/branches/main/protection", effect.ReasonWeakensProtection},
		{"change required_linear_history to false", "PUT", "/repos/o/r/branches/main/protection", effect.ReasonWeakensProtection},
		{"force-push main (update the ref)", "PATCH", "/repos/o/r/git/refs/heads/main", effect.ReasonNotAllowlisted},
		{"admin-bypass merge", "PUT", "/repos/o/r/pulls/1/merge", effect.ReasonNotAllowlisted},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			re := refusal(t, effect.CheckRequest(tc.method, tc.path, nil))
			if re.Reason != tc.reason {
				t.Errorf("reason = %s, want %s", re.Reason, tc.reason)
			}
			if re.Method != tc.method || re.Path != tc.path {
				t.Errorf("refusal records %s %s, want %s %s", re.Method, re.Path, tc.method, tc.path)
			}
		})
	}
}

// Other settings areas are refused the same way: rulesets, hooks, collaborators, keys, secrets.
func TestGuard_OtherSettingsAreas_WeakenProtection(t *testing.T) {
	for _, tc := range []struct{ method, path string }{
		{"POST", "/repos/o/r/rulesets"},
		{"PUT", "/repos/o/r/rulesets/7"},
		{"POST", "/repos/o/r/hooks"},
		{"PUT", "/repos/o/r/collaborators/someone"},
		{"POST", "/repos/o/r/keys"},
		{"PUT", "/repos/o/r/actions/secrets/NAME"},
	} {
		if got := refusal(t, effect.CheckRequest(tc.method, tc.path, nil)).Reason; got != effect.ReasonWeakensProtection {
			t.Errorf("%s %s -> %s, want weakens-protection", tc.method, tc.path, got)
		}
	}
}

// Control: the writes the factory is meant to send pass, and so do reads, even
// reads of protection settings (reading weakens nothing). A guard that refuses
// everything would fail here.
func TestGuard_AllowedWritesAndReads_Pass(t *testing.T) {
	for _, tc := range []struct{ method, path string }{
		{"POST", "/repos/o/r/issues"},
		{"POST", "/repos/o/r/labels"},
		{"PATCH", "/repos/o/r/labels/kind%2Fbug"}, // an encoded slash stays inside one label segment
		{"GET", "/repos/o/r/branches/main/protection"},
		{"GET", "/repos/o/r/issues/1948/comments"},
		{"HEAD", "/repos/o/r/rulesets"},
	} {
		if err := effect.CheckRequest(tc.method, tc.path, nil); err != nil {
			t.Errorf("%s %s refused: %v", tc.method, tc.path, err)
		}
	}
}

// Label deletion is deliberately not on the allowlist.
func TestGuard_LabelDelete_NotAllowlisted(t *testing.T) {
	if got := refusal(t, effect.CheckRequest("DELETE", "/repos/o/r/labels/p0", nil)).Reason; got != effect.ReasonNotAllowlisted {
		t.Errorf("DELETE label -> %s, want not-allowlisted", got)
	}
}

// A request in a shape that could hide a destination is refused as malformed,
// not interpreted: dot segments, empty segments, an absolute URL, control
// characters, a lower-case method and a method-override header.
func TestGuard_MalformedRequests_AreRefused(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		header map[string]string
	}{
		{"dot-dot segment to protection", "PATCH", "/repos/o/r/labels/../branches/main/protection", nil},
		{"encoded dot-dot", "PATCH", "/repos/o/r/labels/%2e%2e/branches/main/protection", nil},
		{"empty segment", "POST", "/repos/o/r//issues", nil},
		{"trailing slash", "POST", "/repos/o/r/issues/", nil},
		{"absolute URL", "POST", "https://api.example.invalid/repos/o/r/issues", nil},
		{"control character", "POST", "/repos/o/r/issues\x00", nil},
		{"lower-case method", "post", "/repos/o/r/issues", nil},
		{"method override header", "POST", "/repos/o/r/issues", map[string]string{"X-HTTP-Method-Override": "DELETE"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := refusal(t, effect.CheckRequest(tc.method, tc.path, tc.header)).Reason; got != effect.ReasonMalformed {
				t.Errorf("reason = %s, want malformed-request", got)
			}
		})
	}
}

// The wording of a path segment alone does not turn an allowed write into a
// refusal: a label whose NAME contains "protection" is one segment of a label
// write and passes (segments are compared whole).
func TestGuard_LabelNamedProtection_IsNotMistakenForTheSetting(t *testing.T) {
	if err := effect.CheckRequest("PATCH", "/repos/o/r/labels/protection", nil); err != nil {
		t.Errorf("a label named protection was refused: %v", err)
	}
}

// Since lane-20 797f4f8174 comment and Issue writes are no longer on the plain
// allowlist: the static guard refuses them as identity-unknown, and they pass only
// through Enforcer.Check, which establishes who is writing (real_enforcer_test.go).
func TestGuard_CommentAndIssueWrites_NeedTheIdentityCheck(t *testing.T) {
	for _, tc := range []struct{ method, path string }{
		{"POST", "/repos/o/r/issues/1948/comments"},
		{"PATCH", "/repos/o/r/issues/comments/501"},
		{"PATCH", "/repos/o/r/issues/1948"},
	} {
		if got := refusal(t, effect.CheckRequest(tc.method, tc.path, nil)).Reason; got != effect.ReasonIdentityUnknown {
			t.Errorf("%s %s -> %s, want identity-unknown (decided only by Enforcer.Check)", tc.method, tc.path, got)
		}
	}
}
