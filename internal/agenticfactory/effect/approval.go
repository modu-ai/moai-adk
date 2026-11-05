package effect

import (
	"context"
	"strings"
)

// A write that passed the Issue-author policy carries proof of it, so that a layer further down (the
// process wrapper) can tell "the policy approved exactly this" from "someone calls it directly". The proof
// lives in a context value of an unexported type: only Enforcer.Approve can put it there, and it names one
// method and one path.

type approvalKey struct{}

type approval struct{ method, path string }

func approvedPath(rawPath string) string { return strings.TrimPrefix(rawPath, "/") }

// Approve is Check that also returns proof of the check. When Check refuses, the returned context carries
// no proof and the error is the refusal. With no Enforcer (or no identity provider) Check refuses, so
// nothing is approved without one.
func (e *Enforcer) Approve(ctx context.Context, method, rawPath string, header map[string]string, body []byte) (context.Context, error) {
	if err := e.Check(ctx, method, rawPath, header, body); err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, approvalKey{}, approval{method: method, path: approvedPath(rawPath)}), nil
}

// Approved reports whether ctx carries the approval of exactly this method and path (with or without the
// leading slash).
func Approved(ctx context.Context, method, rawPath string) bool {
	a, ok := ctx.Value(approvalKey{}).(approval)
	return ok && a.method == method && a.path == approvedPath(rawPath)
}
