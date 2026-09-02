# SPEC Review Report: SPEC-TODO-ARCHIVE-QUERY-001

Iteration: 1/3
Verdict: **FAIL**
Overall Score: **0.775** (threshold 0.80 — see D3: the declared Tier is wrong)

Reasoning context ignored per M1 Context Isolation. This audit reads only the
artifact files and the tree they were written against.

Tree audited: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t394`,
branch `WT-todo-done-history`, HEAD `2c18091d1`. Note: `git rev-list --count
--left-right origin/develop...HEAD` now reads `21 0` — `origin/develop` has
advanced 21 commits since the SPEC was written against `2c18091d1`. Every source
citation below was re-read at `2c18091d1` (this worktree's working tree), which
is the tree the SPEC names, so the drift does not invalidate the citations; it
does mean a run-phase pre-flight must re-run plan.md §C.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -oE 'REQ-TAQ-[0-9]{3}' spec.md
  | sort -u` yields exactly `REQ-TAQ-001` … `REQ-TAQ-015`, 15 ids, sequential,
  no gaps, no duplicates (`grep -oE '^\- \*\*REQ-TAQ-[0-9]{3}' | sort | uniq -d`
  returns empty), consistent 3-digit padding. Definition count = 15.

- **[PASS] MP-2 GEARS format compliance** — judged against the `REQ-XXX`
  requirement layer in `spec.md` §B only; the Given-When-Then entries in
  `acceptance.md` are the verification layer and are graded under Group 4, not
  here. 13 of 15 match a canonical pattern exactly: event-driven (`REQ-TAQ-001`
  "**When** an operator runs `moai todo history <id>` … the CLI shall print"),
  compound Where+When (`REQ-TAQ-004`), Where (`REQ-TAQ-013`), ubiquitous
  (`REQ-TAQ-005`, `007`, `011`, `015`), unwanted (`REQ-TAQ-010` "shall take no
  write lock, shall not call `Mutate`"). Two use a nominal negative subject
  rather than `The <subject> shall not`: `REQ-TAQ-012` ("No storage change shall
  be made") and `REQ-TAQ-014` ("Nothing in this verb shall prompt the user").
  Each still carries `shall` and one unambiguous behavior, so this is a MINOR
  format deviation (D7), not a failure.

- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present
  with correct types: `id`, `title`, `version: "0.1.0"` (quoted semver),
  `status: draft`, `created: 2026-09-01`, `updated: 2026-09-01`, `author`,
  `priority: P2`, `phase`, `module`, `lifecycle: spec-anchored`, `tags`
  (comma-separated string). No rejected snake_case alias (`created_at` /
  `updated_at` / `labels` / `spec_id`) appears. Optional `tier` and
  `related_specs` additionally present.

- **[N/A] MP-4 Section 22 language neutrality** — single-programming-language
  (Go) SPEC; auto-passes. Note the distinct axis it *does* carry: `REQ-TAQ-015`
  imposes template neutrality on the mirrored skill document. Baseline measured:
  `grep -cnE "SPEC-…|REQ-…|20[0-9]{2}-…|\b[0-9a-f]{9,40}\b"
  internal/template/templates/.claude/skills/moai/workflows/todo.md` → `0`. The
  guard currently passes and has a reachable failing input.

- **[PASS] MP-5 D7 cross-SPEC reconciliation** — four SPEC references extracted
  from the body. All four resolve, none is retired / superseded / archived:
  `SPEC-TODO-DESTRUCTIVE-GUARD-001` `status: completed`,
  `SPEC-TODO-SQLITE-001` `status: completed`,
  `SPEC-TODO-ANALYSIS-001` `status: completed`,
  `SPEC-KANBAN-TODO-CLI-001` `status: in-progress`. No BLOCKING finding.

- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` →
  `0`. Auto-pass; no cross-platform concern.

- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'
  .moai/specs/SPEC-TODO-ARCHIVE-QUERY-001/` → exit 1, no matches. No
  `research.md` exists (correct for a non-L Tier).

No must-pass criterion fails. The FAIL verdict rests on the aggregate score
against the correct Tier threshold and on the three blocking defects below.

---

## Independent re-verification of `premise-check.md`

The dispatch required measurements 1, 2, 6 and 9 to be re-run rather than
trusted. All four reproduce, verbatim, at `2c18091d1`:

| # | Claim | Re-run result |
|---|---|---|
| 1 | `done` archives | `internal/cli/todo.go:409` → `return rec.ArchiveCard(id)`; `internal/kanban/backlog_store.go:223` → `func (r *BacklogRecord) ArchiveCard(id string) error`. Reproduces. |
| 2 | archive tables in DDL | `backlog_sqlite.go:113` live `CHECK (state IN ('queued','picked','dropped'))`, `:124` `archived_items`, `:133` `archived_findings`. Reproduces. |
| 6 | 16 verbs, none render the archive; archive loaded on every read | `todo.go:148-152` registers exactly 16 (`add, list, done, undone, next, unpick, edit, move, drop, undrop, analyze, relate, unrelate, why, pr, export-json`); `backlog_migrate.go:102` `if err := e.readArchive(ctx, rec); err != nil`. Reproduces. |
| 9 | operator queue has no archive; installed binary deletes | `sqlite3 …/.moai/state/todo/backlog.db ".tables"` → `findings items meta`; `SELECT count(*) FROM archived_items` → `Error: … no such table: archived_items`; `~/go/bin/moai version` → `moai-adk v3.1.2`; `git show origin/main:internal/cli/todo.go` `newTodoDoneCmd` body → `rec.Items = append(rec.Items[:i], rec.Items[i+1:]...)` + `rec.RemoveFindingsNaming(id)`. Reproduces in full. |

Additional spot-checks, all resolving: `todo_undone.go:68` (`RestoreCard`),
`todo_export.go:100-110` (`discloseArchiveDowngradeCost`), `todo_why.go:34-37`
(the `no findings` conflation), `todo.go:310-311` (`list` with no state filter),
`backlog_sqlite.go:95-101` (the CHECK-rebuild rationale),
`backlog_archive_test.go:98` (`DROP TABLE archived_items`),
`.claude/skills/moai/workflows/todo.md` line 51 (the `pr` row, and its
"card text stays the LAST field" convention that plan.md M1 leans on),
`internal/cli/todo.go:583-589` (`normalizeTodoRef`, which makes `REQ-TAQ-005`
implementable). Every REQ-TDG token cited in §E resolves in
`SPEC-TODO-DESTRUCTIVE-GUARD-001/spec.md` (003:166, 007:176, 013:188, 015:193,
016:194).

**One measurement does NOT reproduce, and it is not a minor one — see D1.**

Live-queue counts have drifted since measurement (expected; the queue is
mutable): now `queued 73 / picked 21 / dropped 19`, total 113, vs the report's
`71 / 22 / 19`, total 112. This is drift, not error — the report attributes the
figure to a time and the argument it supports (default `list` output is large) is
unaffected. Not a defect.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.70 | between 0.50 and 0.75 | Prose is unusually precise — the four-fates table (`spec.md` §A.3), the output shape (`plan.md` M1), the disclosure conditions are each unambiguous. Marked down because §E states a measured fact about develop that is its inverse (D1), and `REQ-TAQ-004` reads as covering a hazard it covers for one `done` (D4). |
| Completeness | 0.80 | 0.75 band | All sections present; four `### Out of Scope — <topic>` H3 sub-headings each with specific `-` bullets; frontmatter complete. Marked down for two substantive omissions: §A.4 never states that the DDL's `IF NOT EXISTS` creates the archive on first open by a newer binary (`backlog_sqlite.go:95-97`), which is what bounds `REQ-TAQ-004`; and the legacy-JSON read path, named in the premise-check's own Residual-risk, is absorbed into no requirement or scope line (D5). |
| Testability | 0.70 | between 0.50 and 0.75 | Every AC names a deciding command; no weasel words anywhere. Marked down because two of fifteen can pass vacuously, and one of them is the AC the SPEC itself grades BLOCKING (D2, D6). |
| Traceability | 0.90 | between 0.75 and 1.0 | `REQ-TAQ-001..015` ↔ `AC-TAQ-001..015`, exactly 1:1, matrix at the head of `acceptance.md`, per-AC severity in §D.5, no orphan and no uncovered REQ (measured). Marked down for one drifted citation (D1). |

**Aggregate: (0.70 + 0.80 + 0.70 + 0.90) / 4 = 0.775.**

Threshold applied: **0.80 (Tier M)**, not 0.75 (declared Tier S) — see D3. At the
correct threshold, 0.775 is a FAIL. This is not an arithmetic technicality: the
Tier declaration is refuted on two independent measured axes, and it was resolved
in the direction that lowers the SPEC's own bar.

---

## Defects Found

**D1. §E dependency note asserts the inverse of what develop does, on a citation that does not resolve** — `spec.md` §E ("Dependency note — card t395"), `plan.md` §B.2, `premise-check.md` § Gaps — Severity: **critical** — Class: **blocking**

The SPEC states: *"develop's `BacklogPathForRoot` (`internal/kanban/backlog_store.go:250`) returns `.moai/state/kanban/backlog.json`"*, and builds a coupling to card t395 on it. Three things are wrong.

1. **The citation does not resolve.** `BacklogPathForRoot` is at
   `internal/kanban/state_dir.go:129`, not `backlog_store.go:250`. Line 250 of
   `backlog_store.go` is the tail of `ArchiveCard` and the doc comment of
   `RestoreCard`. Wrong file, wrong line.
2. **The returned value is the opposite of the claim.**
   `state_dir.go:37` — `const stateDirName = "todo"`;
   `state_dir.go:42` — `const legacyStateDirName = "kanban"`.
   `BacklogPathForRoot` joins `stateDirName`, so develop returns
   `.moai/state/**todo**/backlog.json` — exactly where the operator's live queue
   already sits. `.moai/state/kanban` is the **legacy** name the rename moved
   *away from*, and `ls -d /Users/goos/MoAI/moai-adk-go/.moai/state/kanban` →
   *No such file or directory*. Confirmed independently by
   `factory_slots_test.go:131` (`want := filepath.Join(root, ".moai", "state",
   "todo", "backlog.json")`) and `backlog_sqlite_test.go:273`
   (`.../state/todo/backlog.json` → `.../state/todo/backlog.db`).
3. **t395's scope is mischaracterized.** The queue card's text (read from the
   live queue) is entirely about the migration leaving a stale original
   `backlog.json` beside `backlog.db` **inside `.moai/state/todo/`** — three
   coexisting files, and a lane misled by reading the stale copy. It contains no
   path change.

Consequence: the SPEC records, as a measured fact, a migration risk that does not
exist ("If it starts empty, this verb will report `absent` for every existing
card"), and hands a run-phase implementer a false coupling to look for. Under the
project's own verification-claim doctrine this is an unverified premise presented
as a reason — the more dangerous direction, because nothing downstream contradicts
it.

**Required fix**: correct the citation to `internal/kanban/state_dir.go:129`,
correct the direction (develop's canonical dir is `.moai/state/todo`; `kanban` is
legacy), and either delete the migration-coupling paragraph or restate it as what
t395 actually owns — the stale sibling `backlog.json`, which *is* relevant to this
SPEC (see D5) but for a different reason.

---

**D2. AC-TAQ-011, the SPEC's own BLOCKING invariant guard, is self-certifying as written** — `acceptance.md` § AC-TAQ-011 — Severity: **critical** — Class: **blocking**

The AC reads: *"each output is byte-identical to a golden captured from the same
fixture **before this change**"*, and then: *"The golden files are committed with
the run-phase change"*. No mechanism pins the tree the goldens are captured at.
An implementer who builds the change and then captures the goldens records the
post-change output — including a broken one — and the test passes. The guard for
`REQ-TDG-007`, the invariant this whole SPEC promises not to disturb, can
therefore go green while the invariant is broken. §D.5 grades this AC BLOCKING,
which is the right grade and makes the vacuity worse, not better.

**Required fix**: name the capture tree mechanically. Either capture the goldens
at `origin/develop@<sha>` and commit them in a *separate, earlier* commit whose
SHA the AC cites, or replace byte-identity-to-a-golden with byte-identity between
two builds (pre-change binary vs post-change binary, both run against the same
fixture in the same test), so the comparison cannot be satisfied by one side alone.

---

**D3. Declared Tier S is refuted on two measured axes, and the error lowers the SPEC's own PASS threshold** — `spec.md` frontmatter `tier: S`; `progress.md` §E.1 — Severity: **major** — Class: **blocking**

`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier:

- **Files**: Tier S is `< 5 files`; Tier M is `5 - 15`. `progress.md` counts 5
  and classifies S anyway — 5 is M's lower bound, not S's. The count is also an
  undercount: `plan.md` implies `internal/cli/todo_history.go`, the `todo.go`
  registration edit, a `./internal/cli/` test file, a `./internal/kanban/` test
  file (AC-TAQ-012 runs there), both skill documents, **and** the golden files
  AC-TAQ-011 commits — ≥ 7 paths.
- **REQ/AC ceiling**: Tier S is capped at **8 requirements and 8 acceptance
  criteria**. This SPEC carries **15 and 15** — an 87% overrun. The rule is
  explicit: *"Exceeding either ceiling is a signal to tier up or to split the
  SPEC, not to relax the budget."* 15/15 fits Tier M (ceiling 16/16).
- The SPEC already emits the **Tier M artifact set** (spec + plan + acceptance).

`progress.md` §E.1 is transparent that the file count sits at the boundary — that
candour is creditable — but it never mentions the REQ/AC ceiling, and it resolves
the boundary toward S, which sets the plan-auditor threshold to 0.75 instead of
0.80. Tier is not a label; it is the bar.

**Required fix**: set `tier: M` (and update `progress.md` §E.1's rationale to
cite the REQ/AC ceiling as the deciding axis), or split the SPEC so that both
Tier S budgets are met.

---

**D4. REQ-TAQ-004's disclosure self-disables after a single `done`, while the hazard it exists for persists indefinitely — and AC-TAQ-004 locks that in** — `spec.md` §B.1 REQ-TAQ-004; `acceptance.md` § AC-TAQ-004 — Severity: **major** — Class: **blocking**

`REQ-TAQ-004` detects the pre-archive condition as *"the archive is empty AND at
least one live row exists"*. Measured against the actual upgrade path:
`backlog_sqlite.go:95-97` states the DDL uses `IF NOT EXISTS` so *"a queue created
by an earlier binary gains the tables the first time a newer one opens it"*. So on
the operator's queue the sequence is: upgrade → tables appear, empty → disclosure
fires → **first `moai todo done`** → archive non-empty → disclosure silently stops
— while every card destroyed by `v3.1.2` before the upgrade still reads a bare,
confident `absent`, forever.

`AC-TAQ-004` does not catch this; it *asserts* it: *"**And given** the same queue
after one `moai todo done` … that stderr note is absent."* The SPEC's own §A.4
names the standard the requirement fails: *"a surface that silently reports
`absent` for a destroyed card teaches the operator a false negative"*, and
`plan.md` §G lists *"Reporting `absent` without the empty-archive qualifier"* as
the verb's most dangerous output. The chosen heuristic meets that standard for
about one card's lifetime.

**Required fix**: make the condition a function of the hazard rather than of the
archive's current emptiness — e.g. record a durable marker when an
archive-capable binary first opens a queue whose archive tables it had to create,
and qualify `absent` whenever that marker says the queue predates the archive.
Alternatively, state plainly in §A.4 and in the requirement that the disclosure
is one-`done` wide and why that was accepted. What is not acceptable is the
current state, where the limitation is invisible.

---

**D5. The stale sibling `backlog.json` — the real t395 finding — is a live risk to this verb and appears nowhere** — `spec.md` §D / §E; `plan.md` §B — Severity: **major** — Class: **blocking**

t395's measured finding is that `.moai/state/todo/` holds three coexisting files:
`backlog.db` (live), `backlog.json` (frozen 08/31 21:13), and
`backlog.json.migrated`. A lane already read the stale copy and got a wrong answer
with no error. `premise-check.md` § Residual-risk independently flags that the
legacy-JSON read path (`loadLegacyBacklogJSON`,
`internal/kanban/backlog_migrate.go:411`) was never traced, and that a
JSON-backed queue *"may populate `rec.Archived` by a different route, or not at
all"*.

That is directly load-bearing here. `REQ-TAQ-013` promises graceful degradation
when *the archive tables are missing*; it says nothing about the case where the
read resolves to a stale JSON document that answers every query confidently and
carries no archive at all. This verb's entire value is distinguishing "never
issued" from "closed" — a stale store makes it authoritative and wrong, which is
the exact failure it exists to prevent. Having correctly identified the false
premise it was dispatched with, the SPEC absorbed the wrong half of the sibling
card's finding (an invented path change, D1) and dropped the real one.

**Required fix**: either add a requirement that the verb reports which store
answered (or refuses to answer from a legacy-JSON path it cannot vouch for), or
record the JSON path explicitly in §D Out of Scope **with** the false-negative
consequence stated, so the gap is a decision rather than an oversight.

---

**D6. AC-TAQ-014's negated grep passes when its subject file does not exist** — `acceptance.md` § AC-TAQ-014 — Severity: **minor** — Class: **blocking**

```
! grep -rn "AskUserQuestion\|survey\.\|promptui\|bufio.NewReader(os.Stdin)" internal/cli/todo_history.go
```

`grep` on a missing path exits 2; `!` inverts that to success. If the run phase
registers the verb inside `todo.go` rather than creating `todo_history.go` — a
plausible choice the SPEC does not forbid — this AC reports PASS while its subject
does not exist. The check is additionally near-baseline-vacuous: no existing
`todo` verb prompts, so the pattern has no realistic failing input.

**Required fix**: assert the file exists first (`test -f internal/cli/todo_history.go &&
! grep -qE "…" internal/cli/todo_history.go`), or scope the grep to the verb's
symbol across `internal/cli/` so the check binds wherever the verb lands.

---

**D7. Two requirements use a nominal negative subject rather than the canonical GEARS unwanted form** — `spec.md` §B.3 REQ-TAQ-012, REQ-TAQ-014 — Severity: **minor** — Class: **optional**

*"No storage change shall be made: …"* and *"Nothing in this verb shall prompt the
user"*. The canonical unwanted form is `The <subject> shall not [action]`. Meaning
is unambiguous in both and each carries `shall`, so MP-2 passes; this is style,
not substance.

**Required fix (optional)**: *"The run-phase change shall not add a table, alter a
`CHECK` constraint, add a fourth `BacklogState` value, or change
`schema_version`."* / *"The verb shall not prompt the user."*

---

**D8. REQ-TAQ-011 and AC-TAQ-011 cover different surface sets** — `spec.md` §B.3 vs `acceptance.md` § AC-TAQ-011 — Severity: **minor** — Class: **optional**

The requirement names `list`, `next`, `why`, `analyze` and the state counts; the
AC additionally captures `list --json`. An AC that is a superset of its REQ is
harmless (it tests more), but the two should say the same thing or the next reader
will wonder which is authoritative.

**Required fix (optional)**: add `list --json` to `REQ-TAQ-011`.

---

## Where the SPEC is strong (recorded so a revision does not damage it)

Three things this SPEC does better than most, and which the fixes above must not
regress:

- **§A.4 is genuinely honest about the thing that most undermines the card.** The
  dispatch asked whether the SPEC lets a reader believe the surface closes the
  motivating incidents. It does not: *"It does not, and cannot, recover the
  harness-selection card or t81/t83/t88/t89 — the two incidents that motivated the
  card."* Repeated in `plan.md` §B.1 and `progress.md`. Measurement 9 is reflected
  plainly, and I reproduced it in full. This is the correct handling.
- **§A.5 earns the complexity.** The dispatch asked whether `export-json | jq`
  already answers the question. The SPEC argues three reasons and all three
  verify: it writes to disk (`writeExportAtomic`), it prints a downgrade warning
  on every archive-bearing invocation (`todo_export.go:100-110`), and it
  structurally cannot report on an id in neither array — which is precisely the
  fact both incidents needed. One verb with two forms, each mapped to one of the
  two incidents; search and an `--archived` flag on `list` are both explicitly
  excluded, with reasons. Not over-scoped.
- **Constraint compliance is clean.** `REQ-TAQ-010/011/012` restate REQ-TDG-004,
  005 and 007 as this SPEC's own obligations, `plan.md` §G names the four
  temptations that would break them, and no requirement grows the default `list`
  read. The dispatch's item 4 finds nothing.

---

## Recommendation

**FAIL at iteration 1.** Five blocking defects, of which three change what the
run phase would build:

1. **D1** — correct the §E factual inversion and its unresolvable citation. This
   is the highest priority: it is a false statement about develop's behavior
   presented as measured, and it propagates into `plan.md` §B.2 and
   `premise-check.md` § Gaps. Fix all three surfaces.
2. **D3** — set `tier: M`. The REQ/AC ceiling (8) is exceeded by 15, and the
   artifact set already emitted is M's. Re-audit is then scored at 0.80.
3. **D2** — give `AC-TAQ-011` a capture tree it cannot certify itself against.
   This AC is the only thing standing between the change and a silent
   `REQ-TDG-007` regression.
4. **D4** — either re-base `REQ-TAQ-004`'s condition on the durable fact (this
   queue predates the archive) or state in §A.4 that the disclosure is
   one-`done` wide. `AC-TAQ-004`'s second clause must change either way.
5. **D5** — decide the stale-`backlog.json` question: a requirement, or an
   explicit Out-of-Scope line carrying the false-negative consequence.
6. **D6** — make `AC-TAQ-014` fail when its subject file is absent.

D7 and D8 are optional and may be left; routing them into a revision is not worth
a round trip on its own.

Iteration 2 will be scoped to this enumerated defect delta plus a regression check
over it, per the Retry Loop Contract. Note for the re-audit: `origin/develop` has
moved 21 commits ahead of the tree these artifacts cite, so a revision should
re-run `plan.md` §C and refresh any citation it touches at the new HEAD rather
than carrying the `2c18091d1` line numbers forward unverified.
