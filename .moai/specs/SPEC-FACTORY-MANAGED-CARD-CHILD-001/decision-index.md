# decision-index.md — SPEC-FACTORY-MANAGED-CARD-CHILD-001

> 상태축 없음(stateless). 계획 단계에서 운영자가 인터뷰에서 정하지 않은 결정 3건의 기록이다. 세 건 모두 리더가 결정했다(mission contract 11c79e1a). 근거인 리더 메시지는 커밋된 문서가 아니므로 권한 앵커로 쓰지 않고 `FOUNDER` 로 라우팅한다 — 리더는 나중에 개정(amendment)으로 뒤집을 수 있다.

### Q1: 관리 카드 자식에 카드 시작 프롬프트를 주입하는가?

Label: FOUNDER
Authority anchor: (none — committed tree has no clause deciding this; `.moai/docs/factory-managed-session.md` and the parent SPEC are silent)
Why unresolved: 드라이버는 우선 턴(`managedPrimingPrompt`, `internal/cli/managed_factory_session.go:55`) 뒤 유휴로 들어가 운영자 입력이나 브로커 메시지를 기다린다. 주입하면 카드 시작이 자동이 되지만 레인 루프의 책임 범위와 부모 REQ-MS-014(카드 처분 판정 금지)와의 경계가 걸린다.
Operator verdict: 주입하지 않는다(범위 밖). decided_by=leader · evidence_refs=리더 메시지, mission contract 11c79e1a · ladder_path=lead chat step 5 · 귀결: 무인 레인은 첫 카드 이후로 진행하지 않는다(spec.md Known limits, 문서 AC-CC-012, progress.md Residual-risk, 후속 카드 문안은 progress.md).

### Q2: stdin EOF 인 비대화형 레인에서 관리 카드 세션을 끝낼 방법을 이 SPEC 이 만드는가?

Label: FOUNDER
Authority anchor: (none)
Why unresolved: 드라이버는 EOF 에서 끝나지 않는다(`managed_factory_session.go:315-319`). EOF 를 종료로 읽게 바꾸면 플레인 관리 경로의 현행 동작이 바뀌고, 리더가 소유자·드라이버를 이 카드 범위 밖으로 막았다.
Operator verdict: 드라이버 끝 조건은 그대로 둔다. decided_by=leader · evidence_refs=리더 메시지, mission contract 11c79e1a · ladder_path=lead chat step 5.

### Q3: 연속 시작 실패에 레인 루프의 상한을 두는가?

Label: FOUNDER
Authority anchor: (none)
Why unresolved: 현행 직접 문 루프에도 상한이 없고(`internal/cli/codex_launcher.go:1000-1005`), 관리 경로는 시작 실패 요인(`app-server` 하위 명령 부재, 10초 핸드셰이크 시간 초과)이 더 많다.
Operator verdict: 상한 없음(현행과 같다). decided_by=leader · evidence_refs=리더 메시지, mission contract 11c79e1a · ladder_path=lead chat step 5 · 귀결: 관리 시작이 계속 실패하면 큐 카드가 연속으로 lease 되어 큐가 소진될 수 있다(progress.md Residual-risk).

### Q1–Q3 결정과 plan-audit r2 처분(PASS-WITH-DEBT D2/D3)의 권한 근거

Label: FOUNDER
Authority anchor: (none in-tree — 감사자가 커밋된 트리에서 볼 수 없는 근거를 여기에 기록한다)
Why unresolved: Q1–Q3 결정과 D2·D3 의 PASS-WITH-DEBT 처분이 리더 메시지로만 전달됐다.
Operator verdict: 리더 결정, 미션 계약 11c79e1a(운영자 승인 2026-10-03), 리더 메시지로 전달. decided_by=leader · 범위: Q1·Q2·Q3 와 plan-audit r2(FAIL 0.88)의 PASS-WITH-DEBT 처분(D2/D3), 재감사 없이 run 진입. 부채는 progress.md "Run-phase first obligations" 에 있다.
