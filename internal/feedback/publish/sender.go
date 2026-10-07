package publish

// sender.go — the publication sender (design.md section 7, REQ-ANON-015):
// it walks the user-scoped bugreport queue and files each item through the
// USER'S OWN gh.
//
// Discipline the tests pin:
//   - consent is re-read from the user-scoped file PER ITEM — a user
//     turning participation off mid-run stops the sender at the next item
//     (AC-003/AC-015);
//   - gh absent or unauthenticated leaves every item queued with no
//     prompt and no error (quiet);
//   - a hook dispatch is refused outright (the sentinel the hook command
//     sets);
//   - the whole run is bounded by the caller's context — the flush time
//     box in production;
//   - an existing fingerprint gets ONE occurrence comment below the cap,
//     nothing at the cap, never a body edit, never a label;
//   - the create path uses the deterministic template text (M5; M6's
//     model step slots in ahead of it).

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/bugreport"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/feedback"
	"github.com/modu-ai/moai-adk/internal/feedback/outbox"
)

// onHookPath reports the hook-dispatch sentinel the `moai hook` command
// sets for its process: the sender never publishes from inside a hook
// dispatch (REQ-ANON-015) — flush runs from the CLI's flush paths only.
func onHookPath() bool {
	return os.Getenv(config.EnvHookDispatch) == "1"
}

// The summary decision constants stored on the queue item (design.md
// section 8): which text the create path used — the model's or the
// deterministic template's.
const (
	summaryDecisionModel    = "model"
	summaryDecisionTemplate = "template"
)

// Sender files queued reports through one Runner. The Summarizer is
// INJECTED by internal/cli at the flush call (REQ-ANON-025); nil means the
// deterministic template only. The model-call budget is NOT sender state:
// it persists per user in the outbox store and is counted atomically
// cross-process (review-gate finding 3).
type Sender struct {
	Runner     Runner
	Summarizer Summarizer
}

// NewSender returns a sender over r.
func NewSender(r Runner) *Sender {
	return &Sender{Runner: r}
}

// FlushContext is the production entry the CLI's flush call drives: the
// send half of flush, over the real gh, under the caller's context — no
// summarizer (template only).
//
// @MX:ANCHOR: [AUTO] FlushContext — the CLI flush and the update-end trigger both enter the sender here
// @MX:REASON: a second sender entry could skip the hook-path refusal or the consent re-checks this walk enforces (REQ-ANON-015)
// @MX:WARN: [AUTO] every gh call is a public side effect from the user's account
// @MX:REASON: the sender's caps, consent checks, and refusal gates are the only things standing between a queue and a public post (REQ-ANON-015/016)
func FlushContext(ctx context.Context) error {
	return FlushContextWith(ctx, nil)
}

// FlushContextWith is FlushContext with the model seam INJECTED (DEC-6,
// REQ-ANON-025): internal/cli passes the production summarizer here, so
// publish imports no model helper and no internal/cli.
func FlushContextWith(ctx context.Context, s Summarizer) error {
	sender := NewSender(newExecRunner())
	sender.Summarizer = s
	return sender.Send(ctx)
}

// Flush is FlushContext under the flush time box (design section 10: the
// whole flush is bounded at config.DefaultBugreportFlushTimeBox).
func Flush() error {
	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultBugreportFlushTimeBox)
	defer cancel()
	return FlushContext(ctx)
}

// Send processes every queued item. Never fatal: publication trouble is
// logged to the outbox trail and the items stay queued — the command's own
// work (flush, update) must not fail because a public post did.
func (s *Sender) Send(ctx context.Context) error {
	if onHookPath() {
		slog.Debug("participation send refused: hook dispatch")
		return nil
	}

	store := outbox.BugreportQueueStore()
	rec, err := store.Load()
	if err != nil || len(rec.Items) == 0 {
		return nil // no queue or nothing to send: quiet
	}
	// Consent BEFORE any network work (review-gate finding 2, P2): the gh
	// availability probe is a network round-trip, and a user who never
	// consented must not cause even that. The per-item check below still
	// re-reads consent per item (a mid-run withdrawal).
	if !config.ReadUserParticipation().Enabled {
		return nil
	}
	if !s.Runner.Available(ctx) {
		// gh missing or unauthenticated: leave the items queued, quietly.
		return nil
	}
	repo := config.UserParticipationRepository()

	for _, item := range rec.Items {
		if ctx.Err() != nil {
			return nil // the flush time box: remaining items stay queued
		}
		// Per-item consent re-check: a user can withdraw while the run is
		// in flight; the very next item must see it.
		if !config.ReadUserParticipation().Enabled {
			return nil
		}
		if outbox.AttemptLimitReached(item) {
			s.drop(ctx, store, item, "attempt limit reached")
			continue
		}
		if !s.sendOne(ctx, store, item, repo) {
			return nil
		}
	}
	return nil
}

// beforeClaimForTest parks the sender just before it takes an item's
// claim — the deterministic interleave point for the stale-snapshot tests
// (nil in production).
var beforeClaimForTest func()

// sendOne carries one item from its ownership claim through the outcome
// record, and reports whether the RUN continues (false: stop — the caller
// returns and the remaining items stay queued). A claim the sender cannot
// take means another flush owns the item right now: skip it and keep
// running.
func (s *Sender) sendOne(ctx context.Context, store *feedback.QueueStore, item feedback.QueueItem, repo string) bool {
	if beforeClaimForTest != nil {
		beforeClaimForTest()
	}
	// Cross-process item ownership (review-gate finding, P2): the claim
	// spans the duplicate lookup through the outcome record, so two
	// concurrent flushes can no longer each send the same item.
	release, err := outbox.ClaimItemSend(ctx, item.ID)
	if err != nil {
		return true // another flush owns this item: skip it, keep running
	}
	defer func() { _ = release() }()

	// The claim excluded the concurrent holders but not the SEQUENTIAL
	// one: this flush's snapshot was loaded before the owning flush
	// finished, and that flush's complete() may already have sent and
	// removed the item (review-gate residual, P2). Re-check against the
	// LIVE queue and carry the current copy — gone means our snapshot is
	// stale and the item was already handled: skip.
	live, lerr := store.Load()
	if lerr != nil {
		return true // an unreadable queue is never a reason to send blind
	}
	found := false
	for i := range live.Items {
		if live.Items[i].ID == item.ID {
			item = live.Items[i]
			found = true
			break
		}
	}
	if !found {
		return true
	}

	// The attempt limit was judged against the SNAPSHOT before the claim
	// (review-gate finding, P2): a rival flush's fail() can exhaust the
	// item in between, and the fresh copy this walk now holds is the
	// exhausted one — sending it would burn one more search + create past
	// the limit. Re-judge on the live copy.
	if outbox.AttemptLimitReached(item) {
		s.drop(ctx, store, item, "attempt limit reached (re-checked on the live item)")
		return true
	}

	// A recorded send whose queue cleanup failed leaves the item in the
	// queue with the sent row already recorded — the next flush must not
	// publish again (review-gate hardening round, P2: the first run created
	// the issue; a resent run would add an occurrence comment for the same
	// report). The sent history is the completion record, scoped to the
	// DEC-3 duplicate-suppression window: an expired-window recurrence is a
	// new report and publishes (review-gate amendment).
	if item.Fingerprint != "" && outbox.SentHistoryHasFingerprintWithin(item.Fingerprint, time.Now(), config.DefaultBugreportFingerprintWindowDays) {
		_ = store.MutateContext(ctx, func(rec *feedback.QueueRecord) error {
			kept := rec.Items[:0]
			for _, it := range rec.Items {
				if it.ID == item.ID {
					continue
				}
				kept = append(kept, it)
			}
			rec.Items = kept
			return nil
		})
		_ = outbox.AppendOutbox(outbox.OutboxRow{
			Outcome:  "sent",
			Reason:   "queue cleanup reconciled on a later flush",
			Fingerpr: item.Fingerprint,
		})
		return true
	}

	// Send-time trust boundary (review-gate finding, P1): the queue
	// file is a local file, so the stored body is untrusted. Validate
	// BEFORE any gh call and publish only the regenerated title and
	// body — never the stored text as-is. The create body is rendered
	// again below WITH the item's summary; the stored text itself is
	// never sent.
	payload, title, _, ok := revalidatedItem(item)
	if !ok {
		s.fail(ctx, store, item, errUnvalidatedBody)
		return true
	}

	issue, markerCount, err := findIssue(ctx, s.Runner, repo, item)
	if err != nil {
		// A gh failure is environmental (network, rate limit): stop
		// the run rather than hammering, leave everything queued.
		s.fail(ctx, store, item, err)
		return false
	}

	if issue != nil {
		// Existing fingerprint: one occurrence comment below the cap,
		// nothing at the cap, never a body edit, never a label, and
		// never a model call (AC-016).
		if markerCount >= config.DefaultBugreportOccurrenceCommentsPerIssue {
			_ = outbox.AppendOutbox(outbox.OutboxRow{
				Outcome:  "capped_remote",
				Reason:   "occurrence comment cap reached on the remote issue",
				Fingerpr: item.Fingerprint,
			})
			return true
		}
		comment := OccurrenceCommentFromPayload(payload)
		if err := s.Runner.CommentIssue(ctx, repo, issue.Number, strings.NewReader(comment)); err != nil {
			s.fail(ctx, store, item, err)
			return false
		}
		s.complete(ctx, store, item, "occurrence comment on #"+strconv.Itoa(issue.Number), comment)
		return true
	}

	// No match: the create path. M6 slots the summarizer ahead of the
	// deterministic template text (design section 8; D38), bounded per
	// item (REQ-ANON-017/018) and by the daily cap.
	summary, _ := s.itemSummary(ctx, store, item, payload)
	_, body := outbox.RenderReportWithSummary(payload, summary)
	if err := s.Runner.CreateIssue(ctx, repo, title, strings.NewReader(body)); err != nil {
		s.fail(ctx, store, item, err)
		return false
	}
	s.complete(ctx, store, item, "issue created", body)
	return true
}

// itemSummary carries the per-item model-call bound (REQ-ANON-017/018,
// design section 8): a stored summary is REUSED on every retry; a template
// decision is final; the durable summary_requested marker lands BEFORE the
// call so a crash in the persist window takes the template — one call per
// queue item across any number of retries (D28). Every call counts against
// the rolling daily cap, at attempt time.
func (s *Sender) itemSummary(ctx context.Context, store *feedback.QueueStore, item feedback.QueueItem, payload bugreport.Payload) (string, bool) {
	if item.Summary != "" {
		// Send-time trust boundary (review-gate P1): the queue file is a
		// local file, so a STORED summary is untrusted exactly like the
		// stored body — validate it again before it can reach a public
		// body. A failure discards the text for the deterministic template
		// and records the template decision, so retries never re-ask the
		// model for the bad text.
		if validateSummary(item.Summary) == nil {
			return item.Summary, true
		}
		s.persistSummaryOutcome(ctx, store, item, "", summaryDecisionTemplate)
		return "", false
	}
	if item.SummaryDecision == summaryDecisionTemplate {
		return "", false
	}
	if item.SummaryRequested {
		// The marker without an outcome: a crash between the call and the
		// persist — the template decision, no second call (D28).
		s.persistSummaryOutcome(ctx, store, item, "", summaryDecisionTemplate)
		return "", false
	}
	if s.Summarizer == nil || !outbox.AllowAndRecordModelCall(ctx, time.Now()) {
		s.persistSummaryOutcome(ctx, store, item, "", summaryDecisionTemplate)
		return "", false
	}
	// The durable marker: a crash after the call but before the outcome
	// leaves the marker without a summary, and recovery takes the template.
	if err := store.MutateContext(ctx, func(rec *feedback.QueueRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == item.ID {
				rec.Items[i].SummaryRequested = true
				break
			}
		}
		return nil
	}); err != nil {
		return "", false // the marker could not land: template, never an unbounded call
	}
	// The attempt is already counted (AllowAndRecordModelCall above, one
	// atomic judgment-and-record): a failing endpoint hammers no further
	// than the cap, across senders and processes.
	out, err := s.Summarizer.Summarize(ctx, payload)
	if err != nil || validateSummary(out) != nil {
		s.persistSummaryOutcome(ctx, store, item, "", summaryDecisionTemplate)
		return "", false
	}
	s.persistSummaryOutcome(ctx, store, item, out, summaryDecisionModel)
	return out, true
}

// persistSummaryOutcome stores the summary (or the template decision) on
// the live item — the retry-reuse record.
func (s *Sender) persistSummaryOutcome(ctx context.Context, store *feedback.QueueStore, item feedback.QueueItem, summary, decision string) {
	_ = store.MutateContext(ctx, func(rec *feedback.QueueRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID != item.ID {
				continue
			}
			rec.Items[i].Summary = summary
			rec.Items[i].SummaryDecision = decision
			rec.Items[i].SummaryRequested = true
			break
		}
		return nil
	})
}

// complete records a successful send: the outbox log carries the sent
// payload FIRST (the completion record must survive a cleanup failure —
// the sender's reconcile rule then removes a surviving item on the next
// flush instead of publishing again), and the item leaves the queue.
func (s *Sender) complete(ctx context.Context, store *feedback.QueueStore, item feedback.QueueItem, reason, body string) {
	_ = outbox.AppendOutbox(outbox.OutboxRow{
		Outcome:  "sent",
		Reason:   reason,
		Title:    item.Title,
		Body:     body,
		Fingerpr: item.Fingerprint,
	})
	_ = store.MutateContext(ctx, func(rec *feedback.QueueRecord) error {
		kept := rec.Items[:0]
		for _, it := range rec.Items {
			if it.ID == item.ID {
				continue
			}
			kept = append(kept, it)
		}
		rec.Items = kept
		return nil
	})
}

// drop removes an exhausted item with a decision row (no payload for
// non-queued outcomes). The terminal discard marker lands FIRST (review
// gate finding 8): the item's reservation ends here, and the drain's
// recovery must never re-enroll it inside the dedupe window.
func (s *Sender) drop(ctx context.Context, store *feedback.QueueStore, item feedback.QueueItem, reason string) {
	outbox.RecordTerminalDiscard(ctx, item.Fingerprint)
	_ = store.MutateContext(ctx, func(rec *feedback.QueueRecord) error {
		kept := rec.Items[:0]
		for _, it := range rec.Items {
			if it.ID == item.ID {
				continue
			}
			kept = append(kept, it)
		}
		rec.Items = kept
		return nil
	})
	_ = outbox.AppendOutbox(outbox.OutboxRow{
		Outcome:  "dropped",
		Reason:   "send attempt limit: " + reason,
		Fingerpr: item.Fingerprint,
	})
}

// fail leaves the item queued, bumps its attempt count, and records the
// failure. The caller decides whether to continue or stop the run.
func (s *Sender) fail(ctx context.Context, store *feedback.QueueStore, item feedback.QueueItem, cause error) {
	_ = store.MutateContext(ctx, func(rec *feedback.QueueRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == item.ID {
				rec.Items[i].Attempts++
				break
			}
		}
		return nil
	})
	_ = outbox.AppendOutbox(outbox.OutboxRow{
		Outcome:  "send_failed",
		Reason:   cause.Error(),
		Fingerpr: item.Fingerprint,
	})
}
