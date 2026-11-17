---
id: SPEC-20261117-spec-tooling
title: SPEC 도구 — spec.md 해석기와 moai spec new·check·list·show (SDD-1, SDD-2)
status: approved
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
approval:
  by: user
  at: 2026-11-17
  basis: "in-session selection 2026-11-17: 승인"
  digest: "sha256:e92728ac5905b4b9bd2e5018472aec4d801a1c01d61e5e34c2472aaba9945d0b"
---

## Intent

MoAI는 저장소를 바꾸는 모든 일을 SPEC으로 해요(결정 U-A). 그런데 SPEC 파일을 만들고 검사하는 도구가 없어서,
`.moai/specs/<SPEC-ID>/spec.md`는 손으로 쓰고 눈으로만 확인했어요. 이 SPEC은 SDD 계획의 SDD-1(해석기)과
SDD-2(`moai spec` 명령)를 만들고, 독립 감사가 낸 FAIL(B1–B4, N2–N7)을 같은 가지에서 고쳐요.

## Non-goals

- `spec take`(SDD-7), `spec close`(SDD-8), 승인 기록을 쓰는 명령, 카드 ↔ SPEC 연결(SDD-4).
- 커밋 관문의 `Moai-Spec`·`Moai-Session` 트레일러 검사(SDD-6).
- 완전한 YAML 해석: 우리가 쓰는 모양(한 줄 값, `null`, `approval` 흐름 맵)만 읽어요.
- `.gitignore` 변경(기반 SPEC이 했어요).

## Requirements

- REQ-1: When `moai spec new <slug> --title <t>` runs, the tool shall write only `.moai/specs/SPEC-<yyyymmdd>-<slug>/spec.md` from the built-in template with status `draft` and revision 1.
- REQ-2 (REQ-1): IF the SPEC folder already exists or an input is refused, THEN the tool shall write nothing and exit 1.
- REQ-3: When `moai spec check` reads a SPEC, the checker shall report each problem as a finding with a stable code and exit 0 only if no finding is an error.
- REQ-4 (REQ-3): IF `spec.md` is a link, not a regular file, larger than 256 KiB or not UTF-8, THEN the checker shall report a `read_*` error without following the link.
- REQ-5: The checker shall read each requirement as a GEARS sentence or as an IF/THEN unwanted-behaviour sentence.
- REQ-6: When a GEARS requirement has a `When` trigger, the checker shall require an IF/THEN requirement that names it or a trailing failure-none note. · failure: none — 이 규칙을 어긴 SPEC은 오류 보고로 끝나요.
- REQ-7: Where the status is `approved` or `completed`, the checker shall report open clarification markers and unfilled template placeholders as errors.
- REQ-8: The checker shall warn above 12 requirements or above 400 lines.
- REQ-9: When `--closure` is given, the checker shall require status `completed` and every AC `met` or `withdrawn` in `## Result`. · failure: none — 조건이 맞지 않으면 오류 보고예요.
- REQ-10: The state classification shall read a project `.moai/` that holds only `specs/` as no MoAI state.
- REQ-11: The `spec` command shall run no program and open no network connection.
- REQ-12: The repository SPECs shall pass `moai spec check --all`.

## Acceptance

- AC-1 (REQ-1, REQ-2): 새 SPEC 하나와 폴더만 생기고, 거절은 아무것도 쓰지 않아요.
  verify: `cargo test --locked --offline --test spec_cli new_`
- AC-2 (REQ-3, REQ-4, REQ-5): 해석기 경계(id·상태·앞머리·절·REQ·AC·범위·읽기 한도, N/N+1)가 시험으로 고정돼요.
  verify: `cargo test --locked --offline --test spec_doc`
- AC-3 (REQ-6): `When` 요구마다 IF/THEN 연결이나 줄 끝 메모가 필요해요.
  verify: `cargo test --locked --offline --test spec_doc when_requirements_need_a_failure_path`
- AC-4 (REQ-7, REQ-8): 덜 쓴 글은 승인 뒤 오류, 초안에서 경고이고 크기 예산은 12/13·400/401에서 갈려요.
  verify: `cargo test --locked --offline --test spec_doc -- placeholders_and_clarifications_follow_the_status the_size_budget_warns_above_twelve_requirements_and_four_hundred_lines`
- AC-5 (REQ-9): 마감 검사는 `completed`와 `## Result`의 `met`·`withdrawn`을 요구해요.
  verify: `cargo test --locked --offline --test spec_cli closure_needs_completed_and_a_result`
- AC-6 (REQ-10): `spec new` 앞뒤로 판정·`project-detect`·`project-update-plan`이 같고, `.moai-cli` 이동과 되돌리기가 SPEC을 지켜요(옛 `specs/`는 SPEC 하나씩 합치고, 같은 id는 충돌로 아무것도 옮기지 않아요).
  verify: `cargo test --locked --offline --test spec_cli -- spec_new_keeps_the_state_verdict_and_the_project_answers an_earlier_root_moves_into_a_moai_that_holds_specs_and_back earlier_specs_with_other_ids_merge_and_the_revert_splits_them_again an_earlier_spec_with_the_same_id_blocks_the_move only_a_real_specs_folder_reads_as_no_state`
- AC-7 (REQ-11): 두 소스 파일에 프로세스·소켓 코드가 없어요.
  verify: `grep -n -e 'Command::new' -e 'std::net' src/spec_doc.rs src/ops/spec.rs` → 출력 없음, exit 1
- AC-8 (REQ-12): 저장소의 SPEC이 모두 통과해요.
  verify: `cargo test --locked --offline --test spec_cli the_repository_specs_pass_check`
- AC-9 (REQ-1…REQ-12): 이웃 검사가 모두 통과해요.
  verify: `cargo test --locked --offline --bin moai`; `cargo test --locked --offline --test help_layout`;
  `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test product_cli`;
  `cargo test --locked --offline --test project_update_plan_cli`; `cargo test --locked --offline --test gitignore_block_cli`;
  `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py tests.skill_references_test`

## Scope

- write: `src/spec_doc.rs`, `src/ops/spec.rs`, `src/spec.rs`, `src/state_root.rs`, `src/go_state.rs`,
  `src/project_detect.rs`, `src/lib.rs`, `src/help.rs`, `templates/spec/full.md`, `templates/spec/minimal.md`,
  `tests/spec_doc.rs`, `tests/spec_cli.rs`, `tests/fixtures/help/all.txt`, `tests/fixtures/help/all-60.txt`,
  `tests/fixtures/help/all-80.txt`, `tests/fixtures/help/all-dumb.txt`,
  `docs/design-review/appendix/operations/spec.md`, `.moai/specs/SPEC-20261117-spec-tooling/spec.md`,
  `.moai/specs/SPEC-20261117-sdd-foundation/spec.md`, `tools/repo_scans.json`
- read-only: `src/archive.rs`, `src/main.rs`, `build.rs`, `skills/moai/**`
- forbidden: `.gitignore`, 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. SDD-1: `src/spec_doc.rs` 해석기와 검사, 시험 먼저(RED → GREEN), 검사기 mutation 5개.
2. SDD-2: `src/ops/spec.rs`(`card`처럼 한 낱말에 동작 넷), 템플릿, 도움말·예시·골든, 한국어 동작 문서.
3. 사용자 결정: GEARS + EARS 원치 않는 동작 형태, 실패 경로 규칙.
4. 감사 FAIL 수리: tip 병합, B1 상태 판정, B2 형식·기반 SPEC, B3 덜 쓴 글·크기 예산, B4 이 SPEC, N2–N7.
5. 이웃 시험·형식·clippy·Python 검사, 새 규칙 mutation 5개.

## Risks

- `.moai/`에 `specs/`만 있는 폴더를 상태 없음으로 읽는 규칙이 다른 명령의 가정과 어긋날 수 있어요 →
  `project-detect`·`project-update-plan` 앞뒤 비교 시험과 `.moai-cli` 이동·되돌리기 시험이 지켜요. Go 폴더
  되돌리기처럼 시험하지 않은 길은 남은 위험이에요.
- 자리 표시 검사가 작성자의 `<…>`를 잘못 잡을 수 있어요 → 템플릿에 있는 `<…>`만 봐요.
- 기존 승인 SPEC을 고치면 승인과 내용이 어긋날 수 있어요 → 기반 SPEC은 리더 결정(B2)으로만 고치고 D8에 적었어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | `spec`은 `card`처럼 한 낱말에 동작 넷이고 봉투 `operation`은 `spec` | 구현 판단 | 동작마다 op 파일로 나누기 |
| D2 | SPEC은 언제나 `.moai/specs`(`.moai-cli` 프로젝트도) | 결정 U-A, 리더 B1 | `state_root::path` 사용 |
| D3 | `spec new`는 `.moai/`가 없어도 만들고 표식은 쓰지 않음 | 리더 B1 | `.moai/` 없으면 거절 |
| D4 | 오류가 있는 `check`는 실패 봉투(`invalid_input`, 종료 1, 50개까지) | 구현 판단 | `main.rs`에 종료 특례 |
| D5 | `Non-goals`나 `Risks`가 있으면 full이고 둘 다 필요, `Plan`은 늘 필요 | 구현 판단 | 앞머리 `form` 키 |
| D6 | `approval`은 approved·completed·superseded에 필요, 네 칸(`digest` 포함) | 리더 기본값 | 칸 목록 수정 |
| D7 | Go 세대 SPEC은 읽기·블록·id·상태 코드만 오류 | 구현 판단(1,000개가 넘는 옛 SPEC) | `LEGACY_KEEPS` 넓히기 |
| D8 | 실패 경로 규칙은 오류 `req_failure_path` | 사용자 결정 2026-11-17 | 경고로 바꾸기 |
| D9 | IF/THEN은 `When <조건>, …`으로 바꿔 `parse_gears`로 검사 | 기존 코드 재사용 | 별도 해석기 |
| D10 | `unknown:` AC는 경고, 범위는 `write` 경로 하나 이상 | 리더 형식 | 오류로 바꾸기 |
| D11 | 앞머리 해석기는 따로(보관 해석기는 `approval`을 못 읽음), 새 crate 없음 | 구현 판단 | `archive.rs` 공유 |
| D12 | `.moai/`에 진짜 `specs/` 하나만 있으면 판정은 `.moai/`가 없는 것과 같음 | 리더 B1 | `specs_only` 제거 |
| D13 | `adopt_new`·`.moai-cli` 이동은 그 폴더를 쓰고, 되돌리기는 `specs/`를 돌려놓음 | B1의 결과 | 이동 거절 |
| D14 | 흐름 맵·`null`·범위 참조를 읽음 | 리더 B2 | 들여 쓴 모양만 |
| D15 | 자리 표시는 내장 템플릿의 `<…>`만, 승인·완료에서 오류 | 리더 B3 | 모든 `<…>` |
| D16 | 크기 예산 12개·400줄은 경고 | 리더 B3 | 숫자 바꾸기 |
| D17 | AC·범위 글에 `parse_gears`의 글 검사(`text_unsafe`) | 감사 N2 | 검사 빼기 |
| D18 | `failure: none`은 줄 끝 `· failure: none`만 | 감사 N5 | 아무 자리 허용 |
| D19 | `approval.at`은 시각 없이 날짜만, 정확히 `YYYY-MM-DD`(`2026-11-17T10:00Z`는 거절) | 감사 N6, 재감사 N-g | 앞 10자만 |
| D20 | 이 SPEC의 승인은 사용자의 세션 선택 "승인"(2026-11-17); `digest`는 모든 수정 뒤 마지막에 계산 | 결정 U-C, 재감사 RB2 | `draft`로 되돌리고 `approval` 삭제 |
| D21 | tip 병합 커밋에는 트레일러가 없음(병합 기본 메시지) | 지시는 새 커밋만 요구 | 다음 병합에 트레일러 |
| D22 | 옛 `.moai-cli/specs/`는 `specs/`만 든 `.moai/`로 SPEC 하나씩 옮기고, 같은 id가 양쪽에 있으면 충돌로 이동 전체를 막음(아무것도 안 옮김) | 재감사 RB1 | 폴더째 이동(목적지 충돌로 중간 실패) |
| D23 | `specs/` 옆 Finder `.DS_Store` 파일은 상태가 아님 | 재감사 N-c | 규칙에서 빼기 |
| D24 | `go-state-plan`의 문장은 `specs/`만 든 `.moai/`가 있어도 같음(B1: `spec new`가 계획 답을 바꾸지 않음); 적용의 영수증 문장만 어느 `.moai/`인지 말함 | B1과 재감사 N-f | 계획 문장도 바꾸기 |
| D25 | 되돌리기 계획은 `.moai/specs`에 `restored_to_moai_specs` 항목을 두고, 되돌리기는 그 항목이 있을 때만 `specs/`를 돌려놓음 | 재감사 N-e | 항목 없이 늘 돌려놓기 |
| D26 | 기반 SPEC은 사용자의 세션 선택으로 revision 2로 다시 승인됨(그 SPEC의 D8) | 재감사 RB2 | revision 1과 그 승인으로 |

## Result

구현 에이전트가 2026-11-17에 채웠어요. 감사가 다시 확인해요. 빌드 폴더는 이 worktree의 `target/`,
`CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64, cargo는 한 번에 하나.

커밋: 기준 `6aa92fd0dd` 위 SDD-1/2(`3a09377222`…`c241f48eed`) → 감사 FAIL → tip `da4d9e3d10` 병합 →
`c779c2624c`(B1–B4·N2–N7 수리) → 이 Result와 `tools/repo_scans.json` 항목(다음 커밋). 시험은 모두 먼저
RED(기대한 단언에서 실패)를 본 뒤 GREEN이에요.

- AC-1: met — `spec_cli` exit 0, 12 passed(`new_check_list_show_and_nothing_else_is_written`·`new_refuses_and_writes_nothing` 포함)
- AC-2: met — `spec_doc` exit 0, 23 passed. 검사기 mutation 5개 모두 KILLED(SDD-1 때)
- AC-3: met — 같은 실행. mutation M5(줄 끝 아닌 메모 허용) KILLED
- AC-4: met — 같은 실행. mutation M2(승인 뒤에도 경고)·M3(12개에서 경고) KILLED
- AC-5: met — `spec_cli` 같은 실행
- AC-6: met — `spec_cli` 같은 실행, `project_update_plan_cli` 57 passed, `go_state_cli` 61, `go_state_preview_cli` 44, `state_root_cli` 6, `project_detect_cli` 18, `--lib` state_root·project_detect·go_state·spec·home_root 45 passed. mutation M1(`specs_only` 늘 거짓) KILLED
- AC-7: met — `grep` 출력 없음, exit 1
- AC-8: met — `moai spec check --all --project .` exit 0, 2개 SPEC 오류 0·경고 0. mutation M4(`..` 범위 안 읽음) KILLED
- AC-9: met — `--bin moai` 27, `help_layout` 10, `home_isolation_scan` 16, `product_cli` 75, `gitignore_block_cli` 32, `ac_command_lint` 5, 저장소 검사 대상 7개 모두 exit 0; `cargo fmt --check` 0; `rustfmt --check --edition 2021 src/ops/*.rs` 0; clippy 0; Python 44개 OK(처음에 `repo_scans_test` 1개가 이 시험을 저장소 순회로 판정 → 표에 항목 추가 뒤 OK), `skill_references_test` 2개 OK

실행 안 함: 전체 `cargo test`(바뀐 곳 시험만), Linux·Windows, Go 폴더 되돌리기 길과 `specs/`가 함께 있는 경우.
