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
- `moai harness ledger record` was exercised once this session and works (exit 0, creates a
  **pending** row — not a ledger line; the ledger line appears only at terminal-evidence Stop).
  The recording **obligation already exists** in doctrine (`.claude/skills/moai/SKILL.md` router
  section + `workflows/run.md:196`, shipped by `SPEC-HARNESS-EVOLVE-001` M3 commit `1c54cd9c6`,
  2026-07-12). The ledger stayed empty because no dispatch was driven to a terminal-evidence Stop
  and the LLM-obeyed doctrine is not mechanically enforced — M3 verifies the mechanics + executes
  the AC-HLR-007 falsification; it does not add a second obligation.

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

## §E.2 (cont.) Run-phase Evidence — M2 (route drafts to designed consumer)

M2 = REQ-HLR-004/004b/004c/012, AC-004/005/014/015/016. Implemented in the prior session (commits `1050d6738`·`127bd7bbe`·`1a1d6e206`·`1f78f0da3`·`93dc4b5dd`); evidence backfilled 2026-07-28 (the prior session recorded the work but not this §E.2 block).

### AC verification (backfill, observed 2026-07-28, HEAD `300847a64`)

| AC | Status | Test | Observed |
|---|---|---|---|
| AC-004 | PASS | `TestPromote_CreatesSPECSkeletonWithProvenance` + `TestPromote_ProvenanceRoundTripsDraftID` + `TestPromote_DraftLeavesPendingQueue` | all `--- PASS` |
| AC-005 | PASS | `TestPromote_AppendsExactlyOneAuditRecord` | `--- PASS` |
| AC-014 | PASS | `TestLoadProposalByID_DiscoveryDraftDiagnostic` | `--- PASS` |
| AC-015 | PASS | `TestApply_ApplicabilityGuard_RejectsContentFreeProposal` | `--- PASS` (subtests: all-three-empty, only-target-path-set) |
| AC-016 | PASS | `TestLoadProposalByID_ProducerSchemaMismatch_RetiredByM2` (sentinel) + no `Tier` JSON codec | sentinel `--- PASS`; `grep` confirms no codec (clause 1) |

Design decisions (resolved at the M2 Implementation Kickoff gate, recorded in `plan.md` §A.6): promotion path = `moai harness promote --id`; applicability guard = in-`Apply` pre-flight (before `createSnapshot`); `execute` left dormant (diagnostic only).

The prior M2 session executed the falsifications (revert/observe/restore) per its handoff; this backfill re-ran the tests (all green) and recorded the evidence. See `acceptance.md §D` "M2 — verification evidence (backfilled)".

## §E.2 (cont.) Run-phase Evidence — M3 (dispatch observation)

M3 = REQ-HLR-005 / AC-HLR-007. **The recording obligation and CLI mechanics were already
shipped by `SPEC-HARNESS-EVOLVE-001` M3 (`1c54cd9c6`, 2026-07-12).** M3's deliverable under
the user-approved Option A is: verify the mechanics end-to-end + execute the AC-HLR-007
falsification + correct the stale "obligation missing" premise in plan.md §F.3 and the
progress.md ledger-record line above. No Go code changed.

### AC verification

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-HLR-007 | PASS | full lifecycle in isolated `/tmp` root (worktree binary) | record→pending=1; evidence --terminal; `harness-observe-stop`→ledger 0→**1**, `outcome:"success"`; falsification (no record)→ledger unchanged; gate0 OFF→finalize skipped |

### Falsification executed (both directions, observed not assumed)

Driven in a fresh `/tmp` root with `hook.opt_in.enabled: true` + `learning.enabled` default-ON:

1. `record --session m3a` (stdin `/moai run SPEC-X`) → `routing-pending-m3a.json` created, ledger 0.
2. `evidence --session m3a --kind gate_exit --value 0 --terminal` → terminal evidence on pending.
3. `harness-observe-stop` (stdin `{"session_id":"m3a",...}`) → `FinalizeOnStop` derived terminal
   `success`, appended **1 line** to `routing-ledger.jsonl`, deleted the pending file.
4. **Falsification**: `harness-observe-stop` for a session with NO `record` → ledger unchanged
   (self-gated no-op, `pending.go:148-150`). Confirmed.
5. **Gate cross-check**: `hook.opt_in.enabled: false` → record+evidence+Stop leaves ledger
   unchanged (`finalizeRoutingLedgerOnStop` gate 0, `hook.go:737-738`). Confirmed.
6. **Privacy**: ledger row carries `request_digest: sha256:e0e7dba9545a`, verbatim request
   `grep -c` = 0. Confirmed (no leak).

Verbatim evidence log: `/tmp/hlr-m3-R3-3UJ1/.moai/state/routing-ledger.jsonl` (1 row).

### Lifecycle clarification

The AC's "routing-ledger line count +1 per dispatch" is realized across the dispatch→finalize
lifecycle: `record` writes a *pending* row at dispatch; a terminal-evidence Stop finalizes it
into `routing-ledger.jsonl`. A single `record` call never appends to the ledger. This is the
designed `pending.go` architecture, not a defect; `acceptance.md §E.1` carries the corrected
Baseline wording.

### Known gaps / residual risk (not closed by M3)

- **Obedience gap (residual risk).** The obligation is LLM-obeyed doctrine; no mechanical
  backstop fires `record` at dispatch. The primary-checkout ledger holds 1 row (audit-hand-written)
  despite the doctrine shipping 2026-07-12. Closing this needs a mechanical dispatch-time
  trigger (Option B follow-up), out of scope for M3 per user decision.
- No new Go test added — the existing `TestLedgerVerbs_RecordEvidenceList`, `TestLedgerRecord_LearningDisabledNoOp`,
  and `TestHarnessObserveStop_RoutingLedgerGated` already cover the record/evidence/finalize/gate
  mechanics; M3's contribution is the live end-to-end execution + falsification evidence above.

## §E.2 (cont.) Run-phase Evidence — M4 (generator quality)

M4 = REQ-HLR-009/011, AC-008/017. **§G q2 resolved (AskUserQuestion, 2026-07-28): option (a) exclude `agent_invocation` event type.** Implemented by manager-develop (TDD) + orchestrator-independent-verified.

### Change set

| File | Change |
|---|---|
| `internal/harness/proposalgen/mapper.go` | C1 `excludedEventTypes` (SSOT-derived from `harness.EventTypeAgentInvocation`) + `hasExcludedEventType` + `isActionable` outermost narrowing clause; C2 `buildDraftID` drops `<YYYYMMDD>` |
| `internal/harness/proposalgen/mapper_test.go` | RED tests `TestMapper_AgentInvocationExcludedRegardlessOfTier` (AC-008), `TestMapper_DraftIDStableAcrossDates` (AC-017); format-gate preservation test; reconciled `TestMapper_RealDataSchemaPass`/`TestMapper_DraftIDFormat` |
| `internal/cli/hook_harness_propose_chain_test.go` | cascade fix — seed `agent_invocation` → `tool_failure` (commit `efcb4990c`) |

### AC verification (orchestrator-independent)

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-HLR-008 | PASS | `go test -run TestMapper_AgentInvocationExcludedRegardlessOfTier ./internal/harness/proposalgen/...` | ok; `{agent_invocation:Bash:, auto_update, 0.9, obs:5}` → 0 candidates |
| AC-HLR-017 | PASS | `go test -run TestMapper_DraftIDStableAcrossDates ./internal/harness/proposalgen/...` | ok; same `pattern_key` on 2 dates → identical `PROPOSAL-dc05149f` |

### Falsification executed (manager-develop, observed not assumed)

- C1 revert (drop `hasExcludedEventType` early-return) → AC-008 test FAIL (`got 1 candidate, want 0`, draft `PROPOSAL-20260728-aa401a3a`) → restored → green.
- C2 revert (restore `<YYYYMMDD>`) → AC-017 test FAIL (`day1=...20260713-dc05149f` ≠ `day2=...20260714-dc05149f`) → restored → green.

### Orchestrator independent verification (verification-claim-integrity)

- **Full `go test ./...` exit 0 (105 ok, 0 FAIL)** — confirms no cascade beyond the one manager-develop reported.
- `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 · `golangci-lint run` 0 issues.
- mapper.go function coverage 100% (package 78.1% — pre-existing baseline from scaffolder/reader/candidate, out of scope).
- Subagent boundary (B3): the 1 grep match at `scaffolder.go:162` is a pre-existing string literal in a draft template, NOT a tool invocation, and NOT in the M4 diff. 0 new invocations.

### Commits

- `b010bcfd9` — M4 mapper.go C1+C2 + tests (manager-develop).
- `efcb4990c` — RATCHET-REWIRE propose-chain test seed cascade fix (orchestrator).
- Both **not pushed** — orchestrator pushes after this verification.

## §E.2 (cont.) Run-phase Evidence — M5 (lesson channel)

M5 = REQ-HLR-006/007, AC-HLR-009/010. Implemented by orchestrator-direct execution (user-approved; M5 is doctrine/memory work, not Go code implementation). §G q3 was resolved in the prior session (commit `8eb11902a`, option a — topic-file convention as single lesson store).

### Change set

- `moai-constitution.md` § Lessons Protocol: line 145 (`Store lessons at ... lessons.md` → topic-file convention `feedback_*.md` + `MEMORY.md` as the single designated store); line 149 (archive target `lessons-archive.md` → `memory/_archive/`); line 153 (drain actor + drain trigger named — the orchestrator + the "backlog obscures recurring patterns" trigger). Edited in BOTH `.claude/rules/moai/core/moai-constitution.md` (local) and `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` (mirror).
- `make build` recompiled the binary (`bin/moai` v3.0.1, `8eb11902a`; build smoke `./bin/moai version` → exit 0).
- `lessons.md` (auto-memory, 47 KB, 41 days stale) marked `[SUPERSEDED]` — content not migrated, kept on disk for the audit trail.
- `.moai/lessons-inbox.jsonl` drain: 870 → 149 lines (82.9% reduction). Noise drained = `UnknownFailure(Bash/Read)` 721 one-off dev failures; preserved = VALUE event_keys (OOM/Sandbox/Permission/Timeout/ContextCancelled/Agent/MCP) 149 lines. Backup at `.moai/state/lessons-inbox.backup-20260728.jsonl` (870 lines, sha256 `666b46e5...`).

### AC verification (orchestrator-direct, observed 2026-07-28)

- **AC-HLR-009 (one designated lesson store): PASS.** Constitution § Lessons Protocol now names the topic-file convention (`feedback_*.md` + `MEMORY.md` index) as the single designated store; no rule names the stale `lessons.md` as the store (it is marked `[SUPERSEDED]`). Falsification direction — reverting the constitution edit restores `lessons.md` (mtime 2026-06-17, 41 days stale) as the designated store whose mtime is older than the practiced `feedback_*.md` topic files; the edit is the load-bearing change.
- **AC-HLR-010 (inbox drain named): PASS.** Drain actor (orchestrator) + drain trigger (backlog obscures recurring patterns) both named in constitution line 153. A drain run reduced the undrained entry count: 870 → 149 (observed via `wc -l` before/after; sha256-verified replacement `4841ecc6...`). Falsification direction (count unchanged) is refuted by the measured reduction.

### Orchestrator independent verification (verification-claim-integrity)

- `go test ./...` → exit 0, 0 FAIL (full suite).
- `golangci-lint run --timeout=3m` → exit 0, 0 issues.
- byte-parity local↔template constitution: `diff` → identical.
- template neutrality (§25): `grep` for SPEC-ID / REQ-HLR / AC-HLR / commit-SHA / internal-date in both constitution files → 0 matches.
- `./bin/moai version` → exit 0, `v3.0.1 / 8eb11902a`.

### Known gaps / residual risk (not closed by M5)

- The drain mechanism is doctrine-only (no Go backstop) — parallel to the M3 routing-ledger finding. The orchestrator executes drains per the constitution-named actor/trigger; no mechanical drain hook enforces frequency or pattern conversion. Accepted doctrine-level debt, not closed by M5.
- The drain converted ONE representative feedback topic file (`feedback_lessons_inbox_drain_2026_07.md`, carrying `prediction:` + `verified: false`) — the minimum to exercise M6 AC-011's prediction/verified obligation on this SPEC's own lesson entry. Full conversion of all 149 preserved VALUE stubs into individual feedback files is NOT done (out of scope for M5; the 149 are preserved in the inbox for future drain cycles).

## §E.2 (cont.) Run-phase Evidence — M6 (falsifiability + CLI reporting)

M6 = REQ-HLR-008/010, AC-011/012/013. Implemented by manager-develop (TDD) — the final milestone. §F.5 was the lowest-reversibility milestone (mechanical CLI reporting fixes + doctrine evidence).

### Change set

| File | Change |
|---|---|
| `internal/cli/harness_route.go` | AC-012: replaced the static hand-authored Long string (14/20 verbs) with `buildHarnessRouterLong(cmd)` — derives the Long from `cmd.Commands()` AFTER every AddCommand call, so AddCommand is the single SSOT and the verb enumeration is a round trip. Added `useFirstToken` helper + `sort`/`strings` imports. |
| `internal/cli/harness_route_test.go` | AC-012 RED→GREEN test `TestHarnessRouterHelp_EnumeratesAllVerbs` + `verbDocumentedInLong` helper (first-token check avoids "list" ⊂ "mute-list" false positives). |
| `internal/cli/harness/v4lifecycle_cmd.go` | AC-013: list domain cell for a thin harness changed from `"(manifest missing)"` (defect-suggesting) to `"(command-only thin harness)"` (mirrors doctor's INFO framing). |
| `internal/cli/harness/list_doctor_agreement_test.go` | NEW — AC-013 RED→GREEN test `TestHarnessListAndDoctor_AgreeOnCommandOnlyThinHarness`. |
| `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_harness_loop_repair_m6_2026_07.md` | NEW — AC-011 lesson entry for the M6 harness edits (prediction/verified/surface). |
| `~/.claude/projects/.../memory/feedback_lessons_inbox_drain_2026_07.md` | AC-011: annotated the M5 entry with the observed drain evidence (870→149 mechanism confirmed); `verified:` remains `false` honestly (long-term velocity/conversion effects pending). |
| `~/.claude/projects/.../memory/MEMORY.md` | Index entry for the M6 lesson file. |
| `.moai/specs/SPEC-HARNESS-LOOP-REPAIR-001/acceptance.md` | §B status table: AC-011/012/013 → PASS (17/17). §H.1/§H.2/§H.3 enriched with 5-section PASS evidence. |

### AC verification (orchestrator-independent, observed 2026-07-28)

| AC | Status | Command | Observed |
|---|---|---|---|
| AC-HLR-011 | PASS | `grep -c 'prediction:\|verified:' .../feedback_harness_loop_repair_m6_2026_07.md` | `3`; M5 entry `2`; both entries carry both fields |
| AC-HLR-012 | PASS | `go test -run TestHarnessRouterHelp_EnumeratesAllVerbs ./internal/cli/` | `ok` — all 20 registered verbs documented in Long (was 14/20) |
| AC-HLR-013 | PASS | `go test -run TestHarnessListAndDoctor_AgreeOnCommandOnlyThinHarness ./internal/cli/harness/` | `ok` — list frames thin harness as "command-only"; doctor INFO; no "missing" word |

### Falsification executed (both directions, observed not assumed)

- **AC-012**: temporarily added `if use == "ledger" { continue }` in `buildHarnessRouterLong` (ledger registered but skipped in description) → test FAIL (`verb "ledger" is registered (AddCommand) but NOT documented`) → reverted → green.
- **AC-013**: reverted list domain cell to `"(manifest missing)"` → test FAIL (`list describes the thin harness in defect-suggesting terms (contains "missing")`) → reverted → green.

### Orchestrator independent verification (verification-claim-integrity)

- **Full `go test ./...` exit 0 (all packages)** — no cascade from the help-text or list/doctor changes.
- `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 · `GOOS=linux GOARCH=amd64 go build ./...` exit 0.
- `golangci-lint run --timeout=2m` → exit 0, 0 issues.
- Subagent boundary (B3 / §J item 4): `grep -rn 'AskUserQuestion(\|mcp__askuser(' internal/cli/harness*.go internal/cli/harness/*.go` (function-call pattern, excluding tests) → 0 matches. The broader-text matches are all permitted prose comments / flag descriptions (the M6-edited files contain only `//` comments referencing the `TestPropose_NoAskUserQuestion` static guard).
- Coverage: `internal/cli/harness` 80.6% (pre-existing baseline from M1: 80.9%; M6 added fully-covered new code, the delta is the new test file + one-line edit). `internal/cli` 75.5% (large package, pre-existing). New function coverage: `buildHarnessRouterLong` + `useFirstToken` exercised 100% by `TestHarnessRouterHelp_EnumeratesAllVerbs`.

### Known gaps / residual risk (not closed by M6)

- Package coverage for `internal/cli/harness` remains below the 85% target (80.6%, pre-existing from M1). M6 added tests but did not close the gap (the gap is in scaffolder/reader/candidate, out of M6 scope).
- AC-011's M5 entry remains `verified: false` — the prediction's long-term effects (inbox-noise velocity reduction + VALUE-pattern → feedback conversion) require future drain cycles. M6 does not constitute that observation.

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
