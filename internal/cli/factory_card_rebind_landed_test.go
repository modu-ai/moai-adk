package cli

// factory_card_rebind_landed_test.go — card t1962: a card whose PR landed on
// main as a SQUASH commit has no ancestry link from its audited PR head to the
// worktree HEAD, so `factory rebind-evidence` refused it ("not an ancestor of
// the worktree HEAD"), the row's evidence_sha stayed empty, `factory approve`
// refused ("issuance requires the reviewed evidence hash") and `todo done`
// refused ("receipt absent"). The operator's `todo landed --sha` record names
// the squash commit; rebind-evidence accepts it when it is reachable from the
// landed ref, and an unlanded card stays refused.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// squashLandedRepo builds the CARD worktree of the squash shape: the audited
// PR head P lives only on branch `pr`, while the worktree HEAD is main at the
// base commit, so P is not an ancestor of it. The squash commit itself lives
// in the queue root's repository (squashLandedRoot), where `todo landed`
// records and verifies it.
func squashLandedRepo(t *testing.T) (dir, prHead string) {
	t.Helper()
	dir = filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	fcGit(t, dir, "init", "-q", "-b", "main")
	write := func(name string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, dir, "add", name)
	}
	write("base.txt")
	fcGit(t, dir, "commit", "-q", "-m", "base")
	fcGit(t, dir, "checkout", "-q", "-b", "pr")
	write("pr.txt")
	fcGit(t, dir, "commit", "-q", "-m", "pr head")
	prHead = fcGit(t, dir, "rev-parse", "HEAD")
	fcGit(t, dir, "checkout", "-q", "main")

	reports := filepath.Join(dir, ".moai", "reports", "t1")
	if err := os.MkdirAll(reports, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(reports, "sync-audit-iter1.md"),
		[]byte("verdict: PASS\naudited_sha: "+prHead[:12]+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir, prHead
}

// squashLandedRoot lands the squash commit S on the queue root's repository
// (HEAD and origin/main both at S) and returns S plus unlandedQ, a commit
// that never reached origin/main.
func squashLandedRoot(t *testing.T, root string) (squash, unlandedQ string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "squash.txt"), []byte("squash\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "add", "squash.txt")
	fcGit(t, root, "commit", "-q", "-m", "fix: the change (t1) (#1)")
	squash = fcGit(t, root, "rev-parse", "HEAD")
	unlandedQ = fcGit(t, root, "commit-tree", squash+"^{tree}", "-p", squash+"^", "-m", "never landed")
	fcGit(t, root, "update-ref", "refs/remotes/origin/main", squash)
	return squash, unlandedQ
}

// squashLandedFixture places the linked card t1 in the run stage with an empty
// evidence_sha (the lane landed through the PR path, so no stage evidence was
// ever written) and, when landedSHA is non-empty, the operator's landing record.
func squashLandedFixture(t *testing.T, landedSHA string) (root, uuid, prHead, squash string, store *factory.BacklogStore) {
	t.Helper()
	root, store = fcFixture(t)
	uuid = fcLinkedCard(t, root, store, "t1", "squash landed card")
	dir, prHead := squashLandedRepo(t)
	squash, q := squashLandedRoot(t, root)
	if landedSHA == "unlanded" {
		landedSHA = q
	}
	if landedSHA == "squash" {
		landedSHA = squash
	}
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun,
		OwnerLabel: "lane-1", WorktreePath: dir,
	})
	// The binding a real assignment records; fcPlace inserts the row bare.
	bind := fcOpen(t, root)
	if err := bind.RecordDispatchBinding(context.Background(), "t1", fcRun, fcNow); err != nil {
		t.Fatalf("record dispatch binding: %v", err)
	}
	_ = bind.Close()
	if landedSHA != "" {
		if err := store.Mutate(func(r *factory.BacklogRecord) error {
			for i := range r.Items {
				if r.Items[i].ID == "t1" {
					r.Items[i].Landing = &factory.LandingEvidence{
						Ref: "origin/main", RefHead: squash, ObservedAt: "2026-10-26T00:00:00Z",
						SHA: landedSHA, SHASource: factory.LandingSHASourceOperator,
					}
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	return root, uuid, prHead, squash, store
}

// engageCard records the first factory event naming card t1 in the run, the
// instant the card's engagement is measured from.
func engageCard(t *testing.T, root, at string) {
	t.Helper()
	db := fcOpen(t, root)
	if _, err := db.DB.Exec(`INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,?,?,?)`,
		fcRun, "card.assigned", `{"card_id":"t1"}`, at); err != nil {
		t.Fatalf("record engagement event: %v", err)
	}
	_ = db.Close()
}

// setLandingObservedAt rewrites the observation time on card t1's landing record.
func setLandingObservedAt(t *testing.T, store *factory.BacklogStore, at string) {
	t.Helper()
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == "t1" && r.Items[i].Landing != nil {
				r.Items[i].Landing.ObservedAt = at
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// TestFactoryRebindEvidenceRefusals pins the verb's refusals ahead of the
// verdict and landing paths: each one writes nothing.
func TestFactoryRebindEvidenceRefusals(t *testing.T) {
	rebind := func(t *testing.T, id string) error {
		t.Helper()
		_, _, err := runFactory(t, "rebind-evidence", id, "--run", fcRun, "--reason", "refusal coverage")
		return err
	}
	t.Run("an_id_that_is_not_a_card_id", func(t *testing.T) {
		fcFixture(t)
		if err := rebind(t, "bad/id"); err == nil || !strings.Contains(err.Error(), "is not a card id") {
			t.Fatalf("err = %v, want the not-a-card-id refusal", err)
		}
	})
	t.Run("no_factory_store_yet", func(t *testing.T) {
		fcFixture(t)
		if err := rebind(t, "t1"); err == nil || !strings.Contains(err.Error(), "has no factory record") {
			t.Fatalf("err = %v, want the no-factory-record refusal", err)
		}
	})
	t.Run("the_card_has_no_row_in_the_run", func(t *testing.T) {
		root, _ := fcFixture(t)
		fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardRun, Stage: homestate.CardRun})
		if err := rebind(t, "t1"); err == nil || !strings.Contains(err.Error(), "has no factory record under run") {
			t.Fatalf("err = %v, want the no-row-in-run refusal", err)
		}
	})
	t.Run("the_audit_phase_cannot_be_inferred", func(t *testing.T) {
		root, _ := fcFixture(t)
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned})
		if err := rebind(t, "t1"); err == nil || !strings.Contains(err.Error(), "cannot infer the audit phase") {
			t.Fatalf("err = %v, want the audit-phase refusal", err)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Errorf("audit rows = %d, want 0 on a refusal", n)
		}
	})
	t.Run("a_second_run_on_a_converged_and_audited_record_is_a_duplicate", func(t *testing.T) {
		root, _, _ := rebindFixture(t)
		if out, err := rebindRun(t, "first rebind"); err != nil {
			t.Fatalf("first rebind: %v\nout: %s", err, out)
		}
		before := fcCard(t, root, "t1")
		if _, err := rebindRun(t, "second rebind"); err == nil || !strings.Contains(err.Error(), "converged-and-audited duplicate") {
			t.Fatalf("err = %v, want the duplicate refusal", err)
		}
		if after := fcCard(t, root, "t1"); after.Version != before.Version {
			t.Errorf("a duplicate refusal changed the record: v%d → v%d", before.Version, after.Version)
		}
		if n := mdAuditCount(t, root); n != 1 {
			t.Errorf("audit rows = %d, want the single row of the first run", n)
		}
	})
	t.Run("a_command_without_a_context_still_runs", func(t *testing.T) {
		fcFixture(t) // no factory store yet, so the call stops at the no-record refusal
		err := runFactoryRebindEvidence(&cobra.Command{}, "t1", "no context supplied", fcRun)
		if err == nil || !strings.Contains(err.Error(), "has no factory record") {
			t.Fatalf("err = %v, want the no-factory-record refusal", err)
		}
	})
	t.Run("an_empty_stage_falls_back_to_the_state", func(t *testing.T) {
		// State run with no stage infers the sync-audit phase; with no verdict
		// file and no landing record the verb then refuses for the verdict.
		root, _ := fcFixture(t)
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardRun, WorktreePath: t.TempDir()})
		if err := rebind(t, "t1"); err == nil || !strings.Contains(err.Error(), "no sync-audit verdict") {
			t.Fatalf("err = %v, want a sync-audit verdict refusal", err)
		}
	})
}

func TestFactoryRebindEvidenceSquashLanded(t *testing.T) {
	t.Run("rebinds_to_the_recorded_squash_commit_then_approve_and_the_gate_pass", func(t *testing.T) {
		root, uuid, _, squash, _ := squashLandedFixture(t, "squash")
		out, err := rebindRun(t, "PR squash-landed; audited PR head is not an ancestor (card t1962)")
		if err != nil {
			t.Fatalf("rebind-evidence on a squash-landed card: %v\nout: %s", err, out)
		}
		if got := fcCard(t, root, "t1").EvidenceSHA; got != squash {
			t.Fatalf("evidence_sha = %q, want the squash commit %s", got, squash)
		}
		rows := mdAuditRows(t, root, "t1")
		if len(rows) != 1 || rows[0].New != squash {
			t.Fatalf("audit rows = %+v, want exactly one rebind-evidence row ending at %s", rows, squash)
		}
		if rows[0].Source != homestate.EvidenceSourceLanding {
			t.Errorf("audit row source = %q, want %q", rows[0].Source, homestate.EvidenceSourceLanding)
		}
		if _, _, err := runFactory(t, "approve", "t1", "--run", fcRun, "--issuer", "lead"); err != nil {
			t.Fatalf("approve after the rebind: %v", err)
		}
		db := fcOpen(t, root)
		if err := db.VerifyApprovalReadonly(context.Background(), "t1", uuid); err != nil {
			t.Fatalf("the done-side receipt gate refused the squash-landed card after approve: %v", err)
		}
	})

	t.Run("a_resolvable_verdict_wins_over_a_recorded_landing", func(t *testing.T) {
		root, _, _, squash, _ := squashLandedFixture(t, "squash")
		dir := fcCard(t, root, "t1").WorktreePath
		// The verdict audits the base commit, an ancestor of HEAD, so the
		// verdict path resolves and the landing record must not be consulted.
		base := fcGit(t, dir, "rev-list", "--max-parents=0", "HEAD")
		verdict := filepath.Join(dir, ".moai", "reports", "t1", "sync-audit-iter2.md")
		if err := os.WriteFile(verdict, []byte("verdict: PASS\naudited_sha: "+base[:12]+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if out, err := rebindRun(t, "verdict resolves, landing is only the fallback"); err != nil {
			t.Fatalf("rebind-evidence: %v\nout: %s", err, out)
		}
		if got := fcCard(t, root, "t1").EvidenceSHA; got != base || got == squash {
			t.Fatalf("evidence_sha = %q, want the verdict's audited commit %s, not the landing commit %s", got, base, squash)
		}
		if rows := mdAuditRows(t, root, "t1"); len(rows) != 1 || rows[0].Source != homestate.EvidenceSourceVerdict {
			t.Fatalf("audit rows = %+v, want one row with source %q", rows, homestate.EvidenceSourceVerdict)
		}
	})

	t.Run("rebinds_to_the_recorded_landing_when_no_verdict_file_exists", func(t *testing.T) {
		root, _, _, squash, _ := squashLandedFixture(t, "squash")
		dir := fcCard(t, root, "t1").WorktreePath
		if err := os.RemoveAll(filepath.Join(dir, ".moai", "reports", "t1")); err != nil {
			t.Fatal(err)
		}
		if out, err := rebindRun(t, "no verdict file; the operator recorded the landing"); err != nil {
			t.Fatalf("rebind-evidence: %v\nout: %s", err, out)
		}
		if got := fcCard(t, root, "t1").EvidenceSHA; got != squash {
			t.Fatalf("evidence_sha = %q, want the recorded landing commit %s", got, squash)
		}
	})

	t.Run("verifies_against_the_ref_the_landing_record_names", func(t *testing.T) {
		// The operator recorded the landing against origin/release; the squash
		// commit is NOT on origin/main, so only the recorded ref can admit it.
		root, _, _, squash, store := squashLandedFixture(t, "squash")
		fcGit(t, root, "update-ref", "refs/remotes/origin/main", squash+"^")
		fcGit(t, root, "update-ref", "refs/remotes/origin/release", squash)
		if err := store.Mutate(func(r *factory.BacklogRecord) error {
			for i := range r.Items {
				if r.Items[i].ID == "t1" && r.Items[i].Landing != nil {
					r.Items[i].Landing.Ref = "origin/release"
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if out, err := rebindRun(t, "landing recorded against origin/release"); err != nil {
			t.Fatalf("rebind-evidence: %v\nout: %s", err, out)
		}
		if got := fcCard(t, root, "t1").EvidenceSHA; got != squash {
			t.Fatalf("evidence_sha = %q, want the landing commit %s admitted by the recorded ref", got, squash)
		}
	})

	t.Run("verifies_in_the_queue_root_not_the_card_worktree", func(t *testing.T) {
		// `todo landed --ref HEAD` is resolved in the queue root's repository,
		// where HEAD is the squash commit; the card worktree's HEAD is the old
		// base. Verification must follow the record to the root.
		root, _, _, squash, store := squashLandedFixture(t, "squash")
		dir := fcCard(t, root, "t1").WorktreePath
		if fcGit(t, dir, "rev-parse", "HEAD") == squash {
			t.Fatal("fixture: the card worktree HEAD must differ from the root HEAD")
		}
		if err := store.Mutate(func(r *factory.BacklogRecord) error {
			for i := range r.Items {
				if r.Items[i].ID == "t1" && r.Items[i].Landing != nil {
					r.Items[i].Landing.Ref = "HEAD"
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if out, err := rebindRun(t, "landing recorded against the root HEAD"); err != nil {
			t.Fatalf("rebind-evidence: %v\nout: %s", err, out)
		}
		if got := fcCard(t, root, "t1").EvidenceSHA; got != squash {
			t.Fatalf("evidence_sha = %q, want the landing commit %s", got, squash)
		}
	})

	t.Run("refuses_a_landing_recorded_before_the_run_engaged_the_card", func(t *testing.T) {
		// A reissued card keeps its id, so the previous issuance's landing
		// record is still on the queue row; it predates this run's engagement.
		root, _, _, _, _ := squashLandedFixture(t, "squash")
		engageCard(t, root, "2026-10-26T09:00:00Z") // fixture landing observed 2026-10-26T00:00:00Z
		before := fcCard(t, root, "t1")
		_, err := rebindRun(t, "stale landing from an earlier issuance")
		if err == nil || !strings.Contains(err.Error(), "before this run engaged") {
			t.Fatalf("err = %v, want a refusal naming a landing recorded before this run engaged the card", err)
		}
		if after := fcCard(t, root, "t1"); after.EvidenceSHA != before.EvidenceSHA || after.Version != before.Version {
			t.Errorf("a refusal changed the record: %q v%d → %q v%d", before.EvidenceSHA, before.Version, after.EvidenceSHA, after.Version)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Errorf("audit rows = %d, want 0 on a refusal", n)
		}
	})

	t.Run("accepts_a_landing_recorded_after_the_run_engaged_the_card", func(t *testing.T) {
		root, _, _, squash, _ := squashLandedFixture(t, "squash")
		engageCard(t, root, "2026-10-25T00:00:00Z")
		if out, err := rebindRun(t, "landing recorded after the run engaged the card"); err != nil {
			t.Fatalf("rebind-evidence: %v\nout: %s", err, out)
		}
		if got := fcCard(t, root, "t1").EvidenceSHA; got != squash {
			t.Fatalf("evidence_sha = %q, want %s", got, squash)
		}
	})

	t.Run("accepts_a_landing_recorded_in_the_second_the_run_engaged_the_card", func(t *testing.T) {
		// `todo landed` stores ObservedAt at second precision while the card's
		// events carry fractions: 00:00:00Z can be 0.9 s AFTER an engagement at
		// 00:00:00.1Z, so only a strictly earlier second is "before".
		root, _, _, squash, _ := squashLandedFixture(t, "squash") // observed 2026-10-26T00:00:00Z
		engageCard(t, root, "2026-10-26T00:00:00.100Z")
		if out, err := rebindRun(t, "recorded in the same second as the engagement"); err != nil {
			t.Fatalf("rebind-evidence: %v\nout: %s", err, out)
		}
		if got := fcCard(t, root, "t1").EvidenceSHA; got != squash {
			t.Fatalf("evidence_sha = %q, want %s", got, squash)
		}
	})

	t.Run("refuses_a_landing_recorded_in_the_second_before_the_engagement", func(t *testing.T) {
		// The fraction forgiveness stops at the second: 23:59:59Z is earlier
		// than an engagement at 00:00:00.100Z of the next second.
		root, _, _, _, store := squashLandedFixture(t, "squash")
		engageCard(t, root, "2026-10-26T00:00:00.100Z")
		setLandingObservedAt(t, store, "2026-10-25T23:59:59Z")
		if _, err := rebindRun(t, "recorded one second before the engagement"); err == nil || !strings.Contains(err.Error(), "before this run engaged") {
			t.Fatalf("err = %v, want the before-engagement refusal", err)
		}
	})

	t.Run("refuses_an_unreadable_observation_time", func(t *testing.T) {
		root, _, _, _, store := squashLandedFixture(t, "squash")
		engageCard(t, root, "2026-10-25T00:00:00Z")
		setLandingObservedAt(t, store, "yesterday")
		_, err := rebindRun(t, "the record carries no readable observation time")
		if err == nil || !strings.Contains(err.Error(), "observation time") {
			t.Fatalf("err = %v, want a refusal naming the unreadable observation time", err)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Errorf("audit rows = %d, want 0 on a refusal", n)
		}
	})

	t.Run("refuses_when_no_landing_is_recorded", func(t *testing.T) {
		root, _, _, _, _ := squashLandedFixture(t, "")
		before := fcCard(t, root, "t1")
		_, err := rebindRun(t, "nothing landed")
		if err == nil || !strings.Contains(err.Error(), "ancestor") {
			t.Fatalf("err = %v, want the original not-an-ancestor refusal", err)
		}
		if after := fcCard(t, root, "t1"); after.EvidenceSHA != before.EvidenceSHA || after.Version != before.Version {
			t.Errorf("a refusal changed the record: %q v%d → %q v%d", before.EvidenceSHA, before.Version, after.EvidenceSHA, after.Version)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Errorf("audit rows = %d, want 0 on a refusal", n)
		}
	})

	t.Run("refuses_a_recorded_sha_that_is_not_reachable_from_the_landed_ref", func(t *testing.T) {
		root, _, _, _, _ := squashLandedFixture(t, "unlanded")
		before := fcCard(t, root, "t1")
		_, err := rebindRun(t, "recorded sha never reached main")
		if err == nil {
			t.Fatal("rebind-evidence accepted a landing SHA that origin/main does not contain")
		}
		if after := fcCard(t, root, "t1"); after.EvidenceSHA != before.EvidenceSHA || after.Version != before.Version {
			t.Errorf("a refusal changed the record: %q v%d → %q v%d", before.EvidenceSHA, before.Version, after.EvidenceSHA, after.Version)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Errorf("audit rows = %d, want 0 on a refusal", n)
		}
	})

	t.Run("approve_names_rebind_evidence_when_the_row_has_no_evidence", func(t *testing.T) {
		squashLandedFixture(t, "squash")
		_, _, err := runFactory(t, "approve", "t1", "--run", fcRun, "--issuer", "lead")
		if err == nil || !strings.Contains(err.Error(), "rebind-evidence") {
			t.Fatalf("err = %v, want the evidence-hash refusal to name `factory rebind-evidence` as the remedy", err)
		}
	})
}
