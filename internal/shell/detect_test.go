package shell

import (
	"testing"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		shellEnv string
		want     ShellType
	}{
		{
			name:     "zsh_lowercase",
			shellEnv: "/bin/zsh",
			want:     ShellZsh,
		},
		{
			name:     "zsh_usr_bin",
			shellEnv: "/usr/bin/zsh",
			want:     ShellZsh,
		},
		{
			name:     "bash_lowercase",
			shellEnv: "/bin/bash",
			want:     ShellBash,
		},
		{
			name:     "bash_usr_local",
			shellEnv: "/usr/local/bin/bash",
			want:     ShellBash,
		},
		{
			name:     "fish_lowercase",
			shellEnv: "/usr/local/bin/fish",
			want:     ShellFish,
		},
		{
			name:     "fish_homebrew",
			shellEnv: "/opt/homebrew/bin/fish",
			want:     ShellFish,
		},
		{
			name:     "unknown_sh",
			shellEnv: "/bin/sh",
			want:     ShellUnknown,
		},
		{
			name:     "empty_shell",
			shellEnv: "",
			want:     ShellUnknown,
		},
		{
			name:     "pwsh_shell",
			shellEnv: "/usr/local/bin/pwsh",
			want:     ShellPowerShell,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDetectorWithEnv(func(key string) string {
				if key == "SHELL" {
					return tt.shellEnv
				}
				return ""
			})

			got := d.DetectShell()
			if got != tt.want {
				t.Errorf("DetectShell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsWSL(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{
			name: "wsl_distro_name",
			env:  map[string]string{"WSL_DISTRO_NAME": "Ubuntu"},
			want: true,
		},
		{
			name: "wslenv",
			env:  map[string]string{"WSLENV": "something"},
			want: true,
		},
		{
			name: "wsl_interop",
			env:  map[string]string{"WSL_INTEROP": "/run/WSL/1_interop"},
			want: true,
		},
		{
			name: "not_wsl_empty",
			env:  map[string]string{},
			want: false,
		},
		{
			name: "not_wsl_other_vars",
			env:  map[string]string{"HOME": "/home/user", "PATH": "/usr/bin"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDetectorWithEnv(func(key string) string {
				return tt.env[key]
			})

			got := d.IsWSL()
			if got != tt.want {
				t.Errorf("IsWSL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetShellConfig(t *testing.T) {
	tests := []struct {
		name           string
		env            map[string]string
		wantShell      ShellType
		wantWSL        bool
		wantConfigFile string
	}{
		{
			name: "zsh_macos",
			env: map[string]string{
				"SHELL": "/bin/zsh",
				"HOME":  "/Users/testuser",
			},
			wantShell:      ShellZsh,
			wantWSL:        false,
			wantConfigFile: "/Users/testuser/.zshenv",
		},
		{
			name: "bash_wsl",
			env: map[string]string{
				"SHELL":           "/bin/bash",
				"HOME":            "/home/testuser",
				"WSL_DISTRO_NAME": "Ubuntu",
			},
			wantShell:      ShellBash,
			wantWSL:        true,
			wantConfigFile: "/home/testuser/.profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDetectorWithEnv(func(key string) string {
				return tt.env[key]
			})

			config := d.GetShellConfig()

			if config.Shell != tt.wantShell {
				t.Errorf("Shell = %v, want %v", config.Shell, tt.wantShell)
			}
			if config.IsWSL != tt.wantWSL {
				t.Errorf("IsWSL = %v, want %v", config.IsWSL, tt.wantWSL)
			}
		})
	}
}

func TestDetectShell_PowerShell(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want ShellType
	}{
		{
			name: "windows_powershell",
			env: map[string]string{
				"OS":           "Windows_NT",
				"PSModulePath": "C:\\Users\\test\\Documents\\PowerShell\\Modules",
			},
			want: ShellPowerShell,
		},
		{
			name: "powershell_core_any_platform",
			env: map[string]string{
				"POWERSHELL_DISTRIBUTION_CHANNEL": "PSDocker",
				"PSModulePath":                    "/usr/local/share/powershell/Modules",
			},
			want: ShellPowerShell,
		},
		{
			name: "no_shell_but_psmodulepath",
			env: map[string]string{
				"PSModulePath": "C:\\Users\\test\\Documents\\PowerShell\\Modules",
			},
			want: ShellPowerShell,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDetectorWithEnv(func(key string) string {
				return tt.env[key]
			})

			got := d.DetectShell()
			if got != tt.want {
				t.Errorf("DetectShell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectShell_GitBash(t *testing.T) {
	tests := []struct {
		name    string
		msystem string
		want    ShellType
	}{
		{
			name:    "mingw64",
			msystem: "MINGW64",
			want:    ShellBash,
		},
		{
			name:    "mingw32",
			msystem: "MINGW32",
			want:    ShellBash,
		},
		{
			name:    "ucrt64",
			msystem: "UCRT64",
			want:    ShellBash,
		},
		{
			name:    "msys",
			msystem: "MSYS",
			want:    ShellBash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDetectorWithEnv(func(key string) string {
				if key == "MSYSTEM" {
					return tt.msystem
				}
				return ""
			})

			got := d.DetectShell()
			if got != tt.want {
				t.Errorf("DetectShell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShellTypeString(t *testing.T) {
	tests := []struct {
		shell ShellType
		want  string
	}{
		{ShellZsh, "zsh"},
		{ShellBash, "bash"},
		{ShellFish, "fish"},
		{ShellPowerShell, "powershell"},
		{ShellUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.shell), func(t *testing.T) {
			if got := tt.shell.String(); got != tt.want {
				t.Errorf("ShellType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
