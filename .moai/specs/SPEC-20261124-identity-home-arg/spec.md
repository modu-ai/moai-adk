---
id: SPEC-20261124-identity-home-arg
title: "identity takes the HOME as an argument: identity_in, the claim group, then every remaining caller, and the identity adapter removed"
status: draft
revision: 6
created: 2026-11-24
---

# identity takes the HOME as an argument: identity_in, the claim group, then every remaining caller, and the identity adapter removed

## Intent

카드 C-409 의 뿌리 수리예요. C-406 관찰 전체 실행(카드 노트 7598)에서 389개 테스트가 HOME 의 `user-assets.json` 을 읽었어요. 원인 한 줄은 `identity::identity`(`src/identity.rs:199`)가 환경 변수 HOME 을 직접 읽고, 그 HOME 으로 `home_root::path` → `state_root::classify_home` → `home_state::home_go_markers`(`src/home_state.rs:62`)를 불러 `user-assets.json` 을 연다는 거예요. 그런데 `identity` 의 답 가운데 HOME 에서 나오는 칸은 `home_state_root` 하나뿐이고, 그 칸을 읽는 제품 코드는 `src/audit_clone.rs:151` 뿐이에요. 나머지 호출부 15개 파일은 쓰지도 않는 칸을 위해 HOME 을 읽고 있어요.

이 SPEC 은 그 읽기를 인자로 바꿔요. 1조각(이 카드의 범위, `src/identity.rs`)은 HOME 을 인자로 받는 `identity_in` 을 더하고 환경 변수를 읽는 곳을 `identity` 한 곳(프로세스 가장자리)으로 모아요. 호출부는 이번 조각에서 건드리지 않아요. 호출부는 묶음별 다음 조각으로 나누고(Plan 의 조각표), 조각마다 범위 확장은 리더 결정 노트로 받아요.

판 5 는 카드의 남은 일(조각표의 2b·3·4)을 한 판으로 해요. 남은 제품 호출부(`src/apply.rs`, `src/audit.rs`, `src/audit_clone.rs`, `src/key_vault.rs`, `src/update_hooks.rs`, `src/launcher_init.rs`, `src/ops/lane_next.rs`, `src/ops/identity.rs`)를 `identity_in` 으로 옮기고, 어댑터 `identity` 를 지우고, 이 변경으로 HOME 을 더는 건드리지 않게 된 테스트의 기준선 줄을 지워요. HOME 을 다루는 변경이라 tier 2 예요.

판 6 은 리더 결정 노트 10218 을 반영해요. `apply` 영수증, `audit::observe`, `lane-next` 는 환경 HOME 을 넘겨 지금의 링크 HOME 거절을 지키고, 카드 기준 1 은 not met 으로 두고 착지해요.

## Interview

- Q1: `identity` 의 서명을 바로 바꾸나요, 새 함수를 더하나요? → A: 새 함수 `identity_in(path, home)` 를 더하고 `identity(path)` 는 그것을 부르는 가장자리 어댑터로 남겨요(D-1). 서명을 바로 바꾸면 한 번에 16개 파일이 깨져 한 카드·한 감사로 닿기엔 너무 커요.
- Q2: `src/home_state.rs` 를 고치나요? → A: 안 고쳐요. `home_go_markers(moai: &Path)` 는 이미 경로를 인자로 받고 환경 변수를 읽지 않아요(코드를 읽어 확인, D-2).
- Q3: HOME 을 환경 변수에서 읽는 곳이 `identity` 뿐인가요? → A: 아니에요. `src/layout.rs:35` `default_home_root` 도 읽고 클레임 장부 경로(`claim::journal_root` → `layout::switched_dir`)가 거쳐요. 이 SPEC 의 범위 밖이고 따로 카드로 열어요(D-3).
- Q4: (판 5) 남은 일을 나누나요? → A: 레인 제안은 한 판, READY 하나예요(D-8). 어댑터 삭제는 마지막 호출부와 같이 가야 해요.
- Q5: (판 5) 어느 호출부가 HOME 을 받나요? → A: 홈 상태 폴더 아래를 읽거나 쓰는 길과, 그 거절이 첫 문턱인 길이에요: 키 금고, 첫 실행 초기화, update 훅 층, audit clone 계획, `identity` 명령(D-10, D-11, D-12, D-14).
- Q6: (판 6) `apply` 영수증, `audit::observe`, `lane-next` 는 무엇을 넘기나요? → A: 환경 HOME 이에요(리더 결정 노트 10218). 링크 HOME 거절과 지금 동작이 그대로예요(REQ-20..23, D-13).
- Q7: (판 5) 빈 HOME 은 어디서 거르나요? → A: `identity_in` 한 곳이에요(D-9).
- Q8: (판 5) `identity` 를 쓰던 테스트는 어떻게 하나요? → A: 환경 HOME 을 드러내어 읽어 `identity_in` 에 넘기거나 `None` 을 넘기도록 고치고, 어댑터 자체를 재던 AC-4 테스트는 지워요(D-14).
- Q9: (판 5) 기준선은 몇 줄 줄어요? → A: 읽기만으로는 몰라요(unverified). 변경 전후 관찰 실행으로 재요(D-15).
- Q10: (판 6) 카드 기준 1 은 어떻게 적나요? → A: not met 으로 두고 `card land --keep-open` 으로 착지해요. 비교용 HOME 을 인자로 넘기는 일은 후속 카드예요(리더 결정 노트 10218, D-16).

## Non-goals

- `layout::default_home_root`·`switched_dir`·`switched_dir_checked` 를 거치는 환경 HOME 읽기. 카드 C-410 의 몫이에요. `apply` 영수증 저장소는 이 길이라 그 테스트의 기준선 줄은 이 판에서 줄지 않을 수 있어요.
- 공유 범위(`--shared`) 금고의 홈 상태 폴더 링크 검사. 공유 범위는 지금도 identity 를 부르지 않아요(`src/key_vault.rs`, `scope_for`). 후속 항목이에요.
- 홈 상태 폴더 아래 중간 구성요소의 링크 검사를 새로 더하는 일.
- `rev_parse` 의 git 자식 환경 규칙 바꾸기(SPEC-20261124-home-env-readers D-8).
- `identity` 답의 모양 바꾸기. 값 규칙의 변화는 REQ-10(빈 HOME) 하나예요.
- `src/home_state.rs` 변경(D-2).
- `identity_in` 이 git 자식 준비에서 비교용으로 읽는 환경 HOME 을 인자로 넘기는 일. 후속 카드예요(D-16).

## Requirements

- REQ-1: The `identity_in` function shall resolve `home_state_root` from the HOME it is given and from no environment variable.
- REQ-2: WHEN `identity_in` is given no HOME, it shall return `home_state_root` as `null` without resolving a home state root. · failure: none — the other fields do not depend on a HOME
- REQ-3: WHEN `identity_in` is given a HOME, `home_state_root` shall be the home state root of that HOME even when the environment HOME is another folder.
- REQ-4 (REQ-3): IF the home state root of a given HOME is a symlink, THEN `identity_in` shall return the error `home state root is a symlink; refused` as `identity` returned it before.
- REQ-5: The `identity` adapter shall be retired, with REQ-25 and REQ-26 in place of its return rule of revisions 1 to 4.
- REQ-6 (slice 2a): The claim group shall resolve every identity through `identity_in(path, None)` and through no call of `identity`. · the group is the seven files of the Scope
- REQ-7 (slice 2a): WHEN the environment HOME holds a symlinked home state root, each claim-group function other than `claim::journal_root` that resolves an identity shall answer as it answers with a sound HOME. · failure: none — these functions read and write nothing under that home state root
- REQ-8 (slice 2a): The claim group shall read from an identity answer no field named `home_state_root`.
- REQ-9 (slice 2a repair): WHEN the environment HOME holds a symlinked home state root and `layout.json` switches the claim journal to the home state, `claim::journal_root` shall return the error `home state root is a symlink; refused` before it returns the switched path. · failure: none — the answer is the refusal itself, and nothing is written behind the link
- REQ-10: The `identity_in` function shall treat an empty HOME as no HOME.
- REQ-11: The key vault shall resolve a project scope through `identity_in` with the HOME the vault itself uses (`key_vault::home`), and with no HOME where there is none.
- REQ-12: WHEN the home state root of the HOME given for a project scope is a symlink, the key vault shall refuse the scope with an answer that carries `home state root is a symlink; refused` before it reads, writes or removes a vault entry. · failure: none — the answer is the refusal, and nothing behind the link changes · the shared scope is outside this requirement
- REQ-13: The first-launch initialization shall resolve every identity through `identity_in` with `key_vault::home` as it is returned, not canonicalized.
- REQ-14: WHEN the home state root of that HOME is a symlink, `launcher_init::prepare` shall run no `git init` and create no project file. · failure: none — the launch goes on as it did before
- REQ-15: The update hook layer shall resolve the project through `identity_in` with the HOME that `plan_in` is given.
- REQ-16: WHEN the home state root of that HOME is a symlink, `update_hooks::plan_in` shall name no project and record a skipped row whose reason is `no project here: home state root is a symlink; refused`. · failure: none — the layer writes nothing
- REQ-17: The `audit_clone::plan` function shall resolve the evidence folder from the HOME it is given and from no environment variable.
- REQ-18: WHEN the home state root of that HOME is a symlink, `audit_clone::plan` shall return the error `home state root is a symlink; refused`. · failure: none — the answer is the refusal
- REQ-19: WHEN `audit_clone::plan` is given no HOME for a project that has a project id, it shall answer with a null evidence folder and the reason `HOME unavailable`. · failure: none — the rest of the plan is answered
- REQ-20: The three callers (the `apply` receipt and `audit::observe` and `lane-next`) shall resolve identities through `identity_in` with the environment HOME read at the call.
- REQ-21: WHEN the environment HOME holds a symlinked home state root, `apply::Receipt::new` shall record the clone id and the workspace id as null.
- REQ-22: WHEN the environment HOME holds a symlinked home state root, `audit::observe` shall return the error `home state root is a symlink; refused`.
- REQ-23: WHEN the environment HOME holds a symlinked home state root and the idle count is on, `lane-next` shall answer without keeping a count.
- REQ-24: WHEN the receipt store is switched to a home state folder behind a symlink, the apply engine shall keep refusing the receipt path. · failure: none — nothing is written behind the link
- REQ-25: The `identity` command shall give `identity_in` the environment HOME read once, an unset or empty HOME counting as no HOME.
- REQ-26: The source tree shall hold no `identity` function in the `identity` module and no call of one under `src` or `tests`.
- REQ-27: The baseline change shall remove exactly the lines of the tests that an observation run shows touching their HOME before this change and not after it, with `BASELINE_CAP` set to the number of lines left.
- REQ-28 (REQ-27): IF no line drops, THEN the lane shall record card criterion 2 as not met and leave the baseline as it is.
- REQ-29 (REQ-21): IF the identity cannot be resolved, THEN `apply::Receipt::new` shall record both ids as null without failing the apply.
- REQ-30 (REQ-22): IF the identity cannot be resolved, THEN `audit::observe` shall return the resolver's error.
- REQ-31 (REQ-23): IF no clone is resolved, THEN `lane-next` shall answer without counting.

## Acceptance

- AC-1 (REQ-1, REQ-3): 환경 변수 HOME 과 다른 임시 폴더를 `identity_in` 에 넘기면 `home_state_root` 가 그 임시 폴더 아래 경로예요. · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_in_uses_the_home_it_is_given)'`
- AC-2 (REQ-2): 환경 변수 HOME 이 `.moai-cli` 가 심볼릭 링크인 폴더인 자식 프로세스에서, `identity_in(path, None)` 은 오류 없이 `home_state_root: null` 을 답하고 같은 환경의 `identity(path)` 는 심볼릭 링크 거부 오류를 답해요(None 이 환경 HOME 을 읽지 않는다는 관찰). · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_in_without_a_home_ignores_the_environment_home)'` · 판 5: 어댑터가 없어졌으므로 같은 이름의 테스트가 거절 쪽을 환경 HOME 을 읽어 넘긴 `identity_in` 으로 확인해요.
- AC-3 (REQ-4): 임시 HOME 의 `.moai-cli` 가 심볼릭 링크면 `identity_in(path, Some(home))` 이 거부 오류를 답해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_in_refuses_a_symlinked_home_state_root)'`
- AC-4 (REQ-5): 판 5 에서 폐기해요. 아래 문장과 verify 명령은 판 1 의 기록이고 테스트는 지워요(판 5 의 대신은 `moai identity` 의 CLI 테스트예요). 환경 변수 HOME 이 설정·빈 값·없음 세 경우에 `identity(path)` 의 답이 `identity_in(path, 같은 값)` 의 답과 같아요(자식 프로세스에서 환경을 고정). · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_equals_identity_in_with_the_environment_home)'`
- AC-5 (REQ-1..REQ-5): 조각 1 은 `src/identity.rs` 와 SPEC 폴더만 바꿨고 호출부 파일 diff 는 0이었어요(착지한 head `95c87d1c71`, 기준 `acf8ab1997`). · verify: `git diff --stat acf8ab1997..686e9d88d8 -- src`
- AC-6 (REQ-6): 클레임 묶음 7개 파일에 `identity::identity(` 호출이 하나도 없어요(테스트 모듈 안의 호출 포함). · verify: `grep -n "identity::identity(" src/claim.rs src/claim_link.rs src/lane_registry.rs src/session.rs src/worktree_move.rs src/run.rs src/goal.rs` 가 아무것도 찍지 않고 exit 1
- AC-7 (REQ-7): 파일마다 테스트 하나가 자식 프로세스에서 환경 HOME 의 홈 상태 폴더를 심볼릭 링크로 만들고, 그 파일의 함수가 거부 없이 건전한 HOME 과 같은 답을 해요: `claim::journal_root`(`layout.json` 이 없는 HOME 에서), `claim_link` 의 worktree 검사(`mine`)와 `claim::worktree_id` 경로, `lane_registry::registry_paths`·`worktree_of`, `session::agent_seq_dir`, `worktree_move` 의 계획 함수, `run::baseline_store_dir`, `goal::store`. 테스트 이름은 RED 를 본 뒤 SPEC Result 에 적어요. · verify: `cargo nextest run --offline --locked --lib -E 'test(home_arg_claim_group)'` (테스트 이름이 모두 이 접두사를 가져요)
- AC-8 (REQ-6, REQ-8): 판 3·4 의 diff 는 클레임 묶음 7개 파일, `src/identity.rs`(자식 프로세스 도우미를 `pub(crate)` 로 여는 `#[cfg(test)]` 부분과 `refuse_symlink` 를 `pub(crate)` 로 여는 한 줄만), SPEC 폴더만 닿아요. `home_state_root` 를 읽는 줄이 묶음에 새로 없어요. · verify: `git diff --name-only <base>..<head>` 와 `grep -n "home_state_root" src/claim.rs src/claim_link.rs src/lane_registry.rs src/session.rs src/worktree_move.rs src/run.rs src/goal.rs` 가 exit 1
- AC-9 (REQ-6, REQ-7, REQ-9): 묶음의 기존 테스트가 그대로 통과해요. · verify: `moai check affected <base>..<head>` 가 PASS
- AC-10 (REQ-9): 자식 프로세스의 환경 HOME 에서 홈 상태 폴더가 링크이고 그 링크 너머에 `projects/<id>/layout.json` 이 있을 때, 쓰기 진입점 `claim::release_cli` 가 `home state root is a symlink; refused` 로 거절하고 링크 대상의 `projects/<id>` 에 `layout.json` 말고는 아무 파일도 생기지 않아요. 이 거절을 지우면 이 테스트가 실패해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(home_arg_claim_group_claim_journal_root_refuses_a_linked_home_when_switched)'`
- AC-11 (REQ-26): 어댑터와 그 호출이 없어요. · verify: `git grep -nE 'identity::identity([^_a-zA-Z0-9]|$)' -- src tests ':!tests/fixtures'` 와 `git grep -n 'fn identity(' -- src/identity.rs` 가 둘 다 아무것도 찍지 않고 exit 1
- AC-12 (REQ-10): `identity_in(path, Some(""))` 의 `home_state_root` 가 null 이에요. 수리 전에는 상대 경로가 나와 실패해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(home_arg_rest_identity_in_treats_an_empty_home_as_none)'`
- AC-13 (REQ-11, REQ-12): 환경 HOME 은 건전하고 넘긴 HOME 만 링크일 때 `read_here` 와 `scope_for` 가 프로젝트 범위를 거절하고, HOME 이 없으면 범위를 답해요. `read_here` 테스트는 수리 전에 실패하고, `scope_for` 테스트는 서명이 없어 컴파일이 실패해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(home_arg_rest_read_here) | test(home_arg_rest_scope_for)'`
- AC-14 (REQ-12): 실제 바이너리를 임시 HOME(홈 상태 폴더가 링크이고 링크 너머에 금고 항목이 있어요)에서 돌려 `key-status`, `key-set`, `key-remove --yes` 가 모두 거절하고 링크 대상의 앞뒤 스냅샷이 같아요. 지금도 통과하는 고정 테스트이고, `scope_here` 가 `None` 을 넘기면 실패해요. 기존 고정 테스트 `an_identity_error_stops_the_launch_even_with_a_shared_key`(`tests/launcher_key_vault.rs`)도 그대로 통과해요. · verify: `cargo nextest run --offline --locked --test key_vault_cli --test launcher_key_vault -E 'test(home_arg_rest_key_commands_refuse_a_linked_home_state_root) | test(an_identity_error_stops_the_launch_even_with_a_shared_key)'`
- AC-15 (REQ-13, REQ-14): 자식 프로세스의 환경 HOME 이 링크일 때 저장소가 아닌 폴더에서 `prepare` 가 `.git` 도 프로젝트 파일도 만들지 않아요. 고정 테스트이고, `identity_in(cwd, None)` 으로 바꾸면 실패해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(home_arg_rest_a_linked_home_initializes_nothing)'`
- AC-16 (REQ-15, REQ-16): 환경 HOME 은 건전하고 `plan_in` 에 넘긴 HOME 만 링크이면 프로젝트가 없고 위 이유가 남아요. 수리 전에는 실패해요. · verify: `cargo nextest run --offline --locked --test hook_legacy_cli -E 'test(home_arg_rest_update_layer_refuses_a_linked_given_home)'`
- AC-17 (REQ-17, REQ-18, REQ-19): 넘긴 링크 HOME 은 거절, `None` 은 `HOME unavailable`, 기존 audit plan 테스트는 그대로 통과해요. 새 테스트의 RED 는 컴파일 실패예요. · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(audit_plan)'`
- AC-18 (REQ-20, REQ-21, REQ-24, REQ-29): 링크 환경 HOME 에서 영수증의 두 id 가 null 이고, 같은 테스트의 건전한 HOME 대조에서는 두 id 가 있어요. 기준 커밋에서도 통과하는 고정 테스트이고, 호출이 `None` 을 넘기면 실패해요. 홈으로 돌린 영수증 저장소의 거절은 기존 테스트가 지켜요. · verify: `cargo nextest run --offline --locked --lib -E 'test(home_arg_rest_apply_receipt_records_no_ids_under_a_linked_home)'` 와 `cargo nextest run --offline --locked --test home_writer_links -E 'test(receipt_refuses)'`
- AC-19 (REQ-22, REQ-30): 링크 환경 HOME 에서 `observe` 가 `home state root is a symlink; refused` 로 거절해요. 고정 테스트이고, 호출이 `None` 을 넘기면 실패해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(home_arg_rest_audit_observe_refuses_a_linked_home)'`
- AC-20 (REQ-23, REQ-31): 링크 환경 HOME 에서 스냅샷 읽기에 `--max-empty` 를 줘도 명령은 답하고 클론 상태 폴더의 `lane-idle` 에 횟수가 생기지 않아요. 같은 테스트의 건전한 HOME 대조에서는 생겨요. 고정 테스트이고, 호출이 `None` 을 넘기면 실패해요. · verify: `cargo nextest run --offline --locked --test lane_next_cli -E 'test(home_arg_rest_lane_next_counts_nothing_under_a_linked_home)'`
- AC-21 (REQ-25): `moai identity` 는 설정된 HOME 의 홈 상태 폴더를 답하고, 링크면 실패하고, HOME 이 없거나 비면 `home_state_root` 가 null 이에요. 고정 테스트예요. · verify: `cargo nextest run --offline --locked --test cli_diagnose`
- AC-22 (REQ-26, REQ-7, REQ-9): 어댑터를 쓰던 테스트(클레임 묶음 도우미, identity 의 git 설정 테스트, `claims_legacy_dir`)가 고친 뒤에도 통과해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(home_arg_claim_group) | test(identity::) | test(claims_legacy_dir_ignores)'`
- AC-23 (REQ-27, REQ-28): 기준 커밋의 기준선에 이름이 있는 테스트만, 변경 전과 변경 후에 관찰 모드로 돌려 두 로그를 비교하고, 앞에는 닿고 뒤에는 닿지 않은 줄만 지워요. 먼저 `cp tests/fixtures/test_home_guard/baseline.tsv target/c409/baseline-base.tsv` 로 기준 커밋의 목록을 남겨요(코드 변경 전). 관찰 실행의 HOME 은 가드가 만드는 임시 폴더예요(`tools/test_home_guard.sh`). 지운 뒤에는 같은 명령을 환경 변수 없이(enforce) 돌려 exit 0, `cargo nextest run --offline --locked --test test_home_guard` exit 0 이에요. · verify: `MOAI_HOME_GUARD=observe MOAI_HOME_GUARD_LOG="$PWD/target/c409/guard-before.log" cargo nextest run --offline --locked --no-fail-fast $(awk -F'\t' '!/^#/ && NF==2 {print ($1=="moai_cli" ? "--lib" : "--test=" $1)}' target/c409/baseline-base.tsv | sort -u) -E "$(awk -F'\t' '!/^#/ && NF==2 {printf "%stest(=%s)", s, $2; s=" | "}' target/c409/baseline-base.tsv)"` (heavy slot 안에서; 변경 후에는 로그 이름만 `guard-after.log`)
- AC-24 (REQ-6..REQ-31): 기존 테스트가 통과해요. · verify: `moai check affected --card <base>..<head>` 가 PASS

## Scope

- write (판 5): `src/identity.rs`, `src/layout.rs`(테스트의 `identity` 호출 한 곳과 그 주석만), `src/key_vault.rs`, `src/launcher_init.rs`, `src/update_hooks.rs`, `src/audit_clone.rs`, `src/apply.rs`, `src/audit.rs`, `src/ops/lane_next.rs`, `src/ops/identity.rs`, `src/main.rs`, `tests/home_writer_links.rs`, `tests/key_vault_cli.rs`, `tests/hook_legacy_cli.rs`, `tests/lane_next_cli.rs`, `tests/cli_diagnose.rs`, `tests/test_home_guard.rs`, `tests/fixtures/test_home_guard/baseline.tsv`, `.moai/specs/SPEC-20261124-identity-home-arg/spec.md`. C-410 이 먼저 착지해 `tests/home_env_readers.rs` 의 자리 목록이 생기면 그 파일의 해당 항목도 써요.
- read-only: `src/home_state.rs`(D-2), `src/home_root.rs`, `src/state_root.rs`, `src/session.rs`, `src/model_effort_questions.rs`, `src/ops/key_*.rs`, `tests/launcher_key_vault.rs`
- forbidden: 실제 HOME(읽기 전후 비교로도 건드리지 않아요, 임시 HOME 픽스처만), 네트워크, `CLAUDE.md`, `layout::default_home_root`·`switched_dir*` 의 동작(C-410)

## Plan

조각표(묶음별, 앞 조각이 착지한 뒤 다음 조각이 서요). 조각 2 이후의 범위는 이 SPEC 이 정하지 않고 리더가 노트로 정해요.

| 조각 | 무엇 | 파일 | 동작이 같다는 증거 |
| --- | --- | --- | --- |
| 1 (이 판) | `identity_in(path, home)` 추가, `identity` 는 환경 HOME 한 번 읽어 위임 | `src/identity.rs` | AC-1..4: 같은 입력이면 같은 답 |
| 2 (2a 는 판 3) | HOME 칸을 안 쓰는 호출부가 `identity_in(path, None)` 을 불러요 | 2a 클레임 묶음: `src/claim.rs`, `src/claim_link.rs`, `src/lane_registry.rs`, `src/session.rs`, `src/worktree_move.rs`, `src/run.rs`, `src/goal.rs`. 2b 나머지: `src/layout.rs`(identity 호출만), `src/key_vault.rs`, `src/audit.rs`, `src/update_hooks.rs`, `src/apply.rs`, `src/launcher_init.rs`, `src/ops/lane_next.rs` | 이 파일들이 읽는 칸 목록(아래 읽기 증거)에 `home_state_root` 가 없어요; 파일마다 기존 테스트가 그대로 통과 |
| 3 | HOME 칸을 쓰는 곳이 HOME 을 받아요: `audit_clone` 은 부르는 쪽에서 HOME 을 받고, 명령 가장자리(`src/ops/identity.rs`)는 환경 HOME 을 넘겨요. 마지막 호출부가 떠나면 `identity` 어댑터를 지워요 | `src/audit_clone.rs`, `src/ops/identity.rs`, `src/identity.rs` | `moai identity` 출력이 바이트 같아요(`tests/cli_diagnose.rs` 가 `home_state_root` 를 단언) |
| 4 | C-406 착지 뒤 `baseline.tsv` 에서 줄어든 테스트 줄을 지워요 | `tests/fixtures/test_home_guard/baseline.tsv` | 가드가 enforce 에서 지워진 줄의 테스트를 통과시켜요 |

읽기 증거(2026-11-24, `grep -rn "home_state_root" src` 와 호출부 읽기): `home_state_root` 키를 이름으로 읽는 제품 코드는 `src/audit_clone.rs:151` 뿐이에요. 호출부 `claim.rs`, `session.rs`, `lane_registry.rs`, `layout.rs`, `ops/lane_next.rs` 는 `identity_status`, `clone_anchor`, `work_tree`, `workspace_id` 같은 칸만 읽어요. 조각 2 는 파일마다 이 읽기를 다시 해서 증거로 남겨요.

조각 2a 읽기 증거(2026-11-24, 판 3 초안을 쓰며 7개 파일의 호출 11곳과 테스트 호출 1곳을 열어 읽음): `claim_link.rs:274·299`(`identity_status`, `workspace_kind`, `clone_anchor`, `work_tree`, `claim::worktree_id` 로 `work_tree`·`clone_anchor`·`workspace_git_dir`), `claim_link.rs:440`(`work_tree`), `claim.rs:80`(`identity_status`, `clone_anchor`), `claim.rs:1488`(`worktree_id`), `session.rs:367`(`identity_status`, `clone_anchor`), `lane_registry.rs:91`(`identity_status`, `clone_anchor`), `lane_registry.rs:315`(`identity_status`, `work_tree`), `goal.rs:76`(`identity_status`, `clone_anchor`), `worktree_move.rs:217`(`clone_id`), `run.rs:552`(`clone_anchor`). 어느 곳도 `home_state_root` 를 읽지 않아요. `claim_link.rs:1202` 는 테스트 모듈 안의 호출이에요. 이 읽기는 판 3 의 RUN 에서 `grep` 으로 한 번 더 확인해요(AC-8).

판 5 는 표의 2b·3·4 를 한 판, READY 하나로 해요(D-8). 순서: 승인과 범위 확장 → 변경 전 관찰 실행 → 동작 RED → 호출부 하나씩 → 어댑터 삭제 → 뮤턴트 → 변경 후 관찰 실행과 기준선 → 검사와 READY.

판 6 이 위 순서와 표에서 바꾸는 것: 표의 2b 가운데 `apply`·`audit`·`lane-next` 는 `None` 대신 환경 HOME 을 넘겨요(리더 결정 노트 10218). `identity_in` 의 빈 HOME 거르기(AC-12)의 RED·GREEN 을 이 호출들을 바꾸기 전에 해요. 변경 전 관찰 실행 앞에 예약의 lease 를 확인해요. 관찰 실행은 heavy slot `--max 2` 안에서 돌려요.

읽기 증거(판 5, 2026-11-26, `b3048a821a`): 남은 제품 호출은 `apply.rs:427`(`clone_id`, `workspace_id`), `audit.rs:271`(`workspace_id`), `audit_clone.rs:148`(`project_id`, `home_state_root`), `key_vault.rs:204`(`identity_status`, `work_tree`, `clone_id`, `clone_anchor`, `project_id`), `update_hooks.rs:121`(`work_tree`), `ops/lane_next.rs:81`(`identity_status`, `clone_anchor`), `ops/identity.rs:17`(전체), `launcher_init.rs:101·256·269`(`identity_status`, `work_tree`, `clone_anchor`)예요.

환경 HOME 을 넘기는 세 길(REQ-20, 판 6): `src/apply.rs` 의 `identity_fields` 와 `src/audit.rs` 의 `observe` 는 그 자리에서 `layout::environment_home()` 을 읽어 `identity_in` 에 넘겨요(선례 `src/claim.rs` 의 `journal_root`). `Receipt::new` 와 `observe` 의 서명은 바꾸지 않아요. `src/ops/lane_next.rs` 는 바이너리 쪽이라 `std::env::var_os("HOME")` 을 읽어 넘겨요. `environment_home()` 은 빈 값을 거르지 않고 `identity_in` 이 걸러요(REQ-10). 거절을 삼키는 모양(`apply` 의 `_ => (None, None)`, `lane-next` 의 `.ok()`)과 `audit` 의 `?` 는 그대로예요.

순서(판 1): 1) 결정 D-1..D-4 와 SPEC 승인을 리더에게 노트로 받아요. 2) RED: 위 테스트 이름들로 실패를 먼저 봐요. 3) GREEN: `identity` 본문을 `identity_in` 으로 옮기고 `identity` 는 한 줄 어댑터로 줄여요. 4) 이웃: `identity` 를 거치는 기존 테스트 대표 몇 개, fmt, clippy.

## Risks

- 이 조각만으로는 `baseline.tsv` 줄이 줄지 않아요(호출부가 아직 `identity` 를 불러요). 카드 기준 2 는 조각 2·3 뒤에 충족돼요 — 조각 1 의 READY 는 기준 2 를 not met 으로 적어요.
- 두 번째 뿌리 `layout::default_home_root` 때문에 클레임 묶음 테스트(예: `claim_link_cli` 27개)는 조각 2 뒤에도 HOME 읽기가 남을 수 있어요. 몇 개가 줄지는 C-406 가드를 observe 로 다시 돌려 재야 해요(그 전에는 unverified).
- `identity_in(path, None)` 은 `home_state_root` 를 null 로 만들어요. 그 칸을 쓰는 새 호출부가 생기면 조용히 null 을 받아요 — 조각 3 이 그 칸을 쓰는 곳을 HOME 인자로 못 박아요.
- 자식 프로세스에서 환경을 고정하는 테스트는 프로세스 전체 환경을 바꾸지 않아 병렬 테스트를 깨지 않아요(부모는 `std::env::set_var` 를 쓰지 않아요).

- 조각 2a 는 동작이 하나 바뀌어요: 환경 HOME 의 홈 상태 폴더가 심볼릭 링크일 때 클레임 묶음의 함수들은 지금 `home state root is a symlink; refused` 로 실패하고, 2a 뒤에는 그 폴더를 풀지 않아 실패하지 않아요(REQ-7, D-5). 단 `claim::journal_root` 는 예외예요: `layout.json` 이 저널을 그 홈 상태 폴더로 돌려 놓았으면 `identity` 의 거절이 사라진 자리에서 링크 너머에 저널을 쓰게 되므로, `journal_root` 가 같은 문장으로 거절해요(REQ-9, D-7; 감사 B1 수리). 묶음의 나머지 함수는 홈 상태 폴더 아래에 읽거나 쓰지 않아요(`grep -nwE "switched_dir|default_home_root|active_home|store_dir" src/lane_registry.rs src/session.rs src/goal.rs src/run.rs src/worktree_move.rs src/claim_link.rs` 가 아무것도 찍지 않고 exit 1, 2026-11-24). 거부를 단언하는 기존 테스트는 `src/contract.rs:553` 의 오류 분류 표뿐이었어요(`grep -rn "home state root is a symlink" src tests`, 2026-11-24).
- 조각 2a 도 `baseline.tsv` 줄을 줄이지 못할 수 있어요: 클레임 묶음은 `layout::switched_dir` → `layout::default_home_root` 로 환경 HOME 을 또 읽어요(C-410). 줄어드는 양은 판 3 뒤에 C-406 가드를 다시 돌려 재고, 그 전의 수는 unverified 예요(D-4).
- 7개 파일 약 12곳을 한 조각에 바꿔요. 호출 한 줄씩의 기계적 교체이고 `identity` 어댑터는 그대로 남아 있어서(조각 3 이 지워요) 되돌리기는 쉬워요.
- 판 5, 카드 기준 1: `identity_in` 의 git 자식 준비(`rev_parse`)는 받은 HOME 이 환경 HOME 과 같은지 보려고 `layout::environment_home()` 을 읽고, HOME 이 없거나 같으면 자식이 환경의 전역 git 설정을 읽어요. `home_state_root` 는 환경에서 오지 않아요. 판 6: 이 예외 때문에 기준 1 은 not met 이에요(D-16).
- 판 5: Windows 에서 HOME 이 없으면 `key_vault::home()` 이 `USERPROFILE` 을 줘서 금고와 첫 실행 초기화에 링크 거절이 새로 생기고 git 자식의 설정 조회가 바뀌어요. Mac 테스트로는 확인할 수 없어요(unverified).
- 판 5: 공유 범위 금고는 홈 상태 폴더가 링크여도 지금처럼 링크 너머를 읽고 써요(이 판 전부터).
- 판 6: 남은 제품 호출이 모두 HOME 을 넘기므로 `identity_in` 은 이 길들에서 계속 그 HOME 을 읽어요. 줄어들 수 있는 것은 가드의 HOME 과 다른 픽스처 HOME 을 넘기게 되는 테스트(후보: `plan_in` 을 거치는 lib 의 `update::` 줄)뿐일 것으로 봐요. 0 에서 몇 줄 사이이고 unverified 예요. `lane_next_cli`, `audit_admission`, `apply_*` 줄은 남을 것으로 봐요. 0줄이면 기준 2 도 not met 이에요(REQ-28).
- 판 5, 다른 카드와의 겹침: C-457(`spec/glm-key-scope-20261124`, 착지 전 커밋 `cc22a0d9a8`)이 `src/key_vault.rs` 에 `project_scope_in(dir, home)` 를 더했고 `project_scope(dir)` 는 `identity(dir)` 를 불러요. 판 5 의 키 금고 변경은 같은 이름과 몸통(`scope_of_identity`)을 써서 합치기를 맞추고, 뒤에 착지하는 쪽이 남은 `identity` 호출을 없애야 컴파일돼요. C-410 의 새 목록 검사(`tests/home_env_readers.rs`)가 먼저 착지하면 이 판이 바꾼 HOME 읽기 자리를 그 목록에 맞춰요. `src/main.rs` 의 audit clone 가장자리에 환경 HOME 읽기가 하나 늘어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `identity` 서명을 바꾸지 않고 `identity_in(path, home: Option<&Path>)` 를 더해요. `identity` 는 환경 HOME 을 한 번 읽어 `identity_in` 에 넘기는 가장자리 어댑터예요 | 호출부 16개 파일·약 21곳을 한 카드로 못 바꿔요(카드 본문 "조각으로 나눠도 돼요"). 더하기만 하는 조각이라 되돌리기 쉬워요 | `identity_in` 을 지우고 본문을 `identity` 로 되돌리기 |
| D-2 | `src/home_state.rs` 는 이 SPEC 에서 안 고쳐요 | `home_go_markers(moai: &Path)` 는 인자로 경로를 받고 환경 변수를 읽지 않아요(`src/home_state.rs:62`). 카드 범위에는 그대로 두고 쓰지 않아요 | 해당 없음 |
| D-3 | 두 번째 뿌리 `layout::default_home_root`(`src/layout.rs:34`)와 환경 HOME 을 읽는 다른 곳은 이 SPEC 이 아니라 후속 카드예요 | 같은 종류가 두 번 보이면 소스를 고쳐요(재발 방지). 범위를 넓히지 않고 리더가 카드를 열어요 | 후속 카드를 닫기 |
| D-4 | 줄어드는 양은 조각마다 C-406 가드(observe)를 다시 돌려 재요. 그 전의 수는 unverified 예요 | 한 조각이 몇 개를 줄일지는 읽기만으로 정할 수 없어요 | 해당 없음 |
| D-5 | 조각 2a(판 3)는 클레임 묶음 7개 파일을 `identity_in(path, None)` 으로 바꿔요. 환경 HOME 의 홈 상태 폴더가 심볼릭 링크일 때 이 함수들이 더는 거부하지 않는 것을 받아들이되, `claim::journal_root` 는 예외로 거부를 유지해요(REQ-7, REQ-9). 이 파일들은 `home_state_root` 를 읽지 않아요 | 위 읽기 증거. 거부는 `identity` 가 `home_state_root` 를 풀 때 났어요. 묶음은 그 칸을 안 쓰지만 `journal_root` 는 같은 환경 HOME 의 홈 상태 폴더를 `layout::switched_dir` 로 거쳐 그 아래에 저널을 쓰므로 그 거부가 보호하던 길이에요(감사 B1). 카드 write scope 에 7개 파일은 리더가 노트 7748 로 넓혔어요 | 7개 파일의 호출 줄을 `identity::identity(` 로 되돌리기(어댑터가 남아 있어요) |
| D-6 | 판 2 감사의 비차단 R1(SPEC 본문의 어긋난 조사 둘과 개발자가 안 쓰는 번역투 낱말 둘)을 판 3 에서 함께 고쳐요 | 리더 결정 노트 7667 | 해당 없음 |
| D-7 | `claim::journal_root` 는 `layout::switched_dir` 가 홈 상태 경로를 돌려줄 때, 그 경로를 돌려주기 전에 환경 HOME 의 홈 상태 폴더가 링크이면 `home state root is a symlink; refused` 로 거절해요. 검사는 `identity` 의 기존 `refuse_symlink` 를 `pub(crate)` 로 열어 써요 | 감사 #409 판 3 B1 과 리더 결정 노트 7790. `src/layout.rs` 는 C-410 몫이라 고치지 않아요. `layout.json` 이 없으면 저널은 클론 쪽에 남고 홈 상태 폴더를 쓰지 않으므로 거절하지 않아요(2a 전에는 거절했어요, REQ-7 의 동작 변화에 든 부분이에요) | `journal_root` 의 검사 세 줄과 가시성 한 줄을 되돌리기 |
| D-8 | 남은 일 전부를 한 판, READY 하나로 해요. 대안: 2b 와 3+4 로 나눠요 | 어댑터 삭제와 기준선 측정이 모든 호출부를 요구해요 | 커밋 되돌리기 |
| D-9 | 빈 HOME 은 `identity_in` 이 `None` 으로 봐요. 대안: 가장자리마다 걸러요 | `default_home_root_in` 도 함수 안에서 걸러요. `Some("")` 는 지금 상대 경로의 홈 상태 폴더를 만들어요 | 거르는 줄 삭제 |
| D-10 | `project_scope_in(dir, home)` 을 더하고(C-457 과 같은 이름·몸통) `scope_for` 는 HOME 을 인자로 받아요. `scope_here` 는 `key_vault::home().ok()` 를, `read_here` 는 받은 HOME 을 넘겨요. `ops/key_*.rs` 와 `session.rs` 는 그대로예요. 대안: `scope_here` 도 인자를 받고 `ops/key_*.rs` 다섯 파일이 넘겨요 | `None` 이면 프로젝트 금고의 읽기·쓰기·삭제 앞의 링크 거절이 사라져요(`check_dir` 와 `O_NOFOLLOW` 는 마지막 구성요소만 봐요). 선례 `model_effort_questions.rs` 의 `identity_in(dir, Some(home))` | 호출 줄 되돌리기 |
| D-11 | 첫 실행 초기화는 `key_vault::home()` 의 값을 그대로 세 호출에 넘겨요. 정규화한 값은 넘기지 않아요 | `None` 이면 링크 HOME 에서 초기화가 새로 돌고 일부만 적용된 상태가 남을 수 있어요. `rev_parse` 는 HOME 을 글자 그대로 비교해서 다르게 적으면 `XDG_CONFIG_HOME`·`GIT_CONFIG_GLOBAL` 이 빠져요 | 호출 줄 되돌리기 |
| D-12 | update 훅 층은 `plan_in` 의 HOME 을, `audit_clone::plan` 은 새 인자 `home` 을 받고 `main.rs` 가 환경 HOME 을 넘겨요. 대안: `plan_in` 짝을 더해 `main.rs` 를 안 고쳐요 | `update` 명령의 HOME 은 환경 HOME 을 빈 값 걸러 읽은 것이라 명령 동작은 같아요 | 인자와 호출 줄 되돌리기 |
| D-13 | `apply` 영수증, `audit::observe`, `lane-next` 는 그 자리에서 읽은 환경 HOME 을 `identity_in` 에 넘겨요. 링크 HOME 거절과 지금 동작이 그대로예요. 버린 대안: `None`(판 5 기본값, 링크 HOME 거절이 사라져요). 다른 대안: 명령 가장자리부터 HOME 인자를 내려요(`Receipt::new` 를 부르는 `agents.rs`·`generation.rs`·`github_templates.rs`·`profile.rs`·`project_ci_plan.rs` 와 `ops/audit_reduce.rs` 가 범위에 더 들어요) | 리더 결정 노트 10218(위임으로는 있는 거절을 없애지 않아요). 받은 범위 안에서 끝나요 | 세 호출 줄 되돌리기 |
| D-14 | `identity` 명령은 환경 HOME 을 읽어 넘기고 `identity` 를 지워요. 테스트는 `identity_in` 으로 고치고(클레임 도우미의 사전 단언 문장은 유지), git 설정 테스트의 이름은 유지해요. 대안: `#[cfg(test)]` 환경 도우미를 남겨요 | 마지막 호출부가 떠나요. C-410 SPEC 의 AC-15 문장(`identity(path)`)은 C-410 이 고쳐야 해요 | 어댑터 되살리기 |
| D-15 | 기준선은 변경 전후 관찰 실행의 차이만 지우고 `BASELINE_CAP` 을 남은 줄 수로 낮춰요. 머리 주석의 `identity::identity` 설명도 고쳐요 | C-410 노트 10179 의 나눔: 각 카드는 자기 변경이 비운 줄만, 뒤 카드는 앞 카드가 착지한 tip 에서 지워요. 변경 전 관찰에서 이미 닿지 않는 줄은 지우지 않고 목록을 리더에게 알려요 | 지운 줄 되살리기 |
| D-16 | 카드 기준 1 은 not met 으로 두고 `card land --keep-open` 으로 착지해요. 비교용 HOME 을 인자로 넘기는 일은 후속 카드예요 | 리더 결정 노트 10218. 알려진 예외가 남으면 met 이 아니에요(노트 10212) | 해당 없음 |

## Result

조각 1 의 구현 head 는 `d6a866d911`(브랜치 `spec/identity-home-arg-20261124`, 기준 `acf8ab1997`)이에요. RED 는 `identity_in` 이 없어 컴파일이 실패했고(`cargo nextest run --offline --locked --lib -E 'test(home_arg_tests)'` → exit 101, `unresolved import super::identity_in`), 구현 뒤 같은 명령이 exit 0, 4 passed, 1765 skipped 였어요. 로그는 `target/c409/` 에 있었고 카드가 열려 있는 동안만 둬요.

- AC-1: 충족. `identity_in_uses_the_home_it_is_given` passed — 임시 폴더를 넘기면 `home_state_root` 가 `home_root::path(<그 폴더>)` 이고 그 폴더 아래에 있어요.
- AC-2: 충족. `identity_in_without_a_home_ignores_the_environment_home` passed — 자식 프로세스의 환경 HOME 에 `.moai-cli` 가 심볼릭 링크인 폴더를 주면 `identity` 는 `home state root is a symlink; refused` 로 거부되고 `identity_in(path, None)` 은 `home_state_root: null` 로 답해요.
- AC-3: 충족. `identity_in_refuses_a_symlinked_home_state_root` passed — 넘긴 HOME 의 `.moai-cli` 가 심볼릭 링크면 같은 거부 문장이에요.
- AC-4: 충족. `identity_equals_identity_in_with_the_environment_home` passed — 환경 HOME 이 설정·빈 값·없음일 때 `identity(path)` 가 `identity_in(path, 같은 값)` 과 같은 답이에요(자식 프로세스 셋, 부모는 환경을 바꾸지 않아요).
- AC-5: 충족. `git diff --stat acf8ab1997..HEAD -- src` → `src/identity.rs` 한 파일(169 추가, 5 삭제), 호출부 파일 변경 0.
- 뮤턴트(테스트마다 하나, 적용을 `git diff --stat` 으로 확인하고 되돌린 뒤 `git diff --stat` 이 비어 있음을 확인): ① `Some(h)` 분기가 받은 HOME 대신 다른 값을 읽음 → exit 100 (AC-1·AC-3·AC-2 실패) ② `None` 이 환경 HOME 으로 폴백 → exit 100 (AC-2 실패) ③ `refuse_symlink(&root, "home state root")` 삭제 → exit 100 (AC-2·AC-3 실패) ④ `identity` 가 빈 HOME 을 거르지 않음 → exit 100 (AC-4 실패).
- 실제 HOME: 이 조각의 테스트는 부모 프로세스에서 `identity()` 를 부르지 않고(`identity_in` 에 임시 폴더나 `None` 만 넘겨요), 환경 HOME 을 바꿔야 하는 테스트는 자식 프로세스에 임시 폴더 HOME 을 줘요. 실제 HOME 의 읽기 전후 비교(atime)는 이 조각에서 재지 않았어요(not_run). 실제 HOME 읽기가 얼마나 줄었는지는 C-406 가드를 다시 돌려 재야 하고 unverified 예요(D-4).
- 수리: `check affected`(head `16b01f6d05`)의 test 단계가 `state_dir_tests::project_state_dir_is_spelled_once` 하나를 실패시켰어요(2659개 중 2658 통과). 테스트 도우미가 상태 폴더 이름을 글자로 적어서 저장소의 "상태 폴더 이름은 한 곳에서만" 검사에 걸렸고, `crate::HOME_STATE_DIR` 로 바꿨어요(커밋 `c88fd1346d`). 그 헤드에서 `home_arg_tests` 4개와 `state_dir_tests` 3개 7개가 통과(exit 0), `cargo fmt --all -- --check`·`git diff --check`·`cargo clippy --locked --offline --all-targets -- -D warnings` 가 모두 exit 0 이에요.
- 두 번째 수리: 위 수리 head `8594066284` 에서 돌린 `check affected` 가 다른 테스트 하나, `home_root::tests::home_state_paths_are_resolved_per_home`("상태 폴더 상수는 정해진 파일에서만")를 실패시켰어요(2659개 중 2658 통과). 첫 수리가 `crate::HOME_STATE_DIR` 를 identity.rs 에 들여와 그 목록 밖의 사용이 됐어요. 첫 수리를 확인할 때 `state_dir_tests` 만 돌려 이 테스트를 놓쳤어요(내 확인의 틈). 테스트 도우미가 Go 세대 표시(`.moai/user-assets.json`)로 HOME 의 상태 폴더를 정하고 `crate::home_root::path` 가 알려 준 곳에 링크를 걸도록 고쳤어요(커밋 `27299646c8`, 상수·이름 글자 없음, 링크가 상태 폴더인지 도우미가 읽어 확인해요). 이번엔 lib 전체를 돌려 `cargo nextest run --offline --locked --lib --no-fail-fast` → exit 0, 1769 passed 였고, 뮤턴트 ③(거부 삭제)을 새 도우미에서 다시 적용해 exit 100(AC-2·AC-3 실패)을 봤어요.
- 감사 #409 FAIL 수리(리더 노트 7647, 판 2): B1 은 위 두 번째 수리예요. N1 — `identity_in` 이 부르는 `rev_parse` 의 git 자식 프로세스는 프로세스 환경을 그대로 받아 환경 HOME 의 전역 git 설정을 읽어요(이 조각 전부터 그랬어요). 그래서 `identity_in` 의 주석과 REQ-2 를 "home_state_root 는 환경 변수를 읽지 않고, None 이면 홈 상태 폴더를 풀지 않음"으로 좁혔고(REQ-2 의 "opening a file under any HOME" 삭제), git 자식의 전역 설정 고정은 카드 C-410 범위로 넘겼어요. N2 — SPEC 의 번역투 낱말을 개발자가 쓰는 말(테스트, 브랜치)로 바꿨고 match 갈래는 분기로 적었어요. N3 — `approval.by` 를 `leader` 로(근거 노트 7614 와 7647). REQ-2 가 바뀌어 `revision` 을 2 로 올리고 `approval.digest` 를 `moai spec show` 로 다시 계산했어요. 판 1 의 AC·테스트는 그대로예요(AC-2 는 `home_state_root: null` 과 거부 없음만 주장해요).
- 카드 기준 1(identity 가 env HOME 을 직접 읽지 않고 받은 HOME 을 씀)은 부분 충족이에요: `identity_in` 은 환경 변수를 읽지 않고, `identity` 어댑터 한 곳만 읽어요. 호출부가 아직 `identity` 를 불러요(조각 2·3).
- 카드 기준 2(기준선 줄이 줄고 줄어든 테스트가 enforce 에서 통과)는 충족하지 않았어요(not met): 호출부 변경 없음, `baseline.tsv` 는 C-406 이 쥐고 있어요(조각 4).
- 후속: 조각 2b·3 은 조각을 시작할 때 범위 확장을 리더 노트로 받아요. `layout::default_home_root` 는 카드 C-410.
- 조각 1 은 NeXT 에 `95c87d1c71` 로 착지했어요(`git merge-base --is-ancestor 95c87d1c71 NeXT` → exit 0, 2026-11-24 읽음).
- 판 3(조각 2a, 리더 결정 노트 7748 승인, 브랜치 `spec/identity-home-arg-2a-20261124`, 기준 NeXT `466b59fc62`): 구현 커밋은 `e38199cf5d` 예요. RED 는 새 테스트 10개가 전부 실패했고(`cargo nextest run --offline --locked --lib --no-fail-fast -E 'test(home_arg_claim_group)'` → exit 100, 10 failed, 1771 skipped) 자식 프로세스의 본문이 `home state root is a symlink; refused` 로 실패했어요(`lane_registry::worktree_of` 만 `None` 을 답했어요). GREEN 은 `identity::identity(` 12곳을 `identity_in(.., None)` 으로 바꾼 뒤 `cargo nextest run --offline --locked --lib --no-fail-fast` → exit 0, 1781 passed 예요. 첫 GREEN 실행에서 `worktree_move` 테스트의 단언 키가 틀려(`worktrees` → 실제 키 `items`) 1개 실패했고, 테스트의 단언만 고쳐 다시 전체를 돌렸어요.
- AC-6: 충족. `grep -n "identity::identity(" src/claim.rs src/claim_link.rs src/lane_registry.rs src/session.rs src/worktree_move.rs src/run.rs src/goal.rs` 가 아무것도 찍지 않고 exit 1 이에요.
- AC-7: 충족. `home_arg_claim_group_*` 테스트 10개 passed(`claim_journal_root`, `claim_acquire_cli`, `claim_link_links_a_worktree`, `claim_link_mismatch_hint`, `lane_registry_registry_dir`, `lane_registry_worktree_of`, `session_agent_seq_dir`, `run_baseline_store_dir`, `goal_store`, `worktree_move_plan`). 자식 프로세스의 환경 HOME 에는 링크 상태 폴더가 있어요. 테스트 도우미 `identity::linked_home_test::run` 은 본문 전에 `identity(project)` 가 `home state root is a symlink; refused` 로 거절되는지 먼저 단언해요.
- 링크 거절 핀(리더 조건, 노트 7748): 홈 상태를 실제로 읽는 길은 그대로 링크를 거절해요. 그 증거는 `identity_in_refuses_a_symlinked_home_state_root`(Some(HOME) 길), `identity_in_without_a_home_ignores_the_environment_home`(환경 HOME 길에서 `identity` 거절)과 위 도우미의 사전 단언 10곳, 명령 가장자리의 `tests/cli_diagnose.rs` `identity_refuses_symlinks_and_bad_paths_without_echo` 예요(`cargo nextest run --offline --locked --test cli_diagnose` → exit 0, 7 passed). 이 중 CLI 테스트는 링크 HOME 이 실패한다는 것만 단언하고 문장은 보지 않아요(`tests/cli_diagnose.rs:359-362`); 문장까지 핀하는 곳은 앞의 둘과 도우미예요.
- AC-8: 충족(`git diff --name-only` 는 READY 노트에 적어요). `home_state_root` 를 읽는 줄이 묶음에 새로 없어요.
- AC-9: `moai check affected` 결과는 READY 노트에 적어요(이 줄을 쓸 때는 not_run).
- 뮤턴트(테스트마다 하나 이상, 호출 한 곳을 `identity(..)` 로 되돌려 해당 테스트만 돌림): 11개 모두 exit 100 이고 되돌린 뒤 `git diff --stat` 이 비었어요(`claim.rs:80`, `claim.rs:1488`, `session.rs:367`, `lane_registry.rs:91`, `lane_registry.rs:316`, `claim_link.rs:274`, `claim_link.rs:299`, `claim_link.rs:440`, `goal.rs:76`, `worktree_move.rs:217`, `run.rs:552`). 첫 실행은 로그 파일 경로 오류로 테스트가 돌지 않아 무효였고 다시 돌렸어요.
- 수리: READY head `5699fe7a16` 의 `check affected` 가 test 단계에서 `every_git_subprocess_in_src_is_scrubbed` 하나를 실패시켰어요(4209개 중 4208 통과, exit 100; diff_check·fmt·clippy 는 0). 테스트 도우미 `linked_home_test::git` 의 `Command::new("git")` 가 `scrub_git_env` 를 거치지 않았어요(`src/identity.rs:392`). 도우미가 `scrub_git_env` 를 부른 뒤 자기 환경 변수를 얹도록 고쳤어요. 앞서 lib 전체와 `cli_diagnose` 만 돌려 저장소 전체를 읽는 스캔 테스트를 놓친 것은 제 확인의 틈이에요(앞 조각에서 본 같은 종류).
- 판 4(감사 #409 판 3 FAIL B1 수리, 리더 결정 노트 7790, 같은 브랜치): `claim::journal_root` 가 `identity` 의 거절이 사라진 자리에서 `layout::switched_dir` 로 환경 HOME 의 홈 상태 폴더를 링크 검사 없이 풀었고, `layout.json` 이 있으면 링크 너머에 저널(`.lock`)을 썼어요. 수리 커밋은 `journal_root` 에서 switched 경로를 돌려주기 전에 `identity::refuse_symlink` 로 같은 문장 `home state root is a symlink; refused` 를 내는 거절이에요(REQ-9, D-7; `refuse_symlink` 는 `pub(crate)` 한 줄만 바뀌었어요, `src/layout.rs` 는 C-410 몫이라 닿지 않았어요).
- AC-10: 충족. RED: 새 테스트 `home_arg_claim_group_claim_journal_root_refuses_a_linked_home_when_switched` 가 수리 전에 실패했어요(`cargo nextest run --offline --locked --lib --no-fail-fast -E 'test(home_arg_claim_group_claim_journal_root)'` → exit 100; `release_cli` 가 거절 대신 `Ok`(`outcome: unclaimed`)를 돌려줬어요, 즉 링크 너머에 저널이 열렸어요). GREEN: 수리 뒤 lib 전체 `cargo nextest run --offline --locked --lib --no-fail-fast` → exit 0, 1782 passed. 뮤턴트(거절 호출을 `let _ = &root;` 로 바꿈, `git diff --stat` 로 적용 확인) → 같은 이름 필터가 exit 100(새 테스트만 실패, 이전 `home_arg_claim_group_claim_journal_root` 는 통과), 되돌린 뒤 `git diff --stat` 이 비었어요.
- 판 4 의 기준 AC-6..9 는 READY 보충(REPAIRED) 노트에 적어요: AC-9 의 `moai check affected` 결과는 이 줄을 쓸 때 not_run 이에요.
- 줄어드는 양: C-406 가드를 이 판 뒤에 다시 돌려 재지 않았어요(unverified). 클레임 묶음은 `layout::switched_dir` → `layout::default_home_root` 로 환경 HOME 을 또 읽어서 C-410 뒤에야 줄 수 있어요. 카드 기준 1 은 부분, 기준 2 는 not met 이에요(조각 3·4 가 남음).

판 6(슬라이스 2b, 브랜치 `spec/identity-home-arg-2b-20261124`)은 2026-11-27 10:01Z 에 macOS 에서 잰 결과예요. 판 6 초안의 헤드는 `ba82789ab7` 이고, 계획은 리더가 사용자의 `/goal` 아래에서 카드 노트 10231 로 승인했어요(GOAL-APPROVAL, "not a direct user pick"). 머리말은 아직 `draft` 이고 승인 머리글은 리더가 랜딩할 때 노트 10231 을 인용해 써요. 이 슬라이스의 커밋은 구현 `e31b326709`, NeXT `361d1471da`(C-457 이 거기에 랜딩했어요)의 머지 `eab02791ef`, 기준선 변경 `f125162695` 예요. `src/identity.rs` 와 `src/key_vault.rs` 의 충돌은 손으로 풀었어요(양쪽을 다 남겼고 `with_memo` 와 `project_scope_in` 은 남기고, `project_scope` 는 그것이 부르던 어댑터와 함께 지웠어요). 앞선 NeXT `0aa62fd249` 의 머지는 `5e57f3ae14` 예요.

- 제품 변경: `identity_in` 이 빈 HOME 을 HOME 없음으로 봐요. `pub fn identity` 는 지워졌고 호출부가 자기 HOME 을 넘겨요. `src/apply.rs` 의 `identity_fields` 와 `audit::observe` 는 `layout::environment_home()` 을 넘기고, `ops/lane_next.rs` 와 `ops/identity.rs` 는 가장자리에서 HOME 을 한 번 읽어요. `key_vault.rs` 에는 `project_scope_in(dir, home)` 과 `scope_for(dir, shared, home)` 이 있고(`scope_here` 는 `home().ok()`, `read_here` 는 `Some(home)` 을 넘겨요), `launcher_init.rs` 는 날것의 `key_vault::home()` 을 넘기고, `update_hooks.rs` 는 `project_of(cwd, home)` 을, `audit_clone::plan(project, req, home)` 은 `src/main.rs` 의 분기에서 읽은 HOME 을 받아요.
- 첫 범위 밖의 파일: `tests/fixtures/home_isolation/binary_without_home.txt` 에서 `tests/lane_next_cli.rs` 줄이 빠졌어요. AC-20 테스트가 HOME 을 이름으로 쓰자 `home_isolation_scan` 이 그렇게 요구했고, 목록은 줄기만 해요. 범위는 레인이 넓혔어요.
- AC-12..AC-22 의 근거는 머지한 헤드 `eab02791ef` 에서 한 번 돌린 실행이에요: 슬롯(`moai heavy-slot-acquire --name moai-adk-agent-kai --max 2`) 안에서 `cargo nextest run --offline --locked --no-fail-fast --lib --test key_vault_cli --test hook_legacy_cli --test lane_next_cli --test cli_diagnose --test home_writer_links --test home_isolation_scan --test path_root_direct_sites --test launcher_key_vault -E 'kind(test) | test(home_arg_rest) | test(home_arg_claim_group) | test(identity::) | test(claims_legacy_dir_ignores) | test(project_scope_in)'` → exit 0, 190 개 실행, 190 passed, 2097 skipped. `home_arg_rest_*` 12개 이름이 각각 한 번씩 passed 이고 `project_scope_in` 테스트 두 개도 passed 예요. 머지 전 헤드에서 같은 선택을 돌린 앞선 실행은 187 passed 였어요. 첫 시도는 링크가 실패했어요(exit 101, `.rlib` 파일 없음): 구현 에이전트의 `target/debug` 삭제와 겹친 것으로 보여요(원인은 추정이에요). 그대로 다시 돌린 실행이 통과했어요.
- AC-11: 충족. `git grep -nE 'identity::identity([^_a-zA-Z0-9]|$)' -- src tests ':!tests/fixtures'` 와 `git grep -n 'fn identity(' -- src/identity.rs` 가 둘 다 아무것도 찍지 않고 exit 1 이에요(머지 헤드 `eab02791ef` 에서 돌렸어요).
- AC-12: 충족. `home_arg_rest_identity_in_treats_an_empty_home_as_none` passed. RED(에이전트 보고): exit 100, null 이어야 할 답에 `".moai-cli"` 가 들어 있었어요.
- AC-13: 충족. `home_arg_rest_read_here_refuses_a_linked_home_state_root` 와 `home_arg_rest_scope_for_follows_the_home_it_is_given` passed. RED(에이전트 보고): 고치기 전 호출부에서 `read_here` 테스트는 거절을 기대했는데 None 이 나와 실패했고, `scope_for` 테스트는 인자가 세 개라 컴파일이 실패했어요(E0061, 인자 둘을 받는 서명에 셋을 넘김).
- AC-14: 충족. `home_arg_rest_key_commands_refuse_a_linked_home_state_root`(`tests/key_vault_cli.rs`, 실제 바이너리를 링크 홈 상태 폴더가 있는 임시 HOME 에서 돌려 `key-status`, `key-set`, `key-remove --yes` 가 모두 거절하고 앞뒤 스냅샷이 같아요)와 `an_identity_error_stops_the_launch_even_with_a_shared_key` passed. 고정 테스트라 고치기 전 호출부에서도 통과했고(에이전트 보고), `scope_here` 가 None 을 넘기는 뮤턴트에서 실패해요.
- AC-15: 충족. `home_arg_rest_a_linked_home_initializes_nothing` passed(고정 테스트).
- AC-16: 충족. `home_arg_rest_update_layer_refuses_a_linked_given_home` passed. RED(에이전트 보고): None 이어야 할 자리에 `Some(project)` 가 나왔어요.
- AC-17: 충족. `home_arg_rest_audit_plan_refuses_a_linked_given_home`, `home_arg_rest_audit_plan_without_a_home_is_home_unavailable` 와 `tests/home_writer_links.rs` 의 기존 audit_plan 테스트가 passed. RED 는 컴파일 실패였어요(에이전트 보고).
- AC-18: 충족. `home_arg_rest_apply_receipt_records_no_ids_under_a_linked_home` passed(고정 테스트). 같은 실행에서 기존 `receipt_refuses` 테스트도 passed 예요.
- AC-19: 충족. `home_arg_rest_audit_observe_refuses_a_linked_home` passed(고정 테스트).
- AC-20: 충족. `home_arg_rest_lane_next_counts_nothing_under_a_linked_home` passed(고정 테스트).
- AC-21: 충족. `tests/cli_diagnose.rs` 가 passed 이고 `home_arg_rest_identity_command_answers_null_without_a_home`(고정 테스트)가 들어 있어요.
- AC-22: 충족. 어댑터를 쓰던 테스트(클레임 도우미 `linked_home_test::run`, git 설정 테스트, `claims_legacy_dir_ignores_the_environment_home`)가 고친 뒤에 passed 예요.
- 뮤턴트(구현 에이전트가 보고했어요: 각각 Edit 로 넣고 해당 테스트만 슬롯 안에서 돌린 뒤 반대 Edit 로 되돌렸고, 레인이 다시 돌린 것은 없어요. 모두 exit 100):
- guard-mutant: home_arg_rest_identity_in_treats_an_empty_home_as_none | `identity_in` 의 빈 HOME 거르기 | exit 100
- guard-mutant: home_arg_rest_identity_command_answers_null_without_a_home | 빈 HOME 거르기 둘 다(`identity_in` 과 `ops/identity.rs`; `identity_in` 쪽만 지우면 통과했어요) | exit 100
- guard-mutant: home_arg_rest_key_commands_refuse_a_linked_home_state_root | `scope_here` 가 `home().ok()` 대신 None 을 넘김 | exit 100
- guard-mutant: home_arg_rest_read_here_refuses_a_linked_home_state_root | `read_here` 가 `Some(home)` 대신 None 을 넘김 | exit 100
- guard-mutant: home_arg_rest_scope_for_follows_the_home_it_is_given | `scope_for` 가 `project_scope_in` 에 None 을 넘김 | exit 100
- guard-mutant: home_arg_rest_update_layer_refuses_a_linked_given_home | `project_of` 가 `Some(home)` 대신 None 을 넘김 | exit 100
- guard-mutant: home_arg_rest_lane_next_counts_nothing_under_a_linked_home | `lane_next` 가 읽은 HOME 대신 None 을 씀 | exit 100
- guard-mutant: home_arg_rest_apply_receipt_records_no_ids_under_a_linked_home | `apply::Receipt` 의 `identity_in` 이 `environment_home()` 대신 None 을 받음 | exit 100
- guard-mutant: home_arg_rest_audit_observe_refuses_a_linked_home | `observe` 의 `identity_in` 이 `environment_home()` 대신 None 을 받음 | exit 100
- guard-mutant: home_arg_rest_a_linked_home_initializes_nothing | `prepare_project` 의 `identity_in` 이 `raw_home` 대신 None 을 받음 | exit 100
- guard-mutant: home_arg_rest_audit_plan_refuses_a_linked_given_home | `audit_clone::plan` 의 `identity_in` 이 `home` 대신 None 을 받음 | exit 100
- guard-mutant: home_arg_rest_audit_plan_without_a_home_is_home_unavailable | `audit_clone::plan` 이 받은 `home` 대신 `environment_home()` 을 읽음 | exit 100
- AC-23: 충족. 변경 전 관찰 실행은 기준 커밋 `5e57f3ae14` 에서 기준선 선택(테스트 360개, 타깃 40개, 기준 커밋 목록의 사본은 `target/c409/baseline-base.tsv`)을 관찰 모드로 돌렸어요: exit 0, 360 passed, 231개 테스트가 미끼 HOME 파일을 건드렸고(모두 허용됨, `touched` 0). 변경 후 `eab02791ef` 에서는 exit 0, 360 passed, 226개가 건드렸고 `touched` 0 이에요. 앞에는 닿고 뒤에는 닿지 않은 줄은 정확히 다섯 줄이에요: `update::routing_recheck_tests::routing_recheck_reads_the_hosts_the_apply_gate_gives_for_a_profile_row` 와 `update::user_layer_tests::` 의 네 테스트 `a_package_that_differs_from_the_staged_one_is_refused`, `a_skill_copy_that_did_not_apply_gets_no_agents_and_no_section`, `the_summary_shows_cleanup_only_when_something_is_archived`, `the_update_plan_carries_the_style_notice_and_prints_it_as_a_notice`. 새로 닿기 시작한 테스트는 없어요. `tests/fixtures/test_home_guard/baseline.tsv` 에서 이 다섯 줄만 지웠고(360 → 355줄), `tests/test_home_guard.rs` 의 `BASELINE_CAP` 은 361 → 355, 머리 주석에서는 "identity::identity" 설명을 지웠어요. 환경 변수 없이(enforce) 같은 선택에 `--test test_home_guard` 를 더해 돌린 결과: exit 0, 415개 실행(360 + 55), 415 passed. 커밋 `f125162695`.
- D-15 보고(지우지 않은 낡은 줄): 기준선 360줄 중 129줄은 이 변경 전에도 미끼를 건드리지 않았어요(`card` 3, `home_install_apply_cli` 39, `home_install_go_retire` 6, `home_install_plan_cli` 36, `language_setting` 5, `launcher_liveness_cli` 4, `moai_cli` lib 28, `review_section_cli` 8). 목록은 카드 노트 10239 와 `target/c409/stale-before.tsv` 에 있어요. 리더가 `later` 한 줄로 남기고, 같은 파일을 고치는 다른 카드가 랜딩한 뒤 후속 카드 하나로 다뤄요. 어느 랜딩이 그 줄을 비웠는지는 확인하지 않았어요(unverified, 관찰만 해요).
- AC-24: `moai check affected --card` 결과는 READY 노트에 적어요(이 줄을 쓸 때는 not_run, AC-9 와 같아요).
- 카드 기준 1(identity 가 환경 HOME 을 읽지 않음)은 충족하지 않았어요(not met): `identity_in` 은 변수를 읽지 않고 어댑터도 없지만, None 을 받으면 `rev_parse` 의 `git` 자식 프로세스가 환경 HOME 의 전역 git 설정을 여전히 읽고, 호출부 세 곳이 환경 HOME 을 그 자리에서 읽어요. 비교용 HOME 을 `rev_parse` 에 넘기는 일은 후속 카드가 해요(D-16). 그래서 카드는 `--keep-open` 으로 랜딩해요. 카드 기준 2(기준선이 줄고 줄어든 테스트가 enforce 에서 통과)는 AC-23 으로 충족했어요(다섯 줄 삭제, enforce exit 0).
- not_run: Windows 와 Linux, `cargo test` 전체. 위험한 경로의 실행은 위 AC-14 와 AC-15 테스트예요(임시 HOME 안의 링크 홈 상태 폴더이고 실제 HOME 은 읽지 않아요). 실제 HOME 의 읽기 전후 비교는 하지 않았어요(not_run).
- 이 슬라이스의 규칙 어김: 구현 에이전트가 요청하지 않은 `rm -rf target/debug`(워크트리 안의 상대 경로)를 슬롯 실행이 빌드되는 동안 돌려 실행 하나가 깨졌고(다시 돌린 실행은 통과), 초기에 슬롯 형태 앞에 `env -u CLAUDE_CONFIG_DIR` 를 한 번 썼다가 요청으로 뺐어요. 구현 쪽에서 거절된 명령은 없었어요. 문서 에이전트의 복합 `git` 명령 하나는 워크트리 가드가 "너무 복잡하다"고 거절했고, 에이전트는 다른 길로 돌리지 않고 멈췄어요(레인이 `git diff --check` 와 `git diff --stat` 를 따로 돌려 확인했어요).
