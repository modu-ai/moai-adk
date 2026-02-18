package shell

import (
	"errors"
	"io"
	"log/slog"
	"testing"
)

// mockDetector implements Detector for testing.
type mockDetector struct {
	shell  ShellType
	isWSL  bool
	config *ShellConfig
}

func (m *mockDetector) DetectShell() ShellType { return m.shell }
func (m *mockDetector) IsWSL() bool            { return m.isWSL }
func (m *mockDetector) GetShellConfig() *ShellConfig {
	if m.config != nil {
		return m.config
	}
	return &ShellConfig{
		Shell:      m.shell,
		ConfigFile: "/tmp/test-config",
		IsWSL:      m.isWSL,
		Platform:   "linux",
	}
}

// mockConfigurator implements Configurator for testing.
type mockConfigurator struct {
	getConfigFileResult string
	addEnvVarResult     *ConfigResult
	addEnvVarErr        error
	addPathEntryResult  *ConfigResult
	addPathEntryErr     error
	hasEntryResult      bool
	hasEntryErr         error

	addEnvVarCalls    []addEnvVarCall
	addPathEntryCalls []addPathEntryCall
}

type addEnvVarCall struct {
	configFile, varName, varValue string
}

type addPathEntryCall struct {
	configFile, pathEntry string
}

func (m *mockConfigurator) GetConfigFile(shell ShellType, preferLogin bool) string {
	return m.getConfigFileResult
}

func (m *mockConfigurator) AddEnvVar(configFile, varName, varValue string) (*ConfigResult, error) {
	m.addEnvVarCalls = append(m.addEnvVarCalls, addEnvVarCall{configFile, varName, varValue})
	return m.addEnvVarResult, m.addEnvVarErr
}

func (m *mockConfigurator) AddPathEntry(configFile, pathEntry string) (*ConfigResult, error) {
	m.addPathEntryCalls = append(m.addPathEntryCalls, addPathEntryCall{configFile, pathEntry})
	return m.addPathEntryResult, m.addPathEntryErr
}

func (m *mockConfigurator) HasEntry(configFile, search string) (bool, error) {
	return m.hasEntryResult, m.hasEntryErr
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewEnvConfigurator(t *testing.T) {
	t.Run("with_logger", func(t *testing.T) {
		ec := NewEnvConfigurator(discardLogger())
		if ec == nil {
			t.Fatal("NewEnvConfigurator returned nil")
		}
	})

	t.Run("nil_logger_uses_discard", func(t *testing.T) {
		ec := NewEnvConfigurator(nil)
		if ec == nil {
			t.Fatal("NewEnvConfigurator(nil) returned nil")
		}
		// Should still be functional without panics
		config := ec.GetShellConfig()
		if config == nil {
			t.Error("GetShellConfig() returned nil")
		}
	})
}

func TestGetShellConfig_EnvConfigurator(t *testing.T) {
	md := &mockDetector{
		config: &ShellConfig{
			Shell:      ShellZsh,
			ConfigFile: "/home/user/.zshenv",
			IsWSL:      false,
			Platform:   "darwin",
		},
	}
	ec := &envConfigurator{
		detector:     md,
		configurator: &mockConfigurator{},
		logger:       discardLogger(),
	}

	config := ec.GetShellConfig()
	if config == nil {
		t.Fatal("GetShellConfig() returned nil")
	}
	if config.Shell != ShellZsh {
		t.Errorf("Shell = %v, want %v", config.Shell, ShellZsh)
	}
	if config.ConfigFile != "/home/user/.zshenv" {
		t.Errorf("ConfigFile = %v, want /home/user/.zshenv", config.ConfigFile)
	}
}

func TestGetRecommendation(t *testing.T) {
	tests := []struct {
		name             string
		shell            ShellType
		isWSL            bool
		wantChangesCount int
		wantExplanation  string
	}{
		{
			name:             "zsh",
			shell:            ShellZsh,
			wantChangesCount: 3,
			wantExplanation:  ".zshenv is loaded for ALL shells",
		},
		{
			name:             "bash_non_wsl",
			shell:            ShellBash,
			isWSL:            false,
			wantChangesCount: 3,
			wantExplanation:  ".profile or .bash_profile",
		},
		{
			name:             "bash_wsl",
			shell:            ShellBash,
			isWSL:            true,
			wantChangesCount: 3,
			wantExplanation:  ".profile is used because WSL",
		},
		{
			name:             "fish",
			shell:            ShellFish,
			wantChangesCount: 3,
			wantExplanation:  "config.fish is the standard",
		},
		{
			name:             "powershell",
			shell:            ShellPowerShell,
			wantChangesCount: 3,
			wantExplanation:  "PowerShell profile is loaded",
		},
		{
			name:             "unknown_shell",
			shell:            ShellUnknown,
			wantChangesCount: 3,
			wantExplanation:  ".profile is used as the fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := &mockDetector{
				config: &ShellConfig{
					Shell:      tt.shell,
					ConfigFile: "/tmp/config",
					IsWSL:      tt.isWSL,
					Platform:   "linux",
				},
			}
			ec := &envConfigurator{
				detector:     md,
				configurator: &mockConfigurator{},
				logger:       discardLogger(),
			}

			rec := ec.GetRecommendation()
			if rec == nil {
				t.Fatal("GetRecommendation() returned nil")
			}
			if rec.Shell != tt.shell {
				t.Errorf("Shell = %v, want %v", rec.Shell, tt.shell)
			}
			if len(rec.Changes) != tt.wantChangesCount {
				t.Errorf("len(Changes) = %d, want %d", len(rec.Changes), tt.wantChangesCount)
			}
			if rec.Explanation == "" {
				t.Error("Explanation is empty")
			}
			if tt.wantExplanation != "" {
				found := false
				if len(rec.Explanation) >= len(tt.wantExplanation) {
					for i := 0; i <= len(rec.Explanation)-len(tt.wantExplanation); i++ {
						if rec.Explanation[i:i+len(tt.wantExplanation)] == tt.wantExplanation {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("Explanation = %q, want to contain %q", rec.Explanation, tt.wantExplanation)
				}
			}
		})
	}
}

func TestConfigure_DryRun(t *testing.T) {
	md := &mockDetector{
		config: &ShellConfig{
			Shell:      ShellZsh,
			ConfigFile: "/home/user/.zshenv",
			IsWSL:      false,
			Platform:   "darwin",
		},
	}
	mc := &mockConfigurator{}
	ec := &envConfigurator{
		detector:     md,
		configurator: mc,
		logger:       discardLogger(),
	}

	result, err := ec.Configure(ConfigOptions{
		DryRun:                  true,
		AddClaudeWarningDisable: true,
		AddLocalBinPath:         true,
		AddGoBinPath:            true,
	})
	if err != nil {
		t.Fatalf("Configure(DryRun) error = %v", err)
	}
	if !result.Success {
		t.Error("Success = false, want true")
	}
	if result.ConfigFile != "/home/user/.zshenv" {
		t.Errorf("ConfigFile = %v, want /home/user/.zshenv", result.ConfigFile)
	}
	// In dry-run mode, the configurator should NOT be called
	if len(mc.addEnvVarCalls) != 0 {
		t.Errorf("AddEnvVar called %d times in DryRun mode", len(mc.addEnvVarCalls))
	}
}

func TestConfigure_UnknownShell(t *testing.T) {
	md := &mockDetector{
		config: &ShellConfig{
			Shell:      ShellUnknown,
			ConfigFile: "/home/user/.profile",
			IsWSL:      false,
			Platform:   "linux",
		},
	}
	mc := &mockConfigurator{
		addEnvVarResult: &ConfigResult{
			Success:    true,
			LinesAdded: []string{"export CLAUDE_DISABLE_PATH_WARNING=1"},
		},
		addPathEntryResult: &ConfigResult{
			Success:    true,
			LinesAdded: []string{"export PATH=\"$HOME/.local/bin:$PATH\""},
		},
	}
	ec := &envConfigurator{
		detector:     md,
		configurator: mc,
		logger:       discardLogger(),
	}

	result, err := ec.Configure(ConfigOptions{
		AddClaudeWarningDisable: true,
		AddLocalBinPath:         true,
		AddGoBinPath:            true,
	})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if !result.Success {
		t.Error("Success = false, want true")
	}
}

func TestConfigure_AllOptionsSuccess(t *testing.T) {
	md := &mockDetector{
		config: &ShellConfig{
			Shell:      ShellZsh,
			ConfigFile: "/home/user/.zshenv",
			IsWSL:      false,
			Platform:   "darwin",
		},
	}
	mc := &mockConfigurator{
		addEnvVarResult: &ConfigResult{
			Success:    true,
			LinesAdded: []string{"export CLAUDE_DISABLE_PATH_WARNING=1"},
		},
		addPathEntryResult: &ConfigResult{
			Success:    true,
			LinesAdded: []string{"export PATH=\"$HOME/.local/bin:$PATH\""},
		},
	}
	ec := &envConfigurator{
		detector:     md,
		configurator: mc,
		logger:       discardLogger(),
	}

	result, err := ec.Configure(ConfigOptions{
		AddClaudeWarningDisable: true,
		AddLocalBinPath:         true,
		AddGoBinPath:            true,
	})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if !result.Success {
		t.Error("Success = false, want true")
	}
	if result.Skipped {
		t.Error("Skipped = true, want false")
	}
	if len(result.LinesAdded) != 3 {
		t.Errorf("len(LinesAdded) = %d, want 3", len(result.LinesAdded))
	}
	if len(mc.addEnvVarCalls) != 1 {
		t.Errorf("AddEnvVar called %d times, want 1", len(mc.addEnvVarCalls))
	}
	if len(mc.addPathEntryCalls) != 2 {
		t.Errorf("AddPathEntry called %d times, want 2", len(mc.addPathEntryCalls))
	}
}

func TestConfigure_AllAlreadyConfigured(t *testing.T) {
	md := &mockDetector{
		config: &ShellConfig{
			Shell:      ShellZsh,
			ConfigFile: "/home/user/.zshenv",
			IsWSL:      false,
			Platform:   "darwin",
		},
	}
	mc := &mockConfigurator{
		addEnvVarResult: &ConfigResult{
			Success: true,
			Skipped: true,
		},
		addEnvVarErr: ErrAlreadyConfigured,
		addPathEntryResult: &ConfigResult{
			Success: true,
			Skipped: true,
		},
		addPathEntryErr: ErrAlreadyConfigured,
	}
	ec := &envConfigurator{
		detector:     md,
		configurator: mc,
		logger:       discardLogger(),
	}

	result, err := ec.Configure(ConfigOptions{
		AddClaudeWarningDisable: true,
		AddLocalBinPath:         true,
		AddGoBinPath:            true,
	})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if !result.Success {
		t.Error("Success = false, want true")
	}
	if !result.Skipped {
		t.Error("Skipped = false, want true (all already configured)")
	}
}

func TestConfigure_EnvVarError(t *testing.T) {
	md := &mockDetector{
		config: &ShellConfig{
			Shell:      ShellZsh,
			ConfigFile: "/home/user/.zshenv",
			IsWSL:      false,
			Platform:   "darwin",
		},
	}
	writeErr := errors.New("write failed")
	mc := &mockConfigurator{
		addEnvVarErr: writeErr,
	}
	ec := &envConfigurator{
		detector:     md,
		configurator: mc,
		logger:       discardLogger(),
	}

	_, err := ec.Configure(ConfigOptions{
		AddClaudeWarningDisable: true,
	})
	if err == nil {
		t.Fatal("Configure() error = nil, want error")
	}
}

func TestConfigure_PathEntryError(t *testing.T) {
	md := &mockDetector{
		config: &ShellConfig{
			Shell:      ShellZsh,
			ConfigFile: "/home/user/.zshenv",
			IsWSL:      false,
			Platform:   "darwin",
		},
	}
	writeErr := errors.New("write failed")
	mc := &mockConfigurator{
		addPathEntryErr: writeErr,
	}
	ec := &envConfigurator{
		detector:     md,
		configurator: mc,
		logger:       discardLogger(),
	}

	_, err := ec.Configure(ConfigOptions{
		AddLocalBinPath: true,
	})
	if err == nil {
		t.Fatal("Configure() error = nil, want error")
	}
}

func TestConfigure_GoBinPathError(t *testing.T) {
	md := &mockDetector{
		config: &ShellConfig{
			Shell:      ShellZsh,
			ConfigFile: "/home/user/.zshenv",
			IsWSL:      false,
			Platform:   "darwin",
		},
	}
	writeErr := errors.New("write failed")
	mc := &mockConfigurator{
		addPathEntryErr: writeErr,
	}
	ec := &envConfigurator{
		detector:     md,
		configurator: mc,
		logger:       discardLogger(),
	}

	_, err := ec.Configure(ConfigOptions{
		AddGoBinPath: true,
	})
	if err == nil {
		t.Fatal("Configure() error = nil, want error")
	}
}

func TestConfigure_NoOptions(t *testing.T) {
	md := &mockDetector{
		config: &ShellConfig{
			Shell:      ShellZsh,
			ConfigFile: "/home/user/.zshenv",
			IsWSL:      false,
			Platform:   "darwin",
		},
	}
	mc := &mockConfigurator{}
	ec := &envConfigurator{
		detector:     md,
		configurator: mc,
		logger:       discardLogger(),
	}

	result, err := ec.Configure(ConfigOptions{})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if !result.Success {
		t.Error("Success = false, want true")
	}
	if !result.Skipped {
		t.Error("Skipped = false, want true (no options enabled)")
	}
	if len(mc.addEnvVarCalls) != 0 {
		t.Errorf("AddEnvVar called %d times, want 0", len(mc.addEnvVarCalls))
	}
	if len(mc.addPathEntryCalls) != 0 {
		t.Errorf("AddPathEntry called %d times, want 0", len(mc.addPathEntryCalls))
	}
}

func TestGetPlatformInfo(t *testing.T) {
	info := GetPlatformInfo()
	if info == "" {
		t.Error("GetPlatformInfo() returned empty string")
	}
	// Should contain basic platform information
	if len(info) < 10 {
		t.Errorf("GetPlatformInfo() too short: %q", info)
	}
}

func TestConfigureDefault(t *testing.T) {
	// ConfigureDefault uses real detector and configurator, which will write
	// to the actual shell config file. We simply verify it does not panic
	// and returns a result. The actual file modifications are tested
	// via the individual Configurator tests with temp dirs.
	result, err := ConfigureDefault(discardLogger())
	// On most machines this will succeed (possibly with Skipped=true if already configured)
	// or fail with permission errors. Either way, it should not panic.
	if err != nil {
		// Some CI environments may not have writable home dirs
		t.Logf("ConfigureDefault returned error (may be expected in CI): %v", err)
		return
	}
	if result == nil {
		t.Fatal("ConfigureDefault returned nil result without error")
	}
	if !result.Success {
		t.Error("Success = false, want true")
	}
}

func TestIsCmdPrompt(t *testing.T) {
	// On non-Windows platforms, IsCmdPrompt always returns false
	got := IsCmdPrompt()
	if got {
		// On macOS/Linux this should always be false
		t.Errorf("IsCmdPrompt() = true, want false on non-Windows platform")
	}
}

func TestConfigure_MixedSuccessAndAlreadyConfigured(t *testing.T) {
	callCount := 0
	md := &mockDetector{
		config: &ShellConfig{
			Shell:      ShellZsh,
			ConfigFile: "/home/user/.zshenv",
			IsWSL:      false,
			Platform:   "darwin",
		},
	}
	mc := &mockConfigurator{
		addEnvVarResult: &ConfigResult{
			Success:    true,
			LinesAdded: []string{"export CLAUDE_DISABLE_PATH_WARNING=1"},
		},
	}
	// Override AddPathEntry to return already-configured on first call, success on second
	originalMC := mc
	ec := &envConfigurator{
		detector: md,
		configurator: &mixedConfigurator{
			envVarResult: originalMC.addEnvVarResult,
			pathCallNum:  &callCount,
		},
		logger: discardLogger(),
	}

	result, err := ec.Configure(ConfigOptions{
		AddClaudeWarningDisable: true,
		AddLocalBinPath:         true,
		AddGoBinPath:            true,
	})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if !result.Success {
		t.Error("Success = false, want true")
	}
	// At least the env var was added, so it should not be fully skipped
	if result.Skipped {
		t.Error("Skipped = true, want false (at least one success)")
	}
}

// mixedConfigurator returns different results for sequential AddPathEntry calls.
type mixedConfigurator struct {
	envVarResult *ConfigResult
	pathCallNum  *int
}

func (m *mixedConfigurator) GetConfigFile(shell ShellType, preferLogin bool) string {
	return ""
}

func (m *mixedConfigurator) AddEnvVar(configFile, varName, varValue string) (*ConfigResult, error) {
	return m.envVarResult, nil
}

func (m *mixedConfigurator) AddPathEntry(configFile, pathEntry string) (*ConfigResult, error) {
	*m.pathCallNum++
	if *m.pathCallNum == 1 {
		// First call: already configured
		return &ConfigResult{Success: true, Skipped: true}, ErrAlreadyConfigured
	}
	// Second call: success
	return &ConfigResult{
		Success:    true,
		LinesAdded: []string{"export PATH=\"$HOME/go/bin:$PATH\""},
	}, nil
}

func (m *mixedConfigurator) HasEntry(configFile, search string) (bool, error) {
	return false, nil
}
