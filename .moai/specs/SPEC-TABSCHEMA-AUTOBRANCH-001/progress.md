# SPEC-TABSCHEMA-AUTOBRANCH-001 — Progress

Card: t316
Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t316`
Branch: `WT-tabschema-autobranch`
Plan-phase baseline HEAD: `7ed6edb3e`

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`.

- SPEC ID regex self-check executed as Bash — output `PASS`.
- SPEC ID uniqueness confirmed: `ls -d .moai/specs/SPEC-TABSCHEMA-AUTOBRANCH-001` → exit 1,
  `No such file or directory`.
- All defect coordinates in `spec.md` §1 re-verified on this tree, not carried over.
- All eight acceptance criteria carry a measurement taken on `7ed6edb3e`.
- `tab_schema.json` itself was NOT edited in plan phase.

Amended at v0.1.1 (plan-audit iteration-1 defect closure): `tier: S` added to frontmatter; batch
3.10 provenance disclosed in `spec.md` §1; REQ → AC traceability table added as `spec.md` §6 with
covering-REQ citations on every AC heading; two paired sub-criteria added (`AC-TSA-007b` diff-shape,
`AC-TSA-005b` embedded asset); the schema's pre-existing counter drift recorded out of scope. No
existing REQ or AC was renumbered, reworded, or deleted. `tab_schema.json` remains unedited.

### §E.1 Gaps — what plan phase did NOT observe

- **`make build` has not been run**, by design (it would dirty the tree during plan phase). No claim
  is made here about the embedded asset's state, and `bin/` does not exist on this tree
  (`ls: bin/: No such file or directory`). AC-TSA-005b is therefore adopted with a stated-but-not-
  observed RED cell: its red-in-principle basis is the source-side count that a build made today
  would embed, not an observed failing run.
- **The manual-mode `automation` block was confirmed structurally, not from a rendered config.** The
  claim that the manual profile carries an `automation` block like personal and team rests on the
  Go struct definition (`ModeProfile` is one type used for all three modes), not on a generated
  `git-strategy.yaml` for a manual-mode project.
- **Template neutrality is asserted as a delta, not as an absolute.** AC-TSA-008 asserts that the
  scan output is unchanged from a non-zero baseline; it does not assert that the template copy is
  free of SPEC IDs and dates, because it is not, for reasons predating this card.
- **`moai spec lint` has not been run against this SPEC by the plan-phase author.** It is a
  Definition-of-Done item for run phase.
- **No runtime consumer was exercised.** No code reads `tab_schema.json` (`spec.md` §4), so the
  interview semantics AC-TSA-001 encodes were reconstructed from the schema's own condition data
  and never executed.
- **Now closed, previously open:** the trailing-comma deletion boundary. It was derived by reading
  at v0.1.0 and has since been verified by execution — the two objects span lines `516-536` and
  `726-746`, each is the final element of its `questions` array, and the preceding element closes
  with `},` at lines `515` and `725`.

### §E.1 Residual risk — what could still be wrong despite the above

- **AC-TSA-001's `mode_admits` predicate is a reconstruction no code enforces.** With no runtime
  consumer there is no executable oracle for it. The risk is bounded by the change being purely
  subtractive: it removes questions bound to a path no struct field matches, which holds under any
  interview semantics.
- **AC-TSA-007b compares parsed JSON**, so a pure reformat of untouched regions would still read
  `True`. `plan.md` §B forbids serializer round-tripping and the numstat corroborator would show it,
  but no criterion mechanically rejects it.
- **AC-TSA-005b's binary scan is attributable only while the dead-path strings stay confined to the
  schema.** Measured now: zero occurrences anywhere in `internal/`, `pkg/`, `cmd/` outside the two
  `tab_schema.json` copies. Should a future change introduce those strings elsewhere in compiled
  code, the criterion's `0` stops being attributable and needs re-scoping.
- **`plan.md` §C line coordinates drift** the moment anything above line 516 changes. They are
  correct on `7ed6edb3e` and are anchored by `field` value as well as by line number, so the drift
  is recoverable — but a stale re-read against a moved tree would mis-target.
- **The schema's self-declared counters stay wrong after this card**, and the `total_settings` drift
  widens from `60 vs 48` to `60 vs 46`. Recorded out of scope in `spec.md` §4 so the widened number
  is not later attributed to this deletion.

_Status: plan audit PASS (0.86) and Implementation Kickoff Approval granted; run phase executed._

## §E.2 Run-phase Evidence

Cycle: DDD (behaviour-preserving key alignment; no Go production code touched).
Full verdict with per-command verbatim output: `.moai/reports/t316/run-verdict.md`.

All measurements below were taken in this run, against this tree. The pre-edit working copy of both
`tab_schema.json` files was byte-identical to the pinned baseline `7ed6edb3e` — observed, not
assumed: the post-change `git diff 7ed6edb3e` contains exactly the two deletion hunks and nothing
else.

### AC matrix

| AC | Command | Actual output | Status |
|---|---|---|---|
| AC-TSA-001 | `python3 ac001.py LOCAL` | `personal=1` `team=1` `manual=0` | PASS |
| AC-TSA-001 | `python3 ac001.py TMPL` | `personal=1` `team=1` `manual=0` | PASS |
| AC-TSA-002 | `grep -c 'git_strategy.personal.auto_branch' LOCAL` ; `… .team. …` | `0` ; `0` | PASS |
| AC-TSA-002 | same two greps on `TMPL` | `0` ; `0` | PASS |
| AC-TSA-003 | `grep -cF 'git_strategy.{mode}.automation.auto_branch' LOCAL` ; `… TMPL` | `3` ; `3` | PASS |
| AC-TSA-004 | `grep -c 'auto_branch' LOCAL` ; `… TMPL` | `3` ; `3` | PASS |
| AC-TSA-005 | `diff -q LOCAL TMPL; echo "diff_rc=$?"` | `diff_rc=0` | PASS |
| AC-TSA-005b (1) | `make build; echo "make_build_rc=$?"` | `make_build_rc=0` | PASS |
| AC-TSA-005b (2) | `grep -aoF 'git_strategy.personal.auto_branch' bin/moai \| wc -l` | `0` | PASS |
| AC-TSA-005b (2) | `grep -aoF 'git_strategy.team.auto_branch' bin/moai \| wc -l` | `0` | PASS |
| AC-TSA-005b (3) control | `grep -aoF 'git_strategy.{mode}.automation.auto_branch' bin/moai \| wc -l` | `4` (≥ 1) | PASS |
| AC-TSA-006 | `python3 -m json.tool LOCAL > /dev/null; echo "local_rc=$?"` | `local_rc=0` | PASS |
| AC-TSA-006 | `python3 -m json.tool TMPL > /dev/null; echo "template_rc=$?"` | `template_rc=0` | PASS |
| AC-TSA-007 | `python3 ac007.py LOCAL` ; `… TMPL` | `TOTAL_QUESTIONS = 46` ; `TOTAL_QUESTIONS = 46` | PASS |
| AC-TSA-007b | `python3 ac007b.py LOCAL` | `REMOVED_OBJECTS = 2 ['git_strategy.personal.auto_branch', 'git_strategy.team.auto_branch']` / `IDENTICAL_AFTER_REMOVAL = True` | PASS |
| AC-TSA-007b | `python3 ac007b.py TMPL` | same two lines, `IDENTICAL_AFTER_REMOVAL = True` | PASS |
| AC-TSA-008 | `grep -n 'SPEC-\|20[0-9][0-9]-' TMPL` | 3 lines — `3` schema_updated, `604` / `873` branch-prefix questions; content byte-identical to baseline (`diff` of the two match sets → `neutrality_diff_rc=0`) | PASS |

### Invariants

| Invariant | Command | Actual output | Status |
|---|---|---|---|
| Batch 3.10 untouched | `grep -cF 'git_strategy.{mode}.automation.auto_branch'` on both copies | `3` / `3` | PASS |
| `manual=0` stays `0` (out-of-scope gap not "fixed") | `python3 ac001.py` on both copies | `manual=0` / `manual=0` | PASS |
| Both copies valid JSON after final-element deletion | `python3 -m json.tool` on both | rc `0` / `0` | PASS |
| Template/local byte-identity | `diff -q LOCAL TMPL` | `diff_rc=0` | PASS |
| No question other than the two altered | `ac007b.py` deep equality vs `7ed6edb3e` | `True` on both copies | PASS |
| Scope containment | `git status --short` | exactly the two `tab_schema.json` files (`catalog.yaml` rewritten by the build but byte-identical) | PASS |
| `moai spec lint` clean for this SPEC | `bin/moai spec lint > …/spec-lint.txt 2>&1; echo $?` then `grep -n 'SPEC-TABSCHEMA-AUTOBRANCH-001'` | `spec_lint_rc=0`; zero matches; file ends `0 error(s), 64 warning(s)` (all naming other SPECs) | PASS |

### Deviations recorded rather than smoothed over

1. **AC-TSA-005b control reads `4`, not `3`.** A second embedded template file
   (`.claude/skills/moai/workflows/sync/delivery.md`) carries one occurrence of the canonical string;
   `3 + 1 = 4`. The criterion asserts `>= 1`, so PASS. Measured with
   `grep -rlF … internal/template/templates/`.
2. **AC-TSA-007b's textual corroborator missed its predicted numbers.** Predicted `2` added / `44`
   deleted per copy; measured `0` / `42`. Investigated per the criterion's own instruction: the diff
   is two pure-deletion hunks of 21 lines each with zero additions, because the deleted object's own
   closing line `            }` is byte-identical to the line the preceding element now needs, so git
   models the trailing-comma strip as part of one contiguous deletion instead of a delete+add pair.
   Content is unaffected; the deciding structural check reads `True` on both copies.
3. **AC-TSA-008 discharged on line content, not on the `grep -n` prefix.** The prefixes necessarily
   shift under a deletion above them (`625 → 604`, `915 → 873`, exactly `-21` and `-42`). The match
   set was compared directly against the baseline's and came out byte-identical.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: <M1-PLACEHOLDER — backfill after commit>
run_status: complete
ac_pass_count: 10          # AC-TSA-001..008 + 005b + 007b
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not_run   # no push in run phase per dispatch constraint; branch is local-only
l44_post_push_fetch: not_run    # not pushed — merge window is owned by the coordinating session
new_warnings_or_lints_introduced: 0   # moai spec lint: 0 errors, 64 warnings, none naming this SPEC
cross_platform_build:
  performed: false
  reason: >-
    The change is a JSON schema asset with no OS-conditional content. `make build` (host darwin/arm64)
    exited 0 and the embedded-asset scan is the criterion of record (AC-TSA-005b). No GOOS matrix
    build was run; the cross-platform verdict belongs to CI.
total_run_phase_files: 2   # LOCAL + TMPL (plus SPEC-directory and report artifacts)
m1_to_mN_commit_strategy: >-
  Single commit. M1 (template deletion), M2 (`make build`), M3 (mirror sync), and M4 (acceptance
  battery) are one indivisible change: M2 and M3 have no standalone meaning without M1, and the
  Template-First rule requires the pair to land byte-identical, so splitting would land an
  intentionally-inconsistent tree.
```


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
