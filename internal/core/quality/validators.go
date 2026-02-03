package quality

import (
	"context"
	"fmt"
	"math"
)

// --- TestedValidator ---

// TestedValidator checks the "tested" principle:
// unit test pass status, LSP type/general errors, and test coverage.
type TestedValidator struct {
	lsp             LSPClient
	coverageTarget  int
	currentCoverage int
}

// NewTestedValidator creates a validator for the Tested principle.
// coverageTarget is the minimum required percentage (e.g., 85).
// currentCoverage is the current test coverage percentage.
func NewTestedValidator(lsp LSPClient, coverageTarget, currentCoverage int) *TestedValidator {
	return &TestedValidator{
		lsp:             lsp,
		coverageTarget:  coverageTarget,
		currentCoverage: currentCoverage,
	}
}

// Name returns the principle name.
func (v *TestedValidator) Name() string { return PrincipleTested }

// Validate checks type errors, general errors, and test coverage.
func (v *TestedValidator) Validate(ctx context.Context) (*PrincipleResult, error) {
	result := &PrincipleResult{
		Name:   PrincipleTested,
		Passed: true,
		Score:  1.0,
		Issues: []Issue{},
	}

	diagnostics, err := v.lsp.CollectDiagnostics(ctx)
	if err != nil {
		return nil, fmt.Errorf("tested: collect diagnostics: %w", err)
	}

	var typeErrors, generalErrors int
	for _, d := range diagnostics {
		if d.Source == "typecheck" {
			typeErrors++
			result.Issues = append(result.Issues, Issue{
				File:     d.File,
				Line:     d.Line,
				Severity: SeverityError,
				Message:  d.Message,
				Rule:     "type-error",
			})
		} else if d.Severity == SeverityError {
			generalErrors++
			result.Issues = append(result.Issues, Issue{
				File:     d.File,
				Line:     d.Line,
				Severity: SeverityError,
				Message:  d.Message,
				Rule:     "error",
			})
		}
	}

	// Check coverage threshold.
	if v.coverageTarget > 0 && v.currentCoverage < v.coverageTarget {
		result.Issues = append(result.Issues, Issue{
			Severity: SeverityError,
			Message: fmt.Sprintf("test coverage %d%% is below target %d%%",
				v.currentCoverage, v.coverageTarget),
			Rule: "coverage-threshold",
		})
	}

	// Calculate score: 3 checks weighted equally.
	var score float64
	checks := 3.0

	if typeErrors == 0 {
		score += 1.0
	}
	if generalErrors == 0 {
		score += 1.0
	}
	if v.coverageTarget > 0 && v.currentCoverage >= v.coverageTarget {
		score += 1.0
	} else if v.coverageTarget > 0 {
		score += math.Min(1.0, float64(v.currentCoverage)/float64(v.coverageTarget))
	} else {
		score += 1.0 // No coverage target means this check passes.
	}

	result.Score = math.Round((score/checks)*1000) / 1000
	result.Passed = typeErrors == 0 && generalErrors == 0 &&
		(v.coverageTarget == 0 || v.currentCoverage >= v.coverageTarget)

	return result, nil
}

// --- ReadableValidator ---

// ReadableValidator checks the "readable" principle:
// naming convention compliance and LSP lint errors.
type ReadableValidator struct {
	lsp LSPClient
}

// NewReadableValidator creates a validator for the Readable principle.
func NewReadableValidator(lsp LSPClient) *ReadableValidator {
	return &ReadableValidator{lsp: lsp}
}

// Name returns the principle name.
func (v *ReadableValidator) Name() string { return PrincipleReadable }

// Validate checks for lint errors from the LSP.
func (v *ReadableValidator) Validate(ctx context.Context) (*PrincipleResult, error) {
	result := &PrincipleResult{
		Name:   PrincipleReadable,
		Passed: true,
		Score:  1.0,
		Issues: []Issue{},
	}

	diagnostics, err := v.lsp.CollectDiagnostics(ctx)
	if err != nil {
		return nil, fmt.Errorf("readable: collect diagnostics: %w", err)
	}

	var lintErrors int
	for _, d := range diagnostics {
		if d.Source == "lint" {
			lintErrors++
			result.Issues = append(result.Issues, Issue{
				File:     d.File,
				Line:     d.Line,
				Severity: SeverityError,
				Message:  d.Message,
				Rule:     d.Code,
			})
		}
	}

	if lintErrors > 0 {
		result.Passed = false
		result.Score = math.Max(0, 1.0-float64(lintErrors)*0.1)
		result.Score = math.Round(result.Score*1000) / 1000
	}

	return result, nil
}

// --- UnderstandableValidator ---

// UnderstandableValidator checks the "understandable" principle:
// documentation completeness, code complexity, and LSP warnings.
type UnderstandableValidator struct {
	lsp              LSPClient
	warningThreshold int
	docComplete      bool
	complexityOK     bool
}

// NewUnderstandableValidator creates a validator for the Understandable principle.
// warningThreshold is the maximum allowed warning count (e.g., 10).
func NewUnderstandableValidator(lsp LSPClient, warningThreshold int, docComplete, complexityOK bool) *UnderstandableValidator {
	return &UnderstandableValidator{
		lsp:              lsp,
		warningThreshold: warningThreshold,
		docComplete:      docComplete,
		complexityOK:     complexityOK,
	}
}

// Name returns the principle name.
func (v *UnderstandableValidator) Name() string { return PrincipleUnderstandable }

// Validate checks warnings, documentation, and complexity.
func (v *UnderstandableValidator) Validate(ctx context.Context) (*PrincipleResult, error) {
	result := &PrincipleResult{
		Name:   PrincipleUnderstandable,
		Passed: true,
		Score:  1.0,
		Issues: []Issue{},
	}

	diagnostics, err := v.lsp.CollectDiagnostics(ctx)
	if err != nil {
		return nil, fmt.Errorf("understandable: collect diagnostics: %w", err)
	}

	var warnings int
	for _, d := range diagnostics {
		if d.Severity == SeverityWarning && d.Source != "security" {
			warnings++
		}
	}

	// Calculate score from 3 checks.
	var score float64
	checks := 3.0

	// Check 1: Warnings within threshold.
	if v.warningThreshold > 0 && warnings > v.warningThreshold {
		result.Issues = append(result.Issues, Issue{
			Severity: SeverityWarning,
			Message: fmt.Sprintf("warning count %d exceeds threshold %d",
				warnings, v.warningThreshold),
			Rule: "warnings-threshold",
		})
		score += math.Max(0, 1.0-float64(warnings-v.warningThreshold)*0.05)
	} else {
		score += 1.0
	}

	// Check 2: Documentation completeness.
	if v.docComplete {
		score += 1.0
	} else {
		result.Issues = append(result.Issues, Issue{
			Severity: SeverityWarning,
			Message:  "documentation is incomplete",
			Rule:     "doc-completeness",
		})
	}

	// Check 3: Code complexity.
	if v.complexityOK {
		score += 1.0
	} else {
		result.Issues = append(result.Issues, Issue{
			Severity: SeverityWarning,
			Message:  "code complexity exceeds acceptable threshold",
			Rule:     "complexity",
		})
	}

	result.Score = math.Round((score/checks)*1000) / 1000
	result.Passed = (v.warningThreshold == 0 || warnings <= v.warningThreshold) &&
		v.docComplete && v.complexityOK

	return result, nil
}

// --- SecuredValidator ---

// SecuredValidator checks the "secured" principle:
// security scan results and LSP security warnings.
type SecuredValidator struct {
	lsp LSPClient
}

// NewSecuredValidator creates a validator for the Secured principle.
func NewSecuredValidator(lsp LSPClient) *SecuredValidator {
	return &SecuredValidator{lsp: lsp}
}

// Name returns the principle name.
func (v *SecuredValidator) Name() string { return PrincipleSecured }

// Validate checks for security warnings from the LSP.
func (v *SecuredValidator) Validate(ctx context.Context) (*PrincipleResult, error) {
	result := &PrincipleResult{
		Name:   PrincipleSecured,
		Passed: true,
		Score:  1.0,
		Issues: []Issue{},
	}

	diagnostics, err := v.lsp.CollectDiagnostics(ctx)
	if err != nil {
		return nil, fmt.Errorf("secured: collect diagnostics: %w", err)
	}

	var securityWarnings int
	for _, d := range diagnostics {
		if d.Source == "security" {
			securityWarnings++
			result.Issues = append(result.Issues, Issue{
				File:     d.File,
				Line:     d.Line,
				Severity: SeverityError,
				Message:  d.Message,
				Rule:     d.Code,
			})
		}
	}

	if securityWarnings > 0 {
		result.Passed = false
		result.Score = math.Max(0, 1.0-float64(securityWarnings)*0.2)
		result.Score = math.Round(result.Score*1000) / 1000
	}

	return result, nil
}

// --- TrackableValidator ---

// TrackableValidator checks the "trackable" principle:
// conventional commit messages, structured logging, and diagnostic history tracking.
type TrackableValidator struct {
	git            GitManager
	structuredLogs bool
	diagTracked    bool
}

// NewTrackableValidator creates a validator for the Trackable principle.
func NewTrackableValidator(git GitManager, structuredLogs, diagTracked bool) *TrackableValidator {
	return &TrackableValidator{
		git:            git,
		structuredLogs: structuredLogs,
		diagTracked:    diagTracked,
	}
}

// Name returns the principle name.
func (v *TrackableValidator) Name() string { return PrincipleTrackable }

// Validate checks commit message format, structured logging, and diagnostic tracking.
func (v *TrackableValidator) Validate(ctx context.Context) (*PrincipleResult, error) {
	result := &PrincipleResult{
		Name:   PrincipleTrackable,
		Passed: true,
		Score:  1.0,
		Issues: []Issue{},
	}

	// Calculate score from 3 checks.
	var score float64
	checks := 3.0

	// Check 1: Conventional commit message.
	commitMsg, err := v.git.LastCommitMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("trackable: get commit message: %w", err)
	}

	if IsConventionalCommit(commitMsg) {
		score += 1.0
	} else {
		result.Issues = append(result.Issues, Issue{
			Severity: SeverityError,
			Message:  "commit message does not follow Conventional Commits format",
			Rule:     "conventional-commits",
		})
	}

	// Check 2: Structured logging.
	if v.structuredLogs {
		score += 1.0
	} else {
		result.Issues = append(result.Issues, Issue{
			Severity: SeverityWarning,
			Message:  "structured logging (slog) is not being used",
			Rule:     "structured-logs",
		})
	}

	// Check 3: Diagnostic history tracking.
	if v.diagTracked {
		score += 1.0
	} else {
		result.Issues = append(result.Issues, Issue{
			Severity: SeverityWarning,
			Message:  "LSP diagnostic history is not being tracked",
			Rule:     "diagnostic-history",
		})
	}

	result.Score = math.Round((score/checks)*1000) / 1000
	result.Passed = IsConventionalCommit(commitMsg) && v.structuredLogs && v.diagTracked

	return result, nil
}

// Compile-time interface compliance checks.
var (
	_ Validator = (*TestedValidator)(nil)
	_ Validator = (*ReadableValidator)(nil)
	_ Validator = (*UnderstandableValidator)(nil)
	_ Validator = (*SecuredValidator)(nil)
	_ Validator = (*TrackableValidator)(nil)
)
