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

## AC-GCS-002 — Arm B: real branch-state commands still match (control)

**Given** an ordinary branch-state command carrying no comment at all,
**When** `matchBranchStateCommand` is called with it,
**Then** it returns `(<expected suffix>, true)`.

| Command string | Expected suffix |
|---|---|
| `git merge --ff-only develop` | `git merge` |
| `git switch main` | `git switch` |

Both rows are measured as matching in this tree today (HEAD `3dfae918a`), so this criterion is a
**preservation check**: it holds now and must keep holding. A run in which it flips to no-match is
the signature of a blunted guard, which is the failure Arm A cannot see.

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
**Then** it returns `("git switch", true)` — no comment was opened, so nothing was elided.

| Command string | Expected suffix |
|---|---|
| `git switch feat#123` | `git switch` |
| `git merge topic#7` | `git merge` |

This is the falsification for REQ-GCS-002. An implementation stripping from any `#` to end-of-line
would blind the guard on every branch-state command whose branch name carries a hash.

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

Stated explicitly so no requirement rests on an implicit reading.

| Requirement | Asserted by |
|---|---|
| REQ-GCS-001 (comments elided before matching) | AC-GCS-001 |
| REQ-GCS-002 (`#` at word start only) | AC-GCS-004 |
| REQ-GCS-003 (per-line bound) | AC-GCS-003 |
| REQ-GCS-004 (elide, not placeholder) | AC-GCS-005 |
| REQ-GCS-005 (Arm A — mutation detected) | AC-GCS-001 |
| REQ-GCS-006 (Arm B — guard not blunted) | AC-GCS-002, AC-GCS-003, AC-GCS-004, AC-GCS-006 |
| REQ-GCS-007 (elision runs last) | AC-GCS-006 |
| REQ-GCS-008 (test file name pinned) | every criterion AC-GCS-001 … AC-GCS-006, via the shared verification command — an anchored selector that fails loudly if the test names move |

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
