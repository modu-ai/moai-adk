# SPEC-BACKLOG-HYGIENE-001 — Acceptance Criteria (card t332)

Every criterion below is binary and decided by a command run against the produced artifacts.
`$R` denotes `.moai/reports/t332`.

Sixteen criteria, the Tier M ceiling. v0.1.0 carried 14; plan-audit iter-1 required a digest
criterion (D5) and coverage for five uncovered requirements (D6). Rather than add five siblings and
breach the budget, the coverage additions are folded into existing criteria — AC-BH-012's scope is
widened, the `in-flight-unlanded` cross-check is attached to AC-BH-009, and REQ-BH-003's delta
branch is decided inside AC-BH-001 — and only two criteria are new (AC-BH-006, AC-BH-011).

## §A. Preconditions

- The sweep ran in the t332 worktree; `$R/queue-snapshot.tsv` exists with its capture time recorded.
- `$R/invocations.log` exists and carries every `moai todo` invocation the run phase issued.
- `$R/00-tooling-baseline.md` and `$R/01-scope.md` exist (M1, M2 outputs).

## §B. Traceability map

Every requirement has at least one criterion that can red it; every criterion decides at least one
requirement. Checked by hand at v0.2.0 against spec.md §D.

| REQ | Decided by | REQ | Decided by |
|---|---|---|---|
| REQ-BH-001 | AC-BH-001 | REQ-BH-009 | AC-BH-007 |
| REQ-BH-002 | AC-BH-001 | REQ-BH-010 | AC-BH-009 |
| REQ-BH-003 | AC-BH-001, AC-BH-002 | REQ-BH-011 | AC-BH-010 |
| REQ-BH-004 | AC-BH-004 | REQ-BH-012 | AC-BH-009 |
| REQ-BH-005 | AC-BH-004, AC-BH-006 | REQ-BH-013 | AC-BH-011 |
| REQ-BH-006 | AC-BH-005, AC-BH-006 | REQ-BH-014 | AC-BH-012, AC-BH-013 |
| REQ-BH-007 | AC-BH-004, AC-BH-006 | REQ-BH-015 | AC-BH-014 |
| REQ-BH-008 | AC-BH-008 | REQ-BH-016 | AC-BH-003, AC-BH-016 |

**AC-BH-015 decides a constraint, not a requirement.** It tests that the fan-out batch files carry
pairwise-disjoint id sets, which decides spec.md §E's write-isolation constraint ("read-only fan-out
workers write only their own file; two workers never share an output path"). It is deliberately not
given a requirement of its own — a constraint §E already carries does not need one, and
manufacturing one would cost a slot at the 16-requirement ceiling. v0.2.0's map listed AC-BH-015
under REQ-BH-007, which it does not decide; recording a deliberate absence as a presence is the
worse error of the two, so the cell is corrected rather than the decision revisited.

## §C. AC Matrix

### AC-BH-001 — the snapshot is the source, was not truncated, and any delta is recorded

**Given** `$R/queue-snapshot.tsv` (100 rows) and `$R/01-scope.md`,
**when** the scope file's queued count is compared to
`cut -f2 $R/queue-snapshot.tsv | grep -c '^queued$'`,
**then** the two are equal; **and** `$R/01-scope.md` names the snapshot's capture time and the HEAD
SHA it was captured at; **and** it records the comparison against spec.md §B.1's figure of 68 —
either "no delta" or the delta with the re-derived in-scope set.

The third conjunct is what decides REQ-BH-003's event branch: a sweep that never compared cannot
satisfy it, and one that compared and found nothing records that fact rather than staying silent.

Mutation: truncate the snapshot by 10 rows → the counts diverge and the criterion reds.

### AC-BH-002 — no picked or dropped card was swept

**Given** the picked id list `t278 t333 t338 t341 t346 t350 t354 t356 t357 t358` and the 18
`dropped` ids,
**when** the report's per-card entry ids are intersected with each list,
**then** both intersections are **empty**.

### AC-BH-003 — the report has exactly one entry per in-scope card

**Given** the consolidated `$R/report.md` and the per-batch files under `$R/cards/`,
**when** the per-card entry ids are extracted and counted,
**then** the count equals (snapshot queued count − 1), i.e. **67** at the recorded snapshot, and the
extracted id set equals the queued id set minus `t332`.

Both halves are required: a count alone passes against 67 entries for the wrong cards.

### AC-BH-004 — zero card-mutating `moai todo` invocations

**Given** `$R/invocations.log`,
**when** it is scanned for `moai todo (drop|edit|done|undrop|move|unpick)` and for
`moai todo next <id>` (a `next` carrying any argument),
**then** the match count is **0**, and the log is non-empty (`wc -l` ≥ the number of cards read).

Both conjuncts are required: an empty log also returns 0 matches, and would pass vacuously.

### AC-BH-005 — every relation record is well-formed

**Given** each `moai todo relate` line in `$R/invocations.log`,
**when** each is parsed,
**then** every one carries `--relation` with a value in
`{contains, absorbs, replaces, conflicts}` **and** a non-empty `--note`; **and** `moai todo` re-read
after the sweep shows the same queued/picked/dropped counts as the snapshot.

### AC-BH-006 — the card-row digest is unchanged across the sweep

**Given** the card-row digest produced by **this exact command**, captured into `$R/01-scope.md` at
M2 and re-captured at M5:

**Step 1 — resolve the primary checkout** (one plain call, nothing piped into it):

```bash
git rev-parse --path-format=absolute --git-common-dir
```

This prints the **common** git dir, which every linked worktree shares with the primary checkout
(e.g. `<primary>/.git`). Its **parent directory is the primary checkout**, and the store sits at
`<parent>/.moai/state/todo/backlog.json`. Do NOT use `git rev-parse --show-toplevel`: inside a
worktree that resolves to the *worktree* root, where no queue store exists — measured in the t332
worktree, `ls` on that path returns `No such file or directory`, rc=1. The queue is
primary-checkout-only, so a worktree-relative resolution reads nothing exactly where the run phase
executes.

**Step 2 — digest, taking the resolved path as a literal argument**:

```bash
jq -S -c '[.items[] | {id, state, text}] | sort_by(.id)' \
  <resolved-primary-checkout>/.moai/state/todo/backlog.json | shasum -a 256
```

**Why this is two steps and not one.** The obvious one-liner —
`P="$(dirname "$(git rev-parse --path-format=absolute --git-common-dir)")/…"; jq … "$P"` — is
**refused** by the worktree-isolation guard: *"this command is too complex to verify that it stays
inside the worktree"*. A compound command embedding a `git` call inside a substitution does not run
in a worktree session, which is where this card's run phase executes. Reproduced in the t332
worktree, 2026-08-29. Both captures (M2 and M5) therefore use the same two-step form, and the
resolved path is pasted as a literal rather than recomputed inline. The SPEC records the
**derivation**, never a machine-specific literal path.

**when** the two recorded digests are compared,
**then** they are **byte-identical**; **and** the M4 `relate` invocations, which run between the two
captures, are confirmed to have left them so; **and** each capture records the Step 1 output it
resolved, so a digest taken against the wrong tree is visible rather than silent.

**Why the projection is exactly this.** Verified against the on-disk store at
`.moai/state/todo/backlog.json`, 2026-08-29 (`jq -r 'keys'` → `findings`, `items`, `last_seq`,
`version`; `jq -r '[.items[]|keys]|add|unique'` → `added_at`, `id`, `spec_id`, `state`, `text`;
`jq -r '.items|length'` → 96, matching §B.4's card count). **`findings` is a top-level sibling of
`items`, not a per-item field** — `internal/kanban/backlog_store.go:191`
(`Findings []BacklogFinding \`json:"findings"\``) — so projecting `.items[]` excludes the
relate-appended findings **structurally**, not by an exclusion clause a reader has to honour.
`added_at` and `spec_id` are dropped deliberately: neither is what REQ-BH-005 protects. A wholesale
hash of the file would red on the sweep's own mandated `relate`, and on `last_seq` movement; that is
the wrong observable, not a stricter one.

**Both directions were measured before this criterion was written** (scratch copies via `jq`; the
real store was never written):

| Probe | Digest | Verdict |
|---|---|---|
| baseline, run twice | `56e1387e…b346b` | stable |
| **negative control** — append a `relate`-shaped finding | `56e1387e…b346b` | **unchanged**, as required |
| positive control — append ` MUTANT` to one card's `text` | `46e61f16…5fd28` | changed |
| positive control — flip one card's `state` to `dropped` | `fd9964a7…fb5a4` | changed |
| **two-step form, run from inside the t332 worktree** | `56e1387e…b346b` | reproduces the baseline — the procedure works where the run phase executes |

The negative control is what makes this criterion fail in **both** directions. Without it, an
extraction that wrongly included `findings` would satisfy the positive probe perfectly and then red
on the sweep's own legitimate `relate` — a criterion that punishes compliance. M4 runs `relate`
between M2 and M5, so the control is observed by the sweep's own ordering at no extra cost.

This is the third observable, and the only one not authored by the run phase about itself:
`$R/invocations.log` is written by the worker whose restraint it certifies, and AC-BH-005's count
comparison cannot see a `moai todo edit`, which changes a card's text while leaving all three counts
identical. Three observables that fail independently is the point — AC-BH-004 and AC-BH-005 are
kept, not replaced.

### AC-BH-007 — every landing verdict cites a pinned ref SHA, a commit SHA, and an ancestry result

**Given** `$R/00-tooling-baseline.md` and every card entry whose landing verdict is `landed`,
**when** they are read,
**then** the baseline file records the `git fetch origin develop main` invocation, its time, and the
`git rev-parse origin/develop origin/main` output pinning both refs; **and** every `landed` entry
carries (a) a commit SHA, (b) the **pinned** ref SHA it was found against — not the branch name —
and (c) the `git merge-base --is-ancestor` exit code observed for that pair.

Mutation: strip the `--is-ancestor` line from one entry, or replace a pinned SHA with
`origin/develop` → that entry reds.

### AC-BH-008 — landing was not determined from the lagging installed binary, on either path

**Given** `$R/00-tooling-baseline.md`,
**when** it is read,
**then** it records the measured
`strings ~/go/bin/moai | grep -c 'worktree_base_branch'` value and the chosen path; **and**:

- if the **governing count is `0`**, no card entry cites `moai todo pr` as the basis of a landing
  verdict — **whichever path was declared**. The **governing count** is the one measured against the
  binary that actually produced the landing verdicts: on path A the single pre-sweep measurement, and
  on path B the **post-rebuild** measurement. Path B leaves two counts on the record and only the
  later one governs; naming which is what stops a passing post-rebuild count from being read back
  onto verdicts taken before the rebuild, or a failing pre-rebuild count from condemning verdicts
  taken after it;
- if path **B** (rebuild) was chosen, the file records that post-rebuild count as **≥ 1** together
  with the reinstall command that produced it, and records the order — rebuild first, verdicts after.

The guard is on the measured value, not on the declared path. v0.1.0 guarded the binding conjunct on
"path A was chosen", so a run that declared path B, skipped the rebuild, and cited the installed
landed column anyway made the antecedent false and passed while breaching REQ-BH-008.

Control (the positive half, so the criterion is not vacuous): the baseline file carries the two
t342 control queries — non-empty against the pinned `develop` SHA, empty against the pinned `main`
SHA.

### AC-BH-009 — no landing verdict rests on a branch name, and the in-flight class is populated

**Given** every card entry with a `landed` or `in-flight-unlanded` verdict, and the live worktree
list captured at M1,
**when** the entries and the list are compared,
**then** any branch name an entry names is accompanied by a commit-message id match or a SHA — **no**
entry's sole evidence is a branch-name resemblance; **and** every in-scope card appearing in the
captured worktree list with an unmerged branch carries the `in-flight-unlanded` classification with
its branch and tip SHA.

The second conjunct decides REQ-BH-012's **positive** direction. Quantifying only over entries that
already read `in-flight-unlanded` passes vacuously on the empty set — a sweep that never assigns the
classification would satisfy it.

### AC-BH-010 — `unknown` is preserved, never collapsed

**Given** every card whose landing query could not be answered,
**when** the entry is read,
**then** the verdict reads `unknown` with a stated reason, and does **not** read `not-landed`.

Mutation: relabel one `unknown` as `not-landed` without a query → the reason line has no command to
cite and the criterion reds.

### AC-BH-011 — every entry carries a three-valued premise verdict

**Given** every one of the 67 card entries,
**when** the entry's premise block is read,
**then** it carries the card's central premise restated in one sentence **and** a verdict in
`{holds, falsified, unverified}`; **and** every `unverified` verdict carries the reason it could not
be decided.

Mutation: strip the verdict token from one entry, or leave an `unverified` entry without a reason →
that entry reds.

### AC-BH-012 — the five evidence sections appear on EVERY entry

**Given** every one of the 67 card entries — not only the `falsified` ones —
**when** the entry is read,
**then** all five sections (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) are
present; **and** every `falsified` entry additionally carries the command that falsified the premise
together with that command's verbatim output, not a paraphrase.

v0.1.0 scoped the five-section conjunct to `falsified` entries, so a report whose 60 `holds` entries
carried a bare verdict satisfied every criterion while breaching REQ-BH-014.

### AC-BH-013 — every absence claim names its scanned scope

**Given** every entry asserting that a card is already resolved, already landed, or no longer
reproducible,
**when** the entry is read,
**then** it names the files, refs, or commands scanned to reach that conclusion.

### AC-BH-014 — every overlap candidate names the shared artifact

**Given** `$R/02-overlaps.md`,
**when** each overlap row is read,
**then** each names both card ids and the specific shared file, mechanism, or ref that grounds the
overlap — a bare "these look similar" row fails; **and** no row proposes a fold as performed rather
than proposed.

### AC-BH-015 — fan-out write isolation held

**Given** the batch files under `$R/cards/`,
**when** their per-card entry id sets are pairwise intersected,
**then** every pairwise intersection is **empty**, and the union equals the in-scope id set of
AC-BH-003.

### AC-BH-016 — the disposition list is a proposal, and it is complete

**Given** `$R/report.md`,
**when** the disposition list is read,
**then** it carries one row per in-scope card, each row naming a disposition in
`{keep, drop, fold-into <id>, already-landed, needs-operator-decision}` and the single piece of
evidence it rests on; **and** the report states explicitly that no card was mutated and that the
list awaits the operator; **and** no row proposes an un-drop, or a fold whose counterpart is a
`dropped` card.

## §D. Definition of Done

- M1..M5 closed, each citing the command it ran into `progress.md` §E.2.
- AC-BH-001..016 all decided PASS, each with its deciding command recorded.
- Evidence tree recovered to the primary checkout and `cmp`-verified (plan.md M5).
- The card-row digest is identical across the sweep (AC-BH-006).

## §E. Quality Gates

- No source file modified: `git status --short` shows changes only under
  `.moai/specs/SPEC-BACKLOG-HYGIENE-001/` and `.moai/reports/t332/`. Note that this gate says
  **nothing** about the queue store — `backlog.json` lives in a gitignored state dir and is invisible
  to `git status` whether or not a card was mutated. AC-BH-006 is what decides that question.
- No template mirror obligation (nothing under `internal/template/templates/` is touched).
