// Package agentemit implements the deterministic dual publication of the
// retained agent definitions (SPEC-CODEX-DUAL-AGENTS-001, Option A).
//
// The neutral layer is the pair (.md definitions + this package's embedded
// agents-codex.yaml manifest): the .md files ARE the neutral core, so their
// publication is identity — the emitter never re-renders, reformats, or
// re-orders them (the moai update zero-diff regression ban holds by
// construction). The Codex publication is a deterministic transform of
// (.md x manifest) emitting one TOML per agent under .codex/agents/.
//
// The emitter is fail-closed: an unknown tool token, an unmapped effort
// value, an invalid sandbox value, or an unrepresentable body fails emission
// with a diagnostic naming the offending file and token, and no partial
// artifact set is produced. This exists because codex-cli silently ignores
// unknown config values, so the generator side must validate its own output.
package agentemit

// AgentDoc is the parsed neutral form of one agent .md definition.
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
	// Markdown maps each source .md path to its published bytes — identity
	// pass-through (bytes equal to source; never re-rendered).
	Markdown map[string][]byte
	// CodexTOML maps each emitted TOML path (layout-resolved, e.g.
	// .codex/agents/moai/<name>.toml) to its deterministic rendering.
	CodexTOML map[string][]byte
}
