package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// lspYAMLRoot is the top-level YAML structure matching lsp.yaml.
// Only the lsp.client_impl + lsp.servers sections are consumed; other
// top-level keys are ignored.
type lspYAMLRoot struct {
	LSP struct {
		ClientImpl string                  `yaml:"client_impl"`
		Servers    map[string]ServerConfig `yaml:"servers"`
	} `yaml:"lsp"`
}

// Load reads the lsp.yaml file at path and returns a ServersConfig.
//
// The path should be an absolute path or relative to the process working directory
// (typically the project root). Returns an error if the file cannot be read or
// if the YAML is malformed.
//
// REQ-LC-003: reads lsp.servers.<language> entries.
// @MX:ANCHOR: [AUTO] Load — primary entry point for LSP config; all server startup paths call this
// @MX:REASON: fan_in >= 3 — Manager.New, subprocess.Launcher, and integration tests all invoke Load
func Load(path string) (*ServersConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config.Load: read %q: %w", path, err)
	}

	var root lspYAMLRoot
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("config.Load: parse %q: %w", path, err)
	}

	servers := root.LSP.Servers
	if servers == nil {
		servers = make(map[string]ServerConfig)
	}

	// 각 ServerConfig에 Language 필드를 키에서 역주입 (YAML 키 → 구조체)
	for lang, sc := range servers {
		sc.Language = lang
		servers[lang] = sc
	}

	return &ServersConfig{
		ClientImpl: root.LSP.ClientImpl,
		Servers:    servers,
	}, nil
}

// MergeInitOptions merges src map entries into dst and returns the merged result.
//
// Keys present in both dst and src: src value takes precedence.
// Neither argument is mutated; a new map is always returned.
//
// REQ-LC-003a: per-language init_options are merged into the initialize request.
func MergeInitOptions(dst, src map[string]any) map[string]any {
	result := make(map[string]any, len(dst)+len(src))
	for k, v := range dst {
		result[k] = v
	}
	for k, v := range src {
		result[k] = v
	}
	return result
}
