---
id: SPEC-V3R3-STATUSLINE-FALLBACK-001
version: "1.0.0"
status: completed
created_at: 2026-05-10
updated_at: 2026-05-10
author: GOOS행님
priority: High
labels: [statusline, fallback, cli]
issue_number: null
---

# Progress: SPEC-V3R3-STATUSLINE-FALLBACK-001

## 현재 상태

| 항목 | 상태 |
|------|------|
| Research | 완료 (`research.md`) |
| Spec | 완료 (`spec.md` v0.2.0) |
| Plan | 완료 (`plan.md` v0.1.0) |
| Acceptance | 완료 (`acceptance.md` v0.2.0) |
| Plan Audit | PASS (0.88, iter 2) |
| M1: Cwd Guard + Directory Fallback | 완료 |
| M2: Model Name Fallback Chain | 완료 |
| M3: 통합 테스트 + MX Tags | 완료 |
| Run PR | MERGED (#841, `25dce1c63`) |

## Milestone 추적

### M1: Cwd Guard + Project Directory Fallback
- [x] `internal/cli/statusline.go` cwd guard 구현 (AC-SF-006)
- [x] `internal/statusline/builder.go` `extractProjectDirectory()` 4-level getcwd fallback
- [x] 단위 테스트 작성 (Windows skip — OS prevents cwd deletion)
- [x] Race detector 통과

### M2: Model Name Fallback Chain
- [x] `internal/statusline/model_cache.go` 신규 파일 (ReadModelCache/WriteModelCache)
- [x] `internal/statusline/metrics.go` fallback chain (stdin → env → cache → unavailable)
- [x] `CollectMetrics` signature 변경 + 기존 호출부 업데이트
- [x] 단위 테스트 작성 (metrics_fallback_test.go)
- [x] Race detector 통과

### M3: 통합 테스트 + MX Tags
- [x] `internal/statusline/builder_test.go` 통합 테스트
- [x] 모든 stdin variant (empty / `{}` / `null` / partial / normal) 테스트
- [x] MX tags 6개 적용 (ANCHOR 3, NOTE 3)
- [x] CI 16/16 ALL GREEN (Lint + Test 3OS + Build 5 + CodeQL + Constitution)

## AC 충족 현황

- AC-SF-001 ~ AC-SF-008: GREEN
- EC-SF-001 ~ EC-SF-005: GREEN
- AC-SF-QG-001 ~ QG-004: GREEN

## 메모

- OQ-1 (settings.json statusLine 소멸): 본 SPEC에서 defer. 별도 investigation 필요.
- OQ-2 (git worktree branch mismatch): 확인 후 별도 SPEC 권장. 본 SPEC scope 외.
