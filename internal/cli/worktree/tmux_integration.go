package worktree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/tmux"
)

// TmuxIntegration implements SPEC-WORKTREE-002 R5: Tmux Integration requirements.
// After worktree creation, it automatically creates a tmux session and injects the required environment variables.
//
// @MX:NOTE: SPEC-WORKTREE-002 R5 implementation - automatic tmux session creation and env var injection
// @MX:SPEC: SPEC-WORKTREE-002

// TmuxSessionConfig holds the configuration for creating a tmux session.
type TmuxSessionConfig struct {
	// ProjectName is the project name (e.g., "moai-adk-go")
	ProjectName string

	// SpecID is the SPEC identifier (e.g., "SPEC-WORKTREE-002")
	SpecID string

	// WorktreePath is the absolute path to the worktree
	WorktreePath string

	// ActiveMode is the current LLM mode (cc, glm, cg)
	ActiveMode string

	// GLMEnvVars holds the environment variables to inject in GLM/CG mode.
	// Must be empty in CC mode.
	GLMEnvVars map[string]string
}

// CreateTmuxSession creates a tmux session for the worktree.
//
// R5.1: Session name pattern: moai-{ProjectName}-{SPEC-ID}
// R5.2-5.3: Inject environment variables in GLM/CG mode; no injection in CC mode
// R5.4: After session creation, cd to worktree and execute /moai run command
//
// @MX:ANCHOR: Core entry point for the worktree-based development workflow
// @MX:REASON: tmux session automation is a core feature of SPEC-WORKTREE-002, called by multiple clients
// @MX:SPEC: SPEC-WORKTREE-002
func CreateTmuxSession(ctx context.Context, cfg *TmuxSessionConfig, tmuxMgr tmux.SessionManager) error {
	if cfg == nil {
		return fmt.Errorf("tmux session config is required")
	}

	if tmuxMgr == nil {
		return fmt.Errorf("tmux manager is required")
	}

	// R5.1: Generate session name
	sessionName := GenerateTmuxSessionName(cfg.ProjectName, cfg.SpecID)

	// R5.4: Create tmux session (detached mode)
	sessionCfg := &tmux.SessionConfig{
		Name:       sessionName,
		MaxVisible: 1, // Use a single pane
		Panes: []tmux.PaneConfig{
			{
				SpecID:  cfg.SpecID,
				Command: buildTmuxInitialCommand(cfg),
			},
		},
	}

	result, err := tmuxMgr.Create(ctx, sessionCfg)
	if err != nil {
		return fmt.Errorf("create tmux session: %w", err)
	}

	// R5.2-5.3: Inject environment variables in GLM/CG mode
	if cfg.ActiveMode == "glm" || cfg.ActiveMode == "cg" {
		if len(cfg.GLMEnvVars) > 0 {
			if err := tmuxMgr.InjectEnv(ctx, cfg.GLMEnvVars); err != nil {
				return fmt.Errorf("inject GLM env: %w", err)
			}
		}
	}

	// Print log output
	fmt.Printf("Tmux session created: %s\n", result.SessionName)
	fmt.Printf("Panes created: %d\n", result.PaneCount)
	fmt.Printf("Attached: %v\n", result.Attached)
	fmt.Printf("Worktree path: %s\n", cfg.WorktreePath)
	fmt.Printf("To attach: tmux attach-session -t %s\n", sessionName)

	return nil
}

// buildTmuxInitialCommand builds the initial command to run in the tmux pane.
// R5.4: cd to worktree + execute /moai run
func buildTmuxInitialCommand(cfg *TmuxSessionConfig) string {
	// cd to the worktree path
	cdCmd := fmt.Sprintf("cd %s", cfg.WorktreePath)

	// Execute the /moai run command
	moaiCmd := fmt.Sprintf("/moai run %s", cfg.SpecID)

	// Chain the two commands (separated by ;)
	return fmt.Sprintf("%s ; %s", cdCmd, moaiCmd)
}

// IsTmuxAvailable checks whether tmux is available in the current environment.
// R1: Used for tmux availability detection in the Execution Mode Selection Gate.
//
// @MX:NOTE: SPEC-WORKTREE-002 R1 implementation - tmux availability detection
// @MX:SPEC: SPEC-WORKTREE-002
func IsTmuxAvailable() bool {
	// Check the $TMUX environment variable
	return os.Getenv("TMUX") != ""
}

// GetActiveMode reads the current active mode from .moai/config/sections/llm.yaml.
// R1.1: active mode detection.
//
// Returns: "cc" (default/empty), "glm", or "cg"
//
// @MX:NOTE: SPEC-WORKTREE-002 R1.1 implementation - LLM mode detection
// @MX:SPEC: SPEC-WORKTREE-002
func GetActiveMode(projectRoot string) (string, error) {
	llmConfigPath := filepath.Join(projectRoot, ".moai", "config", "sections", "llm.yaml")

	data, err := os.ReadFile(llmConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "cc", nil
		}
		return "cc", fmt.Errorf("read llm.yaml: %w", err)
	}

	// Simple YAML parsing for llm.team_mode field
	// Look for any line containing "team_mode:" regardless of indentation
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "team_mode:") {
			// Extract value after "team_mode:"
			parts := strings.SplitN(trimmed, "team_mode:", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				// Remove quotes if present
				value = strings.Trim(value, "\"")
				value = strings.Trim(value, "'")
				if value == "" || value == "cc" {
					return "cc", nil
				}
				return value, nil
			}
		}
	}

	return "cc", nil
}

// BuildTmuxSessionConfig builds a tmux session configuration from worktree information.
//
// @MX:NOTE: SPEC-WORKTREE-002 integration function - tmux config builder
// @MX:SPEC: SPEC-WORKTREE-002
func BuildTmuxSessionConfig(projectName, specID, worktreePath, projectRoot string) (*TmuxSessionConfig, error) {
	activeMode, err := GetActiveMode(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("get active mode: %w", err)
	}

	cfg := &TmuxSessionConfig{
		ProjectName:  projectName,
		SpecID:       specID,
		WorktreePath: worktreePath,
		ActiveMode:   activeMode,
		GLMEnvVars:   make(map[string]string),
	}

	// R5.2-5.3: Set environment variables in GLM/CG mode
	if activeMode == "glm" || activeMode == "cg" {
		// Load GLM environment variables from ~/.moai/.env.glm
		homeDir, _ := os.UserHomeDir()
		glmEnvPath := filepath.Join(homeDir, ".moai", ".env.glm")
		if data, err := os.ReadFile(glmEnvPath); err == nil {
			// Parse KEY=VALUE lines
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					// Only include ANTHROPIC_* vars
					if strings.HasPrefix(key, "ANTHROPIC_") {
						cfg.GLMEnvVars[key] = value
					}
				}
			}
		}
		// Fallback to current environment if .env.glm doesn't have them
		if cfg.GLMEnvVars[config.EnvAnthropicDefaultHaikuModel] == "" {
			cfg.GLMEnvVars[config.EnvAnthropicDefaultHaikuModel] = os.Getenv(config.EnvAnthropicDefaultHaikuModel)
		}
		if cfg.GLMEnvVars[config.EnvAnthropicDefaultSonnetModel] == "" {
			cfg.GLMEnvVars[config.EnvAnthropicDefaultSonnetModel] = os.Getenv(config.EnvAnthropicDefaultSonnetModel)
		}
		if cfg.GLMEnvVars[config.EnvAnthropicDefaultOpusModel] == "" {
			cfg.GLMEnvVars[config.EnvAnthropicDefaultOpusModel] = os.Getenv(config.EnvAnthropicDefaultOpusModel)
		}
	}

	return cfg, nil
}
