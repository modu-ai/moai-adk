# Acceptance Criteria — SPEC-GUARD-COMMENT-SCAN-001

## Preamble — why two arms, and why neither is sufficient alone

This card is a **narrowing** change: it makes the guard match less than it does today. A single arm
cannot distinguish the two outcomes it must separate.

- **Arm A alone** ("comment prose no longer matches") is satisfied by a correct narrowing **and** by
  a guard that has been blunted into matching nothing at all. It cannot tell them apart.
- **Arm B alone** ("real commands still match") is satisfied by the current, defective code, which
  does nothing about comments. It measures no change.

Both arms are therefore mandatory criteria, per REQ-GCS-005 and REQ-GCS-006. AC-GCS-001 is Arm A;
AC-GCS-002 through AC-GCS-006 are Arm B, each closing a distinct way the narrowing could overshoot.

**Before adding a row anywhere in this file, read the [HARD] row-discrimination rule in
AC-GCS-004.** It states what makes a row do any work, and it is written down because a row has
already been added to this file that carried the intended character and discriminated nothing. The
rule binds every table here, not only AC-GCS-004's.

**Selector discipline.** Every criterion below selects its test by exact anchored name. Run-phase
verification MUST record the `--- PASS:` / `--- FAIL:` line, not a bare `ok`: a `go test -run` whose
selector matches nothing prints `ok … [no tests to run]` and exits 0, which is an empty sweep, not
a pass.

Verification command shared by all criteria:

```
go test ./internal/hook/ -count=1 -v \
  -run '^TestBranchStatePatterns_(ShellCommentIsNotACommand|CommentCollapseDoesNotBlindTheGuard)$'
```

---

## AC-GCS-001 — Arm A: comment-borne git prose does not match

**Given** a Bash command string in which every occurrence of a branch-state git invocation sits
inside a shell comment,
**When** `matchBranchStateCommand` is called with that string,
**Then** it returns `(_, false)` — no pattern matches.

Table (each row an independent sub-case). Every row places the branch-state text **after** the `#`,
which is what gives an Arm A row its discriminating power — see the row-discrimination rule in
AC-GCS-004.

| # | Case | Command string | Character preceding the `#` | Expected |
|---|---|---|---|---|
| 1 | whole-line comment naming `git merge` | `# align with git merge --ff-only develop` | line start | no match |
| 2 | trailing comment after a harmless command | `moai todo list  # then git switch main` | space | no match |
| 3 | comment after a separator and a space | `ls ; # git reset --hard HEAD` | space | no match |
| 4 | comment opened by a separator with **no** space | `ls ;# git reset --hard HEAD` | `;` | no match |
| 5 | a second `#` inside an already-open comment run | `ls ; # prose about reset --hard HEAD per #123` | space (first `#`) | no match |

**Row 4 carries the separator half of REQ-GCS-002's word-start set**, which rows 1-3 do not: in all
three of those the character immediately preceding the `#` is whitespace or line start, so an
implementation whose word-start set is *only* whitespace and line start satisfies them and is still
wrong. Row 4 is the case where the two sets disagree. Measured in this tree at HEAD `9d822d826`,
this run — `;#` with no intervening space does open a comment in bash:

```
$ bash -c 'echo a;#echo SUPPRESSED
echo SECOND-CMD-RAN'
a
SECOND-CMD-RAN
```

`SUPPRESSED` is absent, so the comment opened. Negative control for the same probe shape — with no
comment, the tail word **does** print, so the absence above is a suppression rather than a probe
that never prints a second word:

```
$ bash -c 'echo a; echo NOT-SUPPRESSED
echo SECOND-CMD-RAN'
a
NOT-SUPPRESSED
SECOND-CMD-RAN
```

**Row 5 pins where the comment run STARTS** (REQ-GCS-004), which no other row does: an
implementation that elides from the *last* word-start `#` on the line rather than the first leaves
`ls ; # prose about reset --hard HEAD per ` in the scan and re-matches `git reset --hard` — the
exact false positive this card removes, in a shape (a comment citing an issue number) that is
common rather than exotic. bash opens no second comment there; the whole tail is one run. Measured
in this tree at HEAD `9d822d826`, this run:

```
$ bash -c 'echo X ; # prose about reset --hard HEAD per #123 echo SUPPRESSED
echo SECOND-CMD-RAN'
X
SECOND-CMD-RAN
```

(Same negative control as above applies: the tail prints when no comment opens.)

**RED-now** (pre-implementation state, measured at HEAD `3dfae918a`): row 1 currently returns
`(suffix="git merge", true)`. That is the defect; the criterion flips it. Rows 4-5 carry no
pre-implementation observation of their own — they were added during plan repair from bash
measurements, not from a matcher probe.

**Binary evidence**: the test's `--- PASS:` line for
`TestBranchStatePatterns_ShellCommentIsNotACommand`, with a non-zero `=== RUN` count.

---

## AC-GCS-002 — Arm B: real branch-state commands still match, comment or no comment

**Given** a real branch-state command — carrying no comment at all (rows 1-2), or carrying a
trailing comment on the **same line** (rows 3-4),
**When** `matchBranchStateCommand` is called with it,
**Then** it returns `(<expected suffix>, true)`.

| # | Command string | Expected suffix |
|---|---|---|
| 1 | `git merge --ff-only develop` | `git merge` |
| 2 | `git switch main` | `git switch` |
| 3 | `git merge --ff-only develop  # per lead instruction` | `git merge` |
| 4 | `git switch main  # move to main` | `git switch` |

**Rows 1-2 are the preservation check.** Both are measured as matching in this tree (HEAD
`3dfae918a`; `branch_guard.go` is byte-identical at HEAD `5bc42a304` — spec.md §G). A run in which
they flip to no-match is the signature of a blunted guard, which is the failure Arm A cannot see.

**Rows 3-4 fix the DIRECTION of the elision, which rows 1-2 cannot.** A comment and a command that
must survive sit on the *same line* here, so the criterion distinguishes "elide from the `#` to
end-of-line, preserving what precedes it" (REQ-GCS-004) from "discard the whole line containing a
`#`". The latter satisfies every criterion built only from comment-only lines and separate-line
pairs — AC-GCS-001's rows carry no surviving command, AC-GCS-003 puts the command on a *different*
line, AC-GCS-004's `#` opens no comment, and AC-GCS-006's `#` is gone before the comment step runs.
Without rows 3-4, nothing in the whole set observes text on the left of a real `#` surviving. The
verdict this SPEC is repairing constructed exactly that implementation and it passed all eleven
pre-repair rows while going blind on row 4 (`.moai/reports/t1056/verdict.md` E2, candidate A).

Rows 3-4 carry no pre-implementation observation: they are deliberately **not** written in the
manufactured-`#` shape (`echo "q"#b ; git switch main`), which a SPEC-conformant implementation
would fail by design — that case is an accepted blinding residual, spec.md §F.

---

## AC-GCS-003 — Arm B: per-line falsification

**Given** a two-line command string whose first line is a comment and whose second line is a real
branch-state command,
**When** `matchBranchStateCommand` is called with it,
**Then** it returns `("git switch", true)` — the elision stopped at the end of the comment line.

```
# this line is prose about git merge
git switch main
```

Without this pin, an implementation that discarded the whole command from the first `#` onward would
satisfy AC-GCS-001 vacuously while un-guarding every command that followed a comment (REQ-GCS-003).

---

## AC-GCS-004 — Arm B: word-boundary falsification

**Given** a real branch-state command whose operand contains a `#` that is **not** at word start,
**When** `matchBranchStateCommand` is called with it,
**Then** it returns `(<expected suffix>, true)` — no comment was opened, so nothing was elided.

| # | Command string | Expected suffix | Character preceding the `#` | Command after the `#`? |
|---|---|---|---|---|
| 1 | `git switch feat#123` | `git switch` | `t` — alphanumeric | no — documentary |
| 2 | `git merge topic#7` | `git merge` | `c` — alphanumeric | no — documentary |
| 3 | `v=bar/#x ; git switch main` | `git switch` | `/` — **non**-alphanumeric | yes |
| 4 | `v=rel-1.0#rc2 ; git switch main` | `git switch` | `.` — **non**-alphanumeric | yes |
| 5 | `v=a=#x ; git switch main` | `git switch` | `=` — **non**-alphanumeric | yes |
| 6 | `v=a-#x ; git switch main` | `git switch` | `-` — **non**-alphanumeric | yes |

This is the falsification for REQ-GCS-002. An implementation stripping from any `#` to end-of-line
would blind the guard on every branch-state command whose branch name carries a hash.

### [HARD] Row-discrimination rule — what makes a row in this table do any work

**The discriminating property of a row is NOT the character preceding the `#`. It is whether a
command that must survive sits AFTER the `#`, on the same line.**

A wrongly-opened comment elides from the `#` to end-of-line. So a row falsifies an implementation
only when that elision would remove something the row's expectation depends on. If the expected
suffix can still be produced from the text **left** of the `#`, every candidate produces it
whatever its word-start rule, and the row passes universally — however exotic the character it was
written to exercise. Such a row documents the intent and discriminates nothing.

This is not hypothetical, and the instance is inside this very table. Row 4 previously read
`git merge --ff-only rel-1.0#rc2`: it carried the `.` character the SPEC wanted pinned, but the
`git merge` sits **before** the hash, so eliding from `#` to end-of-line still leaves
`git merge --ff-only rel-1.0`, which matches. Ten candidate implementations were measured against
the then-current fifteen rows and *all ten* passed that row, the naivest one included
(`.moai/reports/t1056/verdict-iter2.md` E1 — cited, not re-measured here). Row 4 now carries the
same `.` in a shape where the surviving command is on the right of the hash.

**Therefore: any row added to this file — here, to AC-GCS-001, or to AC-GCS-002 — must place a
genuine branch-state command on the far side of the `#` from whatever the row's expectation
depends on.** Rows 1-2 above are retained deliberately and are marked documentary: they state the
ordinary alphanumeric case a reader expects to see, and they are not claimed to exclude anything.

### Why rows 3-6 exist, and why four of them

Rows 1-2 exercise only alphanumeric preceding characters, so an implementation reading the
word-start rule as "the preceding character is non-alphanumeric ⇒ comment" passes them and is still
blind on a path, an assignment, or a dotted version in an operand. The first plan-audit verdict
built that implementation and it passed all eleven pre-repair rows while missing row 3
(`.moai/reports/t1056/verdict.md` E2, candidate B).

Rows 3-6 are **one row per character that spec.md §B.1 measured as leaving the `#` mid-word** —
`/`, `.`, `=`, `-`. One falsifying row per character is what the set needs, because each widening
is independent: an implementation that adds only `.` to the word-start set is caught by row 4 and
by nothing else, and the same holds for `=` (row 5) and `-` (row 6). The second plan-audit verdict
constructed exactly those three widenings and measured each of them passing the whole pre-repair
set while blinding the guard (`.moai/reports/t1056/verdict-iter2.md` E2, candidates J / I / K —
cited, not re-measured here).

That the shell genuinely runs both statements on each of these lines is measured in this tree at
HEAD `9d822d826`, this run (`SECOND-CMD-RAN` stands in for the branch-state command, so no branch
state is mutated):

```
$ bash -c 'v=bar/#x ; echo "$v" ; echo SECOND-CMD-RAN'
bar/#x
SECOND-CMD-RAN
$ bash -c 'v=rel-1.0#rc2 ; echo "$v" ; echo SECOND-CMD-RAN'
rel-1.0#rc2
SECOND-CMD-RAN
$ bash -c 'v=a=#x ; echo "$v" ; echo SECOND-CMD-RAN'
a=#x
SECOND-CMD-RAN
$ bash -c 'v=a-#x ; echo "$v" ; echo SECOND-CMD-RAN'
a-#x
SECOND-CMD-RAN
```

The positive control for this probe shape — a `#` that genuinely opens a comment suppresses the
word after it — is recorded under AC-GCS-001 row 4 and in spec.md §B.1. Without it these four rows
would be consistent with a probe that never opens a comment at all.

---

## AC-GCS-005 — the comment is elided, not replaced by the operand placeholder

**Given** the command `git checkout -b # x`,
**When** `matchBranchStateCommand` is called with it,
**Then** it returns `(_, false)`.

This is the criterion that discriminates the two candidate collapse strategies (REQ-GCS-004,
spec.md §B.3). Replacing the comment run with `quotedArgumentPlaceholder` (`" X "`) would present a
non-flag operand after `-b` and make the `git checkout` pattern match; eliding leaves no operand and
it does not. Eliding is the behaviour that matches the shell: after comment removal that command is
`git checkout -b` with no branch name, which git rejects on its own.

---

## AC-GCS-006 — pipeline ordering: a quoted `#` opens no comment

**Given** the command `echo "text # more" ; git switch main`, in which the only `#` sits inside a
quoted span,
**When** `matchBranchStateCommand` is called with it,
**Then** it returns `("git switch", true)`.

This pins REQ-GCS-007 by its consequence rather than by inspecting call order. Comment elision
running **last** — after `substituteQuotedArguments` has already collapsed the quoted span to
`" X "` — leaves no `#` for the comment step to find, so the trailing `git switch main` survives
into the scan. The opposite order truncates the string at the quoted `#` and un-guards it, which is
the failure this criterion detects.

A companion row covers the heredoc side of the same ordering property:

| Command string | Expected suffix |
|---|---|
| `cat <<EOF > /tmp/note`<br>`# git merge --no-ff WT-card`<br>`EOF`<br>`git switch main` | `git switch` |

The `#` inside the heredoc body is already `" X "` by the time the comment step runs, and the
command after the terminator still matches.

---

## Requirement ↔ criterion map

Stated explicitly so no requirement rests on an implicit reading. Asserting a coverage the criteria
do not actually deliver is worse than asserting none, because it stops the next reader from
checking.

> **[HARD] This map has now been false twice, and both times for the same reason: rows were added
> to a criterion and the map was not re-derived from them.** The first plan-audit verdict found the
> REQ-GCS-006 cell claiming a coverage that did not exist; the second found the REQ-GCS-002 cell
> false in two separate ways after the first repair grew the tables
> (`.moai/reports/t1056/verdict.md` and `verdict-iter2.md`). **Whoever next adds or edits a row
> MUST re-derive every cell of this table from the criteria as they then read — not only the cells
> touching the edited criterion.** The cost of skipping it is invisible: no lint reads this table
> and no test is generated from it, so a false cell stays green forever.
>
> This edition was re-derived row by row from the nineteen rows as they now read, after the rows
> added to AC-GCS-001 and AC-GCS-004 in this repair.

| Requirement | Asserted by | Which rows carry it |
|---|---|---|
| REQ-GCS-001 (comments elided before matching) | AC-GCS-001 | all five rows |
| REQ-GCS-002 (`#` opens a comment at word start **only**) | AC-GCS-001 (positive half), AC-GCS-004 (negative half) | **Positive half** — AC-GCS-001 rows 1-4: line-start (r1), whitespace-preceded (r2, r3) and `;`-preceded-with-no-space (r4) `#` **do** open a comment. **Negative half** — AC-GCS-004 rows 3-6 only: `/` `.` `=` `-` preceded mid-word `#` do **not**. AC-GCS-004 rows 1-2 are documentary (see the row-discrimination rule) and carry no falsification. **Unfalsified residual**: `&`, `|` and `(` are named in REQ-GCS-002's word-start set and have **no row in this file** — an implementation that omits them from its set passes every criterion here. Direction is over-match, which spec.md §C declares the safe side |
| REQ-GCS-003 (elision bounded to the physical line) | AC-GCS-003 | the two-line case |
| REQ-GCS-004 (comment run = the `#` that **opened** the comment → end of line; preceding text survives; elided, not an operand token) | AC-GCS-005, AC-GCS-002, AC-GCS-001 | **Elide vs operand token** — AC-GCS-005's single case. **Preceding text survives** — AC-GCS-002 rows 3-4. **Run starts at the `#` that opened it** — AC-GCS-001 row 5 (a second `#` later in the same run does not restart it) |
| REQ-GCS-005 (Arm A — mutation detected) | AC-GCS-001 | all five rows |
| REQ-GCS-006 (Arm B — guard not blunted) | AC-GCS-002, AC-GCS-003, AC-GCS-004, AC-GCS-006 | every row of those four (4 + 1 + 6 + 2 = 13 rows) expects a non-empty suffix and `true` |
| REQ-GCS-007 (a `#` inside a quoted span or heredoc body opens no comment) | AC-GCS-006 | the quoted case + the heredoc companion row |
| REQ-GCS-008 (test file name pinned) | every criterion AC-GCS-001 … AC-GCS-006, via the shared verification command — an anchored selector that fails loudly if the test names move | n/a |

**Orphan and coverage check, run over this edition**: all nineteen rows are claimed by at least one
requirement above, and all eight requirements are claimed by at least one row or by the shared
selector. No orphan row, no uncovered requirement.

### What this set now excludes (the map's discriminating power, stated so it is checkable)

The map above claims coverage; this sub-section says what that coverage actually rules out, so a
reader can falsify the claim rather than take it. Nine implementations are known to have passed
some earlier edition of this set while disagreeing with the shell. Each was **constructed and
measured** by a plan-audit verdict — two in `.moai/reports/t1056/verdict.md` E2, seven more in
`verdict-iter2.md` E1/E2. The table states which row now excludes each.

| Candidate | What it does | Excluded by | Direction |
|---|---|---|---|
| A — line discard | discards the whole line containing a `#` | AC-GCS-002 rows 3-4 (`git switch main  # move to main` → no match) | blinding |
| B — non-alphanumeric rule | any `#` whose preceding character is non-alphanumeric opens a comment | AC-GCS-004 rows 3-6 | blinding |
| Z — any-hash | any `#`, wherever it sits, opens a comment | AC-GCS-004 rows 3-6 | blinding |
| J — set widened by `.` | REQ-GCS-002's set plus `.` | AC-GCS-004 row 4 | blinding |
| I — set widened by `=` | REQ-GCS-002's set plus `=` (and `-`, `:`) | AC-GCS-004 rows 5, 6 | blinding |
| K — set widened by `-` | REQ-GCS-002's set plus `-` | AC-GCS-004 row 6 | blinding |
| D — whitespace-only word start | set narrowed to whitespace and line start; separators dropped | AC-GCS-001 row 4 (`ls ;# …` → matches, criterion expects no match) | over-match |
| H — field-based | a comment only where a whitespace-separated field *begins* with `#` | AC-GCS-001 row 4 | over-match |
| N — last-hash | elides from the **last** word-start `#` on the line | AC-GCS-001 row 5 | over-match |

**[HARD] Status of this table: derived, not measured.** The candidate definitions and their
pass/fail behaviour on the *pre-repair* rows are measurements, cited above. The "Excluded by"
column is **my derivation** — what each cited definition does on the rows as they now read, worked
out by reading, not by running the candidates again. Run-phase MAY upgrade it by re-running the
ten candidates against the final nineteen rows; until then it is a reasoned claim, not an
observation, and a reader who finds a cell wrong should trust the row over the cell.

This is an exclusion claim, not a completeness claim: it establishes that these nine pass-but-wrong
implementations no longer pass. It does **not** establish that no tenth exists — both verdicts make
that reservation explicitly (the second one closed the first one's by *finding* three more), and it
is carried forward here rather than quietly dropped. The known live residuals are named above: the
`&` `|` `(` separators have no falsifying row, and spec.md §F carries three cases where the stated
rule knowingly differs from the shell.

## Definition of Done

- [ ] All six criteria verified with recorded commands and verbatim output, each showing a
      `--- PASS:` line and a non-zero `=== RUN` count.
- [ ] `go test ./internal/hook/ -count=1` passes whole-package — scoped to the changed package, not
      the full suite (CI owns the full-suite verdict).
- [ ] `gofmt -l internal/hook/` returns empty; `go vet ./internal/hook/...` exits 0.
- [ ] `git diff --stat` names exactly two files: `internal/hook/branch_guard.go` and
      `internal/hook/branch_guard_comment_test.go`. `branchStatePatterns` appears in no hunk.
- [ ] The Known Gap in spec.md §F (hook-path reachability unmeasured) is restated in the run-phase
      evidence rather than quietly dropped.
