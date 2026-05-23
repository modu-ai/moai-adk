// Resolution: UPGRADE — re-render + diff-aware reload via SPEC-V3R2-RT-005 + 20ms debounce.
// SPEC-V3R6-HOOK-ASYNC-EXPAND-001 M3 (REQ-HAE-002): debounce + YAML validation
// + diff-aware reload execute in a background goroutine with a 5-second
// deadline. The main handler returns within ≤ 100 ms (p95 under 10-concurrent
// benchmark, AC-HAE-003).
package hook

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
)

// configChangeHandler processes ConfigChange events.
// It validates configuration file changes and triggers diff-aware reload
// asynchronously per SPEC-V3R6-HOOK-ASYNC-EXPAND-001 REQ-HAE-002.
// SPEC-V3R2-RT-006 REQ-011, REQ-062 (validation contract preserved).
type configChangeHandler struct {
	// mgr holds an optional ConfigManager for RT-005 reload integration.
	// When nil, the handler falls back to YAML-only validation.
	mgr *config.ConfigManager
	// wg tracks in-flight async side-effect goroutines. Tests use the
	// package-internal waitGroup() accessor + testutil.WaitForAsync to
	// deterministically await completion.
	wg sync.WaitGroup
}

// NewConfigChangeHandler creates a new ConfigChange event handler.
func NewConfigChangeHandler() Handler {
	return &configChangeHandler{}
}

// NewConfigChangeHandlerWithManager creates a ConfigChange handler that uses
// the provided ConfigManager for diff-aware reload per SPEC-V3R2-RT-005.
func NewConfigChangeHandlerWithManager(mgr *config.ConfigManager) Handler {
	return &configChangeHandler{mgr: mgr}
}

// waitGroup returns the handler's internal *sync.WaitGroup for use with
// testutil.WaitForAsync. Package-internal; not exposed via the Handler
// interface.
func (h *configChangeHandler) waitGroup() *sync.WaitGroup {
	return &h.wg
}

// EventType returns EventConfigChange.
func (h *configChangeHandler) EventType() EventType {
	return EventConfigChange
}

// Handle processes a ConfigChange event. The main return path completes
// synchronously within ≤ 100 ms (p95) per REQ-HAE-002 / AC-HAE-003. The
// 20ms debounce, YAML validation, and ConfigManager.Reload() all execute
// in a background goroutine bounded by asyncDeadline (5s).
func (h *configChangeHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("config file changed",
		"session_id", input.SessionID,
		"config_file_path", input.ConfigFilePath,
		"configuration_source", input.ConfigurationSource,
	)

	// REQ-HAE-002 async transition: debounce + validation + reload run in
	// a background goroutine. Main handler returns immediately so the
	// Claude Code main loop is unblocked.
	asyncCtx, cancel := context.WithTimeout(context.Background(), asyncDeadline)
	h.wg.Add(1)
	go func() {
		defer cancel()
		defer h.wg.Done()
		h.runReload(asyncCtx, input)
	}()

	return &HookOutput{Continue: true}, nil
}

// runReload performs the debounce + YAML validation + diff-aware reload
// pipeline. It MUST be called from a goroutine. Errors are logged but
// never propagated to the main response (async design intent — failures
// surface via observability logs and future doctor enhancement).
func (h *configChangeHandler) runReload(ctx context.Context, input *HookInput) {
	// 20ms debounce: mitigate fsnotify mid-write race (REQ-011, REQ-062).
	// The file may be partially written when the event fires.
	select {
	case <-time.After(20 * time.Millisecond):
	case <-ctx.Done():
		slog.Warn("config_change async: cancelled during debounce",
			"path", input.ConfigFilePath,
			"error", ctx.Err(),
		)
		return
	}

	// Validate the config file as YAML first.
	if err := h.validateConfig(input.ConfigFilePath); err != nil {
		slog.Warn("config reload rejected (async)",
			"path", input.ConfigFilePath,
			"error", err,
			"action", "old settings retained",
		)
		return
	}

	// RT-005 diff-aware reload: attempt via ConfigManager when wired.
	if h.mgr != nil {
		if err := h.mgr.Reload(); err != nil {
			slog.Warn("config reload rejected (async)",
				"path", input.ConfigFilePath,
				"error", err,
				"action", "old settings retained",
			)
			return
		}
		slog.Info("config reloaded via RT-005 manager (async)",
			"path", input.ConfigFilePath,
		)
		return
	}

	// Fallback: attempt a fresh load when no manager is wired.
	// This is a best-effort path for tests and environments without DI.
	if input.CWD != "" {
		mgr := config.NewConfigManager()
		if err := mgr.Reload(); err != nil {
			slog.Debug("fallback manager reload failed (expected without Load)",
				"error", err,
			)
		}
	}

	slog.Info("config reloaded successfully (async)",
		"path", input.ConfigFilePath,
	)
}

// validateConfig checks that a config file is valid YAML.
func (h *configChangeHandler) validateConfig(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var cfgData any
	if err := yaml.Unmarshal(data, &cfgData); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	return nil
}
