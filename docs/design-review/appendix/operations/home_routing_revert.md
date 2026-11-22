# `home-routing-revert`

operation `home.routing_revert` · 코드 `src/ops/home_routing_revert.rs`, `src/home_routing.rs`, `src/home_routing/rules.rs` (#2638)

`home-routing-revert HOME --receipt ID --permission-reference R` 는 영수증 하나를 되돌린다. 설정 폴더는 선택 인자가 아니라 영수증 항목에서 읽는다. 항목을 마지막부터 거꾸로 처리하고, 적용 뒤 바뀌지 않은 파일만 돌린다: 만든 파일은 지우고(그 apply 가 만든 폴더는 비어 있을 때만 지운다), 바꾼 파일과 지운 옛 블록은 before 바이트로 돌리며 소유 기록도 적용 전 기준으로 되돌린다(옛 블록의 기준이 돌아와 지우기가 다시 제안된다). 그 뒤 사용자가 바꾼 파일은 `user_changed_since_apply`, 링크가 된 폴더·파일은 `host_folder_not_plain`·`file_not_plain` 으로 건너뛰고, 건너뛴 항목은 `host`, `kind`, `path`, `reason` 으로 파일을 이름으로 댄다(은퇴한 호스트 항목은 `host`, `reason: host_retired` 만). 나머지 파일은 계속 되돌린다. `kind` 가 없는 옛 영수증 항목은 옛 블록으로 읽는다. 같은 영수증을 두 번 되돌리지 않는다. 결과: `state`(`reverted`), `receipt`, `reverted`(되돌린 호스트, 한 번씩), `files`(되돌린 파일), `skipped`.
