package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/tmux"
)

// --- Unified Launch ---

// unifiedLaunchFunc is the function used by unifiedLaunch. Override in tests.
var unifiedLaunchFunc = unifiedLaunchDefault

// newDetectorFn constructs the tmux detector used by applyCGMode. It is a
// package-level seam so tests can inject a fake detector (e.g. to simulate the
// REQ-CGH-008 tmux-present-but-unavailable state: InTmuxSession()==true while
// IsAvailable()==false). Production code uses the real SystemDetector.
var newDetectorFn = func() tmux.Detector { return tmux.NewDetector() }

// recordLastProfileFn is the seam unifiedLaunchDefault uses to write the launch
// ledger. It exists so a ledger-write failure can be injected directly
// (SPEC-PROFILE-MEMORY-001 REQ-PM-014). The alternative — provoking a real
// write failure through file permissions — is void when the tests run as root
// and unjudgeable on Windows CI.
var recordLastProfileFn = profile.RecordLastUsedProfileForProject

// launcherStderr is the seam the launch path writes user-facing notices and
// warnings to. warnNoModelResolved takes an io.Writer but its call site
// hardcodes os.Stderr, so it makes the FUNCTION testable while leaving the
// LAUNCH PATH's output uncapturable — and the launch path's output is exactly
// what the fresh-profile notice and the record-failure warning must be judged
// on (SPEC-PROFILE-MEMORY-001 AC-PM-010c / AC-PM-018).
var launcherStderr io.Writer = os.Stderr

// injectTmuxSessionEnvFn is the seam applyCGMode uses to inject GLM credentials
// into the tmux session env. It exists so the REQ-CGH-002 ordering invariant
// (leader-cred strip BEFORE injection) can be tested by forcing an injection
// failure. Production code uses the real injectTmuxSessionEnv.
var injectTmuxSessionEnvFn = injectTmuxSessionEnv

// unifiedLaunch delegates to unifiedLaunchFunc for testability.
func unifiedLaunch(profileName, modeOverride string, extraArgs []string) error {
	return unifiedLaunchFunc(profileName, modeOverride, extraArgs)
}

// resolveMode determines the effective LLM mode.
// Falls back to "claude" when mode is empty.
func resolveMode(mode string) string {
	if mode != "" {
		return mode
	}
	return "claude"
}

// isNamedProfile reports whether name identifies a named profile directory.
// "" and "default" both resolve to the base preferences rather than a
// subdirectory, so they are excluded — this matches the names
// profile.RecordLastUsedProfile refuses and the names profile.GetProfileDir
// maps to "". Letting "default" through would make HasClaudeConfig("default")
// permanently false and emit a bogus notice on every `-p default` launch.
func isNamedProfile(name string) bool {
	return name != "" && name != "default"
}

// warnNoModelResolved reports that the launch resolved no model, so Claude Code
// will apply its own settings.json default rather than a profile value.
//
// profileName is the profile the launch actually read — "" means the base
// preferences.yaml, which is the case a stale launch ledger degrades to.
// Writing to an injected io.Writer keeps the message assertable in tests.
func warnNoModelResolved(w io.Writer, profileName string) {
	read := profileName
	if read == "" {
		read = "default (base preferences)"
	}
	_, _ = fmt.Fprintf(w,
		"Warning: no model resolved from profile %q — Claude Code will use the "+
			"model in .claude/settings.json instead.\n"+
			"  Check with: moai profile current && moai profile list\n"+
			"  Select one with: moai cc -p <profile>\n", read)
}

// warnFreshProfile reports that the target profile carries no Claude Code
// account state yet, so this launch will land on the login / onboarding screen.
//
// The notice is a warning only: nothing is copied, moved, or synthesized
// between profile directories, and no platform credential store is touched.
// That is deliberate — the credential carrier differs by platform, so a "helpful"
// seed would be wrong somewhere. The text names no command this build does not
// ship.
//
// Self-gating: an unnamed profile, or one that already carries state, produces
// no output at all, so the call site does not repeat the condition.
func warnFreshProfile(w io.Writer, profileName string) {
	if !isNamedProfile(profileName) || profile.HasClaudeConfig(profileName) {
		return
	}
	_, _ = fmt.Fprintf(w,
		"Notice: profile %q has no Claude Code configuration yet.\n"+
			"  Claude Code will show the login / onboarding screen on this launch.\n"+
			"  Account state is not inherited between profiles; sign in once and it\n"+
			"  persists for this profile.\n", profileName)
}

// unifiedLaunchDefault centralizes launch logic for all modes (claude, glm, claude_glm).
//
// @MX:ANCHOR: [AUTO] step order is load-bearing: root → resolve → mode → EnsureDir → record → exec
// @MX:REASON: [AUTO] fan_in=3 (runCC/runCG/runGLM via unifiedLaunch). Two orderings are contracts, not
// style: (1) root precedes resolution because project-scoped resolution consumes it — reverting leaves the
// projects map written but never read on the launch path; (2) EnsureDir precedes the ledger write because
// the recorder refuses names whose directory is absent — reverting makes every first-time `-p <new>` launch
// silently unrecorded. The EnsureDir gate reads the resolved profileName while the record gate reads
// originalProfile; they diverge only when originalProfile is "", where no record happens, so the recorded
// name always matches the created directory.
func unifiedLaunchDefault(profileName, modeOverride string, extraArgs []string) error {
	// 1. Determine effective LLM mode (command decides mode, not profile)
	mode := resolveMode(modeOverride)

	// 2. Find project root. This precedes resolution because the fallback is
	// now project-scoped and therefore needs the root. A failure here aborts
	// the launch exactly as it did before the reorder — mode setup below needs
	// the root regardless, so there is nothing to continue with.
	root, err := findProjectRootFn()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}

	// 3. Resolve last-used-profile fallback for bare launches (no -p flag).
	// When profileName is empty and the default profile is unconfigured, fall
	// back to this project's remembered profile, then to the most recently
	// -p-launched named profile recorded in launch.yaml. Explicit -p always
	// wins; MOAI_NO_PROFILE_FALLBACK=1 opts out of both lookups.
	originalProfile := profileName
	resolved := profile.ResolveLaunchProfileForProject(root, profileName)
	if resolved != profileName && resolved != "" {
		_, _ = fmt.Fprintf(launcherStderr, "Using last-used profile '%s' (default profile has no preferences). Use -p default to override.\n", resolved)
		profileName = resolved
	}

	// 4. Apply mode-specific env setup
	switch mode {
	case "glm":
		if err := applyGLMMode(root, profileName); err != nil {
			return err
		}
	case "claude_glm":
		if err := applyCGMode(root, profileName); err != nil {
			return err
		}
	default: // "claude" and any unknown mode
		if err := applyCCMode(root); err != nil {
			return err
		}
	}

	// 4.5. Materialize the profile directory the launch will actually use, and
	// warn when it carries no Claude Code account state yet.
	//
	// The gate is profileName — the RESOLVED name — not originalProfile. A bare
	// launch that resolves through the projects map is still "targeting a named
	// profile", and that path is precisely how a console-created profile (which
	// gets a directory but no .claude.json) reaches its first launch. Gating on
	// originalProfile would skip this block entirely for that case and leave the
	// user with the silent login screen this SPEC exists to remove.
	//
	// Widening the gate cannot create a directory the user did not ask for: a
	// resolved name is only returned after the resolver has confirmed its
	// directory exists, so EnsureDir is a no-op there. Real creation happens
	// only for an explicit -p <new>, which is the intended behavior.
	if isNamedProfile(profileName) {
		if err := profile.EnsureDir(profileName); err != nil {
			return fmt.Errorf("set profile: %w", err)
		}
		warnFreshProfile(launcherStderr, profileName)
	}

	// 5. Record the last-used profile. Only NAMED profiles the user explicitly
	// passed via -p are recorded — this uses originalProfile (what the user
	// typed), not the resolved value: writing back a resolved name would
	// promote the global fallback into a project-scoped entry the user never
	// chose. Best-effort: a write failure is logged but never blocks the
	// launch. This runs AFTER the directory exists (step 4.5 — the recorder
	// refuses names with no directory) and BEFORE launchClaude, which on POSIX
	// does syscall.Exec and replaces the process, so no code runs after it.
	if isNamedProfile(originalProfile) {
		if err := recordLastProfileFn(root, originalProfile); err != nil {
			_, _ = fmt.Fprintf(launcherStderr, "Warning: failed to record last-used profile: %v\n", err)
		}
	}

	// 6. Launch claude
	return launchClaude(profileName, extraArgs)
}

// --- Mode Application ---

// applyCCMode prepares the environment for Claude-only mode.
func applyCCMode(root string) error {
	if err := clearTmuxSessionEnv(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to clear tmux session env: %v\n", err)
	}

	settingsPath := filepath.Join(root, defs.ClaudeDir, defs.SettingsLocalJSON)
	if err := removeGLMEnv(settingsPath); err != nil {
		return fmt.Errorf("remove GLM env: %w", err)
	}

	teamModeMsg := resetTeamModeForCC(root)
	if teamModeMsg != "" {
		fmt.Fprintln(os.Stderr, teamModeMsg)
	}

	worktreeMsg := cleanupMoaiWorktrees(root)
	if worktreeMsg != "" {
		fmt.Fprintln(os.Stderr, worktreeMsg)
	}

	fmt.Fprintln(os.Stderr, "Launching Claude Code...")
	return nil
}

// applyGLMMode prepares the environment for GLM-only mode.
func applyGLMMode(root, profileName string) error {
	glmConfig, err := loadGLMConfig(root)
	if err != nil {
		return fmt.Errorf("load GLM config: %w", err)
	}

	apiKey := getGLMAPIKey(glmConfig.EnvVar)
	if apiKey == "" {
		return fmt.Errorf("GLM API key not found.\n\n"+
			"Save your key first:\n"+
			"  moai glm setup <api-key>\n\n"+
			"Or set the %s environment variable", glmConfig.EnvVar)
	}

	setGLMEnv(glmConfig, apiKey)

	// Auto-enable Z.AI MCP server (Vision, Web Search, Web Reader).
	// Non-blocking: warns on failure, never blocks launch.
	autoEnableMCPServer()

	// settings.local.json injection is intentionally omitted here: setGLMEnv()
	// already sets env for the current process which syscall.Exec inherits into
	// `claude`. Writing to settings.local.json (as previous behavior) would leak
	// GLM env to subsequent `claude` invocations after `moai glm` exits.
	// Tmux team panes still receive env via injectTmuxSessionEnv below (moai cg path).
	// For persistent settings.local.json injection used by `moai --team`, see enableTeamMode().

	if err := persistTeamMode(root, "glm"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to persist team mode: %v\n", err)
	}

	if tmux.NewDetector().InTmuxSession() {
		if err := injectTmuxSessionEnv(glmConfig, apiKey); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to inject GLM env into tmux session: %v\n"+
				"  Teammates spawned in new tmux panes may not have GLM credentials.\n"+
				"  Manually set %s in new panes if needed.\n", err, glmConfig.EnvVar)
		}
	}

	fmt.Fprintln(os.Stderr, "Launching Claude Code with GLM backend...")
	return nil
}

// applyCGMode prepares the environment for Claude + GLM hybrid mode.
func applyCGMode(root, profileName string) error {
	glmConfig, err := loadGLMConfig(root)
	if err != nil {
		return fmt.Errorf("load GLM config: %w", err)
	}

	apiKey := getGLMAPIKey(glmConfig.EnvVar)
	if apiKey == "" {
		return fmt.Errorf("GLM API key not found\n\n"+
			"Set up your API key first, then enable CG mode:\n"+
			"  1. moai glm setup <api-key>   (saves key to ~/.moai/.env.glm)\n"+
			"  2. moai cg                     (enable hybrid mode)\n\n"+
			"Or set the %s environment variable", glmConfig.EnvVar)
	}

	settingsPath := filepath.Join(root, defs.ClaudeDir, defs.SettingsLocalJSON)
	detector := newDetectorFn()
	inTmux := detector.InTmuxSession()

	if !inTmux && os.Getenv(config.EnvTestMode) != "1" {
		return fmt.Errorf("CG mode requires a tmux session.\n\n" +
			"Claude Code itself supports iTerm2 split panes natively (v2.1.186+),\n" +
			"but moai cg injects GLM credentials into teammate panes via tmux\n" +
			"session-level env (set-environment). iTerm2 has no session-level env,\n" +
			"so Leader=Claude / Teammates=GLM isolation requires tmux.\n\n" +
			"  - This pane (lead): uses Claude API\n" +
			"  - New panes (teammates): inherit GLM env for Z.AI API\n\n" +
			"Start a tmux session first:\n" +
			"  tmux new -s moai\n" +
			"  moai cg\n\n" +
			"Or use 'moai glm' for all-GLM mode (no tmux required)")
	}

	// REQ-CGH-008: in a tmux session, the tmux binary must actually be available.
	// A tmux-present-but-binary-missing state (e.g. TMUX env inherited but tmux not
	// on PATH) yields a clear "tmux not installed" error rather than the misleading
	// "restart your tmux session" message emitted on injection failure below.
	if inTmux && !detector.IsAvailable() {
		return fmt.Errorf("tmux is not installed or not executable.\n\n" +
			"CG mode injects GLM credentials into the tmux session env, which " +
			"requires the tmux binary on PATH.\n\n" +
			"Install tmux first:\n" +
			"  macOS:  brew install tmux\n" +
			"  Debian: sudo apt-get install tmux\n\n" +
			"Then start a session and re-run:\n" +
			"  tmux new -s moai\n" +
			"  moai cg")
	}

	// REQ-CGH-002 + REQ-CGH-003: strip stale GLM credentials from the leader config
	// AND set teammateMode=tmux in a SINGLE locked+atomic read-modify-write, BEFORE
	// the failure-prone tmux injection below. This guarantees a tmux-injection
	// failure cannot leave stale GLM credentials in the leader's env block, and no
	// intermediate file state exists where teammateMode is absent.
	if err := mutateSettingsLocal(settingsPath, stripGLMCredsAndSetTeammateMode); err != nil {
		return fmt.Errorf("clean up GLM env for CG mode: %w", err)
	}

	if inTmux {
		if err := injectTmuxSessionEnvFn(glmConfig, apiKey); err != nil {
			return fmt.Errorf("failed to inject GLM env into tmux session: %w\n"+
				"CG mode relies on tmux session env for teammate isolation.\n"+
				"Try restarting your tmux session", err)
		}

		if profileName != "" && profileName != "default" && !isTestEnvironment() {
			profileDir := profile.GetProfileDir(profileName)
			if profileDir != "" {
				tmuxCmd := exec.Command("tmux", "set-environment", "CLAUDE_CONFIG_DIR", profileDir)
				_ = tmuxCmd.Run()
			}
		}
	}

	if err := persistTeamMode(root, "cg"); err != nil {
		return fmt.Errorf("persist team mode: %w", err)
	}

	fmt.Fprintln(os.Stderr, "CG mode: Lead (Claude) + Teammates (GLM)")
	fmt.Fprintln(os.Stderr, "Launching Claude Code...")
	return nil
}

// --- Mode Helpers (moved from cc.go) ---

// removeGLMEnv removes GLM environment variables from settings.local.json.
//
// SPEC-CLIFIX-CRITICAL-001 REQ-CRIT-001-001: round-trips as map[string]any so
// unknown top-level keys survive the write.
func removeGLMEnv(settingsPath string) error {
	// Non-locked pre-check: if the file is absent or empty, there are no GLM env
	// keys to clean. Preserves the original no-op-on-empty behavior.
	preRead, err := readSettingsMap(settingsPath)
	if err != nil {
		return err
	}
	if len(preRead) == 0 {
		return nil
	}

	// SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-001: route through the locked+atomic
	// mutateSettingsLocal seam so concurrent sessions cannot lose updates.
	return mutateSettingsLocal(settingsPath, func(m map[string]any) {
		if len(m) == 0 {
			return
		}

		// Clear teammateMode override so settings.json default ("auto") applies.
		// CG/GLM modes set this to "tmux"; CC mode should restore the default.
		// delete (rather than m["teammateMode"] = "") so the key is omitted entirely,
		// matching the original struct's omitempty behavior for "".
		delete(m, "teammateMode")

		if env, ok := m["env"].(map[string]any); ok {
			// Restore backed-up OAuth token before removing GLM vars
			if backup, bok := env["MOAI_BACKUP_AUTH_TOKEN"].(string); bok && backup != "" {
				env[config.EnvAnthropicAuthToken] = backup
				delete(env, "MOAI_BACKUP_AUTH_TOKEN")
			} else {
				delete(env, config.EnvAnthropicAuthToken)
			}
			delete(env, config.EnvAnthropicBaseURL)
			delete(env, config.EnvAnthropicDefaultHaikuModel)
			delete(env, config.EnvAnthropicDefaultSonnetModel)
			delete(env, config.EnvAnthropicDefaultOpusModel)
			delete(env, config.EnvAnthropicDefaultFableModel)
			// Remove Z.AI proxy compatibility flags (set by moai glm/cg)
			delete(env, config.EnvClaudeCodeDisableExperimentalBetas)
			delete(env, "API_TIMEOUT_MS")
			delete(env, config.EnvClaudeCodeDisableNonessentialTraffic)
			// Remove teammate display env var override (CG/GLM set this)
			delete(env, config.EnvClaudeCodeTeammateDisplay)
			// Issue #742: drop GLM context-size hint when leaving GLM mode so the
			// statusline reverts to the Claude slot's nominal size.
			delete(env, "MOAI_STATUSLINE_CONTEXT_SIZE")
			// SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-002: drop the 1M auto-compact
			// window so it does not persist into subsequent moai cc sessions.
			delete(env, config.EnvClaudeCodeAutoCompactWindow)

			if len(env) == 0 {
				delete(m, "env")
			}
		}
	})
}

// resetTeamModeForCC disables team_mode when switching to CC.
// Returns a message string describing what was changed, or empty if unchanged.
func resetTeamModeForCC(projectRoot string) string {
	mgr := config.NewConfigManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		return ""
	}

	cfg := mgr.Get()
	if cfg == nil || cfg.LLM.TeamMode == "" {
		return ""
	}

	prev := cfg.LLM.TeamMode
	if err := disableTeamMode(projectRoot); err != nil {
		return fmt.Sprintf("Warning: failed to disable team mode: %v", err)
	}
	return fmt.Sprintf("Team mode disabled (was: %s)", prev)
}

// resolveSymlinks returns the symlink-resolved form of path. For paths that
// exist on disk it resolves symlinks via filepath.EvalSymlinks (so macOS
// /var/folders → /private/var/folders prefix matching works); for paths that
// do NOT exist it falls back to filepath.Clean(path) lexically.
//
// The non-existent-path fallback MUST be lexical, NOT EvalSymlinks, because
// filepath.EvalSymlinks's behavior on a non-existent path diverges across
// GOOS: on darwin/linux it returns the cleaned path unchanged (no error),
// while on windows it can partially-resolve the existing prefix (transforming
// 8.3 short names like RUNNER~1 → runneradmin) and leave the non-existent
// tail verbatim — yielding a partially-resolved string that no longer matches
// a prefix built from the same short-name home directory. This divergence
// breaks the launcher's -w validation for absolute paths to worktrees that
// do not exist yet (e.g. re-entry into a not-yet-created path). Returning a
// deterministic lexical Clean on non-existent paths makes prefix-matching
// GOOS-independent for the launcher's pre-creation validation pass.
//
// @MX:NOTE: [AUTO] GOOS-deterministic symlink resolution fallback for non-existent paths (SPEC-WORKTREE-ENTRY-STRATEGY-001 M3a windows CI fix)
// @MX:REASON: Windows filepath.EvalSymlinks on non-existent paths partially-resolves the existing prefix (8.3 short-name → long form), diverging from darwin/linux lexical-clean behavior and breaking prefix-matching for the launcher's pre-worktree-creation -w validation; gating EvalSymlinks behind an existence check restores GOOS-independent determinism while preserving macOS /var → /private/var resolution for paths that do exist.
func resolveSymlinks(path string) string {
	if _, err := os.Lstat(path); err != nil {
		// Path does not exist (or is inaccessible): fall back to lexical Clean
		// so prefix-matching is deterministic across GOOS. See function doc.
		return filepath.Clean(path)
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

// cleanupMoaiWorktrees removes moai-related git worktrees from both the
// local .claude/worktrees/ path and the global ~/.moai/worktrees/*/ paths.
// These are worktrees created by /moai --team with names like worker-SPEC-XXX.
func cleanupMoaiWorktrees(projectRoot string) string {
	// Build the list of base paths that may contain worker worktrees.
	// Resolve symlinks so prefix matching works correctly across platforms
	// (e.g., macOS /var/folders → /private/var/folders).
	var basePaths []string

	// 1. Local Claude Native worktree path.
	localBase := filepath.Join(projectRoot, ".claude", "worktrees")
	if _, err := os.Stat(localBase); err == nil {
		basePaths = append(basePaths, resolveSymlinks(localBase))
	}

	// 2. Global ~/.moai/worktrees/*/ paths (MoAI worktree migration target).
	if homeDir, err := os.UserHomeDir(); err == nil {
		globalBase := filepath.Join(homeDir, ".moai", "worktrees")
		if entries, err := os.ReadDir(globalBase); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					p := resolveSymlinks(filepath.Join(globalBase, entry.Name()))
					basePaths = append(basePaths, p)
				}
			}
		}
	}

	// Skip cleanup when no known worktree locations exist.
	if len(basePaths) == 0 {
		return ""
	}

	gitDir := filepath.Join(projectRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return ""
	}

	output, err := runGitCommand(projectRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return ""
	}

	var cleanedWorktrees []string
	var skippedWorktrees []string

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "worktree ") {
			continue
		}
		rawPath := strings.TrimPrefix(line, "worktree ")
		// Normalize path separators: git on Windows returns forward-slash paths
		// (e.g. C:/Users/...) while filepath.Join produces backslash paths.
		// filepath.FromSlash converts to OS-native separators for correct comparison.
		worktreePath := filepath.FromSlash(rawPath)
		workerName := filepath.Base(worktreePath)
		if !strings.HasPrefix(workerName, "worker-") {
			continue
		}
		for _, base := range basePaths {
			// Use filepath.Rel instead of strings.HasPrefix to avoid false positives
			// from sibling directories sharing a common prefix (e.g. "myproject" vs
			// "myproject-old") and path separator mismatches on Windows.
			rel, err := filepath.Rel(base, worktreePath)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				continue
			}
			// Use the full path so git can locate the worktree regardless
			// of whether it is under .claude/worktrees/ or ~/.moai/worktrees/.
			// Non-force removal: git refuses when the worktree still holds
			// uncommitted or untracked work, so a failure here means "kept",
			// not "broken". See removeWorktree.
			if err := removeWorktree(projectRoot, worktreePath); err == nil {
				cleanedWorktrees = append(cleanedWorktrees, workerName)
			} else {
				skippedWorktrees = append(skippedWorktrees, workerName)
			}
			break
		}
	}

	var msgs []string
	if len(cleanedWorktrees) > 0 {
		msgs = append(msgs, fmt.Sprintf("Cleaned up %d worktree(s): %s", len(cleanedWorktrees), strings.Join(cleanedWorktrees, ", ")))
	}
	if len(skippedWorktrees) > 0 {
		msgs = append(msgs, fmt.Sprintf(
			"Kept %d worktree(s) with local changes: %s (review, then remove explicitly with 'moai worktree clean --stale' or 'git worktree remove --force <path>')",
			len(skippedWorktrees), strings.Join(skippedWorktrees, ", ")))
	}
	return strings.Join(msgs, "\n")
}

// removeWorktree removes a single git worktree without --force.
//
// Omitting --force is deliberate. This runs on every `moai cc` launch via
// cleanupMoaiWorktrees, so an unconditional --force would let an ordinary
// launch delete a worktree that still holds uncommitted or untracked work,
// with no confirmation and no way back. Without --force git refuses on a
// dirty worktree and exits non-zero, which the caller reports as skipped.
// Deliberate removal of a dirty worktree stays an explicit user action
// (`moai worktree remove` / `git worktree remove --force`).
//
// @MX:ANCHOR: [AUTO] launch-path worktree removal is non-destructive by contract
// @MX:REASON: cleanupMoaiWorktrees calls this unconditionally on every `moai cc`
// launch; re-adding --force here silently destroys uncommitted work across every
// worker-* worktree without user confirmation.
func removeWorktree(projectRoot, worktreeName string) error {
	_, err := runGitCommand(projectRoot, "worktree", "remove", worktreeName)
	return err
}

// runGitCommand executes a git command in the given directory.
func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}
	// Do not TrimSpace to preserve trailing newline (matches original exec.Output() behavior)
	return string(out) + "\n", nil
}

// --- Claude Launch ---

// launchClaudeFunc is the function used by launchClaude. Override in tests.
var launchClaudeFunc = launchClaudeDefault

// launchClaude delegates to launchClaudeFunc for testability.
func launchClaude(profileName string, extraArgs []string) error {
	return launchClaudeFunc(profileName, extraArgs)
}

// launchClaudeDefault finds the claude binary, reads DO_CLAUDE_* settings from
// settings.local.json, and replaces the current process with claude via
// syscall.Exec. profileName may be empty for the default profile. extraArgs
// are additional CLI args to pass through to claude.
func launchClaudeDefault(profileName string, extraArgs []string) error {
	// 1. Profile setup
	if profileName != "" && profileName != "default" {
		if err := profile.EnsureDir(profileName); err != nil {
			return fmt.Errorf("set profile: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Profile: %s\n", profileName)
	}

	// 2. Find claude binary
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH. Install Claude Code first")
	}

	// 3. Read profile preferences and sync to project config
	prefs, _ := profile.ReadPreferences(profileName)
	if root, err := findProjectRoot(); err == nil {
		moaiDir := filepath.Join(root, ".moai")
		if info, err := os.Stat(moaiDir); err == nil && info.IsDir() {
			_ = profile.SyncToProjectConfig(root, prefs)
		}
		// Sync permission mode preference to settings.local.json permissions.defaultMode
		settingsLocalPath := filepath.Join(root, defs.ClaudeDir, defs.SettingsLocalJSON)
		if err := syncPermissionModeToSettingsLocal(settingsLocalPath, prefs.PermissionMode); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to sync permission mode setting: %v\n", err)
		}
	}

	// 4. Read project settings.local.json for DO_CLAUDE_* flags (overrides profile)
	settings := readSettingsLocalForLaunch()

	// Profile model is the default; settings.local.json overrides
	if settings["DO_CLAUDE_MODEL"] == "" && prefs.Model != "" {
		settings["DO_CLAUDE_MODEL"] = prefs.Model
	}

	// Permission mode: profile preference is the base default.
	// settings.local.json DO_CLAUDE_PERMISSION_MODE overrides profile.
	// Legacy DO_CLAUDE_BYPASS is also honored for backward compatibility.
	permMode := prefs.PermissionMode
	if settings["DO_CLAUDE_PERMISSION_MODE"] != "" {
		permMode = settings["DO_CLAUDE_PERMISSION_MODE"]
	} else if settings["DO_CLAUDE_BYPASS"] == "true" && permMode == "" {
		permMode = "bypassPermissions"
	}
	chrome := settings["DO_CLAUDE_CHROME"] == "true"
	cont := settings["DO_CLAUDE_CONTINUE"] == "true"
	model := settings["DO_CLAUDE_MODEL"]

	// 5. Parse extra args (overrides)
	var passThrough []string
	for i := 0; i < len(extraArgs); i++ {
		arg := extraArgs[i]
		switch arg {
		case "--chrome":
			chrome = true
		case "--no-chrome":
			chrome = false
		case "-b", "--bypass":
			permMode = "bypassPermissions"
		case "--permission-mode":
			if i+1 < len(extraArgs) {
				permMode = extraArgs[i+1]
				i++
			}
		case "-c", "--continue":
			cont = true
		case "--model", "-m":
			if i+1 < len(extraArgs) {
				model = extraArgs[i+1]
				i++
			}
		default:
			// Handle --permission-mode=value form
			if strings.HasPrefix(arg, "--permission-mode=") {
				permMode = strings.TrimPrefix(arg, "--permission-mode=")
			} else {
				passThrough = append(passThrough, arg)
			}
		}
	}

	// 6. Resolve model string. Under a GLM backend the --model flag MUST carry a
	// slot alias (opus/sonnet/...) so it routes through the ANTHROPIC_DEFAULT_*_MODEL
	// slot env that setGLMEnv configured; under a Claude backend short aliases
	// expand to canonical ids as before (byte-identical to expandModelString).
	glmBackend := false
	if root, err := findProjectRoot(); err == nil {
		glmBackend = resolveGLMBackendForLaunch(root)
	}
	model = resolveMainSessionModel(model, glmBackend)

	// 6b. An empty model here is not neutral: buildArgs below omits --model
	// entirely, so Claude Code falls back to whatever `.claude/settings.json`
	// pins (the MoAI template ships "sonnet"). That looked identical to "my
	// profile was ignored", with nothing on stderr to distinguish the two.
	//
	// The common cause is a profile that was edited but never selected: the
	// launch ledger still points elsewhere, or at a profile whose directory is
	// gone, and profile.ResolveLaunchProfile's stale-record guard degrades to
	// the empty base preferences. Name the resolved profile so the user can see
	// which one was actually read.
	if model == "" {
		warnNoModelResolved(os.Stderr, profileName)
	}

	// 7. Build args
	buildArgs := func(withContinue bool) []string {
		a := []string{"claude"}
		if permMode != "" && permMode != "acceptEdits" {
			a = append(a, "--permission-mode", permMode)
		}
		if !chrome {
			a = append(a, "--no-chrome")
		}
		if withContinue {
			a = append(a, "--continue")
		}
		if model != "" {
			a = append(a, "--model", model)
		}
		a = append(a, passThrough...)
		return a
	}

	// 7. Execute with --continue fallback
	if cont {
		tryCmd := exec.Command(claudeBin, buildArgs(true)[1:]...)
		tryCmd.Stdin = os.Stdin
		tryCmd.Stdout = os.Stdout
		tryCmd.Stderr = os.Stderr
		err := tryCmd.Run()
		if err == nil {
			return nil
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			fmt.Fprintln(os.Stderr, "No previous session found, starting new session...")
		} else {
			return fmt.Errorf("resume session failed: %w", err)
		}
	}

	// NOTE: On POSIX, execOrSpawnClaude replaces the current process entirely
	// (syscall.Exec); no defer() functions run after that point. On Windows it
	// spawns a child and exits with the child's code (syscall.Exec is POSIX-only
	// — REQ-CGH-001). Ensure all cleanup and setup is complete before calling.
	effectiveEffort := resolveLaunchEffort(prefs.EffortLevel, prefs.ModelPolicy)
	var launchEnv []string
	if glmBackend {
		// GLM backend: z.ai honors reasoning_effort, NOT Claude's 5-step effort.
		// Derive ANTHROPIC_REASONING_EFFORT from the effective effort and strip the
		// inert CLAUDE_CODE_EFFORT_LEVEL so a web-set effort reaches z.ai.
		launchEnv = buildEnvForGLMLaunch(effectiveEffort, os.Environ())
	} else {
		// Claude backend: honors the 5-step effort vocabulary (CLAUDE_CODE_EFFORT_LEVEL).
		launchEnv = buildEnvForLaunch(effectiveEffort, os.Environ())
	}
	return execOrSpawnClaude(claudeBin, buildArgs(false), launchEnv)
}

// --- Flag Parsing ---

// parseProfileFlag extracts -p/--profile from args and returns the profile name
// and the remaining args with the flag removed.
// Returns an error if -p/--profile is specified without a value.
func parseProfileFlag(args []string) (string, []string, error) {
	var profileName string
	filtered := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			// Everything after -- is pass-through to claude
			filtered = append(filtered, args[i:]...)
			break
		}
		if args[i] == "--profile" || args[i] == "-p" {
			if i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "-") {
				return "", nil, fmt.Errorf("flag %s requires a profile name\n\nUsage:\n  moai <command> -p <profile-name>\n\nExamples:\n  moai cg -p work\n  moai cc -p default", args[i])
			}
			profileName = args[i+1]
			i++
			continue
		}
		// Handle --profile=value form
		if strings.HasPrefix(args[i], "--profile=") {
			profileName = strings.TrimPrefix(args[i], "--profile=")
			if profileName == "" {
				return "", nil, fmt.Errorf("flag --profile= requires a non-empty profile name\n\nUsage:\n  moai <command> -p <profile-name>\n\nExamples:\n  moai cg -p work\n  moai cc --profile=default")
			}
			continue
		}
		if strings.HasPrefix(args[i], "-p=") {
			profileName = strings.TrimPrefix(args[i], "-p=")
			if profileName == "" {
				return "", nil, fmt.Errorf("flag -p= requires a non-empty profile name\n\nUsage:\n  moai <command> -p <profile-name>\n\nExamples:\n  moai cg -p work\n  moai cc -p=default")
			}
			continue
		}
		filtered = append(filtered, args[i])
	}

	return profileName, filtered, nil
}

// normalizeWorktreeFlag rewrites -w/--worktree[=name] forms (before any "--"
// pass-through marker) into the canonical two-token "--worktree [name]" form
// that Claude Code accepts, preserving argument order. Tokens after "--" are
// left untouched — they are already verbatim pass-through to claude.
// The name is optional: a bare -w lets claude auto-generate a worktree name.
func normalizeWorktreeFlag(args []string) []string {
	normalized := make([]string, 0, len(args)+1)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			// Everything after -- is verbatim pass-through.
			normalized = append(normalized, args[i:]...)
			break
		}
		switch {
		case arg == "-w" || arg == "--worktree":
			normalized = append(normalized, "--worktree")
			// Optional value: consume the next token unless it is another flag
			// or the pass-through marker.
			if i+1 < len(args) && args[i+1] != "--" && !strings.HasPrefix(args[i+1], "-") {
				normalized = append(normalized, args[i+1])
				i++
			}
		case strings.HasPrefix(arg, "--worktree="):
			name := strings.TrimPrefix(arg, "--worktree=")
			normalized = append(normalized, "--worktree")
			if name != "" {
				normalized = append(normalized, name)
			}
		case strings.HasPrefix(arg, "-w="):
			name := strings.TrimPrefix(arg, "-w=")
			normalized = append(normalized, "--worktree")
			if name != "" {
				normalized = append(normalized, name)
			}
		default:
			normalized = append(normalized, arg)
		}
	}

	return normalized
}

// resolveWorktreeL2Path is the MoAI-side pre-resolution step for the
// -w/--worktree flag value. It runs BEFORE normalizeWorktreeFlag in the cc /
// cg / glm launch paths.
//
// SPEC-WORKTREE-ENTRY-STRATEGY-001 M3a (REQ-WES-010). The launcher's -w flag
// accepts three kinds of values:
//
//  1. A short name (e.g. "feat-login"). normalizeWorktreeFlag rewrites it into
//     the canonical "--worktree <name>" two-token form for claude pass-through;
//     claude then resolves the name against .claude/worktrees/<name>. This
//     step MUST NOT interfere with short-name inputs.
//  2. An absolute path under ~/.moai/worktrees/<project>/... (an L2 persistent
//     worktree created by `moai worktree new` or the auto-isolation
//     procedure). claude accepts an absolute --worktree value and uses it
//     directly, so this step accepts the path and leaves args unchanged.
//  3. An absolute path under <project>/.claude/worktrees/... (an L1
//     Claude-native worktree). Treated the same as case 2.
//
// An absolute path NOT under either accepted prefix is REJECTED with a clear
// error so the launcher does not silently fall through to creating a new
// worktree under either prefix (AC-WES-010c).
//
// Tokens after the "--" pass-through marker are not scanned (they are verbatim
// pass-through to claude). Returns nil when no -w value is present or the
// value is a short name; returns a non-nil error only for out-of-prefix
// absolute paths.
//
// This function is ADDITIVE: normalizeWorktreeFlag is unchanged and remains
// the owner of short-name token normalization (AC-WES-010b).
func resolveWorktreeL2Path(args []string) error {
	value, ok := worktreeFlagValue(args)
	if !ok || value == "" {
		// Bare -w (auto-name) or no -w flag: nothing to validate.
		return nil
	}
	if !filepath.IsAbs(value) {
		// Short name: defer to normalizeWorktreeFlag + claude resolution.
		return nil
	}

	// Absolute path: must be under an accepted worktree prefix.
	var acceptedPrefixes []string
	if homeDir, err := os.UserHomeDir(); err == nil {
		acceptedPrefixes = append(acceptedPrefixes, filepath.Join(homeDir, ".moai", "worktrees"))
	}
	if root, err := findProjectRoot(); err == nil {
		acceptedPrefixes = append(acceptedPrefixes, filepath.Join(root, ".claude", "worktrees"))
	}

	for _, prefix := range acceptedPrefixes {
		if isUnderWorktreePrefix(value, prefix) {
			return nil
		}
	}

	return fmt.Errorf(
		"worktree path %q is not under an accepted worktree prefix\n"+
			"  accepted prefixes: ~/.moai/worktrees/ (L2 persistent), .claude/worktrees/ (L1 Claude-native)\n"+
			"  use a short name to create a new worktree under .claude/worktrees/<name>, or an\n"+
			"  absolute path under one of the accepted prefixes to re-enter an existing worktree",
		value,
	)
}

// worktreeFlagValue scans args (stopping at the "--" pass-through marker) for
// the value of the -w / --worktree flag. Returns the value and whether a
// value was supplied. Mirrors the token-shape handling of normalizeWorktreeFlag
// without mutating args.
func worktreeFlagValue(args []string) (string, bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			// Everything after -- is verbatim pass-through — do not scan.
			return "", false
		}
		switch {
		case arg == "-w" || arg == "--worktree":
			// Optional value: the next token is the value unless it is another
			// flag or the pass-through marker.
			if i+1 < len(args) && args[i+1] != "--" && !strings.HasPrefix(args[i+1], "-") {
				return args[i+1], true
			}
			// Bare -w (claude auto-generates a worktree name): no value.
			return "", false
		case strings.HasPrefix(arg, "--worktree="):
			return strings.TrimPrefix(arg, "--worktree="), true
		case strings.HasPrefix(arg, "-w="):
			return strings.TrimPrefix(arg, "-w="), true
		}
	}
	return "", false
}

// isUnderWorktreePrefix reports whether path is contained within prefix.
// Both path and prefix are Clean'd, then symlink-resolved (matching
// cleanupMoaiWorktrees) so that macOS /var/folders → /private/var/folders
// prefix matching works. The comparison uses filepath.Rel, which is
// separator-aware on every GOOS (handles Windows `\` natively); no
// forward-slash string matching is performed.
// The prefix itself is NOT considered "under" (a prefix path is not a valid
// worktree path; only a child of the prefix is).
//
// @MX:NOTE: [AUTO] cross-platform path containment check (SPEC-WORKTREE-ENTRY-STRATEGY-001 M3a)
// @MX:REASON: Windows compatibility — `~` expands via os.UserHomeDir() (USERPROFILE on Windows,
// HOME on Unix); filepath.Rel handles `\` and `/` separators via stdlib, so no manual ToSlash needed.
func isUnderWorktreePrefix(path, prefix string) bool {
	cleanedPath := filepath.Clean(path)
	cleanedPrefix := filepath.Clean(prefix)
	resolvedPath := resolveSymlinks(cleanedPath)
	resolvedPrefix := resolveSymlinks(cleanedPrefix)
	rel, err := filepath.Rel(resolvedPrefix, resolvedPath)
	if err != nil {
		return false
	}
	if rel == "." {
		// path == prefix exactly; not a valid worktree path.
		return false
	}
	// rel must not escape prefix (no ".." or ".."-prefixed relative path).
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

// readSettingsLocalForLaunch reads the env map from .claude/settings.local.json
// in the current directory (or project root). Returns an empty map on error.
func readSettingsLocalForLaunch() map[string]string {
	result := make(map[string]string)

	// Try project root first, fall back to current directory
	settingsPath := filepath.Join(".claude", "settings.local.json")
	root, err := findProjectRoot()
	if err == nil {
		settingsPath = filepath.Join(root, ".claude", "settings.local.json")
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return result
	}

	var settings SettingsLocal
	if err := json.Unmarshal(data, &settings); err != nil {
		return result
	}

	for k, v := range settings.Env {
		result[k] = v
	}
	return result
}

// syncPermissionModeToSettingsLocal persists the profile permission mode
// preference to .claude/settings.local.json so that permissions.defaultMode
// survives across sessions regardless of how Claude Code is launched.
//
// When permissionMode is a non-default value (e.g. "auto", "bypassPermissions"),
// it sets permissions.defaultMode in settings.local.json.
// When permissionMode is empty or "acceptEdits" (matching the project default),
// it removes the defaultMode override so settings.json default applies.
//
// The empty-string normalization for "acceptEdits" is intentional AND surfaced
// to the user: runProfileSetup emits an explicit confirmation line
// (acceptEditsConfirmationLine) so the user does not perceive the selection as
// a silent no-op. See profile_setup.go runProfileSetup normalization block
// (REQ-CCI-006 / REQ-CCI-007 — the normalization is intentional, and it is
// disclosed to the user via the wizard confirmation, not silently applied).
// syncPermissionModeToSettingsLocal persists the profile permission mode
// preference to .claude/settings.local.json so that permissions.defaultMode
// survives across sessions regardless of how Claude Code is launched.
//
// SPEC-CLIFIX-CRITICAL-001 REQ-CRIT-001-001: round-trips as map[string]any so
// unknown top-level keys survive the write.
func syncPermissionModeToSettingsLocal(settingsPath string, permissionMode string) error {
	// SPEC-CLIFIX-CONCURRENCY-001 REQ-CONC-001-001: route through the locked+atomic
	// mutateSettingsLocal seam so concurrent sessions cannot lose updates.
	return mutateSettingsLocal(settingsPath, func(m map[string]any) {
		// Only write an override when the mode differs from the project default.
		// The project settings.json default is "acceptEdits", so we skip writing
		// for empty string and "acceptEdits" to avoid unnecessary overrides.
		if permissionMode != "" && permissionMode != "acceptEdits" {
			perms, _ := m["permissions"].(map[string]any)
			if perms == nil {
				perms = make(map[string]any)
			}
			perms["defaultMode"] = permissionMode
			m["permissions"] = perms
		} else {
			// Remove the override so settings.json default applies
			if perms, ok := m["permissions"].(map[string]any); ok {
				delete(perms, "defaultMode")
				if len(perms) == 0 {
					delete(m, "permissions")
				}
			}
		}
	})
}

// expandModelString normalizes moai-specific model strings into valid Claude
// Code --model values. Short aliases (opus, sonnet, haiku, opusplan) are
// resolved to their canonical Claude Code model id via the central
// template.ModelAliasTable; the "[1m]" suffix is preserved across resolution
// because Claude Code natively supports it (e.g. "opus[1m]",
// "claude-opus-4-7[1m]") to enable the 1M token context window. Values that
// are already canonical ids or are unknown pass through unchanged.
func expandModelString(model string) string {
	if model == "" {
		return model
	}
	base, suffix := splitModelSuffix(model)
	resolved, ok := template.ModelAliasTable[base]
	if !ok {
		return model // already canonical or unknown — pass through unchanged
	}
	if suffix == "" {
		return resolved
	}
	return resolved + suffix
}

// splitModelSuffix separates a model string into its base alias/id and the
// optional "[1m]" context-window suffix. The suffix is recognized only when it
// appears as a literal trailing token; mid-string occurrences are left intact.
func splitModelSuffix(model string) (base, suffix string) {
	const marker = "[1m]"
	if strings.HasSuffix(model, marker) {
		return model[:len(model)-len(marker)], marker
	}
	return model, ""
}

// buildEnvForLaunch returns an environment slice with CLAUDE_CODE_EFFORT_LEVEL
// set to effortLevel when non-empty. Any existing CLAUDE_CODE_EFFORT_LEVEL entry
// in base is replaced to avoid duplicates. When effortLevel is empty, base is
// returned unchanged.
//
// @MX:NOTE: [AUTO] Effort injection point (Claude backend) — model ROUTING (ModelPolicy→model) is orthogonal to effort; effort SOURCING now falls back to a model_policy-derived effort (resolveLaunchEffort→MapModelPolicyToEffort) when prefs.EffortLevel is empty. The routing⊥effort invariant still holds.
func buildEnvForLaunch(effortLevel string, base []string) []string {
	if effortLevel == "" {
		return base
	}
	key := config.EnvClaudeCodeEffortLevel
	entry := key + "=" + effortLevel
	result := make([]string, 0, len(base)+1)
	replaced := false
	for _, e := range base {
		if strings.HasPrefix(e, key+"=") {
			result = append(result, entry)
			replaced = true
		} else {
			result = append(result, e)
		}
	}
	if !replaced {
		result = append(result, entry)
	}
	return result
}

// resolveLaunchEffort resolves the CLAUDE_CODE_EFFORT_LEVEL value for the launch
// from the two profile levers: explicit prefs.EffortLevel always wins; otherwise
// the model_policy-derived effort (template.MapModelPolicyToEffort) is used as a
// fallback; both empty → "" (no override, byte-identical to today's launch).
// model-ROUTING (prefs.Model → DO_CLAUDE_MODEL) remains orthogonal to effort
// sourcing — only effort SOURCING reads model_policy here.
func resolveLaunchEffort(effortLevel, modelPolicy string) string {
	if effortLevel != "" {
		return effortLevel
	}
	if modelPolicy != "" {
		if e := template.MapModelPolicyToEffort(template.ModelPolicy(modelPolicy)); e != "" {
			return e
		}
	}
	return ""
}

// buildEnvForGLMLaunch returns an environment slice for a GLM-backed main
// session. It strips any inherited CLAUDE_CODE_EFFORT_LEVEL (z.ai does NOT
// implement Claude's 5-level effort — that var is inert under the z.ai proxy)
// and injects ANTHROPIC_REASONING_EFFORT derived from the web-set effort via
// the GLM effort overlay, so a web-set effort reaches z.ai through the channel
// it honors. Any pre-existing ANTHROPIC_REASONING_EFFORT (setGLMEnv writes the
// hardcoded coding-max default) is replaced so the prefs-derived value wins.
// When the effort collapse disables thinking, no reasoning-effort entry is
// emitted (reasoning_effort is moot when thinking is off).
func buildEnvForGLMLaunch(effort string, base []string) []string {
	result := make([]string, 0, len(base)+1)
	for _, e := range base {
		if strings.HasPrefix(e, config.EnvClaudeCodeEffortLevel+"=") {
			continue // inert under z.ai — strip
		}
		if strings.HasPrefix(e, config.EnvAnthropicReasoningEffort+"=") {
			continue // re-derived from effort below
		}
		result = append(result, e)
	}
	for k, v := range glmReasoningEnvVarsForEffort(effort) {
		result = append(result, k+"="+v)
	}
	return result
}

// resolveMainSessionModel resolves the --model flag value for the main session,
// GLM-aware. Under a Claude backend (glmBackend==false) it is byte-identical to
// expandModelString (short alias → canonical id). Under a GLM backend
// (glmBackend==true) it REVERSE-maps any canonical id back to its slot alias
// (opus/sonnet/haiku/fable) via template.ModelAliasFromCanonicalID, so the
// --model flag routes through the ANTHROPIC_DEFAULT_*_MODEL slot env that
// setGLMEnv configured — z.ai's Anthropic-compat shim binds those env vars only
// to slot aliases, and a literal canonical claude-* id is sent straight to the
// z.ai proxy base URL, bypassing the slot→GLM-model mapping and silently
// defeating the web-set model preference. Already-alias values pass through
// unchanged; the [1m] suffix is preserved; unknown values pass through.
func resolveMainSessionModel(prefsModel string, glmBackend bool) string {
	if !glmBackend {
		return expandModelString(prefsModel)
	}
	if prefsModel == "" {
		return ""
	}
	base, suffix := splitModelSuffix(prefsModel)
	alias := template.ModelAliasFromCanonicalID(base)
	if suffix == "" {
		return alias
	}
	return alias + suffix
}

// resolveGLMBackendForLaunch reports whether the main-session launch is under a
// GLM backend, by reading the persisted team_mode from llm.yaml
// (template.IsGLMBackend). applyGLMMode / applyCGMode persist team_mode BEFORE
// launchClaude runs, so the llm.yaml signal is authoritative at this point. On
// any read error it resolves false (safe default — the Claude-backend launch
// path), matching the pre-existing behavior for a non-GLM checkout.
func resolveGLMBackendForLaunch(root string) bool {
	sectionsDir := filepath.Join(filepath.Clean(root), defs.MoAIDir, defs.SectionsSubdir)
	llm, err := loadLLMSectionOnly(sectionsDir)
	if err != nil {
		return false
	}
	return template.IsGLMBackend(llm)
}

// syncBypassToSettingsLocal is a backward-compatible wrapper for
// syncPermissionModeToSettingsLocal. It maps bypass=true to "bypassPermissions".
// Deprecated: Use syncPermissionModeToSettingsLocal directly.
func syncBypassToSettingsLocal(settingsPath string, bypass bool) error {
	mode := ""
	if bypass {
		mode = "bypassPermissions"
	}
	return syncPermissionModeToSettingsLocal(settingsPath, mode)
}
