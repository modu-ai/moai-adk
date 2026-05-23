package hook

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolveMemoryDir(t *testing.T) {
	t.Parallel()

	homeDir := "/Users/goos"
	tests := []struct {
		name       string
		projectDir string
		wantSuffix string
	}{
		{
			name:       "simple path without dots",
			projectDir: "/Users/goos/MoAI/moai-adk-go",
			wantSuffix: filepath.Join(".claude", "projects", "-Users-goos-MoAI-moai-adk-go", "memory"),
		},
		{
			name:       "leading dot directory becomes double dash",
			projectDir: "/Users/goos/.moai/worktrees/foo",
			wantSuffix: filepath.Join(".claude", "projects", "-Users-goos--moai-worktrees-foo", "memory"),
		},
		{
			name:       "single dot directory only",
			projectDir: "/Users/goos/.claude",
			wantSuffix: filepath.Join(".claude", "projects", "-Users-goos--claude", "memory"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := resolveMemoryDir(homeDir, tt.projectDir)
			if err != nil {
				t.Fatalf("resolveMemoryDir(%q, %q) returned error: %v", homeDir, tt.projectDir, err)
			}
			want := filepath.Join(homeDir, tt.wantSuffix)
			if got != want {
				t.Errorf("resolveMemoryDir(%q, %q) = %q, want %q", homeDir, tt.projectDir, got, want)
			}
		})
	}
}

func TestResolveMemoryDir_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		homeDir    string
		projectDir string
		wantMsg    string
	}{
		{
			name:       "empty home directory",
			homeDir:    "",
			projectDir: "/Users/goos/MoAI/moai-adk-go",
			wantMsg:    "home directory is empty",
		},
		{
			name:       "empty project directory",
			homeDir:    "/Users/goos",
			projectDir: "",
			wantMsg:    "project directory is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := resolveMemoryDir(tt.homeDir, tt.projectDir)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantMsg)
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestProjectSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"posix simple", "/Users/goos/MoAI/moai-adk-go", "-Users-goos-MoAI-moai-adk-go"},
		{"posix dot dir", "/Users/goos/.moai/worktrees/foo", "-Users-goos--moai-worktrees-foo"},
		{"posix trailing slash collapsed", "/Users/goos/MoAI/", "-Users-goos-MoAI"},
		{"posix dot only", "/Users/goos/.claude", "-Users-goos--claude"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Run POSIX-shaped inputs on non-Windows hosts; on Windows the
			// `filepath.Clean` step would normalize separators and skew the
			// observed-from-macOS expected slugs.
			if runtime.GOOS == "windows" {
				t.Skip("POSIX-shaped path slugs are validated on non-Windows hosts only")
			}
			got := projectSlug(tt.in)
			if got != tt.want {
				t.Errorf("projectSlug(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
