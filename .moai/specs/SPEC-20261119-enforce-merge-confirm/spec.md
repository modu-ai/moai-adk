---
id: SPEC-20261119-enforce-merge-confirm
title: "A direct pull request merge from a Bash command asks for confirmation in a MoAI project, under an off, observe or enforce mode"
status: draft
revision: 3
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
- 셸 문법은 이 SPEC 이 따로 읽지 않고 GREEN 잠금(SPEC-20261119-enforce-spec-hooks REQ-11)과 같은 닫힌 규칙 `open_shell_structure` 로 읽어요: 단어 경계(공백 `;` `&` `|` `(` `)` 따옴표), 구조 단어(`if then elif else fi while until do done ! { }`)는 명령 구분자, 백틱은 `$( )`, `$'…'`·`$"…"` 는 일반 따옴표. 그래서 `;` `&&` `||` `|` 로 이은 줄, `sh -c`·`eval`, `( )`·`$( )`, `if … then gh pr merge 5; fi`, `! gh …`, `{ gh …; }` 를 읽어요(시험 행이 RED 로 먼저 보여 줬어요: 키워드 뒤의 `gh pr merge` 는 처음 구현에서 통과했어요). 이 규칙 밖의 문법은 아래 "다른 길"에 이름이 있어요; 구조 단어가 따옴표 없는 인자로 서도 구조 단어로 읽히는 거짓 확인 요청(`echo do gh pr merge 5`)은 규칙이 막은 쪽으로 틀려요.
- `gh` 의 다른 병합 모양: 옵션이 아닌 단어 기준으로 `pr merge`·`api` 는 첫 단어나(`gh pr merge`) 옵션 값 하나 뒤의 둘째 단어(`gh -R owner/repo pr merge`)에서만 읽어요. `gh api` 는 `-X PUT`·`-XPUT`·`--method PUT`·`--method=PUT` 와 쿼리 문자열이 붙은 주소(`…/merge?x=1`), 인자 어디에든 `mergePullRequest`·`enablePullRequestAutoMerge` 가 든 GraphQL 변이를 읽어요. 이 밖의 옵션 배치(옵션 값이 둘 이상 앞서는 줄 등)는 통과해요.
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
- REQ-10 (REQ-9): IF the call names a session and that session already has two `rule_block` lines for the rule, THEN the engine shall allow the next such command and append one `rule_release` line with the rule id and the guard's fixed sentence.
- REQ-11 (REQ-9): IF a `rule_block` line cannot be written, THEN the engine shall allow the command, because a refusal that cannot be counted could not be bounded.
- REQ-12 (REQ-9): IF the call names no session, THEN the engine shall refuse the command every time and append one `rule_block` line for each refusal, because the lines of every nameless call would otherwise be counted together across all sessions of the project.

## Acceptance

- AC-1 (REQ-1, REQ-4): 모드를 enforce 로 둔 임시 HOME 과 MoAI 임시 저장소에서 `gh pr merge`(여러 모양)·`gh api -X PUT …/merge`·GraphQL 병합 변이는 Claude 에서 `ask`(reason 에 `moai gate-merge`), Codex 에서 `deny`, `echo` 안의 글과 `sh -c` 안의 명령도 같은 방식.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_asks`
- AC-2 (REQ-2, REQ-3): `moai gate-merge`·`moai gate-loop`·`gh pr view/list/comment`·`gh api` GET·병합이 아닌 PUT·`git merge`·`echo gh pr merge` 는 enforce 에서도 통과, `.moai` 없는 저장소에서는 전부 통과하고 로그에 아무것도 안 써요.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_leaves`; `cargo test --locked --offline --test hooks_cli merge_confirm_is_off`
- AC-3 (REQ-5): 세 host 의 `capabilities` 가 직접 병합 확인과 세 모드를 말해요.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_capabilities`
- AC-4 (REQ-6, REQ-7, REQ-8): 설정이 없거나 틀리면 observe(통과 + `would_block` 한 줄, `deny` 줄 아님), off 는 통과하고 로그가 안 생겨요, enforce 는 AC-1 과 같아요.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_mode`
- AC-5 (REQ-9, REQ-10, REQ-11, REQ-12): Codex 에서 같은 규칙이 두 번 막힌 세션의 세 번째는 통과 + `rule_release`, 다른 세션과 Claude 의 ask·observe 로 통과한 호출은 세지 않고, 줄을 못 쓰는 저장소는 통과해요. 세션 이름이 없거나 빈 호출(OMP·Codex)은 몇 번이든 매번 거절하고 줄은 남겨요.
  verify: `cargo test --locked --offline --test hooks_cli merge_confirm_strikes`; `cargo test --locked --offline --test hooks_cli merge_confirm_counts`; `cargo test --locked --offline --test hooks_cli merge_confirm_without_a_session`; `cargo test --locked --offline --test hooks_cli merge_confirm_observed`
- AC-6 (REQ-1…REQ-12): 이웃 시험.
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
- S1 의 PR(#3216)과 같은 파일(`src/hook_engine.rs`, `src/hook_guards.rs`, `tests/hooks_cli.rs`)을 건드려요: #3216 이 먼저 병합되고, 이 SPEC 의 커밋은 그 tip 위에서 `tool_preflight` 인자 충돌을 한 번 손으로 풀어 다시 만들었어요. 셸 문법은 #3216 의 `open_shell_structure` 규칙에 기대므로 그 규칙이 바뀌면 `merge_confirm` 시험 표(`DIRECT_MERGES`)도 같이 봐야 해요.
- 로그에 쓰는 reason 은 규칙의 고정 문장뿐이고 명령·경로·내용은 쓰지 않아요.
- 세션 이름이 없는 호출(OMP 는 `session_id` 를 보내지 않고 Codex 는 빈 값을 보낼 수 있어요)은 풀어 주지 않아요(REQ-12): enforce 로 켜면 그 host 에서 직접 병합은 매번 거절돼요. 거절은 사람이 `moai gate-merge` 로 돌리면 풀리니 갇히지는 않지만 불편해요. 대안은 부모 프로세스 번호로 세션을 근사하는 것인데 유닉스에서만 되고 같은 프로세스 안의 다른 세션을 못 가려요 — 리더 선택으로 남겨 두어요(D-5).
- enforce 가 조용히 거절을 못 세는 경우: hook 의 작업 폴더가 저장소 하위 폴더일 때, 엔진 로그가 크기 한도(4 MiB)에 닿았을 때, `.moai-cli` 폴더가 없는 프로젝트. 이때 `rule_block` 줄을 못 써서 REQ-11 로 통과해요(거절이 안 걸림). `capabilities` 의 확인 문구에 이 세 경우를 적고 시험으로 고정했어요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | enforce 도 거부가 아니라 확인(ask)이에요 | 리더의 직접 병합이 문서에 적혀 있어 거부는 병합 흐름을 멈춰요(card note 2292 D-C = A, 리더 확인) | 거부, 거부 + 환경 표지(선택 B·C) |
| D-2 | MoAI 프로젝트 판정은 작업 트리 루트의 `.moai` 폴더로 해요 | A5 규칙과 같은 판정이고 다른 저장소에서는 건드리지 않아요(AC-5 의 허용 경로) | 설정 파일의 켜기 스위치 |
| D-3 | (a)는 바꾸지 않고, 리더 관문 G9 와 brief 기록은 제안만 해요 | 범위 밖 파일이고 판정문 형식이 바뀌어요(note 2292 D-D, 리더 확인) | 이 SPEC 에서 함께 |
| D-4 | 모드: 설정은 HOME hook-config.toml 의 `[hooks.modes]`, 기본 observe, 틀린 값도 observe, 두 번 규칙은 deny 로 렌더되는 host 에만 세요 | 리더 결정 note 2346 + 기존 stop_block 세기와 설정 읽기를 따라가요; 사람이 답하는 ask 는 갇힘이 없어요 | 모든 확인을 세기(규칙이 약해짐), 환경 변수로 모드 정하기(세션마다 달라져 오탐 관찰이 흩어짐) |
| D-5 | 세션 이름이 없는 호출은 풀지 않고 매번 거절(`rule_block` 은 계속 남겨요) | 이름 없는 세션의 줄을 합쳐 세면 두 번 거절 뒤 프로젝트 안의 모든 이후 세션이 직접 병합을 통과해요(감사 B1) | 부모 프로세스 번호로 세션 근사(유닉스 전용), 이름 없는 호출은 규칙에서 빼기(규칙이 켜진 채로 구멍) |

## Result

구현 에이전트가 채웠어요. 감사가 다시 확인해요. macOS arm64, 이 worktree(`spec/20261119-enforce-merge-confirm`), 빌드 폴더 `~/.moai/build/moai-adk-86-s4`, `CARGO_INCREMENTAL=0`, cargo cache `/tmp/moai-rc-cargo-agent-3`. 시각은 재지 않았어요. 이 절은 재감사 FAIL(B3: 커밋에 `Moai-Spec` 없음, #3216 과 `src/hook_engine.rs` 충돌)을 수리해 #3216 head(871fef0b0e)를 합친 tip 위에 S4 커밋을 다시 만든 뒤의 값이에요(가지 `spec/20261119-enforce-merge-confirm-r2`; 이전 가지의 코드 수리 B1·B2 는 재감사가 인정했어요; 시험 개수는 이 트리에서 파일마다 `cargo test --test <이름>` 로그의 `test result` 줄에서 읽었어요). `tool_preflight` 는 #3216 의 `fixed_reason` 인자와 S4 의 `host`·`now_unix_ms` 인자를 둘 다 가져요. 셸 문법은 `merge_confirm` 이 #3216 의 닫힌 규칙 `open_shell_structure` 를 그대로 써요(키워드 벗기기 코드는 지웠어요).

RED 먼저: 첫 구현(커밋 3f1e28a116)은 새 시험 4개가 모두 실패한 뒤 통과했고, 모드 층(커밋 8a8d77917a)은 시험을 먼저 고쳐 쓴 뒤 구현했어요(RED 실행은 기록하지 않았어요). 수리의 시험(`merge_confirm_without_a_session_id_is_never_released`, `merge_confirm_observed_calls_are_not_strikes`, `merge_confirm_counts_omp_refusals_per_session`)은 구현과 함께 썼고 RED 를 따로 보지 못했어요; 대신 변이 T1·T3(아래)로 시험이 그 결함을 잡는지 확인했어요. 복합문 키워드 수리는 시험 행(`if true; then gh pr merge 5; fi` 등)을 먼저 더해 `merge_confirm_asks…` 가 `left: (0, "") / right: (0, "ask")` 로 실패하는 것을 본 뒤 구현했어요(이전 가지의 일; 다시 만든 가지는 `open_shell_structure` 재사용으로 같은 행이 통과하고, 변이 T12 가 그 재사용을 지우면 실패하는 것으로 확인했어요).

- AC-1: met — `cargo test --locked --offline --test hooks_cli merge_confirm_asks` 통과(`gh pr merge` 여러 모양·`gh api -X PUT …/merge`·GraphQL 변이를 Claude 는 `ask`, Codex 는 `deny` + "Codex cannot ask"; 명령마다 세션을 따로 줘서 세 번째 풀림과 섞이지 않게 했어요; 복합문 행도 같은 표에 있어요: `if`·`else`·`for`·`while`·`until` 몸통, `! gh …`, `{ gh …; }`, 키워드 둘이 겹친 `then ! gh …`·`do { gh …; }`; 이름으로 남긴 거짓 확인 요청 `echo do gh pr merge 5` 도 ask 로 고정). `merge_confirm_reads_the_subcommand_not_the_words_after_it` 통과(`gh -R o/r pr merge`, `gh --hostname h api -X PUT …/merge`, 쿼리 문자열 붙은 주소는 ask; 댓글 글에 든 `pr merge` 는 통과).
- AC-2: met — `merge_confirm_leaves` 와 `merge_confirm_is_off` 통과(`moai gate-merge`·`gate-loop`·`gh pr view/list/comment/checks`·`gh api` GET·병합이 아닌 PUT·`git merge`·`echo gh pr merge`·grep 은 enforce 에서도 통과, `.moai` 폴더가 없는 저장소는 통과하고 로그에 아무것도 안 써요).
- AC-3: met — `merge_confirm_capabilities` 통과(claude·codex·omp 세 host 가 `gh pr merge`·`moai gate-merge`·`[hooks.modes]`·off/observe/enforce·`merge_confirm` 을 말해요).
- AC-4: met — `merge_confirm_mode` 통과(설정 없음·빈 파일·잘못된 단어·문자열 아닌 값·다른 규칙·깨진 TOML 은 observe: 통과 + `would_block` 한 줄, `deny` 줄 없음, 줄에 명령·PR 번호·경로 없음; off 는 통과하고 로그가 안 생겨요) + `--lib hook_config` 11 통과(`rule_modes_default_to_observe`).
- AC-5: met — `merge_confirm_strikes_release_the_third_refusal`, `merge_confirm_counts_only_its_own_blocks`, `merge_confirm_without_a_session_id_is_never_released`, `merge_confirm_observed_calls_are_not_strikes`, `merge_confirm_counts_omp_refusals_per_session` 통과(Codex 에서 같은 세션의 세 번째 거절이 통과 + `rule_release`(이유 포함), 다른 세션은 따로 세고, Claude 의 ask 와 observe 로 통과한 `would_block` 호출은 세지 않고, 로그를 못 쓰는 저장소는 통과; 다른 규칙의 `rule_block` 과 이 규칙의 `rule_release` 를 심어도 횟수가 안 변해요; OMP·Codex 에서 `session_id` 가 없거나 빈 호출은 다섯 번 불러도 매번 거절되고 `rule_block` 줄이 다섯 줄 남고 `rule_release` 는 없어요(REQ-12, 감사 B1); OMP 의 두 세션은 따로 세요).
- AC-6: met — `--test hooks_cli` 48(#3216 의 GREEN 잠금 시험 37 + 이 SPEC 의 11), `spec_before_code_hook` 9, `green_test_lock` 14, `hook_dispatch_cli` 56, `hook_status_cli` 10, `ac_command_lint` 5, `git_env_scrub` 3, `--lib hook_` 44 모두 통과, `cargo fmt --check` 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 0, `python3 -m unittest tests.ci_scope_test` OK, `moai spec check SPEC-20261119-enforce-merge-confirm` 오류 0·경고 0. 이웃 실행 중 observe 시험이 한 번 실패했는데 시험의 결함이었어요(로그의 시간값 `…1831235` 에 우연히 `123` 이 들어 있었어요): PR 번호를 `987654` 로 바꾸고 문자열 값만 검사하게 고쳐 `merge_confirm` 3회 반복과 전체 29개가 통과했어요. 돌리지 않은 것: 전체 `cargo test`(CI 가 돌려요).
- mutation: 첫 구현의 12개(가지 제거·일치 조건 완화 등)는 모두 KILLED(`target/c86s4/mutants.txt`, 모드 층 전의 코드). 모드 층은 12개(규칙 없음을 enforce 로, off 가 ask, observe 가 줄 안 씀, Claude ask 를 셈, 상한 하나 차이, 세션 공유, 규칙 공유, 쓸 수 없어도 거절, release 가 줄 안 씀, release 줄을 셈, 명령을 call 칸에 넣기, 모르는 단어를 enforce 로)를 돌려 10개가 처음부터 KILLED 였고, 규칙 공유와 release 계수 둘은 SURVIVED(규칙이 하나뿐이라 시험이 구별 못 했어요) → 시험 `merge_confirm_counts_only_its_own_blocks` 를 더한 뒤 둘 다 KILLED(`target/c86s4/mutants3.txt`, `mutants3b.txt`, `mutants3c.txt`), `git status` 깨끗이 복구.
- mutation(다시 만든 가지, 커밋한 코드 위, `target/c86s4/mutants11.txt`): 12개 모두 KILLED — T1 `would_block` 을 거절로 셈[감사 M4], T2 호스트 게이트 `!= Codex`[감사 M12], T3 이름 없는 세션도 풀림, T4 쿼리 문자열을 안 자름, T5·T6·T7·T8 `gh` 구독 명령 위치·단어 창, T9 상한 하나 차이, T10·T11 `capabilities` 한계 문구 빼기, T12 `open_shell_structure` 를 거치지 않고 줄을 그대로 읽음. 이전 가지에서는 키워드 벗기기 변이 4개(U1–U4)를 돌려 하나(키워드 하나만 벗김)가 SURVIVED 였다가 키워드 둘이 겹친 행을 더한 뒤 KILLED 였어요; 그 코드는 이제 없고 T12 가 대신 잡아요. 마지막에 `git status` 는 SPEC 만 수정된 상태(내 편집)로 복구.
- 이 조각 밖: `gate-merge` 가 띄우는 `gh api` 자식 프로세스는 hook 이 보지 못해요(hook 은 `moai gate-merge` 한 줄만 봐요; 게이트가 병합을 거절하는 일은 `moai gate-merge` 자신의 몫이에요). 다른 규칙(`green_lock_bash` 등)의 모드는 #3216 이 병합된 뒤 이어가요. G9 의 `Moai-Session` 거절과 감사 brief 기록은 제안만이에요(리더 결정 D-D).
