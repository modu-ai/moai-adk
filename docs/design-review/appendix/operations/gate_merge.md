# `gate-merge`

operation `gate.merge` · 코드 `src/ops/gate_merge.rs`, `src/gate_merge.rs` · Issue #2674 S2

`gate-merge --pr N --permission-reference REF [--repo OWNER/REPO] [--dir .] [--tip …] [--review-confirmed …]` 는 리더 병합의 네 단계를 차례로 한다. 앞 단계가 성공했을 때만 다음으로 간다: ① `gate-check --pr` 와 같은 스냅숏을 모아 규칙 G1~G7 을 본다(멈추면 아무것도 쓰지 않는다), ② head 에 `moai/audit` success 상태를 올린다(`MOAI_AUDIT_APP_TOKEN` 이 있으면 그 토큰으로), ③ `sha` 를 head 로 고정해 병합한다(head 가 바뀌면 GitHub 가 거부; 405 는 상태가 아직 등록되지 않았을 수 있어 5초 뒤 다시, 최대 3번), ④ 병합됐을 때만 병합 기록 댓글을 단다. 결과의 `writes` 는 실제로 한 쓰기(`status`·`merge`·`record`)만 적는다. 권한 참조 없이는 아무것도 읽거나 쓰지 않는다.
