package safety

import (
	"testing"
)

// TestFrozenGuard_IsFrozen verifies the path freeze detection logic of FrozenGuard.
func TestFrozenGuard_IsFrozen(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "default frozen path: moai-constitution.md",
			path: ".claude/rules/moai/core/moai-constitution.md",
			want: true,
		},
		{
			name: "default frozen path: agency constitution.md",
			path: ".claude/rules/agency/constitution.md",
			want: true,
		},
		{
			name: "default frozen path: researcher.md",
			path: ".claude/agents/moai/researcher.md",
			want: true,
		},
		{
			name: "default frozen path: research config.yaml",
			path: ".moai/research/config.yaml",
			want: true,
		},
		{
			name: "non-frozen path",
			path: ".claude/agents/moai/expert-backend.md",
			want: false,
		},
		{
			name: "empty path",
			path: "",
			want: false,
		},
		{
			name: "partial match: absolute path containing frozen path",
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

// TestFrozenGuard_ValidateWrite verifies write validation against frozen paths.
func TestFrozenGuard_ValidateWrite(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "write to frozen path → error",
			path:    ".claude/rules/moai/core/moai-constitution.md",
			wantErr: true,
		},
		{
			name:    "write to allowed path → nil",
			path:    ".claude/agents/moai/expert-backend.md",
			wantErr: false,
		},
		{
			name:    "write to another frozen path → error",
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

// TestFrozenGuardWithPaths verifies that custom frozen paths are applied correctly.
func TestFrozenGuardWithPaths(t *testing.T) {
	custom := []string{"custom/frozen.md", "another/frozen.yaml"}
	g := NewFrozenGuardWithPaths(custom)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "custom frozen path 1",
			path: "custom/frozen.md",
			want: true,
		},
		{
			name: "custom frozen path 2",
			path: "another/frozen.yaml",
			want: true,
		},
		{
			name: "default frozen path is not included",
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
