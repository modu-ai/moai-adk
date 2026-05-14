# moai-adk-go Domain Harness

Project-specific harness for the moai-adk-go project (Go CLI tool + Claude Code orchestration framework).

## Project Domain

moai-adk-go is a Go binary (`moai`) providing:
- CLI commands for project init, template deployment, hook execution, quality gates
- `go:embed` template system (`internal/template/templates/`)
- Claude Code integration via skills, agents, rules, and hooks
- SPEC-based workflow (plan -> run -> sync) with DDD/TDD methodologies
- Learning subsystem (observer + 4-tier evolution + 5-layer safety)

Implementation language: Go
Template language: YAML/JSON/Shell/Markdown
Supported project languages (templates): 16 (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift)

## Team Architecture: Pipeline Pattern

The harness follows a sequential Pipeline pattern with domain specialists at each stage:

```
CLI/Template Specialist --> Quality Specialist --> Workflow Specialist --> Hook/CI Specialist
         (Stage 1)               (Stage 2)             (Stage 3)             (Stage 4)
```

Each specialist focuses on its domain but references existing MoAI agents for execution.

## Agent Roles

| Role | Agent Definition | Delegates To | Domain |
|------|-----------------|--------------|--------|
| CLI/Template Specialist | `my-harness/cli-template-specialist` | expert-backend | Cobra commands, go:embed, template rendering |
| Quality Specialist | `my-harness/quality-specialist` | manager-quality | Testing, linting, coverage, LSP validation |
| Workflow Specialist | `my-harness/workflow-specialist` | manager-spec, manager-develop | SPEC workflow, EARS requirements, MX tags |
| Hook/CI Specialist | `my-harness/hook-ci-specialist` | expert-devops | Shell hooks, CI pipelines, GitHub Actions |

## Key Source Paths

| Domain | Paths |
|--------|-------|
| CLI commands | `internal/cli/*.go` |
| Template system | `internal/template/`, `internal/template/templates/` |
| Configuration | `internal/config/`, `.moai/config/` |
| Hooks | `internal/hook/`, `.claude/hooks/moai/` |
| Tests | `**/*_test.go` |
| CI workflows | `.github/workflows/` |
| SPEC documents | `.moai/specs/` |
| Skills | `.claude/skills/` |
| Agent definitions | `.claude/agents/` |
| Rules | `.claude/rules/moai/` |

## Quality Targets

| Metric | Target |
|--------|--------|
| Package coverage | 85% minimum |
| Critical package coverage (cli, template, hook) | 90%+ |
| Lint | Zero errors (golangci-lint) |
| Vet | Zero errors (go vet) |
| Race | Zero races (go test -race) |
| SPEC completion | All AC criteria met |
| MX tags | Applied per MX protocol |
| Hook handlers | All 27 events covered |

## Skills Reference

| Skill | Domain |
|-------|--------|
| my-harness-cli-template | CLI commands, templates, config |
| my-harness-quality | Testing, linting, coverage |
| my-harness-workflow | SPEC workflow, EARS, MX tags |
| my-harness-hook-ci | Hooks, CI pipelines, GitHub Actions |

## Constraints

- ALL harness artifacts use `my-harness-*` prefix (never `moai-*`)
- Harness specialists delegate to existing MoAI agents; they do not replace them
- Harness does not modify template source directly; changes go through `internal/template/templates/` -> `make build`
- Tests must use `t.TempDir()` for isolation (never write to project root)
- Hooks are shell scripts only (no Python)
- `settings.local.json` is runtime-managed, never in templates

---

Version: 1.0.0
Created: 2026-05-14
Classification: Project-specific harness
