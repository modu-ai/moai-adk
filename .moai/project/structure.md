---
id: STRUCTURE-001
version: 0.1.2
status: active
created: 2025-10-01
updated: 2025-10-21
author: @architect
priority: medium
---

# MoAI-ADK Structure Design

## HISTORY

### v0.1.2 (2025-10-21)
- **UPDATED**: Backup merge - 템플릿 플레이스홀더를 실제 MoAI-ADK 아키텍처로 교체
- **AUTHOR**: @Alfred
- **SECTIONS**: Architecture (6-Layer), Modules (CLI/Core/Template/Project), Integration (Claude Code/Git/PyPI)
- **REASON**: Smart Update 전략에 따라 실제 프로젝트 구조 반영

### v0.1.1 (2025-10-17)
- **UPDATED**: 템플릿 버전 동기화 (v0.3.8)
- **AUTHOR**: @Alfred
- **SECTIONS**: 메타데이터 표준화 (author 필드 단수형, priority 추가)

### v0.1.0 (2025-10-01)
- **INITIAL**: 프로젝트 구조 설계 문서 작성
- **AUTHOR**: @architect
- **SECTIONS**: Architecture, Modules, Integration, Traceability

---

## @DOC:ARCHITECTURE-001 시스템 아키텍처

### 아키텍처 전략

**MoAI-ADK는 계층화된 책임 분리 아키텍처를 채택합니다.** 각 레이어는 명확한 책임을 가지며, 상위 레이어는 하위 레이어에만 의존합니다.

```
MoAI-ADK Architecture (6 Layers)
├── Layer 1: User                    # 최종 의사결정, 요구사항 제공
├── Layer 2: Alfred (SuperAgent)     # 전략적 의사결정, 에이전트 조율
├── Layer 3: Commands                # 워크플로우 템플릿, Phase 구조
├── Layer 4: Sub-agents              # 전문 에이전트 (spec-builder, tdd-implementer 등)
├── Layer 5: Skills                  # 재사용 가능 기능 모듈 (TAG 스캔, TRUST 검증 등)
└── Layer 6: Tools                   # Claude Code 기본 도구 (Read, Write, Bash 등)
```

**선택 이유**:
- **명확한 책임 분리**: 각 레이어는 단일 책임 원칙(SRP)을 준수
- **확장성**: 새로운 에이전트나 Skills 추가가 용이
- **테스트 가능성**: 각 레이어를 독립적으로 테스트 가능
- **유지보수성**: 계층 간 경계가 명확하여 변경 영향 최소화

### 레이어별 책임

| 레이어            | 책임                           | 의사결정 범위                  | 예시                                    |
| ----------------- | ------------------------------ | ------------------------------ | --------------------------------------- |
| **User**          | 최종 승인, 요구사항 제공       | 워크플로우 진행 여부           | "Phase 2 진행", "파일 덮어쓰기 승인"    |
| **Alfred**        | 전략적 판단, 에이전트 조율     | Skills 선택, 순차/병렬 실행    | "spec-builder 호출", "Skills 자동 활성" |
| **Commands**      | 워크플로우 구조 정의           | Phase 순서, 의존성 힌트        | "/alfred:1-plan Phase 1 → Phase 2"      |
| **Sub-agents**    | 전술적 판단, 단일 업무 수행    | 구현 세부사항                  | "SPEC 문서 작성", "TDD 구현"            |
| **Skills**        | 도메인 로직 캡슐화             | 검증 기준 판단                 | "TAG 스캔", "TRUST 원칙 검증"           |
| **Tools**         | 파일/명령 실행                 | N/A (순수 실행 도구)           | Read, Write, Bash, Grep                 |

## @DOC:MODULES-001 모듈별 책임 구분

### 1. CLI 모듈 (`moai_adk/cli/`)

- **책임**: 사용자 명령 처리, 대화형 프롬프트, 배너 출력
- **입력**: CLI 명령어 (`moai-adk init`, `moai-adk doctor` 등)
- **처리**: Click 기반 명령 파싱, Questionary를 통한 사용자 상호작용
- **출력**: 프로젝트 초기화, 상태 보고, 진단 결과

| 컴포넌트          | 역할                | 주요 기능                                           |
| ----------------- | ------------------- | --------------------------------------------------- |
| `cli/main.py`     | CLI 엔트리포인트    | 명령 라우팅, 버전 표시                              |
| `cli/commands/`   | 명령 구현           | init, doctor, status, update, backup                |
| `cli/prompts/`    | 대화형 프롬프트     | 프로젝트 정보 수집, 언어/모드 선택                  |
| `utils/banner.py` | 시각적 피드백       | ASCII 배너, 진행 상태 표시                          |

### 2. Core 모듈 (`moai_adk/core/`)

- **책임**: 핵심 비즈니스 로직, 프로젝트 관리, Git 통합, 품질 검증
- **입력**: 프로젝트 경로, 설정 파일 (`.moai/config.json`)
- **처리**: 프로젝트 초기화, 템플릿 처리, Git 작업 자동화, TRUST 검증
- **출력**: 프로젝트 구조, Git 브랜치/커밋, 검증 보고서

| 컴포넌트                  | 역할              | 주요 기능                                             |
| ------------------------- | ----------------- | ----------------------------------------------------- |
| `core/project/`           | 프로젝트 생명주기 | 초기화, 검증, Phase 실행, 백업 관리                   |
| `core/template/`          | 템플릿 처리       | 플레이스홀더 교체, 병합, 백업                         |
| `core/git/`               | Git 자동화        | 브랜치 생성, 커밋, 체크포인트, 이벤트 감지            |
| `core/quality/`           | 품질 검증         | TRUST 원칙 검증, 테스트 커버리지 확인                 |
| `core/diagnostics/`       | 시스템 진단       | 환경 검증, slash 명령 검증                            |

### 3. Template 모듈 (`moai_adk/templates/`)

- **책임**: 프로젝트 템플릿 제공 (`.claude/`, `.moai/`, `.github/`)
- **입력**: 언어, 모드, 프로젝트 정보
- **처리**: 템플릿 파일 복사, 플레이스홀더 교체
- **출력**: 초기화된 프로젝트 구조

| 컴포넌트                        | 역할                | 주요 기능                                    |
| ------------------------------- | ------------------- | -------------------------------------------- |
| `templates/.claude/agents/`     | 에이전트 정의       | 11개 전문 에이전트 (spec-builder 등)         |
| `templates/.claude/commands/`   | 커맨드 워크플로우   | /alfred:0-project, 1-plan, 2-run, 3-sync     |
| `templates/.claude/skills/`     | Skills 모듈         | Foundation, Essentials, Domain, Language     |
| `templates/.moai/project/`      | 프로젝트 문서       | product.md, structure.md, tech.md            |
| `templates/.moai/memory/`       | 개발 가이드         | development-guide.md, spec-metadata.md 등    |

### 4. Project 모듈 (`moai_adk/core/project/`)

- **책임**: 프로젝트 상태 관리, Phase 실행, 백업/복원
- **입력**: 프로젝트 디렉토리, 사용자 명령
- **처리**: 4-Phase 워크플로우 실행, 충돌 해결, 백업 생성
- **출력**: 초기화된 프로젝트, Phase 실행 보고서

| 컴포넌트               | 역할            | 주요 기능                                         |
| ---------------------- | --------------- | ------------------------------------------------- |
| `initializer.py`       | 프로젝트 초기화 | 4-Phase 실행, 템플릿 복사, 설정 생성              |
| `phase_executor.py`    | Phase 실행      | Phase 1-4 순차 실행, 롤백 지원                    |
| `validator.py`         | 검증            | 프로젝트 구조 검증, 필수 파일 확인                |
| `backup_utils.py`      | 백업 관리       | 타임스탬프 백업 생성, 복원, 정리                  |

## @DOC:INTEGRATION-001 외부 시스템 통합

### Claude Code 연동

- **인증 방식**: N/A (로컬 통합, API 키 불필요)
- **데이터 교환**:
  - `.claude/` 디렉토리를 통한 에이전트/커맨드/Skills 제공
  - CLAUDE.md를 통한 프로젝트 컨텍스트 주입
  - MCP (Model Context Protocol) 준비 중 (planned)
- **장애 시 대체**: CLI 단독 실행 가능 (`moai-adk` 명령)
- **위험도**: 낮음 (Claude Code 없이도 프로젝트 초기화/관리 가능)

**통합 포인트**:
```
MoAI-ADK
    ↓ (템플릿 제공)
.claude/agents/        → Claude Code가 로드
.claude/commands/      → Claude Code가 로드
.claude/skills/        → Claude Code가 로드
CLAUDE.md              → Claude Code가 읽음
```

### Git 연동 (GitPython)

- **용도**: 자동 커밋, 브랜치 생성, 체크포인트 백업
- **의존성 수준**: 중간 (Git 없이도 기본 기능 동작, GitFlow는 Git 필수)
- **성능 요구사항**:
  - 브랜치 생성: <100ms
  - 커밋: <500ms
  - 체크포인트: <1s

**Git 작업 플로우**:
```
/alfred:1-plan
    ↓
git-manager: 브랜치 생성 (feature/SPEC-{ID})
    ↓
/alfred:2-run
    ↓
git-manager: TDD 커밋 (RED → GREEN → REFACTOR)
    ↓
/alfred:3-sync
    ↓
git-manager: 문서 동기화 커밋
```

### PyPI 배포 (pip/uv)

- **용도**: 패키지 배포 및 설치
- **의존성 수준**: 높음 (주 배포 채널)
- **성능 요구사항**:
  - 설치 시간: <30s (uv), <60s (pip)
  - 빌드 시간: <10s

**배포 플로우**:
```
GitHub Release
    ↓ (자동 트리거)
GitHub Actions CI/CD
    ↓
PyPI 배포
    ↓
사용자: uv tool install moai-adk
```

## @DOC:TRACEABILITY-001 추적성 전략

### TAG 체계 적용

**TDD 완벽 정렬**: SPEC → 테스트 → 구현 → 문서
- `@SPEC:ID` (.moai/specs/) → `@TEST:ID` (tests/) → `@CODE:ID` (src/) → `@DOC:ID` (docs/)

**구현 세부사항**: @CODE:ID 내부 주석 레벨
- `@CODE:ID:API` - REST API, GraphQL 엔드포인트
- `@CODE:ID:UI` - 컴포넌트, 뷰, 화면
- `@CODE:ID:DATA` - 데이터 모델, 스키마, 타입
- `@CODE:ID:DOMAIN` - 비즈니스 로직, 도메인 규칙
- `@CODE:ID:INFRA` - 인프라, 데이터베이스, 외부 연동

### TAG 추적성 관리 (코드 스캔 방식)

- **검증 방법**: `/alfred:3-sync` 실행 시 `rg '@(SPEC|TEST|CODE|DOC):' -n`으로 코드 전체 스캔
- **추적 범위**: 프로젝트 전체 소스코드 (.moai/specs/, tests/, src/, docs/)
- **유지 주기**: 코드 변경 시점마다 실시간 검증
- **CODE-FIRST 원칙**: TAG의 진실은 코드 자체에만 존재

### 실제 TAG 예시 (MoAI-ADK 프로젝트)

```python
# @CODE:INIT-001 | SPEC: SPEC-INIT-001.md | TEST: tests/unit/test_initializer.py
# 프로젝트 초기화 핵심 로직
class ProjectInitializer:
    """
    4-Phase 프로젝트 초기화 책임자
    """
    def initialize(self, project_dir: Path) -> None:
        """Phase 1-4 순차 실행"""
        ...
```

```python
# @TEST:INIT-001 | SPEC: SPEC-INIT-001.md
def test_initialize_creates_directory_structure():
    """Phase 1: 디렉토리 구조 생성 검증"""
    ...
```

## Legacy Context

### 기존 시스템 현황

**MoAI-ADK는 새로운 프로젝트로, 레거시 시스템 없음.**

기존에 사용자가 수동으로 관리하던 작업들을 자동화:
- SPEC 문서 작성 (수동) → spec-builder 자동화
- TDD 구현 (수동) → tdd-implementer 자동화
- 문서 동기화 (수동) → doc-syncer 자동화
- TAG 체인 관리 (수동) → tag-agent 자동화

### 마이그레이션 고려사항

**기존 프로젝트를 MoAI-ADK로 마이그레이션할 경우**:

1. **프로젝트 분석** - `moai-adk init` 실행 시 기존 파일 감지 및 백업
2. **템플릿 병합** - 기존 `.claude/`, `.moai/` 디렉토리와 충돌 시 백업 생성
3. **SPEC 역추적** - 기존 코드에서 TAG 부재 시 수동 추가 필요

## TODO:STRUCTURE-001 구조 개선 계획

1. **MCP (Model Context Protocol) 통합** - Claude Desktop과의 직접 통합 지원
2. **플러그인 아키텍처** - 사용자 정의 에이전트/Skills 동적 로드
3. **분산 실행 지원** - 여러 에이전트 병렬 실행 최적화

## EARS 아키텍처 요구사항 작성법

### 구조 설계에서의 EARS 활용

아키텍처와 모듈 설계 시 EARS 구문을 활용하여 명확한 요구사항을 정의하세요:

#### 시스템 아키텍처 EARS 예시
```markdown
### Ubiquitous Requirements (아키텍처 기본 요구사항)
- 시스템은 계층형 아키텍처를 채택해야 한다
- 시스템은 모듈 간 느슨한 결합을 유지해야 한다

### Event-driven Requirements (이벤트 기반 구조)
- WHEN 외부 API 호출이 실패하면, 시스템은 fallback 로직을 실행해야 한다
- WHEN 데이터 변경 이벤트가 발생하면, 시스템은 관련 모듈에 통지해야 한다

### State-driven Requirements (상태 기반 구조)
- WHILE 시스템이 확장 모드일 때, 새로운 모듈을 동적으로 로드할 수 있어야 한다
- WHILE 개발 모드일 때, 시스템은 상세한 디버그 정보를 제공해야 한다

### Optional Features (선택적 구조)
- WHERE 클라우드 환경이면, 시스템은 분산 캐시를 활용할 수 있다
- WHERE 고성능이 요구되면, 시스템은 메모리 캐싱을 적용할 수 있다

### Constraints (구조적 제약사항)
- IF 보안 레벨이 높으면, 시스템은 모든 모듈 간 통신을 암호화해야 한다
- 각 모듈의 복잡도는 15를 초과하지 않아야 한다
```

---

_이 구조는 `/alfred:2-run` 실행 시 TDD 구현의 가이드라인이 됩니다._
