package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/cli/update/report"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/runtime/gobin"
	"github.com/modu-ai/moai-adk/internal/shell"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/tui"
	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// symProgress was retired by SPEC-V3R6-UPDATE-PROGRESS-001 M1. All former
// callers now use tui.ProgressLine which renders its own progress glyph
// from the theme. Kept as a comment for historical reference; the migration
// eliminated 22 "\r"-pair sites that previously caused trailing-character
// corruption when short messages overwrote longer progress messages.

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Sync MoAI-ADK project templates to the latest version",
	GroupID: "project",
	Long:    "Check for binary updates, install if available, then synchronize embedded templates with the project.\n\nNote: moai init / moai update do NOT auto-enter a worktree. To work inside a worktree, use the launcher flag (moai cc -w <name>) or create one explicitly (moai worktree new <SPEC-ID>).",
	PreRunE: validateUpdateFlags,
	RunE:    runUpdate,
}

// validateUpdateFlags validates update flag values before execution.
// SPEC-MODEL-PROFILE-MATRIX-001 (REQ-MPM-015/017): an out-of-set --profile value
// exits non-zero with a usage error naming the closed set {max, medium, low}.
func validateUpdateFlags(cmd *cobra.Command, _ []string) error {
	profileFlag := getStringFlag(cmd, "profile")
	if profileFlag != "" && !config.IsValidProfile(profileFlag) {
		return fmt.Errorf("invalid --profile value %q: must be one of: max, medium, low", profileFlag)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().Bool("check", false, "Check if a newer binary version is available (informational)")
	updateCmd.Flags().Bool("shell-env", false, "Configure shell environment variables for Claude Code")
	updateCmd.Flags().BoolP("config", "c", false, "Re-run the init wizard to edit project configuration (no template sync; bare 'moai update' syncs templates)")
	updateCmd.Flags().Bool("force", false, "Force update: bypass version-match skip, force backup+merge, and overwrite archive drift (backed up to .moai/archive/skills/v2.16-drift-<UTC-timestamp>/)")
	updateCmd.Flags().Bool("yes", false, "Auto-confirm all prompts (CI/CD mode)")
	updateCmd.Flags().Bool("templates-only", false, "Skip binary update, sync templates only")
	updateCmd.Flags().Bool("binary", false, "Update binary only, skip template sync")
	updateCmd.Flags().Bool("dry-run", false, "Show planned archive and install operations without modifying the filesystem")
	updateCmd.Flags().Bool("no-hooks", false, "Skip git hook installation (REQ-CIAUT-002)")
	updateCmd.Flags().String("restore", "", "Restore .moai/config from a backup directory left by a previous update (works on a tree whose .moai/config/sections/system.yaml was destroyed)")
	updateCmd.Flags().Bool("verbose", false, "Show all warnings including acknowledged reserved-name and 3-way merge fallback notices (diagnostic mode; SPEC-V3R6-UPDATE-NOISE-001 REQ-UN-005/010)")

	// SPEC-MODEL-PROFILE-MATRIX-001 (REQ-MPM-015/017): --profile override. When
	// provided, persists the value to llm.profile (no agent frontmatter mutation).
	// The retired --plan-type flag is no longer exposed.
	updateCmd.Flags().String("profile", "", "Override the model+effort profile: max, medium, or low (persists to llm.yaml profile)")
}

// @MX:ANCHOR: [AUTO] runUpdate orchestrates binary update and template synchronization
// @MX:REASON: [AUTO] fan_in=3, called from update.go init(), coverage_test.go, remaining_coverage_test.go
// @MX:NOTE: [AUTO] M4-S4d-1 DDD migration — converts 18 print sites to tui primitives.
// Pattern: tui.KV (Current/Latest/New version), tui.CheckLine (warn/run/err states), tui.Pill
// (final outcomes such as Binary updated · Already up to date · Skipped). Body preserved,
// no impact on external callers. Design source: screens.jsx ScreenUpdate (mirrors the M4-S4a~c IMPROVE pattern).
//
// runUpdate checks for binary updates first, then synchronizes embedded
// templates with the project directory. If a newer binary is installed,
// the process re-execs itself so the latest templates are used.
//
// Reconfigure vs template sync (SPEC-V3R6-CLI-CONFIG-INTEGRITY-001 REQ-CCI-001/002):
//
//	`moai update` (bare)    → binary update + template sync (3-way merge).
//	`moai update -c`        → short-circuits to runInitWizard(cmd, true) and
//	                          performs NO template sync. It re-runs the init
//	                          wizard to edit project configuration (model
//	                          policy, dev mode, git strategy, credentials).
//	                          See the `if editConfig { ... }` block in this
//	                          function for the short-circuit.
//
// Flags:
//
//	-c, --config: Re-run the init wizard to edit project configuration; does NOT
//	              synchronize templates (use bare `moai update` for that).
//	--check: Check if a newer binary version is available (informational)
//	--force: Force update with these effects (SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001):
//	  * bypass version-match skip-sync branch (template sync always runs)
//	  * force backup and merge confirmation through
//	  * overwrite legacy-skill archive drift; the existing archive directory is
//	    moved to .moai/archive/skills/v2.16-drift-<YYYYMMDDTHHMMSSZ>/<id>/
//	    before being re-archived from the live source (lossless backup)
//	  When --force is absent, BC-V3R3-007 idempotency is preserved: archive
//	  drift surfaces as an ARCHIVE_DRIFT error rather than silent overwrite.
//	--shell-env: Configure shell environment variables
//	--yes: Auto-confirm all prompts (CI/CD mode)
//	--templates-only: Skip binary update, sync templates only
//	--binary: Update binary only, skip template sync
func runUpdate(cmd *cobra.Command, _ []string) error {
	checkOnly := getBoolFlag(cmd, "check")
	shellEnv := getBoolFlag(cmd, "shell-env")
	editConfig := getBoolFlag(cmd, "config")
	binaryOnly := getBoolFlag(cmd, "binary")
	templatesOnly := getBoolFlag(cmd, "templates-only")
	out := cmd.OutOrStdout()
	th := resolveTheme()

	// SPEC-V3R6-UPDATE-NOISE-001 REQ-UN-010: propagate --verbose through the
	// package-level updateVerboseMode flag so recordMergeFallback can bypass
	// 3-strike suppression. `moai update` is single-process sequential — no
	// synchronization needed. The flag is reset on function exit so subsequent
	// in-process invocations (CLI tests, helpers) start with the default
	// suppression-on behavior.
	updateVerboseMode = getBoolFlag(cmd, "verbose")
	defer func() { updateVerboseMode = false }()

	// Validate mutually exclusive flags
	if binaryOnly && templatesOnly {
		return fmt.Errorf("--binary and --templates-only are mutually exclusive")
	}

	// Auto-prompt profile setup if no profile exists yet
	nonInteractive := getBoolFlag(cmd, "yes")
	if !nonInteractive && isatty.IsTerminal(os.Stdin.Fd()) {
		profileName := profile.GetCurrentName()
		if !profile.IsSetup(profileName) {
			var wantSetup bool
			confirm := huh.NewConfirm().
				Title("No profile found. Set up profile preferences now?").
				Description("Configure your name, language, and model preferences.").
				Value(&wantSetup)
			// Wrap the standalone confirm in a themed form: field.Run() cannot take
			// a theme, so the MoAI-branded dark-readable theme is applied at the
			// form level (parity with the wizard fix for the other huh surfaces).
			confirmForm := huh.NewForm(huh.NewGroup(confirm)).WithTheme(moaiHuhTheme())
			if err := confirmForm.Run(); err == nil && wantSetup {
				if err := runProfileSetup(cmd, nil); err != nil {
					_, _ = fmt.Fprintf(out, "Warning: profile setup failed: %v\n", err)
				}
			}
		}
	}

	// Handle --config / -c mode (edit configuration only, no template updates)
	// This takes priority over all other flags
	if editConfig {
		return runInitWizard(cmd, true) // true = reconfigure mode
	}

	currentVersion := version.GetVersion()
	_, _ = fmt.Fprintln(out, tui.KV("Current version", "moai-adk "+currentVersion, tui.KVOpts{Theme: &th, KeyWidth: 16}))

	// Handle shell-env mode
	if shellEnv {
		return runShellEnvConfig(cmd)
	}

	// Handle --check mode (informational: check if newer binary exists)
	if checkOnly {
		// Lazily initialize update dependencies
		if deps != nil {
			if err := deps.EnsureUpdate(); err != nil {
				deps.Logger.Debug("failed to initialize update system", "error", err)
			}
		}

		if deps == nil || deps.UpdateChecker == nil {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Update checker", "not available", "using current version", &th))
			return nil
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		info, err := deps.UpdateChecker.CheckLatest(ctx)
		if err != nil {
			return fmt.Errorf("check latest version: %w", err)
		}
		_, _ = fmt.Fprintln(out, tui.KV("Latest version", info.Version, tui.KVOpts{Theme: &th, KeyWidth: 16}))
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.CheckLine("info", "Note", "Binary updates happen automatically at session start", "", &th))
		return nil
	}

	// SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002 REQ-CRR-005 / AC-CRR-005: refuse to
	// operate in a non-project cwd (#1086). The positive marker is
	// `.moai/config/sections/system.yaml`; when it is absent this directory is
	// not a moai project and `moai update` MUST NOT write anything into it.
	//
	// Gating only the clean-reinstall branch is insufficient — that is why
	// #1086 survived M2. Option α reads a missing system.yaml as a POSITIVE
	// Signal 1, so the fingerprint returns IsV2=true in any empty directory;
	// even when the `&& isMoAIProject(cwd)` conjunct below correctly refuses
	// clean-reinstall, control falls through to the v3 file-level sync, which
	// deploys the full embedded template tree just the same. The refusal must
	// therefore precede BOTH paths.
	//
	// Placement is before acquireUpdateLock (and thus before
	// detectV2Fingerprint, per plan.md §F M2) because the lock itself is a
	// writer: it MkdirAll's `.moai/` to host `.moai/.update.lock`, and its
	// release removes only the lock file — leaving an empty `.moai/` behind in
	// the cwd. Gating after the lock still violates AC-CRR-005(a).
	//
	// The `!binaryOnly` conjunct keeps `moai update --binary` project-
	// independent: it upgrades the moai binary itself and performs no template
	// sync, so it has no project-marker precondition. `--check` returns earlier
	// still and is likewise unaffected.
	//
	// The error names the missing marker relative to the cwd and directs the
	// user to `moai init`. It deliberately does NOT echo the absolute cwd
	// (acceptance.md §D.7 Secured: the structured error must not leak absolute
	// paths or environment details).
	// SPEC-UPDATE-DATA-SURVIVAL-001 REQ-UDS-021/022/025: --restore is the
	// lockout escape (plan.md §B.3). It is handled BEFORE the marker gate
	// below, because a destroyed system.yaml is exactly the damage the restore
	// entry point exists to repair — gating it would lock the user out of the
	// only command that can restore the marker. The exemption is scoped to this
	// branch alone; every other path still passes through the gate.
	if restoreDir := getStringFlag(cmd, "restore"); restoreDir != "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory for restore: %w", err)
		}
		return runUpdateRestore(cwd, restoreDir, out)
	}

	if !binaryOnly {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory for project-marker check: %w", err)
		}
		if err := checkProjectMarker(cwd); err != nil {
			return err
		}
	}

	// SPEC-CLIFIX-CRITICAL-001 REQ-CRIT-001-005: acquire update lock before any
	// destructive step so a concurrent moai update fails fast with a diagnostic
	// instead of interleaving destructive deploy/clean/restore operations.
	lockRoot, _ := findProjectRoot()
	if lockRoot == "" {
		lockRoot, _ = os.Getwd()
	}
	if lockRoot != "" {
		releaseLock, lockErr := acquireUpdateLock(lockRoot)
		if lockErr != nil {
			return lockErr
		}
		defer releaseLock()
	}

	// Step 1: Binary update (unless skipped)
	if !shouldSkipBinaryUpdate(cmd) {
		updated, err := runBinaryUpdateStep(cmd)
		if err != nil {
			// Binary update failure is never fatal; warn and continue
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Binary update", "check failed", err.Error(), &th))
		}
		if updated {
			if binaryOnly {
				// --binary mode: skip re-exec and template sync
				_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Binary updated (template sync skipped)", Theme: &th}))
				return nil
			}
			// New binary installed; re-exec so the latest templates are used
			if err := reexecNewBinary(); err != nil {
				_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Re-exec", "failed", err.Error(), &th))
				// Fall through to template sync with the current binary
			}
			// reexecNewBinary replaces the process on success, so we only
			// reach here if it failed.
		} else if binaryOnly {
			_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillInfo, Solid: false, Label: report.RenderOutcome(report.OutcomeAlreadyUpToDate, 0, ""), Theme: &th}))
			return nil
		}
	}

	// Step 2: Template sync (skipped when --binary is set)
	if binaryOnly {
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillNeutral, Solid: false, Label: "Skipped (dev build, --binary)", Theme: &th}))
		return nil
	}

	// --dry-run: print planned operations without mutating the filesystem
	if getBoolFlag(cmd, "dry-run") {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		// SPEC-WORKTREE-BRANCH-GUARD-001 (REQ-WBG-009): surface the worktree
		// advisory on the dry-run path too, so `moai update --dry-run` smoke
		// runs (AC-WBG-009) observe it without mutating the filesystem.
		emitWorktreeAdvisory(out, cwd)
		if archiveErr := dryRunArchiveLegacySkills(cwd, out); archiveErr != nil {
			return archiveErr
		}
		// SPEC-UPDATE-REINSTALL-LOOP-002 REQ-RIL2-024/025 (M4): the v2
		// fingerprint is computed HERE — inside the dry-run branch, above the
		// deny-rule migration below — so the clean-reinstall and
		// residue-cleanup plans become reachable from `moai update --dry-run`.
		// Before M4 this branch returned immediately after the archive
		// summary, leaving both plan renderers dead from the CLI even though
		// each already existed and update.go already wired `DryRun:` into
		// runCleanReinstall.
		//
		// The early return itself does NOT move (REQ-RIL2-026): it stays
		// ABOVE stripRetiredV2DenyEntries, which rewrites settings.json.
		return emitDryRunReinstallPlan(cmd.Context(), cwd, getBoolFlag(cmd, "force"), out, th)
	}

	// Retired-deny-rule migration on the v3 path (issue #1101 follow-up). The
	// same one-shot strip runs inside runCleanReinstall, but that path opens
	// only on a v2 fingerprint — a project already on v3 keeps the retired
	// Write/Grep/Glob entries forever, because every v3 update merge-preserves
	// the user's settings.json. Running it here covers the v3-to-v3 case.
	//
	// Placement is deliberate: after the --binary / --dry-run early-returns (so
	// a dry run never mutates) but BEFORE the version-match short-circuit
	// further below, since a same-version `moai update` is exactly the case
	// that leaves the stale entries in place. stripRetiredV2DenyEntries is a
	// no-op when nothing matches, so a clean v3 file is never rewritten and the
	// second call inside runCleanReinstall stays idempotent.
	{
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory for deny-rule migration: %w", err)
		}
		if stripErr := stripRetiredV2DenyEntries(cwd, out); stripErr != nil {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Deny-rule migration", "failed", stripErr.Error(), &th))
		}
	}

	// SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 REQ-VVCR-002: detect v2 fingerprint
	// and short-circuit to the clean-reinstall code path when the project is
	// v2 (or partial-v2). The detector inspects three signals (system.yaml
	// moai.version, .agency/ presence, DeprecatedPaths enumeration); ANY
	// positive signal triggers runCleanReinstall instead of the v3 file-level
	// sync below. On IsV2: false this is a no-op and the v3 sync proceeds.
	{
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory for v2 detection: %w", err)
		}
		// REQ-CRR-005 / AC-CRR-004: clean-reinstall requires BOTH a v2
		// fingerprint AND a genuine moai-project cwd (positive marker
		// `.moai/config/sections/system.yaml`). A non-project cwd MUST NOT
		// trigger clean-reinstall even if legacy residue drives IsV2=true
		// (#1086 regression repair, M2).
		fingerprint, fpErr := detectV2Fingerprint(cwd)
		if fpErr != nil {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "v2 detection", "failed", fpErr.Error(), &th))
		} else if fingerprint.IsV2 && isMoAIProject(cwd) {
			_, _ = fmt.Fprintln(out, tui.CheckLine("info", "v2 detected",
				"running clean reinstall",
				fmt.Sprintf("signals: version=%v agency=%v deprecated=%v",
					fingerprint.V2DetectedViaVersion,
					fingerprint.V2DetectedViaAgencyDir,
					fingerprint.V2DetectedViaDeprecatedPath),
				&th))

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()

			result, runErr := runCleanReinstall(ctx, cwd, CleanReinstallOptions{
				DryRun: getBoolFlag(cmd, "dry-run"),
				// Propagate the CLI --force flag to the Step 3.5 migration
				// adapter (issue #1132): explicit `moai update --force`
				// performs a real forced migration instead of the
				// already-migrated skip.
				Force: getBoolFlag(cmd, "force"),
				Out:   out,
				// Inject the canonical .agency/ migration adapter. When the
				// project carries a .agency/ legacy directory, this is invoked
				// in Step 3.5 of the canonical order (REQ-VVCR-025), mirroring
				// the migrateLegacyMemoryDir auto-invoke precedent at line 1731.
				RunMigrateAgency: runAgencyMigrationAdapter,
			})
			if runErr != nil {
				return fmt.Errorf("v2-to-v3 clean reinstall: %w", runErr)
			}

			// On successful clean reinstall, the v3 file-level sync below is
			// redundant — runCleanReinstall already invoked deployer.Deploy.
			// Return early so subsequent steps (archive legacy skills, design
			// dir sync, profile sync) skip the v3 file-level pathway.
			_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{
				Kind:  tui.PillOk,
				Solid: false,
				Label: fmt.Sprintf("Clean reinstall complete (%d files preserved, %d deprecated removed)",
					len(result.Inventory.Files), len(result.RemovedPaths)),
				Theme: &th,
			}))
			return nil
		}

		// SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002 REQ-CRR-007 / AC-CRR-007:
		// Fire .agency/ migration INDEPENDENTLY of the v2 fingerprint verdict.
		// A v3 project (IsV2=false per REQ-CRR-001's v3-version negative-
		// override) that still carries a lingering .agency/ directory needs the
		// migration to run, but the Step 3.5 migration inside runCleanReinstall
		// is gated on fp.V2DetectedViaAgencyDir && IsV2 — which never opens for
		// a v3 project. This pre-step closes that gap.
		//
		// Placement rationale: this is reachable ONLY when the clean-reinstall
		// early-return above did NOT fire (detection succeeded, IsV2=false, or
		// IsV2=true but non-moai-project cwd). The gate below narrows to the
		// AC-CRR-007 contract: v3 project (IsV2=false) + genuine moai project
		// + .agency/ present. Genuine-v2 projects (IsV2=true) are handled by
		// the clean-reinstall path above and never reach here, so there is no
		// double-fire (the adapter need not swallow ErrMigrateArchiveExists).
		//
		// SPEC-UPDATE-REINSTALL-LOOP-002 M3 (REQ-RIL2-019..023): this same block
		// is the "v3 project with residue" branch, so the deprecated-path sweep
		// is EXTENDED onto it rather than added elsewhere. runV3ResidueCleanup
		// keeps the agency migration pre-step above as its own step — it is not
		// gated behind the sweep's outcome — and adds backup-then-remove for the
		// deprecated residue. It performs no PRESERVE snapshot, no tree wipe and
		// no forced redeployment (REQ-RIL2-020); the destructive full reinstall
		// stays reachable only through the IsV2 path above.
		if fpErr == nil && !fingerprint.IsV2 && isMoAIProject(cwd) {
			if _, cleanupErr := runV3ResidueCleanup(cwd, getBoolFlag(cmd, "dry-run"), getBoolFlag(cmd, "force"), out); cleanupErr != nil {
				return cleanupErr
			}
		}
	}

	syncSkipped, err := runTemplateSyncWithProgress(cmd)
	if err != nil {
		return err
	}

	// SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 REQ-UAC-004: when the template sync
	// branch short-circuits (version match + !forceUpdate, or user cancelled
	// merge), the legacy-skill archive check MUST also be short-circuited.
	// Pre-fix UX leaked a "Skipping sync" line immediately followed by
	// "Legacy skill archive failed" because the archive ran unconditionally.
	if syncSkipped {
		// SPEC-MODEL-PROFILE-MATRIX-001 (REQ-MPM-016): an explicit --profile override
		// must still persist to llm.profile even when the template sync short-circuits.
		if p := getStringFlag(cmd, "profile"); p != "" {
			if err := applyUpdateProfile(".", p); err != nil {
				return err
			}
		}
		return nil
	}

	// Post-sync follow-up: the "Updated N files" summary has already printed
	// (inside runTemplateSyncWithProgress). The steps below — legacy-skill
	// archive, evolution dir scaffold, profile sync — are a DISTINCT follow-up
	// phase. A section header separates them from the deploy summary so the
	// archive output does not look like it appended to "Updated N files".
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, tui.Section("Post-sync steps", tui.SectionOpts{Theme: &th}))

	// Archive legacy skills (BC-V3R3-007): move 16 removed static skills to
	// .moai/archive/skills/v2.16/ before they are cleaned from .claude/skills/.
	// SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 REQ-UAC-002: --force is propagated
	// so that drift-detection routes through the overwrite + backup path
	// instead of returning ARCHIVE_DRIFT.
	{
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory for archive: %w", err)
		}
		if _, archiveErr := archiveLegacySkills(cwd, out, getBoolFlag(cmd, "force")); archiveErr != nil {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Legacy skill archive", "failed", archiveErr.Error(), &th))
		}
	}

	// Ensure .moai/evolution/ directory tree exists for existing projects
	// that predate the evolution infrastructure (R2: Directory Scaffolding).
	if err := deploy.ScaffoldEvolutionDir("."); err != nil {
		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Evolution dir", "scaffold failed", err.Error(), &th))
	}

	// Sync profile preferences to project config (after template deployment)
	profileName := profile.GetCurrentName()
	prefs, err := profile.ReadPreferences(profileName)
	if err != nil {
		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Profile preferences", "read failed", err.Error(), &th))
	} else {
		if err := profile.SyncToProjectConfig(".", prefs); err != nil {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Profile sync", "failed", err.Error(), &th))
		}
	}

	// SPEC-MODEL-PROFILE-MATRIX-001 (REQ-MPM-016): when --profile is given,
	// persist the override to llm.profile (no agent frontmatter mutation).
	if p := getStringFlag(cmd, "profile"); p != "" {
		if err := applyUpdateProfile(".", p); err != nil {
			return err
		}
	}

	return nil
}

// emitDryRunReinstallPlan renders, for `moai update --dry-run`, either the
// v2-to-v3 clean-reinstall plan (REQ-RIL2-024) or the v3 residue-cleanup plan
// (REQ-RIL2-025) — whichever the project's v2 fingerprint selects. It is the
// dry-run counterpart of the v2-detection block in runUpdate.
//
// It writes nothing (REQ-RIL2-026). detectV2Fingerprint and isMoAIProject only
// read; runCleanReinstall returns from its own dry-run branch before Step 3's
// backup; runV3ResidueCleanup returns from its dryRun branch before its
// backup-then-remove step. Both renderers already existed — M4 makes them
// reachable rather than adding a parallel implementation.
//
// A detection failure degrades to a warning and no plan, matching the
// non-dry-run block: a dry run must never fail the command.
func emitDryRunReinstallPlan(ctx context.Context, cwd string, force bool, out io.Writer, th tui.Theme) error {
	if !isMoAIProject(cwd) {
		return nil
	}
	fingerprint, fpErr := detectV2Fingerprint(cwd)
	if fpErr != nil {
		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "v2 detection", "failed", fpErr.Error(), &th))
		return nil
	}

	if !fingerprint.IsV2 {
		// v3-confirmed project: the residue sweep is the applicable plan.
		if _, cleanupErr := runV3ResidueCleanup(cwd, true, force, out); cleanupErr != nil {
			return cleanupErr
		}
		return nil
	}

	_, _ = fmt.Fprintln(out, tui.CheckLine("info", "v2 detected",
		"clean reinstall planned",
		fmt.Sprintf("signals: version=%v agency=%v deprecated=%v",
			fingerprint.V2DetectedViaVersion,
			fingerprint.V2DetectedViaAgencyDir,
			fingerprint.V2DetectedViaDeprecatedPath),
		&th))

	if ctx == nil {
		ctx = context.Background()
	}
	planCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if _, runErr := runCleanReinstall(planCtx, cwd, CleanReinstallOptions{
		DryRun:           true,
		Force:            force,
		Out:              out,
		RunMigrateAgency: runAgencyMigrationAdapter,
	}); runErr != nil {
		return fmt.Errorf("dry-run clean reinstall plan: %w", runErr)
	}
	return nil
}

// applyUpdateProfile persists a --profile override to llm.profile during
// `moai update` (SPEC-MODEL-PROFILE-MATRIX-001 REQ-MPM-016/024). The former
// plan_type × tier agent-frontmatter re-mutation (ApplyTierProfile) is RETIRED —
// this path writes to llm.yaml only, leaving agent frontmatter at model: inherit.
// An out-of-set profileFlag returns an error naming the closed set (defensive —
// the CLI flag is validated by validateUpdateFlags before this is reached).
func applyUpdateProfile(projectRoot, profileFlag string) error {
	if profileFlag == "" {
		return nil
	}
	if !config.IsValidProfile(profileFlag) {
		return fmt.Errorf("invalid --profile value %q: must be one of: max, medium, low", profileFlag)
	}
	if err := template.ApplyProfile(projectRoot, profileFlag); err != nil {
		return fmt.Errorf("persist profile: %w", err)
	}
	// Keep the legacy performance_tier alias in sync so the separate Tier x Phase
	// axis reads a consistent tier.
	if err := template.ApplyPerformanceTier(projectRoot, profileFlag); err != nil {
		return fmt.Errorf("persist performance_tier: %w", err)
	}
	return nil
}

// shouldSkipBinaryUpdate returns true when the binary update step should
// be skipped. This happens in three cases:
//  1. The --templates-only flag is set (update command only).
//  2. The MOAI_SKIP_BINARY_UPDATE=1 environment variable is set (used by
//     reexecNewBinary to prevent infinite re-exec loops).
//  3. The current binary is a dev build (version contains "dirty", "dev",
//     or "none"), where self-update is meaningless.
func shouldSkipBinaryUpdate(cmd *cobra.Command) bool {
	// Flag check (only the update command registers this flag)
	if f := cmd.Flags().Lookup("templates-only"); f != nil && f.Value.String() == "true" {
		return true
	}

	// Environment variable guard (set by reexecNewBinary)
	if os.Getenv(config.EnvSkipBinaryUpdate) == "1" {
		return true
	}

	// Dev build detection (reuse pattern from buildAutoUpdateFunc in deps.go)
	v := version.GetVersion()
	if strings.Contains(v, "dirty") || v == "dev" || strings.Contains(v, "none") {
		return true
	}

	return false
}

// @MX:NOTE: [AUTO] runBinaryUpdateStep — M4-S4d-1 DDD migration. New-version notice uses
// two tui.KV lines (New / Current), progress is tui.CheckLine "run", and the result is a tui.Pill PillOk.
//
// runBinaryUpdateStep checks whether a newer moai binary is available and,
// if so, downloads and installs it. The caller should re-exec the process
// when updated is true.
//
// Errors are non-fatal by design: the caller should log the error and
// continue with the original operation (template sync or init).
func runBinaryUpdateStep(cmd *cobra.Command) (updated bool, err error) {
	out := cmd.OutOrStdout()
	th := resolveTheme()

	// Lazily initialise update dependencies
	if deps != nil {
		if initErr := deps.EnsureUpdate(); initErr != nil {
			return false, fmt.Errorf("initialize update system: %w", initErr)
		}
	}

	if deps == nil || deps.UpdateChecker == nil {
		return false, nil
	}

	currentVersion := version.GetVersion()

	// Check for available update
	available, info, err := deps.UpdateChecker.IsUpdateAvailable(currentVersion)
	if err != nil {
		return false, fmt.Errorf("check for update: %w", err)
	}
	if !available {
		return false, nil
	}

	_, _ = fmt.Fprintln(out, tui.KV("New version", info.Version, tui.KVOpts{Theme: &th, KeyWidth: 16}))
	_, _ = fmt.Fprintln(out, tui.KV("Current version", currentVersion, tui.KVOpts{Theme: &th, KeyWidth: 16}))
	_, _ = fmt.Fprintln(out, tui.CheckLine("run", "Installing update", "", "", &th))

	if deps.UpdateOrch == nil {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 120*time.Second)
	defer cancel()

	result, err := deps.UpdateOrch.Update(ctx)
	if err != nil {
		return false, fmt.Errorf("install update: %w", err)
	}

	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Updated " + result.PreviousVersion + " → " + result.NewVersion, Theme: &th}))
	return true, nil
}

// reexecNewBinary replaces the current process with the newly installed
// moai binary, preserving the original command-line arguments. It sets
// MOAI_SKIP_BINARY_UPDATE=1 to prevent the re-execed process from
// attempting another binary update.
//
// On Unix this uses syscall.Exec (the process is replaced in-place).
// On Windows syscall.Exec is not available, so we spawn a child process
// and exit the parent.
func reexecNewBinary() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	// Prevent re-exec loop
	if err := os.Setenv("MOAI_SKIP_BINARY_UPDATE", "1"); err != nil {
		return fmt.Errorf("set MOAI_SKIP_BINARY_UPDATE: %w", err)
	}

	if runtime.GOOS == "windows" {
		// Windows: spawn child and exit parent
		child := exec.Command(exe, os.Args[1:]...)
		child.Stdin = os.Stdin
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		if err := child.Run(); err != nil {
			return fmt.Errorf("re-exec on windows: %w", err)
		}
		os.Exit(0)
	}

	// Unix: replace process via execve(2)
	return syscall.Exec(exe, os.Args, os.Environ())
}

// @MX:NOTE: [AUTO] runShellEnvConfig — M4-S4d-2 DDD migration. tui.Section header,
// tui.KV (Shell/Config/Explanation), tui.CheckLine (Changes list), tui.Pill (final outcome).
//
// runShellEnvConfig configures shell environment variables for Claude Code.
func runShellEnvConfig(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	th := resolveTheme()

	_, _ = fmt.Fprintln(out, tui.Section("Configuring shell environment for Claude Code", tui.SectionOpts{Theme: &th}))

	// Get recommendation first
	configurator := shell.NewEnvConfigurator(nil)
	rec := configurator.GetRecommendation()

	_, _ = fmt.Fprintln(out, tui.KV("Shell", string(rec.Shell), tui.KVOpts{Theme: &th, KeyWidth: 12}))
	_, _ = fmt.Fprintln(out, tui.KV("Config file", rec.ConfigFile, tui.KVOpts{Theme: &th, KeyWidth: 12}))
	_, _ = fmt.Fprintln(out, tui.KV("Explanation", rec.Explanation, tui.KVOpts{Theme: &th, KeyWidth: 12}))
	_, _ = fmt.Fprintln(out, tui.Section("Changes to add", tui.SectionOpts{Theme: &th}))
	for _, change := range rec.Changes {
		_, _ = fmt.Fprintln(out, tui.CheckLine("info", "·", change, "", &th))
	}
	_, _ = fmt.Fprintln(out)

	// Execute configuration
	result, err := configurator.Configure(shell.ConfigOptions{
		AddClaudeWarningDisable: true,
		AddLocalBinPath:         true,
		AddGoBinPath:            true,
		PreferLoginShell:        true,
	})
	if err != nil {
		return fmt.Errorf("configure shell environment: %w", err)
	}

	if result.Skipped {
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillInfo, Solid: false, Label: "Shell environment already configured in " + result.ConfigFile, Theme: &th}))
	} else {
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Shell environment configured in " + result.ConfigFile, Theme: &th}))
		_, _ = fmt.Fprintln(out, "Please restart your terminal or run:")
		_, _ = fmt.Fprintln(out, tui.KV("source", result.ConfigFile, tui.KVOpts{Theme: &th, KeyWidth: 8}))
	}

	return nil
}

// runAgencyMigrationAdapter is the auto-invoke entrypoint for the
// .agency/ → .moai/ migration triggered by runCleanReinstall Step 3.5
// (REQ-VVCR-025 of SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001) and by the
// runUpdate v3-project pre-step (REQ-CRR-007).
//
// Unlike runMigrateAgency (which is cobra-driven and reads --dry-run /
// --force / --resume from CLI flags), this adapter constructs the runner
// directly with the values supplied by the clean-reinstall orchestrator.
// It mirrors the auto-invoke precedent of migrateLegacyMemoryDir.
//
// force propagates the CLI --force flag (issue #1132): an explicit
// `moai update --force` performs a real forced migration (overwriting
// existing targets), matching the `moai migrate agency --force` contract.
//
// Swallowed error codes (graceful no-ops):
//   - ErrMigrateNoSource: .agency/ disappeared between detection and
//     adapter invocation (race-safe no-op).
//   - ErrMigrateTargetExists / ErrMigrateArchiveExists (issue #1132):
//     a migration target (design.yaml / research/observations) or the
//     .agency.archived/ backup already exists — the project is already
//     migrated (typically by an earlier interrupted update whose v3
//     template deploy created the targets). Pre-fix, this error aborted
//     clean-reinstall Step 3.5 BEFORE Step 4 removed .agency/, so the v2
//     fingerprint (Signal 2) never converged and every `moai update`
//     re-entered the same abort — a permanent retry loop the CLI --force
//     flag never reached. Treating it as already-migrated lets Step 4
//     clear the residue and the fingerprint converge.
func runAgencyMigrationAdapter(projectRoot string, dryRun, force bool, out io.Writer) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("agency migration adapter: home dir: %w", err)
	}

	r := &migrateAgencyRunner{
		projectRoot: projectRoot,
		homeDir:     homeDir,
		dryRun:      dryRun,
		force:       force,
	}

	if _, runErr := r.Run(); runErr != nil {
		var me *MigrateError
		if errors.As(runErr, &me) {
			switch me.Code {
			case ErrMigrateNoSource:
				return nil
			case ErrMigrateTargetExists, ErrMigrateArchiveExists:
				_, _ = fmt.Fprintf(out,
					"[clean-reinstall] agency migration skipped: target already migrated (%s)\n", me.Code)
				return nil
			}
		}
		return fmt.Errorf("agency migration: %w", runErr)
	}
	return nil
}

// allStatuslineSegments removed (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001): the
// presetToSegments wrapper that consumed it was deleted, leaving this alias
// unused. The canonical segment SSOT remains statusline.CanonicalSegments.

// updateSettingsLocalEnv updates a single environment variable in
// settings.local.json. If the file doesn't exist, it creates a new one. If the
// env map doesn't exist, it creates it.
//
// SPEC-CLIFIX-CRITICAL-001 REQ-CRIT-001-001: round-trips as map[string]any so
// unknown top-level keys (hooks, outputStyle, model, defaultMode, and any future
// keys) survive the write. The closed settingsLocalEnv struct this function
// previously marshaled is removed — marshaling it emitted only the env key and
// silently wiped every other top-level key.
// SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-001: route through the locked+atomic
// mutateSettingsLocal seam so concurrent sessions cannot lose updates.
func updateSettingsLocalEnv(settingsPath, key, value string) error {
	return mutateSettingsLocal(settingsPath, func(m map[string]any) {
		settingsEnvMap(m)[key] = value
	})
}

// globalMoaiHooksDir returns the global moai hooks directory under homeDir.
// It is the single named removal target for ensureGlobalSettingsEnv, so the
// deletion-radius guard has one symbol to assert against instead of
// re-deriving the path independently.
func globalMoaiHooksDir(homeDir string) string {
	return filepath.Join(homeDir, defs.ClaudeDir, "hooks", "moai")
}

// ensureGlobalSettingsEnv cleans up moai-managed settings from ~/.claude/settings.json.
// All settings (env, permissions, teammateMode, hooks) are managed at the project level.
// The global hooks directory (~/.claude/hooks/moai/) is also removed since hooks
// are only deployed to project-level directories via moai init.
func ensureGlobalSettingsEnv() error {
	homeDir, err := userHomeDirFn()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	// Remove global hooks/moai directory if it exists.
	// Hooks are project-level only; the global directory causes "No such file or directory"
	// errors in non-initialized projects that reference $CLAUDE_PROJECT_DIR paths.
	globalHooksDir := globalMoaiHooksDir(homeDir)
	if _, err := os.Stat(globalHooksDir); err == nil {
		_ = os.RemoveAll(globalHooksDir)
	}

	globalSettingsPath := filepath.Join(homeDir, defs.ClaudeDir, defs.SettingsJSON)

	// Read existing global settings
	var existingSettings map[string]any
	if data, err := os.ReadFile(globalSettingsPath); err == nil {
		if err := json.Unmarshal(data, &existingSettings); err != nil {
			return fmt.Errorf("parse existing global settings: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read global settings: %w", err)
	} else {
		// No global settings file, nothing to clean up
		return nil
	}

	needsUpdate := false

	// Clean up legacy hooks including orphaned scripts and deprecated Python hooks
	needsUpdate = cleanLegacyHooks(existingSettings) || needsUpdate

	// Clean up legacy moai-managed env keys that are no longer needed globally.
	// PATH is also removed here so it can be refreshed with the latest SmartPATH below.
	// Preserve any user-added custom env keys but remove moai-specific ones.
	if envVal, exists := existingSettings["env"]; exists {
		if envMap, ok := envVal.(map[string]any); ok {
			moaiKeys := []string{"PATH", "ENABLE_TOOL_SEARCH"}
			for _, key := range moaiKeys {
				if _, exists := envMap[key]; exists {
					delete(envMap, key)
					needsUpdate = true
				}
			}
			// If env is now empty, remove it entirely
			if len(envMap) == 0 {
				delete(existingSettings, "env")
			}
		}
	}

	// Ensure default global settings are present.
	// PATH: Provides a fallback SmartPATH for non-moai directories where no
	// project-level settings.json exists. Without this, Claude Code loses access
	// to basic tools (/usr/bin, /bin, etc.) in non-moai projects. Project-level
	// PATH overrides this when present. Refreshed on every moai update. (issue #598)
	// CLAUDE_DISABLE_PATH_WARNING: Suppresses Claude Code's PATH warning globally.
	// Previously only written to shell config (.zshenv/.profile), which may not be
	// sourced by non-login shells on WSL. (issue #598)
	// CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS: Enables Agent Teams mode by default.
	defaultEnvKeys := map[string]string{
		"PATH":                                 template.BuildSmartPATH(),
		"CLAUDE_DISABLE_PATH_WARNING":          "1",
		"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1",
	}
	for key, value := range defaultEnvKeys {
		if envVal, exists := existingSettings["env"]; exists {
			if envMap, ok := envVal.(map[string]any); ok {
				if _, exists := envMap[key]; !exists {
					envMap[key] = value
					needsUpdate = true
				}
			}
		} else {
			// No env section yet, create it with defaults
			existingSettings["env"] = map[string]any{
				key: value,
			}
			needsUpdate = true
		}
	}

	// Clean up moai-managed permissions if they only contain Task:*
	if permVal, exists := existingSettings["permissions"]; exists {
		if permMap, ok := permVal.(map[string]any); ok {
			if allowVal, exists := permMap["allow"]; exists {
				if allowArr, ok := allowVal.([]any); ok {
					if len(allowArr) == 1 && allowArr[0] == "Task:*" {
						delete(existingSettings, "permissions")
						needsUpdate = true
					}
				}
			}
		}
	}

	// Clean up moai-managed teammateMode
	if mode, exists := existingSettings["teammateMode"]; exists {
		if mode == "auto" {
			delete(existingSettings, "teammateMode")
			needsUpdate = true
		}
	}

	if !needsUpdate {
		return nil
	}

	// Write back
	jsonContent, err := json.MarshalIndent(existingSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal global settings: %w", err)
	}

	if err := os.WriteFile(globalSettingsPath, append(jsonContent, '\n'), defs.FilePerm); err != nil {
		return fmt.Errorf("write global settings: %w", err)
	}

	return nil
}

// cleanLegacyHooks removes legacy hook patterns from global settings.
// This includes orphaned scripts that were never deployed and deprecated Python-based hooks.
// Returns true if any cleanup was performed.
func cleanLegacyHooks(settings map[string]any) bool {
	// List of legacy hook patterns to remove.
	// All moai handle-*.sh hooks belong in project-level settings, not global.
	legacyPatterns := []string{
		"handle-session-end.sh",
		"handle-session-start.sh",
		"handle-stop.sh",
		"handle-pre-tool.sh",
		"handle-post-tool.sh",
		"handle-agent-hook.sh",
		"handle-compact.sh",
		"post_tool__code_formatter.py",
		"post_tool__linter.py",
		"post_tool__ast_grep_scan.py",
	}

	hooksMap, ok := settings["hooks"].(map[string]any)
	if !ok {
		return false
	}

	modified := false
	for hookType, hookListInterface := range hooksMap {
		hookList, ok := hookListInterface.([]any)
		if !ok {
			continue
		}

		var cleanedHooks []any
		for _, hookGroup := range hookList {
			groupMap, ok := hookGroup.(map[string]any)
			if !ok {
				cleanedHooks = append(cleanedHooks, hookGroup)
				continue
			}

			hooksList, ok := groupMap["hooks"].([]any)
			if !ok {
				cleanedHooks = append(cleanedHooks, hookGroup)
				continue
			}

			var cleanedGroupHooks []any
			for _, hookEntry := range hooksList {
				entryMap, ok := hookEntry.(map[string]any)
				if !ok {
					cleanedGroupHooks = append(cleanedGroupHooks, hookEntry)
					continue
				}

				command, ok := entryMap["command"].(string)
				if !ok {
					cleanedGroupHooks = append(cleanedGroupHooks, hookEntry)
					continue
				}

				// Check if command contains any legacy pattern
				shouldRemove := false
				for _, pattern := range legacyPatterns {
					if strings.Contains(command, pattern) {
						shouldRemove = true
						break
					}
				}

				if shouldRemove {
					modified = true
				} else {
					cleanedGroupHooks = append(cleanedGroupHooks, hookEntry)
				}
			}

			if len(cleanedGroupHooks) > 0 {
				groupMap["hooks"] = cleanedGroupHooks
				cleanedHooks = append(cleanedHooks, groupMap)
			} else {
				modified = true
			}
		}

		if len(cleanedHooks) > 0 {
			hooksMap[hookType] = cleanedHooks
		} else {
			delete(hooksMap, hookType)
			modified = true
		}
	}

	if modified && len(hooksMap) == 0 {
		delete(settings, "hooks")
	}

	return modified
}

// execCommand executes a command and returns its output.
func execCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// detectGoBinPathForUpdate detects the Go binary installation path for template rendering.
// Returns the path where Go binaries are installed (e.g., "/home/user/go/bin").
// REQ-V3R2-RT-007-001: deduplicated via the gobin.Detect helper.
func detectGoBinPathForUpdate(homeDir string) string {
	return gobin.Detect(homeDir)
}

// resolveMoaiExecutable returns the running binary's own resolved executable path
// via os.Executable(), or "" when it errors. The running binary IS the installed
// moai binary, so this resolves the installer location (e.g. on Windows) even when
// that directory is not on PATH. An empty result leaves the resolved-executable
// branch out of the rendered status_line.sh (graceful degradation).
func resolveMoaiExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return exe
}

// readHookOptInEnabled reads the SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 master
// toggle from .moai/config/sections/system.yaml at the given project root.
// Returns false on any error (file missing, parse error, key absent) — fail-CLOSED
// per R3 mitigation. Used by `moai update` to render settings.json conditionally.
func readHookOptInEnabled(projectRoot string) bool {
	configPath := filepath.Join(projectRoot, ".moai", "config", "sections", "system.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}
	var doc struct {
		Hook struct {
			OptIn struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"opt_in"`
		} `yaml:"hook"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false
	}
	return doc.Hook.OptIn.Enabled
}
