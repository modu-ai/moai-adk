# `forge-links-probe`

operation `forge.links_probe` · 코드 `src/ops/forge_links_probe.rs` · #2172

`forge-links-probe PROJECT` 는 한 Issue 에 대해 sub-issue·dependency API 가 답하는지 보고, 링크를 API 쓰기 의도 또는 `Parent:`/`Blocked-By:` 본문 줄로 계획한다. GET 만, stdin 의 typed 요청. `src/main.rs` 에서 그대로 옮김(#2834 R2b).
