package hook

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNewStopHandlerWithMarkers(t *testing.T) {
	t.Parallel()

	h := NewStopHandlerWithMarkers([]string{"<done/>"})
	if h == nil {
		t.Fatal("NewStopHandlerWithMarkers returned nil")
	}
	if h.EventType() != EventStop {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventStop)
	}
}

func TestStopHandler_CompletionMarkers_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		markers    []string
		toolOutput json.RawMessage
		// wantAllow: always true (observation-only, never blocks)
		wantAllow bool
	}{
		{
			name:       "DONE marker detected",
			markers:    defaultCompletionMarkers,
			toolOutput: json.RawMessage(`"task complete <moai>DONE</moai>"`),
			wantAllow:  true,
		},
		{
			name:       "COMPLETE marker detected",
			markers:    defaultCompletionMarkers,
			toolOutput: json.RawMessage(`"<moai>COMPLETE</moai>"`),
			wantAllow:  true,
		},
		{
			name:       "output without marker passes through",
			markers:    defaultCompletionMarkers,
			toolOutput: json.RawMessage(`"normal task output"`),
			wantAllow:  true,
		},
		{
			name:       "detection skipped when ToolOutput is empty",
			markers:    defaultCompletionMarkers,
			toolOutput: nil,
			wantAllow:  true,
		},
		{
			name:       "custom marker detected",
			markers:    []string{"<done/>"},
			toolOutput: json.RawMessage(`"task complete <done/>"`),
			wantAllow:  true,
		},
		{
			name:       "detection skipped when marker list is empty slice",
			markers:    []string{},
			toolOutput: json.RawMessage(`"<moai>DONE</moai>"`),
			wantAllow:  true,
		},
		{
			name:       "detection skipped when marker list is nil",
			markers:    nil,
			toolOutput: json.RawMessage(`"<moai>DONE</moai>"`),
			wantAllow:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewStopHandlerWithMarkers(tt.markers)
			ctx := context.Background()

			input := &HookInput{
				SessionID:  "sess-marker-test",
				CWD:        "/tmp",
				ToolOutput: tt.toolOutput,
			}

			got, err := h.Handle(ctx, input)

			if err != nil {
				t.Fatalf("Handle() returned error (should be observation-only): %v", err)
			}
			if got == nil {
				t.Fatal("returned nil output")
			} else if got.Decision != "" {
				// Observation-only: Stop hook always returns empty HookOutput (allow)
				t.Errorf("Decision = %q, want empty (marker detection must not block)", got.Decision)
			} else if got.HookSpecificOutput != nil {
				t.Error("Stop hook must not set HookSpecificOutput")
			}
		})
	}
}

func TestStopHandler_DefaultMarkers_AreSet(t *testing.T) {
	t.Parallel()

	// Verify that NewStopHandler includes the default markers
	h := NewStopHandler()
	impl, ok := h.(*stopHandler)
	if !ok {
		t.Fatal("handler is not *stopHandler")
	}

	if len(impl.completionMarkers) != 2 {
		t.Errorf("default marker count = %d, want 2", len(impl.completionMarkers))
	}

	markerSet := make(map[string]bool, len(impl.completionMarkers))
	for _, m := range impl.completionMarkers {
		markerSet[m] = true
	}

	if !markerSet["<moai>DONE</moai>"] {
		t.Error("default markers do not include <moai>DONE</moai>")
	}
	if !markerSet["<moai>COMPLETE</moai>"] {
		t.Error("default markers do not include <moai>COMPLETE</moai>")
	}
}

func TestStopHandler_StopHookActive_SkipsMarkerCheck(t *testing.T) {
	t.Parallel()

	// When StopHookActive is true, early return before marker check
	h := NewStopHandlerWithMarkers(defaultCompletionMarkers)
	ctx := context.Background()

	input := &HookInput{
		SessionID:      "sess-active",
		CWD:            "/tmp",
		StopHookActive: true,
		ToolOutput:     json.RawMessage(`"<moai>DONE</moai>"`),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if got == nil {
		t.Fatal("returned nil output")
	} else if got.Decision != "" {
		// When StopHookActive, always return empty output (prevent infinite loop)
		t.Errorf("Decision = %q, want empty", got.Decision)
	}
}
