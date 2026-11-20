# `improve-signals`

operation `improve.signals` · 코드 `src/ops/improve_signals.rs` · SPEC-20261121-self-improvement-loop (조각 S1)

`improve-signals <project> [--home DIR]` 는 프로젝트에서 **관찰된 문제**를 모아 "부류(class)"로 묶고 횟수를 세어 보여 주는 읽기 전용 명령이다. 아무것도 쓰지 않고(`writes:false`), 제안도 하지 않는다. 자기 개선 루프(`moai improve`)의 첫 단계로, 이어지는 조각(제안·평가·적용)이 이 결과를 읽는다.

**읽는 곳은 넷뿐이다.**

- 카드 장부의 감사 FAIL 노트: 첫 줄이 `AUDIT ` 로 시작하고 `verdict: FAIL` 을 담은 노트(카드 노트 모양 `AUDIT C-<n> … — verdict: FAIL`, PR 댓글 모양 `AUDIT #<n> …` 도 같다). 본문에서 `src/a.rs` 같은 상대 경로(확장자 rs·md·py·json·yml·yaml·toml·sh·txt)마다 부류 `audit_fail:<경로>` 한 신호이고(뒤의 `:줄` 은 뗀다, 한 노트에서 같은 경로는 한 번), 경로가 없으면 `audit_fail:uncited` 이다.
- 카드 장부의 `LANE-BLOCKED` 노트: 첫 줄의 `kind=<단어>` 로 `block:<kind>`(없으면 `block:unknown`).
- 제목이 `flaky: <test>` 인 카드: 부류 `flake:<test>`. 횟수는 그 카드 1개 + 그 카드에 달린 `observation`·`report` 노트 수다(재발은 노트로 남는다는 병합 열차의 관행에 따른 근사값, SPEC D6).
- 프로젝트 상태 폴더의 `improve/notes.jsonl`: 줄마다 `{"v":1,"at":"<UTC 시각 …Z>","kind":"correction|failure|false_claim","key":"<80자 이하>","text":"<짧은 글>"}` 한 개이고 부류는 `<kind>:<key>` 이다. 깨진 줄·모르는 `kind`·잘못된 시각은 멈추지 않고 `skipped.bad_line` 에 센다. 이 파일을 쓰는 명령은 뒤 조각이며, 지금은 사람이나 다른 도구가 만든다.

**신호가 아닌 것.** 에이전트 자신의 주장(`LANE-READY`·`LANE-DONE`·`CANDIDATE`·`MERGE-READY`·`MERGED-LOCAL` 노트, 커밋 메시지, "고쳤다" 보고)과 위 규칙에 맞지 않는 모든 노트는 세지 않고 `skipped.no_rule` 에 센다. 에이전트의 주장이 틀렸다고 밝혀진 경우만 `false_claim` 종류로 `notes.jsonl` 에 들어올 수 있다.

**부류와 두 번 규칙.** 결과의 `classes` 는 많이 본 것부터이고 각 부류는 `key`, `kind`, `count`, `status`, `first_seen`, `last_seen`(Unix 초), 예시 `signals`(최근 5개, 각 `ref`·`at`·`text`)를 가진다. 한 번 본 부류는 `observed_once`, 두 번 이상이면 `candidate` 이다(고정 2, 경계 시험이 1 과 2 를 가른다).

**읽는 양과 가림.** 장부는 최근 노트 20,000 개, 한 노트는 앞 8,192 바이트만 읽고 더 있으면 `truncated:true` 이다. `notes.jsonl` 은 1 MiB 까지 읽는다. 밖으로 나가는 글은 첫 줄만 비밀 모양을 가린 뒤 120자로 자르고(`redact`), 부류 키에도 비밀 모양은 `[redacted]` 로 바뀐다.

**쓰지 않고, 만들지 않는다.** 파일·폴더·장부 행을 쓰지 않는다. 장부나 상태 폴더가 없어도 만들지 않고 `ledger:"absent"`·`notes_file:"absent"` 로 알린 뒤 종료 코드 0 이다. `notes.jsonl` 자리에 심볼릭 링크나 일반 파일이 아닌 것(이름 붙은 파이프)이 있으면 열기 전에 거절하고(`invalid_input`), 읽을 수 없는 장부는 장부의 오류(`blocked`)를 그대로 낸다.

**호스트 메모리는 읽지 않는다.** Claude 자동 메모리와 Codex memories 를 비롯해 실제 HOME 아래에서 이 프로젝트의 장부 외에는 열지 않으며 결과에 `host_memory:"off"` 가 있다. 읽기는 프로젝트별 옵트인이 있어야 하고 뒤 조각에서 다룬다(SPEC D3).

점검 실패의 종료 코드는 저장된 곳이 없어 이 조각은 읽지 않는다(SPEC D8). 같은 정보는 뒤 조각의 `failure` 종류 노트로 들어온다.
