// integration_wait.go — `moai integration acquire --wait`'s waiting loop
// (card t1479, SPEC-MERGE-WINDOW-QUEUE-001 REQ-MWQ-002/003/004/005).
//
// The Bash tool call is capped at 10 minutes, below the 60-minute default
// wait, so a lane runs this as a background command (spec.md §D) — the loop
// is therefore a plain blocking poll, and its exit IS the verdict: promoted
// (success), bound elapsed (REQ-MWQ-004), dropped by liveness
// (REQ-MWQ-003's scenario 5), or promoted past the bound and released
// onward (REQ-MWQ-005).
package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// AcquireWaitDefaultBound is the bare --wait bound (REQ-MWQ-002): 60 minutes
// from the enqueue instant.
const AcquireWaitDefaultBound = 60 * time.Minute

// parseAcquireWait parses the --wait flag: absent → (0, false, nil); bare
// (empty value) → (60m, true, nil); a duration → (parsed, true, nil).
func parseAcquireWait(value string) (time.Duration, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false, nil
	}
	if value == "true" {
		return AcquireWaitDefaultBound, true, nil
	}
	bound, err := time.ParseDuration(value)
	if err != nil || bound <= 0 {
		return 0, false, fmt.Errorf("acquire --wait: %q is not a positive duration (bare --wait is %s; --wait=2m is two minutes)", value, AcquireWaitDefaultBound)
	}
	return bound, true, nil
}

// integrationLeaseDuration reads the configured lease duration (REQ-MWQ-008):
// absent → 0 (the factory default of 30 minutes decides); zero → negative
// (the lease is disabled); N minutes → N. The config layer owns the
// absent-vs-zero distinction; this translation is the only place the window
// verbs read it.
func integrationLeaseDuration(projectRoot string) time.Duration {
	minutes := integrationLeaseMinutes(projectRoot)
	switch {
	case minutes == nil:
		return 0 // factory default
	case *minutes == 0:
		return -1 // lease disabled
	default:
		return time.Duration(*minutes) * time.Minute
	}
}

// integrationLeaseMinutes reads workflow.integration_lock.lease_minutes.
// Every read failure (no config, unreadable section) reads as absent — the
// factory default decides, and the window still works.
func integrationLeaseMinutes(projectRoot string) *int {
	cfg, err := config.NewLoader().Load(filepath.Join(projectRoot, ".moai"))
	if err != nil {
		return nil
	}
	return cfg.Workflow.IntegrationLock.LeaseMinutes
}

// initWindowLeaseOverride initializes the factory's package-level lease
// override from the configured value, so the INTERNAL callers (the
// release-path promotion, the status refresh) stamp the configured duration
// too — an explicit lease_minutes: 0 disables the lease everywhere, not
// only on acquire (card-review r1 P2-7). Absent config leaves the factory
// default. Assigned on every verb call (the config read is cheap and the
// value is a single assignment) so parallel tests with distinct fixtures
// cannot inherit a stale override.
func initWindowLeaseOverride(projectRoot string) {
	if minutes := integrationLeaseMinutes(projectRoot); minutes != nil {
		factory.WindowLeaseDuration = time.Duration(*minutes) * time.Minute
	} else {
		factory.WindowLeaseDuration = factory.IntegrationLeaseDefault
	}
}

// integrationWaitPollInterval is the loop's sleep between record reads. It
// is a package variable so a test can shorten it — a poll loop that waited
// its own production interval in a unit test would measure the sleep, not
// the queue.
var integrationWaitPollInterval = time.Second

// integrationWaitInQueue is the --wait body after the acquire was refused
// for a live holder or a hold: enqueue (inside the same serialized mutation
// the order is decided in), then poll the record until promoted, dropped, or
// timed out.
func integrationWaitInQueue(root, sessionID string, ticket factory.IntegrationTicket, bound time.Duration, out io.Writer) error {
	// The REAL window policy governs the whole wait — the enqueue's
	// refresh, the bound decision, and every heartbeat renewal
	// (card-review r1 P1-2/P2-3: a hardcoded open here let the wait path's
	// own mutations promote through a hold and skip the liveness drops).
	policy, policyErr := factory.ReadIntegrationWindowPolicy(root)
	if policyErr != nil {
		return policyErr
	}
	// REQ-MWQ-002: one ticket at the tail, the order decided inside the same
	// serialized record mutation that guards acquire and release; the
	// enqueue instant starts the bound's clock.
	err := factory.UpdateIntegrationWindow(root, func(w *factory.IntegrationLock) error {
		return factory.EnqueueTicket(w, ticket, factory.DefaultWindowProcProbe(), factory.WindowClock(), policy)
	})
	if err != nil {
		return err
	}
	enqueuedAt := factory.WindowClock()
	ticket.EnqueuedAt = enqueuedAt.Format(time.RFC3339)
	ticket.Heartbeat = ticket.EnqueuedAt
	deadline := enqueuedAt.Add(bound)
	lastBeat := enqueuedAt
	if out != nil {
		_, _ = fmt.Fprintf(out, "queued at position %d (bound %s, holder ahead); waiting\n", factory.TicketPosition(mustReadWindow(root), sessionID), bound)
	}
	for {
		time.Sleep(integrationWaitPollInterval)
		now := factory.WindowClock()

		// The promotion and the bound are decided in the same mutation
		// (REQ-MWQ-005): PromotedAfterBound releases onward at once when the
		// promotion is observed past the bound, and returns the promotion
		// otherwise — under the real policy.
		var released bool
		var holderNow bool
		mutErr := factory.UpdateIntegrationWindow(root, func(w *factory.IntegrationLock) error {
			outcome, pErr := factory.PromotedAfterBound(w, sessionID, deadline, factory.DefaultWindowProcProbe(), now, factory.WindowLeaseDuration, policy)
			if pErr != nil {
				return pErr
			}
			released = outcome.Released
			holderNow = w.Held() && w.SessionID == sessionID
			return nil
		})
		if mutErr == nil {
			if released {
				// REQ-MWQ-005: promoted past the bound, released onward.
				return fmt.Errorf("integration window: promoted past your %s bound and released onward — re-acquire with --wait to re-enter the queue", bound)
			}
			if holderNow {
				if out != nil {
					_, _ = fmt.Fprintln(out, "integration window acquired from the queue")
				}
				return nil
			}
		} //nolint:staticcheck // QF1001: the early returns above are the clarity this shape exists for

		// A dropped ticket (REQ-MWQ-003 scenario 5): the waiter observes its
		// own disappearance and exits non-zero naming the reason, without
		// re-enqueueing.
		lock, readErr := mustReadWindowErr(root)
		if readErr != nil {
			return readErr
		}
		if factory.TicketPosition(lock, sessionID) == 0 && !(lock.Held() && lock.SessionID == sessionID) { //nolint:staticcheck // QF1001 — the negated conjunction IS the documented drop condition
			return fmt.Errorf("integration window: your ticket was dropped (%s) — re-acquire with --wait re-enters at the tail", waiterDropReason(lock, ticket, now))
		}

		// REQ-MWQ-002: refresh the ticket's heartbeat every 15 seconds while
		// blocking. The renewal is a queue MUTATION, so it runs the liveness
		// refresh under the real policy (card-review r1 P2-3) — a dead
		// ticket cannot stay queued just because its lane is still polling.
		if now.Sub(lastBeat) >= factory.WaiterHeartbeatInterval {
			lastBeat = now
			if err := factory.UpdateIntegrationWindow(root, func(w *factory.IntegrationLock) error {
				report := factory.RefreshWindow(w, policy, factory.DefaultWindowProcProbe(), now, factory.WindowLeaseDuration)
				for i := range w.Queue {
					if w.Queue[i].SessionID == sessionID {
						w.Queue[i].Heartbeat = now.Format(time.RFC3339)
					}
				}
				if len(report.Dropped) > 0 && out != nil {
					for _, d := range report.Dropped {
						_, _ = fmt.Fprintf(out, "dropped ticket: %s\n", d)
					}
				}
				return nil
			}); err != nil {
				return err
			}
		}

		// REQ-MWQ-004: the bound elapsed before promotion — withdraw and
		// name the holder, the last queue position, and the bound.
		if now.After(deadline) {
			holder := "nobody"
			if lock.Held() {
				holder = holderLabel(lock)
			}
			position := factory.TicketPosition(lock, sessionID)
			_ = factory.UpdateIntegrationWindow(root, func(w *factory.IntegrationLock) error {
				factory.WithdrawTicket(w, sessionID)
				return nil
			})
			return fmt.Errorf("integration window: your ticket timed out after %s at queue position %d behind %s — re-acquire with --wait re-enters at the tail", bound, position, holder)
		}
	}
}

// waiterDropReason reconstructs why a waiter's own ticket is gone: the
// waiter is alive by construction (it is running), so the drop it observes
// was owner-gone or heartbeat-stale — decided from what this process can
// see.
func waiterDropReason(lock *factory.IntegrationLock, ticket factory.IntegrationTicket, now time.Time) string {
	if ticket.OwnerPID > 0 && !factory.FactoryProcessAlive(ticket.OwnerPID) {
		return "owner gone"
	}
	if beat, err := time.Parse(time.RFC3339, ticket.Heartbeat); err == nil && now.Sub(beat) > factory.WaiterHeartbeatWindow {
		return "heartbeat stale"
	}
	return "ticket removed from the queue"
}

// mustReadWindow reads the window record, tolerating absence (an empty
// record) — the waiter's poll may run after a queue-less release removed
// the file.
func mustReadWindow(root string) *factory.IntegrationLock {
	lock, err := mustReadWindowErr(root)
	if err != nil {
		return &factory.IntegrationLock{}
	}
	return lock
}

func mustReadWindowErr(root string) (*factory.IntegrationLock, error) {
	return factory.ReadIntegrationLock(root)
}

// holderLabelForCLI renders the holder label for CLI messages (the same
// label the refusals use).
func holderLabel(l *factory.IntegrationLock) string {
	if l == nil {
		return "unknown"
	}
	if l.SessionName != "" {
		return l.SessionName
	}
	if l.SessionID != "" {
		return l.SessionID
	}
	return "unknown"
}

// production.
