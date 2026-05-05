---
id: SPEC-V3R3-RETIRED-AGENT-001
title: Retired Agent Stub Compatibility + manager-cycle Template Alignment
version: "0.1.0"
status: draft
created_at: 2026-05-04
updated_at: 2026-05-04
author: MoAI Plan Workflow
priority: P0
labels: [agent-runtime, templates, retired-stub, manager-cycle, manager-tdd, hooks, bug-fix, v3r3]
issue_number: null
phase: "v3.0.0 — Phase 7 — Agent Runtime Robustness"
module: "internal/template/templates/.claude/agents/moai/, internal/template/templates/.claude/hooks/moai/, internal/hook/, internal/cli/"
dependencies:
  - SPEC-V3R2-ORC-001
related_specs:
  - SPEC-V3R3-HYBRID-001
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "agent-runtime, retired-stub, manager-cycle, template-parity, hook-guard, worktree-empty-object, defect-chain, v3r3"
---

# SPEC-V3R3-RETIRED-AGENT-001: Retired Agent Stub Compatibility + manager-cycle Template Alignment

## HISTORY

| Version | Date       | Author              | Description                                                                                                            |
|---------|------------|---------------------|------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow  | 최초 작성. mo.ai.kr 사이드 프로젝트에서 발생한 5-layer defect chain (retired stub frontmatter → 0 tool_uses → worktreePath empty object → path `{}/{}`) 대응. P0 fix. |

---

## 1. Goal (목적)

`/Users/goos/MoAI/mo.ai.kr` 자매 프로젝트에서 2026-05-04 21:14:54 발생한 critical agent runtime bug — `Agent({subagent_type: "manager-tdd", isolation: "worktree", ...})` 호출이 11.4s에 0 tool_uses + `worktreePath: {}` (empty object) + `worktreeBranch: undefined` 로 반환되어 후속 fallback이 broken state를 propagate, 결국 `[ERROR] Path "/Users/goos/MoAI/mo.ai.kr/{}/{}" does not exist` 으로 manager-cycle 종료한 사건 — 의 5-layer defect chain을 차단한다.

핵심 시정 조치 (P0/P1/P2 분리):

- **P0 (Immediate fix, 본 SPEC 핵심)**:
  1. moai-adk-go template에 `manager-cycle.md` 추가 (현재 absent — `internal/template/templates/.claude/agents/moai/manager-cycle.md` 존재하지 않음)
  2. `manager-tdd.md` retired stub frontmatter standardization (`retired: true`, `retired_replacement: manager-cycle`, `retired_param_hint`, `tools: []`, `skills: []`)
  3. SubagentStart hook retired-rejection guard (`internal/hook/agent_start.go` 신규 + `handle-subagent-start.sh.tmpl` 통합)
- **P1 (Agent runtime robustness)**:
  4. `worktreePath` empty-object guard (Agent() 호출 결과 검증 wrapper)
  5. Type-safe path interpolation (`text/template` 기반, string concat 폐기)
- **P2 (Documentation + Lessons)**:
  6. `lessons.md #11` retired stub anti-pattern + 5-layer defect chain 기록

본 SPEC은 mo.ai.kr 사고가 **moai-adk-go template 자체의 inconsistency**에서 출발했음을 직접 시정한다 (retired stub만 deploy + manager-cycle 미배포 = 항상 broken state). 사용자 의도 그대로:

- mo.ai.kr 자매 프로젝트에서 동일 사고가 다시 일어나지 않도록 **template-side fix만으로 종결 가능한 layer (P0)** 를 우선 수행.
- runtime guard (P1)는 template 외에도 Go runtime + Claude Code Agent() runtime을 모두 건드리지만, P0가 차단하면 P1은 in-depth defense로 작동.
- P2 lessons는 향후 SPEC retirement decision 시 동일 실수 방지.

### 1.1 배경 — 5-Layer Defect Chain (mo.ai.kr 2026-05-04 21:14:54 사건)

#### 1차 — retired stub frontmatter가 Claude Code Agent runtime에 invalid

mo.ai.kr `/Users/goos/MoAI/mo.ai.kr/.claude/agents/moai/manager-tdd.md` (976 bytes stub):

```yaml
---
name: manager-tdd
description: "Retired — use manager-cycle with cycle_type=tdd"
status: retired         # ← custom field, runtime ignores it
# tools field absent → Agent runtime spawns with NO tool_uses
# skills field absent
# permissionMode field absent
---
```

Claude Code Agent runtime은 `name` 필드 존재만으로 valid agent로 간주하여 spawn하지만, `tools` / `skills` / `permissionMode` 부재로 0 tool_uses만 가능한 상태에서 11.4s 동안 retirement 메시지만 반환하고 termination signal 없이 종료한다. `status: retired`는 frontmatter spec에 정의되지 않은 custom field이므로 runtime이 인식하지 못한다.

#### 2차 — `Agent(isolation: "worktree")`는 agent body 실행 전에 worktree allocation

worktree allocation은 spawn 시점에 진행되며, retired stub은 실제 작업을 수행하지 않으므로 worktree allocation result가 정상적으로 caller에게 propagate되지 않는다. caller는 `worktreePath: {}` (empty object), `worktreeBranch: undefined` 반환을 받게 된다. v2.1.97의 worktree CWD isolation fix는 이 시나리오를 다루지 않는다.

#### 3차 — MoAI auto-fallback이 broken state를 검증 없이 propagate

retired response를 detect한 fallback logic이 `manager-cycle`로 re-delegation을 시도하지만, retired stub response에서 `worktreePath` / `worktreeBranch`를 그대로 추출하여 새 spawn prompt에 inject한다. empty object / undefined 검증이 없다.

#### 4차 — Path string interpolation produces literal `{}`

`/Users/goos/MoAI/mo.ai.kr/{worktreeBranch}/{worktreePath}` template에서:
- `worktreeBranch === undefined` → `"undefined"` 또는 `""` (JS template literal)
- `worktreePath === {}` (empty object) → `"[object Object]"` 또는 `"{}"` (JSON.stringify shorthand 또는 toString)

실제 관측된 path: `"/Users/goos/MoAI/mo.ai.kr/{}/{}"` — `{}` 두 번 나타남. shell heredoc 또는 string template이 empty object를 `{}` 로 stringify했음.

#### 5차 — Stream idle partial (side pattern, 직접 cause 아님)

3000-token spawn prompts가 13s 처리 동안 `stream_idle_partial` warning 20+ trigger. `feedback_large_spec_wave_split.md` memory에 기록된 known pattern. 직접 cause는 아니나 debugging 어렵게 만듦.

### 1.2 비교 증거 (Comparative Evidence)

| File | mo.ai.kr (bug-affected) | moai-adk-go (tool source) |
|------|-------------------------|---------------------------|
| `manager-tdd.md` | 976 bytes retired stub | 6407 bytes full definition |
| `manager-ddd.md` | 1000 bytes retired stub | 7628 bytes full definition |
| `manager-cycle.md` | 10245 bytes (unified DDD/TDD) | **MISSING** ⚠️ |

검증 명령:
```
$ ls -la /Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/agents/moai/manager-cycle.md
ls: ... No such file or directory
$ ls -la /Users/goos/MoAI/mo.ai.kr/.claude/agents/moai/manager-cycle.md
-rw-r--r-- 10245 May 1 13:51 ...
```

**핵심 발견**: `manager-cycle.md`가 moai-adk-go template에 없다. mo.ai.kr는 manager-cycle.md를 어딘가 외부 소스에서 deploy 받았으나 manager-tdd.md / manager-ddd.md는 retired stub로 변경되었다. SPEC-V3R2-ORC-001 (Agent Roster Consolidation)이 incomplete 상태로 적용된 것.

### 1.3 비목표 (Non-Goals)

- mo.ai.kr 사이드 프로젝트의 broken state 직접 복구 (mo.ai.kr는 본 SPEC 후 `moai update`로 자동 sync)
- Claude Code Agent runtime의 frontmatter 검증 강화 요청 (Anthropic 측 변경 필요 — 본 SPEC 범위 밖)
- `status` custom frontmatter field를 Claude Code spec에 추가 요청 (외부 의존)
- `manager-tdd` 명령어 자체의 부활 (retirement decision은 SPEC-V3R2-ORC-001에서 완료, 본 SPEC은 retirement 호환성 fix만)
- `manager-ddd` retired stub 같은 시정 (본 SPEC은 manager-tdd 케이스에 집중; manager-ddd 동등 fix는 동일 패턴이지만 별도 검증 필요)
- 모든 retired agent 일반화된 standardization 정책 (단일 케이스 fix가 우선; 일반화는 후속 SPEC에서)
- agency/* retired agent 일괄 처리 (SPEC-AGENCY-ABSORB-001 영역)
- mo.ai.kr 외 다른 사용자 프로젝트의 retroactive fix
- Worktree CWD isolation fix beyond v2.1.97 (Anthropic 측 변경 필요)
- Stream idle partial root cause (`feedback_large_spec_wave_split.md` 영역, wave-split 권장으로 우회)
- 신규 agent retirement workflow 정의 (SPEC retirement governance는 별도 SPEC)

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns (NEW files)**:
  - `internal/template/templates/.claude/agents/moai/manager-cycle.md` (NEW) — mo.ai.kr 10245-byte 버전 reference + moai-adk-go quality 기준 검증
  - `internal/hook/agent_start.go` (NEW) — SubagentStart event handler with retired-rejection guard
  - `internal/hook/agent_start_test.go` (NEW) — TDD test surface for retired-stub detection
  - `internal/template/agent_frontmatter_audit_test.go` (NEW) — moai/* agents의 retired frontmatter standardization audit
  - `.moai/specs/SPEC-V3R3-RETIRED-AGENT-001/` 5-7 SPEC artifacts (본 plan-phase 산출물)

- **Owns (MODIFIED files)**:
  - `internal/template/templates/.claude/agents/moai/manager-tdd.md` — retired stub frontmatter standardization (P0 #2)
  - `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl` — retired-rejection guard 호출 통합 (P0 #3)
  - `internal/cli/launcher.go` (또는 Agent() wrapper layer) — worktreePath empty-object guard (P1 #4)
  - `internal/template/templates/.claude/agents/moai/manager-strategy.md` (line ref `manager-tdd`) — 인용 정합 확인 (drive-by 금지)
  - `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` — Manager Agents (8) → 7 (manager-tdd retired) + Frontmatter Format Rules에 `retired:` 필드 추가
  - `internal/template/templates/.claude/rules/moai/core/agent-hooks.md` — manager-tdd row를 manager-cycle row로 교체
  - `internal/template/templates/.claude/agents/moai/manager-ddd.md` — `manager-tdd` 언급(line 14)을 `manager-cycle` 으로 substitution

- **Modifies (read-only references, 검증만)**:
  - `internal/template/templates/CLAUDE.md §4 Manager Agents (8)` — agent count 정합 확인
  - `internal/template/templates/CLAUDE.md §5 Agent Chain` — manager-tdd 언급 substitution
  - `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` — `manager-ddd/tdd` 언급을 `manager-cycle` 으로

- **CHANGELOG**: 본 SPEC은 BREAKING CHANGE 없음 (P0 fix는 backward-compatible: retired stub은 여전히 retirement message 반환하지만 추가 metadata로 runtime이 더 빠르게 reject). `bc_id: []`.

- **Test surface**:
  - `internal/template/agent_frontmatter_audit_test.go` (REQ-RA-001, REQ-RA-002): retired frontmatter standardization audit
  - `internal/hook/agent_start_test.go` (REQ-RA-007, REQ-RA-008): retired-rejection guard
  - `internal/hook/agents/factory_test.go` extension (REQ-RA-009): factory dispatch for `agent-start` event
  - `internal/template/manager_cycle_present_test.go` (NEW, REQ-RA-003): asserts manager-cycle.md exists in embedded FS

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 구현 코드 (Go function body, hook handler logic detailed implementation) — 본 SPEC은 plan-phase 산출물; 실제 구현은 `/moai run SPEC-V3R3-RETIRED-AGENT-001` 단계.
- mo.ai.kr 사이드 프로젝트 직접 수정 (mo.ai.kr는 본 SPEC merge 후 `moai update` 실행으로 sync 받음)
- `manager-ddd` retired stub의 동등 standardization (동일 패턴 적용 가능하지만 별도 SPEC; 본 SPEC은 manager-tdd case 집중)
- `manager-tdd` 부활 또는 cycle 설계 변경 (SPEC-V3R2-ORC-001 retirement decision 유지)
- agency/* (copywriter, designer 등) retired agent 일괄 표준화 (SPEC-AGENCY-ABSORB-001 영역)
- Claude Code Agent() runtime 자체의 frontmatter 검증 강화 요청 (Anthropic 외부 의존)
- `status: retired` custom field를 Claude Code 공식 frontmatter spec에 추가 (외부 의존)
- 5-layer defect chain의 layer 5 (`stream_idle_partial`) 직접 fix (`feedback_large_spec_wave_split.md` 영역)
- v2.1.97 미만 버전에 대한 worktree CWD isolation backport
- 모든 sub-agent 호출에 worktreePath validation을 적용 (worktree-isolated agent에 한정)
- `text/template` 마이그레이션을 Agent() runtime 외 영역으로 확대 (string concat → text/template은 path interpolation 한정)
- 신규 agent retirement workflow / governance 문서화 (별도 SPEC; 본 SPEC은 단일 케이스 fix)
- mo.ai.kr 직접 모니터링 / alert 시스템 구축

---

## 3. Environment (환경)

- 런타임: Go 1.23+, Cobra CLI, Claude Code Agent() runtime v2.1.97+ (worktree CWD isolation 의존), bash hook wrappers.

- 영향 디렉터리:
  - 신규: `internal/template/templates/.claude/agents/moai/manager-cycle.md`, `internal/hook/agent_start.go`, `internal/hook/agent_start_test.go`, `internal/template/agent_frontmatter_audit_test.go`, `internal/template/manager_cycle_present_test.go`
  - 수정: `internal/template/templates/.claude/agents/moai/{manager-tdd.md, manager-ddd.md, manager-strategy.md}`, `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl`, `internal/template/templates/.claude/rules/moai/{development/agent-authoring.md, core/agent-hooks.md}`, `internal/template/templates/CLAUDE.md` (§4, §5), `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`, `internal/cli/launcher.go` (또는 Agent() wrapper)
  - Reference (no edit): `internal/hook/agents/factory.go` (factory dispatch confirm), `internal/template/embedded.go` (auto-regenerated by `make build`)

- 외부 endpoints / 외부 시스템:
  - Claude Code Agent() runtime: spawn 시점에 frontmatter parse + worktree allocation; behavior는 Anthropic Claude Code release notes (v2.1.97 worktree CWD isolation fix) 기반.
  - SubagentStart hook event: Claude Code v2.1.69+ `agent_id` 필드 포함; v2.1.84+ TaskCreated hook도 dispatch.
  - 외부 reference 없음 (인터넷 endpoint 호출 없음).

- 외부 레퍼런스 (codebase grounded scan):
  - `internal/template/templates/.claude/agents/moai/manager-tdd.md:1-39` (current full definition, 6407 bytes)
  - `mo.ai.kr/.claude/agents/moai/manager-tdd.md:1-35` (deployed retired stub, 976 bytes — 비교 evidence)
  - `mo.ai.kr/.claude/agents/moai/manager-cycle.md:1-?` (deployed unified agent, 10245 bytes — reference for new template file)
  - `internal/template/templates/.claude/hooks/moai/handle-subagent-start.sh.tmpl:1-35` (current bash wrapper, no retired-guard logic)
  - `internal/hook/agents/factory.go` (handler factory dispatch — agent-start event integration target)
  - `internal/template/templates/.claude/rules/moai/core/agent-hooks.md:38-48` (Agent Hook Actions table — manager-tdd row to update)
  - `internal/template/templates/.claude/rules/moai/development/agent-authoring.md:108-118` (Manager Agents (8) — agent count update)
  - `internal/template/templates/CLAUDE.md` Manager Agents (8) section — count + manager-cycle.md addition
  - `internal/template/templates/.claude/agents/moai/manager-strategy.md` line containing `manager-tdd` — substitution target

---

## 4. Assumptions (가정)

- mo.ai.kr의 `manager-cycle.md` (10245 bytes, dated May 1 13:51) is the canonical reference for the unified DDD/TDD agent. moai-adk-go template은 이 파일을 quality 검증 후 그대로 import할 수 있다 (파일 contents는 reference로만 사용; moai-adk-go style/anti-bias 규칙은 final commit 전 검증).
- SPEC-V3R2-ORC-001 (Agent Roster Consolidation)이 manager-tdd / manager-ddd retirement decision을 이미 내렸으므로, 본 SPEC은 retirement 자체를 다시 결정하지 않고 retirement compatibility만 fix한다.
- Claude Code Agent() runtime은 invalid frontmatter (e.g., `tools: []` empty array)에 대해 graceful 처리하며 0 tool_uses만 가능한 silent agent를 spawn한다 (mo.ai.kr 21:14:54 사건의 1차 layer 관측 결과).
- SubagentStart hook은 spawn 시점에 fire되며 hook이 exit 2 + JSON `{"decision":"block","reason":"..."}` 반환 시 spawn을 차단한다 (Claude Code hooks-system.md spec 기반; PreToolUse blocking semantic은 SubagentStart에도 적용된다고 가정 — research.md §2에서 verify 필요).
- `worktreePath` 의 empty-object 반환은 retired stub spawn의 직접 결과이며, P0 #1+#2가 retired stub의 정상 termination을 보장하면 P1 #4 worktreePath validation은 in-depth defense layer로 작동한다 (P0 단독 차단으로도 5-layer chain은 깨진다).
- moai-adk-go의 `internal/template/templates/.claude/agents/moai/` 트리는 16-language neutral하다 (CLAUDE.local.md §15) — agent definition은 language-agnostic이며, 신규 manager-cycle.md도 language-neutral content만 포함한다 (mo.ai.kr 버전을 검증 후 import 시 confirm).
- `manager-cycle.md` 추가는 Manager Agents count를 8 → 9로 변경하지 않는다 (manager-tdd retired = effective count 7; manager-cycle 추가 = effective count 8). Documentation의 "Manager Agents (8)" 표기는 효과적 active agents 기준으로 유지.
- 사용자가 `moai update` 실행 시 mo.ai.kr는 본 SPEC의 fix를 자동 sync 받는다 (P0 #1 manager-cycle.md NEW + P0 #2 manager-tdd.md MODIFIED 반영). 사용자 데이터 (`.moai/specs/`, `.moai/project/`) 는 보존된다 (CLAUDE.local.md §2 protected directories).
- TDD methodology 적용 (`.moai/config/sections/quality.yaml`): RED (M1 test scaffolding) → GREEN (M2-M4 implementation) → REFACTOR (M5 documentation + lessons).

---

## 5. Requirements (EARS 요구사항)

총 **16개 REQs** — Ubiquitous 6, Event-Driven 4, State-Driven 3, Optional 1, Unwanted 2.

### 5.1 Ubiquitous Requirements (always active)

**REQ-RA-001**
The MoAI template tree **shall** include `internal/template/templates/.claude/agents/moai/manager-cycle.md` as a first-class active agent definition file with full frontmatter (name, description, tools, model, permissionMode, memory, skills, hooks).

**REQ-RA-002**
The MoAI template tree **shall** standardize all retired agent stub files (currently `manager-tdd.md`; pattern applicable to future retirements) with the canonical retirement frontmatter: `retired: true` (boolean), `retired_replacement: <agent-name>` (string), `retired_param_hint: <invocation-hint>` (string), `tools: []` (explicit empty), `skills: []` (explicit empty).

**REQ-RA-003**
The MoAI embedded FS regeneration (`make build`) **shall** include `manager-cycle.md` and the standardized `manager-tdd.md` retired stub. A test (`internal/template/manager_cycle_present_test.go`) **shall** verify both files exist in the embedded FS at every CI run.

**REQ-RA-004**
The MoAI template tree **shall** provide a SubagentStart hook handler (`internal/hook/agent_start.go`) that detects retired-frontmatter agents pre-spawn and returns `{"decision":"block","reason":"agent retired, use replacement: <retired_replacement>"}` with exit code 2.

**REQ-RA-005**
The Agent() invocation wrapper layer (`internal/cli/launcher.go` or equivalent) **shall** validate the `worktreePath` field of every Agent() return value when `isolation: "worktree"` was requested: empty object `{}`, `null`, `undefined`, or non-string values **shall** trigger an explicit error with sentinel `WORKTREE_PATH_INVALID` instead of being propagated.

**REQ-RA-006**
The MoAI agent runtime path interpolation **shall not** use ad-hoc string concatenation for paths derived from Agent() return values. Path templates **shall** use Go `text/template` (or equivalent type-safe templating) such that `{}` (empty object) values produce a typed error rather than the literal string `"{}"`.

### 5.2 Event-Driven Requirements

**REQ-RA-007**
**When** the SubagentStart hook receives an event for an agent whose definition file has `retired: true` in frontmatter, the hook **shall**: (a) parse the YAML frontmatter, (b) extract `retired_replacement` and `retired_param_hint`, (c) emit `{"decision":"block","reason":"agent <name> retired (SPEC-V3R2-ORC-001), use <retired_replacement> with <retired_param_hint>"}` to stdout, (d) exit with code 2.

**REQ-RA-008**
**When** the SubagentStart hook is invoked with an unknown agent name (file does not exist in `.claude/agents/moai/` or `.claude/agents/`), the hook **shall** allow the spawn (exit code 0, no decision field) — non-MoAI agents bypass the retired-rejection guard.

**REQ-RA-009**
**When** `internal/hook/agents/factory.go` is dispatched with `event = "agent-start"`, the factory **shall** route to the new `agent_start.go` handler. Unknown action falls through to `default_handler.go` (existing behavior preserved).

**REQ-RA-010**
**When** an Agent() call with `isolation: "worktree"` returns a value where `worktreePath` is empty object, `null`, or non-string, the wrapper layer **shall** raise `WORKTREE_PATH_INVALID` with the agent name + invocation context **and shall not** propagate the broken value to fallback re-delegation.

### 5.3 State-Driven Requirements

**While** retired-stub agent definition files exist in the template (status quo for manager-tdd):

**REQ-RA-011**
**While** `retired: true` is present in an agent frontmatter, the agent body content **shall** describe (a) the retirement reason, (b) the replacement agent name, (c) the migration command pattern (old → new invocation). Documentation lookup at retirement time **shall** be self-contained.

**REQ-RA-012**
**While** the SubagentStart retired-rejection guard is active, the response time **shall** be ≤ 500ms (single YAML parse + stdout write; no network calls, no agent spawn). The 11.4s observed in mo.ai.kr 21:14:54 incident **shall not** recur.

**REQ-RA-013**
**While** `manager-cycle.md` is the active unified DDD/TDD implementation agent, the documentation references in CLAUDE.md (§4 Manager Agents, §5 Agent Chain), `agent-authoring.md` (Manager Agents listing), `agent-hooks.md` (Agent Hook Actions table), `spec-workflow.md` (manager-ddd/tdd reference), and `manager-ddd.md` (line referring to manager-tdd) **shall** all reference `manager-cycle` (or `manager-cycle with cycle_type=tdd`) instead of `manager-tdd`. Inconsistencies **shall** be detected by `internal/template/agent_frontmatter_audit_test.go`.

### 5.4 Optional Requirements

**REQ-RA-014**
**Where** the user wants to introspect retired agents in their installation, the CLI **shall** support `moai agents list --retired` as an optional subcommand (P2; deferred to follow-up SPEC if not feasible in v3R3 first minor release window).

### 5.5 Unwanted Behavior Requirements

**REQ-RA-015 (Unwanted Behavior)**
**If** an Agent() invocation with `isolation: "worktree"` returns `worktreePath` as empty object, null, or non-string type, **then** the orchestrator **shall not** propagate that value to fallback re-delegation, **shall not** perform path string interpolation that would produce literal `{}` segments in the resulting path, **and shall not** silently swallow the error. The orchestrator **shall** raise the `WORKTREE_PATH_INVALID` sentinel and surface it to the user.

**REQ-RA-016 (Unwanted Behavior — Composite)**
**If** a future agent is retired, **then** the retirement workflow **shall** include all of: (a) standardized frontmatter per REQ-RA-002, (b) updated documentation references per REQ-RA-013, (c) `agent_frontmatter_audit_test.go` audit assertion, (d) corresponding active replacement file in template. **If** any one of (a)-(d) is missing, CI **shall** fail with `RETIREMENT_INCOMPLETE_<agent-name>`. **And** the SubagentStart guard (REQ-RA-007) **shall** never produce silent acceptance (exit 0) for an agent whose frontmatter contains `retired: true`.

---

## 6. Acceptance Criteria (수용 기준 요약, 18 ACs ≥ 1-to-1 REQ coverage)

상세 Given/When/Then은 `acceptance.md` 참조. Summary mapping:

- **AC-RA-01** (REQ-RA-001): manager-cycle.md exists in template + has full frontmatter
- **AC-RA-02** (REQ-RA-002): manager-tdd.md retired stub has all 5 standardized fields
- **AC-RA-03** (REQ-RA-003): `make build` regenerates embedded FS with both files
- **AC-RA-04** (REQ-RA-004, REQ-RA-007): SubagentStart hook returns block decision for retired agent
- **AC-RA-05** (REQ-RA-005, REQ-RA-010): launcher.go rejects empty-object worktreePath
- **AC-RA-06** (REQ-RA-006): path interpolation uses text/template, no string concat
- **AC-RA-07** (REQ-RA-007): retired-rejection guard returns proper JSON + exit 2
- **AC-RA-08** (REQ-RA-008): unknown agent name bypasses guard (exit 0)
- **AC-RA-09** (REQ-RA-009): factory.go dispatch for agent-start event
- **AC-RA-10** (REQ-RA-010): WORKTREE_PATH_INVALID sentinel emitted with context
- **AC-RA-11** (REQ-RA-011): retired stub body describes reason + replacement + migration
- **AC-RA-12** (REQ-RA-012): retired-rejection guard ≤500ms response time
- **AC-RA-13** (REQ-RA-013): all 6 documentation references substituted (CLAUDE.md §4, §5, agent-authoring.md, agent-hooks.md, spec-workflow.md, manager-ddd.md)
- **AC-RA-14** (REQ-RA-014): `moai agents list --retired` subcommand surfaced (or deferred with explicit decision)
- **AC-RA-15** (REQ-RA-016): CI rejects `RETIREMENT_INCOMPLETE_<agent>` if any of (a)-(d) missing
- **AC-RA-16** (REQ-RA-001, REQ-RA-013): manager-cycle.md spawn via Agent() succeeds with valid worktreePath
- **AC-RA-17** (REQ-RA-002, REQ-RA-007): manager-tdd retired stub spawn via Agent() blocked at SubagentStart layer (regression test for mo.ai.kr 5-layer chain)
- **AC-RA-18** (REQ-RA-005, REQ-RA-006): Agent() returning empty-object worktreePath triggers WORKTREE_PATH_INVALID instead of `[ERROR] Path "/{}/{}" does not exist`

---

## 7. Constraints (제약)

- **TDD methodology** (per `.moai/config/sections/quality.yaml development_mode: tdd`): RED (M1 test scaffolding) → GREEN (M2-M4 implementation) → REFACTOR (M5 documentation + lessons).
- **16-language neutrality** (CLAUDE.local.md §15): manager-cycle.md content는 language-agnostic. agent definition은 어느 user project 언어에도 적용 가능. (mo.ai.kr reference는 language-neutral 검증 후 import.)
- **Template-First HARD rule** (CLAUDE.local.md §2): `.claude/agents/moai/`, `.claude/hooks/moai/`, `.claude/rules/moai/` 변경은 모두 `internal/template/templates/` 미러 + `make build` 필수.
- **No drive-by refactor** (CLAUDE.md §7 Rule 2 / Agent Core Behavior #5): `manager-strategy.md`, `manager-ddd.md` 등 인접 파일은 substitution scope로 한정. 다른 개선 사항은 별도 SPEC.
- **No flat file** (manager-spec skill HARD rule): `.moai/specs/SPEC-V3R3-RETIRED-AGENT-001/` 디렉토리 + 5-7 files 필수 (spec.md, plan.md, acceptance.md, research.md, progress.md, optionally tasks.md, spec-compact.md).
- **API-key 보안 무관**: 본 SPEC은 API key / token / secret 다루지 않음 (단, retired-rejection guard는 stdin JSON 처리 시 secret leak 방지 위해 `last_assistant_message` 등 sensitive field 출력 금지).
- **Bash hook timeout** (`agent-authoring.md` § Bash Tool Timeout Ceiling): SubagentStart guard handler는 5s timeout 권장 (현행 `handle-subagent-start.sh.tmpl` timeout 5s 유지).
- **No flag in branch protection bypass**: 본 SPEC은 main branch protection rule (CLAUDE.local.md §18.7) 준수 — admin override 없이 PR 정상 리뷰 + CI green 후 merge.
- **Solo mode, no worktree** (사용자 directive): SPEC은 `feature/SPEC-V3R3-RETIRED-AGENT-001` 단일 브랜치에서 작성. Implementation 단계 (`/moai run`)에서 worktree 사용 여부는 별도 결정.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 확률 | 완화 |
|---|---|---|---|
| mo.ai.kr 사이드 프로젝트가 본 SPEC merge 후 `moai update` 미실행 시 same bug recur | H | M | docs-site + CHANGELOG에 "moai update 권장" 명시; 사용자 manual touch (CLAUDE.local.md §10 release process) 알림. |
| `manager-cycle.md` 신규 파일이 mo.ai.kr 버전 (10245 bytes)을 그대로 import 시 moai-adk-go style/anti-bias 규칙 위반 가능 | M | M | M2 implementation 단계에서 `agent-authoring.md` 규칙 + 16-language neutrality + manager-tdd/ddd 기존 패턴과 비교 검증; final commit 전 manual review. |
| SubagentStart hook timing이 spawn pipeline의 frontmatter parse보다 늦으면 retired-rejection이 effective하지 않음 | M | M | research.md §2에서 hook ordering 검증; Claude Code hooks-system.md SubagentStart spec 재확인; 미가용 시 PreToolUse hook으로 fallback. |
| `worktreePath` empty-object validation이 정상 case (worktree 없는 invocation)을 false-positive로 reject | M | L | validation 조건 명확화: `isolation: "worktree"` 명시 호출만 검증; 다른 isolation mode + 미설정은 skip. |
| `text/template` 마이그레이션 (REQ-RA-006)이 기존 path interpolation callsite을 광범위하게 건드림 → drive-by refactor 위험 | M | M | M3 단계에서 callsite 수 measure (target: ≤5); 5+ 발견 시 별도 SPEC으로 분리. |
| Manager Agents count (현행 8) 변경 시 광범위 documentation update 필요 | L | H | "8" 표기는 active agents 기준으로 유지: manager-tdd retired (-1) + manager-cycle 추가 (+1) = effective 8. 단순 substitution으로 처리. |
| `agent_frontmatter_audit_test.go` 추가가 기존 manager-ddd retired stub (mo.ai.kr 1000 bytes evidence)을 audit 대상으로 발견 → 본 SPEC scope 외 작업 발생 | M | H | manager-ddd retired stub 표준화는 본 SPEC scope 밖 (§1.2 명시). audit test는 manager-tdd만 검증; manager-ddd는 후속 SPEC `SPEC-V3R3-RETIRED-DDD-001` (가칭)에서 처리. |
| Claude Code Agent runtime이 `retired: true` custom frontmatter field를 invalid로 reject하여 retirement 자체가 깨짐 | H | L | research.md §2에서 frontmatter spec 재확인; Claude Code는 unknown field 무시 (graceful) — research 결과 기반 confirm. 미가용 시 `description` 필드에 retirement metadata 인코딩 (fallback). |
| Embedded template `make build` regeneration이 Go runtime cache 유발 → 즉시 effective되지 않음 | L | M | M5 단계에서 `make build && make install` 후 Claude Code restart 권장 (CLAUDE.local.md "Hard Constraints"). |
| 5-layer defect chain의 layer 5 (`stream_idle_partial`) 가 본 SPEC fix만으로 해결되지 않아 사용자가 동일 증상 재경험 | M | M | layer 5는 `feedback_large_spec_wave_split.md` lesson #9 영역 (wave-split 권장)으로 분리; 본 SPEC은 layer 1-4 차단. P0 차단으로 layer 5 trigger 자체가 줄어든다 (retired stub spawn 실패 → 13s prompt 처리 발생 안 함). |
| 본 SPEC merge 후 신규 retirement 시 standardized frontmatter 누락 → 동일 패턴 반복 | M | M | REQ-RA-015 (CI assertion) + agent_frontmatter_audit_test.go가 retirement decision 시점에 강제. |
| `moai update` template sync가 사용자 프로젝트의 manager-tdd.md 로컬 수정본을 overwrite | L | M | CLAUDE.local.md §2 protected directories는 `.moai/project/`, `.moai/specs/` 만 보호; `.claude/agents/` 는 template sync 대상. moai update가 backup 생성 후 sync (기존 동작). |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3R2-ORC-001** (Agent Roster Consolidation, status: completed): manager-tdd / manager-ddd retirement decision의 source. 본 SPEC은 이 retirement decision의 실행 incompleteness를 시정하는 follow-up.

### 9.2 Blocks

- 향후 `SPEC-V3R3-RETIRED-DDD-001` (가칭): manager-ddd retired stub standardization — 본 SPEC의 frontmatter standard (REQ-RA-002) 차용.
- 향후 agent retirement workflow / governance SPEC: 본 SPEC의 REQ-RA-015 CI assertion pattern을 base로.

### 9.3 Related

- **SPEC-V3R3-HYBRID-001** (PR #770 merged): provider abstraction 패턴 — 본 SPEC의 retirement standardization 패턴과 유사한 closed-set + audit test 구조.
- **SPEC-V3R2-WF-005** (PR #768 merged): 16-language neutrality — manager-cycle.md 신규 파일이 동일 neutrality 규칙 적용.
- `feedback_large_spec_wave_split.md` (auto-memory lesson #9): layer 5 `stream_idle_partial` 우회 정책 — 본 SPEC layer 1-4 fix와 complementary.
- mo.ai.kr 2026-05-04 21:14:54 incident logs: 본 SPEC의 1차 evidence base.

---

## 10. BC Migration

본 SPEC은 BREAKING CHANGE 없음 (`breaking: false`, `bc_id: []`).

이유: P0 fix는 backward-compatible.
- `manager-cycle.md` 신규 추가: 기존 사용자 프로젝트에 영향 없음 (새 파일 추가만).
- `manager-tdd.md` retired stub frontmatter standardization: 기존 retired stub은 retirement message 반환 후 종료; standardized version도 동일 동작 + 추가 metadata로 SubagentStart guard가 더 빠르게 reject. 기존 `manager-tdd` 호출자는 retirement message + migration hint를 동일하게 받음.
- SubagentStart guard: 기존 hook chain에 추가; 비-retired agent invocation은 변경 없음.
- `worktreePath` validation: 기존 valid worktreePath 반환 케이스 변경 없음; broken empty-object 케이스만 explicit error로 surface.

마이그레이션 절차: 사용자는 `moai update` 실행만 하면 자동 sync. 별도 user action 불필요.

CHANGELOG 항목 (proposed):

```markdown
## [Unreleased]

### Bug Fixes (P0)

- **SPEC-V3R3-RETIRED-AGENT-001**: Fixed 5-layer defect chain in agent runtime when invoking retired agents (e.g., `manager-tdd`).
  - Added `internal/template/templates/.claude/agents/moai/manager-cycle.md` (unified DDD/TDD implementation agent, replaces retired manager-ddd + manager-tdd per SPEC-V3R2-ORC-001).
  - Standardized retired-agent frontmatter: `retired: true`, `retired_replacement`, `retired_param_hint`, `tools: []`, `skills: []`.
  - Added SubagentStart hook guard that blocks retired-agent spawns at the runtime layer (response time ≤500ms; previously 11.4s).
  - Added `worktreePath` empty-object validation in Agent() wrapper to prevent `Path "/{}/{}" does not exist` errors.
  - Added `agent_frontmatter_audit_test.go` to enforce retirement standardization at CI time.
  - User action: run `moai update` to sync the new template files.
```

---

## 11. Traceability (추적성)

- REQ 총 16개: Ubiquitous 6, Event-Driven 4, State-Driven 3, Optional 1, Unwanted 2.
- AC 총 18개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지) — 매트릭스는 plan.md §1.4 참조.
- 5-layer defect chain mapping:
  - Layer 1 (retired stub frontmatter invalid) → REQ-RA-002, REQ-RA-007, REQ-RA-011 → AC-RA-02, AC-RA-07, AC-RA-11, AC-RA-17
  - Layer 2 (worktree allocation timing) → REQ-RA-005, REQ-RA-010 → AC-RA-05, AC-RA-10, AC-RA-18
  - Layer 3 (auto-fallback propagation) → REQ-RA-005, REQ-RA-010 → AC-RA-05, AC-RA-18
  - Layer 4 (path interpolation `{}/{}`)  → REQ-RA-006 → AC-RA-06, AC-RA-18
  - Layer 5 (stream idle, side pattern) → out of scope (§1.3)
- BC 영향: 0건 (`breaking: false`, `bc_id: []`).
- 의존성: SPEC-V3R2-ORC-001 (completed) 1건; SPEC-V3R3-HYBRID-001 (PR #770 merged) related.
- 구현 경로 예상:
  - 신규 1 agent definition file (`manager-cycle.md`)
  - 신규 1 hook handler (`agent_start.go`) + 테스트 + factory dispatch 확장
  - 1 retired stub standardization (`manager-tdd.md`)
  - 6 documentation reference substitutions (CLAUDE.md §4, §5, agent-authoring.md, agent-hooks.md, spec-workflow.md, manager-ddd.md)
  - 1 launcher.go validation guard + path interpolation refactor
  - 2 audit/regression test files
  - lessons.md #11 entry

---

End of SPEC.
