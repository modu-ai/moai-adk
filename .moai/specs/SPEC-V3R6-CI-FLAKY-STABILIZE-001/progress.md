---
id: SPEC-V3R6-CI-FLAKY-STABILIZE-001
artifact: progress
version: "0.1.2"
created: 2026-05-31
updated: 2026-05-31
status: completed
---

# Progress Tracking — SPEC-V3R6-CI-FLAKY-STABILIZE-001

run-phase 구현 진행을 추적한다. M1-M3 milestone evidence(plan.md §F) + sync-phase
manager-docs가 소비할 audit-ready signal을 담는다.

## §A.0 Pre-flight Verification

run-phase 진입 시 plan.md §C 사전 점검 (2026-05-31):

| Check | Command | Result |
|-------|---------|--------|
| Baseline flaky | `go test -race ./internal/spec/ ./internal/cli/` | FAIL (FLAKY-1 git-add race + FLAKY-2 Windows merge-TUI hang, CI-only) |
| Lint baseline | `golangci-lint run` (변경 패키지) | PASS (0 issues) |
| Plan-phase commit | `git log` | 47ac3d31d (manager-spec plan-phase artifacts) |
| Multi-session race | `git fetch && git rev-list --count --left-right origin/main...HEAD` | `0 5` (local ahead, clean) |
| Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 (PASS) |

**구현 인프라 주의**: Claude Code 런타임이 manager-develop을 L1 worktree(base
origin/main `a03868289`)에서 실행 → 공유 checkout(`41f99173b`)과 base mismatch.
코드 commit `0363c54de`는 orchestrator가 cherry-pick으로 통합 후 status transition
대행 (선례: MAIN-RED-REMEDIATION-001 M4 orchestrator 대행).

## §E — Phase 0.95 Mode Selection

Input parameters:
- tier: M
- scope (files): 4 (internal/spec/closer.go, internal/cli ×2 test, internal/merge/confirm.go)
- domain count: 1 (Go source — test-infrastructure stabilization)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy, Finding A4 caveat)
- Agent Teams prereqs: not met (harness standard)

Mode evaluation:

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | semantic change (race fix + TTY guard) |
| 2 background | no | Write/Edit 필요 |
| 3 agent-team | no | prereqs 미충족 + single domain |
| 4 parallel | no | coding-heavy (Finding A4 caveat) |
| 5 sub-agent | **yes** | coding-heavy Go single-domain default |

Decision: sub-agent

Justification: SPEC scope는 Go test-infrastructure 안정화 (single domain, 4 files,
coding-heavy). Anthropic Finding A4 (coding-task parallelism caveat)에 따라 Mode 5
(sub-agent sequential)가 적합. manager-develop 1 spawn (cycle_type=tdd).

## §E.2 Sync-phase Audit-Ready Signal

```yaml
sync_started_at: "2026-05-31"
sync_commit_sha: "3787ace23"
status: implemented
```

Run-phase 구현 evidence (M1-M3):

### M1 — FLAKY-1 internal/spec git-add race (AC-CFS-001..005)
- `closer.go` `Close()`: baseDir를 `filepath.Abs` 정규화 (REQ-CFS-002, 상대 "." 기본값 제거)
- `performAtomicClose`: `git add` / `git restore --staged` 경로를 baseDir 기준
  `filepath.Rel` 상대 경로로 변환 (`relToBaseOrAbs` 헬퍼, §D.2 cross-volume 절대경로 폴백)
- `TestClose_FullClose_ProducesCommit` non-parallel 유지 (REQ-CFS-004 원자성 불변식)
- local: `go test -race ./internal/spec/ -count=20` → 0 race, 0 fail

### M2 — FLAKY-2 caller audit + Windows skip (AC-CFS-006)
- ConfirmMerge 도달 caller audit: 실제 `--yes` 없이 도달하는 caller =
  `update_skip_sync_test.go` `skip_sync_with_force_does_invoke_archive` (`--force` bypass)
- `coverage_improvement_test.go` + `update_skip_sync_test.go` Windows skip 추가 (기존 precedent 동일 idiom)

### M3 — FLAKY-2 Windows TTY guard (AC-CFS-007/008)
- `merge.ConfirmMerge`: `runtime.GOOS=="windows" && !isatty.IsTerminal(os.Stdin.Fd())` →
  TUI 진입 전 fail-open 에러
- 기존 vendored `mattn/go-isatty` 재사용 (신규 의존성 0, go.mod 불변)

AC 검증 matrix (공유 checkout base 재검증 완료):

| AC | Gate | Status | Verification |
|----|------|--------|-------------|
| AC-CFS-001 | local | PASS | `-race -count=20` → 0 race |
| AC-CFS-002 | local | PASS | `TestClose -count=20` PASS |
| AC-CFS-003 | local | PASS | 0 `t.Parallel` in atomic test |
| AC-CFS-004 | local | PASS | `filepath.Rel` 코드 inspection |
| AC-CFS-005 | local | PASS | `filepath.Abs(baseDir)` inspection |
| AC-CFS-006 | local | PASS | grep windows guard = 1 |
| AC-CFS-007 | local | PASS | guard before `tea.NewProgram().Run()` |
| AC-CFS-008 | local | PASS | `go.mod` diff empty |
| AC-CFS-009 | local | PASS | scoped `-race` (internal/template 제외) in-scope ok |
| AC-CFS-010 | local | PASS | `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |
| AC-CFS-011 | **CI-deferred** | PENDING | Windows CI job (darwin 검증 불가) |
| AC-CFS-012 | **CI-deferred** | PENDING | full matrix Test green |

§D.3 HONESTY GATE: AC-CFS-011/012는 Windows CI job이 진짜 gate. local darwin 통과는
necessary but NOT sufficient. SPEC은 `completed`로 표시하지 않음 (status `in-progress` 유지).

## §E.3 Run-phase status field

```yaml
run_started_at: "2026-05-31"
status: in-progress
m1_status: implemented
m2_status: implemented
m3_status: implemented
ci_deferred: [AC-CFS-011, AC-CFS-012]
```

## §E.4 Audit-Ready Signal

### (Migrated from §E.5)

Mx-phase ownership: orchestrator-direct (post-sync). Manager-docs sync-phase does not populate this section.

```yaml
mx_started_at: "2026-05-31"
mx_commit_sha: "7ab847b3a"
status: completed
ci_deferred_resolved: [AC-CFS-011, AC-CFS-012]
```

§D.3 HONESTY GATE 충족: AC-CFS-011/012 (Windows CI-deferred)는 main CI run (commit
5352d49a2)의 `Test (windows-latest) → success` 로 green 확인됨 → deferred-AC gate 통과.
4-phase close: plan (47ac3d31d) → run (a8a7f3c34) → sync (3787ace23) → Mx (this commit).
12/12 AC 충족 (10 local PASS + 2 Windows green), evaluator PASS-WITH-DEBT (CHANGELOG status
표현 cbd7d8264 해소).
