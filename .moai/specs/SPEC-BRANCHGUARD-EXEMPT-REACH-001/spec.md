---
id: SPEC-BRANCHGUARD-EXEMPT-REACH-001
title: "BranchGuard 면제 축의 서브에이전트 도달성 판정 — agent_type 배선 철자와 deny 억제의 실행 기반 확립"
version: "0.1.0"
status: draft
created: 2026-09-21
updated: 2026-09-21
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/hook"
lifecycle: spec-anchored
tags: "branch-guard, hook, exemption, reachability, doctrine, t1064"
---

# SPEC-BRANCHGUARD-EXEMPT-REACH-001

## HISTORY

- 2026-09-21 v0.1.0 — plan 단계 최초 저작 (card t1064). 리드가 「PLAUSIBLE, 확정 아님」 등급으로 넘긴 독트린 주장을 **실행으로** 판정하도록 설계.

## §A 배경 — 무엇이 주장돼 있고 무엇이 측정되지 않았는가

`.claude/rules/moai/workflow/main-checkout-branch-guard.md` § Mechanical Enforcement 는 BranchGuard 면제 두 축이 **도구로 spawn 된 서브에이전트에서 도달 불가**라고 단언한다. 이 트리에서 실측한 원문:

> **Exemptions are unreachable from a tool-spawned subagent.** Both axes work, but neither value
> reaches one: `AgentType` is populated only for a main-thread `claude --agent manager-git` launch,
> and `MOAI_BRANCH_GUARD_EXEMPT=1` is read from the hook process's own environment, which is spawned
> before the guarded command runs.

같은 주장이 코드 주석(`internal/hook/branch_guard.go:25-42`)과 deny reason 문자열(`internal/hook/branch_guard.go:604-606`)에도 박혀 있다. 그런데 **같은 저장소의 다른 문서가 정면으로 반대**를 말한다 — `.claude/rules/moai/core/hooks-system.md:114`:

> All hook events include `agent_id` and `agent_type` fields when triggered from a subagent context (v2.1.69+).

이 모순은 이미 코드베이스에 **CONTESTED 로 기록돼 있다**(`internal/hook/branch_guard_flagclass_test.go:228-246`). 선행 카드가 캡처를 시도했으나 중첩 `claude -p` 프로브가 워크트리 세션 가드에 거부돼 **미측정으로 남았고**, 훅 자체의 trace 기록(`internal/hook/trace/entry.go` `TraceEntry`)에는 에이전트 정체성 필드가 없어 기존 로그로도 가릴 수 없다. 둘 다 이 트리에서 재확인했다(§E, `progress.md` §E.1).

따라서 이 SPEC 이 답하는 것은 **「독트린이 맞는가」가 아니라 「그 주장을 실행으로 판정할 수 있는가, 그리고 그 결과가 무엇인가」**이다.

### §A.1 두 개의 `AgentType` — 배선 철자 축

같은 패키지에 이름이 같고 **JSON 태그 철자가 다른** 필드가 둘 있다(이 트리 실측):

| 선언 | 태그 | 용도 |
|---|---|---|
| `internal/hook/types.go:230` | `agent_type` (snake) | `HookInput` 공용 필드 — `isExemptAgent` 가 읽는 바로 그 필드 |
| `internal/hook/subagent_stop.go:162` | `agentType` (camel) | SubagentStop 전용 구조체 |

`hooks-system.md` 자신도 두 철자를 모두 쓴다 — SubagentStop 행은 `agentType`, 총칙 문장은 `agent_type`. **런타임이 서브에이전트 컨텍스트의 PreToolUse 에서 실제로 어느 철자를 보내는지가 「채워질 수 있는가」를 결정한다.** 이 축을 고정하지 않은 판정은 카드의 질문에 답할 수 없다.

> 배선 철자가 camel 이라면 `HookInput` 의 snake 태그는 값을 **말없이 버린다** — 도달 가능한 값이 도달 불가로 관측되는 제3의 상태이며, 「도달 불가」와 출력이 구별되지 않는다. 이것이 이 축을 별도 요구사항으로 세운 이유다.

### §A.2 주장 보유 지점 (실측 8파일)

`tool-spawned subagent` 리터럴 기준, 이 트리에서 8파일이 같은 주장을 싣고 있다(양성 대조: `branchGuardExemptEnv` 토큰이 8파일 적중 — 계측기 정상 발화):

```
.claude/rules/moai/workflow/main-checkout-branch-guard.md
.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md
internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md
internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md
internal/hook/branch_guard.go
internal/hook/branch_guard_test.go
internal/hook/branch_guard_flagclass_test.go
internal/hook/branch_guard_quoted_test.go
```

**이 리터럴 개수는 하한이다** — 굵게 표기·줄바꿈·다른 표현으로 같은 주장을 싣는 지점은 이 패턴에 안 잡힌다. 처방 단계에서 패턴을 흔들어 재열거해야 한다(AC-BGX-009).

## §B 요구사항 (GEARS)

- **REQ-BGX-001** (배선 철자) — 서브에이전트 컨텍스트에서 발화한 PreToolUse 훅 페이로드가 관측될 때, 시스템은 에이전트 정체성이 `agent_type`·`agentType`·**어느 쪽도 아님** 중 무엇으로 도착하는지를 관측된 원문과 함께 기록해야 한다(shall).
- **REQ-BGX-002** (도달성 판정) — 도구로 spawn 된 서브에이전트가 `isExemptAgent` 를 true 로 만들고 `internal/hook/branch_guard.go:584` 이후의 deny 를 억제할 수 있는지를, 코드 판독이 아니라 **경로 실행**으로 판정해야 한다(shall).
- **REQ-BGX-003** (계측기 양성 대조) — 캡처 계측기가 사용될 때, 시스템은 그 계측기가 **실재하는 필드를 실제로 기록한다**는 양성 대조를 같은 실행에서 함께 남겨야 한다(shall). 대조 없는 필드 부재는 「도달 불가」가 아니라 **「미측정」**으로 기록해야 하며, 부재만으로 결론을 내서는 안 된다(shall not).
- **REQ-BGX-004** (부정 분기의 기제 검증) — 도달 불가로 판정될 때, 시스템은 그 도달 불가의 **기제가 독트린이 서술하는 기제와 동일한지**를 별도로 검증해야 한다(shall). 결론이 같고 서술된 이유가 다르면 그것은 독트린의 결함이며, 통과로 기록해서는 안 된다(shall not).
- **REQ-BGX-005** (처방 분기점) — 도달 가능으로 판정될 때, 시스템은 어떤 수정도 하기 전에 **CODE(면제 범위 축소) / DOC(독트린 정정)** 중 무엇을 처방으로 할지 명시적 결정을 기록해야 한다(shall). 어느 한쪽을 미리 선택해서는 안 된다(shall not).
- **REQ-BGX-006** (변이 두 방향) — 처방이 면제 범위를 축소하는 코드 변경일 때, 시스템은 (a) **무변이 성공** — 정당한 `manager-git` 경로가 여전히 통과, (b) **변이 검출** — 그 정체성을 사칭한 서브에이전트가 실제로 거부됨, 두 방향을 모두 보여야 한다(shall). 한 방향만으로 「가드가 동작한다」를 주장해서는 안 된다(shall not).
- **REQ-BGX-007** (주장 보유 지점 동반 이동) — 판정 결과가 §A.2 지점들의 문안을 바꿀 때, 시스템은 로컬 `.claude/` 사본과 `internal/template/templates/` 미러를 **함께** 갱신해야 한다(shall).
- **REQ-BGX-008** (순서 귀속) — 분기점 결정 기록은 어떤 수정 커밋보다 **앞선 자기 커밋**에 착지해야 한다(shall). 같은 커밋에 묶인 쌍은 순서를 증언하지 못한다.
- **REQ-BGX-009** (등급 보존) — 판정서가 선행 카드의 결론이나 이 SPEC 의 PLAUSIBLE 항목을 인용할 때, 시스템은 원문이 붙인 등급을 **문장 안에** 함께 옮겨야 한다(shall). 등급을 각주·별도 열로 분리해서는 안 된다(shall not).

## §C 완료의 형태

이 카드는 **판정서를 산출물로 하는 조사 카드**다. 코드 변경은 조건부이며, 도달 불가로 판정되고 기제까지 일치하면 **코드 변경 0 으로 닫는 것이 정상 종료**다. 그 경우에도 CONTESTED 표식(`internal/hook/branch_guard_flagclass_test.go:228-246`)은 「미측정」에서 「측정됨」으로 갱신된다 — 측정했는데 표식을 그대로 두면 다음 사람이 같은 조사를 처음부터 반복한다.

세 종결 형태가 모두 정상이다:

| 판정 | 처방 | 남는 것 |
|---|---|---|
| 도달 가능 | CODE 또는 DOC (분기점 결정, AC-BGX-006) | 판정서 + 결정 + 처방 + 변이 두 방향(CODE 인 경우) |
| 도달 불가 · 기제 일치 | 없음 | 판정서 + CONTESTED 표식 갱신 |
| 도달 불가 · 기제 불일치 | DOC (서술된 이유 정정) | 판정서 + 독트린 8지점 정정 |
| 측정 불가 | 없음 | 판정서에 **무엇이 막았는지**와 「미측정」 등급 |

**「측정 불가」는 「도달 불가」가 아니다.** 선행 카드가 남긴 것이 바로 이 구분이며, 이 SPEC 이 그것을 지우지 않고 갱신하는 것이 성공이다.

## §D 배제 범위

### Out of Scope — 인접 훅 결함
- `internal/hook/pre_tool.go:1198` `frozenZonePrefixes` 의 상대경로 가정 대 절대경로 `file_path` 도착 문제. **부수 발견(INCIDENTAL)으로만 기록하며 이 카드의 인수조건을 쓰지 않는다.** 이 카드가 재지 않았으므로 그 자체도 PLAUSIBLE 등급이다.

### Out of Scope — t1056 과의 분리
- **이 카드와 t1056 은 반대 방향이다.** t1056 은 **deny 가 너무 넓다**(과다 매칭)를, 이 카드는 **면제가 너무 넓다**(도달 가능하면 우회 가능)를 다룬다. 한쪽 수정이 다른 쪽을 해결하지 않으며, 묶으면 어느 축의 회귀인지 귀속이 불가능해진다. 나중 독자가 병합하지 않도록 이 문단을 남긴다.
- 두 카드가 공유하는 `internal/hook/branch_guard.go` 는 **통합 창에서 직렬화**할 대상이지 합칠 근거가 아니다.

### Out of Scope — 환경변수 축의 신규 측정
- `MOAI_BRANCH_GUARD_EXEMPT` 축의 도달 불가는 선행 카드가 **uncontested** 로 기록했고 이 카드는 그것을 재측정하지 않는다. 다만 독트린 문장이 **두 축을 한 불릿에 담고 있으므로**, 그 문장을 고칠 때 환경변수 절반을 「이번에 검증됨」으로 조용히 승격시켜서는 안 된다 — 판정서에 **carried-forward-unverified** 로 명시한다(AC-BGX-008).

### Out of Scope — 런타임 자체 수정
- Claude Code 런타임이 어느 철자를 보내는지는 **관측 대상이지 변경 대상이 아니다.** 런타임 동작을 바꾸는 처방은 이 카드가 내지 않는다.

### Out of Scope — 전체 스위트 실행
- `go test ./...` 는 실행하지 않는다. 검증 범위는 `./internal/hook/...` 이며, 전 패키지 판정은 CI 몫이다.

## §E 측정 귀속 (plan 단계)

이 SPEC 본문의 모든 좌표·개수·인용 원문은 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064`, HEAD `69d61d371` 에서 **이번 실행에** 측정했다. 명령과 출력 전문은 `progress.md` §E.1.
