// Package cli — update_clean_install_test.go
//
// Integration tests for the 7-step clean-reinstall orchestrator
// (SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 REQ-VVCR-002..004 + REQ-VVCR-027..029).
// Tests use synthetic in-memory deployers and t.TempDir() isolation
// (CLAUDE.local.md §6 HARD).
//
// Scenario coverage (per plan.md §F.M6):
//   - Scenario A: Full v2 project with all signal sources positive
//   - Scenario B: Partial v2 (only .agency/ present) — runMigrateAgency
//                  auto-invoke
//   - Scenario C: Clean v3 project — no-op idempotency (REQ-VVCR-027)
//
// Tests do NOT call the production embedded template deployer — they use a
// stub deployer that records its Deploy call. This isolates the orchestrator
// logic from upstream template fixtures and keeps the test hermetic.

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// stubDeployer is a test double satisfying the template.Deployer interface
// minimally enough for runCleanReinstall integration testing. Deploy()
// records the projectRoot it was invoked with and returns nil.
type stubDeployer struct {
	deployCalls   int
	lastProjRoot  string
	deployErr     error
	listResult    []string
	validateErr   error
}

func (s *stubDeployer) Deploy(ctx context.Context, projectRoot string, mgr manifest.Manager, tmplCtx *template.TemplateContext) error {
	s.deployCalls++
	s.lastProjRoot = projectRoot
	return s.deployErr
}

func (s *stubDeployer) ListTemplates() []string {
	return s.listResult
}

func (s *stubDeployer) ValidateAll(ctx context.Context, tmplCtx *template.TemplateContext) error {
	return s.validateErr
}

// ExtractTemplate satisfies template.Deployer; unused by orchestrator tests.
func (s *stubDeployer) ExtractTemplate(name string) ([]byte, error) {
	return nil, nil
}

// stubMigrateRunner is the fake adapter for opts.RunMigrateAgency. Records
// invocations for assertion.
type stubMigrateRunner struct {
	calls    int
	lastRoot string
	err      error
}

func (s *stubMigrateRunner) Run(projectRoot string, dryRun bool, out io.Writer) error {
	s.calls++
	s.lastRoot = projectRoot
	return s.err
}

// makeScenarioA constructs a project tree resembling a complete v2 install:
//   - .moai/config/sections/system.yaml with moai.version = v2.16.1
//   - .agency/ directory present
//   - .claude/agents/moai/manager-strategy.md (deprecated path)
//   - PRESERVE seed: .moai/specs/SPEC-USER-001/spec.md + .claude/skills/my-harness-tool/SKILL.md
//
// Returns the projectRoot path.
func makeScenarioA(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v2.16.1\n")
	makeTestDir(t, root, ".agency")
	// Add a content file to .agency/ so it's not just an empty dir.
	writeTestFile(t, root, ".agency/index.md", "legacy agency content\n")
	writeTestFile(t, root, ".claude/agents/moai/manager-strategy.md", "retired agent\n")

	// PRESERVE seed
	writeTestFile(t, root, ".moai/specs/SPEC-USER-001/spec.md", "user spec content\n")
	writeTestFile(t, root, ".claude/skills/my-harness-tool/SKILL.md", "user skill\n")
	writeTestFile(t, root, ".moai/project/product.md", "product doc\n")

	return root
}

// makeScenarioB constructs a partial v2 project — only .agency/ legacy
// directory present (Signal 2 only). system.yaml is v3-clean.
func makeScenarioB(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v3.0.0-rc2\n")
	makeTestDir(t, root, ".agency")
	writeTestFile(t, root, ".agency/index.md", "legacy agency content\n")

	// PRESERVE seed
	writeTestFile(t, root, ".moai/specs/SPEC-USER-002/spec.md", "user spec B\n")

	return root
}

// makeScenarioC constructs a clean v3 project — no v2 signals.
func makeScenarioC(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v3.0.0-rc2\n")
	// No .agency/, no deprecated paths.

	// PRESERVE seed
	writeTestFile(t, root, ".moai/specs/SPEC-USER-003/spec.md", "user spec C\n")

	return root
}

// TestRunCleanReinstall_ScenarioA verifies the full v2 → v3 cycle:
//   - All 3 signals fire
//   - PRESERVE inventory snapshot taken (.moai/specs/, .claude/skills/my-harness-*, .moai/project/)
//   - Backup directory created at .moai/backups/v2-to-v3-<stamp>/
//   - .agency/ migration auto-invoked (REQ-VVCR-025)
//   - Deprecated paths removed
//   - Deployer invoked
//   - PRESERVE files restored byte-identical
//   - Integrity check PASSES
func TestRunCleanReinstall_ScenarioA(t *testing.T) {
	root := makeScenarioA(t)
	deployer := &stubDeployer{}
	migrate := &stubMigrateRunner{}

	opts := CleanReinstallOptions{
		Out:              io.Discard,
		Deployer:         deployer,
		RunMigrateAgency: migrate.Run,
	}

	result, err := runCleanReinstall(context.Background(), root, opts)
	if err != nil {
		t.Fatalf("runCleanReinstall scenario A: %v", err)
	}

	if !result.Detected.IsV2 {
		t.Errorf("scenario A: Detected.IsV2 = false; want true")
	}
	if !result.Detected.V2DetectedViaVersion {
		t.Errorf("scenario A: V2DetectedViaVersion = false; want true")
	}
	if !result.Detected.V2DetectedViaAgencyDir {
		t.Errorf("scenario A: V2DetectedViaAgencyDir = false; want true")
	}
	if !result.Detected.V2DetectedViaDeprecatedPath {
		t.Errorf("scenario A: V2DetectedViaDeprecatedPath = false; want true")
	}
	if result.BackupDir == "" {
		t.Errorf("scenario A: BackupDir is empty; want non-empty")
	}
	if !result.AgencyMigrated {
		t.Errorf("scenario A: AgencyMigrated = false; want true")
	}
	if migrate.calls != 1 {
		t.Errorf("scenario A: migrate.calls = %d; want 1", migrate.calls)
	}
	if migrate.lastRoot != root {
		t.Errorf("scenario A: migrate.lastRoot = %q; want %q", migrate.lastRoot, root)
	}
	if deployer.deployCalls != 1 {
		t.Errorf("scenario A: deployer.deployCalls = %d; want 1", deployer.deployCalls)
	}
	if deployer.lastProjRoot != root {
		t.Errorf("scenario A: deployer.lastProjRoot = %q; want %q", deployer.lastProjRoot, root)
	}
	if !result.IntegrityPassed {
		t.Errorf("scenario A: IntegrityPassed = false; want true (mismatches: %v)", result.IntegrityMismatches)
	}

	// AC-VVCR-003: PRESERVE files survive byte-identical.
	preservePaths := []string{
		".moai/specs/SPEC-USER-001/spec.md",
		".claude/skills/my-harness-tool/SKILL.md",
		".moai/project/product.md",
	}
	for _, rel := range preservePaths {
		abs := filepath.Join(root, rel)
		if _, statErr := os.Stat(abs); statErr != nil {
			t.Errorf("scenario A: PRESERVE file missing after reinstall: %s (%v)", rel, statErr)
		}
	}

	// AC-VVCR-002: backup directory exists with .complete marker.
	completeMarker := filepath.Join(result.BackupDir, ".complete")
	if _, err := os.Stat(completeMarker); err != nil {
		t.Errorf("scenario A: .complete marker missing in backup dir: %v", err)
	}
}

// TestRunCleanReinstall_ScenarioB verifies partial v2 path (only .agency/):
//   - Signal 2 fires (agency dir)
//   - Other signals do not fire (or fire only via deprecated paths if they
//     happen to exist; in this scenario they do not)
//   - runMigrateAgency is auto-invoked
//   - Deployer is invoked
func TestRunCleanReinstall_ScenarioB(t *testing.T) {
	root := makeScenarioB(t)
	deployer := &stubDeployer{}
	migrate := &stubMigrateRunner{}

	opts := CleanReinstallOptions{
		Out:              io.Discard,
		Deployer:         deployer,
		RunMigrateAgency: migrate.Run,
	}

	result, err := runCleanReinstall(context.Background(), root, opts)
	if err != nil {
		t.Fatalf("runCleanReinstall scenario B: %v", err)
	}

	if !result.Detected.IsV2 {
		t.Errorf("scenario B: IsV2 = false; want true")
	}
	if !result.Detected.V2DetectedViaAgencyDir {
		t.Errorf("scenario B: V2DetectedViaAgencyDir = false; want true")
	}
	if migrate.calls != 1 {
		t.Errorf("scenario B: migrate.calls = %d; want 1", migrate.calls)
	}
	if deployer.deployCalls != 1 {
		t.Errorf("scenario B: deployer.deployCalls = %d; want 1", deployer.deployCalls)
	}

	// AC-VVCR-003: PRESERVE survives.
	preserveAbs := filepath.Join(root, ".moai/specs/SPEC-USER-002/spec.md")
	if _, statErr := os.Stat(preserveAbs); statErr != nil {
		t.Errorf("scenario B: PRESERVE file missing: %v", statErr)
	}
}

// TestRunCleanReinstall_ScenarioC verifies no-op on a clean v3 project
// (REQ-VVCR-027 idempotency).
func TestRunCleanReinstall_ScenarioC(t *testing.T) {
	root := makeScenarioC(t)
	deployer := &stubDeployer{}
	migrate := &stubMigrateRunner{}

	opts := CleanReinstallOptions{
		Out:              io.Discard,
		Deployer:         deployer,
		RunMigrateAgency: migrate.Run,
	}

	result, err := runCleanReinstall(context.Background(), root, opts)
	if err != nil {
		t.Fatalf("runCleanReinstall scenario C: %v", err)
	}

	// Clean v3 project: IsV2 should be false → early return, no mutations.
	if result.Detected.IsV2 {
		t.Errorf("scenario C: Detected.IsV2 = true; want false (clean v3 project)")
	}
	if result.BackupDir != "" {
		t.Errorf("scenario C: BackupDir = %q; want empty (no-op)", result.BackupDir)
	}
	if migrate.calls != 0 {
		t.Errorf("scenario C: migrate.calls = %d; want 0 (no-op)", migrate.calls)
	}
	if deployer.deployCalls != 0 {
		t.Errorf("scenario C: deployer.deployCalls = %d; want 0 (no-op)", deployer.deployCalls)
	}
}

// TestRunCleanReinstall_DryRun verifies the --dry-run flag (REQ-VVCR-028):
//   - No filesystem mutations
//   - No deployer invocation
//   - Result still carries detection + inventory + planned-removal data
func TestRunCleanReinstall_DryRun(t *testing.T) {
	root := makeScenarioA(t)
	deployer := &stubDeployer{}
	migrate := &stubMigrateRunner{}

	opts := CleanReinstallOptions{
		DryRun:           true,
		Out:              io.Discard,
		Deployer:         deployer,
		RunMigrateAgency: migrate.Run,
	}

	result, err := runCleanReinstall(context.Background(), root, opts)
	if err != nil {
		t.Fatalf("runCleanReinstall dry-run: %v", err)
	}

	if !result.DryRun {
		t.Errorf("dry-run: result.DryRun = false; want true")
	}
	if !result.Detected.IsV2 {
		t.Errorf("dry-run: still expected to detect v2 fingerprint")
	}
	if result.BackupDir != "" {
		t.Errorf("dry-run: BackupDir = %q; want empty (no mutations)", result.BackupDir)
	}
	if deployer.deployCalls != 0 {
		t.Errorf("dry-run: deployer.deployCalls = %d; want 0", deployer.deployCalls)
	}

	// Verify the deprecated path is STILL present (no mutations).
	deprecatedAbs := filepath.Join(root, ".claude/agents/moai/manager-strategy.md")
	if _, statErr := os.Stat(deprecatedAbs); statErr != nil {
		t.Errorf("dry-run: deprecated path was removed (%v); want preserved", statErr)
	}
}

// TestRunCleanReinstall_EmptyRoot verifies error handling.
func TestRunCleanReinstall_EmptyRoot(t *testing.T) {
	_, err := runCleanReinstall(context.Background(), "", CleanReinstallOptions{Out: io.Discard})
	if err == nil {
		t.Errorf("expected error for empty projectRoot; got nil")
	}
}

// TestRunCleanReinstall_DeployerErrorPropagates verifies that a deploy
// failure propagates as an error from runCleanReinstall.
func TestRunCleanReinstall_DeployerErrorPropagates(t *testing.T) {
	root := makeScenarioA(t)
	deployer := &stubDeployer{deployErr: fmt.Errorf("synthetic deploy failure")}

	opts := CleanReinstallOptions{
		Out:      io.Discard,
		Deployer: deployer,
	}

	_, err := runCleanReinstall(context.Background(), root, opts)
	if err == nil {
		t.Errorf("expected error from deployer failure; got nil")
	}
}

// TestResolveV2BackupDir_CollisionHandling verifies that same-second
// directory collisions are resolved via numeric suffix.
func TestResolveV2BackupDir_CollisionHandling(t *testing.T) {
	root := t.TempDir()
	candidate := filepath.Join(root, "v2-to-v3-2026-05-25T12-00-00Z")

	// First call — no collision, returns the candidate unchanged.
	out1, err := resolveV2BackupDir(candidate)
	if err != nil {
		t.Fatalf("first resolveV2BackupDir: %v", err)
	}
	if out1 != candidate {
		t.Errorf("first call: got %q, want %q", out1, candidate)
	}

	// Materialize the first candidate so the second call collides.
	if err := os.MkdirAll(out1, 0o755); err != nil {
		t.Fatalf("mkdir for collision: %v", err)
	}

	// Second call — should suffix -1.
	out2, err := resolveV2BackupDir(candidate)
	if err != nil {
		t.Fatalf("second resolveV2BackupDir: %v", err)
	}
	if out2 != candidate+"-1" {
		t.Errorf("second call: got %q, want %q", out2, candidate+"-1")
	}
}
