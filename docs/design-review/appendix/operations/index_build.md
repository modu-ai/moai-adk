# `index-build`

operation `index.build` · 코드 `src/ops/index_build.rs` · C6 #2182

`index-build PROJECT` 는 추적 중인 파일을 훑어 바뀌지 않는 manifest 폴더를 만들고, 원자적 포인터 교체로 증분 색인 세대를 게시한다. 실패하면 앞 세대를 두고 partial 로 표시한다(REQ-C6-01). stdin 의 typed 요청, stderr 가 터미널이면 단계 줄을 보인다. `src/main.rs` 에서 그대로 옮김(#2834 R4b).
