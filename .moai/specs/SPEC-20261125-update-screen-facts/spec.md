---
id: SPEC-20261125-update-screen-facts
title: "update-screen-facts"
status: draft
revision: 1
created: 2026-11-25
---

# update-screen-facts

## Intent

사용자가 설치본으로 `moai update` 를 돌렸을 때 화면이 사실과 다르게 읽혔어요. 적용할 변경이 conflict 와 보이지 않는 adopt 뿐인데 모델 질문 다섯 개가 열렸고, 답이 저장됐는지 한 줄도 없었고, 맨 위는 "Everything is up to date" 인데 그 아래에 "4 conflict" 와 "moai.md skipped" 가 있었고, 막힌 부분만 건너뛴 두 번째 실행이 "Nothing can be applied — clear the blocked part first" 로 전부 막힌 것처럼 읽혔어요. 이 SPEC 은 화면 글과 모델 단계 열림 조건만 바로잡아요. 파일을 쓰는 동작은 바꾸지 않아요.

## Non-goals

- 파일을 쓰거나 지우는 동작, 확인 토큰, 계획 digest 는 바꾸지 않아요.
- 모델 질문의 내용과 저장 방식(`update_models`)은 바꾸지 않아요.
- Codex·OMP 삭제, conflict 선택지, installed 줄, routing 행 이름은 이 SPEC 의 일이 아니에요(다른 조각이 해요).

## Requirements

- REQ-1: When a confirmed update ends having published no file, written no agents, routing, style, bin-names, hooks, extra-skills or keys part and having only adopted files or moved receipt baselines, the model step shall open only if no model setting is saved. · failure: none — 열리지 않으면 `models` 는 지금처럼 보고만 해요.
- REQ-2: When the model step ends, the Applied and the nothing-to-apply screens shall show one line for its result: saved with the settings path, left as it was, declined, cancelled or an error with its message. · failure: none — 결과가 `ok`(보고만)이면 줄이 없어요.
- REQ-3: When a host plan carries adopted files, the Checking row of that host shall name how many files are to be adopted. · failure: none — adopt 가 없으면 줄이 그대로예요.
- REQ-4: When a confirmed update adopted files of a host, the host result shall carry the number of adopted files. · failure: none — adopt 가 없으면 결과 JSON 이 그대로예요.
- REQ-5: When the plan writes nothing and any row is blocked, in conflict or skipped, the screen shall say that nothing is left to apply, that the rest is up to date and which rows are still open. · failure: none — 열린 줄이 없으면 지금처럼 "모두 최신" 한 줄이에요.
- REQ-6: When any row is blocked, in conflict or skipped, the screen shall not say that everything is up to date. · failure: none — 이 줄을 빼면 REQ-5 의 문장만 남아요.
- REQ-7: When a confirmed update adopted files of a host, the Applied row of that host shall name how many files were adopted. · failure: none — adopt 가 없으면 줄이 그대로예요.

## Acceptance

- AC-1 (REQ-1): `update::visible_change` 가 adopt 만 한 적용 결과에서 거짓이고, 파일을 하나 쓴 결과와 다른 부분(agents, routing, style, bin-names, hooks, extra-skills, keys)이 쓴 결과에서 참이며, `model_step_opens` 가 그 거짓과 저장된 모델 있음에서 닫히고 저장된 모델 없음에서 열려요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli screen_facts_ac1_`
- AC-2 (REQ-2): 모델 단계의 saved, unchanged, nothing_picked, declined, cancelled, error 결과마다 한국어와 영어 화면이 한 줄을 보여 주고 saved 줄에 저장 경로가 있고 `ok` 결과에는 줄이 없어요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli screen_facts_ac2_`
- AC-3 (REQ-3, REQ-4, REQ-7): adopt 한 임시 HOME 의 미리 보기 화면이 호스트 줄에 adopt 할 파일 수를 보여 주고 확정한 Applied 화면이 adopt 한 수를 보여 주며 결과 JSON 의 호스트에 `adopted` 수가 있어요. adopt 가 없는 호스트의 결과 JSON 에는 그 키가 없어요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli screen_facts_ac3_`
- AC-4 (REQ-5, REQ-6): 쓸 것이 없고 conflict 두 개와 skipped 한 줄이 있는 미리 보기와 막힌 호스트가 있는 미리 보기 화면에 "모두 최신" 문구도 "clear the blocked part first" 도 없고, 나머지가 최신이라는 말과 열린 줄 이름이 있어요. 열린 줄이 없으면 "모두 최신" 한 줄이에요 · verify: `cargo test --locked --offline --test update_conflict_baseline_cli screen_facts_ac4_`
- AC-5 (REQ-1…REQ-7): 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test update_tux_golden`

## Scope

- write: `.moai/specs/SPEC-20261125-update-screen-facts/`
- write: `src/update.rs`, `src/update_tux.rs`
- write: `tests/update_conflict_baseline_cli.rs`, `tests/fixtures/update_tux/*.json`, `tests/fixtures/ui/**/update-*.golden`

## Plan

1. 테스트를 먼저 써요: `visible_change`·화면 줄·adopt 수·머리말(`tests/update_conflict_baseline_cli.rs`).
2. `update.rs`: `visible_change(out)` 와 모델 단계 호출, 호스트 결과의 `adopted`.
3. `update_tux.rs`: 머리말을 행에서 계산, 모델 결과 줄, adopt 수. 골든을 다시 써요.
4. 이웃 테스트와 지킴이 뮤턴트, 카드 영수증.

## Risks

- 골든이 바뀌어요. 바뀐 줄이 이 SPEC 의 줄뿐인지 diff 로 읽어요.
- 결과 JSON 에 `adopted` 키가 늘어요(adopt 한 호스트만). 읽는 쪽이 모르는 키를 무시하는지는 이웃 테스트가 보여요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 모델 단계의 "사람이 볼 변경" 은 파일을 쓴 것(스킬 파일 published, 다른 부분이 applied 또는 쓴 partial)이에요. adopt 와 영수증 baseline 이동은 보이지 않는 변경이에요 | 사용자 실행에서 보이지 않는 adopt 뿐인데 질문 다섯 개가 열렸어요 | `visible_change` 대신 `applied` 로 돌아가요 |
| D-2 | 열린 줄은 Warn·Error 표시의 행(충돌, 막힘, 파일을 건너뛴 라우팅, 붙잡힌 대상)이고, 문구는 행 이름을 한 번씩 이어서 보여요. 순수 Skip 표시의 안내 줄(예: 이 프로젝트에 hook 없음)은 열린 줄이 아니어서 "모두 최신" 한 줄이 그대로예요 | 한 줄로 어느 줄을 볼지 말해야 하고, 안내 줄까지 열린 줄로 세면 흔한 최신 화면이 사라져요 | `open_labels` 의 표시 조건만 바꿔요 |

## Result

구현 head 는 카드의 READY 노트에 있어요(영수증도 거기에 있어요).

- AC-1 met: `screen_facts_ac1_only_a_visible_change_opens_the_model_step` 통과. `update::visible_change` 는 adopt 만 한 결과와 쓰지 않은 partial 에서 거짓, 파일을 쓴 결과와 쓴 부분에서 참이에요. 모델 단계 호출부는 `let wrote = visible_change(&out)` 한 줄이고 터미널이 있어야 열려서 테스트가 못 잡아요(not_run).
- AC-2 met: `screen_facts_ac2_the_model_step_result_has_one_line` 통과(한국어·영어, saved 줄에 설정 파일 경로, `ok` 는 줄 없음, 미리 보기 화면에도 줄이 나와요).
- AC-3 met: `screen_facts_ac3_the_adopted_files_are_counted` 통과(임시 HOME 에서 미리 보기의 "N to adopt", 확정한 결과 JSON 의 `adopted`, Applied 행의 "N adopted", adopt 없는 호스트는 키도 문구도 없음).
- AC-4 met: `screen_facts_ac4_the_headline_follows_the_rows` 통과(충돌 둘, 건너뛴 라우팅, 막힌 호스트에서 "Everything is up to date" 와 "clear the blocked part first" 가 없고 "Nothing left to apply … Still open" 이 있어요).
- AC-5 met: `cargo test --locked --offline --test update_tux_golden` 16개 통과. 골든은 `update-blocked-40` 한국어·영어 두 개만 바뀌었고 바뀐 줄은 이 SPEC 의 머리말 줄뿐이에요. `tests/update_tux_golden.rs` 의 한 줄("Nothing can be applied" → "Nothing left to apply")은 리더 결정으로 범위에 더해 고쳤어요.
- 지킴이 뮤턴트 9개(visible_change 두 곳, `adopted` 키, 머리말 분기, 열린 줄 표시 조건, 모델 결과 두 곳, adopt 수 두 곳)가 모두 지정한 테스트의 단언 실패로 죽었어요(컴파일 오류 아님, 뮤턴트별 로그 확인).
- not_run: 구현 전에 RED 를 돌리지 않았어요. 실제 HOME 의 설치본 화면은 보지 않았어요(임시 HOME 과 고정 JSON 만 썼어요).
