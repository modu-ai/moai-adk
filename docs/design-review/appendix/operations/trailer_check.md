# `trailer-check`

operation `trailers.check` · 코드 `src/ops/trailer_check.rs`, `src/trailers.rs` · SPEC-20261117-session-trailer(SDD-6)

`trailer-check RANGE [--dir .]` 는 RANGE 의 커밋(merge 제외, 오래된 것부터)의 `Moai-Spec:`·`Moai-Session:` trailer 를 git 이 trailer 블록으로 읽는 대로 읽어 `findings`(`missing_spec`, `malformed_spec`, `missing_session`, `malformed_session`)를 낸다. `Moai-Spec` 이 없으면 `spec/` 가지에서만 실패, 값이 둘 이상이거나 모양이 틀리면 늘 실패, `Moai-Session` 이 없으면 알리기만 한다. `ok` 는 실패가 없을 때 참이다. `-` 로 시작하는 RANGE 는 git 을 부르기 전에 거절한다. git 환경은 지우고 읽기만 한다.
