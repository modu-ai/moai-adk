# `home-state-apply`

operation `home_state.apply` · 코드 `src/ops/home_state_apply.rs`, `src/home_state.rs` · Issue #2723

`home-state-apply [--home DIR] --now UTC --plan-digest D --permission-reference REF [--copy NAME]…` 는 같은 입력으로 다시 계산한 미리보기의 digest 가 같고 `action` 이 `migrate` 일 때만 움직인다(아니면 아무것도 쓰지 않음). 순서: 고른 항목 확인(후보이고 `kind: entry` — 비밀·Go 표식(`user-assets.json`, `state`)·도구가 쓰는 이름(`kind: reserved`)은 거부, 링크·특수 파일이 들어 있으면 거부, 10만 항목 상한) → 백업 이름이 아직 비었는지 다시 확인 → `~/.moai` 를 백업으로 rename 1회 → Rust `~/.moai` 생성 → 고른 항목만 복사 → 영수증 `home-state-receipt.json` → 표식 `moai-cli.json` 을 마지막에 원자적으로. rename 뒤 실패하면 지우지 않고 부분 폴더를 `~/.moai.partial-<UTC>` 로 옮기고 Go 폴더를 되돌린다. 다시 미리보기하면 판정 `rust`, `action: none` 이다. 실제 HOME 에서는 사용자 확인 뒤 리더만 실행한다.
