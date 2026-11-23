---
id: SPEC-20261123-ready-mutant-line
title: "A READY note carries one mutant line per new guard test, and bundle check warns when it does not"
status: draft
revision: 1
created: 2026-11-23
---

# READY 노트가 새 지킴 테스트마다 뮤턴트 한 줄을 싣고, bundle check 가 없으면 경고

## Intent

2026-11-23 하루에 같은 종류의 문제가 감사에서 네 번 걸렸어요. 새로 넣은 지킴 테스트(guard test: 어떤 코드를 지키려고 쓴 테스트)가 지키는 코드를 빼도 통과했어요: 테스트 병렬 제한의 cap 파싱, `/clear` 뒤 이어받기의 경로 정규화, 테스트 영역 표의 영역 연결, 공식 LSP 규칙의 가짜 서버(PATH 에서 `touch` 를 못 찾아 실행돼도 표식이 안 생김). 레인의 확인은 모두 초록이었어요. 초록 테스트가 "빼도 통과"인지는 지키는 코드를 빼 봐야 알아요.

이 SPEC 은 되풀이 문제를 소스에서 막아요. (1) READY 안내가 새 지킴 테스트마다 뮤턴트(mutant: 지키는 코드를 빼거나 뒤집은 복사본) 한 번과 그 테스트가 실패한 결과를 한 줄로 싣게 하고, (2) `moai bundle check` 가 그 줄이 없거나 읽히지 않거나 뮤턴트가 살아남았으면 경고하고, (3) 가짜 실행 파일 테스트가 PATH 에 기대지 않고 표식을 남기는 도우미를 `tests/common` 에 두어요.

새 검사는 막지 않고 관찰부터 시작해요(`internal/sdd` 11번): 경고는 `ok`·`findings`·종료 코드를 바꾸지 않아요. 막는 쪽으로 바꾸는 일은 관찰 결과를 본 뒤 기록하는 결정이에요(D-3).

## Non-goals

- 뮤턴트를 자동으로 만들거나 돌리는 도구. 레인이 손으로 돌리고 결과 한 줄을 적어요.
- 줄이 사실인지 기계가 검증하는 일. 이 검사는 줄이 있고 읽히고 실패(`exit` 0 이 아님)를 말하는지만 봐요. 사실 확인은 감사의 몫이에요.
- 기존 테스트의 `touch` 를 새 도우미로 바꾸는 일. 도우미와 눈먼 경우를 재현하는 테스트까지만 하고, 옮기는 일은 각 테스트의 주인 변경이 맡아요.
- `bundle check` 의 기존 `findings` 코드 변경.

## Requirements

- REQ-1: The READY guidance of the flow reference shall require in the body of the READY note one line per new guard test, either `guard-mutant: <test> | <code removed or flipped> | exit <n>` with the non-zero exit code of the failing test, or `guard-mutant: none | <reason>` when the change adds no guard test · failure: none — a guidance text is read by the check of AC-1
- REQ-2: The self-check guidance for guards shall name the `guard-mutant:` line as the record of each mutation and a mutant that survives (`exit 0`) as a missing test · failure: none — a guidance text is read by the check of AC-2
- REQ-3: When `moai bundle check` judges a head, it shall read the body after the first line of the newest READY note that names the head and list the warning `mutant_line_missing` when the body holds no `guard-mutant:` line
- REQ-4 (REQ-3): IF a `guard-mutant:` line fits neither form of REQ-1, THEN `moai bundle check` shall list the warning `mutant_line_unreadable` that names the line number
- REQ-5 (REQ-3): IF a `guard-mutant:` line states `exit 0`, THEN `moai bundle check` shall list the warning `mutant_survived` that names the test
- REQ-6 (REQ-3): IF no READY note names the head, THEN `moai bundle check` shall list the warning `mutant_line_missing` with the detail that no note names the head
- REQ-7: The warnings of `moai bundle check` shall be listed under a `warnings` key and leave `ok`, `findings` and the exit status as they are · failure: none — an unchanged value is the check of AC-5
- REQ-8: The test helper `write_marking_executable` in `tests/common` shall write a fake executable whose marker is created by the shell's own redirection, with no external command and no PATH · failure: none — a missing marker is the failing test
- REQ-9: A test shall reproduce the blind case: a fake executable that creates its marker with `touch` under a PATH that holds no `touch` leaves no marker while it ran, and the helper's fake leaves one · failure: none — the test is its own proof
- REQ-10: The rule provenance table of `CLAUDE.md` shall hold one row that names the recurring cause and the date · failure: none — a missing row is read by the check of AC-8

## Acceptance

- AC-1 (REQ-1): 흐름 문서가 READY 노트의 `guard-mutant:` 두 형식과 "새 지킴 테스트마다 한 줄"을 적어요 · verify: `rg -c "guard-mutant:" skills/moai/references/flows/next-flow.md`
- AC-2 (REQ-2): 자체 점검 문서의 지킴 항목이 그 줄과 "살아남은 뮤턴트는 빠진 테스트"를 적어요 · verify: `rg -c "guard-mutant:" skills/moai/references/internal/self-check.md`
- AC-3 (REQ-3, REQ-4, REQ-5, REQ-6): 줄 읽기가 두 형식을 읽고 줄 없음·못 읽는 줄·`exit 0`·노트 없음을 각각 구분해요 · verify: `cargo test --offline --locked --lib bundle::mutant_line`
- AC-4 (REQ-3, REQ-4, REQ-5, REQ-6): `bundle check` 가 네 경고를 코드와 함께 내고, 줄이 제대로 있으면 경고가 없어요 · verify: `cargo test --offline --locked --test bundle_check_cli mutant_line`
- AC-5 (REQ-7): 경고가 있어도 `ok` 는 참이고 `findings` 는 비었고 종료 코드는 0 이에요 · verify: `cargo test --offline --locked --test bundle_check_cli mutant_line`
- AC-6 (REQ-8, REQ-9): 도우미의 가짜는 PATH 가 비어도 표식을 남기고, `touch` 가짜는 같은 PATH 에서 실행돼도 표식이 없어요 · verify: `cargo test --offline --locked --test fake_exe_marker`
- AC-7 (REQ-3, REQ-4, REQ-5, REQ-6, REQ-8): 새 지킴 테스트마다 지키는 코드를 뺀 뮤턴트가 의도한 테스트에서 실패해요. Result 의 뮤턴트 표에 적어요 · verify: Result 절의 뮤턴트 표.
- AC-8 (REQ-10): `CLAUDE.md` 의 규칙 출처 표에 원인 줄이 있어요 · verify: `rg -c "guard-mutant" CLAUDE.md`

## Scope

- write: `.moai/specs/SPEC-20261123-ready-mutant-line/**`, `skills/moai/references/flows/next-flow.md`, `skills/moai/references/internal/self-check.md`, `skills/moai/references/source-manifest.json`, `src/bundle.rs`, `src/bundle/mutant_line.rs`(새), `src/ops/bundle.rs`(도움말), `docs/design-review/appendix/operations/bundle.md`, `tests/bundle_check_cli.rs`, `tests/fake_exe_marker.rs`(새), `tests/common/fake_exe.rs`(새), `tests/common/mod.rs`(모듈 한 줄), `CLAUDE.md`
- read-only: `src/gate.rs`, `tests/common/**` 의 나머지, 기존 `touch` 가짜를 쓰는 테스트들
- forbidden: 실제 HOME, 새 파이썬 시험, 기존 테스트의 `touch` 교체

## Plan

1. RED: `src/bundle/mutant_line.rs` 의 줄 읽기 단위 테스트와 `tests/bundle_check_cli.rs` 의 경고 테스트, `tests/fake_exe_marker.rs` 를 먼저 써서 실패를 봐요.
2. GREEN: 줄 읽기(`mutant_line`)를 쓰고 `judge()` 가 경고를 모아 `check` 의 `warnings` 로 내요. 도우미 `write_marking_executable` 을 `tests/common/fake_exe.rs` 에 둬요.
3. 안내: 흐름 문서와 자체 점검 문서, 도움말(`src/ops/bundle.rs`)과 `bundle.md` 를 맞추고 매니페스트를 다시 써요.
4. 뮤턴트: 새 지킴 테스트마다 지키는 코드를 뺀 복사본을 돌려 Result 의 표에 적어요(이 SPEC 이 요구하는 줄의 첫 사용).
5. 규칙 출처 표에 원인 줄을 더해요.

## Risks

- 줄 형식이 노트 글에서 우연히 맞거나 틀릴 수 있어요. 줄은 `guard-mutant:` 로 시작할 때만 읽고(앞의 공백과 `- ` 는 허용), 다른 줄은 보지 않아요.
- 경고가 `ok` 를 바꾸지 않아서 레인이 무시할 수 있어요. 관찰이 끝난 뒤 막는 쪽으로 바꾸는 결정이 D-3 이에요.
- 시험 이름에 `|` 가 들어가는 일은 없지만, 들어가면 줄이 읽히지 않는 줄로 경고돼요.

## Decisions

결정은 되돌릴 수 있고 범주는 `none` 이에요. 결정자는 이 SPEC 의 리더이고, 답이 없으면 기본값으로 진행해요.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 줄은 `guard-mutant:` 로 시작하고 필드를 세로 막대로 나눠요: REQ-1 의 두 모양(세 칸 또는 `none` 과 이유 두 칸). 기본값: 이 모양. | 노트 글에서 한 줄로 읽히고 구분자가 시험 이름에 안 나와요. | 문서 한 곳과 줄 읽기 한 함수만 바꾸면 돼요. |
| D-2 | 경고는 `findings` 와 따로 `warnings` 에 담아요. 기본값: 따로. | 기존 `ok`·`findings` 를 읽는 시험과 리더의 읽기를 안 깨요. | 코드 한 곳에서 합쳐요. |
| D-3 | 막는 쪽으로의 전환은 관찰 뒤 기록하는 결정이에요. 관찰 창: 이 검사가 들어간 뒤 `bundle check` 가 도는 번들 열 번 또는 다음 릴리스 중 먼저 오는 것. Result 에 경고 횟수와 거짓 경고를 적어요. 기본값: 관찰 중에는 막지 않음. | `internal/sdd` 11번. | 해당 없음(막지 않음). |
| D-4 | `none` 줄이 새 시험을 더하는 변경에 붙었는지는 기계가 보지 않아요. 기본값: 보지 않고 감사가 봐요. 관찰에서 이 빈틈으로 빠지는 일이 보이면 후속으로 다뤄요. | 규칙 20: 있을 법한 먼 일은 설계하지 않아요. | 해당 없음. |

## Result

레인이 2026-11-23 에 채웠어요. 감사가 다시 확인해요. 가지 `spec/ready-mutant-line-20261123`, 기준 `0dfdfb31dc`, 구현 커밋 `b4dbae14f8` 과 그 위 두 커밋, 이 워크트리의 `target/`, `CARGO_BUILD_JOBS=5`, `RUSTC_WRAPPER=sccache`, 무거운 실행은 `moai heavy-slot-acquire --max 4`, macOS arm64. 같은 머신의 부하가 25~80 이었어요.

- AC-1: met — `rg -c "guard-mutant:" skills/moai/references/flows/next-flow.md` 가 1, exit 0
- AC-2: met — `rg -c "guard-mutant:" skills/moai/references/internal/self-check.md` 가 1, exit 0. `agent_defs_package` 가 고정한 지킴 문장은 그대로 두고 뒤에 더했어요(첫 시도는 낱말을 바꿔 그 시험이 실패했고 고쳤어요).
- AC-3: met — `cargo test --offline --locked --lib bundle::mutant_line` 8 passed
- AC-4: met — `cargo test --offline --locked --test bundle_check_cli` 16 passed(새 5개 `mutant_line_*` 포함)
- AC-5: met — 같은 시험의 `mutant_line_a_note_without_the_line_warns_and_changes_nothing_else` 가 `ok` 참, `findings` 빈 배열, 종료 코드 0 을 단정
- AC-6: met — `cargo test --offline --locked --test fake_exe_marker` 3 passed
- AC-7: met — 아래 뮤턴트 표, 17개 모두 의도한 시험에서 실패
- AC-8: met — `rg -c "guard-mutant" CLAUDE.md` 가 2(지키는 줄과 규칙 출처 줄), exit 0

뮤턴트(커밋한 트리에서 한 번에 하나, 대상 한 개, 돌린 뒤 `git checkout --` 로 되돌림. 스크립트와 로그는 `target/lane/mutants.zsh`, `mutants-first.log`, `mutants.log`):

| 뮤턴트 | 지키는 코드 | 실패한 시험 |
|---|---|---|
| m01 | `exit 0` 을 살아남음으로 읽음 | `a_mutant_that_survived_names_its_test` 외 1 |
| m02 | 줄 없음 경고 | `a_note_without_a_line_is_missing_…` 외 2 |
| m03 | 첫 줄(표식 줄)은 읽지 않음 | `the_marker_line_itself_is_never_read_as_a_mutant_line` |
| m04 | 빈 칸이 있으면 못 읽음 | `a_line_that_fits_neither_form_is_unreadable_…` |
| m05 | `exit` 뒤가 숫자인지 봄 | **첫 시도 생존**(`exit +1` 이 u32 로 읽힘) → 시험에 `exit +1` 을 더하고 같은 시험에서 실패 |
| m06 | 줄머리 `- ` 허용 | `a_bullet_and_leading_spaces_are_allowed` |
| m07 | 앞 공백 허용 | `a_bullet_and_leading_spaces_are_allowed` |
| m08 | 줄이 접두어로 시작해야 읽음 | `a_line_must_start_with_the_prefix` |
| m09 | `none` 모양 | `a_failing_mutant_line_and_a_none_line_give_no_warning` |
| m10 | 줄 번호(첫 줄이 1) | `a_line_that_fits_neither_form_is_…` 외 1 |
| m11 | 가장 새 노트를 읽음 | `mutant_line_is_read_from_the_newest_note_that_names_the_head` |
| m12 | head 를 이름으로 가진 노트만 읽음 | `mutant_line_a_head_no_note_names_is_missing` 외 1 |
| m13 | 경고가 `ok` 를 바꾸지 않음 | `mutant_line_a_note_without_the_line_…` 외 1 |
| m14 | 경고가 head 를 실음 | `mutant_line_unreadable_and_survived_lines_each_warn` |
| m15 | 도우미가 셸 리다이렉션으로 표식을 만듦(`touch` 로 되돌림) | `the_helpers_fake_leaves_its_marker_when_path_holds_nothing` 외 1 |
| m16 | 표식 경로의 따옴표 처리 | `the_marker_path_is_quoted_and_the_rest_of_the_script_runs` |
| m17 | 도우미가 뒤 줄을 이어 붙임 | `the_marker_path_is_quoted_and_the_rest_of_the_script_runs` |

눈먼 경우: `a_touch_fake_under_a_path_without_touch_ran_and_left_no_marker` 가 `touch` 가짜가 PATH 에 `touch` 가 없을 때 실행되고(종료 127) 표식이 없음을 보여요. 도우미의 가짜는 같은 PATH 에서 표식을 남겨요.

실행: `cargo test --offline --locked --lib` 1720 passed, `--bin moai` 28 passed, `--bin moai-cli` 28 passed, `--test` 의 `bundle_body_cli` 11, `bundle_build_cli` 18, `bundle_check_cli` 16, `bundle_ready_cli` 5, `bundle_verify_cli` 12, `bundle_user_assets_cli` 6, `fake_exe_marker` 3, `repo_claude_md` 5, `recurring_problems_rule` 5, `repo_scans_table` 7, `source_manifest_facts` 8, `repo_scans_cli` 19, 바뀐 문서를 `rg` 로 찾아낸 시험 대상(`target/lane/doc-readers.txt`, 첫 실행에서 `agent_defs_package` 가 실패해 고친 뒤 다시 돌려 통과)과 `product_cli skill_package::` 13 모두 exit 0. `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0, `cargo fmt --check` exit 0, `git diff --check` exit 0, `moai spec check` 오류 0·경고 0, `moai trailer-check 0dfdfb31dc..HEAD` failing 0, `moai package-inspect skills/moai` exit 0, `moai repo-scans --refs 0dfdfb31dc HEAD` 가 걸리는 스캔 없음. clippy 는 구현 커밋 위에서 돌렸고 그 뒤 두 커밋은 문서와 매니페스트만 바꿔요.

돌리지 않은 것: 파이썬 시험(규칙), CI, Linux·Windows, `--bin` 이 아닌 나머지 `--test` 전체. 위험 경로: 이 변경은 지우기·프로세스 시작·권한 규칙이 없고 표식은 셸 리다이렉션뿐이라 해당 없음.

관찰(D-3): 이 검사가 들어간 뒤 `bundle check` 가 돈 번들이 아직 없어서 경고 횟수와 거짓 경고는 0 번 관찰이에요. 막는 쪽으로의 전환은 관찰 뒤 기록하는 결정이고, 이 SPEC 은 그것을 하지 않았어요.

승인: 이 SPEC 은 초안(`draft`)이에요. 승인은 사용자의 선택이라 레인이 기록하지 않았어요.
