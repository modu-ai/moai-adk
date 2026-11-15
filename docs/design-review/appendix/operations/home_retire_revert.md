# `home-retire-revert`

operation `home_retire.revert` · 코드 `src/ops/home_retire_revert.rs`, `src/home_retire.rs` · Issue #2871 (U2)

`home-retire-revert [--home DIR] [--stamp UTC] --permission-reference REF` 는 `--stamp` 의 은퇴 저널(없으면 가장 새 `go-retire` 저널)을 거꾸로 되짚어, 옮긴 항목마다 정체(장치:inode 등)가 그대로이고 옛 자리가 비어 있을 때만 `~/.moai` 로 되돌린다. 그 사이 바뀐 항목은 그 자리에 두고 상태가 `revert_partial` 이며, 두 번째 되돌리기는 아무것도 바꾸지 않는다. 지우지 않는다. `~/.moai` 가 링크이거나 폴더가 아니면 어떤 항목도 옮기기 전에 거절한다(링크를 따라가면 비밀을 담은 항목이 링크가 가리키는 곳으로 가고 저널은 성공으로 끝난다); 없는 것은 되돌리기가 그대로 다룬다.
