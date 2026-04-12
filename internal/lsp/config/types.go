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
}

// ServersConfig is the root deserialization type for the lsp.servers section.
// Servers maps language name → ServerConfig.
// @MX:ANCHOR: [AUTO] ServersConfig 최상위 설정 타입 — Load() 반환값이자 Manager 초기화 인자
// @MX:REASON: fan_in >= 3 — config.Load 호출자들(Manager, CLI, 테스트)이 이 타입을 모두 참조
type ServersConfig struct {
	// Servers maps language identifier to its configuration.
	Servers map[string]ServerConfig
}
