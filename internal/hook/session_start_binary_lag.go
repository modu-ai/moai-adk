package hook

import (
	"context"
	"log/slog"
	"time"

	"github.com/modu-ai/moai-adk/internal/binlag"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// binaryLagJoinBound caps how long session start waits for the lag comparison
// before giving up on it for this session.
//
// The bound belongs to the CALLER, not to the comparison. A timeout placed
// inside the comparison would be replaced along with it whenever the seam is
// substituted, so it could not protect the handler from a seam that misbehaves
// — and a seam that misbehaves is the entire reason REQ-BLV-006 exists. The
// shape here is the one the deferred advisory scan already uses a few hundred
// lines up: a timer plus a select over a buffered channel.
//
// The value matches deferredScanJoinBound for the same reason it was chosen
// there: the comparison is two short git invocations that normally finish in
// tens of milliseconds, and a quarter-second caps the worst case a pathological
// repository can add to session start.
//
// @MX:NOTE: the join bound lives in the caller, never inside the seam
// @MX:SPEC: SPEC-BINARY-LAG-VISIBILITY-001
const binaryLagJoinBound = 250 * time.Millisecond

// binaryLagAdvisory returns the deployment-lag notice for this tree, or the
// empty string when there is nothing to say.
//
// Empty covers every non-lag outcome, and they are not distinguished here on
// purpose: a session start is not a diagnostic report, so the only outcome
// worth interrupting a reader for is the one where the binary really is
// running code the tree has moved past. Everything else — matching build,
// branch build, no repository to compare against, comparison too slow, panic
// inside the comparison — is silence.
//
// @MX:WARN @MX:REASON bounded-background-goroutine join — a comparison that
// overruns the bound is abandoned for this session; safe because the verdict
// is read-only and re-derives identically at the next session start.
func binaryLagAdvisory(ctx context.Context, dir string) string {
	if dir == "" {
		return ""
	}
	req := binlag.Request{
		Dir:           dir,
		BinaryCommit:  version.GetCommit(),
		BinaryVersion: version.GetVersion(),
	}

	if !deferredScansAsyncEnabled() {
		// Test path (TestMain sets deferredScansAsync=false): run inline so no
		// goroutine outlives the test boundary, matching the deferred-scan
		// branch a few hundred lines up and for the same reason.
		return binlag.Advisory(binlag.Evaluate(ctx, req))
	}

	// Buffered so a comparison that finishes after the bound has elapsed can
	// still send and exit rather than leaking blocked on an abandoned channel.
	result := make(chan string, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Debug("session start: binary lag comparison panicked (non-blocking)", "recover", r)
			}
		}()
		result <- binlag.Advisory(binlag.Evaluate(ctx, req))
	}()

	timer := time.NewTimer(binaryLagJoinBound)
	defer timer.Stop()
	select {
	case advisory := <-result:
		return advisory
	case <-timer.C:
		slog.Debug("session start: binary lag comparison exceeded join bound (non-blocking)",
			"bound", binaryLagJoinBound.String())
		return ""
	}
}

// appendAdditionalContext adds text to the session-start additionalContext,
// creating the hook-specific output block when this is the first contributor
// and separating from earlier contributors otherwise. Callers upstream open-code
// this same shape; the lag advisory is the last one to join, so it uses a
// helper rather than adding a fifth copy.
func appendAdditionalContext(out *HookOutput, text string) {
	if text == "" {
		return
	}
	if out.HookSpecificOutput == nil {
		out.HookSpecificOutput = &HookSpecificOutput{
			HookEventName: string(EventSessionStart),
		}
	}
	if out.HookSpecificOutput.AdditionalContext == "" {
		out.HookSpecificOutput.AdditionalContext = text
		return
	}
	out.HookSpecificOutput.AdditionalContext += "\n\n" + text
}
