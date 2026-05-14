---
name: my-harness-cli-template
description: >
  CLI/Template domain knowledge for moai-adk-go covering cobra commands, go:embed template
  system, YAML config, and template rendering pipeline.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-05-14"
  modularized: "false"
  tags: "cli, template, cobra, go:embed, config, yaml, moai-cli"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["cobra", "command", "template", "embed", "config", "yaml", "moai init", "moai update", "moai build", "rendering", "internal/cli", "internal/template"]
  agents:
    - "my-harness-cli-template-specialist"
    - "expert-backend"
  phases:
    - "run"
  languages:
    - "go"
---

# CLI/Template Domain Knowledge

Domain-specific knowledge for moai-adk-go's CLI and template system. Supplements `expert-backend` with project-specific patterns.

## Quick Reference

### Architecture Overview

```
moai binary (Go)
  ├── internal/cli/          ~50 cobra command files
  ├── internal/template/
  │   ├── templates/         Source of truth for all templates
  │   ├── embedded.go        Auto-generated (go:embed)
  │   ├── context.go         TemplateContext with GoBinPath, HomeDir
  │   └── renderer.go        Template rendering engine
  └── internal/config/       Configuration loading and defaults
```

### Key Patterns

1. **Template-First Rule**: New files under `.claude/` or `.moai/` must be added to `internal/template/templates/` first, then `make build`
2. **Embedded System**: `//go:embed templates/*` in `embedded.go` -- auto-generated, never edit
3. **Template Variables**: `{{.GoBinPath}}` (init-time absolute), `{{.HomeDir}}` (init-time absolute)
4. **Fallback Paths**: Use `$HOME` (not `.HomeDir`) in `.sh.tmpl` files for runtime flexibility
5. **16-Language Neutrality**: Templates treat all 16 supported languages equally -- no "PRIMARY" language

### Build Cycle

```bash
# 1. Edit templates
vim internal/template/templates/.claude/skills/...

# 2. Regenerate embedded files
make build

# 3. Run tests
go test ./internal/template/...

# 4. Verify
ls -la internal/template/embedded.go
```

### Command Structure

Each cobra command lives in `internal/cli/<command>.go` with corresponding tests in `<command>_test.go`. Commands follow the pattern:
- `rootCmd` in `root.go` with subcommands registered via `init()`
- Each command file: `var <cmd>Cmd = &cobra.Command{...}` + `func init() { rootCmd.AddCommand(...) }`

### Configuration System

- Main config: `.moai/config/config.yaml`
- Sections: `.moai/config/sections/*.yaml` (quality, language, user, workflow, harness, design)
- Priority: env vars > user config > template defaults
- Env var keys: `internal/config/envkeys.go` (constants, never hardcode)

## Implementation Guide

### Adding a New Cobra Command

1. Create `internal/cli/<command>.go` with cobra command struct
2. Register in `init()` function
3. Create `internal/cli/<command>_test.go` using `t.TempDir()`
4. If command modifies templates, add template file to `internal/template/templates/`
5. Run `make build` if templates changed
6. Run `go test ./internal/cli/...` to verify

### Adding a New Template File

1. Add file to `internal/template/templates/<path>`
2. If file needs variable substitution, use `.tmpl` extension
3. Register passthrough tokens in `renderer.go` if needed (e.g., `$HOME`)
4. Run `make build` to regenerate `embedded.go`
5. Run `go test ./internal/template/...`

### Template Rendering Pipeline

```
TemplateContext{GoBinPath, HomeDir}
  -> renderer.Render(templateContent, context)
  -> Output file (variable substitution applied)
```

Reserved passthrough tokens (not substituted, passed through verbatim):
- `$HOME`, `$PATH`, `$CLAUDE_PROJECT_DIR`
- Environment variable references in shell scripts

### Error Wrapping Convention

```go
// Correct: use fmt.Errorf with %w
if err != nil {
    return fmt.Errorf("deploy template: %w", err)
}

// Wrong: string concatenation
if err != nil {
    return fmt.Errorf("deploy template: " + err.Error())
}
```

## Cross-References

- CLAUDE.local.md Section 2: Template-First Rule and file synchronization
- CLAUDE.local.md Section 8: Template Variable Strategy
- `.claude/rules/moai/development/coding-standards.md`: Coding conventions
- `moai-foundation-cc` skill: Claude Code authoring patterns
- `moai-foundation-core` skill: SPEC system and TRUST 5
