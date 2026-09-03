---
id: SPEC-CODEX-E2E-MEASURE-001
title: "Acceptance criteria — codex-axis e2e measurement"
version: "0.2.0"
status: in-progress
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
tier: M
---

# Acceptance — SPEC-CODEX-E2E-MEASURE-001

## §D AC Matrix (Given-When-Then)

Every MUST-PASS AC carries a **RED cell** — the observation that would falsify it. Pinned diff
base for scope ACs: `e9c6a8564` (constant `BASE_SHA`; re-pinned in the same report file if the
lane rebases, never silently).

### Prong A — execution and judgment

- **AC-CEM-001** (MUST, maps REQ-CEM-002) — inventory re-measured. Given the plan-phase baseline (44
  filename = 41 codex-named repo-wide + 6 codexwiring / 7+22 dependency / 50 lexicon delta
  (27+23) / union 126 at `e9c6a8564`), when the
  run phase begins, then `.moai/reports/t462/inventory-run.md` exists, contains the spec.md §A
  commands verbatim with their run-phase counts, the tree SHA each was taken at, and an explicit
  drift statement per axis. RED: a missing drift statement, or a SHA-less count.

- **AC-CEM-002** (MUST, maps REQ-CEM-003, REQ-CEM-006) — tier-1 packages, swept counts
  observable. Given
  `internal/codexwiring` and `internal/codexadapter` are wholly codex-owned, when they are
  executed whole with `go test -count=1 -v ./internal/codexwiring/... ./internal/codexadapter/...`,
  then `execution-log.md` records for each: exit code, swept (`=== RUN`) count ≥ 1 (emitted by
  `-v`), and SHA. RED: swept count 0 recorded as pass, or a run without `-v` (no `=== RUN`
  lines ⇒ count unsatisfiable).

- **AC-CEM-003** (MUST, maps REQ-CEM-006, REQ-CEM-012) — internal/cli recursive standalone.
  Given the measured 626–788s
  package duration and codex tests in `internal/cli/wizard`, when
  `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -timeout 1800s -v ./internal/cli/...` runs
  STANDALONE, then `execution-log.md` records exit code, wall duration, swept count ≥ 1, and
  the output file path. RED: non-recursive `./internal/cli/` (wizard silently skipped), default
  600s timeout, missing `-v`, or concurrency with another lane's verification.

- **AC-CEM-004** (SHOULD, maps REQ-CEM-013) — peripheral packages, recursive. Given the
  lexicon/dependency
  axes reach `internal/config`, `internal/core/project`, `internal/github/workflow`,
  `internal/hook`, `internal/mcp`, `internal/sessionmsg`, `internal/settings`, `internal/spec`,
  `internal/template/...` (recursive, covers `agentemit`), `internal/web`, when they are
  executed whole per plan M4 step 3, then each run's exit code and swept count ≥ 1 is recorded
  in `execution-log.md`. RED: a plain (non-recursive) `./internal/template/` pattern, or a
  peripheral run with no `=== RUN` evidence.

- **AC-CEM-005** (SHOULD, maps REQ-CEM-010, REQ-CEM-014) — live-gate inventory, corrected
  premise. Given
  THREE distinct gates (`MOAI_CODEX_LIVE_PROBE` opt-in; `MOAI_SKIP_LIVE_CODEX` OPT-OUT — those
  tests RUN by default because a functional codex binary IS on PATH at
  `/Users/goos/.local/bin/codex`; `MOAI_AUDIT_PIN_LIVE` opt-in), when the default suite runs
  with `MOAI_SKIP_LIVE_CODEX=1` (lead's kickoff default, M1-D2), then `execution-log.md`
  enumerates which live-gated tests skipped, under which gate. Treating a skip as coverage,
  running the default suite WITHOUT `MOAI_SKIP_LIVE_CODEX=1` (unapproved quota spend), or
  enabling opt-in live gates as a blocking step, fails this AC.

### Ordering gate — vacuous-green guard

- **AC-CEM-006** (MUST, maps REQ-CEM-005) — positive control BEFORE zero-verdicts. Given the
  ground truth
  that moai skill deployment to codex never happened (`~/.codex/skills/` = `.system` +
  `hatch-pet` only), when any zero-count or zero-failure verdict is recorded, then
  `.moai/reports/t462/positive-control.md` already contains (timestamp/ordering before the
  verdict) a deliberately broken wiring point in an ISOLATED `/tmp` fixture AND the observed
  detection of that break — hooks.json-first recipe per plan M3 (a stray hooks.json key the
  `codexadapter.ValidateConfig` whitelist must reject; `codex_readiness.go:116`,
  `doctor_codex.go:56`), executed via `go run ./cmd/moai` from the pinned tree (the installed
  `~/go/bin/moai` is an ancestor build and must not be used). RED: any accepted zero-verdict
  without a preceding positive-control PASS, or a control run against the ancestor binary —
  this is the card's hardest gate.

### Prong B — named gaps

- **AC-CEM-007** (MUST, maps REQ-CEM-007) — gap inventory, named and established. Given the
  plan-phase
  gap list (G1–G8, `inventory-baseline.md` §3), when prong B is reported, then
  `.moai/reports/t462/gap-inventory.md` exists and every gap entry carries: a specific name
  (surface + missing journey, e.g. "no e2e journey covers `moai codex status` six-row readout
  rendering"), the establishing command, its verbatim output, and the SHA. RED: a vague
  "coverage is low" statement, or an unnamed gap count.

- **AC-CEM-008** (SHOULD, maps REQ-CEM-005) — grep positive control. Given
  `grep -ric codex e2e/cli/tux3_journeys.sh → 0` is quotable evidence, when that zero is cited,
  then the same report first shows the method detecting a known-present token (e.g.
  `grep -c doctor e2e/cli/tux3_journeys.sh` > 0) on the same file. RED: citing the codex 0
  without the control.

### Isolation and hygiene

- **AC-CEM-009** (MUST, maps REQ-CEM-004) — CODEX_HOME isolation, blocking precondition. Given
  `resolveCodexHomeDir()` reads `CODEX_HOME` before falling back to `~/.codex`, when any
  integration-style check runs, then it runs with `CODEX_HOME` command-scoped to a `/tmp`
  throwaway, AND `execution-log.md` records a whole-tree snapshot (`git status --porcelain` +
  `git rev-parse --short HEAD`) plus `shasum ~/.codex/config.toml` + `ls ~/.codex/skills/`
  before and after all such checks as identical. RED: any integration check resolving
  `CODEX_HOME` to `~/.codex`, or drift in either the real-home or tree snapshot (reported as a
  process defect, not silently repaired).

- **AC-CEM-010** (MUST, maps REQ-CEM-001, REQ-CEM-009) — measurement pinning. Given adjacent
  cards
  t451/t452 may land at any time, when any report file under `.moai/reports/t462/` is
  finalized, then it names the tree SHA it was measured at, and states whether t451/t452 had
  landed. RED: a SHA-less measurement.

- **AC-CEM-011** (MUST, maps REQ-CEM-008) — scope boundary held. Given the card's scope
  discipline and
  the PINNED base `BASE_SHA=e9c6a8564` (re-pinned explicitly if the lane rebases), when the
  card closes, then `git diff --name-only ${BASE_SHA}..HEAD` contains only paths under
  `.moai/specs/SPEC-CODEX-E2E-MEASURE-001/` and `.moai/reports/t462/`. RED: any journey
  authoring, wiring fix, or `~/.codex` mutation. (Run-phase `progress.md` §E.2 evidence lives
  inside the SPEC directory and does not violate the path bound.)

- **AC-CEM-012** (MUST, maps REQ-CEM-011) — git hygiene. Given the shared-checkout rules,
  when a commit
  is made, then `git branch --show-current` output (`WT-codex-e2e`) and the pre-commit
  `git rev-parse --short HEAD` output appear in the same commit turn's evidence, staging was
  by explicit pathspec, and no push occurred. RED: any `git add -A`-class sweep stage, a commit
  on any other branch, or a push.

## §D.1 Severity

- MUST-PASS: AC-CEM-001, 002, 003, 006, 007, 009, 010, 011, **012** (measurement, swept-count,
  ordering-gate, isolation, scope, and git-hygiene ACs — these ARE the card).
- SHOULD-PASS: AC-CEM-004, 005, 008.

## §D.2 Edge cases

- Base moves mid-measurement (t451/t452 merged into the lane's base): stop, re-pin
  `BASE_SHA`, re-measure inventory; never mix SHAs in one report file.
- codex binary absent on PATH: opt-out live tests then skip by their own guard; still set
  `MOAI_SKIP_LIVE_CODEX=1` (belt-and-braces) and record.
- `internal/cli/...` run exceeds 1800s (wizard adds subpackage time): record the timeout as
  the verdict (fail), report to lead — do not silently raise the timeout.
- Real `~/.codex` or the tree drifts during the measurement window (another lane/user): the
  before/after snapshot mismatch is reported as attribution-ambiguous, not claimed as this
  card's fault or anyone else's.

## §D.3 Definition of Done

All five report files of spec.md §F exist with SHAs; every MUST-PASS AC above has an evidence
row (command + verbatim output + exit code + swept count where applicable); the §E.3 run-phase
audit-ready signal is populated in `progress.md`; commits follow §D.4 of the repo SPEC
conventions; the lead receives the gap list ready to cut follow-up cards.
