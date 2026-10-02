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
**blocked-by** / **shell-error** / **accidental-stop** / **awaited-delegate**)
→ apply the remedy or the ladder (§6) → resume, or record an explicit wait
(§14).

The watchdog is awaken-tick based and **hook-independent by design**: it
binds the awaken turn's FIRST action and introduces no hook surface; who arms
the scheduler that fires the awaken is §5.1. The dual-harness hook-parity
precedent applies only if a future iteration adds a hook surface.

Canonical awaken prompt (the prompt of the lane's standing recheck cron, §5.1):

```text
Each iteration: run the lane stall watchdog FIRST — invoke
Skill("moai-lane-watchdog") and follow it — then resume the card work where
the card's progress record left off. Never resume card work without the
watchdog pass. If the watchdog records an explicit wait, end the iteration;
do not idle-spin on the wait.
```

### 5.1 The standing awaken carrier

A lane that has stopped cannot wake itself, and a turn ended by an API error
(a 429) leaves no model action in which to arm anything. The carrier is
therefore **standing**: armed once at card intake, before the first stage, and
kept until the completion report.

- **Arm.** `CronCreate` with `cron: "7,27,47 * * * *"` (off-minute, every 20
  minutes), `recurring: true`, and the canonical awaken prompt above. Never a
  one-shot at an absolute clock time: a local/UTC slip lands it in the past
  and it never fires (observed on card t1393), and it covers one wait rather
  than the stall that follows it. The 20-minute cadence keeps every fire
  longer than the N-minute stall window after the previous snapshot (§4).
- **Keep.** `CronList` at card intake and after every `/clear`; re-arm when
  the entry is missing. A recurring job expires after 7 days — a card that
  outlives that re-arms it. Delete it (`CronDelete`) when the completion report
  is sent; the next card arms its own.
- **Read disk, not messages.** Every wake — the cron, a leader message, a
  teammate's idle notice — starts with the watchdog pass, which reads the
  evidence on disk (the three channels of §3, the card's progress record, the
  reports, the commits, a delegate's deliverable). A message says when to
  look; it is never evidence of progress or of its absence.
- **Cost.** A fire after the cache window re-writes the prefix once. The
  cadence trades that cost against a stall bounded by about two periods (the
  first observation after a wake yields no verdict, §4) instead of the 87 to
  1606 minutes measured on cards t1393 and t1339.

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
impossible. The awaken carrier on the Claude runner is the standing recheck
cron (§5.1). The codex runner has no session cron tool; its substitute is the
leader's evidence read plus the explicit wait record on disk — a named gap,
never an implicit one.

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
| card pick | AUTONOMOUS inside a batch authorization (the `--auto` invocation IS the approval; §9.3); queue ADMISSION (production) stays the operator's — not a gate on work |
| plan-audit bypass flags | RETIRED into the default path — the audit cross IS the entry evidence (§9.1) |
| Jev capability gate (`workflow.jev.enabled`, default false) | capability switch, not a work gate — listed for completeness; enabling is the operator's |

### 9.1 The default-autonomous Kickoff transition

plan→run entry is AUTONOMOUS when ALL of the following hold: the independent
plan-audit verdict is PASS (FAIL and INCONCLUSIVE stay hard blocks — the §7
authority-gate invariant), the SPEC's plan phase records audit-ready status,
the plan-artifact hashes are unchanged since that verdict, and no blocker is
open. The transition writes a decision record (§10) that the sync audit
re-reads. Keep-set cases keep the operator answer; the contract-signing path
stays as the voluntary equivalent form. Operator-form Kickoff rows that wait
together are presented through §9.2.

### 9.2 The batch gate summary

The batch gate summary is a presentation form for operator-form decisions, not an approval method: it lowers no evidence standard of §9.1. It is distinct from the `--auto` batch authorization of the card pick row, which authorizes the invoked session to take cards on its own judgment, each only through a lease (§9.3).

**Membership**

- The summary applies to the plan→run Kickoff row only. Every other row of the §9 inventory (the factory decide rows, the sync blocking approval, card pick) forms no summary row and is asked individually where an operator answer is required.
- The session that holds the operator dialogue builds it. A lane presents only its own card's gate and never forms a cross-card batch.
- A ready card is never held back to wait for further rows; a row that becomes ready later joins the next summary or is asked individually.
- A single ready Kickoff row is not a summary and is asked individually; a summary needs two or more rows.
- When one judgment is the common decision subject of two or more pending Kickoff rows that are neither reserved nor blocked, ask it once and name every affected card in the question and in the report before it. That question never names a blocked or reserved row.
- When `workflow.autonomy.mode` is `contract`, the contract signature checked by `moai contract kickoff-check` is the plan→run gate and no summary row exists.

**Report and question**

- The report precedes the question in the same response. The report head names the plan→run Kickoff gate row, and each listed card has one row carrying the card id, the SPEC id, the independent plan-audit verdict with its iteration identifier, score, and margin to the tier's PASS threshold, the plan-artifact-hash-unchanged check, a reference to the card's own decision record, and `counter_refs=`. Each row also states the result of the keep-set and leader-held-power check together with the basis on which the row was classified. Rows carrying counter-evidence are listed before rows without it.
- A verdict reference binds to the final iteration of the card's current plan artifacts; an earlier iteration's PASS never covers them.
- Exactly one decision question follows when at least one listed row is approvable, in a call that holds no second batch gate summary question. Its approval covers only the listed approvable rows. The report states that rows added later, reserved rows, and blocked rows are not covered by it. When no listed row is approvable, no question is asked: the report states the blocked and reserved rows and ends there.
- To pull a row out, the operator writes its card id in the question channel's automatic free-text entry, exactly as the report shows it, whatever the number of rows. The question text states that every card id written there is pulled out and every other listed approvable row is approved. An answer that cannot be read as card ids, or that names a card id the report does not list, approves no row and the question is asked again. Pulling a row out leaves the approval of the remaining rows intact, and the pulled-out row is handled individually. Explicit options stay within the channel's per-question limit.
- Recommendation labels follow `interview.recommendation_mode` unchanged. Where the harness has no question channel, the same summary is delivered as a blocker report.

**Counter-evidence**

- Each row carries a `counter_refs=` field naming the strongest evidence against proceeding. Its sources are a closed list: audit warnings or recorded debt, the margin to the PASS threshold, unresolved decision-index rows, divergent audit-cross opinions, open blockers or wait records, and path overlap with another row of the same summary.
- When none is found, the row reads `counter_refs=none searched=<token>`, where the token is one whitespace-free word naming the search, in the report and in the decision record alike. A `counter_refs=none` without `searched=` is not a counter-evidence statement.

**Approvable, blocked, reserved**

- A row is approvable only when all four hold: its most recent independent plan-audit verdict is PASS, the plan phase records audit-ready status, the plan-artifact hashes are unchanged since that verdict, and no blocker is open. Every other row is reported as blocked and excluded from the single approval.
- The verdict must be independent, produced by the plan-auditor; a PASS stated by the session that authored the plan artifacts is blocked.
- Blocked states are: PASS-WITH-DEBT, BYPASSED, FAIL, INCONCLUSIVE, an absent verdict, audit-ready status not recorded, a plan-artifact hash changed since the verdict, and an open blocker.
- A row is reserved, and handled individually outside the single approval, when it falls in a keep-set category (environment-impossible, operator-held, or an irreversible operation on an external shared system) or in a power the leader session keeps: final PASS/FAIL verdicts, final merge approval, operator gates, card issuance and `done` through queue mutations, CodeRabbit slot-wait adjudication, and cross-session dispute coordination. The operator gates item does not include the operator-form plan→run Kickoff row: that row is reserved only when it falls in a keep-set category, and otherwise it is a summary row classified by the approvability rule above.

**Records**

- When the answer arrives and before each approved row is recorded, the session re-reads all four conditions of approvability and the row's reserved classification, an operator hold included. A row that no longer qualifies is refused, not recorded as approved, and the operator is told.
- Each approved row gets its own decision record in the §10 form, as one line carrying the three fields in order, followed by `counter_refs=`. Its `ladder_path` holds the gate row slug followed by `;batch=<id>`, where `<id>` is the UTC time of the decision question as `YYYYMMDDTHHMMSSZ`, identical on every record of one summary and advanced to the next free second when another decision record on the board already carries that value. A row without its own record is not approved.
- The single approval weakens no other Kickoff condition: for each approved card the tier, the mode preference, the PR strategy, and the chain scope are on disk, from the card or SPEC contract or from an operator dialogue held for that card, before run entry.

### 9.3 The card-pick authorization

The card pick row's `--auto` invocation authorizes the invoked session to take
cards from the queue on its own judgment, each only through a lease
(`moai factory next`, or `moai factory next --card <id>` for a judged pick;
the lease is a lane's only pick path) and never a keep-set card. Queue
ADMISSION stays the operator's, and Jev stays display-only.

The keep-set of a pick: a card in the `hold` state or whose text begins with
the `[보류` marker, a card classified blocked, a card the factory record already
owns, a serial card while another serial card is in flight, and a card whose
text hinges on an operator confirmation — payments, secrets, or irreversible
external-shared work. The lease refuses the first four mechanically; the last
is the session's judgment, recorded.

The inputs of the judgment: the card's class, its relation records, its
pull-request and landed state, its worktree presence, its file overlap with
in-flight lanes, and the candidates it skipped. The set is open: a later card
adds an input without amending the keep-set or the lease path. File overlap is a
pluggable input whose fallback — the paths changed by each in-flight lane's
card branch — is inferred; the record does not require it.

A lane that takes a card by lease under this authorization writes one §10 line
in the card's progress record, as evidence, not the decision board (§11):

```text
decision record: decided_by=<runner+role> evidence_refs=card=<id>;class=<class>;relate=<ids>;pr=<state>;wt=<path>;overlap=<paths>;skipped=<id:reason,...> ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9)
```

An input that could not be read is written `unmeasured`, never omitted and
never `none`. The line is self-attested, and no party re-reads the card-pick
record: the §9.1 and §10 sentences that name the sync audit's re-read as the
compensating control do not hold for it. That is a residual risk, not a control.

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

A decision approved through the batch gate summary (§9.2) also carries its
per-row counter-reference field and the batch id; that format is defined only
in §9.2.

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
