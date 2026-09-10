# Progress — SPEC-UPDATE-MERGE-CONFLICT-BLIND-001

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set: Tier M — `spec.md` + `plan.md` + `acceptance.md` (+ this `progress.md`).
- SPEC ID regex check executed as Bash; observed output: `PASS`.
- ID uniqueness: `ls .moai/specs/SPEC-UPDATE-MERGE-CONFLICT-BLIND-001` reported
  `No such file or directory` before authoring — no collision.
- Frontmatter: canonical 12 fields present in `spec.md`, `status: draft`.
- Authoritative input: `.moai/reports/t576/verdict.md` (read in full).
- Not executed at plan-phase: no test, no build, no production-code change.
  `spec.md` §A is a hypothesis until M1 lands (`spec.md` §A.5).

## §E.2 Run-phase Evidence

### M1 — the reproduction (no production code change)

- **PKG** (fixed by M1, per `acceptance.md` preamble): `./internal/cli/update/merge/...`.
  The base derivation the measurement must exercise (`deriveTemplateBase` /
  `derive` / `pruneToShared`) is unexported in that package, so the harness lands
  there and reaches the engine through `internal/merge`'s exported `MergeFile`.
- **Harness**: `internal/cli/update/merge/conflict_blind_repro_test.go` (new file;
  no production file modified).
- **Baseline attribution**: worktree `.claude/worktrees/t576`, branch
  `WT-permissions-ask-empty`, tree `9cf415332` plus the one new untracked test
  file above. All readings below were observed in this run, against this tree.

#### The four control cells — one run, one fixture

Command:

```
go test ./internal/cli/update/merge/... -run 'TestSharedKeyControlCells' -v -count=1
```

Observed (verbatim):

```
=== RUN   TestSharedKeyControlCells
=== RUN   TestSharedKeyControlCells/untouched_shared
    conflict_blind_repro_test.go:165: cell untouched_shared: key=untouched_shared written=["template-old"] HasConflict=false len(Conflicts)=0 predicted=["template-old"] (shared key: derived base carries the template's value, so the template's change reads as no change and the user's value stands)
=== RUN   TestSharedKeyControlCells/emptied_shared
    conflict_blind_repro_test.go:165: cell emptied_shared: key=emptied_shared written=[] HasConflict=false len(Conflicts)=0 predicted=[] (shared key emptied by the user: the empty array is the user's change and is preserved)
=== RUN   TestSharedKeyControlCells/changed_shared
    conflict_blind_repro_test.go:165: cell changed_shared: key=changed_shared written=["user-choice"] HasConflict=false len(Conflicts)=0 predicted=["user-choice"] (shared key changed by the user: the user's value is preserved)
=== RUN   TestSharedKeyControlCells/omitted
    conflict_blind_repro_test.go:165: cell omitted: key=omitted_key written=["template-only"] HasConflict=false len(Conflicts)=0 predicted=["template-only"] (key absent from the user's side is not shared, so it stays out of the base and the template reads as introducing it)
--- PASS: TestSharedKeyControlCells (0.00s)
    --- PASS: TestSharedKeyControlCells/untouched_shared (0.00s)
    --- PASS: TestSharedKeyControlCells/emptied_shared (0.00s)
    --- PASS: TestSharedKeyControlCells/changed_shared (0.00s)
    --- PASS: TestSharedKeyControlCells/omitted (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.645s
```

| Cell | Prediction (`plan.md` §F) | Value written | `HasConflict` | `len(Conflicts)` | Held? |
|---|---|---|---|---|---|
| (i) `untouched_shared` | user's value stands | `["template-old"]` | `false` | `0` | yes |
| (ii) `emptied_shared` | `[]` preserved | `[]` | `false` | `0` | yes |
| (iii) `changed_shared` | user's value preserved | `["user-choice"]` | `false` | `0` | yes |
| (iv) `omitted` — discriminator | template's value lands | `["template-only"]` | `false` | `0` | yes |

Cell (iv) landing the template's value is what separates a merge that ran from a
file that was left alone: a harness that copied the user's document would have
written nothing for `omitted_key`, and the sub-test fails on an absent key rather
than reporting a pass. All four sub-tests appear in the single run — no cell is
absent, so no cell is a gap.

#### The conflict surface — measured, and measured against a control

Command:

```
go test ./internal/cli/update/merge/... -run 'TestConflictSurfaceReachability' -v -count=1
```

Observed (verbatim):

```
=== RUN   TestConflictSurfaceReachability
=== RUN   TestConflictSurfaceReachability/derived_base
    conflict_blind_repro_test.go:196: divergent shared key, base derived as the update path derives it: written="user-value" HasConflict=false len(Conflicts)=0 (prediction: false, 0)
=== RUN   TestConflictSurfaceReachability/genuine_base_control
    conflict_blind_repro_test.go:226: same key and same engine, base differing from both sides: HasConflict=true len(Conflicts)=1 (prediction: true, 1)
--- PASS: TestConflictSurfaceReachability (0.00s)
    --- PASS: TestConflictSurfaceReachability/derived_base (0.00s)
    --- PASS: TestConflictSurfaceReachability/genuine_base_control (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.543s
```

Both readings come from `MergeResult` as the engine returned it. The base is not
re-derived and compared to the value it was derived from, which would assert
nothing.

The control is the load-bearing half. "No conflict was reported" and "this
harness never reads a conflict" produce the same two readings, so a bare `false`
/ `0` would be uninterpreted output. The control runs the **same engine over the
same key** with a base that agrees with neither side, and the surface fires:
`HasConflict=true`, `len(Conflicts)=1`. The `false` / `0` above is therefore a
property of the derived base, not of the instrument.

#### What M1 establishes, and what it does not

- **Established.** `spec.md` §A.4's healing boundary is real and measured: an
  omitted key heals (cell iv), an emptied one does not (cell ii). The
  shared-key conflict surface is unreachable through the derived base while
  being reachable through a base that can differ from `updated` — the two
  sub-tests above differ in exactly that input. `spec.md` §A.5's hypothesis
  marker is discharged for §A.1-§A.4 at the shared-key level.
- **Not established.** Nothing here measures `permissions.ask` in any real
  checkout, nothing measures `moai update` end to end, and nothing measures
  nested-key or YAML behaviour (see Gaps in the M1 report). No production code
  changed, so no `REQ-UMC-008`/`009`/`011` obligation is discharged; M2's design
  choice stays open on this evidence.

#### Scoped verification

```
$ go test ./internal/merge/... ./internal/cli/update/merge/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/merge	0.395s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	0.371s
exit 0

$ go vet ./internal/merge/... ./internal/cli/update/merge/...
(no output)  exit 0

$ golangci-lint run --timeout=5m ./internal/merge/... ./internal/cli/update/merge/...
0 issues.
exit 0
```

Neither package reported `[no test files]`. The full-suite verdict is CI's, not
this run's (`CLAUDE.local.md` §4).

#### AC matrix — M1

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-UMC-001 | PASS | `grep -n 't\.TempDir()' <file>` / `grep -nE '/Users/\|\.claude/settings\.json"\|/moai/moai-adk-go' <file>` | `37:	dir := t.TempDir()` / no match (rc 1) |
| AC-UMC-002 | PASS | `grep -nE 'exec\.Command\|RunUpdate\|runUpdate' <file>` | no match (rc 1) |
| AC-UMC-003 | PASS | `go test $PKG -run 'TestSharedKeyControlCells/untouched_shared' -v -count=1` | `--- PASS`; wrote `["template-old"]` |
| AC-UMC-004 | PASS | `go test $PKG -run 'TestSharedKeyControlCells/emptied_shared' -v -count=1` | `--- PASS`; wrote `[]` (not the template's non-empty value) |
| AC-UMC-005 | PASS | `go test $PKG -run 'TestSharedKeyControlCells/changed_shared' -v -count=1` | `--- PASS`; wrote `["user-choice"]` |
| AC-UMC-006 (MUST) | PASS | `go test $PKG -run 'TestSharedKeyControlCells/omitted' -v -count=1` | `--- PASS`; template's value `["template-only"]` landed |
| AC-UMC-007 (MUST) | PASS | `go test $PKG -run 'TestSharedKeyControlCells' -v -count=1` | exactly 4 sub-tests, each with a `--- PASS` line |
| AC-UMC-008 (MUST) | PASS | `go test $PKG -run 'TestConflictSurfaceReachability' -v -count=1` | `HasConflict=false len(Conflicts)=0` on the derived base; control `true` / `1` |
| AC-UMC-015 (MUST) | PASS | `git status --porcelain -- .claude/settings.json internal/template/templates/.claude/settings.json.tmpl` | no output (both untouched; `REQ-UMC-014` not reached) |
| AC-UMC-016 | N/A | `git diff --name-only` | no settings document changed in M1 — recorded N/A, not PASS |
| AC-UMC-017 (MUST) | PASS | `go test ./internal/merge/... ./internal/cli/update/merge/... -count=1` | two `ok` lines, exit 0, no `[no test files]` |
| AC-UMC-009 / 010 / 011 | not attempted | — | M2 scope; not run in this milestone |
| AC-UMC-012 / 013 / 014 | not attempted | — | M3 scope; not run in this milestone |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: m1-complete            # NOT run-complete — M2 and M3 are unstarted
m1_complete_at: 2026-09-10
m1_commit_sha: pending-backfill    # M1 files authored; commit is the lead's
pkg_selected: ./internal/cli/update/merge/...
ac_pass_count: 10                  # AC-UMC-001..008, 015, 017
ac_fail_count: 0
ac_na_count: 1                     # AC-UMC-016 (no settings document changed)
ac_not_attempted_count: 6          # AC-UMC-009..014 (M2/M3 scope)
production_files_changed: 0
new_test_files: 1
new_warnings_or_lints_introduced: 0
preserve_list_post_run_count: 0    # no PRESERVE-listed file modified
run_complete_at: pending           # awaits M2 + M3
```

Not measured in M1, and therefore gaps rather than passes: cross-platform build
(`GOOS=windows`), coverage delta, nested-key and YAML behaviour of the derived
base, and any end-to-end `moai update` path.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

🗿 MoAI
