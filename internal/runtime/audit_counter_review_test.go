package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// Card-review repair tests (card t1500 card-review P2 findings 1-6): each
// test reproduced its finding on the pre-repair tree before its fix, in
// package (the reviewer's temporary-overlay probes promoted to real tests).

// F1 — audit_ceiling.go loadCeilings: an UNKNOWN on_final_hit NAME passes
// the config load without an error (SPEC-AUDIT-CEILING-002 config matrix
// M2/M12 pass-through; the load-time rejection F1 originally surfaced was
// retired with that convention) but never activates the delta ladder — the
// engine reads the name at evaluation level and grants no delta round for a
// policy it does not implement (fail-closed: deltaRounds forced to zero
// even when auto_delta_rounds is configured), while the configured tier
// ceiling itself is still honored.
func TestEvaluateCeilingUnknownPolicyFailClosed(t *testing.T) {
	f := newCeilingFixture(t)
	harness := "harness:\n  evaluator:\n    memory_scope: per_iteration\n  plan_audit_tier_ceilings:\n    S: 1\n    M: 2\n    L: 1\n  plan_audit_ceiling_policy:\n    auto_delta_rounds: 3\n    on_final_hit: admit\n"
	if err := os.WriteFile(filepath.Join(f.root, ".moai", "config", "sections", "harness.yaml"), []byte(harness), 0o600); err != nil {
		t.Fatal(err)
	}
	tierCeiling, deltaRounds, policyNamed, err := loadCeilings(f.root, "L")
	if err != nil {
		t.Fatalf("an unknown on_final_hit NAME is load pass-through, not a config error: %v", err)
	}
	if policyNamed {
		t.Fatal("an unimplemented policy name was treated as an enforcing hold-and-split")
	}
	if deltaRounds != 0 {
		t.Fatalf("deltaRounds %d, want 0 — an unknown policy name never grants the delta ladder", deltaRounds)
	}
	if tierCeiling != 1 {
		t.Fatalf("tierCeiling %d, want 1 (the configured L ceiling is honored)", tierCeiling)
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

// SPEC-AUDIT-CEILING-REPAIR-001 reproduction tests (card t1560). Each RED
// test below was observed failing on unmodified main 903ccd028 for its
// stated reason before its fix, per the engine family convention
// ("RED is a new test — E8 evidence required").

// AC-ACR-001 (D1 RED) — previousAuditedSHA: the round counter counts BOTH
// families and can select a legacy-family file (<SpecID>-review-<N>.md, the
// card-review F4 stream) as RoundEvidence.LatestPath, so the LATEST round
// number must resolve through the same dual-family parse the prior-round
// scan applies. With the defect, iterationOf (convention-only) returns 0
// for a legacy latest, the latestN <= 0 guard aborts, and the previous
// audited SHA is lost.
func TestPreviousAuditedSHALatestLegacyRound(t *testing.T) {
	specID := "SPEC-ACE-LEG2-001"
	dir := t.TempDir()
	writeReview := func(n int, sha string) {
		t.Helper()
		p := filepath.Join(dir, fmt.Sprintf("%s-review-%d.md", specID, n))
		if err := os.WriteFile(p, []byte("# review\nverdict: FAIL\naudited_sha: "+sha+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeReview(1, "sha-rev1")
	writeReview(2, "sha-rev2")
	writeReview(3, "sha-rev3")
	ev, err := CountAuditRounds(specID, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if ev.LatestPath == "" || !strings.HasSuffix(ev.LatestPath, "-review-3.md") {
		t.Fatalf("latest %q, want the legacy -review-3.md stream tip", ev.LatestPath)
	}
	in := VerdictCeilingInput{SpecID: specID, ProjectRoot: t.TempDir()}
	if got := previousAuditedSHA(in, ev); got != "sha-rev2" {
		t.Fatalf("previous audited SHA %q, want sha-rev2 (largest round strictly below the legacy latest)", got)
	}
}

// AC-ACR-002 (D1 mixed families) — once the latest number resolves through
// the dual-family parse, the prior-round scan's existing dual-family
// behavior finds a convention-family prior below a legacy-family latest.
func TestPreviousAuditedSHAMixedFamilyPrior(t *testing.T) {
	specID := "SPEC-ACE-MIXFAM-001"
	dir := t.TempDir()
	iter2 := writeAuditFixture(t, dir, "plan-audit-iter2.md", specID, "FAIL", 0)
	if err := os.WriteFile(iter2, []byte(replaceSHA(t, iter2, "sha-iter2")), 0o600); err != nil {
		t.Fatal(err)
	}
	rev3 := filepath.Join(dir, specID+"-review-3.md")
	if err := os.WriteFile(rev3, []byte("# review\nverdict: FAIL\naudited_sha: sha-rev3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ev, err := CountAuditRounds(specID, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	in := VerdictCeilingInput{SpecID: specID, ProjectRoot: t.TempDir()}
	if got := previousAuditedSHA(in, ev); got != "sha-iter2" {
		t.Fatalf("previous audited SHA %q, want sha-iter2 (convention prior below the legacy latest)", got)
	}
}

// AC-ACR-003 (D1 preservation, RG) — a latest name parsing under NEITHER
// family keeps the no-prior-round fail-closed semantics: "" before and
// after the fix. Preserve-behavior check — passes on unmodified main and
// must keep passing; never a RED observation.
func TestPreviousAuditedSHAUnparseableLatestStaysEmpty(t *testing.T) {
	specID := "SPEC-ACE-GARBAGE-001"
	dir := t.TempDir()
	p := filepath.Join(dir, "garbage.md")
	if err := os.WriteFile(p, []byte("# verdict\naudited_sha: sha-x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ev := RoundEvidence{Sources: []string{p}, LatestPath: p}
	in := VerdictCeilingInput{SpecID: specID, ProjectRoot: t.TempDir()}
	if got := previousAuditedSHA(in, ev); got != "" {
		t.Fatalf("previous audited SHA %q, want empty for an unparseable latest name (fail-closed)", got)
	}
}

// AC-ACR-014 (D4 RED + preserve arms) — CountAuditRounds: a bare base
// report and an Atoi-overflowing iteration suffix each count as their OWN
// round (the n=1 initialization at audit_counter.go:109 and the Atoi
// collapse at :111 are the defect); parsed-number dedupe, the bare base
// alone, and explicit-0 parity are preserved.
func TestCountAuditRoundsOverflowOwnRound(t *testing.T) {
	specID := "SPEC-ACE-OF-001"

	// (a) base + overflow suffix → 2. RED today: 1 — the Atoi range error
	// leaves n at its initialized 1, deduping into seen[1]; the overflow
	// file is fail-counted and never becomes LatestPath.
	t.Run("base_plus_overflow_suffix_counts_2", func(t *testing.T) {
		dir := t.TempDir()
		writeAuditFixture(t, dir, "plan-audit.md", specID, "FAIL", 0)
		writeAuditFixture(t, dir, "plan-audit-iter99999999999999999999.md", specID, "FAIL", 0)
		ev, err := CountAuditRounds(specID, []string{dir})
		if err != nil {
			t.Fatal(err)
		}
		if ev.Count != 2 {
			t.Fatalf("count %d, want 2 (the base and the overflow suffix are two rounds)", ev.Count)
		}
		if ev.LatestPath == "" || !strings.HasSuffix(ev.LatestPath, "plan-audit.md") {
			t.Fatalf("latest %q, want the base report — an overflow suffix never becomes LatestPath", ev.LatestPath)
		}
	})

	// (b) base vs numbered → 2. RED today: 1 — the convention branch
	// initializes n = 1 BEFORE the Atoi attempt, so base and iter1 collapse
	// into seen[1].
	t.Run("base_versus_numbered_counts_2", func(t *testing.T) {
		dir := t.TempDir()
		writeAuditFixture(t, dir, "plan-audit.md", specID, "FAIL", 0)
		writeAuditFixture(t, dir, "plan-audit-iter1.md", specID, "FAIL", 0)
		ev, err := CountAuditRounds(specID, []string{dir})
		if err != nil {
			t.Fatal(err)
		}
		if ev.Count != 2 {
			t.Fatalf("count %d, want 2 (the base never merges with iter1)", ev.Count)
		}
		if ev.LatestPath == "" || !strings.Contains(ev.LatestPath, "iter1") {
			t.Fatalf("latest %q, want iter1 (the numbered round orders after the base)", ev.LatestPath)
		}
	})

	// (c) one normal + two overflow files → 1+N = 3. RED today: 1 with
	// sources=3 (the collapse the leader's gate measured at 903ccd028).
	t.Run("one_normal_plus_two_overflow_counts_3", func(t *testing.T) {
		dir := t.TempDir()
		writeAuditFixture(t, dir, "plan-audit-iter1.md", specID, "FAIL", 0)
		writeAuditFixture(t, dir, "plan-audit-iter99999999999999999998.md", specID, "FAIL", 0)
		writeAuditFixture(t, dir, "plan-audit-iter99999999999999999999.md", specID, "FAIL", 0)
		ev, err := CountAuditRounds(specID, []string{dir})
		if err != nil {
			t.Fatal(err)
		}
		if ev.Count != 3 {
			t.Fatalf("count %d (sources %d), want 3 (1 normal + 2 unparseable)", ev.Count, len(ev.Sources))
		}
	})

	// Preserve: a bare base alone counts 1 and remains the LatestPath.
	t.Run("bare_base_alone_counts_1_and_stays_latest", func(t *testing.T) {
		dir := t.TempDir()
		writeAuditFixture(t, dir, "plan-audit.md", specID, "PASS", 0)
		ev, err := CountAuditRounds(specID, []string{dir})
		if err != nil {
			t.Fatal(err)
		}
		if ev.Count != 1 {
			t.Fatalf("count %d, want 1", ev.Count)
		}
		if ev.LatestPath == "" || !strings.HasSuffix(ev.LatestPath, "plan-audit.md") {
			t.Fatalf("latest %q, want the base report", ev.LatestPath)
		}
	})

	// Preserve (f): explicit-0 parity — plan-audit-0.md keeps its own
	// seen[0] identity and never merges with the bare base; LatestPath
	// stays the base.
	t.Run("explicit_zero_keeps_own_round_parity", func(t *testing.T) {
		dir := t.TempDir()
		writeAuditFixture(t, dir, "plan-audit.md", specID, "FAIL", 0)
		writeAuditFixture(t, dir, "plan-audit-0.md", specID, "FAIL", 0)
		ev, err := CountAuditRounds(specID, []string{dir})
		if err != nil {
			t.Fatal(err)
		}
		if ev.Count != 2 {
			t.Fatalf("count %d, want 2 (base and plan-audit-0.md are distinct rounds)", ev.Count)
		}
		if ev.LatestPath == "" || !strings.HasSuffix(ev.LatestPath, "plan-audit.md") {
			t.Fatalf("latest %q, want the base report (explicit-0 parity)", ev.LatestPath)
		}
	})
}

// AC-ACR-014 arm (e) (D4 RED) — previousAuditedSHA: the base report orders
// EARLIEST (round 0, the planAuditRoundFile convention), so a numbered
// latest resolves the base as its previous audited round when no numbered
// round orders between them. RED today: "" — the scan skips the base
// (n >= latestN).
func TestPreviousAuditedSHABaseRoundBaseline(t *testing.T) {
	specID := "SPEC-ACE-BASE-001"
	dir := t.TempDir()
	base := writeAuditFixture(t, dir, "plan-audit.md", specID, "FAIL", 0)
	if err := os.WriteFile(base, []byte(replaceSHA(t, base, "sha-base")), 0o600); err != nil {
		t.Fatal(err)
	}
	iter1 := writeAuditFixture(t, dir, "plan-audit-iter1.md", specID, "FAIL", 0)
	if err := os.WriteFile(iter1, []byte(replaceSHA(t, iter1, "sha-iter1")), 0o600); err != nil {
		t.Fatal(err)
	}
	ev, err := CountAuditRounds(specID, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	// The previous-baseline assertion comes first: it is the RED face of
	// this arm. (The counter's own n=1 collapse makes the same fixture
	// count 1 at the pre-repair tree, so the count assertions below read
	// only after the previous-baseline behavior is established.)
	in := VerdictCeilingInput{SpecID: specID, ProjectRoot: t.TempDir()}
	if got := previousAuditedSHA(in, ev); got != "sha-base" {
		t.Fatalf("previous audited SHA %q, want sha-base (the base is round 0, the previous audited round)", got)
	}
	if ev.Count != 2 {
		t.Fatalf("count %d, want 2", ev.Count)
	}
	if ev.LatestPath == "" || !strings.Contains(ev.LatestPath, "iter1") {
		t.Fatalf("latest %q, want iter1", ev.LatestPath)
	}
}
