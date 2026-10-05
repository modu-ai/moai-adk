# SPEC-CLI-COVER-SUPPRESS-001 — progress.md

> Run-phase record (card t481, lane-8). Investigation card: measurement and report only —
> no code or test changes were made (REQ-CSS-006 report-only fence held).

## §E.1 Scope Recap

Three checks from the lead's dispatch: (1) is the failing-package `-cover` suppression real;
(2) does coverage for `internal/cli` actually appear after the t477 landing; (3) where does the
number stand against `CLAUDE.local.md` §6 targets (85% package / 90% critical cli/template/hook).
The dispatch's precondition ("t477 must land first") collapsed before the lane started: t477 was
already on local `develop` @ `8b391bc8c`, so checks (2) and (3) ran in the same session as (1).

## §E.2 Verification Matrix

| AC | Verdict | Evidence |
|---|---|---|
| AC-CSS-001 (mechanism verdict) | PASS — the suppression claim is REFUTED on go1.26.4 | `go -C .moai/cache/t481-mech test -count=1 -cover ./...` → exit 1; the FAILING package prints `coverage: 100.0% of statements` on its own line above the package FAIL line; with `-coverprofile`, the failing package's profile lines ARE merged (3-line profile: mode + `t481mech/bad` + `t481mech/ok`). Raw outputs: `.moai/reports/t481/mechanism-plain.txt`, `.moai/reports/t481/mechanism-coverprofile.txt`, `.moai/reports/t481/mechanism-coverprofile-data.txt`. Process note: an initial grep read 0 profile lines because the pattern carried a stray hyphen (`t481-mech` vs module path `t481mech`); re-read with the correct pattern before any verdict was drawn |
| AC-CSS-002 (post-t477 sweep) | PASS | `go test -count=1 -cover -timeout 900s ./internal/cli/...` (kanban env vars unset in the same invocation) on tree `8b391bc8c` → **exit 0**, 17/17 packages `ok`; main package line: `ok  github.com/modu-ai/moai-adk/internal/cli  438.373s  coverage: 80.6% of statements`. Raw output: `.moai/reports/t481/cli-cover-post-t477-sweep.txt`; exit code read directly from the background-task output (`exit=0`) |
| AC-CSS-003 (target judgment) | PASS as a REPORT — the numbers are BELOW target | main `internal/cli` **80.6%** < 85% (−4.4pp) and < 90% critical (−9.4pp); also below the 85% floor: `internal/cli/harness` 80.9%. Full 17-row table with both readings of the critical-90% rule: `.moai/reports/t481/verdict.md`. NO coverage work was performed in this card |
| AC-CSS-004 (verbatim correction) | PASS | t237 verbatim (tree `32319621b`, progress.md:103): "…and `go test -cover` prints no coverage line for a failing package." recorded next to the measured counter-evidence (AC-CSS-001). The likelier cause of their "Unmeasured" cell is aggregation reading only `ok … coverage: X%` lines — their own subpackage cells quote exactly that glued shape, while a failing package's number sits on a separate line. Their AC-PVM-010 row cites the `FAIL … 462.872s` line only; absence of the number from their summary was real, suppression of the output was not |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: complete
run_complete_at: 2026-09-04
run_baseline_head: 8b391bc8c   # fast-forward absorb of local develop (contains t477 b2f98f6aa); sweep measured on this tree with zero card commits present
cycle_type: investigation (report-only; no implementation)
ac_matrix: "§E.2 — 4 PASS; AC-CSS-003's substance is a BELOW-TARGET finding, reported not repaired"
key_numbers:
  main_internal_cli: "80.6% (438.373s, exit 0)"
  packages_total: 17
  packages_ok: 17
  below_85_floor: ["internal/cli 80.6", "internal/cli/harness 80.9"]
  critical_90_note: "per-package application across internal/cli/* additionally flags agentlint 86.7, preference 85.8, update 88.9, worktree 87.1; family-root-only application flags main 80.6 alone — both readings recorded, adjudication belongs to the lead/operator"
gaps:
  - "go-version history of the suppression behavior NOT investigated (current-toolchain verdict only; t237 ran the same go1.26.4 on the same machine on the same day, so the refutation covers their run too)"
  - "no CI/clean-environment re-measurement — single local darwin/arm64 run"
  - "t237's exact aggregation method is not recorded in their artifacts; ok-line aggregation is a hypothesis consistent with their verbatim record, not a measurement"
  - "other cover flag combinations (-coverpkg etc.) unmeasured"
residual_risk:
  - "single-run sweep — environmental variance of the 438s root-package run not bounded"
  - "the coverage shortfall itself is out of this card's scope — handed to a separate axis"
```

## Notes

- The plan→run gate question for a report-only investigation whose measurement commands were
  ordered verbatim by the dispatch is surfaced to the lead in the completion report — the lane
  does not self-adjudicate the gate.
- Worktree `.claude/worktrees/t481` is kept (branch unpushed per dispatch); `.claude/worktrees/t474`
  untouched.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
sync_complete_at: 2026-09-04
sync_commit_sha: 475fda8a1   # backfilled — the sync commit itself (cannot cite its own SHA)
carried_in_sync_commit:
  - spec.md          # in-progress -> completed frontmatter transition (status only; body untouched)
  - progress.md      # this §E.4 signal
untouched_by_sync: "plan.md / acceptance.md bodies (plan-phase content); all source trees — report-only card, zero source changes"
changelog_decision: "no CHANGELOG entry — report-only investigation with zero user-facing changes; the deliverable is .moai/reports/t481/verdict.md"
ac_matrix: "§E.2 (4 PASS) — unchanged by sync"
gaps: "none new at sync — the §E.3 gaps list stands"
```

