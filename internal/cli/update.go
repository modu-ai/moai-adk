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
	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	updatemerge "github.com/modu-ai/moai-adk/internal/cli/update/merge"
	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/cli/update/report"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/merge"
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
	Long:    "Check for binary updates, install if available, then synchronize embedded templates with the project.",
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
	if !binaryOnly {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory for project-marker check: %w", err)
		}
		if !isMoAIProject(cwd) {
			marker := filepath.ToSlash(filepath.Join(
				defs.MoAIDir, "config", "sections", "system.yaml"))
			return fmt.Errorf(
				"not a moai project: %s not found in the current directory\n\n"+
					"Run `moai init` to initialize a project here, or change to an "+
					"existing project directory and retry", marker)
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
		return dryRunArchiveLegacySkills(cwd, out)
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
		if fpErr == nil && !fingerprint.IsV2 && isMoAIProject(cwd) {
			if _, agencyStatErr := os.Stat(filepath.Join(cwd, ".agency")); agencyStatErr == nil {
				if migrateErr := runAgencyMigrationAdapter(cwd, getBoolFlag(cmd, "dry-run"), getBoolFlag(cmd, "force"), out); migrateErr != nil {
					return fmt.Errorf("pre-step agency migration: %w", migrateErr)
				}
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

// runTemplateSync synchronizes embedded templates with the project directory.
// It performs a quick version comparison first - if the project's template version
// matches the package version, the sync is skipped for performance (70-80% faster).
//
// Template deployment uses a 3-way merge strategy to preserve local modifications.
// Users are prompted to confirm the merge before proceeding.
func runTemplateSync(cmd *cobra.Command) error {
	return runTemplateSyncWithReporter(cmd, nil, false)
}

// @MX:NOTE: [AUTO] runTemplateSyncWithReporter — M4-S4d-2 DDD migration. Top-level header/section/
// final-outcome lines are converted to tui.KV / tui.Section / tui.CheckLine / tui.Pill. Sub-step
// micro messages (\r-prefixed sym* helpers) are preserved because they drive the progress display.
// Design source: screens.jsx ScreenUpdate.
//
// runTemplateSyncWithReporter synchronizes templates with progress reporting.
func runTemplateSyncWithReporter(cmd *cobra.Command, reporter project.ProgressReporter, skipConfirm bool) error {
	out := cmd.OutOrStdout()
	th := resolveTheme()
	ctx := cmd.Context()

	// Get flags for template sync
	forceBackup := getBoolFlag(cmd, "force")
	autoConfirm := getBoolFlag(cmd, "yes")

	// Use current directory as project root
	projectRoot := "."

	currentVersion := version.GetVersion()
	// Identity header band (REQ-TUXIU-015): "◆ MoAI-ADK <version> <go-runtime>
	// · claude" with the version as a solid brand pill.
	_, _ = fmt.Fprintln(out, renderIdentityBand(currentVersion, th))
	_, _ = fmt.Fprintln(out, tui.KV("Current version", "moai-adk "+currentVersion, tui.KVOpts{Theme: &th, KeyWidth: 16}))
	_, _ = fmt.Fprintln(out, tui.CheckLine("run", "Syncing templates", "from embedded filesystem", "", &th))

	if reporter != nil {
		reporter.StepStart("Version Check", "Checking template version...")
	}

	// Stage 2: Config Version Comparison (before template sync)
	// Compare package template_version with project config template_version
	// If versions match, skip sync for performance (70-80% faster)
	packageVersion := version.GetVersion()
	projectVersion, err := plan.GetProjectConfigVersion(projectRoot)
	if err == nil && packageVersion == projectVersion && !forceBackup {
		if reporter != nil {
			reporter.StepComplete("Already up-to-date")
		}
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: report.RenderOutcome(report.OutcomeAlreadyUpToDate, 0, ""), Theme: &th}))
		return nil
	}

	if reporter != nil {
		reporter.StepComplete("Version check complete")
	}

	if reporter != nil {
		reporter.StepStart("Loading Templates", "Reading embedded templates...")
	}

	// Load embedded templates
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		if reporter != nil {
			reporter.StepError(err)
		}
		return fmt.Errorf("load embedded templates: %w", err)
	}

	if reporter != nil {
		reporter.StepComplete("Templates loaded")
	}

	if reporter != nil {
		reporter.StepStart("Loading Manifest", "Reading project manifest...")
	}

	// Initialize manifest manager
	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		if reporter != nil {
			reporter.StepError(err)
		}
		return fmt.Errorf("load manifest: %w", err)
	}

	if reporter != nil {
		reporter.StepComplete("Manifest loaded")
	}

	// Create renderer for template variable substitution
	renderer := template.NewRenderer(embedded)

	// Create deployer with renderer and force update enabled for template sync
	// This ensures template files are rendered (.tmpl -> actual file) and updated even if they exist
	deployer := template.NewDeployerWithRendererAndForceUpdate(embedded, renderer, true)

	// Analyze merge and get user confirmation
	analysis := updatemerge.AnalyzeMergeChanges(deployer, projectRoot)

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, tui.Section("Analyzing merge changes", tui.SectionOpts{Theme: &th}))
	// Card-style classification summary (REQ-TUXIU-010/011): accent box with
	// up to three count pills; zero-count pills omitted; suppressed entirely
	// when the run is clean (all counts zero).
	addCount, updateCount, conflictCount := classifyUpdateCounts(analysis.Files)
	if card := renderClassificationSummary(addCount, updateCount, conflictCount, th); card != "" {
		_, _ = fmt.Fprintln(out, card)
	}

	if reporter != nil {
		reporter.StepUpdate("Found " + fmt.Sprintf("%d files to sync", len(analysis.Files)))
	}

	// Skip confirmation if --yes flag is provided (CI/CD mode) or pre-confirmed
	var proceed bool
	if skipConfirm {
		proceed = true
	} else if autoConfirm {
		proceed = true
		_, _ = fmt.Fprintln(out, tui.CheckLine("info", "Auto-confirm", "CI/CD mode", "", &th))
	} else {
		var err error
		proceed, err = confirmViaPreview(analysis, projectRoot)
		if err != nil {
			if reporter != nil {
				reporter.StepError(err)
			}
			return fmt.Errorf("confirm merge for %d files (risk: %s): %w",
				len(analysis.Files), analysis.RiskLevel, err)
		}
	}

	if !proceed {
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillNeutral, Solid: false, Label: "Merge cancelled by user", Theme: &th}))
		if reporter != nil {
			reporter.StepError(errors.New("cancelled by user"))
		}
		return nil
	}

	// Deploy templates
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, tui.Section("Proceeding with template deployment", tui.SectionOpts{Theme: &th}))
	_, _ = fmt.Fprintln(out)

	// Define deployment steps
	steps := []struct {
		name    string
		message string
		execute func() error
	}{
		{
			name:    "Backup",
			message: "Backing up configuration",
			execute: func() error {
				// Always backup before update (even with --force)
				// --force only skips version check, not backup/merge
				// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces
				// the legacy CR-plus-format pair (REQ-UPR-004).
				pl := tui.ProgressLine(out, "Backing up .moai/config...", nil)
				configBackupPath, backupErr := backup.BackupMoaiConfig(projectRoot)
				if backupErr != nil {
					pl.Fail(fmt.Sprintf("Backup failed: %v", backupErr))
					return backupErr
				}
				if configBackupPath != "" {
					pl.Done(".moai/config backed up")
				} else {
					pl.Done("No config to backup")
				}
				return nil
			},
		},
		{
			name:    "Validate Templates",
			message: "Validating all templates before deployment",
			execute: func() error {
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

				// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces
				// the legacy CR-plus-format pair (REQ-UPR-004).
				pl := tui.ProgressLine(out, "Validating templates...", nil)
				if validateErr := deployer.ValidateAll(ctx, tmplCtx); validateErr != nil {
					pl.Fail(fmt.Sprintf("Template validation failed: %v", validateErr))
					return fmt.Errorf("template validation: %w", validateErr)
				}
				pl.Done("All templates validated")
				return nil
			},
		},
		{
			name:    "Clean Managed Paths",
			message: "Removing old MoAI-managed files",
			execute: func() error {
				return deploy.CleanMoaiManagedPaths(projectRoot, out)
			},
		},
		{
			name:    "Deploy Templates",
			message: "Deploying template files",
			execute: func() error {
				// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces
				// the legacy CR-plus-format pair (REQ-UPR-004).
				pl := tui.ProgressLine(out, "Deploying templates...", nil)

				// Build TemplateContext with detected paths for template rendering
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
					pl.Fail(fmt.Sprintf("Deployment failed: %v", deployErr))
					return fmt.Errorf("deploy templates: %w", deployErr)
				}
				pl.Done("Templates deployed")
				return nil
			},
		},
		{
			name:    "Restore Settings",
			message: "Restoring user settings",
			execute: func() error {
				// This step's status is tracked via configBackupPath variable
				// We'll handle this in the main flow
				return nil
			},
		},
	}

	// Track config backup path for restore step
	var configBackupPath string
	// Backup of user's .gitignore content for EntryMerge after deploy
	var gitignoreBackup []byte
	// Backups of mergeable files for 3-way merge after deploy
	var mergeableBackups []updatemerge.FileBackup

	// collectMergeableFiles returns a list of files that should be merged
	// using the 3-way merge engine during update.
	// Note: .moai/config/sections/*.yaml files are already handled by
	// restoreMoaiConfig with 3-way merge, so they are excluded here.
	collectMergeableFiles := func(projectRoot string) []string {
		// Fixed mergeable files at project root that are NOT handled by restoreMoaiConfig.
		// .mcp.json is intentionally absent: MoAI no longer ships an MCP template
		// (full MCP removal), so a user's .mcp.json is not a merge target.
		return []string{
			".claude/settings.json",
			".moai/status_line.sh",
		}
	}

	// Execute each step with progress reporting. The per-step in-flight line is
	// rendered by the inline tui.ProgressLine primitive inside each step body;
	// the coarse reporter Step wrapper is intentionally NOT driven here — doing
	// so double-rendered every step on the stderr channel (the printer.Step
	// in-flight line) alongside the stdout ProgressLine, producing the two-part
	// "○…○" spinner residue on a TTY (REQ-TUXIU-020/021). Errors still surface
	// via reporter.StepError (orphan-safe) below.
	for i, step := range steps {
		// Special handling for backup/restore steps; default executes normally
		switch step.name {
		case "Backup":
			// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces
			// the legacy CR-plus-format pair (REQ-UPR-004).
			plBackup := tui.ProgressLine(out, "Backing up .moai/config...", nil)
			var backupErr error
			configBackupPath, backupErr = backup.BackupMoaiConfig(projectRoot)
			if backupErr != nil {
				plBackup.Fail(fmt.Sprintf("Backup failed: %v", backupErr))
				if reporter != nil {
					reporter.StepError(backupErr)
				}
				return backupErr
			}
			if configBackupPath != "" {
				plBackup.Done(".moai/config backed up")
			} else {
				plBackup.Done("No config to backup")
			}

			// SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 M3: user-owned namespace backup
			// (REQ-UNP-004). Sequential after .moai/config backup. Skips silently
			// when no user-owned content exists (EC-UNP-001).
			plNsBackup := tui.ProgressLine(out, "Backing up user-owned namespace...", nil)
			nsBackupPath, nsBackupErr := backupUserOwnedNamespace(projectRoot)
			if nsBackupErr != nil {
				plNsBackup.Fail(fmt.Sprintf("Namespace backup failed: %v", nsBackupErr))
				if reporter != nil {
					reporter.StepError(nsBackupErr)
				}
				return nsBackupErr
			}
			if nsBackupPath != "" {
				// Derive display-friendly project-root-relative path
				displayNs := nsBackupPath
				if rel, relErr := filepath.Rel(projectRoot, nsBackupPath); relErr == nil {
					displayNs = rel
				}
				plNsBackup.Done(fmt.Sprintf("User-owned namespace backed up: %s", displayNs))
			} else {
				plNsBackup.Done("No user-owned namespace to back up")
			}

			// SPEC-INTERNAL-SECURITY-001 REQ-SEC-006 (AC-SEC-006a): wire the
			// pre-modification abort sentinel as a REAL gate on the moai update
			// deploy path. verifyNamespaceBackupCoverage runs AFTER the backup
			// completes and BEFORE any destructive deploy step; it aborts with
			// UPDATE_USER_NAMESPACE_VIOLATION when a user-owned namespace file
			// on disk would be overwritten without a backup. Normal updates
			// (no user-owned content, or fully backed up) pass through
			// unchanged (NFR-SEC-003).
			if covErr := verifyNamespaceBackupCoverage(projectRoot, nsBackupPath); covErr != nil {
				plNsBackup.Fail(fmt.Sprintf("Namespace safety check failed: %v", covErr))
				if reporter != nil {
					reporter.StepError(covErr)
				}
				return covErr
			}

			// Also backup .gitignore for EntryMerge after deploy
			gitignorePath := filepath.Join(projectRoot, ".gitignore")
			if data, readErr := os.ReadFile(gitignorePath); readErr == nil {
				gitignoreBackup = data
			}
			// Backup mergeable files for 3-way merge after deploy
			mergeableFiles := collectMergeableFiles(projectRoot)
			for _, mf := range mergeableFiles {
				mfPath := filepath.Join(projectRoot, mf)
				if data, readErr := os.ReadFile(mfPath); readErr == nil {
					mergeableBackups = append(mergeableBackups, updatemerge.FileBackup{Path: mf, Data: data})
				}
			}
		case "Restore Settings":
			// Handle restore step with captured backup path
			if configBackupPath != "" {
				// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces
				// the legacy CR-plus-format pair (REQ-UPR-004).
				plRestore := tui.ProgressLine(out, "Restoring user settings...", nil)
				if restoreErr := backup.RestoreMoaiConfig(projectRoot, configBackupPath, func(pr, relPath string, success bool, errOut io.Writer) {
					// Bridge to the noise-suppression ledger (recordMergeFallback +
					// updateVerboseMode), which stays in package cli. The closure
					// captures updateVerboseMode so the backup subpackage does not
					// need a cross-package mutable-state seam.
					recordMergeFallback(pr, relPath, success, updateVerboseMode, errOut)
				}); restoreErr != nil {
					plRestore.Fail(fmt.Sprintf("Restore failed: %v", restoreErr))
					if reporter != nil {
						reporter.StepError(restoreErr)
					}
					return restoreErr
				}
				plRestore.Done("User settings restored")
				deletedCount := backup.CleanupOldBackups(projectRoot, 5)
				if deletedCount > 0 {
					_, _ = fmt.Fprintf(out, "  %s Cleaned up %d old backup(s)\n", uikit.SymSuccess(), deletedCount)
				}
			}
			// Merge .gitignore: preserve user-added patterns via EntryMerge
			if len(gitignoreBackup) > 0 {
				gitignorePath := filepath.Join(projectRoot, ".gitignore")
				if mergeErr := updatemerge.MergeGitignoreFile(gitignorePath, gitignoreBackup); mergeErr != nil {
					_, _ = fmt.Fprintf(out, "  %s .gitignore merge warning: %v\n", uikit.SymWarning(), mergeErr)
				} else {
					_, _ = fmt.Fprintf(out, "  %s .gitignore user patterns preserved\n", uikit.SymSuccess())
				}
			}
			// Merge user-customized files using 3-way merge engine
			if len(mergeableBackups) > 0 {
				if err := updatemerge.MergeUserFiles(projectRoot, mergeableBackups, out); err != nil {
					_, _ = fmt.Fprintf(out, "  %s File merge warning: %v\n", uikit.SymWarning(), err)
				}
			}
		default:
			// Execute normal step
			if err := step.execute(); err != nil {
				if reporter != nil {
					reporter.StepError(err)
				}
				return err
			}
		}

		// Block progress bar reflecting completed/total deploy steps
		// (REQ-TUXIU-014). Replaces the legacy "N/M steps complete" reporter
		// text; the bar rides the same stdout channel as the step ProgressLines.
		_, _ = fmt.Fprintln(out, renderDeployProgress(i+1, len(steps), th))
	}

	_, _ = fmt.Fprintln(out)
	// Outcome banner (REQ-TUXIU-016): solid success pill + dim detail note.
	renderUpdateOutcome(out, len(analysis.Files), configBackupPath, th)
	report.EmitHooksReviewGuidance(out)

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "To reconfigure project settings (development mode, git, model policy), run:")
	_, _ = fmt.Fprintln(out, "   moai update -c")

	// Ensure global settings.json has required env variables
	if err := ensureGlobalSettingsEnv(); err != nil {
		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Global settings env", "update failed", err.Error(), &th))
	}

	// Install pre-push hook (REQ-CIAUT-002). Non-fatal; --no-hooks opts out.
	installPrePushHookOptional(projectRoot, getBoolFlag(cmd, "no-hooks"), out)

	// Install pre-commit hook (REQ-PC-001). Fast-subset commit tier; --no-hooks opts out.
	installPreCommitHookOptional(projectRoot, getBoolFlag(cmd, "no-hooks"), out)

	return nil
}

// @MX:NOTE: [AUTO] runTemplateSyncWithProgress — M4-S4d-2 DDD migration. Console reporter
// wrapper. Only the key outputs are converted: tui.Pill (skip), tui.Section (analyzing), tui.Pill (cancel).
//
// @MX:NOTE: [AUTO] SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 — return shape extended
// to (skipped bool, err error) so that runUpdate can short-circuit the
// downstream legacy-skill archive block when sync is skipped (version-match
// without --force). skipped == true means "no template files were written";
// callers MUST NOT trigger archive checks in that case.
//
// runTemplateSyncWithProgress runs template sync with simple console output.
// Return values:
//   - skipped: true when version matches and --force is absent (sync was a no-op).
//     When skipped == true, callers should not trigger downstream archive checks.
//   - err: any non-skip error encountered.
//
// A skipped sync is not an error; callers receive (true, nil).
// A user-cancelled merge is also (true, nil) — it is a no-op for downstream
// purposes (no files written), so archive does not need to run.
func runTemplateSyncWithProgress(cmd *cobra.Command) (skipped bool, err error) {
	out := cmd.OutOrStdout()
	th := resolveTheme()
	projectRoot := "."
	autoConfirm := getBoolFlag(cmd, "yes")
	forceUpdate := getBoolFlag(cmd, "force")

	// Use printer-backed console output for progress reporting
	// (REQ-CTX-015: routes step events through the Printer to stderr).
	consoleReporter := newPrinterReporter(
		printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr())))

	// Check for version match before proceeding
	packageVersion := version.GetVersion()
	projectVersion, verr := plan.GetProjectConfigVersion(projectRoot)
	if verr == nil && packageVersion == projectVersion && !forceUpdate {
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: report.RenderOutcome(report.OutcomeAlreadyUpToDate, 0, ""), Theme: &th}))
		return true, nil
	}

	// Confirm merge before proceeding (unless auto-confirm is set)
	if !autoConfirm {
		embedded, eerr := template.EmbeddedTemplates()
		if eerr != nil {
			return false, fmt.Errorf("load embedded templates: %w", eerr)
		}

		deployer := template.NewDeployerWithForceUpdate(embedded, true)
		analysis := updatemerge.AnalyzeMergeChanges(deployer, projectRoot)

		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.Section("Analyzing merge changes", tui.SectionOpts{Theme: &th}))
		proceed, cerr := confirmViaPreview(analysis, projectRoot)
		if cerr != nil {
			return false, fmt.Errorf("confirm merge for %d files (risk: %s): %w",
				len(analysis.Files), analysis.RiskLevel, cerr)
		}
		if !proceed {
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillNeutral, Solid: false, Label: "Merge cancelled by user", Theme: &th}))
			return true, nil
		}
	}

	return false, runTemplateSyncWithReporter(cmd, consoleReporter, true)
}

// toPreviewInputs maps a merge.MergeAnalysis into the neutral
// update.FilePreviewInput slice consumed by the preview entry point (M3c
// convergence, SPEC-CLI-TUX-V3-003 plan Known Issue #7). The Exists/Conflict
// derivation mirrors the same signals the deploy stage enforces — Exists from
// os.Stat (mirrors analyzeFiles), Conflict from RiskLevel=="high" (mirrors
// buildMergeAnalysis hasConflicts). The classification itself (which file is
// preserved/added/updated/conflict) is derived downstream by update.Classify
// via the injected isUserOwnedNamespace predicate (REQ-TUX3-001/002 single
// source of truth — this mapping introduces NO parallel heuristic).
func toPreviewInputs(analysis merge.MergeAnalysis, projectRoot string) []update.FilePreviewInput {
	inputs := make([]update.FilePreviewInput, 0, len(analysis.Files))
	for _, fa := range analysis.Files {
		_, statErr := os.Stat(filepath.Join(projectRoot, fa.Path))
		inputs = append(inputs, update.FilePreviewInput{
			RelPath:  fa.Path,
			Exists:   statErr == nil,
			Conflict: fa.RiskLevel == "high",
			Diff:     fa.Changes,
		})
	}
	return inputs
}

// confirmViaPreview is the single convergence surface for BOTH merge.ConfirmMerge
// call sites in this file (plan Known Issue #7). It launches the Bubble Tea v2
// TUI (AC-TUX3-008/009) when stdin is a terminal, classifying through
// update.Classify with the shared isUserOwnedNamespace predicate
// (REQ-TUX3-001/002/014 — no parallel heuristic, `preserved (user-owned)` label
// derived from the shared predicate).
//
// Behavior preservation relative to the prior merge.ConfirmMerge calls: both
// call sites reach this helper ONLY on the interactive (non-skipped, non-auto)
// path — the skip-confirm and --yes (autoConfirm) branches short-circuit
// upstream and set proceed=true directly, byte-identical to before. When stdin
// is NOT a terminal, the user cannot be asked interactively, so this helper
// returns an error directing them to --yes — mirroring both the explicit
// Windows guard that merge.ConfirmMerge carried (confirm.go REQ-CFS-007/008)
// and the de-facto tea.Run() non-TTY failure on darwin/linux that the v1
// ConfirmMerge relied on. It MUST NOT silently proceed in the non-TTY case:
// that would bypass the confirmation gate the caller explicitly entered (the
// caller did not pass --yes), changing which files get deployed. The
// preview-fallback's proceed=true semantics belong to the --yes abstraction,
// which never reaches this helper.
func confirmViaPreview(analysis merge.MergeAnalysis, projectRoot string) (bool, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return false, fmt.Errorf("merge confirmation UI requires an interactive terminal; " +
			"rerun with --yes to auto-confirm in a non-TTY environment")
	}
	return update.PreviewClassification(toPreviewInputs(analysis, projectRoot), plan.IsUserOwnedNamespace, update.PreviewOptions{Interactive: true})
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

// runInitWizard runs the configuration wizard for reconfiguring an existing project.
// Used by 'moai update -c/--config' to edit project settings.
// @MX:NOTE: [AUTO] runInitWizard — M4-S4d-3 DDD migration. tui.Pill (uninitialized warning,
// cancelled, success) + tui.Section (Reconfiguration header; AC-017 emoji removed).
func runInitWizard(cmd *cobra.Command, reconfigure bool) error {
	out := cmd.OutOrStdout()
	th := resolveTheme()

	// Verify the project is initialized
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if _, err := os.Stat(filepath.Join(cwd, defs.MoAIDir)); os.IsNotExist(err) {
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: "Project not initialized · Run 'moai init' first", Theme: &th}))
		return fmt.Errorf("project not initialized")
	}

	// Print banner and welcome message
	uikit.PrintBanner(version.GetVersion())
	if reconfigure {
		_, _ = fmt.Fprintln(out, tui.Section("Project Reconfiguration Wizard", tui.SectionOpts{Theme: &th}))
		_, _ = fmt.Fprintln(out, "This wizard will help you update your project configuration.")
	} else {
		uikit.PrintWelcomeMessage()
	}

	// REQ-1: Read locale from language.yaml
	locale := wizard.ReadLocaleFromProject(cwd)

	// REQ-2: Read existing username from config (used as default value)
	existingGitHubUsername := wizard.ReadGitHubUsernameFromConfig(cwd)
	existingGitLabUsername := wizard.ReadGitLabUsernameFromConfig(cwd)

	// REQ-3: Check whether gh CLI is authenticated
	ghAuthenticated := wizard.IsGhAuthenticated()

	// Generate default questions and set defaults from existing values
	questions := wizard.ReconfigureQuestions(cwd)
	if existingGitHubUsername != "" {
		if q := wizard.QuestionByID(questions, "github_username"); q != nil {
			q.Default = existingGitHubUsername
		}
	}
	if existingGitLabUsername != "" {
		if q := wizard.QuestionByID(questions, "gitlab_username"); q != nil {
			q.Default = existingGitLabUsername
		}
	}
	// REQ-3: Skip github_token question when gh auth is authenticated
	if ghAuthenticated {
		if q := wizard.QuestionByID(questions, "github_token"); q != nil {
			q.Condition = func(_ *wizard.WizardResult) bool { return false }
		}
	}

	// Run wizard with locale and custom questions
	result, err := wizard.RunWithLocale(questions, nil, locale)
	if err != nil {
		if errors.Is(err, wizard.ErrCancelled) {
			_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillNeutral, Solid: false, Label: "Configuration cancelled", Theme: &th}))
			return nil
		}
		return fmt.Errorf("wizard failed: %w", err)
	}

	// Apply configuration updates to .moai/config/sections/
	// This updates the YAML configuration files based on wizard results
	if err := applyWizardConfig(cwd, result); err != nil {
		return fmt.Errorf("apply configuration: %w", err)
	}

	// F1 security fix: a token entered in the wizard is NOT persisted to any
	// git-tracked config. Tell the user so the credential is not silently
	// dropped — direct them to the provider CLI credential store.
	if result.GitHubToken != "" {
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: "GitHub token was not saved (never stored in plaintext). Run 'gh auth login' to authenticate.", Theme: &th}))
	}
	if result.GitLabToken != "" {
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: "GitLab token was not saved (never stored in plaintext). Run 'glab auth login' to authenticate.", Theme: &th}))
	}

	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Configuration updated successfully", Theme: &th}))

	return nil
}

// applyWizardConfig applies wizard results to the project configuration files.
func applyWizardConfig(projectRoot string, result *wizard.WizardResult) error {
	// F3: validate externally-supplied identity fields before any write so a
	// malformed URL or username never lands in a persisted config file.
	if err := validateWizardInput(result); err != nil {
		return err
	}

	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)

	// user.yaml: Save GitHub/GitLab username plus the wizard-collected user
	// display name. Access tokens are DELIBERATELY NOT persisted here (F1
	// security fix): user.yaml is git-tracked and world-readable, so writing a
	// secret into it leaks the credential. Credentials are delegated to the
	// `gh` / `glab` CLI (see the advisory surfaced by runInitWizard). The block
	// triggers on any non-secret user field (username or display name), so a
	// reconfigure that answers only the user_name question still persists
	// user.name — but a token-only answer creates nothing.
	hasUserFields := result.GitHubUsername != "" ||
		result.GitLabUsername != "" ||
		result.UserName != ""
	if hasUserFields {
		userPath := filepath.Join(sectionsDir, defs.UserYAML)
		// Read existing file
		userData, err := os.ReadFile(userPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read user.yaml: %w", err)
		}

		// Parse YAML
		var user map[string]any
		if len(userData) > 0 {
			if err := yaml.Unmarshal(userData, &user); err != nil {
				return fmt.Errorf("parse user.yaml: %w", err)
			}
		} else {
			user = make(map[string]any)
		}

		// Ensure user.user section exists
		var userConfig map[string]any
		if existingUser, ok := user["user"].(map[string]any); ok {
			userConfig = existingUser
		} else {
			userConfig = make(map[string]any)
		}

		// Save GitHub/GitLab usernames (non-secret). Access tokens
		// (result.GitHubToken / result.GitLabToken) are intentionally NOT
		// written — persisting a secret into the git-tracked user.yaml is the
		// F1 vulnerability this fix closes. Tokens are delegated to the
		// `gh` / `glab` CLI credential store instead.
		if result.GitHubUsername != "" {
			userConfig["github_username"] = result.GitHubUsername
		}
		if result.GitLabUsername != "" {
			userConfig["gitlab_username"] = result.GitLabUsername
		}

		// Save user display name (reconfigure wizard user_name question)
		if result.UserName != "" {
			userConfig["name"] = result.UserName
		}

		user["user"] = userConfig

		// Save to file
		updatedData, err := yaml.Marshal(user)
		if err != nil {
			return fmt.Errorf("marshal user.yaml: %w", err)
		}
		if err := os.WriteFile(userPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write user.yaml: %w", err)
		}
	}

	// language.yaml: Save conversation language (reconfigure wizard language
	// question). Sets both conversation_language and conversation_language_name
	// while preserving all sibling keys (agent_prompt_language, code_comments, ...).
	if result.ConversationLang != "" {
		langPath := filepath.Join(sectionsDir, defs.LanguageYAML)
		langData, err := os.ReadFile(langPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read language.yaml: %w", err)
		}

		var lang map[string]any
		if len(langData) > 0 {
			if err := yaml.Unmarshal(langData, &lang); err != nil {
				return fmt.Errorf("parse language.yaml: %w", err)
			}
		} else {
			lang = make(map[string]any)
		}

		// Ensure language section exists
		var language map[string]any
		if existing, ok := lang["language"].(map[string]any); ok {
			language = existing
		} else {
			language = make(map[string]any)
		}

		language["conversation_language"] = result.ConversationLang
		language["conversation_language_name"] = result.ConversationLang
		lang["language"] = language

		updatedData, err := yaml.Marshal(lang)
		if err != nil {
			return fmt.Errorf("marshal language.yaml: %w", err)
		}
		if err := os.WriteFile(langPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write language.yaml: %w", err)
		}
	}

	// git-strategy.yaml: Save git mode and provider (REQ-4)
	if result.GitMode != "" || result.GitProvider != "" {
		gitStratPath := filepath.Join(sectionsDir, defs.GitStrategyYAML)
		gsData, err := os.ReadFile(gitStratPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read git-strategy.yaml: %w", err)
		}

		var gs map[string]any
		if len(gsData) > 0 {
			if err := yaml.Unmarshal(gsData, &gs); err != nil {
				return fmt.Errorf("parse git-strategy.yaml: %w", err)
			}
		} else {
			gs = make(map[string]any)
		}

		// Ensure git_strategy section exists
		var gitStrategy map[string]any
		if existing, ok := gs["git_strategy"].(map[string]any); ok {
			gitStrategy = existing
		} else {
			gitStrategy = make(map[string]any)
		}

		if result.GitMode != "" {
			gitStrategy["mode"] = result.GitMode
		}
		if result.GitProvider != "" {
			gitStrategy["provider"] = result.GitProvider
		}

		// Save GitLab instance URL (REQ-5)
		if result.GitLabInstanceURL != "" {
			var gitlabSection map[string]any
			if existing, ok := gitStrategy["gitlab"].(map[string]any); ok {
				gitlabSection = existing
			} else {
				gitlabSection = make(map[string]any)
			}
			gitlabSection["instance_url"] = result.GitLabInstanceURL
			gitStrategy["gitlab"] = gitlabSection
		}

		gs["git_strategy"] = gitStrategy

		updatedData, err := yaml.Marshal(gs)
		if err != nil {
			return fmt.Errorf("marshal git-strategy.yaml: %w", err)
		}
		if err := os.WriteFile(gitStratPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write git-strategy.yaml: %w", err)
		}
	}

	// SPEC-MODEL-PROFILE-MATRIX-001 (REQ-MPM-016/024): the former plan_type × tier
	// agent-frontmatter mutation is RETIRED — persist the resolved model policy to
	// llm.profile (normalized {high,medium,low} → {max,medium,low}) instead of
	// mutating agent frontmatter, which stays at model: inherit.
	if result.ModelPolicy != "" {
		policy := template.ModelPolicy(result.ModelPolicy)
		if template.IsValidModelPolicy(string(policy)) {
			if err := template.ApplyProfile(projectRoot, template.NormalizeToTier(result.ModelPolicy)); err != nil {
				return fmt.Errorf("apply profile: %w", err)
			}
			// Persist model_policy to system.yaml so it survives future updates
			systemPath := filepath.Join(sectionsDir, defs.SystemYAML)
			systemData, _ := os.ReadFile(systemPath)
			var sys map[string]any
			if len(systemData) > 0 {
				_ = yaml.Unmarshal(systemData, &sys)
			}
			if sys == nil {
				sys = make(map[string]any)
			}
			moaiSection, _ := sys["moai"].(map[string]any)
			if moaiSection == nil {
				moaiSection = make(map[string]any)
			}
			moaiSection["model_policy"] = string(policy)
			sys["moai"] = moaiSection
			if updatedData, err := yaml.Marshal(sys); err == nil {
				_ = os.WriteFile(systemPath, updatedData, defs.FilePerm)
			}
		}
	}

	// quality.yaml: Save development mode (REQ-4)
	if result.DevelopmentMode != "" {
		qualityPath := filepath.Join(sectionsDir, defs.QualityYAML)
		qualityData, err := os.ReadFile(qualityPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read quality.yaml: %w", err)
		}

		var quality map[string]any
		if len(qualityData) > 0 {
			if err := yaml.Unmarshal(qualityData, &quality); err != nil {
				return fmt.Errorf("parse quality.yaml: %w", err)
			}
		} else {
			quality = make(map[string]any)
		}

		// Ensure constitution section exists
		var constitution map[string]any
		if existing, ok := quality["constitution"].(map[string]any); ok {
			constitution = existing
		} else {
			constitution = make(map[string]any)
		}

		constitution["development_mode"] = result.DevelopmentMode
		quality["constitution"] = constitution

		updatedData, err := yaml.Marshal(quality)
		if err != nil {
			return fmt.Errorf("marshal quality.yaml: %w", err)
		}
		if err := os.WriteFile(qualityPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write quality.yaml: %w", err)
		}
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

// ensureGlobalSettingsEnv cleans up moai-managed settings from ~/.claude/settings.json.
// All settings (env, permissions, teammateMode, hooks) are managed at the project level.
// The global hooks directory (~/.claude/hooks/moai/) is also removed since hooks
// are only deployed to project-level directories via moai init.
func ensureGlobalSettingsEnv() error {
	homeDir, err := userHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	// Remove global hooks/moai directory if it exists.
	// Hooks are project-level only; the global directory causes "No such file or directory"
	// errors in non-initialized projects that reference $CLAUDE_PROJECT_DIR paths.
	globalHooksDir := filepath.Join(homeDir, defs.ClaudeDir, "hooks", "moai")
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
