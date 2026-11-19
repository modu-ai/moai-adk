---
id: SPEC-20261119-models-revert-reporting
title: "models-revert slice 6: marker before revert, honest failure, refusal kinds"
status: draft
revision: 1
created: 2026-11-19
---

# models-revert slice 6: marker before revert, honest failure, refusal kinds

## Intent

C-47 조각 5 재감사 후속(카드 노트 2688의 (2)(3)(4)): 지금 `models revert` 는 설정 파일을 되돌린 뒤에야
되돌림 표식(marker)을 쓰고, 표식 쓰기가 실패하면 이미 되돌린 상태로 internal 오류로 끝나며 재시도는
`changed_since_apply` 로 skipped 돼요 — 되돌림은 일어났는데 영원히 완료 기록이 없어요. 이 SPEC 은 (a) 표식을
되돌림보다 먼저 써서 실패 창을 없애고, 실패해도 "아무것도 바뀌지 않았다"고 정확히 말하며, (b) 표식이 있는데
파일이 아직 after 상태면(이전 실행이 표식 뒤에 죽은 창) 재시도가 거절 대신 되돌림을 이어가게 하고, (c) 모든
거절이 envelope 에서 internal 이 아니라 제 kind 로 분류되게 해요 — 메시지를 기존 분류 표 행에 맞춰
`src/contract.rs` 편집 없이요. 문서(operations)도 거절 목록·종료 코드·재개 규칙을 얻어요.

## Non-goals

- `skills/moai/workflows/update.md:7` 문구: source-manifest.json 이 C-103 예약 안이라 이 조각에서 빼요
  (카드 노트 2740).
- (1) 상태 폴더(~/.moai-cli 자체)가 링크일 때의 revert·save 동작: 사용자 결정 대기.
- `src/contract.rs`·`src/ops/models_revert.rs` 편집 없음: kind 은 기존 분류 행으로 얻어요.
- 영수증 폴더가 링크일 때의 쓰기(조각 5 감사 N2): 후속.

## Requirements

- REQ-1: When the revert marker write fails, the command shall refuse before touching the settings file with a refusal that says nothing was changed. · failure: none
- REQ-2: The command shall finish the revert when the marker is present and the settings file still hashes to the receipt's after-value, which is the window of a run that died after writing the marker.
- REQ-3: The already-reverted refusal shall stay for a marker whose settings file no longer hashes to the receipt's after-value.
- REQ-4: The refusal messages shall classify, through the existing message table, as: a malformed receipt id invalid_input, a missing receipt not_found, a foreign receipt and a read-back mismatch conflict, a damaged receipt invalid_input, a permission reference refusal permission_reference_missing, and a marker write failure io (retryable).
- REQ-5: When `docs/design-review/appendix/operations/models_revert.md` is read, it shall list every refusal with its kind and exit code, and state the marker-before-revert order and the resume rule. · failure: none

## Acceptance

- AC-1 (REQ-1): 표식 경로를 디렉터리로 미리 만들어 쓰기를 막으면 revert 는 설정 파일을 그대로 둔 채 "nothing was changed" 거절로 끝나요.
  verify: `cargo test --offline --locked --test launcher_settings_receipt`
- AC-2 (REQ-2, REQ-3): 표식만 있고 파일이 after 해시면 revert 는 Ok(reverted) 로 끝나요; 표식이 있고 파일이 바뀌었으면 여전히 already-reverted 거절이에요.
  verify: `cargo test --offline --locked --test launcher_settings_receipt`
- AC-3 (REQ-4): CLI envelope 의 kind 가 표와 같아요(잘못된 id invalid_input, 없는 영수증 not_found, 이미 되돌림 conflict, 손상 영수증 invalid_input, 권한 참조 누락 permission_reference_missing).
  verify: `cargo test --offline --locked --test launcher_settings_receipt`
- AC-4 (REQ-5): operations 문서가 거절·kind·종료 코드·재개 규칙을 이름해요.
  verify: `cargo test --offline --locked --test launcher_settings_receipt` (문서 행 포함)
- AC-5 (REQ-1…REQ-5): 이웃 — 기존 수신 거절 시험 전부와 fmt.
  verify: `cargo test --offline --locked --test launcher_settings_receipt`; `cargo fmt --check`

## Scope

- write: `src/update_models.rs`, `tests/launcher_settings_receipt.rs`, `docs/design-review/appendix/operations/models_revert.md`, `.moai/specs/SPEC-20261119-models-revert-reporting/spec.md`
- read-only: `src/key_vault.rs`, `src/contract.rs`, `src/ops/models_revert.rs`, `skills/moai/workflows/update.md`
- forbidden: `skills/**` 편집(manifest 이 C-103 예약), `src/contract.rs`·`src/ops/**` 편집, 실제 HOME·원장

## Plan

1. 시험 먼저: 위 RED 시험을 더하고 현재 코드에서 실패를 읽어요(재개·마커-선쓰기·kind 3계).
2. `revert` 순서 재배치(표식 선쓰기·재개 분기)와 거절 문구를 분류 행에 맞춰 고쳐요.
3. operations 문서를 갱신해요. 이웃 검사와 배터리(lib 전체·scan·fmt·clippy --all-targets·trailer-check)를 돌려요.

## Risks

- 표식 선쓰기 뒤 프로세스가 죽으면 다음 시도가 재개해요 — 재개 분기의 해시 조건이 틀리면 이미 되돌린 파일을
  다시 되돌릴 수 있어요: 재개는 after 해시 일치일 때만(REQ-2)이라 시험이 양쪽을 고정해요.
- 문구 변경이 기존 시험의 includes 문자열을 깰 수 있어요 — 이웃 시험이 잡아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | kind 은 메시지 문구를 기존 분류 행에 맞춰 얻음(표 편집 없음) | 이 카드 범위에 src/contract.rs 가 없음; 행은 이미 존재 | 후속에서 OpError 로 전환 |
| D-2 | 표식 선쓰기 + 재개 분기 | 카드 노트 2688 (2)의 두 대안 중 순서 교체가 재시도까지 완결 | 이전 순서 복원 |
| D-3 | (5) update.md 문구는 이 조각에서 제외 | manifest 이 C-103 예약(카드 노트 2740) | C-103 종료 뒤 소조각 |
