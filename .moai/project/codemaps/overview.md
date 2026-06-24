# moai-adk-go 아키텍처 개요

> 이 문서는 `/moai codemaps --force`로 자동 생성된 아키텍처 설명서입니다.

**모듈**: `github.com/modu-ai/moai-adk`  
**Go 버전**: go 1.26.4  
**코드 규모**: ~645개 Go 소스 파일 (테스트 제외)  
**패키지 수**: 45 internal 디렉터리 = 44 runtime 패키지 + 1 test-only 패키지 (`internal/skills`, 프로덕션 코드 없음) + 2 pkg + 1 cmd

---

## 시스템 개요

moai-adk-go는 Claude Code 내에서 AI 기반 개발 워크플로우를 오케스트레이션하는 통합 Go 바이너리입니다. **4계층 아키텍처**로 설계되었으며, 프레젠테이션 계층(CLI/TUI)에서 인프라 계층(Git/LSP)까지 명확한 책임 분리를 구현합니다.

### 아키텍처 4계층

| 계층 | 책임 | 주요 패키지 |
|------|------|-----------|
| **프레젠테이션** | CLI 명령, 터미널 UI, HTTP 인터페이스 | `cmd/moai`, `internal/cli` (241파일), `internal/tui`, `internal/statusline`, `internal/web` |
| **비즈니스/도메인** | 개발 워크플로우, SPEC 라이프사이클, 정책 | `internal/spec`, `internal/workflow`, `internal/loop`, `internal/harness`, `internal/constitution`, `internal/permission`, `internal/merge` |
| **인프라** | Git 추상화, 템플릿 배포, 설정, 훅, 세션 | `internal/core/git`, `internal/template`, `internal/config`, `internal/hook`, `internal/session`, `internal/lsp/*`, `internal/mx` |
| **계측/지원** | 성능 측정, LSP 통합, 다국어 | `internal/measure`, `internal/astgrep`, `internal/i18n`, `internal/shell` |

---

## 핵심 패키지 역할

### Presentation (프레젠테이션)
- **cmd/moai**: 바이너리 진입점 → `cli.Execute()`
- **internal/cli** (241파일): Cobra 커맨드 트리, composition root, ~48개 내부 패키지 가져옴, 50+개 subcommand 라우팅
- **internal/tui**: Catppuccin 색상 토큰, Box/Pill/Table/Status 컴포넌트, 테마 선택
- **internal/statusline**: Claude Code 상태 렌더러, 3L/5L 레이아웃, pluggable 데이터 제공자
- **internal/web**: loopback HTTP 콘솔, Templ 컴파일 핸들러, 5s 드레인 종료
- **pkg/version**: 빌드타임 버전/commit/date (ldflags)

### Business/Domain (비즈니스 영역)
- **pkg/models**: 공유 config 타입 (매우 높은 팬-인) — ProjectType, DevelopmentMode, ProjectConfig
- **internal/spec** (41파일): SPEC 라이프사이클 — Linter (13+3 규칙), Audit(), ClassifyEra(), DetectDrift(), ClassifyPRTitle()
- **internal/workflow**: Plan-Run-Sync 워크트리 오케스트레이션
- **internal/loop** (18파일): 진단 피드백 루프 — LoopController, DecisionEngine, RalphEngine
- **internal/harness** (64파일): 하네스 학습 — Observer, Learner (4-tier), Applier, 5-phase safety
- **internal/permission** (18파일): 8-tier 권한 스택, 5 모드, bubble 모드
- **internal/merge**: 3-way 파일 병합 (ADR-008), 사용자 커스터마이징 보존
- **internal/constitution**: 동결/진화 구역 모델, 5단계 병합 안전 파이프라인

### Infrastructure (인프라)
- **internal/core/git**: exec 기반 Git 추상화, Repository/BranchManager 인터페이스
- **internal/core/project**: FindProjectRoot() ANCHOR — `.moai/` 발견
- **internal/template** (75파일): go:embed Template-First, Deployer, Renderer, TemplateContext
- **internal/config** (61파일): 계층화 YAML (env > yaml > defaults), 20+ 섹션 구조
- **internal/hook** (143파일): 컴파일된 훅 시스템, 28+ Claude Code 이벤트, JSON 디스패치, exit 0/2 의미
- **internal/session** (25파일): 다중 세션 레지스트리, active-sessions.json, Heartbeat/Purge
- **internal/lsp** (12 sub-packages): JSON-RPC 클라이언트, Gopls 브리지, 집계기, 캐시, 회로 차단기
- **internal/mx**: @MX 태그 스캐너, FanInCounter, 사이드카 JSON 인덱스

---

## 매우 높은 팬-인 패키지 (ANCHOR)

이 패키지들의 변경은 광범위한 영향을 미칩니다:

| 패키지 | 팬-인 수준 | 역할 |
|--------|----------|------|
| `pkg/models` | Very High (45+) | Config 타입 중심 — 모든 패키지가 설정 구조 가져옴 |
| `pkg/version` | High (30+) | 버전 정보 |
| `internal/core/git` | High (35+) | Git 추상화 — workflow/spec/session 사용 |
| `internal/config` | Very High (48+) | YAML 설정 SSOT — CLI가 모든 패키지에 주입 |
| `internal/core/project` | High (28+) | FindProjectRoot() — 프로젝트 루트 발견 (everywhere) |
| `internal/cli` | Very High (50+ import) | Composition root — 모든 subcommand 라우팅 |
| `internal/foundation` | High (32+) | 언어 registry — 지원 언어 쿼리 |
| `internal/hook` | High (18+) | 훅 디스패치 — CLI/test/session |
| `internal/template` | High (15+) | 배포 엔진 — CLI/init/update/migration |

---

## 의존성 다이어그램

```
cmd/moai
    └─→ internal/cli (composition root, 48개 패키지 가져옴)
        ├─→ internal/core/{git,project,quality}
        ├─→ internal/{config,template,manifest,hook}
        ├─→ internal/{spec,workflow,loop,harness}
        ├─→ internal/{lsp/*,mx,astgrep}
        ├─→ internal/{session,state,permission}
        ├─→ internal/{github,update,profile}
        └─→ pkg/{models,version}

순환 의존성: 없음 (검증됨)
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

### CLI 명령 (~50+)
- **프로젝트**: `init`, `update`, `doctor`, `config`, `web`
- **SPEC**: `spec` (audit/lint/close)
- **워크플로우**: `plan`, `run`, `sync`, `loop`, `clean`
- **인프라**: `hook`, `migration`, `worktree`, `session`
- **개발**: `mx`, `fix`, `brain`, `research`

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
- Registry가 28+ 타입 핸들러로 디스패치
- 각 핸들러는 `Handler` 인터페이스 준수

### 5. 멀티 LLM 실행 모드
- `moai cc`: Claude 전용
- `moai glm`: GLM 전용
- `moai cg`: Claude leader + GLM teammates
- GLM tier-models 테이블: Claude tier → GLM 모델 매핑

---

## 상세 참고 문서

이 개요는 빠른 이해를 제공합니다. 더 깊은 분석은 다음을 참고하세요:

- **modules.md**: 44개 runtime 패키지의 함수/타입/역할 상세 설명 + test-only 패키지 목록
- **dependencies.md**: Mermaid 패키지 의존도 그래프 + 팬-인/팬-아웃 정량화
- **entry-points.md**: ~50+ CLI 명령 + 훅 진입점 목록
- **data-flow.md**: 5가지 주요 플로우 시각화 (Mermaid)
- **docs-truth.md**: 문서 검증을 위한 수동 유지보수 사실 체크리스트

---

**생성**: `/moai codemaps --force`로 자동 생성  
**검증**: 순환 의존성 0개, 모든 패키지 경로 존재 확인
