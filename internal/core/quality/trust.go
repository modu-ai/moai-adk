// @MX:ANCHOR: [AUTO] TRUST 5 품질 게이트 프레임워크의 핵심 패키지입니다. 5가지 품질 원칙(Tested, Readable, Understandable, Secured, Trackable)을 자동 검증합니다.
// @MX:REASON: fan_in=15+, 모든 품질 검증의 진입점이며 시스템 전체에서 호출됩니다
// Package quality implements the TRUST 5 Quality Gates framework.
//
// It provides automated validation of five quality principles:
// Tested, Readable, Understandable, Secured, and Trackable.
// The framework supports phase-specific thresholds, regression detection,
// and methodology-aware validation (DDD, TDD).
package quality

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"slices"
	"sync"
	"time"
)

// Principle name constants identify the five TRUST 5 quality pillars.
const (
	PrincipleTested         = "tested"
	PrincipleReadable       = "readable"
	PrincipleUnderstandable = "understandable"
	PrincipleSecured        = "secured"
	PrincipleTrackable      = "trackable"
)

// Severity level constants for Issue classification.
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityHint    = "hint"
	SeverityInfo    = "info"
)

// ValidPrinciples lists all supported principle names in canonical order.
var ValidPrinciples = []string{
	PrincipleTested,
	PrincipleReadable,
	PrincipleUnderstandable,
	PrincipleSecured,
	PrincipleTrackable,
}

// DevelopmentMode represents the configured development methodology.
type DevelopmentMode string

const (
	// ModeDDD uses Domain-Driven Development (ANALYZE-PRESERVE-IMPROVE).
	ModeDDD DevelopmentMode = "ddd"

	// ModeTDD uses Test-Driven Development (RED-GREEN-REFACTOR).
	ModeTDD DevelopmentMode = "tdd"
)

// ValidDevelopmentModes lists all supported development modes.
var ValidDevelopmentModes = []DevelopmentMode{ModeDDD, ModeTDD}

// IsValid checks whether the DevelopmentMode is a recognized value.
func (m DevelopmentMode) IsValid() bool {
	switch m {
	case ModeDDD, ModeTDD:
		return true
	}
	return false
}

// WorkflowPhase represents a workflow execution phase.
type WorkflowPhase string

const (
	// PhasePlan captures the LSP diagnostic baseline.
	PhasePlan WorkflowPhase = "plan"

	// PhaseRun enforces zero-tolerance quality thresholds.
	PhaseRun WorkflowPhase = "run"

	// PhaseSync enforces clean LSP state for documentation.
	PhaseSync WorkflowPhase = "sync"
)

// --- Core types ---

// Issue represents a single quality violation found during validation.
type Issue struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
}

// PrincipleResult is the validation result for a single TRUST principle.
type PrincipleResult struct {
	Name   string  `json:"name"`
	Passed bool    `json:"passed"`
	Score  float64 `json:"score"`
	Issues []Issue `json:"issues"`
}

// Report is the aggregated quality validation report.
type Report struct {
	Passed            bool                       `json:"passed"`
	Score             float64                    `json:"score"`
	DevelopmentMode   string                     `json:"development_mode"`
	Phase             string                     `json:"phase,omitempty"`
	Principles        map[string]PrincipleResult `json:"principles"`
	PhaseIssues       []Issue                    `json:"phase_issues,omitempty"`
	RegressionIssues  []Issue                    `json:"regression_issues,omitempty"`
	MethodologyIssues []Issue                    `json:"methodology_issues,omitempty"`
	Timestamp         time.Time                  `json:"timestamp"`
}

// AllIssues returns all issues from all sources in the report.
func (r *Report) AllIssues() []Issue {
	var all []Issue
	for _, pr := range r.Principles {
		all = append(all, pr.Issues...)
	}
	all = append(all, r.PhaseIssues...)
	all = append(all, r.RegressionIssues...)
	all = append(all, r.MethodologyIssues...)
	return all
}

// --- Diagnostic types ---

// Diagnostic represents an LSP diagnostic for quality analysis.
type Diagnostic struct {
	File     string
	Line     int
	Severity string // "error", "warning", "hint", "info"
	Message  string
	Source   string // "typecheck", "lint", "security"
	Code     string
}

// DiagnosticSnapshot captures diagnostic counts at a point in time.
type DiagnosticSnapshot struct {
	Errors           int
	Warnings         int
	TypeErrors       int
	LintErrors       int
	SecurityWarnings int
	Timestamp        time.Time
}

// ASTMatch represents a structural code pattern match.
type ASTMatch struct {
	File    string
	Line    int
	Pattern string
	Match   string
}

// --- Interfaces ---

// Gate defines the TRUST 5 quality gate interface.
type Gate interface {
	// Validate runs all 5 TRUST principles and returns an aggregated report.
	Validate(ctx context.Context) (*Report, error)

	// ValidatePrinciple runs a single TRUST principle by name.
	ValidatePrinciple(ctx context.Context, principle string) (*PrincipleResult, error)
}

// Validator is the internal interface for individual principle validators.
type Validator interface {
	// Name returns the principle name this validator checks.
	Name() string

	// Validate runs the principle check and returns a result.
	Validate(ctx context.Context) (*PrincipleResult, error)
}

// LSPClient abstracts LSP diagnostic collection for testability.
type LSPClient interface {
	CollectDiagnostics(ctx context.Context) ([]Diagnostic, error)
}

// GitManager abstracts Git operations for trackable validation.
type GitManager interface {
	LastCommitMessage(ctx context.Context) (string, error)
	DiagnosticHistory(ctx context.Context) ([]DiagnosticSnapshot, error)
}

// ASTAnalyzer abstracts AST pattern matching for code analysis.
type ASTAnalyzer interface {
	Analyze(ctx context.Context, patterns []string) ([]ASTMatch, error)
}

// --- Configuration types ---

// QualityConfig holds quality gate configuration.
type QualityConfig struct {
	DevelopmentMode     DevelopmentMode
	EnforceQuality      bool
	TestCoverageTarget  int
	LSPGates            PhaseGates
	RegressionDetection RegressionConfig
	DDDSettings         DDDSettings
	TDDSettings         TDDSettings
	CacheTTL            time.Duration
	Timeout             time.Duration
}

// PhaseGates contains phase-specific quality thresholds.
type PhaseGates struct {
	Plan PlanGate
	Run  RunGate
	Sync SyncGate
}

// PlanGate defines plan phase quality gate thresholds.
type PlanGate struct {
	RequireBaseline bool
}

// RunGate defines run phase quality gate thresholds.
type RunGate struct {
	MaxErrors       int
	MaxTypeErrors   int
	MaxLintErrors   int
	AllowRegression bool
}

// SyncGate defines sync phase quality gate thresholds.
type SyncGate struct {
	MaxErrors       int
	MaxWarnings     int
	RequireCleanLSP bool
}

// RegressionConfig contains regression detection thresholds.
type RegressionConfig struct {
	ErrorIncreaseThreshold     int
	WarningIncreaseThreshold   int
	TypeErrorIncreaseThreshold int
}

// DDDSettings holds DDD-specific quality gate configuration.
type DDDSettings struct {
	RequireExistingTests  bool
	CharacterizationTests bool
	BehaviorSnapshots     bool
	MaxTransformationSize string
	PreserveBeforeImprove bool
}

// TDDSettings holds TDD-specific quality gate configuration.
type TDDSettings struct {
	MinCoveragePerCommit   int
	RequireTestFirst       bool
	MutationTestingEnabled bool
	MutationScoreThreshold int
}

// MethodologyContext provides methodology-specific validation inputs.
type MethodologyContext struct {
	// DDD-specific
	CharacterizationTestsExist bool
	PreserveStepCompleted      bool
	BehaviorSnapshotRegressed  bool

	// TDD-specific
	TestFirstVerified          bool
	CommitCoverage             int
	CoverageExemptionRequested bool

	// Transition tracking
	PreviousMode DevelopmentMode
}

// --- Principle weights ---

// principleWeights defines the scoring weight for each TRUST principle.
var principleWeights = map[string]float64{
	PrincipleTested:         0.30,
	PrincipleReadable:       0.15,
	PrincipleUnderstandable: 0.15,
	PrincipleSecured:        0.25,
	PrincipleTrackable:      0.15,
}

// passScoreThreshold is the minimum overall score for the report to pass.
const passScoreThreshold = 0.7

// conventionalCommitRegex matches Conventional Commits format.
var conventionalCommitRegex = regexp.MustCompile(
	`^(feat|fix|build|chore|ci|docs|style|refactor|perf|test)(\([a-zA-Z0-9_./-]+\))?!?: .+$`,
)

// --- Default configuration ---

// DefaultQualityConfig returns a QualityConfig with default values
// matching the .moai/config/sections/quality.yaml specification.
func DefaultQualityConfig() QualityConfig {
	return QualityConfig{
		DevelopmentMode:    ModeDDD,
		EnforceQuality:     true,
		TestCoverageTarget: 85,
		LSPGates: PhaseGates{
			Plan: PlanGate{RequireBaseline: true},
			Run:  RunGate{MaxErrors: 0, MaxTypeErrors: 0, MaxLintErrors: 0, AllowRegression: false},
			Sync: SyncGate{MaxErrors: 0, MaxWarnings: 10, RequireCleanLSP: true},
		},
		RegressionDetection: RegressionConfig{
			ErrorIncreaseThreshold:     0,
			WarningIncreaseThreshold:   10,
			TypeErrorIncreaseThreshold: 0,
		},
		DDDSettings: DDDSettings{
			RequireExistingTests:  true,
			CharacterizationTests: true,
			BehaviorSnapshots:     true,
			MaxTransformationSize: "small",
			PreserveBeforeImprove: true,
		},
		TDDSettings: TDDSettings{
			MinCoveragePerCommit:   80,
			RequireTestFirst:       true,
			MutationTestingEnabled: false,
		},
		CacheTTL: 5 * time.Second,
		Timeout:  3 * time.Second,
	}
}

// --- TrustGate implementation ---

// TrustGate implements the Gate interface, orchestrating TRUST 5 validation.
type TrustGate struct {
	validators     map[string]Validator
	config         QualityConfig
	phase          WorkflowPhase
	baseline       *DiagnosticSnapshot
	lsp            LSPClient
	methodologyCtx *MethodologyContext
	mu             sync.RWMutex
	logger         *slog.Logger
}

// TrustGateOption is a functional option for configuring TrustGate.
type TrustGateOption func(*TrustGate)

// WithPhase sets the workflow phase for phase-specific validation.
func WithPhase(phase WorkflowPhase) TrustGateOption {
	return func(g *TrustGate) {
		g.phase = phase
	}
}

// WithBaseline sets an existing diagnostic baseline for regression detection.
func WithBaseline(baseline *DiagnosticSnapshot) TrustGateOption {
	return func(g *TrustGate) {
		g.baseline = baseline
	}
}

// WithLSPClient sets the LSP client for phase validation and regression detection.
func WithLSPClient(lsp LSPClient) TrustGateOption {
	return func(g *TrustGate) {
		g.lsp = lsp
	}
}

// WithMethodologyContext sets methodology-specific validation context.
func WithMethodologyContext(mCtx *MethodologyContext) TrustGateOption {
	return func(g *TrustGate) {
		g.methodologyCtx = mCtx
	}
}

// WithLogger sets a custom structured logger.
func WithLogger(logger *slog.Logger) TrustGateOption {
	return func(g *TrustGate) {
		g.logger = logger
	}
}

// NewTrustGate creates a new TrustGate with the given configuration and validators.
func NewTrustGate(config QualityConfig, validators []Validator, opts ...TrustGateOption) *TrustGate {
	vMap := make(map[string]Validator, len(validators))
	for _, v := range validators {
		vMap[v.Name()] = v
	}

	g := &TrustGate{
		validators: vMap,
		config:     config,
		logger:     slog.Default(),
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// Compile-time interface compliance check.
var _ Gate = (*TrustGate)(nil)

// @MX:WARN: [AUTO] validators 수만큼 고루틴을 생성하여 동시에 실행합니다. 고루틴 누수 가능성을 제어해야 합니다.
// @MX:REASON: [AUTO] validators 길이만큼 고루틴이 생성되어 리소스 부하 가능성
// Validate runs all TRUST 5 principle validators and returns an aggregated report.
// Validators are executed concurrently. If context is cancelled, partial results
// are returned along with the context error.
func (g *TrustGate) Validate(ctx context.Context) (*Report, error) {
	g.logger.Info("quality gate validation started",
		"phase", string(g.phase),
		"principles", len(g.validators),
		"development_mode", string(g.config.DevelopmentMode),
	)

	type validationResult struct {
		name   string
		result *PrincipleResult
		err    error
	}

	resultCh := make(chan validationResult, len(g.validators))

	for name, v := range g.validators {
		go func(name string, v Validator) {
			result, err := v.Validate(ctx)
			resultCh <- validationResult{name: name, result: result, err: err}
		}(name, v)
	}

	results := make(map[string]*PrincipleResult)
	var firstErr error

	for i := 0; i < len(g.validators); i++ {
		select {
		case <-ctx.Done():
			report := g.buildReport(results)
			report.Phase = string(g.phase)
			return report, ctx.Err()
		case vr := <-resultCh:
			if vr.err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("quality: %s: %w", vr.name, vr.err)
				}
			} else if vr.result != nil {
				results[vr.name] = vr.result
			}
		}
	}

	report := g.buildReport(results)
	report.Phase = string(g.phase)

	// Collect diagnostic snapshot for phase validation and regression detection.
	var snapshot *DiagnosticSnapshot
	if g.lsp != nil && (g.phase != "" || g.baseline != nil) {
		var err error
		snapshot, err = g.collectSnapshot(ctx)
		if err != nil {
			g.logger.Error("failed to collect diagnostic snapshot", "error", err)
		}
	}

	// Phase-specific validation.
	if g.phase != "" && snapshot != nil {
		phaseIssues := g.validatePhase(snapshot)
		if len(phaseIssues) > 0 {
			report.PhaseIssues = phaseIssues
			for _, issue := range phaseIssues {
				if issue.Severity == SeverityError {
					report.Passed = false
					break
				}
			}
		}
	}

	// Regression detection.
	if g.baseline != nil {
		if snapshot != nil {
			regressionIssues := g.detectRegression(snapshot)
			if len(regressionIssues) > 0 {
				report.RegressionIssues = regressionIssues
				for _, issue := range regressionIssues {
					if issue.Severity == SeverityError {
						report.Passed = false
						break
					}
				}
			}
		}
	} else {
		g.logger.Info("no baseline available, skipping regression detection")
	}

	// Methodology-specific validation.
	if g.methodologyCtx != nil {
		methIssues := g.validateMethodology()
		if len(methIssues) > 0 {
			report.MethodologyIssues = methIssues
			for _, issue := range methIssues {
				if issue.Severity == SeverityError {
					report.Passed = false
					break
				}
			}
		}
	}

	if firstErr != nil {
		return report, firstErr
	}

	return report, nil
}

// ValidatePrinciple runs a single TRUST principle by name and returns its result.
// Returns an error if the principle name is not recognized.
func (g *TrustGate) ValidatePrinciple(ctx context.Context, principle string) (*PrincipleResult, error) {
	if !isValidPrinciple(principle) {
		return nil, fmt.Errorf("quality: unknown principle %q, valid principles: %v", principle, ValidPrinciples)
	}

	v, ok := g.validators[principle]
	if !ok {
		return nil, fmt.Errorf("quality: no validator registered for principle %q", principle)
	}

	return v.Validate(ctx)
}

// Baseline returns the current diagnostic baseline, if any.
func (g *TrustGate) Baseline() *DiagnosticSnapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.baseline
}

// Config returns a copy of the quality configuration (read-only access).
func (g *TrustGate) Config() QualityConfig {
	return g.config
}

// --- Internal methods ---

// collectSnapshot gathers diagnostic counts from the LSP client.
func (g *TrustGate) collectSnapshot(ctx context.Context) (*DiagnosticSnapshot, error) {
	diagnostics, err := g.lsp.CollectDiagnostics(ctx)
	if err != nil {
		return nil, fmt.Errorf("quality: collect diagnostics: %w", err)
	}

	snap := &DiagnosticSnapshot{Timestamp: time.Now().UTC()}
	for _, d := range diagnostics {
		switch d.Source {
		case "typecheck":
			snap.TypeErrors++
		case "lint":
			snap.LintErrors++
		case "security":
			snap.SecurityWarnings++
		}
		switch d.Severity {
		case SeverityError:
			snap.Errors++
		case SeverityWarning:
			snap.Warnings++
		}
	}

	return snap, nil
}

// validatePhase applies phase-specific quality gates.
func (g *TrustGate) validatePhase(snapshot *DiagnosticSnapshot) []Issue {
	switch g.phase {
	case PhasePlan:
		return g.validatePlanPhase(snapshot)
	case PhaseRun:
		return g.validateRunPhase(snapshot)
	case PhaseSync:
		return g.validateSyncPhase(snapshot)
	default:
		return nil
	}
}

// validatePlanPhase captures the diagnostic baseline.
func (g *TrustGate) validatePlanPhase(snapshot *DiagnosticSnapshot) []Issue {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.baseline = snapshot

	g.logger.Info("plan phase baseline captured",
		"errors", snapshot.Errors,
		"warnings", snapshot.Warnings,
		"type_errors", snapshot.TypeErrors,
	)

	return nil
}

// validateRunPhase enforces zero-tolerance thresholds.
func (g *TrustGate) validateRunPhase(snapshot *DiagnosticSnapshot) []Issue {
	var issues []Issue
	gate := g.config.LSPGates.Run

	if snapshot.Errors > gate.MaxErrors {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("run phase requires zero errors, found %d", snapshot.Errors),
			Rule:     "phase-run-errors",
		})
	}

	if snapshot.TypeErrors > gate.MaxTypeErrors {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("run phase requires zero type errors, found %d", snapshot.TypeErrors),
			Rule:     "phase-run-type-errors",
		})
	}

	if snapshot.LintErrors > gate.MaxLintErrors {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("run phase requires zero lint errors, found %d", snapshot.LintErrors),
			Rule:     "phase-run-lint-errors",
		})
	}

	return issues
}

// validateSyncPhase enforces clean LSP state.
func (g *TrustGate) validateSyncPhase(snapshot *DiagnosticSnapshot) []Issue {
	var issues []Issue
	gate := g.config.LSPGates.Sync

	if snapshot.Errors > gate.MaxErrors {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("sync phase requires zero errors, found %d", snapshot.Errors),
			Rule:     "phase-sync-errors",
		})
	}

	if snapshot.Warnings > gate.MaxWarnings {
		issues = append(issues, Issue{
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("sync phase allows maximum %d warnings, found %d", gate.MaxWarnings, snapshot.Warnings),
			Rule:     "phase-sync-warnings",
		})
	}

	return issues
}

// detectRegression compares current diagnostics against the baseline.
func (g *TrustGate) detectRegression(snapshot *DiagnosticSnapshot) []Issue {
	var issues []Issue
	reg := g.config.RegressionDetection

	errorIncrease := snapshot.Errors - g.baseline.Errors
	if errorIncrease > reg.ErrorIncreaseThreshold {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message: fmt.Sprintf("error count increased from %d to %d (threshold: %d)",
				g.baseline.Errors, snapshot.Errors, reg.ErrorIncreaseThreshold),
			Rule: "regression-errors",
		})
	}

	warningIncrease := snapshot.Warnings - g.baseline.Warnings
	if warningIncrease > reg.WarningIncreaseThreshold {
		issues = append(issues, Issue{
			Severity: SeverityWarning,
			Message: fmt.Sprintf("warning count increased from %d to %d (threshold: %d)",
				g.baseline.Warnings, snapshot.Warnings, reg.WarningIncreaseThreshold),
			Rule: "regression-warnings",
		})
	}

	typeErrorIncrease := snapshot.TypeErrors - g.baseline.TypeErrors
	if typeErrorIncrease > reg.TypeErrorIncreaseThreshold {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message: fmt.Sprintf("type error count increased from %d to %d (threshold: %d)",
				g.baseline.TypeErrors, snapshot.TypeErrors, reg.TypeErrorIncreaseThreshold),
			Rule: "regression-type-errors",
		})
	}

	return issues
}

// validateMethodology applies development-mode-specific quality rules.
func (g *TrustGate) validateMethodology() []Issue {
	if g.methodologyCtx == nil {
		return nil
	}

	// Detect mode transition.
	if g.methodologyCtx.PreviousMode != "" && g.methodologyCtx.PreviousMode != g.config.DevelopmentMode {
		g.logger.Warn("development mode changed",
			"from", string(g.methodologyCtx.PreviousMode),
			"to", string(g.config.DevelopmentMode),
		)
	}

	if !g.config.DevelopmentMode.IsValid() {
		return []Issue{{
			Severity: SeverityError,
			Message: fmt.Sprintf(
				"unknown development mode %q, valid modes: ddd, tdd",
				g.config.DevelopmentMode,
			),
			Rule: "methodology-invalid-mode",
		}}
	}

	switch g.config.DevelopmentMode {
	case ModeDDD:
		return g.validateDDDMode()
	case ModeTDD:
		return g.validateTDDMode()
	default:
		return nil
	}
}

// validateDDDMode enforces DDD-specific quality rules.
func (g *TrustGate) validateDDDMode() []Issue {
	var issues []Issue
	mCtx := g.methodologyCtx

	if g.config.DDDSettings.CharacterizationTests && !mCtx.CharacterizationTestsExist {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message:  "characterization tests required for modified files in DDD mode",
			Rule:     "ddd-characterization",
		})
	}

	if g.config.DDDSettings.PreserveBeforeImprove && !mCtx.PreserveStepCompleted {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message:  "PRESERVE step must complete before IMPROVE",
			Rule:     "ddd-preserve-before-improve",
		})
	}

	if g.config.DDDSettings.BehaviorSnapshots && mCtx.BehaviorSnapshotRegressed {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message:  "behavior snapshot regression detected",
			Rule:     "ddd-behavior-snapshot",
		})
	}

	return issues
}

// validateTDDMode enforces TDD-specific quality rules.
func (g *TrustGate) validateTDDMode() []Issue {
	var issues []Issue
	mCtx := g.methodologyCtx

	if g.config.TDDSettings.RequireTestFirst && !mCtx.TestFirstVerified {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message:  "tests must be written before implementation code in TDD mode",
			Rule:     "tdd-test-first",
		})
	}

	if mCtx.CoverageExemptionRequested {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message:  "coverage exemptions not allowed in TDD mode",
			Rule:     "tdd-no-exemption",
		})
	}

	minCov := g.config.TDDSettings.MinCoveragePerCommit
	if minCov > 0 && mCtx.CommitCoverage < minCov {
		issues = append(issues, Issue{
			Severity: SeverityError,
			Message: fmt.Sprintf("commit coverage %d%% is below TDD minimum %d%%",
				mCtx.CommitCoverage, minCov),
			Rule: "tdd-min-coverage",
		})
	}

	return issues
}

// buildReport aggregates validator results into a Report.
func (g *TrustGate) buildReport(results map[string]*PrincipleResult) *Report {
	report := &Report{
		DevelopmentMode: string(g.config.DevelopmentMode),
		Principles:      make(map[string]PrincipleResult, len(ValidPrinciples)),
		Timestamp:       time.Now().UTC(),
	}

	var totalScore float64
	hasErrorIssues := false

	for _, name := range ValidPrinciples {
		if result, ok := results[name]; ok {
			report.Principles[name] = *result
			weight := principleWeights[name]
			totalScore += result.Score * weight
			for _, issue := range result.Issues {
				if issue.Severity == SeverityError {
					hasErrorIssues = true
				}
			}
		} else {
			report.Principles[name] = PrincipleResult{
				Name:   name,
				Passed: false,
				Score:  0.0,
				Issues: []Issue{},
			}
		}
	}

	report.Score = math.Round(totalScore*1000) / 1000
	report.Passed = report.Score >= passScoreThreshold && !hasErrorIssues

	return report
}

// --- Helper functions ---

// isValidPrinciple checks whether the given name is a valid TRUST principle.
func isValidPrinciple(name string) bool {
	return slices.Contains(ValidPrinciples, name)
}

// IsConventionalCommit checks whether a commit message follows Conventional Commits format.
func IsConventionalCommit(msg string) bool {
	return conventionalCommitRegex.MatchString(msg)
}
