// Package runtime provides core runtime utilities for MoAI workflow operations.
// Source: SPEC-WF-AUDIT-GATE-001 REQ-WAG-001..007
package runtime

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Verdict represents the outcome of a plan audit operation.
//
// REQ-WAG-001..007: four possible verdicts that guide Run phase gating.
type Verdict string

const (
	// VerdictPass means all must-pass criteria were met — proceed to Phase 1.
	VerdictPass Verdict = "PASS"

	// VerdictFail means one or more must-pass criteria failed — block Phase 1
	// unless grace window is active, in which case emit warning (FAIL_WARNED).
	VerdictFail Verdict = "FAIL"

	// VerdictFailWarned is FAIL during the 7-day grace window — warn only, no block.
	// REQ-WAG-002 / AC-WAG-08
	VerdictFailWarned Verdict = "FAIL_WARNED"

	// VerdictBypassed records an explicit --skip-audit bypass by the user.
	// REQ-WAG-006 / AC-WAG-06
	VerdictBypassed Verdict = "BYPASSED"

	// VerdictInconclusive records a timeout, error, or malformed response from plan-auditor.
	// REQ-WAG-007 / AC-WAG-07: never equivalent to PASS.
	VerdictInconclusive Verdict = "INCONCLUSIVE"
)

// GracePeriod is the 7-day window after SPEC-WF-AUDIT-GATE-001 merge
// during which FAIL verdicts emit warnings only (not blocking).
const GracePeriod = 7 * 24 * time.Hour

// EnvGateT0 is the environment variable for injecting a custom grace window
// start time in tests. Format: RFC3339 (ISO-8601).
const EnvGateT0 = "MOAI_AUDIT_GATE_T0"

// EnvSkipAudit is the environment variable to bypass the audit gate.
// REQ-WAG-006
const EnvSkipAudit = "MOAI_SKIP_PLAN_AUDIT"

// AuditResult holds the complete result of a single audit gate invocation.
//
// @MX:ANCHOR: [AUTO] AuditResult is the primary output type of the gate
// @MX:REASON: [AUTO] returned by GateConfig.Invoke, consumed by progress.md writer and daily report writer
type AuditResult struct {
	// Verdict is the gate decision: PASS, FAIL, FAIL_WARNED, BYPASSED, INCONCLUSIVE.
	Verdict Verdict

	// ReportPath is the path to the iteration-scoped plan-auditor report.
	// Empty when verdict is BYPASSED and no auditor was invoked.
	ReportPath string

	// AuditAt is the UTC timestamp of the gate invocation.
	AuditAt time.Time

	// AuditorVersion is the identifier of the plan-auditor agent used.
	AuditorVersion string

	// CacheHit is true when a valid 24h cached PASS was used instead of
	// invoking plan-auditor. REQ-WAG-003 / AC-WAG-09
	CacheHit bool

	// CachedAuditAt is the timestamp of the original cache-hit audit.
	// Populated only when CacheHit is true.
	CachedAuditAt time.Time

	// BypassUser is the user identifier when verdict is BYPASSED.
	BypassUser string

	// BypassReason is the sanitized rationale when verdict is BYPASSED.
	BypassReason string

	// InconclusiveAcknowledgedBy is the user identifier when the user chose
	// "proceed-with-acknowledgement" for an INCONCLUSIVE result. REQ-WAG-007.
	InconclusiveAcknowledgedBy string

	// GraceWindowActive indicates whether the grace window was active during this invocation.
	GraceWindowActive bool

	// GraceWindowRemainingDays is the number of days remaining in the grace window.
	// Only meaningful when GraceWindowActive is true.
	GraceWindowRemainingDays int

	// SpecID is the SPEC identifier that was audited.
	SpecID string
}

// PlanAuditor is the interface for invoking the plan-auditor subagent.
//
// Production implementation: invokes Agent(subagent_type="plan-auditor").
// Test implementation: returns deterministic verdicts without actual agent calls.
//
// @MX:ANCHOR: [AUTO] PlanAuditor is the primary abstraction for plan-auditor invocation
// @MX:REASON: [AUTO] fan_in >= 3 (GateConfig.Invoke, cache validation, integration tests)
type PlanAuditor interface {
	// Audit invokes plan-auditor against the SPEC directory and returns the verdict.
	//
	// specDir is the path to the SPEC directory (e.g., ".moai/specs/SPEC-XXX/").
	// The returned Verdict is one of PASS, FAIL, INCONCLUSIVE.
	// BYPASSED is not returned by the auditor — it is set by the gate when --skip-audit is used.
	//
	// REQ-WAG-001: invoked exactly once per /moai run (or 0 times on cache hit).
	Audit(ctx context.Context, specDir string) (Verdict, string, error)
}

// GateConfig holds the configuration for a single audit gate invocation.
//
// Callers construct a GateConfig and call Invoke to execute the full gate logic.
type GateConfig struct {
	// SpecID is the SPEC identifier being audited (e.g., "SPEC-AUTH-001").
	SpecID string

	// SpecDir is the path to the SPEC directory on disk.
	SpecDir string

	// ProjectDir is the root of the project (used for path safety validation).
	ProjectDir string

	// Auditor is the PlanAuditor implementation to use.
	Auditor PlanAuditor

	// Cache provides 24h cache lookup and write operations.
	Cache AuditCache

	// Reporter writes the daily audit report.
	Reporter AuditReporter

	// Clock provides the current time (injectable for tests).
	Clock Clock

	// UserName is the user's name from .moai/config/sections/user.yaml.
	UserName string

	// SkipAudit is true when --skip-audit flag or MOAI_SKIP_PLAN_AUDIT=1 is set.
	SkipAudit bool

	// BypassReason is the rationale for bypassing (collected by orchestrator via AskUserQuestion).
	// When non-interactive, this is "non-interactive". REQ-WAG-006.
	BypassReason string

	// T0 is the grace window start time. Read from .moai/state/audit-gate-merge-at.txt.
	// Zero value means grace window is not active.
	T0 time.Time
}

// isGraceWindowActive returns true if the current time is within 7 days of T0.
func (c *GateConfig) isGraceWindowActive(now time.Time) bool {
	if c.T0.IsZero() {
		// Attempt to read T0 from environment variable (test injection).
		if t0Str := os.Getenv(EnvGateT0); t0Str != "" {
			t0, err := time.Parse(time.RFC3339, t0Str)
			if err == nil {
				c.T0 = t0
			}
		}
	}
	if c.T0.IsZero() {
		return false
	}
	return now.Before(c.T0.Add(GracePeriod))
}

// graceWindowRemainingDays returns the number of full days remaining in the grace window.
func (c *GateConfig) graceWindowRemainingDays(now time.Time) int {
	if !c.isGraceWindowActive(now) {
		return 0
	}
	remaining := c.T0.Add(GracePeriod).Sub(now)
	days := int(remaining.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// Invoke executes the full audit gate logic for the configured SPEC.
//
// The function follows the 5-step protocol defined in run.md Phase 0.5:
// Step 1: compute plan artifact hash
// Step 2: check 24h cache
// Step 3: invoke plan-auditor (if cache miss)
// Step 4: route verdict
// Step 5: persist to progress.md and daily report
//
// Note: AskUserQuestion interactions for FAIL/INCONCLUSIVE decisions are
// orchestrator-level concerns. Invoke returns the raw verdict; the orchestrator
// presents options to the user and calls back into Invoke with the resolved action.
//
// REQ-WAG-001..007
func (c *GateConfig) Invoke(ctx context.Context) (*AuditResult, error) {
	now := c.Clock.Now().UTC()
	result := &AuditResult{
		SpecID:  c.SpecID,
		AuditAt: now,
	}

	// [BYPASS PATH] REQ-WAG-006
	if c.SkipAudit {
		result.Verdict = VerdictBypassed
		result.BypassUser = c.UserName
		result.BypassReason = c.BypassReason
		if c.BypassReason == "" {
			result.BypassReason = "non-interactive"
		}
		if err := c.Reporter.AppendRun(c.SpecID, result); err != nil {
			// Report write failure is non-fatal for the bypass path.
			fmt.Fprintf(os.Stderr, "[plan-audit] warning: failed to write bypass report: %v\n", err)
		}
		return result, nil
	}

	// Step 1: compute plan artifact hash
	hash, err := c.Cache.ComputeHash(c.SpecDir)
	if err != nil {
		// Hash computation failure → INCONCLUSIVE (cannot safely proceed without hash)
		result.Verdict = VerdictInconclusive
		result.AuditorVersion = "plan-auditor/hash-error"
		_ = c.Reporter.AppendRun(c.SpecID, result)
		return result, fmt.Errorf("plan artifact hash: %w", err)
	}

	// Step 2: check 24h cache
	if cached, cacheHit := c.Cache.Lookup(c.SpecID, hash, now); cacheHit {
		result.Verdict = VerdictPass
		result.CacheHit = true
		result.CachedAuditAt = cached.AuditAt
		result.AuditorVersion = cached.AuditorVersion
		result.ReportPath = cached.ReportPath
		fmt.Printf("[plan-audit] cache hit (verdict=PASS, age=%s)\n",
			now.Sub(cached.AuditAt).Round(time.Minute))
		_ = c.Reporter.AppendRun(c.SpecID, result)
		return result, nil
	}

	// Step 3: invoke plan-auditor
	// Log the invocation pattern required by AC-WAG-01.
	fmt.Printf("[plan-audit] invoking plan-auditor for %s (gate=mandatory)\n", c.SpecID)

	verdict, reportPath, auditErr := c.Auditor.Audit(ctx, c.SpecDir)
	result.ReportPath = reportPath
	result.AuditorVersion = "plan-auditor/v1"

	if auditErr != nil {
		// REQ-WAG-007: auditor error → INCONCLUSIVE
		result.Verdict = VerdictInconclusive
		fmt.Printf("[plan-audit] verdict=INCONCLUSIVE, falling back to manual prompt\n")
		_ = c.Reporter.AppendRun(c.SpecID, result)
		return result, nil
	}

	// Step 4: route verdict
	switch verdict {
	case VerdictPass:
		result.Verdict = VerdictPass
		// Store in cache for future 24h reuse.
		c.Cache.Store(c.SpecID, hash, result)
		fmt.Printf("[plan-audit] verdict=PASS, persisted to progress.md, proceeding to Phase 1\n")

	case VerdictFail:
		graceActive := c.isGraceWindowActive(now)
		result.GraceWindowActive = graceActive
		result.GraceWindowRemainingDays = c.graceWindowRemainingDays(now)

		if graceActive {
			result.Verdict = VerdictFailWarned
			fmt.Printf("[plan-audit] verdict=FAIL [grace-window], D-%d until auto-block\n",
				result.GraceWindowRemainingDays)
			fmt.Printf("[grace-window] D-%d (auto-block activates at T0+7)\n",
				result.GraceWindowRemainingDays)
		} else {
			result.Verdict = VerdictFail
			fmt.Printf("[plan-audit] verdict=FAIL, blocking Run phase, report=%s\n", reportPath)
		}

	case VerdictInconclusive:
		result.Verdict = VerdictInconclusive
		fmt.Printf("[plan-audit] verdict=INCONCLUSIVE, falling back to manual prompt\n")

	default:
		// Malformed verdict from auditor → INCONCLUSIVE. REQ-WAG-007.
		result.Verdict = VerdictInconclusive
		fmt.Printf("[plan-audit] verdict=INCONCLUSIVE, falling back to manual prompt\n")
	}

	// Step 5: persist to daily report.
	_ = c.Reporter.AppendRun(c.SpecID, result)
	return result, nil
}

// TeamModeInvoke wraps Invoke with team-mode logging.
//
// In team mode, the gate must execute before TeamCreate. This function
// adds the required log line and delegates to Invoke.
//
// REQ-WAG-005 / AC-WAG-05
func (c *GateConfig) TeamModeInvoke(ctx context.Context) (*AuditResult, error) {
	fmt.Printf("[plan-audit] team mode detected, gate applies before TeamCreate\n")
	return c.Invoke(ctx)
}
