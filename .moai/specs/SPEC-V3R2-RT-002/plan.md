# SPEC-V3R2-RT-002 Implementation Plan

> Implementation plan for **Permission Stack + Bubble Mode**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Authored against branch `plan/SPEC-V3R2-RT-002` at `<worktree-root>` = `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-002`. See §7 for cwd resolution rule.

## HISTORY

| Version | Date       | Author                        | Description                                                              |
|---------|------------|-------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial implementation plan per `.claude/skills/moai/workflows/plan.md` Phase 1B |

---

## 1. Plan Overview

### 1.1 Goal restatement

Concretize spec.md §1 (Goal) into a typed Go subsystem under `internal/permission/` and supporting CLI plumbing under `internal/cli/` so that every tool invocation in moai-adk-go resolves through the canonical 8-source priority chain (`SrcPolicy > SrcUser > SrcProject > SrcLocal > SrcPlugin > SrcSkill > SrcSession > SrcBuiltin`) per master §5.2 and master §4.3 Layer 3 type block.

핵심 요약:

- 기존 `internal/permission/{stack,resolver,bubble}.go` 스켈레톤(약 700 LOC, ~50% REQ 충족)을 24 EARS REQ + 15 AC 가 모두 통과하도록 보강.
- v3.0 의 새 PermissionMode 값인 `bubble` 을 enum 일등 시민으로 격상 (`ModeBubble = "bubble"` 이미 stack.go:37 에 존재 — 호출 경로/lint/AskUserQuestion 라우팅이 비어있음).
- 8-source 모든 tier 의 rule 을 `RulesByTier map[config.Source][]PermissionRule` 형태로 적재하는 reader (RT-005 가 직접 구현. RT-002 는 그 결과를 소비) 와 `Origin` 필드 기반 provenance 의 end-to-end 흐름을 검증.
- `moai doctor permission` 서브커맨드를 보강 — 현 skel(.go 91 LOC)은 hard-coded `RulesByTier` 빈맵으로만 동작. `--all-tiers` 출력, `--trace` JSON 형식 정합성, `--dry-run` 정상화 필요.
- 8 개의 dev-op 패턴(`Bash(go test:*)` 외) 을 `SrcBuiltin` tier 의 pre-allowlist 로 박제 → bubble fatigue 80% 흡수.
- 기존 `bypassPermissions` 사용처를 `security.yaml strict_mode: true` 환경에서 spawn 단계 reject 하는 게이트.
- Agent frontmatter `permissionMode` 값을 5-enum 으로 strict 검증하는 CI lint 추가 (`internal/template/agent_frontmatter_audit_test.go` 확장).

### 1.2 비즈니스 핵심

이 SPEC 은 `bc_id: []` (additive) 입니다. v2.x 의 flat `permissions.allow` 리스트는 reader (SPEC-V3R2-RT-005) 가 자동으로 `SrcProject` 또는 `SrcLocal` tier 로 흡수해 read 호환을 유지하며, 기존 사용자에게 마이그레이션 부담을 주지 않습니다. 본 plan 은 master §8 BC-V3R2-015 의 multi-layer settings reader 위에 8-tier resolver + bubble routing 을 얹는 작업입니다. resolver answer 는 tier-aware (Origin/ResolvedBy 메타 동반) 가 되지만 v2 의 단일-soruce 답에서 동일 결정을 항상 derive 가능 → non-breaking.

### 1.3 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run.

- **RED**: stack.go/resolver.go/bubble.go 가 채우지 못한 동작에 대한 실패 테스트 작성. 기존 `*_test.go` 가 이미 일부 happy-path 를 cover (stack_test.go 215 LOC, resolver_test.go 480 LOC, bubble_test.go 220 LOC) — bypassPermissions strict_mode rejection, frontmatter strict-validation, doctor permission --trace JSON, hook UpdatedInput re-match 등이 미커버.
- **GREEN**: stack.go 의 `IsWriteOperation` 패턴 보강, resolver.go 의 `--trace` JSON 정합성 보강, bubble.go 의 DispatchToParent placeholder 를 orchestrator IPC contract 로 대체 (RT-001 hook protocol 통신 채널 재사용). `internal/cli/doctor_permission.go` 에 `--all-tiers` 출력 + `--mode` 플래그 추가. `internal/template/agent_frontmatter_audit_test.go` 에 `permissionMode oneof` 검증 추가. `.moai/config/sections/security.yaml` 에 `permission.pre_allowlist`, `permission.strict_mode` 키 추가.
- **REFACTOR**: pre-allowlist 패턴 매칭 hot path 를 sort-once + binary-search 로 변환 (1000-rule 기준 p99 500µs 목표). Decision/Source 문자열 변환 통합. spec REQ-V3R2-RT-002-040 (legacy bypassPermissions 마이그레이션 deprecation warning) 을 reader 진입점 (SPEC-V3R2-RT-005 영역) 과 동일 hook 으로 일관화.

### 1.4 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| `Source` enum 8-tier 정합성 확인 (이미 internal/config/source.go 존재) | `internal/config/source.go` (검증) | REQ-001 |
| `PermissionMode` 5-enum 일등 시민 정착 | `internal/permission/stack.go:13-65` (기존 보강) | REQ-003 |
| `PermissionRule` Origin 필드 비빔 검증 | `internal/permission/stack.go:67-86` (기존 — provenance 흐름 end-to-end 테스트 추가) | REQ-002 |
| `PermissionResolver.Resolve` 8-tier 워크 + 첫 매치 반환 | `internal/permission/resolver.go:135-254` (기존 보강 — RulesByTier reader 연결) | REQ-004, REQ-005 |
| Hook UpdatedInput 후 re-match | `internal/permission/resolver.go:147-160` (기존 보강 — 무한루프 방지 가드 검증) | REQ-013 |
| Bubble routing → 부모 AskUserQuestion | `internal/permission/bubble.go:94-114` (DispatchToParent IPC 구현) | REQ-012, REQ-050 |
| Pre-allowlist 8 패턴 시드 | `internal/permission/stack.go:182-233` (기존 — 테스트 강화) | REQ-006 |
| `moai doctor permission` 보강 (`--trace`, `--dry-run`, `--all-tiers`) | `internal/cli/doctor_permission.go` (보강) | REQ-007, REQ-015 |
| Agent frontmatter `permissionMode` strict-validation | `internal/template/agent_frontmatter_audit_test.go` (확장) | REQ-008 |
| Plan mode write-deny 게이트 | `internal/permission/resolver.go:178-191` (기존 — write 패턴 보강) | REQ-020 |
| bypassPermissions strict_mode reject | `internal/permission/resolver.go:383-397` (기존 ValidateMode 호출 사이트 추가) | REQ-022 |
| Fork depth >3 → bubble 강등 | `internal/permission/resolver.go:209-219` (기존 보강) | REQ-023 |
| `permission.session_rules` `SrcSession` tier 적재 | reader 연결 — RT-005 dependency | REQ-030 |
| Legacy `bypassPermissions` action → `acceptEdits` migration | `internal/permission/migration.go` (신규, ~50 LOC) | REQ-040 |
| 비대화형 `ask` → `deny` fail-closed + log | `internal/permission/resolver.go:241-247` (기존 보강) | REQ-041 |
| 동일 tier 충돌 시 specificity-then-fs-order tiebreak | `internal/permission/conflict.go` (신규, ~80 LOC) | REQ-042 |
| Fork session-scope rule SrcPolicy override 차단 | `internal/permission/resolver.go` (보강) | REQ-043 |
| `security.yaml` `permission.pre_allowlist`, `permission.strict_mode`, `permission.session_rules` 키 | `.moai/config/sections/security.yaml` (확장) | REQ-006, REQ-022, REQ-030 |
| CHANGELOG entry | `CHANGELOG.md` Unreleased section | Trackable (TRUST 5) |
| MX tags per §6 | 5 files (per §6 below) | mx_plan |

Embedded-template parity: 본 SPEC 은 `internal/permission/` Go 코드와 `.moai/config/sections/security.yaml` 의 두 표면을 건드립니다. security.yaml 은 `internal/template/templates/.moai/config/sections/security.yaml` 의 미러도 함께 업데이트해야 하므로, M5 단계에 `make build` 를 실행해 `internal/template/embedded.go` 를 재생성합니다.

### 1.5 Traceability Matrix (REQ → AC → Task)

Per plan-auditor PASS criterion #2 (every REQ maps to at least one AC and at least one Task):

| REQ ID | Category | Mapped AC(s) | Mapped Task(s) |
|--------|----------|--------------|----------------|
| REQ-V3R2-RT-002-001 | Ubiquitous | (Source enum coverage; verified by existing `internal/config/source_test.go`) | T-RT002-01 |
| REQ-V3R2-RT-002-002 | Ubiquitous | AC-01, AC-02, AC-05 | T-RT002-02 |
| REQ-V3R2-RT-002-003 | Ubiquitous | AC-03 | T-RT002-03 |
| REQ-V3R2-RT-002-004 | Ubiquitous | AC-01, AC-13 | T-RT002-04, T-RT002-08 |
| REQ-V3R2-RT-002-005 | Ubiquitous | AC-01, AC-02 | T-RT002-04 |
| REQ-V3R2-RT-002-006 | Ubiquitous | AC-02 | T-RT002-05 |
| REQ-V3R2-RT-002-007 | Ubiquitous | AC-05 | T-RT002-12 |
| REQ-V3R2-RT-002-008 | Ubiquitous | AC-09 | T-RT002-14 |
| REQ-V3R2-RT-002-010 | Event-Driven | (default-ask vs default-deny) | T-RT002-08 |
| REQ-V3R2-RT-002-011 | Event-Driven | AC-04 | T-RT002-09 |
| REQ-V3R2-RT-002-012 | Event-Driven | AC-03 | T-RT002-10, T-RT002-21 |
| REQ-V3R2-RT-002-013 | Event-Driven | AC-10 | T-RT002-09 |
| REQ-V3R2-RT-002-014 | Event-Driven | AC-13 | T-RT002-08 |
| REQ-V3R2-RT-002-015 | Event-Driven | AC-05 | T-RT002-12 |
| REQ-V3R2-RT-002-020 | State-Driven | AC-06 | T-RT002-11 |
| REQ-V3R2-RT-002-021 | State-Driven | (bypassPermissions short-circuit) | T-RT002-13 |
| REQ-V3R2-RT-002-022 | State-Driven | AC-07 | T-RT002-13 |
| REQ-V3R2-RT-002-023 | State-Driven | AC-14 | T-RT002-15 |
| REQ-V3R2-RT-002-030 | Optional | (session-scope rule loading) | T-RT002-16 |
| REQ-V3R2-RT-002-031 | Optional | (plugin-tier slot) | T-RT002-17 |
| REQ-V3R2-RT-002-032 | Optional | (CLI behaviour) | T-RT002-12 |
| REQ-V3R2-RT-002-040 | Unwanted | AC-11 | T-RT002-18 |
| REQ-V3R2-RT-002-041 | Unwanted | AC-15 | T-RT002-19 |
| REQ-V3R2-RT-002-042 | Unwanted | AC-12 | T-RT002-20 |
| REQ-V3R2-RT-002-043 | Unwanted | (fork session-rule SrcPolicy override 차단) | T-RT002-22 |
| REQ-V3R2-RT-002-050 | Complex | AC-08 | T-RT002-21 |
| REQ-V3R2-RT-002-051 | Complex | AC-13 | T-RT002-23 |

Coverage: **27 REQs mapped to 15 ACs and 28 tasks** (몇몇 AC/task 는 다중 REQ 와 매핑).

---

## 2. Milestone Breakdown (M1-M5)

각 milestone 은 **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation HARD rule).

### M1: Test scaffolding (RED phase) — Priority P0

기존 테스트: `internal/permission/{stack,resolver,bubble}_test.go` (총 ~28KB).

Owner role: `expert-backend` (Go test) 또는 직접 `manager-tdd` 실행.

Scope:
1. `internal/permission/resolver_test.go` 에 새 테스트 케이스 추가 (기존이 cover 하지 못하는 AC):
   - `TestResolve_FrontmatterUnknownPermissionMode` (AC-09, REQ-008) — agent frontmatter strict-validation lint 의 sentinel 검증.
   - `TestResolve_HookUpdatedInputReMatch` (AC-10, REQ-013) — hook 가 path 를 mutate 한 후 pre-allowlist 가 mutated path 에 대해 다시 매칭되는지.
   - `TestResolve_ForkDepth4DegradeToBubble` (AC-14, REQ-023) — depth=4, mode=acceptEdits → systemMessage emit + decision=ask.
   - `TestResolve_BubbleParentClosed` (AC-08, REQ-050) — IsFork=true, ParentAvailable=false → deny + "parent unavailable" message.
   - `TestResolve_NonInteractiveAskBecomesDeny` (AC-15, REQ-041) — `IsInteractive: false` + 매치 실패 → deny + log entry.
   - `TestResolve_ConflictSpecificityThenFsOrder` (AC-12, REQ-042) — 동일 tier 두 매치 시 specificity 우선, 동률 시 file-system scan order 우선, 충돌 로깅.
   - `TestResolve_PolicyDenyOverridesProjectAllow` (AC-13, REQ-014) — SrcPolicy deny 가 SrcProject allow 보다 우선.
   - `TestResolve_BypassPermissionsRejectedInStrictMode` (AC-07, REQ-022) — strict_mode 에서 spawn 거부.
   - `TestResolve_LegacyBypassActionMigratedWithDeprecationWarning` (AC-11, REQ-040) — v2 `bypassPermissions` action 발견 시 acceptEdits 로 reroute + log.
   - `TestResolve_SessionRulesLoadedAsSrcSession` (REQ-030) — session_rules 키가 SrcSession tier 로 적재.

2. `internal/cli/doctor_permission_test.go` 신규 작성 (기존 doctor_test.go 와 분리):
   - `TestDoctorPermission_AllTiersFlag` (AC-05) — `--all-tiers` 출력 8개 tier 모두 dump.
   - `TestDoctorPermission_TraceJSONFormat` (AC-05, REQ-015) — `--trace` 출력이 JSON 파싱 가능 + 8 tier 모두 `tries[]` 포함.
   - `TestDoctorPermission_DryRun` (REQ-032) — `--dry-run` 시 tool 실행 없음 사인.

3. `internal/template/agent_frontmatter_audit_test.go` 확장:
   - `TestAgentFrontmatter_PermissionModeStrictEnum` (AC-09, REQ-008) — `.claude/agents/**/*.md` 의 모든 frontmatter 의 `permissionMode` 키가 5-enum (`default|acceptEdits|bypassPermissions|plan|bubble`) 안에 있는지 walker 로 검증; 위반 시 sentinel `PERMISSION_MODE_UNKNOWN_VALUE: <file> declares permissionMode: <value>; allowed: default|acceptEdits|bypassPermissions|plan|bubble.`

4. 모든 새 테스트가 RED 상태인지 `go test ./internal/permission/ ./internal/cli/ ./internal/template/` 로 확인. 기존 테스트 (resolver_test.go 의 32 케이스, stack_test.go 의 21 케이스, bubble_test.go 의 16 케이스) 는 baseline GREEN 유지.

검증 게이트 (M2 진행 전): 새 테스트 ≥10 개가 sentinel 메시지와 함께 fail. 기존 테스트는 회귀 없음 (regression baseline).

[HARD] M1 에서는 test 파일 외부에 구현 코드 작성 금지.

### M2: Pre-allowlist 정착 + ValidateMode 호출사이트 + frontmatter strict lint (GREEN, part 1) — Priority P0

Owner role: `expert-backend`.

Scope:
1. `internal/permission/stack.go:182-233` 의 `PreAllowlist()` 함수를 그대로 유지 (skel 가 이미 8 패턴 제공). 단, **lazy-load** 패턴으로 변경: `var preAllowlistOnce sync.Once + var preAllowlistRules []PermissionRule` 로 첫 호출 시에만 빌드. spec.md REQ-006 의 hot-path 비용 제거.
2. `internal/permission/resolver.go:383-397` 의 `ValidateMode` 를 spawn 진입점에 wire — 즉 `internal/cli/run.go` 또는 agent spawn 경로에서 ValidateMode 호출 후 PermissionModeRejected 시 spawn 거부. 본 SPEC 범위 내에서 `internal/permission/spawn.go` 신규로 spawn-side helper `RejectIfStrict(mode, strictMode) error` 만 추가하고, 실제 wire 는 spawn caller 에 1줄 추가 (cross-package coupling 최소화).
3. `internal/template/agent_frontmatter_audit_test.go` (기존 564 LOC) 에 새 테스트 함수 `TestAgentFrontmatter_PermissionModeStrictEnum` 추가. 기존 walker 패턴 재사용.
4. `.moai/config/sections/security.yaml` 에 spec.md §3 envrionment 가 명시한 새 키 3 종 추가:
   ```yaml
   permission:
     strict_mode: false           # bypassPermissions 모드 차단 (SrcPolicy 위배 시 true 권장)
     pre_allowlist:               # SrcBuiltin tier 자동 적재 (REQ-006 시드와 합산)
       - "Bash(go test:*)"
       - "Bash(golangci-lint run:*)"
       - "Bash(ruff check:*)"
       - "Bash(npm test:*)"
       - "Bash(pytest:*)"
       - "Read(*)"
       - "Glob(*)"
       - "Grep(*)"
     session_rules: []            # SrcSession tier 동적 적재 (런타임)
   ```
5. `internal/template/templates/.moai/config/sections/security.yaml` 미러 업데이트.

검증: `TestAgentFrontmatter_PermissionModeStrictEnum`, `TestResolve_BypassPermissionsRejectedInStrictMode`, `TestResolve_FrontmatterUnknownPermissionMode` 가 GREEN 진입. 기존 테스트 회귀 없음.

### M3: Hook UpdatedInput re-match + Conflict tiebreak + Plan-mode write-deny 보강 (GREEN, part 2) — Priority P0

Owner role: `expert-backend`.

Scope:
1. `internal/permission/resolver.go:147-160` 의 hook UpdatedInput re-match 가드 검증. 현재 skel 가 무한루프 방지 위해 `newCtx.HookResponse = nil` 로 clear 하지만, hook 가 중첩 mutate 한 케이스 (UpdatedInput 안에 또 PermissionDecision 이 없는) 에 대한 단일 재실행만 허용하는지 테스트 강화.
2. `internal/permission/stack.go:244-269` 의 `IsWriteOperation` 의 write 패턴 보강:
   - 누락 패턴 추가: `git reset --hard`, `git clean`, `npm run build`, `make install`, `dd of=`.
   - `mv` 중복 항목 (line 252) 수정.
   - 정규식 매처 옵션 추가 — 현재 `strings.Contains` 만 사용 → false-positive 우려 있는 패턴(`echo `) 은 명령어 시작점 매칭으로 변경.
3. `internal/permission/conflict.go` 신규 파일 (~80 LOC):
   - `resolveConflict(rules []*PermissionRule, tool, input string) *PermissionRule` — 동일 tier 다중 매치 시 specificity 점수 (와일드카드 개수의 역수) 우선, 동률 시 `Origin` 파일 path lexicographic 순서 우선.
   - 충돌 발생 시 `.moai/logs/permission.log` 에 entry 기록 (REQ-042 의 충돌 logging).
4. `internal/permission/resolver.go::checkTier` (line 258-299) 가 동일 tier 에서 ≥2 매치 시 `resolveConflict` 호출하도록 변경.

검증: `TestResolve_HookUpdatedInputReMatch`, `TestResolve_ConflictSpecificityThenFsOrder`, `TestResolve_PolicyDenyOverridesProjectAllow` 가 GREEN. AC-04, AC-12, AC-13 충족.

### M4: Bubble dispatch IPC + Fork depth degrade + Non-interactive fail-closed (GREEN, part 3) — Priority P0

Owner role: `expert-backend`.

Scope:
1. `internal/permission/bubble.go:94-102` 의 `DispatchToParent` placeholder 를 contract 만 명시한 IPC stub 으로 격상. 실제 구현은 RT-001 hook protocol 채널 재사용을 가정 (contract: parent session 이 stdin JSON 으로 BubbleRequest 수신 → AskUserQuestion 후 stdout JSON 으로 BubbleResponse 반환). 본 SPEC 은 contract API 만 동결하고 placeholder 가 `Sentinel: orchestrator integration required` 를 emit 하도록 유지. 실제 wiring 은 orchestrator side (`/moai run` skill body) — 본 SPEC 범위 외이지만 contract 는 동결.
2. `internal/permission/resolver.go:209-219` 의 fork depth >3 게이트 정합성 보강:
   - `mode == ModePlan` 이거나 `mode == ModeBubble` 이면 게이트를 통과 (이미 OK).
   - 그 외 모드는 systemMessage emit + decision=ask + ResolvedBy=SrcBuiltin + Origin="fork depth limit" — AC-14 의 sentinel "Fork depth N exceeds limit - mode degraded to bubble" 정렬.
3. `internal/permission/resolver.go:241-247` 의 비대화형 fail-closed 분기 검증 — `IsInteractive: false` 시 default decision 이 `DecisionDeny` 로 변환 + `logUnreachablePrompt` 호출. 로그 entry 가 spec.md REQ-041 의 `.moai/logs/permission.log` path 와 일치하는지 확인.
4. `internal/permission/bubble.go::IsParentAvailable` (line 150-154) 의 placeholder 를 ParentSessionID 비어있지 않음 + 실제 session registry lookup 으로 contract 격상. registry 자체는 SPEC-V3R2-WF-003/RT-004 영역이므로 본 SPEC 은 interface 만 동결.

검증: `TestResolve_BubbleParentClosed`, `TestResolve_ForkDepth4DegradeToBubble`, `TestResolve_NonInteractiveAskBecomesDeny` 가 GREEN. AC-08, AC-14, AC-15 충족.

### M5: doctor permission 보강 + Legacy migration + Session rules + CHANGELOG + MX tags (GREEN, part 4 + REFACTOR + Trackable) — Priority P1

Owner role: `expert-backend` (CLI), `manager-docs` (CHANGELOG, MX tags).

Scope:

#### M5a: `moai doctor permission` CLI 보강

1. `internal/cli/doctor_permission.go` (현 91 LOC) 에 다음 플래그 추가:
   - `--all-tiers` — 모든 tier 의 적재된 rule 목록 dump (기본 false; tool/input 명시 없을 때 자동 활성).
   - `--mode <mode>` — agent permissionMode 시뮬레이션 (default/acceptEdits/bypassPermissions/plan/bubble).
   - `--fork` — fork agent 시뮬레이션 (IsFork=true).
   - 기존 `--trace`, `--dry-run`, `--tool`, `--input` 유지.
2. `result.ExportTrace()` 출력 JSON 이 spec.md AC-05 의 "JSON trace enumerating all 8 tiers inspected with `matched: true|false` per tier" 정합성 검증. 현재 skel 의 `ResolutionTrace.Tries` 는 매치된 tier 만 기록 → 매치 실패 tier 도 기록하도록 `checkTier` 의 unmatched-fallthrough 분기에서도 trace 적재 (이미 `resolver.go:292-296` 에서 push 하는 것으로 보이므로 검증만).
3. 출력 포맷: 인간 가독형은 indent 2-space, JSON 은 `--format json` 별도 플래그 (기본 indent 2-space, no `--format`).

#### M5b: Legacy bypassPermissions migration

1. `internal/permission/migration.go` 신규 (~50 LOC):
   - `MigrateLegacyBypassRules(rules []PermissionRule) ([]PermissionRule, []string)` — 입력 rule 중 `Action == "bypassPermissions"` (legacy v2 form) 을 발견 시:
     - Action 을 `acceptEdits` (mode 가 아닌 action) 로 reroute. 단 PermissionDecision 은 `allow` 로 매핑 (acceptEdits 는 mode 이지 action 이 아님 → 실용적으로는 `Action: DecisionAllow` 로 매핑하고 deprecation warning 만 emit).
     - deprecation warning 문자열 반환 (해당 rule 의 Origin file path 명시).
2. reader (RT-005) 통합 시점에 `MigrateLegacyBypassRules` 를 첫 read 직후 호출.

#### M5c: Session rules SrcSession tier 적재

1. `.claude/settings.local.json` 의 `permissions.session_rules` 키를 SrcSession tier 로 적재하는 reader hook (RT-005 영역). 본 SPEC 은 `permission.SrcSession` tier 의 rule 가 resolver 에서 정상 walk 되는지의 contract 만 동결 + 테스트 추가.

#### M5d: CHANGELOG + MX tags + final verification

1. CHANGELOG entry 추가 (한국어 — `git_commit_messages: ko`):
   ```
   ### 추가
   - SPEC-V3R2-RT-002: Permission Stack + Bubble Mode. `internal/permission/` 8-tier resolver 정착, `bubble` PermissionMode 일등 시민 격상, 8개 dev-op pre-allowlist 고정, `moai doctor permission --all-tiers --trace --mode <m> --fork` CLI 보강, agent frontmatter `permissionMode` strict-validation lint 추가, legacy `bypassPermissions` action 자동 마이그레이션. v2 flat `permissions.allow` 설정 read 호환 유지 (BC-V3R2-015 reader 위에 누적).
   ```
2. §6 의 MX tag 7개 삽입.
3. 워크트리 루트에서 `go test ./...` 전체 실행. ALL GREEN + 0 cascading failures (per `CLAUDE.local.md` §6 HARD rule).
4. `make build` 실행 → `internal/template/embedded.go` 재생성. security.yaml 미러 업데이트 확인.
5. `progress.md` 의 `run_complete_at` + `run_status: implementation-complete` 갱신 (M1-M5 모두 GREEN 후).

[HARD] M5 단계에서 `.moai/specs/` 또는 `.moai/reports/` 에 새 문서 생성 금지 — SPEC 구현 단계임.

---

## 3. File:line Anchors (concrete edit targets)

### 3.1 To-be-modified (existing files)

| File | Anchor | Edit type | Reason |
|------|--------|-----------|--------|
| `internal/permission/stack.go:67-86` | `PermissionRule` struct | Origin/Source provenance 흐름 e2e 테스트 추가 | M1 / REQ-002 |
| `internal/permission/stack.go:182-233` | `PreAllowlist()` | sync.Once lazy-load 변환 | M2 / REQ-006 |
| `internal/permission/stack.go:244-269` | `IsWriteOperation` | write 패턴 보강 (mv 중복 제거 + git reset/clean/npm build/make install/dd 추가 + echo 매처 정밀화) | M3 / REQ-020 |
| `internal/permission/resolver.go:135-254` | `Resolve` 메인 분기 | conflict tiebreak 호출 + non-interactive log path 검증 + fork depth degrade 정합성 | M3, M4 / REQ-004, REQ-023, REQ-041 |
| `internal/permission/resolver.go:147-160` | hook UpdatedInput re-match | 무한루프 방지 가드 + 단일 재실행 enforcement | M3 / REQ-013 |
| `internal/permission/resolver.go:209-219` | fork depth >3 게이트 | systemMessage sentinel 정렬 | M4 / REQ-023 |
| `internal/permission/resolver.go:258-299` | `checkTier` 동일 tier 다중 매치 | `resolveConflict` 호출 분기 | M3 / REQ-042 |
| `internal/permission/resolver.go:383-397` | `ValidateMode` | spawn-side 진입점 wire (RejectIfStrict 호출) | M2 / REQ-022 |
| `internal/permission/bubble.go:94-102` | `DispatchToParent` | IPC contract stub + RT-001 channel 재사용 명시 | M4 / REQ-012 |
| `internal/permission/bubble.go:150-154` | `IsParentAvailable` | placeholder → registry-lookup contract | M4 / REQ-050 |
| `internal/cli/doctor_permission.go` (전체 91 LOC) | `runDoctorPermission` | `--all-tiers --mode --fork` 플래그 추가 + JSON 출력 정합성 | M5a / REQ-007, REQ-015 |
| `internal/template/agent_frontmatter_audit_test.go` | 기존 walker | `TestAgentFrontmatter_PermissionModeStrictEnum` 추가 | M2 / REQ-008 |
| `.moai/config/sections/security.yaml` | top-level `permission:` block | strict_mode/pre_allowlist/session_rules 키 추가 | M2 / REQ-006, REQ-022, REQ-030 |
| `internal/template/templates/.moai/config/sections/security.yaml` | mirror | M2 와 동일 변경 | M2 / template-first |
| `CHANGELOG.md` | `## [Unreleased]` section | Added entry per §M5d | M5d / Trackable |

### 3.2 To-be-created (new files)

| File | Reason | LOC estimate |
|------|--------|--------------|
| `internal/permission/conflict.go` | `resolveConflict` specificity-then-fs-order tiebreak | ~80 |
| `internal/permission/conflict_test.go` | conflict tiebreak 테스트 | ~120 |
| `internal/permission/migration.go` | `MigrateLegacyBypassRules` deprecation warning | ~50 |
| `internal/permission/migration_test.go` | migration 테스트 | ~80 |
| `internal/permission/spawn.go` | `RejectIfStrict(mode, strictMode) error` spawn-side helper | ~30 |
| `internal/permission/spawn_test.go` | spawn helper 테스트 | ~50 |
| `internal/cli/doctor_permission_test.go` | doctor permission CLI 테스트 (3-4 케이스) | ~150 |

Total new: ~560 LOC. Total modified: ~150 LOC. Net additions: ~710 LOC across 7 new files + ~14 modified files.

### 3.3 NOT to be touched (preserved by reference)

다음 파일은 본 SPEC 의 run phase 가 절대 수정하지 않습니다 (cross-SPEC scope 보호).

- `internal/config/source.go:1-end` (8-tier `Source` enum 자체) — SPEC-V3R2-RT-005 의 unique ownership.
- `internal/hook/*.go` — SPEC-V3R2-RT-001 의 unique ownership (본 SPEC 은 `hook.HookResponse` 를 import 만 함; 정의 변경 금지).
- `internal/session/*.go` — SPEC-V3R2-RT-004 의 unique ownership (본 SPEC 은 PhaseState 읽지 않음).
- `internal/cli/run.go` — RT-002 는 spawn caller 위치 1줄 (RejectIfStrict 호출) 만 추가 — broad refactor 금지.
- `.claude/agents/moai/*.md` 의 frontmatter — `permissionMode` 미선언 agent 는 `default` 묵시 처리; 본 SPEC 은 명시 선언만 검증 (frontmatter 강제 추가는 SPEC-V3R2-ORC-001 의 영역).
- `.claude/rules/moai/core/settings-management.md` — SPEC-V3R2-CON-003 의 ownership.

### 3.4 Reference citations (file:line)

Per `spec.md` §10 traceability 와 plan-auditor 의 file:line 앵커 ≥10 요구사항.

1. `spec.md:44-55` (in-scope items 1-9 — typed permission stack 표면)
2. `spec.md:101-148` (24 EARS REQs — REQ-001..051)
3. `spec.md:149-163` (15 ACs — AC-01..15)
4. `spec.md:165-172` (constraints — Go 1.22+, no new direct deps, 500µs p99, 256 KiB 메모리, AskUserQuestion 채널)
5. `internal/permission/stack.go:13-65` (기존 PermissionMode enum 5-value 정의)
6. `internal/permission/stack.go:67-86` (기존 PermissionRule struct + Origin 필드)
7. `internal/permission/stack.go:88-140` (기존 Matches() pattern matcher)
8. `internal/permission/stack.go:182-233` (기존 PreAllowlist 8 패턴 시드)
9. `internal/permission/stack.go:244-269` (기존 IsWriteOperation — M3 보강 대상)
10. `internal/permission/resolver.go:17-48` (기존 ResolveContext struct)
11. `internal/permission/resolver.go:135-254` (기존 Resolve 메인 분기 — M3/M4 보강)
12. `internal/permission/resolver.go:209-219` (기존 fork depth >3 게이트)
13. `internal/permission/resolver.go:258-299` (기존 checkTier — M3 conflict tiebreak)
14. `internal/permission/resolver.go:383-397` (기존 ValidateMode — M2 spawn wire)
15. `internal/permission/bubble.go:46-50` (기존 BubbleDispatcher struct)
16. `internal/permission/bubble.go:94-102` (기존 DispatchToParent placeholder — M4 IPC contract)
17. `internal/permission/bubble.go:138-144` (기존 ValidateForkDepth)
18. `internal/permission/bubble.go:150-154` (기존 IsParentAvailable placeholder)
19. `internal/cli/doctor_permission.go:1-91` (기존 doctor permission CLI 전체)
20. `internal/template/agent_frontmatter_audit_test.go` (기존 walker — M2 확장)
21. `.moai/config/sections/security.yaml:1-24` (기존 security 섹션 — M2 확장)
22. `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary (subagent 금지 규칙 — bubble routing 의 orchestrator-only 정합성)
23. `.claude/rules/moai/workflow/spec-workflow.md:172-204` (Plan Audit Gate)
24. `CLAUDE.local.md:§6` (test isolation — t.TempDir() + filepath.Abs)
25. `CLAUDE.local.md:§14` (no hardcoded paths in `internal/`)
26. `CLAUDE.local.md:§15` (template language neutrality — security.yaml 미러 일치)
27. `docs/design/major-v3-master.md:§4.3 Layer 3` (PermissionMode 5-enum + Source 8-tier 타입블록)
28. `docs/design/major-v3-master.md:§5.2` (Multi-Source Permission Resolution priority chain)
29. `docs/design/major-v3-master.md:§8 BC-V3R2-015` (multi-layer settings reader 누적 layering)
30. `.moai/design/v3-redesign/synthesis/problem-catalog.md` Cluster 5 (P-C01 no permission bubble, P-C04 no provenance)

Total: **30 distinct file:line anchors** (exceeds the §Hard-Constraints minimum of 10).

---

## 4. Technology Stack Constraints

Per `spec.md` §7 Constraints, **새로운 외부 의존성 0**.

- Go 1.22+ (이미 `go.mod` 가 1.26 require — go.mod:3 line 검증 완료).
- `github.com/spf13/cobra` — CLI 의존성, 이미 `go.mod` 에 존재 (`internal/cli/doctor_permission.go:7` 에서 import).
- `github.com/go-playground/validator/v10` — RT-004/SCH-001 동반 도입; 본 SPEC 은 import 안 함 (resolver 자체는 Go 타입 enum 으로 검증).
- `internal/config.Source` 에서 8-tier enum import — RT-005 의 unique ownership.
- `internal/hook.HookResponse` import — RT-001 의 unique ownership.

추가 surface:

- 7 new Go files under `internal/permission/`, `internal/cli/` (per §3.2).
- ~14 modified Go files (per §3.1).
- 1 CHANGELOG entry.
- 7 MX tags across 5 files (per §6).
- 1 security.yaml 새 키 블록 + template mirror.

**No new directory structures** — `internal/permission/` 이미 존재.
**No new YAML schema files** — security.yaml 의 `permission:` 블록만 확장.

---

## 5. Risk Analysis & Mitigations

Extends `spec.md` §8 risks with concrete file-path mitigations.

| Risk | Probability | Impact | Mitigation Anchor |
|------|-------------|--------|-------------------|
| Bubble dispatch 실제 IPC 가 RT-001 hook channel 미정 | M | M | M4 에서 contract 만 동결 (DispatchToParent placeholder 유지). 실제 wire 는 orchestrator integration (스킬 본문) — 본 SPEC 범위 외이지만 contract API freezing → 이후 hookup 무중단. |
| Pre-allowlist 8 패턴 부족 → bubble fatigue 잔존 | M | L | telemetry 채널 (post-beta.1) 로 측정 후 SPEC-V3R2-RT-002 amendment v0.2.0 으로 패턴 추가. master §10 R7. |
| `IsWriteOperation` 의 strings.Contains 매처 false-positive (예: bash command 안의 `echo ` 가 항상 write 로 분류) | H | M | M3 에서 매처 정밀화 — 명령 시작점 기반 매칭 (`strings.HasPrefix(strings.TrimSpace(input), "echo ")`); 일부 패턴은 정규식으로 변경. |
| Conflict tiebreak 의 specificity 점수 함수가 사용자 기대와 어긋남 | M | L | `internal/permission/conflict.go` 가 explicit specificity formula (와일드카드 개수 카운트) 를 godoc 으로 명시 + `moai doctor permission --trace` 가 specificity 점수 dump. |
| Fork depth registry (parent session liveness) 미정 | M | L | M4 에서 `IsParentAvailable` 를 ParentSessionID 비어있지 않음 으로 단순화; registry 통합은 RT-004 SessionStore 와 동시 진행 시 wire (cross-SPEC parallel 진행 가정). |
| Legacy `bypassPermissions` action 마이그레이션이 silent → 사용자 혼란 | M | M | `MigrateLegacyBypassRules` 가 stderr WARN + `.moai/logs/permission.log` 양쪽에 deprecation 메시지 emit (Origin file path 명시). M5b. |
| Agent frontmatter strict-validation 이 기존 미선언 agent 를 break | L | H | walker 가 명시 선언만 검증; 미선언은 default 묵시 — 기존 ~17 agent 중 permissionMode 명시 ≤3 개 추정 (manual sample 검증). 추가로 `--allow-implicit-default` 플래그 (audit test 옵션) 로 후행 가능. |
| Pattern matcher 가 pattern 와일드카드 표현력 부족 (e.g., `Bash(rm -rf /[!.]*)` brace 표현 미지원) | L | L | spec.md §3 의 patten 표현력은 v3.0 범위 내; 추가 표현력은 v3.1+ amendment. |
| `moai doctor permission --trace` JSON 출력이 `Tier: config.Source(999)` (hook tier sentinel) 를 surface | L | L | M5a 에서 hook tier 출력 시 `"tier": "hook"` 으로 stringify. trace JSON 후처리 1줄 추가. |
| 동시성: `PermissionResolver.mu` 가 RWMutex 인데 재귀 호출 (hook UpdatedInput re-match) 시 데드락 가능성 | M | H | M3 의 re-match 가 `r.mu.RUnlock()` 후 `r.Resolve()` 재호출이 아닌 동일 lock 안에서 inline 처리 — 현 skel resolver.go:147-160 을 inline 으로 검증 + 테스트. |
| Concurrent test races on slow CI (Windows) | L | L | 기존 resolver_test.go 의 happy-path 가 race-free; 신규 conflict_test.go/migration_test.go 도 단일 goroutine 으로 작성. `go test -race` 통과 확인. |

---

## 6. mx_plan — @MX Tag Strategy

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` 와 `.claude/skills/moai/workflows/plan.md` mx_plan MANDATORY rule.

### 6.1 @MX:ANCHOR targets (high fan_in / contract enforcers)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/permission/resolver.go:Resolve` | `@MX:ANCHOR fan_in=N - SPEC-V3R2-RT-002 REQ-002, REQ-004, REQ-005 enforcer; every tool invocation walks the 8-tier stack through this method. Tier order, hook overlay, fork-depth gate, non-interactive fail-closed, plan-mode write-deny are all single-path; touching this affects every agent in moai-adk-go.` | Resolve 는 모든 tool 호출의 단일 결정 진입점. 다운스트림 영향이 시스템 전체. |
| `internal/permission/stack.go:PreAllowlist` | `@MX:ANCHOR fan_in=2 - SPEC-V3R2-RT-002 REQ-006 contract; the 8 dev-op patterns embedded here are the SrcBuiltin tier and absorb ~80% of bubble fatigue. Adding/removing requires a SPEC and impacts every agent's first-call experience.` | Pre-allowlist 변경은 모든 agent UX 에 직접 영향. |
| `internal/permission/bubble.go:DispatchToParent` | `@MX:ANCHOR fan_in=2 - SPEC-V3R2-RT-002 REQ-012, REQ-050 contract; bubble routing IPC is THE only channel between fork agents and the parent terminal's AskUserQuestion. Changing the wire format breaks every fork agent simultaneously. Coordinated with RT-001 hook channel.` | Bubble dispatch 는 fork ↔ parent 의 single IPC contract. |

### 6.2 @MX:NOTE targets (intent / context delivery)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/permission/stack.go:PermissionMode` | `@MX:NOTE - SPEC-V3R2-RT-002 PermissionMode 5-enum (default/acceptEdits/bypassPermissions/plan/bubble). bubble is a first-class peer, not a flag — fork agents that inherit parent context bubble decisions to the parent's AskUserQuestion channel rather than silently defaulting. Violation: agent_frontmatter_audit_test.go strict-validation lint (REQ-008).` | 5-enum 의 의도와 bubble 의 일등 시민 위상을 코드 옆에 박제. |
| `internal/permission/conflict.go:resolveConflict` | `@MX:NOTE - SPEC-V3R2-RT-002 REQ-042 same-tier conflict resolution. Specificity = wildcard count의 inverse; ties → file-system scan order (lexicographic). All conflicts logged to .moai/logs/permission.log per REQ-042 sentinel.` | conflict tiebreak 의 결정 규칙을 향후 reader 에게 전달. |

### 6.3 @MX:WARN targets (danger zones)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/permission/resolver.go:147-160 hook UpdatedInput re-match` | `@MX:WARN @MX:REASON - SPEC-V3R2-RT-002 REQ-013 hook UpdatedInput re-match. Re-running Resolve() requires HookResponse=nil to prevent infinite loop. Single re-execution allowed — nested mutate (UpdatedInput inside UpdatedInput) is rejected. Touching this risks recursive re-entry deadlock.` | 가장 쉽게 회귀가 발생할 수 있는 재귀 분기. |
| `internal/permission/resolver.go:241-247 non-interactive fail-closed` | `@MX:WARN @MX:REASON - SPEC-V3R2-RT-002 REQ-041 non-interactive fail-closed. When IsInteractive=false AND no rule matches, default decision MUST be Deny (not Ask). Bypassing this opens permission ambiguity in CI/CD where AskUserQuestion is unreachable. Log path .moai/logs/permission.log is contract.` | CI/CD 안전성의 핵심 경계 — 비활성화 시 silent allow 위험. |

### 6.4 @MX:TODO targets (intentionally NONE for this SPEC)

본 SPEC 은 audit-ready 구현. M5d 종료 시까지 모든 작업이 GREEN 으로 수렴. 어떤 `@MX:TODO` 도 RT-002 implementation phase 안에서 해결되어야 함 (per `.claude/rules/moai/workflow/mx-tag-protocol.md` GREEN-phase resolution rule).

### 6.5 MX tag count summary

- @MX:ANCHOR: 3 targets
- @MX:NOTE: 2 targets
- @MX:WARN: 2 targets
- @MX:TODO: 0 targets
- **Total**: 7 MX tag insertions planned across 5 distinct files.

---

## 7. Worktree Mode Discipline

[HARD] All run-phase work for SPEC-V3R2-RT-002 executes in:

```
/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-002
```

Branch: `plan/SPEC-V3R2-RT-002` (already checked out per session context). Run-phase agent will continue on the same branch or create a sibling branch `feature/SPEC-V3R2-RT-002-permission-stack` per `CLAUDE.local.md` §18.2 branch naming.

[HARD] Worktree is used for this SPEC. All Read/Write/Edit tool invocations use absolute paths under the worktree root.

[HARD] `make build` and `go test ./...` execute from the worktree root: `cd /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-002 && make build && go test ./...`.

> Note: Run-phase agent operates from the actual worktree cwd; absolute paths shown for reference only. The worktree-root resolves to the directory returned by `git -C <worktree> rev-parse --show-toplevel` at run time.

---

## 8. Plan-Audit-Ready Checklist

These criteria are checked by `plan-auditor` at `/moai run` Phase 0.5 (Plan Audit Gate per `spec-workflow.md:172-204`). The plan is **audit-ready** only if all are PASS.

- [x] **C1: Frontmatter v0.1.0 schema** — `spec.md` frontmatter has all required fields (`id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `dependencies`, `bc_id`, `breaking`, `lifecycle`, `tags`). Verified by reading `spec.md:1-24`.
- [x] **C2: HISTORY entry for v0.1.0** — `spec.md:30-32` HISTORY table has v0.1.0 row with description.
- [x] **C3: 24 EARS REQs across 6 categories** — `spec.md:101-148` (Ubiquitous 8, Event-Driven 6, State-Driven 4, Optional 3, Unwanted 4, Complex 2).
- [x] **C4: 15 ACs all map to REQs (100% coverage)** — `spec.md:149-163`. 모든 AC 가 REQ 인용을 명시. 본 plan §1.5 traceability matrix 가 27 REQ → 15 AC → 28 task 매핑 확인.
- [x] **C5: BC scope clarity** — `spec.md:21` (`breaking: false`) + `spec.md:16 bc_id: []` + spec.md §1 (additive — 기존 flat permissions.allow read 호환은 RT-005 reader 가 보장).
- [x] **C6: File:line anchors ≥10** — research.md §10 (cited internally), this plan.md §3.4 (30 anchors).
- [x] **C7: Exclusions section present** — `spec.md:56-63` Out of scope (6 entries explicitly mapped to other SPECs: RT-001/RT-003/RT-005/RT-004/v3.1+/AskUserQuestion UX).
- [x] **C8: TDD methodology declared** — this plan §1.3 + `.moai/config/sections/quality.yaml` `development_mode: tdd`.
- [x] **C9: mx_plan section** — this plan §6 (7 MX tag insertions across 4 categories — 3 ANCHOR + 2 NOTE + 2 WARN + 0 TODO).
- [x] **C10: Risk table with mitigations** — `spec.md:174-184` (6 risks) + this plan §5 (11 risks, file-anchored mitigations).
- [x] **C11: Worktree mode path discipline** — this plan §7 (3 HARD rules, worktree-mode per session context).
- [x] **C12: No implementation code in plan documents** — verified self-check: this plan, research.md, acceptance.md, tasks.md contain only natural-language descriptions, regex patterns, file paths, code-block templates, and pseudo-Go for interface declarations. No executable Go function bodies.
- [x] **C13: Acceptance.md G/W/T format with edge cases** — verified in acceptance.md AC-01..15.
- [x] **C14: tasks.md owner roles aligned with TDD methodology** — verified in tasks.md M1-M5 (manager-tdd / expert-backend / manager-docs assignments).
- [x] **C15: Cross-SPEC consistency** — blocked-by dependencies verified: SPEC-V3R2-RT-001 (hook protocol — provides PermissionDecision enum + HookResponse — at-risk if not yet merged at run-phase plan-audit gate; mitigation: RT-002 imports `internal/hook.HookResponse` 만 사용, 정의는 RT-001 unique ownership), SPEC-V3R2-RT-005 (Source enum + reader — merged status needed), SPEC-V3R2-CON-001 (FROZEN zone declaration — 8-source ordering as constitutional clause). RT-002 blocks RT-003 (sandbox launcher uses PermissionMode), ORC-001 (agent roster cleanup uses validated permissionMode), ORC-004 (worktree+acceptEdits at project tier).

All 15 criteria PASS → plan is **audit-ready**.

---

## 9. Implementation Order Summary

Run-phase agent executes in this order (P0 first, dependencies resolved):

1. **M1 (P0)**: ~10 새 테스트 케이스 across `internal/permission/resolver_test.go`, `internal/cli/doctor_permission_test.go`, `internal/template/agent_frontmatter_audit_test.go`. RED 모두 확인; 기존 테스트 GREEN 유지.
2. **M2 (P0)**: PreAllowlist sync.Once 변환 + ValidateMode spawn-side wire (`internal/permission/spawn.go` 신규) + frontmatter strict lint + security.yaml 신규 키 + template mirror. AC-07, AC-09 GREEN.
3. **M3 (P0)**: hook UpdatedInput re-match 가드 + IsWriteOperation 패턴 보강 + `internal/permission/conflict.go` 신규 (specificity-then-fs-order tiebreak). AC-04, AC-12, AC-13 GREEN.
4. **M4 (P0)**: bubble.go DispatchToParent IPC contract + IsParentAvailable contract + fork depth >3 systemMessage sentinel + non-interactive fail-closed log. AC-08, AC-14, AC-15 GREEN.
5. **M5 (P1)**: doctor_permission.go `--all-tiers --mode --fork --format json` 보강 + `internal/permission/migration.go` (legacy bypassPermissions deprecation) + session_rules SrcSession tier 적재 contract + CHANGELOG + 7 MX tags + final `make build` + `go test ./...`. AC-05, AC-10, AC-11, AC-15 GREEN. progress.md 갱신.

Total milestones: 5. Total file edits (existing): ~14. Total file creations (new): 7. Total CHANGELOG entries: 1. Total MX tag insertions: 7.

---

End of plan.md.
