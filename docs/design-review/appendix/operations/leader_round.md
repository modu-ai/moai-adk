# `leader-round`

operation `leader.round` · 코드 `src/ops/leader_round.rs`, `src/leader_round.rs`, `src/lane_next.rs`(`collect_fleet`) · Issue #2902 O3b

`leader-round PROJECT --observations FILE [--snapshot FILE | --repo OWNER/REPO] [--now UTC]` 는 리더의 한 회차에서 레인마다 할 일 하나와 회차 공통 후보를 만든다. 읽기만 한다.

- 관측 파일(`schema_version` 1, 모르는 키 거부, 레인 50개까지): 레인마다 `lane`·`host`·`read_at`·`source`·`terminal`(Orca `terminal read --screen --json` 그대로)·`pane`·`turn`·`input_box`·`context_used_pct`(0–100)·`wake_armed`·`background_jobs`·`reset`(claims·session_cwd). 턴·입력창·문맥 % 는 관측자가 적고 기계는 화면에서 추측하지 않는다.
- `LANE-*` 줄: 앞뒤 공백과 글머리(`⏺`·`•`) 하나를 뗀 줄이 표지 낱말로 시작할 때만 센다. `>` 인용 줄과 코드 울타리 안은 빼고, 둘째 낱말이 다른 레인이면 `ignored`.
- 판정(위에서 처음 맞는 줄): 0 관측이 600초 넘음·미래·`source≠screen`·`status≠running` → leave · 1 pane unreachable/permission_prompt/usage_limit → leave(권한 질문에 답하지 않음) · 2 턴 running/unknown → leave · 3 입력창 text/unknown → leave · 4 wake 걸림, 기한 남은 `LANE-WAIT` → leave · 5 마지막 표지 `LANE-BLOCKED` → answer(그 줄은 데이터) · 6 카드 경계(DONE, 또는 REPORT + 문맥 85 % 이상, 그 뒤 `moai-claim:` 없음) → `lane-safe-to-reset` 이 safe 면 `clear_redispatch`(`/clear` + `lane-duty-line`, 세션 등록 해제 명령), 아니면 leave + findings · 7 그 밖 → restart(`순환 계속 — self-pick (1)부터`). 끝 조건이 충족되면 `end_condition_met`, 화면 마지막 줄이 이미 그 문장이면 `restart_already_sent` 로 leave.
- `lane-safe-to-reset` 은 레인마다 `lane_reset::check` 를 부른다(git `--no-optional-locks` 읽기, fetch 없음). 화면 관측은 `read_at`·`background_jobs` 로 채운다. 해제할 instance 는 `session-register --show` 와 같은 `lane_registry::show` 로 읽는다(쓰지 않음).
- 회차 공통(`round`): `CLOSE-TABLE` 댓글이 달린 열린 Issue(`close_candidates`), 해제·인계되지 않고 리스가 끝난 claim(`handoff_candidates` — 리스 끝 초는 끝난 것, D2 의 나머지 두 조건은 리더 확인), `moai gate-loop --once …` 명령 한 줄(`gate`, 실행하지 않음), `startable::end_condition`. GitHub 상태가 없으면 넷 모두 `not_run` 과 이유.
- GitHub 읽기: `--snapshot` 은 저장한 `lane-next` 스냅숏, `--repo` 만 있으면 `lane_next::collect_fleet` 가 한 번에 읽는다 — 열린 PR 과 댓글, 열린 Issue 전부와 최신 댓글 100개(50개씩 GraphQL `query`). 요청은 30개까지, 못 읽은 것은 `partial` 에 적는다. REST 쓰기·GraphQL mutation 은 없다.
- 출력: `lanes`(레인마다 `action`·`reason`·`send`·`release`·`data`·`findings`·`notes`·`ignored`), `round`, `partial`, `api_calls`, `writes: false`.
