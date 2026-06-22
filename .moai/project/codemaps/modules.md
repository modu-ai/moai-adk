# 패키지 모듈 상세 설명

> 이 문서는 `/moai codemaps --force`로 자동 생성된 패키지 목록입니다.

**모듈**: `github.com/modu-ai/moai-adk`  
**Go 버전**: go 1.26.4

---

## 프레젠테이션 계층

### cmd/moai
진입점: `main()` → `cli.Execute()`  
의존성: `internal/cli`

### internal/cli (241파일)
**역할**: Cobra 커맨드 트리, composition root  
**팬-아웃**: ~48개 internal 패키지  
**핵심**: `Execute()`, `InitDependencies()`, 50+ subcommand 라우팅

### internal/tui (32파일)
**역할**: Bubbletea TUI 요소, 28개 색상 토큰  
**기본**: Box, Pill, Table, Status, ProgressLine  
**의존성**: lipgloss

### internal/statusline (32파일)
**역할**: Claude Code 상태 렌더러, 3/5L 레이아웃  
**기능**: GitDataProvider, UpdateProvider, UsageProvider  
**의존성**: internal/core/git, internal/config

### internal/web (37파일)
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

### internal/spec (41파일)
**역할**: SPEC 라이프사이클 엔진  
**핵심**: Linter (13+3 규칙), ClassifyEra(), Audit(), DetectDrift(), ClassifyPRTitle()

### internal/constitution (23파일)
**역할**: 동결/진화 구역 모델, 5단계 병합 안전  
**기능**: FrozenGuard, Canary, ContradictionDetector, RateLimiter, HumanOversight

### internal/workflow
**역할**: Plan-Run-Sync 워크트리 오케스트레이션  
**기능**: `WorktreeOrchestrator`, `PhaseExecutor`, 품질 게이트

### internal/loop (18파일)
**역할**: 진단 피드백 루프 컨트롤러  
**핵심**: `LoopController`, `DecisionEngine`, `GoFeedbackGenerator`

### internal/ralph
**역할**: Ralph 의사결정 엔진  
**기능**: `Decide()` (max_iter > perfect_gate > stagnation > human_review)

### internal/harness (64파일)
**역할**: 하네스 학습 서브시스템  
**기능**: Observer, Learner (4-tier), Applier, 5단계 safety

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

### internal/design
**역할**: 디자인 시스템 도구  
**기능**: DTCG 토큰 검증, Path A/B1/B2 선택, BrandConflictAnalyzer

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

### internal/template (75파일)
**역할**: go:embed Template-First 시스템  
**소스**: internal/template/templates/ (단일 진실 공급원)  
**생성**: embedded.go (자동 생성, 편집 금지)  
**기능**: Deployer (원자적), Renderer (strict mode), Manifest.Track()

### internal/config (61파일)
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

### internal/migrate
**역할**: 마이그레이션 중 hook 정리  
**기능**: CleanupUserSettings, 아카이브 우선

### internal/update
**역할**: self-update  
**기능**: Checker, Updater, Rollback, 체크섬 gate, 원자적 replace

### internal/i18n
**역할**: 다국어 GitHub 코멘트  
**언어**: en, ko, ja, zh

### internal/hook (143파일)
**역할**: 컴파일된 훅 시스템  
**이벤트**: 28+ (SessionStart, PostToolUse, Stop, etc)  
**기능**: Registry.Dispatch(), exit 0/2, 순차 + short-circuit  
**서브**: trace, memo, quality, security, mx, handoff, lifecycle, dbsync

### internal/sandbox (19파일)
**역할**: OS 샌드박스 (seatbelt, bubblewrap, docker)  
**기능**: Launcher.Launch(), GenerateSBPL(), deny-by-default

### internal/shell
**역할**: shell 감지 및 config 변경  
**기능**: Configurator, AddEnvVar, AddPathEntry (멱등성)

### internal/astgrep (14파일)
**역할**: ast-grep CLI 래퍼  
**기능**: Scanner.Scan(), Finding 타입, SARIF

### internal/lsp (12 sub-packages)
**역할**: 다중언어 LSP 클라이언트  
**sub**: core, aggregator, gopls, cache, config, hook, subprocess, transport

### internal/mx (27파일)
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

### internal/session (25파일)
**역할**: 다중 세션 조율 레지스트리  
**기능**: Registry, FileSessionStore, PhaseState, advisory lock

### internal/state
**역할**: prompt-cache 사용량 원격측정  
**기능**: CacheUsageEntry JSONL, windowed 집계

### internal/tmux (12파일)
**역할**: tmux 감지, CG/GLM 모드  
**기능**: IsCGMode(), SessionManager

### internal/worktree
**역할**: 작업 트리 상태 가드  
**기능**: Capture(), Diff(), DivergenceLog

### internal/profile
**역할**: 사용자 프로필 관리  
**기능**: ProfilePreferences, GetCurrentName(), Sync

### internal/research
**역할**: 연구/계측 서브시스템  
**sub**: eval, experiment, safety, observe, dashboard

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
**패키지 수**: 45 internal 디렉터리 = 44 runtime 패키지 + 1 test-only + 2 pkg + 1 cmd = 48개 (runtime만 카탈로그에 포함)

---

**생성**: `/moai codemaps --force`로 자동 생성
