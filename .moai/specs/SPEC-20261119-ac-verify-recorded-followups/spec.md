---
id: SPEC-20261119-ac-verify-recorded-followups
title: "Follow-up of the C-86 S3 audit, slice 1: the product path names the hook modes and the AC verify record (card C-130)"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-enforce-ac-verify-recorded
---

# Follow-up of the C-86 S3 audit, slice 1: the product path names the hook modes and the AC verify record (card C-130)

## Intent

C-86 S3(SPEC-20261119-enforce-ac-verify-recorded)는 tier 2 감사에서 차단 B1 을 고친 뒤 PASS 를 받았지만, 감사의 비차단 N1 이 남았어요: `docs/rust-product-path.md` 가 `[hooks.modes]` 표와 `ac_verify_recorded` 규칙을 말하지 않고, 46줄(`hook.dispatch` 행)과 961줄(`hooks-plan` 설명)이 코드와 어긋나요. 46줄은 엔진 로그가 쓰는 줄 종류를 `observe`·`stop_block`·`deny`·`config_error` 넷만 말하는데 코드는 `ac_run`·`would_block`·`rule_block`·`rule_release` 도 써요. 961줄은 "turn-validation rules are empty" 라고 말하는데 `STOP_RULES` 에는 `ac_verify_recorded` 가 있어요.

이 SPEC 은 카드 C-130 의 기준 1(N1, 문서)만 다뤄요. 기준 2~5(한도 문서, 입력 견고성, 읽기 한도, 호스트 모양)는 `src/hook_engine.rs`·`src/hook_guards.rs`·`tests/hooks_cli.rs` 를 쓰고, 그 파일들은 지금 C-86(S3·S3b)과 C-112 가 예약하고 있어서 예약이 풀린 뒤 다음 SPEC 으로 가요. 독자: 이 문서로 훅 설정을 켜는 사용자와, 훅 코드를 고치는 다음 레인.

## Non-goals

- 기준 2~5 의 코드 변경(`src/hook_engine.rs`, `src/hook_guards.rs`, `tests/hooks_cli.rs`)과 그 SPEC 문장. 예약이 풀린 뒤 별도 SPEC 으로 해요.
- `ac_verify_recorded` 를 `enforce` 로 올리는 결정(N10): 사용자 몫이고 이 SPEC 은 문서를 고치기만 해요. 제안은 카드 노트로만 남겨요.
- 문서 구조 개편, 다른 행의 문장 손질, `docs/hooks-generation.md` 의 변경.

## Requirements

- REQ-1: When a reader opens the `hook.dispatch` row of `docs/rust-product-path.md`, the row shall name the `[hooks.modes]` table of the HOME `hook-config.toml`, its three words and its default, the two rule ids `merge_confirm` and `ac_verify_recorded`, and every kind of line the hook engine writes to its log.
- REQ-2 (REQ-1): IF the `hooks-plan` description of `docs/rust-product-path.md` calls the turn-validation rules empty, THEN the description shall instead name `ac_verify_recorded` as the one Stop rule and say that the generated hook set does not call `tool-observe` or `turn-validation` yet.
- REQ-3 (REQ-1): The documented kinds of engine log line and the documented mode words shall equal the ones the code writes and reads, so that the sentence stays true until the code changes.

## Acceptance

- AC-1 (REQ-1, REQ-2): `docs/rust-product-path.md` 의 `hook.dispatch` 행이 `[hooks.modes]`·`merge_confirm`·`ac_verify_recorded`·`off`/`observe`/`enforce`·기본 `observe`·`ac_run`·`would_block`·`rule_block`·`rule_release` 를 이름으로 말하고, `hooks-plan` 설명에 "turn-validation rules are empty" 가 없고 `ac_verify_recorded` 가 있어요. · verify: `grep -cF '[hooks.modes]' docs/rust-product-path.md`; `grep -cF 'ac_verify_recorded' docs/rust-product-path.md`; `grep -cF rule_release docs/rust-product-path.md`; `sh -c '! grep -qF "turn-validation rules are empty" docs/rust-product-path.md'`
- AC-2 (REQ-3): 문서의 로그 줄 종류 여덟 개와 모드 낱말 셋이 코드에서 읽히는 값과 같아요: 코드의 `"kind"` 값에서 문서가 말하지 않는 것이 없고, 문서가 말하는 것은 코드가 쓰는 것이에요. · verify: `sh -c 'for k in observe ac_run stop_block deny config_error would_block rule_block rule_release; do grep -q "\"kind\": \"$k\"" src/hook_engine.rs || grep -q "line(\"$k\")" src/hook_engine.rs || { echo "$k"; exit 1; }; grep -qF "$k" docs/rust-product-path.md || { echo "doc lacks $k"; exit 1; }; done'`; `grep -nE '"(off|observe|enforce)" => Some\(RuleMode' src/hook_config.rs`
- AC-3 (REQ-1, REQ-2, REQ-3): 이웃 검사: 문서를 읽는 시험과 SPEC 검사가 그대로 통과해요. · verify: `cargo test --locked --offline --test doc_references`; `cargo test --locked --offline --test spec_cli the_repository_specs_pass_check`

## Scope

- write: `.moai/specs/SPEC-20261119-ac-verify-recorded-followups/spec.md`, `docs/rust-product-path.md`
- read-only: `src/hook_engine.rs`, `src/hook_config.rs`, `src/hook_guards.rs`, `src/hooks.rs`, `tests/hooks_cli.rs`
- forbidden: `src/` 와 `tests/` 의 모든 쓰기(이 조각), 실제 HOME·실제 프로젝트의 훅 설정, 네트워크, 새 의존성

## Plan

1. 이 SPEC 을 쓰고 `spec check` 로 확인해요.
2. `docs/rust-product-path.md` 의 46줄과 961줄을 코드에서 읽은 값으로 고쳐요(엔진 로그 줄 종류는 `src/hook_engine.rs` 의 `kind` 값, 모드 낱말은 `src/hook_config.rs` 의 `RuleMode::parse`).
3. AC 의 명령과 이웃 검사를 돌리고 변이(문서의 낱말 하나 지우기)로 AC-1·AC-2 가 실패하는지 봐요.
4. 기준 2~5 는 예약이 풀린 뒤 다음 SPEC 에서 해요.

## Risks

- 문서의 낱말이 코드보다 먼저 낡을 수 있어요. AC-2 의 명령이 그 어긋남을 잡지만, 코드가 줄 종류를 새로 더할 때 문서에 그 낱말이 없으면 AC-2 의 목록에도 없어서 놓쳐요. 목록의 갱신은 코드를 바꾸는 쪽의 몫이고, 다음 SPEC 이 시험으로 고정할 수 있어요.
- 이 조각만으로 기준 2~5 는 열린 채예요. 카드는 닫지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 카드 C-130 을 문서 조각과 코드 조각 둘로 나눈다 | C-86 이 예약한 파일이 코드 조각과 겹치고, 아직 못 쓰는 시험을 SPEC 에 적으면 거짓 증거가 돼요(레인 agent-16, 리더 확인 대기) | 다음 SPEC 을 이 SPEC 에 합치면 돼요 |

## Result

- AC-1: met — RED: 문서를 고치기 전에 네 명령이 모두 종료 1(`[hooks.modes]`, `ac_verify_recorded`, `rule_release` 없음, "turn-validation rules are empty" 있음); GREEN: 고친 뒤 네 명령 모두 종료 0(`docs/rust-product-path.md` 46줄 `hook.dispatch` 행과 961줄 `hooks-plan` 설명).
- AC-2: met — RED: `doc lacks ac_run`(종료 1); GREEN: 로그 줄 종류 여덟 개(`observe`, `ac_run`, `stop_block`, `deny`, `config_error`, `would_block`, `rule_block`, `rule_release`)가 코드의 `"kind"` 값 또는 `line("…")` 로 나오고 문서에도 있으며, 모드 낱말 셋은 `src/hook_config.rs` 의 `RuleMode::parse` 세 줄(`off`, `observe`, `enforce`)과 같아요. 세 번째 거절이 통과한다는 문장은 `RULE_BLOCK_CAP = 2`(`src/hook_engine.rs:703`)와 `strikes >= RULE_BLOCK_CAP`(:752)에서 읽었어요.
- 변이(임시 폴더의 사본에서, 저장소 파일 불변): M1 문서의 `rule_release` 를 바꿈 → AC-1c·AC-2 FAILED; M2 옛 문장("the turn-validation rules are empty")을 되살림 → AC-1d FAILED; M3 문서의 `would_block` 을 바꿈 → AC-2 FAILED. 셋 모두 KILLED, 복원 뒤 사본이 저장소 문서와 `cmp` 동일.
- AC-3: 이웃 검사(`doc_references`, `spec_cli the_repository_specs_pass_check`)는 READY 노트의 종료 코드로 적어요.
- 열린 채 남은 것: 카드 C-130 의 기준 2~5(한도 문서, 입력 견고성, 읽기 한도, 호스트 모양)는 이 SPEC 의 범위 밖이에요. 쓸 파일(`src/hook_engine.rs`, `src/hook_guards.rs`, `tests/hooks_cli.rs`)을 C-86·C-112 가 예약하고 있어서 예약이 풀린 뒤 다음 SPEC 으로 해요(`src/hook_guards.rs` 는 카드의 쓰기 범위에도 없어요: 기준 3·4 가 거기 있는 `collapse_whitespace`·`head_sha`·`verify_commands` 를 고쳐야 해요). `enforce` 로 올리는 결정(N10)은 사용자 몫이라 카드 노트의 제안으로만 남겨요.
