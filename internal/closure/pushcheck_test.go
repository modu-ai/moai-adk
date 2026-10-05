package closure

import (
	"bytes"
	"slices"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
	"github.com/modu-ai/moai-adk/internal/closure/gitio"
)

// fixtureSeam builds a GitSeam bound at runtime (stateless, but kept as a
// named local for readability).
func fixtureSeam() GitSeam { return GitSeam{} }

func closuretestBranch() string { return closuretest.Branch }

func gitioHeadOfDir(dir string) string {
	sha, err := gitio.Head(dir)
	if err != nil {
		return ""
	}
	return sha
}

func contains(data []byte, needle string) bool {
	return bytes.Contains(data, []byte(needle))
}

// secondReviewAtCommit records a performed pass review at the named commit.
func secondReviewAtCommit(t *testing.T, f *acFixture, head, verdict string) {
	t.Helper()
	rep, _, _, _, _, err := LoadBuildSideFiles(f.CardDir, fixSpecID, closuretest.Autonomy(), nil, nil)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	disagree := false
	rec := SecondReviewRecord{
		SchemaVersion: 1, Card: fixCard, ContractCard: fixCard, SpecID: fixSpecID,
		ContractSHA256: rep.RecordedContractSHA256, HeadSHA: head,
		Target: "baseBranch",
		Scope: SecondReviewScope{
			BaseBranch: "develop", BaseSHA: "base01", HeadSHA: head,
			ChangedFiles: 3, DiffSHA256: "diff01",
		},
		Backends:         []SecondReviewBackend{{Backend: "codex", Gate: "required", Verdict: verdict}},
		ParticipantCount: 1, DisagreementFlag: &disagree,
		RecordedAt: time.Now().UTC().Format(time.RFC3339),
	}
	f.WriteSecondReviewLine(rec)
}

// regenerateReportAtCommit rebuilds the closure report pair at the named
// commit so its head_sha is current.
func regenerateReportAtCommit(t *testing.T, f *acFixture, head string) {
	t.Helper()
	in := f.input()
	in.EvalCommit = head
	r, err := Build(in)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if _, err := WriteReportPair(EvidenceFor(f.CardDir, fixCard), r); err != nil {
		t.Fatalf("write pair: %v", err)
	}
}

// TestEvaluatePushStopAndClear pins the push evaluation end-to-end over the
// acceptance fixture: after the card branch merges into local develop the
// SPEC's own directory changed in the range, so the card is a candidate; an
// empty evidence set yields second_review_not_performed, and a performed
// pass review at the source commit clears it.
func TestEvaluatePushStopAndClear(t *testing.T) {
	f := newACFixture(t)
	f.Git(f.Root, "merge", "--no-ff", closuretestBranch())

	source := gitioHeadOfDir(f.Root)
	in := PushEvalInput{
		Tree: f.Root, Remote: "origin", Integration: "develop",
		Source: source, SecondReview: "required",
	}

	result := EvaluatePush(fixtureSeam(), in)
	if result.Undetermined {
		t.Fatalf("undetermined: %s", result.Cause)
	}
	if !slices.Contains(result.Evaluated, "SPEC-FIXTURE-001") {
		t.Fatalf("evaluated = %v, want the in-range contract", result.Evaluated)
	}
	if len(result.Results) != 1 || result.Results[0].SpecID != "SPEC-FIXTURE-001" ||
		!slices.Contains(result.Results[0].Codes, CodeSecondReviewNotPerformed) {
		t.Fatalf("results = %+v, want SPEC-FIXTURE-001 not-performed", result.Results)
	}

	// A performed pass review at the source commit plus a current report
	// clear the stop.
	secondReviewAtCommit(t, f, source, "pass")
	regenerateReportAtCommit(t, f, source)
	result = EvaluatePush(fixtureSeam(), in)
	if result.Undetermined || len(result.Results) != 0 {
		t.Fatalf("results = %+v undetermined=%v cause=%s, want ready", result.Results, result.Undetermined, result.Cause)
	}
}

// TestEvaluatePushUndeterminedWithoutRemoteRef pins REQ-CLOSURE-017's
// missing-remote-ref path.
func TestEvaluatePushUndeterminedWithoutRemoteRef(t *testing.T) {
	f := newACFixture(t)
	f.Git(f.Root, "merge", "--no-ff", closuretestBranch())
	f.Git(f.Root, "branch", "-dr", "origin/develop")
	result := EvaluatePush(fixtureSeam(), PushEvalInput{
		Tree: f.Root, Remote: "origin", Integration: "develop",
		Source: gitioHeadOfDir(f.Root), SecondReview: "required",
	})
	if !result.Undetermined {
		t.Fatalf("results = %+v, want undetermined", result)
	}
}

// TestGitioTreeReaders covers the tree-read helpers the push evaluation and
// the CLI report command share.
func TestGitioTreeReaders(t *testing.T) {
	f := newACFixture(t)
	data, err := gitio.Blob(f.CardDir, "HEAD", ".moai/specs/SPEC-FIXTURE-001/contract.yaml")
	if err != nil || !contains(data, "card: c1") {
		t.Fatalf("blob = %q err %v", data, err)
	}
	if _, err := gitio.Blob(f.CardDir, "HEAD", ".moai/specs/SPEC-FIXTURE-001/nope.yaml"); err == nil {
		t.Fatalf("missing blob read succeeded")
	}
	paths, err := gitio.PathsUnder(f.CardDir, "HEAD", ".moai/specs/SPEC-FIXTURE-001")
	if err != nil || len(paths) < 3 {
		t.Fatalf("paths = %v err %v, want the SPEC files", paths, err)
	}
	branch, err := gitio.CurrentBranch(f.CardDir)
	if err != nil || branch == "" {
		t.Fatalf("current branch = %q err %v", branch, err)
	}
	up, err := gitio.UpstreamBranch(f.CardDir, "develop")
	if err != nil {
		t.Fatalf("upstream: %v", err)
	}
	if up != "origin/develop" {
		t.Fatalf("upstream = %q, want origin/develop", up)
	}
}
