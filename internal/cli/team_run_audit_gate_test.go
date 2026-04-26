//go:build integration

// Package cli provides integration tests for audit gate team mode behavior.
//
// AC-WAG-05: gate applies before TeamCreate; FAIL blocks teammate spawn.
// Run with: go test -tags=integration -race ./internal/cli/... -run TestTeamRun
package cli

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/runtime"
)

// mockTeamCreator records whether TeamCreate was attempted.
type mockTeamCreator struct {
	createCalled bool
}

func (m *mockTeamCreator) Create(_ context.Context, _ string) error {
	m.createCalled = true
	return nil
}

// orchestrateTeamRun simulates the team mode run phase orchestration.
// It enforces that TeamCreate is only called when audit gate passes.
// Returns true if TeamCreate was attempted.
func orchestrateTeamRun(ctx context.Context, gate *runtime.GateConfig, teamCreator *mockTeamCreator) (bool, error) {
	// Phase 0.5: Plan Audit Gate must pass before TeamCreate.
	result, err := gate.TeamModeInvoke(ctx)
	if err != nil {
		return false, err
	}

	// Only proceed to TeamCreate on PASS, FAIL_WARNED, or BYPASSED.
	// FAIL and INCONCLUSIVE block TeamCreate.
	switch result.Verdict {
	case runtime.VerdictFail, runtime.VerdictInconclusive:
		// Block: do not call TeamCreate.
		return false, nil
	default:
		// Proceed: call TeamCreate.
		teamCreator.createCalled = true
		return true, nil
	}
}

// TestTeamRunBlockedBeforeTeammateSpawn verifies AC-WAG-05:
// FAIL verdict in team mode blocks TeamCreate and any teammate spawn.
//
// Given: SPEC-TEAM-001 with FAIL verdict, grace window expired
// When: /moai run SPEC-TEAM-001 --team is called
// Then: gate verdict=FAIL, TeamCreate is NOT called, no teammate spawned
//
// AC: AC-WAG-05
// REQ: REQ-WAG-005
func TestTeamRunBlockedBeforeTeammateSpawn(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	specDir := fixtureDir(t, "SPEC-DUMMY-FAIL-001")

	// Expired grace window.
	pastT0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	gate := &runtime.GateConfig{
		SpecID:     "SPEC-TEAM-001",
		SpecDir:    specDir,
		ProjectDir: projectDir,
		Auditor:    &deterministicAuditor{verdict: runtime.VerdictFail},
		Cache:      runtime.NewInMemoryCache(),
		Reporter:   runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
		Clock:      runtime.FakeClock{FixedTime: time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)},
		UserName:   "test-user",
		T0:         pastT0,
	}

	teamCreator := &mockTeamCreator{}
	teamCreated, err := orchestrateTeamRun(context.Background(), gate, teamCreator)
	if err != nil {
		t.Fatalf("orchestrateTeamRun(): %v", err)
	}

	if teamCreated {
		t.Error("TeamCreate was called despite FAIL verdict — gate did not block (AC-WAG-05)")
	}
	if teamCreator.createCalled {
		t.Error("teamCreator.createCalled = true, want false (AC-WAG-05)")
	}
}

// TestTeamRunProceedsOnPass verifies AC-WAG-05 positive case:
// PASS verdict allows TeamCreate to proceed.
//
// AC: AC-WAG-05
// REQ: REQ-WAG-005
func TestTeamRunProceedsOnPass(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	specDir := fixtureDir(t, "SPEC-DUMMY-PASS-001")

	gate := &runtime.GateConfig{
		SpecID:     "SPEC-TEAM-PASS-001",
		SpecDir:    specDir,
		ProjectDir: projectDir,
		Auditor:    &deterministicAuditor{verdict: runtime.VerdictPass},
		Cache:      runtime.NewInMemoryCache(),
		Reporter:   runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
		Clock:      runtime.FakeClock{FixedTime: time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)},
		UserName:   "test-user",
	}

	teamCreator := &mockTeamCreator{}
	teamCreated, err := orchestrateTeamRun(context.Background(), gate, teamCreator)
	if err != nil {
		t.Fatalf("orchestrateTeamRun(): %v", err)
	}

	if !teamCreated {
		t.Error("TeamCreate was not called on PASS verdict — gate blocked incorrectly (AC-WAG-05)")
	}
}
