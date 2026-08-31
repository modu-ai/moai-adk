---
id: SPEC-RED-NOW-THRESHOLD-001
title: "RED-now cell adoption gate — a release-blocking criterion must carry an observed, re-executable RED"
version: "0.4.0"
status: completed
created: 2026-08-29
updated: 2026-08-29
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/spec
lifecycle: spec-anchored
tags: verification-completeness, plan-auditor, acceptance-criteria, must-pass, red-now
tier: M
---

# SPEC-RED-NOW-THRESHOLD-001 — RED-now cell adoption gate

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-08-29 | 0.1.0 | Initial draft (card t343). Authored on `WT-red-now-threshold@a6bbbf82b`. |
| 2026-08-29 | 0.4.0 | Coordinator corrections folded in: §A.2 evidence replaced after the original refuting command was found invalid (the named test exists), with the repair recorded in §A.2.1 rather than swapped silently; §A.2 reframed from one cell to nine; REQ-RNT-015 added constraining L2's verdict to executed-test count (`evidence_writer.go` seam). |
| 2026-08-29 | 0.3.0 | Iteration-2 defect closure (FAIL on N1, score 0.800). Ledger admitted as a carrier in the requirement layer and REQ-RNT-008's scope predicate made carrier-independent (N1); REQ-RNT-002 aligned to four elements (N2); expensive-command disposition added to REQ-RNT-013 (N5). |
| 2026-08-29 | 0.2.0 | Iteration-1 audit revision (FAIL 0.75). L1 re-scoped class-independent (D4); "the command" defined as read-only single-invocation (D6, open decision (a)); document-level tree pin rule adopted (open decision (b)); MP-8 execution discipline (REQ-RNT-013) and continued-firing (REQ-RNT-014) added (D6, D7); line numbers removed from REQ-RNT-011 (D8); `module` corrected (D9). |

---

## A. WHY

`.claude/rules/moai/development/verification-completeness.md` §2 already requires that an
acceptance criterion be adopted as a pair of cells, the RED-now cell being *"the criterion
observed red on the pre-implementation tree, pinned to the tree it was measured on"*.
Nothing checks it.

### A.1 The audit has zero coverage — measured, not assumed

The absence is not a threshold set too low. It is the absence of a criterion.

```
$ grep -c "RED\|two-cell\|verification-completeness\|mutant" .claude/agents/moai/plan-auditor.md
0
```

Measured on `WT-red-now-threshold@a6bbbf82b`, against a 482-line file. The must-pass firewall
runs MP-1..MP-7; Group 4 (Acceptance Criteria Quality) runs AC-1..AC-5, which test
Given-When-Then form, binary-testability, weasel words, and REQ traceability in both directions.
No item looks at a RED cell.

**This SPEC therefore introduces a NEW must-pass criterion (MP-8). It is not a numeric threshold
adjustment.** The originating card was titled as a threshold change ("임계를 조정"); the
measurement above corrects that framing, and this SPEC is built on the corrected one. There is no
dial to turn, because there is no instrument.

### A.2 An unobserved RED cell is not merely unverified — it can be false, and here it is false nine times

The live instance is `.moai/specs/SPEC-TODO-SQLITE-001/acceptance.md` (card t306, landed). Its
header pins a tree (`WT-todo-sqlite@d29b8942e`), so the *evidence-pinning* obligation of rule §4
was met. No RED cell in it carries a command output.

**The defect is systemic within that SPEC, not a single slip.** Census:

```
$ awk -F'|' '/missing test|does not exist|no tests to run/ {gsub(/ /,"",$2); print $2}' .moai/specs/SPEC-TODO-SQLITE-001/acceptance.md
AC-TOSQ-001
AC-TOSQ-002
AC-TOSQ-003
AC-TOSQ-004
AC-TOSQ-005
AC-TOSQ-007
AC-TOSQ-008
AC-TOSQ-011
AC-TOSQ-017
AC-TOSQ-018
```

Ten criteria surface; **nine** of them (001, 002, 003, 004, 005, 007, 008, 017, 018) rest on one
false premise — that an absent test turns the suite red. It is false on both routes available to
them: a package suite that simply lacks the test is green, and a `-run` selector that matches
nothing is green. **AC-TOSQ-011 is deliberately excluded from the count**: its RED rests on an
absent CLI verb (`moai todo export-json` returning non-zero), a different mechanism. It was not
measured here, so it is neither counted as defective nor asserted sound.

Nine release-blocking criteria in one SPEC, carrying one false premise, is the argument for MP-8 —
not a single author's slip.

**The premise, measured:**

```
$ go test ./internal/kanban -run TestMigrationParityDoesNotExistXYZ -count=1 ; echo "exit=$?"
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.219s [no tests to run]
exit=0
```

A `-run` selector matching zero tests exits 0 and prints `ok`. The only token separating this from
a genuine pass is `[no tests to run]` — which matters for REQ-RNT-015 below.

#### A.2.1 The first version of this evidence was itself an unobserved claim — recorded, not quietly repaired

Iterations 1 and 2 of this SPEC cited a different command for the premise above:
`go test ./internal/kanban -run TestMigrationParity`, reported as `ok` / `exit=0`. **That command
proved nothing**, because the test it named exists:

```
$ grep -rn "func TestMigrationParity" internal/kanban/
internal/kanban/backlog_migrate_test.go:70:func TestMigrationParity(t *testing.T) {
internal/kanban/backlog_migrate_test.go:424:func TestMigrationParityCatchesTamperedRecord(t *testing.T) {
```

What ran was a real, passing test. The premise of the refuting command was never verified before
the refutation was asserted — the exact unobserved-claim defect this SPEC exists to prohibit,
committed while documenting it, and it survived two audits that both re-executed the command and
observed it reproduce. Re-execution confirms a command's *output*; it does not confirm the command
*measures what its author claims*.

**The conclusion is unchanged and only the evidence is replaced.** AC-TOSQ-001's cell is still
false; the corrected command above establishes it, and the corrected command came from lane-5
(card t341). This subsection exists rather than a silent swap: a SPEC about unobserved claims that
quietly repairs its own is worse than one that records the repair.

#### A.2.2 The failure chain — why the corrective must be mechanical

The chain is recorded because it settles a design question this SPEC would otherwise have to
argue. Full transcript: `.moai/reports/t343/red-now-premeasurement.md` §M3.

- **The author of a card built to catch unobserved claims in RED cells produced one as its
  central refutation.** The rule was known, was being actively applied to someone else's SPEC in
  the same document, and was violated anyway — the command's output was observed, its premise
  never was.
- **A second actor confirmed the cited fact and passed the inference through.** The lead checked
  the cell's text against the `develop` copy and reported independent confirmation. The text was
  confirmed correctly; what went unchecked was whether the refuting command measured what it was
  said to measure.
- **Two audits re-executed the command and watched it reproduce.** Re-execution confirms output,
  not that a command measures its claim.
- **A third party, working a different card, found it** by asking whether the named test existed.

Four actors, one chain, and the defect survived every human-judgment layer in it. That is the
argument for the shape of the corrective: **knowing the rule did not prevent the violation**, so a
doctrine clause instructing authors to verify their premises would have been satisfied — sincerely
— by every actor above, and would have caught nothing. The three-layer mechanism of §B is not
preferred over an authoring guideline on grounds of taste; the guideline is the alternative that
was measured failing here.

The chain also bounds what MP-8 can claim. L2 re-executes a cited command and checks the RED
reproduces; it does **not** verify that the command measures its stated premise. That residual is
carried in `acceptance.md` §D.2 as mutant M-6, not papered over — MP-8 would not, by itself, have
caught this incident either. What MP-8 does catch is the class in §A.2's census: nine cells whose
cited RED never reproduces at all.

The `0.94` figure is **quoted from the card dispatch, not re-derived on this tree** — no audit was
re-run here. It is cited to locate the incident, and no clause of this SPEC rests on its exact
value; the clause rests on the cell being outside the scoring surface at all (§A.1), which was
measured.

Full pre-measurement evidence, including this correction and its own Gaps section, is persisted at
`.moai/reports/t343/red-now-premeasurement.md`.

### A.3 A classification precedent already exists, in one SPEC only

`SPEC-TODO-LANDING-STATE-001` (card t331, **landed**; `status: completed`, resolvable at
`.moai/specs/SPEC-TODO-LANDING-STATE-001/` on `15453140a` — it was unlanded when this SPEC was
first authored and arrived with the `develop` merge) splits its criteria into release-blocking and
regression-guard, and demotes AC-TLS-002 and AC-TLS-007 rather than inventing failures for them.
Its §C states:

> No criterion is classified release-blocking on a RED cell that argues counterfactually or in
> the future tense; where a cell has no observation it says so and the criterion is a
> regression-guard instead.

This SPEC's job is to lift that local convention into the audit layer. The quoted sentence was
**re-read from the landed copy** rather than carried over from the pre-landing worktree text, and
is byte-identical to what was quoted before (`acceptance.md:93-95`); nothing drifted between the
worktree read and the landing.

Now that t331 has landed, the precedent **can** be cited by path, and is:
`.moai/specs/SPEC-TODO-LANDING-STATE-001/acceptance.md` §C, with AC-TLS-002 and AC-TLS-007 as the
two worked demotions. That changes what a reader can verify, not what this SPEC depends on — t331
remains cited for its *disposition*, and no requirement, criterion, or milestone here is
conditioned on that file existing. The citation-by-path is a convenience for the reader, not a
dependency.

---

## B. WHAT

Three layers, each judging only what it can actually decide.

| Layer | Question | Owner | Decidable because |
|-------|----------|-------|-------------------|
| L1 — form | Does a cell carry a command, its verbatim stdout, its exit code, and a tree SHA? | repository-local Go test in `internal/spec/` | structure is mechanically extractable |
| L2 — truth | Does the cited command actually reproduce RED on this tree? | plan-auditor, re-executing it | at plan-phase the pinned pre-implementation tree IS the current tree |
| L3 — threshold | Does a violation block, regardless of score? | plan-auditor MP-8 must-pass | score-independence is a firewall property, not a rubric dimension |

**L1 binds every cell that carries a command, whatever its class.** L2 and L3 bind
release-blocking criteria only. This asymmetry is deliberate and was learned from this SPEC's own
first audit: a regression-guard is rightly exempt from the RED *obligation*, but nothing about
demotion earns an exemption from the command being *syntactically runnable*. Iteration 1 found
both of this SPEC's defective commands sitting in regression-guard cells, where a
release-blocking-only check misses them by construction — the demotion path had become an escape
hatch from command verification. Scoping L1 by class would rebuild that hatch.

**L2 is possible at plan-phase and nowhere else.** The RED cell pins the pre-implementation tree;
at plan-phase that tree is the working tree, so the auditor can re-run the command and watch it
fail. §A.2 is a worked instance of this check.

**L2's trust boundary is the opposite of the auditor's existing one, and that is the real cost.**
plan-auditor already carries `Bash` in its `tools:` frontmatter and already executes shell in its
Group A-D verification batches, so **no tooling capability is added** — its `NOT for: … running
tests` description line is a routing clause, not a capability limit. What changes is *exposure*:
every command the auditor runs today is one it composed itself, whereas an MP-8 command is an
arbitrary string a SPEC author typed into a table cell. Same tool, inverted trust boundary. That
is why REQ-RNT-001 constrains the command's *form* (read-only, single invocation) rather than
merely its content, and why REQ-RNT-013 defines what the auditor does when a cited command fails,
refuses to run, or collides with this repository's execution discipline.

**L3 must not be a score dimension.** An aggregate of 0.94 already absorbed this defect once
(§A.2). MP-8 follows the MP-5/MP-6/MP-7 pattern: folded into `## Defects Found` at
`severity=critical`, forcing `Verdict: FAIL` regardless of aggregate, with the MP-4 `N/A`
auto-pass precedent available and its reason stated.

**The two layers are adopted as a pair, for the same reason §2 pairs its two cells.** L1 alone
optimizes an author toward fabricating a quoted output — a fenced block is cheap to type. Only
L2's re-execution defeats fabrication. L1 exists so that the command is *extractable*, which is
what makes L2 runnable at all. Neither is sufficient alone.

### B.1 Requirements

**REQ-RNT-001** — Where an acceptance criterion is classified release-blocking, `verification-completeness.md` §2 shall require its RED-now cell to carry four elements together: the command that was run, that command's verbatim stdout, that command's exit code, and the tree SHA the measurement was taken on. §2 shall define each of the four:

- **the command** — a read-only shell invocation that completes in a single invocation. Pipes, redirection, `&&`, `;` chaining, and subshells are excluded from the form. A citation that is not of this form is not an error: it takes the REQ-RNT-003 demotion path, and the auditor is never obliged to execute it.
- **verbatim** — the raw file bytes, not the GFM-rendered view. A command is verbatim when it runs unmodified as read from the source file.
- **the exit code** — recorded as its own field. It is a separate element precisely because the single-invocation form forbids `; echo $?`, so the exit code cannot be carried by the command itself. An empty stdout with a non-zero exit is a complete and common observation.
- **the tree SHA** — a commit SHA, never a branch name (rule §4). A document-level pin is permitted and **explicitly binds every criterion that carries no pin of its own**; a criterion-level pin wins wherever present.

§2 shall further define the **carrier** — where in `acceptance.md` the four elements may physically live — and shall admit exactly two: a table cell, or a **fenced evidence-ledger entry** that a table cell cites by id. The ledger is the RECOMMENDED carrier, because a GFM table cell mangles shell metacharacters and that mangling has already produced both a vacuous green and an unrunnable command in this SPEC's own first revision. Admitting the ledger in the requirement layer is what keeps it a carrier rather than an exemption: REQ-RNT-008's check follows the elements to whichever carrier holds them.

**REQ-RNT-002** — The RED-now obligation shall be expressed as a structural test over the criterion's content (presence of a command, its stdout, its exit code, and a pinned SHA — the four elements of REQ-RNT-001), and the rule text and the audit criterion shall not key on the tense, the grammatical mood, or any word list of the prose.

**REQ-RNT-003** — Where a cited RED cannot be re-executed on the current tree — a historical event, an already-merged state, or an externally observed CI result — the criterion shall lose release-blocking eligibility and shall be classified as a regression-guard, and shall not be recorded as a pass.

**REQ-RNT-004** — `plan-auditor.md` shall carry a must-pass criterion MP-8 that, for every release-blocking acceptance criterion, re-executes the RED cell's cited command against the current tree and confirms the RED reproduces.

**REQ-RNT-005** — When MP-8 emits an unresolved violation, the plan-auditor shall fold that violation into `## Defects Found` at `severity=critical` and shall set `Verdict: FAIL` regardless of the aggregate score.

**REQ-RNT-006** — Where no acceptance criterion in the SPEC is classified release-blocking, or where `acceptance.md` is absent, MP-8 shall be marked `N/A` and the auditor shall state the reason, following the MP-4 auto-pass precedent.

**REQ-RNT-007** — `plan-auditor.md` shall carry a Group 4 checklist item and an `## Must-Pass Results` report row for MP-8, consistent with the MP-5 / MP-6 / MP-7 rows already present.

**REQ-RNT-008** — A repository-local Go test in `internal/spec/` shall extract the MP-8 form contract from `plan-auditor.md` between a sentinel comment pair rather than restating it, and shall assert that exactly one non-empty sentinel span exists before performing any comparison. Wherever a command is carried in an audited `acceptance.md` — a table cell, a fenced evidence-ledger entry, or any other fenced block — the same test shall check that command against the REQ-RNT-001 form, regardless of the criterion's class and regardless of the carrier. Neither demotion nor relocation shall remove a command from scope.

**REQ-RNT-009** — The repository-local Go test shall observe both directions on named fixtures: a planted RED cell that violates the form contract shall be reported as violating, and a fixture carrying a legitimately observed RED shall not be reported as violating.

**REQ-RNT-010** — When the local copies of `verification-completeness.md` and `plan-auditor.md` change, their `internal/template/templates/` mirrors shall be updated in the same change and `make build` shall be run.

**REQ-RNT-011** — The MP-8 clause added to the template mirror of `plan-auditor.md` shall not carry a real SPEC ID, a card id, or an internal date. The scope is the added clause, not the whole file: the mirror already carries illustrative `SPEC-AUTH-001` placeholders, which are neutral by construction and are not in scope. Their measured locations are cited, with a tree pin, in `acceptance.md` AC-RNT-011 rather than here — line numbers do not belong in a requirement about a file this SPEC's own M4 edits.

**REQ-RNT-012** — The repository-local Go test and its fixtures shall reside outside the distributed template tree and shall not ship.

**REQ-RNT-013** — When a command cited under MP-8 fails to run, is refused for not matching the REQ-RNT-001 form, names an operation this repository's execution discipline prohibits — a local full test suite being the standing example — or does not return within the auditor's existing Bash timeout, the plan-auditor shall not execute it further, shall not record the criterion as a pass, and shall apply the REQ-RNT-003 demotion. Repository execution discipline shall take precedence over a criterion's citation. The timeout disposition adds no new mechanism: the auditor's Bash tool already bounds every invocation, so a conforming-but-expensive command is refused on the same terms as a prohibited one rather than stalling the audit.

**REQ-RNT-014** — The repository-local Go test shall assert both the presence of the MP-8 sentinel span in `plan-auditor.md` and the presence of the MP-8 row in that file's `## Must-Pass Results` report template, so that MP-8's disappearance from the agent file becomes a test failure rather than a silent non-execution.

**REQ-RNT-015** — When MP-8 re-executes a cited command whose verdict is derived from a test-runner's output, the plan-auditor shall key the verdict on the count of tests actually executed and shall not treat the presence of an `ok` token as evidence that the RED failed to reproduce. Reason, measured on this tree: `internal/hook/evidence_writer.go` `deriveFromOutputText` matches `ok  \t` as a precise pass marker and returns before inspecting any count, so `ok … [no tests to run]` — a run that executed nothing — is recorded as an observed PASS. An L2 verdict routed through that shape would record a zero-test re-run as "RED did not reproduce", inverting exactly what L2 exists to detect. This requirement constrains L2's own verdict rule only; repairing `evidence_writer.go` belongs to card t341 and is out of scope.

---

## C. HOW (approach, not implementation)

The mechanical seam is not invented here. `internal/spec/ac_count_clause_test.go` (card t338,
landed) already establishes the pattern this SPEC reuses: a repository-local Go test that
extracts a command from an agent file between `# MOAI-AC-COUNTER-BEGIN` / `-END` sentinels,
asserts exactly-one-and-non-empty *before* comparing (so an anchor matching zero or two spans
cannot become a vacuous pass), keeps fixtures under `internal/spec/testdata/ac_count/`, and
ships nothing. Its own header records why the anchor is a sentinel pair rather than a prose
structure: a prose anchor breaks the moment the clause it anchors on is rewritten, which is the
twin-drift failure this repository has already met in its `.sh` / `.sh.tmpl` pairs.

Simplicity ladder: reuse the existing pattern rather than build a checker.

---

## D. Exclusions

### Out of Scope — post-implementation audit
- `sync-auditor` and any sync-phase criterion. Sync-phase asks a different question (did the
  implementation satisfy the criterion), and the pre-implementation tree the RED was pinned to is
  no longer the current tree, so L2's re-execution premise does not hold there.

### Out of Scope — retroactive rewriting of existing SPECs
- The `acceptance.md` RED-cell format of the 706 SPEC directories under `.moai/specs/`. MP-8
  binds SPECs audited after it lands; no existing acceptance table is rewritten by this work.
- `SPEC-TODO-SQLITE-001` specifically. It is cited as the diagnostic instance (§A.2), not
  scheduled for repair here.

### Out of Scope — lexical discrimination
- Any tense, mood, or word-list discriminator over RED-cell prose. Card t342 measured
  natural-language lexical discriminators to be unsound in both directions, including 7 false
  positives where a SPEC's live criteria were classified as ghosts on vocabulary alone. This
  exclusion is also stated positively as REQ-RNT-002 because it is a design constraint, not only
  an omission.

### Out of Scope — adjacent cards
- t341 (selector counting). It shares the t241 root but owns a different instrument. Genuine
  scope overlap, if found, is recorded as a coordination note in `plan.md` rather than absorbed.
- t344, t345.

### Out of Scope — the aggregate scoring rubric
- The Group 1-8 category scores and the per-tier PASS thresholds (Tier S 0.75 / M 0.80 / L 0.85).
  MP-8 is a firewall item, deliberately outside the score.

---

## E. Audit debt — the score does not describe this text

> **Section-placement note.** The instruction naming this record asked for `spec.md` §H. This file
> has no §H: its top-level sections are HISTORY, §A WHY, §B WHAT, §C HOW, §D Exclusions. The
> cross-reference section §H lives in `plan.md`. The debt is recorded here as a new §E rather than
> under an invented §H heading, so the reference resolves to something that exists.

[HARD] **`0.800` is the plan-auditor's score for the iteration-2 text. It is not a measurement of
the text you are reading.**

| | |
|---|---|
| Score | **0.800** (Tier M threshold 0.80) |
| What it measured | the SPEC as it stood at plan-audit iteration 2 |
| Verdict at that score | **FAIL** — one critical blocking defect (N1) |
| Audits run since | **none** |

After iteration 2 returned FAIL, this SPEC was revised to close **one critical** defect (N1) and
**four major** ones (N2, N3, N5, N7), plus a minor (N6), and a subsequent correction round replaced
the §A.2 evidence, reframed the census from one cell to nine, and added REQ-RNT-015. **No audit has
seen any of that.** The iteration cap is 2 (`.moai/config/sections/harness.yaml:77`,
`plan_audit_tier_ceilings: M: 2`), it is reached, and the operator ruled to accept within the cap
rather than run a third iteration.

N1's closure was confirmed by the author and by the coordinator. That is **not an independent
audit**, and it is recorded as a debt for exactly the reason this SPEC exists: a claim confirmed by
the parties who produced it is a claim, not a measurement. This SPEC's own §A.2.1 is the standing
counter-example — an assertion that survived two *independent* re-executing audits and was still
wrong.

Attaching the value to what it measured is the same discipline REQ-RNT-001 imposes on every RED
cell, and follows the house form for a score whose subject has moved: a score labelled as belonging
to the version it graded, or as pre-repair. Anyone opening this SPEC and reading `0.800` as its
current quality is making the error the SPEC prohibits.

**What is NOT debt.** Mutant **M-6** — that MP-8 re-executes a cited command but does not verify
the command measures its stated premise — is a **design fact**, not an unpaid item. It is a
deliberate, recorded boundary of the three-layer mechanism (`acceptance.md` §D.2, §D.5; spec.md
§A.2.2), it is not scheduled for closure, and REQ-RNT-015 narrows only the `ok … [no tests to run]`
case by design. Mutant **M-3** and the Go test's own non-execution are likewise carried residuals,
named where they hold. Widening the gate to cover any of them would add criteria, and the AC axis
stands at 16/16 — that is a tier decision belonging to the lead.