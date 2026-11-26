---
id: SPEC-20261126-lane-advisor-opt-in
title: "moai cc -a 는 advisor 를 기본으로 켜지 않고, --advisor 를 준 레인에만 켜요"
status: draft
revision: 2
created: 2026-11-26
branch: spec/lane-advisor-opt-in-20261126
---

# moai cc -a 는 advisor 를 기본으로 켜지 않고, --advisor 를 준 레인에만 켜요

## Interview

- Q1: 문제가 무엇이었나요? -> A: 2026-11-26 리더 대화에서 사용자가 "moai cc -a 사용할 때 Advisor unavailable (too_many_requests) 메시지가 자주 나온 원인은?" 이라고 물었어요. 코드를 읽으면 `-a` 가 모든 `cc` 레인에 `--advisor opus` 를 붙여요. 한도 숫자 자체는 확인하지 않았어요.
- Q2: 사용자가 고른 것은요? -> A: 사용자 말 그대로 "2번으로 가자"(카드 원장의 결정 노트 10315). 2번은 "자문(advisor)을 일부 레인에만 켜기: 감사처럼 판단이 많은 레인만 켜서 요청을 줄임"이에요.
- Q3: 이 SPEC 을 쓰면서 사용자에게 새로 물은 것은요? -> A: 없어요. Decisions 의 D-1~D-3 은 승인할 때 고르는 선택지예요.
- Q4: 개정 2 는 왜 있나요? -> A: 구현 중에 `docs/host-support-matrix.md` 의 호스트 세션 설명에도 "`-a`는 `cc`에서 `--advisor opus`도 붙입니다"라는 문장이 남아 있는 것을 확인했어요. 개정 1 의 Scope 에 이 파일이 빠져 있어서 AC-5 와 Scope 에 더했어요. 요구와 결정은 바뀌지 않았어요.

## Intent

`moai cc -a` 로 연 레인은 advisor 없이 시작해요. advisor 가 필요한 레인은 `--advisor` 를 붙여 열고, 모델을 빼면 `opus` 예요. 이 결정은 SPEC-20261125-agent-team-session-names 의 REQ-9 가운데 "`-a` 의 기본 advisor 는 `opus`"(그 SPEC 의 D-6, D-13, D-14)를 바꿔요. 읽는 사람: 사용자(승인), 레인, 감사자, 착지 담당.

## Non-goals

- `-l`, 역할 없는 launch, `glm`, 옛 `session` 명령의 동작 변경.
- 사용자 설정 `launcher.advisor` 의 뜻이나 값 검사 변경, advisor 모델 순위 검사 변경.
- `-a` 의 권한 모드, 모델, effort, 이름, 첫 입력 변경.
- 이미 떠 있는 세션에 대한 조치. 다음에 새로 시작할 때부터 바뀌어요.
- advisor 요청 한도를 재거나 429 를 다시 시도하는 일.

## Requirements

- REQ-1: When `-a` is given on the `cc` host without `--advisor` and no `launcher.advisor` setting names a model, the launcher shall pass no `--advisor` · failure: none — the session starts without an advisor.
- REQ-2: When `--advisor` is given without a model on the `cc` host, the launcher shall pass `--advisor opus` · failure: none — REQ-7 covers a main model that ranks above `opus`.
- REQ-3: When `--advisor <model>` or the `launcher.advisor` setting names a model, the launcher shall keep the existing precedence of `--no-advisor` or `--advisor` over the setting over no advisor · failure: none — REQ-8 covers both options together.
- REQ-4: When a `glm` launch asks for an advisor in either form, the launcher shall pass no `--advisor` and keep the existing glm note · failure: none — the advisor is left out and the launch goes on.
- REQ-5: When `--dry-run` composes a `-a` launch, the launcher shall show the argv and a note saying that no advisor is passed unless `--advisor [model]` or `launcher.advisor` asks for one · failure: none — the note is part of the composed answer.
- REQ-6: The repository rules and launcher documents shall record the user decision of 2026-11-26 in place of the `-a` advisor default · failure: none — the documents are checked by reading them.
- REQ-7 (REQ-2): IF `--advisor` is given without a model while the main model ranks above `opus`, THEN the launcher shall refuse the launch with the message `advisor must rank at or above the main model` and start no host.
- REQ-8 (REQ-3): IF `--advisor` is given with `--no-advisor`, THEN the launcher shall refuse the launch and start no host.

## Acceptance

- AC-1 (REQ-1): `-a` 만 주면 argv 에 `--advisor` 가 없어요(`cc`, `glm` 둘 다). 주 모델을 `fable` 로 바꿔도 없고, "순위가 낮아 뺐다"는 옛 노트도 없어요. 설정 파일의 `none` 도 같아요 · verify: `cargo test --locked --offline --test launcher_role_flags advisor_`
- AC-2 (REQ-2, REQ-7): `cc -a --advisor` 와 `cc -a --advisor --name x` 처럼 모델을 빼면 `--advisor opus` 가 붙고, `--model fable` 과 함께 모델 없는 `--advisor` 를 주면 거절되고 host 가 시작하지 않아요 · verify: `cargo test --locked --offline --test launcher_role_flags advisor_`
- AC-3 (REQ-3, REQ-4, REQ-8): `--advisor fable` 과 설정 `launcher.advisor` 는 지금처럼 켜고, `--no-advisor` 는 설정을 끄고, 두 옵션을 같이 주면 거절되며, `glm` 은 `--advisor` 를 넘기지 않고 지금의 노트를 보여요 · verify: `cargo test --locked --offline --test launcher_advisor_setting`
- AC-4 (REQ-5): `--dry-run` 의 argv 와 notes 가 advisor 없는 `-a` 와 `-a --advisor` 두 꼴을 보이고, `-a` 노트에 `--advisor [model]` 이 적혀요. 도움말 표의 `--advisor` 설명이 모델을 뺄 수 있다고 말해요 · verify: `cargo test --locked --offline --test launcher_role_flags advisor_`, `cargo test --locked --offline --bin moai-cli help_rules`
- AC-5 (REQ-6): `CLAUDE.md` 의 Scope and sessions 줄, `docs/decisions-2026-11-08.md` 의 새 행, `docs/rust-product-path.md`, `docs/user-entry-draft.md`, `docs/host-support-matrix.md` 의 런처 설명이 새 규칙을 말하고 "`-a` 가 `--advisor opus` 를 붙인다"는 문장이 남지 않아요. SPEC 검사를 통과해요 · verify: `cargo test --locked --offline --test repo_claude_md`, `moai spec check SPEC-20261126-lane-advisor-opt-in --project .`

## Scope

- write: `.moai/specs/SPEC-20261126-lane-advisor-opt-in/**`
- write (승인 뒤): `src/main.rs`, `src/session.rs`, `tests/launcher_role_flags.rs`, `tests/launcher_advisor_setting.rs`, `CLAUDE.md`, `docs/decisions-2026-11-08.md`, `docs/rust-product-path.md`, `docs/user-entry-draft.md`, `docs/host-support-matrix.md`
- shared (line lease): `src/help.rs`(`advisor` 설명), `tests/fixtures/help/*.txt`(골든이 바뀔 때)
- read-only: `src/launcher_terminal.rs`(`--advisor` 가 나오는 곳은 줄바꿈 테스트의 예시값이라 바꾸지 않아요), 앞의 두 런처 SPEC
- forbidden: 사용자의 설정 파일, `skills/**`, 이 목록 밖의 파일

## Plan

1. 이 SPEC 초안을 커밋하고 `SPEC-READY` 로 리더에게 알려요. 승인은 리더가 사용자 결정 노트를 근거로 기록해요. 레인은 승인을 기록하지 않아요.
2. 범위를 넓히고 RED: `tests/launcher_role_flags.rs` 의 `-a` 기본 advisor 테스트 두 개를 `advisor_` 로 시작하는 새 기대값 테스트로 바꿔요.
3. GREEN: `--advisor` 를 값이 없어도 되는 옵션(빠지면 `opus`)으로 바꾸고, `-a` 의 기본 advisor(`default_advisor`)와 그 노트를 지우고, `-a` 노트와 도움말 글을 고쳐요.
4. 문서 다섯 곳을 고치고, 새 가드 테스트마다 뮤턴트 한 줄, `cargo fmt --all`, clippy, `moai check affected --card` 한 번, READY 노트. `CLAUDE.md` 를 고치므로 감사는 tier 2 전체 감사예요.

## Risks

- 감사처럼 advisor 를 원하던 레인이 옵션을 빠뜨리면 advisor 없이 떠요. 완화: `--dry-run` 노트와 도움말이 켜는 법을 말해요.
- `--advisor` 뒤에 붙는 다음 낱말은 모델로 읽혀요(`--advisor` 다음에 PROMPT 를 주는 경우). 지금도 값이 필요해서 같은 꼴이고, 알 수 없는 값은 거절돼요.
- 사용자 설정 `launcher.advisor` 가 켜져 있으면 `-a` 레인에도 계속 붙어요(D-3). 그 경우 요청이 줄지 않아요.
- 요청 한도에 걸리는 빈도가 실제로 줄어드는지는 이 변경의 테스트로 확인하지 않아요.

## Decisions

승인 전 제안이에요. 기본값은 권장안이에요.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 켜는 옵션(기본): 기존 `--advisor` 를 값이 없어도 되는 옵션으로 바꿔 빠지면 `opus` 예요. `cc`·`glm` launch 전체에 같아요. 대안: `-a` 전용 새 옵션을 더해요 | 카드 기준 1 의 예("`-a` 에 advisor 옵션, 기본 모델 opus")이고 새 옵션 이름이 늘지 않아요 | 값을 다시 필수로 해요 |
| D-2 | 모델 없는 `--advisor` 인데 주 모델이 `opus` 보다 높으면(기본): 직접 청한 것이라 지금처럼 거절해요. 대안: 앞 SPEC 의 D-14 처럼 노트를 남기고 빼요 | 사용자가 청한 advisor 를 말없이 빼지 않아요. D-14 는 청하지 않은 기본값에 대한 규칙이었어요 | 빼는 경로를 되살려요 |
| D-3 | 사용자 설정 `launcher.advisor`(기본): `-a` 에도 지금처럼 적용돼요. 대안: `-a` 는 설정을 무시하고 옵션만 받아요 | 설정은 사용자가 직접 켠 것이에요. 바꾸면 설정의 뜻이 레인에서만 달라져요 | `-a` 에서 설정을 거르는 줄을 더해요 |
