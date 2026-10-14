// factory_renew_cadence_test.go — SPEC-FACTORY-LEASE-RENEW-001 M4 acceptance test (card t1663):
// AC-FLR-012 (REQ-FLR-001, REQ-FLR-004, REQ-FLR-007, REQ-FLR-009, REQ-FLR-012). A lane that
// renews on the renewal period keeps its card lease live through 3D; a renewal delayed to one
// second short of the expiry succeeds and extends the lease from its request time; a renewal at
// the expiry instant, or one second later, is refused as expired with no change to the card.
// The clock is the injected factoryCardNow, so the scenario runs without a wall-clock wait.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factoryCadenceConfig writes workflow.factory_lease into the fixture root with the card lease
// duration d and the renewal period i, then reads the effective values back. The read-back makes
// the scenario fail on a configuration it did not write, so it never runs against the defaults
// by accident.
func factoryCadenceConfig(t *testing.T, root string, d, i time.Duration) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	body := fmt.Sprintf("workflow:\n  factory_lease:\n    duration: %dm\n    renew_interval: %dm\n",
		int(d/time.Minute), int(i/time.Minute))
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}
	gotDuration, gotInterval, err := homestate.EffectiveFactoryLease(root)
	if err != nil {
		t.Fatalf("effective factory lease: %v", err)
	}
	if gotDuration != d || gotInterval != i {
		t.Fatalf("effective factory lease = (%s, %s), want (%s, %s)", gotDuration, gotInterval, d, i)
	}
}

// factoryCadenceRenewAt sets the injected clock to at, runs the renew verb, and returns the
// outcome the verb prints together with the verb's error.
func factoryCadenceRenewAt(t *testing.T, at time.Time) (factoryRenewWire, error) {
	t.Helper()
	factoryCardNow = func() time.Time { return at }
	return factoryRenewViaVerb(t)
}

// TestRenewCadenceKeepsLeaseLive is AC-FLR-012. The scenario runs for three (duration,
// renewal period) pairs: the defaults, the boundary (the renewal period is one half of the
// duration), and a non-default pair. The lease is acquired at the start instant T with expiry
// T plus the duration, and every renewal is requested through the renew verb.
func TestRenewCadenceKeepsLeaseLive(t *testing.T) {
	cases := []struct {
		name     string
		duration time.Duration
		interval time.Duration
	}{
		{"defaults", 15 * time.Minute, 5 * time.Minute},
		{"boundary", 10 * time.Minute, 5 * time.Minute},
		{"non-default", 20 * time.Minute, 4 * time.Minute},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// place opens a fresh fixture: the configured pair, lane-1 registered, and card t1
			// leased to lane-1 from T until T plus the duration. It returns the fixture root.
			place := func(t *testing.T) string {
				t.Helper()
				root, _ := sdMoaiFixture(t)
				factoryCadenceConfig(t, root, c.duration, c.interval)
				sdRegisterLane(t, root, "lane-1")
				factoryRenewPlace(t, root, "lane-1", fcNow.Add(c.duration))
				sdLaneEnv(t, "lane-1", "")
				return root
			}

			t.Run("on-schedule renewals at 0, I, 2I, through 3D succeed", func(t *testing.T) {
				root := place(t)
				before, _ := factoryRenewRecord(t, root)
				steps := int(3 * c.duration / c.interval)
				var last factoryRenewWire
				var lastAt time.Time
				for k := 0; k <= steps; k++ {
					offset := time.Duration(k) * c.interval
					at := fcNow.Add(offset)
					wire, err := factoryCadenceRenewAt(t, at)
					if err != nil || wire.Result != "renewed" {
						t.Fatalf("on-schedule renewal %d at +%s: result = %q, err = %v; want renewed",
							k, offset, wire.Result, err)
					}
					last, lastAt = wire, at
				}
				want := lastAt.Add(c.duration).UTC().Format(time.RFC3339Nano)
				if last.ExpiresAt != want {
					t.Errorf("last on-schedule renewal expires_at = %q, want %q (request time plus %s)",
						last.ExpiresAt, want, c.duration)
				}
				after, _ := factoryRenewRecord(t, root)
				if after.LeaseExpiresAt != want {
					t.Errorf("lease expiry after the last renewal = %s, want %s", after.LeaseExpiresAt, want)
				}
				if after.Version != before.Version || after.State != before.State {
					t.Errorf("renewals changed the card: v%d %s → v%d %s",
						before.Version, before.State, after.Version, after.State)
				}
			})

			t.Run("a renewal due at I and delayed to D minus one second succeeds", func(t *testing.T) {
				root := place(t)
				// The last renewal was at T.
				if wire, err := factoryCadenceRenewAt(t, fcNow); err != nil || wire.Result != "renewed" {
					t.Fatalf("renewal at T: result = %q, err = %v; want renewed", wire.Result, err)
				}
				before, eventsBefore := factoryRenewRecord(t, root)
				// Due at T plus I, and requested at T plus I plus (D minus I minus one second).
				delay := c.duration - c.interval - time.Second
				at := fcNow.Add(c.interval + delay)
				wire, err := factoryCadenceRenewAt(t, at)
				if err != nil || wire.Result != "renewed" {
					t.Fatalf("late renewal due at +%s, requested at +%s: result = %q, err = %v; want renewed",
						c.interval, c.interval+delay, wire.Result, err)
				}
				want := at.Add(c.duration).UTC().Format(time.RFC3339Nano)
				if wire.ExpiresAt != want {
					t.Errorf("late renewal expires_at = %q, want %q (request time plus %s)",
						wire.ExpiresAt, want, c.duration)
				}
				after, eventsAfter := factoryRenewRecord(t, root)
				if after.LeaseExpiresAt != want {
					t.Errorf("lease expiry after the late renewal = %s, want %s", after.LeaseExpiresAt, want)
				}
				if after.Version != before.Version || after.State != before.State {
					t.Errorf("late renewal changed the card: v%d %s → v%d %s",
						before.Version, before.State, after.Version, after.State)
				}
				if eventsAfter != eventsBefore {
					t.Errorf("late renewal appended %d event rows, want none", eventsAfter-eventsBefore)
				}
			})

			t.Run("renewals at D and at D plus one second are refused as expired", func(t *testing.T) {
				root := place(t)
				// The last renewal was at T, so the lease expires at T plus the duration.
				if wire, err := factoryCadenceRenewAt(t, fcNow); err != nil || wire.Result != "renewed" {
					t.Fatalf("renewal at T: result = %q, err = %v; want renewed", wire.Result, err)
				}
				before, eventsBefore := factoryRenewRecord(t, root)
				lapsed := fcNow.Add(c.duration).UTC().Format(time.RFC3339Nano)
				for _, offset := range []time.Duration{c.duration, c.duration + time.Second} {
					wire, err := factoryCadenceRenewAt(t, fcNow.Add(offset))
					if err == nil {
						t.Errorf("renewal at +%s: err = nil, want a non-zero exit", offset)
					}
					if wire.Result != "refused-expired" {
						t.Errorf("renewal at +%s: result = %q, want refused-expired", offset, wire.Result)
					}
					if wire.ExpiresAt != lapsed {
						t.Errorf("renewal at +%s: expires_at = %q, want the lapsed expiry %q", offset, wire.ExpiresAt, lapsed)
					}
					after, eventsAfter := factoryRenewRecord(t, root)
					if after.Version != before.Version || after.State != before.State || after.LeaseExpiresAt != before.LeaseExpiresAt {
						t.Errorf("refused renewal at +%s changed the card: v%d %s expiry=%s → v%d %s expiry=%s",
							offset, before.Version, before.State, before.LeaseExpiresAt,
							after.Version, after.State, after.LeaseExpiresAt)
					}
					if eventsAfter != eventsBefore {
						t.Errorf("refused renewal at +%s appended %d event rows, want none", offset, eventsAfter-eventsBefore)
					}
				}
			})
		})
	}
}
