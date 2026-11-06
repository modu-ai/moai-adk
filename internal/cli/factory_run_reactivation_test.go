package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/discovery"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// REQ-FCR-024 (card t1963, ledger L12) at the leader-entry surface: starting or
// resuming a leader under the id of a run a consolidation retired is refused
// with the typed error, and the message names both the consolidated run and the
// run its cards went to. Both writers wrap with %w, so the typed error survives.
func TestLeaderEntryRefusesAConsolidatedRun(t *testing.T) {
	check := func(t *testing.T, err error) {
		t.Helper()
		if !errors.Is(err, homestate.ErrRunRetiredByConsolidation) {
			t.Fatalf("err = %v, want homestate.ErrRunRetiredByConsolidation", err)
		}
		if !strings.Contains(err.Error(), "run-old") || !strings.Contains(err.Error(), "run-new") {
			t.Fatalf("refusal %q must name the consolidated run (run-old) and its target (run-new)", err)
		}
	}
	t.Run("start", func(t *testing.T) {
		_, _ = fcFixture(t)
		root := factoryCardRoot()
		caaSeedMovedCard(t, root)
		t.Setenv(config.EnvFactoryRunID, "run-old")
		check(t, recordLeaderRunStart(root, "", BackendClaude, "", 0))
	})
	t.Run("resume", func(t *testing.T) {
		_, _ = fcFixture(t)
		root := factoryCardRoot()
		caaSeedMovedCard(t, root)
		check(t, resumeDiscoveredRun(root, discovery.VerifiedLeader{RunID: "run-old", PID: 4242, ProcessStart: "fp", Basis: "test"}))
	})
}

// CARRY-5 (card t1963, ledger L14): the refusal of a consolidated run id comes
// BEFORE any queue-side write. recordFactoryRunStart used to write the queue's
// runtime run record first and only then reach the factory guard, so a refused
// start still left a queue-side run row (its manifest) for a retired run. The
// instrument is the canonical queue reader (LoadPure), with a positive control:
// the same start under a live run id does leave a queue-side row, so an empty
// reading under the consolidated id is a measured absence, not a blind reader.
func TestLeaderStartRefusalWritesNoQueueSideRunRecord(t *testing.T) {
	queueRuns := func(t *testing.T, root string) map[string]bool {
		t.Helper()
		record, err := factory.NewBacklogStore(factory.BacklogPathForRoot(root)).LoadPure()
		if err != nil {
			t.Fatalf("load the queue record: %v", err)
		}
		runs := map[string]bool{}
		for _, r := range record.Runtime.Runs {
			runs[r.RunID] = true
		}
		return runs
	}
	_, _ = fcFixture(t)
	root := factoryCardRoot()
	caaSeedMovedCard(t, root)

	t.Setenv(config.EnvFactoryRunID, "run-live-control")
	if err := recordLeaderRunStart(root, "", BackendClaude, "", 0); err != nil {
		t.Fatalf("control start under a live run id: %v", err)
	}
	if !queueRuns(t, root)["run-live-control"] {
		t.Fatalf("control: a start under a live run id left no queue-side run record; the reader cannot see queue-side writes")
	}

	t.Setenv(config.EnvFactoryRunID, "run-old")
	err := recordLeaderRunStart(root, "", BackendClaude, "", 0)
	if !errors.Is(err, homestate.ErrRunRetiredByConsolidation) {
		t.Fatalf("err = %v, want homestate.ErrRunRetiredByConsolidation", err)
	}
	if queueRuns(t, root)["run-old"] {
		t.Fatalf("a refused start under the consolidated run id still wrote a queue-side run record")
	}
}
