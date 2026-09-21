# SPEC-AUDIT-EXPORT-CLAUSE-001 — progress

Card: t1059 · Branch: `WT-audit-export-clause` · Evidence: `.moai/reports/t1059/`

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-22
tier: M
artifacts: spec.md, plan.md, acceptance.md
spec_version: 0.3.4
```

2026-09-22 · v0.3.4 · operator-disposition amendment (mid-run, run open) —
AC-AEC-003 re-baselined to the ruled final state, AC-AEC-004 sweep narrowed to
the twelve wording surfaces, AC-AEC-014 exception table extended 2 → 9.

v0.3.1 lands two bounded lead-review additions and reopens nothing else: the
`check-ignore` discriminant is promoted from a plan instrument note to SPEC §A.2a
with REQ-AEC-014 + AC-AEC-015, and §E records the basis for the 13-file count at
the Tier M band edge. Requirements 13 → 14, criteria 14 → 15, both absorbed out
of existing headroom. AC-AEC-015 is adopted as a **regression-guard**, not a
release-blocking criterion: it is vacuous against the current tree (0 matches
pre-implementation) and its observed red was taken against the v0.2.0 draft.

Plan-phase artifacts authored by manager-spec from card t1059. v0.3.0 rewrites
them for **direction A** — the operator ruled the 2026-09-14 directive canonical,
so lead-verdict evidence stays on disk and the `git add -f` permission clause is
withdrawn rather than repaired. The card now also enacts the ruling by withdrawing
the `.gitignore` negation, and absorbs the paraphrase surfaces v0.2.0 excluded.

All SPEC §A measurements were re-taken in this worktree at HEAD `64c7edbf3`
(local `develop` absorbed); none were carried over from the dispatch or from the
v0.2.0 draft. Two did not survive re-measurement and are recorded as such: the
paraphrase phrase `tracked** path` no longer exists in this tree (SPEC §A.6), and
`git check-ignore -v` was found to exit 0 on a negation match (plan §E).

## §E.2 Run-phase Evidence

Milestones M2-M6 executed 2026-09-22 in `.claude/worktrees/t1059`, branch
`WT-audit-export-clause`, from base `ec522aeed` (M1 `113e487c2` already landed).
Commits: M2 `6a0cf26ec` · M3 `dd81a6538` · M4 `4679f12b9` · M5 `e205987e2`.
Working tree clean after each.

Run closure, 2026-09-22: after the operator disposition and the manager-spec
v0.3.4 amendment (commit `51502212e`), AC-AEC-003/004/014 were re-measured at
HEAD `51502212e` under the amended criteria (all three flipped to PASS,
below), AC-AEC-013 was re-executed at the moved HEAD, and AC-AEC-005/011 were
spot-check re-runs — both confirmed. The remaining §E.2 rows (AC-AEC-001/002,
006–010, 012, 015) are their M-phase PASS records cited verbatim: the
amendment touched only `.moai/specs/**`, so their subjects are unchanged, and
they were not silently re-marked.

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-AEC-001 | PASS | `grep -nE '^!\.moai/reports/[^/]*/.*verdict' .gitignore` ; `grep -nc 'Narrow exception (card t1039)' .gitignore` | no output, `negation_exit=1`; `0`, `block_exit=1`. Controls `1` and `1` |
| AC-AEC-002 | PASS | `git check-ignore --no-index` on the three names + control + two-form self-check | `IGNORED`, `IGNORED`, `NOT-IGNORED`; control `NOT-IGNORED`; self-check `verbose_exit=0` / `plain_exit=1` on `t338/ac-count-baseline.txt` |
| AC-AEC-003 | PASS | anchored `git ls-files` / `git ls-tree origin/develop` + two-way `comm` — re-measured under the amended v0.3.4 criterion after an explicit `git fetch origin develop` (tip `269fb89c1`, fetched 2026-09-22) | `comm -23` empty; `comm -13` exactly `.moai/reports/t1039/verdict.md`, `.moai/reports/t1048/verdict.md` (the two ruled names, no others); `git status` empty, `status_exit=0`. See §E.2a |
| AC-AEC-004 | PASS | amended 12-file sweep `grep -rnE 'tracked\*{0,2} (citation target\|verdict file\|path)'` + control `local\*{0,2} record` — RED probe re-derived on pinned `269fb89c1` blobs in the same run | sweep: no output, `exit=1`; control `6`. RED probes at `269fb89c1`: `1, 1, 2, 1, 1, 2` (8 lines, all exit 0); instrument control `citation target` → `1`. See §E.2b |
| AC-AEC-005 | PASS | `grep -rc 'Never write the verdict to a gitignored location'` ×4 | four lines each `:0`; control `Export mandate` four lines each `:1` |
| AC-AEC-006 | PASS | `grep -rc 'local by design'` ×4 | four lines each `:1` |
| AC-AEC-007 | PASS | `grep -rnE 'add -f\|--force.*reports\|push.*verdict'` over 12 changed files | no output, `exit=1`; control `do not force it into the tree` → `4` |
| AC-AEC-008 | PASS | two FORBIDDEN counts + two normalised `awk` probes over § Where | `:1` ×2 and `:1` ×2; third command `0`; fourth `1`. Negative-branch control on pinned blob `269fb89c1` → `1` |
| AC-AEC-009 | PASS | `grep -rc 'incomplete audit'` ×4 | four lines each `:1` |
| AC-AEC-010 | PASS | two `awk` section reads + three greps | `Audit artifacts are tracked files` → `0`; `ls-files --error-unmatch` → `0`; `disposal hazard` → `1` (control fires). § Committing carries `is a local file`; § What makes carries `ls .moai/reports/<card-id>/` and no reachability check |
| AC-AEC-011 | PASS | `cmp` on the two convention copies | `BYTE-IDENTICAL`, `cmp_exit=0` |
| AC-AEC-012 | PASS | `make agents-emit-check` + anchored greps on the three TOMLs | `exit=0`; plan-auditor `:0`, sync-auditor `:0`, manager-lead `:0`. Controls `:1`, `:1`, `4` |
| AC-AEC-013 | PASS | forbidden-content-class sweep over `develop...HEAD` added lines under the template tree — re-executed at the post-amendment HEAD (the amendment added no template lines) | no output, `exit=1`; control `local by design` → `5` |
| AC-AEC-014 | PASS | amended 13-file stative/adverbial sweep over `develop...HEAD` + two-direction set equality against the 9-entry exception table | exactly 9 matches, `exit=0`; per-file attribution = E1–E9 with no extra line (`.gitignore` 2 = E1/E2; C1 protocol + C1 lead 1+1 = E3/E5; C2 mirrors 3 = E4/E6/E7 incl. emitted toml; convention + mirror 2 = E8/E9). Control `18`; mutant probe `3` / control-mutant `1` (instrument live). See §E.2c |
| AC-AEC-015 | PASS (regression-guard) | `grep -rnE 'check-ignore[^`]*-v'` over the twelve target documents | no output, `exit=1`; reachability control `:1` ×2. RED cell on pinned `622e25d22` → `3`; green path over spec.md §C → `0` |

### §E.2a AC-AEC-003 — resolved by operator disposition; re-measured PASS

The two paths M1 removed from the index at `113e487c2` (2026-09-22 03:52) were
already present on `origin/develop` when it did so:

- `.moai/reports/t1039/verdict.md` landed via `b765153ae` (2026-09-21 17:24),
  an ancestor of `origin/develop` (`git merge-base --is-ancestor` → exit 0).
- `.moai/reports/t1048/verdict.md` landed via `d72df7454` (2026-09-21 22:03).

Re-probed after an explicit `git fetch origin develop`:
`git cat-file -e origin/develop:.moai/reports/t1039/verdict.md` → exit 0, with
the §A.3 control `t965/plan-audit-verdict.md` → exit 0 in the same run. SPEC
§A.3 recorded exit 128 for the same probe, so the reachability premise the
ruling rested on was measured against a stale remote-tracking ref.

Per acceptance §B this is FAIL and a blocker. No index change and no history
rewrite was performed: REQ-AEC-003 forbids both for an already-published path,
and re-opening the disposition is an operator question.

**Disposition resolution (2026-09-22).** The operator's disposition ruled the
observed state final (SPEC §A.3a): the two files stay on `origin/develop` and
stay intentionally absent from this branch's index. manager-spec re-baselined
the Then-clause accordingly in the v0.3.4 amendment (commit `51502212e`). This
run re-measured the criterion at HEAD `51502212e` after an explicit
`git fetch origin develop` (tip `269fb89c1`, fetched 2026-09-22): `comm -23`
empty, `comm -13` exactly the two ruled names, `git status` empty,
`status_exit=0` — the PASS state of the amended criterion. Restoration and
re-staging of either file remain forbidden under REQ-AEC-003.

### §E.2b AC-AEC-004 — resolved by operator disposition; re-measured PASS

The single remaining match was `internal/template/templates/.moai/README.md:77`
(*"regenerable artifacts out of tracked paths"*), introduced by `d97980654`
(2026-07-07) and untouched by this card. It is not in plan §C's scope table, so
REQ-AEC-004 — which binds *the surfaces enumerated in the implementation plan's
scope table* — is satisfied while the criterion is not. This is the same
reach mismatch AC-AEC-001 was narrowed to repair, in the opposite direction:
the criterion's trees were widened past the requirement's scope, and the `path`
alternative then catches a sentence about git conventions rather than a
card-report artifact.

**Disposition resolution (2026-09-22).** The operator's disposition narrowed
the sweep to the twelve wording surfaces of plan §C rows #2–#13 — excluding
`.gitignore` (row #1; its act is the withdrawal, measured by AC-AEC-001/002)
and `internal/template/templates/.moai/README.md` (the pre-existing
out-of-scope line above), with both exclusion reasons recorded. manager-spec
applied the narrowing in the v0.3.4 amendment (`51502212e`). This run
re-measured at HEAD `51502212e`: the 12-file sweep returns no output,
`exit=1`, with the paired control at `6`; the recorded RED was re-derived —
not cited on trust — against the pinned `269fb89c1` blobs in the same run
(probes `1, 1, 2, 1, 1, 2` = 8 lines; instrument control `1`).

### §E.2c AC-AEC-014 — resolved by operator disposition; re-measured PASS

Nine matches against a 2-entry table. The two `.gitignore` matches are E1 and
E2. The seven others, by file:

- `it is gitignored` — subject `.moai/state/verify/<session>/`, the
  machine-local scratch directory, in `agent-common-protocol.md` (C1 + C2) and
  `manager-lead.md` (C1 + C2 + the emitted `manager-lead.toml`): 5 matches.
  Pre-existing prose this card did not author; the lines appear as additions
  only because a line-granularity diff re-adds a whole line whose other half
  changed. REQ-AEC-012 binds sentences *this SPEC introduces*.
- `is already tracked` in `audit-artifact-convention.md` (local + mirror):
  2 matches. This is SPEC §C.2's own specified closing paragraph, which §C.2's
  commentary states is *"general git behaviour plus a conditional about a
  project, not an assertion about this one — which is what keeps it inside
  REQ-AEC-012"*. Identical in kind to E1.

Every extra falls in a class REQ-AEC-012 explicitly permits, so no wording was
changed in response. Per acceptance §B the criterion nonetheless FAILS, and
extending the table is *"a SPEC edit with its own justification, never a
close-time judgement"* — so it is reported rather than performed.

**Disposition resolution (2026-09-22).** The operator's disposition extended
the declared exception set from 2 to 9 entries; manager-spec authored the
extension with per-entry permission classes in the v0.3.4 amendment
(`51502212e`). This run re-measured at HEAD `51502212e`: the verbatim sweep
returns exactly 9 matches (`exit=0`), and the equality check holds in **both**
directions — every match is a table member (per-file attribution: `.gitignore`
2 = E1/E2; C1 `agent-common-protocol.md` + C1 `manager-lead.md` = E3/E5; C2
mirrors of both + the emitted `manager-lead.toml` = E4/E6/E7; convention +
mirror = E8/E9) and every table entry is matched. Paired control `18`;
close-time mutant probe re-run: mutants `3`, control-mutant `1` — the
instrument is live.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-22
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 15
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: performed
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0
cross_platform_build:
  status: not-applicable
  reason: documentation-only change; no Go source modified
total_run_phase_files: 15
m1_to_mN_commit_strategy: one commit per milestone (M2 6a0cf26ec, M3 dd81a6538, M4 4679f12b9, M5 e205987e2); M6 is verification-only and writes no source
blockers:
  - AC-AEC-003 — resolved by operator disposition 2026-09-22: the removed-index state was ruled final (SPEC §A.3a); manager-spec re-baselined the Then-clause in the v0.3.4 amendment (51502212e); re-measured PASS at HEAD 51502212e after an explicit origin/develop fetch
  - AC-AEC-014 — resolved by operator disposition 2026-09-22: the exception table was extended 2 → 9 via the v0.3.4 amendment (51502212e); re-measured PASS at HEAD 51502212e — exactly the 9 table lines, equality holds both directions
  - AC-AEC-004 — resolved by operator disposition 2026-09-22: the sweep was narrowed to the twelve in-scope wording surfaces via the v0.3.4 amendment (51502212e); re-measured PASS at HEAD 51502212e (no output, exit=1; control 6; RED re-derived on pinned 269fb89c1)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
