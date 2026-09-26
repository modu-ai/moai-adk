---
id: SPEC-AUTONOMY-PRECONDITION-001
title: "A3 preconditions: serialize develop pushes behind the push-develop slot lease, deny agent-invoked contract sign and decide outright, and project the contract onto the mission validator"
version: "0.1.1"
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
- **2026-09-26** — A1's open items **R5** and **R6** are both **resolved**, and the resolutions are
  not symmetric: **R6** answers the mechanism (`push_requires_lease`, §C.5) while **R5** *reduces
  this SPEC's scope* — receipt issuance is A3's, so no receipt-path recognition is specified here
  (§C.4). Either way no criterion is conditional on unowned work, which is the defect (finding
  **N5**) that caused the split.
- **2026-09-26** — v0.1.1 after two lead corrections. (a) **Citation base re-pinned** to A1
  SPEC v0.5.0 at commit **`67a2f55cb`**; every A1 citation in this document is pinned to that SHA
  rather than to the moving branch `WT-contract-schema`, per the lead's citation rule — a dispatch
  names a branch, a SPEC pins the commit it read. `b93af204a` (the earlier base) was verified an
  ancestor of `67a2f55cb`, and every coordinate below was re-measured at `67a2f55cb` rather than
  carried over from the earlier read. (b) **Scope reduced at R5**: the receipt-path allowance and
  its two criteria are withdrawn (§C.4, §H); `decide` is added to the deny matcher alongside
  `sign`. Requirement count 10 → 9, criteria 14 → 12.

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

Every A1 citation below is pinned to commit **`67a2f55cb`** (SPEC-AUTONOMY-CONTRACT-001 v0.5.0) and
was read as `git show 67a2f55cb:.moai/specs/SPEC-AUTONOMY-CONTRACT-001/spec.md`. Line numbers are
that blob's.

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
- A1 (SPEC-AUTONOMY-CONTRACT-001, card t1234) is building it: commit `b93af204a` adds
  `internal/contract/` (schema, 23 verify reason codes, `Decode`/`Digest`/`Verify`/`LoadDir`/
  `ResolveSpecDir` stubs) — package only, no CLI yet. `b93af204a` is an ancestor of the pinned
  `67a2f55cb`.
- A1 assigns this deny to this SPEC by name, at `67a2f55cb:…/spec.md:71`:
  > **Push-serialization enforcement** and the **PreToolUse deny on agent-invoked
  > `moai contract sign`** belong to A2b (t1245). A1 supplies the verify primitive and the schema
  > those detectors read; it enforces nothing at tool-call time.

  and again as a hard precondition at `:151`:
  > **Hard precondition for A3.** Before A3 makes the signature replace Implementation Kickoff
  > Approval, the PreToolUse deny on `moai contract sign` issued from agent tool calls shall exist;
  > A2b (t1245) owns it.
- **Consequence for this SPEC**: the guard is specified against the *command line shape*, which is
  observable at PreToolUse whether or not the binary implements the verb. A guard that denies an
  invocation of a not-yet-existing verb is correct and is not vacuous: the denial is what makes the
  verb agent-unreachable from the moment A1 lands it. REQ-AP-004 states this explicitly.

### C.3 `moai contract decide` — **DOES NOT EXIST, and no track defines it**

- No occurrence of `contract decide` or a `decide` verb anywhere in this tree.
- A1's verb set at `67a2f55cb` is three read/sign commands — `show`, `verify`, `sign`; its
  out-of-scope section at `:84-89` gives `moai contract revoke` to A3. Neither A1's assignment text
  (`:71`, `:151`) nor its ordering paragraph (`:134-140`) names `decide`; all three name `sign`
  only.
- **Lead ruling 09-26 (2nd correction)**: the deny covers `sign` **and** `decide`, outright. This
  SPEC follows that ruling. `decide` is therefore **in scope** (reversing the v0.1.0 disposition),
  guarded by the same shape-based matcher as `sign` — which is sound for the same reason
  (§C.2): a deny at the command-line shape does not require the verb to be implemented, and is what
  makes the verb agent-unreachable from the moment any track lands it. The measurement stands
  alongside the ruling rather than against it: no track currently defines `decide`, so REQ-AP-003
  names it and §F O1 records that its owner is unidentified.

### C.4 R5 — receipt issuance — **RESOLVED by reducing this SPEC's scope**

R5 does **not** hand this SPEC a receipt path to recognize. A1 owns the receipt *format*, its
*validator*, and a non-interactive signing path (`67a2f55cb:…/spec.md:56-57`), but receipt
**issuance** is explicitly A3's — `:85`:

> `### Out of Scope — Revocation and receipt issuance (A3, card t1236)`
>
> - Issuing kickoff receipts from moai itself — moai calling Jev directly and appending the Jev
>   result together with the LLM decision record to a moai-owned append-only store — is A3. A1
>   validates a receipt **file**; it cannot establish who wrote it.

**Consequence, and it is a scope reduction rather than an answer.** A criterion asserting that a
receipt-path invocation is *allowed* could only turn green once A3 implements issuance — precisely
the shape of finding **N5**, the defect that caused this split. Repeating it here would make t1245
uncompletable in the same way. Therefore:

- No receipt-path allowance is specified. The guard denies **every** `sign` / `decide` invocation
  outright and recognizes no exemption (REQ-AP-003, REQ-AP-004).
- The transferred criterion `AC-AE-025` is **not carried**; it is recorded in §H as excluded, owned
  by A3 (t1236), reason N5. This is a lead decision, so it is not an open clarification.
- The guard consequently depends on no A1 or A3 interface at all — a strictly smaller surface than
  v0.1.0 specified.

### C.5 R6 — the serialization mechanism — **ANSWERED, and the field is renamed**

`push_requires_window` is a **retired name**. A1 `:121-125` (A-Q2), quoted verbatim:

> **A-Q2** — Pushing `develop` is allowed inside a contract, but pushes are serialized. Lead
> decision (2026-09-26): the serialization mechanism is a `moai slot` lease on the resource
> `push-develop`, not the integration (merge) window — the push happens outside the merge window
> and is a different resource. A1 represents the permission (`push-develop` action plus the
> `push_develop` configuration switch) and the lease requirement (`push_requires_lease`,
> REQ-CONTRACT-018); **A2b (t1245) enforces serialization.**

That last clause fixes the ownership split this SPEC's requirements are written to: A1 *represents*,
A2b *enforces*. This SPEC therefore reads `push_requires_lease` and never
`push_requires_window`, which appears nowhere in its requirements.

### C.6 The A1 → A2b interface: `show --json` — **the field's producer does not exist yet**

A1 `:378-382` names what it supplies and to whom:

> The `push-develop` action shall denote a push of the integration branch performed while holding a
> `moai slot` lease on the resource `push-develop` (not the integration merge window); the verifier
> shall expose this as a derived `push_requires_lease: true` field in `show --json` output for A2b
> to enforce.

Measured at `553e224f3`: **`moai contract show --json` does not exist** — the whole `contract`
command is absent (§C.2), so neither the subcommand nor the field can be read today. REQ-AP-001 is
therefore written to read the field *from that interface*, matching A1's declared contact point
rather than parsing `contract.yaml` directly, and states that A1 is its producer. Pre-flight
(plan.md §C step 3) re-measures the field's presence before M1 and stops on a rename.

### C.7 `MOAI_FACTORY_ROLE` — **DOES NOT EXIST**

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
- **Consequence, with the lead's re-affirmation taken as the decision.** The lead restated the
  requirement as "deny every `sign` / `decide` call from an agent-role session
  (`MOAI_FACTORY_ROLE=agent`)". The *scope* of that ruling — an outright deny of both verbs — is
  adopted in full (REQ-AP-003). Its *predicate* cannot be the named variable, because the variable
  does not exist; writing a requirement that reads it would make the deny unreachable in every
  session and the criteria would pass while protecting nothing. The predicate is therefore the
  **tool-call boundary itself**: every Bash tool call is by construction agent-issued, so the guard
  denies in every session unconditionally — which delivers "전면 거절" exactly, and is strictly
  stronger than any role variable, because an agent can unset a variable (A1 says so itself at
  `:149-150`) and cannot unset the boundary. REQ-AP-009 reuses A1's closed marker set only where a
  marker is genuinely needed. §F O2 records what remains open: whether the lead wants a *role*
  distinction (lead vs lane) that no existing variable expresses.

### C.8 `mission.ValidateMissionDecision` — **EXISTS at the cited line**

`internal/mission/policy.go:200`:

> `func ValidateMissionDecision(sealed SealedContract, snapshot MissionSnapshot, decision Decision, now time.Time) (OperationReceipt, error)`

with the doc comment "treats every Decision field as untrusted data. It never calls a tool or
performs a state change". Its refusal ladder (`mission_not_running`, `mission_mismatch`,
`policy_mismatch`, `stale_snapshot`, `expired_decision`) is the reuse target of REQ-AP-008.

### C.9 Command-parsing helpers — **EXIST at the cited lines**

`substituteQuotedArguments` (`internal/hook/branch_guard.go:197`) and `extractIntegrationCommand`
(`internal/hook/integration_lock_guard.go:122`). §D notes where quote *removal* is needed instead
of quoted-span *scrubbing*, because scrubbing would erase `'moai'` and defeat the matcher.

## §D — Requirements (GEARS)

### D.1 Push serializer

- **REQ-AP-001** (Capability gate + event) — Where `workflow.autonomy.mode` is `contract` and the
  resolved contract's `actions` contains `push-develop` with `push_requires_lease: true` as read
  from the `moai contract show --json` output A1 supplies for this purpose (`67a2f55cb:…/spec.md:382`,
  §C.6) rather than by parsing `contract.yaml` directly, when a
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
  `workflow.autonomy.mode` and with no exemption of any kind, every Bash tool call that invokes
  `moai contract sign` or `moai contract decide`, with a reason prefixed
  `CONTRACT_SIGN_AGENT_VIOLATION:`, recognizing the invocation through shell quoting, leading
  `NAME=value` assignments, the prefix wrappers of REQ-AP-005, any path to an executable whose
  basename is `moai`, global flags preceding the verb, and one level of `sh -c` / `bash -c` /
  `zsh -c`. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AP-004** (Unwanted) — When a command carries the word `contract` together with `sign` or
  `decide` but its structure cannot be classified — command substitution, a variable in program
  position, `eval`, an unenumerated wrapper, or nesting deeper than one `-c` level — the guard shall
  not allow it, and shall deny it with the reason marked unclassified (fail closed). The guard shall
  deny such a command whether or not the invoked verb is implemented in the installed binary (§C.2,
  §C.3).
- **REQ-AP-005** (Ubiquitous; finding N4) — The guard's prefix-wrapper set shall be a closed,
  enumerated list containing at least `env`, `command`, `exec`, `nohup`, `script`, `timeout`,
  `sudo`, `stdbuf`, `nice`, and `xargs`, and shall skip each wrapper's own options before reading
  the program word; a wrapper not in the list, encountered in program position, shall be treated
  as unclassifiable under REQ-AP-004 rather than allowed.
- **REQ-AP-006** — **WITHDRAWN at v0.1.1** (receipt-path allowance). Receipt issuance is A3's
  (§C.4), so this SPEC recognizes no receipt path and grants no exemption. The id is retired and is
  **not** re-used, so a citation of REQ-AP-006 can only refer to the withdrawn meaning (§H).
- **REQ-AP-009** (Ubiquitous; §C.7) — Where the guard needs an agent-environment signal for a
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
- **C4 — Required ordering, and what this SPEC's criteria therefore presume.** A1 `:134-140` states
  it verbatim:

  > **Required ordering: A1 → A2 (t1235) + A2b (t1245) → A3 (t1236).** A3 shall not activate
  > "signature replaces Kickoff" until A2 and A2b have landed — in particular A2b's
  > push-serialization enforcement and PreToolUse deny on agent-invoked `moai contract sign`. A4
  > (t1237) owns the stop before push when the second review was not performed, and the
  > `push-develop` action shall not be activated until A4 has landed. Until then a signed contract
  > and `mode: contract` confer no autonomy: the Kickoff gate stays in force (REQ-CONTRACT-019).

  Three consequences bind this SPEC. (i) It is a **precondition of A3**, so its criteria are
  evaluated against fixtures, never against a live autonomous run — no criterion here asserts that
  any autonomy was granted. (ii) The `push-develop` action is **not activated until A4 lands**, so
  the push-serializer criteria (acceptance.md §A) drive the activation triple from fixture
  configuration rather than from an actually-active action; that is the correct evaluation basis,
  not a weakening. (iii) This SPEC implements the enforcement only: it does not activate the action,
  does not make a signature replace Kickoff, and does not restate the ordering as a requirement of
  its own.
- **C5 — Template neutrality.** Any template text added under REQ-AP-010 carries no card id, SPEC
  id, internal date, or commit SHA (`CLAUDE.local.md` §25).

## §F — Open items

- **O1** [RESOLVED — lead ruling: `decide` is in scope, denied outright alongside `sign`
  (REQ-AP-003).] One measured fact remains recorded without blocking anything: **no track currently
  defines a `decide` verb** (§C.3) — not A1, not A3's declared scope. The deny is therefore
  forward-looking by construction, which is sound (a shape-based deny needs no implementation), and
  the open part is only *who will define the verb*. It blocks no criterion: AC-AP-005 evaluates the
  deny against the command shape today.
- **O2** [PARTIALLY RESOLVED — the lead's scope ruling (deny every `sign` / `decide` call from an
  agent session) is adopted in full.] What stays open is narrower than v0.1.0 stated:
  `MOAI_FACTORY_ROLE` does not exist (§C.7), so the predicate is the tool-call boundary, which
  denies in every session and therefore *covers* the ruling rather than approximating it. The open
  question is whether the lead additionally wants a **role** distinction (lead vs lane) — no
  existing variable expresses it; the closest pair is `MOAI_FACTORY_WORKERS` (lead and worker) and
  `MOAI_FACTORY_WORKER` (worker only). Resolved by the lead either confirming the boundary is the
  whole intent, or naming the distinction and its owner.
- **O3** Whether the audit line of REQ-AP-002 shares a sink with the existing
  branch-guard audit log (`.moai/logs/branch-guard-audit.log`) or takes its own. Design default:
  its own, named in `design.md`; a shared sink is acceptable if the lead prefers one file.

## §G — Exclusions

### Out of Scope — Receipt issuance and receipt-path recognition (A3, card t1236)

- Issuing kickoff receipts from moai itself — A1 `:85` gives it to A3. This SPEC recognizes **no**
  receipt path and grants **no** exemption, because a criterion asserting a receipt invocation is
  allowed could only turn green after A3 (§C.4, finding N5).
- Validating a receipt's contents — A1's validator owns it (REQ-CONTRACT-023).
- The transferred criterion `AC-AE-025` — excluded, not carried (§H).

### Out of Scope — Verbs and surfaces this SPEC does not guard

- `moai contract revoke` — declared out of scope by A1 at `:87`, owned by A3.
- `moai contract show` / `verify` — read-only; they are positive controls, not deny targets. Note
  that `show --json` is *consumed* by REQ-AP-001 (§C.6) and is not a deny target.

### Out of Scope — Bypass shapes that are not observable at PreToolUse

- A sign invocation inside a **script file**, behind a **shell alias**, or in a shell function
  body. The hook sees the script's path, not its contents; a guard that pretended otherwise would
  be claiming an unobservable. `go run ./cmd/moai contract sign` is denied by the fail-closed rule
  of REQ-AP-004 because `contract` and the verb both occur, and that is the coverage this SPEC
  claims.
- Nesting deeper than one `-c` level — denied as unclassifiable (REQ-AP-004), never parsed.

### Out of Scope — Work belonging to adjacent tracks

- Contract schema, verify, digest, signature seal, receipt format and validator, and the
  `show --json` projection this SPEC consumes — A1 (t1234).
- The escalation detector, its classes, records, and the contract resolver — A2 (t1235).
- Gate rewiring, revocation, moai-issued receipt issuance, and the signing-event store — A3
  (t1236).
- Second-review stop, the closure report, and **activating the `push-develop` action** — A4 (t1237),
  per A1 `:136-138` (C4).
- Making a signature replace Implementation Kickoff Approval — A3, and only after this SPEC lands
  (C4).
- Amending the Jev display-only doctrine.

### Out of Scope — Mechanisms deliberately not introduced

- A new lease record, a new slot verb, or a second audit-ceiling key.
- Any widening of `internal/mission`'s exported types to fit the contract (REQ-AP-008 fails closed
  instead).
- A Codex-specific environment marker — A1 records that none exists in this repository, and
  measuring one is run-phase work only if a measurement is actually taken.

## §H — Transfer and ID-reuse record

Written in the form A1 uses for the same hazard (its v0.3.0 changelog row records an "ID reuse
record" for `AC-CONTRACT-024` / `REQ-CONTRACT-021`), so this repository has one convention rather
than two.

### H.1 ID reuse record (source side, t1235)

**The criterion ids `AC-AE-022`..`AC-AE-025` were re-used after the split.** In
`WT-escalation-detector~1` (v0.2.1) `AC-AE-022` and `AC-AE-023` both mapped `REQ-AE-024` (the push
serializer); in `WT-escalation-detector` (v0.3.0) `AC-AE-022` maps `REQ-AE-022` (ownership decode)
and `AC-AE-023` maps `REQ-AE-023` (signed-invalid) — different criteria wearing the same ids. Any
citation of `AC-AE-022`..`AC-AE-025` in a plan-audit report predating t1235 v0.3.0 refers to the
**withdrawn** meanings. This SPEC therefore carries its own id namespace and never an `AE` id.

### H.2 ID retirement record (this SPEC)

**`REQ-AP-006` was withdrawn at v0.1.1** (receipt-path allowance) and is **not** re-used; a citation
of it can only refer to the withdrawn meaning. The criteria `AC-AP-011` and `AC-AP-012`, which
mapped it, were removed at the same revision and their ids are likewise retired, not renumbered
into other criteria.

### H.3 Provenance of carried items

| This SPEC | Source id (v0.2.1, `WT-escalation-detector~1`) | Carried |
|---|---|---|
| REQ-AP-001, REQ-AP-002, REQ-AP-007 | REQ-AE-024 | text, split by fail direction and by record silence; the field renamed `push_requires_lease` and read from `show --json` (§C.5, §C.6) |
| REQ-AP-003, REQ-AP-004, REQ-AP-009 | REQ-AE-025 | text, split by recognition / fail-closed / marker; **receipt limb dropped**, `decide` added |
| REQ-AP-005 | — (finding N4) | new requirement; the finding had no requirement |
| REQ-AP-010 | — (finding N8) | new requirement; the finding had no requirement |
| REQ-AP-008 | O9 / N7 | the projection item, which had no owner |
| AC-AP-001, AC-AP-002 | AC-AE-022 | split: deny path, and the unaffected-first-push / off-condition controls |
| AC-AP-003, AC-AP-004 | AC-AE-023 | split: release, and stale/expired reclaim + fail-open |
| AC-AP-005, AC-AP-006 | AC-AE-024 | split: bypass shapes (extended with `decide`), and positive controls |
| AC-AP-007 .. AC-AP-010, AC-AP-013, AC-AP-014 | — | new: fail-closed, the not-implemented-verb case, N4 wrappers, N8 documentation, the projection. The gap at 011-012 is the retirement of H.2, not a numbering slip |
| Bypass-shape table | spec.md §J | extended with the N4 wrapper rows |
| Mechanics | design.md §G.1, §G.2 | this SPEC's `design.md` §B, §C |
| Milestone | plan.md M6 | split into M1-M4 |

### H.4 Transferred but **excluded** — not carried here

| Source id | Disposition |
|---|---|
| `AC-AE-025` (receipt path allowed / forged receipt denied) | **Not transferred.** Owned by **A3 (card t1236)**, which owns receipt issuance (A1 `:85`). Reason: **N5** — the criterion could only turn green after A3, so carrying it would make t1245 uncompletable the same way t1235 was. Lead decision; not an open question. |

## §I — Cross-references

- SPEC-AUTONOMY-CONTRACT-001 (A1, card t1234) — contract schema, `verify`, the receipt format and
  validator, `push_requires_lease`, `show --json`, the agent-environment marker set, and the
  required ordering. **Every citation in this document is pinned to commit `67a2f55cb`** (that
  SPEC's v0.5.0); the branch `WT-contract-schema` moves, a pinned SHA does not.
- SPEC-AUTONOMY-CONTRACT-001 §§ owned by A3 (card t1236) — receipt issuance, revocation, and the
  signing-event store; the receipt-path criterion this SPEC excludes (§H.4).
- SPEC-AUTONOMY-ESCALATION-001 (A2, card t1235, `WT-escalation-detector`) — the escalation
  detector and the contract resolver; §K of its spec.md is this SPEC's transfer record.
- `.claude/rules/moai/workflow/resource-slot-lease.md` — the slot lease's liveness, bound, and
  takeover rules.
- `CLAUDE.local.md` §4.1 — why a cancelled CI run on `develop` is a lost verdict.
