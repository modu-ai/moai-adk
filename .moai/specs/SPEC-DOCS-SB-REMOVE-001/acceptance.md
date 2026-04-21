---
id: SPEC-DOCS-SB-REMOVE-001
document: acceptance
version: 1.0.0
---

# Acceptance Criteria

## AC-1: 4 locale /simplify·/batch 리터럴 0건

**When**: `grep -rn '/simplify\|/batch' docs-site/content/{ko,en,ja,zh}/core-concepts/`

**Then**: 0 라인

## AC-2: ko auto-quality.md 삭제

**When**: `test -f docs-site/content/ko/core-concepts/auto-quality.md`

**Then**: exit code 1 (파일 부재)

## AC-3: ko auto-quality 참조 전수 제거

**When**: `grep -rn 'auto-quality' docs-site/content/ko/`

**Then**: 0 라인

## AC-4: Hugo 빌드 성공

**When**: `cd docs-site && hugo --minify 2>&1`

**Then**: exit 0, stderr에 "broken"/"ref not found"/"ERROR" 워닝 0건

## AC-5: 핵심 방법론 설명 보존

**When**: 수동 검토 — 4 locale의 what-is-moai-adk.md TDD RED-GREEN-REFACTOR 및 DDD ANALYZE-PRESERVE-IMPROVE 섹션

**Then**: 사이클 설명 자체는 건재 (스킬 참조만 제거)

## AC-6: _meta.yaml 일관성

**When**: `diff <(grep -o '^"[^"]*":' docs-site/content/ko/core-concepts/_meta.yaml | sort) <(grep -o '^"[^"]*":' docs-site/content/en/core-concepts/_meta.yaml | sort)`

**Then**: ko·en 사이에 `auto-quality`·`harness-engineering` 차이가 제거된 상태로 일관 (`harness-engineering`는 ko 전용 유지 — 별도 SPEC 대상)

## AC-7: Pencil/mx/github 일반 명사 보존

**When**: `grep -rn 'batch_design\|batch_get' docs-site/content/`

**Then**: pre/post 커밋 동일 — Pencil MCP tool 이름은 건드리지 않음
