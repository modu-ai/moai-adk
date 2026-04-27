package dtcg_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// TestDetectAliasCycle_NoCycle: 순환 참조 없는 정상 에일리어스 체인.
// [HARD]: A→B→C 형태의 정상 체인은 오류 없어야 함.
func TestDetectAliasCycle_NoCycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		graph map[string]string
	}{
		{
			name: "단일 에일리어스",
			graph: map[string]string{
				"color.primary": "color.blue-500",
			},
		},
		{
			name: "선형 체인 A→B→C",
			graph: map[string]string{
				"color.primary":  "color.blue-500",
				"color.blue-500": "color.blue-base",
			},
		},
		{
			name: "병렬 에일리어스",
			graph: map[string]string{
				"color.primary":   "color.blue-500",
				"color.secondary": "color.green-500",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := dtcg.DetectAliasCycle(tt.graph)
			if err != nil {
				t.Errorf("DetectAliasCycle(%v) = %v; 순환 없어야 함", tt.graph, err)
			}
		})
	}
}

// TestDetectAliasCycle_Cycle: 순환 참조 감지.
// [HARD]: A→B→C→A 순환 및 A→A 자기 참조 모두 감지해야 함.
func TestDetectAliasCycle_Cycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		graph map[string]string
	}{
		{
			name: "자기 참조 A→A",
			graph: map[string]string{
				"color.primary": "color.primary",
			},
		},
		{
			name: "2-노드 순환 A→B→A",
			graph: map[string]string{
				"color.primary":  "color.secondary",
				"color.secondary": "color.primary",
			},
		},
		{
			name: "3-노드 순환 A→B→C→A",
			graph: map[string]string{
				"color.a": "color.b",
				"color.b": "color.c",
				"color.c": "color.a",
			},
		},
		{
			name: "긴 체인 끝에 순환",
			graph: map[string]string{
				"token.1": "token.2",
				"token.2": "token.3",
				"token.3": "token.4",
				"token.4": "token.2", // 여기서 순환
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := dtcg.DetectAliasCycle(tt.graph)
			if err == nil {
				t.Errorf("DetectAliasCycle(%v) = nil; 순환 감지해야 함", tt.graph)
			}
		})
	}
}
