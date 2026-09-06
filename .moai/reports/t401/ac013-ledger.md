# AC-JFM-013 classification ledger — SPEC-JUDGMENT-FIRST-MODE-001 M1 (card t401)

Sweep window: **the matched physical line** (plan.md §F M1).

Sweep command (verbatim):

```
grep -rn -E '\(Recommended\)|\(권장\)' .claude/rules .claude/skills .claude/output-styles \
  | grep -iE '\bfirst\b|첫 |먼저' > .moai/reports/t401/ac013-candidates.txt
```

| measurement | value |
|---|---|
| tree measured | worktree `WT-analysis-pull`, post-M1-edit working tree, parent HEAD `129fe8b88` |
| swept count (`wc -l < ac013-candidates.txt`) | **26** |
| ledger rows below | **26** |
| rows classed `conditioned` | 8 |
| rows classed `unconditioned-by-design` | 18 |
| rows **escalated** (unclassified remainder) | **0** |
| `grep -c 'recommendation_mode' ac013-candidates.txt` | 7 |

Pre-edit baseline in this same tree: swept count **25**. The 26th row is
`branch-origin-protocol.md:26` — the pull-mode branch this milestone added, which
itself matches the sweep selector. It is a product of the edit, not a pre-existing
candidate.

Six of the seven `conditioned` rows carry `recommendation_mode` **on the matched line
itself**. The seventh, `branch-origin-protocol.md:25`, cannot: it is the Frozen
`CONST-V3R5-035` clause text, which REQ-JFM-015 / spec.md §B.2 require to stay verbatim
so the registry `clause:` string still matches and the canary gate stays satisfiable.
It is conditioned by the adjacent pull branch at `:26`. `conditioned` therefore means
*the clause at this coordinate is conditioned on the mode*, and every such row names
where the conditioning text lives.

**A third ground was used, which this ledger's two-class contract does not describe.**
Row 4 (`zone-registry.md:869`) is classed `unconditioned-by-design` on neither a
consequence test nor a membership test, but on an **explicit SPEC exclusion** — REQ-JFM-015,
spec.md §B.2, and plan.md §F M1 each name the file as not-to-be-edited. It is the only row
using that ground. Whether the ledger's contract should become three-class is a SPEC-body
question, recorded here rather than smoothed away.

## Rows

| # | Coordinate | Class | Reason |
|---|---|---|---|
| 1 | `.claude/rules/moai/core/askuser-protocol-reference.md:34` | unconditioned-by-design | Detail companion of `askuser-protocol.md` § Recommendation Placement Principles, paths-scoped to it and loaded only with it. The stub carries the mode branch; the companion states the `push`-branch reasoning it always stated. Carries no independent mandate. |
| 2 | `.claude/rules/moai/core/askuser-protocol-reference.md:49` | unconditioned-by-design | Same companion. This row *is* the adaptive-strength principle the mode axis generalizes, and already describes a label-omitting branch. |
| 3 | `.claude/rules/moai/core/askuser-protocol-reference.md:122` | unconditioned-by-design | Same companion; restates the bias-prevention rule whose home (§ Option Description Standards) AC-JFM-010 requires byte-unchanged. |
| 4 | `.claude/rules/moai/core/zone-registry.md:869` | unconditioned-by-design | The Frozen `CONST-V3R5-035` `clause:` string. REQ-JFM-015 / spec.md §B.2 forbid editing it; conditioning lives at its doctrine site (`branch-origin-protocol.md:25-26`) and its implementing site (`spec-assembly.md:353`), both `conditioned` below. |
| 5 | `.claude/rules/moai/core/askuser-protocol.md:64` | **conditioned** | S1's first-named coordinate (spec.md §B.1). Mode reference on the line. The section-scoped companion criterion AC-JFM-004 does not reach it — AC-JFM-004 is satisfied inside § Recommendation Placement Principles, while `:64` lives in § Socratic Interview Structure. |
| 6 | `.claude/rules/moai/core/askuser-protocol.md:83` | unconditioned-by-design | The § Option Description Standards bias-prevention rule. Two independent reasons: (a) it mandates no first-option label — it constrains *where* a recommendation signal may be carried, so it is not the clause class REQ-JFM-016 names; (b) AC-JFM-010 requires § Option Description Standards byte-unchanged from `ad272be20`, so conditioning it here would fail a regression-guard criterion. ~~Not an S1 coordinate (§B.1 lists `:64`, `:217`, `:245`, `:253`).~~ **[STRUCK]** — see the strike note below. |
| 7 | `.claude/rules/moai/core/askuser-protocol.md:265` | **conditioned** | The other surviving `askuser-protocol.md` S1 row (`:217` at `ad272be20`; renumbered by this milestone's insert). First-Action Sequence Step 2. Mode reference on the line. |
| 8 | `.claude/rules/moai/development/branch-origin-protocol.md:25` | **conditioned** | Doctrine site of Frozen `CONST-V3R5-035`. Text kept verbatim as the `push` branch per REQ-JFM-015 / §B.2; conditioned by the adjacent pull branch at `:26`. |
| 9 | `.claude/rules/moai/development/branch-origin-protocol.md:26` | **conditioned** | The pull-mode branch added by this milestone. Self-conditioning. |
| 10 | `.claude/rules/moai/workflow/archived-agent-rejection.md:127` | unconditioned-by-design | A recovery-procedure code block restating the Socratic option structure; carries no MUST of its own and inherits from `askuser-protocol.md` § Socratic Interview Structure constraint 3, which is now conditioned (row 5). |
| 11 | `.claude/rules/local/ci-watch-protocol.md:98` | unconditioned-by-design | States its mandate "per `.claude/rules/moai/core/askuser-protocol.md`" — an explicit SSOT citation, so it inherits the conditioned SSOT (row 5). Also a `rules/local/` dev-only file that is never distributed. |
| 12 | `.claude/skills/moai/SKILL.md:350` | **conditioned** | Independent `[HARD]` clause: "All AskUserQuestion calls throughout MoAI workflows MUST follow these rules: The first option MUST always be the recommended choice, clearly marked with '(Recommended)' suffix". No SSOT citation, maximal reach — it meets REQ-JFM-016's reachability test. Originally escalated because the file sat outside M1's declared edit surface; **the operator approved the scope expansion**, adding `.claude/skills/moai/SKILL.md` and its template mirror to M1. Mode reference now on the line. Two alternatives were considered and rejected by the operator: replacing the rule with an SSOT citation (changes the behavior of a writer reading the line), and splitting to a separate card. |
| 13 | `.claude/skills/moai/workflows/moai.md:205` | unconditioned-by-design | Descriptive: states that the sync chain surfaces as the first next-step option rather than firing silently. The subject is chaining behavior, not a label mandate. |
| 14 | `.claude/skills/moai/workflows/run.md:97` | unconditioned-by-design | Same chaining statement as row 13, one file over. Descriptive, not a mandate. |
| 15 | `.claude/skills/moai/workflows/run.md:137` | **conditioned** | The Implementation Kickoff Approval `[HARD]` clause (spec.md §E.1). Mode reference on the line; the gate stays mandatory and score-independent in both modes. |
| 16 | `.claude/skills/moai/workflows/harness-build-entry.md:75` | unconditioned-by-design | Procedural step restating the Socratic composition rules; inherits the conditioned SSOT (row 5). No independent MUST. |
| 17 | `.claude/skills/moai/workflows/harness-build-entry.md:120` | unconditioned-by-design | Same file; a canonical-pattern example, not a mandate. |
| 18 | `.claude/skills/moai/workflows/harness.md:75` | unconditioned-by-design | Procedural instruction for one specific gate; no independent MUST. |
| 19 | `.claude/skills/moai/workflows/harness.md:131` | unconditioned-by-design | Names which option is first for a specific verb (`rollback` vs `apply`); the subject is option ordering for that gate, not the label rule. |
| 20 | `.claude/skills/moai/workflows/harness.md:190` | unconditioned-by-design | Carries a MUST but states it "per `.claude/rules/moai/core/askuser-protocol.md` § Option Description Standards" — an explicit SSOT citation, so it inherits. |
| 21 | `.claude/skills/moai/workflows/feedback.md:124` | unconditioned-by-design | Descriptive statement about one concrete option set; no mandate. |
| 22 | `.claude/skills/moai/workflows/plan/spec-assembly.md:212` | **conditioned** | The same Implementation Kickoff Approval clause as row 15, one file over (spec.md §E.1). Mode reference on the line. |
| 23 | `.claude/skills/moai/workflows/plan/spec-assembly.md:353` | **conditioned** | The sole implementing site of Frozen `CONST-V3R5-035` (the Phase 13 BODP gate). Mode reference on the line, so the doctrine site (row 8) and its implementation are conditioned together and the change is not half-applied. |
| 24 | `.claude/skills/moai/workflows/run/mode-orchestration.md:79` | unconditioned-by-design | Descriptive phase note about chaining, same subject as rows 13-14. |
| 25 | `.claude/skills/hns-workflow-ci-loop/SKILL.md:198` | unconditioned-by-design | A checklist item in a dev-only, non-distributed skill; inherits the conditioned SSOT (row 5). |
| 26 | `.claude/output-styles/moai/moai.md:65` | unconditioned-by-design **in M1** | Restatement of the Socratic composition rules in the output style; inherits the conditioned SSOT (row 5). This file is M2's declared edit surface (S2-S5 banner rules), so any conditioning it needs belongs to that milestone, not this one. |

## Resolutions — both open items closed by orchestrator decision

**Row 12 — resolved by an approved scope expansion.** The row was originally escalated as an
unclassified remainder: classing it `conditioned` needed an out-of-scope edit, and classing it
`unconditioned-by-design` would have rested on scope absence rather than a design ground. The
operator approved adding `.claude/skills/moai/SKILL.md` and its template mirror to M1's edit
surface, on the same reachability basis §E.1 uses to admit `run.md:137` and
`spec-assembly.md:212`. The line is now conditioned under minimum-scope discipline: the
`[HARD] Beginner-Friendly Option Design:` block header, its bullet order, and its second bullet
are untouched.

One correction to the escalation's own record: it stated the mirror "was byte-identical to the
live file at `ad272be20`". **That was wrong, and it was an inference rather than a measurement.**
Measured at `ad272be20`, the pair carries 19 pre-existing intentional differences — the live copy
uses `${CLAUDE_SKILL_DIR}` where the template uses literal `.claude/skills/moai/` paths, and the
live copy carries a `Last Updated:` line the template deliberately lacks. Mirroring by `cp` would
have destroyed all 19 and injected an internal date into the template, violating REQ-JFM-022. The
edit was therefore applied to each copy by hand. This is the only one of the six M1 file pairs that
is not byte-identical.

**Row 6 — membership clause struck by orchestrator adjudication.** The reason originally closed
with "Not an S1 coordinate (§B.1 lists `:64`, `:217`, `:245`, `:253`)", which is a **membership
test** — the classification `plan.md` forbids a run phase from making. It is struck above rather
than deleted, so the record shows a strike.

Ground (a) survives on a consequence test, per the orchestrator's reading of
`askuser-protocol.md:74`–`:83`: the line is the **Bias prevention** clause of § Option Description
Standards. It does not mandate that a first-option label exist; it governs where a recommendation
signal may live when one exists, barring the description from carrying it. Under `pull` the label
is withheld and no option carries a preference claim, so there is no signal for the clause to
misplace — it is satisfied vacuously rather than contradicted. Under `pull` plus on-request
emission the labelled form returns and the clause applies unchanged. It therefore does not reach
the pull-mode behavior.

**Rows 4 and 26 — recorded, not reclassified.** Row 26 (`moai.md:65`) stays hedged as
`unconditioned-by-design **in M1**`, deferred to M2 and not settled. Row 4's third ground is named
in the preamble above.
