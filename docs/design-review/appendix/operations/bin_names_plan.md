# `bin-names-plan`

operation `bin_names.plan` · 코드 `src/ops/bin_names_plan.rs`, `src/bin_names.rs` (#2791)

`bin-names-plan [--home DIR]` 는 이 바이너리가 도는 버전 폴더(`~/.moai-cli/versions/<version>/bin/`, 어느 이름으로 실행했든)에서 `moai` 가 `moai-cli` 의 hard link 인지 본다. 아무것도 쓰지 않는다. `state`: `create`(`moai` 없음), `unchanged`(같은 바이트), `replace_copy`(같은 빌드의 따로 된 사본 — `--version` 첫 줄이 이 빌드의 `moai-cli <ver> (rust)` 와 같음), `blocked`(`moai-cli` 가 없거나 일반 파일이 아님, `moai` 가 링크·폴더, 또는 다른 프로그램), `outside_layout`(버전 폴더 밖). 결과에 `reason`, `canonical`, `alias`, 두 sha256, `plan_digest`. PATH·셸 설정과 다른 곳의 `moai`(Go 포함)는 건드리지 않는다 — 그건 `install-inspect` 의 `moai_candidates` 가 보고한다.
