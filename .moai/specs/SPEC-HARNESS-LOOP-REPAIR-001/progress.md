# SPEC-HARNESS-LOOP-REPAIR-001 — Progress

## Phase state

| Phase | State | Evidence |
|---|---|---|
| Audit | complete | findings recorded in `spec.md` §A.2 (all values measured on the live tree 2026-07-27) |
| Plan (spec.md) | complete | commit `71ef81809` — 10 REQ / 13 AC / 6 milestones |
| Plan (plan.md, acceptance.md) | **not written** | orchestrator-direct authoring; agent spawning was constrained this session |
| plan-audit | **not run** | no independent `plan-auditor` verdict exists |
| Kickoff Approval | **obtained** | user selected "M1 바로 착수" at the AskUserQuestion gate |
| M1 implementation | complete | shared accessor + C1/C2/C3 rewired; see §E.2 |

## Decision taken this session

**Proposal layout contract → normalise on the NESTED directory form** (user decision).
Consumers are changed to read `proposals/<ID>/proposal.json`; the producer is unchanged.
This keeps the 52 live drafts visible and preserves `spec.md` (human body) alongside
`proposal.json` (machine metadata). Resolves `spec.md` §G open question 1.

`spec.md` §G questions 2 (promotion narrowing rule) and 3 (lesson store direction)
remain **open** and belong to M4 / M5 respectively.

## M1 — exact fix sites (verified by reading source)

Producer (unchanged): `internal/harness/proposalgen/scaffolder.go:94,101,106`
— `os.MkdirAll(draftDir)` then writes `spec.md` + `proposal.json` inside it.

Three consumers, all assuming a flat `<ID>.json`:

| # | File | Symbol | Current predicate |
|---|---|---|---|
| C1 | `internal/cli/harness.go` 186-198 | `countProposals` | `!e.IsDir() && strings.HasSuffix(e.Name(), ".json")` |
| C2 | `internal/cli/harness.go` 270-274 | apply selector | same predicate |
| C3 | `internal/cli/harness/execute.go` 192-203 | `resolveProposalPath` | `filepath.Join(root, dir, id+".json")` |

Shared constant already present: `harnessDefaultProposalDir = ".moai/harness/proposals"`
(`internal/cli/harness.go:41`) and `execProposalDirRel` (`execute.go:37`) — the same
literal declared twice; the shared accessor should collapse that duplication too.

Package note: `internal/cli` (package `cli`) already imports `internal/cli/harness`
(package `harnesscli`) — it calls `harnesscli.RunExecute`. Either that package or
`internal/harness/proposalgen` can host the shared accessor; placing it next to the
producer in `proposalgen` binds the layout definition to the code that writes it.

## M1 verification (must be run, not assumed)

Reproduction-first per CLAUDE.md Rule 4 — write the failing test BEFORE the fix:

1. Fixture: temp project with `proposals/PROPOSAL-TEST-001/proposal.json`
2. Assert `countProposals` returns 1 → **must FAIL before the fix** (baseline returns 0)
3. Apply the shared accessor to C1/C2/C3
4. Assert the test passes
5. Live check — **must target the primary checkout**, see the trap below:
   `moai harness status --project-root /Users/goos/MoAI/moai-adk-go`
   → `pending proposals: 52 items` (baseline: `0 items`)
6. Full suite `go test ./...` + `go vet ./...` + `golangci-lint run`

### Trap: the 52 drafts are gitignored and exist only in the primary checkout

`.gitignore:229` ignores `.moai/harness/proposals/`, so the worktree has **0** draft
directories while the primary checkout has **52**. Running `moai harness status` from
inside the worktree therefore reports `0 items` **even after a correct fix**, which reads
as a failed repair. Two consequences:

- Unit fixtures MUST build their own `proposals/<ID>/proposal.json` under `t.TempDir()`
  (this is the required practice anyway per CLAUDE.local.md §6 Test Isolation).
- The live 52-count check MUST pass `--project-root` pointing at the primary checkout,
  and MUST use a binary rebuilt from this worktree (`go build`), not the installed `moai`.

Falsification for AC-HLR-006: a regression test must fail if any call site
reintroduces an independent `id + ".json"` path derivation.

## Out of scope (recorded so it is not re-litigated)

- **goal surface** — audited and found fully wired: `moai goal` CLI, `moai hook stop-goal`,
  `handle-stop-goal.sh` registered on the `Stop` event, and template mirrors present
  (`goal.md.tmpl`, `handle-stop-goal.sh.tmpl`, `settings.json.tmpl` ×2 registrations).
  Only gap is adoption — `.moai/state/goal/` has held 0 files since 2026-07-24.
  SPEC-GOAL-DOCS-RETIRE-001 is concurrently active on this surface.
- Wholesale observation-schema redesign — M4 narrows promotion, not recording.
- Retroactive rewrite of the 52 existing drafts.

## Environment constraints carried forward

- The primary checkout was on `sync/spec-goal-docs-retire-001` (another session's branch)
  with that session's uncommitted files. All work for this SPEC happens in the isolated
  worktree `.claude/worktrees/harness-loop-repair` (branch `feat/SPEC-HARNESS-LOOP-REPAIR-001`,
  based on `origin/main` = `760f09f73`). Do not commit this SPEC's work in the primary checkout.
- `moai harness ledger record` was exercised once this session and works (exit 0, one row
  written). The ledger was empty purely because no orchestrator had ever called it — M3.

## §E.2 Run-phase Evidence — M1

Every row below was produced by running the named command in this run, against
this tree. Verbatim logs: `.moai/state/verify/m1/`.

### Change set

| File | Change |
|---|---|
| `internal/harness/proposalgen/layout.go` | NEW — the shared accessor (`ProposalDirRel`, `MetadataFileName`, `ProposalDir`, `ListDraftIDs`, `ProposalPath`) placed next to the producer that defines the layout |
| `internal/cli/harness.go` | C1 `countProposals` + C2 apply selector routed through the accessor; `harnessDefaultProposalDir` now aliases `proposalgen.ProposalDirRel` (duplicate literal removed) |
| `internal/cli/harness/execute.go` | C3 `resolveProposalPath` delegates path + traversal validation to the accessor; `execProposalDirRel` aliases the canonical const; `draftLabel` added so diagnostics name the draft ID rather than the shared `proposal.json` basename |
| `internal/cli/harness_layout_repro_test.go` | NEW — C1/C2 reproduction + AC-HLR-006 flat-derivation guard |
| `internal/cli/harness/layout_repro_test.go` | NEW — C3 reproduction, traversal guard, M2 schema characterization |
| `internal/cli/harness/execute_test.go`, `internal/cli/harness_execute_test.go` | fixtures migrated from the retired flat `<id>.json` to nested `<id>/proposal.json` |

### AC verification

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-HLR-001 | PASS | `moai harness status --project-root <primary>` (binary built from this worktree) | `pending proposals: 52 items` (baseline `0 items`) |
| AC-HLR-002 | PASS | `moai harness apply --project-root <primary>` | emitted the `PROPOSAL-20260617-dc05149f` payload, not `No pending proposals` |
| AC-HLR-003 | PASS | `moai harness execute --id PROPOSAL-20260617-dc05149f --project-root <isolated copy>` | no longer `proposal not found`; reaches the file and fails at the schema layer (exit 1) — see Blocker below |
| AC-HLR-006 | PASS | `go test -run TestNoFlatProposalPathDerivation ./internal/cli/` | no call site re-derives `id + ".json"` |

### Falsification (both directions verified, not assumed)

- Inverting the directory predicate in `ListDraftIDs` → `TestCountProposals_NestedLayout` reports `countProposals = 0, want 2`; `TestHarnessApply_NestedLayout` fails.
- Restoring `draftID+".json"` in `ProposalPath` → `TestResolveProposalPath_NestedLayout` and `TestLoadProposalByID_NestedLayout` fail.
- Both edits were reverted and the suite re-confirmed green.

### Full verification batch

`go test ./...` exit 0 (105 packages ok) · `go vet ./...` exit 0 ·
`golangci-lint run` exit 0 (0 issues) · `GOOS=windows` and `GOOS=linux`
builds exit 0 · subagent-boundary grep: 0 invocations (19 matches are prose
comments and flag descriptions only).

### Blocker discovered — belongs to M2, NOT fixed here

The producer and the consumer carry two different schemas, independent of the
layout seam:

- `proposalgen` writes `"tier": "auto_update"` (string); `harness.Proposal.Tier`
  is the numeric `harness.Tier` → hard unmarshal error.
- The producer emits no `target_path` / `field_key` / `new_value`, so
  `Applier.Apply()` would have nothing to apply even if parsing succeeded.

Repairing the layout makes drafts VISIBLE; it does not make them APPLICABLE.
AC-HLR-004 (first `apply_outcome`) is therefore blocked on reconciling these
schemas — a design decision (change the producer, or add a mapping layer) that
belongs to M2. Pinned by `TestLoadProposalByID_ProducerSchemaMismatch`, a
characterization test that MUST fail once M2 lands.

### Known gaps (not addressed in M1)

- `plan.md` / `acceptance.md` remain unwritten and no `plan-auditor` verdict
  exists; the SPEC is Tier L and nominally expects 5 artifacts.
- `harness execute` diagnostics carry a doubled `harness execute:` prefix. This
  predates M1 (present at `6fe9d89e8`) and was left untouched under scope
  discipline.
- Proposal coverage for `internal/cli/harness` measured 80.9%, below the 85%
  package target. Pre-existing; M1 added tests but did not close the gap.

## §F Phase 4 Mode Selection — M2

Recorded before the first M2 run-phase `Agent()` spawn, per the canonical
mode-logging policy (`orchestration-mode-selection.md` §D). M1's mode was not
logged at the time (retroactively noted: M1 was also Mode 5 — shared accessor +
3-site rewire is coding-heavy single-track work).

**Input parameters**
- tier: L
- scope: ~5-8 Go files (1 new: `harness promote` verb; edits: `applier.go`, `execute.go`, `layout_repro_test.go`; new tests for promote + guard)
- domain count: 2-3 (`internal/cli/harness`, `internal/harness`, `internal/cli`)
- file language mix: 100% Go
- concurrency benefit: LOW — coding-heavy (Anthropic coding-task parallelism caveat)

**Mode evaluation**

| Mode | Selected? | Rationale |
|---|---|---|
| 1 trivial | no | multi-file, semantic change (new verb + guard + de-wire + test retirement) |
| 2 background | no | write-capable implementation, not read-only |
| 3 agent-team | no | RETIRED |
| 4 parallel | no | coding-heavy → Mode 5 preferred (Anthropic caveat) |
| 5 sub-agent | **yes** | coding-heavy default; sequential per-deliverable |
| 6 workflow | no | <30 files, semantic (not mechanical-uniform transform) |

**Decision: sub-agent** (Mode 5)

Justification: M2 is coding-heavy implementation (new CLI verb + Apply
pre-flight guard + execute de-wire/diagnostic + tier tripwire retirement).
Per Anthropic's coding-task parallelism caveat, coding-heavy work with
inter-file dependencies belongs to sequential sub-agent delegation, not
parallel fan-out. Tier L → Section A-E delegation template required. The
4 deliverables carry a strict sequencing constraint (§A.7: the applicability
guard MUST land before any payload-parseable change), which independently
reinforces sequential execution.

Implementation Kickoff Approval: obtained — user selected "M2 착수" at the
AskUserQuestion gate this session (2026-07-28), after the three §A.6 open
decisions were resolved (CLI verb promotion path · in-Apply guard · execute
left dormant).
