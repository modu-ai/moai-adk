# SPEC-RED-NOW-THRESHOLD-001 — Acceptance Criteria

This SPEC is governed by the rule it edits, so its own criteria are written under that rule.

Every criterion names **WHEN** it runs, the **INPUT** that turns it red, and **WHERE** the red
surfaces (`verification-completeness.md` §1.2). Every criterion carries the two-cell pair of §2.

**Document-level tree pin: `a6bbbf82b`** (`WT-red-now-threshold`). This pin **binds every cell and
every ledger entry below that carries no pin of its own** — explicit inheritance, not an implied
one. It is a commit SHA, never a branch name (rule §4).

**Verbatim means the raw file bytes**, not the GFM-rendered view. Every command below is
copy-runnable as it appears in the source.

Criteria with no observed RED are classified **regression-guard** and say so, per the
`SPEC-TODO-LANDING-STATE-001` §C precedent (spec.md §A.3). They are not dressed with an invented
failure and are not counted as release gates.

## D.0 Evidence ledger — every cited command, outside the table

Iteration 1 found two defective commands, both caused by shell syntax living inside a GFM table
cell (escaped pipes). The corrective applied here is structural rather than per-cell: **no command
appears in a table cell at all.** Cells cite a ledger id; the ledger is fenced, so what is written
is what runs.

Every entry is a **read-only, single-invocation** command per REQ-RNT-001 — no pipes, no
redirection, no `&&`, no `;` chaining, no subshells. The exit code is a separate field because
that form forbids `; echo $?`.

```
E-01  grep -c "RED-now cell content" .claude/rules/moai/development/verification-completeness.md
      stdout: 0
      exit:   1

E-02  grep -c -e tense -e mood -e counterfactual -e "future.sense" .claude/rules/moai/development/verification-completeness.md .claude/agents/moai/plan-auditor.md
      stdout: .claude/rules/moai/development/verification-completeness.md:0
              .claude/agents/moai/plan-auditor.md:0
      exit:   1

E-03  grep -c "regression-guard" .claude/rules/moai/development/verification-completeness.md
      stdout: 0
      exit:   1

E-04  grep -c "MP-8" .claude/agents/moai/plan-auditor.md
      stdout: 0
      exit:   1

E-05  ls internal/spec/red_now_cell_test.go
      stdout: (empty)
      stderr: ls: internal/spec/red_now_cell_test.go: No such file or directory
      exit:   1

E-06  grep -c "AC-6:" .claude/agents/moai/plan-auditor.md
      stdout: 0
      exit:   1

E-07  grep -c "MOAI-REDNOW-BEGIN" .claude/agents/moai/plan-auditor.md
      stdout: 0
      exit:   1

E-08  ls internal/spec/testdata/red_now/
      stdout: (empty)
      stderr: ls: internal/spec/testdata/red_now/: No such file or directory
      exit:   1

E-09  grep -c "MOAI-REDNOW-BEGIN" internal/template/templates/.claude/agents/moai/plan-auditor.md
      stdout: 0
      exit:   1

E-10  grep -nE "SPEC-[A-Z]+-[0-9]{3}" internal/template/templates/.claude/agents/moai/plan-auditor.md
      stdout: 446:This agent receives one input: the absolute path to the SPEC directory (e.g., `.moai/specs/SPEC-AUTH-001/`).
              464:- "Use the plan-auditor subagent to audit the SPEC at .moai/specs/SPEC-AUTH-001/ — this is iteration 1"
      exit:   0

E-11  grep -rl "red_now" internal/template/templates/
      stdout: (empty)
      exit:   1

E-15  grep -rn "func TestMigrationParity" internal/kanban/
      stdout: internal/kanban/backlog_migrate_test.go:70:func TestMigrationParity(t *testing.T) {
              internal/kanban/backlog_migrate_test.go:424:func TestMigrationParityCatchesTamperedRecord(t *testing.T) {
      exit:   0
      note:   establishes that iterations 1-2 cited an EXISTING test as their refuting command
              (spec.md §A.2.1). Multi-invocation runners are excluded from the REQ-RNT-001 form;
              E-16 below is recorded as observed evidence, not as a conforming citation.

E-16  go test ./internal/kanban -run TestMigrationParityDoesNotExistXYZ -count=1
      stdout: ok  	github.com/modu-ai/moai-adk/internal/kanban	0.219s [no tests to run]
      exit:   0

E-17  grep -c "hasPrecisePass" internal/hook/evidence_writer.go
      stdout: 2
      exit:   0

E-12  grep -c "^| \*\*AC-RNT-" .moai/specs/SPEC-RED-NOW-THRESHOLD-001/acceptance.md
      stdout: 16
      exit:   0

E-13  grep -c "^| \*\*AC-RNT-.* | release-blocking |" .moai/specs/SPEC-RED-NOW-THRESHOLD-001/acceptance.md
      stdout: 13
      exit:   0

E-14  grep -c "^| \*\*AC-RNT-.* | \*\*regression-guard\*\*" .moai/specs/SPEC-RED-NOW-THRESHOLD-001/acceptance.md
      stdout: 3
      exit:   0
```

All seventeen measured on `a6bbbf82b`, in this worktree, by running the command. E-12, E-13 and
E-14 are anchored at `^| \*\*AC-RNT-`, so a prose or ledger line mentioning them cannot satisfy
them — the property the unanchored predecessor lacked.

### D.0.1 E-02 divergence probe — why this command is not vacuous

E-02 returning zero is only meaningful if the command *can* return non-zero. Iteration 1 found the
previous form (`grep -cE "tense\|mood\|…"`, an escaped pipe inside a table cell) could not: in ERE
`\|` is a literal pipe, so the pattern matched nothing even when a discriminator word was present.
The replacement was adopted only after observing the divergence directly, per rule §5:

```
# fixture A: a file containing the word "mood"; fixture B: a file containing no discriminator
$ grep -cE "tense\|mood\|counterfactual\|future.sense" <fixture-A>
0                                        exit 1      ← old form: FALSE GREEN on a planted word
$ grep -c -e tense -e mood -e counterfactual -e "future.sense" <fixture-A>
1                                        exit 0      ← E-02 form: true RED
$ grep -c -e tense -e mood -e counterfactual -e "future.sense" <fixture-B>
0                                        exit 1      ← E-02 form: true green
```

Measured on `a6bbbf82b` against scratch fixtures. E-02's `0:0` on the real files is therefore a
verdict, not uninterpreted output.

## D. AC Matrix

| AC | Class | Verifies | Scenario (Given-When-Then) | RED-now proof | Green path |
|----|-------|----------|-----------------------------|----------------|------------|
| **AC-RNT-001** | release-blocking | REQ-RNT-001 | **Given** `verification-completeness.md` **when** §2 is read for the release-blocking RED-cell obligation **then** it requires the command, its verbatim stdout, its exit code, and a tree SHA together. | **E-01** — red because the clause does not exist; §2 today names none of the four elements. | M1 lands the clause; **M3 asserts mechanically**, inside the extracted span: four element tokens present (`command`, `stdout`, `exit`, `SHA`), the words `read-only` and `single invocation` present, and the words `raw file` present. Subtest `TestRuleClauseEnumeratesFourElements`. Scoping the assertion to the extracted §2 span defeats a token pasted anywhere **outside** that span; it does **not** defeat one pasted **inside** it, and the §2 span is ~41 lines, so M-3 survives within it. The residual is recorded in §D.2 rather than claimed closed. |
| **AC-RNT-002** | **regression-guard** — *RED cell: none, deliberately* | REQ-RNT-002 | **Given** the two edited surfaces **when** they are scanned for a tense/mood/word-list discriminator **then** none is present. | **No RED, and no invented one.** Measured green: **E-02**, whose non-vacuity is established by the divergence probe in §D.0.1. This is the property being **preserved**, so it is a regression-guard, not a release gate. | E-02 still returns `…:0` / `…:0` at exit 1 after M1, M2, M4. A failure is a contract break; there is no RED to flip. |
| **AC-RNT-003** | release-blocking | REQ-RNT-003 | **Given** a RED that cannot be re-executed **when** §2 is consulted for its disposition **then** the criterion loses release-blocking eligibility and becomes a regression-guard, and is not recorded as a pass. | **E-03** — red because the rule has no disposition for the undecidable case; the term does not appear at all. | M1 lands the clause; **M3 asserts** inside the extracted span that it contains both `regression-guard` and a negation of pass (`not … pass`), and that the demotion is stated as the disposition rather than as an option. Subtest `TestRuleClauseStatesDemotionNotPass`. |
| **AC-RNT-004** | release-blocking | REQ-RNT-004 | **Given** a SPEC with a release-blocking criterion **when** plan-auditor runs MP-8 **then** it re-executes that criterion's cited command against the current tree and confirms the RED reproduces. | **E-04** — red because no MP-8 exists; the firewall runs MP-1..MP-7. | M2 lands MP-8; **M3 asserts** inside the extracted MP-8 span that it names re-execution (`re-execute`) and the current tree, and that the span is reachable from the `### M5` heading rather than from anywhere in the file. Subtest `TestMP8SpanNamesReexecution`. |
| **AC-RNT-005** | release-blocking | REQ-RNT-005 | **Given** an unresolved MP-8 violation **when** the report is produced **then** it appears in `## Defects Found` at `severity=critical` and `Verdict: FAIL` is set regardless of the aggregate score. | **E-05** — red because the test asserting the MP-8 span carries both obligations does not exist. | M2+M3: subtest `TestMP8SpanIsScoreIndependent` asserts `severity=critical` and a score-independence phrase, both **inside** the extracted span. |
| **AC-RNT-006** | release-blocking | REQ-RNT-006 | **Given** a SPEC with no release-blocking criterion or no `acceptance.md` **when** MP-8 is evaluated **then** it is marked `N/A` with a stated reason, per the MP-4 precedent. | **E-04** — red for the same reason as AC-RNT-004 and stated as such: the clause carrying the `N/A` branch does not exist. | M2+M3: subtest `TestMP8SpanCarriesNABranch` asserts, inside the span, both an `N/A` token and a stated-reason obligation, and that the report row pattern includes `N/A` as an admissible value. |
| **AC-RNT-007** | release-blocking | REQ-RNT-007 | **Given** the audit checklist and report template **when** they are read **then** Group 4 carries an MP-8 checklist item and `## Must-Pass Results` carries an MP-8 row. | **E-06** — red because Group 4 today ends at AC-5. | M2+M3: subtest `TestGroup4AndReportRowExist` asserts an `AC-6:` item **within the `### Group 4` section span** and a line matching `- [PASS/FAIL/N/A] MP-8` **within the `## Must-Pass Results` template span**. Six characters pasted elsewhere in the file satisfy neither, because both assertions are section-scoped. |
| **AC-RNT-008** | release-blocking | REQ-RNT-008 | **Given** `plan-auditor.md` carrying the MP-8 form contract between sentinels **when** the Go test runs **then** it extracts that span rather than restating it, and fails if the sentinel pair does not match exactly once with non-empty content. | **E-07** — red because neither the sentinel pair nor the extracting test exists. | M2+M3: the sentinel pair exists exactly once; the test is **observed failing** on a mutated copy carrying zero pairs and again on one carrying two (mutant M-2, §D.2); and subtest `TestCommandScopeIsCarrierIndependent` collects commands from **all three carriers** — table cells, fenced ledger entries, and other fenced blocks — verified against a third fixture `testdata/red_now/ledger/` whose commands live only in a ledger. The subtest is **observed failing** on a mutant that relocates a malformed command from a cell into a ledger entry (M-5, §D.2). |
| **AC-RNT-009a** | release-blocking | REQ-RNT-009 | **Given** a planted fixture whose RED cell is prose-only (no command, no stdout, no exit code, no SHA) **when** the form check runs on it **then** it is reported as violating. | **E-08** — red because the fixture directory does not exist, so no violation can be reported. | M3: `testdata/red_now/violating/` exists and the check reports it violating with a non-zero finding count. |
| **AC-RNT-009b** | release-blocking | REQ-RNT-009 | **Given** a fixture whose RED cell carries a command, its verbatim stdout, its exit code, and a pin **when** the form check runs on it **then** it is **not** reported as violating. | **E-08** — same absence; the pass direction is unobservable for the same reason the fail direction is. | M3: `testdata/red_now/legitimate/` exists and the check reports zero findings on it. |
| **AC-RNT-010** | release-blocking | REQ-RNT-010 | **Given** the local `plan-auditor.md` carrying the MP-8 sentinel span **when** the mirror is checked **then** the mirror carries the same span. | **E-09** — red because the mirror carries no MP-8 clause. | M4: E-09 returns ≥1 on the mirror, the extracted spans compare **byte-equal** across both carriers, and `make build` exits 0. |
| **AC-RNT-011** | **regression-guard** — *RED cell: none, deliberately* | REQ-RNT-011 | **Given** the MP-8 clause as it appears in the template mirror **when** it is scanned for a real SPEC ID, a card id, or an internal date **then** none is present. | **No RED, and no invented one.** It is green today **only because the MP-8 clause does not exist yet** (E-09 → `0`), so there is nothing to be non-neutral — the same forward-looking shape as AC-RNT-012, stated the same way. Scoping evidence: **E-10** shows two illustrative `SPEC-AUTH-001` placeholders already present. A count-equals-zero criterion over the whole file would be red at arrival and red forever for reasons this work never touches — the impossible / wrong-reason-red direction of rule §2 — so the criterion is scoped to the added clause. | M4: the extracted MP-8 span contains no real SPEC ID, no `t343`, and no ISO date. A failure is a contract break; there is no RED to flip. |
| **AC-RNT-012** | **regression-guard** — *RED cell: none, deliberately* | REQ-RNT-012 | **Given** the distributed template tree **when** it is scanned for the new test or its fixtures **then** neither is present. | **No RED, and no invented one.** Measured green: **E-11**. It is green today only because the artifact does not exist yet, which is precisely why it is **not** a release gate: it is a forward-looking non-leak guard, honest about carrying no starting observation. | M3+M4: E-11 still returns empty stdout at exit 1 after the test and fixtures land. |
| **AC-RNT-013** | release-blocking | REQ-RNT-013 | **Given** a cell citing a command that fails, is refused for not matching the REQ-RNT-001 form, or names a locally prohibited operation **when** MP-8 evaluates it **then** the auditor does not execute it, does not record a pass, and applies the REQ-RNT-003 demotion. | **E-05** — red because no test asserts the MP-8 span carries an execution-discipline branch, and E-04 shows the span itself is absent. | M2+M3: subtest `TestMP8SpanCarriesExecutionDiscipline` asserts, inside the span, a refusal branch, a not-a-pass disposition, and a precedence statement naming repository execution discipline. |
| **AC-RNT-014** | release-blocking | REQ-RNT-014 | **Given** a future edit that deletes MP-8 from `plan-auditor.md` **when** the Go test runs **then** it fails, so MP-8's disappearance is distinguishable from MP-8 passing. | **E-05** — red because the test does not exist, so today MP-8's absence and MP-8's success are indistinguishable — which is the §1.3 defect this criterion closes. | M3: subtest `TestMP8LivenessAnchors` asserts BOTH the sentinel span AND the `- [PASS/FAIL/N/A] MP-8` report-template row, and is **observed failing** on a mutated copy with the MP-8 row deleted. |
| **AC-RNT-015** | release-blocking | REQ-RNT-015 | **Given** an MP-8 re-execution whose cited command is a test runner **when** the run matches zero tests **then** the auditor records a failure to reproduce the RED, not a pass. | **E-05** — red because no test asserts the MP-8 span carries an executed-count rule; **E-16** shows the input that makes the naive rule wrong (`ok` with `[no tests to run]`, exit 0) and **E-17** shows the seam that would misread it (`deriveFromOutputText` returns on `hasPrecisePass` before inspecting any count). | M2+M3: subtest `TestMP8SpanKeysOnExecutedCount` asserts, inside the extracted MP-8 span, that the verdict rule names an executed-test count and explicitly rejects an `ok`-token-only verdict. |

## D.1 Traceability

| REQ | Covering AC |
|-----|-------------|
| REQ-RNT-001 | AC-RNT-001 |
| REQ-RNT-002 | AC-RNT-002 |
| REQ-RNT-003 | AC-RNT-003 |
| REQ-RNT-004 | AC-RNT-004 |
| REQ-RNT-005 | AC-RNT-005 |
| REQ-RNT-006 | AC-RNT-006 |
| REQ-RNT-007 | AC-RNT-007 |
| REQ-RNT-008 | AC-RNT-008 |
| REQ-RNT-009 | AC-RNT-009a, AC-RNT-009b |
| REQ-RNT-010 | AC-RNT-010 |
| REQ-RNT-011 | AC-RNT-011 |
| REQ-RNT-012 | AC-RNT-012 |
| REQ-RNT-013 | AC-RNT-013 |
| REQ-RNT-014 | AC-RNT-014 |
| REQ-RNT-015 | AC-RNT-015 |

No orphaned AC; no uncovered REQ. Release-blocking: 13. Regression-guard: 3 (AC-RNT-002, -011,
-012). Tier M budget: 15 REQ / 16 AC. The AC count now sits **exactly at** the Tier M ceiling of
16; a further criterion requires tiering up or splitting, not relaxing the budget.

Counts are measured by ledger entries **E-12, E-13, E-14** (§D.0). Iteration 2 found the previous
release-blocking count command self-matching — it lived in this paragraph unanchored, so the
sentence stating the count was itself counted, returning 13 for a true value of 12. All three
count commands are now anchored to `^| \*\*AC-RNT-` and live in the fenced ledger, so no prose
line can satisfy them. This was the same defect class as D1: a command that does not measure what
the sentence around it claims.

## D.2 Mutant probe (rule §2)

Iteration 1 ran this probe against AC-RNT-008 only — the one criterion already immune to the
mutant that matters. It is now run against the shallow criteria too.

**Mutant M-1 — fabricated output.** An author writes a RED cell carrying a plausible command, a
fenced block of invented output, and a copied SHA, without running the command. This mutant
**satisfies L1 while violating REQ-RNT-001's intent**, and it is writable. L1 is therefore too
shallow to adopt alone. Disposition: it is not adopted alone — AC-RNT-004 (L2 re-execution)
catches M-1 by running the command and observing it not reproduce RED. This is why spec.md §B
states the two layers are adopted as a pair.

**Mutant M-2 — zero-match extraction.** A sentinel pair is added twice, or once with an empty
body, so the extractor reads nothing and every comparison passes vacuously. Caught by the
exactly-one-non-empty assertion running *before* any comparison, inherited from
`ac_count_clause_test.go`. AC-RNT-008's green path requires the test to be **observed failing** on
both a zero-pair and a two-pair mutated copy.

**Mutant M-3 — token insertion (the iteration-1 finding).** For AC-RNT-001, -003, -004, -006 and
-007, the previous green paths were `grep -c "<token>" <file>` ≥ 1 **AND** a prose conjunct no
command could judge. Pasting the token anywhere in the file — a comment, a table of contents —
satisfied the mechanical half while changing no behaviour. AC-RNT-007 was the shallowest: the six
characters `AC-6:` anywhere in a 482-line file.

Disposition: **the five green paths above no longer contain a file-wide grep.** Each is now a
named M3 subtest asserting its predicate **inside an extracted section span** — the MP-8 span, the
`### Group 4` span, the `## Must-Pass Results` template span. M-3 no longer satisfies any of them,
because a token outside the relevant span is not inside it. The span-extraction seam is the one
REQ-RNT-008 already builds, so no new mechanism was introduced. Governing clause: rule §5
Audit-verification — the cheapest modification satisfying a grep is not the modification that
fixes the defect.

**Mutant M-4 — class laundering.** An author moves a criterion carrying a broken command from
release-blocking to regression-guard, and the command is never checked again. This mutant was
**live in iteration 1 of this very SPEC**: both defective commands sat in regression-guard cells,
where MP-8's release-blocking scope missed them by construction. Disposition: closed on the **class** axis by
re-scoping L1 to bind every criterion carrying a command regardless of class (REQ-RNT-008,
spec.md §B). Demotion still exempts a criterion from the RED obligation; it no longer exempts the
command from being runnable. The **structural** axis is M-5.

**Mutant M-5 — carrier relocation (the iteration-2 finding).** An author moves a malformed command
out of the checked carrier. This mutant was **live in iteration 2 of this very SPEC, and this
document created it**: the D1 corrective moved every command from table cells into the §D.0
ledger, while the D4 corrective scoped L1 to "any **cell** … carries a command". The two
correctives cancelled — applied to this document, class-independent L1 had **zero** commands in
scope, and §D.0 institutionalised the evasion as the recommended pattern. Closing only the class
axis left the structural axis wide open, and the structural one is cheaper to walk through.
Disposition: REQ-RNT-001 admits the ledger as a **named carrier** in the requirement layer and
REQ-RNT-008's predicate follows the command to whichever carrier holds it — cell, ledger entry, or
any fenced block. AC-RNT-008 requires `TestCommandScopeIsCarrierIndependent` to be observed
failing on a relocation mutant, so the closure is measured rather than argued.

**Mutant M-6 — a command whose premise is false (the correction-round finding).** An author cites
a command that runs, whose output is exactly as quoted, and whose RED reproduces on re-execution —
but which does not measure the premise it is said to measure. This SPEC's own §A.2 carried one for
two iterations: `go test -run TestMigrationParity` was cited as proving an absent test yields
green, while the test existed, so a real passing test was being read as a vacuous one. It survived
two audits that each re-executed it. **L2 does not catch M-6**, and no clause here claims it does:
re-execution confirms output, not that the command measures its claim. The residual is recorded
because the alternative — asserting MP-8 catches everything — is the overclaim shape this SPEC
exists to prevent. What narrows M-6 in practice is REQ-RNT-015 for the specific and common case of
a test runner that executed nothing (`ok … [no tests to run]`), and nothing else here narrows it
further. The general case is open.

**M-3 residual, stated rather than claimed closed.** Span-scoping defeats a token pasted outside
the relevant span. It does not defeat one pasted inside it, and the §2 span is roughly 41 lines —
so for AC-RNT-001 in particular, M-3 survives within the span. Narrowing further would need a
positional or ordering predicate over clause structure, which buys less than it costs at this
size. The criterion is therefore **narrowed, not immune**, and AC-RNT-001's green path now says so
instead of claiming a comment cannot satisfy it.

## D.3 Continued-firing (rule §1.3 — how a reader learns the check stopped running)

Rule §1.3 asks a question distinct from contract drift: **absent execution, not suppressed
failure.** Iteration 1's answer here addressed only drift. Both axes are separated now.

### Axis 1 — the contract changed (covered)

- The Go test **extracts** the contract from `plan-auditor.md` rather than restating it, so a
  rewritten clause changes what the test asserts instead of leaving it asserting the old thing.
  The `.sh` / `.sh.tmpl` twin-drift failure this repository already met is the reason
  (`ac_count_clause_test.go` header).
- Extraction asserts exactly-one-non-empty **before** comparing, so removing or duplicating the
  sentinels fails loudly rather than passing on an empty extraction (M-2).
- AC-RNT-010 compares the extracted spans across both carriers byte-equal, so mirror drift is a
  test failure rather than a silent divergence.

### Axis 2 — the check stopped running at all

**MP-8: covered, by REQ-RNT-014.** MP-8 is an instruction inside a markdown agent file. If a later
rewrite drops the MP-8 row, the report still yields PASS from MP-1..MP-7 and nothing notices —
non-execution indistinguishable from success, the §1.3 defect exactly. The answer reuses the span
seam: the Go test asserts both the MP-8 sentinel span **and** the `- [PASS/FAIL/N/A] MP-8` report
row, so MP-8 vanishing from the document turns CI red. AC-RNT-014 requires that failure to be
observed on a mutated copy, not merely asserted.

**The Go test's own non-execution: NOT covered, and stated as such.** If the test file is deleted,
build-tagged out, or its package stops being run, CI stays green and nothing here detects it. This
SPEC does not close that axis. Saying so is deliberate: an answer that looks like coverage while
providing none is the §1.3 defect one level up, and the honest gap is the better artifact. Closing
it needs a repository-wide test-inventory guard — a different instrument, and one no artifact in
this SPEC builds. The boundary is stated here, at the point it is reached, because no exclusion
in `spec.md` § Exclusions carries it: the four Out of Scope sub-headings cover post-implementation
audit, retroactive SPEC rewriting, lexical discrimination, adjacent cards, and the scoring rubric,
and none of them names test-inventory liveness. Iteration 2 judged this a legitimate tool boundary
rather than a §1.3 defect wearing a disclaimer; recording it where it actually holds is what keeps
that judgment checkable.

## D.4 Closure gates

1. Every release-blocking AC passes, with the ledger id, the command, its stdout, its exit code,
   and the tree cited in `progress.md` §E.2.
2. Every regression-guard AC still returns its stated green output.
3. Both directions of REQ-RNT-009 observed (AC-RNT-009a **and** AC-RNT-009b). Confirming only the
   pass direction is indistinguishable from not having raised the threshold at all.
4. Mutants M-2 and the MP-8-row deletion of AC-RNT-014 each **observed failing**, not argued.
5. `make build` exits 0 and the mirror span comparison passes.

## D.5 Definition of Done

MP-8 exists, re-executes under a defined trust boundary, blocks score-independently, is mirrored,
and cannot vanish silently.

L1 checks every command **regardless of class and regardless of carrier**. Both qualifiers are
load-bearing and each was learned from an audit of this document: iteration 1 found that
class-scoping let two broken commands hide in regression-guard cells (M-4), and iteration 2 found
that the fix for the first defect had moved every command out of the carrier the class fix named,
leaving **zero** commands in scope (M-5). The claim this section can now make — and could not make
in the previous revision, where it was false — is that this document's own seventeen ledger
entries and its three regression-guard criteria are inside the gate, because the scope predicate
follows commands to the ledger rather than stopping at table cells.

Three residuals are not closed and are named rather than absorbed: M-3 survives inside a span,
M-6 (a command whose premise is false) is not caught by L2 at all — this document carried one for
two iterations — and the Go test's own non-execution is a tool boundary this SPEC does not reach
(§D.2, §D.3).
