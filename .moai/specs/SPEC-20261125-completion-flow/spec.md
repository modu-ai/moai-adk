---
id: SPEC-20261125-completion-flow
title: "The card completion flow: four FAIL kinds, the auditor records its verdict, a light audit that re-runs nothing, a later list and a work cap"
status: draft
revision: 1
created: 2026-11-25
parent: null
---

# The card completion flow: four FAIL kinds, the auditor records its verdict, a light audit that re-runs nothing, a later list and a work cap

## Interview

- Q1: 감사 FAIL 범위를 어디까지로 하나요? -> A: "네 가지만 FAIL — 동작이 기준과 다름, 데이터 손실, 보안, 게이트 약화. 테스트 증거 부족·문서 불일치·문구는 메모로 남기고 통과. 두 번째 FAIL 은 재감사 대신 리더가 결정."(사용자 결정 2026-11-24T21:52:39Z, 카드 노트 8784 항목 1)
- Q2: SPEC 승인은 어디까지 받나요? -> A: "위험한 변경(삭제·권한·게이트·HOME 을 건드리는 tier 2)만 승인. 그 밖은 카드 본문과 기준이 설계이고 레인은 승인을 기다리지 않음."(같은 노트 항목 2)
- Q3: 착지 경로는요? -> A: "주 체크아웃이 깨끗할 때 card land 가 브랜치와 작업 트리를 함께 앞으로 옮기도록 코드를 고침."(같은 노트 항목 3; 코드 변경이라 이 SPEC 밖이에요)
- Q4: 작업 수 상한은요? -> A: "작업 중과 검사 대기를 합쳐 레인 수의 두 배까지. 차면 쉬는 레인은 새 카드 대신 검사와 착지를 도움."(같은 노트 항목 4)
- Q5: 리더가 이 카드에 무엇을 맡겼나요? -> A: "규칙 글을 새 완료 흐름에 맞게 고쳐요. 더하지 말고 빼는 변경이에요. 승인 대기로 멈추지 말고 진행해요."(리더 배차, 카드 노트 8785; 승인 근거는 위 사용자 결정 노트예요)

## Intent

12시간 착지 36장 중 19장이 FAIL 을 받았고, 72시간 동안 FAIL 123건 가운데 42건은 동작 문제가 아니었어요. 결정 요청 209건 중 141건은 SPEC 승인이었어요. 규칙 글이 "조금이라도 어긋나면 FAIL"(`workflows/audit.md` 의 절차 2)과 "조건이 있으면 FAIL"(`CLAUDE.md` 의 Flow 5)이라고 말하고, 모든 변경에 승인된 SPEC 을, 모든 지적에 후속 기준을, 모든 새 테스트에 뮤턴트를 요구해서 감사와 되돌림이 늘었어요.

이 SPEC 은 규칙 글만 고쳐요: 감사 FAIL 의 범위를 좁히고, 감사자가 판정을 직접 기록하고, 가벼운 감사는 영수증을 읽고 아무것도 다시 돌리지 않고, SPEC 승인은 tier 2 에만 두고, 지적과 되풀이는 `docs/later.md` 한 곳에 한 줄씩 적고, 작업 중인 카드 수에 상한을 둬요. 읽는 사람: 레인, 감사자, 리더.

## Non-goals

- 코드와 도구는 바꾸지 않아요. 착지가 브랜치와 작업 트리를 함께 옮기는 것(사용자 결정 3)과 `spec/` 가지에서 SPEC 이 없으면 코드 쓰기를 거절하는 훅의 완화는 각각 다른 카드의 코드 변경이에요. 훅이 그대로인 동안 tier 2 가 아닌 변경의 `spec/` 가지는 SPEC 파일이 필요해요: 이 SPEC 은 그 불일치를 규칙 글에 적고 `docs/later.md` 에 올려요.
- 감사 모델(Sonnet/Opus) 변경은 이 SPEC 이 하지 않아요. 기존 FAIL 표본의 재분류는 다른 카드가 해요.
- tier 바닥 목록(`audit.md` 의 Audit tiers 절의 접두 목록)은 바꾸지 않아요.
- 개선 효과의 측정(착지 시간, FAIL 비율, 감사 토큰)은 리더가 `measure.py` 로 재요. 이 SPEC 은 규칙만 말해요.

## Requirements

<!-- This SPEC adds no blocking check. -->
- REQ-1: When an audit finds anything other than behaviour that differs from the card's criteria, data loss, a security problem or a weakened gate, the auditor shall write it as a non-blocking line of the verdict note and give the verdict PASS. · failure: none — a rule text, nothing runs.
- REQ-2: When a card has been given a second FAIL, the leader shall decide the next step instead of a re-audit. · failure: none — a rule text, nothing runs.
- REQ-3: When an audit ends, the auditing session shall record the verdict itself with `moai card verdict`, never the leader for it. · failure: none — a rule text, nothing runs.
- REQ-4: Where a change is not tier 2 (deleting, permissions, gates or HOME), the lane shall take the card's body and criteria as the design, with no wait for a SPEC approval.
- REQ-5: When a lane hands a card back, the READY note shall carry the `check affected --card` receipt, which a light audit reads before it runs nothing again. · failure: none — a rule text, nothing runs.
- REQ-6: Where a change is tier 2, the READY note shall carry one `guard-mutant:` line per new guard test, and the audit runs at most two named test modules.
- REQ-7: When a card lands, the landing shall close the card, while a non-blocking finding or a failed remote run becomes one line of `docs/later.md` and no criterion. · failure: none — a rule text, nothing runs.
- REQ-8: When a problem of the same kind is seen a second time, the session that sees it shall add one line to `docs/later.md` that names the cause and the change at the source. · failure: none — a rule text, nothing runs.
- REQ-9: While the cards in work and the cards waiting for a check number twice the lanes, a resting lane shall help with checks and landings instead of taking a new card, and the leader opens no card above that number.
- REQ-10: The decisions log shall record the four user decisions with their time, while the rule files say the same thing and `CLAUDE.md` stays within 200 lines.
- REQ-11: The Standing principles section of `CLAUDE.md` shall name recursive self-improvement (RSI) with its targets, its moments, its owner card, its before-and-after measurement and its direction (remove before adding, never weaken a gate, a HOLD or a user rule).
- REQ-12: The rule texts shall keep the submission signal, the independent review and the landing apart, name a landing batch and a release bundle differently, end the flow at the submission when no leader is registered, refresh the remote integration branch only for a flow whose integration branch is remote, and report the source file, the home install and a new session's load as three facts.

## Acceptance

- AC-1 (REQ-1, REQ-2): `CLAUDE.md` Flow 6, `skills/moai/references/audit.md`, `skills/moai/workflows/audit.md` and `docs/leader-rules.md` name the four FAIL kinds, say that every other finding is a non-blocking line with PASS, and send a second FAIL to the leader; the sentences "any drift … is a FAIL finding" and "conditions make it FAIL" are gone · verify: `cargo test --locked --offline --test repo_claude_md --test sdd_guidance`
- AC-2 (REQ-3): the audit contract's section "Who records the verdict" and the AUDIT workflow's verdict step say the auditing session runs `moai card verdict` and that the leader does not record it for the auditor · verify: `cargo test --locked --offline --test next_flow_skill_scan`
- AC-3 (REQ-4): `CLAUDE.md` Flow 3, `docs/lane-rules.md` and `skills/moai/workflows/work.md` ask for a SPEC approval only at tier 2 · verify: `cargo test --locked --offline --test repo_claude_md --test next_flow_skill_scan`
- AC-4 (REQ-5, REQ-6): the four tier lines are the same in the audit contract, the audit agent, RUN's Check-run rules and the AUDIT workflow's Base agents; the code line says a light audit that runs no mutant and the high-risk line names the guard mutants and at most two named modules; `CLAUDE.md`, `docs/lane-rules.md` and the flow reference ask for guard-mutant lines at tier 2 only · verify: `cargo test --locked --offline --test next_flow_skill_scan`
- AC-5 (REQ-7, REQ-8): `docs/later.md` exists with its one-line form; `CLAUDE.md`, `docs/leader-rules.md`, `docs/lane-rules.md`, the flow reference and the core rules no longer add criteria or open cards for findings and recurrences · verify: `cargo test --locked --offline --test repo_claude_md --test next_flow_skill_scan`
- AC-6 (REQ-9): `CLAUDE.md`, `docs/leader-rules.md` and `docs/lane-rules.md` state the work cap · verify: `cargo test --locked --offline --test repo_claude_md`
- AC-7 (REQ-10): the decisions log holds the four decisions with 2026-11-24T21:52:39Z, `docs/claude-md-rule-provenance.md` lists each rule that left a file, and `CLAUDE.md` stays within 200 lines · verify: `cargo test --locked --offline --test repo_claude_md`
- AC-8 (REQ-11): `CLAUDE.md` has the item "Recursive self-improvement (RSI)" with the targets, the moments, the owner-card rule, the measurement and the direction, the recurrence item fixes a second occurrence in that round and sends only what cannot be fixed to `docs/later.md`, and the decisions log and the provenance table hold the instruction with 2026-11-24T22:19:11Z · verify: `cargo test --locked --offline --test card -- recurring_problems`
- AC-9 (REQ-12): `CLAUDE.md` Flow 5 is the submission and Flow 6 the independent review, the batch and release-bundle names differ in `CLAUDE.md` and `docs/leader-rules.md`, `workflows/sync.md` step 1 and `workflows/leader.md` end at the submission without a leader, and `docs/lane-rules.md` names the three facts · verify: `cargo test --locked --offline --test repo_claude_md --test lane_completion_handshake`

## Scope

- write: `.moai/specs/SPEC-20261125-completion-flow/**`, `CLAUDE.md`, `docs/leader-rules.md`, `docs/lane-rules.md`, `docs/decisions-2026-11-08.md`, `docs/claude-md-rule-provenance.md`, `docs/later.md`, `skills/moai/references/audit.md`, `skills/moai/workflows/audit.md`, `skills/moai/workflows/work.md`, `skills/moai/workflows/run.md`, `skills/moai/workflows/sync.md`, `skills/moai/workflows/leader.md`, `skills/moai/references/flows/next-flow.md`, `skills/moai/references/internal/core-rules.md`, `agents/audit.md`(the tier lines stand word for word in these four texts), `skills/moai/references/source-manifest.json` (the hashes only) and `tools/repo_scans.json` (one scope line for `docs/later.md`), both through a line lease, the old rule table `.moai/specs/SPEC-20261123-claude-md-rewrite/rule-table.md` (rows R-23, R-87, R-88 and R-114 as moved-corrected rows), and the tests that pin the changed words: `tests/repo_claude_md.rs`, `tests/sdd_guidance.rs`, `tests/next_flow_skill_scan.rs`, `tests/card/followup_per_spec_rules.rs`, `tests/card/recurring_problems_rule.rs`
- read-only: `src/**`, the other tests, `skills/moai/references/internal/sdd.md`, `.moai/project.yaml`
- forbidden: code and tool changes, the tier floor list, `.claude/`, the real HOME

## Plan

1. SPEC 한 장을 쓰고 커밋해요. 승인은 사용자 결정 노트 8784 가 근거예요; 레인은 승인을 기록하지 않아요.
2. 규칙 글을 한 파일씩 고쳐요: 빼는 쪽으로, 같은 말은 한 곳에만 두고 다른 곳은 가리켜요.
3. 문구를 고정하는 테스트(`repo_claude_md`, `sdd_guidance`, `next_flow_skill_scan`)를 같은 변경에서 새 문구로 고쳐요. 네 곳에 글자 그대로 있는 tier 줄은 네 곳을 함께 바꿔요.
4. `cargo fmt --all` 뒤 `moai check affected --card <base>..<head>` 를 한 번 돌려 영수증을 READY 에 담아요.

## Risks

- tier 줄이 네 곳에 글자 그대로 있고 테스트가 바이트 단위로 맞춰 봐서, 한 곳만 바꾸면 테스트가 깨져요. 네 곳과 테스트를 한 커밋에 넣어요.
- 규칙 글만 바뀌고 코드는 그대로예요: `spec/` 가지 훅이 SPEC 이 없는 코드 쓰기를 여전히 거절하고, `moai bundle check` 와 `card land` 의 판정도 그대로예요. 글과 코드가 어긋난 곳은 규칙 글에 적고 `docs/later.md` 에 올렸어요.
- FAIL 을 네 가지로 좁히면 놓치는 결함이 생길 수 있어요. 사용자가 고른 비용이고, 비차단 줄이 판정 노트에 남고 `docs/later.md` 로 가요.
- 감사자가 테스트를 다시 돌리지 않으므로 영수증의 신뢰가 감사의 바탕이에요. 영수증 판정은 `check affected` 와 `card note` 게이트가 그대로 지켜요.

## Decisions

승인 근거는 사용자 결정 노트 8784(2026-11-24T21:52:39Z)예요. 되돌릴 수 있어요(규칙 글).

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 감사 FAIL 은 동작이 기준과 다름, 데이터 손실, 보안, 게이트 약화 네 가지뿐이다 | 사용자 결정 항목 1; FAIL 123건 중 42건이 동작 문제 없음 | 이전 문구로 되돌려요 |
| D-2 | 감사자가 `moai card verdict` 로 판정을 직접 기록하고 `AUDIT …` 노트의 첫 줄은 노트를 읽는 도구(`bundle check`, 릴리스 PR)를 위해 둔다 | 리더 배차 8785 항목 2; `card land` 는 `card verdict` 이벤트의 독립 PASS 를 읽어요 | 노트만 쓰던 문구로 되돌려요 |
| D-3 | SPEC 승인은 tier 2 에만 둔다 | 사용자 결정 항목 2; 결정 요청 209건 중 141건이 SPEC 승인 | 모든 변경에 승인을 요구하던 문구로 되돌려요 |
| D-4 | 가벼운 감사는 영수증을 읽고 아무것도 다시 돌리지 않는다; guard-mutant 와 감사자의 모듈 실행은 tier 2 만 | 리더 배차 8785 항목 4; 감사 하나 20만~24만 토큰 | 이전 tier 줄로 되돌려요 |
| D-5 | 지적과 되풀이는 `docs/later.md` 한 곳에 한 줄씩 적는다 | 리더 배차 8785 항목 5; 한 카드가 SPEC 의 후속 카드에 기준을 덧붙이던 규칙을 대체 | 기준 덧붙이기 문구로 되돌려요 |
| D-6 | 작업 중 + 검사 대기 카드는 레인 수의 두 배까지다 | 사용자 결정 항목 4 | 상한 문구를 지워요 |
| D-7 | CLAUDE.md 에 재귀적 자기 개선(RSI)을 적는다; 지금 고칠 수 있는 것은 그 라운드에 고치고 못 고치는 것만 나중 목록 한 줄이다 | 사용자 직접 지시 2026-11-24T22:19:11Z(리더 전달 카드 노트) | RSI 항목을 지우고 되풀이 항목을 이전 문구로 되돌려요 |
| D-8 | 감사 보고서 4절의 상충 여섯 개를 규칙 글에서 푼다: 제출 신호와 운영 ready 의 구별, 단건·2~4 묶음·출시 묶음 이름, 레인 반환과 카드 종결 분리, 리더 없을 때 제출까지, NeXT 로컬 흐름의 원격 최신화 제외, 원본·설치본·새 세션 로드 세 사실 | 리더 정정(사용자가 전한 감사 보고서 지시, 2026-11-25) | 해당 문장을 이전 문구로 되돌려요 |

## Result

규칙 글만 바꿨고 코드는 건드리지 않았어요.

- 바꾼 파일: Scope 의 write 목록 그대로예요(규칙 글, 옛 규칙표 네 행, 고정 테스트, 매니페스트 해시).
- 테스트: 아래 대상을 한 슬롯에서 돌렸고 187개가 모두 통과했어요(`repo_claude_md`, `sdd_guidance`, `next_flow_skill_scan`, `review_check_cli`, `core_rules_2876`, `lane_rules_2401`, `agent_defs_package`, `lane_completion_handshake`, `lane_next_rules`, `product_cli` 의 `skill_package`, `card` 의 `followup_per_spec`·`recurring_problems`). 카드 영수증(`check affected --card`)은 READY 노트에 있어요.
- AC-1 ~ AC-9: met. 증거는 위 테스트예요.
- 남은 불일치(코드): `spec/` 가지 훅은 SPEC 이 없는 코드 쓰기를 여전히 거절해요; `card land` 는 아직 작업 트리를 함께 옮기지 않아요; `moai bundle check` 판정은 그대로예요; 원본과 설치본의 일치를 세 사실로 보여 주는 명령은 아직 없어요. 모두 `docs/later.md` 에 올렸어요.
- 이 카드가 닫지 않는 것: 카드 기준 3·4·5(효과 측정, 시범 열 장, LANE-BLOCKED 집계)는 철회해요. 부분 완료 카드의 남은 기준을 자동으로 닫는 규칙은 넣지 않았어요.
