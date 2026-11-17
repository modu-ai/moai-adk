# `trailer-check`

operation `trailers.check` · 코드 `src/ops/trailer_check.rs`, `src/trailers.rs` · SPEC-20261117-session-trailer(SDD-6)

`trailer-check RANGE [--dir .] [--branch NAME]` 는 RANGE 의 커밋(merge 제외, 오래된 것부터)의 `Moai-Spec:`·`Moai-Session:` trailer 를 git 이 trailer 블록으로 읽는 대로 읽어 `findings`(`missing_spec`, `malformed_spec`, `missing_session`, `malformed_session`)를 낸다. `Moai-Spec` 이 없으면 `spec/` 가지에서만 실패, 값이 둘 이상이거나 모양이 틀리면 늘 실패, `Moai-Session` 이 없거나 worktree key 없는 옛 맨 id(`legacy_session`, 한 번의 이행 기간)면 알리기만 한다. 값은 `<session id>:<12자리 16진수>` 여야 한다. `--branch` 는 분리된 checkout(CI)이 판단할 가지 이름을 준다. `ok` 는 실패가 없을 때 참이고, 거짓이면 결과를 그대로 낸 채 exit 1 로 끝난다(감사 B1). 값이 git log 의 필드 구분 문자를 품어 값을 나눌 수 없으면 `malformed_trailers` 로 늘 실패한다(감사 N3). `-` 로 시작하는 RANGE 는 git 을 부르기 전에 거절한다. git 환경은 지우고 읽기만 한다.
