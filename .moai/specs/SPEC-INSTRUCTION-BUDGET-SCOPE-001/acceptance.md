---
id: SPEC-INSTRUCTION-BUDGET-SCOPE-001
title: "Acceptance criteria — instruction-budget scope alignment and four-file reduction"
version: "0.9.0"
created: 2026-09-28
---

# Acceptance Criteria — SPEC-INSTRUCTION-BUDGET-SCOPE-001

**8 acceptance criteria** against the Tier M ceiling of 16 — unit: **logical criteria**, `AC-IBS-001` … `AC-IBS-008`. The count fell from 16 to 8 in v0.5.0 by folding: seven dispositions are recorded and none of those was dropped; one disposition is unrecorded and unrecoverable, so "nothing was dropped" is not claimable — see § Folding record.

**Two counts exist for this file, and they measure different things.** The repo's `ac-baseline-guard` pre-commit hook counts **distinct live AC identifiers including a letter suffix** and reports **10** here, because `AC-IBS-002a` and `AC-IBS-002d` appear as prefixed identifiers in their own right. Both numbers are correct under their own unit; neither is the other's error.

Which one binds depends on the consumer, so both are stated rather than one being picked:

| consumer | unit | value | command |
|---|---|---|---|
| Tier M ceiling (`spec-workflow.md` § SPEC Complexity Tier) | logical criteria | **8** | `grep -ohE 'AC-IBS-[0-9]{3}' acceptance.md \| sort -u \| wc -l` |
| CHANGELOG AC count at sync (`manager-develop-prompt-template.md` § B12, which names this file as SSOT and counts live identifiers) | live identifiers incl. suffix | **10** | `grep -ohE 'AC-IBS-[0-9]{3}[a-z]?' acceptance.md \| sort -u \| wc -l` |
| `ac-baseline-guard` pre-commit hook (`scripts/ac-baseline/check-staged.sh`) | live identifiers incl. suffix — same unit as the row above | **10** | the hook's own counter; it printed `COUNT 10` on every commit of this card |

[HARD] **If `ac-baseline-guard` is ever armed for this SPEC, the recorded baseline takes the guard's unit — `10`, never `8`.** The guard is a maintained repo mechanism (`scripts/ac-baseline/{check-staged,check-armed,install-hook}.sh` with `internal/spec/ac_baseline_commit_guard_test.go` and `ac_baseline_check_armed_test.go` behind it), and it counts identifiers, not logical criteria. It reported rather than blocked throughout this card only because this SPEC is absent from `.moai/reports/t338/ac-count-baseline.txt` — which is exactly what its `unrecorded (report only)` line means — and because `acceptance.md` carries no counter-delimiter markers.

The marker claim needs its predicate stated, because a substring count no longer answers it — **this paragraph mentions the token, so `grep -c "MOAI-AC-COUNTER"` now matches its own description of the guard.** The guard does not use a substring test: `scripts/ac-baseline/check-staged.sh:124-125` compares a **whole stripped line** for equality (`s == "# MOAI-AC-COUNTER-BEGIN"`), so a backticked mention inside prose cannot arm it. The measurable predicate is therefore an exact-line one, and it is `0`:

```
grep -cxF '# MOAI-AC-COUNTER-BEGIN' acceptance.md
```

A check whose measurement its own documentation changes is the hazard REQ-IBS-004 names in another form; here it is recorded rather than worked around.

The trap is the direction of the mistake. Whoever arms it will reach for the number these artifacts emphasise, and that number is `8` — the Tier figure. Record `8` and the guard fires on the very next commit, reporting it as **an AC was deleted** when nothing changed at all: the count did not move, the unit did. A false deletion alarm is a costly shape to diagnose, because the honest reading of the message sends the reader looking for a removal that never happened.

**The Tier ceiling, by contrast, has no mechanical enforcer.** `16/16` is doctrine the plan-auditor applies by reading. The config key exists but is explicitly unread — `internal/config/loader.go:331` carries `"plan_audit_tier_ceilings": true, // prose-consumed by the plan-auditor agent body; no Go reader` (verified verbatim in this tree, this run). So the v0.5.0 folds were never relieving mechanical pressure; they stand on the merits stated in § Folding record, and the `8` they produced is a doctrinal figure with no gate behind it. The `10` is the one a machine actually produces.

Recorded at plan phase because the sync-phase consumer would otherwise read `8` from `progress.md`, count `10` with its own convention, and have to adjudicate a discrepancy nobody had flagged. This is the fourth unit mismatch in this card — after line-vs-file reference counts (twice) and `grep -c` lines vs occurrences — which is why the unit is now stated beside every count rather than left to the reader.

## Document-level measurement pin

[§2.1] Every RED-now cell below is measured on the **branch base, tree SHA `088594d6b`**, in the worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1180`, branch `WT-rules-40k-split`. This document-level pin binds every criterion carrying no pin of its own, per `.claude/rules/moai/development/verification-completeness.md` §2.1.

The pin is the **SHA**, not a branch name, and the SHA is what makes the cells re-checkable: the measurements were taken before any commit landed on this branch, and the plan-artifact commit that followed touched only files under `.moai/specs/SPEC-INSTRUCTION-BUDGET-SCOPE-001/` — none of the four target files, so every RED-now figure below still describes the tree a reader can inspect at `088594d6b`.

`088594d6b` was `develop`'s tip when this worktree was created; it is an **ancestor** of `develop` now, not equal to it. The target files were measured unchanged at each `develop` tip read during this card (plan.md §A carries the command and the empty results), so the pin is a live measurement of those files while a branch name for them would already have expired. Re-measurement in the merged tree at the integration window is a standing obligation, not a contingency — see plan.md §G on why a moving reference in prose carries an invisible expiry.

Two measurement conventions bind every row:

- **Character counts** use the single-invocation form, never a pipe or a redirect: `python3 -c "import io;print(len(io.open('<path>',encoding='utf-8').read()))"`. `wc -c` is bytes and is inadmissible — these files are Korean-and-English mixed, so bytes exceed characters by 200-800, larger than three of the four overages.
- **No `go run`.** Where a deciding command's predicate is an exit code, it uses the `go test` form or a built binary. `go run` collapses every nonzero exit to 1, so an exit-code predicate behind it measures the wrapper's coarsening rather than the tool.

## Criterion classes

Three classes, because most rows cannot be red before the work exists and saying otherwise would be false:

| Class | Meaning | Rows |
|---|---|---|
| **release-blocking** | Carries a RED-now cell observed red on this tree, plus a green path. Blocks closure. | AC-IBS-001, 002, 003 |
| **regression-guard** | Green at arrival; asserts that the work did not break something already true. Per §2.1 it is **not recorded as a pass** and does not gate closure — it gates *regression*. | AC-IBS-004, 005, 006, 007 |
| **process check** | Verifies the evidence record itself, not the tree. | AC-IBS-008 |

A regression-guard failing IS blocking. The class says only that its green establishes nothing on its own.

---

## D. AC Matrix

### AC-IBS-001 — doctrine scope amendment [release-blocking]

**Given** the amended `coding-standards.md` § File Size Limits, **when** it is read, **then** (a) the budget sentence no longer limits itself to files that load at every session launch and instead names every instruction file the InstructionsLoaded hook measures, AND (b) the "Move detailed content to path-scoped rules" remedy bullet carries a qualifier stating that a `paths:`-scoped destination is itself subject to the same budget.

- **Deciding command (a)**: `grep -c "also loads in full at every session launch" .claude/rules/moai/development/coding-standards.md`
- **RED-now**: stdout `1`, exit `0`. Red because the row requires `0`. Red for the right reason: the limiting clause is present and is the defect REQ-IBS-001 removes.
- **Green path**: M1. Passing output becomes stdout `0`, and `grep -c "InstructionsLoaded" <same file>` ≥ 1.
- **Condition (b)** is reviewer-visible and single-line: read the remedy bullet, the qualifier is in the same bullet. Without it the amended section contradicts itself, which is the whole of REQ-IBS-002.

### AC-IBS-002 — the four files under budget, and the route by which they get there [release-blocking]

**Given** each of the four over-budget files, **when** its character count is measured, **then** the value is `< 40000` — and **the route taken to get there satisfies every sub-condition attached to that file's arm below.**

This is the SPEC's only budget gate. Everything folded into it is a constraint on *how* the gate may be satisfied, not a gate of its own — see § Folding record for why each was folded rather than left standalone.

#### Arm 002a — `spec-workflow.md` (compression-first)

- **Deciding command**: `python3 -c "import io;print(len(io.open('.claude/rules/moai/workflow/spec-workflow.md',encoding='utf-8').read()))"`
- **RED-now**: stdout `40797`, exit `0`. Red because 40,797 > 40,000. Red for the right reason: the file is over the budget the hook enforces at `internal/hook/instructions_loaded.go:103`, and it is the file that fires most (216 of 272 observed over-budget events).
- **Green path**: M3, by duplicate removal or in-place compression. 797 characters to recover.
- **Sub-condition 002a-i — no relocation.** No content is relocated out of this file into a new file. Compression-first is lead direction, not a structural finding (spec.md §4). Decided by the companion roster in 002-R below, not by `git diff --stat` — see § Folding record, fold 3.

#### Arm 002b — `worktree-integration.md` (splittable)

- **Deciding command**: `python3 -c "import io;print(len(io.open('.claude/rules/moai/workflow/worktree-integration.md',encoding='utf-8').read()))"`
- **RED-now**: stdout `61435`, exit `0`. Red because 61,435 > 40,000 — 53% over, the largest overage of the four.
- **Green path**: M4. 21,435 characters, not plausibly reachable by dedup alone, so this is where the split belongs. Its `paths:` carries three patterns, so a companion scoped to a proper subset is admissible.

#### Arm 002c — `kanban-dispatch-detail.md` (splittable, naming-constrained)

- **Deciding command**: `python3 -c "import io;print(len(io.open('.claude/rules/moai/workflow/kanban-dispatch-detail.md',encoding='utf-8').read()))"`
- **RED-now**: stdout `41034`, exit `0`. Red because 41,034 > 40,000.
- **Green path**: M5. 1,034 characters, by compression, non-rule relocation, dedup within the co-loading family, or a split — the M5 ladder, in that order.

#### Arm 002d — `session-handoff-examples.md` (foreclosed from splitting)

- **Deciding command**: `python3 -c "import io;print(len(io.open('.claude/rules/moai/workflow/session-handoff-examples.md',encoding='utf-8').read()))"`
- **RED-now**: stdout `41615`, exit `0`. Red because 41,615 > 40,000.
- **Green path**: M6. 1,615 characters, by compression, non-rule relocation, or dedup within the co-loading family — the reopened rung 3 — and never a split: its `paths:` has one pattern, so the only proper subset is empty and no companion can save anything (spec.md §1; plan.md rung 4 remains "never M6", and REQ-IBS-007 still binds).
- **Sub-condition 002d-i — no relocation into a new rule file.** Same decider as 002a-i. Relocation to a destination outside `.claude/rules/` is NOT a violation of this sub-condition, because such a destination is not a companion — see M6's precondition on whether the hook measures it at all.

#### Sub-conditions on every arm

- **002-H — `[HARD]` clause preservation.** Every `[HARD]` clause present in a modified file before the change is present afterwards at a stated destination, evidenced by a per-clause transfer table in `progress.md` §E.2: one row per clause, source location (file + pre-change line) and destination location (file + post-change line), none marked dropped. A before/after `grep -c '\[HARD\]'` equality is **explicitly not sufficient** — an equal count is compatible with one clause lost and another duplicated. The token-count sum (`4 + 21 + 8 + 12 = 45` `grep -c` hits, not re-measured in this run and carried from M0) may be cited as corroboration; citing it alone fails this sub-condition.
- **002-A — trigger affinity, per relocated section.** Each relocated section's transfer-table row names which of the parent's patterns it is needed on, which it is not needed on, and why; and the claimed "needed on" pattern is one the companion actually carries. A companion-level summary in place of per-section rows fails this.
- **002-A-falsifier.** For each relocated section, the section body does not mention, as a literal string, any path matching a pattern it declared itself NOT needed on. A section mentioning `manager-lead.md` cannot declare `**/.claude/agents/moai/manager-lead.md` as not-needed-on. This is the only condition in the SPEC that can **falsify** an affinity claim rather than check it for internal consistency, which is why it is here — without it, the mutant "declare every section's affinity as a single pattern" satisfies 002-A while violating REQ-IBS-014/015. Stated limitation: literal path mention is a **weak proxy** for affinity, so this is an additional gate and never a sufficient condition. Measured for proportionality, **with the denominator stated**: `worktree-integration.md` carries H2 sections of which exactly **1** — the preamble — mentions literal paths from two trigger domains. The previous wording gave the numerator alone ("only the preamble"), which is not a proportion: `1` of `2` and `1` of `40` share that numerator and support opposite conclusions about whether the falsifier is proportionate. The denominator is measured at M4 and recorded there with its command (`grep -c '^## ' .claude/rules/moai/workflow/worktree-integration.md`); until then this is **1 of an unmeasured total**, and the proportionality claim is a **Gap**, not a finding.
- **002-S — span-affinity sections stay.** Any section whose affinity spans two or more of the parent's patterns remains in the parent and is reduced by compression or not at all. No "recorded N/A" escape: a sub-condition of a row that must go green does not need one.
- **002-R — companion roster.** M2 records, in `progress.md` §E.2 before any content moves, the exact filename of every companion it authorises. The set of new files under `.claude/rules/moai/workflow/` after the change equals that roster — set difference empty in both directions. This is what makes 002a-i and 002d-i attributable: a `git diff --stat` cannot say which parent a new file came from, and a legal `worktree-integration.md` companion appears in the same directory as an illegal one would.
  - Decided by: `find .claude/rules/moai/workflow -maxdepth 1 -name '*.md' | sort` compared against `git ls-tree --name-only "$CARD_BASE":.claude/rules/moai/workflow/ | sort`, with the difference read against the roster, where **`CARD_BASE` is resolved at read time**:

    ```
    CARD_BASE=$(git merge-base develop HEAD)
    ```

  - **Pre-merge-only, and the control is mandatory.** See § Card-base resolution below: this decider is evaluated before the card merges into `develop`, and its non-empty control is run before its result is read.
- **002-P — proper subset.** Every companion's `paths:` pattern set is a proper subset of its parent's: every companion pattern appears in the parent, AND at least one parent pattern is absent. Both failure directions are checked — a companion pattern absent from the parent (scope widened) and an empty complement (equals the parent; the degenerate same-glob shard with zero saving on every trigger). Evidenced by an enumerated table, one row per parent pattern marked carried/not-carried, plus the named complement. A proper subset necessarily **overlaps** the parent on the patterns it carries; that overlap is the saving mechanism (spec.md §1), so an overlap check is NOT part of this condition and must not be added to it.
- **002-N — companion naming.** No companion's filename matches a filename-shaped pattern in its parent's `paths:`. For a companion of `kanban-dispatch-detail.md` that means not matching `kanban-dispatch*`. Decided by `python3 -c "import fnmatch,sys;print(fnmatch.fnmatch(sys.argv[1],sys.argv[2]))" <companion-basename> 'kanban-dispatch*'` → `False`. The family census this defends is in spec.md §1: **three** files already match that pattern at base, and one of them (`kanban-dispatch-mechanics.md`) is the realised form of exactly this trap.
- **002-B — companions are themselves under budget.** Each companion measures `< 40000` by the same single-invocation command. A reduction that pushes the parent under while leaving the companion over satisfies nothing.
- **002-G — registry enrolment.** Each companion's path appears in `sanitizedPairPaths` (`internal/template/sanitized_pair_parity_test.go`) and `workflowOptMirroredPaths` (`internal/template/rule_template_mirror_test.go`). Decided by `grep -c "<companion path>" <each file>` → ≥1.
  - **Two registries, not three — measured, no longer a structural read.** `internal/template/template_neutrality_audit_test.go` walks the whole template tree, so a new mirrored `.md` is covered without enrolment. The sentinel demonstration (plan.md §B.4) injected a `C1-macos-bias-path` violation into the mirror of `spec-workflow.md` and the anchored audit reported it by path — `TEMPLATE_NEUTRALITY_VIOLATION: class=C1-macos-bias-path file=.claude/rules/moai/workflow/spec-workflow.md`, exit `1`, green again at exit `0` after revert. The premise this sub-condition rests on is therefore executed rather than inferred.

Sub-conditions 002-R / 002-P / 002-N / 002-B / 002-G / 002-A / 002-A-falsifier / 002-S apply only where a companion is created, and are recorded as **not-applicable-because-no-companion** where none is. That record is itself required: silence is not the same statement.

### AC-IBS-003 — the four `paths:` globs are pinned by a mechanical guard [release-blocking]

**Given** the four files' `paths:` frontmatter, **when** the guard test runs, **then** each matches the value recorded in the test source.

- **Deciding command**: `go test -run '^TestWorkflowRulePathsPinned$' ./internal/template/...`
- **RED-now**: the guard does not exist. Evidenced by `grep -rl "TestWorkflowRulePathsPinned" internal/template/` → **empty stdout, exit `1`** (measured this run).
  - **The command is `-rl`, not `-rc`, and the earlier cell named the wrong one.** `grep -rc` prints a `<file>:0` line for **every** file it scans and exits `0`, so the previous cell's stated output ("no file matches … exit `1`") was not what its own command produces and the exit code it claimed was the opposite of the truth. Measured: `grep -rc …` prints `internal/template/project_continuation_pipeline_signal_test.go:0` and one such line per scanned file. `-rl` lists only matching files, so absence is an empty stdout and exit `1` — which is the observation this cell needs. Recorded rather than quietly swapped, because a RED cell whose output does not match its command is the §2.1 defect this SPEC disciplines elsewhere. **The deciding command is deliberately NOT the RED-now evidence**: an anchored selector matching zero tests exits `0` and prints `ok … [no tests to run]`, so running it now would produce a green that asserts nothing — the empty-sweep shape of `verification-completeness.md` §1.1. Absence is evidenced by the grep, not by the selector.
- **Green path**: the guard is added in M2 and asserts the four recorded globs. Passing output is `ok internal/template`, exit `0`.
- **Why recorded values and not `git show 088594d6b:<path>`**: REQ-IBS-010 preserves the globs *unless this SPEC is amended to argue a change* — a this-card constraint, not an invariant. A tree-pinned guard would outlive the requirement motivating it and become an unattributable stale guard after close. Recording the values in the test source makes a future glob change deliberate (edit the test, visible in review) rather than forbidden, which is the same idiom `sanitizedPairPaths` and `workflowOptMirroredPaths` already use. Provenance — that the recorded values came from `088594d6b` — lives in this document's pin and a source comment, which is the correct home for a statement about how values were obtained rather than an assertion a test can check.
- **The anchored selector is load-bearing, not cosmetic.** `TestWorkflowRulePathsPinned` unanchored would also select any longer name, making a pass unattributable to the test named. The repo's own lint flags the unanchored form as `VacuousTestAssertion`.

### AC-IBS-004 — every cross-reference into a modified file still resolves [regression-guard]

Green at arrival — references resolve today — so this asserts that the work did not break them. Three sub-conditions, each with its own baseline and its own command, because the previous version cited a baseline measured by a command it did not name.

- **004a — `file.md#anchor` links.** Enumerate with `grep -rlE '<filename>\.md#' .claude .moai/docs internal/template/templates`; for each, confirm a heading in the target whose GitHub-style slug equals the fragment. Zero unresolved.
  - Baseline at branch base `088594d6b`, unit **files**, by that command: `spec-workflow.md` **16**, `worktree-integration.md` **0**, `session-handoff-examples.md` **0**, `kanban-dispatch-detail.md` **0**.
  - **004a is inert on every file this SPEC splits, and live only on the one it does not.** The zeros land on `worktree-integration.md` (M4) and `kanban-dispatch-detail.md` (M5) — precisely the two files where a split happens and where a broken anchor could actually be introduced. The 16 lands on `spec-workflow.md`, which is compression-first and gets no split at all.
  - **Recording rule, per target**: record the swept count first, then the verdict. Where the count is `0`, the verdict is **`N/A (0 references measured)`** — never PASS. A PASS on a zero sweep is the empty-swept-set claim of `verification-completeness.md` §1.1, and here it would be recorded against the two highest-risk files in the SPEC.
- **004b — `§`-style section references, both orders.** Forward: `grep -rlE '<filename>\.md[^)]{0,200}§' …`. Reverse: `grep -rlE '§[^)]{0,60}<filename>\.md' …`. For each hit, `grep -c "<section title>" <target>` ≥ 1.
  - Baseline at `088594d6b`, by the forward command at the 0-40 bound previously used: **44 / 22 / 8 / 7**. This corrects a v0.4.0 defect in which the row cited `78 / 38 / 8 / 7` — a figure produced by a *different* command (bare filename mention, § Reference-count scope note) — while naming the `§` command as decider. The mismatch was masked because two of four cells agree under both commands, and they are the two small numbers a reader checks first.
  - The bound is widened from 40 to 200 characters because 40 missed 4 `§` references for `spec-workflow.md` (74 at 0-200 vs 70 at 0-40). Any reference beyond 200 characters of separation is an explicit **Gap**, not a claimed pass.
  - The reverse order is added because **11 references** across the corpus take the form `§ … <filename>.md` and matched neither 004a nor 004b before this amendment. A row asserting "every one resolves" whose commands sweep part of the reference surface is the defect; the count was never the defect.
- **004c — parent pointers.** For each section relocated into a companion, the parent carries a pointer naming **both** the companion file and the relocated section's heading text, so an inbound `§` reference to the parent still leads a reader to the content.
  - **Two conditions, both required, checked per relocated section**: (i) `grep -n "<companion filename>" <parent>` returns a line; (ii) **that same line** also contains the relocated section's heading text. A filename-only `grep -c` satisfies (i) and leaves every inbound `§ … <parent>` reference to that section dangling — the pointer says "some of this moved" without saying which part or where. Decided by reading the matched line, not by a count.
  - **004c carries more weight than the row's symmetry suggests.** For `worktree-integration.md` and `kanban-dispatch-detail.md`, 004a is `N/A (0 references measured)`, so anchor resolution for the two split targets rests on 004b and 004c alone. Of those, 004b was patched for two blind spots mid-plan (§ Reference-count scope note). That makes 004c the least-patched leg under the heaviest load, which is why its check is specified to the line rather than to the file.
- **Scope, declared rather than left open**: a bare filename mention carrying neither `§` nor `#anchor` resolves unless the file is renamed, and no file is renamed under this SPEC, so bare mentions are **out of AC-IBS-004's scope**. Archived SPEC prose under `.moai/specs/_archive/**` is likewise out of scope — a historical record, not repaired.

### AC-IBS-005 — local and template mirror stay byte-identical [regression-guard]

**The property is "no NEW divergence", not byte-identity.** The row guards against a **one-sided edit** — that is the Template-First failure this repo has had before — and byte-identity is only a proxy for it. The proxy is exact for four of the five pairs and provably wrong for the fifth, so the row splits.

**005a — the four workflow files: byte-identical.** **Given** each of the four target files, **when** compared to its `internal/template/templates/.claude/` counterpart, **then** the pair is byte-identical. Decided by `diff <local> <mirror>` per file → exit `0`, no output. Green at arrival, verified by `diff -q` in this run. M1 does not touch these four; M4-M6 do, and this is the row that catches a one-sided reduction.

**005b — the `coding-standards.md` pair: the diff's divergence is unchanged from its pre-M1 baseline.** **Given** the pair, **when** diffed, **then** the output consists of exactly one hunk whose single content line equals the baseline content line recorded below, byte for byte — no hunk added, none removed; only the hunk header's line numbers may move, because a correct M1 amendment above line 141 shifts them. This invariant is the predicate; the recorded baseline as a whole (header numbers included) is its evidence, not itself the criterion.

- **Deciding command**: `diff .claude/rules/moai/development/coding-standards.md internal/template/templates/.claude/rules/moai/development/coding-standards.md`
- **Baseline, measured this run** — exit `1`, two lines of output, one content line:

```
141d140
< - `git commit --no-verify` — bypasses the relocated pre-commit quality gate (the harness safety net at the commit tier; enforced mechanically by the PreToolUse guard at `internal/hook/pre_tool.go` per SPEC-PRETOOL-GATE-MOVE-001 REQ-PGM-006 / F5)
```

- **Why this pair is exempt from byte-identity, permanently.** That line is local-only **by design**: it carries a SPEC ID and REQ tokens, which the template-neutrality `C1-spec-id` class excludes from the distributed template. So the divergence is intended, it is not M1's to repair, and no correct M1 edit removes it. A byte-identity criterion on this pair is **impossible-red** in the `verification-completeness.md` §2 sense — red at arrival and red forever, satisfiable by no correct work.
- **Why not an exclusion list of local-only lines.** It would also work today and ages badly: every legitimately-added internal line needs an edit to the list, and a stale list fails **silently** — the one failure mode this SPEC is least willing to ship. Diff-unchanged needs no maintenance, because a new local-only line changes the baseline visibly and deliberately.
- **Positive control — executed at plan phase, and it found a defect in the obvious form of this check.** A one-sided edit (`<!-- t1180 one-sided-edit control -->` appended to the local copy only) made the deciding command print two extra hunks on top of the baseline:

```
141d140
< - `git commit --no-verify` — bypasses the relocated pre-commit quality gate (…)
172,173d170
<
< <!-- t1180 one-sided-edit control -->
```

  Reverted with `git checkout --`; the output returned to the baseline exactly, and `git status --short` showed the tree clean apart from this SPEC's own artifacts.

  **[HARD] The exit code was `1` in all three states — baseline, injected, and reverted — so the verdict is the OUTPUT comparison and never the exit code.** This is the finding the control exists to produce: `diff` exits `1` for *any* difference, so a check written as "the diff still exits 1, as it did at baseline" passes an injected one-sided edit without noticing. That is precisely the one-sided edit this row is here to catch, and it is the shape a later reader would most plausibly simplify the row into. The predicate is **byte equality of the diff output against the recorded baseline**; the exit code is recorded as context and carries no verdict.

  Run again at M1 after the doctrine amendment lands: M1 may shift the hunk header's line numbers, and **re-recording the baseline at M1 REQUIRES demonstrating the content line unchanged and no hunk added or removed** — recording whatever the diff then says, unverified, would satisfy a byte-equality-against-baseline reading while permitting the one-sided edit this row exists to catch. That is the mutant the invariant above closes: a one-sided edit necessarily changes the content line or the hunk count, so it cannot satisfy the predicate under any re-recorded header.
- **Scope note.** 005b is the ONLY pair leaving the byte-identity clause. Any further file this SPEC creates under `.claude/` enters 005a, not 005b; an exemption is earned by a measured, explained, permanent divergence, never assumed.

### AC-IBS-006 — the template and hook packages pass [regression-guard]

Decided by `go test ./internal/template/...` → exit `0` (full package, per the verification scope in spec.md §4 — template mirror content is guard-tested there), and `go test ./internal/hook/...` → exit `0`.

**Not measured in this run.** Both are heavy package suites, and `.claude/rules/local/gitflow-lane-protocol.md` §8 requires a resource-slot lease before a heavy run rather than a lane starting one unilaterally. Recorded as a Gap at plan phase; measured in run phase under a lease. The hook row guards against an accidental edit to `instructions_loaded.go` and against a guard in that package keyed on the doctrine text.

### AC-IBS-007 — scope containment [regression-guard]

**Given** the branch diff, **when** inspected, **then** `internal/hook/instructions_loaded.go`, `CLAUDE.local.md`, and `.moai/reports/t1180/verdict.md` are all absent from it. Decided by `git diff --name-only "$CARD_BASE"..HEAD` → none of the three appears, with **`CARD_BASE` resolved at read time** and the non-empty control run first — see § Card-base resolution.

Green at arrival, and the reason has to be stated carefully now that the branch carries the plan-artifact commit: it is green because that commit touched only files under `.moai/specs/SPEC-INSTRUCTION-BUDGET-SCOPE-001/`, not because the branch is empty. The earlier wording ("green because the branch carries zero commits") was true when written and expired the moment the commit landed — an instance of the plan.md §G hazard inside the row that checks scope. The three paths are the SPEC's declared non-targets: the hook is read-only for this SPEC (spec.md §4), `CLAUDE.local.md` is out of scope with its figures recorded as a follow-up card candidate (spec.md §5), and the verdict file is referenced, never overwritten.

### AC-IBS-008 — the evidence record is attributable [process check]

**Given** `progress.md` §E.2 and every commit on `WT-rules-40k-split`, **when** read, **then**:

- (a) every character-count claim cites its single-invocation `python3` command, and no `wc -c` output appears as a budget figure — `grep -c "wc -c" progress.md` → `0`;
- (b) every commit message names the card id `t1180` — `git log --format=%s "$CARD_BASE"..HEAD` read for the token, with the count of non-matching subjects being `0`. **`CARD_BASE` resolved at read time**, and the same non-empty control applies: a `0` non-matching count over an empty commit range asserts nothing. (This row was not in the audit's D11 list, which named two sites; it is the same range-predicate class, so it is repaired with them rather than left as the one literal pin the fix missed — flagged here rather than extended silently.);
- (c) every figure in §E.2 names the command that produced it. This condition exists because the v0.4.0 defect in 004b was exactly a figure whose stated decider was not the command that measured it.

## Card-base resolution — read-time, pre-merge-only, control-gated

[HARD] Every range or tree predicate in this document resolves its left end **at read time**, never from a literal SHA:

```
CARD_BASE=$(git merge-base develop HEAD)
```

Per `.claude/rules/local/gitflow-lane-protocol.md` §8 [HARD]. A literal pin stops describing "this card's own contribution" the moment local `develop` is absorbed: another card's commits enter the range, and the predicate reports contact this card never made. Resolved at read time, `merge-base` stays at the last absorbed `develop` commit and keeps answering the intended question however far `develop` advances afterwards.

**These predicates are pre-merge-only, and the reason is that they pass vacuously afterwards.** Once the card merges into `develop`, `merge-base develop HEAD` **becomes the card tip itself**, the range empties, and every range predicate above passes while measuring nothing. This is measured, not predicted: `gitflow-lane-protocol.md` §8 records the empty output from an already-merged card branch (`.moai/reports/t543/repro/limit-a-post-merge-all.txt`). After the merge, the evidence is tree identity between the merge commit and the card branch — not these rows.

[HARD] **Non-empty-range control — run before any range predicate's result is read.**

```
git diff --name-only "$CARD_BASE"..HEAD | wc -l
```

≥ 1 required. Measured this run: **4**.

**Why the control is not optional, and why it is the point of this repair.** Swapping a literal SHA for `merge-base` removes a **false red** and introduces the possibility of a **vacuous green** — the same defect class one step over, and precisely the one this SPEC exists to discipline (`verification-completeness.md` §1.1: a pass whose swept set is empty asserts nothing). A `0` from the control means **not measurable**, never "no contact found"; a row whose control reads `0` is reported as a Gap and its verdict withheld. Without the control this repair would have traded a defect the reader can see for one they cannot.

## Folding record

v0.4.0 carried 16 criteria; this version carries 8. The table below records **7** dispositions (six rows plus one condition), all folded into AC-IBS-002 and none dropped. Each fold has the same justification, applied per row:

| v0.4.0 row | Folded to | Why it was not a gate of its own |
|---|---|---|
| 003 companion under budget | 002-B | Conditional on a companion the work creates — no pre-implementation RED exists, so it could never carry a RED-now cell. |
| 004 proper subset | 002-P | Same. |
| 005 companion naming | 002-N | Same. Its v0.4.0 expected value of `2` was additionally **wrong-reason red at arrival** — the family has three members at base, and no correct work makes it 2. |
| 006 no relocation | 002a-i, 002d-i | The mutant probe settles it: **do nothing at all** satisfies "no content relocated" while both files stay over budget. A criterion the null implementation satisfies is too shallow to adopt; the same mutant dies on the char count. |
| 009 span-affinity stays | 002-S | Conditional on M2's own findings, and carried a "recorded N/A" escape — a second way to pass without asserting. |
| 013 registry enrolment | 002-G | Conditional on a companion. |
| 010 `[HARD]` preservation | 002-H | The null implementation satisfies it trivially. Folding keeps it binding on the route to green while removing a gate that cannot fail on its own. Its prominence is preserved by placement: it is the first sub-condition on every arm. |

**The arithmetic does not close, and the missing disposition is unrecoverable.** `16 - 7 = 9`, not `8`. Seven dispositions are recorded above; the eighth is not, and it cannot be reconstructed — v0.4.0's criterion list was replaced in place rather than amended, so no copy of the 16-row set survives in this branch's history to diff against. What can be stated is bounded: one v0.4.0 criterion left the document in that pass without its disposition being written down, and whether it was folded into an existing row or dropped outright **cannot now be established from the artifacts**.

Recorded as unrecoverable rather than reconstructed, because a plausible reconstruction is indistinguishable from a measured one once written, and this document's whole subject is the difference. The live consequence is bounded too: the 8 criteria now in force each carry their own RED-now cell and deciding command, so the standing criterion set is verifiable on its own terms regardless of how it was reached — an unrecorded disposition damages the audit trail, not the gate.

The effect of the folds is fewer gates, each red today, with the constraints living inside the row they constrain. Headroom against the 16 ceiling is a by-product, not the reason — every recorded fold was ruled on mutant-probe or impossible-RED grounds, and the ceiling turns out to have no mechanical enforcer at all (§ the guard-unit note above).

## Reference-count scope note

Three different counts exist for the same four files, and mixing them caused the v0.4.0 defect. Each is recorded with the command that produces it **and its unit**. A baseline whose unit is unstated is the same defect class as one whose command is unstated, and this row has now produced both.

**Unit: files** (`grep -rl … | wc -l`). Measured at branch base `088594d6b`, this run.

| file | `§`-form forward (004b) | `§`-form reverse (004b) | reverse-only files | `#anchor` (004a) | bare filename (decides nothing) |
|---|---:|---:|---:|---:|---:|
| `spec-workflow.md` | 44 | 1 | **0** | 16 | 78 |
| `worktree-integration.md` | 22 | 2 | **0** | 0 | 38 |
| `session-handoff-examples.md` | 8 | 0 | **0** | 0 | 8 |
| `kanban-dispatch-detail.md` | 7 | 2 | **0** | 0 | 7 |

- forward `§`: `grep -rlE '<filename>\.md[^)]{0,200}§' .claude .moai/docs internal/template/templates | wc -l`
- reverse `§`: `grep -rlE '§[^)]{0,60}<filename>\.md' … | wc -l`
- reverse-only: `comm -23` of the reverse file list against the forward file list, counted
- `#anchor`: `grep -rlE '<filename>\.md#' … | wc -l`
- bare filename: `grep -rlE '<filename>\.md' … | wc -l`

**Both audit blind spots live inside the forward file set — measured, not assumed.** The reverse-only column is `0` for all four files: every file carrying a reverse-order reference also carries a forward-order one, so the reverse form adds **no new files**. The same holds for the bound: widening 40 → 200 gives 44/22/8/7, identical to the 0-40 figures. So at **file** granularity the forward baseline is complete, and both findings are **per-reference completeness inside those files** rather than missing files. The audit's own figures for the same findings (74 vs 70, and 11) are **line** counts and are not comparable to this table — that is the unit mismatch this note exists to prevent.

The practical consequence for 004b: enumerate both orders at the 0-200 bound, because the per-reference gaps are real even though the file set does not grow. Anything beyond 200 characters of separation stays an explicit **Gap**.

**The bare-filename column decides nothing.** It is blast-radius context — 78 files name `spec-workflow.md`, which is why M3 carries the largest anchor-resolution burden — and it is labelled so it is not mistaken for a decider again.

## Traceability

| REQ | AC |
|---|---|
| REQ-IBS-001 — doctrine names the hook's measured set | AC-IBS-001(a) |
| REQ-IBS-002 — section not self-contradictory | AC-IBS-001(b) |
| REQ-IBS-003 — byte-identical template mirror | AC-IBS-005 |
| REQ-IBS-004 — four files under 40,000 | AC-IBS-002 (all four arms), 002-B |
| REQ-IBS-005 — no equal-set companion | AC-IBS-002-P |
| REQ-IBS-006 — proper subset, demonstrated by enumeration | AC-IBS-002-P, 002-R |
| REQ-IBS-007 — one-element parent: removal or compression only | AC-IBS-002d, 002d-i |
| REQ-IBS-008 — every `[HARD]` clause preserved | AC-IBS-002-H |
| REQ-IBS-009 — every cross-reference still resolves | AC-IBS-004 |
| REQ-IBS-010 — the four `paths:` preserved | AC-IBS-003 |
| REQ-IBS-011 — companion enrolled in both registries | AC-IBS-002-G |
| REQ-IBS-012 — `CLAUDE.local.md` untouched | AC-IBS-007 |
| REQ-IBS-013 — companion filename outside the parent's filename-shaped pattern | AC-IBS-002-N |
| REQ-IBS-014 — per-section trigger affinity stated | AC-IBS-002-A, 002-A-falsifier |
| REQ-IBS-015 — span-affinity sections stay in the parent | AC-IBS-002-S |

Four conditions are **constraint-derived** rather than REQ-derived, recorded so no condition is unattributed: AC-IBS-006 guards the verification scope (spec.md §4), AC-IBS-008(a) the character-count measurement unit (spec.md §4), AC-IBS-008(b) the card-id commit convention (spec.md §4), and AC-IBS-002a's compression-first arm the lead direction (spec.md §4). The reduction order is likewise a constraint (spec.md §4, plan.md §F) and deliberately carries no criterion: sequencing is not observable in the closing state, so a criterion asserting it would test the commit log rather than the work.

## Edge cases

- **Exactly 40,000 characters PASSES.** The hook compares `charCount > charBudget` (`internal/hook/instructions_loaded.go:104`), so 40,000 is within budget and 40,001 is not.
- **A companion created then abandoned within the branch**: the companion sub-conditions are recorded not-applicable and `progress.md` records the abandonment with its reason.
- **A relocated section with no single-pattern affinity**: it does not move (002-S). Discovering this mid-M4 is a normal outcome, not a blocker.
- **Compression cannot reach 40,000 without touching a `[HARD]` clause** — the live risk on `session-handoff-examples.md`, already a relocation destination with 1,615 characters to recover. A blocker is the correct outcome only after the M6 ladder is exhausted: compression → relocation outside `.claude/rules/` → dedup within the co-loading family → blocker. 002-H failing together with arm 002d is the signal that the blocker is genuine rather than a report of unexplored options.

## Definition of Done

All three release-blocking rows green with their green paths realised; every regression-guard row green or carried as a named debt with its reason; AC-IBS-006's Gap closed under a resource-slot lease; and `progress.md` §E.2 carrying the four post-change character counts with their single-invocation commands, the per-clause `[HARD]` transfer table with its per-section trigger-affinity column and falsifier result, the companion roster with the enumerated proper-subset comparison and named complement, the companion-filename pattern test, and the three reference-count baselines each naming its own command.
