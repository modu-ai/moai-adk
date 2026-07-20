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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// stubDeployer is a test double satisfying the template.Deployer interface
// minimally enough for runCleanReinstall integration testing. Deploy()
// records the projectRoot it was invoked with and returns nil.
type stubDeployer struct {
	deployCalls  int
	lastProjRoot string
	lastTmplCtx  *template.TemplateContext
	deployErr    error
	listResult   []string
	validateErr  error
}

func (s *stubDeployer) Deploy(ctx context.Context, projectRoot string, mgr manifest.Manager, tmplCtx *template.TemplateContext) error {
	s.deployCalls++
	s.lastProjRoot = projectRoot
	s.lastTmplCtx = tmplCtx
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
//   - PRESERVE seed: .moai/specs/SPEC-USER-001/spec.md + .claude/skills/harness-tool/SKILL.md
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
	writeTestFile(t, root, ".claude/skills/harness-tool/SKILL.md", "user skill\n")
	writeTestFile(t, root, ".moai/project/product.md", "product doc\n")

	return root
}

// makeScenarioB constructs a partial v2 project — a v2-era version with a
// lingering .agency/ legacy directory (Signals 1 + 2). The version is v2.*
// (not v3.*) so REQ-CRR-001's v3-version negative-override does NOT fire —
// this is the genuine partial-v2 migration case per acceptance.md §D.6 Edge-2
// ("a v2-project case (system.yaml carries v2.* version)").
//
// Prior to SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002 this fixture used v3.0.0-rc2,
// which — under the old pure-disjunction aggregation — drove IsV2=true via
// Signal 2 alone. REQ-CRR-001 corrected that: a v3.* version with .agency/
// residue is now IsV2=false (AC-CRR-007: clean-reinstall NOT activated for a
// v3 project). To keep this scenario testing the partial-v2 → clean-reinstall
// path, the version was changed to v2.* — the realistic v2-project setup.
func makeScenarioB(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v2.16.1\n")
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
//   - PRESERVE inventory snapshot taken (.moai/specs/, .claude/skills/harness-*, .moai/project/)
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
		".claude/skills/harness-tool/SKILL.md",
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

// TestRunCleanReinstall_ScenarioB verifies partial v2 path (v2.* version +
// .agency/ residue, no deprecated paths):
//   - Signals 1 (v2.* version) + 2 (agency dir) fire
//   - REQ-CRR-001 v3-version negative-override does NOT fire (version is v2.*)
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

// TestRunCleanReinstall_PopulatesPATHContext is the reproduction test for the
// v2→v3 clean-reinstall PATH regression: the Step 5 deploy used a bare
// template.NewTemplateContext() (no options), leaving SmartPATH/GoBinPath/HomeDir
// empty. settings.json.tmpl then rendered "PATH": "" and status_line.sh rendered
// the "/moai" fallback (posixPath("")+"/moai"), so downstream projects lost the
// moai binary on PATH after upgrading — breaking the statusline and every
// PATH-resolved MCP server (moai-lsp, npx-based servers).
//
// The deploy context MUST carry a populated SmartPATH and GoBinPath, matching
// the normal `moai update` deploy path (update.go "Deploy Templates" step).
func TestRunCleanReinstall_PopulatesPATHContext(t *testing.T) {
	root := makeScenarioA(t)
	deployer := &stubDeployer{}
	migrate := &stubMigrateRunner{}

	opts := CleanReinstallOptions{
		Out:              io.Discard,
		Deployer:         deployer,
		RunMigrateAgency: migrate.Run,
	}

	if _, err := runCleanReinstall(context.Background(), root, opts); err != nil {
		t.Fatalf("runCleanReinstall: %v", err)
	}

	if deployer.lastTmplCtx == nil {
		t.Fatal("deployer was not invoked with a template context")
	}
	// Regression guard: an empty SmartPATH renders settings.json env.PATH = ""
	// which strips the moai binary from PATH in all downstream sessions.
	if deployer.lastTmplCtx.SmartPATH == "" {
		t.Error("clean-reinstall deploy context SmartPATH is empty; settings.json would render \"PATH\": \"\"")
	}
	// Regression guard: an empty GoBinPath renders the status_line.sh fallback as
	// "/moai" instead of the real Go bin path.
	if deployer.lastTmplCtx.GoBinPath == "" {
		t.Error("clean-reinstall deploy context GoBinPath is empty; status_line.sh would render the \"/moai\" fallback")
	}
}

// TestRunUpdate_V3ProjectWithAgencyDir_MigratesIndependently is the M3
// reproduction test for AC-CRR-007 (SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002):
//
// GIVEN a v3 project (system.yaml moai.version = v3.0.0-rc2, so REQ-CRR-001's
// v3-version negative-override forces detectV2Fingerprint.IsV2=false) that
// carries a lingering `.agency/` legacy directory
// WHEN `moai update` is invoked via updateCmd.RunE
// THEN (per AC-CRR-007):
//
//	(a) runMigrateAgency fires INDEPENDENTLY of the v2 fingerprint verdict —
//	    proved by `.agency.archived/` existing (Phase 1 backup side-effect
//	    created only by the real migrateAgencyRunner.Run());
//	(c) full clean-reinstall is NOT activated — proved by absence of any
//	    `.moai/backups/v2-to-v3-*-{stamp}/` directory.
//
// PRE-M3 (RED): the migration lives only inside runCleanReinstall Step 3.5,
// gated on fp.V2DetectedViaAgencyDir. With IsV2=false the gate never opens,
// so `.agency.archived/` is never created → assertion (a) FAILS.
//
// POST-M3 (GREEN): runUpdate carries an independent pre-step that probes
// `.agency/` directly (gated by isMoAIProject) and fires
// runAgencyMigrationAdapter BEFORE detectV2Fingerprint, so `.agency.archived/`
// is created regardless of the v2 fingerprint verdict.
func TestRunUpdate_V3ProjectWithAgencyDir_MigratesIndependently(t *testing.T) {
	// Fixture: v3 project with lingering .agency/. The PRESERVE seed is
	// included so isMoAIProject sees a real moai project (system.yaml is the
	// positive marker; the SPEC file makes the tree look non-empty).
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v3.0.0-rc2\n")
	makeTestDir(t, root, ".agency")
	writeTestFile(t, root, ".agency/index.md", "legacy agency content\n")
	// PRESERVE seed (mirrors ScenarioA/B/C pattern).
	writeTestFile(t, root, ".moai/specs/SPEC-USER-007/spec.md", "user spec\n")

	// runUpdate reads cwd via os.Getwd() — chdir into the fixture root.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir fixture: %v", err)
	}

	// Mirror the TestRunUpdate_DefaultIsTemplateSync call pattern
	// (coverage_test.go:420): stub deps.UpdateChecker so the pre-line-281
	// setup does not attempt a network update probe.
	origDeps := deps
	defer func() { deps = origDeps }()
	deps = &Dependencies{UpdateChecker: &mockUpdateChecker{}}

	updateCmd.SetOut(io.Discard)
	updateCmd.SetErr(io.Discard)
	updateCmd.SetContext(context.Background())
	if err := updateCmd.Flags().Set("check", "false"); err != nil {
		t.Fatalf("set check flag: %v", err)
	}
	if err := updateCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatalf("set yes flag: %v", err)
	}

	// Invoke runUpdate via the cobra RunE handler. We do NOT hard-fail on
	// a non-nil err: the load-bearing signal is the filesystem side-effect
	// (.agency.archived/ existing), not end-to-end runUpdate success. The
	// downstream template sync may emit warnings on this minimal fixture
	// without affecting the AC-CRR-007 invariant.
	if runErr := updateCmd.RunE(updateCmd, []string{}); runErr != nil {
		t.Logf("runUpdate returned err (non-fatal for AC-CRR-007 check): %v", runErr)
	}

	// AC-CRR-007(a): migration fired independently of v2 fingerprint.
	// Prove it by the Phase 1 backup side-effect — `.agency.archived/` is
	// created only by the real migrateAgencyRunner.Run() (Phase 1 copyDir).
	archiveDir := filepath.Join(root, ".agency.archived")
	if _, err := os.Stat(archiveDir); err != nil {
		t.Errorf("AC-CRR-007(a): .agency.archived/ missing after runUpdate (%v); "+
			"independent .agency migration did not fire for this v3-with-agency fixture", err)
	}

	// AC-CRR-007(c): clean-reinstall NOT activated. Prove it by absence of
	// any v2-to-v3 backup directory (runCleanReinstall Step 2 creates these).
	backupMatches, _ := filepath.Glob(filepath.Join(root, ".moai", "backups", "v2-to-v3-*"))
	if len(backupMatches) != 0 {
		t.Errorf("AC-CRR-007(c): v2-to-v3 backup dirs exist at %v; "+
			"clean-reinstall should NOT activate for a v3 project", backupMatches)
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

// TestRunCleanReinstall_RemovalCountLogGating_AC_CRR_006 verifies REQ-CRR-006:
// the REMOVE-phase log reports the ACTUAL filesystem removal count, and the
// `Removed N deprecated paths` line is suppressed entirely when nothing was
// removed (the #1084 phantom log).
//
// Pre-fix, the line was emitted unconditionally with len(plannedList) — so a
// run that removed nothing still announced a removal, contradicting `git diff`.
func TestRunCleanReinstall_RemovalCountLogGating_AC_CRR_006(t *testing.T) {
	const removedLine = "Removed"
	const noneLine = "No deprecated paths found to remove"

	t.Run("zero deprecated paths → no Removed line, informational line instead", func(t *testing.T) {
		// A v2 project carrying NO deprecated residue at all: Signal 1 is
		// positive (moai.version v2.16.1) so IsV2=true and the orchestrator
		// reaches Step 4 — but the REMOVE phase legitimately removes nothing.
		//
		// This fixture is built inline rather than reusing makeScenarioB
		// because `.agency` is ITSELF a DeprecatedPaths entry (defs/dirs.go),
		// so any .agency/-bearing fixture has a non-zero removal count. This
		// is exactly the zero-removal case that produced the #1084 phantom log.
		root := t.TempDir()
		writeTestFile(t, root, ".moai/config/sections/system.yaml",
			"moai:\n    version: v2.16.1\n")
		writeTestFile(t, root, ".moai/specs/SPEC-USER-006/spec.md", "user spec\n")
		var buf bytes.Buffer

		result, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
			Out:              &buf,
			Deployer:         &stubDeployer{},
			RunMigrateAgency: (&stubMigrateRunner{}).Run,
		})
		if err != nil {
			t.Fatalf("runCleanReinstall: %v", err)
		}
		if !result.Detected.IsV2 {
			t.Fatalf("fixture precondition broken: IsV2=false, so Step 4 never ran")
		}
		if len(result.RemovedPaths) != 0 {
			t.Fatalf("fixture precondition broken: expected 0 deprecated paths on disk, got %v",
				result.RemovedPaths)
		}

		got := buf.String()
		// AC-CRR-006(b): no `Removed N deprecated paths` when nothing removed.
		if stringContains(got, removedLine) {
			t.Errorf("AC-CRR-006(b): phantom log emitted with zero actual removals; "+
				"found %q in output:\n%s", removedLine, got)
		}
		// AC-CRR-006(c): the informational line takes its place.
		if !stringContains(got, noneLine) {
			t.Errorf("AC-CRR-006(c): missing %q in output:\n%s", noneLine, got)
		}
	})

	t.Run("real removals → Removed line reports the actual count", func(t *testing.T) {
		// ScenarioA seeds exactly two DeprecatedPaths members on disk:
		// `.agency` (defs/dirs.go:119) and
		// `.claude/agents/moai/manager-strategy.md` (defs/dirs.go:157).
		root := makeScenarioA(t)
		var buf bytes.Buffer

		result, err := runCleanReinstall(context.Background(), root, CleanReinstallOptions{
			Out:              &buf,
			Deployer:         &stubDeployer{},
			RunMigrateAgency: (&stubMigrateRunner{}).Run,
		})
		if err != nil {
			t.Fatalf("runCleanReinstall: %v", err)
		}
		if len(result.RemovedPaths) != 2 {
			t.Fatalf("fixture precondition broken: expected exactly 2 deprecated paths, got %v",
				result.RemovedPaths)
		}

		got := buf.String()
		// AC-CRR-006(d): the count reported is the actual one.
		if !stringContains(got, "Removed 2 deprecated paths") {
			t.Errorf("AC-CRR-006(d): expected %q in output:\n%s",
				"Removed 2 deprecated paths", got)
		}
		if stringContains(got, noneLine) {
			t.Errorf("AC-CRR-006: informational no-op line emitted despite a real "+
				"removal; output:\n%s", got)
		}

		// AC-CRR-006(a): the count is a filesystem diff — the path is really gone.
		if _, statErr := os.Stat(filepath.Join(root,
			".claude", "agents", "moai", "manager-strategy.md")); statErr == nil {
			t.Error("AC-CRR-006(a): deprecated path still exists after REMOVE; " +
				"the reported count would not reflect the filesystem")
		}
	})
}

// ---------------------------------------------------------------------------
// M4 — Reproduction tests (CLAUDE.md §7 Rule 4 Reproduction-First / HARD-4)
//
// Both tests below are authored to compile AND fail on the PRE-FIX codebase
// (commit 2bc543636, before M1/M2/M3). They deliberately reference only
// surfaces that predate the fix — detectV2Fingerprint().IsV2 and the
// runUpdate cobra entry point — and NOT the post-fix helpers
// (isMoAIProject, V2Fingerprint.V3VersionConfirmed). That constraint is what
// makes the fail-pre-fix step mechanically demonstrable: the tests can be
// dropped onto a pre-fix worktree verbatim and observed RED.
// ---------------------------------------------------------------------------

// TestReproduction_FingerprintNonConvergence_Issue1084 reproduces the #1084
// infinite `moai update` loop (REQ-CRR-008; verifies AC-CRR-002(a)).
//
// Root Cause A (plan.md §A.1): the pre-fix heuristic OR'd its three signals
// with no v3-version negative-override, so a genuine v3 project carrying
// stale legacy residue (.agency/ + a deprecated path) classified as v2 on
// EVERY run. Clean-reinstall does not remove `.agency/` and does not rewrite
// the residue away, so the next run saw the same positive signals — the
// fingerprint never converged and the loop never terminated.
//
// The load-bearing assertion is convergence: on a v3 project the verdict MUST
// be IsV2=false, which is what routes runUpdate to file-level sync instead of
// re-entering clean-reinstall (AC-CRR-002(a)).
//
// Pre-fix: IsV2=true (Signal 2 + Signal 3 drive the disjunction) → FAIL.
// Post-fix: IsV2=false (REQ-CRR-001 v3-version negative-override) → PASS.
func TestReproduction_FingerprintNonConvergence_Issue1084(t *testing.T) {
	// Fixture per AC-CRR-002: a v3 project whose system.yaml carries a
	// populated v3.* version, PLUS the exact legacy residue that drove the
	// loop — a lingering .agency/ dir (Signal 2) and a DeprecatedPaths
	// member on disk (Signal 3).
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v3.0.0-rc2\n")
	makeTestDir(t, root, ".agency")
	writeTestFile(t, root, ".agency/index.md", "legacy agency residue\n")
	writeTestFile(t, root, ".claude/agents/moai/manager-strategy.md", "retired agent\n")
	// User-modified config per AC-CRR-002(d) — the asset the loop endangered.
	writeTestFile(t, root, ".moai/config/sections/language.yaml",
		"language:\n    conversation_language: ko\n")

	// Invoke the fingerprint twice on the SAME fixture with no mutation in
	// between (acceptance.md §D.5: "verify fingerprint convergence on the
	// SAME fixture ... no test-fixture drift"). A converged predicate is
	// stable and negative on both reads.
	for run := 1; run <= 2; run++ {
		fp, err := detectV2Fingerprint(root)
		if err != nil {
			t.Fatalf("run %d: detectV2Fingerprint: unexpected error: %v", run, err)
		}

		// Precondition sanity: the legacy residue really is present, so this
		// test is exercising the override — not passing because the fixture
		// silently lost its Signal 2/3 sources.
		if !fp.V2DetectedViaAgencyDir {
			t.Fatalf("run %d: fixture precondition broken: Signal 2 (.agency/) not detected", run)
		}
		if !fp.V2DetectedViaDeprecatedPath {
			t.Fatalf("run %d: fixture precondition broken: Signal 3 (DeprecatedPaths) not detected", run)
		}

		// AC-CRR-002(a): the v3 project must NOT route into clean-reinstall.
		// This is the assertion that FAILS on the pre-fix implementation.
		if fp.IsV2 {
			t.Errorf("run %d: #1084 regression: IsV2 = true on a v3 project "+
				"(moai.version v3.0.0-rc2) carrying .agency/ + deprecated-path residue; "+
				"want false. A positive verdict here re-enters clean-reinstall on every "+
				"run — the fingerprint never converges (infinite loop). signals: "+
				"version=%v agency=%v deprecated=%v details=%v",
				run, fp.V2DetectedViaVersion, fp.V2DetectedViaAgencyDir,
				fp.V2DetectedViaDeprecatedPath, fp.SignalDetails)
		}
	}

	// AC-CRR-002(d): the user's `ko` value survives — the fingerprint probe is
	// read-only and never touches user config.
	got, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "language.yaml"))
	if err != nil {
		t.Fatalf("read language.yaml: %v", err)
	}
	if want := "language:\n    conversation_language: ko\n"; string(got) != want {
		t.Errorf("AC-CRR-002(d): language.yaml drifted: got %q, want %q", string(got), want)
	}
}

// TestReproduction_NonProjectDirectoryPollution_Issue1086 reproduces the #1086
// arbitrary-directory pollution (REQ-CRR-009; verifies AC-CRR-005(a)(b)).
//
// Root Cause C (plan.md §A.1): Option α treats a missing system.yaml as a
// POSITIVE Signal 1, which is correct for pre-system.yaml v2 projects but also
// fires in any directory that has no `.moai/` at all. Pre-fix, `moai update`
// in e.g. /tmp/some-random-dir produced IsV2=true → runCleanReinstall → the
// full embedded template tree written into the cwd.
//
// The assertion is the filesystem side-effect, not the return value: a
// non-project cwd must be left untouched (AC-CRR-005(a)(b)). runUpdate is
// driven through the cobra RunE handler — the same call shape used by
// TestRunUpdate_V3ProjectWithAgencyDir_MigratesIndependently.
//
// Pre-fix: `.moai/` and `.claude/` materialize in the cwd → FAIL.
// Post-fix: the isMoAIProject gate refuses entry, cwd stays clean → PASS.
func TestReproduction_NonProjectDirectoryPollution_Issue1086(t *testing.T) {
	// Fixture per AC-CRR-005: a cwd with NO .moai/ whatsoever. The .agency/
	// dir stands in for the incidental legacy residue an arbitrary directory
	// may carry (acceptance.md §D.6 Edge-3) and makes Signal 2 positive too,
	// so IsV2 is unambiguously true pre-fix.
	root := t.TempDir()
	makeTestDir(t, root, ".agency")
	writeTestFile(t, root, "README.md", "just some random directory\n")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir fixture: %v", err)
	}

	// Stub the network update probe (mirrors coverage_test.go:420 pattern).
	origDeps := deps
	defer func() { deps = origDeps }()
	deps = &Dependencies{UpdateChecker: &mockUpdateChecker{}}

	updateCmd.SetOut(io.Discard)
	updateCmd.SetErr(io.Discard)
	updateCmd.SetContext(context.Background())
	if err := updateCmd.Flags().Set("check", "false"); err != nil {
		t.Fatalf("set check flag: %v", err)
	}
	if err := updateCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatalf("set yes flag: %v", err)
	}

	runErr := updateCmd.RunE(updateCmd, []string{})
	t.Logf("runUpdate returned: %v", runErr)

	// AC-CRR-005(d): the command exits non-zero. A nil error here is the
	// pre-fix behavior — the command "succeeded" precisely by polluting the cwd.
	if runErr == nil {
		t.Errorf("AC-CRR-005(d): runUpdate returned nil in a non-project cwd; " +
			"want a non-nil error so the command exits non-zero")
	} else {
		// AC-CRR-005(c): the error is structured — it names the missing marker
		// file and directs the user to `moai init`.
		msg := runErr.Error()
		for _, want := range []string{
			"not a moai project",
			".moai/config/sections/system.yaml",
			"moai init",
		} {
			if !stringContains(msg, want) {
				t.Errorf("AC-CRR-005(c): error message missing %q; got: %s", want, msg)
			}
		}
		// §D.7 Secured: the structured error must not leak the absolute cwd.
		if stringContains(msg, root) {
			t.Errorf("§D.7 Secured: error leaks the absolute cwd %q; got: %s", root, msg)
		}
	}

	// AC-CRR-005(a): no .moai/ or .claude/ created in the cwd.
	for _, dir := range []string{".moai", ".claude"} {
		if _, err := os.Stat(filepath.Join(root, dir)); err == nil {
			t.Errorf("AC-CRR-005(a): #1086 regression: %s/ was created in a "+
				"non-project cwd; `moai update` must refuse to install anything "+
				"when the .moai/config/sections/system.yaml marker is absent", dir)
		}
	}

	// AC-CRR-005(b): no template files written anywhere under the cwd. The
	// fixture seeds exactly two entries (.agency/ and README.md); anything
	// beyond that is template spill.
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("readdir fixture root: %v", err)
	}
	allowed := map[string]bool{".agency": true, "README.md": true}
	for _, e := range entries {
		if !allowed[e.Name()] {
			t.Errorf("AC-CRR-005(b): #1086 regression: unexpected entry %q "+
				"materialized in a non-project cwd; want only the seeded "+
				".agency/ and README.md", e.Name())
		}
	}
}

// ---------------------------------------------------------------------------
// M5 — Idempotency + cross-platform parity
// ---------------------------------------------------------------------------

// TestRunUpdate_ThreeRunIdempotency_V3Project verifies REQ-CRR-011 / AC-CRR-009:
// three consecutive `moai update` invocations on a v3 project with a
// user-modified language.yaml converge — the user value survives byte-identical
// and runs 2/3 neither create a v2-to-v3 backup dir nor emit a REMOVE-phase log.
//
// This is the multi-run generalization of the #1084 loop: pre-fix, each run
// re-classified the v3 project as v2 (Signal 2/3 residue) and re-entered
// clean-reinstall, so a backup dir and REMOVE log appeared on every run. The
// M1 v3-version negative-override converges the fingerprint; this test pins the
// convergence across three runs (AC-CRR-009(a)(b)(c)).
func TestRunUpdate_ThreeRunIdempotency_V3Project(t *testing.T) {
	// Fixture per AC-CRR-009: a v3 project (system.yaml v3.0.0-rc2) with a
	// user-modified language.yaml. The .agency/ + deprecated residue is the
	// #1084 loop trigger that must NOT re-fire clean-reinstall on a v3 project.
	root := t.TempDir()
	writeTestFile(t, root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v3.0.0-rc2\n")
	const userLang = "language:\n    conversation_language: ko\n"
	writeTestFile(t, root, ".moai/config/sections/language.yaml", userLang)
	// Legacy residue that drove the pre-fix loop (Signals 2 + 3).
	makeTestDir(t, root, ".agency")
	writeTestFile(t, root, ".agency/index.md", "legacy residue\n")
	writeTestFile(t, root, ".moai/specs/SPEC-USER-009/spec.md", "user spec\n")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir fixture: %v", err)
	}

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = &Dependencies{UpdateChecker: &mockUpdateChecker{}}

	langPath := filepath.Join(root, ".moai", "config", "sections", "language.yaml")

	// The idempotency invariant (AC-CRR-009(a)+(d)): run 1 MAY legitimately
	// normalize the minimal fixture via file-level sync (the SHA-256 hash-diff
	// classifies language.yaml as user-modified and preserves conversation_language:
	// ko while merging in the other template keys — parent REQ-VVCR-007). What
	// MUST hold is (i) the user's ko value survives EVERY run, and (ii) the file
	// converges — it is byte-stable from run 1 onward (runs 2/3 identical to run
	// 1, no further drift). firstRunContent captures the post-run-1 reference.
	var firstRunContent string

	for run := 1; run <= 3; run++ {
		var buf bytes.Buffer
		updateCmd.SetOut(&buf)
		updateCmd.SetErr(&buf)
		updateCmd.SetContext(context.Background())
		if err := updateCmd.Flags().Set("check", "false"); err != nil {
			t.Fatalf("run %d: set check flag: %v", run, err)
		}
		if err := updateCmd.Flags().Set("yes", "true"); err != nil {
			t.Fatalf("run %d: set yes flag: %v", run, err)
		}

		if runErr := updateCmd.RunE(updateCmd, []string{}); runErr != nil {
			// Template sync on this minimal fixture may emit warnings; the
			// load-bearing invariants are the filesystem assertions below.
			t.Logf("run %d: runUpdate returned (non-fatal): %v", run, runErr)
		}

		out := buf.String()

		got, readErr := os.ReadFile(langPath)
		if readErr != nil {
			t.Fatalf("run %d: read language.yaml: %v", run, readErr)
		}

		// AC-CRR-009(a): the user's conversation_language: ko value survives on
		// every run — it is never overwritten, removed, relocated, or dropped.
		if !stringContains(string(got), "conversation_language: ko") {
			t.Errorf("run %d: AC-CRR-009(a): conversation_language: ko lost from "+
				"language.yaml; got %q", run, string(got))
		}

		// AC-CRR-009(a)+(d): convergence — run 1 may normalize, but the file is
		// byte-stable from run 1 onward (runs 2/3 identical to run 1). Any
		// per-run drift here is a non-idempotent sync (the multi-run #1084 shape).
		if run == 1 {
			firstRunContent = string(got)
		} else if string(got) != firstRunContent {
			t.Errorf("run %d: AC-CRR-009(a): language.yaml is not idempotent — "+
				"drifted from the post-run-1 state.\nrun 1: %q\nrun %d: %q",
				run, firstRunContent, run, string(got))
		}

		// AC-CRR-009(b)(c): on runs 2 and 3 the v3 project must NOT re-enter
		// clean-reinstall — no v2-to-v3 backup dir, no REMOVE-phase log.
		if run >= 2 {
			backupMatches, _ := filepath.Glob(
				filepath.Join(root, ".moai", "backups", "v2-to-v3-*"))
			if len(backupMatches) != 0 {
				t.Errorf("run %d: AC-CRR-009(b): v2-to-v3 backup dir(s) present %v; "+
					"clean-reinstall must NOT re-activate on a v3 project", run, backupMatches)
			}
			if stringContains(out, "[clean-reinstall] Removed") ||
				stringContains(out, "[clean-reinstall] No deprecated paths found to remove") {
				t.Errorf("run %d: AC-CRR-009(c): a REMOVE-phase log was emitted; "+
					"the v3 project must route to file-level sync, not clean-reinstall.\n%s",
					run, out)
			}
		}
	}
}

// TestFingerprintPredicate_CrossPlatformParity verifies AC-CRR-010 (S2 / SHOULD-1):
// the fingerprint predicate and the positive-marker precondition resolve
// identically regardless of OS path-separator conventions.
//
// The three OS matrices (macOS/Linux/Windows) are exercised by the CI runner
// matrix; this in-process test pins the OS-agnostic property the CI matrix
// relies on — that the predicate builds its marker path via filepath.Join
// (separator-agnostic) and that the verdict is deterministic. A regression that
// reintroduced a hardcoded "/" separator would break the marker resolution on
// Windows while leaving this assertion green only on POSIX, so the test also
// asserts the marker resolves under the current GOOS.
func TestFingerprintPredicate_CrossPlatformParity(t *testing.T) {
	// AC-CRR-010(c): a genuine v3 project marker is detected on the current OS.
	v3Root := t.TempDir()
	writeTestFile(t, v3Root, ".moai/config/sections/system.yaml",
		"moai:\n    version: v3.0.0\n")
	if !isMoAIProject(v3Root) {
		t.Errorf("AC-CRR-010(c): isMoAIProject=false for a v3 project on %s; "+
			"the marker path must resolve via filepath.Join on every OS", runtime.GOOS)
	}

	// AC-CRR-010(b): a non-project cwd is rejected identically on every OS.
	nonProject := t.TempDir()
	makeTestDir(t, nonProject, ".agency")
	if isMoAIProject(nonProject) {
		t.Errorf("AC-CRR-010(b): isMoAIProject=true for a non-project cwd on %s; want false",
			runtime.GOOS)
	}

	// AC-CRR-010(a): the fingerprint verdict is deterministic and stable across
	// repeated evaluations on the same fixture (the property the OS matrices
	// each assert). A v3 project with residue converges to IsV2=false.
	v3WithResidue := t.TempDir()
	writeTestFile(t, v3WithResidue, ".moai/config/sections/system.yaml",
		"moai:\n    version: v3.0.0-rc2\n")
	makeTestDir(t, v3WithResidue, ".agency")
	for i := 0; i < 3; i++ {
		fp, err := detectV2Fingerprint(v3WithResidue)
		if err != nil {
			t.Fatalf("iteration %d: detectV2Fingerprint: %v", i, err)
		}
		if fp.IsV2 {
			t.Errorf("iteration %d: AC-CRR-010(a): IsV2=true on a v3 project; "+
				"verdict must be a stable false on %s", i, runtime.GOOS)
		}
	}
}
