---
id: SPEC-20261121-language-once
title: "The conversation language is asked once at install, one list rules it everywhere, and a later command changes it"
status: draft
revision: 1
created: 2026-11-21
parent: null
---

# The conversation language is asked once at install, one list rules it everywhere, and a later command changes it

## Intent

설치 흐름 보고서 G6(2026-11-21, tip a4d4c9a164 에서 읽음): 설치 메뉴(install/install.sh:148)와 /moai project 의 user_chat(src/language.rs questions)이 대화 언어를 두 번 묻고, 설치는 fr·pt-BR 같은 아무 코드나 받는데 프로젝트는 en/ko/ja/zh 만 받으며, 저장값을 다시 바꿀 명령이 없고(set_home 을 덮어쓰지 않음), update 화면(src/ui_text.rs)과 모델 질문이 설치값을 읽지 않아요. 사용자 결정(2026-11-20T23:1xZ, 리더 결정 노트 4676)과 핀포인트 답(노트 4770 에 대한 리더 답): 대화 언어는 설치 때 한 번, 나중에 `moai language-set --home` 이 유효 코드로 교체, /moai project 는 대화를 묻지 않고 커밋·문서·instruction 만 프로젝트마다, 코드 목록은 en/ko/ja/zh 하나로 통일(moai 화면은 en/ko 만 번역 — ja/zh 면 영어 표시), update 화면과 모델 질문도 설치 언어를 따라요. 받침은 다 있어요: 홈 저장(set_home·home_conversation), 숨김 없는 질문 루프(questions 의 asked 플래그), ui_text 의 Lang(En|Ko).

## Non-goals

- ja·zh 화면 번역: 화면은 en/ko 만 번역하고 ja/zh 대화 언어는 영어 표시(리더 답) — 번역 작업은 별도.
- 모델 질문의 언어 전환 구현: 그 파일들은 진행 중인 GLM 키 카드 예약에 있어 그 착지 뒤 이어 걸어요(REQ-6 은 이 SPEC 이 고정).
- commit·문서·instruction 언어의 목록 변경: en/ko/ja/zh 유지.

## Requirements

- REQ-1: The project setup shall not ask the conversation language: it reads the value the install step saved in the HOME state folder and fills the profile's user_chat with it. · failure: none — with no saved value the setup fills en without asking.
- REQ-2: The project setup shall ask exactly the commit-messages, documents and instruction-files items per project, one question each with the saved or defaulted suggestion. · failure: none — the asked set is the ITEMS table's flags.
- REQ-3: One list rules the conversation codes everywhere: the install menu, the language-set command and the setup validation accept exactly en, ko, ja and zh. · failure: none — the list lives in the language module's CODES table and the install menu reads the same four.
- REQ-4: The language-set command with --home shall replace a stored value with an accepted code, keeping the before bytes until the publish, and shall refuse a code outside the list without touching the tree. · failure: none — a refused code leaves the file and folder as they were.
- REQ-5: The update screens shall follow the installed conversation language: ko renders Korean, and en and every code the screens do not translate (ja, zh) render English, without falling through to the locale or the project profile. · failure: none — the choice reads the HOME value once per stream.
- REQ-6: The model questions shall follow the installed conversation language the same way. · failure: none — the wiring reads the same HOME value; its GREEN waits for the files the GLM key card holds.

## Acceptance

- AC-1 (REQ-1, REQ-2): the driven setup with an installer-saved conversation value asks the three project items, none of them the conversation, and the written profile carries the saved conversation plus the three answers; a rerun keeps them.
  verify: `cargo test --locked --offline --test language_setting` — the setup scenes
- AC-2 (REQ-2, REQ-3): the question set names exactly the three project items with the four-code options, and the terminal step asks three prompts.
  verify: `cargo test --locked --offline --test language_three_questions` — the question scenes
- AC-3 (REQ-3, REQ-4): language-set stores en/ko/ja/zh as given, replaces a stored value with reason replaced, refuses fr·KO·pt-BR·es-419·kor·zh-Hant-TW·a 35-byte tag with the four-code message and an untouched tree.
  verify: `cargo test --locked --offline --test language_set_cli` — the set scenes
- AC-4 (REQ-5): with ko saved the update screens render Korean and with ja or zh saved they render English, the profile and the locale notwithstanding.
  verify: `cargo test --locked --offline --lib ui_text` — the choice scenes
- AC-5 (REQ-6): the model questions render in the saved conversation language.
  verify: the scene joins the model-question tests when the held files free up (recorded here until then)
- AC-6: the wider READY set around the change passes.
  verify: `cargo test --locked --offline --bins`
  verify: `cargo test --locked --offline --lib`

## Scope

- write: `.moai/specs/SPEC-20261121-language-once/**`, `install/install.sh`, `src/language.rs`, `src/ops/language_set.rs`, `src/ui_text.rs`, `tests/language_set_cli.rs`, `tests/language_three_questions.rs`, `tests/language_setting.rs`
- read-only: `src/model_effort_questions.rs`, `src/update_models.rs`(GLM 키 카드 예약 — 착지 뒤 REQ-6 배선에 예약 추가)
- forbidden: UI 문자의 ja·zh 번역 추가, 커밋·문서 목록 변경, 카드 번호의 출력 사용

## Plan

1. 요구 플래그와 저장 교체(language.rs·ops) — 커밋 a9eb9cfe58.
2. ui_text 의 홈값 우선 — 같은 커밋.
3. 남은 6장면 갱신(language_setting 2·language_three_questions 4)과 install 메뉴의 목록 통일.
4. READY 전체 검사 → 감사는 리더에게 → READY. REQ-6 배선은 GLM 키 카드 파일이 풀리면 같은 가지에서 이어요.

## Risks

- 화면이 ja/zh 를 영어로 표시해도 저장값은 ja/zh 그대로예요: 화면 언어와 내용 언어가 달라지는 건 리더 답의 의도(번역은 별도)라서, 문서가 그 차이를 말해요.
- 프로필의 user_chat 은 이전 설치와의 호환을 위해 남아요: 없으면 en 기본이고, 있으면 홈값이 이겨요(ui_text 의 순서).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 대화 코드 목록은 CODES(en/ko/ja/zh) 하나로 통일한다 — 화면 번역이 있는 곳까지가 목록이고, 늘어나면 이 표만 늘면 된다. | 노트 4770 답(리더), src/ui_text.rs:28 | 카드 열림 |
| D-2 | 교체는 language-set --home 만 한다 — set_home 이 유효 코드면 이전 바이트를 지켜 publish_replace 한다. | 노트 4770 답(리더) | 카드 열림 |
| D-3 | ui_text 의 순서는 MOAI_LANG → 홈 저장값 → 프로필 user_chat → 로케일 이다 — 설치값이 프로필과 로케일보다 우선하고, MOAI_LANG 은 명시적 임시 덮어쓰기로 남는다. | src/ui_text.rs choose | 카드 열림 |

## Result

(status draft — written when the slice lands)
