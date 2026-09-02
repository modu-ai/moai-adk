# moai-adk-go 아키텍처 개요

> 이 문서는 `/moai codemaps --force`로 자동 생성된 아키텍처 설명서입니다.

**모듈**: `github.com/modu-ai/moai-adk`  
**Go 버전**: go 1.26.4  
**코드 규모**: 1074개 non-test Go 소스 파일 + 1714개 테스트 파일 (~229.4k non-test LOC, 2026-09-02 실측)  
**패키지 수**: internal 최상위 64개 디렉터리 (`go list ./...` 기준 internal 패키지 131개) + 2 pkg (`models`, `version`) + 1 cmd

---

## 시스템 개요

moai-adk-go는 Claude Code 내에서 AI 기반 개발 워크플로우를 오케스트레이션하는 통합 Go 바이너리입니다. **4계층 아키텍처**로 설계되었으며, 프레젠테이션 계층(CLI/TUI)에서 인프라 계층(Git/LSP)까지 명확한 책임 분리를 구현합니다.

### 아키텍처 4계층

| 계층 | 책임 | 주요 패키지 |
|------|------|-----------|
| **프레젠테이션** | CLI 명령, 터미널 UI, HTTP 인터페이스 | `cmd/moai`, `internal/cli` (264 non-test, 하위 포함), `internal/tui`, `internal/statusline`, `internal/web` |
| **비즈니스/도메인** | 개발 워크플로우, SPEC 라이프사이클, 정책 | `internal/spec`, `internal/workflow`, `internal/loop`, `internal/harness`, `internal/constitution`, `internal/permission`, `internal/merge`, `internal/navigator/*`, `internal/kanban`, `internal/epic` |
| **인프라** | Git 추상화, 템플릿 배포, 설정, 훅, 세션 | `internal/core/git`, `internal/template`, `internal/config`, `internal/hook`, `internal/session`, `internal/lsp/*`, `internal/mx`, `internal/graph`, `internal/settings`, `internal/guardstate`, `internal/guardliveness`, `internal/binlag` |
| **계측/지원** | 성능 측정, LSP 통합, 셸, 재시도 | `internal/measure`, `internal/astgrep`, `internal/shell`, `internal/resilience`, `internal/timing` |

---

## 핵심 패키지 역할

### Presentation (프레젠테이션)
- **cmd/moai**: 바이너리 진입점 → `cli.Execute()`
- **internal/cli** (264 non-test 파일, 하위 포함): Cobra 커맨드 트리, composition root, root 등록 60건 (202 non-test `.AddCommand()` 호출, 33개 파일 분산). `moai codex` 런처(`codex_launcher.go`)와 세 기동 형태(맨몸·`cli`·`app`)가 통과하는 init-offer 게이트(`codex_init.go`)·AGENTS.md ↔ CLAUDE.md 지시 계약(`codex_contract.go`) 포함. 하위 패키지: `agentlint`, `harness`, `pr`, `preference`, `printer`, `specid`, `taskledger`, `uikit`, `update/{backup,deploy,merge,plan,report}`, `wizard`, `worktree`
- **internal/tui**: Catppuccin 색상 토큰, Box/Pill/Table/Status 컴포넌트, 테마 선택, golden 테스트 스냅샷
- **internal/statusline**: Claude Code 상태 렌더러, 3L/5L 레이아웃, pluggable 데이터 제공자, GLM 컨텍스트 윈도우 해석(`memory.go`의 `glmContextWindows`)
- **internal/web**: loopback HTTP 콘솔, Templ 컴파일 핸들러, 5s 드레인 종료
- **pkg/version**: 빌드타임 버전/commit/date (ldflags)

### Business/Domain (비즈니스 영역)
- **pkg/models**: 공유 config 타입 — ProjectType, DevelopmentMode, ProjectConfig
- **internal/spec** (31 non-test 파일): SPEC 라이프사이클 — Linter (19 규칙: 15 단일-문서 + 3 크로스-SPEC + 1 registry), Audit(), ClassifyEra(), DetectDrift(), ClassifyPRTitle()
- **internal/workflow**: Plan-Run-Sync 워크트리 오케스트레이션
- **internal/loop** (6 non-test 파일): 진단 피드백 루프 — LoopController, DecisionEngine, RalphEngine
- **internal/harness** (82 non-test 파일): 하네스 자가학습 — Observer, Learner (4-tier), Applier, v4manifest, routing(위임 관측 원장), delegationmap(위임 맵 분석기), 5-phase safety. 하위 패키지: `capture`, `cluster`, `curator`, `delegationmap`, `harnessrun`, `proposalgen`, `router`, `routing`, `safety`, `seeds`, `throttle`, `tier`, `v4manifest`
- **internal/navigator** (53 non-test 파일): BAS(Blueprint-Anchored Synchronization) 코드 탐색 — 6 서브패키지 `astx`(AST 추출 시임), `detect`, `fix`, `route`, `sync`(3개 Navigator 체인 통합), `tiers`(4-tier 주소 맵 오버레이, tiers.json OVERLAY)
- **internal/kanban** (33 non-test 파일): 칸반 백로그 큐 엔진 — `moai todo` 명령의 기계적 텍스트 분석기(`backlog_analysis.go`), 큐 저장소, 관계 분석
- **internal/epic** (5 non-test 파일): 디스크 기반 에픽 진행 생산자 — `moai epic status <prefix>`
- **internal/permission** (6 non-test 파일): 8-tier 권한 스택, 5 모드, bubble 모드
- **internal/merge**: 3-way 파일 병합 (ADR-008), 사용자 커스터마이징 보존
- **internal/constitution**: 동결/진화 구역 모델, 5단계 병합 안전 파이프라인

### Infrastructure (인프라)
- **internal/core/git**: exec 기반 Git 추상화, Repository/BranchManager 인터페이스
- **internal/core/project**: FindProjectRoot() ANCHOR — `.moai/` 발견
- **internal/template**: go:embed Template-First (`embed.go`가 직접 `//go:embed all:templates` — 별도 `embedded.go` 생성 없음), Deployer, Renderer, profile_matrix (33-cell). 하위: `agentemit`(에이전트 출력 에밋), `scripts`
- **internal/config** (48 non-test 파일, 하위 포함): 계층화 YAML (env > yaml > defaults), section loader chain. 하위: `atomicfile`, `toolpolicy`
- **internal/hook** (128 non-test 파일, 하위 포함): 컴파일된 훅 시스템, 30개 Claude Code 이벤트, branch-state guard, JSON 디스패치. 하위: `handoff`, `memo`, `memo/taxonomy`, `mx`, `mx/complexity`, `perf`, `quality`, `security`, `testutil`, `trace`
- **internal/session** (22 non-test 파일): 다중 세션 레지스트리, active-sessions.json, Heartbeat/Purge
- **internal/sessionmsg** (7 non-test 파일): 단일 머신 세션 메시징 브로커 코어 — Claude↔Codex 양방향 메시징 (`session_msg_*` MCP 도구의 백엔드)
- **internal/lsp** (8 sub-packages): aggregator, cache, config, core, gopls, hook, subprocess, transport — JSON-RPC 클라이언트, 16-언어 자동감지, 회로 차단기
- **internal/graph** (14 non-test 파일): 코드베이스 엣지 산출물(edges.jsonl) + 3-레이어 신선도 게이트(`moai graph check`)·쿼리 시점 갱신·인용 앵커·MCP 코드 쿼리 엔진(graph_shortest_path 포함). 하위: `symbol`(astx 추출 시임)
- **internal/mx**: @MX 태그 스캐너, FanInCounter, 사이드카 JSON 인덱스, `IsDescribedWorthy()` (codemaps 신선도 게이트가 비교하는 "설명 대상" Go 소스 판정)
- **internal/settings** (10 non-test 파일): settings.json / settings.local.json 헬퍼 — 하위 `agentfm`(에이전트 frontmatter), `yamlpatch`
- **internal/paths**: `~/.moai` 디렉터리 트리의 단일 해석 지점
- **internal/guardstate**: 가드 이벤트를 8행 상태표로 분류하는 상태 모델 — `Classification`(닫힌 7값 어휘, `ClassOK`가 유일한 clean 값), `Classify()`/`Evaluate()`/`Produce()`가 판정·집계·산출물 조립을 각각 담당
- **internal/guardliveness**: guardstate 판정 결과를 운영자에게 "묻지 않아도" 드러내는 표면화 계층 — 3-clause 계약(entry당 정확히 1개 분류, clean 값 1개, 결과가 그 값을 기계가 읽을 수 있게 표시)만 소비하고 어휘 자체는 소유하지 않음
- **internal/binlag**: 설치된 `moai` 바이너리가 현재 소스 트리보다 뒤처졌는지 판정하는 단일 비교 지점 — `moai doctor`와 SessionStart 어드바이저 양쪽이 같은 구현을 공유
- **internal/glmcred**: GLM API 자격증명(`~/.moai/.env.glm`) 쓰기·판독의 단일 구현

---

## 매우 높은 팬-인 패키지 (ANCHOR)

이 패키지들의 변경은 광범위한 영향을 미칩니다 (측정: `go list -f '{{range .Imports}}'` 기반 직접 non-test 임포트 수, 2026-09-02 실측):

| 패키지 | 팬-인 (직접 임포트 패키지 수) | 역할 |
|--------|----------|------|
| `internal/config` | 27 | YAML 설정 SSOT — CLI가 모든 패키지에 주입 |
| `internal/defs` | 17 | 디렉토리 레이아웃 상수 |
| `internal/paths` | 11 | `~/.moai` 경로 단일 해석점 |
| `internal/atomicfile` | 11 | 원자적 파일 쓰기 (write-temp + rename) |
| `internal/lsp` | 10 | LSP 클라이언트 집합 |
| `pkg/models` | 9 | Config 타입 중심 |
| `internal/tui` | 9 | 터미널 UI 컴포넌트 |
| `internal/template` | 9 | 배포 엔진 — CLI/init/update/migration |
| `internal/harness` | 8 | 하네스 학습 서브시스템 |
| `internal/execerr` | 8 | 서브프로세스 실패의 출력 보존 래퍼 |
| `internal/cli` | 팬-아웃 93 (내부 패키지 임포트) | Composition root — 모든 subcommand 라우팅 |

---

## 의존성 다이어그램

```
cmd/moai
    └─→ internal/cli (composition root, 내부 패키지 93개 임포트)
        ├─→ internal/core/{git,project,quality}
        ├─→ internal/{config,template,manifest,hook}
        ├─→ internal/{spec,workflow,loop,harness}
        ├─→ internal/{lsp/*,mx,astgrep,navigator/*}
        ├─→ internal/{graph,graph/symbol,kanban,epic}
        ├─→ internal/{session,sessionmsg,permission,paths}
        ├─→ internal/{github,update,profile,settings}
        └─→ pkg/{models,version}

순환 의존성: 없음 (검증됨 — `go build ./...` 성공이 곧 순환 부재의 기계적 증거)
```

---

## 주요 데이터 흐름

### 1. 템플릿 배포
```
EmbeddedTemplates() → Deployer.Deploy()
  → Renderer (엄격 모드, missing key error)
    → 원자적 쓰기 (temp+rename)
      → Manifest.Track() (3중 해시)
```

### 2. SPEC 라이프사이클
```
CLI (/moai plan/run/sync)
  → spec.Linter (frontmatter + ownership)
    → spec.ClassifyEra (grandfather vs V3R6 현대식)
      → spec.Audit/DetectDrift (SyncStatusDrift)
        → spec.ClassifyPRTitle (git 유추)
```

### 3. 훅 이벤트 분배
```
Claude Code → handle-<event>.sh
  → moai hook <event> (stdin JSON)
    → Registry.Dispatch()
      → Handler chain (순차, 2 오류 시 단락)
        → JSON + exit-code (stdout)
```

### 4. Ralph 진단 루프
```
LoopController.Start()
  → FeedbackGenerator (go test/vet + LSP)
    → RalphEngine.Decide (계속/수렴/중단/검토)
      → iterate
```

### 5. 권한 해석
```
PreToolUse hook
  → permission.Resolver.Resolve()
    → 8-tier stack (policy→...→deny)
      → allow/deny/ask
```

### 6. codex 런치 게이트
```
moai codex (맨몸) / cli / app
  → classifyCodexWiring (배선 판정 — 런처의 단일 판정 소비)
    → wired: 기동 (직접 또는 tmux spawn)
    → 불완전: 상태·처방 보고
      → 비대화형: 보고 후 종료 (rc 1, 프롬프트 없음)
      → 거절: 기록 없이 종료 (rc 130 — 취소)
      → 수락: codexwiring.Wire (배선 생성, agent=codex)
        → AGENTS.md ↔ CLAUDE.md 지시 계약 (봉쇄 선행, temp+rename)
          → 기동
```

---

## 진입점 (Entry Points)

### 바이너리 진입점
```bash
cmd/moai/main() → cli.Execute() → cobra rootCmd.Execute()
```

### Composition Root
```go
cli.InitDependencies() // 모든 서브시스템 와이어링
```

### CLI 명령 (root 등록 60건 — 33개 파일 분산, help 노출 52개 + hidden `statusline` + cobra 자동 `help`/`completion`)
- **프로젝트**: `init`, `status`, `doctor`, `update`, `migrate`, `pr`
- **런처**: `cc`, `glm`, `cg`, `codex`
- **SPEC/루프**: `spec` (audit/lint/close), `plan`, `loop`, `goal`, `gate`
- **인프라**: `hook`, `session`, `worktree`, `migration`, `integration`, `graph`, `chain`
- **개발 도구**: `mx`, `ast-grep`, `ast-edit`, `handoff`, `verify`, `todo`, `epic`, `memory`, `model`, `tokens`
- **설정/거버넌스**: `config`, `constitution`, `tool-policy`, `preference`, `telemetry`, `clean`, `inventory`
- **기타**: `github`, `github` 워크플로우, `lsp`, `mcp`, `mcp-server`, `research`, `agent`, `workflow`, `web`, `version`

### 훅 진입점
```bash
moai hook <event>  # SessionStart, PostToolUse, Stop, etc.
```

---

## 아키텍처 특징

### 1. 인터페이스 우선 설계
- 모든 도메인 모듈이 인터페이스 노출
- 구현은 패키지 내부에 숨김
- Hexagonal Architecture (Ports & Adapters)

### 2. 의존성 주입 (Composition Root)
- `internal/cli/deps.go`에서 모든 타입 인스턴스화
- CLI 명령은 인터페이스만 참조

### 3. 임베드된 템플릿 파일시스템
- go:embed로 모든 프로젝트 템플릿 컴파일
- 배포 시에 원자적 쓰기
- 3-way 병합으로 사용자 커스터마이징 보존

### 4. 훅 레지스트리 패턴
- Claude Code JSON 이벤트 → stdin 수신
- Registry가 30개 EventType 핸들러로 디스패치
- 각 핸들러는 `Handler` 인터페이스 준수

### 5. 멀티 LLM 실행 모드
- `moai cc`: Claude 전용
- `moai glm`: GLM 전용
- `moai cg`: Claude leader + GLM teammates
- `moai codex`: Codex CLI/데스크톱 앱 런처 — 인자 없이 부르면 Codex CLI 를 기동하고(`cli` 는 그 명시 별칭, `app` 은 데스크톱 앱, `--spawn` 은 새 tmux 창), 준비 상태 리드아웃은 `moai codex status` 로 분리돼 아무것도 띄우지 않으며, 배선이 불완전한 프로젝트에서 기동 시 init-offer 게이트가 배선 생성을 제안
- GLM tier-models 테이블: 기본 코딩 모델은 `glm-5.3-flash` (대부분 tier), `glm-5.3`는 Fable tier

---

## 상세 참고 문서

이 개요는 빠른 이해를 제공합니다. 더 깊은 분석은 다음을 참고하세요:

- **modules.md**: internal 주요 패키지의 함수/타입/역할 상세 설명
- **dependencies.md**: Mermaid 패키지 의존도 그래프 + 팬-인/팬-아웃 정량화
- **entry-points.md**: root CLI 명령 + 훅 진입점 목록
- **data-flow.md**: 9가지 주요 플로우 시각화 (Mermaid)
- **docs-truth.md**: 문서 검증을 위한 수동 유지보수 사실 체크리스트

---

**생성**: `/moai codemaps --force`로 자동 생성  
**검증**: 순환 의존성 0개, 모든 패키지 경로 존재 확인
