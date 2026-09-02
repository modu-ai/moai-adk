# 패키지 모듈 상세 설명

> 이 문서는 `/moai codemaps --force`로 자동 생성된 패키지 목록입니다.

**모듈**: `github.com/modu-ai/moai-adk`  
**Go 버전**: go 1.26.4

---

## 프레젠테이션 계층

### cmd/moai
진입점: `main()` → `cli.Execute()`  
의존성: `internal/cli`

### internal/cli (264 non-test 파일, 하위 패키지 포함 — 2026-09-02 실측; 최상위 202 + 하위 62)
**역할**: Cobra 커맨드 트리, composition root  
**팬-아웃**: 내부 패키지 93개 임포트  
**핵심**: `Execute()`, `InitDependencies()`, root 등록 60건 (202 non-test `.AddCommand()` 호출, 33개 파일 분산)  
**codex 런처·게이트**: `codex_launcher.go` — `moai codex` 동사 라우팅(폐쇄 집합 — 기동 {bare, cli, app} · 리드아웃 {status})과 그 하류의 argv 번역(실재 서브커맨드 `app` 만 자식에 전달), 6행 리드아웃, 직접/tmux spawn 기동 · `codex_init.go` — init-offer 게이트(세 기동 형태의 단일 통과점, 런처 배선 판정 소비, 거절 rc 130 / 비대화형 rc 1, 수락 시 생성기 위임) · `codex_contract.go` — AGENTS.md ↔ CLAUDE.md 지시 계약(연결 전용, 경로 봉쇄 선행, per-file temp+rename, 멱등)  
**하위 패키지**: `agentlint`, `harness`, `pr`, `preference`, `printer`, `specid`, `taskledger`, `uikit`, `update/{backup,deploy,merge,plan,report}`, `wizard`, `worktree`

### internal/tui (19 non-test 파일)
**역할**: Bubbletea TUI 요소, 28개 색상 토큰  
**기본**: Box, Pill, Table, Status, ProgressLine  
**하위**: `golden` (골든 스냅샷), `internal` (내부 전용 유틸)  
**의존성**: lipgloss

### internal/statusline (19 non-test 파일)
**역할**: Claude Code 상태 렌더러, 3/5L 레이아웃  
**기능**: GitDataProvider, UpdateProvider, UsageProvider, GLM 컨텍스트 윈도우 해석(`memory.go` `glmContextWindows` — glm-5.3 → 1M)  
**의존성**: internal/core/git, internal/config

### internal/web (29 non-test 파일)
**역할**: loopback HTTP 콘솔, Templ + HTMX  
**기능**: host-header validation, graceful shutdown (5s 드레인)  
**의존성**: internal/profile, internal/config

### pkg/version
**역할**: 빌드타임 버전 (ldflags 주입)

---

## 비즈니스/도메인 계층

### pkg/models
**역할**: 공유 config 타입  
**타입**: ProjectType, DevelopmentMode, ProjectConfig, LanguageConfig, QualityConfig

### internal/foundation
**역할**: 언어 레지스트리, TRUST 5, EARS  
**기능**: `LanguageRegistry`, 16개 언어 지원

### internal/spec (31 non-test 파일)
**역할**: SPEC 라이프사이클 엔진  
**핵심**: Linter (19 규칙: 15 단일-문서 + 3 크로스-SPEC + 1 registry — `StatusTransitionValidityRule`·`ArtifactStatusFieldForbiddenRule`·`MovingRefUnpinnedRule`·`SyncSHASlotFormatRule` 포함), ClassifyEra(), Audit(), DetectDrift(), ClassifyPRTitle()

### internal/constitution (14 non-test 파일)
**역할**: 동결/진화 구역 모델, 5단계 병합 안전  
**기능**: FrozenGuard, Canary, ContradictionDetector, RateLimiter, HumanOversight

### internal/workflow (4 non-test 파일)
**역할**: Plan-Run-Sync 워크트리 오케스트레이션  
**기능**: `WorktreeOrchestrator`, `PhaseExecutor`, 품질 게이트

### internal/loop (6 non-test 파일)
**역할**: 진단 피드백 루프 컨트롤러  
**핵심**: `LoopController`, `DecisionEngine`, `GoFeedbackGenerator`

### internal/ralph
**역할**: Ralph 의사결정 엔진  
**기능**: `Decide()` (max_iter > perfect_gate > stagnation > human_review)

### internal/harness (82 non-test 파일)
**역할**: 하네스 학습 서브시스템  
**기능**: Observer, Learner (4-tier), Applier, 5단계 safety  
**하위 패키지**: `capture`, `cluster`, `curator`, `delegationmap` (관측 원장 행 집계 → 위임 맵 개정안 제안, 읽기 전용), `harnessrun`, `proposalgen`, `router`, `routing` (위임 관측 원장, 훅 입력에서 기계적으로 기록), `safety`, `seeds`, `throttle`, `tier`, `v4manifest`

### internal/navigator (53 non-test 파일)
**역할**: BAS(Blueprint-Anchored Synchronization) 코드 탐색 계층  
**하위 패키지**: `astx` (그래프 빌더의 AST 추출 시임), `detect`, `fix`, `route`, `sync` (3개 Navigator 체인 통합 — regen/audit/sync), `tiers` (4-tier 주소 맵 오버레이, tiers.json OVERLAY 산출)

### internal/kanban (33 non-test 파일)
**역할**: 칸반 백로그 큐 엔진 — `moai todo` 명령 백엔드  
**핵심**: `backlog_analysis.go` (`moai todo add`/`analyze`의 기계적 텍스트 분석기), 큐 저장소, 카드 관계 분석

### internal/epic (5 non-test 파일)
**역할**: 디스크 기반 에픽 진행 생산자 — `moai epic status <prefix>`

### internal/permission (6 non-test 파일)
**역할**: 8-tier 권한 스택  
**모드**: default, acceptEdits, bypassPermissions, plan, bubble  
**계층**: policy → project → user → team → builtin → systemDefault → hookOverride → deny

### internal/evolution
**역할**: 반사 학습 Write Phase  
**기능**: LearningEntry (LEARN-YYYYMMDD-NNN), 졸업 신뢰도 (3→5→10)

### internal/merge (6 non-test 파일)
**역할**: 3-way 파일 병합 (ADR-008)  
**전략**: LineMerge, YAMLDeep, JSONMerge, SectionMerge, EvolvableZoneMerge, Overwrite

> **`internal/design`** — v3.0 코드베이스에 독립 패키지로 존재하지 않음 (이전 문서 드리프트). design 관련 로직은 `internal/harness` 등에 분산.

### internal/git (8 non-test 파일)
**역할**: label → branch-prefix 컨벤션  
**기능**: `DetectBranchPrefix()`, `FormatIssueBranch()`  
**하위**: `convention`

### internal/feedback (7 non-test 파일)
**역할**: 피드백 제출 실패 리트라이 큐 (`queue.go`)

---

## 인프라 계층

### internal/core/git
**역할**: exec 기반 Git 추상화  
**인터페이스**: Repository, BranchManager, WorktreeManager

### internal/core/project
**역할**: 프로젝트 루트 발견  
**ANCHOR**: `FindProjectRoot()` — `.moai/` 발견 (everywhere)

### internal/core/quality
**역할**: TRUST 5 gate enforcement  
**기능**: phase-aware 임계값, DDD/TDD 변형

### internal/runtime (10 non-test 파일)
**역할**: 토큰 circuit-breaker, 예산 추적  
**기능**: soft 75% / hard 90%, stall 감지, progress.md auto-save  
**하위**: `gobin`

### internal/template (23 non-test 파일)
**역할**: go:embed Template-First 시스템  
**소스**: internal/template/templates/ (단일 진실 공급원)  
**임베드**: `embed.go`가 직접 `//go:embed all:templates` 사용 (별도 `embedded.go` 자동 생성 없음)  
**기능**: Deployer (원자적), Renderer (strict mode), Manifest.Track(), profile_matrix (11 agents × 3 profiles = 33 cells)  
**하위**: `agentemit` (에이전트 출력 에밋), `scripts`

### internal/config (48 non-test 파일, 하위 포함)
**역할**: 계층화 YAML config SSOT  
**우선순위**: env > yaml > defaults  
**하위**: `atomicfile`, `toolpolicy`

### internal/manifest
**역할**: 파일 출처 3-way 추적  
**기능**: 3중 해시 (template/deployed/current), 손상 복구

### internal/defs (5 non-test 파일)
**역할**: 디렉토리 레이아웃 상수  
**기능**: `.moai/`, `.claude/` 구조, DeprecatedPaths

### internal/paths (11 non-test 파일)
**역할**: `~/.moai` 디렉터리 트리의 단일 해석 지점

### internal/migration (8 non-test 파일)
**역할**: 버전 기반 마이그레이션 실행기  
**기능**: Apply, Status, Rollback, 멱등성  
**하위**: `migrations`  
**CLI**: `moai migration run|status|rollback` (migration.go — `run` 동사는 이 명령의 서브커맨드다)

> **`internal/migrate`** — v3.0에 없음. 마이그레이션은 `internal/migration` (단수형) 사용.

### internal/update (7 non-test 파일)
**역할**: self-update
**기능**: Checker, Updater, Rollback, 체크섬 gate, 원자적 replace

### internal/factory
**역할**: Factory 모드 상태 — `moai cc -f` / `moai glm -f`가 여는 plan→run→verify→sync 체인의 세션 기록과 중복 억제
**핵심**: `record.go` (세션 레코드 기록, `validateSessionID`가 경로 조작 차단, 파일 0600), `revision.go` (`Matches`/`RevisionMatch`/`SuppressStep0551` — 모든 실패 모드가 "검사 수행"으로 수렴하는 fail-safe, rung은 allow-list)
**상태**: 세션 ID 파생 경로의 레코드 파일 + `revision.json`
**진입점**: `internal/cli/factory.go` (플래그 파싱, env 진입/복원), `internal/cli/launcher_blockcap_infinite.go` (Stop-hook block cap 상향)

### internal/goal (6 non-test 파일)
**역할**: 목표 엔진 — 조건 선언형 에이전틱 루프 (`/moai goal`)  
**핵심**: `moai goal arm|status|clear`, Condition {Mechanical,Model}, Stop-hook 평가 계약  
**상태**: `.moai/state/goal/<session-id>.json` (세션별)

### internal/execerr (8 non-test 파일)
**역할**: 서브프로세스 exit 실패를 출력 가능하게 유지 — `*exec.ExitError` 노출 없이 %w 체인에서 인쇄 가능한 실패로 보존

### internal/hook (128 non-test 파일, 하위 포함)
**역할**: 컴파일된 훅 시스템 + main-checkout branch-state guard  
**이벤트**: 30개 EventType (SessionStart, PostToolUse, Stop, etc)  
**기능**: Registry.Dispatch(), Stop은 stdout JSON `decision:"block"` (exit 0), 순차 + short-circuit  
**하위**: `handoff`, `memo`, `memo/taxonomy`, `mx`, `mx/complexity`, `perf`, `quality`, `security`, `testutil`, `trace`

### internal/sandbox (8 non-test 파일)
**역할**: OS 샌드박스 (seatbelt, bubblewrap, docker)  
**기능**: Launcher.Launch(), GenerateSBPL(), deny-by-default

### internal/shell (6 non-test 파일)
**역할**: shell 감지 및 config 변경  
**기능**: Configurator, AddEnvVar, AddPathEntry (멱등성)

### internal/astgrep (13 non-test 파일)
**역할**: ast-grep CLI 래퍼  
**기능**: Scanner.Scan(), Finding 타입, SARIF

### internal/lsp (35 non-test 파일, 8 sub-packages)
**역할**: 다중언어 LSP 클라이언트  
**sub**: aggregator, cache, config, core, gopls, hook, subprocess, transport

### internal/mx (16 non-test 파일)
**역할**: @MX 태그 스캐너/리졸버  
**기능**: Scanner, Resolver, FanInCounter, Sidecar JSON

### internal/graph (14 non-test 파일)
**역할**: 코드베이스 엣지 산출물(edges.jsonl) + 3-레이어 신선도 게이트(`moai graph check`)·쿼리 시점 갱신·인용 앵커·MCP 코드 쿼리 엔진(graph_shortest_path 포함)  
**하위**: `symbol` (astx 추출 시임 — navigator 계층 의존 없이 code-call/code-import 엣지 추출)

### internal/chain
**역할**: 훅 이벤트 체인 코어 (worktree 세션 origin-trail)

### internal/mcp
**역할**: 자체 호스팅 MCP 서버의 도구 카탈로그 단일 선언

### internal/ciwatch
**역할**: CI 체크 분류  
**기능**: Classifier.IsRequired(), Handoff/WatchState

### internal/codexwiring (4 non-test 파일)
**역할**: `moai init --agent codex` 배선 생성기  
**기능**: `Wire()` — `.codex/hooks.json`·`config.toml` 생성, `InspectMCPTable()` — config.toml MCP 테이블 검사. 배선 상태 분류(`classifyCodexWiring`)는 이 패키지가 아니라 `internal/cli/codex_readiness.go` 소속이며, 런처 리드아웃과 init-offer 게이트가 그 반환값을 소비한다

### internal/codexadapter (5 non-test 파일)
**역할**: codex 하네스 어댑터 공통 계층  
**기능**: config/diagnostics/events/output/stderr — codexwiring이 소비

### internal/resilience
**역할**: circuit breaker FSM  
**상태**: closed, open, half-open

### internal/telemetry (5 non-test 파일)
**역할**: 비동기 사용량 기록  
**기능**: AsyncRecorder, bounded channel, 배치 disk I/O

### internal/github (10 non-test 파일)
**역할**: gh CLI 통합  
**기능**: GHClient 인터페이스, SpecLinker, SecretManager  
**하위**: `workflow`

### internal/session (22 non-test 파일)
**역할**: 다중 세션 조율 레지스트리  
**기능**: Registry, FileSessionStore, PhaseState, advisory lock

> **`internal/state`** — v3.0에 독립 패키지 없음. 세션/상태 관리는 `internal/session` (registry, checkpoint, phase).

### internal/sessionmsg (7 non-test 파일)
**역할**: 단일 머신 세션 메시징 브로커 코어 — Claude↔Codex 양방향 메시징 (`session_msg_*` MCP 도구의 백엔드, SPEC-CODEX-SESSION-MSG-001)

### internal/glmcred
**역할**: GLM API 자격증명(`~/.moai/.env.glm`) 쓰기·판독의 단일 구현

### internal/settings (10 non-test 파일)
**역할**: settings.json / settings.local.json 헬퍼  
**하위**: `agentfm` (에이전트 frontmatter), `yamlpatch`

### internal/tmux (4 non-test 파일)
**역할**: tmux 감지, CG/GLM 모드  
**기능**: IsCGMode(), SessionManager

### internal/worktree (4 non-test 파일)
**역할**: 작업 트리 상태 가드  
**기능**: Capture(), Diff(), DivergenceLog

### internal/profile (3 non-test 파일)
**역할**: 사용자 프로필 관리  
**기능**: ProfilePreferences, GetCurrentName(), Sync

> **`internal/research`** — v3.0에 독립 패키지 없음 (이전 문서 드리프트).

### internal/measure
**역할**: zero-dependency 리프 파서  
**기능**: ParseGoTestJSON(), ParseCoverageFile(), CountNonEmptyLines()

### internal/guardstate (4 non-test 파일)
**역할**: 가드 발화 이벤트를 8행 상태표(spec.md §C.2)로 분류하는 상태 모델 (SPEC-GUARD-STATE-MODEL-001)  
**기능**: `Classification`(닫힌 7값 어휘, `ClassOK`가 유일한 clean 값), `Classify()`/`Evaluate()`/`Produce()`(판정→집계→산출물 조립), `KindGitHubWorkflow`(현재 리더를 갖춘 유일한 subject kind)

### internal/guardliveness (4 non-test 파일)
**역할**: guardstate 판정 결과를 운영자가 묻지 않아도 드러내는 표면화 계층 (SPEC-GUARD-LIVENESS-001, card t333)  
**기능**: 3-clause 계약만 소비(entry당 정확히 1개 분류 / clean 값 1개 / 그 값의 기계 판독 가능 표시) — 분류 어휘 자체는 소유하지 않음. `Entry`, `Store`(평가 결과의 나중 활성화 시점 렌더용 왕복 저장)

### internal/binlag
**역할**: 설치된 `moai` 바이너리가 현재 소스 트리보다 뒤처졌는지(commit lag) 판정하는 단일 비교 지점 (SPEC-BINARY-LAG-VISIBILITY-001)  
**기능**: `Status`(`not-applicable`/`fresh`/`divergent`/그 외), `moai doctor` 항목과 SessionStart 어드바이저 양쪽이 같은 `Comparer` 구현을 공유 — `internal/cli`가 `internal/hook`을 임포트하는 방향성 때문에 `internal/hook`보다 아래 계층에 위치

### internal/atomicfile (5 non-test 파일)
**역할**: 원자적 파일 쓰기 (write-temp + rename) — `merge`/`manifest`/`config`가 사용

### internal/lockfile
**역할**: 크로스 플랫폼 잠금 (Unix `flock(2)` / Windows in-process mutex) — `session`이 사용

### internal/tokenusage
**역할**: 토큰 사용량 계수 (statusline 연동) — `moai tokens` 명령

### internal/verify (6 non-test 파일)
**역할**: 검증 서브시스템 — 공유 진단 스냅샷 기록·판독 (`moai verify check`)

### internal/timing
**역할**: 테스트용 보정 지연 상한 — 코드가 측정되는 것과 실행 머신을 구분하는 단언 제공

### internal/report/planhtml
**역할**: plan-phase HTML 리포트 렌더러 — plan-auditor 리뷰 마크다운 파싱 후 HTML 산출

### internal/mirrornotice
**역할**: 템플릿 배포의 스킬 미러 결과를 사용자 공지 행으로 변환

---

## 테스트 전용 패키지 (런타임 카탈로그 제외)

다음 패키지는 테스트 전용이므로 runtime 모듈 카탈로그에서 제외됩니다:

- **internal/skills** — audit-only test fixture (`workflow_split_test.go` 1파일, 프로덕션 코드 없음. `go list ./...` 카탈로그에는 테스트 파일 보유로 등장)

> `internal/evaluator`는 방치된 TDD RED 스캐폴드(SPEC-EVAL-001, sync-auditor 에이전트로 대체)로 SPEC-CLEANUP-EVALUATOR-001에서 제거되었습니다.

---

## 검증

**순환 의존성**: 0개 (검증됨 — `go build ./...` 성공이 순환 부재의 기계적 증거)  
**패키지 수**: 64 internal 디렉터리 (`go list ./...` 기준 internal 패키지 131개) + 2 pkg (`models`, `version`) + 1 cmd = 137개 카탈로그 항목 (scripts 3개 포함)

> 실측 명령(2026-09-02): `ls -d internal/*/ | wc -l` → 64, `go list ./... | grep -c '/internal/'` → 131, `go list ./... | wc -l` → 137. 종전 표기(390 중첩 디렉터리)는 미갱신 값이며, 스탬프 커밋 이후 340커밋 동안 하위 패키지 구조가 다수 변동되었다 (navigator 6 서브패키지, harness 13 서브패키지, cli 11 하위, hook 10 하위 등). `structure.md` 같은 외부 문서의 수치는 이 SPEC의 설명 대상 루트 밖이므로 이번 배치에서 손대지 않았다.

---

**생성**: `/moai codemaps --force`로 자동 생성
