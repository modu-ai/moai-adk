---
id: SPEC-20261124-home-env-readers
title: "환경 HOME 을 직접 읽는 곳을 명령 가장자리로 모아요: 1조각은 layout 이 HOME 을 받아요"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7680 (2026-11-23T16:23:19Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단이고 사용자 직접 선택은 아님, 노트 7677 의 결정 1 (a): 판 1 승인, 조각 1 은 src/layout.rs 와 그 테스트·SPEC 안만); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:f5ca20be04d83880689839b23b79d7a58bca6200a8e07d306bc57039763c579e"
---

# 환경 HOME 을 직접 읽는 곳을 명령 가장자리로 모아요: 1조각은 layout 이 HOME 을 받아요

## Intent

카드 C-410 은 C-409(identity 가 HOME 을 인자로 받음)의 둘째 갈래예요. C-409 노트 7613 의 D-3 대로, `layout::default_home_root`(`src/layout.rs:34`)가 환경 변수 HOME 을 읽고, 클레임 장부 경로(`claim::journal_root` → `layout::switched_dir`)가 그것을 거쳐요. identity 만 고치면 클레임 묶음 테스트(`claim_link_cli` 등)가 실제 HOME 을 읽는 일이 다 사라지지 않아요. 같은 종류(C-406 원인 + C-105 감사 비차단 1)가 두 번이라 소스를 고치는 카드예요.

방식은 C-409 와 같아요(그 SPEC 의 D-1): HOME 을 인자로 받는 `_in` 짝을 더하고, 기존 이름은 환경 HOME 을 한 번 읽어 위임하는 가장자리 어댑터로 남겨요. 이 카드의 쓰기 범위는 `src/layout.rs` 와 이 SPEC 폴더예요. 그래서 1조각은 layout 안만 고치고, 호출부와 다른 읽는 곳은 조각 2 이후에 리더 노트로 범위를 받아 이어가요.

C-409 N1(감사 #409)도 이 카드 몫이에요: `identity_in` 이 부르는 git 자식은 프로세스 환경을 그대로 받아 환경 HOME 의 전역 git 설정을 읽어요. 그 고정은 `src/identity.rs` 안이라 1조각의 범위 밖이고, 조각 2 에 옵션과 함께 올려요(D-4).

## Non-goals

- 호출부 파일(`src/claim.rs`, `src/forge/outbox.rs`, `src/apply.rs`)과 아래 인벤토리의 다른 모듈을 이 조각에서 고치는 일 — 조각 2 이후.
- 명령 가장자리(`src/main.rs`, `src/ops/*::run`)의 읽기를 없애는 일. 가장자리는 한 번 읽는 자리라 남겨요.
- `tests/fixtures/test_home_guard/baseline.tsv` 줄 지우기 — C-406 가드가 아직 착지하지 않았어요(읽어서 확인: NeXT `3d864da6f4` 에 그 폴더가 없어요).
- git 자식의 전역 설정 고정(D-4, 조각 2).

## Requirements

- REQ-1: The `default_home_root_in` function shall return the home root of the HOME it is given and read no environment variable.
- REQ-2: IF the HOME given to `default_home_root_in` is absent or empty, THEN it shall return no home root.
- REQ-3: The `switched_dir_in` function shall resolve the switched store directory from the HOME it is given and from no environment variable.
- REQ-4: The `default_home_root` function shall return what `default_home_root_in` returns for the environment HOME read once, an unset or empty HOME counting as no HOME.
- REQ-5: The `claims_legacy_dir` function shall resolve the clone identity without a HOME, because it reads no field that depends on one.

## Acceptance

- AC-1 (REQ-1, REQ-2): 환경 변수 HOME 과 다른 임시 폴더를 넘기면 홈 뿌리가 그 폴더 아래 경로이고, `None` 과 빈 경로는 `None` 이에요. · verify: `cargo nextest run --offline --locked --lib -E 'test(default_home_root_in)'`
- AC-2 (REQ-3): 넘긴 HOME 아래에 `layout.json` 이 있는 프로젝트는 그 HOME 의 저장소 폴더를 답하고, 자식 프로세스의 환경 HOME 이 다른 폴더이거나 없어도 답이 같아요. · verify: `cargo nextest run --offline --locked --lib -E 'test(switched_dir_in)'`
- AC-3 (REQ-4): 환경 HOME 이 설정·빈 값·없음 세 경우에 `default_home_root()` 의 답이 `default_home_root_in` 에 같은 값을 넘긴 답과 같아요(자식 프로세스에서 환경을 고정, 부모는 환경을 바꾸지 않아요). · verify: `cargo nextest run --offline --locked --lib -E 'test(default_home_root_equals)'`
- AC-4 (REQ-5): 환경 HOME 의 홈 상태 폴더가 심볼릭 링크인 자식 프로세스에서도 `claims_legacy_dir` 가 거부 없이 같은 클레임 원본을 답해요(옛 코드는 `identity` 를 거쳐 거부했어요). · verify: `cargo nextest run --offline --locked --lib -E 'test(claims_legacy_dir_ignores)'`
- AC-5 (REQ-1..REQ-5): 이 조각은 `src/layout.rs` 와 SPEC 폴더만 바꿨어요. · verify: `git diff --stat <base>..<head>`

## Scope

- write: `src/layout.rs`, `.moai/specs/SPEC-20261124-home-env-readers/spec.md`
- read-only: `src/identity.rs`, `src/home_root.rs`, `src/claim.rs`, `src/forge/outbox.rs`, `src/apply.rs`, `src/worktree.rs`
- forbidden: 실제 HOME(읽기 전후 비교로도 건드리지 않아요, 임시 HOME 픽스처만), `tests/fixtures/test_home_guard/baseline.tsv`, `CLAUDE.md`, 네트워크

## Plan

인벤토리(2026-11-24, `rg 'env::var(_os)?\("HOME"\)' src` 와 함수 이름 확인; 본문 분류는 함수 이름만 보았어요 — 조각마다 읽어 확인하고 증거로 남겨요, 그 전은 unverified):

| 구분 | 곳 | 조각 |
| --- | --- | --- |
| 가장자리(한 번 읽는 자리, 남겨요) | `src/main.rs` 여섯 곳, `src/ops/*::run` 열네 곳(`args.home` 이 없으면 환경), `src/ledger/command.rs::resolve_home`, `src/hook_entry.rs::env`, `src/statusline/mod.rs::run`, `src/identity.rs::identity`(C-409) | — |
| layout 안 | `src/layout.rs::default_home_root`(읽는 곳), `switched_dir`, `claims_legacy_dir`(`identity` 호출) | 1(이 판) |
| layout 호출부 | `src/claim.rs::journal_root`(`switched_dir`), `src/forge/outbox.rs:202`(`switched_dir`), `src/apply.rs:368`(`switched_dir`), `src/worktree.rs::open_locked_journal`(`apply_at` 이 부르며 `claim::journal_root` 를 거쳐 환경을 읽어요) | 2 |
| git 자식 | `src/identity.rs::rev_parse`(환경 HOME 의 전역 git 설정을 읽음, C-409 N1) | 2 (D-4) |
| 안쪽 읽기 A | `src/key_vault.rs::home`, `src/home_workdir.rs::home`, `src/hooks.rs::env_home`, `src/hook_config.rs::load`, `src/hook_engine.rs::session_check`, `src/edit_diagnostics.rs::lock_dir`, `src/heavy_slot.rs::cache_dir` | 3 |
| 안쪽 읽기 B | `src/claude.rs::doctor`·`refuse_user_scope`, `src/project_update.rs::go_settings_step`·`settings_step`, `src/baseline.rs::inventory`, `src/migrate.rs::open_cli_home`, `src/update_tux.rs::view_for`, `src/ops/host_wake.rs::default_config_dir`, `src/ops/star_ask.rs::run` | 4 |
| 줄 지우기 | `tests/fixtures/test_home_guard/baseline.tsv` | C-406 착지 뒤 |

1조각 순서: 1) 승인 노트를 받아요. 2) RED: 위 테스트 이름으로 `default_home_root_in`·`switched_dir_in` 이 없어 컴파일이 실패하는 것을 봐요. 3) GREEN: `default_home_root_in`·`switched_dir_in` 을 더하고 `default_home_root`·`switched_dir` 는 어댑터로, `claims_legacy_dir` 는 `identity_in(project, None)` 으로 바꿔요. 4) 이웃: `layout` 모듈 기존 테스트와 `claim` 의 `journal_root` 테스트, fmt, clippy.

## Risks

- 이 조각만으로는 실제 HOME 읽기가 줄었는지 못 재요: 호출부가 아직 `switched_dir`(어댑터)를 불러요. 카드 기준 1 은 부분 충족, 기준 2 는 not met 으로 적고, 줄어든 수는 C-406 가드(observe)가 착지한 뒤에 재요(그 전은 unverified, C-409 D-4 와 같아요).
- REQ-5 는 `plan` 이 홈 상태 폴더가 심볼릭 링크일 때 내던 거부를 클레임 원본 찾기에서 없애요(그 거부는 클레임 원본과 관계없는 칸이었어요). 동작이 바뀌는 곳은 이것 하나이고 D-3 에 적었어요.
- `apply_at` 의 저널 잠금(`open_locked_journal`)은 이 조각이 닿지 않는 `claim::journal_root` 를 거쳐 환경 HOME 을 읽어요. 조각 2 가 닫기 전까지 `apply_cli` 경로는 환경 HOME 을 읽는 채로 남아요.
- 인벤토리의 본문 분류는 함수 이름까지만 확인했어요. 조각마다 읽어서 고쳐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `default_home_root_in(home: Option<&Path>)` 와 `switched_dir_in(project, subdir, home: Option<&Path>)` 를 더하고 기존 두 함수는 환경 HOME 을 한 번 읽어 위임하는 어댑터로 남겨요 | C-409 D-1 과 같은 방식; 호출부 세 곳이 이 카드 범위 밖이라 서명을 바로 바꿀 수 없어요 | `_in` 두 함수를 지우고 본문을 되돌리기 |
| D-2 | 가장자리(`main.rs`, `ops/*::run`, `ledger::resolve_home`)의 읽기는 남겨요 | 카드 기준 1 은 "명령 가장자리 밖에서는 0" 이에요; 가장자리가 한 번 읽는 자리예요 | 해당 없음 |
| D-3 | `claims_legacy_dir` 는 `identity_in(project, None)` 을 불러요 | 그 함수는 `clone_anchor` 만 읽어요(`src/layout.rs:87`); 홈 상태 폴더 칸은 쓰지 않아요 | 이전 줄 `identity(project)` 로 되돌리기 |
| D-4 | git 자식의 전역 설정 고정(C-409 N1)은 조각 2 에서 `src/identity.rs` 범위를 받아 해요. 옵션: (a) `rev_parse` 자식에 `GIT_CONFIG_GLOBAL` 을 받은 HOME 의 `.gitconfig`(없으면 `/dev/null`)로 줘요 (b) `GIT_CONFIG_NOSYSTEM`·`GIT_CONFIG_GLOBAL=/dev/null` 로 둘 다 막아요 (c) 고치지 않고 문서로만 남겨요. 기본 (a) — 단, `safe.directory` 가 전역 설정에서 오므로 (b) 는 소유자가 다른 저장소를 거부할 수 있어요(미확인, 조각 2 에서 git 문서와 실험으로 확인) | 이 조각에서 정하지 않고 범위 확장과 함께 리더 결정을 받아요 | 해당 없음 |

## Interview

- Q1: 서명을 바로 바꾸나요? → A: 아니요(D-1). `switched_dir` 호출부 세 곳이 범위 밖이에요.
- Q2: 환경 HOME 읽기를 전부 이 카드 하나로 끝내나요? → A: 아니요. 인벤토리를 조각으로 나눠요; 조각 2 이후는 시작할 때 리더 노트로 범위를 받아요.

## Result

조각 1 의 구현 커밋은 `6ee7cdefb2`(브랜치 `spec/home-env-readers-20261124`, 기준 `3d864da6f4`)이고 승인은 리더 노트 7680 을 인용했어요. 로그는 `target/c410/` 에 있고 카드가 열려 있는 동안만 둬요.

- RED: `default_home_root_in`·`switched_dir_in` 이 없어 컴파일이 실패했어요(`cargo nextest run --offline --locked --lib -E 'test(home_arg_tests)'` → exit 101, 오류 6개). 두 함수를 더한 뒤 같은 명령이 exit 100: 9개 중 8 통과, `claims_legacy_dir_ignores_the_environment_home` 만 실패했어요(`claims_legacy_dir` 가 환경 HOME 의 심볼릭 링크 상태 폴더 때문에 거부당했어요).
- GREEN: `claims_legacy_dir` 를 `identity_in(project, None)` 으로 바꾼 뒤 `-E 'test(layout::) | test(journal_root) | test(claim::)'` → exit 0, 36 통과(새 테스트 5개 포함).
- AC-1: 충족. `default_home_root_in_uses_the_home_it_is_given`, `default_home_root_in_without_a_home_is_none` 통과(없음과 빈 경로가 `None`).
- AC-2: 충족. `switched_dir_in_follows_the_home_it_is_given_not_the_environment` 통과(자식 프로세스의 환경 HOME 에는 포인터가 없고 넘긴 HOME 에만 있을 때, `switched_dir_in` 은 넘긴 HOME 아래를 답하고 어댑터 `switched_dir` 는 `None`).
- AC-3: 충족. `default_home_root_equals_the_environment_home_read_once` 통과(자식 프로세스 셋: 설정·빈 값·없음, 부모는 환경을 바꾸지 않아요).
- AC-4: 충족. `claims_legacy_dir_ignores_the_environment_home` 통과(자식의 환경 HOME 이 심볼릭 링크 상태 폴더이고 `identity` 는 거부되는데 `claims_legacy_dir` 는 클레임 폴더를 답해요).
- AC-5: 충족. `git diff --stat 3d864da6f4..HEAD` 가 `src/layout.rs` 와 이 SPEC 두 파일만 보여요.
- 뮤턴트(테스트마다 하나, 적용을 `git diff --stat` 으로 확인하고 되돌린 뒤 `git diff --stat` 이 비어 있음을 확인; 모두 `-E 'test(layout::home_arg_tests)'`): ① `default_home_root_in` 의 빈 HOME 거르기 삭제 → exit 100 (2개 실패: `..._without_a_home_is_none`, `..._equals_the_environment_home_read_once`) ② `switched_dir_in` 이 환경 HOME 을 먼저 씀 → exit 100 (1개 실패) ③ `default_home_root` 가 환경 HOME 대신 `None` 을 넘김 → exit 100 (1개 실패) ④ `claims_legacy_dir` 를 `identity(project)` 로 되돌림 → exit 100 (1개 실패).
- 수리: head `fb4305a260` 에서 돌린 `check affected` 가 테스트 하나, `git_env_scrub::every_git_subprocess_in_src_is_scrubbed`(3702개 중 3701 통과)를 실패시켰어요: 새 테스트의 `git init`(`src/layout.rs:771`)이 `scrub_git_env` 를 거치지 않았어요. 앞 확인은 `--lib` 테스트만 돌려 이 저장소 전체 스캔을 놓쳤어요(내 확인의 틈). 테스트의 `git init` 이 `crate::identity::scrub_git_env` 를 거치게 고쳤고(커밋 `c360414924`), `cargo nextest run --lib --test git_env_scrub -E 'test(layout::) | test(every_git_subprocess)'` → exit 0, 21 통과예요.
- 카드 기준 1: 부분 충족. 이 모듈의 환경 HOME 읽기는 `environment_home` 한 곳(어댑터 `default_home_root`·`switched_dir` 가 씀)이고 `claims_legacy_dir` 의 읽기는 사라졌어요. 호출부(`claim::journal_root`, `forge/outbox.rs`, `apply.rs`, `worktree::open_locked_journal`)가 아직 어댑터를 불러요(조각 2).
- 카드 기준 2: 충족하지 않았어요(not met). 호출부 변경이 없고, C-406 가드(`tests/fixtures/test_home_guard/`)가 NeXT 에 없어 기준선 줄을 재지도 못 했어요(unverified).
- 실제 HOME: 새 테스트는 부모 프로세스에서 환경 HOME 을 읽는 함수를 부르지 않고 임시 폴더만 넘겨요. 환경 HOME 이 필요한 경우는 자식 프로세스에 임시 폴더 HOME 을 줘요. 실제 HOME 의 읽기 전후 비교는 재지 않았어요(not_run).
- 후속: 조각 2(호출부 세 곳·`worktree.rs`·`identity.rs` 의 git 자식 고정, 결정 3 은 실험 뒤)는 시작할 때 리더 노트로 범위를 받아요.
