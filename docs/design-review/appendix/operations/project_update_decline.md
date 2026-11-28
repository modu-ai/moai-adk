# `project-update-decline`

operation `project_update.decline` · 코드 `src/ops/project_update_decline.rs`, `src/project_update/declined.rs`

`project-update-decline --step STEP --source-sha256 SHA --permission-reference REF [PROJECT]` 는 `project-update-plan` 의 선택 단계(`claude-md`, `agents-picks`, `agents-local-merge`, `claude-settings-merge`) 하나를 사용자가 거절했다고 프로젝트 상태 폴더의 `manifests/update-declined.json` 에 기록한다. 기록은 단계 이름과 그 단계가 제안하는 글의 출처 sha256 이다. `claude-md` 의 출처 sha256 은 그 단계가 옮길 모든 파일의 경로와 sha256 을 묶은 값이라, 그 파일 하나의 바이트가 바뀌거나 파일이 늘거나 줄면 다시 묻는다([`claude-md-plan`](claude_md_plan.md)). 권한 참조가 비었거나, 모르는 단계 이름이거나, `--source-sha256` 이 지금 그 단계의 출처 sha256 과 다르거나, 단계에 출처가 없거나, 상태 폴더나 기록 파일이 링크이면 아무것도 쓰지 않는다. 같은 기록이 이미 있으면 `state: unchanged` 로 끝난다. 이 명령은 그 파일 하나만 쓰고 프로젝트의 다른 파일은 쓰지도 지우지도 않는다.

`--step leftover:<경로>` 는 `project-update-plan` 의 `leftovers` 다섯 종류(`.moai-cli/`, `.moai.rust-<UTC>/`, `.moai-backups/`, `.moai/.update.lock`, `.moai-backup/.update.lock`) 하나를 사용자가 그대로 두기로 했다고 같은 파일에 `leftover:<경로>` 키로 기록한다. 출처 sha256 은 폴더면 맨 위 항목의 이름·종류·파일 크기의 목록, 잠금 파일이면 그 바이트의 sha256 이다. `--source-sha256` 이 지금 출처와 다르거나 다섯 종류가 아닌 경로이면 아무것도 쓰지 않고, 항목이 바뀌면 기록은 읽히지 않아 다시 묻는다. 남긴 항목은 `all_unchanged` 를 막지 않는다.

`--ask-again` 은 그 단계의 기록만 지우고 다른 단계의 기록은 그대로 둔다(`--source-sha256` 과 함께 쓰지 않는다). 기록이 없으면 `state: unchanged` 로 끝난다.

`project-update-plan` 은 기록의 sha256 이 지금 출처와 같고 단계가 `needs_decision` 이나 `conflict` 이면 그 단계를 `skip` 으로 읽고 이유에 거절했음과 다시 묻는 명령을 적는다. 출처의 바이트가 바뀌면 기록은 읽히지 않고 단계는 다시 묻는다.
