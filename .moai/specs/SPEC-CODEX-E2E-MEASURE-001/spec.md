---
id: SPEC-CODEX-E2E-MEASURE-001
title: "Codex-axis e2e measurement: test-surface execution inventory + named e2e journey gap report"
version: "0.1.0"
status: draft
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/cli,internal/codexwiring,internal/codexadapter,e2e"
lifecycle: spec-anchored
tags: "codex, e2e, measurement, testing, t462"
tier: M
related_specs: [SPEC-CODEX-WIRING-001, SPEC-CODEX-LAUNCHER-001, SPEC-CODEX-REVIEW-TARGET-001]
---

# SPEC-CODEX-E2E-MEASURE-001 — Codex-axis e2e measurement

## §A Context and Base

Operator directive: "proceed with all codex-related e2e". The repo's e2e surface is exactly one
journey script (`e2e/cli/tux3_journeys.sh`, journeys J1–J6) containing **zero** codex references
(`grep -ric codex e2e/cli/tux3_journeys.sh` → `0`). There is therefore no codex e2e journey to
"proceed with" — this card measures the codex-axis test surface that DOES exist (prong A) and
names the missing e2e journey areas as explicit gaps (prong B) so the lead can cut follow-up
cards without re-deriving them.

- **Base SHA (plan-phase measurement)**: `e9c6a8564` (local develop base; branch `WT-codex-e2e`,
  worktree `.claude/worktrees/t462`). Every run-phase measurement re-pins the SHA it was taken at.
- **Adjacent cards are NOT dependencies**: t451 (doctor codex-wiring silence repair) and t452
  (codex skill-axis wholesale absence) may land before or after this card. This card measures the
  state at ITS base and records which of them (if either) had landed at measurement time.
- **Evidence root**: `.moai/reports/t462/` (layout in §F).

### Plan-phase inventory baseline (both axes, measured at `e9c6a8564`)

| Axis | Command (verbatim) | Count |
|---|---|---|
| Filename | `find internal/cli -name '*codex*_test.go' \| wc -l` | 38 |
| Filename | `find internal/codexwiring -name '*_test.go' \| wc -l` | 6 |
| Dependency | `find internal/codexadapter -name '*_test.go' \| wc -l` (whole package missed by the filename glob) | 7 |
| Dependency | `grep -rlE 'runCodexReviewGate\|CodexAudit\|codex_task\|codex_setup\|codex_job\|codex-review-gate\|CodexWiring\|resolveCodexHomeDir\|codexCmd\|codexadapter\|codexwiring\|codexRunner' internal/ --include='*_test.go' \| grep -v -E '/[a-z0-9_]*codex[a-z0-9_]*_test\.go$' \| grep -v '^internal/codexwiring/\|^internal/codexadapter/' \| wc -l` | 22 |
| Lexicon | `grep -ril 'codex' internal/ --include='*_test.go'` minus filename axis minus codex packages minus dependency axis (the residual "delta", classified by per-file ref count) | 50 (27 behavioral ≥5 refs, 23 incidental 1–4 refs) |
| **Union** | filename ∪ dependency ∪ lexicon | **123** |

The lead's relayed "44 files" is confirmed as the **filename-glob lower bound** (38 + 6 = 44), not
the full surface. The dependency axis adds `internal/codexadapter` wholesale (7 files) and 22
symbol/import-referencing files spread across `internal/cli` (16), `internal/core/project` (2),
`internal/hook` (1), `internal/mcp` (1), `internal/template` (1), `internal/web` (1). The lexicon
axis (case-insensitive `codex`, not caught by either axis above) adds 50 further files — 27 with
≥5 refs (behavioral candidates: `internal/cli/mcp_convergence_participant_test.go` ×29,
`internal/template/agentemit/{agentemit,golden}_test.go` ×25/×22,
`internal/config/audit_models_test.go` ×22, `internal/cli/wizard/mcp_audit_test.go` ×17, …) and
23 with 1–4 refs (incidental mentions: config strings, registration lists). Full per-file list
with counts and classification: `.moai/reports/t462/inventory-baseline.md` §1c. Because the
lexicon delta reaches subpackages (`internal/cli/wizard`, `internal/template/agentemit`) and five
further packages (`internal/config`, `internal/sessionmsg`, `internal/settings`,
`internal/spec`, `internal/github/workflow`), the (A) execution surface uses RECURSIVE package
patterns (REQ-CEM-003/012/013).

### Machine-state ground truth (re-verified at plan time — drift vs relayed numbers)

- `~/.codex/skills/` contains exactly `.system` + `hatch-pet` — **zero moai skills deployed**.
  The wiring was never exercised, not broken (core premise CONFIRMED).
- Drift: the relayed "49 moai skills all `enabled = false`" was NOT reproduced. Measured:
  `grep -c '^\[skills\.moai' ~/.codex/config.toml` → `0` (no `[skills.*]` sections exist);
  `[plugins."moai-*@moai-cowork"]` blocks → 18, all `enabled = true`. The load-bearing fact
  (no deployed moai skills) holds either way; any check counting "disabled moai skills" must be
  re-derived against the actual TOML shape before it can mean anything.

## §B Definitions

- **Codex-axis test surface**: the 123-file three-axis union above (filename ∪ dependency ∪
  lexicon). Execution coverage is by recursive package pattern, so subpackages
  (`internal/cli/wizard`, `internal/template/agentemit`) are included by construction.
- **Integration-style check**: any command whose behavior depends on `CODEX_HOME` resolution
  (`resolveCodexHomeDir()`, `internal/cli/mcp_codex.go:1758`, reads env `CODEX_HOME` first,
  falls back to `~/.codex`).
- **Live-test gates (three distinct, verified at `e9c6a8564`)**:
  - `MOAI_CODEX_LIVE_PROBE` — OPT-IN (`codex_live_protocol_probe_test.go:31`); skipped unless
    set to `1` (+ optional `MOAI_CODEX_LIVE_BIN`).
  - `MOAI_SKIP_LIVE_CODEX` — OPT-OUT (`codex_review_gate_live_test.go:35`,
    `codex_review_target_live_test.go:37`); these tests RUN by default when a functional codex
    binary is on PATH — and one is (`/Users/goos/.local/bin/codex`). Default suite execution
    therefore sets `MOAI_SKIP_LIVE_CODEX=1` unless the lead approves quota spend at kickoff.
  - `MOAI_AUDIT_PIN_LIVE` — OPT-IN (`audit_pin_live_test.go:38`); skipped unless `1`.
  In all three cases a skip is UNOBSERVED, never a pass.
- **Positive control**: a deliberately broken wiring point in an ISOLATED fixture, shown to be
  DETECTED by the check under test, executed before any zero-count verdict is accepted
  (technique proven in card t457).

## §C Requirements (GEARS)

- **REQ-CEM-001** (Ubiquitous) — The run-phase report shall pin every measurement to the tree SHA
  it was taken at, and a measurement whose SHA cannot be named shall be recorded as a Gap, not a
  Claim.
- **REQ-CEM-002** (When) — When the run phase begins, the executor shall re-measure both
  inventory axes with the §A commands and record any drift against the plan-phase counts
  (44 filename / 29 dependency / 73 union) before executing any test.
- **REQ-CEM-003** (Where) — Where a package is wholly codex-owned, the executor shall execute
  `internal/codexwiring` and `internal/codexadapter` whole with
  `go test -count=1 -v ./internal/codexwiring/... ./internal/codexadapter/...`, recording exit
  code and swept (`=== RUN`) count per package.
- **REQ-CEM-004** (While) — While an integration-style check runs, the check shall run with
  `CODEX_HOME` pointing at a throwaway directory under `/tmp`, and the real `~/.codex` shall
  never be written (blocking precondition of every integration-style check; same discipline as
  CLAUDE.local.md §13).
- **REQ-CEM-005** (When) — When a zero-count or zero-failure verdict is about to be recorded, the executor shall first record a PASSING positive control in the same report — a deliberately broken fixture plus the observed detection of that break; a zero-verdict without a preceding positive control is uninterpreted.
- **REQ-CEM-006** (Ubiquitous) — The executor shall record the swept count alongside every
  test-run verdict, and a run whose swept count is zero shall be recorded as uninterpreted
  output, never as a pass.
- **REQ-CEM-007** (When) — When prong B is reported, each missing e2e journey area shall be
  named individually with the establishing command and its output, in
  `.moai/reports/t462/gap-inventory.md`.
- **REQ-CEM-008** (Unwanted) — The card shall not author e2e journeys, fix codex wiring, deploy
  moai skills to codex, or otherwise fill any gap it reports.
- **REQ-CEM-009** (While) — While adjacent cards t451/t452 land before or after this card, each report shall state the base SHA it was measured at and whether either card had landed at measurement time; neither card is a dependency of this SPEC.
- **REQ-CEM-010** (Unwanted) — The executor shall not run `go test ./...` locally, shall not run
  verification in parallel with another lane's verification, and shall not enable
  quota-spending live tests (`MOAI_CODEX_LIVE_PROBE=1`, `MOAI_AUDIT_PIN_LIVE=1`) as blocking
  steps.
- **REQ-CEM-011** (When) — When a commit is made at plan or run close, the executor shall stage
  by explicit pathspec on `WT-codex-e2e` in this worktree only, after re-reading
  `git rev-parse --short HEAD` and `git branch --show-current` in the same commit turn.
- **REQ-CEM-012** (Where) — Where the package is `internal/cli`, the executor shall execute it
  whole via the RECURSIVE pattern `go test -count=1 -timeout 1800s -v ./internal/cli/...`
  STANDALONE — codex-axis tests are scattered across 50+ files with heterogeneous names (§A)
  and a non-recursive `./internal/cli/` would silently skip `internal/cli/wizard` — recording
  exit code, duration, and swept count.
- **REQ-CEM-013** (Where) — Where the lexicon axis reaches packages beyond the wholly-codex and cli packages, the executor shall execute each of them whole with recursive `-count=1 -v` patterns — `internal/config`, `internal/core/project`, `internal/github/workflow`, `internal/hook`, `internal/mcp`, `internal/sessionmsg`, `internal/settings`, `internal/spec`, `internal/template/...`, `internal/web` — recording exit code and swept count per run.
- **REQ-CEM-014** (Where) — Where the opt-out live gate `MOAI_SKIP_LIVE_CODEX` exists, the executor shall set `MOAI_SKIP_LIVE_CODEX=1` for the default suite unless the lead explicitly approves quota spend at Implementation Kickoff Approval — live tests RUN by default when a functional codex binary is on PATH (one is, at `/Users/goos/.local/bin/codex`) — and shall record the resulting skips as inventory (a skip is UNOBSERVED, never a pass).

## §D Constraints

1. **Isolation (blocking)**: `CODEX_HOME` override to a `/tmp` throwaway for every
   integration-style check; real `~/.codex` read-only (hash-snapshot before/after proves
   non-mutation). The real `~/.codex` is never a fixture.
2. **Vacuous-green guard**: positive-control ordering gate per REQ-CEM-005. Ground truth makes
   this the card's biggest trap — moai skill deployment NEVER happened, so "all wiring checks
   pass" can be true because there is nothing to check. Zero-verdicts are meaningless until the
   check has been shown to detect a deliberate break.
3. **Load**: codex-axis packages only (recursive patterns, REQ-CEM-003/012/013); `internal/cli`
   whole runs STANDALONE with `-timeout 1800s`; serial execution; no local full suite (CI's job
   — 2026-08-15 load-413 precedent). Default suite sets `MOAI_SKIP_LIVE_CODEX=1` (REQ-CEM-014)
   — without it, opt-out live tests run against the on-PATH codex binary and spend quota.
4. **Scope**: measurement + naming only. Gap-filling is the lead's follow-up decision.
5. **Git hygiene**: `WT-codex-e2e` in this worktree only; explicit pathspec; HEAD/branch re-read
   before commit; no push, no integration window.

## §E Acceptance

Acceptance criteria (Given-When-Then, binary-testable): `acceptance.md` in this directory.

## §F Report layout (`.moai/reports/t462/`)

| File | Content | Owning phase |
|---|---|---|
| `inventory-baseline.md` | Plan-phase two-axis inventory (per-file list, exact commands, SHA `e9c6a8564`) | plan (authored with this SPEC) |
| `inventory-run.md` | Run-phase re-measurement + drift vs baseline | run |
| `positive-control.md` | Broken-fixture detection evidence (must precede any accepted zero-verdict) | run |
| `execution-log.md` | Prong A: per-package command, exit code, duration, swept count, SKIP inventory of live tests | run |
| `gap-inventory.md` | Prong B: named gaps, each with establishing command + output + SHA | run |

## §G Out of Scope

### Out of Scope — Journey authoring
- Authoring any e2e journey (codex or otherwise) — including adding J7+ to
  `e2e/cli/tux3_journeys.sh` or creating a new codex journey script. The lead decides follow-up
  cards after (A) results.

### Out of Scope — Wiring repair and skill deployment
- Repairing codex wiring, deploying moai skills to `~/.codex`, changing `codexwiring` /
  `codexadapter` / `codex_launcher.go` behavior. This card measures; it does not modify the
  measured surface.

### Out of Scope — Full-suite local verification
- `go test ./...` and any package outside the codex-axis union beyond what REQ-CEM-003 names.
  Full-suite verdicts belong to CI.

### Out of Scope — Adjacent-card work
- t451 (doctor codex-wiring silence) and t452 (codex skill-axis absence) are measured-state
  context only; this card neither blocks on nor implements them.

## HISTORY

| Date | Author | Change |
|---|---|---|
| 2026-09-03 | manager-spec | Plan-phase creation (card t462, lane-6, Factory Mode). Two-axis inventory baseline measured at `e9c6a8564`; machine-state ground truth re-verified with drift vs relayed numbers recorded in §A. |
| 2026-09-03 | manager-spec | Plan-audit iter-1 repair (D1–D11): lexicon axis added (union 73→123), recursive execution patterns (REQ-CEM-012/013), live-gate semantics corrected to three distinct gates + `MOAI_SKIP_LIVE_CODEX` default (REQ-CEM-014), compound REQs tightened (003/005/009/010), 17→16 count fix. |
