# `hook-receipt`

operation `hook.receipt` · 코드 `src/ops/hook_receipt.rs`

`hook-receipt --host H` 는 hook 처리기다. 호스트의 hook stdin(앞 4 MiB 만 읽음)에서 허용 목록의 영수증 한 줄을 `./.moai-cli/receipts/` 에 남기고, hook 을 실패시키지 않는다(잘린 객체는 건너뜀). `src/main.rs` 에서 그대로 옮김(#2834 R2c).
