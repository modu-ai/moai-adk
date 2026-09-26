# design.md — SPEC-AUTONOMY-CLOSURE-001

> Tier L design. Decisions most likely to change come first (record shapes, then the readiness
> rule, then layout and wiring). spec.md states observable behavior; everything here is the proposal
> the run phase implements. Contract facts come from A1 (SPEC-AUTONOMY-CONTRACT-001, branch
> `WT-contract-schema`, read at `67a2f55cb`); escalation facts from A2 (SPEC-AUTONOMY-ESCALATION-001
> §I, read at `8c9ee29b7`). Neither schema is copied here.

## §A. Records A4 owns

A4 introduces three files per card, all under `.moai/reports/<card-id>/` (gitignored, local-only;
research.md §B.7). None of them is `verdict.md`, which stays the lead's hand-authored file.

| File | Writer | Shape |
|---|---|---|
| `second-review.jsonl` | the MCP server, inside `audit_multi`, only when `card_id` is passed | append-only, one JSON object per line |
| `closure-report.md` / `closure-report.json` | `moai contract report` | rewritten atomically as a pair |
| `closure-verdict.jsonl` | `moai contract verdict` (human path only) | append-only, one JSON object per line |

### §A.1 Second-review record (`second-review.jsonl`, one line)

```json
{
  "schema_version": 1,
  "card": "t1234",
  "spec_id": "SPEC-EXAMPLE-001",
  "contract_sha256": "<signature.contract_sha256, or \"\" when absent/unsigned>",
  "head_sha": "<HEAD of project_root at audit time>",
  "backends": [
    { "backend": "claude", "gate": "required", "verdict": "pass" },
    { "backend": "codex",  "gate": "required", "verdict": "fail" },
    { "backend": "glm",    "gate": "advisory", "verdict": "inconclusive" }
  ],
  "participant_count": 2,
  "disagreement_flag": true,
  "audit_receipt": "rcpt-…",
  "build_commit": "<commit of the serving binary>",
  "recorded_at": "2026-09-26T09:00:00Z"
}
```

Field sources: `backends[]`, `participant_count`, `disagreement_flag`, `audit_receipt`, and
`build_commit` are copied from the `ConvergenceResult` the handler already assembles
(research.md §B.2); `spec_id` from the queue store; `contract_sha256` from A1 `Verify` on that SPEC;
`head_sha` from the audited tree. `disagreement_flag` keeps its nullable third state (`null`).

### §A.2 Human verdict record (`closure-verdict.jsonl`, one line)

```json
{
  "schema_version": 1,
  "card": "t1234",
  "spec_id": "SPEC-EXAMPLE-001",
  "verdict": "accept",
  "note": "",
  "operator": { "name": "Jane Doe", "email": "jane@example.com" },
  "recorded_at": "2026-09-26T09:00:00Z",
  "report_sha256": "<SHA-256 of closure-report.json raw bytes at recording>",
  "method": "interactive-tty"
}
```

`verdict` ∈ `accept | reject | amend-contract`. The latest line wins; earlier lines stay as history.

### §A.3 Closure report JSON (`closure-report.json`)

Top-level keys in the fixed section order of REQ-CLOSURE-002:

```
schema_version, card, spec_id, head_sha, generated_at, mode, second_review_policy,
summary, kickoff, reconciliation, invariants, ownership, new_apis, escalations,
first_verdict, second_verdict, plan_audit_binding, not_performed[], residual_risk[],
human_verdict, sources{}
```

`sources` maps each evidence file role (`progress`, `acceptance`, `contract`, `receipt`,
`second_review`, `human_verdict`, `escalation_dir`, `plan_audit_report`) to the path actually read,
or `""` when not found (REQ-CLOSURE-024). The Markdown file renders the same values as headings in
the same order; the renderer takes the JSON model as its only input, so the two cannot diverge.

## §B. Not-performed catalogue

Each entry in `not_performed[]` is `{ "item": <token>, "detail": <text> }`. Closed token set:

| Token | Raised when |
|---|---|
| `progress-missing` | no `progress.md` for the SPEC |
| `ac-not-reported` | a live AC ID has no §E.2 row |
| `first-verdict-missing` | §E.3 block absent or a field absent |
| `second-review-not-performed` | REQ-CLOSURE-013 state is not performed |
| `receipt-missing` | signature method `receipt` and no receipt file |
| `plan-audit-self-reported` | binding state `self-reported` |
| `plan-audit-mismatch` | binding state `mismatch` |
| `escalation-unreadable` | an escalation record failed to parse |
| `escalation-detector-not-armed` | A2 card state file absent or not armed for the contract in force |
| `invariant-not-observed` | an invariant result is `not observed` |
| `new-api-comparison-unavailable` | the class-4 comparison could not run |
| `human-verdict-none` | no verdict recorded |
| `human-verdict-stale` | latest verdict bound to a different report hash |

## §C. Push readiness

### §C.1 Which pushes are evaluated

Only under `workflow.autonomy.mode: contract`, only for a Bash tool call classified as a push of the
integration branch. The integration branch is `config.LoadGitFlowDevelopBranch(projectRoot)`
(research.md §B.9). Classification reuses A2b's push-command classifier when it is exported at run
time; otherwise A4 ships a minimal classifier: `git push` whose explicit refspec destination is the
integration branch, or a bare `git push` / `git push <remote>` while the checked-out branch is the
integration branch.

### §C.2 Which contracts are in the push

1. Range: `<remote>/<integration>..HEAD`, non-merge commits only (`git log --no-merges --name-only`).
2. Candidate SPECs: every `.moai/specs/<ID>/contract.yaml` at HEAD whose A1 state is signed (valid or
   invalid) and whose SPEC is not terminal (A1 derived `terminal`).
3. A candidate is **in the push** when a range commit changes a path under `.moai/specs/<ID>/` or a
   path matched by its `ownership.write` globs (A1 glob semantics), and its `actions` list contains
   `push-develop`.

### §C.3 Card of a SPEC and its evidence directory

Card IDs for a SPEC come from the queue store (`spec_id` reverse lookup). Evidence directory per
card, per file, first hit wins (REQ-CLOSURE-024):

1. `<card worktree>/.moai/reports/<card-id>/` — the worktree whose directory base name equals the
   card ID in `git worktree list --porcelain`.
2. `<primary checkout>/.moai/reports/<card-id>/` — the primary checkout is the parent of
   `git rev-parse --git-common-dir`.

No card, or several cards with none carrying evidence, is `push_check_undetermined`.

### §C.4 Readiness rule per in-push contract

Evaluated in full; every applicable code is collected, sorted, de-duplicated.

| Code | Condition |
|---|---|
| `contract_invalid` | A1 `Verify` state ≠ `signed-valid` |
| `closure_report_missing` | no `closure-report.json` found for the card |
| `second_review_not_performed` | policy `required` and REQ-CLOSURE-013 state is not performed, causes `no-record`, `no-second-model`, `unbound`, or `contract-changed` |
| `second_review_stale` | policy `required`, a performed record exists, but a range commit after its `head_sha` changes the contract's paths (cause `stale-head`) |
| `second_review_failed` | policy `required` and the performed verdict is `fail` |
| `human_verdict_reject` / `human_verdict_amend_contract` | per OQ-1 (plan.md); proposed default: latest recorded verdict is `reject` / `amend-contract`, current or stale |
| `push_check_undetermined` | range, card, or evidence cannot be determined; git timeout |

A push is ready when no in-push contract carries a code. A range with no in-push contract is ready
(nothing A4 governs).

### §C.5 Surfaces

| Surface | Ready | Not ready | Error |
|---|---|---|---|
| PreToolUse hook | allow (fall through to the remaining guards) | deny, reason `CLOSURE_PUSH_STOP: <SPEC-ID>=<code>[,<code>]…[; …]` | deny, `push_check_undetermined` |
| `moai contract push-check` | exit 0, one line `ready` | exit 1, one line per SPEC with codes | exit 2 usage / I/O |
| `moai contract push-check` under `guided` | exit 0, one line `inactive (mode guided)` | — | — |

## §D. Human verdict path

`moai contract verdict <card-id> <accept|reject|amend-contract> [--note <text>]`:

1. Refuse `agent_marker` when any variable of A1's closed marker set is non-empty (A1 design
   § Agent-Environment Markers; imported, not copied).
2. Refuse `not_tty` when standard input is not a terminal.
3. Refuse `report_missing` when no `closure-report.json` is found for the card.
4. Print the report summary (card, SPEC, second-review state, open escalation count, report hash
   prefix) and the confirmation token `<verdict> <card-id>`; refuse `confirmation_mismatch` on a
   different line.
5. Refuse `git_identity_missing` when `user.name` or `user.email` is empty.
6. Append one §A.2 line; exit 0.

Refusals exit 1 and write nothing; usage errors exit 2. The PreToolUse deny
(`CLOSURE_VERDICT_HUMAN_ONLY:`) is the agent-side protection, placed beside A2b's deny on
`moai contract sign` when that exists, otherwise as its own check after `checkBashCommand`.

## §E. Hook placement

`internal/hook/pre_tool.go:507` calls `checkBashCommand` under an `@MX:ANCHOR` that forbids any
conditional return above it (research.md §B.8). Both A4 checks run **after** it, next to the branch
guard (`pre_tool.go:534`):

1. `moai contract verdict` deny — string match on the Bash command; every mode; no I/O.
2. Push readiness — first statement is the mode check against the already-loaded configuration;
   under `guided` it returns before any file read or subprocess (REQ-CLOSURE-023). Git work runs
   with a bounded timeout.

## §F. Report assembly

| Section | Inputs | Library |
|---|---|---|
| Summary | card, SPEC, HEAD, mode, policy, A1 verify state | A1 `contract.Verify` |
| Kickoff | contract `signature`, receipt file (lenient decode) | own lenient decoder; A1 verify status |
| Reconciliation | `acceptance.md` live IDs; `progress.md` §E.2 rows | A1 counter semantics for IDs; own row parser |
| Invariants | contract `invariants`; A2 records; A2 card state file | A2 record reader |
| Ownership | A2 `ownership-move` records; card state file | A2 record reader |
| New APIs | A2 class-4 comparison (read-only); A2 class-4 records | A2 detector function |
| Escalations | `.moai/reports/<card-id>/escalation/*.md` | A2 record reader |
| First Verdict | `progress.md` §E.3 fenced YAML | own parser |
| Second Verdict | `second-review.jsonl` | §A.1 decoder |
| Plan-Audit Binding | receipt `inputs.plan_audit_report`; plan-audit report files | `auditreceipt.ParseVerdictLine` |
| Human Verdict | `closure-verdict.jsonl` | §A.2 decoder |

The receipt is decoded leniently for display (unknown fields ignored, any decider token, any number
of decisions) because A1's validator is strict and the decider value set is still moving
(research.md §C.1). The A1 verify result states whether the receipt is valid; the report never
re-validates it.

## §G. Package layout

- `internal/closure/` — report model (§A.3), section builders, Markdown renderer, record decoders
  (§A.1, §A.2), not-performed catalogue, readiness evaluator (§C.4) as a pure function over injected
  inputs. Imports `internal/contract` (A1 core) and A2's record reader; imports no `internal/hook`.
- `internal/closure/gitio/` — range listing, worktree listing, HEAD reads; the only subprocess user,
  behind an interface the pure core receives.
- `internal/cli/contract_report.go`, `contract_verdict.go`, `contract_pushcheck.go` — subcommands on
  A1's `contract` command.
- `internal/cli/mcp_audit_multi.go`, `internal/cli/mcp_server.go` — optional `card_id` argument and
  the record append after the existing convergence persistence.
- `internal/hook/pre_tool.go` — the two checks of §E.
- Template + local mirrors — `moai-ref-cross-model-audit` skill and `sync-auditor` agent text;
  `make agents-emit` regenerates the Codex agent copy.

## §H. Alternatives considered

- **Write the report to `verdict.md`** (lead design wording) — rejected: overwrites the lead's
  hand-authored verdict files (research.md §B.6); OQ-2 asks for confirmation.
- **Amend A1's contract schema with a `second_review_performed` field** — rejected: the contract is
  signed and immutable during a run (A1 effective `never`); a run-time record is the only place a
  post-signing fact can live.
- **Read `.moai/state/audit-multi/<session>.json`** — rejected as the sole source: it is keyed by
  session, written only when a session id is passed, and carries no card, SPEC, or HEAD
  (research.md §B.2).
- **Gate in the git `pre-push` hook** — rejected as the primary layer: this repository disables git
  hooks (`core.hooksPath=/dev/null`) and the hook has a documented bypass variable; kept available
  for manual use through `moai contract push-check`.
- **Allow on an undetermined push** — rejected: A-Q3 forbids silently skipping the stop.
