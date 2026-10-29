// factory_card_spec_propagation_test.go — card t1921 defect ① regression
// tests: the three lease-path record-CREATION sites (arm (b2)'s queue-picked
// rowless candidate, arm (c)'s promoted candidate, and the nominated lease)
// propagate a well-formed queue spec id into the created row; a malformed
// queue spec never turns a working lease into a validation refusal; and an
// existing row keeps whatever it carries (creation-only, the t1533 rule).
//
// Every fixture is built under t.TempDir() with an isolated git config and
// MOAI_HOME sandboxed away (§B of acceptance.md); ./internal/cli runs only
// through the anchored -run selectors naming one of these tests.
package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// spLease runs one `factory next` lease as lane-1 over the fixture, asserts
// the card ended leased by lane-1, and returns the row for field asserts.
func spLease(t *testing.T, root string, nominated bool) homestate.Card {
	t.Helper()
	args := []string{"next", "--run", fcRun}
	if nominated {
		args = append(args, "--card", "t1")
	}
	out, stderr, err := runFactory(t, args...)
	if err != nil {
		t.Fatalf("next: %v\nstdout: %s\nstderr: %s", err, out, stderr)
	}
	row := fcCard(t, root, "t1")
	if row.State != homestate.CardLeased || row.LeaseHolder != "lane-1" {
		t.Fatalf("t1 = %s holder=%q, want leased/lane-1", row.State, row.LeaseHolder)
	}
	return row
}

func TestFactoryNextLeasePropagatesQueueSpecID(t *testing.T) {
	// Arm (b2): a queue-picked card with no record row — the creation the
	// rowless arm records with the queue item's spec id.
	t.Run("b2-creation", func(t *testing.T) {
		root, store := fcFixture(t)
		cepSelfDispatchYAML(t, root, true)
		fcQueue(t, store, factory.BacklogStatePicked)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		cepSetSpecID(t, store, "t1", "SPEC-X-001")
		sdRegisterLane(t, root, "lane-1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		if got := spLease(t, root, false).SpecID; got != "SPEC-X-001" {
			t.Errorf("t1 row SpecID = %q, want %q propagated from the queue item at arm (b2) creation", got, "SPEC-X-001")
		}
	})

	// Arm (c): the promoted queued candidate — the creation the promotion
	// records with the queue item's spec id.
	t.Run("c-creation", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStateQueued)
		cepSetSpecID(t, store, "t1", "SPEC-X-001")
		sdRegisterLane(t, root, "lane-1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		if got := spLease(t, root, false).SpecID; got != "SPEC-X-001" {
			t.Errorf("t1 row SpecID = %q, want %q propagated from the queue item at arm (c) creation", got, "SPEC-X-001")
		}
	})

	// The nominated lease: `next --card` over a queued rowless nominee — the
	// creation the nominated path records with the queue item's spec id.
	t.Run("nominated-creation", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStateQueued)
		cepSetSpecID(t, store, "t1", "SPEC-X-001")
		sdRegisterLane(t, root, "lane-1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		if got := spLease(t, root, true).SpecID; got != "SPEC-X-001" {
			t.Errorf("t1 row SpecID = %q, want %q propagated from the queue item at the nominated creation", got, "SPEC-X-001")
		}
	})

	// A malformed queue spec (the store keeps spec as-is, with no registry
	// validation) keeps today's behavior: the lease succeeds and the row
	// carries no spec — no new claim-refusal mode.
	t.Run("malformed-no-refusal", func(t *testing.T) {
		root, store := fcFixture(t)
		cepSelfDispatchYAML(t, root, true)
		fcQueue(t, store, factory.BacklogStatePicked)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		cepSetSpecID(t, store, "t1", "not-a-spec")
		sdRegisterLane(t, root, "lane-1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		if got := spLease(t, root, false).SpecID; got != "" {
			t.Errorf("t1 row SpecID = %q, want empty for a malformed queue spec (no propagation, no refusal)", got)
		}
	})

	// An existing picked+unowned row (the pre-fix spec-less shape) keeps
	// whatever it carries: the lease updates the row, the spec id stays empty.
	t.Run("existing-row-untouched", func(t *testing.T) {
		root, store := fcFixture(t)
		cepSelfDispatchYAML(t, root, true)
		fcQueue(t, store, factory.BacklogStatePicked)
		fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		cepSetSpecID(t, store, "t1", "SPEC-X-001")
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
		sdRegisterLane(t, root, "lane-1")
		sdLaneEnv(t, "lane-1", "")
		t.Chdir(root)
		if got := spLease(t, root, false).SpecID; got != "" {
			t.Errorf("t1 row SpecID = %q, want empty — creation-only: an existing row keeps what it carries", got)
		}
	})
}
