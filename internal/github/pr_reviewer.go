package github

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/modu-ai/moai-adk/internal/core/quality"
)

// ReviewDecision represents the outcome of a PR review.
type ReviewDecision string

const (
	// ReviewApprove indicates the PR passes all quality checks.
	ReviewApprove ReviewDecision = "APPROVE"

	// ReviewRequestChanges indicates the PR requires changes.
	ReviewRequestChanges ReviewDecision = "REQUEST_CHANGES"

	// ReviewComment indicates the review has observations but no blocking issues.
	ReviewComment ReviewDecision = "COMMENT"
)

// ReviewReport holds the results of an automated PR review.
type ReviewReport struct {
	// PRNumber is the pull request number.
	PRNumber int

	// Decision is the review outcome (APPROVE, REQUEST_CHANGES, COMMENT).
	Decision ReviewDecision

	// QualityReport is the TRUST 5 quality validation result.
	QualityReport *quality.Report

	// CheckStatus is the CI/CD pipeline check result.
	CheckStatus *CheckStatus

	// Summary is a human-readable review summary.
	Summary string

	// Issues lists specific problems found during review.
	Issues []string
}

// PRReviewer generates quality-based PR review reports.
type PRReviewer interface {
	// Review runs quality validation and CI checks, then generates a review decision.
	Review(ctx context.Context, prNumber int, specID string) (*ReviewReport, error)
}

// prReviewer implements PRReviewer.
type prReviewer struct {
	gh          GHClient
	qualityGate quality.Gate
	logger      *slog.Logger
}

// Compile-time interface compliance check.
var _ PRReviewer = (*prReviewer)(nil)

// NewPRReviewer creates a PR reviewer that validates quality gates and CI status.
func NewPRReviewer(gh GHClient, qualityGate quality.Gate, logger *slog.Logger) *prReviewer {
	if logger == nil {
		logger = slog.Default()
	}
	return &prReviewer{
		gh:          gh,
		qualityGate: qualityGate,
		logger:      logger.With("module", "pr-reviewer"),
	}
}

// Review analyzes a PR and returns a review report with a decision.
func (r *prReviewer) Review(ctx context.Context, prNumber int, specID string) (*ReviewReport, error) {
	r.logger.Info("starting PR review", "pr", prNumber, "spec_id", specID)

	report := &ReviewReport{
		PRNumber: prNumber,
		Decision: ReviewApprove,
		Issues:   []string{},
	}

	// Verify PR exists.
	prDetails, err := r.gh.PRView(ctx, prNumber)
	if err != nil {
		return nil, fmt.Errorf("view PR #%d: %w", prNumber, err)
	}

	if prDetails.State != "OPEN" {
		return nil, fmt.Errorf("PR #%d is %s, not OPEN", prNumber, prDetails.State)
	}

	// Run quality validation.
	qualityReport, qualityErr := r.qualityGate.Validate(ctx)
	report.QualityReport = qualityReport

	if qualityErr != nil {
		report.Issues = append(report.Issues,
			fmt.Sprintf("quality validation error: %v", qualityErr))
		report.Decision = ReviewRequestChanges
	} else if qualityReport != nil && !qualityReport.Passed {
		report.Decision = ReviewRequestChanges
		for _, issue := range qualityReport.AllIssues() {
			if issue.Severity == quality.SeverityError {
				report.Issues = append(report.Issues,
					fmt.Sprintf("[%s] %s:%d: %s", issue.Rule, issue.File, issue.Line, issue.Message))
			}
		}
	}

	// Check CI/CD status.
	checkStatus, checkErr := r.gh.PRChecks(ctx, prNumber)
	report.CheckStatus = checkStatus

	if checkErr != nil {
		report.Issues = append(report.Issues,
			fmt.Sprintf("CI check error: %v", checkErr))
		if report.Decision != ReviewRequestChanges {
			report.Decision = ReviewComment
		}
	} else if checkStatus != nil {
		switch checkStatus.Overall {
		case CheckFail:
			report.Decision = ReviewRequestChanges
			report.Issues = append(report.Issues, "CI/CD checks failed")
			for _, check := range checkStatus.Checks {
				if check.Conclusion == "failure" || check.Conclusion == "cancelled" {
					report.Issues = append(report.Issues,
						fmt.Sprintf("check %q: %s", check.Name, check.Conclusion))
				}
			}
		case CheckPending:
			if report.Decision == ReviewApprove {
				report.Decision = ReviewComment
				report.Issues = append(report.Issues, "CI/CD checks still pending")
			}
		}
	}

	// Generate summary.
	report.Summary = r.buildSummary(report)

	r.logger.Info("PR review complete",
		"pr", prNumber,
		"decision", string(report.Decision),
		"issues", len(report.Issues),
	)

	return report, nil
}

// buildSummary generates a human-readable review summary.
func (r *prReviewer) buildSummary(report *ReviewReport) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## PR #%d Review\n\n", report.PRNumber))

	switch report.Decision {
	case ReviewApprove:
		sb.WriteString("**Decision: APPROVE**\n\n")
		sb.WriteString("All quality gates passed and CI/CD checks are green.\n")
	case ReviewRequestChanges:
		sb.WriteString("**Decision: REQUEST CHANGES**\n\n")
		sb.WriteString("Issues found that must be resolved:\n")
	case ReviewComment:
		sb.WriteString("**Decision: COMMENT**\n\n")
		sb.WriteString("Review has observations:\n")
	}

	if len(report.Issues) > 0 {
		sb.WriteString("\n### Issues\n\n")
		for _, issue := range report.Issues {
			sb.WriteString(fmt.Sprintf("- %s\n", issue))
		}
	}

	if report.QualityReport != nil {
		sb.WriteString(fmt.Sprintf("\n### Quality Score: %.1f%%\n", report.QualityReport.Score*100))
	}

	return sb.String()
}
