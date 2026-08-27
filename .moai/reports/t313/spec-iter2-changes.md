# SPEC-WORKTREE-BASEREF-001 — iteration 2 revision, change summary

Card: t313 · Worktree `.claude/worktrees/t313`, branch `WT-worktree-baseref`, HEAD `48eb945df` (re-read, not inherited).
Input: `.moai/reports/t313/plan-audit-iter1.md` — FAIL, 0.78 harmonic, 7/7 must-pass clear.
Scope: artifacts only. No production code, config, hook, or template was written. No `tab_schema.json` touched.

## Distinct-count measurements (commands as mandated, verbatim output)

```
$ grep -oh 'REQ-WBR-[0-9]\{3\}' .moai/specs/SPEC-WORKTREE-BASEREF-001/*.md | sort -u | wc -l
      16
$ grep -o  'AC-WBR-[0-9]\{3\}'  .moai/specs/SPEC-WORKTREE-BASEREF-001/acceptance.md | sort -u | wc -l
      16
```

Id sets are complete and gapless: `REQ-WBR-001 … REQ-WBR-016`, `AC-WBR-001 … AC-WBR-016`. Every AC id appearing in any of the four artifacts is in the acceptance.md set — no dangling reference. Every one of the 16 REQ ids appears in the §D matrix REQ column (measured: `001 002 003 004 005 006 007 008 009 010 011 012 013 014 015 016`), so REQ coverage is 16/16 and orphaned ACs are 0/16.

## Ceiling note — why no new requirement ids

The obvious shape of the D4 and D6 repairs was two new requirements (a primary-checkout narrowing, a round-trip preservation rule) plus two new ACs. That would have produced 18 REQ / 17 AC, breaching the Tier M ceiling on both axes — `spec-workflow.md:152`: *"A Tier M SPEC may therefore carry up to 16 requirements AND up to 16 acceptance criteria. Exceeding either ceiling is a signal to tier up or to split the SPEC, not to relax the budget."* Repairing one audit finding by creating another is not a repair, so both obligations were folded into the requirements that already own their subject matter: REQ-WBR-004 (the firing-point requirement) became compound and absorbed the narrowing; REQ-WBR-013 (the write-path requirement) absorbed the preservation clause. The two firing-point criteria were likewise merged into one two-half AC. Result: 16/16, exactly at the ceiling, with no obligation dropped.

## Per-defect disposition

### R-D1 — REQ-WBR-004 had no acceptance criterion — FIXED

- `acceptance.md:29` — matrix row `| AC-WBR-016 | MUST | 004 | Firing point: exactly once from the primary checkout, never from a linked worktree |`.
- `acceptance.md:287-315` — the scenario. Half 1 asserts the alignment path's **read** seam (distinct from the `git remote set-head` write seam AC-WBR-003/-005 assert on) is invoked exactly **1** time per `Handle` invocation; zero FAILS (task never registered), two or more FAILS. The verdict's preferred fix was a new AC rather than extending AC-WBR-003, and that is what landed.
- The criterion is explicitly written to fail for an implementation that is behaviourally correct but never wired into `Handle`'s errgroup (`internal/hook/session_start.go:120-175`) — the exact `moai doctor`-only implementation the verdict named.

### R-D2 — eight MUST criteria passed vacuously — FIXED (one edit, all criteria)

- `acceptance.md:330-338` — new first bullet of `§D.3 Definition of Done`: *"Every `-run` invocation in this file must report at least one executed test. A `go test … -run '<regex>'` run whose output contains `[no tests to run]` FAILS the criterion it was run for, regardless of its exit code."*
- Binds by name: AC-WBR-003, -004, -005, -006, -010, -011, -012, -015, -016.
- Mechanically pinned as requested: every `-run` invocation carries `-v`, its output is teed, and `grep -c '^=== RUN'` must be `>= 1`; a `[no tests to run]` line is called out as VACUOUS. Both the `=== RUN` count and the exit code are recorded under `.moai/state/verify/t313/`, and a criterion is PASS only when both are satisfied.
- The verdict's own measurement (AC-WBR-011's literal command exiting 0 with `[no tests to run]` on three packages at this HEAD) is quoted inline as the reason, so a later reader cannot mistake the bullet for ceremony.

### R-D3 — AC-WBR-012 verified one third of REQ-WBR-015 — FIXED

- `acceptance.md:195-214` — retitled *"The anti-dead-key regression guard exists and pins all three properties"*. The `Then` is now the explicit three-part conjunction: present in `settings.AllFields()`; present in the rendered console HTML; reaches a consumer.
- Command now targets `./internal/web/... ./internal/hook/... ./internal/cli/...` — `./internal/web/...` added, which is where the guard lives.
- Two additions beyond the literal ask, both from the verdict's A1/A2 notes: the criterion asserts the **guard test's existence** (a named test beside `internal/web/dead_config_guard_test.go`, and the statement that AC-WBR-005/-007's consumer tests do NOT discharge it), and its **failure mode** — verified by mutation: delete the `FieldDef` from `gitStrategyFields()` (`internal/settings/schema_sections.go:160-177`), re-run, observe non-zero, restore. A guard that still passes with the field removed FAILS.

### R-D4 — R1's premise was false; residual risk under-stated — FIXED (both parts)

**(a) Premise corrected.** `plan.md:149-165` — the sentence *"one repository, one key"* is quoted, named false, and replaced. The tracked-file measurement is inline (`git ls-files --error-unmatch .moai/config/sections/git-strategy.yaml` → rc 0), together with the count inversion (one config copy per worktree following its branch; one `refs/remotes/origin/HEAD` per repository), the reason write-only-on-difference does not confine the hazard (lane A's write creates the difference lane B reverses — divergence is continuously re-created), and the external-actor case that makes it repeated even in a single-lane repository. The real steady-state invariant is stated at `plan.md:159`: *"silence holds only while every working tree that runs the alignment step carries the same configured value"*, with the note that this is not a property the SPEC may assume but one it must arrange. Residual risk is stated rather than dismissed at `plan.md:165`.

**(b) Consumer 1 narrowed to the primary checkout.** Implemented as accepted, not re-opened.

- `spec.md:82-84` — REQ-WBR-004 restated as a compound GEARS requirement: the `When … while` primary-checkout clause plus a `while` clause making a linked worktree a total no-op (no read, no write, no output). The discriminant is cited to the existing implementation, measured in this tree: `inGitWorktreeReal` at `internal/cli/session_worktree.go:234-241` compares `git rev-parse --git-dir` against `git rev-parse --git-common-dir`. The block ends by stating explicitly that the narrowing binds consumer 1 only and that consumer 2 is unaffected.
- `plan.md:50-52` — new decision record `D3.1`, so the narrowing is filed as a design decision rather than as a mitigation buried in a risk.
- `plan.md:91` — M2 now gates the helper on the discriminant **before** reading the configured value, reusing `inGitWorktreeReal`'s shape rather than inventing a second test.
- `acceptance.md:287-315` — AC-WBR-016 half 2 is the verifying criterion: read seam 0, write seam 0, stderr empty, nil error, given the exact configuration AC-WBR-005 requires a write for. The pairing is stated so a half-2 test exercising only the empty-value path does not discharge it.
- `acceptance.md:35` — §D.1 preface states that consumer-1 scenarios carry the primary checkout as their implicit `Given` and consumer-2 scenarios carry no such precondition.
- `acceptance.md:325` — matching edge case replacing the mis-analysed "misconfiguration" reading.

### R-A5 — consumer parity asserted in the plan, unbound by the requirements — FIXED

- `spec.md:102` — REQ-WBR-011 extended: "unresolvable" is decided by REQ-WBR-009's predicate, *"implemented **once** as a single shared helper"* and the **sole** resolvability authority for both consumers. The three divergent implementations the verdict named (`git rev-parse --verify`, a `git branch --list` scrape, a local-branch check) are each declared a requirement violation *even when their runtime behaviour agrees*.
- `acceptance.md:134-138` — AC-WBR-008 gains a third assertion, structural as requested: swap the shared helper through the seam mechanism the package already uses for `sessionWorktreeGitWorktreeAdd` (`internal/cli/session_worktree.go:51-53`) and assert it was invoked exactly once with the configured value. The text states that a behavioural-equivalence check does NOT satisfy the assertion, because it passes for a divergent second rule.
- `plan.md:98` — M3 restated: resolvability decided by **calling M2's exported helper**, exposed through a seam so the assertion is testable.

## Debt items

### D5 — AC-WBR-002's grep bound the whole line — FIXED, and made consistent with M1

Chose the value-side narrowing (not the comment prohibition), then aligned M1 to it.

- `acceptance.md:54-62` — the grep is now `grep … | sed 's/#.*//' | cut -d: -f2- | grep -e develop -e main`, with prose stating the prohibition binds the value, not the comment.
- `plan.md:83` — M1 now says the comment MAY name `main` and `develop` as examples and that only a shipped **value** naming a branch is forbidden.
- Measured, in this worktree, that the pipeline actually discriminates:

```
$ printf 'worktree_base_branch: ""  # e.g. main, develop; empty = no action\n' > /tmp/wbr1.yaml
$ grep 'worktree_base_branch' /tmp/wbr1.yaml | sed 's/#.*//' | cut -d: -f2- | grep -e develop -e main; echo "rc=$?"
rc=1                      # comment naming both branches → PASSES the AC

$ printf 'worktree_base_branch: "develop"\n' > /tmp/wbr2.yaml
$ grep 'worktree_base_branch' /tmp/wbr2.yaml | sed 's/#.*//' | cut -d: -f2- | grep -e develop -e main; echo "rc=$?"
 "develop"
rc=0                      # value naming a branch → FAILS the AC
```

### D6 — AC-WBR-014 traced to `R4 / G3` — FIXED (promoted, then folded)

- `spec.md:108` — REQ-WBR-013 absorbs the preservation clause naming the three keys (`manual.develop_branch`, `manual.release_branch_prefix`, `manual.rc_version_format`, at `.moai/config/sections/git-strategy.yaml:15-17`, absent from `ModeProfile` at `internal/config/types.go:106-132`). The verdict's preferred disposition (promote to the requirement layer) was taken; a standalone REQ-WBR-018 was folded into 013 for the ceiling reason above.
- `acceptance.md:27` — matrix REQ column `R4 / G3` → `013`; severity SHOULD → MUST. This is what removes the only orphaned AC.
- `acceptance.md:251, 261` — heading and closing note updated.
- `acceptance.md:331` — the §D.3 "single SHOULD criterion" bullet replaced: the set is now all-MUST, and an AC-WBR-014 failure escalates as a blocker.
- `plan.md:137` — write list gains the missing file: `internal/settings/sectionapply_test.go` (or a `gitstrategy_roundtrip_test.go` sibling), M5. Done as the verdict directed ("add it either way").
- `plan.md:177` — R4 updated: the risk entry survives because the underlying `ModeProfile` gap is unrepaired, but preservation is now required and verified.

### D7 — citation drift at REQ-WBR-008 — FIXED

- `spec.md:92` — `internal/hook/session_start.go:168-171` → `:176-181`, now agreeing with plan.md §D3. Re-measured in this tree: the contract sentence *"Handle never returns a non-nil error from these steps"* begins at `internal/hook/session_start.go:176`.

### D8 — AC-WBR-009 pinned a check name no requirement fixed — FIXED

- `spec.md:106` — REQ-WBR-012 now fixes `DiagnosticCheck.Name` to exactly `Worktree Base Branch`, citing the exact-match filter measured at `internal/cli/doctor.go:232` (`if filterCheck != "" && c.name != filterCheck`), and states the name is part of the contract rather than an implementation detail.

### D9 — AC-WBR-013 carried a judgement step and an uninterpretable loop — FIXED

- `acceptance.md:216-249` — both non-binary checks replaced by diff-scoped ones against `BASE=$(git merge-base HEAD origin/develop)`. Check (1) enumerates changed `.claude/` / `.moai/` paths in this diff (excluding `.moai/specs/`) and reports `NO-TEMPLATE-COUNTERPART` per path whose template counterpart this diff did not also change — no human judgement. Check (2) enumerates only hook wrappers this diff touched on either side and reports `DRIFT` per non-identical pair. The vacuous-by-construction reading is stated explicitly: with no hook wrapper in scope, check (2) enumerates nothing, which is the intended pass, because the criterion is that this change introduces no drift.
- `git merge-base HEAD origin/develop` was run in this worktree and resolves (`48eb945df606eea7d6d3d1b9a1020adbfe79b2e6`), so the commands are runnable as written.

### D10 — `grep -n -A0` — FIXED

- `acceptance.md:54` — `-A0` dropped.

## Artifact-level bookkeeping

- `spec.md:4` — `version: "0.2.0"` → `"0.3.0"`; `updated: 2026-08-27` unchanged (same date).
- `spec.md:27` — HISTORY row for 0.3.0 recording the verdict, the five blocking repairs, the ceiling-driven folding decision, and the new counts.
- `progress.md:5-12` — `plan_audit_iter` now records iteration 1 complete with its verdict and path; counts updated to 16/16 with the ceiling note; an `iter2_repairs` line enumerates D1/D2/D3/D4/A5 and the folded debt; an `iter2_changes` line points at this file. `blocker: none` retained — no repair proved unimplementable, so nothing was written into plan.md §B as a new blocker.

## What was deliberately NOT changed

- The design. D1-D4 operator decisions (both-consumers scope, SessionStart + doctor surfacing, `TypeText` widget, t316 boundary) are untouched; the verdict confirmed all four faithfully implemented.
- `plan.md` §B gaps G1-G7 — preserved verbatim, none deleted.
- Section structure of all four artifacts.
- Both `tab_schema.json` copies (card t316) — not read, not written.
- No production code, config, hook, or template file.

## Residual risk in this revision

- **Unmeasured**: AC-WBR-013's two diff-scoped shell pipelines were reasoned through and their `BASE` resolution was measured, but the pipelines themselves were not executed end to end against a populated diff — this SPEC has no implementation diff yet, so there is nothing for them to enumerate. A run-phase actor should expect to adjust quoting, not semantics.
- **Judgement, not measurement**: the choice to fold two obligations into REQ-WBR-004 and REQ-WBR-013 rather than add ids is a reading of the Tier M ceiling rule. An auditor who prefers two extra ids over a compound requirement would score the compound REQ-WBR-004 lower on Clarity; the obligation itself is present either way.
- **Untested by construction**: AC-WBR-016's read-seam call-count assertion presumes run-phase implements the read path as a seam. `plan.md:93` states that requirement, but no existing seam carries it — an implementer using a direct `exec.Command` would find the criterion unsatisfiable and would have to refactor rather than merely add a test.
