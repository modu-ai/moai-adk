# `language-setup`

operation `language.setup` · 코드 `src/ops/language_setup.rs` · #2293

`language-setup PROJECT [--questions | --answers] [--conversation CODE]` 는 쓰기 언어를 항목마다 물어 `project-plan` 용 완성 프로필을 낸다. 쓰지 않는다. `--questions` 는 호스트 질문 도구용 질문 묶음, `--answers` 는 그 도구의 `{item: code}` JSON 을 stdin 으로 받고, 둘 다 없으면 터미널에서 묻는다(질문은 stderr, 답은 stdin). `src/main.rs` 에서 그대로 옮김(#2834 R4a).
