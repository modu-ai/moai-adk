package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupAgencyFixture creates a minimal .agency/ directory in dir.
// Returns the agencyDir path.
func setupAgencyFixture(t *testing.T, dir string) string {
	t.Helper()

	agencyDir := filepath.Join(dir, ".agency")
	contextDir := filepath.Join(agencyDir, "context")
	learningsDir := filepath.Join(agencyDir, "learnings")

	for _, d := range []string{contextDir, learningsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", d, err)
		}
	}

	files := map[string]string{
		filepath.Join(contextDir, "brand-voice.md"):     "# Brand Voice\ntone: professional\n",
		filepath.Join(contextDir, "visual-identity.md"): "# Visual Identity\ncolor: blue\n",
		filepath.Join(contextDir, "target-audience.md"): "# Target Audience\ndemo: developers\n",
		filepath.Join(agencyDir, "config.yaml"): `agency:
  gan_loop:
    max_iterations: 5
    pass_threshold: 0.75
  evolution:
    require_approval: true
  pipeline:
    phases: [planner, builder]
`,
		filepath.Join(learningsDir, "LEARN-001.md"): "# Learning 001\nObservation: test\n",
		filepath.Join(agencyDir, "fork-manifest.yaml"): "version: 1\nforks: []\n",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile %s: %v", path, err)
		}
	}

	return agencyDir
}

// TestMigrateAgency_HappyPath verifies AC-MIGRATE-001: successful full migration.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-001,004,005,007,REQ-DIR-001
func TestMigrateAgency_HappyPath(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	// Create a stub .moai/config/sections/ dir so design.yaml target parent exists.
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "research"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir, // use dir as home to keep tx files in temp
		dryRun:      false,
		force:       false,
	}

	result, err := m.Run()
	if err != nil {
		t.Fatalf("Run() returned unexpected error: %v", err)
	}

	// brand files copied
	for _, name := range []string{"brand-voice.md", "visual-identity.md", "target-audience.md"} {
		dst := filepath.Join(dir, ".moai", "project", "brand", name)
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", dst)
		}
	}

	// design.yaml generated
	designYAML := filepath.Join(dir, ".moai", "config", "sections", "design.yaml")
	data, err := os.ReadFile(designYAML)
	if err != nil {
		t.Fatalf("design.yaml not created: %v", err)
	}
	if !strings.Contains(string(data), "gan_loop") {
		t.Errorf("design.yaml missing 'gan_loop' key, got:\n%s", data)
	}

	// learnings moved
	obs := filepath.Join(dir, ".moai", "research", "observations", "LEARN-001.md")
	if _, err := os.Stat(obs); os.IsNotExist(err) {
		t.Error("expected LEARN-001.md in observations/")
	}

	// archive created
	archive := filepath.Join(dir, ".agency.archived")
	if _, err := os.Stat(archive); os.IsNotExist(err) {
		t.Error("expected .agency.archived/ to exist")
	}

	// summary reports file count
	if result.filesTransferred == 0 {
		t.Error("expected filesTransferred > 0")
	}

	if result.archivePath == "" {
		t.Error("expected archivePath to be set")
	}
}

// TestMigrateAgency_DryRun verifies AC-MIGRATE-002: dry-run does not modify anything.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-008
func TestMigrateAgency_DryRun(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
		dryRun:      true,
	}

	_, err := m.Run()
	if err != nil {
		t.Fatalf("Run() dry-run returned error: %v", err)
	}

	// .agency.archived/ must NOT be created
	if _, err := os.Stat(filepath.Join(dir, ".agency.archived")); !os.IsNotExist(err) {
		t.Error("dry-run must not create .agency.archived/")
	}

	// .moai/project/brand/ must NOT be created
	if _, err := os.Stat(filepath.Join(dir, ".moai", "project", "brand")); !os.IsNotExist(err) {
		t.Error("dry-run must not create .moai/project/brand/")
	}

	// original .agency/ untouched
	if _, err := os.Stat(filepath.Join(dir, ".agency")); os.IsNotExist(err) {
		t.Error("dry-run must not remove .agency/")
	}
}

// TestMigrateAgency_NoAgencyDir verifies AC-MIGRATE-003: missing source.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-002
func TestMigrateAgency_NoAgencyDir(t *testing.T) {
	dir := t.TempDir()

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
	}

	_, err := m.Run()
	if err == nil {
		t.Fatal("expected error when .agency/ missing, got nil")
	}

	me, ok := err.(*MigrateError)
	if !ok {
		t.Fatalf("expected *MigrateError, got %T: %v", err, err)
	}
	if me.Code != ErrMigrateNoSource {
		t.Errorf("expected code %s, got %s", ErrMigrateNoSource, me.Code)
	}
}

// TestMigrateAgency_AlreadyMigrated verifies AC-MIGRATE-004: target exists without --force.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-003
func TestMigrateAgency_AlreadyMigrated(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	// Pre-create target to simulate already migrated
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "project", "brand"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".moai", "project", "brand", "brand-voice.md"), []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
	}

	_, err := m.Run()
	if err == nil {
		t.Fatal("expected error when target exists without --force")
	}

	me, ok := err.(*MigrateError)
	if !ok {
		t.Fatalf("expected *MigrateError, got %T", err)
	}
	if me.Code != ErrMigrateTargetExists {
		t.Errorf("expected code %s, got %s", ErrMigrateTargetExists, me.Code)
	}
}

// TestMigrateAgency_ForceOverwrite verifies AC-MIGRATE-005: --force overwrites existing target.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-009
func TestMigrateAgency_ForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	// Pre-create target
	brandDir := filepath.Join(dir, ".moai", "project", "brand")
	if err := os.MkdirAll(brandDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(brandDir, "brand-voice.md"), []byte("old-content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create required parent dirs
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "research"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
		force:       true,
	}

	_, err := m.Run()
	if err != nil {
		t.Fatalf("Run() with --force returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(brandDir, "brand-voice.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "old-content" {
		t.Error("expected brand-voice.md to be overwritten with agency content")
	}
}

// TestMigrateAgency_ArchiveExists verifies AC-MIGRATE-006: .agency.archived/ already exists without --force.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-010
func TestMigrateAgency_ArchiveExists(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	// Pre-create archive
	if err := os.MkdirAll(filepath.Join(dir, ".agency.archived"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
	}

	_, err := m.Run()
	if err == nil {
		t.Fatal("expected error when .agency.archived/ already exists")
	}

	me, ok := err.(*MigrateError)
	if !ok {
		t.Fatalf("expected *MigrateError, got %T", err)
	}
	if me.Code != ErrMigrateArchiveExists {
		t.Errorf("expected code %s, got %s", ErrMigrateArchiveExists, me.Code)
	}
}

// TestMigrateAgency_Atomicity verifies AC-MIGRATE-007: rollback on failure.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-006
func TestMigrateAgency_Atomicity(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	// Create a target dir to simulate phase-3 write failure via a read-only observations dir
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "research"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
		// inject a failure hook at phase 3
		failAtPhase: 3,
	}

	_, err := m.Run()
	if err == nil {
		t.Fatal("expected error when phase 3 fails")
	}

	// .agency/ must still exist (rolled back)
	if _, err := os.Stat(filepath.Join(dir, ".agency")); os.IsNotExist(err) {
		t.Error("rollback: .agency/ must be restored")
	}

	// partial .moai/project/brand/ must be cleaned up
	if _, err := os.Stat(filepath.Join(dir, ".moai", "project", "brand")); !os.IsNotExist(err) {
		t.Error("rollback: .moai/project/brand/ must be removed")
	}
}

// TestMigrateAgency_MergeConflict verifies AC-MIGRATE-009: tech-preferences vs tech.md conflict.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-DIR-002
func TestMigrateAgency_MergeConflict(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	// Add tech-preferences.md with conflicting content
	if err := os.WriteFile(
		filepath.Join(dir, ".agency", "context", "tech-preferences.md"),
		[]byte("# Tech Preferences\nFramework: Next.js\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	// Create existing tech.md in .moai/project/ with different content
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "project"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, ".moai", "project", "tech.md"),
		[]byte("# Tech Stack\nFramework: Remix\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
	}

	_, err := m.Run()
	if err == nil {
		t.Fatal("expected error on merge conflict")
	}

	me, ok := err.(*MigrateError)
	if !ok {
		t.Fatalf("expected *MigrateError, got %T", err)
	}
	if me.Code != ErrMigrateMergeConflict {
		t.Errorf("expected code %s, got %s", ErrMigrateMergeConflict, me.Code)
	}

	// Both files must be preserved
	if _, err := os.Stat(filepath.Join(dir, ".moai", "project", "tech.md")); os.IsNotExist(err) {
		t.Error("tech.md must be preserved on conflict")
	}
	if _, err := os.Stat(filepath.Join(dir, ".agency", "context", "tech-preferences.md")); os.IsNotExist(err) {
		t.Error("tech-preferences.md must be preserved on conflict")
	}
}

// TestMigrateAgency_Archive verifies AC-MIGRATE-001 archive creation detail.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-004 step 1
func TestMigrateAgency_Archive(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "research"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
	}

	_, err := m.Run()
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	archive := filepath.Join(dir, ".agency.archived")
	if _, err := os.Stat(archive); os.IsNotExist(err) {
		t.Fatal(".agency.archived/ must exist after successful migration")
	}

	// fork-manifest.yaml must be in archive
	fmPath := filepath.Join(archive, "fork-manifest.yaml")
	if _, err := os.Stat(fmPath); os.IsNotExist(err) {
		t.Error("fork-manifest.yaml must be archived")
	}
}

// TestMigrateAgency_Resume verifies AC-MIGRATE-010: --resume resumes from checkpoint.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-013
func TestMigrateAgency_Resume(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "research"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a fake checkpoint: already completed phases 1 and 2
	txID := "test-tx-001"
	cp := migrationCheckpoint{
		TxID:          txID,
		ProjectRoot:   dir,
		CompletedPhases: []int{1, 2},
		RemainingFiles:  []string{
			filepath.Join(dir, ".agency", "learnings", "LEARN-001.md"),
		},
	}

	homeDir := dir
	cpPath := filepath.Join(homeDir, ".moai", ".migrate-tx-"+txID+".json")
	if err := os.MkdirAll(filepath.Join(homeDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeCheckpoint(cpPath, &cp); err != nil {
		t.Fatalf("writeCheckpoint: %v", err)
	}

	// Pre-create the brand dir as if phases 1-2 already ran
	brandDir := filepath.Join(dir, ".moai", "project", "brand")
	if err := os.MkdirAll(brandDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"brand-voice.md", "visual-identity.md", "target-audience.md"} {
		content, _ := os.ReadFile(filepath.Join(dir, ".agency", "context", f))
		_ = os.WriteFile(filepath.Join(brandDir, f), content, 0o644)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     homeDir,
		resumeTxID:  txID,
		force:       true, // allow existing brand dir
	}

	_, err := m.Run()
	if err != nil {
		t.Fatalf("Resume Run() returned error: %v", err)
	}

	// After resume, LEARN-001 should be migrated
	obs := filepath.Join(dir, ".moai", "research", "observations", "LEARN-001.md")
	if _, err := os.Stat(obs); os.IsNotExist(err) {
		t.Error("resume: LEARN-001.md must be migrated")
	}
}

// TestMigrateAgency_EmptyAgencyDir verifies EC-001: empty .agency/ dir.
func TestMigrateAgency_EmptyAgencyDir(t *testing.T) {
	dir := t.TempDir()

	// Empty .agency/
	if err := os.MkdirAll(filepath.Join(dir, ".agency"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
	}

	result, err := m.Run()
	if err != nil {
		t.Fatalf("expected success for empty .agency/, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// archive should still be created
	if _, err := os.Stat(filepath.Join(dir, ".agency.archived")); os.IsNotExist(err) {
		t.Error(".agency.archived/ must be created even for empty source")
	}
}

// TestMigrateAgency_DiskFull은 AC-MIGRATE-011을 검증한다:
// 가용 디스크 공간이 .agency/ 크기의 2배 미만일 때 MIGRATE_DISK_FULL 오류를 반환해야 한다.
//
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-011
func TestMigrateAgency_DiskFull(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}

	// checkDiskSpaceFn을 디스크 가득 참 상태로 모킹한다.
	// 실제 디스크를 채우지 않고 함수 변수 주입 패턴을 사용한다.
	original := checkDiskSpaceFn
	t.Cleanup(func() { checkDiskSpaceFn = original })
	checkDiskSpaceFn = func(_ string) error {
		return &MigrateError{
			Code:    ErrMigrateDiskFull,
			Message: "테스트: 디스크 공간 부족 시뮬레이션",
		}
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
	}

	_, err := m.Run()
	if err == nil {
		t.Fatal("디스크 가득 참 상태에서 Run()이 오류 없이 반환됨 (오류 반환 필요)")
	}

	me, ok := err.(*MigrateError)
	if !ok {
		t.Fatalf("*MigrateError를 기대했으나 %T: %v", err, err)
	}
	if me.Code != ErrMigrateDiskFull {
		t.Errorf("오류 코드 %s를 기대했으나 %s", ErrMigrateDiskFull, me.Code)
	}

	// 디스크 가득 참 오류 시 파일시스템에 변경이 없어야 한다.
	if _, statErr := os.Stat(filepath.Join(dir, ".agency.archived")); !os.IsNotExist(statErr) {
		t.Error("디스크 가득 참 오류 시 .agency.archived/가 생성되면 안 됨")
	}
	if _, statErr := os.Stat(filepath.Join(dir, ".moai", "project", "brand")); !os.IsNotExist(statErr) {
		t.Error("디스크 가득 참 오류 시 .moai/project/brand/가 생성되면 안 됨")
	}
}

// TestMigrateAgency_DiskFull_AlsoBlocksDryRun은 --dry-run 모드에서도 디스크 검사가 실행됨을 검증한다.
// REQ-MIGRATE-011은 dry-run 예외를 명시하지 않으므로 사전 검사는 모든 경로에서 수행된다.
//
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-011
func TestMigrateAgency_DiskFull_AlsoBlocksDryRun(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	// checkDiskSpaceFn을 항상 실패로 모킹한다.
	original := checkDiskSpaceFn
	t.Cleanup(func() { checkDiskSpaceFn = original })
	checkDiskSpaceFn = func(_ string) error {
		return &MigrateError{
			Code:    ErrMigrateDiskFull,
			Message: "테스트: 디스크 공간 부족 시뮬레이션",
		}
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
		dryRun:      true,
	}

	// REQ-MIGRATE-011은 dry-run 예외를 명시하지 않으므로 MIGRATE_DISK_FULL 오류가 반환되어야 한다.
	_, err := m.Run()
	if err == nil {
		t.Fatal("디스크 가득 참 상태에서 dry-run도 오류를 반환해야 함")
	}
	me, ok := err.(*MigrateError)
	if !ok {
		t.Fatalf("*MigrateError를 기대했으나 %T: %v", err, err)
	}
	if me.Code != ErrMigrateDiskFull {
		t.Errorf("오류 코드 %s를 기대했으나 %s", ErrMigrateDiskFull, me.Code)
	}
}

// TestMigrateAgency_PhaseFailure4 verifies rollback when phase 4 (config conversion) fails.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-006
func TestMigrateAgency_PhaseFailure4(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "research"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
		failAtPhase: 4,
	}

	_, err := m.Run()
	if err == nil {
		t.Fatal("expected error when phase 4 fails")
	}

	// .agency/ must still exist (rolled back)
	if _, err := os.Stat(filepath.Join(dir, ".agency")); os.IsNotExist(err) {
		t.Error("rollback: .agency/ must be preserved")
	}
}

// TestMigrateAgency_ConstitutionRelocation verifies step 6: constitution.md relocation.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-004 step 6
func TestMigrateAgency_ConstitutionRelocation(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	// Place a user constitution.md
	constSrc := filepath.Join(dir, ".claude", "rules", "agency")
	if err := os.MkdirAll(constSrc, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(constSrc, "constitution.md"), []byte("# Agency Constitution\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "research"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
	}

	result, err := m.Run()
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	// Constitution must be relocated
	constDst := filepath.Join(dir, ".claude", "rules", "moai", "design", "constitution.md")
	if _, err := os.Stat(constDst); os.IsNotExist(err) {
		t.Error("constitution.md must be relocated to .claude/rules/moai/design/")
	}

	if result.filesTransferred == 0 {
		t.Error("filesTransferred must include constitution.md")
	}
}

// TestMigrateAgency_MigrateError_Error verifies the MigrateError.Error() format.
func TestMigrateAgency_MigrateError_Error(t *testing.T) {
	me := &MigrateError{Code: ErrMigrateNoSource, Message: "not found"}
	got := me.Error()
	if !strings.Contains(got, ErrMigrateNoSource) {
		t.Errorf("Error() must contain code %s, got: %s", ErrMigrateNoSource, got)
	}
	if !strings.Contains(got, "not found") {
		t.Errorf("Error() must contain message, got: %s", got)
	}
}

// TestMigrateAgency_Checkpoint_WriteRead verifies checkpoint roundtrip.
func TestMigrateAgency_Checkpoint_WriteRead(t *testing.T) {
	dir := t.TempDir()
	cpPath := filepath.Join(dir, ".moai", ".migrate-tx-roundtrip.json")

	cp := &migrationCheckpoint{
		TxID:            "roundtrip-001",
		ProjectRoot:     dir,
		CompletedPhases: []int{1, 2, 3},
		RemainingFiles:  []string{"/tmp/file.md"},
	}

	if err := writeCheckpoint(cpPath, cp); err != nil {
		t.Fatalf("writeCheckpoint: %v", err)
	}

	got, err := readCheckpoint(cpPath)
	if err != nil {
		t.Fatalf("readCheckpoint: %v", err)
	}

	if got.TxID != cp.TxID {
		t.Errorf("TxID mismatch: got %s, want %s", got.TxID, cp.TxID)
	}
	if len(got.CompletedPhases) != len(cp.CompletedPhases) {
		t.Errorf("CompletedPhases mismatch: got %v, want %v", got.CompletedPhases, cp.CompletedPhases)
	}
}

// TestMigrateAgency_Resume_InvalidCheckpoint verifies corrupt checkpoint handling.
// @MX:SPEC: SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-013
func TestMigrateAgency_Resume_InvalidCheckpoint(t *testing.T) {
	dir := t.TempDir()

	txID := "bad-tx"
	cpDir := filepath.Join(dir, ".moai")
	if err := os.MkdirAll(cpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write invalid JSON
	cpPath := filepath.Join(cpDir, ".migrate-tx-"+txID+".json")
	if err := os.WriteFile(cpPath, []byte("{invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
		resumeTxID:  txID,
	}

	_, err := m.Run()
	if err == nil {
		t.Fatal("expected error for corrupt checkpoint")
	}

	me, ok := err.(*MigrateError)
	if !ok {
		t.Fatalf("expected *MigrateError, got %T: %v", err, err)
	}
	if me.Code != ErrMigrateCheckpointCorrupt {
		t.Errorf("expected code %s, got %s", ErrMigrateCheckpointCorrupt, me.Code)
	}
}

// TestMigrateAgency_DefaultWriters verifies nil stdout/stderr defaults to os.Stderr/os.Stdout.
func TestMigrateAgency_DefaultWriters(t *testing.T) {
	m := &migrateAgencyRunner{}
	if m.getStderr() == nil {
		t.Error("getStderr() must not return nil")
	}
	if m.getStdout() == nil {
		t.Error("getStdout() must not return nil")
	}
}

// TestMigrateAgency_CopyFileReadError verifies error propagation when source is unreadable.
func TestMigrateAgency_CopyFileReadError(t *testing.T) {
	dir := t.TempDir()
	m := &migrateAgencyRunner{projectRoot: dir, homeDir: dir}
	tx := &transactionLog{}

	err := m.copyFile(filepath.Join(dir, "nonexistent.md"), filepath.Join(dir, "dst.md"), tx)
	if err == nil {
		t.Error("expected error copying nonexistent file")
	}
}

// TestMigrateAgency_WriteCheckpoint_MkdirError verifies error when checkpoint dir cannot be created.
func TestMigrateAgency_WriteCheckpoint_MkdirError(t *testing.T) {
	// Use a path that cannot be created (file as parent)
	dir := t.TempDir()
	blockingFile := filepath.Join(dir, "blocking")
	if err := os.WriteFile(blockingFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	cpPath := filepath.Join(blockingFile, "subdir", ".migrate-tx-test.json")
	cp := &migrationCheckpoint{TxID: "test"}
	err := writeCheckpoint(cpPath, cp)
	if err == nil {
		t.Error("expected error when checkpoint dir cannot be created")
	}
}

// TestMigrateAgency_NoMergeConflict verifies no error when tech-preferences.md frameworks match.
func TestMigrateAgency_NoMergeConflict(t *testing.T) {
	dir := t.TempDir()
	setupAgencyFixture(t, dir)

	// Same framework in both files — no conflict
	if err := os.WriteFile(
		filepath.Join(dir, ".agency", "context", "tech-preferences.md"),
		[]byte("Framework: Next.js\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, ".moai", "project"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, ".moai", "project", "tech.md"),
		[]byte("Framework: Next.js\n"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, ".moai", "config", "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "research"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &migrateAgencyRunner{
		projectRoot: dir,
		homeDir:     dir,
	}

	_, err := m.Run()
	if err != nil {
		t.Fatalf("expected success when frameworks match, got: %v", err)
	}
}
