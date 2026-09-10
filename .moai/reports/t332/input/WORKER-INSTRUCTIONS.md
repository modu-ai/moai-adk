# t332 per-card sweep — worker instructions (SPEC-BACKLOG-HYGIENE-001, M3)

You are a READ-ONLY investigator. Read this file fully before your first card.

## Working directory

`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t332` (a git worktree). All paths below are
relative to it. Drive git with plain commands from this directory — never `cd` to the primary
checkout, and never write outside your own output file.

## Your input and your single output

- Input: `.moai/reports/t332/input/batch-<k>.tsv` — tab-separated `id \t state \t text`, one card
  per line. These are your cards and no one else's.
- Output: `.moai/reports/t332/cards/batch-<k>.md` — **the only file you write.** Do not append to
  any shared file. Do not create scratch files under `.moai/`.

## HARD prohibitions

1. **Never invoke a card-mutating `moai todo` verb**: `drop`, `edit`, `done`, `undrop`, `move`,
   `unpick`, or `next <id>`. Do not invoke `moai todo relate` either — relations are M4's, not
   yours. In fact **do not invoke `moai todo` at all**; everything you need is in your `.tsv`.
2. **Never fix a defect you find.** A falsified premise is a finding. The repair is a separate card
   the operator issues. If you catch yourself editing source, stop.
3. **Never determine landing from a branch name.** A branch called `WT-foo` resembling card text is
   not evidence. A prior sweep misattributed t342 to `WT-check-must-fail` doing exactly this.
4. **Never report `not-landed` for a query you could not answer.** `unknown` and `not-landed` are
   different facts. An unanswerable query gets `unknown` plus the reason.
5. **Never cite `moai todo pr`'s landed column.** The installed binary (v3.1.2) predates the
   integration-branch fix — measured `strings ~/go/bin/moai | grep -c worktree_base_branch` = 0 —
   so it answers about `origin/main` only and is silent about every `develop` landing.

## Pinned refs — use these SHAs as literals, never the branch names

```
origin/develop  ee50984abe4f11ac337382b48a26328f091e200a
origin/main     48239c7dc7428c8751a04f6321887c2d36123884
```

Both were fetched once at 2026-08-30T11:16:22Z. Do not fetch again; do not write `origin/develop`
into a verdict.

## Landing determination (run per card)

```
git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt<NNN>\b' --oneline
git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt<NNN>\b' --oneline
```

If a commit is found, confirm ancestry and record the exit code:

```
git merge-base --is-ancestor <commit-sha> ee50984abe4f11ac337382b48a26328f091e200a ; echo $?
```

**Watch for false positives**: `--grep='\bt90\b'` can match a commit that merely *mentions* t90
(e.g. another card's message citing it). Read the matched commit's subject and decide whether it
*delivers* the card or merely names it. A mention is not a landing; say so explicitly.

Verdict vocabulary — exactly one of:

- `landed` — a commit delivering this card is an ancestor of a pinned SHA. Cite (a) the commit SHA,
  (b) the **pinned ref SHA** it was found against, (c) the `--is-ancestor` exit code.
- `not-landed` — both queries ran and returned no delivering commit. Say which two queries ran.
- `in-flight-unlanded` — the card has a live worktree with an unmerged branch. Cite the branch and
  its tip SHA from `.moai/reports/t332/00-worktree-list.txt` (read that file; do not re-run
  `git worktree list`), and confirm the tip is not an ancestor of pinned develop.
- `unknown` — the query could not be answered. State why.

## Premise verdict (run per card)

Restate the card's central premise in **one sentence**, then decide exactly one of:

- `holds` — the premise is still true. Say what you checked.
- `falsified` — measurement contradicts it. **You must cite the command that falsified it together
  with that command's verbatim output.** A paraphrase is not evidence.
- `unverified` — you could not decide. **State the reason.** This is an honourable verdict; a
  plausible reading promoted to `holds` is not.

Verify against the tree at the worktree HEAD. Typical checks: `grep -n` for the file/symbol/pattern
the card names; reading the file at the line the card cites; `git log <pinned-develop> --oneline --
<path>` for whether the area was already repaired. **Name the scope you scanned** — an absence claim
whose scanned scope is unnamed is a Gap, not a finding.

## Required output shape — per card, no exceptions

Every card gets all of this, including cards whose premise `holds`. A bare verdict fails the
acceptance criteria.

```markdown
### <id>

**Premise (one sentence).** …

**Premise verdict.** `holds` | `falsified` | `unverified` — <reason; on `falsified`, the falsifying
command and its verbatim output; on `unverified`, why it could not be decided>

**Landing verdict.** `landed` | `not-landed` | `in-flight-unlanded` | `unknown`
- commit: <sha or —>
- pinned ref: <the literal 40-hex SHA, or —>
- `--is-ancestor` exit: <0|1|—>
- branch + tip (in-flight only): <name> @ <sha>

**Claim.** <what you assert about this card, one or two sentences>

**Evidence.** <the commands you ran and their verbatim output — fenced>

**Baseline-attribution.** <which tree/ref/file each figure was measured against, in this run>

**Gaps.** <what you did NOT check — be specific; "none" only if truly nothing was left unobserved>

**Residual-risk.** <what could still be wrong despite what you observed>

**Proposed disposition.** `keep` | `drop` | `fold-into <id>` | `already-landed` |
`needs-operator-decision` — with the single piece of evidence it rests on.
This is a PROPOSAL. You are not deciding; the operator is.

**Overlap candidates.** <other card ids from the full in-scope list that touch the same file,
mechanism, or ref — name the shared artifact. "none observed" is a valid answer.>
```

The full in-scope id list is `.moai/reports/t332/input/inscope-all.txt` (62 ids) — read it so your
overlap candidates can name cards outside your own batch.

## Output file header

Start your file with:

```markdown
# t332 card sweep — batch <k>

Cards: <the ids, space-separated>  (<n> entries)
Worktree HEAD: <git rev-parse --short HEAD>
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884
```

Then one `### <id>` section per card, in the order they appear in your `.tsv`.

## Restraint

Depth per card is bounded: a handful of targeted `grep`/`git log`/`Read` calls. If a card's premise
needs a deep investigation to settle, that IS the finding — record `unverified` with the reason and
propose `needs-operator-decision`. Do not spend the batch on one card.
