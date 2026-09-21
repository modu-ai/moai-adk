# SPEC-SESSION-REGISTRY-READ-ANCHOR-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1058 · worktree `.claude/worktrees/t1058` · branch `WT-read-anchor`
- Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md` (Tier M)
- Evidence base: `.moai/reports/t1058/pre-plan-measurement.md` (tree `3dfae918a`,
  measured 2026-09-21) with raw companions `orphan-file-list.txt`,
  `orphan-file-entry-counts.txt`, `orphan-entry-liveness.txt`,
  `live-pid-identity.txt`. No measurement re-run; no figure introduced that is
  absent from that record.
- Status: `draft`.
- **The card's premise was reversed by measurement and the SPEC is written to the
  measured state.** Anchoring R1 (`LiveAnchoredSessions`) does not neutralise the
  orphan registry files — it removes them from the disposal guard's view, and live
  orphan-only entries go with them. Anchoring R2 (`findRegistryUpwardFrom`) is a
  pure repair. The two are scoped apart (spec.md §C).
- Open at the Implementation Kickoff Approval gate: the S2 migration decision
  (plan.md M1, branches B1/B2/B3) and R2's evidence grade (plan.md M2). Neither is
  selected here.
- Evidence grades carried verbatim and NOT upgraded: Windows `stateanchor`
  behaviour is unmeasured (code reading only); R2's stall is synthetic-fixture
  grade, unreproduced against the real population; a process-existence probe
  identifies a process, not a session.

## §E.2 Run-phase Evidence

**Scope landed: S1 only (the R2 repair).** S2 was NOT approved and did NOT land.
Tree: worktree `.claude/worktrees/t1058`, branch `WT-read-anchor`, base HEAD
`74863c54e`. Every command below was run in this worktree, in this run.

### Change

`internal/hook/cwd_changed_relocate.go` — `relocateSessionCwd` now iterates a
candidate list of REGISTRY FILES rather than a list of candidate working
directories. `relocateRegistryCandidates` builds it in two passes: pass 1 is the
pre-existing upward walk over every candidate working directory (so a relocation
the old implementation already resolved resolves identically — AC-RAR-004), pass
2 adds the repository's PRIMARY registry per candidate, resolved through
`session.RegistryPathFor` (the same anchor seam the write path uses; that
function and `stateanchor` are untouched — REQ-RAR-009). The home boundary that
`findRegistryUpwardFrom` enforces is applied to pass 2 as well. The now-unused
`findRegistryUpward` wrapper was removed; its home-from-environment read moved
into `relocateRegistryCandidates`. Fail-open branches are unchanged.

`internal/hook/cwd_changed_relocate_anchor_test.go` — new.

### AC matrix

| AC | Status | Command | Actual output |
|---|---|---|---|
| AC-RAR-001 | PASS | `git diff --name-only` on the commit | `internal/hook/cwd_changed_relocate.go`, `internal/hook/cwd_changed_relocate_anchor_test.go` — no R1 coordinate (`internal/session/anchor.go`) present |
| AC-RAR-002 | PASS | `go test ./internal/hook/ -run TestRelocateReachesPrimaryRegistryWhenEveryCandidateIsInsideOneWorktree -count=1` | `ok github.com/modu-ai/moai-adk/internal/hook` |
| AC-RAR-003 | PASS | `go test ./internal/hook/ -run TestRelocateFailOpen -count=1` | `ok github.com/modu-ai/moai-adk/internal/hook` (3 sub-cases: absent, unreadable, entry-less) |
| AC-RAR-004 | PASS | `go test ./internal/hook/ -run 'TestCwdChanged\|TestFindRegistryUpward' -count=1` | `ok github.com/modu-ai/moai-adk/internal/hook`; the pre-existing relocation tests are unmodified (`git diff --name-only` does not list `cwd_changed_relocate_test.go`) |
| AC-RAR-005 | **NOT ATTEMPTED** | — | S2-gated; no S2 change was prepared, so no operator decision was owed or taken in this run |
| AC-RAR-006 | **NOT ATTEMPTED** | — | S2-gated; no landing-time population re-measurement was performed |
| AC-RAR-007 | **NOT ATTEMPTED** | — | S2-gated; this run makes no liveness claim about any registry entry |
| AC-RAR-008 | **FIXTURES ESTABLISHED — measured unchanged at S1 landing; behaviour under S2 UNVERIFIED.** Deliberately NOT recorded as PASS | `go test ./internal/cli/worktree/ ./internal/session/ -run TestACRAR008 -count=1 -v` | `ok …/internal/cli/worktree 0.381s`, `ok …/internal/session 0.326s`; 8/8 sub-cases PASS across the 4 consumer coordinates, both directions — § AC-RAR-008 fixtures below |
| AC-RAR-009 | PASS | `find … -path '*/.moai/state/active-sessions.json'` then `shasum -a 256`, before and after | 69 files before, 69 after; sorted content-hash comparison: **68 of 69 byte-identical**. The one difference is the repository's own LIVE primary registry (`/Users/goos/MoAI/moai-adk-go/.moai/state/active-sessions.json`, `8f6c2b59…` → `25edf7d2…`) — see the qualifier below |
| AC-RAR-010 | PASS | `git diff --name-only` | `internal/session/registry_path_anchor_test.go` absent from the changed-file list; `go test ./internal/session/ -count=1` → `ok … 22.229s` |
| AC-RAR-011 | PASS | this document | Windows behaviour of the anchor resolution and R2's reproduction against the real orphan population are both recorded as unmeasured below, at the grade spec.md §A.7 carries |
| AC-RAR-012 | PASS | command record | No `moai gtd` invocation of any kind appears in this run's command record |

**AC-RAR-009 qualifier.** The changed file is not an orphan: it is the primary
checkout's live registry, which other sessions on this host write continuously
(register / heartbeat / purge). Its post-state carries 172 entries and **no
fixture session id from this run's tests** (`sess-t1058-*`, `sess-other`,
`sess-someone-else` — none present). The tests write only inside `t.TempDir()`.
**Which writer produced the change is UNMEASURED** — only the endpoint hashes
were captured, not a writer trace, so "another live session wrote it" is the
attribution, not an observation.

The comparison was re-taken around the second commit (the AC-RAR-008 fixtures),
at execution time rather than against any stored list: 69 files before, 69
after, and this time **all 69 byte-identical** (`diff exit=0` over the sorted
`shasum -a 256` output). That does NOT retroactively revise the first commit's
result above — the live primary registry genuinely did change during the first,
longer window, and the two measurements are of two different windows. The second
window was short enough that no other session wrote in it; that is a fact about
the window, not evidence that the first change did not happen.

### Mutation controls

A passing test over an unexercised path is not evidence. Each control was applied
to the working tree, run, and reverted; the repaired file was restored from a
byte copy and re-verified (`grep -c MUTATION …` → `0`, tests green).

| # | Mutation | Expected | Observed |
|---|---|---|---|
| 1 | Pass 2 removed (`relocateRegistryCandidatesFrom` no longer adds the primary registry) — i.e. the repair reverted | AC-RAR-002 fails | `--- FAIL: TestRelocateReachesPrimaryRegistryWhenEveryCandidateIsInsideOneWorktree` — `primary registry CWD = ".../001", want ".../002/wt/internal/pkg" (the relocation never reached the primary registry)`; also `--- FAIL: TestRelocateFailOpen/unreadable_registry_is_skipped,_not_claimed` |
| 2 | Not-found fail-open guard removed (`if !found { continue }`) | AC-RAR-003 absent + entry-less cases fail | `--- FAIL: …/absent_registry_creates_nothing` — `a registry file was created at …/001 (stat err = <nil>)`; `--- FAIL: …/entry-less_registries_are_left_byte-identical` — `worktree registry rewritten: before: [] after: [\n]` |
| 3 | Both fail-open guards removed (read-error `continue` and not-found `continue`) | AC-RAR-003 unreadable case fails | `--- FAIL: …/unreadable_registry_is_skipped,_not_claimed` — `primary registry CWD = ".../001", want ".../002/wt/sub"` |

The RED state was captured before any implementation: on base tree `74863c54e`
the new tests failed with
`primary registry CWD = "…/001", want "…/002/wt/internal/pkg" (the relocation
never reached the primary registry)` and the matching unreadable-case failure,
while the absent and entry-less sub-cases passed (they are the fail-open
behaviours the repair must preserve, so their controls are the code mutations
above rather than a pre-implementation RED).

**One mutation claim was corrected during the run.** An earlier test comment
asserted that removing the read-error guard alone stops the walk. It does not:
with the not-found guard still standing, a failed read yields no entries and the
walk continues. The comment now states the mutation that is actually observable
(control 3).

Controls 4 and 5 belong to the AC-RAR-008 fixtures and were run in the second
commit; they are listed with that section below and are part of the same table.

### AC-RAR-008 fixtures (second commit)

**[HARD] This is NOT recorded as PASS.** The accurate statement: *fixtures
established, and the live-anchored-worktree judgement measured as unchanged at
S1 landing; behaviour under S2 is unverified.* The first pass recorded
"unchanged by construction" — a property of the diff. These fixtures convert
that into a MEASURED invariance and pre-place the regression guard that will
catch S2 (anchoring `LiveAnchoredSessions`) breaking the disposal path, which
is where the guard needs to be, since S2 is the change that loses live sessions
from view. They do not reach past that: S2 is not in this run.

Two new test files, no production file touched:

- `internal/cli/worktree/anchor_disposal_guard_ac_rar_008_test.go` — the three
  CLI coordinates (`remove.go:51`, `done.go:77` auto-mode, `done.go:176`
  interactive).
- `internal/session/anchor_disposal_guard_ac_rar_008_test.go` — the fourth
  coordinate, `anchor_lock.go:111` (`AnchorDecision`'s registry branch), which
  is not reachable from the CLI package's surface. The lock source is held
  silent (`LockInfo{}` carries NO opinion) so the verdict is attributable to
  the registry branch alone.

Each coordinate carries BOTH directions. The converse control is the
load-bearing half: a fixture measuring only "refuses disposal" is satisfied by
an implementation that refuses everything — the empty-green shape this SPEC
exists to prevent. The converse uses a **present-but-dead** registry row (PID 0,
rejected at `LiveAnchoredSessions`' `e.PID > 0` guard without an OS probe;
heartbeat two hours old, past `DefaultStaleMinutes`) rather than an absent
registry, so it shows the guard evaluates LIVENESS and not mere presence. Every
case pins `CLAUDE_PROJECT_DIR` at an empty directory, so the caller-registry
root cannot contribute a real entry from the host and a green result stays
attributable to the fixture.

| Coordinate | Direction | Result |
|---|---|---|
| `remove.go:51` | live anchored → refused | PASS (`ANCHORED_SESSIONS_PRESENT`, `Remove()` not called) |
| `remove.go:51` | converse: no live session → free | PASS (`Remove()` called, no error) |
| `done.go:77` (auto) | live anchored → skips removal | PASS (not-done reported, no error, `Remove()` not called) |
| `done.go:77` (auto) | converse: no live session → free | PASS (done reported, `Remove()` called) |
| `done.go:176` (interactive) | live anchored → refused | PASS (`ANCHORED_SESSIONS_PRESENT`, `Remove()` not called) |
| `done.go:176` (interactive) | converse: no live session → free | PASS (`Remove()` called, no error) |
| `anchor_lock.go:111` | live anchored → `Anchored=true`, source registry | PASS |
| `anchor_lock.go:111` | converse: dead entry → `Anchored=false`, source none | PASS |

Verbatim: `ok github.com/modu-ai/moai-adk/internal/cli/worktree 0.381s` and
`ok github.com/modu-ai/moai-adk/internal/session 0.326s`, all 8 sub-cases
`--- PASS`.

**Sensitivity — the fixtures were seen red in both directions.** A green fixture
never observed failing proves nothing, so `internal/session/anchor.go` was
mutated in the working tree, run, and restored (restoration verified by content
hash `8ce8f8b0d2767b554d9df5404f7a0f561e72dfb870ce7a20e4ae8ef56a5d7c35`, and
`git status` confirms the file is absent from both commits).

| # | Mutation | Expected | Observed |
|---|---|---|---|
| 4 | `LiveAnchoredSessions` returns `nil` — the S2 failure shape, live sessions leaving the guard's view | all 4 refusal halves fail, all 4 converses still pass | `--- FAIL` on `TestACRAR008_RemoveCoordinate/live_anchored_session_refuses_disposal`, `…_DoneAutoCoordinate/live_anchored_session_skips_removal`, `…_DoneInteractiveCoordinate/live_anchored_session_refuses_disposal`, `…_AnchorDecisionRegistryCoordinate/live_anchored_session_anchors_the_tree`; no converse sub-case failed |
| 5 | `alive := true` — refuse-everything, liveness never evaluated | all 4 converses fail, all 4 refusal halves still pass | `--- FAIL` on all four `converse_control:` sub-cases (`…_RemoveCoordinate`, `…_DoneAutoCoordinate`, `…_DoneInteractiveCoordinate`, `…_AnchorDecisionRegistryCoordinate/converse_control:_a_dead_entry_leaves_the_tree_free`); no refusal sub-case failed |

Controls 4 and 5 are exact complements: each half of every coordinate fails
under exactly one of them and passes under the other. That is what establishes
both directions are load-bearing rather than decorative.

### Quality gates

| Check | Command | Output |
|---|---|---|
| Tests (affected packages) | `go test ./internal/hook/... ./internal/session/... -count=1` | `ok` for all 12 packages, `exited with code 0` (`internal/hook 261.102s`, `internal/session 22.229s`) |
| Vet | `go vet ./internal/hook/... ./internal/session/...` | `vet exit=0` |
| Lint | `golangci-lint run --timeout=5m ./internal/hook/... ./internal/session/...` | `0 issues.` `exit=0` |
| Coverage | `go test -cover ./internal/hook/ ./internal/session/ -count=1` | `internal/hook … coverage: 85.6% of statements`; `internal/session … coverage: 87.9% of statements` — both above the 85% target |
| Format | `gofmt -l internal/hook/` | (no output) |

Second commit (the AC-RAR-008 fixtures), re-run against its own packages:

| Check | Command | Output |
|---|---|---|
| Tests | `go test ./internal/cli/worktree/... ./internal/session/... -count=1` | `ok …/internal/cli/worktree 4.074s`; `ok …/internal/session 6.173s` |
| Vet | `go vet ./internal/cli/worktree/... ./internal/session/...` | `vet exit=0` |
| Lint | `golangci-lint run --timeout=5m ./internal/cli/worktree/... ./internal/session/...` (run inside the worktree; `pwd` confirmed in the same invocation) | `0 issues.` `lint exit=0` |
| Format | `gofmt -l internal/cli/worktree/ internal/session/` | (no output) |

Exported evidence: `.moai/reports/t1058/{orphan-hashes-pre.txt,
orphan-hashes-post.txt,lint-wt.txt,cover.txt}`. **That directory is gitignored in
this repository (`.gitignore:227`, an operator decision), so those files are
machine-local and reach no clone.** The load-bearing evidence is therefore
carried in this tracked section rather than only by citation.

### Gaps — explicitly NOT observed

- **The full test suite was not run locally** (repo policy; CI owns the
  full-suite verdict). Only `./internal/hook/...` and `./internal/session/...`
  were run.
- **Cross-platform build was not measured.** No `GOOS=windows` build was run in
  this run.
- **AC-RAR-008's behaviour under S2 is UNVERIFIED.** The fixtures measure the
  current state and pre-place the regression guard; S2 is not in this run, so
  nothing here establishes how the guard behaves once `LiveAnchoredSessions`
  is anchored. This is a narrower gap than the first pass recorded — the diff
  property ("unchanged by construction") has been converted into a measured
  one — but it is not closed.
- **Windows behaviour of the anchor resolution is UNMEASURED, not absent.**
  `session.RegistryPathFor` → `stateanchor.FromDirectory` carries no
  `runtime.GOOS` branching; that is a code reading, carried verbatim from
  spec.md §A.7, and this run added no execution measurement on Windows. The
  repair calls that seam and does not change it.
- **R2's stall remains UNREPRODUCED against the real orphan population**
  (spec.md Gaps G3). The new test is a synthetic git-repo-plus-linked-worktree
  fixture. Plan.md M2's choice was therefore taken as "leave the grade and say
  so" — this is that statement, not an upgrade.
- **Which writer changed the live primary registry is unmeasured** (AC-RAR-009
  qualifier above).
- **One authoring command was REFUSED by the worktree-isolation guard**: a
  compound `cat > … <<EOF … EOF && go test …` heredoc whose body named git
  commands. The test file was then written with the Write tool and the test run
  issued as a separate command; no measurement was substituted or inferred.

### Residual risk

- Pass 2 spawns `git rev-parse` (through `stateanchor.FromDirectory`) once per
  distinct candidate working directory — at most three per CwdChanged event, and
  only on the enter/exit transition, not per tool call. Not measured against the
  hook time budget in this run.
- `primaryRegistryFor` recovers the anchor root by stripping three path
  components off `RegistryPathFor`'s result. That is correct while
  `DefaultRegistryPath` has three components; a change to that constant would
  silently misalign the home-boundary comparison. No test pins the coupling.
- The repair widens where a relocation may WRITE: a session whose entry lives in
  the primary registry is now reachable from inside a worktree. That is the
  intended behaviour, and it writes only to an entry already carrying that
  session id.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-21
run_commit_sha: pending-backfill-run
run_status: audit-ready
run_scope: S1-only
ac_pass_count: 8          # AC-RAR-001,002,003,004,009,010,011,012
ac_fail_count: 0
ac_not_attempted_count: 3 # AC-RAR-005,006,007 — all three S2-gated
ac_fixtures_established_count: 1 # AC-RAR-008 — measured unchanged at S1 landing; S2 behaviour unverified. NOT counted as a pass
preserve_list_post_run_count: 68  # of 69 orphan registry files byte-identical; the 69th is the live primary registry (§E.2 qualifier)
l44_pre_commit_fetch: not-performed   # lane does not push; integration is the lead's
l44_post_push_fetch: not-performed
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux: not-measured
  darwin: measured-implicitly-by-test-run
  windows: not-measured
total_run_phase_files: 4
m1_to_mN_commit_strategy: two commits on WT-read-anchor — (1) the R2 repair, (2) the AC-RAR-008 disposal-guard fixtures added on a lead scope addition; no amend, both reported
```

## §E.4 Sync-phase Audit-Ready Signal

**Scope closed: S1 only.** S2 (anchoring `LiveAnchoredSessions`) was not
approved at Implementation Kickoff Approval and did not land in this SPEC;
nothing in this sync phase changes that. Sync-phase evidence was read from
the lane-independent re-measurement files
(`.moai/reports/t1058/lane-independent-{lint,run-verify,ac008-verify}.txt`),
not solely from manager-develop's own report, per the SPEC's grading
discipline (spec.md §A.7 / `verification-claim-integrity.md`).

### Deliverables

- `CHANGELOG.md` — one entry under `[Unreleased] / ### Fixed`, sized to an
  internal behaviour fix with no user-facing API change: the R2 relocation
  repair, the deliberately-not-landed R1/S2 scope with its rationale, the
  8/12 PASS + 3 NOT ATTEMPTED + 1 fixtures-established AC breakdown, and the
  unattributed live-primary-registry write named as a known gap rather than
  rounded away.
- `README.md` / docs-site — **NOT touched.** Grepped for the changed symbols
  (`findRegistryUpward`, `relocateSessionCwd`, `cwd_changed_relocate`,
  `session registry`, `active-sessions.json`) across `README*.md` and
  `docs-site/`; the only hit is an unrelated alt-text string on a web-console
  screenshot (`README.md:138`). Nothing documented there describes this
  internal relocation behaviour, so no edit was made — changing nothing is
  the correct outcome when nothing is stale, not an omission.
- `spec.md` frontmatter — `status: in-progress → completed` on this commit.
  `updated:` was already `2026-09-21` (same day as `created:`) in all three
  of `spec.md` / `plan.md` / `acceptance.md`, so no date edit was needed to
  bring it current. No body content in `spec.md` / `plan.md` / `acceptance.md`
  was modified.
- This §E.4 section.

### B12 self-test (mandatory, before CHANGELOG emission)

1. **Pre-emission grep**: `grep -c 'SESSION-REGISTRY-READ-ANCHOR' CHANGELOG.md`
   → `0` before emission, `1` after — no duplicate entry.
2. **AC count match**:
   `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l` → `12`,
   matching the 12 distinct `AC-RAR-NNN` identifiers the CHANGELOG entry's
   breakdown (8 PASS + 3 NOT ATTEMPTED + 1 fixtures-established) sums to.
3. **File path verification**: `ls` confirmed all four files the entry names
   exist — `internal/hook/cwd_changed_relocate.go`,
   `internal/hook/cwd_changed_relocate_anchor_test.go`,
   `internal/cli/worktree/anchor_disposal_guard_ac_rar_008_test.go`,
   `internal/session/anchor_disposal_guard_ac_rar_008_test.go`.

### Process notes (recurrence record)

- **A lead instruction named AC-RAR-005 where AC-RAR-006 was meant** (intent
  — "re-measure at landing time" — carried correctly; the AC number did not).
  The run-phase lane caught it against `acceptance.md`'s own text and the
  lead accepted the correction; both AC entries in this section and in
  `progress.md` §E.2 name the criterion the text actually describes.
- **The AC-RAR-008 fixture scope was added mid-run, by lead decision, after
  the first run-phase pass had already reported it NOT ATTEMPTED** (correctly
  — it stated the true state at that pass rather than a placeholder). The
  fixtures then landed as a second, separate commit rather than an amend of
  the first, so the already-reported first-commit SHA never moved.

### Gaps — explicitly NOT observed in the sync phase

- **AC-RAR-008's behaviour under S2 is unverified** (carried unchanged from
  §E.2 — S2 is not in this SPEC's landed scope).
- **Windows behaviour of the anchor resolution remains unmeasured**, per
  spec.md §A.7 and `internal/stateanchor/stateanchor.go`'s absence of
  `runtime.GOOS` branching — a code reading, not an execution measurement,
  and not upgraded in sync phase.
- **R2's upward-walk stall remains unreproduced against the real orphan
  population** (spec.md Gaps G3) — synthetic-fixture-plus-code-reading grade,
  carried verbatim.
- **The disposition of the sibling orphan-file-deletion card (A6) is not
  decided here** (REQ-RAR-012, AC-RAR-012) — no `moai gtd` mutation was
  issued by this sync phase; any statement above about that card is a
  finding for the operator, not a queue action.
- **The CHANGELOG entry's characterization of the S2 rationale is drawn from
  spec.md §A.3, not independently re-measured in sync phase** — this sync
  phase performed no new measurement of the orphan population; it read and
  summarized the run-phase evidence.

```yaml
sync_complete_at: 2026-09-21
sync_commit_sha: pending-backfill-sync
sync_status: audit-ready
b12_self_test_a: pass
b12_self_test_b: pass
b12_self_test_c: pass
changelog_entry_position: "[Unreleased] / ### Fixed, first entry"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
  plan_md: "n/a (stateless on status axis)"
  acceptance_md: "n/a (stateless on status axis)"
canary_compliance_check: "n/a — this SPEC defines no forward-looking policy tested by its own sync"
```

