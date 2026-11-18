---
id: SPEC-20261119-enforce-spec-hooks
title: "The GREEN test lock also holds against Bash deletions and moves, and a denied call leaves its reason in the audit line"
status: draft
revision: 1
created: 2026-11-19
---

# The GREEN test lock also holds against Bash deletions and moves, and a denied call leaves its reason in the audit line

## Intent

카드 C-86 은 "지침 파일은 권고일 뿐이니 강제하라"(사용자 2026-11-17)는 큰 카드이고, 코드를 읽어 보니 AC 대부분이 이미 부분 구현돼 있어요(card note 2277). 이 SPEC 은 그 첫 조각(S1)만 해요: 카드 AC-2 의 둘째 절반(RED 기록이 적은 시험 파일을 GREEN 동안 편집·삭제하지 못하게)에서 지금 뚫려 있는 길 하나와, AC-5 의 감사 로그 빈틈 하나예요.

- `tdd::lock_refusal` 은 `Write`·`Edit`·`MultiEdit`·`NotebookEdit`·`apply_patch`(Delete 포함)만 막아요. Bash 로 `rm tests/x.rs`, `git rm`, `mv`, `unlink` 를 하면 잠긴 시험이 GREEN 동안 사라지거나 바뀌어도 hook 이 몰라요(ImpossibleBench 가 말한 "시험을 고쳐서 통과"의 쉬운 길).
- `record_deny` 가 deny 마다 `hook-engine.jsonl` 에 한 줄을 남기지만 어느 guard 가 막았는지(이유)는 없어요.

## Non-goals

- 카드의 AC-1(인터뷰 기록), AC-3(Stop hook), AC-4(병합 관문): 각각 별도 SPEC·PR 이에요. AC-1·AC-3 은 이 카드의 쓰기 범위 밖 파일이 필요해 범위 요청을 했어요(card note 2278 D-B).
- "SPEC 이 approved 여야 쓰기 허용"으로 조이기(AC-2 의 첫 절반): 지금 hook 은 SPEC 파일이 있으면 통과해요. 승인은 사용자의 대화 중 선택만 줄 수 있어서 조이면 모든 레인이 멈춰요. 정책 결정이라 리더에게 물었고(note 2278 D-A) 기본값은 그대로 두기예요.
- 셸이 파일을 바꾸는 모든 길: 리다이렉트(`>`, `tee`), `sed -i`, `find -delete`, 스크립트, `cd` 로 폴더를 바꾼 뒤의 상대 경로는 이 조각도 막지 못해요(REQ-6 이 한계를 적어요).
- 실제 `~/.claude` 설정과 hook 의 변경, OMP(편집 경로를 번역하지 않아요).
- `src/tdd.rs` 의 잠금 규칙 바꾸기: 이 조각은 `lock_refusal` 을 그대로 불러요.

## Requirements

- REQ-1: While a worktree's TDD record is in GREEN or REFACTOR, the preflight guard shall refuse a Bash command whose `rm`, `unlink` or `git rm` names a test path the record locks, or whose `mv` or `git mv` names such a path as a source or as a destination, with the reason the lock already gives.
- REQ-2: IF the target of `rm`, `unlink` or `git rm` does not exist, THEN the guard shall not refuse the command for that target. · failure: none — removing nothing changes no test, and the lock's rule against a new test file is not meant for it.
- REQ-3: Where the worktree has no TDD record, the record is in RED, no named path is locked, or the repository holds no MoAI state, the guard shall allow the command as before.
- REQ-4: The guard shall read relative paths against the hook input's `cwd` and shall judge a path where it lands (`..` and symlinks resolved), as the file-write lock does.
- REQ-5: When the engine denies a tool call, the line it appends to `hook-engine.jsonl` shall carry the denial's reason text, and shall carry no command text, path or file content.
- REQ-6: The `capabilities` answer of the hook engine shall say that the Bash lock covers `rm`, `unlink`, `git rm`, `mv` and `git mv` only, and that redirections, `sed -i`, `find -delete`, scripts and a command run after a `cd` are not guarded.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): GREEN 기록이 있는 임시 저장소에서 `rm`·`unlink`·`git rm`·`mv`·`git mv` 가 잠긴 시험 경로를 건드리면 deny, 잠기지 않은 경로·RED 단계·기록 없음·MoAI 상태 없는 저장소·존재하지 않는 `rm` 대상은 allow, 상대 경로와 `..` 는 착지한 곳으로 판단.
  verify: `cargo test --locked --offline --test hooks_cli green_lock_bash`
- AC-2 (REQ-5): deny 한 줄에 reason 이 있고 명령·경로·내용은 없어요.
  verify: `cargo test --locked --offline --test hooks_cli deny_audit_line`
- AC-3 (REQ-6): `capabilities` 가 Bash 잠금의 범위와 못 막는 길을 말해요.
  verify: `cargo test --locked --offline --test hooks_cli capabilities_name_the_bash_lock_limits`
- AC-4 (REQ-1…REQ-6): 이웃 시험.
  verify: `cargo test --locked --offline --test hooks_cli`; `cargo test --locked --offline --test spec_before_code_hook`; `cargo test --locked --offline --test green_test_lock`; `cargo test --locked --offline --lib hook_guards::`; `cargo test --locked --offline --lib hook_engine::`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/hook_guards.rs`, `src/hook_engine.rs`, `tests/hooks_cli.rs`, `.moai/specs/SPEC-20261119-enforce-spec-hooks/spec.md`
- read-only: `src/tdd.rs`, `src/claim_verify.rs`, `src/hooks.rs`, `tests/green_test_lock.rs`, `tests/spec_before_code_hook.rs`, `skills/moai/references/internal/sdd.md`
- forbidden: the real `~/.claude` and `~/.moai-cli` (tests use a temporary HOME), `src/tdd.rs`, `src/spec_doc.rs`, any change that lets a hook allow what it refused before

## Plan

1. 시험 먼저(`tests/hooks_cli.rs`): GREEN 기록(`.git/moai/tdd.json`)을 임시 저장소에 직접 쓰고 `moai hook-dispatch --host claude --event tool-preflight` 로 Bash 입력을 먹여 deny/allow 를 읽어요. 없는 함수·필드 때문에 deny 시험이 실패하는 것(RED)을 봐요.
2. `bash_lock_guard(command, cwd)`: `shell_commands` 로 `rm`·`unlink`·`git rm`·`mv`·`git mv` 의 경로를 모아 `tdd::lock_refusal` 에 넘기고, `hook_engine` 의 Bash 분기에서 `bash_guard` 다음에 불러요.
3. `record_deny` 줄에 `reason`.
4. `capabilities` 문구.
5. 이웃 시험, mutation(가드마다 한 번).

## Risks

- 셸 파싱은 한계가 있어요(Non-goals). 이 조각은 "막을 수 있는 가장 흔한 길"을 막을 뿐이고, 못 막는 길은 `capabilities` 가 말해요. 허위로 안전하다고 말하지 않아요.
- `lock_refusal` 의 둘째 조건("새 시험 파일")은 존재하지 않는 경로를 거부해요. `rm` 대상이 없는 경우를 건너뛰지 않으면 `rm -f tests/gone.rs` 가 헛거부돼요(REQ-2).
- deny 줄에 필드를 더하면 줄 모양을 고정한 시험이 깰 수 있어요(AC-4 가 이웃 시험으로 확인해요).

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | `lock_refusal` 을 바꾸지 않고 불러요 | 규칙은 한 곳(`tdd.rs`)에 있어야 하고 이 카드의 쓰기 범위 밖이에요 | Bash 용 규칙을 따로 만들기(두 곳이 어긋남) |
| D-2 | 존재하지 않는 `rm` 대상은 건너뛰어요 | 지울 게 없으니 시험이 바뀌지 않아요 | 모두 `lock_refusal` 에 넘기기(헛거부) |
| D-3 | 감사 줄에는 정적 reason 문장만 더하고 명령·경로는 안 남겨요 | guard 문장은 정적 글이라 비밀·경로가 없고, 명령 글은 비밀을 품을 수 있어요 | 명령 해시 |
| D-4 | "approved 여야 쓰기 허용"은 이 조각에서 안 해요 | 정책 결정(note 2278 D-A, 기본값 그대로) | 선택 B 또는 C |

## Result

(채워지지 않았어요: 구현 뒤에 AC 마다 한 줄을 적어요.)
