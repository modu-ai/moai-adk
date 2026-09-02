# SPEC Review Report: SPEC-BACKLOG-HYGIENE-001 (card t332)

Iteration: 1/2 (Tier M ceiling per `harness.plan_audit_tier_ceilings`)
Verdict: **FAIL**
Overall Score: **0.74** (harmonic mean of the 4 dimensions; arithmetic 0.75). Tier M PASS threshold is **0.80**.

Reasoning context ignored per M1 Context Isolation. Artifacts audited: `spec.md`, `plan.md`,
`acceptance.md` (the Tier M input contract) plus the two supporting evidence files under
`.moai/reports/t332/`. Tree: worktree `.claude/worktrees/t332`, HEAD `15453140a`, branch
`WT-backlog-hygiene` (`rev-parse --show-toplevel` / `--short HEAD` / `branch --show-current`, this run).

FAIL is driven by **one must-pass failure (MP-3)**, which the M5 firewall makes score-independent,
and is independently confirmed by the aggregate sitting below the Tier M threshold.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency.** `grep -o 'REQ-BH-[0-9]\{3\}' spec.md | sort -u` → exactly
  `REQ-BH-001 … REQ-BH-023`, 23 ids, sequential, zero-padded to 3. `grep -c '^\*\*REQ-BH-'` → `23`
  definition heads; `grep -o '^\*\*REQ-BH-[0-9]\{3\}\*\*' | sort | uniq -d` → empty. No gap, no duplicate.

- **[PASS] MP-2 GEARS/EARS format compliance.** Judgment made against the **requirement layer**
  (`REQ-BH-*` in `spec.md` §D). The Given-When-Then entries in `acceptance.md` are the correct
  verification-layer format and are graded under Group 4, not here. All 23 requirement entries match a
  GEARS pattern: Ubiquitous (001, 002, 005, 008, 010, 014, 015, 016, 019, 021, 022, 023), Unwanted
  `shall not` (003, 006, 011, 017, 020), Event-driven `When …` (004, 012, 018), State-driven `While …`
  (009 — `spec.md:L234`, "While the installed binary lacks the `worktree_base_branch` string … the
  sweep shall not cite"), Where (007, 013). Two semantic deviations are recorded as minor defects
  (D7, D8); neither breaks the pattern match.

- **[FAIL] MP-3 YAML frontmatter validity.** All 12 canonical fields are present under canonical names
  (no rejected snake_case alias), but one carries a value outside its enum: `spec.md:L12` reads
  `lifecycle: spec-first`. The SSOT
  (`.claude/rules/moai/development/spec-frontmatter-schema.md` § Field Reference) defines `lifecycle`
  as `enum spec-anchored | spec-lite | exploratory`. `spec-first` is not a member. Per MP-3 ("Type
  mismatch = FAIL") this is a must-pass failure.

  Notably the mechanical lint cannot catch it: `internal/spec/lint.go:765` carries
  `{"lifecycle", fm.Lifecycle}` inside a presence-only loop, and
  `grep -n 'spec-anchored\|spec-lite\|exploratory' internal/spec/lint.go` returns **no output** — the
  rule tests non-emptiness, never membership. A scoped `moai spec lint` would have reported this SPEC
  clean. That is a §1.3 continued-firing blind spot, not evidence of validity.

- **[N/A] MP-4 Section 22 language neutrality.** The SPEC is scoped to this project's own backlog-queue
  tooling and names no per-language toolchain — no `gopls` / `pylsp` / `rust-analyzer`-class tool name
  appears in any artifact. Single-project, single-language scope ⇒ N/A, auto-passes.

- **[PASS] MP-5 D7 cross-SPEC reconciliation.** `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u`
  → 4 ids (self + 3 references). All three referenced directories exist under `.moai/specs/`. Statuses
  read this run: `SPEC-TODO-LANDING-STATE-001` → `completed`; `SPEC-TODO-ANALYSIS-001` → `completed`;
  `SPEC-KANBAN-QUEUE-PR-SYNC-001` → `in-progress`. None is `retired` / `superseded` / `archived`, so no
  reconciliation is required and no BLOCKING finding is emitted.

- **[PASS] MP-6 D8 cross-platform discipline.** `grep -c 'syscall'` over all four SPEC artifacts → `0`
  for each (`spec.md`, `plan.md`, `acceptance.md`, `progress.md`). D8-4 auto-PASS.

- **[PASS] MP-7 clarification gate.** `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-BACKLOG-HYGIENE-001/`
  → exit 1, no matches. `plan.md` exists; `research.md` is absent by Tier M design. Note (not a
  failure): `spec.md` §F carries two Open Questions — Q1 routed to milestone M1, Q2 (Tier) routed to the
  operator. Neither uses the `[NEEDS CLARIFICATION: …]` marker, so the gate does not fire mechanically;
  Q2 is folded into D2.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.85 | 0.75 | Every requirement carries a single reading; no pronoun ambiguity. Two `Where`-pattern requirements express a runtime condition rather than a capability gate (`spec.md:L223`, `L258`), and REQ-BH-010 (`spec.md:L239`) is passive-voiced ("A `landed` verdict shall be established by…") so the acting subject is unnamed. Matches "minor ambiguity in one or two requirements". |
| Completeness | 0.72 | 0.75 (lower edge) | All required sections present: HISTORY (`spec.md:L21`), WHY (§A `L27`), WHAT (§C `L151`), REQUIREMENTS (§D `L195`), ACCEPTANCE CRITERIA (`acceptance.md` §B), and **five** `### Out of Scope — <topic>` H3 sub-headings at `spec.md:L165,171,178,183,190`, each carrying specific `-` bullets. Frontmatter is complete in field set but carries one out-of-enum value (MP-3). Two substantive omissions: no ref-refresh before the landing queries (D1), and §E's write boundary contradicts REQ-BH-007 (D3). |
| Testability | 0.80 | 0.75 | Weasel-word scan over `acceptance.md` (`appropriate\|adequate\|reasonable\|proper\|as needed\|if necessary\|sufficiently`, case-insensitive) → exit 1, **zero matches**. Most ACs name a decidable command, and several carry an explicit anti-vacuity conjunct — AC-BH-004 (`L43`, "an empty log also returns 0 matches, and would pass vacuously"), AC-BH-002 (`L28`, "a count alone passes against 67 entries for the wrong cards"), AC-BH-007's positive control (`L71`), plus three stated mutation probes (`L20`, `L61`, `L87`). Deductions: AC-BH-007 is vacuous on the path-B branch (D4); AC-BH-002 and AC-BH-010 rest partly on unspecified extraction / "not a paraphrase" judgement with no stated rubric. |
| Traceability | 0.60 | 0.50 | **Multiple REQs lack a deciding AC**: REQ-BH-004 (delta re-derivation branch), REQ-BH-013 (the *positive* direction — no AC asserts that a card with a live worktree IS classified `in-flight-unlanded`; AC-BH-008 quantifies over entries already carrying the verdict and passes vacuously on an empty set), REQ-BH-014 (per-card premise restatement + three-valued verdict — AC-BH-002 counts entries, AC-BH-010 covers `falsified` only), REQ-BH-016 (five evidence sections on **every** entry — AC-BH-010 requires them only for `falsified`), REQ-BH-018 (`unverified` with reason — no AC; AC-BH-009 covers landing `unknown`, a different axis). Conversely AC-BH-013 (write isolation) traces to plan.md M3 and spec.md §E, not to any `REQ-BH-*`. Rubric band 0.50; scored at the band's upper edge because each has partial adjacent coverage. |

Counts were verified independently, not taken from the prose.
`wc -l .moai/reports/t332/queue-snapshot.tsv` → `100`. `cut -f2 … | sort | uniq -c` → `18 dropped`,
`10 picked`, `68 queued`, plus 4 relation rows. `awk -F'\t' '$2=="queued" && $1!="t332"' | wc -l` →
**67**. `awk -F'\t' '$1=="t332"'` → status `queued`. The picked id list at `spec.md:L90` matches
`awk -F'\t' '$2=="picked"{print $1}'` exactly: `t278 t333 t338 t341 t346 t350 t354 t356 t357 t358`.
Every count stated in §B.1, §C, REQ-BH-021, and AC-BH-002 is therefore correct.

---

## Defects Found (structured defect-list)

**D1. LANDING-REFS-UNFETCHED — spec.md:L239-L247 (REQ-BH-010), plan.md:L66 (M1 path A) — Severity: major — Class: blocking**

*Claim*: every landing verdict can be silently wrong because the refs it queries are never refreshed
and never pinned.
*Evidence*: `grep -n 'fetch' spec.md plan.md acceptance.md` → exit 1, **zero matches**. The method
queries `origin/develop` and `origin/main` as branch names.
*Baseline-attribution*: measured in this worktree at HEAD `15453140a`, this run.
*Gaps*: I did not measure how stale this tree's remote-tracking refs currently are.
*Residual-risk*: remote-tracking refs go stale silently, so a card that landed after this worktree's
last fetch reads `not-landed` with no error and no empty-output warning — the same silent-wrong-answer
shape §B.2 quotes `prlink_landed.go` about, reproduced one layer up. The verdict is also unpinned:
`origin/develop` moves under the sweep, so two cards read at different moments are measured against
different trees (`verification-completeness.md` §4). This is the pin-it case, not the provenance
carve-out — the claim is per-card, not about the mainline's own lineage.
*Required fix*: add to M1 a single refresh of both refs; record its time and the resulting
`rev-parse origin/develop origin/main` SHAs in `00-tooling-baseline.md`; amend REQ-BH-010 so every
landing verdict cites those two pinned SHAs rather than the branch names; extend AC-BH-006 to require
the pinned ref SHA alongside the commit SHA and the `--is-ancestor` exit code.

**D2. TIER-BUDGET-EXCEEDED — plan.md:L9, spec.md frontmatter L15 — Severity: major — Class: blocking**

*Claim*: the Tier M justification rests on a misread ceiling table.
*Evidence*: `plan.md:L9` grounds Tier M on "Requirement count 23 … M (S ceiling is smaller; L threshold
is ~25)". `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (REQ/AC budget) sets the
**Tier M requirement ceiling at 16**, Tier L at 25. Measured: 23 requirements (`grep -c '^\*\*REQ-BH-'`),
14 acceptance criteria (`grep -c '^### AC-BH-'`).
*Baseline-attribution*: both counts measured in this tree, this run; ceilings read from the SSOT rule.
*Gaps*: none — both sides of the comparison are mechanical.
*Residual-risk*: per the SSOT an over-budget count "is a signal to tier up or to split the SPEC, not to
relax the budget". The card's own subject is premises that were never re-measured; its tier
justification is one.
*Required fix*: consolidate the requirement set to ≤16 and restate the plan.md grounds row against the
actual ceilings. Consolidation, not a tier-up — see § Tier Judgement.

**D3. WRITE-BOUNDARY-CONTRADICTION — spec.md:L306 (§E) vs spec.md:L223-L226 (REQ-BH-007) — Severity: major — Class: blocking**

*Claim*: two clauses of the SPEC cannot both be obeyed.
*Evidence*: §E states "No file outside `.moai/specs/SPEC-BACKLOG-HYGIENE-001/` and
`.moai/reports/t332/` is written." REQ-BH-007 mandates `moai todo relate`, which writes the queue
store: `internal/cli/todo_relate.go:66` — `newTodoStore().Mutate(func(rec *kanban.BacklogRecord) error {…})`
— appending a finding through the same `Mutate` write path a card mutation uses, into `backlog.json`
(`internal/kanban/state_dir.go:146` `backlogFileName = "backlog.json"`, located by `BacklogPathForRoot`
at `state_dir.go:122` under the project state dir).
*Baseline-attribution*: source read in this tree at `15453140a`, this run.
*Gaps*: I did not execute `moai todo relate`, so the write is established from the code path, not
observed.
*Residual-risk*: `acceptance.md:L137` §D cannot adjudicate the conflict either — the state dir is
gitignored, so `git status --short` shows nothing whichever way a worker resolved it.
*Required fix*: amend §E with an explicit carve-out — "plus the queue store, written only through
`moai todo relate`, which appends a finding and changes no card" — and give the carve-out an observable
in §D so the permitted write is distinguishable from a forbidden one.

**D4. AC-BH-007-VACUOUS-ON-PATH-B — acceptance.md:L63-L72 — Severity: major — Class: blocking**

*Claim*: a mutant satisfies AC-BH-007 while violating REQ-BH-009.
*Evidence*: the binding conjunct is guarded — "**if** the recorded value is `0` **and path A was
chosen**, no card entry cites `moai todo pr` as the basis of a landing verdict." The mutant: record path
B, skip the rebuild, cite the installed `moai todo pr` landed column anyway. The antecedent is false,
the criterion passes, and REQ-BH-009's `While` condition still holds because the binary was never
rebuilt. Re-measured this run: `strings ~/go/bin/moai | grep -c 'worktree_base_branch'` → **0**; this
session's own MCP server banner independently reports the build as `v3.1.2 (commit 343399d2f, built
2026-08-27T14:07:38Z)`.
*Baseline-attribution*: both measurements taken in this session, against the installed binary at
`~/go/bin/moai`, 2026-08-29.
*Gaps*: the mutant was constructed on paper, not executed — the artifacts it would mutate do not exist
until the run phase.
*Residual-risk*: this is the one criterion standing between the run phase and the exact error §B.3 was
written to prevent.
*Required fix*: guard on the **measured** value rather than the declared path — "if the recorded
`strings` count is `0`, no card entry cites `moai todo pr`, whichever path was chosen" — and add a
path-B conjunct: if path B was chosen, the file records the post-rebuild `strings` count as **≥ 1**
and the reinstall command that produced it.

**D5. NO-MUTATION-OBSERVABLE-IS-SELF-AUTHORED — acceptance.md:L36-L52 (AC-BH-004, AC-BH-005) — Severity: major — Class: blocking**

*Claim*: the no-mutation invariant is stated and checkable, but both observables are weaker than a
hash and they fail in the same direction.
*Evidence*: the invariant IS a requirement (REQ-BH-005/006/008) and IS made mechanically checkable —
AC-BH-004 is a verb census over the evidence log, AC-BH-005 re-reads the queue composition. But
AC-BH-004's subject, `$R/invocations.log`, is authored by the very run phase whose restraint it
certifies: a worker that mutated and omitted the line passes. And AC-BH-005 compares
queued/picked/dropped **counts** — `moai todo edit` changes a card's text while leaving all three
identical, so the count comparison cannot see the mutation the log was supposed to catch.
*Baseline-attribution*: read from `acceptance.md` at `15453140a`; the `edit`-leaves-counts-identical
property follows from the verb set enumerated in REQ-BH-006 and the census in `01-scope.md`'s stated
form.
*Gaps*: I did not enumerate every `moai todo` mutating verb against its effect on the three counts;
`edit` alone is sufficient to establish the hole.
*Residual-risk*: two checks that fail on the same card in the same direction read as defense in depth
and are not.
*Required fix*: add a third, independently-derived observable — a digest over the card rows only
(id + state + text, relation findings excluded, since REQ-BH-007's permitted `relate` legitimately
appends to `Findings`), captured at M2 into `01-scope.md` and re-captured at M5, with a new AC
requiring the two digests to be byte-identical. Keep AC-BH-004 and AC-BH-005 — three observables that
fail independently is the point.

**D6. FIVE-REQS-WITHOUT-A-DECIDING-AC — spec.md:L209 (004), L258 (013), L263 (014), L270 (016), L277 (018) — Severity: major — Class: blocking**

*Claim*: five requirements have no criterion that can red them.
*Evidence*: enumerated with their gaps in the Traceability row above. Sharpest two: REQ-BH-016 requires
the five evidence sections on **every** card entry while AC-BH-010 scopes them to `falsified` entries —
a report whose 60 `holds` entries carry a bare verdict satisfies every criterion. REQ-BH-013's positive
direction is uncovered — AC-BH-008 quantifies over entries already reading `in-flight-unlanded`, so a
sweep that never assigns the classification passes on the empty set.
*Baseline-attribution*: REQ↔AC mapping performed by hand over `spec.md` §D and `acceptance.md` §B at
`15453140a`.
*Gaps*: the mapping is a judgement, not a mechanical extraction — see Residual risk.
*Residual-risk*: an auditor who counted REQ-BH-014/016/018 as covered in substance by AC-BH-010 would
score Traceability near 0.80.
*Required fix*: extend AC-BH-010's five-section conjunct from `falsified` entries to all entries; add
one AC asserting every entry carries a premise verdict in `{holds, falsified, unverified}` with
`unverified` entries carrying a reason (covers 014 + 018); attach an `in-flight-unlanded` cross-check
against the live worktree list to AC-BH-008 (covers 013's positive direction); give REQ-BH-004's delta
branch a criterion, even if it reads "no delta was observed, and `01-scope.md` records the comparison".

**D7. WHERE-PATTERN-USED-FOR-A-RUNTIME-CONDITION — spec.md:L223 (REQ-BH-007), L258 (REQ-BH-013) — Severity: minor — Class: optional**

GEARS reframes `Where` as capability gate / feature flag / static config. Both requirements use it for
a per-card runtime condition discovered during the sweep ("Where an overlap … is confirmed by
measurement"; "Where a queued card has a live worktree"), which is event-driven or state-driven
territory. The syntactic pattern matches, so MP-2 is unaffected; the semantics mislead a reader about
when the clause fires.
*Required fix*: recast as `When an overlap between two queued cards is confirmed by measurement, the
sweep shall record it with …` and `While a queued card has a live worktree or an unmerged branch, the
sweep shall classify it …`.

**D8. PASSIVE REQ-BH-010 NAMES NO ACTOR — spec.md:L239 — Severity: minor — Class: optional**

"A `landed` verdict shall be established by querying git directly …" — the GEARS `<subject>` is the
verdict rather than the actor, so the requirement does not say who performs the query. Every sibling
requirement names `the sweep`.
*Required fix*: "The sweep shall establish a `landed` verdict by querying git directly against both
integration refs, and shall cite …".

**D9. M3 BATCH SIZES MISSTATED — plan.md:L90-L91 — Severity: minor — Class: optional**

plan.md states "split into 6 batches of 11-12". Computed from the in-scope list this run (awk bucketing
over the six stated id ranges): **B1=10, B2=11, B3=13, B4=10, B5=13, B6=10, total=67**. The total is
right; the per-batch figure is not, and two batches exceed the stated upper bound. Harm is bounded
because plan.md:L92 already says membership "is re-derived from M2's list, never retyped from here".
*Required fix*: state the measured sizes, or drop the per-batch figure and keep the re-derivation
instruction alone.

**D10. UNSUPPORTED EMBEDDED-NEWLINE CAVEAT — spec.md:L133-L135 — Severity: minor — Class: optional**

§B.4 annotates the `no-link` figure as "91 (line count, not row count: card text contains embedded
newlines)". The arithmetic in the same evidence root contradicts it: `tooling-baseline.md` records
`5 landed` + `91 no-link` = 96, and 96 is exactly the card count (100 snapshot rows − 4 relation rows =
68 + 10 + 18). If embedded newlines were inflating the line count the two figures could not coincide.
The caveat is an unverified hedge inside the one section the SPEC wrote to correct an unverified claim.
*Required fix*: cite the command that demonstrates embedded newlines, or replace the parenthetical with
the arithmetic showing 91 is the row count.

**D11. t256 IS DROPPED, NOT MERELY UNQUEUED — plan.md:L111 — Severity: minor — Class: optional**

M4 says of the existing `t318 ↔ t256` relation "t256 is not in the queued set — verify before
re-recording". Measured: `awk -F'\t' '$1=="t256"'` → `t256  dropped`. Being **dropped** engages a second
exclusion — `spec.md:L180`, "The 18 `dropped` rows are already decided… an un-drop is never proposed" —
which the M4 note does not mention, leaving open whether the relation may be re-recorded against a
decided card at all.
*Required fix*: state in M4 that a relation whose counterpart is `dropped` is reported as a reading only
and never carried into the disposition proposal.

---

## Regression Check

Not applicable — iteration 1.

---

## Tier Judgement (independent of the author's proposal)

**Tier M is the right complexity classification, but this SPEC does not currently fit inside it.**

The complexity signals genuinely read M: no source file lands, no schema moves, the artifact set is the
3-file Tier M set plus `progress.md` (`ls` confirms exactly `spec.md`, `plan.md`, `acceptance.md`,
`progress.md`), and plan.md's argument that there is nothing to design is correct — the deliverable is a
reading, so a `design.md` would carry nothing plan.md does not already carry. Artifact volume does not
argue S either: 67 cards, a six-way fan-out, and a tooling-baseline question gating every landing
verdict is more than one milestone's work.

What does not fit is the requirement count: 23 against Tier M's ceiling of 16 (D2). I do **not**
recommend tiering up to L to accommodate it. Tier L would buy a `design.md` and a `research.md` this
card has no content for, plus a stricter 0.85 threshold, purely to legalize a requirement count that is
inflated rather than substantive. Several requirements are restatements at a different layer:
REQ-BH-003 is a special case of REQ-BH-001; REQ-BH-015/016/017 are three faces of one
verification-claim-integrity obligation; REQ-BH-020 restates REQ-BH-006 in the overlap context;
REQ-BH-023 restates REQ-BH-005/006 at the report layer. Consolidating those recovers the budget without
losing a single obligation.

**Recommendation: keep `tier: M`, reduce the requirement set to ≤16 by consolidation, and record the
consolidation in HISTORY** so the reduction reads as a merge rather than as a deletion.

---

## Recommendation

FAIL. Route the six **blocking** defects (D1-D6) back to manager-spec; D7-D11 are **optional** and left
to the orchestrator's discretion — a long optional list does not by itself justify a FAIL and was not
used to manufacture one. The FAIL rests on MP-3 plus the 0.74 aggregate against a 0.80 threshold.

Ordered fix instructions:

1. **`spec.md:L12`** — change `lifecycle: spec-first` to `lifecycle: spec-anchored`. This alone clears
   MP-3. Note for whoever fixes it: the lint will not confirm the repair, because it never tested the
   enum — verify by eye against the schema SSOT's Field Reference table.
2. **`spec.md` §D / `plan.md` §A** — consolidate to ≤16 requirements per D2, and restate `plan.md:L9`'s
   grounds row against the real ceilings (M=16, L=25).
3. **`spec.md:L239` REQ-BH-010 + `plan.md` M1 + `acceptance.md` AC-BH-006** — add the ref refresh, pin
   both ref SHAs, require the pinned SHA in every landing citation (D1).
4. **`spec.md:L306` §E** — add the queue-store carve-out for `moai todo relate` (D3).
5. **`acceptance.md:L63` AC-BH-007** — re-guard on the measured `strings` value rather than the declared
   path, and add the path-B post-rebuild conjunct (D4).
6. **`acceptance.md`** — add the card-row digest AC (D5) and the coverage ACs (D6). Naive addition would
   take the count from 14 to ~19, over Tier M's AC ceiling of 16 — fold the coverage additions into
   existing criteria (extend AC-BH-010's scope rather than adding a sibling; attach the
   `in-flight-unlanded` cross-check to AC-BH-008) so the count lands at ≤16.
7. Optional, operator's call: D7-D11.

Re-audit on iteration 2 is scoped to this enumerated defect delta plus a regression check over D1-D6.

---

## Gaps — what this audit did NOT observe

- **No lint signal.** `moai spec lint` was not run in any form. The corpus form was prohibited by the
  dispatch (an earlier run exceeded 120s; another session is running one), and I found no single-SPEC
  scoped form to substitute — recorded as a Gap rather than a clean result. MP-3's finding was reached
  by reading the schema SSOT directly; the lint source inspection (`internal/spec/lint.go:765`)
  indicates the lint would not have reported it anyway.
- **No git-history verification.** I did not verify `260ea5369`, the two t342 control queries, or
  `moai todo pr`'s output. Compound commands carrying `git` were refused by this worktree session's
  command guard, and re-running `moai todo pr` risks touching the queue. §B.2 / §B.3 / §B.4 are consumed
  as the orchestrator's stated evidence, not as observations of mine. Two exceptions I re-measured
  independently: `strings ~/go/bin/moai | grep -c 'worktree_base_branch'` → `0`, and this session's MCP
  banner (`v3.1.2`, `343399d2f`, built 2026-08-27T14:07:38Z).
- **`git worktree list` not run**, so §B.5's claim that queued cards t294/t299/t318/t336/t337/t343/t362
  hold live worktrees is unverified here. It bears on D6's REQ-BH-013 gap — but that gap was derived
  from the AC text and does not depend on the claim.
- **The 67 in-scope cards themselves were not read.** This audit judges the SPEC, not the queue. I
  verified the counts and the id sets; I did not sample a card to test whether M3's per-card procedure
  is executable within a worker's budget.
- **Snapshot capture provenance not verified.** I read `queue-snapshot.tsv` and confirmed its internal
  arithmetic; I did not confirm it was captured at 2026-08-29 20:36 from `moai todo`, nor that the queue
  has not moved since. `01-scope.md` (an M2 output) does not exist yet, so AC-BH-001's own preconditions
  are not yet observable.
- **No mutation testing of the ACs.** The three probes stated in `acceptance.md` (L20, L61, L87) were
  read, not executed — the artifacts they would mutate do not exist until the run phase. D4 and D6's
  vacuity findings were reached by constructing mutants on paper, per the mutant-probe discipline.
- **`harness.plan_audit_tier_ceilings` not read**; the "1/2" iteration label assumes the documented
  Tier M ceiling of 2.

## Residual risk

The MP-3 failure is a one-word repair, which creates a real hazard: a fix round that changes
`spec-first` to `spec-anchored`, sees the must-pass table go green, and treats D1-D6 as advisory. It
must not. D1 (unrefreshed, unpinned refs) and D5 (self-authored no-mutation observable) each
independently permit this sweep to produce a confidently-wrong reading — the exact failure the card was
issued to stop — and neither is visible in any output the sweep itself generates. D4's vacuity is worse
in kind: it would let a run phase cite the very column §B.3 proves is answering about the wrong branch,
and pass its own acceptance criteria while doing so.

Second: the 0.74 aggregate rests on a Traceability score of 0.60 derived by mapping 23 requirements onto
14 criteria by hand. That mapping is a judgement, and a reasonable auditor who counted
REQ-BH-014/016/018 as covered in substance would land near 0.80 and score the SPEC at the threshold. The
MP-3 failure decides the verdict independently of where that argument comes out, which is the only
reason I am comfortable reporting the aggregate at all.
