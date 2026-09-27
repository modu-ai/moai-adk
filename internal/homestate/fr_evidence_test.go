package homestate

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func frLeasedCard(repo frRepo, cardID, state string) Card {
	return Card{
		RunID: frRun, CardID: cardID, State: state, Version: 1, OwnerLabel: "worker-1",
		LeaseHolder: "worker-1", LeaseExpiresAt: frLeaseUntil(time.Hour), HeartbeatAt: frNow.Format(time.RFC3339Nano),
		WorktreePath: repo.Dir, Stage: state,
	}
}

func frHolderRequest(c Card, to string) TransitionRequest {
	return TransitionRequest{RunID: c.RunID, CardID: c.CardID, To: to, ExpectedVersion: c.Version, Actor: "worker-1", Now: frNow}
}

// AC-007 — audit entry verifies the SHA and artifact itself; the request type
// carries no field through which a caller could assert the evidence passed.
func TestFR_AC007_AuditEntryEvidence(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	// A commit that exists but is not an ancestor of the worktree HEAD.
	frGit(t, repo.Dir, "checkout", "-q", "-b", "side", repo.Commit)
	frWrite(t, filepath.Join(repo.Dir, "side.txt"), "side\n")
	frGit(t, repo.Dir, "add", "-A")
	frGit(t, repo.Dir, "commit", "-q", "-m", "S")
	side := frGit(t, repo.Dir, "rev-parse", "HEAD")
	frGit(t, repo.Dir, "checkout", "-q", repo.Integration)

	c := frLeasedCard(repo, "entry", CardPlan)
	frPlace(t, db, c)
	before := frRowDump(t, db, frRun, "entry")
	for name, ev := range map[string][2]string{
		"missing sha":      {strings.Repeat("d", 40), repo.Artifact},
		"not an ancestor":  {side, "side.txt"},
		"artifact missing": {repo.Commit, "b.txt"},
	} {
		req := frHolderRequest(c, CardPlanAudit)
		req.SHA, req.ArtifactPath = ev[0], ev[1]
		if _, err := db.Transition(ctx, req); !errors.Is(err, ErrEvidence) {
			t.Fatalf("%s: err = %v, want ErrEvidence", name, err)
		}
		if after := frRowDump(t, db, frRun, "entry"); after != before {
			t.Fatalf("%s: refused entry changed the record", name)
		}
	}
	req := frHolderRequest(c, CardPlanAudit)
	req.SHA, req.ArtifactPath = repo.Commit, repo.Artifact
	got, err := db.Transition(ctx, req)
	if err != nil {
		t.Fatalf("valid entry: %v", err)
	}
	if got.State != CardPlanAudit || got.EvidenceSHA != repo.Commit || got.EvidencePath != repo.Artifact {
		t.Fatalf("entered card = %s sha=%s path=%s", got.State, got.EvidenceSHA, got.EvidencePath)
	}

	// REQ-FR-010: no caller-supplied verdict, pass flag, tree hash, or ancestry claim.
	rt := reflect.TypeOf(TransitionRequest{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Type.Kind() == reflect.Bool {
			t.Fatalf("TransitionRequest.%s is a boolean — a caller could assert a pass", f.Name)
		}
		for _, banned := range []string{"Verdict", "Pass", "Tree", "Ancestor", "Passed", "Green"} {
			if strings.Contains(f.Name, banned) {
				t.Fatalf("TransitionRequest.%s lets a caller assert evidence", f.Name)
			}
		}
	}
}

// AC-008 — the audit exit reads the verdict file's machine lines and binds
// them to the SHA recorded at entry.
func TestFR_AC008_AuditVerdictGate(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	place := func(cardID string) Card {
		c := frLeasedCard(repo, cardID, CardPlanAudit)
		c.EvidenceSHA = repo.Commit
		c.EvidencePath = repo.Artifact
		frPlace(t, db, c)
		return c
	}
	for name, body := range map[string]string{
		"absent":       "",
		"no verdict":   "audited_sha: " + repo.Commit + "\n",
		"fail":         "verdict: FAIL\naudited_sha: " + repo.Commit + "\n",
		"other commit": "verdict: PASS\naudited_sha: " + repo.Merge + "\n",
		"conflicting":  "verdict: PASS\nverdict: FAIL\naudited_sha: " + repo.Commit + "\n",
	} {
		cardID := "v-" + strings.ReplaceAll(name, " ", "-")
		c := place(cardID)
		if body != "" {
			frWrite(t, filepath.Join(repo.Dir, ".moai", "reports", cardID, "plan-audit.md"), "# report\n\n"+body)
		}
		before := frRowDump(t, db, frRun, cardID)
		if _, err := db.Transition(ctx, frHolderRequest(c, CardKickoff)); !errors.Is(err, ErrEvidence) {
			t.Fatalf("%s: err = %v, want ErrEvidence", name, err)
		}
		if after := frRowDump(t, db, frRun, cardID); after != before {
			t.Fatalf("%s: refused exit changed the record", name)
		}
		if name == "fail" {
			back, err := db.Transition(ctx, frHolderRequest(c, CardPlan))
			if err != nil || back.State != CardPlan {
				t.Fatalf("FAIL verdict return edge: state=%s err=%v, want plan", back.State, err)
			}
		}
	}
	c := place("v-pass-with-debt")
	frWriteVerdict(t, repo.Dir, "v-pass-with-debt", "plan-audit-iter2.md", "PASS-WITH-DEBT", repo.Commit)
	got, err := db.Transition(ctx, frHolderRequest(c, CardKickoff))
	if err != nil {
		t.Fatalf("PASS-WITH-DEBT exit: %v", err)
	}
	if got.State != CardKickoff || got.DecisionGate != DecisionGateKickoff {
		t.Fatalf("exit = %s gate=%q, want kickoff / kickoff", got.State, got.DecisionGate)
	}
}

// AC-009 — merged-local requires a real two-parent merge whose tree equals its
// second parent's, reachable from the integration branch, and a re-measure
// file naming it.
func TestFR_AC009_MergeEvidence(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	// A merge whose tree differs from its second parent (the base had moved).
	frGit(t, repo.Dir, "checkout", "-q", "-b", "div", repo.Commit)
	frWrite(t, filepath.Join(repo.Dir, "d.txt"), "d\n")
	frGit(t, repo.Dir, "add", "-A")
	frGit(t, repo.Dir, "commit", "-q", "-m", "D0")
	frGit(t, repo.Dir, "merge", "-q", "--no-ff", "-m", "D", "feature")
	divergent := frGit(t, repo.Dir, "rev-parse", "HEAD")
	// A clean merge (tree == second parent) not reachable from integration.
	frGit(t, repo.Dir, "checkout", "-q", "-b", "other", repo.Commit)
	frGit(t, repo.Dir, "merge", "-q", "--no-ff", "-m", "O", "feature")
	unreachable := frGit(t, repo.Dir, "rev-parse", "HEAD")
	frGit(t, repo.Dir, "checkout", "-q", repo.Integration)
	stale := filepath.Join(t.TempDir(), "stale-remeasure.md")
	frWrite(t, stale, "re-measured something else\n")
	nameFile := func(sha string) string {
		p := filepath.Join(t.TempDir(), "remeasure-"+sha[:8]+".md")
		frWrite(t, p, "merge "+sha+"\n")
		return p
	}

	c := frLeasedCard(repo, "merge", CardMerging)
	c.Stage = CardMergeReady
	frPlace(t, db, c)
	before := frRowDump(t, db, frRun, "merge")
	for name, in := range map[string][3]string{
		"single parent":      {repo.Commit, nameFile(repo.Commit), repo.Integration},
		"tree differs":       {divergent, nameFile(divergent), "div"},
		"not on integration": {unreachable, nameFile(unreachable), repo.Integration},
		"remeasure stale":    {repo.Merge, stale, repo.Integration},
	} {
		req := frHolderRequest(c, CardMergedLocal)
		req.MergeSHA, req.RemeasurePath, req.IntegrationBranch = in[0], in[1], in[2]
		if _, err := db.Transition(ctx, req); !errors.Is(err, ErrEvidence) {
			t.Fatalf("%s: err = %v, want ErrEvidence", name, err)
		}
		if after := frRowDump(t, db, frRun, "merge"); after != before {
			t.Fatalf("%s: refused merge changed the record", name)
		}
	}
	req := frHolderRequest(c, CardMergedLocal)
	req.MergeSHA, req.RemeasurePath, req.IntegrationBranch = repo.Merge, repo.Remeasure, repo.Integration
	got, err := db.Transition(ctx, req)
	if err != nil {
		t.Fatalf("valid merge: %v", err)
	}
	if got.State != CardMergedLocal || got.MergeSHA != repo.Merge || got.MergeTree != repo.MergeTree || got.RemeasurePath != repo.Remeasure {
		t.Fatalf("merged card = %s sha=%s tree=%s remeasure=%s", got.State, got.MergeSHA, got.MergeTree, got.RemeasurePath)
	}
}

// AC-018 (store half) — the push gate checks the remote-tracking ref as it
// stands and never fetches; with no remote the card goes straight to done.
func TestFR_AC018_PushGateStore(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	// A local merge the remote-tracking ref does not yet contain.
	frGit(t, repo.Dir, "checkout", "-q", "-b", "feature2", repo.Commit)
	frWrite(t, filepath.Join(repo.Dir, "e.txt"), "e\n")
	frGit(t, repo.Dir, "add", "-A")
	frGit(t, repo.Dir, "commit", "-q", "-m", "F2")
	frGit(t, repo.Dir, "checkout", "-q", repo.Integration)
	frGit(t, repo.Dir, "merge", "-q", "--no-ff", "-m", "M2", "feature2")
	local := frGit(t, repo.Dir, "rev-parse", "HEAD")

	c := Card{RunID: frRun, CardID: "push", State: CardMergedLocal, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir, MergeSHA: local}
	frPlace(t, db, c)
	decide := func(c Card, to string) (Card, error) {
		return db.Transition(ctx, TransitionRequest{RunID: frRun, CardID: c.CardID, To: to, ExpectedVersion: c.Version, Actor: "lead", Decider: DeciderHuman, IntegrationBranch: repo.Integration, Now: frNow})
	}
	before := frRowDump(t, db, frRun, "push")
	if _, err := decide(c, CardPushed); !errors.Is(err, ErrEvidence) {
		t.Fatalf("push before the ref contains the merge: err = %v, want ErrEvidence", err)
	}
	if _, err := decide(c, CardDone); err == nil {
		t.Fatal("merged-local → done accepted while a remote is configured")
	}
	if after := frRowDump(t, db, frRun, "push"); after != before {
		t.Fatal("refused push changed the record")
	}
	frGit(t, repo.Dir, "push", "-q", "origin", repo.Integration)
	got, err := decide(c, CardPushed)
	if err != nil || got.State != CardPushed {
		t.Fatalf("push after the ref advanced: state=%s err=%v", got.State, err)
	}

	bare := frNewRepo(t, false)
	n := Card{RunID: frRun, CardID: "noremote", State: CardMergedLocal, Version: 1, OwnerLabel: "worker-1", WorktreePath: bare.Dir, MergeSHA: bare.Merge}
	frPlace(t, db, n)
	if target, err := PushGateTarget(ctx, n); err != nil || target != CardDone {
		t.Fatalf("PushGateTarget(no remote) = %q err=%v, want done", target, err)
	}
	done, err := decide(n, CardDone)
	if err != nil || done.State != CardDone {
		t.Fatalf("no-remote push gate: state=%s err=%v", done.State, err)
	}
	events := frEvents(t, db, "card.transition")
	var payload map[string]any
	if err := json.Unmarshal([]byte(events[len(events)-1]), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["note"] != "no remote — no CI verdict" {
		t.Fatalf("no-remote event note = %v, want %q", payload["note"], "no remote — no CI verdict")
	}
}
