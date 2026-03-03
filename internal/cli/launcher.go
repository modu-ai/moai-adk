package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
	gitops "github.com/modu-ai/moai-adk/internal/git/ops"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/tmux"
)

// --- Unified Launch ---

// unifiedLaunchFunc is the function used by unifiedLaunch. Override in tests.
var unifiedLaunchFunc = unifiedLaunchDefault

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

// @MX:ANCHOR: [AUTO] unifiedLaunchDefault centralizes launch logic for all modes
// @MX:REASON: [AUTO] fan_in=3, called from runCC, runCG, runGLM via unifiedLaunch
// unifiedLaunchDefault centralizes launch logic for all modes (claude, glm, claude_glm).
func unifiedLaunchDefault(profileName, modeOverride string, extraArgs []string) error {
	// 1. Determine effective LLM mode (command decides mode, not profile)
	mode := resolveMode(modeOverride)

	// 3. Find project root
	root, err := findProjectRootFn()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
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

	// 5. Launch claude
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

	if err := persistTeamMode(root, "glm"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to persist team mode: %v\n", err)
	}

	settingsPath := filepath.Join(root, defs.ClaudeDir, defs.SettingsLocalJSON)
	if err := injectGLMEnvForTeam(settingsPath, glmConfig, apiKey); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to inject GLM env into settings: %v\n", err)
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
	detector := tmux.NewDetector()
	inTmux := detector.InTmuxSession()

	if !inTmux && os.Getenv("MOAI_TEST_MODE") != "1" {
		return fmt.Errorf("CG mode requires a tmux session.\n\n" +
			"tmux is required because:\n" +
			"  - This pane (lead): uses Claude API\n" +
			"  - New panes (teammates): inherit GLM env for Z.AI API\n\n" +
			"Start a tmux session first:\n" +
			"  tmux new -s moai\n" +
			"  moai cg\n\n" +
			"Or use 'moai glm' for all-GLM mode (no tmux required)")
	}

	if inTmux {
		if err := injectTmuxSessionEnv(glmConfig, apiKey); err != nil {
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

	if err := removeGLMEnv(settingsPath); err != nil {
		return fmt.Errorf("clean up GLM env for CG mode: %w", err)
	}

	if err := ensureSettingsLocalJSON(settingsPath); err != nil {
		return fmt.Errorf("ensure settings.local.json: %w", err)
	}

	fmt.Fprintln(os.Stderr, "CG mode: Lead (Claude) + Teammates (GLM)")
	fmt.Fprintln(os.Stderr, "Launching Claude Code...")
	return nil
}

// --- Mode Helpers (moved from cc.go) ---

// removeGLMEnv removes GLM environment variables from settings.local.json.
func removeGLMEnv(settingsPath string) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read settings.local.json: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var settings SettingsLocal
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parse settings.local.json: %w", err)
	}

	if settings.Env != nil {
		// Restore backed-up OAuth token before removing GLM vars
		if backup, ok := settings.Env["MOAI_BACKUP_AUTH_TOKEN"]; ok && backup != "" {
			settings.Env["ANTHROPIC_AUTH_TOKEN"] = backup
			delete(settings.Env, "MOAI_BACKUP_AUTH_TOKEN")
		} else {
			delete(settings.Env, "ANTHROPIC_AUTH_TOKEN")
		}
		delete(settings.Env, "ANTHROPIC_BASE_URL")
		delete(settings.Env, "ANTHROPIC_DEFAULT_HAIKU_MODEL")
		delete(settings.Env, "ANTHROPIC_DEFAULT_SONNET_MODEL")
		delete(settings.Env, "ANTHROPIC_DEFAULT_OPUS_MODEL")

		if len(settings.Env) == 0 {
			settings.Env = nil
		}
	}

	data, err = json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		return fmt.Errorf("write settings.local.json: %w", err)
	}

	return nil
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

// cleanupMoaiWorktrees removes moai-related git worktrees.
// These are worktrees created by /moai --team with names like worker-SPEC-XXX.
func cleanupMoaiWorktrees(projectRoot string) string {
	worktreeBase := filepath.Join(projectRoot, ".claude", "worktrees")
	if _, err := os.Stat(worktreeBase); os.IsNotExist(err) {
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

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			worktreePath := strings.TrimPrefix(line, "worktree ")
			if strings.HasPrefix(worktreePath, worktreeBase) {
				workerName := filepath.Base(worktreePath)
				if strings.HasPrefix(workerName, "worker-") {
					if err := removeWorktree(projectRoot, workerName); err == nil {
						cleanedWorktrees = append(cleanedWorktrees, workerName)
					}
				}
			}
		}
	}

	if len(cleanedWorktrees) > 0 {
		return fmt.Sprintf("Cleaned up %d worktree(s): %s", len(cleanedWorktrees), strings.Join(cleanedWorktrees, ", "))
	}
	return ""
}

// removeWorktree removes a single git worktree.
func removeWorktree(projectRoot, worktreeName string) error {
	_, err := runGitCommand(projectRoot, "worktree", "remove", "--force", worktreeName)
	return err
}

// runGitCommand executes a git command in the given directory.
// GitManager를 통해 실행하여 타임아웃과 에러 처리를 일관되게 적용한다.
func runGitCommand(dir string, args ...string) (string, error) {
	mgr := gitops.NewGitManager(gitops.ManagerConfig{
		WorkDir:               dir,
		DefaultTimeoutSeconds: 10,
		DefaultRetryCount:     0,
	})
	result := mgr.ExecuteRaw(args, 10)
	if !result.Success {
		if result.Error != nil {
			return "", result.Error
		}
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), result.Stderr)
	}
	// trailing newline 보존을 위해 TrimSpace 하지 않음 (기존 exec.Output() 동작과 동일)
	return result.Stdout + "\n", nil
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
	}

	// 4. Read project settings.local.json for DO_CLAUDE_* flags (overrides profile)
	settings := readSettingsLocalForLaunch()

	// Profile model is the default; settings.local.json overrides
	if settings["DO_CLAUDE_MODEL"] == "" && prefs.Model != "" {
		settings["DO_CLAUDE_MODEL"] = prefs.Model
	}

	// Profile bypass merges with settings.local.json
	bypass := settings["DO_CLAUDE_BYPASS"] == "true" || prefs.Bypass
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
			bypass = true
		case "-c", "--continue":
			cont = true
		case "--model", "-m":
			if i+1 < len(extraArgs) {
				model = extraArgs[i+1]
				i++
			}
		default:
			passThrough = append(passThrough, arg)
		}
	}

	// 6. Build args
	buildArgs := func(withContinue bool) []string {
		a := []string{"claude"}
		if bypass {
			a = append(a, "--dangerously-skip-permissions")
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

	// NOTE: syscall.Exec replaces the current process entirely.
	// No defer() functions will execute after this point.
	// Ensure all cleanup and setup is complete before calling.
	return syscall.Exec(claudeBin, buildArgs(false), os.Environ())
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
