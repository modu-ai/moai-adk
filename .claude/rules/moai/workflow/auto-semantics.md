---
paths: ".claude/skills/moai-lane-watchdog/**"
---

# --auto Semantics — The Lane-Centered Unification

> **Loading scope**: not always-loaded. The lane stall watchdog skill
> (`.claude/skills/moai-lane-watchdog/`) reads this document explicitly at
> each awaken; the `paths:` frontmatter keeps it out of the session prefix
> otherwise. Load it when adjudicating a lane wait, amending a gate
> disposition, or editing the watchdog skill.
>
> **Classification**: canonical reference — the single source of truth for
> what `--auto` means on the lane surface, how the other `--auto` surfaces
> are bounded, and how a lane resolves a judgment without waiting on a reply.

## 1. The common definition

`--auto` means: **minimize human intervention through autonomous adjudication.**

- **Autonomous** — the run resolves its own judgments instead of waiting for
  a reply. The absence of a reply is never read as the absence of progress.
- **Adjudication** — every resolution names its evidence, follows the
  decision ladder (§6), and writes a decision record (§10). An unresolved
  audit is not a pass; availability failures are recorded, never silently
  promoted to positive outcomes.

What `--auto` does NOT mean: it does not touch the three KEEP categories
(§9), does not admit work (queue production stays the operator's), and does
not weaken any gate — it replaces one form of a gate's evidence (a human
answer) with a stronger one (an independent audit cross + the evidence
criteria + a written decision record), where this inventory says so.

## 2. Surface inventory

| Surface | Coverage | Boundary statement |
|---|---|---|
| lane (factory / kanban worker session, `moai cc` / `moai glm` / `moai codex`) | **COVERED** — this document + the watchdog skill | the full semantics below |
| queue batch cycle (`todo --auto`) | boundary only | an operator-unused batch cycle with a 30-minute evidence deadline + automatic unpick of an unprogressing picked card. Its bounded-wait shape is cited as design PRECEDENT for the watchdog window — nothing more. The queue surface keeps its own owner and its own mechanics; this document does not redefine them |
| mission loop (`goal --auto`) | boundary only | already bounded — turn ceiling, stagnation guard, wall-clock bound; untouched |

## 3. Progress measurement — three channels

A lane's progress is measured on exactly three channels:

- **(a) HEAD** — the lane worktree's HEAD commit SHA (`git rev-parse --short
  HEAD`).
- **(b) evidence mtimes** — the newest mtime across the card's evidence set:
  `.moai/reports/<card-id>/` and the active SPEC's `progress.md`.
- **(c) integration-window state** — `moai integration status` (the holder
  id, or a hash of its output).

Stall rule: the lane is STALL only when **all three** channels show no
change across the N-minute window. N defaults to 15 minutes and is a
per-invocation parameter of the watchdog procedure — never a config key (a
new config file is wiped by template redeployment). One channel moving is
progress: docs-only progress (an evidence mtime advancing without a commit)
is progress, not a stall.

## 4. The observation snapshot

The N-minute verdict compares TWO observations, so each watchdog awaken
persists one:

- **path**: `.moai/state/watchdog/<card-id>.json` inside the lane's own
  worktree — gitignored runtime state, tree-local; the lane measures itself,
  so no home-surface visibility is needed
- **fields**: `observed_at` (RFC3339), `head_sha`, `evidence_mtime_max`, and
  `window_state` (the integration-status output hash or holder id)
- **verdict**: stall iff (no channel change vs the previous snapshot) AND
  the previous observation is at least N minutes old
- **single-writer** — the lane itself; no cross-writer discipline applies
- **fail-open first observation**: a missing or corrupt snapshot means "no
  previous observation" — record a fresh one, no stall verdict

## 5. The awaken rule

When a loop iteration re-awakens a lane, the awaken turn runs the watchdog
self-diagnosis **before** resuming card work:

measure progress (§3–4) → classify the cause (**awaited-judgment** /
**blocked-by** / **shell-error** / **accidental-stop**) → apply the remedy
or the ladder (§6) → resume, or record an explicit wait (§14).

The watchdog is awaken-tick based and **hook-independent by design**: it
binds the awaken turn's FIRST action, never the scheduler, and introduces no
hook surface. The dual-harness hook-parity precedent applies only if a
future iteration adds a hook surface.

Canonical awaken prompt (paste into the lane's loop driver):

```text
Each iteration: run the lane stall watchdog FIRST — invoke
Skill("moai-lane-watchdog") and follow it — then resume the card work where
the card's progress record left off. Never resume card work without the
watchdog pass. If the watchdog records an explicit wait, end the iteration;
do not idle-spin on the wait.
```

## 6. The decision ladder

When a lane needs a judgment to proceed, it resolves through the ladder IN
ORDER — the lead chat is the LAST step, not the first:

| Step | Carrier | Autonomous? |
|---|---|---|
| ① disk evidence | read the deciding artifact directly: audit verdict files, the card's progress record, plan-audit verdicts, evidence paths | yes |
| ② decision board | poll the shared decision store (§11) — the lead/orchestrator records judgments there instead of replying | yes |
| ③ audit cross | the `moai` MCP audit tools (`codex_audit` / `glm_audit` / `claude_audit` / `audit_multi`) for an independent second opinion; fail-open on tool absence | yes |
| ④ jev_ask | the MCP judgment tool for a bounded proceed / retry / wait call — ONLY while `workflow.jev.enabled` is true (ships false); repeatable, side-effect-free parameters; never for completion, merge, or queue decisions | yes, gated |
| ⑤ lead chat | the reply path — only when ①–④ are all unavailable or unresolved; the wait is explicit (§14) | last resort |

## 7. Outcome → action transitions

Availability failures fail-open DOWNWARD; negative or inconclusive verdicts
NEVER auto-proceed:

| Step | positive outcome | negative outcome | inconclusive / unavailable |
|---|---|---|---|
| ① | evidence shows proceed → resume | evidence shows do-not-proceed → record wait / escalate; NEVER proceed against evidence | unreadable / absent → ② |
| ② | board judgment recorded → follow it | board judgment says wait → explicit wait record | board empty / not yet filled → ③ (never wait on the board being filled) |
| ③ | verdict positive → proceed per verdict | verdict negative → **fail-closed: no proceed**; record + escalate to ⑤ | tool absent → ④; verdict inconclusive → record it, → ④ |
| ④ | proceed / retry / wait per judgment | retry / wait per judgment | gated off or unavailable → ⑤ (fail-open by design) |
| ⑤ | — | — | explicit wait record (§14); recheck per awaken |

**Authority-gate invariant**: where the ladder adjudicates an AUTHORITY gate
(the plan→run Kickoff, the sync blocking approval), a NEGATIVE or
INCONCLUSIVE audit verdict is FAIL-CLOSED — the gate does not open on an
unresolved audit; only a positive verdict proceeds, and every transition
writes a decision record (§10). Availability failures (tool missing,
timeout) degrade to the next step and are recorded — they never read as a
positive.

## 8. Harness neutrality

Every ladder step carries a per-runner path. The Claude runner is `moai cc`
/ `moai glm` (same runner, different backend); the codex runner is
`moai codex`.

| Ladder step | Claude runner | codex runner |
|---|---|---|
| ① disk evidence + progress detection (SHA / mtime / window state) | shell + git + stat | same (shell-based; harness-indifferent) |
| ② decision board + native view | disk SSOT (§11); view: TaskCreate / TaskUpdate | same disk SSOT; view: `update_plan` (the codex native plan tool; shape `update_plan(explanation?, plan: [{step, status}])` — a reported premise; re-verify against current Codex documentation before relying on it) |
| ③ audit cross | `moai` MCP audit tools | the same MCP server wires into codex sessions; `codex_audit` is codex-native |
| ④ jev_ask | MCP tool — harness-indifferent | same |
| ⑤ reply path | the session messaging tool | NO session-messaging tool on codex — cross-harness notice only via the session messaging broker (`session_msg_register` / `session_msg_send` + poll) and the queue-on-disk delegation channel; the wait reason is recorded on disk |

A step impossible on a runner names its substitute — never implicitly
impossible. The awaken CARRIER is scheduler-mediated and out of scope (§5
binds the turn's first action, not the scheduler).

## 9. The gate inventory and dispositions

The human gates, each with a disposition. KEEP is admissible ONLY in three
categories: **environment-impossible** (execution itself cannot proceed —
missing credentials/runner), **operator-held** work, and **irreversible
operations touching external shared systems**. Any KEEP outside these three
carries its justification right here, in this inventory.

| Gate | Disposition |
|---|---|
| factory decide: kickoff approve/reject | AUTONOMOUS — the independent audit cross is the entry evidence (§9.1) |
| factory decide: push | KEEP — irreversible operation on an external shared system (origin push); keep-set category 3 |
| factory decide: resume / block / unblock | AUTONOMOUS with a decision record; an operator `hold` state forces KEEP (keep-set category 2) |
| factory decide: abandon | AUTONOMOUS only when the branch is already integrated (nothing lost); otherwise KEEP — irreversible disposal of the only copy (keep-set category 3) |
| plan→run Kickoff | AUTONOMOUS (§9.1); the operator question channel survives for keep-set cases and operator dialogue |
| contract signing (`workflow.autonomy.mode: contract`, human signature) | PRESERVED EQUIVALENT FORM — the voluntary human-signature path stays available and unchanged; justification for preserving outside the three keep categories: the path imposes nothing on anyone (voluntary), so the minimization principle does not require removing it |
| sync blocking approval | AUTONOMOUS — the sync-auditor verdict + evidence thresholds adjudicate; external-shared operations inside sync (release push/merge) KEEP (keep-set category 3) |
| card pick | AUTONOMOUS inside a batch authorization (the `--auto` invocation IS the approval); queue ADMISSION (production) stays the operator's — not a gate on work |
| plan-audit bypass flags | RETIRED into the default path — the audit cross IS the entry evidence (§9.1) |
| Jev capability gate (`workflow.jev.enabled`, default false) | capability switch, not a work gate — listed for completeness; enabling is the operator's |

### 9.1 The default-autonomous Kickoff transition

plan→run entry is AUTONOMOUS when ALL of the following hold: the independent
plan-audit verdict is PASS (FAIL and INCONCLUSIVE stay hard blocks — the §7
authority-gate invariant), the SPEC's plan phase records audit-ready status,
the plan-artifact hashes are unchanged since that verdict, and no blocker is
open. The transition writes a decision record (§10) that the sync audit
re-reads. Keep-set cases keep the operator answer; the contract-signing path
stays as the voluntary equivalent form.

## 10. Decision records

Every autonomous gate transition and every ladder-resolved judgment writes a
decision record on the disk SSOT (§11) as ONE grep-able line naming the
three fields in order:

```text
decision record: decided_by=<runner+role> evidence_refs=<paths+verdict-ids> ladder_path=<step>
```

- `decided_by` — the deciding runner and role
- `evidence_refs` — the artifact paths and verdict ids the decision rests on
- `ladder_path` — the ladder step (①–⑤) or the gate row that resolved it

A decision record is self-attested by the writing lane; the sync audit's
re-read of these records is the compensating control (detection, not
prevention). Write them so a stranger can re-verify.

## 11. The decision board and the view–SSOT rule

The decision board's SSOT is a harness-neutral HOME-surface disk store — the
moai home's state directory for the project (`${MOAI_HOME:-$HOME/.moai}`,
keyed by the primary checkout's project key), append-only, one record per
line. A tree-local copy is a ghost — never a board.

Each runner's native task tool is a rendering VIEW only: TaskCreate /
TaskUpdate on the Claude runner, `update_plan` on codex. The rule:
**render state with the native view, read judgments from disk.** A long-form
plan document is reference material, never the board.

## 12. The Jev boundary

The single Jev surface the watchdog consumes is the `jev_ask` MCP tool, and
only while `workflow.jev.enabled` is true (ships false — enabling is the
operator's act). jev_ask is a **display-only** judgment query: it observes
state and returns a bounded proceed / retry / wait verdict. The lane passes
only repeatable, side-effect-free parameters, and never uses it to decide
completion, merge, or queue outcomes.

## 13. The queue read-only boundary

The watchdog and the remedies read evidence and state surfaces only. Queue
mutation verbs (add, drop, done, edit, relate/unrelate) and contract signing
stay prohibited for lanes. A blocked-by remedy resolves by evidence, never
by relation bookkeeping.

## 14. Explicit waits

A wait is legitimate only as an **explicit wait** — a disk record naming:

```text
wait record: waiting_on=<subject> reason=<why> recheck=<condition or next check point>
```

recorded on the card's evidence path or the decision board. The lane does
not idle-spin on a wait: it rechecks per awaken (§5), yields when the
recheck shows no change, and never prompts the operator on its own behalf.
Where no on-disk gate state exists (a gate mid-question), the recheck
degrades to ladder steps ①–④ — the wait note is then the only record, and
progress resumes only on evidence.

## 15. Cross-references

- `.claude/skills/moai-lane-watchdog/` — the executable carrier (one watchdog iteration)
- `.claude/rules/moai/workflow/kanban-dispatch.md` — the dispatch protocol; the explicit-wait posture at dispatch
- `.claude/rules/moai/workflow/cross-session-messaging.md` — the reply-independence boundary step ⑤ rests on
- `.claude/rules/moai/core/agent-common-protocol.md` — the retry ceiling the shell-error remedy inherits
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` — the checkpoint-return convention the accidental-stop remedy follows
- `.claude/rules/moai/workflow/contract-autonomy.md` — the contract-signing equivalent form

---

Version: 1.0.0
Classification: Canonical reference — the lane-surface `--auto` SSOT.
