---
id: SPEC-20261125-skill-versioning-release-text
title: "The shipped versioning text states today's release rule once, in short sentences, and the interview only points to it"
status: draft
revision: 1
created: 2026-11-25
---

# The shipped versioning text states today's release rule once, in short sentences, and the interview only points to it

## Intent

사용자가 2026-11-25 에 정한 버전 규칙(개발 `-dev`, 릴리스는 배포 월요일의 주와 0부터의 순번, 설치 빌드 번호는 moai 가 내줘요)은 이미 승인된 두 SPEC 에 있고, 날짜 규칙은 `versioning.md` 에 들어갔어요. 남은 일은 글의 모양이에요. 첫 문단은 330 단어 한 문장에 유니코드 범위 74개가 들어 있고, `interview.md` 의 `versioning` 문단은 내부 카드 번호를 말하며 같은 규칙을 한 번 더 적고, 주 계산 명령은 이 Mac 에서 돌지 않아요. 이 SPEC 은 규칙의 뜻을 바꾸지 않고 글을 짧게 한 곳에 모으고, 그 글을 고정하던 시험을 같이 고쳐요. 글을 읽는 쪽은 사용자의 프로젝트에서 `/moai project` 와 릴리스를 준비하는 에이전트예요.

## Interview

- Q1: 스킬 지침을 어떻게 손볼까요? → A: "moai 스킬 지침 전체 전수 조사해서 검색 말고 전체 개별로 읽고 개선점을 찾아서 모두 업데이트 개선 보고 후 진행"(사용자 지시 2026-11-25, 조사 보고서 `~/.moai/reports/moai-adk/skill-survey-20261125.md`, 조각 8: 버전 글). 이 카드는 그중 버전 글이에요.
- Q2: 버전 규칙은 무엇인가요? → A: 이미 사용자가 승인했어요(`SPEC-20261125-headver-weekly-release`, 승인 2026-11-25, digest `sha256:3939a62d296de5533d0060cf0d886c34fd66eb3780aaa702696e2899b27b5f4d`): 개발 빌드는 `{head}.{yearweek}.{build}-dev`, 릴리스는 `{head}.{배포 월요일의 주}.{0부터의 순번}`, 릴리스 브랜치는 `release/v{head}.{yearweek}.{n}`. 설치 빌드 번호는 `moai build-number next` 가 내줘요(`SPEC-20261125-install-build-number`, 승인 2026-11-25). 이 SPEC 은 그 규칙을 바꾸지 않고 글만 맞춰요(`headver-weekly-release` 의 AC-5 가 이미 `versioning.md` 의 날짜 규칙을 고쳤고, 이 SPEC 은 남은 조각을 해요).
- Q3: 남은 조각은 무엇인가요? → A: 실제 글을 읽은 결과예요(2026-11-25): (1) `versioning.md:5` 는 한 문장 330 단어이고 유니코드 범위 74개를 나열해요. (2) `interview.md:42` 는 내부 카드 번호(`card C-69`)를 말하고 `versioning.md` 와 같은 규칙(`Other` 답을 기록하는 법, 거절하는 글자)을 한 번 더 적어요. (3) `versioning.md:35` 의 `date -u -d <that Monday> +%G%V` 는 GNU `date` 만 되고 `%G%V` 는 6자리(`202649`)를 내서 글이 그 뒤에 "last two digits" 를 덧붙여 읽어야 해요. 이 Mac 에서 `date -u -j -f %Y-%m-%d 2026-11-30 +%g%V` 는 `2649` 를 내요(2027-01-04 는 `2701`, 2026-12-28 은 `2653`).
- Q4: 아래 선택지(O-1…O-5)는 무엇으로 정했나요? → A: 추천안(첫째)이 초안이에요. 사용자가 아직 고르지 않았어요. 레인은 승인을 기록하지 않아요.

## Non-goals

- 규칙의 뜻을 바꾸는 일. 이름 규칙, `-dev`, 순번, 관문(태그와 릴리스 게시는 사용자의 선택, 태그는 옮기지 않아요)은 그대로예요.
- `CLAUDE.md` 8단계와 HeadVer 절, `docs/decisions-2026-11-08.md`(`SPEC-20261125-headver-weekly-release` 의 AC-6): 그 파일을 범위에 둔 카드가 따로 해요.
- `interview.md` 의 다른 줄(`C-51`, 퇴역 호스트 이름, 옛 SPEC 이름)과 `docs/design-review/appendix/operations/project_plan.md` 의 `card C-69`: 이 카드의 범위가 아니에요. 같은 종류의 글이 남아 있다고 결과에 적어요.
- 코드의 거절 규칙(`src/versioning.rs` 의 표와 검사)과 그 표가 유니코드 원본과 같은지 보는 시험. 글만 줄여요.
- 태그, 릴리스, 푸시, 설치.

## Requirements

- REQ-1: While a project's profile names HeadVer, the shipped versioning text shall state the development form `{head}.{yearweek}.{build}-dev`, the release form `{head}.{release week}.{serial}` and the branch names `release/v{head}.{yearweek}.{n}` and `release/rc/v{head}.{yearweek}.{n}` with no other release-branch shape.
- REQ-2: The shipped versioning text shall state the refusal of default-ignorable and format characters in one line that names the Unicode 18.0.0 sets and lists no code point range.
- REQ-3: The shipped versioning text shall hold the profile key as a short list of three points and the refusal rules of custom text as one list, in sentences of at most 60 words.
- REQ-4: The shipped versioning text shall hold how an `Other` answer to each `versioning` row is recorded and that the keep option can be corrected, so that the interview text does not repeat them.
- REQ-5: The `versioning` paragraph of the interview text shall point to `internal/versioning`, name no card, SPEC or internal number and hold no rule of its own.
- REQ-6: The shipped versioning text shall compute the release week of a Monday with a command form that runs on macOS and on Linux and prints four digits.
- REQ-7: While the unicode range list is gone from the shipped versioning text, the table tests of `src/versioning.rs` shall check the code tables against the Unicode source and the operations appendix and read no range from the shipped versioning text.
- REQ-8: While the shipped versioning text changes, every command it names shall be one that `moai --help` lists. · failure: none — a check that finds an unknown command stops the card, nothing is applied
- REQ-9: While a skill file changes, the skill's source manifest shall hold the new hash of that file.

## Acceptance

- AC-1 (REQ-1, REQ-6): the text has the three forms and both branch names, has no `release/<40hex>` and no `-d <that Monday>` form, and the week command prints `2649`, `2701` and `2653` for the Mondays 2026-11-30, 2027-01-04 and 2026-12-28 · verify: `cargo test --locked --offline --test versioning_release_texts`
- AC-2 (REQ-2, REQ-3): the text holds no `U+` range other than the two separators, the refusal sits in one line, and no sentence of the profile key passage has more than 60 words · verify: `cargo test --locked --offline --test versioning_release_texts -- the_unicode_refusal_is_one_line the_profile_key_passage_has_short_sentences`
- AC-3 (REQ-4, REQ-5): the interview paragraph is one line that names `internal/versioning`, no card, SPEC or internal number and no rule, and the versioning text holds the `Other` answers and the keep correction · verify: `cargo test --locked --offline --test versioning_release_texts -- the_interview_only_points_to_the_versioning_text the_versioning_text_holds_the_other_answers`
- AC-4 (REQ-4, REQ-5, REQ-7): the moved pins pass in their new place and the table tests pass without the text half · verify: `cargo test --locked --offline --test agent_defs_package versioning_`
- AC-5 (REQ-7): the unicode tables still equal the source and the appendix · verify: `cargo test --locked --offline --lib versioning::tests::default_ignorable_table_matches_the_unicode_source`
- AC-6 (REQ-8, REQ-9): every command named in the changed text is listed by `moai --help` of the build of this head, and the manifest hashes agree · verify: `moai check affected --card NeXT..HEAD`

## Scope

- write (this card now): `.moai/specs/SPEC-20261125-skill-versioning-release-text/**`
- write (after the user approves this SPEC): `skills/moai/references/internal/versioning.md`, `skills/moai/references/internal/interview.md` (the `versioning` paragraph only), `src/versioning.rs` (the two table tests, the text half only), `tests/agent_defs_package.rs` (the pins of the two versioning tests), `tests/versioning_release_texts.rs` (new; its first word `versioning` is listed in `tests/areas.toml`, so that file stays as it is)
- write (shared file, by `moai card line-lease`, one line): `skills/moai/references/source-manifest.json` (the two hashes)
- read-only: `CLAUDE.md`, `docs/decisions-2026-11-08.md`, `docs/design-review/appendix/operations/project_plan.md`, `HEADVER`, `src/headver.rs`
- forbidden: any push, tag or release, `Cargo.toml` `version`, `HEADVER`, the real HOME, Python (no version check either)

## Plan

1. 시험 먼저: `tests/versioning_release_texts.rs` 에 AC-1…AC-3 의 시험을 쓰고 RED 를 봐요(지금 글은 범위 74개를 나열하고, `interview.md` 는 `card C-69` 를 말해요).
2. `versioning.md`: 첫 문단을 세 점 목록과 거절 규칙 한 목록으로 나누고, 유니코드 줄을 한 줄로 줄이고, `Other` 답 기록과 keep 정정을 `## Keep the existing scheme` 아래 한 절로 옮기고, 릴리스 주 계산 명령을 macOS·Linux 둘 다 되는 꼴로 바꿔요.
3. `interview.md`: `versioning` 문단을 `internal/versioning` 을 가리키는 한 줄로 줄여요.
4. 고정 시험을 고쳐요: `src/versioning.rs` 의 두 표 시험은 글 쪽 단언(`versioning.md lacks`)을 빼고 부록 쪽을 남겨요. `tests/agent_defs_package.rs` 의 `interview.contains(...)` 세 단언은 `versioning.md` 로 옮기고 이유를 주석에 적어요.
5. 소스 매니페스트의 해시 두 줄을 line-lease 로 같은 커밋에 고쳐요. `moai check affected --card` 를 한 번 돌려 영수증을 남겨요.

## Options

선택마다 첫째가 추천이고 기본값이에요. 사용자가 고른 것은 Decisions 에 기록해요.

| Id | Question | Options (first = recommended = default) | Cost of the others | Reversible |
| --- | --- | --- | --- | --- |
| O-1 | 유니코드 거절 규칙을 글에서 어떻게 보일지 | (a) 한 줄: "유니코드 18.0.0 의 Default_Ignorable_Code_Point 와 General_Category Cf 에 드는 글자(보이지 않거나 둘레 글자의 모양을 바꾸는 글자)를 거절해요" 만 적고 범위는 코드의 표가 갖는다. (b) 범위 74개를 그대로 둔다 | (b) 는 330 단어 문장이 그대로 남고 카드 기준 2 를 못 채워요 | 예 |
| O-2 | `Other` 답을 기록하는 법과 keep 정정을 어디에 둘지 | (a) `versioning.md` 한 곳에 두고 `interview.md` 는 가리키기만 한다(기준 3). (b) 둘 다에 둔다 | (b) 는 같은 규칙이 두 곳에서 어긋날 수 있고 기준 3 을 못 채워요 | 예 |
| O-3 | 범위 목록을 고정하던 시험 | (a) 두 표 시험에서 `versioning.md` 쪽 단언만 빼고(표와 원본, 부록 단언은 남김) 이유를 주석에 적는다. (b) 시험을 통째로 지운다 | (b) 는 코드 표가 원본과 같은지 보는 보호를 잃어요 | 예 |
| O-4 | 릴리스 주 계산 명령 | (a) 날짜를 정한 뒤 `date -u -j -f %Y-%m-%d <그 월요일> +%g%V`(macOS)와 `date -u -d <그 월요일> +%g%V`(GNU/Linux)를 함께 적는다. (b) 명령 없이 "ISO 주-연도의 끝 두 자리와 ISO 주" 만 적는다 | (b) 는 에이전트가 직접 계산해 틀릴 수 있어요 | 예 |
| O-5 | `interview.md` 의 다른 줄(`C-51`, 퇴역 호스트, 옛 SPEC 이름)과 부록의 `card C-69` | (a) 이 카드에서 건드리지 않고 결과에 적어 따로 카드로 남긴다(범위 밖). (b) 같이 고친다 | (b) 는 카드 범위를 넘고 다른 카드와 겹쳐요 | 예 |

## Risks

- GNU 형태 `date -u -d` 는 이 Mac 에서 돌려 볼 수 없어요(not_run). 글에는 널리 알려진 GNU 형태를 적고 결과에 not_run 으로 적어요. macOS 형태는 위 세 날짜로 돌려 확인했어요.
- `tests/agent_defs_package.rs` 는 카드 C-523 도 쥐고 있어요(범위가 겹쳐요, 파일 잠금 없음). 이 카드는 두 시험 함수의 단언만 고쳐요.
- 새 시험이 `versioning.md` 의 줄 모양(한 줄, 60단어)을 고정하므로 다음에 글을 고치는 카드가 시험도 같이 고쳐야 해요. 모양이 아니라 사실(이름 규칙, 한 줄)만 고정하게 시험을 써요.
- `moai build-number` 는 설치된 `moai`(빌드 41)에 아직 없어요(C-515 가 착지한 뒤의 빌드에서만 나와요). 글의 명령 확인은 이 head 의 빌드로 해야 해서 빌드가 필요해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 규칙의 뜻은 바뀌지 않고 글만 맞춰요 | 사용자가 승인한 `SPEC-20261125-headver-weekly-release` 와 `SPEC-20261125-install-build-number` 가 규칙을 정했고 이 SPEC 의 Non-goals 가 그것을 못박아요 | 해당 없음(규칙을 바꾸는 결정이 아니에요) |
| D-2 | O-1…O-5 는 위 Options 의 추천안을 초안으로 삼아요 | 사용자가 고르기 전까지의 기본값이에요(레인이 정하는 것이 아니에요). 승인 기록은 사용자만 남겨요 | Options 에서 다른 안을 고르면 REQ 와 Plan 을 고쳐요 |
