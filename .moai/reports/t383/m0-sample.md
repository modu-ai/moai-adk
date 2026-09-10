# t383 M0 — premise sampling gate

Discharges AC-MSR-015. Four required items in order: (1) the selection and its rule,
(2) a per-file verdict with deciding evidence, (3) the count against the threshold,
(4) the coverage limit.

**Outcome: 0 of 12 superseded. Threshold is ≥ 4 of 12. The gate does NOT trip — PROCEED.**

## 1. The selection and the rule that produced it

Population: the **58** unique missing `.md` link targets, **whole-file scope** — every
markdown link on any line shape, not just `^- \[` entry lines. This is the population M3
copies and the same count `moai memory doctor` reports as `MEMORY_DANGLING_INDEX_LINK`.

Derived from the **live** index at M0 time (plan.md §F M0 step 3), not from any frozen
list, by `.moai/reports/t383/derive-missing.sh` (committed as evidence):

```
$ bash .moai/reports/t383/derive-missing.sh /tmp/t383-missing.txt /tmp/t383-sample.txt
index:                 /Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md
unique targets:        196
unique missing:        58
sampled (every 5th):   12
coverage limit:        indices not reached by the rule are listed below
highest sampled index: 56  (of 58)
```

The live population is **58**, matching plan-phase exactly and matching the doctor's
`.Code` count in the same window. The index grew from 190 to 196 unique targets between
plan and run; the missing count did not move. No forcing of an old number was required.

Rule: sort the 58 lexicographically (`LC_ALL=C`, so the order is locale-independent and
the indices are reproducible); select every 5th, 1-indexed → 1, 6, 11, … 56 → exactly 12.
No discretion.

## 2. Per-file verdict with deciding evidence

Decision rule (plan.md §F M0 step 3): **superseded** iff *either* the body carries a
`[SUPERSEDED by …]` marker *or* the frontmatter `description:` is subsumed by an existing
active-store file (subsuming file named). Otherwise **live**.

Both arms were run mechanically over the whole sample.

**Arm 1 — SUPERSEDED marker scan.** Zero matches across all 12:

```
$ cut -f2 /tmp/t383-sample.txt | sed "s|^|$L/|" | xargs grep -l 'SUPERSEDED'
(no output)
```

**Arm 2 — subsumption scan.** Keyword searches over the active store for each sampled
file's subject (`full suite` / `go test ./...`, `factory`, `navigator`, `update-branch`,
`env -u`, `moai update`) returned only files that *mention* a keyword in passing — none
whose `description:` subsumes a sampled file's. No subsuming file could be named for any
of the 12, so arm 2 fires for none of them.

**Precondition confirmed for all 12** — absent from the active store, present in the
legacy store (so each is genuinely copyable, not lost):

```
$ cut -f2 /tmp/t383-sample.txt | sed "s|^|$D/|" | xargs ls -1
ls: .../feedback_bg_exitcode_direct_verify.md: No such file or directory      (x12, all absent)
$ cut -f2 /tmp/t383-sample.txt | sed "s|^|$L/|" | xargs ls -1
(all 12 listed — all present)
```

| # | File | Verdict | Deciding evidence |
|---|---|---|---|
| 1 | `feedback-resume-race-fresh-spawn.md` | **live** | no marker; description "SendMessage-resume of a paused subagent can delay-execute; combining it with a fresh spawn races on the same files" — no active-store file covers subagent resume-vs-spawn racing |
| 6 | `feedback_bg_exitcode_direct_verify.md` | **live** | no marker; "background-task and `gh pr checks --watch` exit codes can lie (exit 0 while checks actually failed)" — no active-store equivalent |
| 11 | `feedback_ci_race_update_branch.md` | **live** | no marker; "PR CI-running race … resolve via GitHub update-branch API, not force-push" — `update-branch` search found no subsuming active file |
| 16 | `feedback_full_test_suite_verification.md` | **live** | no marker; "Run-phase 검증은 FULL go test ./... 이어야 한다 — cross-cutting CI 가드는 affected-package 집합 밖" — see the tension note below |
| 21 | `feedback_lsel_f1_actual_write_surface.md` | **live** | no marker; "AC/테스트는 선언 의도가 아닌 실제 쓰기 표면(patch path)을 검증해야" — no active equivalent |
| 26 | `feedback_no_local_full_suite.md` | **live** | no marker; "Do not run `go test ./...` locally in this repo" — see the tension note below |
| 31 | `feedback_prompt_premise_verification.md` | **live** | no marker; "지시문에 담아 보내는 전제는 근거를 명시할 것 — 한 세션에 내 전제가 7번 틀렸다" — no active equivalent |
| 36 | `feedback_stale_moai_update_regression.md` | **live** | no marker; "뒤처진 체크아웃에서 moai update가 돌면 구버전 템플릿으로 덮여 구현된 SPEC이 조용히 되돌아감" — `moai update` search found no subsuming file |
| 41 | `feedback_verification_load_discipline.md` | **live** | no marker; "Parallel lanes running full test suites melted the box; a recipe whose cleanup is on a trailing line leaks forever" |
| 46 | `feedback_worktree_guard_env_un_wrapper.md` | **live** | no marker; "the worktree guard refuses `env -u VAR cmd`; use `unset VAR … && cmd`" — independently re-confirmed live during this very run-phase (the guard refused several commands here) |
| 51 | `project_factory_mode_closed.md` | **live** (archive roll-up representative) | no marker. A *closed* record is not a *superseded* one: the index line `- factory/goal·MX·기타: [FACTORY-MODE](project_factory_mode_closed.md) 외 17편` names it as the reader's entry point into 17 archived records (spec.md §A.2.2 reason 1) |
| 56 | `project_navigator_epic.md` | **live** (archive roll-up representative) | no marker. Index line `- BAS/Navigator: [Epic](project_navigator_epic.md) 외 6편` — same reasoning |

**A tension inside the sample, recorded not resolved.** Entries 16 and 26 give opposite
instructions about running the full suite locally (16: run FULL; 26: never locally, push
and read CI). Neither carries a SUPERSEDED marker, so under the stated rule both are
**live** and both are copied. The rule is applied as written; adjudicating which is
current is a per-file judgement and belongs to the deferred triage card, not here.
Recording it matters because a reader who opens both will meet a contradiction, and this
card's copy step is what makes both openable.

## 3. Count against the threshold

| | |
|---|---|
| superseded | **0** |
| live | **12** |
| stop threshold | **≥ 4 of 12** → return a blocker |
| observed | 0 < 4 |
| **decision** | **PROCEED** |

This corroborates rather than inherits the auditor's independent 10-file read (plan.md
§B.1): a different sample, a different rule, the same direction. The copy direction stands.

## 4. Coverage limit

The rule reaches index **56 of 58**. Two entries lie outside the sample and were **not**
examined:

```
57  project_spec_catalog_cleanup.md
58  reference_mcp_2026_07_28_check.md
```

Second limit, named because it is easy to miss: the sample is 12 of 58 — **21%**. A
0-of-12 result bounds the superseded share loosely, not tightly. It is evidence the copy
direction is right, not proof that no individual file among the 58 is superseded. The
per-file question stays open for the deferred triage card, exactly as spec.md §D scopes it.

Third limit: arm 2 (subsumption) was executed as a keyword search over the active store,
not an exhaustive semantic comparison against all 185 active files. A subsuming file whose
wording shares no keyword with the sampled file would have been missed.
