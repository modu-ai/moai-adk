---
id: SPEC-20261119-c41-evidence
title: "C-41 evidence: the pre-PR candidate check and its evidence line are already on the tip (PR #3086·#3098·#3089)"
status: draft
revision: 1
created: 2026-11-19
---

# C-41 evidence (PR #3086·#3098·#3089)

## Intent

카드 C-41 (p2, 재설계 1단계) 을 통합 가지 tip(`7d744a28f3f9860e14152777eadce2017ad3a09b`)에서 확인한 결과, 요구 1·2·3 과 여섯 AC 전부가 **이미 PR #3098**(P2, candidate 모드·long evidence 줄, 병합 2026-11-16)·**#3086**(P1, 게이트가 long 줄을 읽음)·**#3089**(P3, base 조상성 G1)로 들어와 있어요. 이 SPEC 은 새 코드 없이 그 증거를 기록해요(리더 배차 "한 조각만 좁게" — 남은 조각이 리더 소유 지침 파일 하나뿐이라 증거로 닫아요).

## Non-goals

- 새 구현 — tip 에 완성돼 있어요.
- `AGENTS.local.md` 의 옛 전체-검사 문단(:37 "implement and run the full check set"·:42 Check set 블록) 수정 — 리더 소유 파일이라 이 카드에서 건드리지 않고 리더 결정으로 남겨요(카드 요구 4의 마지막 잔여).

## Requirements

- REQ-1: When the pre-PR candidate check is read from the tip, `tools/local_check.sh` shall already run the head merged with the latest base tip in a throwaway worktree and print the long evidence line (head·base·tree·script·rustc·run). · failure: none — 위반 시 PR #3098 의 시험이 어긋나요; 이 SPEC 은 tip 증거 기록이에요.
- REQ-2: When a verdict names a base, the gate shall already stop when that base is not the base-branch tip or one of its ancestors. · failure: none — 위반 시 PR #3089 의 G1 이 어긋나요.

## Acceptance

- AC-1 (REQ-1): tip 시험 증거 — `tests/local_check_test.py:539` `test_head_passing_alone_fails_when_merged`(AC1, #3062 형태)·`:566` 한 줄 통과·`:578` base 이동.
  verify: `python3 -m unittest discover -s tests -p 'local_check_test.py'`
- AC-2 (REQ-2): 게이트 증거 — `src/gate_collect.rs:1026-1029` `verdict_base_in_tip`(is_ancestor), 시험 `tests/gate_collect.rs:173,453,482,487,497,505,509`(조상·비조상·미지 경계).
  verify: `cargo test --locked --offline --lib gate_collect`
- AC-3 (REQ-1): 결과 줄 필드·변조 — `tests/local_check_test.py:391`(필드 클래스·none 규칙)·`:865`(필드 폭 고정)·`src/gate.rs` `parse_local_check_evidence`.
  verify: `python3 -m unittest discover -s tests -p 'local_check_test.py'`
- AC-4 (REQ-1, REQ-2): 격리·단계 일치 — `tests/local_check_test.py:652`(AC5: 폴더·index·HEAD·가지·stash·worktree 목록 불변)·`:167`·`:826`(CI 단계 일치).
  verify: `python3 -m unittest discover -s tests -p 'local_check_test.py'`; `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261119-c41-evidence/**`
- read-only: `tools/local_check.sh`, `src/gate.rs`, `src/gate_collect.rs`, `tests/local_check_test.py`, `tests/gate_collect.rs`, PR #3086·#3098·#3089
- forbidden: 도구·CI·src 변경(tier 2 영역), 실제 태그·릴리스, 네트워크

## Plan

1. tip 확인(완료): 요구 1·2·3 과 AC1–AC6 이 #3098·#3086·#3089 로 충족 — 위 verify 줄로 관찰.
2. 좁게 take(SPEC 폴더만) → 이 SPEC → push(PR 없음) → 기준 met → 근거 PR 연결 → 감사(tier 0) → READY.

## Risks

- 증거 SPEC 이므로 새 위험 없음; 세 PR 의 감사는 이미 지나갔어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | "이미 됨" 경로 — 구현 없이 증거 기록만 | tip 에서 여섯 AC 전부 확인 | 구현 카드로 재발행 |
| D-2 | `AGENTS.local.md` 잔여 문단은 리더 결정으로 남김 | 리더 소유 파일(AGENTS.local.md § Sessions) | 리더가 문단 갱신 |

## Result

- AC-1: met — tip `tests/local_check_test.py:539` `test_head_passing_alone_fails_when_merged`(#3062 형태: origin 이 먼저 규칙을 얻고 head 는 모름 → candidate 실행이 실패), `:578` base 이동이 줄의 base 를 바꿈. 근거 PR #3098.
- AC-2: met — `src/gate_collect.rs:1026-1029`(verdict_base_in_tip, is_ancestor, 미지 커밋은 false)+`tests/gate_collect.rs` 8곳 단언(조상 true·비조상 false·알려진 커밋 경계). 근거 PR #3089.
- AC-3: met — `tests/local_check_test.py:391` 필드 클래스(base·tree none 규칙 포함)·`:865` 필드 폭 고정·`src/gate.rs` `parse_local_check_evidence`(P1, PR #3086).
- AC-4: met — `:652` AC5 격리 시험(폴더·index·HEAD·가지·stash·worktree 목록·무시 파일 불변)·`:167`+`:826` CI 단계 일치.
  관찰(2026-11-19T06:0xZ, tip 7d744a28f3 추출 없이 worktree 트리=tip): `python3 -m unittest discover -s tests -p 'local_check_test.py'` → 36 tests OK(140.2초),
  `cargo test --lib gate_collect` → 1 passed(0.50초), `cargo test --test gate_collect` → 44 passed(18.97초).
- 요구 4 잔여(기준 밖): `AGENTS.local.md:37`("implement and run the full check set")·`:42` Check set 블록이 통일 문장(AGENTS.md:33·skills/moai/workflows/mode-agent.md:63 과 같은 뜻의 targeted+CI 인용)과 아직 다른 표현을 유지 — 리더 소유 파일이라 이 카드에서 고치지 않고 D-2 로 리더 결정을 남겨요.
