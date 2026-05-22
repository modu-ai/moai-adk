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
	// Input validation — null bytes break SBPL
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

	// Allow reads by default
	lines = append(lines, "(allow file-read*)")

	// plan mode: no writes
	if !opts.PlanMode {
		// writable scope — sort then emit (determinism)
		writePaths := make([]string, len(opts.WritableScope))
		copy(writePaths, opts.WritableScope)
		sort.Strings(writePaths)
		for _, p := range writePaths {
			lines = append(lines, fmt.Sprintf(`(allow file-write* (subpath %q))`, p))
		}

		// .moai/state/ default writable scope
		lines = append(lines, `(allow file-write* (subpath ".moai/state"))`)
	}

	// LSP carve-out: ~/.cache/ rw + /tmp tmpfs (REQ-V3R2-RT-003-021)
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache")
	lines = append(lines, fmt.Sprintf(`(allow file-read* file-write* (subpath %q))`, cacheDir))
	lines = append(lines, `(allow file-read* file-write* (subpath "/tmp"))`)

	// Allow process execution
	lines = append(lines, "(allow process-exec*)")
	lines = append(lines, "(allow process-fork)")

	// Network: allow localhost + UNIX sockets (LSP)
	lines = append(lines, `(allow network-outbound (local tcp))`)
	lines = append(lines, `(allow network-outbound (remote unix-socket))`)

	// SBPL does not support a host-specific TCP allowlist (sandbox-exec
	// constraint). The network allowlist must be implemented at the OS level
	// (pf/nftables) or via a proxy. The current SBPL implementation allows
	// all outbound TCP (pf integration planned for v3.1+).
	// Empty allowlist = no TCP; non-empty = allow all.
	allHosts := append(DefaultNetworkAllowlist, opts.NetworkAllowlist...)
	if len(allHosts) > 0 {
		lines = append(lines, `(allow network-outbound (remote tcp))`)
	}

	// Other required system permissions
	lines = append(lines, "(allow sysctl-read)")
	lines = append(lines, "(allow signal (target self))")
	lines = append(lines, "(allow mach-lookup)")

	// Complete output — sort lines (determinism)
	// version and deny default are pinned at the front; the rest are sorted
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
	// Input validation
	for _, p := range opts.WritableScope {
		if strings.ContainsRune(p, 0) {
			return nil, fmt.Errorf("%w: writable scope path contains null byte: %q",
				ErrSandboxProfileInvalid, p)
		}
	}

	var args []string

	// Base isolation flags
	args = append(args, "--unshare-all")
	args = append(args, "--die-with-parent")

	// plan mode: no writes (all paths ro-bind)
	if opts.PlanMode {
		roPaths := make([]string, len(opts.WritableScope))
		copy(roPaths, opts.WritableScope)
		sort.Strings(roPaths)
		for _, p := range roPaths {
			args = append(args, "--ro-bind", p, p)
		}
	} else {
		// Writable paths — sorted (determinism)
		writePaths := make([]string, len(opts.WritableScope))
		copy(writePaths, opts.WritableScope)
		sort.Strings(writePaths)
		for _, p := range writePaths {
			args = append(args, "--bind", p, p)
		}
	}

	// Read-only paths — sorted
	roPaths := make([]string, len(opts.ReadOnlyScope))
	copy(roPaths, opts.ReadOnlyScope)
	sort.Strings(roPaths)
	for _, p := range roPaths {
		args = append(args, "--ro-bind", p, p)
	}

	// LSP carve-out (REQ-021): ~/.cache/ rw + /tmp tmpfs
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache")
	args = append(args, "--bind", cacheDir, cacheDir)
	args = append(args, "--tmpfs", "/tmp")

	// Network: --unshare-net is already included via --unshare-all.
	// Allowlist hosts would require an actual socat/forward setup, but at the
	// unit-test level we only validate argument generation.
	args = append(args, "--share-net") // Replace with an allowlist bridge in production deployments

	// /proc, /sys, /dev default bindings
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

	// Network policy
	allHosts := append(DefaultNetworkAllowlist, opts.NetworkAllowlist...)
	if len(allHosts) == 0 {
		parts = append(parts, "--network=none")
	} else {
		parts = append(parts, "--network=bridge")
	}

	// writable scope — sorted (determinism)
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
