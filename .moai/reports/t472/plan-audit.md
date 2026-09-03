# Plan Audit — SPEC-TODO-LANDING-ATTRIBUTION-001 (card t472)

Iteration: 1/2 (Tier M ceiling)
**Verdict: PASS-WITH-DEBT**
Aggregate score: **0.8125** against the Tier M threshold **0.80**

Reasoning context ignored per M1 Context Isolation. The lane's reports were read as
*claims to be re-measured*, never as evidence.

---

## Audit provenance (VCI §2.2)

| | Value |
|---|---|
| Tree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t472` |
| Branch | `WT-landed-drift-detect` |
| HEAD at audit start | `62cbfdf77` |
| HEAD at audit end | `a72fb378c` — **moved mid-audit** (see D9) |
| Judging binary | `go build -o <scratch>/moai-audit ./cmd/moai` from **this tree**, rc=0. The installed `~/go/bin/moai` (~190 commits behind) was **not** used for any figure below. |
| Corpus refs | `origin/develop` = `7835148d3`, `origin/main` = `7ad9f8534` (re-read at audit time) |
| SPEC citation tree | `4bcac7079`. Verified: `git diff --stat 4bcac7079 HEAD` touches **only** the four SPEC artifacts (611 insertions, 0 code files), so every `file:line` citation measured at `4bcac7079` still resolves at HEAD. |

---

## MUST-PASS criteria

| | Criterion | State | Evidence |
|---|---|---|---|
| MP-1 | REQ number consistency | **PASS** | `grep -ohE 'REQ-TLA-[0-9]{3}' spec.md` → `001…012` in order, one repeat (the §A.7-referencing line), no gaps, no duplicate definitions, uniform 3-digit padding. |
| MP-2 | GEARS format compliance (**requirement layer** — `REQ-TLA-*` in `spec.md`; the `AC-TLA-*` Given-When-Then entries in `acceptance.md` are the verification layer and were **not** graded here) | **PASS** | 001/002/005/007/010 Ubiquitous (`The … shall …`); 003/004/012 Unwanted (`shall not`); 006/009 Event-driven (`When …`); 008 `Where`-gated on a static-config value (config key present-or-empty is exactly GEARS `Where` = static config, so the "(capability gate)" annotation is correct, not a misuse); 011 compound `While … when … the command shall disclose`. 12/12 conform. |
| MP-3 | YAML frontmatter validity | **PASS** | All 12 canonical fields present with canonical names (`created:`, `updated:`, `tags:`, `id:` — no snake_case alias): `spec.md:2-15`. Plus optional `tier: M` and `related_specs:`. Independently confirmed by `moai-audit spec lint --strict .../spec.md` → `✓ No findings`, rc=0. |
| MP-4 | Section 22 language neutrality | **N/A (auto-pass)** | Single-language SPEC — `module: "internal/kanban, internal/cli"`, no multi-language tooling surface. |
| MP-5 | D7 cross-SPEC reconciliation | **PASS** | 3 referenced SPECs, all present: `SPEC-TODO-LANDING-STATE-001` → `status: completed`; `SPEC-KANBAN-QUEUE-PR-SYNC-001` → `in-progress`; `SPEC-WORKTREE-BASEREF-001` → `completed`. None retired/superseded/archived. No BLOCKING finding. |
| MP-6 | D8 cross-platform discipline | **PASS** | `grep -c 'syscall' spec.md` → `0`. Auto-PASS per D8-4. |
| MP-7 | Clarification gate | **PASS** | `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-TODO-LANDING-ATTRIBUTION-001/` → no output. `research.md` absent (correct for Tier M). |

No must-pass failure. The firewall holds.

---

## Category scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.75 | 0.75 — minor ambiguity in one or two requirements | REQ-TLA-002 pins "exactly three forms", but form 3 (`spec.md` §A.4 row 3, shape column: *"a merge subject naming the card"*) is **not a position** — it is an occurrence test, contradicting REQ-TLA-001's own discriminator. See D1. Everything else is single-interpretation. |
| Completeness | 0.75 | 0.75 — one non-critical gap; frontmatter complete | Six `### Out of Scope — <topic>` H3s, each with specific `-` bullets. HISTORY / WHY (§A) / WHAT (§B) / Constraints (§C) / Exclusions (§D) / Cross-refs (§E) all present. Gaps: `acceptance.md` §D mandates a two-attribution ruling that no criterion supplies (D3); `plan.md` M3 item 2 carries work with no requirement and no criterion (D7). |
| Testability | 0.75 | 0.75 — one criterion not precisely binary | 10 of 12 criteria are binary with a named falsifying fixture. AC-TLA-009's second clause asserts a **value** inequality (`is not DefaultLandedRef`) where it means **provenance**, so it false-fails on any repo whose `origin/HEAD` names `main` (D4). AC-TLA-005's "located by grep … declared in exactly one named symbol or table" needs a judgment call, mitigated by its own removal-mutation clause. No weasel words found ("appropriate"/"adequate"/"reasonable"/"proper" → 0 hits). |
| Traceability | **1.00** | 1.0 | 12 REQ, 12 AC. Every `REQ-TLA-001..012` appears in at least one `(maps REQ-…)` clause in `acceptance.md`; every AC maps to a REQ that exists. Zero orphans, zero uncovered. |

Aggregate = (0.75 + 0.75 + 0.75 + 1.00) / 4 = **0.8125** ≥ 0.80.

---

## The six load-bearing claims — reproduction status

### 1. The ordering decision (F before A) — **REPRODUCED, and robust**

Both sides of the arithmetic reproduce exactly, with my own commands, on this tree.

**Current-ref side (2 landed, both false).**
```
$ <tree-built>/moai-audit todo pr        # 58 lines = 1 notice + 57 rows
```
Exactly two rows carry `landed` in the outcome column: `t237` (queued, line 14) and `t312`
(dropped, line 25). The other three `landed` grep hits are the substring inside card *text*
(t359, t456, t472), not outcome values. Both are false:
```
$ git log origin/main --format='%h|%s' | grep -E '\bt237\b'   → (no output)
$ git log origin/main --format='%h|%s' | grep -E '\bt312\b'   → (no output)
```
Zero subject occurrences of either token on `origin/main`, in **any** position. Precision 0/2 confirmed.

**Counterfactual side (9 of 31 live, 7 false).** Live cards parsed from the same `todo pr` run:
`queued 15 + picked 16 = 31` (dropped 26). Whole-message matching over all 5,837
`origin/develop` commits, emulating `--grep=\btNNN\b`:

> t204(1) t216(3) t237(3) t315(4) t359(2) t401(1) t436(6) t440(5) t443(14) — **9 cards**,
> card-for-card identical to `spec.md` §A.2.

Subject-occurrence census over the same 5,837 subjects:

| Card | Subject occurrences | Attributed? |
|---|---|---|
| t204, t237, t315, t359, t436 | **zero — none at all** | no |
| t216 | `673d3d8a0 docs(t263): … behind t216` | no — t263's |
| t443 | `0d26f8a00 chore(catalog): … t443 jurisdiction (t461)` | no — t461's |
| t401 | `d5caf2d8e feat(SPEC-JUDGMENT-FIRST-MODE-001): … (t401)` | **yes** — trailing paren |
| t440 | `b80cc9cf1 docs(t440): …` + `4c3b1653c Merge card t440 (…) into develop: …` + `63ea8693a … (t440)` | **yes** — scope, merge, trailing |

→ 2 true, **7 false**. Reproduced.

**Over/under-count check (the prompt's specific concern).** The lane's admission that "a third
attribution convention would change the split" is **weaker than the evidence requires**, in the
direction that helps the SPEC. Five of the seven false positives have **zero subject occurrence of
any kind** — no convention operating on the subject line can rescue them. The remaining two are
each demonstrably another card's. The 7/9 split is therefore robust to any fourth *subject-position*
convention. It would only move under a **body-trailer** convention (e.g. a `Card: tNNN` trailer),
which the SPEC does not claim to support and which I did not find in the corpus. The central
structural decision is **sound**.

### 2. The mutant set's non-vacuity — **REPRODUCED, with one gap**

Every fixture exists in this tree's history and has the property claimed of it:

- **MUT-WHOLE-MESSAGE / t237** — `git log origin/main --perl-regexp --grep='\bt237\b' --oneline`
  returns `539349c5b docs(t230): …` and `32d2221fa feat(cli): … (t230)`. Both are t230's. Zero
  t237 subjects. Falsifies.
- **MUT-SUBJECT-ONLY / t216 and t443** — the decisive pair. **Both hold.** `673d3d8a0` is
  attributed to t263 and names t216 mid-subject; `0d26f8a00` is attributed to t461 (trailing) and
  names t443 mid-subject. Neither is the queried card's commit. The naive repair *is* excluded.
- **Positive controls t401 / t440** — both real, both attributed, both would survive a
  position-based predicate.
- **MUT-CONST-REF / MUT-CONFIG-ONLY / MUT-SILENT-FALLBACK** — grounded in
  `git symbolic-ref refs/remotes/origin/HEAD` → `refs/remotes/origin/develop` (re-run, rc=0) and
  the primary checkout's empty key. Each is falsified by an AC clause.

**A criterion CAN pass a mutant it is meant to defeat.** Consider `MUT-MERGE-ANY-TOKEN` — an
implementation whose form-3 rule is exactly what REQ-TLA-002 says, "a merge subject naming the
card". It passes AC-TLA-001 through AC-TLA-007 unchanged, because AC-TLA-003 exercises only the
`(card t263)` merge-into-develop shape and the t216 non-merge counter-case. It never tests a merge
subject naming a card that is not the merge's own. Measured population in-corpus (D1 below): 4 such
merges naming two distinct cards, 26 merges **into** a `WT-` branch that name a card, 50 merges
naming a card in a shape the enumeration does not pin. Two swapped-form errors in the fixture
declaration are D5.

### 3. The bidirectional-regression obligation — **REPRODUCED (satisfied)**

Every predicate-behaviour criterion on axis F carries both directions:
AC-TLA-001 (t440 green / t237 red), -002 (t401 green / t443 red), -003 (t263 green / t216 red),
-004 (body-only red / scope green), -007 (`unknown` vs `not-landed` kept distinguishable).
**No criterion asserts only the green direction**, so no implementation that reports everything as
landed can pass. AC-TLA-005 and AC-TLA-006 are structural rather than behavioural and carry a
mutation direction instead — the preamble's blanket "every axis-F criterion asserts green AND red"
over-reaches over those two (D8, MINOR), but the protection is intact.

### 4. Axis A option A3 and its rejections — **REPRODUCED, with one wrong citation and one overstated residual**

**A3's premise holds.** `git symbolic-ref refs/remotes/origin/HEAD` → `refs/remotes/origin/develop`,
rc=0, this tree, this run. Level 2 would answer `develop` today.

**A4's rejection ground is real and the inference is sound.** The quoted text exists verbatim at
`internal/cli/todo.go:86-87`: *"because the queue and the integration branch are properties of one
repository, not of whichever worktree the command happens to run in"*, inside `todoLandedRef`'s doc
comment (func at `:88`). Reading the landed ref from the worktree's own config would make the landed
verdict differ between two worktrees of one repository — which is what C-5 preserves. The rejection
follows. *(Citation nit: `spec.md` §A.1 says `todo.go:82-90`, `plan.md` §A.1 says `todo.go:81-90`
for the same ground — the two artifacts disagree by one line. D6.)*

**A3's recorded residual is BOTH mis-cited AND overstated — and one read closes it.** The residual
says level 2 reads a ref that `internal/hook/worktree_base_branch.go:156` *writes*. `:156` is
`worktreeBaseBranchReadConfigReal`, which **reads** `LoadWorktreeBaseBranch`; the write is
`worktreeBaseBranchSetHeadReal` (`git remote set-head`) reached at `:125`. More importantly, the
coupling is **vacuous for this path**:

```go
// internal/hook/worktree_base_branch.go:92-100
if !WorktreeBaseBranchInPrimaryCheckout() { return data }
configured := worktreeBaseBranchReadConfig(projectRoot)
if configured == "" {
    // REQ-WBR-005: the neutral value performs no git-metadata read at all.
    return data
}
```

The writer no-ops on an empty key and gates on the **primary checkout** — the same root
`LandedRefFor` reads. Chain level 2 fires only when level 1 is empty; the writer fires only when
level 1 is non-empty. **They are mutually exclusive by construction**, so no cycle exists. The plan
records an unbounded coupling where a single read bounds it — an unverified premise dressed as an
accepted risk (VCI §1). D2.

### 5. Scope discipline — **REPRODUCED (exclusions justified)**

- **Axis C is genuinely t359's surface.** `todo pr` row 29: `t359 no-link picked`, and the card text
  reads *"[t331 분할 B — 착지 증거 컬럼 스키마] … plan-audit iter-1 이 지목한 재설계 2건"*. Both
  claims — picked, and carrying plan-audit iteration-1 redesign items — are confirmed from the
  queue, not from the lane's word. **No REQ-TLA requirement touches schema or persistence**: 001-006
  are predicate behaviour, 007-009 ref resolution, 010-011 output lines, 012 a read-only property.
  Zero trespass.
- **Axis E is not quietly re-imported.** REQ-TLA-010 modifies an existing verdict line; REQ-TLA-011
  writes to an existing command's stderr. `plan.md` §C states "no new subcommand", §E names it an
  anti-pattern, and no requirement creates a surface. Confirmed clean.
- **Axis D** is excluded as operational; no requirement mentions `spec_id`. Clean.

### 6. Open items the author flagged — **JUDGED**

**(a) Implementation shape (widened `--grep` vs Go-side subject filter) — acceptable to leave open.**
`plan.md` §A.2 names both options, states the trade-off honestly (a widened regex concentrates
correctness in a pattern whose failure is byte-identical to "not landed"), and binds REQ-TLA-005
either way. The silent-failure hazard is covered shape-independently by AC-TLA-001's green
direction, and `plan.md` M3-1 already requires the tripwire be kept or replaced with the equivalent
for whatever shape is chosen. **This does not need deciding before Kickoff.**

**(b) The two-attribution edge case — must be closed in text before Kickoff, though the risk is nil.**
I independently re-measured the lane's new (mid-audit) table over the same 5,837 subjects and
reproduced it **exactly**, with a passing control:

| Form | Count (mine) | Count (lane) |
|---|---|---|
| `^[a-z]+\(tNNN\)!?:` scope id | 290 | 290 |
| `\((card )?tNNN\)$` trailing id | 869 | 869 |
| `^Merge card tNNN` | 5 | 5 |
| both scope AND trailing | 34 | 34 |
| ├ same id | 34 | 34 |
| └ **different id (ambiguous)** | **0** | **0** |

Control: `docs(t1): x (t1)` → (t1, t1); `docs(t1): x (t2)` → (t1, t2). The predicate discriminates,
so the 0 is a real 0. The *risk* is therefore nil. But `acceptance.md` §D states the obligation in
its own words — *"the criterion set must state which wins … It is not left to the implementation to
decide silently"* — and **no criterion states it**. The tiebreak now lives only in the evidence file
(commit `a72fb378c`), which is not the criterion set. D3.

---

## Defects Found

**D1 — `REQ-TLA-002` form 3 is an occurrence test, not a position test, and no criterion excludes it — Severity: BLOCKING — Class: blocking**
`spec.md` §A.4 row 3 defines the third attributing form as *"a merge subject naming the card"*, and
REQ-TLA-002 binds the enumeration to "exactly the three forms enumerated in §A.4". Forms 1 and 2 are
positional; form 3 is not — it is the very *occurrence* test REQ-TLA-001 exists to reject. Measured
on `origin/develop` (5,837 subjects, 414 merge subjects):

- **4 merge subjects name two distinct cards.** e.g. `9a3837b5c Merge branch
  'WT-audit-evidence-store' into WT-audit-advice-integrity (t387 depends on t386 convention doc)` —
  a merge between two card branches that lands **neither** on the integration branch, yet form 3
  attributes **both** t386 and t387.
- **26 merge subjects merge INTO a `WT-` branch while naming a card** — absorb merges, not landings.
  e.g. `7a9ea9bf2 Merge branch 'develop' into WT-llm-yaml-preserve (absorb t239, lane-15 window)`.
  **9 of those 26 are caught by form 2 as well**, e.g. `c4ae1ecbd Merge origin/develop into
  WT-audit-participant-count — absorb upstream before integration (card t284)` — so the hole is not
  confined to form 3.
- **50 merge subjects name a card in a shape the enumeration pins to nothing** (not `Merge card
  tNNN`, not a trailing `(tNNN)`/`(card tNNN)`).

No acceptance criterion excludes any of this: AC-TLA-003 exercises only the `(card t263)`
merge-into-develop shape and a non-merge counter-case, so the mutant `MUT-MERGE-ANY-TOKEN` passes
the entire criterion set while reintroducing the exact defect class the SPEC exists to remove.
*It does not change today's 7/9 split* — the five body-only false positives have no subject
occurrence at all — which is precisely why it would land silently.
**Required fix:** replace §A.4 row 3's shape with a positional definition covering the two observed
merge attributing shapes (`^Merge card tNNN ` and a trailing `(card tNNN)`/`(tNNN)`), state
explicitly that a merge whose target is a `WT-` branch does **not** attribute, and add a criterion
(or a third clause on AC-TLA-003) whose red direction is a merge subject naming a card other than
the merge's own — citing `9a3837b5c` and `c4ae1ecbd` as the fixtures.

**D2 — `plan.md` §A.1 residual risk: wrong citation, and an unverified premise a one-line read disproves — Severity: SHOULD-FIX — Class: blocking**
`plan.md` §A.1 "Residual risk carried by A3" and `spec.md` §A.6 both state that
`internal/hook/worktree_base_branch.go:156` **writes** `refs/remotes/origin/HEAD`. Measured: `:156`
is `worktreeBaseBranchReadConfigReal`, a config *read*; the write is `worktreeBaseBranchSetHeadReal`
via `git remote set-head`, reached at `:125`. And the residual itself is vacuous for this path:
`RunWorktreeBaseAlignment` returns at `:97-100` when the configured key is empty (REQ-WBR-005) and
gates on the primary checkout at `:92` — the same root `LandedRefFor` reads. Level 2 fires only on
an empty key; the writer only on a non-empty one. Mutually exclusive; no cycle.
**Required fix:** correct the citation to `:125` (write) / `:170` (`SetHeadReal`), and replace the
unbounded-coupling wording with the measured disjointness, keeping only the genuine residual (a
project that *configures* the key never reaches level 2, so the two surfaces never actually
interact on this path).

**D3 — `acceptance.md` §D mandates a two-attribution ruling that no criterion supplies — Severity: SHOULD-FIX — Class: blocking**
§D: *"the criterion set must state which wins, or accept both as attributions; either ruling is
acceptable provided it is stated and tested. It is not left to the implementation to decide
silently."* No criterion AC-TLA-001..012 states or tests it. A tiebreak (trailing-paren first,
scope second) was written into `.moai/reports/t472/axis-bf-measurement.md` at commit `a72fb378c`,
which is evidence, not a criterion. I reproduced the supporting population as **0 ambiguous of 34
dual-form subjects** with a passing control, so the risk is nil — but the SPEC currently violates
its own stated obligation.
**Required fix:** either add the one-line ruling to §D as a stated tiebreak with an `N ≥ 1` revisit
trigger, or amend §D's wording to record the ruling as deliberately deferred on a measured
zero-observation basis. Do not leave the mandate unmet.

**D4 — AC-TLA-009 asserts a value inequality where it means provenance — Severity: SHOULD-FIX — Class: blocking**
Second clause: *"Given the same root with `refs/remotes/origin/HEAD` resolvable, Then the resolved
ref is not `DefaultLandedRef`."* `DefaultLandedRef` is the literal `"origin/main"`
(`internal/kanban/prlink_landed.go:41`). In any repository whose `origin/HEAD` names `main` — the
common case — a *correct* level-2 resolution yields `origin/main`, and the criterion false-fails.
The criterion means "the answer came from level 2, not level 3", which is exactly the provenance
REQ-TLA-011 introduces.
**Required fix:** restate the clause against the chain level that answered (the value REQ-TLA-011
already requires the resolver to carry out), not against the ref string.

**D5 — `acceptance.md` §A swaps the two positive controls' attributing forms — Severity: MINOR — Class: blocking**
§A: *"**t401** and **t440** are the two measured true positives … attributed by conventional-commit
scope and trailing parenthetical respectively."* Measured: t401's only subject occurrence is
`d5caf2d8e feat(SPEC-JUDGMENT-FIRST-MODE-001): … (t401)` — a **trailing parenthetical** under a
non-card scope; t440 carries `b80cc9cf1 docs(t440): …` — a **conventional-commit scope**. The forms
are the other way round. §B contradicts §A within the same file: AC-TLA-001 correctly uses `docs(t440)`
for scope and AC-TLA-002 correctly uses `(t401)` for the trailing form.
**Required fix:** swap the two form names in §A so the positive-control declaration matches the
criteria that consume it.

**D6 — Truncated command output presented as verbatim, in three places — Severity: SHOULD-FIX — Class: blocking**
`spec.md` §A.3 prints `git log origin/develop --perl-regexp --grep='\bt216\b' --oneline` as
returning one commit. Re-run in this tree at `origin/develop` = `7835148d3` (unmoved): it returns
**three** — `48c35a4d4`, `673d3d8a0`, `2f170549b`. The `\bt443\b` query is shown returning one
commit; it returns **15+**. `acceptance.md` §A repeats both. No elision marker is present. The extra
matches are body mentions and other cards' subjects, so they *strengthen* the argument — this is a
presentation defect, not spin — but a reader re-running the cited command sees different output and
cannot reproduce, which is what VCI §3 Evidence forbids ("the command run plus its **verbatim**
output — a summary is not evidence").
**Required fix:** show the full output, or mark the elision and state the selection rule
("first/only subject-occurrence line shown; N body-mention lines elided").

**D7 — `plan.md` M3 item 2 carries work with no requirement and no criterion — Severity: MINOR — Class: optional**
M3-2 ("Extend the `todo pr` outcome documentation to state the predicate's limit") maps to no
REQ-TLA and to no AC, and `acceptance.md` §F's Definition of Done does not gate it. M3-1 maps
loosely to REQ-TLA-005 / AC-TLA-006. Unbounded scope in a milestone.
**Required fix:** either fold M3-2's outcome into AC-TLA-006's scope or state it explicitly as
non-gating documentation work in §F.

**D8 — `acceptance.md` preamble over-states its own bidirectional obligation — Severity: MINOR — Class: optional**
The preamble binds *"every axis-F criterion"* to assert a landed green direction AND a not-landed
red direction. AC-TLA-005 and AC-TLA-006 are structural (enumeration locality, argv construction)
and assert neither; they carry a mutation direction instead, which is equivalent protection.
**Required fix:** scope the sentence to predicate-behaviour criteria, so the two structural criteria
are not read as violations of the SPEC's own rule.

**D9 — Two figures do not reproduce under the SPEC's own definitions; the audited tree took foreign commits mid-audit — Severity: MINOR — Class: optional**
(a) `spec.md` §A.2's control — *"ids attributed in a subject = 291"* — does not reproduce under
§A.4's three-form enumeration: I measure **257** distinct subject-attributed ids (message-mentioned
= 379, reproduced exactly; `subj ⊆ msg` = true). The lane's own report says the proxy was
"paren convention + bare-token", a broader rule than the SPEC defines, so the figure was produced by
a definition the SPEC does not carry. The control's *purpose* — both operands non-empty, comparison
non-vacuous — holds either way.
(b) `progress.md` states, as a present-tense residual-bounding argument, that `git status --short`
returns `?? .moai/specs/SPEC-TODO-LANDING-ATTRIBUTION-001/`. Measured at audit time: `git status
--short` returns **empty** — the SPEC directory was committed at `62cbfdf77`. The cited argument is
no longer reproducible.
(c) **Process observation, not a SPEC defect.** HEAD moved from `62cbfdf77` to `a72fb378c` during
this audit (a lane commit appending 50 lines to `axis-bf-measurement.md`), and `progress.md` was
also rewritten mid-audit. Per `agent-common-protocol.md` § Background Agent Execution, an actively
audited worktree has exactly one writer. Reported here rather than corrected; nothing under
`.moai/reports/t472/` was modified by this audit except this file.
**Required fix:** (a) re-measure the control under §A.4's three forms and record 257, or state the
broader proxy the 291 was measured with; (b) re-measure or date-stamp the `git status` argument;
(c) no fix — lead awareness.

---

## Recommendation

**PASS-WITH-DEBT.** The SPEC clears the Tier M threshold (0.8125 ≥ 0.80) and every must-pass
criterion. Its central structural decision — M1 before M2 — is not merely asserted but fully
reproducible: I re-derived both the 2/2 current-ref figure and the 9-card / 7-false counterfactual
independently, and found the 7-false conclusion *more robust* than the lane's own residual-risk note
concedes, because five of the seven have zero subject occurrence in any position. The mutant set's
decisive pair (t216, t443) holds, so the naive subject-only repair is genuinely excluded.
Traceability is perfect. This is a well-grounded SPEC.

It is not a clean PASS because of one hole with measured instances: **REQ-TLA-002's third form is
not a position**, and the criterion set does not close it. An implementation that satisfies every
criterion as written can still attribute a card from a merge into another card's worktree — 26 such
merges exist in the corpus, 9 of them reachable through form 2 as well. That is the same defect
class the SPEC was written to eliminate, re-entering through its own enumeration, and it would land
silently because it does not perturb today's numbers.

Before Implementation Kickoff Approval, close in this order:

1. **D1** — repair form 3 into a positional definition, exclude merges targeting a `WT-` branch, and
   add the missing red-direction criterion with `9a3837b5c` / `c4ae1ecbd` as fixtures.
2. **D3** — state the two-attribution tiebreak in the criterion set, or record the deferral on its
   measured-zero basis. The evidence file is not the criterion set.
3. **D4** — restate AC-TLA-009's second clause against the answering chain level, not the ref string.
4. **D2**, **D5**, **D6** — citation and evidence-integrity repairs; cheap, and D2 in particular
   *removes* a risk the plan currently carries unbounded.

The open implementation-shape decision (widened `--grep` vs Go-side filter) is correctly left to
run-phase and needs no ruling now. D7-D9 are the orchestrator's discretion.

Re-audit on iteration 2 should be scoped to this enumerated defect delta plus a regression check —
not a from-scratch review. Verdict authority remains with this agent.
