---
id: SPEC-AC-BASELINE-REFRESH-001
title: "Durable AC-count corpus baseline refresh — in-tree regeneration mode, lifecycle-tied cascade procedure, provenance header"
version: "0.1.1"
status: in-progress
created: 2026-09-22
updated: 2026-09-22
author: lane agent-20 (t1068)
priority: P1
phase: "v3.1.0"
module: "internal/spec,.moai/docs,.moai/reports/t338"
lifecycle: spec-anchored
tier: M
depends_on:
  - SPEC-AC-COUNT-DISCRIMINATOR-001
tags: "ac-count,baseline,snapshot,regeneration,test-infra,lifecycle-cascade"
---

# SPEC-AC-BASELINE-REFRESH-001 — Durable AC-count corpus baseline refresh

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-22 | lane agent-20 (t1068) | Initial plan-phase draft (card t1068). Disposition: split — regenerate now + durable in-tree mechanism + lifecycle-tied cascade procedure. Fork A-alone (regenerate only) and Fork B (judgment redesign / auto-absorb) rejected with reasons (§A.6). |
| 0.1.1 | 2026-09-22 | lane agent-20 (t1068) | Plan-audit iter-1 FAIL (0.847) repair — narrow: D1 self-inclusion staleness re-pinned (§A.1/§A.3/§A.4; own acceptance.md entered the glob population, 83→84; "two reads agree" instant-qualified); D2 AC-ABR-002 whole-tree grep predicate narrowed to owned surfaces (two historical hits are outside edit rights by design); D3 count-neutral-split arithmetic recorded (§A.3). Fresh re-measurement pinned with instant+command (§A.3 second block). |

## §A Context

### A.1 Problem — the gate is red for an operational reason, not a judged regression

`TestACCounterFullCorpusMatchesBaseline` (`internal/spec/ac_count_clause_test.go:415`) compares the tracked corpus snapshot `.moai/reports/t338/ac-count-baseline.txt` (715 recorded entries — static between regenerations) against a live run of the frozen corpus glob (798 matches at the 2026-09-22 second measurement, §A.3 — the match count MOVES with every `acceptance.md` authored anywhere in the tree, this SPEC's own plan artifacts included; figures are re-derived at use, never carried). The `internal/spec` package is RED right now. Per card [HARD] 2 the two reported numbers are separated FIRST — reported row counts are not failure counts:

- **absent axis (report-only — not failures)**: 84 rows of `absent-from-snapshot … reported, not failed (spec.md 3.5 rule 4)` at the 2026-09-22 second measurement (§A.3) — `acceptance.md` files added to the corpus since the last regeneration (t573), the 84th being this SPEC's own `acceptance.md` (COUNT 9), which entered the population when plan-phase authored it. By the v0.5.0 narrowing these are observations awaiting blessing, never failures. All 84 rows carry `COUNT` observations; zero carry `HALT` (corpus halting count is 0 at that same measurement).
- **failure axis (the actual red — exactly one row)**: `ac_count_clause_test.go:479` — `.moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/acceptance.md: present in the snapshot but no longer matched by the corpus glob`.

Commit `20cdeb6bd` superseded SPEC-MODEL-PROFILE-MATRIX-002 and split it into four successors (`SPEC-MODEL-MATRIX-{CORE,CONFIG,DOCS,SURFACES}-001`), deleting its `acceptance.md` as part of legitimate cleanup; the directory now carries only `spec.md` with `status: superseded`. The vanish guard (`SPEC-AC-COUNT-DISCRIMINATOR-001` spec.md §3.5 rule 1: a recorded file the glob no longer matches is a failure — "사라짐은 이미 잰 관측이 없어진 것이다") fired exactly as designed. The current red is therefore a **stale-snapshot false positive on a legitimate lifecycle event**: the judgment is not the defect; the missing refresh path is.

### A.2 Original intent — what the gate guards, and what regeneration may not break

`SPEC-AC-COUNT-DISCRIMINATOR-001` (card t338, commit `23df21c9e`) fixed the gate's contract, and its v0.5.0 amendment (2026-08-28) re-adjudicated the failure surface:

- **The disease**: "이미 잰 파일의 조용한 과다 계상 회귀" — silent count drift, state move, or vanishing on files the snapshot ALREADY measured. Those remain hard failures (§3.5 rules 1–3).
- **New files**: absence-first (§3.5 rule 4) — reported, not failed, because "회귀가 아닌 것을 실패로 부르면 1번이 지키려는 신호가 잡음에 묻힌다"; the report is required output, not optional. And the amendment states the absorption path explicitly: "그 파일은 **다음 재생성에서** `COUNT` 또는 `HALT` 항목으로 편입되고, 편입된 뒤부터 1·2·3번이 그 파일을 지킨다" — regeneration absorbing absent files is the DOCUMENTED lifecycle, not a corruption of the gate.
- **The frozen thing is the glob** (depth-1 `.moai/specs/*/acceptance.md`, `_archive` excluded), not the count.

Consequence for this SPEC: the test's judgment semantics are NOT the defect and MUST NOT change. The defect is operational, with three named gaps:

1. **The documented refresh procedure is unreachable.** The snapshot header line 2 says `# Regenerate with .moai/reports/t338/run-scratch/gen-baseline.sh; every line is a measurement.` — that path does not exist in the tree (untracked scratch, lost).
2. **No procedure binds regeneration to lifecycle events.** The t573 precedent (`5f546af2c`) named this duty "cascade" — and the current red is the second instance of the same class: a corpus-affecting commit (`20cdeb6bd`) landed without its cascade.
3. **The snapshot carries no provenance** — no source tree, no date — so snapshot figures cannot be attributed after the fact (the exact failure that produced card [HARD] 1's disputed attributions).

### A.3 Measured evidence

All commands run in worktree t1068 @ `cd99336bf`, 2026-09-22, lane agent-20. **Two measurement instants are recorded.** The corpus population MOVED between them: the corpus glob scans the working tree, so this SPEC's own plan artifacts (its `acceptance.md` first) entered the population when plan-phase authored them. Each block is a complete measurement at its instant; neither is current except when re-derived.

#### First measurement — 2026-09-22, before this SPEC's own artifacts entered the glob population

```
$ go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1
    ac_count_clause_test.go:479: .moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/acceptance.md: present in the snapshot but no longer matched by the corpus glob
    ac_count_clause_test.go:489: AC corpus: 83 file(s) matched by the glob but absent from the snapshot - reported, not failed (spec.md 3.5 rule 4)
    ac_count_clause_test.go:491:   absent-from-snapshot .moai/specs/SPEC-AC-ANCHOR-SCOPE-001/acceptance.md: COUNT 8
    ac_count_clause_test.go:491:   absent-from-snapshot .moai/specs/SPEC-AC-GUARD-001/acceptance.md: COUNT 9
    ac_count_clause_test.go:491:   absent-from-snapshot .moai/specs/SPEC-AGENTS-IGNORE-POLICY-001/acceptance.md: COUNT 6
    … (80 further absent rows, all COUNT)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/spec	5.211s
FAIL
(exit=1; grep -c 'absent-from-snapshot' → 83; grep -c 'HALT' in output → 0)
```

```
$ grep -c '^\.moai' .moai/reports/t338/ac-count-baseline.txt
715
$ find .moai/specs -maxdepth 2 -name acceptance.md -not -path '*_archive*' | wc -l
     797
Arithmetic: 797 matches − (715 recorded − 1 vanished) = 797 − 714 = 83 absent ✓
```

#### Second measurement — 2026-09-22, after this SPEC's own plan artifacts entered the glob population (same tree, same command)

```
$ go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1
    ac_count_clause_test.go:479: .moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/acceptance.md: present in the snapshot but no longer matched by the corpus glob
    ac_count_clause_test.go:489: AC corpus: 84 file(s) matched by the glob but absent from the snapshot - reported, not failed (spec.md 3.5 rule 4)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/spec	5.439s
FAIL
(exit=1; grep -c 'absent-from-snapshot' → 84; grep -c 'HALT' in output → 0)
$ grep 'absent-from-snapshot' <output> | grep 'BASELINE-REFRESH'
    ac_count_clause_test.go:491:   absent-from-snapshot .moai/specs/SPEC-AC-BASELINE-REFRESH-001/acceptance.md: COUNT 9
$ find .moai/specs -maxdepth 2 -name acceptance.md -not -path '*_archive*' | wc -l
     798
Arithmetic: 798 matches − (715 recorded − 1 vanished) = 798 − 714 = 84 absent ✓
   — the 84th row is this SPEC's own acceptance.md (COUNT 9): the plan-authoring act itself moved the population.
```

**Count-neutral split (verified at both instants).** The four successors' live counts in the absent rows sum to 20 (CONFIG) + 13 (CORE) + 21 (DOCS) + 10 (SURFACES) = **64**, exactly equal to the vanished entry's snapshot COUNT — `.moai/reports/t338/ac-count-baseline.txt:328: … SPEC-MODEL-PROFILE-MATRIX-002/acceptance.md  COUNT 64  live=64 excluded=0 ambiguous=0`. The split preserved the AC count: the vanish is purely a lifecycle event, no counted content was lost — and the four successor additions are self-checking lines in the M2 diff review.

```
$ grep -n 'SPEC-MODEL-PROFILE-MATRIX-002' .moai/reports/t338/ac-count-baseline.txt
328:.moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/acceptance.md  COUNT 64  live=64 excluded=0 ambiguous=0
$ ls .moai/reports/t338/run-scratch/
run-scratch MISSING            # the documented generator is gone
$ git log --oneline -1 -- .moai/reports/t338/ac-count-baseline.txt
fe2d2fd05 fix(ci): restore report guard fixtures as tracked exceptions + refresh catalog hashes
   (last content-bearing regen: 5f546af2c fix(t573): regenerate AC counter baseline after corpus rewrite)
$ git show 20cdeb6bd --name-status --format='%s'
refactor(SPEC-MODEL-PROFILE-MATRIX-002): split into four successors on the milestone seams
A  .moai/specs/SPEC-MODEL-MATRIX-CONFIG-001/acceptance.md
A  .moai/specs/SPEC-MODEL-MATRIX-CORE-001/acceptance.md
A  .moai/specs/SPEC-MODEL-MATRIX-DOCS-001/acceptance.md
A  .moai/specs/SPEC-MODEL-MATRIX-SURFACES-001/acceptance.md
D  .moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/acceptance.md
M  .moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/spec.md      # status: superseded
```

```
$ grep -n '479' /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1067/.moai/reports/t1067/baseline-test-merged-tree.log
2:    ac_count_clause_test.go:479: .moai/specs/SPEC-MODEL-PROFILE-MATRIX-002/acceptance.md: present in the snapshot but no longer matched by the corpus glob
```

### A.4 Figure attribution ([HARD] 1 — every figure to its tree + time; the two axes are never merged)

| absent / unmatched | Tree | When | Measurer | Status |
|---|---|---|---|---|
| 68 / 1 | primary checkout (main) | 2026-09-21 | lead, on record | carried attribution — NOT re-measured by this lane |
| 75 / 0 | t1058 merge tree | 2026-09-21 | on record | carried attribution — NOT re-measured by this lane |
| "83 / 0" (dispatch text) vs 83 / 1 (its cited log) | cd99336bf merge tree | 2026-09-22 | lane dispatch prose vs `.moai/reports/t1067/baseline-test-merged-tree.log:2` | dispatch text REFUTED on the unmatched axis — the cited log carries the :479 row |
| 83 / 1 | this worktree t1068 @ cd99336bf | 2026-09-22 | lane agent-20 | §A.3 FIRST measurement (pre-authoring; exit=1) |
| 84 / 1 | this worktree t1068 @ cd99336bf | 2026-09-22 | lane agent-20 | §A.3 SECOND measurement (post-authoring; exit=1); the +1 absent row is this SPEC's own `acceptance.md` (COUNT 9) |

The two PRE-AUTHORING reads of cd99336bf (the t1067-cited log; this lane's first measurement) both showed 83 / 1 **at their instant** — the agreement is instant-dependent, not a property of the tree: after this SPEC's own plan artifacts entered the glob population, the same tree measures 84 / 1 (§A.3 second measurement). A tree-level "agree" claim about the absent axis is meaningless without its instant. The card's monotonic series 68 → 75 → 83 → 84 attributes to the **absent axis**; the **failure axis** is separately 1 → 0 → 1 → 1 and is the red at every measured instant. The series are reported side by side and never merged into one number.

### A.5 Cost of neglect ([HARD] 4 — recorded, with figures)

- **Monotonic growth, no shrink mechanism.** absent 68 → 75 → 83 across three attributed trees in two days (§A.4). The growth engine was measured by the origin SPEC itself: "7일 60 / 3일 29 / 24시간 6" new `acceptance.md` files (origin/develop `947f5cffb`, 2026-08-28) — several per day. Only a regeneration shrinks the axis; nothing else in the tree touches it.
- **Same failure class, third instance.** t348 created the gate + snapshot; t573's corpus rewrite landed WITHOUT its cascade → red on the develop tip itself, repaired by hand through scratch that is now lost (`5f546af2c`: "t573 (d9b472409) rewrote corpus criteria without this cascade; TestACCounterFullCorpusMatchesBaseline went red on the develop tip itself"); `20cdeb6bd`'s superseded-split landed WITHOUT its cascade → today's red. Each repair was heroics; the class recurs because the procedure does not exist.
- **Signal masking is live NOW.** The package is red for a benign reason, so every genuine regression the gate exists to catch — count-move, halt transition, or an unexplained vanish on a recorded file — lands as indistinguishable red in CI. The origin SPEC's rationale for rule 4 ("회귀가 아닌 것을 실패로 부르면 … 신호가 잡음에 묻힌다") describes the state neglect has already produced, one axis over.
- **The report's purpose is being buried.** The 83-row absent backlog is the observation the report exists to surface ("보고는 그 사이의 유일한 가시성"); left unreviewed, the next genuinely-new file drowns in noise.

### A.6 Disposition decision — card t1068's fork, chosen and rejected

Card t1068 forks the disposition in two: regenerate the snapshot (blesses the current state as normal) versus change the test design (auto-absorb new SPECs, or change the criterion — redefines what the gate guards). The intent evidence in §A.2 decides it:

- **Chosen — Fork C (split): regenerate now + durable mechanism + lifecycle-tied procedure.** Grounds: (1) the vanish guard is the regression the gate EXISTS for (origin §3.5 rule 1), so the criterion stays; (2) regeneration absorbing absent files is the origin contract's own documented next step (§3.5 rule 4), so today's regeneration blesses nothing abnormal; (3) the three operational gaps (§A.2) — lost scratch generator, no cascade procedure, no provenance — are each closed by a named requirement (REQ-ABR-001/002/003/004/005) without touching judgment.
- **Rejected — Fork A alone (regenerate, nothing else).** Clears today's red and leaves the mechanism broken: the next supersede, archive, or rewrite re-reddens the gate, and the recipe is again "someone figures it out" — the exact neglect loop §A.5 records at 68 → 75 → 83.
- **Rejected — Fork B (auto-absorb / criterion change).** Inverts the v0.5.0 adjudication on its own terms: auto-absorption converts the REQUIRED report into self-approval — a report nobody reads, blessed by the very run that produced it, making the report-required clause ("보고는 … 선택이 아니다") decorative. Lifecycle-aware vanish handling moves the legitimacy judgment INTO the test, re-implementing SPEC lifecycle policy inside an assertion and opening a silent path (add a marker → delete → no signal). The gate would still exist while no longer guarding what it was built to guard — the hollow-gate outcome card [HARD] 5 exists to prevent.

## §B Requirements (GEARS)

### Layer 1 — durable in-tree regeneration mode

**REQ-ABR-001** — **Where** the operator sets the regeneration environment variable (`MOAI_AC_BASELINE_REGENERATE=1`, unset by default), the `internal/spec` test package shall provide an in-tree regeneration mode that re-runs the frozen corpus glob, measures one snapshot line per matched `acceptance.md` using the SAME counter-extraction and measurement machinery the corpus test uses (`extractCounterCommand` / `runCounter` / `counterLiveCount` / per-state tally), and overwrites the snapshot at its existing tracked path `.moai/reports/t338/ac-count-baseline.txt`. Without the variable the mode shall perform no write.

> **Why inside the test package.** The counter command is extracted at test time from the sentinel-delimited B12 clause in `.claude/agents/moai/manager-docs.md`; those helpers are test-file-local and not importable by a standalone `go run` script. A script under the `internal/template/scripts/gen-catalog-hashes.go` idiom (Makefile:35) would have to REIMPLEMENT the extraction — a second instrument that can drift from the instrument being judged. The scripts idiom is named and rejected on that ground. Environment-gated regeneration of a golden file is a standard Go pattern and keeps ONE implementation of extraction + line format.

**REQ-ABR-002** — The regeneration mode shall emit a provenance header recording (a) the frozen corpus glob statement, verbatim and unchanged, (b) the source tree commit SHA at regeneration time, (c) the regeneration date, and (d) the exact regeneration command — and every data line shall remain one measurement in the current line format (`COUNT <n>` + per-state tally, or `HALT` + `owner=`/`reason=`), sorted by path, so the file round-trips through `parseACBaseline` unchanged.

### Layer 2 — the remedy rides the failure

**REQ-ABR-003** — **When** the corpus test fails on a recorded file — vanish (the `:479` loop) or count/state move (an `acComparison` problem) — the failure output shall name the in-tree regeneration command, so the remedy is carried by the failure itself instead of by tribal memory. What fails and what reports remains exactly as before; only the message text gains the remedy pointer.

### Layer 3 — cascade procedure tied to lifecycle events

**REQ-ABR-004** — The repository shall carry a TRACKED procedure document at `.moai/docs/ac-count-baseline-refresh.md` that (a) names the corpus-affecting lifecycle events that oblige a same-commit cascade — `acceptance.md` removal via superseded/split, SPEC-directory move into `_archive/`, and count-affecting corpus rewrites; (b) fixes the cascade as: run the regeneration mode, review the emitted diff line by line, attribute every changed line to a named cause, land the refreshed snapshot in the SAME commit as the lifecycle event; and (c) is referenced from the snapshot header, the test file's doc comment, and `internal/spec/CLAUDE.md`.

**REQ-ABR-005** — A cascade regeneration commit shall contain the refreshed snapshot and no unrelated corpus change, and every count or state change in its diff shall carry a named cause in the commit message; a diff line without a named cause is an unattributed measurement and fails review.

### Layer 4 — the judgment surface is untouchable

**REQ-ABR-006** — The refresh mechanism shall not alter the corpus test's judgment semantics: absence stays report-only with required per-row output (`SPEC-AC-COUNT-DISCRIMINATOR-001` spec.md §3.5 rule 4), and vanish / count-move / state-move / halting-identifier-set-move on recorded files stay hard failures (§3.5 rules 1–3; AC-ACD-006 item 5). The test shall not absorb absent files into the snapshot by itself — the blessing act is the regeneration run, performed and reviewed by a person.

**REQ-ABR-007** — **When** a snapshot-recorded `acceptance.md` is removed from the corpus WITHOUT a cascade, the corpus test shall fail naming that file at the vanish site, and the absence report-not-fail narrowing shall survive unchanged (`TestACBaselineComparisonTransitions` absent rows keep passing).

## §C Out of Scope

### Out of Scope — judgment-semantics redesign
- Any change to `acComparison`'s problem/report contract, the absence-first ordering, or the vanish hard-fail. Includes lifecycle-aware vanish handling ("vanish is OK when the remaining spec.md says superseded/archived") — that would re-implement SPEC lifecycle policy inside a test assertion, invert the origin SPEC's v0.5.0 adjudication, and open a silent path (add a marker → delete the file → no signal). This is Fork B of card t1068's fork; rejected (§A.6 rationale).

### Out of Scope — snapshot relocation
- Moving `.moai/reports/t338/ac-count-baseline.txt`. It is a tracked exception inside the local-only `.moai/reports/` root (restored by `fe2d2fd05`), is wired into `internal/spec/ac_count_clause_test.go` (`acBaselineSnapshotPath`) and appears as fixture data in `internal/cli/todo_triage_test.go:444`. Same disposition as SPEC-AC-GUARD-001 §7 "Script-file relocation fixes".

### Out of Scope — per-file review of the 83 absent files
- Regeneration blesses them into the snapshot — that is the origin SPEC's documented next step (§3.5 rule 4). AC-quality review of those files belongs to the authoring convention domain (SPEC-AC-GUARD-001's declared, now-completed scope), not here.

### Out of Scope — CLI surface for regeneration
- A `moai spec …` subcommand or a Makefile target wrapping the regeneration mode. The test-package mode plus the documented one-line command is the minimal durable surface; an alias can be added later if the command proves too long to type.

### Out of Scope — mechanical cascade enforcement
- A CI check that diffs corpus vs snapshot per commit to FORCE the same-commit cascade. The procedure document is the contract; if the class recurs after this SPEC lands, automation is a separate card.

## §D Success Criteria

1. The snapshot is regenerated in-tree; `TestACCounterFullCorpusMatchesBaseline` passes with an absent-report of 0 rows; `TestACBaselineComparisonTransitions` still passes (narrowing intact).
2. The regeneration mode exists, is default-off, and is reachable by one documented command; its output round-trips through `parseACBaseline`.
3. Failure output on a recorded-file violation names the remedy command (verified by mutation, AC-ABR-005).
4. The cascade procedure document is tracked and referenced from the snapshot header, the test doc comment, and `internal/spec/CLAUDE.md`.
5. Mutation AC (card [HARD] 5): a recorded `acceptance.md` deleted without a cascade still fails the corpus test, naming the file; the v0.5.0 report-not-fail narrowing survives.
6. `internal/cli` consumer (`TestTodoTriage*`, fixture path use only) stays green; the snapshot path is unchanged.
7. Diff audit of `internal/spec/ac_count_clause_test.go` shows `acComparison`'s contract and the corpus test's error conditions unchanged; additions are the mode, message suffixes, and doc comment only.

## §E Dependencies, Adjacent Work, and Prior Art

- **`SPEC-AC-COUNT-DISCRIMINATOR-001`** (card t338, `23df21c9e`) — origin SPEC; owns the gate's judgment contract (spec.md §3.5 rules 1–4, AC-ACD-006). This SPEC does not modify that contract; it IMPLEMENTS the operational layer that contract already assumes ("다음 재생성에서 편입") and makes the regeneration path durable. Declared dependency (`depends_on`).
- **`SPEC-AC-GUARD-001`** (card t1067, completed 2026-09-22) — declared scope is the guard-refused-AC census, the AC authoring convention, and verified-block rewrites (its §1; its §7 Out of Scope rows leave the baseline mechanism outside its territory). This SPEC modifies no `acceptance.md` body and no guard surface; SPEC-AC-GUARD-001's `acceptance.md` appears among the 83 absent rows and is blessed at regeneration like every other file. No collision.
- **t573 cascade precedent** (`5f546af2c`) — the by-hand execution this SPEC turns into procedure; its commit message is the format model for named-cause cascade commits.
- **t1067 cited log** (`.moai/reports/t1067/baseline-test-merged-tree.log`, worktree t1067) — the attribution-divergence record for §A.4; worktree-relative, dies with the t1067 tree. The deciding line is carried verbatim in §A.3 of this file.

---

**Author**: lane agent-20 (t1068) · plan phase · Tier M (3 artifacts: spec.md, plan.md, acceptance.md)
**Owning SPEC for the judgment contract**: SPEC-AC-COUNT-DISCRIMINATOR-001 (unchanged by this work)
**Evidence home**: §A.3 of this file (verbatim outputs); run-phase evidence to `.moai/reports/t1068/` per audit-artifact convention
