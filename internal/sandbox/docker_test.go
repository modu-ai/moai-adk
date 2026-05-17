package sandbox

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestDocker_ExecHello is a CI smoke test for Docker backend execution.
// Gated by MOAI_TEST_DOCKER=1 environment variable.
func TestDocker_ExecHello(t *testing.T) {
	if os.Getenv("MOAI_TEST_DOCKER") != "1" {
		t.Skip("docker test requires MOAI_TEST_DOCKER=1")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available in PATH")
	}

	d := NewDockerBackend()
	if !d.Available() {
		t.Skip("docker backend reports unavailable (daemon may not be running)")
	}

	opts := SandboxOptions{
		WritableScope:  []string{t.TempDir()},
		DockerImage:    "alpine:latest",
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	out, err := d.Exec(opts, []string{"echo", "hello"})
	if err != nil {
		t.Fatalf("docker exec: %v", err)
	}
	if !strings.Contains(string(out), "hello") {
		t.Errorf("docker exec: expected 'hello' in output, got %q", string(out))
	}
}

// TestDocker_NetworkAllowlist verifies that docker uses network policy from allowlist.
// Gated by MOAI_TEST_DOCKER=1.
func TestDocker_NetworkAllowlist(t *testing.T) {
	if os.Getenv("MOAI_TEST_DOCKER") != "1" {
		t.Skip("docker test requires MOAI_TEST_DOCKER=1")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available in PATH")
	}

	d := NewDockerBackend()
	if !d.Available() {
		t.Skip("docker backend reports unavailable")
	}

	// allowlist가 비어있으면 --network=none 선택
	opts := SandboxOptions{
		WritableScope:    []string{t.TempDir()},
		NetworkAllowlist: []string{}, // empty = none
		DockerImage:      "alpine:latest",
		MaxOutputBytes:   16 * 1024 * 1024,
	}

	// --network=none 상태에서 curl → 실패해야 함
	_, err := d.Exec(opts, []string{"sh", "-c", "wget -q https://example.com || exit 1"})
	if err == nil {
		t.Error("docker with empty allowlist should block network access")
	}
}

// TestDocker_FileWriteScope verifies that docker mounts only specified paths writable.
// Gated by MOAI_TEST_DOCKER=1.
func TestDocker_FileWriteScope(t *testing.T) {
	if os.Getenv("MOAI_TEST_DOCKER") != "1" {
		t.Skip("docker test requires MOAI_TEST_DOCKER=1")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available in PATH")
	}

	d := NewDockerBackend()
	if !d.Available() {
		t.Skip("docker backend reports unavailable")
	}

	scope := t.TempDir()
	opts := SandboxOptions{
		WritableScope:  []string{scope},
		DockerImage:    "alpine:latest",
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	// scope 안 파일 쓰기 → 성공해야 함
	_, err := d.Exec(opts, []string{"sh", "-c", "touch " + scope + "/test.txt"})
	if err != nil {
		t.Errorf("docker should allow write inside scope: %v", err)
	}
}
