# SPEC-RECEIPT-REUSE-001 — Acceptance Criteria

## §A 규율과 트리 핀

모든 기준은 2칸 규율(`.claude/rules/moai/development/verification-completeness.md` §2)을 따른다: 관측된 RED-now 칸 + 이를 뒤집는 마일스톤을 명명하는 GREEN 칸. 문서 단위 트리 핀: **`f97edcc55`** (plan 저작 시점, 브랜치 `WT-receipt-reuse`, 워크트리 `.moai/worktrees/t1562`, status clean). §B 원장 가운데 A-D의 4건은 문서 핀 `f97edcc55`의 저작 턴에서, E-H의 4건은 plan-audit iter1 수리 라운드에서 트리 `77fc999c5`(당시 HEAD)로 실측했다 — 두 트리 사이 코드 면 변화는 없다(델타는 plan 산출물 문서 커밋뿐). **행위적 RED(AC-RR-001과 AC-RR-009의 재현 실패 출력)는 M1에서 관측되어 progress.md §E.2에 기록된다** — 그 전까지 두 기준의 RED 칸은 §B의 존재론적 RED(테스트 부재 · 원인 상수 부재 · 결함 존재)로 대리 증명되며, 이 대리는 acceptance의 명시적 Gap이다. 결함-존재 셀(LEDGER-RRR-C)이 붉은 이유는 결함이 존재하기 때문이다(wrong-reason red 아님 — 수리가 guard.go:116 경계 동작을 대체할 때만 뒤집힌다). 보존-행동 기준(AC-RR-003/005/008)의 시작 관측은 **현존 테스트의 GREEN 실측**(LEDGER-RRR-D)이고, GREEN 칸은 "M2/M3 이후에도 GREEN 유지"이다. 모든 신규 테스트 기반 판정은 swept-count 우선(`-list`로 정확 개수 확인 후 `-run`)을 따른다.

## §B RED-Now Evidence Ledger

```
LEDGER-RRR-A
  cmd:  go test -list 'TestSubagentStop_SequentialAuditorReceiptReuse' ./internal/hook
  out:  ok  	github.com/modu-ai/moai-adk/internal/hook	0.812s
  exit: 0
  why:  패키지 전수 상관 — 선택자에 해당하는 테스트 0개(목록 비어 있음):
        재현 테스트는 미창출이며 M1이 창출한다. 목록-0 공회전 가드
        (verification-completeness §1.1 — 비어 있는 스윕은 통과로 읽지 않는다).
  tree: f97edcc55

LEDGER-RRR-B
  cmd:  grep -n "CauseReceiptReused" internal/auditreceipt/store.go internal/hook/audit_receipt_guard.go
  out:  (출력 없음)
  exit: 1
  why:  재사용 전용 거부 원인 상수가 양쪽 파일 어디에도 없다 — REQ-RR-003이
        요구하는 구별 원인은 M2가 창출한다.
  tree: f97edcc55

LEDGER-RRR-C
  cmd:  grep -n "if _, err := auditreceipt.ReadStartMarker" internal/hook/audit_receipt_guard.go
  out:  116:		if _, err := auditreceipt.ReadStartMarker(g.store, key); err == nil {
  exit: 0
  why:  결함-존재 셀 — 카드가 지목한 keep-earliest 조기 반환이 트리에 있다.
        붉은 이유는 결함의 존재이며, 수리가 이 경계 동작을 대체할 때만
        뒤집힌다(AC-RR-001의 GREEN이 기대하는 변화).
  tree: f97edcc55

LEDGER-RRR-D
  cmd:  go test -count=1 -run '^(TestSubagentStop_ConcurrentBackgroundAuditorsShareEraAnchor|TestSubagentStop_BackgroundSpawnRecycledReceiptStillRefused|TestSubagentStop_BackgroundSpawnAuditorPassIsProvable|TestSubagentStop_AgentIDFailKeepsBackgroundMarker|TestAuditReceiptGuard_NonRequiredTreesAreInert)$' ./internal/hook
  out:  ok  	github.com/modu-ai/moai-adk/internal/hook	1.246s
  exit: 0
  why:  보존-행동 기준(AC-RR-003/005/008)의 시작 관측 — 5개 현존 테스트가
        저작 시점 트리에서 GREEN. 스윕 개수 5 (앵커화된 선택자 각 1개, 전부
        존재 — 앵커화 재측정은 같은 저작 턴 안에서 이뤘고, 선행 비앵커 형태의
        1차 관측(1.227s)도 동일 판정이었다).
  tree: f97edcc55

LEDGER-RRR-E
  cmd:  go test -list '^TestSubagentStop_FirstStopBlockThenOwnReceiptProvable$' ./internal/hook
  out:  ok  	github.com/modu-ai/moai-adk/internal/hook	1.067s
  exit: 0
  why:  패키지 전수 상관 — 앵커화 선택자 0-일치(목록 비어 있음): AC-RR-004의
        신규 테스트는 미창출이며 M1이 창출한다. (iter1 D1 수리 — plan-audit
        반려된 prospective 칸을 실측으로 채움.)
  tree: 77fc999c5

LEDGER-RRR-F
  cmd:  go test -list '^TestSubagentStop_SequentialOwnReceiptStillProvable$' ./internal/hook
  out:  ok  	github.com/modu-ai/moai-adk/internal/hook	0.688s
  exit: 0
  why:  패키지 전수 상관 — 앵커화 선택자 0-일치: AC-RR-006의 신규 테스트는
        미창출이며 M1이 창출한다. (iter1 D1 수리.)
  tree: 77fc999c5

LEDGER-RRR-G
  cmd:  go test -list '^TestSubagentStop_ForegroundSequentialUnchanged$' ./internal/hook
  out:  ok  	github.com/modu-ai/moai-adk/internal/hook	0.659s
  exit: 0
  why:  패키지 전수 상관 — 앵커화 선택자 0-일치: AC-RR-007의 신규 테스트는
        미창출이며 M1이 창출한다. (iter1 D1 수리.)
  tree: 77fc999c5

LEDGER-RRR-H
  cmd:  go test -list '^TestSubagentStop_ReceiptBoundaryAmbiguitySemantics$' ./internal/hook
  out:  ok  	github.com/modu-ai/moai-adk/internal/hook	0.926s
  exit: 0
  why:  패키지 전수 상관 — 앵커화 선택자 0-일치: AC-RR-009(iter1 D2의 M2
        의미 못박기)의 신규 테스트는 미창출이며 M1이 창출한다.
  tree: 77fc999c5
```

## §C Acceptance Criteria

| ID | REQ | Given / When / Then | RED (시작 관측) | GREEN (뒤집는 마일스톤) |
|----|-----|---------------------|-----------------|--------------------------|
| AC-RR-001 | REQ-RR-001 | Given 같은 세션+역할의 선행 감사 인스턴스가 단일-생존(종료 시 생존 개시 1개)으로 종료했고, 후행 인스턴스가 감사 도구 호출 없이 그 선행 시절 영수증 r1만 인용한 PASS로 끝날 때, when SubagentStop 가드가 판정하면, then PASS는 거부된다(첫 정지 block, 재진입 시 거부 지속) — 네 종료 형태 팔 전부: (a) 수락 PASS로 종료, (b) 첫 정지 차단 뒤 재진입 거부로 종료, (c) FAIL 판정으로 종료, (d) 판정 줄 없이 재진입 종료. | A(테스트 부재) + **M1 행위적 RED**(현재 네 팔 모두 수락 → 재현 테스트 적색, 원문 출력 §E.2 기록) + C(결함 존재) | swept-count 우선: `go test -list '^TestSubagentStop_SequentialAuditorReceiptReuseIsRefused$' ./internal/hook` 1개 상관, THEN `go test -run '^TestSubagentStop_SequentialAuditorReceiptReuseIsRefused$' ./internal/hook` exit 0 — 네 팔 전부 (M2) |
| AC-RR-002 | REQ-RR-003 | Given AC-RR-001의 재사용 거부가 지속될 때, when 거부 기록을 읽으면, then 원인이 기존 원인 어휘와 구별되는 재사용 전용 상수로 기록되고, phase-entry 세 스폰 각각 — `manager-develop`·`manager-docs`·`manager-git` — 이 `AUDIT_RECEIPT_VIOLATION` 접두사로 거부되며, 자격 영수증을 인용한 PASS가 그 거부를 해제한다(세 에이전트 모두 재허용). | B(원인 상수 부재) | 재사용 상수명으로 `grep -n` ≥ 1 (store.go의 원인 상수 선언), THEN 재현 테스트의 거부-기록·스폰-거부(세 에이전트 파라미터화)·해제 arm exit 0 (M2) |
| AC-RR-003 | REQ-RR-004 | Given 한 세션+역할에 동시 살아 있는 두 배경 감사 인스턴스가 있고 각자 자신이 주조한(앵커 시작 이후의) 상이한 영수증을 인용할 때, when 각자의 stop이 판정되면, then 둘 다 수락된다 — 수리가 동시성 증명 가능성을 깨지 않는다. | D (해당 1개 GREEN 시작 관측 — `ConcurrentBackgroundAuditorsShareEraAnchor`) | 문장 변경 없이 `go test -run '^TestSubagentStop_ConcurrentBackgroundAuditorsShareEraAnchor$' ./internal/hook` exit 0 유지 (M2·M3 재관측) |
| AC-RR-004 | REQ-RR-005 | Given 감사 인스턴스가 첫 stop에서 차단됐을 때(표식 유지), when 인스턴스가 계속 진행해 영수증을 주조하고 그것을 인용한 PASS로 다시 끝나면, then 수락된다 — 첫 정지 연속성이 보존된다. | E (0-일치 실측 §B) + **M1 행위적 RED**(§E.2 기록) | swept-count 우선 `-list` 1개, THEN `-run` exit 0 (M2) |
| AC-RR-005 | REQ-RR-006 | Given 감사자가 자기 시작 경계 이전에 주조된 영수증을 인용할 때, when stop이 판정되면, then `CauseReceiptBeforeStart` 원인으로 거부된다 — 수리 전과 동일. | D (해당 1개 GREEN 시작 관측 — `BackgroundSpawnRecycledReceiptStillRefused`) | 문장 변경 없이 `-run '^TestSubagentStop_BackgroundSpawnRecycledReceiptStillRefused$'` exit 0 유지 (M2·M3 재관측) |
| AC-RR-006 | REQ-RR-002 | Given 선행 인스턴스가 종료한 뒤 시작한 후행 인스턴스가 **자기 수명 동안 주조한** 영수증만 인용할 때, when stop이 판정되면, then 수락된다 — 수리가 자기 영수증 증명을 막지 않는다(오탐 펜스). | F (0-일치 실측 §B) + **M1 행위적 RED**(§E.2 기록) | swept-count 우선 `-list` 1개, THEN `-run` exit 0 (M2) |
| AC-RR-007 | REQ-RR-002 | Given agent_id를 실은 전경 인스턴스 2가(자기 표식 소유) 선행 전경 인스턴스 1의 영수증을 인용할 때, when stop이 판정되면, then before-start 원인으로 거부되고, 자기 영수증 인용은 수락된다 — 전경 경로는 수리로 변하지 않는다. | G (0-일치 실측 §B) + **M1 행위적 RED**(§E.2 기록) | swept-count 우선 `-list` 1개, THEN `-run` exit 0 (M2) |
| AC-RR-008 | REQ-RR-007 | Given codex gate가 `required`가 아닌 트리(off/advisory/무설정)일 때, when 감사자 start·stop이 발생하면, then 저장소 쓰기도 거부도 없다 — 수리로 변하지 않는다. | D (해당 1개 GREEN 시작 관측 — `NonRequiredTreesAreInert`) | 문장 변경 없이 `-run '^TestAuditReceiptGuard_NonRequiredTreesAreInert$'` exit 0 유지 (M2·M3 재관측) |
| AC-RR-009 | REQ-RR-001, REQ-RR-004 | Given 동일한 사건열(세션 S+역할 R: A 시작 t0, B 시작 t1, B가 r 주조 t2, 종료 stop t3)에 대해, when 종료 t3 시점의 생존 개시 수가 1개면(단일-생존 — 종료가 특정 인스턴스에 귀속) 이후 r 인용 stop은 거부되고(경계 전진), when 2개면(겹침 — 모호 종료) 이후 r 인용 stop은 기존 경계(앵커)로 판정된다(수락 — 경계 동결, 동시성 의미 수리 전과 등가). | H (0-일치 실측 §B) + **M1 행위적 RED**(단일-생존 팔이 현재 수락 → 적색, §E.2 기록) | swept-count 우선: `go test -list '^TestSubagentStop_ReceiptBoundaryAmbiguitySemantics$' ./internal/hook` 1개 상관, THEN `go test -run '^TestSubagentStop_ReceiptBoundaryAmbiguitySemantics$' ./internal/hook` exit 0 — 두 팔 대조 전부 (M2) |

## §D Traceability (AC → REQ)

| REQ | Covered by | Coverage |
|-----|------------|----------|
| REQ-RR-001 | AC-RR-001, AC-RR-009 | 2 (네 종료 형태 팔 + 단일-생존/겹침 경계 대조 의미) |
| REQ-RR-002 | AC-RR-006, AC-RR-007 | 2 (후행 인스턴스 자기 영수증 / 전경 경로 불변) |
| REQ-RR-003 | AC-RR-002 | 1 (원인 구별 + 스폰 거부 + 해제) |
| REQ-RR-004 | AC-RR-003, AC-RR-009 | 2 (동시 상이 인용 보존 + 모호 종료 경계 동결) |
| REQ-RR-005 | AC-RR-004 | 1 (첫 정지 → 계속 → 자기 영수증 증명) |
| REQ-RR-006 | AC-RR-005 | 1 (개시 이전 재활용 펜스) |
| REQ-RR-007 | AC-RR-008 | 1 (불-required 트리 불개입) |

7 REQ 전부 커버, 9 AC 전부 REQ 귀속 — 양방향 100%. 고의 탈락(descoped): 동시 인스턴스의 **교차 인용**(A가 B의 영수증을, B가 A의 것을 인용 — 양쪽 감사가 실제 실행됐으므로 게이트 목적은 충족; §E 잔여 위험)과 **stop 이벤트 없이 소멸한 선행 인스턴스** 경로(§E 잔여 위험).

## §E Edge Cases와 잔여 위험

- **동시 교차 인용**: A가 r2(B 몫), B가 r1(A 몫)을 인용 — 양쪽 감사가 실제 실행된 트리에서의 귀속 교차다. 순차 재사용 결함(카드 본체)과 구별되며 이 SPEC의 범위 밖 — 잔여 위험으로 기록하고, 종료-이벤트 경계 기제(plan §D.3 후보 1)에서 자연히 어디까지 막히는지 M2 관측으로 확인한다.
- **무-정지 소멸 선행**: 선행 인스턴스가 stop 이벤트 없이 사망하면 종료 기록이 없어 후행 인스턴스의 재사용 경로가 남는다(후보 1 기준) — 잔여 위험. 후보 2(소비 원장)는 수락-선행 경로에서만 보조한다.
- **같은 세션 장수**: 세션 id는 재사용되지 않는다는 관측 위에 세션-귀속 기록을 둔다 — 별도 TTL 없음(decision-index Q3). 세션 id 재사용이 관측되면 blocker.
- **재진입(StopHookActive) 재인용**: 차단→재진입 stop에서 같은 영수증 재인용은 첫 차단이 소비가 아니므로 문제없다(연속성 REQ-RR-005).
- **품질 게이트 기준(DoD)**: AC-RR-001..009 전부 GREEN + 영향 패키지 `-count=1` GREEN + `go vet`·golangci-lint 신규 0 + E1-E8 기록 + RED 원문이 progress.md §E.2에 남아 있다.
