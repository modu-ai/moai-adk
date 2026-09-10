# Audit response — SPEC-TODO-ARCHIVE-QUERY-001, plan-audit iteration 2

Responding to `.moai/reports/t394/plan-audit-iter2.md` (PASS-WITH-DEBT, 0.875
against the Tier M threshold of 0.80).

Tree: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t394`, branch
`WT-todo-done-history`, working tree at `2c18091d1`. `status` stays `draft`. No
implementation code, no tests, no commits, no pushes.

**Closed: 7 of 7** — D1-D4 (blocking-class) and D5-D7 (optional). Nothing is left
open. Three things the repair deliberately did **not** do are stated at the end,
because two of them look like closures and are not.

Per the Retry Loop Contract, Tier M's ceiling is 2 iterations and iteration 2
passed, so no iteration 3 is authorised. These repairs are recorded here for the
run phase, not submitted for another audit round.

---

## Defect ledger

| # | Class | Closed | What changed (file + section) | Command that shows it closed |
|---|---|---|---|---|
| D1 | blocking | yes | `audit-response-iter1.md` § D2 deviation — the `fetch-depth: 1` bullet struck in place, marked `[CORRECTED at iteration 2]`, the repository's actual configuration measured and quoted, and the deviation re-justified on a cost premise that holds | `grep -n "CORRECTED at iteration 2" .moai/reports/t394/audit-response-iter1.md` → `50:` ; `grep -c "fetch-depth: 0" …` → `6` |
| D2 | blocking | yes | `acceptance.md` AC-TAQ-011 clause 1 — `git diff --exit-code "$C" -- "$GOLDENS"` added; the header prose replaced (the false "no capture taken from the post-change tree can satisfy this AC" is gone) with an exact statement of what clause 1 does and does not establish | `grep -n 'git diff --exit-code "\$C" -- "\$GOLDENS"' acceptance.md` → `217:` ; `grep -c "no capture taken from the post-change tree can satisfy this AC" acceptance.md` → `0` |
| D3 | blocking | yes | (a) `acceptance.md` clause 1 — singleness asserted. (b) `acceptance.md` clause 1 — scoped to the pre-integration branch; `plan.md` §D — merge-method dependency recorded with the measurement. (c) `acceptance.md` after clause 3 — develop-drift residual stated, including the regenerate-without-reading-the-diff hazard `grep -n "tr -d" acceptance.md` → `210:` ; `grep -n "scoped to the pre-integration branch" acceptance.md` → `238:` ; `grep -n "depends on the merge method" plan.md` → `109:` ; `grep -n "clause 3 reds on develop" acceptance.md` → `273:` |
| D4 | blocking | yes | `spec.md` §A.4 — a `[MEASUREMENT PROVENANCE]` block naming checkout, binary and instant, two timestamped snapshots, and the 112/113 reconciliation. `spec.md` §D (both Out-of-Scope blocks) and `plan.md` §B.1/§B.3/§B.4 each carry an instant and cite the block | `grep -n "MEASUREMENT PROVENANCE" spec.md` → `107:` ; `grep -c "§A.4 provenance" spec.md plan.md` → `spec.md:3`, `plan.md:1` ; `grep -ci "measured this turn" spec.md plan.md` → `spec.md:0`, `plan.md:0` (the unresolvable coordinate is gone from both) |
| D5 | optional | yes | same block as D4 — the transcripts are attributed to the **primary checkout** and the installed `v3.1.2`, with the worktree's own `ls -d .moai/state/todo` failure quoted so the unrunnability is explicit | `ls -d .moai/state/todo` (in the worktree) → `No such file or directory`, exit `1` — the claim the block makes ; `grep -n "primary checkout" spec.md` → `111:` (the Checkout bullet), `328:`, `340:` |
| D6 | optional | yes | `acceptance.md` fixture convention — reworded to permit a named storage-surgery exception, naming AC-TAQ-004 and AC-TAQ-013 and stating for each why the CLI cannot produce the shape | `grep -n "named here. There are exactly two" acceptance.md` → `10:` |
| D7 | optional | yes | `plan.md` M1 — `newTodoHistoryCmd` fixed as a naming contract, with the vacuous-pass consequence for AC-TAQ-011 and the noisy-fail consequence for AC-TAQ-014 stated separately | `grep -n "Naming contract" plan.md` → `187:` |

---

## D1 — the premise was measured, and it is false

The iteration-1 ledger rejected the in-test two-build comparison because *"CI
checks out at `fetch-depth: 1` by default"*. Measured at this tree:

```
$ grep -n "fetch-depth" .github/workflows/ci.yml
129:          fetch-depth: 0    # job `test`               (declared line 114)
264:          fetch-depth: 0    # job `test-race`          (declared line 254)
382:          fetch-depth: 0    # job `test-integration`   (declared line 368)
431:          fetch-depth: 0    # job `lint`               (declared line 422)
486:          fetch-depth: 0    # job `build`              (declared line 463)
543:          fetch-depth: 0    # job `constitution-check` (declared line 533)
```

Seven `actions/checkout` steps exist. The six above cover **every job that
compiles or runs Go**; the seventh (line 51) is `detect`, a paths-filter job that
runs none. So the rejected alternative was viable here, and the ledger now says
so rather than implying an impossibility.

**The mechanism did not change** — the brief said it need not, and there was no
independent reason to change it. What changed is the reason: the adopted form is
now justified on a cost premise (a base-tree build inside the test puts a
`go build ./cmd/moai` on the critical path of `test` at `ci.yml:208` and
`test-race` at `ci.yml:287`, the two jobs whose command reaches
`./internal/cli`; not `test-integration`, scoped to
`./test/integration/harness/...` at `ci.yml:400`). That is a preference, and it
is recorded as one.

The shape is what matters for the record: an unverified premise argued **against**
an action. Nothing downstream contradicts a wrong "don't do it", which is why
this one survived a full iteration.

---

## D2 + D3 — AC-TAQ-011, and what it can and cannot promise

These closed together because they are the same clause. The escape and the
squash hazard were both reproduced before writing the repair, rather than
reasoned about.

**The modify path is real.** In a throwaway repository: commit the goldens,
commit the verb, then commit a rewrite of the golden bytes.
`git log --diff-filter=A --format=%H -- g/` returns the *same* commit before and
after the rewrite — `C` does not move for a modifying commit. Every existing
command in clause 1 still passed. `git diff --exit-code "$C" -- g/` exited `1`.
That command is now clause 1's last line.

**The squash hazard is real.** In a single-commit repository holding both the
goldens and the verb, `C` resolves to `HEAD` and
`git cat-file -e "$C:verb.go"` exits `0` — so `test "$C" !=` and `! cat-file`
both fail, permanently. The response therefore does two things rather than one:
`acceptance.md` scopes clause 1 to the pre-integration branch (where
M0-before-M1 is a property that can still be violated), and `plan.md` §D records
the merge-method dependency together with today's measured merge path
(`auto-merge.yml:262-269` selects `--merge` for `release/*` and `--squash` for
everything else; card branches reach `develop` by a local `--no-ff` per
CLAUDE.local.md §4.1 — both preserve M0 and M1 as distinct commits). Today's
path preserves the guard; the SPEC now says it depends on that rather than
assuming it.

**The prose no longer overclaims.** The sentence *"no capture taken from the
post-change tree can satisfy this AC"* is gone. What replaces it states the two
halves separately: clause 1 establishes provenance and integrity of the bytes;
it does **not** establish that the bytes came from running that tree's binary,
which is clause 2's job and clause 2 is a manual gate.

**Singleness is now asserted, not merely required in prose.** The prose said `C`
must be "a single commit" while the command was `… | tail -1`. A count assertion
precedes the resolution. Note `wc -l` pads with leading spaces on macOS, so the
form carries `| tr -d '[:space:]'` — measured, not assumed.

**The develop-drift residual is recorded**, with the part that is easy to get
wrong: when clause 3 reds because `develop` moved, regenerating the goldens
without first reading the diff converts an unrelated default-read growth into a
silently accepted one — the same failure this AC exists to catch, arriving from
the other side. Regenerating also re-opens clause 1, because a regenerating
commit is a *modify* and does not move `C`.

**One honesty correction inside the repair.** The failing-input table under
clause 1 initially presented every row as observed. Two rows were not: the
shallow-checkout `merge-base` failure and the `git grep -q` failure were reasoned,
not executed. The table now marks each row **measured** or *reasoned* and obliges
the run phase to observe the reasoned ones RED once rather than inherit them.

---

## D4 + D5 — the counters move, and they moved during the repair

These closed together because the fix is one block. `spec.md` §A.4 now carries a
`[MEASUREMENT PROVENANCE]` block naming the checkout (the **primary** checkout,
not the worktree — `ls -d .moai/state/todo` there exits `1`), the binary
(installed `v3.1.2`, main-era), and the instant of every figure, plus the
statement that these are live counters no requirement or AC turns on.

**112 vs 113 was never a population disagreement.** Snapshot B, every command
issued within one second at `2026-09-01T04:30:18+09:00`:

```
last_seq        402
count(*) items  113
by state        dropped|20  picked|22  queued|71     (sum 113)
.tables         findings  items  meta
moai todo | grep -cE '^\s*t[0-9a-z]+'   113
```

`list` and `count(*)` agree exactly. The plan-phase turn recorded `112` from one
and `113` from the other at different instants; each was correct when taken.

**The queue moved twice during this repair, and the second movement is the
SPEC's own subject.** Between the plan-phase snapshot and `04:25` it gained a
card (`401 → 402`, `113 → 114`). Between `04:25` and `04:30:18` it **lost** one
(`queued` `72 → 71`, `count(*)` `114 → 113`) while `last_seq` held — a `done`
under the deleting binary, destroying one more record while the SPEC documenting
the destruction was being written. The unaccounted count went `288 → 289`.

One condition is now stated that was previously implicit: `last_seq − count(*)`
equals the destroyed count **only while the queue has no archive tables**. After
one archive-capable `done`, the row leaves `items` and the subtraction
over-counts by the number of archived rows. REQ-TAQ-004 is not exposed to this
(it is a per-id predicate consulting both stores), but §A.4's arithmetic is.

Also stated where it belongs: in the stale-`backlog.json` block, the durable fact
is the **missing `archived` key**, not the `109` vs `113` gap — that gap widens
on its own, because the JSON is frozen while the database is not.

---

## D6 — the convention was right in spirit and false as written

`acceptance.md` claimed *"no test writes storage rows by hand"* while two of its
own criteria do. The convention now permits a named exception and names both,
with the reason each needs it: AC-TAQ-004 reconstructs the queue a pre-archive
`done` left (the current CLI archives instead of deleting), and AC-TAQ-013
reconstructs a pre-archive database (the DDL recreates the archive tables on
every open, `backlog_sqlite.go:92-94`). Both are artefacts of an *older* binary
that the current CLI cannot produce — which is what keeps the exception narrow.

---

## D7 — a symbol two ACs depend on, mandated by nothing

`plan.md` M1 now fixes `newTodoHistoryCmd` as a contract, with the two failure
directions separated because they are not equally dangerous: under a different
name AC-TAQ-011 clause 1's `git grep` matches nothing anywhere and passes
**vacuously** on a tree where the verb exists (silent), while AC-TAQ-014's
`test -n "$SRC"` fails for a naming reason (noisy). The convention is grounded
rather than invented — all sixteen registered `todo` verbs are built by
`newTodoXxxCmd()` constructors, re-derived at this tree:

```
$ sed -n '148,152p' internal/cli/todo.go | grep -oE 'newTodo[A-Za-z]*Cmd' | wc -l
16
```

The escape hatch is stated: if the run phase names it otherwise, the AC commands
change in the same commit, or match on the Cobra `Use` string instead.

---

## What this repair did NOT close

Three items, stated because two of them could be mistaken for closures.

1. **The citation-freshness gap is open, by construction.** Every `file:line` in
   these artifacts resolves at `2c18091d1`; **none was verified at the current
   `origin/develop` head**, which has moved well past it and is still moving.
   This is the iteration-2 report's own carry-forward note and the repair does
   not touch it. `plan.md` §C carries the run-phase obligation to re-run the
   pre-flight batch and re-verify every citation before the first edit; it is
   also recorded in `progress.md` §E.1.

2. **Clause 2 remains a manual gate.** The iteration-1 residual is unchanged: an
   implementer who skips the local regeneration and reports having run it is not
   caught by clause 1 alone. What changed is that clause 1 now closes the
   *modify* path, so the surviving hole is narrower — and the header prose now
   says which half clause 1 owns rather than implying it owns both.

3. **The N4 "survives verbatim" question stays unresolvable.** The iteration-2
   report could not establish whether §A.4's em-dash clause was lost or was
   quoted from `premise-check.md:199` and attributed to §A.4, because the
   artifacts are untracked and no prior version exists to diff. Nothing in this
   repair makes that decidable, and no attempt was made to restore a clause whose
   loss is unverified — restoring it would be acting on an unobserved premise.
   The substance the report asked not be regressed (§A.4 stating what cannot be
   recovered, before what can) is intact at `spec.md:84-95`.
