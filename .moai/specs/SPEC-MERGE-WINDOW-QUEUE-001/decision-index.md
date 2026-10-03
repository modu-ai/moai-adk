# decision-index — SPEC-MERGE-WINDOW-QUEUE-001

> `interview.decision_gate: on`(.moai/config/sections/interview.yaml)에 따라 작성했다. 각 행은 무엇이 왜 미결인지를 밝힐 뿐 권고를 담지 않는다. Q1은 운영자가 이미 답한 승인의 기록이다. Q2–Q6은 v0.1.0 작성 당시 비어 있었고, v0.2.0에서 리더 결정(미션 계약 07d28c4b)을 판정란에 기록했다.
>
> 라벨 표기: 리더 결정은 `Operator verdict:` 줄에 `리더 결정(LEADER-DECIDED)`으로 남기고, `Label:`은 그대로 둔다. 라벨 어휘는 네 개(DECIDED · POLICY-COVERED · EVIDENCE-NEEDED · FOUNDER)로 고정이고, 커밋된 트리에 근거가 없는 결정은 DECIDED로 바꿔 달지 않는다는 규칙 때문이다. 판정은 기록됐고, 이 행들이 그 결정의 첫 커밋 원장이다.

### Q1: 병합 창의 배정을 리더의 지명에서 선착순 대기열로 바꾸는가?

Label: FOUNDER

Authority anchor: — (승인은 리더 세션 df44e022의 AskUserQuestion 답으로 내려졌고, 커밋된 트리에 그 기록이 없다. 그래서 이 행이 승인의 첫 커밋 기록이다.)

Why unresolved: 해당 없음 — 운영자가 결정했다. 현행 로컬 독트린(AGENTS.local.md §4.1 221행 「리더의 창 지명만이 근거다」)과 정면으로 다른 변경이라, 커밋된 권위 근거 없이 정책을 뒤집을 수는 없었다.

Operator verdict: 승인(2026-10-03, 리더 세션 df44e022, AskUserQuestion) — 「병합 창 선착순 대기열」: 리더 창 지명 폐지, 즉시 구현. 리더에게는 open/hold 정책과 push 전 최종 읽기만 남는다.

### Q2: 창 보유자 임대(lease)의 기본 기간은 얼마이고, 기본값으로 켜는가?

Label: EVIDENCE-NEEDED

Authority anchor: —

Why unresolved: 지금의 창에는 임대 개념이 없다(research.md §R1 E9). 재측정을 창 밖으로 옮기면 창 보유 시간이 크게 줄 것으로 보이지만, 새 순서에서 보유 시간이 어떻게 분포하는지는 아직 잰 적이 없다. 너무 짧으면 병합 중인 보유자를 밀어내고, 너무 길면 소유 프로세스는 살아 있는데 멈춘 레인이 대기열 전체를 막는다. REQ-MWQ-009는 값 0이면 꺼진다는 것만 정했다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b — 운영자: 즉시 구현·v3.2.0 포함). 임대는 기본으로 켜고 기간은 30분, 보유자가 부르는 창 동사마다 갱신한다. PID가 죽었거나 임대가 끝난 보유자는 stale이며 대기열 맨 앞이 승격된다. 근거: 재측정을 창 밖으로 옮긴 뒤의 창 안 경로(트리 항등 + no-ff 병합)는 짧으므로 30분이면 넉넉하고, 멈춘 레인이 대기열을 무기한 막지는 못한다. EVIDENCE-NEEDED였던 부분(새 순서에서 창 보유 시간)은 run M0에서 재며, 그 결과로는 기본값을 줄일 수만 있다. 반영: REQ-MWQ-009/009a, AC-MWQ-009/009a, plan M0.

### Q3: `--wait`를 한계값 없이 쓰면 대기 한계는 얼마인가(무한 포함)?

Label: EVIDENCE-NEEDED

Authority anchor: —

Why unresolved: 대기 한계를 정하려면 대기열 길이와 창 보유 시간의 실측이 필요한데, 둘 다 새 메커니즘이 생겨야 잴 수 있다. 무한 대기는 레인 턴을 붙잡고, 짧은 한계는 재진입을 되풀이하게 만든다(재진입하면 대기열 맨 뒤로 간다 — REQ-MWQ-005).

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b — 운영자: 즉시 구현·v3.2.0 포함). 한계 없는 `--wait`의 기본값은 60분이다. 시간이 다 되면 티켓을 거두고, 대기열 위치를 밝히며 0이 아닌 코드로 끝낸다. 근거: 레인 턴을 무기한 붙잡지 않으면서, 30분 임대를 둘 넘게 기다릴 수 있는 길이다. 반영: REQ-MWQ-005, AC-MWQ-005a.

### Q4: 재측정 이후 develop이 움직여 재대기가 필요할 때, 그 레인은 맨 뒤로 가는가, 아니면 순번을 지키는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: REQ-MWQ-023은 창을 놓고 재흡수·재측정을 요구한다는 것까지만 정했다. 맨 뒤로 보내면 develop이 자주 움직이는 부하 구간에서 같은 레인이 거듭 밀려날 수 있다. 순번을 지키게 하면 재측정하는 동안 비어 있는 순번이 뒤 레인들을 막는다. 공정성과 처리량 가운데 무엇을 앞세울지는 정책 판단이고, 이 질문을 덮는 선행 SPEC이나 설정은 없다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b — 운영자: 즉시 구현·v3.2.0 포함). develop이 움직여 재대기하는 티켓은 순번을 지킨다. 재흡수·재측정을 마치면 한 번에 한해 맨 앞으로 다시 들어오고, 연달아 두 번째로 움직이면 맨 뒤로 간다. 근거: 기아를 막으면서, 비어 있는 순번이 뒤 레인들을 오래 막는 일도 한 번으로 묶는다. 반영: REQ-MWQ-024, AC-MWQ-024, plan M6.

### Q5: t1478(착지 전 후보 CI)이 먼저 착지하면, 재측정 근거로 후보 CI 실행과 로컬 재측정 가운데 무엇을 필수로 하는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: REQ-MWQ-021은 둘 중 하나를 받는다. t1478은 아직 착지하지 않았다(research.md §R5). 후보 CI만 받으면 로컬 부하와는 무관해지지만 CI 대기가 생기고, 로컬 재측정만 받으면 darwin/windows 매트릭스를 보지 못한다. 두 카드 가운데 나중에 착지하는 쪽이 이 선택을 고정하게 된다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b — 운영자: 즉시 구현·v3.2.0 포함). t1478의 후보 CI 설정(candidate_ci)이 켜져 있으면 성공한 후보 CI run id를 재측정 기록으로 반드시 요구하고, 꺼져 있으면 로컬 재측정 기록(명령·exit·테스트 수)을 요구한다. 두 형태 모두 같은 검증기가 받는다. 설정이 없으면 꺼진 것으로 읽는다(t1478 미착지 상태). 근거: 후보 CI가 있으면 darwin/windows 매트릭스까지 보는 더 강한 근거를 쓰고, 없으면 로컬 측정으로 대신한다. 반영: REQ-MWQ-021/021a, AC-MWQ-021/021a.

### Q6: 재측정 기록의 「테스트 수」를 언어 중립적으로 어떻게 얻는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: `moai integration`은 사용자 프로젝트로 배포되는 동사라서, 16개 프로그래밍 언어의 러너 출력을 해석하는 방식은 템플릿 중립성 규율(AGENTS.local.md §15)과 부딪친다. 호출자가 수를 적어 넣게 하면 명령과 exit 코드는 관측값이 되지만 수 자체는 다시 주장이 된다. 인식 가능한 러너만 해석하고 나머지는 CI 실행 id를 요구하는 방식도 있다. 셋 모두 관측 강도와 중립성 사이의 맞교환이다.

Operator verdict: 리더 결정(LEADER-DECIDED, 2026-10-03, 미션 계약 07d28c4b — 운영자: 즉시 구현·v3.2.0 포함). 테스트 수는 선택 항목이며 도구가 보고한 값만 쓴다. 기록에는 명령과 exit 코드를 늘 담고, 테스트 수는 도구가 구조화된 보고를 낼 때만 요구한다(이 저장소에서는 `go test -json`의 수). 다른 언어는 exit와 명령만 기록한다. 보고된 수가 0이면 거부한다. 근거: 호출자가 적어 넣는 주장을 없애고, 배포되는 동사의 프로그래밍 언어 중립성(§15)도 지킨다. 반영: REQ-MWQ-021b, REQ-MWQ-032, AC-MWQ-021b/032.

### Q7: 마지막 origin/develop CI가 아직 진행 중이거나 판정이 없을 때 push를 보류하는가?

Label: POLICY-COVERED

Authority anchor: `.claude/rules/local/gitflow-lane-protocol.md` §4 「초록 조건부」 단락 — "마지막 push의 CI 판정이 아직 없으면(새 develop의 첫 push 등) red가 아니므로 보류 사유가 아니다."

Why unresolved: 해당 없음 — 커밋된 로컬 규율이 문면 그대로 이 질문을 덮는다(REQ-MWQ-041). 다만 CI 상태를 아예 읽지 못하는 경우는 「판정 없음」이 아니라 「측정 못 함」이라서 이 정책이 덮지 않는다. 그 경우는 REQ-MWQ-042가 거부로 정했다.

Operator verdict:
