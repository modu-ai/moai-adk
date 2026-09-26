# SPEC-AUTONOMY-PRECONDITION-001 — Implementation plan

Milestones are ordered by decision-reversibility: the two design decisions (how the deny recognizes
bypass shapes; how the contract projects onto the mission validator) come first, and mechanical
template and documentation work last.

## §A — Context

Card **t1245**, track **A2b**. Base tree `develop` at `553e224f3`, worktree
`.claude/worktrees/t1245`, branch `WT-push-serialize-sign`. Created by lead ruling 09-26 (2) #3,
which split three items out of SPEC-AUTONOMY-ESCALATION-001. Two independent deny components plus
one reuse projection; they share the contract resolver and nothing else.

## §B — Known issues carried in

- Finding **N4** — the sign guard's wrapper coverage was a hole with no requirement. Now
  REQ-AP-005 with a closed, option-aware wrapper list (design.md §C.3) and a mutant-checked
  unknown-wrapper criterion (AC-AP-010).
- Finding **N8** — the sign deny contradicted the "nothing changes under `guided`" promise. Resolved
  in wording, not by weakening the guard (REQ-AP-010, design.md §C.7).
- Finding **N5** — the receipt criterion could only turn green after work this SPEC does not own.
  **Closed**: A1's R5 and R6 are both answered (spec.md §C.4, §C.5), so no criterion here is
  conditional.
- Criterion id re-use on the t1235 branch — this SPEC uses its own namespace; the provenance table
  is spec.md §H.

## §C — Pre-flight (before M1)

1. Re-read branch and HEAD (`git -C <tree> rev-parse --short HEAD`, `git -C <tree> branch --show-current`)
   and confirm `WT-push-serialize-sign` on a tree descended from `553e224f3`.
2. Re-measure every spec.md §C row against the tree at run-phase start. Two rows are expected to
   have **changed** by then: `moai contract sign` exists once A1 lands (§C.2), and the template
   autonomy block exists once A1 lands (AC-AP-013). A changed row is a premise update recorded in
   progress.md, not a silent adjustment.
3. Confirm A1's `push_requires_lease` field is present in its `show --json` projection; where A1
   renamed or dropped it, stop and report rather than inventing a substitute.
4. Capture the `golangci-lint run` baseline on this tree, so "no new issue" in acceptance.md §H is
   attributable.

## §D — Constraints

Carried from spec.md §E: no new lease mechanism (C1); opposite fail directions are deliberate (C2);
neither component writes an escalation record (C3); activation ordering is A1's (C4); template
neutrality (C5).

## §E — Self-verification

Each milestone closes only with the named tests present as `--- PASS:` lines and a non-empty swept
count, plus the cross-platform build. The full gate list is acceptance.md §H.

## §F — Milestones

### M1 — Push serializer (highest reversibility cost: it writes a shared lease record)

- REQ-AP-001, REQ-AP-002, REQ-AP-007; design.md §B.
- Activation triple (mode + action + `push_requires_lease`), `git push develop` matcher, admit /
  deny / release / reclaim, fail-open.
- ACs: AC-AP-001, AC-AP-002, AC-AP-003, AC-AP-004.
- First because it touches a record other sessions read, so a wrong shape here is the most
  expensive to unwind.

### M2 — Contract-sign guard (the parsing decision)

- REQ-AP-003, REQ-AP-004, REQ-AP-005, REQ-AP-006, REQ-AP-009; design.md §C.
- Quote-removing word split, assignment and wrapper stripping, basename match, one-level `-c`,
  fail-closed unclassified branch, receipt-path allowance.
- ACs: AC-AP-005 .. AC-AP-012.
- Second because the matcher's shape is the card's other genuine design decision, and every later
  step assumes it.

### M3 — Documentation of the mode-independent deny (finding N8)

- REQ-AP-010; design.md §C.7.
- Template `workflow.yaml` autonomy comment and the contract-mode rule text; `make build`.
- AC: AC-AP-013.
- Depends on A1's autonomy block existing. Where it has not landed, M3 waits and the wait is
  reported — it is not worked around by creating the block here.

### M4 — Mission-validator projection (mechanical once the direction is fixed)

- REQ-AP-008; design.md §D.
- One-way projection, fail closed on an unmapped field, mission's exported surface untouched.
- AC: AC-AP-014.
- Last because the direction decision is already recorded in design.md §D and the remaining work is
  mapping.

## §G — Anti-patterns

- Scrubbing quoted spans instead of removing quotes — it erases `'moai'` and defeats the matcher
  (design.md §A).
- Weakening the sign deny to preserve the `guided` sentence (design.md §C.7).
- Widening `internal/mission`'s exported types to fit the contract (design.md §D).
- Making the sign deny conditional on the verb existing — the deny is what makes it unreachable
  (AC-AP-008).
- Asserting wrapper coverage from the list alone, without the unknown-wrapper case (AC-AP-010).
- Running the full local suite: change-scoped packages here, CI for the full verdict
  (`CLAUDE.local.md` §4).

## §H — Risks

| Risk | Mitigation |
|---|---|
| A1 renames or drops `push_requires_lease` | Pre-flight step 3 re-checks it; a change stops M1 and is reported |
| A1's autonomy template block has not landed when M3 is reached | M3 waits and reports; it does not create A1's block |
| A wrapper list that looks complete but skips options wrongly | design.md §C.3 states each wrapper's own options; AC-AP-009 asserts the resolved program is named |
| The receipt path stays unusable through A3 | Stated as residual risk (design.md §F), not designed around |
| `internal/cli` package test latency (10-minute timeout) | Scope test runs to the four packages named in acceptance.md §H |

## §I — Open questions for the lead

- **O1** `moai contract decide` — named in the dispatch, exists nowhere, and no track creates it.
  Currently out of scope (spec.md §G). Confirm, or name the owner.
- **O2** `MOAI_FACTORY_ROLE` — does not exist. The guard denies at the tool-call boundary instead,
  which is role-independent and stronger. Confirm that this is the whole intent, or name the
  role distinction wanted (spec.md §C.6, O2).
- **O3** SPEC body register — the dispatch asked for Korean prose *and* for matching the sibling
  SPEC's register. The sibling (SPEC-AUTONOMY-ESCALATION-001) is written in **English** technical
  prose with Korean only in the 「A1 plan-audit 통과본으로 재확인」 dependency tag. This SPEC matched
  the measured sibling register; say the word and it is rewritten in Korean.
- **O4** Audit-line sink — its own file, or shared with `.moai/logs/branch-guard-audit.log`
  (spec.md O4).
