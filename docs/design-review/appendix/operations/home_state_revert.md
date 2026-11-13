# `home-state-revert`

operation `home_state.revert` · 코드 `src/ops/home_state_revert.rs`, `src/home_state.rs` · Issue #2723

`home-state-revert [--home DIR] [--now UTC] --permission-reference REF` 는 apply 가 만든 Rust `~/.moai`(판정 `rust`, 읽을 수 있는 `home-state-receipt.json`)를 `~/.moai.rust-<UTC>` 로 옮기고, 영수증이 가리키는 백업(`~/.moai` 옆의 `.moai-backup*` 폴더)을 `~/.moai` 로 되돌린다. 지우지 않는다. 옮길 이름이 이미 있거나 백업이 폴더가 아니면 아무것도 옮기지 않는다.
