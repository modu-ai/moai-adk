---
id: SPEC-AUTONOMY-PRECONDITION-001
title: "A3 preconditions: serialize develop pushes behind the push-develop slot lease, deny agent-invoked contract signing, and project the contract onto the mission validator"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/hook"
lifecycle: spec-anchored
tags: "autonomy,contract,precondition,push-serialization,sign-guard,hook,mission"
tier: M
related_specs: [SPEC-AUTONOMY-CONTRACT-001, SPEC-AUTONOMY-ESCALATION-001, SPEC-AUTONOMY-TIERS-001]
---

# SPEC-AUTONOMY-PRECONDITION-001 — A3 preconditions (card t1245, track A2b)

## §A — History

- **2026-09-26** — v0.1.0 plan-phase draft. Card **t1245**, contract-based autonomous harness
  track **A2b**. Base tree `develop` at `553e224f3`; worktree `.claude/worktrees/t1245`, branch
  `WT-push-serialize-sign`. This SPEC exists because lead ruling 09-26 (2) #3 split three items
  out of SPEC-AUTONOMY-ESCALATION-001 (card t1235, track A2): the push serializer, the
  contract-sign guard, and the mission-validator projection. The transfer record is that SPEC's
  §K at `WT-escalation-detector`; the transferred text was read from `WT-escalation-detector~1`
  (the commit before the split removed it), through committed refs only — never through the
  t1235 worktree, which another session is writing.
- **2026-09-26** — every premise handed to this SPEC in the dispatch was measured against
  `553e224f3` before any requirement was written. The measured table (premise → exists? →
  evidence) is §C. Two dispatch premises did **not** hold and the requirements below say so
  rather than assuming them (`moai contract sign` does not exist yet; `MOAI_FACTORY_ROLE` does
  not exist at all).
- **2026-09-26** — A1's open items **R5** (receipt path) and **R6** (which window
  `push_requires_window` means) are both **answered** by SPEC-AUTONOMY-CONTRACT-001 v0.4.0/v0.4.1
  at `WT-contract-schema`, so no criterion here is conditional on unowned work. Citations: §C.3,
  §C.4. This is the defect (finding **N5**) that caused the split; it is closed rather than
  carried.

## §B — Problem

Track A3 rewires gates for contract-mode autonomy. It cannot start until two denials exist and
one reuse is settled, and none of the three had an owner:

1. **Back-to-back `develop` pushes cancel CI.** When two lanes push the integration branch in
   quick succession, the earlier push's in-flight CI run is cancelled and its verdict never
   arrives. The lead reads CI as the integration verdict
   (`CLAUDE.local.md` §4.1 규율 5), so a cancelled run is a verdict silently replaced by nothing —
   the most expensive failure shape, because it presents as a pass.
2. **A contract signed by the agent it binds is no contract.** A1 refuses agent-environment
   signing on its human path (REQ-CONTRACT-021) and refuses a non-terminal stdin, but both
   defences live inside the binary being invoked. A guard at the tool-call boundary is defence in
   depth against the shapes A1's own process never sees.
3. **The mission validator already implements fail-closed decision validation.** `internal/mission`
   carries `ValidateMissionDecision` (`internal/mission/policy.go:200`), which treats every field
   as untrusted and returns a prepared receipt only after complete validation. Re-implementing
   that logic for contracts would be a second, divergent copy of the same fail-closed rules.

## §C — Measured premises (the dispatch's premises, verified against `553e224f3`)

Each row was measured in this tree at `553e224f3`. A premise that did not hold is recorded as
absent and the dependent requirement names who creates the surface.

### C.1 `moai slot` and the `push-develop` lease — **EXISTS (the resource does not yet)**

| What | Evidence |
|---|---|
| `moai slot` command with `acquire` / `status` / `release` | `internal/cli/slot.go:87` (`Short: "Acquire, inspect, and release a named heavy-resource slot (cross-session lease)"`) |
| Named resources validated, not enumerated | `internal/cli/slot.go:114` `kanban.ValidateSlotResourceName(resource)`; flag help at `:163` — `a-z, 0-9, '-', 1-64 characters` |
| Lease record read/write | `kanban.ReadSlotLease(root, resource)` (`internal/cli/slot.go:250`) |
| Prior-art PreToolUse guard on the same record | `internal/hook/slot_lease_guard.go` |
| A record for the resource `push-develop` | **absent** — no file, no test, no mention. `push-develop` is a *valid* resource name, so this SPEC creates the first use of it, not a new mechanism. |

### C.2 `moai contract sign` — **DOES NOT EXIST in this tree; A1 creates it**

- `go run ./cmd/moai contract --help` → `Unknown command "contract" for "moai".`, exit 1.
- `internal/cli` carries no `contract.go`; the `contract` string appears only in
  `codex_init_test.go` as an unrelated call-order token.
- A1 (SPEC-AUTONOMY-CONTRACT-001, card t1234, branch `WT-contract-schema`) is building it: its M1
  commit `b93af204a` adds `internal/contract/` (schema, 23 verify reason codes, `Decode`/`Digest`/
  `Verify`/`LoadDir`/`ResolveSpecDir` stubs) — package only, no CLI yet.
- **Consequence for this SPEC**: the guard is specified against the *command line shape*, which is
  observable at PreToolUse whether or not the binary implements the verb. A guard that denies an
  invocation of a not-yet-existing verb is correct and is not vacuous: the denial is what makes the
  verb agent-unreachable from the moment A1 lands it. REQ-AP-004 states this explicitly.

### C.3 `moai contract decide` — **DOES NOT EXIST and is not created by A1**

- No occurrence of `contract decide` or a `decide` verb anywhere in this tree.
- A1's verb set is three read/sign commands — `show`, `verify`, `sign`
  (`WT-contract-schema:.moai/specs/SPEC-AUTONOMY-CONTRACT-001/spec.md`, exit-code table
  `design.md` § Exit Codes row `sign`); `moai contract revoke` is **declared out of scope, owned by
  A3** (that SPEC's v0.3.0 HISTORY row).
- **Consequence**: the dispatch named `decide` alongside `sign`. There is no such surface to guard
  and no owner creating one. Guarding a verb nobody defines would be an unobserved premise, so
  `decide` is **out of scope** here (§G) and recorded as a clarification (§F O1) rather than
  silently dropped.

### C.4 R5 — the moai-issued receipt path — **ANSWERED by A1**

`WT-contract-schema:.moai/specs/SPEC-AUTONOMY-CONTRACT-001/design.md` § Kickoff Receipt:

> `.moai/specs/<SPEC-ID>/kickoff-receipt.json` (fixed name; `--receipt` must resolve to it).
> Strictly decoded JSON

and § Signing Flow (same file): path selection is `--signer` absent or `human` → human path;
`llm | jev | llm+jev` → receipt path. A1 REQ-CONTRACT-024 records the receipt signature's
provenance as `file`. Critically, A1 §C.8 states that the receipt validator treats **any** Jev
decision as unmeasured and routes it to a human (`receipt_requires_human`, REQ-CONTRACT-023), so
`jev` and `llm+jev` receipts **cannot sign in A1**. R5 is therefore answered with a *shape*
(fixed filename, `--receipt`, `--signer`) and a *fact* (the path is defined but currently
unusable). REQ-AP-006 is written to the shape, which makes it evaluable today.

### C.5 R6 — which window `push_requires_window` means — **ANSWERED by A1, and renamed**

`WT-contract-schema:.moai/specs/SPEC-AUTONOMY-CONTRACT-001/spec.md` v0.4.0 HISTORY row:

> push serialization re-specified as a `moai slot` lease on resource `push-develop`
> (`push_requires_window` renamed `push_requires_lease`)

and its requirement body (same file, around `:362`): the `push-develop` action denotes a push of
the integration branch performed while holding a `moai slot` lease on the resource
`push-develop`, surfaced as `push_requires_lease: true` in `show --json` **for A2b to enforce**.
This SPEC therefore reads `push_requires_lease`, never `push_requires_window`; the old name
appears nowhere in its requirements.

### C.6 `MOAI_FACTORY_ROLE` — **DOES NOT EXIST**

- `grep -rn "MOAI_FACTORY_ROLE" .` → no hit outside prose. `internal/config/envkeys.go` carries
  `EnvMoaiFactoryWorkers = "MOAI_FACTORY_WORKERS"` (`:280`, set on the lead **and** every worker)
  and `EnvMoaiFactoryWorker = "MOAI_FACTORY_WORKER"` (`:287`, worker only, value `lane-<n>`).
  `isFactoryRoleToken` (`internal/cli/codex_factory.go:103`) parses a **CLI argument**, not an
  environment variable.
- A1 already owns the marker set this SPEC must key on: `design.md` § Agent-Environment Markers
  lists `CLAUDECODE` (measured present in Claude Code's Bash environment) and
  `CLAUDE_CODE_SESSION_ID` (`internal/config/envkeys.go:494`), and records that **no Codex marker
  exists in this repository**, naming the A2-side sign deny — this SPEC — as the Codex-side
  protection.
- **Consequence**: the dispatch's framing ("deny when `MOAI_FACTORY_ROLE` is `agent`") rests on a
  variable that does not exist. REQ-AP-003 is written instead against the **tool-call boundary
  itself** (every Bash tool call is by construction agent-issued, which is stronger than any
  marker) and REQ-AP-009 reuses A1's closed marker set where a marker is genuinely needed. The
  original framing is recorded as clarification O2.

### C.7 `mission.ValidateMissionDecision` — **EXISTS at the cited line**

`internal/mission/policy.go:200`:

> `func ValidateMissionDecision(sealed SealedContract, snapshot MissionSnapshot, decision Decision, now time.Time) (OperationReceipt, error)`

with the doc comment "treats every Decision field as untrusted data. It never calls a tool or
performs a state change". Its refusal ladder (`mission_not_running`, `mission_mismatch`,
`policy_mismatch`, `stale_snapshot`, `expired_decision`) is the reuse target of REQ-AP-008.

### C.8 Command-parsing helpers — **EXIST at the cited lines**

`substituteQuotedArguments` (`internal/hook/branch_guard.go:197`) and `extractIntegrationCommand`
(`internal/hook/integration_lock_guard.go:122`). §D notes where quote *removal* is needed instead
of quoted-span *scrubbing*, because scrubbing would erase `'moai'` and defeat the matcher.

## §D — Requirements (GEARS)

### D.1 Push serializer

- **REQ-AP-001** (Capability gate + event) — Where `workflow.autonomy.mode` is `contract` and the
  resolved contract's `actions` contains `push-develop` with `push_requires_lease: true`, when a
  Bash tool call is a push of `develop`, the push serializer shall admit it only if the
  `push-develop` slot lease is free, expired, stale, or held by the calling session — acquiring the
  lease for the calling session on admission — and shall otherwise deny it with a reason prefixed
  `PUSH_SERIALIZATION_VIOLATION:` naming the live holder, as a wait rather than an escalation.
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AP-002** (Event-driven) — When an admitted push exits non-zero, the serializer shall
  release the lease; when the lease record cannot be read, the root cannot be resolved, or the
  caller is unknown, it shall allow the call and append one audit line (fail open).
  「A1 plan-audit 통과본으로 재확인」
- **REQ-AP-007** (Ubiquitous) — The serializer shall write no escalation record on any path, and
  shall leave the hook output of an admitted push unchanged.

### D.2 Contract-sign guard

- **REQ-AP-003** (Ubiquitous) — The contract-sign guard shall deny, regardless of
  `workflow.autonomy.mode`, every Bash tool call that invokes `moai contract sign`, with a reason
  prefixed `CONTRACT_SIGN_AGENT_VIOLATION:`, recognizing the invocation through shell quoting,
  leading `NAME=value` assignments, the prefix wrappers of REQ-AP-005, any path to an executable
  whose basename is `moai`, global flags preceding the verb, and one level of `sh -c` / `bash -c` /
  `zsh -c`. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AP-004** (Unwanted) — When a command carries the words `contract` and `sign` but its
  structure cannot be classified — command substitution, a variable in program position, `eval`,
  or nesting deeper than one `-c` level — the guard shall not allow it, and shall deny it with the
  reason marked unclassified (fail closed). The guard shall deny such a command whether or not the
  `moai contract sign` verb is implemented in the installed binary (§C.2).
- **REQ-AP-005** (Ubiquitous; finding N4) — The guard's prefix-wrapper set shall be a closed,
  enumerated list containing at least `env`, `command`, `exec`, `nohup`, `script`, `timeout`,
  `sudo`, `stdbuf`, `nice`, and `xargs`, and shall skip each wrapper's own options before reading
  the program word; a wrapper not in the list, encountered in program position, shall be treated
  as unclassifiable under REQ-AP-004 rather than allowed.
- **REQ-AP-006** (Where; R5) — Where a `moai contract sign` invocation selects A1's receipt path —
  `--signer` naming `llm`, `jev`, or `llm+jev`, with `--receipt` resolving to
  `.moai/specs/<SPEC-ID>/kickoff-receipt.json` — the guard shall allow the invocation and record
  one audit line naming the receipt path; it shall deny an invocation that names the receipt path
  while the file is absent, and shall not itself validate the receipt's contents, which A1's
  validator owns (REQ-CONTRACT-023). 「A1 plan-audit 통과본으로 재확인」
- **REQ-AP-009** (Ubiquitous; §C.6) — Where the guard needs an agent-environment signal for a
  reason other than the tool-call boundary, it shall read A1's closed marker set
  (`design.md` § Agent-Environment Markers) and shall not read `MOAI_FACTORY_ROLE`, which does not
  exist, nor introduce a marker of its own. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AP-010** (Ubiquitous; finding N8) — The guard shall be documented, in the template
  configuration comment and in the rule text that states contract-mode behavior, as a
  mode-independent deny that is active under `guided`, so that the "nothing changes under
  `guided`" promise of SPEC-AUTONOMY-ESCALATION-001 REQ-AE-001 is stated as scoped to the
  escalation detector rather than to contract-mode work as a whole.

### D.3 Mission-validator projection

- **REQ-AP-008** (Event-driven; O9 / N7) — When a signed contract must be checked against a
  proposed action, the projection shall map the contract onto `mission.SealedContract` /
  `mission.MissionSnapshot` / `mission.Decision` and obtain the verdict from
  `mission.ValidateMissionDecision` (`internal/mission/policy.go:200`) rather than re-implementing
  its refusal ladder; where a contract field has no mission counterpart, the projection shall fail
  closed and name the unmapped field, and shall not widen the mission types to accommodate it.

## §E — Constraints

- **C1 — No new lease mechanism.** REQ-AP-001 uses the existing `moai slot` record, resource name
  `push-develop`, its liveness, bound, and takeover rules. No new record format, lock, or verb
  (§C.1). It does not depend on `workflow.slot_lease.enabled`, which keeps gating the generic slot
  guard only.
- **C2 — Opposite fail directions, deliberately.** REQ-AP-002 fails **open** and REQ-AP-004 fails
  **closed**. An overlapping push costs one cancelled CI run; a stuck deny halts the lane. A
  wrongly allowed signature voids the contract model; a wrongly denied sign costs the operator one
  terminal command, which is where signing belongs.
- **C3 — Neither component writes an escalation record.** They deny, which the escalation detector
  never does (SPEC-AUTONOMY-ESCALATION-001 REQ-AE-003). They share only the contract resolver
  (REQ-AE-002).
- **C4 — Activation ordering is A1's, not this SPEC's.** A1 states that the `push-develop` action
  shall not be activated until A4 has landed. This SPEC implements the enforcement; it does not
  activate the action and does not restate the ordering as a requirement of its own.
- **C5 — Template neutrality.** Any template text added under REQ-AP-010 carries no card id, SPEC
  id, internal date, or commit SHA (`CLAUDE.local.md` §25).

## §F — Open items

- **O1** [NEEDS CLARIFICATION: the dispatch named `moai contract decide` alongside `sign`. No such
  verb exists in this tree and A1's verb set is `show` / `verify` / `sign`, with `revoke` owned by
  A3 (§C.3). Resolved by either (a) the lead confirming `decide` is out of scope here, or (b) A3
  defining the verb, after which a one-criterion amendment extends REQ-AP-003's matcher to it.]
  Current disposition: out of scope (§G).
- **O2** [NEEDS CLARIFICATION: whether the lead intends any behavior keyed on a *role* distinction
  (lead vs lane) rather than on agent-vs-human. `MOAI_FACTORY_ROLE` does not exist; the closest
  existing pair is `MOAI_FACTORY_WORKERS` (lead and worker) and `MOAI_FACTORY_WORKER` (worker
  only) (§C.6). Resolved by the lead naming the distinction it wants, or confirming that the
  tool-call boundary of REQ-AP-003 is the whole intent.] Current disposition: REQ-AP-003 denies at
  the tool-call boundary, which is role-independent and strictly stronger.
- **O3** The receipt path is defined but presently unusable, because A1's validator routes every
  Jev decision to a human (§C.4). REQ-AP-006 is evaluable against the *shape* today; whether any
  receipt can in practice sign is A3's question, not this SPEC's.
- **O4** Whether the audit line of REQ-AP-002 and REQ-AP-006 shares a sink with the existing
  branch-guard audit log (`.moai/logs/branch-guard-audit.log`) or takes its own. Design default:
  its own, named in `design.md`; a shared sink is acceptable if the lead prefers one file.

## §G — Exclusions

### Out of Scope — Verbs and surfaces this SPEC does not guard

- `moai contract decide` — no such verb exists and none is being created (§C.3, O1).
- `moai contract revoke` — declared out of scope by A1, owned by A3.
- `moai contract show` / `verify` — read-only; they are positive controls, not deny targets.

### Out of Scope — Bypass shapes that are not observable at PreToolUse

- A sign invocation inside a **script file**, behind a **shell alias**, or in a shell function
  body. The hook sees the script's path, not its contents; a guard that pretended otherwise would
  be claiming an unobservable. `go run ./cmd/moai contract sign` is denied by the fail-closed rule
  of REQ-AP-004 because both words occur, and that is the coverage this SPEC claims.
- Nesting deeper than one `-c` level — denied as unclassifiable (REQ-AP-004), never parsed.

### Out of Scope — Work belonging to adjacent tracks

- Contract schema, verify, digest, signature seal, receipt format and validator — A1 (t1234).
- The escalation detector, its classes, records, and the contract resolver — A2 (t1235).
- Gate rewiring, revocation, moai-issued receipt issuance, and the signing-event store — A3
  (t1236).
- Second-review stop and the closure report — A4 (t1237).
- Activating the `push-develop` action (C4), and amending the Jev display-only doctrine.

### Out of Scope — Mechanisms deliberately not introduced

- A new lease record, a new slot verb, or a second audit-ceiling key.
- Any widening of `internal/mission`'s exported types to fit the contract (REQ-AP-008 fails closed
  instead).
- A Codex-specific environment marker — A1 records that none exists in this repository, and
  measuring one is run-phase work only if a measurement is actually taken.

## §H — Provenance of transferred items

The criterion ids `AC-AE-022`..`AC-AE-025` were **re-used** on the t1235 branch after the split:
in `WT-escalation-detector~1` `AC-AE-022` and `AC-AE-023` both mapped `REQ-AE-024` (the push
serializer), while in `WT-escalation-detector` `AC-AE-022` maps `REQ-AE-022` (ownership decode)
and `AC-AE-023` maps `REQ-AE-023` (signed-invalid) — different criteria wearing the same ids. This
SPEC therefore carries **its own** id namespace and never the `AE` ids. A later reader resolving
"which `AC-AE-022`?" reads the table below.

| This SPEC | Source id (v0.2.1, `WT-escalation-detector~1`) | Carried |
|---|---|---|
| REQ-AP-001, REQ-AP-002, REQ-AP-007 | REQ-AE-024 | text, split by fail direction and by record silence |
| REQ-AP-003, REQ-AP-004, REQ-AP-006, REQ-AP-009 | REQ-AE-025 | text, split by recognition / fail-closed / receipt / marker |
| REQ-AP-005 | — (finding N4) | new requirement; the finding had no requirement |
| REQ-AP-010 | — (finding N8) | new requirement; the finding had no requirement |
| REQ-AP-008 | O9 / N7 | the projection item, which had no owner |
| AC-AP-001, AC-AP-002 | AC-AE-022 | split: deny path, and the unaffected-first-push / mode-off controls |
| AC-AP-003, AC-AP-004 | AC-AE-023 | split: release, and stale/expired reclaim + fail-open |
| AC-AP-005, AC-AP-006 | AC-AE-024 | split: bypass shapes, and positive controls |
| AC-AP-007 | AC-AE-025 | receipt path, rewritten against A1's answered R5 |
| AC-AP-008 .. AC-AP-014 | — | new: N4 wrappers, N8 documentation, the projection, and the not-implemented-verb case |
| Bypass-shape table | spec.md §J | extended with the N4 wrapper rows |
| Mechanics | design.md §G.1, §G.2 | this SPEC's `design.md` §B, §C |
| Milestone | plan.md M6 | split into M1-M4 |

## §I — Cross-references

- SPEC-AUTONOMY-CONTRACT-001 (A1, card t1234, `WT-contract-schema`) — contract schema, `verify`,
  the receipt format and validator, `push_requires_lease`, the agent-environment marker set.
- SPEC-AUTONOMY-ESCALATION-001 (A2, card t1235, `WT-escalation-detector`) — the escalation
  detector and the contract resolver; §K of its spec.md is this SPEC's transfer record.
- `.claude/rules/moai/workflow/resource-slot-lease.md` — the slot lease's liveness, bound, and
  takeover rules.
- `CLAUDE.local.md` §4.1 — why a cancelled CI run on `develop` is a lost verdict.
