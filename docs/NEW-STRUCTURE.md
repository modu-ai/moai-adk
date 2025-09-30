# MoAI-ADK 온라인 문서 구조 재설계

> **설계 원칙**: 사용자 여정 기반, SPEC-First TDD 핵심 가치 전달, 실전 중심

---

## 📚 전체 목차 구조

### 🏠 **1. 홈 & 개요** (Hero + Quick Wins)

#### 1.1 메인 페이지 (index.md)
- **Hero Section**
  - 타이틀: "🗿 MoAI-ADK: SPEC 우선 TDD 개발 도구"
  - 서브타이틀: "TypeScript 기반 범용 언어 지원, Claude Code 완전 통합"
  - CTA: "5분 안에 시작하기" → Quick Start

- **핵심 가치 6가지**
  1. 📝 SPEC 우선 개발 (EARS 방법론)
  2. 🧪 테스트 주도 개발 (Red-Green-Refactor)
  3. 🏷️ @TAG 추적성 시스템 (94% 최적화)
  4. 🌍 범용 언어 지원 (TS/Python/Java/Go/Rust)
  5. ⚡ 초고속 성능 (Bun 98%, Vitest 92.9%)
  6. 🤖 Claude Code 완전 통합 (7 agents + 5 commands)

- **빠른 시작 코드 블록**
  ```bash
  # Bun (권장)
  bun add -g moai-adk
  moai init my-project
  cd my-project && claude
  ```

- **통계 대시보드**
  - ✅ 92% TRUST 원칙 달성
  - ⚡ 182ms 빌드 시간
  - 📦 195KB 패키지 크기
  - 🌐 8개 주요 언어 지원

#### 1.2 소개 (introduction.md) - 새 페이지
- **MoAI-ADK란?**
  - 문제: AI 페어 프로그래밍의 체계 부재
  - 해결: SPEC-First TDD 자동화 프레임워크

- **핵심 개념 3가지**
  1. SPEC-First: 명세 없이는 코드 없음
  2. TDD-First: 테스트 없이는 구현 없음
  3. TAG-First: 추적성 없이는 완성 없음

- **아키텍처 다이어그램**
  ```
  MoAI-ADK Architecture
  ├── TypeScript CLI (도구)
  ├── Universal Language Support (모든 언어)
  └── Claude Code Integration (자동화)
  ```

- **전환 전후 비교 (SPEC-013)**
  - Before: Python 15MB, 4.6s build
  - After: TypeScript 195KB, 182ms build (99% 절감)

#### 1.3 주요 특징 (features.md) - 새 페이지
- **1. SPEC-First 개발**
  - EARS 방법론 소개
  - 자동 SPEC 템플릿 생성
  - GitHub Issues 통합 (Team 모드)

- **2. TDD 자동화**
  - Red-Green-Refactor 강제
  - 언어별 자동 도구 선택
  - 92.9% 테스트 성공률

- **3. @TAG 추적성**
  - Primary Chain: @REQ → @DESIGN → @TASK → @TEST
  - 코드 직접 스캔 (94% 최적화)
  - 실시간 무결성 검증

- **4. 범용 언어 지원**
  - TypeScript, Python, Java, Go, Rust, C++, C#, PHP
  - 자동 언어 감지
  - 도구 자동 매핑

- **5. Claude Code 통합**
  - 7개 전문 에이전트
  - 5개 워크플로우 명령어
  - 8개 이벤트 훅

---

###  **2. 시작하기** (Getting Started)

#### 2.1 설치 (installation.md) - 기존 개선
- **시스템 요구사항**
  - Node.js 18+ / Bun 1.2+
  - Git 2.0+
  - (선택) Git LFS, Claude Code

- **설치 방법 3가지**
  ```bash
  # 1. Bun (권장)
  bun add -g moai-adk

  # 2. npm
  npm install -g moai-adk

  # 3. 소스 빌드
  git clone https://github.com/modu-ai/moai-adk.git
  cd moai-adk/moai-adk-ts && bun install && bun run build
  ```

- **설치 확인**
  ```bash
  moai --version
  moai doctor  # 시스템 진단
  ```

- **트러블슈팅**
  - 권한 오류: sudo 또는 nvm 사용
  - TypeScript 오류: global TypeScript 설치
  - Git 오류: Git LFS 설치

#### 2.2 빠른 시작 (quick-start.md) - 기존 개선
- **5분 튜토리얼**
  1. 프로젝트 초기화 (1분)
  2. SPEC 작성 (1분)
  3. TDD 구현 (2분)
  4. 문서 동기화 (1분)

- **첫 번째 기능 구현**
  - 예시: 사용자 인증 API
  - EARS 요구사항 작성
  - Red-Green-Refactor 실습
  - @TAG 체인 생성

- **다음 단계**
  - 3단계 워크플로우 학습
  - SPEC-First TDD 심화
  - TAG 시스템 이해

#### 2.3 프로젝트 초기화 (project-setup.md) - 새 페이지
- **moai init 상세 가이드**
  - Personal 모드 vs Team 모드
  - 프로젝트 구조 설명
  - 템플릿 커스터마이징

- **생성된 파일 구조**
  ```
  my-project/
  ├── .moai/              # MoAI-ADK 설정
  │   ├── config.json
  │   ├── memory/
  │   ├── project/
  │   └── specs/
  ├── .claude/            # Claude Code 통합
  │   ├── agents/
  │   ├── commands/
  │   ├── hooks/
  │   └── settings.json
  └── your-code/          # 실제 프로젝트 코드
  ```

- **초기 설정 완료 체크리스트**
  - [ ] Git 저장소 초기화
  - [ ] Claude Code 연동 확인
  - [ ] moai doctor 진단 통과
  - [ ] 첫 SPEC 작성 준비

---

### 📖 **3. 핵심 개념** (Core Concepts)

#### 3.1 SPEC-First TDD (spec-first-tdd.md) - 기존 확장
- **SPEC-First 철학**
  - 명세가 곧 계약 (Contract)
  - 코드는 SPEC의 구현체
  - 추적성이 품질을 보장

- **EARS 방법론 완전 가이드**
  1. Ubiquitous: 시스템은 [기능]을 제공해야 한다
  2. Event-driven: WHEN [조건]이면, [동작]해야 한다
  3. State-driven: WHILE [상태]일 때, [동작]해야 한다
  4. Optional: WHERE [조건]이면, [동작]할 수 있다
  5. Constraints: IF [조건]이면, [제약]해야 한다

- **EARS 템플릿 예시**
  - 인증 시스템
  - RESTful API
  - 데이터 처리 파이프라인
  - UI 컴포넌트

- **TDD 사이클**
  - Red: 실패하는 테스트 작성
  - Green: 최소한의 코드로 통과
  - Refactor: 품질 개선 (TRUST 원칙)

- **언어별 TDD 구현**
  - TypeScript: Vitest + strict typing
  - Python: pytest + mypy
  - Java: JUnit + Maven
  - Go: go test + table-driven
  - Rust: cargo test + doc tests

#### 3.2 TAG 시스템 (tag-system.md) - 기존 확장
- **@TAG란 무엇인가?**
  - 코드와 요구사항의 연결고리
  - Primary Chain 4단계
  - 실시간 추적성 보장

- **TAG 체계 완전 가이드**
  - **Primary Chain** (필수)
    - @REQ: 요구사항
    - @DESIGN: 설계
    - @TASK: 구현
    - @TEST: 검증

  - **Implementation Tags**
    - @FEATURE, @API, @UI, @DATA

  - **Quality Tags**
    - @PERF, @SEC, @DOCS, @DEBT

  - **Meta Tags**
    - @OPS, @RELEASE, @DEPRECATED

- **TAG BLOCK 템플릿**
  ```typescript
  // @FEATURE:AUTH-001 | Chain: @REQ:AUTH-001 → @DESIGN:AUTH-001 → @TASK:AUTH-001 → @TEST:AUTH-001
  // Related: @SEC:AUTH-001, @DOCS:AUTH-001
  ```

- **TAG 무결성 관리**
  - 중복 방지: `rg "@REQ:키워드"`
  - 고아 TAG 감지
  - 끊어진 체인 복구

- **언어별 TAG 사용법**
  - TypeScript: `//` 주석
  - Python: `#` 주석
  - Java: `//` 또는 `/* */`
  - Go: `//` 주석
  - Rust: `//` 주석

#### 3.3 3단계 워크플로우 (workflow.md) - 기존 개선
- **Stage 0: 프로젝트 준비** (선택)
  - `/moai:0-project` 실행
  - product/structure/tech 문서 생성
  - 프로젝트 비전 수립

- **Stage 1: SPEC 작성**
  - `/moai:1-spec "기능 제목"`
  - EARS 요구사항 작성
  - @TAG Catalog 생성
  - 사용자 확인 후 브랜치 생성

- **Stage 2: TDD 구현**
  - `/moai:2-build SPEC-ID`
  - Red-Green-Refactor 자동화
  - 언어별 도구 자동 선택
  - TRUST 원칙 자동 검증

- **Stage 3: 문서 동기화**
  - `/moai:3-sync`
  - TAG 체인 검증
  - Living Document 업데이트
  - PR 상태 전환

- **온디맨드 디버깅**
  - `@agent-debug-helper "오류내용"`
  - 시스템 진단 자동화
  - 개발 가이드 검사

- **전체 사이클 예시**
  - 실제 프로젝트 완전 구현 예제
  - 각 단계별 상세 커맨드
  - 출력 결과 해석

#### 3.4 TRUST 5원칙 (trust-principles.md) - 새 페이지
- **T - Test First**
  - SPEC 기반 테스트 작성
  - 언어별 테스트 프레임워크
  - 커버리지 85% 이상

- **R - Readable**
  - 요구사항 주도 가독성
  - 함수 ≤50 LOC, 파일 ≤300 LOC
  - SPEC 용어 사용

- **U - Unified**
  - SPEC 기반 복잡도 관리
  - 복잡도 ≤10
  - 언어 간 일관성

- **S - Secured**
  - SPEC 보안 요구사항
  - 입력 검증
  - Winston logger (민감정보 마스킹)

- **T - Trackable**
  - SPEC-코드 추적성
  - @TAG 시스템
  - 3단계 워크플로우 추적

- **TRUST 준수율 측정**
  - 자동 검증 도구
  - 리포트 생성
  - 개선 가이드

---

### 🛠️ **4. CLI 레퍼런스** (CLI Reference)

#### 4.1 moai init (init.md) - 기존 개선
- **기본 사용법**
  ```bash
  moai init [project-name] [options]
  ```

- **옵션 상세**
  - `--team`: Team 모드 (GitHub Issues/PR)
  - `--personal`: Personal 모드 (기본값)
  - `--backup`: 백업 생성
  - `--force`: 강제 덮어쓰기

- **실행 단계**
  1. System Verification (doctor)
  2. Configuration
  3. Installation (5 phases)

- **생성된 구조 설명**

- **트러블슈팅**
  - TypeScript 오류
  - 권한 문제
  - Git 설정

#### 4.2 moai doctor (doctor.md) - 기존 대폭 개선
- **진단 항목 5가지**
  1. Runtime Requirements (Git, Node.js)
  2. Development Requirements (npm, TypeScript)
  3. Optional Requirements (Git LFS)
  4. Language-Specific (자동 감지)
  5. Performance (빌드 시간, 패키지 크기)

- **언어 자동 감지**
  - JavaScript/TypeScript
  - Python
  - Java
  - Go
  - 기타 언어

- **동적 요구사항 추가**
  - 감지된 언어별 도구 자동 추가
  - 실시간 진단 리포트

- **백업 관리**
  ```bash
  moai doctor --list-backups
  ```

#### 4.3 moai status (status.md) - 기존 개선
- **프로젝트 상태 확인**
  - 설정 요약
  - SPEC 진행률
  - TAG 통계

- **상세 모드**
  ```bash
  moai status --verbose
  ```

- **특정 경로 지정**
  ```bash
  moai status --project-path /path/to/project
  ```

#### 4.4 moai update (update.md) - 기존 개선
- **업데이트 전략**
  - 전체 업데이트
  - 패키지만
  - 리소스만

- **버전 확인**
  ```bash
  moai update --check
  ```

- **백업 옵션**
  ```bash
  moai update --no-backup
  ```

#### 4.5 moai restore (restore.md) - 기존 개선
- **백업 복원**
  ```bash
  moai restore <backup-path>
  ```

- **Dry-run 모드**
  ```bash
  moai restore <backup-path> --dry-run
  ```

- **강제 복원**
  ```bash
  moai restore <backup-path> --force
  ```

#### 4.6 moai help (help.md) - 새 페이지
- **전체 명령어 목록**
- **명령어별 상세 도움말**
- **예시 모음**

---

### 🤖 **5. Claude Code 통합** (Claude Code Integration) - 새 섹션

#### 5.1 에이전트 가이드 (agents.md) - 새 페이지
- **7개 전문 에이전트 소개**

1. **spec-builder**
   - 역할: SPEC 작성 전담
   - 사용: EARS 요구사항 생성
   - 브랜치: 사용자 확인 후 생성

2. **code-builder**
   - 역할: TDD 구현 전담
   - 범용 언어 지원
   - Red-Green-Refactor 자동화

3. **doc-syncer**
   - 역할: 문서 동기화 전담
   - TAG 체인 검증
   - Living Document 업데이트

4. **cc-manager**
   - 역할: Claude Code 설정 전담
   - 권한 최적화
   - 설정 표준화

5. **debug-helper**
   - 역할: 온디맨드 디버깅
   - 시스템 진단
   - 개발 가이드 검사

6. **git-manager**
   - 역할: Git 작업 전담
   - 사용자 확인 후 브랜치/머지
   - 커밋 자동화

7. **trust-checker**
   - 역할: 품질 검증 통합
   - TRUST 5원칙 검사
   - 코드 품질 분석

- **에이전트 호출 방법**
  ```
  @agent-{name} "요청 내용"
  ```

#### 5.2 워크플로우 명령어 (commands.md) - 새 페이지
- **5개 슬래시 명령어**

1. `/moai:0-project` (선택)
   - 프로젝트 비전 수립
   - 3대 문서 생성

2. `/moai:1-spec`
   - SPEC 작성
   - EARS 요구사항
   - TAG Catalog

3. `/moai:2-build`
   - TDD 구현
   - 범용 언어 지원
   - TRUST 검증

4. `/moai:3-sync`
   - 문서 동기화
   - TAG 검증
   - PR 상태 전환

5. `/moai:help` (선택)
   - 사용법 가이드
   - 예시 모음

#### 5.3 이벤트 훅 (hooks.md) - 새 페이지
- **8개 이벤트 훅**

1. **file-monitor**: 파일 모니터링
2. **language-detector**: 언어 자동 감지
3. **policy-block**: 정책 차단
4. **pre-write-guard**: 쓰기 전 검증
5. **session-notice**: 세션 시작 알림
6. **steering-guard**: 방향성 가이드
7. **run-tests-and-report**: 테스트 자동 실행
8. **claude-code-monitor**: Claude Code 감시

- **훅 커스터마이징**
  - 설정 방법
  - 예시 코드
  - 트러블슈팅

#### 5.4 Output Styles (output-styles.md) - 새 페이지
- **5개 출력 스타일**

1. **beginner**: 초보자용 (상세 설명)
2. **study**: 학습용 (교육적 인사이트)
3. **pair**: 페어 프로그래밍용
4. **expert**: 전문가용 (간결)
5. **default**: 기본 스타일

- **스타일 선택 방법**
- **커스텀 스타일 생성**

---

### 🌍 **6. 언어별 가이드** (Language Guides) - 새 섹션

#### 6.1 TypeScript (typescript.md) - 새 페이지
- **프로젝트 설정**
  ```bash
  moai init my-ts-project
  ```

- **TDD 도구 체인**
  - Vitest: 테스트 프레임워크
  - Biome: 린터+포맷터
  - tsup: 빌드 도구

- **SPEC 작성 예시**
  - React 컴포넌트
  - Node.js API
  - CLI 도구

- **TAG 사용 예시**
  ```typescript
  // @FEATURE:AUTH-001 | Chain: @REQ:AUTH-001 → @DESIGN:AUTH-001 → @TASK:AUTH-001 → @TEST:AUTH-001
  interface AuthService {
    authenticate(username: string, password: string): Promise<boolean>;
  }
  ```

- **TDD 실습**
  - Red: 실패 테스트
  - Green: 구현
  - Refactor: 리팩토링

- **TRUST 원칙 적용**

#### 6.2 Python (python.md) - 새 페이지
- **프로젝트 설정**
  ```bash
  moai init my-python-project
  ```

- **TDD 도구 체인**
  - pytest: 테스트 프레임워크
  - mypy: 타입 검사
  - black: 포맷터
  - ruff: 린터

- **SPEC 작성 예시**
  - FastAPI 서버
  - 데이터 파이프라인
  - CLI 도구

- **TAG 사용 예시**
  ```python
  # @FEATURE:AUTH-001 | Chain: @REQ:AUTH-001 → @DESIGN:AUTH-001 → @TASK:AUTH-001 → @TEST:AUTH-001
  class AuthService:
      def authenticate(self, username: str, password: str) -> bool:
          pass
  ```

- **TDD 실습**
- **TRUST 원칙 적용**

#### 6.3 Java (java.md) - 새 페이지
- **프로젝트 설정**
- **TDD 도구: JUnit, Maven/Gradle**
- **SPEC 예시: Spring Boot**
- **TAG 사용법**
- **TDD 실습**

#### 6.4 Go (go.md) - 새 페이지
- **프로젝트 설정**
- **TDD 도구: go test, gofmt**
- **SPEC 예시: HTTP 서버**
- **TAG 사용법 (인터페이스)**
- **Table-driven tests**

#### 6.5 Rust (rust.md) - 새 페이지
- **프로젝트 설정**
- **TDD 도구: cargo test, rustfmt**
- **SPEC 예시: CLI 도구**
- **TAG 사용법 (trait)**
- **Doc tests**

#### 6.6 기타 언어 (other-languages.md) - 새 페이지
- **C++**: GoogleTest, CMake
- **C#**: xUnit, MSBuild
- **PHP**: PHPUnit, Composer
- **Ruby**: RSpec, Bundler
- **확장 가능한 구조**

---

### 🎓 **7. 고급 주제** (Advanced Topics) - 새 섹션

#### 7.1 커스텀 에이전트 (custom-agents.md) - 새 페이지
- **에이전트 생성 가이드**
  - 기본 템플릿
  - 프롬프트 엔지니어링
  - 도구 통합

- **예시: 성능 분석 에이전트**
  ```markdown
  # @agent-perf-analyzer

  당신은 성능 분석 전문가입니다.
  ```

- **에이전트 테스트**
- **배포 및 공유**

#### 7.2 CI/CD 통합 (ci-cd.md) - 새 페이지
- **GitHub Actions 설정**
  ```yaml
  name: MoAI-ADK CI
  on: [push, pull_request]
  jobs:
    validate:
      - run: moai doctor
      - run: moai status --verbose
  ```

- **GitLab CI**
- **Jenkins 통합**
- **자동화 전략**

#### 7.3 팀 협업 (team-collaboration.md) - 새 페이지
- **Team 모드 활용**
  - GitHub Issues 통합
  - PR 자동화
  - 리뷰 프로세스

- **코드 리뷰 가이드**
  - SPEC 검증
  - TAG 체인 확인
  - TRUST 원칙 준수

- **문서화 전략**
  - Living Document 유지
  - API 문서 자동 생성
  - 릴리스 노트

#### 7.4 성능 최적화 (performance.md) - 새 페이지
- **빌드 최적화**
  - Bun vs npm 벤치마크
  - tsup 설정 튜닝
  - 캐싱 전략

- **TAG 시스템 최적화**
  - 코드 스캔 성능
  - 인덱싱 전략
  - 메모리 관리

- **대규모 프로젝트 대응**
  - 10,000+ 파일 처리
  - 모노레포 지원
  - 병렬 처리

#### 7.5 보안 가이드 (security.md) - 새 페이지
- **SPEC 보안 요구사항**
- **Winston logger 활용**
  - 민감정보 마스킹 (15개 필드)
  - 구조화 로깅
  - 감사 로그

- **입력 검증**
- **의존성 보안**
- **비밀 관리**

#### 7.6 마이그레이션 (migration.md) - 새 페이지
- **기존 프로젝트 통합**
  - MoAI-ADK 적용 가이드
  - 점진적 도입 전략
  - TAG 추가 방법

- **Python v0.1.x → TypeScript v2.0 마이그레이션**
  - 변경 사항
  - 마이그레이션 스크립트
  - 트러블슈팅

---

### 📚 **8. 레퍼런스** (Reference) - 새 섹션

#### 8.1 설정 파일 (configuration.md) - 새 페이지
- **`.moai/config.json`**
  ```json
  {
    "version": "0.0.1",
    "mode": "personal | team",
    "features": [],
    "backup": { ... },
    "git": { ... }
  }
  ```

- **`.claude/settings.json`**
- **환경 변수**
- **설정 우선순위**

#### 8.2 EARS 템플릿 (ears-templates.md) - 새 페이지
- **인증 시스템**
- **RESTful API**
- **데이터 처리**
- **UI 컴포넌트**
- **마이크로서비스**

#### 8.3 TAG 레퍼런스 (tag-reference.md) - 새 페이지
- **전체 TAG 목록**
  - Primary (4개)
  - Implementation (4개)
  - Quality (4개)
  - Meta (4개)

- **TAG ID 규칙**
- **TAG 검색 패턴**
  ```bash
  rg "@REQ:AUTH-\d{3}" -n
  ```

#### 8.4 API 문서 (api-docs.md) - 새 페이지
- **TypeScript API**
  - Core Modules
  - CLI Commands
  - Tag System
  - Installer

- **타입 정의**
- **함수 시그니처**

#### 8.5 CLI 치트시트 (cli-cheatsheet.md) - 새 페이지
- **명령어 빠른 참조**
  ```bash
  moai init [project]         # 초기화
  moai doctor                 # 진단
  moai status [-v]            # 상태
  moai update [--check]       # 업데이트
  moai restore <path>         # 복원
  ```

- **에이전트 호출**
  ```
  @agent-debug-helper "문제"
  @agent-code-builder "SPEC-001"
  ```

- **워크플로우 명령어**
  ```
  /moai:1-spec "기능"
  /moai:2-build SPEC-ID
  /moai:3-sync
  ```

#### 8.6 용어집 (glossary.md) - 새 페이지
- **SPEC**: Specification (명세)
- **EARS**: Easy Approach to Requirements Syntax
- **TDD**: Test-Driven Development
- **TAG**: Traceability TAG
- **TRUST**: Test/Readable/Unified/Secured/Trackable
- **Primary Chain**: @REQ → @DESIGN → @TASK → @TEST

---

### 🆘 **9. 트러블슈팅 & FAQ** (Troubleshooting) - 새 섹션

#### 9.1 일반적인 문제 (common-issues.md) - 새 페이지
- **설치 오류**
  - 권한 문제
  - TypeScript 오류
  - Git 설정

- **실행 오류**
  - moai doctor 실패
  - moai init 실패
  - Claude Code 연동 오류

- **성능 문제**
  - 느린 빌드
  - 메모리 부족
  - TAG 스캔 지연

#### 9.2 FAQ (faq.md) - 새 페이지
- **일반**
  - MoAI-ADK는 무엇인가요?
  - 왜 SPEC-First인가요?
  - Python에서 TypeScript로 전환한 이유는?

- **설치 & 설정**
  - 시스템 요구사항은?
  - Personal vs Team 모드 차이는?
  - 기존 프로젝트에 적용 가능한가요?

- **사용법**
  - SPEC 작성은 어떻게 하나요?
  - TAG는 어떻게 관리하나요?
  - 언어 자동 감지는 어떻게 동작하나요?

- **고급**
  - 커스텀 에이전트를 만들 수 있나요?
  - CI/CD 통합은 어떻게 하나요?
  - 대규모 프로젝트에서도 사용 가능한가요?

#### 9.3 디버깅 가이드 (debugging.md) - 새 페이지
- **debug-helper 에이전트 활용**
  ```
  @agent-debug-helper "오류내용"
  @agent-debug-helper "TAG 체인 검증"
  ```

- **로그 분석**
  - `.moai/logs/` 확인
  - Winston logger 로그 레벨
  - 민감정보 마스킹 확인

- **시스템 진단**
  ```bash
  moai doctor --verbose
  ```

- **문제 보고**
  - GitHub Issues 템플릿
  - 필요한 정보
  - 로그 첨부

---

### 📖 **10. 커뮤니티 & 기여** (Community) - 새 섹션

#### 10.1 기여 가이드 (contributing.md) - 새 페이지
- **기여 방법**
  1. Fork & Clone
  2. 개발 환경 설정
  3. SPEC 작성 (`/moai:1-spec`)
  4. TDD 구현 (`/moai:2-build`)
  5. 문서 동기화 (`/moai:3-sync`)
  6. PR 제출

- **코드 스타일**
  - TypeScript strict 모드
  - Biome 린터+포맷터
  - TRUST 5원칙 준수

- **테스트 요구사항**
  - Vitest 테스트 작성
  - 커버리지 85% 이상
  - 모든 테스트 통과

#### 10.2 릴리스 노트 (changelog.md) - 새 페이지
- **v2.0.0 (SPEC-013)** - TypeScript 완전 전환
- **v0.1.28** - Python 최종 버전
- **v0.0.4** - TRUST 92% 달성
- **v0.0.1** - 초기 릴리스

#### 10.3 로드맵 (roadmap.md) - 새 페이지
- **Phase 1 (완료)**: TypeScript 전환
- **Phase 2 (진행 중)**: 범용 언어 지원 강화
- **Phase 3 (계획)**: 웹 대시보드
- **Phase 4 (계획)**: VSCode Extension

#### 10.4 커뮤니티 (community.md) - 새 페이지
- **GitHub Discussions**
- **Discord/Slack** (오픈 예정)
- **공식 문서**: https://adk.mo.ai.kr
- **커뮤니티 포럼**: https://mo.ai.kr (오픈 예정)

---

## 🎯 Navigation 구조 (VitePress config.ts)

```typescript
nav: [
  { text: '홈', link: '/' },
  { text: '가이드', link: '/guide/introduction' },
  { text: '시작하기', link: '/getting-started/installation' },
  {
    text: '핵심 개념',
    items: [
      { text: 'SPEC-First TDD', link: '/concepts/spec-first-tdd' },
      { text: 'TAG 시스템', link: '/concepts/tag-system' },
      { text: '워크플로우', link: '/concepts/workflow' },
      { text: 'TRUST 원칙', link: '/concepts/trust-principles' }
    ]
  },
  { text: 'CLI', link: '/cli/init' },
  {
    text: 'Claude Code',
    items: [
      { text: '에이전트', link: '/claude/agents' },
      { text: '명령어', link: '/claude/commands' },
      { text: '훅', link: '/claude/hooks' }
    ]
  },
  {
    text: '언어별 가이드',
    items: [
      { text: 'TypeScript', link: '/languages/typescript' },
      { text: 'Python', link: '/languages/python' },
      { text: 'Java', link: '/languages/java' },
      { text: 'Go', link: '/languages/go' },
      { text: 'Rust', link: '/languages/rust' }
    ]
  },
  {
    text: '고급',
    items: [
      { text: '커스텀 에이전트', link: '/advanced/custom-agents' },
      { text: 'CI/CD 통합', link: '/advanced/ci-cd' },
      { text: '팀 협업', link: '/advanced/team-collaboration' },
      { text: '성능 최적화', link: '/advanced/performance' }
    ]
  },
  {
    text: '레퍼런스',
    items: [
      { text: '설정', link: '/reference/configuration' },
      { text: 'EARS 템플릿', link: '/reference/ears-templates' },
      { text: 'TAG 레퍼런스', link: '/reference/tag-reference' },
      { text: 'API 문서', link: '/reference/api-docs' },
      { text: 'CLI 치트시트', link: '/reference/cli-cheatsheet' }
    ]
  },
  {
    text: '도움말',
    items: [
      { text: '트러블슈팅', link: '/help/common-issues' },
      { text: 'FAQ', link: '/help/faq' },
      { text: '디버깅', link: '/help/debugging' }
    ]
  },
  {
    text: '커뮤니티',
    items: [
      { text: '기여하기', link: '/community/contributing' },
      { text: '릴리스 노트', link: '/community/changelog' },
      { text: '로드맵', link: '/community/roadmap' },
      { text: 'GitHub', link: 'https://github.com/modu-ai/moai-adk' }
    ]
  }
]
```

---

##  문서 작성 우선순위

### Phase 1: 핵심 문서 (1-2주)
1. ✅ 홈 & 개요 (index.md, introduction.md, features.md)
2. ✅ 시작하기 (installation.md, quick-start.md, project-setup.md)
3. ✅ 핵심 개념 (spec-first-tdd.md, tag-system.md, workflow.md, trust-principles.md)
4. ✅ CLI 레퍼런스 (init.md, doctor.md, status.md, update.md, restore.md, help.md)

### Phase 2: Claude Code & 언어 (2-3주)
5. 🔄 Claude Code 통합 (agents.md, commands.md, hooks.md, output-styles.md)
6. 🔄 언어별 가이드 (typescript.md, python.md, java.md, go.md, rust.md, other-languages.md)

### Phase 3: 고급 & 레퍼런스 (3-4주)
7. 📋 고급 주제 (custom-agents.md, ci-cd.md, team-collaboration.md, performance.md, security.md, migration.md)
8. 📋 레퍼런스 (configuration.md, ears-templates.md, tag-reference.md, api-docs.md, cli-cheatsheet.md, glossary.md)

### Phase 4: 커뮤니티 & 도움말 (4-5주)
9. 📋 트러블슈팅 (common-issues.md, faq.md, debugging.md)
10. 📋 커뮤니티 (contributing.md, changelog.md, roadmap.md, community.md)

---

## 문서 작성 원칙

### 1. 사용자 여정 중심
- 초보자 → 실무자 → 전문가 순차적 학습
- 각 문서는 이전 문서 기반으로 작성
- "다음 단계" 링크로 학습 경로 안내

### 2. 실전 중심
- 모든 개념에 코드 예시 포함
- 실행 가능한 명령어 제공
- 실제 프로젝트 시나리오 사용

### 3. 범용 언어 지원
- TypeScript 예시를 기본으로
- Python, Java, Go, Rust 예시 추가
- 언어 간 일관성 유지

### 4. 시각적 요소
- 다이어그램: mermaid.js 사용
- 스크린샷: 주요 기능 시연
- 코드 블록: 문법 하이라이팅

### 5. 검색 최적화
- 명확한 제목 (H1, H2, H3)
- 키워드 포함
- 메타데이터 설정

---

##  문서 템플릿

### 기본 템플릿
```markdown
---
title: [페이지 제목]
description: [간단한 설명]
---

# [페이지 제목]

> [핵심 요약 1-2줄]

## 개요

[도입부]

## 주요 내용

### 섹션 1

[내용]

```[language]
// 코드 예시
```

### 섹션 2

[내용]

## 실습 예제

[단계별 실습]

## 다음 단계

- [관련 문서 1]
- [관련 문서 2]

## 참고 자료

- [외부 링크]
```

---

이 새로운 구조로 온라인 문서를 재작성하면:
- ✅ 사용자 여정 명확화
- ✅ 학습 곡선 완화
- ✅ 실전 활용 강화
- ✅ 범용 언어 지원 강조
- ✅ 검색 최적화