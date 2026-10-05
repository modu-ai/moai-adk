# SPEC-CODEX-INIT-001 — progress

> Run-phase record. `§F Phase 4 Mode Selection` is the orchestrator's log —
> not authored here. `§E.4` belongs to manager-docs at the sync commit.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-28
plan_audit: round 5 PASS 1.00 (.moai/reports/t340/verdict-iter5.md)

## §E.2 Run-phase Evidence

Baseline: worktree t340, branch `WT-codex-init`, plan-repair commit `6a98ef743`
(base `f5a834fef` = origin/develop). RED-now observations R1-R4 were pinned at
`f5a834fef` (acceptance.md two-cell table) and independently re-executed by
the round-5 auditor on the same tree.

### Milestones

| M | Scope | Cells | Status | Evidence |
|---|---|---|---|---|
| M1 | Proposal gate + state×verb×spawn matrix + injected-state judgement | 24 + 32 | GREEN | `.moai/state/verify/t340/m1-m3-gate-cells.txt` (148 PASS subtests incl. M2/M3) |
| M2 | Generator delegation, decline, non-interactive prompt counting | 20 + 20 + 40 | GREEN | same run, per-test `-v` subtests |
| M3 | Failure paths E1/E2/E3 × verb × spawn | 12 | GREEN | same run |
| M4 | Path containment — guard is the contract's FIRST act (M4 before M5, plan §D) | 84 + 60 + 36 | GREEN | `.moai/state/verify/t340/m4-containment-second-run.txt` (180 PASS, skip list printed: "(none)" on darwin) |
| M5 | Instruction contract + idempotency | 10 + 12 | GREEN | `.moai/state/verify/t340/m5-m6-link-cells-second-run.txt` (26 PASS incl. M6) |
| M6 | Local-file reachability + overlays | 4 | GREEN | same run |

### M1-M3 observed-red record (E8)

- Plan-phase RED-now: acceptance R1-R4 at `f5a834fef` — implementation units
  absent (`ls internal/cli/codex_init.go …` rc 1), launch path direct to
  `codexSpawnLaunch`/`codexDirectLaunch` with no gate, repo itself unwired.
  Auditor round 5 re-executed these on the same tree (verdict §4).
- In-run observed reds (mutant-style, verbatim in the session transcript):
  (a) accept cells red on a chmod-after-close defect — `AGENTS.md missing
  after acceptance` across all 20 cells; (b) E1-E3 cells red on a missing
  interactive flag — gate exited at the non-interactive branch and the
  generator was never reached. Both observed red before the fix, green after.

### AC matrix so far (GREEN per cell = `-v` subtest)

| AC | Cells | Status | Command | Actual output |
|---|---|---|---|---|
| AC-CI-001 | 24 | PASS | `go test ./internal/cli/ -run TestCodexInit -v` | `TestCodexInitGateStateMatrix` 24/24 subtests |
| AC-CI-002 | 32 | PASS | same | `TestCodexInitGateInjectedState` 32/32 |
| AC-CI-003 | 20 | PASS | same | `TestCodexInitDecline` 20/20 |
| AC-CI-004 (accept) | 20 | PASS | same | `TestCodexInitAcceptDelegation` 20/20 |
| AC-CI-004 (prompt) | 40 | PASS | same | `TestCodexInitPromptIssuance` 40/40 |
| AC-CI-010 | 12 | PASS | same | `TestCodexInitFailurePaths` 12/12 |
| AC-CI-011 | 180 | PASS | `go test ./internal/cli/ -run TestCodexPathGuard -v` | 180/180; darwin runs all real fixtures; skip list "(none)" |
| AC-CI-005/006 | 22 | PASS | `go test ./internal/cli/ -run 'TestCodexContractLinkCreation\|TestCodexContractIdempotent' -v` | 22/22 |
| AC-CI-007 | 4 | PASS | `... -run TestCodexLocalReachability -v` | 4/4 |
| AC-CI-008 | overlay | PASS | embedded in every cell above (SNAP outside-set, write-seam paths, process-seam count) | no isolated cells by design |

### Existing-suite protection

`go test ./internal/cli/ -run TestCodex` rc 0 after the gate insertion —
launcher-SPEC tests unbroken; the shared launch-capture helper pins the gate
open (`withCodexGateOpen`) and four manual-seam launcher tests call it.
Evidence: `.moai/state/verify/t340/existing-codex-tests-after-gate.txt`.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-08-28
run_commit_sha: 0a518f1d8
run_status: complete
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch:
l44_post_push_fetch:
new_warnings_or_lints_introduced:
cross_platform_build.darwin_arm64: build+vet pass (in-run)
cross_platform_build.windows: GOOS=windows vet rc=0 (2026-08-28, lead takeover after agent stall)
total_run_phase_files: 5 (codex_init.go, codex_contract.go, codex_init_test.go,
codex_contract_test.go pending, codex_launcher.go +1 call site)
m1_to_mN_commit_strategy: per-milestone commits (M1-M3 gate surface, M4+M5
contract file, M6 reachability)

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-08-28
sync_commit_sha: 4b36b00d6
sync_status: complete
b12_self_test_a: pre-emission grep `grep -c 'SPEC-CODEX-INIT-001' CHANGELOG.md`
→ 0 (clear to emit; no parallel BATCH-SYNC entry)
b12_self_test_b: distinct AC ids in acceptance.md → 13 raw tokens (AC-CI-001..011
= 11 own; AC-CL-004/007 are launcher-SPEC cross-references cited in the shared
definitions). CHANGELOG entry cites 11 (AC-CI-001..011), matching §E.3
ac_pass_count: 11 — non-zero on both sides
b12_self_test_c: every file path claimed in the CHANGELOG entry verified to
exist (`.moai/specs/SPEC-CODEX-INIT-001/spec.md`, `internal/cli/codex_launcher.go`,
`internal/cli/codex_contract.go`, `internal/codexwiring/`) — all present
changelog_entry_position: `[Unreleased]` → `### Added`, first (top) entry
frontmatter_status_transitions.spec_md: in-progress → completed (3-phase close
merged into the single sync commit). plan.md / acceptance.md / progress.md carry
no YAML frontmatter in this SPEC — no further transitions apply
mx_tag_validation: performed as a sync sub-step by direct read — @MX:NOTE on the
gate's single-judgement source (`internal/cli/codex_init.go`) and @MX:ANCHOR +
@MX:REASON on the contract's ordering invariant
(`internal/cli/codex_contract.go`, secureCodexInstructionContract) are present
and well-formed; no missing tags found, none added
codemaps_regeneration: 6 documents (overview / modules / data-flow / dependencies
/ entry-points / docs-truth) + provenance.json updated to describe the `moai
codex` launcher surface (card t197 gap) and this SPEC's init-offer gate surface
(codex_init / codex_contract / single call site in codex_launcher); measured
counts refreshed to this tree (2026-08-28). provenance commit_sha pinned to
`f1d80b305` — the described tree; the sync commit changes no Go file under
internal/, cmd/, or pkg/
