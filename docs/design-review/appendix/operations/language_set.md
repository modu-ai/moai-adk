# `language-set`

operation `language.set` · 코드 `src/ops/language_set.rs` · #2935 L1

`language-set --home CODE` 는 설치할 때 고른 대화 언어(`en`·`ko`·`ja`·`zh`)를 HOME 상태 폴더의 `config/language.yaml` 에 한 번 적는다. 내용은 `language:` 아래 `conversation_language: CODE` 한 줄뿐이다. 이미 파일이 있으면 덮어쓰지 않고(`written: false`, `reason: exists`), HOME 에 MoAI 상태 폴더가 없으면 아무것도 만들지 않는다(`reason: no_home_state`). 만드는 것은 빠진 `config/` 폴더 하나뿐이다. `config` 나 대상 파일이 링크이거나 종류가 다르면 쓰기 전에 거절한다(exit 1). 코드가 허용 목록 밖이면 사용법 오류(exit 2)이고 아무것도 쓰지 않는다. `language-setup` 은 이 값을 `--conversation`·프로필 `user_chat` 다음, Go 백업·`en` 앞의 대화 언어 기본값으로 읽는다. 파일이 없거나 깨졌거나 링크이거나 4096바이트를 넘으면 조용히 `en`. MoAI 가 찍는 글은 어떤 코드여도 영어다.
