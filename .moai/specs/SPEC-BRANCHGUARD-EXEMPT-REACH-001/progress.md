# 진행 기록 — SPEC-BRANCHGUARD-EXEMPT-REACH-001

card: t1064 · 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064` · 브랜치: `WT-branchguard-exempt`

## §E.1 Plan-phase Audit-Ready Signal

### iter-2 개정 (HEAD `039af6915` 기준 측정)

plan-audit iter-1 이 **FAIL 0.8125** (Tier L 임계 0.85) 로 차단 결함 4건을 냈고, 이 절 아래의 iter-1 증거 위에 iter-2 측정을 덧붙인다. **iter-1 측정을 고쳐 쓰지 않는다** — 관측 기록은 나중 사실에 맞춰 수정하지 않는다.

**(8) 탐색 뿌리 변주 — 변주 축이 패턴이 아님을 확인**

```
$ grep -rln "tool-spawned subagent" .claude internal | wc -l
8
$ grep -rln --exclude-dir=.git "tool-spawned subagent" . | wc -l
21
```

감사는 같은 측정에서 **20** 을 얻었다. 지금 **21** 인 차이 1건은 감사 이후 착지한 판정서 자신(`.moai/reports/t1064/verdict.md`)이다 — 확인:

```
$ grep -rln --exclude-dir=.git "tool-spawned subagent" . | grep -i "reports\|verdict"
.moai/reports/t1064/verdict.md
```

**감사의 20 도 지금의 21 도 틀리지 않았다** — 같은 뿌리를 서로 다른 시점에 잰 것이다. 숫자만 옮기면 모순으로 읽히므로 §A.2 는 뿌리와 시점을 함께 적는다.

> **[iter-4 정정]** 위 측정값 21 자체는 옳다(그 시점 워킹 트리). 틀린 것은 **그 다음 문장의 해석**이다 — ~~차이를 판정서 착지(+1) 하나로 설명했다~~. 같은 구간에 반대 방향 움직임이 하나 더 있었다: **이 카드의 iter-2 커밋 `3d636c991` 자신이 `plan.md` M6 에서 세는 대상 리터럴을 지웠다(−1).** 그리고 +1 의 성격도 「나중에 착지」가 아니라 **미추적**이다(§E.1 iter-4 (17)). iter-2 의 원문은 관측 기록이므로 고치지 않고 이 표식만 단다.

**(9) 반대 주장 보유자 쌍 — 리터럴 0회, 양성 대조 발화**

```
$ grep -c "tool-spawned subagent" .claude/rules/moai/core/hooks-system.md internal/template/templates/.claude/rules/moai/core/hooks-system.md
internal/template/templates/.claude/rules/moai/core/hooks-system.md:0
.claude/rules/moai/core/hooks-system.md:0

$ grep -c "when triggered from a subagent context" .claude/rules/moai/core/hooks-system.md internal/template/templates/.claude/rules/moai/core/hooks-system.md
.claude/rules/moai/core/hooks-system.md:1
internal/template/templates/.claude/rules/moai/core/hooks-system.md:1
```

양성 대조(1/1)가 발화했으므로 0/0 은 **부재**다. 두 파일은 어떤 뿌리에서도 이 리터럴로는 안 잡히며, **적중이 아니라 구성으로** 동반이동 집합에 든다(집합 최소 10).

**(10) 부분문자열 위험 — 앵커 없는 수치의 과다계상**

```
$ grep -roh "agent_type" internal/hook | wc -l
48
$ grep -roh "subagent_type" internal/hook | wc -l
20
```

앵커 없는 48 중 20 이 `subagent_type` 이므로 실제 `agent_type` 은 **28**. 앵커 없이 센 48 을 인용하면 71% 과다계상이며 **재측정해도 같은 48 이 나와 검산으로 안 잡힌다.**

**(11) `agent_id` 선언 — 넷째 프로브이자 판별자 후보**

```
$ grep -n '^\s*AgentID' internal/hook/types.go
239:	AgentID              string `json:"agent_id,omitempty"`
```

**(12) 계측기 실행 경로 — moai 재빌드 없이 성립**

```
$ sed -n '37p' .claude/hooks/moai/handle-pre-tool.sh
payload=$(head -c 1048576)
```

래퍼가 stdin 원문을 먼저 보유하므로 C1/C2 계측기는 재빌드·설치본 치환을 거치지 않는다(§A.4). **단 C3(Go `TraceEntry` 확장)를 택하면 그 위험은 되살아난다** — 위험의 부재는 경로에 붙는 성질이다.

**(13) Tier 판정 근거 — 예산과 아티팩트 집합**

`spec-workflow.md` § SPEC Complexity Tier 실측: Tier S 는 파일 <5 · REQ/AC 천장 8/8 · 2아티팩트, Tier M 은 파일 5-15 · 천장 16/16 · 3아티팩트, Tier L 은 파일 >15 또는 constitutional · 천장 25/25 · 5아티팩트. 이 SPEC 은 동반이동 집합 10파일 · REQ 10 / AC 12(실측: `grep -cE '^### AC-BGX-' acceptance.md` = 12) · 3아티팩트로 **세 기준 모두 M**이며, LOC 만 S 를 가리킨다(§A.5 에 그 반대 지표도 함께 기록).

### iter-5 개정 (D11 — 추적 범위 축, HEAD `fe8159ed9` 기준 측정)

**(19) 네 번째 축 재현 — 같은 뿌리·패턴·SHA 에서 두 명령이 다른 답을 낸다**

```
$ git rev-parse --short HEAD
fe8159ed9
$ git grep -l 'tool-spawned subagent' | wc -l
19
$ grep -rln --exclude-dir=.git 'tool-spawned subagent' . | wc -l
20
$ grep -rln --exclude-dir=.git 'tool-spawned subagent' . | grep reports
.moai/reports/t1064/verdict.md
```

델타 1 은 미추적 `verdict.md` 다(§E.1 iter-4 (17) 에서 ignore 아님·미추적임을 이미 확정). **뿌리도 패턴도 시점도 고정인데 수가 갈린다** — 따라서 이것은 앞의 세 축과 독립된 **네 번째 축**이며, 감사의 발견이 맞다.

**조치**: AC-BGX-009 (1)을 네 좌표(뿌리·패턴·시점·추적 범위) 보고 의무로 재구성하고, 추적 범위를 뿌리와 **대칭**으로(두 수를 모두 적도록) 올렸으며, plan 시점 상수를 합격 비교 대상에서 배제했다.

**D11 서술 중 이 트리에서 어긋난 것 2건**(구현은 그대로 했고, 기록만 남긴다):

- **「추적 범위 미결속(✗)」은 `fe8159ed9` 기준으로는 이미 부분 결속돼 있었다.** iter-3 이 (1)에 「(b)와 (c) 중 어느 것을 인용하는지 명시한다」를 넣었기 때문이다. 감사가 `8852cb921` 을 baseline 으로 잡았으므로 그 시점 기준으로는 정확한 지적이다. **남아 있던 실제 결함은 결속의 세기**였다 — 뿌리는 「두 수 모두」인데 추적 범위는 「어느 쪽인지만 밝혀라」로 약했다. 이번에 대칭으로 올렸다.
- **「(1) 본문에 상수 8 이 박혀 있다」의 위치가 다르다.** iter-3 이 (1) 본문의 `(8)` 은 이미 지웠고, 남아 있던 것은 그 아래 중첩 blockquote 의 **「8 재확인으로 지나가면 FAIL」** 문장이었다. **결함 자체는 실재했고**(맨 상수가 합격 판정 문장에 있었다) 이번에 그 문장에서 상수를 제거했다.

**진짜로 결속되지 않았던 것은 시점(SHA) 좌표 하나**다. iter-3 의 (1)은 「그 시점에 다시 재라」고만 하고 **어느 SHA 였는지 기록하라고 요구하지 않았으며**, SHA 병기는 「run 값이 참조값과 다를 때」라는 조건부였다. 이번에 무조건 의무로 올렸다.

---

### iter-4 개정 (D9·D10, HEAD `8852cb921` 기준 측정)

**(17) D10 — 수치 정정: 넓은 뿌리 21 → 20, 이 SPEC 아티팩트 4 → 3**

```
$ git rev-parse --short HEAD
8852cb921
$ grep -rln --exclude-dir=.git "tool-spawned subagent" . | wc -l
20
$ grep -rln "tool-spawned subagent" .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/ | sort
.moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/acceptance.md
.moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/progress.md
.moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/spec.md
```

양성 대조 — `plan.md` 의 0 이 부재이지 프로브 고장이 아님(네 아티팩트 모두 판독 가능):

```
$ grep -c "tool-spawned subagent" .../plan.md
0
$ wc -l .../{acceptance,plan,progress,spec}.md
     223 acceptance.md / 154 plan.md / 227 progress.md / 195 spec.md
```

**원인 — 반대 방향의 두 움직임.** iter-2 커밋 자신이 리터럴을 지웠다:

```
$ git show 3d636c991 -- .../plan.md | grep -n 'tool-spawned subagent'
135:-1. `tool-spawned subagent` 리터럴로 재열거(plan 시점 하한 8파일).
```

커밋별 개수(직접 측정):

```
$ git grep -l "tool-spawned subagent" 039af6915 | wc -l   → 20
$ git grep -l "tool-spawned subagent" 3d636c991 | wc -l   → 19
$ git grep -l "tool-spawned subagent" 8852cb921 | wc -l   → 19
```

**리드 정정에 대한 재정정 1건.** 「21 은 어느 커밋에도 대응하지 않는 편집 중간 상태」는 정확하지 않다. `verdict.md` 는 **무시 대상이 아니라 미추적**이다:

```
$ git check-ignore -v .moai/reports/t1064/verdict.md
.gitignore:259:!.moai/reports/*/verdict.md	.moai/reports/t1064/verdict.md
$ git ls-files --error-unmatch .moai/reports/t1064/verdict.md
error: pathspec '…' did not match any file(s) known to git
```

네거션(`!`)이 걸려 있으므로 ignore 가 아니라 **아직 커밋되지 않은 상태**다. 따라서 21 = `039af6915` 의 **워킹 트리**(커밋 트리 20 + 미추적 1)를 정확히 센 값이고, 편집 중간 상태가 아니다. 커밋 트리에 21 이 없는 이유는 그 +1 이 계속 미추적 쪽에 있기 때문이다. **재현 조건은 커밋 경계가 아니라 「(b) 워킹 트리와 (c) 커밋 트리 중 무엇을 셌는가」이며**, §A.2 는 그렇게 고쳐 썼다.

**(18) D9 — 감사창 중 HEAD 이동: 절차 결함과 그 귀속**

사실관계(직접 측정):

```
$ git show 3d636c991:.../spec.md | grep -c '키 집합 동등 비교'
0
$ grep -c '키 집합 동등 비교' .../spec.md      # HEAD 8852cb921
2
```

감사는 `3d636c991` 을 baseline 으로 브리핑받았는데, 감사가 읽은 워킹 트리에는 **그 커밋에 없던 텍스트**(키 집합 동등 비교 조항)가 들어 있었다. 그 텍스트는 직후 `8852cb921` 로 착지해 결과적으로 무해했으나, **반대로 갈 수도 있었다** — 그랬다면 PASS 가 이력에 없는 텍스트 위에 서게 된다.

**귀속(실제 있었던 대로):**

- **주 원인 — 리드의 배차 오류.** 같은 트리에서 개정과 감사를 **직렬화하지 않고** 동시에 돌렸고, 감사에게 넘긴 `git status` 판독이 시작 시점에 이미 낡아 있었다. 리드가 스스로 이렇게 기록하도록 지시했다.
- **내 몫 — 가진 신호를 추론하지 않았다.** 커밋 직전 `git status --short` 에서 `?? .moai/reports/t1064/` 를 **읽었고**, 그것이 감사 산출물이라는 것도 알고 있었다(그 판정서를 읽고 작업했으므로). 「판정서가 있다」에서 「감사 세션이 지금 살아 있을 수 있다」로 넘어가는 추론을 하지 않았고, 리드에게 창 상태를 묻지 않았다.
- **레인이 구조적으로 볼 수 없었던 것** — 감사 세션의 생존 여부 자체. 레인 위치에서 조회할 수단이 없다.

**재발 방지(이 카드 범위에서 기록만)**: 워크트리에 감사 산출물 경로(`.moai/reports/<card-id>/`)가 미추적으로 보이면, 커밋 전에 리드에게 창 상태를 묻는다. 이 카드는 그 규율을 코드나 룰로 만들지 않는다 — 범위 밖이다.

---

### iter-3 개정 (D8 — 측정 방법 고정, HEAD `3d636c991` 기준 측정)

**(14) camel 축 접두 열거 — 과다매칭원은 셋이고 둘이 대문자로 시작한다**

```
$ grep -rnoE '[A-Za-z]*agentType' --include='*.go' --include='*.json' internal/hook/ | sed 's/.*://' | sort | uniq -c
  22 agentType
   2 subagentType
   3 SubagentType
   2 TestAgentSpawnFixtureCarriesSubagentType

$ grep -rnoE '(^|[^A-Za-z])agentType' --include='*.go' --include='*.json' internal/hook/ | wc -l
22
```

22 + 2 + 3 + 2 = **29**, 앵커 적용값 22 와 **정확히 닫힌다**. 리드가 넘긴 열거를 인용하지 않고 이 트리에서 재현했으며 값이 일치했다.

**판정**: 「`subagentType` 을 빼라」로 쓴 수정은 **대문자로 시작하는 둘(5건)을 통과시킨다.** 앵커는 레시피에 붙는 패치이고 철자 변종마다 다시 뚫리므로, 결정적 판정(REQ-BGX-001)의 측정 방법을 **키 집합 동등 비교**로 고정했다 — `"subagent_type" ≠ "agent_type"` 은 키 비교에서 구조적으로 성립한다.

**(15) snake 축 앵커 직접 측정 — iter-2 의 파생값 Gap 을 닫는다**

```
$ grep -rnoE '(^|[^A-Za-z])agent_type' --include='*.go' --include='*.json' internal/hook/ | wc -l
28
```

iter-2 는 28 을 `48 − 20` 의 **뺄셈 파생값**으로 얻었고 그 사실을 Gap 으로 적었다. 이번에 앵커로 직접 재어 같은 값을 얻었으므로 **그 Gap 은 닫힌다**. (두 측정의 대상 범위가 다르다 — iter-2 의 48/20 은 `internal/hook` 전체 파일, 이번 28 은 `*.go`/`*.json` 한정. 같은 28 이 나온 것은 그 범위 밖에 `agent_type` 이 없다는 뜻이며, 두 수를 하나의 측정으로 합쳐 인용하지 않는다.)

**(16) 오독 방향** — 과다매칭은 **더 큰 수**를 만든다. 이 축에서 큰 수의 종착지는 `tool_input` 안의 `subagent_type` 이 `agent_type` 도착으로 읽혀 카드가 **「도달 가능」이라는 결정적 오답으로 닫히는 것**이다. 그래서 D8 과 D2 는 방향이 반대인데 **한 조항(키 집합 동등 비교)으로 함께 닫힌다** — 키 집합 캡처는 D2 가 「미도착」과 「프로브 안 한 키로 도착」을 가르기 위해 요구한 바로 그것이다.

---

### 흡수 후 재측정과 D9 보충 (2026-09-22, 리드 판정 (a) 기존 plan 채택 후)

리드가 배차문의 class B·신규 트리 전제를 정정하고 기존 plan 채택을 지시했다(리드가 배차 전 §30 전제
판정을 생략한 것을 스스로 기록). run 진입 전 §4.1 흡수를 먼저 집행했다 — 배차문의 `758314007` 은 배차
시점값이고 흡수 시점 로컬 develop 은 `ef3ad83e2`(t1055 병합분 포함)였다.

**(20) D9 보충 — 사건 #2: 판정 확정 중 동결 해제·감사자 미통지 (iter-4 판정서 항목 3의 OPEN을 닫는다)**

iter-4 (18) 이 담은 것은 사건 #1(배차 오류로 개정과 감사가 같은 트리에서 동시 실행)뿐이었다. 사건 #2는
감사가 판정을 확정하는 동안(§12 진행 중) 레인이 동결을 풀고 `8852cb921` 을 커밋했으며 감사자에게 통지하지
않은 것이다. 귀속: 감사 진행 중임을 레인이 알았으므로(그 판정서를 읽고 작업하던 중이었음) 통지 누락은
레인 몫이다. 재발 방지: 감사 창 중 커밋은 감사자에게 사전 통지하거나 감사 종료 뒤로 미룬다 — 이 카드
범위에서는 기록만 남기고 룰화하지 않는다((18) 과 같은 판단).

**(21) 흡수 트리 재측정 — 기준 수치 (b) 의 무시-범위 이동 (HEAD `1b69ee405`)**

```
$ git rev-parse --short HEAD
1b69ee405
$ grep -rln "tool-spawned subagent" .claude internal | wc -l      → 8   (참조값 일치)
$ git grep -l "tool-spawned subagent" | wc -l                      → 19  (참조값 일치)
$ grep -rln --exclude-dir=.git "tool-spawned subagent" . | wc -l   → 19  (참조값 20 과 어긋남)
```

어긋남의 기제(직접 판별, 추론 아님): 이 트리의 `grep` 은 **ugrep 7.8.4** 로 `.gitignore` 를 존중한다.
흡수가 가져온 .gitignore 는 `113e487c2`(card t1059, 09-22, 의도적 철회)가 네거션
`!.moai/reports/*/verdict.md` 를 지운 판이라 `verdict.md` 가 무시 범위로 들어갔다. 파일 자체는 38행으로
여전히 적중한다 — **내용 변화가 아니라 계측기 무시-범위 변화다.** 순회 시작점에 따라 다르게 읽히는
것도 같은 기제의 관측이다(`.moai/` 뿌리 11 건 중 reports 1, 뿌리 순회 reports 0).

따라서: (1) AC-BGX-009 의 네 좌표(뿌리·패턴·시점·추적 범위)에 실무상 **다섯째 축 — 계측기 무시 범위
(.gitignore 상태)** — 이 더해진다. run 단계 비교에서 (b) 참조값 20 은 `8852cb921` 시점 무시 범위로 핀해
읽고, 흡수 트리 값 19 과는 좌표를 명시해 대조한다. (2) 흡수 트리에서 (b) 와 (c) 가 19 로 일치하는 것은
내용 동등이 아니라 **무시 귀산**이므로, 두 수의 일치를 동일성으로 인용하지 않는다.

측정 유의: 산출물 저작 시 복합 명령 한 건이 워크트리 가드에 거부됐고(프로세스 치환 속 git — plan §Gaps
기록과 같은 형태) 단순 명령으로 쪼개 재측정했다. 거부는 어떤 측정도 대체하지 않았다.

---

### Claim (iter-1, HEAD `69d61d371` 기준 — 원문 보존)

plan 단계 산출물 4종(`spec.md` / `plan.md` / `acceptance.md` / `progress.md`)이 저작됐고, SPEC 본문이 인용하는 모든 좌표·개수·원문은 **이 트리에서 이번 실행에** 측정됐다. SPEC ID 는 정규식 검사를 통과했고 기존 SPEC 과 충돌하지 않는다.

### Evidence

**(1) 트리 좌표**

```
$ git rev-parse --short HEAD && git branch --show-current && pwd
69d61d371
WT-branchguard-exempt
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064
```

**(2) `AgentType` 선언 두 곳 — 철자 축 (선언 앵커 `^\s*AgentType`)**

```
$ grep -n '^\s*AgentType' internal/hook/types.go internal/hook/subagent_stop.go
internal/hook/subagent_stop.go:162:	AgentType  string `json:"agentType,omitempty"`
internal/hook/types.go:230:	AgentType string `json:"agent_type,omitempty"` // Custom agent name if --agent flag used
```

**(3) `isExemptAgent` 선언과 유일 호출부**

```
$ grep -n '^func isExemptAgent\|isExemptAgent(' internal/hook/branch_guard.go
517:func isExemptAgent(input *HookInput) bool {
584:	if isExemptAgent(input) {
```

**(4) 주장 보유 지점 8파일 + 양성 대조**

```
$ grep -rln "tool-spawned subagent" .claude internal | sort
.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md
.claude/rules/moai/workflow/main-checkout-branch-guard.md
internal/hook/branch_guard_flagclass_test.go
internal/hook/branch_guard_quoted_test.go
internal/hook/branch_guard_test.go
internal/hook/branch_guard.go
internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard-detail.md
internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md

$ grep -rln "branchGuardExemptEnv" internal | sort
(8 files — 계측기 정상 발화)
```

**(5) 훅 trace 에 에이전트 정체성 필드 없음 — 양성 대조 동반**

```
$ grep -rc "AgentType" internal/hook/trace/entry.go
internal/hook/trace/entry.go:0
$ grep -rc "SessionID" internal/hook/trace/entry.go
internal/hook/trace/entry.go:2
```

양성 대조(`SessionID` 2행)가 발화했으므로 `0` 은 **부재**이며 계측기 고장이 아니다. 선행 카드 D12 의 「기존 로그로 이 축을 가릴 수 없다」가 이 트리에서 재확인된다.

**(6) 반대 주장 원문**

```
$ sed -n '114p' .claude/rules/moai/core/hooks-system.md
All hook events include `agent_id` and `agent_type` fields when triggered from a subagent context (v2.1.69+).

$ grep -c "tool-spawned subagent" .claude/rules/moai/workflow/main-checkout-branch-guard.md
1
```

**(7) SPEC ID 정규식 사전 검사**

```
$ ID="SPEC-BRANCHGUARD-EXEMPT-REACH-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
```

충돌 검사: `.moai/specs/` 의 branch-guard 계열은 `SPEC-WORKTREE-BRANCH-GUARD-{001,DISCRIM-001,FLAGCLASS-001,OPTIN-001}` 4건이며 이 ID 와 겹치지 않는다.

### Baseline-attribution

- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064`
- 측정 시점 HEAD: `69d61d371` (브랜치 `WT-branchguard-exempt`)
- 위 (1)~(7)은 전부 **이번 실행에서, 이 트리에 대해** 실행됐다. 배차문이 넘긴 좌표(types.go:230 / branch_guard.go:517 / :584)도 인용하지 않고 재측정했으며 값이 일치했다.

### Gaps

- **도달성 자체는 이 단계에서 측정되지 않았다.** plan 단계는 「무엇을 어떻게 잴지」를 설계했을 뿐이며, `isExemptAgent` 가 서브에이전트 컨텍스트에서 true 가 되는지는 **미측정**이다. 리드가 넘긴 등급 **PLAUSIBLE** 은 그대로 유지된다.
- **배선 철자 축 미측정.** 런타임이 서브에이전트 PreToolUse 에서 어느 철자를 보내는지는 재지 않았다. `internal/hook/testdata/agent_pretool_payload.json` 은 **메인 세션이 `Agent` 도구를 호출할 때의** PreToolUse 캡처라 이 질문에 답하지 않는다(그 페이로드에 정체성 필드가 없는 것은 예상된 바이며 부재 근거로 쓸 수 없다).
- **환경변수 축 미측정** — 설계상 범위 밖(carried-forward-unverified).
- **`internal/hook/pre_tool.go:1198` `frozenZonePrefixes`** 는 재지 않았다. 부수 발견으로만 기록했고 등급은 **PLAUSIBLE**.
- **테스트 미실행.** plan 단계에서 `go test ./internal/hook/...` 를 돌리지 않았다. 기준선은 run 단계 §C 사전 점검 항목이다.
- **거부된 도구 호출 1건**: 산출물 저작을 `cat > ... <<EOF` 형태의 compound 명령으로 시도했다가 워크트리 세션 가드에 거부됐다(heredoc 본문이 git 토큰을 담아 정적 검증 불가). 거부 후 Write 도구로 전환해 저작했으며, **이 거부는 어떤 측정도 대체하지 않았다.**

### Gaps (iter-2 추가분)

- **`agent_id` 가 PreToolUse 에 실제로 도착하는지 미측정.** 선언 필드의 존재(`types.go:239`)는 도착의 증거가 아니다 — §A.3 의 「판별자 후보」는 **후보**이지 확인된 판별자가 아니며, 그것이 AC-BGX-011 이 존재하는 이유다.
- **계측기를 실제로 넣어 보지 않았다.** §A.4 의 실행 가능성은 **구성요소의 실재**(래퍼가 원문 보유 · 훅이 서브에이전트 컨텍스트에서 발화)로 확립한 것이고 **끝까지 실행한 결과가 아니다.** 등급은 **measured-feasible** 이며 도달성 판정 자체는 여전히 **미측정**이다.
- **동반이동 집합 10 도 하한이다.** 두 뿌리와 두 패턴 축으로 쟀을 뿐, ~~20/21 파일~~ (iter-4 정정: ref `8852cb921` 기준 넓은 뿌리 **20**) 각각의 **문장 단위**가 같은 주장을 싣는지는 파일 단위 적중까지만 확인했다.
- ~~**`agent_type` 실제 적중 28 은 파생값이다.**~~ — **iter-3 에서 닫힘**: 앵커로 직접 재어 28 을 얻었다(§E.1 iter-3 (15)). 다만 그 측정은 `*.go`/`*.json` 한정 범위이므로 iter-2 의 48/20(전체 파일)과 **하나의 측정으로 합쳐 인용하지 않는다**.
- **camel 축 29 의 범위도 `*.go`/`*.json` 한정이다.** 그 밖의 파일 형식에 `agentType` 이 있는지는 재지 않았다 — 열거가 닫힌 것은 **잰 범위 안에서**다.
- **키 집합 동등 비교는 아직 한 번도 수행되지 않았다.** 비교의 입력(캡처된 키 집합)이 존재하지 않기 때문이며, 이 방법이 실제로 판정을 낸다는 것 자체는 **미측정**이다 — 방법을 고정했을 뿐이다.

### Residual-risk

- **계측기 설계가 실패하면 카드가 「미측정」으로 닫힌다.** 선행 카드가 이미 그렇게 닫혔고, 이 SPEC 은 그 결과를 정상 종료의 하나로 인정한다(§C). 다만 그때 CONTESTED 표식은 「또 미측정」으로 갱신돼야 하며 그대로 두면 다음 사람이 같은 벽을 재발견한다.
- **부정 분기의 조용한 통과.** 도달 불가가 관측되면 「독트린이 맞았다」로 닫고 싶은 압력이 강하다 — 기제 검증(AC-BGX-004)이 그 압력을 막는 유일한 장치다.
- **설치본 판정 위험.** 훅 경로가 이 트리 빌드가 아니라 설치본을 실행하면 계측기 없는 옛 빌드가 조용히 「필드 없음」을 낸다(VCI §2.2). run 단계 M2 1항이 이것을 막는다.
- **8파일은 하한.** 굵게 표기·줄바꿈이 낀 같은 주장은 리터럴 grep 에 안 잡힌다. 문안 정정 시 패턴을 흔들지 않으면 일부 지점이 옛 주장을 싣고 남는다.

---

## §E.2 Run-phase Evidence

### M2 선행 측정 — 계측기 발화 여부 (2026-09-23, 새 레인, 리드 배차)

**결과: 미발화. plan.md M2 1-5 는 진행하지 않았다(배차문 분기 규칙).**

- 절차: 세션을 primary 체크아웃에서 시작 → `EnterWorktree(.claude/worktrees/t1064)` → Bash 1회(`echo probe-1`, 08:18:47 로컬) → 이어 Bash 2회로 판독.
- 판독 (세 번째 Bash 호출 시점, 이 트리):
  - `ls -la ~/.moai/logs/t1064-instrument.log` → `No such file or directory`
  - `ls -la ~/.moai/logs/t1064-dump` → `No such file or directory`
  - 계측기 존재: 워크트리 사본 `.claude/hooks/moai/handle-pre-tool.sh` 의 `t1064` 적중 **5**, primary 사본 `/Users/goos/MoAI/moai-adk-go/.claude/hooks/moai/handle-pre-tool.sh` 적중 **0**.
  - 훅 명령 형태(양쪽 `settings.json`): `"${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-pre-tool.sh"` — 경로가 `CLAUDE_PROJECT_DIR` 로 해석된다.
  - `hook-missing.log`: 워크트리 쪽 파일 없음, primary 쪽 최근 2행은 `handle-config-change.sh` 뿐(pre-tool 누락 기록 없음).
  - HEAD `6ed2cb1f2` · 브랜치 `WT-branchguard-exempt` · `git status --porcelain` = ` M .claude/hooks/moai/handle-pre-tool.sh` 1행(전임자 계측기, 무변경).
- **cpd 값**: 계측기가 발화하지 않아 **훅 프로세스의 `CLAUDE_PROJECT_DIR` 은 캡처되지 않았다.** Bash 도구 프로세스 환경에서 읽은 값은 빈 문자열(`CLAUDE_PROJECT_DIR=`)이었으나, 이것은 훅 환경이 아니므로 cpd 근거로 쓰지 않는다.

**Gaps**
- 「primary 훅이 실행된다」는 가설과 **일관**될 뿐 확인은 아니다. 관측된 것은 워크트리 사본이 실행되지 않았다는 것까지이며, primary 사본이 실행됐는지 / PreToolUse 훅이 아예 돌지 않았는지는 이 측정으로 가르지 못한다(primary 사본에 발화 표식이 없음).
- 서브에이전트 Bash, 키 집합 비교, deny 억제는 수행하지 않았다(분기 규칙상 정지).

**Residual-risk**
- 세션 중간 `EnterWorktree` 는 훅 경로의 `CLAUDE_PROJECT_DIR` 을 바꾸지 않는 것으로 보인다 — 그렇다면 워크트리 사본 계측기는 이 방식으로는 원리상 발화할 수 없고, 워크트리에서 **세션을 시작**하는 경로(`moai cc -w t1064`)가 필요하다. 이 판단은 추론이며 우회책으로 실행하지 않았다.


### M2 본 측정 — 도달성 CONFIRMED (2026-09-27, HEAD `d49a4aec8`, 브랜치 `WT-branchguard-exempt`)

**등급: PLAUSIBLE → CONFIRMED.** 면제는 도구로 spawn 된 서브에이전트에서 도달 가능하며, deny 억제까지 실행으로 확인했다. 독트린 `main-checkout-branch-guard.md` § Mechanical Enforcement 의 「두 예외 축 모두 서브에이전트에서 도달 불가」는 **거짓**이다.

#### Claim

1. 이 세션의 Bash 호출에 PreToolUse 훅이 실제로 돈다.
2. 실행되는 래퍼 사본은 워크트리 것이 아니라 primary 체크아웃 것이다 — 세션 중간 `EnterWorktree` 는 훅의 `CLAUDE_PROJECT_DIR` 을 바꾸지 않는다.
3. 도구로 spawn 된 서브에이전트의 PreToolUse 페이로드에는 `agent_type` 과 `agent_id` 가 **snake_case 로 존재**한다. 메인 세션 호출에는 두 키가 **부재**한다.
4. `agent_type` 은 카탈로그 에이전트 이름에 한정되지 않고 **spawn 시 지정한 임의 이름**을 그대로 싣는다.
5. `agent_type == "manager-git"` 인 서브에이전트 형태 입력에서 BranchGuard 의 deny 가 **실제로 억제된다**(allow).

#### Evidence

**(1) 훅 발화 — 유일 표식 양성 대조.** 래퍼는 부분명령 수가 소프트캡 5를 넘으면 경고 1행을 `~/.moai/logs/hook-stderr.log` 에 적는다. 경계 43개 탐침을 쏘고 경계 0의 판독 명령으로 전후를 쟀다.

```
$ wc -l < ~/.moai/logs/hook-stderr.log      # 탐침 전
   53031
$ echo p1; ... ; echo done43                # 경계 43개
$ wc -l < ~/.moai/logs/hook-stderr.log      # 탐침 후
   53034
$ tail -n 1 ~/.moai/logs/hook-stderr.log
[moai:bash-risk] WARN: subcommand count 43 exceeds soft cap 5 — ...
```

애초 후보였던 `count 17` 은 같은 로그에 과거 **362회** 존재해(측정: `grep -c "count 17 exceeds"` = 362) 귀속 표식이 되지 못했다. 값을 43으로 올린 것이 이 측정의 성립 조건이다.

**(2) 실행 사본 판별.** 두 사본 모두 경고 코드를 보유하나(`grep -c bash_subcommand_soft_cap` = 각 5), 계측기는 워크트리 사본에만 있었다.

```
$ grep -c t1064 .claude/hooks/moai/handle-pre-tool.sh                                   # 워크트리
5
$ grep -c t1064 /Users/goos/MoAI/moai-adk-go/.claude/hooks/moai/handle-pre-tool.sh      # primary
0
$ ls ~/.moai/logs/t1064-dump ~/.moai/logs/t1064-instrument.log
No such file or directory  (둘 다)
```

계측기는 `payload=` 직후 행에 있고 **그 아래** 경고 코드는 발화했다. 따라서 실행된 파일은 「경고 코드는 있고 계측기는 없는」 사본 — primary 이다. 이로써 2026-09-23 선행 측정이 **추론**으로만 남겼던 잔여 위험이 확정되고, 동시에 그 측정의 Gap(「primary 가 돌았는지 / 훅이 아예 안 돌았는지 가르지 못함」)이 닫힌다.

**(3)(4) 키 집합 동등 비교.** 운영자 승인(2026-09-27)으로 primary 래퍼에 C2 계측기(최상위 키 이름 + 네 프로브 값만, 프롬프트 본문 없음)를 임시 삽입했다. 관측 원문 `~/.moai/logs/t1064-keys.log` 발췌:

```
tool=Bash keys=cwd,effort,hook_event_name,permission_mode,prompt_id,scratchpad_dir,
               session_id,tool_input,tool_name,tool_use_id,transcript_path
         probes=NONE                                              ← 메인 세션

tool=Bash keys=agent_id,agent_type,cwd,effort,hook_event_name,permission_mode,
               prompt_id,scratchpad_dir,session_id,tool_input,tool_name,
               tool_use_id,transcript_path
         probes={'agent_type': "'plan-auditor'", 'agent_id': "'a8e6f5ecb87b2613d'"}

         probes={'agent_type': "'manager-develop'", 'agent_id': "'a5a593b7e4697e72c'"}
         probes={'agent_type': "'t1261-m2-live'",   'agent_id': "'at1261-m2-live-beccc0362f9f5874'"}
```

판정은 grep 이 아니라 키 집합 동등 비교로 내렸다(plan.md M2 4항). 네 프로브 중 존재한 것은 `agent_type` · `agent_id` 두 snake 철자이며, camel 두 철자(`agentType` / `agentId`)는 어느 행에도 없다. `HookInput.AgentType` 의 태그가 `json:"agent_type"` 이므로 **철자가 일치**한다 — 디코더가 버리지 않는다.

세 번째 행의 `t1261-m2-live` 는 카탈로그 에이전트가 아니라 다른 세션이 spawn 시 붙인 이름이다. 즉 `agent_type` 은 spawn 이름을 그대로 싣는다.

**(5) deny 억제 — 세 팔.** 트리 빌드로 실제 핸들러를 실행했다(설치본 판정 위험 차단, VCI §2.2):

```
$ go build -o <scratch>/moai-t1064 ./cmd/moai
BUILD_OK from HEAD d49a4aec8
```

세 페이로드는 위 (3)에서 **측정된 키 집합**을 그대로 쓰고, `cwd` 는 primary 체크아웃, 명령은 패턴에 걸리되 실행돼도 무해한 `git checkout t1064-no-such-branch-probe` 로 고정했다.

| 팔 | `agent_type` | 결과 |
|---|---|---|
| A | (키 부재 — 메인 세션 형태) | `"permissionDecision":"deny"` · `BRANCH_GUARD_VIOLATION: git checkout <branch/-b> in primary checkout` |
| B | `plan-auditor` (측정된 서브에이전트 형태, 비면제) | `"permissionDecision":"deny"` · 같은 reason |
| C | `manager-git` | **`"permissionDecision":"allow"`** |

세 팔 모두 exit 0. A·B 는 공허 통과를 배제하는 대조다 — deny 는 메인 세션 형태에서도, 실제 관측된 서브에이전트 형태에서도 발화하므로, C 의 allow 는 **면제 외에 귀속될 곳이 없다.**

코드 좌표는 이 트리에서 다시 쟀다(배차문 좌표를 인용하지 않음): `internal/hook/branch_guard.go:593` `isExemptAgent`, `:603` `return input.AgentType == "manager-git"`, `:660` 그 호출이 `isPrimaryCheckout` 판정보다 **위**에 있다. 가드 활성: `.moai/config/sections/workflow.yaml:160-161` `branch_guard.enabled: true`.

#### Baseline-attribution

- 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064` · HEAD `d49a4aec8` · 브랜치 `WT-branchguard-exempt`
- 위 (1)~(5) 전부 이번 실행에서, 이 트리에 대해 실행됐다. (5)의 바이너리는 이 HEAD 에서 빌드한 것이며 설치본이 아니다.
- 계측기 제거 증명: primary 래퍼의 `t1064` 적중 0, 그리고 이 브랜치 HEAD 판과 sha256 **바이트 동일** — 양쪽 `0ebb97828a91f27c08a63362f0b9ec39e04ba6fc2f8934b8980de8bed7ac5536`. 앵커 인접성도 확인(`payload=$(head -c 1048576)` 다음 행이 원래의 `# Only meaningful for the Bash tool ...`).

#### Gaps

- **살아 있는 서브에이전트가 `manager-git` 이라는 이름으로 실제 Bash 를 쏘는 것까지는 실행하지 않았다.** (5)는 관측된 페이로드 형태를 실제 핸들러에 먹인 것이고, 런타임이 그 이름을 실을 것이라는 근거는 (4)의 임의 이름 관측이다. 두 조각을 잇는 추론이 한 단계 남아 있다 — 다만 그 한 단계를 실행하려면 primary 체크아웃에서 브랜치 상태를 실제로 바꾸는 명령이 통과해야 하므로 의도적으로 하지 않았다.
- **계측기 삽입 직전 sha256 을 캡처하지 않았다.** 제거 증명은 「이 브랜치 HEAD 판과 동일」로 성립하지만, primary 가 체크아웃하고 있는 `main` 판과의 대조는 아니다(가드가 cross-tree git 을 거부).
- **`MOAI_BRANCH_GUARD_EXEMPT` 환경변수 축은 재지 않았다.** 이번 측정은 신원 축 하나만 판정한다.
- **`internal/hook/pre_tool.go` `frozenZonePrefixes`** 는 범위 밖. 선행 기록의 부수 발견으로 남고 등급은 PLAUSIBLE 그대로다.
- **테스트 미실행.** 이번 단계에서 `go test ./internal/hook/...` 를 돌리지 않았다.

#### Residual-risk

- **계측기가 다른 세션의 호출까지 기록했다.** 남긴 것은 키 이름과 에이전트 이름뿐이고 프롬프트 본문·명령 문자열은 아니지만, `~/.moai/logs/t1064-keys.log` 는 측정 산물이므로 처분이 필요하다.
- **deny 문안 자체가 틀린 주장을 싣고 있다**(`branch_guard.go:682`: "fire only for main-thread launches, not for tool-spawned subagents"). 문안 정정 시 이 지점은 사용자에게 직접 보이는 표면이므로 동반이동 집합에서 빠지면 안 된다.
- **처방은 아직 정해지지 않았다.** CODE 분기는 `agent_id` 가 페이로드에 실재함이 확인됐으므로 **구성 가능**하다(M4-b 미발동). 즉 「고를 수 없어서 DOC」이 아니라 양쪽이 모두 후보인 상태이며, 결정은 운영자 게이트다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
