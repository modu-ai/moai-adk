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

---
---

# Plan Audit — SPEC-TODO-LANDING-ATTRIBUTION-001 (card t472) — **ITERATION 2**

Iteration: 2/2 (Tier M ceiling — no iteration 3 is available)
**Verdict: PASS-WITH-DEBT**
Aggregate score: **0.875** against the Tier M threshold **0.80** — delta **+0.0625** from iteration 1's 0.8125.
**Ready for Implementation Kickoff Approval: NOT AS WRITTEN.** Ready after two text-only
corrections (D1, D3 below). Neither is a design change. Detail in § Kickoff readiness.

Reasoning context ignored per M1 Context Isolation. `.moai/reports/t472/axis-bf-measurement.md` and
`premise-recheck.md` were read as *records of claims*; every figure below is labelled re-run or accepted.

## Audit provenance (VCI §2.2)

| | Value |
|---|---|
| Tree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t472` |
| Branch | `WT-landed-drift-detect` |
| HEAD at audit start | `165d69d14` |
| HEAD at audit end | `165d69d14` — **unmoved**. Working tree clean at start. The iteration-1 process defect (foreign writes to an actively audited worktree) did **not** recur. |
| Judging binary | `go build -o /tmp/moai-audit2 ./cmd/moai` from this tree, rc=0. The installed build was not used for any figure. |
| Corpus refs | `origin/develop` = `7835148d3`, `origin/main` = `7ad9f8534` (re-read this run) |
| Corpus size | 5,837 subjects, 414 merge subjects (re-counted this run) |
| Lint | `spec lint --strict <SPEC>/spec.md` with the tree-built binary -> `No findings`, rc=0. Scoped to this SPEC only; no whole-corpus scan, no background process. |

---

## MUST-PASS criteria

| | Criterion | State | Evidence (this run) |
|---|---|---|---|
| MP-1 | REQ number consistency | **PASS** | `REQ-TLA-001..012`, one definition each, sequential, uniform three-place padding. `REQ-TLA-001` appears 3x = 1 definition + 2 prose back-references. No gaps, no duplicate definitions. |
| MP-2 | GEARS compliance (**requirement layer** — `REQ-TLA-*` in `spec.md`; `AC-TLA-*` are the verification layer, graded under Group 4, not here) | **PASS** | 001/002/005/007/010 Ubiquitous; 003/004/012 Unwanted (`shall not`); 006/009 Event-driven (`When ...`); 008 `Where` (static-config gate); 011 compound `While ... when ...`. 12/12. REQ-TLA-002 grew at 0.2.0 but stays Ubiquitous (`The set ... shall be exactly ...`). |
| MP-3 | YAML frontmatter validity | **PASS** | 12 canonical fields present with canonical names (`spec.md:2-14`), `version: "0.2.0"` quoted, plus optional `tier: M` / `related_specs:`. Confirmed by the tree-built linter, rc=0. |
| MP-4 | Language neutrality | **N/A (auto-pass)** | Single-language SPEC (`module: "internal/kanban, internal/cli"`). |
| MP-5 | D7 cross-SPEC reconciliation | **PASS** | `SPEC-TODO-LANDING-STATE-001` = `completed`; `SPEC-KANBAN-QUEUE-PR-SYNC-001` = `in-progress`; `SPEC-WORKTREE-BASEREF-001` = `completed`. None retired/superseded/archived. No BLOCKING finding. |
| MP-6 | D8 cross-platform discipline | **PASS** | `grep -c syscall spec.md` -> 0. Auto-PASS per D8-4. |
| MP-7 | Clarification gate | **PASS** | `grep -rn 'NEEDS CLARIFICATION' <SPEC dir>` -> 0 matches. `research.md` absent (correct for Tier M). |

No must-pass failure. The firewall holds.

---

## Category scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.75 | 0.75 | Iteration-1's D1 contradiction is **gone**: forms 1/2/3a are unambiguous positions, so 0.75 is now *earned* rather than generously awarded. Not 1.0: (i) §A.4's preamble — "a card token ... inside a dependency or absorb note — attributes nothing" — contradicts form 3b's literal admission of **any** token inside the trailing group (iter-2 D1); (ii) form 3b's target ("the branch the landed ref resolves to") is unanalysed against the [HARD] M1-before-M2 order, in which the resolved ref is still `origin/main` (iter-2 D3). |
| Completeness | **1.00** | 1.0 | Both iteration-1 completeness gaps closed and verified: §D carries the two-attribution ruling with its population table and `N >= 1` revisit trigger (iter-1 D3), and M3-2 is marked non-gating in both `plan.md:117-120` and `acceptance.md:236-240` (iter-1 D7). Six `### Out of Scope — <topic>` H3s, each with specific `-` bullets. HISTORY / §A / §B / §C / §D / §E present; frontmatter complete. |
| Testability | 0.75 | 0.75 | 11 of 12 criteria are binary with a named, corpus-verified fixture; zero weasel words (`appropriate`/`adequate`/`reasonable`/`proper` -> 0 hits in spec.md and acceptance.md). AC-TLA-009's iter-1 D4 repair verified (asserts *which level answered*, not a ref-string inequality). Not 1.0: **AC-TLA-005 asserts a protection that measurably does not exist for form 3b** — iter-2 D1. |
| Traceability | **1.00** | 1.0 | 12 REQ / 12 AC. Set comparison of defined REQ ids against ids appearing in `(maps REQ-...)` clauses -> **0 uncovered, 0 orphaned**. Re-run this iteration. |

Aggregate = (0.75 + 1.00 + 0.75 + 1.00) / 4 = **0.875** >= 0.80.

---

## Iteration-1 regression check — every defect, with its disposition state

| Iter-1 defect | State | How I established it |
|---|---|---|
| **D1** (BLOCKING) form 3 is an occurrence test | **VERIFIED FIXED** (new residuals raised as iter-2 D1/D3) | §A.4 now enumerates 3a/3b positionally plus the [HARD] non-attribution rule. Re-measured: the occurrence reading admits **146** merge subjects vs **5** form-3a — both reproduce exactly over the 414 merge subjects. Absorb-direction exclusion **21** reproduces with the SPEC's own command verbatim (my first attempt returned 26 through a sed artifact of my own; corrected by running the SPEC's own extraction form). Both new fixtures exist with the claimed subjects verbatim: `9a3837b5c Merge branch 'WT-audit-evidence-store' into WT-audit-advice-integrity (t387 depends on t386 convention doc)` and `c4ae1ecbd Merge origin/develop into WT-audit-participant-count — absorb upstream before integration (card t284)`. `MUT-MERGE-ANY-TOKEN` is genuinely red against AC-TLA-003 clause 3. |
| **D2** citation + overstated residual | **VERIFIED FIXED** | Read the file: `:92` primary-checkout gate, `:97-100` empty-key return, `:125` `WorktreeBaseBranchSetHead(configured)`, `:155` `func worktreeBaseBranchReadConfigReal`, `:170` `func worktreeBaseBranchSetHeadReal` — all exact by line-numbered grep. Disjointness restated correctly and relocated to a `spec.md` §D exclusion. |
| **D3** unmet §D tiebreak mandate | **VERIFIED FIXED** | `acceptance.md:179-212`. I re-derived the population table **independently**: scope 290, trailing 869, `^Merge card` 5, both 34, same-id 34, **different-id 0** — every cell matches. My own control (`docs(t1): x (t2)` -> `t1 vs t2`; `docs(t1): x (t1)` -> `t1 vs t1`) confirms the extractor discriminates, so the 0 is a measured 0. `N >= 1` revisit trigger present. |
| **D4** value-inequality where provenance meant | **VERIFIED FIXED** | `acceptance.md:136-142`: the clause now asserts "the disclosure names **level 2** as the answering level", with the false-fail rationale recorded. |
| **D5** swapped positive-control forms | **VERIFIED FIXED** | Re-measured: t401's only subject occurrence is `d5caf2d8e feat(SPEC-JUDGMENT-FIRST-MODE-001): ... (t401)` = trailing paren; t440 carries `b80cc9cf1 docs(t440): ...` = scope (plus `4c3b1653c Merge card t440 ...` and `63ea8693a ... (t440)`). `acceptance.md:33-40` now matches §B. |
| **D6** truncated output as verbatim | **VERIFIED FIXED** | Re-ran both cited commands: the `t216` query returns **3** lines (spec.md §A.3 shows all 3); the `t443` query returns **14** lines (spec.md states "14 lines; 1 shown" with an elision marker and its selection rule). Counts exact. |
| **D7** M3-2 unmapped work | **VERIFIED FIXED** | `plan.md:117-120` + `acceptance.md:236-240`, explicitly non-gating; M3-1 explicitly gated via AC-TLA-006. |
| **D8** over-stated bidirectional preamble | **VERIFIED FIXED** | `acceptance.md:6-15`: scoped to AC-TLA-001..004 and -007, with a separate paragraph giving -005/-006 a mutation direction instead. |
| **D9a** 291 does not reproduce | **VERIFIED FIXED** | Re-derived the whole control chain: forms 1/2/3a attributed-id set = **257**; adding form 3b = **258**; the set difference between them -> exactly **`t412`**, one line; message-mentioned = **379**; attributed-minus-mentioned -> **0** lines, so the subset holds and both operands are non-empty. Every figure in §A.2 reproduces. |
| **D9b** stale status argument | **VERIFIED FIXED, with a fresh MINOR instance** | `progress.md:43-52` preserves the stale reading, marks it stale, and re-argues from a diff stat. The substance re-measures true. But see iter-2 D5. |
| **D9c** foreign mid-audit commits | **NOT REPEATED** | HEAD `165d69d14` at start and at end; clean tree at start. |

**No iteration-1 defect is unresolved.** No stagnation.

---

## The three questions the dispatch put

### (a) Has the mutant merely moved? — **Partly yes, and by a measurably smaller amount.**

Two mutants survive the repaired §A.4:

- **MUT-TRAILING-GROUP-ANY-TOKEN.** Form 3b attributes when "the card token lies **inside** the subject's trailing parenthetical group". A group carrying more than one token — `... into develop (card t500 — absorb t280, includes t239)` — attributes all three under the literal rule. That is the occurrence reading relocated inside the parenthesis, and it is exactly what §A.4's own preamble forbids. **Measured blast radius: 0.** Of the 96 `into develop` merges with a trailing group, 76 carry a card token and **none names two distinct card tokens**; none carries absorb/dependency wording. I derived both figures myself (trailing-group extraction over the filtered set, then a per-line distinct-token count).
- **MUT-NO-FORM-3B / MUT-HARDCODED-DEVELOP.** See D1 and D3 — no criterion exercises form 3b at all.

So the mutant has not moved *back*: the 146/50/21 silent-false-positive class is genuinely excluded. But the new shape carries no falsifier of its own.

### (b) Is the `WT-...` predicate sound? — **Load-bearing, but under-general, and keyed on a convention the SPEC does not own.**

It *is* load-bearing: `c4ae1ecbd` ends `... (card t284)`, an exact trailing group, so **form 2 matches it** and only the non-attribution rule removes it. Not redundant with form 3b.

Under-general on two measured grounds:

1. This repository's own history contains card/agent worktree branches that are **not** `WT-`-prefixed: the merge-target census over the 414 merge subjects returns `worktree-t176`, `worktree-t166`, and four `worktree-agent-*` targets. A `WT-`-keyed rule does not fire on them. (No false attribution today — those subjects carry no trailing group — so this is soundness, not a live defect.)
2. The prefix-independent formulation is strictly stronger and free: *a merge whose named target is not the branch the landed ref resolves to attributes no card.* That is form 3b's own target test, contraposed. It covers `WT-`, `worktree-`, and any downstream user's prefix, and needs no convention the SPEC does not own. The SPEC cites no doctrine for `WT-` (the owning [HARD] rule is `kanban-dispatch.md` § Isolation is entered, never provisioned).

Graded MINOR (D4): the shipped surface is `moai todo`, which reaches repositories whose branch names MoAI does not prescribe.

### (c) Is leaving the target-derivation to run-phase acceptable? — **No. It is a second occurrence of the D1 shape one layer down.**

The *rule* is specified (§A.4 form 3b names "the branch the landed ref resolves to", and REQ-TLA-002 binds the enumeration to §A.4). What is missing is a falsifier: **AC-TLA-003 clause 1's fixture is `48c35a4d4 Merge branch 'WT-incremental-rebuild' into develop (card t263)`, whose trailing group is exact, so form 2 alone satisfies it.** An implementation that hardcodes `develop`, and an implementation that omits form 3b entirely, both pass all twelve criteria. See D1 and D3.

---

## Defects Found

**D1 — form 3b carries no falsifying criterion, and AC-TLA-005 asserts a protection that measurably does not exist for it — Severity: major (SHOULD-FIX) — Class: blocking**
`acceptance.md:95-100` (AC-TLA-005): *"When one shape — or the non-attribution rule — is removed from that declaration, Then the criterion for it (AC-TLA-001, -002, or -003's matching clause) fails."* For form 3b this is **false**. Measured this run:
- AC-TLA-003 clause 1's fixture `48c35a4d4 ... into develop (card t263)` has an **exact** trailing group and is therefore matched by form 2 alone. Removing form 3b leaves the clause green.
- The only id form 3b adds over forms 1/2/3a is **`t412`** (set difference -> exactly one line), attributed solely by `b6231290d Merge branch 'WT-mx-tag-edges' into develop (card t412 — SPEC-MX-TAG-EDGES-001)`. A grep for `t412` in `acceptance.md` -> **0 hits**. No criterion cites it.
- Consequence: an implementation omitting form 3b passes every criterion (cost: t412 reads `not-landed` — loud), **and** an implementation reading the whole trailing group as an occurrence set passes every criterion (cost: 0 measured instances today — silent).

**Required fix (text only):** add a clause to AC-TLA-003 using `t412` in both directions — green on `b6231290d` (non-exact trailing group, develop-targeted), red on `d8c91d907 Merge branch 'origin/develop' into WT-mx-tag-edges (window absorption, card t412)` and `57d2f3ae3 Merge branch 'WT-edge-confidence' into WT-mx-tag-edges (card t412 dependency absorption)`. All three fixtures exist verbatim in the corpus and exercise form 3b and the non-attribution rule together. Add one sentence to §A.4 form 3b restricting attribution to a **single** card token in the group, so the row stops contradicting its own preamble.

**D2 — the recorded cost of the narrowing is wrong by a factor of seven; "exactly one under-count" does not reproduce — Severity: major (SHOULD-FIX) — Class: blocking**
`spec.md:163-168`: *"One under-count survives ... t250."* `plan.md:142`: *"Measured cost ... exactly **one** card (`t250`)."* Measured this run over `origin/develop` `7835148d3`: ids appearing anywhere in a subject = **347**; attributed under forms 1/2/3a/3b = **258**; difference = **89**. Filtering that difference to ids whose subject occurrence sits in an *attributing-shaped* position the four forms miss yields, verbatim from the corpus:

| Card | Its only subject-position evidence |
|---|---|
| t250 | `t250: graph freshness ... (#1648)` — bare prefix (the one the SPEC records) |
| t40 | `Merge branch 'worktree-t40': t40 — moai update observability (3 quiet failures)` |
| t36 | `test(timing): add in-run calibrated latency bounds; ... (card t36, absorbs t2)` |
| t68 | `merge: Factory Mode -f N worker fan-out (t68, SPEC-FACTORY-WORKER-FANOUT-001)` |
| t46, t73, t74 | `merge: anchor-session guards for worktree disposal + registry CWD relocation (t46/t73/t74)` |

Each was confirmed to have **no** attributed occurrence anywhere in the corpus (all seven sit in the subject-ids-minus-attributed difference). The bare-prefix family alone has three members (`t279`, `t250`, `t225`), not one. The true under-count is **>= 7 cards**, and the missed shape is not only "a bare prefix" but "a trailing group carrying the card plus other text" — the same non-exactness form 3b was added to handle for merges, left unhandled for non-merges. This is an unattributed measured claim under VCI §2, and it is load-bearing: `plan.md` §D uses it as the mitigation ground for accepting the loud-failure direction. Operational impact today is nil (all seven are archived-era ids), which is why the severity is major rather than blocking-equivalent — but the figure a reviewer weighs the decision against is wrong by 7x.

**Required fix (text only):** restate as a measured lower bound naming the seven ids and both missed shapes, or widen form 2 to accept a card token as the **first** token of a trailing group. Do not leave "exactly one" standing.

**D3 — form 3b's target derivation is unfalsified, and the [HARD] M1-before-M2 order makes it resolve to `main` in the intervening window — Severity: major (SHOULD-FIX) — Class: blocking**
§A.4 form 3b keys on "the branch the landed ref resolves to". After M1 lands and before M2 lands — the window the [HARD] ordering decision (`spec.md` §A.7, `plan.md` §A.3) deliberately creates — the resolver is still the un-repaired `LandedRefFor`, which returns `DefaultLandedRef = "origin/main"` (`prlink_landed.go:41,74-80`, read this run). A *correct* form-3b implementation therefore attributes **nothing** from the 76 `into develop` merges in that window, while a hardcoded-`develop` implementation attributes all 76. No criterion distinguishes them: AC-TLA-003 clause 1 is satisfied via form 2 (D1), and `acceptance.md` §F's post-M1 check names only t401 and t440, which are form 1 / form 2. Downstream the divergence is permanent rather than windowed: a user repository whose integration branch is `main` gets false positives from every `into develop` merge and misses every `into main` merge.

**Required fix (text only):** state in §A.4 (or as a clause on AC-TLA-003) that form 3b's target is **derived from the resolved landed ref**, with a criterion that resolves the ref to something other than `develop` and asserts the target moves with it; and record in §A.7 what form 3b yields during the M1-only window.

**D4 — the non-attribution rule keys on a branch-prefix convention the SPEC does not own, and which this corpus violates — Severity: minor — Class: optional**
Grounding measurement in §(b) above: six merge targets in this repository's own history are card/agent worktree branches named `worktree-t176`, `worktree-t166`, `worktree-agent-*` — not `WT-`. No false attribution results today, so this is soundness rather than a live defect. The contraposed form-3b target test is strictly stronger, prefix-independent, and costs nothing.

**Required fix:** restate the rule in target-mismatch terms, keeping `WT-...` as an illustrative example, and cite `kanban-dispatch.md` if the prefix form is retained.

**D5 — the iter-1 D9b replacement figure is itself already stale and is not in R4 form — Severity: minor — Class: optional**
`progress.md:46-48` reads *"... reports 6 files, 1101 insertions, 0 deletions"*. Re-run at HEAD `165d69d14`: **6 files, 1301 insertions, 0 deletions**. `HEAD` is a moving ref; the command is correctly stated first (R4 ordering), but the value is neither dated nor labelled a reference, so it drifted 200 lines within two commits. The argument's substance re-measures true (the name-only form returns only the four SPEC artifacts and the two files under `.moai/reports/t472/`), so nothing load-bearing fails — but the fix for a stale-figure defect reproduced that defect in miniature.

**Required fix:** parenthesize and date the value, per VCI §2.1 remedy R4.

---

## The lint exemption decision — predicate applied independently

The `<!-- moving-ref-ok: ... -->` marker on `acceptance.md:131` (the AC-TLA-009 fixture line) is
**correct, and R3 is the right remedy.** I applied VCI §2.1 myself rather than reading the author's
reasoning:

- **Test 1 (substitution).** Substituting the SHA `refs/remotes/origin/HEAD` currently resolves to turns "constructed so that the symref lookup exits non-zero" into a different and weaker sentence. The claim is *about* the ref's absence. -> **SUBJECT**.
- **Test 3.** Re-running next week gives the same answer — the fixture is constructed, not observed. -> reads ANCHOR. Tests 1 and 3 therefore **disagree**, which per the tie-break is evidence *against* ANCHOR and mandates Test 4.
- **Test 4 (read-time action).** No. The reader builds a fixture by omitting the symref; nothing is measured against the live repository. -> **S1**.
- **S1 -> R3.** Keep the ref, declare the exemption with a stated reason. The reason present is non-empty, names the instance shape, and states which tests it applied — it is not the cheapest available silencer. **R4 would be wrong here** (there is no read-time measuring command), and **R1/R2 would be actively harmful** (a pin names a value the criterion requires not to exist).

Independently: the scoped `spec lint --strict` on `spec.md` returns `No findings`, rc=0. `acceptance.md` and `plan.md` are not SPEC-parseable and the linter rejects them at the frontmatter stage, so no lint verdict on the marker's own file was obtainable from this tool — recorded as a Gap, not as a pass.

---

## Kickoff readiness

**Not ready as written.** The blocker is not the design — the design is sound and, wherever I could
re-derive it, exactly reproducible. The blocker is that **AC-TLA-005 states a protection that does
not exist** (D1): a run-phase that omits form 3b, or that hardcodes `develop` as its target (D3),
passes all twelve criteria and satisfies the Definition of Done. That is the same failure shape as
iteration-1's D1 — an enumeration entry with no falsifier — at roughly one-tenth the measured blast
radius.

**Ready after D1 and D3**, both of which are additions to `acceptance.md` using fixtures that already
exist verbatim in this corpus (`b6231290d`, `d8c91d907`, `57d2f3ae3`). Neither changes a requirement,
a milestone, or the ref-chain design. **D2** is a one-paragraph correction of a figure a reviewer
weighs a decision against and should travel with them. D4 and D5 are the orchestrator's discretion
(M6: an optional finding does not by itself justify a FAIL, and none of them is being used to
manufacture one).

**Ceiling note.** This is iteration 2 of the Tier M ceiling 2. There is no iteration 3, so the
corrections above cannot be confirmed by a further audit round under the standing contract. The
orchestrator either routes them as pre-Kickoff text edits and accepts them unaudited, or accepts the
enumerated debt explicitly. A high aggregate score does not discharge D1 — the criterion set's own
protection claim is measurably untrue, and score is not the instrument that repairs that.

## Evidence discipline

**Re-ran independently this iteration:** corpus size (5,837 / 414); form-3b yield 76 and its 0
multi-card groups; occurrence-reading 146; form-3a 5; absorb-direction 21; attributed 257 / 258 and
the one-line difference `t412`; mentioned 379; the subset check -> 0; the §D population table
(290 / 869 / 5 / 34 / 34 / 0) with my own control; the iter-1 D6 line counts (3 and 14); the iter-1
D5 positive-control forms; all five `worktree_base_branch.go` line citations;
`prlink_landed.go:41,74-80,96-109`; `todo.go:81-90`; the two new AC-TLA-003 fixtures; the
merge-target census; the diff stat and name list; the scoped lint; and the whole must-pass set.

**Accepted from the record without re-derivation:** the 2/9 and 9/31 splits (re-derived in my own
iteration-1 report against the same corpus ref `7835148d3`, so re-running would measure the same
tree twice) and the lane's independent D1 reproduction in `axis-bf-measurement.md` (superseded for my
purposes by my own re-derivation above).

**Gaps — what I did NOT observe.** No whole-corpus lint (dispatch-prohibited; shared machine). No
execution of the repaired predicate — none exists; this is a plan-phase audit of text. No lint
verdict on `acceptance.md` (not SPEC-parseable by this tool). No measurement against `origin/main`
beyond the ref read, so D3's downstream claim about `into main` repositories is reasoned from the
rule, not observed in a corpus that has one. No queue query confirming the seven D2 under-count cards
are archived — that is inspection of their id range, not a measurement.

**Residual risk.** If the D2 under-count shape ("card token first in a non-exact trailing group") is
more common on `origin/main` than on `origin/develop`, the loud-failure population after M1 lands is
larger than either the SPEC or this audit measured — the corpus I judged is `develop`, and M1 ships
against `main`.

---

# Plan Audit — Iteration 3 (FINAL) — SPEC-TODO-LANDING-ATTRIBUTION-001

Auditor: plan-auditor. Card t472. Tree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t472`,
branch `WT-landed-drift-detect`, HEAD `11b0216ae`, working tree clean at audit start.
SPEC version `0.3.0`, `status: draft`, `tier: M`.

**Reasoning context ignored per M1 Context Isolation.** The dispatch's factual assertions were
treated as claims to verify, not as findings to adopt. Every figure below is re-derived in this
tree; the corpus is the pinned commit `7835148d3` (5,837 subjects) and, where stated, `origin/main`
`7ad9f8534`.

**Tool provenance (VCI §2.2).** `moai spec lint` was run from a binary built from THIS tree
(`go build -o /tmp/t472audit/moai ./cmd/moai`, rc=0, at HEAD `11b0216ae`) — not the installed build,
which is ~190 commits behind. Lint scope was the single file
`.moai/specs/SPEC-TODO-LANDING-ATTRIBUTION-001/spec.md`; no whole-corpus scan, no background
process.

**Regex-engine note.** This shell aliases `grep` to `ugrep`, which mis-parses `\(` in ERE and
silently returns zero matches for patterns of the form `\([^()]*t[0-9]+[^()]*\)$`. Every count below
marked "real grep" was re-run through `command grep` (BSD). All form counts agree under both
engines; the divergent pattern is called out where it matters.

## Verdict

    Verdict: PASS-WITH-DEBT — aggregate 0.8375 vs Tier M threshold 0.80 (delta from iter-2: -0.0375)
    Kickoff-ready: NO. One blocking text defect (D3-1) must close first.

The score fell. Under the LEAN retry clause a score regression is a STOP signal; this is also
iteration 3 of a 3-iteration ceiling, so the disposition passes to the operator either way. The
regression is not a re-grading of iteration 2's work — every iteration-2 remediation verified
clean — it is a NEW measurement that falsifies a figure and a tolerance argument the SPEC states.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-TLA-001`..`REQ-TLA-013`, 13 ids, sequential, no
  gaps, no duplicates, uniform 3-digit padding. Measured:
  `command grep -ohE 'REQ-TLA-[0-9]+' spec.md | sort -u` → 13 ids, 001-013 contiguous.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer only**
  (`spec.md` §B `REQ-TLA-*`); the `AC-TLA-*` Given/When/Then entries in `acceptance.md` are the
  verification layer and were graded under Group 4, not here. All 13 match a GEARS pattern:
  Ubiquitous 001/002/005/007/010/013; unwanted (`shall not`) 003/004/012; event-driven 006/009;
  Where-gate 008; compound While+When 011. No entry is a bare occurrence or user-story form.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:2-14`): `id`, `title`, `version: "0.3.0"` (quoted semver), `status: draft`,
  `created: 2026-09-03`, `updated: 2026-09-04`, `author`, `priority: P1`, `phase`, `module`,
  `lifecycle: spec-anchored`, `tags` (comma-separated string). Plus `tier: M`. No rejected
  snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`) present.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go; `internal/kanban`,
  `internal/cli`). No multi-language tooling surface. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — referenced SPEC ids resolved and status-read:
  `SPEC-TODO-LANDING-STATE-001` = completed, `SPEC-KANBAN-QUEUE-PR-SYNC-001` = in-progress,
  `SPEC-WORKTREE-BASEREF-001` = completed, `SPEC-MX-TAG-EDGES-001` = completed. None is
  retired / superseded / archived. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `command grep -c 'syscall' spec.md` → 0. Auto-PASS
  per D8-4.
- **[PASS] MP-7 clarification gate** —
  `command grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-TODO-LANDING-ATTRIBUTION-001/` → no match
  (rc=1). `research.md` does not exist (Tier M does not require it); `plan.md` is clean.

Mechanical confirmation: `moai spec lint <spec.md>` → `No findings — all SPEC documents are valid`.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 0.75-1.0 | Every requirement single-interpretation; fixtures quoted verbatim and all resolve in the corpus. Deduction: `REQ-TLA-002` and `REQ-TLA-013` carry multi-sentence rationale inside the normative statement, and `REQ-TLA-002` folds an implementation-structure clause ("shall live in one named place in the implementation") into a behavioural requirement. |
| Completeness | 0.65 | 0.50-0.75 | All sections present (HISTORY / §A / §B / §C / §D with six `### Out of Scope — …` H3s carrying bullets / §E) and frontmatter complete — but the SPEC's central deliverable, the §A.4 enumeration, is measurably incomplete against its own doctrine, and the stated residual (7) is wrong by a factor of ~6.6. See D3-1. |
| Testability | 0.80 | 0.75-1.0 | Every AC is binary, bidirectional, and named against a mutant; no weasel words; the AC-TLA-005 falsifier map reproduces exactly. Deduction: the criterion set and the §F Definition-of-Done positive controls (`t401`, `t440`) are both drawn from shapes the enumeration already handles, so **no criterion in the file can detect D3-1** — the same vacuity class iter-2 D1 closed for form 3b, recurring at the suite level. |
| Traceability | 1.00 | 1.0 | 13 REQ, 13 AC headings (`AC-TLA-001`..`012` plus `AC-TLA-003b`). Every REQ has at least one AC (`REQ-TLA-013` ← AC-TLA-003 clause 6); every AC names existing REQs; zero orphans. Verified by set comparison of `REQ-TLA-*` across both files. |

Aggregate (unweighted mean): **0.8375**.

## What I verified, could not verify, or accepted from the record

**VERIFIED — independently reproduced in this tree, exact match:**

- Corpus size 5,837 subjects at `7835148d3`.
- Form counts: form 1 = 290, form 2 = 869, form 3a = 5, form 3b = 76, form 3c = 31.
- `347` ids in a subject / `258` attributed by forms 1-3b / `89` unattributed / `270` with form 3c /
  the exact 12 ids form 3c adds (`t106 t110 t113 t114 t32 t36 t56 t59 t69 t79 t98 t99`).
- **The AC-TLA-005 falsifier map, all five rows**: form 1 alone → 42, form 2 alone → 77, form 3a
  alone → 1 (`t244`), form 3b alone → 1 (`t412`), form 3c alone → 12. Computed as
  `comm -23 <form-set> <union of the other four>`. Every row reproduces.
- The residual-7 ids (`t225 t250 t40 t46 t68 t73 t74`) are each genuinely unattributed by all five
  forms.
- **D1 closed**: `t412` appears 11× in `acceptance.md`; all four fixture subjects (`b6231290d` and
  the three `into WT-mx-tag-edges` siblings) exist verbatim in the corpus.
- **D3's refutation**: `origin/main` `7ad9f8534` → 4,457 subjects, 101 `^Merge` subjects, **0**
  `into develop` merges with a card-bearing trailing group, **0** for `into main`. The window claim
  from iteration 2 is refuted; my own iteration-2 Gap ("had not measured the `origin/main` corpus")
  is now closed against me.
- **D4**: 9 non-`WT-` merge targets, three carrying no prefix at all (`t403`, `t78`, `t86`).
- 8 trailing groups corpus-wide naming two or more distinct cards; 0 on form 3b's own subject set.
- 3 bare-`tNNN:`-prefix subjects, of which `t250` / `t225` are residual — matching the SPEC's claim
  of 2 ids in that shape.

**ACCEPTED FROM THE RECORD — not independently re-derived:**

- **D2's refusal of the first-token widening.** The fixture is real (`merge: t79 — glm_task
  delegation family (branch WT-t80)` exists verbatim; `t80`'s only other subject is
  `Merge branch 'release/v3.1.1' into WT-t80`). The reasoning — trading a loud under-count for a
  silent false positive is the wrong direction for a SPEC whose purpose is removing silent false
  positives — is **sound, and I endorse it**. What I did not do is re-derive the widening's full
  blast radius across the corpus.
- **Form 3c's "31 of 31 the first token is the card the merge delivers".** I verified the *shape*
  (31 subjects begin `merge: tNNN`; 110 `merge: ` subjects total, 79 not card-led). The *delivery
  semantics* of each of the 31 requires per-commit content inspection I did not perform.
- **D5** (`1101 → 1301 → 1569`, R4 form) — verified present in `progress.md:38-39,81-84`; the 1569
  measurement itself accepted from the record.

**COULD NOT VERIFY:**

- Whether the 39 ids in D3-1 sit in the operative live queue. The store at
  `.moai/state/todo/backlog.db` (45 `items`, 142 `archived_items`) contains 38 of the 39 in neither
  table, which is consistent with the known `moai todo` / disk-store split. Their live-queue status
  is therefore a **Gap**, and D3-1 below does not rest on it.
- That the residual is exactly 46 or 49 rather than larger. I named one further shape exhaustively
  and two partially; I did not classify the whole residual exhaustively either. Like the author's
  19, my number is a **floor**.

## Defects Found

**D3-1 — `spec.md` §A.4 form 2 + §A.4.2, `plan.md` §D, `acceptance.md` §F — Severity: major — Class: blocking**

Form 2 is anchored `\((card )?t[0-9]+\)$`. That `)$` excludes **43 subjects carrying 40 distinct
card ids, 39 of them attributed by no form at all**, whose shape is form 2's own shape with a
pull-request reference appended:

    fix(hook): stop the registry walk at the home directory (t168) (#1609)
    feat(kanban): moai todo pr — read-only card-to-PR and landed link (t210) (#1628)
    feat: lead-session deputy — background-parallel dispatch/CI-watch via manager-lead (SPEC-LEAD-DEBOTTLENECK-001, t283) (#1664)

Grounding measurement (real BSD grep; the shell's ugrep alias returns 0 for this pattern):

    command grep -E '\([^()]*t[0-9]+[^()]*\) \(#[0-9]+\)$' subj.txt | wc -l        -> 43
    ... same, ids extracted, sort -u                                               -> 40
    comm -12 <those ids> <ids unattributed by all five forms>                      -> 39

Three consequences, each independently blocking:

1. **The stated residual is wrong for the third consecutive version.** `1` (v0.2.0) → `19`, reduced
   to `7` (v0.3.0) → measured **at least 46** (the 7 union the 39), and at least 49 once two smaller
   shapes are added (`t311` via `(closes t311)`; `t121` / `t128` via `merge(lane-1): t121 …`). This
   is the same defect class the SPEC has now restated twice, because the classification method —
   inspect part of the residual and name what is noticed — has never been exhaustive on any
   iteration, mine included.
2. **`plan.md` §D's tolerance argument does not hold.** Its ground is "all seven are archived-era
   ids so today's operational impact is nil". The 39 are the repository's **most recent** merges, PR
   `#1600`-`#1670`. The archived-era premise does not extend to them; whether they are live is a Gap
   (above), but the dating claim is falsified.
3. **This is not a new shape — it is a boundary defect in the SPEC's largest form.** These subjects
   are already form 2 positionally: the card sits in a trailing parenthetical group carrying nothing
   else, followed by a PR reference. And the shape occurs **43 times on `origin/main` `7ad9f8534`** —
   the exact ref M1's window walks (`spec.md` §A.7). An implementation built to the enumeration as
   written would answer `not-landed` for this repository's dominant PR-merge landing path, on the ref
   the repaired predicate first ships against.

No criterion in `acceptance.md` can detect this: the §F Definition-of-Done positive controls are
`t401` (form 2, `)$`-exact) and `t440` (form 1) — both shapes the enumeration handles — so the DoD is
satisfiable while 39 recent landings read `not-landed`.

**Required fix (bounded, text-only):** restate form 2's anchor so the card-bearing group need only be
the **last card-bearing** trailing group rather than the last group — for example by permitting a
trailing non-card reference group (`(#NNNN)`) after it — or add it as a sixth enumerated form;
re-state the residual figure in `spec.md` §A.4.2 and re-argue `plan.md` §D's tolerance at the
corrected number; and add one AC clause with a `(tNNN) (#PR)` fixture so the shape carries a
falsifier. Verifiable by the two commands quoted above.

**D3-2 — `acceptance.md` §D (two-cards-in-one-group bullet) — Severity: minor — Class: blocking**

§D claims the eight multi-card trailing groups were checked "id by id" and enumerates the ids whose
rejection costs nothing. The measured population of residual ids sitting in a non-exact trailing
group is `{t2, t311, t46, t68, t73, t74, t80}`. `t2` and `t80` are ruled on explicitly;
`t46 / t68 / t73 / t74` are in the residual-7; **`t311` is ruled on nowhere.** Its sole subject
occurrence is `merge(WT-codex-init): integrate card t340 … (closes t311)`. Required fix: classify
`t311` — either as a further residual under-count or, like `t2`, as a correctly-unattributed note —
and say which.

**D3-3 — `spec.md` §A.4 preamble — Severity: minor — Class: optional**

"Counts measured over the pinned commit `7835148d3`, 5,837 subjects, **414 of them merges**." 414 is
the count of subjects **beginning** `Merge`. The actual merge-commit count at that corpus is **661**
(`--merges`, one line per commit). 247 merge commits — including all 31 form-3c subjects and 137
others — sit outside the denominator the merge-shape survey is described against. Required fix: say
"414 subjects begin `Merge`", or state 661.

**D3-4 — `spec.md` §A.4 (the nine-target list) — Severity: minor — Class: optional**

The prose says "Nine targets"; the list printed above it has **eleven** entries. Two
(`worktree-agent-a205e7a01ec2e0f27`, `worktree-agent-a350b7a40faaf39c6`) are **not** merge targets
under the command cited beside the list — each appears once as a merge *source*. Re-running the
SPEC's own extraction command yields exactly the nine the prose claims. Required fix: delete the two
non-reproducing entries so the list matches the command quoted with it.

**D3-5 — `acceptance.md` AC-TLA-005 entry test vs `plan.md` §D — Severity: minor — Class: optional**

Asked directly whether the entry test is a real gate or a restatement: **it is a real gate.** It is
mechanically computable, I reproduced all five columns exactly, and it has already done work — form
3b's column of `1` is what exposed the iteration-2 vacuity. But `plan.md` §D restates it as a
rejection rule ("reject it if that count is 0"), which is **stronger than the SPEC's own practice**:
REQ-TLA-013's row reads `n/a` and is falsified by a constructed `release/v9` fixture, not a corpus
id. A zero column means "no corpus falsifier exists", not "no falsifier exists". Required fix: narrow
`plan.md` §D's wording to "admit only with a named falsifier; a zero column means the falsifier must
be a constructed fixture, stated as such".

## Regression Check (iterations 1-2)

| Prior defect | State | Evidence |
|---|---|---|
| iter-2 D1 (form 3b unfalsified; `t412` absent) | **RESOLVED** | `t412` appears 11× in `acceptance.md`; AC-TLA-003 clauses 4-5 added; all four fixtures exist verbatim; form 3b alone → exactly `t412`. The author additionally found and closed the parallel form-3a vacuity (`AC-TLA-003b`, `t244`) — a defect iteration 2 did not find. |
| iter-2 D2 (under-count figure) | **PARTIALLY RESOLVED; the class recurs** | The retraction 1 → 19 → 7 is correct as far as it goes and reproduces exactly. Form 3c is a sound, measured adoption. But the figure is still wrong — see D3-1. |
| iter-2 D3 (M1-window divergence) | **RESOLVED — and my claim was refuted** | The `origin/main` measurement reproduces exactly (4,457 / 101 / 0 / 0). My iteration-2 window claim was an inference recorded as a Gap; it is false on this repository. The substitution (permanent downstream defect, `REQ-TLA-013` + clause 6) is **sound**, and the constructed `release/v9` fixture is **acceptable and in fact necessary** — by the author's own measurement no real subject can distinguish the two implementations, so a corpus fixture is impossible in principle. Leaving the M1→M2 ordering untouched is correct: the ordering's ground (2→9 false-positive growth) is independent, and nothing measured here disturbs it. |
| iter-2 D4 (`WT-` prefix predicate) | **RESOLVED** | Contraposed target test; 9 targets verified, 3 prefix-free. The list defect D3-4 is cosmetic and does not affect the rule. |
| iter-2 D5 (diff-stat figure) | **RESOLVED** | `progress.md` carries 1569 in R4 form with the file-name list as the load-bearing argument. |
| Two-cards-in-one-group axis kept separate from the §D tiebreak | **VERIFIED SEPARATE** | The tiebreak's axis (scope-id ∧ trailing-id, 0 instances) and the one-group axis (8 instances) are counted, captioned, and ruled on separately; the enumeration rejects the second rather than tiebreaking it. Not conflated. Gap: `t311` (D3-2). |

**Stagnation signal.** No individual defect persisted unchanged across all three iterations. But the
*under-count figure* has been materially wrong in all three versions (1 → 19/7 → ≥46). The cause is
methodological rather than clerical: the residual has never been classified exhaustively. That is the
one item a fourth pass would have to change in kind rather than in degree.

## The `AC-TLA-003b` sub-ID convention

Asked directly: **not a problem.** The id is unique, greppable, traced to its REQs, and sits under
the Tier M ceiling of 16 on either count (13 REQ / 12-or-13 AC). Pairing form 3a and form 3c under
one heading is defensible — they are one shape in two spellings, as §A.4 states. The only cost is
that the AC count is ambiguous in prose, and the SPEC already discloses both readings. No change
required.

## Enumerated debt a PASS-WITH-DEBT would carry

1. **[blocking; must close before run-phase]** The §A.4 enumeration misses 39 unattributed card ids
   in form 2's own positional shape with a trailing `(#PR)` group — 43 subjects on the pinned corpus
   and 43 on `origin/main`, the ref M1 first ships against. Accepting this as debt means accepting
   that the shipped predicate answers `not-landed` for the repository's dominant PR-merge landing
   path.
2. `t311` unclassified in a population §D claims to have checked id-by-id.
3. "414 of them merges" understates the merge-commit denominator by 247.
4. The nine-target list prints eleven entries, two unreproducible under its own command.
5. `plan.md` §D's sixth-form rejection rule is stated more strongly than the SPEC's own practice.
6. **Carried from iteration 2 and still open:** the residual's operational impact after M1 lands is
   measured on `develop` while M1 sails against `main`. `t225` / `t250` (bare prefix) and `t40`
   (legacy `worktree-`) may be commoner on `main`. Unmeasured, by author and auditor alike.
7. **New Gap:** neither the author's 19 nor my 46 is a total. Neither party has classified the
   residual exhaustively.

## Kickoff readiness

**Not ready for Implementation Kickoff Approval as written.** Debt items 2-7 are acceptable to carry
into run-phase. Debt item 1 is not: it is a defect in the artifact run-phase would implement from,
its failure direction is exactly the silent-miss-on-the-live-landing-path the SPEC exists to prevent,
and no acceptance criterion in the file can catch it.

**Recommended disposition.** The fix is bounded and mechanical — one enumeration row, two figures,
one AC clause — and its correctness is checkable with the two commands quoted under D3-1. I recommend
the lane apply that amendment and the lead verify it against those commands, rather than opening a
fourth full audit iteration (which the ceiling does not permit) or reducing scope (the scope is not
the problem; the enumeration's completeness is). Once D3-1 is closed and verified, this SPEC is
Kickoff-ready with debt items 2-7 recorded.

## Residual risk

- My 46 is a floor. I named one shape exhaustively and two partially; the residual has still not
  been classified end to end by anyone. A further shape may exist.
- The live-queue impact of the 39 ids is unmeasured (the store split above). The finding does not
  depend on it, but its severity would rise if any of them are live cards.
- All figures here are measured against `7835148d3` and `7ad9f8534`. `origin/develop` has moved to
  `25a3212a9`; a reader re-running against the branch name will measure a different corpus, as the
  SPEC's own pin discipline (VCI §2.1 remedy R1) requires.
