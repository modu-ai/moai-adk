# plan.md — SPEC-CODEX-AGENTS-SURFACE-001

Card t505 · Tier S · documentation-of-judgment (manifest-comment-only; no TDD red-green on behavior — the AC flips are grep-adoption flips, the guards are inherited)

## §A Milestones (serial)

### M1 — Manifest edit (the only source change)

Target: `internal/template/agentemit/agents-codex.yaml`.

1. Refresh the `fields.model` comment block (:28-32), the class-`model` `omit` rationale (:190-194), and the `model-pin-manager-git` documented drop (:214-218) with the 0.153.4 evidence set of spec.md §B.1 (precedence-winning over parent + `[agents]` default; open-bug #32587; model-id churn #42874; P-03 accepted-but-wrong stays attributed to 0.147.0).
2. Refresh the skill-loader class rationale (:117-139) with spec.md §B.2: 0.153.4 tag-pinned `SkillConfig = {path?, name?, enabled}` override-not-grant + `enabled` required; t505 P4–P7 measured silence; the 0.152.1 measured non-read and the "Re-probe when an observable agent-role surface exists" clause survive unchanged.
3. Add the `[agents]` judgment record: a `documented_drops` entry `id: codex-global-agents-table` + an adjacent YAML comment block reproducing the spec.md §B.3 type map line-per-key, stamped `codex-cli 0.153.4`, carrying the dual-source hazard, the rc=1 strict-parse hazard, the auto-discovery redundancy, and the explicit t494 A1 overturn. Follow the file's existing `>-` folded-scalar style.

Flips: AC-CAS-001, AC-CAS-002, AC-CAS-003. Preserves: AC-CAS-004.

### M2 — Regeneration + zero-diff proof

1. `make agents-emit` (REQ-CSL-008 obligation — regeneration stays behind the explicit verb; Makefile:38).
2. `git diff --stat -- internal/template/templates/.codex` → empty stdout, rc 0. Control: the same command form against `internal/template/agentemit/agents-codex.yaml` is non-empty after M1 (diff-visibility catch-all).
3. `make agents-emit-check` → rc 0 (read-only drift gate; Makefile:47-49).
4. Optional mutation check (t452 AC-CSL-012 precedent): touch one committed TOML, observe AC-CAS-005 flip red + `TestCodexAgentsDeployFixture` red, revert, confirm byte identity.

Closes: AC-CAS-005.

### M3 — Test/gate verification

`go test ./internal/template/... -count=1` → green. Coverage of the inherited guards: manifest parse (AC-013 fail-closed, manifest.go:112-120), golden AC-009 no-model (golden_test.go:170,205-207), `TestEmitAllOmitsModel` (agentemit_test.go:264-279), `TestCodexAgentsDeployFixture` (codex_agents_deploy_test.go:60-62,74-88). Lane-local scope only: no `go test ./...`, no go vet / golangci-lint / cross-build needed — zero Go changes (state this explicitly in the §E report).

Closes: AC-CAS-006.

### M4 — Evidence/verdict records

1. `.moai/reports/t505/verdict.md` — 5-section evidence-bearing report (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) citing each AC's command + verbatim output + tree SHA.
2. progress.md §E.2/§E.3 run-phase evidence + audit-ready signal; re-run the mutant probes and attach their verdicts.
3. Per-site AC-CAS-003 record (plan-audit F-2): the verdict must record the AC-CAS-003 grep count contributed by each of the three required refresh sites separately — model rationale, skill-loader rationale, [agents] judgment record — so site placement, not just the ≥3 total, is evidenced.

## §B Delta markers + PRESERVE

- `[MODIFY] internal/template/agentemit/agents-codex.yaml` — the only source edit.
- `[EXISTING-UNTOUCHED]` `internal/template/templates/.codex/**` (must stay byte-identical), every `*.go` file, `internal/codexwiring/**`, `internal/cli/update_codex_wiring.go`, `internal/template/templates/.claude/agents/moai/*.md` (the neutral layer).

PRESERVE list:
- The 11 committed TOMLs byte-identical (guard: `TestCodexAgentsDeployFixture`; regenerate-not-edit).
- Top-level `codex_measured_version: "0.147.0"` (AC-CAS-004; AC-CSL-009 decision — raising it claims unmeasured axis coverage).
- The skill-loader rationale's 0.152.1 stamp and re-probe clause.
- R-011/AC-009 machinery: `fields.model.emit: false`, class-`model` `omit`, `model-pin-manager-git` drop — retained, only rationales refreshed.

## §C Risk analysis

| Risk | Mitigation |
|---|---|
| A YAML comment edit breaks manifest parsing | Parse coverage runs in every emit test (AC-013 fail-closed, manifest.go:112-120); M3 runs the package suite |
| REQ-CSL-008 forgotten (edit without regeneration) | `make build` gates on `agents-emit-check` (Makefile:34) and drift aborts the build (Makefile:47-49); M2 regenerates explicitly |
| Accidental direct TOML edit (violates regenerate-not-edit) | `TestCodexAgentsDeployFixture` byte identity (codex_agents_deploy_test.go:74-88) + M2 zero-diff proof |
| Top-level stamp wrongly raised to 0.153.4 | AC-CAS-004 preservation cell + the AC-CSL-009 recorded decision cited in REQ-CAS-005 |
| A1 overturned silently (reversal without evidence) | AC-CAS-002 forces the type map; spec.md §B.4 carries the full per-key overturn; REQ-CAS-003 names the overturn in the record itself |
| Folded-scalar (`>-`) formatting mistakes in new rationales | Clone the existing entry style; parse tests catch structural breakage in M3 |
| AC-CAS-004 passes vacuously (file untouched) | Non-vacuity is co-observation with AC-CAS-001/003 (0 → ≥1 flips on the same file), stated in both cells |

## §D MX tag plan

**Zero new @MX tags** — recorded, not defaulted:

- The edit surface is a YAML build input, not an @MX-tagged source language.
- The deliverable itself IS rationale comments (evidence documentation), so an @MX:NOTE would duplicate the record it sits beside.
- The manifest header comment (:7-11) already carries the measured-version hygiene rule.
- No behavior change → no ANCHOR/WARN/DEBT/TODO trigger (no new exported function, no fan_in change, no dangerous pattern).
- The MX report at sync states: 0 added / 0 removed / 0 updated.

## §E AC ↔ milestone mapping

| AC | Kind | RED-now | Flipped by |
|---|---|---|---|
| AC-CAS-001 | release-blocking flip | 0 / rc 1 @ 0b1e27877 | M1 (id present → ≥1) |
| AC-CAS-002 | release-blocking flip (depth guard) | 0 / rc 1 @ 0b1e27877 | M1 (type map → ≥3 lines) |
| AC-CAS-003 | release-blocking flip | 0 / rc 1 @ 0b1e27877 | M1 (0.153.4 cites → ≥3) |
| AC-CAS-004 | preservation guard | 1 / rc 0 (stays 1) | co-observed with 001/003 |
| AC-CAS-005 | regression guard (mutants C/D) | inherited green | M2 (regen + diff-empty + control) |
| AC-CAS-006 | regression guard | inherited green | M3 (package suite + agents-emit-check) |

## §F Constraints (forbidden)

- FORBIDDEN: editing any file under `internal/template/templates/.codex/` by hand.
- FORBIDDEN: any `.go` edit (writer, validator, tests) — zero-Go-change SPEC.
- FORBIDDEN: raising `codex_measured_version` above `"0.147.0"`.
- FORBIDDEN: touching `internal/codexwiring/`, `internal/cli/update_codex_wiring.go`, or the neutral `.md` layer.
- FORBIDDEN: `go test ./...` locally (lane-local verification scope rule); full-suite judgment belongs to CI.
- REQUIRED: Conventional Commits per milestone (`feat(SPEC-CODEX-AGENTS-SURFACE-001): M{n} ...`), card id t505 in every commit message.
