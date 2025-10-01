# Test TODO App - MoAI Agentic Development Kit v0.0.1 ✅

**TypeScript 기반 SPEC-First TDD 프로젝트 (완전 Claude Code 통합)**

## 핵심 철학 (MoAI-ADK 준수)

- ✅ **Spec-First**: EARS 방법론으로 명세 없이는 코드 없음
- ✅ **TDD-First**: Jest 기반 테스트 없이는 구현 없음
- ✅ **GitFlow 지원**: Git 작업 자동화, Living Document 동기화, @TAG 추적성

**프로젝트 성과**: TypeScript 5.0 + Jest 100% + EARS 명세 완성 + @TAG 적용

## 🚀 MoAI-ADK 3단계 워크플로우 (완성)

```bash
/moai:1-spec     # ✅ EARS 명세 작성 (Essential, Attributes, Rationale, Suppress)
/moai:2-build    # ✅ TDD 구현 (RED→GREEN→REFACTOR 사이클)
/moai:3-sync     # 🔄 문서 동기화 (@TAG 추적성)
```

**CLI 명령어**: `moai doctor`, `moai init`, `moai status` 전체 지원

## 핵심 에이전트 (7개) - 프로젝트 최적화

| 에이전트 | 프로젝트 적용 상태 | 주요 기능 |
|---------|---------|---------|
| **spec-builder** | ✅ **적용완료** | EARS 명세 작성, TODO 요구사항 정의 |
| **code-builder** | ✅ **적용완료** | TypeScript TDD 구현 (Jest 100% 커버리지) |
| **doc-syncer** | 🔄 **진행중** | README/CLAUDE 문서 동기화 |
| **cc-manager** | ✅ **설정완료** | Claude Code 프로젝트 설정 최적화 |
| **debug-helper** | ✅ **대기중** | 시스템 진단, TypeScript/Jest 오류 해결 |
| **git-manager** | ✅ **준비완료** | Git 자동화, 브랜치 관리 |
| **trust-checker** | ✅ **검증완료** | TRUST 5원칙 검증 및 품질 보장 |

## CLI 및 디버깅 시스템

**CLI 명령어**:
- `moai doctor` - TypeScript/Jest 환경 진단
- `moai status` - 프로젝트 상태 및 TAG 추적성 확인
- `@agent-debug-helper "Jest 테스트 오류"` - 자동 오류 해결

**디버깅 대상**:
- TypeScript 컴파일 오류
- Jest 테스트 실패
- 의존성 해결 문제
- @TAG 불일치

## @TAG 시스템 (프로젝트 적용)

```
@SPEC:CORE-FEATURES → @SPEC:ARCHITECTURE → @CODE:DEVELOPMENT → @TEST:UNIT
SPEC: REQ,DESIGN,TASK | PROJECT: VISION,TECH,DOCS | IMPLEMENTATION: FEATURE,TEST,DATA
```

**프로젝트 TAG 매핑**:
- `@SPEC:CORE-FEATURES-001` → Todo 핵심 기능 요구사항
- `@CODE:CORE-001` → TodoManager 클래스 구현
- `@TEST:UNIT-001` → Jest 단위 테스트
- `@CODE:MODELS-001` → Todo 인터페이스 정의

## TRUST 5원칙 (프로젝트 준수)

- ✅ **T**est First: Jest 100% 커버리지, 3개 핵심 기능 테스트 완료
- ✅ **R**eadable: TypeScript 엄격 타입 검사, 명확한 인터페이스
- ✅ **U**nified: 단일 TodoManager 클래스, 명확한 책임 분리
- ✅ **S**ecured: 입력 검증, 타입 안전성 보장
- ✅ **T**rackable: @TAG 시스템으로 완전한 추적성

## 프로젝트 규칙 (TypeScript 특화)

**코드 품질**:
- 파일 ≤ 50 LOC (현재: index.ts 33줄 ✅)
- 함수 ≤ 20 LOC (현재: 모든 메서드 10줄 이하 ✅)
- 매개변수 ≤ 3개 (현재: 최대 1개 ✅)
- 복잡도 ≤ 5 (현재: 단순 CRUD ✅)

**테스트 정책**:
- Jest 테스트 우선 작성
- 100% 커버리지 유지 (현재 달성 ✅)
- 독립적/결정적 테스트
- 명확한 Given-When-Then 구조

## Living Document 전략

**프로젝트 문서**:
- ✅ `README.md` - EARS 방법론 적용, @TAG 완성
- 🔄 `CLAUDE.md` - 에이전트 시스템 완전 적용
- 📋 `docs/sync-report.md` - 자동 동기화 리포트 (예정)

**자동 동기화**:
- 코드 변경 → 문서 TAG 업데이트
- 테스트 결과 → 커버리지 문서 반영
- SPEC 변경 → 관련 구현 추적

## 프로젝트 설정

```json
{
  "name": "test-todo-app",
  "version": "1.0.1",
  "moai": {
    "mode": "personal",
    "project": "test-todo-app",
    "backup": true,
    "methodology": "SPEC-First TDD",
    "tags": "@TAG",
    "coverage": "100%"
  }
}
```

## 온디맨드 지원

**디버깅**: `@agent-debug-helper "설명"`
**SPEC 생성**: `@agent-spec-builder "TODO 삭제 기능 추가"`
**TDD 구현**: `@agent-code-builder "deleteTodo 메서드 TDD"`
**문서 동기화**: `@agent-doc-syncer "interactive"`

---

**MoAI-ADK 성과**: 이 프로젝트는 MoAI-ADK v0.0.1의 SPEC-First TDD 방법론과 @TAG 시스템을 완전히 적용하여 개발되었습니다. TypeScript + Jest + EARS 방법론의 완벽한 시연 사례입니다.
