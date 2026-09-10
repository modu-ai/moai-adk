# SPEC Review Report: SPEC-TODO-ARCHIVE-QUERY-001

Iteration: 2/3
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.875** (Tier M threshold 0.80)

Reasoning context ignored per M1 Context Isolation. This audit reads the artifact
files and the tree they name; the iteration-1 report and the author's response are
read only as *claims to be re-verified*, never as evidence.

Tree audited: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t394`,
branch `WT-todo-done-history`, working tree `2c18091d1`. All six input artifacts
re-hashed at audit start; all six match the dispatch:

```
6258b71a…d90a8955d  spec.md
2affd439…9d1ce90b202  plan.md
e0865b4a…07df371fffe5f  acceptance.md
d30a2a69…c8bf69b2d7f  progress.md
9f14a8bb…91bcd6a3b95  premise-check.md
0d78ba26…5339fd4f96ce6  audit-response-iter1.md
```

The four SPEC artifacts and both reports are **untracked** (`git status --short`
→ `?? .moai/specs/SPEC-TODO-ARCHIVE-QUERY-001/`, `?? .moai/reports/t394/`), so no
prior committed version exists to diff against. That bounds one verification
below (N4).

This is a **full re-audit**, not a delta: every must-pass criterion was re-run and
all four dimensions re-scored, because the repair touched every artifact.

---

## Part 1 — Independent verification of the eight claimed closures

`audit-response-iter1.md` asserts 8 of 8 closed. Each cited command was re-run at
this tree. Result: **8 of 8 confirmed closed by their own commands.** Two closed
in a deviating form; both are judged on merit in Part 2.

| # | Claim | Command re-run | Observed | Verdict |
|---|---|---|---|---|
| D1 | §E inversion + unresolvable citation corrected on 3 surfaces | `grep -rn "func BacklogPathForRoot" internal/` | `state_dir.go:129`, `state_dir.go:137` — nothing else | **closed** |
| | | `grep -n 'const stateDirName\|const legacyStateDirName' internal/kanban/state_dir.go` | `37:const stateDirName = "todo"`, `43:const legacyStateDirName = "kanban"` | |
| | | `grep -rn "backlog_store.go:250" .moai/specs/SPEC-TODO-ARCHIVE-QUERY-001/` | 2 hits — `spec.md:312`, `plan.md:31`, both read in full and both **inside withdrawal paragraphs** that quote the retracted citation. No asserting use survives | |
| | | `grep -rn "state/kanban" …` | `spec.md:316` ("the **legacy** name"), `plan.md:30` ("was wrong and is withdrawn") | |
| | | `grep -n 'Gaps\|WITHDRAWN' premise-check.md` | `220:## Gaps`, `222:**[WITHDRAWN 2026-09-01, plan-audit iter1 D1]…`, `241` restates the inverse correctly | third surface confirmed |
| D2 | AC-TAQ-011 given a mechanical capture provenance; M0 added | `git log --diff-filter=A --format=%H -- internal/cli/testdata/golden/live-readers/` | **empty** (dir absent: `ls -d` → No such file or directory), so `test -n "$C"` fails | **closed, deviating** — see N1/N2/N3 |
| D3 | Tier S → M | `grep -n '^tier:' spec.md` → `14:tier: M`; `grep -cE '^\- \*\*REQ-TAQ-[0-9]{3}' spec.md` → `15`; `grep -c '^### AC-TAQ-' acceptance.md` → `15` | as claimed; `progress.md` §E.1:8-14 now cites the REQ/AC ceiling as the deciding axis | **closed** |
| D4 | REQ-TAQ-004 re-based on id-space accounting | `grep -n "archive is empty AND" spec.md` → exit 1, no match. REQ-TAQ-004 (`spec.md:176-186`) now keys on `last_seq`; AC-TAQ-004 (`acceptance.md:87-92`) carries the **inverted** regression clause ("that stderr note is **still present**") | the iter-1 assertion is genuinely reversed, not merely reworded | **closed, deviating** — see N5 |
| D5 | REQ-TAQ-013 extended to legacy JSON; §D gains the stale-sibling section | `grep -n '^### Out of Scope' spec.md` → `262:### Out of Scope — the stale sibling backlog.json` present, carrying the false-negative consequence at `spec.md:272-279`. `acceptance.md:247-256` adds the second Given/When/Then with a stated reachable failing input. `premise-check.md:257+` § Residual-risk carries "**Partly closed 2026-09-01**" | **closed** |
| D6 | AC-TAQ-014 located by symbol, existence asserted first | `git grep -l "newTodoHistoryCmd" -- 'internal/cli/*.go' ':!internal/cli/*_test.go'` → exit 1, empty, so `test -n "$SRC"` fails. Old form re-checked: `grep -rn … internal/cli/todo_history.go` → file absent, exit 2, which `!` inverts to PASS | the exact inversion is closed | **closed** |
| D7 | REQ-TAQ-012/014 in canonical `shall not` form | `grep -n "REQ-TAQ-012\|REQ-TAQ-014" spec.md` → `213: … The run-phase change shall not add a table…`, `224: … The verb shall not prompt the user` | **closed** |
| D8 | `list --json` added to REQ-TAQ-011 | `spec.md:210` → names `list`, `list --json`, `next`, `why`, `analyze` and the state counts, matching AC-TAQ-011 clause 3 (`acceptance.md:215-219`) | **closed** |

No defect claimed closed was found open. No false closure claim.

---

## Part 2 — The two deviations, judged on merit

### N1 (D2) — the rejection of the alternative rests on a premise this repository contradicts

`audit-response-iter1.md:44-48` rejects the in-test two-build comparison because
*"CI checks out at `fetch-depth: 1` by default, so the base object is frequently
absent and the test would fail for a reason unrelated to the invariant."*

Measured:

```
$ grep -rn "fetch-depth" .github/workflows/ci.yml
129:  fetch-depth: 0   # job `test`         (line 114)
264:  fetch-depth: 0   # job `test-race`    (line 254)
382:  fetch-depth: 0   # job `test-integration` (line 368)
431:  fetch-depth: 0   # job `lint`         (line 422)
486:  fetch-depth: 0   # job `build`        (line 463)
543:  fetch-depth: 0   # job `constitution-check` (line 533)
```

**Every Go-test job in this repository already checks out at `fetch-depth: 0`**,
with an in-file comment recording why. `fetch-depth: 1` is `actions/checkout`'s
default, not this repository's configuration — and the repository's configuration
is what decides whether the rejected alternative would have worked here. The sole
stated reason for rejecting alternative 1 is therefore an unverified premise
presented as a measured constraint, which is precisely the direction
`verification-claim-integrity.md` §1.1 surface 4 names as the dangerous one:
nothing downstream contradicts a wrong "don't do it" claim.

Two consequences, and they point opposite ways, so both are recorded:

- The deviation was **accepted on a false premise**. The alternative it displaced
  was viable here.
- The same measurement **rescues clause 1**: because CI has full history,
  `git log --diff-filter=A` resolves, and my initial concern that clause 1
  inherits the shallow-checkout fragility it was written to avoid does **not**
  bite in this repository's CI.

The adopted mechanism is not thereby wrong. The ledger entry is.

### N2 (D2) — clause 1 verifies the *introducing* commit, not the *current bytes*

`acceptance.md:173-174` claims: *"no capture taken from the post-change tree can
satisfy this AC."* That claim is false for one reachable path.

Clause 1 resolves `C` with `--diff-filter=A` — the commit that **added** the
goldens. A later commit that **modifies** the golden bytes is `M`, not `A`, so it
never moves `C`. An implementer who lands M0 honestly, breaks the default read in
M1, then edits the golden files to match the broken output, passes clause 1
(`C` still predates the verb), passes clause 3 (the bytes now agree), and the
BLOCKING guard for REQ-TDG-007 goes green while the invariant is broken.

This is the same self-certification AC-TAQ-011 exists to prevent, deferred by one
commit rather than eliminated. It is materially larger than the residual the
response records at `audit-response-iter1.md:73-77`, which names only "a dirty
working directory at capture" and is closed by clause 2. This one is not closed by
clause 2 either, since clause 2 is a manual Definition-of-Done gate.

The repair is still a real improvement — the original defect (capture-after-build
in the same commit) is now mechanically caught. What must change is the AC's own
overclaim, and ideally a clause asserting the goldens' bytes are unmodified since
`C` (`git log --format=%H -- <dir> | wc -l` = 1, or a diff of the working copy
against `C:<dir>`).

### N3 (D2) — clause 1 goes permanently red after a squash merge, and its "single commit" requirement is unasserted

Two mechanical properties of the adopted chain:

- `acceptance.md:181-183` requires `C` to be *"a single commit"*, but the command
  is `… | tail -1`, which silently selects the oldest of however many adding
  commits exist. The prose requirement is not asserted by the commands under it.
- If this branch is **squash-merged**, M0's golden commit and M1's verb commit
  collapse into one commit on the integration branch. `C` then resolves to a
  commit whose tree **does** contain the verb, so
  `! git cat-file -e "$C:internal/cli/todo_history.go"` fails — permanently, on
  every subsequent run, on the integration branch. A guard that is red forever
  after integration is a guard that gets deleted, which is the standard the
  response itself invokes at `audit-response-iter1.md:47-48`.

Whether that fires depends on the merge method used for this card; the SPEC does
not state a dependency on merge method, and it now has one.

Also unstated: the goldens are byte-comparisons against a moving `develop`. Any
unrelated develop change to `list` / `next` / `why` / `analyze` output reds
clause 3 for a reason unrelated to this change. That is arguably the point, but
it is an operational residual the artifacts do not record.

### N4 (D2) — "survives verbatim" could not be confirmed, and one clause is missing

`audit-response-iter1.md:112-114` states that §A.4's original paragraph *"survives
**verbatim**."* The iteration-1 report quotes §A.4 as: *"It does not, and cannot,
recover the harness-selection card or t81/t83/t88/t89 — **the two incidents that
motivated the card**."* Current `spec.md:91-92` reads: *"It does not, and cannot,
recover the harness-selection card or t81/t83/t88/t89. Stating this is part of the
deliverable…"* — the em-dash clause is gone.

```
$ grep -rn "two incidents that motivated" .moai/specs/… .moai/reports/t394/
.moai/reports/t394/premise-check.md:199
.moai/reports/t394/plan-audit-iter1.md:320
```

Because the artifacts are untracked, **no prior version exists to diff**, so I
cannot establish whether §A.4 changed or the iteration-1 auditor quoted
`premise-check.md:199` while attributing it to §A.4. I report this as a **gap, not
a defect**: the "verbatim" claim is unverifiable either way at this tree. The
substance of §A.4's honesty — that the surface cannot recover the motivating
cards, and that saying so is part of the deliverable — is intact
(`spec.md:84-94`), which is the thing the iteration-1 report asked not be
regressed.

### N5 (D4) — the id-space accounting counts the right population; the *narrative* around it does not

The brief asks whether `meta.last_seq` vs `count(*) FROM items` is actually
attributable to destroyed cards. Two separate objects must be distinguished, and
the artifacts conflate them in one place.

**The requirement is sound.** REQ-TAQ-004 (`spec.md:176-186`) is a **per-id
predicate**, not a difference of counts: *ordinal ≤ `last_seq`* **and** *held by
neither the live rows nor the archive*. Because it consults **both** stores, an
archived row is present and does not fire it; a `dropped` row is a live row
(`spec.md:69-82`) and does not fire it; a restored card is back in `items` and
does not fire it. Re-issue is excluded at source —
`internal/kanban/backlog_store.go:14-17` (re-read, resolves): *"The mark — never
max-present-id — decides the next id, because `done` removes rows and a derived
mark would reuse the removed card's id"* — and
`backlog_store.go:772-778` (re-read, resolves) only ever raises the mark. The
residual the SPEC discloses at `spec.md:124-129` is the correct and only one:
`normalizeBacklogRecord` raises the mark to cover ids the record *holds*, so a
hand-edited or imported mark can leave a gap that was never issued. The
requirement is worded as a **qualification** of `absent`, not a claim of
issuance. That is the right shape.

**The narrative arithmetic is population-correct only at its measurement instant.**
`spec.md:115` — *"401 ids issued, 113 live rows … 288 ids issued and now held by
neither store"* — is valid **only because that queue has no archive tables**
(`.tables` → `findings items meta`). Once an archive-capable binary runs one
`done`, `count(*) FROM items` drops while `last_seq` holds, and the same
subtraction over-counts destroyed cards by exactly the number of archived rows.
Nothing in §A.4 says the subtraction is conditional on the archive's absence, so a
later reader re-deriving it gets a wrong number. This does not touch the
requirement — it touches the paragraph that motivates it.

**Deviation verdict**: the deviation from "a durable marker written at first
archive-capable open" is well-founded. The two reasons given are verified:
`backlog_store.go:406-411` does document `QueuedCount` as a PURE read on
*"surfaces that must never move operator bytes"* (re-read, quoted exactly), and a
marker written at open would make an open into a write against REQ-TAQ-010. The
per-id predicate is strictly more informative than a per-queue marker. Sound.

---

## Part 3 — the live counters (brief item 3)

The figures appear at: `spec.md:108,110,115,255,269,270`; `plan.md:19,42,43,48,49`.

**No AC and no REQ turns on any of them.** AC-TAQ-004's fixture uses
`last_seq = 5` (`acceptance.md:78`), a constructed value, not the operator's. So
the moving-ref hazard does not reach the verification layer. That is the load-bearing
finding and it is favourable.

Two problems remain in the context layer:

1. **Two mutually inconsistent values for the same population, both labelled
   "measured", unreconciled.** `spec.md:255` and `plan.md:48` say the live queue
   measured **112** cards; `spec.md:110,115,270` and `plan.md:19,43` say **113**
   live rows. `list` applies no state filter (`internal/cli/todo.go:310-311`,
   re-read, resolves), so both name the same set. The orchestrator re-measured
   **114** minutes later. Nothing in the artifacts notes that these are snapshots
   of a counter that moves, and `plan.md:48-51` — which exists precisely to warn
   that the dispatch's figures did not reproduce — is itself one of the two
   inconsistent surfaces.
2. **No measurement coordinate.** `spec.md:104` and `plan.md:19` say *"Measured
   this turn"*. "This turn" is not resolvable by a later reader, and the tree SHA
   `2c18091d1` cannot pin it because the queue is runtime state that lives in no
   tree.

Judgment: **illustration, not requirement — so SHOULD-FIX, not blocking.** The fix
is a timestamp plus a one-line note that these are live counters, and picking one
value for the live-row count.

## Part 4 — scoping of the runtime measurements (brief item 4)

Confirmed: `ls -d .moai/state/todo` in this worktree → **No such file or
directory**. The `sqlite3` and `ls` measurements are from the primary checkout,
served by the installed main-era `v3.1.2` binary.

**Attribution is partial.** `spec.md:86-89` does name the binary and its era
explicitly — *"the installed binary (`v3.1.2`, main-era) deletes on `done`
(`git show origin/main:internal/cli/todo.go`, measured)"* — and sources the
delete behaviour to `origin/main`, which is the correct tree for that claim. That
is the important half and it is right.

**What is missing is the checkout.** No artifact says these commands were run in
the primary checkout rather than in the tree the SPEC names. `spec.md:106-113`
presents them as a bare shell transcript against a relative path
(`.moai/state/todo/backlog.db`) inside a document whose §A.1 opens *"At
`origin/develop@2c18091d1`…"*. A reader who runs them where the SPEC says it was
written gets a file-not-found. Same for `spec.md:265-270` and `plan.md:39-44`.

**No develop behavioural claim rests on them.** I checked the one that could:
REQ-TAQ-013's second clause rests on *legacy JSON carries no `archived` key*,
which is a claim about a **pre-archive artifact's shape** (correctly main-era),
while the develop-side read path it obliges is separately cited to
`internal/kanban/backlog_store.go:551-568` (re-read, `LoadPure` →
`loadLegacyBacklogJSON` when no db — resolves). The layering is correct.

## Part 5 — §A.4 honesty and M0 (brief item 5)

- **§A.4's honesty survived.** `spec.md:84-94` still states the surface cannot
  recover the destroyed cards, that this SPEC closes the gap *"going forward
  only"*, and that *"a surface that silently reports `absent` for a destroyed card
  teaches the operator a false negative."* The additions (lines 96-137) sit after
  it and **narrow what `absent` means**; they do not soften what is unrecoverable.
  One clause was lost — see N4.
- **M0 smuggles no implementation into plan phase.** `git status --short` shows
  only the two untracked artifact directories; `ls internal/cli/todo_history.go`
  and `ls -d internal/cli/testdata/golden/live-readers/` both → No such file;
  `git grep -l newTodoHistoryCmd` → empty; `progress.md:39` — *"Status: `draft`.
  No implementation code, tests, commits, or pushes exist."* M0 is a described
  run-phase milestone, which is what `plan.md` §F is for.

## Part 6 — citation integrity (brief item 6)

**16 `file:line` citations opened at the tree they name; 16 resolved.** Weighted
toward repair-introduced ones:

| Citation | Where cited | Resolves to |
|---|---|---|
| `backlog_store.go:14-17` | spec.md:98, response:96 | the `last_seq` never-derived comment, quoted accurately |
| `backlog_store.go:223` | spec.md:36 | `func (r *BacklogRecord) ArchiveCard` |
| `backlog_store.go:406-411` | response:87-89 | the PURE-read comment, quoted **verbatim and correctly** |
| `backlog_store.go:411-427` | spec.md:274, plan.md:43 | `QueuedCount`'s layout branch |
| `backlog_store.go:551-568` | spec.md:275, plan.md:43 | `LoadPure` incl. the legacy-JSON branch |
| `backlog_store.go:551-561` | acceptance.md:249 | same function, narrower span — resolves |
| `backlog_store.go:772-778` | spec.md:101 | `normalizeBacklogRecord`'s raise-only logic |
| `backlog_sqlite.go:92-94` | spec.md:132, plan.md:26 | the `IF NOT EXISTS` comment, quoted accurately |
| `backlog_sqlite.go:95-101` | spec.md:249 | the CHECK-rebuild rationale |
| `backlog_sqlite.go:113,124,133` | plan.md:73 | the `state` CHECK; `archived_items`; `archived_findings` |
| `backlog_migrate.go:102` | spec.md:44, plan.md:159 | `e.readArchive(ctx, rec)` |
| `backlog_migrate.go:106-110` | plan.md:148, response:97 | `readLastSeq` → `rec.LastSeq` |
| `backlog_migrate.go:411` | plan.md:75 | `func loadLegacyBacklogJSON` |
| `state_dir.go:37,43,79-104,129` | spec.md:314-319, plan.md:32-36 | consts, `resolveStateDir`, `BacklogPathForRoot` |
| `todo.go:148-152 / 310-311 / 409` | spec.md:35,49,71 | `AddCommand` (**16 constructors — count re-derived and correct**); the unfiltered `for _, it := range rec.Items` render; `return rec.ArchiveCard(id)` |
| `todo_why.go:34-37`, `todo_export.go:100-110`, `todo_undone.go:68`, `backlog_archive_test.go:98`, `todo.md` line 51 | spec.md:66,41; plan.md:77; acceptance.md:243; plan.md:124 | the `no findings` conflation; `discloseArchiveDowngradeCost`; `RestoreCard`; the `DROP TABLE archived_*` fixture; the `pr` text-last convention |

The D1 class of defect — a citation written without opening it — does not recur.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -oE 'REQ-TAQ-[0-9]{3}' spec.md |
  sort -u` → `REQ-TAQ-001` … `REQ-TAQ-015`; definition count
  (`grep -cE '^\- \*\*REQ-TAQ-[0-9]{3}'`) = `15`; no duplicate, no gap, uniform
  3-digit padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the `REQ-XXX`
  requirement layer in `spec.md` §B only; `acceptance.md`'s Given-When-Then
  entries are the verification layer and are graded under Group 4. All 15 now
  match a canonical pattern: event-driven (`REQ-TAQ-001/002/003`), compound
  Where+When (`REQ-TAQ-004`), Where (`REQ-TAQ-013`), ubiquitous
  (`REQ-TAQ-005/006/007/008/009/011/015`), unwanted in canonical
  `shall not` form (`REQ-TAQ-010`, and `012`/`014` — the iter-1 D7 deviation,
  now closed at `spec.md:213,224`).
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with
  correct types: `id`, `title`, `version: "0.2.0"` (quoted semver), `status: draft`,
  `created: 2026-09-01`, `updated: 2026-09-01`, `author`, `priority: P2`, `phase`,
  `module`, `lifecycle: spec-anchored`, `tags` (comma-separated string). No
  rejected snake_case alias. Optional `tier: M` and `related_specs` also present.
- **[N/A] MP-4 language neutrality** — single-programming-language (Go) SPEC.
  The distinct template-neutrality axis it does carry (`REQ-TAQ-015`,
  `AC-TAQ-015`) has a reachable failing input and a baseline of 0 matches.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — four references extracted; all
  four resolve; none retired / superseded / archived:
  `SPEC-TODO-DESTRUCTIVE-GUARD-001` `completed`, `SPEC-TODO-SQLITE-001`
  `completed`, `SPEC-TODO-ANALYSIS-001` `completed`, `SPEC-KANBAN-TODO-CLI-001`
  `in-progress`. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → `0`.
  Auto-pass.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'` over the
  SPEC directory returns only a hit inside `plan-audit-iter1.md`'s own quotation
  of the check. `plan.md` carries no marker; no `research.md` exists (Tier M set).

**All must-pass criteria pass.** The firewall does not fire.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75–1.0 | All 15 requirements have a single unambiguous reading; `spec.md:69-82` fixes the four-fates vocabulary explicitly so `dropped` cannot be read as archived. Deducted for D4 (112 vs 113 unreconciled, `spec.md:255` vs `:115`) and D5 (`spec.md:106-113` reads as runnable from the tree the document names and is not). |
| Completeness | 0.85 | 0.75–1.0 | All sections present; frontmatter complete; five `### Out of Scope — <topic>` headings each with specific bullets (`spec.md:243,252,262,281,289`). Deducted for D2 (`acceptance.md:173-174` overclaims what clause 1 forecloses) and D3 (the merge-method and develop-drift dependencies of AC-TAQ-011 are unrecorded). |
| Testability | 0.85 | 0.75–1.0 | Every AC binary and names the command deciding it; no weasel words. AC-TAQ-004, 011 and 013 each state a *reachable failing input*, and AC-TAQ-014 mandates a RED plant-and-remove because the negated grep is near-baseline-vacuous (`acceptance.md:282-286`) — that is better than most SPECs manage. Deducted for D2/D3 (clause 1's "single commit" is prose, not assertion; the modify-path escapes it) and D6 (AC-TAQ-004's seeding contradicts the file's own fixture convention). |
| Traceability | 0.95 | 0.75–1.0 | 15 REQ ↔ 15 AC, strictly 1:1, matrix at `acceptance.md:14-30`; no orphan AC, no uncovered REQ (verified by id-set comparison). Deducted only for D7: the symbol `newTodoHistoryCmd` is load-bearing for AC-TAQ-011 clause 1 and AC-TAQ-014 but is mandated by no requirement. |

**Aggregate: 0.875** (arithmetic and harmonic mean agree to three digits).
Threshold for Tier M: **0.80**. Score-delta from iteration 1: **0.775 → 0.875**,
monotonic improvement, no regression detected in any dimension.

---

## Defects Found

**D1. The D2 deviation was accepted on a premise this repository's CI contradicts** — `.moai/reports/t394/audit-response-iter1.md:44-48` — Severity: **SHOULD-FIX** — Class: **blocking**

The stated reason for rejecting the in-test two-build comparison is that CI checks
out at `fetch-depth: 1`. Every Go-test job in `.github/workflows/ci.yml` sets
`fetch-depth: 0` (lines 129, 264, 382, 431, 486, 543), with an in-file comment
recording why. The rejected alternative was viable here. Under
`verification-claim-integrity.md` §1.1 surface 4 this is an unverified premise
dressed as a reason — and it argues *against* an action, the direction nothing
downstream contradicts.

**Required fix**: correct the ledger entry to state the repository's actual
checkout depth, and either re-justify the deviation on a premise that holds or
record that alternative 1 was viable and the adopted form was preferred for a
different, stated reason. The adopted mechanism need not change.

---

**D2. AC-TAQ-011 clause 1 does not foreclose what its own prose claims it forecloses** — `.moai/specs/SPEC-TODO-ARCHIVE-QUERY-001/acceptance.md:173-174, 180-192` — Severity: **SHOULD-FIX** — Class: **blocking**

`--diff-filter=A` pins the commit that *added* the goldens. A later commit that
*modifies* the golden bytes is `M` and never moves `C`. So: land M0 honestly,
break the default read in M1, edit the goldens to match — clause 1 passes, clause
3 passes, and the SPEC's only BLOCKING guard for REQ-TDG-007 is green over a
broken invariant. Clause 2 does not close this, being a manual Definition-of-Done
gate. The claim *"no capture taken from the post-change tree can satisfy this AC"*
is therefore false for that path.

**Required fix**: add a clause asserting the goldens' current bytes still match
`C`'s tree — e.g. `git diff --exit-code "$C" -- internal/cli/testdata/golden/live-readers/`
— and soften the prose claim to what clause 1 actually establishes.

---

**D3. AC-TAQ-011 clause 1 has an unrecorded dependency on merge method, and its "single commit" requirement is unasserted** — `.moai/specs/SPEC-TODO-ARCHIVE-QUERY-001/acceptance.md:181-192` — Severity: **SHOULD-FIX** — Class: **blocking**

(a) The prose requires `C` to be *"a single commit"*; the command is `… | tail -1`,
which silently picks the oldest of many. (b) Under a squash merge, M0 and M1
collapse into one commit whose tree contains the verb, so
`! git cat-file -e "$C:internal/cli/todo_history.go"` fails permanently on the
integration branch — a guard red forever is a guard deleted, the exact standard
`audit-response-iter1.md:47-48` invokes. (c) Unstated: clause 3 compares against a
golden captured from an ancestor of a moving `develop`, so any unrelated change to
`list`/`next`/`why`/`analyze` output reds it.

**Required fix**: assert singleness (`test "$(git log --diff-filter=A --format=%H -- <dir> | wc -l)" -eq 1`);
state the merge-method dependency in `plan.md` §D or scope clause 1 to the
pre-integration branch; record the develop-drift residual.

---

**D4. Live counters cited as fixed measurements, with two inconsistent values for the same population** — `spec.md:110,115,255,270`; `plan.md:19,43,48` — Severity: **SHOULD-FIX** — Class: **blocking**

`spec.md:255` / `plan.md:48` say **112** live cards; `spec.md:110,115,270` /
`plan.md:19,43` say **113** live rows. `list` applies no state filter
(`internal/cli/todo.go:310-311`), so both name the same set. Re-measured by the
orchestrator minutes later: **114** (and `last_seq` **402**). Nothing marks these
as snapshots of a moving counter, and the only coordinate offered is *"this
turn"*, which no later reader can resolve — `2c18091d1` cannot pin it, because
the queue is runtime state that lives in no tree. `plan.md:48-51`, whose whole
purpose is to warn that the dispatch's figures did not reproduce, is itself one of
the two inconsistent surfaces.

Mitigating and load-bearing: **no AC and no REQ turns on any of these figures** —
AC-TAQ-004 uses a constructed `last_seq = 5`. The hazard does not reach the
verification layer, which is why this is SHOULD-FIX and not BLOCKING.

**Required fix**: reconcile 112/113 to one value, and label each figure with its
measurement instant plus a note that it is a live counter.

---

**D5. The runtime measurements are not attributed to the checkout that produced them** — `spec.md:104-113, 265-270`; `plan.md:17-21, 39-44` — Severity: **MINOR** — Class: **optional**

`ls -d .moai/state/todo` in the worktree the artifacts name → *No such file or
directory*. The `sqlite3` and `ls` transcripts come from the primary checkout,
served by the installed main-era `v3.1.2` binary. `spec.md:86-89` correctly names
the binary and its era and sources the delete behaviour to
`git show origin/main:internal/cli/todo.go`; what is missing is the **checkout**.
Presented as a bare transcript against a relative path inside a document that
opens *"At `origin/develop@2c18091d1`…"*, it reads as runnable there and is not.
No develop behavioural claim rests on it (checked: REQ-TAQ-013's second clause
separates the main-era *artifact shape* from the develop-side *read path* it cites
at `backlog_store.go:551-568`).

**Required fix**: prefix the transcripts with the checkout they ran in and the
binary that served them.

---

**D6. AC-TAQ-004's fixture contradicts the file's own stated fixture convention** — `acceptance.md:6-8` vs `acceptance.md:78-81` — Severity: **MINOR** — Class: **optional**

Line 6-8: *"a queue seeded through the public CLI … so no test writes storage rows
by hand."* Lines 78-81: *"seeded by adding five cards through the CLI and removing
`t3`'s row from `items` without lowering `last_seq`"* — which is a hand-written
storage mutation. AC-TAQ-013 (`acceptance.md:240-242`) likewise drops tables. The
convention is right in spirit; as absolutely worded it is now false for two of
its own criteria.

**Required fix**: reword the convention to permit an explicitly-named
storage-surgery exception for criteria that must construct a shape the CLI cannot
produce, and name AC-TAQ-004 and AC-TAQ-013 as the two.

---

**D7. A symbol name is load-bearing for two acceptance criteria but is mandated by no requirement** — `acceptance.md:191, 269`; `spec.md` §B — Severity: **MINOR** — Class: **optional**

`newTodoHistoryCmd` decides AC-TAQ-011 clause 1's last check and AC-TAQ-014's
subject location. If the run phase names the constructor anything else, clause 1's
`git grep` passes vacuously on a tree where the verb exists, and AC-TAQ-014's
`test -n "$SRC"` fails for a naming reason. The convention is well-founded —
`internal/cli/todo.go:148-152` registers all sixteen existing verbs through
`newTodoXxxCmd()` constructors — but a convention two ACs depend on should be
written down.

**Required fix**: state the constructor name in `plan.md` M1 as a contract, or
match on the command's `Use` string instead of the Go symbol.

---

## Regression Check (defects from iteration 1)

- **D1** (§E factual inversion, unresolvable citation) — **RESOLVED**: all three
  surfaces corrected; the only two surviving mentions of the retracted citation
  are inside withdrawal paragraphs; the direction is now correct against
  `state_dir.go:37,43`.
- **D2** (AC-TAQ-011 self-certifying) — **RESOLVED IN PART**: the original hole
  (same-commit capture-after-build) is now mechanically caught. A narrower hole
  survives — see new D2/D3. Not a stagnation: the defect changed shape and shrank.
- **D3** (Tier S refuted) — **RESOLVED**: `tier: M`; `progress.md` §E.1 cites the
  REQ/AC ceiling as the deciding axis and states the threshold consequence
  plainly.
- **D4** (disclosure self-disables after one `done`) — **RESOLVED**: re-based on
  the per-id `last_seq` predicate; AC-TAQ-004's second clause is inverted into a
  regression guard; `plan.md` §G carries the matching anti-pattern.
- **D5** (stale sibling `backlog.json` absent) — **RESOLVED**: REQ-TAQ-013
  extended, `### Out of Scope — the stale sibling backlog.json` added with the
  false-negative consequence, AC-TAQ-013 gains a second scenario with a stated
  reachable failing input.
- **D6** (negated grep passes on a missing subject) — **RESOLVED**: subject
  located by symbol, existence asserted first, RED plant-and-remove mandated.
- **D7, D8** (optional) — **RESOLVED**.

No defect appears unchanged across both iterations. No stagnation.

---

## Recommendation

**PASS-WITH-DEBT at iteration 2.** All seven must-pass criteria pass, every
iteration-1 defect is closed under its own command, and the aggregate 0.875 clears
the Tier M threshold of 0.80. The SPEC may enter the Implementation Kickoff
Approval gate.

Three things this SPEC does unusually well, recorded so a later revision does not
damage them: §A.4 states what the surface *cannot* do before stating what it can;
AC-TAQ-004, 011, 013 and 014 each name a *reachable failing input* rather than
asserting non-vacuity; and the repair inverted AC-TAQ-004's second clause into a
regression guard against its own prior defect instead of merely deleting it.

**Debt carried into run phase**, in priority order:

1. **D2 + D3 — AC-TAQ-011's remaining escape and its merge-method dependency.**
   This is the SPEC's only BLOCKING guard for REQ-TDG-007. Fixing the modify-path
   hole is one added command; the merge-method dependency needs a decision, not
   code. Address both before M0 is committed, since M0's commit is what `C` binds
   to.
2. **D1 — correct the deviation ledger.** No artifact changes; the entry is
   wrong about this repository's CI and a future reader will act on it.
3. **D4 — reconcile 112/113 and timestamp the live counters.** Cheap, and it is
   the class of defect this project has been bitten by repeatedly.
4. **D5, D6, D7 — MINOR/optional.** Routing them into a third iteration is not
   worth the round trip; fold them into the run phase's first artifact touch.

Per the Retry Loop Contract, Tier M's ceiling is 2 iterations. This is iteration 2
and it passes, so no iteration 3 is required or authorised. The debt above is
recorded for the run phase, not for another audit round.

Note carried forward for the run phase: `origin/develop` has moved well past
`2c18091d1` and is still moving. `plan.md` §C already carries the obligation to
re-run the pre-flight batch and re-verify every `file:line` at the then-current
head. All 16 citations spot-checked in this audit resolve at `2c18091d1`; none was
verified at the current develop head, and that is a gap by construction, not an
oversight.
