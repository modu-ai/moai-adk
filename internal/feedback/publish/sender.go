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
	"strings"

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

// Sender files queued reports through one Runner.
type Sender struct {
	Runner Runner
}

// NewSender returns a sender over r.
func NewSender(r Runner) *Sender {
	return &Sender{Runner: r}
}

// FlushContext is the production entry the CLI's flush call drives: the
// send half of flush, over the real gh, under the caller's context.
//
// @MX:ANCHOR: [AUTO] FlushContext — the CLI flush and the update-end trigger both enter the sender here
// @MX:REASON: a second sender entry could skip the hook-path refusal or the consent re-checks this walk enforces (REQ-ANON-015)
// @MX:WARN: [AUTO] every gh call is a public side effect from the user's account
// @MX:REASON: the sender's caps, consent checks, and refusal gates are the only things standing between a queue and a public post (REQ-ANON-015/016)
func FlushContext(ctx context.Context) error {
	return NewSender(newExecRunner()).Send(ctx)
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
			s.drop(store, item, "attempt limit reached")
			continue
		}

		issue, markerCount, err := findIssue(ctx, s.Runner, repo, item)
		if err != nil {
			// A gh failure is environmental (network, rate limit): stop
			// the run rather than hammering, leave everything queued.
			s.fail(store, item, err)
			return nil
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
				continue
			}
			body := OccurrenceComment(item)
			if body == "" {
				// A queued body without a marker block never belongs on
				// this path; leave it queued rather than comment blind.
				s.fail(store, item, errNoMarkerBlock)
				continue
			}
			if err := s.Runner.CommentIssue(ctx, repo, issue.Number, strings.NewReader(body)); err != nil {
				s.fail(store, item, err)
				return nil
			}
			s.complete(store, item, "occurrence comment on #"+itoa(issue.Number), body)
			continue
		}

		// No match: the create path. M5 files the deterministic template
		// text; M6 slots the summarizer ahead of CreateBody.
		body := CreateBody(item)
		if err := s.Runner.CreateIssue(ctx, repo, item.Title, strings.NewReader(body)); err != nil {
			s.fail(store, item, err)
			return nil
		}
		s.complete(store, item, "issue created", body)
	}
	return nil
}

// errNoMarkerBlock marks a queued item whose body lost its marker block.
var errNoMarkerBlock = errorString("queued body carries no bugreport marker block")

type errorString string

func (e errorString) Error() string { return string(e) }

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

// complete records a successful send: the item leaves the queue and the
// outbox log carries the sent payload (queued/sent/withheld rows hold the
// exact payload; decision rows do not).
func (s *Sender) complete(store *feedback.QueueStore, item feedback.QueueItem, reason, body string) {
	_ = store.Mutate(func(rec *feedback.QueueRecord) error {
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
		Reason:   reason,
		Title:    item.Title,
		Body:     body,
		Fingerpr: item.Fingerprint,
	})
}

// drop removes an exhausted item with a decision row (no payload for
// non-queued outcomes).
func (s *Sender) drop(store *feedback.QueueStore, item feedback.QueueItem, reason string) {
	_ = store.Mutate(func(rec *feedback.QueueRecord) error {
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
func (s *Sender) fail(store *feedback.QueueStore, item feedback.QueueItem, cause error) {
	_ = store.Mutate(func(rec *feedback.QueueRecord) error {
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
