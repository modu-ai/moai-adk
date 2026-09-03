---
id: SPEC-CODEX-E2E-MEASURE-001
title: "Implementation plan — codex-axis e2e measurement"
version: "0.1.0"
status: completed
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
   + 22 symbol-referencing files in 6 packages; lexicon axis adds 50 more (27 behavioral, 23
   incidental) → three-axis union 126 (spec.md §A; filename axis is the REPO-WIDE glob
   `find internal -name '*codex*_test.go'` = 41 — 3 codex-named files live outside
   `internal/cli` — plus 6 `internal/codexwiring`).
2. **Relayed machine-state numbers drift.** "49 moai skills all enabled=false" NOT reproduced:
   0 `[skills.*]` sections exist; 18 `[plugins."moai-*"]` blocks all `enabled = true`;
   `~/.codex/skills/` = `.system` + `hatch-pet` only. Core premise (deployment never happened)
   CONFIRMED; the numeric detail is wrong and any check built on it would be vacuous or
   wrong-reason red.
3. **e2e surface**: one script (`tux3_journeys.sh`, J1–J6), zero codex matches — measured,
   needs positive control on the grep method before the 0 is quotable (see §G AP-2).
4. **internal/cli duration** 626–788s measured (relay; re-measure at run) — the 600s default
   timeout truncates; `-timeout 1800s` is mandatory. The pattern must be RECURSIVE
   (`./internal/cli/...`): non-recursive silently skips `internal/cli/wizard` (4+ codex-referencing
   files) and `internal/template/agentemit` would be missed the same way.
5. **Live gates are three, not one** (audit D2, verified): only
   `codex_live_protocol_probe_test.go` is opt-in via `MOAI_CODEX_LIVE_PROBE`;
   `codex_review_gate_live_test.go:35` and `codex_review_target_live_test.go:37` are OPT-OUT via
   `MOAI_SKIP_LIVE_CODEX` and RUN by default (functional codex IS on PATH at
   `/Users/goos/.local/bin/codex`); `audit_pin_live_test.go` is opt-in via
   `MOAI_AUDIT_PIN_LIVE`. Default suite sets `MOAI_SKIP_LIVE_CODEX=1` (kickoff item, M1-D2).

## §C Pre-flight (before any test execution)

- [ ] Re-read `git rev-parse --short HEAD` + `git branch --show-current` (expect `WT-codex-e2e`).
- [ ] **Whole-tree snapshot (D10)**: record `git status --porcelain` + `git rev-parse --short
      HEAD` into `execution-log.md` BEFORE and re-check AFTER all isolation checks — a tree that
      changed outside this card's pathspec during the window is reported, not repaired.
- [ ] **Export the isolation env once per invocation** (each Bash call is a fresh process):
      `CODEX_HOME=$TMPDIR/moai-t462-codex-home-$$` mkdir'd, and passed as a command-scoped
      assignment on EVERY integration-style invocation (never `~/.codex`).
- [ ] Read-only non-mutation baseline: `shasum ~/.codex/config.toml` + `ls ~/.codex/skills/`
      recorded into `execution-log.md` (re-checked after all integration-style checks; drift ⇒
      process defect, report immediately — attribution: another lane or this card).
- [ ] Confirm codex binary presence/absence on PATH (`command -v codex`) — recorded, not fixed.
- [ ] Record the installed `moai` binary provenance (`moai version`) — it is NOT this tree's
      build; the positive control must NOT use it (M3).

## §D Constraints (from spec.md §D — binding)

Isolation (CODEX_HOME override, blocking) · positive-control ordering gate · codex-axis-only
load, serial, recursive patterns, `internal/cli` standalone `-timeout 1800s` ·
`MOAI_SKIP_LIVE_CODEX=1` default · no `go test ./...` · no pushes · explicit-pathspec commits
on `WT-codex-e2e` only.

## §E Self-Verification (run-phase close)

Every verdict row carries: command, verbatim output (or bounded tail + file path), exit code,
swept count, tree SHA. Empty Gaps sections must be true. See `acceptance.md` §D matrix.

## §F Milestones

Ordered by decision-reversibility — the execution-surface decision first, mechanical steps last.

### M1 — Execution-surface decisions (HIGH, decide-first — kickoff items)
- **M1-D1 selector**: `internal/cli` runs WHOLE via the RECURSIVE pattern (standalone,
  `-timeout 1800s`) rather than a `-run` selector. Rationale: the three-axis sweep shows codex
  tests scattered across 50+ cli files (plus `internal/cli/wizard`) with heterogeneous names
  (`TestAC_C_009_*`, `TestReviewGate_*`, `TestHarnessCodex*`, `TestLaunchers_*`, …); no `-run`
  regex can be proven complete, and whole-package is the only selector whose completeness is
  trivially established. An override to a selector requires the selector's swept-count proof in
  the same run.
- **M1-D2 live-quota (audit D2)**: default `MOAI_SKIP_LIVE_CODEX=1` on every suite invocation —
  the card spends no quota and touches no live codex state; the skipped tests are recorded as
  inventory. The lead may instead approve one live run at kickoff (explicit opt-in, recorded).
  Opt-in gates (`MOAI_CODEX_LIVE_PROBE`, `MOAI_AUDIT_PIN_LIVE`) stay unset.

### M2 — Inventory re-measurement (HIGH)
Run the §A commands verbatim (all three axes + the lexicon-delta classification count); write
`inventory-run.md`; record drift vs 41+6 / 7+22 / 50 (27+23) / union 126. If the base moved past
`e9c6a8564` (t451/t452 landing), re-pin and re-measure — do not report stale counts.

### M3 — Positive control (HIGH, ordering gate)
Recipe (audit D9 corrected — validate the artifact the code actually validates):
1. **Binary pinning**: the control runs `go run ./cmd/moai ...` from THIS pinned tree (or a
   fresh `go build -o /tmp/moai-t462 ./cmd/moai` from it) — the installed `~/go/bin/moai` is an
   ancestor build (reports v3.1.3-era provenance, not `e9c6a8564`) and must NOT be used.
2. **hooks.json FIRST**: `codexadapter.ValidateConfig` whitelist-validates `.codex/hooks.json`
   (`codex_readiness.go:116`, `doctor_codex.go:56`); `config.toml` is only existence/shape
   checked (`doctor_codex.go:90-98`). So the default control injects a stray key into a
   fixture's `.codex/hooks.json` (the class Codex silently disables an entire hooks file over —
   t83 Finding D), then the check under test must REPORT the violation. A config.toml-based
   control (missing `[mcp_servers.moai]` table) is the secondary variant.
3. Run the check with `CODEX_HOME` command-scoped to the `/tmp` fixture; record the DETECTION
   (non-empty violation output / non-zero signal) in `positive-control.md`. Only after this
   PASS may zero-count verdicts be recorded (spec REQ-CEM-005).
4. Also positive-control the grep method for prong B (`grep -c doctor e2e/cli/tux3_journeys.sh`
   > 0 proves the method sees present tokens before `grep -c codex … → 0` is quotable).

### M4 — Prong A execution (HIGH)
Serial, in this order, never concurrent with another lane's verification. Every step carries
`-v` so `=== RUN` lines exist (swept-count ACs are unsatisfiable without them) and
`MOAI_SKIP_LIVE_CODEX=1` (M1-D2); integration-style steps additionally carry command-scoped
`CODEX_HOME=$TMPDIR/...`. Output redirected to file, bounded tail per the file-redirect
contract:
1. `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -v ./internal/codexwiring/... ./internal/codexadapter/...`
   (exit code + swept count per package)
2. `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -timeout 1800s -v ./internal/cli/...` STANDALONE
   (exit code + duration + swept count; RECURSIVE — `./internal/cli/` plain would skip `wizard`)
3. Peripheral packages, whole, recursive, in two runs:
   a. `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -v ./internal/config/... ./internal/core/project/... ./internal/github/workflow/... ./internal/hook/... ./internal/mcp/... ./internal/sessionmsg/... ./internal/settings/... ./internal/spec/... ./internal/web/...`
   b. `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -v ./internal/template/...` (separate —
      embed-heavy; recursive so `internal/template/agentemit` runs)
4. SKIP inventory: enumerate skipped live-gated tests across all runs
   (`MOAI_SKIP_LIVE_CODEX` opt-outs; `MOAI_CODEX_LIVE_PROBE` / `MOAI_AUDIT_PIN_LIVE` opt-ins).
Optional (non-blocking, only on lead's explicit kickoff approval): one live run.

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
- **AP-6 Ancestor binary**: running the control or journeys against the installed
  `~/go/bin/moai` (an earlier build) and attributing the result to this tree.
- **AP-7 Non-recursive pattern**: `./internal/cli/` or `./internal/template/` (no `/...`)
  silently skipping `wizard` / `agentemit` subpackages while reporting full coverage.

## §H Cross-References

- `spec.md` §A (inventory commands, three axes), §C (REQ-CEM-001..014), §F (report layout)
- `acceptance.md` — AC matrix (Given-When-Then)
- `.moai/reports/t462/inventory-baseline.md` — plan-phase per-file inventory + named gaps
- CLAUDE.local.md §13 (GLM integration isolation — same discipline applied to CODEX_HOME)
- CLAUDE.local.md §6 / memory: no local full-suite runs; `internal/cli` timeout floor
