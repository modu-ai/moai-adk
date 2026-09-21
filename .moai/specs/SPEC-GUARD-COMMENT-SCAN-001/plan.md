# Implementation Plan — SPEC-GUARD-COMMENT-SCAN-001

Card **t1056**. Worktree `.claude/worktrees/t1056`, branch `WT-guard-prose`, base HEAD `3dfae918a`.

Sections are ordered by **decision reversibility**: the choices most likely to change on review come
first (§A-§C), the mechanical steps last (§F). A reviewer reading top-down meets the load-bearing
decisions before the typing.

---

## §A Decision 1 (least reversible) — where comment elision sits in the pipeline

**Decided: LAST, outermost.**

```go
// before
scanned := substituteQuotedArguments(substituteHeredocBodies(command))
// after
scanned := substituteShellComments(substituteQuotedArguments(substituteHeredocBodies(command)))
```

**Why this is the least reversible decision.** It is the only one whose wrong answer *blinds the
guard* rather than merely annoying a user. Every other choice here errs toward under-matching (the
guard's documented fail-open direction); this one, taken the other way, silently removes real
commands from the scan.

**The argument.** A `#` inside a quoted argument is not a comment, and a `#` inside a heredoc body
is not a comment. Both exclusions are wanted, and running last obtains both **for free** rather than
by re-implementing quote and heredoc tracking inside the comment step:

- by the time `substituteShellComments` runs, a quoted `#` is already part of `" X "`;
- a heredoc-body `#` is likewise already `" X "`.

So the comment step never sees a `#` that it should have ignored, and needs no knowledge of quoting
at all.

**The counter-order is measurably wrong**, which is what makes this decidable rather than a matter
of taste. With comment elision first, `echo "text # more" ; git switch main` truncates at the quoted
`#`, and `git switch main` vanishes from the scan — a real branch-state command un-guarded by a
quoted hash. That case is pinned as AC-GCS-006 so the ordering is asserted by consequence rather
than by reading call order.

**Accepted residual, recorded rather than discovered later.** The quoted-span placeholder is
`" X "`, which introduces a space. A construct like `foo"bar"#baz` becomes `foo X #baz`, where the
`#` now *looks* word-initial and is elided though the original was mid-word. The direction is
under-match (less text scanned → allow), consistent with the guard's documented fail-open posture
and with the residual `substituteQuotedArguments` already accepts for `bash -c "git switch main"`.
No branch-state command has this shape. Recorded as accepted, not fixed.

---

## §B Decision 2 — elide, or substitute the operand placeholder?

**Decided: elide (replace the comment run with nothing beyond the whitespace already preceding it).**

The existing discipline substitutes rather than deletes, and the SPEC honours the discipline's
*reason* rather than its *letter*:

| | quoted span | shell comment |
|---|---|---|
| What the shell does with it | passes it to the command as an argument | removes it before execution |
| Faithful model | a non-flag operand — `" X "` | absence |
| Fusion risk | real (arbitrary adjacent tokens) | **structurally impossible** — §B.1 puts a word boundary before the `#`, and end-of-line bounds the far side |

Using `" X "` here would state that the shell hands the command an extra operand, which it does not,
and the misstatement is observable: `git checkout -b # x` would then present a non-flag token after
`-b` and **match**, denying a command the shell itself would reject for having no branch name.
Eliding gives the shell's own answer. Pinned as AC-GCS-005 — the one criterion that discriminates
the two strategies.

---

## §C Decision 3 — the word-start separator set

**Decided:** `#` opens a comment when it is at line start, or immediately preceded by whitespace or
one of `;`, `&`, `|`, `(`.

This is the POSIX rule restricted to what a hook-scanned Bash command actually contains. Everything
outside the set fails **open** (no comment opened → more text scanned → the guard can still match),
which is the safe direction for this decision specifically — the inverse of §A, and the reason §C
sits below §A in this ordering.

Deliberately **not** included: backtick, `{`, and newline-escaped continuations. Each would widen the
elision surface (the unsafe direction) for constructs that do not appear in branch-state commands.

---

## §D Gap decision — hook-path reachability stays a Gap

The over-match is reproduced at the **matcher** level. Whether it reaches a user-visible deny
through the full hook path — the opt-in `Workflow.BranchGuard.Enabled`, primary-checkout
discrimination, and the exemption axes — was not measured.

**Decision: it stays a stated Gap and is NOT closed in this card.** Two reasons, both about scope
rather than effort:

1. `matchBranchStateCommand` is the pattern SSOT that `checkBranchState` consumes. The narrowing is
   correct at that layer whether or not the hook path currently surfaces it; closing the Gap would
   change no line of this SPEC's implementation.
2. Measuring it requires standing up the opt-in flag and a primary-checkout discriminant — a
   different axis with its own fixture, on a Tier S card whose whole diff is one preprocessing step
   and one test file.

**Coordinate for whoever closes it later**, so the Gap is actionable rather than merely admitted:
`internal/hook/pre_tool_branch_guard_integration_test.go` already drives the full path and is the
place an end-to-end comment case belongs. That is a separate card, not a milestone here.

---

## §E Evidence-path status — known untracked, stated before it is cited

The baseline evidence file is `.moai/reports/t1056/reproduction.md`. `.moai/reports/*` is gitignored
(`.gitignore:227`), so **the file is untracked and this worktree holds the only copy.**

Consequences accepted knowingly:

- Disposing of this worktree before the card's branch is merged destroys the baseline every figure
  in `spec.md` §A rests on. The worktree is not disposed until the remote merge lands.
- Run-phase evidence written under the same directory inherits the same status. Anything cited as
  the basis of a verdict is exported to a tracked path first; anything deliberately not exported is
  named in Residual-risk as a known loss rather than cited.

---

## §F Milestones

Priority-ordered; no time estimates.

### M1 (Priority High) — the preprocessing step

- Add `substituteShellComments(command string) string` to `internal/hook/branch_guard.go`, beside
  `substituteHeredocBodies`, carrying a doc comment in the same style: what the shell does, the
  measured false positive it closes, the bound it keeps, and the accepted residual from §A.
- Wire it into `matchBranchStateCommand` per §A. `branchStatePatterns` is not touched.
- Implements REQ-GCS-001 … REQ-GCS-004, REQ-GCS-007.

### M2 (Priority High) — the two-armed test

- New file `internal/hook/branch_guard_comment_test.go` (REQ-GCS-008), mirroring the shape of
  `branch_guard_heredoc_test.go`: a **pair** of tests, one data table plus one explicit falsification
  arm.
  - `TestBranchStatePatterns_ShellCommentIsNotACommand` — Arm A (AC-GCS-001).
  - `TestBranchStatePatterns_CommentCollapseDoesNotBlindTheGuard` — Arm B
    (AC-GCS-002 … AC-GCS-006), each case carrying its expected suffix so a blunted guard fails
    loudly rather than passing on a bare "no match".
- Both tests `t.Parallel()`, matching the package convention.
- Implements REQ-GCS-005, REQ-GCS-006.

### M3 (Priority Medium) — verification and evidence

- Run the anchored selector from `acceptance.md`; record the `--- PASS:` lines and the `=== RUN`
  count verbatim. A bare `ok` is not accepted as evidence (empty-sweep guard).
- `go test ./internal/hook/ -count=1` (package-scoped only — **never** `go test ./...` locally;
  CI owns the full-suite verdict).
- `gofmt -l internal/hook/`, `go vet ./internal/hook/...`.
- Record the Definition-of-Done checklist in `progress.md` §E.2 / §E.3, restating the §D Gap rather
  than dropping it.

---

## §G Anti-patterns for this card

- **Arm A alone.** "Comment prose no longer matches" is the exact output of a guard that matches
  nothing. Both arms, or neither.
- **Chasing E3.** The `cd <primary> && echo "git"` refusal is the Claude Code runtime's, measured by
  sentinel absence. Narrowing this guard to chase it weakens the guard and moves nothing.
- **Touching `branchStatePatterns`.** This card changes what string the patterns are scanned
  against, never the patterns. A diff hunk in that variable is out of scope by construction.
- **Stripping from any `#`.** The word-start rule is the whole difference between narrowing and
  blinding (AC-GCS-004).
- **Running comment elision first** because the pipeline reads left-to-right. §A is why.

---

## §H Convention deviation, recorded

Tier S convention omits `acceptance.md` and inlines acceptance criteria in `spec.md`. This SPEC
carries a separate `acceptance.md` at explicit dispatch instruction. Recorded here rather than
silently chosen; `spec.md` §D is a pointer to it, not a second copy, so there is one AC surface.

---

## §I Cross-references

- `internal/hook/branch_guard.go` — `substituteQuotedArguments` (:197), `substituteHeredocBodies`
  (:234), `matchBranchStateCommand` (:278), `quotedArgumentPlaceholder` (:179).
- `internal/hook/branch_guard_heredoc_test.go` — the two-test structural precedent M2 mirrors.
- `.moai/reports/t1056/reproduction.md` — baseline, untracked (§E).
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` § Mechanical Enforcement — the
  `BRANCH_GUARD_VIOLATION:` sentinel and the documented fail-open posture the residuals above appeal
  to.
