package factory

// base_freshness_target_test.go — the gate measures ONE target, the configured
// integration branch (SPEC-INTEGRATION-STALE-BASE-001, card t1834, sync-audit
// iter10 H1/H2). Two seams are pinned here:
//
//   - an explicitly requested branch other than the configured one is refused
//     (or, gate off, advised) and is never measured;
//   - the re-measure names its absorbed base through the same refs/heads/<branch>
//     spelling the gate reads, so a same-named tag cannot move the base.

import (
	"strings"
	"testing"
)

func TestBaseFreshnessRequestedBranchOtherThanTheTargetIsRefusedNotMeasured(t *testing.T) {
	f := newBFFixture(t)
	f.writeRecord(f.base, nil)
	bfGit(t, f.repo, "branch", "alternate", "main")

	in := f.input(0, true)
	in.RequestedBranch = "alternate"
	res := EvaluateBaseFreshness(in, BaseFreshnessSeams{})
	if res.Admit {
		t.Fatalf("an explicit branch other than the configured target must be refused: %+v", res)
	}
	if !strings.HasPrefix(res.Refusal, BaseFreshnessSentinel) {
		t.Fatalf("the refusal must start with the sentinel: %q", res.Refusal)
	}
	for _, want := range []string{"alternate", "main", "configured integration target"} {
		if !strings.Contains(res.Refusal, want) {
			t.Fatalf("the refusal must carry %q: %q", want, res.Refusal)
		}
	}
	requireOneLine(t, "refusal", res.Refusal)
	if res.Tip != "" || res.Behind != -1 {
		t.Fatalf("a mismatched request must not be measured against either branch: tip=%q behind=%d", res.Tip, res.Behind)
	}
}

func TestBaseFreshnessRequestedBranchOtherThanTheTargetIsAdvisoryWhenGateOff(t *testing.T) {
	f := newBFFixture(t)
	f.writeRecord(f.base, nil)
	bfGit(t, f.repo, "branch", "alternate", "main")

	in := f.input(0, false)
	in.RequestedBranch = "alternate"
	res := EvaluateBaseFreshness(in, BaseFreshnessSeams{})
	if !res.Admit || res.Refusal != "" {
		t.Fatalf("the gate off admits: %+v", res)
	}
	requireOneLine(t, "notice", res.Notice)
	if !strings.HasPrefix(res.Notice, baseFreshnessAdvisoryPrefix) ||
		!strings.Contains(res.Notice, "alternate") || !strings.Contains(res.Notice, "main") {
		t.Fatalf("the advisory must name both branches: %q", res.Notice)
	}
}

func TestBaseFreshnessRequestedBranchEqualToTheTargetIsMeasuredAsBefore(t *testing.T) {
	f := newBFFixture(t)
	f.writeRecord(f.base, nil)

	for _, requested := range []string{"", "main", " main "} {
		in := f.input(0, true)
		in.RequestedBranch = requested
		res := EvaluateBaseFreshness(in, BaseFreshnessSeams{})
		if !res.Admit || res.Notice != "" || res.Behind != 0 {
			t.Fatalf("requested %q equals the target: want a silent measured admit, got %+v", requested, res)
		}
	}
}

func TestRunRemeasureNamesTheBranchTipWhenATagSharesTheBranchName(t *testing.T) {
	// Audit iter10 H2: `git merge-base main HEAD` resolves a tag named `main`
	// before the branch, so the record's base followed an older tag while the gate
	// read refs/heads/main.
	f := newBFFixture(t)
	bfGit(t, f.repo, "tag", "main", f.base)
	f.landEmpty(1)
	bfGit(t, f.card, "merge", "-q", "--no-ff", "-m", "absorb main", "refs/heads/main")
	tip := bfGit(t, f.repo, "rev-parse", "refs/heads/main")

	rec, err := RunRemeasure(f.root, f.card, "main", "true")
	if err != nil {
		t.Fatalf("remeasure of a clean candidate must succeed: %v", err)
	}
	if rec.Base != tip {
		t.Fatalf("the record's base must be the branch tip %s, got %s (the older tag is %s)", tip, rec.Base, f.base)
	}
	res := EvaluateBaseFreshness(f.input(0, true), BaseFreshnessSeams{})
	if !res.Admit {
		t.Fatalf("the gate must admit the candidate the remeasure just recorded: %s", res.Refusal)
	}
}
