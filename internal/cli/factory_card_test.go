package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// The F1 factory-record CLI tests. Every test builds its own project root
// under t.TempDir() (todoFixture) with MOAI_HOME cleared, so factory.db and
// the queue live under that root, never in the real ~/.moai.

const fcRun = "run-cli"

var fcNow = time.Date(2026, 9, 26, 9, 0, 0, 0, time.UTC)

// fcFixture is a project root with an isolated git config and a queue.
func fcFixture(t *testing.T) (string, *kanban.BacklogStore) {
	t.Helper()
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(cfg, []byte("[user]\n\tname = F1 Test\n\temail = f1@example.invalid\n[init]\n\tdefaultBranch = main\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", cfg)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root, store := todoFixture(t)
	prev := factoryCardNow
	factoryCardNow = func() time.Time { return fcNow }
	t.Cleanup(func() { factoryCardNow = prev })
	return root, store
}

// fcQueue adds n queue items (t1..tn) and sets each item's state.
func fcQueue(t *testing.T, store *kanban.BacklogStore, states ...kanban.BacklogState) {
	t.Helper()
	for i := range states {
		if _, _, err := runTodo(t, "add", fmt.Sprintf("factory card %d", i+1)); err != nil {
			t.Fatalf("todo add: %v", err)
		}
	}
	if err := store.Mutate(func(r *kanban.BacklogRecord) error {
		for i := range r.Items {
			if i < len(states) {
				r.Items[i].State = states[i]
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("set queue states: %v", err)
	}
}

func runFactory(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newFactoryCommand()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

func fcOpen(t *testing.T, root string) *homestate.FactoryDB {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// fcPlace inserts a card row directly (a fixture placement) and closes the
// connection, so the command under test opens the database fresh.
func fcPlace(t *testing.T, root string, rows ...homestate.Card) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	for _, c := range rows {
		if c.RunID == "" {
			c.RunID = fcRun
		}
		if c.Version == 0 {
			c.Version = 1
		}
		if c.UpdatedAt == "" {
			c.UpdatedAt = "2026-09-26T00:00:00Z"
		}
		if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,owner_label,state,version,evidence_path,updated_at,stage,lease_holder,lease_expires_at,heartbeat_at,decision_gate,decision_question,decision_resume,merge_sha,worktree_path,evidence_sha,hint_after) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			c.RunID, c.CardID, c.OwnerLabel, c.State, c.Version, c.EvidencePath, c.UpdatedAt, c.Stage, c.LeaseHolder, c.LeaseExpiresAt, c.HeartbeatAt, c.DecisionGate, c.DecisionQuestion, c.DecisionResume, c.MergeSHA, c.WorktreePath, c.EvidenceSHA, c.HintAfter); err != nil {
			t.Fatalf("place %s: %v", c.CardID, err)
		}
	}
}

func fcCard(t *testing.T, root, cardID string) homestate.Card {
	t.Helper()
	db := fcOpen(t, root)
	c, err := db.LoadCard(context.Background(), fcRun, cardID)
	if err != nil {
		t.Fatalf("load %s: %v", cardID, err)
	}
	_ = db.Close()
	return c
}

func fcHasCard(t *testing.T, root, cardID string) bool {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	_, err := db.LoadCard(context.Background(), fcRun, cardID)
	return err == nil
}

func fcEventCount(t *testing.T, root, kind string) int {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(`SELECT count(*) FROM events WHERE kind=?`, kind).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func fcGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// fcRepo builds a card repository with a merge commit on `integration`,
// optionally pushed to a bare origin. It returns the repo dir and the merge SHA.
func fcRepo(t *testing.T, withRemote bool) (string, string) {
	t.Helper()
	base := t.TempDir()
	dir := filepath.Join(base, "repo")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	fcGit(t, dir, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, dir, "add", "-A")
	fcGit(t, dir, "commit", "-q", "-m", "C")
	fcGit(t, dir, "branch", "integration")
	fcGit(t, dir, "checkout", "-q", "-b", "feature")
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, dir, "add", "-A")
	fcGit(t, dir, "commit", "-q", "-m", "F")
	fcGit(t, dir, "checkout", "-q", "integration")
	fcGit(t, dir, "merge", "-q", "--no-ff", "-m", "M", "feature")
	merge := fcGit(t, dir, "rev-parse", "HEAD")
	if withRemote {
		bare := filepath.Join(base, "origin.git")
		fcGit(t, base, "init", "-q", "--bare", bare)
		fcGit(t, dir, "remote", "add", "origin", bare)
	}
	return dir, merge
}

// fcGitFlowConfig names `integration` as the project's git-flow integration branch.
func fcGitFlowConfig(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	body := "git_strategy:\n  mode: manual\n  manual:\n    workflow: git-flow\n    develop_branch: integration\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func fcFileHash(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func fcRows(t *testing.T, root string) string {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	var b strings.Builder
	for _, q := range []string{`SELECT * FROM cards ORDER BY run_id, card_id`, `SELECT seq,run_id,kind,payload_json,created_at FROM events ORDER BY seq`} {
		rows, err := db.DB.Query(q)
		if err != nil {
			t.Fatal(err)
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]sql.NullString, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatal(err)
			}
			for _, v := range vals {
				b.WriteString(v.String + "|")
			}
			b.WriteString("\n")
		}
		_ = rows.Close()
	}
	return b.String()
}

// AC-014 — the `after` hint gates assignment on the predecessor's merge; the
// `prefer` hint is stored and reported but refuses nothing.
func TestFR_AC014_AssignHints(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked, kanban.BacklogStatePicked, kanban.BacklogStatePicked)
	// t1 = c1 (predecessor), t2 = c2, t3 = c3.
	if _, _, err := runFactory(t, "assign", "t2", "--after", "t1", "--run", fcRun); err != nil {
		t.Fatalf("assign t2 --after t1: %v", err)
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardPicked || c.HintAfter != "t1" {
		t.Fatalf("t2 = %s after=%q, want picked / t1", c.State, c.HintAfter)
	}
	_, _, err := runFactory(t, "assign", "t2", "--to", "worker-1", "--run", fcRun)
	if err == nil || !strings.Contains(err.Error(), "unknown predecessor") {
		t.Fatalf("assign with no predecessor record: err = %v, want unknown-predecessor", err)
	}
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardSync, OwnerLabel: "worker-9"})
	_, _, err = runFactory(t, "assign", "t2", "--to", "worker-1", "--run", fcRun)
	if err == nil || !strings.Contains(err.Error(), "not merged") {
		t.Fatalf("assign with predecessor in sync: err = %v, want predecessor-unmerged", err)
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardPicked {
		t.Fatalf("t2 moved to %s on a refused assign", c.State)
	}
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`UPDATE cards SET state='merged-local' WHERE card_id='t1'`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	if _, _, err := runFactory(t, "assign", "t2", "--to", "worker-1", "--run", fcRun); err != nil {
		t.Fatalf("assign after predecessor merged: %v", err)
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardAssigned || c.OwnerLabel != "worker-1" {
		t.Fatalf("t2 = %s owner=%q, want assigned / worker-1", c.State, c.OwnerLabel)
	}

	if _, _, err := runFactory(t, "assign", "t3", "--prefer", "backend=codex", "--run", fcRun); err != nil {
		t.Fatalf("assign t3 --prefer: %v", err)
	}
	out, _, err := runFactory(t, "status", "--json", "--run", fcRun)
	if err != nil {
		t.Fatalf("status --json: %v", err)
	}
	var st struct {
		Cards []struct {
			CardID string `json:"card_id"`
			Prefer string `json:"prefer"`
		} `json:"cards"`
	}
	if err := json.Unmarshal([]byte(out), &st); err != nil {
		t.Fatalf("status json: %v\n%s", err, out)
	}
	found := false
	for _, c := range st.Cards {
		if c.CardID == "t3" {
			found = c.Prefer == "backend=codex"
		}
	}
	if !found {
		t.Fatalf("status --json does not show t3 prefer exactly backend=codex:\n%s", out)
	}
	if _, _, err := runFactory(t, "assign", "t3", "--to", "worker-1", "--run", fcRun); err != nil {
		t.Fatalf("prefer hint refused an assignment: %v", err)
	}
}

// AC-015 — a kickoff batch decides each card independently; a non-human
// decider is refused outright.
func TestFR_AC015_DecideKickoffBatch(t *testing.T) {
	root, _ := fcFixture(t)
	fcPlace(t, root,
		homestate.Card{CardID: "k1", State: homestate.CardKickoff, OwnerLabel: "worker-1", Stage: homestate.CardPlanAudit, DecisionGate: homestate.DecisionGateKickoff},
		homestate.Card{CardID: "k2", State: homestate.CardKickoff, OwnerLabel: "worker-1", Stage: homestate.CardPlanAudit, DecisionGate: homestate.DecisionGateKickoff},
		homestate.Card{CardID: "k3", State: homestate.CardKickoff, OwnerLabel: "worker-1", Stage: homestate.CardPlanAudit, DecisionGate: homestate.DecisionGateKickoff},
		homestate.Card{CardID: "r1", State: homestate.CardRun, OwnerLabel: "worker-1", Stage: homestate.CardRun},
	)
	r1Before, k3Before := fcCard(t, root, "r1"), fcCard(t, root, "k3")
	out, _, _ := runFactory(t, "decide", "k1", "k2", "r1", "--gate", "kickoff", "--choice", "approve", "--run", fcRun)
	for _, id := range []string{"k1", "k2"} {
		c := fcCard(t, root, id)
		if c.State != homestate.CardAssigned || c.Stage != homestate.CardRun || c.Decider != "human" {
			t.Fatalf("%s = %s stage=%s decider=%q, want assigned / run / human", id, c.State, c.Stage, c.Decider)
		}
	}
	if !strings.Contains(out, "r1") || !strings.Contains(out, "refused") {
		t.Fatalf("decide output does not report r1 refused:\n%s", out)
	}
	if fcCard(t, root, "r1") != r1Before || fcCard(t, root, "k3") != k3Before {
		t.Fatal("r1 or k3 changed")
	}
	if _, _, err := runFactory(t, "decide", "k3", "--gate", "kickoff", "--choice", "reject", "--run", fcRun); err != nil {
		t.Fatalf("reject k3: %v", err)
	}
	if c := fcCard(t, root, "k3"); c.State != homestate.CardBlocked {
		t.Fatalf("k3 = %s, want blocked", c.State)
	}
	before := fcRows(t, root)
	if _, _, err := runFactory(t, "decide", "k1", "--gate", "kickoff", "--choice", "approve", "--decider", "llm", "--run", fcRun); err == nil {
		t.Fatal("decide --decider llm was accepted")
	}
	if fcRows(t, root) != before {
		t.Fatal("decide --decider llm changed the record")
	}
}

// AC-016 — needs-decision records the question and resume target; resume,
// block, and abandon move the card as the table says.
func TestFR_AC016_DecideQuestionChoices(t *testing.T) {
	root, _ := fcFixture(t)
	fcPlace(t, root,
		homestate.Card{CardID: "q1", State: homestate.CardRun, OwnerLabel: "worker-1", Stage: homestate.CardRun, LeaseHolder: "worker-1", LeaseExpiresAt: fcNow.Add(time.Hour).Format(time.RFC3339Nano)},
		homestate.Card{CardID: "q2", State: homestate.CardNeedsDecision, OwnerLabel: "worker-1", DecisionGate: homestate.DecisionGateQuestion, DecisionQuestion: "push now?", DecisionResume: homestate.CardMergedLocal},
		homestate.Card{CardID: "q3", State: homestate.CardNeedsDecision, OwnerLabel: "worker-1", DecisionGate: homestate.DecisionGateQuestion, DecisionQuestion: "keep going?", DecisionResume: homestate.CardRun},
		homestate.Card{CardID: "plain", State: homestate.CardRun, OwnerLabel: "worker-1", Stage: homestate.CardRun, LeaseHolder: "worker-1", LeaseExpiresAt: fcNow.Add(time.Hour).Format(time.RFC3339Nano)},
	)
	db := fcOpen(t, root)
	nd, err := db.Transition(context.Background(), homestate.TransitionRequest{RunID: fcRun, CardID: "q1", To: homestate.CardNeedsDecision, ExpectedVersion: 1, Actor: "worker-1", Question: "which schema?", Now: fcNow})
	_ = db.Close()
	if err != nil {
		t.Fatalf("run → needs-decision: %v", err)
	}
	if nd.DecisionQuestion != "which schema?" || nd.DecisionResume != homestate.CardRun || nd.LeaseHolder != "" {
		t.Fatalf("needs-decision card question=%q resume=%q holder=%q", nd.DecisionQuestion, nd.DecisionResume, nd.LeaseHolder)
	}
	if _, _, err := runFactory(t, "decide", "q1", "--choice", "resume", "--run", fcRun); err != nil {
		t.Fatalf("resume q1: %v", err)
	}
	if c := fcCard(t, root, "q1"); c.State != homestate.CardAssigned || c.Stage != homestate.CardRun {
		t.Fatalf("q1 = %s stage=%s, want assigned / run", c.State, c.Stage)
	}
	if _, _, err := runFactory(t, "decide", "q2", "--choice", "resume", "--run", fcRun); err != nil {
		t.Fatalf("resume q2: %v", err)
	}
	if c := fcCard(t, root, "q2"); c.State != homestate.CardMergedLocal {
		t.Fatalf("q2 = %s, want merged-local", c.State)
	}
	if _, _, err := runFactory(t, "decide", "q3", "--choice", "block", "--run", fcRun); err != nil {
		t.Fatalf("block q3: %v", err)
	}
	if c := fcCard(t, root, "q3"); c.State != homestate.CardBlocked {
		t.Fatalf("q3 = %s, want blocked", c.State)
	}
	before := fcCard(t, root, "plain")
	if _, _, err := runFactory(t, "decide", "plain", "--choice", "abandon", "--run", fcRun); err != nil {
		t.Fatalf("abandon plain: %v", err)
	}
	after := fcCard(t, root, "plain")
	if after.State != homestate.CardAbandoned || after.Version != before.Version+1 {
		t.Fatalf("plain = %s v%d, want abandoned v%d", after.State, after.Version, before.Version+1)
	}
	after.State, after.Version, after.UpdatedAt = before.State, before.Version, before.UpdatedAt
	if after != before {
		t.Fatalf("abandon changed other columns:\nbefore %+v\nafter  %+v", before, after)
	}
}

// AC-002 (command half) — a legacy row is abandoned through decide.
func TestFR_AC002_LegacyAbandonViaDecide(t *testing.T) {
	root, _ := fcFixture(t)
	fcPlace(t, root, homestate.Card{CardID: "legacy", State: "in_progress", Version: 2, OwnerLabel: "worker-1"})
	if _, _, err := runFactory(t, "decide", "legacy", "--choice", "resume", "--run", fcRun); err == nil {
		t.Fatal("resume on a legacy row was accepted")
	}
	if _, _, err := runFactory(t, "decide", "legacy", "--choice", "abandon", "--run", fcRun); err != nil {
		t.Fatalf("abandon legacy: %v", err)
	}
	if c := fcCard(t, root, "legacy"); c.State != homestate.CardAbandoned || c.Version != 3 {
		t.Fatalf("legacy = %s v%d, want abandoned v3", c.State, c.Version)
	}
}

// AC-013 (command half) — a kickoff decision made long after the lease
// duration is accepted with no lease expiry.
func TestFR_AC013_DecideAfterClockAdvance(t *testing.T) {
	root, _ := fcFixture(t)
	fcPlace(t, root, homestate.Card{CardID: "kick", State: homestate.CardKickoff, OwnerLabel: "worker-1", Stage: homestate.CardPlanAudit, DecisionGate: homestate.DecisionGateKickoff})
	factoryCardNow = func() time.Time { return fcNow.Add(2 * homestate.FactoryLeaseDuration) }
	transitions := fcEventCount(t, root, "card.transition")
	if _, _, err := runFactory(t, "decide", "kick", "--gate", "kickoff", "--choice", "approve", "--run", fcRun); err != nil {
		t.Fatalf("approve: %v", err)
	}
	c := fcCard(t, root, "kick")
	if c.State != homestate.CardAssigned || c.Stage != homestate.CardRun || c.Decider != "human" {
		t.Fatalf("kick = %s stage=%s decider=%q", c.State, c.Stage, c.Decider)
	}
	if n := fcEventCount(t, root, "lease.expired"); n != 0 {
		t.Fatalf("lease.expired events = %d, want 0", n)
	}
	if n := fcEventCount(t, root, "card.transition") - transitions; n != 1 {
		t.Fatalf("card.transition events for the decision = %d, want 1", n)
	}
}

// AC-017 (command half) — unblock keeps the stage and refuses a card that is
// not blocked.
func TestFR_AC017_DecideUnblock(t *testing.T) {
	root, _ := fcFixture(t)
	fcPlace(t, root,
		homestate.Card{CardID: "b1", State: homestate.CardBlocked, OwnerLabel: "worker-1", Stage: homestate.CardSync},
		homestate.Card{CardID: "r1", State: homestate.CardRun, OwnerLabel: "worker-1", Stage: homestate.CardRun},
	)
	if _, _, err := runFactory(t, "decide", "b1", "--choice", "unblock", "--run", fcRun); err != nil {
		t.Fatalf("unblock b1: %v", err)
	}
	if c := fcCard(t, root, "b1"); c.State != homestate.CardAssigned || c.Stage != homestate.CardSync {
		t.Fatalf("b1 = %s stage=%s, want assigned / sync", c.State, c.Stage)
	}
	before := fcCard(t, root, "r1")
	if _, _, err := runFactory(t, "decide", "r1", "--choice", "unblock", "--run", fcRun); err == nil {
		t.Fatal("unblock on a card not in blocked was accepted")
	}
	if fcCard(t, root, "r1") != before {
		t.Fatal("refused unblock changed r1")
	}
}

// AC-018 (command half) — the push gate reads the remote-tracking ref and
// never fetches; a repository with no remote goes straight to done.
func TestFR_AC018_DecidePushGate(t *testing.T) {
	root, _ := fcFixture(t)
	fcGitFlowConfig(t, root)
	dir, merge := fcRepo(t, true)
	fcPlace(t, root, homestate.Card{CardID: "p1", State: homestate.CardMergedLocal, OwnerLabel: "worker-1", MergeSHA: merge, WorktreePath: dir})
	if _, _, err := runFactory(t, "decide", "p1", "--gate", "push", "--run", fcRun); err == nil {
		t.Fatal("push accepted before the remote-tracking ref contains the merge")
	}
	if c := fcCard(t, root, "p1"); c.State != homestate.CardMergedLocal {
		t.Fatalf("p1 = %s, want merged-local", c.State)
	}
	fcGit(t, dir, "push", "-q", "origin", "integration")
	if _, _, err := runFactory(t, "decide", "p1", "--gate", "push", "--run", fcRun); err != nil {
		t.Fatalf("push after the ref advanced: %v", err)
	}
	if c := fcCard(t, root, "p1"); c.State != homestate.CardPushed {
		t.Fatalf("p1 = %s, want pushed", c.State)
	}

	bare, bareMerge := fcRepo(t, false)
	fcPlace(t, root, homestate.Card{CardID: "p2", State: homestate.CardMergedLocal, OwnerLabel: "worker-1", MergeSHA: bareMerge, WorktreePath: bare})
	if _, _, err := runFactory(t, "decide", "p2", "--gate", "push", "--run", fcRun); err != nil {
		t.Fatalf("push gate with no remote: %v", err)
	}
	if c := fcCard(t, root, "p2"); c.State != homestate.CardDone {
		t.Fatalf("p2 = %s, want done", c.State)
	}
	db := fcOpen(t, root)
	var payload string
	if err := db.DB.QueryRow(`SELECT payload_json FROM events WHERE kind='card.transition' ORDER BY seq DESC LIMIT 1`).Scan(&payload); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	if !strings.Contains(payload, "no remote — no CI verdict") {
		t.Fatalf("no-remote event payload = %s", payload)
	}
}

// AC-021 — assign, status, and decide never alter the queue schema or write a
// queue item row.
func TestFR_AC021_QueueSchemaUntouched(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked, kanban.BacklogStatePicked)
	dbPath := filepath.Join(filepath.Dir(todoBacklogPath(root)), "backlog.db")
	snapshot := func() (string, string) {
		t.Helper()
		db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(dbPath)+"?mode=ro")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = db.Close() }()
		var schema, items strings.Builder
		rows, err := db.Query(`SELECT type,name,COALESCE(sql,'') FROM sqlite_master ORDER BY type,name`)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var a, b, c string
			_ = rows.Scan(&a, &b, &c)
			schema.WriteString(a + "|" + b + "|" + c + "\n")
		}
		_ = rows.Close()
		rows, err = db.Query(`SELECT * FROM items ORDER BY seq`)
		if err != nil {
			t.Fatal(err)
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]sql.NullString, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			_ = rows.Scan(ptrs...)
			for _, v := range vals {
				items.WriteString(v.String + "|")
			}
			items.WriteString("\n")
		}
		_ = rows.Close()
		return schema.String(), items.String()
	}
	schemaBefore, itemsBefore := snapshot()
	if out, _, err := runFactory(t, "assign", "t1", "--to", "worker-1", "--run", fcRun); err != nil || !strings.Contains(out, "t1") {
		t.Fatalf("assign: out=%q err=%v", out, err)
	}
	if _, _, err := runFactory(t, "status", "--run", fcRun); err != nil {
		t.Fatalf("status: %v", err)
	}
	fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardKickoff, OwnerLabel: "worker-1", DecisionGate: homestate.DecisionGateKickoff})
	if _, _, err := runFactory(t, "decide", "t2", "--gate", "kickoff", "--choice", "approve", "--run", fcRun); err != nil {
		t.Fatalf("decide: %v", err)
	}
	schemaAfter, itemsAfter := snapshot()
	if schemaAfter != schemaBefore {
		t.Fatalf("queue sqlite_master changed:\nbefore %s\nafter  %s", schemaBefore, schemaAfter)
	}
	if itemsAfter != itemsBefore {
		t.Fatalf("queue items rows changed:\nbefore %s\nafter  %s", itemsBefore, itemsAfter)
	}
}

// AC-022 — only a queue item in `picked` is admitted to the factory record.
func TestFR_AC022_AssignRequiresQueuePicked(t *testing.T) {
	root, store := fcFixture(t)
	// q1 = t1 queued, q2 = t2 dropped, q3 = t9 never issued, q4 = t3 picked.
	fcQueue(t, store, kanban.BacklogStateQueued, kanban.BacklogStateDropped, kanban.BacklogStatePicked)
	for id, want := range map[string]string{"t1": "queued", "t2": "dropped", "t9": "not in the queue"} {
		_, _, err := runFactory(t, "assign", id, "--run", fcRun)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("assign %s: err = %v, want one naming %q", id, err, want)
		}
		if fcHasCard(t, root, id) {
			t.Fatalf("refused assign %s created a factory record", id)
		}
	}
	if _, _, err := runFactory(t, "assign", "t3", "--run", fcRun); err != nil {
		t.Fatalf("assign t3: %v", err)
	}
	if c := fcCard(t, root, "t3"); c.State != homestate.CardPicked {
		t.Fatalf("t3 = %s, want picked", c.State)
	}
}

// AC-023 — status reports every card, marks an expired lease, carries the
// contract pointer, and writes nothing.
func TestFR_AC023_StatusIsReadOnly(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked, kanban.BacklogStatePicked)
	fcPlace(t, root,
		homestate.Card{CardID: "s-picked", State: homestate.CardPicked},
		homestate.Card{CardID: "s-leased", State: homestate.CardLeased, OwnerLabel: "worker-1", LeaseHolder: "worker-1", LeaseExpiresAt: fcNow.Add(-time.Minute).Format(time.RFC3339Nano)},
		homestate.Card{CardID: "s-question", State: homestate.CardNeedsDecision, OwnerLabel: "worker-1", DecisionGate: homestate.DecisionGateQuestion, DecisionQuestion: "which?", DecisionResume: homestate.CardRun},
	)
	digest, event := strings.Repeat("a", 64), strings.Repeat("b", 64)
	ref := "SPEC-EXAMPLE-001," + digest + ",2026-09-26T09:00:00Z," + event
	if _, _, err := runFactory(t, "assign", "t1", "--contract-ref", ref, "--to", "worker-1", "--run", fcRun); err != nil {
		t.Fatalf("assign with contract ref: %v", err)
	}
	for _, bad := range []string{
		"SPEC-EXAMPLE-001," + strings.Repeat("a", 63) + ",2026-09-26T09:00:00Z," + event,
		"SPEC-EXAMPLE-001," + digest + ",2026-09-26T09:00:00Z," + strings.Repeat("z", 64),
	} {
		if _, _, err := runFactory(t, "assign", "t2", "--contract-ref", bad, "--run", fcRun); err == nil {
			t.Fatalf("malformed contract ref %q accepted", bad)
		}
	}
	dbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	hashBefore, rowsBefore := fcFileHash(t, dbPath), fcRows(t, root)
	text, _, err := runFactory(t, "status", "--run", fcRun)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	jsonOut, _, err := runFactory(t, "status", "--json", "--run", fcRun)
	if err != nil {
		t.Fatalf("status --json: %v", err)
	}
	for _, id := range []string{"s-picked", "s-leased", "s-question", "t1"} {
		if !strings.Contains(text, id) {
			t.Fatalf("status text misses %s:\n%s", id, text)
		}
	}
	for _, want := range []string{"expired", "needs-decision", "question", "SPEC-EXAMPLE-001", digest, event} {
		if !strings.Contains(text, want) {
			t.Fatalf("status text misses %q:\n%s", want, text)
		}
	}
	var st struct {
		Cards []struct {
			CardID       string `json:"card_id"`
			State        string `json:"state"`
			Stage        string `json:"stage"`
			Version      int64  `json:"version"`
			Owner        string `json:"owner"`
			LeaseExpired bool   `json:"lease_expired"`
			Gate         string `json:"decision_gate"`
			Prefer       string `json:"prefer"`
			After        string `json:"after"`
			Contract     *struct {
				SpecID   string `json:"spec_id"`
				SHA256   string `json:"sha256"`
				SignedAt string `json:"signed_at"`
				Event    string `json:"event"`
			} `json:"contract"`
		} `json:"cards"`
	}
	if err := json.Unmarshal([]byte(jsonOut), &st); err != nil {
		t.Fatalf("status json: %v\n%s", err, jsonOut)
	}
	if len(st.Cards) != 4 {
		t.Fatalf("status json cards = %d, want 4:\n%s", len(st.Cards), jsonOut)
	}
	for _, c := range st.Cards {
		switch c.CardID {
		case "s-leased":
			if !c.LeaseExpired || c.State != homestate.CardLeased {
				t.Fatalf("s-leased json = %+v, want state leased with lease_expired", c)
			}
		case "t1":
			if c.Contract == nil || c.Contract.SpecID != "SPEC-EXAMPLE-001" || c.Contract.SHA256 != digest || c.Contract.SignedAt != "2026-09-26T09:00:00Z" || c.Contract.Event != event {
				t.Fatalf("t1 contract = %+v", c.Contract)
			}
		}
	}
	if fcFileHash(t, dbPath) != hashBefore || fcRows(t, root) != rowsBefore {
		t.Fatal("status wrote to the factory database")
	}
	if c := fcCard(t, root, "s-leased"); c.State != homestate.CardLeased {
		t.Fatalf("s-leased = %s after status, want leased (status applies no expiry)", c.State)
	}
}
