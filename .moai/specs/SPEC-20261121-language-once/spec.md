---
id: SPEC-20261121-language-once
title: "The conversation language is asked once at install, one list rules it everywhere, and a later command changes it"
status: draft
revision: 2
created: 2026-11-21
parent: null
---

# The conversation language is asked once at install, one list rules it everywhere, and a later command changes it

## Intent

설치 흐름 보고서 G6(2026-11-21, tip a4d4c9a164 에서 읽음): 설치 메뉴(install/install.sh:148)와 /moai project 의 user_chat(src/language.rs questions)이 대화 언어를 두 번 묻고, 설치는 fr·pt-BR 같은 아무 코드나 받는데 프로젝트는 en/ko/ja/zh 만 받으며, 저장값을 다시 바꿀 명령이 없고(set_home 을 덮어쓰지 않음), 모델 질문이 설치값을 읽지 않아요. 사용자 결정(2026-11-20T23:1xZ, 리더 결정 노트 4676)과 핀포인트 답(노트 4770 에 대한 리더 답): 대화 언어는 설치 때 한 번, 나중에 `moai language-set --home` 이 유효 코드로 교체, /moai project 는 대화를 묻지 않고 커밋·문서·instruction 만 프로젝트마다, 코드 목록은 en/ko/ja/zh 하나로 통일(moai 화면은 en/ko 만 번역 — ja/zh 면 영어 표시), 모델 질문도 설치 언어를 따르고 화면은 영어만 유지해요(사용자 결정 2026-11-21 — 화면 언어 전환은 철회). 받침은 다 있어요: 홈 저장(set_home·home_conversation), 숨김 없는 질문 루프(questions 의 asked 플래그). 감사(노트 5107)의 B1–B7 수리가 이 개정에 들어가요: 고정 시험 갱신, --conversation 목록 검사 복구, 설치가 기록값을 덮어쓰지 않는 기록 전용 모드, 프로필 user_chat 을 대화 근원에서 제외.

## Non-goals

- 화면 언어: 영어만 유지한다(사용자 결정 2026-11-21, #2825 유지) — update 화면의 언어 전환(구 REQ-5)은 철회했고 대화 언어는 AI 답변과 문서에만 영향을 줘요.
- ja·zh 화면 번역: 화면은 en/ko 만 번역하고 ja/zh 대화 언어는 영어 표시(리더 답) — 번역 작업은 별도.
- 모델 질문의 언어 전환 구현: 그 파일들은 진행 중인 GLM 키 카드 예약에 있어 그 착지 뒤 이어 걸어요(REQ-5 은 이 SPEC 이 고정).
- commit·문서·instruction 언어의 목록 변경: en/ko/ja/zh 유지.

## Requirements

- REQ-1: The project setup shall not ask the conversation language: it reads the value the install step saved in the HOME state folder and fills the profile's user_chat with it. · failure: none — with no saved value the setup fills en without asking.
- REQ-2: The project setup shall ask exactly the commit-messages, documents and instruction-files items per project, one question each with the saved or defaulted suggestion. · failure: none — the asked set is the ITEMS table's flags.
- REQ-3: The install menu and the language-set command and the setup validation shall accept exactly the conversation codes en, ko, ja and zh, from the one list the language module holds and the install menu reads too. · failure: none — the list lives in the language module's CODES table and the install menu reads the same four.
- REQ-4: The language-set command with --home shall replace a stored value with an accepted code, keeping the before bytes until the publish, and refuse a code outside the list without touching the tree. · failure: none — a refused code leaves the file and folder as they were; the installer's record-once mode keeps a recorded value so a reinstall never overwrites it.
- REQ-5: The model questions shall follow the installed conversation language the same way, once the files they live in join this slice. · failure: none — the wiring reads the same HOME value.

## Acceptance

- AC-1 (REQ-1, REQ-2): the driven setup with an installer-saved conversation value asks the three project items, none of them the conversation, and the written profile carries the saved conversation plus the three answers; a rerun keeps them.
  verify: `cargo test --locked --offline --test language_setting` — the setup scenes
- AC-2 (REQ-2, REQ-3): the question set names exactly the three project items with the four-code options, and the terminal step asks three prompts.
  verify: `cargo test --locked --offline --test language_three_questions` — the question scenes
- AC-3 (REQ-3, REQ-4): language-set stores en/ko/ja/zh as given, replaces a stored value with reason replaced, refuses fr·KO·pt-BR·es-419·kor·zh-Hant-TW·a 35-byte tag with the four-code message and an untouched tree, and the record-once mode keeps a recorded value.
  verify: `cargo test --locked --offline --test language_set_cli` — the set scenes
- AC-4 (REQ-5): the model questions render in the saved conversation language.
  verify: the scene joins the model-question tests when the held files free up (recorded here until then)
- AC-5: the wider READY set around the change passes.
  verify: `cargo test --locked --offline --bins`
  verify: `cargo test --locked --offline --lib`

## Scope

- write: `.moai/specs/SPEC-20261121-language-once/**`, `install/install.sh`, `src/language.rs`, `src/ops/language_set.rs`, `tests/language_set_cli.rs`, `tests/language_three_questions.rs`, `tests/language_setting.rs`
- read-only: `src/model_effort_questions.rs`, `src/update_models.rs`(GLM 키 카드 예약 — 착지 뒤 REQ-5 배선에 예약 추가)
- forbidden: UI 문자의 ja·zh 번역 추가, 커밋·문서 목록 변경, 카드 번호의 출력 사용

## Plan

1. 요구 플래그와 저장 교체(language.rs·ops) — 커밋 a9eb9cfe58.
2. 남은 6장면 갱신(language_setting 2·language_three_questions 4)과 install 메뉴의 목록 통일 — 커밋 59f0a11c8f.
3. 수리(감사 B1–B7, 노트 5107): ui_text 변경 철회와 REQ-5 제거, --conversation 의 목록 검사 복구, language-set 의 기록 전용 모드(--if-absent), 프로필 user_chat 을 대화 근원에서 제외, 고정 시험 갱신(go_settings 2·install_script_test 언어부), Result 작성.
4. READY 전체 검사 → 감사는 리더에게 → READY. REQ-5(모델 질문) 배선은 GLM 키 카드 파일이 풀리면 같은 가지에서 이어요. install.sh·install.ps1 의 언어 조각(메뉴·valid 검사·record 호출)은 설치 스크립트 예약 조율 뒤 같은 가지에서.

## Risks

- 화면이 ja/zh 를 영어로 표시해도 저장값은 ja/zh 그대로예요: 화면 언어와 내용 언어가 달라지는 건 리더 답의 의도(번역은 별도)라서, 문서가 그 차이를 말해요.
- 프로필의 user_chat 은 대화 언어의 근원이 아니에요: HOME 기록이 없는 이전 프로젝트는 재실행 때 해석값(en)이 user_chat 에 한 번 기록되고, 저장된 물품 답은 그대로 남아요 — `language-set --home` 으로 대화 언어를 되살릴 수 있어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 대화 코드 목록은 CODES(en/ko/ja/zh) 하나로 통일한다 — 화면 번역이 있는 곳까지가 목록이고, 늘어나면 이 표만 늘면 된다. | 노트 4770 답(리더), src/language.rs CODES | 카드 열림 |
| D-2 | 교체는 벌거벗은 language-set --home 만 한다 — set_home 이 유효 코드면 이전 바이트를 지켜 publish_replace 하고, 설치는 기록 전용 모드(--if-absent)로 넘겨 기록값을 덮어쓰지 않는다. | 노트 4770 답(리더), 노트 5107 B3 | 카드 열림 |
| D-3 | 화면은 영어만 유지한다 — update 화면의 언어 전환은 철회하고 대화 언어는 AI 답변과 문서에만 영향을 준다. | 사용자 결정 2026-11-21(#2825 유지, 노트 5107 B5) | 카드 열림 |
| D-4 | 프로필의 user_chat 은 대화 언어의 근원이 아니다 — 순서는 --conversation → HOME 기록값 → Go 백업 기록값 → en 이고, setup 은 해석값을 user_chat 에 기록하며 세션 요약은 HOME 기록값만 읽는다. | 노트 5107 B6 결정(리더) | 카드 열림 |

## Result

- AC-1: met — head 6780397d6b 에서 `cargo test --locked --offline --test language_setting` exit 0 (9 passed): 설치가 저장한 값으로 세 물음만 묻고 user_chat 을 채우며, 기록된 HOME 이 있는 재실행은 모든 값을 지켜요.
- AC-2: met — `cargo test --locked --offline --test language_three_questions` exit 0 (12 passed): 정확히 세 물음·네 코드 옵션·터미널 세 프롬프트.
- AC-3: met — `cargo test --locked --offline --test language_set_cli` exit 0 (19 passed): 네 코드 그대로 저장, 교체는 reason replaced, 목록 밖 코드는 트리 무변화 거절, --if-absent 는 기록값을 지켜요.
- AC-4: withdrawn — 모델 질문 배선은 다른 예약이 쓰는 파일에 있어 이 슬라이스에서 빠져요(AC-4 의 verify 줄이 기록). REQ-5 가 계약을 유지하고 파일이 풀리면 같은 가지에서 이어져요.
- AC-5: 부분 — `cargo test --locked --offline --lib` exit 0 (1542 passed), `--bins` exit 0 (28 passed), `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0, 언어 네 대상(go_settings_defaults 4·language_set_cli 19·language_setting 9·language_three_questions 12) 전부 통과, `git diff --check`·`cargo fmt --check`·ops rustfmt·`sh -n install.sh` 0. 아직 남은 것: tests/install_script_test.py 의 ps1클래스 메뉴 핀 7개가 install.ps1 의 네 항목 메뉴 수정을 기다려요(카드 조율 노트 참조) — 수정 착지 뒤 이 줄을 최종 숫자로 갱신하고 `moai check affected` 를 push 전에 돌려요.

(status draft — 승인과 completed 는 병합 절차가 기록해요)
