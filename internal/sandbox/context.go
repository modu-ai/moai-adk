// Package sandbox provides an ephemeral sandbox execution layer that wraps
// implementer-agent tool invocations in OS-appropriate isolation primitives.
//
// Backends:
//   - Linux: Bubblewrap (bwrap)
//   - macOS: Seatbelt (sandbox-exec)
//   - CI:    Docker
//
// This is the third defense-in-depth layer after RT-001 (audit trail) and
// RT-002 (permission envelope), addressing OWASP Top 10 for Agentic Apps 2025
// and the Claude Code `rm -rf ~/` class of incidents.
//
// @MX:ANCHOR: [AUTO] Sandbox enum is the primary type contract for all backends
// @MX:REASON: Fan_in >= 3: used by launcher.go, doctor_sandbox.go, agent_lint.go,
//             config/types.go — any rename breaks callers across 4+ packages.
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-001/002
package sandbox

// Sandbox is a typed string enum representing the sandbox execution backend.
// Exactly 4 values are valid per REQ-V3R2-RT-003-001.
type Sandbox string

const (
	// SandboxNone disables sandboxing. Permitted only with explicit justification.
	SandboxNone Sandbox = "none"
	// SandboxBubblewrap uses Linux user-namespace-based sandboxing via bwrap.
	SandboxBubblewrap Sandbox = "bubblewrap"
	// SandboxSeatbelt uses macOS sandbox-exec with SBPL profiles.
	SandboxSeatbelt Sandbox = "seatbelt"
	// SandboxDocker uses ephemeral Docker containers (CI fallback).
	SandboxDocker Sandbox = "docker"
)

// DefaultNetworkAllowlist is the pre-seeded list of allowed network hosts
// per REQ-V3R2-RT-003-008.
var DefaultNetworkAllowlist = []string{
	"github.com",
	"registry.npmjs.org",
	"pypi.org",
	"proxy.golang.org",
	"crates.io",
	"repo.maven.apache.org",
	"rubygems.org",
	"pub.dev",
}

// DefaultMaxOutputBytes is the maximum output size before truncation (16 MiB).
// Per REQ-V3R2-RT-003-042.
const DefaultMaxOutputBytes = 16 * 1024 * 1024

// SandboxOptions holds the configuration for a single sandbox invocation.
//
// @MX:ANCHOR: [AUTO] SandboxOptions is consumed by every backend Exec call
// @MX:REASON: Fan_in >= 3: BubblewrapBackend.Exec, SeatbeltBackend.Exec, DockerBackend.Exec
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-006/007/008/031
type SandboxOptions struct {
	// WritableScope is the list of filesystem paths the sandboxed process may write to.
	// Defaults to agent worktree root + .moai/state/ per REQ-V3R2-RT-003-007.
	WritableScope []string

	// ReadOnlyScope is the list of filesystem paths mounted read-only inside the sandbox.
	ReadOnlyScope []string

	// NetworkAllowlist is the list of allowed outbound hosts. Appended to DefaultNetworkAllowlist.
	// Empty list means no additional hosts (default allowlist still applies unless explicitly
	// set to all-deny by the backend). Per REQ-V3R2-RT-003-008 and REQ-V3R2-RT-003-030.
	NetworkAllowlist []string

	// EnvPassthrough lists environment variable names that should survive the default
	// scrubbing pass. Per REQ-V3R2-RT-003-031 and AC-V3R2-RT-003-07.
	EnvPassthrough []string

	// MaxOutputBytes is the maximum combined stdout+stderr before truncation.
	// Defaults to DefaultMaxOutputBytes (16 MiB). Per REQ-V3R2-RT-003-042.
	MaxOutputBytes int64

	// PlanMode indicates the agent is in read-only plan mode. When true, every
	// filesystem path is mounted read-only (writable scope is ignored).
	// Per REQ-V3R2-RT-003-022.
	PlanMode bool

	// DockerImage overrides the default Docker image for the docker backend.
	DockerImage string

	// Justification is the opt-out explanation when Sandbox == SandboxNone.
	// Required when security.yaml sandbox.required = true.
	Justification string
}

// SandboxBackend is the interface implemented by each OS-specific sandbox backend.
//
// @MX:ANCHOR: [AUTO] SandboxBackend is the polymorphism contract for all OS backends
// @MX:REASON: Fan_in >= 3: BubblewrapBackend, SeatbeltBackend, DockerBackend, mockBackend in tests
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-002
type SandboxBackend interface {
	// Available reports whether the backend binary is present and usable on the
	// current host. Must NOT perform any network requests.
	Available() bool

	// Exec runs cmd inside the sandbox with the given options. Returns combined
	// stdout+stderr up to opts.MaxOutputBytes. Returns ErrSandboxBackendUnavailable
	// if Available() is false. Returns ErrSandboxOutputTruncated (alongside truncated
	// output) if the output exceeds opts.MaxOutputBytes.
	Exec(opts SandboxOptions, cmd []string) ([]byte, error)

	// Profile returns a human-readable representation of the sandbox profile that
	// would be applied for the given options. Used by `moai doctor sandbox --profile`.
	Profile(opts SandboxOptions) (string, error)
}
