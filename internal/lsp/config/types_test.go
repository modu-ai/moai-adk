package config_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
)

// TestServerConfig_Defaults verifies zero-value ServerConfig is valid and safe to use.
func TestServerConfig_Defaults(t *testing.T) {
	t.Parallel()

	var sc config.ServerConfig
	// 기본값 검증: 비어 있는 ServerConfig은 zero-value여야 함
	if sc.Language != "" {
		t.Errorf("default Language = %q, want empty", sc.Language)
	}
	if sc.Command != "" {
		t.Errorf("default Command = %q, want empty", sc.Command)
	}
	if sc.IdleShutdownSeconds != 0 {
		t.Errorf("default IdleShutdownSeconds = %d, want 0", sc.IdleShutdownSeconds)
	}
	if len(sc.Args) != 0 {
		t.Errorf("default Args length = %d, want 0", len(sc.Args))
	}
	if len(sc.RootMarkers) != 0 {
		t.Errorf("default RootMarkers length = %d, want 0", len(sc.RootMarkers))
	}
	if len(sc.FileExtensions) != 0 {
		t.Errorf("default FileExtensions length = %d, want 0", len(sc.FileExtensions))
	}
	if sc.InitOptions != nil {
		t.Errorf("default InitOptions = %v, want nil", sc.InitOptions)
	}
}

// TestServerConfig_Fields verifies all exported fields of ServerConfig can be set.
func TestServerConfig_Fields(t *testing.T) {
	t.Parallel()

	sc := config.ServerConfig{
		Language:            "go",
		Command:             "gopls",
		Args:                []string{"serve", "-rpc.trace"},
		InitOptions:         map[string]any{"staticcheck": true},
		IdleShutdownSeconds: 600,
		RootMarkers:         []string{"go.mod", "go.sum"},
		FileExtensions:      []string{".go"},
	}

	if sc.Language != "go" {
		t.Errorf("Language = %q, want %q", sc.Language, "go")
	}
	if sc.Command != "gopls" {
		t.Errorf("Command = %q, want %q", sc.Command, "gopls")
	}
	if len(sc.Args) != 2 {
		t.Errorf("Args length = %d, want 2", len(sc.Args))
	}
	if sc.IdleShutdownSeconds != 600 {
		t.Errorf("IdleShutdownSeconds = %d, want 600", sc.IdleShutdownSeconds)
	}
	if len(sc.RootMarkers) != 2 {
		t.Errorf("RootMarkers length = %d, want 2", len(sc.RootMarkers))
	}
	if len(sc.FileExtensions) != 1 {
		t.Errorf("FileExtensions length = %d, want 1", len(sc.FileExtensions))
	}
	v, ok := sc.InitOptions["staticcheck"]
	if !ok {
		t.Error("InitOptions missing key 'staticcheck'")
	}
	if v != true {
		t.Errorf("InitOptions[staticcheck] = %v, want true", v)
	}
}

// TestServersConfig_EmptyServers verifies empty servers map is valid.
func TestServersConfig_EmptyServers(t *testing.T) {
	t.Parallel()

	sc := config.ServersConfig{
		Servers: nil,
	}
	if sc.Servers != nil {
		t.Errorf("nil Servers = %v, want nil", sc.Servers)
	}
}

// TestServersConfig_MultipleServers verifies multiple servers can be registered.
func TestServersConfig_MultipleServers(t *testing.T) {
	t.Parallel()

	sc := config.ServersConfig{
		Servers: map[string]config.ServerConfig{
			"go": {
				Language: "go",
				Command:  "gopls",
			},
			"python": {
				Language: "python",
				Command:  "pylsp",
			},
			"typescript": {
				Language: "typescript",
				Command:  "typescript-language-server",
				Args:     []string{"--stdio"},
			},
		},
	}

	if len(sc.Servers) != 3 {
		t.Errorf("Servers count = %d, want 3", len(sc.Servers))
	}

	for _, lang := range []string{"go", "python", "typescript"} {
		if _, ok := sc.Servers[lang]; !ok {
			t.Errorf("Servers missing key %q", lang)
		}
	}
}
