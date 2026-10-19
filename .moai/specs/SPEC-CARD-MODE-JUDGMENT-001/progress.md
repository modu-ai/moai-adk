# progress.md — SPEC-CARD-MODE-JUDGMENT-001

> Phase progress record. Only `§E.1` is filled by the plan phase (manager-spec); `§E.2` and `§E.3` belong to the run phase (manager-develop), `§E.4` to the sync phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-19T02:29:45Z
tier: L
artifacts: spec.md (REQ 16) · plan.md (M1-M5) · acceptance.md (AC 17: R 10, G 6, O 1) · design.md (D-1..D-10) · research.md · decision-index.md (Q1-Q6) · progress.md
measured_tree: 74ad197380a0711cdb9816d4e37c94d29af730ea (branch WT-fail-safe-serial, base 74ad197380, clean, unpushed)
open_clarifications: 1 — decision-index Q3 (FOUNDER, product-level, empty verdict): blocks the autonomous Kickoff and gates REQ-MJ-007 / AC-MJ-006 / run-phase M3 step 2 only. Q1, Q2, Q4, Q5 carry `DEFAULT-APPLIED`; Q6 is POLICY-COVERED.
spec_lint: `moai spec lint --strict .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/spec.md` → `✓ No findings — all SPEC documents are valid`, exit 0 (installed build moai_cp/20261007_195927, which predates this tree's HEAD; the lint also scans plan.md and flagged two unanchored `-run` patterns first, repaired)
spec_audit: `mcp__moai__spec_audit` (filter SPEC-CARD-MODE-JUDGMENT-001, project_root = this worktree) before this file existed: one INFO `EraAutoDetected` (H-1, progress.md absent). It is expected to read H-5 (modern) now that `§E.2`-`§E.4` headings and `phase: "v3.2.0 target"` exist; not re-run.

### Premise record (card text and premise-check.md P1-P12, re-measured; commands and outputs in research.md §1 and acceptance.md §2)

| # | Premise | Result |
|---|---|---|
| P1 | template = deployed `factory-dispatch-cards.md` | holds (`cmp` exit 0). Seven of the eight target pairs are byte-identical; `factory-dispatch.md` is **not** (`cmp` exit 1, one declared live-only line, `TestAutoPickMirrorParity`) — premise P1 is false for that file and it is handled as AC-MJ-017 |
| P2 | no mode-judgment step | holds |
| P3 | `--classification-file` documented only in `gtd.md` | holds; additionally `manager-todo.md` and `manager-lead.md` describe **no** issuance step (Q1) |
| P4 | candidate issuance documents | resolved: only `gtd.md:52` and `factory-dispatch.md:30` describe issuing; `doc-generation.md:348` is a third, bare, out of scope by default (Q2) |
| P5 | `gtd.md:56` states the code default | holds; kept (REQ-MJ-015) |
| P6 | `todo_classify.go:41` | notice constant; default decider at line 36 |
| P7 | three renewal sites contradict a672 | holds and widens: five document sites (`factory-dispatch-core.md:28`, `factory-dispatch.md:119` and `:122`, `auto-semantics.md` §5.1, `factory-dispatch-mechanics.md:26`) plus `moai-lane-watchdog/SKILL.md` §0a, plus the Go SessionStart notice (out of scope, Q4) |
| P8 | free-lane wording | holds; `…-detail.md:133` conflicts, `…-mechanics.md:193` is the separate live-session notion |
| P9 | authorities 0c51, a672, 02f4 | read with `moai decision read`; quoted in spec.md §2 |
| P10 | code is t1737's | holds; `SPEC-FACTORY-DEFAULT-MODE-001` is not in this tree |
| P11 | Template-First build chain | read; no Go run; `make build` not required for a documents-only change |
| P12 | `spec_id` set on the card | holds |

### What this plan measured beyond the premise check (summary; detail in research.md)

- Mirror and sync mechanism: Template-First; `*-check` make targets are read-only drift detectors; the real parity guards are `TestRuleTemplateMirrorDrift`, `TestAutoPickMirrorParity`, `TestReviewOwnership_CardReviewDoctrine`, `TestTodoSkillDocumentsClassification`. The `.moai/policies/**` and `.moai/workflows/gtd.md` copies are stale deploy-time projections, not guards.
- Pinning tests that constrain the wording: the binding ledger (rows ALB-0254, -0265, -0268, -0276, -0304 touch the target lines), the renew-interval token test, `TestLaneRecheckDoctrine`, `TestLaneWatchdogRenewsFirst` / `…RecordsRefusalAndMeasures`, the card-id/internal-content leak tests, the AC-ALB-021(2) fragment rule.
- Budget: `factory-dispatch-core.md` measures 9,999 UTF-16 units against 10,000 — one unit of headroom; the marker-form rewrite is +9 and needs an 8-unit offset.
- Catalog: skill-directory hashes in `internal/template/catalog.yaml` change for `moai` (it contains `workflows/gtd.md`) and `moai-lane-watchdog`.
- Sweep for "no statement relies on the serial default when issuing": violating sentence V defined once (research.md §7); fires on exactly `gtd.md:52`; positive control observed on the base commit.

### Gaps (not observed in this phase)

- **G-1** The Tier question that `spec-assembly.md` poses to the operator could not be asked from a subagent; Tier L is the plan author's classification from the files-affected column (decision-index Q6, design.md D-1).
- **G-2** No Go command ran. Every statement about what a guard test pins comes from reading its source; none was executed. AC-MJ-014 takes the undecidable disposition (regression-guard, not recorded as a pass).
- **G-3** The binding ledger tests, the 10,000-unit core gate and the 150,000 subtotal were not run; the unit counts are a Python UTF-16 measurement of the files as stored, while the Go gates render the template first.
- **G-4** The worktree-isolation guard refused several read-only invocations during this phase: a loop that mixed `cmp` with a piped `git ls-files`, two `sed`/`grep` calls whose path was a shell variable, and three `git diff -G` calls whose pattern carried `{n}` quantifiers ("a form too complex to verify"). Each was re-issued as a plain command (one `git`, literal paths) or with the quantifier written out without braces. Recorded here per `verification-claim-integrity.md` §3.1; no result was inferred from a refused command. The same refusal will apply to a run-phase author who writes the AC-MJ-013 pattern with `{n}`.
- **G-5** Every `moai` CLI observation (`decision read`, `todo next --help`, `factory status`, `spec lint`) was judged by the installed build `moai_cp/20261007_195927`, which predates HEAD; `spec lint` in particular was not run from a build made from this tree.
- **G-6** "Class C `--spec` cannot be added after assign" is board 0c51's statement; this plan saw `--spec` on `moai todo next` and not on `add`, and did not read the assign path in code.
- **G-7** Which core lines are safe to trim for the 8-unit offset was measured only against the binding ledger (uncovered lines 3, 26, 28, 48, 54, 62, 64, 84); other pinning tests (for example `role_entry_points_test.go`) were not read.
- **G-8** The Go SessionStart notice was read from source (`session_start_factory_i18n.go`), not observed at a live join.
- **G-9** `grep -c` printed per-file lines in an order different from the argument order; the acceptance expectations are therefore stated as sets.
- **G-10** `TestAutoPickMirrorParity` and `TestRuleTemplateMirrorDrift` were read in the regions that name the target files, not end to end.

### Residual risk

- The mode judgment is text a human or an LLM writes; no command here proves a stated file set is the real one (spec R-2).
- Until the Go SessionStart notice changes (Q4), a lane reads two instructions about the standing cron (spec R-1).
- AC-MJ-006 and REQ-MJ-007 can be invalidated by the Q3 verdict; the plan routes around it (M3 step 2).
- A paraphrased restatement of a rule in a second file evades AC-MJ-016; the sync audit reads for it.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
