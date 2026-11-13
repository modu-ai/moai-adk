# `wait`

operation `wait` · 코드 `src/ops/wait.rs`, `src/forge/wait.rs` · Issue #2676

`wait KIND --repo OWNER/REPO [--timeout 1800] [--interval 30]` 는 GitHub 를 조건부 요청으로 묻다가 조건이 참이 되면 끝나는 한 번 실행이다. `ci --head H`(그 커밋의 workflow run 이 모두 `completed`, 결론 `success`/`failure`), `comment --issue N [--after ID] [--prefix TEXT]`(id 가 더 큰 댓글, 접두어가 있으면 그것으로 시작하는 줄이 있는 댓글), `pr --pr N --head H`(head 가 바뀌거나 닫힘·병합). 요청은 하나씩, 앞 응답의 ETag 로 다시 물어 바뀌지 않으면 304(인증된 304 는 1차 한도를 쓰지 않는다). 간격은 `--interval`(10초 이상)과 `x-poll-interval` 중 큰 값. 결과: `matched`, `condition`, `value`, `polls`, `not_modified`, `rate_limit`(`limit`·`remaining`·`reset`·`resource`), `elapsed_secs`. `--timeout`(3600초 이하)이 지나면 `matched: false` 로 끝난다(오류 아님). 인증 실패·한도 초과는 `blocked`.
