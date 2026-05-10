---
id: SPEC-V3R3-STATUSLINE-FALLBACK-001
version: "0.1.0"
status: draft
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
| Spec | 완료 (`spec.md` v0.1.0) |
| Plan | 완료 (`plan.md` v0.1.0) |
| Acceptance | 완료 (`acceptance.md` v0.1.0) |
| Plan Audit | PASS (0.88, iter 2) |
| M1: Cwd Guard + Directory Fallback | 미시작 |
| M2: Model Name Fallback Chain | 미시작 |
| M3: 통합 테스트 + Template 정리 | 미시작 |

## Milestone 추적

### M1: Cwd Guard + Project Directory Fallback
- [ ] `internal/cli/statusline.go` cwd guard 구현
- [ ] `internal/statusline/builder.go` `extractProjectDirectory()` getcwd fallback
- [ ] 단위 테스트 작성
- [ ] Race detector 통과

### M2: Model Name Fallback Chain
- [ ] `internal/statusline/model_cache.go` 신규 파일 작성
- [ ] `internal/statusline/metrics.go` fallback chain 구현
- [ ] `CollectMetrics` signature 변경 + 기존 호출부 업데이트
- [ ] 단위 테스트 작성
- [ ] Race detector 통과

### M3: 통합 테스트 + Template 정리
- [ ] `internal/statusline/builder_test.go` 통합 테스트
- [ ] 모든 stdin variant (empty / `{}` / `null` / partial / normal) 테스트
- [ ] `status_line.sh.tmpl` deprecation comment (선택)
- [ ] `make build` template 반영
- [ ] MX tags 적용

## 메모

- OQ-1 (settings.json statusLine 소멸): 본 SPEC에서 defer. 별도 investigation 필요.
- OQ-2 (git worktree branch mismatch): 확인 후 별도 SPEC 권장. 본 SPEC scope 외.
