# `install-inspect`

operation `install.inspect` · 코드 `src/ops/install_inspect.rs` · 04-A #1998

`install-inspect [--json]` 는 PATH 의 세대, `~/.moai-cli/installations/` 의 영수증, Go 세대 경로가 있는지 본다. 쓰지 않는다. 출력은 늘 JSON 봉투이고, stderr 가 터미널이면 PATH 맨 앞 안내를 한 번 보인다(#2867). `src/main.rs` 에서 그대로 옮김(#2834 R3).
