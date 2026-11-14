# `gate-merge`

operation `gate.merge` · 코드 `src/ops/gate_merge.rs`, `src/gate_merge.rs` · Issue #2674 S2

`gate-merge --pr N --permission-reference REF [--repo OWNER/REPO] [--dir .] [--tip …] [--review-confirmed …]` 는 리더 병합의 네 단계를 차례로 한다. 앞 단계가 성공했을 때만 다음으로 간다: ① `gate-check --pr` 와 같은 스냅숏을 모아 규칙 G1~G7 을 본다(멈추면 아무것도 쓰지 않는다), ② head 에 `moai/audit` success 상태를 올린다(`MOAI_AUDIT_APP_TOKEN` 이 있으면 그 토큰으로), ③ `sha` 를 head 로 고정해 병합한다(head 가 바뀌면 GitHub 가 거부; 405 는 상태가 아직 등록되지 않았을 수 있어 5초 뒤 다시, 최대 3번), ④ 병합됐을 때만 병합 기록 댓글을 단다. 게이트가 기계적 delta(#2833, `gate-check` 문서)로 통과했으면 ② 앞에 PR 에 `AUTO-DELTA #N <H>→<N>` 첫 줄 댓글(근거: 파일별 patch-id, G5 가 읽은 N 의 `local-check:` pass 줄)을 한 번 단다 — 같은 첫 줄이 이미 있으면 다시 달지 않고, 이 기록이 실패하면 다른 쓰기를 하지 않는다. 결과의 `writes` 는 실제로 한 쓰기(`auto_delta`·`status`·`merge`·`record`)만 적는다. 권한 참조 없이는 아무것도 읽거나 쓰지 않는다.

감사 등급(#2906 T, 사용자 결정 6878414880, 규칙은 `gate-check` 문서): 결과에는 언제나 `effective_tier` 가 붙는다. 실효 0 등급 PR 은 판정 없이 병합되므로, `moai/audit` 상태의 링크는 G5 가 읽은 `local-check:` 줄의 댓글을 가리키고 설명은 `audit-tier 0 (no audit) head …; local-check pass; …` 이다. 병합 기록은 판정 대신 `audit-tier: 0 (실효 등급 0, 감사 없음 …)`, `local-check <댓글 id>:` 다음에 그 줄 전체, 그리고 head 와 merge base 를 적고, 결과의 `verdict` 는 `null`, `local_check` 는 그 댓글 id 다. 0 등급에는 기계적 delta 기록을 쓰지 않는다(옮길 PASS 가 없다). 1·2 등급 기록은 전처럼 `판정 <id> PASS(…)` 로 시작하고 `audit-tier: <등급>` 을 덧붙인다.
