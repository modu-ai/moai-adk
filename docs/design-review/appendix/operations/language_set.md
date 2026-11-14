# `language-set`

operation `language.set` · 코드 `src/ops/language_set.rs` · #2935 L1

`language-set --home CODE` 는 설치할 때 고른 대화 언어(`en`·`ko`·`pt-BR` 처럼 언어 코드 모양이면 무엇이든)를 HOME 상태 폴더의 `config/language.yaml` 에 한 번 적는다. 내용은 `language:` 아래 `conversation_language: CODE` 한 줄뿐이다. 이미 파일이 있으면 덮어쓰지 않고(`written: false`, `reason: exists`), HOME 에 MoAI 상태 폴더가 없으면 아무것도 만들지 않는다(`reason: no_home_state`). 만드는 것은 빠진 `config/` 폴더 하나뿐이다. `config` 나 대상 파일이 링크이거나 종류가 다르면 쓰기 전에 거절한다(exit 1). 코드는 `^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$` 모양에 35자 이하여야 하고, 받은 그대로 적는다(대소문자도 그대로). 이 모양이 아니면 사용법 오류(exit 2)이고 아무것도 쓰지 않는다. `language-setup` 은 저장된 값이 `en`·`ko`·`ja`·`zh` 일 때만 `--conversation`·프로필 `user_chat` 다음, Go 백업·`en` 앞의 대화 언어 기본값으로 쓴다. `fr`·`pt-BR` 처럼 이름을 모르는 코드는 조용히 다음 출처(Go 백업, 그다음 `en`)로 넘어가고, 그 값은 `language::home_conversation` 으로 그대로 읽을 수 있다. 파일이 없거나 깨졌거나 링크이거나 4096바이트를 넘으면 값이 없는 것으로 보고 조용히 `en`. MoAI 가 찍는 글은 어떤 코드여도 영어다.
