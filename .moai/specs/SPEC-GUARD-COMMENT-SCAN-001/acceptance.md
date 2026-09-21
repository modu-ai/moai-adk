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

Table (each row an independent sub-case):

| Case | Command string | Expected |
|---|---|---|
| whole-line comment naming `git merge` | `# align with git merge --ff-only develop` | no match |
| trailing comment after a harmless command | `moai todo list  # then git switch main` | no match |
| comment after a separator | `ls ; # git reset --hard HEAD` | no match |

**RED-now** (pre-implementation state, measured at HEAD `3dfae918a`): row 1 currently returns
`(suffix="git merge", true)`. That is the defect; the criterion flips it.

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

| # | Command string | Expected suffix | Character preceding the `#` |
|---|---|---|---|
| 1 | `git switch feat#123` | `git switch` | `t` — alphanumeric |
| 2 | `git merge topic#7` | `git merge` | `c` — alphanumeric |
| 3 | `v=bar/#x ; git switch main` | `git switch` | `/` — **non**-alphanumeric |
| 4 | `git merge --ff-only rel-1.0#rc2` | `git merge` | `.` — **non**-alphanumeric |

This is the falsification for REQ-GCS-002. An implementation stripping from any `#` to end-of-line
would blind the guard on every branch-state command whose branch name carries a hash.

**Rows 3-4 close the second way this criterion can be passed while still being wrong.** Rows 1-2
exercise only alphanumeric preceding characters, so an implementation reading the word-start rule
as "the preceding character is non-alphanumeric ⇒ comment" passes them and is still blind on a
path, an assignment, or a dotted version in an operand. The verdict this SPEC is repairing built
that implementation and it passed all eleven pre-repair rows while missing row 3
(`.moai/reports/t1056/verdict.md` E2, candidate B).

**Row 3 additionally puts a real branch-state command later on the same line**, which is what makes
it catch the blinding rather than merely disagreeing about a token: if a comment is wrongly opened
at `/#`, elision runs to end-of-line and `git switch main` leaves the scan entirely. That the shell
genuinely runs both statements is measured in spec.md §B.1 (`bash -c 'echo bar/#x ; echo
SECOND-CMD-RAN'` → `bar/#x` then `SECOND-CMD-RAN`, HEAD `5bc42a304`, this run), alongside the
positive control showing the same probe does suppress a word when a comment really opens.

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

Stated explicitly so no requirement rests on an implicit reading. **Re-derived from the criteria as
they now read**, after the rows added to AC-GCS-002 and AC-GCS-004 — not edited in place. Asserting
a coverage the criteria do not actually deliver is worse than asserting none, because it stops the
next reader from checking.

| Requirement | Asserted by | Which rows carry it |
|---|---|---|
| REQ-GCS-001 (comments elided before matching) | AC-GCS-001 | all three rows |
| REQ-GCS-002 (`#` opens a comment at word start **only**) | AC-GCS-001 (positive half), AC-GCS-004 (negative half) | AC-GCS-001 rows 1-3 show line-start, whitespace-preceded and `;`-preceded `#` **do** open one; AC-GCS-004 rows 1-4 show alphanumeric- and non-alphanumeric-preceded mid-word `#` do **not** |
| REQ-GCS-003 (bounded to the physical line) | AC-GCS-003 | the two-line case |
| REQ-GCS-004 (comment run = `#` → end of line; preceding text survives; elided, not an operand token) | AC-GCS-005 (elide vs operand token), AC-GCS-002 (run starts at the `#`) | AC-GCS-005's single case; AC-GCS-002 rows 3-4 |
| REQ-GCS-005 (Arm A — mutation detected) | AC-GCS-001 | all three rows |
| REQ-GCS-006 (Arm B — guard not blunted) | AC-GCS-002, AC-GCS-003, AC-GCS-004, AC-GCS-006 | every row of those four expects a non-empty suffix and `true` |
| REQ-GCS-007 (a `#` inside a quoted span or heredoc body opens no comment) | AC-GCS-006 | the quoted case + the heredoc companion row |
| REQ-GCS-008 (test file name pinned) | every criterion AC-GCS-001 … AC-GCS-006, via the shared verification command — an anchored selector that fails loudly if the test names move | n/a |

### What this set now excludes (the map's discriminating power, stated so it is checkable)

The map above claims coverage; this sub-section says what that coverage actually rules out, so a
reader can falsify the claim rather than take it. Two implementations that passed every criterion
in the pre-repair set — both constructed and measured in `.moai/reports/t1056/verdict.md` E2 — are
now excluded:

| Candidate | What it does | Now fails on |
|---|---|---|
| A — line discard | discards the whole line containing a `#` | AC-GCS-002 rows 3-4 (`git switch main  # move to main` → no match) |
| B — non-alphanumeric rule | treats any `#` whose preceding character is non-alphanumeric as opening a comment | AC-GCS-004 row 3 (`v=bar/#x ; git switch main` → no match) |

This is an exclusion claim, not a completeness claim: it establishes that these two
pass-but-wrong implementations no longer pass. It does **not** establish that no third exists — the
verdict makes the same reservation, and it is carried forward here rather than quietly dropped.

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
