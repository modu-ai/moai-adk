# SPEC-ASTGREP-LANG16-001 — Progress

Card t228, Class C, Tier L. Worktree `.claude/worktrees/t228`, branch `WT-astgrep-16-langs`,
HEAD `294b4b6ab`.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-25
tier: L
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md]
requirements: 23
acceptance_criteria: 19
milestones: 3
scope: contract (M1-M3)
successor: SPEC-ASTGREP-BREADTH-001
audit_iteration: 5 audits (0.68 -> 0.69 -> 0.71 -> 0.84 -> PASS 0.88)
spec_version: 0.7.0
audit_verdict: PASS 0.88 (iter5); open minor findings: N5 §A.6 half (note-half closed), N6
gate_recheck: ".moai/reports/plan-audit/SPEC-ASTGREP-LANG16-001-2026-08-27.md (narrow-delta; F1+F2-note fixed here)"
```

Plan-phase basis: `.moai/reports/t228/plan-measurements.md` (M1-M4; M5 superseded — the
differential corpus is committed on `origin/main` via `a9eb896ce` / PR #1637 / card t227), plus
`.moai/reports/t228/plan-audit-iter1.md`.

Revision history: lead amendments A1-A3 at v0.2.0; corrections C1-C7 at v0.3.0; plan-audit
iteration 1 revision and the D8 split at v0.4.0; iteration 2's two-column criteria rewrite at
v0.5.0; iteration 3's bounded propagation-only pass at v0.6.0.

**Split (D8, binding).** This SPEC now carries the contract half only — M1-M3. The breadth half
(M4-M10) moved to `SPEC-ASTGREP-BREADTH-001`. Total card scope is unchanged: requirements (2) and
(4) are discharged by the successor, in full, with no external gate between the two SPECs. The
split was forced by the budget — REQ and AC were saturated at 25/25, so the audit's own fixes
could not be added without breaching the Tier L ceiling.

Budget after revision: 23 requirements, 19 acceptance criteria against a Tier L ceiling of 25 —
headroom of 2 and 6 respectively. The propagation pass at v0.6.0 added **no** requirement and
**no** criterion; D1, D3, and D2 are amendments to existing entries.

Iteration-2 findings addressed (v0.5.0, `acceptance.md`): E1 the central criterion passed
vacuously — now asserts count-equals-rule-count, which also discharges REQ-A16-010; E5 REQ-A16-010
had no criterion; E8 the corpus-baseline criterion gained a requirement via renumbering. E2, E3,
E4, E6, E7, and E9 were applied to `acceptance.md` only and their other halves were left
unpropagated — that gap is what iteration 3 found, and what v0.6.0 closes.

Iteration-3 findings addressed (v0.6.0, propagation-only across the five artifacts iteration 2 did
not open): D1 requirement/criterion contradiction on `catalog.yaml` — REQ-A16-018 inverted to the
measured truth, `plan.md` §A.3 and §D corrected; D2 AC-A16-016's citation-or-probe half given a
requirement (REQ-A16-011 third clause) and a work item (`plan.md` M3 item 4); D3 REQ-A16-021 and
REQ-A16-001 re-scoped off `internal/template/templates/**`, with §0's discipline extended to the
requirement layer; D4 the pinning test given a path, an M1 item, an explicit PRESERVE carve-out,
and a restated §E diff command; D5 and D11 `design.md` §3.1/§4 D13 placement stated one way plus
the four-class count and three stale citations; D6 nine renumbering sites; D7 four §I rows
re-derived; D8 this signal block; D9 the two missing HISTORY rows; D10 AC-A16-001's open count
clause bound to a derivable number.

Iteration-1 findings addressed: D1 (corpus gate skips, not fails — four passages corrected, no-skip
obligation added), D2 (R2 mitigation withdrawn, `metadata.cwe` anchor added), D3 (archived
`SPEC-ASTG-UPGRADE-001` reconciled, R6 owned here), D4 (contiguous renumbering), D5 (REQ-A16-012
in `Where … shall` form), D6 (corpus baseline given a criterion), D7 (checker class 1 is a set
comparison, class 4 added), D8 (split), D9 (AC-A16-020 names its sanctioned exemption — that criterion was folded into AC-A16-019 at
v0.5.0; the id is retained here as the audit trail of what iteration 1 asked for), D10
(dangling reference removed), D11 (tautological criterion retired). Optional findings applied:
D12 (`catalog.yaml` named), D13 (rule-tests outside the distributed tree), D14 (neutrality widened
to any human-language text), D15 (unobservable "shall read" moved to `plan.md`), D16 (stray
directory in the pre-flight).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

```yaml
tier: L
scope_files: ~8 (template rules YAML, sgconfig.yml, rule-tests tree, corpus pin test, coverage-matrix.md, checker Go test)
domains: 3 (ast-grep YAML configs, Go test code, markdown contract docs)
file_language_mix: yaml/go/markdown
concurrency_benefit: LOW (M1 harness -> M2 matrix -> M3 severity are strictly sequential)
agent_team_prereqs: not requested
```

| Mode    | Selected      | Rationale                                                              |
|---------|---------------|------------------------------------------------------------------------|
| direct  | not selected  | Multi-milestone Tier L, no trivial path                                 |
| serial  | **selected**  | Coding-heavy, sequential dependency chain, one milestone per spawn      |
| fanout  | not selected  | No independent research fan-out; coding caveat binds                    |
| sweep   | not selected  | Not a uniform mechanical transform; worktree-editing                    |

Decision: serial

Justification: Anthropic's coding-task parallelism caveat applies — the three milestones form a
strict dependency chain (harness before matrix before reclassification), so concurrency buys only
cache pressure and reconciliation cost. Each milestone delegates once to manager-develop with the
Tier L Section A-E template. Boundary case: none — all thresholds resolve unambiguously toward
serial.

Kickoff approval: operator approved "승인 — M1~M3 연속 실행" this session, AFTER the lead dispatch
and prior to any run-phase implementation spawn.
