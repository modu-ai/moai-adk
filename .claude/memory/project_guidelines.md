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
- TDD 필수 적용
- 80% 이상 테스트 커버리지 달성
- 자동화된 테스트 파이프라인

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
- 의존성 취약점 스캔
- 비밀키 관리 (환경변수, 암호화)
- 입력값 검증 및 XSS 방지

## 🤝 협업 규칙

### Git 워크플로우
- main/master: 프로덕션 브랜치
- develop: 개발 통합 브랜치
- feature/*: 기능 개발 브랜치
- hotfix/*: 긴급 수정 브랜치

### 커밋 메시지
```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

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

**마지막 업데이트**: 2025-09-15  
**버전**: v0.1.12