# t23 — kanban-strategy 서브에이전트 무보고 idle 조사

- 조사 범위: 정의 파일 판독 + 기록 조회 (read-only, 재현 시도 없음)
- 기준 트리: `main` @ `3b9b3bf99`
- 근본 원인: **미확정**. 아래 가설은 전부 가설로 표기했고, 실측한 것만 Evidence에 넣었다.

---

## 1. Claim

| # | 주장 | 상태 |
|---|---|---|
| C1 | `manager-kanban.md`에는 최종 보고 계약(Output Format / 산출 경로)이 **아예 없다** | 실측 |
| C2 | 두 역할(Role A / Role B) 어느 쪽의 진입 조건도 "분석 위임"으로는 충족되지 않으며, 미충족 시 행동 지침이 없다 | 실측 |
| C3 | 런타임은 `SubagentStop`을 **두 번 정상 발화**했다 — 멈춘 게 아니라 **빈 본문으로 깨끗이 종료**했다 | 실측 |
| C4 | `memory: project` 컨텍스트 과적재는 원인이 **아니다** (해당 디렉터리 0바이트) | 실측(반증) |
| C5 | `eager_context_weight`는 에이전트별 신호가 아니다 (전 에이전트 동일값) | 실측(반증) |
| C6 | 중첩 spawn이 보고를 삼켰을 가능성 | **가설 — 미검증** |
| C7 | 정의 결함이 단독 원인이다 / 런타임 결함이 단독 원인이다 | **둘 다 미판별** |
| C8 | 직전 기록의 무보고 사례와 **형태가 일치**한다 | 실측 |

---

## 2. Evidence

### E1 — 보고 계약 부재 (C1)

`.claude/agents/moai/manager-kanban.md` 전문 179줄을 판독했다. `## Output Format`, `## Report`, `Write the ... report to` 중 어느 것도 없다.

```
$ for f in .claude/agents/moai/*.md; do
    n=$(grep -c -i "^## Output Format\|^## Report\|Write the .*report to" "$f"); echo "$n  $(basename $f)"; done
0  builder-harness.md
0  e2e-tester.md
0  manager-design.md
0  manager-develop.md
0  manager-docs.md
0  manager-git.md
0  manager-kanban.md      ← 조사 대상
0  manager-spec.md
2  plan-auditor.md
0  super-advisor.md
1  sync-auditor.md
```

보고에 가장 근접한 문장은 `manager-kanban.md:40`의 한 구절뿐이다:

> "…reduces schema-driven fan-out returns into a single consolidated report."

"consolidated report"를 만들라는 말은 있으나 **어디에 어떤 형식으로 내놓는지가 없다**. 산출 경로도, 헤딩 스켈레톤도, "응답 본문에 반환하라"는 지시도 없다.

**중요한 단서**: 11개 중 9개가 0점이므로 계약 부재 자체는 manager-kanban 고유 결함이 아니다. 다만 그 9개 중 **`Agent` 도구를 든 유일한 에이전트**가 manager-kanban이다 (`manager-kanban.md:10`, `manager_kanban_depth_test.go`가 이 유일성을 강제). 즉 "보고 계약 없음 × 하위 spawn 가능"의 조합은 manager-kanban에만 있다.

### E2 — 동료 에이전트와의 구체적 차이 (C1)

두 감사관은 최종 산출을 **명시적으로 못 박는다**. 방식은 서로 다르다.

`plan-auditor.md:357-359` — **파일 산출물**:
```
## Output Format

Write the audit report to `.moai/reports/plan-audit/{SPEC-ID}-review-{iteration}.md`.
```
`plan-auditor.md:452`는 실패 경로까지 규정한다 — spec.md가 없으면 `"AUDIT BLOCKED: spec.md not found at {path}"` 한 줄을 반환하고 종료. **무보고 종료가 정의상 허용되지 않는다.**

`sync-auditor.md:68-88` — **응답 본문 + 고정 스켈레톤**:
```
## Output Format

## Evaluation Report
SPEC: {SPEC-ID}
Overall Verdict: PASS | FAIL
### Dimension Scores
| Dimension | Score | Verdict | Evidence |
...
```
`sync-auditor.md:126`는 한술 더 떠 **왜 파일이 아닌지**까지 밝힌다: "RETURN … in the response body for the orchestrator to persist … this agent has no Write tool (`permissionMode: plan`) and MUST NOT attempt a file write."

세 에이전트의 차이를 한 줄로:

| 에이전트 | 최종 산출 지정 | 회수 경로 |
|---|---|---|
| plan-auditor | 파일 경로 명시 + 실패 시 반환 문자열까지 규정 | 파일 — 무보고여도 회수 가능 |
| sync-auditor | 응답 본문 + 표 스켈레톤 + 본문인 이유까지 명시 | 본문 — 회수 불가, 단 형식이 고정돼 있어 "무엇을 못 냈는지"가 드러남 |
| **manager-kanban** | **없음** | **없음** |

### E3 — 런타임은 정상 종료를 두 번 발화했다 (C3)

`.moai/harness/usage-log.jsonl` 105222행 / 105254행:

```
{"timestamp":"2026-08-14T12:30:57.583911Z","event_type":"subagent_stop","agent_type":"kanban-strategy",
 "agent_id":"akanban-strategy-7ce3add58e9ab36b","parent_session_id":"a7e2da9d-…","eager_context_weight":1082787}
{"timestamp":"2026-08-14T12:32:52.602635Z","event_type":"subagent_stop","agent_type":"kanban-strategy",
 "agent_id":"akanban-strategy-7ce3add58e9ab36b","parent_session_id":"a7e2da9d-…","eager_context_weight":1082787}
```

읽어낼 수 있는 것 세 가지:

1. **두 번 다 `subagent_stop`이 발화했다.** 행이 걸린 것도, 타임아웃도, 크래시도 아니다. 에이전트는 매번 **정상적으로 턴을 끝냈고, 그 턴의 본문이 비어 있었다.** 이건 진단 방향을 크게 좁힌다 — "죽어서 못 보냈다"가 아니라 "살아서 아무것도 안 냈다"다.
2. **`agent_id`가 동일하다.** 재요청이 새 에이전트를 띄운 게 아니라 **같은 에이전트를 이어서 돌렸다**. t23이 말하는 "재요청 후에도 동일"은 독립 2회 시행이 아니라 **같은 문맥의 연속 2턴**이다 — 즉 두 번째 실패는 첫 번째와 독립 증거가 아니다.
3. 두 이벤트 간격은 115초.

같은 세션(`a7e2da9d`)의 라우팅 기록 `.moai/state/routing-pending-a7e2da9d-….json`에도 `kanban-strategy` 위임 2건이 `"outcome":"unknown"`으로 남아 있다 — 오케스트레이터 쪽에서도 결과를 귀속하지 못했다.

### E4 — 반증 두 건 (C4, C5)

세워봤다가 측정으로 무너뜨린 가설 두 개를 기록해 둔다.

**메모리 과적재 가설 (기각).** `manager-kanban`은 `memory: project`(`manager-kanban.md:15`)를 선언하므로 spawn 시 `.claude/agent-memory/manager-kanban/`이 주입된다. 이게 비대하면 컨텍스트를 먹어 퇴화 턴을 만들 수 있다는 가설:
```
$ du -sh .claude/agent-memory/*
432K	.claude/agent-memory/plan-auditor
292K	.claude/agent-memory/manager-spec
196K	.claude/agent-memory/manager-develop
  0B	.claude/agent-memory/manager-kanban      ← 비어 있음
```
manager-kanban 메모리는 **0바이트**이고, 디렉터리 자체가 Aug 14 21:27 생성 — 사고 시각(12:30)보다 **9시간 뒤**다. 사고 당시엔 주입될 메모리가 존재하지도 않았다. 기각. (덧붙여 정작 메모리가 가장 비대한 plan-auditor가 가장 안정적으로 보고하는 쪽이다.)

**컨텍스트 무게 가설 (신호 없음).** 로그의 `eager_context_weight: 1082787`(약 1.08MB)이 커 보였으나:
```
$ tail -3000 usage-log.jsonl | grep subagent_stop | (agent_type별 최댓값)
1082787  manager-spec / plan-auditor / manager-develop / manager-git / kanban-strategy / …  전부 동일
```
전 에이전트가 같은 값이다. 프로젝트 단위 상수이지 에이전트별 지표가 아니다. 신호 없음.

### E5 — 역할 미해소 구간 (C2)

두 역할의 진입 조건은 이렇게 정의돼 있다.

Role B (`manager-kanban.md:46`):
> "Role B has a different and simpler entry: **the session's SessionStart context declares Kanban Mode with the `lead` role.**"

Role A (`manager-kanban.md:48-52`) — 세 조건 **AND**:
> "≥3 milestones … AND ≥10 files … AND cross-domain (≥3 distinct domains)"

서브에이전트 spawn에는 SessionStart 컨텍스트가 없으므로 Role B 진입 조건은 구조적으로 만족될 수 없다. 그리고 "전략 분석"은 SPEC 마일스톤도 파일 표면도 없으니 Role A의 세 조건도 못 채운다. **둘 다 아닐 때 무엇을 하라는 지시가 파일 어디에도 없다.**

가장 근접한 지침은 blocker report 반환(`manager-kanban.md:62`)이지만, 그 트리거는 "unresolved input / peer FAIL / `/compact` 불가" 세 가지로 열거돼 있고 **"위임 자체가 내 역할 어디에도 안 맞음"은 그 목록에 없다.** 대조적으로 plan-auditor는 진입 실패 시 반환 문자열까지 규정해 뒀다(E2).

Opus 4.8/5는 지시를 문자 그대로 따르고 암묵적 일반화를 하지 않는다는 것이 프로젝트 규약의 전제다(`moai-constitution.md` § Opus 5 / 4.8 Prompt Philosophy). 그 전제 위에서 "해당 역할 없음 + 폴백 미정의"는 빈 턴으로 귀결될 여지가 있다 — **이것이 가장 그럴듯한 정의 측 설명이지만, 여전히 가설이다** (§4 G2).

### E6 — 선례와의 형태 일치 (C8)

`memory/feedback_agent_idle_without_report.md`(2026-08-14 추가분)가 같은 형태를 이미 기록해 뒀다:

> "SPEC 카탈로그 전수 조사에서 `Agent(Explore)` 5명을 띄웠는데 **5/5 전원이 끝까지 미보고**로 idle. 쓰기 에이전트(`manager-spec`, `manager-docs`)는 같은 세션에서 **2/2 정상 산출**했다. 차이는 산출물의 유무다 — … **read-only 에이전트는 보고서가 곧 유일한 산출물**이라 미보고 = 작업 전량 소실이다."

t23은 이 축의 세 번째 관측이다. 축은 "read-only냐"가 아니라 **"턴 바깥에 남는 산출물이 있느냐"**로 읽는 게 정확하다 — manager-kanban은 Write 도구를 갖고 있으니(`manager-kanban.md:10`) read-only가 아니지만, **분석 위임에는 쓰라고 지시된 파일이 없었다**. 도구 유무가 아니라 지시된 산출물 유무가 갈랐다.

기록된 3회 사이의 공통 구조: (a) 산출물이 응답 본문뿐, (b) 보고 형식이 스폰 프롬프트에도 에이전트 정의에도 고정돼 있지 않음, (c) `subagent_stop`은 정상 발화.

---

## 3. Baseline-attribution

| 측정 | 명령 | 시점 |
|---|---|---|
| 정의 전문 | `Read .claude/agents/moai/manager-kanban.md` (179줄 전량) | 본 조사 |
| 계약 부재 | `grep -c "^## Output Format\|^## Report\|Write the .*report to"` × 11 파일 | 본 조사 |
| 동료 대조 | `grep -n` on `sync-auditor.md`(:68-88, :126), `plan-auditor.md`(:357-359, :452) | 본 조사 |
| 종료 이벤트 | `grep kanban-strategy .moai/harness/usage-log.jsonl` → 105222, 105254행 | 로그 기록 시각 2026-08-14T12:30:57Z / 12:32:52Z |
| 메모리 크기 | `du -sh .claude/agent-memory/*` | 본 조사 (0B, dir mtime Aug 14 21:27) |
| 무게 상수성 | `tail -3000 usage-log.jsonl` agent_type별 max 집계 | 본 조사 |
| 선례 | `memory/feedback_agent_idle_without_report.md` 전문 | 본 조사 |
| 깊이 가드 | `internal/template/manager_kanban_depth_test.go` 헤더 주석 1-60행 | 본 조사 |

기준 트리는 `main` @ `3b9b3bf99`. 사고 세션(`a7e2da9d`, 2026-08-14 12:30 UTC)의 트리 SHA는 확인하지 않았다 — 정의 파일이 그때와 지금 같다는 것은 **검증하지 않았다**(§4 G1).

---

## 4. Gaps — 관측하지 않은 것

| # | 미관측 항목 | 왜 중요한가 |
|---|---|---|
| **G1** | 사고 시점(2026-08-14 12:30)의 `manager-kanban.md` 내용. `git log` 대조를 하지 않았다 | 그 사이 정의가 바뀌었다면 §2의 정의 판독 전체가 다른 파일에 대한 것이 된다 |
| **G2** | **근본 원인.** 정의 결함인지 런타임 결함인지 **판별하지 못했다** | t23이 명시적으로 미확인으로 남긴 바로 그 질문. §2는 정황을 좁혔을 뿐 인과를 세우지 못했다 |
| **G3** | 실제 spawn 파라미터. 로그의 `agent_type`은 spawn의 `name`(`kanban-strategy`)을 기록하므로 **`subagent_type`이 정말 `manager-kanban`이었는지 확인 불가** | 정의 파일이 애초에 로드되지 않았을 가능성이 배제되지 않는다. 이게 사실이면 §2 전체가 무관해진다 |
| **G4** | 스폰 프롬프트 원문 | 프롬프트가 이미 형식을 지정했는지 여부에 따라 "계약 부재" 진단의 무게가 달라진다 |
| **G5** | **중첩 spawn 삼킴 여부 — 추론만 했고 검증하지 않았다** (아래 별항) | |
| G6 | 응답 본문이 진짜 빈 문자열이었는지 vs 전달 실패였는지. transcript 미판독 | "안 냈다"와 "냈는데 안 왔다"는 완전히 다른 결함이다 |
| G7 | 무보고로 끝난 다른 `Agent` 미보유 에이전트와의 대조군 | `Agent` 보유가 유의미한 변수인지 확인 불가 |

### G5 상술 — 깊이 가드와 중첩 spawn (요청 3항)

**검증한 것**: 가드는 실재한다. `internal/template/manager_kanban_depth_test.go`가 세 절을 강제한다 (헤더 주석 3-28행): (1) manager-kanban이 `tools:`에 `Agent`를 든 **유일한** 에이전트, (2) 나머지 retained 에이전트는 모두 `Agent` 누락, (3) `leaf_of: manager-kanban` 선언 파일은 `Agent` 누락 필수. 위반 시 sentinel `DEPTH_SEAL_VIOLATION`.

**검증하지 않은 것**: 중첩 spawn이 실제로 발생했는지, 그리고 그것이 보고를 삼켰는지. 다음은 **추론이며 증거가 아니다**:

- 구조적으로는 삼킴이 가능하다. leaf worker의 반환은 manager-kanban의 컨텍스트로 들어가지 부모로 가지 않는다. manager-kanban이 그 뒤 빈 턴으로 끝나면 하위 트리 전체가 소실된다.
- 하지만 같은 세션 로그의 `agent_type: None` 종료 이벤트들(12:30:40, 12:40:30 등)을 `kanban-strategy`의 자식으로 **귀속할 수 없다**. 부모-자식 관계를 나타내는 필드가 로그에 없다.
- 12:30:40의 `None` 종료는 kanban-strategy 종료 17초 **전**이지만, 이것만으로 자식이라 단정할 수 없다.

따라서 C6은 가설로 남긴다. 판별하려면 transcript에서 실제 `Agent` 호출 유무를 봐야 한다.

---

## 5. Residual-risk

1. **관측된 2회는 독립 2회가 아니다** (E3-2). 같은 `agent_id`의 연속 두 턴이므로 재현 횟수는 실질 1건에 가깝다. "2회 재현"을 근거로 결정적 결함이라 결론 내리면 과대 해석이다.
2. **§2 전체가 무관해질 단일 분기**: G3 — spawn이 실제로 `manager-kanban` 정의를 로드하지 않았다면 정의 판독은 통째로 헛다리다. 이걸 먼저 확인하는 게 비용 대비 판별력이 가장 높다.
3. **계약 부재는 필요조건이지 충분조건이 아니다.** 계약 없는 retained 에이전트가 9개인데 나머지 8개에서 같은 증상이 보고되지 않았다 — 다만 이건 **보고되지 않았을 뿐 측정된 바 없다**(G7). 계약 부재를 단독 원인으로 지목하면 §1 C7의 미판별 상태를 어긴다.
4. **아래 §6 처방은 원인을 고치지 않는다.** 원인 미상 상태에서의 **손실 완화책**이다. 파일 산출물을 요구해도 에이전트가 파일을 쓰기 전에 빈 턴으로 끝나면 똑같이 전량 소실이다 — 회수 확률을 올릴 뿐 보장하지 않는다.
5. 로그는 5분 단위 이벤트 스트림이고 본문을 담지 않는다. 본문 소실 여부의 최종 확인은 transcript로만 가능하다.

---

## 6. 권고 — 위임 회수 가능성 규율

t23이 후보로 지목한 "분석 위임에도 반드시 파일 산출물 요구"를 축으로 하되, 그것만으로는 부족한 지점을 함께 적는다.

### R1 (핵심) — 분석 위임의 산출물은 파일로 지정한다

스폰 프롬프트에 **절대 경로 + "이 파일이 진짜 산출물"** 을 못 박는다. 본 조사 위임이 이미 이 형태였고, 그래서 이 문서가 남았다.

```
DELIVERABLE — mandatory. 다음 파일에 findings를 써라:
  .moai/reports/backlog/<card>-<topic>.md
최종 채팅 응답은 5줄 요약 + 파일 경로. 파일이 진짜 산출물이니 끝내기 전에 써라.
```

**이득**: 무보고 종료가 전량 소실에서 부분 소실로 내려간다. 오케스트레이터가 파일을 직접 읽어 교차 검증할 수 있다(선례 교훈 1~3번 절차가 read-only 위임에도 적용 가능해진다).

**비용 — 축소해서 적지 않는다**:
- `Explore`와 `permissionMode: plan` 에이전트(sync-auditor 포함)는 **Write 도구가 없다.** 이 규율을 전면 적용하려면 그들에게 쓰기를 주거나 예외로 빼야 하는데, 전자는 read-only 보장을 도구 제한으로 세운 설계(`worktree-integration.md` § HARD Rules, `sync-auditor.md:126`)를 깎는다. **권고 R1은 쓰기 가능 에이전트로 한정하고, read-only 위임은 R3으로 보낸다.**
- 위임마다 리포트 파일이 쌓인다. 짧은 조회성 위임까지 걸면 `.moai/reports/`가 잡음으로 찬다 — **비자명한 분석**에만 걸고 단순 조회는 제외한다.
- 파일 쓰기 턴이 하나 더 든다.
- **보장하지 않는다**: 쓰기 전에 끝나면 결과는 동일하다(§5-4).

### R2 — 에이전트 정의에 최종 산출 계약을 넣는다

`manager-kanban.md`에 `## Output Format` 절을 추가한다. 스폰 프롬프트 규율(R1)은 호출자가 매번 기억해야 하지만, 정의의 계약은 항상 적용된다. plan-auditor의 실패 경로 규정(`plan-auditor.md:452`)이 본받을 형태다 — **무보고 종료를 정의상 허용하지 않는 것**이 핵심이다.

비용: 정의 파일 편집이므로 Template-First 규율(템플릿 미러 + `make build`)을 탄다.

### R3 — 역할 미해소 시의 폴백을 명시한다 (E5 대응)

manager-kanban의 blocker-report 트리거 목록(`:62`)에 한 항을 더한다: **"위임된 작업이 Role A 진입 조건도 Role B 진입 조건도 만족하지 않을 때 — 무엇을 위임받았고 왜 어느 역할에도 안 맞는지 blocker report로 반환."** read-only 위임의 회수 경로도 여기에 얹힌다(파일을 못 쓰는 에이전트라도 blocker report는 본문으로 낼 수 있다).

이건 G2가 열려 있는 동안에도 안전한 편집이다 — 원인이 런타임이더라도 손해가 없다.

### R4 (먼저 할 것) — 판별을 위한 단일 측정

처방을 확정하기 전에 G3을 닫는 것이 비용 대비 판별력이 가장 크다: 사고 세션의 transcript에서 (a) `subagent_type`이 실제로 `manager-kanban`이었는지, (b) 응답 본문이 빈 문자열이었는지, (c) 중첩 `Agent` 호출이 있었는지 셋을 확인한다. 이 셋이 정해지면 G2(정의 대 런타임)가 대부분 갈린다.

---

*작성: read-only 조사. 파일 수정 없음. 서브에이전트 미생성. 재현 미시도.*
