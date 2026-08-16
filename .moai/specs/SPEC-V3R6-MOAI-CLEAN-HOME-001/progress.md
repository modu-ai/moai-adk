# Progress — SPEC-V3R6-MOAI-CLEAN-HOME-001

## §E.2 Run-phase Evidence

Run phase executed 2026-08-17 (manager-develop, cycle_type=tdd, worktree `.claude/worktrees/release-v311`, branch `release/v3.1.1`). Base HEAD at run start: `09516ff0c`. Milestone commits: M1 `051a2fa94`, M2 `6ec0ad212`, M3 `a42941195`, M4 = the commit carrying this section. All evidence below was captured in THIS run against THIS tree at the listed HEAD.

### AC matrix

| AC | Status | Verification command | Observed output (this run, this tree) |
|----|--------|---------------------|----------------------------------------|
| AC-MCH-001 Home Disk Usage check present | PASS | `go test ./internal/cli/ -run 'TestCheckHomeDisk' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli`; golden `doctor-nocolor.golden` line `ok Home Disk Usage no ~/.moai home found — nothing to report` (deterministic via HOME-pinned captureDoctorCmd) |
| AC-MCH-002 duplicate cluster names both profiles | PASS | `go test ./internal/cli/ -run 'TestFindDuplicateClusters|TestCheckHomeDisk_Duplicate' -count=1` | `ok ...`; detector unit test asserts alpha+beta cluster with byte-equal size + equal entry count; gamma (differing count) excluded |
| AC-MCH-003 WARN above / OK below threshold | PASS | `go test ./internal/cli/ -run 'TestCheckHomeDisk_Warns|TestCheckHomeDisk_OK' -count=1` | `ok ...`; sparse >500MB aged debug fixture → WARN + `moai clean --home` pointer; small fixture → OK |
| AC-MCH-004 dry-run mutates nothing | PASS | `go test ./internal/cli/ -run 'TestCleanHome_DryRun' -count=1` | `ok ...`; snapshotTree (size+mtime per file) identical before/after; output carries `[dry-run]` + `Would delete` |
| AC-MCH-005 force deletes only allowlisted | PASS | `go test ./internal/cli/ -run 'TestCleanHome_Force' -count=1` | `ok ...`; aged debug/logs/backups/2-oldest-releases deleted; fresh counterparts + version.json survive |
| AC-MCH-006 carve-out matrix survives --force | PASS | `go test ./internal/cli/ -run 'TestCleanHomeCarveOut' -count=1` | `ok ...`; 34-case predicate table + full root/per-profile matrix + credentials-inside-aged-removed-* (R1 whole-dir skip) all survive |
| AC-MCH-007 releases keep current+N | PASS | `go test ./internal/cli/ -run 'TestCleanHomeCarveOut_Releases' -count=1` | `ok ...`; current + rc + 3 newest survive, 2 oldest deleted |
| AC-MCH-008 home-tier retention (3 sub-tests) | PASS | `go test ./internal/cli/ -run 'TestCleanHome_RetentionFromHomeTier' -count=1` | `ok ...`; explicit 10d honored / absent → 30d default / explicit 0 disables |
| AC-MCH-009 project-scope clean unchanged | PASS | `go test ./internal/cli/ -run 'TestClean_ProjectScopeRegression' -count=1` | `ok ...`; SPEC-V3R2-RT-004 contract: dry-run default, --force deletes aged runs/, fresh survives, project state.yaml retention honored |
| AC-MCH-010 builds + package tests | PASS (gap disclosed below) | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...`; `go test ./internal/cli/ -count=1 -skip 'TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey'`; `go test ./internal/config/ -count=1` | darwin `DARWIN_EXIT_0`, windows `WINDOWS_EXIT_0`; scoped cli `ok ... 287.667s`; config `ok ... 6.357s` |
| AC-MCH-011 no inline threshold literals | PASS | `grep -rn '500\s*\*\|524288800' internal/cli/doctor_disk.go internal/cli/clean_home.go` | exit 1 — zero matches; the only literals live in `internal/config/defaults.go` (DefaultHomeDiskWarnBytes) |
| AC-MCH-012 MOAI_HOME redirect | PASS | `go test ./internal/cli/ -run 'TestCleanHomeCarveOut_MOAIHomeRedirect' -count=1` | `ok ...`; override-root aged deleted / fresh survives; default-HOME fixture untouched |

### E-item notes (attribution triples: command / verbatim output / baseline)

- **E2 cross-platform**: `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` → `DARWIN_EXIT_0` + `WINDOWS_EXIT_0` (this run, HEAD a42941195).
- **E3 coverage**: `go test ./internal/cli/ -run 'TestCheckHomeDisk|TestCleanHome|TestClean_ProjectScopeRegression|TestFindDuplicateClusters' -coverprofile -count=1` → package 9.9% (diluted by the untested package bulk); per-file weighted across the two NEW files: **87.6% (276/315 statements)** — `go tool cover -func` per-function: clean_home.go 75.0–92.1% (relFromHomeRoot lowest), doctor_disk.go 50.0–100.0% (homeEntrySize error branches lowest).
- **E4 subagent boundary**: `grep -rn 'AskUserQuestion' internal/cli/doctor_disk.go internal/cli/clean_home.go` → exit 1 (zero matches).
- **E5 lint**: `golangci-lint run internal/cli/... internal/config/... --timeout=2m` → `0 issues.` — zero NEW vs the pre-run baseline (pre-flight baseline was also clean).
- **E6 commits/push**: M1 `051a2fa94` / M2 `6ec0ad212` / M3 `a42941195` / M4 (this commit) + SHA-backfill follow-up. NO push — B9 repo-local override: the lane is pushed by another session. l44 pre-commit discipline: HEAD + `git status --short` re-read before every commit (one transient `index.lock` from a sibling lane session observed and retried after state verification).
- **E8 RED evidence**: M1 — 7 failing tests captured pre-GREEN (`--- FAIL: TestCheckHomeDisk_*` ×6 + `TestFindDuplicateClusters_*`, statuses `""`, detail `""`, cluster count 0). M2 — 6 failing tests (dry-run/force/retention×3/no-home/flag-wiring, empty output, files not deleted). Verbatim outputs preserved in the run transcript.
- **AC-MCH-010 disclosed gap**: the UNSCOPED `go test ./internal/cli/ ./internal/config/ -count=1` exits 1 on `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` (`expected BLOCK on injection+AWS-key fixture; got decision="" err=<nil>`). This test shells out to the REAL codex backend and its own failure note states a non-BLOCK may be "a real result". Non-intersection evidence: `git diff --name-only 09516ff0c..HEAD` (17 files) contains no codex file — the failure is environment/backend-dependent, reproduced 2/2 in isolation, and identical in cause to the pre-run state. CI (which lacks a working codex → the test skips) is the judging gate. Everything else in both packages passes.

### PRESERVE verification

`git diff --name-only 09516ff0c..HEAD` = 17 files: SPEC artifacts (5), internal/cli ×8 (clean.go extended; doctor.go +3-line registration; doctor_disk{,_test}.go, clean_home{,_test}.go, clean_home_carveout_test.go new; doctor_golden_test.go HOME-pinning), goldens ×3, defaults.go (three-constant append only — the authorized surgical carve-out). No PRESERVE-listed file appears: todo.go, todo_test.go, update.go, cache.yaml, memory.go, memory_test.go, migrate_profiles.go, migrate_profiles_test.go, linkage.go, linkage_test.go, .mcp.json, all SPEC-V3R6-MOAI-HOME-PATHS-001 artifacts, .moai/reports/** — untouched.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-17T03:05:00+09:00
run_commit_sha: 46ed09a66
run_status: green
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 11   # named files verified untouched; defaults.go modified only within its authorized carve-out (3 constants)
l44_pre_commit_fetch: "git fetch origin main -> FETCH_HEAD; origin/main...HEAD = 1 53 (release lane 53 ahead / origin/main +1 — normal lane divergence, no rebase action for this SPEC)"
l44_post_push_fetch: "N/A — push prohibited on this lane (B9 repo-local override; lane pushed by another session)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: "exit 0"
  windows_amd64: "exit 0"
total_run_phase_files: 17
m1_to_mN_commit_strategy: "per-milestone conventional commits (M1 feat 051a2fa94 / M2 feat 6ec0ad212 / M3 test a42941195 / M4 docs+verify + sha backfill), no push per B9"
```



## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-08-16T22:33:32+09:00
- plan_status: audit-ready
- plan_audit_verdict: PASS — iteration 2, score 0.9375 (Tier M threshold 0.80; report: `.moai/reports/plan-audit/SPEC-V3R6-MOAI-CLEAN-HOME-001-review-2.md`; iteration ceiling 2/2 reached — final)
- audit_history: iter1 FAIL 0.875 (D1–D7) → remediation 2026-08-17 (orchestrator-direct, user-approved: byte-equal+entry-count / recursive+carve-out-wins / home-tier retention) → iter2 PASS 0.9375; residuals R1–R4 minor, routed into run-phase (R1/R4 as spawn directives, R2 fixed in-session, R3 accepted as measurement archive)

## Decision Point 2 (2026-08-17, post-iteration-2)

Audit iteration 1 FAIL (0.875) remediated via AskUserQuestion round — 3 decisions, all recommended options selected (duplicate-cluster strictness / carve-out depth semantics / retention tier). Iteration 2 delta re-audit: PASS 0.9375, no blocking defects. Independent verdict now exists (plan-auditor spawn healthy post-PR-#1574). Depends_on SPEC-V3R6-MOAI-HOME-PATHS-001 status verified `completed` in this tree.

## Decision Point 1 (2026-08-16T22:33+09:00)

User APPROVED via AskUserQuestion (option "승인 (권장)"). Audit basis: orchestrator mechanical self-check (frontmatter 12/12, REQ 8 / AC 11 within Tier M ceilings, phase release-label, depends_on declared) — no independent plan-auditor verdict exists (session-wide spawn failure, see HOME-PATHS progress.md). Run ordering locked by depends_on: SPEC-V3R6-MOAI-HOME-PATHS-001 run first, then this SPEC.

## §F Phase 4 Mode Selection

Input parameters:
- tier: M
- scope: ~6 files (2 new: internal/cli/doctor_disk.go, internal/cli/clean_home.go; extended: internal/cli/doctor.go, internal/cli/clean.go, internal/config/defaults.go + state loader surface + tests)
- domain count: 1 (Go CLI feature — doctor check + clean subcommand, single domain)
- file language mix: 100% Go (+ test files)
- concurrency benefit: LOW (coding-heavy; ordered milestone dependencies)
- Agent Teams prereqs: N/A (Mode 3 retired)

Mode evaluation:
1. trivial — not selected: multi-file feature with safety-critical deletion logic
2. background — not selected: write-capable implementation work
3. agent-team — RETIRED (never selected)
4. parallel — not selected: coding-heavy per Anthropic's coding-task parallelism caveat
5. sub-agent — **selected**: single manager-develop delegation covering M1→M4 with per-milestone Conventional Commits; orchestrator verifies at commit boundaries
6. workflow — not selected: far below ~30 files; M2/M3 depend on M1 scanners

Decision: sub-agent
Justification: Tier M coding-heavy feature (doctor disk check + clean --home with a deletion-safety carve-out invariant) with ordered milestones. Sequential sub-agent delegation keeps the deletion-safety guard test independently verifiable at the M3 boundary. Implementation Kickoff Approval passed (user selected "승인 · 표준 진행", 2026-08-17) before this spawn.

## Authoring Record

- 2026-08-16: research.md (disk measurements + surface survey, orchestrator-direct in-session investigation), spec.md (8 REQs), plan.md (4 milestones), acceptance.md (11 ACs) authored orchestrator-direct per user directive "제안대로 진행" — continuing the session's approved degradation pattern (subagent spawns dead pre-PR-#1574; see SPEC-V3R6-MOAI-HOME-PATHS-001 progress.md Authoring Record).
- depends_on SPEC-V3R6-MOAI-HOME-PATHS-001 declared: run phase ordering is HOME-PATHS first (paths.MoaiHome() foundation), this SPEC second.
- Pending: user review (Decision Point 1).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-17T04:10:00+09:00
sync_commit_sha: 2d9fcf3cc
sync_status: audit-ready
b12_self_test_a: "pre-emission grep -c 'SPEC-V3R6-MOAI-CLEAN-HOME-001' CHANGELOG.md → 0 (exit 1, duplicate guard PASS — entry appended to the existing [Unreleased] > Added section, directly below the SPEC-V3R6-MOAI-HOME-PATHS-001 entry)"
b12_self_test_b: "AC count match: acceptance.md SSOT = 12 (canonical grep returns exactly 12 distinct tokens AC-MCH-001..012); CHANGELOG entry cites 12"
b12_self_test_c: "all 9 paths named in the entry verified via ls (doctor_disk.go, clean_home.go, clean.go, doctor.go, defaults.go, clean_home_carveout_test.go, clean_home_test.go, doctor_disk_test.go, SPEC spec.md) — all exist"
changelog_entry_position: "CHANGELOG.md [Unreleased] > Added — second entry, below SPEC-V3R6-MOAI-HOME-PATHS-001; no separate doc-check ACs exist (the CHANGELOG entry itself is the docs surface; the only report-only stance the implementation genuinely surfaces is the ~/.claude line — the .env.glm split belongs to the HOME-PATHS entry, not invented here)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (this sync commit; updated: already 2026-08-17 — no date change needed)"
  acceptance_plan: "markdown-header convention — no frontmatter status to transition"
mx_debt_check: "moai mx query --kind DEBT --json → empty output (no dangling DEBT markers; clean_home.go carries @MX:WARN ×2 + @MX:REASON, doctor_disk.go @MX:NOTE ×1 — the R4 isCarvedOut doc comment is present and is a comment, fine)"
lane_routing: "release/v3.1.1 integration lane per user directive — run-phase landed in-lane (M1 051a2fa94 / M2 6ec0ad212 / M3 a42941195 / M4 46ed09a66 + backfill 1b9440595); sync commits on this branch, no per-phase PR, no push (lane push owned by another session)"
close_summary: "3-phase close complete — CHANGELOG [Unreleased] entry emitted, spec.md frontmatter in-progress → completed, §E.4 recorded; AC 12/12 discharged in run phase (12/12 PASS per §E.2; orchestrator independently re-verified builds, scoped tests, predicates, lint 0)"
```
