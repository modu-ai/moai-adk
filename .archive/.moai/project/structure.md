# MoAI-ADK Structure Design

## @DOC:ARCHITECTURE-001 시스템 아키텍처 (v0.0.4 완성)

### TypeScript 기반 모듈화 아키텍처 ✅

MoAI-ADK는 **TypeScript CLI + 모듈화 Core Engine + 분산 TAG 시스템 + Claude Code 통합** 구조로 완성되어 단일 패키지 배포와 고성능 사용자 경험을 제공합니다.

```
MoAI-ADK v0.0.4 Architecture (모듈화 완성)
├── CLI Layer (TypeScript)    # ✅ 7개 명령어 100% 완성
├── Core Engine             # ✅ 모듈화 설계 (Orchestrator 91% 감소)
│   ├── installer/          # ✅ 9개 모듈 분해 (DI 패턴)
│   │   ├── orchestrator.ts (135 LOC)
│   │   ├── context-manager.ts (127 LOC)
│   │   ├── result-builder.ts (124 LOC)
│   │   ├── phase-executor.ts (496 LOC)
│   │   ├── phase-validator.ts (198 LOC)
│   │   ├── resource-installer.ts (241 LOC)
│   │   ├── fallback-builder.ts (297 LOC)
│   │   └── template-processor.ts (225 LOC)
│   ├── system-checker/     # ✅ 시스템 진단 (4모듈)
│   └── git/                # ✅ Git 자동화 (테스트 안정화)
├── Security System         # ✅ Winston logger (97.92% coverage)
├── Distributed TAG System  # ✅ 94% 최적화 (149 TAGs)
└── Claude Extensions        # ✅ 7에이전트+5명령+8훅
```

**성과 (v0.0.4)**:
- TRUST 92% 달성 (Elite 등급)
- Orchestrator 모듈화 (1,467 → 135 LOC, 91% 감소)
- Winston logger 보안 시스템 구축
- 테스트 안정화 완료

## @DOC:MODULES-001 모듈별 책임 구분

### 1. CLI Layer (`src/cli/`) - TypeScript 기반 ✅

- **책임**: 사용자 명령어 처리, 진단 시스템, 프로젝트 관리
- **입력**: CLI 명령어, 사용자 옵션
- **처리**: Commander.js 파싱, 진단 자동화, 상태 보고
- **출력**: 고성능 터미널 출력, 진단 리포트

| 명령어                   | 상태         | 주요 기능                       |
| ------------------------ | ---------- | ------------------------------- |
| `moai init`              | ✅ **완성** | 프로젝트 초기화, 템플릿 설치   |
| `moai doctor`            | ✅ **완성** | 시스템 요구사항 자동 진단       |
| `moai status`            | ✅ **완성** | 프로젝트 상태 및 TAG 추적성        |
| `moai update`            | ✅ **완성** | 템플릿 업데이트, 버전 동기화    |
| `moai restore`           | ✅ **완성** | 백업 복원, 설정 리셋          |
| `moai help`              | ✅ **완성** | 전체 도움말, 사용법 가이드       |
| `moai --version`         | ✅ **완성** | 버전 정보, 빌드 메타데이터       |

### 2. Core Engine (`src/core/`) - 모듈화 완성 ✅

- **책임**: 시스템 진단, 프로젝트 관리, Git 자동화, TAG 시스템, 설치 오케스트레이션
- **입력**: 시스템 상태, 프로젝트 설정, 템플릿 데이터
- **처리**: 요구사항 검증, 자동 배치, 성능 모니터링, 모듈화 설치
- **출력**: 진단 리포트, 프로젝트 구조, 추적성 데이터

| 모듈                     | 상태       | 주요 기능                       |
| ---------------------- | -------- | ------------------------------- |
| `installer/`           | ✅ **v0.0.4 완성** | 모듈화 설치 오케스트레이션 (9개 모듈) |
| `system-checker/`      | ✅ **완성** | Node.js, Git, 버전 자동 검증      |
| `package-manager/`     | ✅ **완성** | npm, Bun, 의존성 관리           |
| `project/`             | ✅ **완성** | 프로젝트 위저드, 템플릿 관리     |
| `git/`                 | ✅ **v0.0.4 완성** | Git 자동화, 테스트 안정화 완료   |
| `tag-system/`          | ✅ **완성** | 분산 16-Core TAG 시스템        |

#### 2.1. Installer System (`src/core/installer/`) - v0.0.4 모듈화 완성 ✅

**Phase 2 대규모 리팩토링 성과**:
- **이전**: orchestrator.ts (1,467 LOC) - 단일 거대 파일
- **이후**: 9개 전문 모듈 (135~496 LOC) - 의존성 주입 패턴
- **감소율**: 91% LOC 감소 (R: Readable 52% → 100%)

| 모듈                          | LOC  | 역할                    | 주요 기능                              |
| ----------------------------- | ---- | ----------------------- | -------------------------------------- |
| `orchestrator.ts`            | 135  | 설치 조정               | PhaseExecutor/ContextManager/ResultBuilder 통합 |
| `context-manager.ts`         | 127  | 컨텍스트 관리           | 설치 컨텍스트 생성, 상태 추적           |
| `result-builder.ts`          | 124  | 결과 수집               | 단계별 결과 집계, 리포트 생성           |
| `phase-executor.ts`          | 496  | 단계 실행               | Phase 실행 조정, 의존성 관리           |
| `phase-validator.ts`         | 198  | 검증 로직               | 의존성 검증, 전제 조건 확인            |
| `resource-installer.ts`      | 241  | 리소스 설치             | 템플릿 복사, 파일 작업 조정            |
| `fallback-builder.ts`        | 297  | 대체 전략               | 실패 시 복구 전략 생성                 |
| `template-processor.ts`      | 225  | 템플릿 처리             | 템플릿 렌더링, 변수 치환               |
| `*-types.ts`                 | -    | 타입 정의               | 공유 인터페이스, 타입 안전성           |

**v0.0.4 적용 패턴**:
- **의존성 주입 (DI)**: 모든 모듈이 생성자 주입 패턴 사용
- **단일 책임 원칙 (SRP)**: 각 모듈이 하나의 책임만 담당
- **인터페이스 분리 (ISP)**: 명확한 타입 정의 및 계약
- **개방-폐쇄 원칙 (OCP)**: 확장 가능하되 수정 불필요

#### 2.1. Documentation System (`src/moai_adk/core/docs/`) - SPEC-010 추가

- **책임**: 온라인 문서 사이트 자동 생성 및 관리
- **입력**: Python 소스코드, sync-report, mkdocs 설정
- **처리**: API 문서 생성, 릴리스 노트 변환, 문서 빌드
- **출력**: 완전 자동화된 MkDocs 기반 온라인 문서

| 모듈                          | 역할                    | 주요 기능                              |
| ----------------------------- | ----------------------- | -------------------------------------- |
| `documentation_builder.py`   | MkDocs 빌드 관리        | 사이트 구조 초기화, 빌드 검증          |
| `api_generator.py`           | API 문서 자동 생성      | 소스코드 → API 문서, 네비게이션 생성   |
| `release_notes_converter.py` | 릴리스 노트 변환        | sync-report → 릴리스 노트 자동 변환    |

### 3. Install System (`src/moai_adk/install/`)

- **책임**: 패키지 설치, 리소스 관리, 버전 동기화
- **입력**: 설치 요청, 환경 정보
- **처리**: 종속성 확인, 리소스 복사, 설정 적용
- **출력**: 설치 완료 상태, 오류 리포트

| 모듈                     | 역할                     | 주요 기능                           |
| ------------------------ | ------------------------ | ----------------------------------- |
| `installer.py`           | 설치 오케스트레이션      | 전체 설치 프로세스 관리             |
| `resource_manager.py`    | 리소스 관리 (통합)       | 전체 리소스 관리 오케스트레이션     |
| `template_manager.py`    | 템플릿 관리              | 템플릿 발견, 로딩, 렌더링           |
| `file_operations.py`     | 파일 작업                | 파일 복사, 디렉토리 작업, 권한 관리 |
| `resource_validator.py`  | 리소스 검증              | 경로 검증, 보안 검사, 리소스 확인   |
| `post_install.py`        | 설치 후 작업             | 권한 설정, 환경 검증                |
| `post_install_hook.py`   | 설치 후 자동화 시스템    | Python 명령어 감지, Claude 설정     |
| `installation_result.py` | 설치 결과 관리           | 설치 상태 추적, 결과 리포트         |

### 3. Claude Extensions (`.claude/`) - 완전 통합 ✅

- **책임**: Claude Code 네이티브 통합, SPEC-First TDD 워크플로우
- **입력**: Claude Code 이벤트, 사용자 명령, 소스 변경
- **처리**: 에이전트 오케스트레이션, 자동 검증, 리빙 독 동기화
- **출력**: 자동화된 3단계 워크플로우, 품질 보장

| 계층               | 상태       | 구성                           |
| ---------------- | -------- | -------------------------------- |
| **Agents**       | ✅ **완성** | 7개 전문 에이전트 (전체 워크플로우) |
| **Commands**     | ✅ **완성** | 5개 핵심 명령어 (0-4단계)        |
| **Hooks**        | ✅ **완성** | 8개 이벤트 훅 (보안, 모니터링)     |
| **Output Styles** | ✅ **완성** | 5개 출력 스타일 (학습, 페어, 초보) |

## @DOC:INTEGRATION-001 외부 시스템 통합

### Git/GitHub 연동

- **인증 방식**: GitHub CLI (`gh`) 토큰 기반
- **Personal 모드**: 로컬 Git 저장소 관리
- **Team 모드**: GitHub API를 통한 Issue/PR 자동화
- **장애 시 대체**: 로컬 Git 명령어로 fallback

### Claude Code API 연동

- **인증 방식**: Claude Code 내장 권한 시스템
- **데이터 교환**: JSON 기반 에이전트 통신
- **변경 불가 시스템**: Claude Code 내부 API (외부 제어 불가)
- **위험도**: 중간 (API 변경 시 호환성 검토 필요)

### 패키지 생태계

- **PyPI**: 주 배포 채널
- **pip/pipx**: 표준 설치 도구
- **conda/mamba**: 선택적 지원 (conda-forge 등록 계획)

## @DOC:TRACEABILITY-001 추적성 전략

### 16-Core TAG 체계 적용

- **Primary Chain**: `@SPEC → @SPEC → @CODE → @TEST`
- **Implementation**: `@CODE → @CODE → @CODE → @CODE`
- **Quality**: `@CODE → @CODE → @DOC → @TAG`

### 분산 TAG 시스템 v4.0 관리 ✅

- **자동 생성**: JSONL 기반 분산 저장, 94% 크기 절감
- **추적 범위**: 149개 TAG, 122개 파일, 100% 추적성
- **성능 지표**: 95% 파싱 속도 향상, 90% 메모리 절약
- **로딩 속도**: 45ms 평균, 487KB 최적화 달성

## SUCCESS:CURRENT-STATE-001 현재 구조 달성 상태 ✅

### TypeScript 기반 프로젝트 구조 (v0.0.1)

```
moai-adk-ts/ (완성)
├── src/
│   ├── cli/                    # ✅ 7개 명령어 100% 완성
│   │   ├── index.ts            # Commander.js 진입점
│   │   └── commands/           # init, doctor, status, update, restore
│   ├── core/                   # ✅ 핵심 엔진 모듈
│   │   ├── system-checker/     # 시스템 진단 (4모듈)
│   │   ├── package-manager/    # 패키지 관리
│   │   ├── project/            # 프로젝트 관리
│   │   ├── git/                # Git 자동화
│   │   └── tag-system/         # 16-Core TAG 시스템
│   ├── utils/                  # ✅ 공통 유틸리티
│   └── index.ts                # 메인 API 진입점
├── templates/                  # ✅ 프로젝트 템플릿
├── __tests__/                  # ✅ 테스트 수이트 (100% 통과)
└── dist/                       # ✅ ESM/CJS 듀얼 배포 (226ms 빌드, 471KB)
```

### SUCCESS:TYPESCRIPT-FOUNDATION-012 TypeScript 기반 구축 완료 ✅

**SPEC-012 Week 1 완성** (@CODE:WEEK1-012) ✅

```
moai-adk-ts/
├── package.json              # npm 패키지 설정 (v0.0.1)
├── tsconfig.json            # TypeScript strict 모드 설정
├── tsup.config.ts           # 고성능 빌드 설정 (686ms)
├── jest.config.js           # Jest 테스트 환경
├── .eslintrc.json          # TypeScript ESLint 규칙
├── .prettierrc             # 코드 포맷팅 규칙
├── src/
│   ├── cli/
│   │   ├── index.ts        # Commander.js CLI 진입점
│   │   └── commands/
│   │       ├── init.ts     # moai init 명령어 구현
│   │       └── doctor.ts   # moai doctor 명령어 구현
│   ├── core/
│   │   └── system-checker/ # 🆕 혁신적 지능형 시스템 진단 (v0.0.3)
│   │       ├── requirements.ts  # RequirementRegistry + addLanguageRequirements()
│   │       ├── detector.ts      # 언어 패턴 감지 + 통계 분석
│   │       └── index.ts         # SystemChecker 통합 클래스 (5-category 진단)
│   ├── utils/
│   │   ├── logger.ts       # 구조화 로깅 시스템
│   │   └── version.ts      # 버전 정보 관리
│   └── index.ts            # 메인 API 진입점
├── __tests__/              # Jest 테스트 수트 (100% 통과)
│   ├── system-checker/     # 시스템 검증 테스트
│   └── cli/               # CLI 테스트
└── dist/                  # ESM/CJS 듀얼 컴파일 결과 (226ms 빌드, 471KB)
```

**핵심 달성 성과:**
1. **CLI 100% 완성**: 7개 명령어 완전 동작 (초기화, 진단, 상태, 업데이트, 복원)
2. **성능 최적화**: Bun 98%, Vitest 92.9%, Biome 94.8% 성능 향상
3. **분산 TAG 시스템**: 94% 크기 절감, 149개 TAG 완전 추적성
4. **TRUST 5원칙**: Test First, Readable, Unified, Secured, Trackable 100% 준수
5. **현대화 스택**: TypeScript 5.9.2 + Bun 1.2.19 + Vitest + Biome 완성

### Claude Code 통합 현황

```
.claude/
├── agents/moai/   # 7개 핵심 에이전트 (2개 비활성화)
├── commands/moai/ # 5개 워크플로우 명령어
├── hooks/moai/    # 8개 이벤트 훅
└── output-styles/ # 5개 출력 스타일
```

### SUCCESS:SYSTEM-DIAGNOSIS-INNOVATION-001 혁신적 시스템 진단 아키텍처 완성 ✅ (v0.0.3)

#### 1. **SystemChecker 아키텍처 혁신** (@SPEC:SYSTEM-CHECKER-001) ✅
   - **RequirementRegistry**: 동적 요구사항 관리 중앙화
   - **addLanguageRequirements()**: 감지된 언어별 맞춤형 도구 자동 추가
   - **5-category 진단**: Runtime(2) + Development(2) + Optional(1) + Language-Specific + Performance

#### 2. **지능형 언어 감지 시스템** (@CODE:LANGUAGE-DETECTION-001) ✅
   - **파일 패턴 인식**: `.ts`, `.js`, `.py`, `.java`, `.go`, `package.json`, `requirements.txt`
   - **통계 분석**: 다중 언어 프로젝트에서 언어 비중 자동 계산
   - **동적 도구 매핑**: JavaScript/TypeScript/Python/Java/Go → 해당 개발 도구 자동 추가

#### 3. **실용성 혁신** (@CODE:PRACTICAL-IMPROVEMENTS-001) ✅
   - ❌ **SQLite3 제거**: 불필요한 데이터베이스 의존성 완전 제거
   - ✅ **npm 추가**: TypeScript 생태계 필수 도구
   - ✅ **TypeScript 추가**: 메인 개발 도구체인 필수 요소
   - ✅ **Git LFS 추가**: 현대적 대용량 파일 지원

#### 4. **성능 최적화 달성** (@CODE:SYSTEM-DIAGNOSIS-001) ✅
   - **빌드 시간**: 226ms (이전 대비 개선)
   - **패키지 크기**: 471KB (v0.0.3 최적화)
   - **진단 속도**: 실시간 언어 감지 및 요구사항 동적 생성

### SUCCESS:REFACTOR-001 구조 개선 완료 성과 ✅

1. **CLI 모듈 분해 완료** (@CODE:CLI-MODULE-SPLIT-001) ✅
   - commands.py (179 LOC) → 4개 전문 모듈로 분해
   - 각 모듈의 단일 책임 원칙 준수 (TRUST-U)
   - 명령어 실행 로직의 명확한 분리

2. **TRUST 원칙 준수** (@CODE:TRUST-COMPLIANCE-001) ✅
   - 모든 새 CLI 모듈이 50 LOC 이하 단일 책임 구현
   - 명확한 에러 처리 및 로깅 전략
   - 보안 검증 로직 분리 (ResourceValidator)

3. **새로운 기능 추가** (@CODE:CROSS-PLATFORM-HOOKS) ✅
   - post_install_hook.py: Python 명령어 자동 감지 시스템
   - 크로스 플랫폼 Claude 설정 자동 생성
   - 환경별 최적화된 훅 명령어 생성

4. **문서 시스템 통합 최적화** (@CODE:DOCS-INTEGRATION-001) - SPEC-010 완료 ✅
   - MkDocs 빌드 성능 0.54초 달성
   - API 문서 85개 모듈 자동 생성
   - 릴리스 노트 변환 로직 고도화 완료

### @CODE:NEW-TAGS-001 새로운 TAG 체인 완성

**CLI 모듈 리팩토링 TAG 체인:**
```
@SPEC:CLI-REFACTOR-001 → @SPEC:CLI-SPLIT-001 →
@CODE:CLI-COMMANDS-001, @CODE:CLI-EXECUTOR-001, @CODE:CLI-OPERATIONS-001, @CODE:CLI-UTILS-001 →
@CODE:CLI-UNIFIED-001 → @TEST:CLI-EXECUTION-001 ✅
```

## @DOC:NEXT-PHASE-001 다음 단계 발전 계획

### Phase 2: 확장 및 통합 (예정)
1. **범용 언어 지원 강화** - Java, Go, Rust, C# 등 추가 언어
2. **웹 대시보드 개발** - 실시간 프로젝트 모니터링 및 분석
3. **GitHub Actions 완전 통합** - CI/CD 파이프라인 자동화
4. **VSCode Extension** - IDE 내 네이티브 통합 및 사용자 경험 개선

### 성능 및 확장성 목표
- **대용량 프로젝트 대응**: 10,000+ 파일 처리 능력
- **클라우드 동기화**: 팀 협업 및 원격 개발 지원
- **AI 도구 통합**: 추가 AI 모델 연동 및 상호 운용성

---

_이 구조는 현재 v0.0.1 달성 상태를 반영하며, `/moai:2-build` 실행 시 TDD 구현의 가이드라인이 됩니다._