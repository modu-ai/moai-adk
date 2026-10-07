// Package template provides template deployment and rendering for MoAI projects.
//
// The templates/ subdirectory contains curated template content that is embedded
// into the moai binary at compile time. The embedded set is bounded by a
// GENERATED per-file embed allowlist (embed_manifest_gen.go, emitted by
// internal/template/embedemit from the git-tracked file set) — not by
// //go:embed all:templates, which embedded every file on disk including
// git-ignored and untracked additions (card t1539). This includes agent
// definitions, skill files, rules, output styles, configuration references,
// and root files (CLAUDE.md, .gitignore).
//
// Runtime-generated files (settings.json, .lsp.json) are
// intentionally excluded from the embedded templates per ADR-011
// (Zero Runtime Template Expansion) and AD-001 (Go compiled hooks).
// These files are generated programmatically via Go struct serialization
// in settings.go (SettingsGenerator).
package template

import (
	"io/fs"
)

// @MX:ANCHOR: [AUTO] go:embed template filesystem access point - depended on by 6 or more callers including init/update/deployer
// @MX:REASON: [AUTO] fan_in=6, sole embedded template source in the binary; the "templates/" prefix strip rule maps 1:1 to deployment paths
// EmbeddedTemplates returns the embedded template filesystem with the
// "templates/" prefix stripped so that paths match deployment targets.
//
// For example, the embedded path "templates/.claude/agents/expert/expert-backend.md"
// becomes ".claude/agents/expert/expert-backend.md" in the returned fs.FS.
//
// In production this fs.FS is passed to NewDeployer() to create a Deployer
// that writes templates to the project root during "moai init".
func EmbeddedTemplates() (fs.FS, error) {
	return fs.Sub(embeddedRaw, "templates")
}
