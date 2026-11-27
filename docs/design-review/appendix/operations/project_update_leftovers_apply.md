# `project-update-leftovers-apply`

operation `project_update.leftovers_apply` · 코드 `src/ops/project_update_leftovers_apply.rs`, `src/project_update/leftovers_archive.rs` · SPEC-20261127-update-leftovers-action

`project-update-leftovers-apply --plan-digest DIGEST --permission-reference REF --item PATH [--item PATH ...] [PROJECT]` 는 `project-update-plan` 이 `action` 과 함께 보여 준 이전 세대의 leftover 를 백업으로 옮긴다. 대상은 이전 상태 폴더, 이전 revert 가 치워 둔 `<상태 폴더>.rust-<UTC>/`, Go `.moai-backups/`, 상태 폴더와 Go 백업 폴더 안의 죽은 Go 업데이트 lock 이다. 폴더는 rename 한 번으로, lock 은 hard link 한 번과 unlink 한 번으로 `<상태 폴더>/backups/<UTC>/leftovers/<같은 경로>` 에 옮기고, 덮어쓰지 않으며 아무것도 지우지 않는다.

권한 참조가 계약에서 거절되거나, `--item` 이 없거나, 알려진 다섯 경로가 아니거나(절대 경로, `..`, 다른 곳), 두 번 적혔거나, 링크를 지나거나, `plan_digest` 가 지금 새로 만든 계획의 digest 가 아니거나(`stale_plan`), 항목이 지금 계획에 없거나 보관할 수 없는 항목(`keep_only`)이면 아무것도 쓰지 않는다. 사전 점검도 쓰기 전에 끝난다: 항목마다 경로 가드와 목록 digest(lock 은 바이트 sha256 과 프로세스가 아직 죽어 있음)를 다시 보고, 백업 영역이 다른 파일 시스템이면 멈춘다(시험용 고장 지점 `leftovers_cross_device` 는 `simulated` 라고 적은 거절을 낸다).

영수증 `<상태 폴더>/backups/<UTC>/leftovers-receipt.json` 은 0600 으로 `pending` 에서 시작해 첫 이동 전에 읽어 확인하고, 모든 항목이 옮겨져 백업에서 다시 읽힌 뒤에만 `committed` 가 된다. 영수증은 항목마다 경로, 종류(`folder`/`file`), 최상위 항목 수, 크기, 목록 digest 만 담고 파일 내용이나 하위 이름, lock 의 내용은 담지 않는다. 영수증을 저장하지 못하거나 읽어 확인하지 못하면 아무것도 옮기지 않는다. 일부만 옮긴 뒤 멈추면 영수증은 `pending` 으로 남고 오류가 몇 개를 옮겼는지와 되돌리는 명령을 말한다.

결과는 `state`, `receipt`(UTC 이름), `items`(`{path, outcome: moved}`), `tracked_moved`(옮긴 파일 중 git 이 추적하는 수, git 이 답하지 못하면 0), `warning`(0 이면 null, 아니면 그 수와 checkout 이 되살린다는 말), `secret_shaped`(옮긴 폴더 최상위의 비밀 모양 이름 수, 이름은 인쇄하지 않고 열지도 않는다), 되돌리는 명령 `revert` 이다.
