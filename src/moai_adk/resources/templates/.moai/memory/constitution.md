# MoAI-ADK 개발 가이드 5원칙

> **🗿 "명세가 없으면 코드도 없다. 테스트가 없으면 구현도 없다."**

MoAI-ADK 의 모든 개발 활동은 이 5원칙을 준수해야 합니다.

## 🏛️ 개발 가이드 5원칙

### Article I: Simplicity (단순성의 원칙)

> "복잡성은 버그의 온상이며, 단순함이야말로 궁극의 정교함이다."

**제1조** 프로젝트 복잡도 제한
- 동시 운영되는 모듈은 **최대 3개**를 초과할 수 없다
- 새로운 모듈 추가 시 기존 모듈 통합 또는 리팩토링을 우선 고려한다
- 과도한 추상화보다는 명확한 구현을 선호한다

**제2조** 기술 스택 최소화
- 핵심 기능에 필수적이지 않은 기술 도입을 금지한다
- 표준 라이브러리로 해결 가능하면 외부 의존성을 추가하지 않는다

### Article II: Architecture (아키텍처의 원칙)

> "좋은 아키텍처는 결정을 늦출 수 있게 해주며, 나쁜 아키텍처는 결정을 서둘게 만든다."

**제3조** 라이브러리 분리 원칙
- 모든 기능은 라이브러리로 구현한다
- 계층형 아키텍처 준수: Domain, Application, Infrastructure 분리
- 의존성 역전 원칙 (Dependency Inversion Principle) 적용

**제4조** 인터페이스 우선 설계
- 구현보다 인터페이스를 먼저 정의한다
- 각 계층의 책임과 경계를 명확히 정의한다

### Article III: Testing (테스트의 원칙)

> "테스트 없는 코드는 설계상 결함이 있는 코드다."

**제5조** Test-Driven Development (TDD) 의무화
- 모든 핵심 기능 개발은 테스트 작성부터 시작한다
- Red-Green-Refactor 사이클을 엄격히 준수한다
- 테스트 없는 Pull Request는 리뷰를 시작할 수 없다

**제6조** 테스트 커버리지 기준
- 전체 코드 베이스의 테스트 커버리지는 **85% 이상**을 유지한다
- 새로운 기능의 테스트 커버리지는 **90% 이상**이어야 한다

### Article IV: Observability (관찰가능성의 원칙)

> "보이지 않는 것은 관리할 수 없다."

**제7조** 구조화된 로깅 의무화
- 모든 중요한 비즈니스 이벤트는 로그로 기록한다
- 로그는 JSON 형식의 구조화된 데이터로 생성한다
- 개인정보나 민감 정보는 절대 로그에 기록하지 않는다

**제8조** 핵심 메트릭 추적
- 응답시간 (평균, 95th percentile)
- 에러율 (4xx, 5xx 응답)
- 처리량 (RPS, TPS)
- 시스템 자원 사용률

### Article V: Versioning (버전관리의 원칙)

> "버전 관리는 시간 여행을 가능하게 하는 마법이다."

**제9조** 시맨틱 버저닝 의무화
- 모든 릴리스는 MAJOR.MINOR.BUILD 형식을 따른다
- MAJOR: Breaking Change 발생 시
- MINOR: 하위 호환 가능한 기능 추가 시
- BUILD: 하위 호환 가능한 버그 수정 시

**제10조** GitFlow 자동화
- 모든 변경사항은 3단계 파이프라인을 거친다: spec → build → sync
- feature 브랜치 자동 생성 및 관리
- 7단계 자동 커밋 시스템 준수
- Draft PR → Ready for Review 자동 전환

## 🎯 MoAI-ADK  특화 원칙

### GitFlow 완전 투명성
- 개발자는 Git 명령어를 몰라도 된다
- 모든 Git 작업은 3개 에이전트가 자동 처리한다
- 브랜치, 커밋, PR 관리는 완전 자동화된다

### 3단계 파이프라인 강제
1. `/moai:1-spec`: 명세 + 브랜치 + Draft PR
2. `/moai:2-build`: TDD 구현 + 3단계 커밋
3. `/moai:3-sync`: 문서 동기화 + PR Ready

### 16-Core TAG 추적성
- Requirements: @REQ:USER-AUTH-001
- Design: @DESIGN:TOKEN-SYSTEM-001
- Tasks: @TASK:AUTH-IMPL-001
- Tests: @TEST:UNIT-AUTH-001
- Quality: @PERF:API-500MS, @SEC:XSS-HIGH

### 품질 게이트
- 개발 가이드 5원칙 자동 검증
- 테스트 커버리지 85% 이상
- 모든 CI/CD 단계 통과 필수
- 보안 스캔 통과 필수

## ⚖️ 자동화된 검증

### check_constitution.py
```bash
python3 .moai/scripts/check_constitution.py
```
- Simplicity: 모듈 수 ≤ 3개 확인
- Architecture: 라이브러리 분리 검증
- Testing: 커버리지 ≥ 85% 확인
- Observability: 구조화 로깅 검증
- Versioning: MAJOR.MINOR.BUILD 체계 확인

### check-traceability.py
```bash
python3 .moai/scripts/check-traceability.py
```
- 16-Core TAG 체인 무결성 검증
- 고아 TAG 및 끊어진 링크 감지
- 추적성 매트릭스 생성

## 📊 성공 지표

### 정량적 지표
- 테스트 커버리지: ≥85%
- 개발 가이드 준수율: 100%
- 개발 시간 단축: 67% (GitFlow 자동화)
- Git 실수 감소: 100% (자동화로 제거)

### 정성적 지표
- Git 학습 부담 제거
- 일관된 개발 워크플로우
- 완전한 추적성 확보
- 품질 게이트 자동화
