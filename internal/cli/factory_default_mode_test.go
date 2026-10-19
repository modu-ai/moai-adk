// factory_default_mode_test.go — SPEC-FACTORY-DEFAULT-MODE-001 M1 acceptance
// tests on the cli side of the default split: the queue-absent factory-card
// read derives parallelizable (AC-FDM-002), a plain admission records it
// (AC-FDM-003), and the decider-failure fail-safe stays serial with exactly
// one notice naming mode=serial (AC-FDM-004, regression guard).
package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
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
