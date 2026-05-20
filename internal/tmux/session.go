package tmux

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultMaxVisible = 3

// PaneConfig describes a single tmux pane.
type PaneConfig struct {
	// SpecID identifies the SPEC this pane is for (e.g., "SPEC-ISSUE-123").
	SpecID string

	// Command is the shell command to execute in this pane.
	Command string
}

// SessionConfig describes the tmux session to create.
type SessionConfig struct {
	// Name is the session name (e.g., "github-issues-2026-02-16-18-30").
	Name string

	// Panes lists the panes to create in the session.
	Panes []PaneConfig

	// MaxVisible is the maximum number of panes using vertical splits.
	// Additional panes use horizontal splits. Zero uses default (3).
	MaxVisible int
}

// SessionResult holds the outcome of session creation.
type SessionResult struct {
	// SessionName is the name of the created session.
	SessionName string

	// PaneCount is the number of panes created.
	PaneCount int

	// Attached indicates whether the session is attached to the terminal.
	Attached bool
}

// SessionManager creates and manages tmux sessions.
type SessionManager interface {
	// Create creates a new tmux session with the specified configuration.
	Create(ctx context.Context, cfg *SessionConfig) (*SessionResult, error)

	// InjectEnv sets environment variables in the current tmux session.
	InjectEnv(ctx context.Context, vars map[string]string) error

	// ClearEnv removes environment variables from the current tmux session.
	ClearEnv(ctx context.Context, vars []string) error
}

// DefaultSessionManager implements SessionManager using tmux commands.
type DefaultSessionManager struct {
	run    RunFunc
	logger *slog.Logger
}

// Compile-time interface compliance check.
var _ SessionManager = (*DefaultSessionManager)(nil)

// SessionManagerOption configures a DefaultSessionManager.
type SessionManagerOption func(*DefaultSessionManager)

// WithSessionRunFunc sets a custom command runner (used for testing).
func WithSessionRunFunc(fn RunFunc) SessionManagerOption {
	return func(m *DefaultSessionManager) {
		m.run = fn
	}
}

// WithSessionLogger sets the logger for the session manager.
func WithSessionLogger(l *slog.Logger) SessionManagerOption {
	return func(m *DefaultSessionManager) {
		m.logger = l
	}
}

// NewSessionManager creates a new DefaultSessionManager.
func NewSessionManager(opts ...SessionManagerOption) *DefaultSessionManager {
	m := &DefaultSessionManager{
		run:    defaultRun,
		logger: slog.Default().With("module", "tmux.session"),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Create creates a new tmux session with the specified pane configuration.
//
// Layout strategy:
//   - First pane: created with the session via new-session.
//   - Panes 2 to MaxVisible: added via vertical splits (split-window -v).
//   - Panes beyond MaxVisible: added via horizontal splits (split-window -h).
//   - After all panes are created, focus returns to pane 0.
func (m *DefaultSessionManager) Create(ctx context.Context, cfg *SessionConfig) (*SessionResult, error) {
	if len(cfg.Panes) == 0 {
		return nil, ErrNoPanes
	}

	maxVisible := cfg.MaxVisible
	if maxVisible <= 0 {
		maxVisible = defaultMaxVisible
	}

	// Step 1: Create the session with the first pane.
	// Use -c flag to set starting directory when the command contains a cd target.
	startDir := extractCdTarget(cfg.Panes[0].Command)
	args := []string{"new-session", "-d", "-s", cfg.Name}
	if startDir != "" {
		args = append(args, "-c", startDir)
	}
	if _, err := m.run(ctx, "tmux", args...); err != nil {
		return nil, fmt.Errorf("create session %q: %w", cfg.Name, err)
	}

	m.logger.Debug("tmux session created", "name", cfg.Name)

	// Step 2: Wait for shell to initialize, then send command.
	// tmux new-session returns before the shell inside the pane is ready
	// to accept input. A brief pause prevents the command from being lost.
	time.Sleep(500 * time.Millisecond)

	// When startDir was set via -c, the shell already starts in the
	// worktree directory. Strip the leading "cd <path> ; " so we only
	// send the actual command (e.g., "moai cc").
	cmd := cfg.Panes[0].Command
	if startDir != "" {
		cmd = stripCdPrefix(cmd)
	}
	if err := m.sendKeys(ctx, cfg.Name, 0, cmd); err != nil {
		m.logger.Warn("failed to send command to first pane",
			"session", cfg.Name,
			"error", err,
		)
	}

	// Step 3: Create additional panes.
	for i := 1; i < len(cfg.Panes); i++ {
		direction := "-v" // Vertical split.
		if i >= maxVisible {
			direction = "-h" // Horizontal split for overflow.
		}

		if _, err := m.run(ctx, "tmux", "split-window", direction, "-t", cfg.Name); err != nil {
			m.logger.Warn("failed to create pane",
				"session", cfg.Name,
				"pane_index", i,
				"error", err,
			)
			continue
		}

		if err := m.sendKeys(ctx, cfg.Name, i, cfg.Panes[i].Command); err != nil {
			m.logger.Warn("failed to send command to pane",
				"session", cfg.Name,
				"pane_index", i,
				"error", err,
			)
		}
	}

	// Step 4: Select the first pane and rebalance layout.
	_, _ = m.run(ctx, "tmux", "select-pane", "-t", fmt.Sprintf("%s:0.0", cfg.Name))
	_, _ = m.run(ctx, "tmux", "select-layout", "-t", cfg.Name, "tiled")

	m.logger.Info("tmux session ready",
		"name", cfg.Name,
		"panes", len(cfg.Panes),
	)

	return &SessionResult{
		SessionName: cfg.Name,
		PaneCount:   len(cfg.Panes),
		Attached:    false,
	}, nil
}

// InjectEnv injects environment variables into the current tmux session.
//
// SECURITY: this method passes the value as a positional argv to `tmux
// set-environment`, which means the value is visible to any local process
// that can read /proc/<pid>/cmdline (Linux) or `ps -ef` output (macOS).
// For credentials use InjectSensitiveEnv instead.
func (m *DefaultSessionManager) InjectEnv(ctx context.Context, vars map[string]string) error {
	for key, value := range vars {
		args := []string{"set-environment", key, value}
		if _, err := m.run(ctx, "tmux", args...); err != nil {
			return fmt.Errorf("set tmux env %s: %w", key, err)
		}
	}
	return nil
}

// sensitiveTempDir is the directory inside the user's home that holds the
// short-lived scripts used by InjectSensitiveEnv. It is created with mode
// 0o700 so that no other local user can list / open files inside it.
//
// @MX:NOTE: [AUTO] tied to SPEC-V3R5-SECURITY-CRIT-001 REQ-SEC-002-005 stale cleanup
// @MX:REASON: see cleanupStaleSensitiveTemp; best-effort cleanup runs on every call.
const sensitiveTempDir = ".moai/run"

// resolveSensitiveTempDir returns the absolute path of the per-user
// sensitive-temp directory (~/.moai/run/). Returns an empty string and an
// error if HOME cannot be resolved.
func resolveSensitiveTempDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, sensitiveTempDir), nil
}

// cleanupStaleSensitiveTemp removes leftover moai-tmux-* files older than 1
// hour from ~/.moai/run/. Best-effort: errors are silently dropped because
// the cleanup is opportunistic and must not break the injection path.
//
// SPEC-V3R5-SECURITY-CRIT-001 REQ-SEC-002-005 / plan-auditor SHOULD S1.
func cleanupStaleSensitiveTemp(dir string) {
	const maxAge = 1 * time.Hour
	cutoff := time.Now().Add(-maxAge)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return // dir does not exist yet (first call) or unreadable
	}
	for _, ent := range entries {
		if !strings.HasPrefix(ent.Name(), "moai-tmux-") {
			continue
		}
		info, err := ent.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, ent.Name()))
		}
	}
}

// InjectSensitiveEnv injects a single environment variable into the current
// tmux session without exposing its value through process argv. Used for
// credentials such as ANTHROPIC_AUTH_TOKEN where `tmux set-environment <k>
// <v>` would leak the value via /proc/<pid>/cmdline or `ps -ef`.
//
// Implementation: write a single-line tmux config (`set-environment <k>
// <v>`) to a 0o600 temp file in ~/.moai/run/, then invoke `tmux source-file
// <path>`. The value never appears in any process argv. The temp script is
// unlinked on every return path (success and error) so the secret never
// lingers on disk longer than one tmux source-file call.
//
// On any failure (mkdir, create temp, write, chmod, source-file), this
// method returns an error that wraps ErrTmuxSensitiveInjectFailed. Callers
// MUST treat this as a hard stop — they must NOT retry via InjectEnv on the
// failure path, because that would reintroduce the argv leak.
//
// SPEC-V3R5-SECURITY-CRIT-001 P0-2 (REQ-SEC-002-001..-007, AC-SEC-005/007).
//
// @MX:ANCHOR: [AUTO] InjectSensitiveEnv is the only safe entry point for tmux env values containing credentials
// @MX:REASON: SPEC-V3R5-SECURITY-CRIT-001 P0-2 (CWE-214). Diverging from this path = regression.
func (m *DefaultSessionManager) InjectSensitiveEnv(ctx context.Context, key, value string) error {
	if key == "" {
		return fmt.Errorf("%w: key cannot be empty", ErrTmuxSensitiveInjectFailed)
	}

	dir, err := resolveSensitiveTempDir()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTmuxSensitiveInjectFailed, err)
	}

	// Mkdir with 0o700 so the directory itself blocks other users from
	// listing the per-session temp scripts. MkdirAll is idempotent on
	// success and only widens permissions when re-creating, so call Chmod
	// to force-correct legacy 0o755 directories.
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("%w: mkdir %s: %v", ErrTmuxSensitiveInjectFailed, dir, err)
	}
	_ = os.Chmod(dir, 0o700) // best-effort tighten if pre-existing

	// Opportunistic cleanup of stale scripts from prior crashes.
	// SPEC-V3R5-SECURITY-CRIT-001 plan-auditor SHOULD S1.
	cleanupStaleSensitiveTemp(dir)

	// Create temp file with secure mode. CreateTemp uses 0o600 by default
	// since Go 1.18, but we Chmod immediately as a defense-in-depth measure
	// against umask oddities.
	tmpFile, err := os.CreateTemp(dir, "moai-tmux-*.sh")
	if err != nil {
		return fmt.Errorf("%w: create temp script: %v", ErrTmuxSensitiveInjectFailed, err)
	}
	tmpPath := tmpFile.Name()
	// Guarantee cleanup on every exit path. We capture err in a closure so
	// the deferred Remove always runs even when Close itself fails.
	defer func() {
		_ = tmpFile.Close()
		if rmErr := os.Remove(tmpPath); rmErr != nil && !os.IsNotExist(rmErr) {
			m.logger.Warn("could not remove sensitive temp script",
				"path", tmpPath,
				"error", rmErr,
			)
		}
	}()

	if err := os.Chmod(tmpPath, 0o600); err != nil {
		return fmt.Errorf("%w: chmod %s: %v", ErrTmuxSensitiveInjectFailed, tmpPath, err)
	}

	// Write a single tmux config statement. We deliberately keep this to
	// one line and do not interpolate the value into a shell command, so
	// even if the file is read mid-injection no shell metacharacters can
	// affect parsing.
	//
	// tmux source-file accepts `set-environment KEY VALUE` natively; the
	// value is parsed by tmux's own tokenizer, not by a shell.
	content := fmt.Sprintf("set-environment %s %s\n", key, escapeTmuxValue(value))
	if _, err := tmpFile.WriteString(content); err != nil {
		return fmt.Errorf("%w: write temp script: %v", ErrTmuxSensitiveInjectFailed, err)
	}
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("%w: sync temp script: %v", ErrTmuxSensitiveInjectFailed, err)
	}

	// Hand the script path (NOT the value) to tmux. Only the path appears
	// in argv; the value is read from disk by the tmux process itself.
	if _, err := m.run(ctx, "tmux", "source-file", tmpPath); err != nil {
		return fmt.Errorf("%w: source-file: %v", ErrTmuxSensitiveInjectFailed, err)
	}

	return nil
}

// escapeTmuxValue quotes a value for safe inclusion in a tmux config line.
// tmux uses single-quoted strings for literal values; only the single quote
// itself needs escaping by closing+escaping+reopening the literal.
func escapeTmuxValue(v string) string {
	// Wrap in double quotes; escape backslash and double-quote characters.
	// tmux accepts both single- and double-quoted strings; double quotes
	// keep parsing simpler when a value may itself contain a single quote.
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	return `"` + replacer.Replace(v) + `"`
}

// ClearEnv removes environment variables from the current tmux session.
func (m *DefaultSessionManager) ClearEnv(ctx context.Context, vars []string) error {
	for _, key := range vars {
		args := []string{"set-environment", "-u", key}
		if _, err := m.run(ctx, "tmux", args...); err != nil {
			return fmt.Errorf("unset tmux env %s: %w", key, err)
		}
	}
	return nil
}

// sendKeys sends a command string to a specific pane in a session.
func (m *DefaultSessionManager) sendKeys(ctx context.Context, session string, paneIndex int, command string) error {
	target := fmt.Sprintf("%s:0.%d", session, paneIndex)
	_, err := m.run(ctx, "tmux", "send-keys", "-t", target, command, "Enter")
	return err
}

// extractCdTarget parses "cd <path> ; <rest>" and returns <path>.
// Returns empty string if the command doesn't start with "cd ".
func extractCdTarget(cmd string) string {
	if !strings.HasPrefix(cmd, "cd ") {
		return ""
	}
	rest := cmd[3:]
	sep := strings.Index(rest, " ; ")
	if sep >= 0 {
		return rest[:sep]
	}
	return rest
}

// stripCdPrefix removes "cd <path> ; " prefix from a command.
// If no cd prefix is found, returns the original command.
func stripCdPrefix(cmd string) string {
	if !strings.HasPrefix(cmd, "cd ") {
		return cmd
	}
	sep := strings.Index(cmd, " ; ")
	if sep >= 0 {
		return cmd[sep+3:]
	}
	return ""
}
