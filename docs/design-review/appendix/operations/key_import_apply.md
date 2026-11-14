# `key-import-apply`

operation `key.import_apply` · 코드 `src/ops/key_import_apply.rs` · #2393 S3

`key-import-apply PROVIDER --plan-digest D --permission-reference R [--shared]` 는 미리보기 digest 와 권한 참조가 있을 때만 옛 키 파일의 키를 금고로 복사한다. 다시 읽은 값이 같아야 하고, 옛 파일은 남긴다. `src/main.rs` 에서 그대로 옮김(#2834 R3).
