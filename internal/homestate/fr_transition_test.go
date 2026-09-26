package homestate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

const frRun = "run-f1"

// frExpectedEdges is design.md § Transition Table's requested-edge set with a
// remote configured: 65 (from, to) pairs. It is written out here rather than
// read from the production table, so the test compares two independent lists.
func frExpectedEdges() map[[2]string]bool {
	out := map[[2]string]bool{}
	add := func(from, to string) { out[[2]string{from, to}] = true }
	for _, e := range [][2]string{
		{CardPicked, CardAssigned}, {CardAssigned, CardLeased},
		{CardPlan, CardPlanAudit}, {CardPlanAudit, CardPlan}, {CardPlanAudit, CardKickoff},
		{CardKickoff, CardAssigned}, {CardKickoff, CardBlocked},
		{CardRun, CardSync}, {CardSync, CardSyncAudit}, {CardSyncAudit, CardSync}, {CardSyncAudit, CardMergeReady},
		{CardMergeReady, CardMerging}, {CardMerging, CardMergeReady}, {CardMerging, CardMergedLocal},
		{CardMergedLocal, CardPushed},
	} {
		add(e[0], e[1])
	}
	for _, s := range []string{CardPlan, CardRun, CardPlanAudit, CardSync, CardSyncAudit, CardMergeReady} {
		add(CardLeased, s)
	}
	for _, s := range append(append([]string{CardPicked, CardAssigned}, leaseHoldingStates...), CardMergedLocal, CardPushed, CardCIGreen) {
		add(s, CardNeedsDecision)
	}
	for _, s := range []string{CardAssigned, CardPicked, CardMergedLocal, CardPushed, CardCIGreen} {
		add(CardNeedsDecision, s)
	}
	add(CardNeedsDecision, CardBlocked)
	add(CardBlocked, CardAssigned)
	for _, s := range cardStates {
		if !IsTerminalCardState(s) {
			add(s, CardAbandoned)
		}
	}
	for _, s := range leaseHoldingStates {
		add(s, CardFailed)
	}
	return out
}

// frFixtureCard places a card in `from` with every field a guard could read set
// to a value that satisfies it for the target `to`.
func frFixtureCard(repo frRepo, cardID, from, to string) Card {
	c := Card{RunID: frRun, CardID: cardID, State: from, OwnerLabel: "worker-1", WorktreePath: repo.Dir, EvidenceSHA: repo.Commit}
	if IsLeaseHoldingState(from) {
		c.LeaseHolder = "worker-1"
		c.LeaseExpiresAt = frLeaseUntil(time.Hour)
		c.HeartbeatAt = frNow.Format(time.RFC3339Nano)
	}
	if from == CardLeased && isResumableStage(to) {
		c.Stage = to
	}
	switch from {
	case CardKickoff:
		c.DecisionGate = DecisionGateKickoff
	case CardNeedsDecision:
		c.DecisionGate = DecisionGateQuestion
		c.DecisionQuestion = "which way?"
		c.DecisionResume = CardRun
		if slices.Contains([]string{CardAssigned, CardPicked, CardMergedLocal, CardPushed, CardCIGreen}, to) {
			c.DecisionResume = to
		}
	case CardMergedLocal, CardPushed, CardCIGreen:
		c.MergeSHA = repo.Merge
		c.MergeTree = repo.MergeTree
	}
	return c
}

func frFullRequest(repo frRepo, c Card, to string) TransitionRequest {
	return TransitionRequest{
		RunID: c.RunID, CardID: c.CardID, To: to, ExpectedVersion: c.Version,
		Actor: "worker-1", Decider: DeciderHuman, Owner: "worker-1",
		SHA: repo.Commit, ArtifactPath: repo.Artifact,
		MergeSHA: repo.Merge, RemeasurePath: repo.Remeasure, IntegrationBranch: repo.Integration,
		Question: "which way?", Reason: "build broken", Now: frNow,
	}
}

func frWriteVerdictsFor(t *testing.T, repo frRepo, cardID, from string) {
	t.Helper()
	switch from {
	case CardPlanAudit:
		frWriteVerdict(t, repo.Dir, cardID, "plan-audit.md", "PASS", repo.Commit)
	case CardSyncAudit:
		frWriteVerdict(t, repo.Dir, cardID, "sync-audit.md", "PASS", repo.Commit)
	}
}

// AC-005 — exactly the 65 requested edges are accepted and the other 296 are
// refused; the production table has no duplicate pair; the T4 stage guard
// refuses a mismatched resume.
func TestFR_AC005_TransitionTableEdgeCount(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	frRegisterWorker(t, db, "worker-1")
	ctx := context.Background()

	seen := map[[2]string]string{}
	for _, e := range TransitionEdges() {
		key := [2]string{e.From, e.To}
		if prev, dup := seen[key]; dup {
			t.Fatalf("transition table carries duplicate pair %v (%s and %s)", key, prev, e.ID)
		}
		seen[key] = e.ID
	}

	expected := frExpectedEdges()
	if len(expected) != 65 {
		t.Fatalf("expected-edge fixture lists %d pairs, want 65", len(expected))
	}
	accepted := map[[2]string]bool{}
	refused := 0
	for i, from := range cardStates {
		for j, to := range cardStates {
			cardID := fmt.Sprintf("e%02d-%02d", i, j)
			c := frFixtureCard(repo, cardID, from, to)
			frPlace(t, db, c)
			frWriteVerdictsFor(t, repo, cardID, from)
			before := frRowDump(t, db, frRun, cardID)
			if _, err := db.Transition(ctx, frFullRequest(repo, Card{RunID: frRun, CardID: cardID, Version: 1}, to)); err == nil {
				accepted[[2]string{from, to}] = true
				continue
			}
			refused++
			if after := frRowDump(t, db, frRun, cardID); after != before {
				t.Fatalf("refused %s → %s changed the record:\nbefore %s\nafter  %s", from, to, before, after)
			}
		}
	}
	t.Logf("requested pairs: %d accepted, %d refused, %d total; production table rows: %d", len(accepted), refused, len(accepted)+refused, len(TransitionEdges()))
	if len(accepted) != 65 || refused != 296 {
		t.Fatalf("accepted %d / refused %d, want 65 / 296", len(accepted), refused)
	}
	for pair := range expected {
		if !accepted[pair] {
			t.Errorf("expected edge %s → %s was refused", pair[0], pair[1])
		}
	}
	for pair := range accepted {
		if !expected[pair] {
			t.Errorf("unexpected edge %s → %s was accepted", pair[0], pair[1])
		}
	}
	for _, pair := range [][2]string{{CardMergedLocal, CardDone}, {CardPushed, CardCIGreen}, {CardCIGreen, CardDone}} {
		if accepted[pair] {
			t.Errorf("%s → %s (T18/T19/T20) accepted, want refused", pair[0], pair[1])
		}
	}

	// T4 stage guard: a leased card whose stage is sync resumes only into sync.
	for _, to := range []string{CardPlan, CardRun, CardSync} {
		cardID := "stage-sync-" + to
		c := frFixtureCard(repo, cardID, CardLeased, CardSync)
		frPlace(t, db, c)
		_, err := db.Transition(ctx, frFullRequest(repo, Card{RunID: frRun, CardID: cardID, Version: 1}, to))
		if to == CardSync && err != nil {
			t.Fatalf("leased(stage=sync) → sync refused: %v", err)
		}
		if to != CardSync && err == nil {
			t.Fatalf("leased(stage=sync) → %s accepted, want guard refusal", to)
		}
	}
}

// AC-002 — a legacy row admits only an operator abandon.
func TestFR_AC002_LegacyStateRefusesAllButAbandon(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	frPlace(t, db, Card{RunID: frRun, CardID: "legacy", State: "in_progress", Version: 3, OwnerLabel: "worker-1", WorktreePath: repo.Dir})
	before := frRowDump(t, db, frRun, "legacy")
	for _, to := range cardStates {
		if to == CardAbandoned {
			continue
		}
		_, err := db.Transition(ctx, frFullRequest(repo, Card{RunID: frRun, CardID: "legacy", Version: 3}, to))
		if !errors.Is(err, ErrLegacyState) {
			t.Fatalf("legacy → %s err = %v, want ErrLegacyState", to, err)
		}
		if after := frRowDump(t, db, frRun, "legacy"); after != before {
			t.Fatalf("legacy → %s changed the record", to)
		}
	}
	got, err := db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: "legacy", To: CardAbandoned, ExpectedVersion: 3, Actor: "operator", Decider: DeciderHuman, Now: frNow})
	if err != nil {
		t.Fatalf("legacy abandon: %v", err)
	}
	if got.State != CardAbandoned || got.Version != 4 {
		t.Fatalf("legacy abandon = %s v%d, want abandoned v4", got.State, got.Version)
	}
}

// AC-003 — one transaction: state, version+1, and exactly one card.transition
// event commit together or not at all.
func TestFR_AC003_TransitionIsAtomic(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	frRegisterWorker(t, db, "worker-1")
	ctx := context.Background()
	c := frFixtureCard(repo, "atomic", CardPlan, CardPlanAudit)
	c.Version = 7
	frPlace(t, db, c)

	// Fault after the state update, before the event insert: nothing commits.
	before := frRowDump(t, db, frRun, "atomic")
	cardTransitionFault = func(stage string) error {
		if stage == "before-event" {
			return errors.New("injected fault")
		}
		return nil
	}
	_, err := db.Transition(ctx, frFullRequest(repo, c, CardPlanAudit))
	cardTransitionFault = nil
	if err == nil || !strings.Contains(err.Error(), "injected fault") {
		t.Fatalf("faulted transition err = %v, want the injected fault", err)
	}
	if after := frRowDump(t, db, frRun, "atomic"); after != before {
		t.Fatalf("faulted transition committed a partial write:\nbefore %s\nafter  %s", before, after)
	}

	got, err := db.Transition(ctx, frFullRequest(repo, c, CardPlanAudit))
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if got.State != CardPlanAudit || got.Version != 8 {
		t.Fatalf("card = %s v%d, want plan-audit v8", got.State, got.Version)
	}
	events := frEvents(t, db, "card.transition")
	if len(events) != 1 {
		t.Fatalf("card.transition events = %d, want 1", len(events))
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(events[0]), &payload); err != nil {
		t.Fatalf("payload: %v", err)
	}
	for k, want := range map[string]any{"card_id": "atomic", "from": CardPlan, "to": CardPlanAudit, "version": float64(8), "actor": "worker-1"} {
		if payload[k] != want {
			t.Fatalf("payload[%s] = %v, want %v (payload %s)", k, payload[k], want, events[0])
		}
	}
}

// AC-004 — an expected version behind the stored one is refused unchanged.
func TestFR_AC004_StaleVersionRefused(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	c := frFixtureCard(repo, "stale", CardPicked, CardAssigned)
	c.Version = 5
	frPlace(t, db, c)
	before := frRowDump(t, db, frRun, "stale")
	req := frFullRequest(repo, c, CardAssigned)
	req.ExpectedVersion = 4
	if _, err := db.Transition(ctx, req); !errors.Is(err, ErrStaleVersion) {
		t.Fatalf("stale transition err = %v, want ErrStaleVersion", err)
	}
	if after := frRowDump(t, db, frRun, "stale"); after != before {
		t.Fatalf("stale transition changed the record")
	}
}

// AC-006 — two writers racing on one version: exactly one wins, 50 times.
func TestFR_AC006_RacingWritersExactlyOneWins(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	other, err := OpenFactoryPath(db.Path)
	if err != nil {
		t.Fatalf("second connection: %v", err)
	}
	t.Cleanup(func() { _ = other.Close() })
	ctx := context.Background()
	for i := 0; i < 50; i++ {
		cardID := fmt.Sprintf("race-%02d", i)
		c := frFixtureCard(repo, cardID, CardPicked, CardAssigned)
		c.Version = 1
		frPlace(t, db, c)
		var wg sync.WaitGroup
		errs := make([]error, 2)
		for k, store := range []*FactoryDB{db, other} {
			wg.Add(1)
			go func(k int, store *FactoryDB) {
				defer wg.Done()
				_, errs[k] = store.Transition(ctx, frFullRequest(repo, c, CardAssigned))
			}(k, store)
		}
		wg.Wait()
		wins := 0
		for _, err := range errs {
			switch {
			case err == nil:
				wins++
			case !errors.Is(err, ErrStaleVersion):
				t.Fatalf("iteration %d: loser err = %v, want ErrStaleVersion", i, err)
			}
		}
		if wins != 1 {
			t.Fatalf("iteration %d: %d writers won, want exactly 1 (errs %v)", i, wins, errs)
		}
		got, err := db.LoadCard(ctx, frRun, cardID)
		if err != nil || got.Version != 2 {
			t.Fatalf("iteration %d: card version = %d err=%v, want 2", i, got.Version, err)
		}
	}
	if n := len(frEvents(t, db, "card.transition")); n != 50 {
		t.Fatalf("card.transition events = %d, want 50 (one per race)", n)
	}
}

// AC-019 — the CI edges are reserved for F3 and refused, even with a
// caller-supplied file claiming a green CI run.
func TestFR_AC019_ReservedCIEdgesRefused(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	claim := filepath.Join(t.TempDir(), "ci-green.json")
	frWrite(t, claim, `{"merge_sha":"`+repo.Merge+`","conclusion":"success"}`)
	for _, pair := range [][2]string{{CardPushed, CardCIGreen}, {CardCIGreen, CardDone}} {
		cardID := "reserved-" + pair[0]
		c := frFixtureCard(repo, cardID, pair[0], pair[1])
		c.Version = 1
		frPlace(t, db, c)
		before := frRowDump(t, db, frRun, cardID)
		for _, withClaim := range []bool{false, true} {
			req := frFullRequest(repo, c, pair[1])
			if withClaim {
				req.ArtifactPath = claim
				req.RemeasurePath = claim
			}
			_, err := db.Transition(ctx, req)
			if !errors.Is(err, ErrReservedEdge) || !strings.Contains(err.Error(), "F3") {
				t.Fatalf("%s → %s (claim=%v) err = %v, want ErrReservedEdge naming F3", pair[0], pair[1], withClaim, err)
			}
			if after := frRowDump(t, db, frRun, cardID); after != before {
				t.Fatalf("%s → %s changed the record", pair[0], pair[1])
			}
		}
	}
}
