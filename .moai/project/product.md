# MoAI-ADK Product Definition

## @VISION:MISSION-001 핵심 미션

**"Spec-First TDD 개발을 Claude Code 환경에서 누구나 쉽게 실행할 수 있도록 하는 완전한 Agentic Development Kit 제공"**

### 핵심 가치 제안

MoAI-ADK는 Claude Code + Spec-First TDD 기반으로 개발자가 체계적인 개발 워크플로우를 수행할 수 있게 하는 간결한 개발 프레임워크입니다.

## @REQ:USER-001 주요 사용자층

### 1차 사용자: 개인 개발자

- **즉시 얻고 싶은 결과**: Git 복잡성 없이 체계적인 TDD 개발 진행
- **핵심 시나리오**: `/moai:0-project → /moai:3-sync` 4단계로 명세→구현→문서화 완성

### 2차 사용자: 팀/조직 개발자

- **즉시 얻고 싶은 결과**: GitHub 연동을 통한 협업 워크플로우 자동화
- **핵심 시나리오**: GitHub Issue 기반 브랜치 생성, 7단계 커밋 템플릿, PR 자동화

### 3차 사용자: Claude Code 생태계 확장자

- **즉시 얻고 싶은 결과**: 사용자 정의 에이전트/명령어를 통한 도메인별 워크플로우 구축
- **핵심 시나리오**: `.claude/agents/` 확장을 통한 특화 개발 프로세스 구현

## @REQ:PROBLEM-001 해결하는 핵심 문제

### 우선순위 높음

1. **SDD (Specification-Driven Development) 부재**: 명세 없이 코딩하는 혼란 해결
2. **TDD 구현 복잡성**: Red-Green-Refactor 사이클을 체계적으로 지원
3. **TAG 추적성 공백**: 요구사항부터 구현까지 16-Core TAG 체계로 추적

### 우선순위 중간

4. **문서 동기화 지연**: Living Document 패턴으로 코드-문서 일치성 보장

### 현재 실패 사례들

- 명세 없이 시작하여 중간에 요구사항 변경으로 인한 재작업
- 테스트 작성 순서 혼란으로 인한 품질 저하
- 요구사항 추적 실패로 인한 scope creep
- 문서와 코드 불일치로 인한 협업 지연

## @VISION:STRATEGY-001 차별점 및 강점

### 경쟁 솔루션 대비 강점

1. **자동화 심도**: Git 명령어 지식 없이도 완전한 GitFlow 지원
   - **발휘 시나리오**: `/moai:git:checkpoint`, `/moai:git:rollback` 등 5종 명령으로 모든 Git 작업 자동화

2. **문서 동기화**: 코드 변경과 동시에 문서 갱신
   - **발휘 시나리오**: `/moai:3-sync`로 TAG 인덱스, 리포트, PR 상태 일괄 동기화

3. **추적성**: 16-Core TAG 시스템으로 요구사항-구현 연결
   - **발휘 시나리오**: `@REQ → @DESIGN → @TASK → @TEST` 체인으로 완전한 추적성 보장

4. **에이전틱 거버넌스**: TRUST 5원칙 기반 품질 자동 검증
   - **발휘 시나리오**: Constitution 위반 자동 감지, Waiver 제도를 통한 예외 관리

## @REQ:SUCCESS-001 성공 지표

### 즉시 측정 가능한 KPI

1. **채택률**: MoAI-ADK 명령어 사용 빈도 (주간 기준)
   - **베이스라인**: 개인 모드 사용자 기준 주 3회 이상 워크플로우 실행

2. **품질**: TRUST 5원칙 준수율
   - **베이스라인**: Constitution 위반 0건, 테스트 커버리지 85% 이상 유지

3. **생태계 기여**: 사용자 정의 에이전트/명령어 작성 사례
   - **베이스라인**: 월 1개 이상 커뮤니티 확장 사례 공유

### 측정 주기

- **일간**: Constitution 위반 모니터링
- **주간**: 워크플로우 실행 통계
- **월간**: 사용자 피드백 및 확장 사례 수집

## Legacy Context

### 기존 자산 요약

- **완성된 Python 패키지**: `src/moai_adk/` 모듈 구조
- **Claude Code 완전 통합**: `.claude/` 디렉토리 기반 에이전트 시스템
- **포괄적 문서화**: README.md, CHANGELOG.md, docs/ 디렉토리
- **품질 도구체인**: pytest, Makefile, pyproject.toml 등

### @DEBT:MIGRATION-001 초기 마이그레이션 계획

1. 모든 Python 코드 및 Claude Code 커맨드 개선 (@TASK:CODE-REVIEW-001)
2. 서브에이전트 지침 체계화 (@TASK:AGENT-GUIDE-001)
3. 크로스 플랫폼 패키지 배포 완성 (@TASK:CROSS-PLATFORM-001)

## @TODO:SPEC-BACKLOG-001 다음 단계 SPEC 후보

1. **SPEC-002**: Python 코드 품질 개선 시스템
2. **SPEC-003**: Claude Code 커맨드 최적화
3. **SPEC-004**: 서브에이전트 가이드라인 표준화
4. **SPEC-005**: 크로스 플랫폼 설치 자동화

---

_이 문서는 `/moai:1-spec` 실행 시 SPEC 생성의 기준이 됩니다._
