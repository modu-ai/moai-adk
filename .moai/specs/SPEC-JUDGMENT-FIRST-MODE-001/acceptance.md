# SPEC-JUDGMENT-FIRST-MODE-001 — Acceptance Criteria

Requirements live in `spec.md` §C (GEARS `REQ-JFM-xxx`). This file is the **verification layer**:
Given-When-Then criteria, each adopted as a **two-cell pair** — a `RED-now:` cell observed on the
pre-implementation tree, and a `Green path:` cell naming the milestone that flips it
(`verification-completeness.md` §2).

**RED-now measurement pin.** All cells below were measured on 2026-09-02 in the worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t401`, branch `WT-analysis-pull`, tree
`ad272be20`. `git rev-parse HEAD` → `ad272be20abff9e4f3b1b363fce3e48dac4c5132`;
`git status --porcelain` → two untracked paths only (`.moai/reports/t401/`,
`.moai/specs/SPEC-JUDGMENT-FIRST-MODE-001/`), so every tracked file measured is byte-identical to
`ad272be20`. Each cell records the single-invocation command, its verbatim stdout, its exit code as
a separate field, and why it is red. A figure carried from another tree or an earlier commit is not
a baseline.

**One exception to the pin, stated rather than left implicit.** The provenance cells added to
AC-JFM-018 and AC-JFM-023 by the collection-owner amendment were measured later, in the same
worktree and branch — first measured at `6352897a5`, re-measured after the 0.2.3 fold at
`095f2799b`; each cell carries its own inline pin (criterion-level pins govern). No cell measured
at `ad272be20` was re-labelled with the newer tree.

**Severity classes.**

- **release-blocking** — red now, flipped by this work. Holds the SPEC open.
- **regression-guard** — green now, and must stay green. Measured today as a baseline; it proves
  the change did not break something, never that the change happened. A regression guard is
  **not** a release blocker: as one it would assert nothing about this work.
- **non-blocking** — a quality signal; verified or recorded as an open gap.

## §D AC Matrix

| AC | REQ | Severity | Two-cell | Axis |
|---|---|---|---|---|
| AC-JFM-001 | REQ-JFM-002, REQ-JFM-004 | regression-guard | baseline-green | template default neutrality |
| AC-JFM-002 | REQ-JFM-001, REQ-JFM-003 | release-blocking | RED-now + green(M3) | config schema |
| AC-JFM-003 | REQ-JFM-001 | non-blocking | RED-now + green(M3) | config round-trip |
| AC-JFM-004 | REQ-JFM-005 | release-blocking | RED-now + green(M1) | S1 doctrine |
| AC-JFM-005 | REQ-JFM-006 | release-blocking | RED-now + green(M2) | S2 / S3 banners |
| AC-JFM-006 | REQ-JFM-007 | non-blocking | RED-now + green(M2) | S4 Insight slot |
| AC-JFM-007 | REQ-JFM-008 | non-blocking | RED-now + green(M2) | S5 Error Recovery order |
| AC-JFM-008 | REQ-JFM-009 | release-blocking | RED-now + green(M2) | S6 `/clear` |
| AC-JFM-009 | REQ-JFM-010 | release-blocking | RED-now + green(M1) | on-request emission |
| AC-JFM-010 | REQ-JFM-011, REQ-JFM-014 | regression-guard | baseline-green | evidence obligations intact |
| AC-JFM-011 | REQ-JFM-012, REQ-JFM-013 | non-blocking | RED-now + green(M1) | adopted conditions present |
| AC-JFM-012 | REQ-JFM-015 | regression-guard | baseline-green | Frozen `clause:` untouched |
| AC-JFM-013 | REQ-JFM-016 | release-blocking | RED-now + green(M1) | label-mandate sweep ledger |
| AC-JFM-014 | REQ-JFM-017 | release-blocking | RED-now + green(M0) | observer emits rows |
| AC-JFM-015 | REQ-JFM-018 | release-blocking | RED-now + green(M0) | observer never denies |
| AC-JFM-016 | REQ-JFM-018 | release-blocking | RED-now + green(M0) | observer fails open |
| AC-JFM-017 | REQ-JFM-019 | release-blocking | RED-now + green(M4) | static CI guard, both directions |
| AC-JFM-018 | REQ-JFM-020, REQ-JFM-025 | release-blocking | RED-now + green(M6) | **vacuity falsifier** |
| AC-JFM-019 | REQ-JFM-021 | release-blocking | RED-now + green(M5) | template mirror parity |
| AC-JFM-020 | REQ-JFM-022 | regression-guard | baseline-green | template neutrality |
| AC-JFM-021 | REQ-JFM-023 | regression-guard | baseline-green | `.sh` / `.sh.tmpl` pair |
| AC-JFM-022 | REQ-JFM-002 | regression-guard | baseline-green | affected-package suite |
| AC-JFM-023 | REQ-JFM-024, REQ-JFM-025 | release-blocking | RED-now + green(M0) | **detector fires — positive control** |

23 criteria: **13 release-blocking**, 6 regression-guard, 4 non-blocking.

## §D.1 Config schema

**AC-JFM-001 — regression-guard** — *Given* a consumer project that does not set
`recommendation_mode`, *When* the template's `interview.yaml` is compared to base, *Then* the only
change is an additive, commented, `push`-valued key, and behavior for that consumer is unchanged.

- Verify: `git diff ad272be20 -- internal/template/templates/.moai/config/sections/interview.yaml`
- Baseline (this run, tree `ad272be20`): stdout empty; exit code `0`; `| wc -l` → `0`. The file is
  untouched, which is the baseline this guard protects — the diff must stay additive-only, never
  become a behavior change for the default consumer.
- Why this is not release-blocking: it is green at arrival, so as a blocker it would assert
  nothing about this work.

**AC-JFM-002 [release-blocking]** — *Given* `recommendation_mode: bogus`, *When* the config is
loaded, *Then* the load succeeds, the resolved mode is `push`, and the unrecognized value is
recorded verbatim.

- Verify: `go test ./internal/config/... -run 'RecommendationMode' -count=1 -v`
- RED-now: `go test ./internal/config/... -run 'RecommendationMode' -count=1` →
  `ok  	github.com/modu-ai/moai-adk/internal/config	0.214s [no tests to run]`; exit code `0`. The line shown is the **first of 3**; the selector matches `internal/config`, `config/atomicfile`, and `config/toolpolicy`, and the `[no tests to run]` token holds on all 3 (re-measured, `wc -l` → `3`, `0` `FAIL` lines).
  **Red because the swept set is empty**, not because a test failed: no test matches the selector,
  so the exit-zero is uninterpreted output rather than a pass
  (`verification-completeness.md` §1.1). The `[no tests to run]` token is the evidence.
- Green path: M3 adds `internal/config/interview_recommendation_mode_test.go`; the same command
  under `-v` names at least one `--- PASS: TestRecommendationMode…Unrecognized` line and the
  `[no tests to run]` token is absent.

**AC-JFM-003 — non-blocking** — *Given* a section file setting `recommendation_mode: pull`, *When*
the loader runs, *Then* the resolved mode is `pull`.

- Verify: `go test ./internal/config/... -run 'RecommendationMode.*RoundTrip' -count=1 -v`
- RED-now: the command actually run in this measurement pass was the **superset** selector,
  `go test ./internal/config/... -run 'RecommendationMode' -count=1` →
  `ok  	github.com/modu-ai/moai-adk/internal/config	0.214s [no tests to run]`; exit code `0`. The line shown is the **first of 3**; the selector matches `internal/config`, `config/atomicfile`, and `config/toolpolicy`, and the `[no tests to run]` token holds on all 3 (re-measured, `wc -l` → `3`, `0` `FAIL` lines).
  A selector matching zero tests cannot have a narrowing sub-selector match more, so
  `RecommendationMode.*RoundTrip` sweeps zero as well. Stated as the deduction it is, rather than
  reported as a measurement of the narrower command.
- Green path: M3; a named `RoundTrip` test appears in the `-v` output.

## §D.2 Doctrine surfaces

**AC-JFM-004 [release-blocking]** — *Given* the amended `askuser-protocol.md`, *When* the file is
read, *Then* the `(권장)`-first requirement is expressed as a mode-conditional clause generalizing
the existing Adaptive-strength principle, and no parallel mechanism is introduced.

- Verify: `grep -n 'recommendation_mode' .claude/rules/moai/core/askuser-protocol.md` returns a
  match inside § Recommendation Placement Principles, **and**
  `grep -c 'Adaptive strength' .claude/rules/moai/core/askuser-protocol.md` still returns `1`.
- RED-now: `grep -n 'recommendation_mode' .claude/rules/moai/core/askuser-protocol.md` → stdout
  empty; exit code `1`. Red because the mode axis does not exist in the doctrine yet. The
  companion guard is green today and must stay so:
  `grep -c 'Adaptive strength' …` → `1`; exit code `0` — a second occurrence would mean a parallel
  mechanism was authored instead of the existing principle being generalized.
- Green path: M1; the first command returns ≥1 line inside § Recommendation Placement Principles
  and the second still returns exactly `1`.

**AC-JFM-005 [release-blocking]** — *Given* the amended `output-styles/moai/moai.md`, *When* the
Discovery banner and the Epic Stats / Epic Status banner rules are read, *Then* the
`⏭️ Recommended action:` and `⏭️ Next:` rules each carry a pull-mode withholding branch, and no
other banner field changes.

- Verify: `grep -c 'recommendation_mode' .claude/output-styles/moai/moai.md` ≥ 2, and
  `git diff ad272be20 -- .claude/output-styles/moai/moai.md` touching no field other than the
  named ones.
- RED-now: `grep -c 'recommendation_mode' .claude/output-styles/moai/moai.md` → `0`; exit code `1`.
  Red because neither banner rule carries a mode branch.
- Green path: M2; the count is ≥ 2 with one hit at each of the two banner rules.

**AC-JFM-006 — non-blocking** — *Given* pull mode, *When* the Insight banner rule is read, *Then*
it carries the literal user-judgment field `Your call: [what the reader decides]` after
`Implications:`, inside the same banner frame.

- Verify: `grep -c 'Your call:' .claude/output-styles/moai/moai.md` ≥ 1 and
  `grep -n -A12 'MoAI ★ Insight' .claude/output-styles/moai/moai.md` showing it inside the frame.
- RED-now: `grep -c 'Your call:' .claude/output-styles/moai/moai.md` → `0`; exit code `1`. Red
  because the slot does not exist. (REQ-JFM-007 now fixes the slot's literal shape, so this
  criterion is mechanical rather than a judgment about "an explicit slot".)
- Green path: M2.

**AC-JFM-007 — non-blocking** — *Given* pull mode, *When* the Error Recovery template is read,
*Then* the options are ordered by increasing cost to the user (`Pause` → `Retry as-is` →
`Alt approach` → `Abort+preserve`) and `Retry as-is` is not first.

- Verify: `grep -c 'Pause.*Retry as-is.*Alt approach.*Abort+preserve' .claude/output-styles/moai/moai.md`
  returns ≥ 1 — the **full four-element sequence, in order, on one physical line**. The pattern
  carries no backticks so it survives shell quoting, and `+` is literal in BRE. A
  `Pause`-ahead-of-`Retry as-is` check alone is insufficient: it admits three different orderings
  of the remaining two options, so the sequence would be normative in `spec.md:212` and
  `design.md:89` and unpinned by any criterion (D-N8). Supplementary context read:
  `grep -n 'Recovery options via AskUserQuestion' -A2 .claude/output-styles/moai/moai.md`.
- RED-now: `grep -c 'A. Retry as-is' .claude/output-styles/moai/moai.md` → `1`; exit code `0` —
  the option line at `moai.md:597` still reads
  `A. Retry as-is  B. Alt approach  C. Pause  D. Abort+preserve`. Red because `Retry as-is` is
  first by convention, which is exactly what REQ-JFM-008 forbids.
- Green path: M2; the same grep returns `0` for `A. Retry as-is` and the reordered line is present.

**AC-JFM-008 [release-blocking]** — *Given* pull mode, *When* step 4 of
`context-window-management.md`'s pre-clear announcement is read, *Then* the `/clear` guidance is
stated as a threshold-crossing fact plus available actions, with the recommendation phrasing
conditioned on the mode.

- Verify: `grep -n 'recommendation_mode' .claude/rules/moai/workflow/context-window-management.md`
  returning a match within the Pre-clear announcement clause.
- RED-now: that command → stdout empty; exit code `1`. Red because S6 sits outside the
  AskUserQuestion channel and inherits none of its guards, so the mode is absent there and cannot
  be reached by cross-reference (spec.md §B.1).
- Green path: M2.

**AC-JFM-009 [release-blocking]** — *Given* pull mode, *When* the user explicitly asks for a
recommendation or an analysis, *Then* the doctrine directs the orchestrator to emit the withheld
recommendation in the same form it would carry under `push`.

- Verify: `grep -c '^### On-request emission' .claude/rules/moai/core/askuser-protocol.md` ≥ 1,
  and both banner rules in `moai.md` referencing that section by name.
- RED-now: `grep -c '^### On-request emission' .claude/rules/moai/core/askuser-protocol.md` → `0`;
  exit code `1`. Red because no such clause exists. Measured context, so the red is not mistaken
  for an adjacent clause: `grep -nE 'on request|explicitly request' …` returns exactly one line,
  `:122`, which is the pre-existing Requested-Deliverable Primacy rule about *withholding* a
  question — the opposite direction, and not an emission clause.
- Green path: M1; the heading exists and is referenced from both banner rules.

## §D.3 The three adopted conditions and existing guards

**AC-JFM-010 — regression-guard** — *Given* pull mode, *When* the Report-Before-Ask gate,
Requested-Deliverable Primacy, the neutral-description rule, and the Implementation Kickoff
Approval mandate clause in `orchestration-mode-selection.md` are read, *Then* each is
byte-unchanged from base commit `ad272be20`.

- Verify: `git diff ad272be20 -- .claude/rules/moai/core/askuser-protocol.md .claude/rules/moai/workflow/orchestration-mode-selection.md`
  showing no hunk touching § Report-Before-Ask Gate, § Requested-Deliverable Primacy, or
  § Option Description Standards (the neutral-description rule's home), and no change at all in
  `orchestration-mode-selection.md` (which carries the Implementation Kickoff Approval mandate
  clause).
- Baseline (this run): stdout empty; exit code `0`; `| wc -l` → `0`. This is the state the guard
  protects — M1 will add hunks to `askuser-protocol.md`, and this criterion asserts that none of
  them lands inside the four named surfaces (the three `askuser-protocol.md` sections above plus
  the whole of `orchestration-mode-selection.md`).
- Why this is not release-blocking: green at arrival by construction (nothing has been edited yet),
  so it measures the absence of collateral damage, not the presence of the work.

**AC-JFM-011 — non-blocking** — *Given* the amended doctrine, *When* the pull-mode section is read,
*Then* all three adopted conditions appear as binding text: "Detect → Explain → Ask, but never
decide", "An LLM 'best practice' is not a policy", and "When uncertain, escalate. Never downgrade."

- Verify: three `grep -c` hits, each ≥ 1, in `.claude/rules/moai/core/askuser-protocol.md`.
- RED-now: `grep -c 'Detect → Explain → Ask' .claude/rules/moai/core/askuser-protocol.md` → `0`;
  exit code `1`. Red because none of the three has landed.
- Green path: M1.

## §D.4 Frozen zone and doctrine consistency

**AC-JFM-012 — regression-guard** — *Given* `CONST-V3R5-035`, *When* `zone-registry.md` is diffed
against base, *Then* the entry's `clause:` string is unchanged; and `branch-origin-protocol.md`
carries the original text as the push-mode branch plus an added pull-mode branch.

- Verify: `git diff ad272be20 -- .claude/rules/moai/core/zone-registry.md` shows no change to the
  `CONST-V3R5-035` `clause:` line, and
  `grep -c 'recommendation_mode' .claude/rules/moai/development/branch-origin-protocol.md` ≥ 1.
- Baseline (this run): the diff → stdout empty; exit code `0`; `| wc -l` → `0`. The registry is the
  canary-gated Frozen surface; keeping this at zero-change is the whole handling decision
  (spec.md §B.2, REQ-JFM-015).
- Companion measurement, red today and flipped by M1:
  `grep -c 'recommendation_mode' .claude/rules/moai/development/branch-origin-protocol.md` → `0`,
  exit code `1`. The blocking half of that pair is carried by AC-JFM-013, which sweeps the whole
  tree; this criterion's own job is the Frozen-clause invariant, which is a guard.

**AC-JFM-013 [release-blocking]** — *Given* the whole rules, skills, and output-styles tree, *When*
it is swept for clauses mandating a `(Recommended)` / `(권장)` **first** option, *Then* every
candidate is classified in a recorded ledger, and the two coordinates this SPEC names are classed
`conditioned`.

- Sweep window (this is the definition the 0.1.0 criterion lacked): **the enclosing markdown block
  of the matched line**. The block is drawn mechanically — scan up and down from the matched line
  and stop at the nearest of:
  1. a blank line;
  2. a heading (`^#{1,6} `);
  3. a code-fence delimiter (a line whose first non-space characters are three backticks);
  4. a line starting a list item (`^\s*(?:[-*+]|\d+\.)\s`) — scanning **up**, that line is the
     block's first line and the scan stops there; scanning **down**, that line is excluded;
  5. a table row (`^\s*\|`) other than the matched line itself.

  So a wrapped paragraph is all of its physical lines; a list item is its own line plus its wrapped
  continuation lines but not its siblings; a table row is a single line. Two implementers drawing
  this window from the same file land on the same span.

  **Why the window is no longer the matched physical line.** The 0.2.2 text fixed it there and
  justified it in plan.md §F M1 with the claim that "every candidate this tree actually contains is
  a single-line paragraph or list item (measured)". That claim was **false**, and the run phase
  falsified it: `.claude/skills/moai/workflows/plan/spec-assembly.md:212` is a continuation line of
  a wrapped `[HARD]` paragraph spanning roughly `:208`-`:214`. REQ-JFM-016's unit is a **clause**;
  a physical line is not one. Under the line window the criterion's only available pass route was
  rewriting the paragraph as one long line so the token would land inside the measuring instrument
  — and a PASS obtained that way shows that a token sits on a line, not that the doctrine is
  conditioned. The long line at `spec-assembly.md:212` is left exactly as the run phase wrote it:
  it is the honest artifact of the mismatch and the evidence for this repair. **Do not "fix" it.**

- **Carrier form — where the conditioning may live.** A row is `conditioned` when EITHER (a) a mode
  reference appears anywhere in its window, OR (b) the clause at that coordinate is one this SPEC
  **forbids editing**, and the ledger row **names the coordinate where the conditioning text
  actually lives** — that named carrier being itself verifiable. Form (b) is not a relaxation; it
  is what §B.2 already decided (the Frozen `CONST-V3R5-035` text keeps its wording as the push-mode
  branch and a pull-mode branch is added beside it), stated here so the run phase does not have to
  invent it. Two rows take form (b), and for the same reason:
  `branch-origin-protocol.md:25` (carrier: the adjacent pull branch at `:26`) and
  `zone-registry.md:869` (carrier: `branch-origin-protocol.md:25-26` and `spec-assembly.md:353`).
  A line-local or block-local test would score both red **forever**, because REQ-JFM-015 / §B.2
  forbid the edit that would turn them green — a criterion no correct work can satisfy is the
  *impossible* direction `verification-completeness.md` §2 names, not a strict one.
  Form (b) requires no edit to `zone-registry.md`, so it does **not** conflict with AC-JFM-012's
  empty-diff invariant.
- Verify, in order:
  1. `grep -rn -E '\(Recommended\)|\(권장\)' .claude/rules .claude/skills .claude/output-styles | grep -iE '\bfirst\b|첫 |먼저' > .moai/reports/t401/ac013-candidates.txt`
  2. `wc -l < .moai/reports/t401/ac013-candidates.txt` — the **swept count**, which MUST be > 0
     (a pass over an empty sweep asserts nothing, `verification-completeness.md` §1.1)
  3. `.moai/reports/t401/ac013-ledger.md` carries exactly that many rows, each classed
     `conditioned` or `unconditioned-by-design: <reason>`, with no unclassified remainder; and the
     rows for `.claude/rules/moai/core/askuser-protocol.md:64` (S1's first coordinate, named at
     spec.md §B and plan.md M1), any other `askuser-protocol.md` (S1) row,
     `.claude/skills/moai/workflows/run.md:137`,
     `.claude/skills/moai/workflows/plan/spec-assembly.md:212`,
     `.claude/skills/moai/workflows/plan/spec-assembly.md:353`, and
     `.claude/rules/moai/core/zone-registry.md:869` are all `conditioned`.

     **Each required coordinate is identified by its anchor text, not by its line number.** The
     line numbers below were measured at `82edb9109` and are a locating aid that decays on the
     next insert; the anchor phrase is the assertion. Where the two disagree, grep the anchor and
     use the line the content is actually on.

     | Required coordinate (as of `82edb9109`) | Anchor text (the assertion) |
     |---|---|
     | `askuser-protocol.md:64` | `**First option label**` |
     | `askuser-protocol.md:265` | `Step 2: Compose AskUserQuestion round` |
     | `run.md:137` | `first option marked "(Recommended)"` |
     | `spec-assembly.md:212` | the phrase: or the (권장) first-option label |
     | `spec-assembly.md:353` | the phrase: First option: the recommended Choice with (권장) suffix |
     | `zone-registry.md:869` | the line beginning: clause: "Skill body BODP gate |
     | `branch-origin-protocol.md:25` | `[ZONE:Frozen] [HARD] Skill body BODP gate MUST follow` |

     Each named coordinate carries its reason inline:
     - `askuser-protocol.md:64` — the coordinate a section-scoped companion criterion
       (AC-JFM-004) does not reach: AC-JFM-004 is satisfied inside § Recommendation Placement
       Principles, while `:64` lives in § Socratic Interview Structure.
     - `spec-assembly.md:212` — a `[HARD]` clause mandating the `(권장)` first-option label at
       the **Implementation Kickoff Approval gate**. It is the same clause as `run.md:137`, one
       file over; conditioning `run.md:137` alone leaves the two-`[HARD]`-clause contradiction
       §E.1 exists to close standing verbatim, at the same gate.
     - `spec-assembly.md:353` — the sole implementing site of the Frozen clause `CONST-V3R5-035`
       (`zone-registry.md:864-870`, `canary_gate: true`), whose clause subject is literally
       "Skill body BODP gate". `spec-assembly.md:330-356` IS that gate and `:353` is its
       `(권장)`-first line. This SPEC already conditions that clause's doctrine site
       (`branch-origin-protocol.md:25` — plan.md M1, AC-JFM-012); leaving its only
       implementation unconditioned would make the change half-applied, so that after M1 the
       doctrine withholds the label under `pull` while `:353` mandates it unconditionally.
     - `zone-registry.md:869` — the Frozen `CONST-V3R5-035` `clause:` string. Admitted on the
       **carrier form** above: its conditioning lives at `branch-origin-protocol.md:25-26` (the
       doctrine site) and `spec-assembly.md:353` (the implementing site), both of which this
       criterion already requires to be `conditioned`. It is `conditioned` for exactly the reason
       `branch-origin-protocol.md:25` is — same shape, same carrier chain — and classing the two
       differently was an artifact of the retired line window, not a real distinction. Nothing in
       this classing requires editing `zone-registry.md`, so AC-JFM-012's empty-diff invariant is
       untouched.

     The first four are admitted on the **consequence** test REQ-JFM-016 states — "any such clause
     **reachable from the six surfaces**" — not on membership in §B.1's S1 coordinate table.
     The table enumerates S1's known instances; it is not S1's definition, and the sweep exists
     precisely to find instances the table missed. Using the table as the scope test would be
     circular. The consequence test is also the one this SPEC already applies: `run.md:137` and
     `branch-origin-protocol.md:25` are **both** outside S1-S6 and both admitted anyway (§E.1,
     plan.md M1, AC-JFM-012).
- RED-now, four measurements re-run in this tree at `HEAD ad272be20abff9e4f3b1b363fce3e48dac4c5132`
  after the selector was made case-insensitive (the sweep was written to a scratch path,
  `/tmp/ac013c.txt`, because `.moai/reports/t401/` is the run-phase artifact location):
  1. `grep -rn -E '\(Recommended\)|\(권장\)' .claude/rules .claude/skills .claude/output-styles | grep -iE '\bfirst\b|첫 |먼저' | tee /tmp/ac013c.txt | wc -l`
     → `25`; exit code `0`. Swept count is non-zero, so the sweep is interpretable.
  2. `grep -c 'recommendation_mode' /tmp/ac013c.txt` → `0`; exit code `1` — **no** candidate
     carries a mode reference.
  3. `ls .moai/reports/t401/ac013-ledger.md` →
     `ls: .moai/reports/t401/ac013-ledger.md: No such file or directory`; exit code `1`.
  4. Coverage control on the selector fix — `grep -c 'askuser-protocol.md:64' /tmp/ac013c.txt` →
     `1`; exit code `0`. Under the 0.2.0 case-sensitive selector this returned `0` / exit `1`
     (`**First option label**` carries a capital `F`, so `\bfirst\b` never matched it), and the
     swept count was `23`. The case-insensitive selector returns `25`; the two rows it adds are
     exactly `askuser-protocol.md:64` and
     `.claude/skills/moai/workflows/plan/spec-assembly.md:353`.
  5. Coverage control on the second required-`conditioned` skill coordinate —
     `grep -c 'plan/spec-assembly.md:212' /tmp/ac013c.txt` → `1`; exit code `0`. This row was
     already inside the case-sensitive `23`, so it has been in the swept set since 0.1.0; what
     0.2.2 adds is its presence in the required-`conditioned` list above, not its presence in
     the sweep.
  6. Coverage control on the third — `grep -c 'plan/spec-assembly.md:353' /tmp/ac013c.txt` →
     `1`; exit code `0`.

  Red on both halves: zero conditioned candidates and no ledger. **Six** candidates in that set
  are coordinates this SPEC must condition — `.claude/rules/moai/core/askuser-protocol.md:64`,
  `.claude/rules/moai/development/branch-origin-protocol.md:25`,
  `.claude/skills/moai/workflows/run.md:137`,
  `.claude/skills/moai/workflows/plan/spec-assembly.md:212`,
  `.claude/skills/moai/workflows/plan/spec-assembly.md:353`, and
  `.claude/rules/moai/core/zone-registry.md:869` — and all six are currently
  unconditioned, which is the `[HARD]`-clause contradiction spec.md §E.1 names. Two of the six —
  `branch-origin-protocol.md:25` and `zone-registry.md:869` — are conditioned via a **carrier**
  rather than in their own window (see the carrier form above), so the count is of coordinates the
  SPEC must condition, not of edits it must make.

  **No candidate may be classed `unconditioned-by-design` on the ground that it is absent from
  §B.1's coordinate table.** The scope test is REQ-JFM-016's reachability criterion, not table
  membership, and the run phase has no authority to re-derive the exclusion rule the 0.2.1 text
  supplied. Where the run phase judges that a swept row genuinely warrants
  `unconditioned-by-design`, that judgment is a **blocker report to the orchestrator**, resolved
  at the Implementation Kickoff Approval gate or a re-delegation — never a classification the run
  phase makes on its own authority and passes on. There is now **no standing exception**: the row
  the 0.2.2 text carved out (`zone-registry.md:869`) is `conditioned` under the carrier form, so
  the two-class contract has no third ground and needs none.
- Why a ledger rather than a zero-count: requiring a mode reference **inside every candidate's own
  window** would be wrong. Two coordinates carry the SPEC's own no-edit constraint
  (`zone-registry.md:869`, `branch-origin-protocol.md:25`) and are conditioned by a carrier
  elsewhere; a window-local zero-count criterion would demand an edit REQ-JFM-015 / §B.2
  prohibits, which is a criterion no correct work can satisfy. The ledger records the carrier
  instead.
- **Re-sweep obligation (run-phase task — named here, not performed here).** The window only
  widens: a matched line is always inside its own block, so **every row already classed
  `conditioned` on a window-local token stays `conditioned`** and needs no re-check. The rows
  classed `unconditioned-by-design` were judged against the retired narrower window and MUST be
  re-checked against the block — a row whose block carries a mode reference outside the matched
  line becomes `conditioned`. Concretely, against the existing ledger
  (`.moai/reports/t401/ac013-ledger.md`, 26 rows: 8 `conditioned`, 18 `unconditioned-by-design`,
  0 escalated): the 8 survive as authored; the 18 are re-checked; and row 4
  (`zone-registry.md:869`) moves to `conditioned` under the carrier form. The re-sweep is
  run-phase work under this criterion, not plan-phase work.
- Green path: M1. Swept count ≥ 25 with a ledger row per candidate, and all six named
  coordinates — `askuser-protocol.md:64`, `branch-origin-protocol.md:25`, `run.md:137`,
  `spec-assembly.md:212`, `spec-assembly.md:353`, `zone-registry.md:869` — classed `conditioned`,
  each on a window-local token or on a named carrier. This is the criterion that mechanically
  closes the two-`[HARD]`-clause contradiction (spec.md §E.1).

## §D.5 Runtime observer

**AC-JFM-014 [release-blocking]** — *Given* the observer wired into the `PreToolUse` handler and an
`AskUserQuestion` matcher registered in `.claude/settings.json`, *When* an `AskUserQuestion` call is
issued, *Then* exactly one JSONL row is appended under `.moai/logs/` carrying the timestamp, session
id, resolved mode, `label_present` boolean, option count, and question count.

- Verify: `go test ./internal/hook/... -run 'AskUserQuestionObserver' -count=1 -v` naming the
  row-shape test, plus one live observation read back, plus the matcher list containing
  `AskUserQuestion`.
- RED-now, two independent measurements:
  1. `go test ./internal/hook/... -run 'AskUserQuestionObserver' -count=1` →
     `ok  	github.com/modu-ai/moai-adk/internal/hook	0.562s [no tests to run]` — the **first of 11** lines, the `[no tests to run]` token holding on all 11 (re-measured, `wc -l` → `11`, `grep -c '[no tests to run]'` → `11`, `0` `FAIL` lines); exit code `0`.
     Red because the swept set is empty — the observer and its tests do not exist.
  2. `python3 -c "import json;print([m.get('matcher') for m in json.load(open('.claude/settings.json'))['hooks']['PreToolUse']])"`
     → `['Write|Edit|Bash', 'Agent|Task', 'SendMessage|TaskStop']`; exit code `0`. Red because
     there is no `AskUserQuestion` matcher, so no call would ever reach a handler.
- Green path: M0. Named tests appear under `-v`; the matcher list carries a fourth entry
  `AskUserQuestion`; one live row is read back from `.moai/logs/`.

**AC-JFM-015 [release-blocking]** — *Given* any `AskUserQuestion` payload, *When* the handler runs,
*Then* the returned decision is never a deny, and no established prior decision is displaced.

- Verify: `go test ./internal/hook/... -run 'AskUserQuestionObserver.*NeverDenies' -count=1 -v`
  asserting the handler's output over a table of payloads including malformed ones.
- RED-now: `go test ./internal/hook/... -run 'AskUserQuestionObserver' -count=1` →
  `ok  	github.com/modu-ai/moai-adk/internal/hook	0.562s [no tests to run]` — the **first of 11** lines, the `[no tests to run]` token holding on all 11 (re-measured, `wc -l` → `11`, `grep -c '[no tests to run]'` → `11`, `0` `FAIL` lines); exit code `0`.
  Red for the empty-swept-set reason above.
- Green path: M0; the `-v` output names the `NeverDenies` table subtests and the `[no tests to run]`
  token is absent (CONST-7 — the observer measures, it never gates).

**AC-JFM-016 [release-blocking]** — *Given* an unparseable payload, an absent config file, and an
unwritable log path, *When* each is exercised, *Then* the call is allowed in every case and no error
propagates to the caller.

- Verify: `go test ./internal/hook/... -run 'AskUserQuestionObserver.*FailOpen' -count=1 -v`
  naming all three subtests.
- RED-now: as for AC-JFM-015, the command actually run was the superset selector,
  `go test ./internal/hook/... -run 'AskUserQuestionObserver' -count=1` →
  `ok  	github.com/modu-ai/moai-adk/internal/hook	0.562s [no tests to run]` — the **first of 11** lines, the `[no tests to run]` token holding on all 11 (re-measured, `wc -l` → `11`, `grep -c '[no tests to run]'` → `11`, `0` `FAIL` lines); exit code `0`. The
  narrower `…FailOpen` selector therefore also sweeps zero — a deduction from the superset
  measurement, not a separate observation.
- Green path: M0; three named `FailOpen` subtests pass.

## §D.6 Static guard and the vacuity falsifier

**AC-JFM-017 [release-blocking]** — *Given* the CI guard workflow, *When* the doctrine text and the
§8 banner templates are made deliberately inconsistent in a scratch commit, *Then* the guard job
fails; and *When* they are consistent, *Then* it passes.

- Verify: both directions exercised — a mutation run showing RED, and the clean run showing GREEN.
  A guard demonstrated only on the passing side is a vacuous guard.
- RED-now: `ls .github/workflows/judgment-first-consistency.yaml` →
  `ls: .github/workflows/judgment-first-consistency.yaml: No such file or directory`; exit code
  `1`. Red because the guard does not exist, so neither direction can be demonstrated.
- Green path: M4. Both directions recorded in `progress.md` §E.2 with the run URL or the local
  job output for each; the RED direction's scratch commit is reverted before merge.

**AC-JFM-018 [release-blocking] — vacuity falsifier** — *Given* this repository dogfooding
`recommendation_mode: pull`, and a window in which at least 20 `AskUserQuestion` calls have been
recorded by the observer, *When* the JSONL log is read, *Then* zero rows carry
`label_present: true`.

- **The denominator, stated once and identically in all three artifacts**: every recorded row with
  `mode == "pull"`. **No `question_type` filter, in any branch.** `design.md` §6.3 owns the rule;
  `plan.md` M6's jq and this criterion are copies of it. The 0.1.0 form filtered on
  `question_type == "decision"`, a field the `AskUserQuestion` payload does not carry — that jq
  returns `n = 0` forever, making the criterion a permanent gap by construction. The requirement
  widened to match what is observable rather than the measurement narrowing to match an
  unobservable requirement (REQ-JFM-005, spec.md §B).
- **The exported artifact carries its provenance, or it asserts nothing.** The window is collected
  by the session that actually asks and whose observer is actually wired — under the kanban division
  of labour the lead session, not the card's lane, and a session started after the `AskUserQuestion`
  matcher landed, since one already running when the matcher was added does not pick it up — and its
  rows land under that session's `CLAUDE_PROJECT_DIR` (cwd when unset), at
  `.moai/logs/askuser-observations.jsonl`, generally **not** this worktree. The exported
  `.moai/reports/t401/pull-window.jsonl` is therefore a copy whose origin the file itself does not
  record. It MUST be accompanied by `pull-window.jsonl.provenance.md` stating: the **source
  absolute path**; the **asking session's `session_id`**; the **collection interval** (the `timestamp`
  of the first and last exported row); the **row count** and the **`label_present: true` count** as
  measured at export time; the asking session's own count of `AskUserQuestion` calls issued during
  the interval (**`calls_issued`** — a value the asking session knows without the observer); and the
  **export command**. A `session_start` timestamp and a matcher SHA are deliberately **not** part of
  the record: the exported window's own existence and row counts already prove the wired-session
  condition, so those fields would be confirmation stamps gating nothing, and fields that gate
  nothing leave the impression verification finished when it did not
  (`.moai/reports/t401/provenance-eligibility-options.md`, Option B rejection — the doc adopts
  Option A + Option C). An artifact without
  that record is an unattributed claim under `verification-claim-integrity.md` §2 — a Gap, never a
  Claim.
- Verify, all three halves required:
  1. `jq -s '[.[] | select(.mode=="pull")] | {n: length, violations: ([.[] | select(.label_present==true)] | length)}' .moai/reports/t401/pull-window.jsonl`
     returning `n >= 20` and `violations == 0`.
  2. `ls .moai/reports/t401/pull-window.jsonl.provenance.md` → exit code `0`, **and** the row count
     recorded in that file equal to the artifact's actual row count
     (`wc -l < .moai/reports/t401/pull-window.jsonl`). A disagreement is a gap: it means the record
     describes a different export than the one being read.
  3. The provenance's `calls_issued` contrasted with the artifact's `rows_recorded` (its actual row
     count) under a four-way reading rule: `rows_recorded == calls_issued` → the window covers the
     interval and the sample stands; `rows_recorded == 0` with `calls_issued > 0` → the observer was
     not wired into that session (exactly the negative observation the lead session recorded on
     2026-09-02: 4 calls issued, 0 rows anywhere); `0 < rows_recorded < calls_issued` → partial row
     loss — the window is a **Gap** and must not be read as a sample; `rows_recorded >
     calls_issued` → rows from other sessions are mixed in — split by `session_id` before any
     reading. The three mismatch states are observable signals, never silent passes: without this
     contrast, observer non-wiring and partial row loss read silently as "no violations".
- RED-now, all three halves, re-measured in this tree at `HEAD 095f2799b` (the provenance-amendment
  commit's parent; the amendment changes only SPEC artifacts, so the probe subjects are identical
  in both trees):
  1. `ls .moai/logs/ | grep -c askuser` → `0`; exit code `1`.
  2. `ls .moai/reports/t401/pull-window.jsonl.provenance.md` →
     `ls: .moai/reports/t401/pull-window.jsonl.provenance.md: No such file or directory`; exit code
     `1`.
  3. Half 3 is red for the same reason as half 2: the provenance record that would carry
     `calls_issued` does not exist, so the contrast it asserts has no input and cannot be read —
     there is no sample, no record, and no issued-call count to contrast.
  Red on all three halves because no observer log exists and nothing has been exported, so there is no
  sample and no record of one; the criterion is unmet. (An absent log is a gap, and a gap is red; it
  is never read as `violations == 0`.)
- Green path: M6, gated on AC-JFM-023 being green first.
- **Why this is the falsifier**: it FAILS when the convention text is present but the orchestrator
  does not follow it. No amount of correct doctrine text can make it pass.
- **Mutant probe, stated explicitly so the run phase cannot claim this criterion from an
  undetecting observer**: an observer whose label detector never fires — a broken regex, an
  unreached payload path, a hardcoded `false` — produces `violations == 0` on every row while the
  convention is entirely unfollowed. That mutant is excluded by AC-JFM-023, which must be green
  **before** this window opens. AC-JFM-018 read without AC-JFM-023 asserts nothing.
- Sample-size note: `n >= 20` is the floor for a claim of adherence, not a statistical power
  target. A run in which `n < 20` is a **gap**, not a pass.

**AC-JFM-023 [release-blocking] — detector positive control** — *Given* the observer's label
detector, *When* it is exercised on a payload that does carry a `(권장)` / `(Recommended)` option
label, *Then* it records `label_present: true`; and on one that does not, `false`.

- Verify, all three surfaces required:
  1. `go test ./internal/hook/... -run 'AskUserQuestionObserver.*LabelDetect' -count=1 -v` naming
     both directions as subtests.
  2. `jq -s '{n: length, positives: ([.[] | select(.label_present==true)] | length)}' .moai/reports/t401/baseline-push-window.jsonl`
     returning `n >= 5` and `positives >= 1` — a **live** pre-landing window recorded while this
     repository is still at `recommendation_mode: push` and the doctrine is unamended.
  3. `ls .moai/reports/t401/baseline-push-window.jsonl.provenance.md` → exit code `0`, **and** the
     row count recorded in that file equal to the artifact's actual row count
     (`wc -l < .moai/reports/t401/baseline-push-window.jsonl`).
  4. The provenance's `calls_issued` contrasted with the artifact's `rows_recorded` (its actual row
     count) under the same four-way reading rule as AC-JFM-018 half 3: `rows_recorded ==
     calls_issued` → the window covers the interval and the control sample stands;
     `rows_recorded == 0` with `calls_issued > 0` → the observer was not wired into that session
     (the 2026-09-02 negative observation: 4 calls issued, 0 rows anywhere); `0 < rows_recorded <
     calls_issued` → partial row loss — the window is a **Gap** and must not be read as a control
     sample; `rows_recorded > calls_issued` → rows from other sessions are mixed in — split by
     `session_id` before any reading. The mismatch states are observable signals, never silent
     passes.
- **The exported artifact carries its provenance, or it asserts nothing.** This window too is
  collected by the session that actually asks and whose observer is actually wired — the lead session
  under kanban division of labour, never the card's lane, which issues no `AskUserQuestion` calls at
  all, and a session started after the `AskUserQuestion` matcher landed, since one already running
  when the matcher was added does not pick it up; for this baseline that session must further
  predate the convention landing, the window being the control recorded before the convention exists
  to suppress the label — and its rows accumulate under that session's `CLAUDE_PROJECT_DIR`
  (cwd when unset), at
  `.moai/logs/askuser-observations.jsonl`, generally **not** this worktree. The exported artifact
  MUST therefore be accompanied by `baseline-push-window.jsonl.provenance.md` recording: the
  **source absolute path**; the **asking session's `session_id`**; the **collection interval** (the
  `timestamp` of the first and last exported row); the **row count** and the **`label_present: true`
  count** as measured at export time; the asking session's own count of `AskUserQuestion` calls
  issued during the interval (**`calls_issued`** — a value the asking session knows without the
  observer); and the **export command**. A `session_start` timestamp and a matcher SHA are
  deliberately **not** part of the record (Option B rejection — the doc adopts Option A + Option
  C, `.moai/reports/t401/provenance-eligibility-options.md`) — confirmation stamps gating nothing. A
  copied JSONL with no record of
  which session collected it, from which tree, over which interval, is an unattributed claim under
  `verification-claim-integrity.md` §2 — and a positive control read from an unattributed sample
  asserts nothing, exactly as this criterion says of a detector never observed firing.
- RED-now, all four halves (halves 2-4 re-measured in this tree at `HEAD 095f2799b` — the
  provenance-amendment commit's parent, whose probe subjects are identical to the amendment
  commit's; half 1 stands as measured at `ad272be20`, unchanged by this amendment):
  1. `go test ./internal/hook/... -run 'AskUserQuestionObserver' -count=1` →
     `ok  	github.com/modu-ai/moai-adk/internal/hook	0.562s [no tests to run]` — the **first of 11** lines, the `[no tests to run]` token holding on all 11 (re-measured, `wc -l` → `11`, `grep -c '[no tests to run]'` → `11`, `0` `FAIL` lines); exit code `0` —
     empty swept set.
  2. `ls .moai/reports/t401/baseline-push-window.jsonl` →
     `ls: .moai/reports/t401/baseline-push-window.jsonl: No such file or directory`; exit code `1`.
  3. `ls .moai/reports/t401/baseline-push-window.jsonl.provenance.md` →
     `ls: .moai/reports/t401/baseline-push-window.jsonl.provenance.md: No such file or directory`;
     exit code `1`. Requiring provenance grows this cell from one absent file to **two**: the
     artifact and its record are both missing, and either one missing keeps the criterion red.
  4. Red for the same reason as half 3: the provenance record that would carry `calls_issued` does
     not exist, so the contrast it asserts has no input — no control sample, no record, and no
     issued-call count to contrast.
- Green path: M0, before M1 amends any doctrine. The ordering is load-bearing and is why M0 leads
  the plan: once the convention lands there is no window left in which a live `label_present: true`
  row can be produced (plan.md §F preamble).

## §D.7 Distribution

**AC-JFM-019 [release-blocking]** — *Given* every file changed under `.claude/`, `.moai/config/`, or
`.github/`, *When* the change set is enumerated, *Then* the change set is **non-empty**, each file
with a template counterpart has a matching change under `internal/template/templates/<same path>`,
and `make build` regenerates the embedded filesystem with no uncommitted drift afterward.

- Verify: `git diff --name-only ad272be20 | grep -cE '^\.claude/|^\.moai/config/'` > 0 (the swept
  count), then the same command without `-c` compared pairwise against the template tree, then
  `make build && git status --porcelain` clean.
- RED-now: `git diff --name-only ad272be20 | grep -cE '^\.claude/|^\.moai/config/'` → `0`; exit
  code `1`. Red because the change set is empty: a mirror-parity claim over zero files is
  uninterpreted output, not a pass (`verification-completeness.md` §1.1). The non-empty clause is
  in the criterion for exactly this reason.
- Green path: M5, with a swept count of at least 6 (`askuser-protocol.md`,
  `branch-origin-protocol.md`, `run.md`, `moai.md`, `context-window-management.md`,
  `settings.json`, plus `interview.yaml`), each paired — and, since the 0.2.2 scope widening,
  `plan/spec-assembly.md` as well (8 files in the enumeration; the floor of 6 stands, the list is
  illustrative of the M0-M3 edit surface).
- Note on `.github/`: `judgment-first-consistency.yaml` has **no template counterpart**, so
  REQ-JFM-021's counterpart clause does not fire on it. This is not because `.github/` is outside
  the template — it is not: `find internal/template/templates/.github -type f` → 4 files, including
  `workflows/label-sync.yml` (research.md §6).

**AC-JFM-020 — regression-guard** — *Given* the template mirror changes, *When* the neutrality
audit runs, *Then* it passes with zero findings introduced by this SPEC.

- Verify: `go test ./internal/template/... -run TestTemplateNeutralityAudit -count=1 -v` and
  `grep -rn 'SPEC-JUDGMENT-FIRST-MODE\|REQ-JFM\|/Users/\|CLAUDE.local' internal/template/templates/`
  returning zero hits.
- Baseline (this run): the grep → stdout empty; exit code `1`; `| wc -l` → `0`. Zero at arrival,
  and the guard's job is that it is still zero after M1-M5 land text in the template tree.
- Why this is not release-blocking: green at arrival. Neutrality is authored from the start
  (REQ-JFM-022, plan.md §D) rather than repaired at the end, so this criterion can only ever detect a
  regression.

**AC-JFM-021 — regression-guard** — *Given* the hook wrappers that exist as `.sh` / `.sh.tmpl`
pairs, *When* the pairs are compared, *Then* none has drifted, and the swept count is reported
and MUST equal the baseline `4` (an empty sweep prints `swept=0 drift=0` and would otherwise read
as a pass — the same explicit-floor shape AC-JFM-013 and AC-JFM-019 carry).

- Verify (the corrected loop — the 0.1.0 recipe was defective):

  ```bash
  n=0; d=0
  for f in internal/template/templates/.claude/hooks/moai/*.tmpl; do
    b=${f%.tmpl}
    [ -f "$b" ] || continue
    n=$((n+1))
    diff -q "$b" "$f" >/dev/null || { d=$((d+1)); echo "DRIFT $(basename "$b")"; }
  done
  echo "swept=$n drift=$d"
  ```

- Baseline (this run): `swept=4 drift=0`; exit code `0`. The four pairs are
  `handle-agent-hook.sh`, `handle-stop-goal.sh`, `handle-task-completed.sh`,
  `handle-teammate-idle.sh`.
- **Why the recipe changed.** The 0.1.0 form was
  `[ -f "$b" ] && diff -q "$b" "$f" >/dev/null || echo "DRIFT …"`, which reports a `.tmpl` with no
  base sibling as drift, because the `[ -f ]` guard falls into the `||` branch. Measured on this
  untouched tree it printed **31** DRIFT lines (`| wc -l` → `31`) out of 35 `.tmpl` files, only 12
  of which have any `.sh` sibling and only 4 of which pair. The criterion was red on arrival and
  red forever — the *impossible* direction `verification-completeness.md` §2 names.
- **Stated gap, not a hidden vacuous pass.** `handle-pre-tool.sh` — the wrapper the 0.1.0 criterion
  named — is **not in the swept set**, because the template tree carries only
  `handle-pre-tool.sh.tmpl` and no base sibling (`ls
  internal/template/templates/.claude/hooks/moai/handle-pre-tool.sh` → `No such file or directory`;
  exit code `1`). Separately, M0 changes **no** wrapper at all: its only rendered-file edit is
  `.claude/settings.json`, whose `.tmpl` counterpart is covered by AC-JFM-019. REQ-JFM-023's `When`
  antecedent — "when a hook wrapper is changed" — is therefore not triggered by this SPEC, which is
  precisely why this criterion is a regression guard and not a release blocker. Should a milestone
  come to touch a wrapper, this criterion is promoted to release-blocking and a RED-now cell is
  authored for that pair before the edit lands.

**AC-JFM-022 — regression-guard** — *Given* the complete change set, *When* the affected Go
packages are tested, *Then* `internal/config`, `internal/hook`, and `internal/template` are green.

- Verify: `go test ./internal/config/... ./internal/hook/... ./internal/template/... -count=1` with
  the exit code recorded. Full-suite judgment is CI's, per the project's lane-local verification
  rule.
- Baseline (this run): exit code `0`; 16 packages `ok`, 1 `[no test files]`
  (`internal/template/scripts`), zero `FAIL` lines. Longest: `internal/hook` 49.020s,
  `internal/hook/perf` 46.126s, `internal/template` 38.776s.
- Why this is not release-blocking: green at arrival, so as a blocker it would assert nothing about
  this work. The criteria that assert the new code exists are AC-JFM-002/003 (config) and
  AC-JFM-014/015/016/023 (hook), each red now by empty swept set.

## Definition of Done

- **All 13 release-blocking criteria** — AC-JFM-002, 004, 005, 008, 009, 013, 014, 015, 016, 017,
  018, 019, 023 — verified with the command run and the output observed, in the tree that merges,
  at the commit that merges. The count is derived from the §D matrix's `Severity` column, never
  restated independently — and the derivation is scoped to the matrix so this sentence does not
  count itself:
  `grep -c '^| AC-JFM-.*| release-blocking |' acceptance.md` → `13` (matrix rows only; this
  sentence does not start with `| AC-JFM-`, so it cannot count itself).
- **All 6 regression-guard criteria** — AC-JFM-001, 010, 012, 020, 021, 022 — re-measured after the
  final commit and still matching their baseline cells above.
- AC-JFM-023 green **before** the AC-JFM-018 window opens. Reading the falsifier from a detector
  never observed firing is a vacuous pass and is the one failure this ordering exists to prevent.
- AC-JFM-018's dogfood sample actually collected — not projected, not assumed. `n < 20` is recorded
  as a gap.
- AC-JFM-017 demonstrated in both directions (mutation RED, clean GREEN).
- Every non-blocking criterion (AC-JFM-003, 006, 007, 011) either verified or explicitly recorded as
  an open gap with its reason.
- Template mirror parity and `make build` cleanliness confirmed after the final commit, not before.
