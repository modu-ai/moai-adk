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
		t.Fatal("NewStopHandlerWithMarkers가 nil 반환")
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
		// wantAllow: 항상 true (관찰 전용, 절대 블록 안 함)
		wantAllow bool
	}{
		{
			name:       "DONE 마커 감지됨",
			markers:    defaultCompletionMarkers,
			toolOutput: json.RawMessage(`"작업 완료 <moai>DONE</moai>"`),
			wantAllow:  true,
		},
		{
			name:       "COMPLETE 마커 감지됨",
			markers:    defaultCompletionMarkers,
			toolOutput: json.RawMessage(`"<moai>COMPLETE</moai>"`),
			wantAllow:  true,
		},
		{
			name:       "마커 없는 output은 그냥 통과",
			markers:    defaultCompletionMarkers,
			toolOutput: json.RawMessage(`"일반 작업 출력"`),
			wantAllow:  true,
		},
		{
			name:       "ToolOutput이 비어있으면 감지 건너뜀",
			markers:    defaultCompletionMarkers,
			toolOutput: nil,
			wantAllow:  true,
		},
		{
			name:       "커스텀 마커 감지됨",
			markers:    []string{"<done/>"},
			toolOutput: json.RawMessage(`"작업 완료 <done/>"`),
			wantAllow:  true,
		},
		{
			name:       "마커 목록이 빈 슬라이스면 감지 건너뜀",
			markers:    []string{},
			toolOutput: json.RawMessage(`"<moai>DONE</moai>"`),
			wantAllow:  true,
		},
		{
			name:       "마커 목록이 nil이면 감지 건너뜀",
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
				t.Fatalf("Handle() 오류 (관찰 전용이어야 함): %v", err)
			}
			if got == nil {
				t.Fatal("nil output 반환")
			}

			// 관찰 전용: Stop hook은 항상 비어있는 HookOutput을 반환 (allow)
			if got.Decision != "" {
				t.Errorf("Decision = %q, want empty (마커 감지는 블록하지 않음)", got.Decision)
			}
			if got.HookSpecificOutput != nil {
				t.Error("Stop hook은 HookSpecificOutput을 설정하면 안 됨")
			}
		})
	}
}

func TestStopHandler_DefaultMarkers_AreSet(t *testing.T) {
	t.Parallel()

	// NewStopHandler가 기본 마커를 포함하는지 확인
	h := NewStopHandler()
	impl, ok := h.(*stopHandler)
	if !ok {
		t.Fatal("핸들러가 *stopHandler가 아님")
	}

	if len(impl.completionMarkers) != 2 {
		t.Errorf("기본 마커 수 = %d, want 2", len(impl.completionMarkers))
	}

	markerSet := make(map[string]bool, len(impl.completionMarkers))
	for _, m := range impl.completionMarkers {
		markerSet[m] = true
	}

	if !markerSet["<moai>DONE</moai>"] {
		t.Error("기본 마커에 <moai>DONE</moai>가 없음")
	}
	if !markerSet["<moai>COMPLETE</moai>"] {
		t.Error("기본 마커에 <moai>COMPLETE</moai>가 없음")
	}
}

func TestStopHandler_StopHookActive_SkipsMarkerCheck(t *testing.T) {
	t.Parallel()

	// StopHookActive가 true면 마커 체크 전에 early return
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
		t.Fatalf("Handle() 오류: %v", err)
	}
	if got == nil {
		t.Fatal("nil output 반환")
	}

	// StopHookActive이면 무조건 빈 output (무한루프 방지)
	if got.Decision != "" {
		t.Errorf("Decision = %q, want empty", got.Decision)
	}
}
