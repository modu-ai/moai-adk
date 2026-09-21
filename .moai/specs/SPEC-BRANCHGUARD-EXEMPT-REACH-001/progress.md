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
- **동반이동 집합 10 도 하한이다.** 두 뿌리와 두 패턴 축으로 쟀을 뿐, 20/21 파일 각각의 **문장 단위**가 같은 주장을 싣는지는 파일 단위 적중까지만 확인했다.
- **`agent_type` 실제 적중 28 은 파생값이다.** 48 − 20 의 뺄셈이고 앵커를 단 직접 측정이 아니다 — run 단계가 인용하려면 앵커로 다시 재야 한다.

### Residual-risk

- **계측기 설계가 실패하면 카드가 「미측정」으로 닫힌다.** 선행 카드가 이미 그렇게 닫혔고, 이 SPEC 은 그 결과를 정상 종료의 하나로 인정한다(§C). 다만 그때 CONTESTED 표식은 「또 미측정」으로 갱신돼야 하며 그대로 두면 다음 사람이 같은 벽을 재발견한다.
- **부정 분기의 조용한 통과.** 도달 불가가 관측되면 「독트린이 맞았다」로 닫고 싶은 압력이 강하다 — 기제 검증(AC-BGX-004)이 그 압력을 막는 유일한 장치다.
- **설치본 판정 위험.** 훅 경로가 이 트리 빌드가 아니라 설치본을 실행하면 계측기 없는 옛 빌드가 조용히 「필드 없음」을 낸다(VCI §2.2). run 단계 M2 1항이 이것을 막는다.
- **8파일은 하한.** 굵게 표기·줄바꿈이 낀 같은 주장은 리터럴 grep 에 안 잡힌다. 문안 정정 시 패턴을 흔들지 않으면 일부 지점이 옛 주장을 싣고 남는다.

---

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
