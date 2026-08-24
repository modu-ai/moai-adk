# SPEC-KANBAN-RECORD-SESSION-KEY-001 — Progress

Tier M — three artifacts (spec.md, plan.md, acceptance.md). 8 requirements, 9 acceptance criteria.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-24
spec_version: 0.2.0
iteration: 2
tier: M
threshold: 0.80
requirements: 8
acceptance_criteria: 9
milestones: 2
measured_at: 3c3a6fbf8
prior_audit: .moai/reports/t207/plan-audit-kanban-record-key-iter1.md
```

## §E.2 Run-phase Evidence

Run-phase base: `aef1b51f3` (branch `WT-web-live-todo`, worktree
`.claude/worktrees/t207`). Every command below was run in this tree, this run.
Source paths are worktree-relative; `.moai/state/…` paths are read under
`/Users/goos/MoAI/moai-adk-go/` per the SPEC's measurement convention.

### Commits (nothing pushed — the branch has no upstream and no remote ref)

| SHA | Subject |
|---|---|
| `41666ca5f` | M1 widen the record schema with lane and card |
| `8838e31ec` | M2 move the record write into the session |
| `170935d18` | derive the card id without a git subprocess |
| `3964ae7c6` | satisfy the two absence greps literally |
| `69ecae652` | pin the factory lane join's third hop |
| `19da7c6d9` | supersede AC-FM-023a's record-writing half |

### AC matrix

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-KRS-001 | PASS | `go test -run TestRecordIsKeyedByTheRuntimeSessionIDNotTheSidecar -v ./internal/hook/` | `--- PASS` @ `19da7c6d9`. Fixture: sidecar holds `T-1111…`, runtime delivers `S-9999…`; `S.json` exists with `session_id=S`, `role=run`, and `T.json` does not exist. Baseline re-measured: `active-sessions.json` carries one live session `72d4805a-5319-42f1-9feb-b6664eed8514`, and `ls .moai/state/kanban/72d4805a-….json` → `No such file or directory`; `cat .moai/state/current-session-id.txt` → that same identifier. The property holds pre-change: a registered session has no record under its own identifier |
| AC-KRS-002 | PASS | `grep -rn 'CurrentSideChannelFile\|current-session-id\|resolveLaunchSessionID\|resolveCurrentSessionID' internal/kanban/ internal/cli/kanban.go` | no output, `rc=1` (baseline: 1 hit, `internal/cli/kanban.go:474`). Load-bearing half: the two-identifier fixture above, which the pre-change writer fails by construction |
| AC-KRS-003 | PASS | (a) `grep -rn 'kanban.WriteBestEffort\|kanban.Write(' internal/cli/` → no output, `rc=1` (baseline: 1 hit at `kanban.go:478`). (b) `go test -run TestLauncherWritesNoKanbanRecord ./internal/cli/` → `ok`; the fixture root's record-directory listing is identical before and after the launcher call. `TestCC_KanbanWritesNoStateRecord` additionally drives `runCC --kanban SPEC-PLACEHOLDER` end to end and asserts no file carrying `session_id` appears | both halves new |
| AC-KRS-004 | PASS | `go test -run 'TestLaneNumberIsRecordedDistinctlyFromRole\|TestFactoryLaneRecordsItsNumberAndLeadRecordsZero' -v` | `--- PASS` both. `lane-3` → lane `3`, role `lane`; kanban lead → lane `0`. Baseline: `grep -n 'Lane\|Card' internal/kanban/record.go` returned exactly two hits (`:119`, `:126`), neither a struct field; `grep -l '"lane"[[:space:]]*:' .moai/state/kanban/*.json` → 0 files (re-measured, still 0) |
| AC-KRS-005 | PASS | `go test -run 'TestCardIdentifierDerivation\|TestCardIdentifierFromADeepCwdInsideACardWorktree\|TestCardIDFromPathRequiresAWorktreesParent' -v` | `--- PASS` all. (a) cwd `…/.claude/worktrees/t207` → `t207`; (b) override `t999` wins; (c) cwd `/Users/goos/moai/moai-adk-go` → the `card_id` key is ABSENT from the encoded record and is not `moai-adk-go`. Baseline: no card field, and `grep -rn 'CARD' internal/config/envkeys.go` returned 0 lines |
| AC-KRS-006 | PASS | (a) `grep -rn 'BackendGLM\|BackendClaude' internal/cli/cc.go internal/cli/glm.go` → the same 8 lines, now `defer exportKanbanLaunchFacts(entry.Spec, kanban.Backend*)()` — no longer an argument to a record write; `grep -rn 'MOAI_KANBAN_BACKEND' internal/config/envkeys.go` → `envkeys.go:213` (baseline `rc=1`, no output). (b) `go test -run TestFactoryLaneRecordsItsNumberAndLeadRecordsZero -v` asserts `backend=glm` on the GLM env and `backend=claude` on the Claude one | both halves new |
| AC-KRS-007 | PASS | `go test -run 'TestPreChangeRecordReadsAndRewritesByteIdentically\|TestNewFieldsAreOmittedWhenEmpty' -v ./internal/kanban/` | `--- PASS` both. (a) fixture produced by marshalling the pre-change struct; read succeeds, `lane=0`/`card_id=""` report as not recorded, and the rewrite is `bytes.Equal` to the input. (b) a non-lane record with no card encodes neither key |
| AC-KRS-008 | PASS | `go test -run TestRecordWriteFailsOpenOnUnwritableStateDirectory -v ./internal/hook/` | `--- PASS`. The session-start write path IS reached (the launch environment marks a kanban session), its attempt fails inside `WriteBestEffort`, no error is returned — `writeKanbanSessionRecord` has no return value, so the handler cannot gate on it — and no record file exists. Baseline: `grep -rln 'kanban\.Write' internal/hook/` had no match |
| AC-KRS-009 | PASS | `go test -run TestFactoryLaneJoinClosesOnTheThirdHop -v ./internal/hook/` | `--- PASS`. `workers.json[lane-5].PID=424242` → `active-sessions.json` entry → `session_id` → `kanban.Read` returns a record whose `session_id` equals that entry's and whose `lane` is `5`. Baseline: the one registered live session has no record file, so the third hop resolves empty for it |

### Verification commands

| Check | Command | Result |
|---|---|---|
| Build | `go build ./...` | exit 0 |
| Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Vet | `go vet ./internal/cli/ ./internal/hook/ ./internal/kanban/ ./internal/config/` | no output |
| Tests — kanban | `go test -count=1 -timeout 600s -cover ./internal/kanban/` | `ok … 62.143s coverage: 84.4% of statements` |
| Tests — hook | `go test -count=1 -timeout 900s -cover ./internal/hook/` | `ok … 97.108s coverage: 84.5% of statements` |
| Tests — config | `go test -count=1 -timeout 900s -cover ./internal/config/` | `ok … 19.924s coverage: 80.6% of statements` |
| Tests — cli | `go test -count=1 -timeout 1200s -cover ./internal/cli/` | `ok … 580.171s coverage: 78.8% of statements` at `19da7c6d9`. An earlier run at `3964ae7c6` FAILed on `TestCC_KanbanWritesStateRecord` — see finding 4 |
| Lint | `golangci-lint run ./internal/kanban/... ./internal/config/... ./internal/hook/...` then `./internal/cli/...` | `0 issues.` exit 0, both |
| One writer | `grep -rn 'kanban.WriteBestEffort\|kanban.Write(' internal/ --include='*.go' \| grep -v _test` | exactly one line: `internal/hook/session_start_record.go:94` |
| Exclusion boundary | `git diff --name-only aef1b51f3..HEAD` | 12 files across `internal/{cli,config,hook,kanban}`; no file under `internal/web/` or `internal/statusline/`; `internal/hook/session_start.go`'s sidecar writer is untouched |
| Existing record files | `ls -la .moai/state/kanban/` before and after, then `diff` | every session-id-named entry present with unchanged size and mtime. The only differing lines are `backlog.json` / `backlog.lock` (the operator's live queue, written by other sessions' `moai todo` work — not a record, and not written by any code this SPEC touches) and the two directory entries' own mtimes |

### Findings raised during the run

1. **A latency regression I introduced and then removed.** Resolving the card
   identifier with `git rev-parse --show-toplevel` put a subprocess on
   SessionStart's synchronous path for every kanban and factory session.
   `TestSessionStart_DeferredScanDoesNotBlockReturn` went from passing to
   650-890ms against its 500ms budget. Attributed by swapping in the
   pre-change `session_start.go` under the identical environment (passed) and
   back (failed). Fixed in `170935d18` — the containment test is a property of
   the path, so the derivation is now pure path arithmetic.

2. **The test process inherits the lane session's launch environment.** These
   tests run inside a `lane-8` factory session, whose `MOAI_FACTORY_WORKER`
   made every case read as lane 8. `scrubKanbanEnv` is the fix, and it is
   documented as load-bearing rather than tidy — this failure reproduces on the
   machine running the lane and never in CI.

3. **Two acceptance greps had one hit each for reasons the requirement does not
   care about** — a doc comment naming the sidecar file, and a test seeding its
   fixture through `kanban.Write`. Both were satisfied literally in `3964ae7c6`
   rather than argued past.

4. **One pre-existing test asserted the behaviour this SPEC removes.**
   `TestCC_KanbanWritesStateRecord` (AC-FM-023a) required the launcher to leave
   a record. Superseded rather than deleted: its fail-open half is unchanged,
   and its record half now asserts the opposite. `TestCompanionRoleBareLabel`
   was removed with its function and its coverage restated in `internal/hook`.

5. **`*.json` in the record directory is not a record glob, and the launch
   proves it.** `runCC --kanban` legitimately writes `leads.json` there, which
   an "empty directory" assertion caught. Records are identified by carrying
   `session_id` — as the SPEC's own convention says.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-24
run_commit_sha: 19da7c6d9
run_status: complete
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 4
l44_pre_commit_fetch: not-run (branch is local-only and unpushed by operator decision)
l44_post_push_fetch: not-run (nothing pushed)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
total_run_phase_files: 12
m1_to_mN_commit_strategy: one commit per milestone plus four follow-up commits (a latency fix, a grep-literalness fix, an added consumer-join test, and a superseded pre-existing test)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-24
sync_commit_sha: pending-backfill-SPEC-KANBAN-RECORD-SESSION-KEY-001
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-KANBAN-RECORD-SESSION-KEY-001' CHANGELOG.md → 0 (rc 1, a genuine zero-match; no duplicate, emission permitted)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 9; the CHANGELOG entry states 9"
b12_self_test_c: "every path named in the CHANGELOG entry verified present with ls before commit (13 paths, all present)"
changelog_entry_position: "[Unreleased] → ### Changed, appended immediately after the SPEC-WEB-TODO-QUEUE-001 entry"
frontmatter_status_transitions:
  spec.md: "in-progress → implemented → completed (updated: already 2026-08-24, the sync date — unchanged)"
  plan.md: "n/a — carries no frontmatter block"
  acceptance.md: "n/a — carries no frontmatter block"
  progress.md: "n/a — carries no frontmatter block"
canary_compliance_check: "n/a — this SPEC defines no forward-looking policy that its own sync tests"
```

### Cross-SPEC effect — AC-FM-023a's record half is superseded

Recorded here, in the SPEC that caused it, so a later reader of *this* SPEC finds the effect
without having to discover it from the other side.

| What | Detail |
|---|---|
| Affected criterion | `SPEC-FACTORY-MODE-001` **AC-FM-023a (state record)** |
| What it pinned | that `moai cc --factory` leaves a record written by the **launcher**, plus a fail-open half |
| Why it no longer holds | REQ-KRS-002/003 removed the launcher-side write. The launcher runs before the session it launches exists, so the identifier it keyed on was not the launched session's; the write moved into that session's own SessionStart hook |
| How it was handled | **superseded, not deleted.** `TestCC_KanbanWritesStateRecord` now asserts the record's ABSENCE at the launcher; the session-side write is covered in `internal/hook`. Run-phase commit `19da7c6d9` |
| The fail-open half | **unchanged and still binding** — an unwritable state directory must not block a launch. The same test still exercises it |
| The annotation | a supersede note was added to `SPEC-FACTORY-MODE-001/acceptance.md` in `1d8b56426`, on the operator's ruling. That SPEC was not otherwise edited, and sync did not touch it |
| One removal, not a supersede | `TestCompanionRoleBareLabel` was removed with its function; its coverage is restated in `internal/hook` |

### Documentation review — what was searched, found and changed

**Changed: nothing.** That is the finding, not a skipped step.

Searched — `.claude/rules/`, its `internal/template/templates/` mirror, `docs-site/content/{en,ko,ja,zh}/`,
and the four READMEs — for any description of the record's writer or of its fields:
`kanban.Record`, `state/kanban`, `kanban/<session`, `current-session-id`, the record's JSON key
names (`session_id`, `entered_at`, `deepscan_dir`, `verify_reentries`, `role`), and the phrase
"launcher writes".

| Hit | Why it is not this record |
|---|---|
| `docs-site/content/<loc>/utility-commands/moai-todo.md` and three template skill bodies naming `.moai/state/kanban/backlog.json` | the backlog **queue**, a different file in the same directory |
| `docs-site/content/<loc>/advanced/moai-web-console.md` state-directory table row `kanban → .moai/state/kanban` | names the directory, not the record's writer or its schema |
| `context-window-management-detail.md` and `goal-directive-detail.md` (both mirror pairs) on `.moai/state/current-session-id.txt` | the goal-arm side channel — a different consumer of the same file, unchanged by this SPEC |
| `docs-site/content/<loc>/advanced/kanban-mode.md` 13-field table carrying `entered_at` | the Origin-Trail Chain's `WorktreeNode`, not `kanban.Record` |
| `docs-site/content/en/multi-llm/kanban-mode.md:114` "the launcher writes a transient settings file" | the cross-session inbound `--settings` file |

No published page and no doctrine file describes the kanban session record at all, so no mirror
pair was touched and `make build` was not run. Writing a page this SPEC did not require would be a
worse outcome than the silence it would fill.
