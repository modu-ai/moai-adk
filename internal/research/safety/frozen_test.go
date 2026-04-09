package safety

import (
	"testing"
)

// TestFrozenGuard_IsFrozen은 FrozenGuard의 경로 동결 판정 로직을 검증한다.
func TestFrozenGuard_IsFrozen(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "기본 동결 경로: moai-constitution.md",
			path: ".claude/rules/moai/core/moai-constitution.md",
			want: true,
		},
		{
			name: "기본 동결 경로: agency constitution.md",
			path: ".claude/rules/agency/constitution.md",
			want: true,
		},
		{
			name: "기본 동결 경로: researcher.md",
			path: ".claude/agents/moai/researcher.md",
			want: true,
		},
		{
			name: "기본 동결 경로: research config.yaml",
			path: ".moai/research/config.yaml",
			want: true,
		},
		{
			name: "동결되지 않은 경로",
			path: ".claude/agents/moai/expert-backend.md",
			want: false,
		},
		{
			name: "빈 경로",
			path: "",
			want: false,
		},
		{
			name: "부분 일치: 절대 경로에 동결 경로 포함",
			path: "/project/.claude/rules/moai/core/moai-constitution.md",
			want: true,
		},
	}

	g := NewFrozenGuard()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.IsFrozen(tt.path)
			if got != tt.want {
				t.Errorf("IsFrozen(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestFrozenGuard_ValidateWrite는 동결된 경로에 대한 쓰기 검증을 테스트한다.
func TestFrozenGuard_ValidateWrite(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "동결 경로에 쓰기 시도 → 에러",
			path:    ".claude/rules/moai/core/moai-constitution.md",
			wantErr: true,
		},
		{
			name:    "허용 경로에 쓰기 시도 → nil",
			path:    ".claude/agents/moai/expert-backend.md",
			wantErr: false,
		},
		{
			name:    "다른 동결 경로에 쓰기 시도 → 에러",
			path:    ".moai/research/config.yaml",
			wantErr: true,
		},
	}

	g := NewFrozenGuard()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := g.ValidateWrite(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWrite(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

// TestFrozenGuardWithPaths는 커스텀 동결 경로가 올바르게 적용되는지 검증한다.
func TestFrozenGuardWithPaths(t *testing.T) {
	custom := []string{"custom/frozen.md", "another/frozen.yaml"}
	g := NewFrozenGuardWithPaths(custom)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "커스텀 동결 경로 1",
			path: "custom/frozen.md",
			want: true,
		},
		{
			name: "커스텀 동결 경로 2",
			path: "another/frozen.yaml",
			want: true,
		},
		{
			name: "기본 동결 경로는 포함되지 않음",
			path: ".claude/rules/moai/core/moai-constitution.md",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := g.IsFrozen(tt.path)
			if got != tt.want {
				t.Errorf("IsFrozen(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
