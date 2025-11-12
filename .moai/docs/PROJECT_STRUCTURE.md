# MoAI-ADK 프로젝트 구조 및 Skills 카테고리화

> **Release Version**: v0.23.0
> **Documentation Date**: 2025-11-12
> **Skills Count**: 125+ Enterprise-Grade Skills

---

## 🗂️ 프로젝트 디렉토리 구조

### 루트 레벨 (Project Root)

```
MoAI-ADK/
├── README.md                    # 프로젝트 개요 및 시작 가이드
├── README.ko.md                 # 한국어 README
├── CHANGELOG.md                 # 버전 히스토리 및 릴리즈 노트
├── CONTRIBUTING.md              # 기여 가이드
├── LICENSE                      # MIT 라이선스
├── CLAUDE.md                    # Claude Code 지침 (v0.23.0)
├── pyproject.toml               # Python 프로젝트 설정
├── uv.lock                      # UV 의존성 잠금 파일
├── .gitignore                   # Git 무시 패턴
├── .editorconfig                # 에디터 설정
│
├── src/                         # 소스 코드
│   └── moai_adk/
│       ├── __init__.py
│       ├── cli.py               # CLI 엔트리포인트
│       ├── core/                # 핵심 기능
│       ├── templates/           # 패키지 템플릿
│       │   ├── .claude/         # 템플릿 에이전트/스킬
│       │   ├── .moai/           # 템플릿 설정
│       │   ├── CLAUDE.md        # 템플릿 지침
│       │   └── pyproject.toml   # 템플릿 프로젝트 설정
│       └── statusline/          # Claude Code 상태바
│
├── tests/                       # 테스트 코드
│   ├── unit/
│   ├── integration/
│   └── e2e/
│
├── docs/                        # 외부 문서
│   ├── getting-started.md
│   ├── api-reference.md
│   └── examples/
│
├── .moai/                       # MoAI 메타데이터 및 구성
│   ├── config/
│   │   └── config.json          # 프로젝트 설정
│   ├── specs/                   # SPEC 문서
│   │   ├── SPEC-001/
│   │   ├── SPEC-002/
│   │   └── ...
│   ├── docs/                    # 동기화된 문서
│   │   ├── architecture.md
│   │   ├── api-guide.md
│   │   └── PROJECT_STRUCTURE.md (현재 파일)
│   ├── reports/                 # 생성된 보고서
│   │   ├── sync/
│   │   ├── analysis/
│   │   ├── validation/
│   │   └── inspection/
│   ├── logs/                    # 로그 파일
│   ├── temp/                    # 임시 파일
│   ├── cache/                   # 캐시 데이터
│   ├── backups/                 # 백업 파일
│   ├── memory/                  # 세션 메모리
│   ├── research/                # 조사 및 분석
│   └── scripts/                 # 유틸리티 스크립트
│
└── .claude/                     # Alfred 에이전트 및 스킬 (프로젝트 특화)
    ├── agents/                  # AI 에이전트 정의
    ├── commands/                # Alfred 커맨드
    ├── skills/                  # Claude Skills (125+ skills)
    ├── hooks/                   # 시스템 훅
    └── mcp.json                 # MCP 서버 설정
```

---

## 📚 .claude/skills/ 구조 - 125+ Skills 카테고리화

### 1. 재단 Skills (Foundation) - 12 Skills

**SPEC 및 요구사항 관리**:
- `moai-foundation-specs` - SPEC 문서 작성 및 관리
- `moai-foundation-spec-validation` - SPEC 검증 및 품질 기준
- `moai-foundation-ears-format` - EARS 형식 및 패턴

**테스트 및 품질**:
- `moai-foundation-tdd` - TDD 원칙 및 패턴
- `moai-foundation-testing-strategy` - 테스트 전략
- `moai-foundation-quality-gates` - 품질 게이트 및 TRUST 5

**TAGs 및 추적**:
- `moai-foundation-tags` - TAG 시스템 및 관리
- `moai-foundation-traceability` - 완전한 추적성 구현

**개발 원칙**:
- `moai-foundation-trust` - TRUST 5 원칙
- `moai-foundation-best-practices` - 개발 최고 실행
- `moai-foundation-git-workflow` - Git 워크플로우

**기타 재단**:
- `moai-foundation-terminology` - MoAI-ADK 용어

### 2. Alfred 에이전트 Skills - 19 Skills

**핵심 에이전트**:
- `moai-alfred-agent-guide` - Alfred 에이전트 선택 및 위임
- `moai-alfred-personas` - Alfred 대응 스타일
- `moai-alfred-context-budget` - 컨텍스트 관리

**워크플로우 및 조율**:
- `moai-alfred-workflow` - 4단계 개발 워크플로우
- `moai-alfred-command-helpers` - 커맨드 헬퍼 함수
- `moai-alfred-git-workflow` - Git 워크플로우 통합

**개발 전문가**:
- `moai-alfred-spec-expert` - SPEC 작성 전문가
- `moai-alfred-tdd-expert` - TDD 구현 전문가
- `moai-alfred-test-expert` - 테스트 엔지니어
- `moai-alfred-doc-syncer` - 문서 동기화 전문가
- `moai-alfred-security-expert` - 보안 전문가

**기타 전문가**:
- `moai-alfred-backend-expert` - 백엔드 아키텍처
- `moai-alfred-frontend-expert` - 프론트엔드 설계
- `moai-alfred-database-expert` - 데이터베이스 최적화
- `moai-alfred-devops-expert` - DevOps 및 배포
- `moai-alfred-code-reviewer` - 코드 리뷰
- `moai-alfred-plan-agent` - 계획 및 분석
- `moai-alfred-qa-validator` - 품질 검증

**상호작용 및 도구**:
- `moai-alfred-ask-user-questions` - 사용자 질문 도구
- `moai-alfred-report-generator` - 보고서 생성
- `moai-alfred-document-management` - 문서 관리

### 3. 필수 Skills (Essentials) - 10 Skills

**테스트 및 디버깅**:
- `moai-essentials-testing` - 전반적인 테스트 전략
- `moai-essentials-debugging` - 디버깅 기법 및 도구
- `moai-essentials-mock-testing` - Mock 및 Stub 테스트

**성능 및 최적화**:
- `moai-essentials-performance` - 성능 최적화 기법
- `moai-essentials-caching` - 캐싱 전략
- `moai-essentials-optimization` - 코드 최적화

**보안 및 모니터링**:
- `moai-essentials-security` - 기본 보안 원칙
- `moai-essentials-monitoring` - 모니터링 및 로깅
- `moai-essentials-error-handling` - 오류 처리
- `moai-essentials-logging` - 로깅 전략

### 4. 도메인 Skills (Domain) - 35+ Skills

#### 백엔드 아키텍처 (8 skills)
- `moai-domain-api-design` - RESTful API 설계
- `moai-domain-graphql` - GraphQL 설계 및 구현
- `moai-domain-microservices` - 마이크로서비스 아키텍처
- `moai-domain-serverless` - 서버리스 아키텍처
- `moai-domain-event-driven` - 이벤트 기반 아키텍처
- `moai-domain-service-patterns` - 서비스 설계 패턴
- `moai-domain-caching-strategy` - 캐싱 전략
- `moai-domain-performance-optimization` - 성능 최적화

#### 프론트엔드 개발 (10+ skills)
- `moai-domain-html-css` - HTML/CSS 기초
- `moai-domain-tailwind-css` - Tailwind CSS
- `moai-domain-react` - React 프레임워크
- `moai-domain-react-patterns` - React 설계 패턴
- `moai-domain-vue` - Vue.js 프레임워크
- `moai-domain-angular` - Angular 프레임워크
- `moai-domain-shadcn-ui` - shadcn/ui 컴포넌트
- `moai-domain-next-js` - Next.js 프레임워크
- `moai-domain-svelte` - Svelte 프레임워크
- `moai-domain-web-components` - Web Components

#### 데이터베이스 (10+ skills)
- `moai-domain-sql-optimization` - SQL 최적화
- `moai-domain-postgresql` - PostgreSQL 고급
- `moai-domain-mysql` - MySQL 데이터베이스
- `moai-domain-mongodb` - MongoDB NoSQL
- `moai-domain-redis` - Redis 캐싱
- `moai-domain-elasticsearch` - Elasticsearch 검색
- `moai-domain-database-design` - 데이터베이스 설계
- `moai-domain-data-migration` - 데이터 마이그레이션
- `moai-domain-database-backup` - 백업 및 복구
- `moai-domain-query-optimization` - 쿼리 최적화

#### DevOps 및 배포 (10+ skills)
- `moai-domain-docker` - Docker 컨테이너화
- `moai-domain-kubernetes` - Kubernetes 오케스트레이션
- `moai-domain-ci-cd` - CI/CD 파이프라인
- `moai-domain-github-actions` - GitHub Actions
- `moai-domain-git-workflow` - Git 워크플로우
- `moai-domain-infrastructure-as-code` - IaC (Terraform)
- `moai-domain-monitoring` - 모니터링 및 알림
- `moai-domain-logging` - 로깅 시스템
- `moai-domain-security-scanning` - 보안 스캔
- `moai-domain-deployment-strategies` - 배포 전략

### 5. 언어별 Skills (Language) - 20+ Skills

**Python 스택**:
- `moai-lang-python` - Python 기초 및 패턴
- `moai-lang-python-django` - Django 프레임워크
- `moai-lang-python-fastapi` - FastAPI 프레임워크
- `moai-lang-python-testing` - Python 테스트

**TypeScript/JavaScript 스택**:
- `moai-lang-typescript` - TypeScript 기초
- `moai-lang-javascript` - JavaScript 기초
- `moai-lang-node-js` - Node.js 백엔드
- `moai-lang-express` - Express.js 프레임워크

**Go 스택**:
- `moai-lang-go` - Go 언어 기초
- `moai-lang-go-gin` - Gin 프레임워크

**Rust 스택**:
- `moai-lang-rust` - Rust 기초
- `moai-lang-rust-tokio` - Tokio 비동기

**기타 언어**:
- `moai-lang-java` - Java 프로그래밍
- `moai-lang-kotlin` - Kotlin 언어
- `moai-lang-csharp` - C# 및 .NET
- `moai-lang-php` - PHP 백엔드
- `moai-lang-ruby` - Ruby 프로그래밍
- `moai-lang-sql` - SQL 쿼리

### 6. BaaS 플랫폼 Skills - 12 Skills

**클라우드 플랫폼**:
- `moai-baas-foundation` - BaaS 아키텍처 패턴
- `moai-baas-supabase` - Supabase PostgreSQL+Auth
- `moai-baas-firebase` - Firebase 플랫폼
- `moai-baas-vercel` - Vercel 엣지 컴퓨팅
- `moai-baas-cloudflare` - Cloudflare Workers
- `moai-baas-auth0` - Auth0 인증

**확장 플랫폼**:
- `moai-baas-convex` - Convex 백엔드
- `moai-baas-railway` - Railway 배포
- `moai-baas-neon` - Neon PostgreSQL
- `moai-baas-clerk` - Clerk 사용자 관리
- `moai-baas-mongodb-atlas` - MongoDB Atlas
- `moai-baas-aws` - AWS 서비스

### 7. 보안 및 규정준수 Skills - 10 Skills

**인증 및 권한**:
- `moai-security-oauth2` - OAuth 2.0 프로토콜
- `moai-security-saml` - SAML 엔터프라이즈 인증
- `moai-security-webauthn` - WebAuthn 생체인증

**암호화 및 데이터 보호**:
- `moai-security-encryption` - 암호화 기법
- `moai-security-data-protection` - 데이터 보호 원칙
- `moai-security-tls-ssl` - TLS/SSL 설정

**규정준수 및 감사**:
- `moai-security-owasp` - OWASP 상위 10
- `moai-security-compliance` - 규정준수 프레임워크
- `moai-security-vulnerability-assessment` - 취약점 평가
- `moai-security-penetration-testing` - 침투 테스트

### 8. 엔터프라이즈 통합 Skills - 15 Skills

**마이크로서비스**:
- `moai-enterprise-microservices-patterns` - 마이크로서비스 패턴
- `moai-enterprise-service-mesh` - 서비스 메시
- `moai-enterprise-api-gateway` - API 게이트웨이

**이벤트 기반 아키텍처**:
- `moai-enterprise-event-driven` - 이벤트 기반 설계
- `moai-enterprise-kafka` - Apache Kafka
- `moai-enterprise-message-queue` - 메시지 큐

**도메인 주도 설계**:
- `moai-enterprise-ddd` - 도메인 주도 설계
- `moai-enterprise-cqrs` - CQRS 패턴
- `moai-enterprise-event-sourcing` - 이벤트 소싱

**워크플로우 및 오케스트레이션**:
- `moai-enterprise-workflow` - 워크플로우 오케스트레이션
- `moai-enterprise-orchestration` - 서비스 오케스트레이션
- `moai-enterprise-saga` - Saga 패턴

**통합 패턴**:
- `moai-enterprise-integration-patterns` - 통합 패턴
- `moai-enterprise-api-integration` - API 통합
- `moai-enterprise-data-sync` - 데이터 동기화

### 9. 고급 DevOps Skills - 12 Skills

**컨테이너 및 오케스트레이션**:
- `moai-devops-kubernetes-advanced` - Kubernetes 고급
- `moai-devops-docker-compose` - Docker Compose
- `moai-devops-container-registry` - 컨테이너 레지스트리

**배포 전략**:
- `moai-devops-blue-green-deployment` - 블루-그린 배포
- `moai-devops-canary-deployment` - 카나리 배포
- `moai-devops-rolling-deployment` - 롤링 배포

**Infrastructure as Code**:
- `moai-devops-terraform` - Terraform IaC
- `moai-devops-ansible` - Ansible 자동화
- `moai-devops-cloudformation` - CloudFormation

**모니터링 및 관찰성**:
- `moai-devops-prometheus` - Prometheus 모니터링
- `moai-devops-grafana` - Grafana 대시보드
- `moai-devops-observability` - 관찰성 패턴

### 10. 데이터 및 분석 Skills - 18 Skills

**데이터 파이프라인**:
- `moai-data-pipeline-architecture` - 데이터 파이프라인
- `moai-data-etl-design` - ETL 설계
- `moai-data-batch-processing` - 배치 처리

**스트리밍 및 실시간**:
- `moai-data-stream-processing` - 스트림 처리
- `moai-data-kafka-streaming` - Kafka 스트리밍
- `moai-data-real-time-analytics` - 실시간 분석

**데이터 웨어하우스**:
- `moai-data-warehouse-design` - 데이터 웨어하우스
- `moai-data-data-lake` - 데이터 레이크
- `moai-data-columnar-storage` - 컬럼 기반 저장소

**머신러닝 운영**:
- `moai-data-mlops` - MLOps 파이프라인
- `moai-data-model-serving` - 모델 서빙
- `moai-data-feature-engineering` - 피처 엔지니어링

**분석 및 시각화**:
- `moai-data-analytics` - 고급 분석
- `moai-data-visualization` - 데이터 시각화
- `moai-data-bi-tools` - BI 도구 활용

**데이터 거버넌스**:
- `moai-data-data-governance` - 데이터 거버넌스
- `moai-data-data-quality` - 데이터 품질
- `moai-data-privacy-compliance` - 프라이버시 규정

### 11. MCP 및 고급 통합 Skills - 8+ Skills

**MCP 개발**:
- `moai-mcp-builder` - MCP 서버 개발
- `moai-mcp-context7` - Context7 통합
- `moai-mcp-playwright` - Playwright 테스트 자동화

**문서 처리**:
- `moai-document-processing` - 문서 처리 (DOCX, PDF, PPTX, XLSX)

**Artifact 빌더**:
- `moai-artifacts-builder` - React/Tailwind/shadcn/ui 컴포넌트

**엔터프라이즈 커뮤니케이션**:
- `moai-internal-comms` - 내부 커뮤니케이션 자동화

**기타 고급 통합**:
- `moai-sequential-thinking` - 단계별 사고 프로토콜
- `moai-context-manager` - 컨텍스트 관리 고급

### 12. 추가 전문화 Skills - 30+ Skills

**아이콘 및 UI**:
- `moai-icons-lucide` - Lucide 아이콘 (1,200+)
- `moai-icons-react-icons` - React Icons (4,000+)
- `moai-icons-tabler` - Tabler 아이콘 (5,000+)
- `moai-icons-phosphor` - Phosphor 아이콘 (7,500+)
- 및 기타 10+ 아이콘 라이브러리

**모바일 개발**:
- `moai-mobile-react-native` - React Native
- `moai-mobile-flutter` - Flutter 프레임워크
- `moai-mobile-swift` - Swift iOS 개발
- `moai-mobile-kotlin` - Kotlin Android 개발

**클라우드 플랫폼 심화**:
- `moai-cloud-aws-advanced` - AWS 심화
- `moai-cloud-gcp` - Google Cloud Platform
- `moai-cloud-azure` - Microsoft Azure
- `moai-cloud-multi-cloud` - 멀티 클라우드

---

## 📊 Skills 분포도

```
Total Skills: 125+

재단 Skills (9.6%)          : 12 skills
Alfred 에이전트 (15.2%)    : 19 skills
필수 Skills (8%)           : 10 skills
도메인 Skills (28%)        : 35+ skills
언어별 Skills (16%)        : 20+ skills
BaaS 플랫폼 (9.6%)        : 12 skills
보안 및 규정 (8%)          : 10 skills
엔터프라이즈 (12%)         : 15 skills
DevOps (9.6%)             : 12 skills
데이터 분석 (14.4%)       : 18 skills
MCP 및 고급 (6.4%)        : 8+ skills
추가 전문화 (24%)         : 30+ skills
```

---

## 🎯 기술 스택 커버리지

### 프로그래밍 언어 (18)
Python, TypeScript, JavaScript, Go, Rust, Java, Kotlin, Swift, Dart, PHP, Ruby, C, C++, C#, Scala, R, SQL, Shell

### 웹 프레임워크 (15+)
Django, FastAPI, Express, Next.js, React, Vue, Angular, Svelte, Nuxt, NestJS, Gin, Echo, Spring Boot, Laravel, Rails

### 클라우드 플랫폼 (11)
Supabase, Firebase, Vercel, Cloudflare, Auth0, Convex, Railway, Neon, Clerk, AWS, GCP, Azure

### 데이터베이스 (10+)
PostgreSQL, MySQL, MongoDB, Redis, Elasticsearch, Cassandra, DynamoDB, Firestore, Cosmos DB, Snowflake

### DevOps 도구 (20+)
Docker, Kubernetes, GitHub Actions, GitLab CI, Terraform, Ansible, CloudFormation, Prometheus, Grafana, ELK Stack

### 개발 도구 (30+)
VS Code, Git, npm, yarn, pnpm, uv, Make, Docker, Postman, Insomnia, Jest, Pytest, RSpec, Cypress, Playwright

---

## 🔗 TAG 시스템 통합

### TAG 카테고리별 Skills 매핑

- `moai-foundation-specs` - SPEC 작성
- `moai-foundation-spec-validation` - 검증

- `moai-domain-api-design` - API 설계
- `moai-domain-database-design` - DB 설계
- 및 15+ 설계 관련 skills

- `moai-lang-*` - 언어별 구현
- `moai-domain-*` - 도메인별 구현

- `moai-foundation-testing-strategy` - 테스트 전략
- `moai-essentials-testing` - 테스트 기법
- `moai-essentials-mock-testing` - Mock 테스트

- `moai-alfred-document-management` - 문서 관리
- `moai-alfred-doc-syncer` - 문서 동기화

---

## 📈 버전 히스토리

| 버전 | 날짜 | Skills | 주요 변화 |
|------|------|--------|----------|
| v0.22.5 | 2025-10-15 | 16 | 초기 기본 Skills |
| v0.23.0 | 2025-11-12 | 125+ | Phase 1 Batch 2 완료 (109+ 추가) |

---

## 🚀 빠른 참조

### 특정 주제에 따른 Skill 찾기

**React로 시작하려면?**
```
moai-domain-react → 기본 패턴
moai-domain-react-patterns → 고급 패턴
moai-lang-typescript → 타입 안전성
```

**API 설계 시?**
```
moai-domain-api-design → RESTful
moai-domain-graphql → GraphQL
moai-domain-microservices → 아키텍처
```

**배포 자동화?**
```
moai-domain-docker → 컨테이너화
moai-domain-kubernetes → 오케스트레이션
moai-devops-ci-cd → CI/CD 파이프라인
moai-domain-github-actions → GitHub 자동화
```

**데이터 처리?**
```
moai-data-pipeline-architecture → 설계
moai-data-stream-processing → 실시간 처리
moai-data-warehouse-design → 데이터 웨어하우스
```

---

## ✅ 품질 보증

모든 125+ Skills은 다음을 충족합니다:

- **TRUST 5 원칙**: Test-First, Readable, Unified, Secured, Trackable
- **문서화**: 완전한 예제 및 참조 문서
- **코드 예제**: 200+ 실행 가능한 예제
- **최신성**: 2025년 안정적인 기술 스택
- **엔터프라이즈급**: 프로덕션 준비 완료

---

**📌 마지막 업데이트**: 2025-11-12 (v0.23.0 Release)
**문서 유지보수**: doc-syncer 에이전트
