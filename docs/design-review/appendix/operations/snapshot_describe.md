# `snapshot-describe`

operation `snapshot.describe` · 코드 `src/ops/snapshot_describe.rs` · 07-D #2011

`snapshot-describe PROJECT` 는 PROJECT 의 HEAD 커밋과 추적 중인 변경의 digest 를 보여 준다. 읽기만, stdin 은 읽지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R2b).
