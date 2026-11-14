# `home-retire-apply`

operation `home_retire.apply` · 코드 `src/ops/home_retire_apply.rs`, `src/home_retire.rs` · Issue #2871 (U2)

`home-retire-apply [--home DIR] --now UTC --plan-digest D --permission-reference REF` 는 같은 입력으로 다시 계산한 미리보기의 digest 가 같고 `action` 이 `retire` 일 때만 움직인다(아니면 아무것도 쓰지 않음). 순서: `~/.moai-go-retired/`(0700)와 `<UTC>/` 만들기 → 매니페스트 `retire-manifest-<UTC>.json`(항목·분류·규칙·병합 후보·옛/새 경로, 남긴 항목과 이유) → 이동 저널 `move-journal-<UTC>.json`(kind `go-retire`)을 먼저 쓰고 항목마다 배타 rename 1회. 적용 때 프로그램이 쥔 항목을 다시 확인해 건너뛴다. 복사도 삭제도 없고, 비밀은 이름만 바꾼다. 적용 뒤 미리보기는 `none`; Go 가 폴더를 다시 만들면 새 stamp 로 또 은퇴한다. 실제 HOME 에서는 사용자 확인 뒤 리더만 실행한다.
