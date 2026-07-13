---
name: hns-moaiadk-patterns
description: >
  moai-adk-go domain-patterns reference for the 4 harness specialists
  (cli-template-specialist, quality-specialist, workflow-specialist,
  hook-ci-specialist). Covers the CLI/template/config/hook/spec subsystem
  architecture, key source paths, the Pipeline specialist delegation map, the
  Template-First build cycle, the namespace separation contract, and common
  add-a-template / add-a-hook / add-an-agent / add-a-SPEC workflows. Loaded by
  the specialists when working on moai-adk-go's own Go codebase and templates.
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "harness/domain-patterns"
  status: "active"
  updated: "2026-06-17"
  tags: "moai-adk-go,cli,template,harness,patterns"
progressive_disclosure:
  level_1_tokens: 120
  level_2_tokens: 4500
  level_3_optional: true
triggers:
  agents:
    - cli-template-specialist
    - quality-specialist
    - workflow-specialist
    - hook-ci-specialist
  keywords: moai-adk-go, internal/cli, internal/template, embed.go, make build, go:embed, template-first, harness namespace
paths: "internal/**/*.go,internal/template/templates/**,.claude/**,.moai/**"
---

# moai-adk-go Domain Patterns

## Architecture Quick Reference

moai-adk-go is a Go binary (`moai`) with four subsystems:

1. **CLI** (`internal/cli/*.go`, `cmd/moai/`) — Cobra commands: `init`,
   `update`, `hook`, `build`, `glm`, `cc`, `cg`, `version`, `doctor`,
   `spec`. Subcommand handlers read stdin JSON for hooks, emit structured
   output for the orchestrator.
2. **Template system** (`internal/template/`) — `go:embed`-based scaffolding.
   Source at `internal/template/templates/`, embedded into the binary via
   `//go:embed all:templates` in `internal/template/embed.go` (no generated
   `.go` file). `make build` recompiles the binary.
   `TemplateContext` (`{{.GoBinPath}}` / `{{.HomeDir}}`) renders at `moai init`.
3. **Config** (`internal/config/`) — `defaults.go` (single source for
   thresholds), `envkeys.go` (env-var constants), `TemplateContext` renderer.
4. **Hook + CI** (`.claude/hooks/moai/*.sh`, `.github/workflows/`) — bash
   wrapper hooks calling `moai hook <event>`; CI guard enforces template
   neutrality.

Plus the **SPEC lifecycle** (`.moai/specs/`) governing the project's own
development (plan→run→sync→Mx).

## Key Source Paths

| Subsystem | Path | Notes |
|-----------|------|-------|
| Cobra commands | `internal/cli/*.go` | wired from `cmd/moai/` |
| Template source | `internal/template/templates/**` | edit HERE first |
| Embedded assets | `internal/template/embed.go` | `//go:embed all:templates` (no generated file) |
| Config defaults | `internal/config/defaults.go` | threshold SSOT |
| Env constants | `internal/config/envkeys.go` | no hardcoded env names |
| SPEC docs | `.moai/specs/SPEC-*/` | spec/plan/acceptance/progress |
| Era classifier | `internal/spec/era.go` | `ClassifyEra()` H-1..H-6 |
| Hook scripts | `.claude/hooks/moai/*.sh` | bash only, no Python |
| CI workflows | `.github/workflows/*.yaml` | neutrality guard active |
| Harness agents | `.claude/agents/harness/*.md` | USER-OWNED (this skill) |

## Pipeline Specialist Delegation Map

This harness is a 4-stage pipeline; each specialist delegates to retained
agents (never archived, never replaces them):

```
CLI/Template ──→ quality ──→ workflow ──→ hook/CI
  │                │            │            │
  ├─ manager-develop (tdd, backend)
  ├─ Explore (read-only)
  ├─ sync-auditor (4-dim scoring)
  ├─ sync-phase-quality-gate.sh (Stop hook)
  ├─ manager-spec (plan)
  ├─ manager-develop (run)
  ├─ manager-docs (sync)
  ├─ plan-auditor (audit)
  ├─ builder-harness (artifact_type=hook|command|plugin)
  └─ Agent(general-purpose, model: opus, tools: ..., prompt: "...CI specialist...")
```

## Template-First Build Cycle

When adding/editing anything that ships to user projects:

1. Edit `internal/template/templates/<path>` FIRST.
2. Run `make build` → recompiles the binary (templates embedded via
   `//go:embed all:templates` in `embed.go`; no generated `.go` file).
3. Sync to local: `moai update` (or manual copy).
4. Verify the local `.claude/` / `.moai/` reflects the template.
5. Run `go test ./internal/template/...` (neutrality audit included).

Never edit `.claude/` or `.moai/` directly without a template source. The
source of truth is `templates/` — edit files there, then `make build`.

## Namespace Separation Contract

Two namespaces, enforced by `moai update`:

| Namespace | Location | Owner | `moai update` behavior |
|-----------|----------|-------|------------------------|
| Template-managed | `internal/template/templates/**` → `.claude/agents/{core,expert,meta}/`, `moai-*` skills | MoAI-ADK distribution | Overwrites local on sync |
| User-owned (this harness) | `.claude/agents/harness/`, `harness-*` skills, `.moai/harness/` | Project developer | NEVER deleted/modified; backup before update |

The canonical user-owned skill prefix is `harness-*` (recognized by Go
enforcement after the namespace catch-up, SPEC-V3R6-HARNESS-NAMESPACE-V2-001).
The legacy `my-harness-*` form is retained during a backward-compat
deprecation window; new skills MUST use the bare `harness-*` prefix.

## Common Workflows

### Add a template

1. Create file at `internal/template/templates/<path>`.
2. `make build`.
3. `moai update` (or test via `./moai init /tmp/test-project`).
4. `go test ./internal/template/... -run TestTemplateNeutralityAudit`.

### Add a hook

1. Write `.claude/hooks/moai/handle-<event>.sh` (bash, reads stdin JSON, calls
   `moai hook <event>`).
2. Wire in `.claude/settings.json` with `"$CLAUDE_PROJECT_DIR/..."` quoting +
   `timeout: 5`.
3. If the hook is template-distributable, add the wrapper template source AND
   the settings.json entry to `internal/template/templates/`.

### Add an agent (harness specialist)

1. Create `.claude/agents/harness/<role>-specialist.md` with `name`,
   trigger-shaped `description`, `skills:` array (companion skill), `tools:`
   (CSV string).
2. Ensure the companion `harness-*` skill exists (else self-activation
   smoke gate FAILs).

### Add a SPEC

1. `/moai plan "<description>"` → `manager-spec` authors plan-phase artifacts.
2. `plan-auditor` independent audit gate.
3. **Implementation Kickoff Approval** human gate (orchestrator runs
   `AskUserQuestion`).
4. `/moai run SPEC-<ID>` → `manager-develop` (cycle_type per quality.yaml).
5. `/moai sync SPEC-<ID>` → `manager-docs`.
6. `sync-auditor` 4-dimension gate.
7. 3-phase close on the single sync commit (populate `sync_commit_sha` in §E.4; the sync commit carries the `implemented → completed` transition — per SPEC-V3R6-LIFECYCLE-REDESIGN-001, the former separate `mx_commit_sha` / §E.5 Mx-phase step is retired; MX Tag validation is a sync sub-step).

## Cross-References

- CLAUDE.local.md §2 (Template-First Rule), §7 (hooks), §21 (dev-only commands)
- `.claude/rules/moai/development/agent-authoring.md` — agent frontmatter schema
- `.claude/rules/moai/development/skill-authoring.md` — skill frontmatter schema
- `.claude/rules/moai/workflow/archived-agent-rejection.md` §C — migration table
- `.claude/skills/moai-meta-harness/SKILL.md` § Namespace Separation
