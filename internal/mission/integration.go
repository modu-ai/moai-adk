package mission

import (
	"errors"
	"path/filepath"
	"strings"
)

type CardIntegrationInput struct {
	OwnerRole            string
	LauncherEntered      bool
	WorktreeBranch       string
	IntegrationWorktree  string
	IntegrationLeaseHeld bool
	LocalDevelopBaseSHA  string
	CardHeadSHA          string
	ExplicitPaths        []string
	SweepStaging         bool
	LanePush             bool
	CardPullRequest      bool
	MergeStrategy        string
	LocalDevelopMergeSHA string
}

type CardIntegrationReceipt struct {
	LocalDevelopBaseSHA  string
	CardHeadSHA          string
	LocalDevelopMergeSHA string
	OwnerRole            string
}

func validExplicitPath(path string) bool {
	clean := filepath.Clean(path)
	return path != "" && clean != "." && !filepath.IsAbs(clean) && clean != ".." && !strings.HasPrefix(filepath.ToSlash(clean), "../") && !strings.HasPrefix(filepath.ToSlash(clean), ".git/")
}

func IntegrateCardIntoLocalDevelop(in CardIntegrationInput) (CardIntegrationReceipt, error) {
	if in.OwnerRole != "manager-git" || !in.LauncherEntered || !strings.HasPrefix(in.WorktreeBranch, "WT-") {
		return CardIntegrationReceipt{}, errors.New("mission integration: owner_or_worktree_invalid")
	}
	if filepath.ToSlash(in.IntegrationWorktree) == "" || !strings.HasSuffix(filepath.ToSlash(in.IntegrationWorktree), "/.claude/worktrees/develop") || !in.IntegrationLeaseHeld {
		return CardIntegrationReceipt{}, errors.New("mission integration: integration_lease_missing")
	}
	if in.LocalDevelopBaseSHA == "" || in.CardHeadSHA == "" || in.LocalDevelopMergeSHA == "" {
		return CardIntegrationReceipt{}, errors.New("mission integration: sha_missing")
	}
	if len(in.ExplicitPaths) == 0 || in.SweepStaging {
		return CardIntegrationReceipt{}, errors.New("mission integration: explicit_staging_required")
	}
	for _, path := range in.ExplicitPaths {
		if !validExplicitPath(path) {
			return CardIntegrationReceipt{}, errors.New("mission integration: unsafe_path")
		}
	}
	if in.LanePush || in.CardPullRequest {
		return CardIntegrationReceipt{}, errors.New("mission integration: lane_push_or_card_pr_forbidden")
	}
	if in.MergeStrategy != "--no-ff" {
		return CardIntegrationReceipt{}, errors.New("mission integration: no_ff_required")
	}
	return CardIntegrationReceipt{LocalDevelopBaseSHA: in.LocalDevelopBaseSHA, CardHeadSHA: in.CardHeadSHA, LocalDevelopMergeSHA: in.LocalDevelopMergeSHA, OwnerRole: in.OwnerRole}, nil
}

type DevelopBatchInput struct {
	OwnerRole            string
	IntegrationLeaseHeld bool
	LocalDevelopHeadSHA  string
	LocalMergeSHAs       []string
	BatchPushCount       int
	OriginDevelopSHA     string
	CISHA                string
	CIPassed             bool
	ReleaseHeadSHA       string
	ReleaseTarget        string
	ReleaseAuditPassed   bool
	ReleaseReviewPassed  bool
	ReleaseCIPassed      bool
	MainLandedSHA        string
	MainContainsRelease  bool
	ProtectedMain        bool
}

type DevelopBatchReceipt struct {
	OriginDevelopSHA string
	ReleaseHeadSHA   string
	MainLandedSHA    string
}

func ValidateDevelopBatchMerge(in DevelopBatchInput) (DevelopBatchReceipt, error) {
	if in.OwnerRole != "manager-git" || !in.IntegrationLeaseHeld || len(in.LocalMergeSHAs) == 0 {
		return DevelopBatchReceipt{}, errors.New("mission delivery: owner_lease_or_merge_missing")
	}
	if in.BatchPushCount != 1 || in.LocalDevelopHeadSHA == "" || in.OriginDevelopSHA != in.LocalDevelopHeadSHA {
		return DevelopBatchReceipt{}, errors.New("mission delivery: origin_develop_sha_mismatch")
	}
	if !in.CIPassed || in.CISHA != in.OriginDevelopSHA {
		return DevelopBatchReceipt{}, errors.New("mission delivery: develop_ci_stale_or_failed")
	}
	if in.ReleaseHeadSHA != in.OriginDevelopSHA || in.ReleaseTarget != "main" || !in.ProtectedMain {
		return DevelopBatchReceipt{}, errors.New("mission delivery: release_lineage_invalid")
	}
	if !in.ReleaseAuditPassed || !in.ReleaseReviewPassed || !in.ReleaseCIPassed {
		return DevelopBatchReceipt{}, errors.New("mission delivery: release_gate_failed")
	}
	if in.MainLandedSHA == "" || !in.MainContainsRelease {
		return DevelopBatchReceipt{}, errors.New("mission delivery: main_landing_unverified")
	}
	return DevelopBatchReceipt{OriginDevelopSHA: in.OriginDevelopSHA, ReleaseHeadSHA: in.ReleaseHeadSHA, MainLandedSHA: in.MainLandedSHA}, nil
}
