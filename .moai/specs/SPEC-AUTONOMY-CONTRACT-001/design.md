# design.md — SPEC-AUTONOMY-CONTRACT-001

> Tier L design. The section **§ Contract Schema** is the single addressable schema draft for the
> downstream SPECs A2 (escalation detectors, card t1235), A2b (push serialization and the
> `moai contract sign` deny, card t1245), A3 (gate rewiring, revocation, receipt
> issuance), and A4 (closure report, second-review stop). Anything those SPECs need from the contract
> is defined here; if a downstream SPEC needs a new field, it is an amendment to this section, not a
> parallel definition.

## Contract Schema

### File

`.moai/specs/<SPEC-ID>/contract.yaml` — one per SPEC, committed alongside the SPEC artifacts.
Strictly decoded: unknown fields are a `schema_invalid` error (REQ-CONTRACT-003).

### Annotated example (schema_version 1)

```yaml
schema_version: 1
spec_id: SPEC-EXAMPLE-001            # MUST equal the containing directory name
card: t1234                          # REQUIRED card id; covered by the digest (see § Card Field)

acceptance:
  file: acceptance.md                # fixed value in v1 (relative to the SPEC directory)
  sha256: "<64 lowercase hex>"       # written by `sign`; LF-normalized, BOM-stripped content
  ac_count: 9                        # written by `sign`; published AC counter semantics (live IDs)

invariants:                          # set; order not significant
  - "constitution:CONST-V3R2-*"      # kind=constitution: glob over registry rule IDs; must match >=1
  - "frozen-files"                   # kind=frozen: see § Frozen Files
  - "go test ./internal/spec/..."    # kind=command: any other string; opaque to A1, run by A2/run-phase

ownership:
  write:                             # set of repo-relative globs; non-empty; must match .moai/specs/<SPEC-ID>/contract.yaml
    - "internal/foo/**"
    - ".moai/specs/SPEC-EXAMPLE-001/**"
  never:                             # set of repo-relative globs; may be empty
    - ".claude/rules/moai/core/**"
  scratch:                           # OPTIONAL set of repo-relative globs; exempt from ownership escalation
    - ".moai/state/verify/**"        # (in addition to the detectors' fixed exemptions, which A2 defines)

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
  audit_retries: 2                   # SHARED retry cap for plan-audit and sync-audit; absent budget is
                                     # filled at sign time from escalation.budget_default

escalate_on:                         # set; MUST equal exactly these six tokens
  - acceptance-change
  - invariant-violation
  - ownership-move
  - new-architecture-or-api
  - contradictory-evidence
  - irreversible-action

plan_audit:
  verdict: PASS                      # PASS | PASS-WITH-DEBT | FAIL — self-reported (spec.md §C.7)

signature:                           # written ONLY by `sign`; excluded from the digest
  signer_kind: human                 # human | llm | llm+jev   (receipt path: the receipt's effective decider)
  operator: { name: "Jane Doe", email: "jane@example.com" }  # from git config user.name / user.email
  signed_at: "2026-09-26T09:00:00Z"                          # UTC RFC 3339
  head_sha: "<40 or 64 lowercase hex>"                       # HEAD at signing (informational; not re-checked)
  contract_sha256: "<64 lowercase hex>"                      # canonical body digest (see Digest)
  acceptance_sha256: "<64 lowercase hex>"                    # acceptance.md hash at signing
  method: interactive-tty                                    # interactive-tty | receipt
  receipt:                                                   # present only when method=receipt
    path: kickoff-receipt.json                               # relative to the SPEC directory (fixed name)
    sha256: "<64 lowercase hex>"                             # raw bytes of the receipt file
    provenance: file                                         # value set: file. A3 verifies receipts against its store ($MOAI_HOME/db/<project-key>/contract/), not via a new provenance value
  batch_id: "<opaque id>"                                    # present only for batch signing
  supersedes: "<64 lowercase hex>"                           # previous contract_sha256 on --resign
  seal: "<64 lowercase hex>"                                 # see § Signature Seal; covers every field above
```

The signature stores the digest of the contract **body** (a file cannot contain its own whole-file
hash) and the hash of `acceptance.md`; together they pin both files as they were at signing (B1). The
record that a second-model review was performed is A4's addition (spec.md §C.1).

### Signature Seal

`seal = SHA-256( canonical_json(signature without seal) ‖ contract_sha256 )`, lowercase hex, where
`canonical_json` marshals the typed signature block with its fixed field order (absent optional fields
omitted) and `‖` is byte concatenation with the 64-character hex digest. The seal covers `signer_kind`,
`operator`, `signed_at`, `head_sha`, `contract_sha256`, `acceptance_sha256`, `method`, `receipt.*`
(`path`, `sha256`, `provenance`), `batch_id`, and `supersedes`. It is keyless: it detects an edit made
without recomputation; it does not stop a forger who recomputes it (spec.md §H).

Signature consistency rules (checked by verify, reason code `signature_inconsistent`):

| `method` | `signer_kind` | `receipt` block | `receipt.provenance` |
|---|---|---|---|
| `interactive-tty` | `human` | absent | — |
| `receipt` | `llm` or `llm+jev` | present | `file` (the only value; any other value is inconsistent — A3 adds no provenance value, it cross-checks the receipt against its store) |

Any other combination is `signature_inconsistent`. In addition, `signature.acceptance_sha256` must equal
the measured acceptance hash (reason code `signature_acceptance_mismatch`).

### Field rules (normative)

| Field | Type | Rule | Reason code on violation |
|---|---|---|---|
| `schema_version` | int | `== 1` | `schema_invalid` |
| `spec_id` | string | equals directory name; passes the SPEC ID regex | `spec_id_mismatch` |
| `card` | string | present, non-empty, matches `^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$` (§ Card Field) | `card_invalid` |
| `acceptance.file` | string | `== "acceptance.md"` | `schema_invalid` |
| `acceptance.sha256` | hex64 | equals measured hash | `acceptance_hash_mismatch` (`acceptance_missing` if the file is absent) |
| `acceptance.ac_count` | int >= 1 | equals measured count | `ac_count_mismatch` / `ac_count_ambiguous` (0 recorded → `schema_invalid`) |
| `invariants[]` | string set | non-empty; `constitution:<glob>` matches >= 1 supplied registry ID | `invariant_unresolved` |
| `ownership.write[]` | glob set | non-empty; relative; no `..`; some glob matches `.moai/specs/<ID>/contract.yaml` | `ownership_invalid` |
| `ownership.never[]` | glob set | relative; no `..`; no identical glob string in `write` or `scratch` | `ownership_invalid` |
| `ownership.scratch[]` | glob set, optional | relative; no `..` | `ownership_invalid` |
| `approach` | string | non-blank | `schema_invalid` |
| `actions[]` | token set | non-empty; subset of the allowed vocabulary | `actions_empty` / `unknown_action` / `forbidden_action` / `push_develop_disabled` |
| `reobserve[]` | string list | contains `contract.yaml` and `acceptance.md` | `reobserve_incomplete` |
| `review.second_model` | enum | `codex \| glm \| none`; `none`/empty invalid under `second_review: required` | `second_review_missing` |
| `review.human` | enum | `closure-report` | `schema_invalid` |
| `budget.*` | int | `turns >= 1`, `operations >= 1`, `audit_retries >= 0` | `budget_invalid` |
| `escalate_on[]` | token set | exactly the six tokens | `escalate_on_incomplete` |
| `plan_audit.verdict` | enum | `PASS \| PASS-WITH-DEBT` for a signable/valid contract | `plan_audit_not_passing` |
| `signature` | block | present; `contract_sha256` equals recomputed digest | `unsigned` / `contract_digest_mismatch` |
| `signature.seal` | hex64 | equals recomputation (§ Signature Seal) | `signature_seal_mismatch` |
| `signature.method` / `signer_kind` / `receipt` / `receipt.provenance` | block | consistency table (§ Signature Seal) | `signature_inconsistent` |
| `signature.acceptance_sha256` | hex64 | equals measured acceptance hash | `signature_acceptance_mismatch` |
| `signature.receipt` | block | when `method: receipt`: file exists at the fixed path and its SHA-256 equals `sha256` | `receipt_mismatch` |

### Card Field

`card` is a REQUIRED top-level string naming the card the contract belongs to. A2's resolver arms a
contract only when its `card` equals the name of the worktree directory the session runs in; without the
field no card arms.

- **Pattern** `^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`. Why this shape: the value is compared with a worktree
  directory name, so it must be exactly one path segment — no `/`, and no `.` so that `.` and `..` cannot
  occur; the first character is alphanumeric so the value cannot be read as a flag; 64 characters is the
  same bound the resource-lease names already use. No project-specific prefix is hard-coded: any card id
  scheme that fits one path segment is accepted.
- **Reason code** `card_invalid` for a missing, empty, or non-matching value.
- **Integrity chain.** `card` is part of the contract body, so it is included in the canonical digest
  (§ Digest); the digest is recorded as `signature.contract_sha256`; the seal covers that recorded value
  (§ Signature Seal). Editing `card` after signing therefore makes the recomputed digest differ
  (`contract_digest_mismatch`), and also editing `signature.contract_sha256` to match without re-sealing
  breaks the seal (`signature_seal_mismatch`).
- `show --json` reports it as `card`.

### Glob Semantics

Ownership globs use doublestar semantics over forward-slash, repo-relative paths: `*` matches within
one path segment, `**` matches zero or more whole segments, `?` matches one non-separator character.
`go.mod` carries no doublestar library, so the matcher is implemented in-package on top of the standard
library's per-segment `path.Match` (no new dependency).
"Covers the SPEC directory" (REQ-CONTRACT-020) means: at least one `write` glob matches the literal
path `.moai/specs/<SPEC-ID>/contract.yaml` under these semantics — so `.moai/**`,
`.moai/specs/**`, and `.moai/specs/<SPEC-ID>/**` all cover it, `.moai/specs/*.md` does not. The
`never`/`write` and `never`/`scratch` overlap rules are literal string comparisons only; semantic
overlap between different glob strings is A2's ownership detector's concern.

### Derived Fields

Computed by the verifier and reported by `show` / `show --json` (REQ-CONTRACT-014, 018, 025):

| Field | Definition |
|---|---|
| `effective_never` | `ownership.never` ∪ {`.moai/specs/<ID>/contract.yaml`, `.moai/specs/<ID>/acceptance.md`} when the contract is signed; `ownership.never` when unsigned. The two files are immutable during a run even when a `write` glob matches them; A2 enforces. |
| `scratch` | `ownership.scratch` (empty list when absent). |
| `frozen_files` | § Frozen Files, when the `frozen-files` invariant is present; empty list otherwise. Every element is a glob under § Glob Semantics. |
| `push_requires_lease` | `true` when `actions` contains `push-develop`: the push must hold a `moai slot` lease on resource `push-develop` (§ Push Lease). |
| `terminal` | `true` when `spec.md` frontmatter `status` (quoted or unquoted) is `completed` or `archived`. A terminal contract is excluded from detection and signature resolution (A2); it is reported, not treated as invalid. |
| `signable_contract_sha256` | Digest of the contract as `sign` would sign it now (acceptance binding measured, budget filled); the value a kickoff receipt's `inputs.contract_sha256` must carry. |

### Frozen Files

The `frozen-files` invariant token denotes the sorted union of three sources (B3), each measured in this
tree at plan time. **Every element of the union is a glob under § Glob Semantics**, so A2 applies one
matching rule to the whole set:

1. **Charter zone registry, Frozen entries** — the distinct `file:` values of entries with
   `zone: Frozen` in `.claude/rules/moai/core/zone-registry.md` (first such entry at line 83; 57
   entries over 12 distinct files at HEAD `ca1d5dc43`), each emitted as a literal repo-relative path
   (a glob with no wildcard, matching exactly that path). Resolved by the caller through the
   constitution registry loader and passed to the verifier.
2. **Frozen instruction files** — `frozenInstructionFiles` at `internal/hook/pre_tool.go:1231`
   (`{"CLAUDE.md", "CLAUDE.local.md"}`) is a **basename** list (`checkHarnessFrozenZone`,
   `pre_tool.go:1242-1246`), emitted as `**/CLAUDE.md` and `**/CLAUDE.local.md` so the glob keeps the
   basename-anywhere semantics (`**` matches zero segments, so the root files are covered). The
   verification core carries its own copy (it must not import `internal/hook`); a test inside
   `internal/hook` pins the copy to the variable.
3. **The contract's `ownership.never`** — already globs, emitted unchanged.

Not included: `frozenZonePrefixes` in the same file (`.claude/commands/`, `.claude/hooks/`,
`.claude/output-styles/`), which the hook applies only to the harness-learner identity. A2 may widen the
set by amending this section.

### Push Lease

Lead decision (2026-09-26): `push-develop` is serialized with a `moai slot` lease, not with the
integration (merge) window — the push happens outside the merge window and is a different resource.
Measured in this tree: `moai slot` has the verbs `acquire | status | release` (`internal/cli/slot.go:86`,
`:111`, `:243`, `:298`); a resource is named with `--resource` (`slot.go:163`), validated by
`^[a-z0-9-]{1,64}$` (`internal/kanban/slot_lease.go:63`); the lease record lives in the primary
checkout's `.moai/state/slot-leases`, visible from every linked worktree; exit code 3 means the resource
is held by another session. The resource name for a contract push is the fixed string `push-develop`
(valid under that pattern). A1 only declares the requirement (`push_requires_lease`); A2b (t1245)
enforces that a push of the integration branch happens while this session holds the lease.

### Action Vocabulary

| Contract token | Status | Mission action (forward reference for A2b, t1245) |
|---|---|---|
| `commit` | allowed | `commit` |
| `worktree` | allowed | (no mission equivalent — contract-only) |
| `local-merge-develop` | allowed | `local_develop_merge` |
| `push-develop` | allowed when `push_develop: true`; implies `push_requires_lease` | `batch_push` |
| `push-main` | **forbidden** | — |
| `merge-main` | **forbidden** | `main_merge` |
| `force-push` | **forbidden** | `force_push` |
| `release-branch` | **forbidden** | `release_branch` |
| `release-pr` | **forbidden** | `release_pr` |
| anything else | unknown | — |

The forbidden set is fixed by the schema, not by configuration. An empty list is `actions_empty`.
Mission-mappability is not required in A1; A2b (t1245) decides whether its projection requires at least one
mappable action.

### Digest

1. Strict-decode the YAML into the typed contract.
2. Drop the `signature` block.
3. Canonicalize: sort every set-valued list (`invariants`, `ownership.write`, `ownership.never`,
   `ownership.scratch`, `actions`, `escalate_on`) lexicographically; keep `reobserve` in author order.
4. Marshal to JSON with the struct's fixed field order; SHA-256; lowercase hex.

This is the same technique as the mission package's contract sealing (clone, sort, marshal, hash),
applied to the SPEC contract type.

### Acceptance hash and count

- Normalize: strip a leading UTF-8 BOM; replace every CRLF with LF.
- Hash: SHA-256 of the normalized bytes; lowercase hex.
- Count: a Go port of the published AC counter (the `MOAI-AC-COUNTER` block in the manager-docs agent
  definition), applied to the **normalized** content: honor the `moai-ac-prefix` declaration, count
  distinct live IDs, exclude IDs marked `[RETIRED]`/`[REF]`, report ambiguity when one ID is both
  marked and unmarked.
- Parity: a test runs both the extracted awk counter (on the same normalized bytes) and the port over
  (a) every `.moai/specs/*/acceptance.md` and (b) synthetic fixtures, one per counter branch: prefix
  declaration, `[RETIRED]` marker, `[REF]` marker, ambiguity, CRLF line endings, UTF-8 BOM. A positive
  control asserts the ambiguity fixture is ambiguous in **both** implementations. The same test also
  runs the awk counter on the CRLF fixture's **raw** bytes and records whether that count differs from
  the normalized one (spec.md §C.5 scope statement).

### Verify Reason Codes

Closed set, emitted in `show --json` / `verify --json` as `reasons: [...]` (sorted, de-duplicated):

`schema_invalid`, `spec_id_mismatch`, `unsigned`, `contract_digest_mismatch`, `acceptance_missing`,
`acceptance_hash_mismatch`, `ac_count_mismatch`, `ac_count_ambiguous`, `actions_empty`,
`unknown_action`, `forbidden_action`, `push_develop_disabled`, `second_review_missing`,
`escalate_on_incomplete`, `ownership_invalid`, `invariant_unresolved`, `budget_invalid`,
`reobserve_incomplete`, `plan_audit_not_passing`, `receipt_mismatch`, `card_invalid`, `signature_seal_mismatch`,
`signature_inconsistent`, `signature_acceptance_mismatch`.

Verify collects every applicable code rather than stopping at the first, except that a
`schema_invalid` decode failure stops evaluation.

### Sign Refusal Codes

Closed set; `sign` prints the code and a one-line cause and exits 1 (REQ-CONTRACT-012, 013, 021–024):

| Code | Cause |
|---|---|
| `agent_marker` | human path, agent-environment marker present (the variable is named) |
| `not_tty` | human path, stdin is not a terminal |
| `confirmation_mismatch` | human path, typed token differs |
| `already_signed` | signed contract without `--resign` |
| `not_signed` | `--resign` on an unsigned contract |
| `draft_acceptance_stale` | unsigned draft records an acceptance hash different from the measured one |
| `ac_count_ambiguous` / `ac_count_zero` | counter result unusable |
| `plan_audit_not_passing` | verdict not `PASS`/`PASS-WITH-DEBT` |
| `git_identity_missing` | `user.name` or `user.email` empty |
| `batch_disabled` | more than one distinct ID with `batch_sign: false` |
| `batch_non_human` | more than one distinct ID on the receipt path |
| `mode_not_contract` | receipt path while `mode` is not `contract` |
| `kickoff_decider_jev_sole` | receipt path while the configured `kickoff.decider` is `jev` (a configuration error: Jev is never a sole decider) |
| `receipt_invalid` | a structural or consistency rule fails (§ Kickoff Receipt, rules 1, 2, 5, 6, 7, 9) |
| `receipt_signer_mismatch` | the decider fields do not match the configuration or `--signer`, or a fallback is unrecorded (§ Kickoff Receipt, rules 3-4) |
| `receipt_input_mismatch` | an input hash differs from the current file / signable digest |
| `receipt_requires_human` | the recorded `outcome` is `human`, or — while A3's principle amendment has not landed — the effective decider is `llm+jev` (interim A1 rule, removed by A3 when it lands) |
| `receipt_rejected` | the recorded `outcome` is `reject` |
| `verify_failed` | any other verify rule fails (the verify codes are printed) |

### `show --json` / `verify --json` object (stable field names)

```json
{
  "spec_id": "SPEC-EXAMPLE-001",
  "card": "t1234",
  "schema_version": 1,
  "state": "signed-valid",
  "valid": true,
  "reasons": [],
  "contract_sha256": "…",
  "recorded_contract_sha256": "…",
  "signable_contract_sha256": "…",
  "acceptance": { "sha256": "…", "measured_sha256": "…", "ac_count": 9, "measured_ac_count": 9 },
  "actions": ["commit", "local-merge-develop", "push-develop", "worktree"],
  "push_requires_lease": true,
  "terminal": false,
  "effective_never": ["…"],
  "scratch": ["…"],
  "frozen_files": ["…"],
  "second_review": "required",
  "mode": "guided",
  "budget": { "turns": 60, "operations": 40, "audit_retries": 2 },
  "signature": { "signer_kind": "human", "operator": {"name": "…", "email": "…"}, "signed_at": "…", "head_sha": "…", "method": "interactive-tty", "receipt": null, "batch_id": "", "supersedes": "" }
}
```

`state` ∈ `unsigned | signed-valid | signed-invalid`. `show` prints the full contract sections in
addition; `verify --json` prints this object only.

### Exit codes

| Command | 0 | 1 | 2 |
|---|---|---|---|
| `verify` | valid | invalid (reasons printed) | usage / I/O error (e.g. contract file absent, path escapes the SPEC directory) |
| `show` | printed (valid or not) | — | usage / I/O error |
| `sign` | signed | refused (a code from § Sign Refusal Codes) | usage / I/O error |

## Kickoff Receipt

`.moai/specs/<SPEC-ID>/kickoff-receipt.json` (fixed name; `--receipt` must resolve to it). Strictly
decoded JSON. A1 owns the fields and a validator that checks **structure and internal consistency
only**; the rules that decide the outcome belong to A3 (spec.md §C.8).

```json
{
  "receipt_version": 1,
  "spec_id": "SPEC-EXAMPLE-001",
  "requested_decider": "llm+jev",
  "effective_decider": "llm+jev",
  "fallback": { "applied": false },
  "issued_at": "2026-09-26T09:00:00Z",
  "inputs": {
    "contract_sha256": "<signable_contract_sha256>",
    "acceptance_sha256": "<acceptance.md hash>",
    "plan_audit_report": { "path": ".moai/reports/plan-audit/SPEC-EXAMPLE-001-review-2.md", "sha256": "<raw bytes>" }
  },
  "llm_answer": { "answer": "approve", "confidence": 0.82,
                  "reason": "Scope and actions match plan.md; no new API.",
                  "reason_refs": ["contract.yaml:12", "contract.yaml:27"] },
  "jev_answer": { "answer": "approve", "confidence": 0.71, "reason": "…", "reason_refs": ["contract.yaml:27"],
                  "request_sha256": "<hex64>", "raw_response": "<verbatim response body>" },
  "outcome": "approve"
}
```

A fallback receipt instead carries `"effective_decider": "llm"`,
`"fallback": { "applied": true, "reason": "jev_disabled" }`, and no `jev_answer`.

### Field rules (A1 validator — structure and internal consistency)

| # | Rule | Refusal code |
|---|---|---|
| 1 | Strict decode; `receipt_version == 1`; `spec_id` equals the SPEC directory; path is the fixed `kickoff-receipt.json` | `receipt_invalid` |
| 2 | `requested_decider`, `effective_decider` ∈ `llm | llm+jev` | `receipt_invalid` |
| 3 | `requested_decider` equals the effective `kickoff.decider`; `effective_decider` equals `--signer` | `receipt_signer_mismatch` |
| 4 | `fallback.applied` is true exactly when `effective_decider` ≠ `requested_decider`, and the only permitted difference is `llm+jev` → `llm` (a difference without `applied: true` is a silent fallback) | `receipt_signer_mismatch` |
| 5 | `fallback.reason` is present exactly when `fallback.applied`, and is one of `jev_disabled`, `jev_low_confidence`, `jev_malformed_response`, `jev_call_failed`, `jev_key_missing` | `receipt_invalid` |
| 6 | `llm_answer` is always present; `jev_answer` is present exactly when `effective_decider` is `llm+jev` | `receipt_invalid` |
| 7 | Each answer: `answer` ∈ `approve | reject | escalate`; `confidence` ∈ [0, 1]; `reason` non-blank; `reason_refs` non-empty, each `contract.yaml:<line>` with `1 <= line <=` the draft's line count; `jev_answer` also carries non-empty `raw_response` and a hex64 `request_sha256` | `receipt_invalid` |
| 8 | `inputs.contract_sha256` equals `show --json`'s `signable_contract_sha256`; `inputs.acceptance_sha256` and `inputs.plan_audit_report.sha256` equal the current files' SHA-256 | `receipt_input_mismatch` |
| 9 | `outcome` ∈ `approve | reject | human`; `outcome: approve` requires `llm_answer.answer: approve` (a Jev approval never suffices alone) | `receipt_invalid` |

After the validator accepts a receipt, the signer applies, in order: (1) **interim A1 rule** — while
A3's amendment of the Jev display-only principle has not landed, an `effective_decider` of `llm+jev` (Jev
actually answered) → refuse `receipt_requires_human`, even when both answers approve; once A3 lands its
amendment, A3 removes this step and an `llm+jev` receipt follows A3's cross-check result; (2) the recorded
outcome: `approve` → sign; `reject` → refuse `receipt_rejected`; `human` → refuse
`receipt_requires_human`. A recorded fallback (requested `llm+jev` → effective `llm`) passes step 1. A1 does not check that
the outcome follows the cross-check rules below — that consistency is A3's to enforce, together with
moai-issued receipts (spec.md §C.6).

### Cross-check Rules (A3-owned, informative)

Recorded here so A3 starts from the operator's decision; A1 does not evaluate them.

- `llm`: the LLM answer decides — `approve` → `approve`, `reject` → `reject`, `escalate` → `human`.
- `llm+jev`: both `approve` → `approve`; any disagreement or `escalate` → `human`; both `reject` →
  `reject`. A Jev approval never suffices without an LLM approval.
- Jev-side failure under `llm+jev` (`jev_disabled`, `jev_low_confidence` — a Jev confidence below
  `kickoff.jev_min_confidence` —, `jev_malformed_response`, `jev_call_failed`, `jev_key_missing`) →
  fall back to `llm` alone, recorded as requested `llm+jev` → effective `llm` with the reason.
- Interim rule until A3 amends the shipped "Jev is display-only" principle: an `llm+jev` receipt in
  which Jev answered → `human`. **While A3's amendment has not landed, A1 enforces this one** (signer
  step 1 above); A3 removes that enforcement when its amendment lands.

## Agent-Environment Markers

Closed set checked by `sign` on the **human** path before anything else (REQ-CONTRACT-021). A marker is
"present" when the variable is set to a non-empty value. The receipt path does not check markers — it
is the path meant for automated deciders, gated instead by `mode: contract`, the receipt validator,
and §C.6 of spec.md.

| Variable | Source of the claim |
|---|---|
| `CLAUDECODE` | Measured in this plan session: present in the Bash tool environment of Claude Code |
| `CLAUDE_CODE_SESSION_ID` | `internal/config/envkeys.go:494` `EnvClaudeCodeSessionID` — stamped into every subprocess Claude Code spawns |

Codex: no Codex session-marker constant exists in this repository (`CODEX_HOME` appears, but it is a
user configuration location that users also set in ordinary shells). A1 ships no Codex marker;
run-phase may add one only after measuring, from a Codex exec'd command, a variable that Codex sets and
ordinary shells do not. The A2b (t1245) sign deny (spec.md §C.2) is the Codex-side protection.

## Signing Flow

Path selection: `--signer` absent or `human` → human path; `llm | llm+jev` → receipt path.

1. Human path: refuse `agent_marker`, then `not_tty`. Receipt path: refuse `batch_non_human` for more
   than one ID, `mode_not_contract` unless `mode: contract`, and `kickoff_decider_jev_sole` when the
   configured decider is `jev`. (`--signer` names the receipt's effective decider, so a recorded
   `llm+jev` → `llm` fallback is signed with `--signer llm`.)
2. Resolve each SPEC ID (de-duplicated) to its directory; load and strictly decode `contract.yaml`
   (a symlink resolving outside the SPEC directory is an I/O error, exit 2).
3. Branch on signature state:
   - **Unsigned, no `--resign`**: measure acceptance hash and count; refuse `ac_count_ambiguous` /
     `ac_count_zero`; refuse `draft_acceptance_stale` when a recorded hash differs.
   - **Unsigned, `--resign`**: refuse `not_signed`.
   - **Signed, no `--resign`**: refuse `already_signed`.
   - **Signed, `--resign`** (REQ-CONTRACT-022): re-measure; refuse on ambiguity or zero; keep the
     recorded values and the plan-audit verdict for the summary; set `supersedes` to the current
     `signature.contract_sha256`.
4. Fill `budget` from `workflow.autonomy.escalation.budget_default` when absent.
5. Run verify on the would-be contract; refuse `verify_failed` on any reason other than `unsigned`
   (and, on `--resign`, the acceptance-binding codes being replaced).
6. Read operator identity from `git config user.name` / `user.email` (refuse `git_identity_missing`)
   and HEAD.
7. Human path: print the summary (per SPEC: ID, acceptance hash prefix and AC count — as `old → new`
   on re-sign — actions, budget, second model, plan-audit verdict marked "carried" on re-sign) and the
   confirmation token (the SPEC ID, or `sign N contracts` in batch mode); read one line; refuse
   `confirmation_mismatch`. Receipt path: validate the receipt (REQ-CONTRACT-023) against the signable
   digest and current files; refuse with its code. No prompt.
8. Compute the digest; write the signature block (`signer_kind`, `method`, and for the receipt path
   `receipt.{path, sha256, provenance: file}`), then compute and write `seal` last; write each file
   atomically; print the
   REQ-CONTRACT-019 notice (both modes, both paths).

Batch mode (human path only): steps 2-6 run for every ID before step 7; one confirmation; step 8 per
file with a shared `batch_id`. A write failure after some files were written is reported with the
exact written/unwritten split (atomic per file, not across files).

## Forward Note — Mission Projection (for A2b, t1245)

Deferred from A1 (spec.md §B Out of Scope — Mission-validator projection). Drafted mapping, for A2b to
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
contains `internal/foo/x.go`. A2b must translate trailing `/**` to a prefix and decide how to treat
globs with inner wildcards, and must require at least one mission-mappable action (the mission sealer
rejects an empty `AllowedActions` as `incomplete_contract`).

## Forward Note — Verdict Binding (for A3 / A4)

`plan_audit.verdict` is self-reported on the human path and carried unchanged by `--resign`
(spec.md §C.7). A3/A4 shall bind it to evidence — for example by requiring the receipt path's
`inputs.plan_audit_report` hash, or an audit record keyed to the current acceptance hash — before a
verdict gates anything.

## Configuration

```yaml
workflow:
  autonomy:
    mode: guided              # guided | contract
    contract:
      batch_sign: false       # sign several SPECs with one confirmation
      second_review: required # required | advisory | off
      push_develop: false     # allow the push-develop action in contracts
    kickoff:
      # decider: human | llm | llm+jev   (omitted: human in guided, llm in contract)
      jev_min_confidence: 0.50
    escalation:
      budget_default: { turns: 60, operations: 40, audit_retries: 2 }
```

This block is the template text; its comments are neutral by rule (REQ-CONTRACT-016): no `A-Q<n>`
identifiers and no statement about this repository's local values. Rationale lives here, outside the
block: `mode: guided` and `second_review: required` are operator decisions A-Q4 and A-Q3;
`batch_sign`/`push_develop` default to `false` because a user project may have no integration branch.
This repository's local `.moai/config/sections/workflow.yaml` sets `batch_sign: true` and
`push_develop: true` (asserted by AC-CONTRACT-020).

Reader behavior: absent section/key → defaults above, except `decider`: when it is absent or empty the
effective decider is derived from the effective mode — `human` under `guided`, `llm` under `contract`.
An explicit value always wins. The template omits the key (REQ-CONTRACT-016), so a project initialized
from it gets `human` under `guided` and `llm` after switching to `mode: contract`. An out-of-set or
out-of-range value for `mode`, `second_review`, `decider` (other than `jev`), or `jev_min_confidence` →
`guided`, `required`, `human`, `0.50` respectively, plus a warning naming the key (fail toward the
stricter value). `kickoff.decider: jev` is a configuration error naming the key (Jev is never a sole
decider): loading still succeeds, no fallback is applied, no command is blocked, and only the receipt
signing path is refused with `kickoff_decider_jev_sole`. `llm+jev` while `workflow.jev.enabled: false`
(the shipped key) loads without error; the Jev-side fallback to `llm` is A3's decision and is recorded in
the receipt with reason `jev_disabled`. No
environment variable is introduced; `MOAI_AUTONOMY_TIER` is untouched.

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
  normalization/hash, AC counter port, glob matcher, all field rules, derived fields, receipt decode and
  validator, verify. Imports the standard library (no `os/exec`, no `net`) and `gopkg.in/yaml.v3`
  (measured: zero `os/exec`/`net` in its deps). It does **not** import `internal/config`,
  `internal/constitution`, `internal/spec`, or `internal/hook`; policy values, registry rule IDs, and
  registry Frozen files are passed in by the caller. Its SPEC ID pattern is pinned by a test in
  `internal/spec`; its frozen-instruction-file copy is pinned by a test in `internal/hook`.
- `internal/contract/sign/` — signing: marker and TTY checks, git identity/HEAD (subprocess),
  confirmation, receipt-path orchestration, atomic writes via `internal/atomicfile`.
- `internal/cli/contract.go` — Cobra `contract` command with `sign`, `show`, `verify`; loads config and
  the constitution registry and passes values into the core.
- `internal/config/` — `AutonomyConfig` under the workflow section, defaults, reader fail-safe.
- Parity test — lives in `internal/spec` (package `spec`, beside the unexported
  `extractCounterCommand`) and imports `internal/contract`; no cycle because `internal/contract` never
  imports `internal/spec`.
- `internal/template/templates/.moai/config/sections/workflow.yaml` and the local
  `.moai/config/sections/workflow.yaml` — the `autonomy` block.
