# `flake-record`

operation `flake.record` · 코드 `src/ops/flake_record.rs`, `src/flake_record.rs` · Issue #2666 F2

`flake-record --classified FILE --run ID --permission-reference REF [--repo OWNER/REPO]` 는 `flake-classify` 결과의 `flaky` 행마다 그 시험을 추적하는 열린 Issue 를 찾는다. Issue 본문 첫 줄의 숨은 표지 `<!-- moai:flake test=<시험> -->` 로 찾고, 있으면 이 run·job 을 댓글로 적고, 없으면 `[flake] <시험>` Issue 를 만든다. run·job 마다 숨은 표지 `<!-- moai:flake-run run=<id> job=<job> -->` 를 남겨, 이미 적힌 run·job 은 다시 쓰지 않는다(같은 명령을 두 번 돌려도 쓰기 0). 같은 시험을 추적하는 열린 Issue 가 둘 이상이면 멈춘다. 쓰기 뒤 전송 오류가 나면 목록을 다시 읽어 실제로 쓰였는지 확인하고, 다시 읽기도 실패하면 `uncertain` 에 적는다. permission reference 가 없으면 아무것도 읽거나 쓰지 않는다.
