---
id: SPEC-20261123-skill-structure
title: "The skill's four roles hold: stage workflows say how to resume, one reference says when a timer may be armed, and the stage order ends in the landing and the next card"
status: draft
revision: 1
created: 2026-11-23
---

# The skill's four roles hold: stage workflows say how to resume, one reference says when a timer may be armed, and the stage order ends in the landing and the next card

## Intent

사용자가 확정한 스킬 구조(2026-11-21)는 네 역할이에요. `SKILL.md` 는 Claude Code 의 입구와 라우팅, 경로표(`routes.json`)와 체인표(`chains/index.md`)는 단계 고르기, 워크플로는 입력·출력·단계·막힘·이어가기, `references/` 는 필요할 때만 읽는 증거와 상세 계약이에요. 경로 여덟 개와 흐름 참조 하나로 줄이는 큰 조각(`SPEC-20261122-next-flow-skills-agents`)은 이미 들어갔고, 이 SPEC 은 그 위에서 남은 빈틈 셋만 닫아요. 재서 확인한 현재 상태는 아래와 같아요(2026-11-23, 이 가지의 `c395e5a50c` 트리).

- `/loop`: `grep -rn -F "/loop" skills` 가 아무것도 찾지 못해요. 예전 규칙은 `workflows/mode-leader.md` 에 있었고 그 파일이 없어지면서 같이 사라졌어요(`SPEC-20261121-loop-leader-session-only` 의 REQ 는 없는 파일을 가리켜요). 그래서 지금 스킬 글에는 "리더로 확인된 창에서만, 한 번만 등록한다"는 말이 없어요.
- 이어가기: 입력·출력·절차·예외·완료 확인 다섯 제목은 모든 워크플로에 있는데 `## Resume` 은 하나도 없어요(`grep -c "^## Resume" skills/moai/workflows/*.md` 가 모두 0). 재시작·사용량 한도 멈춤·압축 뒤에 무엇을 다시 읽고 어디서 이어 가는지가 RUN 7단계와 WORK 1단계에만 흩어져 있어요.
- 단계 순서의 끝: `references/internal/sdd.md` 8번은 `→ merge` 에서 끝나고, 그다음은 `moai spec close` 만 말해요. 착지 명령(`moai card land` 가 통합 브랜치를 옮기고 같은 호출에서 카드를 닫는다)과 다음 카드 받기(`moai card next`)가 순서 한 곳에 없어요.
- 입구가 한 번 더 적은 것: `SKILL.md` 11행이 WORK 의 단계(카드를 받아 계획·실행·동기화·감사·착지)를 다시 풀어 써요. 그 순서의 주인은 `workflows/work.md` 와 `internal/sdd` 예요.

## Non-goals

- `SKILL.md` 의 판단 순서·HOLD 범주·결정 기록 항목을 다른 곳으로 옮기는 일. `SPEC-20261121-report-rules-in-core-rules` 의 D-1 이 일부러 입구에 두기로 했고 여러 테스트가 그 자리를 고정해요. 이 SPEC 은 그 결정을 뒤집지 않아요.
- `workflows/work.md` 와 `references/flows/next-flow.md`(다른 카드가 쥐고 있음), `references/home/**` 와 `workflows/update.md`(홈 설치 카드의 범위): 이어가기 문단과 `/loop` 지침을 거기 넣지 않아요. WORK 와 UPDATE 의 `## Resume` 은 그 카드들이 착지한 뒤의 따로 하는 일이에요.
- 레인이 사용자의 선택으로 스스로 돌리는 `/loop`(끊김 없는 레인): 이 SPEC 의 대상은 리더 라운드를 되풀이하는 타이머뿐이에요.
- 리더를 등록하는 코드(`session-register`)와 훅 규칙. 훅은 코드가 강제하니까 스킬 글에 훅 규칙을 넣지 않아요.
- 어느 단계가 무엇을 하는지를 바꾸는 일.
- 이 변경이 더하는 문장에 카드·Issue·풀 리퀘스트 번호를 넣는 일(넣지 않아요)과 실제 HOME 을 건드리는 일.

## Requirements

- REQ-1: The skill shall carry a reference `references/internal/loop.md` that names the one case in which a session may arm a repeating timer for the leader round.
- REQ-2: While a session is not confirmed as the registered leader, the loop reference shall tell it not to arm `/loop`, a cron task or a wakeup for the leader round.
- REQ-3: The loop reference shall define a confirmed leader as a session whose id, printed by `moai session-current`, equals the `session` of the live leader entry that `moai session-register --show --name leader` prints.
- REQ-4: When a confirmed leader arms a timer for the leader round, the loop reference shall tell it to arm exactly one, after it has listed its timers and removed an earlier one for the same round.
- REQ-5 (REQ-4): IF the session cannot list its timers, THEN the loop reference shall tell it not to arm one and to say so.
- REQ-6: The skill entry shall point to the loop reference in one line.
- REQ-7: The skill entry shall hold no sentence of the loop rule.
- REQ-8: The skill entry shall name the WORK route without listing its steps.
- REQ-9: Each workflow other than WORK and UPDATE shall have a `## Resume` section directly before `## Exceptions and Permissions` that names what the session reads afresh.
- REQ-10: The SDD reference shall say after the stage order that the merge is the landing with `moai card land`, which closes the card in the same call, and that the next card is taken with `moai card next`.
- REQ-11: The skill's source manifest shall match the files after the change.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): the loop reference exists, begins with its "When to read" line, refuses an unconfirmed session and defines the confirmed leader with the two commands · verify: `cargo nextest run --locked --offline --test skill_structure -E 'test(loop_reference_)'`
- AC-2 (REQ-4, REQ-5): the loop reference says to list the timers, remove the earlier one for the same round and arm exactly one, and not to arm when the list cannot be read · verify: `cargo nextest run --locked --offline --test skill_structure -E 'test(loop_reference_arms_one_timer)'`
- AC-3 (REQ-6, REQ-7, REQ-8): the skill entry has the one pointer line, carries none of the loop rule's words and does not list the WORK steps · verify: `cargo nextest run --locked --offline --test skill_structure -E 'test(skill_entry_)'`
- AC-4 (REQ-9): each of the six workflows has `## Resume` directly before `## Exceptions and Permissions`, and its text names what is read afresh · verify: `cargo nextest run --locked --offline --test skill_structure -E 'test(resume_)'`
- AC-5 (REQ-10): the SDD reference's order step names the landing command, its closing of the card and the next-card command, and the full order is still written only there · verify: `cargo nextest run --locked --offline --test skill_structure -E 'test(sdd_order_)'`
- AC-6 (REQ-1, REQ-6, REQ-9, REQ-10): the shipped-text scans and the stage-order tests still pass on the new and changed text · verify: `cargo nextest run --locked --offline --test next_flow_skill_scan --test sdd_guidance --test agent_defs_package --test core_rules_2876 --test report_rules_core`
- AC-7 (REQ-11): the manifest matches the package · verify: `python3 -I tools/source_manifest.py check`

## Scope

- write: `.moai/specs/SPEC-20261123-skill-structure/**`, `skills/moai/SKILL.md`, `skills/moai/references/internal/loop.md`, `skills/moai/references/internal/sdd.md`, `skills/moai/references/chains/index.md` (one rule line that names the loop reference), `tests/agent_defs_package.rs` (its closed list of internal references gains `loop`), `skills/moai/workflows/read.md`, `skills/moai/workflows/plan.md`, `skills/moai/workflows/run.md`, `skills/moai/workflows/sync.md`, `skills/moai/workflows/audit.md`, `skills/moai/workflows/project.md`, `tests/skill_structure.rs`, `tests/sdd_guidance.rs` (only if a pinned sentence of the order step moves), `skills/moai/references/source-manifest.json` (a leased line, written last)
- read-only: `tests/next_flow_skill_scan.rs`, `tests/report_rules_core.rs`, `tests/core_rules_2876.rs`, `tests/agent_defs_package.rs`, `src/leader_round/registration.rs`, `src/card/land.rs`
- forbidden: `skills/moai/workflows/work.md`, `skills/moai/workflows/update.md`, `skills/moai/references/flows/next-flow.md`

## Plan

1. RED first: write `tests/skill_structure.rs` with the tests of AC-1 to AC-5 and run it on the tip, where each fails for its reported reason (no `loop.md`, no pointer, no `## Resume`, no landing sentence).
2. Write `references/internal/loop.md` (English; "When to read" first line) and the one pointer line in `SKILL.md`; replace the WORK step list in `SKILL.md` by the route name and a pointer to `workflows/work.md`.
3. Add the `## Resume` section to the six workflows, each grounded in what that workflow already reads (`moai card show`, `git`, `moai spec check`, the head named by the ready note, `moai project-update-plan`).
4. Extend the order step of `internal/sdd` with the landing and the next card; keep the one full copy of the order there.
5. Run the targeted scans of AC-6, then regenerate the manifest rows under a line lease and run its check.
6. Show that the loop reference's two commands behave as it says by running the existing `leader_registration_cli` tests (they register a leader, read it back with `--show` and refuse a second live one in a temporary clone and HOME), and record the run in the Result. A session isolated in this worktree may not run git against a clone of its own making, so no extra clone is built.
7. Ask for an independent audit; the implementer does not audit.

## Risks

- A shipped-text scan refuses a word of the new text (`resident leader`, `leader mode`, a mode flag, a tool Claude Code no longer has): the new text is written without them and `next_flow_skill_scan` is run before the first commit.
- A test pins the section list of a workflow (`run.md` is counted by its `## ` headings): the test is read before the section is added, and the pin is kept or moved with a reason.
- `source-manifest.json` is held by another card: its rows are written last, under `moai card line-lease`, after the other changes are committed.
- `project.md` is large (42,873 bytes of the 65,536 bound `profile::read` puts on a package file): the Resume text stays under 700 bytes.
- The leader entry is first-come: a leader that never registered cannot arm a timer. The reference says how it registers; whether the launcher should do it is a separate card.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The decision order, HOLD categories and decision record stay in `SKILL.md` | `SPEC-20261121-report-rules-in-core-rules` D-1 (decided for this card) | Move the text and its pins in a SPEC of its own |
| D-2 | The loop rule lives in a new reference `internal/loop.md`, read on demand, not in `SKILL.md`, `work.md` or the flow reference | The entry only routes; `work.md` and `next-flow.md` are held by another change; the rule is read only before a timer is armed | Fold the file into the flow reference after that change lands |
| D-3 | Leader ownership is confirmed by the session registry: the session's own id equals the `session` of the live leader entry | Registry code (`src/leader_round/registration.rs`) records the host session id and refuses a second live leader; no other check in the tree is first-come | Name another source in the reference and its test |
| D-4 | `## Resume` is added to six workflows now; `work.md` and `update.md` follow when their holders land | The dispatch forbids `work.md`; `update.md` is in the scope of the home-install card | Add both sections in one later SPEC |
| D-5 | The SPEC stays `draft`: only the user's selection approves it | `internal/sdd` item 6 | None needed |

## Result

Not filled yet: SYNC writes it after the checks run.
