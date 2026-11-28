# `improve-note`

operation `improve.note` · 코드 `src/ops/improve_note.rs` · SPEC-20261121-self-improvement-loop (조각 S2, REQ-32~REQ-36, AC-15·AC-16)

`improve-note <project> --kind <correction|failure|false_claim> --key <주제> --text <글> [--by <관찰한 쪽>]` 는 프로젝트 상태 폴더의 `improve/notes.jsonl` 에 **한 줄**을 더하는 명령이다. 이 한 곳이 자기 개선 루프 둘째 조각의 유일한 쓰기이고, `improve-signals`(S1)가 이 줄을 `<kind>:<key>` 부류로 읽는다.

**줄의 모양(SPEC D7).** `{"v":1,"at":"<UTC RFC 3339, 초 단위 …Z>","kind":"…","key":"…","text":"…","by":"…"}` 한 줄과 줄바꿈 하나. `at` 은 이 명령이 시계에서 읽는다. `by` 의 기본값은 `agent` 이다. 폴더와 파일이 없으면 만든다(`improve/` 와 그 위의 상태 폴더 포함).

**거절(`improve-note refused: …`, 종류 `invalid_input`, 아무것도 쓰지 않음).**

- `kind` 가 셋 밖, `key` 가 비었거나 80자 초과, `text` 가 200자 초과, `by` 가 80자 초과(글자 수 기준). 이 글들에 제어 문자(줄바꿈·탭 포함)가 있으면 거절한다. `improve-signals` 가 제어 문자가 든 `key` 를 깨진 줄로 세므로 왕복이 조용히 깨지지 않게 하려는 것이다.
- 이 줄을 더하면 파일이 1 MiB(`improve-signals` 의 읽기 한도)를 넘음. 줄바꿈으로 끝나지 않은 파일에는 줄바꿈 하나를 먼저 쓰고 그 바이트도 센다. 정확히 한도에서 끝나는 파일은 받는다.
- `improve/` 나 `notes.jsonl` 이나 그 위의 상태 폴더가 심볼릭 링크(끊어진 링크 포함)이거나, `notes.jsonl` 이 일반 파일이 아님(이름 붙은 파이프, 폴더). 링크는 따라가지 않고 링크가 가리키는 곳에 아무것도 쓰지 않는다.

**링크를 열지 않는다.** 길을 따라 각 폴더를 링크 검사로 보며 없으면 하나씩 만들고(`mkdir`), 파일은 `O_APPEND|O_NOFOLLOW|O_NONBLOCK` 로 열어 일반 파일인지 `fstat` 로 다시 확인한 뒤 한 번의 `write` 로 쓴다. 검사와 열기 사이에 링크로 바뀌어도 `O_NOFOLLOW` 가 막는다.

**가림.** `key`·`text`·`by` 의 비밀 모양 낱말(`redact::redact_secret_shapes`와 `redact::redact_outgoing`)은 `[redacted]` 로 바뀌어 쓰이고, 원래 글은 파일에도 출력에도 남지 않는다.

**결과.** `action:"note"`, `path`(프로젝트 상대 경로), `kind`, `key`(가려진 값), `appended:1`, `file_bytes`, `writes:true`.

**이 명령은 쓰는 명령이다.** 위험한 경로(링크 폴더·링크 파일·이름 붙은 파이프·링크인 상태 폴더)는 진짜 바이너리로 임시 프로젝트에서 돌리는 Rust 시험(`tests/improve_note_cli.rs`)이 지킨다.
