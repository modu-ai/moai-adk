---
spec: SPEC-V3R3-CI-AUTONOMY-001
wave: 7
version: 0.1.0
status: draft
created_at: 2026-05-09
updated_at: 2026-05-09
author: manager-spec
---

# Wave 7 Atomic Tasks (Phase 1 Output)

> Companion to strategy-wave7.md. SPEC-V3R3-CI-AUTONOMY-001 Wave 7 — T8 Branch Origin Decision Protocol (BODP).
> Generated: 2026-05-09. Methodology: TDD (Pure-Go library + table-driven test cases + skill body extension + CLI flag extension).
> Wave Base: `origin/main 78929d058` (post-Wave 6 PR #793 머지 baseline).
> Final Wave (7/7): SPEC closure 직전 단계.

---

## Task Inventory Summary

| Total Tasks | RED Phases | GREEN Phases | REFACTOR Concerns | Dependencies |
|-------------|------------|--------------|-------------------|--------------|
| **6 tasks (W7-T01..W7-T06)** | 5 RED test sets (W7-T01/T03/T04/T05 directly Go-test; W7-T06 markdown structural test) | 6 GREEN implementations | 4 REFACTOR concerns (os/exec wrapping, skill markdown deduplication, audit trail format stability, reminder mute pattern) | W7-T01 → W7-T04 → {W7-T02, W7-T03} → W7-T05 → W7-T06 |

---

## Atomic Task Table

| Task ID | Description | Files (provisional) | REQ | AC | Dependencies | File Ownership Scope | Status |
|---------|-------------|---------------------|-----|----|--------------|---------------------|--------|
| W7-T01 | BODP 라이브러리 핵심 (Pure-Go). `internal/bodp/relatedness.go` — `Check(input CheckInput) (BODPDecision, error)` 함수 + `Choice` enum (`ChoiceMain`/`ChoiceStacked`/`ChoiceContinue`) + `EntryPoint` enum (`EntryPlanBranch`/`EntryPlanWorktree`/`EntryWorktreeCLI`) + `BODPDecision` 출력 struct. 3-signal 평가: (a) `git diff --name-only origin/main..HEAD` + SPEC frontmatter `depends_on:` 파싱 (yaml v3) → SignalA, (b) `git status --porcelain` untracked 매칭 → SignalB, (c) `gh pr list --head <branch> --state open --json number` length ≥ 1 → SignalC (gh 미설치 시 graceful skip + warning). Decision matrix (8 rows) → Recommended Choice + BaseBranch + Rationale (한국어 const map). 외부 의존성 ZERO 추가 (yaml v3는 기존 사용). os/exec 호출은 `gitCommand`/`ghCommand` 변수 wrapping (테스트용). | `internal/bodp/relatedness.go` (new), `internal/bodp/relatedness_test.go` (new) | REQ-CIAUT-044, REQ-CIAUT-045, REQ-CIAUT-046, REQ-CIAUT-047, REQ-CIAUT-047b | AC-CIAUT-018 (transitive), AC-CIAUT-019 (transitive) | none (foundational) | implementer | pending |
| W7-T02 | `/moai plan` skill body Phase 3 확장. `.claude/skills/moai/workflows/plan.md` 의 Phase 3 Branch Path + Worktree Path 양쪽에 BODP gate sub-section 추가: (a) `internal/bodp/Check()` 호출 → `BODPDecision` 결과 수신, (b) AskUserQuestion (`(권장)` 라벨 첫 옵션 + Recommended 카테고리 + 다른 Choice enum 옵션 + "Other" 자동) 호출 — askuser-protocol.md §Socratic Interview Structure 준수, conversation_language=ko, 옵션 ≤4. (c) 사용자 응답 후 manager-git 위임 시 `base=<chosen>` parameter 명시 전달 (Branch Path) 또는 `moai worktree new <SPEC-ID> --base <chosen>` CLI 호출 (Worktree Path). (d) W7-T04 `WriteDecision()` 호출로 audit trail 기록. 두 path 의 BODP gate 공통 부분은 단일 sub-section ("§Phase 3.X: BODP Gate (공통)") 으로 추출. | `.claude/skills/moai/workflows/plan.md` (extend) | REQ-CIAUT-042, REQ-CIAUT-043, REQ-CIAUT-048 | AC-CIAUT-018 (canonical), AC-CIAUT-019 (canonical) | W7-T01, W7-T04 | implementer | pending |
| W7-T03 | `moai worktree new` CLI 확장. `internal/cli/worktree/new.go` 에 `--base <branch>` (string, default `origin/main`) + `--from-current` (bool, default false) flag 추가. Mutual exclusion 검증 (`TestNew_BaseAndFromCurrentMutuallyExclusive`). Base 결정 로직: (a) `--from-current` → 현재 HEAD, (b) `--base <X>` → X, (c) default → `origin/main` (`git fetch origin main` 사전 실행 후 `git worktree add <path> origin/main`). [HARD] AskUserQuestion 호출 절대 금지 (orchestrator-only HARD per agent-common-protocol §User Interaction Boundary; W7-T03 정적 import 검사). 호출 후 `internal/bodp/audit_trail.go` `WriteDecision()` with `EntryPoint: EntryWorktreeCLI` 호출 (signal 검사 결과 기록, UserChoice = Recommended since no prompt). | `internal/cli/worktree/new.go` (extend), `internal/cli/worktree/new_test.go` (extend) | REQ-CIAUT-043, REQ-CIAUT-043b, REQ-CIAUT-048 | AC-CIAUT-019b | W7-T01, W7-T04 | implementer | pending |
| W7-T04 | Audit trail writer 라이브러리. `internal/bodp/audit_trail.go` — `AuditEntry` struct (Timestamp, EntryPoint, CurrentBranch, NewBranch, Decision, UserChoice, ExecutedCmd) + `WriteDecision(repoRoot string, entry AuditEntry) error` (markdown frontmatter + body 작성, `.moai/branches/decisions/<branch>.md`, slash → dash 정규화, `os.MkdirAll` 자동 생성) + `HasAuditTrail(repoRoot, branchName string) bool` (W7-T05 reader). const: `auditTrailDir = ".moai/branches/decisions"`. 신규 디렉토리 보존: `.moai/branches/decisions/.gitkeep` 신규 파일. | `internal/bodp/audit_trail.go` (new), `internal/bodp/audit_trail_test.go` (new), `.moai/branches/decisions/.gitkeep` (new) | REQ-CIAUT-049 | AC-CIAUT-018 (audit trail), AC-CIAUT-019 (audit trail), AC-CIAUT-019b (audit trail) | W7-T01 | implementer | pending |
| W7-T05 | `moai status` off-protocol branch reminder. `internal/cli/status.go` 확장 — 기존 status 출력 끝에 BODP off-protocol section 추가. 흐름: (a) `MOAI_NO_BODP_REMINDER=1` env var 검사 → set이면 return early, (b) currentBranch 가 `main`/`master` 면 return early, (c) `internal/bodp/audit_trail.go` `HasAuditTrail()` true → return early, (d) audit trail 디렉토리 자체 부재 (`os.Stat .moai/branches/decisions/`) → return early (false-positive 방지), (e) 위 모두 false → friendly reminder 출력 (verbatim 한국어 메시지 const, ⚠️ 이모지 + 권장 entry point 명시 + env var 비활성화 방법). exit code 0 (block 안 함). | `internal/cli/status.go` (extend), `internal/cli/status_test.go` (extend) | REQ-CIAUT-050 | AC-CIAUT-024 | W7-T04 (HasAuditTrail consumer) | implementer | pending |
| W7-T06 | 문서화 + Template-First mirror. (a) `CLAUDE.local.md` §18.11 다음에 §18.12 신규 subsection 추가 (BODP 알고리즘 요약, decision matrix, 3 invocation paths verbatim, audit trail 위치/형식, off-protocol reminder + opt-out 방법, Out of Scope). (b) `.claude/rules/moai/development/branch-origin-protocol.md` 신규 rule 파일 (frontmatter `paths:` 로 BODP 코드 경로 자동 적용, HARD rules + 3 entry point 요약 + cross-references). (c) Template-First mirror: `internal/template/templates/CLAUDE.local.md` 에 동일 §18.12 + `internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md` 동일 내용. (d) `make build` 후 embedded.go 재생성 검증. | `CLAUDE.local.md` (extend), `.claude/rules/moai/development/branch-origin-protocol.md` (new), `internal/template/templates/CLAUDE.local.md` (extend mirror), `internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md` (new mirror) | REQ-CIAUT-051 | AC-CIAUT-025 | W7-T02, W7-T03, W7-T04, W7-T05 (모두 사양 fixture 후 문서화) | implementer | pending |

---

## W7-T01 — BODP Library (Pure-Go)

### RED Phase — Test Cases

File: `internal/bodp/relatedness_test.go`

Function-level test cases (table-driven where possible):

1. **`TestRelatedness_AllNegative_RecommendsMain`** (current session replay; AC-CIAUT-018 canonical)
   - Setup: mock `gitCommand` returns clean diff vs main (no path overlap), mock `ghCommand` returns empty PR list, working tree without untracked match for new SPEC ID.
   - Input: `CheckInput{CurrentBranch: "chore/translation-batch-b", NewSpecID: "SPEC-V3R3-CI-AUTONOMY-001", RepoRoot: t.TempDir(), EntryPoint: EntryPlanBranch}`
   - Expected: `BODPDecision{SignalA: false, SignalB: false, SignalC: false, Recommended: ChoiceMain, BaseBranch: "origin/main", Rationale: "현재 브랜치와 무관한 새 작업이므로 main 분기를 권장합니다."}`

2. **`TestRelatedness_SignalA_RecommendsStacked`** (AC-CIAUT-019 canonical)
   - Setup: mock SPEC frontmatter file with `depends_on: [SPEC-AUTH-001]`; current branch name `feat/SPEC-AUTH-001-base` matches via heuristic (branch name contains depends_on SPEC ID).
   - Input: `CheckInput{CurrentBranch: "feat/SPEC-AUTH-001-base", NewSpecID: "SPEC-AUTH-002", ..., EntryPoint: EntryPlanBranch}`
   - Expected: `BODPDecision{SignalA: true, SignalB: false, SignalC: false, Recommended: ChoiceStacked, BaseBranch: "feat/SPEC-AUTH-001-base", Rationale 에 "depends_on" 또는 "stacked PR" 포함}`

3. **`TestRelatedness_SignalB_RecommendsContinue`**
   - Setup: working tree에 `.moai/specs/SPEC-FOO-001/` 디렉토리 (untracked) 생성; mock `git status --porcelain` 가 해당 path 반환.
   - Input: `CheckInput{CurrentBranch: "feat/foo", NewSpecID: "SPEC-FOO-001", ..., EntryPoint: EntryPlanWorktree}`
   - Expected: `BODPDecision{SignalA: false, SignalB: true, SignalC: false, Recommended: ChoiceContinue, BaseBranch: "" (empty for continue)}`

4. **`TestRelatedness_SignalC_RecommendsStackedWithGotchaWarning`**
   - Setup: mock `ghCommand` returns `[{"number":42}]` (1 open PR with current as head).
   - Input: `CheckInput{CurrentBranch: "feat/X", NewSpecID: "SPEC-Y-001", ..., EntryPoint: EntryPlanBranch}`
   - Expected: `BODPDecision{SignalA: false, SignalB: false, SignalC: true, Recommended: ChoiceStacked, BaseBranch: "feat/X"}`. **Rationale 문자열은 "parent-merge gotcha" 또는 "§18.11" 명시 포함** (REQ-CIAUT-047b verbatim).

5. **`TestRelatedness_SignalsAandC_RecommendsStacked`** (decision matrix row 5)
   - Setup: signal A + C 모두 positive (depends_on 매칭 + open PR 존재).
   - Expected: `Recommended: ChoiceStacked, BaseBranch: "<currentBranch>"`.

Optional 테이블 case (RED 추가 — 시간 허락 시):
- `TestRelatedness_AllSignalsPositive_RecommendsContinue` (decision matrix row 8: a+b+c → continue)
- `TestRelatedness_GhCommandUnavailable_GracefulSkip` (W7-R1 mitigation: ghCommand 함수 error 반환 → SignalC=false + warning logged, 다른 시그널 평가 정상 진행)

### GREEN Phase — Minimal Implementation

```go
// Package bodp implements the Branch Origin Decision Protocol.
package bodp

import (
    "fmt"
    "os/exec"
    "path/filepath"
    "strings"
)

type Choice string

const (
    ChoiceMain     Choice = "main"
    ChoiceStacked  Choice = "stacked"
    ChoiceContinue Choice = "continue"
)

type EntryPoint string

const (
    EntryPlanBranch   EntryPoint = "plan-branch"
    EntryPlanWorktree EntryPoint = "plan-worktree"
    EntryWorktreeCLI  EntryPoint = "worktree-cli"
)

type CheckInput struct {
    CurrentBranch string
    NewSpecID     string
    RepoRoot      string
    EntryPoint    EntryPoint
}

type BODPDecision struct {
    SignalA     bool
    SignalB     bool
    SignalC     bool
    Recommended Choice
    Rationale   string
    BaseBranch  string
}

// 의존성 역전을 위한 변수 (테스트에서 fake 주입)
var (
    gitCommand = func(args ...string) (string, error) {
        cmd := exec.Command("git", args...)
        out, err := cmd.Output()
        return string(out), err
    }
    ghCommand = func(args ...string) (string, error) {
        cmd := exec.Command("gh", args...)
        out, err := cmd.Output()
        return string(out), err
    }
)

// Decision matrix 8 rows
var rationaleMessages = map[Choice]string{
    ChoiceMain:     "현재 브랜치와 무관한 새 작업이므로 main 분기를 권장합니다.",
    ChoiceStacked:  "현재 브랜치와의 의존성이 감지되어 stacked PR을 권장합니다.",
    ChoiceContinue: "현재 브랜치에 untracked SPEC plan 또는 진행 중 작업이 있어 계속 작업을 권장합니다.",
}

const (
    defaultBase = "origin/main"
)

func Check(input CheckInput) (BODPDecision, error) {
    decision := BODPDecision{}

    // Signal A: code dependency
    decision.SignalA = checkSignalA(input)
    // Signal B: working tree co-location
    decision.SignalB = checkSignalB(input)
    // Signal C: open PR head
    decision.SignalC = checkSignalC(input)

    // Decision matrix
    decision.Recommended, decision.BaseBranch = applyMatrix(decision, input.CurrentBranch)
    decision.Rationale = rationaleMessages[decision.Recommended]
    if decision.SignalC {
        // REQ-CIAUT-047b verbatim: parent-merge gotcha warning
        decision.Rationale += " (parent-merge gotcha 주의: CLAUDE.local.md §18.11 Case Study 참조)"
    }
    return decision, nil
}

func applyMatrix(d BODPDecision, currentBranch string) (Choice, string) {
    switch {
    case !d.SignalA && !d.SignalB && !d.SignalC:
        return ChoiceMain, defaultBase
    case d.SignalB:
        // b 단독 또는 b 포함 조합 → continue (decision matrix row 3, 6, 7, 8)
        return ChoiceContinue, ""
    case d.SignalA || d.SignalC:
        return ChoiceStacked, currentBranch
    default:
        return ChoiceMain, defaultBase
    }
}

func checkSignalA(input CheckInput) bool {
    // (1) SPEC frontmatter depends_on 파싱
    specFile := filepath.Join(input.RepoRoot, ".moai", "specs", input.NewSpecID, "spec.md")
    if dependsOn, err := parseDependsOn(specFile); err == nil {
        for _, dep := range dependsOn {
            if strings.Contains(input.CurrentBranch, dep) {
                return true
            }
        }
    }
    // (2) diff path overlap
    out, err := gitCommand("diff", "--name-only", "origin/main..HEAD")
    if err != nil {
        return false // graceful skip
    }
    return strings.Contains(out, input.NewSpecID) // simple heuristic
}

func checkSignalB(input CheckInput) bool {
    out, err := gitCommand("status", "--porcelain")
    if err != nil {
        return false
    }
    target := ".moai/specs/" + input.NewSpecID + "/"
    for _, line := range strings.Split(out, "\n") {
        if strings.Contains(line, target) {
            return true
        }
    }
    return false
}

func checkSignalC(input CheckInput) bool {
    out, err := ghCommand("pr", "list", "--head", input.CurrentBranch, "--state", "open", "--json", "number")
    if err != nil {
        // graceful skip — gh 미설치/미인증 시
        return false
    }
    return strings.TrimSpace(out) != "[]"
}

func parseDependsOn(specFile string) ([]string, error) {
    // yaml v3 frontmatter 파싱 stub
    // 실제 구현은 spec frontmatter 파싱 helper 재사용
    return nil, fmt.Errorf("not implemented")
}
```

### REFACTOR Concerns

- **`os/exec` wrapping**: 테스트 가능성 위해 `gitCommand`/`ghCommand` 변수 export. Production 에서는 default `exec.Command` wrapper, 테스트에서는 fake 함수 주입.
- **SPEC frontmatter 재사용**: `internal/spec/` 패키지 또는 yaml v3 직접 호출 — Wave 7 시점에 기존 helper 존재 시 재사용 (만일 부재 시 inline 파싱).
- **Rationale i18n**: ko 단일 (Wave 7 scope). const map 으로 추출하여 향후 i18n SPEC 에서 i18n 라이브러리 적용 가능.

### MX Tag Obligations

- `@MX:NOTE` on `Check()` function: "BODP는 새 슬래시 명령어/CLI 서브명령어 ZERO 원칙. 3개 entry point 공유 라이브러리." (rationale 전달, plan.md §10.5 + 사용자 critique 2026-05-05 reference)
- `@MX:NOTE` on `applyMatrix()`: "Decision matrix verbatim from strategy-wave7.md §4.1; truth table 8 rows; SignalB 우선순위 (continue) > A/C (stacked)" (intent delivery)
- `@MX:ANCHOR` on `Check()` if fan_in ≥ 3 (3개 entry point에서 호출되므로 fan_in = 3, 즉 anchor 후보): "BODP invariant: 3-signal evaluation order is independent; decision matrix is total (no undefined cases)."

### Template-First Mirror

NOT applicable — `internal/bodp/` 은 dev-project Go 코드 (Wave 6 §7 동일 패턴, Wave 4 `.github/workflows/optional/` 선례). [HARD] verify-via-grep: `ls internal/template/templates/internal/bodp/` 빈 결과.

---

## W7-T02 — Skill Body Phase 3 BODP Gate

### RED Phase — Test Cases

본 task는 markdown skill body 확장으로 직접 단위 테스트 부재. 대신 **integration verification** + plan-auditor schema verify:

1. **`TestPlanBranchSkill_BODPGateInvocation`** (optional integration test in `internal/cli/plan_branch_integration_test.go`)
   - Mock orchestrator 흐름: `/moai plan --branch` invocation → BODP `Check()` 호출 → AskUserQuestion mock 응답 → manager-git delegation
   - Verify: manager-git invocation includes `base=<chosen>` parameter

2. **`TestPlanWorktreeSkill_BODPGateInvocation`** (optional integration test)
   - Verify: `moai worktree new <SPEC-ID> --base <chosen>` CLI invocation includes correct `--base` flag value

3. **plan-auditor schema verify** (Phase 1.5 단계, no separate Go test):
   - Skill body 의 BODP gate sub-section 이 다음 조건 만족:
     - `(권장)` 라벨 첫 옵션 명시
     - conversation_language=ko 옵션 텍스트
     - AskUserQuestion options ≤ 4
     - `Other` 자동 옵션 mention
     - manager-git base parameter 전달 명시

### GREEN Phase — Minimal Implementation

`.claude/skills/moai/workflows/plan.md` Phase 3 확장:

```markdown
### Phase 3: Branch / Worktree Path Selection

#### Phase 3.X: BODP Gate (공통)

이 sub-section은 Branch Path와 Worktree Path 모두에서 manager-git 또는 `moai worktree new` 호출 직전에 실행됩니다.

1. **BODP Relatedness Check**
   - Orchestrator는 `internal/bodp/relatedness.go` `Check()` 함수를 호출하여 3-signal 결과 + 권장 옵션 (`Recommended` Choice) 수신.
   - 결과 struct 의 SignalA/B/C 값과 Rationale 문자열을 사용자에게 표시 준비.

2. **AskUserQuestion Gate** (slash path 항상 실행)
   - `ToolSearch(query: "select:AskUserQuestion")` preload (deferred tool — askuser-protocol.md §ToolSearch Preload Procedure).
   - 옵션 구성 (≤4):
     - 1번째: Recommended Choice (`(권장)` 라벨 부착) — Rationale 문자열 description
     - 2번째: 다른 Choice 옵션 (예: Recommended가 ChoiceMain이면 ChoiceStacked + ChoiceContinue 옵션 추가)
     - 3-4번째: 추가 Choice 옵션 (시그널 별 메시지)
     - "Other" 자동 (Claude Code default)
   - conversation_language=ko 준수, 옵션 라벨/설명 모두 한국어.
   - 사용자 응답 = `chosenChoice` (Choice enum) + `chosenBase` (string).

3. **Audit Trail Write**
   - `internal/bodp/audit_trail.go` `WriteDecision(repoRoot, AuditEntry{...})` 호출.
   - EntryPoint = `EntryPlanBranch` (Branch Path) 또는 `EntryPlanWorktree` (Worktree Path).
   - UserChoice = chosenChoice; ExecutedCmd = manager-git 또는 moai worktree 호출에 대응하는 git 명령어.

4. **Path-Specific Delegation**
   - **Branch Path**: manager-git 위임 시 `base=<chosenBase>` parameter 명시 전달.
   - **Worktree Path**: `moai worktree new <SPEC-ID> --base <chosenBase>` CLI 호출.

> **Out of Scope (BODP Gate)**: AskUserQuestion 응답이 "Other" 인 경우 사용자가 자유 텍스트 입력 — orchestrator는 입력을 base branch name 으로 해석 시도; 잘못된 입력 시 fail-safe (origin/main fallback + warning).
```

REFACTOR (post-GREEN): Branch Path와 Worktree Path 의 BODP gate 호출 부분이 동일하므로 위 sub-section을 단일 location 에서 정의하고, 두 path 에서 cross-reference (`See Phase 3.X: BODP Gate (공통)`).

### REFACTOR Concerns

- **Skill body 텍스트 중복 회피**: BODP gate sub-section을 별도 sub-section 으로 단일 정의, Branch Path와 Worktree Path 에서 cross-reference 사용.
- **AskUserQuestion option label 표준화**: spec.md §3.8 REQ-CIAUT-045/046/047/047b 의 한국어 문구 verbatim 사용.

### MX Tag Obligations

- `@MX:NOTE` (skill body markdown comment): "BODP gate sub-section은 두 path 공통; AskUserQuestion preload 필수 (deferred tool)."

### Template-First Mirror

`.claude/skills/moai/workflows/plan.md` 는 이미 Template-First 적용 대상. `internal/template/templates/.claude/skills/moai/workflows/plan.md` 에 동일 변경 필요. **하지만 Wave 7 plan.md (이 SPEC 의 plan.md, NOT skill body) 는 아니다** — 즉, `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/plan.md` 는 read-only.

---

## W7-T03 — `moai worktree new` CLI 확장

### RED Phase — Test Cases

File: `internal/cli/worktree/new_test.go` (extend)

1. **`TestNew_DefaultBaseIsOriginMain`**
   - Setup: `t.TempDir()` 내 git repo 초기화, origin/main 설정.
   - Invoke: `moai worktree new SPEC-FOO-001` (no flag).
   - Expected: `git worktree add` 호출 시 base = `origin/main`. 호출 전 `git fetch origin main` 실행 검증.
   - Audit trail: `.moai/branches/decisions/feat-SPEC-FOO-001.md` 생성, `entry_point: worktree-cli`.

2. **`TestNew_FromCurrentFlagPreservesOldBehavior`**
   - Invoke: `moai worktree new SPEC-FOO-001 --from-current`
   - Expected: `git worktree add` 호출 시 base = current HEAD (no `git fetch`). Audit trail 에 `user_choice: continue` 또는 base 명시.

3. **`TestNew_BaseFlagSpecifiesBranch`**
   - Invoke: `moai worktree new SPEC-FOO-001 --base feat/SPEC-AUTH-001-base`
   - Expected: `git worktree add` base = `feat/SPEC-AUTH-001-base`. (W7-T02 skill 에서 `--base` flag 전달 시뮬레이션.)

4. **`TestNew_NoAskUserQuestion`** (W7-R4 mitigation)
   - Setup: AskUserQuestion API mock with call counter.
   - Invoke: 임의 invocation (default + `--from-current` + `--base` 모든 케이스).
   - Expected: AskUserQuestion 호출 횟수 = 0.
   - **추가 정적 검사**: `import "...AskUserQuestion..."` 가 `internal/cli/worktree/new.go` 에 부재 (godebug-friendly verification).

5. **`TestNew_BaseAndFromCurrentMutuallyExclusive`**
   - Invoke: `moai worktree new SPEC-FOO-001 --base X --from-current`
   - Expected: error message "flags --base and --from-current are mutually exclusive", non-zero exit code.

6. **`TestNew_AuditTrailWritten`**
   - Setup: mock `internal/bodp/audit_trail.go` `WriteDecision` 함수 (variable injection 또는 interface).
   - Invoke: 임의 invocation.
   - Expected: `WriteDecision` 호출 1회, `EntryPoint: EntryWorktreeCLI` 인자 포함.

### GREEN Phase — Minimal Implementation

```go
// internal/cli/worktree/new.go (extend)

package worktree

import (
    "fmt"
    "os/exec"
    "github.com/modu-ai/moai-adk/internal/bodp"
    "github.com/spf13/cobra"
)

func newCommand() *cobra.Command {
    var (
        baseFlag        string
        fromCurrentFlag bool
    )
    cmd := &cobra.Command{
        Use:   "new <SPEC-ID>",
        Short: "Create a new worktree for a SPEC",
        RunE: func(cmd *cobra.Command, args []string) error {
            if baseFlag != "" && fromCurrentFlag {
                return fmt.Errorf("flags --base and --from-current are mutually exclusive")
            }

            specID := args[0]
            base := determineBase(baseFlag, fromCurrentFlag)

            if base == "origin/main" {
                if _, err := exec.Command("git", "fetch", "origin", "main").CombinedOutput(); err != nil {
                    return fmt.Errorf("git fetch origin main failed: %w", err)
                }
            }

            // git worktree add
            worktreePath := fmt.Sprintf("../%s-worktree", specID)
            if _, err := exec.Command("git", "worktree", "add", worktreePath, base).CombinedOutput(); err != nil {
                return fmt.Errorf("git worktree add failed: %w", err)
            }

            // Audit trail
            return writeAuditTrail(specID, base)
        },
    }
    cmd.Flags().StringVar(&baseFlag, "base", "", "base branch (default: origin/main)")
    cmd.Flags().BoolVar(&fromCurrentFlag, "from-current", false, "use current HEAD as base (opt-out of origin/main default)")
    return cmd
}

const defaultBase = "origin/main"

func determineBase(baseFlag string, fromCurrentFlag bool) string {
    if fromCurrentFlag {
        return "HEAD"
    }
    if baseFlag != "" {
        return baseFlag
    }
    return defaultBase
}

func writeAuditTrail(specID, base string) error {
    // Get current branch
    out, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
    currentBranch := string(out)

    // BODP check (signal collection only — CLI path does not prompt)
    decision, err := bodp.Check(bodp.CheckInput{
        CurrentBranch: currentBranch,
        NewSpecID:     specID,
        RepoRoot:      ".",
        EntryPoint:    bodp.EntryWorktreeCLI,
    })
    if err != nil {
        return err
    }

    return bodp.WriteDecision(".", bodp.AuditEntry{
        EntryPoint:    bodp.EntryWorktreeCLI,
        CurrentBranch: currentBranch,
        NewBranch:     fmt.Sprintf("feat/%s", specID),
        Decision:      decision,
        UserChoice:    decision.Recommended, // CLI path has no prompt
        ExecutedCmd:   fmt.Sprintf("git worktree add ../%s-worktree %s", specID, base),
    })
}
```

### REFACTOR Concerns

- **CLI flag naming**: `--base` 와 `--from-current` 사용 (plan.md §9 W7-T03 표기 일치).
- **AskUserQuestion 부재 보장**: 코드 리뷰 시 `import` 목록 정적 검사. lint custom rule 가능 (Wave 7 ship 안 함).
- **`writeAuditTrail` 분리**: testability 위해 별도 함수로 추출.

### MX Tag Obligations

- `@MX:NOTE` on `newCommand()`: "CLI path는 AskUserQuestion 호출 절대 금지 (orchestrator-only HARD per agent-common-protocol §User Interaction Boundary)."
- `@MX:WARN` on `determineBase()`: `--base` 와 `--from-current` mutual exclusive — 둘 동시 지정 시 error. `@MX:REASON`: "사용자 의도 모호성 방지 (BODP rationale clarity)."

### Template-First Mirror

NOT applicable — `internal/cli/worktree/new.go` 는 dev-project Go 코드.

---

## W7-T04 — Audit Trail Writer

### RED Phase — Test Cases

File: `internal/bodp/audit_trail_test.go`

1. **`TestWriteDecision_CreatesFile`**
   - Setup: `t.TempDir()` repoRoot.
   - Invoke: `WriteDecision(repoRoot, AuditEntry{Timestamp: time.Now(), EntryPoint: EntryPlanBranch, CurrentBranch: "chore/x", NewBranch: "feat/SPEC-Y", Decision: BODPDecision{Recommended: ChoiceMain, BaseBranch: "origin/main", ...}, UserChoice: ChoiceMain, ExecutedCmd: "git fetch origin main && git checkout -B feat/SPEC-Y origin/main"})`.
   - Expected: `<repoRoot>/.moai/branches/decisions/feat-SPEC-Y.md` 파일 생성. Frontmatter 검증 (timestamp, entry_point, current_branch, new_branch, user_choice). Body 검증 (signals, decision, executed cmd).

2. **`TestWriteDecision_NormalizesBranchNameForFilename`**
   - Setup: `NewBranch = "feat/SPEC-A/sub-feature"`
   - Expected: 파일명 `feat-SPEC-A-sub-feature.md` (slash → dash).

3. **`TestWriteDecision_CreatesDirectoryIfAbsent`**
   - Setup: `<repoRoot>/.moai/branches/decisions/` 디렉토리 부재 상태.
   - Expected: `os.MkdirAll` 호출 + 파일 생성 성공 (no error).

4. **`TestHasAuditTrail_DetectsExisting`**
   - Setup: pre-existing `<repoRoot>/.moai/branches/decisions/feat-X.md`.
   - Invoke: `HasAuditTrail(repoRoot, "feat/X")`.
   - Expected: `true`.

5. **`TestHasAuditTrail_AbsentBranch`**
   - Setup: 디렉토리 존재, but no matching file.
   - Expected: `false`.

6. **`TestHasAuditTrail_DirAbsentReturnsFalse`** (W7-T05 false-positive 방지 의존)
   - Setup: `.moai/branches/decisions/` 디렉토리 자체 부재 (신규 프로젝트).
   - Expected: `false` (no error).

### GREEN Phase — Minimal Implementation

```go
// internal/bodp/audit_trail.go (new)

package bodp

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
)

const auditTrailDir = ".moai/branches/decisions"

type AuditEntry struct {
    Timestamp     time.Time
    EntryPoint    EntryPoint
    CurrentBranch string
    NewBranch     string
    Decision      BODPDecision
    UserChoice    Choice
    ExecutedCmd   string
}

func WriteDecision(repoRoot string, entry AuditEntry) error {
    if entry.Timestamp.IsZero() {
        entry.Timestamp = time.Now().UTC()
    }
    fname := normalizeBranchName(entry.NewBranch) + ".md"
    fullPath := filepath.Join(repoRoot, auditTrailDir, fname)

    if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
        return fmt.Errorf("create audit trail dir: %w", err)
    }

    content := renderAuditEntry(entry)
    return os.WriteFile(fullPath, []byte(content), 0o644)
}

func HasAuditTrail(repoRoot, branchName string) bool {
    fname := normalizeBranchName(branchName) + ".md"
    fullPath := filepath.Join(repoRoot, auditTrailDir, fname)
    _, err := os.Stat(fullPath)
    return err == nil
}

func normalizeBranchName(name string) string {
    return strings.ReplaceAll(name, "/", "-")
}

func renderAuditEntry(entry AuditEntry) string {
    var sb strings.Builder
    sb.WriteString("---\n")
    fmt.Fprintf(&sb, "timestamp: %s\n", entry.Timestamp.Format(time.RFC3339))
    fmt.Fprintf(&sb, "entry_point: %s\n", entry.EntryPoint)
    fmt.Fprintf(&sb, "current_branch: %s\n", entry.CurrentBranch)
    fmt.Fprintf(&sb, "new_branch: %s\n", entry.NewBranch)
    fmt.Fprintf(&sb, "user_choice: %s\n", entry.UserChoice)
    sb.WriteString("---\n\n")
    fmt.Fprintf(&sb, "# BODP Decision: %s\n\n", entry.NewBranch)
    sb.WriteString("## Signals\n")
    fmt.Fprintf(&sb, "- Signal (a) — Code dependency: %t\n", entry.Decision.SignalA)
    fmt.Fprintf(&sb, "- Signal (b) — Working tree co-location: %t\n", entry.Decision.SignalB)
    fmt.Fprintf(&sb, "- Signal (c) — Open PR head: %t\n", entry.Decision.SignalC)
    sb.WriteString("\n## Decision\n")
    fmt.Fprintf(&sb, "- Recommended: %s\n", entry.Decision.Recommended)
    fmt.Fprintf(&sb, "- User choice: %s\n", entry.UserChoice)
    fmt.Fprintf(&sb, "- Base branch: %s\n", entry.Decision.BaseBranch)
    fmt.Fprintf(&sb, "- Rationale: %s\n", entry.Decision.Rationale)
    sb.WriteString("\n## Executed\n```\n")
    sb.WriteString(entry.ExecutedCmd)
    sb.WriteString("\n```\n")
    return sb.String()
}
```

### REFACTOR Concerns

- **파일 형식 stability**: 향후 audit 분석 도구 (예: `moai branch decisions list`) 가능성 — frontmatter 구조 명확화. 단 Wave 7은 writer + HasAuditTrail reader 만 ship.
- **Concurrent write safety**: orchestrator 단일 세션 가정. file lock 은 follow-up SPEC.

### MX Tag Obligations

- `@MX:NOTE` on `WriteDecision()`: "Audit trail은 BODP 결정의 영구 기록. branch name 정규화 (slash → dash) 로 filesystem-safe."
- `@MX:NOTE` on `HasAuditTrail()`: "W7-T05 reminder false-positive 방지: 디렉토리 자체 부재 시 false 반환 (no error 구분)."

### Template-First Mirror

NOT applicable — `internal/bodp/` 은 dev-project Go 코드. `.moai/branches/decisions/.gitkeep` 은 사용자 프로젝트에서 동일하게 생성될 디렉토리 구조 표시 — Template-First mirror 필요할 수 있음 (Phase 1.5 plan-auditor 가 verify; Wave 7 보수적으로 mirror 포함).

`internal/template/templates/.moai/branches/decisions/.gitkeep` 추가 검토.

---

## W7-T05 — `moai status` Off-Protocol Reminder

### RED Phase — Test Cases

File: `internal/cli/status_test.go` (extend)

1. **`TestStatus_OffProtocolBranchReminder`** (AC-CIAUT-024 canonical)
   - Setup: `t.TempDir()` 내 git repo, current branch `feat/quick-fix` (raw `git checkout -b` 시뮬레이션). `.moai/branches/decisions/` 디렉토리 존재 but `feat-quick-fix.md` 부재.
   - Invoke: `moai status`.
   - Expected: stderr 또는 stdout 에 "off-protocol branch detected" 또는 "Branch `feat/quick-fix` was created without going through MoAI entry points" 포함. exit code 0 (block 안 함).

2. **`TestStatus_AuditTrailExistsNoReminder`**
   - Setup: pre-existing `.moai/branches/decisions/feat-SPEC-X.md` + current branch `feat/SPEC-X`.
   - Expected: reminder 출력 안 함.

3. **`TestStatus_AuditTrailDirAbsentNoFalsePositive`** (acceptance.md AC-CIAUT-024 Failure Mode)
   - Setup: `.moai/branches/decisions/` 디렉토리 자체 부재 (신규 프로젝트).
   - Expected: reminder 출력 안 함 (false-positive 방지).

4. **`TestStatus_EnvVarDisablesReminder`**
   - Setup: `t.Setenv("MOAI_NO_BODP_REMINDER", "1")` + audit trail 부재 상태 (정상이면 reminder 출력 조건).
   - Expected: reminder 출력 안 함.

5. **`TestStatus_MainBranchNoReminder`**
   - Setup: current branch `main` (또는 `master`).
   - Expected: reminder 출력 안 함 (main 자체는 BODP 대상 아님).

### GREEN Phase — Minimal Implementation

```go
// internal/cli/status.go (extend)

package cli

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "github.com/modu-ai/moai-adk/internal/bodp"
)

const (
    envNoReminder      = "MOAI_NO_BODP_REMINDER"
    auditTrailDirCheck = ".moai/branches/decisions"
)

var mainBranches = []string{"main", "master"}

const reminderMessage = `⚠️  Branch ` + "`%s`" + ` was created without going through MoAI entry points.
Future branches: use ` + "`/moai plan --branch <name>`" + ` (SPEC-tied) or ` + "`moai worktree new <SPEC-ID>`" + ` for relatedness check + audit trail.
Skip with ` + "`MOAI_NO_BODP_REMINDER=1`" + ` if intentional.`

// emitOffProtocolReminder는 status 명령어 끝에 호출됨.
// exit code 0 (block 안 함).
func emitOffProtocolReminder(repoRoot string) {
    if os.Getenv(envNoReminder) == "1" {
        return
    }

    out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
    if err != nil {
        return
    }
    currentBranch := strings.TrimSpace(string(out))

    for _, mainName := range mainBranches {
        if currentBranch == mainName {
            return
        }
    }

    if bodp.HasAuditTrail(repoRoot, currentBranch) {
        return
    }

    // false-positive 방지: 디렉토리 자체 부재 시 reminder 출력 안 함
    dirPath := filepath.Join(repoRoot, auditTrailDirCheck)
    if _, err := os.Stat(dirPath); os.IsNotExist(err) {
        return
    }

    fmt.Fprintln(os.Stderr)
    fmt.Fprintf(os.Stderr, reminderMessage, currentBranch)
    fmt.Fprintln(os.Stderr)
}
```

### REFACTOR Concerns

- **Reminder 빈도 제한**: 본 Wave 7 ship 안 함. `.moai/state/bodp-reminder-shown-<branch>` flag 는 follow-up SPEC.
- **i18n**: 한국어 메시지 const string. follow-up SPEC.

### MX Tag Obligations

- `@MX:NOTE` on `emitOffProtocolReminder()`: "Off-protocol detection은 reminder만 — block 안 함 (REQ-CIAUT-050). MOAI_NO_BODP_REMINDER=1 비활성화 가능."
- `@MX:NOTE` on dir-absent check: "신규 프로젝트 false-positive 방지 (acceptance.md AC-CIAUT-024 Failure Mode)."

### Template-First Mirror

NOT applicable — `internal/cli/status.go` 는 dev-project Go 코드.

---

## W7-T06 — Documentation + Template-First Mirror

### RED Phase — Test Cases

본 task는 markdown 문서 작업. Structural verification 만 수행:

1. **`TestDocsCLAUDELocalMd_Section1812Exists`**
   - Verify: `grep '## 18.12\|### §18.12' CLAUDE.local.md` 결과 1+ 매칭.

2. **`TestDocsCLAUDELocalMd_TemplateMirror`** (Template-First HARD)
   - Verify: `internal/template/templates/CLAUDE.local.md` 에도 §18.12 존재.
   - Diff: local CLAUDE.local.md 의 §18.12 와 template mirror §18.12 가 byte-identical (혹은 사용자 환경 specific 부분 제외 동일).

3. **`TestDocsRule_BranchOriginProtocolExists`**
   - Verify: `.claude/rules/moai/development/branch-origin-protocol.md` 파일 존재.
   - Frontmatter `paths:` 필드 포함 검증.

4. **`TestDocsRule_TemplateMirror`** (Template-First HARD)
   - Verify: `internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md` 동일 내용 mirror.

5. **`TestDocsBuildEmbedded`** (post `make build`)
   - Verify: `internal/template/embedded.go` 에 `branch-origin-protocol.md` 내용 포함됨.

### GREEN Phase — Minimal Implementation

#### CLAUDE.local.md §18.12 신규 subsection

`CLAUDE.local.md` §18.11 (line 1036) 다음에 `### §18.12` 추가 (전체 텍스트는 strategy-wave7.md §5 W7-T06 GREEN 섹션 참조).

#### .claude/rules/moai/development/branch-origin-protocol.md

Frontmatter:

```yaml
---
paths:
  - "internal/cli/worktree/**/*.go"
  - "internal/cli/status.go"
  - "internal/bodp/**/*.go"
  - ".claude/skills/moai/workflows/plan.md"
---
```

Body: BODP HARD rules + 3 entry point + cross-references (전체 텍스트는 strategy-wave7.md §5 W7-T06 GREEN 섹션 참조).

#### Template-First Mirror

```bash
cp CLAUDE.local.md internal/template/templates/CLAUDE.local.md
# Or: 부분 patch — §18.12 subsection만 추가
cp .claude/rules/moai/development/branch-origin-protocol.md internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md
make build  # embedded.go 재생성
```

### REFACTOR Concerns

- **Mirror 일관성**: `make build` 후 `embedded.go` 재생성 + git diff 검증 (CLAUDE.local.md §2 Template-First).
- **§18.12 위치**: §18.11 다음, §19 이전 — natural numbering.
- **`paths:` frontmatter 정확성**: BODP 코드 경로만 포함 (over-specification 회피).

### MX Tag Obligations

본 task 는 markdown 문서 작업이므로 MX tag 직접 부착 대상 아님. 단 Wave 7 다른 task 의 코드 변경 시 add `@MX:NOTE` referencing branch-origin-protocol.md rule.

### Template-First Mirror Obligations

[HARD] Template-First Wave 7 specific paths:

- `CLAUDE.local.md` § 18.12 변경 → `internal/template/templates/CLAUDE.local.md` § 18.12 동일 mirror
- `.claude/rules/moai/development/branch-origin-protocol.md` → `internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md` 동일 내용
- `make build` 후 `internal/template/embedded.go` 재생성

verify-via-grep:

```bash
grep '§18.12' internal/template/templates/CLAUDE.local.md         # ≥1 match expected
ls internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md  # exists
grep -l 'branch-origin-protocol' internal/template/embedded.go    # match expected after make build
```

---

## Test Files Inventory

| Test File | Lines (estimate) | Tasks Covered | Type |
|-----------|------------------|---------------|------|
| `internal/bodp/relatedness_test.go` | ~150-200 | W7-T01 (5+ table cases) | Unit (table-driven) |
| `internal/bodp/audit_trail_test.go` | ~120-160 | W7-T04 (6 cases) | Unit |
| `internal/cli/worktree/new_test.go` (extend) | ~100-150 (new section) | W7-T03 (6 cases) | Unit |
| `internal/cli/status_test.go` (extend) | ~80-120 (new section) | W7-T05 (5 cases) | Unit |
| `internal/cli/plan_branch_integration_test.go` | ~80-120 (optional) | W7-T02 (1-2 integration cases) | Integration |
| `internal/cli/plan_worktree_integration_test.go` | ~80-120 (optional) | W7-T02 (1-2 integration cases) | Integration |

Total estimate: ~600-900 LOC for tests across 4 unit + 2 optional integration files.

---

## Wave 7 Definition of Done (DoD) Checklist

### Per-Wave DoD

각 W7 task 완료 시 모두 통과:

- [ ] All 6 W7 tasks complete (W7-T01 ~ W7-T06; see table above)
- [ ] `internal/bodp/relatedness.go` + `audit_trail.go` + 4 test files all created
- [ ] `internal/cli/worktree/new.go` extended with `--base` + `--from-current` flags + audit trail call
- [ ] `internal/cli/status.go` extended with off-protocol reminder
- [ ] `.claude/skills/moai/workflows/plan.md` Phase 3 Branch Path + Worktree Path BODP gate sub-section added
- [ ] `.claude/rules/moai/development/branch-origin-protocol.md` 신규 rule 파일 (frontmatter `paths:` + HARD rules + cross-references)
- [ ] `CLAUDE.local.md` §18.12 신규 subsection 추가
- [ ] `internal/template/templates/CLAUDE.local.md` §18.12 mirror
- [ ] `internal/template/templates/.claude/rules/moai/development/branch-origin-protocol.md` mirror
- [ ] `.moai/branches/decisions/.gitkeep` 신규 파일 + (선택) template mirror
- [ ] `make build` 후 `internal/template/embedded.go` 재생성 검증
- [ ] Test cases per W7-T01..T05 RED phases all written and initially fail
- [ ] All tests PASS post-GREEN with `go test -race ./internal/bodp/... ./internal/cli/...`
- [ ] `go test -cover ./internal/bodp/...` ≥ 85%
- [ ] `golangci-lint run ./internal/bodp/... ./internal/cli/worktree/... ./internal/cli/status.go` 0 issue
- [ ] `make ci-local` passes (Wave 1 ci-mirror framework regression 없음)
- [ ] AC-CIAUT-018 fixture replay (current session): `Check()` with all-negative input returns `Recommended: ChoiceMain`
- [ ] AC-CIAUT-019 fixture replay (signal A): `depends_on` 매칭 시 `Recommended: ChoiceStacked`
- [ ] AC-CIAUT-019b CLI verification: `moai worktree new` default base = `origin/main`, `--from-current` opt-out, no AskUserQuestion
- [ ] AC-CIAUT-024 reminder verification: off-protocol branch + audit trail dir 존재 시 reminder 출력 + env var 비활성화 + main 브랜치 false-positive 없음 + dir 부재 false-positive 없음
- [ ] AC-CIAUT-025 §18.12 verification: `grep '§18.12' CLAUDE.local.md` + Template-First mirror 동일성
- [ ] All const extracted (CLAUDE.local.md §14): `auditTrailDir`, `defaultBase`, `envNoReminder`, `mainBranches`, `rationaleMessages`
- [ ] No release/tag automation introduced
- [ ] PR labeled with `type:feature`, `priority:P0`, `area:cli`, `area:bodp` (or `area:workflow`)
- [ ] Conventional Commits + 🗿 MoAI co-author trailer all 12 commits
- [ ] CHANGELOG.md updated with Wave 7 entry on merge

### SPEC-Level Closure Checklist (Wave 7 = Final Wave)

본 Wave 7 머지로 SPEC-V3R3-CI-AUTONOMY-001 closure 진입:

- [ ] All 7 Waves complete (Wave 1-7 PR 모두 main merged)
- [ ] All 25 acceptance scenarios pass (acceptance.md §1 AC index)
- [ ] AC-CIAUT-020 (5-PR sweep replay manual validation) — post-merge 30일 grace window 활성화 (SPEC closure 시점에는 deferred 가능)
- [ ] CLAUDE.local.md §18.12 BODP subsection 추가됨
- [ ] `.claude/rules/moai/development/branch-origin-protocol.md` 신규 규칙 추가됨
- [ ] `.claude/rules/moai/workflow/worktree-state-guard.md` (Wave 5)
- [ ] `.claude/rules/moai/workflow/ci-watch-protocol.md` (Wave 2)
- [ ] `.claude/rules/moai/workflow/ci-autofix-protocol.md` (Wave 3)
- [ ] `.github/required-checks.yml` SSoT (Wave 1)
- [ ] CHANGELOG.md SPEC closure entry
- [ ] No release/tag automation across SPEC (T5 후 main 머지 시 tag/release 자동 생성 없음)
- [ ] No hardcoded URLs/models (CLAUDE.local.md §14)
- [ ] 16-language neutrality 검증 (T1, T7만 해당; Wave 7는 언어 중립)

---

## Suggested Commit Cadence

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

Optional REFACTOR commits between feat steps if shared `os/exec` wrapping (W7-T01) requires extraction or if W7-T02 markdown duplication elimination becomes large.

Each commit body ends with the verbatim trailer:

```
🗿 MoAI <email@mo.ai.kr>
```

---

## §5 Phase 1 + Closure Entry for progress.md (Copy-Paste Block)

The orchestrator should append the following block to `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/progress.md` after Phase 1 (manager-spec strategy + tasks) completion + audit:

```markdown
## Wave 7 — Phase 1 (Strategy + Tasks)

- date: 2026-05-09
- author: manager-spec
- artifacts:
  - .moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave7.md (status: draft, version 0.1.0)
  - .moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave7.md (status: draft, version 0.1.0)
- summary:
  - Wave 7 scope: T8 Branch Origin Decision Protocol (P0). 새 슬래시 명령어 / CLI 서브명령어 ZERO 원칙 (plan.md §10.5 + 사용자 critique 2026-05-05).
  - 6 atomic tasks (W7-T01..W7-T06) covering pure-Go bodp library + audit trail writer + skill Phase 3 BODP gate + worktree CLI flag extension + status off-protocol reminder + CLAUDE.local.md §18.12 documentation.
  - 3 entry points reuse shared `internal/bodp/` library: `/moai plan --branch`, `/moai plan --worktree`, `moai worktree new`.
  - REQ mapping: REQ-CIAUT-042..051 (11 requirements) all bound.
  - AC mapping: AC-CIAUT-018, 019, 019b, 024, 025 all bound.
  - 3 Wave-local Open Questions resolved inline (strategy §6 Wave7-Q1..Q3): library location (`internal/bodp/`), Signal A heuristic (depends_on + path overlap both), AskUserQuestion option order (`(권장)` first per askuser-protocol.md).
  - Solo mode (--branch pattern, lessons #13). Wave Base: origin/main 78929d058.
  - Template-First mirror: CLAUDE.local.md §18.12 + branch-origin-protocol.md rule (W7-T06).
  - Final Wave (7/7) — closure 진입 준비.
- next:
  - Phase 1.5 PLAN AUDIT: plan-auditor review (must-pass: REQ coverage 11/11, EARS compliance, AC coverage 5/5 with verification path, file ownership clarity, no-new-command 원칙 verify)
  - On PASS: Phase 2 (manager-tdd 위임 또는 main-session 직접 구현 fallback per Wave 4/5/6 §C-7 lesson)

## Wave 7 — Phase 1.5 (Plan Audit) — Pending

- date: <pending>
- author: plan-auditor
- previous_verdict: N/A (initial audit)
- verdict: <pending>
- report: .moai/reports/plan-audit/SPEC-V3R3-CI-AUTONOMY-001-2026-05-09.md
```

---

## Honest Scope Concerns

1. **`/moai plan` skill body 의 BODP gate sub-section 위치 (W7-T02)**: 현재 plan.md skill body 가 어떤 Phase 구조인지 직접 확인 미완 (read-only scope 한계). Phase 3 가 Branch Path + Worktree Path 분기로 구성되어 있다는 가정 하 작업. Phase 2 단계에서 구현자가 실제 skill body 를 읽고 적절한 위치에 BODP gate sub-section 삽입 필요. Mitigation: strategy-wave7.md §5 W7-T02 에서 "공통 sub-section 추출" 방향 제시; 실제 위치는 구현자 판단.

2. **Signal A path overlap heuristic 정밀도 (W7-T01 + W7-R2)**: depends_on 미선언이지만 같은 internal/cli/ 디렉토리 작업 케이스에서 false-positive 가능. Wave 7는 simple substring match 사용; 향후 정밀도 향상은 follow-up SPEC. AC-CIAUT-019 canonical case (depends_on 명시) 는 정확히 처리.

3. **CLI path AskUserQuestion 부재 정적 검사 (W7-T03 + W7-R4)**: import 검사는 Wave 7 단일 검증 기제; lint custom rule 도입은 over-engineering. Phase 2 코드 리뷰 + PR description 체크리스트 항목 강조.

4. **Audit trail concurrent write race (W7-R5)**: orchestrator 단일 세션 가정. 두 세션이 동시에 `moai worktree new SPEC-X` 호출 시 file overwrite race 가능. Wave 7 ship 안 함; risk 수용. Follow-up SPEC 에서 `internal/lockedfile/` 패턴 (lessons #11) 적용 검토.

5. **Reminder 빈도 (W7-R6)**: 매 `moai status` 호출마다 reminder 출력 noise 가능. `MOAI_NO_BODP_REMINDER=1` env var 1회 비활성화만 ship. 1회 표시 후 자동 mute pattern 은 over-engineering 회피.

6. **`/moai plan --branch` skill body 의 manager-git delegation 메커니즘 (W7-T02)**: manager-git agent 가 `base=` parameter 를 어떻게 수신하는지 (agent prompt 형식, parameter 전달 방식) 직접 검증 미완. Phase 2 단계에서 manager-git agent body 검토 후 호환 가능한 prompt 형식으로 BODP 결과 전달. Mitigation: AskUserQuestion 응답 → manager-git delegation prompt 에 "base branch: <chosen>" 명시 텍스트 포함 (자연어 prompt).

7. **`internal/cli/worktree/new.go` 기존 구조 (W7-T03)**: 본 파일 read-only scope 외, 현재 구조 (cobra.Command vs flag-based) 미확인. Wave 7 strategy 는 cobra 가정 (다른 moai-adk-go CLI 와 일관). Phase 2 구현 시 실제 구조에 맞게 조정.

8. **`.moai/branches/decisions/.gitkeep` Template-First mirror 필요성 (W7-T04)**: `.gitkeep` 은 사용자 프로젝트 환경에서 BODP audit trail 디렉토리 존재 보장 위해 필요. Wave 7 는 보수적으로 mirror 포함 (`internal/template/templates/.moai/branches/decisions/.gitkeep`); plan-auditor 검토 후 결정.

9. **Wave 5/6 sub-agent 1M context inheritance error precedent**: Phase 2 manager-tdd 위임 시 동일 error 가능. Mitigation: Wave 7 Go code volume estimate ~600-900 LOC across 4 source files (`relatedness.go` ~250, `audit_trail.go` ~150, `worktree/new.go` extend ~100, `status.go` extend ~60) + 4-6 test files (~600-900 LOC) — main-session 직접 구현 fallback 충분히 manageable. Phase 2 plan에 fallback path 명시.

10. **Final Wave closure dependencies**: 본 Wave 7 머지 = SPEC-V3R3-CI-AUTONOMY-001 closure 직전. CHANGELOG, AC-CIAUT-020 (manual validation), retrospective 작성은 본 Wave scope 외 (Wave 7 PR 머지 후 별도 SPEC closure tasks). progress.md 업데이트는 본 Wave 포함.

No hard blockers identified. Wave 7 ready for Phase 1.5 (plan-auditor) upon strategy + tasks approval, then Phase 2 (manager-tdd 위임 또는 main-session 직접 구현 fallback).

---

Version: 0.1.0
Status: pending Phase 1.5 PLAN AUDIT (plan-auditor initial audit)
Last Updated: 2026-05-09
