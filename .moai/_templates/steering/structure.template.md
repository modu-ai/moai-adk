# 구조 및 기술 설계서: $PROJECT_NAME

**@STRUCT:ARCH** - $PROJECT_NAME 아키텍처 구조
**@TECH:STACK** - $PROJECT_NAME 기술 스택
**작성일**: $CREATED_DATE
**버전**: $VERSION

## MoAI-ADK Constitution 5원칙 준수 아키텍처

$PROJECT_NAME는 Constitution 5원칙을 준수하는 단순하고 확장 가능한 아키텍처로 설계됩니다.

```
$PROJECT_NAME/
├── presentation/          # 프레젠테이션 계층
├── application/          # 애플리케이션 계층
├── domain/              # 도메인 계층
├── infrastructure/      # 인프라 계층
└── shared/             # 공통 모듈
```

## 계층별 상세 설계

### Presentation Layer (프레젠테이션 계층)
- **책임**: 사용자 인터페이스 및 외부 시스템 인터페이스
- **구성요소**: Controllers, Middleware, Validators
- **기술 스택**: [선택된 프론트엔드/API 기술]

### Application Layer (애플리케이션 계층)
- **책임**: 비즈니스 로직 조합 및 워크플로우
- **구성요소**: Services, Use Cases, DTOs
- **패턴**: Command, Query, Event Handlers

### Domain Layer (도메인 계층)
- **책임**: 핵심 비즈니스 규칙 및 로직
- **구성요소**: Entities, Value Objects, Domain Services
- **원칙**: Domain-Driven Design (DDD) 적용

### Infrastructure Layer (인프라 계층)
- **책임**: 외부 시스템 연동 및 기술적 세부사항
- **구성요소**: Repositories, Database, External APIs
- **기술 스택**: [선택된 데이터베이스 및 인프라 기술]

## 모듈 의존성 규칙

### 의존성 방향
```
Presentation → Application → Domain ← Infrastructure
```

### 규칙
1. **도메인 계층**은 다른 계층에 의존하지 않음
2. **애플리케이션 계층**은 도메인 계층에만 의존
3. **프레젠테이션 계층**은 애플리케이션 계층에만 의존
4. **인프라 계층**은 도메인 인터페이스를 구현

## 핵심 설계 패턴

### 1. Repository Pattern
- 데이터 액세스 추상화
- 테스트 가능성 향상
- 데이터 소스 독립성 확보

### 2. CQRS (Command Query Responsibility Segregation)
- 명령과 조회 분리
- 성능 최적화
- 확장성 향상

### 3. Event-Driven Architecture
- 느슨한 결합
- 확장성 및 유연성
- 비동기 처리

## 데이터 흐름

### 요청 처리 흐름
1. **HTTP Request** → Controller
2. **Controller** → Application Service
3. **Application Service** → Domain Service
4. **Domain Service** → Repository (Infrastructure)
5. **Repository** → Database
6. **Response** ← 역방향으로 반환

### 이벤트 처리 흐름
1. **Domain Event** 발생
2. **Event Handler** 처리
3. **Side Effects** 실행
4. **Notification** 발송

## 보안 설계

### 인증 및 인가
- JWT 기반 인증
- 역할 기반 접근 제어 (RBAC)
- API Rate Limiting

### 데이터 보안
- 민감 데이터 암호화
- SQL Injection 방지
- XSS 공격 방어

### 감사 및 로깅
- 모든 중요 작업 로그
- 사용자 활동 추적
- 보안 이벤트 모니터링

## 성능 고려사항

### 캐싱 전략
- **Application Level**: Redis/Memcached
- **Database Level**: Query Result Cache
- **CDN**: 정적 자원 캐싱

### 확장성 설계
- **수평 확장**: Load Balancer + Multiple Instances
- **수직 확장**: Resource Optimization
- **데이터베이스 샤딩**: 필요시 적용

### 모니터링
- APM (Application Performance Monitoring)
- 로그 집계 및 분석
- 실시간 알람 시스템

## GitFlow 통합 배포 전략

### 3단계 배포 파이프라인
| 환경 | 트리거 | 배포 방식 | 검증 단계 |
|------|--------|-----------|----------|
| **Development** | `/moai:2-build` | Docker Compose | TDD + Constitution |
| **Staging** | Draft → Ready PR | K8s Preview | E2E + 성능 테스트 |
| **Production** | main 브랜치 머지 | Blue-Green | 카나리 + 모니터링 |

### GitFlow 자동화 통합
- **feature 브랜치**: spec-builder 자동 생성
- **PR 기반 배포**: code-builder TDD 완료 시
- **main 보호**: doc-syncer 문서 동기화 완료 후만 머지

## 기술 스택 및 선정 근거

### 핵심 기술 결정 원칙
- **Constitution 5원칙 기반** 기술 선택
- **GitFlow 투명성** 지원 도구 우선
- **자동화 가능성** 및 **유지보수성** 고려

### Backend Stack
| 기술 영역 | 선택 기술 | 선정 근거 |
|----------|-----------|----------|
| **Language** | $BACKEND_LANG | [Constitution 원칙 기반 선정] |
| **Framework** | $BACKEND_FRAMEWORK | [단순성 + 아키텍처 품질] |
| **Database** | $DATABASE | [관찰가능성 + 확장성] |
| **Cache** | Redis | [성능 + 운영 단순성] |
| **Queue** | $MESSAGE_QUEUE | [비동기 처리 + 모니터링] |

### Frontend Stack
| 기술 영역 | 선택 기술 | 선정 근거 |
|----------|-----------|----------|
| **Language** | TypeScript | [타입 안전성 + 개발자 경험] |
| **Framework** | $FRONTEND_FRAMEWORK | [컴포넌트 재사용성] |
| **State Management** | $STATE_MANAGER | [예측 가능한 상태 관리] |
| **UI Library** | $UI_LIBRARY | [디자인 시스템 일관성] |

### Infrastructure & DevOps
| 기술 영역 | 선택 기술 | 선정 근거 |
|----------|-----------|----------|
| **Cloud** | $CLOUD_PROVIDER | [GitFlow 자동화 지원] |
| **Container** | Docker + K8s | [배포 일관성 + 스케일링] |
| **CI/CD** | GitHub Actions | [코드와 파이프라인 통합] |
| **Monitoring** | $MONITORING_STACK | [관찰가능성 완전 구현] |

## 품질 속성 달성 전략

### 가용성 (Availability)
- 99.9% 목표 SLA
- 헬스체크 엔드포인트
- 자동 복구 메커니즘

### 성능 (Performance)
- 응답시간 < 200ms
- 처리량 > 1000 TPS
- 동시 사용자 > 10,000명

### 보안 (Security)
- OWASP Top 10 준수
- 정기 보안 감사
- 침입 탐지 시스템

### 유지보수성 (Maintainability)
- 코드 커버리지 > 80%
- 순환 복잡도 < 10
- 기술 부채 정기 해소

---

**@TAG 연결**
- @VISION:CORE → 핵심 비전 구현
- @STRUCT:ARCH → 아키텍처 구조 정의
- @TECH:STACK → 기술 스택 선정 및 근거
- @DESIGN:SECURITY → 보안 설계
- @DESIGN:PERFORMANCE → 성능 설계

**MoAI-ADK 0.2.1 Constitution 검토**:
- [ ] **단순성**: 3개 이하 주요 컴포넌트 원칙 준수
- [ ] **아키텍처**: 모든 기능이 라이브러리로 분리 가능
- [ ] **테스트**: TDD 기반 아키텍처 설계
- [ ] **관찰가능성**: 구조화된 로깅 및 모니터링 계획
- [ ] **버전관리**: GitFlow 투명성 지원 구조

**GitFlow 워크플로우 준비**:
- [ ] `/moai:1-spec` - 이 구조 문서 작성 완료
- [ ] `/moai:2-build` - TDD 기반 구현 준비
- [ ] `/moai:3-sync` - 문서 동기화 및 PR 자동화