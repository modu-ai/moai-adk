package effect

import (
	"context"
	"testing"
)

// AC-AF-008 / design §11.3: each protection-weakening operation W1..W7 is refused at the effect boundary
// before anything is sent. The send count is zero, no second record appears, and the recorded reason is
// exactly one: permission-denied. A refusal that an implementation does not make is a FAIL here, never a
// skip.

// guardedSender is the shape every real sender has: it asks the guard first and only then "sends".
type guardedSender struct {
	method, path string
	sends        int
}

func (g *guardedSender) Send(context.Context, Intent) (string, error) {
	if err := CheckRequest(g.method, g.path, nil); err != nil {
		return "", err
	}
	g.sends++
	return "sent", nil
}

func TestProtectionGuard_RefusesEveryWeakeningOperation(t *testing.T) {
	for name, w := range map[string]struct{ method, path string }{
		"W1 delete a required check":            {"DELETE", "/repos/o/r/branches/main/protection/required_status_checks"},
		"W2 weaken branch protection":           {"DELETE", "/repos/o/r/branches/main/protection/enforce_admins"},
		"W3 change the linear-history setting":  {"DELETE", "/repos/o/r/branches/main/protection/required_linear_history"},
		"W4 force-push main":                    {"PATCH", "/repos/o/r/git/refs/heads/main"},
		"W5 delete a protected branch":          {"DELETE", "/repos/o/r/git/refs/heads/main"},
		"W6 admin-bypass merge":                 {"PUT", "/repos/o/r/pulls/9/merge"},
		"W7 reach a rule through a label route": {"PATCH", "/repos/o/r/labels/x%2F..%2Fbranches%2Fmain%2Fprotection"},
	} {
		t.Run(name, func(t *testing.T) {
			gs := &guardedSender{method: w.method, path: w.path}
			mem := NewMemIntentStore()
			ex := &Executor{Store: mem, Send: gs, Lookup: notFoundLookup{}, Timeouts: casTimeouts()}
			req := Request{Kind: LabelApply, Target: "o/r/" + name, Digest: "d", Payload: `{}`}
			in, err := ex.Execute(t.Context(), req)
			if err != nil {
				t.Fatal(err)
			}
			if gs.sends != 0 {
				t.Errorf("sends = %d, want 0", gs.sends)
			}
			if in.Status != StatusParked || in.ParkReason != "permission-denied" {
				t.Errorf("recorded %s / %q, want parked / permission-denied", in.Status, in.ParkReason)
			}
			again, err := ex.Execute(t.Context(), req)
			if err != nil || again.Status != StatusParked || again.Attempts != in.Attempts {
				t.Errorf("repeat changed the record: %+v %v", again, err)
			}
			if all, _ := mem.List(); len(all) != 1 {
				t.Errorf("records = %d, want exactly one reason record", len(all))
			}
		})
	}
}
