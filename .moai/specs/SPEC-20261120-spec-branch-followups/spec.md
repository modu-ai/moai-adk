---
id: SPEC-20261120-spec-branch-followups
title: "Follow-ups of the SPEC branch name-first work: the remote-name boundary, the one-form hint and the hook reason"
status: draft
revision: 1
created: 2026-11-20
parent: SPEC-20261119-spec-branch-name-first
---

# Follow-ups of the SPEC branch name-first work: the remote-name boundary, the one-form hint and the hook reason

## Intent

C-112 의 재감사(카드 노트 3418, PASS)가 비차단 후속 넷을 남겼어요(카드 C-133).
- N8: `src/spec_take.rs` 의 `remote_branch_exists` 는 원격 가지 이름이 SPEC 의 옛 꼴 이름과 정확히 같을 때만 막아야 해요. 이 비교를 고정한 시험이 없어서, `==` 를 `starts_with` 로 바꾼 변이 R4 가 살아남았어요. 옛 이름으로 시작할 뿐인 원격 가지(수리 가지 `-r2` 같은 것)가 `spec take` 를 막으면 안 돼요.
- N9: 옛 꼴 가지가 원격에만 있을 때의 거절은 `git switch <옛 가지>` 를 하라고 말해요. 새 꼴 가지가 이미 로컬에 있으면 그 안내를 따른 뒤 두 꼴이 다 생겨서 다음 `spec take` 가 `both forms` 로 또 거절해요. 안내가 한 꼴로 끝나는 단계를 말해야 해요.
- N5: 수리 접미사 때문에 두 SPEC id 로 읽히는 가지 이름(`spec/20261117-demo-r2` 는 `…-demo-r2` 와 `…-demo`)을 hook 이 거절할 때, 이유 문구(`SPEC_TWO_IDS`)는 옛 꼴과 새 꼴만 원인으로 말해요.
- N7: C-112 SPEC 의 `status` 가 `draft` 예요. `approved` 로 바꾸려면 승인 기록(`approval` 의 by·at·basis·digest)이 있어야 하고 그건 승인한 사람의 몫이라, 이 조각은 값을 바꾸지 않고 리더 결정으로 올려요(Decisions D-3).

## Non-goals

- `spec take` 의 거절 조건을 바꾸지 않아요: 어떤 경우에 거절하는지는 그대로이고 문구와 고정 시험만 더해요.
- hook 의 판정(허용·거절)은 바꾸지 않아요: 이유 문구만 고쳐요.
- `moai spec status` 로 승인을 만들어 넣지 않아요.

## Requirements

- REQ-1: The command `spec take` shall not refuse a SPEC because a remote branch only starts with the older branch name of that SPEC.
- REQ-2: The command `spec take` shall name a step that leaves one branch form when the older branch exists only on a remote and the name-first branch exists locally.
- REQ-3: The SPEC-first hook shall name the repair suffix in its reason when a branch name reads as two SPEC ids.

## Acceptance

- AC-1 (REQ-1): 원격 추적 가지 `origin/spec/20261117-login-fix-r2` 와 `origin/spec/20261117-login-fix-extra` 가 있어도 `spec take SPEC-20261117-login-fix` 가 성공하고 name-first 가지를 만들어요. 변이 R4(`==` → `starts_with`)가 이 시험을 실패시켜요.
  verify: `cargo test --locked --offline --test spec_take_followups a_remote_branch_that_only_starts_with_the_old_name_does_not_block_take`
- AC-2 (REQ-2): 거절 문구가 합치고 지우는 단계와 기준 가지로 돌아가는 단계를 말하고, 그 단계를 그대로 따른 뒤의 `spec take` 가 `both forms` 로 거절되지 않아요(성공해요). 새 꼴 가지가 없는 경우의 `git switch` 안내도 기준 가지로 돌아가는 단계를 말하고 따르면 성공해요. 옛 가지 위에서 곧바로 take 하면 `no base` 로 거절돼요.
  verify: `cargo test --locked --offline --test spec_take_followups the_remote_only_hint_ends_in_one_form`
- AC-3 (REQ-3): `spec/20261117-demo-r2` 에 SPEC 폴더가 없을 때 hook 의 거절 이유가 `reads as two SPEC ids` 와 `repair suffix` 를 둘 다 말하고, 입력의 가지 이름은 이유에 나오지 않아요(고정 문구).
  verify: `cargo test --locked --offline --test spec_take_followups the_hook_reason_names_the_repair_suffix`

## Scope

- write: `.moai/specs/SPEC-20261120-spec-branch-followups/spec.md`, `src/spec_take.rs`(거절 문구 한 곳), `src/hook_guards.rs`(`SPEC_TWO_IDS` 문구), `tests/spec_take_followups.rs`(새)
- read-only: `src/spec_branch.rs`, `tests/spec_branch_name_first.rs`
- forbidden: 판정(거절 조건) 바꾸기, 네트워크, 실제 HOME

## Plan

1. RED: 새 시험 파일에 세 시험을 써서 AC-2·AC-3 은 지금 실패하고 AC-1 은 이미 통과함(변이로 의미를 확인)을 눈으로 봐요.
2. GREEN: `spec_take.rs` 의 원격 전용 거절 문구를 나눠 쓰고(새 꼴 로컬 가지가 있으면 합치고 지우는 단계), `hook_guards.rs` 의 이유 문구에 수리 접미사를 더해요.
3. 변이: `remote_branch_exists` 의 `==` → `starts_with`(R4) 가 AC-1 을 실패시키는지, 문구를 옛 값으로 되돌리면 AC-2·AC-3 이 실패하는지 확인하고 되돌려요.
4. 이웃 시험과 여섯 검사를 돌려요.

## Risks

- 이 가지는 C-112 의 READY head(`bd23d9d68e`) 위에서 시작해요: 코드가 로컬 tip 에 아직 없기 때문이에요. 리더가 C-112 를 먼저 합치면 이 가지의 앞부분이 그대로 겹쳐 충돌이 없어요.
- 안내 문구의 단계가 틀리면 사용자가 또 막혀요: 시험이 단계를 실제로 따라 해서 막아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 새 꼴 로컬 가지가 있고 옛 꼴이 원격에만 있을 때의 안내는 `git switch <옛 가지>`, 새 꼴 가지를 그 안에 합치기, 새 꼴 가지 지우기, 기준 가지(예: main)로 돌아가기, 다시 take 의 다섯 단계예요 | 그 단계 뒤에는 로컬에 옛 꼴 하나만 남아서 `both forms` 에 걸리지 않아요; 옛 가지 위에서 곧바로 take 하면 기록된 base 가 없어 `no base` 로 거절돼서(감사 보고의 발견) 돌아가기 단계가 필요해요 | 한 줄 안내로 되돌리기 |
| D-2 | 새 꼴 로컬 가지가 없을 때의 안내는 `git switch <옛 가지>` 한 번 뒤에 기준 가지로 돌아가 take 하라는 말을 더해요 | 그 단계 뒤에 로컬은 옛 꼴 하나뿐이고, 같은 이유로 SPEC 가지 위에서는 take 가 `no base` 로 거절돼요(이 변경 전의 한 줄 안내도 이 단계가 빠져 있었어요) | 없음 |
| D-3 | N7(C-112 SPEC `status: draft`)은 값을 바꾸지 않고 리더에게 결정을 물어요 | `approved` 는 `approval`(by·at·basis·digest)이 있어야 하고 승인은 사용자·리더의 기록이라 레인이 만들어 넣지 않아요; 이 저장소 SPEC 대부분(138개)이 `draft` 예요 | 승인 기록과 함께 값 바꾸기 |

## Result

- RED 먼저(커밋 7851bd3d6c): `cargo test --locked --offline --test spec_take_followups` 에서 AC-1 은 이미 통과(고정 시험을 더한 것이라 변이로 의미를 확인), AC-2(`the_remote_only_hint_ends_in_one_form`)와 AC-3(`the_hook_reason_names_the_repair_suffix`)은 실패했어요(`merge`·`delete` 안내 없음, 이유에 `repair suffix` 없음).
- GREEN(커밋 aa2ad043b1): `src/spec_take.rs` 의 원격 전용 거절이 새 꼴 로컬 가지가 있으면 `git switch` → 합치기 → `git branch -d` → 다시 take 를 말하고(D-1), 없으면 지금 그대로예요(D-2). `src/hook_guards.rs` 의 `SPEC_TWO_IDS` 가 이름 먼저·날짜 먼저 읽기와 수리 접미사 `-r<N>` 의 유무를 원인으로 말해요. 시험 3개 통과(3.28초). 시험 중 `git switch <옛 가지>` 를 실제로 따라 하려면 원격 설정이 있어야 해서 시험 도우미가 `origin` 을 설정해요(주소는 쓰이지 않고 fetch 는 없어요).
- AC-1: met — 변이 R4(`remote_branch_exists` 의 `==` → `starts_with`)가 `a_remote_branch_that_only_starts_with_the_old_name_does_not_block_take` 를 실패시켰어요(`tests/spec_take_followups.rs:31`).
- AC-2: met — 새 꼴 가지 안내를 끈 변이(`if has_new` → `if false`)가 `the_remote_only_hint_ends_in_one_form` 을 실패시켰어요(`:201`, merge·delete 안내 단언). 안내의 단계를 시험이 실제로 따라 한 뒤 `spec take` 가 `both forms` 없이 성공해요.
- AC-3: met — 이유 문구에서 수리 접미사 말을 뺀 변이가 `the_hook_reason_names_the_repair_suffix` 를 실패시켰어요(`:297`). 변이 셋 모두 `git restore` 뒤 `git status` 깨끗.
- 감사 보고(audit-c133, head d52415831f, 차단 0)의 비차단 발견 반영(수리 1): 안내를 글자 그대로 따라 옛 가지 위에서 take 하면 `both forms` 는 안 나오지만 `no base` 로 거절돼요(옛 가지가 원격에서 와서 기록된 base 가 없고 `git branch -d` 가 새 꼴 가지의 기록도 지워요). 시험이 중간에 `git switch main` 을 몰래 더해서 가렸어요. 안내 두 갈래에 "기준 가지로 돌아가 take"를 더하고(D-1·D-2), 시험이 그 말을 단언하고, 옛 가지 위의 take 가 `no base` 로 거절되는 것과 기준 가지에서 성공하는 것을 모두 고정해요. 시험 3개 통과.
- N7: 값을 바꾸지 않았어요(D-3): `approved` 는 승인 기록이 필요해서 리더 결정으로 올렸어요.
- 이웃 시험 통과(가지 시작 시 head `aa2ad043b1`): `--lib` 1568, `spec_take_cli` 15, `spec_branch_name_first` 13, `spec_before_code_hook` 9, `spec_close_cli` 10, `green_test_lock` 14, `trailer_check_cli` 3, `commit_guard_users` 7, `sdd_guidance` 6, `agent_defs_package` 19.
- 판정(거절 조건·허용 조건)은 바꾸지 않았어요: 바뀐 것은 거절 문구 두 곳과 hook 이유 한 문자열이에요.
- 돌리지 않은 것: CI, 실제 HOME, 실제 원격(fetch).
