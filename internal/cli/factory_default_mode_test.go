// factory_default_mode_test.go — SPEC-FACTORY-DEFAULT-MODE-001 M1 acceptance
// tests on the cli side of the default split: the queue-absent factory-card
// read derives parallelizable (AC-FDM-002), a plain admission records it
// (AC-FDM-003), and the decider-failure fail-safe stays serial with exactly
// one notice naming mode=serial (AC-FDM-004, regression guard).
package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestQueueAbsentCardClassificationDefault — AC-FDM-002: a card the queue
// record does not classify — a nil record, an id absent from items and
// archived, and a carried item whose classification field is absent — reads
// parallelizable through the same absent-field derivation, never a serial
// fallback.
func TestQueueAbsentCardClassificationDefault(t *testing.T) {
	if got := factoryQueueClassification(nil, "t9"); got.Mode != factory.ClassModeParallelizable {
		t.Errorf("nil record classification mode = %q, want parallelizable", got.Mode)
	}
	_, store := todoFixture(t)
	if _, _, err := store.Add("legacy card without classification"); err != nil {
		t.Fatalf("seed add: %v", err)
	}
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	if rec.Items[0].Classification != nil {
		t.Fatalf("fixture card %s carries a classification; the absent-field case needs one without", rec.Items[0].ID)
	}
	if got := factoryQueueClassification(rec, "t404"); got.Mode != factory.ClassModeParallelizable {
		t.Errorf("id-absent classification mode = %q, want parallelizable", got.Mode)
	}
	got := factoryQueueClassification(rec, rec.Items[0].ID)
	if got.Mode != factory.ClassModeParallelizable || got.Priority != factory.ClassPriorityNormal || got.Blocked || got.Decider != factory.DeciderIdentityDefault {
		t.Errorf("absent-field item classification = %+v, want parallelizable/normal/false/default", got)
	}
}

// TestTodoAddDefaultClassificationParallelizable — AC-FDM-003: a plain
// admission (no --classification-file) records mode parallelizable under
// decider default with a reason naming the deterministic default, and prints
// no fallback notice (the deterministic default is a judgment, not a
// failure).
func TestTodoAddDefaultClassificationParallelizable(t *testing.T) {
	_, store := todoFixture(t)
	installTodoCardDecider(t, factory.DefaultCardDecider{})

	out, errOut, err := runTodo(t, "add", "plain admission card")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if strings.Contains(errOut, "classification decider unavailable") {
		t.Errorf("plain add printed a fallback notice on stderr %q, want none", errOut)
	}
	id := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec.Items {
		if it.ID != id {
			continue
		}
		c := it.Classification
		if c == nil {
			t.Fatalf("card %s carries no classification", id)
		}
		if c.Mode != factory.ClassModeParallelizable || c.Decider != factory.DeciderIdentityDefault || c.Priority != factory.ClassPriorityNormal || c.Blocked {
			t.Errorf("default admission classification = %+v, want parallelizable under decider default (normal, not blocked)", c)
		}
		if !strings.Contains(c.Reason, "deterministic default") {
			t.Errorf("default admission reason %q does not name the deterministic default", c.Reason)
		}
	}
}

// TestTodoAddDeciderFailureFailsafeSerial — AC-FDM-004 (regression guard,
// demoted per acceptance.md §D.0 — green-now, kept green through the flip): a
// decider failure promotes the SERIAL fail-safe — never the absent-read
// default — with exactly one stderr notice naming mode=serial, and admission
// is not refused. The guard keeps the two defaults separated (plan D2) once
// the absent-read default flips to parallelizable.
func TestTodoAddDeciderFailureFailsafeSerial(t *testing.T) {
	_, store := todoFixture(t)
	installTodoCardDecider(t, failingCardDecider{})

	out, errOut, err := runTodo(t, "add", "fail-safe card")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if n := strings.Count(errOut, "classification decider unavailable"); n != 1 {
		t.Errorf("stderr carries %d fallback notices, want exactly 1 (got %q)", n, errOut)
	}
	if !strings.Contains(errOut, "mode=serial") {
		t.Errorf("fallback notice %q does not name mode=serial", errOut)
	}
	id := strings.TrimSpace(strings.SplitN(out, " ", 2)[0])
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range rec.Items {
		if it.ID != id {
			continue
		}
		c := it.Classification
		if c == nil {
			t.Fatalf("fail-safe card %s carries no classification", id)
		}
		if c.Mode != factory.ClassModeSerial || c.Decider != factory.DeciderIdentityDefault {
			t.Errorf("fail-safe classification = %+v, want mode serial under decider default (the fail-safe is NOT the absent-read default)", c)
		}
	}
}

// TestFactorySelectionReadsReclassifiedMode — AC-FDM-008 (guard-addition,
// green-by-construction once M1/M2 land): a factory-recorded in-flight card
// reclassified by the verb is observed by BOTH the next selection pass (the
// serial exclusivity bites against the reclassified mode) and the status
// view (the mode cell renders the new value) — with no write beyond the
// verb's single locked write (REQ-FDM-005, the single-store property).
func TestFactorySelectionReadsReclassifiedMode(t *testing.T) {
	root, store := fcFixture(t)
	// t1: in flight (live lease) as a parallelizable card; t2/t3: queued
	// serial candidates.
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStateQueued, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t3", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardRun, OwnerLabel: "lane-2", LeaseHolder: "lane-2", LeaseExpiresAt: live, Stage: homestate.CardRun},
	)
	sdRegisterLane(t, root, "lane-1")

	// The verb reclassifies the in-flight t1 to serial — allowed, because no
	// other serial card is in flight yet.
	out, _, err := runTodo(t, "classify", "t1", "--mode", "serial", "--reason", "t1 edits the serial-slot surface")
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	if !strings.Contains(out, "classified t1 mode=serial") {
		t.Fatalf("stdout = %q, want the confirmation line", out)
	}

	// The next selection pass honors the reclassified mode: the queued
	// serial t2/t3 wait against the reclassified t1 instead of leasing
	// beside it.
	got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-1")
	if err != nil {
		t.Fatalf("next after reclassification: %v", err)
	}
	if owned {
		t.Fatalf("lane-1 leased %s although the reclassified serial t1 holds the serial slot", got.CardID)
	}

	// The status view renders the reclassified mode from the same queue
	// read — the one write the verb made is the only synchronization point.
	statusOut, _, err := runFactory(t, "status", "--json")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	var report struct {
		Cards []struct {
			CardID string `json:"card_id"`
			Mode   string `json:"mode"`
		} `json:"cards"`
	}
	if err := json.Unmarshal([]byte(statusOut), &report); err != nil {
		t.Fatalf("parse status json: %v\n%s", err, statusOut)
	}
	found := false
	for _, c := range report.Cards {
		if c.CardID != "t1" {
			continue
		}
		found = true
		if c.Mode != factory.ClassModeSerial {
			t.Errorf("status view mode for t1 = %q, want the reclassified serial", c.Mode)
		}
	}
	if !found {
		t.Errorf("status view carries no t1 card: %s", statusOut)
	}
}

// TestSerialExclusivityAfterDefaultFlip — AC-FDM-009 (regression guard): the
// serial-vs-serial exclusivity is unchanged by the default flip. A serial
// card holding the slot blocks a queued serial candidate, while an
// UNCLASSIFIED candidate — the flip's parallelizable default — leases beside
// the holder (REQ-FDM-007).
func TestSerialExclusivityAfterDefaultFlip(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	// t3 reverts to the absent-field shape the flip governs: no record.
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID == "t3" {
				rec.Items[i].Classification = nil
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardRun, OwnerLabel: "lane-9", LeaseHolder: "lane-9", LeaseExpiresAt: live, Stage: homestate.CardRun},
	)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")

	got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-1")
	if err != nil {
		t.Fatalf("lane-1 next: %v", err)
	}
	if !owned || got.CardID != "t3" {
		t.Fatalf("lane-1 next = (%s, owned=%v), want the unclassified t3 leased beside the serial holder", got.CardID, owned)
	}
	got2, owned2, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-2")
	if err != nil {
		t.Fatalf("lane-2 next: %v", err)
	}
	if owned2 {
		t.Fatalf("lane-2 leased %s while the serial t1 holds the slot — exclusivity did not survive the flip", got2.CardID)
	}
}
