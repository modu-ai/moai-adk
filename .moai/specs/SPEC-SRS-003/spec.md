---
id: SPEC-SRS-003
version: "1.0.0"
status: draft
created: "2026-04-09"
updated: "2026-04-09"
author: GOOS
priority: high
issue_number: 0
---

# SPEC-SRS-003: Dashboard + CLI + Agency 통합

## Overview

Self-Research System의 TUI 대시보드, `moai research` CLI 명령, 그리고
Agency Learner의 research 엔진 위임 패턴을 구현한다.

Source: `.moai/plans/DESIGN-SRS-002-unified.md` Phase 4 + Phase 5

## What NOT to Build (Exclusions)

- HTML 대시보드 (`--html` 모드) - TUI 기본 모드만 구현
- Researcher 에이전트의 실제 AI 변이 로직 - 인프라만
- Hook handler 내 관찰 수집 코드 통합 - 별도 SPEC
- `moai research` 전체 실험 루프 자동화 - 상태 표시 + baseline 측정만

---

## Requirements

### REQ-1: TUI Dashboard (`internal/research/dashboard/`)

**[EARS: Event-driven]**

WHEN `moai research --status` is invoked, the system SHALL render a terminal dashboard showing research progress using lipgloss.

```go
type DashboardData struct {
    Target       string
    Baseline     float64
    CurrentScore float64
    TargetScore  float64
    Experiments  int
    MaxExperiments int
    KeepCount    int
    DiscardCount int
    PerCriterion []CriterionStatus
}

type CriterionStatus struct {
    Name     string
    PassRate float64
    Weight   string // "MUST" or ""
}

func RenderDashboard(data *DashboardData) string
func RenderCompact(data *DashboardData) string
```

출력 예시:
```
 Research: moai-lang-go ──────────────────────
 Baseline: 0.533  →  Current: 0.867  (+0.334)
 ████████████████████░░░░░  86.7%  (target: 95%)

 Experiments: 7/20  │  Keep: 4  │  Discard: 3

 Per-Eval:
  compiles        ███████████████████████████ 100%  MUST
  error-handling  ███████████████████████████ 100%  MUST
  modern-go       ██████████████████░░░░░░░░░  67%
  test-included   ██████████████████░░░░░░░░░  67%
```

Acceptance:
- `RenderDashboard`가 위 형식의 문자열을 반환
- 프로그레스 바 폭은 25자 고정
- MUST 레이블은 must_pass criteria에만 표시
- 점수 변화량에 색상: 양수=녹색, 음수=빨간색, 0=회색
- lipgloss 사용 (이미 go.mod 의존성)
- `RenderCompact`는 1줄 요약 (상태라인용)
- 테스트 85%+

### REQ-2: `moai research` CLI 명령 (`internal/cli/research.go`)

**[EARS: Event-driven]**

WHEN `moai research` subcommand is invoked, the system SHALL provide research status and baseline management.

서브커맨드:
```
moai research status              # TUI 대시보드 표시
moai research baseline <target>   # baseline 측정 (eval suite 실행)
moai research list                # 등록된 eval suite 목록
```

```go
func newResearchCmd() *cobra.Command
func newResearchStatusCmd() *cobra.Command
func newResearchBaselineCmd() *cobra.Command
func newResearchListCmd() *cobra.Command
```

Acceptance:
- `moai research status` → 가장 최근 실험의 대시보드 출력
- `moai research baseline <target>` → eval suite 로드 → 현재 상태 baseline으로 저장
- `moai research list` → `.moai/research/evals/` 디렉토리의 eval YAML 목록
- 데이터가 없으면 안내 메시지 출력
- cobra 명령으로 root에 등록
- 테스트 가능한 구조 (io.Writer 주입)

### REQ-3: Researcher 에이전트 정의

**[EARS: Ubiquitous]**

The system SHALL provide a researcher agent definition and research workflow skill.

파일:
- `.claude/agents/moai/researcher.md` - Researcher 에이전트
- `.claude/skills/moai-workflow-research/SKILL.md` - Research 워크플로우 스킬

에이전트 정의:
```yaml
name: researcher
description: >
  Active self-research agent that optimizes moai-adk components
  through iterative experimentation with binary eval criteria.
tools: Read, Write, Edit, Grep, Glob, Bash
model: opus
permissionMode: acceptEdits
memory: project
skills:
  - moai-workflow-research
```

스킬 정의:
```yaml
name: moai-workflow-research
description: >
  Self-research workflow for optimizing moai-adk components through
  binary eval experimentation loops.
user-invocable: false
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
```

Acceptance:
- 에이전트 파일이 frontmatter 스키마를 준수
- 스킬 파일이 frontmatter 스키마를 준수
- 템플릿 디렉토리에도 동일 파일 생성

### REQ-4: Agency Learner 통합

**[EARS: Ubiquitous]**

The system SHALL update the Agency Learner agent to delegate evolution analysis to the research engine.

변경:
- `.claude/agents/agency/learner.md` EVOLVABLE ZONE 수정
- 기존 자체 진화 파이프라인 → `moai research --observe` + `moai research --evolve` 위임 패턴

수정 후 Learner Evolution Pipeline:
```
1. Analyze observations: moai research --observe (CLI) 결과 참조
2. Filter agency-relevant patterns (brand, copy, design 카테고리)
3. Validate against Brand Context (.agency/context/)
4. Propose evolution via moai research --evolve (CLI) 위임
5. Present to user for approval
6. On approval: apply changes, bump version, snapshot
```

Acceptance:
- Learner EVOLVABLE ZONE이 research 엔진 위임 패턴으로 업데이트
- FROZEN ZONE은 변경 없음
- Brand Context 검증 로직은 Learner에 유지
- 템플릿 디렉토리에도 동일 변경 적용

### REQ-5: Eval Suite 예시 파일

**[EARS: Ubiquitous]**

The system SHALL include example eval suite YAML files for reference.

파일:
- `.moai/research/evals/skills/moai-lang-go.eval.yaml` (예시)

Acceptance:
- YAML 스키마가 `internal/research/eval` 패키지의 EvalSuite 타입과 일치
- 최소 3개 test_inputs, 4개 evals
- must_pass와 nice_to_have 모두 포함

---

## Technical Approach

### 파일 생성

**Go 패키지:**
- `internal/research/dashboard/terminal.go` + `terminal_test.go`
- `internal/research/dashboard/types.go`
- `internal/cli/research.go` + `research_test.go`

**에이전트/스킬:**
- `.claude/agents/moai/researcher.md`
- `.claude/skills/moai-workflow-research/SKILL.md`
- `internal/template/templates/.claude/agents/moai/researcher.md`
- `internal/template/templates/.claude/skills/moai-workflow-research/SKILL.md`

**Agency:**
- `.claude/agents/agency/learner.md` (EVOLVABLE ZONE 수정)
- `internal/template/templates/.claude/agents/agency/learner.md` (동일)

**예시:**
- `.moai/research/evals/skills/moai-lang-go.eval.yaml`

### 의존성

```
dashboard/ → lipgloss (이미 go.mod에 있음)
cli/research.go → research/dashboard, research/eval, research/experiment
```
