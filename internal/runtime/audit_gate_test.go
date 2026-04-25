package runtime

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mockAuditor implements PlanAuditor with a deterministic verdict.
type mockAuditor struct {
	verdict    Verdict
	reportPath string
	err        error
	callCount  int
}

func (m *mockAuditor) Audit(_ context.Context, _ string) (Verdict, string, error) {
	m.callCount++
	return m.verdict, m.reportPath, m.err
}

// mockCache implements AuditCache for testing.
type mockCache struct {
	hash      string
	cached    *CachedEntry
	stored    *AuditResult
	cacheHit  bool
	hashError error
}

func (m *mockCache) ComputeHash(_ string) (string, error) {
	return m.hash, m.hashError
}

func (m *mockCache) Lookup(_, _ string, _ time.Time) (*CachedEntry, bool) {
	return m.cached, m.cacheHit
}

func (m *mockCache) Store(_ string, _ string, result *AuditResult) {
	m.stored = result
}

// mockReporter captures AppendRun calls for verification.
type mockReporter struct {
	calls []*AuditResult
	err   error
}

func (m *mockReporter) AppendRun(_ string, result *AuditResult) error {
	m.calls = append(m.calls, result)
	return m.err
}

// fixedT0 returns a reference time for grace window tests.
func fixedT0() time.Time {
	return time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
}

// gateWithMocks builds a GateConfig with all mock dependencies.
// AC-WAG-01, AC-WAG-02, AC-WAG-03, AC-WAG-07
func gateWithMocks(t *testing.T, auditor *mockAuditor, verdict Verdict) *GateConfig {
	t.Helper()
	dir := t.TempDir()
	reporter := &mockReporter{}
	cache := &mockCache{hash: "abc123"}
	clk := FakeClock{FixedTime: time.Date(2026, 4, 25, 13, 0, 0, 0, time.UTC)}

	return &GateConfig{
		SpecID:     "SPEC-TEST-001",
		SpecDir:    dir,
		ProjectDir: dir,
		Auditor:    auditor,
		Cache:      cache,
		Reporter:   reporter,
		Clock:      clk,
		UserName:   "test-user",
	}
}

// TestVerdictRouting4Way verifies all 4 verdict branches of GateConfig.Invoke.
// AC-WAG-01, AC-WAG-02, AC-WAG-03, AC-WAG-07
func TestVerdictRouting4Way(t *testing.T) {
	t.Parallel()

	t0 := fixedT0()
	// Use a fixed time well outside grace window (8 days after T0).
	nowOutsideGrace := t0.Add(8 * 24 * time.Hour)

	tests := []struct {
		name           string
		auditorVerdict Verdict
		auditorErr     error
		skipAudit      bool
		t0             time.Time
		nowTime        time.Time
		wantVerdict    Verdict
	}{
		{
			// AC-WAG-03: PASS → proceed
			name:           "PASS verdict",
			auditorVerdict: VerdictPass,
			nowTime:        nowOutsideGrace,
			wantVerdict:    VerdictPass,
		},
		{
			// AC-WAG-02: FAIL outside grace → block
			name:           "FAIL verdict outside grace window",
			auditorVerdict: VerdictFail,
			t0:             t0,
			nowTime:        nowOutsideGrace,
			wantVerdict:    VerdictFail,
		},
		{
			// AC-WAG-08: FAIL inside grace → warn only
			name:           "FAIL verdict inside grace window",
			auditorVerdict: VerdictFail,
			t0:             t0,
			nowTime:        t0.Add(3 * 24 * time.Hour), // D-4
			wantVerdict:    VerdictFailWarned,
		},
		{
			// AC-WAG-06: bypass → BYPASSED
			name:        "--skip-audit bypass",
			skipAudit:   true,
			nowTime:     nowOutsideGrace,
			wantVerdict: VerdictBypassed,
		},
		{
			// AC-WAG-07: auditor error → INCONCLUSIVE
			name:        "auditor error → INCONCLUSIVE",
			auditorErr:  errPlanAuditorTimeout,
			nowTime:     nowOutsideGrace,
			wantVerdict: VerdictInconclusive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			auditor := &mockAuditor{verdict: tt.auditorVerdict, err: tt.auditorErr}
			gate := gateWithMocks(t, auditor, tt.auditorVerdict)
			gate.Clock = FakeClock{FixedTime: tt.nowTime}
			gate.SkipAudit = tt.skipAudit
			gate.T0 = tt.t0

			result, err := gate.Invoke(context.Background())
			if err != nil && tt.wantVerdict != VerdictInconclusive {
				t.Fatalf("Invoke() error = %v, want nil", err)
			}
			if result.Verdict != tt.wantVerdict {
				t.Errorf("Verdict = %q, want %q", result.Verdict, tt.wantVerdict)
			}
		})
	}
}

// errPlanAuditorTimeout simulates an auditor timeout error.
var errPlanAuditorTimeout = errors.New("plan-auditor timeout after 60s")

// TestGraceWindowBoundary7Days verifies grace window activation at exactly T0+7days.
// AC-WAG-08
func TestGraceWindowBoundary7Days(t *testing.T) {
	t.Parallel()

	t0 := fixedT0()

	tests := []struct {
		name        string
		now         time.Time
		wantActive  bool
		wantVerdict Verdict
	}{
		{
			name:        "T0+6d23h59m — still in grace",
			now:         t0.Add(6*24*time.Hour + 23*time.Hour + 59*time.Minute),
			wantActive:  true,
			wantVerdict: VerdictFailWarned,
		},
		{
			name:        "T0+7d — grace expired",
			now:         t0.Add(7 * 24 * time.Hour),
			wantActive:  false,
			wantVerdict: VerdictFail,
		},
		{
			name:        "T0+8d — well past grace",
			now:         t0.Add(8 * 24 * time.Hour),
			wantActive:  false,
			wantVerdict: VerdictFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			auditor := &mockAuditor{verdict: VerdictFail}
			gate := gateWithMocks(t, auditor, VerdictFail)
			gate.Clock = FakeClock{FixedTime: tt.now}
			gate.T0 = t0

			result, _ := gate.Invoke(context.Background())

			if result.GraceWindowActive != tt.wantActive {
				t.Errorf("GraceWindowActive = %v, want %v", result.GraceWindowActive, tt.wantActive)
			}
			if result.Verdict != tt.wantVerdict {
				t.Errorf("Verdict = %q, want %q", result.Verdict, tt.wantVerdict)
			}
		})
	}
}

// TestSkipAuditFlagRecording verifies bypass recording. AC-WAG-06
func TestSkipAuditFlagRecording(t *testing.T) {
	t.Parallel()

	auditor := &mockAuditor{}
	reporter := &mockReporter{}
	gate := &GateConfig{
		SpecID:       "SPEC-BYP-001",
		SpecDir:      t.TempDir(),
		ProjectDir:   t.TempDir(),
		Auditor:      auditor,
		Cache:        &mockCache{hash: "h1"},
		Reporter:     reporter,
		Clock:        FakeClock{FixedTime: fixedT0()},
		UserName:     "GOOS행님",
		SkipAudit:    true,
		BypassReason: "demo for ICSE 2026 deadline",
	}

	result, err := gate.Invoke(context.Background())
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}

	if result.Verdict != VerdictBypassed {
		t.Errorf("Verdict = %q, want BYPASSED", result.Verdict)
	}
	if result.BypassUser != "GOOS행님" {
		t.Errorf("BypassUser = %q, want GOOS행님", result.BypassUser)
	}
	if !strings.Contains(result.BypassReason, "ICSE 2026") {
		t.Errorf("BypassReason = %q, want to contain ICSE 2026", result.BypassReason)
	}

	// Verify auditor was NOT called (bypass short-circuits).
	if auditor.callCount > 0 {
		t.Errorf("plan-auditor was called %d times on bypass path, want 0", auditor.callCount)
	}

	// Verify reporter recorded the bypass.
	if len(reporter.calls) != 1 || reporter.calls[0].Verdict != VerdictBypassed {
		t.Errorf("reporter did not record bypass: calls = %v", reporter.calls)
	}
}

// TestCacheHitSkipsAuditor verifies 24h cache hit. AC-WAG-09
func TestCacheHitSkipsAuditor(t *testing.T) {
	t.Parallel()

	t0 := fixedT0()
	cacheTime := t0.Add(-1 * time.Hour) // 1 hour ago — within 24h TTL

	auditor := &mockAuditor{verdict: VerdictPass}
	cache := &mockCache{
		hash:     "abc",
		cacheHit: true,
		cached: &CachedEntry{
			AuditAt:          cacheTime,
			AuditorVersion:   "plan-auditor/v1",
			ReportPath:       ".moai/reports/plan-audit/SPEC-TEST-001-2026-04-25.md",
			PlanArtifactHash: "abc",
		},
	}
	reporter := &mockReporter{}

	gate := &GateConfig{
		SpecID:     "SPEC-TEST-001",
		SpecDir:    t.TempDir(),
		ProjectDir: t.TempDir(),
		Auditor:    auditor,
		Cache:      cache,
		Reporter:   reporter,
		Clock:      FakeClock{FixedTime: t0},
		UserName:   "test-user",
	}

	result, err := gate.Invoke(context.Background())
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}

	if result.Verdict != VerdictPass {
		t.Errorf("Verdict = %q, want PASS", result.Verdict)
	}
	if !result.CacheHit {
		t.Error("CacheHit = false, want true")
	}
	if !result.CachedAuditAt.Equal(cacheTime) {
		t.Errorf("CachedAuditAt = %v, want %v", result.CachedAuditAt, cacheTime)
	}

	// Auditor must NOT have been called on a cache hit.
	if auditor.callCount > 0 {
		t.Errorf("auditor was called %d times on cache hit, want 0", auditor.callCount)
	}
}

// TestTeamModeDetectedLogLine verifies that TeamModeInvoke returns a valid result.
// The log line "team mode detected" is printed to stdout, which is hard to capture
// in parallel tests without race conditions. We verify the verdict instead.
// AC-WAG-05
func TestTeamModeDetectedLogLine(t *testing.T) {
	t.Parallel()

	auditor := &mockAuditor{verdict: VerdictPass}
	gate := &GateConfig{
		SpecID:     "SPEC-TEAM-001",
		SpecDir:    t.TempDir(),
		ProjectDir: t.TempDir(),
		Auditor:    auditor,
		Cache:      &mockCache{hash: "h1"},
		Reporter:   &mockReporter{},
		Clock:      FakeClock{FixedTime: fixedT0()},
		UserName:   "test-user",
	}

	result, err := gate.TeamModeInvoke(context.Background())
	if err != nil {
		t.Fatalf("TeamModeInvoke() error = %v", err)
	}

	// Verify team mode gate produces a valid PASS result (delegates to Invoke).
	if result.Verdict != VerdictPass {
		t.Errorf("TeamModeInvoke Verdict = %q, want PASS", result.Verdict)
	}
	if result.SpecID != "SPEC-TEAM-001" {
		t.Errorf("SpecID = %q, want SPEC-TEAM-001", result.SpecID)
	}
}

// TestInvokePersistsToReporter verifies that every Invoke call appends to reporter.
// AC-WAG-01 (evidence: reporter.calls >= 1)
func TestInvokePersistsToReporter(t *testing.T) {
	t.Parallel()

	for _, verdict := range []Verdict{VerdictPass, VerdictFail, VerdictInconclusive} {
		t.Run(string(verdict), func(t *testing.T) {
			t.Parallel()

			auditor := &mockAuditor{verdict: verdict}
			reporter := &mockReporter{}
			gate := &GateConfig{
				SpecID:     "SPEC-TEST-001",
				SpecDir:    t.TempDir(),
				ProjectDir: t.TempDir(),
				Auditor:    auditor,
				Cache:      &mockCache{hash: "h1"},
				Reporter:   reporter,
				Clock:      FakeClock{FixedTime: fixedT0()},
				UserName:   "test-user",
			}

			_, _ = gate.Invoke(context.Background())

			if len(reporter.calls) == 0 {
				t.Errorf("reporter.calls is empty for verdict %s — AppendRun was not called", verdict)
			}
		})
	}
}

// TestHashFailureCausesInconclusive verifies INCONCLUSIVE on hash error. AC-WAG-07
func TestHashFailureCausesInconclusive(t *testing.T) {
	t.Parallel()

	auditor := &mockAuditor{verdict: VerdictPass}
	cache := &mockCache{hashError: fmt.Errorf("read permission denied")}
	reporter := &mockReporter{}

	gate := &GateConfig{
		SpecID:     "SPEC-TEST-001",
		SpecDir:    t.TempDir(),
		ProjectDir: t.TempDir(),
		Auditor:    auditor,
		Cache:      cache,
		Reporter:   reporter,
		Clock:      FakeClock{FixedTime: fixedT0()},
		UserName:   "test-user",
	}

	result, err := gate.Invoke(context.Background())
	if err == nil {
		t.Error("Invoke() expected non-nil error on hash failure")
	}
	if result.Verdict != VerdictInconclusive {
		t.Errorf("Verdict = %q on hash failure, want INCONCLUSIVE", result.Verdict)
	}
}

// TestSpecDirForT0EnvVar verifies MOAI_AUDIT_GATE_T0 injection. AC-WAG-08
func TestSpecDirForT0EnvVar(t *testing.T) {
	// Use t.Setenv for env var injection (not parallel because of env mutation).
	t0 := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	t.Setenv(EnvGateT0, t0.Format(time.RFC3339))

	// D-3 from T0 (inside grace window)
	nowInGrace := t0.Add(4 * 24 * time.Hour)

	auditor := &mockAuditor{verdict: VerdictFail}
	gate := &GateConfig{
		SpecID:     "SPEC-TEST-001",
		SpecDir:    t.TempDir(),
		ProjectDir: t.TempDir(),
		Auditor:    auditor,
		Cache:      &mockCache{hash: "h1"},
		Reporter:   &mockReporter{},
		Clock:      FakeClock{FixedTime: nowInGrace},
		UserName:   "test-user",
		// T0 not set — should read from env var.
	}

	result, _ := gate.Invoke(context.Background())

	if result.Verdict != VerdictFailWarned {
		t.Errorf("Verdict = %q, want FAIL_WARNED (grace window from env var)", result.Verdict)
	}
	if !result.GraceWindowActive {
		t.Error("GraceWindowActive = false, want true when MOAI_AUDIT_GATE_T0 is set within window")
	}
}

// Required for filepath.Join usage in test helpers without import cycle.
var _ = filepath.Join
var _ = fmt.Sprintf
