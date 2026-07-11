---
id: SPEC-SESSIONSTART-PERF-001
title: "Session-Start Performance Durability — Design"
version: "0.1.0"
status: draft
created: 2026-07-11
updated: 2026-07-11
author: manager-spec
---

# Design — SPEC-SESSIONSTART-PERF-001

This document elaborates the HOW for the three milestones. It is a design aid for run-phase; final function names / signatures are indicative and may be refined during TDD/DDD provided the acceptance criteria hold.

## §M1 — Single-pass drift detection + HEAD-SHA cache

### §M1.1 Current control flow (as verified at HEAD 0303e8c7)

`DetectDrift(baseDir)` (drift.go:31):
1. `os.ReadDir(.moai/specs)` → iterate 477 entries.
2. Per SPEC: `ParseStatus` → skip on error.
3. Terminal early-return (`isTerminalStatus`) → append `Drifted:false`, continue (mechanism ③).
4. Era-final early-return (`LoadEraSignalsFromDir` + `ClassifyEra().EraFinal()`) → append `Drifted:false`, continue (mechanism ④).
5. `getGitImpliedStatus(specID)` → **subprocess #1** (`git log --grep=specID -50`).
6. Combined-scope fallback: if frontmatter=`completed` but git≠`completed` and not terminal → `resolveCombinedScopeClose(specID)` → **subprocess #2** (`git log --grep=<scope-prefix> -50`).
7. `drifted := frontmatterStatus != gitStatus`; append record; increment count on drift.

The two subprocess sites (steps 5 and 6) are the O(n) cost. Steps 3-4 already pre-filter cheaply (no git) — they stay.

### §M1.2 Target control flow (single global pass + in-memory walk)

```
DetectDrift(baseDir):
  # (A) HEAD-SHA cache short-circuit  [REQ-SSP-004]
  head := gitHeadSHA()                      # one cheap `git rev-parse HEAD` (or reuse cache)
  if cached := loadDriftCache(baseDir, head); cached != nil:
      return cached                          # zero further git work

  # (B) enumerate + cheap pre-filter  [REQ-SSP-003]
  active := []                               # SPECs needing git classification
  for each spec dir:
      status := ParseStatus(dir); skip on err
      if isTerminalStatus(status) or eraFinal(dir):
          emit record{Drifted:false}; continue     # PRESERVE emission contract
      active = append(active, {specID, status})

  # (C) single git log pass → in-memory FULL-MESSAGE index  [REQ-SSP-001, REQ-SSP-002]
  #     MUST capture the FULL commit message (subject + body), NOT --oneline (subject-only):
  #     the current per-SPEC `git log --grep=<specID>` matches the FULL message (subject OR body),
  #     so the candidate key MUST be a full-message substring match to replicate it exactly.
  commits := gitLogAllFullMessage()          # ONE subprocess: `git log <branch> --no-merges
                                             #   --format=<delimited hash + subject + body>`; newest-first
  index := buildIndex(commits)               # per-record {subject, body}; candidate match = strings.Contains(subject+body, key)
                                             #   map[specID][]record  AND  map[scopePrefix][]record

  # (D) per-SPEC in-memory walk — replicate the CURRENT TWO-STAGE semantics EXACTLY (drift.go:184-235)
  for each a in active:
      # stage 1 (candidate set): newest gitLogWindowSize records whose FULL message contains a.specID
      #   → replicates `git log --grep=<a.specID> -50` (full-message match + newest-N window)
      # stage 2 (re-filter + classify), newest-first: shouldSkipCommitTitle(subject) → then
      #   commitMatchesSPECID(subject, a.specID) SUBJECT word-boundary → then ClassifyPRTitle(subject);
      #   first non-empty classification wins
      gitStatus := inMemImpliedStatus(index, a.specID)
      if a.status == "completed" and gitStatus != "completed" and !terminal(gitStatus):
          # combined-scope fallback: newest-50 FULL-message matches of the scope-prefix,
          #   re-filtered via combinedScopeCloseMatches(subject) → replicates resolveCombinedScopeClose
          if inMemCombinedScopeClose(index, a.specID):
              gitStatus = "completed"
      emit record{Drifted: a.status != gitStatus}

  result := {records sorted by specID, count}
  saveDriftCache(baseDir, head, result)      # persist keyed on HEAD  [REQ-SSP-006]
  return result
```

`inMemImpliedStatus` replicates the current two-stage walk EXACTLY: (stage 1) the candidate set is the newest `gitLogWindowSize` records whose FULL commit message (subject+body) contains the specID substring — the in-memory equivalent of `git log --grep=<specID> -50`; (stage 2) each candidate is re-filtered newest-first via `shouldSkipCommitTitle(subject)` (chore-skip LSCSK-001) + `commitMatchesSPECID(subject, specID)` (subject word-boundary LSGF-001) + `ClassifyPRTitle(subject)` (first non-empty classification wins). `inMemCombinedScopeClose` replicates `resolveCombinedScopeClose`: newest-50 FULL-message matches of the scope-prefix, re-filtered via `combinedScopeCloseMatches(subject)`. **No classification helper changes** — only the commit source changes from a per-SPEC subprocess to the shared full-message index; the matching basis (full-message candidate + subject re-filter + window) is preserved bit-for-bit.

### §M1.3 Behavior-preservation invariant (the key risk — corrected per plan-audit D1/D2)

The current walker is **TWO-stage and its candidate set is body-dependent**: `git log --grep=<specID> -50` (drift.go:184) returns the newest 50 commits whose FULL message (subject OR body) contains the specID substring, then `commitMatchesSPECID(subject, specID)` (drift.go:218) re-filters to those whose SUBJECT carries the exact specID token via `ExtractSPECIDs(subject)`. The `-50` window is therefore applied to the FULL-MESSAGE-matched set, NOT the subject-matched set.

A single-pass index keyed on `ExtractSPECIDs(subject)` (subject tokens only) builds a **DIFFERENT candidate set** and can flip the count: a specID appearing in the BODIES of ≥ `gitLogWindowSize` commits newer than its own newest classifiable subject-close commit would exhaust the current walker's window (all body-only matches, all filtered out at stage 2 → walker returns error → skip), whereas a subject-only index would reach the deeper subject-close commit and classify it. Because the real-corpus count is **exactly 78** (plan.md §C baseline), a single such flip violates HARD REQ-SSP-005.

- **Chosen approach (exact two-stage replication)**: the single `git log` pass captures FULL messages (subject + body). Per active SPEC, replicate stage 1 (candidate set = newest `gitLogWindowSize` FULL-message-substring matches, newest-first — the in-memory equivalent of `--grep=<specID> -50`) then stage 2 (subject re-filter via `commitMatchesSPECID` + `shouldSkipCommitTitle` + `ClassifyPRTitle`, first classifiable wins). The invariant is stated over the *classifiable-candidate set* (full-message-matched, window-bounded, subject-re-filtered) — NOT a loose subject-only superset.
- **Equivalence proof**: for any SPEC, the in-memory candidate list (full-message substring matches, newest-first, capped at `gitLogWindowSize`) is order-identical to the set `git log --grep=<specID> -50` returns; applying the identical stage-2 re-filter + classifier yields the identical first-classifiable status. The single global pass is unbounded (no `-N`) so every per-SPEC 50-window is reachable; the 50-cap is applied per-SPEC in memory. Guarded by AC-SSP-005a/005b (synthetic) + AC-SSP-005c (real 78-count corpus).
- **Rejected alternative A (subject-only index)**: keying the index on `ExtractSPECIDs(subject)` only — REJECTED; it is the divergence hazard above (different candidate set → count flip). This is the plan-audit D1/D2 correction.
- **Rejected alternative B (bounded global window)**: a global `-K` cap (e.g. `-2000`) — REJECTED (AP-1); it could truncate before an older-but-active SPEC's window is reachable.
- **Cost note**: the full-message pass carries more bytes than `--oneline`, but it is still ONE streamed subprocess vs ~700 — the asymptotic O(1)-subprocess win (REQ-SSP-001) is preserved.

### §M1.4 HEAD-SHA cache design [REQ-SSP-004, REQ-SSP-006]

- **Key**: current HEAD SHA (`git rev-parse HEAD`).
- **Location**: `.moai/state/drift-cache.json` (gitignored runtime state, same family as `context-usage.json`, `active-sessions.json`; follows the `.moai/state/*.last` sentinel precedent from `sync-phase-quality-gate.sh`).
- **Payload**: `{ head_sha, computed_at, count, records[] }`.
- **Hit**: HEAD unchanged → return cached count with zero git subprocesses (AC-SSP-004).
- **Miss/stale**: HEAD changed or file absent/unparseable → recompute + overwrite (AC-SSP-006, AC-SSP-022).
- **Fail-open**: cache read/write errors never fail drift detection — on any cache error, fall through to a full compute (the current behavior with no cache).
- **On-demand freshness bypass (`--no-cache`)** [REQ-SSP-006a / AC-SSP-006a]: `moai spec drift --no-cache` bypasses the HEAD-SHA cache and recomputes freshly — the authoritative on-demand path. This flag is what makes the stale-frontmatter window (below) safe: the cached session-start count is advisory, and `--no-cache` gives the fresh authoritative view whenever an operator needs it.
- **Correctness note**: the cache is keyed ONLY on HEAD SHA. Uncommitted frontmatter edits do not advance HEAD, so a stale-frontmatter window exists between an edit and its commit — this matches the current advisory nature of the check (drift is best-effort, non-blocking) and is acceptable because the `--no-cache` bypass above (REQ-SSP-006a) provides the fresh authoritative view on demand. Recorded as residual risk.

### §M1.5 Reused infrastructure (NO reinvention)

- `cachedMainBranch()` (gitquery_cache.go:89) — branch detection, already REQ-PERF-001-A-cached. Reuse for the single-pass `git log <branch>`.
- `isTerminalStatus` (drift.go:323), `ParseStatus` (status.go:78), `LoadEraSignalsFromDir` + `ClassifyEra` + `EraFinal` (era.go) — reuse verbatim for the pre-filter.
- `ExtractSPECIDs`, `deriveScopePrefix`, `combinedScopeCloseMatches`, `shouldSkipCommitTitle`, `commitMatchesSPECID`, `ClassifyPRTitle` — reuse verbatim in the index + in-memory walk.

## §M2 — SPEC auto-archive

### §M2.1 Eligibility predicate

A SPEC is archive-eligible iff BOTH hold:
1. **Terminal status**: frontmatter status ∈ {`completed`, `superseded`, `archived`, `rejected`}. Per task, `completed` AND era-final/grandfather SPECs are INCLUDED — but only via criterion (2), never by grandfather status alone (REQ-SSP-010 / AP-2).
2. **Past grace window**: the SPEC's terminal-transition timestamp predates now − grace-days. Timestamp source (design decision, see research.md §F): prefer the git-log close-commit date (accurate); fall back to frontmatter `updated:` when git history is unavailable.

Grandfather protection is orthogonal: an era-final SPEC is eligible ONLY when (1)+(2) independently hold. The eligibility function does NOT consult `EraFinal()` as a gate — it consults status + date. (Era classification remains relevant only to the M1 drift path, not to archive eligibility.)

### §M2.2 Archive location (DECIDED: `.moai/archive/specs/<year>/`)

Decision D1 (research.md §F, user-approved 2026-07-11): the archive location is **`.moai/archive/specs/<year>/SPEC-...`** — consistent with the existing `.moai/archive/skills/` precedent (verified present). The rejected alternative `.moai/specs/archive/<year>/SPEC-...` would keep everything under `.moai/specs/` but require the drift scanner to skip the `archive` sub-dir.

Rationale for the decided path: the archived dir lands OUTSIDE the `os.ReadDir(.moai/specs)` SPEC-scan set automatically (different parent), so it no longer contributes to drift cost with zero scanner change. (The rejected alternative already skips non-`SPEC-`… entries via `ParseStatus` error, but an explicit skip would have been cleaner — moot under the decided path.)

### §M2.3 CLI surface

- New `internal/cli/spec_archive.go`, registered as a sibling of `spec_drift.go` under the `moai spec` command group.
- Flags: `--dry-run` (report only), `--grace-days N` (override configured default), optional `--json`.
- Move mechanism: `git mv` (or move + `git add`) so the relocation is a tracked rename — preserves history and grep-discoverability (REQ-SSP-007, AC-SSP-023).
- Reports moved (or eligible, for dry-run) set with counts.

### §M2.4 Config (Template-First) [REQ-SSP-012]

- Grace-window default key added to a config section. Options: a new `spec.yaml` section, or a key under an existing section (e.g. `system.yaml` or a new `archive.yaml`). Recommended: a small new `archive.yaml` (or a `spec_archive:` block) to keep the concern isolated.
- Template source FIRST: `internal/template/templates/.moai/config/sections/<name>.yaml` → `make build` → mirror local.
- Default grace-days value: **90** (DECIDED — research.md §F, decision D2, user-approved 2026-07-11). The `moai spec archive --grace-days` flag defaults to this config value; the config default and the `config/defaults.go` fallback (REQ-SSP-018) both carry 90.

### §M2.5 On-close trigger [REQ-SSP-013] / critical-path exclusion [REQ-SSP-011]

- The eligibility+relocation entry point is a plain function callable from (a) the `moai spec archive` CLI and (b) the `/moai sync` close path (documented trigger point; the actual sync invocation can be a follow-up wiring or a doc note — the CAPABILITY and its reachability are the deliverable).
- Guard: archive code is NEVER referenced from `internal/hook/session_start.go` (AC-SSP-011 grep guard).

## §M3 — Regression guard

### §M3.1 Performance-budget test [REQ-SSP-014]

- New test (e.g. `internal/spec/drift_perf_test.go`) builds N=500 synthetic SPEC dirs in `t.TempDir()` with a real (small) git repo, runs drift detection, and asserts elapsed < budget (default 2s, from config/defaults.go).
- The fixture must exercise the git path (a tiny git repo with a few commits) so the test measures the real algorithm, not a stub. Uses `t.TempDir()` for isolation (CLAUDE.local.md §6).
- Wired into `go test ./...` — no separate CI job needed; a budget violation fails the package test.

### §M3.2 Session-start time-box [REQ-SSP-015]

- Add a context-aware entry point, e.g. `DriftCountCtx(ctx context.Context, baseDir string) (int, error)`; keep `DriftCount(baseDir)` as a thin `context.Background()` wrapper for backward-compat with existing callers.
- In `detectStatusDrift` (session_start.go:935): wrap the call in `context.WithTimeout(ctx, timeBoxDeadline)` (default 2s, from config/defaults.go). On `ctx.Err() == context.DeadlineExceeded`, skip and emit the existing advisory string instead of blocking (AC-SSP-015).
- With M1 in place, the deadline is never hit at normal corpus sizes; the time-box is a safety net for pathological repos and a guarantee that the advisory check can never block the critical path unboundedly.

### §M3.3 Codified principle (Template-First) [REQ-SSP-016, REQ-SSP-017]

- Add a HARD note to `internal/template/templates/.claude/rules/moai/development/coding-standards.md` (or the hooks rule) → mirror local. Suggested wording: "Session-start / Stop advisory checks MUST be cached, asynchronous, or on-demand — never unbounded-blocking. An advisory computation on the session-start critical path MUST be time-boxed (context deadline) and degrade to an advisory message on timeout."

## §M4 — Threshold inventory (config/defaults.go) [REQ-SSP-018]

| Threshold | Default | Home | Consumer |
|-----------|---------|------|----------|
| Grace-window days | 90 (proposed) | template `.moai/config/` + config/defaults.go fallback | M2 eligibility |
| Perf budget | 2s | config/defaults.go | M3 perf test |
| Time-box deadline | 2s | config/defaults.go | M3 session-start |
| Git-log window (existing) | 50 (`gitLogWindowSize`) | already a const in drift.go | M1 in-mem walk cap |

## §M5 — Milestone dependency order

M1 → M2 → M3 (priority order). M1 is independent and is the actual latency fix. M2 depends on nothing in M1 but shares the era/terminal helpers. M3's time-box (§M3.2) is most valuable AFTER M1 (as a safety net), and its perf-budget test (§M3.1) validates M1's asymptotics — so M3 lands last.
