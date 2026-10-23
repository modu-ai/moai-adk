// Package agentemit renders the agent definitions from one edit surface, the
// agent source (SPEC-CODEX-DUAL-AGENTS-001, SPEC-AGENT-CORE4-001).
//
// The agent source is a directory of .md files, one per definition: the
// shipped definition file plus optional claude-only and codex-only block
// markers (render.go). Both publications are rendered from it. The Claude
// publication is the source's Claude view; a source without markers renders to
// itself, so the shipped Markdown bytes are carried unchanged. The Codex
// publication is a deterministic transform of (Codex view x this package's
// embedded agents-codex.yaml manifest) emitting one TOML per agent. The copies
// on disk are generated outputs — `make agents-emit` writes them and
// `make agents-emit-check` compares them without writing (sets.go).
//
// The emitter is fail-closed: an unknown tool token, an unmapped effort
// value, an invalid sandbox value, or an unrepresentable body fails emission
// with a diagnostic naming the offending file and token, and no partial
// artifact set is produced. This exists because codex-cli silently ignores
// unknown config values, so the generator side must validate its own output.
package agentemit

// AgentDoc is the parsed form of one agent .md definition (its Codex view).
type AgentDoc struct {
	File        string   // source .md file name (base)
	Name        string   // frontmatter name
	Description string   // decoded block-scalar description
	Tools       []string // decoded tools CSV tokens
	Model       string   // frontmatter model (documented-drop carrier, never emitted in v1)
	Effort      string   // frontmatter effort
	Skills      []string // optional skills preload (M1 seam; not emitted in v1)
	Body        []byte   // verbatim body bytes after the closing frontmatter delimiter
}

// Publication is a deterministic dual publication of one agent set.
type Publication struct {
	// Markdown maps each source .md path to its published Claude definition:
	// the Claude view of the source (equal to the source bytes when the source
	// carries no block markers).
	Markdown map[string][]byte
	// CodexTOML maps each emitted TOML path (layout-resolved, e.g.
	// .codex/agents/moai/<name>.toml) to its deterministic rendering.
	CodexTOML map[string][]byte
}
