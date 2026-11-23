---
id: SPEC-20261124-home-env-readers
title: "환경 HOME 을 직접 읽는 곳을 명령 가장자리로 모아요: 조각 1 은 layout, 조각 2 는 클레임 장부 호출부와 git 자식"
status: draft
revision: 3
created: 2026-11-24
---

# 환경 HOME 을 직접 읽는 곳을 명령 가장자리로 모아요: 조각 1 은 layout, 조각 2 는 클레임 장부 호출부와 git 자식

## Intent

카드 C-410 은 C-409(identity 가 HOME 을 인자로 받음)의 둘째 갈래예요. C-409 노트 7613 의 D-3 대로, `layout::default_home_root`(`src/layout.rs:34`)가 환경 변수 HOME 을 읽고, 클레임 장부 경로(`claim::journal_root` → `layout::switched_dir`)가 그것을 거쳐요. identity 만 고치면 클레임 묶음 테스트(`claim_link_cli` 등)가 실제 HOME 을 읽는 일이 다 사라지지 않아요. 같은 종류(C-406 원인 + C-105 감사 비차단 1)가 두 번이라 소스를 고치는 카드예요.

방식은 C-409 와 같아요(그 SPEC 의 D-1): HOME 을 인자로 받는 `_in` 짝을 더하고, 기존 이름은 환경 HOME 을 한 번 읽어 위임하는 가장자리 어댑터로 남겨요. 이 카드의 쓰기 범위는 `src/layout.rs` 와 이 SPEC 폴더예요. 그래서 1조각은 layout 안만 고치고, 호출부와 다른 읽는 곳은 조각 2 이후에 리더 노트로 범위를 받아 이어가요.

C-409 N1(감사 #409)도 이 카드 몫이에요: `identity_in` 이 부르는 git 자식은 프로세스 환경을 그대로 받아 환경 HOME 의 전역 git 설정을 읽어요. 그 고정은 `src/identity.rs` 안이라 1조각의 범위 밖이었고, 조각 2 에서 해요(D-4, D-8).

판 2(조각 2)는 세 가지를 해요. ① 클레임 장부 호출부가 HOME 을 받아요: `claim::journal_root_in`, `worktree::open_locked_journal_in` 을 더하고 `layout::apply_at` 의 저널 잠금이 환경 HOME 을 읽지 않게 해요. ② git 자식의 전역 설정을 받은 HOME 으로 고정하고, 조각 1 테스트의 부모 쪽 git 자식이 임시 폴더 HOME 을 쓰게 해요(감사 #410 N2). ③ C-406 가드가 착지했으니(NeXT `132728a026`) 이 변경으로 더는 HOME 을 건드리지 않게 된 테스트의 기준선 줄을 같은 변경에서 지워요. 줄이 몇 개 줄어드는지는 변경 뒤 관찰 실행으로 재요(그 전은 unverified).

판 3(조각 2 의 마무리)은 하나를 더해요. 클레임 연결 테스트(`tests/claim_link_cli.rs`)의 도우미 `record()` 가 `identity::identity` 를 불러 그 테스트 스물여섯 개가 가드 기준선에 남아 있었어요. 그 도우미가 읽는 칸(`clone_anchor`, `workspace_git_dir`, `work_tree`)은 HOME 과 무관하므로 `identity_in(.., None)` 으로 바꾸고, 관찰 실행으로 가드에서 빠진 것을 확인한 스물다섯 줄을 같은 변경에서 지워요(리더 결정 노트 7830).

## Non-goals

- 아래 인벤토리의 안쪽 읽기 A·B 모듈 — 조각 3 이후. `src/forge/outbox.rs`, `src/apply.rs` 는 관찰 실행 뒤에 정해요(D-7).
- 명령 가장자리(`src/main.rs`, `src/ops/*::run`)의 읽기를 없애는 일. 가장자리는 한 번 읽는 자리라 남겨요.
- `identity::identity` 호출을 `identity_in` 으로 바꾸는 일 — C-409 의 몫이에요. 이 카드는 `journal_root_in` 이 필요한 곳만 바꿔요.

## Requirements

- REQ-1: The `default_home_root_in` function shall return the home root of the HOME it is given and read no environment variable.
- REQ-2: IF the HOME given to `default_home_root_in` is absent or empty, THEN it shall return no home root.
- REQ-3: The `switched_dir_in` function shall resolve the switched store directory from the HOME it is given and from no environment variable.
- REQ-4: The `default_home_root` function shall return what `default_home_root_in` returns for the environment HOME read once, an unset or empty HOME counting as no HOME.
- REQ-5: The `claims_legacy_dir` function shall resolve the clone identity without a HOME, because it reads no field that depends on one.
- REQ-6: The `journal_root_in` function shall resolve the claim journal folder from the HOME it is given and from no environment variable.
- REQ-7: The `journal_root` function shall return what `journal_root_in` returns for the environment HOME read once.
- REQ-8: The `open_locked_journal_in` function shall lock the journal that `journal_root_in` names for the HOME it is given.
- REQ-9: The `apply_at` function shall lock the project-side claim journal without reading the environment HOME, because it refuses an already switched pointer before it locks.
- REQ-10: The `identity_in` function shall run its git children with the global git configuration of the HOME it is given, and with the process's own lookup when it is given none.
- REQ-11: The layout tests shall run every git child they start with a HOME in a temporary folder.
- REQ-12: The baseline of `tests/fixtures/test_home_guard/baseline.tsv` shall list no test that the observation run of the changed tree shows leaving its HOME untouched.
## Acceptance

- AC-1 (REQ-1, REQ-2): 환경 변수 HOME 과 다른 임시 폴더를 넘기면 홈 뿌리가 그 폴더 아래 경로이고, `None` 과 빈 경로는 `None` 이에요. · verify: `cargo nextest run --offline --locked --lib -E 'test(default_home_root_in)'`
- AC-2 (REQ-3): 넘긴 HOME 아래에 `layout.json` 이 있는 프로젝트는 그 HOME 의 저장소 폴더를 답하고, 자식 프로세스의 환경 HOME 이 다른 폴더이거나 없어도 답이 같아요. · verify: `cargo nextest run --offline --locked --lib -E 'test(switched_dir_in)'`
- AC-3 (REQ-4): 환경 HOME 이 설정·빈 값·없음 세 경우에 `default_home_root()` 의 답이 `default_home_root_in` 에 같은 값을 넘긴 답과 같아요(자식 프로세스에서 환경을 고정, 부모는 환경을 바꾸지 않아요). · verify: `cargo nextest run --offline --locked --lib -E 'test(default_home_root_equals)'`
- AC-4 (REQ-5): 환경 HOME 의 홈 상태 폴더가 심볼릭 링크인 자식 프로세스에서도 `claims_legacy_dir` 가 거부 없이 같은 클레임 원본을 답해요(옛 코드는 `identity` 를 거쳐 거부했어요). · verify: `cargo nextest run --offline --locked --lib -E 'test(claims_legacy_dir_ignores)'`
- AC-5 (REQ-1..REQ-5): 조각 1 은 `src/layout.rs` 와 SPEC 폴더만 바꿨어요(승인된 판 1 의 범위). · verify: `git diff --stat 3d864da6f4..4d22c99aac`
- AC-6 (REQ-6, REQ-7): 환경 HOME 의 프로젝트에는 전환 포인터가 없고 넘긴 HOME 에만 있을 때 `journal_root_in` 은 넘긴 HOME 아래 클레임 폴더를, `journal_root` 는 프로젝트 쪽 폴더를 답해요; `None` 은 프로젝트 쪽 폴더예요(자식 프로세스에서 환경 HOME 을 고정). · verify: `cargo nextest run --offline --locked --lib -E 'test(journal_root_in)'`
- AC-7 (REQ-8): 넘긴 HOME 에 전환 포인터가 있으면 `open_locked_journal_in` 이 잠그는 저널이 그 HOME 아래 클레임 폴더의 것이고, 같은 폴더를 이미 잠근 뒤에는 두 번째 호출이 사용 중 응답을 해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(open_locked_journal_in)'`
- AC-8 (REQ-9): `apply_at` 이 환경 HOME 에 전환 포인터가 있는 자식 프로세스에서도 프로젝트 쪽 저널을 잠가요(환경 HOME 을 읽지 않으니 답이 같아요); 이미 전환된 홈 폴더에는 지금처럼 `conflict` 로 거절해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(apply_at_locks)'`
- AC-9 (REQ-10): 소유자가 다른 저장소(`GIT_TEST_ASSUME_DIFFERENT_OWNER`)에서, 받은 HOME 의 `.gitconfig` 에 `safe.directory` 가 있으면 `identity_in(path, Some(home))` 이 성공하고 없으면 거절 문구로 실패해요; 자식 프로세스의 환경 HOME 이 어느 쪽이든 답이 같아요. · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_in_git_config)'`
- AC-10 (REQ-11): 조각 1 테스트(`claims_legacy_dir_ignores_the_environment_home`)의 `git init` 자식이 임시 폴더 HOME 으로 돌아요. · verify: `cargo nextest run --offline --locked --lib -E 'test(claims_legacy_dir_ignores)'` 와 `git diff <base>..<head> -- src/layout.rs` 로 `HOME` 을 주는 줄을 읽어요
- AC-11 (REQ-12): 줄을 지운 기준선이 `tests/test_home_guard.rs` 의 기준선 테스트를 통과하고(상한은 남은 줄 수 361 로 낮춰요), 지운 줄의 테스트를 관찰 모드로 돌렸을 때 가드 줄이 나오지 않아요. · verify: `cargo nextest run --offline --locked --test test_home_guard --test claim_link_cli --test hook_dispatch_cli` 와 `MOAI_HOME_GUARD=observe MOAI_HOME_GUARD_LOG=<파일> cargo nextest run --offline --locked --test claim_link_cli --test claim_verify_cli --test hook_dispatch_cli`
- AC-12 (REQ-6..REQ-12): 조각 2 의 변경 경로는 이 SPEC 의 Scope write 목록 안이에요. · verify: `git diff --stat <base>..<head>`
- AC-13 (REQ-12): 클레임 연결 테스트 도우미는 HOME 없이 클론 신원을 풀어요. `tests/claim_link_cli.rs` 의 `record()` 가 `identity_in(&self.root, None)` 을 부르고, 되돌리면(`identity(&self.root)`) 기준선에서 지운 스물다섯 테스트가 enforce 에서 실패해요. · verify: `git diff <base>..<head> -- tests/claim_link_cli.rs` 와 `cargo nextest run --offline --locked --test claim_link_cli`

## Scope

- write(조각 2): `src/layout.rs`, `src/claim.rs`, `src/worktree.rs`, `src/identity.rs`, 클레임 장부를 프로세스 안에서 부르는 기준선 테스트 파일(`tests/claim_link_cli.rs`, `tests/claim_verify_cli.rs`, `tests/hook_dispatch_cli.rs` 와 관찰 실행이 더 보여 주는 같은 종류 파일; 확정 목록은 Result), `tests/test_home_guard.rs`(상한 한 줄), `tests/fixtures/test_home_guard/baseline.tsv`, `.moai/specs/SPEC-20261124-home-env-readers/spec.md`
- 차례: `src/claim.rs`·`src/identity.rs` 는 identity 를 HOME 인자로 바꾸는 카드의 조각 2a 가 착지한 뒤(NeXT `5b6fcb4689`)에 이 카드가 예약을 넓혔어요(리더 결정 노트 7763). 기준선·`tools/repo_scans.json` 쓰기의 파일 차례는 지났어요.
- read-only: `src/home_root.rs`, `src/forge/outbox.rs`, `src/apply.rs`, `src/handoff.rs`, `src/scope.rs`
- forbidden: 실제 HOME(읽기 전후 비교로도 건드리지 않아요, 임시 HOME 픽스처만), `CLAUDE.md`, 네트워크, `src/` 안의 `identity::identity` → `identity_in` 호출 교체 줄(identity 를 HOME 인자로 바꾸는 카드의 몫). `tests/claim_link_cli.rs` 도우미 한 줄은 이 금지에서 빠져요(D-10).

## Plan

인벤토리(2026-11-24, `rg 'env::var(_os)?\("HOME"\)' src` 와 함수 이름 확인; 본문 분류는 함수 이름만 보았어요 — 조각마다 읽어 확인하고 증거로 남겨요, 그 전은 unverified):

| 구분 | 곳 | 조각 |
| --- | --- | --- |
| 가장자리(한 번 읽는 자리, 남겨요) | `src/main.rs` 여섯 곳, `src/ops/*::run` 열네 곳(`args.home` 이 없으면 환경), `src/ledger/command.rs::resolve_home`, `src/hook_entry.rs::env`, `src/statusline/mod.rs::run`, `src/identity.rs::identity`(C-409) | — |
| layout 안 | `src/layout.rs::default_home_root`(읽는 곳), `switched_dir`, `claims_legacy_dir`(`identity` 호출) | 1(이 판) |
| layout 호출부(장부) | `src/claim.rs::journal_root`(`switched_dir`), `src/worktree.rs::open_locked_journal`(`apply_at` 이 부르며 `claim::journal_root` 를 거쳐 환경을 읽어요) | 2 |
| layout 호출부(저장소) | `src/forge/outbox.rs:202`(`switched_dir`), `src/apply.rs:368`(`switched_dir`) | 관찰 실행 뒤 정해요(D-7) |
| git 자식 | `src/identity.rs::rev_parse`(환경 HOME 의 전역 git 설정을 읽음, C-409 N1) | 2 (D-4, D-8) |
| 안쪽 읽기 A | `src/key_vault.rs::home`, `src/home_workdir.rs::home`, `src/hooks.rs::env_home`, `src/hook_config.rs::load`, `src/hook_engine.rs::session_check`, `src/edit_diagnostics.rs::lock_dir`, `src/heavy_slot.rs::cache_dir` | 3 |
| 안쪽 읽기 B | `src/claude.rs::doctor`·`refuse_user_scope`, `src/project_update.rs::go_settings_step`·`settings_step`, `src/baseline.rs::inventory`, `src/migrate.rs::open_cli_home`, `src/update_tux.rs::view_for`, `src/ops/host_wake.rs::default_config_dir`, `src/ops/star_ask.rs::run` | 4 |
| 줄 지우기 | `tests/fixtures/test_home_guard/baseline.tsv` | C-406 착지 뒤 |

1조각 순서: 1) 승인 노트를 받아요. 2) RED: 위 테스트 이름으로 `default_home_root_in`·`switched_dir_in` 이 없어 컴파일이 실패하는 것을 봐요. 3) GREEN: `default_home_root_in`·`switched_dir_in` 을 더하고 `default_home_root`·`switched_dir` 는 어댑터로, `claims_legacy_dir` 는 `identity_in(project, None)` 으로 바꿔요. 4) 이웃: `layout` 모듈 기존 테스트와 `claim` 의 `journal_root` 테스트, fmt, clippy.

조각 2 순서(앞 단계가 끝나야 다음으로 가요):
1. 승인(판 2)을 리더 노트로 받아요.
2. git 전역 설정 실험(D-4): 소유자가 다른 저장소(`GIT_TEST_ASSUME_DIFFERENT_OWNER=1`)에서 아래 아홉 경우를 돌리고 결과를 Result 에 적어요 — HOME 만 다른 경우, `GIT_CONFIG_GLOBAL` 을 받은 HOME 의 `.gitconfig` 로 준 경우, `/dev/null` 로 준 경우, `GIT_CONFIG_NOSYSTEM` 을 더한 경우, `XDG_CONFIG_HOME` 이 있는 경우. 그 결과로 옵션 (a)·(b)·(c) 중 하나를 고르고 D-8 에 적어요.
3. `src/layout.rs`(C-409 와 겹치지 않아요): RED→GREEN 으로 N2(조각 1 테스트의 `git init` 자식 HOME)와 `apply_at` 의 `open_locked_journal_in(project, None)` 호출(AC-8, AC-10). `open_locked_journal_in` 이 먼저 있어야 하므로 이 단계 뒤쪽은 5 번과 함께 가요.
4. C-409 조각 2a 가 착지한 뒤: `moai card scope` 로 `src/claim.rs`·`src/identity.rs` 와 테스트 파일을 같은 generation 에 더해요(리더 승인 필요).
5. `src/claim.rs::journal_root_in` → `src/worktree.rs::open_locked_journal_in` → `src/identity.rs` 의 git 전역 설정 고정(AC-6, AC-7, AC-9).
6. 클레임 장부를 프로세스 안에서 부르는 기준선 테스트가 `journal_root_in(..., None)` 을 쓰게 바꿔요. 바꾼 파일마다 관찰 모드(`MOAI_HOME_GUARD=observe`)로 그 바이너리를 돌려 가드 줄이 사라진 테스트만 확인해요.
7. 기준선 줄 지우기와 상한 낮추기(AC-11)는 맨 마지막이에요. 줄을 지운 테스트는 `MOAI_HOME_GUARD=enforce` 로 통과해야 하고, 파일 차례(C-418 다음)를 기다려요.

## Risks

- 이 조각만으로는 실제 HOME 읽기가 줄었는지 못 재요: 호출부가 아직 `switched_dir`(어댑터)를 불러요. 카드 기준 1 은 부분 충족, 기준 2 는 not met 으로 적고, 줄어든 수는 C-406 가드(observe)가 착지한 뒤에 재요(그 전은 unverified, C-409 D-4 와 같아요).
- REQ-5 는 `plan` 이 홈 상태 폴더가 심볼릭 링크일 때 내던 거부를 클레임 원본 찾기에서 없애요(그 거부는 클레임 원본과 관계없는 칸이었어요). 동작이 바뀌는 곳은 이것 하나이고 D-3 에 적었어요.
- `apply_at` 의 저널 잠금(`open_locked_journal`)은 이 조각이 닿지 않는 `claim::journal_root` 를 거쳐 환경 HOME 을 읽어요. 조각 2 가 닫기 전까지 `apply_cli` 경로는 환경 HOME 을 읽는 채로 남아요.
- 인벤토리의 본문 분류는 함수 이름까지만 확인했어요. 조각마다 읽어서 고쳐요.
- 조각 2: `apply_at` 의 잠금이 `None`(프로젝트 쪽 저널)을 쓰면, 환경 HOME 과 다른 `home_root` 로 `apply_at` 을 부르는 호출에서 잠그는 저널이 달라질 수 있어요. 지금 그렇게 부르는 곳은 없어요(`apply_cli` 는 `default_home_root()` 를 넘겨요; `rg 'apply_at\(' src tests` 로 확인, 판 2 승인 뒤 한 번 더). `check_destinations` 가 이미 전환된 포인터를 잠금 앞에서 거절하므로(`src/layout.rs:274`) 잠금에 닿는 호출의 저널은 늘 프로젝트 쪽이에요.
- 조각 2: 기준선 줄이 몇 개 줄지는 변경 뒤 관찰 실행 전에는 몰라요(unverified). `journal_root_in` 으로 바꾸는 테스트 도우미는 `identity::identity`(C-409 몫)도 불러서, 두 카드가 모두 끝나야 그 테스트가 HOME 을 안 건드릴 수 있어요. 한쪽만 끝난 파일은 줄을 지우지 않아요.
- 조각 2: 실험(D-4)을 돌리려면 `HOME`·`GIT_CONFIG_GLOBAL` 을 정한 `git` 명령이 필요한데, 워크트리 격리 훅이 그런 명령을 막았어요(이 세션, 2026-11-23). 그 명령줄 실험은 다시 돌리지 않았고, 리더 결정(노트 7763)대로 자식 프로세스에만 HOME 을 주는 Rust 테스트로 D-8 의 동작을 RED 먼저 확인했어요(Result).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `default_home_root_in(home: Option<&Path>)` 와 `switched_dir_in(project, subdir, home: Option<&Path>)` 를 더하고 기존 두 함수는 환경 HOME 을 한 번 읽어 위임하는 어댑터로 남겨요 | C-409 D-1 과 같은 방식; 호출부 세 곳이 이 카드 범위 밖이라 서명을 바로 바꿀 수 없어요 | `_in` 두 함수를 지우고 본문을 되돌리기 |
| D-2 | 가장자리(`main.rs`, `ops/*::run`, `ledger::resolve_home`)의 읽기는 남겨요 | 카드 기준 1 은 "명령 가장자리 밖에서는 0" 이에요; 가장자리가 한 번 읽는 자리예요 | 해당 없음 |
| D-3 | `claims_legacy_dir` 는 `identity_in(project, None)` 을 불러요 | 그 함수는 `clone_anchor` 만 읽어요(`src/layout.rs:87`); 홈 상태 폴더 칸은 쓰지 않아요 | 이전 줄 `identity(project)` 로 되돌리기 |
| D-4 | git 자식의 전역 설정 고정(C-409 N1)은 조각 2 에서 `src/identity.rs` 범위를 받아 해요. 옵션: (a) `rev_parse` 자식에 `GIT_CONFIG_GLOBAL` 을 받은 HOME 의 `.gitconfig`(없으면 `/dev/null`)로 줘요 (b) `GIT_CONFIG_NOSYSTEM`·`GIT_CONFIG_GLOBAL=/dev/null` 로 둘 다 막아요 (c) 고치지 않고 문서로만 남겨요. 기본 (a) — 단, `safe.directory` 가 전역 설정에서 오므로 (b) 는 소유자가 다른 저장소를 거부할 수 있어요(미확인, 조각 2 에서 git 문서와 실험으로 확인) | 판 1 에서 정하지 않았고, 판 2 의 D-8 이 이어 받아요 | 해당 없음 |
| D-5 | `claim::journal_root_in(project, home: Option<&Path>)` 는 `identity_in(project, None)` 으로 클레임 원본을 찾고 `switched_dir_in(project, "claims", home)` 으로 전환 폴더를 봐요. `journal_root` 는 환경 HOME 을 한 번 읽어 넘기는 어댑터예요. `worktree::open_locked_journal_in` 도 같은 꼴이에요 | D-1 과 같은 방식; 호출부가 공개 함수 서명에 의존하는 곳이 많아서(`rg '[^_a-z]journal_root\(' src tests` 가 클레임 장부 함수 호출 스무 곳을 보여요) 서명을 바로 바꾸지 않아요 | `_in` 함수를 지우고 어댑터 본문을 되돌리기 |
| D-6 | `apply_at` 은 `open_locked_journal_in(project, None)` 을 불러요 | 잠금 앞의 `check_destinations` 가 이미 전환된 포인터를 거절하므로 잠금 시점의 저널은 프로젝트 쪽이에요(읽어서 확인: `src/layout.rs:274`) | 이전 줄 `open_locked_journal(project)` 로 되돌리기 |
| D-7 | `src/forge/outbox.rs`·`src/apply.rs` 의 `switched_dir` 호출은 조각 2 의 6단계 관찰 실행 결과를 보고 정해요: 그 뒤에도 기준선에 남는 줄이 이 두 모듈 때문이면 판 3 으로 `_in` 을 더하고, 아니면 조각 3 으로 미뤄요 | 기준선 바이너리 중 `tests/forge_report_cli.rs`(5줄)가 `outbox::load` 를, `tests/project_ci_plan_lib.rs`(9줄)가 `apply::recover` 를 프로세스 안에서 불러요(읽어서 확인: `rg 'outbox::|apply::recover' tests/<바이너리>.rs`). 그 줄이 이 두 호출 때문인지 `identity` 때문인지는 아직 못 쟀어요(unverified); 소비자가 확인되기 전에 `_in` 을 더하지 않아요(핵심 규칙 20) | 해당 없음 |
| D-8 | git 전역 설정 고정의 기본은 (a)이고, `home` 이 `None` 이면 자식의 전역 설정 찾기를 건드리지 않아요(프로세스 환경 그대로). 실험 결과가 (a) 와 다르면 이 줄을 고쳐 다시 승인받아요 | (b) 는 `safe.directory` 를 막을 수 있어 실험 전에는 고르지 않기로 한 리더 결정(노트 7680); `None` 을 `/dev/null` 로 막으면 `identity_in(path, None)` 을 부르는 명령 경로의 사용자가 자기 `safe.directory` 설정을 잃어요 | `rev_parse` 의 환경 줄 삭제 |
| D-9 | 기준선 줄은 마지막 단계에서 관찰 실행 결과로만 지워요; 상한(`BASELINE_CAP`)은 남은 줄 수로 같은 변경에서 낮춰요 | 줄 수는 "이 변경 뒤 가드가 건드렸다고 말하지 않는 테스트"만 줄일 수 있어요(핵심 규칙 10: 글자 일치는 가설) | 지운 줄을 되살리기 |
| D-10 | `tests/claim_link_cli.rs` 도우미 `record()` 의 `identity::identity(&self.root)` 한 줄을 `identity_in(&self.root, None)` 으로 바꿔요. 금지 문장은 `src/` 안의 교체 줄로 좁혀요 | 리더 결정 노트 7830(7828 에 답함; 리더 판단, 사용자 위임). 그 도우미가 읽는 칸은 HOME 과 무관하고, 관찰 실행에서 그 한 줄이 스물다섯 테스트를 기준선에서 빼요(`guard-after2.log`). identity 를 HOME 인자로 바꾸는 카드의 범위에는 `tests/` 가 없어 부딪치는 카드가 없어요 | 그 한 줄을 `identity(&self.root)` 로 되돌리고 지운 스물다섯 줄을 되살리기 |

## Interview

- Q1: 서명을 바로 바꾸나요? → A: 아니요(D-1). `switched_dir` 호출부 세 곳이 범위 밖이에요.
- Q2: 환경 HOME 읽기를 전부 이 카드 하나로 끝내나요? → A: 아니요. 인벤토리를 조각으로 나눠요; 조각 2 이후는 시작할 때 리더 노트로 범위를 받아요.
- Q3: 조각 2 가 C-409 조각 2a 와 같은 파일(`src/claim.rs`, `src/identity.rs`)을 만지는데 어떻게 하나요? → A: C-409 가 먼저예요(카드 본문의 순서). 이 카드는 겹치지 않는 부분(`src/layout.rs`)을 먼저 하고, 겹치는 파일은 C-409 착지 뒤 `moai card scope` 로 예약을 넓혀 해요.
- Q4: `journal_root` 호출부(서른 곳 넘음)를 모두 `journal_root_in` 으로 옮기나요? → A: 아니요. 어댑터를 남기고 기준선 테스트가 부르는 곳과 `apply_at` 만 옮겨요. 나머지 호출부는 가장자리에서 HOME 을 받아 내려오는 조각에서 옮겨요.

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

**조각 2 결과** (브랜치 `spec/20261124-home-env-readers`, 기준 `29e15b53cc` 를 쫓지 않았어요; 구현 커밋 `5cfe929a58`(N2), `da9fda941c`(단계 5), `09947f9e71`·`98fa1b9636`(단계 6·7); 승인은 판 2 가 리더 노트 7763, 판 3 이 리더 노트 7830 이고 둘 다 리더 판단이며 사용자 직접 선택은 아니에요; 로그는 `target/c410/` 에 있고 카드가 열려 있는 동안만 둬요)

- 단계 5 RED: 새 함수가 없어 컴파일이 exit 101(오류 12개). 함수를 더한 뒤 `cargo nextest run --offline --locked --lib -E 'test(journal_root_in) | test(open_locked_journal) | test(apply_at_locks) | test(identity_in_git_config)'` 가 exit 100, 6개 중 4 통과·2 실패(`apply_at` 이 환경 HOME 쪽 저널을 잠가 busy 대신 stale plan 이 나오고, `identity_in` 이 환경 HOME 의 `safe.directory` 를 따라요). GREEN: 같은 명령 exit 0, 6/6, `--lib` 전체 1786/1786.
- AC-6: 충족. `journal_root_in_follows_the_home_it_is_given_not_the_environment`(자식 둘: 환경 HOME 에 포인터가 없을 때와 있을 때, 어댑터 `journal_root` 는 환경 HOME 의 답과 같고 `journal_root_in(None)` 은 클론 쪽), `journal_root_in_refuses_a_linked_given_home_when_switched`(링크된 홈 상태 폴더 아래 포인터는 같은 문장으로 거절하고 링크 너머에 `claims` 가 안 생겨요).
- AC-7: 충족. `open_locked_journal_in_locks_the_journal_of_the_home_it_is_given`(두 번째 잠금은 `unavailable` busy, 클론 쪽 저널은 따로 잠겨요), `open_locked_journal_locks_the_journal_of_the_environment_home`(어댑터가 환경 HOME 의 저널을 잠가요).
- AC-8: 충족. `apply_at_locks_the_project_side_journal_without_reading_the_environment_home`(환경 HOME 이 전환된 자식에서도 프로젝트 쪽 저널을 잠가 busy 를 답하고 환경 쪽 저널을 열지 않아요; 이미 전환된 홈 뿌리는 `already switched` 로 거절). `rg 'apply_at\(' src tests` 의 호출은 `apply_cli`(`default_home_root()` 를 넘김)와 `layout` 안 테스트뿐이에요.
- AC-9: 충족. `identity_in_git_config_follows_the_home_it_is_given`: 소유자가 다른 저장소(`GIT_TEST_ASSUME_DIFFERENT_OWNER`, 자식 프로세스만)에서 받은 HOME 의 `.gitconfig` 에 `safe.directory` 가 있으면 성공하고 없으면 `dubious ownership` 으로 거절해요. 환경 HOME 이 어느 쪽이든 같고 `None` 은 환경 HOME 을 따라요(통제).
- D-8 결과: 옵션 (a) 를 확인했어요. git 문서(`man git`, git 2.54.0, ENVIRONMENT): "if GIT_CONFIG_GLOBAL is set, neither $HOME/.gitconfig nor $XDG_CONFIG_HOME/git/config will be read", `/dev/null` 은 해당 층을 건너뛰어요. (b) 는 같은 테스트의 "받은 HOME 이 신뢰함" 경우를 깨므로 고르지 않아요. 훅이 막은 명령줄 실험은 돌리지 않았고 위 Rust 테스트가 그 동작을 대신 확인해요(`not_run`: 명령줄 아홉 경우 실험).
- AC-10: 충족. 단계 3(`5cfe929a58`)의 `init_repo_ignores_the_global_git_config_of_the_environment_home`.
- D-7 결과: 미해결(unverified). `tests/forge_report_cli.rs`·`tests/project_ci_plan_lib.rs` 의 기준선 줄이 `outbox`·`apply` 때문인지 이 조각에서 재지 않았어요. 이 조각은 두 모듈을 건드리지 않았고 줄도 그대로예요 → 다음 조각 몫.
- AC-11: 부분 충족. 관찰 실행(`MOAI_HOME_GUARD=observe`, `claim_link_cli`·`claim_verify_cli`·`hook_dispatch_cli` 한 명령): 변경 전 41개(27+12+2, 기준선 줄과 같음) → 테스트 도우미를 `journal_root_in(.., None)` 으로 바꾼 뒤 38개 → `record()` 한 줄까지 바꾼 뒤 13개(`claim_verify_cli` 12 + `claim_link_cli` 1). 빠진 28개 줄만 지우고 `BASELINE_CAP` 을 389 에서 361 로 낮췄어요. 지운 줄의 바이너리는 enforce 에서 통과했어요(`test_home_guard`·`claim_link_cli`·`hook_dispatch_cli` 엔 가드 줄 없음). 남은 `claim_verify_cli` 12개는 `src/` 안의 어댑터(`journal_root`, `identity::identity`)를 거치는 읽기라 `src/` 호출을 옮기는 다음 조각 몫이에요.
- AC-12: 변경 경로는 `src/claim.rs`, `src/identity.rs`, `src/layout.rs`, `src/worktree.rs`, `tests/claim_link_cli.rs`, `tests/claim_verify_cli.rs`, `tests/hook_dispatch_cli.rs`, `tests/test_home_guard.rs`, `tests/fixtures/test_home_guard/baseline.tsv` 와 이 SPEC 이에요(`git diff --stat` 은 READY 노트에 적어요).
- AC-13: 충족. `record()` 를 `identity(&self.root)` 로 되돌리는 뮤턴트가 `claim_link_cli` 를 enforce 에서 exit 100(28개 중 25개 실패, `HOME-GUARD: touched` 줄 25개)으로 깨요.
- 뮤턴트(적용을 `git diff --stat` 으로 확인하고 되돌린 뒤 `git diff --stat` 이 비어 있음을 확인; 모두 위 `-E` 필터, exit 100): ① `journal_root_in` 이 받은 HOME 대신 환경 HOME 을 씀(5개 실패) ② `journal_root_in` 의 링크 거절 삭제(1개) ③ `open_locked_journal_in` 이 `journal_root` 를 씀(3개) ④ `open_locked_journal` 어댑터가 `None` 을 넘김(1개) ⑤ `apply_at` 이 `open_locked_journal` 로 돌아감(1개) ⑥ `rev_parse` 의 `GIT_CONFIG_GLOBAL` 고정이 절대 경로 HOME 을 건너뜀(1개).
- 카드 기준 1: 부분 충족. 클레임 저널·`apply_at`·`identity_in` 의 git 자식은 환경 HOME 을 읽지 않고 읽는 곳은 `layout::environment_home` 어댑터 경로로 모였어요. 안쪽 읽기 A·B 모듈, `forge/outbox.rs`·`apply.rs` 의 `switched_dir`, `claim.rs` 안 `journal_root` 어댑터 호출이 남아요.
- 카드 기준 2: 충족. 기준선 줄이 389 에서 361 로 줄었고 줄어든 테스트가 enforce 에서 통과해요(위 AC-11).
- 합치기 주의: 이 브랜치는 조각 2a 가 착지하기 전 기준에서 이어졌어요. 조각 2a 가 같은 `claim::journal_root` 본문(`identity_in(.., None)` 과 링크 거절)과 `refuse_symlink` 가시성 한 줄을 바꿔서, 착지 때 `src/claim.rs` 의 `journal_root` 한 곳에서 합치기 충돌이 나요. 합친 결과는 이 브랜치의 `journal_root_in` 이 그 둘을 모두 가진 모양(받은 HOME 으로 `switched_dir_in`·`default_home_root_in` 을 부르고 링크 거절)이고 `journal_root` 는 그것을 부르는 어댑터예요.
- not_run: `cargo test --locked` 전체(리더 몫), Windows·Linux, 명령줄 git 실험, 실제 HOME 읽기 전후 비교, `forge/outbox.rs`·`apply.rs` 원인 재기.
