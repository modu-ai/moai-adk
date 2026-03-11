# ADR-011: Template Passthrough Tokens

## Status
Accepted (2022-06-15, updated 2026-03-11)

## Context
The template renderer uses Go's `text/template` package with strict mode (`missingkey=error`). After template execution, we validate that no unexpanded tokens remain. However, certain environment variables are resolved at runtime by Claude Code or the shell and must not be flagged as errors.

## Decision
We maintain a whitelist of passthrough tokens in `internal/template/renderer.go`:

```go
var claudeCodePassthroughTokens = []string{
    "$CLAUDE_PROJECT_DIR", // Set by Claude Code, points to project root
    "$CLAUDE_SKILL_DIR",   // Set by Claude Code, points to skill directory
    "$ARGUMENTS",          // Set by Claude Code, command arguments
    "$HOME",               // Set by the shell, user home directory
}
```

### Token Resolution Responsibilities

| Token | Resolved By | When | Scope |
|-------|-------------|------|-------|
| `$CLAUDE_PROJECT_DIR` | Claude Code | Runtime | Project |
| `$CLAUDE_SKILL_DIR` | Claude Code | Runtime | Skill-specific |
| `$ARGUMENTS` | Claude Code | Runtime | Command |
| `$HOME` | Shell | Environment | System |
| `{{.Variable}}` | Go template | Build time | Template |

### File Type Distinction

| File Type | Extension | Renderer Handles | ${VAR} Support |
|-----------|-----------|------------------|----------------|
| Template | `.tmpl` | ✅ Yes | ❌ Only via data |
| Markdown | `.md` | ❌ No | ✅ Runtime vars OK |
| JSON | `.json` | ❌ No | ✅ Runtime vars OK |

**Critical:** `.md` files are NOT processed by the template renderer. They are read directly by Claude Code at runtime, so `${CLAUDE_SKILL_DIR}` and similar variables are resolved by Claude Code, not the renderer.

## Consequences

### When Adding New Runtime Variables

1. **Add to passthrough list:** Update `claudeCodePassthroughTokens` in `internal/template/renderer.go`
2. **Add test:** Update `TestRendererPassthroughTokens` in `internal/template/renderer_test.go`
3. **Add completeness check:** Update `TestPassthroughTokensCompleteness` required tokens map
4. **Document:** Update this ADR and `.coderabbit.yaml` path instructions

### Negative Consequences

- Every new runtime variable requires explicit whitelisting
- Forgetting to whitelist causes CI failures (unexpanded token detection)
- CodeRabbit may flag legitimate usage as errors if not documented in `.coderabbit.yaml`

### Positive Consequences

- Explicit list prevents accidental unexpanded tokens in production
- Test coverage ensures consistency
- Clear separation between build-time (Go template) and runtime (Claude Code) variables

## References

- Implementation: `internal/template/renderer.go:39-43`
- Tests: `internal/template/renderer_test.go:154-172` (passthrough), `:414+` (completeness)
- CodeRabbit config: `.coderabbit.yaml:48-54` (template instructions)
