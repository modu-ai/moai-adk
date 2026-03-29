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

// @MX:ANCHOR: [AUTO] go:embed 템플릿 파일시스템 접근점 - init/update/deployer 등 6개 이상의 호출자가 의존
// @MX:REASON: fan_in=6, 바이너리에 임베드된 유일한 템플릿 소스이며 "templates/" prefix 스트립 규칙이 배포 경로와 1:1 대응됨
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
