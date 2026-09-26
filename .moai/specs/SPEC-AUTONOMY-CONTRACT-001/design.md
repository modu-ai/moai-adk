# design.md — SPEC-AUTONOMY-CONTRACT-001

> Tier L design. The section **§ Contract Schema** is the single addressable schema draft for the
> downstream SPECs A2 (escalation detectors), A3 (gate rewiring), and A4 (closure report). Anything
> those SPECs need from the contract is defined here; if a downstream SPEC needs a new field, it is an
> amendment to this section, not a parallel definition.

## Contract Schema

### File

`.moai/specs/<SPEC-ID>/contract.yaml` — one per SPEC, committed alongside the SPEC artifacts.
Strictly decoded: unknown fields are a `schema_invalid` error (REQ-CONTRACT-003).

### Annotated example (schema_version 1)

```yaml
schema_version: 1
spec_id: SPEC-EXAMPLE-001            # MUST equal the containing directory name

acceptance:
  file: acceptance.md                # fixed value in v1 (relative to the SPEC directory)
  sha256: "<64 lowercase hex>"       # written by `sign`; LF-normalized, BOM-stripped content
  ac_count: 9                        # written by `sign`; published AC counter semantics (live IDs)

invariants:                          # set; order not significant
  - "constitution:CONST-V3R2-*"      # kind=constitution: glob over registry rule IDs; must match >=1
  - "frozen-files"                   # kind=frozen: the constitution frozen-zone file set
  - "go test ./internal/spec/..."    # kind=command: any other string; opaque to A1, run by A2/run-phase

ownership:
  write:                             # set of repo-relative globs; non-empty; must cover .moai/specs/<SPEC-ID>/**
    - "internal/foo/**"
    - ".moai/specs/SPEC-EXAMPLE-001/**"
  never:                             # set of repo-relative globs; may be empty
    - ".claude/rules/moai/core/**"

approach: "<one-paragraph summary of plan.md's approach; approved once here (Rule 1)>"   # non-empty

actions:                             # set; closed vocabulary (see Action Vocabulary)
  - commit
  - worktree
  - local-merge-develop
  - push-develop

reobserve:                           # ordered list; MUST contain "contract.yaml" and "acceptance.md"
  - contract.yaml
  - acceptance.md
  - "git rev-parse HEAD"
  - progress.md

review:
  second_model: codex                # codex | glm | none   ("none" invalid when second_review=required)
  human: closure-report              # v1 value set: closure-report

budget:                              # all integers >= 0; turns and operations >= 1
  turns: 60
  operations: 40
  audit_retries: 2                   # absent budget is filled at sign time from escalation.budget_default

escalate_on:                         # set; MUST equal exactly these six tokens
  - acceptance-change
  - invariant-violation
  - ownership-move
  - new-architecture-or-api
  - contradictory-evidence
  - irreversible-action

plan_audit:
  verdict: PASS                      # PASS | PASS-WITH-DEBT | FAIL  (sign refuses FAIL)
  score: 0.88                        # 0.0..1.0
  report: ".moai/reports/plan-audit/SPEC-EXAMPLE-001-review-1.md"   # repo-relative; must exist at sign time

signature:                           # written ONLY by `sign`; excluded from the digest
  signer: { name: "Jane Doe", email: "jane@example.com" }   # from git config user.name / user.email
  signed_at: "2026-09-26T09:00:00Z"                          # UTC RFC 3339
  head_sha: "<40 or 64 lowercase hex>"                       # HEAD at signing (informational; not re-checked)
  contract_sha256: "<64 lowercase hex>"                      # canonical digest (see Digest)
  mission_contract_sha256: "<64 lowercase hex>"              # sealed hash of the mission projection
  method: interactive-tty                                    # v1 value set: interactive-tty
  batch_id: "<opaque id>"                                    # present only for batch signing
  supersedes: "<64 lowercase hex>"                           # previous contract_sha256 on --resign
```

### Field rules (normative)

| Field | Type | Rule | Reason code on violation |
|---|---|---|---|
| `schema_version` | int | `== 1` | `schema_invalid` |
| `spec_id` | string | equals directory name; passes the SPEC ID regex | `spec_id_mismatch` |
| `acceptance.file` | string | `== "acceptance.md"` | `schema_invalid` |
| `acceptance.sha256` | hex64 | equals measured hash | `acceptance_hash_mismatch` (`acceptance_missing` if the file is absent) |
| `acceptance.ac_count` | int >= 1 | equals measured count | `ac_count_mismatch` / `ac_count_ambiguous` |
| `invariants[]` | string set | non-empty; `constitution:<glob>` matches >= 1 registry ID | `invariant_unresolved` |
| `ownership.write[]` | glob set | non-empty; relative; no `..`; covers `.moai/specs/<ID>/**` | `ownership_invalid` |
| `ownership.never[]` | glob set | relative; no `..`; no literal overlap with `write` | `ownership_invalid` |
| `approach` | string | non-blank | `schema_invalid` |
| `actions[]` | token set | subset of the allowed vocabulary | `unknown_action` / `forbidden_action` / `push_develop_disabled` |
| `reobserve[]` | string list | contains `contract.yaml` and `acceptance.md` | `reobserve_incomplete` |
| `review.second_model` | enum | `codex \| glm \| none`; `none`/empty invalid under `second_review: required` | `second_review_missing` |
| `review.human` | enum | `closure-report` | `schema_invalid` |
| `budget.*` | int | `turns >= 1`, `operations >= 1`, `audit_retries >= 0` | `budget_invalid` |
| `escalate_on[]` | token set | exactly the six tokens | `escalate_on_incomplete` |
| `plan_audit.verdict` | enum | `PASS \| PASS-WITH-DEBT` for a signable/valid contract | `plan_audit_not_passing` |
| `signature` | block | present; `contract_sha256` equals recomputed digest | `unsigned` / `contract_digest_mismatch` |

### Action Vocabulary

| Contract token | Status | Mission projection (`internal/mission` Action) |
|---|---|---|
| `commit` | allowed | `commit` |
| `worktree` | allowed | (no mission equivalent — contract-only) |
| `local-merge-develop` | allowed | `local_develop_merge` |
| `push-develop` | allowed when `push_develop: true`; implies `push_requires_window` | `batch_push` |
| `push-main` | **forbidden** | — (projected into `prohibited_actions` as `main_merge`) |
| `merge-main` | **forbidden** | `main_merge` (prohibited) |
| `force-push` | **forbidden** | `force_push` (prohibited) |
| `release-branch` | **forbidden** | `release_branch` (prohibited) |
| `release-pr` | **forbidden** | `release_pr` (prohibited) |
| anything else | unknown | — |

The forbidden set is fixed by the schema, not by configuration: no contract can authorize a main or
release action (the card: "never main/release").

### Digest

1. Strict-decode the YAML into the typed contract.
2. Drop the `signature` block.
3. Canonicalize: sort every set-valued list (`invariants`, `ownership.write`, `ownership.never`,
   `actions`, `escalate_on`) lexicographically; keep `reobserve` in author order (it is a procedure).
4. Marshal to JSON with the struct's fixed field order; SHA-256; lowercase hex.

This is the same technique as the mission package's contract sealing (clone, sort, marshal, hash),
applied to the SPEC contract type.

### Acceptance hash and count

- Hash: read `acceptance.md`; strip a leading UTF-8 BOM; replace every CRLF with LF; SHA-256; hex.
- Count: a Go port of the published AC counter (the `MOAI-AC-COUNTER` block in the manager-docs agent
  definition): honor the `moai-ac-prefix` declaration, count distinct live IDs, exclude IDs marked
  `[RETIRED]`/`[REF]`, report ambiguity when one ID is both marked and unmarked. A parity test runs
  both the extracted awk counter and the port over every `.moai/specs/*/acceptance.md` and requires
  identical results. The port exists because verify must not spawn a process (REQ-CONTRACT-009) and
  must run on windows.

### Verify Reason Codes

Closed set, emitted in `show --json` / `verify --json` as `reasons: [...]` (sorted, de-duplicated):

`schema_invalid`, `spec_id_mismatch`, `unsigned`, `contract_digest_mismatch`, `acceptance_missing`,
`acceptance_hash_mismatch`, `ac_count_mismatch`, `ac_count_ambiguous`, `unknown_action`,
`forbidden_action`, `push_develop_disabled`, `second_review_missing`, `escalate_on_incomplete`,
`ownership_invalid`, `invariant_unresolved`, `budget_invalid`, `reobserve_incomplete`,
`plan_audit_not_passing`.

Verify collects every applicable code rather than stopping at the first, except that a
`schema_invalid` decode failure stops evaluation (no typed value exists to check further).

### `show --json` / `verify --json` object (stable field names)

```json
{
  "spec_id": "SPEC-EXAMPLE-001",
  "schema_version": 1,
  "state": "signed-valid",
  "valid": true,
  "reasons": [],
  "contract_sha256": "…",
  "recorded_contract_sha256": "…",
  "acceptance": { "sha256": "…", "measured_sha256": "…", "ac_count": 9, "measured_ac_count": 9 },
  "actions": ["commit", "local-merge-develop", "push-develop", "worktree"],
  "push_requires_window": true,
  "second_review": "required",
  "mode": "guided",
  "budget": { "turns": 60, "operations": 40, "audit_retries": 2 },
  "signature": { "signer": {"name": "…", "email": "…"}, "signed_at": "…", "head_sha": "…", "batch_id": "" }
}
```

`state` ∈ `unsigned | signed-valid | signed-invalid`. `show` prints the full contract sections in
addition; `verify --json` prints this object only.

### Exit codes

| Command | 0 | 1 | 2 |
|---|---|---|---|
| `verify` | valid | invalid (reasons printed) | usage / I/O error (e.g. contract file absent) |
| `show` | printed (valid or not) | — | usage / I/O error |
| `sign` | signed | refused (validation, ambiguity, audit verdict, non-TTY, confirmation mismatch, batch disabled) | usage / I/O error |

## Signing Flow

1. Resolve each SPEC ID to its directory; load and strictly decode `contract.yaml`.
2. Refuse when stdin is not a terminal (method `interactive-tty` is the only v1 method).
3. Measure acceptance hash and count; refuse on ambiguity; refuse when a recorded hash differs from
   the measured one (the draft was reviewed against different acceptance criteria).
4. Fill `budget` from `workflow.autonomy.escalation.budget_default` when absent.
5. Run verify on the would-be contract, treating `unsigned` as expected; refuse on any other reason.
6. Refuse when already signed and valid, unless `--resign` (which records `supersedes`).
7. Read signer identity from `git config user.name` / `user.email` (refuse when either is empty) and
   HEAD.
8. Print the summary (per SPEC in batch mode: ID, acceptance hash prefix, AC count, actions, budget,
   second model, plan-audit verdict) and the confirmation token (the SPEC ID, or `sign N contracts`
   in batch mode); read one line; refuse on mismatch.
9. Compute the digest and the mission projection's sealed hash; write the signature block; write each
   file atomically. In `guided` mode print the REQ-CONTRACT-019 notice.

Batch mode (REQ-CONTRACT-013): steps 1-7 run for every ID before step 8; one confirmation; step 9 per
file with a shared `batch_id`. A write failure after some files were written is reported with the
exact written/unwritten split (atomic per file, not across files).

## Mission-Contract Projection

A signed SPEC contract projects onto `mission.MissionContract` so A2 can reuse
`ValidateMissionDecision` rather than build a second decision validator:

| MissionContract field | Source |
|---|---|
| `MissionID` | `spec_id` |
| `Goal` | `approach` |
| `CompletionEvidence` | `["acceptance:" + acceptance.sha256, "ac_count:" + N]` |
| `Scope` | `ownership.write` |
| `AllowedActions` | mapped allowed actions (Action Vocabulary table) |
| `MergeTarget` | `develop` |
| `ResourceLimits` | `{MaxOperations: budget.operations, MaxRetries: budget.audit_retries}` |
| `ProhibitedActions` | `main_merge, force_push, release_branch, release_pr` (always) |
| `StopConditions` | `escalate_on` |
| `RecoveryConditions` | `["re-sign after acceptance change"]` (v1 constant) |
| `RevocationBehavior` | `"stop-before-next-action"` (v1 constant) |
| `PolicyVersion` | `"contract-v1"` |
| `Approved` | `true` only at sign time, after confirmation |

The sealed hash from `SealMissionContract` is recorded as `signature.mission_contract_sha256`.
`ownership.never`, `invariants`, `reobserve`, and `review` have no mission equivalent; they are
covered by the SPEC contract digest, which is why the digest — not the mission hash — is the tamper
authority.

## Configuration

```yaml
workflow:
  autonomy:
    mode: guided              # guided | contract   (template default: guided; A-Q4)
    contract:
      batch_sign: false       # template default false; this repository sets true
      second_review: required # required | advisory | off   (default required; A-Q3)
      push_develop: false     # template default false; this repository sets true (A-Q2)
    escalation:
      budget_default: { turns: 60, operations: 40, audit_retries: 2 }
```

Reader behavior: absent section/key → defaults above; invalid `mode` → `guided` plus a warning;
invalid `second_review` → `required` plus a warning (fail toward the stricter value). No environment
variable is introduced; `MOAI_AUTONOMY_TIER` is untouched.

## F1 Reference Shape

A factory card record that points at a signed contract carries:

```yaml
contract_ref:
  spec_id: SPEC-EXAMPLE-001
  contract_sha256: "<signature.contract_sha256 at the time of linking>"
  signed_at: "2026-09-26T09:00:00Z"
```

The record is a pointer. The authority is `moai contract verify` on the file; a pointer whose
`contract_sha256` differs from the file's current signed digest means the contract was re-signed or
tampered after linking, and the card must re-link. F1 decides storage; A1 guarantees the three fields
are stable and available from `show --json`.

## Package Layout (proposed, run-phase may adjust)

- `internal/contract/` — new package: schema types, strict decode, digest, acceptance hash, AC counter
  port, verify, sign core (TTY and git seams injectable), mission projection.
- `internal/cli/contract.go` — Cobra `contract` command with `sign`, `show`, `verify` subcommands.
- `internal/config/` — `AutonomyConfig` under the workflow section, defaults, reader fail-safe.
- `internal/template/templates/.moai/config/sections/workflow.yaml` and the local
  `.moai/config/sections/workflow.yaml` — the `autonomy` block.
