---
id: SPEC-HOOK-PRETOOL-PERF-001
title: "PreToolUse hook path lightweighting: collapse the fork+exec+full-config-per-invocation cost via config disk caching and a lazy config slice"
version: "0.1.0"
status: completed
created: 2026-08-12
updated: 2026-08-13
author: manager-spec
priority: P1
phase: "v3.0.2 target"
module: "internal/hook, internal/config"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "hook, pretool, performance, config-cache, lazy-config, latency, root-cause"
related_specs: "SPEC-PRETOOL-GATE-MOVE-001"
---

## HISTORY

| Version | Date       | Author        | Change |
|---------|------------|---------------|--------|
| 0.1.0   | 2026-08-12 | manager-spec  | Initial draft — plan-phase artifacts (Tier M). Root-cause lightweighting of the PreToolUse fork+exec+full-config-per-invocation cost. The 5→10s timeout widening already landed in `internal/template/templates/.claude/settings.json.tmpl` + `internal/hook/CLAUDE.md` is the IMMEDIATE mitigation; this SPEC is the ROOT fix that should let the timeout return toward 5s once per-invocation cost drops. Complementary to (non-overlapping with) SPEC-PRETOOL-GATE-MOVE-001, which relocated the synchronous vet/lint/test gate OFF the PreToolUse budget but did not touch the residual fork+exec+config cost this SPEC attacks. |

## A. Context (Why)

### A.1 The defect (measured baseline, captured this session)

Every PreToolUse invocation forks a fresh `moai` process via `.claude/hooks/moai/handle-pre-tool.sh` → `moai hook pre-tool`, and that fresh process loads the **full** merged config (`internal/config/manager.go:59` `ConfigManager.Load` → `internal/config/loader.go:31` `Loader.Load`, which runs ~20 per-section loaders, each reading a `.moai/config/sections/*.yaml` file). The normal-path cost is small:

| Phase | Measured cost | Measurement |
|-------|---------------|-------------|
| Full PreToolUse path (`handle-pre-tool.sh` + `moai hook pre-tool`) | 0.05s | 5/5 runs stable |
| `moai` binary cold start (`moai version`) | 0.056s | fresh process, no hook work |
| git rev-parse subprocesses (branch-guard) | 0.045–0.062s each | NOT the cause — `Workflow.BranchGuard.Enabled` defaults `false` (`internal/config/defaults.go`), so these do NOT run in this repo |
| OTEL/exporter | n/a | no OTEL env configured → no remote call; NOT the cause |
| `observability_master` (`internal/hook/observability_master.go`) | fast | sync.Once file read; NOT the cause |

But under system load and/or concurrent hook execution, the fork+exec+full-config cost amplifies intermittently:

- **33 timeouts across 30 days**, ALL on `PreToolUse:Bash`.
- The spikes are **scattered, NOT cold-start-clustered**: only ~9% fall in the first 10% of each session's hook stream.
- The spikes are **<1% of total PreToolUse invocations** — rare enough to be invisible to a smoke test, frequent enough to stall real sessions.

### A.2 Root cause (structural, not a deterministic hot spot)

There is **NO deterministic 5s hot spot in the code**. The spike is the **fork+exec+full-config-load STRUCTURE** amplifying under system load:

1. **fork+exec** — a fresh `moai` binary process per invocation. Binary cold start alone is 0.056s; under load, process creation contends for CPU/memory/pages.
2. **full config load** — ~20 per-section YAML reads + parse + merge, charged to every invocation, even though the PreToolUse handler only consumes a thin slice (security policy, branch-guard flag, gate config).
3. **concurrent hook execution** — when multiple PreToolUse events fire near-simultaneously (compound Bash commands, agent fan-out), each forks its own `moai` process and each performs the full config load, multiplying contention.

The 5→10s timeout widening (already landed) absorbs the spike as immediate mitigation. This SPEC is the ROOT fix that attacks the structure, so the timeout can return toward 5s once per-invocation cost drops.

### A.3 Non-goals (why this is not "just find the slow function")

The spike is intermittent and **not reproducible at will**. A profiler attached to a single invocation measures the 0.05s normal path, not the 5s tail. Therefore the SPEC cannot promise "the function X is the bottleneck" — it must instead collapse the structural cost that AMPLIFIES under load, and must include a milestone that CONFIRMS the collapse moves the needle before the implementation is declared sufficient.

## B. Scope

### B.1 In scope

- The PreToolUse hot path ONLY: `handle-pre-tool.sh` → `moai hook pre-tool` → `preToolHandler.Handle` (`internal/hook/pre_tool.go:370`).
- Config disk caching: persist the merged config to a mtime-keyed cache file under `.moai/state/`, invalided on section-file mtime change OR deletion.
- Lazy config slice: load only the config slices the PreToolUse handler actually consumes.
- A profiling milestone (before AND after implementation) that captures per-phase timing under simulated concurrent-hook stress, to CONFIRM the cache+lazy approach moves the needle.
- The explicit security trade-off of any future fast-path, recorded as a binding REQ + AC even though the fast-path itself is deferred.

### B.2 Exclusions (deferred to §E)

The full exclusion list (daemon, fast-path implementation, other hook events, destructive-primitive set expansion, timeout narrowing itself) lives in §E with `### Out of Scope —` H3 sub-headings and bullet items; this subsection is a pointer, not the authoritative exclusion surface.

## C. Requirements (GEARS)

### REQ-PERF-001 (Ubiquitous) — Config disk cache

The config loader SHALL persist the merged config to a cache file under `.moai/state/` and, on subsequent PreToolUse invocations, SHALL serve the cached config directly when no `.moai/config/sections/*.yaml` file's mtime has changed since the cache was written, skipping the per-section read+parse+merge.

### REQ-PERF-002 (When / event-detected) — Cache invalidation on section mtime change

**When** any `.moai/config/sections/*.yaml` file's mtime is newer than the cache's recorded mtime fingerprint, the config loader SHALL re-read and re-merge the affected section files, SHALL rewrite the cache, and SHALL serve the freshly merged config to the PreToolUse handler.

### REQ-PERF-003 (shall not) — Cache invalidation on section deletion

The config disk cache SHALL NOT serve a stale cached config when a section file present at cache-write time has since been removed. A deleted section file is an invalidation signal equal in force to an mtime change.

### REQ-PERF-004 (When / event-detected) — Fail-open on corrupt or schema-mismatched cache

**When** the cache file is corrupt, unreadable, or carries a schema version that does not match the running binary's config schema, the config loader SHALL fall back silently to the full re-merge path, SHALL emit no user-facing error, and SHALL rewrite the cache on successful re-merge.

### REQ-PERF-005 (While / state-driven) — Lazy config slice for PreToolUse

**While** servicing a PreToolUse event with a cache miss, the config provider SHALL load only the config slices the PreToolUse handler actually consumes (security policy inputs, branch-guard enable flag, gate config) rather than the full ~20-section set, UNLESS a full config is required for correctness. On a cache hit, the full cached config is served directly at cache-read cost and this lazy-load path is not exercised.

### REQ-PERF-006 (Where / capability gate) — Profiling milestone is a implementation-gate

**Where** the profiling milestone has not produced BOTH a baseline (pre-change) AND a post-change per-phase timing measurement (fork/exec, config load, security scan) under simulated concurrent-hook stress, the implementation SHALL NOT be declared sufficient. The measurement, not the code change, is the evidence of root-cause resolution.

### REQ-PERF-007 (When / event-detected) — SECURITY: fast-path must not bypass destructive-primitive scan

**When** a fast-path that short-circuits any subset of PreToolUse invocations is introduced — whether by this SPEC, a follow-up SPEC, or any other change — the fast-path SHALL NOT bypass the dangerous-pattern / security scan for any Bash command matching the Bash Risk-Amplifier destructive-primitive set documented in `.claude/rules/moai/development/coding-standards.md` § Bash Risk-Amplifier Doctrine (3): `rm -rf`, `git push --force` (and `git push -f`), `git push --no-verify`, `git commit --no-verify`, `git reset --hard`, SQL `DROP TABLE` / `TRUNCATE`, `chmod -R 777`; nor for any Bash command whose compound-subcommand count exceeds `BASH_SUBCOMMAND_SOFT_CAP` (5). This REQ binds the fast-path design space even though this SPEC defers the fast-path itself to a follow-up.

### REQ-PERF-008 (While / state-driven) — Atomic cache write

**While** two or more concurrent PreToolUse processes attempt to write the cache, the cache write SHALL be atomic (write-to-temporary-file + `os.Rename`), so that no reader ever observes a partially-written cache file. A reader that loses the race SHALL either read the previous valid cache or trigger the fail-open re-merge path (REQ-PERF-004).

### REQ-PERF-009 (Where / capability gate) — Cache location under state dir

**Where** the project resolves its state directory (default `.moai/state/`), the config cache file SHALL live under that directory so it inherits the existing state-dir lifecycle (gitignore, cleanup, worktree-aware path resolution) and does not introduce a new top-level path.

### REQ-PERF-010 (When / event-detected) — Timeout may return toward 5s only after cost drops

**When** the post-change profiling milestone (REQ-PERF-006) demonstrates that per-invocation PreToolUse cost under concurrent-hook stress has dropped below the threshold that previously produced the 33-timeouts/30-days tail, the 10s timeout configuration in `internal/template/templates/.claude/settings.json.tmpl` + `internal/hook/CLAUDE.md` SHALL be eligible to return toward the 5s MoAI policy default. The timeout SHALL NOT be narrowed speculatively; the measurement is the trigger.

## D. Constraints

- **C-1 (fail-open invariant)**: The PreToolUse hook is a critical-path gate. Any cache/scan change that could block, panic, or exit non-zero on a corrupt cache, a missing state dir, or a filesystem error MUST fail open (serve the full re-merge path or allow the tool call) — never stall the user's session. This invariant composes with the existing fail-open discipline in `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline.
- **C-2 (no daemon)**: This SPEC SHALL NOT introduce a long-lived hook daemon (process persistence). A daemon is a large architectural change with its own lifecycle, IPC, and security surface. The config cache + lazy slice directions are preferred; a daemon is justified only by a follow-up SPEC if the profiling milestone shows them insufficient.
- **C-3 (no new config format)**: The cache file format SHALL be a forward-compatible encoding (JSON or gob) of the existing `*Config` struct. This SPEC SHALL NOT invent a new config schema. A schema-version field in the cache file guards against binary/version skew (REQ-PERF-004).
- **C-4 (template neutrality / CLAUDE.local.md §2)**: Any change to `handle-pre-tool.sh` or `settings.json.tmpl` MUST be made in the template source (`internal/template/templates/`) first, regenerated via `make build`, and verified to carry no SPEC IDs, commit SHAs, or moai-adk-internal development state per the template-neutrality doctrine.
- **C-5 (worktree path awareness)**: The cache invalidation logic MUST use the same `$CLAUDE_PROJECT_DIR`-first resolution priority as the rest of the hook package (`internal/hook/observer.go`), so that a hook running inside a worktree does not read or write the primary checkout's cache.
- **C-6 (profiling reproducibility)**: Because the spike is not reproducible at will, the profiling milestone MUST use simulated concurrent-hook stress (multiple parallel `moai hook pre-tool` invocations against a fixture project) rather than waiting for an organic spike.

## E. Out of Scope

### Out of Scope — Long-lived hook daemon

- A persistent `moai` process (daemon) that PreToolUse invocations IPC into is explicitly OUT of scope. It is a large architectural change with lifecycle, IPC, and security implications that dwarf the cache+lazy fix. It MAY be revisited in a follow-up SPEC only if the profiling milestone (REQ-PERF-006) shows the cache+lazy approach insufficient.

### Out of Scope — Fast-path implementation

- The actual fast-path (a lightweight pre-check that short-circuits obviously-safe commands before paying the full config + scan cost) is DEFERRED to a follow-up SPEC. This SPEC records the security trade-off as a binding REQ (REQ-PERF-007) and AC (AC-PERF-007) so any future fast-path inherits the constraint, but does NOT implement the fast-path itself. The shell wrapper's existing bash-risk warn (`handle-pre-tool.sh`) is the only in-shell short-circuit and is unchanged by this SPEC.

### Out of Scope — Other hook events

- SessionStart, PostToolUse, Stop, SubagentStop, and other hook events are OUT of scope even if they share the config-load cost. They are NOT on the 33-timeouts/30-days hot list. A later SPEC MAY port the cache to other events if profiling justifies it; this SPEC targets PreToolUse only.

### Out of Scope — Dangerous-primitive set expansion

- This SPEC does NOT expand the Bash Risk-Amplifier destructive-primitive set. It references the existing set from `coding-standards.md` § Bash Risk-Amplifier Doctrine (3) verbatim. Expanding the set is an independent doctrine change.

### Out of Scope — Timeout narrowing itself

- The actual narrowing of the 10s timeout back toward 5s is the LAST step and is owned by a follow-up after the profiling milestone demonstrates the cost drop. This SPEC makes the timeout ELIGIBLE to be narrowed (REQ-PERF-010) but does NOT narrow it within this SPEC's run-phase scope, because the measurement must precede the narrowing.

## F. Cross-References

- `.claude/rules/moai/development/coding-standards.md` § Bash Risk-Amplifier Doctrine — the destructive-primitive set REQ-PERF-007 binds to, and the `BASH_SUBCOMMAND_SOFT_CAP` (5) doctrine constant.
- `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline — the fail-open / time-box / constant-cost doctrine C-1 composes with.
- `internal/hook/CLAUDE.md` — confirms the 10s timeout widening as immediate mitigation pending this root-fix SPEC.
- `internal/hook/pre_tool.go:370` (`preToolHandler.Handle`) — the handler entry point; consumer of the config slice REQ-PERF-005 targets.
- `internal/config/manager.go:59` (`ConfigManager.Load`) + `internal/config/loader.go:31` (`Loader.Load`) — the config-load path the cache intercepts.
- `internal/hook/branch_guard.go` + `internal/config/defaults.go` (`Workflow.BranchGuard.Enabled` default `false`) — evidence the git-rev-parse subprocesses are NOT the cause (the guard is off by default).
- `SPEC-PRETOOL-GATE-MOVE-001` (completed) — complementary; relocated the synchronous vet/lint/test gate OFF the PreToolUse budget but did not touch the residual fork+exec+config cost this SPEC attacks. Non-overlapping.
- `internal/config/atomicfile` (SPEC-CONFIG-ATOMIC-WRITE-001) — the atomic-write helper available for REQ-PERF-008; reuse, do not reinvent.
