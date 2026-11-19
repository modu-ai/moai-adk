---
id: SPEC-20261119-enforce-merge-confirm
title: "A direct pull request merge from a Bash command asks for confirmation in a MoAI project, under an off, observe or enforce mode"
status: draft
revision: 2
created: 2026-11-19
---

# A direct pull request merge from a Bash command asks for confirmation in a MoAI project, under an off, observe or enforce mode

## Intent

카드 C-86 AC-4 의 마지막 절반은 "병합 관문 밖의 `gh pr merge` 는 hook 이 막는다"예요. 코드를 읽어 보니(card note 2292) AC-4 는 세 부분이고 상태가 달라요.

- (a) trailer 세션의 PASS 거부: CI 의 `moai audit-gate` 에 이미 있어요 — `src/audit_gate.rs:215-228` 가 작성자(`Moai-Session` trailer 또는 lane 이름)의 PASS 만 있으면 `missing`(rule `non-author-pass`)으로 만들고, `src/audit_gate.rs` 의 시험과 `tests/audit_gate_cli.rs` 가 고정해요. 이 SPEC 은 바꾸지 않아요.
- (b) 감사자 brief 기록: 코드에 없어요(`brief` 검색 0건). 규칙은 `skills/moai/workflows/mode-agent.md` 의 글뿐이에요.
- (c) 관문 밖 직접 병합: hook 이 없어요. 병합 관문 `moai gate-merge` 는 `gh api -X PUT …/pulls/<n>/merge` 를 자식 프로세스로 띄우고, hook 이 보는 것은 Bash 줄(`moai gate-merge …`)뿐이에요.

이 SPEC 은 (c)만 해요: MoAI 저장소에서 Bash 가 `gh pr merge` 나 `gh api` 로 PR 을 직접 병합하려 하면 확인을 요청하고(Claude), 확인을 못 하는 host 에서는 거절해요(기존 확인 규칙의 방식). 관문을 지나는 `moai gate-merge`·`moai gate-loop` 는 그대로 통과해요.

리더 결정(card note 2346, 2026-11-18): 이 카드의 hook 규칙은 **off / observe / enforce** 세 모드를 가져요. 기본은 observe — 거부할 것을 감사 로그에만 남기고 통과시키고, 오탐을 본 뒤 사용자 선택으로 enforce 로 올려요. enforce 에서 한 세션이 같은 규칙으로 두 번 막히면 다음은 통과시키고 이유를 남겨 갇히지 않게 해요. 이 SPEC 이 그 모드 층을 처음 만들고 `merge_confirm` 규칙에 적용해요(S1 의 `green_lock_bash` 는 #3216 이 끝난 뒤 후속 조각이에요).

## Non-goals

- (a)의 리더 병합 관문 쪽(`src/gate.rs` G9 는 `Moai-Claim` 작성자만 비교하고 `Moai-Session` 은 보지 않아요): 범위 밖(`src/gate.rs`, `src/gate_collect.rs`)이라 별도 카드로 제안만 해요. (b) 감사자 brief 기록도 판정문 형식과 `agents/audit.md` 를 바꾸므로 별도 카드로 제안만 해요(card note 2292 D-D, 리더 기본안 확인).
- 거부(deny)로 조이기: 문서에 적힌 리더의 직접 병합(`skills/moai/workflows/sync.md` 8단계, `AGENTS.local.md`)을 막으므로 enforce 도 확인 요청이에요(D-C = A). 거부로 바꾸는 것은 한 줄이에요.
- 다른 규칙에 모드 붙이기: 이 SPEC 은 모드 층과 `merge_confirm` 만 해요. S1 의 GREEN 잠금(`#3216`)과 이미 있던 규칙(`.git` 보호, SPEC 먼저 등)은 그대로예요.
- 셸이 병합을 하는 다른 길: 스크립트 안의 `gh`, `curl` 로 API 직접 호출, 변수로 만든 명령, 웹 화면의 병합 버튼은 읽지 않아요.
- Claude 가 `bypassPermissions` 나 `-p` 로 돌 때의 `ask` 동작은 문서에 없어요(기존 확인 규칙과 같은 한계).
- 실제 host 설정과 hook 설치의 변경: 시험은 임시 저장소와 임시 HOME 만 써요.

## Requirements

- REQ-1: While the mode of the rule `merge_confirm` is enforce and the session's work tree is a MoAI project (a `.moai` folder at its root), the preflight guard shall ask for confirmation of a Bash command that runs `gh pr merge`, sends a PUT to a `…/pulls/<n>/merge` endpoint through `gh api`, or runs the GraphQL mutation `mergePullRequest` or `enablePullRequestAutoMerge` through `gh api`, with a reason that names `moai gate-merge`.
- REQ-2: The preflight guard shall leave `moai gate-merge`, `moai gate-loop` and every `gh` command that only reads or comments on pull requests unchanged, including a GET of a `…/merge` endpoint and a PUT to an endpoint that does not merge.
- REQ-3: Where the session's work tree is not a MoAI project, the preflight guard shall allow the command as before and write nothing.
- REQ-4: Where the host cannot ask, the engine shall render this confirmation as a refusal by its existing confirmation rule.
- REQ-5: The `capabilities` answer of the hook engine shall list the direct pull request merge and the three modes among each host's confirmation facts.
- REQ-6: The mode of the rule `merge_confirm` shall be read from the key `merge_confirm` of `[hooks.modes]` in the HOME hook config, a missing file, table or key and any value other than `off`, `observe` and `enforce` meaning observe.
- REQ-7: While the mode is observe and the guard would have asked, the engine shall allow the command and append one `would_block` line, with the rule id and the guard's fixed sentence and no command text, path or content, to the engine log.
- REQ-8: While the mode is off, the engine shall allow the command without reading or writing the engine log.
- REQ-9: While the mode is enforce and the host cannot ask, the engine shall append one `rule_block` line with the rule id for each refusal.
- REQ-10 (REQ-9): IF the session already has two `rule_block` lines for the rule, THEN the engine shall allow the next such command and append one `rule_release` line with the rule id and the guard's fixed sentence.
- REQ-11 (REQ-9): IF a `rule_block` line cannot be written, THEN the engine shall allow the command, because a refusal that cannot be counted could not be bounded.

## Acceptance

- AC-1 (REQ-1, REQ-4): 모드를 enforce 로 둔 임시 HOME 과 MoAI 임시 저장소에서 `gh pr merge`(여러 모양)·`gh api -X PUT …/merge`·GraphQL 병합 변이는 Claude 에서 `ask`(reason 에 `moai gate-merge`), Codex 에서 `deny`, `echo` 안의 글과 `sh -c` 안의 명령도 같은 방식.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_asks`
- AC-2 (REQ-2, REQ-3): `moai gate-merge`·`moai gate-loop`·`gh pr view/list/comment`·`gh api` GET·병합이 아닌 PUT·`git merge`·`echo gh pr merge` 는 enforce 에서도 통과, `.moai` 없는 저장소에서는 전부 통과하고 로그에 아무것도 안 써요.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_leaves`; `cargo test --locked --offline --test hooks_cli merge_confirm_is_off`
- AC-3 (REQ-5): 세 host 의 `capabilities` 가 직접 병합 확인과 세 모드를 말해요.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_capabilities`
- AC-4 (REQ-6, REQ-7, REQ-8): 설정이 없거나 틀리면 observe(통과 + `would_block` 한 줄, `deny` 줄 아님), off 는 통과하고 로그가 안 생겨요, enforce 는 AC-1 과 같아요.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_mode`
- AC-5 (REQ-9, REQ-10, REQ-11): Codex 에서 같은 규칙이 두 번 막힌 세션의 세 번째는 통과 + `rule_release`, 다른 세션과 Claude 의 ask 는 세지 않고, 줄을 못 쓰는 저장소는 통과해요.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_strikes`
- AC-6 (REQ-1…REQ-11): 이웃 시험.
  verify: `cargo test --locked --offline --test hooks_cli`; `cargo test --locked --offline --test hook_dispatch_cli`; `cargo test --locked --offline --test hook_status_cli`; `cargo test --locked --offline --test spec_before_code_hook`; `cargo test --locked --offline --test green_test_lock`; `cargo test --locked --offline --lib hook_`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/hook_guards.rs`, `src/hook_engine.rs`, `src/hook_config.rs`, `tests/hooks_cli.rs`, `.moai/specs/SPEC-20261119-enforce-merge-confirm/spec.md`
- read-only: `src/hooks.rs`, `src/gate_merge.rs`, `src/audit_gate.rs`, `src/claim_verify.rs`, `tests/hook_dispatch_cli.rs`, `skills/moai/workflows/mode-leader.md`, `skills/moai/workflows/sync.md`
- forbidden: the real host settings and MoAI state under the user's home (tests use a temporary HOME), `src/gate.rs`, `src/gate_collect.rs`, `agents/audit.md`, any change that lets a hook allow what it asked about or refused before outside the new modes

## Plan

1. 시험 먼저(`tests/hooks_cli.rs`): 임시 저장소(`.git` 폴더를 손으로 만들고 `.moai` 폴더 유무만 바꿔요)와 임시 HOME 의 `hook-config.toml` 로 모드를 바꿔 가며 `moai hook-dispatch --host claude|codex --event tool-preflight` 에 Bash 입력을 먹이고, 엔진 로그를 읽어요. 모드를 읽는 코드가 없으니 enforce·observe 시험이 RED 로 실패해요.
2. `hook_guards::merge_confirm(command, cwd)`: `shell_commands` 로 `gh pr merge`·`gh api`(PUT `…/merge`, GraphQL 변이 이름)를 읽고 `.moai` 폴더 확인 뒤 문장을 돌려줘요(만든 뒤 시험으로 고정됨).
3. `hook_config`: `[hooks.modes]` 를 읽어 `HookConfig::modes`(규칙 id → 모드)로 두고, 없거나 틀리면 observe.
4. `hook_engine`: 확인 단계에서 `bash_confirm` 다음에 부르고, 모드에 따라 off 통과 / observe 통과 + `would_block` / enforce ask·거절 + `rule_block`·`rule_release`(`consecutive_blocks_in` 처럼 로그를 세요). `CONFIRM_COMMANDS` 와 `capabilities` 문구에 더해요.
5. 이웃 시험과 mutation(가드마다 한 번, 12개 이하, 위험한 것부터).

## Risks

- 문서화된 리더의 직접 병합(`gh pr merge --auto --merge --match-head-commit`)은 enforce 에서 확인을 물어요: 기본이 observe 라 켜기 전에는 아무것도 막히지 않고, 켠 뒤에도 거부가 아니라 확인이에요.
- 기본이 observe 라서 설정을 안 한 설치에서는 이 규칙이 아무것도 막지 않아요: 의도한 것이고(리더 결정 note 2346), 오탐을 로그로 본 뒤 사용자가 enforce 를 골라요. 로그를 읽는 화면은 이 SPEC 에 없어요.
- 두 번 규칙은 deny 로 렌더되는 host(Codex, OMP)에만 세요. Claude 의 ask 는 사람이 매번 답하니 갇힘이 없고, 세면 한 세션의 세 번째 병합부터 확인 없이 통과해 규칙이 약해져요(D-4).
- `gh api` 의 옵션 해석은 근사예요(`-X PUT`, `-XPUT`, `--method=PUT`, 대소문자). 모르는 모양은 통과해요(Non-goals 에 적은 한계).
- S1 의 PR(#3216)과 같은 파일(`src/hook_engine.rs`, `tests/hooks_cli.rs`)을 건드려요: 시험은 다른 위치에 더했고, 엔진 함수 모양이 겹치면 늦게 병합되는 쪽이 한 번 손으로 합쳐요.
- 로그에 쓰는 reason 은 규칙의 고정 문장뿐이고 명령·경로·내용은 쓰지 않아요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | enforce 도 거부가 아니라 확인(ask)이에요 | 리더의 직접 병합이 문서에 적혀 있어 거부는 병합 흐름을 멈춰요(card note 2292 D-C = A, 리더 확인) | 거부, 거부 + 환경 표지(선택 B·C) |
| D-2 | MoAI 프로젝트 판정은 작업 트리 루트의 `.moai` 폴더로 해요 | A5 규칙과 같은 판정이고 다른 저장소에서는 건드리지 않아요(AC-5 의 허용 경로) | 설정 파일의 켜기 스위치 |
| D-3 | (a)는 바꾸지 않고, 리더 관문 G9 와 brief 기록은 제안만 해요 | 범위 밖 파일이고 판정문 형식이 바뀌어요(note 2292 D-D, 리더 확인) | 이 SPEC 에서 함께 |
| D-4 | 모드: 설정은 HOME hook-config.toml 의 `[hooks.modes]`, 기본 observe, 틀린 값도 observe, 두 번 규칙은 deny 로 렌더되는 host 에만 세요 | 리더 결정 note 2346 + 기존 stop_block 세기와 설정 읽기를 따라가요; 사람이 답하는 ask 는 갇힘이 없어요 | 모든 확인을 세기(규칙이 약해짐), 환경 변수로 모드 정하기(세션마다 달라져 오탐 관찰이 흩어짐) |

## Result

(채워지지 않았어요: 구현 뒤에 AC 마다 한 줄을 적어요.)
