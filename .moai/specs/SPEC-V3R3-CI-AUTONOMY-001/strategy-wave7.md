---
spec: SPEC-V3R3-CI-AUTONOMY-001
wave: 7
version: 0.1.0
status: draft
created_at: 2026-05-09
updated_at: 2026-05-09
author: manager-spec
---

# Wave 7 Execution Strategy (Phase 1 Output)

> Audit trail. manager-spec output for Wave 7 of SPEC-V3R3-CI-AUTONOMY-001 — T8 Branch Origin Decision Protocol (BODP).
> Generated: 2026-05-09. Methodology: TDD (Go pure-library + table-driven test fixtures + skill body extension + CLI flag extension).
> Wave Base: `origin/main 78929d058` (post-Wave 6 PR #793 머지 baseline; Wave 7은 다른 Wave와 무관 독립 실행).
> Final Wave (7/7): 본 Wave 머지 후 SPEC-V3R3-CI-AUTONOMY-001 closure 진입.

---

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-09 | manager-spec | Initial draft. Wave 7 — T8 BODP (P0). 6 atomic tasks (W7-T01..W7-T06) covering pure-Go relatedness library + audit trail writer + skill Phase 3 extension + `moai worktree new` `--from-current` flag + `moai status` off-protocol reminder + CLAUDE.local.md §18.12 documentation. Architecture: 3 existing entry points (`/moai plan --branch`, `/moai plan --worktree`, `moai worktree new`) reuse a shared `internal/bodp/` library — ZERO new slash commands or CLI subcommands per plan.md §10.5 (resolves user critique 2026-05-05 "추가 명령어는 자제"). REQ mapping: REQ-CIAUT-042..051 all bound. AC mapping: AC-CIAUT-018/019/019b/024/025. Independent of all other Waves (plan.md §9 line 304). |

---

## 1. Wave 7 Goal & Scope

### 1.1 Goal

Wave 7 는 **새 슬래시 명령어 또는 CLI 서브명령어 ZERO** 원칙 아래, 기존 3개 entry point (`/moai plan --branch`, `/moai plan --worktree`, `moai worktree new`)에 BODP (Branch Origin Decision Protocol) 동작을 내장한다. 신규 SPEC 작업 진입 시점에 새 브랜치 base 결정을 자동화하여, 5-PR sweep (2026-05-05) 사건의 R7 ("새 브랜치가 main이 아닌 현재 브랜치 base로 default") + P10 ("Sub-agent가 plan 파일을 untracked 상태로 남김; 새 브랜치가 main이 아닌 feature branch에서 분기됨") 근본 원인을 해결한다.

핵심 동작:

1. **3-signal relatedness check**: 새 SPEC이 (a) 현재 브랜치 commit과 의존, (b) untracked plan 파일과 SPEC-ID 매칭, (c) 현재 브랜치를 head로 하는 open PR 보유 — 세 시그널을 평가
2. **Default = origin/main 분기**: 모든 시그널 negative 시 main 분기를 권장 옵션으로 제시 (CLAUDE.local.md §18 Enhanced GitHub Flow 정합)
3. **AskUserQuestion gate** (slash 경로): 시그널이 positive면 stacked / continue / main 옵션을 제시 + "Other" 자동 포함
4. **CLI flag-only path** (`moai worktree new`): orchestrator-only HARD 준수, AskUserQuestion 미사용, default `origin/main` + `--from-current` opt-out flag
5. **Audit trail**: 모든 결정 사항을 `.moai/branches/decisions/<branch>.md`에 기록 (timestamp, entry point, signals, choice, command)
6. **Off-protocol reminder**: raw `git checkout -b` 우회 시 다음 `moai status` 호출에서 친절한 reminder 출력 (block 안 함)

User-facing surface는 **변경 없이 안전성 향상**: 사용자는 새 명령어를 학습할 필요 없으며, 기존 `/moai plan --branch` / `/moai plan --worktree` / `moai worktree new` 흐름이 자동으로 더 안전해진다.

### 1.2 What This Wave Does NOT Do (P0 priority but minimal-surface rationale)

- Does NOT introduce new slash commands (예: `/moai branch new`) — plan.md §10.5 + 사용자 명시 critique (2026-05-05)
- Does NOT introduce new CLI subcommands (예: `moai branch new`) — 동일 critique
- Does NOT intercept raw `git checkout -b` invocations — REQ-CIAUT-050 + spec.md §7 OQ4 RESOLVED ("opt-in 유지")
- Does NOT auto-delete existing branches — Out of Scope per spec.md §3.8 design principle
- Does NOT migrate already-existing off-protocol branches — `moai status` reminder만 출력
- Does NOT modify Wave 1-6 산출물 (ci-mirror framework, ci-watch skill, worktree state guard, i18n validator 모두 read-only consumer 관점에서 참조)
- Does NOT propagate BODP to monorepo path-based stacked PR 자동 분해 — spec.md §2 Out of Scope 명시
- Does NOT plumb BODP into `/moai run` Phase 2 (구현 단계는 base 결정 이후 단계)

P0 priority 정당화: R7/P10은 5-PR sweep 가시 사건의 직접 원인 (사용자 명시 직접 지시 "main에서 분기 + T8 추가 (권장)") 이지만, **구현 surface가 작고** (3개 entry point 확장 + 1개 신규 라이브러리 + 문서 1개) **Wave 7 단독으로 즉시 가치 제공** 가능하다. 따라서 P0이지만 Wave 7 (final) 으로 배치하여 다른 Wave와 충돌 없는 독립 머지를 보장한다.

### 1.3 Why a Pure-Go Library (not a CLI tool)

Plan.md §9 specifies `internal/bodp/relatedness.go` as a Go module shared by 3 entry points. Rationale:

- **CLI 도입 금지** (사용자 critique 2026-05-05): CLI subcommand 추가 시 사용자가 새 명령어를 학습해야 함; 본 Wave는 인지 부담 ZERO 원칙 준수
- **3개 entry point 공유**: `/moai plan` (orchestrator-driven Skill body), `moai worktree new` (Go CLI), `moai status` (Go CLI) 가 동일 라이브러리 호출 — Pure-Go 모듈이 가장 자연스러운 공유 surface
- **Testability**: AST/git CLI shell-out과 달리 BODP 시그널 검사는 결정론적 함수 (3개 입력 → 3개 boolean) — 테이블 기반 단위 테스트로 4-5개 case 완전 커버
- **No external dependencies**: `os/exec` (git CLI shell-out), `path/filepath`, `os` (file existence) 만 사용; CLAUDE.local.md §14 no-hardcoding 준수

---

## 2. Requirements Mapping

| REQ ID | Description (verbatim from spec.md §3.8) | Implementing Task |
|--------|------------------------------------------|-------------------|
| REQ-CIAUT-042 (Event-Driven) | When `/moai plan --branch` is invoked, the orchestrator shall run BODP relatedness check before delegating to manager-git, present base-branch choice via AskUserQuestion when signals positive, and pass chosen base to manager-git as parameter. | W7-T02 (skill Branch Path 확장) |
| REQ-CIAUT-043 (Event-Driven) | When `/moai plan --worktree` is invoked, the orchestrator shall run same BODP check before invoking `moai worktree new`. Chosen base shall be propagated via new `--base <branch>` flag. | W7-T02 (skill Worktree Path 확장) + W7-T03 (`--base` flag 신설) |
| REQ-CIAUT-043b (Event-Driven) | When `moai worktree new <SPEC-ID>` CLI is invoked directly, it shall default to `origin/main` as base. `--from-current` flag shall opt-out. CLI shall NOT invoke AskUserQuestion (orchestrator-only HARD). | W7-T03 |
| REQ-CIAUT-044 (Event-Driven) | BODP relatedness check shall evaluate three signals on every invocation: (a) SPEC depends_on / file-path overlap with current branch diff, (b) untracked SPEC-ID directory match, (c) open PR with current branch as head. | W7-T01 (라이브러리 핵심 함수) |
| REQ-CIAUT-045 (Event-Driven) | When all 3 signals negative, orchestrator (slash path) shall recommend "main에서 분기" as first option marked `(권장)`. CLI path uses `origin/main` automatically without prompting. | W7-T01 (decision matrix) + W7-T02 (skill AskUserQuestion 옵션 순서) |
| REQ-CIAUT-046 (Event-Driven) | When signal (a) positive, orchestrator shall recommend "현재 브랜치에서 분기 (stacked PR)" as first option with rationale. | W7-T01 (decision matrix) + W7-T02 |
| REQ-CIAUT-047 (Event-Driven) | When signal (b) positive, orchestrator shall recommend "현재 브랜치에 계속 작업" as first option (no new branch). | W7-T01 (decision matrix) + W7-T02 |
| REQ-CIAUT-047b (Event-Driven) | When signal (c) positive, orchestrator shall recommend "stacked PR (base=현재 브랜치)" as first option + warn about parent-merge gotcha (CLAUDE.local.md §18.11 Case Study). | W7-T01 (decision matrix) + W7-T02 |
| REQ-CIAUT-048 (Ubiquitous) | After confirmation (slash AskUserQuestion or CLI flag default), the existing branch-creation handler shall execute appropriate git commands with chosen base (main: `git fetch origin main && git checkout -B <new> origin/main`; stacked: `git checkout -B <new>`; continue: no-op). | W7-T02 (manager-git parameter) + W7-T03 (worktree CLI git invocation) |
| REQ-CIAUT-049 (Ubiquitous) | Every BODP decision shall be recorded at `.moai/branches/decisions/<branch-name>.md` with timestamp, invocation entry point, current branch, relatedness signals (a/b/c with evidence), user choice, executed command. | W7-T04 (audit trail writer) |
| REQ-CIAUT-050 (Unwanted Behavior) | If user invokes raw `git checkout -b` (bypassing all MoAI entry points), then system shall NOT intervene, but next `moai status` invocation shall detect off-protocol branch (absence of audit trail) and emit friendly reminder. | W7-T05 (`moai status` 확장) |
| REQ-CIAUT-051 (Ubiquitous) | CLAUDE.local.md §18 (Enhanced GitHub Flow) shall be amended with new subsection §18.12 documenting algorithm, three existing entry points, raw-git-checkout reminder. **No new slash command or CLI subcommand introduced.** | W7-T06 (문서화 + Template-First mirror) |

All 11 spec.md §3.8 requirements (REQ-042..051 including 043b, 047b) are mapped. No requirement is unbound. No task is unmapped.

---

## 3. Acceptance Criteria Mapping

| AC ID | Description (verbatim from acceptance.md) | Validating Task | Verification |
|-------|-------------------------------------------|-----------------|--------------|
| AC-CIAUT-018 | BODP recommends "main에서 분기" via `/moai plan --branch` (current session replay: chore/translation-batch-b + untracked SPEC-MX-INJECT-001 + 신규 SPEC-CI-AUTONOMY-001 → main 권장). | W7-T01 + W7-T02 + W7-T04 | `internal/bodp/relatedness_test.go` table-driven test case `TestRelatedness_AllNegative_RecommendsMain` (current session 시나리오 replay) + skill 통합 manual test (Phase 2 단계). Audit trail 파일 생성 검증. |
| AC-CIAUT-019 | BODP recommends stacked PR via `/moai plan --branch` (signal a positive: SPEC-AUTH-002 with `depends_on: [SPEC-AUTH-001]` on current branch `feat/SPEC-AUTH-001-base`). | W7-T01 + W7-T02 + W7-T04 | `internal/bodp/relatedness_test.go` 케이스 `TestRelatedness_SignalA_RecommendsStacked` (depends_on 시뮬레이션) + manager-git에 `base=feat/SPEC-AUTH-001-base` parameter 전달 검증. |
| AC-CIAUT-019b | BODP via `moai worktree new` CLI (default origin/main + `--from-current` opt-out). | W7-T03 | `internal/cli/worktree/new_test.go` 케이스 `TestNew_DefaultBaseIsOriginMain` + `TestNew_FromCurrentFlagPreservesOldBehavior` + `TestNew_NoAskUserQuestion` (orchestrator-only HARD 준수 verify). |
| AC-CIAUT-024 | `moai status` off-protocol branch reminder (audit trail 부재 시 친절한 reminder, hard-block 안 함, `MOAI_NO_BODP_REMINDER=1` 환경변수로 비활성화). | W7-T05 | `internal/cli/status_test.go` 케이스 `TestStatus_OffProtocolBranchReminder` + `TestStatus_AuditTrailDirAbsentNoFalsePositive` + `TestStatus_EnvVarDisablesReminder`. |
| AC-CIAUT-025 | CLAUDE.local.md §18.12 BODP subsection added (3개 entry point 명시, default = "main에서 분기", audit trail 위치, opt-out 방법). | W7-T06 | `grep -A 30 '§18.12' CLAUDE.local.md` 출력 검증 + `internal/template/templates/CLAUDE.local.md` mirror diff 검증 + `.claude/rules/moai/development/branch-origin-protocol.md` 신규 rule 파일 검증. |

5개 Wave 7 관련 AC 항목 모두 validating task에 매핑됨. AC-CIAUT-018/019는 동일 라이브러리 (W7-T01) + 동일 skill 확장 (W7-T02) 의 다른 시그널 조합으로 검증되는 자매 케이스 (table-driven test의 인접 row).

---

## 4. Architecture

### 4.1 3-Signal BODP Algorithm

Algorithm spec.md §3.8 REQ-CIAUT-044 + plan.md §10.5 정의:

```
Input:
  - currentBranch: string (예: "chore/translation-batch-b")
  - newSpecID:     string (예: "SPEC-V3R3-CI-AUTONOMY-001")
  - workingTree:   git working tree state (status --porcelain, diff vs main)

Signals (independent, evaluated in parallel):
  Signal (a) — Code dependency
    - SPEC newSpecID 의 frontmatter `depends_on:` 가 현재 브랜치에서 in-progress 인 SPEC을 참조?
    - 현재 브랜치 diff vs main 의 변경 파일 경로가 newSpecID 의 module/scope과 겹침?
    → boolean

  Signal (b) — Working tree co-location
    - `git status --porcelain` untracked entries 중 `.moai/specs/<newSpecID>/` 매칭?
    → boolean

  Signal (c) — Open PR head
    - `gh pr list --head <currentBranch> --state open` 결과 ≥1?
    → boolean

Output:
  BODPDecision struct {
    SignalA       bool
    SignalB       bool
    SignalC       bool
    Recommended   string  // "main" | "stacked" | "continue"
    Rationale     string  // 사용자에게 보여줄 한국어 근거
    BaseBranch    string  // "origin/main" | "<currentBranch>" | ""
  }
```

Decision matrix (3-signal × 권장 옵션):

| Signal (a) | Signal (b) | Signal (c) | Recommended | BaseBranch | Rationale (Korean) |
|------------|------------|------------|-------------|------------|---------------------|
| F | F | F | `main` | `origin/main` | 현재 브랜치와 무관한 새 작업이므로 main 분기를 권장합니다. |
| T | F | F | `stacked` | `<currentBranch>` | SPEC depends_on 또는 코드 경로 의존이 감지되어 stacked PR을 권장합니다. |
| F | T | F | `continue` | (no branch) | 현재 브랜치에 untracked SPEC plan 파일이 있어 계속 작업을 권장합니다. |
| F | F | T | `stacked` | `<currentBranch>` | 현재 브랜치를 head로 하는 open PR이 있어 stacked PR을 권장합니다 (parent-merge gotcha 주의: CLAUDE.local.md §18.11). |
| T | F | T | `stacked` | `<currentBranch>` | 코드 의존 + open PR 이중 시그널: stacked PR 권장. |
| T | T | F | `continue` | (no branch) | untracked plan + 코드 의존 동시 발생: 현재 브랜치 계속 작업 권장 (충돌 방지). |
| F | T | T | `continue` | (no branch) | untracked plan + open PR: 현재 브랜치 계속 작업 권장. |
| T | T | T | `continue` | (no branch) | 모든 시그널 positive: 안전한 선택은 현재 브랜치에 머무르는 것입니다. |

Truth table 8 rows 모두 결정 가능 (no-undefined). Rationale 문자열은 conversation_language=ko 준수.

> **Replay scenario (plan.md §9 line 297)**: chore/translation-batch-b + untracked `.moai/specs/SPEC-V3R3-MX-INJECT-001/` + 사용자 요청 "신규 SPEC-CI-AUTONOMY-001 작성"
> - Signal (a): SPEC-CI-AUTONOMY-001 ≠ chore/translation-batch-b 의 i18n 파일 → F
> - Signal (b): untracked SPEC-MX-INJECT-001 ≠ 새 SPEC-CI-AUTONOMY-001 (다른 SPEC ID) → F
> - Signal (c): chore/translation-batch-b head open PR 없음 (이미 머지) → F
> - **결과: `main` 권장** ✓ AC-CIAUT-018 충족

### 4.2 3 Entry Points + Shared Library Architecture

```
                      ┌─────────────────────────────────────┐
                      │   internal/bodp/                    │
                      │   ┌──────────────────────────────┐  │
                      │   │ relatedness.go               │  │
                      │   │   func Check(ctx, ...) → BODPDecision │
                      │   │   3-signal + decision matrix │  │
                      │   └──────────────────────────────┘  │
                      │   ┌──────────────────────────────┐  │
                      │   │ audit_trail.go               │  │
                      │   │   func WriteDecision(...)    │  │
                      │   │   .moai/branches/decisions/  │  │
                      │   └──────────────────────────────┘  │
                      └────┬───────────┬──────────────┬─────┘
                           │           │              │
        ┌──────────────────┘           │              └──────────────────┐
        │                              │                                  │
        ▼                              ▼                                  ▼
┌────────────────────┐   ┌──────────────────────────┐   ┌────────────────────────────┐
│ /moai plan         │   │ /moai plan --worktree    │   │ moai worktree new <SPEC-ID>│
│   --branch         │   │   (skill Worktree Path)  │   │   (Go CLI direct)          │
│   (skill Branch    │   │                          │   │                            │
│    Path)           │   │ orchestrator → BODP      │   │ default base=origin/main   │
│                    │   │   → AskUserQuestion      │   │ --from-current opt-out     │
│ orchestrator       │   │   → moai worktree new    │   │ NO AskUserQuestion         │
│   → BODP           │   │       <SPEC> --base <X>  │   │ (orchestrator-only HARD)   │
│   → AskUserQuestion│   │                          │   │ → BODP audit trail         │
│   → manager-git    │   │                          │   │                            │
│      (with base    │   │                          │   │                            │
│       param)       │   │                          │   │                            │
└────────────────────┘   └──────────────────────────┘   └────────────────────────────┘
        │                              │                                  │
        └──────────────────────────────┼──────────────────────────────────┘
                                       ▼
                          ┌─────────────────────────────────┐
                          │ .moai/branches/decisions/       │
                          │   <branch-name>.md              │
                          │   (audit trail; 1 entry/branch) │
                          └─────────────────────────────────┘

                                       │
                                       ▼ (off-protocol detection 우회 경로)
                          ┌─────────────────────────────────┐
                          │ raw `git checkout -b ...`       │
                          │   (BODP 무관)                   │
                          │                                 │
                          │ next `moai status` invocation:  │
                          │   audit_trail 부재 → reminder   │
                          │   (block 안 함, MOAI_NO_BODP_   │
                          │    REMINDER=1 비활성화 가능)    │
                          └─────────────────────────────────┘
```

**핵심 설계**: 3개 entry point 모두 동일 `internal/bodp/` 라이브러리 호출, **새 명령어 ZERO**. CLI path는 AskUserQuestion 미사용 (orchestrator-only HARD 준수).

### 4.3 Package Structure

```
internal/
└── bodp/                                  # NEW Go package (Wave 7 신규)
    ├── relatedness.go                     # NEW (W7-T01: 3-signal + decision matrix + BODPDecision struct)
    ├── relatedness_test.go                # NEW (W7-T01: 테이블 기반 4-5 case)
    ├── audit_trail.go                     # NEW (W7-T04: .moai/branches/decisions/<branch>.md writer)
    └── audit_trail_test.go                # NEW (W7-T04: write/read/missing-dir cases)

internal/cli/
├── worktree/
│   └── new.go                             # EXTEND (W7-T03: --base flag 신설 + --from-current opt-out + default origin/main + audit trail 호출)
└── status.go                              # EXTEND (W7-T05: off-protocol branch detection + friendly reminder + MOAI_NO_BODP_REMINDER 환경변수)

.claude/skills/moai/workflows/
└── plan.md                                # EXTEND (W7-T02: Phase 3 Branch Path + Worktree Path 둘 다 BODP 검사 + AskUserQuestion + manager-git base parameter 전달)

.claude/rules/moai/development/
└── branch-origin-protocol.md              # NEW (W7-T06: BODP rule, paths frontmatter로 자동 적용 범위 정의)

internal/template/templates/.claude/rules/moai/development/
└── branch-origin-protocol.md              # NEW (W7-T06: Template-First mirror)

internal/template/templates/CLAUDE.local.md  # EXTEND (W7-T06: §18.12 mirror)
CLAUDE.local.md                              # EXTEND (W7-T06: §18.12 신규 subsection)

.moai/branches/
└── decisions/
    └── .gitkeep                           # NEW (W7-T04: 디렉토리 보존; .gitignore 미포함)
```

### 4.4 Public API (Go) — `internal/bodp/relatedness.go`

```go
// Package bodp implements the Branch Origin Decision Protocol.
// BODP는 새 브랜치 생성 직전 3개 시그널을 평가하여 권장 base 옵션을 산출한다.
// 3개 entry point에서 공유: /moai plan --branch (skill), /moai plan --worktree (skill),
// moai worktree new (CLI). 새 슬래시 명령어 또는 CLI 서브명령어를 도입하지 않는다.
package bodp

// Choice는 BODP 결정 카테고리를 나타낸다.
type Choice string

const (
    ChoiceMain     Choice = "main"     // origin/main 분기 (default 권장)
    ChoiceStacked  Choice = "stacked"  // 현재 브랜치에서 분기 (stacked PR)
    ChoiceContinue Choice = "continue" // 새 브랜치 생성 안 함, 현재 브랜치 계속 작업
)

// EntryPoint는 BODP를 호출한 entry point를 나타낸다.
type EntryPoint string

const (
    EntryPlanBranch    EntryPoint = "plan-branch"
    EntryPlanWorktree  EntryPoint = "plan-worktree"
    EntryWorktreeCLI   EntryPoint = "worktree-cli"
)

// CheckInput은 Check() 입력 구조체이다.
type CheckInput struct {
    CurrentBranch string // git rev-parse --abbrev-ref HEAD
    NewSpecID     string // 예: "SPEC-V3R3-CI-AUTONOMY-001"
    RepoRoot      string // git rev-parse --show-toplevel
    EntryPoint    EntryPoint
}

// BODPDecision은 Check() 결과 구조체이다.
type BODPDecision struct {
    SignalA     bool   // 코드 의존 (depends_on or diff path overlap)
    SignalB     bool   // working tree co-location (untracked SPEC-ID)
    SignalC     bool   // open PR with current as head
    Recommended Choice
    Rationale   string // 사용자 표시용 한국어 근거 (conversation_language=ko)
    BaseBranch  string // "origin/main" | "<CurrentBranch>" | ""
}

// Check는 3개 시그널을 평가하고 권장 옵션을 반환한다.
// git CLI shell-out (os/exec) + 파일 시스템 검사 (os.Stat).
// 외부 의존성 ZERO (third-party 모듈 미사용).
func Check(input CheckInput) (BODPDecision, error)
```

### 4.5 Public API (Go) — `internal/bodp/audit_trail.go`

```go
// AuditEntry는 .moai/branches/decisions/<branch>.md 파일에 기록되는 결정 항목이다.
type AuditEntry struct {
    Timestamp     time.Time
    EntryPoint    EntryPoint
    CurrentBranch string
    NewBranch     string // 사용자 confirm 후 생성된 브랜치 이름 (continue 케이스는 "")
    Decision      BODPDecision
    UserChoice    Choice // 사용자가 최종 선택한 옵션 (Recommended와 다를 수 있음)
    ExecutedCmd   string // 예: "git fetch origin main && git checkout -B feat/SPEC-XXX origin/main"
}

// WriteDecision은 .moai/branches/decisions/<branch-name>.md 에 결정 사항을 기록한다.
// 파일 형식: Markdown frontmatter + body (timestamp, signals, choice, command).
// branch-name 정규화: '/' → '-' 치환 (filesystem-safe).
// 부재 디렉토리 자동 생성 (mkdir -p .moai/branches/decisions/).
func WriteDecision(repoRoot string, entry AuditEntry) error

// HasAuditTrail은 .moai/branches/decisions/<branch>.md 파일 존재 여부를 반환한다.
// W7-T05 (moai status off-protocol detection) 에서 사용.
func HasAuditTrail(repoRoot, branchName string) bool
```

### 4.6 Dependencies (External Go Stdlib + zero third-party)

- `os/exec` — `git rev-parse`, `git status`, `git diff`, `git fetch`, `git checkout` 호출
- `os/exec` — `gh pr list --head <branch> --state open --json number` 호출 (Signal C)
- `os` — `os.Stat`, `os.MkdirAll`, `os.WriteFile` (audit trail 디렉토리/파일)
- `path/filepath` — 경로 조합
- `time` — `AuditEntry.Timestamp`
- `encoding/yaml` (gopkg.in/yaml.v3 via go.mod, already used elsewhere) — SPEC frontmatter `depends_on:` 파싱 (Signal A)
- `strings` — 경로 비교, 브랜치명 정규화

[HARD] No third-party HTTP / API libraries. CLAUDE.local.md §14 준수: const 추출 항목 — `auditTrailDir = ".moai/branches/decisions"`, `defaultBase = "origin/main"`, `envNoReminder = "MOAI_NO_BODP_REMINDER"`, recommendation rationale messages.

---

## 5. TDD Strategy per Sub-task

### W7-T01 — BODP Library (Pure-Go)

**RED — Test cases (test file only, no implementation):**

1. `TestRelatedness_AllNegative_RecommendsMain` — input: CurrentBranch="chore/translation-batch-b", NewSpecID="SPEC-V3R3-CI-AUTONOMY-001", working tree에 다른 SPEC untracked, open PR 없음 → expected: `BODPDecision{SignalA:F, SignalB:F, SignalC:F, Recommended:ChoiceMain, BaseBranch:"origin/main"}`. **AC-CIAUT-018 canonical replay (current session 시나리오)**.
2. `TestRelatedness_SignalA_RecommendsStacked` — input: CurrentBranch="feat/SPEC-AUTH-001-base", NewSpecID="SPEC-AUTH-002" with `depends_on: [SPEC-AUTH-001]` 명시 → expected: `BODPDecision{SignalA:T, SignalB:F, SignalC:F, Recommended:ChoiceStacked, BaseBranch:"feat/SPEC-AUTH-001-base"}`. **AC-CIAUT-019 canonical**.
3. `TestRelatedness_SignalB_RecommendsContinue` — input: CurrentBranch="feat/foo", NewSpecID="SPEC-FOO-001", working tree에 `.moai/specs/SPEC-FOO-001/` untracked → expected: `BODPDecision{SignalA:F, SignalB:T, SignalC:F, Recommended:ChoiceContinue, BaseBranch:""}`.
4. `TestRelatedness_SignalC_RecommendsStackedWithGotchaWarning` — input: CurrentBranch="feat/X", NewSpecID="SPEC-Y-001", `gh pr list --head feat/X` returns 1 open PR → expected: `Recommended:ChoiceStacked, BaseBranch:"feat/X", Rationale 에 "parent-merge gotcha" 또는 "§18.11" 포함`.
5. `TestRelatedness_SignalsAandC_RecommendsStacked` — input: 시그널 a + c 모두 positive → expected: `Recommended:ChoiceStacked` (decision matrix row 5 검증).

**GREEN — Minimal implementation:**

- `Check(input CheckInput) (BODPDecision, error)`:
  - Signal A 검사: `os/exec` 로 `git diff --name-only origin/main..HEAD` 실행 + SPEC frontmatter `depends_on:` 파싱 (yaml unmarshal). 매칭 시 SignalA=true.
  - Signal B 검사: `git status --porcelain` 출력 파싱; untracked entries 중 `.moai/specs/<NewSpecID>/` prefix 매칭.
  - Signal C 검사: `gh pr list --head <CurrentBranch> --state open --json number` 호출; 결과 array length ≥ 1 이면 SignalC=true. `gh` 미설치/미인증 시 graceful skip + log warning.
  - Decision matrix lookup: 위 §4.1 truth table 8 rows 그대로 if-else chain.
  - Rationale 메시지: const map `rationaleMessages map[Choice]string` 에서 lookup.

**REFACTOR concerns:**

- `os/exec` 호출 wrapping: 테스트 가능성 위해 `gitCommand func(args ...string) (string, error)` 와 `ghCommand func(args ...string) (string, error)` 변수 export → 테스트에서 fake 함수 주입 (의존성 역전). production 에서는 default `exec.Command` wrapper.
- SPEC frontmatter 파싱: `internal/spec/` 패키지가 이미 존재하면 재사용 (다른 Wave 산출물 검토 필요; 부재 시 inline `gopkg.in/yaml.v3` 호출).
- Rationale 메시지 i18n: Wave 7은 conversation_language=ko 단일 언어; 향후 follow-up SPEC에서 i18n 라이브러리화 가능 (현 시점 over-engineering 회피).

### W7-T02 — Skill Body Phase 3 확장

**RED — Test cases:**

본 task는 skill body markdown 확장으로 직접 단위 테스트 부재. 대신 **integration verification**:

1. `internal/cli/plan_branch_integration_test.go` (or shell-based harness): mock `manager-git` invocation 상태에서 `/moai plan --branch` 흐름 시뮬레이션 → BODP `Check()` 호출 → AskUserQuestion mock 응답 → manager-git에 `base=<chosen>` parameter 포함 verify.
2. `internal/cli/plan_worktree_integration_test.go`: 위와 유사하게 `/moai plan --worktree` 흐름에서 `moai worktree new <SPEC-ID> --base <chosen>` 호출 검증.

(Skill body 자체는 markdown — 테스트는 orchestrator behavior에 대한 검증으로 한정. plan-auditor 가 skill body 텍스트를 schema verify.)

**GREEN — Minimal implementation:**

- `.claude/skills/moai/workflows/plan.md` Phase 3 Branch Path 섹션:
  - manager-git 위임 직전 단계 추가: "Step 3.X: BODP relatedness check"
    - `internal/bodp/relatedness.go` `Check()` 호출 (orchestrator 측에서 직접 또는 helper script 경유)
    - `BODPDecision` 결과를 AskUserQuestion options 로 변환:
      - 첫 옵션 = `Recommended` 카테고리 (`(권장)` 라벨 부착)
      - 추가 옵션 = 다른 `Choice` enum 값 (사용자 over-ride 가능)
      - "Other" 옵션 자동 (Claude Code 기본 동작)
    - 사용자 confirm 후 manager-git 호출 시 `base=<chosen>` parameter 명시 전달
- `.claude/skills/moai/workflows/plan.md` Phase 3 Worktree Path 섹션:
  - 동일 BODP 검사 + AskUserQuestion + 결과를 `moai worktree new <SPEC-ID> --base <chosen>` CLI 호출 시 flag 로 전달.
- 두 path 모두 W7-T04 audit trail writer 호출 (orchestrator 측에서 `internal/bodp/audit_trail.go` `WriteDecision()` 호출).

**REFACTOR concerns:**

- skill body 텍스트 중복 회피: Branch Path와 Worktree Path 의 BODP 검사 부분은 공통 sub-section ("§Phase 3.X: BODP Gate (공통)") 으로 추출하여 두 path 모두 참조.
- AskUserQuestion option label 표준화: spec.md §3.8 REQ-CIAUT-045/046/047/047b의 한국어 문구를 verbatim 사용 (예: "main에서 분기 (권장)", "현재 브랜치에서 분기 (stacked PR)", "현재 브랜치에 계속 작업").

### W7-T03 — `moai worktree new` CLI 확장

**RED — Test cases:**

1. `TestNew_DefaultBaseIsOriginMain` — input: `moai worktree new SPEC-FOO-001` (no flag) → expected: `git worktree add ... origin/main` 호출 + audit trail 기록 (entry point: `worktree-cli`).
2. `TestNew_FromCurrentFlagPreservesOldBehavior` — input: `moai worktree new SPEC-FOO-001 --from-current` → expected: `git worktree add ... <currentHEAD>` 호출 (기존 동작 opt-out 유지).
3. `TestNew_BaseFlagSpecifiesBranch` — input: `moai worktree new SPEC-FOO-001 --base feat/SPEC-AUTH-001-base` → expected: `git worktree add ... feat/SPEC-AUTH-001-base` 호출 (W7-T02 skill 에서 `--base` flag 전달 사용).
4. `TestNew_NoAskUserQuestion` — input: 임의 invocation; mock AskUserQuestion API → expected: 호출 횟수 0 (orchestrator-only HARD 준수 verify).
5. `TestNew_BaseAndFromCurrentMutuallyExclusive` — input: `moai worktree new SPEC-FOO-001 --base X --from-current` → expected: error "flags are mutually exclusive".
6. `TestNew_AuditTrailWritten` — `internal/bodp/audit_trail.go` mock; verify `WriteDecision()` 호출 with `EntryPoint: EntryWorktreeCLI`.

**GREEN — Minimal implementation:**

- `internal/cli/worktree/new.go` 확장:
  - 기존 `cobra.Command` 또는 동등 CLI 구조에 `--base <branch>` (string flag, default `origin/main`) + `--from-current` (bool flag, default false) 추가.
  - Mutual exclusion 검증: 두 flag 동시 지정 시 error.
  - Base 결정 로직:
    - `--from-current` 지정 → use current HEAD
    - else `--base` 지정 → use specified branch (skill path 에서 W7-T02 가 사용)
    - else → use `origin/main` (default; CLI direct path)
  - `git fetch origin main` 사전 실행 (default base 케이스만; `--from-current` 또는 explicit `--base` 는 skip).
  - `git worktree add <path> <base>` 호출.
  - 호출 후 `internal/bodp/audit_trail.go` `WriteDecision()` with `EntryPoint: EntryWorktreeCLI` + 시그널 검사 결과 기록 (CLI path 는 AskUserQuestion 미호출 → 시그널만 기록, UserChoice = Recommended).

**REFACTOR concerns:**

- CLI flag naming: `--base` vs `--from`: spec.md §3.8 + plan.md §9 W7-T03 표기 `--base` 사용 (skill `--worktree` path 와 일관).
- AskUserQuestion 미호출 enforcement: 코드 리뷰에서 `internal/cli/worktree/` 패키지 import 목록에 `AskUserQuestion` 패키지 부재 verify (정적 검사). 추가로 lint custom rule 가능 (over-engineering 회피).

### W7-T04 — Audit Trail Writer

**RED — Test cases:**

1. `TestWriteDecision_CreatesFile` — input: `AuditEntry{Timestamp:t, EntryPoint:EntryPlanBranch, CurrentBranch:"chore/x", NewBranch:"feat/SPEC-Y", Decision:..., UserChoice:ChoiceMain, ExecutedCmd:"..."}` → expected: `.moai/branches/decisions/feat-SPEC-Y.md` 생성, frontmatter + body 검증.
2. `TestWriteDecision_NormalizesBranchNameForFilename` — input: NewBranch="feat/SPEC-A/sub-feature" → expected: 파일명 `feat-SPEC-A-sub-feature.md` (slash → dash).
3. `TestWriteDecision_CreatesDirectoryIfAbsent` — input: `.moai/branches/decisions/` 부재 상태 → expected: `os.MkdirAll` 호출 + 파일 생성 성공.
4. `TestHasAuditTrail_DetectsExisting` — pre-existing `.moai/branches/decisions/feat-X.md` → expected: `HasAuditTrail(repo, "feat/X")` returns true.
5. `TestHasAuditTrail_AbsentBranch` — pre-existing 파일 없음 → expected: `HasAuditTrail` returns false.
6. `TestHasAuditTrail_DirAbsentReturnsFalse` — `.moai/branches/decisions/` 자체 부재 → expected: `HasAuditTrail` returns false (no error, no false positive — W7-T05 에서 신규 프로젝트 false-positive 방지에 의존).

**GREEN — Minimal implementation:**

- `WriteDecision`:
  - 파일 경로: `filepath.Join(repoRoot, auditTrailDir, normalizeBranchName(entry.NewBranch) + ".md")`
  - `os.MkdirAll(filepath.Dir(path), 0o755)` 사전 실행.
  - Markdown 형식:
    ```markdown
    ---
    timestamp: 2026-05-09T12:34:56Z
    entry_point: plan-branch
    current_branch: chore/translation-batch-b
    new_branch: feat/SPEC-V3R3-CI-AUTONOMY-001
    user_choice: main
    ---
    # BODP Decision: feat/SPEC-V3R3-CI-AUTONOMY-001

    ## Signals
    - Signal (a) — Code dependency: false (no diff overlap, no depends_on match)
    - Signal (b) — Working tree co-location: false (untracked SPEC-MX-INJECT-001 ≠ this SPEC ID)
    - Signal (c) — Open PR head: false (no open PR with chore/translation-batch-b as head)

    ## Decision
    - Recommended: main
    - User choice: main
    - Base branch: origin/main
    - Rationale: 현재 브랜치와 무관한 새 작업이므로 main 분기를 권장합니다.

    ## Executed
    ```
    git fetch origin main && git checkout -B feat/SPEC-V3R3-CI-AUTONOMY-001 origin/main
    ```
    ```
  - `os.WriteFile(path, content, 0o644)` 호출.
- `HasAuditTrail`: `os.Stat(filepath.Join(repoRoot, auditTrailDir, normalizeBranchName(branchName) + ".md"))` 결과로 boolean 반환. error 시 false (false-positive 방지).

**REFACTOR concerns:**

- 파일 형식 stability: 향후 audit trail 분석 도구 (예: `moai branch decisions list`) 가능성 — frontmatter 구조 명확화. 단 Wave 7은 writer + reader (HasAuditTrail) 만 ship.
- Concurrent write safety: orchestrator 단일 세션 가정, 동시 write 시나리오는 Wave 7 scope 외.

### W7-T05 — `moai status` Off-Protocol Reminder

**RED — Test cases:**

1. `TestStatus_OffProtocolBranchReminder` — pre-existing branch `feat/quick-fix` 가 raw `git checkout -b` 로 생성됨 (`.moai/branches/decisions/feat-quick-fix.md` 부재) → expected: `moai status` stdout/stderr 에 "off-protocol branch detected" 포함된 friendly reminder 출력. exit code 0 (block 안 함).
2. `TestStatus_AuditTrailExistsNoReminder` — pre-existing branch `feat/SPEC-X` + `.moai/branches/decisions/feat-SPEC-X.md` 존재 → expected: reminder 출력 안 함.
3. `TestStatus_AuditTrailDirAbsentNoFalsePositive` — `.moai/branches/decisions/` 디렉토리 자체 부재 (신규 프로젝트) → expected: reminder 출력 안 함 (false-positive 방지 — acceptance.md AC-CIAUT-024 Failure Mode).
4. `TestStatus_EnvVarDisablesReminder` — `MOAI_NO_BODP_REMINDER=1` 환경변수 설정 + audit trail 부재 → expected: reminder 출력 안 함.
5. `TestStatus_MainBranchNoReminder` — current branch 가 `main` → expected: reminder 출력 안 함 (main 자체는 BODP 대상 아님).

**GREEN — Minimal implementation:**

- `internal/cli/status.go` 확장:
  - 기존 status 출력 끝에 BODP off-protocol check section 추가.
  - 흐름:
    1. `os.Getenv("MOAI_NO_BODP_REMINDER") == "1"` → return early
    2. `git rev-parse --abbrev-ref HEAD` → currentBranch
    3. currentBranch == "main" 또는 "master" → return early
    4. `internal/bodp/audit_trail.go` `HasAuditTrail(repoRoot, currentBranch)` 호출
       - true → return early
       - false → audit trail 디렉토리 자체 부재인지 추가 검사 (`os.Stat(.moai/branches/decisions/)`)
         - 부재 → return early (false-positive 방지)
         - 존재 → reminder 출력
  - Reminder 메시지 (verbatim, conversation_language=ko):
    ```
    ⚠️  Branch `<currentBranch>` was created without going through MoAI entry points.
    Future branches: use `/moai plan --branch <name>` (SPEC-tied) or `moai worktree new <SPEC-ID>` for relatedness check + audit trail.
    Skip with `MOAI_NO_BODP_REMINDER=1` if intentional.
    ```

**REFACTOR concerns:**

- Reminder 메시지의 i18n: 본 Wave는 ko 단일 언어; 향후 i18n SPEC 에서 메시지 i18n 처리 (현 시점 const string).
- `moai status` 의 다른 출력과 시각 분리: ⚠️ 이모지 + blank line separator.

### W7-T06 — 문서화 + Template-First Mirror

**RED — Test cases:**

본 task는 markdown 문서 작업으로 직접 단위 테스트 부재. 대신 **structural verification**:

1. `TestDocsCLAUDELocalMd_Section1812Exists` — `grep '## 18.12\|### §18.12' CLAUDE.local.md` 결과 1+ 매칭.
2. `TestDocsCLAUDELocalMd_TemplateMirror` — `internal/template/templates/CLAUDE.local.md` 에도 §18.12 존재 (Template-First 준수).
3. `TestDocsRule_BranchOriginProtocolExists` — `.claude/rules/moai/development/branch-origin-protocol.md` 파일 존재 + frontmatter `paths:` 필드 명시.
4. `TestDocsRule_TemplateMirror` — `internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md` 동일 내용 mirror.

(Phase 1.5 plan-auditor 가 markdown schema 검증; Phase 2 manager-tdd 가 위 4개 테스트 stub 작성.)

**GREEN — Minimal implementation:**

- `CLAUDE.local.md` §18.11 다음에 §18.12 신규 subsection 추가:
  ```markdown
  ### §18.12 [HARD] Branch Origin Decision Protocol (BODP)

  새 브랜치 생성 시 base 결정을 자동화. **새 슬래시 명령어 또는 CLI 서브명령어 ZERO** 원칙. 기존 3개 entry point에 동작 내장.

  #### 3-Signal Algorithm
  - Signal (a) — 코드 의존: SPEC depends_on 또는 현재 브랜치 diff 경로 overlap
  - Signal (b) — Working tree co-location: untracked `.moai/specs/<SPEC-ID>/` 매칭
  - Signal (c) — Open PR head: `gh pr list --head <current> --state open` ≥ 1

  #### Decision Matrix
  - 모든 시그널 negative → `main에서 분기` 권장 (default; `origin/main`)
  - Signal (a) positive → `stacked PR` 권장 (base = 현재 브랜치)
  - Signal (b) positive → `현재 브랜치에 계속 작업` 권장 (no new branch)
  - Signal (c) positive → `stacked PR` 권장 + parent-merge gotcha 경고 (§18.11 Case Study)

  #### 3 Invocation Paths (모두 기존 entry, 새 명령어 ZERO)
  1. `/moai plan --branch [name]` — SPEC-tied feature 브랜치 (skill Phase 3 Branch Path)
  2. `/moai plan --worktree` — SPEC-tied worktree (skill Phase 3 Worktree Path; `moai worktree new <SPEC-ID> --base <chosen>` 호출)
  3. `moai worktree new <SPEC-ID>` — 직접 worktree CLI (default base = `origin/main`; `--from-current` flag로 opt-out)

  #### Audit Trail
  - 위치: `.moai/branches/decisions/<branch-name>.md` (slash → dash 정규화)
  - 형식: Markdown frontmatter (timestamp, entry_point, current_branch, new_branch, user_choice) + body (signals/decision/executed command)
  - 1 entry per branch

  #### Off-Protocol Reminder
  - raw `git checkout -b` 우회 시 BODP 무관, 다음 `moai status` 호출에서 친절한 reminder 출력 (block 안 함)
  - `MOAI_NO_BODP_REMINDER=1` 환경변수로 비활성화 가능
  - main/master 브랜치 또는 audit trail 디렉토리 부재 (신규 프로젝트) 시 reminder 출력 안 함

  #### Out of Scope
  - raw `git checkout -b` 가로채기 — opt-in 유지 (REQ-CIAUT-050 + spec.md §7 OQ4)
  - 기존 off-protocol 브랜치 자동 마이그레이션 — reminder만 출력
  - Multi-repo 일괄 적용 — 단일 repo scope
  ```
- `internal/template/templates/CLAUDE.local.md` 동일 §18.12 mirror.
- `.claude/rules/moai/development/branch-origin-protocol.md` 신규 rule 파일:
  ```markdown
  ---
  paths:
    - "internal/cli/worktree/**/*.go"
    - "internal/cli/status.go"
    - "internal/bodp/**/*.go"
    - ".claude/skills/moai/workflows/plan.md"
  ---

  # Branch Origin Decision Protocol (BODP)

  See CLAUDE.local.md §18.12 for full algorithm + entry point spec.
  This rule auto-loads when modifying BODP-related code paths.

  ## HARD Rules
  - [HARD] BODP introduces ZERO new slash commands or CLI subcommands.
  - [HARD] CLI path (moai worktree new) MUST NOT invoke AskUserQuestion (orchestrator-only HARD).
  - [HARD] Audit trail MUST be written for every BODP decision (3 entry points).
  - [HARD] Off-protocol reminder MUST NOT block (`exit 0`).

  ## Invocation Paths
  - Skill paths (`/moai plan --branch`, `/moai plan --worktree`): orchestrator runs BODP, AskUserQuestion gate, manager-git delegation.
  - CLI path (`moai worktree new`): default base = origin/main, `--from-current` opt-out, no AskUserQuestion.

  ## Cross-References
  - SPEC: SPEC-V3R3-CI-AUTONOMY-001 (Wave 7, T8)
  - Acceptance: AC-CIAUT-018, 019, 019b, 024, 025
  - Library: `internal/bodp/relatedness.go`, `internal/bodp/audit_trail.go`
  ```
- `internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md` 동일 내용 Template-First mirror.

**REFACTOR concerns:**

- Mirror 일관성: `make build` 후 embedded.go 재생성 검증 (CLAUDE.local.md §2 Template-First).
- §18.12 위치: §18.11 (v2.14 Case Study) 다음, §19 (AskUserQuestion Enforcement) 이전 — natural numbering.

---

## 6. Open Questions (Resolved Inline)

> **Namespace note**: Wave7-Q1..Q3는 본 Wave 7 strategy local namespace. spec.md §7 OQ1..OQ4 (이미 resolved) 와 무관.

### Wave7-Q1: BODP Library를 `internal/bodp/` vs `pkg/bodp/` 어디에 둘 것인가?

**Decision**: `internal/bodp/`.

**Rationale**: moai-adk-go 의 컨벤션 (CLAUDE.local.md 참조) — 외부 import 가능한 public API 는 `pkg/`, dev project 내부 사용 전용은 `internal/`. BODP 는 (a) 3개 entry point 모두 dev project 내부 호출, (b) 외부 라이브러리 사용자 노출 의도 없음 → `internal/` 적합. 향후 외부 SDK 화 필요 시 follow-up SPEC.

### Wave7-Q2: Signal A 검사 — SPEC frontmatter `depends_on:` 만 vs `depends_on:` + diff path overlap?

**Decision**: BOTH.

**Rationale**: spec.md §3.8 REQ-CIAUT-044 line 238 verbatim "SPEC `depends_on:` field referencing an in-progress SPEC on this branch; **file-path overlap with current branch's diff vs main**". 둘 중 하나만 positive 여도 SignalA=true. depends_on 미선언이지만 동일 패키지 작업하는 케이스 (예: 같은 internal/cli/ 하위 SPEC 두 개) 까지 포착 필요. plan-auditor 가 spec.md verbatim 준수 검증.

### Wave7-Q3: AskUserQuestion options 순서 — Recommended를 항상 첫 번째? 아니면 사용자 시그널에 따라 동적 정렬?

**Decision**: Recommended 항상 첫 번째, `(권장)` 라벨 부착 (askuser-protocol.md §Socratic Interview Structure 준수).

**Rationale**: askuser-protocol.md `First option label: MUST carry the (권장) suffix` HARD 규칙. 시그널 결과는 Recommended 카테고리만 변경 (main/stacked/continue 중 하나가 첫 번째로 회전); 옵션 라벨/설명은 시그널별 메시지 (REQ-045/046/047/047b verbatim) 사용. "Other" 옵션 자동.

---

## 7. Constraints Compliance

Reference: spec.md §4.1 Hard Constraints + CLAUDE.local.md §2 (Template-First) + §14 (no hardcoding) + askuser-protocol.md (AskUserQuestion HARD).

| Constraint | Wave 7 Compliance |
|------------|-------------------|
| **Template-First** (CLAUDE.local.md §2) | W7-T06 신규 rule 파일 (`.claude/rules/moai/development/branch-origin-protocol.md`) 과 CLAUDE.local.md §18.12 변경이 `internal/template/templates/` 에 mirror됨. `internal/bodp/`, `internal/cli/worktree/new.go`, `internal/cli/status.go`, `.claude/skills/moai/workflows/plan.md` 는 dev-project Go 코드 또는 skill body — Template-First 적용 대상 아님 (Wave 6 §7 동일 패턴, Wave 4 `.github/workflows/optional/` 선례). [HARD] verify-via-grep: `ls internal/template/templates/internal/bodp/` 빈 결과 + `grep '§18.12' internal/template/templates/CLAUDE.local.md` 매칭 + `ls internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md` 존재. |
| **16-language neutrality** (CLAUDE.local.md §15) | Wave 7 산출물 (`internal/bodp/`, `moai worktree new` CLI, `moai status` CLI, BODP rule, CLAUDE.local.md §18.12) 은 **언어 중립**. SPEC frontmatter 파싱 (Signal A) + audit trail 형식은 markdown 기반 — 프로젝트 구현 언어와 무관. 16개 언어 모두 동등 사용 가능. |
| **No hardcoding** (CLAUDE.local.md §14) | 모든 path/branch-name/env-var/rationale 메시지를 const 추출: `auditTrailDir = ".moai/branches/decisions"`, `defaultBase = "origin/main"`, `envNoReminder = "MOAI_NO_BODP_REMINDER"`, `mainBranches = []string{"main", "master"}`, `rationaleMessages = map[Choice]string{...}`. CLI flag default values 도 const. |
| **AskUserQuestion HARD** (askuser-protocol.md, agent-common-protocol.md §User Interaction Boundary) | (a) Skill path (W7-T02): orchestrator AskUserQuestion 호출 + `(권장)` 라벨 첫 옵션 + `Other` 자동 + conversation_language=ko 준수. (b) CLI path (W7-T03 `moai worktree new`): AskUserQuestion 호출 절대 금지 (orchestrator-only HARD); flag-based decision only. 정적 검사: `internal/cli/worktree/` import에서 AskUserQuestion 패키지 부재. |
| **No release/tag automation** (feedback_release_no_autoexec.md) | Wave 7 BODP 는 git tag, gh release, goreleaser 호출 ZERO. branch 생성 + audit trail 작성만. |
| **Conventional Commits + 🗿 MoAI co-author** | All Wave 7 commits follow this format (see §10 Commit Cadence). |
| **AskUserQuestion Channel Monopoly** (CLAUDE.md §1, askuser-protocol.md §Channel Monopoly) | W7-T02 skill body 의 모든 사용자 결정 요청은 AskUserQuestion tool call 경유; 산문 질문 prohibited. W7-T05 `moai status` reminder 는 statement (not question), AskUserQuestion 불필요. |

---

## 8. File Ownership

### Implementer Scope (write access)

```
internal/bodp/relatedness.go                                              # NEW (W7-T01)
internal/bodp/audit_trail.go                                              # NEW (W7-T04)
internal/cli/worktree/new.go                                              # EXTEND (W7-T03: --base, --from-current flags + audit trail call)
internal/cli/status.go                                                    # EXTEND (W7-T05: off-protocol reminder)
.claude/skills/moai/workflows/plan.md                                     # EXTEND (W7-T02: Phase 3 Branch + Worktree Path BODP gate)
.claude/rules/moai/development/branch-origin-protocol.md                  # NEW (W7-T06)
internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md  # NEW (W7-T06 mirror)
internal/template/templates/CLAUDE.local.md                               # EXTEND (W7-T06: §18.12 mirror)
CLAUDE.local.md                                                           # EXTEND (W7-T06: §18.12 신규 subsection)
.moai/branches/decisions/.gitkeep                                         # NEW (W7-T04: 디렉토리 보존)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave7.md                   # this file
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave7.md                      # companion
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/progress.md                         # extend (Phase 1 entry; Wave 7 종료 entry)
```

### Tester Scope (test files only)

```
internal/bodp/relatedness_test.go                                         # NEW (W7-T01: 테이블 기반 4-5 case)
internal/bodp/audit_trail_test.go                                         # NEW (W7-T04: write/read/normalize/missing-dir cases)
internal/cli/worktree/new_test.go                                         # EXTEND (W7-T03: 6 case 포함)
internal/cli/status_test.go                                               # EXTEND (W7-T05: 5 case 포함)
internal/cli/plan_branch_integration_test.go                              # OPTIONAL (W7-T02: integration verification; sub-agent 직접 구현 fallback 가능)
internal/cli/plan_worktree_integration_test.go                            # OPTIONAL (W7-T02)
```

### Read-Only Scope

```
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/spec.md                             # frozen for Wave 7 (§3.8 REQ-042..051 source)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/plan.md                             # frozen for Wave 7 (§9 Wave 7 outline + §10.5 BODP Embedded pattern)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/acceptance.md                       # frozen for Wave 7 (AC-018/019/019b/024/025 source)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave[1-6].md               # prior waves audit trail
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave[1-6].md                  # prior waves audit trail
internal/template/templates/CLAUDE.local.md                               # before W7-T06 baseline (read for diff context)
.claude/rules/moai/core/askuser-protocol.md                               # AskUserQuestion HARD (W7-T02 source of constraint)
.claude/rules/moai/core/agent-common-protocol.md                          # User Interaction Boundary (W7-T03 CLI no-AskUser HARD source)
CLAUDE.md, CLAUDE.local.md                                                # project rules (§14 no-hardcoding, §15 16-language, §18 GitHub Flow)
```

### Mode

[HARD] Solo mode (single-implementer, `--branch` pattern; lessons #13). Wave 7 scope 7 source files (2 Go production new + 2 Go production extend + 2 markdown extend + 1 markdown new + 1 mirror set + 1 .gitkeep) + 4-6 Go test files — `--team` parallelism 가치 < 협업 오버헤드. lessons #12 (수동 worktree isolation 위배) 재발 방지 위해 `Agent(isolation: "worktree")` 위임 권장 (sub-agent → worktree 격리), 단 sub-agent 1M context error fallback 시 main-session 직접 구현 (Wave 4/5/6 동일 패턴, Wave 6 §C-7 lesson).

---

## 9. Acceptance Criteria Mapping (Detailed)

| Wave 7 Task | Drives AC | Validation Method |
|-------------|-----------|-------------------|
| W7-T01 (BODP library) | AC-CIAUT-018 (transitive: all-negative → main), AC-CIAUT-019 (transitive: signal A → stacked) | 테이블 기반 단위 테스트 5 case (current session replay + signal-a + signal-b + signal-c + a+c). `t.TempDir()` 격리 + git CLI mock injection. |
| W7-T02 (skill Phase 3 BODP gate) | AC-CIAUT-018 (canonical: `/moai plan --branch` flow + manager-git base parameter), AC-CIAUT-019 (canonical: `/moai plan --branch --resume` with `depends_on`) | Integration verification (Phase 2 단계). skill body markdown은 plan-auditor schema verify (`(권장)` 라벨, conversation_language=ko, AskUserQuestion options ≤4). |
| W7-T03 (`moai worktree new` CLI) | AC-CIAUT-019b (default origin/main + `--from-current` opt-out + no AskUserQuestion) | `internal/cli/worktree/new_test.go` 6 case. AskUserQuestion absence 정적 검사. |
| W7-T04 (audit trail writer) | AC-CIAUT-018 (audit trail 파일 생성), AC-CIAUT-019 (stacked rationale 기록), AC-CIAUT-019b (entry point: worktree-cli 기록) | `internal/bodp/audit_trail_test.go` 6 case. 파일 형식 + 디렉토리 자동 생성 + branch name normalization. |
| W7-T05 (`moai status` reminder) | AC-CIAUT-024 (off-protocol reminder + false-positive 방지 + env var 비활성화) | `internal/cli/status_test.go` 5 case. `t.TempDir()` 격리 + `t.Setenv("MOAI_NO_BODP_REMINDER", "1")`. |
| W7-T06 (CLAUDE.local.md §18.12 + rule + Template-First mirror) | AC-CIAUT-025 (§18.12 존재 + 3 entry point 명시 + audit trail 위치 + opt-out 방법) | `grep -A 30 '§18.12' CLAUDE.local.md` + `grep '§18.12' internal/template/templates/CLAUDE.local.md` + `ls .claude/rules/moai/development/branch-origin-protocol.md` + Template-First mirror diff. |

> **Note for spec.md/acceptance.md follow-up (no edits in this Wave 7)**:
> - REQ-CIAUT-051 (line 259-260 of spec.md) is present and binds to W7-T06. ✓
> - AC-CIAUT-018, 019, 019b, 024, 025 (lines 358-440 of acceptance.md) are present and bind to W7-T01..T06. ✓
> - **No gaps detected**. spec.md §3.8 + acceptance.md §9 mappings are fully consistent with plan.md §9 W7-T01..T06 scope. plan-auditor verdict expected: PASS (no must-pass criterion failure).

---

## 10. Commit Cadence

12 commits, Conventional Commits format. Every commit body MUST end with the verbatim trailer (separated by a blank line):

```
🗿 MoAI <email@mo.ai.kr>
```

1. `test(bodp): W7-T01 RED — relatedness check cases (all-negative, signal-a, signal-b, signal-c, a+c)`
2. `feat(bodp): W7-T01 implement Check() + 3-signal evaluation + decision matrix`
3. `test(bodp): W7-T04 RED — audit trail writer cases (create, normalize, mkdir, has-trail, missing-dir)`
4. `feat(bodp): W7-T04 implement WriteDecision + HasAuditTrail`
5. `test(cli/worktree): W7-T03 RED — new --base / --from-current / no-AskUser cases`
6. `feat(cli/worktree): W7-T03 add --base + --from-current flags + audit trail call`
7. `test(cli/status): W7-T05 RED — off-protocol reminder + false-positive + env-var cases`
8. `feat(cli/status): W7-T05 implement off-protocol branch reminder`
9. `feat(skill/plan): W7-T02 extend Phase 3 Branch + Worktree Path with BODP gate`
10. `docs(rules): W7-T06 add branch-origin-protocol.md rule + CLAUDE.local.md §18.12`
11. `chore(template): W7-T06 mirror branch-origin-protocol.md + CLAUDE.local.md §18.12 to templates/`
12. `chore(spec): SPEC-V3R3-CI-AUTONOMY-001 Wave 7 progress.md Phase 1 + closure entries`

Optional REFACTOR commits between feat steps if shared `os/exec` wrapping (W7-T01) requires extraction or if W7-T02 markdown duplication elimination (Branch + Worktree common sub-section) becomes large.

---

## 11. Risk Register

| ID | Risk | Mitigation |
|----|------|-----------|
| W7-R1 | False-positive Signal C: `gh pr list` 호출 실패 (gh 미설치/미인증) → SignalC=false 강제, 사용자가 stacked 권장을 못 받음. | `gh` 호출 실패 시 stderr warning + SignalC=false (graceful skip). 사용자가 수동으로 "stacked" 옵션 선택 가능 (AskUserQuestion 옵션 항상 모두 표시). 추가로 `gh auth status` 사전 검사 W7-T01에 포함 가능 (over-engineering 회피하고자 Wave 7는 graceful skip만). |
| W7-R2 | False-negative Signal A: depends_on 미선언이지만 다른 SPEC 변경과 path overlap. diff path overlap heuristic 부정확. | 시그널 둘 다 검사 (Wave7-Q2 결정). depends_on 만 매칭하지 않으면 path overlap fallback. 단 path overlap heuristic 은 internal/cli/ 같은 일반 디렉토리에서 false-positive 가능 → minimum overlap threshold (예: 2+ 파일) 적용. plan-auditor 가 verbatim spec.md 준수 verify. |
| W7-R3 | AskUserQuestion option 순서 혼란: signal 결과에 따라 첫 옵션이 변하면 사용자가 단축키 학습 어려움. | Wave7-Q3 결정 따라 첫 옵션이 항상 `(권장)` 라벨 부착 — 사용자는 라벨로 권장 옵션 식별. CLAUDE.local.md §19 AskUserQuestion Enforcement 준수. 옵션 4개 한도 내 회전 (main/stacked/continue/Other). |
| W7-R4 | CLI path (`moai worktree new`) AskUserQuestion 누설: 코드 변경 중 실수로 AskUserQuestion 호출 도입. | (a) `internal/cli/worktree/new.go` import에 AskUserQuestion 패키지 부재 정적 검사 (W7-T03 단위 테스트 case `TestNew_NoAskUserQuestion`). (b) 코드 리뷰 체크리스트 항목 추가 (PR description 명시). (c) Future: lint custom rule (Wave 7 scope 외). |
| W7-R5 | Audit trail concurrent write race condition: 두 orchestrator 세션이 동시에 같은 branch에 BODP 실행 → file overwrite race. | Wave 7 scope: orchestrator 단일 세션 가정. 동시 write는 follow-up SPEC 또는 file lock 라이브러리 (`internal/lockedfile/` 패턴; lessons #11 referenced) 도입. 현 시점 risk 수용. |
| W7-R6 | `moai status` reminder 빈도 과다: 사용자가 매 status 호출마다 reminder 보면 noise. | (a) `MOAI_NO_BODP_REMINDER=1` 환경변수 비활성화 (acceptance.md AC-CIAUT-024 명시). (b) reminder 1회만 출력 후 `.moai/state/bodp-reminder-shown-<branch>` flag 추가 검토 (over-engineering 회피, Wave 7 ship 안 함). (c) 신규 프로젝트 (audit trail 디렉토리 자체 부재) false-positive 방지 명시 처리. |
| W7-R7 | spec.md §3.8 `gh pr list` 호출이 audit trail 작성 전에 실패해서 audit trail 파일 영영 생성 안 됨. | W7-T01 graceful skip (W7-R1) + W7-T04 WriteDecision 은 SignalC 결과와 무관하게 audit trail 작성 (SignalC=false 로 기록). 사용자가 audit trail 파일을 수동 검토 시 SignalC=false + rationale 에 "gh unavailable" 명시 가능. |

---

## 12. Out of Wave 7 Scope

- Raw `git checkout -b` 가로채기 (REQ-CIAUT-050 명시 — opt-in 유지). `moai status` reminder만 출력.
- Multi-repo 일괄 BODP 적용 — 단일 repo scope.
- Branch deletion 자동화 — Wave 7는 branch 생성만.
- Audit trail concurrent write file lock — orchestrator 단일 세션 가정.
- Audit trail 파일 분석 도구 (예: `moai branch decisions list`) — follow-up SPEC.
- `moai status` reminder 1회 표시 후 자동 mute — over-engineering 회피.
- BODP를 `/moai run` Phase 2에 plumb — 구현 단계는 base 결정 이후, 본 Wave scope 외.
- Signal A path overlap heuristic 의 정밀화 (예: ML 기반 코드 의존 분석) — Wave 7는 directory prefix 기반 단순 heuristic.
- BODP rationale 메시지 i18n (en/ja/zh) — conversation_language=ko 단일.

---

## 13. Configuration Precedence

`moai worktree new` CLI 의 base 결정 우선순위 (높음 → 낮음):

1. **`--from-current` flag** (explicit opt-out): 현재 HEAD 사용. `--base` 와 mutual exclusive (W7-T03 `TestNew_BaseAndFromCurrentMutuallyExclusive` 검증).
2. **`--base <branch>` flag** (explicit base): 지정 브랜치 사용. skill path (W7-T02) 가 BODP 결과를 이 flag 로 전달.
3. **Default (no flag)**: `origin/main` 사용. `git fetch origin main` 사전 실행.

`/moai plan --branch` / `/moai plan --worktree` skill path 의 base 결정 우선순위:

1. **사용자 AskUserQuestion 응답** (slash path 항상 prompt): 응답된 옵션 사용.
2. **BODP Recommended (자동 첫 옵션)**: 사용자가 첫 옵션을 그대로 선택하면 Recommended 사용.
3. **Default (무응답 시 첫 옵션)**: AskUserQuestion이 응답 없으면 (timeout 가정) 첫 옵션 자동 선택 (방어적; 실제 운영에서는 사용자 응답 대기).

`moai status` reminder 출력 우선순위 (W7-T05):

1. **`MOAI_NO_BODP_REMINDER=1` 환경변수**: reminder 출력 안 함 (highest priority).
2. **Branch가 main/master**: reminder 출력 안 함.
3. **Audit trail 디렉토리 부재 (신규 프로젝트)**: reminder 출력 안 함 (false-positive 방지).
4. **Audit trail 파일 존재**: reminder 출력 안 함 (BODP 경로 사용 확인).
5. **Above 모두 false**: reminder 출력 (off-protocol 의심).

---

Version: 0.1.0
Status: pending plan-auditor audit (Phase 1.5)
Last Updated: 2026-05-09
