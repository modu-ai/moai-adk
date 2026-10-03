# decision-index — SPEC-MERGE-WINDOW-QUEUE-001

> `interview.decision_gate: on`(.moai/config/sections/interview.yaml)에 따라 작성했다. 각 행은 무엇이 왜 미결인지를 밝힐 뿐 권고를 담지 않는다. Q1은 운영자가 이미 답한 승인의 기록이고, Q2 이하의 `Operator verdict:`는 작성 시점에 비어 있다.

### Q1: 병합 창의 배정을 리더의 지명에서 선착순 대기열로 바꾸는가?

Label: FOUNDER

Authority anchor: — (승인은 리더 세션 df44e022의 AskUserQuestion 답으로 내려졌고, 커밋된 트리에 그 기록이 없다. 그래서 이 행이 승인의 첫 커밋 기록이다.)

Why unresolved: 해당 없음 — 운영자가 결정했다. 현행 로컬 독트린(AGENTS.local.md §4.1 221행 「리더의 창 지명만이 근거다」)과 정면으로 다른 변경이라, 커밋된 권위 근거 없이 정책을 뒤집을 수는 없었다.

Operator verdict: 승인(2026-10-03, 리더 세션 df44e022, AskUserQuestion) — 「병합 창 선착순 대기열」: 리더 창 지명 폐지, 즉시 구현. 리더에게는 open/hold 정책과 push 전 최종 읽기만 남는다.

### Q2: 창 보유자 임대(lease)의 기본 기간은 얼마이고, 기본값으로 켜는가?

Label: EVIDENCE-NEEDED

Authority anchor: —

Why unresolved: 지금의 창에는 임대 개념이 없다(research.md §R1 E9). 재측정을 창 밖으로 옮기면 창 보유 시간이 크게 줄 것으로 보이지만, 새 순서에서 보유 시간이 어떻게 분포하는지는 아직 잰 적이 없다. 너무 짧으면 병합 중인 보유자를 밀어내고, 너무 길면 소유 프로세스는 살아 있는데 멈춘 레인이 대기열 전체를 막는다. REQ-MWQ-009는 값 0이면 꺼진다는 것만 정했다.

Operator verdict:

### Q3: `--wait`를 한계값 없이 쓰면 대기 한계는 얼마인가(무한 포함)?

Label: EVIDENCE-NEEDED

Authority anchor: —

Why unresolved: 대기 한계를 정하려면 대기열 길이와 창 보유 시간의 실측이 필요한데, 둘 다 새 메커니즘이 생겨야 잴 수 있다. 무한 대기는 레인 턴을 붙잡고, 짧은 한계는 재진입을 되풀이하게 만든다(재진입하면 대기열 맨 뒤로 간다 — REQ-MWQ-005).

Operator verdict:

### Q4: 재측정 이후 develop이 움직여 재대기가 필요할 때, 그 레인은 맨 뒤로 가는가, 아니면 순번을 지키는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: REQ-MWQ-023은 창을 놓고 재흡수·재측정을 요구한다는 것까지만 정했다. 맨 뒤로 보내면 develop이 자주 움직이는 부하 구간에서 같은 레인이 거듭 밀려날 수 있다. 순번을 지키게 하면 재측정하는 동안 비어 있는 순번이 뒤 레인들을 막는다. 공정성과 처리량 가운데 무엇을 앞세울지는 정책 판단이고, 이 질문을 덮는 선행 SPEC이나 설정은 없다.

Operator verdict:

### Q5: t1478(착지 전 후보 CI)이 먼저 착지하면, 재측정 근거로 후보 CI 실행과 로컬 재측정 가운데 무엇을 필수로 하는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: REQ-MWQ-021은 둘 중 하나를 받는다. t1478은 아직 착지하지 않았다(research.md §R5). 후보 CI만 받으면 로컬 부하와는 무관해지지만 CI 대기가 생기고, 로컬 재측정만 받으면 darwin/windows 매트릭스를 보지 못한다. 두 카드 가운데 나중에 착지하는 쪽이 이 선택을 고정하게 된다.

Operator verdict:

### Q6: 재측정 기록의 「테스트 수」를 언어 중립적으로 어떻게 얻는가?

Label: FOUNDER

Authority anchor: —

Why unresolved: `moai integration`은 사용자 프로젝트로 배포되는 동사라서, 16개 프로그래밍 언어의 러너 출력을 해석하는 방식은 템플릿 중립성 규율(AGENTS.local.md §15)과 부딪친다. 호출자가 수를 적어 넣게 하면 명령과 exit 코드는 관측값이 되지만 수 자체는 다시 주장이 된다. 인식 가능한 러너만 해석하고 나머지는 CI 실행 id를 요구하는 방식도 있다. 셋 모두 관측 강도와 중립성 사이의 맞교환이다.

Operator verdict:

### Q7: 마지막 origin/develop CI가 아직 진행 중이거나 판정이 없을 때 push를 보류하는가?

Label: POLICY-COVERED

Authority anchor: `.claude/rules/local/gitflow-lane-protocol.md` §4 「초록 조건부」 단락 — "마지막 push의 CI 판정이 아직 없으면(새 develop의 첫 push 등) red가 아니므로 보류 사유가 아니다."

Why unresolved: 해당 없음 — 커밋된 로컬 규율이 문면 그대로 이 질문을 덮는다(REQ-MWQ-041). 다만 CI 상태를 아예 읽지 못하는 경우는 「판정 없음」이 아니라 「측정 못 함」이라서 이 정책이 덮지 않는다. 그 경우는 REQ-MWQ-042가 거부로 정했다.

Operator verdict:
