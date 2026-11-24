# `home-archive-revert`

operation `home_archive.revert` · 코드 `src/ops/home_archive_revert.rs`, `src/home_state_retire.rs` · SPEC-20261123-home-state-retire

`home-archive-revert [--home DIR] [--stamp UTC] --permission-reference REF` 는 `--stamp` 의 보관 저널(없으면 가장 새 `state-retire` 저널)을 거꾸로 되짚어, 보관 폴더가 옮긴 그 항목(장치:inode 등)이 그대로이고 옛 자리가 비어 있을 때만 옛 상태 폴더 자리로 되돌린다. 옛 폴더가 다시 생겼거나 보관 항목이 바뀌었으면 그 자리에 두고 상태가 `revert_partial` 이며, 두 번째 되돌리기는 아무것도 바꾸지 않는다. 지우지 않는다(`archive` 폴더, 저널, 매니페스트는 남는다). `~/.moai` 가 링크이거나 폴더가 아니면 옮기기 전에 거절한다.
