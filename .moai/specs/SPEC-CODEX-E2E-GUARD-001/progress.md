# Progress — SPEC-CODEX-E2E-GUARD-001

## §E.1 Plan-phase Audit-Ready Signal

plan_phase:
  spec: SPEC-CODEX-E2E-GUARD-001
  base_sha: "ace1c5440"
  branch: WT-codex-e2e-guard
  worktree: .claude/worktrees/t500
  tier: M
  measurements: spec.md §A + §F (commands and outputs in spec.md HISTORY)
  premise_refutations:
    - axis-2 card premise REFUTED (spec.md §F.1) — existing TestStatusLineDefaultSubsetOfAllowlist
    - lane "41 matches no population" PARTIALLY SUPERSEDED — 41 is the repo-wide codex-named
      test-file population at this base (spec.md §F.2)
  baseline_greens:
    - "go test ./internal/codexwiring/ -run TestStatusLine -count=1 → ok 0.617s"
    - "go test ./internal/cli/ -run 'TestCodexSpecFiles|TestCodexCommand_NeutralityScan|TestCodexSpawn_TmuxDiagnosticSingleSource' -count=1 -timeout 600s → ok 0.929s"
  lint: pending run-phase (moai spec lint at sync)

## §E.2 Run-phase Evidence

All run-phase measurements taken this run in worktree `.claude/worktrees/t500` at HEAD
`a3153c215` (branch WT-codex-e2e-guard), before the single run-phase commit. Full verbatim
records: `.moai/state/verify/t500/run-evidence-20260907.md`.

### E.2.0 Pre-flight (M0)

- `go build ./...` → rc=0; `GOOS=windows GOARCH=amd64 go build ./...` → rc=0.
- `golangci-lint run --timeout=2m ./internal/cli/... ./internal/codexwiring/...` → `0 issues.` rc=0 (baseline).
- `go test ./internal/codexwiring/ -run TestStatusLine -count=1` → `ok … 0.567s`.
- `go test ./internal/cli/ -run 'TestCodexSpecFiles|TestCodexCommand_NeutralityScan|TestCodexSpawn_TmuxDiagnosticSingleSource' -count=1 -timeout 600s` → `ok … 0.882s`.
- Exec call sites re-read (plan §G anti-pattern): codex_launcher.go:457 (`exec.Command(req.Program, …)`), mcp_codex.go:350/433/1889 (`exec.CommandContext(ctx, binaryPath, …)` ×3), codex_review_gate.go:129 (`exec.Command("git", …)`).

### E.2.1 AC matrix (verbatim outputs at a3153c215)

| AC | Status | Command | Observed output (verbatim, trimmed to the verdict lines) |
|----|--------|---------|----------------------------------------------------------|
| AC-CEG-001 | PASS | `go test ./internal/cli/ -run TestRunInit_ThenDoctorCodexWiringHealthy -count=1 -timeout 600s -v` | `--- PASS: TestRunInit_ThenDoctorCodexWiringHealthy (0.43s)` / `ok github.com/modu-ai/moai-adk/internal/cli 1.128s` |
| AC-CEG-002 | PASS | `go test ./internal/cli/ -run 'TestRunInit_ThenDoctorCodexWiringHealthy\|TestRunInit_ClaudeOnlyThenDoctorStaysSilent' -count=1 -timeout 600s` | `ok github.com/modu-ai/moai-adk/internal/cli 1.505s` (companion lands — no descope needed) |
| AC-CEG-003 | PASS | `go test ./internal/codexwiring/ -run TestStatusLineDefaultSubsetOfAllowlist -count=1` (mutant / revert) | RED: `--- FAIL: … statusline_test.go:57: default token "git-branchx" is not in statusLineAllowlist — Codex would drop or reject it` / FAIL rc=1 → revert GREEN: `ok … 0.394s` → `git status --porcelain internal/codexwiring/configtoml.go` EMPTY; `git log --oneline -- internal/codexwiring/configtoml.go` head = `7b217da7c` (pre-existing; no mutant commit) |
| AC-CEG-004 | (sync-owned) | `grep -n "^## §F\|^### §F" .moai/specs/SPEC-CODEX-E2E-GUARD-001/spec.md` | §F.1/§F.2 present since plan phase; the sync-phase verdict flips this AC — not judged at run phase |
| AC-CEG-005 | PASS | `go test ./internal/cli/ -run TestCodexSpecFiles_NoBuildTagsOrSyscall -count=1 -timeout 600s` over the 12-file set | `--- PASS …` in the verbose batch below; swept-count mutant: `codex_launcher_guards_test.go:179: guard swept 11 files, want 12 — vacuous green` FAIL rc=1 (one entry transiently removed, restored, never staged); independent count `ls internal/cli/*codex*.go \| grep -v _test \| wc -l` → `12` |
| AC-CEG-006 | PASS | `go test ./internal/cli/ -run TestCodexSpecFiles_ExecPrimitivesCodexOnly -count=1 -timeout 600s` | `--- PASS: TestCodexSpecFiles_ExecPrimitivesCodexOnly (0.00s)`; per-file table req.Program / binaryPath / `"git"` / zero-call ×9; comment-line skip preserved |
| AC-CEG-007 | PASS | `go test ./internal/cli/ -run TestCodexCommand_GuardFileLiteralsNeutral -count=1 -timeout 600s` | `--- PASS: TestCodexCommand_GuardFileLiteralsNeutral (0.01s)`; baseline measured FIRST (violations observed, then narrowed — E.2.2); canary `"git"` @ codex_review_gate.go observed; 0-literal sweep is RED |

Verbose guard batch (all five guard tests GREEN on the 12-file form):

```
=== RUN   TestCodexCommand_NeutralityScan
--- PASS: TestCodexCommand_NeutralityScan (0.00s)
=== RUN   TestCodexSpecFiles_NoBuildTagsOrSyscall
--- PASS: TestCodexSpecFiles_NoBuildTagsOrSyscall (0.00s)
=== RUN   TestCodexSpecFiles_ExecPrimitivesCodexOnly
--- PASS: TestCodexSpecFiles_ExecPrimitivesCodexOnly (0.00s)
=== RUN   TestCodexSpawn_TmuxDiagnosticSingleSource
--- PASS: TestCodexSpawn_TmuxDiagnosticSingleSource (0.01s)
=== RUN   TestCodexCommand_GuardFileLiteralsNeutral
--- PASS: TestCodexCommand_GuardFileLiteralsNeutral (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.964s
```

### E.2.2 M4 literal-scan narrowing (REQ-CEG-009 justification record)

Baseline measured BEFORE landing assertions (plan M4.2): the first widened-scan run OBSERVED —
`CLAUDE.local` ×2 (`codex_contract.go:31/37` — the contract check names the dev-only files it
inspects), `.moai/reports` ×1 (`codex_review_gate.go:38` — the constant IS the
runtime-managed-path filter), and non-ASCII ×many (deliberate bilingual runtime copy: Korean
progress messages in codex_task.go / mcp_codex.go, em-dash typography). Narrowing applied to the
LITERAL scan only: kept the seven leakage classes (SPEC- id, REQ- id, card id, ISO date, 7+ hex,
/Users/, /home/ — all clean on baseline); dropped `CLAUDE.local`, `.moai/reports`, and the
non-ASCII count, with the justification in the guard-file comment
(`codexLiteralNeutralityPatterns`). `TestCodexCommand_NeutralityScan` (AC-CL-013) keeps the FULL
table — untouched, never weakened. (plan.md M4 named "this plan section" as a record carrier;
plan.md body is not run-phase-editable per ownership, so the justification is recorded here and
in the guard-file comment.)

Exec-table mechanics note: `exec.CommandContext` takes `ctx` as its first argument and the
binary second, so the exec guard captures the binary argument per primitive shape
(CommandContext → 2nd arg; Command/StartProcess → 1st arg). The original 2-file regex captured
the literal first argument and sufficed only because codex_launcher.go uses `exec.Command`.

### E.2.3 Scoped verification batch (M5)

```
$ go build ./...                                                     → rc=0
$ GOOS=windows GOARCH=amd64 go build ./...                           → rc=0
$ GOOS=windows GOARCH=amd64 go vet ./internal/cli/ ./internal/codexwiring/  → rc=0
$ golangci-lint run --timeout=2m ./internal/cli/... ./internal/codexwiring/... → "0 issues." rc=0
$ go test ./internal/cli/ -count=1 -timeout 600s -cover
ok  	github.com/modu-ai/moai-adk/internal/cli	356.855s	coverage: 80.7% of statements
$ go test ./internal/codexwiring/ -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.605s	coverage: 88.2% of statements
```

Lint NEW-vs-baseline: baseline `0 issues.` (pre-change, same command) — NEW issues introduced: 0.
Coverage: pre-change baselines NOT separately measured (Gap, recorded honestly); production
code is untouched, so statement denominators are identical and only test-side coverage was added.

### E.2.4 Files touched

- `internal/cli/doctor_codex_e2e_test.go` — NEW (M2, 110 lines).
- `internal/cli/codex_launcher_guards_test.go` — EXTENDED (M3/M4; 2-file set → 12-file length-pinned set, per-file exec table, literal scan + canary).
- `.moai/specs/SPEC-CODEX-E2E-GUARD-001/progress.md` — this record.
- `internal/codexwiring/configtoml.go` — touched TRANSIENTLY for the M1 mutant, byte-restored, porcelain-empty gate passed; never staged, never committed.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-09-07"
run_commit_sha: "667509ac9"   # single run-phase commit; backfilled after landing
run_status: "complete"
ac_pass_count: 6        # AC-CEG-001/002/003/005/006/007
ac_fail_count: 0        # AC-CEG-004 is sync-owned (spec.md §F verdict), not a run-phase fail
preserve_list_post_run_count: 0   # zero production files modified (test-only card)
l44_pre_commit_fetch: "n/a — lane worktree; lane never pushes (gitflow lane protocol §4)"
l44_post_push_fetch: "n/a — push is the lead's batch act; lane reports merge SHA only"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  native: "ok (go build ./... rc=0)"
  windows: "ok (GOOS=windows GOARCH=amd64 go build ./... rc=0)"
  windows_vet_test_typecheck: "ok (GOOS=windows go vet ./internal/cli/ ./internal/codexwiring/ rc=0)"
total_run_phase_files: 4
m1_to_mN_commit_strategy: "single run-phase commit carrying M1-M5 (evidence-first; transient mutants never staged)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-09-07"
sync_commit_sha: "2d98654af"   # resolved from the pending-backfill-sync placeholder (a commit cannot cite its own SHA)
sync_status: "complete"
b12_self_test_a_duplicate_gate: "grep -c 'SPEC-CODEX-E2E-GUARD-001' CHANGELOG.md → 0 (pre-emission), entry emitted once"
b12_self_test_b_ac_count: "acceptance.md live AC identifiers: 8 (AC-CEG-001..007 + AC-CL-007 cross-ref to SPEC-CODEX-LAUNCHER-001); CHANGELOG entry states 7/7 AC-CEG matrix with AC-CL-007 cross-referenced, no double-claim"
b12_self_test_c_file_paths: "internal/cli/doctor_codex_e2e_test.go, internal/cli/codex_launcher_guards_test.go — both verified present via ls before emission"
changelog_entry_position: "first bullet under [Unreleased] → Added"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (full transition on the single sync commit)"
  plan_md: "n/a — no status field (artifact statelessness)"
  acceptance_md: "n/a — no status field"
  progress_md: "n/a — no status field"
updated_field_refresh: "spec.md updated: 2026-09-07 (unchanged date, refreshed on the sync commit)"
ac_ceg_004_verdict: "PASS — sync-owned AC; AC matrix 7/7"
docs_site_readme_disposition: "no edits — test-only internal surface; no user-facing behavior changed"
canary_compliance_check:
  template_tree_touched: false
  mx_tags_added_at_run: 0   # test-only helpers, no exported surface — nothing to annotate; confirmed in sync
  mx_sync_contradiction: false
spec_body_edits: "none — spec.md/plan.md/acceptance.md bodies untouched (frontmatter status: + updated: only)"
run_commit_sha: "pending-backfill-run"   # set at run phase; resolved to 667509ac9 in the backfill commit
```


## §F Phase 4 Mode Selection

Input parameters: tier M; scope 2 test files + 1 new test file (~4-6 files); domains 1 (Go tests,
single package pair); language mix 100% Go; concurrency benefit LOW (coding-heavy, sequential
milestones M1→M5 with dependencies); Agent Teams prerequisites not requested.

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | not selected | multi-milestone Tier M delegation, not a trivial edit |
| serial | **selected** | coding-heavy test authoring; one write-capable agent at a time (concurrency safeguard) |
| fanout | not selected | no multi-domain research; milestone order is a dependency chain |
| sweep | not selected | semantic new-code work, not mechanical-uniform transformation |

Decision: serial — single manager-develop spawn carrying M1→M5 in order.

Justification: this is coding-heavy work (Anthropic's coding-task parallelism caveat — sequential
is the safe default), the milestones are strictly ordered (mutant proof before guard extension,
extension before scoped verification), and a single writer eliminates file-write races inside one
worktree. Factory card context: lane-6 is itself the orchestrator; no fan-out value exists.

Kickoff: Implementation Kickoff Approval obtained from the operator via the lead (2026-09-07,
AskUserQuestion — lead relay, not lead-proxy). Plan-audit iter2 PASS 1.00; Phase-1 gate skip
conditions hold at run entry (verdict PASS, 1.00 ≥ 0.80, artifact hash unchanged).
