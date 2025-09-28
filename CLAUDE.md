# ${PROJECT_NAME} - MoAI Agentic Development Kit (TypeScript)

**TypeScript-First TDD 개발 가이드**

## 핵심 철학

- **Spec-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **GitFlow 지원**: Git 작업 자동화, Living Document 동기화, 16-Core @TAG 추적성

**TypeScript 기반**: 타입 안전성과 고성능, 분산 TAG 시스템 v4.0 (94% 크기 절감)

## 3단계 핵심 워크플로우

```bash
/moai:1-spec     # 명세 작성 (EARS 방식, 브랜치/PR 생성)
/moai:2-build    # TDD 구현 (RED→GREEN→REFACTOR)
/moai:3-sync     # 문서 동기화 (PR 상태 전환)
```

**반복 사이클**: 1-spec → 2-build → 3-sync → 1-spec (다음 기능)

## 핵심 에이전트 (5개)

| 에이전트 | 역할 | 자동화 |
|---------|------|--------|
| **spec-builder** | EARS 명세 작성 | 브랜치/PR 생성 |
| **code-builder** | 범용 언어 TDD 구현 | Red-Green-Refactor (Python, TypeScript, Java, Go, Rust 등) |
| **doc-syncer** | 문서 동기화 | PR 상태 전환/라벨링 |
| **cc-manager** | Claude Code 관리 | 설정 최적화/권한 |
| **debug-helper** | 오류 진단 | 개발 가이드 검사 |

## 디버깅 & Git 관리

**디버깅**: `@agent-debug-helper "오류내용"` 또는 `@agent-debug-helper --trust-check`
**Git 자동화**: 모든 워크플로우에서 자동 처리 (99% 케이스)
**Git 직접**: `@agent-git-manager "명령"` (1% 특수 케이스)

## 16-Core @TAG 시스템 v4.0 (분산)

```
@REQ → @DESIGN → @TASK → @TEST
SPEC: REQ,DESIGN,TASK | PROJECT: VISION,STRUCT,TECH,ADR
IMPLEMENTATION: FEATURE,API,TEST,DATA | QUALITY: PERF,SEC,DEBT,TODO
```

**분산 TAG 데이터베이스 v4.0**:
- **카테고리별 저장**: `.moai/indexes/categories/*.jsonl` (JSONL 분산)
- **관계 매핑**: `.moai/indexes/relations/chains.jsonl` (체인 추적)
- **캐시 시스템**: `.moai/indexes/cache/summary.json` (고속 검색)
- **성능 개선**: 94% 크기 절감, 95% 파싱 속도 향상

## TRUST 5원칙 (범용 언어 지원)

**MoAI-ADK 도구**: TypeScript 기반 CLI, **사용자 프로젝트**: 모든 주요 언어 지원
- **T**est First: 언어별 최적 도구 (Jest, pytest, go test, cargo test, JUnit 등)
- **R**eadable: 언어별 린터 (ESLint, ruff, golint, clippy 등)
- **U**nified: 타입 안전성 (TypeScript, Go, Rust, Java) 또는 런타임 검증 (Python, JS)
- **S**ecured: 언어별 보안 도구 및 정적 분석
- **T**rackable: 분산 16-Core @TAG 시스템 v4.0 (94% 최적화)

상세: @.moai/memory/development-guide.md

## 언어별 코드 규칙

**공통**: 파일≤300 LOC, 함수≤50 LOC, 매개변수≤5, 복잡도≤10
**품질**: 언어별 최적 도구 자동 선택, 의도 드러내는 이름, 가드절 우선
**테스트**: 언어별 표준 프레임워크, 독립적/결정적, 커버리지≥85%

## 메모리 전략

**핵심 메모리**: @.moai/memory/development-guide.md (TRUST+16-Core TAG)
**프로젝트 컨텍스트**:
- @.moai/project/product.md
- @.moai/project/structure.md
- @.moai/project/tech.md
**TAG 시스템**: 분산 v4.0 아키텍처 (JSONL 기반, 487KB, 94% 최적화)
**검색 도구**: 고속 카테고리별 검색, 45ms 평균 로딩, rg/grep/find 지원