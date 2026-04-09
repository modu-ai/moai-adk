# DESIGN-SRS-002: moai-adk 통합 자가 연구 시스템

> **Status**: Draft
> **Author**: MoAI Orchestrator
> **Date**: 2026-04-09
> **Supersedes**: DESIGN-SRS-001 (Agency 별도 유지 방식)

---

## Part 1: moai-adk-go 코드맵

### 1.1 패키지 구조 (42개 패키지, ~42K 소스 라인)

```
cmd/moai/main.go                          # 바이너리 진입점 (~15 lines)
│
internal/cli/                              # 조합 루트 + CLI 명령 (~7,087 lines)
├── internal/cli/wizard/                   # 대화형 프로젝트 초기화 위저드 (~1,007)
└── internal/cli/worktree/                 # moai worktree 하위 명령 (~1,344)
│
├── internal/config/                       # YAML 설정 관리자 (~1,514)
├── internal/defs/                         # 전역 상수 (디렉토리, 파일명, 권한) (~120)
├── internal/foundation/                   # 에러 타입, 언어 감지, 타임아웃 (~468)
│
├── internal/core/                         # [DEAD] 플랫폼 경로 유틸 (~86) ← 삭제 대상
├── internal/core/git/                     # Git 작업 추상화 (~1,288)
├── internal/core/project/                 # 프로젝트 감지/초기화/검증 (~2,053)
├── internal/core/quality/                 # TRUST 5 품질 프레임워크 (~1,348)
│
├── internal/hook/                         # Claude Code 훅 프로토콜 (~4,908)
│   ├── internal/hook/agents/              # [STUB] 에이전트별 훅 핸들러 (~541)
│   ├── internal/hook/lifecycle/           # 세션 생명주기 관리 (~562)
│   ├── internal/hook/memo/               # 세션 메모 시스템 (~189)
│   ├── internal/hook/mx/                 # @MX 태그 검증 (~802)
│   ├── internal/hook/quality/            # 품질 게이트 (린터, 포매터) (~1,197)
│   ├── internal/hook/security/           # AST 보안 스캐닝 (~1,098)
│   └── internal/hook/trace/              # 관찰/추적 (JSONL) (~395)
│
├── internal/lsp/                          # LSP 데이터 모델 (~203)
│   └── internal/lsp/hook/                # LSP 진단 수집 (~1,518)
│
├── internal/astgrep/                      # AST 패턴 매칭 (sg CLI) (~740)
├── internal/resilience/                   # 서킷 브레이커 패턴 (~260)
├── internal/manifest/                     # 파일 추적 (provenance) (~360)
├── internal/template/                     # 임베디드 템플릿 시스템 (~1,289)
├── internal/merge/                        # 3-way 병합 엔진 (~1,788)
├── internal/statusline/                   # 터미널 상태 라인 (~2,612)
├── internal/update/                       # 바이너리 자동 업데이트 (~1,185)
├── internal/loop/                         # Ralph 피드백 루프 상태 머신 (~717)
├── internal/ralph/                        # Ralph 결정 엔진 (~90)
├── internal/workflow/                     # SPEC 워크플로우 오케스트레이터 (~444)
├── internal/github/                       # GitHub CLI 통합 (~1,471)
├── internal/git/                          # 경량 git 브랜치 감지 (~38)
├── internal/git/convention/               # 커밋 메시지 컨벤션 (~584)
├── internal/i18n/                         # 다국어 템플릿 (~162)
├── internal/shell/                        # 셸 환경 감지/설정 (~776)
├── internal/tmux/                         # tmux 세션 관리 (~333)
├── internal/profile/                      # 사용자 프로필 관리 (~470)
├── internal/mcp/                          # MCP 서버 구현 (~443)
│
├── pkg/version/                           # 빌드 버전 정보 (~63)
└── pkg/models/                            # 공유 데이터 타입 (~347)
```

### 1.2 의존성 그래프

```
cmd/moai
  └── internal/cli  ← 모든 패키지를 주입하는 조합 루트
        │
        ├── [Infrastructure]
        │   ├── config ← defs, pkg/models
        │   ├── manifest ← defs
        │   ├── template ← manifest, config, pkg/models
        │   ├── merge (독립)
        │   └── resilience (독립)
        │
        ├── [Core Domain]
        │   ├── core/git ← foundation
        │   ├── core/project ← foundation, manifest, template, pkg/models
        │   ├── core/quality ← config, core/git
        │   └── workflow ← core/git, core/quality
        │
        ├── [Hook System]
        │   ├── hook ← config, hook/trace, lsp/hook, astgrep
        │   ├── hook/agents ← hook  [DEAD - never wired]
        │   ├── hook/lifecycle ← defs
        │   ├── hook/memo ← defs
        │   ├── hook/mx ← defs
        │   ├── hook/quality ← lsp/hook, astgrep
        │   ├── hook/security ← astgrep
        │   └── hook/trace (독립)
        │
        ├── [Feedback Loop]
        │   ├── loop (독립)
        │   └── ralph ← config, loop
        │
        ├── [External Integration]
        │   ├── github ← git, i18n
        │   ├── statusline ← defs, core/git, pkg/version
        │   ├── shell (독립)
        │   ├── tmux (독립)
        │   └── mcp ← lsp
        │
        └── [Leaf Packages]
            ├── pkg/version (독립)
            ├── pkg/models (독립)
            ├── defs (독립)
            ├── foundation (독립)
            ├── astgrep (독립)
            ├── i18n (독립)
            ├── git (독립)
            └── git/convention ← pkg/models
```

### 1.3 템플릿/설정 인벤토리

| 영역 | 파일 수 | 동기화 상태 |
|------|---------|------------|
| 에이전트 (`.claude/agents/`) | 26 (moai:20, agency:6) | SYNC |
| 스킬 (`.claude/skills/`) | 62 SKILL.md (404 전체) | SYNC |
| 룰 (`.claude/rules/`) | 32 | SYNC |
| 커맨드 (`.claude/commands/`) | 23 | DRIFT: 로컬 2개 추가 (`98-github`, `99-release`) |
| 훅 (`.claude/hooks/`) | 26 | DRIFT: 이름 불일치 1개 |
| Agency (`.agency/`) | 8 | SYNC |
| 설정 (`.moai/config/`) | 23 | DRIFT: 템플릿에 3개 추가 미반영 |
| 출력 스타일 | 3 | SYNC |

---

## Part 2: 데드코드 및 개선점

### 2.1 삭제 대상 (Dead Code)

#### P1: 죽은 패키지

| 패키지 | 파일 | 이유 |
|--------|------|------|
| `internal/core/` (최상위) | `temp.go`, `pathutil_*.go` | 전체 프로젝트에서 import 0회 |

#### P2: 미사용 export (삭제 가능)

| 파일 | 미사용 식별자 |
|------|-------------|
| `internal/core/project/validator.go` | `BackupTimestampFormat`, `BackupsDir` (deprecated 주석 포함) |
| `internal/foundation/timeouts.go` | `DefaultCLITimeout`, `DefaultSearchTimeout`, `DefaultLSPTimeout` |
| `internal/foundation/errors.go` | `ErrInvalidRequirementType`, `ErrInvalidPillar` 외 5개 에러 타입 |
| `internal/shell/errors.go` | `ErrUnsupportedShell`, `ErrConfigNotFound` |
| `internal/defs/timeouts.go` | `HookDefaultTimeout`, `HookPostToolTimeout`, `GitShortTimeout`, `GitLongTimeout` |
| `internal/defs/perms.go` | `CredDirPerm`, `CredFilePerm` |
| `internal/defs/files.go` | `GithubSpecRegistryJSON`, `MCPJSON` |
| `internal/defs/paths.go` | `StatusLinePath` |
| `internal/defs/dirs.go` | `SpecsSubdir`, `ReportsSubdir` |
| `internal/core/project/root.go` | `MustFindProjectRoot()` (테스트에서만 사용) |

#### P3: TODO 스텁 (구현 필요 또는 삭제)

| 파일 | 내용 |
|------|------|
| `internal/hook/agents/*.go` (10개) | 모든 핸들러가 `NewAllowOutput()` 반환 - 미구현 |
| `internal/mcp/lsp_stub.go` (6개 메서드) | "Phase 2 will replace" - LSP 브릿지 미연결 |
| `internal/cli/worktree/new.go:189` | `isTmuxPreferred()` 항상 false 반환 |

### 2.2 수정 대상 (Bugs & Inconsistencies)

#### SPEC ID 정규식 불일치 (잠재적 버그)

```
internal/hook/user_prompt_submit.go:21  → SPEC-[A-Z0-9]+-\d+  (숫자 허용)
internal/hook/task_completed.go:15      → SPEC-[A-Z]+-\d+      (숫자 미허용)
internal/workflow/specid.go:20          → ^SPEC-ISSUE-\d+$      (ISSUE 고정)
```

`[A-Z0-9]+` vs `[A-Z]+` 차이는 `SPEC-CC297-001` 같은 ID에서 다른 결과를 낸다. `internal/workflow/specid.go`에 통합 필요.

#### 타임아웃 상수 중복

```
internal/defs/timeouts.go:17     → GitShortTimeout = 5s  (미사용)
internal/foundation/timeouts.go  → DefaultGitTimeout = 5s (사용됨)
```

`defs/timeouts.go` 전체 삭제 가능 (사용되는 값은 모두 foundation에 있음).

#### 템플릿 동기화 불일치

| 문제 | 위치 | 조치 |
|------|------|------|
| 훅 이름 불일치 | `handle-permission-request.sh` vs `handle-permission-denied.sh` | 통일 |
| 설정 누락 | `gate.yaml`, `memo.yaml`, `observability.yaml` | 로컬에 반영 |
| 스킬 파일명 | `moai-ref-*/skill.md` (소문자) | `SKILL.md`로 통일 |

### 2.3 테스트 갭

| 패키지 | 테스트 파일 | 상태 |
|--------|-----------|------|
| `internal/core/` (최상위) | 없음 | 삭제 예정 → 무관 |
| `internal/defs/` | 없음 | 상수만 포함 → 낮은 우선순위 |

### 2.4 미연결 시스템 (Unwired)

| 시스템 | 코드 위치 | 상태 |
|--------|----------|------|
| TRUST 5 Validators | `internal/core/quality/validators.go` | 5개 validator 정의됨, production에서 미사용 |
| Agent Hook Handlers | `internal/hook/agents/` | 10개 핸들러 정의됨, `deps.go`에서 미연결 |
| Worktree Validator | `internal/core/quality/worktree_validator.go` | 인터페이스 참조됨, 생성자 미사용 |

---

## Part 3: Agency + Autoresearch 통합 설계

### 3.1 현재 상태 vs 목표 상태

```
[현재]                              [목표]
┌────────────────────┐             ┌────────────────────────────────┐
│ moai-adk Core      │             │ moai-adk Core                  │
│ ├── hooks          │             │ ├── hooks                      │
│ ├── templates      │             │ ├── templates                  │
│ ├── loop (Ralph)   │             │ ├── loop (Ralph) ← 확장       │
│ └── workflow       │             │ ├── workflow                   │
├────────────────────┤             │ └── research/ ← NEW           │
│ Agency (별도)      │             │     ├── eval/    (바이너리)    │
│ ├── 6 agents       │    ──→     │     ├── experiment/ (루프)     │
│ ├── 5 skills       │             │     ├── observe/  (수동)      │
│ ├── Learner (자체) │             │     ├── safety/   (안전)      │
│ └── Evolution      │             │     └── dashboard/ (TUI+HTML) │
├────────────────────┤             ├────────────────────────────────┤
│ autoresearch       │             │ Agency (research 엔진 사용)    │
│ (외부, 미연결)     │             │ ├── 6 agents (유지)           │
│ ├── SKILL.md       │             │ ├── 5 skills (유지)           │
│ └── eval-guide.md  │             │ └── Learner → research 위임   │
└────────────────────┘             └────────────────────────────────┘
```

### 3.2 핵심 설계 원칙

1. **Research Engine은 Go 패키지** - eval 실행, 상태 관리, 안전 검증은 Go로 (빠르고 결정론적)
2. **변이(Mutation)는 Claude 에이전트** - 스킬/에이전트 내용 수정은 researcher 에이전트가 Edit 도구로
3. **Ralph 확장, 교체 아님** - 기존 `internal/loop` 상태 머신에 eval 단계를 추가
4. **Agency Learner는 thin wrapper** - 자체 진화 코드 대신 research 엔진에 위임
5. **기존 훅에 관찰 레이어 추가** - 새로운 훅 타입 불필요, 기존 핸들러에 observation writing 추가

### 3.3 새로운 Go 패키지 설계

```
internal/research/                    # Self-Research 코어
├── eval/                             # Binary Eval 엔진
│   ├── engine.go                    # EvalEngine 인터페이스
│   │   type EvalEngine interface {
│   │       LoadSuite(path string) (*EvalSuite, error)
│   │       RunBaseline(suite *EvalSuite, target string) (*Baseline, error)
│   │       Evaluate(suite *EvalSuite, output []byte) (*EvalResult, error)
│   │   }
│   ├── suite.go                     # EvalSuite YAML 로더/세이버
│   │   type EvalSuite struct {
│   │       Target    TargetSpec
│   │       Inputs    []TestInput
│   │       Criteria  []EvalCriterion
│   │       Settings  EvalSettings
│   │   }
│   ├── criterion.go                 # Binary 판정 로직
│   │   type EvalCriterion struct {
│   │       Name     string
│   │       Question string          // Yes/No 질문
│   │       Pass     string          // 통과 조건
│   │       Fail     string          // 실패 조건
│   │       Weight   CriterionWeight // must_pass | nice_to_have
│   │   }
│   ├── result.go                    # 점수 집계
│   │   type EvalResult struct {
│   │       Overall    float64
│   │       PerCriterion map[string]CriterionResult
│   │       MustPassOK bool
│   │   }
│   └── types.go
│
├── experiment/                       # 실험 루프
│   ├── loop.go                      # ExperimentLoop (Ralph 패턴 활용)
│   │   type ExperimentLoop struct {
│   │       controller *loop.LoopController  // 기존 Ralph 재사용
│   │       eval       eval.EvalEngine
│   │       safety     safety.SafetyChecker
│   │       store      ResultStore
│   │   }
│   │   func (l *ExperimentLoop) Run(ctx context.Context, target string) error
│   ├── baseline.go                  # Baseline 측정/저장/비교
│   ├── result.go                    # 실험 결과 저장
│   └── types.go                     # Experiment, Hypothesis, Mutation
│
├── observe/                          # 수동 관찰 수집
│   ├── collector.go                 # ObservationCollector
│   │   type ObservationCollector struct {
│   │       storage  ObservationStorage
│   │       detector PatternDetector
│   │   }
│   │   func (c *ObservationCollector) RecordCorrection(agent, file string, before, after []byte)
│   │   func (c *ObservationCollector) RecordFailure(agent, task string, err error)
│   │   func (c *ObservationCollector) RecordSuccess(agent, task string, score float64)
│   ├── pattern.go                   # 패턴 감지 (3x→heuristic, 5x→rule)
│   │   type PatternDetector struct {
│   │       threshold PatternThreshold  // {Heuristic:3, Rule:5, HighConf:10}
│   │   }
│   │   func (d *PatternDetector) Detect(obs []Observation) []Pattern
│   ├── storage.go                   # JSONL 파일 저장
│   └── types.go                     # Observation, Pattern, Classification
│
├── safety/                           # 안전 레이어
│   ├── frozen.go                    # FrozenGuard (수정 불가 파일 체크)
│   │   func IsFrozen(path string) bool
│   ├── canary.go                    # Canary 회귀 검증
│   │   func CanaryCheck(baselines []Baseline, proposed EvalResult) (bool, error)
│   ├── limiter.go                   # Rate Limiter
│   │   func CheckRateLimit(config RateLimitConfig) error
│   └── types.go
│
├── dashboard/                        # 연구 대시보드 (TUI 기본, HTML 선택)
│   ├── terminal.go                  # lipgloss/bubbletea 터미널 렌더링 (기본)
│   │   - 바 차트: 점수 진행 (baseline → current)
│   │   - 테이블: per-eval pass/fail 현황
│   │   - 요약: 실험 수, keep 비율, 잔여 실패
│   ├── html.go                      # Chart.js HTML 생성 + 브라우저 open (--html)
│   │   - `go tool cover -html` 패턴: 파일 생성 → exec.Command("open", path)
│   │   - 자동 갱신 (10초), 점수 추이 라인차트, eval별 breakdown
│   └── types.go                     # DashboardData (양쪽 렌더러 공유)
│
├── config.go                        # ResearchConfig (YAML 섹션)
└── types.go                         # 패키지 공통 타입
```

### 3.4 기존 패키지 확장

#### `internal/loop` 확장 (Ralph 상태 머신)

```go
// 현재 단계
type LoopPhase string
const (
    PhaseAnalyze   LoopPhase = "analyze"
    PhaseImplement LoopPhase = "implement"
    PhaseTest      LoopPhase = "test"
    PhaseReview    LoopPhase = "review"
)

// 추가 단계 (Research 모드)
const (
    PhaseEvaluate  LoopPhase = "evaluate"   // eval suite 실행
    PhaseMutate    LoopPhase = "mutate"     // 단일 변이 적용
    PhaseScore     LoopPhase = "score"      // 점수 비교 → keep/discard
)
```

#### `internal/ralph` 확장 (결정 엔진)

```go
// 기존 결정 유형
const (
    ActionContinue      = "continue"
    ActionConverge      = "converge"
    ActionRequestReview = "request_review"
    ActionAbort         = "abort"
)

// 추가 결정 유형 (Research 모드)
const (
    ActionKeep    = "keep"     // 변이 유지
    ActionDiscard = "discard"  // 변이 폐기
    ActionStagnate = "stagnate" // 정체 → 다른 차원 시도
)
```

#### `internal/hook` 확장 (관찰 수집)

```go
// 기존 PostToolUse 핸들러에 관찰 기록 추가
func (h *PostToolHandler) Handle(input *HookInput) (*HookOutput, error) {
    result := h.existing_logic(input)

    // NEW: 사용자 교정 감지 (에이전트 Edit 직후 사용자 Edit)
    if h.observer != nil && isCorrection(input) {
        h.observer.RecordCorrection(input.Agent, input.File, input.Before, input.After)
    }

    return result, nil
}
```

#### `internal/hook/agents` 재활용 (연구용 eval 핸들러)

```go
// 현재: 빈 스텁
func (h *backendHandler) Handle(input *hook.HookInput) (*hook.HookOutput, error) {
    // TODO: Implement backend-specific logic
    return hook.NewAllowOutput(), nil
}

// 변경: 연구용 평가 로직
func (h *backendHandler) Handle(input *hook.HookInput) (*hook.HookOutput, error) {
    // SubagentStop 이벤트에서 에이전트 결과 품질 평가
    if input.EventType == hook.EventSubagentStop {
        h.observer.RecordSuccess(
            "expert-backend",
            input.TaskName,
            h.evaluateBackendOutput(input.Output),
        )
    }
    return hook.NewAllowOutput(), nil
}
```

### 3.5 Agency 통합 방식

#### Agency Learner 변경

```markdown
# 현재 Learner EVOLVABLE ZONE (agent prompt 내 진화 로직)
1. Read .agency/learnings/learnings.md
2. Detect patterns: 3x -> Heuristic, 5x -> Rule
3. Validate against Brand Context
4. Check for contradictions
5. Generate evolution proposal with diff preview
6. On approval: modify skill SKILL.md Dynamic Zone
7. Bump version, create snapshot
8. Record in evolution-log.md
9. Archive applied learnings

# 변경 후 (research 엔진 위임)
1. Run `moai research --observe` to get pattern summary
2. Filter patterns relevant to Agency (brand, copy, design categories)
3. Validate against Brand Context (.agency/context/) ← Agency 전용 로직 유지
4. Run `moai research --evolve` for graduation proposals
5. Present to user for approval ← 기존 safety Layer 5 재사용
```

#### Agency 데이터 디렉토리 통합

```
# 현재 (분산)                        # 통합 후
.agency/learnings/                   → .moai/research/observations/agency/
.agency/learnings/archive/           → .moai/research/observations/agency/archive/
.agency/evolution/                   → .moai/research/experiments/agency/
.agency/evolution/snapshots/         → .moai/research/experiments/agency/snapshots/

# Agency 전용 유지 (이동 없음)
.agency/config.yaml                  ← 파이프라인 설정
.agency/context/                     ← 브랜드 컨텍스트 (인간 전용)
.agency/templates/                   ← BRIEF 템플릿
.agency/fork-manifest.yaml           ← 포크 추적
.agency/briefs/                      ← 프로젝트 BRIEF
```

### 3.6 통합 데이터 디렉토리

```
.moai/research/                              # 통합 연구 데이터
├── config.yaml                              # 연구 시스템 설정
├── evals/                                   # Eval suite 정의
│   ├── skills/
│   │   ├── moai-lang-go.eval.yaml
│   │   └── ...
│   ├── agents/
│   │   ├── expert-backend.eval.yaml
│   │   └── ...
│   ├── rules/
│   │   └── coding-standards.eval.yaml
│   └── agency/                              # Agency 전용 eval
│       ├── copywriter.eval.yaml
│       └── designer.eval.yaml
├── baselines/                               # Baseline 스냅샷
├── experiments/                             # 실험 로그
│   ├── moai-lang-go/
│   ├── expert-backend/
│   └── agency/                              # Agency 실험 (병합)
│       ├── copywriter/
│       └── designer/
├── observations/                            # 수동 관찰 (통합)
│   ├── corrections.jsonl                    # 사용자 교정
│   ├── failures.jsonl                       # 에이전트 실패
│   ├── successes.jsonl                      # 성공 기록
│   ├── agency/                              # Agency 관찰 (병합)
│   │   └── learnings.jsonl
│   └── patterns.yaml                        # 감지된 패턴
├── results/
│   ├── results.json
│   └── results.tsv
└── dashboard.html
```

### 3.7 CLI 명령 통합

```
# moai-adk CLI (Go binary)
moai research <target>                 # 능동적 실험 루프
moai research --observe                # 수동 관찰 요약
moai research --evolve                 # 졸업 제안
moai research --eval <target>          # eval suite 편집
moai research --baseline <target>      # baseline 측정
moai research --dashboard              # 터미널 TUI 대시보드 (기본, lipgloss)
moai research --dashboard --html       # HTML 생성 + 브라우저 open (go tool cover 패턴)
moai research --status                 # 현재 상태 (한 줄 요약)

# TUI 대시보드 출력 예시:
#  Research: moai-lang-go ──────────────────────
#  Baseline: 0.533  →  Current: 0.867  (+0.334)
#  ████████████████████░░░░░  86.7%  (target: 95%)
#
#  Experiments: 7/20  │  Keep: 4  │  Discard: 3
#
#  Per-Eval:
#   compiles        ███████████████████████████ 100%  MUST
#   error-handling  ███████████████████████████ 100%  MUST
#   modern-go       ██████████████████░░░░░░░░░  67%
#   no-placeholders ██████████████████░░░░░░░░░  67%
#   test-included   ██████████████████░░░░░░░░░  67%
#  ──────────────────────────────────────────────

# Claude Code 슬래시 명령 (agent orchestration)
/moai research <target>                # Researcher 에이전트 실행
/moai research --observe               # 관찰 분석 + 연구 트리거 제안
/moai research --evolve                # 졸업 제안 → 사용자 승인

# Agency 명령 (변경 없음, 내부적으로 research 사용)
/agency learn                          # → moai research --observe (agency 필터)
/agency evolve                         # → moai research --evolve (agency 필터)
```

### 3.8 설정 스키마

```yaml
# .moai/config/sections/research.yaml (NEW)
research:
  version: "1.0.0"
  enabled: true

  # 수동 관찰
  passive:
    enabled: true
    correction_window_seconds: 60      # Edit 후 N초 내 재Edit = 교정
    pattern_thresholds:
      heuristic: 3                     # 3회 → 연구 후보
      rule: 5                          # 5회 → 자동 연구 트리거
      high_confidence: 10              # 10회 → 자동 제안
    retention_days: 90

  # 능동 실험
  active:
    defaults:
      runs_per_experiment: 3
      max_experiments: 20
      pass_threshold: 0.80
      target_score: 0.95
      budget_cap_tokens: 500000
      stagnation_threshold: 0.03
      stagnation_patience: 2
    tier_overrides:
      agent:
        runs_per_experiment: 2
        max_experiments: 10
        budget_cap_tokens: 1000000
      rule:
        max_experiments: 5
      config:
        max_experiments: 5

  # 안전
  safety:
    worktree_isolation: true
    canary_regression_threshold: 0.10
    canary_baseline_count: 3
    auto_accept: false
    auto_accept_threshold: 0.15
    require_approval_for_rules: true
    rate_limits:
      max_experiments_per_session: 20
      max_accepted_per_session: 5
      max_auto_research_per_week: 3
      cooldown_hours: 24

  # 대시보드
  dashboard:
    default_mode: terminal             # terminal | html
    html_auto_refresh_seconds: 10      # HTML 모드에서만 적용
    html_open_browser: true            # HTML 생성 후 자동 브라우저 open
```

---

## Part 4: 구현 로드맵

### Phase 0: 데드코드 정리 (1주)

```
- [ ] internal/core/ (최상위) 삭제
- [ ] 미사용 export 삭제 (P2 목록)
- [ ] SPEC ID 정규식 통합 → workflow/specid.go
- [ ] defs/timeouts.go 삭제 (foundation으로 통합)
- [ ] 훅 이름 불일치 수정
- [ ] 3개 config 섹션 로컬 반영
- [ ] moai-ref-* skill.md → SKILL.md 통일
- [ ] go test ./... 전체 통과 확인
```

### Phase 1: Research 기반 패키지 (2주)

```
- [ ] internal/research/eval/ 구현 (EvalEngine, EvalSuite, Criterion)
- [ ] internal/research/safety/ 구현 (FrozenGuard, CanaryCheck, RateLimiter)
- [ ] internal/research/config.go (YAML 스키마)
- [ ] .moai/config/sections/research.yaml 템플릿
- [ ] .moai/research/ 디렉토리 구조 생성
- [ ] 단위 테스트 85%+
```

### Phase 2: 실험 루프 + CLI (2주)

```
- [ ] internal/research/experiment/ 구현 (Loop, Baseline, ResultStore)
- [ ] internal/loop 확장 (PhaseEvaluate, PhaseMutate, PhaseScore)
- [ ] internal/ralph 확장 (ActionKeep, ActionDiscard, ActionStagnate)
- [ ] internal/cli/research.go (moai research 명령)
- [ ] internal/cli/deps.go에 research 의존성 주입
- [ ] .claude/skills/moai-workflow-research/SKILL.md 작성
- [ ] .claude/agents/moai/researcher.md 작성
```

### Phase 3: 수동 관찰 시스템 (2주)

```
- [ ] internal/research/observe/ 구현 (Collector, PatternDetector, Storage)
- [ ] internal/hook에 관찰 기록 통합
- [ ] internal/hook/agents/ 재활용 (연구용 eval 핸들러)
- [ ] Agency learnings → .moai/research/observations/agency/ 마이그레이션
- [ ] Agency Learner 에이전트 프롬프트 업데이트
```

### Phase 4: Agency 통합 (1주)

```
- [ ] Agency evolution 디렉토리 → research 디렉토리 마이그레이션
- [ ] Agency Learner → research 엔진 위임 패턴 적용
- [ ] /agency learn → moai research --observe (agency 필터) 연결
- [ ] /agency evolve → moai research --evolve (agency 필터) 연결
- [ ] Agency 전용 eval suite 작성 (copywriter, designer, builder)
- [ ] 브랜드 컨텍스트 검증은 Agency Learner에 유지
```

### Phase 5: 대시보드 + 스킬 eval (2주)

```
- [ ] internal/research/dashboard/ 구현 (terminal.go: lipgloss TUI, html.go: Chart.js + browser open)
- [ ] 주요 스킬 eval suite 작성 (moai-lang-go, moai-domain-backend, moai-lang-typescript)
- [ ] 주요 에이전트 eval suite 작성 (expert-backend, manager-tdd)
- [ ] /moai research 슬래시 커맨드 통합
- [ ] End-to-end 테스트: 스킬 1개에 대한 전체 연구 사이클
```

---

## Part 5: 아키텍처 결정 기록

### ADR-1: Research Engine을 Go 패키지로 구현

**결정**: Eval 실행, 상태 관리, 안전 검증은 Go로 구현. 변이(mutation)만 Claude 에이전트에 위임.

**이유**:
- Binary eval 판정은 결정론적이어야 함 (LLM 비결정성 제거)
- 실험 상태는 crash-resistant 해야 함 (프로세스 종료 후 재개 가능)
- 안전 레이어(frozen guard, rate limit)는 LLM이 우회할 수 없어야 함
- Go 코드는 `moai research` CLI 명령으로 직접 실행 가능 (Claude 없이도)

### ADR-2: Ralph 확장, 교체 아님

**결정**: 기존 `internal/loop` + `internal/ralph`에 새 단계를 추가. 새로운 루프 엔진 미생성.

**이유**:
- Ralph은 이미 상태 머신 + 결정 엔진 + 지속성을 구현
- Research 루프는 Ralph 루프의 특수 형태 (eval 기반 scoring)
- 코드 중복 방지

### ADR-3: Agency Learner를 thin wrapper로 변경

**결정**: Learner의 자체 진화 파이프라인 코드를 제거하고 research 엔진에 위임.

**이유**:
- 진화 로직 중복 제거 (Agency Learner vs Research Engine)
- Agency 전용 로직(브랜드 컨텍스트 검증)만 Learner에 유지
- 단일 졸업 프로토콜로 통일

### ADR-4: 관찰 데이터를 통합 디렉토리에 저장

**결정**: Agency의 `.agency/learnings/`를 `.moai/research/observations/agency/`로 마이그레이션.

**이유**:
- 단일 관찰 저장소로 패턴 감지의 시야 확대
- Agency 관찰과 moai 관찰을 교차 분석 가능
- Agency 전용 필터로 기존 기능 유지

### ADR-5: hook/agents 패키지를 연구 eval 핸들러로 재활용

**결정**: 현재 TODO 스텁인 10개 에이전트 핸들러를 연구용 품질 평가 로직으로 재구현.

**이유**:
- 패키지 구조와 인터페이스가 이미 존재
- SubagentStop 이벤트에서 에이전트 결과 품질을 자동 측정
- 수동 관찰의 핵심 데이터 소스가 됨

---

**Version**: 2.0.0 (Unified Design)
**Classification**: Design Proposal
**Next Step**: GOOS행님 검토 후 Phase 0 (데드코드 정리) 시작
