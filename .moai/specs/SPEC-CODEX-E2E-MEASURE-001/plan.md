---
id: SPEC-CODEX-E2E-MEASURE-001
title: "Implementation plan — codex-axis e2e measurement"
version: "0.1.0"
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
tier: M
---

# Plan — SPEC-CODEX-E2E-MEASURE-001

## §A Context

Card t462 (lane-6, Factory Mode). Base `e9c6a8564`, branch `WT-codex-e2e`, worktree
`.claude/worktrees/t462`. Two prongs: (A) execute + judge the codex-axis test surface,
(B) name the e2e journey gaps. Journey authoring is OUT of scope. Full baseline:
`spec.md` §A and `.moai/reports/t462/inventory-baseline.md`.

## §B Known Issues (plan-phase findings)

1. **The lead's 44 is a lower bound.** Filename glob = 44 (38 `internal/cli` + 6
   `internal/codexwiring`); dependency axis adds `internal/codexadapter` (7, whole package)
   + 22 symbol-referencing files in 6 packages → union 73.
2. **Relayed machine-state numbers drift.** "49 moai skills all enabled=false" NOT reproduced:
   0 `[skills.*]` sections exist; 18 `[plugins."moai-*"]` blocks all `enabled = true`;
   `~/.codex/skills/` = `.system` + `hatch-pet` only. Core premise (deployment never happened)
   CONFIRMED; the numeric detail is wrong and any check built on it would be vacuous or
   wrong-reason red.
3. **e2e surface**: one script (`tux3_journeys.sh`, J1–J6), zero codex matches — measured,
   needs positive control on the grep method before the 0 is quotable (see §G AP-2).
4. **internal/cli duration** 626–788s measured (relay; re-measure at run) — the 600s default
   timeout truncates; `-timeout 1800s` is mandatory.
5. **Live tests skip by default** (`MOAI_CODEX_LIVE_PROBE=1` gate) — a green suite with them
   skipped is expected and must be recorded as such, not as coverage.

## §C Pre-flight (before any test execution)

- [ ] Re-read `git rev-parse --short HEAD` + `git branch --show-current` (expect `WT-codex-e2e`).
- [ ] Create throwaway `CODEX_HOME=$TMPDIR/moai-t462-codex-home-$$` (mkdir; never `~/.codex`).
- [ ] Read-only non-mutation baseline: `shasum ~/.codex/config.toml` + `ls ~/.codex/skills/`
      recorded into `execution-log.md` (re-checked after all integration-style checks; drift ⇒
      process defect, report immediately — attribution: another lane or this card).
- [ ] Confirm codex binary presence/absence on PATH (`command -v codex`) — recorded, not fixed.

## §D Constraints (from spec.md §D — binding)

Isolation (CODEX_HOME override, blocking) · positive-control ordering gate · codex-axis-only
load, serial, `internal/cli` standalone `-timeout 1800s` · no `go test ./...` · no pushes ·
explicit-pathspec commits on `WT-codex-e2e` only.

## §E Self-Verification (run-phase close)

Every verdict row carries: command, verbatim output (or bounded tail + file path), exit code,
swept count, tree SHA. Empty Gaps sections must be true. See `acceptance.md` §D matrix.

## §F Milestones

Ordered by decision-reversibility — the execution-surface decision first, mechanical steps last.

### M1 — Execution-surface decision confirmation (HIGH, decide-first)
The plan-phase decision: **`internal/cli` runs WHOLE (standalone, `-timeout 1800s`) rather than
via a `-run` selector.** Rationale: the dependency sweep shows codex-axis tests scattered across
50+ cli files with heterogeneous test names (`TestAC_C_009_*`, `TestReviewGate_*`,
`TestHarnessCodex*`, `TestLaunchers_*`, …); no `-run` regex can be proven complete, and the
swept-count discipline makes whole-package the only selector whose completeness is trivially
established. The lead confirms or overrides at Implementation Kickoff Approval; an override to a
selector requires the selector's swept-count proof in the same run.

### M2 — Inventory re-measurement (HIGH)
Run the four §A commands verbatim; write `inventory-run.md`; record drift vs 44/29/73. If the
base moved past `e9c6a8564` (t451/t452 landing), re-pin and re-measure — do not report stale
counts.

### M3 — Positive control (HIGH, ordering gate)
In the `/tmp` fixture: inject a deliberate wiring break (e.g. a `.codex/config.toml` with a key
the `codexadapter` `ValidateConfig` whitelist must reject, or a broken hooks.json stray key —
the class Codex silently disables an entire hooks file over, per `internal/codexwiring`
package doc / t83 Finding D), run the relevant check (e.g. `moai codex status` / the wiring
inspection path) with `CODEX_HOME` overridden, and record it DETECTING the break. Only after
this PASS may zero-count verdicts be recorded (spec REQ-CEM-005). Also positive-control the
grep method for prong B (`grep -c doctor e2e/cli/tux3_journeys.sh` > 0 proves the method sees
present tokens before `grep -c codex … → 0` is quotable).

### M4 — Prong A execution (HIGH)
Serial, in this order, never concurrent with another lane's verification:
1. `go test ./internal/codexwiring/... -count=1` (exit code + swept count)
2. `go test ./internal/codexadapter/... -count=1` (exit code + swept count)
3. `go test ./internal/cli/ -count=1 -timeout 1800s -v` STANDALONE (exit code + duration +
   swept count; redirect to file, bounded tail per the file-redirect contract)
4. Peripheral dependency-axis packages, whole, cheap:
   `go test ./internal/core/project/ ./internal/hook/ ./internal/mcp/ ./internal/web/ -count=1`
   and `go test ./internal/template/ -count=1` (separate — embed-heavy)
5. SKIP inventory: enumerate live-gated tests skipped in run 3 (`MOAI_CODEX_LIVE_PROBE` gate).
Optional (non-blocking, only if lead pre-approves quota): one opt-in live probe run.

### M5 — Prong B gap inventory + report assembly (MEDIUM)
Author `gap-inventory.md`: each named gap (G1–G8 baseline in
`.moai/reports/t462/inventory-baseline.md` §3) re-established at run time with command + output
+ SHA. Assemble `execution-log.md`. No gap is filled.

### M6 — Commit (LOW)
Single plan/run-phase commit(s) on `WT-codex-e2e`, explicit pathspec
(`.moai/specs/SPEC-CODEX-E2E-MEASURE-001/ .moai/reports/t462/`), after re-reading
HEAD + branch in the same turn. Subject:
`feat(SPEC-CODEX-E2E-MEASURE-001): plan-phase artifacts (Tier M, 3 artifacts)` (plan close) /
`test(SPEC-CODEX-E2E-MEASURE-001): codex-axis measurement evidence` (run close). No push.

## §G Anti-Patterns

- **AP-1 Vacuous zero**: recording "0 failures" without the M3 positive control PASS on record
  first. The skills axis never had anything deployed — checks pass because there is nothing to
  check.
- **AP-2 Uncontrolled grep**: quoting `grep -ric codex e2e/ → 0` without first showing the same
  grep sees a known-present token (empty-sweep discipline).
- **AP-3 Relay reuse**: carrying the lead's 626–788s or 44-file numbers into the run report
  without this run's own measurement (a relayed value is not your measurement).
- **AP-4 Real-home fixture**: any integration check whose `CODEX_HOME` resolves to `~/.codex`.
- **AP-5 Scope creep**: "while measuring, I fixed the wiring" — forbidden (spec REQ-CEM-008).

## §H Cross-References

- `spec.md` §A (inventory commands), §C (REQ-CEM-001..011), §F (report layout)
- `acceptance.md` — AC matrix (Given-When-Then)
- `.moai/reports/t462/inventory-baseline.md` — plan-phase per-file inventory + named gaps
- CLAUDE.local.md §13 (GLM integration isolation — same discipline applied to CODEX_HOME)
- CLAUDE.local.md §6 / memory: no local full-suite runs; `internal/cli` timeout floor
