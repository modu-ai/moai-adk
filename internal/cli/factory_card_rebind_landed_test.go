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

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// squashLandedRepo builds the squash shape: the audited PR head P lives only
// on branch `pr`; main carries a separate squash commit S (what the PR landed
// as) and origin/main points at S. The worktree HEAD is main, so P is not an
// ancestor of it. unlandedQ is a commit on `pr` that never reached origin/main.
func squashLandedRepo(t *testing.T) (dir, prHead, squash, unlandedQ string) {
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
	write("pr2.txt")
	fcGit(t, dir, "commit", "-q", "-m", "pr follow-up never landed")
	unlandedQ = fcGit(t, dir, "rev-parse", "HEAD")
	fcGit(t, dir, "checkout", "-q", "main")
	write("squash.txt")
	fcGit(t, dir, "commit", "-q", "-m", "fix: the change (t1) (#1)")
	squash = fcGit(t, dir, "rev-parse", "HEAD")
	fcGit(t, dir, "update-ref", "refs/remotes/origin/main", squash)

	reports := filepath.Join(dir, ".moai", "reports", "t1")
	if err := os.MkdirAll(reports, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(reports, "sync-audit-iter1.md"),
		[]byte("verdict: PASS\naudited_sha: "+prHead[:12]+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir, prHead, squash, unlandedQ
}

// squashLandedFixture places the linked card t1 in the run stage with an empty
// evidence_sha (the lane landed through the PR path, so no stage evidence was
// ever written) and, when landedSHA is non-empty, the operator's landing record.
func squashLandedFixture(t *testing.T, landedSHA string) (root, uuid, prHead, squash string, store *factory.BacklogStore) {
	t.Helper()
	root, store = fcFixture(t)
	uuid = fcLinkedCard(t, root, store, "t1", "squash landed card")
	dir, prHead, squash, q := squashLandedRepo(t)
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
