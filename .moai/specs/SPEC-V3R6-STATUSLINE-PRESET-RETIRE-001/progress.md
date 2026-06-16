---
id: SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001
title: "Progress — Retire statusline preset system + remove web-console statusline panel"
version: "0.2.0"
status: draft
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "progress"
lifecycle: spec-anchored
tags: "progress, statusline, preset, mode, retire"
tier: M
---

# Progress — SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001

## §A. Current Phase

**run-phase** — Implementation Kickoff Approval obtained; plan-auditor verdict
PASS-WITH-DEBT 0.86 iter-2 (Tier M threshold 0.80). cycle_type=ddd. Mode 5
(sub-agent, sequential) selected at Phase 0.95. M1 baseline captured.

Execution location: L1 isolation worktree
`.claude/worktrees/agent-adc364de9bbfaf4ec`
(branch `worktree-agent-adc364de9bbfaf4ec`, base `b3c5dd12e2`).

## §B. Artifact Status

| Artifact | Path | Status |
|----------|------|--------|
| spec.md | `.moai/specs/SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001/spec.md` | draft v0.2.0 (24 REQs, 28 ACs, 8 exclusions) |
| plan.md | `.moai/specs/SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001/plan.md` | draft v0.2.0 (6 milestones M1-M6, D1-D7 remediation) |
| acceptance.md | `.moai/specs/SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001/acceptance.md` | draft v0.2.0 (25 MUST + 3 SHOULD ACs) |
| progress.md | this file | §E skeleton emitted |

## §C. Milestone Tracker

| Milestone | Scope | Status | Evidence |
|-----------|-------|--------|----------|
| M1 | Research / baseline capture | pending | _<pending run-phase>_ |
| M2 | Go code removal (models/statusline/cli/profile) | pending | _<pending run-phase>_ |
| M3 | Web console removal (fieldsets/handlers/validate + 3 tests) | pending | _<pending run-phase>_ |
| M4 | Wizard cleanup (profile_setup.go) | pending | _<pending run-phase>_ |
| M5 | Template + docs-site (4-locale) | pending | _<pending run-phase>_ |
| M6 | Final verification + sync prep | pending | _<pending run-phase>_ |

## §D. Blockers / Open Items

None at plan-phase. The 3 open questions in `spec.md §F.2` are run-phase
decisions, not user-blocking.

---

## §E.1 Plan-phase Audit-Ready Signal

The plan-phase artifact set is complete and internally consistent (v0.2.0,
iter-1 FAIL remediation applied):

- **spec.md**: 12 canonical frontmatter fields present; SPEC ID matches the
  canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (decomposition: SPEC ✓ |
  V3R6 ✓ | STATUSLINE ✓ | PRESET ✓ | RETIRE ✓ | 001 ✓ → PASS); 24 GEARS-format
  requirements (REQ-SPR-001 through REQ-SPR-024, with REQ-SPR-022/023/024
  added in v0.2.0 for D2 fieldset-caller, D3 wizard mode Select, D6 i18n
  cleanup); 8 explicit exclusions; 3 confirmed user decisions captured in §A.1
  (preset retire + web panel removal + mode preferences axis retire).
- **plan.md**: 6 milestones (M1-M6) with ordered edits, dependency-aware
  sequencing (struct fields first, then consumers); D1-D7 iter-1 defects
  addressed (lowercase presetToSegments wrapper D1, fieldset caller D2, mode
  axis D3, 7 test files D4, statuslineData D5, mode i18n D6, docs mode: D7);
  8 anti-patterns documented; pre-flight checklist present.
- **acceptance.md**: 28 ACs (25 MUST + 3 SHOULD), all observable (grep/build/
  test/byte-diff verifiable); traceability matrix covers 24/24 requirements;
  6 edge cases enumerated. AC-SPR-025/026/027/028 added in v0.2.0 for D3
  wizard mode Select, D6 i18n, D4 test files, D3 Builder API preservation.
- **progress.md**: this file — §E.2 through §E.5 emitted as placeholder
  headings only (no populated evidence; that belongs to run/sync/Mx phases).

**Era classification (per `.claude/rules/moai/workflow/lifecycle-sync-gate.md`)**:
H-2 fallback would fire on this minimal progress.md (no `§E.2`-`§E.5` markers
populated yet) — but the SPEC is freshly created at plan-phase, so the H-5
tie-breaker (created ≥ 2026-04-01) AND the `era: V3R6` override (to be set if
needed) keep it in the V3R6 bucket. No grandfather-clause risk: this is a new
SPEC, subject to modern-era drift detection once run-phase populates §E.2.

**Ready for**: plan-auditor independent audit → Implementation Kickoff
Approval (human gate, §19.1 CLAUDE.local.md) → `/moai run` delegation to
manager-develop with `cycle_type=ddd`.

---

## §E — Phase 0.95 Mode Selection

**Input parameters**:
- tier: M (300–1000 LOC, 5–15 files)
- scope: ~25 files (Go source + templ + tests + template + 4-locale docs)
- domain count: 3 (Go code / web templ / docs-site)
- file language mix: Go + templ + YAML + Markdown
- concurrency benefit: LOW (coding-heavy removal — mechanical but interdependent; struct-field removal ripples into consumers in the same package)
- Agent Teams prereqs: not met (harness standard, team.enabled not asserted)

**Decision**: `sub-agent` (Mode 5, sequential) — but executed orchestrator-direct
inside this L1 worktree (single manager-develop agent, no further sub-spawn).

**Mode evaluation**:
- Mode 1 trivial: NO (multi-package removal, not a typo)
- Mode 2 background: NO (writes files; background Write-denied)
- Mode 3 agent-team: NO (prereqs not met; scope under the 3-domain team threshold anyway)
- Mode 4 parallel: NO (coding-heavy; per Anthropic coding-task parallelism caveat)
- Mode 5 sub-agent: YES (selected — coding-heavy sequential default)
- Mode 6 workflow: NO (under ~30-file mechanical threshold; inter-file dependencies via struct fields)

**Justification**: This is coding-heavy removal work (struct field deletion
rippling into sync/handlers/wizard/tests). Per Anthropic's coding-task parallelism
caveat (Finding A4), sequential single-agent execution is the safe default.
The work is bounded enough (~25 files) for a single manager-develop pass.

---

## §E.2 Run-phase Evidence

### M1 — Pre-retire baseline (captured 2026-06-17, worktree HEAD b3c5dd12e2)

**Build baseline** (must remain green after every milestone):

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

**Lint baseline**:

```
$ golangci-lint run --timeout=2m
0 issues.
```

**Test + coverage baseline** (5 affected packages):

```
$ go test -cover ./internal/statusline/... ./internal/profile/... ./internal/web/... ./internal/cli/... ./pkg/models/...
FAIL    github.com/modu-ai/moai-adk/internal/statusline  4.208s  coverage: 82.9% of statements
ok      github.com/modu-ai/moai-adk/internal/profile      0.418s  coverage: 80.5% of statements
ok      github.com/modu-ai/moai-adk/internal/web          1.203s  coverage: 72.4% of statements
ok      github.com/modu-ai/moai-adk/internal/cli          9.800s  coverage: 71.8% of statements
ok      github.com/modu-ai/moai-adk/pkg/models            1.852s  coverage: 100.0% of statements
```

**PRE-EXISTING baseline failures (NOT caused by this SPEC)** —
`internal/statusline` memory_test.go has 6 failing subtests
(`TestCollectMemory/current_usage_calculation` +
`TestCollectMemory_AutoCompactScaling` × 5) expecting `TokenBudget = 200000`
but getting `1000000`. These live in `memory.go`/`memory_test.go` which the
SPEC explicitly PRESERVES (§A.3 segment rendering, AC-SPR-019 byte-parity).
They are a pre-existing baseline condition unrelated to preset/mode removal.
Post-retire statusline test result MUST remain "same 6 pre-existing failures,
no NEW failures introduced by this SPEC".

**Coverage reference** (post-retire MUST be ≥ these per-package values):
- statusline 82.9% (note: memory.go paths PRESERVED; preset.go PresetToSegments
  removal will shift denominator — net delta reported at M6)
- profile 80.5%
- web 72.4%
- cli 71.8%
- pkg/models 100.0%

**Toolchain**: templ v0.3.1020 available at `/Users/goos/go/bin/templ`.

**Target-location drift check** (spec.md §A.2 vs actual code):
- `pkg/models/config.go:199` Preset field — CONFIRMED
- `internal/statusline/preset.go:16-69` PresetToSegments — CONFIRMED
- `internal/cli/update.go:2732-2738` lowercase wrapper — CONFIRMED
- `internal/cli/statusline.go:52-57` preset fallback — CONFIRMED
- `internal/cli/statusline.go:123-127` statuslineFileConfig.Preset — CONFIRMED
- `internal/profile/preferences.go:45-46` StatuslineMode + StatuslinePreset — CONFIRMED
- `internal/profile/sync.go:80-91` statuslineData.Preset — CONFIRMED
- `internal/profile/sync.go:114-142` preset default + expansion switch — CONFIRMED
- `internal/web/fieldsets.templ:120-153` fieldsetStatusline — CONFIRMED
- `internal/web/root.templ:108` @fieldsetStatusline(view) caller — CONFIRMED
- `internal/web/handlers.go:368` StatuslinePreset binding (plan said 373; actual 368 — minor drift, non-blocking) — CONFIRMED
- `internal/web/validate.go:43-44` statuslinePresetCanonical — CONFIRMED
- `internal/web/validate.go:134-136` preset validation rule — CONFIRMED
- docs-site 4-locale statusline.md L211 (mode:) + L213 (preset:) — CONFIRMED all 4 locales

**Web-test scope expansion discovered** (plan §A.2 enumerated 4 test-file
touchpoints; actual grep reveals additional compile-breakable references in
`handlers_test.go`, `i18n_test.go`, `integration_test.go`, `restyle_test.go` —
all asserting presence of the statusline fieldset / segment checkboxes / preset
form values). These are expected mechanical refactor work per plan §A.2
"retire or refactor to match new surface" — handled in M3.

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — emitted by manager-develop upon M6 completion.>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates this section. Records the
sync commit SHA, CHANGELOG entry, README update (if any), and the
frontmatter status transition (in-progress → implemented).>_

---

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase — orchestrator-direct or manager-docs populates this
section with the Mx commit SHA, 4-phase close confirmation, and final
audit-ready YAML block.>_

---

## §F. Cross-References

- `spec.md` (canonical SSOT — §A overview, §B requirements, §C constraints,
  §E exclusions, §F risks)
- `plan.md` (M1-M6 milestone scope, anti-patterns, pre-flight checks)
- `acceptance.md` (30 ACs, traceability matrix, quality gates)
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` (era classification,
  §E.1-§E.5 marker semantics)
- `.claude/rules/moai/development/manager-develop-prompt-template.md §E`
  (E1-E7 self-verification deliverables — the run-phase evidence shape)

---

## Out of Scope (progress-phase)

The following are explicitly out of scope for this progress tracker:

### 1. Out of Scope — populated §E.2-§E.5 evidence

- The §E.2 Run-phase Evidence, §E.3 Run-phase Audit-Ready Signal, §E.4
  Sync-phase Audit-Ready Signal, and §E.5 Mx-phase Audit-Ready Signal sections
  are placeholder skeletons at plan-phase. Populating them belongs to
  manager-develop (§E.2/§E.3), manager-docs (§E.4), and orchestrator-direct
  Mx (§E.5) per the SPEC artifact ownership matrix. This progress.md MUST NOT
  carry populated evidence at plan-phase.

### 2. Out of Scope — non-SPEC runtime state

- This progress tracker does NOT record runtime-managed state
  (`.moai/state/*`, `.moai/logs/*`, `.moai/cache/*`). Those are owned by the
  runtime and hooks, not by the SPEC lifecycle.
