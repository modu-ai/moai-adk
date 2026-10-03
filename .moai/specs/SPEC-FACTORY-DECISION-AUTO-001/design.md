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
 "supersedes":""}
```

Kinds (closed enumeration): `ruling`, `standing-rule`, `wait-resolution`, `hold`,
`split-proposal-ack`, `supersede`. A `standing` scope requires a predicate in `body` stated as
the situation it governs (e.g. "PASS-family verdict, blocking 0, hash unchanged → Kickoff
autonomous"). `read` renders each record also as the §10 one-line form so existing greps keep
working.

### 1.3 CLI

- `moai decision record --scope card:<id>|standing --kind <k> --body <text|-> --evidence <refs>
  [--ladder <step>] [--supersedes <id>]` — refuses under lane refusal (same predicate as
  `factory decide`, REQ-SD-015) before any I/O.
- `moai decision read [--scope card:<id>] [--json] [--all]` — card scope returns card + standing
  records; `--all` includes superseded ones; status line reports `board=absent|empty|ok
  unparseable=<n>`.
- An MCP mirror (`decision_read`) is optional for the run phase; the CLI is the contract.

### 1.4 Lanes

Lanes do not write the board. A lane's waits, DEFAULT-APPLIED rows, ceiling-extension records, and
split proposals go to the card's progress record and evidence path (doctrine §14 already allows the
card evidence path). The wait-resolution link is: a board record with `scope=card:<id>` created
after the wait line resolves it.

## 2. Authority register extension

A `decision-index.md` row may cite `board:<record-id>#<sha256-prefix>` or
`mission:<mission-id>#<contract-hash-prefix>`. The citation is valid only when the row also
carries the cited record line verbatim in a fenced block (the pin), and the digest of that pinned
line matches. Because the row is committed, `git show` resolves it — the committed-only principle
holds; the board is the source, the pin is the citation. Only `standing` records and signed mission
contracts are admissible; a `card:` ruling is evidence for that card, not authority for others.

## 3. PASS-WITH-DEBT predicate

One function (proposed `contract.AdmitsKickoffVerdict(v VerdictFile, tier) (bool, reason)`) used by
`internal/contract/rules.go`, `internal/contract/kickoff/decide.go`, and
`internal/homestate/card_transition.go`:

| Verdict | Admitted when |
|---|---|
| `PASS` | score ≥ tier threshold, must-pass all pass, blocking = 0 |
| `PASS-WITH-DEBT` | the PASS conditions AND a `debts:` section with ≥1 item, each item carrying `id`, `description`, `dispose_in` (`run` or `sync`) |
| `FAIL`, `INCONCLUSIVE`, `BYPASSED`, absent | never |

The auditor's verdict block gains `blocking_findings: <n>` and `debts:` fields. Doctrine: §9.1
"verdict is PASS" becomes "verdict is PASS-family per the admission predicate"; §9.2's blocked list
drops PASS-WITH-DEBT and adds "PASS-WITH-DEBT without enumerated debts". Run entry copies each debt
into `progress.md` under a `### Binding run conditions` heading inside §E.2; the sync-auditor reads
that heading and scores each item disposed / undisposed.

## 4. Ceiling policy

```yaml
harness:
  plan_audit_tier_ceilings: {S: 1, M: 2, L: 3}
  plan_audit_ceiling_policy:
    auto_delta_rounds: 1
    on_second_hit: hold-and-split   # only value in v1
```

Delta eligibility is declared by the auditor (it owns the verdict) as a verdict-block field
`delta_eligible: true|false` with three sub-checks it states: `confined_to_required_fix`,
`req_ac_scope_unchanged` (mechanical — REQ/AC counts and §F headings compared to the previous
iteration), `no_regression` (no previously fixed defect id reappears). The lane reads the field;
it does not judge eligibility itself. Score-regression STOP always takes the second-hit path.
The auditor's "Max 3" text and spec-workflow `:158` are rewritten to cite the tier map and the
policy; the user-channel routing is replaced by the policy, with keep-set and product-level
escalation unchanged.

## 5. Factory audit decider

- `homestate.DeciderAudit = "audit"`.
- New edge `T8a {CardKickoff → CardRun, guardKickoffAudit}`: the guard loads the verdict file
  referenced by the card, applies the §3 predicate, recomputes the plan-artifact hash, and compares.
  Lease and owner fields are carried over unchanged. T8 (human → assigned) is untouched.
- `factory decide`: `--decider audit` is admitted only with `--gate kickoff --choice approve`; under
  lane refusal it is admitted only when the caller's lease holds the named card. All other lane
  calls still return `factoryDecideLaneRefusal()`.

## 6. FOUNDER defaults

Row shape extension (gate on):

```text
### Q<N>: <question>
Label: FOUNDER
Class: implementation-level | product-level
Default: <option> (reversibility rule: <which rule selected it>)
Alternate: <option>
Authority anchor: —
Why unresolved: <...>
Operator verdict:
```

Reversibility rule (ordered): the option that keeps the current behavior; else the option whose
undo is a single revert of this SPEC's own commits; else the option with the smaller user-visible
surface. A row the rule cannot rank carries no Default and blocks. `product-level` covers: any
user-visible CLI/config/output change outside the SPEC's own new surface, data or state-format
compatibility, a keep-set category, and pricing/licensing. Kickoff writes
`Operator verdict: DEFAULT-APPLIED <UTC> <runner+role>`.

The existing manager-spec rule "never carries an embedded recommendation or preferred answer" is
amended: a Default selected by the stated rule is a declared fallback, not a recommendation; rows
must still carry Detect → Explain → Ask text. This amendment is decision-index Q2.

## 7. Short recheck

The watchdog, after writing a wait line with `waiting_on=leader`, arms a second recurring
`CronCreate` (`*/5` offset minute) whose prompt is the canonical awaken prompt; on each pass it
deletes the short carrier when a resolving board record exists. Codex runner: no cron; named gap,
the leader's board write plus the nudge is the substitute. Cadence key:
`workflow.watchdog.wait_recheck_minutes` (default 5).

## 8. Messaging hygiene

- Bind cache: `<moai home>/projects/<key>/factory/bind-cache/<session>.json` holding
  `{session, run, pid, process_start, bound_at}`; a full match returns before
  `factorymsg.Open`. Invalidation: any field mismatch, run retirement, or a bind error.
- Degraded inbox state: both call sites pass the state to `slog.Warn` (`.moai/logs/hook-runtime.log`);
  a notice "factory messaging degraded: <state>" is added to the hook context at most once per
  `workflow.factory.degraded_notice_minutes` (default 10) per session, tracked in the same cache file.
- MCP staleness: intake reads the MCP server build from the server instructions line
  (`build vX (commit: SHA ...)`) or `moai doctor`, compares with `moai version`; on mismatch writes
  `mcp: server=<build> cli=<build> fallback=CLI` to progress and uses `moai verify codex-review
  --project-root`, `moai spec audit`, etc.
