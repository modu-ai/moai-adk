---
id: SPEC-AGENTIC-FACTORY-001
title: "AgenticFactory 실행 계약 — GitHub Issue 원장 기반 소유권·전이·안전 게이트"
version: "0.6.7"
status: draft
created: 2026-10-29
updated: 2026-10-30
author: GOOS
priority: P0
phase: "v3.3.0 target"
module: "internal/agenticfactory"
lifecycle: spec-anchored
tags: "agentic-factory, github-issues, control-ref, ownership, lease, policy, intake, context-ledger, rate-limit, checkpoint, contract"
tier: L
issue_number: 1948
related_specs: [SPEC-ASIDE-BROWSER-001, SPEC-RESOURCE-SLOT-LEASE-001, SPEC-FACTORY-LEASE-RENEW-001, SPEC-AGENTIC-FACTORY-IDENTITY-001]
---

# SPEC-AGENTIC-FACTORY-001 — AgenticFactory 실행 계약

## HISTORY

| 버전 | 날짜 | 작성 | 내용 |
|---|---|---|---|
| 0.1.0 | 2026-10-29 | manager-spec | 최초 작성 (draft). 규범 원천은 GitHub Issue #1948 하나뿐이다. 조정자 추가 요구(단일 Issue 시나리오 7종·6개 결정 사항, 세션 구조·명칭·CLI 재정의, 명칭 미확정 표시·lane-22 대응표 정렬), 입력 보고서 세 건(lane-22 갭 매트릭스·이관 대응표, lane-21 브라우저 계약), 상위 설계 권고(Agentic Factory / ControlService / Coordinator / Worker, `moai factory start·status·attach`, TTL 대 epoch 교차 점검)를 같은 판에 PROPOSAL로 반영했다. |
| 0.2.0 | 2026-10-29 | manager-spec | plan-audit iteration-1 FAIL 0.69(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter1.md`)의 결함 D1~D18을 앞으로 고치는 판. **REQ ID 재사용(상한 25 유지)**: 옛 REQ-AF-002(차단·보류)는 REQ-AF-001에 합치고 REQ-AF-002는 `done` 수용 조건(원격 반영 + 로컬 동기화 + 배포 증거 또는 기록된 N/A)으로, 옛 REQ-AF-005(비밀 가림)는 REQ-AF-004에 합치고 REQ-AF-005는 관찰 모니터 표시 항목으로 다시 썼다. REQ-AF-015는 환경별 단계 실제 실행, REQ-AF-016은 3선택 선택 모델(선택·저장·다시 읽기), REQ-AF-014는 거울 쓰기 실패 규칙 하나로 정리(D2), REQ-AF-010은 활성 판정 정본을 `homestate.ClassifyOwnerWith`로 정함(D12, §F.2), REQ-AF-025는 단계별 호환 기준(D10). AC는 25개 그대로이며 시나리오 하네스 계약(D5), 증거 원장의 "대상 부재 참고 관측" 전환(D1), AC-AF-019·024 기준점(D1·D9), AC-AF-018 여섯 증거 제거 사례(D6), AC-AF-002 미전달 dispatch 사례(D7), AC-AF-020 운영 기록 점검(D8), 마일스톤 재배치(D11)를 반영했다. `factory.go:284-286`·`approve` `Use:` 문자열 사실 정정(D17), progress §E.1 기록(D18). 명칭·명령은 계속 미확정이다. |
| 0.3.0 | 2026-10-29 | manager-spec | plan-audit iteration-2 FAIL 0.69(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter2.md`)의 차단 결함 R-D5·N1~N8과 선택 O1~O3·O6·O7을 앞으로 고치는 판. REQ·AC 수는 25/25 그대로이며 새 ID를 만들지 않고 기존 문장에 합쳤다. N1: 결정 D2를 고쳐 Codex가 아닌 **공급자 중립 리뷰 경로**(독립 리뷰 세션 또는 운영자 기록 리뷰)를 정의하고 M4에 둔다(§F.7, design §11.1). N2: 증거 생산자 역할 결속(REQ-AF-018·020, design §7.2·§11). N3: 로컬 동기화를 경로별 내용 digest로(REQ-AF-002 (b), design §4). N4~N7: AC-AF-018·003·007·020 사례 보강(N7은 커밋 그래프 확인). N8: §A.3과 acceptance §E에 3선택 UI 항목 미완료 명시. R-D5: acceptance §A.1에 05→06 증거 생산 단계, 분기 사례의 격리 실행, 재실행 규칙, 두 번째 Issue의 사후 상태 추가. O6: `cc -l` 리더 검증 위치(`internal/discovery/factory_discovery.go` `verifyCandidate`) 기록. ID 재사용 안내(O8): 1회차 판정 파일의 REQ-AF-002·005는 0.1.0의 뜻(차단 기록·비밀 가림)이며 지금은 REQ-AF-001·004 안에 있다. |
| 0.4.0 | 2026-10-29 | manager-spec | plan-audit iteration-3 FAIL 0.69(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter3.md`) 뒤의 델타 수리 판. 운영자 결정(2026-10-29, 리더 경유 전달)을 **결정 DEC-AF-01**(§C.5.1, design §11.2)로 기록하고 REQ-AF-018·AC-AF-018에 합쳤다: 증거를 내는 역할은 ControlService만 운영자 인증 아래 부여하고, Worker는 자기에게도 남에게도 역할을 줄 수 없으며, 한 세션은 run 안에서 역할 하나만 갖고, 세션 신원은 ControlService가 관측한 pid·시작 시각·session id·epoch에 묶는다(D1). D2: 격리 사본을 독립 gitdir·인덱스의 별도 클론으로 하고 실제 트리 HEAD·인덱스·경로별 digest 불변을 단언. D3: AC-AF-018에 코드 리뷰·보안 리뷰·QA 각각의 Worker 생산 사례와 이전 attempt Worker의 자기 코드 리뷰 사례 추가. D4: AC-AF-020 넷째 검사를 커밋마다 값 비교로 바꾸고 `llm.yaml` 두 벌 추가. D5: AC-AF-007에 인덱스만·모드만 바뀐 변형 추가. D6: 보류 사유 `operator-hold`를 design §2.1a에 넣고 REQ-AF-001·§A.1을 맞춤. D7: AC-AF-003 ④가 `received`를 세우지 않게 고침. 선택 O1~O5·O7 반영(O6 S4 AC는 남김, OQ-20). 전 AC 변형 점검표(acceptance §D) 추가와 그에 따른 AC 보강. REQ·AC 수는 25/25 그대로이며 ID 집합은 바뀌지 않았다. |
| 0.5.0 | 2026-10-29 | manager-spec | plan-audit iteration-4 FAIL 0.76(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter4.md`)의 분할 제안과 운영자 재결정(2026-10-29, 리더 경유)에 따른 델타 판. 역할 부여·세션 신원·운영자 인증(결정 DEC-AF-01, REQ-AF-018의 해당 조항, design §11.2와 §7.2·§11·§11.1의 해당 내용, AC-AF-018 R1~R10, OQ-19)을 하위 SPEC-AGENTIC-FACTORY-IDENTITY-001로 옮기고, 그 자리에 하위 SPEC의 귀속 조회 의존과 **잠금 결정 DEC-AF-02**(§C.5.1: 하위 SPEC 완료 전에는 리뷰·감사·QA·coordinator 증거가 모두 귀속 불가로 막힘)를 넣었다. N3: AC-AF-008을 M1에서 닿는 상태 엔진 조회 수준으로 다시 썼고, 관찰 출력·종료 코드 단언은 M6의 AC-AF-022 (e)로 옮겼다(0.6.0 정정: 이 행은 원래 "AC-AF-024"로 잘못 적혀 있었다 — iteration-5 D6). 변형 점검표(acceptance §D)의 잘못 귀속된 두 행(005 하네스 행, 018 R1/R2 행)과 이관 항목 행을 고쳤다. REQ·AC 수는 25/25 그대로다. |
| 0.6.0 | 2026-10-29 | manager-spec | 정본 정합 판. 규범 원천을 Issue 본문만에서 **Issue #1948 본문 + 댓글 11개(가장 새 revision 우선)**로 넓혔다(§A.1). plan-audit iteration-5 FAIL(`.moai/reports/SPEC-AGENTIC-FACTORY-001/plan-audit-iter5.md`)의 D1~D6을 원인별로 고쳤다: (a) 잠금을 코드 수준 상수로 정하고 설정·환경 변수로는 풀 수 없게 하며 시나리오 하네스까지 덮음(DEC-AF-04, REQ-AF-011, AC-AF-011·025, plan M7) — D1~D3; (b) 귀속 결과 하나 → 게이트 사유 하나의 단일 표(design §2.1b) — D4; (c) writer 인계의 세션 변경·epoch 변경 분리(AC-AF-009·011) — D5; (d) 위 0.5.0 행 정정과 두 가지 셈 규칙(acceptance §A) — D6. 구조는 리더 결정 K.6 (마)를 따른다: 판단이 날카로운 항목(상태·소유·병합)은 묶음 REQ, 전이표·record·template·시험 manifest는 design 표(DEC-AF-10). **REQ·AC ID의 뜻이 바뀌었다**: REQ-AF-001~015 = 묶음 G01~G15, REQ-AF-016~025 = 표 T01~T10, AC-AF-0NN = 리더 배정표 K.6b의 AC NN. 옛 ID 대응은 research.md R12. 새로 넣은 것: 정본 상태 모형(READY·BUILDING·REVIEWING·MERGING·DONE·CANCELLED + PARKED, E01~E27), 권위 = 온라인 control ref + Issue(로컬 claim 저장소는 재구축 가능한 캐시), PR 인계·수리 단일 소유·PR 적체 상한, 백그라운드 접수·준비 7행, Issue별 맥락 장부·접근성 분류·정제 범주 C1~C5·증거 위치 단계, GitHub/GitLab 호출 효율·429/403 판별, Issue 쓰기 권한 정책과 협업 옵션, checkpoint 복구, 효율 측정 어휘, 명칭(Coordinator 내부 / leader 표시 / agent 실행자). REQ 25 / AC 25(상한과 같음). 이 판은 감사받지 않았다(새 감사 승인 없음). |
| 0.6.1 | 2026-10-29 | manager-spec | 조정자 후속 지시에 따른 정정 판(REQ·AC 수와 ID는 그대로). (1) 0.6.0이 §C.7에 예약으로 넣었던 운영자 결정 행 하나와 §H·progress의 해당 언급을 지웠다. 그 지시는 잘못 전달된 것이었다. 다른 DEC 번호는 다시 매기지 않았다(깨지는 참조 없음). (2) `pr_queue.wip_limit` 미설정 시 배차 정지와 checkpoint 상한 미설정 시 거절은 새 정책 선택이므로 채택하지 않고 §G.1 OD-7·OD-8(운영자 결정 대기)로 옮겼다. 옛 OQ-24는 OD-7로 옮겨 G.2에서 지웠고, OQ-22는 숫자 값 질문만 남겼다. design §4.6·§6, acceptance AC-AF-007, plan §D·§G를 맞췄다. (3) 닫힌 사유 문자열 8개를 lane-18 트리의 Go 상수와 바이트 단위로 대조해 design §2.1a에 결과를 적었다(8개 모두 일치, 문자열 수정 없음). (4) lane-18·lane-20 보고 사실과 lane-18이 밝힌 공백을 §B.2·§H에 넣었다. |
| 0.6.2 | 2026-10-29 | manager-spec | 조정자 후속 지시에 따른 정정 판(REQ·AC 수와 ID는 그대로). (1) AC-AF-012 (d)의 "`intake.max_concurrent` 미설정 → 거절"은 채택한 규칙이 아니므로 사례에서 빼고 §G.1 OD-9(운영자 결정 대기)로 옮겼다. (d)는 설정값 N에서 N+1번째 대기만 남아 사례 수는 5 그대로이고 기대 PASS 수(10 + 5)도 그대로다. design §4.2 실행 규칙과 §6 설정 표를 맞췄다. (2) lane-18 코드와 SPEC의 차이 다섯 건(§H (a)~(e))을 리더 결정에 따라 관찰로만 기록한 공백으로 고쳐 쓰고 출처를 붙였다. SPEC을 코드에 맞추거나 코드 변경을 요구하는 문장은 없앴다. (3) 리더 정정 반영: `authz.go:208`은 상수 `ReasonOpNotDefined`를 반환할 뿐이고, 따옴표 붙은 `"op-not-defined"`는 `state.go:86`에만 있으며, `authz.go:96` `"unknown-operation"`과 `:90` `"allow"`는 `AuthzVerdict` 값이다(§B.2, §H, design §2.1a). (4) collector 기준 스냅숏이 claim 문서 한도(256 KiB, 가정값)를 넘을 때의 규칙을 §G.1 OD-10(운영자 결정 대기, 선택지 A~E와 결정 입력 세 가지, 권고 없음)으로 넣고 §H 공백 (e)와 §B.2를 그쪽으로 연결했다. 요구사항·수용 기준은 만들지 않았다. (5) PARKED 전용 사유 철자를 정본 S46(댓글 6489610367) 50·72행에 맞춰 `stop-unknown` → `stop_unknown`, `start-unknown` → `start_unknown`으로 바꿨다(design §2.1a·§2.2, §H (b)). 나머지 넷(`dependency`·`unknown-effect`·`pending-effect`·`contract-change`)은 정본 48·57·60·70행과 같아 그대로다. (6) checkpoint push 재시도 횟수를 §G.1 OD-11로 넣고 design §4.6 재개 규칙의 "한 번"을 그 결정에 걸었다. (7) `operator-hold`를 PARKED 사유로 둘지를 §G.1 OD-12로 넣고, 이 값을 쓰는 design §2.1a·§8과 acceptance §A.1이 그 결정을 가리키게 했다(값과 단언은 그대로 두고 결정을 가리키는 주석만 더함). (8) design §2.3 대응표를 리더 결정('leader decision, relayed via the leader session')으로 고쳤다: 이름 기준으로 `verifying`·`ready` → `BUILDING/IMPLEMENT` + 검사 증거, `landing` → 인계 receipt가 있으면 `REVIEWING`·병합 의무 진행 중이면 `MERGING`, `blocked` → 현 단계 + `PARKED`(사유). `on-hold`는 OD-12. [0.6.3에서 증거 기준 규칙으로 대체됨 — 0.6.3 행 참조] (9) 관찰 기록을 §B.2와 §H (f)~(j)에 더했다: checkpoint 일곱 상태·K1~K4는 정본에 없음, 코드 상태 넷과 필드로 같은 정보 표현 가능(lane-18 보고), 코드 상태 모형은 계약 0.4.0 기준(리더 보고), 옛 상태 키 이전 세 단계(리더 결정, lane-18 구현 중). REQ·AC 수와 ID는 그대로다. |
| 0.6.3 | 2026-10-29 | manager-spec | 리더 정정('leader decision, relayed via the leader session')에 따른 판(REQ·AC 수와 ID는 그대로). design §2.3의 옛 상태 대응을 이름 기준에서 **증거 기준**으로 바꿨다. 정본 E03을 따라, 옛 이름과 상관없이 writer 인계 receipt와 leader 인계 수락이 둘 다 있으면 `REVIEWING`, 병합 의무가 진행 중(`merged=true`)이면 `MERGING`, 그 밖의 모든 경우(`verifying`·`ready`, 인계 증거 하나라도 없는 `landing` 포함)는 `BUILDING/IMPLEMENT` + 검사 증거다. `blocked`는 그대로 현 단계 + `PARKED`(사유), `on-hold`는 OD-12. §H (i)·(j)를 이 규칙으로 닫힌 관찰 기록으로 고쳐 쓰고(출처 유지), 규칙 문언에 없는 우선순위와 E03 guard 목록과의 관계를 §H (k)로 더했다. §B.2에 리더 보고(lane-18의 커밋되지 않은 `core/legacy.go`가 `landing`을 `HasReceipt && LeaderAccepted`로 나눔, `ready`에도 같은 규칙 적용 요청)를 미검증으로 더했고 코드 변경은 요구하지 않는다. 0.6.2 HISTORY (8)에 대체 표시를 달고, design §2.2 끝의 `landing → verifying` 문장에 증거 기준 참조를 더했다. 상태 이전을 다루는 AC·시험 줄은 acceptance.md에 없어 고칠 곳이 없었다. 같은 판에 lane-18 커밋 `af80e0bdec`의 잠정 값(번호 1·2·3·5·7·6, 4번 미전달)과 PARKED 중 단계 변경 간선 거절(리더 결정)을 §H·§B.2에 관찰로 더하고, 정본 S46의 E05·E15·E16·E18·E19·E20·E24 줄과 대조해 정본이 침묵하지 않는 곳((3)·(5)·(7))을 충돌 관찰로 적었다. (6)의 순환(코드가 감사받지 않은 SPEC 0.6.1에서 값을 가져옴)을 그대로 적었다. 권고와 코드 변경 요구는 없다. |
| 0.6.4 | 2026-10-29 | manager-spec | 리더 정정('leader correction, relayed via the leader session; the leader read S46 lines 49, 61, 65, 70 directly')을 §H·§B.2에 관찰로 기록한 판(REQ·AC 수와 ID는 그대로). 리더가 (3) E24·(5) 종결과 overlay 해제·(7) E19를 "정본 침묵"으로 잘못 분류했고 정본은 S46 70·61·65행에서 침묵하지 않는다는 정정, 그리고 리더가 lane-18에 보낸 코드 쪽 지시 네 가지(E15 정리 불명 시 현 단계 + PARKED 유지, E19 사유를 실제 오류 분류에서 파생·검증, E03 guard 증거 키가 49행 목록 전체를 덮음, E24를 정본 근거로 표시)를 리더 보고·미검증 사실로 적었다. "4번 항목 미전달" 표기가 맞다는 확인을 더했다. PARKED 중 단계 변경 간선 거절(리더 결정, 정본에 없음)은 문언 그대로 두었다. SPEC을 코드에 맞추거나 코드 변경을 요구하는 문장은 없다. |
| 0.6.5 | 2026-10-29 | manager-spec | 관찰 기록 판(REQ·AC 수와 ID는 그대로). §H·§B.2에 더한 것 두 가지. (A) lane-18 커밋 `52363f7ac7`: E03 guard가 S46 49행 아홉 항목을 덮음, E15·E16 결과별 guard, E19 사유를 `core.ClassifiedReasons` 밖이면 거절. 이 집합의 값 11개는 이 실행이 그 커밋에서 직접 읽었다. (B) 리더 결정 두 가지: 옛 상태를 진척 기록으로 남기고 `FromLegacy` 파생값을 함께 기록한다, 병합 게이트 사유(G)를 overlay에서 단계를 유지하는 게이트 기록으로 옮긴다(동작 변경: 'Resume after blocked' → '같은 단계에서 새 증거로 다시 검증'). progress.md에 착지 규칙(리더 결정) 기록 §K를 더했다. 정본 근거는 판단하지 않았고, SPEC을 코드에 맞추거나 코드 변경을 요구하는 문장은 없다. |
| 0.6.6 | 2026-10-29 | manager-spec | 관찰 기록 판(REQ·AC 수와 ID는 그대로). §H·§B.2에 설정 읽기 커밋의 계보(`6c833ee4b4`·`8920c0ecf8` → `5c278d8b51` → 병합 `7a88d26259` → 통합 HEAD)와 `wip_limit` 읽기가 OD-7을 미리 정하지 않는다는 리더 판독을 'leader direct reading of the settings-tree log and diff' 출처로 기록하고, 이 실행이 읽기 전용 git으로 따로 관측한 것(조상 관계, 병합 부모, 주석 원문, 시험 이름과 단언 범위)을 분리해 적었다. progress §K에 `8920c0ecf8`의 출처(리더 지시 메시지 (a)(b)(c)로 만들어졌고, 리더는 그 커밋을 확인한 기록이 없었음)와 잠정 키, 바뀐 파일 두 개를 더했다. |
| 0.6.7 | 2026-10-30 | manager-spec | revised, unaudited. 리더 결정(리더 세션 경유)을 기존 REQ·AC 문장에 합친 판이다. 새 REQ·AC ID는 없고 25/25 그대로다(Tier L 상한). (1) 끝까지 책임: 배정된 agent가 구현 → CI 실패 수리 → 리뷰 대응 → 정상 병합 → Issue 종료 확인까지 책임을 지고, PR 인계로 책임이 끝나지 않는다(REQ-AF-009, design §2.1·§2.2 E03·E06·§7). 정본 S46의 E03 "종결 책임 leader로"·E06 "가용 agent에 배차" 문언과의 차이는 §H 0.6.7에 적었다. (2) leader 역할 = 변경 알림·실패 복구·배정 조정·통합 지원, agent 일을 대신하지 않음(REQ-AF-010, §A.4). (3) 대기 기록 필수 필드 여섯 개, 하나라도 없으면 무효(REQ-AF-016, design §2.1c). (4) PR을 연 agent가 다른 Issue를 맡는 조건: 원 Issue 수리 우선, `wip_limit`(미설정 동작은 OD-7 그대로), 별도 쓰기 worktree, crash-safe 인계(REQ-AF-009). (5) 시험 세 가지를 기존 AC 안의 사례로 넣음: CI 실패 수리(AC-AF-009 (f)), crash 인계(AC-AF-004 (e)), 답장 유실 복구(AC-AF-005 (e)). 그 밖에 AC-AF-009 (g), AC-AF-010 (f), AC-AF-016 (e). (6) 열린 결정 OD-13~OD-17 추가(claim 저장소 디렉터리 관례, 'AF configured' 판정 키, cursor 저장 위치, S46 대조 세 항목, MaxChunks·SeenLimit 키 이름). (7) `moai update`의 provider template 보존 동작을 알려진 동작으로 기록(§B.2, design §16). 운영자의 가장 새 정본 Page 'seq57'은 이 SPEC 작성자가 읽지 않았다(§H). 하위 SPEC은 고치지 않았다. 후속 정정(같은 판, 판 번호 그대로): 리더 판정(리더 세션 경유)을 §A.1·§C.7 DEC-AF-23 출처 칸·§H 0.6.7에 적었다. 정본 seq57(운영자가 리더 세션에 직접 전달, Page 원문은 리더도 이 SPEC 작성자도 읽지 않음)이 S46의 E03·E06 문언을 대체하며, 원문 대조는 하지 않았다. E03·E06 바꿔 쓰기와 design §2.2의 S46 병기는 그대로다. 두 번째 후속 정정(같은 판): 리더 결정(리더 세션 경유)과 lane-19 대조(리더 전달, 코드는 이 작성자가 다시 읽지 않음)에 따라 (a) OD-16에서 E18 채택(adopt) 간선 — BUILDING/FIX + PARKED `contract-change`에서 승인된 revision 채택 시 BUILDING/SPEC, owner·receipt·epoch 유지 — 을 '리더 결정, 정본 대조 안 함'으로 옮기고 나머지 하위 항목은 열어 두었으며, design §2.2 E18 행에 채택 결과와 revision 승인 증거를 더했다. (b) 책임과 실행 자리는 따로 있는 두 자원이 아니라 효과로 읽는다는 리더 판정을 적었다: Holder가 책임을 계속 갖고, 새 Issue 수락은 acquire 검사가 판정한다(§H 0.6.7, DEC-AF-23, design §2.2 E03). (c) OD-7·OD-17에 구현 관찰(요구사항 아님)을 적었다: `wip_limit` 미설정은 거절, 재시도 상한 `0`은 허용 안 함(1 이상). OD-7 기본값 질문은 여전히 열려 있다. (d) 후보 OD-18(REQ-AF-010 역할 쓰기 경계를 effect 경계에서 집행할지)을 기록만 했다. 세 번째 후속 정정(같은 판): 리더 판정(리더 세션 경유)에 따라 REQ-AF-009 SHALL 문장과 AC-AF-009 사례 (c)의 "agent claim·slot 반환"을 효과 읽기로 바꿨다 — 실행 자리를 비우고(새 Issue 수락은 acquire 검사가 판정) 책임 기록(Holder)은 같은 agent에 남는다(DEC-AF-23과 같은 뜻, S46 원문은 design §2.2 E03에 그대로 병기). 네 번째 후속 정정(같은 판): 리더 결정(리더 세션 경유)에 따라 design §6.1 "제안 설정 키 표 (리더 결정 대기)"를 더해 OD-13·OD-14·OD-17을 묶었다 — 모든 키 이름은 provisional, 기본값 칸은 모두 `—`, 새 파일에만 두며 새 파일의 이름·위치와 webhook 비밀의 위치는 열려 있다. 근거는 lane-20 통지(`efeb3dc87c`)와 lane-20의 정정(하드코딩 값 14a~14c, host 헤더 값은 정책 상수 아님)이며 이 작성자는 다시 읽지 않았다. 다섯 번째 후속 정정(같은 판): 리더 결정(리더 세션 경유)에 따라 REQ-AF-009·design §2.2 E03·§4.8·AC-AF-009 (a)의 인계 receipt 필드에 lane-18 이름(`2251e77393`)을 붙이고, 수리 세션 한정(`67defdeb1f`)·`repair-<n>`(`06f8ccc3c7`)·PARKED 진입 Wait 기록(`deb7551aac`)·worktree 경로 정규화(`150d49be8f`)를 관찰로 적었으며, SPEC 문언과 어긋나는 곳은 고치지 않고 §H 0.6.7에 차이로 남겼다(코드는 이 작성자가 다시 읽지 않음). 여섯 번째 후속 정정(같은 판): 리더 결정(Jev 상담; 리더 세션 경유)을 DEC-AF-25(D1~D5, Jev 점수는 advisory signal)로 §C.7에 적고, OD-13·OD-14를 해결로 표시(행 유지)하고, OD-17을 남은 네 항목(Tick 주기, webhook 비밀 참조 위치, 14a/b/c, `forge.Adaptive` 검증 채택)으로 줄이고, OD-19(template 배포, 배포 파일 만들지 않음)를 더했으며, design §6.1 표를 확정/대기로 나누고 inbox 행을 분리하고 `enabled`·`claim_store.dir` 행을 더했다. 일곱 번째 후속 정정(같은 판, 리더 결정(리더 세션 경유), 사실은 전달받은 것이며 다시 읽지 않음) — **A(충돌 판정)**: §H 0.6.7의 세 차이를 판정으로 바꿨다. A1 필드 대응(계약=AdoptedSpec, 보존·quiescence=`WriterStopped`+`BoundaryHead`==`HEAD`+`branch`+`worktree`, 원 writer 출처=claim 획득 때 `ErrNotIssueAuthor` 검사, receipt 필드 추가 없음)에 맞춰 REQ-AF-009·design §4.8·E03의 이름을 고쳤다. A2 옛 receipt는 다시 검증하지 않으며 "모든 필드"는 새 인계(E03 쓰기)에만 적용된다. A3 수리 세션과 agent는 충돌이 아니다(`TakeOver`가 epoch+1과 함께 `Session`을 옮김), PARKED 인계 때 Wait 처리는 열린 관찰. **B(설정 필드)**: design §6.1의 14a·14b를 lane-20 새 필드 `ghintake.GH.ListLimit`·`adapt.GHReader.PRListLimit`(`af72c90458`, 키 이름 provisional)로 바꾸고 14c 행을 지웠으며(재시도 한도는 `claim.UpdateDoc`), OD-17에서 열린 하드코딩을 빼고 두 키 이름을 열린 항목으로 적었다. **C(`enabled` 행)**: design §6 본 표의 `enabled` 행을 기본값 `—`·필수·bool 외 거절로 고치고, 키가 없거나 `false`이면 오류가 아니라 설정되지 않음(이유 표시)이라고 적었으며, §6.1의 차이 기록과 OD-14를 해결로 바꿨다. 여덟 번째 후속 정정(같은 판, 리더 결정(리더 세션 경유)): `enabled` 판정을 두 단계로 적었고(없거나 false → 미설정·오류 아님, true인데 다른 필수 키 누락·범위 위반 → 키 이름을 담은 검증 오류이며 결과도 미설정; design §6·§6.1·OD-14), DEC-AF-25 끝에 "현재 이름은 design §6.1"만 붙였으며, 설정 읽기 구현(`dc1bff9643`, 다시 읽지 않음)을 design §6.1 아래 관찰로, SPEC이 정하지 않은 동작 (a)~(i)를 §H 0.6.7 열린 관찰로 적었다((a) `max_id_bytes` 필수 범위 불일치는 리더 판정 필요). 같은 차례에 이어진 리더 결정(리더 세션 경유): **D1** `max_id_bytes`는 언제나 필수(§6.1 두 행 고침, (a) 해결). **D2** `intake.list_limit`·`monitor.pr_list_limit` 확정(필수, 정수 >= 1); OD-17에는 `secret_ref`·`tick_period`·비밀 참조 위치·`forge.Adaptive` 검증 채택만 남김. **D3** `max_doc_bytes`로 끝나는 세 키는 >= 1이며 `0` 거절(256 KiB 사본은 lane-20 내부 상수이지 설정 값이 아님), 부등식은 언제나 검사; `cursor.max_chunks`의 `0` = 나누기 끔은 그대로. **D4** (b)~(h)는 리더가 수용한 관찰로 표시하고 (i)는 D3로 대체되어 뺐다. |

## §A 범위와 규범 원천

### A.1 규범 원천

이 SPEC의 규범 원천은 GitHub Issue #1948 "[P0] AgenticFactory 우선 완성"의 **본문과 댓글 11개**다. 이 실행에서 `gh issue view 1948 --comments --json comments`(읽기 전용)로 받은 사본은 댓글 11개였고, 앞서 저장한 사본(v6)과 같았다. 문장이 서로 부딪히면 **가장 새 revision이 이긴다**.

| 번호 | revision | 다루는 것 |
|---|---|---|
| [0]~[2] | S28 기본 설계 1/3~3/3 | 상태 모형 E01~E27, PARKED overlay, 권위·effect FSM, 시험 manifest M01~M56 |
| [3] | DELTA1 | 수집기·delta·record 5종(CollectorCheckpoint·SourceRevision·ConsumerDelivery·ConsumerACK·SourceTombstone), 공유 API 예산, SYNC01~12 |
| [4] | DELTA2 | Issue·PR template, 댓글 묶음, WT 시험 |
| [5] | DELTA3 | 계층·분류·설정, GH 시험 |
| [6] | DELTA4 | 레인별 소유(lane-17~22) |
| [7] | S46 `AF-AGENT-HANDOFF-S46` | PR 인계(E03), 수리 재배차(E06), PRQueue, HANDOFF01~08, CLI19~23 |
| [8] | S51 `AF-INTAKE-CONTEXT-S51` | 백그라운드 접수·준비 7행, Issue 맥락 장부, PREP01~10, CONTEXT01~06 |
| [9] | S52 `AF-EFFICIENCY-S52` | 완료까지 지속 개선, 효율 측정, EFF01~06 |
| [10] | S53 `AF-RATE-BUDGET-S53` | webhook 우선, ETag 범위, 429/403 판별, 안전 쓰기 예산, SYNC13~18 |

정본에 없는 내용은 이 SPEC의 결정(§C.7 DEC 표)과 리더·레인이 보고한 사실(§B.2)로만 채웠고, 각각 출처를 붙였다. 확인하지 못한 것은 §G·§H로 넘겼다.

0.6.7: 리더는 운영자의 가장 새 정본으로 Page 'seq57'("agent가 Issue 종결까지 책임 유지", 제품 시험 NOT_RUN)을 전달했다. seq57은 운영자가 리더 세션에 직접 전달한 것이며, Page 원문은 리더도 이 SPEC 작성자도 읽지 않았다(canon seq57: operator direct relay to the leader session; page original not read by the leader or by this SPEC author). 리더 판정(리더 세션 경유)에 따라 seq57이 S46의 E03·E06 문언을 대체하며, 원문 대조는 하지 않았다. 위 표에는 넣지 않았고, 이 판에서 그 Page에 기댄 문장은 모두 리더 결정으로만 적었다(§H 0.6.7).

### A.2 이 SPEC이 다루는 것

이 SPEC은 **계약 + 정책 + 설정 초안**이다. 코드를 쓰지 않는다. 다루는 것: 정본 상태 모형과 전이 guard, 권위·캐시·reconcile 방향, 소유권·단일 writer·fencing, effect 불명 처리, DONE 순서, 병합 게이트와 PR 적체 상한, 보호 규칙 비약화, PR 인계와 수리, 독립성과 역할 잠금, 접수·준비, 비밀·공개 검토, 인증·rate limit·예산, 공급자 관문(GLM 제외·Codex 명시 설정), Issue 쓰기 권한과 협업 옵션, checkpoint 복구, 맥락 장부와 증거 위치, 관찰 보기, 설정·브라우저, CLI, template 참조, 환경 행렬, 효율 측정 어휘.

### A.3 "구현 완료"의 정의 경계

Issue #1948은 이 문서들이 "상위 착수 원장이지 구현 완료나 실행 시작의 증명이 아니다"라고 적는다. 이 SPEC이 PASS를 받아도 AgenticFactory가 구현된 것이 아니고, Phase 2(남은 카드의 Issue 이관)를 한 것도 아니다. 정본 S51은 "병목을 줄이기 위한 실행 계약이지 **병목 0을 보장하는 주장이 아니다**"라고 적으며, 이 SPEC도 그렇다. 정본 S52에 따라 문서 작성만으로 성능 검증이나 전체 완료를 선언하지 않는다.

사람이 보는 설정 화면(브라우저 3선택 UI)의 시험은 이 SPEC 완료로 닫히지 않는다(REQ-AF-021, §I).

### A.4 명칭 (리더 결정 DEC-AF-08)

| 개념 | 내부 역할 이름 | 사용자·CLI 표시 | 비고 |
|---|---|---|---|
| LLM 조율 | `Coordinator` | `leader` | 기존 `moai cc -f` 진입 |
| 실행자 | `agent` | `agent` | "worker"는 폐기. 레거시 이름 `worker-<n>` 거절은 유지 |
| 결정적 제어·복구(LLM 없음) | `ControlService` | (상주 서비스) | 이름은 미확정(provisional) |
| 실행 단위·시도 | `run`, `attempt` | — | 미확정 |

0.5.0까지의 `Worker`는 이 판에서 모두 `agent`로 읽는다.

leader(내부 `Coordinator`)의 역할(0.6.7, 리더 결정(리더 세션 경유)): 변경 알림, 실패 복구, 배정 조정, 통합 지원. 배정된 agent의 구현·수리·리뷰 대응을 대신하지 않는다(REQ-AF-010). 이름 규칙은 위 표 그대로다: 사용자·CLI에는 "leader", 내부에는 "Coordinator", 실행자는 "agent"이며 "worker"는 쓰지 않는다.

## §B 배경 — 사실

### B.1 이 트리에서 읽거나 실행한 것

트리: worktree `af-contract-spec`, 브랜치 `WT-agentic-factory-contract`, HEAD `3275706cc9`(이 실행 시작 때 깨끗함). 상세 명령은 research.md.

- `git ls-remote origin refs/heads/moai-control` → `ae3a4edcaa529bc927357b345051b96cd04605a1	refs/heads/moai-control`(이 실행, exit 0). control ref가 원격에 genesis 커밋으로 있다. 그 안의 `control/schema.json` sha256 `1fc23184522e104b985bc6a8429a347836aaaab4714526939d5bc9363b290640`은 **리더 보고**(원격 readback 일치)이며 이 실행은 다시 재지 않았다.
- `internal/agenticfactory/`는 이 트리에 없다(0.5.0 EV-REF-0, 트리 `5c3f0ebf27`; 이 트리에서 다시 실행하지 않음).
- 활성 판정: `homestate.ClassifyOwnerWith`(`internal/homestate/factory_run_retire.go:152-168`)는 `ProbeProcessIdentity`(`internal/homestate/profile_lease.go:210-220`)가 돌려준 fingerprint와 기록값을 **문자열 동등 비교**만 한다(`fingerprint == processStart` → live, 다르면 dead). fingerprint 원천: Linux 등 `ps -o lstart= -p <pid>` 출력 문자열(`process_fingerprint_unix.go:11-17`, 초 단위, 순서 비교 정의 없음), darwin `"%d.%06d"` 초.마이크로초(`process_fingerprint_darwin.go`), Windows 생성 시각(`process_fingerprint_windows.go`). 모두 이 실행에서 읽음(darwin·Windows는 0.5.0 이전 판에서 읽은 함수 위치를 다시 쓰지 않고 인용).
- `FactoryProcessAlive`(`internal/factory/factory_alive_unix.go:22-28`)는 PID만 본다(0.2.0 측정).
- `main` 보호 설정(0.2.0 `gh api` 측정, acceptance EV-AC008-BASE): 필수 검사 3개, `enforce_admins: true`, 강제 푸시·삭제 금지, `required_linear_history: false`.
- 이 프로젝트의 로컬 `workflow.yaml`은 Codex 기능을 켜 두었고(`audit.model: codex` 등), 코드 기본값과 배포 템플릿은 꺼 둔다(0.1.0 측정).
- (0.6.1, 트리 HEAD `89f4e812a3`) 닫힌 사유 문자열 8개를 lane-18 트리(worktree `af-core-claim-runner`, 체크아웃된 브랜치 `WT-agentic-factory-core-claim-runner`의 ref 파일 값 `fa781ef0e9`)의 `core/state.go`·`core/authz.go`와 바이트 단위로 대조했다. 그 트리에는 git을 실행하지 않았고 읽기만 했다. 같은 두 파일을 이 트리에서 `git show f9f44e5ecc:<경로>`로도 읽었다. `git merge-base --is-ancestor f9f44e5ecc fa781ef0e9`는 종료 0이었고, `git diff --stat f9f44e5ecc fa781ef0e9 -- <두 파일>`의 출력은 비어 있었다. 결과는 design §2.1a.
- (0.6.1) `git cat-file -t`로 `da7f3a4290`·`3c7db03d7b`·`f9f44e5ecc`·`e6499b731a`·`51edce26a8`·`fa781ef0e9`가 모두 `commit`임을 이 트리에서 관측했다. 앞의 세 커밋은 리더가 확인한 사실과 같다.
- lane-17 설정 읽기 코드(worktree `af-settings-keys`, `internal/config/agentic_factory.go`)를 이 실행에서 grep으로 읽었다: 키 `workflow.agentic_factory.browser.selected`·`codex.enabled`·`pr_queue.wip_limit`·`collaboration.enabled`·`collaboration.allowed_repositories`·`collaboration.permissions.{edit_body,close,labels,assignees}`, 함수 `NormalizeRepositoryPath(provider, s)`, `ResolveAgenticFactoryBrowser`, `CodexOptIn`.

### B.2 리더·레인이 보고한 사실 (이 실행이 검증하지 않음)

| 사실 | 출처 | 상태 |
|---|---|---|
| 접수 상태 7행을 core에 구현 | lane-18 커밋 `da7f3a4290`, 리더 보고 | 미검증 |
| 닫힌 사유 문자열 8개(design §2.1a의 `unclassified-403`과 표 끝 7행) | lane-18 커밋 `f9f44e5ecc`(브랜치 `WT-agentic-factory-core-claim-runner`), **lane-18 보고·리더 확인, `core/state.go`의 임시 상수** | 0.6.1에서 바이트 대조함: 8개 모두 일치(design §2.1a) |
| 커밋 `da7f3a4290`, `3c7db03d7b`, `f9f44e5ecc`가 저장소에 있다 | 'reported by lane-18 / lane-20, relayed by the leader; commit existence confirmed by the leader via cat-file' | 리더 확인. 이 실행도 `cat-file -t` → `commit` 관측 |
| `f9f44e5ecc`에서 `op-not-defined`는 `core/authz.go`에, `foreign-closing-keyword`는 `core/state.go`에 있다 | 'reported by lane-18 / lane-20, relayed by the leader; commit existence confirmed by the leader via cat-file' | 리더 확인. 이 실행의 바이트 대조: `"foreign-closing-keyword"`는 `state.go:83`에 있다. 따옴표 붙은 `"op-not-defined"`는 `state.go:86`에만 있다. `authz.go:208`은 상수 `ReasonOpNotDefined`를 반환할 뿐이다(리더 정정 반영, design §2.1a) |
| claim 문서 종류 `collector`(Safety 등급) | 커밋 `fa781ef0e9`, 'reported by lane-18 / lane-20, relayed by the leader; commit existence confirmed by the leader via cat-file' | 내용은 미검증. 이 실행도 `cat-file -t fa781ef0e9` → `commit` 관측(lane-18 브랜치 ref 파일 값과 같음) |
| checkpoint 상태 모형 구현: `local`·`pending`·`remote-durable`·`recovered` | 커밋 `e6499b731a`, 'reported by lane-18 / lane-20, relayed by the leader; commit existence confirmed by the leader via cat-file' | 내용은 미검증(§H 공백 (d)) |
| checkpoint push는 원격에 없다고 확인되면 최대 3번까지 다시 시도한다(`checkpoint.go:28`) | 커밋 `e6499b731a`, lane-18 보고, 리더 인용 | 내용은 미검증(§G.1 OD-11) |
| 코드의 checkpoint 상태 넷과 필드로 design §4.6의 상태 일곱과 같은 정보를 나타낼 수 있다(`checkpoint.go:125-146`) | lane-18 보고 | 미검증(§H 공백 (g)) |
| 코드의 상태 모형은 계약 0.4.0 기준에 머물러 있고 정본 S46과 다르다. 리더가 lane-18에 맞춤 방안을 요청했다 | 리더 보고 | 관찰로만 기록(§H 공백 (h)) |
| 코드는 보류 상태 이름으로 `on-hold`를 쓴다 | lane-18 보고, 리더 전달 | 미검증(§G.1 OD-12) |
| 옛 상태 키 이전은 세 단계로 나란히 진행된다: 새 키 추가 → 전환(옛 값을 새 값에서 파생) → 옛 키 제거 | 'leader decision / lane-18 implementation in progress' | 관찰로만 기록. 이 SPEC은 코드에 요구하지 않는다(design §2.3) |
| 옛 상태 → 정본 값 대응은 이름이 아니라 증거로 정한다(0.6.3 정정): writer 인계 receipt와 leader 인계 수락이 둘 다 있으면 `REVIEWING`, `merged=true`면 `MERGING`, 그 밖은 `BUILDING/IMPLEMENT` + 검사 증거. `blocked`는 현 단계 + `PARKED`(사유), `on-hold`는 OD-12 | 'leader decision, relayed via the leader session' | design §2.3에 반영 |
| lane-18 커밋 `af80e0bdec`: core S46 모형, E01~E27 표, `FromLegacy`. 정본이 정하지 않은 범위에 둔 잠정 값 6건(번호 1·2·3·5·7·6, 4번은 전달되지 않음)과 PARKED 중 단계 변경 간선 거절(리더 결정) | 'reported by lane-18, relayed by the leader; commit existence/race test confirmed by the leader' | 커밋 존재는 이 실행도 `cat-file -t` → `commit` 관측. 내용·경쟁 시험은 미검증. 정본 대조는 관찰로만(§H lane-18 잠정 값) |
| 설정 읽기 커밋 `6c833ee4b4`·`8920c0ecf8`은 `5c278d8b51`의 조상이고, lane-22의 병합 `7a88d26259`가 함께 가져왔으며, 셋 다 통합 HEAD의 조상이다(의도한 흐름). `6c833ee4b4`의 `wip_limit` 읽기는 미설정을 nil로 두고 OD-7을 미리 정하지 않는다 | 'leader direct reading of the settings-tree log and diff, relayed via the leader session' | 계보와 주석·시험 이름은 이 실행도 읽기 전용으로 관측(§H 0.6.6). OD-7 판단은 리더의 것 |
| lane-18 커밋 `52363f7ac7`: E03 guard가 S46 49행 아홉 항목 전부를 덮음, E15·E16 결과별 guard(정리 불명 → 현 단계 + PARKED), E19 사유가 `core.ClassifiedReasons`(11개) 밖이면 거절 | 'reported by lane-18 / leader decision, relayed via the leader session; leader-confirmed race test ok' | 커밋 존재는 이 실행도 `cat-file -t` → `commit` 관측. 11개 값은 이 실행이 읽음(§H 0.6.5 (A3)). 나머지 미검증 |
| 리더 결정: 옛 상태는 진척 기록으로 남기고 `FromLegacy`로 파생한 새 단계·하위 단계·PARK를 함께 기록. 병합 게이트 사유(G)는 overlay에서 단계를 유지하는 게이트 기록으로 옮기며, 'Resume after blocked'가 '같은 단계에서 새 증거로 다시 검증'으로 바뀜(동작 변경) | 'leader decision, relayed via the leader session' | 관찰로만 기록(§H 0.6.5 (B1)·(B2)) |
| 리더 정정: (3) E24·(5) 종결과 overlay 해제·(7) E19는 정본이 침묵하는 범위가 아니다(S46 70·61·65행). 리더가 lane-18에 E15 정리 불명 시 현 단계 + PARKED 유지, E19 사유를 실제 오류 분류에서 파생·검증, E03 guard 증거 키가 49행 목록 전체를 덮음, E24를 정본 근거로 표시하라고 지시했다 | 'leader correction, relayed via the leader session; the leader read S46 lines 49, 61, 65, 70 directly' | 코드 쪽 지시는 리더 보고, 미검증(§H 리더 정정) |
| lane-18의 커밋되지 않은 진행 중 `core/legacy.go`는 이미 `landing`을 `HasReceipt && LeaderAccepted`이면 `REVIEWING`, 아니면 `BUILDING/IMPLEMENT`로 나눈다. 리더가 lane-18에 같은 증거 규칙을 `ready`에도 적용하라고 요청했다 | 리더 보고 | 미검증(커밋되지 않은 코드, 이 실행은 읽지 않음). 이 SPEC은 코드 변경을 요구하지 않는다 |
| 약 190 KB를 넘는 collector 스냅숏은 저장되지 않는다(lane-20 한계) | 커밋 `51edce26a8`, 'reported by lane-18 / lane-20, relayed by the leader; commit existence confirmed by the leader via cat-file' | 내용은 미검증(§H 공백 (e), 운영자 결정 대기 OD-10) |
| 협업 판정 규칙은 `core.Collaboration.Authorize` 한 곳. 설정 키 → `core.Collaboration{Provider, Enabled, AllowedRepositories, EditBody, Close, Labels, Assignees}` 연결은 아직 없음(주입 필드만) | lane-18 보고, 리더 전달 | 미검증, 열린 항목(§G.1) |
| 자동 종결 키워드 목록 = GitHub 문서의 아홉 낱말(close/closes/closed, fix/fixes/fixed, resolve/resolves/resolved). 대소문자 혼합·URL 형태는 문서에 없지만 안전 쪽으로 포함. 문서 확인은 WebFetch 요약이며 원문 대조 안 함(uncollated) | lane-18 보고, 리더 전달 | 미검증 |
| 설정 읽기 코드 `5c278d8b51`(브랜치 `WT-agentic-factory-settings-keys`): 리더가 독립 재검증(2파일, 깨끗한 트리, 판정 도우미 0, `internal/config` 시험 ok) | 리더 보고 | 이 실행은 grep만 함 |
| `--agent` 인자가 Claude로 그대로 넘어감(argv 수준 특성화 시험) | lane-22 커밋 `3c7db03d7b`, 리더 보고. Claude 쪽 효과는 단언하지 않음 | 미검증 |
| `TestProtectionGuard_*` e2e 시험 3개 작성됨(레이블 쓰기 endpoint 고정, 쓰기 기본 꺼짐 시 전송 0, 제품 소스의 보호 설정 용어 0). "약화 연산 일곱 개가 각각 사유 하나로 거절"은 거절 경로가 없어 NOT_RUN | lane-22 보고(리더 전달) | 미검증 |
| 종결 순서(Issue 종료 → DONE)는 lane-18 코드의 기존 시험으로 증명됨 | lane-22 보고 | 미검증, 시험 이름 미전달 |
| **알려진 동작(결함 아님)**: 프로젝트에 Agentic Factory provider template 파일이 이미 있지만 manifest에 없으면 `moai update`는 그 파일을 `user_created`로 기록하고 건너뛴다. 그래서 그런 파일은 뒤의 template 개정을 받지 못한다. 사용자 파일 보존이 우선한다. 근거 코드: 브랜치 `WT-agentic-factory-templates`의 커밋 `8ebbacad30`(`internal/template/user_owned_templates.go`의 `isUserOwnedTemplatePath`, `deployer.go` 한 줄 변경), 보호 시험 승격 `47d8b782b7` | 'leader decision, relayed via the leader session' | 알려진 동작으로 기록(design §16). 이 실행은 두 커밋을 읽지 않았다 |
| GitHub REST: "The name is not case sensitive."(https://docs.github.com/en/rest/repos/repos, 2026-10-29 확인, WebFetch 요약·uncollated). 저장소 이름 변경·이전 때 301 응답(uncollated) | 리더 전달 | 원문 대조 안 함 |

## §C 요구사항 (GEARS)

요구사항은 묶음 15개(G01~G15)와 표 10개(T01~T10)로 25개다(Tier L 상한 25). 표 REQ는 "design.md의 표를 지킨다"는 한 줄 계약이며 표 본문이 세부다(DEC-AF-10). 각 REQ의 소유 레인은 design §10.

### C.1 묶음 요구사항 — 상태·소유·병합

| ID | 묶음 | 요구사항 |
|---|---|---|
| REQ-AF-001 | G01 활성 claim·fencing | The claim authority SHALL hold at most one active claim and exactly one accountable writer per Issue: of concurrent claims for the same Issue exactly one SHALL succeed, every reissue, attempt, and its history SHALL stay under the same Issue key, and every state write SHALL carry the attempt epoch so that a write carrying an epoch older than the current one is refused. When a dispatch names an Issue that already has an active claim, a live attempt, or an undelivered dispatch, the dispatcher SHALL refuse it and SHALL record the claim it collided with. |
| REQ-AF-002 | G02 권위·재구축·checkpoint 복구 | (a) The online control ref (`refs/heads/moai-control`) SHALL be the authority for technical state (claim, attempt, epoch, effect intent and receipt, checkpoint manifest) and the Issue SHALL be the authority for business fields (`design.md` §4); the local claim store SHALL be a cache rebuildable from the control ref and approved immutable artifacts and SHALL NOT by itself confer a claim, a grant, or a completion. (b) When the control ref cannot be read, or a compare-and-set write against it fails, the ControlService SHALL keep the dependent transition from proceeding and SHALL NOT proceed on the local cache alone. (c) A checkpoint SHALL pass four independent steps — local commit, remote acknowledgement, control-ref manifest verification, Issue summary link — each with its own evidence and failure state (`design.md` §4.6); a checkpoint SHALL be `remote-durable` only after the remote acknowledgement and the manifest verification are both observed, a failed Issue link SHALL NOT revoke `remote-durable`, and before any push the checkpoint writer SHALL detect a ref and SHA already acknowledged and SHALL NOT push them again. (d) When a checkpoint is reused or recovered on the same or another host, the ControlService SHALL first establish that the previous writer is `dead` or recorded a safe stop and SHALL advance the epoch, and SHALL NOT start a second writer. (e) The checkpoint path SHALL NOT stage or commit the user's uncommitted changes automatically, push unfinished work to `main`, force-push or rewrite history, or carry any value of redaction categories C1–C5, and checkpoint storage SHALL stay within the configured cap and retention parameters. |
| REQ-AF-003 | G03 소유·Issue 쓰기 권한 확인 | When ownership verification fails or cannot be completed (claim absent, holder mismatch, authority unreadable), the writer SHALL refuse the write and SHALL NOT fall back to an unverified write. When AgenticFactory is about to write to an Issue, the single Issue-write decision (`design.md` §4.1) SHALL allow a body edit or a close only on an Issue whose author's immutable numeric id equals the authenticated user's numeric id, never by comparing login names; on another user's Issue it SHALL allow only related comments, plus a claim only where the collaboration option is enabled and the repository is on its allow-list, and a body edit, close, label change, or assignee change only where additionally that operation's own permission key is true; it SHALL refuse label, assignee, title, milestone, and reopen operations on the user's own Issue whatever any permission key holds; it SHALL refuse with exactly one reason of `design.md` §2.1a when an identity cannot be established, the repository is not allowed or answers with a different repository or any redirect, the operation is not permitted or not defined, a foreign owner's handoff is unconfirmed, or a PR body or commit carries a closing keyword aimed at an Issue it may not close. |
| REQ-AF-004 | G04 활성 판정·인계 | While an attempt's lease is expired and the liveness classification of its holder (`homestate.ClassifyOwnerWith` over the pid and process-start fingerprint recorded at claim time) is `live` or `indeterminate`, the ControlService SHALL refuse a takeover and SHALL mark the attempt `lapsed`; a new attempt SHALL be issued only when the classification is `dead` or the holder recorded a safe stop, with the epoch incremented and the prior attempt's worktree, branch, SHA, and dirty state preserved and referenced. When a session disconnects, a lease expires, the ControlService restarts, a stale HEAD result arrives, or a step fails partway, the recovery path SHALL first query the actual execution state and SHALL then resume, reissue, or park; it SHALL NOT force-kill a process and SHALL NOT start a duplicate attempt. |
| REQ-AF-005 | G05 effect 불명 재조회·전달 상태 | The delivery tracker SHALL record `sent`, `received`, `executed`, and `completed` as four separate states per command, each set only from its own evidence source (`design.md` §3); a send result or an idle notice SHALL NOT set a later state. When the outcome of an external effect (Issue creation, comment, branch push, merge request, close) is unknown, the effect executor SHALL have recorded the effect intent before sending, SHALL re-query the remote fact before any re-issue, SHALL NOT re-issue blindly, and SHALL keep the item `EFFECT_UNKNOWN` or `MERGING` with `PARKED unknown-effect` until the fact is established. |
| REQ-AF-006 | G06 DONE 순서 | When the remote integration of a merge is confirmed (E12), the state engine SHALL record `DONE` only after every completion obligation of E13 is met — acceptance evidence, merge commit, an approved local sync whose per-path content digest, index entry, and mode are unchanged for every path that was modified, staged, or untracked before the sync, a deploy record or a recorded `deploy: N/A` with its reason, and the Issue close — so that the Issue close precedes `DONE`; when the sync or the close fails it SHALL keep `MERGING` with `PARKED pending-effect` and SHALL NOT revert the merge (E14); the next dispatch for the freed capacity SHALL be admitted only after `DONE`; and migration of remaining cards SHALL NOT be recorded as `DONE` or as implementation completion. |
| REQ-AF-007 | G07 병합 게이트·PR 적체 상한 | The merge gate SHALL admit E08 only when, for the latest HEAD, every required check actually ran and passed, a same-HEAD code review and security review each carry `PASS`, the required independent audit is a debt-free `PASS`, the official QA skill verdict is `PASS`, and a valid PR handoff receipt exists (REQ-AF-009); every review, audit, QA, writer-handoff, and leader-acceptance item SHALL count only when the attribution source (REQ-AF-011) returns `attributed` with the role that evidence kind requires, and every other result SHALL be recorded through the single reason table of `design.md` §2.1b; a skipped, missing, failed, cancelled, stale, `FAIL`, `BLOCKED`, unparseable, or `PASS-WITH-DEBT` item or an exhausted audit budget SHALL be recorded as a block and SHALL NOT count as a pass. While open PR handoffs plus the repair backlog of a repository are at or above `workflow.agentic_factory.pr_queue.wip_limit`, the dispatcher SHALL dispatch no new Issue while closure and repair continue; the key SHALL be an integer of at least 1, and `0`, a negative, a non-integer, or an out-of-range value SHALL be refused, never read as "off". |
| REQ-AF-008 | G08 보호 비약화 | AgenticFactory SHALL NOT delete a required check, weaken branch protection, change the repository's linear-history setting, force-push `main`, delete a protected branch, merge through an administrator bypass, or change a protection rule through any label or settings endpoint; the effect boundary SHALL refuse each such operation (`design.md` §11.3) with exactly one recorded reason before any request is sent. |
| REQ-AF-009 | G09 PR 인계·수리 단일 소유 | When an agent publishes a PR, the ControlService SHALL move the Issue to `REVIEWING` (E03) only when the new handoff it writes carries all of the remote PR URL (`PR URL`, http or https), remote HEAD (`HEAD`) and base (`base`), the contract as the adopted SPEC (`adopted SPEC`), task results (`task result`), check state (`check status`), the responsible party (`Session`/`Agent`; the original writer's provenance is the Issue-author check at claim acquisition), the preserved-change and quiescence evidence (`WriterStopped`, `BoundaryHead` equal to `HEAD`, `branch`, `worktree`), and the leader's acceptance (`LeaderAccepted`) (field names per `design.md` §4.8 `PRHandoff`; a receipt already recorded before this rule is read as past state and not re-validated), and SHALL then vacate the agent's execution place (whether the agent accepts a new Issue is decided by the acquire check) while the responsibility record (Holder) stays with the same agent, preserving its worktree, branch, and changes; PR publication alone SHALL NOT be recorded as done. The agent assigned to an Issue SHALL remain that Issue's single responsible party through implementation, CI-failure repair, review response, normal merge, and confirmation that the Issue is closed, and the PR handoff (E03) SHALL NOT end that responsibility; the responsibility record SHALL live on the control ref so that, when the responsible agent is classified `dead` or recorded a safe stop, responsibility passes with the epoch under REQ-AF-004 and never lapses unowned. When a repair is decided (E06) or a merging Issue's HEAD or protection changes (E10 → E06), the repair SHALL be dispatched as a new attempt under an exclusive claim on the original Issue and branch, keeping Issue, branch, worktree, and epoch continuity, to the Issue's responsible agent, and SHALL NOT overwrite that agent's other Issue. While an agent holds responsibility for an Issue whose PR is open, the dispatcher SHALL dispatch another Issue to it only when no repair of its earlier Issue is pending (that repair takes priority), the backlog condition of REQ-AF-007 admits a new Issue, the new Issue gets a separate write worktree, and the responsibility record of every Issue it holds is on the control ref. |
| REQ-AF-010 | G10 독립성·역할 쓰기 경계 | The ControlService SHALL be the only role that writes claim, lease, liveness, reconciliation, and resume-decision state; the Coordinator SHALL act on that state only through ControlService commands, SHALL limit its own actions to change notification, failure recovery, assignment adjustment, and integration support, and SHALL NOT perform the implementation, repair, or review response of an Issue assigned to an agent; an agent SHALL write only within its own attempt. No review, security-review, audit, or QA evidence SHALL count when its producer was the writer of any attempt on the same Issue, or when the Coordinator implemented the change it would judge. |
| REQ-AF-011 | G11 역할 귀속·잠금 | The merge gate SHALL judge every evidence item through the read-only attribution query of SPEC-AGENTIC-FACTORY-IDENTITY-001 (REQ-AFI-007) and never through a value the item states about itself. Until the lock is lifted, the production attribution source SHALL be a fixed implementation whose behaviour is selected by a single compile-time constant and by no configuration key, environment variable, file, or flag, and SHALL return `unattributable` with reason `identity-spec-incomplete` for every evidence kind, so that the gate records `producer-not-allowed`; the lock SHALL cover every caller including the scenario harness, attribution stubs SHALL be confined to gate unit tests, and the lock SHALL be lifted only by a single code change that lands together with SPEC-AGENTIC-FACTORY-IDENTITY-001 `completed`, its acceptance criteria green, and the operator's recorded risk acceptance of its relay-path residual risk. |
| REQ-AF-012 | G12 접수·준비 | When a work request is accepted, the intake service SHALL record a durable request with its idempotency key, canonical payload hash, and approval scope on the control ref before acknowledging, and SHALL keep acknowledgement, Issue publication, branch registration, agent reservation, worktree preparation, READY acknowledgement, and start as separate states (`design.md` §4.2); it SHALL run preparations in a bounded-concurrency queue with backpressure, ordering, stage deadlines, and restart reconciliation of recorded requests only; it SHALL prepare a worktree only after a slot and an exclusive reservation exist, SHALL let an agent write only after the preparer ended its writes and handed over READY, and SHALL NOT create a new ledger store, unbounded subtasks, or an always-on lane. |
| REQ-AF-013 | G13 비밀·공개 검토 | Before any ledger entry, comment, checkpoint manifest, Issue summary, or refined evidence file is published, the refiner SHALL remove every value of redaction categories C1–C5 (`design.md` §4.4) with a deterministic non-LLM check as the last gate, SHALL refuse the whole publication when any category value remains, and SHALL record material that cannot be published as `RESTRICTED` with only its state, a short summary, and its SHA-256. The ledger writer SHALL refuse a write to a field by a role that is not that field's declared owner, and observation reads SHALL NOT confer write authority. |
| REQ-AF-014 | G14 인증·rate limit·예산 | When a GitHub or GitLab response is 401, 403, 404, or 429, the API client SHALL classify it by the table of `design.md` §4.5 before any retry decision, recording a 403 that matches neither a rate-limit row nor a permission row as `unclassified-403` with zero retries; it SHALL NOT send a request against a rate-limited resource before the time the response names, SHALL take every limit, remaining count, and reset time only from response headers and record `UNKNOWN` when they are absent, SHALL keep no numeric limit as a code constant, SHALL NOT switch key, token, account, app installation, or endpoint to obtain the same data or effect, and SHALL send state-safety writes before safety notices and safety notices before progress or report comments while the shared budget is below its reserve; an operation that would add a server, credential, scope, permission, or cost SHALL wait for a separate operator approval. |
| REQ-AF-015 | G15 공급자 관문 | AgenticFactory implementation work and local-environment verification SHALL NOT invoke a GLM-managed backend. Where `workflow.agentic_factory.codex.enabled` is not explicitly `true`, AgenticFactory SHALL make zero Codex calls. The Codex default migration SHALL follow the order of `plan.md` §F.3 — introduce the opt-in gate, verify zero calls while unset, then change defaults — and no existing Codex feature SHALL be removed before its replacement is verified. |

### C.2 표 요구사항 — design.md 표가 본문

| ID | 표 | 요구사항 |
|---|---|---|
| REQ-AF-016 | T01 전이표 | The state engine SHALL apply a transition only when it appears in the transition table E01–E27 of `design.md` §2.2 with its guard evidence present and the requester is that edge's writer, SHALL refuse every other request with the prior state unchanged, and SHALL record each `PARKED` overlay with exactly one reason of the closed set of `design.md` §2.1a, its evidence, and its resume condition. Every wait record — a `PARKED` overlay and any other recorded wait of an Issue or an agent — SHALL carry all six fields of `design.md` §2.1c (single responsible party, next automatic action, re-check time, diagnosis, retry cap, escalation target), and the state engine SHALL refuse as invalid a wait record missing any of them. |
| REQ-AF-017 | T02 record | The ControlService SHALL keep the records of `design.md` §4 (CollectorCheckpoint, SourceRevision, ConsumerDelivery, ConsumerACK, SourceTombstone, PRQueue, PRHandoff, effect intent and receipt) with their declared fields, run exactly one active collector per repository, and give no agent a polling path of its own. |
| REQ-AF-018 | T03 준비 상태표 | The intake service SHALL move a request only along the seven-row preparation table of `design.md` §4.2, and E01 SHALL require that request's preparation receipt and the reserved agent's READY acknowledgement. |
| REQ-AF-019 | T04 맥락 장부·증거 위치 | The context ledger SHALL keep the entry fields, stale rule, and accessibility classes of `design.md` §4.3, SHALL link evidence only by full-commit-SHA URLs, SHALL mark an invalidated entry `STALE` without deleting it, and SHALL publish refined evidence only under `.moai/evidence/issue-<N>/` through the stages `PREPARED → MASKED → COMMITTED → PUSHED → REMOTE_VERIFIED`, reporting `AVAILABLE` only after a remote lookup. |
| REQ-AF-020 | T05 관찰 보기 | The observation view SHALL show, per project, Issue, and owner, the items of `design.md` §8.1, each with its source and read time, SHALL show the preparation states and the PR backlog distinctly from started work, and running it SHALL change no state. |
| REQ-AF-021 | T06 설정·브라우저 | The configuration loader SHALL refuse any `workflow.agentic_factory.*` value outside the rules of `design.md` §6, naming the key, the allowed range, and the file, without clamping; a read error SHALL NOT be treated as an unset key. The browser selection SHALL offer exactly `egolite`, `aside`, and `chrome`, SHALL use `chrome` only while nothing is selected, SHALL invoke only an explicitly selected browser, and SHALL block without falling back when it is not installed or cannot connect. |
| REQ-AF-022 | T07 CLI | The launcher SHALL follow the command table of `design.md` §9.0 and §8: existing entry commands SHALL keep their behaviour through the deprecation stages of §9.4, a new syntax SHALL require an explicit version declaration and refuse an undeclared mix with a compatibility error, and no argument SHALL silently change a session's role. |
| REQ-AF-023 | T08 template·분류·설정 | Issue and PR templates, taxonomy, and setup SHALL follow the reference table of `design.md` §16, SHALL report a provider capability the provider lacks as `not supported`, and SHALL batch non-urgent progress comments per its window rule without delaying security failures, approvals, or actual blocks. |
| REQ-AF-024 | T09 환경 행렬 | The environment matrix SHALL record for each environment of `design.md` §5.1 the actually executed command, exit code, and output excerpt per step, SHALL record `PASS` only when every step ran and succeeded, and SHALL record a not-run step as `not-run`, never as success. |
| REQ-AF-025 | T10 효율 측정 어휘 | Every efficiency comparison record SHALL carry the fields of `design.md` §4.7 and keep measured, estimated, `UNKNOWN`, and `NOT_RUN` values separate; the API, token, and time cost of checkpoints SHALL be recorded only as measured values or as a recorded Gap; and no comparison SHALL report an efficiency gain that hides a quality, safety, or correctness regression. |

### C.3 Issue 쓰기 권한과 협업 옵션 (REQ-AF-003의 근거)

- **Issue 쓰기 권한 정책**(DEC-AF-17, 출처 '리더 세션의 사용자 채널로 받음, 출처 표기 없음'): 본문 편집은 사용자 **자기** Issue에서만 한다. 판정은 인증된 사용자의 불변 **숫자** id와 Issue 작성자의 숫자 id가 같을 때이며, 로그인 이름은 비교하지 않는다. 남의 Issue에는 관련 댓글만 쓴다. 신원이 불분명하면 쓰기를 막는다. 따로 명시된 허락 없이 남의 본문 편집·인수·레이블/담당자 변경·종료를 하지 않는다. PR의 자동 종결 효과는 미리 점검한다. 남의 작업을 가로채는 자기 Issue 중복 발행은 금지다. 이 정책은 코드 수리를 막지 않는다. 자기 Issue 종료는 허용한다(E13/E14).
- **협업 옵션**(DEC-AF-18, 출처 '부모 대화에서 사용자가 확인, 리더 세션 경유'): 제품 기본값은 자기 Issue만 다룬다. 남의 Issue는 명시적으로 켠 협업 옵션 **그리고** 명시적 저장소 허용 목록 안에서만 claim할 수 있다. 남의 Issue의 본문 편집·종료·레이블·담당자 권한은 각각 따로이며, 옵션만으로는 아무것도 허용되지 않는다. claim 전에 인증 id, 작성자 id, 기존 owner의 인계를 확인한다. 신원 불분명이면 막는다. PR 자동 종결을 점검한다. 중복 claim을 막는다. 이 기록은 지금 남의 Issue를 바꾸라는 승인이 아니며, 새 writer·권한 확대·추가 감사를 승인하지 않는다.
- **자기 Issue의 정의되지 않은 동작**(DEC-AF-19, 리더 결정(리더 세션 경유)): 자기 Issue의 제목 편집·마일스톤 변경·다시 열기는 정본 근거가 없어 `op-not-defined`로 거절한다. 자기 Issue의 레이블·담당자 변경은 `permission-not-granted`로 거절하며, 어떤 `permissions.*` 키가 `true`여도 거절이다. `collaboration.permissions.*` 키는 **남의 Issue에 대한 쓰기에만** 적용되고 자기 Issue의 어떤 동작도 열지 않는다. 이것들은 정책 확장이 아니라 공백 기록이며(§H), 결정이 필요해지면 리더에게 보고한다.
- 판정 규칙은 한 곳(lane-18 `core.Collaboration.Authorize`)에만 있다. 설정 읽기 코드(lane-17)는 값을 읽고 검증만 한다(design §4.1 두 열 표).

### C.4 잠금 (REQ-AF-011의 근거, DEC-AF-02·DEC-AF-04)

- **무엇이 잠기나**: 귀속이 필요한 모든 증거 종류(코드 리뷰, 보안 리뷰, 독립 감사, QA, writer 인계, leader 수락). 잠긴 동안 제품 귀속 원천은 언제나 `unattributable{identity-spec-incomplete}`를 돌려주고, 게이트는 design §2.1b 표에 따라 `producer-not-allowed`를 기록한다.
- **어디에 있나**: 제품 코드의 **컴파일 시점 상수 하나**. 게이트는 그 상수만 읽는다. 설정 키, 환경 변수, 파일, 플래그 어느 것도 이 상수를 대신하거나 덮지 못한다(AC-AF-011 L3).
- **누구를 덮나**: 시나리오 하네스를 포함한 모든 호출자. 귀속 대역(stub)은 게이트 단위 시험 안에서만 허용되며, 시나리오 묶음(`af_e2e`)은 제품 원천을 쓰고 시작 때 원천 종류를 기록해 대역이면 FAIL한다(AC-AF-011 L4). 그래서 잠긴 트리에서 시나리오의 병합 단계는 초록이 될 수 없다.
- **어떻게 풀리나**: 그 상수 한 곳을 바꾸는 코드 변경 하나로만. 그 변경은 하위 SPEC `completed`, 하위 AC 전부 PASS, 하위 OQ-AFI-1 중계 경로 잔여 위험에 대한 운영자 위험 수용 기록과 **같은 착지**여야 한다(plan M7). 풀 때 AC-AF-011의 L1·L2를 "제품 원천이 등록된 증거를 귀속한다"로 다시 쓴다(acceptance §E).
- **사유 문자열** `identity-spec-incomplete`의 소유: 이 SPEC(상위)이 값을 정하고, 하위 SPEC은 같은 값을 귀속 이유 집합에 둔다.

### C.5 checkpoint 복구 (REQ-AF-002 (c)~(e)의 근거, DEC-AF-20)

- 출처: '상위 대화에서 사용자가 승인, 리더 세션 경유'. 이 SPEC은 이미 정해진 정책을 계약 문장으로 옮겼을 뿐이다.
- 단계·상태·부분 실패·manifest 필드·저장 상한은 design §4.6 표다. 소유: 상태·git·manifest = lane-18, Issue 링크 = lane-20, 충돌 회귀·측정 = lane-22.
- 묶음 G02에 넣은 이유: checkpoint는 권위(control ref)에 기술 상태를 남기고 재구축하는 일이라 G02와 같은 판단 축이다. 새 묶음을 만들면 REQ가 26이 되어 상한을 넘는다.

### C.6 정본과 부딪힌 문장의 처리

| 대상 | 0.5.0 문장 | 정본 문장(인용) | 0.6.0 처리 |
|---|---|---|---|
| 권위 저장소 | §F.3 "로컬 claim 저장소가 권위, Issue는 거울" | [3] DELTA1 "local cache는 control ref와 승인된 immutable 원본 artifact에서 재구축 가능해야 한다"; SYNC12 "local cache 소실·비LLM 서비스 재시작 → online control ref와 원본 artifact로 복구"; [8] S51 "receipt와 재개 checkpoint는 기존 온라인 control ref 권위에 두고" | REQ-AF-002 (a): control ref가 권위, 로컬은 캐시. §F.3 폐기 |
| 상태 모형 | design §2.1 `queued`…`done`, `blocked`/`on-hold` | [0]§8 READY/BUILDING/REVIEWING/MERGING/DONE/CANCELLED + PARKED, E01~E27 | design §2 전면 교체. 옛 이름 대응은 design §2.3 |
| 종결 순서 | REQ-AF-021 "`done` 기록 뒤 Issue 종료" | [0]§8 E13 "수용증거·M·승인된 local sync·close·정리 → DONE", E14 "sync/close 실패 → MERGING + PARKED pending-effect" | REQ-AF-006: 종료가 DONE보다 먼저 |
| 수리 경로 | 0.5.0 `verifying → implementing` | E06 REVIEWING → BUILDING/FIX(S46 개정), E10 MERGING → REVIEWING | REQ-AF-009: E10 → E06. 0.5.0에 없던 간선 `landing → verifying`은 E10이 메운다. 옛 이름의 정본 단계는 증거로 정한다(design §2.3) |
| PR 발행 | 0.5.0에 PR 인계 개념 없음 | [7] S46 E03 "PR 원격 H/base·계약·SPEC·task 결과·검사 상태·원 writer·leader 수락 receipt·quiescence·보존 확인 → REVIEWING; … agent claim/slot 반환" | REQ-AF-009 |
| 하위 Issue 발행 | §I "하위 이슈를 만들거나 기존 이슈를 대량 발행·종료하지 않는다" | [8] S51 `ISSUE_CONFIRMED`(접수 요청의 Issue 발행) | §I 문장을 "Phase 2 이관 문맥의 대량 발행"으로 좁힘 |
| 429 판별 | REQ-AF-014 "authentication failure, permission denial, or HTTP 429" | [10] S53 SYNC13 "GitHub primary/secondary 403·429와 권한 403 → 오류 분류·정상 재개 조건 분리" | REQ-AF-014: 403도 rate limit일 수 있음, 판별표 |
| 증거 위치 | design §11 "파일 배치는 현행 관례(`.moai/reports/<id>/`)" | [8] S51 "`.moai/report`와 `.moai/reports` 중 하나라고 추정하거나 존재하지 않는 경로를 기록하지 않는다", "접근 가능한 위치가 확인되기 전에는 LINK_BLOCKED" | REQ-AF-019, DEC-AF-13: 정제본만 `.moai/evidence/issue-<N>/` |
| 접수 상태 수 | (없음) | [8] S51 준비 상태표 7행 | REQ-AF-018. lane-20 표시 이름 6개는 정렬 중인 표시 이름일 뿐이며 코드 단계 수로 쓰지 않음 |
| 역할 부여 운영자 인증 | DEC-AF-01 ② 요청마다 운영자 인증 | [0]§2 "매 카드 사람 승인을 새로 추가하거나 기존 required review를 삭제하지 않음", [0]§1 "사람은 시작 시 승인 범위를 정하고 마지막에 검수한다" | **좁게 기록**: 정본이 현재 문언보다 넓게 허용하는 부분은 적용하지 않는다. 문언 유지. 잔여 충돌은 운영자 열린 항목(§G.1) |

### C.7 결정 표

| ID | 내용 | 출처 | 반영 |
|---|---|---|---|
| DEC-AF-01 | ① 리뷰·감사 역할은 ControlService만 부여 ② 등록에 운영자 인증 필요 ③ agent는 자기에게도 남에게도 역할을 줄 수 없음 ④ 한 세션이 agent이면서 감사자일 수 없음. 정본과의 충돌 1은 좁게 기록(§C.6 마지막 행) | 운영자 결정 (2026-10-29, 리더 경유. 0.4.0 기록) | 하위 SPEC §F.1 |
| DEC-AF-02 | 하위 SPEC 완료 전 역할 증거 효력 없음(잠금) | 운영자 결정 (0.5.0 HISTORY의 운영자 재결정) | REQ-AF-011 |
| DEC-AF-04 | 잠금은 코드 수준 상수 하나, 설정·환경 변수로 풀 수 없음, 시나리오 하네스 포함, 단일 코드 변경으로만 해제, 해제 때 L1·L2 재작성 | 리더 결정 (리더 세션 경유) | REQ-AF-011, §C.4 |
| DEC-AF-05 | SPEC은 둘뿐(이 SPEC + IDENTITY-001). IDENTITY-001은 넓히지 않는다 | 리더 결정 (리더 세션 경유) | §I |
| DEC-AF-06 | 권위 = 온라인 control ref(기술) + Issue(업무). 로컬 claim 저장소는 재구축 가능한 캐시·시험 구현 | 리더 결정 (리더 세션 경유) | REQ-AF-002 |
| DEC-AF-07 | Issue 종료가 DONE보다 먼저(E13/E14, 시험 M34). 수리 E10 → E06. Issue당 활성 claim 최대 1, 이력은 같은 키. PR 적체 상한 키 `workflow.agentic_factory.pr_queue.wip_limit`(1 이상, 범위 밖 거절, `0`=끔 금지) | 리더 결정 (리더 세션 경유) | REQ-AF-001·006·007·009 |
| DEC-AF-08 | 명칭: 내부 Coordinator / 표시 leader / 실행자 agent | 리더 결정 (리더 세션 경유) | §A.4 |
| DEC-AF-09 | CLI: `-f/--factory`, 인자형 `--leader <name>`, `-w/--worktree` 유지, agent 진입 `-a/--agent`, 이전 `-l/--leader` 제안 폐기. 남은 4건은 운영자 판단 대기(OD-2) | 리더 결정 (리더 세션 경유) | design §9.0 |
| DEC-AF-10 | K.6 (마): 판단 항목은 묶음 REQ, 전이표·record·template·manifest는 design 표, S52 루프는 plan 진행 원칙 | 리더 결정 (리더 세션 경유) | §C 구조, plan §B |
| DEC-AF-11 | 시험마다 주 AC 하나. 보조 근거는 별도 열이며 충족에 세지 않음 | 리더 결정 (리더 세션 경유) | design §15 |
| DEC-AF-12 | 정제 기본선 C1~C5, 게시 불가 자료는 상태·요약·hash만 | 리더 결정(기본값) (리더 세션 경유) | REQ-AF-013 |
| DEC-AF-13 | 원본은 `.moai/reports`(ignore 규칙 그대로), 공개 안전 정제본만 추적 디렉터리 `.moai/evidence/issue-<N>/`, 커밋 고정 링크, 단계 PREPARED→MASKED→COMMITTED→PUSHED→REMOTE_VERIFIED, AVAILABLE은 원격 조회 뒤 | 리더 결정 (리더 세션 경유) | REQ-AF-019 |
| DEC-AF-14 | rate limit 문서값은 provider 설정 스키마의 기본값으로만, 출처 URL·조회 날짜·"uncollated" 표시. 없으면 UNKNOWN, 보수적 처리. 코드에 숫자 금지 | 리더 결정 (리더 세션 경유) | REQ-AF-014, design §6 |
| DEC-AF-15 | 접수 상태는 정본 S51의 7행을 따름 | 리더 결정 (리더 세션 경유) | REQ-AF-018 |
| DEC-AF-16 | 닫힌 사유 8개: `unclassified-403`, `not-issue-author`, `identity-unknown`, `repository-not-allowed`, `permission-not-granted`, `owner-unconfirmed`, `foreign-closing-keyword`, `op-not-defined` | 구현 사실 — lane-18 보고·리더 확인, `core/state.go`의 임시 상수(이 실행은 검증 안 함) | design §2.1a |
| DEC-AF-17 | Issue 쓰기 권한 정책(§C.3 첫 항목) | '리더 세션의 사용자 채널로 받음, 출처 표기 없음' | REQ-AF-003 |
| DEC-AF-18 | 협업 옵션(§C.3 둘째 항목) | '부모 대화에서 사용자가 확인, 리더 세션 경유' | REQ-AF-003, design §4.1 |
| DEC-AF-19 | 자기 Issue의 제목·마일스톤·다시 열기 = `op-not-defined`, 레이블·담당자 = `permission-not-granted`, `permissions.*` 키는 남의 Issue에만 적용 | 리더 결정 (리더 세션 경유). 공백 기록 | REQ-AF-003, §H |
| DEC-AF-20 | checkpoint 복구 설계(§C.5) | '상위 대화에서 사용자가 승인, 리더 세션 경유' | REQ-AF-002 |
| DEC-AF-21 | 하위 SPEC의 정본 충돌 2·3·4를 정본 쪽으로 고침(Issue 단위 writer 배제, Coordinator를 부여 대상에서 뺌, 부여 기록은 control ref에서 재구축) | 리더 결정 (리더 세션 경유) | 하위 SPEC 0.2.0 |
| DEC-AF-22 | 약화 거절(REQ-AF-008, 0.5.0의 REQ-AF-019)은 lane-20이 effect 경계에서 집행, 사유 타입은 lane-18, 시험은 lane-22 | 리더 결정 (리더 세션 경유) | REQ-AF-008, design §10 |
| DEC-AF-23 | 배정된 agent가 구현 → CI 실패 수리 → 리뷰 대응 → 정상 병합 → Issue 종료 확인까지 책임 유지, PR 인계로 끝나지 않음. leader 역할은 변경 알림·실패 복구·배정 조정·통합 지원이며 agent 일을 대신하지 않음. 대기 기록 필수 필드 여섯 개. PR을 연 뒤 다른 Issue를 맡는 조건(원 Issue 수리 우선, `wip_limit`, 별도 쓰기 worktree, crash-safe 인계). 시험 세 가지(CI 실패 수리, crash 인계, 답장 유실 복구) | 리더 결정 (리더 세션 경유, 0.6.7). 근거는 정본 seq57(운영자가 리더 세션에 직접 전달, Page 원문은 리더도 이 SPEC 작성자도 읽지 않음). 리더 판정: seq57이 S46의 E03·E06 문언을 대체함. 원문 대조 안 함. 리더 판정(리더 세션 경유): 책임과 실행 자리는 따로 있는 두 자원이 아니며 효과로 읽는다 — Holder가 책임을 계속 갖고, 새 Issue 수락은 acquire 검사가 판정 | REQ-AF-009·010·016, AC-AF-004·005·009·010·016 |
| DEC-AF-24 | `moai update`가 manifest에 없는 기존 provider template을 `user_created`로 두고 건너뛰는 것은 알려진 동작이며 결함이 아님(사용자 파일 보존 우선) | 리더 결정 (리더 세션 경유, 0.6.7) | §B.2, design §16 |
| DEC-AF-25 | AF 설정 키 확정. **D1**: AF 키는 새 파일 `.moai/config/sections/agentic-factory.yaml`에만 두고 기존 설정 파일은 바꾸지 않는다. **D2**: claim 저장소 디렉터리는 키 `workflow.agentic_factory.claim_store.dir`이며 기본값이 없다. **D3**: "AF가 설정되었다"는 새 파일이 있고 + `workflow.agentic_factory.enabled: true`이고 + 필수 키가 모두 검증을 통과할 때만이다. 하나라도 어긋나면 설정되지 않은 것으로 다루고 실패 이유를 보인다. **D4**: inbox 키는 webhook 큐와 나눈다 — `workflow.agentic_factory.notify.inbox.{max_doc_bytes,seen_limit,max_id_bytes}`는 별도 키, `workflow.agentic_factory.queue.webhook.*`는 webhook 큐 전용. **D5**: design §6.1의 나머지 이름은 확정(provisional 표시 제거). 기본값 칸은 모든 행이 "—"로 남고, 모든 키가 필수이며 빠지면 거절한다. 이름이 아직 정해지지 않은 행(Tick 주기, 14a/b/c, webhook 비밀 참조)은 이 결정에 들지 않는다(OD-17). Jev 점수는 결정이 아닌 advisory signal로만 적는다: D1 0.97, D2 0.73, D3 0.83, D4 0.56(모호하여 리더가 판단). 현재 이름은 design §6.1 | 리더 결정 (Jev 상담; 리더 세션 경유, 0.6.7) | design §6.1, OD-13, OD-14, OD-17, OD-19 |

## §D 수용 기준 요약

`acceptance.md`에 25개(AC-AF-001~025)가 있다. AC-AF-0NN의 주 REQ는 REQ-AF-0NN(NN = 1~15), AC-AF-016~024의 주 REQ는 각각 REQ-AF-016·017·019·020·021·022·023·024·025, AC-AF-025는 실제 Issue 시나리오이며 주 REQ는 REQ-AF-006이다. REQ-AF-018(준비 상태표)은 AC-AF-012가 보조로 덮는다. 모든 AC는 `EXPECTED_RED`가 기록되기 전까지 채택 전이다.

## §E 추적표 (REQ → AC)

| REQ | 주 AC | 보조 AC |
|---|---|---|
| REQ-AF-001 | AC-AF-001 | AC-AF-025 |
| REQ-AF-002 | AC-AF-002 | — |
| REQ-AF-003 | AC-AF-003 | — |
| REQ-AF-004 | AC-AF-004 | AC-AF-025 |
| REQ-AF-005 | AC-AF-005 | AC-AF-025 |
| REQ-AF-006 | AC-AF-006, AC-AF-025 | — |
| REQ-AF-007 | AC-AF-007 | AC-AF-009 |
| REQ-AF-008 | AC-AF-008 | — |
| REQ-AF-009 | AC-AF-009 | — |
| REQ-AF-010 | AC-AF-010 | — |
| REQ-AF-011 | AC-AF-011 | AC-AF-025 |
| REQ-AF-012 | AC-AF-012 | — |
| REQ-AF-013 | AC-AF-013 | AC-AF-018 |
| REQ-AF-014 | AC-AF-014 | — |
| REQ-AF-015 | AC-AF-015 | AC-AF-023 |
| REQ-AF-016 | AC-AF-016 | — |
| REQ-AF-017 | AC-AF-017 | — |
| REQ-AF-018 | — | AC-AF-012 |
| REQ-AF-019 | AC-AF-018 | — |
| REQ-AF-020 | AC-AF-019 | — |
| REQ-AF-021 | AC-AF-020 | — |
| REQ-AF-022 | AC-AF-021 | — |
| REQ-AF-023 | AC-AF-022 | — |
| REQ-AF-024 | AC-AF-023 | — |
| REQ-AF-025 | AC-AF-024 | — |

## §F 유지되는 결정 (0.5.0에서 이어짐)

### F.1 claim은 slot lease가 아니다

Issue claim은 `internal/factory/slot_lease.go`의 자원 slot lease가 아니다. slot lease는 선언 기한이 지나면 소유 프로세스가 살아 있어도 `--force` 없이 인계하고(REQ-RSL-006), 키가 자원 이름이라 Issue·attempt 식별자가 없다. claim은 Issue 하나에 키 하나다. #1938(죽은 run에 남은 카드)과 #1813(같은 이름의 다른 트리)의 대체 동작은 design §12.

### F.2 활성 판정 정본과 그 한계

- 정본: `homestate.ClassifyOwnerWith`. `live`와 `indeterminate`는 인계를 막고 `lapsed`로 표시, `dead` 또는 안전 정지 기록만 인계(REQ-AF-004).
- **한계(정직한 기록)**: 비교는 문자열 동등뿐이다. Linux 원천은 초 단위 `ps lstart` 문자열이라 **같은 초 안의 PID 재사용은 구분하지 못한다**. 그 위험은 epoch fencing(REQ-AF-001)과 하위 SPEC의 계보 검사가 보완한다. 더 정밀한 원천을 요구할지는 OQ-17이다.
- `FactoryProcessAlive`(PID만)는 claim 판정에 쓰지 않으며 기존 호출처도 고치지 않는다.

### F.3 (폐기) 로컬 claim 저장소 권위

0.5.0의 "로컬 claim 저장소가 권위, Issue 댓글은 거울"은 정본 DELTA1·S51과 충돌해 폐기했다(§C.6). 대체는 REQ-AF-002.

### F.4 Codex 명시 설정 키

`workflow.agentic_factory.codex.enabled`(bool, 미설정 허용). 미설정·`false`는 호출 0건, `true`만 켜짐. 기존 `workflow.codex.*`·`audit.model: codex`는 이 SPEC 단계에서 바꾸지 않고 plan §F.3 순서를 따른다. 읽기는 lane-17 `CodexOptIn`(§B.1).

### F.5 "공식 QA 스킬"

`/moai gate`(같은 HEAD의 lint·format·type·test)로 정하고, 웹·브라우저 표면을 건드릴 때만 `/moai e2e`를 붙인다. 최소 계약: 판정 HEAD 전체 SHA 기록, 결과 `PASS`/`FAIL`/`BLOCKED`, 빈 실행은 PASS 아님, 검사별 명령·종료 코드·분류 기록, 비밀 미출력. 운영자 의도 확인은 OQ-5.

### F.6 선형 이력

측정한 `main` 보호 설정을 바꾸지 않고 `main` 이력을 다시 쓰지 않는다. `required_linear_history`는 꺼져 있으며 AgenticFactory가 켜지도 기대지도 않는다. 켤지는 OQ-6.

### F.7 병합 게이트 D1~D5 (0.3.0 결정 유지)

D1 보호 규칙 불변. D2 Codex 미설정 프로젝트의 리뷰는 공급자 중립 경로(design §11.1). D3 무부채 PASS는 AgenticFactory 경로의 Issue에만, 착지한 카드를 다시 열지 않음. D4 SHA 결합 증거는 코드 후보 브랜치 밖(0.6.0: 정제본 위치는 DEC-AF-13, 원본 위치는 `.moai/reports`). D5 필수 검사 = 보호 규칙의 3개. 근거와 출처는 0.5.0 본문(커밋 `3275706cc9`의 이 파일 §F.7)과 research.md.

### F.8 카드 TTL/renew 대 지속 배정 + epoch

TTL·renew는 활성 힌트, epoch가 권위(design §14). 기한이 지난 뒤 살아 있는 소유자의 renew 처리는 OQ-11이며 그전까지 현행 거절 유지.

## §G 미결 질문

### G.1 운영자 결정 대기 (열린 채로 둠)

| ID | 결정 | 막는 것 |
|---|---|---|
| OD-1 | S52 임계값·표본 설계의 승인(정본이 "제안과 승인 상태를 구분"하라고 요구) | REQ-AF-025 비교 기준 |
| OD-2 | CLI 4건: 버전 선언 방식, 새 문법의 대상 leader 플래그 이름, `-l/--lane` 장기 처분, 기본 문법 전환 시점. **충돌 미해결**. 어느 답이든 명시적 버전 선언, 호환 오류, 역할 뒤바뀜 0·조용한 흡수 0 회귀 시험, `--` 경계가 필요하다 | REQ-AF-022 |
| OD-3 | DEC-AF-01 ② 문언과 정본 [0]§2의 잔여 충돌: 실제 운용에서는 리뷰·감사가 Issue마다 새 세션이라 요청별 운영자 인증이 카드마다 사람 승인이 된다. 원문 뜻이 "요청마다"인지 "run마다"인지와 이 충돌의 수용 여부 | 하위 REQ-AFI-002 |
| OD-4 | 하위 SPEC 중계 경로 잔여 위험(하위 OQ-AFI-1)의 운영자 위험 수용. 수용 기록 없이는 잠금이 풀리지 않는다 | REQ-AF-011 해제, plan M7 |
| OD-5 | 추가 감사 라운드(새 감사는 지금 승인되지 않았다) | 이 판의 채택 |
| OD-6 | 실제 Issue 시나리오(AC-AF-025)에 쓸 Issue 한 쌍과 보호된 `main` 실제 병합·Issue 종료 승인(keep-set) | AC-AF-025 |
| OD-7 | **대상**: `workflow.agentic_factory.pr_queue.wip_limit` 키가 설정되지 않았을 때 새 Issue 배차를 어떻게 할지. **선택지**: (가) 새 배차를 `not-configured`로 막는다. (나) 설정 단계에서 명시 값을 요구하고, 값이 없으면 factory 시작 자체를 거절한다. (다) 문서에 적은 보수적 기본값을 쓰되, 그 값이 운영자가 정하지 않은 기본값임을 설정 스키마와 관찰 보기에 표시한다. **근거**: 정본은 상한 키와 범위 검사(1 이상, `0`=끔 금지)만 정하고 미설정 동작은 정하지 않는다. 어느 쪽이든 새 정책이며, 이 SPEC은 숫자를 지어내지 않는다. (가)·(나)는 배차가 멈출 수 있고 (다)는 근거 없는 숫자가 들어간다. 결정 전에는 미설정 사례를 채택 근거로 쓰지 않는다. 0.6.7: REQ-AF-009의 "PR을 연 agent가 다른 Issue를 맡는 조건"도 이 키를 쓰며, 미설정일 때 그 조건이 어떻게 판정되는지도 이 결정에 달려 있다. "미설정이면 새 Issue를 주지 않는다" 같은 보수적 기본값을 이 SPEC이 정하지 않는다. **구현 관찰(0.6.7, 요구사항 아님)**: 구현은 `wip_limit` 미설정을 거절한다 — 새 Issue를 받지 않는다. 출처: 'lane-19 comparison, relayed by the leader'. 이 작성자는 그 코드를 다시 읽지 않았다. **이 SPEC은 이 동작을 기본값으로 채택하지 않으며, 미설정 기본값 질문 자체는 여전히 열려 있다** | REQ-AF-007, REQ-AF-009, AC-AF-007 적체 사례, AC-AF-009 (g) |
| OD-8 | **대상**: `workflow.agentic_factory.checkpoint.{max_per_issue,retention,max_manifest_bytes}`가 설정되지 않았을 때 새 checkpoint를 어떻게 할지. **선택지**: (가) 새 checkpoint를 `not-configured`로 거절한다. (나) 설정 단계에서 명시 값을 요구한다. (다) 문서에 적은 보수적 기본값을 쓰되, 운영자가 정하지 않은 기본값임을 표시한다. **근거**: 정본은 checkpoint의 단계·상태·금지 사항을 정하지만 저장 상한 값과 미설정 동작은 정하지 않는다(숫자 자체는 OQ-22). (가)·(나)는 복구 지점이 생기지 않을 수 있고, (다)는 근거 없는 숫자가 들어간다. 결정 전에는 미설정 사례를 채택 근거로 쓰지 않는다 | REQ-AF-002 (c)~(e), design §4.6 |
| OD-9 | **대상**: `workflow.agentic_factory.intake.max_concurrent`가 설정되지 않았을 때 새 접수·준비 요청을 어떻게 할지. **선택지**: (가) 새 준비 요청을 `not-configured`로 거절한다. (나) 설정 단계에서 명시 값을 요구한다. (다) 문서에 적은 보수적 기본값을 쓰되, 운영자가 정하지 않은 기본값임을 표시한다. **근거**: 정본 S51은 동시 준비 상한이 있어야 한다는 것과 포화를 수락으로 숨기지 않는다는 것만 정하고, 미설정 동작과 숫자는 정하지 않는다(숫자 자체는 OQ-22, 포화 처리는 OQ-21). (가)·(나)는 접수가 멈출 수 있고 (다)는 근거 없는 숫자가 들어간다. 결정 전에는 미설정 사례를 채택 근거로 쓰지 않는다 | REQ-AF-012, AC-AF-012, design §4.2 |
| OD-10 | **대상**: collector 기준 스냅숏(baseline)이 claim 문서 한도(256 KiB, base64 뒤 본문 약 195 KB — **가정값이며 이 실행의 측정이 아님**)를 넘을 때의 규칙. **현재 동작(보고)**: Ack 오류 뒤 같은 변경이 매 Tick마다 끝없이 되풀이된다. **선택지**: A. 추적 투영(projection)을 줄인다. B. 키마다 문서를 나누고 색인 문서를 둔다. C. 키 → 정규형 hash 표로 바꾼다. D(보조). gzip으로 압축한다. E. 넘치면 멈추고 차단을 기록한다 — 이를 위해 새 사유 값을 만들면 design §2.1a의 닫힌 사유 집합(lane-18이 대조한 사유 문자열 8개 묶음 포함)이 **바뀐다**. **결정 입력**: (1) 추적하는 수집 항목의 최대 개수, (2) 필드 단위 변경 내용이 필요한지, (3) 멈춘 상태를 누가 풀지. **근거·출처**: 'reported by lane-20, relayed by the leader; locations confirmed by the leader'. 위치: `claim/docs.go:18`(한도), `forge/collector.go:24`(수집기). 커밋 `51edce26a8`. 용량 숫자는 모두 가정값이다. 이 SPEC은 어느 선택지도 채택하지 않는다 | REQ-AF-017, design §4.8 `CollectorCheckpoint`, §H 공백 (e) |
| OD-11 | **대상**: checkpoint push를 다시 시도하는 횟수. **사실**: design §4.6 `committed-unacked` 재개 규칙(K2)은 원격 ref를 조회한 뒤 **한 번** push한다고 적었다(0.6.1까지의 문언). 코드는 원격에 없다고 확인되면 최대 3번 다시 시도한다(커밋 `e6499b731a`, `checkpoint.go:28`; lane-18 보고, 리더 인용, 이 실행은 읽지 않음). 정본은 재시도 횟수를 정하지 않는다(정본 댓글에 checkpoint push 규칙 없음, §H 공백 (f)). **선택지**: (가) 조회 뒤 한 번만 push한다. (나) 최대 횟수를 명시한 제한 재시도로 하고, 매번 그 전에 멱등성 확인(같은 ref·SHA의 ACK 존재 여부)을 한다. (다) 재시도를 공용 API 예산 규칙 아래 둔다. **결정 입력**: 중복 push 금지(design §4.6), 실패 형태(부분 ACK 등 결과 불명), 시도 한 번의 비용. 이 SPEC은 어느 선택지도 채택하지 않는다 | REQ-AF-002 (c)~(e), design §4.6 |
| OD-12 | **대상**: `operator-hold`를 PARKED 사유로 둘지. **사실**: 정본 S46(댓글 6489610367)에는 CLI14 행(125행, 두 칸 원문 "pause/resume"과 "새 배차만 제어, 강제 종료 0, HOLD/보호 우회 0")만 있고, 이에 대응하는 PARKED 사유는 없다. 정본 다른 댓글(6487472895, 158행)의 "diagnostic HOLD"는 수리 재발급을 멈추는 다른 개념이며 E20으로 재개한다. 코드는 `on-hold`를 쓴다(lane-18 보고, 리더 전달). 이 SPEC은 0.4.0부터 `operator-hold`를 design §2.1a에 두었고 design §8 `moai factory hold`, acceptance §A.1 시나리오 묶음 07이 이 값을 쓴다. **선택지**: (가) `operator-hold`를 PARKED 사유로 둔다. (나) 배차 일시 정지로만 모형화하고 Issue별 상태는 두지 않는다. (다) `on-hold`를 계약 밖 구현 내부 상태로 둔다. **결정 입력**: PARKED overlay 의미(단계·owner 불변)와 E 전이(E19·E20)에 주는 영향, 재개 권한이 누구에게 있는지, 닫힌 사유 집합(design §2.1a)의 변경 여부. 이 SPEC은 어느 선택지도 채택하지 않으며, 결정 전까지 위 세 곳은 현재 문언을 둔 채 이 결정을 가리킨다 | REQ-AF-001, design §2.1a·§2.3·§8, acceptance §A.1 |
| OD-13 | **대상**: 로컬 claim 저장소(캐시)의 디렉터리 경로 관례. **사실**: 리더가 정본 사본(Issue #1948 댓글 사본)에서 claim 저장소 경로를 찾았으나 디렉터리 관례가 없었다. 소스와 설정에도 없다. lane-21 보고: `claim.Open`에는 시험이 아닌 호출자가 없다. 출처: 'leader decision, relayed via the leader session; lane-21 report relayed by the leader'. 이 실행은 그 검색과 코드를 다시 하지 않았다. **막힌 쪽**: lane-22의 웹 콘솔 연결(대기 중). 이 SPEC은 경로를 지어내지 않는다. **해결(0.6.7, DEC-AF-25 D1·D2)**: 디렉터리는 키 `workflow.agentic_factory.claim_store.dir`(필수, 기본값 없음)로 정하고, 새 파일 `.moai/config/sections/agentic-factory.yaml`에 둔다. 경로 값 자체는 이 SPEC이 정하지 않는다(기본값 없음). 이 행은 기록으로 남긴다. 남은 것: 그 파일의 template 배포는 별도 결정이다(OD-19). 출처: 'leader decision (Jev consultation; relayed via the leader session)' | REQ-AF-002 (a), design §1 로컬 캐시, design §6.1 |
| OD-14 | **대상**: "AF가 설정되었다"를 판정하는 설정 키. **사실**: 정본 사본에서 `agentic_factory:` 키를 찾았으나 그런 판정 키가 없었다. `internal/config`에도 AF 판정 키가 없다(lane-21 보고). design §6의 `workflow.agentic_factory.enabled`는 이 SPEC의 **제안** 행이며 lane-17 읽기 코드에 없다. 그 키가 이 판정 키인지도 정해지지 않았다. 출처: 위 OD-13과 같다. **막힌 쪽**: lane-22의 웹 콘솔 연결(대기 중). 이 SPEC은 키를 지어내지 않는다. **해결(0.6.7, DEC-AF-25 D1·D3)**: "AF가 설정되었다"는 새 파일 `.moai/config/sections/agentic-factory.yaml`이 있고 + `workflow.agentic_factory.enabled: true`이고 + 필수 키가 모두 검증을 통과할 때만이다. 어긋나면 설정되지 않은 것으로 다루고 실패 이유를 보인다. 판정은 두 단계다(리더 결정, 리더 세션 경유): enabled가 없거나 false → AF 미설정(오류 아님). enabled가 true인데 다른 필수 키가 없거나 범위 위반 → 검증 오류(키 이름 포함)이며 그 결과도 AF 미설정으로 표시. design §6.1 표에 `enabled` 행을 더했다. design §6 본 표의 같은 키 행이 기본값 `false`로 어긋났던 것은 리더 결정(리더 세션 경유)으로 해결했다: 두 표 모두 기본값 `—`·필수·bool 외 거절이며, 키가 없거나 `false`이면 오류가 아니라 설정되지 않음으로 다루고 이유를 보인다. 이 행은 기록으로 남긴다. 남은 것: template 배포(OD-19). 출처: 'leader decision (Jev consultation; relayed via the leader session)' | REQ-AF-021, design §6, design §6.1 |
| OD-15 | **대상**: 수집 cursor를 어디에 둘지(ref 모드의 권위 저장소 안인지 로컬 저장소인지). **사실**: cursor를 ref 권위 저장소에 chunk 문서로 두면 Ack 한 번마다 커밋이 chunk 수 + 1개 생기고, `DeleteDoc`은 git 이력에서 내용을 지우지 못하므로 수백 KB 스냅숏이 쌓일 수 있다. 출처: lane-19 리뷰 4b, 'relayed by the leader'. 이 실행은 그 코드와 리뷰를 읽지 않았다. 기록만 하며 이 SPEC은 어느 쪽도 채택하지 않는다 | REQ-AF-002 (a), REQ-AF-017, design §4.8 `CollectorCheckpoint`, OD-10 |
| OD-16 | **대상**: lane-19의 정본 읽기 대조에서 나온 항목들. **열린 부분은 기록만 하며, SPEC 변경은 운영자 승인 뒤다.** **열림**: (i) S46 E24: 정본 표(Issue 사본 192행·2760행)는 "변경 없이 PARKED contract-change, E18 뒤 SPEC 단계"인데 정본 그림(605행·2896행)은 FIX → SPEC을 바로 그린다. (ii) E20 재개는 단계를 바꾸지 않는다. E20의 사유별 재개 guard는 정해지지 않았다. (iii) 정본 침묵: E19 허용 사유 집합(11개 값), E10·E15·E16·E18·E23의 PARK 사유. **리더 결정으로 정해진 부분(0.6.7, 정본 대조 안 함)**: BUILDING/FIX에서 SPEC으로 돌아가는 간선은 E18의 '채택(adopt)' 간선으로 정한다 — BUILDING/FIX + PARKED `contract-change`에서 승인된 revision을 채택하면 BUILDING/SPEC으로 가고, owner·receipt·epoch는 그대로 둔다. 채택 간선은 revision 승인 증거가 있어야만 일어난다(design §2.2 E18). 구현: lane-18 `core/transition.go` 137~138행, runner `AdoptRevision`. 출처: 'leader decision, relayed via the leader session; lane-19 comparison relayed by the leader'. 이 작성자는 그 코드를 다시 읽지 않았고, 정본 원문과의 대조는 하지 않았다. **출처(열린 부분)**: 'lane-19 read comparison, relayed by the leader'. 행 번호는 전달받은 그대로이며 이 실행은 다시 읽지 않았다(0.6.7 첫 판에 적었던 `core/transition.go:139`는 PARK만 한다는 전달도 다시 읽지 않았다). §H 0.6.3~0.6.5의 S46 대조 관찰과 겹치는 부분이 있으며, 서로 어긋나는지는 판단하지 않는다 | REQ-AF-016, design §2.2 |
| OD-17 | **대상(DEC-AF-25 뒤 남은 것만)**: (i) Tick 주기 키(`workflow.agentic_factory.loop.tick_period`, provisional)의 이름과 값, (ii) webhook 비밀 참조 키(`...webhook.routes[].secret_ref`, provisional)의 이름, (iii) 그 비밀 참조를 어디서 읽을지(비밀의 위치), (iv) `forge.Adaptive` 세 필드의 검증 규칙 채택. **해결된 부분**: lane-20 새 필드 `intake.list_limit`·`monitor.pr_list_limit`(커밋 `af72c90458`)는 리더 결정으로 확정됐다(필수, 정수 >= 1). 예전 하드코딩 14a/b는 이 두 필드로 바뀌었고 14c는 제거됐으므로 열린 하드코딩은 없다. 또 `MaxChunks`(chunk 상한, `0` = 나누기 끔)·`SeenLimit`(webhook 중복 제거 창, lane-20 커밋 `527ca7e3a7`)을 포함한 나머지 design §6.1 이름은 DEC-AF-25 D5로 확정됐다(기본값 없음, 필수). 이 실행은 그 커밋을 읽지 않았다. 숫자 값과 "`0` = 끔" 의미가 design §6의 "`0`을 끔으로 읽지 않음" 원칙(`wip_limit`)과 어떻게 맞는지는 기본값이 없으므로 값 결정과 함께 남는다. **구현 관찰(0.6.7, 요구사항 아님)**: 구현은 재시도 상한(retry cap)에 `0`을 허용하지 않는다(1 이상이어야 함). 출처: 'lane-19 comparison, relayed by the leader'. 이 작성자는 그 코드를 다시 읽지 않았다. 이 SPEC은 이 값 규칙을 채택하지 않았고 design §2.1c 재시도 상한 필드의 값은 여전히 열려 있다. 0.6.7: 키 표는 design §6.1에 있다(확정 행은 'confirmed (leader decision)', 위 (i)·(ii) 행 `loop.tick_period`·`secret_ref`만 provisional·'리더 결정 대기'; ListLimit 두 행의 근거는 lane-20 `af72c90458`, 리더 전달, 이 작성자는 다시 읽지 않음). 모든 키는 새 파일 `.moai/config/sections/agentic-factory.yaml`에만 둔다(DEC-AF-25 D1). webhook 경로 서명 비밀은 커밋되는 설정에 값이 아닌 참조로만 두며, **비밀을 어디에 둘지는 이 SPEC 소유자의 결정이고 열려 있다**. 표의 근거는 lane-20 사실 통지(커밋 `efeb3dc87c`, 브랜치 `WT-agentic-monitor-effect-reconcile`)와 lane-20의 정정이며, 리더가 전달했고 이 작성자는 다시 읽지 않았다 | REQ-AF-021, design §6, design §6.1 |
| OD-18 | **대상(후보, 기록만)**: REQ-AF-010의 역할 쓰기 경계 — leader(모니터)가 agent 일을 대신하지 않는다 — 를 effect 경계에서 집행할지. **사실**: 코드는 이 경계를 직접 거절하지 않는다. 출처: 'lane-19 comparison, relayed by the leader'. 이 작성자는 그 코드를 다시 읽지 않았다. 집행 위치는 정해지지 않았다. 이 SPEC은 어느 쪽도 채택하지 않는다 | REQ-AF-010, AC-AF-010 (f) |
| OD-19 | **대상**: `.moai/config/sections/agentic-factory.yaml`의 template 배포. **사실**: DEC-AF-25 D1로 파일 이름과 위치는 정해졌지만, 그 파일을 `internal/template/templates/` 아래에 배포 파일로 둘지는 정해지지 않았다. template 배포 파일을 만들면 `moai update`가 사용자 프로젝트에 그 파일을 쓰게 되므로, 이 결정 전에는 **template 배포 파일을 만들지 않는다**. 출처: 'leader decision (Jev consultation; relayed via the leader session)'. 이 SPEC은 어느 쪽도 채택하지 않는다 | REQ-AF-021, design §6.1, §B.2 |

### G.2 그 밖의 미결 질문

| ID | 질문 | 막는 것 |
|---|---|---|
| OQ-2 | `run`의 범위와 실행 자리 `slot`이 자원 slot lease와 이름이 겹치는 문제 | REQ-AF-010·022 |
| OQ-5 | "공식 QA 스킬" 결정(§F.5)이 운영자 의도와 맞는가 | REQ-AF-007 |
| OQ-6 | `required_linear_history`를 켤 의도가 있는가, `git-strategy.yaml` live 값과 로컬 규칙의 차이 | REQ-AF-008 문구 |
| OQ-7 | egolite·aside의 연결 실패 문구·종료 코드, aside 로그인 감지 명령 미측정 | REQ-AF-021 |
| OQ-11 | 기한이 지난 뒤 살아 있는 소유자의 renew 처리 | REQ-AF-004 세부 |
| OQ-12 | 제안 `moai factory status`와 기존 명령 이름 충돌 | REQ-AF-022 |
| OQ-13 | `CI=true`에서 aside 처리, aside 권한 기본값 | REQ-AF-021 |
| OQ-15 | 필수 검사 집합 확장 여부 | REQ-AF-007 |
| OQ-17 | 같은 초 PID 재사용을 구분하는 더 정밀한 시작 시각 원천을 요구할지 | REQ-AF-004 |
| OQ-18 | 브라우저 3선택 화면 표시 | REQ-AF-021 표시 범위 |
| OQ-20 | 명령 제거 단계(S4) 판정 형식 | REQ-AF-022 S4 |
| OQ-21 | 큐 포화 시 새 접수를 받고 대기시키는지 거절하는지(정본은 "포화를 수락으로 숨기지 않는다"까지만 정함) | REQ-AF-012 |
| OQ-22 | 수치 설정값: `intake.max_concurrent`, `intake.stage_timeout`, polling 간격, backoff 배수·횟수, 예산 reserve, 댓글 묶음 window, checkpoint 상한·보존 기간. 모두 **OPEN**이며 이 SPEC은 숫자를 정하지 않는다 | REQ-AF-002·012·014·023 |
| OQ-23 | GitLab을 이 SPEC 범위에 넣을지, 정본 [1]§19 순서 6("승인된 adapter·GitLab·부하를 확대")으로 미룰지 | REQ-AF-014 GitLab 행 |
| OQ-25 | 자기 Issue 레이블·담당자 변경을 `op-not-defined`가 아닌 `permission-not-granted`로 거절하는 리더 교정은 받았다. 이 동작들을 언젠가 허용할지 | REQ-AF-003 |

해소된 0.5.0 질문: OQ-1(시나리오 Issue → OD-6로 옮김), OQ-3(lane-22 대응표 — 정본 DELTA4와 리더 배정으로 대체), OQ-4(여섯 범위 → 레인 소유 표 design §10), OQ-8(단일 owner → DEC-AF-07), OQ-9(PR #1937 → plan §F.3 순서 안에서 판정), OQ-10(환경 목록 → design §5.1), OQ-14(증거 ref → DEC-AF-13), OQ-16(`slots` 상한 → OQ-22에 합침), OQ-19(하위 OQ-AFI-1). 0.6.1에서 OQ-24(`wip_limit` 미설정 동작)는 OD-7로 옮겼다.

## §H 확인 공백

- lane-18이 밝힌 공백(리더 전달): 실제 GitHub 원격 조회와 댓글 멱등성은 가짜(fake)로만 시험했다. CONTEXT06 "새 대상의 증거가 필요하다"는 **호출자 규칙**이며 모형이 집행하지 않는다.
- 아래 다섯 항목 (a)~(e)는 **관찰로만 기록한 공백**이다. 출처: 'reported / measured by the SPEC writer on 2026-10-29'. 코드 쪽 사실은 리더가 lane-18·lane-20에 따로 확인한다. 이 SPEC은 감사받지 않았다.
  - (a) lane-18 `core/authz.go:96`의 `"unknown-operation"`과 `core/authz.go:90`의 `"allow"`는 `AuthzVerdict` 값이며 `BlockReason`이 아니다. 이 SPEC의 닫힌 사유 집합(design §2.1a)에는 없다. `BlockReason` 상수 26개는 모두 그 집합 안에 있다(측정).
  - (b) design §2.1a의 PARKED 전용 사유 6개(`dependency`·`stop_unknown`·`unknown-effect`·`pending-effect`·`contract-change`·`start_unknown`, 0.6.2부터 정본 S46 철자)는 lane-18 `core/state.go`에 없다(측정).
  - (c) lane-18 `core/state.go:9-17`의 상태 상수는 0.5.0 모형(`queued`…`on-hold`)이고, design §2.1의 모형은 READY…PARKED다(측정).
  - (d) checkpoint 상태: 보고된 구현(`e6499b731a`)은 넷(`local`·`pending`·`remote-durable`·`recovered`)이고 design §4.6은 일곱을 적는다. `committed-unacked`·`acked-manifest-unverified`·`durable-link-pending`은 보고에 없다(보고. 이 실행은 그 커밋 내용을 읽지 않았다).
  - (e) collector 스냅숏이 claim 문서 한도(256 KiB, base64 뒤 본문 약 190~195 KB — 보고된 값이며 가정)를 넘으면 저장되지 않고, Ack 오류 뒤 같은 변경이 매 Tick마다 되풀이된다는 lane-20 보고(`51edce26a8`, `claim/docs.go:18`·`forge/collector.go:24`는 리더가 위치 확인)가 있다. 이 SPEC에는 그 경우의 규칙이 없다. 규칙은 운영자 결정 대기 OD-10(§G.1)으로 둔다(보고. 이 실행은 그 코드를 읽지 않았다).
- 아래 (f)~(k)도 **관찰로만 기록한 사실**이다(0.6.2, (k)는 0.6.3). 각 항목에 출처를 따로 적는다. 이 SPEC을 코드에 맞추거나 코드 변경을 요구하지 않는다.
  - (f) design §4.6의 checkpoint 상태 일곱 개와 단계 K1~K4는 이 SPEC이 DEC-AF-20을 풀어 쓴 것이며 정본에는 없다. 출처: 'measured by the SPEC writer on 2026-10-29' — 정본 댓글 사본 두 벌(10개 묶음, 11개 묶음)에서 `remote-durable`·`committed-unacked`·`durable-link`를 찾은 결과 0건. 같은 검색에 `stop_unknown`·`CLI14`를 넣으면 댓글 세 개가 나와 검색이 작동함을 확인했다. 정본에서 checkpoint는 E25 guard와 collector checkpoint(`CollectorCheckpoint`)로만 나온다.
  - (g) 코드의 checkpoint 상태 넷(`local`·`pending`·`remote-durable`·`recovered`)과 필드로 위 일곱 상태와 같은 정보를 나타낼 수 있다(`checkpoint.go:125-146`). 출처: lane-18 보고. 이 실행은 그 코드를 읽지 않았다.
  - (h) 코드의 상태 모형은 계약 0.4.0 기준에 머물러 있고 정본 S46과 다르다. 리더가 lane-18에 맞춤 방안을 요청했다. 출처: 리더 보고. 참고로 이 SPEC 0.4.0(커밋 `8dc835756`) design §2.1은 상태 아홉 개(`queued`…`on-hold`)를 적었고, 이는 (c)에서 잰 `core/state.go:9-17`의 이름과 같다('measured by the SPEC writer on 2026-10-29', `git show 8dc835756:…/design.md`).
  - (i) **[0.6.3에서 닫힘]** 0.4.0 design §2.1은 `ready`를 "writer 인계와 coordinator 인계가 모두 있다"로, `verifying`을 "후보 HEAD 하나에 대해 검사·리뷰·감사·QA를 모은다"로 정의했다('measured by the SPEC writer on 2026-10-29', `git show 8dc835756:…/design.md`). 0.6.2의 이름 기준 대응은 두 값을 모두 `BUILDING/IMPLEMENT`로 보냈고, 인계 증거가 있는 `ready`를 E03(PR 인계 → `REVIEWING`)과 다르게 다뤘다. 0.6.3의 증거 기준 규칙(design §2.3, 'leader decision, relayed via the leader session')은 이름과 상관없이 인계 증거 둘이 있으면 `REVIEWING`으로 보내므로 이 어긋남은 남지 않는다. 기록으로만 남긴다.
  - (j) **[0.6.3에서 닫힘]** 0.6.2 대응표에서는 `landing`이 인계 receipt도 없고 진행 중인 병합 의무도 없는 경우가 정해지지 않았다. 0.6.3 규칙은 그 경우를 "그 밖의 모든 경우" → `BUILDING/IMPLEMENT` + 검사 증거로 정한다. 기록으로만 남긴다.
  - (k) 0.6.3 규칙에서 인계 증거 둘과 `merged=true`가 함께 관측될 때 어느 행이 먼저인지는 리더 결정 문언에 없다. 또 이 규칙이 이름 붙인 증거 둘(writer 인계 receipt, leader 인계 수락)은 정본 E03 guard 목록(S46 49행: PR 원격 H/base·계약·SPEC·task 결과·검사 상태·원 writer·leader 수락 receipt·quiescence·보존 확인)의 일부다. 둘 다 관찰로만 적는다('measured by the SPEC writer on 2026-10-29').
- **lane-18 잠정 값 (0.6.3, 관찰)**: lane-18 커밋 `af80e0bdec`(core S46 모형, E01~E27 표, `FromLegacy`)에서 lane-18이 정본이 정하지 않은 범위에 **잠정으로** 둔 값이다. 출처: 'reported by lane-18, relayed by the leader; commit existence/race test confirmed by the leader'. 이 실행은 그 커밋 내용을 읽지 않았고 `git cat-file -t af80e0bdec` → `commit`만 관측했다. 번호는 리더가 붙인 그대로이며 **4번 항목은 전달되지 않았다(item 4 not supplied)**. 0.6.4에서 리더가 이 표기가 맞다고 확인했다. 4번은 별도 지시로 나갔다(전달 문언 그대로: 'the leader sent that instruction to lane-18, separately to lane-17'). 그 내용은 이 SPEC에 전달되지 않았다. 정본 대조 줄은 S46(댓글 6489610367) 본문에서 그대로 옮겼다('measured by the SPEC writer on 2026-10-29'). 아래 어느 항목도 권고나 코드 변경 요구가 아니다.
  - (1) E05 REVIEWING → BUILDING의 하위 단계를 `IMPLEMENT`로 둔다. 정본 51행 E05 도착 칸은 "새 epoch·attempt로 BUILDING, 보존후 재검증"이며 하위 단계를 적지 않는다. 정본이 정하지 않은 범위가 맞다. design §2.2 E05도 하위 단계를 적지 않는다.
  - (2) E18 overlay 사유를 `contract-change`로 둔다. 정본 64행 E18 도착 칸은 "현 phase + PARKED, 새 attempt에 승인 revision 적용"이며 사유 이름이 없다. 다만 정본 70행 E24 도착 칸은 "변경없이 PARKED contract-change; E18 후 SPEC단계"로 같은 이름을 E18과 이어 쓴다. 이 SPEC의 design §2.1a·§2.2 E18 행은 이미 `contract-change`를 적는다. E18 행 자체는 사유를 적지 않지만 정본 전체가 이 범위에서 완전히 침묵하는 것은 아니다.
  - (3) E24는 BUILDING/FIX에서만 허용한다. **충돌 관찰**: 정본은 이 범위에서 침묵하지 않는다. 정본 70행 E24의 출발 칸이 "BUILDING/FIX"다. 값은 정본과 같고, 어긋나는 것은 "정본이 정하지 않은 범위"라는 분류다.
  - (5) 종결 때 overlay를 해제한다. 정본에 종결 시 overlay 해제를 적은 문장은 없다. 다만 정본 61행 E15 도착 칸은 "CANCELLED; 정리불명은 current+PARKED"로, 정리가 불명이면 종결하지 않고 현 단계 + PARKED로 둔다. 62행 E16 도착 칸도 "CANCELLED 또는 MERGING unknown"이다. **충돌 관찰**: 종결 간선과 PARKED의 관계에 정본이 일부 말하고 있으므로 완전한 공백은 아니다. 두 규칙이 실제로 어긋나는지는 판단하지 않는다.
  - (7) E19 사유는 호출자가 준 분류다. 정본 65행 E19 guard 칸은 "실제오류분류", 도착 칸은 "phase + PARKED, 정상 재개조건 기록"이다. **충돌 관찰**: 정본은 guard로 실제 오류 분류를 요구하므로 이 범위에서 완전히 침묵하지 않는다. 호출자가 준 분류가 그 guard를 채우는지는 판단하지 않는다. design §2.2 E19 guard는 "§4.5 분류 결과"다.
  - (6) 정본이 이름 붙인 PARKED 사유 여섯 개 밖에서 overlay에 쓸 수 있는 사유는 감사받지 않은 이 SPEC 0.6.1 design §2.1a의 P 열에서 가져왔다. **순환이 있다**: 코드는 감사받지 않은 SPEC에서 값을 가져왔고, 이 SPEC은 그 코드를 다시 사실로 적고 있다. 따라서 이 값들은 어느 쪽에서도 독립 근거가 되지 못한다. 0.6.1 P 열(`git grep -n -E … 5b7035403e -- .moai/specs/SPEC-AGENTIC-FACTORY-001/design.md`)에는 값 21개가 있다. 그 가운데 `stop-unknown`·`start-unknown`은 0.6.2에서 정본 철자 `stop_unknown`·`start_unknown`으로 바뀌었고, `operator-hold`는 0.6.2에서 OD-12 대기가 됐다. 코드가 어느 철자를 쓰는지는 이 실행이 읽지 않았다.
- **리더 결정 사실 (0.6.3, 관찰)**: PARKED 동안 단계를 바꾸는 간선은 거절한다. 예외는 E20 재개, 종결 간선, 사유 갱신이다. 출처: 'leader decision, relayed via the leader session'. 정본 대조: S46 66행 E20은 "overlay 해제, phase와 owner 임의변경 없음"만 적는다. 정본 댓글 6487471042 160행은 "`PARKED`는 이유·재개 guard·담당·시각이 있는 overlay이며 별도 업무 단계가 아니다"와 "아래 없는 전이는 rejected receipt를 남기고 변경하지 않는다"를 적는다. PARKED 중 단계 변경 간선을 거절한다는 문장은 정본에 없다.
- **리더 정정 (0.6.4, 관찰)**: 출처: 'leader correction, relayed via the leader session; the leader read S46 lines 49, 61, 65, 70 directly'. 이 실행은 그 판단을 다시 하지 않았고 아래 코드 쪽 지시도 검증하지 않았다. 권고나 코드 변경 요구가 아니다.
  - 리더는 위 lane-18 잠정 값 가운데 (3) E24, (5) 종결과 overlay 해제, (7) E19를 "정본이 정하지 않은 범위"로 잘못 분류했다고 정정했다. 정본은 그 셋에서 침묵하지 않는다: S46 70행(E24 출발 칸 "BUILDING/FIX"), 61행(E15 "CANCELLED; 정리불명은 current+PARKED"), 65행(E19 guard "실제오류분류"). 위 (3)·(5)·(7)의 충돌 관찰과 같은 줄이다.
  - 리더가 lane-18에 보낸 코드 쪽 지시(리더 보고, 미검증): E15에서 정리가 불명이면 CANCELLED 대신 현 단계 + PARKED를 유지한다. E19 사유는 실제 오류 분류에서 파생하고 그 분류와 대조해 검증한다. E03 guard 증거 키가 S46 49행 목록 전체를 덮는다. E24는 정본 근거 값으로 표시한다. 이 SPEC은 이 지시를 사실로만 적는다.
  - 위 "리더 결정 사실"(PARKED 동안 단계를 바꾸는 간선은 거절, 예외 E20 재개·종결 간선·사유 갱신)은 정본에 없는 리더 결정이며 기록한 문언 그대로 둔다.
- **lane-18 커밋 `52363f7ac7`과 리더 결정 (0.6.5, 관찰)**: 출처: 'reported by lane-18 / leader decision, relayed via the leader session; leader-confirmed race test ok'. 정본 근거 여부는 여기서 판단하지 않는다. 권고나 코드 변경 요구가 아니다.
  - (A1) E03 guard가 이제 S46 49행의 아홉 항목을 모두 덮는다(lane-18 보고, 미검증).
  - (A2) E15·E16에 결과별 guard가 있다. 정리 결과가 불명이면 현 단계 + PARKED(lane-18 보고, 미검증).
  - (A3) E19 사유가 분류 집합 `core.ClassifiedReasons` 밖이면 거절한다(lane-18 보고). 집합의 값 11개는 이 실행이 읽었다('measured by the SPEC writer on 2026-10-29'): `git grep`으로 이 트리에서 커밋 `52363f7ac7`의 `internal/agenticfactory/core/park.go:107-110`(`classified` 목록)과 `state.go:53-75`(문자열 상수)를 읽었고, lane-18 worktree `af-core-claim-runner`의 체크아웃 파일도 grep으로만 읽어 같은 줄을 봤다(그 트리에 git 실행 없음). 값: `not-installed`, `connection-failed`, `not-logged-in`, `not-configured`, `auth-failed`, `permission-denied`, `rate-limited`, `unclassified-403`, `capacity-short`, `budget-exhausted`, `store-unavailable`. 11개 모두 design §2.1a에 같은 문자열로 있다. 거절 동작 자체와 경쟁 시험은 읽지 않았다.
  - (B1) 리더 결정: 옛 상태를 진척 기록으로 남기고, `FromLegacy`로 파생한 새 단계·하위 단계·PARK를 함께 기록한다. 이유: 새 모형만으로는 옛 상태의 세밀한 진척을 복원할 수 없다.
  - (B2) 리더 결정: 병합 게이트 사유(design §2.1a "쓰는 곳" G)를 overlay에서 **단계를 유지하는 게이트 기록**으로 옮긴다. **동작이 바뀐다**: 지금까지의 'Resume after blocked'(차단 뒤 재개)가 '같은 단계에서 새 증거로 다시 검증'이 된다. 리더 결정 사실로만 적는다.
- **설정 읽기 커밋의 계보와 `wip_limit` 읽기 (0.6.6, 관찰)**: 출처: 'leader direct reading of the settings-tree log and diff, relayed via the leader session'. 권고나 코드 변경 요구가 아니다.
  - 리더 판독: `6c833ee4b4`(`pr_queue.wip_limit` 읽기)와 `8920c0ecf8`은 `5c278d8b51`의 조상이다. 리더가 lane-22에 병합 대상으로 `5c278d8b51`을 주었고, 그 병합(`7a88d26259`)이 두 조상을 함께 가져왔다. 세 커밋 모두 통합 브랜치 HEAD의 조상이다(리더 확인). 의도한 흐름이다.
  - 리더 판독: `6c833ee4b4`의 `wip_limit` 읽기는 미설정을 nil로 남기고, "숫자 기본값은 열린 결정이며 여기서 만들지 않는다"는 주석과 "미설정을 무제한으로 다루지 않는다" 시험을 가진다. 그래서 OD-7을 미리 정하지 않는다(리더 판단).
  - 이 실행의 따로 읽은 관찰('measured by the SPEC writer on 2026-10-29', 이 트리에서 읽기 전용 git): `git merge-base --is-ancestor`로 `6c833ee4b4`·`8920c0ecf8` → `5c278d8b51` 둘 다 종료 0, `5c278d8b51` → `WT-agentic-factory-integration` 끝(`e8106e3093`) 종료 0. `7a88d26259`의 부모는 `0372f9d215`와 `5c278d8b51`이고 제목은 "merge(WT-agentic-factory-settings-keys): periodic integration of lane-17 reader @5c278d8b51". `git grep 6c833ee4b4`로 본 `internal/config/agentic_factory.go`: 53행 `PRQueueWIPLimit *int`, 60~64행 `WIPLimit()`은 nil이면 `0, false`를 돌려준다. 58~59행 주석 원문은 "unset value (0, false) must NOT be treated as unlimited by the caller; the numeric default is an open decision and is deliberately not invented here." 시험 `TestAgenticFactoryWIPLimitUnsetIsNotUnlimited`(`agentic_factory_test.go:257`)는 미설정·null에서 필드가 nil인지를 단언한다(278~279행). 호출자가 무제한으로 다루지 않는지는 그 시험이 단언하지 않는다. 이 실행은 OD-7을 미리 정했는지 판단하지 않는다.
- **0.6.7 확인 공백과 정본 차이 (관찰)**: 출처: 'leader decision, relayed via the leader session'. 권고나 코드 변경 요구가 아니다.
  - 운영자의 가장 새 정본은 Page 'seq57' "agent가 Issue 종결까지 책임 유지"이며 그 제품 시험은 NOT_RUN이라고 리더가 전달했다. 정본 seq57은 운영자가 리더 세션에 직접 전달한 것이며, Page 원문은 리더도 이 SPEC 작성자도 읽지 않았다(canon seq57: operator direct relay to the leader session; page original not read by the leader or by this SPEC author). 이 기록의 출처는 리더 결정(리더 세션 경유)이다. DEC-AF-23과 그에 따라 바꾼 REQ-AF-009·010·016 문장은 이 리더 결정에 근거한다. seq57 원문과의 대조는 하지 않았다.
  - 정본 차이와 리더 판정: design §2.2의 E03 도착 칸 "종결 책임 leader로"와 E06 "가용 agent에 배차"는 S46([7])에서 옮긴 문언이다. **리더 판정(리더 세션 경유)**: 정본 seq57이 S46의 E03·E06 문언을 대체한다. 원문 대조는 하지 않았다(original-text comparison not done). 그래서 0.6.7의 바꿔 쓴 문언 — 책임은 배정된 agent에 남고 수리는 그 Issue의 책임 agent에게 간다 — 을 그대로 둔다. **책임과 실행 자리(리더 판정, 리더 세션 경유)**: 둘은 따로 있는 두 자원이 아니며 효과로 읽는다. Holder가 책임을 계속 갖고, 새 Issue를 받을지는 acquire 검사가 판정한다. S46의 E03 "agent claim·slot 반환" 문언은 이 읽기로 해석하며, 별도 자원으로 나눈다는 뜻이 아니다(DEC-AF-23, REQ-AF-009, design §2.2 E03). design §2.2 두 행에는 S46 문언을 0.6.7 문언 옆에 그대로 남겨 두었다.
  - 대기 기록 여섯 필드(design §2.1c)의 값 — 재확인 시각, 재시도 상한, 상승 대상 — 은 정하지 않았다. 필드가 있어야 한다는 것만 정한다. 숫자는 OQ-22와 같은 원칙으로 열려 있다.
  - 시험 세 가지의 이름은 "제안"이며 lane-22가 정한다(acceptance §A).
  - **lane-18 계약 변경 관찰 (다섯 번째 후속 정정)**: 아래는 리더가 lane-18 코드에서 전달한 사실이며, **이 작성자는 그 코드를 다시 읽지 않았다**. 출처: 'leader decision, relayed via the leader session'. 구현된 동작의 관찰이며 요구사항이 아니다.
    - 인계 receipt(`2251e77393`, 브랜치 `WT-agentic-factory-core-claim-runner`): 새 필수 필드 `PR URL`(http/https), `base`, `adopted SPEC`, `task result`, `check status`. 유지 필드 `handoff id`, `LeaderAccepted`, `PR number`, `HEAD`, `branch`, `worktree`, `WriterStopped`, `BoundaryHead`. 옛 receipt도 읽을 수 있다(읽기 호환). REQ-AF-009·design §2.2 E03·§4.8·AC-AF-009 (a)에 이 이름을 붙였다.
    - **필드 대응(리더 판정, 일곱 번째 후속 정정 A1)**: "계약" = `adopted SPEC`(AdoptedSpec). "보존한 변경과 quiescence 증거" = `WriterStopped` + `BoundaryHead`==`HEAD` + `branch` + `worktree`(예약된 실제 경로)이며 이것으로 충족된다. "원 writer의 출처"는 receipt가 아니라 claim 획득 때의 Issue 작성자 검사(`ErrNotIssueAuthor`)가 충족하고, receipt의 책임 주체 출처는 `Session`/`Agent`다. receipt 필드는 새로 더하지 않는다. REQ-AF-009와 design §4.8은 이 이름으로 고쳤다(새 의무 없음). 출처: 'leader decision, relayed via the leader session'. 코드는 이 작성자가 다시 읽지 않았다.
    - **옛 receipt(리더 판정, A2)**: 이미 기록된 과거 상태이며 다시 검증하지 않는다(읽기 호환). REQ-AF-009의 "모든 필드가 있어야 함"은 새 인계(E03 쓰기)에만 적용되며, 그 범위를 문장에 적었다. 출처: 'leader decision, relayed via the leader session'.
    - **수리 세션과 agent(리더 판정, A3 — 충돌 아님)**: 리더가 `af-core-claim-runner`의 `claim/next.go` 262~275행을 읽었다: `TakeOver`는 epoch+1과 함께 receipt의 `Session`을 새 holder로 옮긴다("이제 PR 책임자는 새 holder"). 그래서 "원래 책임 세션에만"(`67defdeb1f`)은 인계 뒤에는 새 책임 세션을 뜻하며, REQ-AF-004·AC-AF-004 (e)와 맞는다. 충돌 목록에서 뺐다. **열린 관찰**: PARKED Issue가 인계될 때 Wait 기록을 어떻게 다룰지는 lane-18이 고치는 중이다. 출처: 'leader decision, relayed via the leader session'. 이 작성자는 그 코드를 다시 읽지 않았다.
  - **설정 읽기 구현 관찰(여덟 번째 후속 정정, 요구사항 아님)**: 커밋 `dc1bff9643`(브랜치 `WT-agentic-factory-settings-keys`)의 `internal/config/agentic_factory_settings.go`가 SPEC이 정하지 않은 곳에서 고른 동작이다. 출처: 'leader decision, relayed via the leader session'. 이 작성자는 그 코드를 다시 읽지 않았다.
    - (a) **해결(리더 결정, D1)**: `max_id_bytes`는 언제나 필수다(DEC-AF-25 D5가 우선). 읽기 코드는 `seen_limit`이 `0`이어도 이 키를 요구한다. design §6.1의 "`SeenLimit>0`이면 필수" 문언은 고쳤다.
    - (b) YAML 키 중복은 오류다 — 리더가 수용함(accepted by the leader).
    - (c) 경로 중간이 map이 아니면(예: `workflow: 5`) 첫 단계에서도 오류다. 빈 파일이나 `workflow` 블록이 없으면 "설정되지 않음"이다 — 리더가 수용함.
    - (d) 모든 필수 키에서 `null`은 빠진 것으로 센다. `enabled: null`만 "설정되지 않음"이다 — 리더가 수용함.
    - (e) duration은 YAML 문자열이어야 한다(0은 `0s`로 쓴다) — 리더가 수용함.
    - (f) 비밀 보호: `webhook.routes[]` 아래 이름이 정확히 `secret`인 키는 거절하며, 값은 출력하지 않는다 — 리더가 수용함.
    - (g) 첫 오류만 보고한다(실행마다 키 하나) — 리더가 수용함.
    - (h) `collector.poll` min/max/safety는 존재와 해석만 검사한다 — 리더가 수용함. OD-17의 "필요한 규칙"은 여전히 채택되지 않았다.
    - 처음 적었던 (i)(큐의 `max_doc_bytes`가 `0`이면 부등식 검사를 건너뜀)은 뺐다: 리더 결정 D3로 `max_doc_bytes`로 끝나는 키는 모두 >= 1이고 `0`은 거절되므로, 부등식은 언제나 검사된다.
    - 수리는 새 attempt이며 이름은 `repair-<n>`이다(`06f8ccc3c7`). REQ-AF-009의 "new attempt"와 어긋나지 않는다. 이름 형식은 이 SPEC이 정하지 않는다.
    - 새 `PARKED` 상태로 들어가는 쓰기는 Wait 기록을 요구한다(`deb7551aac`). REQ-AF-016의 대기 기록 규칙과 같은 방향이다. 그 Wait 기록이 design §2.1c 여섯 필드를 모두 요구하는지는 전달되지 않았다.
    - worktree 경로는 실제 경로로 정규화하고, 존재하지 않는 경로는 거절한다(`150d49be8f`). 이 SPEC의 REQ 문장은 worktree 경로 처리를 적지 않으며, AC-AF-001의 전용 시험 `TestClaim_WorktreeIdentity_NormalizedPathPreventsMismatch`가 정규화 방향과 같다. "존재하지 않는 경로 거절"을 단언하는 AC 문장은 없다.
- lane-18 트리에 git을 실행하지 않았으므로 그 트리의 커밋되지 않은 변경 여부는 모른다. 대조한 줄 번호는 체크아웃된 파일과 `f9f44e5ecc`의 파일에서 같았다.
- 자기 Issue의 레이블·담당자·제목·마일스톤·다시 열기 동작에는 정본 근거가 없다(DEC-AF-19). 결정이 필요해지면 리더에게 보고한다.
- 설정 키 → `core.Collaboration` 연결은 아직 없다(lane-18 보고). 연결 담당은 lane-18, 읽기는 lane-17.
- 저장소 이름 정규화: GitHub 경로는 ASCII 소문자로 바꿔 비교한다(GitHub 문서 "The name is not case sensitive.", WebFetch 요약·uncollated). GitLab과 알 수 없는 공급자는 정확히 일치로 비교한다. GitLab의 대소문자 규칙은 2026-10-29에 공식 문장을 찾지 못했고, 실제 인스턴스 관측은 하지 않는다.
- 잔여 위험: GitHub은 저장소 이름 변경·이전에 301로 답한다(uncollated). 완화: 어떤 redirect도 허용으로 보지 않고 `repository-not-allowed`로 막는다(AC-AF-003 IA9·IA10).
- 자동 종결 키워드 목록의 문서 확인은 WebFetch 요약이며 원문 대조가 아니다(§B.2).
- `control/schema.json`의 sha256은 리더 보고이며 이 실행이 다시 재지 않았다.
- 레인 보고 사실(§B.2)은 하나도 이 실행이 검증하지 않았다.
- 효율 측정의 토큰 receipt 경로, 구 Factory baseline 원천은 측정하지 않았다(정본 S52 → NOT_RUN 또는 비교 불가).
- GitLab 조건부 요청(ETag)·429 헤더 누락 규칙·403 사용 범위는 공식 문서에서 확인하지 못했다(research R10).
- 이 실행 중 worktree 가드가 한 줄 반복문(SPEC ID 두 개를 한 번에 검사)을 "너무 복잡"하다고 거절했다. 각 ID를 따로 검사해 둘 다 `PASS`를 얻었다.
- 같은 가드가 `git show … | awk …` 형태의 명령(0.5.0 REQ 의미를 뽑으려던 것)을 "명령을 실행할 수 있는 awk 프로그램"이라며 거절했다. awk 없이 `git show 3275706cc9:…spec.md | grep -o -E`로 대신 읽었다(research R12).
- lint 판정 빌드(`moai-tree5`)의 커밋 표기는 빌드 때 손으로 넣은 `8dc835756`이고 트리 HEAD는 `3275706cc9`다. 표기가 판정한 트리를 가리키지 않는다(progress §E.1).
- 0.5.0 §H의 나머지 공백(PR 변경 파일 목록 미열람, `moai codex -l` 구현 미열람, Aside 공식 문서 미열람, #1926 둘째 댓글 일부만 읽음, `auto-merge.yml` 일부만 읽음)은 그대로다.

## §I 범위 밖

### Out of Scope — 별도 SPEC 분할

- 이 프로그램의 SPEC은 둘뿐이다(DEC-AF-05). 나머지는 이 SPEC의 절이다.
- 하위 SPEC-AGENTIC-FACTORY-IDENTITY-001은 넓히지 않는다. D2 완화 구현, 기준점 여러 개(D8), 응답자 인증은 범위 밖이다.

### Out of Scope — 남은 카드의 Issue 이관 (Phase 2)

- 남은 카드·열린 PR·이슈의 대사, 중복 확인, 재우선순위, 기존 owner 인계는 이 SPEC 뒤의 단계다.
- 이관 자체는 구현 완료로 세지 않는다.
- 이관 문맥에서 기존 Issue를 대량 발행·종료하지 않는다. 접수 요청의 Issue 발행(REQ-AF-012)은 이 금지에 들지 않는다.

### Out of Scope — 기존 설정·코드 수정

- `.moai/config/**`, `internal/config`, `internal/settings`, 템플릿 트리는 이 SPEC이 고치지 않는다. 설정 읽기는 lane-17의 별도 작업이다.
- `FactoryProcessAlive`와 그 호출처, SPEC-ASIDE-BROWSER-001은 고치지 않는다.

### Out of Scope — 남의 Issue 변경 승인

- 협업 옵션 기록(DEC-AF-18)은 지금 남의 Issue를 바꾸라는 승인이 아니다. 새 writer, 권한 확대, 추가 감사를 승인하지 않는다.

### Out of Scope — 기존 결함의 종료와 기존 작업의 처분

- #1938 / #1926 / #1924 / #1847 / #1813을 자동으로 닫지 않는다. 기존 PR을 닫거나 브랜치를 덮어쓰거나 카드를 일괄 done 처리하지 않는다.
- 강제 프로세스 종료, 작업 트리 삭제, 미커밋 변경 폐기는 하지 않는다.

### Out of Scope — provider template 작성과 그림

- Issue·PR template의 작성은 다른 worktree의 별도 작업이다. 이 SPEC은 참조만 한다(design §16).
- manifest에 없는 기존 template 파일에 새 template 개정을 강제로 덮어쓰는 일은 하지 않는다(DEC-AF-24, 알려진 동작).
- Pele SVG는 생성하지 않았다(기록된 공백).

### Out of Scope — 설정 화면(웹 UI) 표시

- 브라우저 3선택을 사람이 보는 화면에 그리는 일은 다루지 않는다(OQ-18).

### Out of Scope — 공급자·비용

- GLM managed 경로, Codex 자동 활성화, 새 키·유료 용량·추가 지출, Cloud 개발 재개, 앱 설치, webhook 수신 서버 운영은 별도 승인 대상이며 이 SPEC이 승인하지 않는다.
