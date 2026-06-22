---
id: SPEC-CLEANUP-EVALUATOR-001
title: "Progress tracking — remove orphaned internal/evaluator package"
version: "0.1.0"
status: in-progress
created: 2026-06-22
updated: 2026-06-22
author: Goos Kim
---

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored (spec.md + plan.md + acceptance.md + progress.md);
`status: draft`. SPEC ID self-check PASS; 12-field frontmatter validated; Out of Scope
section present (5 H3 sub-headings); Tier S justified. Grounding pre-verified.

## §E.2 Run-phase Evidence

- `internal/evaluator/` 제거 (orchestrator-direct — manager-develop가 L1 격리 worktree에서 메인 트리 미커밋 상태 접근 불가 blocker 반환, 복구 경로로 직접 수행).
- codemaps `overview.md`/`modules.md` 동기화: 46→45 internal, test-only 2→1 (`internal/skills`만 잔존).
- `spec.md` status draft→in-progress, `tier: S` 추가.
- 보호 산출물 무변경: `sync-auditor.md`, `evaluator-profiles/`, SPEC-EVAL-001/EVALLIB-001, `internal/skills`.

## §E.3 Run-phase Audit-Ready Signal

검증 7/7 PASS (orchestrator 독립 배치):
- `go build ./...` → exit 0
- `go test ./...` → 0 FAIL (전체 통과)
- `grep -rn 'internal/evaluator' --include='*.go'` → 0 매치
- `go list ./internal/evaluator/` → directory not found (패키지 제거 확인)
- `go vet ./...` → 클린

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
