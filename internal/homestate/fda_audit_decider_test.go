package homestate

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// fdaKickoffFixture places a kickoff card owned by worker-1 whose plan
// artifacts, verdict, progress record, and decision index satisfy every
// T8a condition; mutate adjusts the fixture before the decision runs.
type fdaKickoff struct {
	db   *FactoryDB
	repo frRepo
	card Card
}

const fdaProgressReady = "# progress\n\n## §E.1 Plan-phase Audit-Ready Signal\n\n```yaml\naudit_ready: true\n```\n\n## §E.2 Run-phase Evidence\n\n_<pending run-phase>_\n"

func newFDAKickoff(t *testing.T, id string) fdaKickoff {
	t.Helper()
	db := frOpen(t)
	repo := frNewRepo(t, true)
	frRegisterWorker(t, db, "worker-1")
	specDir := filepath.Join(repo.Dir, ".moai", "specs", frSpecID)
	frWrite(t, filepath.Join(specDir, "progress.md"), fdaProgressReady)
	frWrite(t, filepath.Join(specDir, "decision-index.md"),
		"### Q1: x?\nLabel: FOUNDER\nClass: implementation-level\nDefault: keep (rule: preserve current behavior)\nAlternate: change\nOperator verdict: DEFAULT-APPLIED 2026-10-03T00:00:00Z claude+lane\n")
	frWriteVerdict(t, repo.Dir, id, "plan-audit.md", "PASS", repo.Commit)
	c := Card{RunID: frRun, CardID: id, State: CardKickoff, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir,
		EvidenceSHA: repo.Commit, SpecID: frSpecID, DecisionGate: DecisionGateKickoff, DecisionResume: CardRun}
	frPlace(t, db, c)
	return fdaKickoff{db: db, repo: repo, card: c}
}

func (f fdaKickoff) approve(decider, queue string) (Card, error) {
	return f.db.Transition(context.Background(), TransitionRequest{RunID: frRun, CardID: f.card.CardID, To: CardRun,
		ExpectedVersion: f.card.Version, Actor: "worker-1", Decider: decider, QueueHold: queue, Now: frNow})
}

// AC-FDA-015 — an audit approval moves kickoff → run and leases the card to
// its record owner in the same transition.
func TestFDA_AuditDeciderLeasesToTheRecordOwner(t *testing.T) {
	f := newFDAKickoff(t, "ok")
	got, err := f.approve(DeciderAudit, QueueHoldClear)
	if err != nil {
		t.Fatalf("audit approval: %v", err)
	}
	if got.State != CardRun || got.LeaseHolder != "worker-1" || got.LeaseExpiresAt == "" || got.Decider != DeciderAudit {
		t.Fatalf("after audit approval: %+v", got)
	}
}

// AC-FDA-014 — every failed condition refuses and leaves the human path.
func TestFDA_AuditDeciderRefusals(t *testing.T) {
	specFile := func(f fdaKickoff, name string) string {
		return filepath.Join(f.repo.Dir, ".moai", "specs", frSpecID, name)
	}
	cases := map[string]struct {
		mutate func(t *testing.T, f fdaKickoff)
		queue  string
		want   string
	}{
		"hash mismatch (plan edited after audit)": {func(t *testing.T, f fdaKickoff) {
			frWrite(t, specFile(f, "plan.md"), "edited after the audit\n")
		}, QueueHoldClear, "plan_artifact_hash"},
		"decision-index reclassified after audit": {func(t *testing.T, f fdaKickoff) {
			frWrite(t, specFile(f, "decision-index.md"), "### Q1: x?\nLabel: FOUNDER\nClass: implementation-level\nDefault: a\nOperator verdict: DEFAULT-APPLIED now\n")
		}, QueueHoldClear, "plan_artifact_hash"},
		"FAIL verdict": {func(t *testing.T, f fdaKickoff) {
			frWriteVerdict(t, f.repo.Dir, f.card.CardID, "plan-audit.md", "FAIL", f.repo.Commit)
		}, QueueHoldClear, "verdict FAIL"},
		"audit-ready absent": {func(t *testing.T, f fdaKickoff) {
			frWrite(t, specFile(f, "progress.md"), "# progress\n\n## §E.1 Plan-phase Audit-Ready Signal\n\n_<pending plan-phase>_\n")
		}, QueueHoldClear, "audit-ready"},
		"queue item on hold": {nil, QueueHoldHeld, "hold"},
		"queue unreadable":   {nil, QueueHoldUnreadable, "unreadable"},
		"queue not read":     {nil, "", "unreadable"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFDAKickoff(t, "r")
			if c.mutate != nil {
				c.mutate(t, f)
			}
			_, err := f.approve(DeciderAudit, c.queue)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err=%v, want a refusal naming %q", err, c.want)
			}
			if got, herr := f.db.Transition(context.Background(), TransitionRequest{RunID: frRun, CardID: "r", To: CardAssigned,
				ExpectedVersion: f.card.Version, Actor: "operator", Decider: DeciderHuman, Now: frNow}); herr != nil || got.State != CardAssigned {
				t.Fatalf("human path after refusal: state=%s err=%v", got.State, herr)
			}
		})
	}
}

// Sync-audit F1/F2/F3 — gate bypasses closed at the decider.
func TestFDA_AuditDeciderClosesSyncAuditBypasses(t *testing.T) {
	cases := map[string]func(t *testing.T, f fdaKickoff){
		"NaN score": func(t *testing.T, f fdaKickoff) {
			p := filepath.Join(f.repo.Dir, ".moai", "reports", f.card.CardID, "plan-audit.md")
			frWrite(t, p, strings.Replace(frReadAbs(t, p), "Overall Score: 0.90", "Overall Score: NaN", 1))
		},
		"DECIDED row holding DEFAULT-APPLIED on product-level": func(t *testing.T, f fdaKickoff) {
			dir := filepath.Join(f.repo.Dir, ".moai", "specs", frSpecID)
			frWrite(t, filepath.Join(dir, "decision-index.md"), "### Q1: x?\nLabel: DECIDED\nClass: product-level\nOperator verdict: DEFAULT-APPLIED x\n")
			frWriteVerdict(t, f.repo.Dir, f.card.CardID, "plan-audit.md", "PASS", f.repo.Commit)
		},
		"audit_ready false": func(t *testing.T, f fdaKickoff) {
			frWrite(t, filepath.Join(f.repo.Dir, ".moai", "specs", frSpecID, "progress.md"),
				"## §E.1 Plan-phase Audit-Ready Signal\n\n```yaml\naudit_ready: false\n```\n")
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFDAKickoff(t, "b")
			mutate(t, f)
			if _, err := f.approve(DeciderAudit, QueueHoldClear); err == nil {
				t.Fatalf("%s: audit approval accepted, want refusal", name)
			}
		})
	}
}

// AC-FDA-014 — FOUNDER row fixtures, each written before the audit so the hash matches.
func TestFDA_AuditDeciderFounderRows(t *testing.T) {
	rows := map[string]struct {
		index string
		ok    bool
	}{
		"empty product-level":               {"Label: FOUNDER\nClass: product-level\nOperator verdict:\n", false},
		"empty implementation, no default":  {"Label: FOUNDER\nClass: implementation-level\nOperator verdict:\n", false},
		"DEFAULT-APPLIED on product-level":  {"Label: FOUNDER\nClass: product-level\nDefault: a\nOperator verdict: DEFAULT-APPLIED x\n", false},
		"DEFAULT-APPLIED on Class-less row": {"Label: FOUNDER\nDefault: a\nOperator verdict: DEFAULT-APPLIED x\n", false},
		"DEFAULT-APPLIED without Default":   {"Label: FOUNDER\nClass: implementation-level\nOperator verdict: DEFAULT-APPLIED x\n", false},
		"operator verdict on product-level": {"Label: FOUNDER\nClass: product-level\nOperator verdict: keep the CLI default\n", true},
	}
	for name, r := range rows {
		t.Run(name, func(t *testing.T) {
			db := frOpen(t)
			repo := frNewRepo(t, true)
			frRegisterWorker(t, db, "worker-1")
			specDir := filepath.Join(repo.Dir, ".moai", "specs", frSpecID)
			frWrite(t, filepath.Join(specDir, "progress.md"), fdaProgressReady)
			frWrite(t, filepath.Join(specDir, "decision-index.md"), "### Q1: x?\n"+r.index)
			frWriteVerdict(t, repo.Dir, "fr", "plan-audit.md", "PASS", repo.Commit)
			c := Card{RunID: frRun, CardID: "fr", State: CardKickoff, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir,
				EvidenceSHA: repo.Commit, SpecID: frSpecID, DecisionGate: DecisionGateKickoff, DecisionResume: CardRun}
			frPlace(t, db, c)
			_, err := db.Transition(context.Background(), TransitionRequest{RunID: frRun, CardID: "fr", To: CardRun,
				ExpectedVersion: 1, Actor: "worker-1", Decider: DeciderAudit, QueueHold: QueueHoldClear, Now: frNow})
			if (err == nil) != r.ok {
				t.Fatalf("err=%v, want ok=%v", err, r.ok)
			}
		})
	}
}

func TestFDA_DeciderVocabulary(t *testing.T) {
	f := newFDAKickoff(t, "v")
	if _, err := f.approve("foo", QueueHoldClear); err == nil || !errors.Is(err, ErrDecider) {
		t.Fatalf("decider foo: err=%v, want ErrDecider", err)
	}
	if _, err := f.approve(DeciderHuman, QueueHoldClear); err == nil {
		t.Fatalf("human decider on T8a accepted; human keeps T8 (kickoff → assigned)")
	}
	got, err := f.db.Transition(context.Background(), TransitionRequest{RunID: frRun, CardID: "v", To: CardAssigned,
		ExpectedVersion: 1, Actor: "operator", Decider: DeciderHuman, Now: frNow})
	if err != nil || got.State != CardAssigned {
		t.Fatalf("T8 human: state=%s err=%v", got.State, err)
	}
}
