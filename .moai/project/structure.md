---
id: STRUCTURE-001
version: 0.1.1
status: active
created: 2025-10-01
updated: 2025-10-17
author: @Goos
priority: high
---

# MoAI-ADK Structure Design

## HISTORY

### v0.1.1 (2025-10-17)
- **UPDATED**: 템플릿 기본값을 실제 MoAI-ADK 아키텍처 내용으로 갱신
- **AUTHOR**: @Alfred
- **SECTIONS**: Architecture, Modules, Integration, Traceability 실제 구조 반영

### v0.1.0 (2025-10-01)
- **INITIAL**: 프로젝트 구조 설계 문서 작성
- **AUTHOR**: @architect
- **SECTIONS**: Architecture, Modules, Integration, Traceability

---

## @DOC:ARCHITECTURE-001 시스템 아키텍처

### 아키텍처 전략

**MoAI-ADK는 Python 기반 CLI + AI 에이전트 시스템으로 구성된 모듈형 아키텍처를 채택합니다.**

```
MoAI-ADK Architecture
├── CLI Layer                    # 사용자 인터페이스
│   ├── init                    # 프로젝트 초기화
│   ├── doctor                  # 환경 진단
│   ├── status                  # 프로젝트 상태 확인
│   ├── update                  # 템플릿 업데이트
│   ├── restore                 # 백업 복구
│   └── backup                  # 수동 백업
│
├── Agent Layer                  # AI 에이전트 시스템
│   ├── Alfred (Orchestrator)   # 중앙 조율자
│   ├── spec-builder            # SPEC 작성
│   ├── code-builder            # TDD 구현
│   ├── doc-syncer              # 문서 동기화
│   ├── tag-agent               # TAG 관리
│   ├── git-manager             # Git 워크플로우
│   ├── debug-helper            # 오류 진단
│   ├── trust-checker           # TRUST 검증
│   ├── cc-manager              # Claude Code 설정
│   └── project-manager         # 프로젝트 초기화
│
├── Core Layer                   # 핵심 비즈니스 로직
│   ├── git/                    # Git 워크플로우
│   ├── tags/                   # TAG 시스템
│   ├── templates/              # 템플릿 처리
│   ├── trust/                  # TRUST 검증
│   └── init/                   # 초기화 로직
│
└── Integration Layer            # 외부 시스템 통합
    ├── Claude Code Hooks       # SessionStart, PreToolUse, PostToolUse
    ├── GitHub API              # PR 생성, 브랜치 관리
    └── PyPI                    # 패키지 배포
```

**선택 이유**:
1. **관심사 분리**: CLI, Agent, Core, Integration 계층 명확히 분리
2. **확장성**: 새로운 에이전트/커맨드 추가 용이
3. **테스트 용이성**: 각 계층을 독립적으로 테스트 가능
4. **재사용성**: Core Layer는 CLI/Agent 계층과 독립적으로 재사용 가능

## @DOC:MODULES-001 모듈별 책임 구분

### 1. CLI Layer (`src/moai_adk/cli/`)

- **책임**: 사용자 명령어 처리 및 프로젝트 관리
- **입력**: CLI 명령어 (init, doctor, status, update, restore, backup)
- **처리**: 명령어별 핸들러 실행, 사용자 상호작용
- **출력**: Rich 포맷팅된 출력, 에러 메시지

| 컴포넌트 | 역할 | 주요 기능 |
|----------|------|-----------|
| `init.py` | 프로젝트 초기화 | 신규 프로젝트 생성, 기존 프로젝트에 설치 |
| `doctor.py` | 환경 진단 | 의존성 체크, 설정 검증, 권한 확인 |
| `status.py` | 상태 확인 | 프로젝트 정보, Git 상태, SPEC 진행도 표시 |
| `update.py` | 템플릿 업데이트 | Alfred 시스템 업데이트, 백업 생성 |
| `restore.py` | 백업 복구 | checkpoint/백업에서 복구 |
| `backup.py` | 수동 백업 | 현재 상태 백업 생성 |

### 2. Git Module (`src/moai_adk/git/`)

- **책임**: Git 워크플로우 자동화 및 checkpoint 관리
- **입력**: 브랜치명, 커밋 메시지, PR 정보
- **처리**: Git 명령 실행, GitHub API 호출, checkpoint 생성
- **출력**: Git 작업 결과, PR URL

| 컴포넌트 | 역할 | 주요 기능 |
|----------|------|-----------|
| `branch.py` | 브랜치 관리 | feature 브랜치 생성, develop 기반 분기 |
| `commit.py` | 커밋 관리 | TDD 단계별 커밋 (RED/GREEN/REFACTOR) |
| `checkpoint.py` | Checkpoint 시스템 | 자동 백업, FIFO 정리, 복구 |
| `manager.py` | Git 워크플로우 | PR 생성/머지, 브랜치 정리, 자동화 |

### 3. Tags Module (`src/moai_adk/tags/`)

- **책임**: @TAG 시스템 관리 및 추적성 검증
- **입력**: 소스 코드 (코드 스캔)
- **처리**: TAG 패턴 매칭, 체인 검증, 고아 TAG 탐지
- **출력**: TAG 목록, 무결성 검증 결과

| 컴포넌트 | 역할 | 주요 기능 |
|----------|------|-----------|
| `project.py` | 프로젝트 TAG 관리 | 프로젝트 레벨 TAG 생성 및 검증 |
| `context.py` | TAG 컨텍스트 | TAG 관계 그래프, 의존성 분석 |
| `tags.py` | TAG 스캔 | `rg` 기반 코드 직접 스캔 (CODE-FIRST) |

### 4. Templates Module (`src/moai_adk/templates/`)

- **책임**: 프로젝트 템플릿 처리 및 병합
- **입력**: 템플릿 파일, 사용자 설정
- **처리**: 템플릿 복사, 변수 치환, 지능형 병합
- **출력**: 프로젝트 파일 생성 및 업데이트

| 컴포넌트 | 역할 | 주요 기능 |
|----------|------|-----------|
| `processor.py` | 템플릿 처리 | 변수 치환, 파일 복사 |
| `merger.py` | BackupMerger | product/structure/tech.md 병합 |
| `backup.py` | 백업 관리 | Alfred 폴더 자동 백업 |
| `config.py` | 설정 관리 | .moai/config.json 생성/업데이트 |
| `languages/` | 언어별 템플릿 | Python, TypeScript, Java 등 17개 언어 |

### 5. Trust Module (`src/moai_adk/trust/`)

- **책임**: TRUST 5원칙 검증
- **입력**: 소스 코드, 테스트 커버리지, 린터 결과
- **처리**: 원칙별 검증 로직 실행
- **출력**: 준수 여부, 위반 항목

| 컴포넌트 | 역할 | 주요 기능 |
|----------|------|-----------|
| `trust_checker.py` | TRUST 검증 | 5가지 원칙 통합 검증 |
| `base_validator.py` | 기본 검증자 | 언어별 검증 로직 기반 클래스 |

### 6. Init Module (`src/moai_adk/init/`)

- **책임**: 프로젝트 초기화 및 환경 검증
- **입력**: 프로젝트 정보, 언어 스택
- **처리**: 의존성 체크, 설정 생성, 템플릿 적용
- **출력**: 초기화 완료 메시지

| 컴포넌트 | 역할 | 주요 기능 |
|----------|------|-----------|
| `checker.py` | 환경 체크 | Python, uv, git, rg 설치 확인 |
| `detector.py` | 프로젝트 감지 | 언어 스택 자동 감지 |
| `validator.py` | 설정 검증 | config.json 유효성 검증 |
| `initializer.py` | 초기화 실행 | 전체 초기화 오케스트레이션 |

### 7. Utils Module (`src/moai_adk/utils/`)

- **책임**: 공통 유틸리티 제공
- **입력**: 다양한 데이터 타입
- **처리**: 로깅, 배너 출력, 파일 I/O
- **출력**: 포맷팅된 출력

| 컴포넌트 | 역할 | 주요 기능 |
|----------|------|-----------|
| `logger.py` | 로깅 시스템 | Rich 기반 로깅 |
| `banner.py` | 배너 출력 | pyfiglet 기반 ASCII 아트 |

## @DOC:INTEGRATION-001 외부 시스템 통합

### Claude Code Hooks 연동

- **인증 방식**: Claude Code 내장 인증 (API 키 불필요)
- **데이터 교환**: JSON (stdin/stdout)
- **장애 시 대체**: Hooks 실패 시 에이전트 직접 호출
- **위험도**: Low (읽기 전용 데이터 제공, 차단 메시지만 반환)

**Hooks 구조**:
```
.claude/hooks/
├── session_start.py           # 세션 시작 시 프로젝트 정보 표시
├── pre_tool_use.py            # 위험 작업 차단, 자동 checkpoint
└── post_tool_use.py           # TAG 동기화, 컨텍스트 정리
```

**주요 기능**:
1. **SessionStart**: 프로젝트 정보, Git 상태, SPEC 진행도 표시
2. **PreToolUse**: 위험 작업 (`rm -rf`, 병합, 스크립트 실행) 차단, 자동 checkpoint
3. **PostToolUse**: TAG 동기화, 토큰 사용량 >70% 시 Compaction 권장

### GitHub API 연동

- **용도**: PR 생성, 브랜치 관리, 자동 머지
- **의존성 수준**: Medium (GitFlow 워크플로우 필수)
- **성능 요구사항**: 응답 시간 <5초

**주요 API**:
- `POST /repos/{owner}/{repo}/pulls`: Draft PR 생성
- `PATCH /repos/{owner}/{repo}/pulls/{number}`: PR Ready 전환
- `PUT /repos/{owner}/{repo}/pulls/{number}/merge`: 자동 머지 (squash)

**인증**:
- GitHub CLI (`gh`) 또는 Personal Access Token (PAT)
- 권한: `repo` 스코프 필수

### PyPI 연동

- **용도**: 패키지 배포 및 버전 관리
- **의존성 수준**: Low (개발자 전용 기능)
- **성능 요구사항**: N/A (비동기 배포)

**배포 프로세스**:
1. GitHub Actions 트리거 (tag push)
2. `hatch build` 실행 (wheel + sdist)
3. `twine upload` 실행 (PyPI)
4. 배포 완료 알림

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

**TAG 검증 명령어**:
```bash
# 전체 TAG 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# 특정 도메인 TAG 검색
rg '@SPEC:AUTH' -n .moai/specs/

# 특정 ID 검색 (중복 확인)
rg 'AUTH-001' -n

# 고아 TAG 탐지
rg '@CODE:AUTH-001' -n src/          # CODE는 있는데
rg '@SPEC:AUTH-001' -n .moai/specs/  # SPEC이 없으면 고아
```

### 디렉토리 구조 및 TAG 매핑

```
MoAI-ADK/
├── .moai/
│   ├── specs/
│   │   └── SPEC-{ID}/            # @SPEC:ID
│   │       ├── spec.md           # YAML Front Matter + EARS 요구사항
│   │       └── diagrams/         # (선택) 다이어그램
│   ├── project/
│   │   ├── product.md            # @DOC:MISSION-001, @SPEC:USER-001 등
│   │   ├── structure.md          # @DOC:ARCHITECTURE-001 등
│   │   └── tech.md               # @DOC:STACK-001 등
│   ├── memory/
│   │   ├── development-guide.md  # TRUST 5원칙, SPEC-TDD 워크플로우
│   │   └── spec-metadata.md      # SPEC 메타데이터 표준
│   └── config.json               # 프로젝트 설정
│
├── src/moai_adk/                 # @CODE:ID
│   ├── cli/                      # @CODE:CLI-*
│   ├── git/                      # @CODE:GIT-*
│   ├── tags/                     # @CODE:TAG-*
│   ├── templates/                # @CODE:TEMPLATE-*
│   ├── trust/                    # @CODE:TRUST-*
│   ├── init/                     # @CODE:INIT-*
│   └── utils/                    # @CODE:UTIL-*
│
├── tests/                        # @TEST:ID
│   ├── unit/                     # 단위 테스트
│   ├── integration/              # 통합 테스트
│   └── e2e/                      # E2E 테스트
│
├── docs/                         # @DOC:ID (Living Document)
│   ├── api/                      # API 문서
│   ├── guides/                   # 가이드
│   └── architecture/             # 아키텍처 문서
│
└── .claude/
    ├── commands/alfred/          # Alfred 커맨드 (1-spec, 2-build, 3-sync)
    ├── agents/                   # 9개 전문 에이전트
    └── hooks/                    # Hooks (session_start, pre_tool_use, post_tool_use)
```

## Legacy Context

### 기존 시스템 현황

**MoAI-ADK v0.3.1 (현재 버전)**

```
기존 구조/
├── src/moai_adk/              # 핵심 로직 완성
├── .claude/                   # Alfred 시스템 구축
├── tests/                     # 테스트 커버리지 87.66%
└── .moai/                     # SPEC 및 문서 체계
```

### 마이그레이션 고려사항

#### 현재 완료된 항목

1. **CLI Layer**: 6개 명령어 모두 구현 완료
2. **Agent Layer**: 10개 에이전트 (Alfred + 9개) 구현 완료
3. **Core Layer**: Git, Tags, Templates, Trust, Init 모듈 구현 완료
4. **Integration Layer**: Claude Code Hooks, GitHub API, PyPI 연동 완료

#### 개선 계획

1. **Template Processor 병합 로직 고도화** (v0.3.x)
   - 우선순위: high
   - 목표: 사용자 커스터마이징 보존율 100%

2. **Event-Driven Checkpoint 시스템 개선** (v0.3.x)
   - 우선순위: high
   - 목표: 위험 작업 탐지율 100%, 복구 성공률 100%

3. **언어별 TRUST 검증 규칙 확장** (v0.3.x)
   - 우선순위: medium
   - 목표: 17개 언어 모두 언어별 최적 도구 자동 선택

## TODO:STRUCTURE-001 구조 개선 계획

### 단기 (v0.3.x)

1. **모듈 간 인터페이스 정의**
   - 각 모듈의 public API 명확화
   - 타입 힌트 100% 적용 (mypy strict 모드)

2. **의존성 관리 전략**
   - 의존성 그래프 시각화
   - 순환 의존성 제거

3. **확장성 확보 방안**
   - Plugin 시스템 설계 (v0.5.x 목표)
   - 사용자 정의 에이전트 지원

### 중기 (v0.4.x)

4. **마이크로서비스 아키텍처 검토**
   - CLI, Agent, Core 계층 독립 배포 가능성 검토
   - API Gateway 도입 검토

5. **성능 최적화**
   - TAG 스캔 캐싱 (CODE-FIRST 원칙 유지)
   - 병렬 처리 (pytest-xdist 활용)

### 장기 (v0.5.x+)

6. **Web UI 대시보드**
   - 프로젝트 전체 현황 시각화
   - TAG 의존성 그래프 인터랙티브 뷰

7. **멀티 프로젝트 지원**
   - 여러 프로젝트 동시 관리
   - 프로젝트 간 SPEC 공유

## EARS 아키텍처 요구사항 작성법

### 구조 설계에서의 EARS 활용

아키텍처와 모듈 설계 시 EARS 구문을 활용하여 명확한 요구사항을 정의하세요:

#### 시스템 아키텍처 EARS 예시

```markdown
### Ubiquitous Requirements (아키텍처 기본 요구사항)
- 시스템은 모듈형 아키텍처를 채택해야 한다
- 시스템은 CLI, Agent, Core, Integration 계층을 명확히 분리해야 한다

### Event-driven Requirements (이벤트 기반 구조)
- WHEN 위험한 작업이 감지되면, 시스템은 자동으로 checkpoint를 생성해야 한다
- WHEN TAG 불일치가 발견되면, 시스템은 자동으로 수정을 권장해야 한다

### State-driven Requirements (상태 기반 구조)
- WHILE Personal 모드일 때, 시스템은 로컬 브랜치만 생성해야 한다
- WHILE Team 모드일 때, 시스템은 자동으로 PR을 생성해야 한다

### Optional Features (선택적 구조)
- WHERE GitHub API 접근 가능하면, 시스템은 자동 머지를 수행할 수 있다
- WHERE 네트워크 오프라인이면, 시스템은 로컬 모드로 동작할 수 있다

### Constraints (구조적 제약사항)
- IF 보안 레벨이 높으면, 시스템은 모든 외부 통신을 로깅해야 한다
- 각 모듈의 복잡도는 15를 초과하지 않아야 한다
- 함수는 50 LOC, 파일은 300 LOC를 초과하지 않아야 한다
```

---

_이 구조는 `/alfred:2-build` 실행 시 TDD 구현의 가이드라인이 됩니다._
