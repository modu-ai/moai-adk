---
id: SPEC-20261119-codex-conflict-choices
title: "A Codex config conflict comes with keep, diff and replace-with-backup"
status: draft
revision: 1
created: 2026-11-19
---

# A Codex config conflict comes with keep, diff and replace-with-backup

## Intent

`/moai project`가 Codex 설정을 계획할 때, 사용자가 이미 `.codex/config.toml`의 `tui.status_line`을 직접 정해 두었으면 `codex-config` 단계가 `conflict`로 끝나고 **고를 수 있는 것이 없어요**. 규칙("손으로 고치지 않는다")과 합쳐지면 막다른 길이에요. mo.ai.kr에서 `/moai project`를 돌린 기록(2026-11-16)에서 실제로 일어났어요.

이 SPEC은 그 충돌 하나에 세 가지 선택지를 붙여요. 각 선택지는 세션이 사용자의 선택 뒤에 그대로 돌릴 수 있는 명령을 가져요(권장이 맨 앞).

| 선택지 | 하는 일 | 명령 |
| --- | --- | --- |
| `keep` (권장) | 사용자의 파일을 그대로 둬요. 아무것도 쓰지 않아요 | 없음 |
| `diff` | 사용자의 값과 MoAI가 쓸 값을 나란히 보여 줘요(읽기 전용) | `codex-config-plan … --replace-with-backup` |
| `replace` | 원래 파일을 백업한 뒤 MoAI 값으로 바꿔요 | `codex-config-apply … --replace-with-backup --plan-digest <d>` |

같은 카드의 나머지 조각은 이 SPEC 밖이에요(Non-goals).

## Non-goals

- `codex-hooks` 충돌(`.codex/hooks.json` 전체를 바꾸는 일): 파일 전체 교체라서 쓰는 쪽 규칙이 달라요 — 다음 SPEC.
- 표식 없는 `AGENTS.md`, 번들 풀기, `CLAUDE.md` 그림자: `src/agents.rs` 조각 — 그 파일이 풀리면 별도 SPEC.
- 워크플로 글(`skills/moai/**`)에 선택지 문장 넣기: manifest 예약이 풀려야 해요 — 별도 SPEC.
- 백업을 되돌리는 명령: 백업은 원래 파일 그대로의 사본이라 사용자가 복사해 되돌릴 수 있어요. 별도 명령은 만들지 않아요.
- 신뢰(trust) 키 변경: 지금처럼 쓰지 않아요.

## Requirements

- REQ-1: The codex-config plan shall list, for each conflicting key, the value the file holds and the value MoAI would write.
- REQ-2: WHEN `--replace-with-backup` is given to `codex-config-plan` and a key conflicts, the plan shall report the action `replace`, the replaced keys, the backup path and its own digest without writing anything.
- REQ-3 (REQ-2): IF `codex-config-apply --replace-with-backup` is given a digest other than the one of the replace plan, THEN it shall refuse the apply as a stale plan and write nothing.
- REQ-4: WHEN `codex-config-apply --replace-with-backup` is given the digest of the replace plan, the apply shall write the backup, then the config, then the ownership receipt.
- REQ-5 (REQ-4): IF a backup file of the planned name holds other bytes, THEN the apply shall refuse before it writes the config.
- REQ-6: The backup shall hold the exact bytes the config had before the apply, under the project state folder.
- REQ-7: IF `codex-config-apply` is given a conflicting config without `--replace-with-backup`, THEN it shall refuse with a conflict and write nothing.
- REQ-8: IF `--replace-with-backup` is given and no key conflicts, THEN the plan and the apply shall behave as without the flag and write no backup.
- REQ-9: WHEN `project-update-plan` finds the `codex-config` conflict, its step shall carry the choices `keep`, `diff` and `replace` in that order with `keep` marked recommended.
- REQ-10 (REQ-9): IF the replace plan cannot be made, THEN the step shall carry the choices `keep` and `diff` only.
- REQ-11: The choices shall carry these commands: `keep` none, `diff` the read-only replace plan, `replace` the apply with the digest of the replace plan and the backup path.
- REQ-12: The plan of the same request that follows a replace apply shall be `unchanged` with every other key and comment of the config kept byte for byte.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-8): 충돌 설정의 계획에 키별 `diff`가 있고, `--replace-with-backup` 계획이 `replace`·백업 경로·다른 digest를 내며 아무것도 쓰지 않아요. 충돌이 없으면 플래그가 있어도 달라지지 않아요.
  verify: `cargo test --locked --offline --test codex_conflict_choices_cli plan`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-12): replace 적용이 백업 → 설정 → 영수증 순으로 쓰고, 백업은 원본 바이트이며, 플래그 없는 적용·일반 계획의 digest·다른 내용의 같은 이름 백업은 거절돼요. 적용 뒤 계획은 `unchanged`예요.
  verify: `cargo test --locked --offline --test codex_conflict_choices_cli apply`
- AC-3 (REQ-9, REQ-10, REQ-11): `project-update-plan`의 `codex-config` 단계가 `keep`·`diff`·`replace`를 순서대로 싣고 출력된 명령이 그대로 돌아가요. replace 계획이 없으면 `keep`·`diff`만 남아요.
  verify: `cargo test --locked --offline --test codex_conflict_choices_cli step`; `cargo test --locked --offline --lib project_hosts::tests::conflict_choices`
- AC-4: 이웃 검사.
  verify: `cargo test --locked --offline --test codex_config_cli`; `cargo test --locked --offline --test codex_config`; `cargo test --locked --offline --test ac_command_lint`

## Scope

- write: `.moai/specs/SPEC-20261119-codex-conflict-choices/spec.md`, `src/hosts/codex_config.rs`, `src/project_hosts.rs`, `src/ops/codex_config_plan.rs`, `src/ops/codex_config_apply.rs`, `tests/codex_conflict_choices_cli.rs`
- read-only: `src/project_update.rs`, `src/hooks.rs`, `src/agents.rs`
- forbidden: 실제 HOME, `skills/moai/**`, `src/agents.rs` 편집

## Plan

1. 시험을 먼저 써요(RED): 플래그가 없어서 모두 실패해야 해요.
2. `hosts/codex_config.rs`: 충돌을 값 쌍으로 싣고, replace 모드(충돌 키를 바꿈으로 취급, digest에 모드 포함)와 백업 쓰기를 더해요. 기존 `config_plan`·`config_apply`는 모드 없는 형태로 남겨 호출 자리를 건드리지 않아요.
3. `ops/codex_config_{plan,apply}.rs`: `--replace-with-backup` 플래그.
4. `project_hosts.rs`: 충돌 단계의 `preview.choices`.
5. 이웃 시험, 변이 시험, fmt·clippy.

## Risks

- replace 모드 digest가 일반 모드와 같으면 일반 계획을 본 사용자의 승인이 교체에 쓰여요 → REQ-3과 시험이 지켜요.
- 백업을 쓰다 중간에 멈추면 원래 파일은 그대로고 백업만 있어요 → 순서가 백업 먼저라서 안전한 쪽이에요. 같은 바이트의 백업이 이미 있으면 그대로 두고 이어가요(재시도 가능).
- 백업 이름이 겹쳐 다른 내용을 덮어쓰면 원본이 사라져요 → 이름에 원본 해시를 넣고, 같은 이름에 다른 바이트는 거절해요(REQ-5).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 이번 SPEC은 Codex 설정만, 훅은 다음 SPEC | 훅은 파일 전체 교체라 쓰는 쪽이 달라 한 PR에 섞으면 감사가 어려워요 | 없음(범위 결정) |
| D-2 | 백업은 상태 폴더(`.moai-cli/backups/`) 아래 | 상태 폴더는 관리 `.gitignore` 블록이 이미 무시해서, 추적 파일 옆에 사본이 생기지 않아요 | 파일 위치만 바뀌어요 |
| D-3 | 선택지는 `preview.choices`에 실어요 | `HostStep.preview`는 값 그대로 단계 JSON에 실려서 `project_update.rs`를 고치지 않아도 돼요 | 없음 |
| D-4 | 명령 인자로 `--replace-with-backup` | 교체가 이름에 드러나고, 플래그 없는 명령은 지금과 똑같이 충돌을 거절해요 | 플래그만 지우면 돼요 |

## Result

(구현 뒤에 채워요)
