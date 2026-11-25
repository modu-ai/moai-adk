---
id: SPEC-20261125-check-fast-and-fix
title: "check fast 와 check fix: 장부 테스트를 맨 앞에 돌리고 손으로 맞추는 표를 명령 하나로 맞춰요"
status: draft
revision: 1
created: 2026-11-25
---

# check fast 와 check fix: 장부 테스트를 맨 앞에 돌리고 손으로 맞추는 표를 명령 하나로 맞춰요

## Intent

테스트 단계에서 실패한 검사의 절반은 손으로 맞추는 표(장부)를 빠뜨린 실패인데, 몇 초면 알 수 있는 그 실패가 중앙값 447초짜리 테스트 실행 끝에 나와요(측정: `~/.moai/reports/moai-adk/2026-11-25-build-test-load-rootcause.md` 3-4절). 이 SPEC 은 `moai check fast <range>` 가 가벼운 검사와 장부 테스트만 먼저 돌려 실패를 한 번에 모아 말하게 하고, `moai check affected` 가 그것을 clippy 와 긴 테스트 앞에 놓게 하며, `moai check fix <range>` 가 기계가 만들 수 있는 표 줄과 해시를 써넣고 사람이 정할 것은 그대로 출력하게 해요. 관문은 그대로예요: 장부 테스트는 지워지지 않고 넓은 검사에서도 돌아요. 읽는 사람: 레인, 감사자, 리더.

## Requirements

- REQ-1: When `check fast <base>..<head>` runs on a clean checkout at the head, the command shall run `git diff --check`, `cargo fmt --check`, `rustfmt --check` of each changed `src/ops` file, the table check, the end-line check and the ledger tests.
- REQ-2 (REQ-1): IF a step of `check fast` fails, THEN the command shall run the remaining steps and answer `result` FAIL with each failed step, its exit code, its log and the names of its failing tests.
- REQ-3: When the end-line check reads the changed files of a range, it shall name each regular, non-empty text file whose last byte is not a newline.
- REQ-4 (REQ-3): IF a changed file cannot be read, THEN the end-line check shall name it as a failure of the check.
- REQ-5: When `check affected` plans a Rust repository's steps, the plan shall place the ledger step after the cheap steps and before `clippy`.
- REQ-6 (REQ-5): IF the repository has no ledger test file, THEN the plan shall keep its earlier steps and add no ledger step.
- REQ-7: When the ledger step fails, `check affected` shall stop before `clippy` and `test` with `stopped_at=ledger` in its summary.
- REQ-8 (REQ-7): IF the ledger step cannot start, THEN `check affected` shall give it the failing exit code of the start error and stop as REQ-7 says.
- REQ-9: When `check fix <base>..<head>` finds a fact it can write without a human choice, it shall write that fact and re-run the same in-process checks.
- REQ-10 (REQ-9): IF the file to write is a link or lies under a linked folder, THEN `check fix` shall write nothing there and name the path.
- REQ-11 (REQ-9): IF the remaining violation needs a human choice (an allow-list row, an exemption reason, a manifest entry's origin, an area for a new first word), THEN `check fix` shall write nothing for it and print the missing line as it is.
- REQ-12: When `check fix` adds a registration to `tools/repo_scans.json`, the new text shall be one that `gate::registry_only_grew` reads as growth.
- REQ-13 (REQ-12): IF the new text would not read as growth, THEN `check fix` shall print it and write nothing.
- REQ-14: When the ledger list is resolved against the test files of a checkout, each name shall resolve to a test target, to a module of an area crate or to a library test.
- REQ-15 (REQ-14): IF a name resolves to none, THEN the ledger step shall fail and name it as missing.
- REQ-16: While a change is judged by the wide set, the ledger tests shall run in it unchanged.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 저장소에서 `diff_check`, 끝줄, `test_tables`, 장부 단계에 실패를 하나씩 심어 `check fast` 를 돌리면 한 답에 네 실패가 모두 들어 있고 `result` 가 FAIL 이에요 · verify: `cargo test --offline --locked --test check_fast_cli check_fast_collects_every_failure`
- AC-2 (REQ-3, REQ-4): 끝줄이 없는 파일은 이름으로 나오고, 빈 파일·NUL 이 든 파일·`templates/bundle-packs/` 아래 파일은 나오지 않으며, 읽을 수 없는 파일은 검사의 실패로 나와요 · verify: `cargo test --offline --locked --test check_fast_cli the_endline_check_flags_a_missing_final_newline`
- AC-3 (REQ-5, REQ-6, REQ-7, REQ-8): `check affected` 의 계획에서 장부 단계는 `clippy` 앞에 있고, 장부 단계가 실패하면 `clippy` 와 `test` 가 시작되지 않고 요약이 `stopped_at=ledger` 예요 · verify: `cargo test --offline --locked --test check_fast_cli a_failed_ledger_step_stops_clippy_and_test`
- AC-4 (REQ-9, REQ-11, REQ-12, REQ-13): 새 테스트 파일의 등록 줄과 source-manifest 해시를 써넣고, 사람이 정할 줄은 쓰지 않고 출력하며, 쓴 표는 `registry_only_grew` 가 더하기로 읽고, 두 번째 실행은 할 일이 없어요 · verify: `cargo test --offline --locked --test check_fix_cli check_fix_writes_what_is_mechanical_and_prints_the_rest`
- AC-5 (REQ-10): 링크인 표 파일은 쓰지 않고 이름으로 거절해요 · verify: `cargo test --offline --locked --test check_fix_cli check_fix_refuses_a_linked_table`
- AC-6 (REQ-14, REQ-15): 장부 목록의 모든 이름이 이 저장소에서 실제 대상으로 풀리고, 풀리지 않는 이름은 빠졌다고 나와요 · verify: `cargo test --offline --locked --lib check_fast::tests::the_ledger_names_resolve_in_this_repository`
- AC-7 (REQ-16): 넓은 변경의 전체 선택이 장부 대상을 모두 담고 있어요 · verify: `cargo test --offline --locked --lib check_fast::tests::the_full_pick_of_a_wide_change_holds_every_ledger_target`
- AC-8 (REQ-1, REQ-5): `moai check affected --card <base>..<head>` 의 요약이 `result=PASS` 예요 · verify: `moai check affected --card <base>..<head>` writes `result=PASS` in `summary.txt`

## Scope

- write: `.moai/specs/SPEC-20261125-check-fast-and-fix/**`, `src/check_fast.rs`, `src/check_fix.rs`, `tests/check_fast_cli.rs`, `tests/check_fix_cli.rs`
- write by `card line-lease` (shared files): `src/lib.rs` (two `pub mod` lines), `src/help.rs` (the `check` example lines)
- write after the card that changes the exit code of `check affected` lands (it holds them now): `src/check_affected.rs`, `src/ops/check.rs`
- widen later by `card scope`: `tests/source_manifest_facts.rs` (one copy of the manifest render), `docs/design-review/appendix/operations/check.md`, `docs/lane-rules.md`
- forbidden: `tools/repo_scans.json`, `tests/areas.toml` and `skills/moai/references/source-manifest.json` by hand (only the product command writes them, under `card line-lease`), `CLAUDE.md`, the real HOME

## Plan

1. SPEC 초안과 RED 테스트: 새 파일 `tests/check_fast_cli.rs`·`tests/check_fix_cli.rs` 에 AC-1, AC-2, AC-4, AC-5 를 임시 저장소와 가짜 cargo 로 써요. 올바른 이유로 실패하는지 읽어요.
2. A 조각(새 파일만): `src/check_fast.rs`(장부 목록, 이름 풀기, 필터셋, 끝줄 검사, 계획, 실행)와 `src/check_fix.rs`(매니페스트 해시 렌더, 계획, 쓰기)와 `src/lib.rs` 두 줄.
3. 단위 테스트 AC-6, AC-7 은 `src/check_fast.rs` 의 `mod tests` 에서 이 저장소의 실제 글을 읽어요(레지스트리 줄이 필요 없게).
4. B 조각(종료 코드를 고치는 다른 카드의 착지 뒤, 범위 넓힘): `src/ops/check.rs` 의 `Fast`·`Fix` 동작, `plan_steps` 에 장부 단계 끼우기, 멈춤 규칙(AC-3), `TMPDIR` 확장, 도움말과 문서.
5. 확인: `cargo fmt --all`, 새 파일 `moai check fast` 로 자기 자신 검사, 위험 경로 실행(멈춘 가짜 cargo, 링크인 매니페스트), `moai check affected --card` 를 READY 에서 한 번.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 장부 목록의 라이브러리 단위 테스트 둘(`the_repository_tests_pass_*`)은 `kind(lib)` 테스트로 돌려요(기본). 대안: 같은 판정 함수 `table_violations`·`scan_violations` 를 프로세스 안에서 부름 | 카드 본문이 이 테스트들을 이름으로 말해요. 측정에서 라이브러리 테스트 빌드가 시간을 지배하면 대안으로 바꿔요 | 목록의 두 줄을 바꿔요 |
| D-2 | 끝줄 검사는 바뀐 일반 파일의 마지막 바이트가 줄바꿈인지예요(기본). 빈 파일, NUL 이 든 파일, `templates/bundle-packs/**` 는 건너뛰고 끝의 빈 줄은 `git diff --check` 가 맡아요. 대안: 커밋 꼬리말 문단 검사 | 저장소에 기존 끝줄 검사가 없어요(검색 결과). 꼬리말은 `moai trailer-check` 가 따로 있어요 | 검사 한 함수를 지워요 |
| D-3 | 새 첫 낱말의 영역은 `--area <낱말>=<영역>` 으로 받고, 없으면 출력만 해요(기본). 대안: 같은 폴더 이웃에서 추측 | 영역 선택은 사람이 정하는 일이에요 | 옵션을 없애요 |
| D-4 | `check fix` 는 범위의 커밋과 깨끗한 head 만 봐요(기본). 대안: 스테이징·추적 안 된 파일도 봐요 | `check affected` 와 같은 전제예요. 레인이 고친 것을 한 커밋으로 더하고 다시 돌려요 | 해당 없음 |
| D-5 | 매니페스트 렌더는 `check_fix::render_manifest` 한 곳이고 `tests/source_manifest_facts.rs` 의 `rendered()` 가 그것을 불러요(기본). 대안: 두 사본과 같음 테스트 | 사본이 둘이면 어긋나요. 이 파일은 범위를 넓혀야 해요 | 사본으로 되돌려요 |

## Result

(구현 뒤에 AC 별 met·not_run 을 적어요.)
