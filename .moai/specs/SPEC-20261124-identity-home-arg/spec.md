---
id: SPEC-20261124-identity-home-arg
title: "identity takes the HOME as an argument: slice 1 adds identity_in and keeps one process-edge reader"
status: approved
revision: 2
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7614 (2026-11-23T15:09:12Z, 사용자 위임 2026-11-23T13:57Z, 노트 7613 의 결정 1 (a), 판 1 승인)과 노트 7647 (2026-11-23T15:49:32Z, 감사 #409 FAIL 수리 결정: N1 REQ-2 를 'home_state_root 는 환경 변수를 읽지 않음' 쪽으로 좁힘, N2 낱말, N3 승인자 leader); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:79126f9c8b4653f288c59959da91819980ed9730b332c925b72513c729cbc283"
---

# identity takes the HOME as an argument: slice 1 adds identity_in and keeps one process-edge reader

## Intent

카드 C-409 의 뿌리 수리예요. C-406 관찰 전체 실행(카드 노트 7598)에서 389개 테스트가 HOME 의 `user-assets.json` 을 읽었어요. 원인 한 줄은 `identity::identity`(`src/identity.rs:199`)가 환경 변수 HOME 을 직접 읽고, 그 HOME 으로 `home_root::path` → `state_root::classify_home` → `home_state::home_go_markers`(`src/home_state.rs:62`)를 불러 `user-assets.json` 을 연다는 거예요. 그런데 `identity` 의 답 가운데 HOME 에서 나오는 칸은 `home_state_root` 하나뿐이고, 그 칸을 읽는 제품 코드는 `src/audit_clone.rs:151` 뿐이에요. 나머지 호출부 15개 파일은 쓰지도 않는 칸을 위해 HOME 을 읽고 있어요.

이 SPEC 은 그 읽기를 인자로 바꿔요. 1조각(이 카드의 범위, `src/identity.rs`)은 HOME 을 인자로 받는 `identity_in` 을 더하고 환경 변수를 읽는 곳을 `identity` 한 곳(프로세스 가장자리)으로 모아요. 호출부는 이번 조각에서 건드리지 않아요. 호출부는 묶음별 다음 조각으로 나누고(Plan 의 조각표), 조각마다 범위 확장은 리더 결정 노트로 받아요.

## Interview

- Q1: `identity` 의 서명을 바로 바꾸나요, 새 함수를 더하나요? → A: 새 함수 `identity_in(path, home)` 를 더하고 `identity(path)` 는 그것을 부르는 가장자리 어댑터로 남겨요(D-1). 서명을 바로 바꾸면 한 번에 16개 파일이 깨져 한 카드·한 감사로 닿기엔 너무 커요.
- Q2: `src/home_state.rs` 를 고치나요? → A: 안 고쳐요. `home_go_markers(moai: &Path)` 는 이미 경로를 인자로 받고 환경 변수를 읽지 않아요(코드를 읽어 확인, D-2).
- Q3: HOME 을 환경 변수에서 읽는 곳이 `identity` 뿐인가요? → A: 아니에요. `src/layout.rs:35` `default_home_root` 도 읽고 클레임 장부 경로(`claim::journal_root` → `layout::switched_dir`)가 거쳐요. 이 SPEC 의 범위 밖이고 따로 카드로 열어요(D-3).

## Non-goals

- `layout::default_home_root` 와 `key_vault.rs`·`agents.rs`·`claude.rs` 등 환경 변수 HOME 을 읽는 다른 곳(src 에 40곳 가까이 있어요). 각자 HOME 을 받는지는 별도 판단이에요(D-3).
- `tests/fixtures/test_home_guard/baseline.tsv` 줄 지우기 — C-406 이 그 파일을 쥐고 있어서 C-406 착지 뒤 조각이에요(조각 4).
- 호출부 파일 변경(조각 2·3). 이 판은 `src/identity.rs` 만 써요.
- `identity` 답의 모양이나 `home_state_root` 의 값 규칙 바꾸기.

## Requirements

- REQ-1: The `identity_in` function shall resolve `home_state_root` from the HOME it is given and from no environment variable.
- REQ-2: WHEN `identity_in` is given no HOME, it shall return `home_state_root` as `null` without resolving a home state root. · failure: none — the other fields do not depend on a HOME
- REQ-3: WHEN `identity_in` is given a HOME, `home_state_root` shall be the home state root of that HOME even when the environment HOME is another folder.
- REQ-4 (REQ-3): IF the home state root of a given HOME is a symlink, THEN `identity_in` shall return the error `home state root is a symlink; refused` as `identity` returned it before.
- REQ-5: The `identity` function shall return exactly what `identity_in` returns for the environment HOME read once, an unset or empty HOME counting as no HOME.

## Acceptance

- AC-1 (REQ-1, REQ-3): 환경 변수 HOME 과 다른 임시 폴더를 `identity_in` 에 넘기면 `home_state_root` 가 그 임시 폴더 아래 경로예요. · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_in_uses_the_home_it_is_given)'`
- AC-2 (REQ-2): 환경 변수 HOME 이 `.moai-cli` 가 심볼릭 링크인 폴더인 자식 프로세스에서, `identity_in(path, None)` 은 오류 없이 `home_state_root: null` 을 답하고 같은 환경의 `identity(path)` 는 심볼릭 링크 거부 오류를 답해요(None 이 환경 HOME 을 읽지 않는다는 관찰). · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_in_without_a_home_ignores_the_environment_home)'`
- AC-3 (REQ-4): 임시 HOME 의 `.moai-cli` 가 심볼릭 링크면 `identity_in(path, Some(home))` 이 거부 오류를 답해요. · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_in_refuses_a_symlinked_home_state_root)'`
- AC-4 (REQ-5): 환경 변수 HOME 이 설정·빈 값·없음 세 경우에 `identity(path)` 의 답이 `identity_in(path, 같은 값)` 의 답과 같아요(자식 프로세스에서 환경을 고정). · verify: `cargo nextest run --offline --locked --lib -E 'test(identity_equals_identity_in_with_the_environment_home)'`
- AC-5 (REQ-1..REQ-5): 이 판은 `src/identity.rs` 와 SPEC 폴더만 바꿨고 호출부 파일 diff 는 0이에요. · verify: `git diff --stat <base>..<head> -- src`

## Scope

- write: `src/identity.rs`, `.moai/specs/SPEC-20261124-identity-home-arg/spec.md`
- read-only: `src/home_state.rs`(고치지 않아요, D-2), `src/home_root.rs`, `src/state_root.rs`, 호출부 파일들
- forbidden: 실제 HOME(읽기 전후 비교로도 건드리지 않아요, 임시 HOME 픽스처만), `tests/fixtures/test_home_guard/baseline.tsv`(C-406 이 쥐어요), 네트워크, `CLAUDE.md`

## Plan

조각표(묶음별, 앞 조각이 착지한 뒤 다음 조각이 서요). 조각 2 이후의 범위는 이 SPEC 이 정하지 않고 리더가 노트로 정해요.

| 조각 | 무엇 | 파일 | 동작이 같다는 증거 |
| --- | --- | --- | --- |
| 1 (이 판) | `identity_in(path, home)` 추가, `identity` 는 환경 HOME 한 번 읽어 위임 | `src/identity.rs` | AC-1..4: 같은 입력이면 같은 답 |
| 2 | HOME 칸을 안 쓰는 호출부가 `identity_in(path, None)` 을 불러요 | 2a 클레임 묶음: `src/claim.rs`, `src/claim_link.rs`, `src/lane_registry.rs`, `src/session.rs`, `src/worktree_move.rs`, `src/run.rs`, `src/goal.rs`. 2b 나머지: `src/layout.rs`(identity 호출만), `src/key_vault.rs`, `src/audit.rs`, `src/update_hooks.rs`, `src/apply.rs`, `src/launcher_init.rs`, `src/ops/lane_next.rs` | 이 파일들이 읽는 칸 목록(아래 읽기 증거)에 `home_state_root` 가 없어요; 파일마다 기존 테스트가 그대로 통과 |
| 3 | HOME 칸을 쓰는 곳이 HOME 을 받아요: `audit_clone` 은 부르는 쪽에서 HOME 을 받고, 명령 가장자리(`src/ops/identity.rs`)는 환경 HOME 을 넘겨요. 마지막 호출부가 떠나면 `identity` 어댑터를 지워요 | `src/audit_clone.rs`, `src/ops/identity.rs`, `src/identity.rs` | `moai identity` 출력이 바이트 같아요(`tests/cli_diagnose.rs` 가 `home_state_root` 를 단언) |
| 4 | C-406 착지 뒤 `baseline.tsv` 에서 줄어든 테스트 줄을 지워요 | `tests/fixtures/test_home_guard/baseline.tsv` | 가드가 enforce 에서 지워진 줄의 테스트를 통과시켜요 |

읽기 증거(2026-11-24, `grep -rn "home_state_root" src` 와 호출부 읽기): `home_state_root` 키를 이름으로 읽는 제품 코드는 `src/audit_clone.rs:151` 뿐이에요. 호출부 `claim.rs`, `session.rs`, `lane_registry.rs`, `layout.rs`, `ops/lane_next.rs` 는 `identity_status`, `clone_anchor`, `work_tree`, `workspace_id` 같은 칸만 읽어요. 조각 2 는 파일마다 이 읽기를 다시 해서 증거로 남겨요.

순서(이 판): 1) 결정 D-1..D-4 와 SPEC 승인을 리더에게 노트로 받아요. 2) RED: 위 테스트 이름들로 실패를 먼저 봐요. 3) GREEN: `identity` 본문을 `identity_in` 으로 옮기고 `identity` 는 한 줄 어댑터로 줄여요. 4) 이웃: `identity` 를 거치는 기존 테스트 대표 몇 개, fmt, clippy.

## Risks

- 이 조각만으로는 `baseline.tsv` 줄이 줄지 않아요(호출부가 아직 `identity` 를 불러요). 카드 기준 2 는 조각 2·3 뒤에 충족돼요 — 조각 1 의 READY 는 기준 2 를 not met 으로 적어요.
- 두 번째 뿌리 `layout::default_home_root` 때문에 클레임 묶음 테스트(예: `claim_link_cli` 27개)는 조각 2 뒤에도 HOME 읽기가 남을 수 있어요. 몇 개가 줄지는 C-406 가드를 observe 로 다시 돌려 재야 해요(그 전에는 unverified).
- `identity_in(path, None)` 은 `home_state_root` 를 null 로 만들어요. 그 칸을 쓰는 새 호출부가 생기면 조용히 null 을 받아요 — 조각 3 이 그 칸을 쓰는 곳을 HOME 인자로 못 박아요.
- 자식 프로세스에서 환경을 고정하는 테스트은 프로세스 전체 환경을 바꾸지 않아 병렬 테스트을 깨지 않아요(부모는 `std::env::set_var` 를 쓰지 않아요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `identity` 서명을 바꾸지 않고 `identity_in(path, home: Option<&Path>)` 를 더해요. `identity` 는 환경 HOME 을 한 번 읽어 `identity_in` 에 넘기는 가장자리 어댑터예요 | 호출부 16개 파일·약 21곳을 한 카드로 못 바꿔요(카드 본문 "조각으로 나눠도 돼요"). 더하기만 하는 조각이라 되돌리기 쉬워요 | `identity_in` 을 지우고 본문을 `identity` 로 되돌리기 |
| D-2 | `src/home_state.rs` 는 이 SPEC 에서 안 고쳐요 | `home_go_markers(moai: &Path)` 는 인자로 경로를 받고 환경 변수를 읽지 않아요(`src/home_state.rs:62`). 카드 범위에는 그대로 두고 쓰지 않아요 | 해당 없음 |
| D-3 | 두 번째 뿌리 `layout::default_home_root`(`src/layout.rs:34`)와 환경 HOME 을 읽는 다른 곳은 이 SPEC 이 아니라 후속 카드예요 | 같은 종류가 두 번 보이면 소스를 고쳐요(재발 방지). 범위를 넓히지 않고 리더가 카드를 열어요 | 후속 카드를 닫기 |
| D-4 | 줄어드는 양은 조각마다 C-406 가드(observe)를 다시 돌려 재요. 그 전의 수는 unverified 예요 | 한 조각이 몇 개를 줄일지는 읽기만으로 정할 수 없어요 | 해당 없음 |

## Result

조각 1 의 구현 head 는 `d6a866d911`(브랜치 `spec/identity-home-arg-20261124`, 기준 `acf8ab1997`)이에요. RED 는 `identity_in` 이 없어 컴파일이 실패했고(`cargo nextest run --offline --locked --lib -E 'test(home_arg_tests)'` → exit 101, `unresolved import super::identity_in`), 구현 뒤 같은 명령이 exit 0, 4 passed, 1765 skipped 였어요. 로그는 `target/c409/` 에 있었고 카드가 열려 있는 동안만 둬요.

- AC-1: 충족. `identity_in_uses_the_home_it_is_given` passed — 임시 폴더를 넘기면 `home_state_root` 가 `home_root::path(<그 폴더>)` 이고 그 폴더 아래에 있어요.
- AC-2: 충족. `identity_in_without_a_home_ignores_the_environment_home` passed — 자식 프로세스의 환경 HOME 에 `.moai-cli` 가 심볼릭 링크인 폴더를 주면 `identity` 는 `home state root is a symlink; refused` 로 거부되고 `identity_in(path, None)` 은 `home_state_root: null` 로 답해요.
- AC-3: 충족. `identity_in_refuses_a_symlinked_home_state_root` passed — 넘긴 HOME 의 `.moai-cli` 가 심볼릭 링크면 같은 거부 문장이에요.
- AC-4: 충족. `identity_equals_identity_in_with_the_environment_home` passed — 환경 HOME 이 설정·빈 값·없음일 때 `identity(path)` 가 `identity_in(path, 같은 값)` 과 같은 답이에요(자식 프로세스 셋, 부모는 환경을 바꾸지 않아요).
- AC-5: 충족. `git diff --stat acf8ab1997..HEAD -- src` → `src/identity.rs` 한 파일(169 추가, 5 삭제), 호출부 파일 변경 0.
- 뮤턴트(테스트마다 하나, 적용을 `git diff --stat` 으로 확인하고 되돌린 뒤 `git diff --stat` 이 비어 있음을 확인): ① `Some(h)` 분기가 받은 HOME 대신 다른 값을 읽음 → exit 100 (AC-1·AC-3·AC-2 실패) ② `None` 이 환경 HOME 으로 폴백 → exit 100 (AC-2 실패) ③ `refuse_symlink(&root, "home state root")` 삭제 → exit 100 (AC-2·AC-3 실패) ④ `identity` 가 빈 HOME 을 거르지 않음 → exit 100 (AC-4 실패).
- 실제 HOME: 이 조각의 테스트은 부모 프로세스에서 `identity()` 를 부르지 않고(`identity_in` 에 임시 폴더나 `None` 만 넘겨요), 환경 HOME 을 바꿔야 하는 테스트은 자식 프로세스에 임시 폴더 HOME 을 줘요. 실제 HOME 의 읽기 전후 비교(atime)는 이 조각에서 재지 않았어요(not_run). 실제 HOME 읽기가 얼마나 줄었는지는 C-406 가드를 다시 돌려 재야 하고 unverified 예요(D-4).
- 수리: `check affected`(head `16b01f6d05`)의 test 단계가 `state_dir_tests::project_state_dir_is_spelled_once` 하나를 실패시켰어요(2659개 중 2658 통과). 테스트 도우미가 상태 폴더 이름을 글자로 적어서 저장소의 "상태 폴더 이름은 한 곳에서만" 검사에 걸렸고, `crate::HOME_STATE_DIR` 로 바꿨어요(커밋 `c88fd1346d`). 그 헤드에서 `home_arg_tests` 4개와 `state_dir_tests` 3개 7개가 통과(exit 0), `cargo fmt --all -- --check`·`git diff --check`·`cargo clippy --locked --offline --all-targets -- -D warnings` 가 모두 exit 0 이에요.
- 두 번째 수리: 위 수리 head `8594066284` 에서 돌린 `check affected` 가 다른 테스트 하나, `home_root::tests::home_state_paths_are_resolved_per_home`("상태 폴더 상수는 정해진 파일에서만")를 실패시켰어요(2659개 중 2658 통과). 첫 수리가 `crate::HOME_STATE_DIR` 를 identity.rs 에 들여와 그 목록 밖의 사용이 됐어요. 첫 수리를 확인할 때 `state_dir_tests` 만 돌려 이 테스트을 놓쳤어요(내 확인의 틈). 테스트 도우미가 Go 세대 표시(`.moai/user-assets.json`)로 HOME 의 상태 폴더를 정하고 `crate::home_root::path` 가 알려 준 곳에 링크를 걸도록 고쳤어요(커밋 `27299646c8`, 상수·이름 글자 없음, 링크가 상태 폴더인지 도우미가 읽어 확인해요). 이번엔 lib 전체를 돌려 `cargo nextest run --offline --locked --lib --no-fail-fast` → exit 0, 1769 passed 였고, 변이 ③(거부 삭제)을 새 도우미에서 다시 적용해 exit 100(AC-2·AC-3 실패)을 봤어요.
- 감사 #409 FAIL 수리(리더 노트 7647, 판 2): B1 은 위 두 번째 수리예요. N1 — `identity_in` 이 부르는 `rev_parse` 의 git 자식 프로세스는 프로세스 환경을 그대로 받아 환경 HOME 의 전역 git 설정을 읽어요(이 조각 전부터 그랬어요). 그래서 `identity_in` 의 주석과 REQ-2 를 "home_state_root 는 환경 변수를 읽지 않고, None 이면 홈 상태 폴더를 풀지 않음"으로 좁혔고(REQ-2 의 "opening a file under any HOME" 삭제), git 자식의 전역 설정 고정은 카드 C-410 범위로 넘겼어요. N2 — SPEC 낱말 시험→테스트, 가지→브랜치(match 갈래는 분기). N3 — `approval.by` 를 `leader` 로(근거 노트 7614 와 7647). REQ-2 가 바뀌어 `revision` 을 2 로 올리고 `approval.digest` 를 `moai spec show` 로 다시 계산했어요. 판 1 의 AC·시험은 그대로예요(AC-2 는 `home_state_root: null` 과 거부 없음만 주장해요).
- 카드 기준 1(identity 가 env HOME 을 직접 읽지 않고 받은 HOME 을 씀)은 부분 충족이에요: `identity_in` 은 환경 변수를 읽지 않고, `identity` 어댑터 한 곳만 읽어요. 호출부가 아직 `identity` 를 불러요(조각 2·3).
- 카드 기준 2(기준선 줄이 줄고 줄어든 테스트가 enforce 에서 통과)는 충족하지 않았어요(not met): 호출부 변경 없음, `baseline.tsv` 는 C-406 이 쥐고 있어요(조각 4).
- 후속: 조각 2a·2b·3 은 조각을 시작할 때 범위 확장을 리더 노트로 받아요. `layout::default_home_root` 는 카드 C-410.
