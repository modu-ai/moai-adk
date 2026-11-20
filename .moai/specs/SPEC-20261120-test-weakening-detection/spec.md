---
id: SPEC-20261120-test-weakening-detection
title: "Detect a weakened test in the diff and hand the findings to the auditor"
status: draft
revision: 1
created: 2026-11-20
---

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
- REQ-10: While the weakening gate is in observe, `effective_tier` shall ignore the findings. · failure: none
- REQ-13: While the weakening gate is in observe, the gate's result shall name the tier it would give with the findings. · failure: none
- REQ-12: While the weakening gate is enforcing, `effective_tier` shall be at least 1 when the findings are not empty or are unknown. · failure: none
- REQ-11: When the quality rule and the agent workflow speak of a deleted test, the text shall say in one connecting sentence that a deletion whose reason is written in the card and the SPEC is not a finding. · failure: none

## Acceptance

- AC-1 (REQ-1): 시험 함수를 지우는 diff 는 `test_deleted` 를 `파일:줄` 과 함께 내고, 같은 이름을 다른 파일에 다시 더하는 diff(옮기기)는 내지 않아요(Rust `#[test]`, Python `def test_` 각각 한 쌍의 고정물) · verify: `cargo test --locked --offline --test test_weakening each_kind_fires_on_its_fixture_and_stays_quiet_on_its_pair`
- AC-2 (REQ-2): 건너뛰기 다섯 표기를 더하는 diff 는 `skip_added` 를 내고, 건너뛰기를 지우거나 설명 글에만 쓰는 diff 는 내지 않아요 · verify: `cargo test --locked --offline --test test_weakening each_kind_fires_on_its_fixture_and_stays_quiet_on_its_pair`
- AC-3 (REQ-3): 단언이 순감소한 시험 파일만 `assertions_reduced` 를 내고, 단언을 옮기거나 늘린 파일과 시험 파일이 아닌 파일은 내지 않아요 · verify: `cargo test --locked --offline --test test_weakening each_kind_fires_on_its_fixture_and_stays_quiet_on_its_pair`
- AC-4 (REQ-4): 시험 파일에 sleep 을 더하면 `sleep_added`, 시험 파일이 아닌 곳에 더하면 내지 않아요 · verify: `cargo test --locked --offline --test test_weakening each_kind_fires_on_its_fixture_and_stays_quiet_on_its_pair`
- AC-5 (REQ-5): 골든 파일만 바뀐 diff 는 `golden_only`, 같은 골든과 함께 소스 코드도 바뀐 diff 는 내지 않아요 · verify: `cargo test --locked --offline --test test_weakening each_kind_fires_on_its_fixture_and_stays_quiet_on_its_pair`
- AC-6 (REQ-6): 시험 파일만 바꾸고 다섯 가지에 걸리지 않는 diff 는 빈 목록이고 `tier_of(Some(0), 그 파일들)` 는 여전히 0 이에요 · verify: `cargo test --locked --offline --test test_weakening a_test_only_change_that_weakens_nothing_stays_tier_zero`
- AC-7 (REQ-7): `review-package` 결과의 `test_weakening` 이 약화 diff 에서는 목록, 아닌 diff 에서는 빈 배열이에요 · verify: `cargo test --locked --offline --test test_weakening review_package_carries_the_findings`
- AC-8 (REQ-8, REQ-9): `gate-check` 의 결과가 걸린 목록을 보이고, diff 를 읽을 수 없으면 `unknown` 을 보여요 · verify: `cargo test --locked --offline --test gate_check_cli the_gate_lists_the_findings_and_says_unknown_when_unread`
- AC-9 (REQ-10, REQ-12, REQ-13): 관찰 중에는 `effective_tier` 가 그대로이고 결과가 올렸을 때의 tier 를 말하며, 켜면 목록이 비어 있지 않거나 unknown 일 때 최소 1 이에요 · verify: `cargo test --locked --offline --test test_weakening observe_reports_and_enforce_raises_the_tier`
- AC-10 (REQ-11): 두 문서에 연결 문장이 한 번씩 있어요 · verify: `cargo test --locked --offline --test test_weakening the_two_documents_carry_the_connecting_sentence`
- AC-11: 최근 병합된 PR 50개(로컬 병합 커밋의 두 부모 사이)에 돌린 결과 — 걸린 수, 그중 실제 약화 수 — 가 이 SPEC 의 Result 에 있어요 · verify: manual: Result 표를 읽어요(명령과 출력이 Result 에 적혀 있어요)

## Scope

- write: 이 SPEC, `src/test_weakening.rs`(새), `src/gate.rs`, `src/gate_collect.rs`, `src/ops/gate_check.rs`, `src/review.rs`(조각 3b), `tests/test_weakening.rs`(새), `tests/gate_check_cli.rs`, `skills/moai/references/internal/quality.md`와 `skills/moai/workflows/mode-agent.md`(조각 4)
- 다른 카드가 쥔 파일(2026-11-20 예약 읽음): `src/review.rs`(조각 3 의 `review-package` 키), `skills/moai/references/internal/quality.md`와 `skills/moai/workflows/mode-agent.md`(조각 4 의 연결 문장)는 쥔 카드가 풀거나 리더가 순서를 정한 뒤에 써요. 그 전까지 조각 2와 `gate-check` 쪽(`src/gate.rs`, `src/gate_collect.rs`)을 먼저 해요
- read-only: `src/audit_gate.rs`, `tools/protected_checks.json`, `skills/moai/workflows/audit.md`
- forbidden: 모델 호출, 네트워크, 약화가 걸린 PR 을 막는 규칙, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것

## Plan

1. 조각 1(이번): SPEC, 조사 보고, PLAN 노트. 범위 확장 요청(`src/gate.rs`, `src/review.rs` 가 다른 카드에 잡혀 있으면 순서를 리더에게).
2. 조각 2(검사기): `src/test_weakening.rs` 순수 함수(diff 글 → 목록)와 다섯 가지 고정물 쌍 시험(AC-1~6). 별도 명령은 만들지 않아요(D-7).
3. 조각 3a(소비, gate-check): 스냅샷 키 `test_weakening`, 수집기가 diff 를 읽음, 결과에 목록·`tier_with_findings`, 관찰/켜짐 스위치(AC-8, AC-9). 조각 3b: `review-package` 의 새 키(AC-7, `src/review.rs` 를 쥔 카드가 풀린 뒤).
4. 조각 4(문서와 결과): 연결 문장 두 곳(AC-10), 최근 PR 50개 실행 결과를 Result 에(AC-11).

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

## Result

이 조각(2·3a: 검사기와 `gate-check` 연결)의 근거예요. 코드 head `46d1766cc4`. 카드는 부분 READY 로 열어 두어요: AC-7(`review-package` 키, `src/review.rs` 를 다른 카드가 쥠)과 AC-10(문서 연결 문장, `quality.md`·`mode-agent.md` 를 다른 카드가 쥠)은 이번 조각에 없고 후속이에요. 관찰/켜짐 스위치(`TEST_WEAKENING_ENFORCED`)는 `false` 로 두었어요: 이 조각이 머지돼도 `effective_tier` 는 달라지지 않고 결과가 목록과 올렸을 때의 tier 를 보이기만 해요. 켜는 일은 리더 결정이에요.

TDD 이탈: 검사기(`src/test_weakening.rs`)를 시험보다 먼저 썼어요. 시험은 그 뒤에 다섯 가지 고정물 쌍으로 쓰고, 변형 10개로 시험이 검사기를 정말 지키는지 확인했어요.

### 변형 시험(커밋한 뒤, 한 번에 하나씩 바꾸고 되돌림)

| 변형 | 바꾼 것 | 잡은 시험 |
| --- | --- | --- |
| M1 | Rust 건너뛰기 표기가 맞지 않음 | collector 시험(`the_collector_lists_weakened_tests_…`) |
| M2 | 옮긴 시험도 삭제로 셈 | `each_kind_fires_on_its_fixture_and_stays_quiet_on_its_pair` |
| M3 | 단언 수가 같아도 줄었다고 셈 | 같은 시험 |
| M4 | 새 골든 파일도 다시 쓴 것으로 셈 | 같은 시험 |
| M5 | 골든 옆의 코드 변경을 무시 | 같은 시험 |
| M6 | 시험 파일이 아닌 곳의 단언·sleep 도 셈 | 같은 시험 |
| M7 | 관찰 중에도 tier 를 올림 | `observe_reports_and_enforce_raises_the_tier`, `the_gate_lists_the_findings_and_says_unknown_when_unread` |
| M8 | diff 를 못 읽었을 때 tier 를 안 올림 | `observe_reports_and_enforce_raises_the_tier` |
| M9 | git diff 실패를 빈 목록으로 읽음 | collector 시험 |
| M10 | 결과가 못 읽은 diff 를 빈 목록으로 보임 | `observe_reports_and_enforce_raises_the_tier`, `the_gate_lists_the_findings_and_says_unknown_when_unread` |

기준 실행은 세 대상 모두 0 이었고 변형 10개는 모두 시험이 실패해서 잡혔어요(생존 0).

### 최근 병합 50개에 돌린 결과(AC-11)

명령: `git log --first-parent --merges -50 v2026.11.0` 의 병합 50개마다 `git diff --no-ext-diff --no-color --no-renames --unified=0 <병합>^1 <병합>` 을 파일로 저장하고, 커밋한 `src/test_weakening.rs` 를 그대로 가져와 빌드한 임시 실행기로 `detect` 를 돌렸어요(PR API 는 쓰지 않았어요). 출력 맨 끝 줄: `TOTAL diffs=50 with_findings=6`.

| 병합 | 걸린 것 | diff 로 읽은 판단 |
| --- | --- | --- |
| 언어 세 질문(`language-three-questions`) | `assertions_reduced` `tests/language_setting.rs:211` 8 줄 빠짐·6 줄 더함 | 묻는 항목을 8개에서 3개로 줄이는 동작 변경에 맞춘 기대값 축소. 이유가 있는 변경이지만 목록에 올라야 할 변경 |
| CI 기준 범위(`ci-base-scope`) | `test_deleted` 2건(`test_an_unknown_base_falls_back_to_the_full_matrix`, `test_targets_diff_uses_no_renames`) | 같은 diff 에 다른 이름의 새 시험이 있어요(이름이 바뀐 대체). 검사기는 이름이 같을 때만 옮긴 것으로 봐서 삭제로 보여요 |
| 릴리스 4(#3234) | `test_deleted` 2건(`ci_superseded_test.py`) | 같은 diff 에 이름이 바뀐 대체 시험이 있어요(위와 같은 한계) |
| 릴리스 1(#3231) | `sleep_added` `tests/profile_plan_cli.rs:493` | 30 초 마감이 있는 `try_wait` 반복 안의 50ms 대기(기다림 루프). 고정 sleep 이 아니라서 사람이 보면 바로 넘어갈 오탐 |
| 묶음 6(#3230) | `sleep_added` `tests/hooks_cli.rs:2202` | 로그 시각이 달라지게 하는 5ms 고정 대기(`ac_pause`). 진짜 sleep 이고 이유를 물을 만해요 |
| 세 도구 launcher(#3211) | `assertions_reduced` 1건 + `test_deleted` 4건(`tests/launcher_three_hosts.rs`) | 제거한 호스트의 시험 파일 전체를 지움(단언 19). 기능 제거에 따른 의도된 삭제 |

결과: 50개 중 6개(12%)가 걸리고 걸린 줄은 모두 12건이에요. 걸린 6개 모두 기능 변경·제거·이름 바꾸기·기다림이 이유로 보이는 변경이고, 이유 없이 시험을 약하게 만든 diff 는 diff 만 봐서는 하나도 확인하지 못했어요(작성자에게 이유를 묻지는 않았어요). 사람이 바로 넘길 오탐은 12건 중 1건(기다림 루프)이에요. 켜면 50개 중 6개가 tier 1 이 되므로 켜는 일은 리더가 이 표를 보고 정해요.

### 알려진 한계

- 이름이 바뀐 시험은 삭제로 보여요(위 두 건). 한 파일에서 지운 시험을 다른 파일에 같은 이름으로 다시 더하면 옮긴 것으로 보지만, 지운 쪽 파일의 단언 순감소는 파일 단위로 그대로 세요(D-8).
- 단언은 파일 단위로 세요. 시험 함수 안에서 단언을 다른 곳으로 옮기면 줄었다고 볼 수 있어요.
- 8MiB 가 넘는 diff 는 읽지 않고 `unknown` 으로 두어요.

### 수락 기준

- AC-1: met — `--test test_weakening each_kind_fires_on_its_fixture_and_stays_quiet_on_its_pair`(Rust `#[test]`·Python `def test_` 각 한 쌍) 통과
- AC-2: met — 같은 시험(건너뛰기 표기 다섯 개와 그 짝) 통과
- AC-3: met — 같은 시험(단언 순감소만) 통과
- AC-4: met — 같은 시험(시험 파일의 sleep 만) 통과
- AC-5: met — 같은 시험(골든만 바뀐 diff 와 코드가 함께 바뀐 diff) 통과
- AC-6: met — `--test test_weakening a_test_only_change_that_weakens_nothing_stays_tier_zero` 통과
- AC-8: met — `--test gate_check_cli the_gate_lists_the_findings_and_says_unknown_when_unread` 통과, 수집기 쪽은 `--lib gate_collect::tests::the_collector_lists_weakened_tests_and_says_unknown_when_the_diff_is_unreadable` 통과
- AC-9: met — `--test test_weakening observe_reports_and_enforce_raises_the_tier` 통과
- AC-11: met — 위 표와 명령(병합 50개, 걸린 6개)

AC-7 과 AC-10 은 위에 쓴 대로 이번 조각에 없어요(후속 조각).

READY 검사 종료 코드(코드 head `46d1766cc4`, 기준 `341f6c7abb`, 범위를 좁히지 않았고 `CARGO_TARGET_DIR` 는 이 worktree 전용): diff-check 0, fmt 0, ops rustfmt 0, clippy `--all-targets -D warnings` 0, `--lib` 전체 0(1639 통과), `--bins` 0, 영역 시험 대상(`gate_*`·`test_weakening`, 7개 대상, 155 통과) 0, 저장소 전체 검사용 시험(`git_env_scrub`, `home_isolation_scan`, `windows_unix_gate_scan`, `product_cli`) 0, `tools/repo_scans.json` 검사 41개 실패 0, 파이썬 `repo_scans_test`·`source_manifest_test`·`local_check_test` 0, `trailer-check` 0, 이 Mac 에서 도는 `cfg(macos)` 시험(7개 대상) 0. 처음 돈 READY 는 clippy 두 줄과 저장소 검사(`docs/` 글자) 한 건에서 걸렸고, 고친 뒤 처음부터 다시 돌려 위 값을 얻었어요.

돌리지 않은 것: GitHub CI(PR 없음), Windows, `review-package` 와 문서 연결 문장(후속).
