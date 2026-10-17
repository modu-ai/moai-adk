# Acceptance: SPEC-CONTAINMENT-MAPPING-DOC-001

Verification layer for `spec.md`. Every command runs from the repository (worktree) root. `DOC` below means `.moai/docs/multi-agent-containment-mapping.md`.

**Command form.** Every command is one plain invocation: no pipe, no `;`, no redirection, no subshell. For awk there is also no `system()` and no `|` in the program text; a pattern that needs `|` is passed with `-v NAME='…'` and each backslash doubled, and the program matches with `$0 ~ NAME`, `match(…, NAME)`, or `!~`. The worktree-session guard refuses awk programs it cannot show to be harmless, and the run phase executes in a card worktree, so a command the guard refuses cannot be run there. The commands below were each run in a worktree session to confirm the guard admits them; the refusals seen on the way are recorded verbatim in § Guard measurements.

**Document-level pin.** Every RED-now observation in this file, except where a ledger entry carries its own `tree:` line, was taken on tree `cf5a52f55992e59fe9593fa508ace5d4dba59fc7` (the plan-phase commit, working tree clean at observation time), before the document existed. It differs from the earlier pin `629bb51cda73fc70422769988b8c28182c9bd2d5` in exactly the four SPEC files (`git diff --name-only 629bb51cda73fc70422769988b8c28182c9bd2d5 cf5a52f55992e59fe9593fa508ace5d4dba59fc7` listed `acceptance.md`, `plan.md`, `progress.md`, `spec.md` under this SPEC directory; re-observed in fix round 2), and the document does not exist on either, so the commands that read only the document are unaffected by the difference. A criterion-level pin overrides this one: the E08 and E10 ledger entries carry their own, because both were rebuilt in fix round 2. The commit `629bb51cda73fc70422769988b8c28182c9bd2d5` also serves as the comparison base of E10; there it is a base, not a measurement tree. Later card commits add tracked files under `.moai/reports/t1760/` (observed with `git ls-files .moai/reports/t1760`), so the difference between that base and a later `HEAD` is wider than these four files; E10 excludes `.moai/reports/` for that reason (AC-CMM-010).

**Carrier.** The four RED-now elements (command, verbatim stdout, exit code, tree SHA) live in the evidence ledger in § Evidence ledger, because a table cell mangles pipes and quotes. Each criterion cites its ledger entries by id.

**Reading the ledger fields.** Each entry has separate `stdout`, `stderr`, and `exit` fields. For a command run on the absent document the grep family prints nothing on stdout and a missing-file warning on stderr (in this tree `grep` resolves to a shell function over ugrep), awk prints `can't open file` on stderr, and both exit 2; `test -f` exits 1 with no output. For an absence criterion (E08, E09) green is stdout `0` with the document present; `grep` exits 1 on a zero count, and exit 2 means the file is missing, which is red. The observation wrapper used to split the streams (`<command> 2><scratch file>; echo "exit=$?"; cat <scratch file>`) is not part of the cited command.

## Criteria

### AC-CMM-001 — The document exists
- **Given** the run phase has finished, **When** the existence command in E01 runs, **Then** it exits 0.
- Release-blocking. RED reason: the document does not exist yet. Green path: M2 creates it. Ledger: E01.

### AC-CMM-002 — The reference-only statement is present
- **Given** the document exists, **When** E02 counts occurrences of the literal token `reference-only`, **Then** stdout is a number of 1 or more.
- Release-blocking. RED reason: the document does not exist. Green path: M2. Ledger: E02.

### AC-CMM-003 — The source register has exactly four rows, each verified-or-flagged
- **Given** the document exists, **When** E03 reads every row whose first cell is BP-1, BP-2, BP-5, or BP-6, **Then** it exits 0 and prints nothing: each of the four ids is the first cell of exactly one row, and every such row carries a double-quoted fragment of at least 20 characters and ends with `| VERIFIED |` or `| UNVERIFIED |`.
- Release-blocking. RED reason: the document does not exist (exit 2). Mutants this closes: four rows of one source (a plain row count reads 4), a fragment shorter than 20 characters, a flag other than the two tokens. Not closed by a command: a well-formed fragment that is not on the page, or a wrong URL; both are reviewer-checked under AC-CMM-011. Green path: M2 writes the register, M1's flags fill it. Ledger: E03.

### AC-CMM-004 — Every URL line carries a flag
- **Given** the document exists, **When** E04a counts lines containing `https://` and E04b counts lines containing `https://` that end with a flag cell, **Then** the two numbers are equal and each is 4 or more.
- Release-blocking. RED reason: the document does not exist. Mutant that this closes: a URL pasted into prose without a flag raises E04a above E04b. Green path: M2. Ledger: E04a, E04b.

### AC-CMM-005 — The mapping table covers the three themes with well-formed rows
- **Given** the document exists, **When** E05 scans the mapping table, **Then** it exits 0 and prints nothing: the header row `| Theme | Source | Source claim | MoAI defense | Fit |` appears exactly once, every theme row ends in `aligned`, `differs`, or `gap`, every non-gap row's MoAI-defense cell (the fifth `|`-separated field) contains a `.claude/…md` or `.moai/…md` path followed by `§` (the path may sit in backticks), and each of the three themes has at least one row.
- Release-blocking. RED reason: the document does not exist (exit 2). Mutants this closes: a `§` only in the claim cell, a defense cell with a path but no section marker, swapped column order, a Fit outside the three tokens, a missing theme. Green path: M3. Ledger: E05.

### AC-CMM-006 — Every claim with no located defense is a gap row
- **Given** the document exists and the run-phase evidence records G, the number of seed gap claims (`plan.md` § D carries two) for which no defense was located, **When** E06 counts rows ending in `| gap |`, **Then** stdout is a number of G or more. A seed gap claim for which the author did locate a defense is carried as a non-gap row and listed in the run-phase evidence with the file and section opened; G is then smaller, down to 0, and E06 printing `0` with exit 1 is conforming in that case.
- Release-blocking. RED reason: the document does not exist (exit 2). This is a floor tied to what the author found, not a demand for a gap: REQ-CMM-008 permits a gap row only where no defense is found, so an unconditional minimum would force either an unsupported gap or a contradiction of the requirement. Not closed by a command: a defense recorded in the evidence that was not actually opened, or a G recorded too low; a reviewer compares the recorded G with the seed claims (reviewer procedure of AC-CMM-014). Green path: M3. Ledger: E06.

### AC-CMM-007 — Every cited path resolves in full, and enough distinct paths are cited
- **Given** the document exists, **When** E07 extracts every token that starts `.claude/` or `.moai/` and contains `.md`, **Then** it exits 0 and prints nothing: no token continues past `.md` (a `.md.NONEXISTENT` suffix prints `BAD-SUFFIX`), every distinct path opens (`MISSING` otherwise), and there are at least eight distinct paths (`TOO-FEW-DISTINCT` otherwise).
- Release-blocking. RED reason: the document does not exist (exit 2). The distinct-path floor exists because a sweep that selected nothing asserts nothing, and because eight citations of one path would otherwise satisfy a count. Existence is tested by reading the path with awk's `getline < path`, which returns -1 for a path that cannot be opened and runs no command; a `system("test -f …")` form is refused by the worktree guard. Limits: a path that exists as an untracked file passes, and a cited directory or a non-`.md` path (for example a `.yaml` file) is not counted, so only `.md` citations are checked. A renamed section is caught by M3 re-reading, not by this command. Green path: M3. Ledger: E07.

### AC-CMM-008 — No identifiers, years, dates, or the word `develop`
- **Given** the document exists, **When** E08 counts lines containing a card identifier shape (`t` or `T` and three or four digits), a SPEC identifier (`SPEC-` and an uppercase token), a four-digit year starting 19 or 20, a calendar date, or the word `develop` in any letter case, each matched at explicit ASCII boundaries, **Then** stdout is `0`. A calendar date is any of: a zero-padded month-day pair such as `10-16` or `10/16`; an English month name, full or abbreviated, in either order with a one- or two-digit day, in any letter case except that lowercase `may` is not a month (`October 16`, `oct 16`, `OCTOBER 16`, `16 October`, `Sept. 9`, `May 4`), with an optional ordinal suffix on the day (`October 16th`, `16th October`); or a Korean month such as `3월`. The shapes the command does not cover are listed under Known limit.
- Release-blocking. RED reason: the document does not exist (exit 2). The boundaries are the explicit class `[^A-Za-z0-9_]`, not `grep -w`: the document body is Korean, and `-w` treats a Hangul particle as part of the word, which hides `2026년`, `t1760은`, and `develop를`. Month names are spelled with bracket classes (`[Oo][Cc][Tt]` and so on) so that the case-insensitivity covers only them, the card shape, and `develop`; the `SPEC-` branch stays case-sensitive so that `spec-anchored` does not count. Lowercase `may` stays out because it is an ordinary verb (`may 4`); `May` and `MAY` count. Green path: M2 to M4. Ledger: E08.

### AC-CMM-009 — No rule-marker tokens
- **Given** the document exists, **When** E09 counts lines containing `[HARD]` or `[ZONE`, **Then** stdout is `0`. This is the mechanical proxy for "no new `[HARD]` clause and no restated rule text".
- Release-blocking. RED reason: the document does not exist (exit 2). Green path: M2 to M4. Ledger: E09.

### AC-CMM-010 — Nothing outside the document and the SPEC directory changed
- **Given** the pinned base commit `629bb51cda73fc70422769988b8c28182c9bd2d5`, **When** E10a compares that commit with the working tree excluding the document, this SPEC directory, and the card-evidence directory `.moai/reports/`, and E10b lists uncommitted and untracked paths outside the same three, **Then** E10a exits 0 and E10b prints nothing.
- **Regression-guard, not release-blocking.** A scope guard is green on the pre-implementation tree by construction, so no RED-now observation exists for it; its red was instead observed on constructed probes (ledger E10a/E10b probes). E10a compares the pinned commit with the working tree, so it sees committed, staged, and unstaged tracked changes; it cannot see untracked files, which is why E10b exists. `.moai/reports/` is excluded from both commands by pathspec, on purpose. Card evidence is written there during the plan and run phases; `.gitignore` ignores the directory (observed `.gitignore:235:.moai/reports/*`), but an ignore rule hides only untracked files, and evidence that was force-added is tracked. In this tree two files under `.moai/reports/t1760/` are tracked (observed `git ls-files .moai/reports/t1760`), which made the earlier form of E10a exit 1 at the tip whatever the run phase did: a red that no correct work could clear. The cost of the exclusion is stated rather than hidden: neither command sees a change under `.moai/reports/`, so the criterion says nothing about evidence files; the reader of the evidence owns that. Evaluated at the run-phase tip (see Definition of Done). Ledger: E10a, E10b.

## Reviewer-checked criteria (no mechanical command)

These requirements have no discriminator that is sound on structure alone, and a vocabulary check would be unsound in both directions. Each is a criterion with a reviewer procedure. The reviewer records the result; none is recorded as a mechanical pass.

### AC-CMM-011 — Every VERIFIED flag has an authoring-run record (REQ-CMM-004)
- **Given** the document carries a source register, **When** a reviewer reads the run-phase evidence in `.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001/progress.md` § E.2, **Then** each source flagged `VERIFIED` has there, from the authoring run, the fetch and the raw-byte fragment match output (command and verbatim output per quoted fragment).
- Reviewer-checked. This is also where a fragment that is well formed but absent from the page, or a wrong URL, is caught.

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

Measured on tree `cf5a52f55992e59fe9593fa508ace5d4dba59fc7` (clean working tree), with the document absent. An earlier revision of this ledger put the tool's stderr text in a field named `stdout`; that was a mislabel, and the entries below carry the three streams separately.

```text
E01
command: test -f .moai/docs/multi-agent-containment-mapping.md
stdout:
stderr:
exit: 1
```

```text
E02
command: grep -c -i -F 'reference-only' .moai/docs/multi-agent-containment-mapping.md
stdout:
stderr: ugrep: warning: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E03
command: awk -F'|' -v rowre='^\\| BP-(1|2|5|6) \\|' -v fragre='"[^"]{20,}"' -v flagre='\\| (VERIFIED|UNVERIFIED) \\|$' '$0 ~ rowre { id=$2; gsub(/ /,"",id); c[id]++; if ($0 !~ fragre) { print "BAD-FRAGMENT: " $0; bad=1 } if ($0 !~ flagre) { print "BAD-FLAG: " $0; bad=1 } } END { if (c["BP-1"]!=1) { print "COUNT BP-1=" c["BP-1"]+0; bad=1 } if (c["BP-2"]!=1) { print "COUNT BP-2=" c["BP-2"]+0; bad=1 } if (c["BP-5"]!=1) { print "COUNT BP-5=" c["BP-5"]+0; bad=1 } if (c["BP-6"]!=1) { print "COUNT BP-6=" c["BP-6"]+0; bad=1 } exit bad }' .moai/docs/multi-agent-containment-mapping.md
stdout:
stderr: awk: can't open file .moai/docs/multi-agent-containment-mapping.md
 source line number 1
exit: 2
```

```text
E04a
command: grep -c -F 'https://' .moai/docs/multi-agent-containment-mapping.md
stdout:
stderr: ugrep: warning: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E04b
command: grep -c -E 'https://.*\| (VERIFIED|UNVERIFIED) \|$' .moai/docs/multi-agent-containment-mapping.md
stdout:
stderr: ugrep: warning: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E05
command: awk -F'|' -v rows='^\\| (correlated-failure|isolation-first|merge-queue-alternative) \\|' -v hdrre='^\\| Theme \\| Source \\| Source claim \\| MoAI defense \\| Fit \\|$' -v pathre='\\.(claude|moai)/[A-Za-z0-9_./-]+\\.md`? § ' '$0 ~ hdrre { hdr++ } $0 ~ rows { k=$2; gsub(/ /,"",k); t[k]++; f=$(NF-1); gsub(/ /,"",f); if (f!="aligned" && f!="differs" && f!="gap") { print "BAD-FIT: " $0; bad=1 } if (f!="gap" && !match($5, pathre)) { print "NO-PATH-SECTION: " $0; bad=1 } } END { if (hdr!=1) { print "HEADER-COUNT=" hdr+0; bad=1 } if (t["correlated-failure"]<1) { print "THEME-MISSING: correlated-failure"; bad=1 } if (t["isolation-first"]<1) { print "THEME-MISSING: isolation-first"; bad=1 } if (t["merge-queue-alternative"]<1) { print "THEME-MISSING: merge-queue-alternative"; bad=1 } exit bad }' .moai/docs/multi-agent-containment-mapping.md
stdout:
stderr: awk: can't open file .moai/docs/multi-agent-containment-mapping.md
 source line number 1
exit: 2
```

```text
E06
command: grep -c -E '\| gap \|$' .moai/docs/multi-agent-containment-mapping.md
stdout:
stderr: ugrep: warning: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E07
command: awk -v re='\\.(claude|moai)/[A-Za-z0-9_./-]+' '{ s=$0; while (match(s, re)) { t=substr(s,RSTART,RLENGTH); s=substr(s,RSTART+RLENGTH); sub(/\.+$/,"",t); if (t !~ /\.md/) continue; if (t !~ /\.md$/) { print "BAD-SUFFIX: " t; bad=1; continue } if (!(t in seen)) { seen[t]=1; n++; if ((getline junk < t) < 0) { print "MISSING: " t; bad=1 } else close(t) } } } END { if (n<8) { print "TOO-FEW-DISTINCT n=" n; bad=1 } exit bad }' .moai/docs/multi-agent-containment-mapping.md
stdout:
stderr: awk: can't open file .moai/docs/multi-agent-containment-mapping.md
 source line number 1
exit: 2
```

```text
E08
command: grep -c -E '(^|[^A-Za-z0-9_])(([Tt][0-9]{3,4})|SPEC-[A-Z][A-Z0-9-]*|(19|20)[0-9]{2}|[Dd][Ee][Vv][Ee][Ll][Oo][Pp]|(0[1-9]|1[0-2])[-/](0[1-9]|[12][0-9]|3[01])|([Jj][Aa][Nn]([Uu][Aa][Rr][Yy])?|[Ff][Ee][Bb]([Rr][Uu][Aa][Rr][Yy])?|[Mm][Aa][Rr]([Cc][Hh])?|[Aa][Pp][Rr]([Ii][Ll])?|May|MAY|[Jj][Uu][Nn][Ee]?|[Jj][Uu][Ll][Yy]?|[Aa][Uu][Gg]([Uu][Ss][Tt])?|[Ss][Ee][Pp]([Tt]([Ee][Mm][Bb][Ee][Rr])?)?|[Oo][Cc][Tt]([Oo][Bb][Ee][Rr])?|[Nn][Oo][Vv]([Ee][Mm][Bb][Ee][Rr])?|[Dd][Ee][Cc]([Ee][Mm][Bb][Ee][Rr])?)[ .]+[0-9]{1,2}([Ss][Tt]|[Nn][Dd]|[Rr][Dd]|[Tt][Hh])?|[0-9]{1,2}([Ss][Tt]|[Nn][Dd]|[Rr][Dd]|[Tt][Hh])?[ .]+([Jj][Aa][Nn]([Uu][Aa][Rr][Yy])?|[Ff][Ee][Bb]([Rr][Uu][Aa][Rr][Yy])?|[Mm][Aa][Rr]([Cc][Hh])?|[Aa][Pp][Rr]([Ii][Ll])?|May|MAY|[Jj][Uu][Nn][Ee]?|[Jj][Uu][Ll][Yy]?|[Aa][Uu][Gg]([Uu][Ss][Tt])?|[Ss][Ee][Pp]([Tt]([Ee][Mm][Bb][Ee][Rr])?)?|[Oo][Cc][Tt]([Oo][Bb][Ee][Rr])?|[Nn][Oo][Vv]([Ee][Mm][Bb][Ee][Rr])?|[Dd][Ee][Cc]([Ee][Mm][Bb][Ee][Rr])?)|[0-9]{1,2}월)([^A-Za-z0-9_]|$)' .moai/docs/multi-agent-containment-mapping.md
tree: 1b2319571070c4d3b60d270037d2f90f7d7ffa8d (criterion-level pin, overrides the document-level pin; fix round 2 observation, document absent at this tree)
stdout:
stderr: ugrep: warning: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E09
command: grep -c -F -e '[HARD]' -e '[ZONE' .moai/docs/multi-agent-containment-mapping.md
stdout:
stderr: ugrep: warning: .moai/docs/multi-agent-containment-mapping.md: No such file or directory
exit: 2
```

```text
E10a
command: git diff --quiet 629bb51cda73fc70422769988b8c28182c9bd2d5 -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports'
RED-now: not applicable (regression-guard; green on the pre-implementation tree by construction)
tree: 1b2319571070c4d3b60d270037d2f90f7d7ffa8d, working tree = that HEAD plus uncommitted edits confined to this SPEC directory (fix round 2 observation)
green (this tree):                                      exit 0
the fix-round-1 form, without ':!.moai/reports', same tree:  exit 1   (two tracked files under .moai/reports/t1760 differ from the base; this is the defect the exclusion removes)
probe, older pinned base 46960ac7d8043f1bee3f352688cceba8fb405738, this form:  exit 1   (changes outside the three allowed paths are still seen)
carried from fix round 1, tree cf5a52f55992e59fe9593fa508ace5d4dba59fc7, fix-round-1 form, NOT re-run in fix round 2:
  probe, untracked file added under .moai/docs:         exit 0   (E10a cannot see untracked files)
  probe, tracked file .moai/docs/agent-lint.md appended: exit 1
```

```text
E10b
command: git status --porcelain --untracked-files=all -- . ':!.moai/docs/multi-agent-containment-mapping.md' ':!.moai/specs/SPEC-CONTAINMENT-MAPPING-DOC-001' ':!.moai/reports'
RED-now: not applicable (regression-guard; green on the pre-implementation tree by construction)
tree: 1b2319571070c4d3b60d270037d2f90f7d7ffa8d, working tree = that HEAD plus uncommitted edits confined to this SPEC directory (fix round 2 observation)
green (this tree):                                      stdout empty, exit 0
probe, untracked file .moai/docs/zz-e10-probe.md added: stdout `?? .moai/docs/zz-e10-probe.md`, exit 0
after the probe:                                        the probe file was removed with rm; the unscoped `git status --porcelain --untracked-files=all` printed nothing, exit 0
carried from fix round 1, tree cf5a52f55992e59fe9593fa508ace5d4dba59fc7, fix-round-1 form, NOT re-run in fix round 2:
  probe, tracked file .moai/docs/agent-lint.md appended: stdout ` M .moai/docs/agent-lint.md`, exit 0
```

E10b's verdict is its stdout, not its exit code (git status exits 0 either way); that is why the green state is stated as "stdout empty". The earlier E10 probe `HEAD~1 HEAD` named moving refs and no longer reproduces (it exits 0 on this tree); the probes above use pinned commits and constructed edits instead. A change under `.moai/reports/` is invisible to both commands by design (AC-CMM-010); no probe was run against that directory because the exclusion is the intended behavior there.

## Guard measurements

The worktree-session guard decides which of the commands above can run in a card worktree. Observed in the fix round, in this session, quoting the guard's own words:

- The original E07 (`system("test -f " p)` inside awk) was refused: `this command runs awk with a program that can execute commands or that it cannot read in a plain command, so what it runs cannot be shown not to be git. Refusing to run it`.
- An awk program carrying `|` inside a regex literal next to `print` and `getline` was refused: `this command runs awk with a | after print/printf or before getline in its program, which awk can run as a pipe to a command, whatever the command is (a | in a regex, string or comment at such a spot is refused too: pass the pattern with -v re='a|b', doubling any backslash in it, and use $0 ~ re or gsub(re, "x"); print a literal | as "\174")`. The same pattern passed through `-v` was admitted, which is the form E03, E05, and E07 use.
- `awk --version` was refused with the same "cannot be shown not to be git" refusal, so no awk command above probes the awk build.
- The original E05 and the original E08, run verbatim on the absent document, were admitted (exit 2 each). E05 is rebuilt anyway because of D4.
- E01 to E10b were each run in this session and admitted (some with the observation wrapper around them).
- Fix round 2, in a worktree session: the revised E08 (no `awk`, one `grep -c -E` with the pattern in the command line) and the revised E10a and E10b (extra `':!.moai/reports'` pathspec) were each run against the fixtures and the tree named in their ledger entries and admitted. The E08 run on the absent document used a `2>` redirect of stderr to a scratch file as an observation wrapper, outside the cited command.

## Instrument validation (observed failure on constructed inputs)

A check is complete only after its failure has been observed on a known failing input. The fixtures live outside the repository (session scratch, not committed; recreate them from these descriptions): a passing fixture with a Korean body (four register rows, seven mapping rows across the three themes including one gap row, nine distinct existing paths cited as `path § section`, and the allowed look-alikes `v3.2.0은`, `30,000건`, `2,500건`, `developer를`, `development`, the first four with a Hangul particle attached); and the mutant fixtures named in the table.

| Command | Passing input | Failing inputs |
|---|---|---|
| E01 | not re-run in the fix round (unchanged; a fixture file exists, so it is not discriminating) | n/a |
| E02 | 1, exit 0 | plan run: 0, exit 1 (not re-run in the fix round) |
| E03 | exit 0, no output | four BP-1 rows only: exit 1, `COUNT BP-1=4`, `COUNT BP-2=0`, `COUNT BP-5=0`, `COUNT BP-6=0` (the earlier `grep -c` read 4 on this input and passed); two BP-1 rows and no BP-6: exit 1, `COUNT BP-1=2`, `COUNT BP-6=0`; a 9-character fragment and a flag `MAYBE`: exit 1, `BAD-FRAGMENT`, `BAD-FLAG` |
| E04a / E04b | 4 / 4, exit 0 | plan run: 2 / 0, unequal (not re-run in the fix round) |
| E05 | exit 0, no output | `§` only in the claim cell, vague defense: exit 1, `NO-PATH-SECTION` (the plan audit measured the earlier check exiting 0 on a comparable input; the earlier command was not re-run here); path without a `§`: exit 1, `NO-PATH-SECTION`; swapped header columns: exit 1, `HEADER-COUNT=0`; Fit `maybe`: exit 1, `BAD-FIT`; one theme absent: exit 1, `THEME-MISSING: merge-queue-alternative` |
| E06 | 1, exit 0 (one gap row) | plan run: 0, exit 1 (not re-run in the fix round) |
| E07 | exit 0, no output, on the passing fixture (nine distinct paths) and on a punctuation fixture (nine paths, some followed by `를`, `.`, `)`, `,`, `;`, or a closing backtick, plus a cited directory and a cited `.yaml` that must not count) | one path cited eight times: exit 1, `TOO-FEW-DISTINCT n=1` (the plan audit reported the earlier command counting 8 occurrences and passing; it is refused by the guard and cannot be re-run here); a real `.md` followed by `.NONEXISTENT`: exit 1, `BAD-SUFFIX`; eight real paths plus one nonexistent `.md`: exit 1, `MISSING`; seven distinct paths: exit 1, `TOO-FEW-DISTINCT n=7` |
| E08 | 0 on the passing fixture; 0 on the 26-line must-stay-zero fixture and on 21 further must-stay-zero lines (see below) | 24 of 24 lines count on the must-hit fixture; 18 of 18 on the new date-shape lines; a three-line fixture of exactly `october 16`, `OCTOBER 16`, `16 October`: 3 of 3 (the fix-round-1 command counted 0 of those 3); a shortened copy of the passing fixture plus one `16 October` line: 1; seven uncovered shapes: 0 (listed under Known limit) |
| E09 | 0, exit 1 | plan run: 1, exit 0 (not re-run in the fix round) |
| E10a / E10b | green: exit 0 / stdout empty | probes in the ledger entries |

E08 must-hit lines, each counted (24 of 24 with the fix-round-2 command; fixture `e08_hit.txt`, 24 lines): `t1760은`, `2026년`, `develop를`, `SPEC-CONTAINMENT-MAPPING-DOC-001을`, bare `t1760`, `T1760`, `Develop`, `DEVELOP`, `1999`, `SPEC-AUTH-001`, `2026-10-16`, `10-16`, `10/16`, `Oct 16`, `October 16`, `10월 16일`, `3월`, `(t1760)`, `[t1234]`, `Sept. 9`, `Dec 5에`, `카드t1760`, `한글2026년`, `t100을`. The first-round `grep -c -w -i` command read 10 of the first 19 (round 1 measurement, not re-run): it missed `t1760은`, `2026년`, `develop를` (the three Hangul-attached cases), and the six date shapes it never covered (`10-16`, `10/16`, `Oct 16`, `October 16`, `10월 16일`, `3월`).

E08 new date-shape lines, each counted (18 of 18, fix round 2; fixture `e08_hit_new.txt`): `october 16`, `OCTOBER 16`, `16 October`, `16 Oct`, `16 oct를`, `OCT 16`, `October 16th`, `16th October`, `1st Dec`, `Dec. 5`, `sept 9`, `June 3`, `3 june`, `Sep. 30`, `mar 3`, `MAY 4`, `May 4`, `4 May`. The fix-round-1 command counted 4 of these 18, and 0 of the three-line fixture `october 16` / `OCTOBER 16` / `16 October`.

E08 must-stay-zero lines, none counted (the 26-line fixture `e08_zero.txt`, 0): `developer`, `developer를`, `development`, `30,000건`, `2,500건`, `v3.2.0`, `v3.2.0은`, `93%`, `t99`, `spec-anchored`, `Opus 5.5`, `opus-5-5`, `3-5개`, `a1760`, `bt1760`, `1234567`, `ID20260`, `Marketing 5`, `Mark 5`, `may 4`, a bare `Mar`, `§ 5.1`, `§ 3.`, the two quoted source figures (`"around 2,500 code changes merged per day"`, `"we merged over 30,000 per day"`), and `prompting-claude-opus-5-5`. The two volume figures the document will quote, and the version token, do not trip the check.

E08 further must-stay-zero lines, added in fix round 2 because the date branch now matches case-insensitively and in both orders (21 lines, 0; fixture `e08_zero_new.txt`): `5 Marketing`, `3 Mark`, `4 may`, `may 4`, `2 Augment`, `10 Decisions`, `1 Septic`, `Dec2Hex`, `Mayor 5`, `Junction 5`, `Augment 5`, `Novel 3`, `Decimal 4`, `Octal 7`, `Marching 5`, `Sepal 2`, `Janet 3`, `Aprilia 2`, `Julian 4`, `Februaryish 3`, `Opus 5.5 and 3 Marks`. The Korean passing fixture also reads 0, and a shortened copy of it (the same opening paragraph and one mapping row) with one `16 October` line added reads 1.

Known limit: the E08 `develop` token, the years, and the dates count lines, not occurrences; a failing line reports 1 however many tokens it holds. Date shapes the command does not cover, observed uncounted on a seven-line fixture (0 of 7, fix round 2): `Oct16`, `Oct-16`, `October, 16`, `10.16`, `the 16th of October`, `20261016`, and `October ` followed by full-width digits. Lowercase `may` is not treated as a month by design. HTML entities that spell a forbidden token were not tried. A reviewer reads the document for these shapes (the date half of REQ-CMM-010 is mechanical only for the shapes above).

## Definition of Done

- AC-CMM-001 through AC-CMM-009 green on the final document, each with its command and verbatim output in the run-phase evidence, against the tree HEAD named there.
- AC-CMM-010 green against the pinned base commit, evaluated at the run-phase tip: E10a exits 0 and E10b prints nothing before the sync phase begins.
- AC-CMM-011 through AC-CMM-014 recorded by a reviewer.
- Every source flag in the document equals the outcome of the authoring-run fetch and fragment match.
- At the run-phase tip no file outside the document, this SPEC directory, and the card-evidence directory `.moai/reports/` changed (E10a and E10b), and in no phase does a file under `internal/template/templates/` change. The sync phase may add a changelog entry and update this SPEC directory's own status and progress files; those fall outside AC-CMM-010, which REQ-CMM-001 binds to the run phase only.
