# MoAI-ADK Agents Reference

Alfred의 에이전트 위임 참조. 각 에이전트는 특정 작업에 최적화되어 있다.

## Planning & Specification

- `spec-builder`: EARS 포맷 SPEC 생성
- `plan`: 복잡한 작업을 단계별로 분해

## Implementation

- `tdd-implementer`: TDD 사이클 (RED-GREEN-REFACTOR) 실행
- `backend-expert`: 백엔드 아키텍처 및 API 개발
- `frontend-expert`: 프론트엔드 UI 컴포넌트 개발
- `database-expert`: 데이터베이스 설계 및 최적화

## Quality & Testing

- `security-expert`: 보안 분석 및 OWASP 검증
- `quality-gate`: 코드 품질 검증 (TRUST 5)
- `test-engineer`: 테스트 전략 및 구현

## Architecture & Design

- `api-designer`: REST/GraphQL API 설계
- `component-designer`: 재사용 가능한 컴포넌트 설계
- `ui-ux-expert`: 사용자 경험 및 인터페이스 설계

## DevOps & Infrastructure

- `devops-expert`: CI/CD 파이프라인 및 배포
- `monitoring-expert`: 모니터링 및 관찰성
- `performance-engineer`: 성능 최적화 및 분석

## Data & Integration

- `migration-expert`: 데이터베이스 마이그레이션
- `data-engineer`: 데이터 파이프라인 개발

## Documentation & Process

- `docs-manager`: 기술 문서 및 API 문서 생성
- `git-manager`: Git 워크플로우 및 버전 관리
- `project-manager`: 프로젝트 조정 및 계획

## Specialized Services

- `accessibility-expert`: WCAG 접근성 검증
- `debug-helper`: 오류 분석 및 해결책 제시
- `agent-factory`: 새로운 에이전트 생성 및 설정
- `skill-factory`: Skill 정의 생성 및 관리
- `format-expert`: 코드 포매팅 및 스타일 일관성

## System Agents

- `Explore`: 코드베이스 탐색 및 파일 시스템 분석
- `Plan`: 전략 분해 및 계획 수립

---

**위임 원칙**:
1. Alfred는 항상 Task()로 전문 에이전트에게 위임한다.
2. 요청의 복잡도와 의존성을 분석하여 순차 또는 병렬로 조율한다.
3. 각 에이전트의 결과를 다음 에이전트의 컨텍스트로 전달한다.

**에이전트 선택 기준**:
- 단순 작업 (1개 파일): 1-2개 에이전트 순차 실행
- 중간 작업 (3-5개 파일): 2-3개 에이전트 순차 실행
- 복잡한 작업 (10+개 파일): 5+개 에이전트 병렬/순차 혼합

---

자세한 에이전트 설명은 CLAUDE.md를 참고한다.
