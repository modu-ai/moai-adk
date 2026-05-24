---
id: SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001
title: "Hook cwd leak audit + resolveProjectRoot consistency"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook"
lifecycle: spec-anchored
tags: "hook, cwd, audit, refactoring, tier-s, v3.0, sprint-2"
tier: S
issue_number: null
depends_on: [SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001]
related_specs: [SPEC-V3R6-HOOK-ASYNC-EXPAND-001, SPEC-V3R6-HOOK-CONTRACT-FIX-001]
---

# SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001: Hook cwd Leak Audit + resolveProjectRoot Consistency

## Section A — Pre-existing State Survey (codebase audit prior to scoping)

This section enumerates pre-existing infrastructure facts discovered during plan-phase scoping (2026-05-23) after the prior fix commit `a9b3e8cd8` addressed only 1 of 10 instances of the same pattern. Every fact below was verified against working-tree HEAD `eaff5f272` on `main` via `grep -rn "os.Getwd" internal/hook/`.

### A.1 — Pre-existing infrastructure facts (5)

1. **9 `os.Getwd()` cwd-leak patterns remain in `internal/hook/`** across 5 files in 2 packages (`hook` and `hook/quality`). Exact locations verified via `grep -rn "os\.Getwd" internal/hook/ | grep -v "_test.go"`:
   - `internal/hook/subagent_start.go:37` (in `NewSubagentStartHandlerWithConfig`, sets `projectDir` field)
   - `internal/hook/subagent_start.go:211` (in `Handle`, sets local `projectDir` variable for `loadAgentFrontmatter`)
   - `internal/hook/pre_tool.go:326` (in `NewPreToolHandler`, sets `projectDir` field)
   - `internal/hook/pre_tool.go:336` (in `NewPreToolHandlerWithScanner`, sets `projectDir` field)
   - `internal/hook/observability_master.go:82` (in `loadObservabilityMaster`, sets local `root` variable for YAML file read)
   - `internal/hook/quality/gate.go:276` (in `executeStep`, sets local `projectDir` for ast-grep gate)
   - `internal/hook/quality/gate.go:297` (in `detectToolchain`, sets local `dir` for marker-file detection)
   - `internal/hook/quality/gate.go:399` (in `executeStep` ext-filter path, sets local `dir` for staged-files lookup)
   - `internal/hook/quality/gate.go:480` (in `anyConfigFileExists`, sets local `dir` for config-file existence check)

2. **Reference fix pattern exists and is shipping** in `internal/hook/post_tool_metrics.go:98-113`:
   ```go
   // resolveProjectRoot returns the MoAI project root for task metrics logging.
   // It prefers CLAUDE_PROJECT_DIR (set by the Claude Code hook system when invoking
   // hook scripts) over input.CWD. Returns empty string when the resolved path does
   // not already contain a .moai/ directory, preventing accidental creation of .moai/
   // in subdirectories or unrelated directories.
   func resolveProjectRoot(input *HookInput) string {
       root := os.Getenv(config.EnvClaudeProjectDir)
       if root == "" {
           root = input.CWD
       }
       if root == "" {
           return ""
       }
       if _, err := os.Stat(filepath.Join(root, ".moai")); err != nil {
           return ""
       }
       return root
   }
   ```
   Resolution priority: `CLAUDE_PROJECT_DIR` env var → `input.CWD` → empty (skip). The `.moai/` existence guard is the critical anti-leak safety: it returns empty string when invoked from a subdirectory or non-MoAI directory, preventing accidental `.moai/` creation. Currently consumed by `post_tool_metrics.go` (lines 48, 158), `post_tool_duration.go:105`, and `subagent_stop.go:210`.

3. **Commit `a9b3e8cd8` "fix(hook): subagent_stop dispatchCapture cwd leak via resolveProjectRoot"** (2026-05-23, prior commit on `main`) applied the `resolveProjectRoot` pattern to `subagent_stop.go:210` only. The 9 other instances were not addressed in that commit and remain leaking. This was confirmed by reading `internal/hook/subagent_stop.go:205-213` which contains the comment "Path resolution: delegates to resolveProjectRoot to prefer CLAUDE_PROJECT_DIR over input.CWD and guard against writes outside a valid MoAI project root."

4. **Cwd-leak discovery context — post HOI merge**. The leak class was discovered when SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 was implemented and the team noticed `internal/hook/.moai/harness/observations.yaml` (418L) and `.moai/specs/.moai/harness/usage-log.jsonl` (3L) being created at cwd-relative paths during hook execution. The `subagent_stop` fix (`a9b3e8cd8`) addressed the most visible leak path. This SPEC addresses the full audit. Memory reference: `[[project-v3r6-sprint2-hoi-revise-ready]]` § "cwd-leak fix" sub-section.

5. **Two distinct refactor strategies are required** due to package and type boundaries:
   - **Strategy A — `package hook` callers** (`subagent_start.go`, `pre_tool.go`, `observability_master.go`): These have access to `HookInput` (or can be refactored to receive it) and can directly call `resolveProjectRoot(input)`. The `observability_master.go:82` case is a special branch — it has no `HookInput` parameter and reads from a global toggle path; here the refactor strategy is to extract a thin helper `resolveProjectRootFromEnvOrEmpty()` that mirrors the env→empty fallback **without** the `.moai/` guard (since loading a config file from a non-MoAI directory is benign), OR factor a parameterless variant.
   - **Strategy B — `package hook/quality` callers** (4 sites in `gate.go`): These use `g.config.ProjectDir` (the `GateConfig.ProjectDir` field) as the primary source and only fall back to `os.Getwd()` when the field is empty. The `hook/quality` package does NOT import the parent `hook` package and does NOT have `HookInput` available. The refactor strategy here is local: extract a private helper `resolveQualityProjectDir(cfg GateConfig) string` inside `package quality` that prefers `cfg.ProjectDir` → `CLAUDE_PROJECT_DIR` env var → `os.Getwd()` fallback with a warning log. The `.moai/` existence guard is OPTIONAL here because quality-gate operations (linting, vet, test) are intended to run inside a project root by contract.

### A.2 — Verification commands (reproducible)

```bash
# Verify the 9 leak sites:
grep -rn "os\.Getwd" internal/hook/ | grep -v "_test.go"
# Expected: 9 matches across 5 files (subagent_start:37,211 / pre_tool:326,336 /
#           observability_master:82 / quality/gate:276,297,399,480)

# Verify the reference fix:
grep -n "func resolveProjectRoot" internal/hook/post_tool_metrics.go
# Expected: line 98

# Verify the prior fix commit:
git log --oneline --all | grep "a9b3e8cd8"
# Expected: a9b3e8cd8 fix(hook): subagent_stop dispatchCapture cwd leak via resolveProjectRoot
```

---

## Background

The MoAI hook subsystem (`internal/hook/`) invokes Go handler functions in response to Claude Code lifecycle events (SessionStart, PreToolUse, PostToolUse, SubagentStart, SubagentStop, etc.). Hooks need to resolve the **MoAI project root** to read configuration, write logs, and update telemetry under `.moai/`. The correct resolution order is:

1. `input.CWD` (when present; this is the most authoritative — Claude Code sets it from the active session)
2. `CLAUDE_PROJECT_DIR` env var (set by Claude Code hook system)
3. `os.Getwd()` (last-resort fallback)

When `os.Getwd()` is used directly without the env-var preference and without a `.moai/` existence guard, two failure modes occur:

1. **Cwd leak (writes)** — When a hook is invoked from a subdirectory (e.g., `internal/hook/` during test runs), `os.Getwd()` returns the subdirectory path, and downstream `.moai/` writes create a stray `.moai/` tree inside the subdirectory. This is the historical bug class addressed by the `.moai/` guard in `resolveProjectRoot`.
2. **Cwd leak (reads)** — When a hook reads from `<cwd>/.moai/config/sections/*.yaml`, an unintended cwd produces stale or missing config and can mask production issues during local development.

Commit `a9b3e8cd8` fixed the most visible leak (subagent_stop dispatchCapture writing observations.yaml). This SPEC closes the audit by applying the same pattern (or its package-local variant) to the remaining 9 sites.

---

## Requirements (EARS Format)

### Ubiquitous Requirements

**REQ-HCWA-001** — The `internal/hook/` package shall route every project-root resolution through either `resolveProjectRoot(input *HookInput)` (parent package `hook`) or a package-local helper `resolveQualityProjectDir(cfg GateConfig) string` (sub-package `hook/quality`). Direct `os.Getwd()` calls outside these helpers shall not exist in any non-test file under `internal/hook/`.

**REQ-HCWA-002** — The helper `resolveProjectRoot(input *HookInput) string` in `internal/hook/post_tool_metrics.go` shall remain the canonical reference pattern. Its public surface, behavior, and `.moai/` existence guard shall not change as part of this SPEC.

### Event-Driven Requirements

**REQ-HCWA-003** — WHEN `NewSubagentStartHandlerWithConfig(cfg ConfigProvider)` is called and `CLAUDE_PROJECT_DIR` is unset, THEN the handler shall resolve `projectDir` via a path that gives `input.CWD` first preference and `os.Getwd()` only as last-resort fallback, with a warning log emitted on `os.Getwd()` fallback.

**REQ-HCWA-004** — WHEN `(*subagentStartHandler).Handle(ctx, input)` is called, THEN the handler shall prefer `input.CWD` over `CLAUDE_PROJECT_DIR` over `os.Getwd()` for the `projectDir` resolution that feeds `loadAgentFrontmatter`. This is the existing order; the audit shall verify it is preserved and route through a shared helper rather than inline code.

**REQ-HCWA-005** — WHEN `NewPreToolHandler(cfg, policy)` or `NewPreToolHandlerWithScanner(cfg, policy, scanner)` is called, THEN the handler shall resolve `projectDir` via a helper that prefers `CLAUDE_PROJECT_DIR` env var over `os.Getwd()` and emits a warning log when `os.Getwd()` fallback is taken.

**REQ-HCWA-006** — WHEN `loadObservabilityMaster()` is called and `CLAUDE_PROJECT_DIR` is unset, THEN the function shall fall back to `os.Getwd()` via a shared helper that emits a warning log. The `.moai/` existence guard is NOT applied here because the function reads from a path that may not include `.moai/config/sections/observability.yaml` (the file is the toggle target itself).

**REQ-HCWA-007** — WHEN any `package quality` function (`executeStep`, `detectToolchain`, `anyConfigFileExists`) needs a project directory and `g.config.ProjectDir` is empty, THEN the function shall delegate to a package-local helper `resolveQualityProjectDir(cfg GateConfig) string` that prefers `cfg.ProjectDir` → `CLAUDE_PROJECT_DIR` env var → `os.Getwd()` fallback with a warning log.

### State-Driven Requirements

**REQ-HCWA-008** — WHILE the resolution helper is in effect, the helper shall log a warning at `slog.Warn` level with key `"cwd_fallback": true` and `"caller": "<function-name>"` whenever `os.Getwd()` is used as the last-resort fallback. This makes cwd-fallback events observable in `.moai/logs/` and in production telemetry.

### Unwanted Behavior Requirements

**REQ-HCWA-009** — The system shall not introduce any new `os.Getwd()` call site in `internal/hook/` outside the two helper functions. Any future hook handler that needs a project directory shall route through `resolveProjectRoot(input)` or `resolveQualityProjectDir(cfg)`.

**REQ-HCWA-010** — The system shall not modify `internal/hook/subagent_stop.go` (the prior fix from commit `a9b3e8cd8`), `internal/hook/post_tool_metrics.go` (the reference pattern), or `internal/hook/cohabitation_guard_test.go` (the HOI safety net). These are explicit PRESERVE targets — see Exclusions below.

### Optional Requirements

**REQ-HCWA-011** — WHERE feasible without expanding scope, the audit shall also identify and document any `os.Getwd()` calls in `internal/hook/*_test.go` files (test scope) for awareness. Test-scope leaks are not in scope for fix, but the documentation aids future hardening.

---

## Acceptance Criteria

Acceptance criteria are defined in [acceptance.md](acceptance.md). Summary of the 7 binary ACs:

- **AC-HCWA-001** — `grep -rn "os\.Getwd" internal/hook/ | grep -v "_test.go"` returns 0 matches (all 5 files clean).
- **AC-HCWA-002** — `resolveProjectRoot` call site count increases by +3 (subagent_start.go × 2 + pre_tool.go × 2 + observability_master.go × 1 = +5 in package hook, OR via thin shared helper = +3 net new call sites depending on refactor choice — see plan.md M1/M2). `resolveQualityProjectDir` is introduced with +4 call sites in package quality.
- **AC-HCWA-003** — Existing `cohabitation_guard_test.go` (HOI safety net) test cases continue to pass without modification.
- **AC-HCWA-004** — `go test -race ./internal/hook/...` exits 0 with no new race detector warnings.
- **AC-HCWA-005** — `golangci-lint run --timeout=2m` reports 0 NEW issues vs the M0 baseline captured at plan-phase entry (`HEAD eaff5f272`).
- **AC-HCWA-006** — `internal/hook/subagent_stop.go` is byte-identical pre-/post-SPEC (PRESERVE).
- **AC-HCWA-007** — `internal/hook/post_tool_metrics.go` `resolveProjectRoot` function body and signature are byte-identical pre-/post-SPEC (PRESERVE).

---

## Exclusions (What NOT to Build)

### Out of Scope — Hook cwd audit boundary

The following are explicitly out of scope for this SPEC. Any work in these areas requires a separate SPEC.

- **Modifying `internal/hook/subagent_stop.go`** — The `dispatchCapture` fix from commit `a9b3e8cd8` is shipping and correct. This SPEC PRESERVES that file unchanged.
- **Modifying `internal/hook/post_tool_metrics.go`** — `resolveProjectRoot` is the reference pattern. Its signature, body, and behavior PRESERVE unchanged.
- **Modifying `internal/hook/cohabitation_guard_test.go`** — HOI cohabitation safety net (SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 deliverable). PRESERVE unchanged.
- **Fixing `os.Getwd()` calls in `*_test.go` files** — Test-scope leaks are documented (REQ-HCWA-011) but not fixed in this SPEC. Future SPEC may address.
- **Refactoring `resolveProjectRoot` itself** — Including changing its signature, weakening the `.moai/` guard, or moving it to a different package. Reference pattern is frozen.
- **Cross-package consolidation** — Moving `resolveProjectRoot` and `resolveQualityProjectDir` into a shared `internal/hook/internal/pathutil/` sub-package. Considered, but defers to a follow-up SPEC if/when fan-in grows beyond the current 2-package boundary.
- **Hook handlers outside `internal/hook/`** — Any cwd resolution in `internal/cli/`, `internal/harness/`, or `cmd/moai/` is out of scope. This SPEC audits `internal/hook/` only.
- **Behavior change in CLAUDE_PROJECT_DIR semantics** — The env var contract with Claude Code is fixed by the upstream hook protocol. This SPEC consumes the env var as-is.
- **Performance optimization** — `os.Stat` calls in `resolveProjectRoot` add ~1µs latency per hook invocation. Optimization (caching) is out of scope.
- **Adding `.moai/` existence guard to `resolveQualityProjectDir`** — Quality-gate operations are contractually invoked from project roots; adding the guard would change semantics (linting from a subdirectory currently works). PRESERVE current behavior.

---

## Risks

| Risk | Severity | Mitigation |
|------|---------|------------|
| Refactor breaks an undocumented call pattern (e.g., a test that relies on cwd-fallback behavior) | Medium | Run full `go test -race ./internal/hook/...` before/after; preserve all `*_test.go` files unchanged; rely on AC-HCWA-003 (cohabitation guard) and AC-HCWA-004 (race detection) as safety net. |
| `observability_master.go` semantic shift — adding env-var preference may change behavior for users who relied on cwd-based config discovery | Low | Document in `plan.md` M2; the env var was always the documented contract per `config.EnvClaudeProjectDir` usage in sibling files; cwd-only behavior was an undocumented fallback. |
| `package quality` helper duplicates `package hook` logic, drift over time | Low | Document the duplication rationale in `plan.md` M3 § Notable Implementation Decisions; future SPEC can consolidate if cross-package shared helper proves necessary. |
| Warning log spam from `cwd_fallback: true` in production if `CLAUDE_PROJECT_DIR` is consistently unset | Low | Warning is at `slog.Warn` level (filterable); production deployments set `CLAUDE_PROJECT_DIR` via Claude Code; local-dev fallback is the intended log target. |
| Lint baseline drift catches false positives during refactor (e.g., unused variable warnings) | Low | M0 baseline measurement at plan-phase entry; AC-HCWA-005 is "0 NEW" relative to baseline, not absolute zero. |

---

## Cross-references

- **Prior fix** — commit `a9b3e8cd8` (`fix(hook): subagent_stop dispatchCapture cwd leak via resolveProjectRoot`)
- **Reference pattern** — `internal/hook/post_tool_metrics.go:98-113` (resolveProjectRoot)
- **Related SPECs**:
  - SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 (parent context — cwd-leak discovered post-merge)
  - SPEC-V3R6-HOOK-ASYNC-EXPAND-001 (sibling — async hook handlers)
  - SPEC-V3R6-HOOK-CONTRACT-FIX-001 (sibling — hook handler contract)
- **Memory references**:
  - `[[project-v3r6-sprint2-hoi-revise-ready]]` § "cwd-leak fix" sub-section
- **Project documentation**:
  - `.moai/project/tech.md` § Hook subsystem
  - `.claude/rules/moai/development/agent-authoring.md` § Background Agent Write Restriction

---

Version: 0.2.0
Status: implemented
Tier: S (minimal — 4 artifacts, no separate design.md/research.md)
Created: 2026-05-23
Last Updated: 2026-05-23
