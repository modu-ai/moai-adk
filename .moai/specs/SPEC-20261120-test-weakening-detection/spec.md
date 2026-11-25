---
id: SPEC-20261120-test-weakening-detection
title: "Detect a weakened test in the diff and hand the findings to the auditor"
status: approved
revision: 2
created: 2026-11-20
approval:
  by: leader
  at: 2026-11-25
  basis: "리더 위임 승인 노트 9487 (사용자 위임 2026-11-25 ~14:50Z \"나머진 네가 알아서 추천 해주는데로 선택해서\", 리더 판단이고 사용자 직접 선택은 아님); 리더가 착지 직전 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:8618b6fc5614139212f8c787f9ca3d3947d553205973f41fb16fd98a7a01073f"
---

## Interview

- none: revision 2 came from the lane's plan under the overnight delegation; the leader decided the open choices and no question went to the user.

# Detect a weakened test in the diff and hand the findings to the auditor

## Intent

변경은 자기를 판정하는 검사를 고칠 수 없어야 해요(검증 2차, 2f). 지금은 `src/gate.rs` 의 `tier0_safe` 가 `tests/` 로 시작하는 경로를 모두 안전으로 봐서, 시험만 고친 PR 은 tier 0 이고 tier 0 은 감사가 없어요. 시험을 지우거나 약하게 만든 변경을 잡는 규칙은 `skills/moai/references/internal/quality.md` 의 한 문장뿐이고 그 문장을 읽는 도구가 없어요. 이 SPEC 은 base..head diff 를 읽는 결정적 검사 하나(모델 호출 없음, 네트워크 없음)를 더해요. 잡는 것은 다섯 가지예요: 시험 함수 삭제, 건너뛰기 추가, 바뀐 시험 파일의 단언 줄 순감소, 시험의 sleep 추가, 코드 변경 없이 골든 파일만 다시 쓴 변경. 걸린 `파일:줄` 목록을 `review-package` 결과와 `gate-check` 결과에 실어서 감사자에게 넘기고, 켜면 `effective_tier` 를 최소 1 로 올려 감사를 붙여요. 막는 것이 아니라 감사를 붙이는 것이에요. 새 게이트는 관찰(observe)로 시작해요: 지금은 목록만 보고하고 tier 를 올리지 않으며, 켜는 스위치는 상수 하나이고 리더가 켜요.

## Non-goals

- 모델이 읽고 판단하는 약화 검사: 하지 않아요. 이 검사는 줄 단위 규칙만 써요.
- 약화가 걸린 PR 을 막는 일: 하지 않아요. 감사를 붙일 뿐이고, 합쳐도 되는지는 감사가 정해요.
- `audit-gate` CI 작업의 tier 계산 바꾸기: 이번 조각에서는 하지 않아요(그 작업은 파일 이름만 읽어요). 켜는 조각이 두 곳을 함께 맞춰요(Risks).
- Rust 단위 시험이 소스 파일 안의 `#[cfg(test)]` 모듈에 있을 때 그 모듈의 단언 수를 세는 일: 하지 않아요. 단언 순감소는 시험 파일 경로 규칙(Decisions D-2)으로만 세요.

## Requirements

- REQ-1: When the detector reads a unified diff, it shall report each test function the diff deletes or disables, as a `test_deleted` finding with the old file's path and line, unless the same test name is added again in the same diff. · failure: none
- REQ-2: When the detector reads a unified diff, it shall report each added line that skips a test (`#[ignore`, `@unittest.skip`, `@pytest.mark.skip`, `t.Skip(`, `.skip(`) as a `skip_added` finding with the new file's path and line. · failure: none
- REQ-3: When a changed test file loses more assertion lines than it gains, the detector shall report it as an `assertions_reduced` finding with the counts and the first removed assertion's line. · failure: none
- REQ-4: When the diff adds `thread::sleep` or `time.sleep(` to a test file, the detector shall report each such line as a `sleep_added` finding. · failure: none
- REQ-5: When the diff changes a golden file and no file outside tests, golden files and documents, the detector shall report a `golden_only` finding for each golden file. · failure: none
- REQ-6: When a diff holds no finding of the five kinds, the detector shall return an empty list, so a test-only change that weakens nothing keeps the tier its paths give. · failure: none
- REQ-7: When `review-package` builds a package, the result shall carry a `test_weakening` array with the findings of the package's diff. · failure: none
- REQ-8: When `gate-check` collects a snapshot, the gate's result shall list the findings of the diff from the merge base to the head. · failure: none
- REQ-9 (REQ-8): IF the diff cannot be read, THEN the gate's result shall show the findings as `unknown`, never as an empty list.
- REQ-14: The detector module `src/test_weakening.rs` shall be on both protected-path lists (`gate::TIER2_PREFIXES` and `tools/protected_checks.json`), so a change to it is judged tier 2. · failure: none
- REQ-10: While the weakening gate is in observe, `effective_tier` shall ignore the findings. · failure: none
- REQ-13: While the weakening gate is in observe, the gate's result shall name the tier it would give with the findings. · failure: none
- REQ-12: While the weakening gate is enforcing, `effective_tier` shall be at least 1 when the findings are not empty or are unknown. · failure: none
- REQ-11: When the quality rule speaks of a deleted test, the quality rule shall say in one connecting sentence that a deletion whose reason is written in the card and the SPEC is not a finding. · failure: none

## Acceptance

- AC-1 (REQ-1): 시험 함수를 지우는 diff 는 `test_deleted` 를 `파일:줄` 과 함께 내고, 같은 이름을 다른 파일에 다시 더하는 diff(옮기기)는 내지 않아요(Rust `#[test]`, Python `def test_` 각각 한 쌍의 고정물). 속성만 떼고 같은 파일에 다른 새 시험을 더하는 diff 도 `test_deleted` 를 내요(같은 이름이거나 속성을 다른 속성으로 바꾼 경우만 상쇄) · verify: `cargo test --locked --offline --test test_weakening -- a_deleted_test_fires a_new_test_in_the_same_file`
- AC-2 (REQ-2): 건너뛰기 다섯 표기를 더하는 diff 는 `skip_added` 를 내고, 건너뛰기를 지우거나 설명 글에만 쓰는 diff 는 내지 않아요 · verify: `cargo test --locked --offline --test test_weakening a_skip_added_fires_in_the_five_notations_and_stays_quiet_on_its_pairs`
- AC-3 (REQ-3): 단언이 순감소한 시험 파일만 `assertions_reduced` 를 내고, 단언을 옮기거나 늘린 파일과 시험 파일이 아닌 파일은 내지 않아요 · verify: `cargo test --locked --offline --test test_weakening removed_assertions_fire_and_a_rewritten_or_grown_count_stays_quiet`
- AC-4 (REQ-4): 시험 파일에 sleep 을 더하면 `sleep_added`, 시험 파일이 아닌 곳에 더하면 내지 않아요 · verify: `cargo test --locked --offline --test test_weakening a_sleep_added_to_a_test_fires_and_a_comment_or_source_file_stays_quiet`
- AC-5 (REQ-5): 골든 파일만 바뀐 diff 는 `golden_only`, 같은 골든과 함께 소스 코드도 바뀐 diff 는 내지 않아요 · verify: `cargo test --locked --offline --test test_weakening a_golden_file_changed_alone_fires_and_one_that_follows_code_stays_quiet`
- AC-6 (REQ-6): 시험 파일만 바꾸고 다섯 가지에 걸리지 않는 diff 는 빈 목록이고 `tier_of(Some(0), 그 파일들)` 는 여전히 0 이에요 · verify: `cargo test --locked --offline --test test_weakening a_test_only_change_that_weakens_nothing_stays_tier_zero`
- AC-7 (REQ-7): `review-package` 결과의 `test_weakening` 이 약화 diff 에서는 목록, 아닌 diff 에서는 빈 배열이에요 · verify: `cargo test --locked --offline --test test_weakening review_package_carries_the_findings`
- AC-8 (REQ-8, REQ-9): `gate-check` 의 결과가 걸린 목록을 보이고, diff 를 읽을 수 없으면 `unknown` 을 보여요 · verify: `cargo test --locked --offline --test gate_check_cli the_gate_lists_the_findings_and_says_unknown_when_unread`
- AC-9 (REQ-10, REQ-12, REQ-13): 관찰 중에는 `effective_tier` 가 그대로이고 결과가 올렸을 때의 tier 를 말하며, 켜면 목록이 비어 있지 않거나 unknown 일 때 최소 1 이에요 · verify: `cargo test --locked --offline --test test_weakening observe_reports_and_enforce_raises_the_tier`
- AC-10 (REQ-11): `skills/moai/references/internal/quality.md` 에 연결 문장이 한 번 있고, 그 문장이 이유를 카드와 SPEC 에 적은 삭제는 지적이 아니라고 말해요 · verify: `cargo test --locked --offline --test test_weakening the_quality_rule_carries_the_connecting_sentence`
- AC-12 (REQ-9): 수집기는 8 MiB(8388608바이트) 이하의 diff 는 읽고 한 바이트 더 긴 diff 는 `unknown` 으로 두며, 이름 바꾸기 검출을 끄고 읽어서 시험 파일을 시험 폴더 밖으로 옮긴 diff 는 지운 단언으로 세요 · verify: `cargo test --locked --offline --lib gate_collect::tests`
- AC-13 (REQ-14): `src/test_weakening.rs` 가 `TIER2_PREFIXES` 와 `tools/protected_checks.json` 에 모두 있고 `tier_of(Some(0), …)` 가 2 예요 · verify: `cargo test --locked --offline --test test_weakening the_detector_module_is_a_protected_path_on_both_lists`
- AC-11: NeXT 의 최근 착지 50개(first-parent 병합 커밋마다 첫 부모와 병합 사이의 diff, D-9)에 돌린 결과 — 걸린 수, 그중 실제 약화 수 — 가 이 SPEC 의 Result 에 있어요 · verify: manual: Result 표를 읽어요(명령과 출력이 Result 에 적혀 있어요)

## Scope

- write: 이 SPEC, `src/test_weakening.rs`(새), `src/gate.rs`, `src/gate_collect.rs`, `src/ops/gate_check.rs`, `src/review.rs`, `tests/test_weakening.rs`(새), `tests/gate_check_cli.rs`, `tools/protected_checks.json`, `skills/moai/references/internal/quality.md`, `skills/moai/references/audit.md`(감사 규약의 tier 2 접두사 목록에 `src/test_weakening` 한 단어, 영수증이 짚었어요)
- 이 SPEC 이 쓰는 파일을 다른 카드가 쥐고 있지 않아요(2026-11-26 읽음): `src/review.rs` 와 `quality.md` 를 쥐던 카드는 끝났어요
- read-only: `src/audit_gate.rs`, `tools/protected_checks.json`, `skills/moai/workflows/audit.md`
- forbidden: 모델 호출, 네트워크, 약화가 걸린 PR 을 막는 규칙, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것

## Plan

1. 조각 2(검사기): `src/test_weakening.rs` 순수 함수(diff 글 → 목록)와 다섯 가지 고정물 쌍 시험(AC-1~6). 별도 명령은 만들지 않아요(D-7). 이전 가지에서 끝나 NeXT 끝에 합쳤어요.
2. 조각 3a(소비, gate-check): 스냅샷 키 `test_weakening`, 수집기가 diff 를 읽음, 결과에 목록·`tier_with_findings`, 관찰/켜짐 스위치(AC-8, AC-9). 같은 가지에서 끝나 합쳤어요.
3. 조각 3b(소비, review-package): `review-package` 결과의 새 키 `test_weakening`(AC-7, D-11).
4. 조각 4(문서와 결과): `quality.md` 의 연결 문장(AC-10, D-10), NeXT 의 최근 착지 50개 실행 결과를 Result 에(AC-11, D-9).

## Risks

- **tier 2 변경이에요.** `src/gate*.rs` 는 보호 경로(합칠 PR 의 감사 단계를 정하는 규칙)예요. 관찰로 시작해서 켜기 전까지 tier 는 그대로예요. 켜는 일은 별도 결정이에요.
- `audit-gate` CI 작업은 파일 이름만 읽고 diff 를 읽지 않아서, 켠 뒤에는 `gate-check` 와 `audit-gate` 가 서로 다른 tier 를 낼 수 있어요. 켜는 조각이 두 곳을 함께 맞춰야 해요(이 SPEC 은 관찰만 해서 어긋남이 없어요).
- 줄 규칙은 오탐이 있을 수 있어요(예: 옮긴 시험, 주석의 `assert`). 옮기기는 같은 이름 재추가로 거르고, 최근 PR 50개 결과로 오탐 수를 Result 에 적어요. 오탐은 감사를 한 번 더 붙일 뿐 합치기를 막지 않아요.
- `review-package` 의 결과 모양이 커져서 그 결과를 읽는 시험·문서가 바뀔 수 있어요(조각 3 에서 모두 돌려요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 걸려도 막지 않고 감사자에게 목록을 넘겨요. 켜기 전까지 관찰이에요: 목록과 "올렸다면 tier"만 보고하고 `effective_tier` 는 그대로예요 | 카드 본문("막는 것이 아니라 감사를 붙이는 것")과 공통 규칙("새 게이트는 observe 로 시작")을 함께 지켜요 | 켜는 상수를 true 로 |
| D-2 | 시험 파일은 경로로 정해요: `tests/` 아래, 파일 이름이 `test_*.py`·`*_test.py`·`*_test.rs`. 시험 함수 삭제와 건너뛰기는 경로와 상관없이 줄 표지(`#[test]`, `def test_`, 건너뛰기 표기)로 잡아요 | 소스 파일 안의 시험 모듈은 diff 줄만으로 경계를 알 수 없어요 | 시험 모듈을 읽는 파서로 확장 |
| D-3 | 시험 삭제의 이름은 지워진 `#[test]` 다음 줄의 `fn 이름`(없으면 문맥 줄), Python 은 `def test_이름` 이에요. 같은 이름이 같은 diff 에 시험으로 다시 더해지면 옮기기라서 걸지 않아요 | 파일을 옮기거나 쪼개는 PR 의 오탐을 줄여요 | 옮기기 면제를 걷기 |
| D-4 | 단언 줄은 Rust `assert!(`·`assert_eq!(`·`assert_ne!(`, Python `assert `·`self.assert`·`assertEqual(` 등으로 세고, 바뀐 시험 파일마다 지운 수에서 더한 수를 뺀 값이 양수일 때만 걸어요 | 한 파일 안에서 단언을 옮기는 변경을 거르는 가장 단순한 규칙이에요 | 함수 단위 비교 |
| D-5 | 골든 파일은 경로로 정해요: `/golden/` 이 든 경로, `.golden`·`.snap`·`.expected` 로 끝나는 파일, `tests/fixtures/` 아래. 골든이 바뀌고 시험·골든·문서 밖의 파일이 하나도 안 바뀌면 걸어요 | 코드 변경 없이 기대값만 다시 쓴 경우가 약화의 전형이에요 | 경로 목록 조정 |
| D-6 | 관찰/켜짐 스위치는 `gate.rs` 의 상수 하나(`TEST_WEAKENING_ENFORCED`, 지금 false)예요 | 켜는 일은 리더 결정이라 코드 한 줄로 되돌릴 수 있게 해요 | 상수 값 |
| D-7 | 검사기 모듈은 `src/lib.rs` 가 아니라 `src/gate.rs` 가 `#[path]` 로 선언하고(`gate::test_weakening`), 별도 명령 대신 `gate-check` 결과로만 보여요 | `src/lib.rs`·`src/help.rs` 같은 공용 한 줄 파일을 다른 카드가 오래 쥐고 있어서, 게이트 파일만 건드리게 해요 | `lib.rs` 로 옮기기, 명령 더하기 |
| D-8 | 한 파일에서 지운 시험을 다른 파일에서 같은 이름으로 다시 시작하면 시험 삭제로 세지 않지만, 지운 쪽 파일의 단언 순감소는 파일 단위로 그대로 세요 | 옮긴 시험의 오탐은 감사 한 번이고, 단언을 파일 단위로 세는 가장 단순한 규칙이에요 | 시험 함수 단위로 단언을 세기 |
| D-9 | 카드 기준 3 의 "최근 병합된 PR 50개"를 "NeXT 의 최근 착지 50개"로 읽어요: `git log --first-parent --merges -50 <NeXT 끝>` 의 병합마다 첫 부모와 병합 사이의 diff 에 검사기를 돌려요 | 카드는 PR 시대(2026-11-20)에 쓰였고 지금 흐름(NeXT-Flow)은 PR 없이 로컬 병합으로 착지해서, 같은 뜻의 단위가 착지예요. 리더가 2026-11-26 디스패치에서 이렇게 읽고 SPEC 에 결정으로 적으라고 했어요 | PR 이 다시 생기면 PR 단위로 |
| D-10 | 연결 문장은 `quality.md` 에만 넣어요. 카드가 짝으로 든 `mode-agent.md` 는 더는 없어요: 그 문장("원인 불명이면 삭제")은 `check-rules.md` 로 옮겨 갔고 이미 "이유를 카드와 SPEC 에 적는다"가 들어 있어서, `quality.md` 의 한 문장이 둘을 잇기에 충분해요 | 2026-11-26 에 `skills/moai` 를 읽었어요(`mode-agent.md` 없음, `check-rules.md` 38줄) | `check-rules.md` 에도 같은 문장을 넣기 |
| D-11 | `review-package` 결과의 `test_weakening` 은 `gate-check` 와 같은 글(`kind path:line detail`)의 배열이고, 패키지의 diff 가 8 MiB 를 넘으면 배열은 비우고 `skipped` 에 `test-weakening: diff-over-limit` 한 줄을 남겨요 | 두 결과가 한 가지 모양을 써야 감사자가 같은 방식으로 읽어요. 패키지는 이미 `skipped` 로 못 읽은 것을 말해요 | 구조체 배열로 바꾸기 |

## Result

조각 2·3a(검사기와 `gate-check` 연결)는 이전 가지에서 끝나 NeXT 끝(`882cba7497`) 위에 병합했고, 조각 3b(`review-package` 키)와 4(`quality.md` 문장, 최근 착지 50개 결과)를 이번에 끝냈어요. 아래 변형 시험·감사 수리 표는 이전 가지에서 돌린 것이고(코드는 그 뒤로 `gate_collect.rs` 의 `weakened_in` 가시성 한 단어만 달라졌어요), 이번 가지에서 새로 돌린 변형은 "이번 가지의 변형" 절에 있어요. 관찰/켜짐 스위치(`TEST_WEAKENING_ENFORCED`)는 `false` 로 두었어요: 이 조각이 머지돼도 `effective_tier` 는 달라지지 않고 결과가 목록과 올렸을 때의 tier 를 보이기만 해요. 켜는 일은 리더 결정이에요.

TDD 이탈: 검사기(`src/test_weakening.rs`)를 시험보다 먼저 썼어요. 시험은 그 뒤에 다섯 가지 고정물 쌍으로 쓰고, 변형 10개로 시험이 검사기를 정말 지키는지 확인했어요.

### 변형 시험(커밋한 뒤, 한 번에 하나씩 바꾸고 되돌림)

| 변형 | 바꾼 것 | 잡은 시험 |
| --- | --- | --- |
| M1 | Rust 건너뛰기 표기가 맞지 않음 | collector 시험(`the_collector_lists_weakened_tests_…`) |
| M2 | 옮긴 시험도 삭제로 셈 | `a_deleted_test_fires_and_a_moved_or_swapped_one_stays_quiet` |
| M3 | 단언 수가 같아도 줄었다고 셈 | `removed_assertions_fire_and_a_rewritten_or_grown_count_stays_quiet` |
| M4 | 새 골든 파일도 다시 쓴 것으로 셈 | `a_golden_file_changed_alone_fires_and_one_that_follows_code_stays_quiet` |
| M5 | 골든 옆의 코드 변경을 무시 | 같은 시험 |
| M6 | 시험 파일이 아닌 곳의 단언·sleep 도 셈 | `removed_assertions_fire_…` 와 `a_sleep_added_to_a_test_fires_…` |
| M7 | 관찰 중에도 tier 를 올림 | `observe_reports_and_enforce_raises_the_tier`, `the_gate_lists_the_findings_and_says_unknown_when_unread` |
| M8 | diff 를 못 읽었을 때 tier 를 안 올림 | `observe_reports_and_enforce_raises_the_tier` |
| M9 | git diff 실패를 빈 목록으로 읽음 | collector 시험 |
| M10 | 결과가 못 읽은 diff 를 빈 목록으로 보임 | `observe_reports_and_enforce_raises_the_tier`, `the_gate_lists_the_findings_and_says_unknown_when_unread` |

기준 실행은 세 대상 모두 0 이었고 변형 10개는 모두 시험이 실패해서 잡혔어요(생존 0). M2~M6 은 시험 하나를 종류별 다섯 개로 나누기 전에 돌렸고, 위 표의 시험 이름은 나눈 뒤의 이름이에요(변형은 아래 수리 변형으로 다시 확인했어요).

### 감사 수리(같은 브랜치, 같은 감사자 델타)

감사가 막은 셋과 보호 경로 지적을 이 순서로 고쳤어요. 시험은 고치기 전에 B3 만 RED 로 확인했어요(B1·B2 는 지킬 코드가 이미 있어서 변형으로 확인).

| 감사 | 고친 곳 | 시험 |
| --- | --- | --- |
| B1 8 MiB 경계 | `src/gate_collect.rs` 의 한도 검사를 `weakened_in` 으로 분리 | `the_diff_limit_reads_exactly_eight_mebibytes_and_not_one_byte_more`(한도−1·한도는 읽고 한도+1 은 `None`, 상수 8 MiB) |
| B2 `--no-renames` | 같은 파일 | `a_test_file_moved_out_of_the_test_folders_is_read_as_a_delete_and_an_add` |
| B3 속성만 떼고 새 시험 더하기 | `src/test_weakening.rs` 의 상쇄 수: 이름 없는 표지만 | `a_new_test_in_the_same_file_does_not_offset_a_removed_test_attribute` |
| N5 보호 경로 | `src/gate.rs` `TIER2_PREFIXES`, `tools/protected_checks.json` | `the_detector_module_is_a_protected_path_on_both_lists` |
| N6 고정물 분리 | `tests/test_weakening.rs` 의 큰 시험을 종류별 다섯 개로 | 8개(+1) 모두 통과 |

수리 변형(커밋한 뒤 하나씩, 되돌린 뒤 작업 트리 깨끗함 확인): MA1 한도 `>` 를 `>=` 로 → 잡힘, MA2 한도 검사 없음 → 잡힘, MF 이름 바꾸기 검출 켬(`--no-renames` 삭제) → 잡힘, MB3 새 시험이 하나라도 있으면 상쇄 → 잡힘, MP1 `TIER2_PREFIXES` 에서 모듈 빼기 → 잡힘, MP2 `protected_checks.json` 에서 빼기 → 잡힘. 생존 0.

### NeXT 의 최근 착지 50개에 돌린 결과(AC-11, D-9)

이 절은 이전 가지가 PR 시대의 병합 50개로 돌린 결과를 대신해요. 명령: `git log --first-parent --merges -50 882cba7497`(NeXT 끝)의 병합 50개마다 `git diff --no-ext-diff --no-color --no-renames --unified=0 <병합>^1 <병합>` 을 파일로 저장하고, 커밋한 `src/test_weakening.rs` 에서 serde 파생 한 단어만 뺀 복사본을 `rustc` 로 빌드한 던져 버릴 실행기로 `detect` 를 돌렸어요(PR API 는 쓰지 않았어요, 쓴 도구는 셸과 `rustc` 뿐이고 파일은 `target/` 아래에 두었다가 지웠어요). 출력 맨 끝 줄: `TOTAL diffs=50 with_findings=6 findings=27`.

| 착지한 SPEC | 걸린 것 | diff 와 병합 메시지로 읽은 판단 |
| --- | --- | --- |
| `update-declined-steps` | `golden_only` 1건(`tests/fixtures/state_folder_literals/allow.tsv:198`) | 허용 목록의 개수표 한 줄(70 에서 72)이 같은 병합에서 시험 파일에 더한 줄을 따라갔어요. 시험 파일도 함께 바뀌었고 소스는 없어서 걸렸어요. 사람이 바로 넘길 오탐 |
| `codex-home-update-removed` | `test_deleted` 9건 + `assertions_reduced` 2건(`tests/codex_home_update_removed.rs`, `tests/home_update_notice_follow_ups.rs`, `src/ui_text.rs`) | 같은 병합이 소스(`ui_text.rs`, `update.rs`, `update_tux.rs`)에서 알림 기능을 지워서 그 기능의 시험을 함께 지웠어요. 기능 제거에 따른 의도된 삭제 |
| `home-guard-flake` | `sleep_added` 1건(`tests/check_affected_tmpdir.rs:512`) | 400 번 도는 기다림 반복 안의 50ms 대기(마감이 있어요). 고정 sleep 이 아니라서 오탐 |
| `heavy-slot-release-on-exit` | `sleep_added` 1건(`tests/launcher_auto_init.rs:537`) | 10 초 마감이 있는 반복 안의 50ms 대기. 오탐 |
| `card-no-file-locks` | `test_deleted` 9건 + `assertions_reduced` 3건(`tests/card/` 아래 세 파일) | 파일 잠금을 없애는 변경이라 잠금 양보 시험을 지우고 새 시험 파일(527줄)로 바꿨어요. 의도된 교체 |
| `integration-guard-install` | `sleep_added` 1건(`tests/integration_guard_install.rs:460`) | 1100ms 고정 대기이고 이유가 주석에 있어요(정리 폴더 이름이 시계의 초 단위). 진짜 sleep 이지만 이유가 적혀 있어요 |

결과: 50개 중 6개(12%)가 걸리고 걸린 줄은 모두 27건이에요. 이유 없이 시험을 약하게 만든 변경은 하나도 확인하지 못했어요(diff 와 병합 메시지만 읽었고 작성자에게 이유를 묻지는 않았어요): 실제 약화 0건. 6개 중 사람이 바로 넘길 오탐은 3개(개수표 1, 기다림 반복 안의 sleep 2), 기능 제거에 따른 의도된 삭제·교체가 2개, 이유가 적힌 고정 sleep 이 1개예요. 켜면 50개 중 6개가 tier 1 이 되고 그중 3개는 사람이 보면 바로 넘길 변경이라, 켜기 전 조건에 아래 N5·N6 을 더했어요. 켜는 일은 리더가 이 표를 보고 정해요.

### 알려진 한계

- 이름이 바뀐 시험은 삭제로 보여요(위 두 건). 한 파일에서 지운 시험을 다른 파일에 같은 이름으로 다시 더하면 옮긴 것으로 보지만, 지운 쪽 파일의 단언 순감소는 파일 단위로 그대로 세요(D-8).
- 단언은 파일 단위로 세요. 시험 함수 안에서 단언을 다른 곳으로 옮기면 줄었다고 볼 수 있어요.
- 8MiB 가 넘는 diff 는 읽지 않고 `unknown` 으로 두어요.
- 강제(`TEST_WEAKENING_ENFORCED = true`)로 켜기 전에 풀어야 하는 조건(감사 N1~N4, 관찰 중인 지금은 막지 않아요. 풀기 전에 켜면 정당한 변경이 tier 1 이 돼요):
  - N1: 정당한 파일 이동이 오탐이에요. `git mv` 로 시험 파일을 옮기거나 파이썬 시험을 Rust 로 이식하면 지운 쪽 단언이 줄어든 것으로 세어져요(이름 바꾸기 검출을 끄고 읽으므로, 위 B2 시험이 이 동작을 고정하고 있어요). 이름이 바뀐 대체 시험도 같은 원인이에요(예: 이미 병합된 작업량 카운터 변경이 벽시계 시험 둘을 카운터로 바꾼 의도된 교체가 걸려요).
  - N2: 단언을 지우고 문서 블록(주석·문서 글)에 같은 줄을 쓰면 단언 수가 상쇄돼요.
  - N3: Rust 시험 이름을 바꾸면(속성은 그대로, 함수 이름만 변경) 삭제와 추가로 보여서 이름 바꾸기를 구분하지 못해요.
  - N4: 카드 범위 밖의 약화 형태(단언을 항상 참인 것으로 바꾸기, 시험 본문을 비우기, 기대값 완화 등)는 다섯 가지에 걸리지 않아요.
  - N5: 마감이 있는 기다림 반복 안의 `sleep` 은 약화가 아니에요. 최근 착지 50개에서 sleep 3건 중 2건이 그런 반복이었어요(위 표).
  - N6: 시험 파일의 변경을 따라가는 개수표(예: 허용 목록의 줄 수)를 고친 diff 는 골든만 다시 쓴 변경이 아니에요. 최근 착지 50개에서 `golden_only` 1건이 그랬어요.
  - 이 여섯 가지를 시험이 고정하는 형태로 줄이는 일은 다음 조각이에요. 켜는 일은 리더 결정이고, 그 조각이 끝나기 전에는 켜지 않기를 권해요.

### 수락 기준

- AC-1: met — `--test test_weakening -- a_deleted_test_fires a_new_test_in_the_same_file`(Rust `#[test]`·Python `def test_` 각 한 쌍, 속성만 뗀 diff) 통과
- AC-2: met — `--test test_weakening a_skip_added_fires_in_the_five_notations_and_stays_quiet_on_its_pairs`(건너뛰기 표기 다섯 개와 그 짝) 통과
- AC-3: met — `--test test_weakening removed_assertions_fire_and_a_rewritten_or_grown_count_stays_quiet`(단언 순감소만) 통과
- AC-4: met — `--test test_weakening a_sleep_added_to_a_test_fires_and_a_comment_or_source_file_stays_quiet`(시험 파일의 sleep 만) 통과
- AC-5: met — `--test test_weakening a_golden_file_changed_alone_fires_and_one_that_follows_code_stays_quiet`(골든만 바뀐 diff 와 코드가 함께 바뀐 diff) 통과
- AC-12: met — `--lib gate_collect::tests`(한도−1·한도는 읽고 한도+1 은 unknown, 이름 바꾸기 끄고 읽기) 통과
- AC-13: met — `--test test_weakening the_detector_module_is_a_protected_path_on_both_lists` 통과
- AC-6: met — `--test test_weakening a_test_only_change_that_weakens_nothing_stays_tier_zero` 통과
- AC-8: met — `--test gate_check_cli the_gate_lists_the_findings_and_says_unknown_when_unread` 통과, 수집기 쪽은 `--lib gate_collect::tests::the_collector_lists_weakened_tests_and_says_unknown_when_the_diff_is_unreadable` 통과
- AC-9: met — `--test test_weakening observe_reports_and_enforce_raises_the_tier` 통과
- AC-7: met — `--test test_weakening review_package_carries_the_findings`(약화 diff 는 `test_deleted tests/thing.rs:` 줄, 늘린 diff 는 빈 배열과 빈 `skipped`) 통과, 한도 쪽은 `--lib review::tests::a_diff_over_the_weakening_limit_gives_no_findings_and_one_skipped_entry`(한도 길이는 읽고 한 바이트 더 긴 diff 는 빈 목록과 `skipped` 한 줄) 통과
- AC-10: met — `--test test_weakening the_quality_rule_carries_the_connecting_sentence`(`quality.md` 에 문장이 한 번, "is not a finding" 과 `check-rules.md` 를 담은 줄) 통과
- AC-11: met — 위 표와 명령(NeXT 의 최근 착지 50개, 걸린 6개, 실제 약화 0건)

### 이번 가지의 변형(커밋한 뒤 하나씩 바꾸고 `git checkout` 으로 되돌림, 되돌린 뒤 작업 트리 깨끗함 확인)

| 변형 | 바꾼 것 | 결과 |
| --- | --- | --- |
| MN1 | `review.rs`: 패키지의 `test_weakening` 을 항상 빈 배열로 | `review_package_carries_the_findings` 실패(종료 100, 컴파일 오류 0) |
| MN2 | `review.rs`: 한도 초과 `skipped` 항목의 경로를 다른 글자로 | `a_diff_over_the_weakening_limit_gives_no_findings_and_one_skipped_entry` 실패(종료 100, 컴파일 오류 0) |
| MN3 | `quality.md`: 연결 문장을 지움 | `the_quality_rule_carries_the_connecting_sentence` 실패(종료 100, 컴파일 오류 0) |

세 변형 모두 시험이 실패해서 잡혔어요(생존 0).

영수증 수리: 이번 가지의 첫 `check affected --card`(머리 `0bc16399a2`)는 `result=FAIL`(test=100, 4121개 중 1개)이었어요. `next_flow_skill_scan refs_audit_tiers_match_the_code_floor` 가 감사 규약의 tier 2 접두사 목록에 `src/test_weakening` 이 없다고 알렸어요(코드의 `TIER2_PREFIXES` 에는 이미 있었어요). 목록에 한 단어를 더해 고쳤어요. 지정한 시험 몇 개만 돌린 앞선 실행에는 이 시험이 없어서 영수증이 처음 잡았어요.

이전 가지의 READY 검사 종료 코드(코드 head `46d1766cc4`, 기준 `341f6c7abb`, 이 가지의 것이 아니고 이력으로만 두어요; 이번 가지의 영수증은 READY 노트가 인용해요. 범위를 좁히지 않았고 `CARGO_TARGET_DIR` 는 이 worktree 전용): diff-check 0, fmt 0, ops rustfmt 0, clippy `--all-targets -D warnings` 0, `--lib` 전체 0(1639 통과), `--bins` 0, 영역 시험 대상(`gate_*`·`test_weakening`, 7개 대상, 155 통과) 0, 저장소 전체 검사용 시험(`git_env_scrub`, `home_isolation_scan`, `windows_unix_gate_scan`, `product_cli`) 0, `tools/repo_scans.json` 검사 41개 실패 0, 파이썬 `repo_scans_test`·`source_manifest_test`·`local_check_test` 0, `trailer-check` 0, 이 Mac 에서 도는 `cfg(macos)` 시험(7개 대상) 0. 처음 돈 READY 는 clippy 두 줄과 저장소 검사(`docs/` 글자) 한 건에서 걸렸고, 고친 뒤 처음부터 다시 돌려 위 값을 얻었어요.

돌리지 않은 것: GitHub CI(PR 없음), Windows. 이전 가지의 변형 16개는 이 가지에서 다시 돌리지 않았어요(코드 변화는 `weakened_in` 가시성 한 단어뿐이에요). 켜는 스위치를 `true` 로 한 실행은 하지 않았어요(관찰로 시작하는 결정이 D-1 이에요).
