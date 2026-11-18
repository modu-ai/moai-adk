---
id: SPEC-20261119-enforce-merge-confirm
title: "A direct pull request merge from a Bash command asks for confirmation in a MoAI project"
status: draft
revision: 1
created: 2026-11-19
---

# A direct pull request merge from a Bash command asks for confirmation in a MoAI project

## Intent

카드 C-86 AC-4 의 마지막 절반은 "병합 관문 밖의 `gh pr merge` 는 hook 이 막는다"예요. 코드를 읽어 보니(card note 2292) AC-4 는 세 부분이고 상태가 달라요.

- (a) trailer 세션의 PASS 거부: CI 의 `moai audit-gate` 에 이미 있어요 — `src/audit_gate.rs:215-228` 가 작성자(`Moai-Session` trailer 또는 lane 이름)의 PASS 만 있으면 `missing`(rule `non-author-pass`)으로 만들고, `src/audit_gate.rs` 의 시험과 `tests/audit_gate_cli.rs` 가 고정해요. 이 SPEC 은 바꾸지 않아요.
- (b) 감사자 brief 기록: 코드에 없어요(`brief` 검색 0건). 규칙은 `skills/moai/workflows/mode-agent.md` 의 글뿐이에요.
- (c) 관문 밖 직접 병합: hook 이 없어요. 병합 관문 `moai gate-merge` 는 `gh api -X PUT …/pulls/<n>/merge` 를 자식 프로세스로 띄우고, hook 이 보는 것은 Bash 줄(`moai gate-merge …`)뿐이에요.

이 SPEC 은 (c)만 해요: MoAI 저장소에서 Bash 가 `gh pr merge` 나 `gh api` 로 PR 을 직접 병합하려 하면 확인을 요청하고(Claude), 확인을 못 하는 host 에서는 거절해요(기존 확인 규칙의 방식). 관문을 지나는 `moai gate-merge`·`moai gate-loop` 는 그대로 통과해요.

## Non-goals

- (a)의 리더 병합 관문 쪽(`src/gate.rs` G9 는 `Moai-Claim` 작성자만 비교하고 `Moai-Session` 은 보지 않아요): 범위 밖(`src/gate.rs`, `src/gate_collect.rs`)이라 별도 카드로 제안만 해요. (b) 감사자 brief 기록도 판정문 형식과 `agents/audit.md` 를 바꾸므로 별도 카드로 제안만 해요(둘 다 card note 2292 D-D, 리더 기본안 확인).
- 거부(deny)로 조이기: 문서에 적힌 리더의 직접 병합(`skills/moai/workflows/sync.md` 8단계, `AGENTS.local.md`)을 막으므로 이번엔 확인 요청이에요(D-C = A). 거부로 바꾸는 것은 한 줄이에요.
- 셸이 병합을 하는 다른 길: 스크립트 안의 `gh`, `curl` 로 API 직접 호출, 변수로 만든 명령, 웹 화면의 병합 버튼은 읽지 않아요.
- Claude 가 `bypassPermissions` 나 `-p` 로 돌 때의 `ask` 동작은 문서에 없어요(기존 확인 규칙과 같은 한계).
- 실제 host 설정과 hook 설치의 변경: 시험은 임시 저장소와 임시 HOME 만 써요.

## Requirements

- REQ-1: While the session's work tree is a MoAI project (a `.moai` folder at its root), the preflight guard shall ask for confirmation of a Bash command that runs `gh pr merge`, sends a PUT to a `…/pulls/<n>/merge` endpoint through `gh api`, or runs the GraphQL mutation `mergePullRequest` or `enablePullRequestAutoMerge` through `gh api`, with a reason that names `moai gate-merge`.
- REQ-2: The preflight guard shall leave `moai gate-merge`, `moai gate-loop` and every `gh` command that only reads or comments on pull requests unchanged, including a GET of a `…/merge` endpoint.
- REQ-3: Where the session's work tree is not a MoAI project, the preflight guard shall allow the command as before.
- REQ-4: Where the host cannot ask, the engine shall render this confirmation as a refusal by its existing confirmation rule.
- REQ-5: The `capabilities` answer of the hook engine shall list the direct pull request merge among each host's confirmation commands.

## Acceptance

- AC-1 (REQ-1, REQ-4): MoAI 임시 저장소에서 `gh pr merge`(여러 모양)·`gh api -X PUT …/merge`·GraphQL 병합 변이는 Claude 에서 `ask`(reason 에 `moai gate-merge`), Codex 에서 `deny`, `echo` 안의 글과 `sh -c` 안의 명령도 같은 방식.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_asks`
- AC-2 (REQ-2, REQ-3): `moai gate-merge`·`moai gate-loop`·`gh pr view/list/comment`·`gh api` GET·`git merge`·`echo gh pr merge` 는 통과, `.moai` 없는 저장소에서는 전부 통과.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_leaves`; `cargo test --locked --offline --test hooks_cli merge_confirm_is_off`
- AC-3 (REQ-5): 세 host 의 `capabilities` 가 직접 병합 확인을 말해요.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_capabilities`
- AC-4 (REQ-1…REQ-5): 이웃 시험.
  verify: `cargo test --locked --offline --test hooks_cli`; `cargo test --locked --offline --test hook_dispatch_cli`; `cargo test --locked --offline --test spec_before_code_hook`; `cargo test --locked --offline --test green_test_lock`; `cargo test --locked --offline --lib hook_`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/hook_guards.rs`, `src/hook_engine.rs`, `tests/hooks_cli.rs`, `.moai/specs/SPEC-20261119-enforce-merge-confirm/spec.md`
- read-only: `src/gate_merge.rs`, `src/audit_gate.rs`, `src/claim_verify.rs`, `tests/hook_dispatch_cli.rs`, `skills/moai/workflows/mode-leader.md`, `skills/moai/workflows/sync.md`
- forbidden: the real host settings and MoAI state under the user's home (tests use a temporary HOME), `src/gate.rs`, `src/gate_collect.rs`, `agents/audit.md`, any change that lets a hook allow what it asked about or refused before

## Plan

1. 시험 먼저(`tests/hooks_cli.rs`): 임시 저장소(`.git` 폴더를 손으로 만들고 `.moai` 폴더 유무만 바꿔요)에서 `moai hook-dispatch --host claude|codex --event tool-preflight` 로 Bash 입력을 먹여 `ask`·`deny`·통과를 읽어요. 함수가 없으니 ask 시험이 RED 로 실패해요.
2. `hook_guards::merge_confirm(command, cwd)`: `shell_commands` 로 `gh pr merge`·`gh api`(PUT `…/merge`, GraphQL 변이 이름)를 읽고 `.moai` 폴더 확인 뒤 문장을 돌려줘요.
3. `hook_engine` 의 확인 단계에서 `bash_confirm` 다음에 부르고, `CONFIRM_COMMANDS` 문구에 더해요.
4. 이웃 시험과 mutation(가드마다 한 번, 12개 이하).

## Risks

- 문서화된 리더의 직접 병합(`gh pr merge --auto --merge --match-head-commit`)이 이제 확인을 물어요: 리더 세션에서 한 번 승인하면 되고, 거부가 아니라 확인이라 흐름은 멈추지 않아요. 확인을 못 하는 Codex 세션은 거절돼요(그 세션은 병합하면 안 돼요: lane 은 병합하지 않아요).
- `gh api` 의 옵션 해석은 근사예요(`-X PUT`, `-XPUT`, `--method=PUT`, 대소문자). 모르는 모양은 통과해요(위 Non-goals 에 적은 한계).
- S1 의 PR 과 같은 파일(`src/hook_engine.rs`, `tests/hooks_cli.rs`)을 건드려요: 다른 위치에 더해 병합 충돌을 줄였어요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 거부가 아니라 확인(ask)으로 시작해요 | 리더의 직접 병합이 문서에 적혀 있어 거부는 병합 흐름을 멈춰요(card note 2292 D-C = A, 리더 확인) | 거부, 거부 + 환경 표지(선택 B·C) |
| D-2 | MoAI 프로젝트 판정은 작업 트리 루트의 `.moai` 폴더로 해요 | A5 규칙과 같은 판정이고 다른 저장소에서는 건드리지 않아요(AC-5 의 허용 경로) | 설정 파일의 켜기 스위치 |
| D-3 | (a)는 바꾸지 않고, 리더 관문 G9 와 brief 기록은 제안만 해요 | 범위 밖 파일이고 판정문 형식이 바뀌어요(note 2292 D-D, 리더 확인) | 이 SPEC 에서 함께 |

## Result

(채워지지 않았어요: 구현 뒤에 AC 마다 한 줄을 적어요.)
