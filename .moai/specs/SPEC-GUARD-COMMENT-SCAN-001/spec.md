---
id: SPEC-GUARD-COMMENT-SCAN-001
title: "BranchGuard scan preprocessing: a shell comment is text the shell never executes, and must not be scanned as the command being run (card t1056)"
version: "0.1.0"
status: completed
created: 2026-09-21
updated: 2026-09-22
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "internal/hook"
lifecycle: spec-anchored
tags: "branch-guard, scan-preprocessing, shell-comment, false-positive, two-armed-mutation, t1056"
tier: S
---

# SPEC-GUARD-COMMENT-SCAN-001 — Shell Comments Are Not Commands

## §A Context and Problem

`matchBranchStateCommand` (`internal/hook/branch_guard.go:278`) decides whether a Bash command is a
branch-state mutation the guard should refuse. Before scanning, it runs a preprocessing pipeline
whose whole purpose is to separate **the command being run** from **text the command merely
carries**:

```go
scanned := substituteQuotedArguments(substituteHeredocBodies(command))
```

Two axes of that separation already exist, each landed after a measured false positive:

- **`substituteQuotedArguments`** (`branch_guard.go:197`) collapses a quoted span to the single
  non-flag placeholder word `" X "`. A `moai todo add "… git switch …"` call was denied because the
  guarded text sat inside an argument that would never execute.
- **`substituteHeredocBodies`** (`branch_guard.go:234`) collapses heredoc bodies, because a heredoc
  body is stdin DATA. A `moai handoff save --stdin … <<EOF … EOF` was denied because the resume body
  it was saving named `git merge --no-ff <sha>`; the lane then skipped the save, closing the
  handoff-record path.

**A shell comment is the missing third axis, and it is the same defect class: prose read as a
call.** A comment line is text the shell strips before execution — it can never be the command being
run. Measured in this tree at HEAD `3dfae918a`:

```
matched=true  suffix="git merge"  cmd=# align with git merge --ff-only develop
```

The pipeline has no comment handling at all. Measured, with positive controls so the absence is not
an instrument failure:

```
$ grep -c '#'  internal/hook/branch_guard.go
0
$ grep -c '<<' internal/hook/branch_guard.go     # same file, same form, control
6
$ grep -rc '#' internal/hook/ --include='*.go' | grep -v ':0' | wc -l   # the pattern CAN hit
     112
```

The `#` character does not occur anywhere in the file; the two controls fire, so the zero is a
measured absence rather than a broken grep.

All three figures were **re-measured in this tree** — the first two at HEAD `5bc42a304`, all three
again at HEAD `9d822d826` with the same outputs — and the blocks above are that run's verbatim
stdout.

The third figure has been corrected twice, and both corrections are recorded because the second one
was made inside the sentence repairing the first. An early draft cited it as "3 files", which does
not reproduce. The repair that replaced it wrote **"112 (of 282 `.go` files at
`internal/hook/*.go`)"** — a numerator and a denominator drawn from **different populations**: 112
comes from a *recursive* scan of `internal/hook/`, whose population is 412; 282 is the file count
of the *non-recursive* glob `internal/hook/*.go`, whose numerator is 72. Each command and its
observed output, measured at HEAD `9d822d826`, this run:

```
$ grep -rc '#' internal/hook/ --include='*.go' | grep -v ':0' | wc -l   # recursive numerator
     112
$ grep -rc '#' internal/hook/ --include='*.go' | wc -l                  # recursive population
     412
$ grep -c '#' internal/hook/*.go | grep -v ':0' | wc -l                 # non-recursive numerator
      72
$ ls internal/hook/*.go | wc -l                                         # non-recursive population
     282
```

The attributable pair is therefore **112 of 412 (recursive)**, equivalently 72 of 282 for the
top-level directory alone. The block's conclusion is unchanged in every reading — the pattern
reaches non-zero files in this very directory, so the `0` above is a measured absence rather than
an instrument failure.

### The discriminant — a refusal without the sentinel was not produced by this guard

[HARD] **A refusal message carrying no `BRANCH_GUARD_VIOLATION:` prefix was NOT produced by this
repository's BranchGuard.** Recorded here because asserting the refuser from the symptom alone is
what reversed this card's original premise: the card was dispatched as "BranchGuard over-matches
E3", and E3's refusal turned out to come from the Claude Code **runtime** worktree-isolation guard
— its text carries no sentinel, and `matchBranchStateCommand` returns `false` for that exact
command. Nothing in this repository can change E3. The symptom ("a command containing `git` was
refused") does not name the refuser; the prefix does. Read the prefix before attributing.

### Why this narrowing is safe to make and how it can fail

The change makes the guard match **less**. A narrowing cannot be validated by a single arm: a test
showing "comment prose no longer matches" is satisfied equally well by a correct narrowing and by a
guard that has been blunted into matching nothing. Both arms are therefore requirements below
(REQ-GCS-005, REQ-GCS-006), not conveniences.

## §B Design Constraints

Four constraints bind any implementation. Each is a property the shell has, and each has a failure
mode if ignored.

**B.1 — `#` opens a comment only at WORD START.** POSIX: `#` begins a comment when it is the first
character of a word — at line start, or preceded by whitespace or a command separator (`;`, `&`,
`|`, `(`). A `#` inside a word does not: `git switch feat#123` and a path `a#b` carry a literal
hash. A rule that stripped from any `#` to end-of-line would blind the guard on real commands whose
operands contain one.

**"Word start" is not "preceded by an alphanumeric character", and the difference is measurable.**
A character can be non-alphanumeric and still leave the `#` mid-word — `/`, `=`, `.`, `-` all do.
An implementation reading the rule as "preceding character is non-alphanumeric ⇒ comment" satisfies
a word-boundary criterion built only from alphanumeric examples, and then blinds the guard on
operands carrying a path or an assignment. Measured in this tree at HEAD `5bc42a304`, this run
(`SECOND-CMD-RAN` stands in for a branch-state command, so no branch state is mutated):

```
$ bash -c 'echo bar/#x ; echo SECOND-CMD-RAN'
bar/#x
SECOND-CMD-RAN
$ bash -c 'echo a=#b ; echo SECOND-CMD-RAN'
a=#b
SECOND-CMD-RAN
$ bash -c 'echo a.#b ; echo SECOND-CMD-RAN'
a.#b
SECOND-CMD-RAN
$ bash -c 'echo a-#b ; echo SECOND-CMD-RAN'
a-#b
SECOND-CMD-RAN
```

Positive control for the same instrument — a `#` that genuinely **does** open a comment suppresses
the following word, so the probe is not simply printing everything:

```
$ bash -c $'echo a ; # echo B\necho SECOND-CMD-RAN'
a
SECOND-CMD-RAN
```

(`B` is absent: the comment opened. Without this control the four literal-hash rows above would be
consistent with a probe that never opens a comment at all.)

The falsification for this constraint is AC-GCS-004, which carries **one row per character listed
above** — `/`, `.`, `=`, `-` — each written so a real branch-state command sits on the far side of
the `#`. One row per character is what the constraint needs: the widenings are independent, so an
implementation that adds a single one of these four to its word-start set is caught by that
character's row and by no other. A row that merely *contains* the character, without a command
after the `#` that the elision would remove, excludes nothing — see the row-discrimination rule in
`acceptance.md` AC-GCS-004, which records the measured instance of that mistake in this very SPEC.

**B.2 — collapse is PER LINE.** The comment runs to the end of its own line and no further. A
comment line followed by a real command on the next line must leave that command fully scannable —
the same bound `substituteHeredocBodies` already keeps for heredoc bodies.

**B.3 — the comment is ELIDED, not replaced by the operand placeholder.** The existing collapse
discipline substitutes rather than deletes, so neighbouring tokens cannot fuse into an accidental
match. That discipline is honoured here by a different substitution, for a stated reason: a quoted
span **is** an operand the shell passes to the command, so `" X "` models it faithfully; a comment
is **removed** by the shell and is not an operand, so modelling it as one would misstate the shell.
Fusion — the hazard the discipline exists to prevent — is structurally impossible on this axis:
B.1 guarantees the `#` is always preceded by a word boundary, and end-of-line bounds the other
side, so nothing on either side can join. The observable consequence is pinned as AC-GCS-005:
`git checkout -b # x` MUST NOT match, because after the shell strips the comment that command has
no branch operand — exactly what the shell itself would run.

**B.4 — ordering within the pipeline is load-bearing and is decided, not left implicit.** A `#`
inside a quoted argument, and a `#` inside a heredoc body, are not comments. Comment collapse
therefore runs **last** — outermost, on the already-quote-and-heredoc-collapsed string:

```go
scanned := substituteShellComments(substituteQuotedArguments(substituteHeredocBodies(command)))
```

Running it last is not a stylistic preference; it is what makes B.1's quoted-`#` case correct for
free, because by then a quoted `#` has already become part of `" X "` and a heredoc-body `#` has
already become `" X "`. The opposite order is demonstrably wrong and blinds the guard:
`echo "text # more" ; git switch main` — with comment collapse first, the line is truncated at the
quoted `#` and `git switch main` disappears from the scan. That case is AC-GCS-006.

## §C Requirements

Modality is `SHALL`, matching this repository's measured convention.

- **REQ-GCS-001** — The scan-preprocessing pipeline SHALL elide shell comments before pattern
  matching, so a branch-state pattern matches the command being run rather than prose carried in a
  comment.
- **REQ-GCS-002** — The preprocessing step SHALL treat a `#` as opening a comment only where it
  appears at word start (line start, or preceded by whitespace or one of `;`, `&`, `|`, `(`), and
  SHALL NOT treat a `#` appearing inside a word as opening one.
- **REQ-GCS-003** — The elision SHALL be bounded to the **physical** line the comment opens on.
  **When** a comment line is followed by a further line, the preprocessing step SHALL leave that
  following line intact and scannable. (Physical, not logical: a backslash-newline continuation is
  a known residual recorded in §F, not a second bound this requirement asserts.)
- **REQ-GCS-004** — The comment run SHALL be elided rather than replaced by a non-flag operand
  token, because a comment is removed by the shell rather than passed to the command as an argument
  (§B.3). The **comment run** is the span beginning at the `#` that opened the comment and ending
  at the end of that physical line, inclusive of the `#` and exclusive of the newline; the
  preprocessing step SHALL leave the text **preceding** that `#` on the same line intact and
  scannable, so a branch-state command carrying a trailing comment still matches.
- **REQ-GCS-005** — Comment-borne git prose SHALL NOT match any branch-state pattern (the
  mutation-detected arm).
- **REQ-GCS-006** — The guard SHALL continue to match real branch-state commands after the change
  (the no-mutant-success arm). This requirement exists because REQ-GCS-005 alone is satisfied by a
  guard that has been disabled.
- **REQ-GCS-007** — A `#` appearing inside a quoted span or inside a heredoc body SHALL NOT open a
  comment, so a branch-state command sharing a line with such a `#` remains scannable. (§B.4
  records the pipeline ordering that obtains this outcome; the requirement states the outcome, and
  AC-GCS-006 observes it.)
- **REQ-GCS-008** — The two-armed test SHALL live in `internal/hook/branch_guard_comment_test.go`,
  mirroring the per-axis file convention `branch_guard_heredoc_test.go` and
  `branch_guard_quoted_test.go` already establish.

## §D Acceptance Criteria

Enumerated in `acceptance.md` (AC-GCS-001 … AC-GCS-006) with Given-When-Then scenarios and the
requirement ↔ criterion map. Both mutation arms are mandatory criteria there, per REQ-GCS-005 and
REQ-GCS-006.

## §E Scope

### In Scope

- One preprocessing step added to `internal/hook/branch_guard.go`, plus its wiring into
  `matchBranchStateCommand`'s pipeline expression.
- One new test file, `internal/hook/branch_guard_comment_test.go`, carrying both mutation arms.

### Exclusions — what this SPEC does NOT build

Each entry is a decision recorded with its reason, not an omission.

#### Out of Scope — E3, the `cd <primary> && echo "git"` refusal

- **Measured to be outside this repository entirely.** The refuser is the Claude Code **runtime**
  worktree-isolation guard, not this repository's BranchGuard: the refusal text carries no
  `BRANCH_GUARD_VIOLATION:` sentinel, and `matchBranchStateCommand` returns `false` for that exact
  command (`matched=false suffix=""`).
- **No source change here can alter it.** Narrowing the BranchGuard to chase E3 would weaken the
  guard while leaving the symptom exactly where it was.

#### Out of Scope — E6, command substitution inside loops

- A separate axis of the same general family, with its own reproduction and its own risk direction.
  It was not measured in this card and is not a card. Folding it in would widen what this change
  must be right about, in the same change that is establishing the comment axis.

#### Out of Scope — reinstalling the stale `~/go/bin/moai` binary

- **Not a code change.** The installed binary (`…-1452-gf67d2193f`, built 2026-09-17) predates the
  heredoc repair `3eda5f028` (2026-09-18) — measured, `git merge-base --is-ancestor 3eda5f028
  f67d2193f` → exit 1, with a three-way positive control confirming the verb discriminates in this
  tree. The source is already repaired; the binary running the hook does not carry the repair.
- The remedy is a reinstall, tracked as milestone M1 in the card's evidence file and deliberately
  held for operator judgment: several lanes in this batch share that hook, so swapping the installed
  binary mid-batch would move other lanes' measurement baselines.

#### Out of Scope — any change to the branch-state pattern set

- `branchStatePatterns` is untouched. This SPEC changes only what string the patterns are scanned
  against, never which patterns exist or what they match.

## §F Known Gaps

Carried forward rather than papered over.

- **The over-match is reproduced at the MATCHER level only.** Whether a comment-borne false positive
  reaches a user-visible deny through the full hook path — the opt-in flag
  `Workflow.BranchGuard.Enabled`, primary-checkout discrimination, and the exemption axes — was NOT
  measured. `plan.md` §D records the decision to keep this a Gap rather than close it, with the
  coordinate for whoever closes it later.
- **The card's originating incident is not attributed.** The verbatim text of the lead's 2026-09-21
  refusal was never captured, so the stale-binary account above shows *that this symptom follows
  from a stale binary*, not that *that event was this*. Stated as cause-established,
  attribution-unestablished.

### Accepted residuals — cases the stated rule gets wrong, and the direction of each

These are not gaps in measurement; they are places where the rule in §B / §C is knowingly not the
shell's answer. Each names its direction, because the two directions have very different costs:
**blinding** (the guard scans less than the shell runs — a real command escapes the guard) is the
unsafe one; **over-match** (the guard scans text the shell discards) is the defect class this card
is narrowing, and is merely noisy.

- **Manufactured `#` after a quoted span — direction: BLINDING. Decided: ACCEPT.** The
  quoted-argument placeholder `" X "` introduces a space, so `foo"bar"#baz` reaches the comment
  step as `foo X #baz` and its `#` reads as word-initial. Because elision runs to end-of-line, the
  blinding is **not** limited to that token: any branch-state command later on the same line
  disappears from the scan. Measured in this tree at HEAD `5bc42a304`, this run — `bash -c 'echo
  "q"#b ; echo SECOND-CMD-RAN'` printed `q#b` then `SECOND-CMD-RAN`, so bash runs both statements
  while a conforming implementation would scan neither past the manufactured `#`. Accepted rather
  than repaired: stopping the elision at a command separator would re-admit the over-match this
  card removes, and the sound fix belongs to `substituteQuotedArguments`, a different step outside
  this SPEC's In Scope. Rationale, both rejected repairs, and the follow-up coordinate: `plan.md`
  §A.
- **Backslash-newline continuation — direction: BLINDING. Decided: ACCEPT (scope).** A
  backslash-newline is removed during line-joining before tokenization — **and only where no
  whitespace precedes the backslash** does joining leave the next physical line's `#` mid-word and
  literal to the shell. That condition is load-bearing and it belongs to the cited source, which
  measured the two variants separately. Where whitespace *does* precede the backslash, joining
  leaves the `#` at word start and bash opens a comment there as well — the rule in §B.2 and the
  shell then agree, and no residual exists. The residual is the no-whitespace case only: there the
  rule in §B.2 / REQ-GCS-003 treats that physical line as a line-start comment and elides it,
  dropping any real command riding on it. An earlier edition of this entry stated the consequence
  without the condition, which asserted a wider residual than the source established; the condition
  is restored here, and restoring it narrows the claim rather than widening it.
  **Not re-measured in either repair run, and not by the second auditor either**: every probe form
  for this construct (`bash -c $'…\\\n…'`, a `printf … | bash` pipe, and a plain-character control)
  was **refused** by the Claude Code runtime worktree-isolation guard — a refusal carrying no
  `BRANCH_GUARD_VIOLATION:` sentinel, i.e. not this guard (§A discriminant). Routing around the
  guard by moving the probe into a script file was available and was **not** taken. The second
  plan-audit verdict records the same refusal against the same construct
  (`.moai/reports/t1056/verdict-iter2.md` Gap 1). The direction stated here — and the whitespace
  condition restored above — therefore rest on the POSIX line-joining rule plus the **first**
  verdict's bash measurement (`.moai/reports/t1056/verdict.md` E1), and are carried as
  cited-not-re-measured figures rather than as observations of any repair run. The correction is a
  correction of the citation, not a new measurement, and is not to be read as one.
  Handling continuations would *narrow* the elision (the safe direction) and is omitted on scope
  grounds only — see `plan.md` §C.
- **`)` and `}` absent from the word-start set — direction: OVER-MATCH. Decided: ACCEPT (record
  only).** The plan-audit verdict (`.moai/reports/t1056/verdict.md` E1) measured `(echo a)#echo B`
  and `{ echo a;}#b` as opening comments in bash, which the §C set does not admit; the guard would
  therefore scan text the shell discards. **Not re-measured in this repair run** — the probe form
  was refused by the same runtime guard described above — so this entry is cited, not observed
  here. Recorded so the next reader does not re-derive it; it does not blind the guard, which is
  why it is not repaired in this card.

## §G Baseline Attribution

All figures were measured in the worktree `.claude/worktrees/t1056`, branch `WT-guard-prose`.
Two measurement points exist, and each figure states which one it belongs to.

- **HEAD `3dfae918a` (plan-phase authoring)** — the matcher figures (`matched=true suffix="git
  merge"` in §A, and the AC-GCS-001 / AC-GCS-002 RED-now and preservation observations in
  `acceptance.md`). They come from a temporary probe test invoking `matchBranchStateCommand`
  directly; the probe was removed afterwards and the tree left clean.
- **HEAD `5bc42a304` (first plan repair)** — the first two grep figures in §A, and the bash
  observations in §B.1 and §F.
- **HEAD `9d822d826` (second plan repair, this run)** — the four-command population block in §A
  (the N3 correction), the bash observations backing the rows added to `acceptance.md` AC-GCS-001
  rows 4-5 and AC-GCS-004 rows 3-6, and a re-run of the §A `0` / `6` / `112` figures, which
  reproduced unchanged.

The two points are interchangeable **for this SPEC's subject**, and that is measured rather than
assumed: `git rev-parse 3dfae918a:internal/hook/branch_guard.go` and
`git rev-parse HEAD:internal/hook/branch_guard.go` both return `70d6da28da9449c9d4accee9c654c597112a92a1`
— the file under change is byte-identical across the two, so a matcher figure taken at the earlier
point describes the same code the later figures were taken against.

Two residuals in §F carry figures that were **not** re-measured in this run: the continuation-line
direction and the `)` / `}` over-match. Both probe forms were refused by the Claude Code runtime
worktree-isolation guard, so both are cited to the plan-audit verdict
(`.moai/reports/t1056/verdict.md` E1) with that status stated inline, per the refused-tool
degradation rule (`verification-claim-integrity.md` §3.1).

Evidence file: `.moai/reports/t1056/reproduction.md`. That path is gitignored
(`.gitignore:227`), so the file is untracked and this worktree holds the only copy — see `plan.md`
§E.
