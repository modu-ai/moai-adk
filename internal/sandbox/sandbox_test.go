package sandbox

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

// @MX:TODO: 테스트 커버리지를 85% 이상으로 높이고 경계 조건(edge cases)에 대한 테스트를 추가

func TestSandboxEnumValues(t *testing.T) {
	tests := []struct {
		name    string
		sandbox Sandbox
		want    string
	}{
		{"None", None, "none"},
		{"Bubblewrap", Bubblewrap, "bubblewrap"},
		{"Seatbelt", Seatbelt, "seatbelt"},
		{"Docker", Docker, "docker"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.sandbox); got != tt.want {
				t.Errorf("Sandbox.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectPlatform(t *testing.T) {
	sandbox := DetectPlatform()

	switch runtime.GOOS {
	case "darwin":
		if sandbox != Seatbelt && sandbox != Docker {
			t.Errorf("DetectPlatform() on darwin = %v, want Seatbelt or Docker", sandbox)
		}
	case "linux":
		if os.Getenv("CI") == "1" {
			if sandbox != Docker {
				t.Errorf("DetectPlatform() on linux with CI=1 = %v, want Docker", sandbox)
			}
		} else {
			if sandbox != Bubblewrap && sandbox != Docker {
				t.Errorf("DetectPlatform() on linux = %v, want Bubblewrap or Docker", sandbox)
			}
		}
	default:
		if sandbox != None {
			t.Errorf("DetectPlatform() on unsupported platform = %v, want None", sandbox)
		}
	}
}

func TestDefaultSandboxOptions(t *testing.T) {
	worktreeRoot := t.TempDir()
	opts := DefaultSandboxOptions(worktreeRoot)

	if opts.WorktreeRoot != worktreeRoot {
		t.Errorf("DefaultSandboxOptions() WorktreeRoot = %v, want %v", opts.WorktreeRoot, worktreeRoot)
	}

	if len(opts.ReadOnlyDirs) == 0 {
		t.Error("DefaultSandboxOptions() ReadOnlyDirs is empty")
	}

	// 네트워크 허용목록 확인
	expectedHosts := []string{
		"github.com",
		"registry.npmjs.org",
		"pypi.org",
		"proxy.golang.org",
		"crates.io",
		"repo.maven.apache.org",
		"rubygems.org",
		"pub.dev",
	}

	for _, host := range expectedHosts {
		found := false
		for _, allowed := range opts.NetworkAllow {
			if allowed == host {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DefaultSandboxOptions() NetworkAllow missing host: %v", host)
		}
	}
}

func TestScrubEnv(t *testing.T) {
	tests := []struct {
		name        string
		env         []string
		passthrough []string
		want        []string
	}{
		{
			name:        "scrub secrets",
			env:         []string{"AWS_ACCESS_KEY_ID=secret", "PATH=/usr/bin"},
			passthrough: []string{},
			want:        []string{"PATH=/usr/bin"},
		},
		{
			name:        "passthrough preserves",
			env:         []string{"AWS_ACCESS_KEY_ID=secret", "CUSTOM_VAR=value"},
			passthrough: []string{"CUSTOM_VAR"},
			want:        []string{"CUSTOM_VAR=value"},
		},
		{
			name:        "scrub all secrets",
			env:         []string{"GITHUB_TOKEN=token", "ANTHROPIC_API_KEY=key", "OPENAI_API_KEY=key2", "NPM_TOKEN=token3", "GH_TOKEN=token4"},
			passthrough: []string{},
			want:        []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scrubEnv(tt.env)

			// 길이 확인
			if len(got) != len(tt.want) {
				t.Errorf("scrubEnv() length = %v, want %v", len(got), len(tt.want))
			}

			// 내용 확인
			for _, w := range tt.want {
				found := false
				for _, g := range got {
					if g == w {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("scrubEnv() missing wanted entry: %v", w)
				}
			}

			// 비밀 누출 확인
			for _, g := range got {
				if strings.Contains(g, "AWS_") || strings.Contains(g, "TOKEN") || strings.Contains(g, "API_KEY") {
					t.Errorf("scrubEnv() leaked secret: %v", g)
				}
			}
		})
	}
}

func TestGenerateSeatbeltProfile(t *testing.T) {
	tests := []struct {
		name    string
		opts    SandboxOptions
		wantErr bool
	}{
		{
			name: "valid profile",
			opts: SandboxOptions{
				WorktreeRoot:   t.TempDir(),
				ReadOnlyDirs:   []string{"/"},
				WritablePaths:  []string{},
				NetworkAllow:   []string{},
				EnvPassthrough: []string{},
			},
			wantErr: false,
		},
		{
			name: "with network allowlist",
			opts: SandboxOptions{
				WorktreeRoot:   t.TempDir(),
				ReadOnlyDirs:   []string{"/"},
				WritablePaths:  []string{},
				NetworkAllow:   []string{"github.com"},
				EnvPassthrough: []string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := generateSeatbeltProfile(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateSeatbeltProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 프로필 구문 확인
				if !strings.HasPrefix(profile, "(version 1)") {
					t.Error("generateSeatbeltProfile() profile must start with '(version 1)'")
				}

				// 필수 규칙 확인
				requiredRules := []string{
					"(deny default)",
					"(allow process*)",
				}

				for _, rule := range requiredRules {
					if !strings.Contains(profile, rule) {
						t.Errorf("generateSeatbeltProfile() profile missing required rule: %v", rule)
					}
				}
			}
		})
	}
}

func TestGenerateBubblewrapArgs(t *testing.T) {
	tests := []struct {
		name    string
		opts    SandboxOptions
		wantErr bool
	}{
		{
			name: "valid args",
			opts: SandboxOptions{
				WorktreeRoot:   t.TempDir(),
				ReadOnlyDirs:   []string{"/"},
				WritablePaths:  []string{},
				NetworkAllow:   []string{},
				EnvPassthrough: []string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := generateBubblewrapArgs(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateBubblewrapArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 필수 인수 확인
				requiredArgs := []string{
					"--unshare-all",
					"--die-with-parent",
				}

				for _, arg := range requiredArgs {
					found := false
					for _, a := range args {
						if a == arg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("generateBubblewrapArgs() missing required arg: %v", arg)
					}
				}
			}
		})
	}
}

func TestTruncateOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   []byte
		wantTrunc bool
	}{
		{
			name:     "no truncation",
			output:   []byte("small output"),
			wantTrunc: false,
		},
		{
			name:     "truncation needed",
			output:   make([]byte, maxOutputSize+1),
			wantTrunc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateOutput(tt.output)

			if tt.wantTrunc {
				if len(got) <= maxOutputSize {
					t.Errorf("truncateOutput() length = %v, want > %v", len(got), maxOutputSize)
				}

				// 시스템 메시지 확인
				if !strings.Contains(string(got), "[System]: Output truncated") {
					t.Error("truncateOutput() missing system message")
				}
			} else {
				if len(got) != len(tt.output) {
					t.Errorf("truncateOutput() length = %v, want %v", len(got), len(tt.output))
				}
			}
		})
	}
}

func TestDetectSudoAttempts(t *testing.T) {
	tests := []struct {
		name     string
		output   []byte
		wantDetect bool
	}{
		{
			name:     "no sudo",
			output:   []byte("normal output"),
			wantDetect: false,
		},
		{
			name:     "sudo detected",
			output:   []byte("sudo command"),
			wantDetect: true,
		},
		{
			name:     "su detected",
			output:   []byte("su - user"),
			wantDetect: true,
		},
		{
			name:     "setuid detected",
			output:   []byte("setuid root"),
			wantDetect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectSudoAttempts(tt.output)
			if got != tt.wantDetect {
				t.Errorf("detectSudoAttempts() = %v, want %v", got, tt.wantDetect)
			}
		})
	}
}

func TestDetectPrivilegeEscalation(t *testing.T) {
	tests := []struct {
		name     string
		cmd      []string
		wantDetect bool
	}{
		{
			name:     "normal command",
			cmd:      []string{"ls", "-la"},
			wantDetect: false,
		},
		{
			name:     "sudo command",
			cmd:      []string{"sudo", "apt-get", "update"},
			wantDetect: true,
		},
		{
			name:     "su command",
			cmd:      []string{"su", "-"},
			wantDetect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectPrivilegeEscalation(tt.cmd)
			if got != tt.wantDetect {
				t.Errorf("DetectPrivilegeEscalation() = %v, want %v", got, tt.wantDetect)
			}
		})
	}
}

func TestValidateNetworkAllowlist(t *testing.T) {
	tests := []struct {
		name      string
		allowlist []string
		wantErr   bool
	}{
		{
			name:      "empty allowlist",
			allowlist: []string{},
			wantErr:   false,
		},
		{
			name:      "valid allowlist",
			allowlist: []string{"github.com", "registry.npmjs.org"},
			wantErr:   false,
		},
		{
			name:      "invalid allowlist",
			allowlist: []string{"github.com", "example.com"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNetworkAllowlist(tt.allowlist)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNetworkAllowlist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSBPLProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		wantErr bool
	}{
		{
			name:    "valid profile",
			profile: "(version 1)\n(deny default)\n(allow process*)\n",
			wantErr: false,
		},
		{
			name:    "missing version",
			profile: "(deny default)\n(allow process*)\n",
			wantErr: true,
		},
		{
			name:    "unbalanced parentheses",
			profile: "(version 1\n(deny default)\n",
			wantErr: true,
		},
		{
			name:    "missing deny default",
			profile: "(version 1)\n(allow process*)\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSBPLProfile(tt.profile)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSBPLProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckSandbox(t *testing.T) {
	available := CheckSandbox()

	// 맵이 비어 있지 않은지 확인
	if len(available) == 0 {
		t.Error("CheckSandbox() returned empty map")
	}

	// 현재 플랫폼에 맞는 백엔드 확인
	switch runtime.GOOS {
	case "darwin":
		if _, ok := available["seatbelt"]; !ok {
			t.Error("CheckSandbox() on darwin missing seatbelt key")
		}
	case "linux":
		if _, ok := available["bubblewrap"]; !ok {
			t.Error("CheckSandbox() on linux missing bubblewrap key")
		}
	}
}

func TestSandboxLauncherExecNone(t *testing.T) {
	launcher := &SandboxLauncher{}

	// 간단한 명령으로 테스트
	cmd := []string{"echo", "hello"}

	output, err := launcher.Exec(cmd, None, SandboxOptions{})
	if err != nil {
		t.Errorf("SandboxLauncher.Exec() error = %v", err)
	}

	// 출력 확인
	if !strings.Contains(string(output), "hello") {
		t.Errorf("SandboxLauncher.Exec() output = %v, want contains 'hello'", string(output))
	}
}

func TestSandboxLauncherExecBackendUnavailable(t *testing.T) {
	launcher := &SandboxLauncher{}

	// 지원되지 않는 백엔드로 테스트
	cmd := []string{"echo", "hello"}

	// 현재 플랫폼에서 지원되지 않는 백엔드 선택
	var unsupported Sandbox
	switch runtime.GOOS {
	case "darwin":
		unsupported = Bubblewrap
	case "linux":
		unsupported = Seatbelt
	default:
		unsupported = Bubblewrap
	}

	_, err := launcher.Exec(cmd, unsupported, SandboxOptions{})
	if err == nil {
		t.Error("SandboxLauncher.Exec() expected error for unsupported backend, got nil")
	}

	if !strings.Contains(err.Error(), "unavailable") {
		t.Errorf("SandboxLauncher.Exec() error = %v, want contains 'unavailable'", err)
	}
}

// Mock LookPath를 사용한 기능 감지 테스트
func TestCheckBubblewrapAvailableMock(t *testing.T) {
	// bwrap이 있는 시나리오만 테스트 (실제 시스템에 의존)
	// bwrap은 Linux에서만 사용 가능하므로 macOS에서는 false를 반환
	result := checkBubblewrapAvailable()

	// macOS에서는 false가 예상됨
	if runtime.GOOS == "darwin" && result {
		t.Error("checkBubblewrapAvailable() on darwin should return false")
	}

	// Linux에서는 시스템에 bwrap이 있는지 확인
	if runtime.GOOS == "linux" {
		// 실제 Linux 시스템에서는 bwrap이 있을 수 있으므로 결과만 확인
		t.Logf("checkBubblewrapAvailable() on linux = %v", result)
	}
}

func TestCheckSeatbeltAvailableMock(t *testing.T) {
	// sandbox-exec는 macOS에서만 사용 가능
	result := checkSeatbeltAvailable()

	// Linux에서는 false가 예상됨
	if runtime.GOOS == "linux" && result {
		t.Error("checkSeatbeltAvailable() on linux should return false")
	}

	// macOS에서는 시스템에 sandbox-exec가 있는지 확인
	if runtime.GOOS == "darwin" {
		t.Logf("checkSeatbeltAvailable() on darwin = %v", result)
	}
}

func TestCheckDockerAvailableMock(t *testing.T) {
	// CI 환경 변수 저장 및 복원
	originalCI := os.Getenv("CI")
	defer func() {
		if originalCI == "" {
			if err := os.Unsetenv("CI"); err != nil {
				t.Logf("failed to unset CI: %v", err)
			}
		} else {
			if err := os.Setenv("CI", originalCI); err != nil {
				t.Logf("failed to restore CI: %v", err)
			}
		}
	}()

	// CI가 설정되지 않은 시나리오
	if err := os.Unsetenv("CI"); err != nil {
		t.Fatalf("failed to unset CI: %v", err)
	}

	result := checkDockerAvailable()
	if result {
		t.Error("checkDockerAvailable() = true, want false when CI not set")
	}

	// CI가 설정된 시나리오 (실제 Docker 설치 여부는 시스템에 따라 다름)
	if err := os.Setenv("CI", "1"); err != nil {
		t.Fatalf("failed to set CI: %v", err)
	}
	result = checkDockerAvailable()

	t.Logf("checkDockerAvailable() with CI=1 = %v", result)
}

func TestValidateSandboxConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		opts    SandboxOptions
		wantErr bool
	}{
		{
			name: "valid configuration",
			opts: SandboxOptions{
				WorktreeRoot:   t.TempDir(),
				ReadOnlyDirs:   []string{"/"},
				WritablePaths:  []string{},
				NetworkAllow:   []string{},
				EnvPassthrough: []string{},
			},
			wantErr: false,
		},
		{
			name: "empty worktree root",
			opts: SandboxOptions{
				WorktreeRoot:   "",
				ReadOnlyDirs:   []string{"/"},
				WritablePaths:  []string{},
				NetworkAllow:   []string{},
				EnvPassthrough: []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid network allowlist",
			opts: SandboxOptions{
				WorktreeRoot:   t.TempDir(),
				ReadOnlyDirs:   []string{"/"},
				WritablePaths:  []string{},
				NetworkAllow:   []string{"example.com"},
				EnvPassthrough: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSandboxConfiguration(tt.opts)
			// For the "valid configuration" case, the recommended backend may not be installed
			// in CI or dev environments. Skip if the only error is backend unavailability.
			if !tt.wantErr && err != nil {
				if isErrSandboxRequired(err) {
					t.Skipf("skipping: %v", err)
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSandboxConfiguration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		want  string
	}{
		{
			name:  "normal path",
			path:  "/usr/bin",
			want:  "/usr/bin",
		},
		{
			name:  "path traversal attempt",
			path:  "../../../etc/passwd",
			want:  "/etc/passwd",
		},
		{
			name:  "windows path traversal",
			path:  "..\\..\\windows\\system32",
			want:  "windowssystem32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizePath(tt.path)
			// 경로 순회 패턴이 제거되었는지 확인
			if strings.Contains(got, "..") {
				t.Errorf("SanitizePath() = %v, want no '..' patterns", got)
			}
		})
	}
}

// isErrSandboxRequired reports whether err wraps ErrSandboxRequired.
// Used to skip tests when the recommended backend is simply not installed.
func isErrSandboxRequired(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), ErrSandboxRequired.Error())
}

func TestGetSandboxBackendForRole(t *testing.T) {
	tests := []struct {
		role string
	}{
		{"implementer"},
		{"tester"},
		{"designer"},
		{"reviewer"},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			backend := GetSandboxBackendForRole(tt.role)

			// 모든 롤은 OS별 권장 백엔드를 사용
			expected := DetectPlatform()
			if backend != expected {
				t.Errorf("GetSandboxBackendForRole() = %v, want %v", backend, expected)
			}
		})
	}
}
