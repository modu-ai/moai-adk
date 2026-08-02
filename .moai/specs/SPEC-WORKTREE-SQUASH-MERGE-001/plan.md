# Implementation Plan — SPEC-WORKTREE-SQUASH-MERGE-001

> Tier M. Milestones are ordered by decision-reversibility: the predicate's signal set and
> its interface shape come first because they are the decisions most likely to change under
> review; test scaffolding and documentation follow.

---

## §A Context

`IsBranchMerged` answers a reachability question (`git branch --merged`) while its two
callers need an equivalence question ("is this branch's work already in base?"). Under a
squash-merge workflow the two diverge permanently. Full root cause, the executed detector
matrix, and the four resolved scope decisions are in `spec.md` §1, §2, and §5.

---

## §B Tier Classification and Reasoning

**Assigned Tier: M** (artifact set: `spec.md` + `plan.md` + `acceptance.md`, plus
`progress.md`).

The mechanical thresholds alone would suggest Tier S: the change is one production
function in one file, well under 300 LOC, and — because the interface signature is
unchanged — it touches no test doubles and no call sites. Tier M is assigned on the
implementer-judgment clause of the Tier rule, for two reasons:

1. The function gates a **bulk directory-deletion** command. A false positive destroys
   unmerged user work with no undo. The behaviour therefore needs a separately reviewable
   acceptance matrix rather than criteria folded into `spec.md` §3.
2. The correctness argument rests on a seventeen-scenario × five-signal matrix, each cell of
   which must be independently verifiable. Inlining that into `spec.md` §3 would obscure
   both the requirements and the criteria.

### The AC budget is exceeded at v0.6.0 — declared, not absorbed

**This SPEC now carries 14 requirements and 17 acceptance criteria. The Tier M ceiling is 16
of each, so the criterion budget is over by one.** That is stated here rather than engineered
away, because both available ways of hitting 16 are worse than the overage:

- **Cramming.** The four new scenarios could be forced into existing criteria. SC-14 and
  SC-15 genuinely belong in AC-WSM-006 and are placed there — SC-15 is literally an
  under-inclusive path list, which is that criterion's original scope, and SC-14 widens it by
  one clause. But P3c and P4c are a different claim (the comparison must be mode-sensitive)
  whose falsification is a mechanism swap rather than a flag removal, and folding them in
  would leave AC-WSM-006 carrying eight scenarios and six falsifications. The ceiling exists
  because an over-budget SPEC "lands hardest on the plan-auditor, which must hold every
  requirement and criterion in view at once" — and a single criterion carrying six
  falsifications loads the auditor *more* than a seventeenth criterion does. Cramming defeats
  the ceiling's own purpose.
- **Manufacturing a consolidation.** Two criteria elsewhere could be merged to free a slot;
  AC-WSM-013 and AC-WSM-014 are the plausible pair, since the full-suite judge strictly
  contains the pre-existing-predicate-tests judge. But this SPEC committed at v0.4.0 that
  "no consolidation elsewhere freed a slot, and **none was manufactured to create one**", and
  the consolidations it has performed (AC-WSM-004, AC-WSM-005) each met a stated test: same
  requirement, shared falsification. AC-WSM-013 and AC-WSM-014 meet neither. Breaking that
  commitment to hit a number is the same class of move as the closure sentence this revision
  deletes — optimising the artifact's appearance against its own recorded standard.

**The correct disposition is a tier-up to L** (ceiling 25/25). The judgment clause that
carried this SPEC from Tier S to Tier M applies again unchanged in kind and larger in degree:
the correctness argument is now five failure axes, eight guards, seventeen scenarios, a
187-cell mutation grid, six audit iterations, and two commissioned empirical studies. The
implementation is still one function in one file, so the mechanical thresholds still say
Tier S — the same mismatch §B already resolves by judgment.

**The tier-up is deliberately NOT taken in this revision**, and the reason is scope rather
than disagreement. Tier L's artifact set is five files, adding `design.md` and `research.md`.
Producing those honestly means relocating the axis map, the guard inventory, and the
rejected-alternatives record out of `spec.md` §2 — a restructure of content that six audit
iterations have verified in place, on a revision whose stated worst outcome is a regression
in the SC-1..SC-13 rows. Creating two thin files that merely cross-reference §2 would satisfy
the count while duplicating the content, which is the over-formalization the Tier taxonomy
exists to prevent.

So the overage is carried openly for one iteration, and the tier decision is left to the
user with its cost named: **either** accept 17 criteria at Tier M as a declared exception,
**or** authorise a Tier L migration as a separate scoped change that moves `spec.md` §2's
design content into `design.md` and the two studies' findings into `research.md`. What is
not acceptable is silently exceeding the budget, and this note exists so that it is not
silent.

---

## §C Constraints

- **The interface does not change.** `IsBranchMerged(branch, base string) (bool, error)`
  keeps its exact signature. This is deliberate: it means no test double, no mock, and no
  call site needs editing, and both `--stale` and `--merged-only` pick up the fix at once
  (`spec.md` §5 decision 3).
- **Reachability must be retained, not replaced.** `git branch --merged` is the only signal
  that recognises a true merge commit and a strictly-behind branch. Replacing it with the
  patch-id probes would regress two scenarios that work today.
- **Fail-safe on error.** Any probe error propagates as an error; `staleKeepReason` turns
  that into a keep-and-report, never a removal.
- **No new dependency.** Plain `git` subprocesses through the existing `execGit` helper.
- **No repository-wide object reclamation.** `git prune` and `git gc` are prohibited
  (REQ-WSM-009).

---

## §D Technical Approach

`IsBranchMerged` becomes an ordered OR over the standalone signals S1 and S2, followed by
the two history signals S3 and S4, each of which is **conjoined with the state check S5**.
It short-circuits on the first term that fires. Ordering is by cost, cheapest first, so the
object-creating probe runs last.

```
IsBranchMerged(branch, base) -> (bool, error)

  S1  reachability     git branch --merged <base>  lists <branch>          -> merged
      (existing behaviour, including the "* "/"+ " marker stripping)

      mb <- git merge-base <base> <branch>       (computed once, shared by S2..S5)
      exit 1 with empty stdout = unrelated histories -> (false, nil), keep

  S2  empty diff       git diff --quiet <mb> <branch>   (rc 0 = merged, rc 1 = continue)

      names <- git diff --ignore-submodules=none --no-renames --name-only -z <mb> <branch>
              (computed once, used by S5)
              split the output on NUL, never on newline
              --no-renames is REQUIRED: without it a rename reports only its destination
              and the branch's deleted source path never enters the pathspec
              -z is REQUIRED: without it a path containing a backslash, tab, quote, or
              control character -- or, under the default core.quotePath=true, any
              non-ASCII byte -- is emitted C-quoted, and the quoted form matches nothing
              --ignore-submodules=none is REQUIRED HERE TOO, and is NOT redundant with the
              same flag on the comparison below: diff.ignoreSubmodules=all and
              submodule.<name>.ignore=all drop a changed gitlink from THIS enumeration, so
              without it the submodule pointer never reaches the comparison at all

  S5  state check      empty names                              -> holds
                       git --literal-pathspecs diff --quiet --no-textconv \
                           --ignore-submodules=none <branch> <base> -- <names as separate args>
                       rc 0 -> holds ; rc 1 -> withholds
                       NOT a merged verdict on its own
                       --literal-pathspecs is REQUIRED and is a GIT-LEVEL option, so it
                       precedes the subcommand: without it a byte-perfect path whose first
                       byte is ':' is parsed as pathspec magic and names no file, and
                       ':!'/':^' forms parse as EXCLUDE and invert what is compared
                       --no-textconv is REQUIRED and is a DIFF-level option, so it follows
                       the subcommand: without it a .gitattributes textconv driver makes
                       git compare the driver's RENDERING of each blob rather than the
                       blobs, and two genuinely different files report no difference
                       --ignore-submodules=none is REQUIRED and is DIFF-level: without it a
                       changed gitlink is ignored even when the path list names it

              flag placement is loud, not silent -- every misplacement exits 129, which is
              distinct from both verdict codes (measured):
                git --no-textconv diff ...              -> 129 unknown option
                git --ignore-submodules=none diff ...   -> 129 unknown option
                git diff --literal-pathspecs ...        -> 129 usage error

  S3  rebase-merge     git cherry <base> <branch>
                       -> merged only when output is non-empty
                          AND every line is prefixed '-'
                          AND S5 holds

  S4  squash-merge     synth <- git commit-tree <branch^{tree}> -p <mb> -m <fixed msg>
                       git cherry <base> <synth>
                       -> merged only when output is non-empty
                          AND every line is prefixed '-'
                          AND S5 holds

  otherwise -> not merged
```

Eight guards carry the whole safety argument. The first three constrain the signals; the
fourth and fifth constrain the *input* to a signal; the sixth constrains how that input is
*read back*; and the last two constrain what git *concludes* about an input that is already
correct — which is why each successive group was missed while only the previous one was
under review:

- **Non-empty output required (S3, S4).** Empty `git cherry` output means "no commits
  compared"; it must never be read as a merged verdict.
- **Every line prefixed `-` required (S3, S4).** A single `+` line means part of the branch's
  work is absent from base. This is what keeps the partially-applied case safe.
- **S5 must hold (S3, S4).** `git cherry` answers a history question — did an equivalent
  patch-id ever exist in base since the merge-base. S5 answers the state question — does the
  content survive in base HEAD. A branch whose patches were applied and then reverted
  satisfies the first and fails the second (`spec.md` §2 SC-8). S5 is a conjunct only: it
  can withhold a verdict, never grant one.
- **S5's path list must be complete (`--no-renames`).** A correct conjunct fed an
  incomplete path list returns `OK` and grants the verdict. `git diff --name-only` applies
  rename detection by default and reports only a rename's destination, so the path the
  branch deleted is never checked (`spec.md` §2 SC-10, SC-11).
- **S5's path list must round-trip (`-z`).** A correct conjunct fed a *corrupted* path list
  fails the same way as one fed an incomplete list — it returns `OK` and grants the verdict.
  `git diff --name-only` emits a C-quoted rendering rather than the path itself whenever the
  path contains a backslash, tab, double quote, or control character, and (under the default
  `core.quotePath=true`) whenever it contains any non-ASCII byte. The quoted rendering
  matches no file (`spec.md` §2 SC-12). This guard is independent of the previous one:
  `--no-renames` decides *which* paths are enumerated, `-z` decides *what form they arrive
  in*, and getting one right does not repair the other.
- **S5's comparison must match literally (`--literal-pathspecs`).** A correct conjunct fed a
  complete, byte-perfect path list *still* fails when git re-parses that list as pathspec
  expressions rather than filenames. Git reads a pathspec's leading characters as magic
  before matching, so a repo-root path whose first byte is `:` names no file, and `:!`/`:^`
  forms parse as EXCLUDE and invert the comparison (`spec.md` §2 SC-13). This guard is
  independent of the previous two: they govern the list's contents and its bytes, this one
  governs its interpretation, and a list that is perfect on both prior axes still fails here.
  The flag disables all pathspec magic and globbing categorically; losing glob expansion is
  free on the safety axis, since over-matching can only add differences and an added
  difference withholds a verdict.
- **S5's comparison must not be narrowed by diff configuration (`--no-textconv`).** The three
  guards above all govern the path list. This one governs the *answer*. `git diff --quiet`
  decides equality under the repository's diff configuration, and a `.gitattributes` rule
  binding a `textconv` driver makes it compare the driver's rendering rather than the blobs —
  so a complete, byte-perfect, literally-matched list is judged wrongly and the branch is
  reported merged (`spec.md` §2 SC-14). The trigger is an ordinary documented recipe
  (`*.pdf diff=pdf` is an example in git's own `gitattributes(5)`), and the predicate runs in
  whatever repository the user has. Note the exposure is read from the **worktree's**
  `.gitattributes`, not from the tree under judgement, so it depends on checkout state.
- **S5 must see a changed gitlink, at BOTH stages (`--ignore-submodules=none`, twice).**
  `diff.ignoreSubmodules=all` and `submodule.<name>.ignore=all` suppress a moved submodule
  pointer when the path list is built *and* when the comparison is made (`spec.md` §2 SC-15).
  Closing one stage leaves the other open, which is why this is the only guard applied twice
  and why the enumeration-side occurrence must not be deleted as duplication — measured,
  removing it from either stage alone reproduces the false positive. `submodule.<name>.ignore`
  is readable from a **tracked** `.gitmodules`, so this is the one exposure in the whole
  design that a repository can ship to every clone without any operator configuration.

Since S5 evaluates identically for S3 and S4, compute it once, after `names`, and reuse it.

### The comparison must stay mode-sensitive — a property to preserve, not to add

`git diff` compares file **modes** as well as content, and two scenarios in the matrix depend
on it. A symlink and a regular file whose content is exactly the symlink's target **share one
blob OID**, and a `chmod` changes no content at all:

```
P3c   feat 100755 blob 587be6b4…  x.txt      main 100644 blob 587be6b4…  x.txt
P4c   feat 100644 blob 1de56593…  link.txt   main 120000 blob 1de56593…  link.txt
```

The specified mechanism already reports both correctly, so no change is required — this is
recorded as a constraint on **future** edits. Any substitution of the comparison for
something that compares content hashes alone (a `git rev-parse <rev>:<path>` lookup, a
`git cat-file --batch-check` batch, or an `ls-tree` map that stores only the OID) turns P3c
and P4c into silent fail-open false positives under **default** git configuration — no
`.gitattributes`, no config key. That is a strictly worse exposure than any of the five axes,
all of which require non-default state. `git cat-file --batch-check` is additionally
disqualified: it cannot resolve a gitlink and reports `missing` for one on both sides at exit
0, so two different submodule pointers compare equal. AC-WSM-017 exists to fail if the
comparison is made mode-blind.

### Building and passing `names` — NUL-split, and a `[]string` never a joined string

Three separate obligations apply to the path list, and satisfying one does not satisfy the
others. **How the list is built** is governed below under "splitting the enumeration
output"; **how the list is handed to git** is governed by this subsection; **how git reads
it back** is governed by the `--literal-pathspecs` guard above. All three failures produce a
pathspec that matches nothing, and by `spec.md` REQ-WSM-007 that reads as merged.

The path list MUST reach the exec helper as a `[]string` argument list, one element per
path. It MUST NOT be joined into a single string and re-split, and the shell snippets in
`spec.md` / `acceptance.md` MUST NOT be transcribed literally into Go.

This is not style. `git diff --quiet` exits **0** for a pathspec that matches nothing
(`spec.md` REQ-WSM-007), so a malformed pathspec does not fail — it reads as "no
difference", which the predicate reads as *merged*. A joined list becomes one pathspec that
matches nothing, and the branch is reported merged. Measured, where the correct answer is
`rc=1`:

```
argv: <feat> <main> <--> <old.txt >   (joined into one argument)   rc=0   <- WRONG
argv: <feat> <main> <--> <old.txt>    (one argument per path)      rc=1   <- right
```

The joined form also behaves differently per shell — `bash` word-splits an unquoted
`$names` and survives by luck, `zsh` does not split and silently returns the wrong answer
— which is precisely why the list must never travel through a shell string. `execGit`
already takes variadic arguments, so the correct form is the natural one; the hazard is
only in copying a snippet.

Go's `exec.Command` passes arguments directly without shell interpretation, so the
argument-list half of the problem is solved simply by never building a joined string.

**No shell — a checked property, not an assumption.** The argv axis is closed *only* while
every git invocation in the predicate reaches the process directly. `execGit` must continue
to build its command with `exec.Command`/`exec.CommandContext` and an explicit argument
slice; it must not route any invocation through `sh -c`, `bash -c`, `exec.Command("sh", ...)`,
or any string-interpolated command line. Re-introducing a shell anywhere on this path
re-opens the argv axis in full, and it does so silently: a shell would word-split the path
list, and by the unmatched-pathspec property above the resulting mismatch reads as *merged*
rather than as an error.

This is stated as a constraint because the evidence base does not cover it. The measurements
throughout these artifacts were produced with direct `subprocess` argv and no shell, which
models the intended Go construction but does not verify it — no Go code existed when they
were taken. The run-phase implementer therefore verifies `execGit`'s construction directly
rather than inheriting the assumption:

```bash
grep -nE 'exec\.Command(Context)?\((ctx, )?"(sh|bash|zsh)"|"-c"' internal/core/git/*.go
```

Executed against the pre-implementation tree: **no match** (exit 1). The construction in
place today is `exec.CommandContext(ctx, gitPath, args...)` at `manager.go:257`, which is
already correct — so this check begins satisfied and exists to keep it that way.

The check is falsifiable, verified by injection rather than asserted. Appending
`func probeShell() { _ = exec.Command("sh", "-c", "git diff") }` to `manager.go` makes it
report `manager.go:303: ... exec.Command("sh", "-c", "git diff")` and exit 0; reverting the
file returns it to no-match. A guard that cannot be made to fire is not a guard.

### Splitting the enumeration output — on NUL, never on newline

The other half is **not** solved by `exec.Command`, and it is the half a previous version of
this plan got wrong. That version stated that a `[]string` built from
`strings.Split(out, "\n")` is *safe* provided empty trailing elements are dropped. **That
claim is false**, and it is false in the unsafe direction: it endorses the exact
construction that produces the SC-12 false positive.

Dropping empty elements is necessary but nowhere near sufficient. `git diff --name-only`
does not emit paths verbatim — it emits a **C-quoted rendering** for any path containing a
backslash, tab, double quote, or control character, and (under git's documented default
`core.quotePath=true`) for any non-ASCII path. Splitting that output on newline therefore
yields elements that are *not paths*: passing `"\353\254\270\354\204\234.txt"` back as a
pathspec matches no file, `git diff --quiet` exits 0, S5 holds vacuously, and the branch is
reported merged while its work is absent from base. Measured, where the correct answer is
`rc=1` in every row:

```
core.quotePath=true (git's documented default) — each --name-only line round-tripped
  "a\\b.txt"                      rc=0   <- matched NOTHING
  "caf\303\251.txt"               rc=0   <- matched NOTHING
  plainname.txt                   rc=1   (matched)
  "tab\tname.txt"                 rc=0   <- matched NOTHING
  "\355\225\234\352\270\200.txt"  rc=0   <- matched NOTHING

-z (NUL-delimited) — same five paths, round-trip correct
  a\b.txt   café.txt   plainname.txt   tab<TAB>name.txt   한글.txt      all rc=1
```

The required construction is therefore
`git diff --ignore-submodules=none --no-renames --name-only -z` split on `\x00` — in Go,
`bytes.Split(out, []byte{0})` (or `strings.Split(out, "\x00")`), dropping the empty trailing
element that a NUL terminator leaves behind. `strings.Split(out, "\n")` MUST NOT be used on
this command's output.

Do **not** substitute `core.quotePath=false` for `-z`. It suppresses quoting for non-ASCII
paths only; backslash and control-character paths stay quoted regardless of the setting —
measured in `acceptance.md` §C.2 route 3, second block, where `"a\\b.txt"` and
`"tab\tname.txt"` still fail to round-trip under `quotePath=false`. The config change
therefore closes part of the hole while reading as though it closed all of it.

Exit-code handling is not uniform across probes: `git diff --quiet` uses rc `1` as a
verdict and `git merge-base` uses rc `1` to mean unrelated histories. `spec.md` REQ-WSM-007
carries the measured per-probe table; treating any non-zero as a failure would make S3, S4,
and S5 unreachable.

The synthetic commit in S4 pins its identity so the object is deterministic
(REQ-WSM-008), which bounds the unreferenced-object count to the number of distinct branch
states rather than the number of sweeps:

```
GIT_AUTHOR_NAME / GIT_COMMITTER_NAME    fixed
GIT_AUTHOR_EMAIL / GIT_COMMITTER_EMAIL  fixed
GIT_AUTHOR_DATE / GIT_COMMITTER_DATE    fixed ("@0 +0000")
commit message                          fixed
```

`execGit` currently builds its environment as `os.Environ()` plus two fixed entries and
exposes no per-call environment hook, so M2 adds one — either a variant helper or an
optional parameter. Whichever shape is chosen must leave the existing `execGit` callers
untouched.

### Anchoring note

The `@MX:ANCHOR` block on `cleanStaleWorktrees` documents the two-condition gate; it stays
as-is. A new `@MX:ANCHOR` on the predicate should record why reachability is retained
alongside the patch-id probes, since that is the non-obvious invariant a future edit is
most likely to break.

---

## §E Milestones

Ordered most-reversible-decision-first.

### M1 — Predicate signal set and its guards (highest change likelihood)

The decision under review here is *which* signals compose the predicate and *how* their
guards are written. Implement `IsBranchMerged` as the composition above: S1 and S2
standalone, S3 and S4 each conjoined with S5, with the non-empty-output and all-lines-`-`
guards on S3 and S4, and merge-base plus `names` computed once. Retain the existing
`git branch --merged` path and its marker-stripping helper unchanged. Apply the per-probe
exit-code table from `spec.md` REQ-WSM-007 rather than treating non-zero as failure.

Exit: the seventeen-scenario matrix in `acceptance.md` §C reproduces the expected verdict for
every scenario, including SC-8 (partially reverted), which the history signals alone report
as merged; SC-10 / SC-11 (rename + re-add), which they report as merged whenever the path
enumeration is not rename-blind; SC-12 (non-ASCII path), which they report as merged
whenever the enumeration output is split on newline instead of NUL; SC-13 (leading-colon
path), which they report as merged whenever the comparison omits `--literal-pathspecs`, even
though its path list is byte-perfect; SC-14 (textconv driver) and SC-15 (submodule pointer
under an ignore directive), which they report as merged whenever the comparison is issued
under the repository's own diff configuration; and P3c / P4c (mode-only divergence and
symlink/file blob-OID collision), which the specified mechanism already handles and which
fail only if the comparison is made mode-blind.

### M2 — Deterministic synthetic-commit identity

Add the per-call environment mechanism `execGit` lacks, and pin the six identity variables
plus the commit message. Existing `execGit` callers must be unaffected.

Exit: two successive evaluations of one unchanged branch produce the same object hash, and
the second adds no new dangling object.

### M3 — Real-repository scenario tests

Add table-driven tests in `internal/core/git` using the existing `initTestRepo`, `runGit`,
and `writeTestFile` helpers, one case per scenario in the matrix. Mock-based CLI tests
cannot cover this: `clean_stale_test.go` stubs `IsBranchMerged` through
`mockIsBranchMergedFunc`, so the CLI layer never exercises the real predicate.

Each guard test must be falsifiable — removing the guard it protects has to make the test
fail. `acceptance.md` §D records the falsification procedure per criterion.

Exit: `go test ./internal/core/git/...` green, including the two pre-existing
`TestWorktreeIsBranchMerged*` tests, which encode scenarios the change must not alter.

### M4 — Regression and full-suite verification

Run the full suite and `go vet`. Confirm the `--stale` and `--merged-only` CLI tests still
pass unchanged, which is the evidence that the interface really did stay fixed.

Exit: `go test ./...` and `go vet ./...` both clean.

### M5 — Mechanical follow-through

Update the `clean` command's `Long` help text only if the observed behaviour no longer
matches its wording, and add the `@MX:ANCHOR` described in §D. No other files.

---

## §F Risks

| Risk | Severity | Mitigation |
|---|---|---|
| A patch-id false positive removes a worktree holding unmerged work | **Critical** | The guards on S3/S4; the seventeen-scenario matrix, in which all eleven must-keep scenarios were observed as keep under the composed predicate; the independent working-tree-cleanliness condition, which `--stale` retains |
| Replacing rather than extending the reachability signal regresses merge-commit and behind-branch cases | High | S1 retained verbatim; the two pre-existing `TestWorktreeIsBranchMerged*` tests cover exactly these and must keep passing |
| `git cherry` cost grows with the number of base commits since merge-base | Low | Real, and it applies to **both** S3 and S4 — see the measured note below. Bounded in practice at roughly a quarter-second per branch on this repository; no budget is imposed, and the sweep is a user-initiated maintenance command rather than an interactive path |
| Unreferenced objects accumulate | Low | Deterministic identity (M2) bounds them by distinct branch state; git auto-gc reclaims them; `git prune` is prohibited (REQ-WSM-009) |
| A branch merged by individual replay and then amended is not detected | Low | Conservative false negative — the worktree is kept. Recorded in `spec.md` §2 as an accepted limitation |
| A history signal fires on a branch whose work was reverted out of base | **Critical** | The S5 state conjunct on S3 and S4 (`spec.md` §2 SC-8). Mutation-checked: removing S5 alone flips SC-8 to merged |
| A history signal fires on a branch whose renamed-away path base still holds | **Critical** | `--no-renames` on the S5 path enumeration (`spec.md` §2 SC-10, SC-11). Mutation-checked: the `folded-names` mutation flips exactly SC-10 and SC-11 and nothing else. The conjunct alone does not catch this — it is fed an incomplete path list and holds vacuously |
| A history signal fires on a branch whose only absent path has a name git C-quotes | **Critical** | `-z` on the S5 enumeration plus a NUL split (`spec.md` §2 SC-12). Mutation-checked: the `newline-split` mutation flips exactly SC-12 and nothing else. Orthogonal to `--no-renames` — the rename fix left this route completely open. Note `core.quotePath=false` is **not** a mitigation: backslash and control-character paths quote regardless |
| A history signal fires on a branch whose only absent path is parsed as pathspec magic | **Critical** | `--literal-pathspecs` as a git-level option on the S5 comparison (`spec.md` §2 SC-13). Mutation-checked: the `no-literal` mutation flips exactly SC-13 and nothing else. Orthogonal to both `--no-renames` and `-z` — the path list is already complete and byte-perfect when this one fires, which is why the two prior repairs left it open |
| A history signal fires on a branch whose only absent path is compared under a `textconv` diff driver | **Critical** | `--no-textconv` on the S5 comparison (`spec.md` §2 SC-14). Mutation-checked: the `no-textconv` mutation flips exactly SC-14 and nothing else. Orthogonal to all three round-trip repairs — the path list is complete, byte-perfect, and literally matched when this one fires; only git's *judgement* of it is wrong. Trigger is an ordinary documented recipe, so the exposure is not exotic |
| A history signal fires on a branch whose only absent change is a moved submodule pointer | **Critical** | `--ignore-submodules=none` on **both** the S5 enumeration and the S5 comparison (`spec.md` §2 SC-15). Mutation-checked: `no-ignoresub-cmp` and `no-ignoresub-enum` each flip exactly SC-15 on their own, which is the evidence both occurrences are needed. Uniquely, `submodule.<name>.ignore=all` is readable from a tracked `.gitmodules`, so this exposure can arrive inside the repository with no operator configuration |
| A future edit swaps the S5 comparison for a mode-blind mechanism, and mode-only divergence is silently missed | **Critical** | `git diff` compares modes; P3c and P4c pin this (`spec.md` §2). Mutation-checked: an OID-only comparison flips exactly P3c and P4c. Worse than any axis member because it fails under **default** configuration — a symlink and a regular file with identical content share one blob OID. AC-WSM-017 is the criterion that fails if the swap is made; `plan.md` §D records why `rev-parse`, `cat-file --batch-check`, and OID-only `ls-tree` maps are all disqualified |
| The path list is joined into one shell string, matching nothing, and reads as merged | **Critical** | Pass `names` as a `[]string` to `execGit` (§D). The failure is silent: an unmatched pathspec exits 0, so this surfaces as a wrong removal rather than an error |
| A shell is re-introduced on the git-invocation path, silently re-opening the argv axis | **High** | `execGit` keeps its direct `exec.CommandContext` construction; the §D grep is the check, verified falsifiable by injection. Note the evidence base for every measurement in these artifacts used direct argv with no shell, so this is a constraint the implementation must hold rather than one the measurements established |

### Measured note — `git cherry` cost (corrects an earlier claim)

An earlier version of this plan asserted that the synthetic commit's merge-base parent
bounds the compared range "by the branch's own divergence, not by the repository's
history". That is **wrong**, and it was wrong for S4 as well as S3.

`git cherry <upstream> <head>` computes patch-ids on *both* sides: the head side
(`upstream..head`) and the upstream side (`head..upstream`). The synthetic parent shrinks
only the head side, to a single commit. The upstream side is unaffected and grows with
base's divergence since the merge-base. Measured on git 2.50.1 in a synthetic repository:

```
base commits since merge-base=50    S3 real=0.01s   S4 real=0.01s
base commits since merge-base=200   S3 real=0.13s   S4 real=0.12s
base commits since merge-base=800   S3 real=0.23s   S4 real=0.26s
   (S3 upstream-side commits = 800, head-side = 1)
```

S4 tracks S3 almost exactly — the synthetic parent buys no asymptotic saving. Measured on
this repository's real branches:

```
branch=backup-main-full-20260706   upstream-side=1143   git cherry elapsed=244ms
branch=backup-handoff-deps-fix     upstream-side=918    git cherry elapsed=225ms
branch=GoosLab/main-fork           upstream-side=87     git cherry elapsed=122ms
```

Practical bound for a full `--stale` sweep here: two `git cherry` calls per branch at
roughly 0.25s each, over the 45 worktree branches in `spec.md` §1, is on the order of 20
seconds. That is acceptable for a user-initiated maintenance command and no latency budget
is set, but the honest statement is that the cost scales with base's history, not with the
branch's. No acceptance criterion measures predicate latency; if a repository with a far
longer main makes the sweep unpleasant, bounding it is follow-up work, not part of this
change.

---

## §G Anti-Patterns

- Using plain `git cherry` as the squash detector. Falsified — see `spec.md` §2.
- Dropping `git branch --merged` once the patch-id probes are in place.
- Treating empty `git cherry` output as a merged verdict.
- Accepting a merged verdict when any `git cherry` line is prefixed `+`.
- Reporting merged on a patch-id history signal (S3 or S4) without the S5 state conjunct.
  This is the partially-reverted false positive; it is the one mutation that flips SC-8 on
  its own.
- Promoting S5 to a standalone merged signal. It is a conjunct only — it may withhold a
  verdict, never grant one.
- Enumerating S5's paths with `git diff --name-only` (rename detection on). This folds a
  rename's deleted source path out of the pathspec, so S5 holds vacuously and the rename +
  re-add false positive returns. Dropping `--no-renames` as "redundant" is the specific
  edit AC-WSM-006's first falsification exists to fail.
- Splitting the S5 enumeration's output on `\n` instead of `\x00`, or dropping `-z` as
  "noise". A quoted path is not the path; it matches nothing and S5 holds vacuously. This
  is a *different* defect from the rename fold and is not repaired by `--no-renames` —
  it is the specific edit AC-WSM-006's second falsification exists to fail.
- Substituting `core.quotePath=false` for `-z`. It suppresses quoting for non-ASCII paths
  only; backslash, tab, quote, and control-character paths stay quoted under any setting.
  The config change looks like a fix and is not one.
- Omitting `--literal-pathspecs` from the S5 comparison, or dropping it as "redundant now
  that the bytes are correct". Correct bytes are not the property at stake: git parses a
  pathspec before matching it, so a byte-perfect leading-colon path still names no file.
  This is the specific edit AC-WSM-006's third falsification exists to fail.
- Placing `--literal-pathspecs` after the subcommand (`git diff --literal-pathspecs ...`).
  It is a git-level option; git rejects that form with a usage error. This is the one repair
  on this surface whose misapplication is loud rather than silent — do not "fix" the usage
  error by deleting the flag.
- Substituting a per-element `:(literal)` prefix for the flag. It is mechanically correct
  but re-introduces a per-path string transformation, which is the operation that has failed
  on three of the four round-trip axes. Prefer the flag, which touches no path.
- Omitting `--no-textconv` from the S5 comparison, or dropping it because "the path list is
  already correct". Correctness of the list is not the property at stake: `git diff --quiet`
  compares a `textconv` driver's *rendering* of each blob, so two genuinely different files
  report no difference on a perfect list. This is the specific edit AC-WSM-006's fourth
  falsification exists to fail.
- Dropping `--ignore-submodules=none` from the S5 **enumeration** as redundant with the same
  flag on the comparison. It is not redundant — the two submodule-ignore keys suppress the
  gitlink at both stages, so without the enumeration-side flag the pointer never enters
  `names` and the comparison is never asked. Measured: removing it from *either* stage alone
  reproduces the SC-15 false positive. This is the specific edit AC-WSM-006's fifth
  falsification exists to fail, and it is the likeliest deletion in the whole mechanism
  because the flag appears twice and reads as a copy-paste error.
- Assuming a submodule exposure requires the operator's own config. `submodule.<name>.ignore`
  is read from `.gitmodules`, a tracked file, so a repository ships it to every clone. A
  repair reasoned as "this needs a non-default local setting, so it is rare" closes the
  wrong member.
- Replacing the S5 comparison with a blob-OID comparison (`git rev-parse <rev>:<path>`,
  `git cat-file --batch-check`, or an `ls-tree` map storing only the OID). It looks stronger
  — content hashes cannot be fooled by diff configuration — and is measurably weaker: it is
  mode-blind, and a symlink and a regular file with identical content share one blob OID, so
  P3c and P4c become silent false positives under **default** configuration.
  `cat-file --batch-check` is doubly disqualified, reporting `missing` for a gitlink on both
  sides at exit 0 so two different submodule pointers compare equal. If a future
  comparison-side defect ever forces this route, the mechanism must carry the **mode**
  alongside the OID; `spec.md` §2 records the full weighing.
- Routing any git invocation through a shell (`sh -c`, `bash -c`, an interpolated command
  string). It re-opens the argv axis in full and fails silently in the merged direction.
- Reconstructing the path list as a single shell string, or transcribing the document's
  shell snippets into Go. One pathspec matching nothing exits 0, which reads as merged.
- Reading an S5 result of "no difference" as reassuring without checking that the path list
  was non-empty and well-formed. S5 fails open: silence there is indistinguishable between
  "the branch's work is present" and "we asked about nothing".
- Treating every non-zero probe exit as a failure. `git diff --quiet` rc 1 and
  `git merge-base` rc 1 are verdicts; see `spec.md` REQ-WSM-007.
- Calling `git prune` or `git gc` to clean up the synthetic objects.
- Weakening the working-tree-cleanliness condition on the strength of better merge
  detection. The two conditions are independent (REQ-WSM-010).
- Widening the interface signature. It would force edits to four test doubles for no gain.

---

## §H Cross-References

- `spec.md` §2 — the executed detector matrix and the falsified alternative
- `spec.md` §5 — the four resolved scope decisions
- `acceptance.md` — per-criterion verification and falsification procedures
- `.claude/rules/moai/core/verification-claim-integrity.md` — evidence obligations
- `internal/cli/worktree/clean.go` — `cleanStaleWorktrees` `@MX:ANCHOR` (two-condition gate)
