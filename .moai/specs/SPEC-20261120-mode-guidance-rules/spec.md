---
id: SPEC-20261120-mode-guidance-rules
title: "Lane and leader guidance: background long commands, waiting lanes take the next card, compact on request, no hand-written times, supervision tools, repair evidence"
status: draft
revision: 1
created: 2026-11-20
---

# Lane and leader guidance: background long commands, waiting lanes take the next card, compact on request, no hand-written times, supervision tools, repair evidence

## Intent

사용자 지시(2026-11-20): 리더·에이전트 모드와 워크플로에서 고칠 부분을 찾아 스킬과 지침을 고쳐요. 오늘 레인을 지켜본 기록(보고서 `skills-agents-gap-2026-11-20.md` §6)에서 되풀이된 일이 여섯 가지예요.

1. `mode-agent.md` 가 기다릴 때 "`moai wait` 를 앞에서 돌리고 턴을 끝내지 말라"고 해서, 레인이 앞선 명령에 화면을 몇 시간 쥐고(한 레인 3시간, 다른 레인 24분) 리더가 보낸 메시지가 쌓였다가 사라졌어요. 규칙: 2분 넘게 걸릴 명령은 호스트 백그라운드 작업으로 돌리고 레인은 다른 일을 이어 가요. 앞에서 도는 명령은 120초 이하의 시간 제한을 둬요.
2. 검사·감사·병합을 기다리는 레인은 놀고 있는 레인이에요(오늘 30~60분을 배차 없이 기다렸어요). 다음 카드를 두 번째 워크트리에서 잡고(동시에 둘까지) `LANE-WAIT` 를 적어요. 리더는 기다리는 레인을 배차 후보로 읽어요.
3. 맥락이 한도에 가까운 레인은 `LANE-COMPACT` 를 적고, 리더가 `/compact` 와 이어 가기 줄을 보내요.
4. 노트에 손으로 시각을 적지 않아요. 레인이 한국 시각에 `Z` 를 붙여(1~9시간 어긋남) 단계 시간 계산을 깨뜨렸어요. `date -u` 출력을 붙여 넣거나 시각을 빼요.
5. 리더의 감독 단계가 `moai lane-screen --scan` 과 `moai card timeline` 을 이름으로 가리켜요.
6. 수리 후 READY 줄은 지적마다 고치기 전·후의 재현을 싣고, 쓰기 범위 밖의 일은 `Not done` 아래 그 파일을 쥔 카드와 넘김 노트를 적어요.

이 변경은 문서와 그 문서를 읽는 검사만 바꿔요. 사용자마다 같은 지침이고 프로젝트의 규칙이 아니에요.

## Non-goals

- `SKILL.md` 문장: 그 파일을 쥔 카드가 있어서 이 SPEC 의 쓰기 범위 밖이에요. 필요한 한 문장을 그 카드에 노트로 넘겨요(Plan).
- `card note` 가 먼 시각을 경고하는 코드: 따로 있는 카드의 일이에요.
- 새 명령이나 새 훅: 이미 있는 `moai lane-screen`, `moai card timeline`, `moai lane-safe-to-reset` 만 이름으로 가리켜요.

## Requirements

- REQ-1: The MODE-AGENT workflow shall tell a lane to run a command expected to take longer than two minutes as a host background task and to give a command it runs in the foreground a timeout of at most 120 seconds. · failure: none
- REQ-2: The MODE-AGENT workflow shall no longer tell a lane to run `moai wait` in the foreground. · failure: none
- REQ-3: The MODE-AGENT workflow shall tell a lane that waits for a check, an audit or a merge to take the next card in a second worktree, at most two cards at once, and to post `LANE-WAIT <lane> on=check|audit|merge`. · failure: none
- REQ-4: The MODE-LEADER workflow shall tell the leader to give a lane that posted `LANE-WAIT` the next ready card unless it already holds two cards with a worktree. · failure: none
- REQ-5: The MODE-AGENT workflow shall tell a lane near its context limit to post `LANE-COMPACT` once its resume note is current and to start again with an `ACK <lane> resume` line. · failure: none
- REQ-6: The MODE-LEADER workflow shall tell the leader to answer `LANE-COMPACT` with `/compact` and one resume line, and to count the lane as working only after it has read the ACK. · failure: none
- REQ-7: The MODE-AGENT workflow shall forbid typing a time into a card note, a report line or a resume note and tell the lane to paste the output of `date -u` or leave the time out. · failure: none
- REQ-8: The MODE-LEADER `supervise` step shall name `moai lane-screen --scan` and `moai card timeline` as the tools to read the lanes and a card's history. · failure: none
- REQ-9: The MODE-AGENT `report` step shall tell a lane that a repair READY line carries, for each finding, the reproduction before and after the fix and lists work outside the write scope under `Not done` with the holder of those files and the handoff note. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): the long-command rule is in MODE-AGENT and the foreground-wait sentence is gone.
  verify: `cargo test --locked --offline --test mode_guidance_scan agent_runs_long_commands`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6): the waiting-lane and compact rules are in both workflows.
  verify: `cargo test --locked --offline --test mode_guidance_scan waiting_lanes`
  verify: `cargo test --locked --offline --test mode_guidance_scan compact_request`
- AC-3 (REQ-7, REQ-8, REQ-9): no hand-written times, the supervision tools by name, the repair evidence line.
  verify: `cargo test --locked --offline --test mode_guidance_scan no_hand_written_times`
  verify: `cargo test --locked --offline --test mode_guidance_scan supervision_tools`
  verify: `cargo test --locked --offline --test mode_guidance_scan repair_evidence`
- AC-4: neighbours.
  verify: `cargo test --locked --offline --test mode_guidance_scan`
  verify: `cargo test --locked --offline --test agent_defs_package`

## Scope

- write: `.moai/specs/SPEC-20261120-mode-guidance-rules/**`, `tests/mode_guidance_scan.rs`, `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/mode-leader.md`, `skills/moai/references/source-manifest.json` (regenerated after the two workflow edits)
- read-only: `skills/moai/SKILL.md`, `src/lane_screen*.rs`, `src/card/timeline.rs`
- forbidden: a card, Issue or pull request number in the sentences this change adds, and any edit of a file another card holds (the card that holds the two workflow files and the manifest must have released them first)

## Plan

1. SPEC and a RED test, `tests/mode_guidance_scan.rs` (reads the two workflow files by path, never walks the repository).
2. The sentences below, in the two workflows, when their holder has released them; regenerate the source manifest in the same commit; run the scan test and the package tests.
3. The `SKILL.md` sentence goes to its holder as a card note: "A command expected to run longer than two minutes runs as a host background task, and a foreground command carries a timeout of at most 120 seconds, so a lane or a leader stays reachable (MODE-AGENT `Long commands`)", placed after the sentence that says both modes do not end the turn.
4. A read-only audit by another agent, then READY.

### Sentences

MODE-AGENT, new section `## Long commands` after `## Watching on your own`:

- A lane that holds its screen in a foreground command cannot read the leader's messages, and queued messages were lost for hours that way.
- A command expected to run longer than two minutes (a build, a test run, `moai check affected`, an audit agent, `moai wait`) runs as a host background task (Claude Code: `run_in_background` or the Monitor tool) and you go on with other work in the same turn; its completion starts your next turn.
- A command you run in the foreground carries a timeout of at most 120 seconds.
- Never type a time into a card note, a report line or a resume note: paste the output of `date -u` from the same step, or leave the time out. Lanes wrote local times with a trailing `Z`, 1 to 9 hours off, and that broke stage timing; the ledger's `created_at` is the time of a note.

MODE-AGENT `Watching on your own`, third rung: "(3) in the turn, run `moai wait` as a background task (`Long commands`) and do not end the turn".

MODE-AGENT `self-pick`, after the resume-note paragraph:

- A lane that waits for a check, an audit or a merge is an idle lane. After its resume point it takes the next card of `moai card next` in a second worktree (at most two cards at once) and posts `LANE-WAIT <lane> on=check|audit|merge card=<C-n>` once, on the card and on its screen.
- A lane whose context is near its limit (about 85%) posts `LANE-COMPACT <lane> card=<C-n> resume=<note ref>` at a card boundary, once its resume note is current, then waits for the leader's `/compact` and resume line; its first line afterwards is `ACK <lane> resume <ref> <UTC>`.

MODE-AGENT `report` step, appended: "A repair READY line, after a FAIL, carries for each finding the reproduction before and after the fix (the command and its result on the failing head and on the repaired head), and lists work outside the write scope under `Not done` with the holder of those files and the handoff note you wrote."

MODE-LEADER `supervise` step, after the sentence about a lane waiting on CI or a local check:

- A lane that posted `LANE-WAIT on=check|audit|merge` is waiting on its first card: give it the next ready card unless it already holds two cards with a worktree, and count the waiting lanes in the round summary.
- A lane that posted `LANE-COMPACT` gets `/compact` when its input box is empty and `moai lane-safe-to-reset` does not answer `not_safe`, then one resume line that names its resume note; count it as working only after you have read its `ACK … resume` line.
- Read the lanes with `moai lane-screen --scan` (one state per terminal) and a card's history with `moai card timeline <card>` before you give an action. The commands you run yourself follow MODE-AGENT `Long commands`.

## Risks

- The two workflow files and the manifest are held by another card; the sentences wait for it, and the RED test stays red until they are applied. If that card changes the same lines, its text and this one are merged by hand, with the scan test as the check.
- Substring scans pin wording, not behaviour; a reworded sentence has to keep the named words (`LANE-WAIT`, `LANE-COMPACT`, `Long commands`, the tool names).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A new `## Long commands` section, not a change inside a step | One place a lane reads before any wait; the steps stay untouched | Move the text into `heavy-checks` |
| D-2 | At most two cards at once for a waiting lane | The card's rule (limit two); a third card makes the resume notes unreadable | Change the number in one sentence |
| D-3 | Hand the `SKILL.md` sentence over as a note | The file is held by another card | Apply it after its holder releases |

## Result

(written after the checks)
