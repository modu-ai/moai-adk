package statusline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionCollector_CheckUpdate(t *testing.T) {
	tests := []struct {
		name          string
		setupConfig   func(t *testing.T) string
		wantVersion   string
		wantAvailable bool
		wantErr       bool
	}{
		{
			name: "valid config with version",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".moai", "config")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "config.yaml")
				content := []byte("moai:\n  version: 1.14.0\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantVersion:   "1.14.0",
			wantAvailable: true,
		},
		{
			name: "valid config with v prefix",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".moai", "config")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "config.yaml")
				content := []byte("moai:\n  version: v2.0.0\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantVersion:   "2.0.0",
			wantAvailable: true,
		},
		{
			name: "no config file",
			setupConfig: func(t *testing.T) string {
				return t.TempDir()
			},
			wantAvailable: false,
		},
		{
			name: "empty version",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".moai", "config")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "config.yaml")
				content := []byte("moai:\n  version: ''\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to test directory
			testDir := tt.setupConfig(t)
			originalDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(originalDir) }()
			if err := os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}

			// Clear any cached state by creating a new collector
			v := NewVersionCollector()
			ctx := context.Background()

			got, err := v.CheckUpdate(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Available != tt.wantAvailable {
				t.Errorf("CheckUpdate() Available = %v, want %v", got.Available, tt.wantAvailable)
			}

			if tt.wantVersion != "" && got.Current != tt.wantVersion {
				t.Errorf("CheckUpdate() Current = %v, want %v", got.Current, tt.wantVersion)
			}
		})
	}
}

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"v1.14.0", "1.14.0"},
		{"1.14.0", "1.14.0"},
		{"v2.0.0", "2.0.0"},
		{"2.0.0", "2.0.0"},
		{"v", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := formatVersion(tt.input); got != tt.want {
				t.Errorf("formatVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
