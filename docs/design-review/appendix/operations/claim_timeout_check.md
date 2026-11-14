# `claim-timeout-check`

operation `claim.timeout_check` · 코드 `src/ops/claim_timeout_check.rs`, `src/claim_timeout.rs` · Issue #2906

`claim-timeout-check --issue N --claim HOLDER_ID --request REQUEST_ID --lane ME [--repo OWNER/REPO] [--now UTC]` 는 다른 레인의 claim 을 `moai-claim-handoff: comment=<HOLDER_ID> to=<ME> by=timeout` 로 닫아도 되는지(결정 D2, #1975 6840449133) 세 조건을 읽어 판정한다. GitHub 를 GET 으로만 읽고, 아무것도 쓰지 않는다.

- 리스 만료: `now >= lease_until` (`is_live = lease_until > now` 와 같은 경계).
- 요청 무응답: 요청 댓글이 holder 의 claim id 를 한 낱말로 적고(아니면 `request_names_no_claim`), 이 레인의 요청이며(첫 비지 않은 줄이 `--lane` 을 한 낱말로 적고 holder 레인 댓글이 아님, 아니면 `request_not_mine`), claim 뒤에 썼고(요청 id > claim id, 아니면 `request_before_claim`), `now >= 요청 시각 + 1800` 이며, 요청 뒤에 holder 레인의 댓글도, 그 claim 을 닫는 release·handoff 도 없다.
- 하트비트 없음: `now >= max(최근 활동, claim 작성) + 7200`. 활동 = claim 댓글의 `updated_at`(리스 연장), holder 레인의 댓글(이 Issue 전체 + 저장소 전체 댓글 중 최근 2시간), claim 의 `branch=` 위 `Moai-Claim: <issue>#<claim>` 을 단 커밋(결정 Q4; 가지가 없으면 404 = 커밋 0). 활동 시각이 `now` 보다 뒤면 충족하지 않는다(`future_activity`).
- holder 레인 댓글 = 본문 어느 줄에든 holder 의 lane 이름이나 session 을 한 낱말로 적은 댓글(`lane=<x>`·`session=<x>`·`auditor: <x>`·`감사: <x>`·`LANE-… <x>` 포함; 모든 레인이 한 계정으로 쓰므로 글자 기준). 더 많이 맞으면 인계가 막히는 쪽이다. holder 레인 댓글은 요청으로 받지 않고, 하트비트·답에서 빼지도 않는다.
- 출력: `met`(bool), `state`(`met`·`not_met`·`uncertain`), `reasons`, 조건마다 근거(`conditions`), 읽은 양(`activity_read`). 셋 다 충족하고 `--claim` 이 지금 holder 일 때만 `line` 과 근거 본문 `body`(요청 id·읽은 활동)를 준다.
- 판정이 나오면 충족 여부와 상관없이 exit 0 이다(`met` 로 읽는다). 목록마다 10쪽을 넘거나 읽기가 실패하면 `uncertain`(`met: false`, 줄 없음). 인자 오류(`--now`·`--repo` 형식, 시스템 시계보다 300초 넘게 늦은 `--now` 등)만 오류 envelope 이다.
- `claim-verify` 는 그대로 리스만 본다(결정 Q2). 리더는 같은 명령을 `--now` 로 다시 돌려 확인한다.
