# `audit-verdict-line`

operation `audit.verdict_line` · 코드 `src/ops/audit_verdict_line.rs`, `src/gate_collect.rs` · Issue #2833 D4

`audit-verdict-line --pr N --verdict PASS|FAIL --title TEXT [--base 40HEX] [--repo OWNER/REPO] [--dir .] [--tip …]` 는 감사 판정 댓글의 첫 줄을 만든다. 출력 직전에 GitHub 에서 PR 의 지금 head 를 읽어(GET 만) `AUDIT #N (TEXT) head <head> base <base> — verdict: <V>` 를 결과 `line` 에 낸다. base 는 `--base` 를 주면 그 값(delta 재감사는 앞 head), 아니면 로컬 저장소에서 head 와 `--tip` 의 merge base 다(head 를 먼저 fetch). 줄은 게이트 G1 이 읽는 `gate::parse_verdict` 로 같은 head·base·판정이 다시 읽힐 때만 낸다(16진수 40자가 아니거나 제목이 여러 줄이면 거부). 감사 도중 head 가 움직여 판정이 옛 head 를 적는 일(#2821)을 막는다. 아무것도 쓰지 않는다.
