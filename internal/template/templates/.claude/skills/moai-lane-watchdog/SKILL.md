---
name: moai-lane-watchdog
description: >
  One lane stall watchdog iteration: measure lane progress on three
  channels, classify the stall cause, apply the remedy or resolve the
  judgment through the decision ladder — instead of waiting on a reply —
  then record the outcome and report. Reads the unified --auto doctrine at
  .claude/rules/moai/workflow/auto-semantics.md for the ladder, the gate
  inventory, and the record formats.
when_to_use: >
  Use when a loop iteration re-awakens a lane session (the awaken rule: the
  watchdog pass runs BEFORE card work resumes), when a lane shows no
  progress across the watch window, or when a lane needs a judgment to
  proceed and no reply has arrived.
license: Apache-2.0
compatibility: Designed for Claude Code and Codex sessions
allowed-tools: Read, Grep, Glob, Bash(git rev-parse:*), Bash(git status:*), Bash(git log:*), Bash(stat:*), Bash(ls:*), Bash(test:*), Bash(moai integration status:*)
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  tags: "watchdog, stall, lane, ladder, decision-ladder, unattended, auto"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
---

# Lane Stall Watchdog — One Iteration

<!-- moai:role-rules-required -->
[HARD] Before the first watchdog pass of a session running this skill, read BOTH role-gated rule files in full: `.claude/rules/moai/workflow/factory-dispatch.md` and `.claude/rules/moai/workflow/cross-session-messaging.md`. A Codex-only project deploys the same files under `.moai/policies/` (for example `.moai/policies/workflow/factory-dispatch.md`) — read whichever layout this project carries. The always-loaded surface carries only their stubs; the lane obligations and the messaging rules the ladder operates under live in those two files.

One watchdog pass for a lane session (a factory lane running
under `moai cc` / `moai glm` / `moai codex`). The law — the ladder, the
outcome transitions, the gate inventory, the view–SSOT rule, the record
formats — lives in `.claude/rules/moai/workflow/auto-semantics.md`. Read
that doctrine before the first pass and whenever a step below needs its
detail; this body is the executable procedure.

## 0. Never prompt

The watchdog never asks the operator a question and never answers an
approval gate on anyone's behalf. Everything that would have been a question
becomes either a resolved judgment (the ladder) or an explicit wait record.
A keep-set gate — environment-impossible, operator-held, or an irreversible
external-shared operation — is never satisfied here: the recheck reads the
gate surface, and finding no change, it re-records the wait and yields.

## 1. Measure progress

Three channels (doctrine §3–4):

- **HEAD**: `git rev-parse --short HEAD` in the lane's own worktree.
- **evidence mtimes**: the newest mtime across `.moai/reports/<card-id>/`
  and the active SPEC's `progress.md`.
- **integration-window state**: `moai integration status`.

Compare against the previous snapshot at
`.moai/state/watchdog/<card-id>.json` (fields: `observed_at`, `head_sha`,
`evidence_mtime_max`, `window_state`). The lane is stalled only when ALL
THREE channels are unchanged AND the previous observation is at least N
minutes old. N defaults to 15 minutes and is a per-invocation parameter —
never a config key. No previous snapshot → record a fresh one, no stall
verdict (fail-open first observation). One channel moving is progress —
evidence mtime advancing without a commit included: end the iteration with
a one-line status — unless the lane itself is parked on a delegate (§3.5):
then the movement is the delegate's, and the deliverable read in §3.5
decides, not this rule. Write the new snapshot after the verdict.

## 2. Classify the cause

| Cause | Signature |
|---|---|
| **awaited-judgment** | the lane stopped where a judgment was needed (a gate, a verdict, a choice) and no reply arrived |
| **blocked-by** | the card waits on a predecessor card; the predecessor's landing state is the open question |
| **shell-error** | the lane halted on an unadjudicated command failure (a non-zero exit with no recorded disposition) |
| **accidental-stop** | the session died or was interrupted mid-card — an API error (a 429) ends the turn with no further model action — and no deliberate stop was recorded |
| **awaited-delegate** | the lane parked on delegated work — a spawned agent's report, a background run's completion — and the report or notice has not arrived; an `available` idle notice that promised a later report counts, because that agent sends nothing more unless it is messaged |

No stall / progress detected → end the iteration.

## 3. Apply the remedy

### 3.1 awaited-judgment — the decision ladder

Resolve through the ladder IN ORDER (doctrine §6–7). Availability failures
fail-open downward; negative or inconclusive verdicts never auto-proceed,
and at an authority gate they fail closed.

1. **Disk evidence** — read the deciding artifact directly: audit verdict
   files, the card's progress record, plan-audit verdicts, evidence paths.
   Evidence shows proceed → resume. Evidence shows do-not-proceed → record
   the wait; never proceed against evidence. Unreadable or absent → step 2.
2. **Decision board** — run `moai decision read --scope card:<id>` (the
   card's rulings plus every standing ruling; doctrine §11: the append-only
   board under the moai home, keyed by the project). A recorded judgment →
   follow it. A judgment that says wait → explicit wait record (with an id;
   doctrine §14). A record whose `resolves` names an open wait ends that
   wait. `board=absent` / `board=empty` → step 3; never wait on the board
   being filled. While a wait on the leader stays open, keep the one-shot
   short recheck of doctrine §14 armed.
3. **Audit cross** — run one `moai` MCP audit tool (`codex_audit`,
   `glm_audit`, `claude_audit`, or `audit_multi`) for an independent second
   opinion. Positive → proceed per verdict. Negative → fail-closed: no
   proceed; record + escalate to step 5. Tool absent → step 4. Inconclusive
   → record it, → step 4.
4. **jev_ask** — only while `workflow.jev.enabled` is true; a bounded
   proceed / retry / wait judgment with repeatable, side-effect-free
   parameters; never for completion, merge, or queue outcomes. Gated off or
   unavailable → step 5 (fail-open by design).
5. **Lead chat** — last resort. Record the explicit wait (§5 below), then
   yield.

Write a decision record naming whichever step resolved the judgment (§5
below).

### 3.2 blocked-by

Read the blocking predecessor's on-disk evidence directly (its
`.moai/reports/<card-id>/` path, its progress record). Evidence shows the
predecessor landed → resume. Otherwise record an explicit wait: reason +
predecessor id + recheck point. Where the predecessor was dropped, the wait
record names the queue relation-removal verb as the operator escape — the
lane does NOT run it (queue mutation is prohibited). The queue-level pickup
filter is another surface's decision; this remedy only reads.

### 3.3 shell-error

Adjudicate the error: transient (network blip, lock contention, flaky
timing) vs determinate (missing remote, bad ref, usage error).

- transient → retry at most 3 times, recording the error and the retry
  count each time; retries exhausted → structured blocker report.
- determinate → ZERO retries; straight to a structured blocker report with
  the verbatim error text.

Never halt indefinitely on an unadjudicated error; never loop silently.

### 3.4 accidental-stop

Resume by checkpoint return: use the session-handoff resume message with the
worktree-anchored Block 0 (`git rev-parse --show-toplevel` → this worktree),
resuming from the last recorded checkpoint in the card's progress record.
Never restart the card from zero.

### 3.5 awaited-delegate

Read the delegate's deliverable on disk — the artifacts it was asked to write,
its output file, its commits — against the delegation record (what was asked,
where it was to land). Never read the absence of its message as the absence of
its work.

- Deliverable present and consistent with the delegation → resume and consume
  it; never wait for the report or the notice. An agent that went `available`
  sends nothing more unless it is messaged, so the report may never come.
- Deliverable absent or partial and the delegate evidently ended (its idle or
  completion notice arrived, or its process is gone) → re-delegate with the
  partial state, or resume the delegate with `SendMessage` only while it is a
  live teammate. Never address a teammate stopped with `TaskStop` by name — one
  message revives it as an ownerless writer
  (`cross-session-messaging.md` § Rules). Record which.
- Deliverable absent and the evidence still moving (the delegate is working) →
  explicit wait whose recheck point is the next cron fire.
- Deliverable absent, no sign the delegate ended, and the evidence unchanged
  across two consecutive fires → structured blocker naming the delegate, what
  was asked of it, and the evidence read; do not wait a third time.

## 4. Decision board protocol

- **SSOT**: the disk board (doctrine §11). Append-only, one record per
  line; never rewrite board history.
- **Views**: render state with the runner's native task tool — TaskCreate /
  TaskUpdate on the Claude runner, `update_plan` on codex — and read
  judgments ONLY from disk. A view is never the SSOT; a long-form plan
  document is never the board.
- **Codex step ⑤**: no session-messaging tool exists there — cross-harness
  notice rides the broker (`session_msg_register` / `session_msg_send` +
  poll) and the queue-on-disk delegation channel; the wait reason is
  recorded on disk either way.

## 5. Decision records

Every autonomous gate transition and every ladder-resolved judgment writes
ONE grep-able line on the disk board:

```text
decision record: decided_by=<runner+role> evidence_refs=<paths+verdict-ids> ladder_path=<step>
```

`decided_by` names the deciding runner and role; `evidence_refs` names the
artifact paths and verdict ids the decision rests on; `ladder_path` names
the ladder step (①–⑤) or the gate row that resolved it. The sync audit
re-reads these records — write them so a stranger can re-verify.

## 6. Composition boundary

The queue-level pickup decision — the pickup filter that excludes unrelated
or blocked cards before a pick — belongs to the queue surface's owner. This
watchdog is queue-readonly: it reads evidence and state surfaces, performs
zero queue mutation, and never picks, drops, edits, or unrelates anything.

## 7. Output shape — escalate or record

Every iteration ends in exactly one of:

- **resumed** — progress detected or a judgment resolved; say which step,
  and cite the decision record line.
- **explicit wait** — the wait record line and its recheck point, nothing
  more; end the iteration (no idle-spin).
- **structured blocker** — determinate failure or retries exhausted: the
  verbatim error, what was attempted, the evidence read, and the decision
  owed. Return it to the orchestrating session; do not retry past the
  ceiling; do not prompt anyone.

## Boundaries (hard)

1. Never prompt the operator; never answer an approval gate. Keep-set gates
   (environment-impossible, operator-held, irreversible external-shared
   operations) are rechecked and yielded to, never satisfied by this skill.
2. Queue read-only (§6).
3. One writer per tree: the watchdog writes only its own snapshot, its own
   evidence records, and the disk board.
4. Verification is lane-local; no background load.
5. Hook-independent: awaken-tick only; no hook surface. The scheduler that
   fires the awaken is the lane's standing recheck cron (doctrine §5.1), armed
   at card intake — this skill is its prompt, not its owner.
