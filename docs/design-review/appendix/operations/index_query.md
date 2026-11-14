# `index-query`

operation `index.query` · 코드 `src/ops/index_query.rs` · C6 #2182

`index-query PROJECT` 는 증분 색인의 게시된 세대를 찾아 fresh/stale/partial/unavailable 중 하나와 찾은 변경(추가·삭제·수정·추적 안 된 새 파일·가지 전환)을 알린다. 읽기만, stdin 의 typed 요청. `src/main.rs` 에서 그대로 옮김(#2834 R2c).
