---
id: SPEC-20261119-reports-outside-repo
title: "Reports are saved outside the repository in the user-level MoAI home"
status: draft
revision: 1
created: 2026-11-19
---

# Reports are saved outside the repository in the user-level MoAI home

## Intent

사용자 지시(2026-11-19, 리더 대화): 로컬 보고서를 저장소 안에 두지 않아요. 기본 체크아웃에 보고서·스크립트·Issue 본문 사본 42개가 추적되지 않은 채 쌓였고(추적 파일은 137개 더 있어요) 개인적인 글이 들어갈 수 있는데 매일 커밋 대기 파일로 보였어요. 새 자리는 사용자 홈의 MoAI 폴더 아래 `~/.moai/reports/<project>/` 예요(`<project>` 는 worktree 경로가 쓰는 프로젝트 이름과 같아요).

이 SPEC 은 이 일의 카드가 가진 기준 1-4(경로 해석기·스캔 규칙·지침·결정 노트)예요. 기준 5(`/moai update` 가 옛 보고서를 옮김)는 이어진 SPEC `SPEC-20261119-update-moves-reports`(parent 가 이 SPEC)가 맡아요: 요구가 12개를 넘으면 나누라는 SPEC 형식 규칙 때문이고, 그 조각은 실제 HOME 에 쓰는 길이라 tier 2 감사가 따로 필요해서예요.

기존 해석기를 그대로 써요(새로 만들지 않아요): MoAI 홈은 `home_workdir::home()`/`root(home)`, 프로젝트 이름은 `worktree_name` 이 쓰는 `safe_name(<기본 체크아웃 폴더 이름>)` 이에요.

## Non-goals

- 추적 파일 137개 지우기: 삭제는 사용자 선택이라 이 SPEC 은 목록과 묶음별 제안만 결정 노트로 남겨요(AC-6). 지우는 일은 별도 카드예요.
- `.gitignore` 에 `.moai-cli/reports/` 더하기: 카드 4번 항목이 "추적 파일 처리 결정 뒤"라고 정했어요. 사용자 선택이 난 뒤 후속 카드에서 해요.
- 옛 보고서 옮기기(`/moai update`): 이어진 SPEC 이 해요.
- 기본 체크아웃의 다른 호스트 폴더 정리(따로 있는 카드), Go 세대 도구가 쓰는 `~/.moai/reports` 의 옛 파일 손대기.

## Requirements

- REQ-1: The function `home_workdir::reports_dir` shall return `<HOME>/.moai/reports/<project>` where `<project>` is the safe form of the primary checkout folder name that `worktree_name` uses.
- REQ-2: The function `home_workdir::reports_dir` shall refuse a project folder whose name has no usable characters and create nothing.
- REQ-3: The HOME name table shall list the name `reports` as a folder this generation writes and keeps as user data.
- REQ-4: The command `optimize-plan` shall list the project folder under the HOME reports folder as a row that is kept and never trimmed.
- REQ-5: The file `AGENTS.md` shall state in one sentence the place `<HOME>/.moai/reports/<project>/` and that a session never writes a report under the repository.
- REQ-6: The moai skill shall state the same place in one sentence and carry the new hash of every changed skill file in the source manifest of the same commit.
- REQ-7: The repository scans and the CI scope rule that name `.moai-cli/reports/` shall keep covering the tracked files under that folder until a later card removes them.
- REQ-8: A decision note on this change's card shall list every tracked file under `.moai-cli/reports/` by group with a keep or remove proposal while this change removes no tracked file.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 HOME 에서 `reports_dir` 가 `<HOME>/.moai/reports/<프로젝트>` 를 주고(공백·한글 같은 문자는 `-` 로 바뀌어요) 쓸 수 있는 문자가 없는 폴더 이름은 거절하며 아무 폴더도 만들지 않아요.
  verify: `cargo test --locked --offline --lib reports_dir`
- AC-2 (REQ-3): HOME 이름표의 `reports` 행이 이 세대가 쓰는 사용자 데이터(보존)로 바뀌고 옛 Go 폴더 분류 시험이 새 값을 지켜요.
  verify: `cargo test --locked --offline --lib home_layout`; `cargo test --locked --offline --lib home_retire`
- AC-3 (REQ-4): `optimize-plan` 이 임시 HOME 의 `reports/<프로젝트>` 폴더를 `kind: report`·`action: keep` 행으로 보여 주고 `trim_paths` 는 비어 있어요.
  verify: `cargo test --locked --offline --test optimize_plan_cli`
- AC-4 (REQ-5, REQ-6): `AGENTS.md` 와 moai 스킬이 새 자리를 한 문장씩 말하고 스킬 패키지 시험이 매니페스트 해시를 확인해요.
  verify: `cargo test --locked --offline --test product_cli skill_package`; `python3 -m unittest tests.source_manifest_test tests.skill_references_test`
- AC-5 (REQ-7): 추적된 보고서 파일이 남아 있는 동안 스캔·CI 범위 규칙이 계속 그 파일들을 덮고, 세 시험 모듈이 통과해요.
  verify: `python3 -m unittest tests.repo_scans_test tests.ci_scope_test tests.local_check_test`
- AC-6 (REQ-8): 결정 노트가 추적 파일 137개를 묶음별로 나열하고(개수·크기·제안) 이 변경으로 지워진 추적 파일이 0개예요.
  manual: 이 카드의 결정 노트 번호를 Result 에 적고 `git ls-files .moai-cli/reports | wc -l` 이 137 임을 보인다.

## Scope

- write: `.moai/specs/SPEC-20261119-reports-outside-repo/spec.md`, `src/home_workdir.rs`, `src/home_layout.rs`, `src/home_retire.rs`(옛 분류 시험 한 줄), `src/optimize.rs`, `tests/optimize_plan_cli.rs`, `docs/design-review/appendix/operations/optimize_plan.md`, `AGENTS.md`, `skills/moai/SKILL.md`, `skills/moai/references/source-manifest.json`
- read-only: `src/state_root.rs`, `src/go_state.rs`, `tools/ci_scope.py`, `tools/check_ac_commands.py`, `tools/repo_scans.json`
- forbidden: 실제 HOME 에 쓰는 시험, 네트워크, 추적 파일 삭제, `.gitignore` 수정

## Plan

1. SPEC 을 쓰고 `spec check` 로 확인한 뒤 `card spec` 으로 카드에 연결해요.
2. 결정 노트(AC-6): 추적 파일 137개를 묶음별로 읽기 전용으로 세어 카드에 남겨요.
3. RED: `reports_dir` 시험, HOME 이름표 행 시험, `optimize-plan` 시험을 먼저 쓰고 실패를 눈으로 봐요.
4. GREEN: `home_workdir.rs`(프로젝트 이름 규칙을 `worktree_name` 과 한 함수로 나눠요), `home_layout.rs`, `optimize.rs`와 그 문서 조각.
5. 지침: `AGENTS.md`·`SKILL.md` 한 문장씩, 매니페스트 해시를 같은 커밋에서 갱신해요.
6. 시험을 돌려요: 위 verify 명령들, fmt, ops rustfmt, clippy `--all-targets`, `--bins`(기존 실패 5개가 그대로인지 따로 알려요).

## Risks

- 홈의 `reports` 가 지금은 "주인 모르는 Go 폴더(보존)"로 분류돼요: 행을 바꿔도 보존 규칙은 같아서 `home-retire` 가 지우거나 옮기지 않아요(AC-2 가 지켜요).
- 홈 이전(`home-state-plan`/`home-state-apply`)이 Go 세대 홈 폴더를 통째로 이름 바꿔 백업으로 옮길 때 `reports/` 는 지금 "그냥 항목"이라 사용자 보고서가 `.moai-backup` 으로 따라가요(지우지는 않고, 사용자 승인 뒤 적용이에요). `worktrees`·`projects` 처럼 작업 폴더로 만들면 Go 세대 홈의 옛 `reports` 때문에 통째 이전이 막히는 부작용이 있어서 이 조각에서는 바꾸지 않아요(결정 D-6). 후속 카드로 리더가 정해요.
- `optimize-plan` 행이 늘어서 그 출력을 읽는 시험이나 골든이 달라질 수 있어요: 먼저 찾아 같이 고쳐요.
- 이 저장소의 설치된 `moai` 는 옛 빌드라서 카드 노트에 옵션-등호 모양을 적으면 거절돼요: 노트는 풀어 써요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 프로젝트 이름은 `worktree_name` 과 같은 규칙(`safe_name(기본 체크아웃 폴더 이름)`)을 한 함수로 나눠 둘이 함께 쓴다 | 카드 본문: 기존 해석기를 쓰고 두 번째를 만들지 않는다 | 함수를 되돌려 두 곳에 두기 |
| D-2 | HOME 이름표 `reports` 는 사용자 데이터·보존으로 바꾼다(홈에서 지우거나 옮기는 규칙은 없어요) | 이 세대가 직접 쓰는 폴더라 "주인 모르는 폴더"가 틀린 표시예요 | 행을 옛 값으로 |
| D-3 | 스캔·CI 범위 규칙은 `.moai-cli/reports/` 를 계속 덮는다: 추적 파일이 남아 있는 동안은 그 파일의 문서 규칙이 유효해요. 새 자리는 저장소 밖이라 스캔·CI 가 볼 일이 없어요 | 카드 기준 2 의 "갱신"은 추적 파일 처리(사용자 선택) 뒤에 규칙을 지우는 쪽이에요; 지금 지우면 추적된 보고서의 AC 명령 검사가 멈춰요 | 사용자 선택 뒤 후속 카드에서 규칙 삭제 |
| D-4 | `optimize-plan` 은 저장소 안 옛 보고서 행을 그대로 두고 HOME 행을 더한다 | 추적 파일이 남는 동안 두 자리 모두 공간을 차지해요 | HOME 행 빼기 |
| D-5 | `.gitignore` 규칙은 추적 파일 사용자 선택 뒤 후속 카드에 둔다 | 카드 본문 4번 항목의 순서 | 지금 더하기 |
| D-6 | `src/home_state.rs` 의 작업 폴더 목록(`is_work_folder`)과 선택 표는 바꾸지 않는다 | 목록에 더하면 Go 세대 홈의 옛 `reports` 가 통째 이전을 막아요(`work_folder_present:reports`); 그 영향은 이 조각에서 재지 않았어요 | 후속 카드에서 측정한 뒤 더하기 |

## Result

구현됨(2026-11-20, 가지 `spec/20261119-reports-outside-repo`).

- AC-1: met — `cargo test --lib reports_dir` 통과(전체 `--lib` 종료 코드 0, head `0a6b335558`)
- AC-2: met — `--lib home_layout`·`--lib home_retire` 와 `tests/home_retire_cli.rs`(24개) 통과
- AC-3: met — `tests/optimize_plan_cli.rs` 통과(영역 대상 실행 종료 코드 0)
- AC-4: met — `product_cli skill_package` 와 `tests.source_manifest_test`·`tests.skill_references_test` 종료 코드 0
- AC-5: met — `tests.repo_scans_test`·`tests.ci_scope_test`·`tests.local_check_test` 종료 코드 0
- AC-6: met — 결정 노트 3280 이 137개를 묶음별로 나열했고 `git ls-files .moai-cli/reports | wc -l` 이 137, 지운 추적 파일 0개

- 보고서 자리는 `home_workdir::reports_dir` 가 `<HOME>/.moai/reports/<프로젝트>` 로 풀어요(AC-1). HOME 이름표의 `reports` 행은 이 세대가 쓰는 사용자 데이터(보존)이고 옛 분류 시험 두 곳(`src/home_retire.rs` 의 단위 시험, `tests/home_retire_cli.rs`)이 새 값을 지켜요(AC-2). `optimize-plan` 은 그 폴더를 `report`·`keep` 행으로 보여 주고 `trim_paths` 는 비어 있어요(AC-3).
- `AGENTS.md` 는 `## Reporting` 의 겹친 두 문단을 항목 하나도 빼지 않고 한 문단으로 합쳐 9201 bytes(한도 9216)로 맞추고 한 문장 `Reports go to ~/.moai/reports/<project>/, never into the repo.` 를 더했어요. moai 스킬에도 한 문장을 더했고 매니페스트 해시를 같은 커밋에서 갱신했어요(AC-4).
- 스캔·CI 범위 규칙은 바꾸지 않았고 추적된 보고서 파일이 남아 있는 동안 계속 덮어요. `tests.repo_scans_test`·`tests.ci_scope_test`·`tests.local_check_test` 통과(AC-5).
- 추적 파일 137개의 묶음 제안은 카드 노트 3280 에 있고, 이 가지는 추적 파일을 하나도 지우지 않았어요(`git diff --diff-filter=D` 0개, `git ls-files .moai-cli/reports .moai/reports` 137개 그대로)(AC-6).
- 증거는 `SPEC-20261119-update-moves-reports` 의 Result 에 같은 줄을 적었어요(head `0a6b335558`의 전체 검사 종료 코드). 돌리지 않은 것: 병합된 트리에서의 골든 실행, CI.
