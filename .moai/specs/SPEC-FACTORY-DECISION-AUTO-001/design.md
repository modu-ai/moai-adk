# design.md — SPEC-FACTORY-DECISION-AUTO-001

Mechanism-level design. Names below are proposals for the run phase; spec.md binds behavior only.

## 1. Decision board

### 1.1 Location

`<moai home>/projects/<project-key>/decisions/board.jsonl`, resolved through the existing
`homestate` project-directory helper (`homestate.ProjectDir`, beside `FactoryDir`/`TodoDir`), keyed
by the primary checkout so every linked worktree reads the same file (doctrine §11). Appends use
`O_APPEND` with a single `write` per record and an advisory file lock, so concurrent writers never
interleave a line.

### 1.2 Record shape

```json
{"id":"d-20261003T091500Z-3f2a","scope":"card:t1409","kind":"ruling",
 "decided_by":"claude+leader","evidence_refs":".moai/reports/t1409/plan-audit-iter3.md",
 "ladder_path":"②","body":"one delta round; then hold+split","created":"2026-10-03T09:15:00Z",
 "supersedes":"","resolves":"w-t1409-20261003T090000Z"}
```

Kinds (closed enumeration): `ruling`, `standing-rule`, `wait-resolution`, `hold`,
`split-proposal-ack`, `ceiling-exception`, `supersede`. A `standing` scope requires a predicate in
`body` stating the situation it governs (e.g. "plan verdict admitted by the plan-phase predicate,
hash unchanged → Kickoff autonomous"). `supersedes` and `resolves` must name an existing record id or
wait id; `record` refuses an unknown reference. `read` also renders each record in the §10 one-line
form so existing greps keep working.

### 1.3 CLI

- `moai decision record --scope card:<id>|standing --kind <k> --body <text|-> --evidence <refs>
  [--ladder <step>] [--supersedes <id>] [--resolves <wait-id>]` — refuses under lane refusal (the
  same predicate as `factory decide`, REQ-SD-015) before any I/O.
- `moai decision read [--scope card:<id>] [--json] [--all]` — card scope returns card + standing
  records; `--all` includes superseded ones; status line `board=absent|empty|ok unparseable=<n>`.
- An MCP mirror (`decision_read`) is optional for the run phase; the CLI is the contract.

### 1.4 Lanes and waits (decision-index Q1, Q15)

Lanes do not write the board. A lane's waits, DEFAULT-APPLIED rows, ceiling records, and split
proposals go to the card's progress record and evidence path. Wait lines gain an id:

```text
wait record: id=w-<card>-<UTC> waiting_on=leader reason=<why> recheck=<condition>
```

A wait is resolved only by a board record whose `resolves` equals that id. Any other record for the
same card leaves the wait open.

## 2. Authority register extension

A `decision-index.md` row may cite `board:<record-id>#<sha256-prefix>` or
`mission:<mission-id>#<contract-hash-prefix>`. The citation is valid only when the row also carries
the cited record line verbatim in a fenced block (the pin) and the digest of that pinned line matches.
Because the row is committed, `git show` resolves it, so the committed-only principle holds: the board
is the source and the pin is the citation. Only `standing` records and signed mission contracts are
admissible; a `card:` ruling is evidence for that card, not authority for others.

## 3. Admission predicate (decision-index Q8, Q9)

One function, `contract.AdmitVerdict(v VerdictFile, phase Phase, tier Tier) (ok bool, reason string)`,
is used by `internal/contract/rules.go`, `internal/contract/kickoff/decide.go`, and
`internal/homestate/card_transition.go` (T7 with `PhasePlan`, T13 with `PhaseSync`).

| Phase | Label | Admitted when |
|---|---|---|
| plan | `PASS` | score ≥ tier plan threshold (S 0.75 / M 0.80 / L 0.85), `must_pass_failed` = 0, `blocking_findings` = 0, plan-artifact hash matches |
| plan | `PASS-WITH-DEBT` | the PASS conditions AND `debts` with ≥ 1 item, each carrying `id`, `description`, `dispose_in` (`run` or `sync`) |
| sync | `PASS`, `PASS-WITH-DEBT` | the existing sync thresholds and semantics (unchanged); binding run conditions all disposed (§3.1) |
| both | `FAIL`, `INCONCLUSIVE`, `BYPASSED`, absent | never |

Today T7 admits on the label alone, so plain PASS at T7 gains the score, must-pass, and blocking
checks. Run M2 characterizes the existing T7/T13 tests before the change. The plan-auditor's verdict
block gains `must_pass_failed`, `blocking_findings`, `fix_scope`, `debts`, and (on final hits)
`defect_class` and `reread_hunks`. Doctrine: §9.1 "verdict is PASS" becomes "verdict admitted by the
plan-phase predicate". §9.2's blocked list drops bare PASS-WITH-DEBT and adds "PASS-WITH-DEBT not
admitted by the predicate".

### 3.1 Binding run conditions (decision-index Q14)

Run entry copies each debt into `progress.md` §E.2 under `### Binding run conditions` as
`- <debt-id>: <description> — dispose_in=<phase> — disposition: <pending|disposed: evidence>`. Both
sync verdict owners read that heading: the `sync-audit-4dim` workflow (happy path) and the
sync-auditor agent (fallback). An item still `pending` is a failed must-pass criterion, and the sync
verdict is `FAIL`.

## 4. Ceiling policy (decision-index Q7, Q12, Q13)

```yaml
harness:
  plan_audit_tier_ceilings: {S: 1, M: 2, L: 3}
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_final_hit: hold-and-split   # only value in v1
```

- **Who.** Every session that runs a plan audit (lane, kanban companion, or plain orchestrator).
- **Delta eligibility (mechanical).** The ceiling-hit verdict lists `fix_scope` entries
  (`<file>#<anchor>`, where an anchor is a markdown heading slug or a REQ/AC id). After the fix, the
  session computes `git diff --name-only` and the per-file hunk ranges between the ceiling-hit
  `audited_sha` and the delta round's `audited_sha`. It is eligible iff every changed hunk falls inside
  a `fix_scope` anchor's section, or the file is `progress.md`, `decision-index.md`, or under
  `.moai/reports/`, and the REQ id set and AC id set extracted at both SHAs are identical. A missing
  `fix_scope`, an unreachable SHA, or any other failed check makes it ineligible.
- **Final hit.** Iteration count = tier ceiling + `auto_delta_rounds` without admission, an
  ineligible delta, or a score-regression STOP → hold wait record + split proposal. A lane writes them
  to the card evidence path; a non-lane orchestrator writes them to the SPEC's `progress.md`, tells the
  user, and may offer an override through the question channel (the only question asked).
- **Release-blocking AC-wording exception.** A final hit is excepted only when all three hold:
  1. The card is release-blocking: a queue relation path leads from it to a card inside a release
     scope that a mission contract or a standing board record marks operator-approved.
  2. The final verdict has `blocking_findings: 1` and `defect_class: ac-wording`.
  3. The verdict lists `reread_hunks`.

  The leader then writes a `card:<id>` board record of kind `ceiling-exception`. The session changes
  only the listed hunks, and the auditor returns a re-read confirmation verdict scoped to those hunks
  instead of a full re-audit. Without the board record the hold stands.

The auditor's "Max 3" text and spec-workflow `:158` are rewritten to cite the tier map and the policy.

## 5. Factory audit decider (decision-index Q6, Q10)

- `homestate.DeciderAudit = "audit"`.
- New edge `T8a {CardKickoff → CardRun, guardKickoffAudit}`. The guard: loads the card's plan-audit
  verdict through the existing evidence reader (`audited_sha` == evidence SHA); applies
  `AdmitVerdict(…, PhasePlan, tier)`; recomputes the plan-artifact hash; requires audit-ready status
  in `progress.md` §E.1; requires no open blocker and no operator hold on the card row; and parses
  `decision-index.md` for a `Class: product-level` row (or a row with no Class) whose operator verdict
  is empty. Any failure refuses with a reason and leaves T8 (human) available. Lease and owner fields
  carry over unchanged. T8 is untouched.
- `factory decide`: `--decider audit` is admitted only with `--gate kickoff --choice approve`. Under
  lane refusal it is admitted only when the caller's lease holds the named card. All other lane calls
  still return `factoryDecideLaneRefusal()`. The push gate keeps requiring `human`.

## 6. FOUNDER defaults (decision-index Q2, Q3)

Row shape extension (gate on):

```text
### Q<N>: <question>
Label: FOUNDER
Class: implementation-level | product-level
Default: <option> (rule: <which published rule selected it>)
Alternate: <option>
Authority anchor: —
Why unresolved: <...>
Operator verdict:
```

Published Default rule (ordered): the option that preserves current behavior; else the option whose
undo is a single revert of this SPEC's own commits; else the option with the smaller user-visible
surface. A row the rule cannot rank carries no Default and blocks. `product-level` (closed list): a
change to a shipped command's default user-visible behavior, removal of a user-facing feature, or a
change to a template default. Kickoff writes `Operator verdict: DEFAULT-APPLIED <UTC> <runner+role>`.

The manager-spec clause "never carries an embedded recommendation or preferred answer" is amended:
it governs judgment calls only, and a Default selected by the published rule is a policy
application. Template source first, then the local copy, then `make agents-emit`.

## 7. Wait recheck (decision-index Q4, Q15)

After writing a wait line with `waiting_on=leader`, the watchdog arms a one-shot `CronCreate`
(`recurring: false`) at the configured delay from now, with the canonical awaken prompt. Each fire
runs the watchdog pass. If no board record `resolves` the wait id, it re-arms one more one-shot;
otherwise it arms nothing. Delay key: `workflow.watchdog.wait_recheck_minutes` (default 5; values
below 5 clamped to 5; M0(b) may raise the default). The fire time is computed from the cron tool's own
clock to avoid the local/UTC slip §5.1 records. Codex runner: no cron, a named gap.

## 8. Messaging hygiene (decision-index Q5, Q11)

- **Bind cache.** `<moai home>/projects/<key>/factory/bind-cache/<session>.json` holds
  `{session, run, pid, process_start, bound_at}`. On every prompt the hook still runs
  `factoryHookProbeRun`. On a full four-field match AND a probe result of the same live run, it
  returns before `factorymsg.Open` and the peer query. A probe reporting retired or a different run
  deletes the cache file and continues into the full bind (the stale-run rebind of
  SPEC-FACTORY-STALE-RUN-HEAL-001) in the same invocation. A probe error on a hit is reported
  degraded. The hook is the cache's only writer: it writes after a successful bind and deletes on
  invalidation.
- **Degraded inbox state.** Both call sites pass the state to `slog.Warn`
  (`.moai/logs/hook-runtime.log`). A notice "factory messaging degraded: <state>" is added to the hook
  context at most once per `workflow.factory.degraded_notice_minutes` (default 10) per session,
  tracked in the same cache file.
- **MCP staleness.** Intake reads the MCP server build from the server instructions line
  (`build vX (commit: SHA ...)`) or `moai doctor` and compares it with `moai version`. On mismatch it
  writes `mcp: server=<build> cli=<build> fallback=CLI` to progress and uses
  `moai verify codex-review --project-root`, `moai spec audit`, and similar commands.
