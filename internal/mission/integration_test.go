package mission

import "testing"

func TestIntegrateCardIntoLocalDevelop(t *testing.T) {
	in := CardIntegrationInput{
		OwnerRole: "manager-git", LauncherEntered: true, WorktreeBranch: "WT-gtd-autonomy",
		IntegrationWorktree: "/repo/.claude/worktrees/develop", IntegrationLeaseHeld: true,
		LocalDevelopBaseSHA: "base", CardHeadSHA: "card", ExplicitPaths: []string{"internal/mission/policy.go"},
		MergeStrategy: "--no-ff", LocalDevelopMergeSHA: "merge",
	}
	got, err := IntegrateCardIntoLocalDevelop(in)
	if err != nil || got.CardHeadSHA != "card" || got.LocalDevelopMergeSHA != "merge" {
		t.Fatalf("integration = %+v err=%v", got, err)
	}
	mutants := []CardIntegrationInput{in, in, in, in}
	mutants[0].LanePush = true
	mutants[1].CardPullRequest = true
	mutants[2].SweepStaging = true
	mutants[3].MergeStrategy = "--ff-only"
	for _, mutant := range mutants {
		if _, err := IntegrateCardIntoLocalDevelop(mutant); err == nil {
			t.Fatalf("unsafe integration allowed: %+v", mutant)
		}
	}
}

func TestValidateDevelopBatchMerge(t *testing.T) {
	in := DevelopBatchInput{
		OwnerRole: "manager-git", IntegrationLeaseHeld: true, LocalDevelopHeadSHA: "develop-1",
		LocalMergeSHAs: []string{"merge-1", "merge-2"}, BatchPushCount: 1, OriginDevelopSHA: "develop-1",
		CISHA: "develop-1", CIPassed: true, ReleaseHeadSHA: "develop-1", ReleaseTarget: "main",
		ReleaseAuditPassed: true, ReleaseReviewPassed: true, ReleaseCIPassed: true,
		MainLandedSHA: "main-2", MainContainsRelease: true, ProtectedMain: true,
	}
	if _, err := ValidateDevelopBatchMerge(in); err != nil {
		t.Fatal(err)
	}
	mutants := []DevelopBatchInput{in, in, in, in, in}
	mutants[0].OriginDevelopSHA = "stale"
	mutants[1].CISHA = "other"
	mutants[2].ReleaseHeadSHA = "other"
	mutants[3].ReleaseReviewPassed = false
	mutants[4].MainContainsRelease = false
	for _, mutant := range mutants {
		if _, err := ValidateDevelopBatchMerge(mutant); err == nil {
			t.Fatalf("stale delivery allowed: %+v", mutant)
		}
	}
}
