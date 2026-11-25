---
id: SPEC-20261125-agent-names-bible
title: "Lane session names come at random from a list of Biblical names, among the names that are free"
status: draft
revision: 1
created: 2026-11-25
branch: spec/agent-names-bible-20261125
---

# Lane session names come at random from a list of Biblical names, among the names that are free

## Interview

- Q1: 레인 세션 이름의 꼴을 바꾸나요? → A: "moai-adk-agent-woody (Recommended)"(사용자 2026-11-25, 리더 대화의 선택지, 카드 본문이 옮긴 말). 꼴은 그대로 `<프로젝트 이름>-agent-<이름>` 이에요.
- Q2: 이름은 어떻게 고르고 어떤 이름을 쓰나요? → A: "1번으로 가고 성경 인물만 사용하자"(같은 결정, 카드 본문이 옮긴 말). 뜻은 비어 있는 이름 가운데 무작위로 고르고, 목록은 성경 인물만 쓰고, 토이 스토리 이름은 쓰지 않는다는 거예요.

## Intent

`moai cc|glm -a` 는 지금 영어 이름 24개(`ben` … `zoe`) 가운데 앞에서부터 첫 빈 이름을 골라요(`src/session.rs` 의 `AGENT_NAMES` 와 `pick_agent_name`, `SPEC-20261125-agent-team-session-names` 의 결정 D-8 과 요구 3). 사용자 결정에 따라 이 SPEC 은 목록을 성경 인물 이름으로 바꾸고, 고르는 법을 "비어 있는 이름 가운데 무작위"로 바꿔요. 세션 이름의 꼴 `<프로젝트 이름>-agent-<이름>`, 레인을 알아보는 접두사 규칙, `--name` 이 이름을 정하는 길, 모든 이름이 쓰일 때 `--name` 을 알리는 한 메시지는 그대로예요.

이 SPEC 은 `SPEC-20261125-agent-team-session-names` 의 D-8 과 요구 3, 그리고 `SPEC-20261125-session-names-project-prefix` 의 "첫 이름" 문구를 대체해요.

읽어서 확인한 사실(2026-11-25, NeXT `59ac3344cd`):

- `src/session.rs` 의 `AGENT_NAMES` 는 24개 상수이고, `max_project_len` 이 가장 긴 이름 길이로 프로젝트 이름의 한도를 정해요(등록부 이름 한도 64자 − `-agent-` 7자 − 가장 긴 이름). 지금 한도는 53자예요.
- `pick_agent_name(dir, now)` 는 목록 순서대로 돌며 `.live` 잠금이 잡혔거나 등록부 항목의 임대가 이어지는 이름을 건너뛰고 첫 이름을 돌려줘요. 부르는 곳은 `src/main.rs` 하나예요.
- "첫 빈 이름" 문구는 `src/main.rs` 의 도움말 주석, `src/help.rs` 의 한국어 도움말, 도움말 골든 4개, `docs/user-entry-draft.md` 에 있어요.
- 같은 파일 `src/session.rs` 를 다른 카드(`-a` 이름 충돌을 다루는 등록 열쇠 카드)가 예약하고 있어요. 먼저 착지하는 쪽이 정해지면 나중 쪽이 합쳐요(Risks R-2).

읽는 사람: 이 변경을 승인하는 사용자, 구현하는 레인, 독립 감사자.

## Non-goals

- 세션 이름의 꼴, `-l` 의 이름 `<프로젝트 이름>-leader`, `--name`, 레인 확인(`--lane-check`), 등록부, 모델·advisor·권한 규칙은 바뀌지 않아요.
- 이미 떠 있는 세션의 이름은 바꾸지 않아요. 새 목록은 다음 `-a` 시작부터 써요.
- 난수 외부 crate 는 더하지 않아요. 같은 이름을 동시에 고른 두 시작은 지금처럼 독점 잠금이 나중 쪽을 거절해요(다시 고르지 않아요).
- `CLAUDE.md`, 스킬 글(`skills/**`), `<영어 이름>` 자리표시자의 철자는 바꾸지 않아요.

## Requirements

- REQ-1: The launcher shall take the names of `-a` from one constant list of 48 lowercase ASCII names of Biblical people, one word each, with no duplicate and none longer than 9 characters.
- REQ-2: When `moai cc` or `moai glm` is given `-a` or `--agent` without `--name`, the launcher shall start the host with the session name `<project name>-agent-<name>`, the name chosen at random with equal chance among the names that REQ-3 does not exclude.
- REQ-3: When the launcher chooses a name, it shall exclude every name whose session live lock is held or whose registry entry has a running lease. · failure: none — a lock or an entry that cannot be read refuses the launch with its own error, as the launcher does today.
- REQ-4 (REQ-2): IF every name of the list is excluded, THEN the launcher shall refuse the launch with one message that names `--name` and start no host.
- REQ-5: The launcher shall cut the project name so that `<project name>-agent-<name>` fits the registry's 64 characters for the longest name of the list.
- REQ-6: When `--name <name>` is given with `-a`, the launcher shall use the name as given and choose none. · failure: none — a name the registry refuses is refused as it is today.
- REQ-7: When the launcher help or a repository document says how `-a` takes its name, it shall say that the name is chosen at random among the free names. · failure: none — the help golden files and the scan tests pin the text.
- REQ-8: When this change lands, the decisions log shall carry one line that replaces D-8 and REQ-3 of `SPEC-20261125-agent-team-session-names` and the first-name wording of `SPEC-20261125-session-names-project-prefix`. · failure: none — the test of AC-6 reads the line.

## Acceptance

- AC-1 (REQ-1, REQ-5): the list is pinned literally (48 names, lowercase ASCII, no duplicate, none over 9 characters), every `<project name>-agent-<name>` passes the registry's name check, and the project name keeps room for the longest name · verify: `cargo test --offline --locked --test launcher_role_flags bible_ac1_`
- AC-2 (REQ-2, REQ-3): with an injected seed, the pick reaches exactly the free names and never a name whose live lock is held or whose entry has a running lease; with 47 names held, `moai cc -a --dry-run` prints the one free name; repeated unseeded picks give at least two different names · verify: `cargo test --offline --locked --test launcher_role_flags bible_ac2_`
- AC-3 (REQ-4): with every name held, `moai cc -a --dry-run` is refused with one message that names `--name` and starts no host · verify: `cargo test --offline --locked --test launcher_role_flags bible_ac3_`
- AC-4 (REQ-6): `moai cc -a --name scout --dry-run` uses `scout` as given · verify: `cargo test --offline --locked --test launcher_role_flags bible_ac4_`
- AC-5 (REQ-7): the help goldens and `docs/user-entry-draft.md` say the name is chosen at random among the free names and no live text says "first free name" · verify: `cargo test --offline --locked --test launcher_role_flags bible_ac5_`
- AC-6 (REQ-8): the decisions log carries the replacing line · verify: `cargo test --offline --locked --test launcher_role_flags bible_ac6_`
- AC-7 (REQ-2, REQ-3): lane check, the session prefix and the citation scans still pass · verify: `cargo test --offline --locked --test session_lane_check_cli`

## Scope

- write: `.moai/specs/SPEC-20261125-agent-names-bible/**`, `src/session.rs`, `src/main.rs`, `tests/launcher_role_flags.rs`, `docs/user-entry-draft.md`, `docs/decisions-2026-11-08.md`, `docs/rust-product-path.md`
- write (한 줄, line-lease): `src/help.rs`, `tests/fixtures/help/all.txt`, `tests/fixtures/help/all-60.txt`, `tests/fixtures/help/all-80.txt`, `tests/fixtures/help/all-dumb.txt`
- read-only: `src/lane_registry.rs`, `tests/help_layout.rs`, `tests/launcher_retire_role_flags.rs`, `tests/repo_claude_md.rs`, `tests/cli_text_citations.rs`, `tests/help_text_names_moai.rs`
- forbidden: `CLAUDE.md`, `skills/**`, 실제 HOME, 네트워크, Python

## Plan

1. RED: `tests/launcher_role_flags.rs` 에 `bible_ac1_` … `bible_ac6_` 를 더하고 옛 첫-빈-이름 시험을 새 뜻으로 고쳐요. 없는 함수의 컴파일 오류와 목록 시험 실패를 RED 로 기록해요.
2. GREEN: `src/session.rs` 의 `AGENT_NAMES` 를 48개로 바꾸고, 비어 있는 이름을 모아 `seed % 개수` 로 고르는 `pick_agent_name_seeded(dir, now, seed)` 를 더하고, `pick_agent_name` 은 시계·프로세스 id·세어 가는 수로 시드를 만들어 그것을 불러요(crate 추가 없음). 모두 쓰이면 `--name` 을 알리는 한 메시지예요.
3. 글: `src/main.rs` 도움말 주석, `src/help.rs` 의 한국어 도움말 한 줄(line-lease), 도움말 골든 4개 재생성(line-lease), `docs/user-entry-draft.md`, `docs/rust-product-path.md`, 결정 로그 한 줄.
4. 확인: `cargo fmt --all`, 이름 있는 시험들, 만든 바이너리로 `-a --dry-run` 을 여섯 번 돌려 이름이 갈리는 것을 보이고, `moai check affected --card` 영수증 한 번.

## Risks

- R-1: 프로젝트 이름의 한도가 53자에서 48자로 줄어요(가장 긴 이름이 `priscilla` 9자). 더 긴 이름은 이미 잘리고 있고, 같은 클론의 겹침은 잠금이 막아요.
- R-2: `src/session.rs` 를 다른 카드가 함께 예약해서 합치기가 필요할 수 있어요. 이 SPEC 의 변경은 목록 상수, 고르는 함수, 두 함수의 주석이라 겹치는 줄이 적어요.
- R-3: 같은 순간에 두 시작이 같은 이름을 고르면 나중 쪽이 독점 잠금에서 "already running" 으로 거절돼요. 다시 고르지 않아요.
- R-4: 목록이 성경 인물이라는 사실은 Rust 로 검사할 수 없어서 시험이 확정된 목록을 글자 그대로 고정해요. 목록을 바꾸면 그 시험도 함께 바뀌어요.

## Decisions

- D-1 (이름 목록): (A, 추천) 48개, 소문자 한 낱말, 코드 안의 상수: adam, eve, noah, abraham, sarah, isaac, rebekah, jacob, rachel, leah, joseph, moses, aaron, miriam, joshua, deborah, gideon, samson, ruth, naomi, boaz, hannah, samuel, david, saul, solomon, elijah, elisha, isaiah, jeremiah, ezekiel, daniel, esther, jonah, mary, martha, lazarus, peter, andrew, james, john, thomas, matthew, luke, paul, timothy, lydia, priscilla. 카드 본문의 예시 50개에서 `job` 과 `mark` 는 이름이 아니라 흔한 낱말로 읽혀서 뺐어요. (B) 24개: adam, eve, noah, sarah, isaac, jacob, rachel, leah, joseph, moses, aaron, miriam, joshua, ruth, naomi, samuel, david, elijah, daniel, esther, peter, john, paul, luke. 기본값은 A.
- D-2 (난수 원천): (A, 추천) 표준 라이브러리만: 시계 나노초, 프로세스 id, 프로세스 안에서 세어 가는 수를 섞어요. (B) 외부 crate 를 더해요. (C) 운영체제 난수를 읽어요. 기본값은 A.
- D-3 (같은 이름을 동시에 고른 두 시작): (A, 추천) 다시 고르지 않아요. 독점 잠금이 나중 쪽을 거절하고 사용자가 다시 시작해요. (B) 거절되면 다른 이름으로 다시 골라요. 기본값은 A.
- D-4 (자리표시자 철자): (A, 추천) `<영어 이름>` 을 `CLAUDE.md`, 스킬 글, 골든에 그대로 둬요(성경 이름도 영어 철자라서 어긋나는 곳이 없어요). (B) `<이름>` 으로 바꿔요: `CLAUDE.md` 카드와 스킬 글 변경이 따로 필요해요. 기본값은 A.

## Result

구현 커밋 `f859b741dd` 는 RED 커밋 `756d8aae8b` 위에 있어요. 아래는 이 실행에서 본 값이에요.

- RED: 구현 전 실행은 컴파일에 실패했어요(E0425 오류 3개, `pick_agent_name_seeded` 가 없었어요. 기록 `target/agent-names-bible/red.log`).
- AC-1: met — `cargo test --offline --locked --test launcher_role_flags` 를 슬롯에서 돌려 19개가 통과하고 0개가 실패했어요. `bible_ac1_the_list_is_pinned_and_every_name_fits_the_registry` 가 목록을 글자 그대로 고정하고 등록부 이름 검사를 통과해요. 목록은 `sort | uniq -d` 로 중복 0개, `wc -w` 로 48개, `awk length` 로 가장 긴 이름 `priscilla` 9자예요.
- AC-2: met — 같은 19개 실행에서 `bible_ac2_the_seeded_pick_reaches_exactly_the_free_names`, `bible_ac2_the_binary_takes_the_one_free_name`, `bible_ac2_unseeded_picks_vary` 가 통과해요. 만든 바이너리로 임시 git 프로젝트 `demo`(HOME 은 따로 둔 폴더)에서 `moai cc -a --dry-run` 을 여섯 번 돌리니 이름이 `demo-agent-luke`, `deborah`, `aaron`, `esther`, `aaron`, `timothy` 로 나와 서로 다른 이름이 다섯 개였어요(스크립트 `target/agent-names-bible/real-run.sh`, 출력 `target/agent-names-bible/real-run.out`).
- AC-3: met — `bible_ac3_a_used_up_list_refuses_and_names_the_name_option` 통과(같은 19개 실행). 모든 이름이 잡히면 `--name` 을 알리는 한 메시지로 거절해요.
- AC-4: met — `bible_ac4_name_is_used_as_given` 통과(같은 19개 실행).
- AC-5: met — `bible_ac5_help_and_docs_say_the_name_is_chosen_at_random` 통과(같은 19개 실행). 도움말 골든 4개는 `MOAI_HELP_GOLDEN_WRITE=1 cargo test --offline --locked --test help_layout` 로 다시 만들었고 11개가 통과해요.
- AC-6: met — `bible_ac6_the_decisions_log_carries_the_replacing_line` 통과(같은 19개 실행).
- AC-7: met — 구현 실행에서 `session_lane_check_cli` 5개가 통과하고, 같은 대상 모음의 `cli_text_citations`(5), `help_text_names_moai`(2), `launcher_retire_role_flags`(14), `retired_host_words`(9), `repo_claude_md`(27), `lane_team_names`(10) 도 모두 통과했어요.
- 남은 것: 독립 감사(tier 1)는 아직 안 했어요. 돌리지 않은 것은 clippy 와 넓은 `check affected` 모음(READY 영수증이 담아요), Windows 와 Linux 예요(Mac 에서만 돌렸어요).
