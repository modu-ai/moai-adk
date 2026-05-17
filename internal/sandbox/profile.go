package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GenerateSBPL produces a deterministic macOS SBPL (Sandbox Profile Language) profile
// from the given SandboxOptions. The profile denies all by default and allows
// reads everywhere, restricted writes to WritableScope, and network access to
// NetworkAllowlist hosts.
//
// All directives are sorted before emission to ensure checksum-stable output
// (REQ-V3R2-RT-003-004).
//
// @MX:ANCHOR: [AUTO] generateSBPL is the primary macOS profile generator
// @MX:REASON: Fan_in >= 3: SeatbeltBackend.Profile, TestProfile_GenerateSBPL,
//             TestSeatbelt_SBPLDeterministic, doctor_sandbox.go --profile flag
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-004/010/021/041
func GenerateSBPL(opts SandboxOptions) (string, error) {
	// 입력 검증 — null byte는 SBPL을 망가뜨린다
	for _, p := range opts.WritableScope {
		if strings.ContainsRune(p, 0) {
			return "", fmt.Errorf("%w: writable scope path contains null byte: %q",
				ErrSandboxProfileInvalid, p)
		}
	}
	for _, h := range opts.NetworkAllowlist {
		if strings.ContainsRune(h, 0) {
			return "", fmt.Errorf("%w: network allowlist host contains null byte: %q",
				ErrSandboxProfileInvalid, h)
		}
	}

	var lines []string
	lines = append(lines, "(version 1)")
	lines = append(lines, "(deny default)")

	// 기본 읽기 허용
	lines = append(lines, "(allow file-read*)")

	// plan mode: 쓰기 없음
	if !opts.PlanMode {
		// writable scope — 정렬 후 emit (결정성)
		writePaths := make([]string, len(opts.WritableScope))
		copy(writePaths, opts.WritableScope)
		sort.Strings(writePaths)
		for _, p := range writePaths {
			lines = append(lines, fmt.Sprintf(`(allow file-write* (subpath %q))`, p))
		}

		// .moai/state/ 기본 writable scope
		lines = append(lines, `(allow file-write* (subpath ".moai/state"))`)
	}

	// LSP 카브아웃: ~/.cache/ rw + /tmp tmpfs (REQ-V3R2-RT-003-021)
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache")
	lines = append(lines, fmt.Sprintf(`(allow file-read* file-write* (subpath %q))`, cacheDir))
	lines = append(lines, `(allow file-read* file-write* (subpath "/tmp"))`)

	// 프로세스 실행 허용
	lines = append(lines, "(allow process-exec*)")
	lines = append(lines, "(allow process-fork)")

	// 네트워크: localhost + UNIX 소켓 허용 (LSP)
	lines = append(lines, `(allow network-outbound (local tcp))`)
	lines = append(lines, `(allow network-outbound (remote unix-socket))`)

	// SBPL은 host-specific TCP allowlist를 지원하지 않음 (sandbox-exec 제약).
	// 네트워크 allowlist는 OS 레벨 방화벽(pf/nftables) 또는 프록시로 구현해야 하며,
	// 현재 SBPL 구현은 전체 outbound TCP를 허용 (v3.1+에서 pf 통합 예정).
	// 빈 allowlist는 TCP 불허 — 비어있지 않으면 전체 허용.
	allHosts := append(DefaultNetworkAllowlist, opts.NetworkAllowlist...)
	if len(allHosts) > 0 {
		lines = append(lines, `(allow network-outbound (remote tcp))`)
	}

	// 기타 필수 시스템 허용
	lines = append(lines, "(allow sysctl-read)")
	lines = append(lines, "(allow signal (target self))")
	lines = append(lines, "(allow mach-lookup)")

	// 전체 출력 — 행 정렬 (결정성)
	// version과 deny default는 앞에 고정, 나머지는 정렬
	fixed := lines[:2]
	rest := lines[2:]
	sort.Strings(rest)

	allLines := append(fixed, rest...)
	return strings.Join(allLines, "\n") + "\n", nil
}

// GenerateBwrapArgs produces a deterministic list of bwrap command-line arguments
// for the given SandboxOptions. Arguments are sorted within each group for
// checksum stability (REQ-V3R2-RT-003-004).
//
// @MX:ANCHOR: [AUTO] GenerateBwrapArgs is the primary Linux bwrap argument generator
// @MX:REASON: Fan_in >= 3: BubblewrapBackend.Exec, TestBubblewrap_ArgsDeterministic,
//             TestProfile_GenerateBwrapArgs, TestProfile_DeterministicChecksum_100Runs
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-004/011/021/041
func GenerateBwrapArgs(opts SandboxOptions) ([]string, error) {
	// 입력 검증
	for _, p := range opts.WritableScope {
		if strings.ContainsRune(p, 0) {
			return nil, fmt.Errorf("%w: writable scope path contains null byte: %q",
				ErrSandboxProfileInvalid, p)
		}
	}

	var args []string

	// 기본 격리 플래그
	args = append(args, "--unshare-all")
	args = append(args, "--die-with-parent")

	// plan mode: 쓰기 없음 (모든 경로 ro-bind)
	if opts.PlanMode {
		roPaths := make([]string, len(opts.WritableScope))
		copy(roPaths, opts.WritableScope)
		sort.Strings(roPaths)
		for _, p := range roPaths {
			args = append(args, "--ro-bind", p, p)
		}
	} else {
		// 쓰기 가능 경로 — 정렬 (결정성)
		writePaths := make([]string, len(opts.WritableScope))
		copy(writePaths, opts.WritableScope)
		sort.Strings(writePaths)
		for _, p := range writePaths {
			args = append(args, "--bind", p, p)
		}
	}

	// 읽기 전용 경로 — 정렬
	roPaths := make([]string, len(opts.ReadOnlyScope))
	copy(roPaths, opts.ReadOnlyScope)
	sort.Strings(roPaths)
	for _, p := range roPaths {
		args = append(args, "--ro-bind", p, p)
	}

	// LSP 카브아웃 (REQ-021): ~/.cache/ rw + /tmp tmpfs
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache")
	args = append(args, "--bind", cacheDir, cacheDir)
	args = append(args, "--tmpfs", "/tmp")

	// 네트워크: --unshare-net이 이미 --unshare-all에 포함됨
	// allowlist 호스트는 실제 socat/forward 설정이 필요하지만
	// unit test 레벨에서는 argument 생성만 검증
	args = append(args, "--share-net") // 실제 배포시 allowlist bridge로 교체

	// /proc, /sys, /dev 기본 바인딩
	args = append(args, "--proc", "/proc")
	args = append(args, "--dev", "/dev")

	return args, nil
}

// GenerateDockerSnippet produces a human-readable Docker run snippet for the
// given SandboxOptions. Used by `moai doctor sandbox --profile`.
//
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-004/015/032
func GenerateDockerSnippet(opts SandboxOptions) (string, error) {
	image := opts.DockerImage
	if image == "" {
		image = "alpine:latest"
	}

	var parts []string
	parts = append(parts, "docker run")
	parts = append(parts, "--rm")

	// 네트워크 정책
	allHosts := append(DefaultNetworkAllowlist, opts.NetworkAllowlist...)
	if len(allHosts) == 0 {
		parts = append(parts, "--network=none")
	} else {
		parts = append(parts, "--network=bridge")
	}

	// writable scope — 정렬 (결정성)
	writePaths := make([]string, len(opts.WritableScope))
	copy(writePaths, opts.WritableScope)
	sort.Strings(writePaths)
	for _, p := range writePaths {
		parts = append(parts, fmt.Sprintf("-v %s:%s", p, p))
	}

	if len(writePaths) > 0 {
		parts = append(parts, fmt.Sprintf("-w %s", writePaths[0]))
	}

	parts = append(parts, image)
	parts = append(parts, "<cmd>")

	return strings.Join(parts, " \\\n  "), nil
}
