# design.md — SPEC-AUTONOMY-CLOSURE-001

> Tier L design. Decisions most likely to change come first (record shapes, then the readiness
> rule, then the push classifier, then layout and wiring). spec.md states observable behavior; this
> file is the proposal the run phase implements. Contract facts come from A1
> (SPEC-AUTONOMY-CONTRACT-001 v0.5.1, read at `65e0a9167` on branch `WT-contract-schema`); escalation
> facts from A2 (SPEC-AUTONOMY-ESCALATION-001 v0.3.0 §I, read at `8c9ee29b7`). Neither schema is
> copied here.

## §A. Records A4 owns

All three live in the **card evidence directory** `<card evidence home>/.moai/reports/<card-id>/`
(spec.md §C.7; gitignored, research.md §B.7). None of them is `verdict.md`.

| File | Writer | Shape |
|---|---|---|
| `second-review.jsonl` | the MCP server, inside `audit_multi`, only when `card_id` is passed | append-only, one JSON object per line |
| `closure-report.md` / `closure-report.json` | `moai contract report` | rewritten atomically as a pair |
| `closure-verdict.jsonl` | `moai contract verdict` (human path only) | append-only, one JSON object per line |

### §A.1 Second-review record (one line)

```json
{
  "schema_version": 1,
  "card": "c1",
  "contract_card": "c1",
  "spec_id": "SPEC-EXAMPLE-001",
  "contract_sha256": "<signature.contract_sha256, or \"\">",
  "head_sha": "<HEAD of project_root at audit time>",
  "target": "baseBranch",
  "scope": { "base_branch": "<name from resolveReviewBaseBranchName>", "base_sha": "<merge-base of head and base_branch>",
             "head_sha": "<same as head_sha>",
             "changed_files": 7, "diff_sha256": "<SHA-256 of git diff base..head>" },
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

Sources: `backends[]`, `participant_count`, `disagreement_flag` (nullable, kept as `null`),
`audit_receipt`, `build_commit` come from the `ConvergenceResult` the handler already builds
(research.md §B.2); `spec_id` from the queue store (card → SPEC); `contract_card` and
`contract_sha256` from A1 `Verify` on that SPEC; `head_sha` from git in `project_root`; `scope` from
the same base resolution the `baseBranch` review backends use (`resolveReviewBaseBranchName` →
remote default head, then `main`, `internal/cli/mcp_review_material.go:131`; merge base as in
`resolveReviewMergeBase`, `:92`), so the recorded scope is the reviewed scope even where the remote
default head is not the integration branch.

Failure paths (spec.md REQ-CLOSURE-012): unknown card or card without SPEC → record written with
`spec_id: ""` (the reader classifies it `unbound`); contract absent → `contract_card: ""`,
`contract_sha256: ""`; git failure → `scope` fields empty (reader: `scope-not-covered`); append failure
→ result gains `second_review_record_error: "<cause>"`. The audit result itself is never altered.

### §A.2 Human verdict record (one line)

```json
{
  "schema_version": 1, "card": "c1", "spec_id": "SPEC-EXAMPLE-001",
  "verdict": "accept", "note": "",
  "operator": { "name": "Jane Doe", "email": "jane@example.com" },
  "recorded_at": "2026-09-26T09:00:00Z",
  "report_sha256": "<SHA-256 of closure-report.json raw bytes at recording>",
  "method": "interactive-tty"
}
```

`verdict` ∈ `accept | reject | amend-contract`. The latest line wins; earlier lines stay as history.

### §A.3 Closure report JSON

Top-level keys in the fixed section order of REQ-CLOSURE-002:

```
schema_version, card, spec_id, head_sha, generated_at, mode, second_review_policy,
summary, kickoff, reconciliation, invariants, ownership, new_apis, escalations,
first_verdict, second_verdict, plan_audit_binding, not_performed[], residual_risk[],
human_verdict, sources{}
```

`head_sha` is the card evidence home's HEAD at build time; the push evaluator applies the currency
rule to it (`closure_report_stale`). `sources` maps each evidence role to the path read, or `""`. The
Markdown renderer takes the JSON model as its only input.

## §B. Not-performed catalogue

`not_performed[]` entries are `{ "item": <token>, "detail": <text> }`. Closed token set:

| Token | Raised when |
|---|---|
| `progress-missing` | no `progress.md` |
| `ac-not-reported` | a live AC ID has no §E.2 row |
| `first-verdict-missing` | §E.3 block or one of its fields absent |
| `second-review-not-performed` | REQ-CLOSURE-013 state `not performed` |
| `second-review-stale` | REQ-CLOSURE-013 state `stale` |
| `receipt-missing` | method `receipt` and no receipt file |
| `receipt-field-unrecognized` | a receipt field the display decoder does not know |
| `plan-audit-self-reported` / `plan-audit-mismatch` | binding states |
| `escalation-unreadable` | an escalation record failed to parse |
| `escalation-detector-not-armed` | A2 card state file absent, disarmed, or bound to another contract |
| `invariant-not-observed` | an invariant result is `not observed` |
| `new-api-comparison-unavailable` | the class-4 comparison could not run |
| `human-verdict-none` / `human-verdict-stale` | verdict states |

## §C. Push readiness

### §C.1 Push classification (fail-closed)

Evaluated only under `mode: contract`, on the Bash command text. Integration branch
`I = config.LoadGitFlowDevelopBranch(root)` (research.md §B.9); the tree is the `-C <path>` argument
when present, else the tool call's working directory. Reuse A2b's classifier if it is exported at run
time; otherwise A4's:

| Form | Destination / source | Result |
|---|---|---|
| `git [-C p] [-c k=v] push <remote> <src>:<dst>` (optional leading `+`) | dst normalized (`refs/heads/X` → `X`); src = named ref, `HEAD`, or SHA | evaluate when dst = `I` |
| `git push <remote> <name>` / `+<name>` / `refs/heads/<name>` | dst = src = `<name>` | evaluate when name = `I` |
| `git push` / `git push <remote>` | dst = upstream of the tree's current branch | evaluate when upstream branch = `I`; unresolvable upstream → undetermined |
| `--all`, `--mirror` | may include `I` | undetermined (deny under contract mode) |
| `--tags`, `--delete`, refspec with dst ≠ `I` | not an integration push | not evaluated (A2 class 6 still reports) |
| `git push` inside `sh -c`, `bash -c`, `$(…)`, backticks, `eval`, or with a `$VAR` operand | cannot be proven | undetermined |

Undetermined → `push_check_undetermined` → deny.

### §C.2 Which contracts are in the push

1. Source commit `S` = the resolved source of the classified refspec, in the tree of §C.1.
2. Range `<remote>/I..S` (non-merge commits, `git log --no-merges --name-only`). A missing
   remote-tracking ref → undetermined.
3. Candidates: every `.moai/specs/<ID>/contract.yaml` **in the tree of `S`** (`git show S:<path>`,
   passed to A1 `Verify` as inputs), signed, `actions` containing `push-develop`. **No terminal
   filter:** the sync commit sets `status: completed` inside the card worktree before the merge, so in
   `S` every normally-closed card is terminal; excluding terminal SPECs would exempt exactly the cards
   the stop exists for. Membership is scoped by the range (step 4) instead.
4. In the push: a range commit changes a governed path of the contract or a path under
   `.moai/specs/<ID>/`.

### §C.3 Card evidence home

For a contract's signed `card` value `C`: `git worktree list --porcelain` in the tree of §C.1; the
entry whose directory base name equals `C` is the home; none → the primary checkout (parent of
`git rev-parse --git-common-dir`). The same rule serves `moai contract report`, `moai contract verdict`,
and the `audit_multi` record writer. Unreadable worktree list → undetermined.

### §C.4 Readiness per in-push contract

Collected in full, sorted, de-duplicated:

| Code | Condition |
|---|---|
| `contract_invalid` | A1 `Verify` state ≠ `signed-valid` |
| `closure_report_missing` | no `closure-report.json` in the card evidence directory |
| `closure_report_stale` | report `head_sha` fails the currency rule for `S` |
| `human_verdict_reject` / `human_verdict_amend_contract` | latest verdict line is `reject` / `amend-contract` (current or stale) |
| `second_review_not_performed` | policy `required`, REQ-CLOSURE-013 state `not performed` (evaluation commit `S`) |
| `second_review_stale` | policy `required`, state `stale` |
| `second_review_failed` | policy `required`, state `performed` with verdict `fail` |
| `push_check_undetermined` | §C.1-§C.3 undetermined, or a git call timed out |

A range with no in-push contract is ready. A push is ready when no in-push contract carries a code.

### §C.5 Surfaces

| Surface | Ready | Not ready | Error |
|---|---|---|---|
| PreToolUse hook | fall through to remaining guards | deny `CLOSURE_PUSH_STOP: <SPEC-ID>=<code>[,<code>]…[; …]` | deny with `push_check_undetermined` |
| `moai contract push-check` | exit 0, `ready` | exit 1, one line per SPEC with codes (undetermined included) | exit 2 on usage error only |
| `moai contract push-check` under `guided` | exit 0, `inactive (mode guided)` | — | — |

## §D. Second-review selection (REQ-CLOSURE-013)

```
records := decode(second-review.jsonl)             // unknown schema_version → skipped, listed
for filter in [bound, same-contract, scope-covered, second-model, in-history]:
    kept := records passing filter
    if kept is empty: return not-performed(cause = filter's cause)   // no-record if records empty at start
    records = kept
r := latest(records by recorded_at)
verdict := fail if any counted codex/glm verdict == fail else pass
if not current(r.head_sha, P): return stale(superseding commit)
return performed(verdict)
```

`P` is the report HEAD (report) or `S` (push). Currency and "in history" share one ancestry check.

## §E. Human verdict path

`moai contract verdict <card-id> <accept|reject|amend-contract> [--note <text>]`: refuse
`agent_marker` (A1 marker set, imported), `not_tty`, `report_missing`, `confirmation_mismatch`
(token `<verdict> <card-id>`), `git_identity_missing`; else append §A.2. Refusals exit 1, usage 2.
The PreToolUse deny `CLOSURE_VERDICT_HUMAN_ONLY:` is a convenience guard (spec.md §H).

## §F. Hook placement

`internal/hook/pre_tool.go:507` calls `checkBashCommand` under an `@MX:ANCHOR` forbidding a
conditional return above it. Both A4 checks run after it, beside the branch guard (`:534`): the verdict
deny (string match, every mode, no I/O), then push readiness (mode check first against the loaded
configuration; under `guided` it returns before any file read or subprocess).

## §G. Report assembly and template markers

| Section | Inputs | Library |
|---|---|---|
| Summary | card, SPEC, HEAD, mode, policy, A1 verify state | A1 `contract.Verify` |
| Kickoff | contract `signature`; receipt (display decoder for v0.5.1 fields: `requested_decider`, `effective_decider`, `fallback{applied,reason}`, `llm_answer`, `jev_answer`, `outcome`) | own display decoder; A1 verify status |
| Reconciliation | `acceptance.md` live IDs; `progress.md` §E.2 rows | A1 counter semantics; own row parser |
| Invariants / Ownership / Escalations | A2 records; A2 card state file | A2 record reader |
| New APIs | A2 class-4 comparison (read-only); A2 class-4 records | A2 detector function |
| First Verdict | `progress.md` §E.3 fenced YAML | own parser |
| Second Verdict | `second-review.jsonl` | §D |
| Plan-Audit Binding | receipt input path; `plan-audit*.md`; `plan-audit/<SPEC-ID>-review-<n>.md` | `auditreceipt.ParseVerdictLine` |
| Human Verdict | `closure-verdict.jsonl` | §A.2 decoder |

The display decoder maps unknown fields to `receipt-field-unrecognized` entries instead of dropping
them. Template auditor text sits between the literal lines
`<!-- moai:closure-second-review:start -->` and `<!-- moai:closure-second-review:end -->` in the
`moai-ref-cross-model-audit` skill and the `sync-auditor` agent (template and local copies).

## §H. Package layout

- `internal/closure/` — model (§A.3), section builders, Markdown renderer, record decoders (§A.1,
  §A.2), not-performed catalogue, selection (§D), readiness evaluator (§C.4) as pure functions over
  injected inputs. Imports `internal/contract` and A2's record reader; never `internal/hook`.
- `internal/closure/gitio/` — push classifier helpers, range listing, worktree listing, blob reads;
  the only subprocess user, behind an interface.
- `internal/cli/contract_report.go`, `contract_verdict.go`, `contract_pushcheck.go`.
- `internal/cli/mcp_audit_multi.go`, `mcp_server.go`, `mcp_convergence.go` — `card_id` argument and
  record append after the existing persistence step.
- `internal/hook/pre_tool.go`, `internal/hook/closure_push.go` — the two checks of §F.
- Template + local mirrors of the skill and agent text; `make agents-emit`.

## §I. Alternatives considered

- **`verdict.md` as the report file** — rejected by lead decision (OQ-2).
- **Amend A1's schema with a performed flag** — rejected: the contract is immutable after signing.
- **Range from `HEAD`** — rejected (plan-audit D5): a push of the integration branch issued from a tree
  on another branch evaluated the wrong range and passed.
- **Exact HEAD equality for review currency** — rejected (D6): the pushed integration commit is a
  `--no-ff` merge and never equals the card commit that was reviewed.
- **Allow on an undetermined push** — rejected: A-Q3 forbids silently skipping the stop.
