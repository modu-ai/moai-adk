---
id: SPEC-V3R6-CG-MODE-HARDENING-001
artifact: progress
version: "0.1.0"
created: 2026-06-22
updated: 2026-06-22
author: orchestrator
run_commit_sha: "23901497a"
sync_commit_sha: "47368ca0d"
---

# Progress — SPEC-V3R6-CG-MODE-HARDENING-001

## §E.1 Plan-phase Audit-Ready Signal

- **Tier**: M (justified in plan.md §A — multi-file + detector redesign + conditional template sync + security validation; not L, not S).
- **Requirements**: 10 (REQ-CGH-001 .. REQ-CGH-010), grouped: launch-safety (001), atomicity cluster (002/003/005), detector SSOT headline (006), doc (004), precondition (008), security (007), coverage (009), regression-safety (010).
- **Acceptance criteria**: 10 AC groups (AC-CGH-001 .. AC-CGH-010), each mechanically verifiable; 6 supporting edge cases (EC-1..EC-6).
- **Defect verification**: all CONFIRMED + POTENTIAL findings re-verified against cited source during plan-phase (spec.md §A.1 table). AGENT-REPORTED security finding (REQ-CGH-007) confirmed real at `glm.go:742-778` + `validation.go:349-352`. Disproven process-env-pollution hypothesis explicitly excluded (§A.2 / §H).
- **Sibling-SPEC reconciliation**: `cg_detect.go` / `REQ-WTL-009` owned by SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001; REQ-CGH-006 reconciles (does not delete), enforced by C-7 + AC-CGH-006 Scenario 6b.
- **Artifacts**: spec.md, plan.md, acceptance.md, design.md, progress.md (5-file Tier M set).
- **Frontmatter**: 12 canonical fields present; `status: draft`; `id` regex self-check PASS (decomposition printed in agent response).
- **Out of Scope**: 5 `### Out of Scope —` H3 sub-headings present in spec.md §H (satisfies OutOfScopeRule).

_Run-phase (§E.2/§E.3) and sync-phase (§E.4) sections below are placeholder headings only at plan-phase._

## §E.2 Run-phase Evidence

- **cycle_type**: tdd (RED-GREEN-REFACTOR; quality.yaml `development_mode: tdd`)
- **구현 파일 (production 8)**: `internal/cli/launcher.go` (cleanup 순서 — `removeGLMEnv` 선행), `internal/cli/launch_exec_posix.go` (신규, `syscall.Exec`), `internal/cli/launch_exec_windows.go` (신규, spawn-child + `os.Exit`), `internal/cli/glm.go` (단일 atomic teammateMode write, tmux `IsAvailable()` 전제), `internal/cli/settings.go` (신규, flock+atomic `settings.local.json` RMW), `internal/tmux/cg_detect.go` (`IsCGMode` layered-OR + `sessionEnvReaderFn` seam, 2-arg 유지), `internal/config/validation.go` (GLM `base_url` allowlist 검증), `CLAUDE.local.md` (§22.3 정합)
- **테스트 파일 (4)**: `internal/cli/cg_mode_hardening_test.go`, `internal/cli/launch_exec_test.go`, `internal/tmux/cg_detect_ssot_test.go`, `internal/config/validation_glm_baseurl_test.go`
- **AC 매트릭스**: AC-CGH-001..010 전원 PASS (manager-develop §E 자가검증 + orchestrator 독립 재검증 일치)
- **독립 검증 (orchestrator, main 통합 트리)**: `go test ./...` ALL PASS · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 · `go vet` clean · `golangci-lint run ./internal/{cli,tmux,config}/...` 0 issues · subagent-boundary grep 0 (매치 전부 주석/help text 부정문)
- **커버리지**: credential-routing 코어가 production path로 이동 — `buildTmuxInjectVars`/`buildTmuxClearVars` 100%, `IsCGMode` 94% (baseline 10.5%/25% 대비 상승)

## §E.3 Run-phase Audit-Ready Signal

- **run_commit_sha**: `23901497a` (코드 실제 안착 commit — 병렬 `SPEC-V3R6-DEV-HARNESS-SPLIT-001` 세션이 working tree에 staged된 cg 코드 11파일을 흡수한 shared-main orphan race; `git diff HEAD 4e12f9ea0` = 0 으로 byte-identical 검증, origin push 확인. lifecycle status `draft→in-progress` 전환은 `8e78530bb`)
- **AC 결과**: 10/10 PASS, 0 discrepancies (orchestrator 독립 재검증 V1-V7 매트릭스)
- **sibling 보존**: SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 `IsCGMode` 테스트 green (layered-OR 2번째 disjunct + 워닝 문자열 리터럴 `"GLM env vars are absent"` 보존, REQ-WTL-009 reconcile)
- **반증 처리**: "leader process-env pollution" 가설 §H 제외 — `applyCGMode`가 `setGLMEnv()` 미호출 직접 확인
- **통합 경로**: L1 worktree(4e12f9ea0) → main specific-path checkout (코드 11 + `CLAUDE.local.md`); 병렬 harness 세션 변경(`M` 3파일) 보존

## §E.4 Sync-phase Audit-Ready Signal

- **CHANGELOG entry**: Added to `[Unreleased]` section on 2026-06-22; entry count = 1 (verified `grep -c 'CG-MODE-HARDENING'` = 1)
- **File path verification**: All 7 production + 4 test files confirmed via `ls` (see §E.2)
- **AC count reconciliation**: 10 AC PASS entries (acceptance.md line count = 10, matching CHANGELOG narrative)
- **Status transition**: spec.md frontmatter `status: in-progress → implemented → completed` atomically in this sync commit
- **Frontmatter updated**: `updated:` field refreshed to 2026-06-22
- **Quality validation**: sync-auditor independent 4-dimension scoring (Functionality/Security/Craft/Consistency) — baseline expected ≥ 0.88 per run-phase audit signal
- **Deterministic close check**: `moai spec audit --json --filter-spec=SPEC-V3R6-CG-MODE-HARDENING-001` post-merge expected drift 0, era V3R6
- **Close-subject format**: `chore(SPEC-V3R6-CG-MODE-HARDENING-001): sync-phase artifacts + 3-phase close` (full SPEC-ID mandate satisfied)
