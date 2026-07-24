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
	"runtime"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	updatemerge "github.com/modu-ai/moai-adk/internal/cli/update/merge"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
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

	// Force: propagate the CLI --force flag to the Step 3.5 agency-migration
	// adapter (issue #1132). When true, an explicit `moai update --force`
	// performs a real forced migration (overwriting existing targets) instead
	// of the already-migrated skip no-op.
	Force bool

	// RunMigrateAgency: optional override for the .agency/ → .moai/ migration
	// invocation (REQ-VVCR-025). nil defaults to a no-op when not testing;
	// production callers from M5 inject the canonical runMigrateAgency
	// adapter that proxies cobra command flags.
	RunMigrateAgency func(projectRoot string, dryRun, force bool, out io.Writer) error
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
		_, _ = fmt.Fprintln(out, "[clean-reinstall] not a v2 project — no-op")
		return result, nil
	}

	_, _ = fmt.Fprintf(out, "[clean-reinstall] v2 fingerprint detected (signals: version=%v agency=%v deprecated=%v)\n",
		fp.V2DetectedViaVersion, fp.V2DetectedViaAgencyDir, fp.V2DetectedViaDeprecatedPath)

	// ---------------------------------------------------------------
	// Step 2 — PRESERVE inventory snapshot
	// ---------------------------------------------------------------
	inv, err := buildPreserveInventory(projectRoot)
	if err != nil {
		return result, fmt.Errorf("step 2: build PRESERVE inventory: %w", err)
	}
	result.Inventory = inv

	_, _ = fmt.Fprintf(out, "[clean-reinstall] PRESERVE inventory: %d files\n", len(inv.Files))

	// Pre-snapshot hashes for Step 7 integrity verification.
	hashesPre, err := computeInventoryHashes(projectRoot, inv)
	if err != nil {
		return result, fmt.Errorf("step 2: compute pre-snapshot hashes: %w", err)
	}

	// Dry-run early return — emit planned actions and stop before any
	// filesystem mutation.
	if opts.DryRun {
		_, _ = fmt.Fprintln(out, "[clean-reinstall] DRY-RUN — no filesystem mutations performed")
		_, _ = fmt.Fprintf(out, "[clean-reinstall] Would back up %d files into .moai/backups/v2-to-v3-<stamp>/\n", len(inv.Files))
		if planRemoved, scanErr := scanDeprecatedPaths(projectRoot); scanErr == nil {
			_, _ = fmt.Fprintf(out, "[clean-reinstall] Would remove %d deprecated paths\n", len(planRemoved))
			result.RemovedPaths = planRemoved
		}
		if fp.V2DetectedViaAgencyDir {
			_, _ = fmt.Fprintln(out, "[clean-reinstall] Would auto-invoke `moai migrate agency` for .agency/ contents")
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

	_, _ = fmt.Fprintf(out, "[clean-reinstall] Backup created at %s\n", finalBackupDir)

	// ---------------------------------------------------------------
	// Step 3.5 — Auto-invoke .agency/ migration if present (REQ-VVCR-025)
	// ---------------------------------------------------------------
	if fp.V2DetectedViaAgencyDir && opts.RunMigrateAgency != nil {
		if err := opts.RunMigrateAgency(projectRoot, opts.DryRun, opts.Force, out); err != nil {
			return result, fmt.Errorf("step 3.5: auto-invoke migrate agency: %w", err)
		}
		result.AgencyMigrated = true
		_, _ = fmt.Fprintln(out, "[clean-reinstall] .agency/ → .moai/ migration completed")
	}

	// ---------------------------------------------------------------
	// Step 3.9 — Backup-coverage abort gate (pre-destructive invariant)
	// ---------------------------------------------------------------
	// Clean-path counterpart of the normal update path's
	// verifyNamespaceBackupCoverage gate: every PRESERVE-inventory file still
	// on disk MUST have a copy in the Step 3 backup before the first
	// destructive step (Step 4 REMOVE / Step 5 Deploy) runs. Verifies against
	// the STRICT inventory the snapshot actually captured — not the
	// conservative superset — so real v2 upgrades do not spuriously abort.
	// Placed AFTER Step 3.5 so anything the agency migration did to the
	// backup directory is also caught. On violation the backup dir is left in
	// place for forensic inspection (HARD-5 backup-before-removal).
	if covErr := verifyPreserveBackupCoverage(projectRoot, inv, finalBackupDir); covErr != nil {
		return result, fmt.Errorf("step 3.9: backup coverage gate: %w", covErr)
	}

	// ---------------------------------------------------------------
	// Step 4 — REMOVE deprecated paths
	// ---------------------------------------------------------------
	//
	// REQ-CRR-006 / AC-CRR-006: the removal count reported below is derived
	// from the FILESYSTEM DIFF, not from the planned-list length.
	//
	// scanDeprecatedPaths already existence-filters (os.Lstat + silent skip on
	// IsNotExist), so its return value is the set of deprecated paths that
	// actually exist pre-REMOVE. Re-scanning post-REMOVE yields those that
	// survived, and the difference is what was genuinely removed. On a project
	// with no deprecated residue both scans return empty and the count is 0 —
	// the case that previously emitted the phantom `Removed N deprecated paths`
	// line reported by #1084 (the log printed the planned count while `git diff`
	// showed nothing removed).
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

	// Post-REMOVE re-scan → actual removal count (AC-CRR-006(a)).
	remaining, err := scanDeprecatedPaths(projectRoot)
	if err != nil {
		return result, fmt.Errorf("step 4: re-scan deprecated paths: %w", err)
	}
	removedCount := len(deprecated) - len(remaining)

	// AC-CRR-006(b)(c)(d): emit the removal line ONLY when the actual count is
	// positive; otherwise emit the informational no-op line.
	if removedCount > 0 {
		_, _ = fmt.Fprintf(out, "[clean-reinstall] Removed %d deprecated paths\n", removedCount)
	} else {
		_, _ = fmt.Fprintln(out, "[clean-reinstall] No deprecated paths found to remove")
	}

	// ---------------------------------------------------------------
	// Step 4.5 — Capture user config for merge-preservation
	// (SPEC-UPDATE-REINSTALL-LOOP-001 R3, REQ-RIL-007/008/009/010)
	// ---------------------------------------------------------------
	// The Step 5 force-deploy (forceUpdate=true) overwrites .claude/settings.json
	// and every .moai/config/sections/*.yaml with template defaults, bypassing the
	// normal `moai update` path's 3-way merge and silently dropping user
	// customizations (effortLevel/permissions/model, operator name, conversation
	// language, brand tokens — issue #1084). Capture the user's config NOW (after
	// Step 4 removed deprecated sections, before Step 5 clobbers the rest) so
	// Step 5.5 can restore it with the SAME merge machinery the normal path uses.
	// The clean-reinstall must not be a lower-protection bypass of the normal path.
	configBackupPath, cfgBackupErr := backup.BackupMoaiConfig(projectRoot)
	if cfgBackupErr != nil {
		return result, fmt.Errorf("step 4.5: backup .moai/config for merge-preservation: %w", cfgBackupErr)
	}
	// Mergeable root files handled by the normal path's 3-way engine (identical
	// set to update.go collectMergeableFiles). settings.json base is unavailable
	// in the embedded FS (it ships as settings.json.tmpl), so MergeUserFiles
	// preserves the user's file wholesale — matching the normal path exactly.
	var mergeableBackups []updatemerge.FileBackup
	for _, mf := range []string{".claude/settings.json", ".moai/status_line.sh"} {
		if data, readErr := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(mf))); readErr == nil {
			mergeableBackups = append(mergeableBackups, updatemerge.FileBackup{Path: mf, Data: data})
		}
	}
	// Backup .gitignore for the user-pattern EntryMerge after deploy (issues
	// #1131/#1094): the Step 5 force-deploy overwrites .gitignore with the
	// template version, silently dropping user-added patterns. This mirrors
	// the normal `moai update` path's Backup-step capture. .gitignore is
	// intentionally NOT part of the PRESERVE inventory — Step 7 requires
	// byte-identical restore, while .gitignore must MERGE (deployed template
	// content + user additions appended).
	var gitignoreBackup []byte
	if data, readErr := os.ReadFile(filepath.Join(projectRoot, ".gitignore")); readErr == nil {
		gitignoreBackup = data
	}

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
		if _, loadErr := mgr.Load(projectRoot); loadErr != nil {
			return result, fmt.Errorf("step 5: load manifest: %w", loadErr)
		}
	}

	// Build the deploy context with detected paths — identical to the normal
	// `moai update` "Deploy Templates" step (see runUpdate). A bare context
	// would leave SmartPATH/GoBinPath/HomeDir empty, which renders settings.json
	// "PATH": "" and the status_line.sh "/moai" fallback, stripping the moai
	// binary from PATH in every downstream session after a v2→v3 upgrade and
	// breaking the statusline plus all PATH-resolved MCP servers (moai-lsp, npx).
	homeDir, _ := userHomeDir()
	goBinPath := detectGoBinPathForUpdate(homeDir)
	tmplCtx := template.NewTemplateContext(
		template.WithGoBinPath(goBinPath),
		template.WithResolvedMoaiPath(resolveMoaiExecutable()),
		template.WithHomeDir(homeDir),
		template.WithSmartPATH(template.BuildSmartPATH()),
		template.WithPlatform(runtime.GOOS),
		template.WithVersion(version.GetVersion()),
		template.WithHookOptIn(readHookOptInEnabled(projectRoot)),
	)

	if deployErr := deployer.Deploy(ctx, projectRoot, mgr, tmplCtx); deployErr != nil {
		return result, fmt.Errorf("step 5: reinstall templates: %w", deployErr)
	}
	_, _ = fmt.Fprintln(out, "[clean-reinstall] Embedded templates reinstalled")

	// ---------------------------------------------------------------
	// Step 5.5 — Restore user config via the normal-path 3-way merge
	// (SPEC-UPDATE-REINSTALL-LOOP-001 R3, REQ-RIL-007/008/009/010)
	// ---------------------------------------------------------------
	// Mirrors the "Restore Settings" step of the normal `moai update` path
	// (update.go): backup.RestoreMoaiConfig 3-way merges .moai/config/sections/*.yaml
	// (user values win over template defaults; template additions merge in), and
	// updatemerge.MergeUserFiles restores .claude/settings.json / status_line.sh.
	// Using the SAME functions makes the clean path's config protection equivalent
	// to the normal path's (AC-RIL-007), not a lower-protection bypass. The
	// preserved config files are OUTSIDE the PRESERVE inventory (inv.Files), so the
	// Step 7 byte-identical integrity check does not flag their merged content.
	// Log-claim accuracy: each restoration below reports its own success line
	// scoped to what actually ran, instead of one unconditional
	// "merge-preserved" claim emitted even when nothing was backed up.
	if configBackupPath != "" {
		// nil recorder: the noise-suppression merge-history ledger stays in the
		// normal update flow; the clean-reinstall path does not need it.
		if restoreErr := backup.RestoreMoaiConfig(projectRoot, configBackupPath, nil); restoreErr != nil {
			return result, fmt.Errorf("step 5.5: restore .moai/config sections: %w", restoreErr)
		}
		_, _ = fmt.Fprintln(out, "[clean-reinstall] .moai/config/sections/*.yaml merge-restored (user values preserved)")
		// Backup-dir accumulation cap — the same pruning the normal path
		// performs after its restore step (backup.CleanupOldBackups). Only
		// timestamped config-backup dirs under .moai-backups/ are candidates;
		// the v2-to-v3-* forensic backups live under .moai/backups/ and are
		// never touched by this cleanup.
		if deleted := backup.CleanupOldBackups(projectRoot, 5); deleted > 0 {
			_, _ = fmt.Fprintf(out, "[clean-reinstall] Cleaned up %d old config backup(s)\n", deleted)
		}
	}
	if len(mergeableBackups) > 0 {
		// MergeUserFiles preserves the user's version on any merge failure, so a
		// warning (not a hard error) matches the normal path's tolerance.
		if mergeErr := updatemerge.MergeUserFiles(projectRoot, mergeableBackups, out); mergeErr != nil {
			_, _ = fmt.Fprintf(out, "[clean-reinstall] settings merge warning: %v\n", mergeErr)
		} else {
			_, _ = fmt.Fprintln(out, "[clean-reinstall] settings.json / status_line.sh merge-preserved")
		}
	}
	// Merge .gitignore: preserve user-added patterns via EntryMerge (issues
	// #1131/#1094) — the same updatemerge.MergeGitignoreFile call the normal
	// `moai update` path performs after deploy. Warn-not-fail matches that
	// path's tolerance.
	if len(gitignoreBackup) > 0 {
		gitignorePath := filepath.Join(projectRoot, ".gitignore")
		if mergeErr := updatemerge.MergeGitignoreFile(gitignorePath, gitignoreBackup); mergeErr != nil {
			_, _ = fmt.Fprintf(out, "[clean-reinstall] .gitignore merge warning: %v\n", mergeErr)
		} else {
			_, _ = fmt.Fprintln(out, "[clean-reinstall] .gitignore user patterns preserved")
		}
	}

	// One-shot v2->v3 deny-rule migration (issue #1101): the wholesale
	// settings.json preservation above also preserves retired v2-era deny
	// entries (Write/Grep/Glob x 4 protected paths) that the v3 template no
	// longer ships, causing Claude Code startup warnings every session. Strip
	// exact-match retired entries only; user customizations are untouched.
	if stripErr := stripRetiredV2DenyEntries(projectRoot, out); stripErr != nil {
		_, _ = fmt.Fprintf(out, "[clean-reinstall] deny-rule migration warning: %v\n", stripErr)
	}

	// ---------------------------------------------------------------
	// Step 6 — MERGE-back PRESERVE inventory
	// ---------------------------------------------------------------
	if err := mergeBackPreserveInventory(projectRoot, inv, finalBackupDir); err != nil {
		return result, fmt.Errorf("step 6: merge-back PRESERVE inventory: %w", err)
	}
	_, _ = fmt.Fprintln(out, "[clean-reinstall] PRESERVE inventory restored")

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
		_, _ = fmt.Fprintf(out, "[clean-reinstall] Integrity check FAILED: %d mismatches\n", len(mismatches))
		return result, fmt.Errorf("step 7: PRESERVE integrity violation on %d paths (backup retained at %s)", len(mismatches), finalBackupDir)
	}
	// Log-claim accuracy: the integrity check covers ONLY the PRESERVE
	// inventory hashes (inv.Files) — not merged config, settings, or
	// .gitignore, whose content legitimately changes via the merges above.
	_, _ = fmt.Fprintf(out, "[clean-reinstall] Integrity check PASSED (%d PRESERVE-inventory files verified)\n", len(hashesPre))

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
