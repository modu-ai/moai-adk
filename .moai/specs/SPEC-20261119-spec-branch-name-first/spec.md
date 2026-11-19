---
id: SPEC-20261119-spec-branch-name-first
title: "A SPEC work branch is named spec/<slug>-<yyyymmdd>, and every reader still resolves the old date-first form"
status: draft
revision: 2
created: 2026-11-19
---

# A SPEC work branch is named spec/<slug>-<yyyymmdd>, and every reader still resolves the old date-first form

## Intent

카드 C-112(사용자 결정 2026-11-19, 세 선택지 중 1번): SPEC 작업 가지 이름을 `spec/<yyyymmdd>-<slug>` 에서 `spec/<slug>-<yyyymmdd>`(이름이 앞)로 바꿔요. 레인 상태 줄이 가지 이름을 잘라 보여 주면 열두 줄이 모두 `spec/20261119-spe…` 로 보여 구분이 안 돼요. 이름이 앞이면 잘려도 구분되고 탭 완성도 쉬워요. SPEC 폴더와 id 는 그대로 `SPEC-<yyyymmdd>-<slug>` 예요(폴더가 날짜순으로 정렬돼요).

지금 가지 이름을 읽는 곳은 세 군데(tip `d25ecb1ba1`, `git grep` 으로 읽음)예요: `src/spec_take.rs:130` 이 id 에서 가지를 만들고, `src/tdd.rs:376` 이 가지를 날짜와 이름으로 쪼개 SPEC 을 찾고, `src/hook_guards.rs:1051` 이 SPEC 작업 가지를 알아봐요. 원격에 SPEC 가지가 이미 120개 있고(`git branch -r --list 'origin/spec/*'`), 그중 9개는 수리 가지 `-r2`·`-r3` 예요. 그래서 옛 이름도 계속 읽어야 해요. 이 조각은 가지 이름을 읽는 하나의 해석기를 만들어 세 군데가 같은 규칙을 쓰게 하고, `spec take` 가 새 이름으로 만들게 하고, 스킬·문서가 새 이름을 보이게 해요. 해석이 두 SPEC 으로 읽히는 이름은 추측하지 않고 이유와 함께 거절해요.

이 SPEC 은 계획과 RED 시험까지예요(리더 배차 note 2851). 구현은 `src/hook_guards.rs` 를 C-86 S4(묶음 5)·S3 가 바꾸므로 묶음 5 가 병합된 새 tip 에서 시작해요.

## Interview

- Q1: 가지 이름은 어느 꼴로 하나요? → A: `spec/<slug>-<yyyymmdd>` (사용자 결정 2026-11-19, 카드 C-112 본문, 세 선택지 중 1번). SPEC id·폴더·`Moai-Spec` 트레일러는 그대로예요.
- Q2: 이미 올라간 옛 이름 가지는요? → A: 계속 읽어요. 이름을 바꾸는 도구나 옮기기는 이 조각에 없어요(120개 중 병합된 것은 저절로 사라져요).
- Q3: `-r2` 같은 수리 접미사는요? → A: 두 꼴 모두에서 접미사를 떼고 같은 SPEC 으로 읽어요. 지금은 옛 꼴의 `-r2` 가 SPEC 이름의 일부로 읽혀 존재하지 않는 `SPEC-…-r2` 를 찾아 SPEC-first hook 이 막혀요(S4 r2 때 겪은 일).
- Q4: 두 SPEC 으로 읽히는 이름은요? → A: 거절해요(카드: "추측하지 않고 거절"). 읽는 쪽(hook·tdd)은 두 후보 중 SPEC 폴더가 하나만 있으면 그 폴더를 증거로 쓰고, 둘 다 있거나 둘 다 없으면 거절해요. `spec take` 는 새 가지를 만들기 전에 폴더를 보지 않고 곧바로 거절해요(나중에 다른 SPEC 이 생겨도 가지가 모호해지지 않게).
- Q5: 구현은 언제 시작하나요? → A: 이 조각은 SPEC·계획·RED 시험까지고, 구현 커밋은 묶음 5(C-86 S4)가 병합된 새 tip 에서 시작해요(리더 note 2851).

## Non-goals

- 가지 이름을 바꾸는 도구, 옛 가지의 일괄 이름 바꾸기, 원격 가지 삭제.
- SPEC id·SPEC 폴더 이름·`Moai-Spec`/`Moai-Session` 트레일러의 꼴 바꾸기(그대로예요).
- `moai trailer-check` 의 판정 바꾸기: `spec/` 접두사만 보므로 두 꼴 모두 이미 SPEC 가지로 읽어요. 이 조각은 그 사실을 시험으로 고정만 해요.
- 날짜 검증 강화: 날짜는 지금처럼 숫자 8자리만 봐요(달력상 유효한지는 보지 않아요).
- `spec/` 가 아닌 가지(`card/`, `issue/` 등)와 SPEC 폴더 이름이 아닌 이름은 다루지 않아요.
- 리더의 배차 문구가 저장소 밖에 있는 경우의 수정(저장소 안 스킬·문서만 다뤄요).

## Requirements

- REQ-1: When `moai spec take <SPEC-ID>` starts work on a SPEC of which no branch exists in either form, the command shall create the branch `spec/<slug>-<yyyymmdd>` from the base, switch to it, and name that branch in its result and in the commands it prints.
- REQ-2 (REQ-1): IF a branch of the SPEC already exists in exactly one form, THEN `moai spec take` shall switch to that branch and create no second one.
- REQ-3: The shared branch reader of the SPEC-first hook and `moai tdd` and `moai spec take` shall read `spec/<slug>-<yyyymmdd>` and `spec/<yyyymmdd>-<slug>` and either of them followed by a repair suffix `-r<N>` as one SPEC id `SPEC-<yyyymmdd>-<slug>`.
- REQ-4 (REQ-3): IF a branch name has no SPEC id in either form (a character outside `a-z`, `0-9` and `-`, a date of other than eight digits, nothing before or after the date, or a further `/`), THEN the readers shall treat the branch as not a SPEC work branch.
- REQ-5 (REQ-3): IF a branch name reads as two SPEC ids, THEN the SPEC-first hook and `moai tdd` shall refuse with a reason that says the name reads as two SPEC ids, unless exactly one of the two SPEC folders exists under `.moai/specs/`, in which case that SPEC is the one.
- REQ-6 (REQ-1): IF the branch `moai spec take` would create reads as two SPEC ids, or the SPEC has a branch in both forms, THEN the command shall refuse before it changes anything and name both ids or both branches.
- REQ-7: The skills and the operation pages shall show the work branch as `spec/<slug>-<yyyymmdd>` and say that the date-first form is still read.
- REQ-8: The shipped SPEC-first hook reason shall name the new form `spec/<slug>-<yyyymmdd>` without echoing any part of the branch name.

## Acceptance

- AC-1 (REQ-1, REQ-2): `spec take` 가 새 이름 `spec/login-fix-20261117` 을 만들어 옮기고 결과와 다음 명령(`git push -u origin …`, `--head …`)에 그 이름을 쓰며 옛 이름 가지는 만들지 않아요. 옛 이름 가지가 이미 있으면 그리로 옮기기만 하고 새 가지를 더 만들지 않아요.
  verify: `cargo test --locked --offline --test spec_branch_name_first name_first_take`
- AC-2 (REQ-3, REQ-4, REQ-8): SPEC-first hook 이 `spec/demo-20261117`·`spec/20261117-demo`·각각의 `-r2` 를 모두 `SPEC-20261117-demo` 로 읽어, SPEC 이 없으면 코드 쓰기를 거절하고(이유는 고정 문구, 새 이름을 보임) 있으면 통과시켜요. SPEC 가지가 아닌 이름은 그냥 둬요.
  verify: `cargo test --locked --offline --test spec_branch_name_first name_first_hook`
- AC-3 (REQ-3): `moai tdd red` 가 위 네 이름에서 같은 SPEC id 를 기록해요.
  verify: `cargo test --locked --offline --test spec_branch_name_first name_first_tdd`
- AC-4 (REQ-3): `moai trailer-check` 가 두 꼴과 수리 접미사 이름을 SPEC 가지로 읽고(트레일러가 없으면 실패), 다른 가지는 그렇게 읽지 않아요(지금도 참이라 고정만 해요).
  verify: `cargo test --locked --offline --test spec_branch_name_first name_first_trailer`
- AC-5 (REQ-5, REQ-6): 두 SPEC 으로 읽히는 이름은 hook·tdd 가 이유와 함께 거절하고(폴더 하나는 증거, 둘은 거절), `spec take` 는 두 꼴 가지가 함께 있거나 만들 이름이 두 SPEC 으로 읽히면 아무것도 바꾸지 않고 거절해요.
  verify: `cargo test --locked --offline --test spec_branch_name_first name_first_ambiguous`
- AC-6 (REQ-7): `skills/moai/references/internal/sdd.md` 와 `interview.md` 가 새 이름을 보여요.
  verify: `cargo test --locked --offline --test spec_branch_name_first name_first_skills`
- AC-7 (REQ-1…REQ-8): 이웃 시험은 새 이름 기준으로 고쳐져 통과하고 해석기의 단위 시험이 통과해요.
  verify: `cargo test --locked --offline --test spec_take_cli` `cargo test --locked --offline --test spec_before_code_hook` `cargo test --locked --offline --test green_test_lock` `cargo test --locked --offline --lib spec_branch`
- AC-8 (REQ-7): `docs/design-review/appendix/operations/spec.md`·`tdd.md` 가 새 이름을 보여요.
  verify: manual: 문서 두 쪽이라 시험으로 고정하지 않아요(시험이 `docs/` 경로를 읽으면 `tools/ci_scope.py` 의 거부 목록에 더해야 해요). 감사가 `git grep -n 'spec/<' -- docs/design-review/appendix/operations` 로 읽어요.

## Scope

- write(이 조각, 계획과 RED): `.moai/specs/SPEC-20261119-spec-branch-name-first/**`, `tests/spec_branch_name_first.rs`
- write(구현 조각, 묶음 5 병합 뒤 범위를 새로 받아요): `src/spec_branch.rs`(새 해석기; `src/lib.rs` 를 다른 카드 C-119 가 잡고 있어서 `src/spec_take.rs` 안에서 `#[path]` 로 등록해요, D-8), `src/hook_guards.rs`, `src/tdd.rs`, `src/spec_take.rs`, `src/spec_close.rs`(합칠 가지를 두 꼴에서 찾아요, D-9), `src/ops/spec.rs`, `src/ops/tdd.rs`, `tests/spec_take_cli.rs`, `tests/spec_before_code_hook.rs`, `tests/green_test_lock.rs`, `tests/trailer_check_cli.rs`, `tests/spec_close_cli.rs`, `tests/commit_guard_users.rs`, `tests/sdd_guidance.rs`, `tests/agent_defs_package.rs`, `skills/moai/references/internal/sdd.md`, `skills/moai/references/internal/interview.md`(+ 원본 목록 `skills/moai/references/source-manifest.json` 을 C-59 가 놓은 뒤), `docs/design-review/appendix/operations/spec.md`, `docs/design-review/appendix/operations/tdd.md`
- read-only: `src/trailers.rs`(`spec/` 접두사만 봐요), `src/spec_doc.rs`(`valid_slug`), `src/card/spec_link.rs`, `docs/decisions-2026-11-08.md`(옛 결정 기록이라 고치지 않아요)

## Plan

1. (이 조각) SPEC 과 RED 시험: `tests/spec_branch_name_first.rs` 시험 11개를 먼저 쓰고, 지금 tip 에서 어느 것이 왜 실패하는지 기록해요(RED 커밋에 `Moai-Phase: red`). 가지 `spec/20261119-spec-branch-name-first` 에 push 만 해요(M-Flow 로컬 묶음).
2. 묶음 5(C-86 S4)가 병합되면 새 tip 을 이 가지에 합쳐요(일반 병합). `hook_guards.rs` 의 `spec_branch_id`·S3 의 `ac_work` 가 같은 해석기를 쓰는지 읽어서 확인해요.
3. `src/spec_branch.rs`: 해석기 `resolve(branch, specs_dir) -> NotSpec | One(id) | Two(a, b)` 와 만들 이름을 주는 `names_of(id) -> (name_first, date_first)` 와 단위 시험. 접미사는 끝의 `-r<숫자>` 하나만 떼고, 후보는 (이름 그대로, 접미사 뗀 것) × (이름 앞, 날짜 앞) 중 SPEC 이름 모양에 맞는 것이에요.
4. `hook_guards::spec_branch_id`·`tdd::spec_of_branch`·`spec_take::branch_of` 를 해석기로 바꾸고, `spec take` 가 두 꼴을 찾는 순서(둘 다 있으면 거절, 하나면 그리로, 없으면 새 이름 생성)를 구현해요.
5. 이웃 시험의 옛 이름 고정을 새 이름 기준으로 고치고 옛 이름 시험은 "여전히 읽는다"는 시험으로 남겨요.
6. 스킬(`sdd.md`·`interview.md`)과 문서 두 쪽을 고쳐요. `skills/moai/**` 는 C-59 의 원본 목록 예약이 풀린 뒤 해요.

## Risks

- `src/hook_guards.rs` 를 C-86 S4(묶음 5)·S3(감사 중)가 바꿔요: 구현은 묶음 5 병합 뒤 새 tip 에서 시작하고, S3 의 `ac_work` 가 `spec_branch_id` 를 거치면 새 해석기를 그대로 써요(구현 때 읽어 확인해요).
- `skills/moai/**` 와 `source-manifest.json` 은 C-59 가 쥐고 있어요: 스킬 수정은 그 예약이 풀린 뒤 범위를 새로 받아 해요.
- 폴더 증거는 작업 폴더에 있는 SPEC 폴더만 봐요: 다른 가지의 SPEC 은 병합 전엔 안 보여요. 그래서 두 후보 중 하나만 보이면 그것으로 정해요. 나중에 다른 폴더가 합쳐져 둘 다 보이면 그 가지는 거절돼요(가지 이름을 바꾸세요: `git branch -m`).
- `-r<숫자>` 로 끝나는 SPEC 슬러그는 접미사로도 읽혀 두 후보가 돼요: 지금 SPEC 폴더 86개에 그런 것은 없어요(`ls .moai/specs | grep -E -e '-r[0-9]+$'` 가 비었어요).
- 시험이 `docs/` 경로를 따옴표로 읽으면 `tools/ci_scope.py` 거부 목록에 더해야 해요: 이 조각의 시험은 `skills/` 만 읽고 `docs/` 는 읽지 않아요.
- 새 시험 파일은 `cfg(unix)` 예요(`sh` 로 시험 스크립트를 돌려요): Windows 빌드는 이 파일을 건너뛰고, `tests/windows_unix_gate_scan.rs` 는 `os::unix` 를 이 파일에서 찾지 않아요.
- 120개 옛 가지는 이름이 그대로라 새 이름과 한동안 섞여요: 읽는 쪽이 두 꼴을 영원히 읽어요(Non-goals).

## Decisions

- D-1: 가지 이름은 `spec/<slug>-<yyyymmdd>` 예요(사용자 결정 2026-11-19, 세 선택지 중 1번). 되돌릴 수 있고 범주는 none 이에요.
- D-2: 날짜는 숫자 8자리만 봐요(달력 검사 없음). 지금 hook 이 `spec/20269999-x` 를 SPEC 가지로 읽으므로 그 동작을 약하게 만들지 않으려는 거예요.
- D-3: 수리 접미사는 이름 끝의 `-r<숫자>` 하나예요(숫자는 한 자리 이상). 접미사를 떼고 SPEC 이름 모양이 되면 두 후보로 둬요.
- D-4: 읽는 쪽(hook·tdd)은 후보 중 SPEC 폴더가 하나만 있으면 그것으로 정하고, `spec take` 는 만들기 전에 폴더를 보지 않아요(이유: 나중에 생기는 다른 SPEC 이 가지를 모호하게 만들지 않게).
- D-5: 카드 본문 "슬러그가 8자리 숫자로 끝난다"는 옛 꼴의 모호함이고, 새 꼴은 반대로 "슬러그가 날짜로 시작한다"가 모호해요. 둘 다 "두 SPEC 으로 읽힘"이라는 하나의 판정(해석기가 후보를 둘 만들 때)으로 처리해요. 리더가 보기에 카드 뜻과 다르면 note 로 알려 주세요.
- D-6: hook 의 이유 문구는 고정 문자열이고 가지 이름을 되풀이하지 않아요(지금 시험이 고정한 성질). 명령줄 읽는 쪽(`tdd`·`spec take`)은 후보 id 를 이유에 적어요(`a-z0-9-` 로만 이뤄진 것을 확인한 id 만이에요).
- D-7: 이 조각의 가지 이름은 아직 옛 꼴이에요(`spec/20261119-spec-branch-name-first`): 코드가 새 꼴을 알기 전엔 SPEC-first hook 이 새 꼴 가지를 SPEC 가지로 읽지 못해요.
- D-8: 해석기는 `src/spec_branch.rs` 에 있고 `src/spec_take.rs` 안에서 `#[path = "spec_branch.rs"] pub mod spec_branch;` 로 등록해요(호출은 `crate::spec_take::spec_branch::…`). 이유: `src/lib.rs` 를 C-119(bundle-tool)가 예약하고 있어 C-112 가 take 할 수 없었어요. C-119 가 풀리면 `src/lib.rs` 에 `pub mod spec_branch;` 한 줄을 두고 `#[path]` 줄을 지우면 되고 동작은 같아요. 되돌릴 수 있고 범주는 none 이에요(리더에게 note 로 알렸어요).
- D-9: `spec close` 도 합칠 가지를 찾아야 해서 `src/spec_close.rs` 가 범위에 들어왔어요: 옛 꼴 가지만 있으면 그것을, 아니면 새 꼴을 가리켜요(`spec_take::existing_branch_of`). 둘 다 있는 SPEC 은 `spec take` 가 거절하므로 close 는 만나지 않아요.
- D-10: 해석기의 반환은 SPEC 본문의 `Two(a, b)` 대신 `Ambiguous(Vec<String>)` 예요: 수리 접미사가 붙으면 후보가 둘보다 많아질 수 있어서요(이름 그대로·접미사 뗀 것 × 이름 앞·날짜 앞). 판정은 같아요: 후보가 하나면 `One`, 둘 이상이면 SPEC 폴더가 정확히 하나일 때만 그것, 아니면 `Ambiguous`.
- D-11: 슬러그의 길이 상한(`spec_doc::valid_slug` 의 48자)은 해석기가 적용하지 않아요. 지금 hook 은 상한 없이 읽으므로, 상한을 두면 긴 이름의 가지에서 hook 이 조용히 꺼져요(강제력이 약해져요). `moai spec new` 의 상한은 그대로예요.

## Result

구현 조각(리더 배차 note 2851, 묶음 5 병합 `747738af63` 뒤 새 tip 위)의 결과예요. 먼저 RED 조각(`a2db8d8f69`)에서는 같은 시험 11개가 3 통과 8 실패였고, 해석기 단위 시험은 껍데기(stub)로 11개 중 10개가 실패(1 통과)하는 것을 먼저 봤고, 12번째(긴 슬러그)는 추가한 뒤 고치기 전에 11 통과 1 실패를 따로 봤어요(로그는 scratchpad 의 `red-lib.txt`, `red-long.txt`). 아래 줄은 각 AC 의 `verify:` 명령을 하나씩 그대로 돌린 결과예요. `moai spec check` 는 구현 조각 마지막에 다시 돌려요.

- AC-1: 통과. `cargo test --locked --offline --test spec_branch_name_first name_first_take` → exit 0, 2 passed(새 이름 `spec/login-fix-20261117` 을 만들고 결과·다음 명령에 쓰며 옛 이름은 안 만들어요, 옛 이름 가지가 있으면 그리로 옮기기만 해요).
- AC-2: 통과. `… name_first_hook` → exit 0, 2 passed(새·옛 이름과 각각의 `-r2` 가 `SPEC-20261117-demo` 로 읽혀요; 이유는 고정 문구이고 `demo` 를 되풀이하지 않아요; SPEC 가지가 아닌 이름은 그냥 둬요).
- AC-3: 통과. `… name_first_tdd` → exit 0, 1 passed(네 이름 모두 같은 SPEC id 를 기록해요).
- AC-4: 통과(고정하는 시험). `… name_first_trailer` → exit 0, 1 passed.
- AC-5: 통과. `… name_first_ambiguous` → exit 0, 4 passed(hook·tdd 는 두 id 를 이유에 적고 거절하되 폴더 하나는 증거로 받아들이고 둘은 거절해요; `spec take` 는 두 꼴 가지가 함께 있거나 만들 이름이 두 SPEC 으로 읽히면 아무것도 바꾸지 않아요).
- AC-6: 아직 안 했어요(실패). `… name_first_skills` → exit 101, 1 failed: `skills/moai/references/internal/sdd.md`·`interview.md` 는 C-59 가 `skills/moai/**` 와 원본 목록을 예약하고 있어서 이 조각의 범위가 아니에요. C-59 가 풀린 뒤 고쳐요.
- AC-7: 통과. `cargo test --locked --offline --test spec_take_cli` → exit 0, 15 passed(옛 이름을 고정하던 곳을 새 이름으로 고쳤어요: `BRANCH` 와 `moai-base` 설정 키 두 곳). `--test spec_before_code_hook` → exit 0, 9 passed. `--test green_test_lock` → exit 0, 14 passed. `--lib spec_branch` → exit 0, 12 passed(해석기 단위 시험).
- AC-8: 했어요(`manual:`). `docs/design-review/appendix/operations/spec.md`(take·close 두 줄)와 `tdd.md`(가지 한 줄)가 새 이름과 옛 이름 읽기, 두 SPEC 으로 읽히는 이름의 거절을 말해요. 감사가 `git grep -n 'spec/<' -- docs/design-review/appendix/operations` 로 읽어요.

이웃으로 더 돌린 것(이 AC 의 명령은 아니에요): `--test trailer_check_cli` 3, `commit_guard_users` 7, `sdd_guidance` 6, `agent_defs_package` 19, `spec_close_cli` 8, `hooks_cli` 48, `hook_dispatch_cli` 56, `workflow_e2e` 1, `worktree_cli` 23, `audit_admission` 7, `run_verify` 13, `boundary_repro` 16 모두 exit 0. `cargo fmt --all -- --check` exit 0, `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0, `python3 tests/ci_scope_test.py` 40 OK.

변이 시험(5개, 커밋 `2d4df10602` 뒤 작업 폴더에서 하나씩 적용하고 `git checkout` 으로 복원, 마지막에 깨끗함을 확인): 폴더 증거 무시(`--lib spec_branch` 2 실패), 수리 접미사 안 떼기(같은 시험 3 실패), 두 꼴 가지 거절 빼기(`name_first_ambiguous_take` 1 실패), 두 SPEC 으로 읽힘 거절 빼기(같은 시험 1 실패), SPEC 폴더 쓰기 허용 빼기(`name_first_ambiguous_hook` 1 실패) 모두 KILLED.

돌리지 않은 것: AC-6(스킬, 위 이유), 실제 windows 빌드 잡(병합 뒤 CI), 호스트의 실제 세션에서 새 이름 가지로 hook 이 도는지(실제 세션은 아니에요: 시험은 모두 임시 저장소의 실제 바이너리예요).
