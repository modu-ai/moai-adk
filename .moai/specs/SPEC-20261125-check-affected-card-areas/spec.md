---
id: SPEC-20261125-check-affected-card-areas
title: "check affected --card: area sets and module filters instead of the full set"
status: draft
revision: 1
created: 2026-11-25
---

# check affected --card: area sets and module filters instead of the full set

## Intent

레인의 READY 영수증 `moai check affected --card` 는 바뀐 소스 파일의 이름 낱말로 테스트를 좁혀요. 그런데 그 좁히기가 실패하면 전체 세트로 떨어져서, 하루 검사 193번 중 좁은 세트(`selection=card`)로 끝난 것이 2번(1%)뿐이고 나머지 191번은 대상 중앙값 116개의 넓은 세트였어요. 테스트 실행이 검사 시간의 82% 이고 한 번에 중앙값 3,356개를 돌려요(근거 보고서 `~/.moai/reports/moai-adk/2026-11-25-build-test-load-rootcause.md` 3-1, 3-2절, 2026-11-25 읽음). 이 SPEC 은 전체 세트로 떨어지는 세 길을 영역 세트와 모듈 필터로 바꿔서, 한 카드의 영수증이 그 카드가 건드린 영역만 돌게 해요.

읽어서 확인한 사실(`src/check_affected.rs`, NeXT `a35a8e06a0`, 2026-11-25):

- `select_card` 는 바뀐 `src/**/*.rs` 경로마다 `module_of` 로 낱말과 경로 열쇠를 얻어요. `mod.rs`·`lib.rs`·`main.rs` 는 모듈이 없어 "no test reached" 로, 낱말을 가진 대상이 10개(`COMMON_WORD_CAP`)를 넘으면 "common name word" 로, 가진 대상이 없으면 "no test reached" 로 전체 세트가 돼요.
- 그 밖의 전체 세트 길은 `wide_class` 예요(의존성, 빌드 설정, `.moai/project.yaml`, 원장 저장 계층, 권한·훅 가드, 감사 하한 목록, git 훅, 착지 코드 `src/card/land/`, 읽을 수 없는 스캔 표). 이 SPEC 은 그 길들을 바꾸지 않아요.
- `area_filterset` 는 좁은 세트에서도 항상 `kind(lib) | kind(bin)` 을 통째로 넣어요. 마지막 전체 실행(3834개)에서 라이브러리 단위 테스트가 951개, 바이너리 단위 테스트가 28개였어요(로그를 세어 읽음). 그래서 영역 실행 파일을 좁혀도 단위 테스트 약 980개는 늘 돌아요.
- 옮겨 간 영역(`tests/<영역>/main.rs` 가 있는 영역)은 지금 `card` 하나이고, 그 실행 파일은 모듈 이름 `<모듈>::` 접두사 필터로 좁혀져요. 나머지 영역은 아직 파일마다 따로 실행 대상이에요(`tests/areas.toml`).
- `tests/areas.toml` 의 `[areas]` 는 테스트 파일 이름의 첫 낱말(밑줄 앞)을 영역에 이어요. 같은 규칙을 소스 경로의 첫 낱말에 쓰면 소스 경로의 영역이 나와요(`src/check_affected.rs` → 낱말 `check` → 영역 `spec`, `src/card/…` → `card`).

읽는 사람: 레인, 이 변경의 감사자, 착지 묶음을 도는 리더.

## Non-goals

- 넓은 세트(맨 `moai check affected`, `card land-batch` 의 검사)의 선택은 바뀌지 않아요. `--lane` 선택도 그대로예요(REQ-9).
- `wide_class` 에 든 경로는 이전처럼 전체 세트예요(REQ-7). 이 SPEC 은 그 목록을 넓히거나 줄이지 않아요.
- 영역 이름표(`tests/areas.toml`)를 바꾸거나 새 영역을 옮기지 않아요. 표는 읽기만 해요.
- 스캔(`tools/repo_scans.json`)이 고르는 대상은 지금처럼 더해져요.
- 단위 테스트를 모듈로 좁히는 것은 라이브러리(`kind(lib)`)만이에요. 바이너리의 단위 테스트(`kind(bin)`)는 28개라 통째로 둬요.

## Requirements

<!-- This SPEC adds no blocking check and weakens none: every path that is a wide path today stays one, and a card run with no way to narrow a path falls to the area, never to nothing. -->
- REQ-1: When `check affected --card` reads a changed `src/**/*.rs` path that is no wide path and whose name word is carried by more than 10 test targets, the command shall select the carrying targets that belong to the path's area instead of the full set. · failure: none — REQ-3 covers the case that no carrying target belongs to the area.
- REQ-2: When `check affected --card` reads a changed `src/**/*.rs` path that is no wide path and is a `mod.rs` or reaches no test target, the command shall select the area set of the path instead of the full set, where the area of a path is the area that `tests/areas.toml` lists for the first word of the path's top folder or file name. · failure: none — REQ-4 covers a path that has no area.
- REQ-3 (REQ-1): IF the carrying targets of a path that REQ-1 or REQ-2 reads include none of its area, THEN the command shall select the whole area set of the path, so that a module filter that picks nothing never runs nothing.
- REQ-4 (REQ-2): IF a changed source path has no area (`src/lib.rs`, `src/main.rs`, or a first word that no area lists), THEN the command shall add neither an integration test nor a full-set reason for that path, since the unit tests of REQ-5 to REQ-7 still run.
- REQ-5: When the card set holds a changed library module (a `src/**/*.rs` path other than `lib.rs`, `main.rs` and `src/ops/**`), the command shall replace the library's whole unit tests with the unit tests whose names start with the module path (`kind(lib)` and `test(/^<module path>::/)`), one prefix per changed module. · failure: none — REQ-6 keeps at least the binary's unit tests in every selection.
- REQ-6 (REQ-5): The command shall keep the binary's unit tests whole in every card selection, so that no card selection is empty.
- REQ-7 (REQ-5): IF the card set holds no changed library module, THEN the command shall keep the library's whole unit tests as before.
- REQ-8: IF a changed path is a wide path (`wide_class`), or `tests/areas.toml` is absent or unreadable while a path needs an area, THEN the command shall select the full set and name the reason, as it does for a wide path before this change.
- REQ-9: When `check affected --card` selects the card set, the summary line shall carry `selection=card` and the word `card_by=` followed by the sorted, comma-separated reasons (`area:<name>` for an area set, `module:<word>` for a name word that narrowed an area, `units:<module path>` for a library unit prefix). · failure: none — a run with no narrowed path writes `card_by=none`.
- REQ-10 (REQ-9): The answer of `check affected --card` shall carry the same reasons per changed path as `card_picks`.
- REQ-11: The bare `check affected` and `check affected --lane` shall plan the same steps, select the same targets and write the same summary line as before this change.
- REQ-12 (REQ-9): The READY gate of `card note` shall count a card receipt with a `card_by=` word as it counted a card receipt before this change. · failure: none — the gate reads `result=` and `selection=` only, and a lane receipt stays refused.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 임시 저장소의 `--card --list` 가 이름 낱말을 가진 대상이 10개를 넘는 소스 변경, 맞는 테스트가 없는 소스 변경, `mod.rs` 변경, 낱말을 가진 대상 중 영역에 속한 것이 없는 소스 변경에서 전체 세트가 아니라 그 경로의 영역 세트(와 영역 안의 이름 낱말 대상)를 골라요(`selection` 이 `card`, `card_wide` 가 null) · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(card_areas_replace_the_full_set)'`
- AC-2 (REQ-3): 모듈 필터가 영역 안에서 아무것도 고르지 못하면 그 영역 전체가 선택에 들고, 어떤 `--card` 선택도 단위 테스트와 합쳐 비지 않아요 · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(an_empty_module_filter_runs_the_whole_area)'`
- AC-3 (REQ-4): `src/lib.rs`·`src/main.rs` 와 영역 표에 없는 낱말의 소스 변경은 통합 테스트를 더하지 않고도 전체 세트로 떨어지지 않아요 · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(a_path_without_an_area_adds_no_integration_test)'`
- AC-4 (REQ-5, REQ-6, REQ-7): 라이브러리 모듈 하나만 바뀐 범위의 test 단계 필터식은 `kind(bin)` 과 `kind(lib) & test(/^<모듈 경로>::/)` 를 담고 `kind(lib)` 통째가 없어요; 바뀐 라이브러리 모듈이 없는 범위(문서나 테스트 파일만)는 `kind(lib) | kind(bin)` 통째를 그대로 담아요 · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(the_library_unit_tests_follow_the_changed_module)'`
- AC-5 (REQ-8): `Cargo.toml`, `Cargo.lock`, `.moai/project.yaml`, 원장 저장 계층, 권한·훅 가드 파일, `src/card/land/` 의 변경, 읽을 수 없는 `tests/areas.toml` 은 이전처럼 전체 세트이고 `card_wide.reason` 이 이유를 말해요 · verify: `cargo nextest run --locked --offline --test check_affected_card_select` and `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(wide_paths_keep_the_full_set)'`
- AC-6 (REQ-9, REQ-10, REQ-12): 좁은 `--card` 실행의 요약 줄이 `selection=card` 와 `card_by=` 낱말을 담고, 그 줄의 PASS 영수증을 READY 관문이 받고 `selection=lane` 영수증은 이전처럼 거절해요 · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(the_summary_line_names_why_the_set_is_narrow)'` and `cargo nextest run --locked --offline --test card -E 'test(ready_receipt)'`
- AC-7 (REQ-11): 같은 범위에서 맨 `check affected --list` 와 `--lane --list` 의 출력이 변경 전과 같아요(골든 비교) · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(the_bare_and_lane_selections_are_unchanged)'`
- AC-8 (REQ-1, REQ-2, REQ-3, REQ-4): 최근 착지 50건의 범위에 새 규칙을 적용한 표(범위마다 옛 대상 수, 새 대상 수, 새 테스트 수)를 카드 노트에 남기고 좁은 세트 비율이 80% 이상이에요 · verify: `moai check affected --card <기준>..<머리> --list` 를 50건에 되풀이해 잰 값(READY 의 Result 에 표와 비율을 적어요)

## Scope

- write: `.moai/specs/SPEC-20261125-check-affected-card-areas/**`, `src/check_affected.rs`, `src/ops/check.rs`, `tests/check_affected_card_select.rs`, `tests/check_affected_card_areas.rs`

## Risks

- 단위 테스트를 모듈로 좁히면 바뀐 모듈을 부르는 다른 모듈의 단위 테스트는 레인 영수증에서 빠져요. 그 보호는 리더가 착지 묶음마다 한 번 도는 넓은 세트가 해요(이 SPEC 은 그 관문을 바꾸지 않아요).
- 착지 코드 `src/card/land/` 는 감사 하한 성격이라 `wide_class` 에 남아요. 그 경로를 바꾸는 카드의 영수증은 계속 전체 세트예요(좁히려면 별도 결정이 필요해요).
- 영역 세트는 영역 전체라 `card` 영역은 723개예요. 그래도 전체 3834개보다 훨씬 작지만 가장 큰 영역의 변경은 여전히 무거워요.

## Plan

1. 소스 경로의 영역을 읽는 순수 함수와 `select_card` 의 세 후퇴 길을 영역 세트로 바꿔요(REQ-1 ~ REQ-4, REQ-7).
2. `area_filterset` 가 라이브러리 단위 테스트를 모듈 접두사로 좁히게 해요(REQ-5, REQ-6). 영역의 모든 모듈이 골라지면 `binary(=<영역>)` 하나로 줄여요.
3. 요약 줄과 답에 이유를 적어요(REQ-8). 맨 명령과 `--lane` 은 건드리지 않아요(REQ-9).
4. 도움말 문단(`src/ops/check.rs`)과 기존 `check_affected_card_select` 테스트의 옛 후퇴 기대를 새 규칙으로 바꿔요.
5. 최근 착지 50건으로 좁은 세트 비율을 재요(AC-8).
