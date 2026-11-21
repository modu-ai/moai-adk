---
id: SPEC-20261120-mode-guidance-rules
title: "Lane and leader guidance: background long commands, waiting lanes take the next card, compact on request, no hand-written times, supervision tools, repair evidence"
status: draft
revision: 1
created: 2026-11-20
---

# Lane and leader guidance: background long commands, waiting lanes take the next card, compact on request, no hand-written times, supervision tools, repair evidence

## Intent

사용자 지시(2026-11-20): 리더·에이전트 모드와 워크플로에서 고칠 부분을 찾아 스킬과 지침을 고쳐요. 오늘 레인을 지켜본 기록(보고서 `skills-agents-gap-2026-11-20.md` §6)에서 되풀이된 일이 여섯 가지고, 같은 날 사용자가 일곱 번째를 더했어요.

1. `mode-agent.md` 가 기다릴 때 "`moai wait` 를 앞에서 돌리고 턴을 끝내지 말라"고 해서, 레인이 앞선 명령에 화면을 몇 시간 쥐고(한 레인 3시간, 다른 레인 24분) 리더가 보낸 메시지가 쌓였다가 사라졌어요. 규칙: 2분 넘게 걸릴 명령은 호스트 백그라운드 작업으로 돌리고 레인은 다른 일을 이어 가요. 앞에서 도는 명령은 120초 이하의 시간 제한을 둬요.
2. 검사·감사·병합·파일 반환·짝 줄·열차 차례를 기다리는 레인은 놀고 있는 레인이에요(오늘 30~60분을 배차 없이 기다렸어요). 사용자 지시(2026-11-21): 기다리지 않고 전환해요. 기다리게 된 카드에 `RESUME` 노트를 남기고, 예약을 반환하고(`/clear` 뒤에는 세션 id 가 바뀌어 옛 예약을 못 다뤄요), `SWITCH <카드> → <다음>` 한 줄을 남기면, 리더가 `/clear`(Claude Code) 또는 `/new`(Codex CLI)와 다음 카드를 보내요. 기다리던 신호가 오면 리더가 그 카드를 다시 배차하고 새 세션이 `RESUME` 노트를 읽고 다시 잡아요. 한 레인이 동시에 쥐는 카드는 셋까지이고 카드마다 임대를 갱신해요. 턴을 끝내는 것은 기다리는 것이 아니고 금지예요.
3. 맥락이 한도(호스트가 알려 주는 값의 80% 이상)에 가까운 레인은 `LANE-COMPACT` 를 적고, 리더가 `/compact` 와 이어 가기 줄을 보내요.
4. 노트에 손으로 시각을 적지 않아요. 레인이 한국 시각에 `Z` 를 붙여(1~9시간 어긋남) 단계 시간 계산을 깨뜨렸어요. `date -u` 출력을 붙여 넣거나 시각을 빼요.
5. 리더의 감독 단계가 `moai lane-screen --scan` 과 `moai card timeline` 을 이름으로 가리켜요.
6. 수리 후 READY 줄은 지적마다 고치기 전·후의 재현을 싣고, 쓰기 범위 밖의 일은 `Not done` 아래 그 파일을 쥔 카드와 넘김 노트를 적어요.
7. 사용자 지시(2026-11-20): 컨텍스트 창이 가득 찰 때까지 끌지 않아요. 카드 하나가 끝나면(READY·MERGE-READY 노트까지) 레인은 `LANE-DONE` 을 남기고, 리더가 `/clear` 로 세션을 비운 뒤 다음 카드를 줘요. 한계에 가까우면 `/compact` 로 정리하고 이어 가요. Claude Code 와 Codex CLI 둘 다 같은 규칙이에요(명령 이름은 두 호스트의 공식 문서로 확인, Decisions D-4).

이 변경은 문서와 그 문서를 읽는 검사만 바꿔요. 사용자마다 같은 지침이고 프로젝트의 규칙이 아니에요.

## Non-goals

- `SKILL.md` 문장: 그 파일을 쥔 카드가 있어서 이 SPEC 의 쓰기 범위 밖이에요. 필요한 한 문장을 그 카드에 노트로 넘겨요(Plan).
- `card note` 가 먼 시각을 경고하는 코드: 따로 있는 카드의 일이에요.
- 새 명령이나 새 훅: 이미 있는 `moai lane-screen`, `moai card timeline`, `moai lane-safe-to-reset` 만 이름으로 가리켜요.

## Requirements

- REQ-1: The MODE-AGENT workflow shall tell a lane to run a command expected to take longer than two minutes as a host background task and to give a command it runs in the foreground a timeout of at most 120 seconds. · failure: none
- REQ-2: The MODE-AGENT workflow shall no longer tell a lane to run `moai wait` in the foreground. · failure: none
- REQ-3: The MODE-AGENT workflow shall tell a lane whose card would wait for a check or an audit another session runs, a file another card holds, a pair line or a train turn to leave a `RESUME` note and stop its own background jobs, give the reservation back, arm its wake, print `SWITCH <card> → <next>` as the one line a turn may end on while a card waits, and wait for the leader's `/clear`, never naming a card in `review` (it keeps the session) and holding at most three cards whose leases it renews. · failure: none
- REQ-4: The MODE-LEADER workflow shall tell the leader to answer a `SWITCH` line with `/clear` (Claude Code) or `/new` (Codex CLI) and the next card, and to dispatch the earlier card again when the awaited signal arrives. · failure: none
- REQ-5: The MODE-AGENT workflow shall tell a lane near its context limit to post `LANE-COMPACT` once its resume note is current and to start again with an `ACK <lane> resume` line. · failure: none
- REQ-6: The MODE-LEADER workflow shall tell the leader to answer `LANE-COMPACT` with `/compact` and one resume line, and to count the lane as working only after it has read the ACK. · failure: none
- REQ-7: The MODE-AGENT workflow shall forbid typing a time into a card note, a report line or a resume note and tell the lane to paste the output of `date -u` or leave the time out. · failure: none
- REQ-8: The MODE-LEADER `supervise` step shall name `moai lane-screen --scan` and `moai card timeline` as the tools to read the lanes and a card's history. · failure: none
- REQ-9: The MODE-AGENT `report` step shall tell a lane that a repair READY line carries, for each finding, the reproduction before and after the fix and lists work outside the write scope under `Not done` with the holder of those files and the handoff note. · failure: none
- REQ-10: The MODE-AGENT workflow shall tell a lane that finishes a card (its `READY` or `MERGE-READY` note posted, its resume note current, nothing uncommitted or unpushed in its worktrees) to post `LANE-DONE` and wait for the leader's `/clear` before it takes a card from `moai card next`, and to go on with `self-pick` in the same context when no `/clear` has arrived by its next wake. · failure: none
- REQ-11: The MODE-AGENT workflow shall name the reset and compact commands of both hosts, Claude Code `/clear` and `/compact`, Codex CLI `/new` (or `/clear`) and `/compact`, with the context limit at 80% or more of the figure the host shows. · failure: none
- REQ-12: The MODE-LEADER workflow shall tell the leader to answer a `LANE-DONE` line with `/clear` and the next dispatch when `moai lane-safe-to-reset` answers `safe`, to answer `LANE-COMPACT` with `/compact` and one resume line, and to read the lane's `ACK … resume` line before it counts the lane as working. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): the long-command rule is in MODE-AGENT and the foreground-wait sentence is gone.
  verify: `cargo test --locked --offline --test mode_guidance_scan agent_runs_long_commands`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6): the waiting-lane and compact rules are in both workflows.
  verify: `cargo test --locked --offline --test mode_guidance_scan waiting_is_switching`
  verify: `cargo test --locked --offline --test mode_guidance_scan compact_request`
- AC-3 (REQ-7, REQ-8, REQ-9): no hand-written times, the supervision tools by name, the repair evidence line.
  verify: `cargo test --locked --offline --test mode_guidance_scan no_hand_written_times`
  verify: `cargo test --locked --offline --test mode_guidance_scan supervision_tools`
  verify: `cargo test --locked --offline --test mode_guidance_scan repair_evidence`
- AC-4 (REQ-10, REQ-11, REQ-12): a finished card ends in `/clear`, a full context in `/compact`, for both hosts, in both workflows.
  verify: `cargo test --locked --offline --test mode_guidance_scan session_reset`
- AC-5: neighbours.
  verify: `cargo test --locked --offline --test mode_guidance_scan`
  verify: `cargo test --locked --offline --test agent_defs_package`

## Scope

- write: `.moai/specs/SPEC-20261120-mode-guidance-rules/**`, `tests/mode_guidance_scan.rs`, `tests/mode_workflow_steps.rs` and `tests/mode_cards_53.rs` (existing tests that pin sentences this change rewrites, and the card action list that lacked `timeline`), `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/mode-leader.md`, `skills/moai/references/source-manifest.json` (regenerated after the two workflow edits)
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

MODE-AGENT, new section `## Waiting is switching` after `## Long commands`:

- Never wait. When a card of yours would have to wait for a check or an audit that another session runs, a file another card holds, a pair line or a train turn, first finish whatever else the card still has, then do these in order: (1) leave a `RESUME` note on the card with what it waits for, the watch id, the next step, the worktree, the branch and the head, and stop your own background jobs and watches (the note names them, the next session arms them again), because `moai lane-safe-to-reset` answers `not_safe` while one runs; (2) give the reservation back with `card release`, because `/clear` gives the session a new id that cannot touch the old reservation; (3) print `SWITCH <card> → <next>` as the last line of the screen and add it as a note on the card; (4) end the turn on that line and wait for the leader to send `/clear` (Claude Code) or `/new` (Codex CLI) and the next card, which the host runs only after the turn has ended, after arming the wake of the `wake` step so that a `/clear` that never comes does not leave you stopped. The `SWITCH` line is the only line a turn may end on while a card waits; ending a turn on a wait without it is forbidden.
- A wait that one of your own jobs carries is not switched away, because `/clear` would end the job: a check or an audit agent this session started runs in the background (`Long commands`), you go on with other work of the card, and its completion starts your next turn.
- A card in `review` is never named in a `SWITCH` line: `card release` refuses it and a new session cannot take it until its lease ends (the ledger answers `held`; measured by the auditor on a temporary ledger with `release`, `renew`, `take`, `reopen` and `reclaim` from a second session), so the card keeps this session; renew its lease and go on with the next card of `moai card next` in the same context until the card is done or reopened.
- If no `/clear` or `/new` has arrived by your next wake, go on with `self-pick` in the same context and skip the cards you gave back until the signal they wait for has come.
- When the signal you waited for comes, the leader dispatches the earlier card again, and the new session reads its `RESUME` note and takes the card again.
- A lane holds at most three cards at once and renews the lease of each (`card renew`).

MODE-AGENT, new section `## Session reset` after `## Waiting is switching`:

- One context does not carry many cards. A card is finished for this session when its `READY` or `MERGE-READY` note is posted, its resume note is current, each of its worktrees shows nothing uncommitted or unpushed and no card of yours is in `review` (a `/clear` would leave it held by a session that no longer exists): post `LANE-DONE`, then wait for the leader's `/clear` and take the next card only afterwards, from the leader's dispatch or from `moai card next`.
- If no `/clear` has arrived by your next wake (`wake` step), go on with `self-pick` in the same context, so the wait never leaves the lane idle.
- Claude Code: `/clear` starts a new conversation with empty context, and `/compact [instructions]` summarizes the conversation so far. Codex CLI: `/new` starts a new chat in the same session (`/clear` also starts one and clears the terminal view first), and `/compact` summarizes the visible chat to free tokens. At 80% or more of the context figure the host shows, post `LANE-COMPACT` (`self-pick` step).
- After a `/clear` the session id is new: read `moai card list`, the `RESUME` notes of your cards and the dispatch, and make your first line `ACK <lane> resume <ref> <UTC>`.

MODE-AGENT `report` step: "A `LANE-DONE` line is a progress record, not a stop: after writing it, follow `Session reset` and go on in the same turn, and never make it the turn's last line unless the `keep-going` step lets the turn end."

MODE-AGENT `Watching on your own`, third rung: "(3) in the turn, run `moai wait` as a background task (`Long commands`), go on with other work in the same turn and do not end the turn on it; when nothing else is left, follow `Waiting is switching`".

MODE-AGENT, old sentences that told a lane to take other work at once (the audit found them next to the new rule): the `claim` step's reply wait (`go on with other work of the card in the same turn, or follow Waiting is switching when it has none`), the `heavy-checks` step (`run it as a background task, go on with other work of the card … follow Waiting is switching when only the wait is left: a wait is not work`), the `keep-going` step's `waiting:` line (`once the card has nothing else to do, follow Waiting is switching; when the reply comes, the card is dispatched again`), the `hand-off` step (`Then do not wait for the merge: follow Waiting is switching`, a card in `review` stays held) and the `report` step's `LANE-BLOCKED` sentence (`then the lane follows Waiting is switching`).

MODE-AGENT `self-pick`, after the resume-note paragraph:

- Waiting is switching (new section `## Waiting is switching` after `## Long commands`): see the sentences of that section below.
- A lane whose context is at 80% or more of the figure the host shows posts `LANE-COMPACT <lane> card=<C-n> resume=<note ref>` at a card boundary, once its resume note is current, then waits for the leader's `/compact` and resume line; its first line afterwards is `ACK <lane> resume <ref> <UTC>`.

MODE-AGENT `report` step, appended: "A repair READY line, after a FAIL, carries for each finding the reproduction before and after the fix (the command and its result on the failing head and on the repaired head), and lists work outside the write scope under `Not done` with the holder of those files and the handoff note you wrote."

MODE-LEADER `supervise` step, after the sentence about a lane waiting on CI or a local check:

- A lane that posted `SWITCH <card> → <next>` gets `/clear` (Claude Code) or `/new` (Codex CLI) once `moai lane-safe-to-reset` answers `safe` and the card the line names is released, and then, as the first message after the reset, the next card from the urgent queue or `moai card next`; when the awaited signal arrives, dispatch the earlier card again (its `RESUME` note says where it stopped), and count the `SWITCH` lines in the round summary. A lane that holds a card in `review` gets no `/clear` or `/new`: `card release` refuses that card and a new session cannot take it until its lease ends, so it keeps its session and goes on with its next pick until the card is done or reopened. A lane that posted `LANE-DONE` for a card at `READY` or `MERGE-READY`, with none of its cards in `review`, gets `/clear` (Claude Code) or `/new` (Codex CLI) as the next bullet says.
- A lane that posted `LANE-COMPACT` gets `/compact` when its input box is empty and `moai lane-safe-to-reset` does not answer `not_safe`, then one resume line that names its resume note; count it as working only after you have read its `ACK … resume` line.
- A lane that posted `LANE-DONE` for a card at `READY` or `MERGE-READY` gets `/clear` once `moai lane-safe-to-reset` answers `safe`, and its next dispatch as the first message after the reset (`clear_redispatch`); a `not_safe` or `unknown` answer is a line to the lane that names what to push or note first. Claude Code and Codex CLI both take `/clear` and `/compact` (MODE-AGENT `Session reset`).
- Read the lanes with `moai lane-screen --scan` (one state per terminal) and a card's history with `moai card timeline <card>` before you give an action. The commands you run yourself follow MODE-AGENT `Long commands`.

## Risks

- The two workflow files and the manifest are held by another card; the sentences wait for it, and the RED test stays red until they are applied. If that card changes the same lines, its text and this one are merged by hand, with the scan test as the check.
- A lane that cannot type `/clear` into its own session depends on the leader sending it; REQ-10 therefore ends the wait at the lane's next wake instead of blocking.
- Substring scans pin wording, not behaviour; a reworded sentence has to keep the named words (`SWITCH`, `LANE-COMPACT`, `Long commands`, the tool names).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A new `## Long commands` section, not a change inside a step | One place a lane reads before any wait; the steps stay untouched | Move the text into `heavy-checks` |
| D-2 | A lane holds at most three cards at once and gives a waiting card back before it switches | User instruction 2026-11-21: after `/clear` the session id is new and an old reservation cannot be handled, so the card is released with a `RESUME` note first | Change the number in one sentence |
| D-3 | Hand the `SKILL.md` sentence over as a note | The file is held by another card | Apply it after its holder releases |
| D-4 | After every finished card the leader sends `/clear`; at 80% of the context the leader sends `/compact`; both hosts, same words | User instruction 2026-11-20. Command names read in the vendors' own pages on 2026-11-20 between the clock reads 21:02:34Z and 21:03:32Z (`date -u`) and 21:04:36Z. Claude Code, code.claude.com/docs/en/commands: "`/clear [name]` Start a new conversation with empty context." and "`/compact [instructions]` Free up context by summarizing the conversation so far." Codex CLI, learn.chatgpt.com/docs/cli/slash-commands (also reached from developers.openai.com/codex/cli/slash-commands): "Clear the terminal and start a fresh chat." (`/clear`), "Start a new chat inside the same CLI session." (`/new`), "Summarize the visible chat to free tokens." (`/compact`). The Codex page shows `/clear` and `/new` with a name argument but `/compact` with none, so the text sends the resume line as a message after a Codex `/compact` | Reword one sentence if either vendor renames a command |
| D-5 | After `/clear` a lane continues from its resume notes and its `ACK` line; early takeover of a reservation held under the old session id is not in this SPEC | Taking over a reservation early is the work of the card that adds that path; this text names only paths that exist (`card handoff`, `card reclaim` after the lease ends) | Add one sentence when the path ships |

## Result

Checks (2026-11-21, head `e433afdb0d`, base `efedc6608d`, the audited mode-file head the branch merged):

- `moai-cli check affected efedc6608dce..e433afdb0df7` in the worktree under a heavy slot: first line `result=PASS`, `targets=42 py=8 scans=listed:41 diff_check=0 fmt=0 clippy=0`, exit 0. Its log (`06-test_targets.log`) shows `mode_guidance_scan` running 7 tests, all ok, and `agent_defs_package` 19 passed.
- Red before: the same scan strings were looked up in the base files with a script (no cargo run): every one of the 7 tests has missing words in `efedc6608d`'s `mode-agent.md` and `mode-leader.md` (3 to 15 missing words per test), so each test fails on the base.
- `python3 -I tools/source_manifest.py check`: ok after `write` changed only the two workflow rows.
- `git diff --check`, `rustfmt --check tests/mode_guidance_scan.rs` and `spec check` printed no error.

- AC-1: met — `agent_runs_long_commands_in_the_background` ok in the check log; `Long commands` is in MODE-AGENT and the foreground-wait sentence of the watch step now says `as a background task`.
- AC-2: met — `waiting_is_switching_and_the_leader_answers_it` and `compact_request_is_posted_by_the_lane_and_answered_by_the_leader` ok; the waiting rule is `Waiting is switching` (RESUME note, give the reservation back, `SWITCH` line, the leader's `/clear` or `/new`, three cards at most) in MODE-AGENT and the answer to a `SWITCH` line in MODE-LEADER.
- AC-3: met — `no_hand_written_times_in_notes`, `supervision_tools_are_named_in_the_leader_workflow` and `repair_evidence_is_asked_for_in_the_report_step` ok.
- AC-4: met — `session_reset_follows_every_card_and_compact_follows_a_full_context` ok; Claude Code `/clear`, `/compact`; Codex CLI `/new` (and `/clear`), `/compact`, from the documents read on 2026-11-20.
- AC-5: met — the whole `mode_guidance_scan` target (7 tests) and `agent_defs_package` (19 tests) pass in the same check.

Audit repair (the independent audit of head `6b6e95cb42` recommended FAIL; repaired in `89a95fc8ce` and `399c9e1e5e`):

- The first `check affected` selected 42 targets and none of the tests that pin the workflow sentences, so its `PASS` hid four red tests (`tests/mode_workflow_steps.rs` 3, `tests/mode_cards_53.rs` 1). The auditor read the selection with `check affected <base>..<head> --list`. This repair therefore runs every test file that names the two workflow files (19 files: `cargo test --locked --offline --no-fail-fast` with `--test` for each of `agent_defs_package`, `claim_renew_doc_3042`, `claim_verify_cli`, `claim_verify_flags_2838`, `close_step_2903`, `flaky_followups`, `lane_next_rules`, `lane_rules_2401`, `lane_rules_2836`, `lane_rules_2879`, `mode_cards_53`, `mode_guidance_scan`, `mode_workflow_steps`, `orca_alignment`, `orca_optional_3005`, `product_cli`, `rule_history_guard`, `self_audit_cli`, `workflow_size_budget`): on head `89a95fc8ce` 250 passed, 3 failed (`both_modes_name_every_card_action_of_the_round`, `blocked_lines_name_kind_owner_and_next_check`, `ops_named_by_the_mode_workflows_exist_with_their_flags`), all three in the two files I had widened the scope for; after `399c9e1e5e` the three changed or related targets read `mode_cards_53` 9 passed, `mode_guidance_scan` 8 passed, `mode_workflow_steps` 43 passed, exit 0 (the other 16 targets were not run again; the workflow text did not change between the two runs).
- B1 (`mode_workflow_steps`): the pinned sentences follow the new text (`do not end the turn` stays in the watch step; the `LANE-DONE` sentence and the `LANE-BLOCKED` sentence say `follow Session reset` and `follow Waiting is switching`; the op list gains `check`, `lane-safe-to-reset` and `lane-screen --scan`).
- B2 (`mode_cards_53`): `timeline` is allowed in a code span (`LEADER_ONLY`) without being required of both files.
- B3: `card release` refuses a card in `review` (`src/ledger/reservation.rs`), so the rule keeps such a card held and writes `review` after its name in the `SWITCH` line; the leader answers when the card is released or still in `review`. Not run against a live ledger (the ledger code was read, not executed).
- B4: five old sentences that told a lane to take other work at once now point to `Waiting is switching` (claim, heavy-checks, keep-going, hand-off, report); a turn may end only on the `SWITCH` line because the host runs a queued `/clear` after the turn; the lane stops its own background jobs first because `lane-safe-to-reset` answers `not_safe` while one runs; with no `/clear` by the next wake the lane goes on with `self-pick`. New scan tests pin the order (1)<(2)<(3)<(4), the safety words and the absence of the old sentences.
- Non-blocking: the leader's context figure is 80% like the lane's (N1), the `lane-done` step names `/clear` or `/new` (N3), `Long commands` names a detached start for hosts without a background task (N5), the pinned words that survived the auditor's mutants are in the new scan test (N4).

Not done: `moai-cli check affected` on the final head is run after this edit; no live lane was sent `/clear` or `/new` (the leader's round code reads the `SWITCH` line later); the two workflow files still carry older issue and decision citations that the citation SPEC removes when it lands, and the train merge resolves the overlap.
