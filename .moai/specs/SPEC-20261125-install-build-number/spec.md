---
id: SPEC-20261125-install-build-number
title: "설치용 빌드 번호는 moai 가 내주고, 검사용 빌드는 번호 없이"
status: draft
revision: 1
created: 2026-11-25
---

# 설치용 빌드 번호는 moai 가 내주고, 검사용 빌드는 번호 없이

## Intent

개발 빌드의 `{build}` 번호(`MOAI_LOCAL_BUILD`)를 지금은 사람이 손으로 정해요. 그래서 설치본은 `4.2648.26` 인데 검사 영수증용 빌드가 27~33 을 나눠 썼고(레인 세 곳), 리더의 설치 빌드가 33 과 겹칠 뻔해 34 로 바꿨어요. 사용자가 2026-11-25 에 정했어요: "용 번호는 moai 가 내주게 하자". 리더의 제안은 "검사용 빌드는 번호 없이 돌고, 설치용 번호는 moai 가 내주게 하면 26 → 27 처럼 이어져요" 였어요.

이 SPEC 은 그 결정을 두 조각으로 정해요. (1) `moai build-number next` 가 다음 설치용 번호를 원자적으로 내주고 기록해요. (2) 검사용 빌드(`moai check affected` 와 그 안의 `cargo`)는 번호 환경 변수를 물려받지 않아 번호 없는 빌드(`unnumbered build`, build 0)로 돌아요. HeadVer 규칙(번호는 올라가기만 하고 다시 쓰지 않음, `MOAI_LOCAL_BUILD` 와 `MOAI_BUILD_NUMBER` 를 함께 두지 않음, 버전만 옮기는 커밋 없음)은 그대로예요. 번호의 기록은 사용자 HOME 의 프로젝트 상태 폴더에 쓰이므로 이 카드는 tier 2 이고, 코드는 사용자가 이 SPEC 을 승인한 뒤에 써요.

## Interview

- Q1: 빌드 번호를 누가 정하는가? -> A: moai 가 내줘요. 사용자(2026-11-25, 이 SPEC 의 카드 본문 그대로): "용 번호는 moai 가 내주게 하자"; 이는 리더의 제안 '검사용 빌드는 번호 없이 돌고, 설치용 번호는 moai 가 내주게 하면 26 → 27 처럼 이어져요'에 대한 답이에요.
- Q2: HeadVer 규칙은 바뀌는가? -> A: 바뀌지 않아요(이 SPEC 의 카드 본문): 번호는 올라가기만 하고 다시 쓰지 않으며, `MOAI_LOCAL_BUILD` 와 `MOAI_BUILD_NUMBER` 를 함께 두지 않고, 버전만 옮기는 커밋은 없어요.
- Q3: 번호를 어디에 기록하고 어떻게 시작하는가, 번호만 받고 빌드가 실패하면 어떻게 하는가? -> A: 아래 Options 의 추천안이 초안이에요(O-1…O-5). 사용자가 아직 고르지 않았어요.

## Non-goals

- 릴리스 번호(`release/v{head}.{yearweek}.{n}` 의 `{n}`)와 CI 번호(`MOAI_BUILD_NUMBER`): `SPEC-20261125-headver-weekly-release` 와 `SPEC-20261117-headver-ci-number` 가 정해요. 이 SPEC 은 건드리지 않아요.
- `HEADVER`(head 번호)를 올리는 일, `Cargo.toml` 의 `version`, 태그나 릴리스 게시.
- 설치 자체(`~/.moai/versions/<버전>` 폴더를 만들고 `bin/moai` 링크를 옮기는 일): 이 SPEC 은 번호만 내주고 설치하지 않아요.
- 번호를 내준 뒤 실제로 그 번호로 빌드하게 강제하는 일. 내준 번호는 쓰였든 아니든 다시 내주지 않아요(D-2).

## Requirements

- REQ-1: When `moai build-number next` runs, the command shall record one new build number that is larger than every number it recorded before and larger than every build number of an installed version folder of the same head, and then print it.
- REQ-2 (REQ-1): IF two runs of the command overlap, THEN the command shall give them different numbers.
- REQ-3 (REQ-1): IF the record cannot be read as whole numbers, locked or written, THEN the command shall print no number, record nothing and name the cause.
- REQ-4: When `moai build-number show` runs, the command shall print the last recorded number with its time, its session and the number the next run would give, and write nothing. · failure: none — `show` only reads, so a failing read has nothing to undo; an unreadable record is named as in REQ-3
- REQ-5: While `moai check affected` runs its steps, it shall remove `MOAI_LOCAL_BUILD` and `MOAI_BUILD_NUMBER` from the environment of every child process it starts.
- REQ-6: While the repository's instruction texts describe a development build number, they shall say that an install build takes its number from `moai build-number next`, that a check build takes none, and that no person picks a number by hand.

## Acceptance

- AC-1 (REQ-1, REQ-3): in a temporary HOME the command hands out 1, 2, 3 on three runs, starts above the highest installed `<head>.<yearweek>.<n>[-dev]` folder of its head, ignores a folder of another head, and refuses an unreadable record without recording · verify: `cargo test --locked --offline --test build_number_cli`
- AC-2 (REQ-2): eight concurrent runs in one temporary HOME print eight different numbers and the record holds exactly those eight · verify: `cargo test --locked --offline --test build_number_cli -- concurrent_runs_get_different_numbers`
- AC-3 (REQ-4): `show` prints the last number and the next one and the record is byte-identical after it · verify: `cargo test --locked --offline --test build_number_cli -- show_writes_nothing`
- AC-4 (REQ-5): a `check affected` run started with both variables set reports a build that has neither (the child's environment holds none of the two) · verify: `cargo test --locked --offline --lib check_affected::unnumbered_check_environment`
- AC-5 (REQ-6): `CLAUDE.md` (HeadVer section, in pieces), `docs/local-rc1-kit.md` and the `moai` skill text say the new rule and no longer tell a person to pick a number · verify: `cargo test --locked --offline --test build_number_texts`
- AC-6 (REQ-1…REQ-6): each new guard is killed by a mutant, and the neighbour tests pass · verify: `moai check affected --card NeXT..HEAD`

## Scope

- write (this card now): `.moai/specs/SPEC-20261125-install-build-number/**`
- write (after the user approves this SPEC, each by a scope request to the leader): `src/build_number.rs` (new: the record, the lock, the floor from the version folders), `src/ops/build_number.rs` (new: the `moai build-number` command, one file per command like its neighbours), `src/check_affected.rs` (the child environment only), `tests/build_number_cli.rs` (new), `tests/build_number_texts.rs` (new), `docs/local-rc1-kit.md`, `skills/moai/**` (the lines that name the build number), the source manifest hashes of those skill files
- write (shared files, by `moai card line-lease`, one line each): `src/lib.rs` (`pub mod build_number;`), `src/ops/mod.rs` and the file that registers a top-level command, `src/help.rs` (the one help entry), `tests/areas.toml` (the first word `build`) and `skills/moai/references/source-manifest.json`
- write (tier 2, a card whose scope names it, with the user's approval of each piece): `CLAUDE.md` (the HeadVer section only)
- read-only: `HEADVER`, `build.rs`, `src/headver.rs`, `src/local_build.rs` (the numbering rules stay as they are), the installed version folders of the home state (listed by the command, never written)
- forbidden: the real HOME in a test (every test uses a temporary HOME), any change of an installed binary or link, any push, tag or release, `Cargo.toml` `version`, `HEADVER`, Python

## Plan

1. 사용자가 Options 를 고르고 이 SPEC 을 승인해요(레인은 승인을 기록하지 않아요). 고른 안이 추천안과 다르면 REQ 와 이 Plan 을 먼저 고쳐요.
2. 시험 먼저(모두 RED 를 봐요): `build_number_cli`(세 번 내주기, 설치 폴더 바닥, 다른 head 무시, 깨진 기록 거절, 여덟 동시 실행, `show` 가 쓰지 않음), `check_affected::unnumbered_check_environment`(두 변수를 준 채 돌려도 자식 환경에 없음), `build_number_texts`(글이 새 규칙을 말함).
3. `src/build_number.rs`: 기록 파일 `~/.moai/projects/<project>/build-number.log`(한 줄에 `<번호> <UTC 시각> <세션>`, 덧붙이기만 해요), 배타 잠금 아래에서 `max(기록의 마지막, 같은 head 의 설치 폴더 최대) + 1` 을 쓰고 `fsync` 한 뒤에만 번호를 출력해요. 잠금은 기다리지 않고(짧은 재시도 뒤 거절) 프로세스를 죽이지 않아요.
4. `src/ops/build_number.rs` 와 등록 줄: `next` 와 `show`, 둘 다 `--project`·`--home`(시험용 임시 HOME)을 받아요. 출력은 JSON 한 개(`build`, `env`: `MOAI_LOCAL_BUILD=<n>`, `version`: `{head}.{yearweek}.{n}-dev`, `recorded`).
5. `src/check_affected.rs`: 모든 자식 프로세스에 `env_remove` 두 개(한 곳에 모아요). 검사용 빌드는 번호 없는 빌드가 돼요.
6. 글: `docs/local-rc1-kit.md` 와 스킬 글의 번호 줄, 그리고 `CLAUDE.md` HeadVer 절의 `{build}` 항목을 "설치용은 `moai build-number next`, 검사용은 번호 없음"으로 조각마다 사용자 승인을 받아 고쳐요. 소스 매니페스트 해시를 다시 계산해요.
7. 위험한 경로(HOME 에 쓰는 명령)는 임시 HOME 두 개와 동시 실행으로 READY 에 직접 돌려 보여요. 뮤턴트(잠금 제거, 바닥 제거, `fsync` 제거, `env_remove` 제거)와 `moai check affected --card` 한 번.

## Options

선택마다 첫째가 추천이고 기본값이에요. 사용자가 고른 것은 Decisions 에 기록해요.

| Id | Question | Options (first = recommended = default) | Cost of the others | Reversible |
| --- | --- | --- | --- | --- |
| O-1 | 번호를 어디에 기록할지 | (a) 사용자 HOME 의 프로젝트 상태 폴더 `~/.moai/projects/<project>/build-number.log`: 모든 워크트리·세션이 한 번호열을 봐요(카드 원장과 같은 자리). (b) 저장소에 추적하는 파일 | (b) 는 번호를 올리는 커밋이 생겨서 "버전만 옮기는 커밋 없음" 규칙과 부딪히고 워크트리마다 갈라져요 | 예(파일 하나) |
| O-2 | 처음 시작하는 번호 | (a) 기록이 없으면 같은 head 의 설치 폴더(`~/.moai/versions/<head>.<yearweek>.<n>[-dev]`) 최대 번호 + 1 로 이어가요: 지금 설치본 위에서 자동으로 이어져요. (b) 첫 실행에 `--first <n>` 을 꼭 받아요 | (b) 는 사람이 번호를 정하는 자리가 한 번 남아요 | 예 |
| O-3 | 내준 번호로 빌드·설치가 실패하면 | (a) 번호는 쓰였든 아니든 다시 내주지 않아요(빈 번호가 생겨도 괜찮아요: 규칙 "다시 쓰지 않음"). (b) `moai build-number release <n>` 으로 되돌려 다시 내줘요 | (b) 는 다시 쓰지 않는다는 규칙을 깨고 되돌리는 경로에서 번호가 겹칠 수 있어요 | 예 |
| O-4 | 명령이 빌드까지 맡을지 | (a) 번호만 내줘요: 쓰는 쪽이 `MOAI_LOCAL_BUILD=<n> cargo build …` 로 빌드해요(설치는 이 SPEC 밖). (b) `moai build-number next -- <명령>` 이 번호를 환경에 넣고 명령을 대신 돌려요 | (b) 는 프로세스를 시작하는 새 경로라 위험한 경로 시험과 중단 처리가 늘어요 | 예 |
| O-5 | 검사용 빌드에서 번호 환경을 없애는 자리 | (a) `moai check affected` 의 자식 프로세스에서만 없애요(한 곳). (b) 레인 시작 때 환경 전체를 정리해요 | (b) 는 레인 밖 세션의 환경까지 건드려요(보안·권한 변경 성격) | 예 |

## Risks

- 기록이 HOME 에 있어서 다른 사용자 계정이나 다른 기계의 번호와는 이어지지 않아요. 한 기계의 한 사용자가 설치용 번호를 내주는 경우만 다뤄요(지금의 사용 그대로예요). 다른 기계의 설치 폴더는 바닥에 읽히지 않아요.
- 번호만 받고 쓰지 않으면 번호에 빈 자리가 생겨요. 규칙은 "올라가기만 하고 다시 쓰지 않음"이라 빈 자리는 허용돼요(O-3 (a)). 버전 비교는 숫자 순서라 영향이 없어요.
- 검사용 빌드가 번호를 잃어서 영수증의 `--version` 둘째 줄이 `unnumbered build` 가 돼요. 영수증이 그 줄을 읽는 시험이 있으면 마지막 `moai check affected` 가 잡아요.
- 기록 파일이 사람 손으로 줄어들면(예: 마지막 줄 삭제) 같은 번호가 다시 나올 수 있어요. 바닥에 설치 폴더 최대 번호를 함께 쓰는 것이 이를 막지만, 설치되지 않은 번호는 막지 못해요(O-2 (a)). 기록의 마지막 줄이 깨졌으면 REQ-3 대로 거절해요.
- `CLAUDE.md` 의 HeadVer 절은 tier 2 파일이에요. 조각마다 사용자 승인을 받아야 하고, 승인 전에는 글만 초안으로 이 SPEC 에 남겨요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 설치용 빌드 번호는 `moai build-number next` 가 내주고, 검사용 빌드는 번호 없이 돌아요 | 사용자 결정 2026-11-25(이 SPEC 의 카드 본문, 말 그대로): "용 번호는 moai 가 내주게 하자" — 리더의 제안 '검사용 빌드는 번호 없이 돌고, 설치용 번호는 moai 가 내주게 하면 26 → 27 처럼 이어져요'에 대한 답 | 새 사용자 결정으로만 |
| D-2 | 내준 번호는 쓰였든 아니든 다시 내주지 않아요 | HeadVer 규칙(`CLAUDE.md` Development build versions): "`{build}` is one counter that only goes up and never restarts; never reuse a number, also not for a rebuild of the same commit" | 새 사용자 결정으로만 |
| D-3 | 기록 자리·시작 번호·실패한 번호·명령의 범위·환경을 없애는 자리는 위 Options 의 추천안을 초안으로 삼아요 | 사용자가 고르기 전까지의 기본값이에요(레인이 정하는 것이 아니에요). 승인 기록은 사용자만 남겨요 | Options 에서 다른 안을 고르면 REQ·Plan 을 고쳐요 |
