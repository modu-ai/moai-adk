# `agents-merge-apply`

operation `agents_merge.apply` · 코드 `src/ops/agents_merge_apply.rs`, `src/agents_local_merge.rs` · Issue #2936 (R10)

`agents-merge-apply --plan-digest DIGEST --permission-reference REF [PROJECT]` 는 사용자가 확인한 `agents-merge-plan` 을 적용한다. 권한 참조가 없거나 `plan_digest` 가 지금 계획과 다르면(그 사이 `AGENTS.md`·출처가 바뀜) 아무것도 쓰지 않는다. `action` 이 `merge` 일 때만 쓴다: `AGENTS.md` 끝에 빈 줄 하나 뒤 `<!-- moai:merged-local sha256=<출처 sha256> -->` … `<!-- /moai:merged-local -->` 블록을 더한다(관리 구역은 건드리지 않고, 관리 구역을 다시 만들어도 블록은 남는다). `unchanged`·`none` 은 `state: unchanged` 로 끝나고, `conflict`·`blocked` 는 오류다. 영수증 `<상태 폴더>/receipts/agents-local-merge-<id>/` 에 이전 `AGENTS.md`(`before/`)를 두고, 제자리 파일이 출처면 사본(`source/`)도 둔다. 영수증 폴더는 모두 0700, 파일은 모두 0600 이다(umask 와 무관; 사본에 토큰이 있을 수 있다). `receipt.json` 은 먼저 `pending`(쓰기 전, `agents_after` 에 쓸 바이트의 해시를 미리 적음)으로 두고 `AGENTS.md` 를 쓴 뒤 `applied` 로 바꾼다. 출처 파일은 옮기거나 지우지 않는다.
