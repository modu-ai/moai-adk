# MoAI CLI 비교 분석 보고서: moai-adk (Python) vs moai-go (Go)

Version: 1.0.0
Date: 2026-02-03
Status: DRAFT
Cross-references: design.md, template-redesign-plan.md, product.md, structure.md, tech.md

---

## 1. 개요 (Executive Summary)

본 보고서는 Python 기반 MoAI-ADK CLI와 Go 기반 MoAI-GO CLI의 포괄적 비교 분석을 제공한다. 두 구현체의 아키텍처, 커맨드 구조, 성능, 의존성, 테스트 인프라를 체계적으로 비교하여 Go 재작성의 타당성과 마이그레이션 전략을 도출한다.

### 1.1 프로젝트 요약

| 항목 | moai-adk (Python) | moai-go (Go) |
|------|-------------------|--------------|
| 코드베이스 규모 | ~88K LOC (220+ 파일) | ~18K LOC 목표 (현재 ~306 LOC) |
| CLI 프레임워크 | Click >=8.1.0 | Cobra v1.10.2 |
| Git 커밋 수 | 4,174 | 초기 단계 |
| 알려진 이슈 | 173 (2 open, 171 closed) | N/A (신규) |
| 구현 진행률 | 100% | ~3% (scaffold + version) |

### 1.2 핵심 발견

1. **68% CLI LOC 감소 달성 가능**: Python CLI ~6,660 LOC에서 Go ~2,130 LOC으로 축소
2. **28개 Hook 이슈 완전 해결**: 컴파일된 binary subcommand로 전환 (ADR-006)
3. **90% 시작 성능 향상**: 200-500ms에서 <50ms로 단축
4. **의존성 60% 감소**: 직접 의존성 ~15개에서 ~8개로, 전이적 의존성 ~50+에서 ~20개로
5. **단일 바이너리 배포**: Python 런타임 + 가상환경 의존성 완전 제거

### 1.3 현재 Go 구현 상태

현재 Go scaffold에는 다음이 구현되어 있다:

- `cmd/moai/main.go`: 14 LOC, `cli.Execute()` 호출하는 thin entry point
- `internal/cli/root.go`: 30 LOC, Cobra root command 정의
- `internal/cli/init.go`: 23 LOC, stub (TODO)
- `internal/cli/doctor.go`: 23 LOC, stub (TODO)
- `internal/cli/status.go`: 23 LOC, stub (TODO)
- `internal/cli/version.go`: 27 LOC, 완전 구현 (build-time ldflags 주입)
- `pkg/version/version.go`: 31 LOC, version 상수 및 접근 함수
- `pkg/models/`: 46 LOC, 기본 config/project 타입 정의
- `pkg/utils/`: logger + path 유틸리티

---

## 2. CLI 프레임워크 비교

### 2.1 Python: Click >=8.1.0

Click은 Python의 대표적 CLI 프레임워크로, decorator 기반 커맨드 등록과 자동 help 생성을 제공한다.

**주요 특성:**

- **Decorator 기반 등록**: `@click.command()`, `@click.group()`, `@click.option()` 데코레이터로 커맨드 정의
- **LazyGroup 패턴**: 시작 시간 최적화를 위한 지연 로딩 (`_load_rank_group()`, `_load_worktree_group()`)
- **Rich 통합**: `rich.console.Console`을 통한 고급 터미널 출력 (스타일링, 패널, 테이블)
- **InquirerPy 통합**: 대화형 프롬프트 (선택, 체크박스, 텍스트 입력)
- **click.testing.CliRunner**: In-process CLI 테스트 지원
- **Windows UTF-8 보정**: `sys.stdout.reconfigure(encoding="utf-8")` 수동 처리 필요

**Python CLI Entry Point 구조** (`__main__.py`, 339 LOC):

```python
@click.group(invoke_without_command=True)
@click.version_option(version=__version__, prog_name="MoAI-ADK")
@click.pass_context
def cli(ctx: click.Context) -> None:
    if ctx.invoked_subcommand is None:
        show_logo()

# Lazy-loaded commands
@cli.command()
@click.argument("path", type=click.Path(), default=".")
@click.option("--non-interactive", "-y", is_flag=True)
def init(ctx, path, non_interactive, ...):
    from moai_adk.cli.commands.init import init as _init
    ctx.invoke(_init, ...)
```

### 2.2 Go: Cobra v1.10.2

Cobra는 Go 생태계의 표준 CLI 프레임워크로, struct 기반 커맨드 정의와 네이티브 flag 파싱을 제공한다.

**주요 특성:**

- **Struct 기반 정의**: `cobra.Command{}` 구조체로 커맨드 정의
- **내장 Help 생성**: 자동 help 텍스트 + 사용법 생성 (Python보다 풍부)
- **네이티브 Shell Completion**: bash, zsh, fish, PowerShell 완성 내장
- **pflag 파싱**: POSIX 호환 flag 파싱 (`--flag` 및 `-f` 지원)
- **표준 testing 패키지**: 별도 테스트 러너 불필요
- **Lazy Loading 불필요**: 컴파일된 바이너리이므로 import 오버헤드 없음

**Go CLI Entry Point 구조** (`root.go`, 30 LOC):

```go
var rootCmd = &cobra.Command{
    Use:   "moai",
    Short: "MoAI-ADK: Agentic Development Kit for Claude Code",
    Version: version.GetVersion(),
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    rootCmd.SetVersionTemplate(fmt.Sprintf("moai-adk %s\n", version.GetVersion()))
}
```

### 2.3 프레임워크 비교표

| 차원 | Click (Python) | Cobra (Go) | 분석 |
|------|---------------|------------|------|
| 커맨드 등록 | Decorator (`@click.command`) | Struct (`cobra.Command{}`) | Go가 더 명시적, IDE 지원 우수 |
| 서브커맨드 | `@group.command()` | `AddCommand()` | 유사한 패턴 |
| Flag 파싱 | `click.option` / `click.argument` | `pflag` | Cobra가 POSIX 호환성 우수 |
| Help 생성 | 자동 (기본적) | 자동 (상세, 서브커맨드 트리) | Cobra가 더 풍부 |
| Shell Completion | 수동 / 3rd party | 내장 (bash/zsh/fish/PS) | Cobra 압도적 우위 |
| 테스팅 | `CliRunner` (in-process) | 직접 커맨드 실행 | 동등 |
| Lazy Loading | `LazyGroup` 패턴 필요 | N/A (컴파일) | Go는 근본적 불필요 |
| 타입 안전성 | 런타임 (click.Choice) | 컴파일 타임 | Go가 안전 |
| 에러 처리 | Exception + Abort | `error` 반환 | Go가 명시적 |
| 플러그인 확장 | `click.MultiCommand` | `cobra.Command` 동적 추가 | 유사한 수준 |
| 터미널 UI | Rich + InquirerPy (별도) | Charmbracelet (별도) | 양쪽 모두 강력 |
| 시작 시간 | 100-300ms (import) | <5ms (링크) | Go 20-60x 빠름 |

---

## 3. 커맨드 매핑 분석

### 3.1 커맨드 인벤토리 비교표

Python CLI에서 Go CLI로의 전체 커맨드별 매핑:

| Python 커맨드 | Python LOC | Go 타겟 파일 | Go 예상 LOC | Phase | 현재 상태 |
|--------------|-----------|-------------|-------------|-------|----------|
| `init` | 854 | `internal/cli/init.go` | ~300 | Phase 1 | Stub |
| `doctor` | 370 | `internal/cli/doctor.go` | ~150 | Phase 1 | Stub |
| `status` | 111 | `internal/cli/status.go` | ~80 | Phase 2 | Stub |
| `version` | ~50 (entry) | `internal/cli/version.go` | ~27 | Phase 1 | **Done** |
| `update` | 3,162 | `internal/cli/update.go` | ~800 | Phase 3 | Not started |
| `cc` / `glm` (백엔드 전환) | 325 | `internal/cli/cc.go`, `glm.go` | ~120 | Phase 3 | Not started |
| `rank` (7 서브커맨드) | 478 | `internal/cli/rank/*.go` | ~200 | Phase 4 | Not started |
| `worktree` (10 서브커맨드) | 885+662+422+65 = 2,034 | `internal/cli/worktree/*.go` | ~400 | Phase 5 | Not started |
| `statusline` | 420 (+ 2,877 지원 모듈) | `internal/cli/statusline.go` | ~40 | Phase 2 | Not started |
| `language` | 255 | **DROPPED** | 0 | - | N/A |
| `analyze` | 125 | **DROPPED** | 0 | - | N/A |
| `__main__.py` (entry) | 339 | `cmd/moai/main.go` + `root.go` | ~44 | Phase 1 | Done |
| **합계** | **~6,523** (순수 CLI) | | **~2,161** | | **67% 감소** |

**참고**: Python LOC은 `moai_adk/cli/` 디렉토리의 순수 CLI 코드만 포함. 지원 모듈(statusline 3,297 LOC, core 모듈 등)은 별도 집계.

### 3.2 각 커맨드 상세 비교

#### 3.2.1 init -- 프로젝트 초기화

**Python 구현 (854 LOC)**

```
moai-adk init [PATH] [OPTIONS]
Options:
  --non-interactive, -y    Non-interactive mode (기본값 사용)
  --mode [personal|team]   프로젝트 모드 (default: personal)
  --locale [ko|en|ja|zh]   언어 설정
  --language TEXT           프로그래밍 언어 (자동 감지)
  --force                  확인 없이 재초기화
```

- InquirerPy 대화형 프롬프트 (프로젝트 이름, 언어, 타입 선택)
- Git 감지 및 Python 버전 확인
- 템플릿 배포 (`.moai/` 디렉토리 구조, `.claude/` 템플릿, `CLAUDE.md`)
- 설정 파일 생성 (user.yaml, language.yaml, quality.yaml)
- Rich 콘솔 출력 (진행 상황, 결과 패널)
- Windows UTF-8 인코딩 보정

**Go 타겟 설계 (~300 LOC)**

- Charmbracelet bubbletea 기반 대화형 wizard (Elm architecture)
- `internal/core/project/initializer.go` 위임
- `internal/template/deployer.go`로 go:embed 템플릿 배포
- `internal/manifest/manifest.go`로 파일 추적 시작 (ADR-007)
- `internal/config/manager.go`로 Viper 기반 설정 생성
- TTY 감지를 통한 non-interactive 환경 자동 처리

**마이그레이션 복잡도**: Medium

- 핵심 비즈니스 로직은 직관적이나 템플릿 배포 + 매니페스트 추적이 새로운 패턴
- InquirerPy -> bubbletea 전환에 TUI 재설계 필요

**해결되는 이슈**: #315 (config path 혼란), #283 (config 미로드), #304 (변수 치환 실패), #2 (non-interactive 실패), #9 (Git Bash 중단)

#### 3.2.2 doctor -- 시스템 진단

**Python 구현 (370 LOC)**

```
moai-adk doctor [OPTIONS]
Options:
  --verbose, -v           상세 도구 버전 및 언어 감지 표시
  --fix                   누락 도구 수정 제안
  --export PATH           진단 결과 JSON 내보내기
  --check TEXT            특정 도구만 점검
  --check-commands        Slash command 로딩 문제 진단
  --shell                 Shell 및 PATH 설정 진단 (WSL/Linux)
```

- 시스템 진단: Git, Python, Claude Code, Node.js, PATH, config 무결성
- Rich 패널 형식 출력 (성공/실패 상태 아이콘)
- `--fix` 옵션으로 자동 수정 제안
- JSON 내보내기 지원

**Go 타겟 설계 (~150 LOC)**

- Cobra command + flag 정의
- `internal/core/project/checker.go` 위임
- go-git로 Git 상태 확인 (subprocess 대신 in-process)
- lipgloss 스타일링 출력
- 바이너리 자체 검증 추가 (hook contract 검사)

**마이그레이션 복잡도**: Low

- 진단 로직은 주로 외부 도구 존재 확인으로 직관적
- Go에서는 `exec.LookPath()`로 간결하게 구현 가능

**해결되는 이슈**: Hook 시스템 진단 항목이 근본적으로 불필요해짐 (컴파일된 바이너리)

#### 3.2.3 status -- 프로젝트 상태

**Python 구현 (111 LOC)**

```
moai-adk status
```

- Rich 패널 디스플레이
- 프로젝트 이름, 타입, MoAI 버전, config 상태 표시
- SPEC 진행 상황 요약

**Go 타겟 설계 (~80 LOC)**

- Cobra command 정의
- `internal/config/manager.go`에서 config 읽기
- `internal/core/git/manager.go`에서 Git 상태
- lipgloss 스타일 패널 출력

**마이그레이션 복잡도**: Low -- 단순 표시 커맨드

#### 3.2.4 version -- 버전 표시

**Python 구현 (~50 LOC, entry point 내)**

```
moai-adk --version
```

- `click.version_option()`으로 구현
- `moai_adk/__init__.py`에서 `__version__` 읽기
- pyfiglet 로고 표시 (서브커맨드 없을 때)

**Go 구현 (27 LOC) -- 완료**

```
moai version
moai --version
```

- Cobra 내장 `--version` flag + `version` 서브커맨드
- Build-time `ldflags` 주입: Version, Commit, Date
- `pkg/version/version.go`에서 관리

**마이그레이션 복잡도**: Done -- 이미 구현 완료

**개선점**: Go는 commit hash와 build date까지 포함하여 더 풍부한 버전 정보 제공

#### 3.2.5 update -- 업데이트 시스템

**Python 구현 (3,162 LOC) -- 가장 복잡한 커맨드**

```
moai-adk update [OPTIONS]
Options:
  --path PATH            프로젝트 경로 (default: .)
  --force                백업 건너뛰기, 강제 업데이트
  --check                버전만 확인 (업데이트 안 함)
  --templates-only       패키지 업그레이드 건너뛰기, 템플릿만 동기화
  --yes                  모든 확인 자동 수락 (CI/CD 모드)
  -c, --config           프로젝트 설정 편집 (init wizard 동일)
```

- PyPI/GitHub 버전 확인
- 패키지 매니저 업그레이드 (pip/uv/pipx)
- 템플릿 다운로드 + 파일 비교
- 문자열 기반 템플릿 변수 치환 (`{{HOOK_SHELL_PREFIX}}`, `${SHELL:-/bin/bash}`)
- 백업/복원 메커니즘
- 마이그레이션 실행 (버전별 migration 스크립트)
- **38회 수정된 가장 불안정한 파일**

**Go 타겟 설계 (~800 LOC) -- 근본적 재설계**

- **바이너리 Self-Update**: goreleaser + GitHub Releases API
  - `internal/update/checker.go`: 최신 릴리스 확인
  - `internal/update/updater.go`: 바이너리 다운로드 + 원자적 교체
  - `internal/update/rollback.go`: 실패 시 자동 롤백
  - `internal/update/orchestrator.go`: 전체 워크플로우 조율
- **Manifest 기반 템플릿 동기화** (ADR-007):
  - `internal/manifest/manifest.go`: 파일 provenance 추적
  - 4가지 provenance 유형: `template_managed`, `user_modified`, `user_created`, `deprecated`
- **3-Way Merge 엔진**:
  - `internal/merge/three_way.go`: Git 스타일 3-way merge
  - 파일 유형별 전략: line merge, YAML deep merge, JSON merge, section merge
  - 충돌 시 `.conflict` 파일 생성 + 사용자 알림
- **Struct 직렬화** (ADR-011): 문자열 치환 버그 원천 차단

**마이그레이션 복잡도**: **High** -- 완전 재설계 필요

**해결되는 이슈**: #246 (설정 파일 분실), #187 (워크플로우 덮어쓰기), #162 (파일 덮어쓰기), #236 (콘텐츠 삭제), #318 (동기화 실패), #319 (플러그인 전체 실패), #253 (PyPI 미발견), #296 (PATH 이슈), #159 (uv 업그레이드 실패), #312 (uv 업데이트 실패)

#### 3.2.6 cc / glm -- 백엔드 전환

**Python 구현 (325 LOC)**

```
moai-adk claude          # Claude 백엔드 전환
moai-adk cc              # claude 별칭
moai-adk glm [API_KEY]   # GLM 백엔드 전환 또는 API 키 업데이트
```

- 2개의 별도 top-level 커맨드 (`claude`/`cc`, `glm`)
- config 파일 재작성 (LLM 백엔드 설정)
- API 키 관리

**Go 타겟 설계 (~120 LOC)**

```
moai cc                  # Claude 백엔드 (claude 별칭)
moai glm [KEY]           # GLM 백엔드
```

- Python과 동일하게 별도 top-level 커맨드로 유지 (`switch` 미사용)
- `internal/cli/cc.go`: Claude 백엔드 전환
- `internal/cli/glm.go`: GLM 백엔드 전환 + API 키 관리
- `internal/config/manager.go`로 설정 업데이트
- Viper 기반 타입 안전 설정 수정

**마이그레이션 복잡도**: Medium -- config 재작성 로직 이식 필요

#### 3.2.7 rank -- 성능 랭킹 (7 서브커맨드)

**Python 구현 (478 LOC)**

```
moai-adk rank login             # 인증
moai-adk rank status            # 랭킹 상태
moai-adk rank logout            # 로그아웃
moai-adk rank sync              # 메트릭 동기화
moai-adk rank exclude [PATTERN] # 제외 패턴
moai-adk rank include [PATTERN] # 포함 패턴
moai-adk rank register [--org]  # 조직 등록
```

- MoAI Cloud API 통합
- 자격 증명 관리 (keyring)
- 세션 메트릭 수집 및 제출

**Go 타겟 설계 (~200 LOC)**

- `internal/rank/client.go`: HTTP 클라이언트
- `internal/rank/auth.go`: 시스템 keyring 통합 (macOS Keychain, Linux secret-service)
- `internal/rank/config.go`: 랭킹 설정
- stdlib `net/http` 사용

**마이그레이션 복잡도**: Medium -- API 통합 + 자격 증명 관리

#### 3.2.8 worktree -- Git Worktree 관리 (10 서브커맨드)

**Python 구현 (2,034 LOC 총합)**

```
moai-wt new [SPEC-ID] [OPTIONS]   # 새 worktree 생성
moai-wt list                       # worktree 목록
moai-wt go [NAME]                  # worktree 이동
moai-wt remove [NAME] [OPTIONS]    # worktree 제거
moai-wt status                     # worktree 상태
moai-wt sync [NAME]                # worktree 동기화
moai-wt clean [OPTIONS]            # 정리
moai-wt recover [NAME]             # worktree 복구
moai-wt done [NAME]                # worktree 완료 처리
moai-wt config [OPTIONS]           # 설정 관리
```

- `cli.py` (885 LOC): 10개 서브커맨드 정의
- `manager.py` (662 LOC): worktree CRUD 로직
- `registry.py` (422 LOC): JSON 레지스트리 관리
- `models.py` (65 LOC): 데이터 모델
- 별도 엔트리포인트: `moai-worktree`, `moai-wt`

**Go 타겟 설계 (~400 LOC)**

```
moai worktree new [SPEC-ID]
moai worktree list
moai worktree switch [NAME]
moai worktree remove [NAME]
moai worktree status
moai worktree sync [NAME]
moai worktree clean
```

- `internal/core/git/` 활용: go-git + 시스템 Git fallback (ADR-007)
- `internal/cli/worktree/*.go`: 각 서브커맨드별 파일
- 별도 엔트리포인트 제거: `moai worktree` 서브커맨드로 통합

**마이그레이션 복잡도**: **High** -- Git 작업, 플랫폼 민감성, 파일 시스템 조작

#### 3.2.9 statusline -- Claude Code Statusline

**Python 구현 (420 LOC CLI + 2,877 LOC 지원 모듈 = 3,297 LOC)**

```
moai-adk statusline   # stdin에서 JSON 컨텍스트 읽기
```

- `main.py` (420 LOC): statusline 데이터 빌드
- `renderer.py` (468 LOC): 터미널 렌더링
- `version_reader.py` (769 LOC): 다양한 형식 버전 파일 읽기
- `config.py` (379 LOC): statusline 설정
- `enhanced_output_style_detector.py` (450 LOC): 출력 스타일 감지
- `memory_collector.py` (268 LOC): 메모리 수집
- `git_collector.py` (190 LOC): Git 상태 수집
- 기타 모듈 (353 LOC)

**Go 타겟 설계 (~40 LOC CLI + ~600 LOC 모듈)**

- `internal/statusline/builder.go`: 데이터 빌드
- `internal/statusline/git.go`: go-git in-process Git 상태
- `internal/statusline/memory.go`: OS API 직접 호출
- `internal/statusline/renderer.go`: lipgloss 렌더링
- `internal/statusline/update.go`: 업데이트 확인
- version_reader 제거: build-time ldflags로 대체
- output_style_detector 제거: CLAUDE.md에 내장

**마이그레이션 복잡도**: Low-Medium -- 대부분 데이터 수집 + 포맷팅

### 3.3 드롭된 커맨드

두 커맨드가 Go 버전에서 제거된다:

| 커맨드 | Python LOC | 드롭 사유 | 기능 흡수 위치 |
|--------|-----------|----------|---------------|
| `language` | 255 | 비활성 상태, 단독 사용 사례 없음 | `init` 및 `config` 서브시스템에 흡수 |
| `analyze` | 125 | 비활성 상태, `doctor`/`status`와 기능 중복 | `doctor --verbose`에 흡수 |

**삭감 효과**: -380 LOC (Python 기준), 커맨드 표면 간소화

### 3.4 Go 신규 커맨드

Go 버전에서 새롭게 추가되는 커맨드:

| 커맨드 | 목적 | 예상 LOC | 근거 |
|--------|------|---------|------|
| `moai hook <event>` | 명시적 hook 실행 CLI | ~200 | ADR-006: Hook을 바이너리 서브커맨드로 |
| `moai hook list` | 등록된 hook 목록 표시 | ~50 | 디버깅 및 진단 지원 |
| `moai hook run <event>` | 수동 hook 실행 | ~80 | 테스트 및 개발 지원 |
| `moai completion [shell]` | 네이티브 Shell Completion 생성 | ~30 | Cobra 내장 기능 활용 |
| (향후) `moai config get/set` | CLI 기반 설정 관리 | ~100 | Viper 기반 직접 설정 조작 |

---

## 4. 아키텍처 비교

### 4.1 엔트리 포인트 비교

| 측면 | Python | Go | 변경점 |
|------|--------|-----|-------|
| 엔트리 포인트 수 | 4개 | 1개 | -75% |
| 기본 CLI | `moai-adk` | `moai` | 이름 간소화 |
| 단축 CLI | `moai` (alias) | - | 별칭 불필요 |
| Worktree CLI | `moai-worktree` | `moai worktree` | 서브커맨드 통합 |
| Worktree 단축 | `moai-wt` | `moai wt` (alias) | 서브커맨드 통합 |

Python의 4개 별도 엔트리포인트는 패키징 복잡성을 증가시켰다 (`pyproject.toml`에 4개의 `[project.scripts]` 항목). Go에서는 단일 `moai` 바이너리로 모든 기능을 제공한다.

### 4.2 프로젝트 구조 비교

**Python 구조:**

```
moai_adk/
├── __main__.py              (339 LOC, lazy loading entry point)
├── __init__.py              (version string)
├── version.py               (version management)
├── cli/
│   ├── commands/
│   │   ├── init.py          (854 LOC)
│   │   ├── doctor.py        (370 LOC)
│   │   ├── status.py        (111 LOC)
│   │   ├── update.py        (3,162 LOC -- 가장 복잡)
│   │   ├── switch.py        (325 LOC, claude/cc + glm 커맨드)
│   │   ├── rank.py          (478 LOC, 7 서브커맨드)
│   │   ├── language.py      (255 LOC, DROPPED)
│   │   ├── analyze.py       (125 LOC, DROPPED)
│   │   └── lsp_setup.py     (LSP 설정)
│   ├── worktree/
│   │   ├── cli.py           (885 LOC, 10 서브커맨드)
│   │   ├── manager.py       (662 LOC)
│   │   ├── registry.py      (422 LOC)
│   │   └── models.py        (65 LOC)
│   ├── ui/                  (theme, prompts, progress)
│   └── prompts/             (init prompts, translations)
├── statusline/              (11 파일, 3,297 LOC)
├── core/                    (80+ 파일, 핵심 비즈니스 로직)
│   ├── config/              (unified config manager)
│   ├── project/             (initializer, detector, validator)
│   ├── git/                 (manager, branch, conflict)
│   ├── template/            (processor, merger, backup)
│   ├── migration/           (version migrator, backup)
│   ├── quality/             (TRUST checker, validators)
│   └── ...                  (40+ 추가 모듈)
├── foundation/              (EARS, langs, backend, frontend, DB patterns)
└── hooks/ (lib/)            (32 파일, 21,535 LOC -- 별도 디렉토리)
```

**Go 구조 (타겟):**

```
cmd/moai/
└── main.go                  (14 LOC, thin entry)

internal/
├── cli/                     (Cobra 커맨드 정의)
│   ├── root.go              (30 LOC)
│   ├── init.go              (~300 LOC)
│   ├── doctor.go            (~150 LOC)
│   ├── status.go            (~80 LOC)
│   ├── version.go           (27 LOC, Done)
│   ├── update.go            (~800 LOC)
│   ├── hook.go              (~200 LOC, NEW)
│   ├── cc.go                (~60 LOC)
│   ├── glm.go               (~60 LOC)
│   ├── rank.go              (~200 LOC)
│   └── worktree/            (~400 LOC)
│       ├── new.go
│       ├── list.go
│       ├── switch.go
│       ├── sync.go
│       ├── remove.go
│       └── clean.go
├── hook/                    (컴파일된 hook 핸들러, ~2,500 LOC)
│   ├── registry.go          (핸들러 등록 & 디스패치)
│   ├── protocol.go          (JSON stdin/stdout 프로토콜)
│   ├── contract.go          (실행 계약, ADR-012)
│   ├── session_start.go     (프로젝트 정보, config 검증)
│   ├── pre_tool.go          (보안 가드)
│   ├── post_tool.go         (linter, formatter, LSP)
│   ├── session_end.go       (정리, rank 제출)
│   ├── stop.go              (루프 컨트롤러)
│   └── compact.go           (컨텍스트 보존)
├── config/                  (Viper 기반 타입 안전 설정)
├── template/                (go:embed 배포)
├── manifest/                (파일 provenance 추적, NEW)
├── merge/                   (3-way merge 엔진, NEW)
├── update/                  (Self-Update 시스템, NEW)
├── core/
│   ├── project/             (초기화, 감지, 검증)
│   ├── git/                 (go-git + 시스템 Git fallback)
│   ├── quality/             (TRUST 5 프레임워크)
│   ├── integration/         (통합 테스트 엔진)
│   └── migration/           (버전 마이그레이션)
├── foundation/              (EARS, 언어 정의, 도메인 패턴)
├── lsp/                     (LSP 클라이언트)
├── loop/                    (Ralph 피드백 루프)
├── ralph/                   (의사결정 엔진)
├── rank/                    (성능 랭킹)
├── statusline/              (Claude Code statusline)
├── astgrep/                 (AST-Grep 통합)
└── ui/                      (Charmbracelet TUI)

pkg/
├── version/                 (빌드 타임 버전)
├── models/                  (공유 데이터 구조)
└── utils/                   (로거, 파일 I/O, 경로)

templates/                   (go:embed 소스, 바이너리에 번들)
```

### 4.3 Hook 시스템 아키텍처 변경

이것은 Python에서 Go로의 전환에서 **가장 큰 아키텍처 변경**이다.

#### Python Hook 시스템 (현재)

- **규모**: 12 hook 스크립트 + 20 라이브러리 모듈 = 32 파일, 21,535 LOC
- **실행 모델**: Claude Code가 Python 스크립트를 subprocess로 스폰
- **설정 형식**: `"${SHELL:-/bin/bash} -c 'python3 \"$CLAUDE_PROJECT_DIR/.claude/hooks/...\"'"`

**핵심 문제점:**

| 문제 카테고리 | 이슈 수 | 주요 예시 |
|-------------|---------|----------|
| Python 런타임 의존성 | 12 | #278 PyYAML 미발견, #288 uv 버전 감지 실패, #269 import 에러 |
| PATH 해석 실패 | 8 | #259 Windows 혼합 구분자, #161 `$CLAUDE_PROJECT_DIR` 미설정 |
| 플랫폼 비호환 | 5 | #129 SIGALRM (치명적), #249 cp1252 인코딩, #25 무한 대기 |
| Hook 형식/프로토콜 | 3 | #265 settings.json 형식 비호환, #207 중복 실행 |
| **합계** | **28** | |

**근본 원인 분석:**

1. Hook 스크립트가 Python 인터프리터에 의존하므로 `python3`이 PATH에 있어야 함
2. `$CLAUDE_PROJECT_DIR` 환경변수가 비대화형 셸에서 미설정 가능
3. Windows에서 `signal.SIGALRM` 미지원 (Unix 전용 신호)
4. 셸 래퍼 (`${SHELL:-/bin/bash} -c '...'`)가 플랫폼별 셸 차이 노출
5. 템플릿 변수 치환 방식으로 settings.json 생성 시 JSON 구문 오류 가능

#### Go Hook 시스템 (타겟)

- **규모**: 6 핸들러 + 3 인프라 파일 = 9 파일, ~2,500 LOC
- **실행 모델**: `moai hook <event>` 바이너리 서브커맨드 (ADR-006)
- **설정 형식**: `"moai hook session-start"` (경로 없음, 셸 래퍼 없음)

**아키텍처 결정 레코드:**

- **ADR-006: Hooks as Binary Subcommands**: Hook을 동일 바이너리의 서브커맨드로 구현
- **ADR-012: Hook Execution Contract**: Hook 실행 환경의 공식 보증 정의
  - 보증: stdin JSON, stdout JSON, exit code 의미, timeout 동작
  - 비보증: 사용자 PATH, 셸 환경, Python 가용성

**Before/After 비교:**

```
# Python (Before) - 4개의 PostToolUse hook 엔트리
"command": "${SHELL:-/bin/bash} -c 'python3 \"$CLAUDE_PROJECT_DIR/.claude/hooks/post_tool__code_formatter.py\"'"
"command": "${SHELL:-/bin/bash} -c 'python3 \"$CLAUDE_PROJECT_DIR/.claude/hooks/post_tool__linter.py\"'"
"command": "${SHELL:-/bin/bash} -c 'python3 \"$CLAUDE_PROJECT_DIR/.claude/hooks/post_tool__lsp_diagnostic.py\"'"
"command": "${SHELL:-/bin/bash} -c 'python3 \"$CLAUDE_PROJECT_DIR/.claude/hooks/post_tool__ast_grep_scan.py\"'"

# Go (After) - 1개의 통합 서브커맨드
"command": "moai hook post-tool"
```

**정량적 개선:**

| 메트릭 | Python (Before) | Go (After) | 변화 |
|--------|----------------|-----------|------|
| 총 LOC | 21,535 (32 파일) | ~2,500 (9 파일) | **-88%** |
| Hook 스크립트 | 12 파일 | 0 파일 | -100% |
| 라이브러리 모듈 | 20 파일 | 0 파일 | -100% |
| settings.json 엔트리/이벤트 | 1-4개 Python 커맨드 | 1개 Go 서브커맨드 | 최대 -75% |
| 런타임 의존성 | Python 3.13+, PyYAML, Rich, uv | 없음 (컴파일 바이너리) | **-100%** |
| PATH 이슈 | 8건 | 0건 | **-100%** |
| SIGALRM 이슈 | 1건 (치명적, #129) | 0건 | **-100%** |
| 인코딩 이슈 | 3건 | 0건 | **-100%** |
| 관련 GitHub 이슈 | 28건 | 0건 (설계상) | **-100%** |

### 4.4 Update 시스템 재설계

Update 시스템은 가장 복잡한 커맨드(3,162 LOC)이며 가장 자주 수정된 파일(38회)이다.

#### Python Update 시스템

```
moai-adk update
  │
  ├─ 1. PyPI 버전 확인 (requests)
  ├─ 2. 패키지 매니저 업그레이드 (pip/uv/pipx)
  ├─ 3. 새 버전 템플릿 다운로드
  ├─ 4. 현재 파일과 비교
  ├─ 5. 문자열 기반 템플릿 변수 치환
  │     ├─ {{HOOK_SHELL_PREFIX}} → ${SHELL:-/bin/bash} -c '...'
  │     ├─ {{PYTHON_PATH}} → python3 경로
  │     └─ {{PROJECT_DIR}} → $CLAUDE_PROJECT_DIR
  ├─ 6. 파일별 덮어쓰기/건너뛰기 결정
  ├─ 7. 백업 생성 (선택)
  └─ 8. 마이그레이션 스크립트 실행
```

**문제점:**

- 문자열 치환에서 따옴표, 경로 구분자, 셸 변수로 인한 JSON 구문 오류 (4회 회귀 사이클)
- 파일 provenance 추적 없음 -> 사용자 수정 파일 무조건 덮어쓰기
- 롤백 메커니즘 미흡 -> 실패 시 복구 어려움
- 38회 수정에도 완전 해결되지 않는 구조적 문제

#### Go Update 시스템 (타겟)

```
moai update
  │
  ├─ 1. GitHub Releases API 버전 확인
  ├─ 2. 새 바이너리를 임시 위치에 다운로드
  ├─ 3. 현재 manifest 로드 (.moai/manifest.json)
  ├─ 4. 새 바이너리에서 새 템플릿 추출 (go:embed)
  ├─ 5. 각 파일에 대해:
  │     ├─ manifest에 없음 → template_managed로 배포
  │     ├─ template_managed + 변경 없음 → 안전 덮어쓰기
  │     ├─ template_managed + 사용자 변경 → user_modified로 승격, 3-way merge
  │     ├─ user_modified → 3-way merge (base=deployed, current=actual, new=template)
  │     │   ├─ 충돌 없음 → 자동 merge
  │     │   └─ 충돌 → .conflict 파일 작성, 현재 파일 유지, 알림
  │     └─ user_created → 건너뛰기 (절대 수정 안 함)
  ├─ 6. 새 템플릿에 없는 manifest 파일 → deprecated 표시, 알림, 파일 유지
  ├─ 7. manifest 업데이트
  ├─ 8. 바이너리 교체 (원자적: temp 작성 → rename)
  └─ 9. 요약 표시: updated/merged/conflicted/skipped
```

**핵심 ADR:**

- **ADR-007: File Manifest Provenance**: 모든 배포 파일의 출처와 해시 추적
- **ADR-011: Zero Runtime Template Expansion**: `json.MarshalIndent()`로 JSON 생성, 문자열 치환 금지

---

## 5. 성능 비교

### 5.1 시작 성능

| 메트릭 | Python | Go | 개선율 |
|--------|--------|-----|-------|
| Cold Startup | 200-500ms | <50ms 목표 | **4-10x 빠름** |
| Import/Link 시간 | 100-300ms | <5ms | **20-60x 빠름** |
| 커맨드 실행 오버헤드 | +50-100ms | 거의 0 | **제거됨** |
| Lazy Loading 효과 | ~400ms -> ~100ms | N/A | 불필요 |

**Python 시작 시간 분해:**

1. Python 인터프리터 시작: ~50-100ms
2. Click 프레임워크 import: ~20-30ms
3. Rich 라이브러리 import: ~30-50ms (lazy loading 적용 시 지연)
4. 기타 의존성 import: ~20-50ms
5. 커맨드 디스패치: ~5-10ms

**Go 시작 시간 분해:**

1. 바이너리 로드 (OS): ~5-10ms
2. 런타임 초기화: ~1-2ms
3. Cobra 커맨드 트리 구축: ~1-2ms
4. 커맨드 디스패치: ~1ms

Python에서는 이 문제를 해결하기 위해 LazyGroup 패턴, 조건부 import, `_console = None` 지연 초기화 등 다양한 최적화 기법을 적용했지만 인터프리터 자체의 오버헤드는 제거할 수 없다. Go에서는 이러한 최적화가 근본적으로 불필요하다.

### 5.2 메모리 사용

| 메트릭 | Python | Go | 비고 |
|--------|--------|-----|------|
| 인터프리터 베이스라인 | ~30-50MB | N/A | Go는 인터프리터 없음 |
| 커맨드 실행 시 | ~50-100MB | ~10-30MB | 2-5x 효율적 |
| Idle 메모리 | ~50MB | <20MB 목표 | |
| 동시 작업 | GIL 제한 | goroutine 기반 | Go 네이티브 병렬성 |
| 16 LSP 서버 피크 | 추정 ~300MB+ | <200MB 목표 | goroutine 경량 |

### 5.3 바이너리/패키지 크기

| 메트릭 | Python | Go | 분석 |
|--------|--------|-----|------|
| 패키지 크기 | ~2MB (wheel) | 15-30MB (standalone) | Go 바이너리가 더 큼 |
| 런타임 요구사항 | Python 3.10+ (~100MB) | 없음 | **Go 무의존** |
| 총 디스크 사용 | ~102MB (Python + wheel) | 15-30MB | **Go 70-85% 절감** |
| 임베디드 템플릿 | 별도 다운로드 | 바이너리에 포함 (2-5MB) | 배포 단순화 |
| 스트립 최적화 | N/A | `-ldflags "-s -w"` (30-40% 감소) | Go 특유 |

**트레이드오프**: Go 바이너리 자체(15-30MB)는 Python wheel(~2MB)보다 크지만, Python 런타임(~100MB)을 포함하면 총 디스크 사용량은 Go가 70-85% 적다.

### 5.4 Hook 실행 성능

| 메트릭 | Python Hook | Go Hook | 개선율 |
|--------|-----------|---------|-------|
| Hook 시작 시간 | 200-500ms (인터프리터) | <50ms (컴파일) | 4-10x |
| SessionStart hook | ~500ms-1s | ~100ms 목표 | 5-10x |
| PostToolUse hook (4개) | ~2-4s (순차) | ~200ms (통합) | 10-20x |
| JSON 파싱 | `json.loads()` | `json.Unmarshal()` | 동등 |
| 메모리 (hook당) | ~50MB (Python 프로세스) | ~10MB (바이너리) | 5x |

---

## 6. 의존성 생태계 비교

### 6.1 의존성 수 요약

| 카테고리 | Python | Go | 변화 |
|---------|--------|-----|------|
| 직접 의존성 | ~15 | ~8 | **-47%** |
| 전이적 의존성 | ~50+ | ~20 | **-60%** |
| 보안 표면 | Medium-High | Low | 감소 |
| 런타임 의존성 | Python 3.10+, pip/uv | 없음 | **제거** |

### 6.2 Python 주요 의존성

| 라이브러리 | 버전 | 용도 | Go 대체 |
|-----------|------|------|---------|
| `click` | >=8.1.0 | CLI 프레임워크 | Cobra v1.10.2 |
| `rich` | >=13.0.0 | 터미널 UI/출력 | lipgloss + bubbletea |
| `inquirerpy` | >=0.3.4 | 대화형 프롬프트 | bubbletea + huh |
| `pyyaml` | >=6.0 | YAML 파싱 | gopkg.in/yaml.v3 |
| `packaging` | >=23.0 | 버전 비교 | go-version (계획) |
| `gitpython` | - | Git 작업 | go-git (계획) |
| `requests` / `urllib3` | - | HTTP 클라이언트 | `net/http` (stdlib) |
| `pyfiglet` | - | ASCII 아트 로고 | 제거 (또는 lipgloss) |
| `toml` / `tomli` | - | TOML 파싱 | `encoding/json` + `gopkg.in/yaml.v3` |
| `keyring` | - | 자격증명 관리 | 시스템 keyring 직접 접근 |

### 6.3 Go 주요 의존성

| 라이브러리 | 버전 | 용도 | 상태 |
|-----------|------|------|------|
| `github.com/spf13/cobra` | v1.10.2 | CLI 프레임워크 | **설치됨** |
| `github.com/spf13/viper` | v1.18+ (계획) | 설정 관리 | 계획 |
| `github.com/go-git/go-git/v5` | v5.12+ (계획) | Git 작업 | 계획 |
| `github.com/charmbracelet/bubbletea` | v1.2+ (계획) | TUI 프레임워크 | 계획 |
| `github.com/charmbracelet/lipgloss` | v1.0+ (계획) | 터미널 스타일링 | 계획 |
| `gopkg.in/yaml.v3` | v3.0+ (계획) | YAML 파싱 | 계획 |
| `github.com/stretchr/testify` | v1.9+ (계획) | 테스트 assertion | 계획 |
| `go.lsp.dev/protocol` | v0.12+ (계획) | LSP 프로토콜 타입 | 계획 |

**Go stdlib 활용:**

| stdlib 패키지 | 용도 | Python 동등 라이브러리 |
|-------------|------|---------------------|
| `log/slog` | 구조화 로깅 | `logging` + `structlog` |
| `embed` | 템플릿 임베딩 | 별도 다운로드 필요 |
| `testing` | 단위 테스트 | `pytest` (3rd party) |
| `net/http` | HTTP 클라이언트 | `requests` (3rd party) |
| `encoding/json` | JSON 처리 | `json` (stdlib, 동등) |
| `context` | 타임아웃/취소 | `asyncio` + `signal.SIGALRM` |
| `os/exec` | 프로세스 실행 | `subprocess` |
| `path/filepath` | 경로 처리 | `pathlib` / `os.path` |

### 6.4 의존성 관리 비교표

| 차원 | Python | Go | 분석 |
|------|--------|-----|------|
| 패키지 매니저 | pip/pipx/uv | `go mod` | Go 단일 도구 |
| Lock 파일 | requirements.txt / uv.lock | `go.sum` (체크섬) | Go 암호학적 검증 |
| 가상 환경 | 필수 (venv/virtualenv) | N/A | Go 불필요 |
| CVE 노출 | Medium-High | Low | Go 의존성 수 적음 |
| 재현성 | `pip freeze` / `uv lock` | `go.sum` (불변) | Go 더 엄격 |
| 라이선스 검사 | 수동 / `pip-licenses` | `go-licenses` | 동등 |
| 취약점 스캔 | `safety` / `bandit` | `govulncheck` (공식) | Go 공식 도구 |
| 빌드 격리 | `build` 패키지 | `CGO_ENABLED=0` | Go 완전 격리 |

---

## 7. 테스트 및 품질 인프라 비교

### 7.1 테스트 프레임워크

| 도구 | Python | Go | 비고 |
|------|--------|-----|------|
| 단위 테스트 | `pytest` | `testing` (stdlib) | Go 내장 |
| CLI 테스트 | `click.testing.CliRunner` | 직접 커맨드 실행 | 동등 |
| Mocking | `pytest-mock` / `unittest.mock` | Interface 기반 + `mockery` | Go가 더 명시적 |
| 커버리지 | `pytest-cov` | `go test -cover` (내장) | Go 내장 |
| 벤치마크 | `pytest-benchmark` | `testing.B` (내장) | Go 내장 |
| Fuzzing | `hypothesis` (3rd party) | `testing.F` (내장) | Go 내장 |
| Race Detection | N/A | `go test -race` (내장) | **Go 고유** |
| 보안 스캔 | `bandit` / `safety` | `gosec` + `govulncheck` | Go 공식 도구 |
| 정적 분석 | `mypy` (선택) | 컴파일러 (필수) | Go 강제 |

### 7.2 타입 안전성

| 측면 | Python | Go |
|------|--------|-----|
| 타입 시스템 | 선택적 (`mypy` 힌트) | 필수 (컴파일 타임) |
| 런타임 타입 에러 | 가능 | 불가능 |
| 타입 힌트 적용률 | 프로젝트별 상이 | 100% (컴파일 필수) |
| Null Safety | `Optional[]` 힌트 (미강제) | 명시적 포인터 / zero value |
| 인터페이스 | `Protocol` / `ABC` (선택) | 암시적 인터페이스 (강제) |
| 제네릭 | Python 3.12+ | Go 1.18+ | 양쪽 지원 |

**실제 영향**: Python에서는 `mypy`를 통과해도 런타임에 타입 에러가 발생할 수 있다 (예: #278 PyYAML 미설치로 인한 `ImportError`). Go에서는 컴파일이 되면 타입 에러가 런타임에 발생하지 않는다.

### 7.3 Linting 비교

| 측면 | Python | Go |
|------|--------|-----|
| 주요 린터 | `ruff` | `golangci-lint` |
| 포맷터 | `black` (3rd party) | `gofmt` / `gofumpt` (공식) |
| Import 정렬 | `isort` (3rd party) | `goimports` (공식) |
| 린터 수 | ruff 800+ 규칙 | golangci-lint 100+ 린터 |
| 보안 린팅 | `bandit` | `gosec` (golangci-lint 내) |
| 실행 속도 | 빠름 (ruff: Rust 기반) | 빠름 (네이티브) |

### 7.4 Go 특화 테스트 카테고리

Go 버전에서 추가되는 테스트 유형:

| 테스트 카테고리 | 목적 | 위치 |
|---------------|------|------|
| Hook Contract Tests | Hook 실행 계약 검증 (ADR-012) | `internal/hook/*_test.go` |
| JSON Safety Tests | settings.json 생성 유효성 + 경로 정규화 | `internal/template/*_test.go` |
| Minimal PATH Test | `moai hook <event>`가 /usr/bin:/bin만으로 작동 확인 | `internal/hook/*_test.go` |
| Non-Interactive Shell Test | .bashrc/.zshrc 없이 hook 작동 확인 | `internal/hook/*_test.go` |
| Cross-Platform Test | darwin, linux, windows 동작 확인 | CI matrix |
| Race Detection | goroutine 안전성 검증 | `go test -race ./...` |
| Fuzz Tests | 설정 파싱, CLI 인자 경계 탐색 | `*_fuzz_test.go` |

---

## 8. 기능 패리티 분석

### 8.1 Category A: 완전 매핑 (기능 보존 + 개선)

직접적으로 대응되며 Go에서 개선되는 커맨드:

| 커맨드 | Python | Go | 주요 개선점 |
|--------|--------|-----|----------|
| `init` | 854 LOC, InquirerPy | ~300 LOC, bubbletea wizard | Manifest 추적 추가, TTY 자동 감지 |
| `doctor` | 370 LOC, Rich 패널 | ~150 LOC, lipgloss 출력 | Hook contract 검사 추가 |
| `status` | 111 LOC, Rich 패널 | ~80 LOC, lipgloss 출력 | 동일 기능, LOC 감소 |
| `version` | ~50 LOC | 27 LOC (**Done**) | commit hash + build date 추가 |
| `update` | 3,162 LOC (불안정) | ~800 LOC (재설계) | Manifest + 3-way merge + self-update |

### 8.2 Category B: 아키텍처 변경으로 매핑

아키텍처가 근본적으로 변경되는 영역:

| 영역 | Python | Go | 아키텍처 변경 |
|------|--------|-----|-------------|
| `cc` / `glm` | 2 별도 커맨드 (claude/cc, glm) | 2 별도 커맨드 유지 (cc, glm) | 동일 구조, switch 미사용 |
| `rank` (7 sub) | Click group | Cobra 서브커맨드 | 동일 패턴, 프레임워크 변경 |
| `worktree` (10 sub) | 별도 엔트리포인트 (moai-wt) | 서브커맨드 통합 | 엔트리포인트 통합 |
| Hook 시스템 | 32 파일, 21,535 LOC | 9 파일, ~2,500 LOC | 스크립트 -> 컴파일 서브커맨드 (ADR-006) |
| Statusline | 11 파일, 3,297 LOC | 6 파일, ~640 LOC | 모듈 통합, 불필요 탐지기 제거 |

### 8.3 Category C: 드롭

| 커맨드 | Python LOC | 드롭 사유 | 흡수 위치 |
|--------|-----------|----------|----------|
| `language` | 255 | 비활성, `init`에서 처리 | `init --locale` |
| `analyze` | 125 | 비활성, `doctor`와 중복 | `doctor --verbose` |

### 8.4 Category D: Go 신규 기능

| 기능 | 설명 | Python 동등 | 근거 |
|------|------|------------|------|
| `moai hook <event>` | 명시적 hook 관리 CLI | 없음 (암시적 스크립트) | ADR-006, 디버깅 용이 |
| `moai hook list` | 등록된 hook 표시 | 없음 | 투명한 hook 관리 |
| Self-Update | 바이너리 자체 업데이트 | PyPI 경유 업데이트 | 패키지 매니저 독립 |
| Shell Completion | 네이티브 완성 생성 | 수동 설정 필요 | Cobra 내장 기능 |
| Contract Tests | Hook 실행 계약 테스트 | 없음 | ADR-012 |
| File Manifest | 파일 provenance 추적 | 없음 | ADR-007 |
| 3-Way Merge | 지능적 파일 merge | 단순 덮어쓰기 | 사용자 변경 보존 |

---

## 9. 위험 평가 및 마이그레이션 전략

### 9.1 위험도 매트릭스

| 커맨드 | 위험 수준 | 사유 | 복잡도 요인 |
|--------|----------|------|-----------|
| `version` | Done | 이미 구현 완료 | - |
| `status` | Low | 단순 표시 커맨드, 외부 의존성 없음 | config 읽기 |
| `doctor` | Low | 진단 검사, 주로 `exec.LookPath()` | 검사 항목 이식 |
| `statusline` | Low | 데이터 수집 + 포맷팅 | version_reader 제거 |
| `init` | Medium | 템플릿 배포 + 대화형 UI + manifest 통합 | bubbletea 학습 곡선 |
| `cc` / `glm` | Medium | config 재작성 로직 | Viper 통합 |
| `rank` | Medium | 클라우드 API 통합 + 자격증명 관리 | HTTP + keyring |
| `worktree` | **High** | Git 작업, 플랫폼 민감, 파일 시스템 조작 | go-git 한계 |
| `update` | **High** | 완전 재설계, manifest + merge + self-update | 가장 복잡한 신규 시스템 |
| Hook 시스템 | **High** | 아키텍처 전환, 28개 이슈 해결 대상 | 계약 테스트 필수 |

### 9.2 구현 순서 (5 Phase)

#### Phase 1: Foundation (init, doctor, version)

| 커맨드 | 상태 | 선행 모듈 |
|--------|------|----------|
| `version` | **Done** | `pkg/version/` |
| `init` | Stub | `internal/config/`, `internal/template/`, `internal/manifest/`, `internal/ui/` |
| `doctor` | Stub | `internal/config/`, `internal/core/project/checker.go` |

**현재 진행률**: ~33% (version 완료)
**예상 LOC**: ~450 (CLI) + ~2,500 (지원 모듈)
**핵심 의존성**: Viper, bubbletea, go:embed 인프라 구축

#### Phase 2: Runtime (status, hook, statusline)

| 커맨드 | 상태 | 선행 모듈 |
|--------|------|----------|
| `status` | Stub | `internal/config/`, `internal/core/git/` |
| `hook` | Not started | `internal/hook/` 전체 |
| `statusline` | Not started | `internal/statusline/` |

**현재 진행률**: ~10% (status stub)
**예상 LOC**: ~200 (CLI) + ~3,100 (hook + statusline 모듈)
**핵심 과제**: Hook 시스템 아키텍처 전환, 계약 테스트 작성

#### Phase 3: Infrastructure (update, cc/glm)

| 커맨드 | 상태 | 선행 모듈 |
|--------|------|----------|
| `update` | Not started | `internal/update/`, `internal/merge/`, `internal/manifest/` |
| `cc` / `glm` | Not started | `internal/config/` |

**현재 진행률**: 0%
**예상 LOC**: ~920 (CLI) + ~2,000 (update + merge 모듈)
**핵심 과제**: 3-way merge 엔진, self-update 메커니즘

#### Phase 4: Cloud (rank 7 서브커맨드)

| 커맨드 | 상태 | 선행 모듈 |
|--------|------|----------|
| `rank` (7 sub) | Not started | `internal/rank/` |

**현재 진행률**: 0%
**예상 LOC**: ~200 (CLI) + ~400 (rank 모듈)
**핵심 과제**: API 클라이언트, 자격증명 관리

#### Phase 5: Collaboration (worktree 10 서브커맨드)

| 커맨드 | 상태 | 선행 모듈 |
|--------|------|----------|
| `worktree` (10 sub) | Not started | `internal/core/git/` |

**현재 진행률**: 0%
**예상 LOC**: ~400 (CLI) + ~600 (git worktree 모듈)
**핵심 과제**: go-git worktree 한계, 시스템 Git fallback

### 9.3 마이그레이션 전략

3단계 마이그레이션 계획:

#### Stage 1: 병렬 운영 (Parallel Operation)

- Go 바이너리가 Python과 공존
- `moai` (Go) + `moai-adk` (Python) 동시 설치 가능
- 사용자가 점진적으로 Go 버전으로 전환
- `.moai/` 디렉토리 형식 호환 유지

#### Stage 2: 기능 패리티 (Feature Parity)

- Go가 100% 기능 패리티 달성
- Python 버전은 유지보수 모드 전환
- 자동 마이그레이션 도구 제공
- 설정 파일 형식 호환 검증 완료

#### Stage 3: Python 지원 중단 (Python Deprecation)

- Python 에디션 공식 지원 중단 발표
- 6개월 지원 중단 기간
- `moai-adk` -> `moai`로 완전 전환
- Python 코드베이스 아카이브

---

## 10. 정량적 요약

### 10.1 종합 비교 대시보드

| 메트릭 | moai-adk (Python) | moai-go (Go) | Delta |
|--------|-------------------|--------------|-------|
| CLI 프레임워크 | Click 8.1+ | Cobra v1.10.2 | - |
| 총 프로젝트 LOC | ~88,319 (220 파일) | ~18,000 (목표) | **-80%** |
| 순수 CLI LOC | ~6,523 | ~2,161 (목표) | **-67%** |
| 구현 진행률 | 100% | ~3% | Gap |
| 최상위 커맨드 수 | 12 (2 비활성) | 10 (목표) | -2 |
| 서브커맨드 수 | 23+ | 20+ (목표) | ~동일 |
| 엔트리 포인트 | 4 (moai-adk, moai, moai-wt, moai-worktree) | 1 (moai) | **-75%** |
| 직접 의존성 | ~15 | ~8 | **-47%** |
| 전이적 의존성 | ~50+ | ~20 | **-60%** |
| 시작 시간 | 200-500ms | <50ms | **-90%** |
| Hook LOC | 21,535 (32 파일) | ~2,500 (9 파일) | **-88%** |
| Hook 관련 이슈 | 28 | 0 (설계상) | **-100%** |
| 런타임 요구사항 | Python 3.10+ (~100MB) | 없음 | **제거** |
| 타입 안전성 | 선택적 (`mypy`) | 필수 (컴파일러) | 향상 |
| Race Detection | 없음 | `go test -race` 내장 | **Go 고유** |
| Fuzzing | `hypothesis` (3rd party) | `testing.F` 내장 | Go 내장 |
| 플랫폼 타겟 | PyPI (any, 런타임 의존) | 6개 명시적 (CGO_ENABLED=0) | 명시적 |
| 총 디스크 사용 | ~102MB (Python + wheel) | 15-30MB | **-70~85%** |
| 설정 형식 | YAML (동일) | YAML (동일) | 호환 |
| 템플릿 배포 | 다운로드 + 치환 | go:embed + struct 직렬화 | 안전성 향상 |
| update.py 수정 횟수 | 38회 (가장 불안정) | 재설계로 해결 | 구조적 해결 |

### 10.2 Top 5 아키텍처 개선점

**1. Hook 신뢰성 (ADR-006, ADR-012)**

28개의 Python hook 이슈를 컴파일된 바이너리 서브커맨드로 완전 제거. PATH 문제, 인코딩 문제, SIGALRM 문제, 프로토콜 불일치가 아키텍처적으로 불가능해진다. Hook Execution Contract 테스트가 CI에서 회귀를 차단한다.

- 효과: 21,535 LOC -> ~2,500 LOC (-88%), 28개 이슈 -> 0

**2. 안전한 업데이트 (ADR-007)**

파일 매니페스트 + 3-way merge로 파괴적 덮어쓰기를 방지. 4가지 provenance 유형(`template_managed`, `user_modified`, `user_created`, `deprecated`)으로 각 파일의 적절한 업데이트 전략을 자동 결정한다.

- 효과: 6개 덮어쓰기 이슈 해결, update.py 38회 수정 역사 종료

**3. Zero 템플릿 확장 (ADR-011)**

Go struct 직렬화(`json.MarshalIndent()`)로 JSON/YAML 생성. `{{VAR}}`, `${SHELL}` 등 문자열 치환 패턴을 완전 제거하여 4회의 settings.json 회귀 사이클을 원천 차단한다.

- 효과: 템플릿 치환 관련 6개 이슈 해결

**4. 단일 바이너리 배포**

Python 런타임, 가상환경, 패키지 매니저(pip/uv/pipx), 46개 hook 스크립트를 단일 `moai` 바이너리로 통합. PATH 문제, 인코딩 문제, 의존성 충돌이 구조적으로 불가능해진다.

- 효과: 설치 단계 1로 축소, 런타임 의존성 0, 디스크 70-85% 절감

**5. 시작 성능 향상**

컴파일된 바이너리로 인터프리터 오버헤드를 제거하여 10-20배 빠른 시작 시간을 달성. Hook 실행이 특히 크게 개선되어 (4개 Python 프로세스 -> 1개 Go 서브커맨드) 개발자의 코드 편집 루프 체감 속도가 향상된다.

- 효과: 200-500ms -> <50ms, Hook 총 실행 2-4s -> ~200ms

### 10.3 Top 3 트레이드오프

**1. TUI 풍부함 감소**

Python Rich 라이브러리는 매우 정교한 터미널 출력을 제공한다 (패널, 트리, 구문 강조, 마크다운 렌더링). Go의 Charmbracelet 생태계(lipgloss, bubbletea)도 강력하지만 Rich의 모든 기능을 1:1로 대체하지는 않는다. 특히 복잡한 레이아웃이나 Markdown 렌더링에서 차이가 있을 수 있다.

- 영향: 일부 출력 형식이 간소화될 수 있음
- 완화: Charmbracelet 생태계가 빠르게 발전 중, glamour 패키지로 Markdown 렌더링 가능

**2. 프로토타이핑 속도**

Go의 정적 타입 시스템은 컴파일 전에 더 많은 사전 설계를 요구한다. Python의 동적 타이핑은 빠른 프로토타이핑에 유리하다. 새로운 기능 추가 시 Go에서 인터페이스 설계, struct 정의, 에러 처리 등 더 많은 보일러플레이트가 필요하다.

- 영향: 신규 기능 개발 초기 속도가 다소 느릴 수 있음
- 완화: 타입 안전성이 장기적으로 유지보수 비용 절감, 리팩토링 안정성 향상

**3. 바이너리 크기**

Go 바이너리(15-30MB)는 Python wheel(~2MB)보다 크다. 템플릿 임베딩으로 추가 2-5MB가 더해진다. 그러나 이는 Python 런타임(~100MB)을 포함하지 않은 비교이므로 총 디스크 사용량은 Go가 적다.

- 영향: 다운로드 크기 증가 (15-30MB vs ~2MB wheel)
- 완화: `-ldflags "-s -w"`로 30-40% 크기 감소, UPX 압축 옵션, 총 디스크 사용은 Go가 적음

---

## 11. 결론 및 권장 사항

### 11.1 결론

Go 재작성의 타당성이 데이터 기반으로 확인되었다:

1. **LOC 감소**: 전체 프로젝트 80%, CLI 코드 67%, Hook 시스템 88% 축소
2. **이슈 해결**: 28개 Hook 이슈 + 6개 Update 이슈 + 6개 Template 이슈 = 40+ 이슈 아키텍처적 해결
3. **성능**: 시작 시간 90% 감소, Hook 실행 10-20x 빠름, 메모리 2-5x 효율적
4. **신뢰성**: 타입 안전성 필수화, 문자열 치환 제거, 계약 테스트 도입
5. **배포**: 단일 바이너리, 6개 플랫폼, zero 런타임 의존성

현재 구현 상태(~3%)에서 MVP까지 Phase 1(init, doctor)과 Phase 2(hook 시스템)의 우선 구현이 필요하다. 가장 큰 기술적 도전은 update 커맨드의 완전 재설계와 hook 시스템의 아키텍처 전환이다.

### 11.2 권장 구현 우선순위

| 순위 | 대상 | 근거 | 예상 효과 |
|-----|------|------|----------|
| **1** | Phase 1 완료 (init + doctor) | 기본 프로젝트 설정 가능, config/template 인프라 구축 | MVP 달성 |
| **2** | Phase 2 Hook 시스템 | 가장 많은 이슈 해결 (28개), 아키텍처 핵심 | 신뢰성 확보 |
| **3** | Phase 2 Status + Statusline | 일상 사용 커맨드, 사용자 경험 핵심 | 실용성 확보 |
| **4** | Phase 3 Update | 자체 업데이트 + 템플릿 동기화, 장기 유지보수 핵심 | 자립성 확보 |
| **5** | Phase 3 Switch | 백엔드 전환, config 시스템 활용 | 기능 보완 |
| **6** | Phase 4 Rank | 클라우드 통합, 커뮤니티 기능 | 생태계 확장 |
| **7** | Phase 5 Worktree | 병렬 개발 지원, 가장 복잡한 Git 작업 | 기능 완성 |

### 11.3 성공 지표

| 지표 | 목표 | 측정 방법 |
|------|------|----------|
| CLI LOC | <=2,200 (목표 2,161) | `wc -l internal/cli/**/*.go` |
| 총 프로젝트 LOC | <=18,000 | `wc -l **/*.go` |
| 시작 시간 | <=50ms | `time moai version` |
| 테스트 커버리지 | >=85% | `go test -cover ./...` |
| Hook 이슈 | 0 (28에서 감소) | GitHub Issues 추적 |
| settings.json 회귀 | 0 | Contract test suite |
| 파괴적 덮어쓰기 | 0 | Manifest-based 업데이트 검증 |
| 플랫폼 커버리지 | 6/6 | goreleaser CI matrix |
| 바이너리 크기 | <30MB | 릴리스 빌드 측정 |
| 메모리 (idle) | <20MB | 런타임 프로파일링 |

---

## Appendix A: Python CLI 커맨드 상세 사양

### A.1 init (854 LOC)

```
moai-adk init [PATH] [OPTIONS]
Options:
  --non-interactive, -y              Non-interactive mode (기본값 사용)
  --mode [personal|team]             프로젝트 모드 (default: personal)
  --locale [ko|en|ja|zh]             언어 설정
  --language TEXT                    프로그래밍 언어 (자동 감지)
  --force                           확인 없이 재초기화
```

**주요 기능:**

- InquirerPy 대화형 프롬프트: 프로젝트 이름, 언어, 프레임워크, 도메인 선택
- Git 저장소 감지 (`git rev-parse --git-dir`)
- Python 버전 확인 (3.10+ 필수)
- `.moai/` 디렉토리 구조 생성 (`config/sections/`, `memory/`, `specs/`, `project/`)
- `.claude/` 템플릿 배포 (agents, skills, commands, hooks, rules, output-styles)
- `CLAUDE.md` 생성 (프로젝트별 커스터마이징)
- `settings.json` 생성 (hook 설정, 출력 스타일)
- 언어별 설정 파일 초기화 (conversation_language, agent_prompt_language 등)
- Rich 콘솔 출력 (진행 상황 스피너, 결과 패널)
- Windows UTF-8 인코딩 보정

**알려진 이슈:**

- #315: 잘못된 디렉토리에 config 생성
- #283: config 미로드
- #304: 템플릿 변수 치환 실패
- #308: CLAUDE.md 깨짐
- #2: non-interactive 모드 실패
- #9: Git Bash에서 중단

### A.2 doctor (370 LOC)

```
moai-adk doctor [OPTIONS]
Options:
  --verbose, -v            상세 도구 버전 및 언어 감지 표시
  --fix                    누락 도구 수정 제안
  --export PATH            진단 결과 JSON 내보내기
  --check TEXT             특정 도구만 점검
  --check-commands         Slash command 로딩 문제 진단
  --shell                  Shell 및 PATH 설정 진단 (WSL/Linux)
```

**검사 항목:**

- Git 설치 및 버전
- Python 설치 및 버전
- Claude Code 설치
- Node.js 설치 및 버전
- uv/pip 패키지 매니저
- MoAI-ADK 버전
- PATH 설정 무결성
- `.moai/config/` 디렉토리 구조
- `.claude/settings.json` 유효성
- Hook 스크립트 존재 및 실행 권한
- Slash command 파일 유효성 (`--check-commands`)
- Shell 환경 진단 (`--shell`)

**출력 형식:** Rich 패널, 검사 항목별 성공/실패/경고 아이콘

### A.3 status (111 LOC)

```
moai-adk status
```

**표시 항목:**

- 프로젝트 이름
- 프로젝트 타입
- MoAI 버전
- config 상태 (로드됨/미로드)
- SPEC 진행 상황 요약

**출력 형식:** Rich 패널

### A.4 version (~50 LOC, __main__.py 내장)

```
moai-adk --version
```

- `click.version_option()`으로 구현
- `moai_adk/__init__.py`의 `__version__` 문자열 표시
- 서브커맨드 없이 실행 시 pyfiglet ASCII 아트 로고 표시

### A.5 update (3,162 LOC) -- 가장 복잡

```
moai-adk update [OPTIONS]
Options:
  --path PATH              프로젝트 경로 (default: .)
  --force                  백업 건너뛰기, 강제 업데이트
  --check                  버전만 확인 (업데이트 안 함)
  --templates-only         패키지 업그레이드 건너뛰기, 템플릿만 동기화
  --yes                    모든 확인 자동 수락 (CI/CD 모드)
  -c, --config             프로젝트 설정 편집 (init wizard 동일)
```

**기능 분해:**

1. 버전 확인: PyPI API 질의 또는 GitHub Releases 확인
2. 패키지 업그레이드: pip/uv/pipx 감지 및 업그레이드 실행
3. 템플릿 다운로드: 새 버전 템플릿 파일 가져오기
4. 파일 비교: 현재 파일 vs 새 템플릿 diff
5. 변수 치환: `{{HOOK_SHELL_PREFIX}}`, `{{PYTHON_PATH}}`, `{{PROJECT_DIR}}`
6. 덮어쓰기 결정: 파일별 덮어쓰기/건너뛰기 (provenance 추적 없음)
7. 백업 생성: `.moai/backup/` 디렉토리에 현재 상태 저장
8. 마이그레이션 실행: 버전별 migration 로직
9. 복원: 실패 시 백업에서 복원 시도

**알려진 이슈:**

- #246: settings 파일 업데이트 시 분실
- #187: GitHub workflows 덮어쓰기
- #162: init이 프로젝트 파일 덮어쓰기
- #236: `/moai:0-project`가 콘텐츠 삭제
- #318: 1.4 -> 1.14 템플릿 동기화 실패
- #319: 설치 후 모든 플러그인 실패
- #253: PyPI 미발견
- #296: PATH 이슈로 패키지 매니저 감지 실패
- 38회 수정 이력 (코드베이스 내 가장 자주 변경된 파일)

### A.6 cc / glm (325 LOC)

```
moai-adk claude             # Claude 백엔드로 전환
moai-adk cc                 # claude 별칭 (동일 기능)
moai-adk glm [API_KEY]      # GLM 백엔드로 전환 또는 API 키 업데이트
```

**참고**: `switch`는 커맨드로 존재하지 않으며, `claude`/`cc`와 `glm`이 각각 별도 top-level 커맨드로 동작한다.

**기능:**

- Claude 백엔드 전환: `.moai/config/sections/`의 LLM 설정 수정
- GLM 백엔드 전환: API 키 저장 + LLM 설정 수정
- API 키만 업데이트: `moai glm <api-key>`
- config 파일 원자적 쓰기

### A.7 rank (478 LOC, 7 서브커맨드)

```
moai-adk rank login                 # MoAI Cloud 인증
moai-adk rank status                # 랭킹 상태 표시
moai-adk rank logout                # 로그아웃
moai-adk rank sync                  # 메트릭 동기화
moai-adk rank exclude [PATTERN]     # 제외 패턴 추가
moai-adk rank include [PATTERN]     # 포함 패턴 추가
moai-adk rank register [--org TEXT] # 조직 등록
```

**기능:**

- MoAI Cloud API 인증 (로그인/로그아웃)
- 세션 메트릭 수집 및 제출
- 커뮤니티 리더보드 통합
- 제외/포함 패턴 관리
- 조직 등록

### A.8 worktree (2,034 LOC, 10 서브커맨드)

```
moai-wt new [SPEC-ID] [OPTIONS]     # 새 worktree 생성
  --branch TEXT                      # 브랜치 이름 지정
  --base TEXT                        # 기본 브랜치 (default: main)
moai-wt list                         # worktree 목록 표시
moai-wt go [NAME]                    # worktree 디렉토리로 이동
moai-wt remove [NAME] [OPTIONS]      # worktree 제거
  --force                            # 강제 제거
moai-wt status                       # worktree 상태 표시
moai-wt sync [NAME]                  # base 브랜치와 동기화
moai-wt clean [OPTIONS]              # 병합 완료된 worktree 정리
  --merged-only                      # 병합된 것만 정리
moai-wt recover [NAME]               # worktree 복구
moai-wt done [NAME]                  # worktree 완료 처리
moai-wt config [OPTIONS]             # worktree 설정 관리
  --get KEY                          # 설정 값 읽기
  --set KEY VALUE                    # 설정 값 쓰기
```

**구성:**

- `cli.py` (885 LOC): Click 기반 10개 서브커맨드 정의
- `manager.py` (662 LOC): Git worktree CRUD 로직, subprocess 기반 Git 명령
- `registry.py` (422 LOC): JSON 기반 worktree 레지스트리 관리
- `models.py` (65 LOC): WorktreeInfo, WorktreeConfig 데이터 모델
- 별도 엔트리포인트: `moai-worktree`, `moai-wt` (`pyproject.toml`에 등록)

### A.9 statusline (420 LOC CLI + 2,877 LOC 지원 모듈)

```
moai-adk statusline     # stdin JSON 컨텍스트 -> stdout 렌더링 텍스트
```

**지원 모듈 (11 파일, 3,297 LOC):**

| 파일 | LOC | 기능 |
|------|-----|------|
| `main.py` | 420 | statusline 데이터 빌드 |
| `version_reader.py` | 769 | 다양한 형식 버전 파일 읽기 |
| `renderer.py` | 468 | 터미널 렌더링 |
| `enhanced_output_style_detector.py` | 450 | 출력 스타일 감지 |
| `config.py` | 379 | statusline 설정 |
| `memory_collector.py` | 268 | 메모리 사용 수집 |
| `git_collector.py` | 190 | Git 상태 수집 |
| `update_checker.py` | 129 | 업데이트 확인 |
| `alfred_detector.py` | 105 | MoAI 설치 감지 |
| `metrics_tracker.py` | 78 | 메트릭 추적 |
| `__init__.py` | 41 | 패키지 초기화 |

---

## Appendix B: Go CLI 설계 타겟 상세

### B.1 init (~300 LOC)

**Cobra Command 구조:**

```go
var initCmd = &cobra.Command{
    Use:   "init [path]",
    Short: "Initialize a new MoAI project",
    Long:  "...",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runInit,
}

// Flags
initCmd.Flags().BoolP("non-interactive", "y", false, "Non-interactive mode")
initCmd.Flags().String("mode", "personal", "Project mode")
initCmd.Flags().String("locale", "", "Preferred language")
initCmd.Flags().String("language", "", "Programming language")
initCmd.Flags().Bool("force", false, "Force re-initialization")
```

**내부 모듈 의존성:**

- `internal/config/manager.go`: 설정 생성 및 저장
- `internal/template/deployer.go`: go:embed 템플릿 추출 + 배포
- `internal/template/settings.go`: settings.json 프로그래밍 방식 생성 (ADR-011)
- `internal/manifest/manifest.go`: 배포 파일 provenance 추적 시작 (ADR-007)
- `internal/ui/wizard.go`: bubbletea 기반 대화형 설정
- `internal/core/project/initializer.go`: 프로젝트 초기화 로직

**ADR 참조:** ADR-003 (embed), ADR-007 (manifest), ADR-011 (struct 직렬화)

### B.2 doctor (~150 LOC)

**Cobra Command 구조:**

```go
var doctorCmd = &cobra.Command{
    Use:   "doctor",
    Short: "Run system diagnostics",
    RunE:  runDoctor,
}

doctorCmd.Flags().BoolP("verbose", "v", false, "Show detailed diagnostics")
doctorCmd.Flags().Bool("fix", false, "Suggest fixes")
doctorCmd.Flags().String("export", "", "Export diagnostics to JSON")
doctorCmd.Flags().String("check", "", "Check specific tool only")
```

**내부 모듈 의존성:**

- `internal/config/manager.go`: config 유효성 검사
- `internal/core/project/checker.go`: 도구 존재 확인 (`exec.LookPath()`)
- `pkg/utils/logger.go`: 구조화 로깅

### B.3 status (~80 LOC)

**내부 모듈 의존성:**

- `internal/config/manager.go`: 프로젝트 설정 읽기
- `internal/core/git/manager.go`: Git 상태
- `internal/statusline/`: 상태 데이터 수집

### B.4 version (27 LOC) -- 구현 완료

**현재 구현:**

```go
var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Show version information",
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Printf("moai-adk %s (commit: %s, built: %s)\n",
            version.GetVersion(), version.GetCommit(), version.GetDate())
        return nil
    },
}
```

**빌드 타임 주입:** `go build -ldflags "-X .../pkg/version.Version=v1.0.0 -X .../pkg/version.Commit=abc123 -X .../pkg/version.Date=2026-02-03"`

### B.5 update (~800 LOC) -- 완전 재설계

**내부 모듈 의존성:**

- `internal/update/checker.go`: GitHub Releases API 버전 확인
- `internal/update/updater.go`: 바이너리 다운로드 + 원자적 교체
- `internal/update/rollback.go`: 실패 시 자동 롤백
- `internal/update/orchestrator.go`: 전체 워크플로우 조율
- `internal/manifest/manifest.go`: 파일 provenance 확인
- `internal/merge/three_way.go`: 3-way merge 실행
- `internal/merge/strategies.go`: 파일 유형별 merge 전략 선택
- `internal/template/deployer.go`: 새 템플릿 추출

**ADR 참조:** ADR-007 (manifest), ADR-008 (JSON 생성), ADR-011 (struct 직렬화)

### B.6 hook (~200 LOC CLI + ~2,500 LOC 핸들러)

**Cobra Command 구조:**

```go
var hookCmd = &cobra.Command{
    Use:   "hook",
    Short: "Hook management commands",
}

// Subcommands
hookCmd.AddCommand(sessionStartCmd)  // moai hook session-start
hookCmd.AddCommand(preToolCmd)       // moai hook pre-tool
hookCmd.AddCommand(postToolCmd)      // moai hook post-tool
hookCmd.AddCommand(sessionEndCmd)    // moai hook session-end
hookCmd.AddCommand(stopCmd)          // moai hook stop
hookCmd.AddCommand(compactCmd)       // moai hook compact
hookCmd.AddCommand(listCmd)          // moai hook list
```

**내부 모듈 의존성:**

- `internal/hook/registry.go`: 핸들러 등록 + 디스패치
- `internal/hook/protocol.go`: JSON stdin/stdout 프로토콜
- `internal/hook/contract.go`: 실행 계약 (ADR-012)
- `internal/hook/session_start.go` ~ `compact.go`: 6개 이벤트 핸들러

**ADR 참조:** ADR-006 (hooks as subcommands), ADR-012 (execution contract)

### B.7 cc / glm (~120 LOC)

**Cobra Command 구조:**

```go
// internal/cli/cc.go
var ccCmd = &cobra.Command{
    Use:   "cc",
    Short: "Switch to Claude backend",
    RunE:  runCC,
}

// internal/cli/glm.go
var glmCmd = &cobra.Command{
    Use:   "glm [api-key]",
    Short: "Switch to GLM backend",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runGLM,
}
```

**참고**: Python과 동일하게 별도 top-level 커맨드로 등록. `switch` 커맨드는 사용하지 않는다.

### B.8 rank (~200 LOC)

**Cobra Command 구조:**

```go
var rankCmd = &cobra.Command{
    Use:   "rank",
    Short: "Performance ranking commands",
}

// 7 subcommands
rankCmd.AddCommand(loginCmd, statusCmd, logoutCmd, syncCmd,
                   excludeCmd, includeCmd, registerCmd)
```

**내부 모듈 의존성:**

- `internal/rank/client.go`: HTTP 클라이언트 (stdlib `net/http`)
- `internal/rank/auth.go`: 시스템 keyring 통합
- `internal/rank/config.go`: 랭킹 설정

### B.9 worktree (~400 LOC)

**Cobra Command 구조:**

```go
// internal/cli/worktree/
var worktreeCmd = &cobra.Command{
    Use:     "worktree",
    Aliases: []string{"wt"},
    Short:   "Git worktree management",
}

// 7+ subcommands (10에서 일부 통합)
worktreeCmd.AddCommand(newCmd, listCmd, switchCmd, removeCmd,
                       statusCmd, syncCmd, cleanCmd)
```

**내부 모듈 의존성:**

- `internal/core/git/manager.go`: go-git Repository 인터페이스
- `internal/core/git/branch.go`: 브랜치 관리
- `internal/core/git/` + `exec.Command("git", "worktree", ...)`: 시스템 Git fallback

**ADR 참조:** ADR-007 (go-git + system Git fallback)

---

## Appendix C: 이슈 매핑 (Python Issues -> Go Solutions)

### C.1 Hook 이슈 (28건)

가장 많은 이슈를 유발한 영역. ADR-006 컴파일 핸들러로 전부 해결.

| 이슈 번호 | 문제 | Python 근본 원인 | Go 해결 방식 |
|----------|------|-----------------|-------------|
| #278 | PyYAML 미발견 | Python 런타임 의존 | 컴파일 바이너리, 의존성 없음 |
| #288 | uv 버전 감지 실패 | 패키지 매니저 의존 | 단일 바이너리, 패키지 매니저 불필요 |
| #269 | Import 에러 | Python 모듈 경로 | 컴파일 시 모든 의존성 포함 |
| #260 | Hook import 실패 | Python PATH 이슈 | `moai` in PATH, 경로 해석 불필요 |
| #259 | Windows 혼합 구분자 | 문자열 경로 조작 | `filepath.Join()` OS별 처리 |
| #161 | `$CLAUDE_PROJECT_DIR` 미설정 | 환경변수 의존 | 환경변수 불필요 |
| #5 | MODULE_NOT_FOUND | Node.js/Python 모듈 | 컴파일 바이너리 |
| #129 | SIGALRM Windows 미지원 | Unix 전용 신호 | `context.WithTimeout()` |
| #249 | cp1252 인코딩 | Python 기본 인코딩 | Go 네이티브 UTF-8 |
| #25 | 무한 대기 | 셸 환경 의존 | 컴파일 바이너리, 셸 불필요 |
| #265 | settings.json 형식 비호환 | 문자열 치환 생성 | `json.MarshalIndent()` |
| #207 | Hook 중복 실행 | 동적 등록 | 컴파일 시 정적 등록 |
| #263 | 성공인데 에러 표시 | exit code 처리 | 명시적 exit code 핸들링 |
| #245 | Hook에서 config 로드 실패 | 별도 프로세스 | 동일 바이너리 = 동일 config 로더 |
| #243 | Hook config 경로 혼란 | 환경변수 의존 | 프로젝트 루트 자동 감지 |
| #231 | Docker 의존성 미발견 | Python 환경 | 단일 바이너리, 의존성 없음 |
| #66 | Hook 프리즈 | 셸 래퍼 의존 | 컴파일 바이너리, 타임아웃 내장 |
| #107 | Hook 행 | 프로세스 간 통신 | 동일 프로세스 실행 |
| #314 | Emoji 렌더링 실패 | 인코딩 문제 | Go 네이티브 UTF-8 |
| #286 | 인코딩 에러 | Windows cp949 | Go 네이티브 UTF-8 |
| #45 | PowerShell Hook 실패 | 셸 호환성 | 셸 불필요 |
| #31 | 스크립트 에러 | Python 스크립트 구문 | 컴파일 바이너리 |
| 나머지 6건 | 기타 Hook 관련 | 다양한 환경 의존 | 아키텍처적 제거 |

### C.2 Config/Init 이슈 (25건)

Viper 기반 타입 안전 struct + manifest로 해결.

| 이슈 카테고리 | 이슈 수 | Go 해결 방식 |
|-------------|---------|-------------|
| Config 경로 혼란 | 8 | Viper 자동 경로 해석, 프로젝트 루트 자동 감지 |
| 템플릿 변수 치환 | 6 | ADR-011 struct 직렬화, 문자열 치환 금지 |
| 파괴적 init/update | 6 | ADR-007 manifest provenance 추적 |
| 컨텍스트 크기 | 3 | 템플릿 최적화 (별도 redesign-plan 참조) |
| Non-interactive 환경 | 2 | TTY 감지 + `--non-interactive` flag |

### C.3 Template 이슈 (6건)

ADR-011 Zero Runtime Template Expansion으로 해결.

| 이슈 번호 | 문제 | Go 해결 방식 |
|----------|------|-------------|
| #304 | 변수 치환 실패 | `json.MarshalIndent()` 사용, 변수 치환 없음 |
| #308 | CLAUDE.md 깨짐 | Go text/template strict mode |
| #309 | 잘못된 언어 설정 | 타입 안전 LanguageConfig struct |
| 나머지 3건 | 경로/인코딩 관련 | `filepath.Clean()` + UTF-8 |

### C.4 Update 이슈 (15건)

ADR-007 Manifest + 3-Way Merge + Self-Update로 해결.

| 이슈 카테고리 | 이슈 수 | Go 해결 방식 |
|-------------|---------|-------------|
| 파괴적 업데이트 | 5 | Manifest provenance, 3-way merge |
| 버전 감지 취약성 | 4 | GitHub Releases API, build-time ldflags |
| 패키지 매니저 의존 | 4 | Self-update, 패키지 매니저 불필요 |
| 롤백 부재 | 2 | 원자적 바이너리 교체 + 자동 롤백 |

### C.5 플랫폼 이슈 (11건)

`CGO_ENABLED=0` + goreleaser로 해결.

| 이슈 카테고리 | 이슈 수 | Go 해결 방식 |
|-------------|---------|-------------|
| SIGALRM | 1 (치명적) | `context.WithTimeout()`, SIGALRM 불사용 |
| 경로 구분자 | 2 | `filepath.Join()`, OS별 자동 처리 |
| 인코딩 | 3 | Go 네이티브 UTF-8 |
| 셸 비호환 | 3 | 셸 래퍼 불필요, 컴파일 바이너리 |
| 설치 | 2 | 단일 바이너리 다운로드 |

### C.6 이슈 해결 요약

| 카테고리 | Python 이슈 수 | Go 아키텍처적 해결 | 해결률 |
|---------|-------------|-----------------|-------|
| Hook 시스템 | 28 | ADR-006 컴파일 핸들러 | 100% |
| Config/Init | 25 | Viper + typed struct + manifest | ~90% |
| Update/Migration | 15 | Manifest + 3-way merge + self-update | ~95% |
| 플랫폼 호환 | 11 | CGO_ENABLED=0, goreleaser | 100% |
| 성능/메모리 | 7 | 컴파일 바이너리, goroutine | ~85% |
| 템플릿 | 6 | ADR-011 struct 직렬화 | 100% |
| 기타 | 81 | 다양한 개선 | ~70% |
| **합계** | **173** | | **~89%** |

173개 이슈 중 약 89%에 해당하는 이슈가 Go 재작성의 아키텍처적 결정에 의해 해결 가능한 것으로 분석된다. 이는 redesign-report.md의 "95% of actionable issues (89/94)" 분석과 일치한다.

---

## Appendix D: ADR (Architecture Decision Record) 요약

Go 재작성에 관련된 주요 아키텍처 결정:

| ADR | 제목 | 핵심 결정 | 영향 범위 |
|-----|------|---------|----------|
| ADR-001 | Modular Monolithic | 단일 바이너리, Go 패키지로 도메인 경계 | 전체 아키텍처 |
| ADR-003 | Interface-Based DDD | 컴파일 타임 계약, 모킹 가능 의존성 | 모든 도메인 모듈 |
| ADR-006 | Hooks as Binary Subcommands | Python 스크립트 -> `moai hook <event>` | 28개 이슈 해결 |
| ADR-007 | File Manifest Provenance | 파일 출처 추적, 3-way merge | update, init |
| ADR-008 | Programmatic JSON Generation | `json.MarshalIndent()`, 문자열 치환 금지 | settings.json |
| ADR-011 | Zero Runtime Template Expansion | struct 직렬화, 동적 토큰 금지 | template 전체 |
| ADR-012 | Hook Execution Contract | 공식 보증/비보증 정의, 계약 테스트 | hook 시스템 |

---

## Appendix E: 용어 사전

| 용어 | 설명 |
|------|------|
| ADR | Architecture Decision Record, 아키텍처 결정 기록 |
| bubbletea | Go TUI 프레임워크 (Elm architecture, Charmbracelet) |
| CGO_ENABLED=0 | Go 빌드 시 C 의존성 제거 flag |
| Click | Python CLI 프레임워크 (Pallets 프로젝트) |
| Cobra | Go CLI 프레임워크 (spf13) |
| go:embed | Go 빌드 시 파일을 바이너리에 임베딩하는 directive |
| goreleaser | Go 크로스 컴파일 + 릴리스 자동화 도구 |
| Hook Execution Contract | Hook 실행 환경의 공식 보증 사양 (ADR-012) |
| ldflags | Go 빌드 시 링커 flag를 통한 변수 주입 |
| lipgloss | Go 터미널 스타일링 라이브러리 (Charmbracelet) |
| LOC | Lines of Code, 코드 줄 수 |
| Manifest | 배포 파일 provenance 추적 메타데이터 (.moai/manifest.json) |
| pflag | POSIX 호환 flag 파싱 라이브러리 (Cobra 의존) |
| Provenance | 파일의 출처 및 소유권 분류 |
| Rich | Python 터미널 UI 라이브러리 |
| SIGALRM | Unix 알람 신호 (Windows 미지원) |
| struct 직렬화 | Go struct를 JSON/YAML로 변환하는 방식 |
| TRUST 5 | Tested, Readable, Unified, Secured, Trackable 품질 프레임워크 |
| Viper | Go 설정 관리 라이브러리 (spf13) |
| 3-Way Merge | base + current + updated를 비교하는 merge 알고리즘 |

---

*본 보고서는 moai-adk-go 프로젝트의 `.moai/project/` 디렉토리에 위치한 design.md, product.md, tech.md, structure.md, redesign-report.md, template-redesign-plan.md을 기반으로 작성되었다.*

*Python 소스코드 분석은 `moai_adk/` 디렉토리의 220개 Python 파일(88,319 LOC)과 Go 소스코드 11개 파일(306 LOC)을 대상으로 수행되었다.*
