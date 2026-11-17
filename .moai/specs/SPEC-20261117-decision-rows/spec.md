---
id: SPEC-20261117-decision-rows
title: "결정 기록 행 추가와 AGENTS.md 병합 조건의 tier 구분"
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

2026-11-17 오전의 리더 결정 1건과 사용자 결정·지시 4건이 결정 기록(`docs/decisions-2026-11-08.md`)에 아직
없어요. 그리고 `AGENTS.md` 작업 흐름 7(37행)은 병합에 "성공한 `ci`와 그 실행을 인용한 작성자 아닌 PASS"를
tier 구분 없이 요구해서, 같은 날 결정("CI 성공 + PASS", FAIL은 모든 tier에서 병합을 멈춤)과 모드 파일
`skills/moai/workflows/mode-leader.md`의 merge-train 문장과 달라요. 이 SPEC은 행 5개를 더하고 37행을 그
결정에 맞춰요.

## Requirements

요구 문장은 영어 GEARS예요(`src/spec.rs` `parse_gears`가 읽는 형식). `When` 요구가 없어 실패 경로 요구도 없어요.

- REQ-1: The decisions log shall hold one row in the existing table format for each of the five 2026-11-17 decisions this SPEC records.
- REQ-2: Each such row shall name the record it replaces or state that it replaces none.
- REQ-3: Each such row shall describe a card by its subject instead of a personal card number.
- REQ-4: Workflow step 7 of `AGENTS.md` shall require a non-author PASS that cites the `ci` run only at audit tier 1 or 2.
- REQ-5: Workflow step 7 of `AGENTS.md` shall state that a posted FAIL verdict stops the merge at every tier.
- REQ-6: The repository `AGENTS.md` shall stay below 9,216 bytes.
- REQ-7: The provenance table of `AGENTS.local.md` shall name the decisions behind the new step-7 sentence.
- REQ-8: The change shall carry its own SPEC at `.moai/specs/SPEC-20261117-decision-rows/spec.md`.

## Acceptance

- AC-1 (REQ-1): 시각이 적힌 2026-11-17 행이 5개 있어요.
  verify: `grep -c -E '^\| (04:1xZ|07:1xZ|08:0xZ|08:22:21Z|08:28:10Z)\(11-17\) \|' docs/decisions-2026-11-08.md` → `5`
- AC-2 (REQ-2, REQ-3): 그 5행 모두 `**대체**` 또는 `대체 없음`이 있고, 카드 번호가 없어요.
  verify: `grep -E '^\| (04:1xZ|07:1xZ|08:0xZ|08:22:21Z|08:28:10Z)\(11-17\) \|' docs/decisions-2026-11-08.md | grep -v -c -e '\*\*대체\*\*' -e '대체 없음'` → `0`;
  `grep -E '^\| (04:1xZ|07:1xZ|08:0xZ|08:22:21Z|08:28:10Z)\(11-17\) \|' docs/decisions-2026-11-08.md | grep -c -E '(^|[^A-Za-z])C-[0-9]'` → `0`
- AC-3 (REQ-4, REQ-5, REQ-6): 새 문장 2개가 고정되고, 옛 tier 구분 없는 문장은 없고, 크기는 9 KiB 미만이에요.
  verify: `cargo test --locked --offline --test repo_agents_md`; `wc -c < AGENTS.md` → 9216 미만
- AC-4 (REQ-7): 출처 표에 새 행이 있어요.
  verify: `grep -c 'FAIL 판정과 tier 0 병합' AGENTS.local.md` → `1`
- AC-5 (REQ-1, REQ-8): 이 SPEC과 결정 기록이 도구 검사를 통과해요.
  verify: `moai spec check --all --project .`(이 가지에서 빌드한 바이너리); `cargo test --locked --offline --test decision_append_cli`
- AC-6 (REQ-1…REQ-8): 바뀐 곳에 닿는 검사가 모두 통과해요.
  verify: `cargo fmt --check`; `cargo test --locked --offline --test spec_cli`; `cargo test --locked --offline --test ac_command_lint`;
  `cargo test --locked --offline --bin moai`; `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py tests/source_manifest_test.py`

## Scope

- write: `docs/decisions-2026-11-08.md`, `AGENTS.md`, `AGENTS.local.md`, `tests/repo_agents_md.rs`,
  `.moai/specs/SPEC-20261117-decision-rows/spec.md`
- read-only: `skills/moai/workflows/mode-leader.md`, `src/spec_doc.rs`, `.moai/specs/SPEC-20261117-sdd-foundation/spec.md`
- forbidden: 실제 HOME, 네트워크 쓰기(push, PR, Issue), 다른 worktree

## Plan

1. 결정 기록 표(11-17 행 뒤)에 행 5개: 04:1xZ GitHub 설정, 07:1xZ FAIL 판정, 08:0xZ HeadVer 먼저, 08:22:21Z 사전 허락,
   08:28:10Z 자율 진행.
2. RED: `tests/repo_agents_md.rs`에 새 문장 2개 고정과 옛 문장 부재 검사 → 실패 확인.
3. GREEN: `AGENTS.md` 37행에서 PASS 조건을 빼고, 아래 하위 항목 한 줄에 tier별 조건과 FAIL 규칙.
4. `AGENTS.local.md` 출처 표에 행 추가.
5. AC-1…AC-6 검사, `## Result` 기록.

## Decisions

- D1 행 시각은 근거 문서의 시각을 그대로 써요(`04:1xZ`, `07:1xZ`, `08:0xZ`, `08:22:21Z`, `08:28:10Z`). 그래서 기반
  SPEC의 AC-1(`xx:xxZ(11-17)` 행 16개)은 그대로예요.
- D2 08:0xZ 행의 사용자 원문은 카드 번호 대신 "HeadVer"로 적고, 그렇게 바꿨다고 행에 밝혀요.
- D3 07:1xZ 행은 "대체 없음"이에요: 같은 날 "병합 조건" 행에 tier 0의 FAIL 처리를 더하는 결정이고, 고친 `AGENTS.md`
  문장은 결정 기록이 아니라 그 결정을 옮긴 문장이에요.
- D4 04:1xZ 행은 지운 비밀값의 이름을 적지 않아요(개수와 readback만). 이름은 바꾸기 전 값 폴더에 있어요.
- D5 37행이 400자를 넘지 않게(시험의 한 줄 상한) tier 조건은 하위 항목 한 줄로 나눴어요.
- D6 브리프의 "6열 형식"은 실제 표가 5열(시각·주제·결정·적용·비고)이라 5열로 썼어요.

## Result

구현 에이전트가 2026-11-17에 채웠어요. SYNC와 AUDIT가 다시 확인해요. 기준 `c0b858de8c`(origin/v2026.11.0), 빌드 폴더는
이 worktree의 `target/`, `CARGO_INCREMENTAL=0`, `CARGO_BUILD_JOBS=4`, macOS arm64, cargo는 한 번에 하나.
`moai heavy-slot-acquire`가 75(스왑 82%로 시작 안 함, 빌드 폴더 경고)를 내서 리더 지시대로 슬롯 밖에서 돌렸어요.

- AC-1: met — `grep -c -E …` → `5`. 기반 SPEC의 `xx:xxZ(11-17)` 행 수는 그대로 `16`.
- AC-2: met — 대체 표시 없는 행 `0`(`대체 없음` 5행), 카드 번호 `0`.
- AC-3: met — RED(새 고정 문장, `AGENTS.md` 수정 전): exit 101, `tests/repo_agents_md.rs:86`
  "AGENTS.md must say `Tier 1 or 2 also needs a non-author PASS that cites that run; tier 0 needs `ci` alone.`".
  GREEN: exit 0, 3 passed. `wc -c < AGENTS.md` → `9174`(전 `9079`), 가장 긴 줄 373바이트.
  mutation(커밋 `c680428206` 뒤, 각각 `git restore`로 되돌리고 tree 깨끗 확인): FAIL 문장 삭제 → exit 101
  "must say `A posted FAIL stops the merge at every tier.`"; 37행에 옛 PASS 문구 복원 → exit 101 "still says
  `a successful `ci` run and a non-author PASS that cites that run`". 되돌린 뒤 exit 0.
- AC-4: met — `1`.
- AC-5: met — `moai spec check --all --project .`(이 가지 `target/debug/moai`) exit 0, errors 0, warnings 0, SPEC 3개;
  `decision_append_cli` exit 0(17 passed, 실제 결정 기록 미리보기 포함).
- AC-6: met — `cargo fmt --check` 0; `spec_cli` 0(17); `ac_command_lint` 0(5); `--bin moai` 0(27);
  Python 83개 OK(exit 0).

실행 안 함: 전체 `cargo test`(바뀐 곳 시험만), clippy(Rust 코드는 시험 파일의 문자열·주석만 바뀜), Linux·Windows.
