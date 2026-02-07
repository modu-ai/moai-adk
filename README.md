# MoAI-ADK (Go Edition)

**[English](README.en.md)** | **[한국어](README.md)**

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Copyleft%203.0-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-30%20packages-brightgreen)](./internal/)
[![Coverage](https://img.shields.io/badge/Coverage-85--100%25-brightgreen)](#test-coverage)
[![Version](https://img.shields.io/badge/Version-2.0.0-success)](CHANGELOG.md)

Claude Code를 위한 고성능 Agentic Development Kit -- Python 기반 MoAI-ADK(~73,000줄)을 Go로 완전히 재작성했습니다.

**모듈:** `github.com/modu-ai/moai-adk`

---

## 개요

MoAI-ADK (Go Edition)는 Claude Code 내에서 MoAI 프레임워크의 런타임 백본으로 작동하는 컴파일된 개발 툴킷입니다. CLI 도구, 구성 관리, LSP 통합, Git 작업, 품질 게이트, 자율 개발 루프 기능을 제공하며 모두 단일 바이너리로 배포되어 런타임 의존성이 없습니다.

### 왜 Go인가?

| 항목          | Python Edition                 | Go Edition                            |
| ------------- | ------------------------------ | ------------------------------------- |
| 배포          | pip install + venv + 의존성    | 단일 바이너리, 의존성 없음            |
| 시작 시간     | ~800ms 인터프리터 부팅         | ~5ms 네이티브 실행                    |
| 동시성        | asyncio / threading            | 네이티브 goroutines                   |
| 타입 안전성   | 런타임 (mypy 선택)             | 컴파일 타임 강제                      |
| 크로스 플랫폼 | Python 런타임 필요             | 모든 플랫폼용 프리빌트 바이너리       |
| 훅 실행       | Shell 래퍼 + Python 인터프리터 | 컴파일된 바이너리, 직접 JSON 프로토콜 |

### 핵심 특징

- **32,977줄**의 Go 코드, 30개 패키지
- **85-100% 테스트 커버리지** (30개 테스트 패키지)
- **네이티브 동시성** via goroutines (병렬 LSP, 품질 검사, Git 작업)
- **내장된 템플릿** using `go:embed`
- **크로스 플랫폼** 빌드 (macOS arm64/amd64, Linux arm64/amd64, Windows)
- **64개 moai-* 스킬** 프로그레시브 디스클로저 시스템으로 최적화
- **Agent Teams v2.0** (실험적) - 이중 모드 실행 엔진

---

## 새로운 기능 (v2.0)

### Agent Teams v2.0 (실험적 기능)

이중 모드 실행 엔진으로 서브 에이전트와 Agent Teams 모드를 모두 지원합니다:

- **Sub-agent 모드**: 순차적 전문 에이전트 위임 (기존 방식)
- **Agent Teams 모드**: 병렬 팀 기반 개발, 공유 작업 목록
- **Auto 모드**: 복잡도 분석을 통한 지능형 모드 선택

```
User Request
    |
    v
MoAI Orchestrator
    |
    +-- Mode Selector
        |
        +-- Sub-Agent Mode (Task() 단일 에이전트)
        +-- Agent Teams Mode (TeamCreate + SendMessage)
        +-- Auto Mode (복잡도 기반 자동 선택)
```

**팀 구성 패턴:**
- Plan 팀: researcher + analyst + architect (병렬 탐색)
- Run 팀: backend-dev + frontend-dev + tester (병렬 구현)
- Sync 페이즈: 항상 sub-agent 모드

**활성화 방법:**
1. `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` 환경 변수 설정
2. `workflow.team.enabled: true` 설정
3. `--team` 또는 `--auto` 플래그 사용

자세한 내용은 [workflow-v2.md](.moai/docs/workflow-v2.md)를 참조하세요.

### 자동 업데이트 및 캐싱

- **자동 업데이트 알림**: 세션 시작 시 자동으로 새 버전 확인
- **업데이트 결과 캐싱**: 불필요한 API 호출 최소화
- **롤백 지원**: 업데이트 실패 시 자동 복구

### StatusLine 개선

- **개선된 버전 표시**: 더 명확한 버전 정보 렌더링
- **확장된 테스트 커버리지**: 안정성 향상

### Hybrid 방법론 (기본값)

신규 프로젝트와 지속적 개발을 위한 권장 방법론:
- **새 코드**: TDD (RED-GREEN-REFACTOR)
- **기존 코드**: DDD (ANALYZE-PRESERVE-IMPROVE)

### 64개 moai-* 스킬

프로그레시브 디스클로저 시스템으로 최적화된 포괄적인 스킬 컬렉션:

**Foundation Skills:**
- moai-foundation-claude, moai-foundation-core, moai-foundation-philosopher, moai-foundation-quality, moai-foundation-context

**Workflow Skills:**
- moai-workflow-spec, moai-workflow-project, moai-workflow-jit-docs, moai-workflow-templates, moai-workloop, moai-workflow-ddd, moai-workflow-tdd, moai-workflow-testing, moai-workflow-worktree, moai-workflow-thinking

**Domain Skills:**
- moai-domain-backend, moai-domain-frontend, moai-domain-database, moai-domain-uiux, moai-formats-data

**Language Skills (18개 언어):**
- Go, Python, TypeScript, JavaScript, Java, Kotlin, Swift, C++, C#, Ruby, PHP, Rust, Elixir, Scala, R, Flutter

**Library Skills:**
- moai-library-shadcn, moai-library-nextra, moai-library-mermaid

**Platform Skills:**
- Vercel, Supabase, Neon, Convex, Firebase, Firestore, Auth0, Clerk, Railway

**Tool Skills:**
- moai-tool-ast-grep, moai-tool-svg

**Specialist Skills:**
- Figma integration, Flutter development, Rust engineering, Performance optimization, Guidelines

---

## 설치

### 빠른 설치 (권장)

간단한 원라인 명령어로 설치하세요. OS와 아키텍처를 자동 감지하여 적합한 바이너리를 다운로드하고 설치합니다.

#### macOS, Linux, WSL

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.sh | bash
```

#### Windows PowerShell

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.ps1 | iex
```

또는:

```powershell
Invoke-Expression (Invoke-RestMethod -Uri "https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.ps1")
```

#### Windows CMD

```batch
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.bat -o install.bat
install.bat
```

curl이 없는 경우:
```batch
powershell -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.bat' -OutFile 'install.bat'"
install.bat
```

### 소스에서 빌드

Go 1.25 이상이 필요합니다.

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
make build
```

컴파일된 바이너리는 `bin/moai`에 생성됩니다.

### GOPATH에 설치

```bash
make install
```

### 프리빌트 바이너리

[릴리즈](https://github.com/modu-ai/moai-adk/releases) 페이지에서 플랫폼별 바이너리를 다운로드하세요. 다음 플랫폼용 아카이브가 제공됩니다:

- `darwin_arm64` (macOS Apple Silicon)
- `darwin_amd64` (macOS Intel)
- `linux_arm64`
- `linux_amd64`
- `windows_amd64`

---

## 빠른 시작

### 프로젝트 초기화

```bash
moai init my-project-name
```

언어, 프레임워크, 방법론을 자동 감지하는 대화형 프로젝트 설정 마법사를 실행하여 적절한 구성과 Claude Code 통합 파일을 생성합니다. 기본 방법론은 **Hybrid**입니다.

### 시스템 상태 확인

```bash
moai doctor
```

개발 환경을 진단하여 도구 가용성, 구성 무결성, LSP 서버 준비 상태를 확인합니다.

### 프로젝트 상태 확인

```bash
moai status
```

Git 브랜치, 품질 메트릭, 구성 상태를 포함한 프로젝트 상태 요약을 표시합니다.

### 업데이트 확인

```bash
# 업데이트 확인만 수행
moai update --check

# 최신 버전으로 프로젝트 템플릿 동기화
moai update --project
```

### Git Worktree 관리

```bash
moai worktree new feature/auth
moai worktree list
moai worktree switch feature/auth
moai worktree sync
moai worktree remove feature/auth
moai worktree clean
```

병렬 브랜치 개발을 위한 전체 worktree 수명 주기 관리.

---

## 아키텍처

```
moai-adk/
+-- cmd/moai/             # 애플리케이션 진입점
|   +-- main.go
+-- internal/             # 프라이빗 애플리케이션 패키지
|   +-- astgrep/          # AST 기반 코드 분석
|   +-- cli/              # Cobra 명령어 정의
|   +-- config/           # YAML 구성 관리
|   +-- core/
|   |   +-- git/          # Git 작업
|   |   +-- project/      # 프로젝트 초기화 및 감지
|   |   +-- quality/      # TRUST 5 품질 게이트
|   +-- foundation/       # EARS 패턴, TRUST 5, 언어 정의
|   +-- git/              # Git 작업 래퍼
|   +-- hook/             # Claude Code 훅 시스템
|   |   +-- lifecycle/    # 훅 생명주기 관리
|   |   +-- quality/      # 훅 기반 품질 검사
|   |   +-- security/     # 훅 기반 보안 스캐닝
|   +-- loop/             # Ralph 피드백 루프 및 상태 머신
|   +-- lsp/              # LSP 클라이언트 (16개 이상의 언어)
|   |   +-- hook/         # LSP 훅 통합
|   +-- manifest/         # 파일 출적 추적 (SHA-256)
|   +-- merge/            # 3-way 병합 엔진
|   +-- ralph/            # 수렴 결정 엔진
|   +-- rank/             # 세션 랭킹 (HMAC-SHA256)
|   +-- resilience/       # 회복성 패턴 (회로 차단기, 재시도)
|   +-- shell/            # 셸 실행 및 환경 감지
|   +-- statusline/       # Claude Code 상태줄 통합
|   +-- template/         # 템플릿 배포 및 보안
|   +-- ui/               # 대화형 TUI 컴포넌트
|   +-- update/           # 자체 업데이트 및 롤백
+-- pkg/                  # 퍼블릭 라이브러리 패키지
|   +-- models/           # 공유 데이터 모델
|   +-- utils/            # 공통 유틸리티
|   +-- version/          # 빌드 버전 메타데이터
+-- templates/            # 내장된 프로젝트 템플릿
+-- Makefile              # 빌드 자동화
+-- .goreleaser.yml       # 릴리즈 구성
```

### 패키지 개요

| 패키지         | 목욱                                                              | 커버리지 |
| -------------- | ----------------------------------------------------------------- | -------- |
| `config`       | 스레드 안전 병렬 액세스가 가능한 모듈식 YAML 구성                 | 94.1%    |
| `foundation`   | EARS 패턴, TRUST 5 원칙, 18개 언어 정의, 방법론 엔진              | 98.4%    |
| `core/git`     | Git 작업 (브랜치, worktree, 충돌, 이벤트 감지)                    | 88.1%    |
| `core/project` | 프로젝트 초기화, 언어/프레임워크 감지, 방법론 자동 감지           | 89.2%    |
| `core/quality` | 병렬 검증기 및 페이즈 게이트가 있는 TRUST 5 품질 게이트           | 96.8%    |
| `hook`         | Claude Code용 컴파일된 훅 시스템 (6개 이벤트 타입, JSON 프로토콜) | 90.0%    |
| `lsp`          | 16개 이상의 언어를 지원하는 LSP 클라이언트, 병렬 서버 관리        | 91.3%    |
| `template`     | 템플릿 배포, 설정 생성, 경로 보안                                 | 85.7%    |
| `manifest`     | SHA-256 무결성 검증이 있는 파일 출적 추적                         | 88.0%    |
| `ui`           | 대화형 TUI (선택기, 체크박스, 프롬프트, 진행률, 마법사)           | 96.8%    |
| `statusline`   | git/메모리/품질 메트릭이 있는 Claude Code 상태줄                  | 100%     |
| `astgrep`      | AST 기반 코드 분석 및 패턴 매칭                                   | 89.4%    |
| `rank`         | HMAC-SHA256 인증이 있는 세션 랭킹                                 | 85.1%    |
| `update`       | SHA-256 검증 및 자동 롤백이 있는 자체 업데이트                    | 87.6%    |
| `merge`        | 6가지 전략 및 충돌 마커가 있는 3-way 병합 엔진                    | 90.3%    |
| `loop`         | 상태 머신 및 수렴 감지가 있는 Ralph 피드백 루프                   | 92.7%    |
| `ralph`        | 자율 반복을 위한 수렴 결정 엔진                                   | 100%     |
| `cli`          | Cobra 명령어 (init, doctor, status, version, worktree)            | 92.0%    |
| `cli/worktree` | Git worktree 하위 명령어 (new, list, switch, sync, remove, clean) | 100%     |

### 핵심 개념

**TRUST 5 품질 프레임워크** -- 모든 코드 변경은 5가지 핵심 요소에 대해 검증됩니다:

- **Tested**: 85%+ 커버리지, 기존 코드용 특성 테스트
- **Readable**: 명확한 명명 규칙, 일관된 코드 스타일
- **Unified**: 일관된 포맷팅, 가져오기 순서
- **Secured**: OWASP 준수, 입력 유효성 검사
- **Trackable**: 컨벤셔널 커밋, 이슈 참조

**훅 실행 계약** -- 컴파일된 바이너리 훅이 shell 래퍼를 대체하여 JSON 프로토콜을 통해 6가지 Claude Code 이벤트 타입(PreToolUse, PostToolUse, SessionStart, SessionEnd, PreCompact, Notification)을 지원합니다. 프로토콜 준수를 위해 모든 훅 출력은 `hookSpecificOutput`에 `hookEventName` 필드를 포함해야 합니다.

**터치리스 템플릿 업데이트** -- 파일 출적 추적이 있는 3-way 병합 엔진을 통해 사용자 정의를 잃지 않고 자동 템플릿 업데이트가 가능합니다.

---

## CLI 명령어

| 명령어                        | 설명                                               |
| ----------------------------- | -------------------------------------------------- |
| `moai init`                   | 언어/프레임워크 감지가 포함된 대화형 프로젝트 설정 |
| `moai doctor`                 | 시스템 상태 진단 및 환경 검증                      |
| `moai status`                 | git 및 품질 메트릭이 포함된 프로젝트 상태 개요     |
| `moai version`                | 버전, 커밋 해시, 빌드 날짜 정보                    |
| `moai hook <event>`           | Claude Code 통합용 훅 디스패처                     |
| `moai worktree new <name>`    | 새 Git worktree 생성                               |
| `moai worktree list`          | 활성 worktree 목록                                 |
| `moai worktree switch <name>` | 기존 worktree로 전환                               |
| `moai worktree sync`          | 업스트림과 worktree 동기화                         |
| `moai worktree remove <name>` | worktree 제거                                      |
| `moai worktree clean`         | 오래된 worktree 정리                               |
| `moai update`                 | 최신 버전으로 업데이트 (자동 롤백 포함)            |
| `moai update --check`         | 설치하지 않고 업데이트 확인                        |
| `moai update --project`       | 바이너리 업데이트 없이 프로젝트 템플릿 동기화      |

### 업데이트 명령어

`moai update` 명령어는 최신 릴리스를 확인하고 설치합니다. 다음을 지원합니다:

- **Dev 버전**: `go-v*` 태그가 지정된 릴리스 자동 확인 (Go 에디션)
- **Production 버전**: 최신 안정 릴리스 확인
- **환경 변수 재정의**: `MOAI_UPDATE_URL`을 사용하여 다른 저장소 확인

```bash
# 업데이트 확인
moai update --check

# 최신 버전으로 업데이트
moai update

# 프로젝트 템플릿만 동기화 (바이너리 업데이트 없음)
moai update --project

# 사용자 정의 저장소 사용 (환경 변수)
export MOAI_UPDATE_URL="https://api.github.com/repos/owner/repo/releases/latest"
moai update
```

#### 자동 업데이트 알림

세션 시작 시 훅을 통해 자동으로 새 버전을 확인합니다. 업데이트가 가능한 경우 상태줄에 알림이 표시됩니다. 업데이트 확인 결과는 캐시되어 불필요한 API 호출을 줄입니다.

#### 릴리즈 태깅

릴리스 태그는 `v*` 형식을 사용합니다. 설치 스크립트는 `v*`와 `go-v*` 태그를 모두 인식합니다:

```bash
# 릴리스 태그 (권장)
git tag v2.0.0
git push origin v2.0.0
```

설치 스크립트는 태그에서 접두사를 자동으로 제거하여 올바른 다운로드 URL을 생성합니다.

---

## 개발

### 전제 조건

- Go 1.25 이상
- `golangci-lint` (린팅용)
- `gofumpt` (포맷팅용)

### 빌드

```bash
# bin/moai에 바이너리 빌드
make build

# git 태그의 버전 정보로 빌드
make build VERSION=v1.0.0

# 빌드 및 실행
make run
```

### Makefile 타겟

| 타겟                | 설명                                  |
| ------------------- | ------------------------------------- |
| `make build`        | `bin/moai`에 바이너리 빌드            |
| `make install`      | `$GOPATH/bin`에 바이너리 설치         |
| `make test`         | 레이스 감지 및 커버리지로 테스트 실행 |
| `make test-verbose` | 상세 출력으로 테스트 실행             |
| `make coverage`     | HTML 커버리지 리포트 생성             |
| `make lint`         | golangci-lint 실행                    |
| `make vet`          | go vet 실행                           |
| `make fmt`          | gofumpt로 코드 포맷팅                 |
| `make tidy`         | go 모듈 정리                          |
| `make clean`        | 빌드 아티팩트 제거                    |
| `make all`          | 린트, 테스트, 빌드 실행               |

### 빌드 플래그

버전 메타데이터는 `-ldflags`를 통해 빌드 타임에 주입됩니다:

```bash
go build -ldflags "-s -w \
  -X github.com/modu-ai/moai-adk/pkg/version.Version=v1.0.0 \
  -X github.com/modu-ai/moai-adk/pkg/version.Commit=$(git rev-parse --short HEAD) \
  -X github.com/modu-ai/moai-adk/pkg/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o bin/moai ./cmd/moai
```

---

## 테스트

### 모든 테스트 실행

```bash
# 레이스 감지가 있는 표준 실행
go test -race ./... -count=1

# Makefile을 통한 실행 (커버리지 출력 포함)
make test
```

### 테스트 커버리지

HTML 커버리지 리포트 생성:

```bash
make coverage
# coverage.html 열기
```

### 패키지별 테스트 커버리지

| 패키지         | 커버리지 |
| -------------- | -------- |
| `config`       | 94.1%    |
| `foundation`   | 98.4%    |
| `core/quality` | 96.8%    |
| `ui`           | 96.8%    |
| `loop`         | 92.7%    |
| `cli`          | 92.0%    |
| `lsp`          | 91.3%    |
| `merge`        | 90.3%    |
| `hook`         | 90.0%    |
| `astgrep`      | 89.4%    |
| `core/project` | 89.2%    |
| `manifest`     | 88.0%    |
| `core/git`     | 88.1%    |
| `update`       | 87.6%    |
| `template`     | 85.7%    |
| `rank`         | 85.1%    |
| `ralph`        | 100%     |
| `statusline`   | 100%     |
| `cli/worktree` | 100%     |

### 개발 방법론

프로젝트는 **Hybrid** 방법론을 기본으로 사용합니다:

- **Hybrid (기본값)**: 신규 프로젝트 및 지속적 개발에 권장. 새 코드는 TDD, 기존 코드는 DDD 적용
- **DDD (Domain-Driven Development)**: 기존/브라운필드 프로젝트 전용. 기존 동작 분석(ANALYZE), 특성 테스트로 보존(PRESERVE), 점진적 개선(IMPROVE)
- **TDD (Test-Driven Development)**: 격리된 신규 모듈 전용. 실패하는 테스트 작성(RED), 통과 구현(GREEN), 리팩토링(REFACTOR)

방법론은 `moai init` 시 프로젝트 상태에 따라 자동 선택되며, `.moai/config/sections/quality.yaml`에서 수동 변경 가능합니다.

---

## 릴리즈

릴리스는 [GoReleaser](https://goreleaser.com/)로 자동화됩니다. 각 릴리스는 다음을 생성합니다:

- 지원되는 모든 플랫폼용 정적으로 연결된 바이너리 (`CGO_ENABLED=0`)
- `tar.gz` 아카이브 (Linux, macOS) 및 `zip` 아카이브 (Windows)
- `checksums.txt`의 SHA-256 체크섬

---

## 기여

기여를 원하시나요? [CONTRIBUTING.ko.md](CONTRIBUTING.ko.md)에서 상세한 가이드를 확인하세요.

### 빠른 시작

1. 저장소를 포크하세요
2. 기능 브랜치 생성 (`git checkout -b feature/my-feature`)
3. 테스트 우선 작성 (새 코드는 TDD, 기존 코드는 특성 테스트)
4. 모든 테스트 통과 확인: `make test`
5. 린팅 통과 확인: `make lint`
6. 코드 포맷팅: `make fmt`
7. 컨벤셔널 커밋 메시지로 커밋
8. 풀 리퀘스트 오픈

### 코드 품질 요구사항

- 모든 패키지는 85%+ 테스트 커버리지 유지
- 0개의 린트 오류, 0개의 타입 오류
- 기존 패키지 구조 및 명명 규칙 준수
- 적절한 경우 테이블 기반 테스트 포함

### 개발 참여자

모든 기여자들을 환영합니다! 작은 버그 수정부터 대규모 기능 추가까지 모든 기여는 소중합니다.

---

## 라이선스

Copyleft 3.0 - 자세한 내용은 [LICENSE](LICENSE)를 참조하세요.

---

## 관련 프로젝트

- [MoAI-ADK (Python)](https://github.com/modu-ai/moai-adk) -- 원본 Python 구현 (~73,000줄)
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) -- MoAI-ADK가 확장하는 AI 개발 환경
