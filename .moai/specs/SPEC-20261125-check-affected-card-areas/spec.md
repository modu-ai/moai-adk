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
- 그 밖의 전체 세트 길은 `wide_class` 예요(의존성, 빌드 설정, `.moai/project.yaml`, 원장 저장 계층, 권한·훅 가드, 감사 하한 목록, git 훅, 착지 코드 `src/card/land/`, 스캔 표 `tools/repo_scans.json`). 최근 착지 50건을 재어 보니 29건이 이 길로 전체 세트가 됐고, 그중 28건이 "audit floor" 였어요. 스캔 표가 19건, `src/update*` 가 11건, 착지 코드가 4건으로 가장 많아서 REQ-13·REQ-14 가 이 셋을 좁혀요. 게이트·훅·비밀·설치·이주 코드는 그대로 넓어요.
- `area_filterset` 는 좁은 세트에서도 항상 `kind(lib) | kind(bin)` 을 통째로 넣어요. 마지막 전체 실행(3834개)에서 라이브러리 단위 테스트가 951개, 바이너리 단위 테스트가 28개였어요(로그를 세어 읽음). 그래서 영역 실행 파일을 좁혀도 단위 테스트 약 980개는 늘 돌아요.
- 옮겨 간 영역(`tests/<영역>/main.rs` 가 있는 영역)은 지금 `card` 하나이고, 그 실행 파일은 모듈 이름 `<모듈>::` 접두사 필터로 좁혀져요. 나머지 영역은 아직 파일마다 따로 실행 대상이에요(`tests/areas.toml`).
- `tests/areas.toml` 의 `[areas]` 는 테스트 파일 이름의 첫 낱말(밑줄 앞)을 영역에 이어요. 같은 규칙을 소스 경로의 첫 낱말에 쓰면 소스 경로의 영역이 나와요(`src/check_affected.rs` → 낱말 `check` → 영역 `spec`, `src/card/…` → `card`).

읽는 사람: 레인, 이 변경의 감사자, 착지 묶음을 도는 리더.

## Non-goals

- 넓은 세트(맨 `moai check affected`, `card land-batch` 의 검사)의 선택은 바뀌지 않아요. `--lane` 선택도 그대로예요(REQ-9).
- `wide_class` 에 든 경로는 이전처럼 전체 세트예요(REQ-8). 이 SPEC 이 줄이는 것은 REQ-13·REQ-14 가 이름 붙인 경로뿐이고 목록을 넓히지는 않아요.
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
- REQ-9: When `check affected --card` selects the card set, the summary line shall carry `selection=card` and the word `card_by=` followed by the sorted, comma-separated reasons (`area:<name>` for an area set, `module:<word>` for a name word that picked targets, `units:<module path>` for a library unit prefix). · failure: none — a run with no narrowed path writes `card_by=none`.
- REQ-10 (REQ-9): The answer of `check affected --card` shall carry the same reasons per changed path as `card_picks`.
- REQ-11: The bare `check affected` and `check affected --lane` shall plan the same steps, select the same targets and write the same summary line as before this change.
- REQ-12 (REQ-9): The READY gate of `card note` shall count a card receipt with a `card_by=` word as it counted a card receipt before this change. · failure: none — the gate reads `result=` and `selection=` only, and a lane receipt stays refused.
- REQ-13: When `check affected --card` reads a changed Rust source path under `src/update`, `src/home`, `src/ops/home`, `src/ops/agents_home`, `src/agent_install`, `src/go_state`, `src/state_root` or `src/card/land/`, the command shall treat it as an ordinary source path (REQ-1 to REQ-7) instead of an audit-floor path. · failure: none — the gate, hook, secret, install and migration paths of the floor stay wide.
- REQ-14: When `check affected --card` reads a changed `tools/repo_scans.json`, the command shall select the test targets whose source names that path, name the reason `pins:<path>`, and not fall to the full set. · failure: none — the wide check of a landing batch still runs every scan.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 임시 저장소의 `--card --list` 가 이름 낱말을 가진 대상이 10개를 넘는 소스 변경, 맞는 테스트가 없는 소스 변경, `mod.rs` 변경, 낱말을 가진 대상 중 영역에 속한 것이 없는 소스 변경에서 전체 세트가 아니라 그 경로의 영역 세트(와 영역 안의 이름 낱말 대상)를 골라요(`selection` 이 `card`, `card_wide` 가 null) · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(card_areas_replace_the_full_set)'`
- AC-2 (REQ-3): 모듈 필터가 영역 안에서 아무것도 고르지 못하면 그 영역 전체가 선택에 들고, 어떤 `--card` 선택도 단위 테스트와 합쳐 비지 않아요 · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(an_empty_module_filter_runs_the_whole_area)'`
- AC-3 (REQ-4): `src/lib.rs`·`src/main.rs` 와 영역 표에 없는 낱말의 소스 변경은 통합 테스트를 더하지 않고도 전체 세트로 떨어지지 않아요 · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(a_path_without_an_area_adds_no_integration_test)'`
- AC-4 (REQ-5, REQ-6, REQ-7): 라이브러리 모듈 하나만 바뀐 범위의 test 단계 필터식은 `kind(bin)` 과 `kind(lib) & test(/^<모듈 경로>::/)` 를 담고 `kind(lib)` 통째가 없어요; 바뀐 라이브러리 모듈이 없는 범위(문서나 테스트 파일만)는 `kind(lib) | kind(bin)` 통째를 그대로 담아요 · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(the_library_unit_tests_follow_the_changed_module)'`
- AC-5 (REQ-8): `Cargo.toml`, `Cargo.lock`, `.moai/project.yaml`, 원장 저장 계층, 권한·훅 가드 파일, 감사 하한의 게이트·설치 코드(`src/gate.rs`, `src/self_update.rs`), 읽을 수 없는 `tests/areas.toml` 은 이전처럼 전체 세트이고 `card_wide.reason` 이 이유를 말해요 · verify: `cargo nextest run --locked --offline --test check_affected_card_select` and `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(wide_paths_keep_the_full_set)'`
- AC-6 (REQ-9, REQ-10, REQ-12): 좁은 `--card` 실행의 요약 줄이 `selection=card` 와 `card_by=` 낱말을 담고, 그 줄의 PASS 영수증을 READY 관문이 받고 `selection=lane` 영수증은 이전처럼 거절해요 · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(the_summary_line_names_why_the_set_is_narrow)'` and `cargo nextest run --locked --offline --test card -E 'test(ready_receipt)'`
- AC-7 (REQ-11): 같은 범위에서 맨 `check affected --list` 와 `--lane --list` 의 출력이 변경 전과 같아요(골든 비교) · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(the_bare_and_lane_selections_are_unchanged)'`
- AC-8 (REQ-1, REQ-2, REQ-3, REQ-4): 최근 착지 50건의 범위에 새 규칙을 적용한 표(범위마다 옛 대상 수, 새 대상 수, 새 테스트 수)를 카드 노트에 남기고 좁은 세트 비율이 80% 이상이에요 · verify: `moai check affected --card <기준>..<머리> --list` 를 50건에 되풀이해 잰 값(READY 의 Result 에 표와 비율을 적어요)

- AC-9 (REQ-13, REQ-14): `src/update_models.rs`, `src/home_style.rs`, `src/card/land/check.rs` 의 변경은 전체 세트가 아니라 자기 영역 세트를 고르고, `tools/repo_scans.json` 의 변경은 그 경로를 이름으로 쓰는 테스트만 `pins:` 이유로 고르며, `src/gate.rs` 는 `src/update_models.rs` 와 함께 바뀌어도 전체 세트예요 · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(narrowed_floor_paths_take_their_area_and_the_scan_table_pins_its_tests)'`

## Scope

- write: `.moai/specs/SPEC-20261125-check-affected-card-areas/**`, `src/check_affected.rs`, `src/ops/check.rs`, `src/card/ready_receipt.rs`, `tests/check_affected_card_select.rs`, `tests/check_affected_card_areas.rs`

## Risks

- 단위 테스트를 모듈로 좁히면 바뀐 모듈을 부르는 다른 모듈의 단위 테스트는 레인 영수증에서 빠져요. 그 보호는 리더가 착지 묶음마다 한 번 도는 넓은 세트가 해요(이 SPEC 은 그 관문을 바꾸지 않아요).
- 착지 코드 `src/card/land/` 와 `src/update*`·`src/home*` 는 감사 하한 성격이었지만 REQ-13 이 영역 세트로 좁혀요. 그 경로를 바꾼 카드는 이제 레인 영수증이 좁아지고, 감사 하한의 전체 검사는 독립 감사와 착지 묶음의 넓은 세트가 맡아요. 되돌리려면 `NARROWED_FLOOR` 에서 접두사를 빼면 돼요(되돌릴 수 있는 결정이에요).
- 영역 세트는 영역 전체라 `card` 영역은 723개예요. 그래도 전체 3834개보다 훨씬 작지만 가장 큰 영역의 변경은 여전히 무거워요.

## Plan

1. 소스 경로의 영역을 읽는 순수 함수와 `select_card` 의 세 후퇴 길을 영역 세트로 바꿔요(REQ-1 ~ REQ-4, REQ-7).
2. `area_filterset` 가 라이브러리 단위 테스트를 모듈 접두사로 좁히게 해요(REQ-5, REQ-6). 영역의 모든 모듈이 골라지면 `binary(=<영역>)` 하나로 줄여요.
3. 요약 줄과 답에 이유를 적어요(REQ-8). 맨 명령과 `--lane` 은 건드리지 않아요(REQ-9).
4. 도움말 문단(`src/ops/check.rs`)과 기존 `check_affected_card_select` 테스트의 옛 후퇴 기대를 새 규칙으로 바꿔요.
5. 최근 착지 50건으로 좁은 세트 비율을 재요(AC-8).

## Result

구현 에이전트 agent-7 이 채워요(2026-11-25T05:53Z 시계 기준). 감사가 다시 확인해요. SPEC revision 1, macOS arm64, 이 워크트리, 기준은 NeXT `a35a8e06a0` 위의 SPEC 커밋 `eee2cec344` 예요. 이 변경은 tier 2 가 아니에요(`src/check_affected.rs` 는 `gate.rs TIER2_PREFIXES` 에 없어요).

- AC-1…AC-7, AC-9 met: `tests/check_affected_card_areas.rs` 8개 통과, `check_affected_card_select`·`check_affected::` 단위 테스트·`ready_receipt` 단위 테스트를 합친 66개 통과(`target/c504/new-tests3.log`), 표 검사 5개와 `check_affected_*` 이웃을 합친 157개 통과(`target/c504/neighbours.log`).
- AC-8 met (measured, 최근 착지 50건, `target/c504/measure.sh` 와 `moai check affected <머지>^1..<머지> --card --list`, 옛 선택은 설치된 `moai` 4.2648.35, 새 선택은 이 빌드): 좁은 세트(`selection=card`) 41/50 = 82%, 옛 설치본은 10/50 = 20%예요. 새 선택의 테스트 수 평균은 2130(옛 평균 3753, 같은 방식으로 `cargo nextest list` 로 센 값), 중앙값 1697, 최대 4002이에요. 표 전체는 카드 노트에 있어요(`target/c504/table.md`).
- 1차 측정(감사 하한 좁히기 전)은 21/50 = 42% 였고, 남은 29건 중 28건이 "audit floor" 였어요(스캔 표 19건, `src/update*` 11건, 착지 코드 4건이 가장 많았어요). 그래서 REQ-13·REQ-14 를 더했어요. 되돌릴 수 있는 결정이에요(`NARROWED_FLOOR` 의 접두사를 빼면 돼요).
- 아직 전체 세트인 9건은 게이트·훅·설치·`CLAUDE.md`·`.github` 경로예요(`src/gate.rs`, `src/hook_engine.rs`, `install/*`, `CLAUDE.md`, `agents/audit.md`, `.github/workflows/*`). 이 SPEC 이 넓게 두기로 한 경로예요.
- 빈틈: 선택이 `card` 이면서 테스트가 3000개를 넘는 착지가 6건이에요(`update`·`home` 처럼 대상이 많은 영역의 변경이라 영역 전체가 골라져요). 비율(82%)은 선택 종류의 비율이고 테스트 수가 모두 줄었다는 뜻은 아니에요. 영역 표를 더 잘게 나누는 일은 이 SPEC 밖이에요(`tests/areas.toml` 은 읽기만 해요).
- AC-7 의 골든 비교: 같은 50개 범위에서 맨 `check affected --list` 와 `--lane --list` 를 옛 설치본과 이 빌드로 돌려 선택·`rust_targets`·단계(`label`, `command`)를 비교했어요(`target/c504/compare-wide.sh`): 100번 비교에서 차이 0.
- 기존 테스트 `tests/check_affected_card_select.rs` 의 옛 후퇴 기대를 새 규칙으로 바꿨어요: 영역 표가 없는 고정물에서 `src/ops/card.rs`·`src/ledger/mod.rs`·`src/lonely.rs`·`src/card/land/check.rs` 는 이유 `area table` 로 전체 세트가 돼요(표 없이는 영역을 못 읽기 때문이에요, REQ-8).
- `src/card/ready_receipt.rs`: 단위 테스트 한 개만 더했어요(`a_card_summary_line_with_its_reasons_is_a_receipt_and_a_lane_one_is_not`). 코드는 바꾸지 않았어요(REQ-12).
- 단위 테스트를 모듈로 좁히면 레인 영수증에서 다른 모듈의 단위 테스트가 빠져요(Risks). 보호는 착지 묶음의 넓은 세트가 맡아요(맨 `check affected` 와 `land-batch` 는 그대로예요).
- guard-mutant: none | tier 2 변경이 아니라 가드 뮤턴트를 묻지 않아요.
- 돌리지 않은 것: 전체 `cargo test`(리더 몫), Windows 와 Linux, 다른 호스트. READY 영수증(`check affected --card`)은 이 Result 를 커밋한 뒤 한 번 내요.
