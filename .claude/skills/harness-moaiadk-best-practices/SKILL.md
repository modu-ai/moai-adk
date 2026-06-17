---
name: harness-moaiadk-best-practices
description: >
  moai-adk-go best-practices reference for the 4 harness specialists
  (cli-template-specialist, quality-specialist, workflow-specialist,
  hook-ci-specialist). Covers TRUST 5 gates, Go test isolation
  (t.TempDir, no OTEL env in parallel tests), hardcoding-prevention rules
  (env constants in envkeys.go, thresholds in defaults.go), the
  AskUserQuestion orchestrator-only boundary, the deferred-tool preload rule,
  the archived-agent rejection contract, and verification-claim integrity.
  Loaded by the specialists when authoring or reviewing moai-adk-go code.
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "harness/best-practices"
  status: "active"
  updated: "2026-06-17"
  tags: "moai-adk-go,best-practices,trust5,testing,hardcoding"
progressive_disclosure:
  level_1_tokens: 120
  level_2_tokens: 4000
  level_3_optional: true
triggers:
  agents:
    - cli-template-specialist
    - quality-specialist
    - workflow-specialist
    - hook-ci-specialist
  keywords: TRUST 5, t.TempDir, envkeys.go, defaults.go, AskUserQuestion, archived-agent, verification-claim, deferred tool
paths: "internal/**/*.go,**/*_test.go,.claude/rules/**"
---

# moai-adk-go Best Practices

## TRUST 5 Quality Gates

Every change must pass all five dimensions before completion:

| Pillar | Gate | Failure action |
|--------|------|----------------|
| **Tested** | `go test ./...` with coverage | Block merge; generate missing tests |
| **Readable** | `golangci-lint run` | Warn; suggest refactoring |
| **Unified** | `go fmt` + `goimports` | Auto-format or warn |
| **Secured** | OWASP-aligned review (per-spawn opus agent) | Block; require review |
| **Trackable** | Conventional Commits regex | Suggest format |

Coverage targets: 85% package minimum; 90%+ for critical packages
(`internal/cli`, `internal/template`, `internal/hook`).

## Test Isolation

- **Always** `t.TempDir()` for temp dirs — auto-cleanup, under `os.TempDir()`.
- **macOS path pitfall**: `t.TempDir()` returns `/var/folders/...`. Go's
  `filepath.Join(cwd, absPath)` does NOT strip the leading `/`:
  `filepath.Join("/a/b", "/var/folders/x")` → `"/a/b/var/folders/x"` (WRONG).
  Use `filepath.Abs()` when resolving user-supplied paths in CLI commands.
- **No OTEL env in parallel tests** (CLAUDE.local.md §WARN): never
  `t.Setenv("OTEL_EXPORTER_*", ...)` in parallel tests — the OTEL SDK
  initializes global state from env vars on first use, causing data races. Use
  a fake/no-op exporter instead; or make the parent test non-parallel.
- **No `t.Setenv("HOME", tmpDir)`** in GLM integration tests — parallel-test
  pollution. Use `t.TempDir()` + explicit path construction.
- **After fixing any test**, run the FULL suite (`go test ./...`) to catch
  cascading failures. Use `-count=1` to disable caching when debugging flaky
  tests; use `-race` for concurrency-safety checks.

## Hardcoding Prevention

- **URLs / model names / org names / API headers** → extract to `const`.
- **Environment variable names** → define in `internal/config/envkeys.go` as
  constants; reference the constant everywhere. Never inline a raw env string.
- **Thresholds** → single source in `internal/config/defaults.go`. Never
  duplicate a threshold across packages.
- **Cross-platform paths** → prefer `$HOME`, `HOMEBREW_PREFIX`, etc. In
  `.sh.tmpl` fallback paths use `$HOME` (not `.HomeDir`), because `.HomeDir`
  freezes at `moai init` time and breaks for users with non-standard layouts.
- **Hardcoding allowed** only in `CLAUDE.local.md`, `settings.local.json`,
  and `_test.go` files inside `t.TempDir()`.

## AskUserQuestion Boundary (orchestrator-only)

- `AskUserQuestion` is the ONLY user-facing question channel, and it is
  reserved for the MoAI orchestrator (main session).
- **Subagents (including these harness specialists) MUST NOT invoke
  AskUserQuestion.** If user input is required, return a structured blocker
  report to the orchestrator (see `.claude/rules/moai/core/askuser-protocol.md`
  § Blocker Report Format).
- **Deferred-tool preload**: `AskUserQuestion`, `TaskCreate`, `TaskUpdate`,
  `TaskList`, `TaskGet` are deferred tools — schema not loaded at session
  start. The orchestrator MUST call
  `ToolSearch(query: "select:AskUserQuestion,TaskCreate,...")` before first
  use. Subagents inherit this constraint.
- Free-form prose questions in response text are prohibited — always route
  through AskUserQuestion (orchestrator) or a blocker report (subagent).

## Archived-Agent Rejection Contract

12 agents are ARCHIVED and MUST NOT be referenced anywhere in generated
harness files — no `delegates-to`, no prose, no examples. The full list of
archived names lives in the canonical SSOT at
`.claude/rules/moai/workflow/archived-agent-rejection.md` §B; this skill does
not repeat the literal names (repeating them in every generated file would
re-seed the exact tokens the rejection contract is meant to suppress).

The 8 RETAINED agents are the only valid delegation targets:

```
manager-spec, manager-develop, manager-docs, manager-git,
plan-auditor, sync-auditor, builder-harness, Explore (Anthropic built-in)
```

For domain expertise formerly provided by the archived domain-expert agents,
use the per-spawn pattern: `Agent(subagent_type: "general-purpose", model:
"opus", tools: "<whitelist>", prompt: "...<domain> specialist:
<conventions>...")` at delegation time. See
`.claude/rules/moai/workflow/archived-agent-rejection.md` §C for the full
migration table (rows #1-#12), which maps each archived agent to its
canonical retained-agent or per-spawn replacement.

## Verification-Claim Integrity

Per `.claude/rules/moai/core/verification-claim-integrity.md`:

- **No unobserved claims.** A "tests pass" / "coverage 87%" / "lint clean"
  assertion is valid ONLY when the actor ran the command and observed the
  output. An unran command is a gap, never a pass.
- **No unobserved defect claims.** Inferring a defect/debt/drift from
  frontmatter text or grep matches alone — without the domain's dedicated tool
  (`moai spec audit`, `go test -cover`, `golangci-lint`) — is a hypothesis,
  not a verified defect. The 2026-06-17 incident (29 SPECs wrongly flagged as
  "Mx-close debt"; `moai spec audit` showed all 29 were grandfather-protected)
  is the canonical worked example.
- **Baseline attribution.** Every verification claim names the command run +
  the verbatim output observed, measured against this tree in this run. A
  number from a different SPEC/package/time is a carry-over, not a baseline.
- **5-section report format**: Claim / Evidence / Baseline-attribution / Gaps
  / Residual-risk. The Gaps section is the defense — force yourself to
  enumerate what was NOT observed.

## Cross-References

- CLAUDE.local.md §6 (testing), §14 (hardcoding), §19 (AskUserQuestion)
- `.claude/rules/moai/core/verification-claim-integrity.md` v1.1.0
- `.claude/rules/moai/core/askuser-protocol.md`
- `.claude/rules/moai/workflow/archived-agent-rejection.md`
- `.claude/rules/moai/development/coding-standards.md`
