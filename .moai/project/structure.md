# MoAI-ADK Structure Design

## @STRUCT:ARCHITECTURE-001 시스템 아키텍처

### 모듈 전략: 핵심 3모듈 + 도메인 확장

MoAI-ADK는 **핵심 3모듈(CLI/Core/Install) + Claude Code 확장** 구조로 설계되어 단일 패키지로 배포하되, 사용자 정의 확장을 지원합니다.

```
MoAI-ADK Architecture
├── CLI Layer          # 사용자 인터페이스
├── Core Engine        # 핵심 비즈니스 로직
├── Install System     # 설치/배포 관리
└── Claude Extensions  # 에이전트/명령어/훅
```

**선택 이유**: 단일 패키지로 배포 단순성을 유지하면서, Claude Code 생태계를 통한 확장성 제공

## @STRUCT:MODULES-001 모듈별 책임 구분

### 1. CLI Layer (`src/moai_adk/cli/`)

- **책임**: 사용자 명령어 처리, 대화형 인터페이스
- **입력**: CLI 명령어, 사용자 응답
- **처리**: 명령어 파싱, 위저드 실행, 진행률 표시
- **출력**: 터미널 출력, 상태 메시지

| 모듈                     | 역할                        | 주요 기능                       |
| ------------------------ | --------------------------- | ------------------------------- |
| `commands.py`            | 명령어 엔트리포인트         | `init`, `status`, `doctor` 등   |
| `command_executor.py`    | 기본 명령어 실행 로직       | `init`, `restore`, `doctor` 구현 |
| `command_operations.py`  | 복잡 명령어 처리 로직       | `status`, `update` 구현         |
| `command_utils.py`       | CLI 유틸리티 함수           | 모드 설정, 설정 관리            |
| `wizard.py`              | 대화형 설치 가이드          | 프로젝트 유형별 질문/응답 처리  |
| `banner.py`              | UI/UX 요소                  | 브랜딩, 진행률 표시             |

### 2. Core Engine (`src/moai_adk/core/`)

- **책임**: 핵심 비즈니스 로직, 파일/Git 관리, 문서 생성
- **입력**: 설정 데이터, 템플릿 경로, 소스 코드
- **처리**: 파일 생성/수정, Git 작업, 보안 검증, 문서 자동화
- **출력**: 프로젝트 구조, 설정 파일, 온라인 문서

| 모듈                   | 역할               | 주요 기능                       |
| ---------------------- | ------------------ | ------------------------------- |
| `directory_manager.py` | 디렉토리 구조 관리 | `.moai/`, `.claude/` 생성       |
| `git_manager.py`       | Git 작업 자동화    | 저장소 초기화, 브랜치/커밋 관리 |
| `config_manager.py`    | 설정 관리          | Personal/Team 모드 전환         |
| `file_manager.py`      | 파일 작업          | 템플릿 복사, 권한 설정          |
| `security.py`          | 보안 검증          | 민감정보 검사, 권한 확인        |

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

### 4. Claude Extensions (`.claude/`)

- **책임**: Claude Code 환경 통합, 워크플로우 자동화
- **입력**: Claude Code 이벤트, 사용자 명령
- **처리**: 에이전트 실행, 훅 처리, 정책 검증
- **출력**: 자동화된 작업 실행, 정책 위반 알림

| 계층         | 구성 요소   | 역할                     |
| ------------ | ----------- | ------------------------ |
| **Agents**   | `moai/*.md` | 워크플로우 에이전트 정의 |
| **Commands** | `moai/*.md` | 4단계 파이프라인 명령어  |
| **Hooks**    | `moai/*.py` | 이벤트 처리, 정책 검증   |

## @STRUCT:INTEGRATION-001 외부 시스템 통합

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

## @STRUCT:TRACEABILITY-001 추적성 전략

### 16-Core TAG 체계 적용

- **Primary Chain**: `@REQ → @DESIGN → @TASK → @TEST`
- **Implementation**: `@FEATURE → @API → @UI → @DATA`
- **Quality**: `@PERF → @SEC → @DOCS → @TAG`

### TAG 인덱스 관리

- **자동 생성**: `/moai:3-sync` 실행 시 `.moai/indexes/tags.json` 갱신
- **추적 범위**: 소스 코드, 문서, 커밋 메시지, Issue/PR
- **유지 주기**: 매 동기화 사이클 (워크플로우 실행 시)

## Legacy Context

### 기존 모듈 구조 현황

```
src/moai_adk/
├── cli/           # 완성된 CLI 인터페이스 (13개 모듈)
├── core/          # 핵심 엔진 (14개 모듈) - SPEC-010 추가
│   ├── docs/      # 🆕 온라인 문서 시스템 (3개 모듈)
│   ├── quality/   # 품질 개선 시스템
│   └── (기타)     # 기존 모듈들
├── install/       # 설치 시스템 (8개 모듈)
├── utils/         # 공통 유틸리티 (3개 모듈)
└── resources/     # 템플릿/스크립트 (9개 템플릿)
```

### Claude Code 통합 현황

```
.claude/
├── agents/moai/   # 7개 핵심 에이전트 (2개 비활성화)
├── commands/moai/ # 5개 워크플로우 명령어
├── hooks/moai/    # 8개 이벤트 훅
└── output-styles/ # 5개 출력 스타일
```

### @SUCCESS:REFACTOR-001 구조 개선 완료 성과 ✅

1. **CLI 모듈 분해 완료** (@TASK:CLI-MODULE-SPLIT-001) ✅
   - commands.py (179 LOC) → 4개 전문 모듈로 분해
   - 각 모듈의 단일 책임 원칙 준수 (TRUST-U)
   - 명령어 실행 로직의 명확한 분리

2. **TRUST 원칙 준수** (@TASK:TRUST-COMPLIANCE-001) ✅
   - 모든 새 CLI 모듈이 50 LOC 이하 단일 책임 구현
   - 명확한 에러 처리 및 로깅 전략
   - 보안 검증 로직 분리 (ResourceValidator)

3. **새로운 기능 추가** (@FEATURE:CROSS-PLATFORM-HOOKS) ✅
   - post_install_hook.py: Python 명령어 자동 감지 시스템
   - 크로스 플랫폼 Claude 설정 자동 생성
   - 환경별 최적화된 훅 명령어 생성

4. **문서 시스템 통합 최적화** (@TASK:DOCS-INTEGRATION-001) - SPEC-010 완료 ✅
   - MkDocs 빌드 성능 0.54초 달성
   - API 문서 85개 모듈 자동 생성
   - 릴리스 노트 변환 로직 고도화 완료

### @TASK:NEW-TAGS-001 새로운 TAG 체인 완성

**CLI 모듈 리팩토링 TAG 체인:**
```
@REQ:CLI-REFACTOR-001 → @DESIGN:CLI-SPLIT-001 →
@TASK:CLI-COMMANDS-001, @TASK:CLI-EXECUTOR-001, @TASK:CLI-OPERATIONS-001, @TASK:CLI-UTILS-001 →
@FEATURE:CLI-UNIFIED-001 → @TEST:CLI-EXECUTION-001 ✅
```

## @TODO:MIGRATION-002 Initial Migration Tasks

1. **모듈 간 인터페이스 문서화** - 각 모듈의 공개 API 명세 작성
2. **에이전트 오케스트레이션 규칙** - 에이전트 간 호출 규칙 및 데이터 흐름 정의
3. **에러 처리 표준화** - 전역 에러 처리 및 사용자 피드백 메커니즘 통일
4. **크로스 플랫폼 호환성 검증** - Windows/macOS/Linux 환경별 동작 검증

---

_이 구조는 `/moai:2-build` 실행 시 TDD 구현의 가이드라인이 됩니다._