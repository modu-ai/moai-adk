# @SPEC:DOCS-001: VitePress 문서 사이트 구축 - 구현 계획

## 우선순위 기반 구현 계획

이 문서는 SPEC-DOCS-001의 구현 계획을 우선순위 기반으로 정의합니다.

---

## Phase 1: 핵심 경로 (P0 - 최우선)

**목표**: 사용자가 MoAI-ADK를 처음 접하고 시작하는 데 필요한 최소한의 문서 제공

**범위**: 5개 페이지

### 1.1 홈페이지
- **파일**: `docs/index.md`
- **콘텐츠 소스**: README.md 1-85줄 (Alfred 소개, 핵심 가치)
- **구성**:
  - Hero 섹션: 타이틀, 서브타이틀, CTA 버튼
  - Features 섹션: 4가지 핵심 가치 (일관성, 품질, 추적성, 범용성)
  - Quick Links: Getting Started, Concepts, Examples
- **우선순위**: P0 - 최우선

### 1.2 Getting Started
- **파일**: `docs/guide/getting-started.md`
- **콘텐츠 소스**: README.md 87-156줄 (Quick Start)
- **구성**:
  - 준비물 (Bun, Claude Code, Git)
  - 3단계 설치 및 초기화
  - 첫 기능 개발 예제 (3단계 워크플로우)
- **우선순위**: P0

### 1.3 What is MoAI-ADK
- **파일**: `docs/guide/what-is-moai-adk.md`
- **콘텐츠 소스**: README.md 183-267줄 (The Problem, The Solution)
- **구성**:
  - 바이브 코딩의 한계 (5가지 문제)
  - MoAI-ADK의 3단계 워크플로우
  - Alfred SuperAgent 역할
- **우선순위**: P0

### 1.4 SPEC-First TDD
- **파일**: `docs/concepts/spec-first-tdd.md`
- **콘텐츠 소스**: development-guide.md 9-23줄 + README.md 관련 부분
- **구성**:
  - SPEC 우선 TDD 워크플로우
  - 3단계 개발 루프 (SPEC → TDD → Sync)
  - EARS 요구사항 작성법
- **우선순위**: P0

### 1.5 FAQ
- **파일**: `docs/guide/faq.md`
- **콘텐츠 소스**: README.md FAQ 섹션 (신규 작성 필요)
- **구성**:
  - 자주 묻는 질문 10개
  - Q: MoAI-ADK는 무료인가요?
  - Q: 어떤 언어를 지원하나요?
  - Q: Team 모드와 Personal 모드 차이는?
  - Q: 기존 프로젝트에 적용 가능한가요?
- **우선순위**: P0

**Phase 1 완료 기준**:
- 5개 페이지 작성 완료
- VitePress 빌드 성공
- 로컬 개발 서버 정상 동작 (localhost:5173)

---

## Phase 2: 학습 경로 (P1 - 높음)

**목표**: 사용자가 MoAI-ADK의 핵심 개념과 사용법을 이해하도록 지원

**범위**: 11개 페이지 (누적 16개)

### 2.1 핵심 개념 (5개 페이지)
- `docs/concepts/alfred-superagent.md` - Alfred의 역할과 10개 에이전트 생태계
- `docs/concepts/trust-principles.md` - TRUST 5원칙 (Test, Readable, Unified, Secured, Trackable)
- `docs/concepts/tag-system.md` - @TAG 시스템과 추적성
- `docs/concepts/ears-syntax.md` - EARS 요구사항 작성 방법론
- `docs/concepts/gitflow-integration.md` - GitFlow 워크플로우와 자동화

**콘텐츠 소스**:
- development-guide.md 핵심 개념 섹션
- README.md Alfred 소개 및 TRUST 원칙

### 2.2 설치 가이드 (6개 페이지)
- `docs/installation/prerequisites.md` - 사전 준비사항 (Bun, Git, Claude Code)
- `docs/installation/init-new-project.md` - 새 프로젝트 초기화
- `docs/installation/init-existing-project.md` - 기존 프로젝트 통합
- `docs/installation/init-options.md` - `moai init` 옵션 설명 (SPEC-INSTALL-001 참조)
- `docs/installation/non-interactive.md` - 비대화형 초기화 (SPEC-INIT-001 참조)
- `docs/installation/windows-wsl.md` - Windows/WSL 환경 설정 (SPEC-INIT-002 참조)

**콘텐츠 소스**:
- README.md Quick Start
- SPEC-INSTALL-001, SPEC-INIT-001, SPEC-INIT-002

**Phase 2 완료 기준**:
- 11개 페이지 추가 작성 (누적 16개)
- Sidebar 구성 완료 (guide, concepts, installation)
- 내부 링크 유효성 검증

---

## Phase 3: 실전 활용 (P1 - 높음)

**목표**: 사용자가 실제 프로젝트에서 MoAI-ADK를 활용하도록 실전 가이드 제공

**범위**: 29개 페이지 (누적 45개)

### 3.1 CLI 레퍼런스 (10개 페이지)
- `docs/cli-reference/init.md` - 프로젝트 초기화
- `docs/cli-reference/doctor.md` - 시스템 진단
- `docs/cli-reference/status.md` - 프로젝트 상태 확인
- `docs/cli-reference/update.md` - 도구 업데이트
- `docs/cli-reference/restore.md` - 프로젝트 복원
- `docs/cli-reference/help.md` - 도움말 표시
- `docs/cli-reference/version.md` - 버전 확인
- `docs/cli-reference/alfred-1-spec.md` - /alfred:1-plan 명령어
- `docs/cli-reference/alfred-2-build.md` - /alfred:2-run 명령어
- `docs/cli-reference/alfred-3-sync.md` - /alfred:3-sync 명령어

**콘텐츠 소스**:
- development-guide.md CLI 명령어 섹션
- README.md CLI Reference
- 기존 CLI 도움말 출력

### 3.2 언어별 가이드 (10개 페이지)
- `docs/language-guides/typescript.md` - TypeScript 프로젝트
- `docs/language-guides/python.md` - Python 프로젝트
- `docs/language-guides/java.md` - Java 프로젝트
- `docs/language-guides/go.md` - Go 프로젝트
- `docs/language-guides/rust.md` - Rust 프로젝트
- `docs/language-guides/dart-flutter.md` - Dart/Flutter 모바일
- `docs/language-guides/swift-ios.md` - Swift/iOS
- `docs/language-guides/kotlin-android.md` - Kotlin/Android
- `docs/language-guides/react-native.md` - React Native
- `docs/language-guides/csharp.md` - C#/.NET

**콘텐츠 소스**:
- development-guide.md 언어별 규칙
- 신규 작성 (언어별 도구 체인, 예제 코드)

### 3.3 실전 예제 (4개 페이지)
- `docs/examples/jwt-auth.md` - JWT 인증 구현 예제
- `docs/examples/rest-api.md` - REST API 서버 예제
- `docs/examples/cli-tool.md` - CLI 도구 개발 예제
- `docs/examples/mobile-app.md` - 모바일 앱 예제 (Flutter)

**콘텐츠 소스**:
- 기존 SPEC 문서 (AUTH-001 등)
- 실제 구현 코드에서 추출

### 3.4 레퍼런스 (5개 페이지)
- `docs/reference/CONFIG-SCHEMA.md` - .moai/config.json 스키마 (SPEC-CONFIG-001 참조)
- `docs/reference/spec-template.md` - SPEC 템플릿
- `docs/reference/tag-reference.md` - @TAG 레퍼런스
- `docs/reference/trust-checklist.md` - TRUST 체크리스트
- `docs/reference/agent-api.md` - 에이전트 API (온디맨드 호출)

**콘텐츠 소스**:
- development-guide.md 레퍼런스 섹션
- SPEC-CONFIG-001

**Phase 3 완료 기준**:
- 29개 페이지 추가 작성 (누적 45개)
- 모든 CLI 명령어 문서화
- 언어별 가이드 10개 완료
- 실전 예제 4개 완료

---

## Phase 4: 심화 및 기여 (P2 - 중간)

**목표**: 고급 사용자와 기여자를 위한 심화 주제 제공

**범위**: 13개 페이지 (누적 58개, 여유분 5개 포함)

### 4.1 문제 해결 (4개 페이지)
- `docs/troubleshooting/installation-errors.md` - 설치 오류 해결
- `docs/troubleshooting/build-errors.md` - 빌드 에러 해결
- `docs/troubleshooting/git-issues.md` - Git 관련 문제
- `docs/troubleshooting/common-errors.md` - 자주 발생하는 에러

**콘텐츠 소스**:
- README.md 문제 해결 섹션
- GitHub Issues 자주 묻는 질문

### 4.2 고급 주제 (4개 페이지)
- `docs/advanced/context-engineering.md` - 컨텍스트 엔지니어링
- `docs/advanced/custom-agents.md` - 커스텀 에이전트 작성
- `docs/advanced/multi-repo.md` - 멀티 리포지토리 관리
- `docs/advanced/ci-cd-integration.md` - CI/CD 통합

**콘텐츠 소스**:
- development-guide.md Context Engineering 섹션
- 신규 작성 (고급 활용법)

### 4.3 기여하기 (5개 페이지)
- `docs/contributing/index.md` - 기여 가이드 개요
- `docs/contributing/code-contribution.md` - 코드 기여 방법
- `docs/contributing/documentation.md` - 문서 기여 방법
- `docs/contributing/bug-report.md` - 버그 리포트 작성
- `docs/contributing/feature-request.md` - 기능 요청 작성

**콘텐츠 소스**:
- CONTRIBUTING.md (신규 작성 예정)
- GitHub 기여 가이드 템플릿

**Phase 4 완료 기준**:
- 13개 페이지 추가 작성 (누적 58개)
- 모든 섹션 완료 (8개 디렉토리)
- 검색 기능 정상 동작
- 콘텐츠 소스 비율 검증 (README 30%, dev-guide 60%, 신규 10%)

---

## VitePress 설정 및 커스터마이징

### config.ts 설정
- **파일**: `docs/.vitepress/config.ts`
- **설정 항목**:
  - 사이트 메타데이터 (title, description)
  - Sidebar 구성 (8개 섹션)
  - Navbar 메뉴
  - 검색 기능 (local search provider)
  - Dark Mode (기본 활성화)
  - 소셜 링크 (GitHub, Discord)

### 테마 커스터마이징
- **로고**: Alfred 로고 추가 (docs/public/alfred_logo.png)
- **색상**: MoAI-ADK 브랜딩 컬러 적용
- **폰트**: Pretendard (한글 최적화)

### 빌드 스크립트
- `bun run docs:dev` - 개발 서버 (핫 리로드)
- `bun run docs:build` - 프로덕션 빌드
- `bun run docs:preview` - 프로덕션 미리보기

---

## 콘텐츠 마이그레이션 전략

### README.md 활용 (30%)
- **총 1097줄 중 330줄 활용**
- **주요 섹션**:
  - Alfred 소개 (1-85줄) → index.md, guide/what-is-moai-adk.md
  - Quick Start (87-156줄) → guide/getting-started.md
  - The Problem (183-267줄) → guide/what-is-moai-adk.md
  - CLI Reference → cli-reference/*.md
  - FAQ → guide/faq.md

### development-guide.md 활용 (60%)
- **총 391줄 중 235줄 활용**
- **주요 섹션**:
  - SPEC 우선 TDD (9-23줄) → concepts/spec-first-tdd.md
  - EARS 작성법 (25-55줄) → concepts/ears-syntax.md
  - Context Engineering (58-100줄) → advanced/context-engineering.md
  - TRUST 원칙 → concepts/trust-principles.md
  - @TAG 시스템 → concepts/tag-system.md

### 신규 작성 (10%)
- 언어별 가이드 (10개 페이지)
- 실전 예제 (4개 페이지)
- 기여 가이드 (5개 페이지)

---

## 품질 보증 계획

### 빌드 검증
- VitePress 빌드 성공 확인 (`bun run docs:build`)
- 빌드 에러 0개
- 경고 최소화

### 링크 검증
- 모든 내부 링크 유효성 확인
- 외부 링크 응답 확인 (선택적)
- 깨진 링크 0개

### 콘텐츠 검증
- 맞춤법 검사 (한글, 영문)
- 코드 예제 실행 가능성 확인
- 스크린샷/이미지 최적화 (WebP 변환)

### 성능 검증
- 페이지 로딩 시간 < 2초
- 검색 속도 < 500ms
- 빌드 시간 < 30초

---

## 배포 준비 (Phase 4 완료 후)

### GitHub Pages 배포
- `.github/workflows/deploy.yml` 작성
- `gh-pages` 브랜치 자동 배포
- 커스텀 도메인 설정 (선택적)

### Vercel 배포 (대안)
- Vercel 프로젝트 연결
- 자동 빌드 및 배포
- 프리뷰 환경 제공

---

## 다음 단계 안내

SPEC-DOCS-001 작성 완료 후:
1. `/alfred:2-run DOCS-001` - VitePress 설정 및 Phase 1 페이지 작성
2. `/alfred:3-sync` - 문서 동기화 및 TAG 체인 검증
3. 반복 사이클: Phase 2-4 순차 진행

---

**작성일**: 2025-10-06
**버전**: 0.1.0
**관련 SPEC**: @SPEC:DOCS-001
