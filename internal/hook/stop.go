package hook

import (
	"context"
	"log/slog"
	"strings"
)

// defaultCompletionMarkers는 기본 완료 마커 목록.
// Claude가 작업 완료 시 출력에 포함하는 마커.
var defaultCompletionMarkers = []string{
	"<moai>DONE</moai>",
	"<moai>COMPLETE</moai>",
}

// stopHandler processes Stop events.
// It performs graceful shutdown, saves in-progress work state, and preserves
// loop controller (Ralph) state (REQ-HOOK-035). Always returns "allow".
type stopHandler struct {
	// completionMarkers는 ToolOutput에서 감지할 완료 마커 목록.
	completionMarkers []string
}

// NewStopHandler creates a new Stop event handler.
func NewStopHandler() Handler {
	return &stopHandler{completionMarkers: defaultCompletionMarkers}
}

// NewStopHandlerWithMarkers는 커스텀 완료 마커를 사용하는 Stop 이벤트 핸들러를 생성.
// markers가 nil이거나 빈 슬라이스면 완료 마커 감지를 건너뜀.
func NewStopHandlerWithMarkers(markers []string) Handler {
	return &stopHandler{completionMarkers: markers}
}

// EventType returns EventStop.
func (h *stopHandler) EventType() EventType {
	return EventStop
}

// Handle processes a Stop event. It logs the stop request, preserves
// any active state, and returns an appropriate response.
//
// Per Claude Code protocol:
// - Return empty JSON {} to allow Claude to stop
// - Return {"decision": "block", "reason": "..."} to keep Claude working
// - Check stop_hook_active to prevent infinite loops
//
// Errors are non-blocking: the handler logs warnings and returns empty output.
func (h *stopHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("stop requested",
		"session_id", input.SessionID,
		"stop_hook_active", input.StopHookActive,
	)

	// IMPORTANT: Prevent infinite loop per Claude Code protocol
	// If stop_hook_active is true, Claude is already continuing due to a previous
	// stop hook decision. Allow Claude to stop to prevent infinite loops.
	if input.StopHookActive {
		slog.Debug("stop_hook_active is true, allowing Claude to stop")
		return &HookOutput{}, nil
	}

	// ToolOutput에서 완료 마커 감지 (관찰 전용, 절대 블록하지 않음)
	if len(input.ToolOutput) > 0 && len(h.completionMarkers) > 0 {
		output := string(input.ToolOutput)
		for _, marker := range h.completionMarkers {
			if strings.Contains(output, marker) {
				slog.Info("완료 마커 감지됨",
					"marker", marker,
					"session_id", input.SessionID,
				)
				break
			}
		}
	}

	// Stop hooks use top-level decision/reason fields per Claude Code protocol
	// Return empty JSON {} to allow Claude to stop (default behavior)
	// To keep Claude working, return: {"decision": "block", "reason": "..."}
	return &HookOutput{}, nil
}
