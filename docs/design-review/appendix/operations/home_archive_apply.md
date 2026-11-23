# `home-archive-apply`

operation `home_archive.apply` · 코드 `src/ops/home_archive_apply.rs`, `src/home_state_retire.rs` · SPEC-20261123-home-state-retire

`home-archive-apply [--home DIR] --now UTC (--plan-digest D | --content-digest D) --permission-reference REF` 는 같은 입력으로 다시 계산한 미리보기의 digest 가 같고 동작이 `retire` 일 때만, 먼저 `~/.moai/archive/state-retire-manifest-<stamp>.json` 과 이동 저널(`move-journal-<stamp>.json`)을 쓰고 옛 상태 폴더를 `~/.moai/archive/moai-cli-<stamp>/` 로 rename 한 번에 옮긴다(덮어쓰지 않음, 같은 파일 시스템, 링크는 따라가지 않음). 복사도 열기도 삭제도 없어서 바이트·모드·vault 의 0700/0600 이 그대로이고, `~/.moai` 의 `archive` 밖은 바꾸지 않는다. 가지(`~/.moai`, `archive`)가 없으면 0700 으로 만든다. 허가 참조가 없거나 digest 가 어긋나면 아무것도 쓰지 않는다. 같은 `--now` 의 저널이 크래시로 남았으면(매니페스트의 digest 가 같을 때) 그 저널을 끝내고, 다른 계획의 저널이면 거절한다. 시험용 크래시 지점은 디버그 빌드의 `MOAI_CLI_TEST_STATE_RETIRE_FAULT`. 실제 HOME 에서는 사용자가 확인한 뒤에만 리더가 실행한다.
