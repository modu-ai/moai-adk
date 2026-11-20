---
id: SPEC-20261120-protected-paths-one-list
title: "Protected check paths: one list, pinned by a test"
status: draft
revision: 1
created: 2026-11-20
---

# Protected check paths: one list, pinned by a test

## Intent

검증 2차 W2-2 예요. 지금 "이 파일을 바꾸면 감사 등급 2" 라는 목록이 둘이에요. `check-guard`(`tools/check_guard.py`)는 기준 가지의 `tools/protected_checks.json` 을 읽고, 병합 게이트는 `src/gate.rs` 의 `TIER2_PREFIXES` 로 실효 등급(`effective_tier`)을 정해요. 두 목록은 서로 달라요. `tools/ci_scope.py`, `tools/check_guard.py`, `tools/protected_checks.json`, `src/audit_gate.rs`, `REVIEW.md`, `agents/audit.md` 는 보호 목록에만 있어서, 본문에 `audit-tier: 0` 이라고 쓴 PR 이 이 파일을 바꿔도 게이트의 경로 규칙은 등급 2 로 올리지 않아요(읽은 코드: `src/gate.rs` 의 `tier_of` 는 `TIER2_PREFIXES` 의 접두어만 봐요). 어느 목록에도 없는 판정용 파일도 있어요: `tools/repo_scans.json`(저장소 전체 스캔 표), `tools/self_check.sh`(자체 점검 스크립트), `skills/moai/workflows/audit.md`(감사 절차), `AGENTS.md`(모든 에이전트의 규칙). 목록이 둘이면 한쪽만 고쳐져요. 이 SPEC 은 보호 목록의 모든 항목이 게이트에서 등급 2 로 나오는지 고정하는 시험 하나와, 빠진 판정용 파일 넷의 추가를 정해요.

## Non-goals

- `check-guard` 를 필수 검사(저장소 규칙의 required check)로 만드는 일: 저장소 보호 설정이라 사용자 결정이에요. 이 SPEC 은 규칙 설정을 읽지도 바꾸지도 않아요.
- `ci.yml` 이 PR 쪽 `tools/ci_scope.py` 대신 기준 가지의 것을 쓰게 하는 일: 별도 카드예요.
- `TIER2_PREFIXES` 의 항목을 보호 목록에 거꾸로 맞추는 일(예: `tools/local_check.sh` 를 보호 목록에 넣기): 게이트 쪽 항목 대부분은 설치·훅·HOME 이주 같은 경로 규칙이라 보호 목록과 뜻이 달라요. 이 SPEC 은 한 방향, 보호 목록 ⊆ 등급 2 만 고정해요.
- `check_guard.py` 의 매칭 규칙과 `gate::tier_of` 의 판정 방식 바꾸기: 그대로 두고 목록만 맞춰요.

## Requirements

- REQ-1: When the pull request body declares audit-tier 0, the gate shall judge every path that an entry of the protected check list matches as effective tier 2.
- REQ-2: The protected check list shall name `tools/repo_scans.json`, `tools/self_check.sh`, `skills/moai/workflows/audit.md` and `AGENTS.md`.
- REQ-3 (REQ-1): IF an entry of the protected check list matches a path that the gate does not judge as effective tier 2, THEN the list-agreement test shall fail and name that entry and that path.
- REQ-4 (REQ-1): IF an entry of the protected check list holds a pattern character that the list-agreement test cannot turn into sample paths, THEN the test shall fail and name that entry.
- REQ-5: While the list-agreement test runs, it shall read the real `tools/protected_checks.json` of the checkout and write nothing.

## Acceptance

- AC-1 (REQ-1, REQ-5): 실제 `tools/protected_checks.json` 의 항목마다(`/**` 접두어, `*` 글롭, 정확한 이름) 그 항목이 맞는 대표 경로를 만들어 본문 등급 0 으로 `effective_tier` 를 구하면 모두 2 가 나와요 · verify: `cargo test --locked --offline --test protected_paths_one_list every_protected_entry_is_judged_tier_2`
- AC-2 (REQ-2): 보호 목록이 `tools/repo_scans.json`, `tools/self_check.sh`, `skills/moai/workflows/audit.md`, `AGENTS.md` 를 정확한 이름으로 가져요 · verify: `cargo test --locked --offline --test protected_paths_one_list the_judging_files_are_in_the_protected_list`
- AC-3 (REQ-3): 게이트가 등급 2 로 보지 않는 경로를 맞추는 항목이 든 시험용 목록을 주면 검사 함수가 그 항목과 경로를 이름으로 돌려주고, 덮인 항목만 든 목록은 빈 결과를 줘요. 항목 하나를 일부러 `TIER2_PREFIXES` 에서 뺀 변이가 AC-1 시험을 실패시켜요 · verify: `cargo test --locked --offline --test protected_paths_one_list a_gap_between_the_two_lists_is_named`
- AC-4 (REQ-4): `?` 나 `[` 가 든 항목은 건너뛰지 않고 항목 이름과 함께 실패로 돌아와요 · verify: `cargo test --locked --offline --test protected_paths_one_list an_entry_the_test_cannot_sample_is_reported`
- AC-5 (REQ-1): 대표 경로는 항목이 실제로 맞는 경로여야 해요. `check_guard.py` 와 같은 규칙(`/**` 접두어 · `*` 전체 경로 글롭 · 정확한 이름)으로 다시 맞춰 보고, 항목이 맞지 않는 이웃 경로(`src/audit_gate.rs` 항목에 대한 `docs/src/audit_gate.rs`)는 본문 등급 0 에서 2 로 올라가지 않는 것까지 확인해요 · verify: `cargo test --locked --offline --test protected_paths_one_list a_sample_path_matches_its_own_entry_and_a_neighbour_is_not_raised`

## Scope

- write: 이 SPEC, `tools/protected_checks.json`, `src/gate.rs`(`TIER2_PREFIXES` 만), `tests/protected_paths_one_list.rs`(새), 이웃 시험이 깨질 때만 `tests/gate_check_cli.rs`
- read-only: `tools/check_guard.py`(매칭 규칙), `tests/check_guard_test.py`, `.github/workflows/check-guard.yml`, `src/audit_gate.rs`, `src/bundle.rs`(`tier_of` 호출처)
- forbidden: 저장소 규칙·필수 검사 설정, `.github/` 의 어떤 파일, `tools/check_guard.py` 의 판정 방식, 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 실제 HOME

## Plan

1. RED: 새 시험 파일을 먼저 써서, 목록을 고치기 전에 AC-1 이 보호 목록 항목 여섯(`tools/ci_scope.py` 등)에서 실패하는지 본다.
2. GREEN: `TIER2_PREFIXES` 에 보호 목록의 빠진 여섯 항목과 새 넷을 더하고, 보호 목록에 새 넷을 더한다.
3. 이웃 시험: 등급을 읽는 시험(`gate_check_cli`, `gate_merge`, `audit_gate_cli`, 묶음 등급 바닥을 쓰는 시험)과 `check_guard` 시험을 돌린다.
4. 변이: `TIER2_PREFIXES` 항목 하나를 빼고 AC-1 이 실패하는지, 보호 목록의 새 항목 하나를 빼고 AC-2 가 실패하는지 본다.

## Risks

- **보호 목록이 늘면 PR 쓰는 쪽이 바뀌어요.** `tools/repo_scans.json` 을 고치는 PR 은 이제 `audit-tier: 2` 를 본문에 써야 `check-guard` 를 통과하고, 게이트도 등급 2 로 봐서 리더가 감사해요. 새 시험 파일을 스캔 표에 올리는 일이 잦아서 체감이 커요. 카드가 정한 설계(스캔 표를 판정용 파일로 본다)라서 그대로 하고, 리더가 되돌릴 수 있게 결정 표에 적어요.
- `AGENTS.md` 접두어는 저장소 루트의 `AGENTS.md` 와 `AGENTS.md…` 로 시작하는 루트 파일에 맞아요. 하위 폴더의 `AGENTS.md` 는 맞지 않아요(그쪽은 기존대로 등급 1, `tier0_safe` 가 에이전트 지시 파일을 제외해요).
- 접두어 일치는 이름이 더 긴 루트 파일도 올려요(예: `REVIEW.md.bak`). 보호 항목은 정확한 이름이라 이 쪽이 더 넓은 방향(안전한 쪽)이에요.
- 보호 목록은 PR 이 아니라 기준 가지에서 읽혀요(`check-guard` 의 설계). 그래서 이 PR 이 합쳐진 다음 PR 부터 새 항목이 효력을 가져요. 이 PR 자신은 이미 `src/gate.rs` 와 `tools/protected_checks.json` 을 바꾸므로 `audit-tier: 2` 를 써요.
- 시험은 `tier_of` 대신 `effective_tier` 를 불러 `audit-gate` 와 `gate-check` 가 같은 결과를 읽게 해요. `audit_gate.rs` 는 `tier_of` 를 직접 써서, 시험은 둘 다 확인해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 한 방향만 고정해요: 보호 목록의 모든 항목은 등급 2. 게이트 쪽 항목을 보호 목록에 거꾸로 맞추지 않아요 | 카드 설계("`TIER2_PREFIXES` 가 보호 목록의 모든 항목을 덮는지 확인하는 시험 하나"), 게이트 쪽에는 설치·훅 같은 보호 목록과 뜻이 다른 항목이 있어요 | 시험을 양방향으로 넓히기(항목마다 뜻을 가려야 해서 하지 않아요) |
| D-2 | `tools/local_check.sh` 는 보호 목록에 넣지 않아요(게이트 쪽에만 이미 있어요). 판정용이 아니라 증거용 스크립트이고, `ci` 를 대신하지 않아요 | `AGENTS.local.md` ("a `local-check:` line never replaces `ci`"), 카드가 넣을 파일로 든 넷에 없어요 | 보호 목록에 한 줄 더하기 |
| D-3 | 새 시험은 `src/gate.rs` 의 `effective_tier` 와 `tier_of` 를 라이브러리로 직접 불러요(`gate-check` 명령을 띄우지 않아요) | 명령 시험은 스냅숏 입력이 커서 느리고, 두 목록의 어긋남은 함수 하나로 잡혀요 | 명령 시험으로 바꾸기 |
| D-4 | 접두어는 정확한 파일 이름으로 더해요(`AGENTS.md`, `tools/repo_scans.json` …). 폴더 접두어로 넓히지 않아요 | 보호 목록의 항목이 정확한 이름이에요 | 접두어 넓히기 |

## Result

(아직 구현 전이에요.)
