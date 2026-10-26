# decision-index.md — SPEC-FACTORY-COMPLETION-RECOVERY-001

v0.2.0(카드 t1963, M2-B) 개정 중 드러난 미결 판정이다. `interview.decision_gate: on`(`.moai/config/sections/interview.yaml`)에 따라 작성했다. 권위 등록부는 커밋된 산출물뿐이며, 리더 분석 `.moai/reports/t1963/zombie-analysis.md`와 카드 배차 문구는 미추적 카드 증거라 권위로 인용하지 않는다. 기본값 규칙(순서대로): 현재 동작을 보존하는 선택지 → 이 SPEC 자신의 커밋 되돌리기 한 번으로 원복되는 선택지 → 사용자에게 보이는 면이 더 작은 선택지.

### Q1: `moai factory runs`의 재결속 표면을 `--rebind-to` 플래그로 둘 것인가, 하위 명령으로 둘 것인가?

Label: FOUNDER
Class: product-level
Authority anchor: 운영자 판정의 출처 — 카드 t1963 배차 본문(운영자 지시 2026-10-26), 레인 결정 기록 `.moai/reports/t1963/progress.md`. 커밋된 권위 등록부의 항목은 아니므로 라벨은 FOUNDER로 두고 판정은 아래 운영자 판정 줄에 기록한다
Why unresolved: 카드는 `moai factory runs --retire <id> --rebind-to <live-run>` 모양과 "SPEC-FACTORY-RETIRE-GUARD-001 시험은 바꾸지 않고 통과"를 함께 요구한다. 그런데 그 SPEC의 `TestFactoryRunsHasNoForceFlag`(`internal/cli/factory_runs_retire_guard_test.go:234-242`)는 플래그 집합이 정확히 `[retire]`임을 단언한다(acceptance.md §0.1 L3에서 PASS 관측). 플래그를 더하면 이 시험이 실패하고, 시험을 지키면 카드가 지정한 명령 모양이 불가능하다. 선택지: (a) 시험의 허용 집합을 `[rebind-to retire]`로 넓힌다 (b) 하위 명령(예: `moai factory runs rebind <id> --to <run>`)으로 두고 시험을 그대로 둔다.
Default: (b) 하위 명령 (rule: 현재 동작 보존 — runs 명령의 플래그 집합과 그 시험이 바뀌지 않는다)
Alternate: (a) 플래그 + 시험 허용 집합 확대
Operator verdict: DECIDED 2026-10-26 lane-11 (authority anchor: card t1963 body; see .moai/reports/t1963/progress.md decision record) — (a): `--rebind-to`는 `moai factory runs --retire <id> --rebind-to <live-run>`의 플래그다. 기존 시험 수정은 `TestFactoryRunsHasNoForceFlag`의 허용 플래그 집합 단언을 `[retire]`에서 `[rebind-to retire]`로 넓히는 한 건만 허용된다. `--rebind-to` 없는 은퇴는 SPEC-FACTORY-RETIRE-GUARD-001 거부 동작 그대로다.

### Q2: 리더 진입이 `--resume` 없이 죽은 소유자 run을 채택하고 나머지 죽은 run을 통합하게 할 것인가?

Label: FOUNDER
Class: product-level
Authority anchor: 운영자 판정의 출처 — 카드 t1963 배차 본문(운영자 지시 2026-10-26), 레인 결정 기록 `.moai/reports/t1963/progress.md`. 커밋된 권위 등록부의 항목은 아니므로(현행 커밋 동작은 resume 인자를 요구한다, `internal/cli/factory_run_adopt.go:165`) 라벨은 FOUNDER로 두고 판정은 아래 운영자 판정 줄에 기록한다
Why unresolved: 출시된 `moai cc -f`·`moai glm -f`의 기본 동작이 바뀐다 — 새 리더가 이전 run id를 이어받고, 다른 죽은 run의 카드를 옮기고 은퇴시킨다. 카드(리더 배차)가 요청한 동작이지만 그 요청은 커밋된 권위가 아니다. 기존 시험 `TestAdoptResumedLeaderRunWarnsOnMultipleCandidates`의 기대도 반전된다. 선택지: (a) REQ-FCR-017/018대로 자동 채택·통합 (b) resume 게이트를 유지하고 통합만 운영자 표면(REQ-FCR-021)으로 제공.
Default: (b) resume 게이트 유지 (rule: 현재 동작 보존)
Alternate: (a) resume 없이 자동 채택·통합 — 카드의 완료 기준("리더 진입 한 번에 활성 run 1개")은 (a)에서만 성립한다
Operator verdict: DECIDED 2026-10-26 lane-11 (authority anchor: card t1963 body; see .moai/reports/t1963/progress.md decision record) — (a): 리더 진입은 `--resume` 없이 가장 최신의 죽은 소유자 run을 채택하고 나머지를 통합한다.

### Q3: 통합 시 대상 run에 같은 카드 행이 이미 있으면 어느 행을 남기는가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 없음
Why unresolved: `cards`의 키가 `(run_id, card_id)`라 재키가 충돌한다. 어느 행이 더 진행된 상태인지는 일반적으로 알 수 없다. 선택지: (a) 대상 run의 행을 남기고 원천 행은 `abandoned` + `card.rebound`(이전 상태 전부) (b) `updated_at`이 더 최근인 행을 남김.
Default: (a) 대상 행 유지 + 원천 행 포인터 포기 (rule: 현재 동작 보존 — 현행 코드는 현재 run의 행만 임대·전이 대상으로 삼는다, `internal/homestate/card_picked.go:159-176`)
Alternate: (b) 최근 갱신 행 유지
Operator verdict: DEFAULT-APPLIED 2026-10-26T01:29:43Z lane-11+manager-spec

### Q4: 임대(`moai factory next`) 시 죽은 run의 중복 행을 재키할 것인가, 포인터 포기할 것인가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 없음
Why unresolved: 카드는 둘 다 허용했다. 재키는 원래 행의 상태를 그대로 옮기지만 임대 경로의 T1(`picked` 생성)·이어받기(t1930) 의미론과 겹친다. 포인터 포기는 기존 생성·이어받기 경로를 그대로 쓰고 원천 행을 종결시킨다.
Default: 포인터 포기 (rule: 사용자에게 보이는 면이 더 작음 — 기존 `RecordPicked`·이어받기 경로를 바꾸지 않는다; 앞의 두 규칙은 두 선택지를 가르지 못한다)
Alternate: 재키
Operator verdict: DEFAULT-APPLIED 2026-10-26T01:29:43Z lane-11+manager-spec

### Q5: 합류 프롬프트 `m<n>`은 산 run이 둘 이상일 때도 제시하는가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 없음
Why unresolved: 대상 run이 여럿이면 어디로 옮길지 추가 선택이 필요하다.
Default: 산 소유자 활성 run이 정확히 하나일 때만 제시 (rule: 사용자에게 보이는 면이 더 작음)
Alternate: 대상 선택 단계를 덧붙여 항상 제시
Operator verdict: DEFAULT-APPLIED 2026-10-26T01:29:43Z lane-11+manager-spec

### Q6: worker 생존 판정에 프로세스 시작 지문 열을 추가할 것인가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 없음
Why unresolved: `workers`에는 지문 열이 없어(`internal/homestate/factory.go:30-38`) pid만으로 판정하면 pid 재사용 시 죽은 worker가 살아 보인다(안전한 방향의 오판). 열을 더하면 정확해지지만 스키마 변경이다.
Default: pid만으로 판정, 확실한 dead만 제외 (rule: 현재 동작 보존 — 스키마 불변)
Alternate: 지문 열 추가
Operator verdict: DEFAULT-APPLIED 2026-10-26T01:29:43Z lane-11+manager-spec

### Q7: 산 소유자 활성 run이 이미 있을 때 리더 진입이 죽은 run을 채택하는가?

Label: FOUNDER
Class: implementation-level
Authority anchor: 없음
Why unresolved: 산 리더가 있는 상태에서 두 번째 리더가 죽은 run을 채택하면 활성 run이 둘(둘 다 산 소유자)이 되어 모호성이 정당하게 남는다. 채택하지 않으면 현재처럼 새 run을 발급한다.
Default: 채택하지 않음 — 변경 전 동작 (rule: 현재 동작 보존)
Alternate: 그래도 채택
Operator verdict: DEFAULT-APPLIED 2026-10-26T01:29:43Z lane-11+manager-spec
