---
id: SPEC-20261123-lane-next-card-rule
title: "레인이 다음 카드를 가져가는 규칙을 등록된 리더 조건에 맞춤"
status: draft
revision: 1
created: 2026-11-23
---

# 레인이 다음 카드를 가져가는 규칙을 등록된 리더 조건에 맞춤

## Intent

SPEC-20261123-lane-completion-handshake 의 독립 감사(AUDIT #373, head a83a94af89, verdict PASS)가 남긴 비차단 지적 A~D 의 후속 한 조각이에요. 핸드셰이크 SPEC 은 "등록된 리더가 없을 때만 레인이 다음 카드를 스스로 가져가고, 리더가 있으면 그 리더의 배차를 기다린다"를 `next-flow.md` 두 곳과 `work.md` 예외 절에 넣었어요. 같은 규칙을 말해야 하는 다른 글 네 곳이 아직 무조건이라 서로 어긋나요. 이 SPEC 은 그 네 곳을 맞추고 테스트로 고정해요. 글만 바꾸고 명령·구조·설치 변경은 없어요.

### 읽은 것 (2026-11-23, 이 워크트리 NeXT `90d3844167`)

- `CLAUDE.md` 12줄: `LANE-BLOCKED kind=decision` 노트를 쓴 뒤 "and take the next card" 가 무조건이에요(지적 A). 45줄 흐름 9단계: "then take the next card" 가 무조건이에요(지적 A).
- `skills/moai/references/flows/next-flow.md` 42줄(Transition): 리더의 `/clear` 줄이 "nothing is open" 조건만 말하고 원장이 done 또는 released 를 보여야 한다는 조건이 없어요. 39줄에는 있지만 42줄이 일반 규칙이라 같이 써야 해요(지적 B, SPEC-20261123-next-flow-clear-text REQ-2 의 문구).
- `skills/moai/workflows/work.md` 21줄(7단계 repeat): "take the next card once, when the ledger shows the previous card done or released" 에 등록된 리더 조건이 없어요(지적 C). 23줄: "with no matching row, send nothing" 을 테스트가 고정하지 않아요. 감사 뮤턴트 M7 이 이 문장을 빼도 살아남았어요(지적 D).
- `tests/lane_completion_handshake.rs`: `BLOCKED`, ac6, ac2 가 이미 next-flow 두 곳과 work.md 예외 절을 고정해요. 이 SPEC 이 고칠 곳 네 곳은 고정되어 있지 않아요.

## Non-goals

- `E`(핸드셰이크 SPEC 크기)와 `F`(trailer-check 범위) — 감사 후속 중 글이 아닌 일이라 이 카드에 없어요.
- 전역 흐름 규칙 글(`references/home/**`)과 출력 스타일 원본 — 다른 카드가 쥐고 있어요.
- 레인이 다음 카드를 가져가는 시점 규칙 자체의 변경. 이미 정해진 규칙을 글 네 곳에 맞출 뿐이에요.

## Requirements

- REQ-1: The `LANE-BLOCKED kind=decision` sentence of `CLAUDE.md` shall tell a lane to take the next card only when no leader is registered and otherwise to wait for that leader's dispatch.
- REQ-2: Step 9 of the flow in `CLAUDE.md` shall tell a lane to take the next card only when no leader is registered and otherwise to wait for that leader's dispatch.
- REQ-3: Step 7 (repeat) of `skills/moai/workflows/work.md` shall tell a lane to take the next card once only when no leader is registered and otherwise to wait for that leader's dispatch.
- REQ-4: The general `/clear` line of the Transition part of `skills/moai/references/flows/next-flow.md` shall let the leader send `/clear` only after the ledger shows the lane's card done or released and only when nothing is open.
- REQ-5: A test shall pin the same two phrases, "no registered leader" and "wait for that leader's dispatch", in `CLAUDE.md`, `next-flow.md` and `work.md` together.
- REQ-6: A test shall pin the `work.md` sentence `with no matching row, send nothing`, so that removing it fails the test.
- REQ-7: `CLAUDE.md` shall stay within 200 lines without growing in line count.
- REQ-8: `CLAUDE.md` shall keep the gate and HOLD sentences that `shortened_claude_md_keeps_its_gate_and_hold_sentences` pins.
- REQ-9: The source manifest shall carry the new sha256 of each changed shipped file.
- REQ-10: `next-flow.md` shall stay within its 8192-byte budget.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-5): `CLAUDE.md` 12줄과 9단계에 같은 조건 문장이 있고, 세 글이 같은 두 구절을 모두 가져요 · verify: `cargo nextest run --test lane_completion_handshake -E 'test(/ac7_/)'`
- AC-2 (REQ-3, REQ-5): `work.md` 7단계가 등록된 리더 조건을 말해요 · verify: `cargo nextest run --test lane_completion_handshake -E 'test(/ac7_/)'`
- AC-3 (REQ-4): `next-flow.md` 의 일반 `/clear` 줄이 원장 조건을 말해요 · verify: `cargo nextest run --test lane_completion_handshake -E 'test(/ac8_/)'`
- AC-4 (REQ-6): `work.md` 의 `with no matching row, send nothing` 문장을 빼면 `ac2_` 가 실패해요 · verify: `cargo nextest run --test lane_completion_handshake -E 'test(/ac2_/)'` 와 뮤턴트 한 줄
- AC-5 (REQ-7, REQ-8, REQ-9, REQ-10): `CLAUDE.md` 줄 수·고정 문장과 매니페스트·단어 스캔이 통과해요 · verify: `cargo nextest run --offline --locked -E 'binary(repo_claude_md) | binary(next_flow_skill_scan) | binary(core_rules_2876) | (binary(product_cli) & test(/skill_package::/))'`

## Decisions

- D-1 (2026-11-23, 리더 배차, 되돌릴 수 있음, 분류 none): 감사 후속 A~D 를 한 카드에서 하고, E(SPEC 크기)·F(trailer-check 범위)는 글이 아니라서 하지 않아요.
- D-2 (2026-11-23, 되돌릴 수 있음, 분류 none): `CLAUDE.md` 는 tier 2 바닥 목록이라 문장 두 곳만 같은 줄 안에서 고치고 줄 수를 늘리지 않아요. 새 테스트는 `tests/lane_completion_handshake.rs` 에 두고 `tests/repo_claude_md.rs` 는 건드리지 않아요(이미 있는 고정 문장만 지켜져야 해요).

## Scope

- write: `.moai/specs/SPEC-20261123-lane-next-card-rule/**`, `CLAUDE.md`, `skills/moai/references/flows/next-flow.md`, `skills/moai/workflows/work.md`, `tests/lane_completion_handshake.rs`
- write (한 줄, 임대): `skills/moai/references/source-manifest.json` — `moai card line-lease` 로
- 카드의 쓰기 범위에는 `tests/repo_claude_md.rs` 도 있지만 이 SPEC 은 그 파일을 바꾸지 않아요(D-2).
- read-only: 나머지 전부. 코드·구조·설치·실제 HOME 변경 없음.

## Plan

1. RED: `tests/lane_completion_handshake.rs` 에 ac7(CLAUDE.md 두 곳·work.md 7단계·세 글 공통 구절), ac8(next-flow Transition `/clear` 줄), ac2 의 `with no matching row, send nothing` 바늘을 더해요. 현재 글에서 실패해야 해요.
2. GREEN: `CLAUDE.md` 두 곳, `next-flow.md` 한 줄, `work.md` 한 줄을 고쳐요.
3. 매니페스트: 줄 임대를 받아 두 파일의 digest 를 다시 써요.
4. 좁은 테스트와 가드 뮤턴트, `check affected` 한 번, READY 노트와 LANE-DONE.

## Risks

- R-1: `CLAUDE.md` 의 12줄에는 고정 문장("Documents describe requirements; …")이 있어요. 같은 줄의 끝 문장만 고쳐서 고정 문장이 그대로 남게 해요.
- R-2: `next-flow.md` 는 7712 바이트로 상한 8192 까지 480 바이트가 남아요. 추가 문장은 100 바이트 안쪽이에요.
- R-3: 단어 스캔이 금지하는 낱말(`resident leader`, `leader mode`)을 새 문장에 쓰지 않아요.

## Result

Status: SPEC draft (revision 1); 승인은 사용자 몫이에요. 글 네 곳과 테스트 한 파일, 매니페스트 두 줄을 바꿨어요.

- `CLAUDE.md`: 12줄과 9단계의 끝 문장만 바꿨어요(`with no registered leader take the next card, otherwise wait for that leader's dispatch`). 줄 수는 120줄 그대로이고 12줄의 고정 문장은 그대로예요.
- `skills/moai/references/flows/next-flow.md`: Transition 의 일반 `/clear` 줄에 `only after the ledger shows the lane's card done or released` 를 되살렸어요(7712 → 7778 바이트, 상한 8192).
- `skills/moai/workflows/work.md`: 7단계 repeat 에 등록된 리더 조건과 `wait for that leader's dispatch` 를 더했어요.
- `skills/moai/references/source-manifest.json`: 두 파일의 sha256 두 줄(`tools/source_manifest.py write`, 바뀐 항목은 두 개뿐이에요).
- `tests/lane_completion_handshake.rs`: ac7(세 글과 세 곳 문장), ac8(`/clear` 줄)을 더하고 ac2 에 `with no matching row, send nothing` 바늘을 더했어요. RED 는 `target/next-card-rule/red.log`(7개 중 ac7·ac8 실패).
- AC-1: met — `cargo nextest run --offline --locked -E 'binary(lane_completion_handshake) | binary(repo_claude_md) | binary(next_flow_skill_scan) | binary(core_rules_2876) | (binary(product_cli) & test(/skill_package::/))'` 108 passed, 0 failed (`ac7_`); 뮤턴트 M2·M3 이 `ac7_` 에서 exit 100.
- AC-2: met — 같은 실행의 `ac7_`; 뮤턴트 M4 가 `ac7_` 에서 exit 100.
- AC-3: met — 같은 실행의 `ac8_`; 뮤턴트 M5 가 `ac8_` 에서 exit 100.
- AC-4: met — 같은 실행의 `ac2_`; 뮤턴트 M1(문장 제거)이 `ac2_` 에서 exit 100.
- AC-5: met — 같은 실행에서 `repo_claude_md`, `next_flow_skill_scan`, `core_rules_2876`, `product_cli skill_package::` 가 전부 통과해요. 한도는 `wc -l CLAUDE.md` 120, `wc -c next-flow.md` 7778.
- 이 SPEC 이 하지 않은 일: 감사 후속 E(핸드셰이크 SPEC 크기)·F(trailer-check 범위), 전역 흐름 규칙 글과 출력 스타일 원본(다른 카드).
