package github

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// MergeOptions configures the merge operation.
type MergeOptions struct {
	// AutoMerge indicates whether --auto-merge was specified.
	AutoMerge bool

	// Method is the merge strategy (merge, squash, rebase).
	Method MergeMethod

	// DeleteBranch indicates whether to delete the branch after merge.
	DeleteBranch bool

	// RequireReview indicates whether review approval is required.
	RequireReview bool

	// RequireChecks indicates whether CI checks must pass.
	RequireChecks bool

	// SpecID is the SPEC identifier for quality validation context.
	SpecID string
}

// MergeResult summarizes the merge outcome.
type MergeResult struct {
	// Merged indicates whether the PR was actually merged.
	Merged bool

	// PRNumber is the pull request number.
	PRNumber int

	// Method is the merge strategy used.
	Method MergeMethod

	// BranchDeleted indicates whether the feature branch was deleted.
	BranchDeleted bool

	// MergedAt is the timestamp of the merge.
	MergedAt time.Time
}

// PrerequisiteCheck lists which merge conditions are met.
type PrerequisiteCheck struct {
	// AllMet indicates all prerequisites are satisfied.
	AllMet bool

	// AutoMergeFlag indicates --auto-merge was specified.
	AutoMergeFlag bool

	// ReviewApproved indicates the PR review approved the changes.
	ReviewApproved bool

	// ChecksPassed indicates CI/CD checks passed.
	ChecksPassed bool

	// QualityPassed indicates TRUST 5 quality gates passed.
	QualityPassed bool

	// Mergeable indicates no merge conflicts exist.
	Mergeable bool

	// FailureReasons lists the unmet prerequisites.
	FailureReasons []string
}

// PRMerger handles conditional PR merge operations.
type PRMerger interface {
	// Merge attempts to merge a PR if all conditions are met.
	Merge(ctx context.Context, prNumber int, opts MergeOptions) (*MergeResult, error)

	// CheckPrerequisites verifies all merge conditions without merging.
	CheckPrerequisites(ctx context.Context, prNumber int, opts MergeOptions) (*PrerequisiteCheck, error)
}

// prMerger implements PRMerger.
type prMerger struct {
	gh       GHClient
	reviewer PRReviewer
	logger   *slog.Logger
}

// Compile-time interface compliance check.
var _ PRMerger = (*prMerger)(nil)

// NewPRMerger creates a merger that checks prerequisites before merging.
// Returns an error if gh or reviewer is nil.
func NewPRMerger(gh GHClient, reviewer PRReviewer, logger *slog.Logger) (*prMerger, error) {
	if gh == nil {
		return nil, ErrNilGHClient
	}
	if reviewer == nil {
		return nil, ErrNilReviewer
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &prMerger{
		gh:       gh,
		reviewer: reviewer,
		logger:   logger.With("module", "pr-merger"),
	}, nil
}

// Merge attempts to merge a PR after verifying all prerequisites.
func (m *prMerger) Merge(ctx context.Context, prNumber int, opts MergeOptions) (*MergeResult, error) {
	m.logger.Info("starting merge attempt", "pr", prNumber, "auto_merge", opts.AutoMerge)

	// Auto-merge flag is mandatory.
	if !opts.AutoMerge {
		return nil, fmt.Errorf("PR #%d: %w", prNumber, ErrAutoMergeNotRequested)
	}

	// Check all prerequisites.
	prereqs, err := m.CheckPrerequisites(ctx, prNumber, opts)
	if err != nil {
		return nil, fmt.Errorf("check prerequisites for PR #%d: %w", prNumber, err)
	}

	if !prereqs.AllMet {
		return nil, fmt.Errorf("PR #%d has unmet prerequisites: %v: %w",
			prNumber, prereqs.FailureReasons, ErrMergeBlocked)
	}

	// Execute the merge.
	method := opts.Method
	if method == "" {
		method = MergeMethodMerge
	}

	if err := m.gh.PRMerge(ctx, prNumber, method, opts.DeleteBranch); err != nil {
		return nil, fmt.Errorf("merge PR #%d: %w", prNumber, err)
	}

	result := &MergeResult{
		Merged:        true,
		PRNumber:      prNumber,
		Method:        method,
		BranchDeleted: opts.DeleteBranch,
		MergedAt:      time.Now(),
	}

	m.logger.Info("pull request merged", "pr", prNumber, "method", string(method))

	return result, nil
}

// CheckPrerequisites verifies all merge conditions without actually merging.
// It fetches PR data once and passes it to the reviewer to avoid redundant API calls.
func (m *prMerger) CheckPrerequisites(ctx context.Context, prNumber int, opts MergeOptions) (*PrerequisiteCheck, error) {
	check := &PrerequisiteCheck{
		AllMet:         false,
		AutoMergeFlag:  opts.AutoMerge,
		FailureReasons: []string{},
	}

	// Check 1: Auto-merge flag.
	if !opts.AutoMerge {
		check.FailureReasons = append(check.FailureReasons, "--auto-merge flag not specified")
	}

	// Fetch PR details once for all downstream checks.
	prDetails, err := m.gh.PRView(ctx, prNumber)
	if err != nil {
		return nil, fmt.Errorf("view PR #%d: %w", prNumber, err)
	}

	// Check 2: Mergeability.
	switch prDetails.Mergeable {
	case "MERGEABLE":
		check.Mergeable = true
	case "CONFLICTING":
		check.FailureReasons = append(check.FailureReasons, "merge conflicts detected")
	default:
		// UNKNOWN or empty: treat as potentially mergeable.
		check.Mergeable = true
	}

	// Fetch CI/CD status once if required (shared with reviewer).
	var checkStatus *CheckStatus
	if opts.RequireChecks {
		checkStatus, err = m.gh.PRChecks(ctx, prNumber)
		if err != nil {
			check.FailureReasons = append(check.FailureReasons,
				fmt.Sprintf("CI check error: %v", err))
		} else if checkStatus != nil {
			check.ChecksPassed = checkStatus.Overall == CheckPass
			if !check.ChecksPassed {
				check.FailureReasons = append(check.FailureReasons,
					fmt.Sprintf("CI/CD status: %s", checkStatus.Overall))
			}
		}
	} else {
		// No checks required: mark as passed.
		check.ChecksPassed = true
	}

	// Check 3: Review via PRReviewer (pass pre-fetched data to avoid redundant calls).
	if opts.RequireReview {
		reviewReport, reviewErr := m.reviewer.Review(ctx, prNumber, opts.SpecID, &ReviewInput{
			PRDetails:   prDetails,
			CheckStatus: checkStatus,
		})
		if reviewErr != nil {
			check.FailureReasons = append(check.FailureReasons,
				fmt.Sprintf("review error: %v", reviewErr))
		} else {
			check.ReviewApproved = reviewReport.Decision == ReviewApprove
			if reviewReport.QualityReport != nil {
				check.QualityPassed = reviewReport.QualityReport.Passed
			}
			if !check.ReviewApproved {
				check.FailureReasons = append(check.FailureReasons,
					fmt.Sprintf("review decision: %s", reviewReport.Decision))
			}
		}
	} else {
		// No review required: mark as passed.
		check.ReviewApproved = true
		check.QualityPassed = true
	}

	// Determine overall result.
	check.AllMet = check.AutoMergeFlag &&
		check.ReviewApproved &&
		check.ChecksPassed &&
		check.QualityPassed &&
		check.Mergeable

	return check, nil
}
