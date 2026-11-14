# `key-remove`

operation `key.remove` · 코드 `src/ops/key_remove.rs` · #2393 · #2806

`key-remove PROVIDER [--yes] [--shared]` 는 금고에서 공급자 키를 지운다. `--yes` 없이는 무엇을 지울지만 알리고, stdin 과 stderr 가 모두 터미널이며 지울 항목이 있으면 한 번 묻는다. 기계 키와 옛 키 파일은 건드리지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R3).
