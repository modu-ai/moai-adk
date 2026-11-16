# `session-current`

operation `session.current` · 코드 `src/ops/session_current.rs`, `src/session_current.rs` · Issue #3015 C4 (계획 6892828490, 조사 N10)

`session-current` 는 **이 세션 자신의** 번호를 호스트가 이 프로세스 환경에 넣어 준 변수에서만 읽어 `result` 로 돌려준다. 파일을 읽지도 쓰지도 않고 네트워크도 쓰지 않는다. 인자는 없다.

- 순서: `CLAUDE_CODE_SESSION_ID` → `CODEX_THREAD_ID`. 앞의 것이 번호 모양이면 뒤는 보지 않는다. 번호 모양은 16진수 숫자와 `-` 만, 8~64자(UUID). 앞뒤 공백은 잘라 낸다. 빈 값·공백뿐인 값은 없는 것으로 본다. 번호 모양이 아닌 값은 쓰지 않고 `ignored` 에 **변수 이름만** 적는다(값은 되풀이해 출력하지 않는다).
- 찾았을 때: `{"source_session_id": "<번호>", "source": "env:<변수>", "ignored": [...]}`.
- 못 찾았을 때(종료 코드 0): `source_session_id` 는 `null`, `source` 는 `none`, `fallback` 은 Go 빌드가 내던 줄(`source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>`) 그대로, `hint` 는 “세션 시작 훅이 알려 준 번호를 쓰라”는 안내다.
- **프로젝트 파일(`.moai/state/current-session-id.txt`)은 읽지 않는다.** 그 파일은 프로젝트당 하나라서 세션이 시작될 때마다 덮어쓰여, 레인 안에서 읽으면 리더의 번호가 나온다(Go 빌드의 `session current` 가 그렇게 동작한다).
- `CODEX_THREAD_ID` 는 Codex 문서에 없는 변수다. Codex 레인이 실제 값을 재어 확인하기 전에는 “미확인”으로 본다. 확인은 진짜 HOME 의 인증 파일을 복사하지 않고 한다.
- 이름이 `session-` 으로 시작하는 새 단어라, 사용자 셔심(`~/go/bin/moai`)은 이미 이 명령을 Rust 로 보낸다. Go 의 `moai session current` 는 바뀌지 않는다.
