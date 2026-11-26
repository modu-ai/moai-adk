---
id: SPEC-20261126-profile-command-values
title: "프로필의 결정 값에는 명령의 '/' 를 받고, 거절할 때는 필드 이름을 대요"
status: draft
revision: 1
created: 2026-11-26
---

# 프로필의 결정 값에는 명령의 `/` 를 받고, 거절할 때는 필드 이름을 대요

## Interview

- Q1: 무엇이 문제인가요? → A: 카드 본문(2026-11-19 사용자 결정을 옮긴 부모 카드에서 나온 조각)이에요. `project-plan` 이 `/` 가 든 결정 값을 모두 거절해서 `go test ./...` 나 `./gradlew build` 를 저장할 수 없고, 거절 문장이 필드를 알려 주지 않아요.
- Q2: 어디까지 지켜야 하나요? → A: 카드 기준 둘이에요. 두 명령이 든 프로필이 plan 과 apply 를 통과하고, 절대 홈 경로와 비밀 모양 값은 계속 거절돼요. 거절 오류는 필드 id 를 대고 값을 되비추지 않아요.

## Intent

프로필의 글 값 검사(`src/profile.rs` 의 `portable`)는 `/` 가 든 글을 모두 거절해요. 경로와 홈 경로가 프로필에 남지 않게 하려는 검사인데, 인터뷰가 결정(`decisions`)의 값으로 저장하는 시험·빌드 명령까지 걸려요. 이 변경은 결정의 값에만 명령을 받는 검사를 따로 두고, 다른 모든 글에는 지금 검사를 그대로 써요. 거절 문장은 걸린 필드의 id 를 대요.

읽는 사람: 이 변경을 판단하는 리더, 구현하는 레인, 독립 감사자.

## Non-goals

- `purpose`, `tools`, `development_methods`, `interview_revision`, `versioning.format`, 결정의 `topic`·`scope` 의 검사는 바꾸지 않아요.
- 비밀 모양 목록과 게이트는 느슨해지지 않아요.
- 이미 필드 이름을 대는 하위 검사(`language`, `scopes`, `detected`, `work_flow`, `lane_run`, `ci`, `checks`, `test_paths`, `leader`, `versioning`)의 문장은 바꾸지 않아요.
- 디스크의 프로필을 읽다 나는 형식 오류와 `project-apply` 의 입력 형식 오류의 문장은 바꾸지 않아요.

## Requirements

- REQ-1: The profile screen shall accept a decision value whose every "/" follows a name character, which is an ASCII letter, an ASCII digit or one of ". _ - * /", when no rule of REQ-2 or REQ-3 refuses the value.
- REQ-2: The profile screen shall refuse a decision value that is blank, is over 256 bytes, holds a control character or a backslash, matches the secret screen or the outgoing redaction screen, or holds a home path written with a tilde or with the HOME variable in any letter case.
- REQ-3: The profile screen shall refuse a decision value in which a "~" at the start or after a character that is no name character leads to a "/" through user-name characters, or in which a flag that starts with "-" there is directly followed by a "/" or by such a "~" path.
- REQ-4: The profile screen shall keep the earlier screen for every profile text other than a decision value.
- REQ-5: When the profile screen refuses a profile, the plan shall answer "invalid or nonportable profile field <id>; no values were logged" with the kind invalid_input, naming a decision by its id when that id is a portable id that no secret screen matches and by its position otherwise. · failure: none — a refused profile writes nothing.
- REQ-6: When the project-plan command cannot read its input as a profile, the command shall name in the sentence of REQ-5 the first known key that is missing or has the wrong type, and keep the earlier sentence without a field when the input holds an unknown key or no known key is the cause. · failure: none — the input is refused either way and nothing is written.
- REQ-7: The refusal of a profile shall carry no value read from the input.

## Acceptance

- AC-1 (REQ-1): a profile whose test command is `go test ./...` and whose build command is `./gradlew build` plans, applies, reads back with both values and plans again as unchanged · verify: `cargo test --locked --offline --test profile_command_values accept_`
- AC-2 (REQ-2, REQ-3, REQ-7): each refused value of a table (absolute, home, tilde, flag-glued, backslash, control, over-long and secret-shaped forms) is refused with its field id and the kind invalid_input, appears in neither output stream and leaves the project folder untouched · verify: `cargo test --locked --offline --test profile_command_values refused_`
- AC-3 (REQ-4): a "/" in each other profile text is still refused · verify: `cargo test --locked --offline --test profile_command_values other_fields_`
- AC-4 (REQ-5, REQ-7): the refusal names the field for a plain field, for a decision by id, and for a decision by position when the id is unusable, and a decision id such as `temporary` keeps the kind invalid_input · verify: `cargo test --locked --offline --test profile_command_values field_`
- AC-5 (REQ-6, REQ-7): a missing or wrong-typed known key is named without echoing its value, and an unknown key gets the earlier sentence · verify: `cargo test --locked --offline --test profile_command_values shape_`
- AC-6 (REQ-4, REQ-5): the profile and classification unit tests pass · verify: `cargo test --locked --offline --lib profile::` and `cargo test --locked --offline --lib contract::`
- AC-7 (REQ-4, REQ-6): the earlier refusal test and the protocol fixtures of the product commands pass · verify: `cargo test --locked --offline --test product_cli nonportable_or_secret_shaped` and `cargo test --locked --offline --test product_cli protocol_fixtures_record_current_contract`

## Scope

- write: `.moai/specs/SPEC-20261126-profile-command-values/**`, `src/contract.rs`, `src/ops/project_plan.rs`, `tests/profile_command_values.rs`, `docs/design-review/appendix/operations/project_plan.md`
- write (line-lease): `src/profile.rs`
- read-only: `src/generation.rs`, `src/agents.rs`, `src/project_detect.rs`, `tests/product_cli.rs`
- forbidden: 게이트 완화, 비밀 모양 목록 축소, 실제 HOME, 네트워크, Python

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 명령을 받는 검사는 결정의 값 하나에만 써요 | 시험·빌드 명령은 결정의 값으로 저장돼요. 다른 글은 경로 성격이라 그대로예요 | 검사를 지우고 `portable` 로 되돌려요 |
| D-2 | `/` 앞의 글자는 ASCII 만 이름 글자예요 | 보이지 않는 유니코드 글자(U+3164)가 이름 글자로 통과하면 홈 경로가 들어가요 | 범위를 넓혀요 |
| D-3 | `)` `]` `}` 는 이름 글자가 아니에요 | `$(echo ~)/.ssh`, `${HOME:-}/.ssh` 가 통과해요 | 뒤 글자 검사를 더해 허용해요 |
| D-4 | 결정 값에는 밖으로 내보내는 글의 가림 검사(`redact_outgoing`)도 돌려요 | 기존 코드를 부르고 거절만 늘어요 | 호출을 빼요 |
| D-5 | 상대 경로의 `..` 와 `$HOME` 밖의 환경 변수 꼴은 받아요 | 절대 홈 경로가 아니에요. 목록을 늘려도 다 막지 못해요 | 조각 검사를 더해요 |
| D-6 | 필드를 대는 곳은 plan 의 검사와 `project-plan` 의 입력 형식이에요 | 카드가 말하는 문장이 여기서 나와요. 모르는 키는 사용자가 쓴 글이라 대지 않아요 | 다른 파서에도 더해요 |

## Plan

1. RED: `tests/profile_command_values.rs` 에 `accept_`, `refused_`, `other_fields_`, `field_`, `shape_` 시험을 먼저 써서 실패를 기록해요.
2. GREEN: `src/contract.rs` 의 분류 줄, `src/profile.rs` 의 값 검사·필드별 거절·입력 형식 거절, `src/ops/project_plan.rs` 의 호출.
3. 운영 문서, `cargo fmt --all`, `check affected --card` 한 번, READY.

## Risks

- R-1: 정당한데 계속 거절되는 값이 있어요: `2>/dev/null`, `scheme://` 주소, Bazel 의 `//...`, 산문의 `a / b`, `/p:` 꼴 플래그, `$PWD:/app`, `"$PWD"/x`, `app/[slug]/page`, `(auth)/login`, `{a,b}/x`, `-Iinclude/x`, 한글 낱말 바로 뒤의 `/`, 모든 백슬래시.
- R-2: 전에는 `/` 금지 때문에 거절됐지만 이제 받는 값이 있어요: `../../Users/name`, `x/Users/name`, `$USERPROFILE/x`, scheme 없는 `host.invalid/path`, `user:pw@host/db`, 목록의 비밀 모양에 걸리지 않는 `/` 든 비밀. 뒤의 둘은 우연히 얻던 보호가 없어지는 것이라 감사가 봐야 해요.
- R-3: 이 변경 전의 `moai` 는 `/` 값이 든 프로필을 읽을 때마다 거절해요. 설치된 바이너리를 올린 뒤에 그런 값을 저장해야 해요.
- R-4: `docs/rust-product-path.md` 의 "같은 portable 검사" 문장이 낡아요. 그 파일은 다른 카드가 잡고 있어 이 변경에서 고치지 않아요.

## Result

2026-11-26 에 macOS 에서 돌렸어요. 기준은 NeXT `b3048a821a` 이고, 확인 명령은 커밋 직전의 워크트리에서 적힌 그대로 차례로 돌렸어요(모두 exit 0, 기록 `target/dana/ac.log`).

- RED: 소스를 바꾸기 전 `cargo test --locked --offline --test profile_command_values` 는 13개 가운데 11개가 실패하고 2개가 통과했어요(exit 101). 통과한 둘은 바뀌지 않는 동작을 고정하는 시험이에요.
- AC-1: met — 3개 통과.
- AC-2: met — 2개 통과.
- AC-3: met — 1개 통과.
- AC-4: met — 5개 통과.
- AC-5: met — 2개 통과.
- AC-6: met — `profile::` 23개, `contract::` 7개 통과.
- AC-7: met — 두 줄 각각 1개 통과.

더 적을 것:

- 필드의 순서: 하위 검사(`versioning`, `shared_files` 들)가 먼저 돌고, 그다음 프로필 키 표의 순서로 첫 필드를 대요. 결정 안에서는 `id`, `topic`, `value`, `source`, `scope` 순서이고 위치는 0 부터 세요.
- 비밀 모양의 결정 id 는 다른 검사에 걸리지 않으면 전처럼 받아요. 거절 문장에는 보이지 않아요.
- `project_id`, `hosts`, `package_pin`, 결정의 `id`·`source` 의 `/` 거절은 각 필드의 형식 검사가 이미 하던 일이에요.
- 영향 범위 검사(`check affected --card`)와 독립 감사는 카드에 남겨요. Windows 와 Linux 는 돌리지 않았어요.
