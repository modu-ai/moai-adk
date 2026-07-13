package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
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

// fileBackup holds a file path and its backed-up content for merging.
type fileBackup struct {
	path string
	data []byte
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Sync MoAI-ADK project templates to the latest version",
	GroupID: "project",
	Long:    "Check for binary updates, install if available, then synchronize embedded templates with the project.",
	PreRunE: validateUpdateFlags,
	RunE:    runUpdate,
}

// validateUpdateFlags validates update flag values before execution.
// SPEC-MODEL-TIER-PLANTYPE-001 M3 (REQ-MTP-018): an out-of-set --plan-type value
// exits non-zero with a usage error naming the closed set {api, subscription}.
func validateUpdateFlags(cmd *cobra.Command, _ []string) error {
	planType := getStringFlag(cmd, "plan-type")
	if planType != "" && !config.IsValidPlanType(planType) {
		return fmt.Errorf("invalid --plan-type value %q: must be one of: api, subscription", planType)
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

	// SPEC-MODEL-TIER-PLANTYPE-001 M3 (REQ-MTP-018): plan_type override. When
	// provided, persists the new value to llm.yaml and re-applies the tier profile;
	// when absent, update reads the persisted llm.plan_type (absent → subscription).
	updateCmd.Flags().String("plan-type", "", "Override the billing plan type: api or subscription (persists to llm.yaml plan_type and re-applies the tier profile)")
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
			if err := confirm.Run(); err == nil && wantSetup {
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
			_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillInfo, Solid: false, Label: "Already up to date", Theme: &th}))
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
		fingerprint, fpErr := detectV2Fingerprint(cwd)
		if fpErr != nil {
			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "v2 detection", "failed", fpErr.Error(), &th))
		} else if fingerprint.IsV2 {
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
				Out:    out,
				// Inject the canonical .agency/ migration adapter. When the
				// project carries a .agency/ legacy directory, this is invoked
				// in Step 3.5 of the canonical order (REQ-VVCR-025), mirroring
				// the migrateLegacyMemoryDir auto-invoke precedent at line 1731.
				RunMigrateAgency: func(projectRoot string, dryRun bool, out io.Writer) error {
					return runAgencyMigrationAdapter(projectRoot, dryRun, out)
				},
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
		// SPEC-MODEL-TIER-PLANTYPE-001 M3 (REQ-MTP-018): an explicit --plan-type
		// override must still persist + re-apply the tier profile even when the
		// template sync short-circuits (the agent files already exist on disk).
		if pt := getStringFlag(cmd, "plan-type"); pt != "" {
			if err := applyUpdateTierProfile(".", pt); err != nil {
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
	if err := scaffoldEvolutionDir("."); err != nil {
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

	// SPEC-MODEL-TIER-PLANTYPE-001 M3 (REQ-MTP-018): when --plan-type is given,
	// persist the override and re-apply the plan_type × tier profile to the
	// freshly-synced agent files.
	if pt := getStringFlag(cmd, "plan-type"); pt != "" {
		if err := applyUpdateTierProfile(".", pt); err != nil {
			return err
		}
	}

	return nil
}

// applyUpdateTierProfile re-applies the plan_type × tier profile during
// `moai update` (SPEC-MODEL-TIER-PLANTYPE-001 M3, REQ-MTP-018). When planTypeFlag
// is non-empty it PERSISTS the override to llm.yaml first (D3); otherwise the
// effective plan type is read from the persisted llm.plan_type (absent →
// subscription). The tier is read from the persisted performance_tier (absent →
// medium). Agent frontmatter is then patched via ApplyTierProfile.
//
// Returns a graceful nil when the manifest cannot be loaded (a non-initialized
// directory), mirroring the applyWizardConfig guard. An out-of-set planTypeFlag
// returns an error naming the closed set (defensive — the CLI flag is validated
// by validateUpdateFlags before this is reached).
func applyUpdateTierProfile(projectRoot, planTypeFlag string) error {
	if planTypeFlag != "" {
		if !config.IsValidPlanType(planTypeFlag) {
			return fmt.Errorf("invalid --plan-type value %q: must be one of: api, subscription", planTypeFlag)
		}
		if err := template.ApplyPlanType(projectRoot, planTypeFlag); err != nil {
			return fmt.Errorf("persist plan_type: %w", err)
		}
	}

	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		return nil // non-initialized project — nothing to re-apply
	}

	planType := template.ResolveProjectPlanType(projectRoot)
	tier := template.ResolveProjectPerformanceTier(projectRoot)
	if err := template.ApplyTierProfile(projectRoot, planType, tier, mgr); err != nil {
		return fmt.Errorf("apply tier profile: %w", err)
	}
	return nil
}

// emitHooksReviewGuidance writes the advisory /hooks review guidance line.
//
// Claude Code captures a snapshot of .claude/settings.json hooks at session
// startup and uses that snapshot throughout the session. When `moai update`
// re-renders the template (the "Template sync complete" branch), any running
// CC session's snapshot is stale — new/changed hooks do not take effect until
// the user reviews the /hooks menu or restarts Claude Code. This is ADVISORY
// only (CC auto-warns on external modification); it is a stdout message, never
// an interactive prompt (C-HRA-008 / REQ-PGN-012 subagent boundary).
//
// @MX:NOTE: [AUTO] Advisory /hooks review guidance (SPEC-HOOK-CONFIG-SAFETY-001).
// @MX:SPEC: SPEC-HOOK-CONFIG-SAFETY-001
func emitHooksReviewGuidance(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Hooks: running Claude Code sessions need a /hooks review (or a Claude Code restart) for new or changed hooks to take effect.")
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
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Template version up-to-date · Skipping sync", Theme: &th}))
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
	analysis := analyzeMergeChanges(deployer, projectRoot)

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, tui.Section("Analyzing merge changes", tui.SectionOpts{Theme: &th}))

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
				return cleanMoaiManagedPaths(projectRoot, out)
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
	var mergeableBackups []fileBackup

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

	// Execute each step with progress reporting
	for i, step := range steps {
		if reporter != nil {
			reporter.StepStart(step.name, step.message)
		}

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
					mergeableBackups = append(mergeableBackups, fileBackup{path: mf, data: data})
				}
			}
			if reporter != nil {
				reporter.StepComplete("Configuration backed up")
			}
		case "Restore Settings":
			// Handle restore step with captured backup path
			if configBackupPath != "" {
				if reporter != nil {
					reporter.StepStart("Restore Settings", "Restoring user settings")
				}
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
				if reporter != nil {
					reporter.StepComplete("Settings restored")
				}
			}
			// Merge .gitignore: preserve user-added patterns via EntryMerge
			if len(gitignoreBackup) > 0 {
				gitignorePath := filepath.Join(projectRoot, ".gitignore")
				if mergeErr := mergeGitignoreFile(gitignorePath, gitignoreBackup); mergeErr != nil {
					_, _ = fmt.Fprintf(out, "  %s .gitignore merge warning: %v\n", uikit.SymWarning(), mergeErr)
				} else {
					_, _ = fmt.Fprintf(out, "  %s .gitignore user patterns preserved\n", uikit.SymSuccess())
				}
			}
			// Merge user-customized files using 3-way merge engine
			if len(mergeableBackups) > 0 {
				if err := mergeUserFiles(projectRoot, mergeableBackups, out); err != nil {
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
			if reporter != nil {
				reporter.StepComplete(fmt.Sprintf("%s complete", step.name))
			}
		}

		// Update progress for remaining steps
		if reporter != nil && i < len(steps)-1 {
			reporter.StepUpdate(fmt.Sprintf("%d/%d steps complete", i+1, len(steps)))
		}
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Template sync complete", Theme: &th}))
	emitHooksReviewGuidance(out)

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
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Template version up-to-date · Skipping sync", Theme: &th}))
		return true, nil
	}

	// Confirm merge before proceeding (unless auto-confirm is set)
	if !autoConfirm {
		embedded, eerr := template.EmbeddedTemplates()
		if eerr != nil {
			return false, fmt.Errorf("load embedded templates: %w", eerr)
		}

		deployer := template.NewDeployerWithForceUpdate(embedded, true)
		analysis := analyzeMergeChanges(deployer, projectRoot)

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

// mergeGitignoreFile reads the newly deployed .gitignore template and merges
// user-specific patterns from the backup. Template content is kept as-is;
// user-added lines (not present in the template) are appended under a
// "User Custom Patterns" header.
func mergeGitignoreFile(gitignorePath string, userBackup []byte) error {
	templateContent, err := os.ReadFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("read new .gitignore: %w", err)
	}

	// Build a set of non-blank, non-comment lines from the template
	templateLines := strings.Split(string(templateContent), "\n")
	templateSet := make(map[string]bool, len(templateLines))
	for _, line := range templateLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			templateSet[trimmed] = true
		}
	}

	// Collect user-specific lines that are not in the new template
	userLines := strings.Split(string(userBackup), "\n")
	var userAdditions []string
	for _, line := range userLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !templateSet[trimmed] {
			userAdditions = append(userAdditions, line)
		}
	}

	if len(userAdditions) == 0 {
		return nil // No user-specific patterns to preserve
	}

	// Append user additions to the template content
	result := string(templateContent)
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	result += "\n# User Custom Patterns (preserved by moai update)\n"
	for _, line := range userAdditions {
		result += line + "\n"
	}

	return os.WriteFile(gitignorePath, []byte(result), defs.FilePerm)
}

// mergeUserFiles performs 3-way merge for user-customized files after template deployment.
// It uses the manifest's TemplateHash as the base, user's backed-up content as current,
// and the newly deployed template as updated. This preserves user customizations while
// incorporating template changes.
func mergeUserFiles(projectRoot string, backups []fileBackup, out io.Writer) error {
	// Load embedded templates to get original template content for base version
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("load embedded templates: %w", err)
	}

	// Load manifest to get template hashes for base version
	mgr := manifest.NewManager()
	if _, loadErr := mgr.Load(projectRoot); loadErr != nil {
		return fmt.Errorf("load manifest: %w", loadErr)
	}

	// Create merge engine
	engine := merge.NewEngine()

	var mergedCount int
	for _, fb := range backups {
		destPath := filepath.Join(projectRoot, fb.path)

		// Read newly deployed file (updated version)
		updatedContent, err := os.ReadFile(destPath)
		if err != nil {
			// File might not exist in new template version - keep user's version
			if writeErr := os.WriteFile(destPath, fb.data, defs.FilePerm); writeErr != nil {
				return fmt.Errorf("restore removed file %s: %w", fb.path, writeErr)
			}
			_, _ = fmt.Fprintf(out, "  %s %s preserved (removed in new template)\n", uikit.SymSuccess(), fb.path)
			mergedCount++
			continue
		}

		// Get original template content from embedded filesystem for base version
		// Try both with and without leading dot
		possiblePaths := []string{fb.path, strings.TrimPrefix(fb.path, ".")}
		var baseContent []byte
		for _, p := range possiblePaths {
			if data, readErr := fs.ReadFile(embedded, p); readErr == nil {
				baseContent = data
				break
			}
		}

		// Perform 3-way merge: base (original template), current (user's backup), updated (new template)
		// If base is not available, treat as new file - preserve user content
		if baseContent == nil {
			// No base available - this might be a user-created file
			// Prefer user's content but merge if compatible
			if string(fb.data) == string(updatedContent) {
				continue // No change needed
			}
			// Keep user's version as-is
			if err := os.WriteFile(destPath, fb.data, defs.FilePerm); err != nil {
				return fmt.Errorf("restore user file %s: %w", fb.path, err)
			}
			_, _ = fmt.Fprintf(out, "  %s %s user content preserved\n", uikit.SymSuccess(), fb.path)
			mergedCount++
			continue
		}

		// Use merge engine for proper 3-way merge
		result, mergeErr := engine.MergeFile(context.Background(), fb.path, baseContent, fb.data, updatedContent)
		if mergeErr != nil {
			// Merge failed - preserve user's version
			_, _ = fmt.Fprintf(out, "  %s %s merge failed, preserving user version: %v\n", uikit.SymWarning(), fb.path, mergeErr)
			if err := os.WriteFile(destPath, fb.data, defs.FilePerm); err != nil {
				return fmt.Errorf("preserve user file %s: %w", fb.path, err)
			}
			mergedCount++
			continue
		}

		// Write merged result
		if err := os.WriteFile(destPath, result.Content, defs.FilePerm); err != nil {
			return fmt.Errorf("write merged file %s: %w", fb.path, err)
		}

		// Report merge status
		if result.HasConflict {
			_, _ = fmt.Fprintf(out, "  %s %s merged with conflicts (user version preferred)\n", uikit.SymWarning(), fb.path)
		} else {
			_, _ = fmt.Fprintf(out, "  %s %s user customizations preserved\n", uikit.SymSuccess(), fb.path)
		}
		mergedCount++
	}

	if mergedCount > 0 {
		_, _ = fmt.Fprintf(out, "  %s Merged %d file(s) with 3-way merge engine\n", uikit.SymSuccess(), mergedCount)
	}

	return nil
}

// buildMergeAnalysis creates a summary from individual file analysis results.
// It counts high/medium/low risk files, determines overall risk level,
// identifies conflicts, and generates a human-readable summary.
func buildMergeAnalysis(files []merge.FileAnalysis) merge.MergeAnalysis {
	var highRisk, medRisk, lowRisk int
	for _, f := range files {
		switch f.RiskLevel {
		case "high":
			highRisk++
		case "medium":
			medRisk++
		case "low":
			lowRisk++
		}
	}

	overallRisk := "low"
	hasConflicts := false
	if highRisk > 0 {
		overallRisk = "high"
		hasConflicts = true
	} else if medRisk > 0 {
		overallRisk = "medium"
	}

	summary := fmt.Sprintf("Found %d files to sync", len(files))
	if highRisk > 0 {
		summary += fmt.Sprintf(" (%d high-risk files)", highRisk)
	}

	return merge.MergeAnalysis{
		Files:        files,
		HasConflicts: hasConflicts,
		SafeToMerge:  highRisk == 0,
		Summary:      summary,
		RiskLevel:    overallRisk,
	}
}

// analyzeMergeChanges performs a quick analysis of template files that will be modified.
// It evaluates risk levels based on file types and existing content:
//   - High risk: CLAUDE.md, settings.json (core config files)
//   - Medium risk: Existing files being updated
//   - Low risk: New files being created
//
// Returns a MergeAnalysis with file-by-file risk assessment and recommended strategies.
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

func analyzeMergeChanges(deployer template.Deployer, projectRoot string) merge.MergeAnalysis {
	templates := deployer.ListTemplates()
	files := plan.AnalyzeFiles(templates, projectRoot)
	return buildMergeAnalysis(files)
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
// cleanMoaiManagedPaths removes MoAI-managed directories and files before template
// deployment. This ensures stale files are cleaned up during version upgrades.
// The .moai/config/ directory is deleted entirely (backup was done by the Backup step).
// Paths that do not exist are silently skipped.
func cleanMoaiManagedPaths(projectRoot string, out io.Writer) error {
	type cleanTarget struct {
		// displayPath is shown in progress messages (e.g., ".claude/settings.json")
		displayPath string
		// fullPath is the absolute filesystem path to delete
		fullPath string
		// isGlob indicates the target uses filepath.Glob matching
		isGlob bool
	}

	targets := []cleanTarget{
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.SettingsJSON),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.SettingsJSON),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.CommandsMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.CommandsMoaiSubdir),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.AgentsMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.AgentsMoaiSubdir),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai*"),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.SkillsSubdir, "moai*"),
			isGlob:      true,
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.RulesMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.RulesMoaiSubdir),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.OutputStylesSubdir, "moai"),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.OutputStylesSubdir, "moai"),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.HooksMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.HooksMoaiSubdir),
		},
	}

	// Process standard targets (files and directories)
	// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces the legacy
	// CR-plus-format pair in this hot path (REQ-UPR-004).
	for _, t := range targets {
		pl := tui.ProgressLine(out, fmt.Sprintf("Removing %s...", t.displayPath), nil)

		if t.isGlob {
			matches, err := filepath.Glob(t.fullPath)
			if err != nil {
				pl.Fail(fmt.Sprintf("Failed to glob %s: %v", t.displayPath, err))
				return fmt.Errorf("glob %s: %w", t.displayPath, err)
			}
			for _, match := range matches {
				if err := os.RemoveAll(match); err != nil {
					pl.Fail(fmt.Sprintf("Failed to remove %s: %v", t.displayPath, err))
					return fmt.Errorf("remove %s: %w", match, err)
				}
			}
			pl.Done(fmt.Sprintf("Removed %s", t.displayPath))
			continue
		}

		if _, err := os.Stat(t.fullPath); err != nil {
			if os.IsNotExist(err) {
				// Use Done with a "skipped" marker — semantically a successful
				// no-op (target already absent). The leading symbol shifts from
				// "-" to "✓" but the message text is preserved; this aligns
				// the visual category with the success branch.
				pl.Done(fmt.Sprintf("Skipped %s (not found)", t.displayPath))
				continue
			}
			pl.Fail(fmt.Sprintf("Failed to stat %s: %v", t.displayPath, err))
			return fmt.Errorf("stat %s: %w", t.displayPath, err)
		}

		if err := os.RemoveAll(t.fullPath); err != nil {
			pl.Fail(fmt.Sprintf("Failed to remove %s: %v", t.displayPath, err))
			return fmt.Errorf("remove %s: %w", t.displayPath, err)
		}
		pl.Done(fmt.Sprintf("Removed %s", t.displayPath))
	}

	// Clean .moai/config/ entirely - backup was already done by the Backup step.
	// For v1.x -> v2.x: old config is incompatible, fresh install needed.
	// For v2.x -> v2.x: backup includes sections/, restore will merge values back.
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)
	configDisplayPath := filepath.Join(defs.MoAIDir, defs.ConfigSubdir)
	// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces the legacy
	// CR-plus-format pair (REQ-UPR-004).
	plConfig := tui.ProgressLine(out, fmt.Sprintf("Removing %s...", configDisplayPath), nil)

	if err := os.RemoveAll(configDir); err != nil {
		if !os.IsNotExist(err) {
			plConfig.Fail(fmt.Sprintf("Failed to remove %s: %v", configDisplayPath, err))
			return fmt.Errorf("remove %s: %w", configDisplayPath, err)
		}
	}
	plConfig.Done(fmt.Sprintf("Removed %s", configDisplayPath))

	// Migrate legacy .moai/memory/ to .moai/state/.
	// Prior to v2.x, state files (checkpoints, coverage, diagnostics) lived under
	// .moai/memory/. If the old directory still exists, migrate or remove it.
	if err := migrateLegacyMemoryDir(projectRoot, out); err != nil {
		return err
	}

	return nil
}

// migrateLegacyMemoryDir handles the .moai/memory/ → .moai/state/ migration.
// If only the old directory exists, it is renamed. If both exist, the old one
// is removed (the new directory takes precedence). If neither exists, this is
// a no-op because template deployment will create .moai/state/.
func migrateLegacyMemoryDir(projectRoot string, out io.Writer) error {
	legacyDir := filepath.Join(projectRoot, defs.MoAIDir, "memory")
	stateDir := filepath.Join(projectRoot, defs.MoAIDir, defs.StateSubdir)

	legacyDisplayPath := filepath.Join(defs.MoAIDir, "memory")

	legacyExists := false
	if _, err := os.Stat(legacyDir); err == nil {
		legacyExists = true
	}

	if !legacyExists {
		return nil
	}

	// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces the legacy
	// CR-plus-format pair (REQ-UPR-004).
	plLegacy := tui.ProgressLine(out, fmt.Sprintf("Migrating %s...", legacyDisplayPath), nil)

	stateExists := false
	if _, err := os.Stat(stateDir); err == nil {
		stateExists = true
	}

	if !stateExists {
		// Rename .moai/memory/ → .moai/state/ (fast atomic move).
		if err := os.Rename(legacyDir, stateDir); err != nil {
			plLegacy.Fail(fmt.Sprintf("Failed to migrate %s: %v", legacyDisplayPath, err))
			return fmt.Errorf("migrate %s to %s: %w", legacyDisplayPath, defs.StateSubdir, err)
		}
		plLegacy.Done(fmt.Sprintf("Migrated %s → %s", legacyDisplayPath, filepath.Join(defs.MoAIDir, defs.StateSubdir)))
	} else {
		// Both exist — state directory takes precedence; remove legacy.
		if err := os.RemoveAll(legacyDir); err != nil {
			plLegacy.Fail(fmt.Sprintf("Failed to remove %s: %v", legacyDisplayPath, err))
			return fmt.Errorf("remove legacy %s: %w", legacyDisplayPath, err)
		}
		plLegacy.Done(fmt.Sprintf("Removed legacy %s", legacyDisplayPath))
	}

	return nil
}

// runAgencyMigrationAdapter is the auto-invoke entrypoint for the
// .agency/ → .moai/ migration triggered by runCleanReinstall Step 3.5
// (REQ-VVCR-025 of SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001).
//
// Unlike runMigrateAgency (which is cobra-driven and reads --dry-run /
// --force / --resume from CLI flags), this adapter constructs the runner
// directly with the values supplied by the clean-reinstall orchestrator.
// It mirrors the auto-invoke precedent of migrateLegacyMemoryDir.
//
// When `.agency/` is absent at projectRoot, the underlying migrateAgency
// runner returns ErrMigrateNoSource — which this adapter swallows
// gracefully because the clean-reinstall flow already verified .agency/
// presence via Signal 2.
func runAgencyMigrationAdapter(projectRoot string, dryRun bool, out io.Writer) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("agency migration adapter: home dir: %w", err)
	}

	r := &migrateAgencyRunner{
		projectRoot: projectRoot,
		homeDir:     homeDir,
		dryRun:      dryRun,
		// force=false: respect existing .moai/ contents; clean-reinstall
		// pre-emptively preserves user-owned namespaces via Step 2 inventory
		// so any forced overwrite here would risk re-introducing collisions.
		force: false,
	}

	if _, runErr := r.Run(); runErr != nil {
		// ErrMigrateNoSource is acceptable when .agency/ disappeared
		// between detection and adapter invocation (race-safe no-op).
		if me, ok := runErr.(*MigrateError); ok && me.Code == ErrMigrateNoSource {
			return nil
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
	questions := wizard.DefaultQuestions(cwd)
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

	_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Configuration updated successfully", Theme: &th}))

	return nil
}

// applyWizardConfig applies wizard results to the project configuration files.
func applyWizardConfig(projectRoot string, result *wizard.WizardResult) error {
	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)

	// user.yaml: Save GitHub/GitLab username and token (REQ-4, REQ-5)
	hasUserFields := result.GitHubUsername != "" || result.GitHubToken != "" ||
		result.GitLabUsername != "" || result.GitLabToken != ""
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

		// Save GitHub credentials
		if result.GitHubUsername != "" {
			userConfig["github_username"] = result.GitHubUsername
		}
		if result.GitHubToken != "" {
			userConfig["github_token"] = result.GitHubToken
		}

		// Save GitLab credentials (REQ-5)
		if result.GitLabUsername != "" {
			userConfig["gitlab_username"] = result.GitLabUsername
		}
		if result.GitLabToken != "" {
			userConfig["gitlab_token"] = result.GitLabToken
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

	// Apply the plan_type × tier profile to agent definition files (project-level,
	// not profile-level). A single pass patches both model: and effort: frontmatter
	// atomically (replace-both precedence). The effective plan type is read from the
	// persisted llm.yaml (absent → subscription); the tier is resolved from
	// result.ModelPolicy (legacy {high, medium, low} → {max, medium, low}).
	if result.ModelPolicy != "" {
		policy := template.ModelPolicy(result.ModelPolicy)
		if template.IsValidModelPolicy(string(policy)) {
			mgr := manifest.NewManager()
			if _, err := mgr.Load(projectRoot); err == nil {
				planType := template.ResolveProjectPlanType(projectRoot)
				tier := template.NormalizeToTier(result.ModelPolicy)
				if err := template.ApplyTierProfile(projectRoot, planType, tier, mgr); err != nil {
					return fmt.Errorf("apply tier profile: %w", err)
				}
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

// scaffoldEvolutionDir ensures the .moai/evolution/ directory tree exists.
// Called during both init and update so that existing projects that predate
// the evolution infrastructure receive the directory structure automatically.
// Non-destructive: only creates missing files; never overwrites existing ones.
func scaffoldEvolutionDir(projectRoot string) error {
	evolutionDir := filepath.Join(projectRoot, ".moai", "evolution")

	// Sub-directories to create.
	subdirs := []string{
		"telemetry",
		"learnings",
		"new-skills",
	}

	for _, sub := range subdirs {
		dir := filepath.Join(evolutionDir, sub)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}

		// Ensure .gitkeep exists so the directory is tracked by git.
		gitkeep := filepath.Join(dir, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			if err := os.WriteFile(gitkeep, []byte{}, 0o644); err != nil {
				return fmt.Errorf("create %s: %w", gitkeep, err)
			}
		}
	}

	// Create default manifest.yaml if missing.
	manifestPath := filepath.Join(evolutionDir, "manifest.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		const defaultManifest = `schema_version: 1
evolved_skills: []
new_skills: []
learnings_count: 0
last_evolution_date: ""
rate_limit:
  week_start: ""
  proposals_this_week: 0
  last_proposal_time: ""
`
		if err := os.WriteFile(manifestPath, []byte(defaultManifest), 0o644); err != nil {
			return fmt.Errorf("create %s: %w", manifestPath, err)
		}
	}

	// Create default changelog.md if missing.
	changelogPath := filepath.Join(evolutionDir, "changelog.md")
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		const defaultChangelog = `# MoAI Evolution Changelog

All notable skill evolutions and learning graduations will be documented here.

## Format

Each entry: date, learning ID, skill affected, change summary.
`
		if err := os.WriteFile(changelogPath, []byte(defaultChangelog), 0o644); err != nil {
			return fmt.Errorf("create %s: %w", changelogPath, err)
		}
	}

	return nil
}
