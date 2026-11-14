# `hooks-revert`

operation `hooks.revert` · 코드 `src/ops/hooks_revert.rs` · #2853

`hooks-revert PROJECT --receipt ID --permission-reference R` 는 확인된 `update` 한 번이 적용한 hook 층을 그 영수증으로 되돌린다. 호스트의 파일이 모두 update 가 쓴 그대로일 때만 앞 바이트로 돌리고, 그 뒤 바뀐 호스트는 건너뛰고 이름을 댄다. `src/main.rs` 에서 그대로 옮김(#2834 R4b).
