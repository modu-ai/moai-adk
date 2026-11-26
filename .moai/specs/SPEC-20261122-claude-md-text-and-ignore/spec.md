---
id: SPEC-20261122-claude-md-text-and-ignore
title: "Shipped and generated text calls the instruction files CLAUDE.md and CLAUDE.local.md"
status: draft
revision: 3
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Shipped and generated text calls the instruction files CLAUDE.md and CLAUDE.local.md

## Intent

moai 의 지침 파일은 `CLAUDE.md`(팀)와 `CLAUDE.local.md`(개인)예요. 지금도 `/moai project` 의 언어 질문과 몇몇 도움말·안내 문장이 `AGENTS.md` 를 moai 의 파일처럼 불러요. 그 문장을 고치고, 같은 문제가 다시 들어오면 실패하는 저장소 글 테스트를 둬요. 사용자의 예전 `AGENTS.md`·`AGENTS.local.md` 를 옮기고, 읽고, 백업하고, 되돌리는 코드와 그 안내는 예전 이름을 그대로 쓰고, 테스트가 그 예외를 줄 단위로 이유와 함께 적어요. 카드의 남은 기준(배포되는 글이나 만들어지는 파일이 `AGENTS.md` 를 moai 의 파일로 부르지 않음, 저장소 글 테스트)을 이 개정이 맡아요.

## Non-goals

- 개정 1 의 REQ-1~REQ-4(무시 블록, 업데이트 미리보기 경고, 이 저장소의 `.gitignore` 허용 목록)와 REQ-7(배너)은 이 개정에서 빼요(D-10).
- 이전 동작은 바꾸지 않아요: 지금도 새로 `AGENTS.md`·`AGENTS.local.md` 를 만드는 코드는 없고, 되돌리기만 예전 파일을 복원해요.
- 배포되지 않는 글(`docs/`, `.github/`, 저장소 도구 파일, 코드 주석)은 고치지 않아요. 바뀌는 제안 문장을 설명하는 `docs/claude-adapter-contract.md` 의 그 대목만 고쳐요.

## Interview

- Q1: 무엇을 고치나요? -> A: 사용자(2026-11-26, 카드 노트 10213, 말 그대로) "/moai project 설정시 AGENTS.md 오류 표기 등 AGETNS.md 관련 모든 내요을 수정해야한다. 기존의 카드가 있으면 패스 없으면 추가"
- Q2: moai 의 지침 파일 이름은 무엇인가요? -> A: 사용자(2026-11-22, 카드 노트 6484, 말 그대로) "AGENTS.md 제거해라. CLAUDE.md CLAUDE.local.md 만 사용한다"

## Requirements

- REQ-1: The instruction-file label shall name `CLAUDE.md` and `CLAUDE.local.md` in both languages and name no old instruction file.
- REQ-2: When `generation-apply` answers its next actions, the result shall name the project instructions and no old instruction file as what the new session loads.
- REQ-3 (REQ-2): IF the apply wrote no `CLAUDE.md` because the user's old file stays, THEN the next action shall claim no newly written instruction file.
- REQ-4: The update preview shall name `CLAUDE.local.md` in its tracked-or-ignored reason, `CLAUDE.md` in its undo-order note and no project instruction file in the reason of a not-supported user instruction row.
- REQ-5: The suggestions of `claude-doctor` for a suppressed `AGENTS.md` shall advise moving its text into `CLAUDE.md` and never an `@AGENTS.md` import or a setting that reads both files.
- REQ-6: The command help shall name `CLAUDE.md` for `agents-plan` and `agents-apply` and call the file that `claude-doctor` checks the user's own `AGENTS.md`.
- REQ-7: The shipped skill texts shall name `CLAUDE.md` as the file the install step writes and name no retired host's instruction file.
- REQ-8: When the repository text test reads a shipped line that names an old instruction file outside a listed exception, the test shall fail with the path and the line.
- REQ-9 (REQ-8): IF a scanned folder cannot be read, an exception's count differs from its row, a row has no reason or its fragment is only a bare file name, THEN the test shall fail naming it.
- REQ-10: The exception list shall hold one row per line fragment, with its exact count and its reason, only for text that moves, reads, backs up, guards, undoes or advises about the user's old files.
- REQ-11: When a temporary project is set up in English or Korean, the files moai writes in it and under its HOME shall name no old instruction file.
- REQ-12 (REQ-11): IF a setup command fails or writes no instruction file, THEN the test shall fail.

"배포되는 줄"은 Plan 1 이 정해요: `src/` 의 줄 주석과 `#[cfg(test)]` 모듈은 빼고, 명령 도움말이 되는 문서 주석은 넣어요.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-11, REQ-12): 임시 HOME 과 임시 프로젝트를 영어와 한국어로 설정하면 쓰인 파일, 설정 질문, 다음 할 일 문장에 예전 이름이 없고 새 라벨이 쓰여요.
  verify: `cargo test --locked --offline --test claude_md_layout_scan generated_project`
- AC-2 (REQ-3): 예전 파일을 그대로 둔 프로젝트에서 다음 할 일 문장이 새 지침 파일을 썼다고 말하지 않아요.
  verify: `cargo test --locked --offline --test claude_md_layout_scan generated_next_action`
- AC-3 (REQ-4…REQ-8, REQ-10): 배포되는 줄의 예전 이름은 목록의 행에만 있고, 고친 문장들이 새 이름을 담아요.
  verify: `cargo test --locked --offline --test claude_md_layout_scan shipped_`
- AC-4 (REQ-8, REQ-9): 심어 둔 줄, 틀린 개수, 잘못된 행, 읽을 수 없는 폴더에서 검사가 실패해요.
  verify: `cargo test --locked --offline --test claude_md_layout_scan planted_`
- AC-5 (REQ-6): 도움말 출력이 골든 파일과 같아요.
  verify: `cargo test --locked --offline --test help_layout`
- AC-6 (REQ-7): 스킬 패키지 목록(source manifest)의 해시와 고정 문장이 맞아요.
  verify: `cargo test --locked --offline --test source_manifest_facts`
  verify: `cargo test --locked --offline --test go_state_classes_cli`
- AC-7 (REQ-1, REQ-4): 라벨과 되돌리기 순서의 기존 고정 테스트가 통과해요.
  verify: `cargo test --locked --offline --test language_setting`
  verify: `cargo test --locked --offline --test project_harness_label`
- AC-8 (REQ-4): 고친 미리보기 픽스처로 업데이트 화면 테스트가 통과해요.
  verify: `cargo test --locked --offline --test update_tux_golden`
- AC-9 (REQ-8): 새 테스트가 저장소 검사 표에 올라 있고 읽는 곳이 덮여요.
  verify: `cargo test --locked --offline --test repo_scans_table every_`
- AC-10 (REQ-5): 예전 파일이 가려진 임시 프로젝트에서 `claude-doctor` 가 내용을 `CLAUDE.md` 로 옮기라고 권하고, import 나 두 파일 함께 읽기 설정은 권하지 않아요.
  verify: `cargo test --locked --offline --test product_cli claude_doctor_agents_loading_combinations`

## Scope

- write: `.moai/specs/SPEC-20261122-claude-md-text-and-ignore/**`, `tests/claude_md_layout_scan.rs`(새 파일), `tools/repo_scans.json`(한 행), `src/language.rs`, `src/generation.rs`, `src/project_update.rs`, `src/home_routing.rs`, `tests/fixtures/update_tux/preview-blocked.json`, `tests/fixtures/update_tux/preview-changes.json`, `tests/fixtures/update_tux/preview-latest.json`, `src/ops/agents_apply.rs`, `src/ops/agents_plan.rs`, `src/ops/claude_doctor.rs`, `src/help.rs`, `tests/fixtures/help/all.txt`, `tests/fixtures/help/all-60.txt`, `tests/fixtures/help/all-80.txt`, `tests/fixtures/help/all-dumb.txt`, `tests/fixtures/help/root.txt`, `tests/fixtures/help/root-60.txt`, `tests/fixtures/help/root-80.txt`, `tests/fixtures/help/root-dumb.txt`, `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`, `src/claude.rs`(C-336 의 READY 뒤에 범위를 넓혀요), `tests/product_cli.rs`, `docs/claude-adapter-contract.md`
- 공유 파일(line-lease 로만 씀): `tools/repo_scans.json`, `src/help.rs`, `tests/fixtures/help/*.txt`, `skills/moai/references/source-manifest.json`
- read-only: `src/agents.rs`, `src/agents_local_merge.rs`, `src/agents_local_merge/**`
- forbidden: `.gitignore`, `CLAUDE.md`, `CLAUDE.local.md`, 실제 HOME, 네트워크

## Plan

1. `tests/claude_md_layout_scan.rs` 를 먼저 써서 실패를 봐요. 검사 대상은 `skills/`, `agents/`, `templates/`, `schemas/`, `install/`, `README.md`, `README.ko.md`, `src/**`. 찾는 이름은 `AGENTS.md`, `AGENTS.local.md`, `AGENTS.override.md`, `AGENTS files` 와, `src/agents.rs`·`src/agents_local_merge*` 에서는 이름을 넣는 상수 표기(`{AGENTS_FILE}`, `{LEGACY_TARGET}`, `{LOCAL}`). `src/` 의 `//`·`//!` 줄은 빼고, `src/ops/**`·`src/main.rs`·`src/help.rs`·`src/ui.rs` 의 `///` 줄은 읽어요. `#[cfg(test)]` 뒤의 `mod x {` 블록만 건너뛰고, 모듈이 아닌 항목에서는 건너뛰기를 시작하지 않아요. `tools/repo_scans.json` 에 자기 행 하나를 더해요.
2. 라벨(`src/language.rs`): `Instruction files (CLAUDE.md, CLAUDE.local.md, rules)`, `지침 파일(CLAUDE.md, CLAUDE.local.md, 규칙)`.
3. `src/generation.rs` 다음 할 일: "start a new host session in this project so it loads the project instructions and the generated files". `src/project_update.rs` 의 두 문장.
4. `src/home_routing.rs` Cursor 이유: "Cursor CLI has no user-level instruction file (rules live in the app); not written", 픽스처 세 개.
5. 명령 도움말(`src/ops/*`, `src/help.rs`)과 골든 파일 8개(`MOAI_HELP_GOLDEN_WRITE=1` 로 다시 씀), `skills/moai/workflows/project.md:38`("writes a new `CLAUDE.md`"), `source-manifest.json` 의 옛 Codex 이유 문구와 해시(`moai check fix`).
6. `src/claude.rs` 의 제안 두 문장을 하나로 다시 써요: "move the text of your AGENTS.md into CLAUDE.md (for a file of this project, /moai update offers the pick `move`)". `tests/product_cli.rs` 의 그 테스트가 새 문장을 확인하고, `docs/claude-adapter-contract.md` 의 설명 한 대목을 맞춰요.
7. 임시 프로젝트 생성 파일 테스트, 가드 뮤턴트, `moai check affected --card`, READY. 착지하면 리더가 스킬 사본을 맞춰요.

## Decisions

각 결정의 기본값이 이 개정의 내용이에요. 승인 때 바꿀 수 있어요.

| Id | Decision | Basis | Undo / alternative |
| --- | --- | --- | --- |
| D-1 | 글 테스트는 Rust 테스트예요. | Python 테스트 금지(2026-11-21). | 없음. |
| D-2 | 예외는 테스트 파일 안의 (경로, 줄 조각, 정확한 개수, 이유) 행이에요. 파일 단위 건너뛰기는 쓰지 않아요. | 예외 파일에 새 문장이 몰래 들어오지 못해요. | 경로 단위 목록. |
| D-3 | 검사 대상과 주석·테스트 코드 규칙은 Plan 1 이에요. | 배포되는 것과 도움말이 되는 것만 봐요. | 대상 축소. |
| D-4 | Cursor 이유 문장을 여기서 고쳐요. `src/home` 접두라 이 SPEC 은 tier 2 예요. | 지금 문장은 프로젝트 `AGENTS.md` 가 규칙을 나른다고 말해요. 리더 결정(노트 10248). | 별도 카드로 빼고 날짜 있는 예외 행. |
| D-5 | `claude-doctor` 는 동작을 그대로 두고 "your own AGENTS.md" 라고 불러요. | 그 명령은 실제로 그 파일을 확인해요. | 문장을 그대로 두고 예외 행. |
| D-6 | `src/claude.rs` 의 제안 두 문장(`@AGENTS.md` import 추가, 두 파일 함께 읽기 설정)은 "내용을 `CLAUDE.md` 로 옮기기" 한 문장으로 다시 써요. | 리더 결정(노트 10248): 사용자 요청(노트 10213)과 노트 6484("@AGENTS.md 줄 없음")에 맞아요. 그 파일은 C-336 이 예약 중이라 그 READY 뒤에 범위를 넓혀요. | 예외 행으로 남기기. |
| D-7 | 예전 파일을 지키거나 검토하는 줄(`src/scope.rs`, `src/gate.rs`, `src/review.rs`, 감사 글, 프로토콜 스키마 설명)은 "읽기" 예외예요. | 지우면 보호 목록이 줄어요. | 지우기. |
| D-8 | 다음 할 일 문장은 "the project instructions" 예요. | `CLAUDE.md` 를 쓰지 않은 경우에도 맞아요. | "CLAUDE.md". |
| D-9 | 라벨은 두 파일 이름을 다 적어요. 이 설정이 있는 프로젝트는 다음 업데이트에서 관리 블록이 한 번 바뀌어요. | 팀·개인 파일이 모두 그 언어 설정의 대상이에요. | `CLAUDE.md` 만. |
| D-10 | 개정 1 의 REQ-1~4·7 은 빼고, 무시 파일은 리더가 후속 카드로 열어요(노트 10248). 배너는 지금 글이 형식을 지켜 새 일이 없어요. | 저장소 `.gitignore` 허용 목록은 `CLAUDE.md` 의 "`.claude/` 는 스테이징하지 않는다" 와 부딪치고, 결정 기록에서 그것을 푸는 행을 찾지 못했어요. layout SPEC 의 조각 4 는 아직 이 SPEC 을 가리켜요. | 이 개정에 남기기(그 결정이 먼저 필요해요). |
| D-11 | 저장해 둔 업데이트 미리보기 픽스처 세 개도 새 문장으로 고쳐요. | 제품 문장과 같게 해요. | 그대로 두기. |

## Risks

- R-1: `src/project_update.rs` 는 C-330·C-569·C-572, `skills/moai/workflows/project.md` 는 C-330·C-569, `src/claude.rs` 는 C-336 이 예약 중이에요(2026-11-26 읽음). 순서를 따로 세우지 않아요: 먼저 READY 를 받은 카드가 먼저 착지하고, 뒤에 착지하는 쪽이 합쳐요(노트 10248).
- R-2: 예외 줄의 문장을 바꾸되 조각이 남고 이름이 늘지 않으면 테스트가 잡지 못해요.
- R-3: Claude Code 와 Cursor 가 어떤 지침 파일을 읽는지는 이 개정에서 새로 확인하지 않았어요.

## Result

구현했어요(브랜치 `spec/claude-md-text-and-ignore-20261122`, 기반 `0aa62fd249`, 코드 head `33051a6bf53638a75a5fc3b62af4597385b4d0de`). 커밋은 red 테스트 `fcca95a7b6`, 글 수정 `6f8dc9f879`, 소스 매니페스트 해시 `33051a6bf5` 예요. AC 명령은 적힌 그대로, `6f8dc9f879` 를 커밋하기 직전에 제품 수정이 커밋되지 않은 작업 트리에서 돌렸어요. AC-6 의 두 명령만 해시를 고친 `33051a6bf5` 내용에서 다시 돌렸어요.

- RED: 제품 수정 전의 `fcca95a7b6` 내용에서 새 테스트 6개 가운데 심어 둔 줄 테스트(`planted_`) 2개는 통과했고 4개(`shipped_` 2개, `generated_` 2개)는 실패했어요.
- 관련 테스트 대상 26개의 nextest 실행: 324개 실행, 323개 통과, 1개 실패(옛 `project.md` 해시).
- 바뀐 것: 지침 파일 라벨(영어, 한국어), `generation-apply` 의 다음 할 일 문장, 업데이트 미리보기의 두 문장과 저장해 둔 픽스처 3개, Cursor 이유 문장, `agents-plan`·`agents-apply`·`claude-doctor` 도움말과 골든 파일, `claude-doctor` 제안(한 문장), `skills/moai/workflows/project.md` 와 `source-manifest.json`(이유 문구, 해시), `docs/claude-adapter-contract.md` 의 해당 대목, 새 테스트 `tests/claude_md_layout_scan.rs`, `tests/product_cli.rs` 의 단언.

- AC-1: met — `cargo test --locked --offline --test claude_md_layout_scan generated_project` 1 passed
- AC-2: met — `cargo test --locked --offline --test claude_md_layout_scan generated_next_action` 1 passed
- AC-3: met — `cargo test --locked --offline --test claude_md_layout_scan shipped_` 2 passed
- AC-4: met — `cargo test --locked --offline --test claude_md_layout_scan planted_` 2 passed
- AC-5: met — `cargo test --locked --offline --test help_layout` 11 passed
- AC-6: met — `33051a6bf5` 내용에서 `cargo test --locked --offline --test source_manifest_facts` 8 passed, `cargo test --locked --offline --test go_state_classes_cli` 71 passed(`source_manifest_facts` 는 해시를 고치기 전에 옛 `project.md` 해시로 한 번 실패했어요)
- AC-7: met — `cargo test --locked --offline --test language_setting` 9 passed, `cargo test --locked --offline --test project_harness_label` 6 passed
- AC-8: met — `cargo test --locked --offline --test update_tux_golden` 17 passed
- AC-9: met — `cargo test --locked --offline --test repo_scans_table every_` 8 passed
- AC-10: met — `cargo test --locked --offline --test product_cli claude_doctor_agents_loading_combinations` 1 passed

가드 뮤턴트: `33051a6bf5` 에서 7개 가운데 7개가 exit 100 이고 각각 지목한 테스트에서 실패했어요(새 테스트마다 하나와 `claude-doctor` 단언 하나). 줄은 READY 노트에 있어요.

계획에 더한 것:

- `tools/repo_scans.json` 의 새 행은 scope prefix 에 `CLAUDE.md` 와 `REVIEW.md` 도 적어요. 테스트가 그 이름들을 읽기 때문에 `repo_scans_table` 이 요구했어요. 그 읽기는 임시 프로젝트의 파일이고, 이 테스트는 이 저장소의 `CLAUDE.md` 나 `REVIEW.md` 를 검사하지 않아요.

알려 둘 것:

- `src/project_update.rs` 와 `src/ops` 의 두 파일에 더한 존재 단언은 제품 수정 뒤에 썼어요. 실패를 본 것은 `src/project_update.rs` 쪽(`shipped_rewritten_sentences_name_the_new_files` 의 위 뮤턴트)뿐이에요.
- 설정의 대화형 터미널 출력이 새 라벨을 찍는지는 확인하지 않았어요. 생성 파일 테스트가 라벨을 확인하는 곳은 설정 질문 글과 쓰인 파일이에요.
- 공유 파일(`src/help.rs`, 도움말 골든 파일, `source-manifest.json`, `tools/repo_scans.json`)은 카드 line-lease 로 썼어요.
- `src/claude.rs` 는 아직 착지하지 않은 다른 카드도 바꿔요. 뒤에 착지하는 쪽이 합쳐요(R-1).
- SPEC front matter 는 `status: draft` 이고 `approval` 블록이 없어요. 승인은 노트 10251 이고, 블록은 착지 때 리더가 써요.
- `moai check affected` 영수증은 이 Result 가 아니라 카드 노트에 적어요.
- 스킬 사본은 착지 뒤에 리더가 맞춰요.

not_run:

- 전체 `cargo test`, Windows, Linux.
- 실제 Claude Code 세션이 이 파일들을 읽는지(R-3).
