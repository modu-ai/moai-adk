package cli

// factory_card_rebind_test.go — card t1930 defect ④: the formal
// `moai factory rebind-evidence <card> --run <run> --reason <text>` verb that
// re-binds a card's evidence_sha to its newest audit verdict's audited_sha.
// The normal repair loop (fix commits land AFTER the audit) used to diverge
// the recorded evidence commit from the verdict's audited_sha permanently —
// the T6/T13 guards refused forever, and lanes resorted to direct SQL
// (reproduced five times: t1832, t1536, t1929, t1828, and the t1828 re-issue
// workaround). The verb is the formalization of that operator-approved
// exception: the value comes only from the verdict file, never from the
// caller, and every mutation lands one card_metadata_events audit row.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// rebindDivergedSHA is the stale evidence commit a repair loop leaves behind:
// a full 40-hex value that names neither the base nor the tip.
const rebindDivergedSHA = "c0ffee0000000000000000000000000000000000"

// rebindRepo builds the repair-loop shape: a git repository whose HEAD
// carries two plan-audit verdict files — iter1 auditing an unrelated commit,
// iter2 (the newest, per the newest-selection rule) auditing the BASE commit,
// an ancestor of HEAD. The recorded evidence diverged from both.
func rebindRepo(t *testing.T) (dir, baseFull string) {
	t.Helper()
	dir = filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	fcGit(t, dir, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "base.txt"), []byte("base\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, dir, "add", "base.txt")
	fcGit(t, dir, "commit", "-q", "-m", "base")
	baseFull = fcGit(t, dir, "rev-parse", "HEAD")

	reports := filepath.Join(dir, ".moai", "reports", "t1")
	if err := os.MkdirAll(reports, 0o700); err != nil {
		t.Fatal(err)
	}
	// iter1 audits an unrelated commit — the newest-selection rule must skip it.
	if err := os.WriteFile(filepath.Join(reports, "plan-audit-iter1.md"),
		[]byte("verdict: FAIL\naudited_sha: 123456789012\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, dir, "add", ".moai")
	fcGit(t, dir, "commit", "-q", "-m", "verdict iter1")
	// iter2 audits the BASE commit: the auditor judged the pre-repair tree,
	// then fix commits (iter2's own landing included) moved HEAD past it.
	if err := os.WriteFile(filepath.Join(reports, "plan-audit-iter2.md"),
		[]byte("verdict: PASS\naudited_sha: "+baseFull[:12]+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, dir, "add", ".moai")
	fcGit(t, dir, "commit", "-q", "-m", "verdict iter2 after the fix commits")
	return dir, baseFull
}

// rebindFixture places a plan-audit card whose recorded evidence diverged.
func rebindFixture(t *testing.T) (root, wt, baseFull string) {
	t.Helper()
	root, _ = fcFixture(t)
	wt, baseFull = rebindRepo(t)
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: homestate.CardPlanAudit, Stage: homestate.CardPlanAudit,
		EvidenceSHA: rebindDivergedSHA, WorktreePath: wt,
	})
	return root, wt, baseFull
}

func rebindRun(t *testing.T, reason string) (string, error) {
	t.Helper()
	out, _, err := runFactory(t, "rebind-evidence", "t1", "--run", fcRun, "--reason", reason)
	return out, err
}

func TestFactoryRebindEvidence(t *testing.T) {
	t.Run("rebinds_to_the_newest_verdicts_audited_sha", func(t *testing.T) {
		root, _, baseFull := rebindFixture(t)
		out, err := rebindRun(t, "repair loop landed the verdict after the fix commits (card t1930)")
		if err != nil {
			t.Fatalf("rebind-evidence: %v\nout: %s", err, out)
		}
		c := fcCard(t, root, "t1")
		if c.EvidenceSHA != baseFull {
			t.Fatalf("evidence_sha = %q, want the newest verdict's audited commit %s (never a caller-supplied SHA)", c.EvidenceSHA, baseFull)
		}
		rows := mdAuditRows(t, root, "t1")
		if len(rows) != 1 {
			t.Fatalf("audit rows = %d, want exactly one rebind-evidence row", len(rows))
		}
		r := rows[0]
		if r.Kind != "rebind-evidence" || r.Field != "evidence_sha" {
			t.Errorf("audit row kind/field = %s/%s, want rebind-evidence/evidence_sha", r.Kind, r.Field)
		}
		if r.Old != rebindDivergedSHA || r.New != baseFull {
			t.Errorf("audit row old→new = %q→%q, want %q→%q", r.Old, r.New, rebindDivergedSHA, baseFull)
		}
		if r.Reason != "repair loop landed the verdict after the fix commits (card t1930)" {
			t.Errorf("audit row reason = %q, want the caller's verbatim reason", r.Reason)
		}
		if !strings.Contains(out, baseFull[:12]) {
			t.Errorf("output does not name the rebound commit:\n%s", out)
		}
	})

	t.Run("refuses_when_no_verdict_file_exists", func(t *testing.T) {
		root, wt, _ := rebindFixture(t)
		if err := os.RemoveAll(filepath.Join(wt, ".moai", "reports", "t1")); err != nil {
			t.Fatal(err)
		}
		before := fcCard(t, root, "t1")
		_, err := rebindRun(t, "nothing to read")
		if err == nil || !strings.Contains(err.Error(), "verdict") {
			t.Fatalf("err = %v, want a refusal naming the missing verdict file", err)
		}
		after := fcCard(t, root, "t1")
		if after.EvidenceSHA != before.EvidenceSHA || after.Version != before.Version {
			t.Errorf("a missing-verdict refusal changed the record: evidence %q v%d → %q v%d",
				before.EvidenceSHA, before.Version, after.EvidenceSHA, after.Version)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Errorf("audit rows = %d, want 0 on a refusal", n)
		}
	})

	t.Run("refuses_when_the_audited_commit_is_invalid", func(t *testing.T) {
		root, wt, _ := rebindFixture(t)
		reports := filepath.Join(wt, ".moai", "reports", "t1")
		if err := os.WriteFile(filepath.Join(reports, "plan-audit-iter3.md"),
			[]byte("verdict: PASS\naudited_sha: deadbeefcafe\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		fcGit(t, wt, "add", ".moai")
		fcGit(t, wt, "commit", "-q", "-m", "verdict iter3 names a commit that does not resolve")
		before := fcCard(t, root, "t1")
		_, err := rebindRun(t, "the verdict names a ghost")
		if err == nil || !strings.Contains(err.Error(), "deadbeefcafe") {
			t.Fatalf("err = %v, want a refusal naming the unresolvable audited commit", err)
		}
		after := fcCard(t, root, "t1")
		if after.EvidenceSHA != before.EvidenceSHA || after.Version != before.Version {
			t.Errorf("an invalid-commit refusal changed the record: evidence %q → %q", before.EvidenceSHA, after.EvidenceSHA)
		}
		if n := mdAuditCount(t, root); n != 0 {
			t.Errorf("audit rows = %d, want 0 on a refusal", n)
		}
	})

	t.Run("refuses_a_missing_reason", func(t *testing.T) {
		root, _, _ := rebindFixture(t)
		before := fcCard(t, root, "t1")
		if _, _, err := runFactory(t, "rebind-evidence", "t1", "--run", fcRun, "--reason", "  "); err == nil {
			t.Fatal("rebind-evidence ran without a reason; --reason is required (the audit row records why)")
		}
		if after := fcCard(t, root, "t1"); after.Version != before.Version {
			t.Errorf("a reason refusal changed the record: v%d → v%d", before.Version, after.Version)
		}
	})

	t.Run("refuses_a_stale_version_at_the_record_layer", func(t *testing.T) {
		// The version fence is the record layer's own contract (the same
		// ErrStaleVersion refusal mutateCardSpec carries), so it is pinned at
		// the layer that owns it.
		root, _, baseFull := rebindFixture(t)
		db := fcOpen(t, root)
		defer func() { _ = db.Close() }()
		c := fcCard(t, root, "t1")
		if _, err := db.SetCardEvidenceSHA(context.Background(), fcRun, "t1", baseFull, c.Version-1, fcNow); err == nil ||
			!strings.Contains(err.Error(), homestate.ErrStaleVersion.Error()) {
			t.Fatalf("err = %v, want the stale-version refusal (expected v%d, card at v%d)", err, c.Version-1, c.Version)
		}
		if got := fcCard(t, root, "t1").EvidenceSHA; got != rebindDivergedSHA {
			t.Errorf("a stale write changed evidence_sha to %q", got)
		}
	})
}
