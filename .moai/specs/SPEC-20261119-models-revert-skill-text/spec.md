---
id: SPEC-20261119-models-revert-skill-text
title: "update.md names the models-revert command and its permission rule (C-47 slice 7)"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-launcher-settings-receipt
---

# update.md names the models-revert command and its permission rule (C-47 slice 7)

## Intent

C-47 조각 5(SPEC-20261119-models-revert-command)가 되돌리기 명령 `models-revert <HOME> --receipt models-… --permission-reference <ref>` 를 만들었어요. 그런데 이 스킬이 세션에게 읽어 주는 문서 `skills/moai/workflows/update.md` 의 7번 줄은 아직 "되돌리는 라이브러리 함수는 아직 명령이 아니니 사람에게 되돌리기 명령을 약속하지 말라" 고 말해요. 문서가 코드보다 낡았고(core-rules 21: 문서와 코드가 다르면 코드가 이겨요), 사람이 "방금 저장한 모델 설정을 되돌려 줘" 라고 하면 세션이 명령이 있는데도 없다고 답해요.

이 조각은 그 한 문장만 고쳐요: 명령의 이름과 인자, 사람이 되돌리기를 요청했을 때만 사람이 준 권한 참조로 실행한다는 규칙, 두 가지 결과(`reverted`, `skipped` 와 `changed_since_apply`)와 거절의 모양. 이 일은 manifest 가 풀리기를 기다렸어요(리더 결정 노트 2400, 2742 점(5)): C-103 과 C-59 가 manifest 를 놓았고 C-114 도 놓아서 이제 받을 수 있어요.

이 조각은 조각 6 의 head(a54c62a237, 가지 spec/20261119-models-revert-reporting) 위에 서요. 명령과 그 운영 문서(`docs/design-review/appendix/operations/models_revert.md`)가 그 가지에만 있으므로, 조각 5 와 6 이 먼저 묶음에 들어가야 해요.

## Non-goals

- 명령 코드(`src/ops/models_revert.rs`, `src/update_models.rs`), 도움말, 운영 문서 바꾸기 — 조각 5 와 6 이 끝냈어요.
- 카드 기준 5 와 7(런처가 저장된 값을 읽는 일), 조각 5 재감사 N2(영수증 폴더가 링크일 때의 쓰기), 노트 2688 의 (1)(상태 폴더가 링크일 때: 사용자 결정 대기).
- `moai update` 터미널 화면에 명령 힌트 줄 더하기.
- `update.md` 의 다른 줄이나 다른 스킬 문서.

## Requirements

- REQ-1: Where `skills/moai/workflows/update.md` describes the receipt of a saved models choice, it shall name the command `moai models-revert` with its `--receipt` and `--permission-reference` arguments as the way to put the earlier file back.
- REQ-2: Where `skills/moai/workflows/update.md` describes the receipt of a saved models choice, it shall not say that the undo is not a command yet.
- REQ-3: Where `skills/moai/workflows/update.md` names that command, it shall say that the session runs it only when the person asks for the undo and only with a permission reference the person gave, and that the `receipt` id printed by the save is the `--receipt` argument.
- REQ-4: Where `skills/moai/workflows/update.md` names that command, it shall name its two answers, `reverted` and `skipped` with the reason `changed_since_apply`, and say that a refusal exits 1 with an error envelope and changes nothing.
- REQ-5: When a file under `skills/moai/` changes, `skills/moai/references/source-manifest.json` shall hold the sha256 that `tools/source_manifest.py write` writes for that file. · failure: none — `tools/source_manifest.py check` is the check that fails when a hash is stale.
- REQ-6: When `tests/models_revert_cli.rs` runs, it shall read the real `skills/moai/workflows/update.md` and pin that it names what REQ-1 to REQ-4 require, so a later edit that drops a name turns a test red. · failure: none — a test that does not find the name fails, there is no softer outcome.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): `update.md` names `moai models-revert`, both arguments, the person-asks-and-gives-the-reference rule and the two answers, and the words "not a command yet" are gone.
  verify: `rg -n -e "moai models-revert" -e "changed_since_apply" -e "not a command yet" skills/moai/workflows/update.md`
  verify: `cargo test --locked --offline --test models_revert_cli update_md_names_the_models_revert_command`
- AC-2 (REQ-5): the manifest is current and the package fits.
  verify: `python3 -I tools/source_manifest.py check`
  verify: `cargo run --offline --locked --quiet --bin moai -- package-inspect skills/moai`
- AC-3 (REQ-6): the document pin passes and the tests that read this file still pass.
  verify: `cargo test --locked --offline --test models_revert_cli`
  verify: `cargo test --locked --offline --test worktree_include_guard`
  verify: `cargo test --locked --offline --test agent_defs_package`
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.skill_references_test tests.source_manifest_test`

## Scope

- write: `.moai/specs/SPEC-20261119-models-revert-skill-text/spec.md`, `skills/moai/workflows/update.md`, `skills/moai/references/source-manifest.json`, `tests/models_revert_cli.rs`
- read-only: `src/ops/models_revert.rs`, `src/update_models.rs`, `src/help.rs`, `docs/design-review/appendix/operations/models_revert.md`, `tests/worktree_include_guard.rs`, `tests/agent_defs_package.rs`
- forbidden: 위 write 밖의 `skills/` 파일, 실제 HOME 과 원장, 네트워크, 키

## Plan

1. SPEC 을 쓰고 `spec check` 오류 0 을 확인한다.
2. 첫 RED: `tests/models_revert_cli.rs` 에 `update.md` 가 이름을 적는지 읽는 시험을 먼저 쓰고, 지금 글로는 실패하는 것을 본다.
3. `update.md` 7번 줄의 한 문장을 고친다. `wc -c` 로 65,536 바이트 한도와 여유를 확인한다.
4. `python3 -I tools/source_manifest.py write` 뒤 `check`, 빌드한 바이너리로 `package-inspect`.
5. 점검: 위 AC 의 명령, 이름을 하나씩 지우는 변이(시험이 빨개지는지), 팁과 조각 6 head 에 대한 `git merge-tree` 와 병합 결과의 manifest 검사, tier 1 독립 감사.

## Risks

- `skills/moai/references/source-manifest.json` 은 다른 카드도 쓰는 파일이에요: 이 조각은 마지막 커밋에서만 manifest 를 쓰고, 풀어 준 뒤에도 병합 결과에서 `check` 를 돌려요(항목별 줄이라 다른 카드의 항목과 안 겹쳐요).
- 문장이 코드보다 먼저 낡을 수 있어요: 명령 이름과 두 결과 이름은 `tests/models_revert_cli.rs` 가 같은 파일에서 이미 실제 이진으로 확인하는 값과 같은 이름만 읽어 맞춰 봐요.
- `worktree_include_guard`, `project_decide_model` 같은 시험이 `update.md` 를 읽어요: 줄 하나만 바꾸고 그 시험들을 돌려요.

## Decisions

| id | decision | why | reversible |
| --- | --- | --- | --- |
| D-1 | `update.md` 7번 줄의 한 문장만 고치고 줄을 나누지 않는다 | 이 스킬의 규칙들이 한 줄에 한 규칙이고, 시험도 그 줄을 이름으로 읽어요 | 문장 이동 |
| D-2 | 시험은 이름과 짧은 꼬리만 읽는다(문장 전체 못박기 없음) | 문장을 다듬을 때마다 시험이 깨지는 것을 막으려고 이름 목록만 읽어요(C-114 조각 3 의 D-3 와 같은 방침) | 시험 추가 |
| D-3 | manifest 해시는 이 가지의 마지막 커밋에서만 쓴다 | manifest 는 여러 카드가 기다리는 파일이라 쥐는 시간을 줄여요(리더 결정, C-59 노트 3081) | 마지막 단계에서 `write` |
