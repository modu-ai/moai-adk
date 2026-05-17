package sandbox

import (
	"os/exec"
	"runtime"
	"testing"
)

// BenchmarkSandbox_SeatbeltHello measures seatbelt startup + execution latency.
// AC-V3R2-RT-003-17 (macOS): p95 must be <= 50ms.
// Gated on macOS.
func BenchmarkSandbox_SeatbeltHello(b *testing.B) {
	if runtime.GOOS != "darwin" {
		b.Skip("seatbelt benchmark requires macOS")
	}
	if _, err := exec.LookPath("sandbox-exec"); err != nil {
		b.Skip("sandbox-exec not available")
	}

	s := NewSeatbeltBackend()
	if !s.Available() {
		b.Skip("seatbelt backend unavailable")
	}

	opts := SandboxOptions{
		WritableScope:  []string{b.TempDir()},
		MaxOutputBytes: DefaultMaxOutputBytes,
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = s.Exec(opts, []string{"echo", "hello"})
	}
}

// BenchmarkSandbox_BwrapHello measures bubblewrap startup + execution latency.
// AC-V3R2-RT-003-18 (Linux): p95 must be <= 50ms.
// Gated on Linux.
func BenchmarkSandbox_BwrapHello(b *testing.B) {
	if runtime.GOOS != "linux" {
		b.Skip("bubblewrap benchmark requires Linux")
	}
	if _, err := exec.LookPath("bwrap"); err != nil {
		b.Skip("bwrap not available")
	}

	bw := NewBubblewrapBackend()
	if !bw.Available() {
		b.Skip("bubblewrap backend unavailable")
	}

	opts := SandboxOptions{
		WritableScope:  []string{b.TempDir()},
		MaxOutputBytes: DefaultMaxOutputBytes,
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = bw.Exec(opts, []string{"echo", "hello"})
	}
}

// BenchmarkSandbox_DockerHello measures Docker container startup + execution latency.
// AC-V3R2-RT-003-19 (CI): p99 must be <= 5s (CI-only, startup cost is acceptable).
// Gated on MOAI_TEST_DOCKER=1.
func BenchmarkSandbox_DockerHello(b *testing.B) {
	if _, err := exec.LookPath("docker"); err != nil {
		b.Skip("docker not available")
	}

	d := NewDockerBackend()
	if !d.Available() {
		b.Skip("docker backend unavailable (daemon may not be running)")
	}

	opts := SandboxOptions{
		WritableScope:  []string{b.TempDir()},
		DockerImage:    "alpine:latest",
		MaxOutputBytes: DefaultMaxOutputBytes,
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = d.Exec(opts, []string{"echo", "hello"})
	}
}
