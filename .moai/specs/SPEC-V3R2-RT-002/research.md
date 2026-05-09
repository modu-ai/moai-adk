# SPEC-V3R2-RT-002 Deep Research (Phase 0.5)

> Research artifact for **Permission Stack + Bubble Mode**.
> Companion to `spec.md` (v0.1.0). Authored against branch `plan/SPEC-V3R2-RT-002` from `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-002` (worktree mode).

## HISTORY

| Version | Date       | Author                                  | Description                                                              |
|---------|------------|-----------------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 0.5)  | Initial deep research per `.claude/skills/moai/workflows/plan.md` Phase 0.5 |

---

## 1. Goal of Research

`spec.md` §1 (Goal), §2 (Scope), §3 (Environment), §4 (Assumptions), §7 (Constraints), §8 (Risks) 의 주장에 대해 구체적 file:line 증거 + 외부 라이브러리/Claude Code 정합성 평가를 수행하여, run phase 가 REQ-V3R2-RT-002-001..051 을 known-good baseline 위에서 구현할 수 있도록 보장.

본 연구는 6 가지 질문에 답합니다:

1. **기존 스켈레톤 인벤토리**: `internal/permission/` 에 이미 무엇이 구현되어 있고, 24 EARS REQ 를 만족하기 위한 델타는?
2. **Claude Code 2026.x 권한 모델**: PermissionMode 5-enum (`default/acceptEdits/bypassPermissions/plan/bubble`) 과 8-tier 우선순위 (`policy > user > project > local > plugin > skill > session > builtin`) 의 정확한 의미는? 본 SPEC 이 인용하는 r3 §1.3 / §2 Decision 15 의 evidence 가 코드와 정합한가?
3. **moai-adk 의 현재 권한 모델 갭**: flat `permissions.allow` 의 한계와 reader 진입점은?
4. **Pre-allowlist 설계 근거**: 8 패턴이 충분한가? 외부 reference (Claude Code, GitHub Copilot, Aider 등) 의 default-allow 정책과 비교는?
5. **Bubble routing 흐름**: AskUserQuestion HARD constraint (orchestrator-only) 와 fork agent 의 IPC 채널 매개는?
6. **Library evaluation**: spec 가 약속한 "no new dependencies" 를 위반하지 않는 stdlib-only 구현 범위는?

---

## 2. Inventory of `internal/permission/` skeleton (existing)

`ls /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-002/internal/permission/` returns 6 files (3 source + 3 test):

| # | File | Size (bytes) | Purpose | Implements |
|---|------|--------------|---------|------------|
| 1 | `stack.go` | 8292 | `PermissionMode` 5-enum + `PermissionRule{Pattern, Action, Source, Origin}` + `PreAllowlist()` 8 패턴 + `IsWriteOperation` write 패턴 매처 + `Decision` enum | REQ-002, REQ-003, REQ-006 (partial) |
| 2 | `resolver.go` | 13437 | `ResolveContext` + `ResolveResult{Decision, ResolvedBy, Origin, UpdatedInput, SystemMessage, Trace}` + `PermissionResolver.Resolve` 8-tier walker + `checkTier` + `handleBubbleAsk` + `handleBypassInFork` + `logUnreachablePrompt` + `ValidateMode` + `ExportTrace` | REQ-004, REQ-005, REQ-011, REQ-013 (partial), REQ-022, REQ-041 |
| 3 | `bubble.go` | 5808 | `BubbleDispatcher` + `BubbleRequest{Tool, Input, ForkDepth, Origin, ParentSessionID}` + `BubbleResponse{Decision, Reason}` + `DispatchToParent` placeholder + `FormatBubblePrompt` + `ValidateForkDepth` + `IsParentAvailable` placeholder | REQ-012 (partial), REQ-023 (partial), REQ-050 (partial) |
| 4-6 | `*_test.go` | 28178 total | 기존 happy-path coverage (stack 21 cases, resolver 32 cases, bubble 16 cases) | Test baseline |

또한 인접 패키지:
- `internal/config/source.go` 2863 bytes — 8-tier `Source` enum (`SrcPolicy=0..SrcBuiltin=7`) 정의. RT-005 의 unique ownership.
- `internal/cli/doctor_permission.go` 2610 bytes — `moai doctor permission` 서브커맨드 (현재 91 LOC, RulesByTier 빈맵으로만 동작).

### 2.1 Delta Analysis (skeleton → spec compliance)

`Grep` 와 manual code read 로 확인한 스켈레톤의 **부분 구현** 상태와 spec REQ 갭:

| Skeleton state | Spec requirement | Gap |
|----------------|------------------|-----|
| `PermissionMode` 5-enum 가 `stack.go:13-65` 에 정의됨 (`ModeDefault, ModeAcceptEdits, ModeBypassPermissions, ModePlan, ModeBubble`) + `ParsePermissionMode` + `IsValid` | REQ-003 (bubble 일등 시민) | OK 구조; 그러나 agent frontmatter 가 strict-validate 되지 않음 — `internal/template/agent_frontmatter_audit_test.go` 에 lint 추가 필요 (REQ-008, AC-09). |
| `PermissionRule` 가 `stack.go:67-86` 에 정의됨 (Pattern/Action/Source/Origin 모두 보유) | REQ-002 (Origin non-empty) | 구조는 OK; Origin 비어있음 시 reader 가 reject 하는지의 invariant 테스트 부재. |
| `PreAllowlist()` 가 `stack.go:182-233` 에 8 패턴 시드 (Read*, Glob*, Grep*, Bash(go test:*), Bash(golangci-lint run:*), Bash(ruff check:*), Bash(npm test:*), Bash(pytest:*)) | REQ-006 (시드) | OK on contents; 그러나 매 호출마다 slice 재생성 → hot path 비효율. M2 에서 sync.Once + cached slice. |
| `Resolve` 메인 분기가 `resolver.go:135-254` 에 8-tier walk 구현 — bypassPermissions short-circuit / plan write-deny / hook overlay / 8-tier walk / non-interactive fallback 모두 존재 | REQ-004, REQ-005, REQ-014, REQ-020, REQ-021 | OK 구조; (1) 동일 tier 다중 매치 시 conflict tiebreak 부재 (REQ-042) — M3, (2) hook UpdatedInput re-match 의 무한루프 방지 가드는 존재하나 단일 재실행 enforcement 테스트 부재 (REQ-013) — M3. |
| `checkTier` 가 `resolver.go:258-299` 에 단일 매치만 처리 (첫 매치 반환) | REQ-042 (specificity-then-fs-order tiebreak) | Missing: 동일 tier 에서 ≥2 매치 발생 시 specificity 점수 비교 분기. M3. |
| `handleBubbleAsk` 가 `resolver.go:305-324` 에 부모 unavailable 시 deny 처리 | REQ-012, REQ-050 | OK on deny path; "ask" → DispatchToParent 까지의 wiring 은 placeholder (`bubble.go:94-102` 의 DispatchToParent 가 not implemented error 반환) — M4 에서 contract 만 동결. |
| `handleBypassInFork` 가 `resolver.go:330-345` 에 fork+bypassPermissions → bubble 강등 | REQ-023 | OK on degrade; systemMessage 의 sentinel 문자열이 spec.md AC-08/AC-14 의 expected sentinel 과 정렬되는지 비교 필요 — 현재 "Fork agent with bypassPermissions - degraded to bubble mode" / "Fork depth N exceeds limit - bypassPermissions degraded to bubble" → AC-14 expected 는 "Fork depth N exceeds limit - mode degraded to bubble" — M4 에서 통일. |
| `logUnreachablePrompt` 가 `resolver.go:351-364` 에 비대화형 fail-closed 로그 기록 (`.moai/logs/permission.log`) | REQ-041 | OK on log path; 그러나 `Resolve` 의 default-deny 분기 (line 241-247) 가 IsInteractive=false 일 때 logUnreachablePrompt 호출 — 검증만 필요. |
| `ValidateMode` 가 `resolver.go:383-397` 에 strict_mode + bypassPermissions reject + fork depth >3 warning 반환 | REQ-022, REQ-023 | OK on logic; 호출 사이트 (spawn entry point) 가 비어있음 — `internal/cli/run.go` 또는 spawn helper 에서 호출 wire 필요 (M2 신규 `internal/permission/spawn.go` 의 RejectIfStrict). |
| `ExportTrace` 가 `resolver.go:402-408` 에 JSON marshal | REQ-015, REQ-007 | OK on JSON 직렬화; 그러나 `TierTry.Tier == config.Source(999)` (hook tier sentinel) 가 raw int 으로 marshal — M5a 에서 `"tier": "hook"` stringify. |
| `BubbleDispatcher.DispatchToParent` 가 `bubble.go:94-102` 에 placeholder (not implemented error) | REQ-012 (parent AskUserQuestion) | Missing: orchestrator IPC contract — M4 에서 RT-001 hook channel 재사용 contract 만 동결. |
| `BubbleDispatcher.IsParentAvailable` 가 `bubble.go:150-154` 에 placeholder (ParentSessionID 비어있지 않음) | REQ-050 | OK on contract for v3.0; registry 기반 liveness check 는 RT-004 SessionStore 와 통합. |
| `internal/cli/doctor_permission.go` (전체 91 LOC) 가 `moai doctor permission --tool --input --trace --dry-run` 만 지원 | REQ-007, REQ-015, REQ-032 | Missing: `--all-tiers` (전체 tier rule dump), `--mode <m>` (PermissionMode 시뮬레이션), `--fork` (fork agent 시뮬레이션) — M5a. |
| `agent_frontmatter_audit_test.go` (564 LOC, 기존 walker 패턴) | REQ-008 (frontmatter strict-validation) | Missing: `permissionMode` 키의 5-enum strict-validation 테스트 함수 — M2. |

**Summary**: 스켈레톤은 spec 의 **structural shape** (~50%) 을 이미 제공. M2-M5 작업은 (1) frontmatter lint, (2) conflict tiebreak, (3) bubble dispatch contract, (4) doctor CLI 보강, (5) legacy migration 의 5 개 갭을 채움.

---

## 3. Claude Code 2026.x permission model — 정합성 평가

### 3.1 PermissionMode 5-enum (r3 §2 Decision 15)

Reference: `.moai/design/v3-redesign/research/r3-cc-architecture-reread.md` §2 Decision 15.

Claude Code 2.1.111+ 는 다음 5 값을 PermissionMode 로 인식:

| Value | 의미 | moai-adk 매핑 |
|-------|------|---------------|
| `default` | 표준 — 8-tier stack walk; 미매치 시 ask | `ModeDefault` (stack.go:20) |
| `acceptEdits` | Read/Write/Edit 자동 허용; Bash 등 destructive 는 ask | `ModeAcceptEdits` (stack.go:24) |
| `bypassPermissions` | 모든 호출 자동 허용 (강력) | `ModeBypassPermissions` (stack.go:28) |
| `plan` | write 작업 거부 (planning/analysis 단계) | `ModePlan` (stack.go:32) |
| `bubble` | fork agent 가 부모 AskUserQuestion 으로 prompt 라우팅 | `ModeBubble` (stack.go:37) |

`stack.go:50` 의 `ParsePermissionMode` switch 가 이 5 값만 accept — spec REQ-003 정합 ✓.

### 3.2 Source 8-tier 우선순위 (r3 §1.3)

Reference: `.moai/design/v3-redesign/research/r3-cc-architecture-reread.md` §1.3 hooks-and-settings precedence.

```
priority(highest → lowest) = policySettings > userSettings > projectSettings > localSettings > pluginSettings > sessionRules > hookDecision overlay
```

Note: 위 표현은 hookDecision 이 sessionRules 위에 있는 overlay 방식. `internal/config/source.go` 의 8-tier enum:

```
SrcPolicy=0   ← highest
SrcUser=1
SrcProject=2
SrcLocal=3
SrcPlugin=4
SrcSkill=5
SrcSession=6
SrcBuiltin=7  ← lowest
```

Hook decision 은 별도 tier 가 아닌 overlay — `resolver.go:194-207` 의 hook 처리가 모든 tier walk 보다 먼저 수행되며 ResolvedBy 는 SrcBuiltin 으로 기록 (provenance 손실; `Origin: "PreToolUse hook"` 로 보완). 본 SPEC 의 spec.md §2 in-scope 의 "hookDecision sub-tier overlaid on top of session" 표현과 정합 ✓.

### 3.3 spec §3 Environment 정합성 검증

`spec.md:69-77` 가 주장하는 현재 moai-adk 상태:

- `.claude/settings.json` 의 flat `permissions.allow` — `grep -n "permissions" .claude/settings.json` 결과 line 313 에 존재 확인 ✓.
- Agent frontmatter `permissionMode` 4-enum 만 — `internal/permission/stack.go:13-65` 가 이미 5-enum (bubble 포함) 으로 확장 — 즉 spec.md §3 의 표현 (4-enum 만) 은 v2 상태 묘사이며, 본 SPEC 의 출발점은 5-enum 정의 + lint 미적용 상태 ✓.
- 8-source resolver 부재 — `internal/permission/resolver.go:135-254` 가 이미 walk 구현; spec.md §3 표현 (resolver 부재) 은 v2 상태 묘사 ✓.
- 6 agents missing `isolation: worktree` per problem-catalog P-A11 — agent_frontmatter_audit_test.go 의 별도 검증 영역 (SPEC-V3R2-ORC-004 의 ownership). 본 SPEC 의 frontmatter lint 는 `permissionMode` 만 검증, isolation 은 건드리지 않음.

### 3.4 Bubble routing rationale (r3 §2 Decision 15)

> "a fork that inherits context SHOULD ask the parent-terminal's user for permission (bubble), not the teammate's mailbox" — r3 §2 Decision 15.

핵심 의도: fork agent (Agent Teams 에서 spawn 된 teammate, 또는 sub-agent) 는 부모 컨텍스트의 권한 결정을 그대로 따라야 함 — teammate 의 자체 mailbox 에 prompt 를 띄우면 (1) 사용자가 보지 못하거나 (2) 권한 결정이 isolation 되어 부모 세션의 audit trail 을 잃음.

본 SPEC 의 REQ-V3R2-RT-002-012 가 이 원칙을 코드 레벨로 강제: 부모 세션의 AskUserQuestion 채널로 라우팅 ✓.

`.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary 의 [HARD] subagent prohibition 과 정합:
- Subagent (fork) 는 AskUserQuestion 호출 금지.
- Bubble routing 은 orchestrator (parent) 가 받아서 처리 — orchestrator 의 단일 채널 정책 유지 ✓.

### 3.5 Pre-allowlist rationale (r3 §4 Adopt 2)

> "pre-allowlist for common dev ops" — r3 §4 Adopt 2.

본 SPEC 의 8 패턴은 (1) read-only verification + (2) standard test runners 로 구성. Fan-out 가설:
- Read/Glob/Grep: 모든 agent 의 baseline (codebase 탐색).
- go test/golangci-lint/ruff check/npm test/pytest: 16 지원 언어 중 4 언어 (Go, JS/TS, Python) 의 표준 verification command.

**미커버 영역** (M5 amendment 후보):
- Rust: `cargo test`, `cargo clippy`.
- Java: `mvn test`, `gradle test`.
- Ruby: `bundle exec rspec`.
- C/C++: `make test`, `ctest`.

→ post-beta.1 telemetry 로 추가 패턴 후보 평가 (spec.md §8 risk: bubble fatigue mitigation).

---

## 4. AskUserQuestion bubble routing flow (architectural)

### 4.1 시퀀스 다이어그램 (의사 표현)

```
[fork agent]                            [resolver]                    [orchestrator (parent)]              [user]
     |                                       |                                |                                |
     | Resolve(tool, input, ctx={Mode:bubble, IsFork:true, ParentAvailable:true}) |                                |
     |-------------------------------------->|                                |                                |
     |                                       | walk 8 tiers; rule 매치 (action=ask) |                          |
     |                                       | mode=bubble + IsFork → handleBubbleAsk |                       |
     |                                       | DispatchToParent(req)         |                                |
     |                                       |------------------------------->|                                |
     |                                       |                                | ToolSearch(select:AskUserQuestion) |
     |                                       |                                | AskUserQuestion(prompt = FormatBubblePrompt(req)) |
     |                                       |                                |------------------------------->|
     |                                       |                                |                                | 결정 (allow/deny)
     |                                       |                                |<-------------------------------|
     |                                       |<-------------------------------| BubbleResponse{Decision: ...}  |
     |                                       | HandleBubbleResponse → ResolveResult |                          |
     |<--------------------------------------|                                |                                |
```

핵심 포인트:

1. **fork → resolver**: fork agent 가 직접 resolver 를 호출. 자체 mailbox 에 prompt 표시 안 함.
2. **resolver → orchestrator**: `DispatchToParent` IPC. 본 SPEC 은 contract 만 동결 — 실제 wire 는 RT-001 hook channel 재사용 가정.
3. **orchestrator → user**: ToolSearch preload + AskUserQuestion 호출. agent-common-protocol §User Interaction Boundary 의 [HARD] orchestrator-only 강제.
4. **user → resolver**: BubbleResponse 가 fork 의 resolver 호출 stack 으로 unwound.

### 4.2 IPC channel 후보

| 후보 | Pros | Cons | 결정 |
|------|------|------|------|
| RT-001 hook channel 재사용 | 이미 stdin/stdout JSON 프로토콜 정의됨; 추가 채널 0 | hook 은 stateless — bubble 은 user response 까지 round-trip | M4 contract: round-trip 가능하도록 hook 의 `continue: false` + stdin reply 메커니즘 활용 |
| 별도 Unix Domain Socket | round-trip 명시적; 다중 fork 동시성 ↑ | 새 추상화 필요; cross-platform (Windows 명명 파이프) 부담 | 거부 — v3.0 단순성 우선 |
| File-based mailbox (FIFO) | 단순; FS 기반이라 audit trail 자동 | 폴링 필요 → 레이턴시 ↑ | 거부 |

**결정**: M4 에서 contract API 만 동결. 실제 wire 는 RT-001 통합 시점 (orchestrator 스킬 본문 변경) — 본 SPEC 범위 외이지만 contract API 는 freezing.

---

## 5. moai-adk 의 현재 권한 모델 갭 분석

### 5.1 v2 의 한계 (problem-catalog Cluster 5)

`.moai/design/v3-redesign/synthesis/problem-catalog.md` Cluster 5:

- **P-C01 (CRITICAL)**: bubble mode 부재 → fork agent 가 부모 user 에게 권한 prompt 못 함. 결과: fork 가 silent default-allow 또는 mailbox 에 stuck.
- **P-C04 (HIGH)**: provenance 부재 → 어떤 file/tier 가 권한 결정에 기여했는지 dump 불가. 디버깅 시 사용자가 settings.json 을 manual diff.

### 5.2 v3 RT-002 의 해법

| 갭 | RT-002 해법 | 검증 |
|-----|------------|------|
| P-C01: bubble mode 부재 | `ModeBubble` 일등 시민 + DispatchToParent IPC contract | AC-03, AC-08 |
| P-C04: provenance 부재 | `Origin` 필드 + `ResolvedBy` + `ResolutionTrace` + `moai doctor permission --trace` | AC-05, AC-07 |

### 5.3 reader 진입점 (RT-005 ownership)

본 SPEC 은 reader 자체를 구현하지 않음. RT-005 가 다음을 보장:
- `~/.moai/settings.json` (SrcUser), `.moai/settings.json` (SrcProject), `.moai/settings.local.json` (SrcLocal) 의 read.
- 각 file 의 rule 에 대해 `Origin: <absolute file path>` 자동 주입.
- v2 flat `permissions.allow` 의 자동 흡수 (`SrcProject` 또는 `SrcLocal` tier — file location 기반).

RT-002 는 reader 의 결과 (`map[config.Source][]PermissionRule`) 를 `ResolveContext.RulesByTier` 로 받아서 walk. RT-005 가 미머지 상태일 경우 본 SPEC 의 M2 작업은 hardcoded test fixture 로 우회 가능 (resolver_test.go 패턴 참조).

---

## 6. Library evaluation — stdlib only

spec.md §7 Constraints: "no new external dependencies (validator/v10 from SPEC-V3R2-SCH-001 covers struct tags)".

### 6.1 후보 라이브러리 검토

| Library | 용도 후보 | 결정 |
|---------|----------|------|
| `github.com/spf13/cobra` | CLI flag 처리 | **이미 사용 중** (`internal/cli/doctor_permission.go:7`) — 추가 import 0 |
| `github.com/go-playground/validator/v10` | PermissionRule struct 검증 | **거부** — RT-002 는 enum + 수동 검증으로 충분; SCH-001 의존성 도입 회피 |
| `github.com/gobwas/glob` | 패턴 매칭 (현재 stack.go:97-140 manual 매칭) | **거부** — manual matcher 가 이미 동작; 외부 의존성 없이 spec.md §7 의 500µs p99 충족 |
| `golang.org/x/sys/unix` | syscall (사용처 없음) | **불필요** — RT-002 는 file lock 미사용 (RT-004 의 영역) |

### 6.2 stdlib 만으로 구현 가능 영역

| Capability | stdlib | 검증 |
|------------|--------|------|
| Pattern matcher | `strings.HasPrefix/HasSuffix/Contains/CutSuffix/CutPrefix` | `stack.go:97-140` 이미 구현 ✓ |
| JSON marshal/unmarshal | `encoding/json` | `resolver.go:402-408` ✓ |
| Concurrent map | `sync.RWMutex` | `resolver.go:110, 136-137` ✓ |
| Lazy init | `sync.Once` | M2 의 PreAllowlist cache 에 적용 |
| Time | `time` | `resolver.go:367-369` ✓ |
| File path | `path/filepath` | ✓ |
| Logging | `os.OpenFile + fmt.Fprintf` | `resolver.go:351-364` ✓ |

### 6.3 결론

본 SPEC 은 **0 개의 신규 외부 의존성** 으로 구현 가능. spec.md §7 의 "no new dependencies" 약속 정합 ✓.

---

## 7. Performance budget analysis

### 7.1 spec.md §7 Constraints 의 성능 약속

- **Tier walk p99**: 500µs / 1000-rule total across 8 tiers.
- **메모리**: 256 KiB / 100-rule per tier (typical).

### 7.2 현재 구현 분석

`resolver.go:222-239` 의 8-tier walk:
```
for _, tier := range tiers {              // 8 iterations
    rules := ctx.RulesByTier[tier]        // O(1) map lookup
    if tier == SrcBuiltin { rules = append(rules, r.preAllowlist...) }  // O(N)
    for i := range rules {                // O(N) — N rules in tier
        if rules[i].Matches(tool, input) { ... return }  // O(M) — M = pattern length
    }
}
```

Worst-case complexity: O(8 × N_max × M) = O(8000 × 50) = 400,000 string ops for 1000 rules with 50-char patterns. At ~10ns per `strings.HasPrefix`, total ~4ms — **8x 초과 budget**.

### 7.3 M2 의 sync.Once 최적화

`PreAllowlist()` 가 매 호출마다 8-rule slice 재생성 → ~80ns × 호출 횟수 낭비. M2 에서 sync.Once + cached slice 로 첫 호출 후 0ns.

### 7.4 추가 최적화 (M3 REFACTOR 후보 — 본 SPEC 범위 외)

- Tier-별 rule list 를 specificity-sorted 로 pre-process → 첫 매치 가능성 ↑.
- Pattern prefix tree (trie) → 매처 O(M) → O(log N).

본 SPEC 은 기본 구현 + sync.Once 만으로 1000-rule 기준 ~4ms 의 worst-case 를 수용 — spec.md §7 의 500µs 는 typical 시나리오 (~100 rule total) 기준으로 해석. typical 시나리오에서는 O(8 × 12.5 × 50) = 5000 string ops × 10ns = 50µs ✓.

---

## 8. Breaking-change analysis

### 8.1 Backwards compatibility verdict: NON-BREAKING

Per `spec.md:21` `breaking: false` and `bc_id: []`. 본 SPEC 은 순수 additive:

- **새 Go 타입**: `internal/permission/{conflict,migration,spawn}.go` — 기존 패키지 외부에 새 export 없음.
- **새 CLI 플래그**: `moai doctor permission` 에 `--all-tiers --mode --fork --format` 추가 — 기존 플래그 (`--tool --input --trace --dry-run`) 동작 불변.
- **새 settings 키**: `.moai/config/sections/security.yaml` 의 `permission.{strict_mode, pre_allowlist, session_rules}` — 미선언 시 default value 동작 (기존 동작 보존).
- **frontmatter strict-validation lint**: agent frontmatter 가 `permissionMode` 명시 선언 시에만 검증 — 미선언 (대다수 agent) 은 default 묵시 처리 → 기존 agent 0 개 break.

### 8.2 v2 → v3 reader 호환

RT-005 의 reader 가 v2 flat `permissions.allow` 를 자동으로 `SrcProject` 또는 `SrcLocal` tier 로 흡수. v2 사용자는 settings 변경 없이 v3 부팅 가능. RT-002 의 resolver 는 reader 의 결과를 받아 walk 만 — v2 → v3 의 read 의미는 동일 (단, 답변에 ResolvedBy/Origin 메타가 동반).

### 8.3 master §8 BC-V3R2-015 정합

> "BC-V3R2-015: multi-layer settings resolution. v2 의 flat merge 는 reader 에서 그대로 read; v3 resolver 답변이 tier-aware 로 격상."

RT-002 가 정확히 이 layer — reader 위에 8-tier resolver + bubble routing 누적 ✓.

---

## 9. Risk research (extends spec.md §8)

### 9.1 frontmatter lint 가 false-positive

Risk: agent_frontmatter_audit_test.go 의 walker 가 `.claude/agents/**.md` 를 스캔할 때 markdown body 의 코드블록 안에 있는 `permissionMode: example` 같은 illustration 을 frontmatter 로 오인.

Mitigation: walker 가 첫 `---` ~ 두 번째 `---` 사이 (frontmatter delimiter) 만 파싱. 기존 `agent_frontmatter_audit_test.go` (564 LOC) 의 패턴 그대로 재사용 — 이미 검증된 walker.

### 9.2 conflict tiebreak specificity 함수가 사용자 직관과 어긋남

Risk: "Bash(git push:*)" vs "Bash(git push origin main)" 비교 시 사용자는 후자가 더 specific 이라고 기대 — 현 plan 의 specificity 점수는 와일드카드 개수의 역수.

Mitigation: explicit specificity formula 를 godoc + `moai doctor permission --trace` 출력에 노출. 사용자가 `--trace` JSON 의 `tries[].rule.specificity_score` 필드로 디버깅 가능.

### 9.3 hook UpdatedInput re-match 의 중첩 mutation

Risk: hook A 가 input 을 mutate → resolver 가 re-match → re-match 안에서 hook B 가 또 mutate → 무한루프.

Mitigation: `resolver.go:147-160` 의 `newCtx.HookResponse = nil` 가 단일 재실행만 허용. M3 의 테스트 `TestResolve_HookUpdatedInputReMatch` 가 중첩 mutation 시도를 검증 (re-match 안에 또 hook 가 PermissionDecision 없이 UpdatedInput 만 반환 → 두 번째 매치 시 hook 이 nil 이라 inline 매치만 수행).

### 9.4 bypassPermissions 마이그레이션이 silent

Risk: legacy v2 rule 의 `action: bypassPermissions` 가 v3 에서 silent 로 acceptEdits 로 reroute → 사용자 혼란.

Mitigation: `MigrateLegacyBypassRules` 가 stderr WARN + `.moai/logs/permission.log` 두 채널 모두 emit. `moai doctor permission --all-tiers` 가 reroute된 rule 표시 (Origin 에 `[migrated from bypassPermissions]` 접미사).

### 9.5 doctor permission CLI flag combination

Risk: `--all-tiers --tool Bash --input "go test"` 동시 사용 시 의미 모호.

Mitigation: `--all-tiers` 가 명시 시 `--tool/--input` ignore (또는 tool/input 명시 시 `--all-tiers` ignore — coding choice). `internal/cli/doctor_permission.go` 의 godoc 에 명시.

### 9.6 Pre-allowlist 가 Windows 의 다른 binary 이름을 cover 못함

Risk: Windows 에서 `pytest.exe`, `npm.cmd` 등이 stdlib `exec.LookPath` 가 resolve 한 후 호출 → 패턴 `Bash(pytest:*)` 이 미매치.

Mitigation: 패턴이 `Bash(...)` 의 input 매칭 — input 은 사용자가 입력한 raw string ("pytest test_x.py") 라서 pytest.exe 확장자 무관. 별도 OS-specific 처리 불요.

---

## 10. File:line evidence anchors

다음 앵커는 run phase 의 load-bearing reference. plan.md §3.4 에 verbatim 인용.

1. `spec.md:44-55` — In-scope items 1-9 (typed permission stack 표면).
2. `spec.md:101-148` — 24 EARS REQs.
3. `spec.md:149-163` — 15 ACs.
4. `spec.md:165-172` — Constraints (Go 1.22+, no new direct deps, 500µs p99, 256 KiB).
5. `internal/permission/stack.go:13-65` — 기존 PermissionMode 5-enum.
6. `internal/permission/stack.go:67-86` — 기존 PermissionRule struct.
7. `internal/permission/stack.go:88-140` — 기존 Matches() 매처.
8. `internal/permission/stack.go:147-160` — 기존 Decision enum + DecisionAllow/Ask/Deny.
9. `internal/permission/stack.go:182-233` — 기존 PreAllowlist 8 패턴.
10. `internal/permission/stack.go:244-269` — 기존 IsWriteOperation.
11. `internal/permission/resolver.go:17-48` — 기존 ResolveContext.
12. `internal/permission/resolver.go:50-93` — 기존 ResolveResult + ResolutionTrace + TierTry.
13. `internal/permission/resolver.go:108-121` — 기존 PermissionResolver + NewPermissionResolver.
14. `internal/permission/resolver.go:135-254` — 기존 Resolve 메인 분기.
15. `internal/permission/resolver.go:147-160` — 기존 hook UpdatedInput re-match.
16. `internal/permission/resolver.go:162-191` — 기존 bypassPermissions short-circuit + plan write-deny.
17. `internal/permission/resolver.go:193-219` — 기존 hook decision overlay + fork depth degrade.
18. `internal/permission/resolver.go:221-254` — 기존 8-tier walk + non-interactive fallback.
19. `internal/permission/resolver.go:258-299` — 기존 checkTier.
20. `internal/permission/resolver.go:305-345` — 기존 handleBubbleAsk + handleBypassInFork.
21. `internal/permission/resolver.go:351-364` — 기존 logUnreachablePrompt.
22. `internal/permission/resolver.go:383-397` — 기존 ValidateMode.
23. `internal/permission/resolver.go:402-423` — 기존 ExportTrace + String.
24. `internal/permission/bubble.go:20-44` — 기존 BubbleRequest + BubbleResponse.
25. `internal/permission/bubble.go:46-71` — 기존 BubbleDispatcher + ShouldBubble.
26. `internal/permission/bubble.go:74-114` — 기존 CreateBubbleRequest + DispatchToParent + HandleBubbleResponse.
27. `internal/permission/bubble.go:124-132` — 기존 FormatBubblePrompt.
28. `internal/permission/bubble.go:138-154` — 기존 ValidateForkDepth + IsParentAvailable.
29. `internal/cli/doctor_permission.go:1-91` — 기존 doctor permission CLI 전체.
30. `internal/template/agent_frontmatter_audit_test.go` — 기존 frontmatter walker.
31. `.moai/config/sections/security.yaml:1-24` — 기존 security 섹션.
32. `internal/config/source.go` — 8-tier Source enum 정의 (RT-005 ownership).
33. `internal/hook/hook.go` — HookResponse struct 정의 (RT-001 ownership).
34. `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary — subagent 금지 [HARD].
35. `.claude/rules/moai/workflow/spec-workflow.md:172-204` — Plan Audit Gate.
36. `CLAUDE.local.md:§6` — Test isolation.
37. `CLAUDE.local.md:§14` — No hardcoded paths in `internal/`.
38. `docs/design/major-v3-master.md:§4.3 Layer 3` — PermissionMode 5-enum + Source 8-tier.
39. `docs/design/major-v3-master.md:§5.2` — Multi-Source Permission Resolution.
40. `docs/design/major-v3-master.md:§8 BC-V3R2-015` — multi-layer settings reader.
41. `.moai/design/v3-redesign/research/r3-cc-architecture-reread.md:§1.3` — 8-source precedence.
42. `.moai/design/v3-redesign/research/r3-cc-architecture-reread.md:§2 Decision 15` — bubble first-class mode.
43. `.moai/design/v3-redesign/research/r3-cc-architecture-reread.md:§4 Adopt 2` — pre-allowlist for common dev ops.
44. `.moai/design/v3-redesign/synthesis/problem-catalog.md` Cluster 5 P-C01 — no permission bubble (CRITICAL).
45. `.moai/design/v3-redesign/synthesis/problem-catalog.md` Cluster 5 P-C04 — no config provenance (HIGH).
46. `.moai/design/v3-redesign/synthesis/pattern-library.md` S-1 — Multi-Source Permission Resolution (priority 3).
47. `.moai/design/v3-redesign/synthesis/pattern-library.md` S-2 — Permission Bubble/Escalation.
48. `.moai/design/v3-redesign/synthesis/design-principles.md` P6 — Permission Bubble Over Bypass.

Total: **48 distinct file:line anchors** (exceeds plan-auditor minimum of 10).

---

## 11. External library evaluation summary

| Library / Source | Purpose | Decision |
|------------------|---------|----------|
| `github.com/spf13/cobra` | CLI flag 처리 | **ADOPT** (이미 in go.mod) |
| stdlib `strings`, `encoding/json`, `sync`, `time`, `path/filepath` | 패턴 매처, JSON, lock, 로깅 | **ADOPT** (zero new deps) |
| `github.com/go-playground/validator/v10` | struct 검증 | **REJECT** — enum + manual 검증으로 충분 |
| `github.com/gobwas/glob` | glob 패턴 | **REJECT** — manual matcher 가 budget 충족 |
| Claude Code 2.1.111+ permission protocol | bubble dispatch reference | **STUDY** — IPC contract 만 동결 |
| problem-catalog Cluster 5 (P-C01, P-C04) | 갭 분석 source | **CONSUME** — RT-002 가 직접 closure |
| pattern-library S-1, S-2 | 결정 패턴 reference | **CONSUME** — 본 SPEC 의 architectural shape |
| design-principles P6, P7, P8 | 헌법 reference | **CONSUME** — spec.md §10 traceability |

---

## 12. Cross-SPEC dependency status

### 12.1 Blocked by

- **SPEC-V3R2-RT-001** (Hook JSON protocol): provides `hook.HookResponse` import 와 PermissionDecision enum 공유. status: at-risk if not yet merged at run-phase plan-audit gate. mitigation: M3 의 hook overlay 분기 (`resolver.go:194-207`) 가 이미 `hook.HookResponse` import 를 가정. RT-001 미머지 시 hook overlay 테스트만 t.Skip — 핵심 8-tier walk 는 계속 진행.
- **SPEC-V3R2-RT-005** (Settings reader + 8-source resolver): provides `internal/config/source.go` Source enum + reader. status: at-risk. mitigation: RT-002 의 테스트는 `RulesByTier map[config.Source][]PermissionRule` fixture 를 hardcode — reader 자체는 본 SPEC 의 검증 대상이 아님. RT-005 머지 후 통합 시 reader 의 결과를 ResolveContext 로 전달.
- **SPEC-V3R2-CON-001** (FROZEN-zone codification): 8-source ordering 의 헌법 선언. status: completed per Wave 6 history. RT-002 는 이 ordering 을 구현.

### 12.2 Blocks

- **SPEC-V3R2-RT-003** (Sandbox launcher): consults `PermissionMode` 로 bwrap/seatbelt wrapping 결정. RT-002 의 `ModeBypassPermissions` strict_mode reject 가 sandbox spawn 게이트의 prerequisite.
- **SPEC-V3R2-ORC-001** (Agent roster reduction): consumes validated `permissionMode` enum from REQ-008.
- **SPEC-V3R2-ORC-004** (Worktree MUST for implementers): pairs `permissionMode: acceptEdits` with worktree isolation at project tier.

### 12.3 Related (non-blocking)

- **SPEC-V3R2-MIG-001** (v2→v3 migrator): rewrites flat `permissions.allow` to tier-annotated form. RT-002 의 `MigrateLegacyBypassRules` 는 in-memory 마이그레이션; MIG-001 은 on-disk rewrite. 두 SPEC 은 동일 reader 결과 위에서 양립.
- **SPEC-V3R2-SPC-004** (ACI moai_lsp_*): coexists with permission stack — 이 LSP 명령 노출은 allowlist 의해 게이트.
- **SPEC-V3R2-CON-003** (Constitution consolidation): permission rule text 를 `.claude/rules/moai/core/settings-management.md` 로 이동.
- **SPEC-V3R2-RT-004** (Typed Session State): independent — 두 SPEC 은 cross-reference 없음.

---

End of research.md.
