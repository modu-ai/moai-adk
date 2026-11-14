# `key-set`

operation `key.set` · 코드 `src/ops/key_set.rs` · #2393

`key-set PROVIDER [--shared]` 는 공급자 키를 암호화해 금고에 넣는다(지금 프로젝트, 또는 `--shared` 로 모든 프로젝트). 값은 stdin 으로만 받고(argv 아님), 대화형 터미널이면 거절한다. 답에는 값이 없다. `src/main.rs` 에서 그대로 옮김(#2834 R3).
