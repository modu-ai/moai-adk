# SPEC-OBS-ENABLED-GATE-001 — progress.md

> Progress tracking for the observability `enabled`-key gate in `internal/cli/deps.go`.
> Section §E and its `§E.N` sub-sections are parser-load-bearing: `internal/spec/era.go`
> `ClassifyEra()` string-matches the literal `§E.2` / `§E.3` / `§E.4` heading tokens and
> the literal `sync_commit_sha` field name. Renaming any of them silently breaks era
> classification. The headings below are preserved verbatim per the Section Map SSOT in
> `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-02
plan_auditor_verdict: PASS
plan_auditor_score: 0.81
plan_auditor_iteration: iter-2
```

- Plan-auditor iter-2 PASS (0.81) — above the Tier S PASS threshold (0.75). Final
  iteration accepted without further escalation.
- Plan artifacts: spec.md (2-file Tier S default + acceptance.md) + plan.md. The
  Tier S artifact set is 2 files; acceptance.md was carried as a third artifact per
  the authoring judgment on a contract-violation SPEC (8 ACs, sabotage round-trip
  protocol).
- Design fork resolved: Option (a) — surgical Tier S fix reusing
  `hook.IsObservabilityEnabled()` as the SSOT; Option (b) (structural unification
  under package `hook`) explicitly deferred per spec.md §B Out of Scope.

## §E.2 Run-phase Evidence

### Implementation summary

- `internal/cli/deps.go:259` `enableObservabilityIfConfigured(reg, cwd)` — added a
  two-step gate:
  1. `os.Stat` file-existence gate (preserved — REQ-OEG-005 skip-on-missing
     intact, and the heavier `IsObservabilityEnabled` parse is skipped when the
     file is absent).
  2. `hook.IsObservabilityEnabled()` SSOT read (REQ-OEG-007) — consulted ONLY
     when the file exists. Returns false on absent key, false on `enabled: false`,
     false on parse error (safe-defaults). The gate returns early when the SSOT
     returns false, so `EnableObservability` is NOT called and the log directory
     is NOT created (REQ-OEG-002 / REQ-OEG-004).
- `internal/cli/deps_coverage_test.go` — added two regression tests using the
  `hook.SetObservabilityMasterForTesting(false)` + non-parallel +
  `t.Cleanup(hook.ResetObservabilityMasterForTesting)` seam pattern (per
  `internal/hook/notification_test.go:87` precedent):
  - `TestEnableObservabilityIfConfigured_DisabledByEnabledKey` (AC-OEG-001)
  - `TestEnableObservabilityIfConfigured_AbsentKeyDefaultsDisabled` (AC-OEG-003)

### 7-AC PASS matrix

| AC-ID | Status | Verification command | Verbatim output |
|-------|--------|---------------------|-----------------|
| AC-OEG-001 | PASS | `go test -run TestEnableObservabilityIfConfigured_DisabledByEnabledKey ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli  0.850s` (scoped run) |
| AC-OEG-002 | PASS | `go test -run TestEnableObservabilityIfConfigured ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli  0.850s` (existing `WithConfigFile` preserved) |
| AC-OEG-003 | PASS | `go test -run TestEnableObservabilityIfConfigured_AbsentKeyDefaultsDisabled ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli  0.850s` |
| AC-OEG-004 | PASS | `go test -run TestEnableObservabilityIfConfigured ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli  0.850s` (existing `NoConfigFile` preserved) |
| AC-OEG-005 | PASS | `go test ./internal/hook/...` | `ok` across all hook subpackages (4-quadrant `AC-HOI-007` unchanged) |
| AC-OEG-006 | PASS | `grep -c 'IsObservabilityEnabled' internal/cli/deps.go` → `4`; companion `grep -n 'hook.observability_events\|hook.opt_in.enabled' internal/cli/deps.go` excluding comments → `0` matches | SSOT reuse verified; no system.yaml key reads in the gate function |
| AC-OEG-007 | PASS | Sabotage round-trip (RED sabotage + restored GREEN captured by run-phase implementer in `deps_coverage_test.go` evidence block) | RED: under sabotaged `os.Stat`-only gate, `SetObservabilityMasterForTesting(false)` is ignored and `EnableObservability` is called → stub call-count assertion fails. Restored GREEN: scoped run exits 0. |

### Coverage

- `enableObservabilityIfConfigured`: 100% function-level coverage (all four branches
  exercised: file-missing, file-present+enabled-false, file-present+enabled-true,
  file-present+key-absent).
- `./internal/cli/` package: 76.4% (per the run-phase coverage measurement; the
  surgical scope of this SPEC touches one gate function + its tests, not the
  package-wide surface).

### Run-phase HEAD

- Run-phase commit: `468eb584a` — `fix(SPEC-OBS-ENABLED-GATE-001): M1 add enabled-key gate via IsObservabilityEnabled reuse`
- Single-milestone Tier S run (M1). No M2+ iterations.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-08-02
run_commit_sha: 468eb584a
```

- Orchestrator verification matrix (7-item read-only batch, 7/7 PASS):
  1. Scoped test suite — PASS (`ok  internal/cli  0.850s`)
  2. Hook cohabitation suite — PASS (`ok` across `internal/hook/...`)
  3. Full `internal/cli/` build — PASS
  4. Cross-platform build (windows/amd64) — PASS
  5. Subagent-boundary grep — PASS (no `AskUserQuestion` in CLI code)
  6. AC-OEG-006 SSOT-reuse grep — PASS (`IsObservabilityEnabled` x4; system.yaml keys 0)
  7. Coverage measurement — PASS (gate function 100%; package 76.4%)
- All 7 ACs GREEN with attributed command output (per §E.2 matrix).
- `AC-HOI-007` 4-quadrant cohabitation test untouched (no edit to the test file).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-08-02
sync_commit_sha: pending-backfill-20260802-220412
```

- The `sync_commit_sha` placeholder follows the self-referential-hazard workaround
  documented in `.claude/rules/moai/development/spec-frontmatter-schema.md`
  § Status Transition Ownership Matrix (SHA placeholder backfill exemption D3):
  a sync-phase commit cannot reference its own SHA — a commit does not know its
  own hash until after it lands. The real SHA will be backfilled in a follow-up
  commit by `manager-docs` (per the phase-owning-agent backfill exemption).
- Frontmatter transition carried by this single sync commit (3-phase close):
  `in-progress → implemented → completed` on `spec.md` (manager-docs owned; merged
  into the sync commit per the canonical 3-phase close — no separate Mx commit).
- MX Tag validation: no new MX tags required for this SPEC (the surgical gate change
  reuses an existing SSOT and adds a doc-comment referencing
  `SPEC-OBS-ENABLED-GATE-001 REQ-OEG-007` inside `deps.go`; no ANCHOR/WARN/TODO
  candidates surfaced — the function is not fan_in ≥ 3 and carries no dangerous
  pattern). MX is a cross-cutting sync concern, not a separate phase.

## §F Phase 4 Mode Selection

```yaml
tier: S
mode: sub-agent
mode_selection: Mode 5 (sequential sub-agent, single-milestone Tier S)
```

- **Input parameters**: tier = S; scope ≤ 2 files (deps.go + deps_coverage_test.go);
  domain count = 1 (`internal/cli` observability gate); file language mix = 100% Go;
  concurrency benefit = LOW (coding-heavy surgical work per Anthropic's coding-task
  parallelism caveat).
- **Decision**: Mode 5 (`sub-agent`, sequential). The coding-task parallelism caveat
  routes coding-heavy work to the sequential sub-agent path; Tier S + single-domain
  + sub-300-LOC scope makes Mode 5 the default fallback with no boundary-case
  ambiguity.
- **Justification**: Mode 4 (parallel) was not selected — single-domain coding work
  does not meet the multi-domain ≥3-domains / ≥10-files threshold and the
  coding-task parallelism caveat explicitly favors sequential sub-agents for
  coding-heavy tasks. Mode 6 (workflow) was not selected — scope is < 30 files and
  the change is surgical, not a uniform mechanical transform. Mode 5 is the safe
  default for coding-heavy Tier S work.
