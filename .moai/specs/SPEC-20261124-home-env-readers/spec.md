---
id: SPEC-20261124-home-env-readers
title: "환경 HOME 을 직접 읽는 곳을 명령 가장자리로 모아요: layout, 클레임 장부, 읽기 목록 검사와 남은 묶음"
status: draft
revision: 6
created: 2026-11-24
---

# 환경 HOME 을 직접 읽는 곳을 명령 가장자리로 모아요: layout, 클레임 장부, 읽기 목록 검사와 남은 묶음

## Intent

카드 C-410 은 C-409(identity 가 HOME 을 인자로 받음)의 둘째 갈래예요. C-409 노트 7613 의 D-3 대로, `layout::default_home_root`(`src/layout.rs:34`)가 환경 변수 HOME 을 읽고, 클레임 장부 경로(`claim::journal_root` → `layout::switched_dir`)가 그것을 거쳐요. identity 만 고치면 클레임 묶음 테스트(`claim_link_cli` 등)가 실제 HOME 을 읽는 일이 다 사라지지 않아요. 같은 종류(C-406 원인 + C-105 감사 비차단 1)가 두 번이라 소스를 고치는 카드예요.

방식은 C-409 와 같아요(그 SPEC 의 D-1): HOME 을 인자로 받는 `_in` 짝을 더하고, 기존 이름은 환경 HOME 을 한 번 읽어 위임하는 가장자리 어댑터로 남겨요. 이 카드의 쓰기 범위는 `src/layout.rs` 와 이 SPEC 폴더예요. 그래서 1조각은 layout 안만 고치고, 호출부와 다른 읽는 곳은 조각 2 이후에 리더 노트로 범위를 받아 이어가요.

C-409 N1(감사 #409)도 이 카드 몫이에요: `identity_in` 이 부르는 git 자식은 프로세스 환경을 그대로 받아 환경 HOME 의 전역 git 설정을 읽어요. 그 고정은 `src/identity.rs` 안이라 1조각의 범위 밖이었고, 조각 2 에서 해요(D-4, D-8).

판 2(조각 2)는 세 가지를 해요. ① 클레임 장부 호출부가 HOME 을 받아요: `claim::journal_root_in`, `worktree::open_locked_journal_in` 을 더하고 `layout::apply_at` 의 저널 잠금이 환경 HOME 을 읽지 않게 해요. ② git 자식의 전역 설정을 받은 HOME 으로 고정하고, 조각 1 테스트의 부모 쪽 git 자식이 임시 폴더 HOME 을 쓰게 해요(감사 #410 N2). ③ C-406 가드가 착지했으니(NeXT `132728a026`) 이 변경으로 더는 HOME 을 건드리지 않게 된 테스트의 기준선 줄을 같은 변경에서 지워요. 줄이 몇 개 줄어드는지는 변경 뒤 관찰 실행으로 재요(그 전은 unverified).

판 3(조각 2 의 마무리)은 하나를 더해요. 클레임 연결 테스트(`tests/claim_link_cli.rs`)의 도우미 `record()` 가 `identity::identity` 를 불러 그 테스트 스물여섯 개가 가드 기준선에 남아 있었어요. 그 도우미가 읽는 칸(`clone_anchor`, `workspace_git_dir`, `work_tree`)은 HOME 과 무관하므로 `identity_in(.., None)` 으로 바꾸고, 관찰 실행으로 가드에서 빠진 것을 확인한 스물다섯 줄을 같은 변경에서 지워요(리더 결정 노트 7830).

판 4(조각 2 수리, 감사 #410 막는 것 B1, 리더 결정 노트 7851)는 판 2 ② 의 git 자식 방식을 바꿔요. 판 2 는 받은 HOME 이 있으면 `GIT_CONFIG_GLOBAL` 을 그 HOME 의 `.gitconfig`(없으면 `/dev/null`)로 고정했어요. git 은 `GIT_CONFIG_GLOBAL` 이 있으면 `$XDG_CONFIG_HOME/git/config` 를 읽지 않아서(git 2.54.0, git(1) ENVIRONMENT), `identity()` 가 XDG 파일에 `safe.directory` 를 둔 사용자와 자기 `GIT_CONFIG_GLOBAL` 을 둔 사용자의 설정을 버렸어요. 판 4 는 그 고정을 없애요: 받은 HOME 이 프로세스의 환경 HOME 과 같거나 HOME 이 없으면 자식은 환경을 그대로 받고, 다르면 자식의 HOME 을 받은 HOME 으로 두고 `XDG_CONFIG_HOME` 과 `GIT_CONFIG_GLOBAL` 을 지워요. 같은 판에서 감사 #410 비차단 N2(Scope 의 금지 문장과 D-5 의 어긋남)와 N3(`spec show` 가 Scope 를 잘못 읽음)의 문구를 고쳐요.

판 5(조각 3)는 카드 기준 1 의 남은 몫을 작은 묶음으로 나눠요. 이 판이 요구로 묶는 것은 첫 묶음 하나예요: ① 환경 HOME 을 읽는 곳과 어댑터 호출 자리의 목록을 테스트로 만들어 소스와 어긋나면 실패하게 해요. ② 잡힌 파일이 없는 두 곳(`baseline::inventory`, `host-wake` 의 설정 폴더 기본값)에서 읽기를 명령 가장자리로 올려요. 남은 묶음은 Plan 의 표에만 두고, 파일을 잡은 카드가 놓는 차례에 다음 판에서 요구로 올려요.

판 6(리더 결정 노트 10217)은 ① 의 검사를 관찰 모드(observe)로 시작하게 바꿔요. 새 관문은 관찰부터라는 사용자 규칙(2026-11-20, `skills/moai/references/internal/sdd.md` 11 항)에 따라, 목록과 소스가 달라도 알려 주기만 하고 착지를 막지 않아요. 판단 기준은 "정확히 같아야" 그대로이고, 막기(enforce)로 바꾸는 일은 측정한 뒤 사용자에게 물어요.

## Non-goals

- Plan 의 B2 이후 묶음 — 이 판의 요구가 아니에요. 다음 판에서 요구로 올려요.
- 명령 가장자리(`src/main.rs`, `src/ops/*::run`)의 읽기를 없애는 일. 가장자리는 한 번 읽는 자리라 남겨요.
- `identity::identity` 호출을 `identity_in` 으로 바꾸는 일 — C-409 의 몫이에요. 이 카드는 `journal_root_in` 이 필요한 곳만 바꿔요.
- 감사 #410 비차단 N1(`journal_root_in` 의 git 자식이 환경 HOME 설정을 읽음)과 R2(같은 HOME 을 다른 철자로 주면 사용자 git 설정을 잃음)의 수리 — 어느 카드가 받을지는 리더가 정해요.

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
- REQ-10: The `identity_in` function shall run its git children with the HOME it is given as their HOME and without `XDG_CONFIG_HOME` and `GIT_CONFIG_GLOBAL` when that HOME is not the process's own HOME, and with the process's own environment when it is that HOME or none.
- REQ-11: The layout tests shall run every git child they start with a HOME in a temporary folder.
- REQ-12: The baseline of `tests/fixtures/test_home_guard/baseline.tsv` shall list no test that the observation run of the changed tree shows leaving its HOME untouched.
- REQ-13: The `home_env_readers` comparison shall report one line, naming the file and the function, for every product source expression under `src/` that reads `HOME` through `env::var` or `env::var_os` whose file, enclosing function, count or class (edge or inner) differs from the list.
- REQ-14: The `home_env_readers` comparison shall report one line for every call site of a HOME adapter outside a command edge that is added, removed or moved to another function without the same change in the list.
- REQ-15: The `baseline::inventory` function shall report each HOME entry from the HOME its caller hands over, `null` when none is handed over, without reading the environment HOME.
- REQ-16: The `host-wake` operation `run` shall be the only place in its module that reads `HOME`, handing it to `default_config_dir` with the precedence of `CLAUDE_CONFIG_DIR` and the relative `.claude` for an unset or empty HOME unchanged.
- REQ-17 (REQ-13, REQ-14): IF `MOAI_HOME_READERS` is `enforce` and the comparison reports a line, THEN the `home_env_readers` test of the checked-in sources shall fail with that line.
- REQ-18: WHILE `MOAI_HOME_READERS` is not `enforce`, the `home_env_readers` test of the checked-in sources shall pass whatever the comparison reports, writing each line with the prefix `home-env-readers observe:` to the output that `cargo nextest run` and `cargo test` show for a passing run.

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
- AC-14 (REQ-10): 소유자가 다른 저장소에서, 신뢰(`safe.directory`)를 받은 HOME 의 `.config/git/config` 에만 둔 HOME 을 넘기면 `identity_in(path, Some(home))` 이 성공하고, 신뢰가 없는 HOME 은 거절 문구로 실패해요(판 2 의 `GIT_CONFIG_GLOBAL` 고정은 이 파일을 버렸어요). · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_in_reads_the_xdg)'`
- AC-15 (REQ-10): 환경이 `$XDG_CONFIG_HOME/git/config` 나 `GIT_CONFIG_GLOBAL` 이 가리키는 파일로만 신뢰를 두면 환경 HOME 과 같은 HOME 을 넘긴 `identity_in` 이 성공하고, 그 둘이 없으면 거절 문구로 실패해요; 어댑터 `identity(path)` 가 있는 동안은 그것도 같아요. · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_keeps_the)'`
- AC-16 (REQ-10): 환경이 `.gitconfig`·`$XDG_CONFIG_HOME/git/config`·`GIT_CONFIG_GLOBAL` 로 모두 신뢰해도, 환경 HOME 과 다른 HOME 을 넘기고 그 HOME 이 신뢰를 두지 않으면 `identity_in` 이 거절 문구로 실패해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_in_leaves_the)'`
- AC-17 (REQ-13): 목록과 같은 가짜 소스에서는 보고 줄이 없고, 목록에 없는 읽기를 더한 가짜 소스에서는 파일과 함수 이름을 담은 줄이 나와요. 주석 줄과 첫 칸의 `#[cfg(test)]` 모듈은 세지 않아요. 가짜 소스로 재는 테스트는 모드와 상관없이 늘 단언해요. · verify: `cargo nextest run --offline --locked --test home_env_readers`
- AC-18 (REQ-14): 어댑터(`key_vault::home`, `home_workdir::home`, `hooks::env_home`, `layout::environment_home`, `default_home_root`, `switched_dir`, `switched_dir_checked`, `claim::journal_root`, `worktree::open_locked_journal`)의 호출 자리를 더하거나 빼거나 다른 함수로 옮긴 가짜 소스에서 그 줄이 나오고, 검사가 스캔 표에 `always` 로 등록돼요. 별칭으로 들여온 맨 `home()` 호출과 상수에 담은 이름은 못 보는 것이 한계예요. · verify: `cargo nextest run --offline --locked --test home_env_readers` 와 `cargo nextest run --offline --locked --test repo_scans_table`
- AC-19 (REQ-15): HOME 이 설정·빈 값·없음 세 경우에 `baseline-inventory` 의 HOME 항목 답이 바뀌기 전과 같아요(옛 코드에서 먼저 통과하는 특성 테스트). · verify: `cargo nextest run --offline --locked --test baseline_git_env -E 'test(baseline_inventory_reports_home_entries)'`
- AC-20 (REQ-16): `CLAUDE_CONFIG_DIR` 이 있을 때와 없을 때, HOME 이 설정·빈 값·없음일 때 `host-wake` 가 쓰는 설정 폴더가 바뀌기 전과 같아요(자식 프로세스 환경만 바꾸는 특성 테스트). · verify: `cargo nextest run --offline --locked --test host_wake_cli -E 'test(host_wake_default_config_dir)'`
- AC-21 (REQ-13, REQ-15, REQ-16): `baseline::inventory`·`default_config_dir` 의 읽기가 안쪽 목록에서 빠지고 두 명령 `run` 이 가장자리 목록에 들어간 목록이 통과해요(`b3048a821a` 에 이 묶음만 더하면 가장자리 30, 안쪽 17; 먼저 착지한 카드가 읽기를 바꿨으면 그 tip 의 수에 맞춰요). `MOAI_HOME_READERS=enforce` 로 통과하고, 목록을 먼저 고치면 옛 코드에서 enforce 로 실패해요(RED). · verify: `MOAI_HOME_READERS=enforce cargo nextest run --offline --locked --test home_env_readers`
- AC-22 (REQ-17): 차이가 있는 가짜 소스를 enforce 로 판정하면 그 줄을 담은 실패가 나와요. · verify: `cargo nextest run --offline --locked --test home_env_readers -E 'test(enforce_mode_fails_on_a_difference)'`
- AC-23 (REQ-18): 차이가 있는 가짜 소스를 enforce 가 아닌 값(설정 안 함, 빈 값, `observe`)으로 판정하면 통과하고, 줄마다 `home-env-readers observe:` 로 시작해요. · verify: `cargo nextest run --offline --locked --test home_env_readers -E 'test(observe_mode_reports_and_passes)'`
- AC-24 (REQ-18): 목록 한 줄을 일부러 어긋나게 둔 채 `cargo nextest run --offline --locked --test home_env_readers` 를 돌리면 exit 0 이고 그 출력에 `home-env-readers observe:` 줄이 보이며, `cargo test` 로 돌려도 보여요. 실행 시험 목록(`tests/spawn_cap_test.rs`)은 그대로 통과해요. · manual: 두 실행의 명령·exit·보인 줄을 Result 에 적고 어긋낸 줄은 되돌려요

## Scope

- write: `src/baseline.rs`, `src/ops/baseline_inventory.rs`, `src/ops/host_wake.rs`, `tests/home_env_readers.rs`, `tests/baseline_git_env.rs`, `tests/host_wake_cli.rs`, `tools/repo_scans.json`, `.config/nextest.toml`, `.moai/specs/SPEC-20261124-home-env-readers/spec.md`
- read-only: `src/home_root.rs`, `src/layout.rs`, `src/claim.rs`, `src/worktree.rs`, `src/identity.rs`, `src/key_vault.rs`, `src/main.rs`
- forbidden: `CLAUDE.md`

판 5·6 의 쓰기 칸은 첫 묶음의 파일이에요. `.config/nextest.toml` 은 이 검사의 통과 출력을 보이게 하는 `overrides` 한 묶음만 더해요(첫 `filter =` 줄보다 뒤에; `tests/spawn_cap_test.rs` 가 첫 줄을 읽어요). `tools/repo_scans.json` 은 공유 파일이라 예약 범위에 넣지 않고 `card line-lease` 로 한 줄만 더해요. `src/key_vault.rs`·`src/identity.rs`·`src/layout.rs` 는 identity 를 HOME 인자로 바꾸는 카드와 나눠 쓰기로 해서 이 묶음은 읽기만 해요. 앞 조각의 쓰기 칸은 각 조각의 Result 에 남아 있어요.

조각 2 의 쓰기 칸에서 `tests/claim_link_cli.rs`, `tests/claim_verify_cli.rs`, `tests/hook_dispatch_cli.rs` 는 클레임 장부를 프로세스 안에서 부르는 기준선 테스트 파일이고, 관찰 실행이 같은 종류 파일을 더 보여 주면 확정 목록은 Result 에 적어요. `tests/test_home_guard.rs` 는 상한 한 줄만 바꿔요. `src/claim.rs`·`src/identity.rs` 는 identity 를 HOME 인자로 바꾸는 카드의 조각 2a 가 착지한 뒤(NeXT `5b6fcb4689`)에 이 카드가 예약을 넓혔어요(리더 결정 노트 7763). 기준선·`tools/repo_scans.json` 쓰기의 파일 차례는 지났어요.

금지(위 `forbidden:` 줄의 나머지): 실제 HOME 은 읽기 전후 비교로도 건드리지 않아요(임시 HOME 픽스처만). 네트워크도 쓰지 않아요. `src/` 안에 이미 있는 `identity::identity` 호출 줄을 `identity_in` 으로 바꾸는 일은 identity 를 HOME 인자로 바꾸는 카드의 몫이라 하지 않아요. 새 함수 `journal_root_in`·`open_locked_journal_in` 이 처음부터 `identity_in(project, None)` 을 부르는 줄은 교체가 아니라 새 줄이라 이 금지에 걸리지 않아요(D-5). `tests/claim_link_cli.rs` 도우미 한 줄은 이 금지에서 빠져요(D-10).

## Plan

인벤토리(판 5, 기준 `b3048a821a`; `rg` 로 `env::var(_os)?("HOME")` 를 찾고 감싸는 함수를 읽었어요. 글자 검색이라 상수에 담은 이름으로 읽는 곳이 있으면 못 봐요). 제품 코드의 직접 읽기는 47곳이에요: 가장자리 28, 안쪽 19. 테스트 안의 3곳은 세지 않아요.

| 구분 | 곳 |
| --- | --- |
| 가장자리 28 | `src/main.rs` 6, `src/ops/*::run` 18(`star_ask::run` 포함), `src/ledger/command.rs::resolve_home`, `src/hook_entry.rs::env`, `src/statusline/mod.rs::run`, `src/identity.rs::identity`(identity 를 HOME 인자로 바꾸는 카드의 몫) |
| 안쪽 19 | `src/agents.rs` `tilde`·`default_global_excludes`, `src/baseline.rs::inventory`, `src/claude.rs` `doctor`·`refuse_user_scope`, `src/edit_diagnostics.rs::lock_dir`, `src/heavy_slot.rs::cache_dir`, `src/home_workdir.rs::home`, `src/hook_config.rs::load`, `src/hook_engine.rs` `session_check`·`bypass_boundary_deny`, `src/hooks.rs::env_home`, `src/key_vault.rs::home`, `src/layout.rs::environment_home`, `src/migrate.rs::open_cli_home`, `src/ops/host_wake.rs::default_config_dir`, `src/project_update.rs` `go_settings_step`·`settings_step`, `src/update_tux.rs::view_for` |
| 간접 읽기 | `src/host_home.rs` 의 `key_vault::home()` 대체 길, `src/statusline/state.rs` 의 환경 통째 담기(가장자리에서 받아요). 그 밖의 클로저는 `CLAUDE_CONFIG_DIR`·`CODEX_HOME` 같은 다른 이름만 읽어요 |
| 어댑터 호출 | `key_vault::home` 14(가장자리 밖 7), `home_workdir::home` 12(6), `hooks::env_home` 6(5), `layout::environment_home` 6(6), `claim::journal_root` 7(7), `worktree::open_locked_journal` 5(5), `default_home_root` 2(2), `switched_dir_checked` 2(2) |

묶음(판 5). 요구로 묶는 것은 첫 묶음(B0+B1)이고, 나머지는 파일을 잡은 카드가 놓는 차례에 다음 판에서 올려요. 파일을 잡은 카드는 2026-11-26 에 원장에서 읽은 값이라 묶음마다 다시 읽어요.

| 묶음 | 내용 | 기다리는 것 |
| --- | --- | --- |
| B0+B1 | 목록·호출 자리 검사(`tests/home_env_readers.rs`)와 `baseline::inventory`, `host-wake` 설정 폴더 기본값 | 없음 |
| B2 | `migrate::open_cli_home` 의 HOME 거절을 받은 사용자 HOME 으로(거절은 약해지지 않아야 해요) | `src/main.rs` |
| B3 | 훅 분배: 세션 확인, 권한 우회 경계 판정, 진단 잠금 폴더 | `src/main.rs`, `src/hook_engine.rs` 를 바꾼 카드의 착지 |
| B4 | 영수증 읽기·쓰기와 `hooks::env_home` 호출 자리 | B3, `src/hooks.rs`, `src/project_update.rs` |
| B5 | `claude::doctor` | `src/claude.rs` |
| B6 | `project_update::plan` 의 설정 단계 둘 | `src/project_update.rs` |
| B7 | `update_tux::view_for` | `src/main.rs`, `src/update.rs` |
| B8 | `hook_config::load` | B4 와 그 호출부 파일 |
| B9 | `refuse_user_scope`(안전 거절) | `src/claude.rs`, `src/hooks.rs` |
| B10 | `agents` 의 전역 제외 파일 두 곳 | `src/agents.rs`, `src/project_update.rs` |
| B11 | `key_vault::home` 의 안쪽 호출 자리 | `src/main.rs`, identity 를 HOME 인자로 바꾸는 카드의 `src/key_vault.rs` |
| B12 | `home_workdir::home` 의 안쪽 호출 자리(링크된 이동 뿌리 거절 유지) | `src/home_workdir.rs`·`src/worktree_move.rs` |
| B13 | layout·클레임 장부 어댑터의 호출 자리(링크 거절 유지) | `src/worktree.rs`, `src/main.rs` |
| B14 | `heavy_slot::cache_dir` | `src/heavy_slot.rs`, `src/check_affected.rs` |
| B15 | 안쪽 목록이 비고 이름 붙은 예외만 남는지 닫기 | 위 모두 |

기준선(`tests/fixtures/test_home_guard/baseline.tsv`) 줄은 관찰 실행으로 확인한 묶음에서만 지워요(D-9). 첫 묶음은 그 파일을 쓰지 않아요.

조각 3 첫 묶음 순서: 1) 승인 노트를 받아요. 2) 특성 테스트(AC-19, AC-20)를 옛 코드에서 통과시켜요. 3) 목록 검사를 옛 소스에 맞춰 enforce 로 통과시키고, 목록의 두 줄을 가장자리로 먼저 옮겨 enforce 실패(RED)를 봐요. 4) `inventory` 가 HOME 을 인자로 받고 `host_wake::run` 이 읽어 넘기게 바꿔요(GREEN). 5) 스캔 표 한 줄을 line-lease 로 더해요. 그 항목은 `"always": true` 와 `src/` 범위여야 해요: 목록이 `src/` 파일 경로를 글자로 적어서, 항상 스캔이 아니면 `every_file_reader_has_an_always_scan_or_an_exemption`(`tests/repo_scans_table.rs`)이 실패해요. 그래서 `src/` 의 `.rs` 를 바꾼 착지는 모두 이 검사를 돌리지만, 관찰 모드라 막지 않아요(D-13). `tests/repo_scans_table.rs` 는 고치지 않아요(다른 카드가 예약 중). `.config/nextest.toml` 에는 `binary(home_env_readers)` 의 `success-output = "immediate"` 를 첫 override 뒤에 더해요(D-15; 통과 출력은 기본으로 숨겨져요, https://nexte.st/docs/configuration/per-test-overrides/ 와 https://nexte.st/docs/reporting/ 를 2026-11-26 에 읽음; 통과 테스트에서 실제로 보이는지는 AC-24 로 확인). 그 파일을 바꾼 뒤에는 읽는 테스트 둘(`--test spawn_cap_test`, `--test test_home_guard`)을 이름으로 돌려요. 6) 새 테스트마다 뮤턴트, fmt, clippy, `check affected --card` 한 번.

관찰 창: 이 묶음의 착지부터예요. 착지·검사 로그의 `home-env-readers observe:` 줄을 세어, 이 검사가 막았을 횟수와 오탐을 Result 에 적어요. 막기로 바꾸는 일은 그 측정을 보여 주고 사용자에게 묻는 기록된 결정이에요(리더 결정 노트 10217).

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
- 판 6: 목록 검사는 관찰 모드라 다른 카드가 HOME 읽기나 어댑터 호출을 더하거나 옮겨도 착지돼요. 알림은 로그의 줄뿐이라 아무도 읽지 않으면 목록이 낡고 새 안쪽 읽기가 들어올 수 있어요. 관찰 창의 횟수는 로그를 모아야 나오고, 자동으로 세는 장치는 만들지 않아요.
- 판 5: 목록 검사는 글자 검색이에요. 별칭으로 들여온 맨 `home()` 호출, 상수에 담은 변수 이름, git 같은 자식 프로세스가 HOME 을 읽는 것은 못 봐요.
- 판 5: identity 를 HOME 인자로 바꾸는 카드의 승인 대기 초안(커밋 `3076d1543d`)은 어댑터 `identity` 를 지우고 `identity` 명령과 audit clone 가장자리의 환경 HOME 읽기를 바꿔요. 목록은 먼저 착지하는 쪽에 뒤 쪽이 맞춰요: 그 카드가 먼저면 이 묶음이 그 tip 에서 목록을 만들고, 이 묶음이 먼저면 그 카드가 `tests/home_env_readers.rs` 를 함께 고치기로 했어요(관찰 모드라 막지는 않아요).
- 판 5: `identity.rs` 가 일부러 `environment_home` 을 부르는 곳(D-8)과 `identity::identity` 는 이 카드만으로는 0 이 되지 않아요. 예외로 남길지는 B13·B15 에서 승인자가 정해요.
- 판 5: B12·B13 은 호출 사슬의 위쪽(`apply::receipts_dir`, `outbox::outbox_dir`, 워크트리 정리 쪽)을 아직 끝까지 따라가지 않았어요(unverified). 그 묶음을 요구로 올리기 전에 따라가요.

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
| D-8 | git 자식의 전역 설정은 `GIT_CONFIG_GLOBAL` 로 고정하지 않아요(판 4; 판 2 의 옵션 (a) 고정은 감사 #410 B1 이 막았어요). 받은 HOME 이 프로세스의 환경 HOME 과 같으면 자식은 환경을 그대로 받아 사용자의 `XDG_CONFIG_HOME`·`GIT_CONFIG_GLOBAL` 이 살아요. 다르면 자식의 HOME 을 받은 HOME 으로 두고 `XDG_CONFIG_HOME`·`GIT_CONFIG_GLOBAL` 을 지워 git 이 `<home>/.gitconfig` 와 `<home>/.config/git/config` 를 둘 다 읽어요. `home` 이 `None` 이면 건드리지 않아요(프로세스 환경 그대로) | git(1) ENVIRONMENT(git 2.54.0): `GIT_CONFIG_GLOBAL` 이 있으면 `$HOME/.gitconfig` 도 `$XDG_CONFIG_HOME/git/config` 도 읽지 않아요, 그래서 고정은 둘 중 하나를 버려요; (b) 는 `safe.directory` 를 막을 수 있어 고르지 않아요(리더 결정 노트 7680); `None` 을 `/dev/null` 로 막으면 `identity_in(path, None)` 을 부르는 명령 경로의 사용자가 자기 `safe.directory` 설정을 잃어요; 같은 HOME 의 판정은 `layout::environment_home` 한 곳을 거쳐요; 리더 결정 노트 7851 | `rev_parse` 의 환경 분기 삭제 |
| D-9 | 기준선 줄은 마지막 단계에서 관찰 실행 결과로만 지워요; 상한(`BASELINE_CAP`)은 남은 줄 수로 같은 변경에서 낮춰요 | 줄 수는 "이 변경 뒤 가드가 건드렸다고 말하지 않는 테스트"만 줄일 수 있어요(핵심 규칙 10: 글자 일치는 가설) | 지운 줄을 되살리기 |
| D-10 | `tests/claim_link_cli.rs` 도우미 `record()` 의 `identity::identity(&self.root)` 한 줄을 `identity_in(&self.root, None)` 으로 바꿔요. 금지 문장은 `src/` 안의 교체 줄로 좁혀요 | 리더 결정 노트 7830(7828 에 답함; 리더 판단, 사용자 위임). 그 도우미가 읽는 칸은 HOME 과 무관하고, 관찰 실행에서 그 한 줄이 스물다섯 테스트를 기준선에서 빼요(`guard-after2.log`). identity 를 HOME 인자로 바꾸는 카드의 범위에는 `tests/` 가 없어 부딪치는 카드가 없어요 | 그 한 줄을 `identity(&self.root)` 로 되돌리고 지운 스물다섯 줄을 되살리기 |
| D-11 | 여럿이 부르는 어댑터는 모든 호출 자리가 가장자리일 때만 가장자리로 쳐요. 그래서 목록 검사는 직접 읽기와 어댑터 호출 자리를 둘 다 적어요 | 어댑터를 늘 가장자리로 치면 안쪽 읽기를 도우미로 감싸기만 해도 카드 기준 1 이 통과해요. 어댑터를 모두 없애는 안은 `identity` 쪽 설계(D-8)와 부딪쳐요 | REQ-14 와 두 번째 목록을 지우기 |
| D-12 | 호출자가 모두 그 묶음 안에 있으면 읽는 어댑터를 남기지 않고 함수 서명을 바로 바꿔요(D-1 을 좁혀요) | `baseline::inventory` 와 `default_config_dir` 의 호출자는 각각 한 곳이에요(읽어서 확인) | 어댑터를 되살리기 |
| D-13 | 목록 검사의 판단은 "목록과 정확히 같아야"이고, 처음에는 관찰 모드로 둬요: 다르면 알려 주기만 해요. 테스트 코드 안의 읽기는 세지 않아요 | 줄어들기만 허용하면 새 안쪽 읽기가 숨어요. 목록은 줄 번호가 아니라 파일·함수·개수라 다른 편집에는 안 흔들려요. 관찰부터는 리더 결정 노트 10217 과 사용자 규칙 2026-11-20(`sdd.md` 11 항) | 검사를 "늘지 않음"으로 바꾸기; enforce 전환은 측정 뒤 사용자에게 물어요 |
| D-14 | 안전 거절(`migrate::open_cli_home`, `refuse_user_scope`, 링크 거절)에 넘기는 HOME 은 가장자리가 읽은 프로세스 HOME 이고 명령줄 `--home` 이 아니에요. 빈 HOME 같은 기존의 특이한 동작은 그대로 두고, 고치려면 따로 카드를 내요 | `--home` 을 받으면 거절을 비켜 갈 수 있어요. 이 카드는 읽는 자리를 옮기는 일이에요 | 해당 없음 |
| D-15 | 관찰 모드의 스위치는 환경 변수 `MOAI_HOME_READERS=enforce` 예요. 알림은 테스트가 프로세스 stderr 에 직접 쓰고(`cargo test` 에서 보이게), `cargo nextest run` 에는 `.config/nextest.toml` 의 `success-output` override 를 더해요. 검출기를 가짜 소스로 재는 테스트는 모드와 상관없이 단언해요 | 저장소의 앞선 방식(`MOAI_HOME_GUARD`, `tests/freshness_fixture.rs` 의 직접 쓰기)과 같아요. `tools/repo_scans.json` 에는 알림만 하는 항목 칸이 없어서 모드는 테스트 안에 둬요 | override 묶음을 지우고 라이브 비교를 늘 단언으로 바꾸기 |

- 판 4: 받은 HOME 이 환경 HOME 과 같은지는 `Path` 비교예요. 구성요소 비교라 끝 슬래시만 다른 경로는 같은 HOME 이고(감사 #410 R1 이 앞 문장을 바로잡았어요), 심볼릭 링크 경유나 대소문자만 다른 경로는 다른 HOME 으로 보아, 그때는 환경의 `XDG_CONFIG_HOME`·`GIT_CONFIG_GLOBAL` 이 자식에게 가지 않아요. `identity()` 는 환경 HOME 을 그대로 넘기므로 이 경우를 만나지 않아요(unverified: 다른 호출자가 그렇게 부르는 곳은 이 판에서 읽지 않았어요).
- 판 4: 자식에게 HOME 을 주는 방식은 유닉스 가정이에요(Windows 의 git 은 `USERPROFILE` 도 읽어요); Windows 결과는 not_run 이에요.

## Interview

- Q1: 서명을 바로 바꾸나요? → A: 아니요(D-1). `switched_dir` 호출부 세 곳이 범위 밖이에요. 판 5 는 호출자가 모두 그 묶음 안에 있는 함수만 서명을 바로 바꿔요(D-12).
- Q2: 환경 HOME 읽기를 전부 이 카드 하나로 끝내나요? → A: 아니요. 인벤토리를 조각으로 나눠요; 조각 2 이후는 시작할 때 리더 노트로 범위를 받아요.
- Q3: 조각 2 가 C-409 조각 2a 와 같은 파일(`src/claim.rs`, `src/identity.rs`)을 만지는데 어떻게 하나요? → A: C-409 가 먼저예요(카드 본문의 순서). 이 카드는 겹치지 않는 부분(`src/layout.rs`)을 먼저 하고, 겹치는 파일은 C-409 착지 뒤 `moai card scope` 로 예약을 넓혀 해요.
- Q4: `journal_root` 호출부(서른 곳 넘음)를 모두 `journal_root_in` 으로 옮기나요? → A: 아니요. 어댑터를 남기고 기준선 테스트가 부르는 곳과 `apply_at` 만 옮겨요. 나머지 호출부는 가장자리에서 HOME 을 받아 내려오는 조각에서 옮겨요.
- Q5: 판 5 는 남은 읽기를 한꺼번에 요구로 묶나요? → A: 아니요. 리더 지시가 "한 번에 한 묶음씩 작게"이고 많은 파일을 다른 카드가 잡고 있어요. 첫 묶음만 REQ-13~16 으로 묶고 나머지는 Plan 표에 둬요(요구가 열두 개를 넘어 `spec check` 경고가 나는지는 승인 요청 노트에 적어요).
- Q6: 새 목록 검사가 다른 카드의 착지를 막을 수 있는데 괜찮나요? → A: 처음에는 막지 않아요. 관찰 모드로 시작하고 판단 기준은 "정확히 같아야"예요(리더 결정 노트 10217, D-13).

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

**판 4 결과** (조각 2 수리; 감사 #410 FAIL 의 막는 것 B1; 승인은 리더 결정 노트 7858 이고 리더 판단이며 사용자 직접 선택은 아니에요; 수리 커밋 `b94c064c22`, 초안 커밋 `585f777a69`; 로그는 `target/c410/` 에 있고 카드가 열려 있는 동안만 둬요)

- RED: 새 테스트 셋을 더하고 수리 전에 `cargo nextest run --offline --locked --lib -E 'test(identity_in_) | test(identity_keeps_the)'` → exit 100, 8개 중 6 통과·2 실패(`identity_in_reads_the_xdg_git_config_of_the_home_it_is_given`: `dubious ownership` 거절, `identity_keeps_the_git_config_of_the_environment`). 세 번째 `identity_in_leaves_the_environment_git_config_out_for_another_home` 는 옛 고정도 지켜서 수리 전에 통과했어요(회귀 방지용이고 RED 가 아니에요). GREEN: 같은 명령에 `test(identity::)` 를 더해 exit 0, 10/10.
- 수리: `rev_parse`(`src/identity.rs`)가 `GIT_CONFIG_GLOBAL` 을 고정하지 않아요. 받은 HOME 이 `layout::environment_home` 과 다르면 자식 HOME 을 받은 HOME 으로 두고 `XDG_CONFIG_HOME`·`GIT_CONFIG_GLOBAL` 을 지우고, 같거나 없으면 환경 그대로예요. 자식 프로세스 도우미 `run_in_child_of`(`src/layout.rs`)는 개발자 환경의 `XDG_CONFIG_HOME`·`GIT_CONFIG_GLOBAL` 이 자식에게 새지 않게 먼저 지워요.
- AC-14: 충족. `identity_in_reads_the_xdg_git_config_of_the_home_it_is_given`(`<home>/.config/git/config` 에만 `safe.directory` 가 있는 받은 HOME 은 성공, 신뢰가 없는 HOME 은 `dubious ownership`).
- AC-15: 충족. `identity_keeps_the_git_config_of_the_environment`(자식 셋: XDG 파일만, `GIT_CONFIG_GLOBAL` 만, 둘 다 없음 = 통제는 거절; 신뢰가 있는 둘은 `identity()` 와 환경 HOME 을 넘긴 `identity_in` 이 모두 성공).
- AC-16: 충족. `identity_in_leaves_the_environment_git_config_out_for_another_home`(환경이 `.gitconfig`·XDG·`GIT_CONFIG_GLOBAL` 로 모두 신뢰하고 통제로 `identity()` 가 성공하는데, 다른 받은 HOME 은 거절돼요).
- 뮤턴트(적용을 `git diff --stat` 으로 확인하고 되돌린 뒤 `git status --short` 가 비어 있음을 확인; 모두 위 `-E` 필터, exit 100): ① `GIT_CONFIG_GLOBAL` 을 `/dev/null` 로 고정하는 판 2 모양으로 되돌림 → 2개 실패(XDG 테스트와 `identity_in_git_config_follows_the_home_it_is_given`) ② 같은 HOME 판정 삭제(항상 대체) → 1개 실패(`identity_keeps_the_git_config_of_the_environment`) ③ `XDG_CONFIG_HOME`·`GIT_CONFIG_GLOBAL` 을 지우지 않음 → 1개 실패(`identity_in_leaves_the_environment_git_config_out_for_another_home`).
- N2·N3: 고쳤어요. `moai spec show` 가 이제 write 10개·read_only 5개·forbidden 1개(`CLAUDE.md`)로 읽고(옛 판은 write 2개, forbidden 에 `src/` 포함) `moai spec check` 가 오류 0·경고 0 이에요. N1(`journal_root_in` 의 git 자식이 환경 HOME 설정을 읽음)은 D-5·D-8 대로 후속 조각이고, N4(Windows)는 not_run 이에요.
- not_run: 전체 `cargo test --locked`(리더 몫), Windows·Linux, 명령줄 git 실험(훅 거절, Rust 자식 프로세스 테스트로 대신), 실제 HOME 읽기 전후 비교.

**조각 3 첫 묶음 결과** (판 6; 브랜치 `spec/home-env-readers-20261124-r2`, 기준 `b3048a821a`, 구현 커밋 `dada13ba6b`; 승인은 리더 노트 10234 이고 사용자 목표 아래의 리더 판단이며 사용자 직접 선택은 아니에요; 로그는 `target/c410s3/` 에 있고 카드가 열려 있는 동안만 둬요)

- 특성 테스트(옛 코드): `baseline_inventory_reports_home_entries`, `host_wake_default_config_dir` 2/2 통과.
- 목록(옛 소스): `MOAI_HOME_READERS=enforce` 5/5 통과. 직접 읽기는 가장자리 28, 안쪽 19, 가장자리 밖 어댑터 호출 40줄이에요.
- RED: 두 줄만 가장자리로 옮기고 enforce → exit 100. 보인 줄: `unlisted inner HOME read in src/baseline.rs fn inventory`, `unlisted inner HOME read in src/ops/host_wake.rs fn default_config_dir`, `listed edge HOME read not found in src/ops/baseline_inventory.rs fn run`, `listed edge HOME read not found in src/ops/host_wake.rs fn run`.
- GREEN: `baseline::inventory` 가 HOME 을 인자로 받고 `default_config_dir` 는 `host_wake::run` 이 읽은 HOME 을 받아요. enforce 5/5, `--test baseline_git_env --test host_wake_cli` 14/14.
- AC-17: 충족. `cargo nextest run --offline --locked --test home_env_readers` exit 0, 5/5.
- AC-18: 충족. 위 명령 exit 0(5/5)과 `--test repo_scans_table` exit 0(34/34). 스캔 표 항목 `home-env-readers` 는 `"always": true` 예요(line-lease 로 쓰고 돌려줬어요).
- AC-19: 충족. 적힌 명령 exit 0, 1개 선택·통과.
- AC-20: 충족. 적힌 명령 exit 0, 1개 선택·통과.
- AC-21: 충족(이 head 와 기준에서). 가장자리 30, 안쪽 17 이고 `MOAI_HOME_READERS=enforce` 명령 exit 0, 5/5. 착지 tip 의 차이는 아래 "착지 tip" 줄을 보세요.
- AC-22: 충족. 적힌 명령 exit 0, 1개 선택·통과.
- AC-23: 충족. 적힌 명령 exit 0, 1개 선택·통과.
- AC-24: 충족(수동). `("src/update_tux.rs", "view_for", 1)` 을 2 로 어긋낸 채 `cargo nextest run --offline --locked --test home_env_readers` exit 0(5/5)과 `cargo test --offline --locked --test home_env_readers` exit 0(5 통과)이 모두 `home-env-readers observe: listed inner HOME read not found in src/update_tux.rs fn view_for` 를 보였어요. 줄은 1 로 되돌렸어요. `tests/spawn_cap_test.rs` 는 아래 읽는 테스트 실행에서 통과했어요.
- 읽는 테스트: `--test spawn_cap_test --test test_home_guard --test repo_scans_table --test state_folder_literal_scan --test ac_command_lint` 세 번째 실행 122/122 통과. 앞 두 번은 실패했어요: ① `every_file_reader_has_an_always_scan_or_an_exemption`(`tests/home_env_readers.rs reads src/ops` 가 `.rs` 접미사 범위 밖) ② `test_home_guard` 6개(새 override 블록의 위치).
- 새 테스트 파일 검사: `--test test_area_layout --test source_manifest_facts` 27/27 통과, `cargo clippy ... -- -D warnings`(lib, bin, 바뀐 테스트 셋) exit 0, `cargo fmt --all -- --check` 통과.
- 계획과 다른 점 ① nextest override 는 파일 맨 끝에 있어요(첫 override 뒤라는 조건은 지켜요). `tests/test_home_guard.rs` 의 `config_without_overrides` 가 override 블록을 `test-group` 줄까지 지워서, 그 줄이 없는 블록이 앞에 있으면 홈 가드 설정까지 지워져요. 그 파일은 이 카드의 범위 밖이라 고치지 않았고 설정 파일에 이유를 주석으로 남겼어요.
- 계획과 다른 점 ② 검출기는 첫 칸의 `#[cfg(test)]` 뿐 아니라 `#[cfg(all(test, …))]` 모듈도 세지 않아요(D-13 의 "테스트 코드는 세지 않아요"). 관찰 실행에서 `src/identity.rs` 의 테스트 읽기 2개가 잡혀서 고쳤고, `all(not(test), …)` 모듈은 그대로 센다는 가짜 소스 단언을 넣었어요.
- 계획과 다른 점 ③ 스캔 항목 범위는 접미사 없는 `src/` 예요(`src/` 의 `.rs` 가 아닌 파일 두 개를 바꿔도 이 항목이 골라져요).
- 뮤턴트(한 줄 바꾸고 그 테스트만 실행, 저장해 둔 파일로 되돌린 뒤 `git status --short` 0줄; 뮤턴트가 먹었는지는 `git diff --numstat` +1 -1):
  - guard-mutant: baseline_inventory_reports_home_entries | 빈 HOME 거르기를 더함 | exit 100
  - guard-mutant: host_wake_default_config_dir | `CLAUDE_CONFIG_DIR` 우선 가지를 못 타게 함 | exit 100
  - guard-mutant: the_checked_in_sources_match_the_list | `baseline::inventory` 에 HOME 읽기를 되살림(enforce) | exit 100 (같은 뮤턴트를 기본 모드로 돌리면 설계대로 exit 0 이고 알림 줄이 보여요)
  - guard-mutant: an_added_read_is_reported_with_its_file_and_function | `all(…)` 테스트 모듈 건너뛰기를 지움 | exit 100
  - guard-mutant: a_changed_adapter_call_is_reported | 옮긴 호출의 짝짓기를 지움 | exit 100
  - guard-mutant: enforce_mode_fails_on_a_difference | enforce 가 늘 통과 | exit 100
  - guard-mutant: observe_mode_reports_and_passes | 알림 줄을 버림 | exit 100
- 착지 tip: NeXT `f48ea9d815` 의 `src/` 를 임시 폴더에 풀어 이 목록으로 관찰 실행했어요(병합 없음). 이 묶음의 이동 네 줄 밖의 차이는 `listed call of hooks::env_home not found in src/lane_loop.rs fn stop_hook` 한 줄이에요(먼저 착지한 카드가 그 호출을 없앴어요). 직접 읽기 수는 같아요. 처리는 리더 결정 노트 10347 의 답을 따라요.
- 카드 기준 1: 충족하지 않았어요(not met). 안쪽 직접 읽기 17곳과 가장자리 밖 어댑터 호출이 남아요(Plan 의 다음 묶음).
- 카드 기준 2: 조각 2 에서 충족했고 이 묶음은 기준선 줄을 바꾸지 않았어요.
- not_run: 전체 `cargo test --locked`(리더 몫), Windows·Linux, 실제 HOME 읽기 전후 비교.

조각 3 둘째 묶음(리더 결정 노트 10986, 브랜치 `spec/home-env-readers-20261124-r3`, 기준 NeXT `af867cf151`, 예약 세대 6): 목록 파일 `tests/home_env_readers.rs` 의 세 줄만 고쳤어요. 코드는 바꾸지 않았어요. 구현 커밋은 `de6d5c73b7` 이에요. 세 줄은 에이전트 없이 레인이 직접 고쳤어요.

- 기준선 관찰: 기준 tip 에서 `MOAI_HOME_READERS=enforce cargo nextest run --offline --locked --no-fail-fast --test home_env_readers` → exit 100, 5개 중 4 통과, `the_checked_in_sources_match_the_list` 가 정확히 세 줄 차이로 실패했어요(앞서 착지한 조각들이 열한 줄의 차이를 이미 맞췄어요): `unlisted edge HOME read in src/main.rs fn run_internal_classified`, `unlisted inner HOME read in src/hook_engine.rs fn tool_preflight_deny`, `call of home_workdir::home moved from src/worktree_move.rs fn default_root to fn home_for_default_root`. 로그는 `target/c410-3/enforce-before.txt`.
- 세 줄 뒤: 같은 명령 → exit 0, 5개 모두 통과(`target/c410-3/enforce-after.txt`). 목록의 차이는 0이에요.
- `run_internal_classified` 5→6: 새 읽기는 `AuditClonePlan` 갈래의 입구(`src/main.rs`)에서 한 번 읽어 `plan_from` 에 `home` 과 `process_home` 으로 같이 넘겨요. 함수 안의 읽기 여섯 줄은 서로 다른 명령 갈래의 입구예요. 근거는 이 SPEC 의 D-2("가장자리(`main.rs`, `ops/*::run`, `ledger::resolve_home`)의 읽기는 남겨요")와 `SPEC-20261124-identity-home-arg` 의 REQ-36(가장자리가 프로세스 HOME 을 한 번 읽어 넘겨요)이고 REQ-13(목록과 소스의 개수가 어긋나면 한 줄 보고)이 이 줄을 요구해요.
- `tool_preflight_deny`: 안쪽 읽기라 `INNER_READS` 에 한 줄을 올렸고 행 위에 "입구로 옮길 것"이라는 사유 주석을 적었어요. 이것은 위반을 없앤 것이 아니라 알려진 안쪽 읽기로 세는 일이에요. 같은 파일의 `bypass_boundary_deny`, `session_check` 와 같은 처리이고, 세 읽기를 훅 입구 하나로 옮기는 일은 사용자 결정 뒤의 후속이에요(결정 노트 10986).
- `default_root` → `home_for_default_root`: 같은 호출(`home_workdir::home`)이 같은 모듈 안에서 이름만 바뀌었어요. 근거는 REQ-14(옮겨진 호출은 목록에도 같은 변경)와 REQ-17(enforce 에서 차이는 실패)이에요. 리더 메시지가 인용한 "REQ-22" 는 이 SPEC 에 없어요(요구사항은 REQ-18 까지)라 REQ-14 와 REQ-17 로 적었어요. 동작 변화는 없어요.
- 카드 기준 1: 충족하지 않았어요(not met). 안쪽 직접 읽기 18곳과 가장자리 밖 어댑터 호출이 남아요. 글을 바꿀지, 후속 카드를 열지는 사용자 결정이 먼저예요(결정 노트 10868 은 기준의 '거절'을 관찰로 읽는 것이고 '가장자리 밖 0'을 낮추는 결정이 아니에요). 이 묶음은 기준 1 을 표시하지 않아요.
- 카드 기준 2: 이전 묶음에서 충족했고 이 묶음은 기준선 파일을 바꾸지 않았어요.
- not_run: 전체 `cargo test --locked`(리더 몫), Windows·Linux, 실제 HOME 읽기 전후 비교, 뮤턴트(새 가드 테스트가 없어요. 목록 파일의 표 줄만 바꿨어요).
