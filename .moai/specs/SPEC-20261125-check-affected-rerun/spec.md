---
id: SPEC-20261125-check-affected-rerun
title: "A repeated receipt run re-runs only the tests that failed, and a test that passes on retry is recorded as FLAKY"
status: approved
revision: 1
created: 2026-11-25
approval:
  by: user
  at: 2026-11-25
  basis: "사용자가 리더 대화의 선택지에서 직접 고른 답(카드 노트 9000): \"승인, 재시도는 모든 검사에 (Recommended)\"; 리더가 기록 직전에 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:205410128c25a3340624375cde90e5570416fcf5f062253b6be2de8aa5081db7"
---

# A repeated receipt run re-runs only the tests that failed, and a test that passes on retry is recorded as FLAKY

## Interview

- Q1: 이미 통과한 테스트를 다시 돌리지 않아도 되나요? 부하에 흔들리는 테스트는 어떻게 하나요? → A: "부하에 민감한 테스트는 패슬르하거나 매번 똑같이해야하나? 이미 테스트 통과한건 또 안해도되지않나?"(사용자 물음 2026-11-25, 카드 본문), 기존 지시 "같은 검증을 되풀이하지 않는다"(2026-11-24)
- Q2: 재시도는 영수증 실행에만 적용할까요, 모든 검사에 적용할까요(D-2)? → A: "승인, 재시도는 모든 검사에 (Recommended)"(사용자 선택 2026-11-25T05:38:32Z, 카드 노트 9000)
- Q3: 좁히는 방법(D-1), FLAKY 기록(D-3), 좁히기 끄기(D-4), 켜지는 때(D-5)는 어떻게 할까요? → A: 리더가 모두 A(기본)로 정했어요(카드 노트 9000, LEADER-DECISION, 되돌릴 수 있음, 분류 none)

## Intent

사용자 물음(2026-11-25, 말 그대로): "부하에 민감한 테스트는 패슬르하거나 매번 똑같이해야하나? 이미 테스트 통과한건 또 안해도되지않나?" 와 기존 지시 "같은 검증을 되풀이하지 않는다"(2026-11-24)가 근거예요.

- 측정(카드 본문): 24시간 동안 검사 실행 193번 중 실패 38%, 착지 한 건당 평균 3.0번 실행. 한 건이 4000개 넘는 시험 중 1개만 실패해도(부하에 흔들린 시험이나 표 등록 누락) 15~20분짜리 전체 영수증을 처음부터 다시 받았어요.
- 이 SPEC 이 바꾸는 것: 같은 head 에서 `moai check affected` 를 다시 부르면 앞 실행에서 통과한 시험은 다시 돌리지 않고, 실패한 시험만 돌려요. 시험 단계는 실패한 시험을 정해진 횟수만 다시 돌려 보고, 다시 돌려 통과한 시험은 숨기지 않고 FLAKY 로 남겨요.
- 바꾸지 않는 것: 착지 전 넓은 검사(`card land-batch` 가 부르는 전체 집합)는 늘 전체를 그대로 돌려요. 안전망이에요. 게이트를 느슨하게 하지 않아요.

조사(2026-11-25T05:18Z, 설치된 `cargo-nextest 0.9.152` 의 `cargo nextest run --help` 와 `cargo nextest store --help`, 읽은 그대로):

- `--retries <N>` 은 실패한 시험을 N번 다시 돌려요. `--flaky-result <pass|fail>` 은 재시도로 통과한 시험을 통과로 칠지 실패로 칠지 정해요. 둘 다 안정 기능이에요. 재시도로 통과한 시험은 nextest 가 `FLAKY` 로 표시해요.
- `-R, --rerun <RUN_ID_OR_RECORDING>` 은 앞 기록 실행에서 통과하지 못한 시험만 돌리고 새 시험도 넣어요. 그런데 기록 저장소는 실험 기능이라 `NEXTEST_EXPERIMENTAL_RECORD=1` 이 있어야 하고, 이 컴퓨터의 `cargo nextest store list` 는 `no recorded runs` 였어요. 그래서 이 기능에 기대지 않아요(Decision D-1).
- 이 저장소의 `moai check affected` 는 이미 단계마다 로그를 `target/check-affected-<head12>/NN-<단계>.log` 에 남겨요. 시험 단계 로그에는 `FAIL [..] (i/N) <binary-id> <시험 이름>` 줄이 있어요(실제 로그에서 읽음).

## Non-goals

- 넓은 검사(`card land-batch`, `--full`)를 좁히거나 재시도로 가리지 않아요.
- 시험 코드나 시험 표(`allow.tsv`, `repo_scans.json`)를 바꾸지 않아요. 흔들리는 시험을 고치는 일은 FLAKY 기록이 가리키는 후속이에요.
- 원격 CI 를 바꾸지 않아요.
- nextest 설정 파일(`.config/nextest.toml`)을 추가하지 않아요. 플래그로 넘겨요.

## Requirements

<!-- This SPEC changes the check command that produces the READY receipt; every change makes the set of tests smaller only when a previous run of the same head proved those tests, and the wide set is untouched. -->
- REQ-1: When `moai check affected` runs on a head whose check folder holds an earlier finished run with the same base and the same selection whose test step failed, the test step shall run only the tests that failed or did not finish there and the tests the range holds that the earlier run did not run. · failure: none — the earlier run proved the other tests on this same head; REQ-5 covers a log that names no failure.
- REQ-2: When the test step ran as in REQ-1, the result shall be `PASS` only when every step of this run exits 0 and every test of the earlier failure list passed now. · failure: none — a test that still fails gives the result FAIL, and the run names it.
- REQ-3: When the head has moved since the earlier run, the test step shall run the targets the diff between the two heads reaches and the tests that failed earlier. · failure: none — REQ-4 covers a diff that touches a wide path.
- REQ-4: When the diff between the two heads touches a wide path, the run shall use the full set with `rerun=full reason=<class>` in the summary line. · failure: none — a wide path always gives the full set.
- REQ-5: When the earlier test log is missing or unreadable or names no failed test or the earlier run stopped before the test step, the run shall use the full set with `rerun=full reason=<why>` in the summary line. · failure: none — the full set is the safe default.
- REQ-6: The test step shall pass `--retries <n>` and `--flaky-result pass` to nextest, where `<n>` is 2 unless D-2 says otherwise. · failure: none — a test that fails every attempt still fails (REQ-8).
- REQ-7: A test that failed first and passed on a retry shall be listed as `FLAKY` in the summary line (`flaky=<count>`), in the step log and in `flaky.txt` of the check folder. · failure: none — the retry never hides the first failure.
- REQ-8: A test that fails every attempt shall make the test step exit non-zero. · failure: the step exits non-zero and the result is FAIL.
- REQ-9: Each name listed `FLAKY` shall be appended to a ledger file of the project state folder with the time, the head and the check folder. · failure: a ledger that cannot be written is reported on stderr and does not change the result of the run.
- REQ-10: `moai check flaky` shall print each ledger name with its count and mark a name seen twice as `fix`. · failure: an unreadable ledger answers an error and prints no count.
- REQ-11: A run through `moai card land-batch` or a run given `--full` shall run the full set without reading an earlier run. · failure: none — the wide check is the safety net; it still retries and records FLAKY by REQ-6 and REQ-7.
- REQ-12: The `summary.txt` of a combined run shall carry `rerun_of=<earlier folder>` and stay a receipt for `card note`. · failure: `card note` refuses a READY note whose summary is not a receipt, as today.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-12): in a temporary project whose check run failed one test, a second run on the same head runs only that test, the summary line carries `rerun_of=` and `result=PASS` after the test passes, and it stays `FAIL` while the test still fails · verify: `cargo test --offline --locked --test check_affected_rerun ac1_`
- AC-2 (REQ-3, REQ-4): a run after a head move runs the changed targets and the earlier failures only, and a diff that touches a wide path gives the full set with `rerun=full reason=` · verify: `cargo test --offline --locked --test check_affected_rerun ac2_`
- AC-3 (REQ-6, REQ-7, REQ-8): a test that fails once and passes on retry is listed in `flaky=<count>`, in the step log and in `flaky.txt`, the step exits 0, and a test that fails every attempt exits non-zero · verify: `cargo test --offline --locked --test check_affected_rerun ac3_`
- AC-4 (REQ-5): a missing, empty or failure-free earlier log and an earlier run that stopped before the test step give the full set with `rerun=full reason=` · verify: `cargo test --offline --locked --test check_affected_rerun ac4_`
- AC-5 (REQ-9, REQ-10): the ledger gets one line per FLAKY name, `moai check flaky` prints counts, a name seen twice shows `fix`, and an unwritable ledger does not change the run result · verify: `cargo test --offline --locked --test check_affected_rerun ac5_`
- AC-6 (REQ-11): a `--full` run and the run `card land-batch` makes read no earlier run, run the full set, and still record FLAKY · verify: `cargo test --offline --locked --test check_affected_rerun ac6_`
- AC-7 (REQ-12): a combined `summary.txt` is accepted by `moai card note` as a READY receipt and names both folders · verify: `cargo test --offline --locked --test check_affected_rerun ac7_`
- AC-8: the numbers before and after (check runs per landed card and the median time of a retaken receipt) are written in the card note · verify: not testable by `cargo`; the lane measures with `moai flow-measure` and writes the note (`not_run` until then).

## Decisions

각 결정은 되돌릴 수 있고 분류는 none 이에요. 추천이 맨 앞이고, 사용자에게 선택 질문으로 올려요.

- D-1 (방법): (A, 기본·권장) `moai` 가 앞 실행의 시험 단계 로그에서 `FAIL` 줄의 이름을 읽어 nextest 필터(`-E`)를 만들어요. 안정 기능만 쓰고 기록 저장소가 필요 없어요. (B) nextest 의 `--rerun latest` 를 써요: 코드가 짧지만 실험 기능이고 기록이 없으면 동작하지 않아요. 기본값 A.
- D-2 (재시도 횟수와 범위): (A, 기본·권장) 재시도 2번, 영수증 실행과 넓은 검사 모두에 적용하고 FLAKY 로 남겨요. (B) 영수증 실행에만 적용하고 넓은 검사는 재시도 없이 한 번에 통과해야 해요(더 엄격하지만 부하에 흔들리면 착지가 멈춰요). 기본값 A.
- D-3 (FLAKY 기록 위치와 읽는 명령): (A, 기본·권장) 프로젝트 상태 폴더의 추가 전용 파일(한 줄에 시험 이름, 시간, head, 폴더)과 읽기 명령 `moai check flaky`. (B) 카드 노트로만 남겨요(새 명령 없음, 합계는 사람이 세요). 기본값 A.
- D-4 (좁히기를 끄는 방법): (A, 기본·권장) `--full` 플래그 하나와 `card land-batch` 의 내부 호출은 늘 전체. (B) 환경 변수로 끄기. 기본값 A.
- D-5 (좁히기가 켜지는 때): (A, 기본·권장) 같은 head 와 base 에서 앞 실행이 실패로 끝났으면 자동으로 켜져요(READY 영수증을 다시 받는 흐름 그대로). (B) 사용자가 `--again` 을 줘야 켜져요(안전하지만 지금 습관이 안 바뀌어요). 기본값 A.
- D-6 (범위 선행 조건, 리더가 알아야 할 것): `src/check_affected.rs` 와 `src/ops/check.rs` 는 영수증 선택을 고치는 다른 카드가 쥐고 있어요. 이 카드는 그 카드가 반납한 뒤에 코드 범위를 넓혀요(그 전에는 SPEC 과 시험 파일만).

## Scope

- write: `.moai/specs/SPEC-20261125-check-affected-rerun/**`, `tests/check_affected_rerun.rs`
- write (다른 카드의 반납 뒤에 `card scope` 로 넓힘): `src/check_affected.rs`, `src/ops/check.rs`, `src/check_flaky.rs`(새 모듈), `src/card/land/batch.rs`(전체 집합을 못 박는 한 줄)
- write (한 줄, line-lease): `src/lib.rs`(새 모듈 등록), `src/help.rs`(`check flaky`, `--full`), `tests/fixtures/help/*.txt`(도움말 골든), `tools/repo_scans.json`(새 시험 파일 등록)
- read-only: `src/heavy_slot.rs`, `src/check_pipeline*.rs`, `src/card/note.rs`(영수증을 읽는 부분)
- forbidden: `.config/nextest.toml`, 시험 표(`allow.tsv`)의 값 바꾸기, 전역 설치본과 실제 HOME, 네트워크, Python, 넓은 검사를 약하게 하는 변경

## Plan

1. 조사는 끝났어요(위 조사). 단계 로그의 `FAIL` 줄 모양은 실제 로그로 한 번 더 확인하고 시험 입력에 고정해요.
2. RED: `tests/check_affected_rerun.rs`(AC-1..AC-7)가 지금 동작에서 실패해요. 시험은 임시 프로젝트와 가짜 단계 로그를 만들어요(실제 HOME 을 안 써요).
3. 앞 실행 읽기(REQ-1, REQ-3, REQ-5)와 필터 만들기, 합친 요약(REQ-2, REQ-8).
4. 재시도와 FLAKY 표시(REQ-4), 기록 파일과 `moai check flaky`(REQ-6), `--full` 과 land-batch(REQ-7). GREEN.
5. 가드 뮤턴트(새 시험마다 한 줄), 위험 경로 실행: 임시 프로젝트에서 흔들리는 시험 하나와 실패하는 시험 하나로 실제 `check affected` 를 두 번 돌려요.
6. `moai check affected` 영수증, AC-8 숫자를 노트에 적고 READY.

## Risks

- R-1: 앞 실행의 통과를 믿고 시험을 건너뛰면, 같은 head 라도 환경이 달라 통과가 달라지는 시험을 놓칠 수 있어요. 그래서 같은 head 와 base, 같은 선택일 때만 좁히고 넓은 검사는 늘 전체예요(REQ-7).
- R-2: 로그 줄 모양이 nextest 버전에서 바뀌면 이름을 못 읽어요. 못 읽으면 전체로 돌아가요(REQ-5)라서 안전한 쪽으로 실패해요.
- R-3: 재시도가 진짜 결함을 가릴 수 있어요. FLAKY 로 숨기지 않고 남기고, 두 번 잡히면 `fix` 로 보여요(REQ-4, REQ-6).
- R-4: 다른 카드가 같은 파일을 고쳐서 병합 충돌이 날 수 있어요(D-6). 그 카드가 먼저 착지하면 이 카드는 그 tip 위에서 코드를 시작해요.
