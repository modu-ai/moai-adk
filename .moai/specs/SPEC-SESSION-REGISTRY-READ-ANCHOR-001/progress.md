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
| AC-RAR-008 | **NOT ATTEMPTED** | — | Directed by the lead to be recorded as not attempted, not as passed. S1 does not touch `internal/session/anchor.go`, so the disposal guard's behaviour is **unchanged by construction** — that is a property of the diff, NOT a verification: the per-coordinate refusal fixtures and the converse free-worktree control the criterion requires were not built and not run |
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

### Quality gates

| Check | Command | Output |
|---|---|---|
| Tests (affected packages) | `go test ./internal/hook/... ./internal/session/... -count=1` | `ok` for all 12 packages, `exited with code 0` (`internal/hook 261.102s`, `internal/session 22.229s`) |
| Vet | `go vet ./internal/hook/... ./internal/session/...` | `vet exit=0` |
| Lint | `golangci-lint run --timeout=5m ./internal/hook/... ./internal/session/...` | `0 issues.` `exit=0` |
| Coverage | `go test -cover ./internal/hook/ ./internal/session/ -count=1` | `internal/hook … coverage: 85.6% of statements`; `internal/session … coverage: 87.9% of statements` — both above the 85% target |
| Format | `gofmt -l internal/hook/` | (no output) |

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
- **AC-RAR-008 was not verified** (see the matrix row). Unchanged-by-construction
  is a diff property, not a guard measurement.
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
ac_not_attempted_count: 4 # AC-RAR-005,006,007,008 (005/006/007 S2-gated; 008 per lead direction)
preserve_list_post_run_count: 68  # of 69 orphan registry files byte-identical; the 69th is the live primary registry (§E.2 qualifier)
l44_pre_commit_fetch: not-performed   # lane does not push; integration is the lead's
l44_post_push_fetch: not-performed
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux: not-measured
  darwin: measured-implicitly-by-test-run
  windows: not-measured
total_run_phase_files: 2
m1_to_mN_commit_strategy: single-commit (S1 is one atomic repair; plan.md M3 only)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
