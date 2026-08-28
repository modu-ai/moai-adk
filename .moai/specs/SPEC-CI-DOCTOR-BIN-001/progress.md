# progress.md — SPEC-CI-DOCTOR-BIN-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-28

## §E.2 Run-phase Evidence

- 측정 트리: `.claude/worktrees/t346`, 브랜치 `WT-ci-doctor-bin`, base `origin/develop d566ecc75` 병합 후.
- 증거 전문(5절 형식 — Claim / Evidence / Baseline-attribution / Gaps / Residual-risk): `.moai/reports/t346/verdict.md`.
- AC 매트릭스: AC-CDB-001 PASS · AC-CDB-002 PASS · AC-CDB-003 PASS · AC-CDB-004 PASS (RED 전문 + 9/9 GREEN, 선택자 매치 수 계수 포함).
- 게이트: `go vet ./internal/cli/` rc=0 · `GOOS=windows GOARCH=amd64 go build ./...` rc=0 · `golangci-lint run --timeout=5m ./internal/cli/...` `0 issues.` · `go test ./internal/cli/... -count=1 -cover` rc=0 (17패키지 전부 ok, `internal/cli` 79.9%).
- 라이브 대조: bin 있음 → `11/11 embedded agent-emit artifacts match the committed set`; bin 없음 → `skipped: no readable binary to judge at …`.

### develop CI 적색 전수 (한 계열만 지목하지 않음)

run `33128899299` (head `d566ecc75`) 실패 15건 = 5계열. 본 SPEC 이 닫는 것은 doctor 계열 9종 **1계열뿐**이며, 나머지는 착지 후에도 남는다: t349(codex init 3종) · t322(`TestGitDiffNameCount_Predicate`) · t326(`TestSessionStart_BlockingComparerDoesNotStallSessionStart`) · t306(`TestConcurrencyStress`). 귀속 근거는 verdict.md 표.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-08-28

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
