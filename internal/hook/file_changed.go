// Resolution: UPGRADE — MX re-scan for 16 supported language extensions on FileChanged.
// SPEC-V3R6-HOOK-ASYNC-EXPAND-001 M2 (REQ-HAE-001): MX delta scan + sidecar
// update execute in a background goroutine with a 5-second deadline. The main
// handler returns within ≤ 100 ms (p95 under 10-concurrent benchmark, AC-HAE-002).
package hook

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// Supported language extensions for MX tag scanning.
var supportedExtensions = map[string]bool{
	".go":    true,
	".py":    true,
	".ts":    true,
	".js":    true,
	".rs":    true,
	".java":  true,
	".kt":    true,
	".cs":    true,
	".rb":    true,
	".php":   true,
	".ex":    true,
	".exs":   true,
	".cpp":   true,
	".cc":    true,
	".cxx":   true,
	".h":     true,
	".hpp":   true,
	".scala": true,
	".r":     true,
	".dart":  true,
	".swift": true,
}

// asyncDeadline is the canonical deadline for hook handler side-effect
// goroutines per SPEC-V3R6-HOOK-ASYNC-EXPAND-001 REQ-HAE-005. After this
// duration the goroutine MUST self-cancel via its context.
const asyncDeadline = 5 * time.Second

// fileChangedHandler processes FileChanged events.
// It scans changed files for MX tag deltas asynchronously.
type fileChangedHandler struct {
	// wg tracks in-flight async side-effect goroutines. Tests use the
	// package-internal waitGroup() accessor + testutil.WaitForAsync to
	// deterministically await completion. Production callers MUST NOT
	// block on this WaitGroup.
	wg sync.WaitGroup
}

// NewFileChangedHandler creates a new FileChanged event handler.
func NewFileChangedHandler() Handler {
	return &fileChangedHandler{}
}

// waitGroup returns the handler's internal *sync.WaitGroup for use with
// testutil.WaitForAsync. Package-internal; not exposed via the Handler
// interface. Tests obtain a typed reference via a package-internal
// type-assert or via the typed constructor.
func (h *fileChangedHandler) waitGroup() *sync.WaitGroup {
	return &h.wg
}

// EventType returns EventFileChanged.
func (h *fileChangedHandler) EventType() EventType {
	return EventFileChanged
}

// Handle processes a FileChanged event. It returns within ≤ 100 ms (p95
// under 10-concurrent benchmark per AC-HAE-002) by deferring the MX delta
// scan + sidecar update to a background goroutine bounded by asyncDeadline.
//
// REQ-HAE-001: main return path completes synchronously with HookOutput{}.
// All side-effects (scan, sidecar update) execute in the spawned goroutine.
func (h *fileChangedHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("file changed externally",
		"session_id", input.SessionID,
		"file_path", input.FilePath,
		"change_type", input.ChangeType,
	)

	// Skip deleted files (synchronous fast-path — no side-effect work).
	if input.ChangeType == "deleted" {
		return &HookOutput{}, nil
	}

	// Check if file has supported extension (synchronous fast-path).
	ext := strings.ToLower(filepath.Ext(input.FilePath))
	if !supportedExtensions[ext] {
		slog.Debug("unsupported file extension for MX scan",
			"path", input.FilePath,
			"ext", ext,
		)
		return &HookOutput{}, nil
	}

	// REQ-HAE-001 async transition: MX scan + sidecar update run in a
	// background goroutine with a 5-second deadline. The main handler
	// returns immediately so the Claude Code main loop is unblocked.
	//
	// context.Background() decouples from the request ctx — request
	// cancellation MUST NOT cancel the side-effect (it runs to completion
	// or the asyncDeadline expires).
	asyncCtx, cancel := context.WithTimeout(context.Background(), asyncDeadline)
	h.wg.Add(1)
	go func() {
		defer cancel()
		defer h.wg.Done()
		h.runMXScan(asyncCtx, input)
	}()

	return &HookOutput{}, nil
}

// runMXScan executes the MX tag scan + sidecar update for the changed file.
// It MUST be called from a goroutine; the deadline-aware ctx is honored by
// downstream MX scanner / manager calls when they support context.
//
// Errors are logged and swallowed — async side-effects never propagate
// failure to the main handler response (REQ-HAE-005 design intent).
func (h *fileChangedHandler) runMXScan(ctx context.Context, input *HookInput) {
	// Respect deadline cancellation defensively.
	select {
	case <-ctx.Done():
		slog.Warn("file_changed async: context cancelled before scan",
			"path", input.FilePath,
			"error", ctx.Err(),
		)
		return
	default:
	}

	// Scan file for MX tags.
	scanner := mx.NewScanner()
	tags, err := scanner.ScanFile(input.FilePath)
	if err != nil {
		slog.Warn("failed to scan file for MX tags",
			"path", input.FilePath,
			"error", err,
		)
		return
	}

	// Compare with existing sidecar tags and update index.
	// This tracks MX tag deltas across file edits for monitoring and validation.
	projectDir := input.CWD
	if projectDir != "" {
		stateDir := filepath.Join(projectDir, ".moai", "state")
		manager := mx.NewManager(stateDir)
		if _, uerr := manager.UpdateFile(input.FilePath, tags); uerr != nil {
			slog.Warn("failed to update MX sidecar",
				"path", input.FilePath,
				"error", uerr,
			)
		}
	}

	// Format summary message and log for observability.
	// Note: in async mode the SystemMessage is NOT returned to Claude Code
	// (REQ-HAE-001 design — main response is fixed empty payload). The
	// message is logged at info level for observability.
	if msg := h.formatTagDelta(input.FilePath, tags); msg != "" {
		slog.Info("mx tag delta (async)", "summary", msg)
	}
}

// formatTagDelta creates a summary message for MX tag changes.
func (h *fileChangedHandler) formatTagDelta(filePath string, tags []mx.Tag) string {
	if len(tags) == 0 {
		return ""
	}

	// Count tags by kind
	counts := make(map[mx.TagKind]int)
	for _, tag := range tags {
		counts[tag.Kind]++
	}

	var sb strings.Builder
	sb.WriteString(filePath)
	sb.WriteByte(':')
	for kind, count := range counts {
		fmt.Fprintf(&sb, " %s=%d", kind, count)
	}

	return "MX tag delta on " + sb.String()
}
