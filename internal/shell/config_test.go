package shell

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetConfigFile(t *testing.T) {
	c := NewConfigurator()

	tests := []struct {
		name         string
		shell        ShellType
		preferLogin  bool
		wantSuffixes []string // Accept multiple valid suffixes per platform
	}{
		{
			name:         "zsh_prefer_login",
			shell:        ShellZsh,
			preferLogin:  true,
			wantSuffixes: []string{".zshenv"},
		},
		{
			name:         "zsh_no_prefer_login",
			shell:        ShellZsh,
			preferLogin:  false,
			wantSuffixes: []string{".zshrc"},
		},
		{
			name:         "bash_prefer_login",
			shell:        ShellBash,
			preferLogin:  true,
			wantSuffixes: []string{".profile", ".bash_profile"}, // macOS may use .bash_profile
		},
		{
			name:         "bash_no_prefer_login",
			shell:        ShellBash,
			preferLogin:  false,
			wantSuffixes: []string{".bashrc"},
		},
		{
			name:         "fish",
			shell:        ShellFish,
			preferLogin:  false,
			wantSuffixes: []string{"config.fish"},
		},
		{
			name:        "unknown_defaults_to_profile",
			shell:       ShellUnknown,
			preferLogin: true,
			// On Windows, ShellUnknown defaults to PowerShell profile
			wantSuffixes: func() []string {
				if runtime.GOOS == "windows" {
					return []string{".ps1", "profile.ps1"}
				}
				return []string{".profile", ".bash_profile"}
			}(),
		},
		{
			name:         "powershell",
			shell:        ShellPowerShell,
			preferLogin:  true,
			wantSuffixes: []string{".ps1", "profile.ps1"}, // Accept both PowerShell profile formats
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.GetConfigFile(tt.shell, tt.preferLogin)
			found := false
			for _, suffix := range tt.wantSuffixes {
				if strings.HasSuffix(got, suffix) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("GetConfigFile() = %v, want one of suffixes %v", got, tt.wantSuffixes)
			}
		})
	}
}

func TestHasEntry(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".zshenv")

	content := `# Shell config
export PATH="$HOME/.local/bin:$PATH"
export CLAUDE_DISABLE_PATH_WARNING=1
`
	if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := NewConfigurator()

	tests := []struct {
		name   string
		search string
		want   bool
	}{
		{
			name:   "find_existing_var",
			search: "CLAUDE_DISABLE_PATH_WARNING",
			want:   true,
		},
		{
			name:   "find_local_bin",
			search: ".local/bin",
			want:   true,
		},
		{
			name:   "not_found",
			search: "NONEXISTENT_VAR",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.HasEntry(configFile, tt.search)
			if err != nil {
				t.Fatalf("HasEntry() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("HasEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasEntry_FileNotExists(t *testing.T) {
	c := NewConfigurator()

	got, err := c.HasEntry("/nonexistent/path/.zshenv", "test")
	if err != nil {
		t.Fatalf("HasEntry() error = %v, expected nil for non-existent file", err)
	}
	if got {
		t.Error("HasEntry() = true, want false for non-existent file")
	}
}

func TestAddEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".zshenv")

	c := NewConfigurator()

	// First addition should succeed
	result, err := c.AddEnvVar(configFile, "TEST_VAR", "test_value")
	if err != nil {
		t.Fatalf("AddEnvVar() error = %v", err)
	}
	if !result.Success {
		t.Error("AddEnvVar() Success = false, want true")
	}
	if result.Skipped {
		t.Error("AddEnvVar() Skipped = true, want false")
	}

	// Verify content
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if !strings.Contains(string(content), "export TEST_VAR=test_value") {
		t.Errorf("Config file doesn't contain expected export: %s", content)
	}
	if !strings.Contains(string(content), "# Added by MoAI-ADK") {
		t.Errorf("Config file doesn't contain MoAI-ADK comment: %s", content)
	}
}

func TestAddEnvVar_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".zshenv")

	// Pre-create file with existing var
	content := "export CLAUDE_DISABLE_PATH_WARNING=1\n"
	if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := NewConfigurator()

	// Second addition should return ErrAlreadyConfigured
	result, err := c.AddEnvVar(configFile, "CLAUDE_DISABLE_PATH_WARNING", "1")
	if err != ErrAlreadyConfigured {
		t.Errorf("AddEnvVar() error = %v, want ErrAlreadyConfigured", err)
	}
	if result == nil {
		t.Fatal("AddEnvVar() result = nil, want non-nil")
	}
	if !result.Skipped {
		t.Error("AddEnvVar() Skipped = false, want true")
	}
}

func TestAddEnvVar_FishSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	fishDir := filepath.Join(tmpDir, ".config", "fish")
	if err := os.MkdirAll(fishDir, 0o755); err != nil {
		t.Fatalf("Failed to create fish config dir: %v", err)
	}
	configFile := filepath.Join(fishDir, "config.fish")

	c := NewConfigurator()

	result, err := c.AddEnvVar(configFile, "TEST_VAR", "test_value")
	if err != nil {
		t.Fatalf("AddEnvVar() error = %v", err)
	}
	if !result.Success {
		t.Error("AddEnvVar() Success = false, want true")
	}

	// Verify fish syntax
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if !strings.Contains(string(content), "set -gx TEST_VAR test_value") {
		t.Errorf("Config file doesn't contain fish syntax: %s", content)
	}
}

func TestAddPathEntry(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".zshenv")

	c := NewConfigurator()

	result, err := c.AddPathEntry(configFile, "$HOME/.local/bin")
	if err != nil {
		t.Fatalf("AddPathEntry() error = %v", err)
	}
	if !result.Success {
		t.Error("AddPathEntry() Success = false, want true")
	}

	// Verify content
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if !strings.Contains(string(content), `export PATH="$HOME/.local/bin:$PATH"`) {
		t.Errorf("Config file doesn't contain expected PATH export: %s", content)
	}
}

func TestAddPathEntry_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".zshenv")

	// Pre-create file with existing PATH
	content := `export PATH="$HOME/.local/bin:$PATH"` + "\n"
	if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := NewConfigurator()

	result, err := c.AddPathEntry(configFile, "$HOME/.local/bin")
	if err != ErrAlreadyConfigured {
		t.Errorf("AddPathEntry() error = %v, want ErrAlreadyConfigured", err)
	}
	if result == nil {
		t.Fatal("AddPathEntry() result = nil, want non-nil")
	}
	if !result.Skipped {
		t.Error("AddPathEntry() Skipped = false, want true")
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get user home directory")
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "expand_tilde",
			path: "~/.zshenv",
			want: filepath.Join(home, ".zshenv"),
		},
		{
			name: "absolute_path_unchanged",
			path: "/etc/profile",
			want: "/etc/profile",
		},
		{
			name: "relative_path_unchanged",
			path: ".bashrc",
			want: ".bashrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.path)
			if got != tt.want {
				t.Errorf("expandPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddEnvVar_PowerShellSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	psDir := filepath.Join(tmpDir, "Documents", "PowerShell")
	if err := os.MkdirAll(psDir, 0o755); err != nil {
		t.Fatalf("Failed to create PowerShell config dir: %v", err)
	}
	configFile := filepath.Join(psDir, "Microsoft.PowerShell_profile.ps1")

	c := NewConfigurator()

	result, err := c.AddEnvVar(configFile, "TEST_VAR", "test_value")
	if err != nil {
		t.Fatalf("AddEnvVar() error = %v", err)
	}
	if !result.Success {
		t.Error("AddEnvVar() Success = false, want true")
	}

	// Verify PowerShell syntax
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if !strings.Contains(string(content), "$env:TEST_VAR = \"test_value\"") {
		t.Errorf("Config file doesn't contain PowerShell syntax: %s", content)
	}
}

func TestAddPathEntry_PowerShellSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	psDir := filepath.Join(tmpDir, "Documents", "PowerShell")
	if err := os.MkdirAll(psDir, 0o755); err != nil {
		t.Fatalf("Failed to create PowerShell config dir: %v", err)
	}
	configFile := filepath.Join(psDir, "Microsoft.PowerShell_profile.ps1")

	c := NewConfigurator()

	result, err := c.AddPathEntry(configFile, "$HOME/.local/bin")
	if err != nil {
		t.Fatalf("AddPathEntry() error = %v", err)
	}
	if !result.Success {
		t.Error("AddPathEntry() Success = false, want true")
	}

	// Verify PowerShell syntax (Windows path format)
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	// Should contain PowerShell path with semicolon separator
	if !strings.Contains(string(content), "$env:PATH") {
		t.Errorf("Config file doesn't contain PowerShell PATH syntax: %s", content)
	}
}

func TestAddPathEntry_GoBin(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".zshenv")

	c := NewConfigurator()

	result, err := c.AddPathEntry(configFile, "$HOME/go/bin")
	if err != nil {
		t.Fatalf("AddPathEntry() error = %v", err)
	}
	if !result.Success {
		t.Error("AddPathEntry() Success = false, want true")
	}

	// Verify content
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if !strings.Contains(string(content), `export PATH="$HOME/go/bin:$PATH"`) {
		t.Errorf("Config file doesn't contain expected Go bin PATH export: %s", content)
	}
}

func TestAddPathEntry_GoBin_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".zshenv")

	// Pre-create file with existing go/bin PATH
	content := `export PATH="$HOME/go/bin:$PATH"` + "\n"
	if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	c := NewConfigurator()

	result, err := c.AddPathEntry(configFile, "$HOME/go/bin")
	if err != ErrAlreadyConfigured {
		t.Errorf("AddPathEntry() error = %v, want ErrAlreadyConfigured", err)
	}
	if result == nil {
		t.Fatal("AddPathEntry() result = nil, want non-nil")
	}
	if !result.Skipped {
		t.Error("AddPathEntry() Skipped = false, want true")
	}
}

func TestAddPathEntry_BothPaths(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".zshenv")

	c := NewConfigurator()

	// Add .local/bin first
	result1, err := c.AddPathEntry(configFile, "$HOME/.local/bin")
	if err != nil {
		t.Fatalf("AddPathEntry(.local/bin) error = %v", err)
	}
	if !result1.Success {
		t.Error("AddPathEntry(.local/bin) Success = false, want true")
	}

	// Add go/bin second - should succeed (different path)
	result2, err := c.AddPathEntry(configFile, "$HOME/go/bin")
	if err != nil {
		t.Fatalf("AddPathEntry(go/bin) error = %v", err)
	}
	if !result2.Success {
		t.Error("AddPathEntry(go/bin) Success = false, want true")
	}

	// Verify both paths are in the file
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if !strings.Contains(string(content), ".local/bin") {
		t.Errorf("Config file doesn't contain .local/bin: %s", content)
	}
	if !strings.Contains(string(content), "go/bin") {
		t.Errorf("Config file doesn't contain go/bin: %s", content)
	}
}

func TestAddPathEntry_FishSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	fishDir := filepath.Join(tmpDir, ".config", "fish")
	if err := os.MkdirAll(fishDir, 0o755); err != nil {
		t.Fatalf("Failed to create fish config dir: %v", err)
	}
	configFile := filepath.Join(fishDir, "config.fish")

	c := NewConfigurator()

	result, err := c.AddPathEntry(configFile, "$HOME/.local/bin")
	if err != nil {
		t.Fatalf("AddPathEntry() error = %v", err)
	}
	if !result.Success {
		t.Error("AddPathEntry() Success = false, want true")
	}

	// Verify fish syntax
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if !strings.Contains(string(content), "set -gx PATH") {
		t.Errorf("Config file doesn't contain fish PATH syntax: %s", content)
	}
}

func TestAddPathEntry_PowerShellSyntax_GoBin(t *testing.T) {
	tmpDir := t.TempDir()
	psDir := filepath.Join(tmpDir, "Documents", "PowerShell")
	if err := os.MkdirAll(psDir, 0o755); err != nil {
		t.Fatalf("Failed to create PowerShell config dir: %v", err)
	}
	configFile := filepath.Join(psDir, "Microsoft.PowerShell_profile.ps1")

	c := NewConfigurator()

	result, err := c.AddPathEntry(configFile, "$HOME/go/bin")
	if err != nil {
		t.Fatalf("AddPathEntry() error = %v", err)
	}
	if !result.Success {
		t.Error("AddPathEntry() Success = false, want true")
	}

	// Verify PowerShell syntax with Windows path
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if !strings.Contains(string(content), "$env:PATH") {
		t.Errorf("Config file doesn't contain PowerShell PATH syntax: %s", content)
	}
}

func TestAddPathEntry_GenericPath(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".zshenv")

	c := NewConfigurator()

	result, err := c.AddPathEntry(configFile, "/opt/custom/bin")
	if err != nil {
		t.Fatalf("AddPathEntry() error = %v", err)
	}
	if !result.Success {
		t.Error("AddPathEntry() Success = false, want true")
	}

	// Verify content
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if !strings.Contains(string(content), "/opt/custom/bin") {
		t.Errorf("Config file doesn't contain expected PATH export: %s", content)
	}
}

func TestSelectConfigFile_Fish(t *testing.T) {
	// Fish shell should always return config.fish path
	configFile := selectConfigFile(ShellFish, true)
	if !strings.HasSuffix(configFile, filepath.Join(".config", "fish", "config.fish")) {
		t.Errorf("selectConfigFile(fish) = %q, want suffix .config/fish/config.fish", configFile)
	}
}

func TestSelectConfigFile_UnknownNonLogin(t *testing.T) {
	configFile := selectConfigFile(ShellUnknown, false)
	if runtime.GOOS == "windows" {
		if !strings.Contains(configFile, "PowerShell") && !strings.HasSuffix(configFile, ".ps1") {
			t.Errorf("selectConfigFile(unknown, windows) = %q, want PowerShell path", configFile)
		}
	} else {
		if !strings.HasSuffix(configFile, ".profile") {
			t.Errorf("selectConfigFile(unknown, non-windows) = %q, want .profile suffix", configFile)
		}
	}
}

func TestGetPowerShellProfilePath(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("no_directories_exist", func(t *testing.T) {
		// When no directories exist, should return PowerShell Core path
		result := getPowerShellProfilePath(tmpDir)
		if !strings.Contains(result, "PowerShell") {
			t.Errorf("getPowerShellProfilePath() = %q, want to contain 'PowerShell'", result)
		}
	})

	t.Run("ps_core_directory_exists", func(t *testing.T) {
		dir := t.TempDir()
		psCoreDir := filepath.Join(dir, "Documents", "PowerShell")
		if err := os.MkdirAll(psCoreDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		result := getPowerShellProfilePath(dir)
		if !strings.Contains(result, "PowerShell") {
			t.Errorf("result = %q, want to contain 'PowerShell'", result)
		}
	})

	t.Run("windows_ps_directory_exists", func(t *testing.T) {
		dir := t.TempDir()
		winPSDir := filepath.Join(dir, "Documents", "WindowsPowerShell")
		if err := os.MkdirAll(winPSDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		result := getPowerShellProfilePath(dir)
		if !strings.Contains(result, "WindowsPowerShell") {
			t.Errorf("result = %q, want to contain 'WindowsPowerShell'", result)
		}
	})

	t.Run("ps_core_profile_exists", func(t *testing.T) {
		dir := t.TempDir()
		psCoreDir := filepath.Join(dir, "Documents", "PowerShell")
		if err := os.MkdirAll(psCoreDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		profilePath := filepath.Join(psCoreDir, "Microsoft.PowerShell_profile.ps1")
		if err := os.WriteFile(profilePath, []byte("# profile"), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}
		result := getPowerShellProfilePath(dir)
		if result != profilePath {
			t.Errorf("result = %q, want %q", result, profilePath)
		}
	})

	t.Run("windows_ps_profile_exists", func(t *testing.T) {
		dir := t.TempDir()
		winPSDir := filepath.Join(dir, "Documents", "WindowsPowerShell")
		if err := os.MkdirAll(winPSDir, 0o755); err != nil {
			t.Fatalf("MkdirAll error: %v", err)
		}
		profilePath := filepath.Join(winPSDir, "Microsoft.PowerShell_profile.ps1")
		if err := os.WriteFile(profilePath, []byte("# profile"), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}
		result := getPowerShellProfilePath(dir)
		if result != profilePath {
			t.Errorf("result = %q, want %q", result, profilePath)
		}
	})
}

func TestIsPowerShellProfile(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		want       bool
	}{
		{
			name:       "ps1_extension",
			configFile: "/Users/test/Documents/PowerShell/Microsoft.PowerShell_profile.ps1",
			want:       true,
		},
		{
			name:       "windows_powershell_dir",
			configFile: "C:\\Users\\test\\Documents\\WindowsPowerShell\\profile.ps1",
			want:       true,
		},
		{
			name:       "powershell_in_path",
			configFile: "/home/user/Documents/PowerShell/profile.ps1",
			want:       true,
		},
		{
			name:       "zshenv_not_powershell",
			configFile: "/Users/test/.zshenv",
			want:       false,
		},
		{
			name:       "bashrc_not_powershell",
			configFile: "/home/user/.bashrc",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPowerShellProfile(tt.configFile)
			if got != tt.want {
				t.Errorf("isPowerShellProfile(%q) = %v, want %v", tt.configFile, got, tt.want)
			}
		})
	}
}
