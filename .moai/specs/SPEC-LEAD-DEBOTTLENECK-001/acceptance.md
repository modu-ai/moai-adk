# acceptance.md — SPEC-LEAD-DEBOTTLENECK-001

검증 계층. 모든 AC는 `AC-XXX` 라벨의 Given-When-Then 형식이며 이진 판정 가능해야 한다. 요구사항 계층(GEARS)은 spec.md §3가 소유한다 — 여기에 GEARS를 재진술하지 않는다.

## §D.0 RED-now 기준 (2026-08-26, t283 워크트리 HEAD `175d63f3f`에서 실측)

| 항목 | 명령 | 관측 출력 | 의미 |
|---|---|---|---|
| 도구 부재 | `grep -c "SendMessage\|ListAgents" .claude/agents/moai/manager-lead.md` | `0` | AC-001 RED-now — 오늘 deputy는 발송 자체를 시도할 수 없음 |
| deputy 절 부재 | `grep -ci "deputy" .claude/agents/moai/manager-lead.md` | `0` | AC-002 RED-now |
| 교리 부재 | `grep -ci "deputy" .claude/rules/moai/workflow/kanban-dispatch.md .claude/rules/moai/workflow/kanban-dispatch-detail.md` | `:0` 각각 | AC-008 RED-now |
| depth seal baseline | `go test ./internal/template/ -run 'TestManagerLeadIsSoleAgentCarrier|TestManagerLeadCarriesAgent|TestNoNestedLeafWorkerCarrier' -count=1` | `ok github.com/modu-ai/moai-adk/internal/template 3.305s` | AC-006 green-baseline — 구현 후 동일 green이 불변식 |
| mirror 동일성 baseline | `diff -q .claude/agents/moai/manager-lead.md internal/template/templates/.claude/agents/moai/manager-lead.md` | (출력 없음, exit 0) | 구현 후에도 local==mirror여야 함 (AC-007) |
| 런타임 SendMessage 존재 | `orchestration-mode-selection.md:124` 기록 인용 | "SendMessage present" | AC-004의 green path가 런타임에 의해 차단되지 않음의 근거 |

## §D AC Matrix

**AC-LDB-001** (M1, MUST, REQ-LDB-001) — Given 사전구현 트리에서 `grep -c "SendMessage" .claude/agents/moai/manager-lead.md`가 0을 반환하고, When M1이 `tools:` 허용목록에 `SendMessage, ListAgents`를 추가하면, Then 동일 grep이 각각 1 이상을 반환하고 허용목록이 단일 CSV 문자열로 유효 YAML을 유지한다.

**AC-LDB-002** (M1, MUST, REQ-LDB-002) — Given agent 본문에 "Deputy dispatch surface" 절이 없고(grep `-i deputy` = 0), When M1이 절을 추가하면, Then 해당 절이 spec.md §4 매트릭스의 위임 항목 5종(디스패치 발송+형상검증 / CI watch 보고 / CodeRabbit 2조건 판독 보고 / 1차 증거 읽기+권고 / 요약 보고)과 유지 항목을 코드화하고, `grep -c "Deputy dispatch surface" .claude/agents/moai/manager-lead.md`가 1을 반환한다.

**AC-LDB-003** (M1, MUST, REQ-LDB-003) — Given 유지 권한 금지가 grep 불가능한 산문이면 뮤턴트가 통과하고, When M1이 `DEPUTY-RETAINED-BY-LEAD` 리터럴 마커로 유지 권한 6종(최종 머지 승인 / 운영자 게이트 / `moai todo` 변이 / 최종 PASS-FAIL / CodeRabbit 규율 판정 / 크로스세션 분쟁 조율)을 열거하면, Then `grep -c "DEPUTY-RETAINED-BY-LEAD" .claude/agents/moai/manager-lead.md`가 1 이상이고 각 항목이 마커 하위에 존재한다.

**AC-LDB-004** (M3, MUST, REQ-LDB-001/004) — 런타임 프로브. Given M3 시나리오 수행자(리드 세션)가 시나리오 시작 시 named 테스트 컴패니언 세션을 직접 기동해 두고(컴패니언 기동은 시나리오 프로토콜의 첫 단계 — deputy가 아닌 리드가 수행), When 리드 세션이 UNNAMED 백그라운드 manager-lead deputy를 spawn하여 named 세션으로 `SendMessage`를 발송하게 하면, Then deputy의 반환 보고에 발송 결과 형상이 관측되고 `routing` 객체가 없음이 기록된다. 발송 결과에 `routing` 객체가 나타나면 유실로 간주하고 `name [ref]` 재발송이 뒤따른다(AC-005). RED-now: 오늘은 tools 부재로 발송 시도 자체가 불가(§D.0 1행) — 이 AC는 현재 실패 상태이며 M1 후 통과 가능해진다.

**AC-LDB-005** (M1, MUST, REQ-LDB-004) — Given 디스패치가 in-process mailbox로 유실될 수 있고(routing 객체), When deputy가 발송 결과에서 `routing` 객체를 관측하면, Then agent 본문이 `name [ref]` 재발송 프로토콜을 지시함을 `grep -c "routing" .claude/agents/moai/manager-lead.md` ≥ 1과 해당 절의 재발송 문구로 확인한다.

**AC-LDB-006** (M1, MUST, REQ-LDB-010) — Given baseline에서 depth-seal 가드가 green이고(§D.0), When 구현이 tools를 추가한 뒤, Then `go test ./internal/template/ -run 'TestManagerLeadIsSoleAgentCarrier|TestManagerLeadCarriesAgent|TestNoNestedLeafWorkerCarrier' -count=1`이 `ok`를 반환하고, `.claude/agents/moai/**/*.md` 전수에서 `Agent`를 `tools:`에 가진 파일이 manager-lead.md 유일임을 grep으로 확인한다.

**AC-LDB-007** (M2, MUST, REQ-LDB-009) — Given 분배 표면 3종(agent 1 + 규칙 2)이 편집되고, When M2가 완료되면, Then (a) `diff` 각 쌍이 **금지 토큰 행을 제외한 나머지에서 동일** — 금지 토큰(SPEC ID·REQ 토큰·감사 인용·내부 일자) 행만 로컬 사본이 추가로 가질 수 있으므로 동일성 판정은 modulo 금지 토큰 (b) `make build` 재생성 후 `bin/moai` 타임스탬프 갱신 또는 빌드 성립 (c) 템플릿 사본에서 `grep -rc "SPEC-LEAD-DEBOTTLENECK\|REQ-LDB" internal/template/templates/` = 0 (중립성 — 로컬 사본은 SPEC 출처를 가질 수 있으나 미러는 중립화).

**AC-LDB-008** (M2, MUST, REQ-LDB-007) — Given 기존 `kanban-dispatch.md`의 [HARD] 절 집합이 baseline이고, When deputy 표면이 추가된 뒤, Then `[HARD]` 출현 수가 baseline 이상이고 기존 절 토큰 전부(`The lead is the queue's sole producer`, `Completion is read, never trusted`, `the verdict's home` 등 핵심 문구)가 잔존하며, `grep -ci "deputy" kanban-dispatch.md` ≥ 1 (RED-now = 0).

**AC-LDB-009** (M3, MUST, 카드 검증 시나리오) — 2+ 레인 동시 진행 시나리오에서 리드 턴 점유 감소 실측. Given 동일한 조율 작업량(2개 레인 디스패치 + 2개 PR의 CI 종착 + CodeRabbit 2조건 판독 + 레인 응답 1차 처리)을 (a) 구현 전 리드 세션이 (b) 구현 후 deputy가 수행하는 두 실행의 리드 세션 transcript(`~/.claude/projects/<hash>/<session>.jsonl`)에서, When 조율 툴콜(`SendMessage` 발송, `gh pr checks`/`gh api` CodeRabbit 폴링)을 포함한 assistant 턴 수를 `grep -c` 계수로 세면, Then (b)의 리드 세션 조율 턴 수가 (a) 대비 50% 미만으로 감소한다. "체감 개선"은 판정 근거가 아니다.

**AC-LDB-010** (M3, MUST, REQ-LDB-003, 카드 검증 시나리오) — 머지 승인 누락 0 + 운영자 게이트 우회 0. Given M3 시나리오 프로토콜이 (i) 시나리오 로그의 모든 머지 행위를 `git merge` 커맨드 토큰으로 계수하고 (ii) 리드 세션이 각 승인을 **고정 토큰** `LEAD-MERGE-APPROVED <PR-번호-또는-SHA>` 형식으로 transcript에 기록하도록 요구하며, When 시나리오가 완료되면, Then (a) 승인 기록 수 ≥ 머지 행위 수 (`grep -c 'git merge' <시나리오로그>` ≤ `grep -c 'LEAD-MERGE-APPROVED' <lead>.jsonl` — 토큰이 프로토콜에 고정되므로 ko-lead transcript에서도 결정론적이며, 머지 0건과 '승인 없는 머지'를 구분한다) (b) subagent 맥락에서 발행된 `AskUserQuestion` 호출 수 = 0 (transcript grep `"name":"AskUserQuestion"`의 subagent 파일 한정).

**AC-LDB-011** (M1..M3 전체, MUST, REQ-LDB-011) — Given 구현이 완료되고, When `git diff --stat <base>..<head>`를 검사하면, Then `internal/`, `pkg/`, `cmd/` 하위 Go 경로가 diff에 0회 등장한다.

**AC-LDB-012** (M3, MUST, REQ-LDB-003, REQ-LDB-006) — 금지 행위 뮤턴트의 기계적 FAIL 검사. Given M3 시나리오 동안 deputy가 생성한 transcript가 존재하고, 프로토콜이 deputy의 금지 출력 토큰(`FINAL VERDICT:`)과 권고 전용 토큰(`RECOMMEND:`)을 고정하며, When 아래 계수를 실행하면, Then 전부 0: (i) `grep -c '"moai todo' <deputy>.jsonl` — 큐 변이 시도 0 (ii) `grep -c 'FINAL VERDICT:' <deputy>.jsonl` — 최종 판정 토큰 0. 0이 아닌 값은 뮤턴트 A의 기계적 FAIL이다.

**AC-LDB-013** (M3, MUST, REQ-LDB-012) — 동시 write-capable 단독성. Given 시나리오 로그가 deputy 활성 구간과 레인 커밋 창을 기록하고, When deputy transcript와 시나리오 로그를 검사하면, Then (i) deputy의 Write/Edit 대상 경로 중 `.moai/reports/`·`.moai/state/` 밖 = 0 (`grep -o '"file_path":"[^"]*"' <deputy>.jsonl | grep -vc '\.moai/\(reports\|state\)/'`) (ii) deputy 활성 구간과 write-capable 레인 커밋 창이 겹치는 기록 = 0.

### 뮤턴트 방어 (채택 전 탐침)

- **뮤턴트 A (금지 행위 수행 deputy)**: deputy가 머지, 큐 변이 또는 최종 판정을 수행하는 시나리오 — 교리상 방어선은 AC-003의 grep-able `DEPUTY-RETAINED-BY-LEAD` 마커 + REQ-LDB-006의 거부·블로커 반환 지시이고, **기계적 검사**는 AC-012가 담당한다(`"moai todo` / `FINAL VERDICT:` 계수 — 0이 아니면 FAIL). 마커만 있고 계수가 없으면 이 뮤턴트가 통과하므로, AC-003은 마커의 존재를, AC-012는 위반의 부재를 각각 기준으로 삼는다.
- **뮤턴트 B (유실 무시 deputy)**: 발송 결과를 읽지 않고 성공으로 보고 — AC-005가 `routing` 재발송 프로토콜의 존재를, AC-004가 실제 발송 형상 관측을 각각 잡는다.
- **뮤턴트 C (판정 승격 deputy)**: CodeRabbit 판독 보고를 판정으로 대신 수행 — AC-002가 매트릭스의 "판독·보고" vs "판정" 분리를 코드화했음을 확인하고, AC-010(b)가 게이트 우회 0을 잡는다.

## §D.1 심각도·추적·종결 관문

| AC | 심각도 | 추적 REQ | 종결 관문 |
|---|---|---|---|
| AC-001 | MUST | REQ-001 | M1 |
| AC-002 | MUST | REQ-002 | M1 |
| AC-003 | MUST | REQ-003, REQ-006 | M1 |
| AC-004 | MUST | REQ-001, REQ-004 | M3 (M1 후 실행 가능) |
| AC-005 | MUST | REQ-004 | M1 |
| AC-006 | MUST | REQ-010 | M1 |
| AC-007 | MUST | REQ-009 | M2 |
| AC-008 | MUST | REQ-007, REQ-008 | M2 |
| AC-009 | MUST | REQ-005, REQ-008, 카드 검증 지시 | M3 |
| AC-010 | MUST | REQ-003, 카드 검증 지시 | M3 |
| AC-011 | MUST | REQ-011 | 전 마일스톤 closure |
| AC-012 | MUST | REQ-003, REQ-006 | M3 |
| AC-013 | MUST | REQ-012 | M3 |

## §D.2 정의된 측정 명령 요약 (AC-009/010 — 판정 반증 가능 형태)

```bash
# 리드 세션 조율 턴 계수 (before/after 동일 명령)
grep -c '"name":"SendMessage"\|"gh pr checks"\|"gh api.*status"' <lead-session>.jsonl

# 머지 행위 vs 승인 기록 (AC-010: approvals >= merges — 고정 토큰)
grep -c 'git merge' <scenario-log>.jsonl            # 머지 행위 수
grep -c 'LEAD-MERGE-APPROVED' <lead-session>.jsonl  # 승인 기록 수 (프로토콜 고정 토큰)

# deputy 금지 행위 계수 (AC-012: 전부 0이어야 PASS)
grep -c '"moai todo' <deputy>.jsonl                 # 큐 변이 시도
grep -c 'FINAL VERDICT:' <deputy>.jsonl              # 최종 판정 토큰 (권고는 RECOMMEND: 만)

# subagent 발행 AskUserQuestion = 0 확인 (AC-010b)
grep -c '"name":"AskUserQuestion"' <subagent-session>.jsonl

# deputy 쓰기 표면 (AC-013i: 보고/상태 경로 밖 = 0)
grep -o '"file_path":"[^"]*"' <deputy>.jsonl | grep -vc '\.moai/\(reports\|state\)/'
```

정의상 unfalsifiable한 AC("빨라졌다")는 채택 금지 — 위 계수가 기준이다. 토큰 `LEAD-MERGE-APPROVED` / `FINAL VERDICT:` / `RECOMMEND:`는 M3 시나리오 프로토콜에 고정되므로 grep이 ko-lead transcript에서도 결정론적이다.

## §D.3 간접 검증·품질 관문

- AC-007의 중립성 grep은 템플릿 트리 전체 한정(로컬 제외)으로 스캔 범위를 명시한다 (부재 주장은 훑은 경로 한정).
- 모든 grep 기준은 사전구현 트리(base 0히트) 조건을 §D.0 RED-now 표로 이미 충족했다.
- 종결(Definition of Done): AC-001..013 전부 PASS + §D.0 baseline 표의 green 항목(depth seal, mirror 동일성)이 구현 후에도 동일 명령으로 green + 템플릿 중립성 0히트 + Go diff 0.
