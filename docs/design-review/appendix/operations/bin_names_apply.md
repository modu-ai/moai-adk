# `bin-names-apply`

operation `bin_names.apply` · 코드 `src/ops/bin_names_apply.rs`, `src/bin_names.rs` (#2791)

`bin-names-apply [--home DIR] --plan-digest D --permission-reference R [--now UTC]` 는 다시 계획해 digest 가 같을 때만 움직인다(permission 이 비면 거부). `create` 는 `bin/moai` 를 `bin/moai-cli` 의 hard link 로 만든다. `replace_copy` 는 따로 된 사본을 `moai.replaced-<UTC>` 로 옮긴 뒤(지우지 않음, 이름이 있으면 거부) 링크하고, 링크가 실패하면 사본을 제자리로 되돌린다. 그 밖의 상태는 쓰지 않는다. 링크 뒤 `bin/moai` 의 sha256 이 `bin/moai-cli` 와 같은지 확인한다. 결과: `state`(`applied` 또는 계획 상태), `linked`, `to`, `moved_aside`, `written`. `moai-cli update` 미리보기의 `bin_names` 부분과 같은 계획이다.
