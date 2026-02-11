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
		binaryVersion string
		wantVersion   string
		wantAvailable bool
		wantUpdate    bool
		wantLatest    string
		wantErr       bool
	}{
		{
			name: "valid config with version, same as binary",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".moai", "config", "sections")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "system.yaml")
				content := []byte("moai:\n  version: 1.14.0\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			binaryVersion: "v1.14.0",
			wantVersion:   "1.14.0",
			wantAvailable: true,
			wantUpdate:    false,
		},
		{
			name: "binary newer than template",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".moai", "config", "sections")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "system.yaml")
				content := []byte("moai:\n  version: v2.0.0\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			binaryVersion: "v2.0.1",
			wantVersion:   "2.0.0",
			wantAvailable: true,
			wantUpdate:    true,
			wantLatest:    "2.0.1",
		},
		{
			name: "valid config with v prefix",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".moai", "config", "sections")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "system.yaml")
				content := []byte("moai:\n  version: v2.0.0\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			binaryVersion: "v2.0.0",
			wantVersion:   "2.0.0",
			wantAvailable: true,
			wantUpdate:    false,
		},
		{
			name: "no config file falls back to binary version",
			setupConfig: func(t *testing.T) string {
				return t.TempDir()
			},
			binaryVersion: "v2.0.0",
			wantVersion:   "2.0.0",
			wantAvailable: true,
			wantUpdate:    false,
		},
		{
			name: "empty version falls back to binary version",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".moai", "config", "sections")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "system.yaml")
				content := []byte("moai:\n  version: ''\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			binaryVersion: "v2.0.0",
			wantVersion:   "2.0.0",
			wantAvailable: true,
			wantUpdate:    false,
		},
		{
			name: "no config and no binary version",
			setupConfig: func(t *testing.T) string {
				return t.TempDir()
			},
			binaryVersion: "",
			wantAvailable: false,
		},
		{
			name: "no binary version provided",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				configDir := filepath.Join(dir, ".moai", "config", "sections")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				configPath := filepath.Join(configDir, "system.yaml")
				content := []byte("moai:\n  version: v2.0.0\n")
				if err := os.WriteFile(configPath, content, 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			binaryVersion: "",
			wantVersion:   "2.0.0",
			wantAvailable: true,
			wantUpdate:    false,
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
			v := NewVersionCollector(tt.binaryVersion)
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

			if got.UpdateAvailable != tt.wantUpdate {
				t.Errorf("CheckUpdate() UpdateAvailable = %v, want %v", got.UpdateAvailable, tt.wantUpdate)
			}

			if tt.wantLatest != "" && got.Latest != tt.wantLatest {
				t.Errorf("CheckUpdate() Latest = %v, want %v", got.Latest, tt.wantLatest)
			}
		})
	}
}

func TestVersionCollector_PrefersTemplateVersion(t *testing.T) {
	// Reproduction test: when moai.template_version differs from moai.version,
	// the collector should use moai.template_version (updated by moai update)
	// rather than moai.version (only set during moai init).
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "system.yaml")
	// Simulate a project initialized at v0.40.1, then updated to v2.2.1
	content := []byte("moai:\n  version: 0.40.1\n  template_version: 2.2.1\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	v := NewVersionCollector("v2.2.1")
	got, err := v.CheckUpdate(context.Background())
	if err != nil {
		t.Fatalf("CheckUpdate() error = %v", err)
	}

	// Should read template_version (2.2.1), not moai.version (0.40.1)
	if got.Current != "2.2.1" {
		t.Errorf("CheckUpdate() Current = %q, want %q (should prefer moai.template_version)", got.Current, "2.2.1")
	}
	// Binary matches template_version, so no update should be available
	if got.UpdateAvailable {
		t.Errorf("CheckUpdate() UpdateAvailable = true, want false (binary matches template_version)")
	}
}

func TestVersionCollector_FallbackToMoaiVersion(t *testing.T) {
	// When moai.template_version is missing, fall back to moai.version
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "system.yaml")
	content := []byte("moai:\n  version: 1.5.0\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	v := NewVersionCollector("v1.5.0")
	got, err := v.CheckUpdate(context.Background())
	if err != nil {
		t.Fatalf("CheckUpdate() error = %v", err)
	}

	if got.Current != "1.5.0" {
		t.Errorf("CheckUpdate() Current = %q, want %q", got.Current, "1.5.0")
	}
	if got.UpdateAvailable {
		t.Errorf("CheckUpdate() UpdateAvailable = true, want false")
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
