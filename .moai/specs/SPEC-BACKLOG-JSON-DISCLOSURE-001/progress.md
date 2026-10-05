# Progress — SPEC-BACKLOG-JSON-DISCLOSURE-001

Card: t395 · Worktree `.claude/worktrees/t395` · Branch `WT-stale-backlog-json`

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (3-file artifact set: spec.md + plan.md + acceptance.md; plan-auditor PASS threshold 0.80)
- Requirements: 14 (ceiling 16) · Acceptance criteria: 16 (ceiling 16)
- Artifacts emitted: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- Status: `draft` — awaiting Implementation Kickoff Approval
- Evidence base: `.moai/reports/t395/{premise-verdict,reader-surfaces,r1-repro}.md`

### Plan-audit iterations

| Iter | Verdict | Score | Outcome |
|---|---|---|---|
| 1 | PASS-WITH-DEBT | 0.85 (Tier M threshold 0.80) | 5 blocking-class + 3 optional defects — `.moai/reports/t395/plan-audit.md` |
| 2 | PASS | 0.94 | iter-1's 8 defects all RESOLVED; 3 new optional raised — `.moai/reports/t395/plan-audit-iter2.md`. spec.md 0.1.0 → 0.2.0 |
| 3 | — | — | D9 closed (one clause). D10/D11/D12 accepted as residual risk by lead decision. spec.md 0.2.0 → 0.2.1 |
| 4 | — | — | Disclosure breadth DECIDED by the operator — read surface in full. No criterion changed. spec.md 0.2.1 → 0.2.2 |

Iteration 2 closures: D5 (AC-BJD-015 single-regex blindness → two enumerated
greps), D2 (AC-BJD-007 → runnable commands), D4 (AC-BJD-010 Given → constructible
and self-evidencing, unbuilt Given = Gap), D3 (verb breadth unified at the
read-surface floor; open decision recorded in `plan.md` §D.1), D1 (rebinding →
directory-based), D6 (three files / four sites; REQ-BJD-003 `Where` → `While`),
D7 (§A.2 R5 citation replaced with the two greps that actually cover the sites),
D8 (AC-BJD-002 count scoped to this SPEC's line + archive-tables Given).

**Disclosure breadth: DECIDED 2026-09-02** (operator, relayed by the lead) — the
read surface **in full**, not the `todo history` precedent alone; write verbs stay
out. Recorded in `plan.md` §D.1 with the deciding evidence. No criterion changed:
REQ-BJD-002 already bound a read command. The concrete verb list is enumerated
from the code at run-phase (M3) with its derivation basis recorded — no artifact
names it from recall.

## §E.2 Run-phase Evidence

Baseline attribution for every row below: measured in THIS run, against THIS
tree — worktree `.claude/worktrees/t395`, branch `WT-stale-backlog-json`,
working tree on top of `4a4bbe396` (base `origin/develop@ad272be20`), macOS
darwin/arm64. Commands are quoted as run from the worktree root.

### The read-verb enumeration, and how it was derived

REQ-BJD-002's breadth was decided as "the read surface in full" and the
concrete verb list was deliberately left unnamed, to be derived from the code
rather than recalled (plan.md §D.1, §G). The derivation:

1. The registered verbs come from the single registration site,
   `internal/cli/todo.go:148` `cmd.AddCommand(...)` — 17 verbs: add, list,
   done, undone, next, unpick, edit, move, drop, undrop, analyze, relate,
   unrelate, why, pr, export-json, history.
2. Each verb's RunE path was classified by whether it reaches
   `BacklogStore.Mutate` (the locked write) or only `Load` / `LoadPure`:

```
$ awk '/^func /{fn=$2; sub(/\(.*/,"",fn)} /Mutate\(func\(/{print FILENAME"  MUTATES inside "fn} \
  /store.Load\(\)|newTodoStore\(\).Load\(\)|store.LoadPure\(\)/{print FILENAME"  READ inside "fn}' \
  internal/cli/todo*.go
internal/cli/todo.go  MUTATES inside runTodoAddAppend
internal/cli/todo.go  MUTATES inside runTodoAddPick
internal/cli/todo.go  READ inside runTodoList
internal/cli/todo.go  MUTATES inside newTodoDoneCmd
internal/cli/todo.go  READ inside newTodoNextCmd
internal/cli/todo.go  MUTATES inside newTodoNextCmd
internal/cli/todo.go  MUTATES inside newTodoUnpickCmd
internal/cli/todo_analysis.go  MUTATES inside newTodoAnalyzeCmd
internal/cli/todo_drop.go  MUTATES inside newTodoDropCmd
internal/cli/todo_drop.go  MUTATES inside newTodoUndropCmd
internal/cli/todo_edit_move.go  MUTATES inside newTodoEditCmd
internal/cli/todo_edit_move.go  MUTATES inside newTodoMoveCmd
internal/cli/todo_export.go  MUTATES inside runTodoExportJSON
internal/cli/todo_history.go  READ inside runTodoHistory
internal/cli/todo_pr.go  READ inside runTodoPR
internal/cli/todo_relate.go  MUTATES inside runTodoRelate
internal/cli/todo_relate.go  MUTATES inside newTodoUnrelateCmd
internal/cli/todo_undone.go  MUTATES inside newTodoUndoneCmd
internal/cli/todo_why.go  READ inside newTodoWhyCmd
```

**The read surface is four verbs**: `list`, `why`, `pr`, `history` — plus the
bare `moai todo` invocation, whose RunE with no args is `runTodoList`
(`internal/cli/todo.go:144`), so it is the same code path as `list` and
carries the disclosure for free. `next` reads AND mutates and is therefore a
write verb; the disclosure does not ride it. The enumeration is recorded in
the test file header (`internal/cli/todo_json_disclosure_test.go`) so the
basis travels with the check.

### The R1 red, demonstrated before the repair (AC-BJD-008)

The shipped watch block was run VERBATIM (rebinding only the process working
directory to a fixture project root, which rebinds every relative target the
script resolves) against an isolated fixture queue, across a real mutation:

```
$ go test ./internal/kanban/ -run TestForemanQueueWatch -v -timeout 300s   # PRE-REPAIR TREE
--- FAIL: TestForemanQueueWatch_FiresOnMutation (40.06s)
    --- FAIL: TestForemanQueueWatch_FiresOnMutation/local (20.03s)
        foreman_queue_watch_test.go:180: no change event within 16s across a real queue
        mutation — the local watch does not observe the authoritative store
    --- FAIL: TestForemanQueueWatch_FiresOnMutation/template (20.03s)
        foreman_queue_watch_test.go:180: no change event within 16s across a real queue
        mutation — the template watch does not observe the authoritative store
--- PASS: TestForemanQueueWatch_ShippedJSONTargetIsSilent (20.03s)
--- FAIL: TestForemanQueueWatch_FiresWithStaleJSONPresent (20.04s)
--- FAIL: TestForemanQueueWatch_WatchTargetsAgree (0.00s)
```

The falsifiability condition is demonstrated, not asserted: the same check is
run against the pinned pre-repair block in every later run
(`TestForemanQueueWatch_ShippedJSONTargetIsSilent`), and it stays silent. A
check that fired on both targets would assert nothing.

### The WAL measurement decided the watch target (AC-BJD-010)

The Given was BUILT and EVIDENCED, not assumed. `wal_autocheckpoint` is
per-connection and unset anywhere in `internal/kanban`, so the deferred state
was produced from a connection that sets `wal_autocheckpoint(0)` and stays
open — the audit's suggested side-connection technique would have read as
configured while controlling nothing:

```
$ go test ./internal/kanban/ -run TestForemanQueueWatch_SeesWALDeferredCommit -v
    baseline cksum(backlog.db) = 1178858955 40960
    Given established: backlog.db-wal = 12392 bytes, cksum(backlog.db) unmoved at 1178858955 40960
--- PASS (20.05s)
```

`TestForemanQueueWatch_DBOnlyTargetMissesWALDeferral` pins the control that
DECIDED the target: a watch polling `backlog.db` alone stays silent under a
committed-but-uncheckpointed write. Covering `backlog.db-wal` is therefore a
measured requirement, not a precaution — and the naive `cksum backlog.db`
repair would have traded one silent blind spot for another. The observation
window was NOT widened: every armed watch uses the same 16s bound.

### AC matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-BJD-001 | PASS | `go test ./internal/kanban/ -run TestInspectArchiveVouch -v` | `--- PASS: TestInspectArchiveVouch_StateD_ReportsNonAuthoritativeJSON` + `_SteadyState_NoJSONToDisclose` + `_LegacyJSON_IsAuthoritative` + `_NoStore_NothingToDisclose`; `ok  internal/kanban  0.424s` |
| AC-BJD-002 | PASS | `go test ./internal/cli/ -run TestTodoReadSurface_DisclosesNonAuthoritativeJSON -v` | PASS on all 5 subtests (`bare`, `list`, `why`, `pr`, `history`); each asserts **exactly one** line carrying `is NOT the queue`, naming `the SQLite backlog store` and `backlog.json` |
| AC-BJD-003 | PASS | `go test ./internal/cli/ -run TestTodoReadSurface_StdoutUnpolluted -v` | PASS ×5 — stdout byte-identical with and without the `backlog.json`, zero disclosure bytes on stdout, line present on stderr |
| AC-BJD-004 | PASS | `go test ./internal/cli/ -run TestTodoReadSurface_SilentWithoutJSON -v` | PASS ×5 — zero disclosure lines on either stream; fixture archive tables present, so the REQ-TAQ-013 line cannot co-fire |
| AC-BJD-005 | PASS | `go test ./internal/cli/ -run TestTodoDisclosure_LeavesBacklogJSONUntouched -v` | PASS — sha256 and mtime unchanged after all five read invocations (mtime backdated 2h first, so a byte-preserving rewrite would still be caught) |
| AC-BJD-006 | PASS | `go test ./internal/kanban/ -run TestInspectArchiveVouch_ReadOnlyDatabase -v` | PASS — inspector returns its report against a `0o400` database, and no queue lock file exists after it runs |
| AC-BJD-007 | PASS | `grep -c 'func Inspect.*Backlog.*\(Store\|Vouch\)' internal/kanban/*.go` · `grep -rn 'BacklogStore\(SQLite\|LegacyJSON\|None\) *=' internal/kanban/ \| grep -v _test.go` | summed count **1**, sole non-zero file `backlog_archive_vouch.go:1`; all three constants at `backlog_archive_vouch.go:30,32,34`; the new fact is `NonAuthoritativeJSON bool` **inside** `type BacklogArchiveVouch struct` (`:41-53`) — a field, not a second inspector |
| AC-BJD-008 | PASS | `go test ./internal/kanban/ -run TestForemanQueueWatch -timeout 400s` | `ok  internal/kanban  120.476s`; pre-repair RED quoted above; falsifiability control (`ShippedJSONTargetIsSilent`) green in the same run |
| AC-BJD-009 | PASS | (same run) `TestForemanQueueWatch_FiresWithStaleJSONPresent` | PASS — Case B observed, closing the Gap `r1-repro.md` recorded as reasoned-not-observed |
| AC-BJD-010 | PASS | `go test ./internal/kanban/ -run 'TestForemanQueueWatch_(SeesWAL\|DBOnly)' -v` | Given evidenced (`backlog.db-wal = 12392 bytes`, `cksum(backlog.db)` unmoved before AND after the window); event observed; db-only control silent |
| AC-BJD-011 | PASS | `go test ./internal/kanban/ -run TestForemanQueueWatch_WatchTargetsAgree -v` | `--- PASS (0.00s)` — the two extracted blocks are byte-identical, neither names `backlog.json`, both name `backlog.db` |
| AC-BJD-012 | PASS | `grep -rn 'state/todo/backlog\.json' .claude/skills/` · `grep -rn 'moai/todo/<project-key>/backlog\.json' .claude/skills/` | both: no output, exit 1 (zero matches). Residual form `grep -rn -e … -e … \| grep -v -e 'backlog\.db' -e 'export'` → zero matches |
| AC-BJD-013 | PASS | `grep -n 'backlog\.db' .claude/skills/moai/SKILL.md` · read `workflows/todo.md:17-30` | `SKILL.md:170` now `State: \`.moai/state/todo/backlog.db\` …`; `todo.md` names `.moai/state/todo/backlog.db` and the home fallback `~/.moai/todo/<project-key>/backlog.db` |
| AC-BJD-014 | PASS | `git diff --exit-code -- .moai/docs/todo-queue-storage.md internal/template/templates/.moai/docs/todo-queue-storage.md` | exit 0, no output — the control document is untouched on both trees |
| AC-BJD-015 | PASS | `go test ./internal/template/ -run TestBacklogJSONDisclosure_TemplateMirrorIsComplete -v` | PASS — pattern 1 leaves exactly one match, `templates/.moai/docs/todo-queue-storage.md` (the export-json control); pattern 2 (home fallback) zero matches. **Both** patterns run; a single-regex form would have passed over three of four sites |
| AC-BJD-016 | PASS | `make build` then `go test ./internal/template/ -run TestBacklogJSONDisclosure_EmbeddedTemplatesMatchSource -v`; plus `strings bin/moai \| grep -c 'cksum "$d"/backlog.db'` | test PASS for all **three** mirrored files; `strings` → `1` for the repaired watch line and `0` for `f=.moai/state/todo/backlog.json`, so the shipped binary carries the corrected templates and no longer carries the old block |

### Independent verification batch

| Check | Command | Output |
|---|---|---|
| kanban tests | `go test ./internal/kanban/... -timeout 900s` | `ok  github.com/modu-ai/moai-adk/internal/kanban  137.182s` |
| cli tests | `go test ./internal/cli/ -timeout 900s` | `ok  github.com/modu-ai/moai-adk/internal/cli  377.506s` |
| template tests | `go test ./internal/template/... -timeout 900s` | `ok  …/internal/template  25.843s` · `ok  …/internal/template/agentemit  0.671s` |
| vet | `go vet ./internal/kanban/... ./internal/cli/... ./internal/template/...` | no output, exit 0 |
| lint | `golangci-lint run --timeout=5m ./internal/kanban/... ./internal/cli/... ./internal/template/...` | `0 issues.` |
| host build | `go build ./...` | exit 0 |
| windows cross-build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| coverage (package) | `go test -cover ./internal/kanban/ ./internal/cli/ ./internal/template/` | kanban **86.5%**, cli **80.2%**, template **86.3%** |
| coverage (changed code) | `go tool cover -func` on the same profiles | `todo_disclosure.go` both functions **100.0%**; `InspectBacklogArchiveVouch` **100.0%**; `runTodoList` 93.2%, `runTodoPR` 88.1%, `newTodoWhyCmd` 86.7%, `runTodoHistory` 72.7% |

**Tree attribution for the two coverage rows above.** They are the run-phase
batch, measured on the **pre-merge run-phase tree `d7813ec2b`** — not on the
current tree. On the current tree `41a3a500b`, after the develop merge brought
`internal/cli/memory_drain.go` (178 lines of production code) into the package
and so changed the denominator, `internal/cli` measures **80.3%** (measured
twice independently there). Both figures are exact; Go coverage is
deterministic per tree, so the difference is the tree, not noise. The branch is
currently **56 commits behind `develop`** — absorbing those in the merge window
can move the figure again, and a pre-merge measurement is not a post-merge
basis (`CLAUDE.local.md` §4.1 rule 5). **Re-measure after the merge** rather
than assuming these still apply.

### Gaps — explicitly NOT observed

- **`internal/cli` package coverage (80.2%) was not baselined against the
  pre-change tree.** The figure is measured in this run; whether it moved is
  unattributed. The changed code is **not uniformly above the package figure**.
  By name and value: `discloseNonAuthoritativeBacklogJSON` 100.0%,
  `discloseQueueLayout` 100.0%, `InspectBacklogArchiveVouch` 100.0%,
  `runTodoList` 93.2%, `runTodoPR` 88.1%, `newTodoWhyCmd` 86.7%,
  `runTodoHistory` **72.7%** — the last below the 80.2% package figure. So the
  changed code cannot be used to argue the shortfall away, and for the ones
  above it that would be reasoning rather than a measurement. Recorded as a
  Gap, not a claim.
- **AC-BJD-016's Go check proves embed-directive ↔ source parity at test-build
  time**, not that the on-disk `bin/moai` was rebuilt. The `strings bin/moai`
  probe above is the separate evidence for the shipped artifact; both are
  reported rather than conflated.
- **The watch tests skip on a platform without `sh` and `cksum`** (Windows CI
  leg). The foreman arms this loop through a POSIX shell, so there is nothing
  to observe there; the skip is explicit (`t.Skipf`) rather than a silent pass.
- **The full test suite was not run locally** — deliberately, per
  `CLAUDE.local.md` §4/§6 (parallel lanes running `go test ./...` drove machine
  load to 413 on 2026-08-15). The full-suite verdict is CI's.
- **`origin/develop` CI is already red on two `internal/web` i18n failures**,
  unrelated to this card and dispatched elsewhere. Not re-measured here and
  not attributed to this change.

### Adjacent defect found, deliberately NOT fixed (out of scope)

`internal/cli/todo.go:94` — the `moai todo` command's own `Long` help text is
a fifth instance of the same defect class: it says "Operate the kanban backlog
queue at `.moai/state/todo/backlog.json`" and "keeps its queue at
`~/.moai/todo/<project-key>/backlog.json`". Both are false after the SQLite
cutover, and `moai todo --help` is a reader surface. It is NOT one of the four
sites §A.2 names, REQ-BJD-011/012 bind only the two skill files, and
AC-BJD-012's greps are scoped to `.claude/skills/` — so repairing it here would
be scope expansion. Reported for the lead to card.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-02
run_commit_sha: pending-backfill-lead-commits
run_status: audit-ready
ac_pass_count: 16
ac_fail_count: 0
ac_gap_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (lane does not commit; the lead commits)
l44_post_push_fetch: not-run (lane does not push)
new_warnings_or_lints_introduced: 0   # golangci-lint → "0 issues."
cross_platform_build:
  darwin_arm64: pass          # go build ./... exit 0
  windows_amd64: pass         # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 20   # 14 modified (tracked) + 6 new (untracked)
m1_to_mN_commit_strategy: none — the lane leaves the tree uncommitted by dispatch instruction
files:
  source:
    - internal/kanban/backlog_archive_vouch.go        # M1 — the new field
    - internal/cli/todo_disclosure.go                 # M3 — the disclosure (new)
    - internal/cli/todo.go                            # M3 — list call site
    - internal/cli/todo_why.go                        # M3 — why call site
    - internal/cli/todo_pr.go                         # M3 — pr call site
    - internal/cli/todo_history.go                    # M3 — history call site
  tests:
    - internal/kanban/backlog_json_disclosure_test.go       # new
    - internal/kanban/foreman_queue_watch_test.go           # new
    - internal/kanban/foreman_queue_watch_wal_test.go       # new
    - internal/cli/todo_json_disclosure_test.go             # new
    - internal/template/backlog_json_disclosure_mirror_test.go  # new
  skills_local:
    - .claude/skills/moai-kanban-foreman/SKILL.md     # M2 — watch target
    - .claude/skills/moai/SKILL.md                    # M4 — state location
    - .claude/skills/moai/workflows/todo.md           # M4 — state location ×2
  skills_template:
    - internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md
    - internal/template/templates/.claude/skills/moai/SKILL.md
    - internal/template/templates/.claude/skills/moai/workflows/todo.md
  generated:
    - internal/template/catalog.yaml                  # M5 — regenerated by `make build` (skill-hash cascade)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-02
sync_commit_sha: pending-backfill-lead-commits   # the lane does not commit; the lead commits
sync_status: audit-ready
b12_self_test_a: pass   # grep -c 'SPEC-BACKLOG-JSON-DISCLOSURE-001' CHANGELOG.md -> 0 before emission (no duplicate)
b12_self_test_b: pass   # distinct AC ids in acceptance.md -> 16; CHANGELOG entry states "All 16 ACs green" (non-zero, not a vacuous 0 == 0)
b12_self_test_c: pass   # every path named in the CHANGELOG entry verified present via ls
changelog_entry_position: "[Unreleased] -> ### Added, first bullet (CHANGELOG.md:12)"
frontmatter_status_transitions:
  spec.md: "in-progress -> completed (single sync commit; updated: already 2026-09-02, unchanged)"
  plan.md: "n/a - no YAML frontmatter block (file opens with an H1)"
  acceptance.md: "n/a - no YAML frontmatter block (file opens with an H1)"
  progress.md: "n/a - no YAML frontmatter block (file opens with an H1)"
canary_compliance_check: n/a   # this SPEC defines no forward-looking policy that its own sync tests
```

### Documentation surfaces checked

| Surface | Scope of the check | Finding |
|---|---|---|
| `README.md` / `.ko` / `.ja` / `.zh` | `grep -n -e 'backlog' -e 'state/todo' -e 'moai todo'` over all four files | **No statement about where queue state lives exists in any locale.** Every hit is either a board-column name (`backlog` the column), a `/moai todo` usage line, or a docs-site link. Nothing to repair. |
| `docs-site/content/{en,ko,ja,zh}/utility-commands/moai-todo.md` §State file | read | **Already correct** — names `.moai/state/todo/backlog.db` as the one SQLite database and points at `.moai/docs/todo-queue-storage.md`. Not false, not edited. |
| `docs-site/content/{en,ko,ja,zh}/utility-commands/moai-todo.md` `export-json` row | read | Not false but **incomplete after this SPEC**: it said later `todo` verbs leave the exported `backlog.json` where it was put, which is still true, but they now also disclose it. One clause appended in all four locales. |
| `docs-site/content/{en,ko,ja,zh}/advanced/moai-web-console.md:108` | read | the `kanban` row names `.moai/state/todo` — the **directory**, not a file. Correct as written; untouched. |
| `docs-site` repo-wide | `grep -rn 'state/todo\|backlog' docs-site/content/` | No remaining occurrence claims the queue is a JSON file. The only `backlog.json` mentions are the `export-json` rows, which describe it correctly as a legacy export. |
| `.moai/docs/todo-queue-storage.md` (+ template mirror) | `git diff --exit-code` | exit 0 — byte-identical on both trees, AC-BJD-014 re-verified after all sync-phase edits. |

### Sync-phase verification

| Check | Command | Output |
|---|---|---|
| CHANGELOG duplicate guard (B12-a) | `grep -c 'SPEC-BACKLOG-JSON-DISCLOSURE-001' CHANGELOG.md` | `0` before emission, `1` after |
| AC count (B12-b) | `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md \| sort -u \| wc -l` | `16` (AC-BJD-001..016) |
| docs-site build | `hugo --source docs-site --destination /tmp/t395-hugo-out2` | exit 0, `Total in 3320 ms`, 187/185/185/185 pages; `grep -ci 'warn\|error'` over the log -> `0`; `sitemap.xml` present |
| 4-locale heading parity | `grep -c '^#'` on the four `moai-todo.md` | `28` in every locale (unchanged — the edit added prose to an existing table row, no heading) |
| frozen control doc | `git diff --exit-code -- .moai/docs/todo-queue-storage.md internal/template/templates/.moai/docs/todo-queue-storage.md` | exit 0 |
| Go vet (MX edit) | `go vet ./internal/cli/` | exit 0, no output |
| gofmt (MX edit) | `gofmt -l internal/cli/todo_disclosure.go` | no output |
| disclosure ACs re-run | `go test ./internal/cli/ -run 'TestTodoReadSurface\|TestTodoDisclosure' -count=1` | `ok  github.com/modu-ai/moai-adk/internal/cli  7.434s` |

### MX Tag validation (sync sub-step)

`discloseQueueLayout` (`internal/cli/todo_disclosure.go`) has **three** non-test
callers — `todo.go:314` (bare/`list`), `todo_why.go:28`, `todo_pr.go:120` —
crossing the `fan_in >= 3` threshold at which the constitution requires an
`@MX:ANCHOR`. One was added, in the established repo form
(`// @MX:ANCHOR fan_in=3 - <SPEC-ID> <REQ> …`, matching
`internal/runtime/gobin/resolver.go:9`). Comment-only; `go vet` and `gofmt`
clean and the disclosure ACs re-run green above.

Below threshold and deliberately not annotated:
`discloseNonAuthoritativeBacklogJSON` (2 callers) and
`InspectBacklogArchiveVouch` (2 non-test callers).

### Sync-phase gaps — explicitly NOT observed

- **The docs-site clause was not proof-read by a native reader in ja / zh / ko.**
  It was authored as native prose rather than a word-for-word carry of the
  English, and the page's own register was matched, but no second reader
  confirmed it. The hugo build proves it renders, not that it reads well.
- **The docs-site edit is not covered by any AC.** No criterion in
  `acceptance.md` asserts it, so it carries no mechanical regression guard —
  it is sync-phase documentation synchronisation, verified only by the build
  and the parity counts above.
- **`internal/cli` was not re-run in full after the MX comment edit.** Only the
  disclosure ACs were re-run (7.4s). A comment cannot change behaviour and
  `go vet` is clean, but that is reasoning, not a full-package measurement.
- **`origin/develop`'s two `internal/web` i18n failures were not measured on
  this branch**, in either direction. This branch has absorbed
  `origin/develop`; whether the merge carries their fix is unknown here and
  is deliberately not attributed to or away from this card.
- **The CHANGELOG entry restates run-phase figures** (coverage, AC counts,
  lint) from `§E.2`; they were not independently re-measured in sync-phase.
  Their baseline attribution is `§E.2`, not this section.
