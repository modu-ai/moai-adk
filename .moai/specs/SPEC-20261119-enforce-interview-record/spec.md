---
id: SPEC-20261119-enforce-interview-record
title: "An approved SPEC carries an interview record, and spec check refuses one that does not"
status: draft
revision: 1
created: 2026-11-19
---

# An approved SPEC carries an interview record, and spec check refuses one that does not

## Intent

카드 C-86 의 AC-1 이에요("인터뷰해서 SPEC 을 쓰게 하라"를 지침이 아니라 검사로). 지금은 SPEC 을 `approved` 로 만드는 데 인터뷰가 있었다는 흔적이 필요 없어서, 에이전트가 사용자에게 한 번도 묻지 않고 SPEC 을 쓰고 승인 기록(`approval`)까지 채울 수 있어요. 이 SPEC 은 승인된 SPEC 에 `## Interview` 절을 요구해요: 질문과 답을 한 줄씩, 또는 "물을 것이 없었다"는 이유 한 줄. 이 절이 없거나 모양이 틀린 승인된 SPEC 을 `moai spec check` 가 오류로 거절해요.

이 검사는 기록의 모양만 확인해요. 질문 도구로 실제로 물었는지는 파일만 보고 알 수 없어요(Risks).

## Non-goals

- 실제로 질문 도구를 썼는지 확인하기. 호스트마다 질문 도구가 다르고 대화 기록은 이 검사가 읽지 않아요.
- PLAN 워크플로 문서(`skills/moai/workflows/plan.md`)와 SPEC 템플릿(`templates/spec/*.md`)에 인터뷰 절차를 적기: 이 조각의 쓰기 범위 밖이라 리더에게 제안만 해요(별도 조각).
- `moai spec approve` 같은 새 명령. 승인은 지금처럼 `approval` 을 적는 일이고 이 조각은 그 결과를 검사해요.
- 2026-11-19 전에 만든 SPEC, `draft` SPEC, Go 세대 SPEC 에 인터뷰를 요구하기.
- 인터뷰 절을 승인 `digest` 에 넣기: 기존 승인의 digest 가 모두 깨져요(Risks).

## Requirements

- REQ-1: The SPEC checker shall report `interview_missing` as an error for a SPEC with a new-form id, a status of approved, completed or superseded, a `created` date of 2026-11-19 or later, and no `## Interview` section.
- REQ-2: The SPEC checker shall read each list line of `## Interview` as an entry, either `- Q<n>: <question> → A: <answer>` (the arrow may be written `->`) or `- none: <reason>`.
- REQ-3 (REQ-2): IF an entry is of neither form, or has an empty question or answer, or is a `none` entry whose reason is shorter than 10 characters, or is a `none` entry beside any other entry, or the section has no entry, THEN the SPEC checker shall report `interview_entry`, as an error for an approved, completed or superseded SPEC and as a warning for any other status.
- REQ-4: The SPEC checker shall ask for no interview section from a SPEC created before 2026-11-19, a SPEC whose status is not approved, completed or superseded, or a Go-generation SPEC, while reading an `## Interview` section that such a SPEC has.
- REQ-5: The SPEC checker shall accept `## Interview` as a known section of every SPEC, as one section that appears at most once.
- REQ-6: The `moai spec check` command shall refuse an approved SPEC created on or after 2026-11-19 that has no valid interview record, and pass it once the record is valid.

## Acceptance

- AC-1 (REQ-1, REQ-4): 승인된 새 SPEC(2026-11-19 이후)에 `## Interview` 가 없으면 `interview_missing` 오류. 2026-11-18 에 만든 SPEC·`draft`·Go 세대 SPEC 은 없어도 조용해요.
  verify: `cargo test --locked --offline --test spec_doc interview_is_required`
- AC-2 (REQ-2, REQ-3): 질문·답 줄과 `none` 줄의 읽기, 틀린 모양 여섯 가지(모양 아님, 빈 질문, 빈 답, 짧은 이유, `none` 과 다른 줄이 함께, 줄이 하나도 없음)가 `interview_entry`, 승인된 SPEC 은 오류이고 그 밖의 상태는 경고.
  verify: `cargo test --locked --offline --test spec_doc interview_entries`
- AC-3 (REQ-5): `## Interview` 는 알려진 절이라 `section_unknown` 이 없고, 두 번 쓰면 `section_duplicate`, 인터뷰가 없던 옛 SPEC 의 읽기는 그대로예요.
  verify: `cargo test --locked --offline --test spec_doc interview_section`
- AC-4 (REQ-6): `moai spec check` 가 기록 없는 승인된 SPEC 을 거절(오류 1, 코드 `interview_missing`)하고 기록을 더하면 통과해요.
  verify: `cargo test --locked --offline --test spec_cli interview`
- AC-5 (REQ-1…REQ-6): 이웃 시험과 범위.
  verify: `cargo test --locked --offline --test spec_doc`; `cargo test --locked --offline --test spec_cli`; `cargo test --locked --offline --test card_spec_cli`; `cargo test --locked --offline --test spec_contract`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `cargo fmt --check`

## Scope

- write: `src/spec_doc.rs`, `tests/spec_doc.rs`, `tests/spec_cli.rs`, `.moai/specs/SPEC-20261119-enforce-interview-record/spec.md`
- read-only: `src/spec.rs`, `src/spec_cli.rs`, `templates/spec/full.md`, `templates/spec/minimal.md`, `skills/moai/workflows/plan.md`
- forbidden: the real HOME and the real host settings, `skills/**`, `templates/**`, `src/hook_engine.rs`, `src/hook_guards.rs`

## Plan

1. 시험 먼저: `tests/spec_doc.rs` 에 `interview_is_required`, `interview_entries`, `interview_section`, `tests/spec_cli.rs` 에 `interview` 를 더해 실패(RED)를 봐요.
2. `src/spec_doc.rs`: `SECTIONS` 에 `Interview`, 파서(`interview` 함수), 상수 `INTERVIEW_FROM`(2026-11-19), 두 오류 코드.
3. 오래된 승인된 SPEC(`created` 2026-11-17)이 계속 깨끗한지, `spec check --all` 이 이 저장소에서 새 오류를 내지 않는지 확인.
4. 이웃 시험, clippy, fmt, mutation(≤12), SPEC `## Result`.

## Risks

- 기록은 자기가 쓴 글이라 검사는 모양만 봐요. 질문 없이 `- Q1: x → A: y` 를 써 넣을 수도 있어요. 이 조각은 "아예 흔적이 없는 승인"을 막을 뿐이고, 진짜 인터뷰였는지는 감사자가 대화와 비교해요(감사 brief 쪽은 카드의 다른 조각).
- 인터뷰 절은 digest 밖이라 승인 뒤에 고쳐도 `spec check` 는 모르고 git 기록에만 남아요.
- 기준일(2026-11-19)보다 일찍 만든 승인된 SPEC 은 인터뷰 없이 통과해요(기존 SPEC 을 깨지 않으려는 경과 조치). 기준일을 `created` 로 읽으므로 날짜를 거짓으로 쓰면 피할 수 있어요; 이 한계는 `created` 가 git 기록과 맞는지 보는 일(감사)에 달려 있어요.
- `none` 의 최소 길이 10자는 형식 요건일 뿐 이유의 질을 보지 않아요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 인터뷰 기록은 SPEC 본문의 `## Interview` 절 | 승인 근거(`approval.basis`)와 같은 파일에 있어 따로 읽을 곳이 없고, 파서가 이미 절을 읽어요 | 카드 note 에 두기(카드가 없는 저장소에서는 못 읽어요) |
| D-2 | 기준일 2026-11-19 이전에 만든 SPEC 은 제외 | 이미 승인된 SPEC 5개를 깨지 않아요 | 모든 승인된 SPEC 에 요구하고 5개에 사후 기록 추가(사실이 아닌 기록이 생겨요) |

## Result

구현 에이전트가 채웠어요. 감사가 다시 확인해요. macOS arm64, 이 worktree(`spec/20261119-enforce-interview-record`), 빌드 폴더 `~/.moai/build/moai-adk-86-s2`, `CARGO_INCREMENTAL=0`, cargo cache `/tmp/moai-rc-cargo-agent-3`. 시험 개수는 `cargo test --test <이름>` 을 파일마다 돌려 그 파일 로그의 `test result` 줄에서 읽었어요.

RED 먼저: 구현 전에 새 시험 4개(`interview_is_required`, `interview_entries`, `interview_section`, `interview_is_checked_by_the_command`)가 모두 실패하는 것을 봤어요. 구현 뒤 `interview_entries` 가 한 번 실패했는데 시험의 결함이었어요(`none` 이유 문자열이 8자·9자라서 "9자는 오류, 10자는 통과" 경계를 못 짚었어요): 글자 수를 정확히 9자·10자로 고쳐 통과했어요.

- AC-1: met — `cargo test --locked --offline --test spec_doc interview_is_required` 통과(approved·completed·superseded × created 2026-11-19·2026-11-20·2027-01-01 은 `interview_missing` 오류 하나, 기록을 더하면 조용; 2026-11-18, `draft`, 승인 전 상태 세 가지, Go 세대 SPEC, 달력에 없는 `created`(`created_invalid` 만)는 인터뷰를 요구하지 않아요).
- AC-2: met — `interview_entries` 통과(질문·답 줄과 `->` 형태, `none` 줄, 질문 안에 화살표가 든 줄, 줄 사이의 산문은 통과; 틀린 모양 여섯 가지는 승인된 SPEC 에서 `interview_entry` 오류, `draft` 에서 경고, `created` 가 옛날인 승인된 SPEC 에서도 오류; `none` 이유는 글자 수 9 는 오류, 10 은 통과).
- AC-3: met — `interview_section` 통과(`## Interview` 는 `section_unknown` 없이 알려진 절, 두 번 쓰면 `section_duplicate`, 인터뷰 유무가 `digest` 를 바꾸지 않고, 2026-11-17 에 만든 승인된 SPEC 은 그대로 조용해요).
- AC-4: met — `cargo test --locked --offline --test spec_cli interview` 통과(`moai spec check <id>` 가 기록 없는 승인된 SPEC 을 코드 1 과 메시지의 `interview_missing` 으로 거절하고, 거절은 아무것도 쓰지 않고, 기록을 더하면 오류 0 으로 통과).
- AC-5: met — `--test spec_doc` 27, `spec_cli` 18, `card_spec_cli` 11, `spec_contract` 8, `ac_verify_required` 6, `audit_admission` 7, `boundary_repro` 16, `plan_admission` 10, `spec_close_cli` 8, `ac_command_lint` 5, `git_env_scrub` 3, `--lib spec` 15 모두 통과, `cargo fmt --check` 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 0, `python3 -m unittest tests.ci_scope_test` OK. 이 저장소의 SPEC 에 새 오류가 없는지: 이 worktree 의 `moai spec check --all`(빌드한 Rust binary)이 SPEC 68개에서 오류 0·경고 5(기존 경고)를 내고, 통합 tip `origin/v2026.11.0` 의 SPEC 74개 중 기준일 이후에 만들어진 승인된 SPEC 은 없고, 열려 있는 `spec/20261119-*` 가지 29개의 SPEC 중에도 없어요(`git ls-tree`/`git show` 로 status·created·`## Interview` 줄을 읽어 셌어요). 돌리지 않은 것: 전체 `cargo test`(CI 가 돌려요), `moai spec check --all` 을 tip 위에서 실제로 돌리는 것(통합 때 리더가 돌려요).
- mutation: 코드에 12개를 돌려 11개 KILLED(기준일 당일 제외, 상태를 안 봄, id 형태를 안 봄, 이유를 바이트로 셈, 최소 9자, `interview_entry` 가 오류가 안 됨, `none` 과 다른 줄 허용, 빈 절 통과, `->` 안 읽음, 빈 답 통과, 첫 화살표만 인정), 하나(달력에 없는 날짜도 묻기)는 SURVIVED 였어요(`target/c86s2/mutants.txt`). 이유: `created` 는 달력에 있는 날짜일 때만 저장돼서(`spec_doc.rs` 의 frontmatter 읽기) 그 검사가 중복이었고, 동작이 같으므로 중복 코드를 지웠어요(지운 뒤 이웃 시험 통과). `git status` 깨끗이 복구.
- 이 조각 밖: PLAN 워크플로 문서(`skills/moai/workflows/plan.md`)와 SPEC 템플릿(`templates/spec/*.md`)에 인터뷰 절차를 적는 일은 쓰기 범위 밖이라 리더에게 제안만 해요(별도 조각). 질문 도구를 실제로 썼는지는 이 검사가 알 수 없어요(Risks).
