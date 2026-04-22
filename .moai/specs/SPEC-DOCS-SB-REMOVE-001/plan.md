---
id: SPEC-DOCS-SB-REMOVE-001
document: plan
version: 1.0.0
---

# Implementation Plan

## Technical Approach

- 4개 locale의 `what-is-moai-adk.md` 동시 편집 (CLAUDE.local.md §17.3 의무).
- ko-only 기형 파일 제거(auto-quality.md) + ko 네비게이션 정리.
- Hugo 빌드로 깨진 참조가 없는지 검증.

## Milestones

### M1 — ko 정본 4 파일 정리
- what-is-moai-adk.md: line ~186, ~197, ~230-237의 /simplify·/batch 표·산문 제거
- auto-quality.md: `git rm` 삭제
- harness-engineering.md: auto-quality 참조 bullet 제거
- _meta.yaml: `auto-quality` 엔트리 제거 (ko→en/ja/zh 구조 일관화)

### M2 — en/ja/zh what-is-moai-adk.md 동기화
- 동일한 3개 구간에서 /simplify·/batch 언급 제거
- REFACTOR/IMPROVE 단락이 자연스럽게 읽히도록 접속사 정리

### M3 — Hugo 빌드 검증
- `cd docs-site && hugo --minify`
- stderr "broken" / "ref not found" 0건
- 생성된 public/ 디렉터리에 core-concepts 네비 정상 노출

### M4 — 커밋
- 단일 `docs(site): SPEC-DOCS-SB-REMOVE-001` 커밋 (ko+en+ja+zh 한 번에)

## Commit Strategy

```
docs(site): SPEC-DOCS-SB-REMOVE-001 — 4개국어에서 /simplify·/batch 홍보 제거

SPEC-SKILL-GATE-001로 MoAI 워크플로우에서 두 스킬을 완전 제거했음에도
공식 문서 사이트가 여전히 "핵심 자동 품질 레이어"로 홍보 중이던 허위 정보
를 일괄 정리. CLAUDE.local.md §17.3 4-locale 동시 업데이트 의무 준수.

변경:
- ko/en/ja/zh what-is-moai-adk.md: line ~186,197,230-237 정리
- ko auto-quality.md: 삭제 (en/ja/zh에 부재 — 4-locale 일관화)
- ko harness-engineering.md: auto-quality 링크 제거
- ko _meta.yaml: auto-quality 엔트리 제거

검증: hugo --minify 성공, grep-based AC 전수 통과.

🗿 MoAI <email@mo.ai.kr>
```
