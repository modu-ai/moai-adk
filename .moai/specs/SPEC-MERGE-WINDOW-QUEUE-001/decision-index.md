# decision-index — SPEC-MERGE-WINDOW-QUEUE-001

> `interview.decision_gate: on`(.moai/config/sections/interview.yaml)에 따라 작성했다. 각 행은 무엇이 왜 미결인지를 밝힐 뿐 권고를 담지 않는다. Q1은 운영자가 이미 답한 승인의 기록이다. Q2–Q6은 v0.1.0 작성 당시 비어 있었고, v0.2.0에서 리더 결정(미션 계약 07d28c4b)을 판정란에 기록했다. v0.3.0에서 plan-audit 1차(`.moai/reports/t1479/plan-audit-iter1.md`)를 닫는 리더 결정을 Q8–Q13으로 더했고, REQ/AC 번호를 001–025로 새로 매겼다 — 각 행의 「반영」은 새 번호다.
>
> v0.5.0(plan-audit 2차 이후 범위 축소, Q17)에서 번호가 다시 바뀌었다. Q8–Q16의 「반영」은 v0.3.0–v0.4.0 번호이고, v0.5.0 번호는 아래 표로 읽는다(REQ와 AC 공통).
>
> | v0.4.0 | 003 | 004 | 005 | 006 | 007 | 008 | 009 | 010 | 011 | 012 | 013 | 014 | 015 | 016 | 017 | 018 | 019 | 020 | 021 | 022 | 023 | 024 | 025 |
> |---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
> | v0.5.0 | 삭제 | 003 | 004 | 005 | 006 | 007 | 008 | 009 | 010 | 011 | 012 | 013 | 014 | 015 | 016 | 017 | 018 | 삭제 | 019 | 020 | 021 | 022 | 023 |
>
> 001·002는 번호가 그대로다. v0.5.0의 REQ-MWQ-017은 옛 018(창 안 단계)에 병합 동사(Q18)를 접어 넣은 것이다.
>
> 라벨 표기: 리더 결정은 `Operator verdict:` 줄에 `리더 결정(LEADER-DECIDED)`으로 남기고, `Label:`은 그대로 둔다. 라벨 어휘는 네 개(DECIDED · POLICY-COVERED · EVIDENCE-NEEDED · FOUNDER)로 고정이고, 커밋된 트리에 근거가 없는 결정은 DECIDED로 바꿔 달지 않는다는 규칙 때문이다. 판정은 기록됐고, 이 행들이 그 결정의 첫 커밋 원장이다.

### Q1: 병합 창의 배정을 리더의 지명에서 선착순 대기열로 바꾸는가?

Label: FOUNDER

Authority anchor: — (승인은 리더 세션 df44e022의 AskUserQuestion 답으로 내려졌고, 커밋된 트리에 그 기록이 없다. 그래서 이 행이 승인의 첫 커밋 기록이다.)

Why unresolved: 해당 없음 — 운영자가 결정했다. 현행 로컬 독트린(AGENTS.local.md §4.1 221행 「리더의 창 지명만이 근거다」)과 정면으로 다른 변경이라, 커밋된 권위 근거 없이 정책을 뒤집을 수는 없었다.

Operator verdict: 승인(2026-10-03, 리더 세션 df44e022, AskUserQuestion) — 「병합 창 선착순 대기열」: 리더 창 지명 폐지, 즉시 구현. 리더에게는 open/hold 정책과 push 전 최종 읽기만 남는다(push 쪽은 v0.3.0부터 SPEC-CANDIDATE-CI-001 소관 — Q8).

### Q2: 창 보유자 임대(lease)의 기본 기간은 얼마이고, 기본값으로 켜는가?

Label: EVIDENCE-NEEDED

Authority anchor: —

Why unresolved: 지금의 창에는 임대 개념이 없다(research.md §R1 E9). 재측정을 창 밖으로 옮기면 창 보유 시간이 크게 줄 것으로 보이지만, 새 순서에서 보유 시간이 어떻게 분포하는지는 아직 잰 적이 없다. 너무 짧으면 병합 중인 보유자를 밀어내고, 너무 길면 소유 프로세스는 살아 있는데 멈춘 레인이 대기열 전체를 막는다. REQ-MWQ-009는 값 0이면 꺼진다는 것만 정했다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b — 운영자: 즉시 구현·v3.2.0 포함). 임대는 기본으로 켜고 기간은 30분, 보유자가 부르는 창 동사마다 갱신한다. PID가 죽었거나 임대가 끝난 보유자는 stale이며 대기열 맨 앞이 승격된다. 근거: 재측정을 창 밖으로 옮긴 뒤의 창 안 경로(트리 항등 + no-ff 병합)는 짧으므로 30분이면 넉넉하고, 멈춘 레인이 대기열을 무기한 막지는 못한다. EVIDENCE-NEEDED였던 부분(새 순서에서 창 보유 시간)은 run M0에서 재며, 그 결과로는 기본값을 줄일 수만 있다. 반영(v0.3.0 번호): REQ-MWQ-007, -009, AC-MWQ-007, -009, plan M0.

### Q3: `--wait`를 한계값 없이 쓰면 대기 한계는 얼마인가(무한 포함)?

Label: EVIDENCE-NEEDED

Authority anchor: —

Why unresolved: 대기 한계를 정하려면 대기열 길이와 창 보유 시간의 실측이 필요한데, 둘 다 새 메커니즘이 생겨야 잴 수 있다. 무한 대기는 레인 턴을 붙잡고, 짧은 한계는 재진입을 되풀이하게 만든다(v0.2.0 기준: 재진입하면 대기열 맨 뒤로 갔다 — v0.3.0에서는 Q11의 슬라이스 재진입이 순번을 지킨다).

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b — 운영자: 즉시 구현·v3.2.0 포함). 한계 없는 `--wait`의 기본값은 60분이다. 시간이 다 되면 티켓을 거두고, 대기열 위치를 밝히며 0이 아닌 코드로 끝낸다. 근거: 레인 턴을 무기한 붙잡지 않으면서, 30분 임대를 둘 넘게 기다릴 수 있는 길이다. 반영(v0.3.0 번호): REQ-MWQ-002, -005, AC-MWQ-002, -005.

### Q4: 재측정 이후 develop이 움직여 재대기가 필요할 때, 그 레인은 맨 뒤로 가는가, 아니면 순번을 지키는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: (v0.2.0 기준) 창을 놓고 재흡수·재측정을 요구한다는 것까지만 정해져 있었다. 맨 뒤로 보내면 develop이 자주 움직이는 부하 구간에서 같은 레인이 거듭 밀려날 수 있다. 순번을 지키게 하면 재측정하는 동안 비어 있는 순번이 뒤 레인들을 막는다. 공정성과 처리량 가운데 무엇을 앞세울지는 정책 판단이고, 이 질문을 덮는 선행 SPEC이나 설정은 없다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b — 운영자: 즉시 구현·v3.2.0 포함). develop이 움직여 재대기하는 티켓은 순번을 지킨다. 재흡수·재측정을 마치면 한 번에 한해 맨 앞으로 다시 들어오고, 연달아 두 번째로 움직이면 맨 뒤로 간다. 근거: 기아를 막으면서, 비어 있는 순번이 뒤 레인들을 오래 막는 일도 한 번으로 묶는다. 반영(v0.3.0 번호): REQ-MWQ-019, -020, AC-MWQ-019, -020, plan M6 — 예약 티켓의 대기 방식은 Q10이 구체화했다.

Superseded (v0.5.0, Q17): 순번 유지·한 번 맨 앞 규칙은 폐지됐다. develop이 움직이면 레인은 다시 재고 대기열 맨 뒤로 다시 들어간다(REQ-MWQ-018).

### Q5: t1478(착지 전 후보 CI)이 먼저 착지하면, 재측정 근거로 후보 CI 실행과 로컬 재측정 가운데 무엇을 필수로 하는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: (v0.2.0 기준) 재측정 기록 요구사항이 둘 중 하나를 받았다. t1478은 아직 착지하지 않았다(research.md §R5). 후보 CI만 받으면 로컬 부하와는 무관해지지만 CI 대기가 생기고, 로컬 재측정만 받으면 darwin/windows 매트릭스를 보지 못한다. 두 카드 가운데 나중에 착지하는 쪽이 이 선택을 고정하게 된다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b — 운영자: 즉시 구현·v3.2.0 포함). t1478의 후보 CI 설정(candidate_ci)이 켜져 있으면 성공한 후보 CI run id를 재측정 기록으로 반드시 요구하고, 꺼져 있으면 로컬 재측정 기록(명령·exit·테스트 수)을 요구한다. 두 형태 모두 같은 검증기가 받는다. 설정이 없으면 꺼진 것으로 읽는다(t1478 미착지 상태). 근거: 후보 CI가 있으면 darwin/windows 매트릭스까지 보는 더 강한 근거를 쓰고, 없으면 로컬 측정으로 대신한다. 반영(v0.3.0 번호): REQ-MWQ-015, AC-MWQ-015. 키 경로는 SPEC-CANDIDATE-CI-001 REQ-CCI-023의 `workflow.candidate_ci.enabled`다.

### Q6: 재측정 기록의 「테스트 수」를 언어 중립적으로 어떻게 얻는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: `moai integration`은 사용자 프로젝트로 배포되는 동사라서, 16개 프로그래밍 언어의 러너 출력을 해석하는 방식은 템플릿 중립성 규율(AGENTS.local.md §15)과 부딪친다. 호출자가 수를 적어 넣게 하면 명령과 exit 코드는 관측값이 되지만 수 자체는 다시 주장이 된다. 인식 가능한 러너만 해석하고 나머지는 CI 실행 id를 요구하는 방식도 있다. 셋 모두 관측 강도와 중립성 사이의 맞교환이다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b — 운영자: 즉시 구현·v3.2.0 포함). 테스트 수는 선택 항목이며 도구가 보고한 값만 쓴다. 기록에는 명령과 exit 코드를 늘 담고, 테스트 수는 도구가 구조화된 보고를 낼 때만 요구한다(이 저장소에서는 `go test -json`의 수). 다른 언어는 exit와 명령만 기록한다. 보고된 수가 0이면 거부한다. 근거: 호출자가 적어 넣는 주장을 없애고, 배포되는 동사의 프로그래밍 언어 중립성(§15)도 지킨다. 반영(v0.3.0 번호): REQ-MWQ-016, -023, AC-MWQ-016, -023 — 빈 실행 거부는 Q12가 더했다.

### Q7: 마지막 origin/develop CI가 아직 진행 중이거나 판정이 없을 때 push를 보류하는가?

Label: POLICY-COVERED

Authority anchor: `.claude/rules/local/gitflow-lane-protocol.md` §4 「초록 조건부」 단락 — "마지막 push의 CI 판정이 아직 없으면(새 develop의 첫 push 등) red가 아니므로 보류 사유가 아니다."

Why unresolved: 해당 없음 — 커밋된 로컬 규율이 문면 그대로 이 질문을 덮는다. 다만 CI 상태를 아예 읽지 못하는 경우는 「판정 없음」이 아니라 「측정 못 함」이라서 이 정책이 덮지 않는다. v0.3.0에서 push 동사가 SPEC-CANDIDATE-CI-001로 옮겨 갔으므로(Q8) 이 행의 적용 대상도 그쪽 REQ-CCI-013이다. 이 SPEC에는 더 이상 해당 요구사항이 없다.

Operator verdict:

### Q8: push 동사를 이 SPEC에 두는가, t1478 쪽으로 옮기는가? (감사 D1·D2·D8)

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. 감사는 REQ 32개·AC 35개가 Tier L 상한 25/25를 넘었다고 지적했고(D1), push 설계의 세 결함도 찾았다(D8: 빨간 원격 팁 수리 push 차단, `--limit 1` CI 판독, 브랜치 이름 push).

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). push 동사(옛 REQ-040..044와 그 AC)를 이 SPEC에서 빼고 SPEC-CANDIDATE-CI-001(카드 t1478)로 옮긴다. 그쪽 push 게이트가 이미 `origin/develop..develop` 범위 전체를 검사한다. 이 SPEC에는 한 줄 포인터만 남기고(spec.md §E), D8 지적은 t1478에 넘기는 인계 기록으로 남긴다(research.md §R6). 문자 접미 항목(009a·021a·021b)은 부모에 접고 REQ/AC를 001–025로 연속 번호화한다. 반영: spec.md v0.3.0 전체 번호, §E, research.md §R6.

### Q9: hold 정책 중에 보유자가 release하면 대기열 맨 앞을 승격하는가? (감사 D3)

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. v0.2.0은 release·stale 때 승격한다는 요구와, hold 중에는 새 보유자를 주지 않는다는 요구가 같은 상황에서 부딪쳤다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). hold 중에는 어떤 승격도 일어나지 않는다. hold 중의 release는 대기열을 그대로 두고, 정책이 open으로 돌아오면 순서대로 승격을 다시 시작한다. 한 REQ에 담는다. 반영: REQ-MWQ-008, AC-MWQ-008, design.md D1(보유자 없는 대기열 허용 조건).

### Q10: develop 이동으로 재대기한 예약 티켓은 재측정하는 동안 대기열을 막는가? (감사 D4)

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. Q4의 「순번 유지·한 번 맨 앞」만으로는 예약 티켓이 재측정 중에 승격되는지, 막는지, 끝없이 남는지가 정해지지 않았다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). 예약 티켓은 대기열을 막지 않는다. 주인이 재측정하는 동안 「준비 안 됨」이고, 다음 준비된 티켓이 승격된다. 준비되면 다음 승격을 한 번 받는다(front-once). 한계(기본 30분) 안에 준비되지 않으면 사유와 함께 빠진다. 이것이 기아 방지 보장이다. 반영: REQ-MWQ-019, -020, AC-MWQ-019, -020. 이 규칙으로도 남는 기아 가능성은 research.md §R7에 잔여 위험으로 적었다.

Superseded (v0.5.0, Q17): 예약 티켓·준비 한계·front-once는 폐지됐다.

### Q11: 티켓의 생존은 무엇으로 판정하고, 시간 초과와 승격이 겹치면 어느 쪽이 이기는가? (감사 D5)

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. 세션 소유 pid로 판정하면, 죽은 대기 프로세스의 티켓이 살아 있는 것으로 읽혀 승격된 뒤 창을 막는다. Bash 도구 호출 한계(10분)가 기본 대기 60분보다 짧다는 점도 함께 드러났다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). 티켓 생존은 세션이 아니라 대기 중인 프로세스(pid + 프로세스 시작 시각)에 묶는다. 대기 프로세스가 heartbeat을 갱신하고, 죽은 대기자의 티켓은 다음 대기열 변경 때 빠진다. 시간 초과와 승격은 둘 다 잠금 기록 변경 안에서 정하며, 먼저 기록된 쪽이 이긴다. 이미 시간이 초과된 채 승격된 대기자는 즉시 창을 놓는다. 레인은 `acquire --wait`를 백그라운드로 돌리거나 짧은 한계로 반복 호출하며, 짧은 한계로 재진입해도 순번을 잃지 않는다. 반영: REQ-MWQ-001..006, AC-MWQ-001..006, plan §D.

Partly superseded (v0.5.0, Q17): 대기 프로세스 생존과 heartbeat 판정, 시간 초과·승격 경합 규칙은 유지한다. 보유자 생존은 다시 세션 소유 pid로 판정한다(티켓에 owner pid를 싣고 승격 때 보유자 기록에 찍는다 — REQ-MWQ-001/006). `--slice` 재진입은 폐지됐다. 레인은 `acquire --wait`를 백그라운드로 돌린다.

### Q12: 재측정 명령이 실제로 아무 테스트도 돌리지 않았을 때 완료 게이트는 받아 주는가? (감사 D6)

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. v0.2.0 규칙대로라면 `true`나 `[no tests to run]`인 실행도 유효한 기록이 됐다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). 완료 게이트는 러너가 스스로 빈 실행이라고 보고한 경우(`no tests to run`, `[no test files]`)와, 도구가 구조화된 출력을 지원하는데 그것 없이 돌린 명령(이 저장소에서는 `-json` 없는 `go test`)을 거부한다. 반영: REQ-MWQ-016, AC-MWQ-016. 인식된 보고를 내지 않는 도구(예: `true`)는 Q6대로 exit 0이면 유효로 남는다 — 이 경계는 리더 결정의 문면을 따른 것이다.

### Q13: 작업 트리가 더러운 상태에서 잰 재측정을 HEAD 트리 키로 받아 주는가? (감사 D7)

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. 커밋되지 않은 편집이 있으면 측정한 트리와 키로 쓴 트리가 달라질 수 있었다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). 재측정은 실행 전후 모두 깨끗한 작업 트리를 요구한다(`git status --porcelain`가 비어 있어야 한다). 아니면 거부한다. 반영: REQ-MWQ-017(실행 중 HEAD 불변 확인을 함께 담았다), AC-MWQ-017.

### Q14: heartbeat 간격, heartbeat 유효 창, 슬라이스 재진입 유예는 각각 얼마인가?

Label: EVIDENCE-NEEDED

Authority anchor: —

Why unresolved: Q11은 생존을 heartbeat과 슬라이스 재진입으로 판정한다고 정했지만 값은 정하지 않았다. 유효 창이 짧으면 부하 120–255에서 살아 있는 대기자가 heartbeat를 늦게 갱신해 빠질 수 있고, 길면 죽은 대기자의 티켓이 그만큼 오래 남는다. 재진입 유예가 짧으면 레인 턴 사이 간격에서 순번을 잃고, 길면 돌아오지 않는 레인이 맨 앞 자리를 붙잡는다. 부하 상태의 heartbeat 지연과 레인 재호출 간격 모두 실측이 없다. (v0.3.0 기준) spec.md §F가 이 값들 없이는 M1에 들어가지 않는다고 적었다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). heartbeat 간격 15초, heartbeat 유효 창 60초(박동 4회를 놓치면 죽은 대기자), 슬라이스 재진입 유예 120초(`--slice` 대기자가 120초 안에 다시 들어오면 순번 유지). M0 측정은 이 값들을 조이기만 할 수 있다. 반영: REQ-MWQ-003, -004, AC-MWQ-003, -004(경계값 시나리오 ±1초), spec.md §F.

Partly superseded (v0.5.0, Q17): heartbeat 15초·유효 창 60초는 유지한다. 슬라이스 재진입 유예 120초는 `--slice`와 함께 폐지됐다.

### Q15: 다른 보유자의 병합 때문에 develop이 움직인 경우도 「두 번째 이동 → 맨 뒤」로 세는가? (research.md §R7)

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. v0.3.0 규칙대로라면 준비를 마친 예약 티켓이 다른 레인의 창을 기다리는 사이 그 레인의 병합으로 develop이 움직여도 두 번째 이동으로 세어져, 부하 구간에서 같은 레인이 계속 맨 뒤로 밀릴 수 있었다(§R7).

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). 「두 번째 이동 → 맨 뒤」에는 티켓 자신의 재측정 중에 develop이 움직인 경우만 센다. 준비된 티켓이 다른 보유자의 창을 기다리는 동안 생긴 이동은 세지 않고, 그 티켓은 맨 앞 자리를 지킨다. 한계: 종류를 가리지 않고 연속 3회 재대기한 티켓은 맨 뒤로 가고, 그 사건을 기록한다. 반영: REQ-MWQ-020, AC-MWQ-020 시나리오 4–6, research.md §R7.

Superseded (v0.5.0, Q17): 재대기 계수와 연속 3회 한계는 폐지됐다.

### Q16: 인식된 테스트 구조를 보고하지 않는 명령을 exit 0으로 받아 주는 경계(Q12)를 그대로 두는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. Q12 반영 후에도 `true` 같은 비테스트 명령은 검증기가 가려내지 못한다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). 그대로 받아들인다. 인식된 테스트 구조를 보고하지 않는 명령은 exit 0이면 유효로 남는다. 검증기가 비테스트 명령을 탐지하지 못한다는 점과, 후보 CI 경로(t1478)가 더 강한 기록이라는 점을 잔여 위험에 적는다. 반영: spec.md §D 잔여 위험 항목, AC-MWQ-016(`true` 행은 유효).

### Q17: plan-audit 2차가 퇴행(0.75 → 0.69)했을 때, 대기열 공정성 장치를 고쳐 세 번째로 반복하는가, 범위를 줄이는가? (감사 N1–N4)

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. 감사는 같은 서브시스템(티켓 상태 모델)에서 모순 네 건을 찾았다. 보유자의 owner pid 소실(N1), 예약 티켓이 생존 판정에 걸려 빠지는 문제(N2), 슬라이스 사이 상태의 생존·승격 모순(N3), 도달 불가능한 재대기 계수(N4)다. 그리고 STOP 신호와 함께 범위 축소·부채 수용·명시적 재반복을 선택지로 올렸다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b) — 범위 축소(일반 반복이 아님). 근거: 재측정이 창 밖으로 나가 창은 몇 초(트리 항등 + no-ff 병합)만 쥐어지므로, 정교한 공정성 장치는 얻는 것이 없다. 예약 티켓, `--slice`/슬라이스 사이 상태, front-once, 재대기 계수, 연속 3회 규칙을 없앤다(Q4·Q10·Q15 전체, Q11·Q14의 해당 부분을 대체). 대기열은 `acquire --wait[=한계]`(기본 60분)가 만드는 단순 FIFO다. 티켓 생존은 오늘의 잠금처럼 세션 소유자로 판정하고(N1: `Stale`·`releasableBy`가 기대는 owner pid 의미 복원), 대기 프로세스는 자기 티켓을 지키려고만 heartbeat하며, 죽은 대기자의 티켓은 다음 변경 때 빠진다. 승격은 잠금 변경 안에서 일어나고, 시간 초과와 승격은 먼저 기록된 쪽이 이기며, 시간이 초과된 채 승격된 대기자는 즉시 놓는다. 창 안에서는 재측정 기록과 트리 항등만 본다. 재측정 뒤 develop이 움직였으면 `integration merge`가 창을 놓고 「재측정 후 재획득」 코드로 끝나고, 레인은 다시 재고 맨 뒤로 재획득한다. 유지: hold 정책(hold 중 승격 없음), 실질적 완료 게이트, 전후 깨끗한 트리, 커밋된 no-wait 기준선, 레인 병합 동사(Q18). 반영: spec.md v0.5.0 REQ-MWQ-001..011, -017, -018, §E 「Queue fairness machinery」, research.md §R7.

### Q18: 레인의 develop 병합을 동사로 노출하고, t1478의 공용 landing check를 부르게 하는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. 지금은 레인이 통합 워크트리에서 손으로 `git merge --no-ff`를 친다. 그래서 t1478의 공용 landing check(REQ-CCI-011)를 거치게 할 수도, 독트린이 손 병합을 금지할 수도 없다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). 이 SPEC이 레인 병합 동사를 소유한다. 창 안 단계(트리 항등 확인 + develop으로 `--no-ff` 병합 + release)를 `moai integration merge --card <id>`로 노출한다. 카드의 WT 브랜치는 SPEC-CANDIDATE-CI-001 REQ-CCI-004와 같은 방식으로 찾고, SPEC-CANDIDATE-CI-001의 공용 landing check(REQ-CCI-011)를 부른다(`workflow.candidate_ci.enabled`가 false면 아무 일도 하지 않음). 이 동사가 착지하면 독트린이 레인의 손 `git merge`를 금지할 수 있다. 25/25 안에서 기존 창 안 REQ에 접는다. 반영: REQ-MWQ-017(옛 018에 접음), REQ-MWQ-013(손 병합 금지 문장), AC-MWQ-013, -017, plan M5.

### Q19: `factory complete`와 병합 동사는 병합 경로를 하나로 쓰는가, 창 안 실패는 어떻게 끝나는가, 승격된 보유자는 통합 대상을 어디서 얻는가? (감사 B1–B3)

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. plan-audit 3차(0.75, 첫 상한 도달)가 막힘 세 건을 찾았다. 하나는 complete와 병합 동사가 서로 다른 병합 경로를 가진 문제다. 동사로 병합한 카드는 merged-local에 닿지 못하고, 그 뒤 complete를 부르면 재측정 루프에 빠지며, complete는 병합한 뒤에야 게이트를 돌린다(B1). 또 창 안 실패의 결말이 정해지지 않았고 SHA가 고정되지 않았다(B2). 마지막으로 승격 때 통합 대상(`branch`·`branch_source`·`worktree`)이 사라진다(B3).

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b — 감사자의 fix_scope 안에서 한 번의 델타 라운드, 범위 변경 없음). B1: 병합 경로는 하나다. `moai factory complete`는 `moai integration merge --card <id>`와 같은 창 안 단계(base 확인·트리 항등·landing check·고정 SHA 병합·release)를 불러 병합하고, 그 뒤 카드를 merged-local로 기록한다. `factory_card.go:1409`의 자체 병합은 그 호출로 바뀌어 게이트가 언제나 병합보다 먼저 돈다. 같은 카드를 `integration merge`로 이미 병합한 뒤의 complete는 그 카드 브랜치의 병합 커밋이 develop에서 닿는지 확인하고 상태만 기록한다(재측정 루프 없음). 독트린(REQ-013): 레인은 `integration merge` 또는 그것을 부르는 `factory complete`로만 병합하며, AGENTS.local.md:219와 gitflow-lane-protocol.md:99를 같은 뜻으로 고친다. B2: 창 안 실패는 모두 창을 놓고 원인마다 다른 코드로 끝난다(트리 불일치, landing check 거부, 병합 충돌 — `git merge --abort` 후 통합 트리가 깨끗한지 확인하고 놓는다 — 그 밖의 오류). 보유자가 아닌 호출은 잠금을 건드리지 않고 거부한다. 병합은 항등 확인 때 고정한 SHA로 한다(`git merge --no-ff <sha>`, 브랜치 이름은 쓰지 않는다). B3: 티켓은 acquire 때 branch·branch_source·worktree를 기록하고, 승격이 그것을 보유자 기록에 복사한다. 그래서 승격된 보유자에게도 factory complete의 `lock.Branch` 소유 확인이 성립한다(AC 픽스처 추가). O1–O4는 한 줄로 끝나는 곳만 고친다. 반영: REQ-MWQ-001, -003, -006, -009, -013, -017, -018, -019, AC-MWQ-001, -003, -006, -009, -013, -017, -018, -019, design.md D1·D3, plan.md M5·M6, research.md §R5(O3·O4).

작성자 메모(manager-spec): 리더 지시는 「abort 후 깨끗한지 확인하고 놓는다」까지였고, 확인이 실패하는 경우는 정하지 않았다. 감사자 B2가 「더러운 트리 위로 조용히 승격하지 않는다」는 선택지로 `hold`를 제시했고, REQ-MWQ-018은 그 경우 정책을 `hold`로 두고 놓는다고 적었다. 리더가 다른 결말을 원하면 이 한 절만 바꾸면 된다.

### Q20: plan-audit 4차(0.75, claude·codex 일치) 뒤에 한 번 더 고치는가?

Label: FOUNDER

Authority anchor: — (결정은 2026-10-03 운영자의 AskUserQuestion 답으로 내려졌고, 커밋된 트리에 그 기록이 없다. 이 행이 첫 커밋 기록이다.)

Why unresolved: 해당 없음 — 운영자가 결정했다. 감사는 막힘 네 건을 찾았다(`.moai/reports/t1479/plan-audit-iter4.md`).
- C1: 3차 수리 뒤 complete의 카드 게이트가 develop이 움직인 다음에 돈다.
- C2: 채택(adoption) 조건이 브랜치의 현재 끝과 유효 기록에 묶이지 않았고, 절의 우선순위가 정해지지 않았다.
- C3: 병합 커밋이 생긴 뒤의 실패가 정의되지 않았다.
- C4: 보유자가 아닌 호출의 거부와 「병합도 대기열 변경」이 서로 모순된다.

Operator verdict: 운영자 결정(2026-10-03, AskUserQuestion) — 좁은 범위로 델타 한 라운드를 더 돈다. 새 REQ는 만들지 않는다. 리더 결정(LEADER-DECIDED, 미션 계약 07d28c4b)으로 받은 수정 지시는 다음과 같다.
- C1: complete의 모든 카드 게이트(T14 리스 보유자·버전 확인)는 develop이 움직이기 전에 돈다. REQ-019를 고치고 design·plan을 맞추며, 만료된 리스와 남의 리스 AC를 둔다.
- C2: 채택하려면 병합 커밋의 두 번째 부모가 카드 브랜치의 **현재** 끝이어야 하고(이후 커밋이 있으면 채택하지 않고 재병합 경로로 간다), 기록이 REQ-014/015로 유효해야 한다. 거부 절은 채택보다 먼저 평가하고, 순서는 design에서 REQ-019 본문으로 옮긴다. 「동사 병합 → 추가 커밋 → 채택 안 됨」 픽스처 AC를 둔다.
- C3: 병합 커밋이 생긴 뒤의 실패는 결말을 정한다. 병합 커밋을 그대로 두고, 원인을 담아 `hold`를 걸고, 창을 놓고, 고유 코드로 끝난다. AC-018은 REQ가 보장하지 않는 「병합 커밋이 남지 않는다」를 주장하지 않는다.
- C4: 보유자 판정을 위한 읽기는 허용하고, 쓰기·대기열 변경은 허용하지 않는다. REQ-009·D1과 REQ-017의 순서를 하나로 맞춘다.
- 선택: 병합 단계의 `hold` 쓰기를 REQ-012 레인 거부의 예외로 둔다. REQ-018과 AC-018의 종료 코드 수를 맞춘다. AC-018 병합 실패 행의 설정을 분명히 한다.

반영(v0.7.0): REQ-MWQ-009, -017(보유자 확인 먼저·유효성·조상 관계 전제), -018(원인 9가지·병합 뒤 실패는 커밋 유지 + `hold`·시스템 쓰기), -019(게이트 순서 1–4·채택 조건), AC-MWQ-017 시나리오 4–5, AC-MWQ-018(9행 표·병합 커밋 유무 열·병합 seam 주입), AC-MWQ-019 시나리오 5–7, design.md D1·D3, plan.md M5·M6. REQ·AC 개수는 23/23 그대로다.

### Q21: 고정 SHA·기록 base·develop 끝이 모두 같아 병합할 것이 없을 때(plan-audit 5차 codex P2) 어떻게 끝나는가?

Label: FOUNDER

Authority anchor: — (운영자 결정, 2026-10-03. 커밋된 근거가 없어 이 행이 첫 커밋 기록이다.)

Why unresolved: 해당 없음 — 운영자가 결정했다. 5차 감사(`.moai/reports/t1479/plan-audit-iter5.md`)는 필수 codex 게이트에서만 FAIL했고, 지적은 P2 한 건이다. 세 값이 모두 같으면 조상 관계·트리 확인은 통과하지만, `git merge --no-ff`가 "Already up to date"로 끝나 병합 커밋이 생기지 않는데, 이때의 결말이 정해져 있지 않았다.

Operator verdict: 운영자 결정(2026-10-03) — 한 줄 수정 후 해당 부분만 다시 읽는다. 고정 SHA가 기록의 base와 같으면(병합할 것이 없으면) 병합 단계는 `git merge`를 부르기 전에 거부하고, 창을 놓고, 고유 코드로 끝난다. REQ-019의 병합 뒤 절은 「병합 커밋이 생긴 뒤의 모든 상태 전이 실패」로 넓힌다. 다른 수정은 하지 않는다. 반영: REQ-MWQ-017(같은 값 거부), REQ-MWQ-018(원인 10, 코드 열 개), REQ-MWQ-019(병합 뒤 절 확장), AC-MWQ-018 10행, AC-MWQ-019 시나리오 8의 코드 수, design.md D3의 코드 수.

### Q22: 부분 재독(`.moai/reports/t1479/plan-audit-reread.md`)에서 필수 codex 게이트가 새로 찾은 B1·B2를 어떻게 닫고, plan 반복은 어디서 멈추는가?

Label: FOUNDER

Authority anchor: — (리더 결정, 미션 계약 11c79e1a. 커밋된 근거가 없어 이 행이 첫 커밋 기록이다.)

Why unresolved: 해당 없음 — 리더가 결정했다. 원인 10 수정은 확인됐다. 다만 필수 codex 게이트가 원래부터 있던 공백 두 건을 새로 찾았다.
- B1: 병합 동사가 카드 게이트를 거치지 않고, 요청한 카드와 창 기록의 카드가 같은지도 확인하지 않는다.
- B2: 병합 전후로 통합 워크트리가 깨끗한지 확인하지 않는다(autostash 잔여물 포함).

Operator verdict: 리더 결정(LEADER-DECIDED, 미션 계약 11c79e1a).
- B1: `moai integration merge --card`도 complete와 같은 카드 게이트(merge-ready·만료되지 않은 카드 리스 보유·버전)를 적용하고, 요청한 카드가 창 기록의 카드와 같아야 한다. 모두 develop이 움직이기 전에 확인하며, 거부하면 카드와 develop을 그대로 두고 고유 코드로 끝난다.
- B2: 통합 워크트리가 깨끗한지 병합 전후에 확인한다. 병합 전에 더러우면 고유 코드로 거부한다. 병합 뒤에 더러우면 원인 8처럼 처리한다(커밋 유지, 원인과 SHA를 담은 `hold`, 창 해제, 고유 코드).
- **수렴 규칙**: 이번이 마지막 plan 라운드다. 이후에 나오는 치명적이지 않은 지적은 run 단계로 넘기는 기록된 의무로 다루고, 판정은 PASS-WITH-DEBT로 한다.

반영: REQ-MWQ-017(카드 게이트·카드 일치·병합 전후 깨끗한 트리), REQ-MWQ-018(원인 11·12, 원인 8에 병합 뒤 더러움 포함, 코드 열두 개), REQ-MWQ-019(코드 수), AC-MWQ-018 행 11a–11c·12·8b, AC-MWQ-019 시나리오 8의 코드 수, design.md D3의 코드 수.

### Q23: 최종 재독(`.moai/reports/t1479/plan-audit-final.md`)에서 나온 치명 지적 한 건(데이터 유실)은 어떻게 닫는가?

Label: FOUNDER

Authority anchor: — (리더 결정, 미션 계약 11c79e1a. Q22 수렴 규칙 「치명 지적은 반드시 고친다」를 적용했다.)

Why unresolved: 해당 없음 — 리더가 결정했다. 고정 SHA가 새로 추가하는 경로가 통합 워크트리에 무시되었거나 추적되지 않은 파일로 이미 있으면, 병합하는 과정에서 그 파일의 내용을 잃을 수 있다.

Operator verdict: 리더 결정(LEADER-DECIDED, 미션 계약 11c79e1a). REQ-MWQ-017은 `git merge` 전에 고정 SHA가 통합 브랜치 끝보다 새로 추가하는 경로를 계산한다. 그중 하나라도 통합 워크트리에 무시되었거나 추적되지 않은 파일·디렉터리로 이미 있으면, 원인 13(고유 코드)으로 병합 전에 거부하고, 그 파일의 내용은 건드리지 않은 채 창을 놓는다. 원인 수는 열셋으로 맞춘다. 남은 치명적이지 않은 지적 O1–O4는 Q22 규칙에 따라 run 단계 의무로 progress.md §E.1에 기록한다. 반영: REQ-MWQ-017, REQ-MWQ-018(원인 13, 열세 개), REQ-MWQ-019(코드 수), AC-MWQ-018 13행, AC-MWQ-019 시나리오 8의 코드 수, design.md D3, progress.md.

### Q24: 델타 재독(`.moai/reports/t1479/plan-audit-delta-d468ff19c.md`, FAIL 0.80)에서 나온 D1(치명)·D2(주요)는 어떻게 닫는가?

Label: FOUNDER

Authority anchor: — (리더 결정, 카드 t1479. Q22 수렴 규칙과 Q23의 연장이며 커밋된 근거가 없어 이 행이 첫 커밋 기록이다.)

Why unresolved: 해당 없음 — 리더가 결정했다. 원인 13은 고정 SHA가 추가하는 경로 자체가 이미 있을 때만 거부해서, `runtime.local/payload`를 추가하는 후보가 무시된 파일 `runtime.local`을 디렉터리로 바꾸는 경우(감사자가 재현: 병합 exit 0, 파일 소실)는 통과했다. plan.md M5는 여전히 「원인 9가지」였고 병합 전 검사 순서에 카드 게이트·깨끗한 트리·병합할 것 없음·원인 13이 빠져 있었다.

Operator verdict: 리더 결정(LEADER-DECIDED, 카드 t1479). **이번이 리더의 마지막 수리 라운드다.** 새 REQ도, 범위 확대도 없다.
- D1: REQ-MWQ-017의 충돌 검사와 REQ-MWQ-018 원인 13은 추가된 경로, 추가된 경로의 상위 경로, 추가된 경로 아래의 경로 중 어느 하나라도 통합 워크트리에 무시되었거나 추적되지 않은 파일·디렉터리로 이미 있으면 `git merge` 전에 거부하고, 충돌한 바이트는 그대로 두고, 창을 놓고, 원인 13과 같은 코드로 끝난다. 원인 수는 열셋 그대로다. 감사자의 스크래치 재현은 AC-MWQ-018 13b 행이며 run 단계가 가장 먼저 쓰는 RED 픽스처다.
- D2: plan.md M5를 원인 열셋과 spec.md REQ-MWQ-017/018의 병합 전 순서로 고친다.
- D3은 run 의무 O2에 합치고(`status.showUntrackedFiles=no` 회귀), D4(심볼릭 링크·대소문자 비구분 파일시스템)는 run 의무 O5로 기록한다. 둘 다 새 REQ를 만들지 않는다.

반영: REQ-MWQ-017, REQ-MWQ-018(원인 13), AC-MWQ-018 행 13a–13c와 RED 픽스처 문단, design.md D3, plan.md M5, progress.md §E.1(O2 확장·O5). REQ·AC 개수는 23/23, 원인 수는 열셋 그대로다.

### Q25: 델타 재독(`.moai/reports/t1479/plan-audit-delta-9d9d5fffa.md`, FAIL 0.88)에서 나온 D5(주요)·D6·D7(선택)은 어떻게 닫는가?

Label: FOUNDER

Authority anchor: — (운영자 결정, 카드 t1479. Q22 수렴 규칙과 Q23·Q24의 연장이며 커밋된 근거가 없어 이 행이 첫 커밋 기록이다.)

Why unresolved: 해당 없음 — 운영자가 결정했다. 원인 13의 「경로」가 디렉터리 항목을 세는지 정하지 않아, 추적되던 디렉터리 `runtime.local/`을 파일 `runtime.local`로 바꾸는 후보가 통합 워크트리의 무시된 파일 `runtime.local/secret`이 있는 채로 검사를 통과했다(감사자 재현: 병합 exit 0, 파일 소실, codex도 동의).

Operator verdict: 운영자 결정(OPERATOR-DECIDED, 카드 t1479). **이번이 카드 t1479의 마지막 수리 라운드다.** 범위는 D5와 D6·D7뿐이고, 새 REQ도 다른 REQ·AC의 변경도 없다.
- D5: REQ-MWQ-017은 「경로」를 잎 항목(파일·심볼릭 링크·서브모듈 항목, `git ls-tree -r`이 나열하는 것, 트리 항목 아님)으로 정의하고, 디렉터리가 잎으로 바뀌는 변경도 추가된 경로로 센다. 원인 13은 같은 정의를 쓰며 원인 수는 열셋, 종료 코드는 원인 13과 같다. AC-MWQ-018 행 13d를 13b 옆의 둘째 RED 픽스처로 더한다.
- D6: progress.md §E.1의 O2 확장을 REQ-MWQ-017의 순서(깨끗한 트리 확인인 원인 12가 충돌 검사보다 앞)에 맞춰 다시 쓴다. 추적되지 않은 충돌은 설계상 원인 12가 먼저 막는다.
- D7: 행 13b의 RED-now 칸을 다시 실행할 수 있는 명령 순서와 실측 출력으로 바꾼다.

반영: REQ-MWQ-017, REQ-MWQ-018(원인 13), AC-MWQ-018 행 13d와 RED 픽스처 문단, design.md D3, plan.md M5, progress.md §E.1(O2 확장). REQ·AC 개수는 23/23, 원인 수는 열셋 그대로다.
