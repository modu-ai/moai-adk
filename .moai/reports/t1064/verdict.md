# plan 감사 판정서 — SPEC-BRANCHGUARD-EXEMPT-REACH-001 (card t1064)

- 감사자: plan-auditor · 반복 **iter-1 / 3**
- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064` · 브랜치 `WT-branchguard-exempt` · HEAD `039af6915`
- **Verdict: FAIL** · Overall Score **0.8125** (Tier 판정은 D1 참조 — 프론트매터가 `tier:` 를 선언하지 않아 Tier L 로 해소되고, 그 임계는 **0.85**)
- M1 Context Isolation: 배차문이 전달한 리드의 판독 결론은 **입력이 아니라 검증 대상**으로만 취급했다. 배차문이 넘긴 좌표(`types.go:230`, `subagent_stop.go:162`, `branch_guard_flagclass_test.go:228-246`, `branch_guard.go:30-33`)는 인용하지 않고 이 트리에서 전부 재측정했다.

---

## 1. Must-Pass 결과

| 항목 | 판정 | 근거 |
|---|---|---|
| MP-1 REQ 번호 정합 | **PASS** | `REQ-BGX-001`~`009`, `spec.md:71-79` 연속 9건·중복 0·자릿수 3자리 균일 |
| MP-2 GEARS 형식 | **PASS (요건층)** | 9건 전부 `shall` / `shall not` 모달리티와 응답을 명시. 8건은 `…때, 시스템은 …해야 한다(shall)` 사건구동형. **판정 층 명시**: 이 판단은 `spec.md` 의 `REQ-XXX` **요건층**에 대해 내렸고, `acceptance.md` 의 Given-When-Then `AC-XXX` 는 검증층이므로 MP-2 대상이 아니다(§4 에서 별도 채점) |
| MP-3 프론트매터 유효성 | **PASS** | 정본 12필드 전부 존재(`spec.md:2-13`), 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건 |
| MP-4 언어 중립성 | **N/A (auto-pass)** | `module: "internal/hook"` — 단일 프로그래밍 언어(Go) 범위 SPEC |
| MP-5 D7 교차 SPEC | **PASS** | `spec.md` 본문의 SPEC-ID 참조는 자기 자신 1건뿐. `progress.md` 가 참조하는 형제 4건은 전부 `status: completed`(retired/superseded/archived 0건) |
| MP-6 D8 크로스플랫폼 | **PASS (auto)** | `syscall` 리터럴 0건 |
| MP-7 해명 게이트 | **PASS** | `[NEEDS CLARIFICATION` 0건, 양성 대조 발화 확인 |

### MP-5 / MP-6 / MP-7 증거

```
$ grep -oE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/spec.md | sort -u
SPEC-BRANCHGUARD-EXEMPT-REACH-001

$ for s in SPEC-WORKTREE-BRANCH-GUARD-001 SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001 \
           SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001 SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001; do
      printf '%s: ' "$s"; grep -m1 '^status:' ".moai/specs/$s/spec.md"; done
SPEC-WORKTREE-BRANCH-GUARD-001: status: completed
SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001: status: completed
SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001: status: completed
SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001: status: completed

$ grep -rn 'syscall' .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/ ; echo "syscall_exit=$?"
syscall_exit=1

$ grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/ ; echo "exit=$?"
exit=1
$ grep -rc 'HARD' .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/plan.md   # 양성 대조
.moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/plan.md:3
```

양성 대조(`HARD` 3행)가 발화했으므로 `NEEDS CLARIFICATION` 의 0 은 **부재**이며 grep 고장이 아니다.

---

## 2. 차원 점수

| 차원 | 점수 | 루브릭 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 — 한두 항목에 경미한 모호성 | `spec.md:72` REQ-BGX-002 는 주어가 생략돼 `<subject> shall` 형태가 아님 · `acceptance.md:102-112` AC-BGX-007b 의 「사칭한 서브에이전트 형태 입력」이 미정의(D4) |
| Completeness | 0.75 | 0.75 — 비핵심 누락 1건 이상, 프론트매터는 완전 | `tier:` 미선언(D1) · §A.2 동반이동 집합이 `hooks-system.md` ×2 를 누락(D3) · 프로브 집합이 `agent_id` 누락(D2) |
| Testability | 0.75 | 0.75 — 판단이 개입하는 AC 존재 | AC-BGX-004 가 결과 (3) 에서 판정 불능(D2 결과) · AC-BGX-007b 가 기존 테스트로 만족 가능(D4). 반면 AC-BGX-006 의 `--is-ancestor X X` 공허 차단과 AC-BGX-010 의 타임아웃 판별식은 **모범적** |
| Traceability | 1.00 | 1.0 — 고아 0·미커버 0 | `acceptance.md:13-24` 매트릭스에서 REQ-BGX-001~009 전부 1개 이상 AC 로 덮이고, AC 11건 전부 유효 REQ 또는 §C/§D 를 지목 |

**Overall = (0.75 + 0.75 + 0.75 + 1.00) / 4 = 0.8125**

`tier:` 부재로 Tier L(임계 0.85)이 적용되므로 **0.8125 < 0.85 → FAIL**. 참고로 `tier: M` 이 선언되면 임계는 0.80 이 되어 같은 점수가 PASS 가 된다 — 이것이 D1 을 optional 이 아니라 blocking 으로 분류한 이유다.

---

## 3. 배차문 5개 질문에 대한 답

### Q1 — 이 SPEC 은 실행 가능한가, 아니면 D12 가 impracticable 로 닫은 프로브를 조용히 재계획하는가

**Claim.** 재계획이 **아니다**. SPEC 은 D12 가 부딪힌 중첩 `claude -p` 경로를 **명시적으로 포기**하고 다른 경로로 대체했으며, 그 대체 경로는 이 트리에서 **실행 가능하다**. 결정적 측정을 싣는 AC 는 **AC-BGX-002**(배선 철자)와 **AC-BGX-003**(deny 억제까지의 경로 실행) 둘이다.

**Evidence.**

(1) SPEC 이 D12 의 벽을 입력으로 받아 재시도를 금지한 원문 — `plan.md` §B:

```
| 중첩 `claude -p` 프로브 | 워크트리 세션 가드가 거부(선행 카드가 2회 관측) | **재시도하지 않는다.** M2 의 계측기 경로로 대체 |
```

(2) D12 의 기록이 커밋된 소스에 실재함을 재확인 — `internal/hook/branch_guard_flagclass_test.go:233-245` 발췌:

```
//   - AgentType axis (CONTESTED — left contested by this card, see below):
// CONTESTED-AXIS CAPTURE OUTCOME (D12, audit iteration 2): the mandated
// capture of one real tool-spawned PreToolUse payload is IMPRACTICABLE in this
// card's environment — the nested `claude -p` probe is refused by the runtime
// worktree-session guard ...
// The axis therefore REMAINS CONTESTED;
```

(3) **대체 경로의 실행 가능성 — 이 감사가 직접 측정한 핵심 근거.** PreToolUse 훅은 Go 바이너리가 아니라 **워크트리 자신의 셸 래퍼**를 통해 돈다:

```
$ python3 -c "...json.load(open('.claude/settings.json'))['hooks']['PreToolUse']..."
   "args": [ "-c", "[ -f \"$0\" ] && exec bash \"$0\"; ...",
             "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-pre-tool.sh" ],
   "matcher": "Write|Edit|Bash"
```

그 래퍼는 stdin 원문을 셸 변수에 **통째로 담아 두고** moai 에 넘긴다 — `.claude/hooks/moai/handle-pre-tool.sh` 발췌:

```
payload=$(head -c 1048576)
...
if command -v moai &> /dev/null; then
    printf '%s' "$payload" | moai hook pre-tool
```

따라서 M1 후보 C1/C2(원문 덤프 / 키 집합 기록)는 **moai 재빌드도, 설치본 치환도 없이** 이 워크트리의 래퍼 한 줄로 성립한다. `plan.md` M2 1항과 `acceptance.md` §D.1 셋째 경계사례가 경고한 「설치본이 판정」 위험은 이 경로에서는 **구조적으로 발생하지 않는다**.

(4) **훅이 tool-spawned subagent 컨텍스트에서 실제로 발화함을 이 트리에서 관측했다.** 이 감사 세션 자체가 Agent 도구로 spawn 된 서브에이전트이며, 그 Bash 호출이 워크트리의 `.moai/logs` 에 PreToolUse 를 남겼다:

```
$ head -1 .moai/logs/trace-8a224cdc-702d-44e5-af56-54e041b79b5c.jsonl
{"ts":"2026-09-21T22:18:56.895180083+09:00","event":"PreToolUse","handler":"*hook.preToolHandler",
 "tool":"Bash","duration_ms":29,"decision":"allow","session_id":"8a224cdc-702d-44e5-af56-54e041b79b5c"}
```

`session_id` 가 비어있지 않은 값으로 실려 있다(양성 대조). 그리고 같은 행에 에이전트 정체성 필드는 없는데, 그것은 스키마가 8필드뿐이기 때문이다:

```
$ grep -nE 'json:' internal/hook/trace/entry.go
11: ts / 13: event / 15: handler / 17: tool / 19: duration_ms / 21: decision / 23: reason / 25: error / 27: session_id
```

즉 **기존 로그로는 이 축을 가릴 수 없다**(D12 의 둘째 관측이 이 트리에서 재확인됨) — 그러나 **훅 경로 자체는 살아 있고 서브에이전트 컨텍스트에서 발화한다**. 계측기를 래퍼에 넣기만 하면 되는 상태다.

**Baseline-attribution.** 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064`, HEAD `039af6915`. (1)~(4) 전부 이번 실행에서 이 트리에 대해 실행했다. `.moai/logs/trace-…jsonl` 은 이 감사 세션(2026-09-21 22:18 이후)이 생성한 것이다.

**Gaps.**
- **계측기를 실제로 넣어 보지는 않았다.** 감사 범위가 판정서 + 스크래치패드 쓰기로 한정돼 있어 `handle-pre-tool.sh` 를 편집하지 않았다. 따라서 「경로가 실행 가능하다」는 **구성요소의 실재**(래퍼가 원문을 보유 / 훅이 서브에이전트에서 발화)로 뒷받침한 것이고, **끝까지 실행한 결과는 아니다**. 이 항목의 등급은 **실행 가능성 확립(measured-feasible)** 이며 **도달성 판정 자체는 여전히 미측정**이다.
- 서브에이전트 spawn 이 run 단계에서 실제로 일어나는지(2차 spawn 의 권한·깊이)는 재지 않았다. 다만 이 세션의 존재가 depth-1 spawn 이 성립함을 보인다.
- **거부된 도구 호출 2건**(§3.1 규율에 따라 명시): ① `python3` 키 집합 집계를 `&&` 체인으로 묶은 명령이 워크트리 가드에 거부(`"too complex to verify that it stays inside the worktree"`) ② `comm -23 <(...)` 명령이 `.git` 문자열 때문에 거부(`"names git in a form too complex"`). 둘 다 **단순 명령으로 분할해 재실행했고, 어떤 측정도 대체하지 않았다.**

**Residual-risk.** 래퍼 편집은 워크트리 안이지만 `.claude/hooks/moai/` 는 `moai update` 관리 대상 뿌리이며 `handle-pre-tool.sh` 는 템플릿에서 배포된다(CLAUDE.local.md §2.3). run 단계가 계측기를 넣고 **되돌리지 않으면** 미추적/수정 상태가 남는다. SPEC 은 M1 에서 「위생 처분(커밋 여부)」을 산출로 요구하지만 **되돌림 자체를 요구하는 AC 는 없다**(D6, optional).

---

### Q2 — 배선 철자를 측정값으로 고정하는가

**Claim.** 고정한다. 다만 **프로브 집합이 좁다** — 정작 다투는 독트린 문장이 이름을 대는 `agent_id` 가 빠져 있다(D2, blocking).

**Evidence.** 두 `AgentType` 선언은 이 트리에서 재측정했고 배차문 좌표와 일치한다:

```
$ grep -n '^\s*AgentType' internal/hook/types.go internal/hook/subagent_stop.go
internal/hook/subagent_stop.go:162:	AgentType  string `json:"agentType,omitempty"`
internal/hook/types.go:230:	AgentType string `json:"agent_type,omitempty"` // Custom agent name if --agent flag used
```

`isExemptAgent` 가 읽는 것은 snake 쪽이다 — `internal/hook/branch_guard.go:517-528`:

```
func isExemptAgent(input *HookInput) bool {
	if os.Getenv(branchGuardExemptEnv) == "1" { return true }
	if input == nil { return false }
	return input.AgentType == "manager-git"
}
```

SPEC 은 두 철자를 **모두** 프로브하라고 [HARD] 로 못박는다 — `acceptance.md:45-47`:

```
**[HARD]** 두 철자를 **모두** 프로브해야 한다. 한쪽만 재고 「부재」로 적으면 camel 도착을
구조적으로 못 보며, 그것은 도달 가능한 값이 도달 불가로 보이는 상태다.
```

그리고 §D.1 이 「camel 로 도착하고 값이 비어있음」을 「도착 아님」으로 미리 가른다. **이 두 장치는 정확하고 잘 설계됐다.**

문제는 집합의 폭이다. 다투는 독트린 문장 자체가 **두 필드**를 말한다 — `.claude/rules/moai/core/hooks-system.md:114`:

```
All hook events include `agent_id` and `agent_type` fields when triggered from a subagent context (v2.1.69+).
```

그리고 `agent_id` 는 이미 `HookInput` 의 선언된 필드다 — `internal/hook/types.go:238`:

```
	AgentID              string `json:"agent_id,omitempty"`
```

AC-BGX-002 의 프로브 집합은 `agent_type` / `agentType` 둘뿐이고 `agent_id` 가 없다.

**Baseline-attribution.** 트리 `t1064`, HEAD `039af6915`, 이번 실행 측정.

**Gaps.** 런타임이 실제로 무엇을 보내는지는 재지 않았다 — 그것이 이 카드의 본체이고 이 감사의 일이 아니다. `agent_id` 가 PreToolUse 에 실제로 도착하는지도 **미측정**이다(선언 필드의 존재는 도착의 증거가 아니다).

**Residual-risk.** `agent_id` 를 프로브에 넣지 않은 채 결과 (3)「둘 다 부재」로 닫으면, 정체성이 `agent_id` 로만 도착하는 세계를 **구조적으로 못 본다**. 그것은 AC-BGX-002 의 [HARD] 가 camel 에 대해 막으려는 실패 모드와 **정확히 같은 형태**이며, 한 필드만 빠진 것이다.

---

### Q3 — AC-BGX-004 가 결론 일치만으로 통과하는 것을 실제로 금지하는가

**Claim.** **금지한다.** 문언이 명시적이고, 이 항목은 이 SPEC 에서 가장 잘 쓰인 AC 중 하나다. 다만 **결과 (3) 분기에서 판정 불능 상태에 도달할 수 있다**(D2 의 결과, blocking 아님 — D2 수정으로 함께 닫힘).

**Evidence.** `acceptance.md:62-70` 원문:

```
**Then** 판정서는 **결론**(면제가 발화하지 않음)과 **서술된 기제**(메인 스레드 런치에서만
채워짐)를 **분리해** 각각 판정하고, 기제가 관측과 다르면 그 불일치를 독트린 결함으로 기록한다.

**실패 조건**: 결론 일치만으로 「독트린 정확」이라고 적으면 이 AC 는 FAIL 이다.
```

`plan.md` M3 이 같은 것을 세 명제로 분해해 (1)·(2) 가 (3) 의 이유인지를 별개 명제로 못박고, §E 반패턴이 「부정 분기를 무료 통과로 읽기」를 명시적으로 금지한다. **이 세 겹은 배차문이 요구한 함정 방어를 충족한다.**

**Gaps.** AC-BGX-004 는 「기제」 판정을 요구하는데, 프로브 집합이 두 철자로 닫혀 있으므로 결과 (3)에서 관측은 「이 두 키가 없다」까지만 말한다. 「런타임이 정체성을 안 보낸다」(독트린의 기제)와 「제3의 키로 보낸다」(다른 기제)를 **그 관측으로는 가를 수 없다**. AC-BGX-001 은 양성 대조로 **필드 1개**(`hook_event_name` 또는 `session_id`)만 요구하고 **최상위 키 집합 전체**를 요구하지 않는다 — `plan.md` M1 의 후보 C2 가 그 집합을 기록하는 형태지만 **AC 로 승격돼 있지 않다**.

**Residual-risk.** 계측기가 C1/C3 로 선택되고 키 집합이 기록되지 않으면, AC-BGX-004 는 **판정해야 하는데 판정 근거가 없는** 상태로 도달한다. 그때 가장 쉬운 출구가 「결론이 맞으니 독트린도 맞다」이며, 그것이 바로 이 AC 가 막으려는 것이다.

---

### Q4 — 변이 두 방향이 존재하고 서로의 증거로 만족되지 않는가

**Claim.** **둘 다 존재하고, 서로의 증거로는 만족되지 않는다.** 그러나 **AC-BGX-007b 는 기존 테스트로 만족 가능한 형태로 쓰여 있고**, 판별자(discriminator)가 미정의다(D4, blocking).

**Evidence.** 두 방향의 입력과 기대 출력이 정반대다 — `acceptance.md:92-112`:

```
### AC-BGX-007a — 무변이 성공 … 정당한 `manager-git` 정체성 경로를 입력할 때,
**Then** 면제가 여전히 발화해 deny 가 나오지 않아야 한다.
**배제하는 실패 모드**: 「가드가 전부 거부한다」.

### AC-BGX-007b — 변이 검출 … 그 정체성을 사칭한 서브에이전트 형태 입력을 줄 때,
**Then** deny 가 실제로 발생해야 한다.
**배제하는 실패 모드**: 「가드가 전부 허용한다」.
```

입력(정당 `manager-git` vs 사칭 형태)과 기대 출력(no-deny vs deny)이 둘 다 반대이므로, 한쪽 증거가 다른 쪽을 만족시키는 경로는 없다. **[HARD] 로 둘 다 필요함도 명시돼 있다.**

문제는 **「사칭한 서브에이전트 형태 입력」이 무엇인지 정의되지 않았다**는 점이다. 현행 코드에서 unit 층의 「서브에이전트 형태」는 이미 **`AgentType` 제로값**으로 구현돼 있고, 그 테스트가 이미 존재한다 — `internal/hook/branch_guard_flagclass_test.go:254-259`:

```
	input := &HookInput{
		ToolName:  "Bash", CWD: repo,
		ToolInput: json.RawMessage(`{"command": "git branch -f x y"}`),
		// AgentType deliberately left zero-valued: subagent-shaped payload.
	}
	decision, reason := checkBranchState(input, repo)
	if decision != DecisionDeny { t.Fatalf(...) }
```

이 테스트는 **AC-BGX-007b 의 문언을 이미 만족시킨다** — 그러나 아무것도 새로 고정하지 않는다. 진짜 사칭 형태는 `AgentType == "manager-git"` **이면서** 서브에이전트인 입력이고, 그것을 unit 층에서 만들려면 **「서브에이전트임」을 나타내는 판별 필드가 필요하다**. 현행 `HookInput` 에서 그 후보는 `agent_id`(`types.go:238`) 하나뿐이며, 그 도착 여부가 바로 AC-BGX-002 가 **프로브하지 않는** 것이다(D2).

**Gaps.** 판별자 후보가 `agent_id` 말고 더 있는지는 전수 조사하지 않았다. `HookInput` 선언(`types.go:208-250`)만 읽었고 런타임 페이로드는 재지 않았다.

**Residual-risk.** 측정 결과가 (1)「snake 로 도착」인데 `agent_id` 가 도착하지 않으면, **unit 층에 판별자가 존재하지 않아 M4 의 CODE 분기가 구현 불가능**해진다. `plan.md` M4 는 CODE/DOC 를 대등한 두 선택지로 제시할 뿐 **CODE 가 성립 불가능할 수 있다는 분기를 다루지 않는다.** 그 상태에서 007b 를 「형태만 맞는 기존 테스트」로 닫으면 가드가 강화됐다는 잘못된 기록이 남는다.

---

### Q5 — 8파일이 유지되는가 (다른 패턴 형태로 재측정)

**Claim.** **리터럴 8 은 `.claude` + `internal` 범위에서 유지된다.** 마크업·대소문자·하이픈·공백을 흔들어도 숫자가 **움직이지 않았다**. 숫자를 움직인 것은 **패턴이 아니라 탐색 뿌리**였다(8 → 20). 그리고 **동반이동 집합은 8 이 아니라 최소 10 이다** — 다투는 반대 주장을 싣는 `hooks-system.md` 와 그 템플릿 미러가 §A.2 에도 AC-BGX-009 에도 없다(D3, blocking).

**Evidence.**

(P1) manager-spec 의 패턴 재현 — 8 유지:

```
$ grep -rln "tool-spawned subagent" .claude internal | wc -l
8
```

(P2) 뿌리를 저장소 전체로 넓힘 — **20**:

```
$ grep -rln --exclude-dir=.git "tool-spawned subagent" . | wc -l   # 목록 20행
… .claude ×2, internal ×6, internal/template ×2 (= 위 8)
+ .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/{spec,plan,acceptance,progress}.md (4)
+ .moai/specs/SPEC-RC-TESTBED-001/{spec,plan,acceptance,research}.md (4)
+ .moai/specs/SPEC-WORKTREE-BRANCH-GUARD-FLAGCLASS-001/{spec,plan,acceptance}.md (3)
+ CHANGELOG.md (1)
```

(P3) **패턴을 흔든 재측정** — 대소문자 무시 + 하이픈/공백/마크업 허용:

```
$ grep -rlniE 'tool[-–[:space:]*_]*spawned[^A-Za-z]{0,6}subagent' . --exclude-dir=.git | wc -l
20
```

**P2 와 동일한 20.** 즉 마크업·대소문자·하이픈 변형으로 놓친 지점은 **0건**이다.

(P4) 코드 형태 식별자 프로브 `isExemptAgent` — 21파일이지만 `.claude`+`internal` 범위의 claim 보유자는 위 8과 동일(나머지는 `.moai/specs` 이력과 CHANGELOG).

(P5) 면제 토큰 보유자와의 차집합 — 6파일이 추가로 나오나, **주장 보유자가 아님**을 확인:

```
$ grep -rlE 'MOAI_BRANCH_GUARD_EXEMPT' .claude internal      # 13파일
… 8 외의 6: .claude/rules/local/gitflow-lane-protocol.md, internal/config/types.go,
   internal/hook/{branch_guard_worktree_test.go, pre_tool_branch_guard_optin_test.go,
   pre_tool_test.go, pre_tool.go}
$ grep -niE 'unreachable|subagent|main-thread' <위 6파일> | grep -iE 'exempt|guard|agent_type|AgentType|reach'
internal/config/types.go:972:	// but unreachable from any config file.
internal/config/types.go:1716:// the false branch. Unbound, that branch was unreachable from configuration:
```

두 적중 모두 「설정 파일에서 도달 불가」라는 **무관한 문맥**이다. 6파일은 환경변수 토큰만 싣고 도달성 주장은 싣지 않는다.

(P6) **누락된 동반이동 대상 — 반대 주장 보유자 2파일:**

```
$ ls -1 .claude/rules/moai/core/hooks-system.md \
        internal/template/templates/.claude/rules/moai/core/hooks-system.md
.claude/rules/moai/core/hooks-system.md
internal/template/templates/.claude/rules/moai/core/hooks-system.md

$ grep -n 'agent_id.*agent_type' .claude/rules/moai/core/hooks-system.md \
        internal/template/templates/.claude/rules/moai/core/hooks-system.md
.claude/…/hooks-system.md:114:All hook events include `agent_id` and `agent_type` fields when triggered from a subagent context (v2.1.69+).
internal/template/…/hooks-system.md:114:All hook events include `agent_id` and `agent_type` fields when triggered from a subagent context (v2.1.69+).
```

두 파일 모두 리터럴 `tool-spawned subagent` 를 싣지 않아 8에도 20에도 **이 축으로는** 잡히지 않는다(20 목록에 없음). 판정이 어느 쪽으로 나든 **모순하는 두 문장 중 하나는 반드시 바뀌어야 하므로**, 동반이동 집합은 8 이 아니라 **최소 10** 이다.

(P7) **`TestBranchGuard_DenyReasonRemediationContract` 는 deny reason 문안을 실제로 고정한다 — 확인됨.** `internal/hook/branch_guard_test.go:694-696`:

```go
	if !strings.Contains(reason, branchGuardExemptEnv) || !strings.Contains(reason, "tool-spawned subagents") {
		t.Fatalf("deny reason lacks the exemption reachability qualifier (main-thread-only): %q", reason)
	}
```

그리고 고정되는 문자열의 생산지 — `internal/hook/branch_guard.go:604-606`:

```go
	reason = fmt.Sprintf("%s: %s in primary checkout (use a worktree; the manager-git identity and %s exemptions fire only for main-thread launches, not for tool-spawned subagents)", ...)
```

즉 **축이 도달 가능으로 판정되어 deny reason 을 고치면 이 테스트가 즉시 깨진다.** `plan.md` §F 가 이 테스트를 「동반 이동 대상」으로 이름을 대고 있는 것은 정확하다 — 다만 **어떤 AC 도 이것을 묶지 않는다**(AC-BGX-009 는 `.claude` 사본 + 템플릿 미러 + `make build` 만 다룬다). D5(optional).

**Baseline-attribution.** 전부 트리 `t1064`, HEAD `039af6915`, 이번 실행. P1 은 `progress.md` §E.1 (4) 가 HEAD `69d61d371` 에서 잰 것과 같은 값이며, 이 감사는 그 값을 인용하지 않고 현재 HEAD 에서 재측정했다.

**Gaps.** 20파일 각각의 **문장 단위**가 같은 주장을 싣는지는 파일 단위 적중까지만 확인했고 전문 판독은 하지 않았다. `CHANGELOG.md` 와 닫힌 SPEC 아티팩트가 정정 대상인지(「관측 기록은 나중 사실에 맞춰 고치지 않는다」) 는 이 감사가 판정하지 않는다.

**Residual-risk.** AC-BGX-009 의 [HARD] 는 「패턴을 흔들어라」고만 말한다. 이 감사의 측정은 **흔들어도 숫자가 안 움직이고 뿌리를 넓혀야 움직인다**는 것을 보였으므로, 지시받은 대로만 수행한 run 단계는 **여전히 8 을 재확인하고 지나간다**.

---

## 4. 발견된 결함 (구조화 목록)

- **D1** — `spec.md:2-13` — 프론트매터에 `tier:` 가 없다. 스키마상 optional 이지만 부재는 **Tier L 로 해소**되고(`spec-frontmatter-schema.md:167`), Tier L 은 5아티팩트(`design.md`+`research.md` 포함)를 요구하는데 이 SPEC 은 3아티팩트만 있다. 동시에 감사 임계가 0.85 로 올라간다. REQ 9 / AC 11 은 Tier S 예산(8/8)을 초과하고 Tier M(16/16)에 들어맞으며, 현재 아티팩트 집합도 Tier M 과 정확히 일치한다. — **Severity: major · Class: blocking** — **필요 수정**: 프론트매터에 `tier: M` 추가(+ `updated:` 갱신). 다른 판정으로 Tier L 을 의도했다면 `design.md` 와 `research.md` 를 저작할 것.

- **D2** — `acceptance.md:39-47` (AC-BGX-002) — 프로브 집합이 `agent_type` / `agentType` 둘로 닫혀 있고 **`agent_id` 가 빠졌다.** `agent_id` 는 ① 다투는 독트린 문장(`hooks-system.md:114`)이 `agent_type` 과 **나란히** 이름을 대는 필드이고 ② 이미 선언된 `HookInput` 필드(`internal/hook/types.go:238`)이며 ③ AC-BGX-007b 의 「사칭 형태」를 unit 층에서 구성할 **유일한 판별자 후보**다. 이 누락은 AC-BGX-002 의 [HARD] 가 camel 에 대해 막는 실패 모드와 같은 형태다. 결과 (3) 분기에서 AC-BGX-004 가 기제를 판정할 근거도 함께 사라진다. — **Severity: major · Class: blocking** — **필요 수정**: AC-BGX-002 의 프로브 집합을 `agent_type` · `agentType` · `agent_id` · `agentId` 로 확장하고, AC-BGX-001 에 **최상위 키 집합 전체를 기록**하는 의무를 추가(= `plan.md` M1 후보 C2 를 AC 로 승격). 그래야 「제3의 키로 도착」이 「부재」와 구별된다.

- **D3** — `spec.md:60-70` (§A.2) 및 `acceptance.md:124-132` (AC-BGX-009) — 동반이동 집합이 8파일인데, **다투는 반대 주장을 싣는 `.claude/rules/moai/core/hooks-system.md:114` 와 그 템플릿 미러가 빠졌다.** 두 파일은 리터럴 `tool-spawned subagent` 를 싣지 않아 8에도 저장소 전체 20에도 이 축으로는 잡히지 않는다(이 감사 실측). 판정이 어느 방향으로 나든 모순하는 두 문장 중 하나는 반드시 정정되므로 집합은 최소 10이다. 덧붙여 AC-BGX-009 의 [HARD] 는 「패턴을 흔들어라」고만 지시하는데, 실측상 **흔들어도 0 이 움직이고 뿌리를 넓혀야 8→20 이 움직인다** — 변주 축 지정 자체가 틀렸다. — **Severity: major · Class: blocking** — **필요 수정**: §A.2 에 `hooks-system.md` 로컬+미러 2건을 명시 추가, AC-BGX-009 (1)의 재열거 의무를 「패턴 흔들기 **그리고** 탐색 뿌리 확대(저장소 전체)」로 고쳐 쓰고, 닫힌 SPEC 아티팩트·`CHANGELOG.md` 는 정정 대상에서 제외함을 명시(관측 기록 불변 원칙).

- **D4** — `acceptance.md:102-112` (AC-BGX-007b) — 「그 정체성을 사칭한 서브에이전트 형태 입력」이 정의되지 않았다. 현행 unit 층의 「서브에이전트 형태」는 `AgentType` 제로값이고, 그 형태의 deny 를 고정하는 테스트가 **이미 존재한다**(`branch_guard_flagclass_test.go:250-266`). 따라서 007b 는 **아무것도 새로 고정하지 않는 기존 테스트로 문언상 만족된다**. 진짜 사칭 형태(`AgentType=="manager-git"` **이면서** 서브에이전트)는 판별 필드 없이 구성 불가이며, `plan.md` M4 는 **CODE 분기가 성립 불가능할 수 있다는 경우를 다루지 않는다.** — **Severity: major · Class: blocking** — **필요 수정**: AC-BGX-007b 에 「사칭 형태」의 구성 요건을 명시(`AgentType == "manager-git"` **이면서** 서브에이전트 판별 필드가 부재/상이)하고, 그 판별자가 측정 결과 존재하지 않을 경우 **CODE 분기를 선택 불가로 선언**하는 조항을 `plan.md` M4 에 추가.

- **D5** — `plan.md` §F / `acceptance.md:124-132` — `TestBranchGuard_DenyReasonRemediationContract` 가 `branch_guard_test.go:694` 에서 `strings.Contains(reason, "tool-spawned subagents")` 로 deny reason 문안을 **기계적으로 고정**함을 이 감사가 확인했다. `plan.md` §F 는 이것을 동반이동 대상으로 이름을 댔지만 **어떤 AC 도 묶지 않는다** — AC-BGX-009 의 동반이동 의무는 `.claude` 사본·템플릿 미러·`make build` 만 다룬다. — **Severity: minor · Class: optional** — **필요 수정**: AC-BGX-009 (2)에 Go 테스트 동반이동(`branch_guard_test.go:694`, `branch_guard.go:604-606`)을 한 항목으로 추가.

- **D6** — `plan.md` M1 — 계측기의 **되돌림(revert)** 을 요구하는 AC 가 없다. M1 은 「위생 처분(커밋 여부)」을 산출로 요구할 뿐이다. 이 감사 실측상 계측기가 들어갈 곳은 `.claude/hooks/moai/handle-pre-tool.sh` 이며, 이는 `moai update` 관리 대상 뿌리이자 템플릿 배포 파일이다. 되돌리지 않으면 미추적/수정 상태로 남는다. — **Severity: minor · Class: optional** — **필요 수정**: AC-BGX-010 에 「계측기 제거 후 `git status --porcelain -- .claude/hooks/` 가 0행」을 한 줄 추가.

- **D7** — `spec.md:72` (REQ-BGX-002) — GEARS `<subject> shall <response>` 형태에서 주어가 생략됐다(다른 8건은 「시스템은」 또는 명시 주어를 가진다). 모달리티와 응답이 명시돼 있어 MP-2 는 통과하나 문형 일관성이 깨진다. — **Severity: minor · Class: optional** — **필요 수정**: 「…시스템은 … 경로 실행으로 판정해야 한다(shall)」로 주어 삽입.

---

## 5. 권고 (FAIL — iter-2 를 위한 수정 지시)

iter-2 재감사는 **위 결함 델타에만** 범위를 한정한다(전면 재감사 아님).

1. **D1** — `spec.md` 프론트매터에 `tier: M` 추가, `updated: 2026-09-21` 유지/갱신. (Tier L 을 의도했다면 대신 `design.md` + `research.md` 저작.)
2. **D2** — AC-BGX-002 프로브 집합을 4철자(`agent_type`/`agentType`/`agent_id`/`agentId`)로 확장. AC-BGX-001 에 **최상위 키 집합 전량 기록** 의무 추가.
3. **D3** — `spec.md` §A.2 에 `hooks-system.md` 로컬+템플릿 미러 2건 추가(집합 8 → 10). AC-BGX-009 (1) 을 「패턴 흔들기 + 탐색 뿌리 확대」로 개정하고, 닫힌 SPEC·CHANGELOG 제외를 명시.
4. **D4** — AC-BGX-007b 에 사칭 형태의 구성 요건 명시. `plan.md` M4 에 「판별자 부재 시 CODE 분기 선택 불가」 조항 추가.
5. (선택) D5·D6·D7 — optional. 운영자 재량. D5 는 한 줄로 끝나고 회귀 비용이 실재하므로 함께 처리할 것을 권한다.

**수정해서는 안 되는 것 — 이 SPEC 의 강점:**

- AC-BGX-004 의 결론/기제 분리와 그 실패 조건(Q3) — 배차문이 지목한 부정 분기 함정을 정확히 막는다.
- AC-BGX-005 의 「측정 불가 ≠ 도달 불가」 보존 — D12 가 남긴 구분을 지우지 않는 유일한 장치다.
- AC-BGX-006 의 `--is-ancestor X X` 공허 차단, AC-BGX-010 의 「순수 타임아웃은 `--- FAIL` 을 안 남긴다」 판별식 — 둘 다 이 저장소가 실제로 당한 실패에서 나온 정확한 조항이다.
- `plan.md` §B 의 「중첩 `claude -p` 재시도하지 않는다」 — Q1 의 답이 「재계획 아님」인 근거다.

---

## 6. 등급 보존 명세 (AC-BGX-009 / REQ-BGX-009 준수)

이 판정서가 인용한 선행 결론의 등급을 **문장 안에** 싣는다:

- 리드가 넘긴 「`AgentType` 축이 도달 가능할 수 있다」는 **PLAUSIBLE 등급이며 확정이 아니다** — 판독은 됐고 deny 억제까지의 경로는 실행되지 않았다. 이 감사도 그 등급을 **올리지 않았다**.
- D12 가 기록한 「중첩 `claude -p` 프로브 IMPRACTICABLE」은 **커밋된 소스에 CONTESTED 로 남아 있는 관측**이며, 이 감사가 이 트리에서 원문을 재확인했다.
- 「기존 훅 trace 로 이 축을 가릴 수 없다」는 **이 감사가 이 트리에서 이번 실행에 재측정한 값**이다(`trace/entry.go` 8필드, 양성 대조 `session_id` 실값 관측).
- 「계측기 경로가 실행 가능하다」는 **이 감사의 measured-feasible 등급**이며 **도달성 판정 자체는 여전히 미측정**이다 — 래퍼를 편집하지 않았다.
- `MOAI_BRANCH_GUARD_EXEMPT` 환경변수 축의 도달 불가는 선행 카드가 uncontested 로 기록한 것이고 **이 감사도 이 SPEC 도 재측정하지 않았다 — carried-forward-unverified 다.**
- `internal/hook/pre_tool.go:1198` `frozenZonePrefixes` 건은 이 감사가 재지 않았으며 **PLAUSIBLE 등급 그대로**다.

---

## 7. 검증 범위

```
$ go test ./internal/hook/... -timeout 30m
EXIT=0
ok  	github.com/modu-ai/moai-adk/internal/hook	229.447s
ok  	github.com/modu-ai/moai-adk/internal/hook/handoff	15.759s
ok  	github.com/modu-ai/moai-adk/internal/hook/memo	0.324s
ok  	github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy	0.690s
ok  	github.com/modu-ai/moai-adk/internal/hook/mx	18.503s
ok  	github.com/modu-ai/moai-adk/internal/hook/mx/complexity	0.346s
ok  	github.com/modu-ai/moai-adk/internal/hook/perf	41.893s
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	28.563s
ok  	github.com/modu-ai/moai-adk/internal/hook/security	9.910s
ok  	github.com/modu-ai/moai-adk/internal/hook/testutil	0.365s
ok  	github.com/modu-ai/moai-adk/internal/hook/trace	0.378s
```

종료 코드 **0**, 요약 행 전부 `ok` — **run 단계 기준선은 초록이다**(`plan.md` §C 사전 점검 항목의 선취). 판별식은 `--- FAIL` 세기가 아니라 요약 행 + 종료 코드다(AC-BGX-010 [HARD] 준수). `go test ./...` 는 **실행하지 않았다.**

---

## 8. 이 판정서 자신의 Gaps / Residual-risk

**Gaps.**
- 계측기를 실제로 넣어 도달성을 측정하지 **않았다** — 감사 범위 밖이며, 그것은 run 단계의 일이다.
- `.moai/specs/` 아래 어떤 파일도 편집하지 않았다(배차 제약 준수).
- 거부된 도구 호출 2건을 §3 Q1 Gaps 에 원문 요지와 함께 기록했다. 둘 다 분할 재실행했고 어떤 측정도 대체하지 않았다.
- 20파일 각각의 문장 전문 판독은 하지 않았다(파일 단위 적중까지).
- `agent_id` 가 PreToolUse 에 실제로 도착하는지는 **미측정** — 선언 필드의 존재는 도착의 증거가 아니다.

**Residual-risk.**
- 이 판정서의 점수 0.8125 는 Tier 해소에 의존한다. D1 이 `tier: M` 으로 닫히면 같은 점수가 임계 0.80 을 넘어 PASS 가 된다 — 그러나 D2·D3·D4 는 점수와 무관하게 **수정 대상**이며, 그 넷을 닫지 않고 tier 만 선언해 PASS 를 얻는 것은 이 판정의 의도가 아니다.
- 이 카드가 세 번째로 「미측정」으로 닫힐 위험이 여전히 가장 크다. §3 Q1 이 보인 바 계측기 경로는 실행 가능하므로, 그 위험은 이제 **환경 제약이 아니라 실행 여부**의 문제다.

---

## 9. 추가 측정 — 부분문자열 과다매칭 위험과 SPEC 레시피의 앵커 여부

리드가 배차 후 위험 하나를 전달했다. **인용하지 않고 이 트리에서 재측정했다**(리드 전달값은 아래 재현 결과와 일치했다).

### Claim

`agent_type` 은 `subagent_type` 의 **부분문자열**이므로 앵커 없는 grep 은 과다계상한다. 이 위험은 **이 SPEC 의 결정적 측정(AC-BGX-002)에 직접 걸리며**, 그 AC 는 **어떤 패턴도 지정하지 않아 앵커 의무를 상속하지 않는다.** 나아가 camel 축에는 과다매칭 토큰이 **하나가 아니라 셋**이라 「`subagentType` 을 빼면 된다」는 교정도 불완전하다. → **D8, blocking.**

### Evidence

**(1) snake 축 재현 — 리드 전달값과 일치**

```
$ grep -rno --include='*.go' --include='*.json' 'agent_type'            internal/hook/ | wc -l
48
$ grep -rno --include='*.go' --include='*.json' 'subagent_type'         internal/hook/ | wc -l
20
$ grep -rnoE --include='*.go' --include='*.json' '(^|[^a-z])agent_type' internal/hook/ | wc -l
28
```

`48 = 28 + 20` — 앵커 없는 수는 **71% 가 부풀려진 값**(48/28)이다.

**(2) camel 축 — 과다매칭원이 셋이다(리드가 다루지 않은 축)**

```
$ grep -rno  --include='*.go' --include='*.json' 'agentType'                 internal/hook/ | wc -l
29
$ grep -rno  --include='*.go' --include='*.json' 'subagentType'              internal/hook/ | wc -l
2
$ grep -rnoE --include='*.go' --include='*.json' '(^|[^a-zA-Z])agentType'    internal/hook/ | wc -l
22
```

`29 − 2 = 27 ≠ 22`. 차이 5의 정체를 접두사 전수로 확정했다:

```
$ grep -rnoiE --include='*.go' --include='*.json' '[a-z]*agentType' internal/hook/ \
    | sed -E 's/.*:([A-Za-z]*agentType)/\1/I' | sort | uniq -c | sort -rn
 108 AgentType        ← Go 필드명(와이어 태그 아님)
  22 agentType        ← 실제 와이어 철자
   3 SubagentType
   2 TestAgentSpawnFixtureCarriesSubagentType
   2 subagentType
   2 OfficialAgentType
   2 AbsentAgentType
```

소문자 `agentType` 을 부분문자열로 품는 것은 `agentType`(22) + `subagentType`(2) + `SubagentType`(3) + `TestAgentSpawnFixtureCarriesSubagentType`(2) = **29** — 산술이 정확히 닫힌다. 즉 camel 축의 과다매칭 토큰은 **`subagentType` 하나가 아니라 셋**이며, 그중 둘은 대문자 `S` 로 시작해 `subagentType` 만 빼는 교정에 잡히지 않는다.

**(3) 위험이 결정적 방향으로 실현된 실례 — `agent_pretool_payload.json`**

```
$ grep -no  'agent_type'                internal/hook/testdata/agent_pretool_payload.json
14:agent_type
$ grep -noE '(^|[^a-z])agent_type'      internal/hook/testdata/agent_pretool_payload.json ; echo "exit=$?"
exit=1
$ grep -n   'agent_type'                internal/hook/testdata/agent_pretool_payload.json
14:    "subagent_type": "Explore"
```

앵커 없는 1적중은 **정체성 필드가 아니라 `tool_input` 안의 Agent 도구 인자**다. 키 집합으로 확정(양성 대조 동반):

```
$ python3 -c "...json.load(...)..."
top-level keys: ['cwd', 'effort', 'hook_event_name', 'permission_mode', 'prompt_id',
                 'session_id', 'tool_input', 'tool_name', 'tool_use_id', 'transcript_path']
tool_input keys: ['description', 'prompt', 'run_in_background', 'subagent_type']
```

최상위에 `agent_type` · `agentType` · `agent_id` **모두 부재**. 양성 대조로 `hook_event_name` 과 `session_id` 가 실값으로 존재한다 — 공교롭게도 AC-BGX-001 이 양성 대조로 지명한 바로 그 두 필드다. 따라서 이 부재는 **실제 부재**이고 계측 고장이 아니다.

덧붙여 **이 키 집합 덤프가 D2 권고의 실물 예시**다: 최상위 키 전량을 찍으면 「어느 철자로 도착하는가」와 「제3의 키로 도착하는가」가 한 번에 갈리고, 부분문자열 과다매칭이 **구조적으로 발생할 수 없다**(키 비교는 완전일치이므로).

**(4) SPEC 자신의 레시피는 앵커돼 있는가 — 배차문이 물은 것**

| 레시피 위치 | 패턴 | 앵커 | 판정 |
|---|---|---|---|
| `progress.md` §E.1 (2) | `grep -n '^\s*AgentType'` | **있음**(줄머리 선언 앵커) | 이 SPEC 에서 유일하게 앵커된 철자 레시피. 다만 **AC 가 아니라 progress 기록**이라 run 단계를 구속하지 않는다 |
| `progress.md` §E.1 (5) | `grep -rc "AgentType" internal/hook/trace/entry.go` | **없음** | 결과가 0 이라 이번엔 무해했으나, 그 파일에 `SubagentType` 이 있었다면 조용히 과다계상됐다 |
| `progress.md` §E.1 (4) | `grep -rln "tool-spawned subagent"` | 해당 없음(구절) | 부분문자열 위험 없음 |
| **`acceptance.md` AC-BGX-002** | **패턴 미지정** | **없음** | **결정적 AC 가 철자 이름만 대고 측정 형태를 지정하지 않는다** |
| `acceptance.md` AC-BGX-009 | `tool-spawned subagent` + 「패턴을 흔든다」 | 없음 | 흔들라는 지시만 있고 앵커 의무가 없다 |

즉 **앵커된 레시피는 하나뿐이고 그것은 AC 가 아니다.** 결정적 AC(002)는 `agent_type` 과 `agentType` 을 **이름으로만** 지목하며, 이 감사의 (1)(2) 측정에 따르면 **두 철자 모두 부분문자열 과다매칭에 노출돼 있다.**

### Baseline-attribution

트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064`, HEAD `039af6915`. (1)~(4) 전부 이번 실행에 이 트리에서 실행했다. (1) 의 세 수(48/20/28)는 리드가 전달한 값과 일치하나, **인용이 아니라 재현**이다.

### Gaps

- 저장소 전체(`internal/hook/` 밖)의 과다매칭 노출은 재지 않았다. 범위를 `internal/hook/`, `*.go`+`*.json` 으로 한정했다.
- 런타임이 실제로 보내는 철자는 **여전히 미측정**이다 — 이 절은 「어떻게 재야 하는가」만 다룬다.
- `agent_pretool_payload.json` 이 어느 커밋에서 왔는지는 확인하지 않았다(배차문이 `b0d3b61f8` 을 댔으나 이 감사가 재지 않았으므로 **인용하지 않는다**).

### Residual-risk

이 위험의 방향이 나쁘다. 과다계상은 **큰 수**를 내고, 큰 수는 「더 많이 찾았다」로 읽혀 안전해 보인다. AC-BGX-002 의 결과 (1)/(2)—「비어있지 않은 값으로 도착」—을 앵커 없는 grep 으로 판정하면 **`subagent_type` 이 `agent_type` 도착으로 읽히고**, 그것은 이 카드를 **「도달 가능」이라는 결정적 오답으로 닫는다.** 리드가 실제로 이 오독을 겪고 스스로 정정했다(전달 메시지 기재 — 이 감사가 그 경위를 재측정하지는 않았으므로 **전달된 자기보고 등급 그대로** 옮긴다). D2 의 실패 방향(도달 가능한 값을 못 봄)과 **정반대이면서 같은 뿌리**다.

### 이 발견이 판정에 미치는 영향

**Verdict(FAIL)와 Overall(0.8125)은 변하지 않는다.** D8 은 Testability 0.75 밴드를 **보강**할 뿐 밴드를 옮기지 않는다 — AC 들은 여전히 이진 판정 가능하며, 결함은 판정 기준이 아니라 **측정 레시피의 미지정**에 있다. 점수를 내리는 것이 더 엄격해 보이지만, 그것은 발견 하나로 밴드를 조작하는 것이므로 하지 않는다.

---

## 4-bis. 추가 결함

- **D8** — `acceptance.md:39-47` (AC-BGX-002), 파생으로 `acceptance.md:124-132` (AC-BGX-009) — 결정적 AC 가 **측정 패턴을 지정하지 않아** 부분문자열 과다매칭에 무방비다. 이 감사 실측: snake 축 앵커 없는 `agent_type` 48 = 실제 28 + `subagent_type` 20(71% 과다). camel 축은 과다매칭 토큰이 **셋**(`subagentType` 2 · `SubagentType` 3 · `TestAgentSpawnFixtureCarriesSubagentType` 2)이라 하나만 빼는 교정도 불완전. 실현 사례: `internal/hook/testdata/agent_pretool_payload.json` 을 앵커 없이 재면 1적중이 나오는데 그것은 `tool_input` 안의 `subagent_type` 이다(앵커 재측정 0적중, 최상위 키 집합에 정체성 필드 전무, 양성 대조 `hook_event_name`·`session_id` 실값 확인). 과다계상은 **큰 수**를 내고 큰 수는 안전해 보이므로 이 오류는 **스스로 신호를 내지 않는다.** — **Severity: major · Class: blocking** — **필요 수정**: AC-BGX-002 에 ① 철자 판정을 **완전일치 키 비교**(최상위 JSON 키 집합)로 수행하고 ② 부득이 grep 을 쓸 때는 경계 앵커(`(^|[^a-zA-Z_])`)를 의무화하며 ③ 앵커/비앵커 두 수를 **나란히** 기록해 차이가 0 이 아니면 과다매칭을 보고하도록 조항 추가. AC-BGX-009 의 재열거에도 같은 앵커 의무를 건다. **이 수정은 D2(키 집합 전량 기록)와 같은 한 줄로 동시에 닫힌다.**

## 5-bis. 권고 갱신

§5 의 수정 지시에 다음을 추가한다:

6. **D8** — AC-BGX-002 의 철자 판정을 **키 집합 완전일치**로 규정하고, grep 사용 시 경계 앵커 의무 + 앵커/비앵커 병기 조항 추가. AC-BGX-009 재열거에도 동일 적용. (D2 의 「최상위 키 집합 전량 기록」과 **한 조항으로 합쳐 처리 가능** — 키 비교는 부분문자열 과다매칭이 구조적으로 불가능하다.)

**추가로 확인된 SPEC 의 강점**: `progress.md` §E.1 (2) 가 철자 선언을 `^\s*AgentType` 로 **앵커해서** 쟀다 — 이 SPEC 에서 유일하게 앵커된 철자 레시피이고 정확하다. 문제는 그 규율이 **AC 로 승격되지 않아 run 단계를 구속하지 못한다**는 점뿐이다. 또한 `progress.md` Gaps 의 「`agent_pretool_payload.json` 은 메인 세션의 `Agent` 도구 호출 캡처라 이 질문에 답하지 않으며, 그 페이로드의 정체성 필드 부재를 부재 근거로 쓸 수 없다」는 **이 감사의 키 집합 측정으로 옳음이 확인됐다** — manager-spec 의 판단이 맞았다.

---
---

# plan 감사 판정서 — iter-2 (재감사)

> **위 §1~§9 는 iter-1 기록이며 고쳐 쓰지 않는다.** 이 절부터가 iter-2 다.

- 감사자: plan-auditor · 반복 **iter-2 / 3**
- 감사 대상 커밋: **`3d636c991`** (`fix(SPEC-…): close iter-1 blocking defects D2/D3/D4 and declare tier`)
- iter-1 기준선 커밋 `039af6915` 는 **온전**하다(`git cat-file -t 039af6915` → `commit`)
- 측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064` · 브랜치 `WT-branchguard-exempt`
- **Verdict: PASS-WITH-DEBT** · Overall **0.9375** · Tier M 임계 **0.80** → 점수상 통과
- 범위: iter-1 이 열거한 결함 델타 + 그 수정이 만든 새 표면. 전면 재감사 아님.

## iter-2 §0 — 감사 시작 시 관측된 트리 불일치 (먼저 보고한다)

**Claim.** 배차문은 「working tree carries only `?? .moai/reports/t1064/`」로 확인했다고 밝혔으나, **이 감사가 잰 워킹 트리는 그것과 다르다.** 추적 파일 `acceptance.md` 에 **미커밋 수정 1건**이 있다.

**Evidence.**

```
$ git status --porcelain
 M .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/acceptance.md
?? .moai/reports/t1064/

$ git diff --stat -- .moai/specs/
 .../acceptance.md | 4 +++-
 1 file changed, 3 insertions(+), 1 deletion(-)
```

내용은 AC-BGX-001 **강화**다 — 키 집합을 「키 단위로 분리된 형태(JSON 키 목록 또는 한 줄 한 키)」로 남기라는 요건과 그 [HARD] 근거(「원문 JSON 한 덩어리로만 남기면 비교가 다시 문자열 검색으로 퇴화한다」) 추가.

**Baseline-attribution.** 트리 `t1064`, HEAD `3d636c991`, 이번 실행.

**판정과 처분.**
- **감사 기준은 커밋된 `3d636c991` 상태로 잡았다.** 어느 브랜치에도 커밋되지 않은 워킹 사본은 정본으로 인용할 수 없다(인용 일반 규칙) — 다른 사람이 같은 것을 읽었는지 확인할 방법이 없다.
- 따라서 **이 미커밋 델타는 어떤 결함의 닫힘 근거로도 쓰지 않았다.** 다행히 D2/D8 닫힘은 커밋된 AC-BGX-001 의 「최상위 키 집합 전량」 조항만으로 성립하므로 판정이 델타에 의존하지 않는다.
- **이것은 정리 대상이 아니라 보고 대상이다.** 감사 중인 워크트리는 쓰기 주체가 하나여야 하고, 예기치 않은 변경 관측은 조용히 넘기지 않고 리드에게 보고한다. 되돌리지도, 커밋하지도 않았다.
- → **D9(process)**.

**Residual-risk.** 배차문의 확인 시점 이후에 편집이 착지했다면 감사 창 안에서 다른 쓰기 주체가 움직인 것이고, 확인 자체가 이 파일을 못 봤다면 확인 절차가 추적 파일을 놓친 것이다. **어느 쪽인지 이 감사는 재지 않았다** — 판별에는 편집 시각과 세션 귀속이 필요한데 둘 다 관측하지 않았다.

---

## iter-2 §1 — Must-Pass 재판정

| 항목 | 판정 | 근거 |
|---|---|---|
| MP-1 REQ 번호 정합 | **PASS** | `REQ-BGX-001`~`010` 연속 10건, 앵커 카운트 `grep -cE '^\- \*\*REQ-BGX-[0-9]{3}\*\*'` → **10**, 중복 0, 자릿수 균일 |
| MP-2 GEARS (요건층) | **PASS** | 10건 전부 `shall`/`shall not` + 응답 명시. iter-1 의 유일한 약점(REQ-BGX-002 주어 생략)이 닫혔다(D7). REQ-BGX-010 은 「…측정될 때, 시스템은 …해야 한다(shall)」 사건구동형. **판정 층: `spec.md` 의 `REQ-XXX` 요건층**이며 `AC-XXX` 검증층은 대상이 아니다 |
| MP-3 프론트매터 | **PASS** | 정본 12필드 전부 존재 + optional `tier: M`(`spec.md:13`). 거부 별칭 0건. `version: "0.3.0"` 로 갱신 |
| MP-4 언어 중립성 | **N/A (auto-pass)** | 단일 프로그래밍 언어(Go) 범위 |
| MP-5 D7 교차 SPEC | **PASS** | 변동 없음 — 형제 4건 전부 `completed` |
| MP-6 D8 크로스플랫폼 | **PASS (auto)** | `syscall` 0건 (양성 대조: 같은 경로에서 `HARD` 가 plan.md 6 / acceptance.md 18 적중 — grep 정상) |
| MP-7 해명 게이트 | **PASS** | `[NEEDS CLARIFICATION` 0건, 같은 양성 대조 발화 |

---

## iter-2 §2 — 차원 점수

| 차원 | iter-1 | iter-2 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | **1.00** | D7(주어) 닫힘 · AC-BGX-007b 의 「사칭 형태」가 두 조건으로 정의됨 · AC-BGX-003 이 세 칸으로 분화. 인용 가능한 모호성이 남지 않았다 |
| Completeness | 0.75 | **0.75** | tier 선언·`agent_id` 축·집합 하한 10·§A.4 실행 가능성 근거가 모두 들어왔다. **그러나 §A.2 의 저장소 전체 수치와 그 내역이 배포 커밋에서 틀리다**(D10) |
| Testability | 0.75 | **1.00** | 결정적 판정의 **측정 방법이 고정**됐고(키 집합 동등 비교), 공허 통과가 AC-007b·AC-011 양쪽에서 봉쇄됐으며, AC-009 는 「앵커를 짐작으로 달지 말고 접두 열거로 과다매칭원을 먼저 드러내라」까지 규정한다 |
| Traceability | 1.00 | **1.00** | REQ-BGX-001~010 전부 매트릭스에 등장, AC 12건 전부 유효 REQ 또는 §C/§D 지목, `AC-BGX-011 → REQ-BGX-010` 신규 행 포함 |

**Overall = (1.00 + 0.75 + 1.00 + 1.00) / 4 = 0.9375** · Tier M 임계 0.80 → **점수상 통과**

---

## iter-2 §3 — 결함별 CLOSED / OPEN 판정

각 항목은 「보고받은 대로」가 아니라 **이 트리에서 이번 실행에 재측정**해 판정했다.

### D1 (tier) — **CLOSED**

`spec.md:13` 에 `tier: M`. Tier 판단의 독립 검증은 §4 에 따로 둔다.

### D2 + D8 (키 집합 병합 조항) — **CLOSED**

배차문의 물음: 「키 집합 동등 비교가 실제로 결정 AC 의 지정된 측정 방법인가, 아니면 앵커 없는 grep 으로도 여전히 만족되는가.」

**답: 지정돼 있고, 앵커 없는 grep 으로는 만족되지 않는다.** 세 층이 겹쳐 있다.

1. **§A.2 [HARD] 측정 방법 조항** — 「정체성 필드의 도착 여부(§B REQ-BGX-001)는 **문자열 검색으로 판정하지 않는다.** 캡처한 페이로드의 **최상위 키 집합**과 프로브 이름을 **정확 일치**로 대조한다.」 그리고 왜 「앵커를 달면 된다」가 불충분한지를 **측정으로** 논증한다(camel 축 과다매칭원 셋 중 둘이 대문자 시작).
2. **AC-BGX-001** — 키 집합 전량 기록이 **의무**이고, 「네 철자가 없다」만 적으면 **FAIL**, 키 집합이 없으면 이 실행의 모든 부재 관측이 **무효**가 되어 AC-BGX-005 로 내려간다.
3. **AC-BGX-002 [HARD] 경계 앵커 의무** — grep 을 쓸 경우 JSON 키 경계(`"agent_type"`) 또는 단어 경계 앵커를 강제하고 **앵커 없는 수치 인용을 금지**하며, 「재측정해도 같은 48 이 나와 검산으로 잡히지 않는다」는 이유까지 적혀 있다.

허용된 두 앵커 형태가 실제로 충분한지 **경험적으로 확인**했다(`internal/hook/testdata/agent_pretool_payload.json`):

```
$ grep -c '"agent_type"' <payload>        → 0   (JSON 키 경계 형태)
$ grep -cw 'agent_type'  <payload>        → 0   (단어 경계 형태)
$ grep -c  '"session_id"' <payload>       → 1   ← 양성 대조: 계측기 발화
$ grep -c  'subagent_type' <payload>      → 1   ← 대조: 과다매칭원이 실재
```

앵커 없는 형태가 1을 내는 바로 그 파일에서 **허용된 두 형태 모두 0**이고, 양성 대조 두 개가 발화한다. 즉 AC 가 허용하는 측정 형태는 이 위험에 대해 **충분**하다.

프로브 집합도 넷으로 넓어졌고(`agent_type`·`agentType`·`agent_id`·`agentId`), AC-BGX-003 이 「넷 다 부재」/「snake·camel 도착」/「`agent_id` 계열만 도착」 **세 칸**으로 분화했으며, 세 번째 칸이 배차문이 말한 **「결론은 맞고 기제는 틀림」**에 정확히 대응한다. AC-BGX-004 는 판정 근거를 **키 집합으로 못박고** 「그것 없이 이 AC 를 닫아서는 안 된다」를 추가했다 — iter-1 이 지적한 「(3) 분기에서 판정 불능」이 닫혔다.

### D3 (변주 축 + 집합 하한) — **부분 CLOSED, 수치는 OPEN(D10)**

**닫힌 부분**: 변주 축이 패턴 → 탐색 뿌리로 재작성됐고(AC-BGX-009 (1) + [HARD]), 정정 제외 목록이 명시됐으며, `hooks-system.md` 쌍이 **적중이 아니라 구성으로** 집합에 들어가 하한 10 이 선언됐다. 쌍의 0/0 + 양성 대조를 재현했다:

```
$ grep -c "tool-spawned subagent" .claude/…/hooks-system.md …/templates/…/hooks-system.md
.claude/rules/moai/core/hooks-system.md:0
internal/template/templates/.claude/rules/moai/core/hooks-system.md:0
$ grep -c 'All hook events include' <같은 두 파일>          ← 양성 대조
.claude/rules/moai/core/hooks-system.md:1
internal/template/templates/.claude/rules/moai/core/hooks-system.md:1
```

0/0 이 부재이고 파일 판독 실패가 아님이 확정된다. **구성상 포함 논거는 타당하다** — 판정이 어느 방향이든 모순하는 두 문장 중 하나는 반드시 바뀐다.

**열린 부분**: 저장소 전체 수치 → **D10**(아래 §5).

### D4 (사칭 형태 + CODE 불가 분기) — **CLOSED**

배차문의 물음: 「007b 가 기존 테스트로 더는 만족될 수 없는가.」 **답: 구조적으로 불가능하다.**

AC-BGX-007b 의 [HARD] 구성 요건은 **두 조건 동시 만족**을 요구한다: ① `AgentType == "manager-git"` (**제로값이 아니다**) ② 판별 필드가 서브에이전트임을 가리킴. 그런데 기존 테스트는 `AgentType` 을 **제로값으로 둔다** — iter-1 에서 원문을 읽었다(`branch_guard_flagclass_test.go:254-259`, 주석 `// AgentType deliberately left zero-valued: subagent-shaped payload.`). 조건 ①과 **정면으로 배타적**이므로 그 테스트로는 007b 를 만족시킬 수 없다.

덧붙여 AC-BGX-007b 가 그 테스트를 **이름과 좌표로 지목해 배제 기준선으로 고정**하고(「그 형태의 deny 는 이미 존재하는 테스트가 고정하고 있으므로 … 아무것도 새로 고정하지 않은 채 통과한다」), 새 검증이 **다른 입력**을 쓴다는 명시를 판정서에 요구한다.

CODE 불가 분기는 REQ-BGX-010 + AC-BGX-011 로 신설됐고, **실패 귀속을 뒤집는 조항**까지 있다 — 판별자가 부재인데 「CODE 로 갔고 007b 를 만족시켰다」는 보고가 오면 **007b 가 아니라 AC-BGX-011 이 FAIL**이다. 「고르지 않았다」와 「구성할 수 없었다」의 구별이 AC 문언에 들어가 있다.

### D5 (Go 층 동반이동) — **CLOSED**

AC-BGX-009 **(3)** 이 생산지(`branch_guard.go:604-606`)와 고정 테스트(`branch_guard_test.go:694`)를 **둘 다** 좌표로 묶고, `strings.Contains(reason, "tool-spawned subagents")` 를 인용하며 「문서만 고치고 이 테스트를 두면 즉시 깨진다」를 적는다.

### D6 (계측기 되돌림) — **CLOSED**

AC-BGX-010 **(3)**: `git status --porcelain -- .claude/hooks/` **0행**. [HARD] 근거로 「되돌리지 않으면 다음 `moai update` 가 말없이 덮어 계측기가 사라진 사실이 어디에도 안 나타난다」까지 적혀 있다 — iter-1 이 지적한 것보다 한 단계 깊다.

### D7 (REQ-002 주어) — **CLOSED**

```
- **REQ-BGX-002** … 억제할 수 있는지를, 시스템은 코드 판독이 아니라 **경로 실행**으로 판정해야 한다(shall).
```

---

## iter-2 §4 — Tier 독립 판정 (배차문 [HARD])

**판정: `tier: M` 은 정직하다.** 임계를 낮추려 고른 편의적 Tier 가 아니다.

| 지표 | 측정값 | Tier 함의 |
|---|---|---|
| 아티팩트 집합 | spec + plan + acceptance = **3** | **M 과 정확히 일치**(S=2, L=5) |
| REQ 수 | **10** (앵커 카운트) | S 상한 8 **초과** → S 배제 · M 상한 16 이내 |
| AC 수 | **12** (앵커 카운트) | S 상한 8 **초과** → S 배제 · M 상한 16 이내 |
| 동반이동 파일 | **10** (하한) | M 대역 5-15 **적중** |
| LOC | 조사 카드 — 코드 변경 조건부·소규모 | **S 방향**(반대 지표) |

**근거의 질.** 서로 독립인 두 지표(파일 수 · REQ/AC 예산)가 모두 M 을 가리키고, 아티팩트 집합이 M 과 일치한다. 결정적으로 **반대 지표(LOC → S)를 숨기지 않고 명시**했다 — 편의적 Tier 선택의 전형은 유리한 지표만 열거하는 것이므로, 반대 지표의 자진 기록은 반대 증거다.

**L 이 강제되는가?** 아니다. L 요건은 「>1000 LOC 또는 constitutional」인데, 이 카드가 만지는 `main-checkout-branch-guard.md` 는 운영 규칙이지 `moai-constitution.md` 가 아니며, 산출물도 판정서다.

**S 로 내려갈 수 있는가?** 아니다 — REQ 10·AC 12 가 S 예산 8/8 을 **둘 다** 초과한다. 예산 초과는 「Tier 를 올리라는 신호이지 예산을 완화하라는 신호가 아니다」(spec-workflow.md).

**임계 효과.** M 선언으로 임계가 0.85 → 0.80 이 된다. 다만 **이 판정은 그 완화에 기대지 않는다**: 점수 0.9375 는 iter-1 임계 0.85 로도 통과한다. 즉 Tier 선언이 없었어도 결론은 같다 — 편의적 Tier 로 임계를 넘은 사례가 아니다.

---

## iter-2 §5 — 20 대 21 재조정 (배차문이 직접 요청)

**Claim.** **양쪽이 맞는 것이 아니다.** 배포 커밋 `3d636c991` 에서 저장소 전체 적중은 **20** 이며 **21 이 아니다**. 그리고 그 차이는 판정서 착지(+1) **하나가 아니라**, 같은 커밋이 `plan.md` 에서 리터럴을 **제거한 것(−1)** 과 합쳐진 결과다. manager-spec 의 재조정 서사는 **+1 만 설명하고 −1 을 가린다.**

**Evidence.**

(1) 지금 수치 — 20:

```
$ grep -rln --exclude-dir=.git "tool-spawned subagent" . | wc -l
20
$ grep -rln "tool-spawned subagent" .claude internal | wc -l
8
```

(2) 이 SPEC 자신의 아티팩트는 **4 가 아니라 3**(§A.2 내역은 4 라고 적는다):

```
$ grep -rln "tool-spawned subagent" .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/
.moai/specs/…/acceptance.md
.moai/specs/…/progress.md
.moai/specs/…/spec.md          ← 3건. plan.md 없음
$ grep -rlc "SPEC-BRANCHGUARD" .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/ | wc -l
4                               ← 양성 대조: 4개 아티팩트 모두 판독 가능
```

양성 대조가 4를 내므로 `plan.md` 의 0 은 **실제 부재**다.

(3) `plan.md` 는 iter-1 시점에 리터럴을 **1회 싣고 있었고**, 그것을 없앤 것은 **이 커밋 자신**이다:

```
$ git show 039af6915:.moai/specs/…/plan.md | grep -c "tool-spawned subagent"
1
$ grep -c "tool-spawned subagent" .moai/specs/…/plan.md   → 0   (exit 1)
$ grep -c "subagent" .moai/specs/…/plan.md                → 2   ← 양성 대조

$ git show 3d636c991 -- .moai/specs/…/plan.md | grep -n "tool-spawned subagent"
135:-1. `tool-spawned subagent` 리터럴로 재열거(plan 시점 하한 8파일).
```

선두 `-` — M6 단계 1을 재작성하면서 **삭제**됐다.

(4) 따라서 정확한 재조정은:

| 좌표 | 적중 |
|---|---|
| iter-1 측정 (HEAD `039af6915`, 판정서 착지 전) | **20** |
| iter-2 배포 커밋 (HEAD `3d636c991`, 판정서 착지 후) | **20** = 20 **−1**(plan.md 제거) **+1**(verdict.md 추가) |
| manager-spec 이 적은 값 | **21** — 두 좌표 어디에서도 재현되지 않음 |

21 이 나오려면 `plan.md` 가 아직 리터럴을 싣고 **있으면서** `verdict.md` 가 이미 존재해야 한다. 그것은 **어떤 커밋에도 대응하지 않는 편집 중간 상태**다.

**Baseline-attribution.** 트리 `t1064`, HEAD `3d636c991`, 이번 실행. (3)의 `039af6915` 판독은 같은 실행에서 `git show` 로 수행했다.

**Gaps.** 21 이 정확히 언제 측정됐는지는 재지 않았다 — 편집 시각을 관측하지 않았다. 「편집 중간 상태」는 산술이 지지하는 **가장 단순한 설명**이지 관측된 사실이 아니다.

**Residual-risk.** 이 오류의 형태가 나쁘다. 서사(「감사의 20 도 지금의 21 도 틀리지 않았다 — 같은 뿌리를 서로 다른 시점에 잰 것이다」)가 **그럴듯하고 교훈적이라서** 검산을 유발하지 않는다. 게다가 그 문장은 §A.2 에서 **「뿌리와 시점을 함께 적어야 하는 이유의 실례」로 인용된다** — 즉 방법론의 모범 사례로 제시된 수치가 배포 커밋에서 틀렸고, 틀리게 만든 것이 그 커밋 자신이다. 실질 피해는 작지만(집합 **멤버십**은 정확하고 처방도 안 바뀐다), **하한 집합의 수치가 다음 사람의 재열거 기준선이 된다**는 점에서 방치하면 안 된다.

---

## iter-2 §6 — manager-spec 이 자진 신고한 두 항목에 대한 판단

**(가) 「28 은 뺄셈 파생값」** — **적절히 처리됐다.** `progress.md` 가 그 Gap 을 `~~취소선~~` 으로 표시하고 「iter-3 에서 닫힘: 앵커로 직접 재어 28 을 얻었다」로 갱신하되 **iter-2 기록을 고쳐 쓰지 않았다**. 범위 차이(전체 파일 vs `*.go`/`*.json`)도 스스로 적고 「두 수를 하나의 측정으로 합쳐 인용하지 않는다」고 보수적으로 닫았다.

이 감사가 그 보수성을 검증했다 — 범위를 **넓혀서** 재측정:

```
$ grep -rno 'agent_type'                      internal/hook/ | wc -l  → 48
$ grep -rno 'subagent_type'                   internal/hook/ | wc -l  → 20
$ grep -rnoE '(^|[^a-z])agent_type'           internal/hook/ | wc -l  → 28
```

**범위 제한 없이도 48/20/28 로 동일**하다. 즉 `internal/hook` 안에 `*.go`/`*.json` 밖에서 이 토큰을 싣는 파일이 없다. 그들의 「범위가 달라 합쳐 인용하지 않는다」는 주의는 **과했지만 틀리지 않았고**, 과한 방향이 안전한 쪽이다. 결함으로 잡지 않는다.

**(나) 「동반이동 집합 10 은 파일 입도 하한」** — **정직하고 정확하다.** 문장 단위 판독은 이 감사도 하지 않았고(iter-1 Gaps 에 같은 내용을 적었다), 「하한」이라는 등급이 문장 안에 실려 있다. 결함 아님.

---

## iter-2 §7 — 미측정으로 남은 것 (등급 보존)

manager-spec 의 자기 신고와 이 감사의 측정이 일치한다:

- **도달성 자체 — 미측정.** 리드가 넘긴 **PLAUSIBLE 등급이 유지**되며, 이 SPEC 도 이 감사도 올리지 않았다.
- **`agent_id` 의 실제 도착 — 미측정.** 선언(`types.go:239`)은 도착의 증거가 아니다. **그것이 AC-BGX-011 이 존재하는 이유**이며, 이 인과가 SPEC 문언에 적혀 있다.
- **계측기 실제 삽입 — 미수행.** §A.4 의 등급은 **measured-feasible**(구성요소 실재로 확립)이지 **executed-through 가 아니다.** 이 감사 역시 래퍼를 편집하지 않았다.
- **키 집합 동등 비교 — 한 번도 수행되지 않음.** 비교 입력이 아직 없다. **방법을 고정했을 뿐**이라는 것이 SPEC 자신의 Gaps 에 적혀 있다.
- **`go test ./internal/hook/...` 기준선 — plan 단계 미실행.** run 단계 §C 사전 점검 항목. (참고: **iter-1 감사가 실행해 exit 0 / 11개 패키지 전부 `ok` 를 이미 기록**했다 — 위 iter-1 §7 참조. 그것은 감사의 측정이지 SPEC 의 것이 아니므로 SPEC 의 Gap 은 그대로 유효하다.)
- **`MOAI_BRANCH_GUARD_EXEMPT` 환경변수 축 — carried-forward-unverified.** 재측정되지 않았다.
- **`internal/hook/pre_tool.go` `frozenZonePrefixes` — PLAUSIBLE 등급 그대로.**

---

## iter-2 §8 — 남은 결함

- **D9** — 워킹 트리 — 추적 파일 `.moai/specs/…/acceptance.md` 에 **미커밋 수정 1건**이 감사 창 안에서 관측됐고, 배차문의 확인 내용(「`?? .moai/reports/t1064/` 뿐」)과 **불일치**한다. 내용은 AC-BGX-001 강화(키 단위 분리 형태 요건)로 **개선**이지만, 커밋되지 않았으므로 **어떤 닫힘의 근거로도 쓰지 않았다.** 감사 중인 워크트리의 단일 쓰기 주체 원칙에 걸리는 관측이라 조용히 넘기지 않는다. — **Severity: major · Class: blocking(process)** — **필요 조치**: 리드가 처분을 정한다(커밋해 iter-3 범위에 넣거나, 의도치 않은 것이면 귀속을 확인). **감사가 되돌리거나 커밋하지 않았다.**

- **D10** — `spec.md` §A.2 「변주 축」 표 및 그 내역 문단, 동일 주장이 `progress.md` §E.1 (8) — 저장소 전체 적중을 **21** 로, 내역을 「이 SPEC 자신의 아티팩트 **4**」로 적었으나 배포 커밋 `3d636c991` 에서 실측은 **20 / 3** 이다. 차이는 판정서 착지(+1) 하나가 아니라 **같은 커밋이 `plan.md` M6 에서 리터럴을 제거한 것(−1)** 과 합쳐진 결과이며, 재조정 서사는 +1 만 설명해 −1 을 가린다. 21 은 **어떤 커밋에도 대응하지 않는 편집 중간 상태**의 값이다. 집합 **멤버십**과 처방은 영향받지 않으나, 이 수치는 §A.2 에서 **방법론의 모범 사례로 인용**되고 다음 사람의 재열거 기준선이 된다. — **Severity: major · Class: blocking** — **필요 수정**: §A.2 표의 「저장소 전체 21」을 **20** 으로, 내역을 「이 SPEC 아티팩트 **3**(`plan.md` 은 이 커밋에서 리터럴이 제거됨) + RC-TESTBED 4 + FLAGCLASS 3 + CHANGELOG 1 + 판정서 1 = 12, 8+12=20」으로 정정. 재조정 문장은 **제거(−1)를 명시**하도록 고쳐 쓴다 — 「+1 만 있는 것이 아니라 −1 도 있었고 우연히 상쇄되지 않았다」가 요점이다. `progress.md` §E.1 (8)은 **관측 기록이므로 고쳐 쓰지 않고**, iter-3 항목으로 **정정 추가**한다.

---

## iter-2 §9 — 판정 근거와 권고

**PASS-WITH-DEBT** 인 이유:

- Must-Pass **7/7 PASS**(MP-4 N/A). 필수 통과 실패 0.
- Overall **0.9375** ≥ Tier M 임계 **0.80**, 그리고 **iter-1 임계 0.85 로도 통과** — Tier 선언에 기대지 않은 점수다.
- iter-1 의 차단 결함 **D1·D2·D3(축)·D4·D8 전부, 그리고 optional D5·D6·D7 까지 8건이 실측으로 닫혔다.** 문언만 응답한 것이 아니라 **기제가 바뀌었다** — 특히 D4 의 배제 기준선과 AC-BGX-011 의 실패 귀속 반전, D8 의 「앵커 대신 키 집합 동등 비교」는 iter-1 이 요구한 것보다 강하다.
- 남은 둘(D9·D10)은 **처방·집합 멤버십·AC 어느 것도 바꾸지 않으며**, D10 은 수치 정정 한 곳 + 서사 한 문장, D9 는 리드 판정 사항이다.

**FAIL 로 내리지 않은 이유를 명시한다.** D10 은 관측 없는 주장 계열이고 가볍게 볼 것이 아니지만, 8건을 실측으로 닫은 개정에 수치 한 건으로 FAIL 을 매기는 것은 **선택적 결함으로 FAIL 을 제조하는 것**이다. 대신 **debt 로 명명해 run 단계 진입 전 정정을 요구**한다 — 그 편이 결함을 더 확실히 닫는다.

**iter-3(있다면) 범위**: D9 처분 + D10 정정 **둘뿐**. 다른 항목은 재감사하지 않는다.

**절대 되돌리지 말 것** — 이번 개정에서 가장 잘 된 것:
- §A.2 의 **「앵커를 짐작으로 먼저 달지 마라 — 접두 열거로 과다매칭원을 먼저 드러내고 합이 닫히는지 보여라」**. 이것은 iter-1 이 요구한 것보다 한 수 위다.
- AC-BGX-011 의 **실패 귀속 반전**(공허 통과 보고가 오면 007b 가 아니라 011 이 FAIL).
- AC-BGX-010 (3)의 근거 — 되돌리지 않은 계측기를 **`moai update` 가 말없이 덮어 흔적이 사라진다**는 2차 효과까지 짚었다.
- `progress.md` 의 **취소선 + 「iter-N 에서 닫힘」** 방식 — 관측 기록을 고쳐 쓰지 않으면서 갱신을 표시한다.

---

## iter-2 §10 — 이 판정 자신의 Gaps / Residual-risk

**Gaps.**
- 감사 기준을 **커밋 `3d636c991`** 로 잡았고, 미커밋 `acceptance.md` 델타는 판정 근거에서 **제외**했다(D9 로 보고).
- `.moai/specs/` 아래 어떤 파일도 편집하지 않았다.
- 21 의 측정 시각은 재지 않았다 — 「편집 중간 상태」는 산술이 지지하는 가장 단순한 설명이지 관측이 아니다.
- 20파일 각각의 **문장 단위** 판독은 하지 않았다(파일 입도).
- 도달성·`agent_id` 도착·계측기 삽입은 이 감사도 재지 않았다 — SPEC 의 Gaps 와 동일하다.
- `go test ./internal/hook/...` 는 iter-2 에서 **재실행하지 않았다**. iter-1 에서 exit 0 를 얻었고, 이번 커밋은 `.moai/specs/` 문서 4개만 바꾸므로 Go 결과가 달라질 경로가 없다 — **다만 이것은 추론이지 재측정이 아니다.**

**Residual-risk.**
- D10 을 「수치 오타」로 축소해 서사(−1 은닉)를 고치지 않으면, 다음 사람이 §A.2 의 21 을 재열거 기준선으로 집어 **제거된 지점을 눈치채지 못한다.**
- D9 의 귀속이 확인되지 않으면, 감사 창 안에 다른 쓰기 주체가 있었는지가 미해결로 남는다.
- 이 카드가 **세 번째로 「미측정」으로 닫힐 위험**은 여전히 가장 크다. 계측 설계는 이제 충분히 단단하므로, 남은 것은 **실행 여부**다.

---

## iter-2 §11 — 감사 종료 시 부록: HEAD 가 감사 창 안에서 움직였다 (귀속 정정)

**이 절은 위 §0 의 귀속 진술을 정정한다.** 판정과 점수는 바뀌지 않는다.

### Claim

감사 창 안에서 **HEAD 가 `3d636c991` → `8852cb921` 로 이동**했고, 시작 시점에 미커밋이던 편집이 그 커밋으로 착지했다. 따라서 **위 §0 의 「감사 기준은 커밋된 `3d636c991` 상태로 잡았다」는 진술은 틀렸다** — 실제로 판독한 것은 워킹 트리이고, 그 워킹 트리는 `spec.md` 에도 미커밋 v0.3.0 내용을 싣고 있었다(§0 시점의 `git status` 는 `acceptance.md` 만 보여 줬다). 모든 표제 수치를 **새 HEAD 에서 재측정해 재귀속**했으며 값은 전부 동일했다.

### Evidence

**(1) HEAD 이동**

```
$ git rev-parse --short HEAD          # 감사 시작 시
3d636c991
$ git rev-parse --short HEAD          # 감사 종료 시
8852cb921
$ git log --oneline -2
8852cb921 fix(SPEC-…): close D8 by fixing the measurement method, not the grep recipe
3d636c991 fix(SPEC-…): close iter-1 blocking defects D2/D3/D4 and declare tier
$ git status --porcelain              # 종료 시 — ` M acceptance.md` 가 사라졌다
?? .moai/reports/t1064/
```

**(2) 어느 조항이 어느 커밋에 있는가 — D2 와 D8 의 근거가 서로 다른 커밋이다**

```
$ git show 3d636c991:…/acceptance.md | grep -c '최상위 키 집합 전량'   → 2   ← D2 근거: 3d636c991 에 이미 있었다
$ git show 3d636c991:…/spec.md       | grep -c '키 집합 동등 비교'     → 0
$ grep -c '키 집합 동등 비교' …/spec.md                                → 2   ← D8 근거: 8852cb921 에서 들어왔다
```

즉 **D2(키 집합 기록 의무)는 `3d636c991`, D8(키 집합 동등 비교를 측정 방법으로 고정)은 `8852cb921`** 이 근거다. 둘 다 이제 커밋돼 있으므로 인용 가능하지만, §3 에서 둘을 하나의 커밋으로 귀속한 것은 부정확했다.

**(3) 표제 수치 전량 재측정 — HEAD `8852cb921`**

```
repo-wide "tool-spawned subagent" : 20
.claude + internal                : 8
이 SPEC 아티팩트                   : 3
plan.md                           : 0
REQ (앵커 카운트)                  : 10
AC  (앵커 카운트)                  : 12
tier:                              : M
```

**전부 §1~§5 에서 보고한 값과 동일하다.**

**(4) D10 은 새 HEAD 에서도 살아 있다**

```
$ git show 3d636c991:…/spec.md | grep -A1 '저장소 전체' | head -1
| 저장소 전체 (`.git` 제외) | 21 |
$ grep -A1 '저장소 전체' …/spec.md | head -1        # HEAD 8852cb921
| 저장소 전체 (`.git` 제외) | 21 |
```

21 은 두 커밋 모두에 있고 실측은 두 커밋 모두에서 20 이다 — **D10 은 정정되지 않았다.**

### Baseline-attribution

트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064`. **판정의 귀속 커밋은 `8852cb921`** 이다(§0 이 적은 `3d636c991` 이 아니다). (3)의 재측정은 그 HEAD 에서 이번 실행에 수행했다.

### Gaps

- **감사 창 중간의 어느 시점에 어떤 파일이 수정 상태였는지**는 연속 관측하지 않았다. `git status` 를 한 번 찍었고 그 스냅샷이 즉시 낡았다 — `spec.md` 의 미커밋 v0.3.0 내용을 그 스냅샷은 보여 주지 않았는데, 그것이 스냅샷 이후에 쓰인 것인지 스냅샷이 놓친 것인지 **재지 않았다**.
- 커밋 `8852cb921` 의 저자·세션 귀속은 확인하지 않았다.

### Residual-risk

**이것이 D9 의 실체이며, 처음 본 것보다 크다.** 처음에는 「미커밋 파일 1개」로 보였으나 실제로는 **감사가 도는 내내 같은 트리에 쓰기가 계속되고 있었다.** 그 결과:

- 내가 읽은 `spec.md` §A.2 는 **커밋되지 않은 텍스트였고**, 나는 그것을 커밋된 것으로 적었다. 결론적으로는 그 텍스트가 곧 커밋돼 무해했지만, **커밋되지 않았을 수도 있었다** — 그랬다면 이 판정서는 이력에 없는 텍스트를 근거로 PASS 를 준 것이 된다.
- 한 번 찍은 `git status` 를 감사 내내 유효한 것으로 취급한 것이 내 쪽 결함이다. **감사 창 안에서 트리 상태는 재확인 대상이지 기억 대상이 아니다.**

**판정 변동 없음** — (3)의 재측정이 모든 표제 수치를 새 HEAD 에서 재현했고, D2·D8 의 근거 조항이 둘 다 커밋된 상태로 확인됐다. **Verdict PASS-WITH-DEBT, Overall 0.9375 유지.** D9 의 필요 조치만 바뀐다: 「미커밋 파일 처분」이 아니라 **「감사 창 안에 다른 쓰기 주체가 있었던 사실의 귀속 확인」**이다.

---

## iter-2 §12 — 재기준화(`8852cb921`) · 세 점 시계열의 완전 귀속 · 이동 ref 분류 감사

**이 절이 iter-2 의 최종 판정이다.** §0 의 귀속 진술과 §3 의 D2/D8 인용 조항, §2 의 Testability 점수를 정정한다.

### §12.1 감사 좌표 — `8852cb921`

**감사 좌표는 `8852cb921` 이다.** 배차 시점의 `3d636c991` 이 아니다.

리드가 원인을 밝혔다 — **리드의 배차 오류**: D8 보정을 manager-spec 에 보낸 뒤 같은 트리에 iter-2 감사를 띄워, **감사 창 안에 쓰기 주체가 둘 있었다**. 독트린은 감사 창 동안 쓰기 주체를 하나로 제한한다. 이것은 각주가 아니라 **모든 수치의 귀속에 걸리는 사실**이므로 판정서 본문에 남긴다.

**읽기별 SHA 귀속 — 섞지 않는다:**

| 판독 | 취한 좌표 | 재취 여부 |
|---|---|---|
| iter-1 전량 (§1~§9) | `039af6915` | 해당 없음(그 좌표가 기준선) |
| iter-2 §1 Must-Pass, §2 점수, §3 결함별, §4 Tier, §5 수치 | 워킹 트리(당시 `3d636c991` HEAD + 미커밋 v0.3.0) | **§12.2 에서 `8852cb921` 로 전량 재측정** |
| 코드측 측정(48/20/28, camel 29 인구조사, 페이로드 키 집합, `isExemptAgent`, 테스트 본문) | `039af6915` 이후 불변 | **재취 불필요 — 근거는 §12.2 (3)** |
| 이 절 전량 | `8852cb921` | — |

### §12.2 `8852cb921` 에서의 재측정

**(1) 표제 수치 — 전부 동일**

```
repo-wide "tool-spawned subagent"  : 20   (verdict.md 포함, 이 SPEC 의 plan.md 불포함)
.claude + internal                 : 8
이 SPEC 아티팩트                    : 3
REQ / AC (앵커 카운트)              : 10 / 12
tier:                              : M
```

**(2) 코드측 측정이 HEAD 이동에 영향받지 않는 근거 — 두 커밋 모두 `internal/hook` 을 건드리지 않았다**

```
$ git show --name-only --format= 8852cb921
.moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/{acceptance,plan,progress,spec}.md
$ git show --name-only --format= 3d636c991
.moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/{acceptance,plan,progress,spec}.md
```

네 파일씩, 전부 `.moai/specs/` 아래. **`internal/hook` 파일 0건.**

> **[계측기 사고 기록]** 이 사실을 처음에 `git show --stat | grep -c 'internal/hook'` → `0` 으로 재려 했는데, **양성 대조(`grep -c 'specs'`)도 0 을 냈다.** `--stat` 이 경로를 `.../acceptance.md` 로 줄이기 때문이다 — 그 0 은 부재가 아니라 **미측정**이었다. `--name-only` 로 바꿔 재측정했다. 양성 대조가 없었으면 「건드리지 않았다」를 근거 없이 적을 뻔했다.

### §12.3 20 → 21 → 20 — 세 점 시계열은 **완전히 귀속된다**

리드의 관측: 「무언가가 집합에서 빠져나갔는데 그것이 무엇인지 모르겠고 추측하지 않겠다.」 **이 감사가 그것을 찾았다.**

**Claim.** 셋 다 옳고, 움직인 원인은 **둘**이다 — 추적 집합에서 **1건이 빠지고**(`plan.md`), 미추적 판정서가 **1건 들어왔다**. 두 사건의 **순서** 때문에 합이 20 → 21 → 20 이 된다.

**Evidence.**

(1) 추적 집합만 — `git grep`(미추적 미포함):

```
$ git grep -l 'tool-spawned subagent' 039af6915 | wc -l   → 20
$ git grep -l 'tool-spawned subagent' 3d636c991 | wc -l   → 19
$ git grep -l 'tool-spawned subagent' 8852cb921 | wc -l   → 19
```

(2) 빠져나간 1건의 정체 — 이 SPEC 의 `plan.md`:

```
$ git show 039af6915:…/plan.md | grep -c 'tool-spawned subagent'   → 1
$ git show 3d636c991:…/plan.md | grep -c 'tool-spawned subagent'   → 0
$ git show 8852cb921:…/plan.md | grep -c 'tool-spawned subagent'   → 0
$ git show 8852cb921:…/plan.md | grep -c 'subagent'                → 2   ← 양성 대조
```

양성 대조 2 가 발화하므로 0 은 **실제 부재**다. 제거는 **`3d636c991` 에서** 일어났다(M6 단계 1 재작성, iter-2 §5 (3)에 `-` 행 원문 인용).

(3) 완전 분해:

| 좌표 | 추적 | 미추적 `verdict.md` | 합 | 누가 읽었나 |
|---|---|---|---|---|
| `039af6915` | 20 | 없음 | **20** | iter-1 감사 |
| `3d636c991` **커밋 전** 워킹 트리 | 20 | 있음 | **21** | manager-spec |
| `3d636c991` | 19 | 있음 | **20** | iter-2 감사 · 리드 |
| `8852cb921` | 19 | 있음 | **20** | 이 절 |

**21 은 「`plan.md` 이 아직 리터럴을 싣고 있으면서 판정서가 이미 존재하는」 상태**이며, 그 상태는 **커밋 전 워킹 트리에만 존재한다.**

(4) **리드의 Test 2 가 귀속을 못 찾은 이유 — 프로브가 파일 하나였다.** 리드의 측정 자체는 정확하고 재현된다:

```
$ git show 3d636c991:…/spec.md | grep -c 'tool-spawned subagent'   → 3
$ git show 8852cb921:…/spec.md | grep -c 'tool-spawned subagent'   → 3
```

`spec.md` 는 실제로 불변이다. 그러나 **집합의 변화를 한 파일로 물은 것**이 문제다 — 빠져나간 것은 `plan.md` 였다. 「이 축은 배제됐다 → 귀속 없음」은 **한 파일의 불변을 집합의 불변으로 일반화**한 것이다.

**Baseline-attribution.** 트리 `t1064`, HEAD `8852cb921`, 이번 실행. (1)(2)의 과거 커밋 판독은 같은 실행에서 `git grep` / `git show` 로 수행했다.

**Gaps.** manager-spec 의 21 이 **정확히 언제** 측정됐는지는 여전히 재지 않았다(편집 시각 미관측). (3)의 둘째 행은 **산술이 유일하게 지지하는 상태**이지 관측된 타임스탬프가 아니다.

**§A.2 의 재조정은 세 번째 측정을 견디는가 — 아니다.** 현행 문안은 단일 원인(판정서 착지 +1)만 적고 **`plan.md` 제거(−1)를 적지 않는다.** 그래서 「지금은 21」이 되고, 실제로는 20 이다. 리드의 지적이 옳다.

### §12.4 이동 ref 술어 감사 — **분류는 유지, 근거 하나는 정정**

리드의 분류를 채택하지 않고 **네 테스트를 이 좌표에서 독립 실행**했다.

- **Test 1(치환)** — 후속 독자가 행동할 값으로 치환하면 이미 낡았다(21 → 실측 20). → **SUBJECT**. **리드와 일치.**
- **Test 2(반증 출처, 조건부)** — 리드는 「귀속 없음」으로 닫았으나, **이 감사는 귀속을 찾았다**(§12.3: `3d636c991` 의 `plan.md` 제거). **정정.** 다만 리드가 정확히 적었듯 **Test 2 는 등급이 아니라 귀속을 돌려주므로 나머지를 게이트하지 않는다** — 따라서 이 정정은 **분류를 바꾸지 않는다.**
- **Test 3(재측정 기대)** — 다음 주 이 카드에 아무 작업 없이 재실행하면 같은 답이 아니다. 다른 레인이 이 리터럴을 싣는 카드를 계속 착지시키고, AC-BGX-009 의 존재 이유 자체가 처방 시점 재열거다. **변동이 주장의 요점이다.** → **SUBJECT**. **리드와 일치.**
- **Test 4(독시 행동)** — AC-BGX-009 가 이 열거를 처방 시점 동반이동 집합으로 **소비**하므로 run 단계 행위자가 다시 잰다. 서술이 아니다. → **S2**.

**판정: SUBJECT / S2 → R4.** 리드의 분류가 **유지된다** — Test 1·3 이 일치하고 동률이 아니며, Test 4 가 S2 를 낸다. R1/R2(고정)는 이 주장이 존재하는 이유를 파괴하고, R3(면제)은 독자가 실제로 행동하므로 불가.

### §12.5 R4 의 대가가 실제로 치러지는가 — **명령이 네 요소를 다 가져야 한다**

리드의 요구(「명명된 명령은 실제로 개수를 정하는 명령이어야 하고, 뿌리와 패턴을 포함해야 한다」)에 **한 축을 더한다.**

이 트리에서 개수를 움직이는 축은 **넷**이며, 그중 **셋이 명령에 있고 하나는 번호에서 안 보인다**:

| 축 | 실측 |
|---|---|
| **뿌리** | `.claude`+`internal` **8** vs 저장소 전체 **20** |
| **패턴** | 앵커 없음/있음 — 이 리터럴에서는 0만큼 이동, 그러나 철자 축에서는 48 vs 28 |
| **시점(SHA)** | `039af6915` 20 · `3d636c991` 20 · 커밋 전 워킹 21 |
| **추적/미추적 범위** | `git grep` **19** vs `grep -r` **20** — **±1, 번호만 봐서는 보이지 않는다** |

넷째 축은 §A.2 에도 리드의 처방 초안에도 없다. `grep -r` 는 미추적 파일을 세고 `git grep` 은 세지 않으므로, **같은 뿌리·같은 패턴·같은 SHA 에서도 두 명령이 다른 수를 낸다.** 이 카드에서 그 차이는 정확히 판정서 자신이다.

**따라서 R4 가 값을 치르려면 명명된 명령이 ① 뿌리 ② 패턴 ③ 추적/미추적 여부를 모두 드러내야 하고, 값은 ④ SHA 가 붙은 날짜 있는 참조로 강등돼야 한다.** 이 넷 중 하나라도 빠지면 R4 는 **값싼 침묵기**로 쓰인 것이다.

### §12.6 AC-BGX-009 도 같은 처치가 필요한가 — **필요하다** (신규 결함 D11)

**측정한 결과, AC-BGX-009 는 넷 중 둘만 묶는다.**

| 요소 | AC-BGX-009 가 묶는가 | 근거 |
|---|---|---|
| 뿌리 | **예** | (1) 「두 뿌리 양쪽에서 재고 두 수를 모두 보고한다」 |
| 패턴 | **예** | [HARD] 접두 열거 — 「`[A-Za-z]*<토큰>` 으로 먼저 열거해 과다매칭원을 드러내고 합이 닫히는지 보인 뒤에 앵커를 정한다」 |
| **시점(SHA)** | **아니오** | AC-BGX-009 전문에 측정 시점·커밋을 기록하라는 의무가 **없다** |
| **추적/미추적 범위** | **아니오** | 언급 자체가 없다 |

게다가 (1)은 **plan 시점 값 `8` 을 기준선으로 본문에 박아 두고** [HARD] 로 「패턴만 흔들고 **「8 재확인」**으로 지나가면 FAIL」이라고 쓴다. 즉 **run 단계의 읽기를 plan 단계의 상수와 대조**하는 구조다 — 리드가 「한 층 아래의 같은 결함」이라 부른 것이 정확히 이것이다. 8 자체도 `.claude`+`internal` 의 추적 집합이므로 다른 레인이 움직일 수 있는 **이동 좌표**다.

→ **D11, blocking.**

### §12.7 `8852cb921` 이 바꿨다고 주장하는 것 — 검증

| 주장 | 판정 | 근거 |
|---|---|---|
| REQ-BGX-001 / AC-BGX-002 가 판정 방법을 **키 집합 정확 일치**로 고정, 문자열 검색·부분문자열 포함·정규식 매칭을 `shall not` 으로 배제 | **확인** | REQ-001 원문: 「…**최상위 키 집합과의 정확 일치 비교**로 내려야 하며(shall), 문자열 검색·부분문자열 포함·정규식 매칭을 판정 근거로 삼아서는 안 된다(shall not).」 AC-002 가 같은 것을 [HARD] 로 재천명 |
| 앵커 단 grep 도 그 배제에 포함 | **확인** | 앵커 grep 은 정규식 매칭이므로 `shall not` 에 걸린다 |
| AC-BGX-001 이 비교 입력이고, 키를 **분리된 항목**으로 고정 | **확인** | `acceptance.md:34` 「**키 단위로 분리된 형태**(JSON 키 목록 또는 한 줄 한 키)」 + `:36` [HARD] 근거 |
| 「한 덩어리로 남기면 AC-BGX-005 측정 불가」 경계 사례 | **확인** | `acceptance.md:215` 원문 존재 |
| AC-BGX-009 는 키 집합 비교를 **상속하지 않고** 접두 열거를 먼저 요구, 앵커 선짐작 금지 | **확인** | [HARD] 절 원문 확인 |
| `28` 뺄셈 파생 Gap 을 앵커 직접 측정으로 닫되 범위(`*.go`/`*.json`)를 명시해 48/20 과 합쳐 인용 금지 | **확인** | 취소선 + 「iter-3 에서 닫힘」, 범위 주의 병기 |

**§3 의 D2/D8 인용 정정.** iter-2 §3 은 **「AC-BGX-002 [HARD] 경계 앵커 의무」**를 인용해 「허용된 앵커 형태가 충분하다」고 적었다. 그 조항은 `3d636c991` 의 것이고 **`8852cb921` 에서는 대체됐다** — 이제 앵커 grep 자체가 판정 근거에서 배제된다. 따라서:

- 제가 실행한 앵커 충분성 실험(`'"agent_type"'` 0 / `-w` 0 / 양성 대조 1·1)은 **측정으로서 유효하나 AC 판정에는 무의미해졌다** — 더 강한 조항이 그것을 덮었다.
- **D2 와 D8 은 여전히 CLOSED 이며, 근거가 `8852cb921` 에서 더 강해졌다.**
- 근거 조항의 커밋 귀속: **D2**(키 집합 기록 의무, `최상위 키 집합 전량`) = `3d636c991` 에 이미 존재(실측 2 hits) · **D8**(키 집합 동등 비교를 방법으로 고정) = `8852cb921` 에서 도입(`3d636c991` 실측 0 hits).

### §12.8 점수 정정과 최종 판정

| 차원 | iter-2 §2 | **§12 최종** | 사유 |
|---|---|---|---|
| Clarity | 1.00 | **1.00** | 변동 없음 |
| Completeness | 0.75 | **0.75** | D10 — §A.2 재조정이 세 번째 측정을 못 견딘다 |
| Testability | 1.00 | **0.75** | **정정** — D11: AC-BGX-009 가 시점·추적범위를 안 묶고 plan 시점 상수 `8` 을 run 단계 대조 기준선으로 박아 둔다. 기준이 시점에 따라 다르게 판정되므로 Testability 결함이다 |
| Traceability | 1.00 | **1.00** | 변동 없음 |

**Overall = (1.00 + 0.75 + 0.75 + 1.00) / 4 = 0.875** · Tier M 임계 0.80 · iter-1 임계 0.85 — **양쪽 다 통과.**

**Verdict: PASS-WITH-DEBT** (유지). Must-Pass 7/7 은 `8852cb921` 에서 재확인됐다(§12.2 (1)).

### §12.9 남은 결함 (최종)

- **D9**(process, blocking) — 감사 창 안 쓰기 주체 둘. **원인이 밝혀졌다**: 리드의 배차 오류(D8 보정 발송 후 같은 트리에 감사 spawn). 필요 조치는 「미커밋 파일 처분」이 아니라 **재발 방지** — 감사 창 동안 해당 트리에 쓰기 금지를 배차문에 명시.
- **D10**(blocking) — §A.2 재조정이 **단일 원인**만 적어 세 번째 측정을 못 견딘다. **필요 수정(R4 형태로)**: 단일 원인 서사를 폐기하고 ① 개수를 정하는 명령을 **뿌리·패턴·추적범위까지 드러나게** 명명 ② 값은 **SHA 가 붙은 날짜 있는 참조**로 강등 ③ 세 점 시계열(§12.3 (3) 표)을 그대로 싣되 **두 원인(`plan.md` −1, `verdict.md` +1)과 그 순서**를 적는다. 「원인을 모른다」로 적어서는 안 된다 — 이 감사가 찾았다.
- **D11**(blocking, 신규) — AC-BGX-009 가 **시점(SHA)과 추적/미추적 범위를 묶지 않고**, plan 시점 상수 `8` 을 run 단계 대조 기준선으로 박아 둔다. **필요 수정**: (1)에 「재열거 시 **측정 SHA 와 명령 전문(뿌리·패턴·추적범위 포함)을 함께 기록**한다」를 추가하고, 「8 재확인으로 지나가면 FAIL」을 **「plan 시점 값과 **다를 수 있으며**, 다르면 그 차이를 두 원인 축(추적 집합 변화 / 미추적 산출물 도착)으로 분해해 보고한다」**로 바꾼다. 상수 대조를 요구하는 문언이 남으면 D10 이 한 층 아래에서 재생산된다.

**iter-3 범위: D9 재발 방지 문안 + D10 + D11. 셋뿐이다.**

### §12.10 이 절 자신의 Gaps / Residual-risk

**Gaps.**
- manager-spec 의 21 측정 시각 **미관측**. §12.3 (3)의 둘째 행은 산술이 유일하게 지지하는 상태이지 관측된 타임스탬프가 아니다.
- 20파일 각각의 **문장 단위** 판독은 여전히 하지 않았다(파일 입도).
- 커밋 `8852cb921` 의 세션 귀속은 확인하지 않았다 — 리드의 자기 신고를 그대로 옮겼다.
- `go test ./internal/hook/...` 는 재실행하지 않았다. **근거가 추론에서 측정으로 올라갔다** — §12.2 (2)가 두 커밋 모두 `internal/hook` 파일 0건임을 `--name-only` 로 보였다. 그래도 **이 좌표에서 테스트를 돌린 것은 아니다.**
- 도달성·`agent_id` 도착·계측기 삽입은 이 감사도 재지 않았다 — **PLAUSIBLE / 미측정 등급 그대로.**

**Residual-risk.**
- **D11 이 D10 보다 오래 산다.** D10 은 문서 한 곳의 수치이고 D11 은 **기준 자체**다 — 고치지 않으면 run 단계가 8 을 상수로 대조하다 정상적인 집합 변화를 결함으로 읽거나, 반대로 실제 누락을 정상 변동으로 넘긴다.
- **추적/미추적 축이 가장 조용하다.** 뿌리와 패턴은 명령을 보면 드러나지만 `grep -r` 와 `git grep` 의 차이는 **번호에 흔적을 안 남긴다.** 이 카드에서 그 차이가 정확히 판정서 자신이었다는 사실이 그 은밀함의 실례다.
- 이 카드가 **세 번째로 「미측정」으로 닫힐 위험**은 그대로다. 계측 설계는 이제 충분하고, 남은 것은 실행이다.

---

## iter-2 §13 — 판정 확정 직전 관측: 쓰기 동결 중 세 번째 쓰기

**Claim.** 리드가 「내 판정이 착지할 때까지 개정을 승인하지 않으며 쓰기 동결」이라고 선언했고 「manager-spec 은 파일을 건드리지 않았고 쓰기를 중단했다」고 전달했으나, **§12 를 판정서에 append 한 직후 `spec.md` 와 `acceptance.md` 에 미커밋 수정이 다시 나타났다.** 감사 창 안 쓰기 관측은 이번이 **세 번째**다(§0 · §11 · 이 절).

**Evidence.**

```
$ git rev-parse --short HEAD
8852cb921                       ← HEAD 는 이동하지 않았다(미커밋 쓰기)

$ git status --porcelain
 M .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/acceptance.md
 M .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/spec.md
?? .moai/reports/t1064/

$ git diff --stat -- .moai/specs/
 .../acceptance.md |  6 ++-
 .../spec.md       | 43 ++++++++++++++++++----
 2 files changed, 41 insertions(+), 8 deletions(-)
```

내용은 **§A.2 의 R4 형태 개정**이다 — 감사 좌표의 표 행이 명령 명명 형태로 대체되고, 이 감사가 §12.5 에서 요구한 **추적/미추적 축**이 들어와 있다:

```
# 커밋된 8852cb921 (감사 좌표)
86:| 저장소 전체 (`.git` 제외) | 21 |

# 워킹 사본 (미커밋 — 감사 대상 아님)
# (a) 좁은 뿌리 …  grep -rln … .claude internal | wc -l
# (b) 넓은 뿌리 — 저장소 전체(.git 제외). 미추적 파일을 포함한다
# (c) 커밋 트리만 — 같은 질문에서 미추적을 뺀 값   git grep -l … <ref> | wc -l
```

**Baseline-attribution.** 트리 `t1064`, HEAD `8852cb921`(불변), 이번 실행.

**판정에 미치는 영향 — 없다. 그리고 그 이유가 요점이다.**

- **감사 좌표는 `8852cb921` 이고 그 커밋에서 §A.2 는 여전히 `| 저장소 전체 (.git 제외) | 21 |` 이다.** D10 은 그 좌표에서 **열린 채로 확정된다.**
- **워킹 사본의 개정은 어떤 결함도 닫지 않는다.** 어느 브랜치에도 커밋되지 않은 텍스트는 정본으로 인용할 수 없고, 닫힘의 근거로도 쓸 수 없다 — §0 에서 같은 규칙을 적용했고 여기서도 동일하다.
- 내용이 이 감사의 처방과 방향이 같다는 사실은 **판정을 바꾸지 않는다.** 처방과 일치하는 미커밋 텍스트도 미커밋이다. 커밋되면 **iter-3 에서** 판정한다.

**Residual-risk — 이번 관측이 가장 위험한 형태다.**

앞선 두 번(§0 미커밋 1건 · §11 HEAD 이동)은 **내가 읽은 것이 커밋되지 않은 텍스트였다**는 귀속 오류를 낳았고, 그것은 사후에 정정할 수 있었다. 이번은 다르다 — **내가 방금 쓴 처방이 판정서에 착지하기 전에 대상 파일이 그 처방 방향으로 움직이고 있다.** 이 상태가 계속되면 다음 감사는 「내 처방을 반영한 결과」와 「내가 감사한 대상」을 구별할 수 없고, **감사는 자기가 만든 변화를 자기 근거로 읽게 된다.**

되돌리거나 커밋하지 않았다. **처분은 리드가 정한다.** D9 의 심각도를 여기서 올린다 — 단일 사건이 아니라 **감사 창 내내 지속된 상태**이며, 재발 방지 문안만으로는 부족하고 **쓰기 동결이 실제로 집행되는지**가 확인돼야 한다.

---

## iter-2 §14 — §13 귀속 정정, D9 재집계, 그리고 이 판정서의 유효 경계

**이 절은 §13 의 귀속을 정정하고 D9 를 재집계한다. 새 감사가 아니다** — 리드의 지시로 다음 창까지 감사를 멈춘 상태에서, **내가 §13 에 적은 사실 주장이 틀렸음이 확인됐으므로** 그것만 고친다.

### §14.1 §13 귀속 정정 — 쓰기 동결 위반이 아니었다

**§13 은 세 번째 쓰기를 「쓰기 동결 중 위반」으로 적었다. 그것은 틀렸다.**

리드가 밝힌 바: iter-2 판정이 착지한 뒤 **D9 + D10 범위로 동결을 해제**하고 R4 처방을 포함한 iter-3 지시를 manager-spec 에 보냈다. 내가 관측한 미커밋 편집은 **승인된 iter-3 작업의 실행**이었다. 해제 사실이 나에게 전달되지 않았을 뿐이다.

- **manager-spec 은 위반하지 않았다.** §13 의 「쓰기 동결 중」이라는 수식을 철회한다.
- 귀책은 **해제를 알리지 않은 전달 누락**에 있고, 리드가 자기 것으로 밝혔다.
- **나에게도 몫이 있다.** 나는 「동결이 선언됐다」는 **전달받은 상태를 현재 상태로 취급**했다. 그 진술도 전달 시점의 스냅샷이었고, 이 감사가 내내 지적해 온 **「스냅샷을 기억으로 쓰지 말고 재확인하라」**가 나 자신에게 적용되는 자리였다. 트리 상태는 두 번 재확인했으면서 **권한 상태는 한 번도 재확인하지 않았다.**

**철회되지 않는 것 — §13 의 추론은 그대로 유효하다.** 승인 여부는 **누구의 잘못인가**를 바꾸지만 **위험 자체를 없애지 않는다**: 내 처방이 적용되는 것과 내가 감사한 대상을 구별하지 못하는 감사는 결국 **자기 효과를 자기 근거로 읽는다.** 승인된 개정이라도 감사 창과 겹치면 그 구별이 무너지는 것은 같다. 리드도 같은 판단을 했다 — 「Authorization changes who is at fault; it does not change that hazard.」

### §14.2 D9 재집계 — 한 건이 아니라 **두 건**이다

같은 형태가 두 번 났고, **둘째는 첫째의 정정을 쓰는 중에 났다.** 한 건으로 적으면 재발 방지 문안이 **틀린 빈도를 겨눈다.**

| # | 사건 | 원인 | 내 판정서에 나타난 곳 |
|---|---|---|---|
| 1 | 개정과 감사를 같은 트리에 **동시 배차** → 감사 창 중 HEAD `3d636c991` → `8852cb921` 이동 | 리드 배차 오류 | §0 · §11 · §12.1 |
| 2 | 판정 확정 중 **동결을 해제하고 알리지 않음** → 미커밋 쓰기 3파일 | 리드 전달 누락 | §13(귀속은 이 절에서 정정) |

**공통 형태**: 감사 창의 경계가 **감사자에게 관측 가능한 형태로 유지되지 않았다.** 두 번 다 감사자는 트리가 고정돼 있다고 믿을 근거를 가졌고, 두 번 다 그 믿음이 조용히 낡았다.

**따라서 D9 의 필요 조치를 고쳐 쓴다.** 「재발 방지 문안 한 줄」이 아니라:
1. 배차문이 감사 창 동안 **해당 트리 쓰기 금지**를 명시하고,
2. **창의 시작·종료 SHA 를 감사자에게 명시**하며,
3. **창이 변경되면(해제·확대·축소 포함) 감사자에게 통지**한다 — 통지 없는 해제는 해제가 아니라 위 #2 의 재현이다.
4. 감사자 측: **권한 상태도 트리 상태와 같이 재확인 대상**으로 취급한다(내 몫).

### §14.3 리드가 수용·재현한 두 항목

- **Test 2 정정 수용됨.** 프로브가 `spec.md` 한 파일이었고, 한 파일의 불변(3/3)을 집합의 불변으로 일반화한 것이 오류였다. 빠진 파일은 `plan.md`, 제거한 커밋은 `3d636c991` 자신.
- **넷째 축(추적/미추적) 재현됨** — 리드가 `fe8159ed9` 에서 같은 ±1 을 얻었다고 전달했다. **이 감사는 그 수치를 재측정하지 않았으므로 인용하지 않는다**(§14.4). 축의 존재 자체는 이 감사가 `8852cb921` 에서 직접 쟀다(`git grep` 19 vs `grep -r` 20).
- **R4 의 값은 네 축**이다: 뿌리 · 패턴 · 시점(SHA) · 추적/미추적 범위.

### §14.4 [HARD] 이 판정서의 유효 경계 — `fe8159ed9` 는 **감사되지 않았다**

이 절을 쓰는 시점의 트리 상태를 재확인했다:

```
$ git rev-parse --short HEAD
fe8159ed9
$ git status --porcelain
?? .moai/reports/t1064/
$ git log --oneline -1
fe8159ed9 fix(SPEC-…): correct §A.2 counts to 20/3 and record the audit-window HEAD move
```

- **이 판정서의 감사 좌표는 `8852cb921` 이며 그 이상으로 확장되지 않는다.**
- **`fe8159ed9` 는 감사되지 않았고, 이 판정서는 그 커밋에 대해 어떤 판정도 내리지 않는다.** D10 은 `8852cb921` 에서 **열린 채 확정**됐고, `fe8159ed9` 가 그것을 닫았는지는 **미판정**이다 — 커밋 메시지가 「correct §A.2 counts to 20/3」이라고 말하는 것은 **주장이지 관측이 아니다.**
- 리드가 전달한 `fe8159ed9` 의 수치(19 / 20)도 **이 감사가 재측정하지 않았으므로 근거로 쓰지 않는다.**
- **D11 은 열려 있다.** iter-3 지시에 포함되지 않았고 `fe8159ed9` 가 AC-BGX-009 를 건드리지 않았다고 전달받았다 — 그 역시 **미검증 전달값**이다.

**다음 창에서 재측정할 것**(리드가 SHA 와 함께 창을 지명할 때):
1. `fe8159ed9`(또는 그때의 SHA)에서 **표제 수치 전량 재측정** — `8852cb921` 판독을 이월하지 않는다.
2. §A.2 의 R4 재작성이 **네 축 값을 실제로 치르는지** — 뿌리 · 패턴 · SHA · 추적/미추적. 하나라도 빠지면 R4 가 **값싼 침묵기**로 쓰인 것이다.
3. D9 기록이 **두 사건을 모두** 담는지.
4. D11(AC-BGX-009 의 시점·추적범위 미구속 + 상수 `8` 대조)의 처리 여부.

**판정 자체는 변하지 않는다 — Verdict PASS-WITH-DEBT, Overall 0.875, 좌표 `8852cb921`.** 이 절은 귀속과 경계를 고쳤을 뿐 어떤 결함도 닫지 않았다.

---

## iter-2 §15 — 21 의 귀속 정정: 「편집 중간 상태」가 아니라 **039af6915 의 워킹 트리**

**이 절은 §5 · §12.3 에 내가 쓴 21 의 성격 규정을 정정한다. 새 감사가 아니다** — 내가 쓴 사실 주장이 틀렸으므로 그것만 고친다. 어떤 결함도 닫지 않는다.

### §15.1 내가 틀린 것

iter-2 §5 와 §12.3 에서 나는 21 을 이렇게 적었다:

> 「21 은 **어떤 커밋에도 대응하지 않는 편집 중간 상태**의 값이다.」
> §12.3 표: 「`3d636c991` **커밋 전** 워킹 트리 · 추적 20 · 미추적 있음 · 합 21」

**「어떤 커밋에도 대응하지 않는다」가 틀렸다.** 21 은 **정확히 귀속되는 상태**의 값이다 — **커밋 `039af6915` 의 워킹 트리**, 즉 그 커밋 트리(추적 20) + 미추적 `verdict.md` 1.

내 §12.3 표의 둘째 행은 **부분 편집을 가정**해야 성립했다(`spec`/`acceptance` 는 고쳤고 `plan.md` 는 아직 안 고친 상태). 리드가 제시한 설명은 **어떤 편집도 가정하지 않는다** — 커밋 하나와 미추적 파일 하나로 완전히 지정된다. 둘 다 21 을 내지만, **재현 가능한 쪽은 후자뿐이다.**

### §15.2 이 감사가 직접 잰 것

(1) **추적 집합** — §12.3 (1)에서 이미 측정했고 값이 그대로다:

```
$ git grep -l 'tool-spawned subagent' 039af6915 | wc -l   → 20
$ git grep -l 'tool-spawned subagent' 3d636c991 | wc -l   → 19
```

(2) **`verdict.md` 는 무시된 것이 아니라 미추적이다** — 이 감사가 iter-1 에서 **미해결로 남긴 모호성을 여기서 확정한다**:

```
$ git check-ignore .moai/reports/t1064/verdict.md ; echo $?
1                                        ← 무시되지 않음

$ git check-ignore -v .moai/reports/t1064/verdict.md ; echo $?
.gitignore:259:!.moai/reports/*/verdict.md	.moai/reports/t1064/verdict.md
0                                        ← 규칙이 매치했다는 뜻이지 「무시된다」가 아니다

$ git ls-files --error-unmatch .moai/reports/t1064/verdict.md ; echo $?
error: pathspec … did not match any file(s) known to git
1                                        ← 추적되지 않음
```

따라서 `20(추적) + 1(미추적, 무시 아님) = 21`. **리드의 정정이 이 감사의 측정으로 확인된다.**

> **[계측기 함정 — 기록해 둘 값어치가 있다]** `git check-ignore -v` 의 **종료 코드 0 은 「무시됨」이 아니다.** 네거션 규칙(`!`)이 매치해도 0 을 낸다. iter-1 에서 나는 `-v` 형태만 돌려 exit 0 을 보고 결론을 **규칙 문자열을 읽어서** 냈다 — 맞는 답이었지만 **근거가 판별식이 아니었다.** 판별식은 **플레인 `git check-ignore` 의 종료 코드**이며, 추적 여부는 **별개 질문**이라 `ls-files --error-unmatch` 가 따로 필요하다. 세 명령이 세 가지를 묻고, 하나로 둘을 답할 수 없다.

### §15.3 왜 현학이 아닌가 — 재발 교훈이 바뀐다

내 문안대로면 다음 독자는 **커밋 경계를 맞추면 재현된다**고 믿는다. 재현되지 않는다. **정확히 같은 SHA 를 골라도 재현 못 한다** — 고를 것이 하나 더 있기 때문이다: **커밋 트리인가 워킹 트리인가.**

이것이 §12.5 의 **넷째 축(추적/미추적)** 이며, §15.2 의 측정이 그 축의 **기제**를 보여준다. 미추적이되 무시되지 않은 파일은 **워킹 트리 순회(`grep -r`)에는 잡히고 커밋 트리 조회(`git grep`)에는 안 잡힌다.** 그래서 같은 뿌리·같은 패턴·같은 SHA 에서 두 수가 갈린다.

**정정된 교훈: 판별식은 「어느 커밋인가」가 아니라 「어느 트리를 셌는가」다.**

### §15.4 §12.3 표 정정

| 좌표 | 추적 | 미추적 `verdict.md` | 합 | 측정자 |
|---|---|---|---|---|
| `039af6915` **커밋 트리** | 20 | — | **20** | iter-1 감사 (판정서 저작 **전**) |
| `039af6915` **워킹 트리** | 20 | 있음 | **21** | manager-spec |
| `3d636c991` / `8852cb921` 워킹 트리 | 19 | 있음 | **20** | iter-2 감사 · 리드 |

두 움직임(`verdict.md` +1, `plan.md` −1)이 여전히 전체 델타를 설명하며, **D10 은 내가 쓴 그대로 유지된다** — 정정되는 것은 21 의 **성격 규정**이지 그 존재나 D10 의 필요 수정이 아니다. 다만 D10 의 필요 수정 문안에서 **「어떤 커밋에도 대응하지 않는다」를 「`039af6915` 의 워킹 트리 값이다」로** 바꿔 적어야 한다.

### §15.5 관측되지 않은 채 남는 것

**manager-spec 이 실제로 `039af6915` 의 워킹 트리를 쟀는지는 관측되지 않았다.** 21 을 내는 상태는 둘 이상 있고(그 워킹 트리 · 내가 가정했던 부분 편집 상태), 어느 쪽이었는지는 **측정 시각을 관측하지 않는 한 결정되지 않는다.** 리드의 설명이 우월한 이유는 **관측됐기 때문이 아니라 편집 순서를 가정하지 않고 완전히 지정되기 때문**이다. 그 등급 그대로 적는다 — **재현 가능한 설명이지 관측된 사실이 아니다.**

**판정 불변: Verdict PASS-WITH-DEBT, Overall 0.875, 감사 좌표 `8852cb921`.** `fe8159ed9` 와 그 이후는 여전히 **미감사**다(§14.4).

---
---

# plan 감사 판정서 — iter-3

> 위 iter-1(§1~§9) · iter-2(§0~§15) 는 그대로 둔다. 이 절부터가 iter-3 다.

- 감사자: plan-auditor · 반복 **iter-3 / 3**
- **감사 좌표: `bbc2df7b8`** (리드가 SHA 와 함께 지명한 창)
- 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064` · 브랜치 `WT-branchguard-exempt`
- **Verdict: PASS** · Overall **1.00** · Tier M 임계 0.80
- **`8852cb921` 판독을 이월하지 않았다** — 표제 수치 전량을 이 좌표에서 새로 쟀다.

## iter-3 §1 — 창 상태 확인

```
$ git rev-parse --short HEAD
bbc2df7b8
$ git status --porcelain
?? .moai/reports/t1064/
$ git log --oneline -3
bbc2df7b8 fix(…): bind all four count-moving coordinates in AC-BGX-009
fe8159ed9 fix(…): correct §A.2 counts to 20/3 and record the audit-window HEAD move
8852cb921 fix(…): close D8 by fixing the measurement method, not the grep recipe
```

**창이 깨끗하다.** 감사 대상 외 수정 0건. iter-2 의 세 차례 관측(§0·§11·§13)과 달리 이번 창에서는 **HEAD 이동도 외부 수정도 관측되지 않았다.** 리드가 명시한 창 조건(시작 SHA 고지 · 쓰기 금지 · 변경 시 사전 통지)이 실제로 집행됐다.

## iter-3 §2 — 표제 수치 (이 좌표에서 새로 측정)

```
$ grep -rl 'tool-spawned subagent' . | grep -v '^./.git/' | wc -l   → 20   # (b) 워킹 트리
$ git grep -l 'tool-spawned subagent' | wc -l                       → 19   # (c) 커밋 트리
$ grep -rl 'tool-spawned subagent' .claude internal | wc -l         → 8    # (a) 좁은 뿌리
REQ 10 · AC 12 · tier: M · version: "0.5.0"
```

(c) 19 는 리드가 이 좌표에서 낸 값과 일치한다 — **인용이 아니라 재현이다.**

## iter-3 §3 — Must-Pass 재판정 (전 항목 PASS)

| 항목 | 판정 | 근거 |
|---|---|---|
| MP-1 | PASS | `REQ-BGX-001`~`010` 연속, 중복 0 |
| MP-2 | PASS | 10건 전부 `shall`/`shall not` + 주어 + 응답. **요건층(`REQ-XXX`) 기준 판정** |
| MP-3 | PASS | 정본 12필드 + optional `tier: M`. 거부 별칭 0 |
| MP-4 | N/A | 단일 프로그래밍 언어(Go) |
| MP-5 | PASS | 변동 없음 |
| MP-6 | PASS | `syscall` 0 (양성 대조 `HARD` 19행 발화) |
| MP-7 | PASS | `[NEEDS CLARIFICATION` 0 (같은 양성 대조) |

## iter-3 §4 — 다섯 항목 판정

### 항목 1 — 표제 수치 신규 측정: **완료**

§2. 이월 없음.

### 항목 2 — §A.2 의 R4 가 네 축 값을 치르는가: **치른다 (CLOSED)**

**Claim.** R4 가 요구하는 것(결정 명령을 먼저 적고 값을 날짜·ref 붙은 참조로 강등)이 실제로 이행됐고, **네 축이 모두 명령 자체에서 드러난다.**

**Evidence.** §A.2 원문:

> **[HARD] 이 개수는 값이 아니라 명령으로 산다.** 아래 셋이 판정식이며, 인용할 때는 **명령을 먼저 적고 값은 괄호 안에 날짜·측정 ref 와 함께 참조로만** 단다.

```bash
# (a) 좁은 뿌리 — 룰과 코드만
grep -rln "tool-spawned subagent" .claude internal | wc -l
# (b) 넓은 뿌리 — 저장소 전체(.git 제외). 미추적 파일을 포함한다
grep -rln --exclude-dir=.git "tool-spawned subagent" . | wc -l
# (c) 커밋 트리만 — 같은 질문에서 미추적을 뺀 값
git grep -l "tool-spawned subagent" <ref> | wc -l
```

> 참조값 — **2026-09-21, ref `8852cb921` 에서 측정**(읽는 시점의 사실이 아니라 그때의 관측): (a) **8**, (b) **20**, (c) **19**.

| 축 | 어디서 드러나는가 | 판정 |
|---|---|---|
| 뿌리 | (a) 대 (b) — 두 명령이 나란히 | **치름** |
| 추적/미추적 | (b) 대 (c) — 「미추적 파일을 포함한다」/「미추적을 뺀 값」이 명령 주석에 | **치름** |
| 시점 | 참조값에 날짜 + `ref 8852cb921`, 그리고 (c) 가 `<ref>` 를 인자로 요구 | **치름** |
| 패턴 | 「경계 앵커가 필요 없다 — **다만 접두 열거 의무는 면제되지 않는다**: 「부분문자열이 아니다」는 이 시점의 관측이고 그 판정은 세는 시점에 다시 내린다」 | **치름** |

넷째가 특히 잘 됐다 — 「앵커 불필요」라는 **면제 주장 자체에 시효를 붙였다.** 값싼 침묵기가 아니다.

**참조값 3건을 이 감사가 검증했다**: `8852cb921` 에서 (a) 8 · (b) 20 · (c) 19 — iter-2 §12.2·§12.3 의 내 측정과 **전부 일치**.

### 항목 3 — D9 기록이 두 사건을 다 담는가: **아니다 (OPEN, optional)**

**Claim.** `progress.md` 의 D9 기록은 **사건 1건만** 담는다. 사건 #2(판정 확정 중 동결 해제, 감사자 미통지)가 없다.

**Evidence.**

```
$ grep -c '배차 오류' progress.md        → 1     ← 사건 #1 (양성 대조)
$ grep -c '해제\|통지\|알리지' progress.md → 0     ← 사건 #2 부재
```

양성 대조가 발화하므로 0 은 **실제 부재**다. 기록된 재발 방지도 **레인 쪽 한 줄**뿐이다(「감사 산출물 경로가 미추적으로 보이면 커밋 전에 리드에게 창 상태를 묻는다»). 사건 #2 의 방지 조항(**창 변경 시 감사자에게 사전 통지**)은 리드 쪽 규율이라 여기에 없다.

**등급을 optional 로 두는 이유**: 이 SPEC 은 D9 를 **기록만 하고 규율화는 명시적으로 범위 밖**으로 선언했고(「이 카드는 그 규율을 코드나 룰로 만들지 않는다 — 범위 밖이다」), 사건 #2 는 레인 행동으로 예방되지 않는다. 또한 **두 사건 모두 이 판정서 §14.2 에 영구 기록돼 있어 프로젝트에서 소실되지 않는다.** 다만 리드가 지적한 위험(「한 건에 대고 쓴 방지선은 틀린 빈도를 겨눈다」)은 실재하므로 보고한다.

### 항목 4 — D11 처리: **CLOSED, 내가 지정한 것보다 강하게**

AC-BGX-009 (1)이 **네 좌표 보고 의무 표**로 재작성됐고, 각 행이 **그 축이 실제로 수를 움직인다는 실측**을 달고 있다(8 대 20 · 0 이동과 48 대 28 · 20·20·21 · 19 대 20).

내가 요구한 것보다 강한 지점 둘:

1. **추적 범위가 「어느 쪽을 썼는지 밝힌다」에서 「두 수를 모두 적는다」로 승격** — 뿌리와 대칭. 근거도 정확하다: 「한쪽만 적으면 읽는 사람은 그 수가 다른 쪽에서 몇인지 알 길이 없다」.
2. **시점이 무조건이 됐다** — 이전에는 run 값이 참조값과 다를 때만 ref 를 적게 돼 있었다.

그리고 상수 대조가 **명문으로 금지**됐다: 「어떤 plan 시점 상수도 합격 판정의 비교 대상이 되지 않는다 … 맨 상수 대 측정값의 대조는 비교가 아니다.» 잔존 확인:

```
$ sed -n '/AC-BGX-009/,/AC-BGX-010/p' acceptance.md | grep -n '8 재확인\|재확인했다'
18:**[HARD] 어떤 plan 시점 상수도 …** 「그 수를 재확인했다」는 합격 사유가 아니다.
```

유일한 적중이 **그것을 금지하는 문장**이다. 상수 비교항 0건.

**패턴 축을 의무에 남긴 판단이 옳다**: 「「움직이지 않았다」를 **보인 것**과 **재지 않은 것**이 출력으로 구별되지 않기 때문」 — 이 감사가 iter-1부터 반복해 온 원리를 SPEC 이 스스로 적용했다.

### 항목 5 — D10 이 이 좌표에서 실제로 닫혔는가: **CLOSED**

**Claim.** 닫혔다. 수치·내역·경위·성격 규정 넷 다 정정됐고, 표는 **커밋에 대고 직접 검증된다**.

**Evidence — §A.2 표를 커밋 3개에 대고 재측정:**

```
$ git grep -l 'tool-spawned subagent' 039af6915 | wc -l                    → 20
$ git grep -l … 039af6915 -- .moai/specs/SPEC-BRANCHGUARD-EXEMPT-REACH-001/ → 4
$ git grep -l … 3d636c991 | wc -l                                          → 19
$ git grep -l … 3d636c991 -- …/SPEC-BRANCHGUARD-EXEMPT-REACH-001/          → 3
$ git grep -l … 8852cb921 -- …/SPEC-BRANCHGUARD-EXEMPT-REACH-001/          → 3
```

SPEC 의 표와 **한 칸도 어긋나지 않는다.**

셋째 정정도 확인했다 — **−1 이 명시**되고(`3d636c991` 의 `-` 행 원문 인용), **은폐가 명명**되고(「iter-2 가 첫째만 적어 둘째를 가렸다»), **21 의 성격이 정정**됐다(「편집 중간 상태가 아니다 … `.gitignore:259` 네거션이 걸려 무시 대상이 아니라 단지 미추적이고, `039af6915` 의 워킹 트리를 정확히 센 값」). 마지막 것은 내 iter-2 §5·§12.3 의 오기를 고치는 내용이며 **내 §15 와 일치한다.**

## iter-3 §5 — 리드가 전달한 「브리핑 stale」 주장 검증

전달받은 세 항목을 `fe8159ed9` 에 대고 직접 쟀다. **셋 다 정확하다.**

| 전달된 주장 | 판정 | 근거 (`fe8159ed9` AC-BGX-009) |
|---|---|---|
| 추적 범위가 **부분적으로는 이미** 묶여 있었다 | **맞음** | (1)에 「(b)와 (c) 중 **어느 것을 인용하는지 명시한다**」 — 강도가 「어느 쪽인지」였지 부재가 아니었다 |
| 상수 `8` 은 (1) 본문에서 이미 제거됐고 **중첩 인용구에** 남아 있었다 | **맞음** | 본문에 `(8)` 없음. 인용구에 「패턴만 흔들고 「8 재확인」으로 지나가면 이 항목은 FAIL 이다」 |
| 실제로 안 묶인 축은 **시점** | **맞음** | 「run 시점 값이 **다르면** … ref 를 나란히」 — 조건부였다 |

**따라서 내 D11 은 `8852cb921` 기준으로는 옳았고 `fe8159ed9` 기준으로는 부분적으로 낡았다.** 남아 있던 것은 추적 범위의 **부재가 아니라 강도**, 상수의 **부재가 아니라 위치**, 그리고 시점의 **조건부성**이었다. manager-spec 이 그 차이를 스스로 재고 기록한 것은 정확한 처리다.

내 쪽 교훈: **결함을 등급과 함께 넘길 때 좌표도 함께 넘겨야 한다.** 「D11」이 아니라 「`8852cb921` 기준 D11」로 적었어야 했다.

## iter-3 §6 — 남은 결함

- **D9-rec** — `progress.md` D9 기록이 동시 쓰기 **2건 중 1건만** 담는다(§4 항목 3). — **Severity: minor · Class: optional** — **권고**: 사건 #2(동결 해제 미통지)를 한 줄 추가하고, 방지 조항을 레인 쪽(현행)과 **리드 쪽(창 변경 시 사전 통지)** 둘로 적는다. 두 사건 모두 이 판정서 §14.2 에 영구 보존돼 있다.

- **D12**(신규) — `spec.md:57` · `:81` 의 제목이 **「변주 축은 패턴이 아니라 탐색 뿌리다」** 로 단일 축을 선언하는데, 같은 SPEC 의 AC-BGX-009 (1)은 이미 **네 좌표 모델**이다. 제목만 읽는 독자는 원래의 단일 축 오류를 그대로 재생산한다. — **Severity: minor · Class: optional** — **권고**: 두 제목을 「변주 축은 넷이다 — 뿌리·패턴·시점·추적 범위」류로 고친다. 본문은 이미 맞으므로 제목만 움직이면 된다.

**차단 결함 0건.**

## iter-3 §7 — 점수와 판정

| 차원 | iter-2 | **iter-3** | 사유 |
|---|---|---|---|
| Clarity | 1.00 | **1.00** | D12 는 제목 한 줄이며 요건 모호성이 아니다 |
| Completeness | 0.75 | **1.00** | D10 닫힘. 전 절 존재, 프론트매터 완전, Out of Scope 5개 |
| Testability | 0.75 | **1.00** | D11 닫힘. 상수 비교항 0, 네 좌표 의무, 「보인 것 대 안 잰 것」 구별까지 |
| Traceability | 1.00 | **1.00** | REQ 001~010 · AC 12건 전부 매핑 |

**Overall = 1.00** · Must-Pass 7/7 · 차단 결함 0 → **Verdict: PASS**

`version` 0.5.0 및 HISTORY 한 줄의 범위 밖 기재에 **이견 없다** — AC 가 실질적으로 바뀌었으므로 버전 고정이 오히려 드리프트다. 리드의 수용에 동의한다.

## iter-3 §8 — 여전히 미측정 (등급 보존)

**이 PASS 는 「계획이 건전하다」이지 「질문이 답해졌다」가 아니다.**

- **도달성 자체 — 미측정.** 리드가 넘긴 **PLAUSIBLE 등급 그대로**이며 세 차례 감사 어디서도 올리지 않았다.
- **`agent_id` 실제 도착 — 미측정.** 선언은 도착의 증거가 아니다.
- **계측기 삽입 — 미수행.** §A.4 는 **measured-feasible** 이지 executed-through 가 아니다.
- **키 집합 동등 비교 — 한 번도 수행되지 않음.** 방법을 고정했을 뿐.
- **`MOAI_BRANCH_GUARD_EXEMPT` 축 — carried-forward-unverified.**
- **`go test ./internal/hook/...`** — iter-1 에서 exit 0 / 11패키지 `ok` 를 얻었고, iter-3 에서는 **재실행하지 않았다.** `bbc2df7b8`·`fe8159ed9` 가 `.moai/specs/` 만 건드렸다는 것은 iter-2 §12.2 에서 두 커밋에 대해 쟀으나 **이 두 커밋에 대해서는 재지 않았다** — 추론이다.

## iter-3 §9 — 이 판정 자신의 Gaps / Residual-risk

**Gaps.**
- `.moai/specs/` 무편집(창 규율 준수). 판정서만 썼다.
- §A.2 표의 (b) 열 중 `039af6915` 의 **21 은 재측정 불가**다 — 과거 커밋의 워킹 트리(미추적 포함)는 복원되지 않는다. 커밋 트리 20 + 미추적 1 의 **파생값**이며, SPEC 본문이 그 유도를 적고 있어 근거는 충분하나 **직접 측정은 아니다**.
- §A.2 는 21 을 「정확히 센 값」이라 적는다. 이 감사는 §15.5 에서 그것을 **재현 가능하지만 관측되지 않은 것**으로 등급했고, SPEC 본문에는 그 caveat 이 없다. 결함으로 잡지 않는다 — 유도가 명시돼 있다.
- 20파일 각각의 **문장 단위** 판독은 세 차례 감사 내내 하지 않았다(파일 입도).
- `bbc2df7b8` 의 세션 귀속은 확인하지 않았다.

**Residual-risk.**
- **가장 큰 위험은 그대로다: 이 카드가 세 번째로 「미측정」으로 닫히는 것.** 계측 설계는 이제 세 번의 감사를 거쳐 단단하고, §A.4 가 실행 경로까지 확립했다. **남은 것은 오직 실행 여부다.** 설계 품질 1.00 이 측정을 대신하지 않는다.
- D12 를 방치하면 제목이 본문을 이긴다 — 사람은 제목을 먼저 믿는다.
- iter-3 의 창은 깨끗했으나 그것은 **이번 창의 관측**이다. 다음 창에서 같은 조건이 성립한다는 보장이 아니다.

---

## M2 판정 — 도달 가능 (CONFIRMED), 2026-09-27

**측정 기준선**: 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064` · HEAD `d49a4aec8` · 브랜치 `WT-branchguard-exempt`. 전 구간 이번 실행에서 측정. 상세 증거(명령 + 원문)는 `progress.md` §E.2 § M2 본 측정.

### 판정

`BranchGuard` 의 `manager-git` 신원 면제는 **도구로 spawn 된 서브에이전트에서 도달 가능하다.** 독트린 `main-checkout-branch-guard.md` § Mechanical Enforcement 의 「예외 축 둘 다 서브에이전트에서는 도달 불가」와, `internal/hook/branch_guard.go:682` deny 문안의 같은 주장은 **거짓**이다.

### 두 조각

**(가) 런타임은 정체성을 보낸다.** 서브에이전트의 PreToolUse 페이로드에 `agent_type` · `agent_id` 가 snake_case 로 실린다. 관측 원문: `.moai/reports/t1064/m2-keyset-observation.log` (57행, sha256 `d11ee8c90981426ce9ce33e415925dc0c81c4d1f5b06cdde2b51167598c9cf91`).

| 형태 | 행 수 | `agent_type` |
|---|---|---|
| 메인 세션 | 27 | 키 부재 |
| 서브에이전트 | 19 | `plan-auditor` |
| 서브에이전트 | 7 | `manager-develop` |
| 서브에이전트 | 4 | `t1261-m2-live` |

camel 두 철자(`agentType` / `agentId`)는 57행 어디에도 없다. snake 철자가 `HookInput` 의 태그와 일치하므로 디코더가 버리지 않는다. 네 번째 행의 `t1261-m2-live` 는 카탈로그 에이전트가 아니라 다른 세션이 spawn 시 붙인 이름이며, 이것이 사칭 경로를 구체화한다 — `agent_type` 은 **이름을 그대로 싣는다.**

**(나) 그 정체성이면 deny 가 억제된다.** 트리 빌드 바이너리에 측정된 키 집합 형태를 먹인 세 팔:

| 팔 | `agent_type` | 결과 |
|---|---|---|
| A | 키 부재(메인 세션 형태) | `deny` · `BRANCH_GUARD_VIOLATION: git checkout <branch/-b> in primary checkout` |
| B | `plan-auditor` | `deny` · 같은 reason |
| C | `manager-git` | **`allow`** |

A·B 가 공허 통과를 배제한다 — deny 는 두 형태 모두에서 발화하므로 C 의 allow 는 면제 외에 귀속되지 않는다. 코드 좌표: `isExemptAgent` (`branch_guard.go:593`), `input.AgentType == "manager-git"` (`:603`), 그 호출이 `isPrimaryCheckout` 보다 위(`:660`). 가드 활성 `workflow.yaml:160-161`.

### 처방 분기점 — 미결정

**CODE 분기는 구성 가능하다.** M4-b 는 발동하지 않는다: 서브에이전트임을 말해 주는 판별 필드 `agent_id` 가 페이로드에 실재함이 (가)에서 확인됐다. 따라서 이 카드의 상태는 「판별자가 없어서 DOC 단독」이 **아니라** CODE·DOC 양쪽이 후보인 상태이며, 선택은 운영자 게이트다.

- **CODE**: `agent_id` 존재를 판별자로 써서 면제를 메인 스레드 런치로 좁힌다. 정당한 `manager-git` 경로를 깰 위험이 있어 M5 의 양방향 변이가 필수다.
- **DOC**: 독트린 문안 · 주석 · deny reason 을 관측에 맞게 정정한다. 우회 가능성은 남고, 그것을 알려진 잔여 위험으로 명시해야 한다.

문안을 고치는 경우 `branch_guard.go:682` 의 deny reason 은 **사용자에게 직접 보이는 표면**이므로 동반이동 집합(M6)에서 빠질 수 없다.

### 남은 Gap

- 살아 있는 서브에이전트를 `manager-git` 이라는 이름으로 띄워 실제 Bash 를 쏘는 한 단계는 실행하지 않았다. 그 단계는 primary 체크아웃에서 브랜치 상태를 바꾸는 명령이 실제로 통과해야 성립하므로 의도적으로 하지 않았고, (가)와 (나)를 잇는 추론이 한 단계 남아 있다.
- `MOAI_BRANCH_GUARD_EXEMPT` 환경변수 축은 이번 판정의 범위가 아니다.
- `frozenZonePrefixes` 는 범위 밖 · 부수 발견 · 등급 PLAUSIBLE 유지.

---

## 증거 처분 (2026-09-27, 리드 지시)

### 보존 경로와 해시

| 파일 | 처분 | sha256 |
|---|---|---|
| `.moai/reports/t1064/verdict.md` | 이 브랜치에 `git add -f` 로 추적(develop 관행 — 판정서 531개가 같은 방식) | 커밋 시점 값은 `git log` 로 확인 |
| `.moai/reports/t1064/m2-keyset-observation.log` | 같음. 57행 관측 원문 | `d11ee8c90981426ce9ce33e415925dc0c81c4d1f5b06cdde2b51167598c9cf91` |
| `~/.moai/logs/t1064-keys.log` | 계측 원본. 위 사본과 바이트 동일함을 확인한 뒤 **삭제** | 삭제 전 `d11ee8c90981426ce9ce33e415925dc0c81c4d1f5b06cdde2b51167598c9cf91`(동일) |

리드가 제시한 두 선택지 중 **`git add -f` 추적**을 택했다. primary 체크아웃에 파일을 쓰지 않으므로 공유 트리 쓰기를 피하고, 증거가 병합과 함께 이동해 원격 착지 후에도 남는다. `.gitignore:235` 의 `.moai/reports/*` 는 `-f` 로 넘어간다.

### 위생 확인 — 관측 원문에 페이로드 내용이 실리지 않았음

계측기는 C2(키 이름 + 네 프로브 값)였고, 실제로 그 형태만 남았음을 형태 단언으로 확인했다:

```
$ grep -c '"command"' .moai/reports/t1064/m2-keyset-observation.log
0
$ grep -vcE '^[0-9T:Z-]+ cpd=[^ ]+ tool=[A-Za-z]+ keys=[a-z_,]+ probes=(NONE|\{.*\})$' \
    .moai/reports/t1064/m2-keyset-observation.log
0
```

57행 전부가 그 형태이며(형태 위반 0행), 명령 값 표식은 0행이다. 로그에 실린 식별 정보는 최상위 **키 이름**, `tool_name`, 그리고 `agent_type` / `agent_id` 값뿐이다 — 프롬프트 본문도 명령 문자열도 없다. `keys=` 목록에 `tool_input` 이 보이는 것은 **키 이름**이며 그 값이 아니다.

---

## 독립 sync-audit 기록 (2026-09-27, 리드 지시)

### 감사자 모델 귀속 — 자기선언이 아닌 관측 두 개

감사 판정을 인용하려면 그 감사가 어느 모델에서 돌았는지가 먼저 확립돼야 한다. 같은 날 GLM 레인 두 곳에서 `model: opus` 로 띄운 감사자가 실제로는 `glm-5.3-flash` 로 돌았고, 리드가 그 사고를 근거로 모델 id 표기를 재감사 조건으로 걸었다.

감사자 자신은 `auditor-model: claude-opus-5[1m]` 을 보고하면서 **그 값이 런타임의 자기선언이며 독립 측정이 아니라는 한계를 스스로 고지**했다 — 자기 스폰의 model 파라미터를 조회할 도구가 자기 `tools:` 에 없다는 이유까지 적었다. 그 고지가 맞으므로 이 값 하나로는 귀속이 성립하지 않는다.

**관측 1 — 서브에이전트 트랜스크립트의 `.message.model`** (이 레인이 직접 측정):

```
$ python3 <count .message.model in
  ~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go--claude-worktrees-t1064/
  337ea462-74d9-49b6-ae11-96d132b67a4c/subagents/agent-at1064-sync-audit-9eca3c6b97d79d83.jsonl>
rows carrying .message.model: 100
  claude-opus-5 100
```

**관측 2 — 스폰 메타** (같은 디렉터리의 `.meta.json`):

```
{"agentType":"t1064-sync-audit", … "model":"claude-opus-5[1m]",
 "customAgentType":"sync-auditor","permissionMode":"bypassPermissions"}
```

리드도 같은 트랜스크립트에서 독립적으로 96/96 `claude-opus-5` 를 얻었다(측정 시점이 이르러 행 수만 다르다). 1차 감사와 재감사가 같은 감사자 기록에 이어져 있으므로 **두 판정 모두 Opus 귀속**이다.

**막힌 경로 하나를 기록한다**: 감사자가 교차검증 경로로 가리킨 `.moai/logs/agent-model-audit.jsonl` 은 이 스폰에 대해 **0행**이었다(`t1064` + `audit` 매칭 0건). 그 로그는 advisory 이고 이 경로를 기록하지 않았으므로, 모델 귀속의 근거는 위 관측 두 개다.

### 판정

- **1차 감사: FAIL.** must-pass 는 중립성 단독이었다 — Functionality 62/100 이 must-pass 로 걸렸고 그 원인이 기존 가드 `TestTemplateNoInternalContentLeak` 의 적색(이 카드가 미러에 심은 `SPEC-` 접두)이었다. Security 88 PASS(must-pass), Craft 78 PASS, Consistency 55 FAIL(must-pass 아님).
- **재감사: PASS-WITH-DEBT.** 커밋 `43697af85` 가 세 리터럴을 0행으로 만들고 리크·중립성 가드를 초록으로 되돌렸음을 감사가 직접 실행해 확인했다. 중립성 쪽은 클래스별 서브테스트(`C1-macos-bias-path`·`C2-bare-narrative-v3r`·`C4-feedback-memory-ref`·`C5-claude-local-ref`·`C6-pr-number-ref`·`C9-natural-language-canonical-form`)가 전부 발화했으므로 공허 초록이 아니다.

### 감사가 확인해 준 것

- **렌즈 ① 통과.** 57행을 감사가 재계수해 카드 수치와 완전 일치(NONE 27 / nonNONE 30 / 식별자 7·19·4, camel 0). 교차 오염 0행 — NONE 행에 agent 언급 0, nonNONE 행에 `agent_id` 누락 0. camel 0 은 같은 프로브의 snake 두 철자 발화가 양성 대조로 붙는다고 판정했다.
- **렌즈 ③ 통과, 수리 후에도 회귀 없음.** 미러가 좌표를 잃었을 뿐 신원 축 실질 여섯 조각(측정됨·추론 아님 / snake_case / spawn 이름 verbatim / 세 팔 결과 / 우회 존재 / 거짓 주장 철회)과 센티널 축의 축 분리가 전부 남아 있고, 센티널이 검증된 척하는 지점은 0곳이다.
- **미러 분기는 기계적으로 승인된 형태다.** 감사가 `rule_template_mirror_test.go:65` 주변을 읽어, 바이트 동일 allowlist 부재가 누락이 아니라 **§25 sanitized pair 등록**이며 전용 가드 `sanitized_pair_parity_test.go` 의 `TestSanitizedPairParity` 가 독트린 등가성을 직접 판정해 통과함을 확인했다. 즉 이번 수리는 새 정책 도입이 아니라 원상복귀다.
- **비공허성** — 감사가 변이를 직접 실행해 `branch_guard_test.go:699` 에서 죽고 복원 후 ok 를 확인했으며, 트리 원복을 sha256(`7bdb76676ba2…`)과 빈 `status --porcelain` 으로 이중 확인했다.
- **부수 2 방어 가능.** 고치지 않은 14파일은 전부 `.moai/specs/**` + `CHANGELOG.md` 이고 템플릿 트리에 하나도 없다. live holder 의심 1건(`SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001/acceptance.md:307`)은 그 SPEC 자기 가드의 문안 제약이며, 같은 SPEC `spec.md` §F 가 BranchGuard 면제 도달성 질문을 별도 카드로 명시 유보한다.
- **부수 3.** deny 문안은 정확하고 실행 가능하다고 판정했다. 약 210자로 길지만 축약하면 비공허성 단언(`:699`)이 물 곳을 잃으므로 축약 비권고.
- **M4 커밋 격리** — `8b3464757` 은 `progress.md` 1파일, 문안 변경 0.

### 감사의 틀린 정정 1건 — 둘째 부모 미측정

감사는 재감사에서 `TestRuleTemplateMirrorDrift`(`worktree-integration.md`) 실패의 귀속을 「develop 은 깨끗하고 이 브랜치의 흡수 병합 `6ed2cb1f2` 가 분기를 만들었다. 병합 전에 처리해야 하며, 병합되면 develop 의 초록 가드를 적색으로 만든다」로 정정했다. **이 정정은 틀렸다.** 근거로 든 것은 첫 부모(58473/58473 동일)와 병합 결과(60597/60762)이며, **둘째 부모를 재지 않았다.**

```
$ git rev-parse --short 6ed2cb1f2^2
00e761af8
$ git show 00e761af8:.claude/rules/moai/workflow/worktree-integration.md | wc -c
   60597
$ git show 00e761af8:internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md | wc -c
   60762
```

흡수된 develop 스냅숏이 **이미** 갈려 있었다. 병합이 이 파일을 손대지 않았음은 해시로 확정된다 — HEAD 양쪽이 둘째 부모와 바이트 동일:

```
00e761af8  source fa93b1d76acd6a98ef19a8f32bce846eb01df0ab3bf193ddb062d036947274f3
           mirror 028609a2e46bba85279ef8e1776e32e07e7ce5cf4c36d23f293e93b535256c63
HEAD       source fa93b1d76acd6a98ef19a8f32bce846eb01df0ab3bf193ddb062d036947274f3
           mirror 028609a2e46bba85279ef8e1776e32e07e7ce5cf4c36d23f293e93b535256c63
$ git merge-base --is-ancestor 6ed2cb1f2^2 develop   →  yes
$ git log --oneline 00e761af8..develop -- <미러>      →  5커밋
$ git show develop:<양쪽> | wc -c                     →  61749 / 61749
```

즉 드리프트는 흡수 시점 develop 의 결함이고 develop 이 그 뒤 고쳤다. 「develop 은 깨끗하다」는 감사의 관측은 **현재** develop 에 대해서만 참이며, 흡수 시점 develop 에 대해서는 거짓이다. 결론도 뒤집힌다 — 병합 전 별도 수리는 필요하지 않고, 신선한 흡수가 `61749/61749` 를 가져오면 해소된다. 리드가 develop head `b59a5d69c` 에서 이 테스트를 돌려 해당 서브테스트까지 PASS 를 확인한 것과 같은 결론이다.

**기록하는 이유**: 감사 판정을 그대로 받았다면 병합 전에 불필요한 수리를 했을 것이고, 그 수리는 develop 이 이미 고친 파일을 이 브랜치에서 손으로 되돌리는 형태가 됐을 것이다. 감사자의 판정은 근거를 보고 받는 것이며, 근거가 닿지 않는 축이 있으면 그 축을 재는 것이 수리자 몫이다.

### 감사가 남긴 Gap — 창의 병합 트리 재측정으로 이월

- `golangci-lint`·`gofmt`·`go vet` 을 감사가 재실행하지 않았다(카드의 v2.1.6 0 issues 는 감사 미검증).
- `make build` 후 임베드 축 미측정 — 리크 2건이 설치 바이너리에 실렸는지 미확인.
- 미러 전역 중립성 스윕 미수행 — 두 미러 파일 범위만 쟀고 C7(커밋 SHA) 축은 미측정.
- 살아 있는 `manager-git` 서브에이전트의 실 Bash 발사는 감사도 하지 않았다(카드와 같은 이유).
- `TestSessionStart_MissPathSpendsNoJoinBudgetOnDrift` 1회 사망(`Handle took 293.716459ms` vs 벽시계 예산 250ms). 격리 `-count=3` 3/3 통과로 부하 의존 flake 로 판단했으나 표본 4회이며, `session_start` 소관으로 `branch_guard` 와 무관하다.

---

## 병합 트리 재측정 (2026-09-27, 통합 창 — agent-48)

창을 `moai integration acquire --name agent-48 --card t1064` 로 잡았고(settings 드리프트 적중 없음 — 거절 없이 통과), 로컬 develop `b59a5d69c`(= `origin/develop`)를 흡수했다. 흡수 전 HEAD `c900d0193`, 흡수 후 **`82a3a3d3c`**, 충돌 0.

카드 작업 전체가 미흡수 트리에서 이뤄졌으므로(배차 지시 미이행) 여기의 측정이 판정의 근거이며, 앞선 §E.2 측정은 미흡수 트리 기준으로 남긴다.

### 흡수가 해소한 것 — 드리프트 귀속의 확인

```
$ wc -c < .claude/rules/moai/workflow/worktree-integration.md
   61749
$ wc -c < internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
   61749
```

흡수 전 60597 / 60762 였던 두 사본이 일치했다. 이 카드가 아무것도 고치지 않고 흡수만으로 해소됐으므로, § 독립 sync-audit 기록의 귀속 판정(감사의 정정이 아니라 레인 측정이 맞다)이 결과로도 확인된다.

### 재측정 결과 — 전항목 초록

| 측정 | 명령 | 결과 |
|---|---|---|
| 훅 전 패키지 | `go test ./internal/hook/... -count=1` | **exit 0** — 11패키지 전부 ok(`internal/hook` 430.893s) |
| 템플릿 가드 4종 | `go test ./internal/template/ -run 'Leak\|Neutral\|MirrorDrift\|SanitizedPair' -count=1` | **exit 0** — ok 1.845s |
| 바이너리 지연 | `go test ./internal/cli -run TestBinaryLag -count=1` | **exit 0** — ok 1.011s |
| 빌드 | `make build` | **exit 0** — commit stamp `82a3a3d3c`, 빌드 후 `status --porcelain` **0행**(재생성 산물이 커밋본과 동일) |
| 임베드 축 | `make embed-check` | **exit 0** — `Agent Emit Embed: 12/12 embedded agent-emit artifacts match the committed set`, 1 ok / 0 warn / 0 fail |
| lint (CI 판) | `golangci-lint run ./internal/hook/... ./internal/template/...` (v2.1.6) | **0 issues** |
| 포맷 | `gofmt -l internal/` | **0행** |
| 정적검사 | `go vet ./internal/hook/... ./internal/template/...` | **exit 0** |

**감사가 flake 로 판단한 항목**: `TestSessionStart_MissPathSpendsNoJoinBudgetOnDrift` 는 이번 병합 트리 전 패키지 실행에서 **발화하지 않았다**(exit 0). 부하 의존이라는 감사 판단과 일관되나, 이 1회 통과가 flake 아님을 확립하지는 않는다 — 벽시계 예산 테스트이므로 병렬 레인 환경에서 재발 가능성은 남는다.

### 미러 전역 중립성 스윕 — 측정의 한계를 함께 적는다

리드가 요구한 C7 포함 전역 스윕을 돌렸으나, **crude regex 스윕은 가드가 아니다.** `internal/template/templates/` 전역에 `SPEC-`·날짜·`card t<n>`·9자리 hex 적중이 다수 있고 그것들은 전부 기존 파일이다. 실제 C 클래스 규칙은 `TestTemplateNoInternalContentLeak` 과 중립성 가드가 갖고 있으며, 그 가드들이 초록이다. 따라서 텍스트 적중을 결함으로 주장하지 않는다.

귀속 가능한 형태로 잰 것은 **이 카드가 만진 두 미러 파일**이며, 네 축 전부 0이다:

```
$ for pat in 'SPEC-[A-Z0-9-]+-[0-9]{3}' '20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]' \
             'card t[0-9]+' '\b[0-9a-f]{9}\b'; do
    grep -rlE "$pat" <두 미러> | wc -l
  done
0
0
0
0
```

### Baseline-attribution

- 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064` · 브랜치 `WT-branchguard-exempt` · 흡수 후 HEAD **`82a3a3d3c`** · 흡수원 로컬 develop `b59a5d69c`(= `origin/develop`)
- 위 표의 모든 행은 **이번 실행에서 병합 트리에 대해** 돌린 명령의 결과다. 흡수 전 측정값을 옮긴 것은 없다.
- 측정 중 `git status --porcelain` 0행 — `make build` 직후에도 0행.

### Gaps

- **살아 있는 `manager-git` 서브에이전트의 실 Bash 발사는 여전히 미실행.** 병합 트리에서도 하지 않았다 — 성립 조건이 primary 체크아웃의 브랜치 상태 변경이 실제로 통과하는 것이므로 의도적이다.
- **`MOAI_BRANCH_GUARD_EXEMPT` 환경변수 축 미측정** — 이 카드의 범위가 아니며 문안에도 「미재측정」으로 적혀 있다.
- **전 패키지 스위트(`go test ./...`)는 돌리지 않았다** — §4.1 대로 변경 영향 범위(`internal/hook`, `internal/template`, `internal/cli` 의 BinaryLag)만 쟀고, 전 패키지 판정은 CI 몫이다.
- **`TestSessionStart…` flake 판정의 표본은 여전히 작다**(감사 4회 + 이번 1회).
- **crude 중립성 스윕은 가드를 대체하지 않는다**(위 절에 명시).
