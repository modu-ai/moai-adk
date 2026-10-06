package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// Card-review repair tests (card t1500 card-review P2 findings 1-6): each
// test reproduced its finding on the pre-repair tree before its fix, in
// package (the reviewer's temporary-overlay probes promoted to real tests).

// F1 — audit_ceiling.go loadCeilings: an INVALID on_final_hit is a config
// error, not a missing-file default. The harness config's own load
// validation rejects the value; the engine must surface that refusal, not
// silently proceed on the default ceilings.
func TestEvaluateCeilingInvalidPolicyRefused(t *testing.T) {
	f := newCeilingFixture(t)
	harness := "harness:\n  evaluator:\n    memory_scope: per_iteration\n  plan_audit_tier_ceilings:\n    S: 1\n    M: 2\n    L: 1\n  plan_audit_ceiling_policy:\n    auto_delta_rounds: 0\n    on_final_hit: admit\n"
	if err := os.WriteFile(filepath.Join(f.root, ".moai", "config", "sections", "harness.yaml"), []byte(harness), 0o600); err != nil {
		t.Fatal(err)
	}
	path := f.writeIter(t, 1, "PASS")
	fields, hashOK := f.fieldsFromHash(t, path)
	_, _, err := EvaluateCeiling(f.input(), fields, hashOK, nil)
	if err == nil {
		t.Fatal("an invalid on_final_hit policy was swallowed; want a config error")
	}
}

// F2 — audit_counter.go: legacy plan-audit-N.md shapes (the kickoff reader's
// planAuditNameRe accepts them) are iteration evidence too — a counter that
// ignores the very file the seam reads can never reach its ceiling.
func TestCountAuditRoundsLegacyPlanAuditNumbered(t *testing.T) {
	specID := "SPEC-ACE-NUM-001"
	dir := t.TempDir()
	writeAuditFixture(t, dir, "plan-audit-1.md", specID, "PASS", 0)
	writeAuditFixture(t, dir, "plan-audit-2.md", specID, "FAIL", 0)

	ev, err := CountAuditRounds(specID, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if ev.Count != 2 {
		t.Fatalf("count %d, want 2 (plan-audit-N.md shapes are evidence)", ev.Count)
	}
	if ev.LatestPath == "" || !strings.Contains(ev.LatestPath, "plan-audit-2.md") {
		t.Fatalf("latest %q, want the numbered shape with the highest N", ev.LatestPath)
	}
}

// F4 — audit_ceiling.go previousAuditedSHA: the previous round is the
// largest round strictly below the latest after dedupe, across BOTH families
// — a latest convention iteration with legacy prior rounds still has a
// previous audited SHA.
func TestPreviousAuditedSHALegacyPriorRound(t *testing.T) {
	specID := "SPEC-ACE-PREV-001"
	dir := t.TempDir()
	iter3 := writeAuditFixture(t, dir, "plan-audit-iter3.md", specID, "FAIL", 0)
	if err := os.WriteFile(filepath.Join(dir, "plan-audit-iter3.md"), []byte(replaceSHA(t, iter3, "sha-iter3")), 0o600); err != nil {
		t.Fatal(err)
	}
	rev2 := filepath.Join(dir, specID+"-review-2.md")
	if err := os.WriteFile(rev2, []byte("# review\nverdict: FAIL\naudited_sha: sha-rev2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rev1 := filepath.Join(dir, specID+"-review-1.md")
	if err := os.WriteFile(rev1, []byte("# review\nverdict: FAIL\naudited_sha: sha-rev1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ev, err := CountAuditRounds(specID, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	in := VerdictCeilingInput{SpecID: specID, ProjectRoot: t.TempDir()}
	if got := previousAuditedSHA(in, ev); got != "sha-rev2" {
		t.Fatalf("previous audited SHA %q, want sha-rev2 (largest round below the latest, legacy included)", got)
	}
}

func replaceSHA(t *testing.T, path, sha string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	for i, l := range lines {
		if strings.HasPrefix(l, "audited_sha:") {
			lines[i] = "audited_sha: " + sha
		}
	}
	return strings.Join(lines, "\n")
}

// F5 — audit_counter.go attribution: a convention file belongs to the SPEC
// named in its REPORT HEADER, compared exactly — another SPEC's report that
// merely mentions the target SPEC in its body is not its round.
func TestCountAuditRoundsExactHeaderAttribution(t *testing.T) {
	target := "SPEC-ACE-TARGET-001"
	dir := t.TempDir()
	body := "# SPEC Review Report: SPEC-ACE-FOREIGN-001\nverdict: PASS\nOverall Score: 0.90\nmust_pass_failed: 0\nblocking_count: 0\nplan_artifact_hash: h\naudited_sha: a\nnotes: cross-references SPEC-ACE-TARGET-001 in prose\n"
	if err := os.WriteFile(filepath.Join(dir, "plan-audit-iter1.md"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	ev, err := CountAuditRounds(target, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if ev.Count != 0 {
		t.Fatalf("count %d, want 0 (the report belongs to SPEC-ACE-FOREIGN-001)", ev.Count)
	}
}

// F6 — audit_counter.go RoundReportDirs: a directory carrying ONLY legacy
// stream files (<SPEC-ID>-review-N.md, which name the SPEC in their file
// name) is round evidence and is discovered.
func TestRoundReportDirsLegacyOnlyDir(t *testing.T) {
	specID := "SPEC-ACE-DIRS2-001"
	root := t.TempDir()
	reports := filepath.Join(root, ".moai", "reports")
	legacyOnly := filepath.Join(reports, "t7001")
	if err := os.MkdirAll(legacyOnly, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyOnly, specID+"-review-1.md"), []byte("# review\nverdict: PASS\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	dirs := RoundReportDirs(root, specID, "")
	if !strings.Contains(strings.Join(dirs, "\n"), legacyOnly) {
		t.Fatalf("legacy-only evidence directory missing from %v", dirs)
	}
}

// F3 — audit_ceiling.go diffInsideAnchors: the #anchor is load-bearing — an
// edit hunk in the anchored file that sits OUTSIDE the anchor's section does
// not verify, while an edit inside the anchor's vicinity does.
func TestDiffInsideAnchorsHunkScope(t *testing.T) {
	root := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		out, err := auditreceipt.RunScrubbedGit(root, args...)
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return strings.TrimSpace(out)
	}
	specDir := filepath.Join(root, ".moai", "specs", "SPEC-ACE-HUNK-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	git("init", "-q")
	git("config", "user.email", "t@example.com")
	git("config", "user.name", "t")
	spec := "# spec\n\n## alpha\nalpha 1\nalpha 2\nalpha 3\nalpha 4\nalpha 5\n\n## REQ-ACE-001 anchor section\nREQ-ACE-001 detail\n\n## tail\ntail 1\ntail 2\ntail 3\ntail 4\ntail 5\ntail 6\ntail 7\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(spec), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	base := git("rev-parse", "HEAD")

	anchor := []string{".moai/specs/SPEC-ACE-HUNK-001/spec.md#REQ-ACE-001"}

	// An in-anchor edit: the REQ-ACE-001 line itself.
	inAnchor := strings.Replace(spec, "REQ-ACE-001 detail", "REQ-ACE-001 detail repaired", 1)
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(inAnchor), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "in-anchor")
	mid := git("rev-parse", "HEAD")
	if !diffInsideAnchors(root, base, mid, anchor) {
		t.Fatal("an edit hunk inside the anchor's vicinity did not verify")
	}

	// An out-of-anchor edit in the SAME file: the tail section, whose hunk
	// context carries no REQ-ACE-001.
	outAnchor := strings.Replace(inAnchor, "tail 4", "tail 4 repaired", 1)
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(outAnchor), 0o600); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "outside")
	tip := git("rev-parse", "HEAD")
	if diffInsideAnchors(root, mid, tip, anchor) {
		t.Fatal("an out-of-anchor hunk in the anchored file verified")
	}
}
