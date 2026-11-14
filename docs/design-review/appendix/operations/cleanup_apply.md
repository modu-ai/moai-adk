# `cleanup-apply`

operation `cleanup.apply` · 코드 `src/ops/cleanup_apply.rs`, `src/update_cleanup.rs` · Issue #2872 (1단계)

`cleanup-apply HOME STAGE --confirm-cleanup TOKEN --permission-reference REF [--now UTC]` 는 `cleanup-plan` 을 다시 계산하고, 그 `confirm_token` 이 TOKEN 과 같을 때만 archive 항목을 옮긴다. 다르면 아무것도 쓰지 않는다.

- 보관 폴더는 `<상태 폴더>/cleanup/<YYYYMMDDTHHMMSSZ>` 이다(상태 폴더는 `home_root::path`). 새로 만들고(`create_dir`), 이미 있으면 `.2`, `.3` … `.64` 를 쓴다.
- 옮기기 전에 모든 원본의 부모 폴더와 보관 폴더가 같은 파일 시스템인지(unix 장치 번호) 본다. 하나라도 다르면 아무것도 옮기지 않는다.
- `manifest.tsv`(`<원래 절대 경로>\t<보관 절대 경로>` 줄)와 `receipt.json`(상태 `pending`, 항목·digest)을 임시 파일·fsync·rename 으로 먼저 쓴다.
- 항목마다 digest 가 그대로인지 다시 보고(다르면 그 항목에 `changed` 를 적고 멈추며 receipt 는 `pending` 으로 남는다), 보관 폴더 안에 HOME 기준 경로를 그대로 만들어 `rename` 한다. 옮긴 항목은 `moved: true`. 끝나면 `committed`. 그 뒤 `cleanup-plan` 의 archive 항목은 0 이다. 지우지 않는다.
