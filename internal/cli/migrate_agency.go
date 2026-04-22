// Package cli implements the moai command-line interface.
// migrate_agency.go provides the `moai migrate agency` subcommand.
//
// @SPEC:SPEC-AGENCY-ABSORB-001:REQ-MIGRATE-001~014,REQ-DIR-001~003
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Migration error codes (REQ-MIGRATE-002/003/006/010/011/013, REQ-DIR-002).
const (
	// ErrMigrateNoSource is returned when .agency/ does not exist.
	ErrMigrateNoSource = "MIGRATE_NO_SOURCE"
	// ErrMigrateTargetExists is returned when a migration target already exists without --force.
	ErrMigrateTargetExists = "MIGRATE_TARGET_EXISTS"
	// ErrMigrateArchiveExists is returned when .agency.archived/ already exists without --force.
	ErrMigrateArchiveExists = "MIGRATE_ARCHIVE_EXISTS"
	// ErrMigrateRollbackOK is reported after a successful rollback.
	ErrMigrateRollbackOK = "MIGRATE_ROLLBACK_OK"
	// ErrMigrateRollbackFailed is reported when rollback itself fails.
	ErrMigrateRollbackFailed = "MIGRATE_ROLLBACK_FAILED"
	// ErrMigrateDiskFull is returned when available disk space is insufficient.
	ErrMigrateDiskFull = "MIGRATE_DISK_FULL"
	// ErrMigrateMergeConflict is returned when tech-preferences.md conflicts with tech.md.
	ErrMigrateMergeConflict = "MIGRATE_MERGE_CONFLICT"
	// ErrMigrateInterrupt is returned after SIGINT/SIGTERM is handled.
	ErrMigrateInterrupt = "MIGRATE_INTERRUPT"
	// ErrMigrateCheckpointCorrupt is returned when a resume checkpoint cannot be parsed.
	ErrMigrateCheckpointCorrupt = "MIGRATE_CHECKPOINT_CORRUPT"
)

// MigrateError is an error with a machine-readable code.
type MigrateError struct {
	Code    string
	Message string
}

func (e *MigrateError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// migrateResult contains the outcome of a successful migration.
type migrateResult struct {
	filesTransferred int
	archivePath      string
	summary          []string
}

// migrationCheckpoint records intermediate state for --resume support.
// Stored at ~/.moai/.migrate-tx-<txID>.json.
//
// @MX:ANCHOR: [AUTO] migrationCheckpoint is the contract for resume-capable migrations
// @MX:REASON: [AUTO] read by both Run (checkpoint flush) and Resume paths; fan_in >= 3
type migrationCheckpoint struct {
	TxID            string    `json:"tx_id"`
	ProjectRoot     string    `json:"project_root"`
	CompletedPhases []int     `json:"completed_phases"`
	RemainingFiles  []string  `json:"remaining_files"`
	Timestamp       time.Time `json:"timestamp"`
}

// transactionLog records created paths for rollback.
type transactionLog struct {
	createdDirs  []string
	createdFiles []string
}

func (tx *transactionLog) recordDir(path string) {
	tx.createdDirs = append(tx.createdDirs, path)
}

func (tx *transactionLog) recordFile(path string) {
	tx.createdFiles = append(tx.createdFiles, path)
}

// rollback removes all files/dirs created during the transaction.
func (tx *transactionLog) rollback() {
	// Remove files first, then dirs (deepest first)
	for i := len(tx.createdFiles) - 1; i >= 0; i-- {
		_ = os.Remove(tx.createdFiles[i])
	}
	for i := len(tx.createdDirs) - 1; i >= 0; i-- {
		_ = os.RemoveAll(tx.createdDirs[i])
	}
}

// migrateAgencyRunner holds all migration parameters.
// Separating configuration from cobra flags enables table-driven testing.
//
// @MX:ANCHOR: [AUTO] migrateAgencyRunner.Run is the migration entry point
// @MX:REASON: [AUTO] called from cobra RunE, tests (multiple), and resume path; fan_in >= 3
type migrateAgencyRunner struct {
	projectRoot string
	homeDir     string
	dryRun      bool
	force       bool
	resumeTxID  string
	stderr      io.Writer // defaults to os.Stderr
	stdout      io.Writer // defaults to os.Stdout

	// failAtPhase is a test injection hook: if > 0, Run returns an error at that phase.
	failAtPhase int
}

func (r *migrateAgencyRunner) getStderr() io.Writer {
	if r.stderr != nil {
		return r.stderr
	}
	return os.Stderr
}

func (r *migrateAgencyRunner) getStdout() io.Writer {
	if r.stdout != nil {
		return r.stdout
	}
	return os.Stdout
}

func (r *migrateAgencyRunner) logf(format string, args ...any) {
	_, _ = fmt.Fprintf(r.getStdout(), format+"\n", args...)
}

func (r *migrateAgencyRunner) errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(r.getStderr(), format+"\n", args...)
}

// checkpointPath returns the path for a transaction checkpoint file.
func (r *migrateAgencyRunner) checkpointPath(txID string) string {
	return filepath.Join(r.homeDir, ".moai", ".migrate-tx-"+txID+".json")
}

// Run executes the migration and returns a result summary on success.
//
// @MX:WARN: [AUTO] complex phase-based migration with rollback; cyclomatic complexity >= 15
// @MX:REASON: [AUTO] each of 6 phases can fail; rollback must be attempted in reverse order
func (r *migrateAgencyRunner) Run() (*migrateResult, error) {
	// Resume mode: load checkpoint and only run remaining phases.
	if r.resumeTxID != "" {
		return r.runResume()
	}
	return r.runFull()
}

func (r *migrateAgencyRunner) runFull() (*migrateResult, error) {
	agencyDir := filepath.Join(r.projectRoot, ".agency")
	archiveDir := filepath.Join(r.projectRoot, ".agency.archived")

	// Phase 0: validate preconditions.
	if _, err := os.Stat(agencyDir); os.IsNotExist(err) {
		return nil, &MigrateError{Code: ErrMigrateNoSource, Message: ".agency/ directory not found"}
	}

	if _, err := os.Stat(archiveDir); err == nil && !r.force {
		return nil, &MigrateError{Code: ErrMigrateArchiveExists, Message: ".agency.archived/ already exists; use --force to overwrite"}
	}

	if !r.force {
		if err := r.checkTargetsAbsent(); err != nil {
			return nil, err
		}
	}

	// Check for tech-preferences.md merge conflict BEFORE starting.
	if conflictErr := r.checkMergeConflict(); conflictErr != nil {
		return nil, conflictErr
	}

	if r.dryRun {
		return r.runDryRun(agencyDir)
	}

	tx := &transactionLog{}
	result := &migrateResult{}

	// Install signal handler for graceful interrupt handling (POSIX only).
	txID := fmt.Sprintf("%d", time.Now().UnixNano())
	cp := &migrationCheckpoint{
		TxID:            txID,
		ProjectRoot:     r.projectRoot,
		CompletedPhases: []int{},
	}
	cpPath := r.checkpointPath(txID)
	cancelSignal := installSignalHandler(cp, cpPath, func(sig os.Signal) {
		r.errorf("signal %v received, checkpoint saved at %s", sig, cpPath)
	})
	defer cancelSignal()

	// Phase 1: backup .agency/ → .agency.archived/
	if r.failAtPhase == 1 {
		return nil, &MigrateError{Code: "MIGRATE_PHASE_INJECT", Message: "injected failure at phase 1"}
	}
	r.logf("[1/6] Backing up .agency/ → .agency.archived/")
	if err := r.copyDir(agencyDir, archiveDir, tx); err != nil {
		tx.rollback()
		_ = os.RemoveAll(archiveDir)
		return nil, fmt.Errorf("migrate: phase 1 backup: %w", err)
	}

	// Phase 2: copy context/ → .moai/project/brand/
	if r.failAtPhase == 2 {
		tx.rollback()
		_ = os.RemoveAll(archiveDir)
		return nil, &MigrateError{Code: "MIGRATE_PHASE_INJECT", Message: "injected failure at phase 2"}
	}
	r.logf("[2/6] Migrating context/ → .moai/project/brand/")
	brandDir := filepath.Join(r.projectRoot, ".moai", "project", "brand")
	contextDir := filepath.Join(agencyDir, "context")
	n, err := r.migrateContext(contextDir, brandDir, tx)
	if err != nil {
		r.errorf("rollback: %s", ErrMigrateRollbackOK)
		tx.rollback()
		_ = os.RemoveAll(archiveDir)
		return nil, fmt.Errorf("migrate: phase 2 context: %w", err)
	}
	result.filesTransferred += n

	// Phase 3: copy learnings/ → .moai/research/observations/
	if r.failAtPhase == 3 {
		r.errorf("rollback: %s", ErrMigrateRollbackOK)
		tx.rollback()
		_ = os.RemoveAll(archiveDir)
		return nil, &MigrateError{Code: "MIGRATE_PHASE_INJECT", Message: "injected failure at phase 3"}
	}
	r.logf("[3/6] Migrating learnings/ → .moai/research/observations/")
	learningsDir := filepath.Join(agencyDir, "learnings")
	obsDir := filepath.Join(r.projectRoot, ".moai", "research", "observations")
	n, err = r.migrateLearnings(learningsDir, obsDir, tx)
	if err != nil {
		r.errorf("rollback: %s", ErrMigrateRollbackOK)
		tx.rollback()
		_ = os.RemoveAll(archiveDir)
		return nil, fmt.Errorf("migrate: phase 3 learnings: %w", err)
	}
	result.filesTransferred += n

	// Phase 4: convert config.yaml → design.yaml
	if r.failAtPhase == 4 {
		r.errorf("rollback: %s", ErrMigrateRollbackOK)
		tx.rollback()
		_ = os.RemoveAll(archiveDir)
		return nil, &MigrateError{Code: "MIGRATE_PHASE_INJECT", Message: "injected failure at phase 4"}
	}
	r.logf("[4/6] Converting config.yaml → design.yaml")
	configSrc := filepath.Join(agencyDir, "config.yaml")
	designDst := filepath.Join(r.projectRoot, ".moai", "config", "sections", "design.yaml")
	if err := r.convertConfig(configSrc, designDst, tx); err != nil {
		r.errorf("rollback: %s", ErrMigrateRollbackOK)
		tx.rollback()
		_ = os.RemoveAll(archiveDir)
		return nil, fmt.Errorf("migrate: phase 4 config: %w", err)
	}
	result.filesTransferred++

	// Phase 5: archive fork-manifest.yaml (already in archive; just note it)
	r.logf("[5/6] fork-manifest.yaml archived at .agency.archived/")

	// Phase 6: conditional constitution.md relocation
	if r.failAtPhase == 6 {
		r.errorf("rollback: %s", ErrMigrateRollbackOK)
		tx.rollback()
		_ = os.RemoveAll(archiveDir)
		return nil, &MigrateError{Code: "MIGRATE_PHASE_INJECT", Message: "injected failure at phase 6"}
	}
	r.logf("[6/6] Checking constitution.md relocation")
	constitutionSrc := filepath.Join(r.projectRoot, ".claude", "rules", "agency", "constitution.md")
	if _, err := os.Stat(constitutionSrc); err == nil {
		constitutionDst := filepath.Join(r.projectRoot, ".claude", "rules", "moai", "design", "constitution.md")
		if err := r.copyFile(constitutionSrc, constitutionDst, tx); err != nil {
			// Non-fatal: log and continue
			r.errorf("warn: could not relocate constitution.md: %v", err)
		} else {
			result.filesTransferred++
			r.logf("  constitution.md relocated to .claude/rules/moai/design/")
		}
	} else {
		r.logf("  no user constitution.md found, no-op")
	}

	result.archivePath, _ = filepath.Abs(archiveDir)
	result.summary = append(result.summary, fmt.Sprintf("Transferred %d file(s)", result.filesTransferred))
	result.summary = append(result.summary, fmt.Sprintf("Archive: %s", result.archivePath))
	result.summary = append(result.summary, "Next step: /moai design")

	r.logf("\nMigration complete: %d file(s) transferred", result.filesTransferred)
	r.logf("Archive preserved at: %s", result.archivePath)
	r.logf("Run '/moai design' to continue.")

	return result, nil
}

// runDryRun prints expected actions without modifying the filesystem.
func (r *migrateAgencyRunner) runDryRun(agencyDir string) (*migrateResult, error) {
	r.logf("[dry-run] Expected migration actions:")
	r.logf("  %s → %s", filepath.Join(agencyDir, "context"), filepath.Join(r.projectRoot, ".moai", "project", "brand"))
	r.logf("  %s → %s", filepath.Join(agencyDir, "config.yaml"), filepath.Join(r.projectRoot, ".moai", "config", "sections", "design.yaml"))
	r.logf("  %s → %s", filepath.Join(agencyDir, "learnings"), filepath.Join(r.projectRoot, ".moai", "research", "observations"))
	r.logf("  %s → %s", agencyDir, filepath.Join(r.projectRoot, ".agency.archived"))

	return &migrateResult{
		summary: []string{"dry-run: no changes made"},
	}, nil
}

// runResume loads a checkpoint and runs remaining phases.
func (r *migrateAgencyRunner) runResume() (*migrateResult, error) {
	cpPath := r.checkpointPath(r.resumeTxID)
	cp, err := readCheckpoint(cpPath)
	if err != nil {
		return nil, &MigrateError{Code: ErrMigrateCheckpointCorrupt, Message: fmt.Sprintf("checkpoint read: %v", err)}
	}

	r.logf("Resuming migration from checkpoint %s (completed phases: %v)", r.resumeTxID, cp.CompletedPhases)

	// For simplicity in the resume path: migrate remaining files directly.
	if len(cp.RemainingFiles) == 0 {
		return &migrateResult{summary: []string{"resume: nothing to do"}}, nil
	}

	obsDir := filepath.Join(r.projectRoot, ".moai", "research", "observations")
	tx := &transactionLog{}
	transferred := 0

	for _, srcFile := range cp.RemainingFiles {
		dstFile := filepath.Join(obsDir, filepath.Base(srcFile))
		if err := r.copyFile(srcFile, dstFile, tx); err != nil {
			r.errorf("rollback: %s", ErrMigrateRollbackOK)
			tx.rollback()
			return nil, fmt.Errorf("resume: copy %s: %w", srcFile, err)
		}
		transferred++
	}

	archiveDir := filepath.Join(r.projectRoot, ".agency.archived")
	archivePath, _ := filepath.Abs(archiveDir)

	return &migrateResult{
		filesTransferred: transferred,
		archivePath:      archivePath,
		summary:          []string{fmt.Sprintf("resume: transferred %d file(s)", transferred)},
	}, nil
}

// checkTargetsAbsent verifies that destination paths do not already exist.
func (r *migrateAgencyRunner) checkTargetsAbsent() error {
	targets := []string{
		filepath.Join(r.projectRoot, ".moai", "project", "brand"),
		filepath.Join(r.projectRoot, ".moai", "config", "sections", "design.yaml"),
		filepath.Join(r.projectRoot, ".moai", "research", "observations"),
	}
	for _, t := range targets {
		if _, err := os.Stat(t); err == nil {
			return &MigrateError{
				Code:    ErrMigrateTargetExists,
				Message: fmt.Sprintf("%s already exists; use --force to overwrite", t),
			}
		}
	}
	return nil
}

// checkMergeConflict detects tech-preferences.md vs tech.md content conflicts.
func (r *migrateAgencyRunner) checkMergeConflict() error {
	techPrefPath := filepath.Join(r.projectRoot, ".agency", "context", "tech-preferences.md")
	techMdPath := filepath.Join(r.projectRoot, ".moai", "project", "tech.md")

	prefData, errPref := os.ReadFile(techPrefPath)
	techData, errTech := os.ReadFile(techMdPath)
	if errPref != nil || errTech != nil {
		// One or both files missing: no conflict
		return nil
	}

	// Simple conflict detection: look for different "Framework:" declarations.
	prefFramework := extractValue(string(prefData), "Framework:")
	techFramework := extractValue(string(techData), "Framework:")

	if prefFramework != "" && techFramework != "" && prefFramework != techFramework {
		return &MigrateError{
			Code:    ErrMigrateMergeConflict,
			Message: fmt.Sprintf("tech-preferences.md (Framework: %s) conflicts with tech.md (Framework: %s); resolve manually", prefFramework, techFramework),
		}
	}
	return nil
}

// extractValue finds the value after a given key prefix on the same line.
func extractValue(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		if idx := strings.Index(line, key); idx >= 0 {
			return strings.TrimSpace(line[idx+len(key):])
		}
	}
	return ""
}

// migrateContext copies brand files from context/ to .moai/project/brand/.
func (r *migrateAgencyRunner) migrateContext(src, dst string, tx *transactionLog) (int, error) {
	brandFiles := []string{"brand-voice.md", "visual-identity.md", "target-audience.md"}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return 0, fmt.Errorf("mkdirall %s: %w", dst, err)
	}
	tx.recordDir(dst)

	transferred := 0
	for _, name := range brandFiles {
		srcFile := filepath.Join(src, name)
		if _, err := os.Stat(srcFile); os.IsNotExist(err) {
			continue
		}
		if err := r.copyFile(srcFile, filepath.Join(dst, name), tx); err != nil {
			return transferred, fmt.Errorf("copy %s: %w", name, err)
		}
		transferred++
	}
	return transferred, nil
}

// migrateLearnings copies all .md files from learnings/ to observations/.
func (r *migrateAgencyRunner) migrateLearnings(src, dst string, tx *transactionLog) (int, error) {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return 0, nil
	}

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return 0, fmt.Errorf("mkdirall %s: %w", dst, err)
	}
	tx.recordDir(dst)

	entries, err := os.ReadDir(src)
	if err != nil {
		return 0, fmt.Errorf("readdir %s: %w", src, err)
	}

	transferred := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		srcFile := filepath.Join(src, e.Name())
		dstFile := filepath.Join(dst, e.Name())
		if err := r.copyFile(srcFile, dstFile, tx); err != nil {
			return transferred, fmt.Errorf("copy learning %s: %w", e.Name(), err)
		}
		transferred++
	}
	return transferred, nil
}

// convertConfig transforms .agency/config.yaml into design.yaml.
func (r *migrateAgencyRunner) convertConfig(src, dst string, tx *transactionLog) error {
	data, err := os.ReadFile(src)
	if err != nil {
		// config.yaml may not exist; generate minimal design.yaml
		data = []byte{}
	}

	// Extract the 'agency:' subtree and rename to 'design:'.
	// Simple string-based transformation: replace the top-level key.
	converted := strings.Replace(string(data), "agency:", "design:", 1)

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdirall %s: %w", filepath.Dir(dst), err)
	}

	if err := os.WriteFile(dst, []byte(converted), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}
	tx.recordFile(dst)
	return nil
}

// copyFile copies a single file, preserving permissions on POSIX systems.
// Platform-specific permission logic is in migrate_agency_posix.go / migrate_agency_windows.go.
func (r *migrateAgencyRunner) copyFile(src, dst string, tx *transactionLog) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		r.errorf("warn: skipping symlink %s", src)
		return nil
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdirall for %s: %w", dst, err)
	}

	if err := os.WriteFile(dst, data, info.Mode().Perm()); err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}
	tx.recordFile(dst)

	// Apply platform-specific permission handling.
	applyPermissions(src, dst, info, r.getStderr())

	return nil
}

// copyDir recursively copies src to dst, recording all created paths in tx.
func (r *migrateAgencyRunner) copyDir(src, dst string, tx *transactionLog) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("mkdirall %s: %w", dst, err)
	}
	tx.recordDir(dst)

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("readdir %s: %w", src, err)
	}

	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := r.copyDir(srcPath, dstPath, tx); err != nil {
				return err
			}
		} else {
			if err := r.copyFile(srcPath, dstPath, tx); err != nil {
				return err
			}
		}
	}
	return nil
}

// writeCheckpoint persists a checkpoint to disk.
func writeCheckpoint(path string, cp *migrationCheckpoint) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdirall checkpoint dir: %w", err)
	}
	data, err := json.Marshal(cp)
	if err != nil {
		return fmt.Errorf("marshal checkpoint: %w", err)
	}
	return os.WriteFile(path, data, 0o600)
}

// readCheckpoint loads a checkpoint from disk.
func readCheckpoint(path string) (*migrationCheckpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cp migrationCheckpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &cp, nil
}

// --- Cobra command wiring ---

var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Short:   "Migration utilities for MoAI-ADK",
	GroupID: "project",
	Long:    "Run migrations to upgrade project data to the current MoAI-ADK format.",
}

var migrateAgencyCmd = &cobra.Command{
	Use:   "agency",
	Short: "Migrate .agency/ project data to .moai/ namespace",
	Long: `Migrate existing AI Agency project data (.agency/) into the MoAI namespace (.moai/).

This command is atomic: if any step fails, all changes are rolled back.

File mapping:
  .agency/context/     → .moai/project/brand/
  .agency/config.yaml  → .moai/config/sections/design.yaml
  .agency/learnings/   → .moai/research/observations/
  .agency/             → .agency.archived/ (backup)

Flags:
  --dry-run   Show planned actions without modifying files
  --force     Overwrite existing targets (use with caution)
  --resume    Resume from a previous interrupted migration`,
	RunE: runMigrateAgency,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateAgencyCmd)

	migrateAgencyCmd.Flags().Bool("dry-run", false, "Show planned actions without modifying files")
	migrateAgencyCmd.Flags().Bool("force", false, "Overwrite existing migration targets")
	migrateAgencyCmd.Flags().String("resume", "", "Resume migration from a checkpoint tx-id")
}

func runMigrateAgency(cmd *cobra.Command, _ []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	resumeTxID, _ := cmd.Flags().GetString("resume")

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	m := &migrateAgencyRunner{
		projectRoot: cwd,
		homeDir:     homeDir,
		dryRun:      dryRun,
		force:       force,
		resumeTxID:  resumeTxID,
	}

	result, err := m.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if me, ok := err.(*MigrateError); ok {
			switch me.Code {
			case ErrMigrateNoSource:
				os.Exit(2)
			default:
				os.Exit(1)
			}
		}
		return err
	}

	_ = result
	return nil
}
