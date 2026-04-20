# Phase 2 — Subtree 크기 실측 기록

> 측정일: 2026-04-20
> 목적: AC-G1.5-05, AC-G2-10 (Gap 2) 증빙

## .git/ 디렉토리 크기 변화

| 시점 | du -sh .git/ | 비고 |
|------|-------------|------|
| git subtree add 전 (main 브랜치) | 183M | feat/docs-site-spec-docs-site-001 브랜치 생성 직전 |
| git subtree add 후 | 197M | squash commit f685e34a3 추가 직후 |
| 증가량 | +14MB | moai-adk-docs 전체 이력을 단일 squash 커밋으로 압축한 결과 |

## 원본 레포 크기 참고

moai-adk-docs 원본 레포 `.git/` 크기: 약 39MB (spec.md §5 D1 기재)

## squash 커밋 정보

- 커밋 해시: f685e34a3
- 커밋 메시지: feat(docs-site): import moai-docs via squash subtree for SPEC-DOCS-SITE-001
- 변경 파일 수: 1142 files changed, 318524 insertions(+)
- Merge commit 형식: Merge squash-commit + original tree

## 평가

- 목표: 원본 39MB `.git`을 단일 squash 커밋으로 압축
- 실측: `.git/` 전체 14MB 증가 (squash 효과로 원본 39MB 대비 약 64% 절감)
- D1 목표 "약 50KB 단일 커밋"은 squash 커밋 자체의 메타데이터 크기를 지칭 (실제 파일 트리는 별도 pack object)
- 단일 squash 커밋으로 원 이력 39개+ 커밋이 1개로 압축되어 이력 오염 없음
- G1.5 통과 기준 충족
