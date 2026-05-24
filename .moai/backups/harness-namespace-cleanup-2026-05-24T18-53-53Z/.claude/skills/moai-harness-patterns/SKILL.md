---
name: moai-harness-patterns
description: >
  MoAI-ADK harness pattern library — unified domain knowledge covering hook/CI dispatch
  (PostToolUse, SessionStart, GitHub Actions, release automation), workflow patterns
  (SPEC structure, EARS, MX tags, plan-run-sync pipeline), and Go quality gates (testing,
  linting, coverage, race detection, LSP). Use for moai-adk-go harness work — NOT for
  general MoAI agent patterns (see moai-foundation-cc).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "0.1.0"
  category: "domain"
  status: "active"
  updated: "2026-05-22"
  tags: "harness, hooks, ci, github-actions, spec, plan-run-sync, ears, mx-tag, go-test, golangci-lint, coverage, lsp, quality-gate"

progressive_disclosure:
  enabled: true
  level1_tokens: 120
  level2_tokens: 5000

triggers:
  keywords: ["hook", "PostToolUse", "SessionStart", "CI", "GitHub Actions", "GoReleaser", "handle-", "settings.json", "SPEC", "plan", "run", "sync", "EARS", "MX tag", "AC-", "REQ-", "go test", "lint", "vet", "coverage", "golangci-lint", "LSP", "race", "t.TempDir", "flaky"]
  agents: ["moai-harness-hook-ci-specialist", "moai-harness-workflow-specialist", "moai-harness-quality-specialist", "manager-spec", "manager-develop", "manager-docs", "manager-quality", "expert-devops"]
  phases: ["plan", "run", "sync"]
  languages: ["go"]
---

# MoAI-ADK Harness Patterns (`moai-harness-patterns`)

Unified harness domain knowledge for `moai-adk-go`. This skill consolidates three
specialty areas previously split across separate skills:

1. **Hook & CI** — Shell hook wrappers, GitHub Actions, release automation
2. **Workflow** — SPEC structure, EARS requirements, plan-run-sync pipeline, MX tags
3. **Quality** — Go testing, linting, coverage, race detection, LSP gates

Each harness specialist agent (`moai-harness-hook-ci-specialist`,
`moai-harness-workflow-specialist`, `moai-harness-quality-specialist`) loads this single
skill for shared context. Supplements general agents (`expert-devops`, `manager-spec`,
`manager-quality`) with project-specific patterns.

## Quick Reference

### Hook System Architecture

```
Claude Code hook event
  -> .claude/hooks/moai/handle-{event}.sh  (shell wrapper)
  -> moai hook {event}                     (Go binary handler)
  -> internal/hook/{handler}.go            (Go handler logic)
```

**27 hook events** spanning Session, Tool, Agent, State, Permissions, Interaction, Team,
and Worktree categories. Verify: `ls .claude/hooks/moai/handle-*.sh | wc -l` == 27. Full
event list in `.claude/rules/moai/core/agent-hooks.md`.

### SPEC Workflow Pipeline

```
/moai plan "description"  → manager-spec       → SPEC document with EARS requirements
/moai run SPEC-XXX        → manager-develop    → Implementation (TDD or DDD)
/moai sync SPEC-XXX       → manager-docs       → Documentation + PR creation
```

### Quality Targets

`go vet ./...` and `golangci-lint run` must be zero errors at all phases.
`go test ./...` all-pass at all phases; `-race` must be zero races at run.
Coverage: per-package >= 85% at sync (`go test -cover ./...`); critical packages
(`internal/cli/`, `internal/template/`) >= 90%. LSP gates: run = zero errors / zero type
errors / zero lint errors; sync adds warnings cap (max 10).

## Implementation Guide

### Section 1 — Hook & CI Patterns

**Hook wrapper pattern** (thin shell wrapper, Go does the work): read stdin JSON via
`INPUT=$(cat)`, pipe to `moai hook {event} <<< "$INPUT"`. Lives at
`.claude/hooks/moai/handle-{event}.sh`.

**settings.json invocation rules**: always quote `$CLAUDE_PROJECT_DIR` (e.g.
`"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-X.sh"`); set timeout (default 5s, max
10s for post-processing); use full path, do not rely on PATH.

**Agent-scoped hooks** declared in agent YAML frontmatter under `hooks:` map (per event:
`PreToolUse`, `PostToolUse`, `SubagentStop`, etc.) with `matcher` regex and `command`
shell call to `handle-agent-hook.sh {action}`. Default timeout 5s (10s for post-processing).

**GitHub Actions workflows** in `.github/workflows/`: `ci.yml` (push/PR to main — matrix
ubuntu/macos/windows × Go 1.24, lint + test + build); `release.yml` (tag `v*` —
GoReleaser 5 platforms); `release-drafter.yml` (PR merge to main — auto-label + draft
changelog); `auto-merge.yml` (Dependabot CI pass — squash); `codeql.yml` (push/PR — Go
security analysis); `spec-lint.yml` (PR — SPEC frontmatter validation);
`spec-status-auto-sync.yml` (schedule — status drift detection); `docs-i18n-check.yml`
(PR touching docs-site — 4-locale sync verification).

**Release process** — always use `./scripts/release.sh vX.Y.Z "description"` (or
`--hotfix`); never manual tag push. Chain: script → tag → `release.yml` → GoReleaser →
5-platform binaries → GitHub Release → Release Drafter updates changelog.

**Hook handler testing**: Go unit tests in `internal/hook/` read JSON from stdin and
exercise the handler logic directly.

### Section 2 — Workflow Patterns

**SPEC structure** under `.moai/specs/<SPEC-ID>/`: `spec.md` (EARS + AC), `plan.md` (M1..Mn
milestones), `acceptance.md` (binary AC + REQ↔AC traceability, Tier M+), `scenarios.md`
(Tier L), `risks.md` (Tier L), `progress.md` (auto-generated during run). Tier S = 2
artifacts (spec + plan, AC inline); Tier M = 3 (adds acceptance.md); Tier L = 5 (adds
design + research).

**SPEC naming**: `SPEC-V{major}R{minor}-{CATEGORY}-{NUMBER}`. Categories: WF, ORC, RT,
SPC, HOOK, CI, CON, MX, CLI, TUI, BRAIN, STATUSLINE, HYBRID, MIG, GLM.

**EARS requirement patterns**: Ubiquitous (`The system shall [action]`); Event-Driven
(`When [event], the system shall [action]`); Unwanted (`If [bad condition], the system
shall [action]`); State-Driven (`While [state], the system shall [action]`); Optional
(`Where [feature] enabled, the system shall [action]`).

**MX tag types**: `@MX:NOTE` (context/intent, new exported funcs); `@MX:WARN` (danger
zone, requires `@MX:REASON` — goroutines, complexity >= 15); `@MX:ANCHOR` (invariant
contract, high fan_in functions >= 3 callers); `@MX:TODO` (incomplete work, untested
public functions). Per-file caps: 3 ANCHOR / 5 WARN / 10 NOTE / 5 TODO.

**Plan-in-main doctrine**: SPEC plan PRs merge to main (not feature branches). Run phase
uses worktree isolation. Sync PRs merge to main with full review history.

**Milestones**: M1 Foundation → M2 Core → M3 Integration → M4 Edge cases → M5 Polish.
**Waves**: 30+ task SPECs split into wave-PRs; track in `progress.md`.

**AC format**: `AC-{SHORT}-{NN}: {verifiable condition}` with `Verification:` command
and `Priority: P0|P1|P2`.

**SPEC status lifecycle (8-value enum)**:
`draft → planned → in-progress → implemented → completed`, alternates: `superseded`,
`archived`, `rejected`.

### Section 3 — Quality Patterns

**Critical test isolation rules**: (1) always `t.TempDir()` — auto-cleanup under `/tmp`;
(2) `filepath.Join` trap: `Join("/a/b", "/var/folders/x")` = `/a/b/var/folders/x` (WRONG)
— use `filepath.Abs()`; (3) `OTEL_*` env vars MUST NOT use `t.Setenv` in parallel tests
(global state race); (4) after any fix, run FULL suite (`go test ./...`); (5) flaky debug
disables cache via `go test -count=1 ./...`.

**Test execution**: `go test ./...` (full), `-race` (race detection), `-cover` (coverage),
`-run TestX ./pkg/` (specific), `-count=1` (disable cache for flaky debug), `-v` (verbose).

**Table-driven test pattern**: standard Go convention — `tests := []struct{name, input,
want string; wantErr bool}{...}` + `for _, tt := range tests { t.Run(tt.name, ...) }`.
Cover happy path + error path + edge cases per row.

**LSP quality gate thresholds by phase**:

| Phase | LSP Errors | Type Errors | Lint Errors | Warnings |
|---|---|---|---|---|
| plan | Baseline captured | Baseline | Baseline | Baseline |
| run | Zero required | Zero required | Zero required | — |
| sync | Zero required | Zero required | Zero required | Max 10 |

**Pre-commit quality gate**: `go vet ./... && golangci-lint run && go test ./...`.

**Coverage report**: `go test -coverprofile=coverage.out ./...` then
`go tool cover -html=coverage.out`.

**Common quality issues**: `filepath.Join` with absolute path → use `filepath.Abs()`;
OTEL data race in parallel tests → fake/no-op exporter; test writes outside temp dir →
always `t.TempDir()`; flaky CI → `go test -race -count=1`.

## Cross-References

`.claude/rules/moai/core/agent-hooks.md` (hook lifecycle), `.claude/rules/moai/workflow/
spec-workflow.md` (canonical SPEC workflow), `.claude/skills/moai/references/mx-tag.md`
(MX tag protocol), `.moai/config/sections/quality.yaml` (quality config), `CLAUDE.local.md`
§6 (testing) + §7 (hook development), `CLAUDE.md` §6 (TRUST 5), `.github/workflows/ci.yml`
and `release.yml`.

## Works Well With

3 harness specialists (`moai-harness-hook-ci-specialist`, `moai-harness-workflow-specialist`,
`moai-harness-quality-specialist`) load this skill. Also: `manager-develop` (MX tags +
EARS + quality), `manager-quality` (TRUST 5), `manager-git` (CI auto-fix push),
`expert-devops` (release + Actions), `moai-workflow-ci-loop` (CI runtime sibling),
`moai-foundation-quality` (broader orchestration), `moai-foundation-core` (TRUST 5 + SPEC
overview).

## Verification

- [ ] 27 hook handler scripts present (`ls .claude/hooks/moai/handle-*.sh | wc -l` == 27)
- [ ] `go vet ./... && golangci-lint run && go test ./...` passes
- [ ] `go test -race ./...` passes (no goroutine leaks)
- [ ] Package coverage >= 85%; `internal/cli/` and `internal/template/` >= 90%
- [ ] SPEC frontmatter passes `spec-lint.yml`
- [ ] MX tag count per file ≤ 3 ANCHOR / 5 WARN / 10 NOTE / 5 TODO
- [ ] `golangci-lint run --timeout=2m` zero NEW issues vs baseline
- [ ] Release uses `scripts/release.sh`, never manual `git tag` push

<!-- absorbed from moai-harness-hook-ci + moai-harness-workflow + moai-harness-quality (SPEC-V3R6-SKILL-CONSOLIDATE-001, 2026-05-22) -->
