---
id: SPEC-20261126-profile-command-values
title: "프로필의 결정 값에는 명령의 '/' 를 받고, 거절할 때는 필드 이름을 대요"
status: draft
revision: 2
created: 2026-11-26
---

# 프로필의 결정 값에는 명령의 `/` 를 받고, 거절할 때는 필드 이름을 대요

## Interview

- Q1: 무엇이 문제인가요? → A: 카드 본문(2026-11-19 사용자 결정을 옮긴 부모 카드에서 나온 조각)이에요. `project-plan` 이 `/` 가 든 결정 값을 모두 거절해서 `go test ./...` 나 `./gradlew build` 를 저장할 수 없고, 거절 문장이 필드를 알려 주지 않아요.
- Q2: 어디까지 지켜야 하나요? → A: 카드 기준 둘이에요. 두 명령이 든 프로필이 plan 과 apply 를 통과하고, 절대 홈 경로와 비밀 모양 값은 계속 거절돼요. 거절 오류는 필드 id 를 대고 값을 되비추지 않아요.
- Q3: 개정 1 의 감사에서 "/" 뒤의 비밀 모양 값이 통과하는 것이 드러났어요. 어떻게 하나요? → A: 사용자가 직접 "먼저 고친 뒤 승인" 을 골랐어요(카드 노트 10227). 이 답은 수리 방향이고, 이 개정의 승인은 따로 받아요.
- Q4: 전에는 `/` 때문에 거절되던 값은 어디까지 다시 거절하나요? → A: 리더 결정(카드 노트 10327)이에요. 홈 변수 이름과 `@` 앞 암호 꼴까지 더하고, 정당한 명령을 막게 되는 나머지는 알려진 한계로 남겨 승인할 때 보여요.

## Intent

프로필의 글 값 검사(`src/profile.rs` 의 `portable`)는 `/` 가 든 글을 모두 거절해요. 경로와 홈 경로가 프로필에 남지 않게 하려는 검사인데, 인터뷰가 결정(`decisions`)의 값으로 저장하는 시험·빌드 명령까지 걸려요. 이 변경은 결정의 값에만 명령을 받는 검사를 따로 두고, 다른 모든 글에는 지금 검사를 그대로 써요. 거절 문장은 걸린 필드의 id 를 대요.

읽는 사람: 이 변경을 판단하는 사용자와 리더, 구현하는 레인, 독립 감사자.

## Non-goals

- `purpose`, `tools`, `development_methods`, `interview_revision`, `versioning.format`, 결정의 `topic`·`scope` 의 검사는 바꾸지 않아요.
- 비밀 모양 목록과 게이트는 느슨해지지 않아요.
- 이미 필드 이름을 대는 하위 검사(`language`, `scopes`, `detected`, `work_flow`, `lane_run`, `ci`, `checks`, `test_paths`, `leader`, `versioning`)의 문장은 바꾸지 않아요.
- 디스크의 프로필을 읽다 나는 형식 오류와 `project-apply` 의 입력 형식 오류의 문장은 바꾸지 않아요.
- 결정 값의 검사는 일반 비밀 탐지기가 아니에요. R-2 의 남는 값은 이 변경에서 막지 않아요.

## Requirements

- REQ-1: The profile screen shall accept a decision value whose every "/" follows a name character, which is an ASCII letter, an ASCII digit or one of ". _ - * /", when no rule of REQ-2, REQ-3, REQ-8 or REQ-9 refuses the value.
- REQ-2: The profile screen shall refuse a decision value that is blank, is over 256 bytes, holds a control character or a backslash, holds a tilde that a "/" directly follows, matches the secret screen, or matches the outgoing redaction screen as a whole or in any piece between "/" and "=".
- REQ-3: The profile screen shall refuse a decision value in which a "~" at the start or after a character that is no name character leads to a "/" through user-name characters, or in which a flag that starts with "-" there is directly followed by a "/" or by such a "~" path.
- REQ-4: The profile screen shall keep the earlier screen for every profile text other than a decision value.
- REQ-5: When the profile screen refuses a profile, the plan shall answer "invalid or nonportable profile field <id>; no values were logged" with the kind invalid_input, naming a decision by its id when that id is a portable id that no secret screen matches and by its position otherwise. · failure: none — a refused profile writes nothing.
- REQ-6: When the project-plan command cannot read its input as a profile, the command shall name in the sentence of REQ-5 the first known key that is missing or has the wrong type, and keep the earlier sentence without a field when the input holds an unknown key or no known key is the cause. · failure: none — the input is refused either way and nothing is written.
- REQ-7: The refusal of a profile shall carry no value read from the input.
- REQ-8: The profile screen shall refuse a decision value in which a word ended by ASCII white space holds "$" or "${" followed in any letter case by HOME, USERPROFILE, HOMEDRIVE or HOMEPATH, no ASCII letter, ASCII digit or "_" continues that name, and a "/" comes later in the same word.
- REQ-9: The profile screen shall refuse a decision value in which a word ended by ASCII white space has an "@" whose preceding text, counted from the word start or the last "/" before it, holds a ":" with at least one character on each side, and a "/" comes after that "@" in the same word.

## Acceptance

- AC-1 (REQ-1, REQ-8, REQ-9): a profile whose test command is `go test ./...` and whose build command is `./gradlew build` plans, applies, reads back with both values and plans again as unchanged; the neighbour values of REQ-8 and REQ-9 and the known limits of R-2 are taken · verify: `cargo test --locked --offline --test profile_command_values accept_`
- AC-2 (REQ-2, REQ-3, REQ-7, REQ-8, REQ-9): each refused value of a table (absolute, home, tilde, home-variable, flag-glued, backslash, control, over-long, password-before-"@" and secret-shaped forms, bare and behind a path or a flag) is refused with its field id and the kind invalid_input, appears in neither output stream and leaves the project folder untouched · verify: `cargo test --locked --offline --test profile_command_values refused_`
- AC-3 (REQ-4): a "/" in each other profile text is still refused · verify: `cargo test --locked --offline --test profile_command_values other_fields_`
- AC-4 (REQ-5, REQ-7): the refusal names the field for a plain field, for a decision by id, and for a decision by position when the id is unusable, and a decision id such as `temporary` keeps the kind invalid_input · verify: `cargo test --locked --offline --test profile_command_values field_`
- AC-5 (REQ-6, REQ-7): a missing or wrong-typed known key is named without echoing its value, and an unknown key gets the earlier sentence · verify: `cargo test --locked --offline --test profile_command_values shape_`
- AC-6 (REQ-4, REQ-5): the profile and classification unit tests pass · verify: `cargo test --locked --offline --lib profile::` and `cargo test --locked --offline --lib contract::`
- AC-7 (REQ-4, REQ-6): the earlier refusal test and the protocol fixtures of the product commands pass · verify: `cargo test --locked --offline --test product_cli nonportable_or_secret_shaped` and `cargo test --locked --offline --test product_cli protocol_fixtures_record_current_contract`

## Scope

- write: `.moai/specs/SPEC-20261126-profile-command-values/**`, `src/contract.rs`, `src/ops/project_plan.rs`, `tests/profile_command_values.rs`, `docs/design-review/appendix/operations/project_plan.md`, `docs/rust-product-path.md`
- write (line-lease): `src/profile.rs`
- read-only: `src/generation.rs`, `src/agents.rs`, `src/project_detect.rs`, `src/redact.rs`, `tests/product_cli.rs`
- forbidden: 게이트 완화, 비밀 모양 목록 축소, 실제 HOME, 네트워크, Python

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 명령을 받는 검사는 결정의 값 하나에만 써요 | 시험·빌드 명령은 결정의 값으로 저장돼요. 다른 글은 경로 성격이라 그대로예요 | 검사를 지우고 `portable` 로 되돌려요 |
| D-2 | `/` 앞의 글자는 ASCII 만 이름 글자예요 | 보이지 않는 유니코드 글자(U+3164)가 이름 글자로 통과하면 홈 경로가 들어가요 | 범위를 넓혀요 |
| D-3 | `)` `]` `}` 는 이름 글자가 아니에요 | `$(echo ~)/.ssh`, `${HOME:-}/.ssh` 가 통과해요 | 뒤 글자 검사를 더해 허용해요 |
| D-4 | 결정 값에는 밖으로 내보내는 글의 가림 검사(`redact_outgoing`)를 값 전체와 `/`·`=` 에서 나눈 각 조각에 돌려요 | 가림 검사는 낱말 머리의 키 모양만 봐서 `x/<모양>` 이 통과했어요(감사 노트 10199 지적 1, 사용자 선택 10227) | 조각 검사 줄을 빼요 |
| D-5 | 상대 경로의 `..` 와 홈 변수가 아닌 환경 변수 꼴은 받아요. 홈 변수(`HOME`·`USERPROFILE`·`HOMEDRIVE`·`HOMEPATH`)는 같은 낱말에 `/` 가 뒤따르면 거절해요 | 홈 경로를 만드는 이름만 좁게 막아요(노트 10227, 10327). 목록을 늘려도 다 막지는 못해요 | 이름 목록을 고쳐요 |
| D-6 | 필드를 대는 곳은 plan 의 검사와 `project-plan` 의 입력 형식이에요 | 카드가 말하는 문장이 여기서 나와요. 모르는 키는 사용자가 쓴 글이라 대지 않아요 | 다른 파서에도 더해요 |
| D-7 | `@` 앞 암호 꼴은 그 `@` 뒤의 같은 낱말에 `/` 가 있을 때만 거절해요 | 노트 10327. `/` 가 없는 값은 이 카드 전에도 받았고, 이렇게 좁혀야 다이제스트로 고정한 이미지 이름(`이름:태그@sha256:…`)과 `묶음:이름:판@확장자` 표기가 막히지 않아요. 노트 10323 에 "같이 거절되는 정당한 명령 없음" 이라고 쓴 것은 틀렸어요(R-1) | 조건을 빼거나 넓혀요 |
| D-8 | R-2 의 남는 값에는 규칙을 더하지 않아요 | 노트 10327: 규칙을 더하면 `go test github.com/…`, `docker run -u 1000:1000`, `Features/Users/…` 같은 명령이 막혀요 | 값마다 규칙을 더하는 개정을 해요 |

## Plan

1. RED: `tests/profile_command_values.rs` 에 `accept_`, `refused_`, `other_fields_`, `field_`, `shape_` 시험을 먼저 써서 실패를 기록해요.
2. GREEN: `src/contract.rs` 의 분류 줄, `src/profile.rs` 의 값 검사·필드별 거절·입력 형식 거절, `src/ops/project_plan.rs` 의 호출.
3. 운영 문서, `cargo fmt --all`, `check affected --card` 한 번, READY.
4. 감사 수리(개정 2): 새 거절 시험을 먼저 써서 실패를 보고, 조각 검사·홈 변수 규칙·`@` 규칙을 더해요. 지키는 시험마다 변이를 한 줄씩 남기고, 두 문서를 맞춘 뒤 `check affected --card` 한 번, READY.

## Risks

- R-1: 정당한데 계속 거절되는 값이 있어요: `2>/dev/null`, `scheme://` 주소, Bazel 의 `//...`, 산문의 `a / b`, `/p:` 꼴 플래그, `$PWD:/app`, `"$PWD"/x`, `app/[slug]/page`, `(auth)/login`, `{a,b}/x`, `-Iinclude/x`, 한글 낱말 바로 뒤의 `/`, 모든 백슬래시. 개정 2 의 규칙 때문에 거절되는 값: `PATH=$HOME:./bin go test`(홈 변수와 `/` 가 한 낱말), `npm:cowsay@1.5.0/cli` 꼴(`@` 앞에 `:` 가 있고 뒤에 `/`).
- R-2: 전에는 `/` 때문에 거절됐지만 이제 받는 값이 있어요. 감사(노트 10199 지적 2)가 든 열한 값 가운데 개정 2 가 다시 거절하는 것은 넷이에요: `$USERPROFILE/x`, `$HOMEDRIVE$HOMEPATH/x`, `user:pw@host/db`, `psql 이름:암호@호스트/db`. **계속 받는 것(알려진 한계)** 은 일곱이에요: `../../Users/name`, `x/Users/name`, `./../../../Users/name`, scheme 없는 `host.invalid/path`, scheme 없는 webhook 경로, `curl -u 이름:암호 호스트/경로`, AWS 문서 예시 모양의 `/` 둘 든 비밀 키. 일곱은 같은 모양의 값을 시험 `accept_the_known_limits_of_the_spec` 이 고정해요(감사가 쓴 글자 그대로는 아니에요). `/`·`=` 가 아닌 구분자 뒤의 키 모양(`x:<모양>`, 같은 시험)과 두 비밀 검사에 걸리지 않는 다른 비밀도 받아요. 감사 지적 3 의 경계 값(`~jöhn/x`, `~$USER/x`, `./~name/x`, `*/Users/name`, `-I_/Users/name`)은 바꾸지 않았고 이번에 다시 재지 않았어요. 자격 증명이 든 값은 `/` 금지가 우연히 막아 주던 것이라, 검사를 통과했다고 비밀이 없는 것은 아니에요.
- R-3: 이 변경 전의 `moai` 는 `/` 값이 든 프로필을 읽을 때마다 거절해요. 설치된 바이너리를 올린 뒤에 그런 값을 저장해야 해요.
- R-4: `docs/rust-product-path.md` 의 문장은 이 개정에서 고쳐요. 그 파일은 다른 카드 둘의 쓰기 범위와 겹쳐, 착지 순서에 따라 병합할 때 손이 필요할 수 있어요.
- R-5: 이 변경 전의 `moai` 로 저장한 프로필에 키 모양의 결정 값이 맨 꼴이나 `=` 뒤(`--key=<모양>`)에 있으면, 전에는 읽혔지만 이제 읽기가 거절돼요. 읽을 때도 같은 검사를 돌려요(코드 읽기, 재지 않았어요).

## Result

2026-11-26 에 macOS 에서 돌렸어요. 기준은 NeXT `b3048a821a` 예요. 아래 AC 줄은 개정 2 의 코드 커밋 `35e40621f1` 에서 확인 명령을 적힌 그대로 차례로 돌린 결과예요(모두 exit 0, 기록 `target/dana/repair/v-ac*.log`).

- AC-1: met — `accept_` 6개 통과.
- AC-2: met — `refused_` 7개 통과.
- AC-3: met — `other_fields_` 1개 통과.
- AC-4: met — `field_` 5개 통과.
- AC-5: met — `shape_` 3개 통과(이름에 `shape_` 가 든 거절 시험 하나가 함께 걸려요).
- AC-6: met — `profile::` 23개, `contract::` 7개 통과.
- AC-7: met — 두 줄 각각 1개 통과.

개정 2(감사 수리):

- RED: 커밋 `2f8bc8236b` 의 소스에 새 시험을 더해 돌리니 21개 가운데 새 거절 시험 4개가 실패하고 17개가 통과했어요(exit 100, `target/dana/repair/red.log`). 통과한 새 시험은 바뀌지 않는 동작을 고정하거나, 이미 있던 비밀 검사를 지키는 시험이에요.
- 변이: 이 카드가 더한 거절 시험마다 변이를 하나씩 돌려 모두 실패했어요(exit 100). 줄은 병합 준비 노트에 적어요.
- 감사의 30값(키 모양 다섯 × 앞붙임 여섯)은 `refused_secret_shapes_behind_a_path_or_a_flag_are_named_and_write_nothing` 이 모두 거절해요.
- 영향 범위 검사(`check affected --card`)의 줄은 병합 준비 노트에 있어요. Windows 와 Linux 는 돌리지 않았어요.

개정 1(커밋 `2f8bc8236b`): RED 13개 가운데 11개 실패(exit 101), 구현 뒤 확인 명령 전부 통과. 필드는 하위 검사(`versioning`, `shared_files` 들)가 먼저 돌고 그다음 프로필 키 표의 순서로 첫 필드를 대요. 결정 안에서는 `id`, `topic`, `value`, `source`, `scope` 순서이고 위치는 0 부터 세요. 비밀 모양의 결정 id 는 다른 검사에 걸리지 않으면 전처럼 받고, 거절 문장에는 보이지 않아요. `project_id`, `hosts`, `package_pin`, 결정의 `id`·`source` 의 `/` 거절은 각 필드의 형식 검사가 이미 하던 일이에요.
