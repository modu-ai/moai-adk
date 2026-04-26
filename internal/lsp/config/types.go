package config

// ServerConfig holds configuration for a single language server (REQ-LC-003).
//
// Fields map to lsp.servers.<language> entries in lsp.yaml.
// @MX:ANCHOR: [AUTO] ServerConfig is the central config type consumed by subprocess.Launcher, core.Client, and Manager
// @MX:REASON: fan_in >= 3 — subprocess launcher, core client, and manager all read ServerConfig fields
type ServerConfig struct {
	// Language is the canonical language identifier (e.g., "go", "python", "typescript").
	// Set by the loader from the map key; not present in YAML.
	Language string `yaml:"-"`

	// Command is the language server binary name (e.g., "gopls", "pylsp").
	Command string `yaml:"command"`

	// Args are additional command-line arguments passed to the server binary.
	Args []string `yaml:"args"`

	// InitOptions are merged into the LSP initialize request as initializationOptions (REQ-LC-003a).
	// Values may be any YAML-compatible type: bool, int, float, string, or nested map.
	InitOptions map[string]any `yaml:"init_options"`

	// IdleShutdownSeconds is the inactivity timeout after which the server is shut down (REQ-LC-050).
	// Zero means no automatic idle shutdown.
	IdleShutdownSeconds int `yaml:"idle_shutdown_seconds"`

	// RootMarkers are filenames that identify the project root for this language (e.g., "go.mod").
	RootMarkers []string `yaml:"root_markers"`

	// FileExtensions are the file suffixes (including the dot) that this server handles (e.g., ".go").
	FileExtensions []string `yaml:"file_extensions"`

	// RootDir is the workspace root directory sent as rootUri in the LSP initialize request.
	// When empty, rootUri is omitted (nil) from the initialize request.
	// Set this to the project root when starting a client for a specific workspace.
	RootDir string `yaml:"-"`

	// InstallHint is a human-readable install command shown when the binary is missing (REQ-LM-004).
	// Example: "go install golang.org/x/tools/gopls@latest"
	InstallHint string `yaml:"install_hint"`

	// FallbackBinaries is an ordered list of alternative server binaries to try when
	// the primary Command is not found in PATH (REQ-LM-008).
	// Tried in order; the first found binary is used.
	FallbackBinaries []string `yaml:"fallback_binaries"`

	// ProjectMarkers are filenames whose presence in the project root signals that
	// this language server should be activated (REQ-LM-001, REQ-LM-002).
	// Distinct from RootMarkers: RootMarkers identify the workspace root for LSP,
	// while ProjectMarkers control whether the server is spawned at all.
	ProjectMarkers []string `yaml:"project_markers"`
}

// ServersConfig is the root deserialization type for the lsp.servers section.
// Servers maps language name → ServerConfig.
// @MX:ANCHOR: [AUTO] ServersConfig — top-level config type returned by Load() and used as a Manager initialization argument
// @MX:REASON: fan_in >= 3 — callers of config.Load (Manager, CLI, tests) all reference this type
type ServersConfig struct {
	// ClientImpl selects which LSP client implementation handles language
	// server communication (SPEC-LSP-CORE-002 AC10, REQ-LC-010).
	//
	// Accepted values: "gopls_bridge" (legacy Go-only bridge, SPEC-GOPLS-BRIDGE-001)
	// and "powernap_core" (multi-language foundation, SPEC-LSP-CORE-002).
	// Empty string is treated as the default "gopls_bridge".
	//
	// Runtime toggling without restart is an explicit SPEC requirement;
	// Manager.Reload() consumers should re-read this value when the file
	// changes. The initial value is captured at startup.
	ClientImpl string `yaml:"client_impl"`

	// Servers maps language identifier to its configuration.
	Servers map[string]ServerConfig
}

// Valid LSP client implementations (REQ-LC-010).
const (
	ClientImplGoplsBridge = "gopls_bridge"
	ClientImplPowernapCore = "powernap_core"
)

// ResolveClientImpl returns the effective client implementation, falling back
// to the default "gopls_bridge" when the config value is empty or unknown.
// @MX:NOTE: [AUTO] ResolveClientImpl — default resolution for SPEC-LSP-CORE-002 AC10 Feature Flag
func (c *ServersConfig) ResolveClientImpl() string {
	if c == nil {
		return ClientImplGoplsBridge
	}
	switch c.ClientImpl {
	case ClientImplGoplsBridge, ClientImplPowernapCore:
		return c.ClientImpl
	default:
		return ClientImplGoplsBridge
	}
}
