---
id: SPEC-20261125-check-affected-card-areas
title: "check affected --card: area sets and module filters instead of the full set"
status: approved
revision: 1
created: 2026-11-25
approval:
  by: user
  at: 2026-11-25
  basis: "사용자가 리더 대화의 선택지에서 직접 고른 답(카드 노트 9147): \"승인, 개정 1 (Recommended)\"; 리더가 기록 직전에 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:63e2f1d22fb6f58a9b61815f513c100317ca33e780eaa0c7fbf51b2eb414a0e1"
---

# check affected --card: area sets and module filters instead of the full set

## Intent

레인의 READY 영수증 `moai check affected --card` 는 바뀐 소스 파일의 이름 낱말로 테스트를 좁혀요. 그런데 그 좁히기가 실패하면 전체 세트로 떨어져서, 하루 검사 193번 중 좁은 세트(`selection=card`)로 끝난 것이 2번(1%)뿐이고 나머지 191번은 대상 중앙값 116개의 넓은 세트였어요. 테스트 실행이 검사 시간의 82% 이고 한 번에 중앙값 3,356개를 돌려요(근거 보고서 `~/.moai/reports/moai-adk/2026-11-25-build-test-load-rootcause.md` 3-1, 3-2절, 2026-11-25 읽음). 이 SPEC 은 전체 세트로 떨어지는 세 길을 영역 세트와 모듈 필터로 바꿔서, 한 카드의 영수증이 그 카드가 건드린 영역만 돌게 해요.

읽어서 확인한 사실(`src/check_affected.rs`, NeXT `a35a8e06a0`, 2026-11-25):

- `select_card` 는 바뀐 `src/**/*.rs` 경로마다 `module_of` 로 낱말과 경로 열쇠를 얻어요. `mod.rs`·`lib.rs`·`main.rs` 는 모듈이 없어 "no test reached" 로, 낱말을 가진 대상이 10개(`COMMON_WORD_CAP`)를 넘으면 "common name word" 로, 가진 대상이 없으면 "no test reached" 로 전체 세트가 돼요.
- 그 밖의 전체 세트 길은 `wide_class` 예요(의존성, 빌드 설정, `.moai/project.yaml`, 원장 저장 계층, 권한·훅 가드, 감사 하한 목록, git 훅, 착지 코드 `src/card/land/`, 스캔 표 `tools/repo_scans.json`). 최근 착지 50건을 재어 보니 29건이 이 길로 전체 세트가 됐고, 그중 28건이 "audit floor" 였어요. 스캔 표가 19건, `src/update*` 가 11건, 착지 코드가 4건으로 가장 많아서 REQ-13·REQ-14 가 이 셋의 일부를 이름으로 좁혀요. 좁혀지는 파일은 `src/update.rs`, `src/update_models.rs`, `src/update_tux.rs`, `src/update/switch.rs`, `src/update/extra_skill.rs`, `src/home_style.rs`, `src/home_routing.rs`, `src/home_routing/rules.rs`, 착지 코드 `src/card/land/` 의 `.rs` 와 스캔 표 `tools/repo_scans.json` 뿐이에요. 접두사로 좁히지 않아요: 비밀(`src/update_keys.rs`), 훅(`src/update_hooks.rs`), 설치(`src/home_install*.rs`, `src/ops/home_install_*.rs`, `src/agent_install.rs`), HOME 이동·되돌림(`src/update_cleanup.rs`, `src/home_retire*.rs`, `src/ops/home_*_revert.rs`), 상태 폴더·HOME 경로(`src/state_root.rs`, `src/home_root.rs`, `src/home_workdir.rs`, `src/go_state.rs`)와 앞으로 더해지는 파일은 감사 하한으로 남아 전체 세트예요(사용자가 좁힘을 받아들인 조건, 아래 Interview).
- `area_filterset` 는 좁은 세트에서도 항상 `kind(lib) | kind(bin)` 을 통째로 넣어요. 마지막 전체 실행(3834개)에서 라이브러리 단위 테스트가 951개, 바이너리 단위 테스트가 28개였어요(로그를 세어 읽음). 그래서 영역 실행 파일을 좁혀도 단위 테스트 약 980개는 늘 돌아요.
- 옮겨 간 영역(`tests/<영역>/main.rs` 가 있는 영역)은 지금 `card` 하나이고, 그 실행 파일은 모듈 이름 `<모듈>::` 접두사 필터로 좁혀져요. 나머지 영역은 아직 파일마다 따로 실행 대상이에요(`tests/areas.toml`).
- `tests/areas.toml` 의 `[areas]` 는 테스트 파일 이름의 첫 낱말(밑줄 앞)을 영역에 이어요. 같은 규칙을 소스 경로의 첫 낱말에 쓰면 소스 경로의 영역이 나와요(`src/check_affected.rs` → 낱말 `check` → 영역 `spec`, `src/card/…` → `card`).

읽는 사람: 레인, 이 변경의 감사자, 착지 묶음을 도는 리더.

## Interview

- Q1: 감사 하한 경로 일부(`update`·`home`·`agent_install`·`go_state`·`state_root`, 스캔 표)의 제출 검사를 좁은 세트로 해도 되나요(REQ-13·REQ-14)? → A: 받아들여요. 사용자 선택은 "받아들이기, 묶음 넓은 검사가 안전망 (Recommended)"예요(리더 대화 2026-11-25T06:03:49Z, 리더가 이 SPEC 의 카드 note 9042 에 옮겨 적었어요). 이 답은 좁힘을 받아들인다는 뜻이고 이 SPEC 의 승인이 아니에요.
- Q2: 어떤 조건이 붙었나요? → A: 게이트·훅·설치·비밀 코드는 그대로 넓게 둬야 하고, 감사자가 관문·감사 tier 하한·묶음 넓은 검사가 그대로임을 확인한 경우에만 착지해요(note 9042 와 리더의 수리 지시 note 9043). 독립 감사 note 9040 이 처음 제출본(접두사 일곱 개)이 이 조건을 어긴다고 FAIL 을 냈고, 이 개정이 접두사를 이름 있는 경로로 바꿔 고쳐요.
- Q3: 이름 붙은 여덟 파일 목록(`src/update.rs`, `src/update_models.rs`, `src/update_tux.rs`, `src/update/switch.rs`, `src/update/extra_skill.rs`, `src/home_style.rs`, `src/home_routing.rs`, `src/home_routing/rules.rs`)을 사용자가 보고 골랐나요? → A: 아직 묻지 않았어요. 이 목록은 독립 감사 note 9040 의 "고치는 길 A" 예시와 리더 지시 note 9043 을 따른 것이고, 사용자가 이 목록 자체를 확인했다는 기록은 없어요. 그 확인은 리더가 사용자에게 물을 일이에요.

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
- REQ-13: When `check affected --card` reads a changed Rust source path that is exactly one of `src/update.rs`, `src/update_models.rs`, `src/update_tux.rs`, `src/update/switch.rs`, `src/update/extra_skill.rs`, `src/home_style.rs`, `src/home_routing.rs`, `src/home_routing/rules.rs` or is under `src/card/land/`, the command shall treat it as an ordinary source path (REQ-1 to REQ-7) instead of an audit-floor path. · failure: none — any other audit-floor path, by name or prefix (`src/update_keys.rs`, `src/update_hooks.rs`, `src/update_cleanup.rs`, `src/home_install*.rs`, `src/home_retire*.rs`, `src/home_root.rs`, `src/ops/home_*_revert.rs`, `src/agent_install.rs`, `src/go_state.rs`, `src/state_root.rs`, a file added later), stays on the audit floor and gives the full set.
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
- AC-10 (REQ-13): 이름이 붙은 여덟 파일은 각각 영역 세트를 고르고(`selection` 이 `card`), 그 옆의 비밀·훅·설치·HOME 이동·상태 폴더 파일(`src/update_keys.rs`, `src/update_hooks.rs`, `src/update_cleanup.rs`, `src/home_install.rs`, `src/home_retire.rs`, `src/home_root.rs`, `src/state_root.rs`, `src/agent_install.rs`, `src/go_state.rs`, `src/ops/home_install_apply.rs`, `src/ops/home_style_revert.rs`)은 `card_wide.reason` 이 `audit floor` 인 전체 세트이며, 이름 붙은 파일을 비밀 파일과 함께 바꿔도 비밀 파일이 전체 세트를 정해요; 단위 테스트는 이름 붙은 파일이 실제로 있고 형제 파일과 앞으로 더해질 파일이 `audit floor` 로 남는 것을 지켜요 · verify: `cargo nextest run --locked --offline --test check_affected_card_areas -E 'test(floor_files_next_to_the_narrowed_ones_stay_wide)'` and `cargo nextest run --locked --offline --lib -E 'test(the_named_floor_files_are_narrowed_and_their_siblings_stay_wide)'`

## Scope

- write: `.moai/specs/SPEC-20261125-check-affected-card-areas/**`, `src/check_affected.rs`, `src/ops/check.rs`, `src/card/ready_receipt.rs`, `tests/check_affected_card_select.rs`, `tests/check_affected_card_areas.rs`

## Risks

- 단위 테스트를 모듈로 좁히면 바뀐 모듈을 부르는 다른 모듈의 단위 테스트는 레인 영수증에서 빠져요. 그 보호는 리더가 착지 묶음마다 한 번 도는 넓은 세트가 해요(이 SPEC 은 그 관문을 바꾸지 않아요).
- 착지 코드 `src/card/land/` 와 이름 붙은 여덟 파일은 감사 하한 성격이었지만 REQ-13 이 영역 세트로 좁혀요. 그 경로를 바꾼 카드는 이제 레인 영수증이 좁아지고, 감사 하한의 전체 검사는 독립 감사와 착지 묶음의 넓은 세트가 맡아요. 되돌리려면 `NARROWED_FLOOR` 에서 이름을 빼면 돼요(되돌릴 수 있는 결정이에요). 이름 목록은 접두사가 아니라 정확한 경로라서, 옆에 새 파일이 생겨도 자동으로 좁혀지지 않아요.
- 이름 붙은 파일 중 `src/update/switch.rs`(이름 바꾸기 스크립트)와 `src/home_routing/rules.rs`(`~/.claude/rules` 를 쓰는 코드)는 HOME 쪽 파일을 만지는 코드예요. 감사 노트가 고치는 길 A 의 목록에 이 둘을 이름으로 넣어서 따랐고, 사용자가 이 목록을 직접 확인한 기록은 아직 없어요(Interview 의 Q3). 둘을 빼면 비율이 조금 낮아지는 대신 HOME 쪽 파일은 모두 넓게 남아요.
- 영역 세트는 영역 전체라 `card` 영역은 723개예요. 그래도 전체 3834개보다 훨씬 작지만 가장 큰 영역의 변경은 여전히 무거워요.

## Plan

1. 소스 경로의 영역을 읽는 순수 함수와 `select_card` 의 세 후퇴 길을 영역 세트로 바꿔요(REQ-1 ~ REQ-4, REQ-7).
2. `area_filterset` 가 라이브러리 단위 테스트를 모듈 접두사로 좁히게 해요(REQ-5, REQ-6). 영역의 모든 모듈이 골라지면 `binary(=<영역>)` 하나로 줄여요.
3. 요약 줄과 답에 이유를 적어요(REQ-8). 맨 명령과 `--lane` 은 건드리지 않아요(REQ-9).
4. 도움말 문단(`src/ops/check.rs`)과 기존 `check_affected_card_select` 테스트의 옛 후퇴 기대를 새 규칙으로 바꿔요.
5. 최근 착지 50건으로 좁은 세트 비율을 재요(AC-8).

## Result

구현 에이전트 agent-7 이 채워요(수리 revision, 2026-11-25T06:27Z 시계 기준). 감사가 다시 확인해요. SPEC revision 1, macOS arm64, 이 워크트리, 기준은 NeXT `a35a8e06a0` 위의 SPEC 커밋 `eee2cec344` 예요. 첫 제출(구현 커밋 `c69082533c`)은 독립 감사에서 FAIL 을 받았어요: 좁힘이 접두사 일곱 개라서 비밀·훅·설치·HOME 이동 코드의 이름이 비슷한 파일까지 좁혔고, 사용자가 받아들인 조건("게이트·훅·설치·비밀 코드는 그대로 넓게")과 REQ-13 의 failure 문장을 어겼어요. 이 수리는 접두사를 이름 있는 경로로 바꿔요. 이 변경은 tier 2 가 아니에요(`src/check_affected.rs` 는 `gate.rs TIER2_PREFIXES` 에 없어요).

- 수리한 것: `NARROWED_FLOOR` 가 접두사 일곱 개에서 정확한 파일 경로 여덟 개(`src/update.rs`, `src/update_models.rs`, `src/update_tux.rs`, `src/update/switch.rs`, `src/update/extra_skill.rs`, `src/home_style.rs`, `src/home_routing.rs`, `src/home_routing/rules.rs`)로 바뀌었어요. `src/card/land/` 의 좁힘(카드 본문대로)과 스캔 표(REQ-14)는 그대로예요. 이제 `src/update_keys.rs`, `src/update_hooks.rs`, `src/update_cleanup.rs`, `src/home_install*.rs`, `src/home_retire*.rs`, `src/home_root.rs`, `src/ops/home_*_revert.rs`, `src/agent_install.rs`, `src/go_state.rs`, `src/state_root.rs`와 앞으로 더해질 파일은 `audit floor` 로 남아요. SPEC 에는 좁혀지는 경로를 이름으로 적고(Intent, REQ-13, Risks), AC-10 과 `## Interview` 절을 더했어요.
- AC-1…AC-7, AC-9, AC-10 met: `tests/check_affected_card_areas.rs` 9개 통과, 단위 테스트 `the_named_floor_files_are_narrowed_and_their_siblings_stay_wide` 통과, `check_affected_card_select`·`check_affected::` 단위·`ready_receipt` 단위를 합친 68개 통과(`target/c504/repair-tests.log`).
- AC-8 not met (measured, 최근 착지 50건, `target/c504/measure-new.sh` 와 `check affected <머지>^1..<머지> --card --list`, 새 선택은 이 빌드, 옛 선택은 설치된 `moai` 4.2648.35): 수리한 head 의 좁은 세트(`selection=card`)는 39/50 = 78% 이고 AC-8 의 기준 80% 에 못 미쳐요(옛 설치본은 10/50 = 20%). 접두사 판(감사 FAIL 을 받은 head)은 41/50 = 82% 였고, 줄어든 두 건은 바뀐 파일이 `src/agent_install.rs`(설치) 한 개인 착지와 `src/go_state.rs`·`src/state_root.rs`(상태 폴더·HOME 경로)뿐인 착지예요. 사용자가 받아들인 조건상 넓게 둬야 하는 파일이라 비율을 맞추려고 풀지 않았어요. 기준 80% 를 두고 정하는 일(78% 를 받아들이거나, 사용자가 이 파일들을 좁히기로 고르거나)은 리더와 사용자 몫이에요.
- 전체 세트로 남은 11건: 게이트·훅·설치·상태 폴더·`CLAUDE.md`·`.github` 경로예요(`audit floor` 10건, `permission` 1건: `src/hook_engine.rs`). 새 선택의 테스트 수(좁은 39건): 최소 76, 중앙값 1550, 최대 4004, 전체 50건 평균 2186(옛 설치본 평균 3753). 표는 카드 노트에 있어요.
- 빈틈: 선택이 `card` 이면서 테스트가 3000개를 넘는 착지가 6건이에요(`update`·`home` 처럼 대상이 많은 영역의 변경이라 영역 전체가 골라져요). 비율(78%)은 선택 종류의 비율이고 테스트 수가 모두 줄었다는 뜻은 아니에요.
- AC-7 의 골든 비교: 수리한 빌드로 같은 50개 범위에서 맨 `check affected --list` 와 `--lane --list` 를 옛 설치본과 비교해 선택·`rust_targets`·단계(`label`, `command`)가 100번 모두 같아요(`target/c504/compare-wide.sh`, 로그 `compare-wide-repair.log`).
- 기존 테스트 `tests/check_affected_card_select.rs` 의 옛 후퇴 기대를 새 규칙으로 바꿨어요(영역 표가 없는 고정물에서 영역이 필요한 경로는 이유 `area table` 의 전체 세트, REQ-8). `src/card/ready_receipt.rs` 에는 단위 테스트 한 개만 더했고 코드는 안 바꿨어요(REQ-12).
- 단위 테스트를 모듈로 좁히면 레인 영수증에서 다른 모듈의 단위 테스트가 빠져요(Risks). 보호는 착지 묶음의 넓은 세트가 맡아요(맨 `check affected` 와 `land-batch` 는 그대로예요).
- guard-mutant: floor_files_next_to_the_narrowed_ones_stay_wide, the_named_floor_files_are_narrowed_and_their_siblings_stay_wide | 이름 목록 조회를 접두사 검사로 되돌림(`NARROWED_FLOOR.contains(&path)` → 접두사) | exit 100
- guard-mutant: floor_files_next_to_the_narrowed_ones_stay_wide, the_named_floor_files_are_narrowed_and_their_siblings_stay_wide | 비밀 파일 `src/update_keys.rs` 를 이름 목록에 넣음 | exit 100
- guard-mutant: floor_files_next_to_the_narrowed_ones_stay_wide, the_named_floor_files_are_narrowed_and_their_siblings_stay_wide | 감사 하한 검사에서 이름 목록 조건을 `true` 로(이름 붙은 파일도 넓음) | exit 100
- guard-mutant: floor_files_next_to_the_narrowed_ones_stay_wide, the_named_floor_files_are_narrowed_and_their_siblings_stay_wide | 이름 목록 조회를 `true` 로(감사 하한 전체가 좁음) | exit 100
  (셸 스크립트 `target/c504/mutants-repair.sh` 를 한 슬롯에서 돌렸고 `target/c504/mutants-repair.out` 의 마지막 줄은 `restored: identical to the start copy` 예요. 네 변형 모두 두 테스트가 함께 실패했어요.)
- 돌리지 않은 것: 전체 `cargo test`(리더 몫), Windows 와 Linux, 다른 호스트. 수리 영수증(`check affected --card`)은 이 Result 를 커밋한 뒤 한 번 내요.
