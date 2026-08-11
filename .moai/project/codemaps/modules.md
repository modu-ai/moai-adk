# 패키지 모듈 상세 설명

> 이 문서는 `/moai codemaps --force`로 자동 생성된 패키지 목록입니다.

**모듈**: `github.com/modu-ai/moai-adk`  
**Go 버전**: go 1.26.4

---

## 프레젠테이션 계층

### cmd/moai
진입점: `main()` → `cli.Execute()`  
의존성: `internal/cli`

### internal/cli (109 non-test 파일)
**역할**: Cobra 커맨드 트리, composition root  
**팬-아웃**: ~48개 internal 패키지  
**핵심**: `Execute()`, `InitDependencies()`, ~40 root verbs (152 non-test `.AddCommand()` 호출)

### internal/tui (19 non-test 파일)
**역할**: Bubbletea TUI 요소, 28개 색상 토큰  
**기본**: Box, Pill, Table, Status, ProgressLine  
**의존성**: lipgloss

### internal/statusline (15 non-test 파일)
**역할**: Claude Code 상태 렌더러, 3/5L 레이아웃  
**기능**: GitDataProvider, UpdateProvider, UsageProvider  
**의존성**: internal/core/git, internal/config

### internal/web (18 non-test 파일)
**역할**: loopback HTTP 콘솔, Templ + HTMX  
**기능**: host-header validation, graceful shutdown (5s 드레인)  
**의존성**: internal/profile, internal/config

### pkg/version
**역할**: 빌드타임 버전 (ldflags 주입)  
**팬-인 (High)**: 30+개 패키지

---

## 비즈니스/도메인 계층

### pkg/models
**역할**: 공유 config 타입 (매우 높은 팬-인)  
**타입**: ProjectType, DevelopmentMode, ProjectConfig, LanguageConfig, QualityConfig  
**팬-인 (Very High)**: 45+개 패키지

### internal/foundation
**역할**: 언어 레지스트리, TRUST 5, EARS  
**기능**: `LanguageRegistry`, 16개 언어 지원  
**팬-인 (High)**: 32+개 패키지

### internal/spec (24 non-test 파일)
**역할**: SPEC 라이프사이클 엔진  
**핵심**: Linter (13+3 규칙), ClassifyEra(), Audit(), DetectDrift(), ClassifyPRTitle()

### internal/constitution (13 non-test 파일)
**역할**: 동결/진화 구역 모델, 5단계 병합 안전  
**기능**: FrozenGuard, Canary, ContradictionDetector, RateLimiter, HumanOversight

### internal/workflow
**역할**: Plan-Run-Sync 워크트리 오케스트레이션  
**기능**: `WorktreeOrchestrator`, `PhaseExecutor`, 품질 게이트

### internal/loop (6 non-test 파일)
**역할**: 진단 피드백 루프 컨트롤러  
**핵심**: `LoopController`, `DecisionEngine`, `GoFeedbackGenerator`

### internal/ralph
**역할**: Ralph 의사결정 엔진  
**기능**: `Decide()` (max_iter > perfect_gate > stagnation > human_review)

### internal/harness (80 non-test 파일)
**역할**: 하네스 학습 서브시스템  
**기능**: Observer, Learner (4-tier), Applier, 5단계 safety  
**하위 패키지**: `routing` — 위임 관측 원장(routing ledger), 훅 입력에서 기계적으로 기록 · `delegationmap` — 관측된 원장 행을 집계해 위임 맵 개정안을 제안(읽기 전용, 적용은 Tier-4 승인 게이트)

### internal/permission (18파일)
**역할**: 8-tier 권한 스택  
**모드**: default, acceptEdits, bypassPermissions, plan, bubble  
**계층**: policy → project → user → team → builtin → systemDefault → hookOverride → deny

### internal/evolution
**역할**: 반사 학습 Write Phase  
**기능**: LearningEntry (LEARN-YYYYMMDD-NNN), 졸업 신뢰도 (3→5→10)

### internal/merge
**역할**: 3-way 파일 병합 (ADR-008)  
**전략**: LineMerge, YAMLDeep, JSONMerge, SectionMerge, EvolvableZoneMerge, Overwrite

> **`internal/design`** — v3.0 코드베이스에 독립 패키지로 존재하지 않음 (이전 문서 드리프트). design 관련 로직은 `internal/harness` 등에 분산.

### internal/bodp
**역할**: Branch Origin Decision Protocol  
**기능**: 3-signal 검사 → 8-row 의사결정 매트릭스 → main/stacked/continue

### internal/git
**역할**: label → branch-prefix 컨벤션  
**기능**: `DetectBranchPrefix()`, `FormatIssueBranch()`

---

## 인프라 계층

### internal/core/git
**역할**: exec 기반 Git 추상화  
**인터페이스**: Repository, BranchManager, WorktreeManager  
**팬-인 (High)**: 35+개 패키지

### internal/core/project
**역할**: 프로젝트 루트 발견  
**ANCHOR**: `FindProjectRoot()` — `.moai/` 발견 (everywhere)

### internal/core/quality
**역할**: TRUST 5 gate enforcement  
**기능**: phase-aware 임계값, DDD/TDD 변형

### internal/runtime
**역할**: 토큰 circuit-breaker, 예산 추적  
**기능**: soft 75% / hard 90%, stall 감지, progress.md auto-save

### internal/template
**역할**: go:embed Template-First 시스템
**소스**: internal/template/templates/ (단일 진실 공급원)
**임베드**: `embed.go`가 직접 `//go:embed all:templates` 사용 (별도 `embedded.go` 자동 생성 없음)
**기능**: Deployer (원자적), Renderer (strict mode), Manifest.Track(), profile_matrix (11 agents × 3 profiles = 33 cells)

### internal/config (35 non-test 파일)
**역할**: 계층화 YAML config SSOT  
**우선순위**: env > yaml > defaults  
**팬-인 (Very High)**: 48+개 패키지

### internal/manifest
**역할**: 파일 출처 3-way 추적  
**기능**: 3중 해시 (template/deployed/current), 손상 복구

### internal/defs
**역할**: 디렉토리 레이아웃 상수  
**기능**: `.moai/`, `.claude/` 구조, DeprecatedPaths

### internal/migration
**역할**: 버전 기반 마이그레이션 실행기  
**기능**: Apply, Status, Rollback, 멱등성

> **`internal/migrate`** — v3.0에 없음. 마이그레이션은 `internal/migration` (단수형) 사용.

### internal/update
**역할**: self-update  
**기능**: Checker, Updater, Rollback, 체크섬 gate, 원자적 replace

### internal/goal
**역할**: 목표 엔진 — 조건 선언형 에이전틱 루프 (`/moai goal`)
**핵심**: `moai goal arm|status|clear`, Condition {Mechanical,Model}, Stop-hook 평가 계약
**상태**: `.moai/state/goal/<session-id>.json` (세션별)

### internal/factory
**역할**: Factory 모드 상태 — `moai cc -f` / `moai glm -f`가 여는 plan→run→verify→sync 체인의 세션 기록과 중복 억제
**핵심**: `record.go` (세션 레코드 기록, `validateSessionID`가 경로 조작 차단, 파일 0600), `revision.go` (`Matches`/`RevisionMatch`/`SuppressStep0551` — 모든 실패 모드가 "검사 수행"으로 수렴하는 fail-safe, rung은 allow-list)
**상태**: 세션 ID 파생 경로의 레코드 파일 + `revision.json`
**진입점**: `internal/cli/factory.go` (플래그 파싱, env 진입/복원), `internal/cli/launcher_blockcap_infinite.go` (Stop-hook block cap 상향)

### internal/hook
**역할**: 컴파일된 훅 시스템 + main-checkout branch-state guard
**이벤트**: 30개 EventType (SessionStart, PostToolUse, Stop, etc), 35개 `handle-*.sh` 래퍼
**기능**: Registry.Dispatch(), Stop은 stdout JSON `decision:"block"` (exit 0), 순차 + short-circuit
**서브**: trace, memo, quality, security, mx, handoff, lifecycle, dbsync, branch_guard

### internal/sandbox (19파일)
**역할**: OS 샌드박스 (seatbelt, bubblewrap, docker)  
**기능**: Launcher.Launch(), GenerateSBPL(), deny-by-default

### internal/shell
**역할**: shell 감지 및 config 변경  
**기능**: Configurator, AddEnvVar, AddPathEntry (멱등성)

### internal/astgrep (5 non-test 파일)
**역할**: ast-grep CLI 래퍼  
**기능**: Scanner.Scan(), Finding 타입, SARIF

### internal/lsp (8 sub-packages: aggregator, cache, config, core, gopls, hook, subprocess, transport)
**역할**: 다중언어 LSP 클라이언트  
**sub**: core, aggregator, gopls, cache, config, hook, subprocess, transport

### internal/mx (12 non-test 파일)
**역할**: @MX 태그 스캐너/리졸버  
**기능**: Scanner, Resolver, FanInCounter, Sidecar JSON

### internal/ciwatch
**역할**: CI 체크 분류  
**기능**: Classifier.IsRequired(), Handoff/WatchState

### internal/resilience
**역할**: circuit breaker FSM  
**상태**: closed, open, half-open

### internal/telemetry
**역할**: 비동기 사용량 기록  
**기능**: AsyncRecorder, bounded channel, 배치 disk I/O

### internal/github (26파일)
**역할**: gh CLI 통합  
**기능**: GHClient 인터페이스, SpecLinker, SecretManager

### internal/session (12 non-test 파일)
**역할**: 다중 세션 조율 레지스트리  
**기능**: Registry, FileSessionStore, PhaseState, advisory lock

> **`internal/state`** — v3.0에 독립 패키지 없음. 세션/상태 관리는 `internal/session` (registry, checkpoint, phase).

### internal/tmux (4 non-test 파일)
**역할**: tmux 감지, CG/GLM 모드  
**기능**: IsCGMode(), SessionManager

### internal/worktree
**역할**: 작업 트리 상태 가드  
**기능**: Capture(), Diff(), DivergenceLog

### internal/profile
**역할**: 사용자 프로필 관리  
**기능**: ProfilePreferences, GetCurrentName(), Sync

> **`internal/research`** — v3.0에 독립 패키지 없음 (이전 문서 드리프트).

### internal/measure
**역할**: zero-dependency 리프 파서  
**기능**: ParseGoTestJSON(), ParseCoverageFile(), CountNonEmptyLines()

---

## 테스트 전용 패키지 (런타임 카탈로그 제외)

다음 패키지는 테스트 전용이므로 runtime 모듈 카탈로그에서 제외됩니다:

- **internal/skills** — audit-only test fixture (LOC-ceiling / template-mirror-parity test suite, 프로덕션 코드 없음)

> `internal/evaluator`는 방치된 TDD RED 스캐폴드(SPEC-EVAL-001, sync-auditor 에이전트로 대체)로 SPEC-CLEANUP-EVALUATOR-001에서 제거되었습니다.

---

## 검증

**순환 의존성**: 0개 (검증됨)  
**패키지 수**: 49 internal 디렉터리 (323 중첩 디렉터리) + 2 pkg (`models`, `version`) + 1 cmd = 52개 경로

> 실측 명령: `ls -d internal/*/ | wc -l` → 49, `find internal -type d -mindepth 2 | wc -l` → 323. 종전 표기(46 / 318)는 갱신이 밀린 값이며, 세 패키지 차이 중 `internal/factory` 하나만 SPEC-FACTORY-MODE-001이 추가한 것이고 나머지 둘은 그 이전부터 있었다. `structure.md` 71행의 같은 수치도 함께 갱신했다.

---

**생성**: `/moai codemaps --force`로 자동 생성
