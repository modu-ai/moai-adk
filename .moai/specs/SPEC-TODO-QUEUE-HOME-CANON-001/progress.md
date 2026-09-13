# progress.md — SPEC-TODO-QUEUE-HOME-CANON-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-13
spec: SPEC-TODO-QUEUE-HOME-CANON-001
tier: M
survey_baseline: "b66789479 (worktree t658, branch WT-home-queue-canonical, clean)"
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
key_finding: "queue path resolution already single-seam; residual work = foreman-skill statement fix + guard tests + disposition ledger; card premise (statusline/backlog.go:50 direct JSON reader) is stale on this tree"
open_clarifications: 0
```

## §E.2 Run-phase Evidence

Measured in this run, this worktree, base `4dfc163a0` → run-phase commits on
`WT-home-queue-canonical`. All commands run from the worktree root.

### M1 — Disposition ledger (REQ-004, frozen)

Re-verified against the run-phase HEAD (not carried from plan-phase). Per-row
justification:

| Row | Reader | Disposition | Run-phase re-verification |
|-----|--------|-------------|---------------------------|
| B1 | `LoadPure` → `loadLegacyBacklogJSON` | **keep-read-only** — pre-cutover readability + downgrade route; path derived via the A3/A5 seam | `backlog_migrate.go:511` unchanged on this HEAD |
| B2 | `migrateLegacyBacklog` (adopting path) | **keep** — adoption window; retirement owned by SPEC-TODO-QUEUE-HOME-MERGE-001 M5, not this SPEC | adoption machinery in `state_dir.go` untouched by this run |
| B3 | `backlog_export.go` / `todo_export.go` | **keep** — the documented `moai todo export-json` downgrade route | `todo_export.go:33-45` help text names the route verbatim |
| B4 | `queueExists` / `relocateQueueArtifacts` | **keep** until M5 retirement | reads legacy locations by design; untouched |
| B5 | `backlog_archive_vouch.go` classification | **keep** — naming only, no path read | unchanged |
| B6 | `cmd/t657-merge/main.go:51` | **boundary-exception** — merge CORE is card t835's; recorded, not converged | `filepath.Join(abs, "backlog.json")` present; exempted in the seam guard |
| B7 | `todo_disclosure.go` prose | **keep** — this IS the convergence messaging | unchanged |
| B8 | `statusline/backlog.go` (card premise) | **already-converged** — premise stale | re-verified: `git log` shows t306/t510 landed the seam; file delegates to `kanban.BacklogCountsForRoot` (`backlog.go:36,48`); no direct JSON read |
| B9 | `web/events.go` SHM :167-168, WAL :169-172; `screens_templ.go:1180` render comment | **keep** — SQLite-aware watch/render, not a JSON remnant | spans re-read on this HEAD; unchanged |

Zero blank rows. No reader beyond B1-B9 was discovered during the run
(REQ-004 run-phase sweep below).

### M2 — Foreman skill canonical-path correction (REQ-001/REQ-002)

`internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md` queue
watch rewritten: the fragment now derives the queue directory the way
`kanban.StateDirForRoot` does for a standard git-repository project —
`${MOAI_HOME:-$HOME/.moai}/db/<project-key>/todo`, project key = sanitized
primary-checkout basename + first 8 hex of sha256 over the canonical root
(`git worktree list` first entry → `pwd -P`), with a `sha256sum`/`shasum`
fallback chain. Checksum-poll change detection (db + WAL) is unchanged. The
prose below the fragment states the resolver agreement.

- Local copy synced: `diff -q .claude/skills/moai-kanban-foreman/SKILL.md internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md` → exit 0.
- Catalog hashes regenerated (`go run ./internal/template/scripts/gen-catalog-hashes.go --all`): only the `moai-kanban-foreman` entry changed
  (`ec5e5517…` → `9e2ebb72…`), confirmed via `git diff internal/template/catalog.yaml`.
- `make build` → exit 0.
- Neutrality greps on the template file: card ids 0, internal dates 0, 40-hex SHAs 0, `/Users/` paths 0, `state/todo` 0. The one `SPEC-` hit is the pre-existing `<SPEC-ID>` dispatch placeholder (identical count in HEAD).
- Emit surfaces verified, not assumed: `templates/.agents/skills/` carries only the 16 published `moai-<command>` skill files (no foreman); `templates/.codex/agents/moai/*.toml` are agent emissions (no foreman). `make agents-emit` / `make commands-emit` do not apply to this skill edit; `make build`'s `agents-emit-check` passed as part of the build.

### M3 — Guard tests (REQ-003) with mutation controls

New files: `internal/kanban/foreman_queue_statement_test.go` (docs-match,
AC-003a) and `internal/kanban/queue_path_seam_scan_test.go` (literal scan,
AC-003b). Fixture trap honored: both docs-side fixtures pin `MOAI_HOME` to a
fixture-local absolute path (SPEC-TODO-HOME-TEMP-GUARD-001's control pattern)
so a `t.TempDir()` origin still resolves home, and `HOME` is canaried.

**(a) RED on the defective doc (pre-fix, verbatim):**

```
$ go test ./internal/kanban/ -run 'TestForemanQueueWatchResolvesCanonicalStateDir' -count=1 -v
=== RUN   TestForemanQueueWatchResolvesCanonicalStateDir
    foreman_queue_statement_test.go:92: the skill names the project-local .moai/state/todo path — the dead watch target this SPEC removed
--- FAIL: TestForemanQueueWatchResolvesCanonicalStateDir (0.00s)
FAIL
```

The first post-fix run also caught a real fragment bug (trailing `-` in the
key: `001--31cbba72` vs `001-31cbba72` — a piped `basename` newline translated
by `tr -c`), fixed before green.

**(a) GREEN:**

```
$ go test ./internal/kanban/ -run 'TestForemanQueueWatchResolvesCanonicalStateDir' -count=1 -v
=== RUN   TestForemanQueueWatchResolvesCanonicalStateDir
--- PASS: TestForemanQueueWatchResolvesCanonicalStateDir (0.47s)
PASS
```

**(a) Mutation control (restore `d=.moai/state/todo`, verbatim):**

```
$ go test ./internal/kanban/ -run 'TestForemanQueueWatchResolvesCanonicalStateDir' -count=1 -v   # d= line mutated
    foreman_queue_statement_test.go:92: the skill names the project-local .moai/state/todo path — the dead watch target this SPEC removed
--- FAIL: ...  go-test-exit=1
# mutation reverted byte-identically → PASS again
```

**(b) GREEN (swept set counted):**

```
$ go test ./internal/kanban/ -run 'TestBacklogJSONLiteralStaysSeamScoped' -count=1 -v
    queue_path_seam_scan_test.go:93: exempt: internal/kanban/state_dir.go — the seam const every queue path enters through (backlogFileName)
    queue_path_seam_scan_test.go:93: exempt: cmd/t657-merge/main.go — one-off merge utility boundary exception, owned by card t835
    queue_path_seam_scan_test.go:101: swept 1218 non-test Go files under internal/, pkg/, cmd/; 2 file(s) carry the literal
--- PASS
```

**(b) Mutation control (inject `internal/cli/t658_scratch_violation.go` with
`filepath.Join(".", "backlog.json")`, verbatim):**

```
$ go test ./internal/kanban/ -run 'TestBacklogJSONLiteralStaysSeamScoped' -count=1 -v
    queue_path_seam_scan_test.go:91: internal/cli/t658_scratch_violation.go constructs the legacy queue-path literal "backlog.json" (1 occurrence(s)); the queue path is built only through the internal/kanban seam
    queue_path_seam_scan_test.go:101: swept 1219 non-test Go files under internal/, pkg/, cmd/; 3 file(s) carry the literal
--- FAIL: TestBacklogJSONLiteralStaysSeamScoped (0.18s)
go-test-exit=1
# scratch file deleted → exit 0 again
```

The scratch file was never committed.

### Existing t395 watch tests — adapted to the new contract

`foreman_queue_watch_test.go` / `foreman_queue_watch_wal_test.go`
(SPEC-BACKLOG-JSON-DISCLOSURE-001) build their fixture at the retired
project-local path, which the repaired script no longer watches — with the
temp-dir fixtures that disagreement was invisible before this SPEC and is the
exact acceptance.md fixture trap. `watchFixture` now creates a git-repo
fixture, pins `MOAI_HOME` to a fixture-local absolute path, and builds the
store at `StateDirForRoot`'s answer; the script re-derives the same directory
from the inherited environment. `dbOnlyWatchScript` (a synthetic control, not
shipped text) follows the queue via `MOAI_TEST_QUEUE_DIR`;
`legacyForemanWatchScript` keeps its historical verbatim pin — under the new
contract its silence is structural (the watched directory is absent), which is
this SPEC's defect statement, noted in the test header.

```
$ go test ./internal/kanban/ -run 'TestForemanQueueWatch' -count=1 -v
--- PASS: TestForemanQueueWatch_FiresOnMutation (local + template) x2
--- PASS: TestForemanQueueWatch_ShippedJSONTargetIsSilent
--- PASS: TestForemanQueueWatch_FiresWithStaleJSONPresent
--- PASS: TestForemanQueueWatch_WatchTargetsAgree
--- PASS: TestForemanQueueWatch_DBOnlyTargetMissesWALDeferral
--- PASS: TestForemanQueueWatch_SeesWALDeferredCommit
PASS  ok  github.com/modu-ai/moai-adk/internal/kanban 124.668s
```

### M4 — Verification sweep

| Claim | Command | Result |
|---|---|---|
| AC-005 retention | `go test ./internal/kanban/ -run 'TempOrigin' -count=1 -v` | 5/5 PASS, `TempOriginReason`/`StateDirForRoot` branches byte-unchanged (no `state_dir.go` diff in the run) |
| Affected package | `go test ./internal/kanban/ -count=1` | `ok ... 220.026s`, exit 0 |
| Catalog parity | `go test ./internal/template/ -run 'Catalog' -count=1` | `ok ... 0.453s`, exit 0 |
| Build | `go build ./...` | exit 0 |
| Vet | `go vet ./internal/kanban/` | exit 0 |
| Format | `gofmt -l internal/kanban/` | empty |
| REQ-006 scan | `grep -rn "state/todo\|state/kanban" internal/template/templates/` | 2 hits only: t704's owned `todo-queue-storage.md:22` (legacy NAMING, excluded per REQ-006) and the pre-fix foreman line (fixed under REQ-001). `.claude/skills/moai/SKILL.md:170` and `.claude/skills/moai/workflows/todo.md:26` re-read: correct home-canonical statements, untouched. `.claude/rules/`: no canonical claims. No new false statements discovered → REQ-006 ledger row: none to adjudicate. |

### Gaps

- Windows: the docs-match guard and the t395 watch tests are POSIX-sh and
  skip on GOOS=windows (`sh`/`cksum` LookPath guards). The fragment's
  sha256 fallback (`sha256sum` → `shasum`) is unmeasured on git-bash/Windows.
  CI on the darwin/linux legs is the verdict surface.
- The guard test pins the fragment's git fallback order (`worktree list`
  first-entry form) as an approximation of `CanonicalProjectRoot`'s
  separate-git-dir branch; equality is measured for standard and worktree
  repos, not for `--separate-git-dir` checkouts.
- Full-suite (`go test ./...`) intentionally not run locally (CLAUDE.local.md
  §4); CI on `origin/develop` after the lead's batch push is the full verdict.
- Non-ASCII repository basenames: the fragment's byte-wise `tr` sanitization
  diverges from Go's rune-wise mapping for multi-byte names (both tests use
  ASCII fixtures). Recorded as a known approximation, not a defect claim.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-13
run_commit_sha: "d26091f5b"
run_status: complete
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 4
l44_pre_commit_fetch: "n/a (isolated card worktree, no shared-checkout commit)"
l44_post_push_fetch: "n/a (lane does not push; lead batch-pushes develop)"
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin: "go build ./... exit 0"
cross_platform_build.windows: "not built locally; CI matrix is the verdict surface"
total_run_phase_files: 9
m1_to_mN_commit_strategy: "2 commits — (1) code+tests+template artifacts with the draft→in-progress transition, (2) this progress record"
```

AC matrix: AC-001 PASS (grep 0 + docs-match guard), AC-002 PASS (parity diff +
catalog hash refresh), AC-003 PASS (a/b guards + both mutation controls red
once), AC-004 PASS (ledger complete, zero blank rows), AC-005 PASS
(TempOrigin 5/5 + byte-unchanged branches), AC-006 PASS (scan clean; nothing
to adjudicate).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-13
sync_commit_sha: "da3e72ea5"   # backfilled (D3) — sync commit above
sync_status: complete
b12_self_test_a: "grep -c SPEC-TODO-QUEUE-HOME-CANON-001 CHANGELOG.md → 0 (pre-emission, safe)"
b12_self_test_b: "acceptance.md distinct AC = 6; §E.3 ac_pass_count = 6; match"
b12_self_test_c: "all CHANGELOG-cited paths ls-verified (.claude/skills/moai-kanban-foreman/SKILL.md)"
changelog_entry_position: "[Unreleased] → Fixed, first bullet"
frontmatter_status_transitions.in_progress_to_implemented: merged into sync commit
frontmatter_status_transitions.implemented_to_completed: merged into sync commit (3-phase close, no separate Mx commit)
frontmatter_updated_refresh: "2026-09-13 (all 4 artifacts unchanged date)"
mx_tag_validation: "sync sub-step — no new exported symbols in sync scope; template artifact only"
canary_compliance_check: "n/a — this SPEC defines no forward-looking policy for its own sync tests"
```
