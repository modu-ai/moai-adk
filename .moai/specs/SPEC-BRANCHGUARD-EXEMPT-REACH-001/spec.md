---
id: SPEC-BRANCHGUARD-EXEMPT-REACH-001
title: "BranchGuard 면제 축의 서브에이전트 도달성 판정 — agent_type 배선 철자와 deny 억제의 실행 기반 확립"
version: "0.5.0"
status: draft
created: 2026-09-21
updated: 2026-09-21
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/hook"
lifecycle: spec-anchored
tier: M
tags: "branch-guard, hook, exemption, reachability, doctrine, t1064"
---

# SPEC-BRANCHGUARD-EXEMPT-REACH-001

## HISTORY

- 2026-09-21 v0.1.0 — plan 단계 최초 저작 (card t1064). 리드가 「PLAUSIBLE, 확정 아님」 등급으로 넘긴 독트린 주장을 **실행으로** 판정하도록 설계.
- 2026-09-21 v0.2.0 — plan-audit iter-1 FAIL(0.8125) 수정. `tier: M` 선언(§A.5), 프로브 집합에 `agent_id` 축 추가(§A.3), 동반이동 집합 8 → **최소 10**(§A.2), 계측기 실행 경로의 실행 가능성 근거 기록(§A.4), REQ-BGX-002 주어 삽입, REQ-BGX-010 신설(판별자 부재 시 CODE 분기 선택 불가).
- 2026-09-21 v0.3.0 — plan-audit D8 수정. 결정적 판정(REQ-BGX-001)의 측정 방법을 grep 에서 **키 집합 동등 비교**로 고정. 근거: camel 축 과다매칭원이 셋이고 둘이 대문자로 시작해(실측) 「앵커를 달아라」식 수정이 말없이 뚫린다. D8 과 D2 가 이 한 조항으로 함께 닫힌다.
- 2026-09-21 v0.4.0 — plan-audit iter-2(PASS-WITH-DEBT 0.9375) 잔여 결함 D9·D10 수정. §A.2 의 개수를 **R4 형태**(명령을 먼저, 값은 날짜·ref 를 단 참조)로 고쳐 쓰고, 수치를 넓은 뿌리 **20** / 이 SPEC 아티팩트 **3** 으로 정정하며 **두 방향의 움직임**(판정서 미추적 +1, `plan.md` 리터럴 삭제 −1)을 모두 서술. AC-BGX-009 에 「plan 시점 숫자를 run 시점 측정과 한 측정처럼 대조하지 않는다」 추가. 감사창 중 HEAD 이동은 `progress.md` §E.1 iter-4 에 절차 결함으로 귀속 기록.
- 2026-09-21 v0.5.0 — plan-audit D11 수정. AC-BGX-009 (1)을 개수를 움직이는 **네 좌표**(뿌리·패턴·시점 SHA·추적 범위) 보고 의무로 재구성. 추적 범위를 뿌리와 대칭으로 올리고(두 수 모두 보고), 어떤 plan 시점 상수도 합격 비교 대상이 되지 않도록 배제. 근거: 같은 뿌리·패턴·SHA 에서 `git grep -l` 19 대 `grep -rln` 20(실측, 델타는 미추적 `verdict.md`).

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

### §A.2 동반이동 집합 — 최소 10파일, 그리고 변주 축은 패턴이 아니라 탐색 뿌리다

**도달-불가 주장을 싣는 8파일** (`tool-spawned subagent` 리터럴, 뿌리 `.claude`+`internal`):

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

**반대 주장을 싣는 2파일 — 이 리터럴에는 0회 적중한다:**

```
.claude/rules/moai/core/hooks-system.md                                    (:114)
internal/template/templates/.claude/rules/moai/core/hooks-system.md        (:114)
```

두 파일은 `tool-spawned subagent` 를 **한 번도 싣지 않는다**(실측 0/0, 양성 대조로 문제의 문장이 두 파일 모두에 1회씩 실재함을 확인). 그러나 **판정이 어느 방향으로 나든 모순하는 두 문장 중 하나는 반드시 바뀐다** — 도달 가능이면 도달-불가 주장 8지점이, 도달 불가이면 `hooks-system.md:114` 의 총칙 문장이 (적어도 PreToolUse 범위에 대해) 정정돼야 한다. 따라서 두 파일은 **적중 여부와 무관하게 구성상 집합의 일부**이며, 동반이동 집합은 **최소 10**이다.

#### 변주 축 — 패턴이 아니라 탐색 뿌리다

plan iter-1 은 「패턴을 흔들어 재열거하라」고 지시했는데 **그 축은 틀렸다.** 마크업·대소문자·하이픈·공백을 흔들어도 숫자는 **0만큼 움직인다**. 숫자를 움직이는 것은 **탐색 뿌리**다.

**[HARD] 이 개수는 값이 아니라 명령으로 산다.** 아래 셋이 판정식이며, 인용할 때는 **명령을 먼저 적고 값은 괄호 안에 날짜·측정 ref 와 함께 참조로만** 단다. 값을 앞에 두면 그 값이 사실로 굳고 뿌리도 시점도 따라오지 않는다.

```bash
# (a) 좁은 뿌리 — 룰과 코드만
grep -rln "tool-spawned subagent" .claude internal | wc -l
# (b) 넓은 뿌리 — 저장소 전체(.git 제외). 미추적 파일을 포함한다
grep -rln --exclude-dir=.git "tool-spawned subagent" . | wc -l
# (c) 커밋 트리만 — 같은 질문에서 미추적을 뺀 값
git grep -l "tool-spawned subagent" <ref> | wc -l
```

패턴은 리터럴 `tool-spawned subagent` 이며 다른 토큰의 부분문자열이 아니다 — 그래서 §A.2 의 `agent_type` 계열과 달리 경계 앵커가 필요 없다. **다만 AC-BGX-009 의 접두 열거 의무는 면제되지 않는다**: 「부분문자열이 아니다」는 이 시점의 관측이고, 그 판정은 세는 시점에 다시 내린다.

참조값 — **2026-09-21, ref `8852cb921` 에서 측정**(읽는 시점의 사실이 아니라 그때의 관측): (a) **8**, (b) **20**, (c) **19**.

##### 이 수가 움직인 경위 — 움직임은 둘이었고 iter-2 는 하나만 적었다

iter-2 는 (b)를 **21** 로 적고 차이를 「판정서가 감사 이후에 착지했다」(+1) 하나로 설명했다. **그 설명은 절반이다.** 같은 구간에 반대 방향의 움직임이 하나 더 있었다:

```
$ git show 3d636c991 -- .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/plan.md | grep -n 'tool-spawned subagent'
135:-1. `tool-spawned subagent` 리터럴로 재열거(plan 시점 하한 8파일).
```

앞머리 `-` — **iter-2 커밋 `3d636c991` 자신이 `plan.md` M6 에서 그 리터럴을 지웠다**(−1). 재열거 지시를 「뿌리를 넓혀라」로 고쳐 쓰면서 리터럴이 본문에서 빠졌다. 즉 **이 카드가 세는 대상을 이 카드가 바꿨다.**

| ref | (c) 커밋 트리 | (b) 넓은 뿌리(미추적 `verdict.md` 포함) | 이 SPEC 아티팩트 |
|---|---|---|---|
| `039af6915` | 20 | 21 | 4 |
| `3d636c991` | 19 | 20 | 3 |
| `8852cb921` | 19 | 20 | 3 |

**세 관측이 각각 내부적으로 옳았고, 어긋나 보인 것은 축이 둘이었기 때문이다** — 판정서의 유무(+1)와 `plan.md` 리터럴의 유무(−1). iter-2 가 첫째만 적어 둘째를 가렸다.

[HARD] **21 은 「어느 커밋에도 대응하지 않는 편집 중간 상태」가 아니다.** `.moai/reports/*/verdict.md` 는 `.gitignore:259` 의 네거션이 걸려 **무시 대상이 아니라 단지 미추적**이고, 21 은 `039af6915` 의 **워킹 트리**(커밋 트리 20 + 미추적 1)를 정확히 센 값이다. 커밋 트리에 21 이 없는 이유는 편집 중간이어서가 아니라 **그 +1 이 계속 미추적 쪽에 있기 때문**이다. 이 구별이 실무에서 갈리는 지점: 「중간 상태였다」로 적으면 다음 사람이 **커밋 경계만 맞추면 재현된다**고 믿지만, 실제 재현 조건은 **(b)와 (c) 중 어느 것을 셌는지**를 적는 것이다.

[HARD] 넓힌 뿌리가 데려오는 것 중 **정정 대상이 아닌 것**: 닫힌 SPEC 아티팩트(`SPEC-RC-TESTBED-001`, `SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001`), `CHANGELOG.md`, 그리고 판정서 — 전부 **그 시점의 관측 기록**이며 나중 사실에 맞춰 고치지 않는다.

#### [HARD] 측정 방법 — 결정적 판정은 grep 이 아니라 키 집합 동등 비교다

정체성 필드의 도착 여부(§B REQ-BGX-001)는 **문자열 검색으로 판정하지 않는다.** 캡처한 페이로드의 **최상위 키 집합**과 프로브 이름을 **정확 일치**로 대조한다.

이유는 「앵커를 달면 된다」가 **성립하지 않기 때문**이다. 두 축의 과다매칭원 수가 다르고, camel 축은 **대문자로 시작하는 것이 둘**이라 소문자 기준으로 쓴 수정이 통과시킨다 — 이 트리 실측:

| 축 | 앵커 없음 | 과다매칭원 | 앵커 적용 |
|---|---|---|---|
| snake `agent_type` | 48 | `subagent_type` 20 | **28** |
| camel `agentType` | 29 | `subagentType` 2 · `SubagentType` 3 · `TestAgentSpawnFixtureCarriesSubagentType` 2 | **22** |

camel 축을 「`subagentType` 을 빼라」로 고치면 **대문자 시작 둘(5건)을 놓친다.** 앵커는 레시피에 붙이는 패치이고 다음 철자 변종에 말없이 뚫리는 반면, **키 `"subagent_type"` 은 키 `"agent_type"` 과 같지 않다** — 키 비교에서는 이 계열 오류가 구조적으로 불가능하다.

**오독의 방향이 나쁘다.** 과다매칭은 **더 큰 수**를 만들고, 큰 수는 「더 많이 찾았다」로 읽히며 아무것도 신고하지 않는다. 이 축에서 그 오독의 종착지는 **「도달 가능」이라는 결정적 오답으로 카드가 닫히는 것**이다.

grep 을 피할 수 없는 곳(§A.2 의 산문 리터럴 열거)에서는 **접두 열거**로 대체한다 — `[A-Za-z]*<토큰>` 으로 과다매칭원을 **먼저 드러내고**, 열거 합과 앵커 적용값이 닫히는지 보인 뒤에 앵커를 정한다(AC-BGX-009). 앵커를 짐작으로 먼저 다는 것은 금지다.

### §A.3 `agent_id` — 넷째 프로브이자 유일한 판별자 후보

다투는 문장(`hooks-system.md:114`)은 `agent_id` 와 `agent_type` 을 **나란히** 이름을 댄다. 그리고 `agent_id` 는 이미 선언된 `HookInput` 필드다 — `internal/hook/types.go:239`: `AgentID string \`json:"agent_id,omitempty"\``.

이 필드가 프로브 집합에서 빠지면 **정체성이 `agent_id` 로만 도착하는 세계를 구조적으로 못 본다.** 그것은 §A.1 의 camel 철자 누락과 **정확히 같은 실패 형태**이며, 필드 이름 하나가 다를 뿐이다. 따라서 프로브 집합은 네 철자다: `agent_type` · `agentType` · `agent_id` · `agentId`.

`agent_id` 는 두 번째 역할도 갖는다 — **사칭 입력의 유일한 판별자 후보**다. 「서브에이전트이면서 `manager-git` 을 자칭하는 입력」은 서브에이전트임을 말해 주는 필드가 있어야만 구성된다. 그 후보가 이것뿐이므로, **`agent_id` 가 도착하지 않으면 CODE 분기(면제 범위 축소)는 unit 층에서 구성 불가능**하다(REQ-BGX-010).

### §A.4 실행 가능성 근거 — 계측기는 moai 재빌드 없이 들어간다

PreToolUse 훅은 Go 바이너리로 직행하지 않고 **이 워크트리 자신의 셸 래퍼**를 통과한다. 그 래퍼는 stdin 원문을 셸 변수에 통째로 담아 둔 뒤 moai 에 넘긴다 — `.claude/hooks/moai/handle-pre-tool.sh:37`:

```
payload=$(head -c 1048576)
```

따라서 원문 덤프·키 집합 기록 계측기(M1 후보 C1/C2)는 **moai 재빌드도 설치본 치환도 없이** 이 한 지점에서 성립하며, 「설치본이 트리를 판정한다」는 위험(VCI §2.2)은 **이 경로에서는 구조적으로 발생하지 않는다.**

[HARD] 단, 계측기를 `TraceEntry` 확장(후보 C3)으로 택하면 **그 위험은 되살아난다** — 그쪽은 Go 코드이므로 트리 빌드가 실제로 실행되는지 보여야 한다. 위험의 부재는 **경로에 붙는 성질이지 카드에 붙는 성질이 아니다.**

반대급부: 그 래퍼는 `moai update` 관리 대상 뿌리(`.claude/hooks/moai/`)에 있고 템플릿에서 배포된다. 계측기를 **되돌리지 않으면** 수정 상태가 남는다(AC-BGX-010).

### §A.5 Tier 판정 — M, 그리고 그 근거

**Tier M 으로 선언한다.** 근거는 셋이고, 반대 방향을 가리키는 지표 하나도 함께 적는다.

| 기준 | 관측 | 가리키는 Tier |
|---|---|---|
| 영향 파일 수 | 동반이동 집합 최소 10 (5-15 밴드) | **M** |
| REQ / AC 예산 | REQ 10 · AC 12 — Tier S 천장(8/8) **초과**, M 천장(16/16) 내 | **M** |
| 아티팩트 집합 | spec + plan + acceptance (+progress) — M 의 3파일 집합과 일치 | **M** |
| 코드 LOC | CODE 분기라도 `isExemptAgent` 주변 + 테스트로 300 LOC 미만 | S |

LOC 만 S 를 가리킨다. 그러나 **천장 초과는 tier 를 낮출 근거가 아니라 올릴 신호**이며(`spec-workflow.md` § SPEC Complexity Tier), 파일 수와 아티팩트 집합이 독립적으로 M 을 가리킨다. Tier L 은 성립하지 않는다 — 1000 LOC 도 15파일 초과도 아니고, `design.md`·`research.md` 를 요구할 설계 결정이 이 카드에 없다(이 카드는 **측정 카드**이지 설계 카드가 아니다).

[HARD] **이 선언은 임계를 낮추려고 한 것이 아니다.** 선언의 부수 효과로 감사 임계가 0.85 에서 0.80 으로 내려가는 것은 사실이며 여기 명시해 둔다 — 그러나 iter-1 의 차단 결함(D2/D3/D4)은 **임계와 무관하게 각자의 근거로** 수정했다. 임계를 근거로 tier 를 고른 것이라면 그것은 측정이 아니라 거래다.

## §B 요구사항 (GEARS)

- **REQ-BGX-001** (배선 철자) — 서브에이전트 컨텍스트에서 발화한 PreToolUse 훅 페이로드가 관측될 때, 시스템은 에이전트 정체성이 네 철자(`agent_type`·`agentType`·`agent_id`·`agentId`) 중 무엇으로 도착하는지, 또는 **넷 다 아님**인지를 관측된 원문과 함께 기록해야 한다(shall). 이 판정은 캡처된 페이로드의 **최상위 키 집합과의 정확 일치 비교**로 내려야 하며(shall), 문자열 검색·부분문자열 포함·정규식 매칭을 판정 근거로 삼아서는 안 된다(shall not).
- **REQ-BGX-002** (도달성 판정) — 도구로 spawn 된 서브에이전트가 `isExemptAgent` 를 true 로 만들고 `internal/hook/branch_guard.go:584` 이후의 deny 를 억제할 수 있는지를, 시스템은 코드 판독이 아니라 **경로 실행**으로 판정해야 한다(shall).
- **REQ-BGX-003** (계측기 양성 대조와 키 집합) — 캡처 계측기가 사용될 때, 시스템은 페이로드의 **최상위 키 집합 전량**을 **키 단위로 분리된 형태**로 기록해야 하며(shall), 그 기록 자체가 계측기의 양성 대조이자 REQ-BGX-001 비교의 입력이다. 원문 JSON 을 한 덩어리로만 남겨서는 안 된다(shall not) — 그러면 비교가 문자열 검색으로 퇴화한다. 네 철자의 부재만 기록하고 키 집합을 남기지 않아서는 안 된다(shall not) — 그러면 「런타임이 정체성을 안 보낸다」와 「프로브하지 않은 제3의 키로 보낸다」가 구별되지 않는다. 키 집합 없는 필드 부재는 「도달 불가」가 아니라 **「미측정」**으로 기록해야 한다(shall).
- **REQ-BGX-004** (부정 분기의 기제 검증) — 도달 불가로 판정될 때, 시스템은 그 도달 불가의 **기제가 독트린이 서술하는 기제와 동일한지**를 별도로 검증해야 한다(shall). 결론이 같고 서술된 이유가 다르면 그것은 독트린의 결함이며, 통과로 기록해서는 안 된다(shall not).
- **REQ-BGX-005** (처방 분기점) — 도달 가능으로 판정될 때, 시스템은 어떤 수정도 하기 전에 **CODE(면제 범위 축소) / DOC(독트린 정정)** 중 무엇을 처방으로 할지 명시적 결정을 기록해야 한다(shall). 어느 한쪽을 미리 선택해서는 안 된다(shall not).
- **REQ-BGX-006** (변이 두 방향) — 처방이 면제 범위를 축소하는 코드 변경일 때, 시스템은 (a) **무변이 성공** — 정당한 `manager-git` 경로가 여전히 통과, (b) **변이 검출** — 그 정체성을 사칭한 서브에이전트가 실제로 거부됨, 두 방향을 모두 보여야 한다(shall). 한 방향만으로 「가드가 동작한다」를 주장해서는 안 된다(shall not).
- **REQ-BGX-007** (주장 보유 지점 동반 이동) — 판정 결과가 §A.2 집합의 문안을 바꿀 때, 시스템은 로컬 `.claude/` 사본과 `internal/template/templates/` 미러를 **함께** 갱신해야 하며(shall), 그 집합에는 반대 주장을 싣는 `hooks-system.md` 쌍과 deny reason 문자열을 고정하는 Go 테스트가 포함돼야 한다(shall). 리터럴 적중 여부를 집합 판별식으로 삼아서는 안 된다(shall not).
- **REQ-BGX-008** (순서 귀속) — 분기점 결정 기록은 어떤 수정 커밋보다 **앞선 자기 커밋**에 착지해야 한다(shall). 같은 커밋에 묶인 쌍은 순서를 증언하지 못한다.
- **REQ-BGX-009** (등급 보존) — 판정서가 선행 카드의 결론이나 이 SPEC 의 PLAUSIBLE 항목을 인용할 때, 시스템은 원문이 붙인 등급을 **문장 안에** 함께 옮겨야 한다(shall). 등급을 각주·별도 열로 분리해서는 안 된다(shall not).
- **REQ-BGX-010** (판별자 부재 시 CODE 분기 폐쇄) — 서브에이전트를 식별할 판별 필드가 페이로드에 도착하지 않는 것으로 측정될 때, 시스템은 **CODE 분기(면제 범위 축소)를 선택 불가로 선언**하고 그 사실과 근거를 판정서에 기록해야 한다(shall). 판별자 없이 구성한 「사칭 형태」 입력으로 변이 검출을 주장해서는 안 된다(shall not) — 그 입력은 이미 존재하는 테스트가 고정하는 형태와 같아져 공허 통과가 된다.

## §C 완료의 형태

이 카드는 **판정서를 산출물로 하는 조사 카드**다. 코드 변경은 조건부이며, 도달 불가로 판정되고 기제까지 일치하면 **코드 변경 0 으로 닫는 것이 정상 종료**다. 그 경우에도 CONTESTED 표식(`internal/hook/branch_guard_flagclass_test.go:228-246`)은 「미측정」에서 「측정됨」으로 갱신된다 — 측정했는데 표식을 그대로 두면 다음 사람이 같은 조사를 처음부터 반복한다.

세 종결 형태가 모두 정상이다:

| 판정 | 처방 | 남는 것 |
|---|---|---|
| 도달 가능 · 판별자 도착 | CODE 또는 DOC (분기점 결정, AC-BGX-006) | 판정서 + 결정 + 처방 + 변이 두 방향(CODE 인 경우) |
| 도달 가능 · 판별자 부재 | **DOC 만** — CODE 분기는 선택 불가(REQ-BGX-010) | 판정서 + 선택 불가 선언과 그 근거 + DOC 처방 |
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
