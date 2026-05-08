---
spec: SPEC-V3R3-CI-AUTONOMY-001
wave: 5
version: 0.1.0
status: draft
created_at: 2026-05-08
updated_at: 2026-05-08
author: manager-strategy
---

# Wave 5 Execution Strategy (Phase 1 Output)

> Audit trail. manager-strategy output for Wave 5 of SPEC-V3R3-CI-AUTONOMY-001 — T6 Worktree State Guard.
> Generated: 2026-05-08. Methodology: TDD (Go unit tests for state-guard primitives + manual replay for orchestrator-driven invocation pattern).
> Base: `origin/main 311f27a2a` (Wave 4 PR #791 merge baseline; Wave 5 는 plan.md §7 에 따라 독립적이며 Wave 1-4 산출물과 무관).

---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-08 | manager-strategy | Initial draft. Resolved OQ1-OQ6 inline (§7 Decisions). Reframed plan.md `internal/orchestrator/` 신규 패키지 가정을 기존 `internal/worktree/` (신규 lib) + `internal/cli/worktree/` (CLI 확장) 통합 패턴으로 정정 (OQ1). claude-code-guide 가 dev-project 와 user-template 양쪽에 부재함을 발견 → "extend" wording 을 "create new agent" 로 정정 (§12 Honest Concerns C-1). AC mapping: AC-CIAUT-014 + AC-CIAUT-015. |

---

## 1. Wave 5 Goal

Wave 5 는 5-PR sweep (2026-05-05) 에서 P7 으로 식별된 R4 (`Agent(isolation: "worktree")` 회귀 — sub-agent 4/4 모두 main workspace 변경, `worktreePath: {}` empty 응답) 를 차단한다. 회귀 자체는 Anthropic upstream 에서 fix 해야 하는 외부 버그이므로 본 Wave 는 **방어 레이어 (guard layer)** 만 구현한다. 핵심 산출물 3가지: (1) state snapshot/diff/restore primitives 를 pure-Go 라이브러리로 캡슐화, (2) 오케스트레이터가 Bash 로 호출 가능한 `moai worktree` CLI 서브커맨드 (`snapshot`, `verify`, `restore`) 노출, (3) `worktree-state-guard.md` 규칙 + 신규 `claude-code-guide` 에이전트 (Anthropic upstream 회귀 조사) 정의. AC-CIAUT-014 (divergence detection) 와 AC-CIAUT-015 (empty worktreePath suspect handling) 가 본 Wave 외에서는 충족 불가능하다.

**핵심 비-자명 사실 (Phase 2 위임 시 manager-tdd 가 인지해야 함)**: orchestrator (Claude Code runtime) 는 Go 코드가 아니다. 따라서 Go 라이브러리는 `Agent()` 호출을 가로챌 수 없으며, snapshot/verify 패턴은 **orchestrator 가 매 Agent() 호출 직전·직후에 Bash 로 `moai worktree snapshot` / `moai worktree verify` 를 명시적으로 invoke** 하는 documentation-driven 패턴으로만 동작한다. Go 라이브러리는 primitive 만 제공하고, "언제 호출할 것인가" 는 `worktree-state-guard.md` rule 문서가 강제한다.

---

## 2. Architecture

### 2.1 Package Structure (resolves OQ1)

```
internal/
├── worktree/                            # NEW package (Wave 5)
│   ├── state_guard.go                   # Snapshot, Diff, Restore primitives (pure-Go)
│   ├── state_guard_test.go              # 4 unit-test cases (no diff / untracked added / untracked removed / branch changed)
│   ├── snapshot_io.go                   # JSON serialization + .moai/state/ persistence
│   ├── divergence_log.go                # .moai/reports/worktree-guard/<DATE>.md writer (markdown)
│   └── doc.go                           # Package doc with public API contract
└── cli/worktree/                        # EXISTING package (extend)
    ├── guard.go                         # NEW — `moai worktree snapshot|verify|restore` subcommands
    └── guard_test.go                    # NEW — CLI integration test (cobra-driven)
```

### 2.2 Data Model

```go
// Snapshot captures working tree state at a point in time.
type Snapshot struct {
    SchemaVersion  string    `json:"schema_version"`     // "1.0.0"
    CapturedAt     time.Time `json:"captured_at"`
    HeadSHA        string    `json:"head_sha"`           // git rev-parse HEAD
    Branch         string    `json:"branch"`             // git rev-parse --abbrev-ref HEAD
    PorcelainLines []string  `json:"porcelain_lines"`    // git status --porcelain (sorted, deterministic)
    UntrackedSpecs []string  `json:"untracked_specs"`    // sorted relative paths under .moai/specs/
    SnapshotID     string    `json:"snapshot_id"`        // unique identifier (UUID or timestamp-based)
}

// Divergence categorizes how post-state differs from pre-state.
type Divergence struct {
    HeadChanged       bool     `json:"head_changed"`         // SHA mismatch
    BranchChanged     bool     `json:"branch_changed"`       // branch name mismatch
    UntrackedAdded    []string `json:"untracked_added"`      // paths present post but not pre
    UntrackedRemoved  []string `json:"untracked_removed"`    // paths present pre but not post
    PorcelainDelta    []string `json:"porcelain_delta"`      // unified diff of porcelain lines
}

// IsDivergent returns true when any dimension differs (binary detection per OQ4).
func (d Divergence) IsDivergent() bool { ... }

// SuspectFlag indicates the orchestrator's response showed an empty worktreePath despite isolation request.
type SuspectFlag struct {
    SnapshotID  string `json:"snapshot_id"`
    AgentName   string `json:"agent_name"`
    Reason      string `json:"reason"`              // "empty_worktree_path" | "post_state_divergence"
    DetectedAt  time.Time `json:"detected_at"`
    PushBlocked bool   `json:"push_blocked"`        // true → require manual override
}
```

### 2.3 Orchestrator Invocation Pattern

```
orchestrator (Claude Code runtime, prose)
   │
   │ before Agent(isolation: "worktree") call:
   ├── Bash: moai worktree snapshot --out .moai/state/worktree-snapshot-<id>.json
   │
   │ Agent(...) executes
   │
   │ after Agent() returns:
   ├── Bash: moai worktree verify --snapshot .moai/state/worktree-snapshot-<id>.json --agent-response <response.json>
   │            └─ exit 0 = no divergence + non-empty worktreePath
   │               exit 1 = divergence detected → orchestrator runs AskUserQuestion(restore/accept/abort)
   │               exit 2 = suspect flag (empty worktreePath) → orchestrator warns + blocks subsequent push
   │
   │ on divergence + user choice "restore":
   └── Bash: moai worktree restore --snapshot <path>
              └─ git restore --source=<HEAD-SHA> --staged --worktree :/ + untracked path notification
```

**Key constraint**: AskUserQuestion is orchestrator-only HARD (agent-common-protocol §User Interaction Boundary). Therefore the Go CLI returns structured exit codes + JSON-on-stdout reports; the orchestrator (in prose / `Skill("moai")` body) translates these into the AskUserQuestion call.

---

## 3. Approach §1 — Snapshot Capture (W5-T01)

**Strategy**: pure-Go invocation of `git` subprocess via `os/exec.Command`, no `go-git` dependency.

설계:
- `git rev-parse HEAD` (get current SHA)
- `git rev-parse --abbrev-ref HEAD` (get branch name; "HEAD" 인 경우 detached state)
- `git status --porcelain` (whole tree status; sorted lines for deterministic output)
- `git ls-files --others --exclude-standard .moai/specs/` (untracked under specs/, respects .gitignore)
- All results captured into Snapshot struct, JSON-serialized to `.moai/state/worktree-snapshot-<UUID>.json`

이유 (vs `go-git` library):
- `go-git` 추가 의존성 + binary size 증가 + git 동작 미묘한 차이 (예: index lock handling)
- `os/exec` + `git` 바이너리는 모든 dev 환경에서 가용 + Wave 1 의 ci-mirror 패턴 계승
- 테스트 시 `t.TempDir()` 에 `git init` 후 fixture 조작이 더 직관적

**보호 장치**:
- `git rev-parse HEAD` 실패 (예: 빈 repo) → Snapshot 의 `HeadSHA = ""` + warning log; verify 단계에서 명시적 처리
- 결과 라인 sorted (Go `sort.Strings`) + UTF-8 정규화로 OS 간 일관성 보장
- 30s timeout 으로 git 명령 hang 방지 (`exec.CommandContext` with `context.WithTimeout`)

---

## 4. Approach §2 — State Diff Logic (W5-T02)

**Strategy**: 4-dimension binary comparison (resolves OQ4).

설계:
- **HEAD diff**: `pre.HeadSHA != post.HeadSHA` → `Divergence.HeadChanged = true`
- **Branch diff**: `pre.Branch != post.Branch` → `Divergence.BranchChanged = true`
- **Untracked added**: `set(post.UntrackedSpecs) - set(pre.UntrackedSpecs)` → `Divergence.UntrackedAdded`
- **Untracked removed**: `set(pre.UntrackedSpecs) - set(post.UntrackedSpecs)` → `Divergence.UntrackedRemoved`
- **Porcelain delta**: line-by-line unified diff (단순 string 비교, 변경된 라인 집합)

이유 (vs configurable threshold):
- Plan.md REQ-CIAUT-033 wording 이 binary ("**If** post-call state diverges...") — 임계값 도입 시 spec drift
- false positive 는 **scope** 으로 해결 (OQ3): `.moai/specs/` 아래만, `.gitignore` 매칭 항목 제외, `.moai/reports/evaluator-active/` 등 Wave 외 untracked 영역 비교 제외
- 임계값 (예: untracked 5개 이상 시 critical) 은 Wave 5 scope 외 — Phase 2 또는 follow-up SPEC 에서 `--threshold` flag 추가 검토

**보호 장치**:
- `.gitignore` 매칭 untracked 자동 제외 (`git ls-files --others --exclude-standard` 가 이미 처리)
- 추가 exclusion list (`.moai/reports/**`, `.moai/cache/**`, `.moai/logs/**`, `.moai/state/**`) 를 const 로 추출 (CLAUDE.local.md §14 no-hardcoding)
- diff 결과 stable ordering (Go `sort` 사용) → snapshot regeneration 시 false-positive 차이 0

---

## 5. Approach §3 — Orchestrator Wrapper / CLI Surface (W5-T03)

**Strategy**: `moai worktree` 기존 CLI 에 `snapshot`, `verify`, `restore` 3 subcommand 추가 (cobra pattern, `internal/cli/worktree/guard.go`).

설계:

```
moai worktree snapshot [--out <path>] [--agent-name <name>]
   - .moai/state/worktree-snapshot-<UUID>.json 생성 (또는 --out 지정 path)
   - stdout: snapshot ID + path
   - exit 0 정상

moai worktree verify --snapshot <path> [--agent-response <json>]
   - <path> 의 snapshot 을 pre-state 로 로드
   - 현재 working tree 를 post-state 로 capture
   - Divergence + SuspectFlag 계산
   - stdout: JSON report (orchestrator 가 parse)
   - exit 0 = clean (no divergence + non-empty worktreePath)
   - exit 1 = divergence detected
   - exit 2 = suspect (empty worktreePath in agent-response)
   - exit 3 = both divergence + suspect

moai worktree restore --snapshot <path> [--dry-run]
   - <path> 의 snapshot 의 HEAD SHA 로 git restore
   - 명령: git restore --source=<HeadSHA> --staged --worktree :/
   - untracked file paths: stdout 에 list 출력 + "manual recreation required" 명시
   - --dry-run 시 실행 없이 명령만 출력
   - exit 0 정상
```

이유 (vs in-process Go API only):
- Orchestrator 는 Go process 가 아니므로 Go API 직접 호출 불가 — Bash 가 유일한 인터페이스
- `internal/cli/worktree/` 기존 cobra 트리에 추가 → user-facing `moai worktree` 명령 확장 + dev-internal 사용 양쪽 지원
- exit code 기반 통신 → orchestrator 의 prose 단순 conditional ("if exit 1, run AskUserQuestion")

**대안 (rejected)**:
- 신규 `moai guard` top-level 명령 → 새로운 사용자 surface 도입 (CLAUDE.local.md §16 자가 점검 4번 "수량 기반 트리거"), 기존 `moai worktree` 와 의미 분리 모호
- 별도 `internal/orchestrator/` 패키지 (plan.md 원안) → 새 패키지 + 새 import 경로 + cobra 트리 별도 구성 필요; `internal/cli/worktree/` 와의 응집도 손상

---

## 6. Approach §4 — Divergence Logging + Empty Worktree Suspect (W5-T04, W5-T05)

**Strategy**: Markdown structured report + JSON sidecar.

설계 (W5-T04 divergence 로깅):
- 위치: `.moai/reports/worktree-guard/<YYYY-MM-DD>.md` (per-day rolling)
- 동일 날짜 다회 호출 시 append (Wave 4 의 `release-drafter-cleanup-log.md` 패턴 계승)
- 항목 schema:

```markdown
## <ISO-8601 timestamp> — Divergence detected

- Snapshot ID: <id>
- Agent: <agent-name (if known)>
- Divergence dimensions:
  - HeadChanged: pre=<sha> → post=<sha>
  - BranchChanged: pre=<name> → post=<name>
  - UntrackedAdded: [<path>, ...]
  - UntrackedRemoved: [<path>, ...]
- Report sidecar: .moai/reports/worktree-guard/<DATE>-<id>.json
```

설계 (W5-T05 empty worktreePath suspect):
- Agent response JSON 에 `worktreePath` 필드 검사 (`{}` 또는 missing 또는 empty string → suspect)
- Suspect 시 `.moai/state/worktree-suspect-<id>.flag` 파일 생성 (subsequent `moai sync` 또는 push 명령이 이 파일 존재 시 차단)
- orchestrator 가 push 직전에 `moai worktree status` (기존 명령) 또는 신규 helper 로 suspect flag 확인 (Phase 2 가 push pre-check 와이어업 결정)

이유 (markdown vs JSON-only):
- Markdown 은 사용자 직접 열람 (debug 시) + grep-friendly
- JSON sidecar 는 자동화 (예: 후속 SPEC 에서 dashboard 생성) 용
- 양자 병행 → 사람과 도구 모두 친화적

**보호 장치**:
- `.moai/reports/worktree-guard/` 디렉터리는 `.gitignore` 에 추가 (.moai/reports/ 이미 ignore 되어 있을 가능성 — Phase 2 가 검증)
- suspect flag 파일 stale 처리 (1시간 이상 된 flag 는 warning 만, hard block 안 함; orchestrator restart 시 recovery)

---

## 7. Decisions (OQ1-OQ6 Resolutions)

### OQ1: Package Location

**Decision**: 옵션 (b) 변형 — `internal/worktree/state_guard.go` (신규 lib) + `internal/cli/worktree/guard.go` (기존 CLI 확장).

**Rationale**:
- `internal/orchestrator/` 패키지가 dev project 에 부재 — plan.md 의 원안은 새 패키지 신설 가정
- `internal/cli/worktree/` 는 이미 `new.go`, `status.go`, `list.go` 등 cobra 서브커맨드 트리 보유 → 일관된 사용자 surface
- `Agent()` 호출은 Claude Code runtime 의 동작 (Go 코드 부재) → orchestrator 가 Bash CLI 로 primitive invoke 하는 패턴이 사실상 유일한 통신 경로
- pure-Go primitives 를 `internal/worktree/` 새 라이브러리로 분리 → CLI 와 라이브러리 사용자 (예: 향후 다른 Go 코드) 양쪽 지원
- 새 패키지 이름이 `internal/cli/worktree/` 와 동일 last-segment 이지만 import path 가 다름 (`internal/worktree` vs `internal/cli/worktree`) → 충돌 없음

**Action**: tasks-wave5.md W5-T01 ~ W5-T06 의 file path 를 `internal/worktree/state_guard.go` (lib) + `internal/cli/worktree/guard.go` (CLI) 로 명시.

### OQ2: Snapshot Serialization

**Decision**: 옵션 (a) — 디스크 기반 JSON, `.moai/state/worktree-snapshot-<id>.json`.

**Rationale**:
- orchestrator 는 Claude Code runtime (Go process 아님) → Go 메모리 공유 불가능
- in-memory only 옵션 은 단일 프로세스 호출에서만 가치 있음 (snapshot + verify 가 별개 프로세스)
- 하이브리드 옵션 (primary disk + cache) 은 over-engineering — 현재 wave 의 사용 빈도 (Agent() 호출 인접) 에서 캐시 필요 없음
- JSON schema 는 forward-compatible (`schema_version` 필드 포함) → 향후 변경 시 graceful migration

**Action**: W5-T01 의 file IO 부분을 `internal/worktree/snapshot_io.go` 에 분리; schema_version "1.0.0" 명시.

### OQ3: Untracked File Scope

**Decision**: 옵션 (a) — `.moai/specs/` only 우선 + 명시적 exclusion list.

**Rationale**:
- spec.md REQ-CIAUT-031 wording 이 strict ("untracked files under `.moai/specs/`") — 확장 시 spec drift
- 더 넓은 `.moai/` 는 false positive 위험 (`.moai/reports/evaluator-active/` 등 정상 untracked)
- 전체 working tree 는 user-installed dev tool 산출물 (예: IDE 캐시) 까지 포함 → 노이즈 큼
- Plan.md §11 R-CIAUT-5 risk 에 명시된 "`.gitignore` 항목 비교 제외" mitigation 과 정합

**Exclusion list** (const 로 추출, CLAUDE.local.md §14):
```go
const (
    untrackedScope = ".moai/specs/"
)

var defaultExclusions = []string{
    ".moai/reports/",
    ".moai/cache/",
    ".moai/logs/",
    ".moai/state/",  // self-protection: snapshot files 자체가 noise 되지 않도록
}
```

**Action**: W5-T01 의 untracked enumeration 함수가 `untrackedScope` const 사용 + `git ls-files --others --exclude-standard <scope>` 호출. exclusions 는 `.gitignore` 가 이미 처리 가정 (추가 검증 없음 — Phase 2 가 .gitignore 정합성 확인 필요).

### OQ4: Divergence Threshold

**Decision**: 옵션 (a) — binary detection (any difference = divergence).

**Rationale**:
- Plan.md + spec.md REQ-CIAUT-033 wording 이 binary
- configurable threshold 는 false-positive 발생 시 사용자 워크플로우 차단 → 에스컬레이션 확률 ↑
- false-positive 완화는 OQ3 의 scope 제한 + exclusions 로 충분
- 임계값은 Phase 2 follow-up 또는 별도 SPEC (Wave 5 외) 에서 `--threshold-untracked-added <N>` 같은 opt-in flag 로 추가 검토 가능

**Action**: W5-T02 의 `Divergence.IsDivergent()` 메서드는 4 boolean OR + 2 slice non-empty check.

### OQ5: claude-code-guide Invocation Trigger

**Decision**: 옵션 (a) — 자동 trigger on first `worktreePath: {}` suspect detection within session.

**Rationale**:
- 실시간 trigger 가 context freshest (orchestrator 가 방금 받은 agent response 와 결합)
- Wave 5 SPEC completion 1-shot 은 정보 손실 (어떤 agent invocation 이 회귀였는지 불명확)
- user-opt-in 은 friction 추가 + first-time 사용자가 trigger 방법 모름
- 단, 동일 session 내 후속 suspect 는 re-trigger 안 함 (single bug report per session — claude-code-guide 가 동일 정보 중복 수집 방지)
- session 종료 후 다음 session 에서 새 suspect 발견 시 다시 trigger (per-session bookkeeping via `.moai/state/upstream-investigation-<session-id>.flag`)

**Action**: W5-T07 의 claude-code-guide trigger 정의를 `worktree-state-guard.md` rule 에 명시; 자동 invocation 로직은 orchestrator (Skill body) 가 책임지고, Go 라이브러리는 suspect flag 만 노출.

### OQ6: Template-First Mirror Scope

**Decision**: 옵션 (a) — both new docs MUST live in `internal/template/templates/` per CLAUDE.local.md §2.

**Rationale**:
- 본 wave 의 산출물 중 `.claude/rules/moai/workflow/worktree-state-guard.md` 와 `.claude/agents/moai/claude-code-guide.md` 는 user-facing rule + agent → user project 에 배포되는 것이 정상 (사용자도 같은 worktree guard 가 필요)
- Wave 4 의 `.github/workflows/*` 와 다름 (Wave 4 는 dev-only CI)
- Go 코드 (`internal/worktree/`, `internal/cli/worktree/`) 는 `internal/` 에 위치 — Go 패키지는 binary 에 컴파일되어 user 에게 배포 (template 미러 불필요)
- placeholder report (`.moai/reports/upstream/agent-isolation-regression.md`) 는 runtime artifact (사용자 환경마다 다름) → template 미러 안 함

**File ownership matrix** (§4 참조):

| File | Source path | Mirror required? | Rationale |
|------|-------------|------------------|-----------|
| `worktree-state-guard.md` rule | `internal/template/templates/.claude/rules/moai/workflow/worktree-state-guard.md` | YES (primary) — `.claude/...` 사본은 `make build` 로 derive | user-facing rule; user project 가 worktree guard 활용 |
| `claude-code-guide.md` agent | `internal/template/templates/.claude/agents/moai/claude-code-guide.md` | YES (primary) — `.claude/...` 사본은 `make build` 로 derive | user-facing agent; user project 도 동일 회귀 발생 가능 |
| `internal/worktree/*.go` | dev project | NO (Go 패키지는 binary 컴파일) | Go 코드는 template 미러 대상 아님 |
| `internal/cli/worktree/guard.go` | dev project | NO (CLI는 binary 일부) | 동일 |
| `.moai/reports/upstream/agent-isolation-regression.md` | dev project (placeholder) | NO (runtime artifact) | 사용자 환경마다 다름; placeholder 만 dev 에 존재 |

**Action**: tasks-wave5.md W5-T07 + W5-T08 file path 를 template-first source 로 명시 + Phase 2 위임 시 "make build 후 .claude/ 사본 자동 갱신 검증" 강제.

---

## 8. File Ownership Matrix

### Implementer Scope (write access)

```
internal/worktree/state_guard.go                                          # NEW (W5-T01, W5-T02)
internal/worktree/snapshot_io.go                                          # NEW (W5-T01 sub-task: JSON IO)
internal/worktree/divergence_log.go                                       # NEW (W5-T04: markdown writer)
internal/worktree/doc.go                                                  # NEW (package doc)
internal/cli/worktree/guard.go                                            # NEW (W5-T03, W5-T05, W5-T06: subcommands)
internal/cli/worktree/root.go                                             # EXTEND (register guard subcommand) — read existing first
internal/template/templates/.claude/rules/moai/workflow/worktree-state-guard.md  # NEW (W5-T08)
internal/template/templates/.claude/agents/moai/claude-code-guide.md      # NEW (W5-T07; NOT extension — see C-1)
.moai/reports/upstream/agent-isolation-regression.md                      # NEW placeholder (W5-T07)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave5.md                   # this file
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave5.md                      # companion
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/progress.md                         # extend (Phase 1 + 1.5 entries)
```

### Tester Scope (test 파일만 write)

```
internal/worktree/state_guard_test.go                                     # NEW (4 cases: no-diff / untracked-added / untracked-removed / branch-changed)
internal/worktree/snapshot_io_test.go                                     # NEW (JSON roundtrip)
internal/cli/worktree/guard_test.go                                       # NEW (cobra integration: snapshot/verify/restore exit codes)
```

### Read-Only Scope (cross-Wave consumer / SSoT source)

```
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/spec.md                             # frozen for Wave 5
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/plan.md                             # frozen for Wave 5
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/acceptance.md                       # frozen for Wave 5
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave[1-4].md               # prior waves (audit trail)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave[1-4].md                  # prior waves
internal/cli/worktree/new.go, status.go, list.go, ...                     # existing cobra patterns (read-only reference)
.claude/rules/moai/workflow/worktree-integration.md                       # existing worktree rule (worktree-state-guard.md 가 cross-reference)
.claude/rules/moai/core/agent-common-protocol.md                          # AskUserQuestion HARD rule (orchestrator-only)
.claude/agents/moai/expert-debug.md                                       # agent frontmatter pattern (claude-code-guide 작성 시 참조)
.claude/agents/moai/researcher.md                                         # agent frontmatter pattern (researcher 와 유사한 read+write 권한)
CLAUDE.md, CLAUDE.local.md                                                # project rules
```

### Implicit Read Access (모든 task)

- `.claude/rules/moai/**` (auto-loaded rules)
- 모든 Wave 1-4 산출물 (read-only consumer)

---

## 9. Implementation Sequence (W5-T01 → W5-T08 Dependency Graph)

```
W5-T01 (snapshot capture: git rev-parse + porcelain + ls-files)
    │
    ├─→ W5-T02 (diff function: 4-dimension binary)
    │       │
    │       ├─→ W5-T03 (orchestrator wrapper / CLI subcommands snapshot+verify)
    │       │       │
    │       │       ├─→ W5-T04 (divergence logging + markdown writer)
    │       │       │
    │       │       ├─→ W5-T05 (empty worktreePath suspect detection + flag file)
    │       │       │
    │       │       └─→ W5-T06 (state restore: git restore + untracked notification)
    │       │
    │       └── (parallel) W5-T08 (rule doc — independent of CLI implementation)
    │
    └── (parallel) W5-T07 (claude-code-guide NEW agent — independent of guard primitives)
```

**Sequential dependencies**: T01 → T02 → T03 (T03 imports T01+T02 primitives). T04, T05, T06 all depend on T03 (extend CLI) but can run in any order after T03.

**Parallel opportunities**: T07 (agent file) and T08 (rule doc) are pure-markdown deliverables independent of Go code; can run in parallel with T01-T06.

**Suggested commit cadence (Conventional Commits)**:

1. `test(worktree): W5-T01 RED — Snapshot capture cases (test file only)`
2. `feat(worktree): W5-T01 implement Snapshot + git rev-parse/porcelain/ls-files`
3. `test(worktree): W5-T02 RED — Diff cases (no-diff/added/removed/branch-changed)`
4. `feat(worktree): W5-T02 implement Diff + IsDivergent`
5. `feat(cli): W5-T03 add moai worktree snapshot|verify subcommands`
6. `feat(worktree): W5-T04 add divergence markdown logger`
7. `feat(cli): W5-T05 detect empty worktreePath + suspect flag file`
8. `feat(cli): W5-T06 add moai worktree restore subcommand`
9. `chore(agents): W5-T07 add claude-code-guide agent + upstream report placeholder`
10. `chore(rules): W5-T08 add worktree-state-guard.md rule`

각 commit 은 🗿 MoAI co-author trailer 포함.

---

## 10. Risks & Mitigations

| ID | Risk | Mitigation |
|----|------|-----------|
| W5-R1 | `git ls-files --others --exclude-standard` 가 거대한 untracked tree 에서 느림 (>1s) | scope 을 `.moai/specs/` 로 한정 (OQ3); Phase 2 perf 측정 시 5s 초과 시 `--max-depth` flag 도입 검토 |
| W5-R2 | `git rev-parse HEAD` 가 detached HEAD / empty repo 에서 실패 → snapshot 실패 | `HeadSHA = ""` + warning log 로 graceful degradation; verify 단계에서 명시적 처리 |
| W5-R3 | `.moai/reports/evaluator-active/` 같은 정상 untracked 가 OQ3 scope 외이지만 false-positive 가능 | OQ3 scope `.moai/specs/` 로 한정 → reports/ 는 자동 제외; W5-T02 가 scope 외 path 비교 안 함 |
| W5-R4 | `git restore --source=<sha> --staged --worktree :/` 가 untracked 파일 보존 | 의도된 동작 (untracked 는 git restore 영향 안 받음); W5-T06 가 stdout 에 untracked path list + "manual recreation required" 출력 |
| W5-R5 | `worktreePath: {}` empty 응답이 정상 케이스 (예: Anthropic API 변경) 와 회귀를 구분 못 함 | suspect flag 는 "warn" level 로 기본 처리 (push block 은 추가 confirm 필요); Phase 2 가 sample agent response 수집 후 known-good schema 정의 |
| W5-R6 | claude-code-guide 가 Anthropic 외부 API 호출 (회귀 보고서 작성 시 Anthropic console scrape) → 외부 의존성 | claude-code-guide 는 **로컬 분석 + 추측 + 사용자 보고만** 수행; Anthropic upstream 은 사용자가 수동 보고; 본 wave 는 placeholder report 만 생성 |
| W5-R7 | Wave 4 lessons (manager-tdd/expert-devops sub-agent 1M context inheritance error) 재발 — Phase 2 위임 실패 시 implementation 막힘 | Phase 2 plan 에 fallback 명시: "if sub-agent delegation fails with context error, proceed with direct main-session implementation"; 본 wave 의 Go 코드 분량 (~600 LOC) 은 main-session 직접 구현 가능 범위 |
| W5-R8 | `internal/cli/worktree/root.go` 에 `guard` subcommand 등록 시 기존 cobra 트리 회귀 | Phase 2 가 etxisting `new`, `status`, `list` 등록 패턴 inspect 후 동일 패턴 적용; 회귀 시 `internal/cli/worktree/root_test.go` (existing) 가 즉시 fail |

---

## 11. Test Plan

### 11.1 Unit Tests (`internal/worktree/state_guard_test.go`)

**4 mandatory cases** (per plan.md §7):

```go
func TestSnapshot_NoDiff(t *testing.T) {
    // Setup: t.TempDir() + git init + commit + .moai/specs/SPEC-X/file.md
    // Capture: pre := Snapshot(); post := Snapshot()
    // Assert: Diff(pre, post).IsDivergent() == false
}

func TestSnapshot_UntrackedAdded(t *testing.T) {
    // Setup: as above
    // Capture: pre := Snapshot(); create new file under .moai/specs/SPEC-X/; post := Snapshot()
    // Assert: Divergence.UntrackedAdded contains new file path
}

func TestSnapshot_UntrackedRemoved(t *testing.T) {
    // Setup: as above with pre-existing untracked under .moai/specs/SPEC-X/
    // Capture: pre := Snapshot(); rm new untracked; post := Snapshot()
    // Assert: Divergence.UntrackedRemoved contains removed path
}

func TestSnapshot_BranchChanged(t *testing.T) {
    // Setup: as above
    // Capture: pre := Snapshot() on branch-A; git checkout -B branch-B; post := Snapshot()
    // Assert: Divergence.BranchChanged == true; pre.Branch != post.Branch
}
```

**Test isolation** (CLAUDE.local.md §6):
- 모든 test 는 `t.TempDir()` 사용 (auto-cleanup)
- `exec.Command("git", "init", tmpDir)` 로 isolated repo
- host project 의 git state 와 절대 격리 (호스트 repo HEAD 변경 금지)

### 11.2 IO Tests (`internal/worktree/snapshot_io_test.go`)

- JSON roundtrip: `Marshal(snapshot)` → `Unmarshal` → 동일 struct
- Schema version forward compatibility: unknown 필드 무시

### 11.3 CLI Integration Tests (`internal/cli/worktree/guard_test.go`)

- `moai worktree snapshot --out <path>` → exit 0 + 파일 생성 확인
- `moai worktree verify --snapshot <path>` (no diff) → exit 0
- `moai worktree verify --snapshot <path>` (diff) → exit 1
- `moai worktree verify --agent-response <empty-worktreepath>` → exit 2
- `moai worktree restore --snapshot <path> --dry-run` → exit 0 + stdout 에 명령 출력

### 11.4 Manual Replay Scenarios

- **AC-CIAUT-014 replay**: 4회 연속 fake `Agent(isolation:)` 호출 시뮬레이션 (test fixture 가 worktree state 변경) → 1회차에서 alert
- **AC-CIAUT-015 replay**: agent response JSON `{"worktreePath": {}}` fixture 로 verify 호출 → suspect flag 파일 생성 확인 + stdout 에 warning

---

## 12. Acceptance Mapping

| REQ-CIAUT | Wave 5 Task | AC-CIAUT | Verification |
|-----------|-------------|----------|--------------|
| 031 (Pre-call snapshot Event-Driven) | W5-T01 | 014 | unit test TestSnapshot_NoDiff + manual replay |
| 032 (Post-call verify Event-Driven) | W5-T02, W5-T03 | 014 | unit test 4 cases + CLI integration test |
| 033 (Divergence handling Unwanted) | W5-T04 | 014 | manual replay: divergence → markdown log + JSON sidecar |
| 034 (Empty worktreePath suspect Unwanted) | W5-T05 | 015 | CLI integration test (exit 2) + suspect flag file 검증 |
| 035 (claude-code-guide upstream investigation Ubiquitous) | W5-T07 | 015 | agent file YAML frontmatter valid + placeholder report file 존재 |
| 036 (Restore option Optional) | W5-T06 | 014 (restore path) | CLI integration test with --dry-run + manual restore replay |
| (governance) | W5-T08 | 014 + 015 transitive | rule doc grep checks: when-to-snapshot 명시, divergence threshold 명시, escalation path 명시 |

---

## 13. Rollback Plan

본 wave 의 산출물이 main 머지 후 issue 발견 시:

1. **Go 라이브러리 회귀** (`internal/worktree/`): 이전 commit 으로 revert (Wave 5 Phase 3 commit hash). Go 패키지는 cobra 트리에 등록되지 않은 상태에서는 dead code → 회귀 영향 0.
2. **CLI 회귀** (`internal/cli/worktree/guard.go`): `internal/cli/worktree/root.go` 에서 guard subcommand registration 만 제거 (1-line revert). 라이브러리는 보존 → forward 호환.
3. **Rule doc 회귀** (`worktree-state-guard.md`): 단순 revert 가능. orchestrator (Skill body) 가 아직 이 rule 을 reference 하지 않은 상태이므로 영향 0.
4. **Agent 회귀** (`claude-code-guide.md`): 신규 agent 이므로 단순 revert 가능. 기존 agent 동작 영향 0.

**Forward 호환성**:
- snapshot JSON schema 의 `schema_version` 필드 → 향후 변경 시 graceful migration
- CLI exit codes (0/1/2/3) → orchestrator 의 prose conditional 이 unknown exit 시 default-safe (assume divergence)

---

## 14. Open Concerns (NOT resolved — flagged for Phase 2 or follow-up)

- **C-1: claude-code-guide 가 NEW 인지 EXTENSION 인지 plan.md wording 모호**: `git ls -la .claude/agents/moai/ | grep claude-code-guide` 결과 0 hit + `internal/template/templates/.claude/agents/moai/` 비어있음. **NEW agent 작성 필요** (extension 아님). 본 strategy 는 NEW 로 간주하고 W5-T07 task 구체화. Phase 2 가 commit 메시지 wording 명확화 필요 (`feat(agents): add` not `extend`).
- **C-2: orchestrator 측 wiring 은 본 Wave scope 외**: Go primitive + CLI 만 제공; `Skill("moai") body 또는 `/moai run`/`/moai sync` workflow body 에서 실제 invoke 는 별도 SPEC. `worktree-state-guard.md` rule 이 invocation pattern 정의 + 향후 SPEC 이 wiring 구현. 본 wave 머지 후 즉시 활성화 안 됨 (단계적 rollout).
- **C-3: untracked 파일 content snapshot 부재 → W5-T06 restore 한계**: untracked 파일은 git stage 되지 않았으므로 `git restore` 가 복원 불가. snapshot 은 path 만 저장 → user 가 untracked 파일 내용을 수동 복원해야 함. 대안 (full-content snapshot) 은 본 wave scope 비대화. W5-T06 stdout 에 명시적 안내 필수.
- **C-4: false positive from `.moai/state/` self-modification**: snapshot 파일 자체가 `.moai/state/` 에 저장되므로 (OQ2), 동일 디렉터리의 untracked 변경이 noise. mitigation: OQ3 의 untrackedScope 가 `.moai/specs/` 만 → `.moai/state/` 자동 제외. 단, future-proof 를 위해 `defaultExclusions` 에 `.moai/state/` 명시.
- **C-5: AskUserQuestion 호출은 본 wave 산출물 외**: Go CLI 는 exit code + JSON report 만 출력; AskUser 는 orchestrator 책임. 본 wave 의 rule doc 이 orchestrator 측 책임을 명시하지만 실제 AskUser wiring 은 Skill body 변경 필요 → C-2 와 동일 follow-up.
- **C-6: Wave 4 sub-agent 1M context inheritance error 재발 가능성**: Phase 2 에서 manager-tdd 위임 시 동일 error 가능. mitigation (W5-R7): Go 코드 분량 (~600 LOC) 은 main-session 직접 구현 fallback 가능 — Phase 2 가 위임 실패 시 즉시 main-session 진행.
- **C-7: `internal/cli/worktree/root.go` 의 guard subcommand 등록 위치**: 기존 root.go 가 cobra `rootCmd.AddCommand(...)` 패턴 사용 가정. Phase 2 가 root.go 를 read 한 후 동일 패턴 적용. 등록 누락 시 CLI 가 호출 안 됨 — guard_test.go 가 이를 즉시 catch.

---

## 15. TRUST 5 Targets

| Pillar | Target | Verification |
|--------|--------|--------------|
| **Tested** | `internal/worktree/` 패키지 coverage ≥ 85%; `internal/cli/worktree/guard.go` ≥ 80%; 4 mandatory unit cases + JSON roundtrip + 5 CLI integration cases 모두 PASS; 모든 test 가 `t.TempDir()` isolation | `go test -cover ./internal/worktree/...` ≥ 85%; `go test -cover ./internal/cli/worktree/...` ≥ 80%; race detector clean (`go test -race`) |
| **Readable** | godoc 모든 exported type/func 에 doc comment + package doc.go 작성; `golangci-lint run` 0 issue; markdown rule + agent file 의 H2/H3 구조 명확 | `golangci-lint run ./internal/worktree/...`; `markdownlint internal/template/templates/.claude/rules/moai/workflow/worktree-state-guard.md` |
| **Unified** | gofmt + goimports 일관성; cobra subcommand naming `<verb>` (snapshot/verify/restore — 명사 아닌 동사) 일관; rule doc 헤딩 구조가 worktree-integration.md 와 동일 톤 | `gofmt -l ./internal/...` empty; `goimports -l` empty; rule doc cross-reference 검증 |
| **Secured** | snapshot 파일에 secrets 노출 없음 (porcelain 이 untracked 파일 path 만 포함, content 미포함); `git restore` 호출이 사용자 confirm (orchestrator 측 AskUser) 후에만; `.moai/state/` permission 0644 (world-readable 안 함은 over-engineering — local dev tool) | snapshot JSON 검사: 모든 path 만 포함, content 0; restore CLI 가 `--dry-run` 기본 안 함 단 명시적 확인 메시지 출력 |
| **Trackable** | 모든 commit 이 SPEC-V3R3-CI-AUTONOMY-001 W5 reference 포함; Conventional Commits + 🗿 MoAI co-author trailer; divergence log 가 audit trail 보존 (`.moai/reports/worktree-guard/<DATE>.md`) | `git log --grep='SPEC-V3R3-CI-AUTONOMY-001 W5'`; divergence log 파일 schema 검증 |

---

## 16. Per-Wave DoD Checklist

- [ ] 모든 8개 W5 task 완료 (W5-T01 ~ W5-T08)
- [ ] **Template-First mirror 검증** — `worktree-state-guard.md` rule + `claude-code-guide.md` agent 양쪽 `internal/template/templates/.claude/...` 에 source 위치
- [ ] `make build` 후 embedded.go 갱신 + `.claude/rules/moai/workflow/worktree-state-guard.md` + `.claude/agents/moai/claude-code-guide.md` derived 사본 자동 생성
- [ ] `internal/worktree/state_guard.go` + `snapshot_io.go` + `divergence_log.go` + `doc.go` 신규 작성
- [ ] `internal/cli/worktree/guard.go` 신규 작성 + `internal/cli/worktree/root.go` 에 subcommand 등록
- [ ] 4 mandatory unit tests (no-diff / untracked-added / untracked-removed / branch-changed) PASS
- [ ] JSON roundtrip test PASS
- [ ] CLI integration tests (snapshot / verify clean / verify diff / verify suspect / restore dry-run) PASS
- [ ] `go test -race ./...` 통과 (concurrency safety)
- [ ] `go test -cover ./internal/worktree/...` ≥ 85%
- [ ] `golangci-lint run ./internal/...` 0 issue
- [ ] `make ci-local` 통과 (Wave 1 framework 회귀 없음)
- [ ] `claude-code-guide.md` agent YAML frontmatter valid (description + tools + model + permissionMode + memory + skills 필드)
- [ ] `.moai/reports/upstream/agent-isolation-regression.md` placeholder 파일 작성 (frontmatter only, 본문 _TBD_)
- [ ] AC-CIAUT-014 manual replay 통과 (4회 연속 fake Agent invocation 시뮬레이션 → 1회차에서 alert)
- [ ] AC-CIAUT-015 manual replay 통과 (empty worktreePath fixture → suspect flag + warning)
- [ ] `worktree-state-guard.md` 가 worktree-integration.md cross-reference + invocation pattern 명시
- [ ] No release/tag automation 도입
- [ ] No hardcoded paths/URLs/models (CLAUDE.local.md §14)
- [ ] PR labeled with `type:feature`, `priority:P1`, `area:cli` + `area:workflow`
- [ ] Conventional Commits + 🗿 MoAI co-author trailer 모든 commit
- [ ] CHANGELOG.md 에 Wave 5 머지 entry

---

## 17. Out-of-Scope (Wave 5 Exclusions)

- Wave 6 (T7 i18n validator) — 독립적, 별도 Wave
- Wave 7 (T8 BODP) — 독립적, 별도 Wave
- Anthropic upstream `Agent(isolation:)` 회귀 fix — claude-code-guide 가 보고만, 본 wave 는 guard layer 만 (spec.md §2 Out of Scope)
- Orchestrator (Skill body) 측 invocation wiring — 본 wave 는 primitive + rule 만; 실제 `/moai run` 등에서 invoke 는 별도 SPEC (C-2)
- Untracked 파일 content snapshot — paths-only restoration (C-3)
- Configurable divergence threshold — binary detection 만 (OQ4)
- AskUserQuestion 직접 호출 from Go — orchestrator-only HARD; Go 는 exit code + JSON 만 (C-5)
- Multi-platform `git` binary 호환성 검증 — Wave 1 에서 이미 git-bash on Windows 가정 (Wave 5 도 동일)
- 16-language neutrality — Wave 5 는 Go-only orchestrator primitive (사용자 프로젝트 언어와 무관)
- claude-code-guide 의 자동 Anthropic console scrape — 사용자 수동 보고만; 본 wave placeholder 는 정적 markdown
- Concurrency safety for parallel `moai worktree snapshot` invocations — 단일 invocation 가정 (orchestrator 가 sequential)
- Rollback to pre-snapshot state for tracked file deletions — `git restore` 가 처리하지 않는 케이스 (예: 새로 만든 commit 후 reset 필요) 는 wave 5 외

---

## 18. Honest Methodology Note

본 wave 산출물은 (1) Go 코드 (W5-T01 ~ W5-T06) + (2) markdown 문서 (W5-T07 agent, W5-T08 rule) 혼합. quality.yaml `development_mode: tdd` 적용:

- **Go 코드**: 표준 RED → GREEN → REFACTOR 사이클. 각 task 는 test 먼저 (`*_test.go` RED commit) → implementation (GREEN commit) → cleanup (REFACTOR commit, optional).
- **Markdown 문서**: TDD 부적절 — verify-via-grep 로 대체. W5-T07 + W5-T08 는 RED commit 없이 직접 작성 + 검증은 grep 기반 (예: `grep "## " worktree-state-guard.md` 으로 H2 구조 확인).

Phase 2 manager-tdd 위임 시 본 적응 명시 필요. Wave 4 의 verify-via-replay 패턴 (CI/CD config) 과 다르며, Wave 1-3 의 mixed Go+shell 패턴에 더 가깝다.

---

Version: 0.1.0
Status: pending Phase 2 (manager-tdd 위임 대기)
Last Updated: 2026-05-08
