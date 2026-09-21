---
id: SPEC-GUARD-COMMENT-SCAN-001
title: "BranchGuard scan preprocessing: a shell comment is text the shell never executes, and must not be scanned as the command being run (card t1056)"
version: "0.1.0"
status: draft
created: 2026-09-21
updated: 2026-09-21
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
$ grep -c '#'  internal/hook/branch_guard.go                    → 0
$ grep -c '<<' internal/hook/branch_guard.go   # same file, same form, control → 6
$ grep -rc '#' internal/hook/ --include='*.go' | grep -v ':0'   # the pattern CAN hit → 3 files
```

The `#` character does not occur anywhere in the file; the two controls fire, so the zero is a
measured absence rather than a broken grep.

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
already become `" X "`. The opposite order is measurably wrong and blinds the guard:
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
- **REQ-GCS-003** — The elision SHALL be bounded to the line the comment opens on. **When** a
  comment line is followed by a further line, the preprocessing step SHALL leave that following line
  intact and scannable.
- **REQ-GCS-004** — The comment run SHALL be elided rather than replaced by the operand placeholder
  `quotedArgumentPlaceholder`, because a comment is removed by the shell rather than passed to the
  command as an argument (§B.3).
- **REQ-GCS-005** — Comment-borne git prose SHALL NOT match any branch-state pattern (the
  mutation-detected arm).
- **REQ-GCS-006** — The guard SHALL continue to match real branch-state commands after the change
  (the no-mutant-success arm). This requirement exists because REQ-GCS-005 alone is satisfied by a
  guard that has been disabled.
- **REQ-GCS-007** — Comment elision SHALL run last in the pipeline, after quoted-argument and
  heredoc collapse, so that a `#` inside a quoted span or a heredoc body does not open a comment
  (§B.4).
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

## §G Baseline Attribution

All figures in §A were measured in the worktree `.claude/worktrees/t1056`, branch `WT-guard-prose`,
at HEAD `3dfae918a`, in this session. The matcher figures come from a temporary probe test invoking
`matchBranchStateCommand` directly; the probe was removed afterwards and the tree left clean. The
grep figures in §A were re-measured during plan-phase authoring, on the same tree.

Evidence file: `.moai/reports/t1056/reproduction.md`. That path is gitignored
(`.gitignore:227`), so the file is untracked and this worktree holds the only copy — see `plan.md`
§E.
