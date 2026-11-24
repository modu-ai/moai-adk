---
id: SPEC-20261121-glm-key-at-pick
title: "Picking glm in the model step asks for the key right there: hidden, vaulted, skippable"
status: approved
revision: 1
created: 2026-11-20
parent: null
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8327 (2026-11-24T07:39:10Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단이고 사용자 직접 선택은 아님, 노트 8326 에 답함: 판 1 승인, head 39ed0d9fb4); 기능 자체는 사용자 결정 2026-11-20 (노트 4678); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:746e8acf504074aee716940f85028859a3f92a39d8904b5193d69f3938e3c579"
---

# Picking glm in the model step asks for the key right there: hidden, vaulted, skippable

## Intent

설치 흐름 보고서 G8(2026-11-21, 로컬 tip a4d4c9a164 에서 읽음): 모델 단계는 glm 을 제안하지(src/model_effort_questions.rs TOOL_CHOICES) 키를 묻는 곳이 없어서, 첫 `moai glm` 실행이 key_absent 로 막혀요(src/session.rs). 사용자 결정(2026-11-20T23:1xZ, 리더 결정 노트 4678로 기록): glm 을 고르면 그 자리에서 키를 물어요 — 숨김 입력, 키 보관소 저장, 세션·argv·출력에 흔적 없음, 건너뛰기 가능. 보관소에 키가 이미 있으면 '그대로 쓰기(기본) / 새로 입력'을 물어요. 받침은 다 있어요: 숨김 입력은 key_prompt 의 read_hidden, 저장은 key set 명령이 지나는 보관소 경로, 존재 확인은 key_vault 의 read_entry.

## Interview

- Q1: glm 을 고르면 키를 어디서 어떻게 받나요? → A: 모델 단계에서 그 자리에서 숨김 입력으로 받고 키 보관소에 저장해요. 세션·argv·출력에 안 남고 건너뛸 수 있어요(사용자 결정 2026-11-20, 리더 결정 노트 4678).
- Q2: 보관소에 키가 이미 있으면요? → A: '그대로 쓰기(기본) / 새로 입력'을 물어요(노트 4678).
- Q3: 구현이 `src/update_models.rs` 도 건드려도 되나요? → A: 예, 범위에 더해요(리더 결정 노트 8327, 2026-11-24).
- Q4: 시험은 어떻게 하나요? → A: 임시 HOME 과 가짜 키만 쓰고 진짜 보관소 경로는 시험에서 열지 않아요. 키 값이 출력·세션·argv·로그·시험 스냅숏 어디에도 안 남음을 시험이 고정해요(노트 8327).

## Non-goals

- 다른 제공자(claude·codex)의 키 질문: 이 조각은 glm 만.
- 키 보관소의 저장 방식 변경: key set 이 쓰는 경로를 그대로 재사용해요.
- 첫 `moai glm` 실행의 key_absent 동작 변경: 건너뛴 사용자는 오늘과 같이 실행 때 안내를 받아요.

## Requirements

- REQ-1: When the model step's pick names glm, the step shall ask for the key before the launcher record is written, and the question says the input is hidden and may be skipped. · failure: none — the question adds no new flow step outside the model step.
- REQ-2: The key question shall read the answer through the hidden input path, so the typed value appears in no echo, no session record and no output. · failure: none — the reading is the existing hidden reader.
- REQ-3: The asked key shall be stored in the key vault through the same path the key set command uses, and nowhere else. · failure: none — the storage call is the one the key set command makes.
- REQ-4: When the vault already holds a usable glm key, the step shall ask between keeping the existing key, the default, and entering a new one. · failure: none — keeping needs no input of the value.
- REQ-5: An empty answer shall skip the key entry, the step still records the glm pick, and the first glm run keeps asking for the key. · failure: none — the skip writes nothing to the vault.
- REQ-6: When the model step's pick names another provider, the step shall ask nothing about a key. · failure: none — the question is tied to the glm pick only.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-6): with a glm pick the driven interview shows the key question once with the hidden-and-skippable words, a claude pick shows none, and no answer bytes appear in any output.
  verify: `cargo test --locked --offline --test model_effort_questions` — the question scenes
- AC-2 (REQ-3, REQ-4, REQ-5): a typed key lands in the vault readable by the key set path, a keep answer stores nothing new, an empty answer records the pick without a vault write.
  verify: `cargo test --locked --offline --test model_effort_questions` — the vault scenes
- AC-3: the wider READY set around the change passes.
  verify: `cargo test --locked --offline --bins`
  verify: `cargo test --locked --offline --lib`

## Scope

- write: `.moai/specs/SPEC-20261121-glm-key-at-pick/**`, `src/model_effort_questions.rs`, `src/key_prompt.rs`, `tests/model_effort_questions.rs`, 그리고 질문이 앉는 흐름 파일(구현에서 이름이 정해지면 예약에 추가)
- read-only: `src/key_vault.rs`, `src/ops/key_set.rs`, `src/session.rs`
- forbidden: 키 값의 출력·세션 기록, 보관소 경로 변경, 다른 제공자 질문 추가, 카드 번호의 출력 사용

## Plan

1. 상호작용 계층 확정: model_effort_questions 를 부르는 흐름에서 glm 픽 뒤에 질문을 앉힌다(파일이 정해지면 예약에 추가).
2. RED: 구동 하니스 시험이 질문·숨김·보관소 저장·건너뛰기·기존 키 선택을 먼저 요구해요.
3. GREEN: read_hidden 재사용 + key set 경로 저장 + 기존 키 분기.
4. READY 전체 검사 → 감사는 리더에게 → READY.

## Risks

- 숨김 입력이 터미널이 없는 환경(파이프)에서 오면 read_hidden 의 거절이 흐름을 멈출 수 있어요: 건너뛰기 경로가 같은 조건에서 열려 흐름은 계속되도록 해요.
- 기존 키의 '그대로 쓰기'는 값을 보여 주지 않아요: 질문은 존재 여부만 말해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 숨김 입력과 저장은 기존 read_hidden 과 key set 경로를 재사용한다 — 새 입력·저장 기계를 만들지 않는다. | src/key_prompt.rs:132, src/ops/key_set.rs:77 | 카드 열림 |

## Result

(status draft — written when the slice lands)
