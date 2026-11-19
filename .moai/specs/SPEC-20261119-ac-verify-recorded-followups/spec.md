---
id: SPEC-20261119-ac-verify-recorded-followups
title: "Follow-up of the C-86 S3 audit, slice 1: the product path names the hook modes and the AC verify record (card C-130)"
status: draft
revision: 2
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
- REQ-4 (REQ-1): The `hook.dispatch` row shall say that only a successful, finished Bash call leaves an `ac_run` line, and that an enforce refusal whose `rule_block` line cannot be written goes through (tier 1 audit of head e709e4396f, B1 and B2).

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-4): `docs/rust-product-path.md` 의 `hook.dispatch` 행 안에서(파일 전체가 아니라 그 행만 읽어요) `[hooks.modes]`·`merge_confirm`·`ac_verify_recorded`·`enforce`·기본값 낱말 `observe`("reads as `observe`)" 꼴)·"goes through instead"·"successful, finished"·"record nothing"·"Codex and OMP refuse"·"third refusal" 이 나오고, 코드(`src/hook_engine.rs`)의 `"kind"` 값과 `line("…")` 값 전부가 그 행에 "`종류` line" 꼴로 나와요(코드에 새 줄 종류가 생기면 이 명령이 실패해요). · verify: `sh -c 'row=$(grep -E "^\| .hook\.dispatch. \(rc\.2" docs/rust-product-path.md | head -1); [ -n "$row" ] || { echo no row; exit 1; }; for w in "[hooks.modes]" merge_confirm ac_verify_recorded enforce "absent or any other word reads as" "line per tool-observe" "goes through instead" "successful, finished" "record nothing" "Codex and OMP refuse" "third refusal"; do printf %s "$row" | grep -qF -- "$w" || { echo "row lacks: $w"; exit 1; }; done; printf %s "$row" | grep -qE "reads as .observe.[)]" || { echo "default word"; exit 1; }; for k in $(grep -ohE "\"kind\": \"[a-z_]+\"|line\(\"[a-z_]+\"\)" src/hook_engine.rs | sed -E "s/.*\"([a-z_]+)\".*/\1/" | sort -u); do printf %s "$row" | grep -qE "$k. line" || { echo "row lacks kind: $k"; exit 1; }; done'`
- AC-2 (REQ-2): `hooks-plan` 설명 줄이 `ac_verify_recorded`·"turn-validation rule, "·`--host claude`·`--host codex`·`--host omp`·`tool_result` 를 말하고 "rules are empty" 가 없어요. · verify: `sh -c 'l=$(grep -E "^.hooks-plan PROJECT --host claude.codex. previews" docs/rust-product-path.md | head -1); [ -n "$l" ] || { echo no line; exit 1; }; for w in "turn-validation rule, " ac_verify_recorded "--host claude" "--host codex" "--host omp" tool_result; do printf %s "$l" | grep -qF -- "$w" || { echo "line lacks: $w"; exit 1; }; done; ! printf %s "$l" | grep -qF "rules are empty"'`
- AC-3 (REQ-1, REQ-2, REQ-3, REQ-4): 이웃 검사: 문서를 읽는 시험과 SPEC 검사가 그대로 통과해요. · verify: `cargo test --locked --offline --test doc_references`; `cargo test --locked --offline --test spec_cli the_repository_specs_pass_check`

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

- 문서의 낱말이 코드보다 먼저 낡을 수 있어요. AC-1 의 명령이 코드의 줄 종류 값을 직접 읽어 문서 행에 없으면 실패해요. 반대로 문서가 코드에 없는 줄 종류를 더 말하는 것과, 한 낱말이 행에 두 번 나올 때 한 곳만 바뀐 것은 못 잡아요(알려진 한계).
- 이 조각만으로 기준 2~5 는 열린 채예요. 카드는 닫지 않아요.
- 문서의 enforce 설명은 줄 쓰기 실패 시 통과라는 코드의 선택(`Err(_) => Verdict::Allow`)을 그대로 옮긴 것이고, 그 선택이 옳은지는 이 SPEC 의 판단 밖이에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 카드 C-130 을 문서 조각과 코드 조각 둘로 나눈다 | C-86 이 예약한 파일이 코드 조각과 겹치고, 아직 못 쓰는 시험을 SPEC 에 적으면 거짓 증거가 돼요(레인 agent-16, 리더 확인 대기) | 다음 SPEC 을 이 SPEC 에 합치면 돼요 |

## Result

- 감사 반영(head e709e4396f 의 tier 1 감사, 권고 FAIL): B1 "enforce 거절" 문장에 "줄을 못 쓰면 거절 없이 통과" 절을 더했어요(`src/hook_engine.rs:755-759` 의 `Err(_) => Verdict::Allow`, `src/hooks.rs` `append_receipt_line_in` 의 `.moai-cli` 없음·로그 가득·일반 파일 아님). B2 `ac_run` 문장을 "성공한 끝난 Bash 호출만 기록, 실패·백그라운드·중단은 기록 안 됨"으로 고쳤어요(`observe_outcome`: Claude 는 `PostToolUseFailure` 가 아님, Codex 는 `exit_code` 0, OMP 는 `is_error` false; `tool_observe` 가 `outcome == "success"` 일 때만 `record_ac_run` 을 불러요). 비차단 8(961줄)은 "`--host claude`·`codex` 에서는 부르지 않고 OMP 확장은 `tool_result` 에서 `tool-observe` 를 불러요"로 한정했어요. 비차단 1~4(AC 가 파일 전체·낱말만 봄)는 AC 를 `hook.dispatch` 행 안, 낱말마다 "`종류` line" 꼴, 코드의 `"kind"`·`line("…")` 값을 자동으로 읽어 전부 요구하도록 바꿨어요. 비차단 5~7(`config_error` 절의 옛 `~/.moai-cli/config` 이름, `reason` 목록의 `invalid_edit_diagnostics`, `--host` 목록의 `omp`)은 같은 행의 원래 문서 어긋남이라 이 조각 밖이고 카드 노트로 넘겨요.
- AC-1: met — 같은 명령을 기준 문서(e84d7c46ce)에서 돌리면 `row lacks: [hooks.modes]` 종료 1(RED), 고친 문서에서는 종료 0(GREEN).
- AC-2: met — 기준 문서에서 `line lacks: turn-validation rule, ` 종료 1(RED), 고친 문서에서 종료 0.
- 변이(임시 폴더의 사본에서, 저장소 파일 불변, 열두 개): 열한 개 KILLED — 감사의 MA(`deny` 절 삭제)·MB(`merge_confirm` 이름 바꿈)·MC(기본 낱말 `observe`→`enforce`)·MD(세 번째→첫 번째)와 내 것(통과 절 삭제, "successful" 삭제, `--host omp` 삭제, `stop_block`·`rule_release` 이름 바꿈, "record nothing" 삭제, "refuse"→"ask"). SURVIVED 하나: `would_block` 의 한 곳만 바꾸기(같은 행에 두 번 나와서 다른 한 곳이 남아요). 알려진 한계예요.
- AC-3: 이웃 검사(`doc_references`, `spec_cli the_repository_specs_pass_check`)는 READY 노트의 종료 코드로 적어요.
- 열린 채 남은 것: 카드 C-130 의 기준 2~5(한도 문서, 입력 견고성, 읽기 한도, 호스트 모양)는 이 SPEC 의 범위 밖이에요. 쓸 파일(`src/hook_engine.rs`, `src/hook_guards.rs`, `tests/hooks_cli.rs`)을 C-86·C-112 가 예약하고 있어서 예약이 풀린 뒤 다음 SPEC 으로 해요(`src/hook_guards.rs` 는 카드의 쓰기 범위에도 없어요). `enforce` 로 올리는 결정(N10)은 사용자 몫이라 카드 노트의 제안으로만 남겨요.
