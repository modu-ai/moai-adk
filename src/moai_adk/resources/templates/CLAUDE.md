# ${PROJECT_NAME} - MoAI Agentic Development Kit

**Spec-First TDD 개발 가이드**

## 핵심 철학

- **Spec-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **GitFlow 지원**: Git 작업 자동화, Living Document 동기화, 16-Core @TAG 추적성

## 4단계 개발 워크플로우

```bash
/moai:0-project  # 프로젝트 문서 초기화 (product/structure/tech)
/moai:1-spec     # 명세 작성 (EARS 방식, 브랜치/PR 생성)
/moai:2-build    # TDD 구현 (RED→GREEN→REFACTOR)
/moai:3-sync     # 문서 동기화 (PR 상태 전환)
```

## 핵심 에이전트 (5개)

| 에이전트 | 역할 | 자동화 |
|---------|------|--------|
| **spec-builder** | EARS 명세 작성 | 브랜치/PR 생성 |
| **code-builder** | TDD 구현 | Red-Green-Refactor |
| **doc-syncer** | 문서 동기화 | PR 상태 전환/라벨링 |
| **cc-manager** | Claude Code 관리 | 설정 최적화/권한 |
| **debug-helper** | 오류 진단 | 개발 가이드 검사 |

## 디버깅 & Git 관리

**디버깅**: `/moai:debug "오류내용"` 또는 `/moai:debug --trust-check`
**Git 자동화**: 모든 워크플로우에서 자동 처리 (99% 케이스)
**Git 직접**: `@agent-git-manager "명령"` (1% 특수 케이스)

## 16-Core @TAG 시스템

```
@REQ → @DESIGN → @TASK → @TEST
SPEC: REQ,DESIGN,TASK | PROJECT: VISION,STRUCT,TECH,ADR
IMPLEMENTATION: FEATURE,API,TEST,DATA | QUALITY: PERF,SEC,DEBT,TODO
```

## TRUST 5원칙

**T**est First, **R**eadable, **U**nified, **S**ecured, **T**rackable
상세: @.moai/memory/development-guide.md

## 코드 규칙

**크기**: 파일≤300 LOC, 함수≤50 LOC, 매개변수≤5, 복잡도≤10
**품질**: 명시적 코드, 의도 드러내는 이름, 가드절, 구조화 로깅, 입력 검증
**테스트**: 새 코드=새 테스트, 독립적/결정적, 성공/실패 경로, 커버리지≥80%

## 메모리 전략

**핵심 메모리**: @.moai/memory/development-guide.md (TRUST+16-Core TAG)
**프로젝트 컨텍스트**: 
- @.moai/project/product.md
- @.moai/project/structure.md
- @.moai/project/tech.md
**검색 도구**: rg(권장), grep, find 지원