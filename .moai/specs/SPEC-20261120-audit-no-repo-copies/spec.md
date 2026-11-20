---
id: SPEC-20261120-audit-no-repo-copies
title: "Audits and experiments never copy a whole repository; a guard flags copies in the reports and evidence folders"
status: draft
revision: 1
created: 2026-11-20
---

# Audits and experiments never copy a whole repository; a guard flags copies in the reports and evidence folders

## Intent

사용자 결정(2026-11-20, 리더 대화): 감사와 실험은 저장소를 통째로 복사하지 않아요. 한 프로젝트의 reports 폴더에 PR 감사 때 만든 저장소 사본 67개(64G, 그 안 `node_modules` 28.4G)가 쌓였어요. 사본은 `git worktree add --detach`(객체를 함께 써요)나 임시 폴더의 깨끗한 clone 으로 만들고, 끝나면 지워요. 보고서·증거는 프로젝트 밖에 두고 추적하지 않아요.

이 SPEC 은 코드 쪽 두 가지만 해요. 지침 글(감사 에이전트 지침, 감사·검증 워크플로, OPTIMIZE 워크플로)은 같은 파일을 쥔 다른 변경이 끝난 뒤에 따로 넣어요(Non-goals).

1. **가드**: `moai optimize-plan` 이 목록에 올리는 reports 폴더 행(`kind: report`)에 `findings` 를 달아요. 폴더 이름이 `node_modules`, `.venv`, `target`, `.next`, `.git` 인 곳(저장소 사본의 표시)과, 카드 하나의 증거 폴더가 100 MiB 를 넘는 곳이에요. 일반 폴더는 아무것도 달지 않아요. 가드는 아무것도 지우거나 옮기지 않고, 링크를 따라가지 않고, 표시된 폴더 안으로 들어가지 않아요(크기를 세는 비용을 피해요).
2. **감사 도구 안내**: `moai audit-clone-plan`(감사 clone 계획)의 답에 규칙을 달아요: 통째로 복사하지 않기, 별도 worktree 나 임시 clone 만 쓰기, 증거 폴더에는 로그만 두고 트리를 넣지 않기(100 MiB 넘는 파일은 경로만 적기), 끝나면 지우기.

## Non-goals

- 지침 글 변경(`agents/audit.md`, `skills/moai/workflows/audit.md`, `verify.md`, `optimize.md`, 매니페스트): 그 파일들을 쥔 변경이 끝나면 같은 규칙 문장을 그쪽에 넣어요. 이 SPEC 의 시험은 그 글을 보지 않아요.
- 지우기·옮기기 적용: 가드는 목록에 표시만 하고, 정리는 기존 미리보기 → 사용자 선택 → 적용 흐름을 써요.
- 프로젝트 밖 홈 보고서 폴더(`~/.moai/reports/<project>/`)의 스캔: 그 자리를 만드는 변경 뒤에 같은 함수를 부르면 돼요.
- 증거 폴더에 파일을 넣을 때 막는 훅: 이번 조각은 표시까지만 해요.

## Requirements

- REQ-1: When `moai optimize-plan` lists a reports folder, the row shall carry a finding for each folder named `node_modules`, `.venv`, `target`, `.next` or `.git` found anywhere below it. · failure: none
- REQ-2: When a folder directly below the reports folder holds more than 100 MiB, counted as the sum of the logical sizes of all files in it and its nested folders, the row shall carry a finding that names that folder, its size and the cap. · failure: none
- REQ-3: When a reports folder holds none of those folders and no folder over the cap, the row shall carry an empty `findings` list and the reasons it had before. · failure: none
- REQ-4 (REQ-1): IF a finding is raised, THEN the row shall add the reason `evidence_guard` and keep the action `keep`.
- REQ-5: While the guard scans, it shall follow no link and enter no folder it has flagged. · failure: none
- REQ-6: When `moai audit-clone-plan` answers, it shall add the rules for audit copies and the evidence cap. · failure: none
- REQ-7 (REQ-1): IF the guard cannot read a folder or an entry, or the deadline passes, THEN the row shall set `findings_partial` to true, so that an empty `findings` list is read as "nothing seen".

## Acceptance

- AC-1 (REQ-1, REQ-4, REQ-5): a fixture reports folder with a repository copy (`.git`, `node_modules`) is flagged, the row keeps its action, and a linked `node_modules` is flagged but not entered.
  verify: `cargo test --locked --offline --test audit_copy_guard flags_`
- AC-2 (REQ-2, REQ-3): a card folder over the cap is flagged with its size, also when no single file is over it; a normal folder passes with an empty list.
  verify: `cargo test --locked --offline --test audit_copy_guard passes_`
  verify: `cargo test --locked --offline --test audit_copy_guard adds_up_`
- AC-5 (REQ-7): a folder the guard cannot read sets `findings_partial` and the cards after it are still checked.
  verify: `cargo test --locked --offline --test audit_copy_guard says_so_`
- AC-3 (REQ-6): the plan carries the rules and the cap.
  verify: `cargo test --locked --offline --test audit_clone_cli rules_`
- AC-4: neighbours.
  verify: `cargo test --locked --offline --test optimize_plan_cli --test audit_clone_cli --test audit_copy_guard`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `.moai/specs/SPEC-20261120-audit-no-repo-copies/**`, `src/optimize.rs`, `src/audit_clone.rs`, `tests/audit_copy_guard.rs`, `tests/optimize_plan_cli.rs`, `tests/audit_clone_cli.rs`
- read-only: `src/state_root.rs`, `src/hooks.rs`
- forbidden: any deletion or move by the guard, any write in the real HOME, a new card, Issue or pull request number in the text this change adds

## Plan

1. SPEC, then RED tests (`tests/audit_copy_guard.rs`, one case in `tests/audit_clone_cli.rs`).
2. `src/optimize.rs`: the guard function and the `findings` on the report row. `src/audit_clone.rs`: the rules.
3. Checks, mutants, a read-only audit by another agent.

## Risks

- A project that keeps a folder named `target` or `.git` on purpose under its reports folder gets a finding; the finding only lists it, nothing is removed.
- The scan reads folder names only (not file contents), so a copied tree under a different folder name is found only through its `.git`, `node_modules` and similar folders.
- The 100 MiB cap counts file sizes under one card folder and stops counting at the cap, so a huge folder costs little.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The guard lives in the OPTIMIZE preview, not in a new command | The preview already lists reports and is read-only; no new word in the binary | Move the function |
| D-2 | A finding never changes the action | Deleting stays a guarded choice after the user's pick | none needed |

## Result

`moai optimize-plan` 이 이제 reports 폴더 행에 `findings` 를 달고(`forbidden_folder`: `node_modules`·`.venv`·`target`·`.next`·`.git` 이름의 폴더, 링크, worktree 의 `.git` 파일 / `over_cap`: 카드 폴더가 100 MiB 를 넘을 때 `bytes_at_least` 와 `cap_bytes`), 찾으면 이유 `evidence_guard` 를 더해요. 행의 동작은 `keep` 그대로이고 아무것도 지우거나 옮기지 않아요. `moai audit-clone-plan` 의 답에는 `rules` 네 줄과 `evidence_cap_bytes` 가 들어갔어요. 아래 종료 코드는 모두 0 이에요.

- 새 시험 8개(`--test audit_copy_guard`)와 1개(`--test audit_clone_cli rules_`), 이웃 `optimize_plan_cli` 19개, `audit_clone_cli` 11개 통과. 경계 시험: 정확히 100 MiB 는 통과하고 1바이트 더는 걸려요.
- 묶음 71개 명령 전부 종료 0: `clippy --all-targets -D warnings`, `--lib` 1619, `--bins` 56, `audit_*`·`optimize_*`·`worktree_*` 시험 대상 전부, `tools/repo_scans.json` 검사 39개, 이 Mac 에서 도는 macOS 시험 파일 10개, 파이썬 `audit_eval_test`. 가벼운 검사(`cargo fmt --check`, `rustfmt --edition 2021 --check`, `git diff --check`, `moai trailer-check`, `moai spec check` 0/0)도 모두 통과.
- 변이 5개 모두 잡혔어요: `.venv` 를 표시에서 뺌, 상한을 이상(≥)으로 바꿈, `.git` 파일을 표시에서 뺌, 표시한 폴더 안으로 들어가 크기를 셈, 찾아도 이유를 더하지 않음. 변이 하나(`.git` 파일)와 하나(표시한 폴더의 크기를 세지 않음)는 처음 시험이 없어서 시험을 더했어요.

아직 안 한 것: 지침 글(감사 에이전트 지침, 감사·검증 워크플로, OPTIMIZE 워크플로)에 같은 규칙 문장 넣기. 그 파일들을 다른 변경이 쥐고 있어서 그 변경이 끝난 뒤에 따로 넣어요. 프로젝트 밖 홈 보고서 폴더(`~/.moai/reports/<project>/`) 스캔과, 증거 폴더에 파일을 넣을 때 막는 훅도 후속이에요.
