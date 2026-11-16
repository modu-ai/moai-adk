# `hook-receipt`

operation `hook.receipt` · 코드 `src/ops/hook_receipt.rs`

`hook-receipt --host H` 는 hook 처리기다. 호스트의 hook stdin(앞 4 MiB 만 읽음)에서 허용 목록의 영수증 한 줄을 `./.moai-cli/receipts/` 에 남기고, hook 을 실패시키지 않는다(잘린 객체는 건너뜀). 프로필에 고정 `project_id` 가 있고 HOME 의 상태 폴더가 진짜 폴더이면 줄은 대신 HOME 의 프로젝트 폴더 `projects/<project_id>/receipts/` 에 남는다(#3009 S4b: 같은 프로젝트의 모든 worktree 가 로그 하나를 공유, 폴더 0700·파일 0600, 링크는 거절). 읽는 쪽은 프로젝트의 옛 로그와 HOME 로그를 둘 다 읽고, 옛 로그는 복사·이동·삭제하지 않는다. 로그 한도(4 MiB)는 그대로이고 닿으면 새 줄을 버린다. `src/main.rs` 에서 그대로 옮김(#2834 R2c).
