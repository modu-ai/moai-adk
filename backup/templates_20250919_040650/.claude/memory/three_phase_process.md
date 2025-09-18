# 3단계 작업 프로세스(필수)

> 모든 요청은 탐색 → 계획 → 구현 3단계로 수행하며, 각 단계의 산출물을 명확히 제공합니다.

## Phase 1: 탐색(Exploration)
- 코드베이스 분석: 관련 파일/디렉터리/모듈 목록화
- 규칙 파악: 네이밍/아키텍처/스타일/에러 처리 패턴 기록
- 키워드/함수/클래스/패턴 전역 검색

출력 형식 예시:
```
### Codebase Analysis Results
**Relevant Files Found:**
- path/to/file: 이유

**Code Conventions Identified:**
- Naming: 규칙
- Architecture: 패턴
- Styling: 포맷

**Key Dependencies & Patterns:**
- lib: 사용 패턴
```

## Phase 2: 계획(Planning)
- 변경 사항 로드맵 수립(작고 안전한 작업 단위)
- 작업/수용 기준 명세화

출력 형식 예시:
```
## Implementation Plan

### Module: 모듈명
**Summary:** 한두 문장 요약

**Tasks:**
- [ ] 구체 작업
- [ ] 구체 작업

**Acceptance Criteria:**
- [ ] 정량/정성 기준
- [ ] 품질/성능 기준
```

## Phase 3: 구현(Implementation)
- 계획에 따라 변경 적용, 수용 기준 검증
- 기존 컨벤션 준수 + 최소주의(KISS/Minimal)

품질 게이트:
- [ ] 모든 수용 기준 충족
- [ ] 코드 규칙 준수
- [ ] 최소주의/전문가 수준 충족

## 응답 구조(항상 동일)
1) Phase 1 Results  2) Phase 2 Plan  3) Phase 3 Implementation
