package sandbox

import (
	"testing"
)

// TestSandbox_EnumExhaustive verifies that the Sandbox type has exactly 4 valid values.
// RED: fails until Sandbox enum is defined in context.go.
func TestSandbox_EnumExhaustive(t *testing.T) {
	t.Parallel()

	// 4개 열거값 모두 검증
	validValues := []Sandbox{
		SandboxNone,
		SandboxBubblewrap,
		SandboxSeatbelt,
		SandboxDocker,
	}

	seen := make(map[Sandbox]bool)
	for _, v := range validValues {
		if seen[v] {
			t.Errorf("duplicate sandbox value: %q", v)
		}
		seen[v] = true

		// 각 값이 비어있지 않아야 함
		if string(v) == "" {
			t.Error("sandbox value must not be empty string")
		}
	}

	if len(seen) != 4 {
		t.Errorf("expected exactly 4 sandbox values, got %d", len(seen))
	}

	// 각 값의 문자열 표현 확인
	if string(SandboxNone) != "none" {
		t.Errorf("SandboxNone: got %q, want %q", SandboxNone, "none")
	}
	if string(SandboxBubblewrap) != "bubblewrap" {
		t.Errorf("SandboxBubblewrap: got %q, want %q", SandboxBubblewrap, "bubblewrap")
	}
	if string(SandboxSeatbelt) != "seatbelt" {
		t.Errorf("SandboxSeatbelt: got %q, want %q", SandboxSeatbelt, "seatbelt")
	}
	if string(SandboxDocker) != "docker" {
		t.Errorf("SandboxDocker: got %q, want %q", SandboxDocker, "docker")
	}
}

// TestSandboxOptions_Validate verifies that SandboxOptions holds all required fields.
// RED: fails until SandboxOptions is defined in context.go.
func TestSandboxOptions_Validate(t *testing.T) {
	t.Parallel()

	opts := SandboxOptions{
		WritableScope:    []string{"/tmp/worktree"},
		ReadOnlyScope:    []string{"/usr", "/lib"},
		NetworkAllowlist: []string{"github.com", "pypi.org"},
		EnvPassthrough:   []string{"GH_TOKEN"},
		MaxOutputBytes:   16 * 1024 * 1024, // 16 MiB
	}

	if len(opts.WritableScope) == 0 {
		t.Error("WritableScope must not be empty")
	}
	if opts.MaxOutputBytes != 16*1024*1024 {
		t.Errorf("MaxOutputBytes: got %d, want %d", opts.MaxOutputBytes, 16*1024*1024)
	}
	if len(opts.NetworkAllowlist) == 0 {
		t.Error("NetworkAllowlist must not be empty")
	}
}

// TestSandboxBackend_InterfaceContract verifies that the SandboxBackend interface
// is satisfied by a simple mock (ensuring the interface exists with correct methods).
// RED: fails until SandboxBackend interface is defined in context.go.
func TestSandboxBackend_InterfaceContract(t *testing.T) {
	t.Parallel()

	// mockBackend implements SandboxBackend for contract verification
	var _ SandboxBackend = (*mockBackend)(nil)
}

// mockBackend is a test double implementing SandboxBackend.
type mockBackend struct {
	available bool
	execErr   error
	output    []byte
}

func (m *mockBackend) Available() bool {
	return m.available
}

func (m *mockBackend) Exec(opts SandboxOptions, cmd []string) ([]byte, error) {
	if m.execErr != nil {
		return nil, m.execErr
	}
	return m.output, nil
}

func (m *mockBackend) Profile(opts SandboxOptions) (string, error) {
	return "mock-profile", nil
}
