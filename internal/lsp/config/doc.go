// Package config provides types and a YAML loader for LSP server configuration.
//
// Configuration is read from .moai/config/sections/lsp.yaml under the
// lsp.servers.<language> key (REQ-LC-003).
//
// Usage:
//
//	cfg, err := config.Load(".moai/config/sections/lsp.yaml")
//	if err != nil {
//	    // handle error
//	}
//	goServer, ok := cfg.Servers["go"]
package config
