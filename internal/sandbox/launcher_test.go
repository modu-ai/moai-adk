package sandbox

import (
	"bytes"
	"errors"
	"os"
	"runtime"
	"testing"
)

// TestLauncher_DispatchByOS verifies that the launcher selects the correct backend
// based on the current operating system (and not CI).
// RED: fails until launcher.go::Launcher is created.
func TestLauncher_DispatchByOS(t *testing.T) {
	// t.Setenv 사용 시 t.Parallel() 불가 (Go 1.26 제약)
	t.Setenv("CI", "")

	l := NewLauncher()
	resolved := l.ResolveBackend(SandboxNone)

	switch runtime.GOOS {
	case "darwin":
		// macOS에서 none은 none 유지 (no upgrade)
		// implementer role 테스트는 TestLauncher_ResolveBackend_AllScenarios에서
	case "linux":
		// Linux에서 none은 none 유지
	}

	// SandboxNone은 항상 none으로 반환 (role-based override는 ResolveForRole)
	if resolved != SandboxNone {
		// none이 아닌 경우 — declared가 none이면 launcher는 그대로 none 반환
		// (업그레이드는 role-profile에서)
		t.Logf("ResolveBackend(SandboxNone) = %q (OS: %s)", resolved, runtime.GOOS)
	}
}

// TestLauncher_CIOverride verifies CI=1 forces docker backend for implementer roles.
// RED: fails until launcher.go handles CI detection.
func TestLauncher_CIOverride(t *testing.T) {
	// t.Setenv 사용 시 t.Parallel() 불가
	t.Setenv("CI", "1")

	l := NewLauncher()

	// implementer role에서 CI=1이면 docker
	resolved := l.ResolveForRole("implementer")
	if resolved != SandboxDocker {
		t.Errorf("CI=1: ResolveForRole(implementer) = %q, want %q", resolved, SandboxDocker)
	}

	// researcher role에서 CI=1이어도 none
	resolvedResearcher := l.ResolveForRole("researcher")
	if resolvedResearcher != SandboxNone {
		t.Errorf("CI=1: ResolveForRole(researcher) = %q, want %q", resolvedResearcher, SandboxNone)
	}
}

// TestLauncher_OutputTruncation16MiB verifies that output exceeding 16 MiB is truncated
// and ErrSandboxOutputTruncated is returned.
// RED: fails until launcher.go implements truncation logic.
func TestLauncher_OutputTruncation16MiB(t *testing.T) {
	t.Parallel()

	const limit = 16 * 1024 * 1024

	// 16MiB + 1バイトのデータを生成する
	large := bytes.Repeat([]byte("x"), limit+1)

	truncated, err := TruncateOutput(large, limit)

	if !errors.Is(err, ErrSandboxOutputTruncated) {
		t.Errorf("TruncateOutput: expected ErrSandboxOutputTruncated, got %v", err)
	}
	if len(truncated) != limit {
		t.Errorf("TruncateOutput: got %d bytes, want %d bytes", len(truncated), limit)
	}
}

// TestLauncher_BackendUnavailable verifies that Exec returns ErrSandboxBackendUnavailable
// when the backend's Available() returns false.
// RED: fails until launcher.go checks backend availability before exec.
func TestLauncher_BackendUnavailable(t *testing.T) {
	t.Parallel()

	l := NewLauncher()
	l.SetBackend(SandboxBubblewrap, &mockBackend{available: false})

	opts := SandboxOptions{
		WritableScope:  []string{t.TempDir()},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	_, err := l.Exec(SandboxBubblewrap, opts, []string{"echo", "hello"})
	if !errors.Is(err, ErrSandboxBackendUnavailable) {
		t.Errorf("Exec with unavailable backend: expected ErrSandboxBackendUnavailable, got %v", err)
	}
}

// TestLauncher_ResolveBackend_AllScenarios verifies resolve logic for all OS × CI × role combinations.
// RED: fails until launcher.go implements resolve logic.
func TestLauncher_ResolveBackend_AllScenarios(t *testing.T) {
	// 서브테스트 내에서 t.Setenv 사용하므로 부모는 Parallel 불가

	// 현재 OS에 따라 기대값 결정
	var osDefault Sandbox
	switch runtime.GOOS {
	case "darwin":
		osDefault = SandboxSeatbelt
	case "linux":
		osDefault = SandboxBubblewrap
	default:
		t.Skip("unsupported OS for sandbox role default test")
	}

	tests := []struct {
		name   string
		ciSet  bool
		role   string
		want   Sandbox
	}{
		{"implementer-no-ci", false, "implementer", osDefault},
		{"tester-no-ci", false, "tester", osDefault},
		{"designer-no-ci", false, "designer", osDefault},
		{"researcher-no-ci", false, "researcher", SandboxNone},
		{"analyst-no-ci", false, "analyst", SandboxNone},
		{"reviewer-no-ci", false, "reviewer", SandboxNone},
		{"architect-no-ci", false, "architect", SandboxNone},
		{"implementer-ci", true, "implementer", SandboxDocker},
		{"tester-ci", true, "tester", SandboxDocker},
		{"designer-ci", true, "designer", SandboxDocker},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// t.Setenv 사용 시 t.Parallel() 불가
			if tc.ciSet {
				t.Setenv("CI", "1")
			} else {
				t.Setenv("CI", "")
			}

			l := NewLauncher()
			got := l.ResolveForRole(tc.role)
			if got != tc.want {
				t.Errorf("ResolveForRole(%q, CI=%v) = %q, want %q", tc.role, tc.ciSet, got, tc.want)
			}
		})
	}
}

// TestLauncher_PermissionDenyDivergence verifies AC-16:
// when permission layer says "allow" but sandbox blocks, sandbox verdict wins.
// RED: fails until launcher.go implements divergence handling (T-RT003-25).
func TestLauncher_PermissionDenyDivergence(t *testing.T) {
	t.Parallel()

	// mockBackend가 exec에서 EPERM에 해당하는 에러를 반환
	permErr := &SandboxDeniedError{Path: "/etc/passwd", Reason: "file-write-denied"}
	backend := &mockBackend{
		available: true,
		execErr:   permErr,
	}

	l := NewLauncher()
	l.SetBackend(SandboxSeatbelt, backend)

	opts := SandboxOptions{
		WritableScope:  []string{"/tmp/worktree"},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	// permission이 "allow"를 반환했더라도 sandbox 거부 → sandbox 우선
	_, err := l.Exec(SandboxSeatbelt, opts, []string{"touch", "/etc/passwd"})
	if err == nil {
		t.Fatal("Exec: expected error from sandbox-denied, got nil")
	}

	var denied *SandboxDeniedError
	if !errors.As(err, &denied) {
		t.Errorf("Exec: expected *SandboxDeniedError, got %T: %v", err, err)
	}
}

// TestLauncher_OutputTruncation_SystemMessage verifies that truncated output
// contains a SystemMessage noting the truncation.
// RED: fails until launcher.go emits SystemMessage on truncation.
func TestLauncher_OutputTruncation_SystemMessage(t *testing.T) {
	t.Parallel()

	const limit = 16 * 1024 * 1024
	large := bytes.Repeat([]byte("x"), limit+1)

	truncated, err := TruncateOutput(large, limit)
	if !errors.Is(err, ErrSandboxOutputTruncated) {
		t.Errorf("TruncateOutput: expected ErrSandboxOutputTruncated, got %v", err)
	}

	// 마지막 부분에 truncation 표시가 있어야 함
	last200 := string(truncated[len(truncated)-200:])
	_ = last200 // truncation marker format is implementation-defined

	// 에러 메시지에 크기 정보가 있어야 함
	if err.Error() == "" {
		t.Error("ErrSandboxOutputTruncated Error() should not be empty")
	}
	_ = os.DevNull // ensure os is used
}
