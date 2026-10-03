# decision-index — SPEC-MERGE-WINDOW-QUEUE-001

> `interview.decision_gate: on`(.moai/config/sections/interview.yaml)에 따라 작성했다. 각 행은 무엇이 왜 미결인지를 밝힐 뿐 권고를 담지 않는다. Q1은 운영자가 이미 답한 승인의 기록이다. Q2–Q6은 v0.1.0 작성 당시 비어 있었고, v0.2.0에서 리더 결정(미션 계약 07d28c4b)을 판정란에 기록했다. v0.3.0에서 plan-audit 1차(`.moai/reports/t1479/plan-audit-iter1.md`)를 닫는 리더 결정을 Q8–Q13으로 더했고, REQ/AC 번호를 001–025로 새로 매겼다 — 각 행의 「반영」은 새 번호다.
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

### Q11: 티켓의 생존은 무엇으로 판정하고, 시간 초과와 승격이 겹치면 어느 쪽이 이기는가? (감사 D5)

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. 세션 소유 pid로 판정하면, 죽은 대기 프로세스의 티켓이 살아 있는 것으로 읽혀 승격된 뒤 창을 막는다. Bash 도구 호출 한계(10분)가 기본 대기 60분보다 짧다는 점도 함께 드러났다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). 티켓 생존은 세션이 아니라 대기 중인 프로세스(pid + 프로세스 시작 시각)에 묶는다. 대기 프로세스가 heartbeat을 갱신하고, 죽은 대기자의 티켓은 다음 대기열 변경 때 빠진다. 시간 초과와 승격은 둘 다 잠금 기록 변경 안에서 정하며, 먼저 기록된 쪽이 이긴다. 이미 시간이 초과된 채 승격된 대기자는 즉시 창을 놓는다. 레인은 `acquire --wait`를 백그라운드로 돌리거나 짧은 한계로 반복 호출하며, 짧은 한계로 재진입해도 순번을 잃지 않는다. 반영: REQ-MWQ-001..006, AC-MWQ-001..006, plan §D.

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

### Q15: 다른 보유자의 병합 때문에 develop이 움직인 경우도 「두 번째 이동 → 맨 뒤」로 세는가? (research.md §R7)

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. v0.3.0 규칙대로라면 준비를 마친 예약 티켓이 다른 레인의 창을 기다리는 사이 그 레인의 병합으로 develop이 움직여도 두 번째 이동으로 세어져, 부하 구간에서 같은 레인이 계속 맨 뒤로 밀릴 수 있었다(§R7).

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). 「두 번째 이동 → 맨 뒤」에는 티켓 자신의 재측정 중에 develop이 움직인 경우만 센다. 준비된 티켓이 다른 보유자의 창을 기다리는 동안 생긴 이동은 세지 않고, 그 티켓은 맨 앞 자리를 지킨다. 한계: 종류를 가리지 않고 연속 3회 재대기한 티켓은 맨 뒤로 가고, 그 사건을 기록한다. 반영: REQ-MWQ-020, AC-MWQ-020 시나리오 4–6, research.md §R7.

### Q16: 인식된 테스트 구조를 보고하지 않는 명령을 exit 0으로 받아 주는 경계(Q12)를 그대로 두는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: 해당 없음 — 리더가 결정했다. Q12 반영 후에도 `true` 같은 비테스트 명령은 검증기가 가려내지 못한다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b). 그대로 받아들인다. 인식된 테스트 구조를 보고하지 않는 명령은 exit 0이면 유효로 남는다. 검증기가 비테스트 명령을 탐지하지 못한다는 점과, 후보 CI 경로(t1478)가 더 강한 기록이라는 점을 잔여 위험에 적는다. 반영: spec.md §D 잔여 위험 항목, AC-MWQ-016(`true` 행은 유효).
