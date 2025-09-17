# 프로젝트 가이드라인

> MoAI-ADK 기반 프로젝트의 표준 개발 가이드라인

## 📋 개발 원칙

### 1. Spec-First 개발 방법론
- 모든 기능 개발은 @SPEC 태그로 시작
- 요구사항 → 계획 → 작업 → 구현 순서 엄수
- EARS 형식의 명확한 요구사항 작성

### 2. 16-Core TAG 시스템 (4개 카테고리)

#### SPEC 카테고리 (문서 추적 - 필수)
- **@REQ**: 요구사항 정의
- **@DESIGN**: 설계 문서
- **@TASK**: 구현 작업

#### STEERING 카테고리 (원칙 추적 - 필수)
- **@VISION**: 프로젝트 비전
- **@STRUCT**: 구조 설계
- **@TECH**: 기술 선택
- **@STACK**: 기술 스택

#### IMPLEMENTATION 카테고리 (코드 추적 - 필수)
- **@FEATURE**: 기능 개발
- **@API**: API 설계 및 구현
- **@TEST**: 테스트 케이스
- **@DATA**: 데이터 모델링

#### QUALITY 카테고리 (품질 추적 - 선택)
- **@PERF**: 성능 최적화
- **@SEC**: 보안 검토
- **@DEBT**: 기술 부채
- **@TODO**: 할 일 추적


## 🏗️ Constitution 5원칙

### 1. Simplicity (단순성)
- 최대 3개 프로젝트/서비스로 제한
- 복잡한 아키텍처보다 명확한 구조 선호
- Over-engineering 방지

### 2. Architecture (아키텍처)
- 표준 라이브러리 우선 사용
- 검증된 패턴과 프레임워크 선택
- 기술 스택 통일성 유지

### 3. Testing (테스트)
- 테스트는 TDD(Red-Green-Refactor) 사이클을 따른다
- 기본 커버리지 목표는 80% 이상이며, 세부 규칙은 @.claude/memory/tdd_guidelines.md 참고
- 테스트/PR/보안 체크는 @.claude/memory/shared_checklists.md를 사용

### 4. Observability (관찰 가능성)
- 로깅, 메트릭, 트레이싱 구현
- 모니터링 및 알림 시스템 구축
- 장애 대응 플레이북 작성

### 5. Versioning (버전 관리)
- MAJOR.MINOR.BUILD 시맨틱 버전
- 브랜치 전략 및 릴리스 프로세스
- 변경사항 문서화

## 🔄 개발 워크플로우

### Phase 1: SPECIFY
1. 요구사항 분석 및 @REQ 태그 생성
2. 비기능 요구사항 정의
3. 수용 기준 (Acceptance Criteria) 작성

### Phase 2: PLAN
1. 아키텍처 설계 (@DESIGN)
2. 기술 스택 선정
3. 개발 일정 계획

### Phase 3: TASKS
1. 작업 분해 (@TASK)
2. 우선순위 설정
3. 책임자 할당

### Phase 4: IMPLEMENT
1. TDD 기반 개발
2. 코드 리뷰 (@REVIEW)
3. 통합 테스트

## 📏 품질 기준

### 코드 품질
- 린터 및 포매터 사용 (ESLint, Prettier, Black 등)
- 코드 복잡도 관리 (McCabe < 10)
- 중복 코드 최소화 (DRY 원칙)

### 문서화
- README.md 필수 작성
- API 문서화 (OpenAPI/Swagger)
- 인라인 주석 적절히 활용

### 보안
- 보안 설계/코딩/개인정보 규칙은 @.claude/memory/security_rules.md 준수
- 의존성 취약점 스캔과 비밀키/입력 검증은 PR 체크리스트(@.claude/memory/shared_checklists.md)에 포함

## 🤝 협업 규칙

### Git 워크플로우
- main/master: 프로덕션 브랜치
- develop: 개발 통합 브랜치
- feature/*: 기능 개발 브랜치
- hotfix/*: 긴급 수정 브랜치

### 커밋 메시지
- Conventional Commit 형식을 따른다. 세부 규칙과 예시는 @.claude/memory/git_commit_rules.md 참고

### Pull Request
- 리뷰어 최소 1명 지정
- CI/CD 파이프라인 통과 필수
- 브랜치 보호 규칙 적용

## 📊 메트릭 및 KPI

### 개발 효율성
- 리드 타임 (Lead Time)
- 배포 빈도 (Deployment Frequency)
- 평균 복구 시간 (MTTR)

### 품질 지표
- 테스트 커버리지
- 버그 발생률
- 고객 만족도

---
## 🧭 에이전트 운영 원칙(요약)

- 대화/문서/커밋은 한국어로 작성
- 변경 전: 관련 파일을 처음부터 끝까지 읽고, 정의·참조·호출·테스트·문서를 전역 검색으로 확인
- 변경은 작게: 커밋/PR 최소화, 영향도 1–3줄 요약, 가정은 Issue/PR/ADR에 기록
- 최소 2가지 대안 비교 후 가장 단순한 해법 선택(장점/단점/위험 각 1줄)
- 보안 기본: 시크릿 미노출, 입력 검증/정규화/인코딩, 파라미터화, 최소 권한
- 코드 한도: 파일 ≤ 300 LOC, 함수 ≤ 50 LOC, 파라미터 ≤ 5, 순환복잡도 ≤ 10
- 테스트 원칙: 새 코드→새 테스트, 버그→회귀 테스트(먼저 실패), E2E 성공/실패 경로 포함
- 로그/시간대: 구조화 로깅(+상관관계 ID), 저장은 UTC, 표시만 로컬
