---
id: SPEC-SESSIONSTART-PERF-001
title: "Session-Start Performance Durability — Research"
version: "0.1.0"
status: completed
created: 2026-07-11
updated: 2026-07-11
author: manager-spec
---

# Research — SPEC-SESSIONSTART-PERF-001

## §A Root-cause analysis (verified this session, HEAD 0303e8c7)

### §A.1 Reproduced measurement

```
$ time moai spec drift --count
78
moai spec drift --count  7.99s user  4.02s system  91% cpu  13.140 total
```

- Wall-clock 13.14s; CPU-bound (91% CPU, user+sys ≈ 12s). 78 SPECs currently drift.
- SPEC directory count: 477 (`ls -1d .moai/specs/SPEC-* | wc -l`).
- Task-brief independent reproduction: real 12.78s / user 7.9s / sys 4.3s; 50,282 involuntary context switches (subprocess-spawn thrash). Consistent.

### §A.2 Blocking chain (source-verified)

| Layer | File:line (HEAD 0303e8c7) | Fact |
|-------|---------------------------|------|
| Hook config | `.claude/settings.json:4-10` | `SessionStart` runs `handle-session-start.sh`, `type=command`, synchronous, `timeout: 30` |
| Handler | `internal/hook/session_start.go:185` | calls `detectStatusDrift(input.ProjectDir)` synchronously |
| Handler | `internal/hook/session_start.go:935-947` | `detectStatusDrift` → `spec.DriftCount(projectDir)`; warns when count ≥ 5 |
| Drift | `internal/spec/drift.go:31-153` | `DetectDrift`: `os.ReadDir(.moai/specs)`, per-SPEC classification |
| Drift | `internal/spec/drift.go:184-185` | subprocess #1: `git log <branch> --oneline --no-merges --grep=<specID> -50` |
| Drift | `internal/spec/drift.go:122,468-469` | subprocess #2 (fallback): `git log ... --grep=<scope-prefix> -50` |

Net: up to ~2× active-SPEC-count `git log` subprocess spawns per session start; O(n) in SPEC count.

## §B Existing infrastructure inventory (REUSE, do not reinvent)

| Capability | Symbol / path | Notes |
|-----------|---------------|-------|
| Branch detection (cached) | `cachedMainBranch()` — `internal/spec/gitquery_cache.go:89` | REQ-PERF-001-A per-run cache; reuse for single-pass `git log` |
| Git-query cache scaffolding | `startGitQueryCache`/`stopGitQueryCache` — `gitquery_cache.go` | wraps `Lint()`; drift path bypasses it currently |
| Terminal-status test | `isTerminalStatus` — `drift.go:323` | superseded/archived/rejected |
| Status parse | `ParseStatus` — `internal/spec/status.go:78` | frontmatter status |
| Era signals load | `LoadEraSignalsFromDir` — `internal/spec/era.go:300` | reads progress.md + frontmatter |
| Era classify | `ClassifyEra` / `(Era).EraFinal()` — `era.go:119 / :49` | grandfather protection |
| SPEC-ID token extraction | `ExtractSPECIDs` — `internal/spec/transitions.go` | exact-token index key |
| Scope-prefix derivation | `deriveScopePrefix` — `drift.go:345` | combined-scope index key |
| Combined-scope 3-gate matcher | `combinedScopeCloseMatches` — `drift.go:372` | reuse over indexed prefix commits |
| Classification | `ClassifyPRTitle`, `shouldSkipCommitTitle`, `commitMatchesSPECID` | preserve verbatim |
| State dir precedent | `.moai/state/` (present; `.gitkeep`, `context-usage.json`, `active-sessions.json`) | HEAD-SHA cache home |
| Sentinel `.last` precedent | `.moai/state/sync-quality-gate.last` — `sync-phase-quality-gate.sh:178` | cache-file pattern |
| Archive dir precedent | `.moai/archive/skills/` (present) — `internal/cli/update_archive.go` | archive-location pattern |
| CLI command precedent | `internal/cli/spec_drift.go:19` (`moai spec drift`) | sibling command shape for `moai spec archive` |
| Config defaults | `internal/config/defaults.go` (present), `internal/config/envkeys.go` (present) | threshold homes |
| Template config sections | `internal/template/templates/.moai/config/sections/` (present) | Template-First config source |

## §C Prior / related SPECs (avoid collision, build upon)

| SPEC | Relationship |
|------|--------------|
| `SPEC-DRIFT-001` | Original drift-detection SPEC. This SPEC hardens its performance without changing its semantics. |
| `SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001` / `-002` | Drift follow-ups. No overlap — this is a perf/architecture concern, not a semantic one. |
| `SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001` / `SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001` | Own the combined-scope + chore-skip + close-infix semantics that this SPEC PRESERVES verbatim. |
| `SPEC-INTERNAL-PERF-001` | Earlier internal perf SPEC; introduced `REQ-PERF-001-A` per-run git-query cache (`cachedMainBranch`). This SPEC extends that cache philosophy to the drift path with a HEAD-SHA result cache. Related, non-colliding. |
| `SPEC-HOOK-SESSIONSTART-PROBE-001` | Session-start probe SPEC. Related surface (session-start), distinct concern (this SPEC targets the drift bottleneck specifically). No file collision expected; run-phase should re-verify no overlapping edits. |

No SPEC currently owns "drift subprocess-count reduction" or "SPEC auto-archive" — this SPEC is the owner.

## §D Behavior-preservation risk analysis (M1)

The dominant risk (corrected per plan-audit D1/D2): the current walker is **TWO-stage** — `git log --grep=<specID> -50` (drift.go:184) matches the FULL commit message (subject OR body), then `commitMatchesSPECID(subject, specID)` (drift.go:218) re-filters on the SUBJECT. A naive single-pass index keyed on `ExtractSPECIDs(subject)` (subject-only) builds a DIFFERENT candidate set and can flip the count when a specID appears in the BODIES of ≥ `gitLogWindowSize` commits newer than its own subject-close commit. Because the real-corpus count is exactly 78, a single flip violates HARD REQ-SSP-005. A secondary variant of AP-1 (a bounded global `-K` window truncating an older-but-active SPEC's reach) also holds. Mitigation:

- **Capture full messages, replicate two-stage exactly**: one **unbounded** single `git log --no-merges` pass carrying subject + body (one subprocess, streamed); per SPEC, replicate stage 1 (newest `gitLogWindowSize` FULL-message-substring matches) + stage 2 (subject re-filter via `commitMatchesSPECID` + `shouldSkipCommitTitle` + `ClassifyPRTitle`). NOT a subject-only superset. See design.md §M1.3.
- **Empirical backstop**: gate with characterization tests (AC-SSP-005a/005b synthetic) PLUS AC-SSP-005c — the real 78-count corpus equivalence check that the synthetic fixture cannot reach. This is a DDD ANALYZE-PRESERVE-IMPROVE cycle.
- **Cost**: one full-message `git log` over history carries more bytes than `--oneline` but is still ONE streamed subprocess vs ~700 — the asymptotic O(1)-subprocess win (REQ-SSP-001) is preserved.

Secondary risk: the HEAD-SHA cache serves a stale count while frontmatter is edited but uncommitted (HEAD unchanged). Accepted — the check is advisory / non-blocking by design; the on-demand `moai spec drift --no-cache` bypass (REQ-SSP-006a / AC-SSP-006a) gives the fresh authoritative view whenever needed. Recorded as residual risk.

## §E Verification methodology for subprocess-count claim

To satisfy AC-SSP-001 without relying on rusage alone, run-phase should introduce a seam: an injectable git-command runner (interface) so the test can COUNT `git log` invocations directly and assert the count is constant as N grows. rusage/context-switch counts are corroborating evidence, not the primary assertion (per verification-claim-integrity: claims must be mechanically observed). The seam also enables AC-SSP-004 (assert 0 git invocations on cache hit).

## §F Resolved decisions (user-approved 2026-07-11 — "proceed as proposed")

Both former open decisions were resolved with the recommended defaults after an orchestrator AskUserQuestion round; the user directed "proceed as proposed". No open markers remain. All other "e.g." values from the task brief are adopted as recommended defaults (perf budget 2s, time-box 2s, N=500 fixture).

- **D1 — archive location path → DECIDED: `.moai/archive/specs/<year>/`**. Rationale: matches the existing `.moai/archive/skills/` precedent (verified present) and lands OUTSIDE the `os.ReadDir(.moai/specs)` drift-scan set automatically, so archived SPECs no longer contribute to drift cost. The rejected alternative `.moai/specs/archive/<year>/` would have required an explicit scanner skip. M2 REQ-SSP-007 and AC-SSP-007/023 name this concrete path.
- **D2 — grace-window default value → DECIDED: 90 days**. Rationale: terminal SPECs older than a quarter are unlikely to need active drift tracking. Defined as a Template-First config default (REQ-SSP-012) in `internal/template/templates/.moai/config/` (mirrored locally); the `moai spec archive --grace-days` flag defaults to this config value (REQ-SSP-009). Remains user-tunable via config or the flag override.

Additional design decision (not blocking, recorded for run-phase):
- **D3 — terminal-transition timestamp source for grace-window**: prefer git-log close-commit date; fall back to frontmatter `updated:` when git history is unavailable. Recorded in design.md §M2.1; no user input required.

## §G External references

- Go `os/exec` subprocess cost: each `exec.Command().Output()` forks+execs a `git` process — the 50,282 involuntary context switches directly reflect this thrash. A single streamed `git log` amortizes to one fork+exec.
- MoAI hook timeout policy: MoAI tightens the Claude Code 10-min default to 5s (internal/hook/CLAUDE.md); the local dev settings.json uses `timeout: 30` for SessionStart. The time-box (M3, 2s) is well within either.
