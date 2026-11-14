# `forge-read`

operation `forge.read` · 코드 `src/ops/forge_read.rs`

`forge-read PROJECT` 는 GitHub Issue 와 댓글을 할 일(TODO)로 읽는다. GET 만, stdin 의 typed 요청(64 KiB 이하), 봉투는 stdout. 폴링 상태는 `<PROJECT>/.moai-cli/forge/poll-state.json`. stderr 가 터미널이면 진행 줄을 보인다. `src/main.rs` 에서 그대로 옮김(#2834 R2b).
