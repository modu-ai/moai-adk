# Acceptance: SPEC-CONTAINMENT-MAPPING-DOC-001

Verification layer for `spec.md`. Every command runs from the repository (worktree) root. `DOC` below means `.moai/docs/multi-agent-containment-mapping.md`.

**Document-level pin.** Every RED-now observation in this file was taken on tree `629bb51cda73fc70422769988b8c28182c9bd2d5`, before the document existed. A criterion-level pin would override this one; none does.

**Carrier.** The four RED-now elements (command, verbatim stdout, exit code, tree SHA) live in the evidence ledger in § Evidence ledger, because a table cell mangles pipes and quotes. Each criterion cites its ledger entries by id.

**Reading the stdout field.** For a command run on the absent document the field holds the tool's own error text, which is the observed output. For an absence criterion (E08, E09) green is stdout `0` with the document present; `grep` exits 1 on a zero count, and exit 2 means the file is missing, which is red.

## Criteria

### AC-CMM-001 — The document exists
- **Given** the run phase has finished, **When** the existence command in E01 runs, **Then** it exits 0.
- Release-blocking. RED reason: the document does not exist yet. Green path: M2 creates it. Ledger: E01.

### AC-CMM-002 — The reference-only statement is present
- **Given** the document exists, **When** E02 counts occurrences of the literal token `reference-only`, **Then** stdout is a number of 1 or more.
- Release-blocking. RED reason: the document does not exist. Green path: M2. Ledger: E02.

### AC-CMM-003 — The source register has four verified-or-flagged rows
- **Given** the document exists, **When** E03 counts register rows whose first cell is BP-1, BP-2, BP-5, or BP-6, that carry a double-quoted fragment of at least 20 characters, and that end with `VERIFIED` or `UNVERIFIED`, **Then** stdout is `4`.
- Release-blocking. RED reason: the document does not exist. Green path: M2 writes the register, M1's flags fill it. Ledger: E03.

### AC-CMM-004 — Every URL line carries a flag
- **Given** the document exists, **When** E04a counts lines containing `https://` and E04b counts lines containing `https://` that end with a flag cell, **Then** the two numbers are equal and each is 4 or more.
- Release-blocking. RED reason: the document does not exist. Mutant that this closes: a URL pasted into prose without a flag raises E04a above E04b. Green path: M2. Ledger: E04a, E04b.

### AC-CMM-005 — The mapping table covers the three themes with well-formed rows
- **Given** the document exists, **When** E05 scans the theme rows, **Then** it exits 0: every theme row ends in `aligned`, `differs`, or `gap`, every non-gap row contains a `§`, and each of the three themes has at least one row.
- Release-blocking. RED reason: the document does not exist. Green path: M3. Ledger: E05.

### AC-CMM-006 — At least one gap row exists
- **Given** the document exists, **When** E06 counts rows ending in `| gap |`, **Then** stdout is 1 or more.
- Release-blocking. RED reason: the document does not exist. The seed (`plan.md` § D) carries two gap rows; a document that finds a defense for both must say so in the row text. Green path: M3. Ledger: E06.

### AC-CMM-007 — Every cited path resolves, and enough paths are cited
- **Given** the document exists, **When** E07 extracts each `.claude/…md` and `.moai/…md` path and tests it with `test -f`, **Then** it exits 0, which requires no missing path and at least eight citations.
- Release-blocking. RED reason: the document does not exist. The count floor exists because a sweep that selected nothing asserts nothing. Green path: M3. Ledger: E07.

### AC-CMM-008 — No identifiers, years, or the word `develop`
- **Given** the document exists, **When** E08 counts lines containing a card identifier shape, a SPEC identifier, a four-digit year starting 19 or 20, or the word `develop`, **Then** stdout is `0`.
- Release-blocking. RED reason: the document does not exist (exit 2). Green path: M2 to M4. Ledger: E08.

### AC-CMM-009 — No rule-marker tokens
- **Given** the document exists, **When** E09 counts lines containing `[HARD]` or `[ZONE`, **Then** stdout is `0`. This is the mechanical proxy for "no new `[HARD]` clause and no restated rule text".
- Release-blocking. RED reason: the document does not exist (exit 2). Green path: M2 to M4. Ledger: E09.

### AC-CMM-010 — Nothing outside the document and the SPEC directory changed
- **Given** the plan-phase commit SHA is known, **When** E10 diffs it against HEAD excluding the document and this SPEC directory, **Then** it exits 0.
- **Regression-guard, not release-blocking.** A scope guard is green on the pre-implementation tree by construction, so no RED-now observation exists for it; its red was instead observed on a known-different pair (ledger E10 probes). Fill `<PLAN_COMMIT_SHA>` with the commit that carries these plan artifacts. Ledger: E10.

## Reviewer-checked criteria (no mechanical command)

These requirements have no discriminator that is sound on structure alone, and a vocabulary check would be unsound in both directions. Each is a criterion with a reviewer procedure. The reviewer records the result; none is recorded as a mechanical pass.

### AC-CMM-011 — Every VERIFIED flag has an authoring-run record (REQ-CMM-004)
- **Given** the document carries a source register, **When** a reviewer reads the run-phase evidence, **Then** each source flagged `VERIFIED` has, from the authoring run, the fetch and the fragment match output.
- Reviewer-checked.

### AC-CMM-012 — Rows citing an UNVERIFIED source read as leads (REQ-CMM-005)
- **Given** a source is flagged `UNVERIFIED`, **When** a reviewer reads the mapping rows citing it, **Then** each row reads as a lead awaiting confirmation. Dormant while all four flags are `VERIFIED`.
- Reviewer-checked.

### AC-CMM-013 — Attribution and scale framing of the merge-queue rows (REQ-CMM-009)
- **Given** the document carries `merge-queue-alternative` rows, **When** a reviewer reads those rows and the scale caveat, **Then** the source's volume figures and its post-merge-testing conclusion are attributed to the source, and no adoption by the factory is recommended.
- Reviewer-checked.

### AC-CMM-014 — Defense cells paraphrase and point (REQ-CMM-010, second clause)
- **Given** the document carries defense cells, **When** a reviewer reads each against its cited section, **Then** the cell paraphrases and points rather than copying rule text. AC-CMM-009 cannot see a paraphrased restatement.
- Reviewer-checked.

## Traceability

| Requirement | Criterion |
|---|---|
| REQ-CMM-001 | AC-CMM-001, AC-CMM-010 |
| REQ-CMM-002 | AC-CMM-002 |
| REQ-CMM-003 | AC-CMM-003, AC-CMM-004 |
| REQ-CMM-004 | AC-CMM-011 |
| REQ-CMM-005 | AC-CMM-012 |
| REQ-CMM-006 | AC-CMM-005 |
| REQ-CMM-007 | AC-CMM-005 (section marker), AC-CMM-007 |
| REQ-CMM-008 | AC-CMM-006 |
| REQ-CMM-009 | AC-CMM-013 |
| REQ-CMM-010 | AC-CMM-008, AC-CMM-009, AC-CMM-014 |
| REQ-CMM-011 | AC-CMM-010 |

## Evidence ledger

Measured on tree `629bb51cda73fc70422769988b8c28182c9bd2d5`, with the document absent. Exit codes are recorded as their own field because the single-invocation command form cannot carry `; echo $?`.

```text
E01
command: test -f .moai/docs/multi-agent-containment-mapping.md
stdout:
exit: 1
```

```text
E02
command: grep -c -i -F 'reference-only' .moai/docs/multi-agent-containment-mapping.md
stdout: grep: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E03
command: grep -c -E '^\| BP-(1|2|5|6) \|.*"[^"]{20,}".*\| (VERIFIED|UNVERIFIED) \|$' .moai/docs/multi-agent-containment-mapping.md
stdout: grep: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E04a
command: grep -c -F 'https://' .moai/docs/multi-agent-containment-mapping.md
stdout: grep: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E04b
command: grep -c -E 'https://.*\| (VERIFIED|UNVERIFIED) \|$' .moai/docs/multi-agent-containment-mapping.md
stdout: grep: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E05
command: awk -F'|' '/^\| (correlated-failure|isolation-first|merge-queue-alternative) \|/ { k=$2; gsub(/ /,"",k); t[k]++; f=$(NF-1); gsub(/ /,"",f); if (f!="aligned" && f!="differs" && f!="gap") { print "BAD-FIT: " $0; bad=1 } if (f!="gap" && index($0,"§")==0) { print "NO-SECTION: " $0; bad=1 } } END { if (t["correlated-failure"]<1 || t["isolation-first"]<1 || t["merge-queue-alternative"]<1) { print "THEME-MISSING"; bad=1 } exit bad }' .moai/docs/multi-agent-containment-mapping.md
stdout: awk: can't open file .moai/docs/multi-agent-containment-mapping.md
 source line number 1
exit: 2
```

```text
E06
command: grep -c -E '\| gap \|$' .moai/docs/multi-agent-containment-mapping.md
stdout: grep: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E07
command: awk '{ s=$0; while (match(s, /\.(claude|moai)\/[A-Za-z0-9_.\/-]+\.md/)) { p=substr(s,RSTART,RLENGTH); n++; if (system("test -f " p) != 0) { print "MISSING: " p; bad=1 } s=substr(s,RSTART+RLENGTH) } } END { if (n<8) { print "TOO-FEW-PATHS n=" n; bad=1 } exit bad }' .moai/docs/multi-agent-containment-mapping.md
stdout: awk: can't open file .moai/docs/multi-agent-containment-mapping.md
 source line number 1
exit: 2
```

```text
E08
command: grep -c -w -i -E 't[0-9]{3,4}|SPEC-[A-Z][A-Z0-9-]*|(19|20)[0-9]{2}|develop' .moai/docs/multi-agent-containment-mapping.md
stdout: grep: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E09
command: grep -c -F -e '[HARD]' -e '[ZONE' .moai/docs/multi-agent-containment-mapping.md
stdout: grep: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E10
command: git diff --quiet <PLAN_COMMIT_SHA> HEAD -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001'
RED-now: not applicable (regression-guard; green on the pre-implementation tree by construction)
red probe (known-different pair): git diff --quiet HEAD~1 HEAD -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001'  ->  exit 1
green probe (identical pair):     git diff --quiet HEAD HEAD -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001'  ->  exit 0
```

## Instrument validation (observed failure on constructed inputs)

A check is complete only after its failure has been observed on a known failing input. E01 to E09 were run on two constructed fixture documents kept outside the repository: a passing fixture (four register rows, nine mapping rows across the three themes, eight resolving paths, one gap row, no forbidden token) and a failing fixture (an unflagged URL, a flag of `MAYBE`, a theme row with Fit `maybe`, a non-gap row with no `§`, one missing theme, a nonexistent path, and the tokens a card identifier, a SPEC identifier, a dated year, `[HARD]`, and `develop`).

| Command | Passing fixture | Failing fixture |
|---|---|---|
| E01 | exit 0 | exit 0 (fixture exists; not a discriminating command) |
| E02 | 1 | 0, exit 1 |
| E03 | 4 | 0, exit 1 |
| E04a / E04b | 4 / 4 | 2 / 0 (unequal, red) |
| E05 | exit 0, no output | exit 1, `BAD-FIT`, `NO-SECTION`, `THEME-MISSING` |
| E06 | 1 | 0, exit 1 |
| E07 | exit 0, no output | exit 1, `MISSING` and `TOO-FEW-PATHS` |
| E08 | 0, exit 1 | 2, exit 0 |
| E09 | 0, exit 1 | 1, exit 0 |

The first run of E07 on the passing fixture exited 1 with `TOO-FEW-PATHS n=7`, because the fixture cited seven paths; adding one path made it exit 0. That run is the observed red of the count floor.

Token-by-token probe of E08 on one-line inputs: a card identifier `1`, `t99` `0` (shorter than three digits, not a card shape), a four-digit year starting 20 `1`, a four-digit year starting 19 `1`, a SPEC identifier `1`, `develop` `1`, `Develop` `1`, `v3.2.0` `0`, `30,000` `0`, `2,500` `0`, plain text `0`. The two volume figures the document will quote do not trip the check.

Known limit: the E08 `develop` token and the four-digit year check count lines, not occurrences; a failing line reports 1 however many tokens it holds.

## Definition of Done

- AC-CMM-001 through AC-CMM-009 green on the final document, each with its command and verbatim output in the run-phase evidence, against the tree HEAD named there.
- AC-CMM-010 green against the plan-commit SHA.
- AC-CMM-011 through AC-CMM-014 recorded by a reviewer.
- Every source flag in the document equals the outcome of the authoring-run fetch and fragment match.
- No file outside the document and this SPEC directory changed, and no file under `internal/template/templates/` changed.
