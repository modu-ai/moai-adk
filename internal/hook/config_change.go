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
	"path/filepath"
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

// waitGroup returns the handler's internal *sync.WaitGroup for use with
// testutil.WaitForAsync. Package-internal; not exposed via the Handler
// interface.
func (h *configChangeHandler) waitGroup() *sync.WaitGroup {
	return &h.wg
}

// joinAsync implements asyncJoiner: it waits for the handler's in-flight
// side-effect goroutine, bounded by timeout, and reports whether the budget
// was exhausted.
//
// Handle deliberately returns before that goroutine finishes, which is correct
// while a long-lived process remains to run it. A `moai hook <event>` process
// is not long-lived: it exits as soon as Dispatch returns, and the goroutine
// dies mid-debounce with its verdict unwritten. This method is the barrier the
// one-shot entrypoint crosses so the verdict actually gets produced; the
// teardown caller is registry.Shutdown.
func (h *configChangeHandler) joinAsync(timeout time.Duration) error {
	done := make(chan struct{})
	// The WaitGroup is joined on a helper goroutine rather than inline,
	// because sync.WaitGroup.Wait cannot itself be bounded. On the timeout
	// branch this helper is abandoned, which is safe: it holds no lock, writes
	// nothing, and the process is exiting.
	go func() {
		h.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("config_change: async side effect still running after %s", timeout)
	}
}

// EventType returns EventConfigChange.
func (h *configChangeHandler) EventType() EventType {
	return EventConfigChange
}

// Handle processes a ConfigChange event. The main return path completes
// synchronously within ≤ 100 ms (p95) per REQ-HAE-002 / AC-HAE-003. The
// 20ms debounce, YAML validation, and ConfigManager.Reload() are all
// DISPATCHED to a background goroutine carrying an asyncDeadline (5s) ctx.
// Dispatched, not delivered: see the spawn-site note below.
func (h *configChangeHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	// Official stdin field is config_source; configuration_source is the
	// legacy MoAI field name (kept as fallback for old payloads).
	configSource := input.ConfigSource
	if configSource == "" {
		configSource = input.ConfigurationSource
	}

	slog.Info("config file changed",
		"session_id", input.SessionID,
		"config_file_path", input.ConfigFilePath,
		"config_source", configSource,
	)

	// REQ-HAE-002 async transition: debounce + validation + reload run in
	// a background goroutine. Main handler returns immediately so the
	// Claude Code main loop is unblocked.
	//
	// asyncDeadline is not the outcome that decides this goroutine's fate.
	// `moai hook` is a one-shot process: it exits once the main handler
	// returns, and exit kills the goroutine wherever it has reached — the
	// deadline only ever fires when the process outlives it, which the hook
	// process does not. Compounding it here: the 20ms debounce runs BEFORE
	// any work, so this goroutine is still sleeping when the process exits.
	// Same shape as file_changed.go; see the longer note there.
	asyncCtx, cancel := context.WithTimeout(context.Background(), asyncDeadline)
	h.wg.Add(1)
	go func() {
		defer cancel()
		defer h.wg.Done()
		h.runReload(asyncCtx, input)
	}()

	// Empty output = continue (official spec: "continue": false is the only
	// meaningful value; the affirmative case omits the field entirely).
	return &HookOutput{}, nil
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
		appendConfigChangeAudit(input, configChangeResultRejected, err.Error())
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
			appendConfigChangeAudit(input, configChangeResultRejected, err.Error())
			return
		}
		slog.Info("config reloaded via RT-005 manager (async)",
			"path", input.ConfigFilePath,
		)
		appendConfigChangeAudit(input, configChangeResultReloaded, "rt005-manager")
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
	appendConfigChangeAudit(input, configChangeResultReloaded, "fallback")
}

// configChangeAuditRelPath is the durable record of what the async reload
// pipeline actually decided, relative to the project root.
//
// The slog calls in runReload cannot serve as that record on the path that
// matters. resolveLoggingDecision (internal/cli/logging.go) routes EVERY record
// emitted under `moai hook` to io.Discard, unconditionally — stdout carries the
// hook's JSON contract and stderr is read by the Claude Code runtime, so a
// stray record would corrupt the exchange, and MOAI_LOG_LEVEL does not re-open
// the carve-out. A validation verdict reached inside a one-shot hook process is
// therefore invisible unless it is written somewhere of its own.
//
// The file is that somewhere, modeled on the sibling advisory logs
// (branchGuardAuditRelPath, preEditAdvisoryLogRelPath).
const configChangeAuditRelPath = ".moai/logs/config-change-audit.log"

// The two terminal outcomes of the async reload pipeline, recorded verbatim in
// the audit line's result field so a reader can grep one without matching the
// other.
const (
	configChangeResultRejected = "rejected"
	configChangeResultReloaded = "reloaded"
)

// appendConfigChangeAudit appends one structured line recording how the async
// reload pipeline ended. Both outcomes are recorded, not only the rejection:
// an audit trail that logs failures alone cannot distinguish "the config was
// fine" from "the check never ran", which is precisely the ambiguity this file
// exists to remove.
//
// Every error is swallowed. The audit trail is observability, and losing it
// must never disturb the hook's own outcome — the handler has already decided
// by the time this is called.
func appendConfigChangeAudit(input *HookInput, result, detail string) {
	projectDir := resolveProjectRootFromInputOrEnv(input, "config_change")
	if projectDir == "" {
		return
	}
	sessionID := ""
	configPath := ""
	source := ""
	if input != nil {
		sessionID = input.SessionID
		configPath = input.ConfigFilePath
		source = input.ConfigSource
		if source == "" {
			source = input.ConfigurationSource
		}
	}

	entry := fmt.Sprintf("[%s] session=%s path=%s source=%s result=%s detail=%q\n",
		time.Now().UTC().Format(time.RFC3339), sessionID, configPath, source, result, detail)
	logPath := filepath.Join(projectDir, configChangeAuditRelPath)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(entry)
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
