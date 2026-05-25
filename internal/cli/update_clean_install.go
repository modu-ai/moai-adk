// Package cli — update_clean_install.go
//
// Clean reinstall orchestration for the v2-to-v3 upgrade path
// (SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 REQ-VVCR-002..004, REQ-VVCR-010..025,
// AC-VVCR-002 / AC-VVCR-006..013).
//
// Runs the 7-step canonical order:
//
//	Step 1 — Detect v2 fingerprint (delegate to detectV2Fingerprint).
//	Step 2 — PRESERVE inventory snapshot (buildPreserveInventory).
//	Step 3 — Backup at .moai/backups/v2-to-v3-{stamp}/ (snapshotPreserveInventory).
//	Step 4 — REMOVE deprecated paths (scanDeprecatedPaths + backupDeprecatedPaths).
//	Step 5 — Reinstall embedded templates (deployer.Deploy).
//	Step 6 — MERGE-back PRESERVE inventory (mergeBackPreserveInventory).
//	Step 7 — Integrity verification (compare pre/post hashes).
//
// Step 3 also auto-invokes runMigrateAgency when `.agency/` is present
// (REQ-VVCR-025) — the migration runs BEFORE any REMOVE operation so that
// .agency/ contents are preserved into .moai/ before the REMOVE step
// purges legacy paths.
//
// This orchestration is invoked by M5 from runUpdate when
// detectV2Fingerprint returns IsV2: true.

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
)

// CleanReinstallOptions configures runCleanReinstall behavior.
//
// All fields are optional; sensible defaults apply when zero-valued. The
// struct enables M5 callers to inject dependencies cleanly without
// re-plumbing the cobra command layer.
type CleanReinstallOptions struct {
	// DryRun: when true, emit planned actions but make no filesystem mutations.
	// REQ-VVCR-028.
	DryRun bool

	// Out: progress / diagnostic writer. nil defaults to os.Stderr.
	Out io.Writer

	// Deployer: optional injected template deployer. When nil, the orchestrator
	// constructs the default deployer via template.NewDeployerWithRendererAndForceUpdate
	// using the supplied EmbeddedFS (or template.NewEmbeddedFS()).
	Deployer template.Deployer

	// EmbeddedFS: optional injected embedded template filesystem. nil defaults
	// to the result of template.EmbeddedTemplates(). Type is fs.FS to match
	// the upstream constructor signature.
	EmbeddedFS fs.FS

	// Manifest: optional injected manifest manager. nil defaults to a
	// new manifest.Manager using projectRoot.
	Manifest manifest.Manager

	// RunMigrateAgency: optional override for the .agency/ → .moai/ migration
	// invocation (REQ-VVCR-025). nil defaults to a no-op when not testing;
	// production callers from M5 inject the canonical runMigrateAgency
	// adapter that proxies cobra command flags.
	RunMigrateAgency func(projectRoot string, dryRun bool, out io.Writer) error
}

// CleanReinstallResult summarizes the outcome of runCleanReinstall for
// telemetry / dry-run output.
type CleanReinstallResult struct {
	// Detected: the fingerprint that triggered the reinstall.
	Detected V2Fingerprint

	// BackupDir: absolute path of the namespace backup directory created in
	// Step 3 (.moai/backups/v2-to-v3-<stamp>/). Empty when inventory is empty.
	BackupDir string

	// RemovedPaths: project-root-relative paths removed in Step 4.
	RemovedPaths []string

	// AgencyMigrated: true when runMigrateAgency was auto-invoked (REQ-VVCR-025).
	AgencyMigrated bool

	// Inventory: the PRESERVE inventory snapshot from Step 2.
	Inventory PreserveInventory

	// IntegrityPassed: true when Step 7 confirmed PRESERVE hashes match
	// before/after. False on any mismatch (with details in IntegrityMismatches).
	IntegrityPassed bool

	// IntegrityMismatches: project-root-relative paths whose hash differed
	// between snapshot time and post-reinstall.
	IntegrityMismatches []string

	// DryRun: echo of the DryRun input flag for caller convenience.
	DryRun bool
}

// runCleanReinstall executes the 7-step canonical clean-reinstall order.
//
// Pre-conditions:
//   - projectRoot exists and is a valid project directory
//   - detectV2Fingerprint already returned IsV2: true (caller's responsibility)
//
// Post-conditions on success:
//   - Backup exists at result.BackupDir (when inventory non-empty)
//   - All deprecated paths from defs.DeprecatedPaths are removed
//   - Embedded templates are deployed
//   - PRESERVE inventory is restored byte-identical (verified in Step 7)
//
// Errors abort the operation mid-flight; the partially-written backup
// directory is left in place for forensic inspection (HARD-5 atomicity
// guarantee — backup-before-removal).
//
// @MX:ANCHOR: 7-step canonical clean-reinstall orchestrator. Called by M5
// from runUpdate when V2Fingerprint.IsV2 is true.
// @MX:REASON: REQ-VVCR-004 specifies the canonical step order; deviations
// risk PRESERVE data loss (HARD-1..HARD-5). Step reordering MUST trigger
// a spec amendment.
func runCleanReinstall(ctx context.Context, projectRoot string, opts CleanReinstallOptions) (CleanReinstallResult, error) {
	out := opts.Out
	if out == nil {
		out = os.Stderr
	}

	if projectRoot == "" {
		return CleanReinstallResult{}, errors.New("runCleanReinstall: empty projectRoot")
	}

	result := CleanReinstallResult{DryRun: opts.DryRun}

	// ---------------------------------------------------------------
	// Step 1 — Detect v2 fingerprint
	// ---------------------------------------------------------------
	fp, err := detectV2Fingerprint(projectRoot)
	if err != nil {
		return result, fmt.Errorf("step 1: detect v2 fingerprint: %w", err)
	}
	result.Detected = fp
	if !fp.IsV2 {
		// Caller should not have invoked us; return gracefully without
		// mutations. This is REQ-VVCR-027 (idempotency on clean v3 projects).
		fmt.Fprintln(out, "[clean-reinstall] not a v2 project — no-op")
		return result, nil
	}

	fmt.Fprintf(out, "[clean-reinstall] v2 fingerprint detected (signals: version=%v agency=%v deprecated=%v)\n",
		fp.V2DetectedViaVersion, fp.V2DetectedViaAgencyDir, fp.V2DetectedViaDeprecatedPath)

	// ---------------------------------------------------------------
	// Step 2 — PRESERVE inventory snapshot
	// ---------------------------------------------------------------
	inv, err := buildPreserveInventory(projectRoot)
	if err != nil {
		return result, fmt.Errorf("step 2: build PRESERVE inventory: %w", err)
	}
	result.Inventory = inv

	fmt.Fprintf(out, "[clean-reinstall] PRESERVE inventory: %d files\n", len(inv.Files))

	// Pre-snapshot hashes for Step 7 integrity verification.
	hashesPre, err := computeInventoryHashes(projectRoot, inv)
	if err != nil {
		return result, fmt.Errorf("step 2: compute pre-snapshot hashes: %w", err)
	}

	// Dry-run early return — emit planned actions and stop before any
	// filesystem mutation.
	if opts.DryRun {
		fmt.Fprintln(out, "[clean-reinstall] DRY-RUN — no filesystem mutations performed")
		fmt.Fprintf(out, "[clean-reinstall] Would back up %d files into .moai/backups/v2-to-v3-<stamp>/\n", len(inv.Files))
		if planRemoved, scanErr := scanDeprecatedPaths(projectRoot); scanErr == nil {
			fmt.Fprintf(out, "[clean-reinstall] Would remove %d deprecated paths\n", len(planRemoved))
			result.RemovedPaths = planRemoved
		}
		if fp.V2DetectedViaAgencyDir {
			fmt.Fprintln(out, "[clean-reinstall] Would auto-invoke `moai migrate agency` for .agency/ contents")
			result.AgencyMigrated = true
		}
		return result, nil
	}

	// ---------------------------------------------------------------
	// Step 3 — Backup at .moai/backups/v2-to-v3-{stamp}/
	// ---------------------------------------------------------------
	stamp := newNamespaceBackupStamp()
	baseBackupRoot := filepath.Join(projectRoot, defs.MoAIDir, defs.NamespaceBackupsSubdir)
	v2BackupDir := filepath.Join(baseBackupRoot, "v2-to-v3-"+stamp)

	// Collision handling: append -1, -2, ... if same-second directory exists.
	finalBackupDir, err := resolveV2BackupDir(v2BackupDir)
	if err != nil {
		return result, fmt.Errorf("step 3: resolve backup dir: %w", err)
	}
	if mkErr := os.MkdirAll(finalBackupDir, 0o755); mkErr != nil {
		return result, fmt.Errorf("step 3: create backup dir: %w", mkErr)
	}

	if err := snapshotPreserveInventory(projectRoot, inv, finalBackupDir); err != nil {
		return result, fmt.Errorf("step 3: snapshot PRESERVE inventory: %w", err)
	}
	result.BackupDir = finalBackupDir

	fmt.Fprintf(out, "[clean-reinstall] Backup created at %s\n", finalBackupDir)

	// ---------------------------------------------------------------
	// Step 3.5 — Auto-invoke .agency/ migration if present (REQ-VVCR-025)
	// ---------------------------------------------------------------
	if fp.V2DetectedViaAgencyDir && opts.RunMigrateAgency != nil {
		if err := opts.RunMigrateAgency(projectRoot, opts.DryRun, out); err != nil {
			return result, fmt.Errorf("step 3.5: auto-invoke migrate agency: %w", err)
		}
		result.AgencyMigrated = true
		fmt.Fprintln(out, "[clean-reinstall] .agency/ → .moai/ migration completed")
	}

	// ---------------------------------------------------------------
	// Step 4 — REMOVE deprecated paths
	// ---------------------------------------------------------------
	deprecated, err := scanDeprecatedPaths(projectRoot)
	if err != nil {
		return result, fmt.Errorf("step 4: scan deprecated paths: %w", err)
	}
	for _, rel := range deprecated {
		abs := filepath.Join(projectRoot, filepath.FromSlash(rel))
		if rmErr := os.RemoveAll(abs); rmErr != nil {
			return result, fmt.Errorf("step 4: remove %s: %w", rel, rmErr)
		}
	}
	result.RemovedPaths = deprecated
	fmt.Fprintf(out, "[clean-reinstall] Removed %d deprecated paths\n", len(deprecated))

	// ---------------------------------------------------------------
	// Step 5 — Reinstall embedded templates
	// ---------------------------------------------------------------
	deployer := opts.Deployer
	if deployer == nil {
		embedded := opts.EmbeddedFS
		if embedded == nil {
			embeddedFS, embErr := template.EmbeddedTemplates()
			if embErr != nil {
				return result, fmt.Errorf("step 5: load embedded templates: %w", embErr)
			}
			embedded = embeddedFS
		}
		renderer := template.NewRenderer(embedded)
		deployer = template.NewDeployerWithRendererAndForceUpdate(embedded, renderer, true)
	}

	mgr := opts.Manifest
	if mgr == nil {
		mgr = manifest.NewManager()
	}

	// Use a default template context — same as runTemplateSync default path.
	tmplCtx := template.NewTemplateContext()

	if deployErr := deployer.Deploy(ctx, projectRoot, mgr, tmplCtx); deployErr != nil {
		return result, fmt.Errorf("step 5: reinstall templates: %w", deployErr)
	}
	fmt.Fprintln(out, "[clean-reinstall] Embedded templates reinstalled")

	// ---------------------------------------------------------------
	// Step 6 — MERGE-back PRESERVE inventory
	// ---------------------------------------------------------------
	if err := mergeBackPreserveInventory(projectRoot, inv, finalBackupDir); err != nil {
		return result, fmt.Errorf("step 6: merge-back PRESERVE inventory: %w", err)
	}
	fmt.Fprintln(out, "[clean-reinstall] PRESERVE inventory restored")

	// ---------------------------------------------------------------
	// Step 7 — Integrity verification (REQ-VVCR-023)
	// ---------------------------------------------------------------
	hashesPost, err := computeInventoryHashes(projectRoot, inv)
	if err != nil {
		return result, fmt.Errorf("step 7: compute post-reinstall hashes: %w", err)
	}

	var mismatches []string
	for rel, preHash := range hashesPre {
		postHash, ok := hashesPost[rel]
		if !ok {
			mismatches = append(mismatches, rel)
			continue
		}
		if preHash != postHash {
			mismatches = append(mismatches, rel)
		}
	}
	result.IntegrityMismatches = mismatches
	result.IntegrityPassed = len(mismatches) == 0

	if !result.IntegrityPassed {
		fmt.Fprintf(out, "[clean-reinstall] Integrity check FAILED: %d mismatches\n", len(mismatches))
		return result, fmt.Errorf("step 7: PRESERVE integrity violation on %d paths (backup retained at %s)", len(mismatches), finalBackupDir)
	}
	fmt.Fprintln(out, "[clean-reinstall] Integrity check PASSED")

	return result, nil
}

// resolveV2BackupDir handles same-second collision avoidance for the
// .moai/backups/v2-to-v3-<stamp>/ directory. Mirrors resolveNamespaceBackupDir
// semantics (NFR-UNP-004 numeric suffix).
//
// Returns the absolute path of a directory that does NOT exist yet (suitable
// for os.MkdirAll). The directory is NOT created here.
func resolveV2BackupDir(candidate string) (string, error) {
	if _, err := os.Stat(candidate); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
		return "", fmt.Errorf("stat backup candidate: %w", err)
	}
	for i := 1; i < 1000; i++ {
		next := fmt.Sprintf("%s-%d", candidate, i)
		if _, err := os.Stat(next); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return next, nil
			}
			return "", fmt.Errorf("stat backup candidate %d: %w", i, err)
		}
	}
	return "", fmt.Errorf("resolveV2BackupDir: exhausted 1000 candidates for %s", candidate)
}
