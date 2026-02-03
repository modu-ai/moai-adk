// Package template provides template deployment and rendering for MoAI projects.
//
// The templates/ subdirectory contains curated template content that is embedded
// into the moai binary at compile time via //go:embed. This includes agent
// definitions, skill files, rules, output styles, configuration references,
// and root files (CLAUDE.md, .gitignore).
//
// Runtime-generated files (settings.json, .mcp.json, .lsp.json) are
// intentionally excluded from the embedded templates per ADR-011
// (Zero Runtime Template Expansion) and AD-001 (Go compiled hooks).
// These files are generated programmatically via Go struct serialization
// in settings.go (SettingsGenerator, MCPGenerator).
package template

import (
	"embed"
	"io/fs"
)

// embeddedRaw holds the raw embedded filesystem with the "templates/" prefix.
// The all: prefix ensures dot-prefixed directories (.claude/, .moai/) and
// dot-prefixed files (.gitignore) are included.
//
//go:embed all:templates
var embeddedRaw embed.FS

// EmbeddedTemplates returns the embedded template filesystem with the
// "templates/" prefix stripped so that paths match deployment targets.
//
// For example, the embedded path "templates/.claude/agents/moai/expert-backend.md"
// becomes ".claude/agents/moai/expert-backend.md" in the returned fs.FS.
//
// In production this fs.FS is passed to NewDeployer() to create a Deployer
// that writes templates to the project root during "moai init".
func EmbeddedTemplates() (fs.FS, error) {
	return fs.Sub(embeddedRaw, "templates")
}
