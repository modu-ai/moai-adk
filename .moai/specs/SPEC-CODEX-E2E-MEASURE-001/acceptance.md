---
id: SPEC-CODEX-E2E-MEASURE-001
title: "Acceptance criteria — codex-axis e2e measurement"
version: "0.1.0"
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
tier: M
---

# Acceptance — SPEC-CODEX-E2E-MEASURE-001

## §D AC Matrix (Given-When-Then)

### Prong A — execution and judgment

- **AC-CEM-001** (inventory re-measured) — Given the plan-phase baseline (44 filename / 29
  dependency / 73 union at `e9c6a8564`), when the run phase begins, then
  `.moai/reports/t462/inventory-run.md` exists, contains the four spec.md §A commands verbatim
  with their run-phase counts, the tree SHA each was taken at, and an explicit drift statement
  per axis. A missing drift statement fails this AC.

- **AC-CEM-002** (tier-1 packages) — Given `internal/codexwiring` and `internal/codexadapter`
  are wholly codex-owned, when they are executed whole with `go test -count=1`, then
  `execution-log.md` records for each: exit code, swept (`=== RUN`) count ≥ 1, and SHA. Swept
  count 0 recorded as pass fails this AC.

- **AC-CEM-003** (internal/cli standalone) — Given the measured 626–788s package duration, when
  `go test ./internal/cli/ -count=1 -timeout 1800s` runs STANDALONE, then `execution-log.md`
  records exit code, wall duration, swept count ≥ 1, and the output file path (redirect per the
  file-redirect contract). A run at the default 600s timeout, or concurrent with another
  lane's verification, fails this AC.

- **AC-CEM-004** (peripheral packages) — Given the dependency axis reaches 5 non-codex packages
  (`internal/core/project`, `internal/hook`, `internal/mcp`, `internal/template`,
  `internal/web`), when they are executed whole per plan M4 step 4, then each exit code and
  swept count is recorded in `execution-log.md`.

- **AC-CEM-005** (live-test SKIP inventory) — Given live tests skip unless
  `MOAI_CODEX_LIVE_PROBE=1`, when the default suite runs, then `execution-log.md` enumerates
  which live-gated tests skipped. Treating a skip as coverage, or enabling quota-spending live
  probes as a blocking step, fails this AC.

### Ordering gate — vacuous-green guard

- **AC-CEM-006** (positive control BEFORE zero-verdicts) — Given the ground truth that moai
  skill deployment to codex never happened (`~/.codex/skills/` = `.system` + `hatch-pet` only),
  when any zero-count or zero-failure verdict is recorded, then
  `.moai/reports/t462/positive-control.md` already contains (timestamp/ordering before the
  verdict) a deliberately broken wiring point in an ISOLATED `/tmp` fixture AND the observed
  detection of that break. Any accepted zero-verdict without a preceding positive-control PASS
  fails this AC — this is the card's hardest gate.

### Prong B — named gaps

- **AC-CEM-007** (gap inventory, named and established) — Given the plan-phase gap list
  (G1–G8, `inventory-baseline.md` §3), when prong B is reported, then
  `.moai/reports/t462/gap-inventory.md` exists and every gap entry carries: a specific name
  (surface + missing journey, e.g. "no e2e journey covers `moai codex status` six-row readout
  rendering"), the establishing command, its verbatim output, and the SHA. A vague
  "coverage is low" statement, or an unnamed gap count, fails this AC.

- **AC-CEM-008** (grep positive control) — Given `grep -ric codex e2e/cli/tux3_journeys.sh → 0`
  is quotable evidence, when that zero is cited, then the same report first shows the method
  detecting a known-present token (e.g. `grep -c doctor e2e/cli/tux3_journeys.sh` > 0) on the
  same file. Citing the codex 0 without the control fails this AC.

### Isolation and hygiene

- **AC-CEM-009** (CODEX_HOME isolation, blocking precondition) — Given
  `resolveCodexHomeDir()` reads `CODEX_HOME` before falling back to `~/.codex`, when any
  integration-style check runs, then it runs with `CODEX_HOME` pointed at a `/tmp` throwaway,
  AND `execution-log.md` records `shasum ~/.codex/config.toml` + `ls ~/.codex/skills/` before
  and after all such checks as identical. Any drift in the real-home snapshot fails this AC as
  a process defect to report, not to repair silently.

- **AC-CEM-010** (measurement pinning) — Given adjacent cards t451/t452 may land at any time,
  when any report file under `.moai/reports/t462/` is finalized, then it names the tree SHA it
  was measured at, and states whether t451/t452 had landed. A SHA-less measurement fails
  this AC.

- **AC-CEM-011** (scope boundary held) — Given the card's scope discipline, when the card
  closes, then no file outside `.moai/specs/SPEC-CODEX-E2E-MEASURE-001/` and
  `.moai/reports/t462/` was created or modified by this card (verifiable:
  `git diff --name-only e9c6a8564..HEAD` contains only those paths). Any journey authoring,
  wiring fix, or `~/.codex` mutation fails this AC.

- **AC-CEM-012** (git hygiene) — Given the shared-checkout rules, when a commit is made, then
  `git branch --show-current` output (`WT-codex-e2e`) and the pre-commit
  `git rev-parse --short HEAD` output appear in the same commit turn's evidence, staging was
  by explicit pathspec, and no push occurred.

## §D.1 Severity

- MUST-PASS: AC-CEM-001, 002, 003, 006, 007, 009, 010, 011 (the measurement, ordering-gate,
  isolation, and scope ACs — these ARE the card).
- SHOULD-PASS: AC-CEM-004, 005, 008, 012.
- Edge AC-CEM-011 note: run-phase `progress.md` §E.2 evidence lives inside the SPEC directory
  and does not violate the path bound.

## §D.2 Edge cases

- Base moves mid-measurement (t451/t452 merged into the lane's base): stop, re-pin SHA,
  re-measure inventory; never mix SHAs in one report file.
- codex binary absent on PATH: expected; live tests skip; record, do not install.
- `internal/cli` run exceeds 1800s: record the timeout as the verdict (fail), report to lead —
  do not silently raise the timeout.
- Real `~/.codex` drifts during the measurement window (another lane/user): the before/after
  snapshot mismatch is reported as attribution-ambiguous, not claimed as this card's fault or
  anyone else's.

## §D.3 Definition of Done

All five report files of spec.md §F exist with SHAs; every MUST-PASS AC above has an evidence
row (command + verbatim output + exit code + swept count where applicable); the §E.3 run-phase
audit-ready signal is populated in `progress.md`; commits follow §D.4 of the repo SPEC
conventions; the lead receives the gap list ready to cut follow-up cards.
