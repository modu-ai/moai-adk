# `project-update-leftovers-revert`

operation `project_update.leftovers_revert` · 코드 `src/ops/project_update_leftovers_revert.rs`, `src/project_update/leftovers_archive.rs` · SPEC-20261127-update-leftovers-action

`project-update-leftovers-revert --receipt UTC --permission-reference REF [PROJECT]` 는 `project-update-leftovers-apply` 한 번이 옮긴 항목을 영수증 `<상태 폴더>/backups/<UTC>/leftovers-receipt.json` 으로 제자리에 되돌린다. 권한 참조가 계약에서 거절되거나 영수증 이름이 UTC 이름이 아니거나 영수증을 읽을 수 없거나 다른 프로젝트의 것이면 오류로 끝나고 아무것도 쓰지 않는다.

영수증은 믿지 않는다. 항목의 경로가 알려진 다섯 leftover 경로가 아니거나(절대 경로와 `..` 포함), 영수증이 말하는 종류와 다르거나, 두 번 적혔거나, 링크를 지나면 그 항목은 `conflict` 이고 영수증 전체가 아무것도 옮기지 않는다(`ok: true`, `state: conflict`, 영수증 파일도 그대로). 가드를 통과한 항목은 디스크로 판단한다: 원래 자리가 비어 있고 백업에 있는 사본의 목록 digest(lock 은 sha256)가 영수증과 같을 때만 되돌리고(`restored`), 원래 자리가 차 있거나 사본이 다르면 양쪽을 그대로 두고 `conflict`, 사본도 원래 자리도 없으면 `missing` 이다. 아직 옮기지 않아 원래 자리에 영수증이 적은 그대로 있는 항목은 `restored` 로 읽는다. 영수증의 `moved` 값은 판단에 쓰지 않는다. 덮어쓰지 않고 지우지 않는다.

모든 항목이 `restored` 일 때만 영수증이 `reverted` 가 되고 결과 `state` 도 `reverted` 이다. 그렇지 않으면 결과 `state` 는 `conflict` 이고 영수증은 그대로다. 이미 `reverted` 인 영수증은 `note: already reverted` 로 끝난다. 일부만 되돌린 뒤 멈추면 오류가 몇 개를 되돌렸는지 말하고, 같은 명령을 다시 실행하면 이어진다.

`project-update-plan` 의 `undo_order` 는 `reverted` 가 아닌 leftovers 영수증이 있는 동안만 이 명령을 첫 행(`leftovers`)으로 적는다. Go 백업 폴더에서 옮긴 lock 이 `go-state-revert` 가 그 폴더의 digest 를 읽기 전에 제자리로 돌아와야 하기 때문이다.
