package tmux

import (
	"context"
	"fmt"
	"log/slog"
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
	if _, err := m.run(ctx, "tmux", "new-session", "-d", "-s", cfg.Name); err != nil {
		return nil, fmt.Errorf("create session %q: %w", cfg.Name, err)
	}

	m.logger.Debug("tmux session created", "name", cfg.Name)

	// Step 2: Send command to the first pane.
	if err := m.sendKeys(ctx, cfg.Name, 0, cfg.Panes[0].Command); err != nil {
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

// sendKeys sends a command string to a specific pane in a session.
func (m *DefaultSessionManager) sendKeys(ctx context.Context, session string, paneIndex int, command string) error {
	target := fmt.Sprintf("%s:0.%d", session, paneIndex)
	_, err := m.run(ctx, "tmux", "send-keys", "-t", target, command, "Enter")
	return err
}
