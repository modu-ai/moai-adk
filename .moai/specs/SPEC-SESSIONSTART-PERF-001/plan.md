---
id: SPEC-SESSIONSTART-PERF-001
title: "Session-Start Performance Durability — Implementation Plan"
version: "0.1.0"
status: completed
created: 2026-07-11
updated: 2026-07-11
author: manager-spec
---

# Implementation Plan — SPEC-SESSIONSTART-PERF-001

## §A Context

Session-start blocks ~13s (verified: `moai spec drift --count` → `13.140 total` at HEAD `0303e8c7`, 477 SPECs) due to O(n) `git log` subprocess spawns in `internal/spec/drift.go`. This plan delivers three milestones in priority order: **M1** (the actual latency fix — algorithmic hardening), **M2** (dataset bounding — SPEC auto-archive), **M3** (regression guard). See `spec.md` §A for the full root-cause chain and `research.md` for the infrastructure inventory and behavior-preservation risk analysis.

Recommended cycle_type per milestone:
- **M1 = ddd** (behavior-preserving refactor; characterization tests capture current output FIRST)
- **M2 = tdd** (new capability; test-first)
- **M3 = tdd** (new tests + rule codification; test-first)

## §B Known Issues / Constraints Carried In

- The current algorithm has subtle, hard-won correctness: chore-skip (LSCSK-001), word-boundary SPEC-ID match (LSGF-001), terminal-status authority (mechanism ③), grandfather era exemption (mechanism ④), and the combined-scope secondary prefix-grep fallback (mechanism ①). ALL of this behavior MUST be preserved — it is guarded by existing tests and the drift-convention SPEC line.
- `getGitImpliedStatus` walks commits **newest-first** and adopts the first classifiable status, capped at `gitLogWindowSize = 50` per SPEC. Any single-pass rewrite must reproduce this newest-first-first-classifiable-within-cap semantics (see design.md §M1).
- Per-run branch detection already uses `cachedMainBranch()` (from `internal/spec/gitquery_cache.go`) per REQ-PERF-001-A — REUSE it; do not re-add `git rev-parse` spawns.

## §C Pre-flight Checklist

- [ ] Read `internal/spec/drift.go`, `internal/spec/gitquery_cache.go`, `internal/spec/era.go`, `internal/spec/status.go`, `internal/spec/transitions.go` in full before editing.
- [ ] Capture the current `moai spec drift --json` output on the real corpus as the characterization baseline (store under test fixtures or `/tmp` and cite in §E).
- [ ] Confirm `.moai/state/` gitignore status and the `.moai/archive/skills/` precedent path shape.
- [x] Prior open decisions RESOLVED (user-approved 2026-07-11 — "proceed as proposed"): D1 archive location → `.moai/archive/specs/<year>/`; D2 grace-window default → 90 days (Template-First config). See research.md §F.

## §D Constraints

### §D.1 Go project rules (CLAUDE.local.md)

- No hardcoding: extract new thresholds to `config/defaults.go`; env overrides (if any) to `internal/config/envkeys.go` constants.
- File naming `snake_case.go` / `snake_case_test.go`; English comments and godoc.
- Error wrapping `fmt.Errorf("...: %w", err)`.
- Coverage: package ≥ 85%; critical packages (`internal/cli`, `internal/hook`) ≥ 90%.
- Observation-only discipline in `internal/spec` lint/drift code: no file-modify primitives inside classification helpers (drift.go stays read-only w.r.t. the working tree; the NEW archive code lives in `internal/cli` and DOES move files — that is its purpose and is outside the observation-only drift path).

### §D.2 Template-First (HARD — called out explicitly per task)

Two milestones touch template-managed surfaces and MUST follow the Template-First rule (edit `internal/template/templates/...` FIRST → `make build` → mirror to local):

- **M2 config** (REQ-SSP-012): the grace-window default config file lives at `internal/template/templates/.moai/config/sections/<name>.yaml` (or an added key in an existing section) FIRST, then mirrored to local `.moai/config/`.
- **M3 rule** (REQ-SSP-016 / REQ-SSP-017): the codified "advisory checks must be cached/async/on-demand" principle is added to `internal/template/templates/.claude/rules/moai/development/coding-standards.md` (or the hooks rule) FIRST, then mirrored to local `.claude/rules/...`.

Verification before commit: every new file under `.claude/` or `.moai/config/` has a corresponding template source file (CLAUDE.local.md §2 Template-First Rule).

### §D.3 Behavior preservation (HARD — M1)

The M1 refactor is behavior-preserving. The gate is the characterization test: current `DetectDrift` output (count + per-SPEC records) captured BEFORE the refactor must match AFTER the refactor on the representative fixture. This is a DDD ANALYZE-PRESERVE-IMPROVE cycle, not a rewrite.

## §E Self-Verification Plan (populated at run-phase; skeleton only at plan-phase)

Run-phase (manager-develop) will populate progress.md §E.2/§E.3 with:
- Characterization baseline capture + post-refactor equivalence proof (M1).
- `go test ./internal/spec/... ./internal/hook/... ./internal/cli/...` output.
- Coverage per package.
- Subprocess-count evidence (e.g., a test asserting a single `git log` invocation via an injected git runner, OR a strace/rusage-based comparison).
- `moai spec drift` before/after wall-clock.
- Lint clean (`golangci-lint run`).

## §F Milestones

### §F.1 M1 — Algorithmic hardening of drift detection [PRIORITY: HIGHEST] (cycle_type=ddd)

Files: `internal/spec/drift.go` (primary), `internal/spec/drift_test.go`, possibly a new `internal/spec/drift_cache.go` for the HEAD-SHA cache, `internal/spec/gitquery_cache.go` (reuse only), `config/defaults.go` (time-box / window constants if extracted).

Steps (ANALYZE → PRESERVE → IMPROVE):
1. **ANALYZE**: enumerate every semantic branch of `DetectDrift` + `getGitImpliedStatus` + `resolveCombinedScopeClose` (see design.md §M1.1). Identify the two subprocess sites (drift.go:184, drift.go:468).
2. **PRESERVE**: write characterization tests capturing current `DetectDrift` output on a fixture that exercises: a drifted SPEC, a non-drifted SPEC, a terminal SPEC, a grandfather SPEC, and a combined-scope-close SPEC. Confirm they pass on the pre-refactor code.
3. **IMPROVE (single-pass index)**: rewrite so `DetectDrift`:
   - Pre-filters terminal + era-final SPECs BEFORE any git work (REQ-SSP-003) — this preserves the existing early `records = append(..., Drifted:false)` emission contract (drift.go:71-98).
   - Runs ONE `git log --oneline --no-merges` pass (single subprocess, streamed), builds an in-memory index keyed by exact SPEC-ID token (via `ExtractSPECIDs`) AND by scope-prefix (via `deriveScopePrefix`) for combined-scope fallback (REQ-SSP-001, REQ-SSP-002).
   - For each active SPEC, performs the SAME newest-first-first-classifiable walk over its indexed commits, capped at `gitLogWindowSize` for exact equivalence (REQ-SSP-005).
   - Preserves `shouldSkipCommitTitle`, `commitMatchesSPECID`, `ClassifyPRTitle`, and the combined-scope 3-gate matcher verbatim.
4. **IMPROVE (HEAD-SHA cache)**: add a cache under `.moai/state/` keyed on current HEAD SHA (REQ-SSP-004, REQ-SSP-006). On a HEAD-SHA hit, return the cached `count` without git work. On miss, compute + persist. Follow the `.moai/state/*.last` sentinel pattern used by `sync-phase-quality-gate.sh`.
5. Re-run characterization tests → must match. Re-measure `moai spec drift` wall-clock → target well under 2s.

### §F.2 M2 — SPEC auto-archive [PRIORITY: MEDIUM] (cycle_type=tdd)

Files: new `internal/cli/spec_archive.go` + `internal/cli/spec_archive_test.go`; a new archive-eligibility helper in `internal/spec/` (e.g. `archive.go`) reusing `ParseStatus`, `isTerminalStatus`, `LoadEraSignalsFromDir`, `ClassifyEra`; config default in `internal/template/templates/.moai/config/sections/` + local mirror; wiring into the `moai spec` command group (sibling of `spec_drift.go`).

Steps (RED → GREEN → REFACTOR):
1. Define the eligibility predicate (RED test first): a SPEC is archive-eligible iff (a) frontmatter status is terminal (`completed` included, and era-final/grandfather included per task), AND (b) its terminal transition (git-log-derived close date, or frontmatter `updated`) predates the grace window (REQ-SSP-009). Grandfather-protected SPECs are eligible ONLY when they independently satisfy (a)+(b) — the guard for REQ-SSP-010.
2. Implement `moai spec archive [--dry-run] [--grace-days N]` (REQ-SSP-008): dry-run reports eligible set; non-dry-run moves each eligible SPEC dir to `<archive-location>/<year>/SPEC-...` via `git mv` (or move + stage) so it stays git-tracked and grep-discoverable (REQ-SSP-007).
3. Grace-window default: add to template config FIRST (REQ-SSP-012), read via config loader.
4. On-close trigger point (REQ-SSP-013): expose the eligibility+archive entry point so `/moai sync` close can invoke it; MUST NOT run in session-start (REQ-SSP-011 — assert archive code is never called from `internal/hook/session_start.go`).
5. Extract grace-window default + any literals to config/defaults.go (REQ-SSP-018).

### §F.3 M3 — Regression guard [PRIORITY: MEDIUM] (cycle_type=tdd)

Files: new perf-budget test in `internal/spec/` (e.g. `drift_perf_test.go`) or `internal/hook/`; time-box change in `internal/hook/session_start.go` (`detectStatusDrift`); a context-aware drift entry point in `internal/spec/` if `DriftCount` must accept a deadline; rule codification in `internal/template/templates/.claude/rules/moai/development/coding-standards.md` + local mirror; `make test` / CI wiring (the test runs under the standard `go test ./...`).

Steps (RED → GREEN → REFACTOR):
1. **Perf-budget test** (REQ-SSP-014): build a synthetic fixture of N=500 SPEC dirs in `t.TempDir()`, run drift detection, assert completion under budget (default 2s, extracted to config/defaults.go per REQ-SSP-018). Fails CI on regression.
2. **Time-box** (REQ-SSP-015): wrap the session-start drift call in a `context.WithTimeout` deadline (default 2s, extracted); on deadline exceed, skip and emit the existing advisory string rather than block. This likely requires a context-accepting variant of the drift entry point (e.g. `DriftCountCtx(ctx, baseDir)`), keeping the existing `DriftCount` as a thin wrapper for backward-compat.
3. **Codify principle** (REQ-SSP-016 / REQ-SSP-017): add a HARD note to the coding-standards (or hooks) rule template FIRST, mirror to local.
4. Confirm the time-box does NOT regress the M1 fast path (with M1, the deadline is never hit at normal corpus sizes; the time-box is a safety net for pathological repos).

## §G Anti-Patterns to Avoid

- **AP-1 (window shrink)**: making the single global `git log` window SMALLER than the per-SPEC 50-cap could reach — would silently drop close commits for older-but-active SPECs and change the count. The global pass must be a superset of every per-SPEC window (design.md §M1.3).
- **AP-2 (grandfather over-archive)**: archiving an era-final SPEC merely because it is grandfather-protected. Grandfather status alone is NOT archive-eligibility — the terminal+grace criteria must independently hold (REQ-SSP-010).
- **AP-3 (session-start archive)**: calling archive from the hot path. Archive is on-demand / on-sync-close only (REQ-SSP-011). A grep guard (`grep -n archive internal/hook/session_start.go` → 0) belongs in the AC set.
- **AP-4 (hardcoded thresholds)**: inlining 2s / 90-day / 500-fixture literals in business logic. All go to config/defaults.go (REQ-SSP-018).
- **AP-5 (local-first template edit)**: editing local `.claude/rules/` or `.moai/config/` before the template source — Template-First inversion (REQ-SSP-012, REQ-SSP-017).
- **AP-6 (semantic drift in refactor)**: "cleaning up" the chore-skip / word-boundary / combined-scope logic during the M1 rewrite. PRESERVE it verbatim; the characterization tests are the gate.

## §H Cross-References

- `spec.md` §A (root cause), §B (REQ), §C (exclusions)
- `design.md` (M1 single-pass index algorithm, HEAD-SHA cache, M2 archive design, M3 time-box design)
- `research.md` (infra inventory, prior SPECs, behavior-preservation risk, §F resolved decisions D1/D2/D3)
- `acceptance.md` (AC-SSP-001 .. AC-SSP-024)
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` (era classification + grandfather clause SSOT — REUSE read-only)
- `internal/spec/gitquery_cache.go` (`cachedMainBranch` — REUSE)
- CLAUDE.local.md §2 (Template-First), §14 (no hardcoding)
