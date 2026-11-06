package effect

import (
	"context"
	"errors"
	"testing"
)

// Approve is Check that also hands back proof: a context that carries the approval of exactly this method
// and path, and only when Check passed. Only this package can put the proof there.
func TestApproveReturnsProofOfExactlyTheCheckedWriteAndOnlyWhenItPassed(t *testing.T) {
	e := &Enforcer{Identity: world()}
	bg := context.Background()

	ctx, err := e.Approve(bg, "POST", "/repos/o/r/issues/7/comments", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !Approved(ctx, "POST", "/repos/o/r/issues/7/comments") || !Approved(ctx, "POST", "repos/o/r/issues/7/comments") {
		t.Error("the approved write is not recognised (with or without the leading slash)")
	}
	for name, tc := range map[string][2]string{
		"another issue": {"POST", "/repos/o/r/issues/8/comments"},
		"another verb":  {"PATCH", "/repos/o/r/issues/7/comments"},
		"another repo":  {"POST", "/repos/x/y/issues/7/comments"},
		"a read":        {"GET", "/repos/o/r/issues/7/comments"},
	} {
		if Approved(ctx, tc[0], tc[1]) {
			t.Errorf("%s: the approval was taken for a different write", name)
		}
	}
	if Approved(bg, "POST", "/repos/o/r/issues/7/comments") {
		t.Error("a context that was never approved reads as approved")
	}

	// a refusal returns no proof
	refused, err := e.Approve(bg, "PATCH", "/repos/o/r/issues/comments/502", nil, nil) // their comment
	var re *RefusedError
	if !errors.As(err, &re) || Approved(refused, "PATCH", "/repos/o/r/issues/comments/502") {
		t.Errorf("err=%v approved=%v: a refused write must not come back approved", err, Approved(refused, "PATCH", "/repos/o/r/issues/comments/502"))
	}
	// nothing wired: fail closed
	var none *Enforcer
	if ctx, err := none.Approve(bg, "POST", "/repos/o/r/issues/7/comments", nil, nil); err == nil || Approved(ctx, "POST", "/repos/o/r/issues/7/comments") {
		t.Errorf("no enforcer: err=%v", err)
	}
}
