package e2e

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/checkpoint"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect/ctxledger"
)

// Cost of the four checkpoint stages, counted against fakes (no real GitHub, no real
// git remote). Stages 1 and 3 are local store work and make no API call; stage 2 is
// remote lookups (a Head and a Contains); stage 4 is the only one that posts a comment,
// counted by lane-20's CountingRemote. Elapsed time and tokens are NOT measured here.

func cpCheckpoint() ctxledger.Checkpoint {
	d := ctxledger.Durable{SHA: cpSHA, RemoteVerified: true}
	return ctxledger.Checkpoint{
		Issue: 1948, Commit: d, ACK: d,
		Manifest:     ctxledger.Manifest{Path: ".moai/specs/SPEC-X/manifest.json", Hash: "sha256-abc", Durable: d},
		LinkTemplate: "https://example.test/org/repo/blob/{sha}/{path}",
	}
}

func countedLedger() (*ctxledger.Ledger, *ctxledger.CountingRemote, *ctxRemote) {
	inner := &ctxRemote{}
	cr := &ctxledger.CountingRemote{Inner: inner}
	return newCtxLedger(cr, effect.NewMemIntentStore()), cr, inner
}

func TestMeasureCheckpoint_FourStages_APICalls(t *testing.T) {
	ctx := context.Background()
	s, _ := cpSaver(t)

	// Stage 1 (commit) and stage 3 (manifest): store only, no remote involved.
	r := checkpoint.NewRecord(cpManifest())
	if err := s.Save(&r); err != nil {
		t.Fatal(err)
	}
	r.MarkPushSent()
	// Stage 2 (remote ACK): lookups only.
	rem := &cpRemote{found: true, contains: true}
	if err := r.ConfirmRemote(rem); err != nil {
		t.Fatal(err)
	}
	if err := s.Verify(&r); err != nil {
		t.Fatal(err)
	}
	t.Logf("MEASURED stage2 remote lookups: head=%d contains=%d (git plumbing, 0 API calls); stages 1+3: 0 remote calls (local store)", rem.heads, rem.checks)

	// Stage 4 (summary link): first publish.
	l, cr, inner := countedLedger()
	res := l.PublishSummaryLink(ctx, cpCheckpoint())
	if res.Status != ctxledger.StepPublished {
		t.Fatalf("first publish: %+v", res)
	}
	first := cr.Snapshot()
	t.Logf("MEASURED stage4 first publish: %+v total=%d comments=%d", first, first.Total(), len(inner.comments))
	if len(inner.comments) != 1 {
		t.Errorf("comments = %d, want 1", len(inner.comments))
	}

	// Replay of the same checkpoint: no second comment.
	cr.Reset()
	if res := l.PublishSummaryLink(ctx, cpCheckpoint()); res.Status != ctxledger.StepPublished {
		t.Fatalf("replay: %+v", res)
	}
	replay := cr.Snapshot()
	t.Logf("MEASURED stage4 replay of the same checkpoint: %+v total=%d comments=%d", replay, replay.Total(), len(inner.comments))
	if replay.Create != 0 || len(inner.comments) != 1 {
		t.Errorf("replay created again: %+v comments=%d", replay, len(inner.comments))
	}

	// Not durable yet: nothing is sent at all.
	l2, cr2, inner2 := countedLedger()
	notYet := cpCheckpoint()
	notYet.ACK.RemoteVerified = false
	if res := l2.PublishSummaryLink(ctx, notYet); res.Status != ctxledger.StepSkipped {
		t.Fatalf("not durable: %+v", res)
	}
	if cr2.Snapshot().Total() != 0 || len(inner2.comments) != 0 {
		t.Errorf("a skipped step made calls: %+v", cr2.Snapshot())
	}
	t.Logf("MEASURED stage4 skipped (an earlier step not remote-confirmed): total=%d", cr2.Snapshot().Total())
}

// A lost create response and a comment failure both leave exactly one comment at most,
// and a failed summary never changes the recovery verdict decided by the earlier stages.
func TestMeasureCheckpoint_Stage4_LostResponseAndFailure(t *testing.T) {
	ctx := context.Background()

	l, cr, inner := countedLedger()
	inner.loseNextCreate = true
	first := l.PublishSummaryLink(ctx, cpCheckpoint())
	afterLost := cr.Snapshot()
	second := l.PublishSummaryLink(ctx, cpCheckpoint())
	total := cr.Snapshot()
	t.Logf("MEASURED stage4 lost create response: first=%s calls=%+v; retry=%s total=%+v comments=%d",
		first.Status, afterLost, second.Status, total, len(inner.comments))
	if len(inner.comments) != 1 {
		t.Errorf("duplicate comments after a lost response: %d, want 1", len(inner.comments))
	}

	l2, cr2, inner2 := countedLedger()
	inner2.createErr = httpStatus(502, "bad gateway")
	failed := l2.PublishSummaryLink(ctx, cpCheckpoint())
	if failed.Status != ctxledger.StepFailed || failed.Err == nil {
		t.Fatalf("failing comment: %+v", failed)
	}
	t.Logf("MEASURED stage4 comment failure (502): status=%s calls=%+v comments=%d", failed.Status, cr2.Snapshot(), len(inner2.comments))
	if len(inner2.comments) != 0 {
		t.Errorf("comments = %d after a failure, want 0", len(inner2.comments))
	}

	// The step reports failure by value; the earlier stages' verdict is untouched.
	r, _ := durable(t)
	r.RecordSummary(failed.Err)
	if err := r.CanRecover(okFacts); err != nil {
		t.Errorf("a failed summary closed recovery: %v", err)
	}
	if errors.Is(failed.Err, checkpoint.ErrNotDurable) {
		t.Error("summary failure surfaced as a durability failure")
	}
}
