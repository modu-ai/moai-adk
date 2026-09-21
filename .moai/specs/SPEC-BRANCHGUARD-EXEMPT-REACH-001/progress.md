# 진행 기록 — SPEC-BRANCHGUARD-EXEMPT-REACH-001

card: t1064 · 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1064` · 브랜치: `WT-branchguard-exempt`

## §E.1 Plan-phase Audit-Ready Signal

### Claim

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
