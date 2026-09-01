# Audit response — SPEC-TODO-ARCHIVE-QUERY-001, plan-audit iteration 1

Responding to `.moai/reports/t394/plan-audit-iter1.md` (FAIL, 0.775 against the
Tier M threshold of 0.80).

Tree: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t394`, branch
`WT-todo-done-history`, working tree at `2c18091d1`. Every source citation below
was re-read at that tree this turn. `origin/develop` has moved ahead of it; the
artifacts deliberately keep citing `2c18091d1` and now carry an explicit
run-phase obligation to re-verify every `file:line` at the then-current develop
(`plan.md` §C).

Status stays `draft`. No implementation code, no tests, no commits, no pushes.

**Closed: 8 of 8 (D1-D6 blocking, D7-D8 optional).** Two of them close in a form
that deviates from the auditor's suggested fix; both deviations are named and
justified below rather than presented as compliance.

---

## Defect ledger

| # | Severity | Closed | What changed | Command that shows it closed |
|---|---|---|---|---|
| D1 | critical | yes | `spec.md` §E rewritten; `plan.md` §B.2 rewritten; `premise-check.md` § Gaps entry withdrawn in place | `grep -rn "func BacklogPathForRoot" internal/` → only `state_dir.go:129`, `:137`; `grep -n 'const stateDirName\|const legacyStateDirName' internal/kanban/state_dir.go` → `37:…"todo"`, `43:…"kanban"`; `grep -rn "backlog_store.go:250" .moai/specs/SPEC-TODO-ARCHIVE-QUERY-001/` → exactly 2 hits, `spec.md:312` and `plan.md:31`, **both inside the withdrawal paragraphs that quote the retracted citation** — no asserting use of it survives (the same holds for `grep -rn "state/kanban"`: `spec.md:316` calls it the legacy name, `plan.md:30` records the withdrawn claim) |
| D2 | critical | yes (deviating form) | `acceptance.md` AC-TAQ-011 rewritten as three clauses with a mechanical capture provenance; `plan.md` gains **M0**, which commits the goldens before any verb code exists; `plan.md` §G gains the matching anti-pattern | `git log --diff-filter=A --format=%H -- internal/cli/testdata/golden/live-readers/` → empty at this tree, so `test -n "$C"` fails — the AC cannot pass without a golden commit that precedes the verb |
| D3 | major | yes | `spec.md` frontmatter `tier: S` → `tier: M`; `progress.md` §E.1 rewritten to cite the REQ/AC ceiling as the deciding axis; `spec.md` §C tier note corrected | `grep -n '^tier:' spec.md` → `tier: M`; `grep -cE '^\- \*\*REQ-TAQ-[0-9]{3}' spec.md` → `15`; `grep -c '^### AC-TAQ-' acceptance.md` → `15` (M ceiling 16/16) |
| D4 | major | yes (deviating form) | `spec.md` §A.4 gains the `last_seq` and `IF NOT EXISTS` measurements plus the residual; REQ-TAQ-004 re-based on the id-space accounting; AC-TAQ-004 rewritten with three clauses, the second of which is a regression guard against the archive-emptiness condition; `plan.md` M2 carries the implementation coupling and §G the anti-pattern | `sqlite3 .moai/state/todo/backlog.db "SELECT value FROM meta WHERE key='last_seq';"` → `401`; `… "SELECT count(*) FROM items;"` → `113`; `… ".tables"` → `findings items meta`; `grep -n "archive is empty AND" spec.md` → no match |
| D5 | major | yes | REQ-TAQ-013 extended to the legacy-JSON store; AC-TAQ-013 gains a second Given/When/Then; `spec.md` §D gains `### Out of Scope — the stale sibling backlog.json` carrying the false-negative consequence; `plan.md` §B.3 records the measurement; `premise-check.md` § Residual-risk partly closed | `ls -la .moai/state/todo/backlog.json*` → `backlog.json` (08-31 21:13) and `backlog.json.migrated` (08-27 23:01) beside `backlog.db` (09-01 02:37); `python3 -c "import json;print(list(json.load(open('.moai/state/todo/backlog.json')).keys()))"` → `['version','last_seq','items','findings']` (no `archived`) |
| D6 | minor/blocking | yes | AC-TAQ-014 rewritten: the subject is located by symbol, existence is asserted first, and a RED plant-and-remove observation is required | old form: `grep -rn "…" internal/cli/todo_history.go` → exit `2`, which `!` inverts to PASS; new form: `git grep -l "newTodoHistoryCmd" -- 'internal/cli/*.go' ':!internal/cli/*_test.go'` → exit `1`, empty, so `test -n "$SRC"` fails |
| D7 | minor/optional | yes | REQ-TAQ-012 and REQ-TAQ-014 rewritten in the canonical `The <subject> shall not` form | `grep -n "REQ-TAQ-012\|REQ-TAQ-014" spec.md` → `The run-phase change shall not add a table…` / `The verb shall not prompt the user` |
| D8 | minor/optional | yes | `list --json` added to REQ-TAQ-011, matching AC-TAQ-011's surface set | `grep -n "REQ-TAQ-011" spec.md` → names `list`, `list --json`, `next`, `why`, `analyze` and the state counts |

---

## The two deviations, stated plainly

### D2 — why not the in-test two-build comparison

The dispatch preferred comparing a pre-change binary against a post-change
binary inside one test. I rejected the two obvious implementations and took a
third form:

- **Building the base tree inside the test.** ~~needs the base commit present in
  every checkout that runs the test. CI checks out at `fetch-depth: 1` by
  default, so the base object is frequently absent and the test would fail for a
  reason unrelated to the invariant — a guard that goes red for the wrong reason
  gets skipped, which is how a BLOCKING guard becomes decorative.~~

  **[CORRECTED at iteration 2 — the struck reason was an unverified premise and
  it is false for this repository.]** `fetch-depth: 1` is `actions/checkout`'s
  default; it is not this repository's configuration, and this repository's
  configuration is what decides whether the alternative would have worked here.
  Measured at worktree `2c18091d1`, this turn:

  ```
  $ grep -n "fetch-depth" .github/workflows/ci.yml
  129:          fetch-depth: 0   # job `test`               (declared line 114)
  264:          fetch-depth: 0   # job `test-race`          (declared line 254)
  382:          fetch-depth: 0   # job `test-integration`   (declared line 368)
  431:          fetch-depth: 0   # job `lint`               (declared line 422)
  486:          fetch-depth: 0   # job `build`              (declared line 463)
  543:          fetch-depth: 0   # job `constitution-check` (declared line 533)
  ```

  Seven `actions/checkout` steps exist; the six above cover **every job that
  compiles or runs Go**, each at full history, four of them carrying an in-file
  comment recording why (`SPEC-V3R4-CI-INFRA-FIX-001 D-3`). The seventh, at line
  51, is the `detect` paths-filter job, which runs no Go.

  **So alternative 1 was viable here, and the record now says so.** The adopted
  form is nevertheless kept, on a premise that does hold and is stated instead:
  building the base tree inside the test would make every run of a `./internal/cli/`
  test compile a second binary from a second worktree, putting a `go build` of
  `./cmd/moai` on the critical path of both jobs whose command reaches that
  package — `test` (`ci.yml:208`, `go test … ./...`) and `test-race`
  (`ci.yml:287`, `go test -json -race -count=1 ./...`). Not `test-integration`,
  whose command is scoped to `./test/integration/harness/...` (`ci.yml:400`) and
  never compiles `./internal/cli`. The adopted form gets the same guarantee — a
  comparison against a tree that genuinely predates the change — from commit
  metadata, at the cost of one `git log` in CI, and pays the build cost once,
  locally, in clause 2. That is a cost preference, not an impossibility claim,
  and it is recorded as one.

  Recorded plainly because the correction matters more than the outcome: the
  original bullet argued **against** an action on a premise nobody had measured,
  and nothing downstream contradicts a wrong "don't do it"
  (`verification-claim-integrity.md` §1.1 surface 4). The mechanism did not have
  to change; the reason did.
- **Compiling the verb out behind a build tag** removes the git dependency but is
  *vacuous for the hazard that matters*: a change to a shared render helper
  appears identically in both builds, so the comparison passes while the default
  read has grown. That is precisely the failure `plan.md` §G names ("a shared
  render helper touched in M1 is exactly how a default read grows by accident").

The adopted form keeps the comparison against a tree that genuinely predates the
change, and proves that mechanically rather than by discipline:

1. **Clause 1** (in CI, cheap) resolves the goldens' introducing commit `C` and
   asserts `C != HEAD`, `C` is an ancestor of `HEAD`, and `C`'s tree contains
   neither `internal/cli/todo_history.go` nor the symbol `newTodoHistoryCmd`. A
   golden captured after the verb existed fails this. An unreachable `C` fails
   `merge-base --is-ancestor` — the safe direction, since unverifiable
   provenance must not read as verified.
2. **Clause 2** (local, cited in `progress.md` §E.2) builds the binary from `C`
   in a throwaway worktree and diffs its output against the committed golden,
   closing the gap clause 1 leaves: that the golden bytes came from `C`'s tree
   rather than from a dirty working directory.
3. **Clause 3** is the comparison the AC was always about.

`plan.md` gains **M0** so the ordering is structural, not a convention: the
goldens are captured and committed alone, before M1 writes a line of the verb.

**Residual**: clause 2 is a Definition-of-Done gate run by the implementer, not a
CI assertion. An implementer who skips it and lies about having run it is not
caught by clause 1 alone. That is a smaller hole than the one that existed
(clause 1 already makes a post-change capture fail mechanically), and it is
recorded here rather than papered over.

### D4 — why not a durable marker written at first archive-capable open

The dispatch suggested recording a marker when an archive-capable binary first
opens a queue whose archive tables it had to create. I did not take it, for two
reasons that are constraints this SPEC already carries:

- It makes an **open into a write**. `QueuedCount` is documented as a PURE read
  on "the statusline's per-render path and the SessionStart notice's, both of
  which run on surfaces that must never move operator bytes"
  (`internal/kanban/backlog_store.go:406-411`). A marker written at open would
  either violate that or need a second, inconsistent open path.
- It puts the change in `internal/kanban`'s open path, outside this SPEC's
  module, and sits awkwardly against REQ-TAQ-010's read-only obligation.

`meta.last_seq` already carries the durable fact and costs nothing: it is a
persisted high-water mark, never derived from present ids, explicitly so that
"`done` removes rows and a derived mark would reuse the removed card's id"
(`internal/kanban/backlog_store.go:14-17`), and `readRecord` already loads it
(`internal/kanban/backlog_migrate.go:106-110`). No write, no schema change, no
new read — and it is strictly *more* informative than the marker, because it is
per-id rather than per-queue: `k <= last_seq` and absent means the id may have
been issued and destroyed, `k > last_seq` means never issued.

**Residual, recorded in `spec.md` §A.4**: `normalizeBacklogRecord` raises the
mark to cover ids the record holds and never lowers it, so a hand-edited or
imported mark can leave a gap that was never an issued card. The disclosure is
therefore worded as a qualification of `absent`, not as a claim that the id
certainly was issued.

---

## What the repair deliberately did not touch

`spec.md` §A.4's original paragraph — that this surface cannot recover the
harness-selection card or t81/t83/t88/t89 — survives **verbatim**. The additions
sit after it and narrow what `absent` means; they do not soften what is
unrecoverable. The audit's three "where the SPEC is strong" items (§A.4's
honesty, §A.5's ladder argument, the REQ-TDG-004/005/007 constraint compliance)
are unchanged apart from D7's grammatical rewording of REQ-TAQ-012.

## Requirement and criterion budget after the repair

15 requirements, 15 acceptance criteria, 1:1, no orphan and no uncovered REQ —
inside Tier M's 16/16 with one slot of headroom. D5 was closed by extending
REQ-TAQ-013 rather than adding a sixteenth requirement, deliberately: landing
exactly on a ceiling is the signal the ceiling exists to raise.
