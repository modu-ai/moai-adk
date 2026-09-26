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
  write:                             # set of repo-relative globs; non-empty; must match .moai/specs/<SPEC-ID>/contract.yaml
    - "internal/foo/**"
    - ".moai/specs/SPEC-EXAMPLE-001/**"
  never:                             # set of repo-relative globs; may be empty
    - ".claude/rules/moai/core/**"

approach: "<one-paragraph summary of plan.md's approach; approved once here (Rule 1)>"   # non-empty

actions:                             # non-empty set; closed vocabulary (see Action Vocabulary)
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

budget:                              # turns and operations >= 1; audit_retries >= 0
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
  verdict: PASS                      # PASS | PASS-WITH-DEBT | FAIL  (sign refuses FAIL; verify reports it)

signature:                           # written ONLY by `sign`; excluded from the digest
  signer: { name: "Jane Doe", email: "jane@example.com" }   # from git config user.name / user.email
  signed_at: "2026-09-26T09:00:00Z"                          # UTC RFC 3339
  head_sha: "<40 or 64 lowercase hex>"                       # HEAD at signing (informational; not re-checked)
  contract_sha256: "<64 lowercase hex>"                      # canonical digest (see Digest)
  method: interactive-tty                                    # v1 value set: interactive-tty
  batch_id: "<opaque id>"                                    # present only for batch signing
  supersedes: "<64 lowercase hex>"                           # previous contract_sha256 on --resign
```

A2 will add the "second review performed" record (spec.md §C.1); it is intentionally absent from v1.

### Field rules (normative)

| Field | Type | Rule | Reason code on violation |
|---|---|---|---|
| `schema_version` | int | `== 1` | `schema_invalid` |
| `spec_id` | string | equals directory name; passes the SPEC ID regex | `spec_id_mismatch` |
| `acceptance.file` | string | `== "acceptance.md"` | `schema_invalid` |
| `acceptance.sha256` | hex64 | equals measured hash | `acceptance_hash_mismatch` (`acceptance_missing` if the file is absent) |
| `acceptance.ac_count` | int >= 1 | equals measured count | `ac_count_mismatch` / `ac_count_ambiguous` (0 recorded → `schema_invalid`) |
| `invariants[]` | string set | non-empty; `constitution:<glob>` matches >= 1 supplied registry ID | `invariant_unresolved` |
| `ownership.write[]` | glob set | non-empty; relative; no `..`; some glob matches `.moai/specs/<ID>/contract.yaml` | `ownership_invalid` |
| `ownership.never[]` | glob set | relative; no `..`; no identical glob string in `write` | `ownership_invalid` |
| `approach` | string | non-blank | `schema_invalid` |
| `actions[]` | token set | non-empty; subset of the allowed vocabulary | `actions_empty` / `unknown_action` / `forbidden_action` / `push_develop_disabled` |
| `reobserve[]` | string list | contains `contract.yaml` and `acceptance.md` | `reobserve_incomplete` |
| `review.second_model` | enum | `codex \| glm \| none`; `none`/empty invalid under `second_review: required` | `second_review_missing` |
| `review.human` | enum | `closure-report` | `schema_invalid` |
| `budget.*` | int | `turns >= 1`, `operations >= 1`, `audit_retries >= 0` | `budget_invalid` |
| `escalate_on[]` | token set | exactly the six tokens | `escalate_on_incomplete` |
| `plan_audit.verdict` | enum | `PASS \| PASS-WITH-DEBT` for a signable/valid contract | `plan_audit_not_passing` |
| `signature` | block | present; `contract_sha256` equals recomputed digest | `unsigned` / `contract_digest_mismatch` |

### Glob Semantics

Ownership globs use doublestar semantics over forward-slash, repo-relative paths: `*` matches within
one path segment, `**` matches zero or more whole segments, `?` matches one non-separator character.
`go.mod` carries no doublestar library, so the matcher is implemented in-package on top of the standard
library's per-segment `path.Match` (no new dependency).
"Covers the SPEC directory" (REQ-CONTRACT-020) means: at least one `write` glob matches the literal
path `.moai/specs/<SPEC-ID>/contract.yaml` under these semantics — so `.moai/**`,
`.moai/specs/**`, and `.moai/specs/<SPEC-ID>/**` all cover it, `.moai/specs/*.md` does not. The
`write`/`never` overlap rule is a literal string comparison only; semantic overlap between different
glob strings is A2's ownership detector's concern.

### Action Vocabulary

| Contract token | Status | Mission action (forward reference for A2) |
|---|---|---|
| `commit` | allowed | `commit` |
| `worktree` | allowed | (no mission equivalent — contract-only) |
| `local-merge-develop` | allowed | `local_develop_merge` |
| `push-develop` | allowed when `push_develop: true`; implies `push_requires_window` | `batch_push` |
| `push-main` | **forbidden** | — |
| `merge-main` | **forbidden** | `main_merge` |
| `force-push` | **forbidden** | `force_push` |
| `release-branch` | **forbidden** | `release_branch` |
| `release-pr` | **forbidden** | `release_pr` |
| anything else | unknown | — |

The forbidden set is fixed by the schema, not by configuration: no contract can authorize a main or
release action (the card: "never main/release"). An empty list is `actions_empty`. Mission-mappability
is not required in A1: with the projection deferred to A2, a contract such as `[worktree]` is valid
schema-wise; A2 decides whether its projection requires at least one mappable action.

### Digest

1. Strict-decode the YAML into the typed contract.
2. Drop the `signature` block.
3. Canonicalize: sort every set-valued list (`invariants`, `ownership.write`, `ownership.never`,
   `actions`, `escalate_on`) lexicographically; keep `reobserve` in author order (it is a procedure).
4. Marshal to JSON with the struct's fixed field order; SHA-256; lowercase hex.

This is the same technique as the mission package's contract sealing (clone, sort, marshal, hash),
applied to the SPEC contract type.

### Acceptance hash and count

- Normalize: strip a leading UTF-8 BOM; replace every CRLF with LF.
- Hash: SHA-256 of the normalized bytes; lowercase hex.
- Count: a Go port of the published AC counter (the `MOAI-AC-COUNTER` block in the manager-docs agent
  definition), applied to the **normalized** content: honor the `moai-ac-prefix` declaration, count
  distinct live IDs, exclude IDs marked `[RETIRED]`/`[REF]`, report ambiguity when one ID is both
  marked and unmarked. The port exists because verify must not spawn a process (REQ-CONTRACT-009) and
  must run on windows.
- Parity: a test runs both the extracted awk counter (on the same normalized bytes) and the port over
  (a) every `.moai/specs/*/acceptance.md` and (b) synthetic fixtures, one per counter branch: prefix
  declaration, `[RETIRED]` marker, `[REF]` marker, ambiguity, CRLF line endings, UTF-8 BOM. A positive
  control asserts the ambiguity fixture is ambiguous in **both** implementations.

### Verify Reason Codes

Closed set, emitted in `show --json` / `verify --json` as `reasons: [...]` (sorted, de-duplicated):

`schema_invalid`, `spec_id_mismatch`, `unsigned`, `contract_digest_mismatch`, `acceptance_missing`,
`acceptance_hash_mismatch`, `ac_count_mismatch`, `ac_count_ambiguous`, `actions_empty`,
`unknown_action`, `forbidden_action`, `push_develop_disabled`, `second_review_missing`,
`escalate_on_incomplete`, `ownership_invalid`, `invariant_unresolved`, `budget_invalid`,
`reobserve_incomplete`, `plan_audit_not_passing`.

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
  "signature": { "signer": {"name": "…", "email": "…"}, "signed_at": "…", "head_sha": "…", "batch_id": "", "supersedes": "" }
}
```

`state` ∈ `unsigned | signed-valid | signed-invalid`. `show` prints the full contract sections in
addition; `verify --json` prints this object only.

### Exit codes

| Command | 0 | 1 | 2 |
|---|---|---|---|
| `verify` | valid | invalid (reasons printed) | usage / I/O error (e.g. contract file absent, path escapes the SPEC directory) |
| `show` | printed (valid or not) | — | usage / I/O error |
| `sign` | signed | refused (validation, ambiguity, zero ACs, audit verdict, non-TTY, agent marker, confirmation mismatch, batch disabled, empty git identity) | usage / I/O error |

## Agent-Environment Markers

Closed set checked by `sign` before anything else (REQ-CONTRACT-021). A marker is "present" when the
variable is set to a non-empty value.

| Variable | Source of the claim |
|---|---|
| `CLAUDECODE` | Measured in this plan session: present in the Bash tool environment of Claude Code |
| `CLAUDE_CODE_SESSION_ID` | `internal/config/envkeys.go` `EnvClaudeCodeSessionID` — stamped into every subprocess Claude Code spawns |

Codex: no Codex session-marker constant exists in this repository (`CODEX_HOME` appears, but it is a
user configuration location that users also set in ordinary shells, so it cannot be a marker). A1
therefore ships no Codex marker. Run-phase may add one only after measuring, from a Codex exec'd
command, a variable that Codex sets and ordinary shells do not; until then the A3 precondition deny
(spec.md §C.2) is the Codex-side protection. The set is extended by amending this table.

These markers raise the bar only: an agent can unset them. They make an accidental agent-issued sign
fail loudly; they are not the A3 gate.

## Signing Flow

1. Refuse when any agent-environment marker is present (name it).
2. Refuse when stdin is not a terminal (method `interactive-tty` is the only v1 method).
3. Resolve each SPEC ID (de-duplicated) to its directory; load and strictly decode `contract.yaml`
   (path containment: a `contract.yaml` symlink resolving outside the SPEC directory is an I/O error,
   exit 2).
4. Branch on signature state:
   - **Unsigned draft**: measure acceptance hash and count; refuse on ambiguity or zero; refuse when a
     recorded hash differs from the measured one (the draft was reviewed against different acceptance
     criteria).
   - **Signed, no `--resign`**: refuse ("already signed; use --resign").
   - **Signed, `--resign`** (REQ-CONTRACT-022): re-measure acceptance hash and count; refuse on
     ambiguity or zero; keep the recorded values for the summary; set `supersedes` to the current
     `signature.contract_sha256`. This is the recovery path after an `acceptance-change` escalation.
5. Fill `budget` from `workflow.autonomy.escalation.budget_default` when absent.
6. Run verify on the would-be contract (new acceptance binding, no signature); refuse on any reason
   other than `unsigned`.
7. Read signer identity from `git config user.name` / `user.email` (refuse when either is empty) and
   HEAD.
8. Print the summary (per SPEC: ID, acceptance hash prefix and AC count — as `old → new` on re-sign —
   actions, budget, second model, plan-audit verdict) and the confirmation token (the SPEC ID, or
   `sign N contracts` in batch mode); read one line; refuse on mismatch.
9. Compute the digest; write the signature block; write each file atomically; print the
   REQ-CONTRACT-019 notice (both modes).

Batch mode (REQ-CONTRACT-013): steps 3-7 run for every ID before step 8; one confirmation; step 9 per
file with a shared `batch_id`. A write failure after some files were written is reported with the
exact written/unwritten split (atomic per file, not across files).

## Forward Note — Mission Projection (for A2)

Deferred from A1 (spec.md §B Out of Scope — Mission-validator projection). Drafted mapping, for A2 to
adopt or revise:

| MissionContract field | Source |
|---|---|
| `MissionID` | `spec_id` |
| `Goal` | `approach` |
| `CompletionEvidence` | `["acceptance:" + acceptance.sha256, "ac_count:" + N]` |
| `Scope` | `ownership.write` — **needs glob→prefix translation** (below) |
| `AllowedActions` | mapped allowed actions (Action Vocabulary table) |
| `MergeTarget` | `develop` |
| `ResourceLimits` | `{MaxOperations: budget.operations, MaxRetries: budget.audit_retries}` |
| `ProhibitedActions` | `main_merge, force_push, release_branch, release_pr` |
| `StopConditions` | `escalate_on` |
| `RecoveryConditions` | `["sign --resign after acceptance change"]` (matches REQ-CONTRACT-022) |
| `RevocationBehavior` | `"stop-before-next-action"` |
| `PolicyVersion` | `"contract-v1"` |

Known issue: `internal/mission/policy.go` `targetInsideScope` is exact-or-prefix
(`target == allowed || HasPrefix(target, allowed+"/")`), so a glob such as `internal/foo/**` never
contains `internal/foo/x.go`. A2 must translate trailing `/**` to a prefix and decide how to treat
globs with inner wildcards, and must require at least one mission-mappable action (the mission sealer
rejects an empty `AllowedActions` as `incomplete_contract`).

## Configuration

```yaml
workflow:
  autonomy:
    mode: guided              # guided | contract   (template default: guided; A-Q4)
    contract:
      batch_sign: false       # template default false; this repository's local config sets true
      second_review: required # required | advisory | off   (default required; A-Q3)
      push_develop: false     # template default false; this repository's local config sets true (A-Q2)
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

## Package Layout

Import constraints are measured facts, not preferences: `go list -deps ./internal/constitution` and
`go list -deps ./internal/config` both include `net` in this tree, and `internal/spec` hosts the
parity test's counter extractor.

- `internal/contract/` — **pure verification core**: schema types, strict decode, digest, acceptance
  normalization/hash, AC counter port, glob matcher, all field rules, verify. Imports the standard
  library (no `os/exec`, no `net`) and `gopkg.in/yaml.v3` (measured: zero `os/exec`/`net` in its
  deps). It does **not** import `internal/config`, `internal/constitution`, or `internal/spec`; policy
  values (mode, second_review, push_develop) and the registry rule IDs are passed in by the caller.
  Its own SPEC ID pattern is a local copy of the lint pattern, pinned by a test in `internal/spec`
  comparing the two.
- `internal/contract/sign/` — signing: TTY and agent-marker checks, git identity/HEAD (subprocess),
  confirmation, atomic writes via `internal/atomicfile`. Imports `internal/contract`.
- `internal/cli/contract.go` — Cobra `contract` command with `sign`, `show`, `verify`; loads config and
  the constitution registry and passes values into the core.
- `internal/config/` — `AutonomyConfig` under the workflow section, defaults, reader fail-safe.
- Parity test — lives in `internal/spec` (package `spec`, beside the unexported
  `extractCounterCommand`) and imports `internal/contract`; no cycle because `internal/contract` never
  imports `internal/spec`.
- `internal/template/templates/.moai/config/sections/workflow.yaml` and the local
  `.moai/config/sections/workflow.yaml` — the `autonomy` block.
