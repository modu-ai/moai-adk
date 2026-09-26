---
id: SPEC-AUTONOMY-PRECONDITION-001
title: "A3 preconditions: serialize develop pushes behind the push-develop slot lease, deny the human contract-signing path at the tool-call boundary, deny the non-interactive sign path and decide from agent-role sessions, and project the contract onto the mission validator"
version: "0.1.3"
status: in-progress
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
  §K at that SPEC's branch; the transferred text was read from commit **`d8926ff9a`** — that
  SPEC's v0.2.1, the last commit before the split removed it, verified by
  `git show d8926ff9a:…/SPEC-AUTONOMY-ESCALATION-001/spec.md | grep '^version:'` → `version: "0.2.1"`
  — through committed refs only, never through the t1235 worktree, which another session is writing.
  <!-- moving-ref-ok: this ref is the SUBJECT of a recorded correction, not an address into a tree; replacing it with a SHA would erase the record of which citation was wrong -->
  The earlier citation `WT-escalation-detector~1` was **wrong and is corrected here**: `~1` is a
  relative ref, it slid as that branch advanced, and it now resolves to a post-split commit in which
  `AC-AE-025` does not occur at all. A pinned SHA is the only citation form this document uses for
  transferred text.
- **2026-09-26** — every premise handed to this SPEC in the dispatch was measured against
  `553e224f3` before any requirement was written. The measured table (premise → exists? →
  evidence) is §C. Two dispatch premises did **not** hold and the requirements below say so
  rather than assuming them (`moai contract sign` does not exist yet; `MOAI_FACTORY_ROLE` does
  not exist at all).
- **2026-09-26** — A1's open items **R5** and **R6** are both **resolved**, and the resolutions are
  not symmetric: **R6** answers the mechanism (`push_requires_lease`, §C.5) while **R5** *reduces
  this SPEC's scope* — receipt issuance is A3's, so no receipt-path recognition is specified here
  (§C.4).
- **2026-09-26** — v0.1.2, after the iter-1 plan audit (FAIL 0.71, measured at `87929d6df`) and
  three further lead rulings. Six changes bear on how this SPEC is read.
  (a) **Every citation is re-pinned to a SHA.** Transferred text is cited at `d8926ff9a`; A1 is
  cited at **`25283ebf8`** (A1 SPEC v0.5.2), and every coordinate in §C was re-measured against that
  blob rather than carried over from the earlier `67a2f55cb` read. Four remaining branch names in
  this document are *subjects* of a claim rather than addresses into a tree, and each carries an
  inline exemption marker stating why (§H.1, §I).
  (b) **The deny is no longer one rule but three** (§C.3, §C.7, REQ-AP-003 / REQ-AP-011): the human
  signing path is refused on every tool call, while the non-interactive sign path and `decide` are
  refused only from a session whose `MOAI_FACTORY_ROLE` is `worker`. The reason is concrete and was
  the lead's: `decide` is a path the **lead session's own LLM** invokes as a tool call when the
  decider is `llm` or `llm+jev`, so an outright deny of every tool call would block a legitimate
  path. An outright deny is therefore *not* strictly stronger, as v0.1.1 claimed — it is wrong.
  (c) **This SPEC defines that marker's name and value constants itself** (REQ-AP-012), which is what
  makes the role-scoped criteria evaluable now rather than after card t1240.
  (d) **REQ-AP-010 is narrowed to a documentation surface this SPEC creates**, so no criterion
  depends on A1's template autonomy block any more (§C.10). v0.1.1's claim that no criterion was
  conditional was **false** — `AC-AP-013` was, and the audit was right to call it a recurrence of
  finding N5. It is repaired by moving the surface, not by softening the claim.
  (e) **The mission-validator projection is specified** (§C.11, REQ-AP-008): A1's drafted mapping is
  adopted with three measured corrections.
  (f) **§H was rebuilt by reading `d8926ff9a`** rather than reconstructed from a summary, which is
  what produced its three mis-attributions.
- **2026-09-26** — v0.1.3. One structural change: **the role value is no longer pinned by a
  ruling, it is pinned by an assertion** (REQ-AP-013, AC-AP-018). The value has now been renamed
  twice under this card — `agent`, then `worker` when `agent` turned out to be the retired spelling —
  and the operator has since decided to rename the factory role vocabulary again, to `lane`, under
  card **t1256**. The failure each rename risks is the same one and it is silent: a guard keyed on a
  dead spelling denies nothing **while every criterion still passes**. So this revision does not chase
  the string. It requires the guard's expected value to be pinned to the same vocabulary the launcher
  stamps from, by an equality assertion that goes RED when a rename touches one side only; §F O6
  records that the literal spelling now follows t1256 rather than being decided here. Nothing else
  changes: the value stays `worker` in this document, because `worker` is what develop `553e224f3`
  carries and renaming it is t1256's work, not this SPEC's.

- **2026-09-26** — v0.1.1 after two lead corrections. (a) **Citation base re-pinned** to A1
  SPEC v0.5.0 at commit **`67a2f55cb`**; every A1 citation in this document is pinned to that SHA
  <!-- moving-ref-ok: the subject is the citation rule this revision applied, so the branch name is what the sentence is about rather than a coordinate anything is read from -->
  rather than to the moving branch `WT-contract-schema`, per the lead's citation rule — a dispatch
  names a branch, a SPEC pins the commit it read. `b93af204a` (the earlier base) was verified an
  ancestor of `67a2f55cb`, and every coordinate below was re-measured at `67a2f55cb` rather than
  carried over from the earlier read. This pin is superseded by v0.1.2's (`25283ebf8`).
  (b) **Scope reduced at R5**: the receipt-path allowance and
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

Every A1 citation below is pinned to commit **`25283ebf8`** (SPEC-AUTONOMY-CONTRACT-001 v0.5.2) and
was read as `git show 25283ebf8:.moai/specs/SPEC-AUTONOMY-CONTRACT-001/spec.md`, verified
`version: "0.5.2"` at that blob's line 4. Line numbers below are that blob's. **v0.5.2 is used rather
than the v0.5.1 (`65e0a9167`) the dispatch named**, for two measured reasons: v0.5.2 is the newer
commit on the same history and is the revision that *corrected the mission-projection owner and the
sign-deny attribution to A2b* — the very statements this SPEC cites — and the sibling A2 SPEC already
pins there (`78a95ee03`, "pin A1 citations to v0.5.2 25283ebf8"). The v0.5.1→v0.5.2 diff shifts
later lines **non-uniformly**, which is why a v0.5.1 pin cannot be rescued by a constant offset:
measured by locating each v0.5.2 line verbatim in the v0.5.1 blob, the `### Out of Scope —
Mission-validator projection` heading moves `78 → 79` (**+1**) while `:329`, `:410` and `:462` each
move by **+3**, and `:231` — the interim-A1-rule line — has **no v0.5.1 counterpart at all**. The
file grows 520 → 524 lines. A v0.5.1 pin would therefore carry line numbers off by a varying amount
against the text actually quoted here, and would cite one statement that did not yet exist.

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
- `internal/cli` carries no `contract.go`, and no cobra command is registered under that name:
  `grep -rn 'Use:.*"contract' internal/cli/` returns **0 rows**. That zero is the whole of the claim.
  The word `contract` does occur in seven `internal/cli` filenames — `codex_contract.go`,
  `codex_contract_test.go`, `codex_contract_link_test.go`, `codex_contract_fixture_unix_test.go`,
  `codex_contract_fixture_windows_test.go`, `exitcode_contract_test.go`,
  `hook_dest_contract_test.go` — all of them the unrelated codex contract-link feature and two
  unrelated output-contract test files. None registers a `contract` command, which is why the
  registration count rather than the string count is the evidence. (v0.1.1 claimed the string
  occurred only in `codex_init_test.go`; that was an overstatement and is corrected here.)
- A1 (SPEC-AUTONOMY-CONTRACT-001, card t1234) is building it: commit `b93af204a` adds
  `internal/contract/` (schema, 23 verify reason codes, `Decode`/`Digest`/`Verify`/`LoadDir`/
  `ResolveSpecDir` stubs) — package only, no CLI yet. `b93af204a` is an ancestor of the pinned
  `25283ebf8` (`git merge-base --is-ancestor b93af204a 25283ebf8` → exit 0).
- A1 assigns this deny to this SPEC by name, at `25283ebf8:…/spec.md:72-74`:
  > **Push-serialization enforcement** and the **PreToolUse deny on agent-invoked
  > `moai contract sign`** belong to A2b (t1245). A1 supplies the verify primitive and the schema
  > those detectors read; it enforces nothing at tool-call time.

  and again as a hard precondition at `:158-159`:
  > **Hard precondition for A3.** Before A3 makes the signature replace Implementation Kickoff
  > Approval, the PreToolUse deny on `moai contract sign` issued from agent tool calls shall exist;
  > A2b (t1245) owns it.
- **The signing paths A1 defines, which is what the three-way split keys on.** A1 `:329`:
  > Where the signer is human (no `--signer`, or `--signer human`), the signer shall refuse to sign and exit

  and `:462`:
  > Where the signer is `llm` or `llm+jev`, `moai contract sign --signer <kind> --receipt <path>`

  So the two signing paths are distinguishable **on the command line**, which is the only thing
  PreToolUse sees: the human path is `sign` with no `--signer` or with `--signer human`; the
  non-interactive path is `sign --signer llm` or `sign --signer llm+jev`. REQ-AP-003 and REQ-AP-011
  are written against that distinction rather than against a property of the process.
- **Consequence for this SPEC**: the guard is specified against the *command line shape*, which is
  observable at PreToolUse whether or not the binary implements the verb. A guard that denies an
  invocation of a not-yet-existing verb is correct and is not vacuous: the denial is what makes the
  verb agent-unreachable from the moment A1 lands it. REQ-AP-004 states this explicitly.

### C.3 `moai contract decide` — **DOES NOT EXIST, and no track defines it**

- No occurrence of `contract decide` or a `decide` verb anywhere in this tree.
- A1's verb set at `25283ebf8` is three read/sign commands — `show`, `verify`, `sign`; its
  out-of-scope section at `:87` (heading) / `:89` gives `moai contract revoke` to A3. Neither A1's assignment text
  (`:72-74`, `:158-159`) nor its ordering paragraph (`:142-148`) names `decide`; all three name
  `sign` only.
- **`decide` is owned by A3 (card t1236).** A1 `:90-92` gives A3 "the kickoff decision rules … the
  `llm+jev` cross-check agreement … and the Jev-failure → `llm` fallback decision", which is the
  decision this verb would carry. This SPEC guards it; it does not define it.
- **Lead ruling 09-26 (3rd correction) — the deny is three rules, not one.** The outright deny of
  v0.1.1 is withdrawn, for a measured reason rather than a preference: `decide` is a path the **lead
  session's own LLM** invokes as a tool call when the decider is `llm` or `llm+jev` (the decider set
  is `human | llm | llm+jev`, A1 v0.5.1 changelog row `:30` — `llm` is not the only LLM decider), so
  a rule that denied every tool call would deny a legitimate path. The three rules are:

  | Invocation | Refused where | Requirement |
  |---|---|---|
  | `moai contract sign` on the **human** path (no `--signer`, or `--signer human`) | **every** tool call, in every session, no exemption | REQ-AP-003 |
  | `moai contract sign --signer llm` / `--signer llm+jev` (the non-interactive path) | only a session whose `MOAI_FACTORY_ROLE` is `worker` | REQ-AP-011 |
  | `moai contract decide` | only a session whose `MOAI_FACTORY_ROLE` is `worker` | REQ-AP-011 |

  A session carrying **no** marker — a human-launched session, or the lead — is **allowed** to
  `decide`. The tool-call boundary is the right discriminator for the human path only, because human
  signing happens at a terminal and therefore never as a tool call; it is the wrong discriminator for
  the other two, because those paths have a legitimate tool-call caller.
- **The deny is still sound while the verb does not exist**, for the same reason as `sign` (§C.2): a
  deny at the command-line shape does not require the verb to be implemented, and is what makes it
  agent-unreachable from the moment A3 lands it.

### C.4 R5 — receipt issuance — **RESOLVED by reducing this SPEC's scope**

R5 does **not** hand this SPEC a receipt path to recognize. A1 owns the receipt *format*, its
*validator*, and a non-interactive signing path (`25283ebf8:…/spec.md:59-60`), but receipt
**issuance** is explicitly A3's — `:87`, `:94-98`:

> `### Out of Scope — Revocation and receipt issuance (A3, card t1236)`
>
> - Issuing kickoff receipts from moai itself — moai calling Jev directly and appending the Jev
>   result together with the LLM decision record to a moai-owned append-only store — is A3. A1
>   validates a receipt **file**; it cannot establish who wrote it.

**Consequence, and it is a scope reduction rather than an answer.** A criterion asserting that a
receipt-path invocation is *allowed* could only turn green once A3 implements issuance — precisely
the shape of finding **N5**, the defect that caused this split. Repeating it here would make t1245
uncompletable in the same way. Therefore:

- No receipt-path allowance is specified, and the guard grants **no** exemption — it never reads
  `--receipt`. Every **human-path** `sign` invocation is denied whatever a `--receipt` argument says
  (REQ-AP-003). The non-interactive path and `decide` are gated on the role marker rather than on a
  receipt (REQ-AP-011): `--signer` selects which rule applies, and `--receipt` travels through as an
  opaque argument the guard does not inspect.
- The transferred criterion `AC-AE-025` is **not carried**; it is recorded in §H as excluded, owned
  by A3 (t1236), reason N5. This is a lead decision, so it is not an open clarification.
- The guard consequently depends on no A1 or A3 interface at all — a strictly smaller surface than
  v0.1.0 specified.

### C.5 R6 — the serialization mechanism — **ANSWERED, and the field is renamed**

`push_requires_window` is a **retired name**. A1 `:128-133` (A-Q2), quoted verbatim:

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

A1 `:404-410` (REQ-CONTRACT-018) names what it supplies and to whom:

> The `push-develop` action shall denote a push of the integration branch performed while holding a
> `moai slot` lease on the resource `push-develop` (not the integration merge window); the verifier
> shall expose this as a derived `push_requires_lease: true` field in `show --json` output for A2b
> to enforce.

Measured at `553e224f3`: **`moai contract show --json` does not exist** — the whole `contract`
command is absent (§C.2), so neither the subcommand nor the field can be read today. REQ-AP-001 is
therefore written to read the field *from that interface*, matching A1's declared contact point
rather than parsing `contract.yaml` directly, and states that A1 is its producer. Pre-flight
(plan.md §C step 3) re-measures the field's presence before M1 and stops on a rename.

### C.7 `MOAI_FACTORY_ROLE` — **DOES NOT EXIST; this SPEC defines it**

- `grep -rn 'MOAI_FACTORY_ROLE' internal cmd pkg` → **0 rows** (re-measured at `eee5f635e`).
  `internal/config/envkeys.go` carries `EnvMoaiFactoryWorkers = "MOAI_FACTORY_WORKERS"` (`:280`, set
  on the lead **and** every worker) and `EnvMoaiFactoryWorker = "MOAI_FACTORY_WORKER"` (`:287`,
  worker only, value `lane-<n>`). `isFactoryRoleToken` (`internal/cli/codex_factory.go:103-105`)
  parses a **CLI argument**, not an environment variable.
- A1 owns a closed *harness* marker set, and deliberately does not apply it to the path REQ-AP-011
  covers. `25283ebf8:…/design.md:416-426` lists `CLAUDECODE` (measured present in Claude Code's Bash
  environment) and `CLAUDE_CODE_SESSION_ID` (`internal/config/envkeys.go:494`), states the set is
  "checked by `sign` on the **human** path before anything else (REQ-CONTRACT-021)", and then states
  at `:419-421` that **"the receipt path does not check markers — it is the path meant for automated
  deciders"**. That sentence is why REQ-AP-011 exists: A1 leaves the non-interactive path ungated by
  any marker, so the role gate at the tool-call boundary is this SPEC's to supply.
- **Consequence — the predicate is the variable, and this SPEC creates the variable.** v0.1.1
  reasoned that a requirement reading `MOAI_FACTORY_ROLE` would be unreachable because the variable
  does not exist, and substituted the tool-call boundary. The premise was true and the substitution
  is still correct **for the human path**, but wrong for the other two: `decide` has a legitimate
  tool-call caller (§C.3), so a boundary-keyed deny of `decide` denies the lead. The reachability
  problem is therefore solved at its source instead — **REQ-AP-012 defines the name and value
  constants in `internal/config/envkeys.go` as part of this SPEC's own change**, so the guard reads a
  constant that exists in the tree it ships in, and the role-scoped criteria (AC-AP-015, AC-AP-016)
  turn green with the variable set by the test alone. No dependency on card t1240 remains.
- **The value spelling, measured and then decided.** An earlier ruling named the value `agent`. In
  `internal/cli/factory.go` the live CLI role token is `worker` (`:59`
  `factoryWorkerRoleToken = "worker"`) and **`agent` is its retired pre-rename spelling** (`:64`
  `factoryLegacyAgentRoleToken = "agent"`, kept as a parsing alias that prints a deprecation hint).
  The failure mode of keying on the retired spelling was concrete: **if card t1240 stamps the live
  spelling, a guard keyed on `agent` never fires and AC-AP-016's allow arm passes anyway** — green
  criteria over zero protection. The operator therefore resolved §F O5 to **`worker`**, the canonical
  spelling, and only `worker` is accepted; `agent` is not. The value carried to card t1240 is
  `worker` (§G).
- **The environment an agent can unset.** A1 `:155-156` records that "an agent can unset environment
  variables". REQ-AP-011's gate is therefore **nominal in production** — in both directions: until
  t1240 stamps the marker nothing is denied, and after it does, an agent that unsets the variable is
  not denied either. This is recorded, not designed around: the human path, which is the one that
  actually voids a contract, is keyed on the boundary precisely because the boundary cannot be unset
  (REQ-AP-003). §F O2 carries the residual exposure.
- **What a wrapper cannot do to the read, and what stays unmeasured.** The guard reads the
  **calling session's** environment at PreToolUse, before any wrapper in the command executes, so
  `sudo` (which drops the environment) and `env -i` cannot affect the role read — they alter the
  child's environment, not the hook's. What is **not** measured is whether every harness populates
  the hook process's environment identically to the session's; that is recorded as a Gap in
  progress.md §E.1 rather than asserted here.

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

### C.10 The documentation surfaces of finding N8 — **BOTH absent, and only one of them is A1's**

The audit found `AC-AP-013` conditional on A1's template autonomy block, which is a recurrence of
finding **N5**. Measured at `eee5f635e`, the position is worse than the audit stated — *neither*
surface v0.1.1 named exists:

| Surface v0.1.1 named | Measured | Owner |
|---|---|---|
| The template `workflow.yaml` autonomy block and its comment | `grep -n autonomy internal/template/templates/.moai/config/sections/workflow.yaml` → **no output**; the block does not exist | A1 (REQ-CONTRACT-016 template default and neutrality) |
| "the rule text that states contract-mode behavior" | `grep -rln 'workflow.autonomy' .claude/rules/ internal/template/templates/.claude/rules/` → **no file**; no such rule text exists, and no track in this epic claims it | **unowned** |

So the surface could not be re-pointed at "the part this SPEC owns", because v0.1.1 owned neither.
The repair is to give the requirement a surface **this SPEC creates**: a new `paths:`-scoped rule file
documenting this SPEC's own guard, together with its template mirror, which is ordinary
Template-First work for a guard this card implements (`CLAUDE.local.md` §2 Template-First Rule). It is
`paths:`-scoped rather than always-loaded deliberately — an always-loaded rule file spends the
always-loaded budget for every session, which is not warranted by a guard's documentation.

REQ-AP-010 is narrowed to that surface, and the two limbs this SPEC does not own are transferred
(§H.5): the template comment goes with the block, to A1; the re-scoping of the sentence "nothing
changes under `guided`" belongs to the SPEC that wrote it, A2 (REQ-AE-001). What stays here is
evaluable today against a file this change creates, so no criterion is conditional on unowned work.

### C.11 The mission projection A1 hands over (R8) — **drafted, with three measured corrections**

A1 v0.5.2 assigns the projection to this SPEC and supplies a draft. `25283ebf8:…/spec.md:79-85`:

> `### Out of Scope — Mission-validator projection (A2b, card t1245)`
>
> - Projecting a signed SPEC contract onto the `/moai goal --auto` mission contract so that the
>   mission decision validator can judge individual operations is deferred to A2b (t1245), its only
>   consumer. The projection needs a glob-to-prefix scope translation that the mission validator's
>   exact-or-prefix containment check requires. Forward note for A2b: `design.md` § Forward Note —
>   Mission Projection.
> - A1 records no `mission_contract_sha256`; the contract digest is the sole tamper authority.

The drafted 12-row mapping is at `25283ebf8:…/design.md:470-494`. **This SPEC adopts it** — it maps
onto the real `mission.MissionContract` field set — **with three corrections, each measured in this
tree rather than inferred:**

1. **The table is one field short, and the missing field is the one that decides admission.**
   `mission.MissionContract` has **13** fields (`internal/mission/contract.go:32-46`); the drafted
   table covers 12 and omits `Approved bool`. `contractComplete` (`:55-61`) requires
   `c.Approved` true, so a projection built from the 12 drafted rows alone is rejected by the sealer
   as `incomplete_contract` (`internal/mission/contract.go:82`) and **every** verdict fails closed.
   REQ-AP-008 therefore requires the projection to set `Approved` from the contract's signature
   state, and AC-AP-014 asserts a signed-valid fixture reaches a verdict at all.
2. **The glob defect is confirmed at the line A1 names.** `targetInsideScope`
   (`internal/mission/policy.go:112-118`) is exact-or-prefix —
   `target == allowed || strings.HasPrefix(target, allowed+"/")` — so `internal/foo/**` never
   contains `internal/foo/x.go`. The projection translates a trailing `/**` to its prefix; a glob
   with an **inner** wildcard (`internal/*/x.go`) has no prefix that preserves its meaning and is
   therefore **unmappable**, which under REQ-AP-008 fails closed and names the field rather than
   silently widening or narrowing scope. Widening is the hazard that matters: a scope quietly
   widened by a translation admits writes the contract did not permit.
3. **`ResourceLimits.MaxOperations` must be positive, and `card` has no mission counterpart.**
   `contractComplete` also requires `MaxOperations > 0` (`:58`), so a contract whose
   `budget.operations` is zero or absent is unmappable rather than mapped to zero. And A1 v0.5.2 adds
   a **required top-level `card` field** covered by the canonical digest (REQ-CONTRACT-001,
   `:260-266`), reported by `show --json` (REQ-CONTRACT-014, `:362-369`) — this SPEC's own input.
   Under a bare "unmapped field fails closed" rule `card` would fail **every** contract, so
   REQ-AP-008 distinguishes a field that is *deliberately not projected* (enumerated, with `card` on
   that list) from one that is *unmapped* (not enumerated → fail closed). Without that distinction
   R9 silently breaks the projection.

Not adopted, and recorded rather than assumed: `mission_contract_sha256` is not recorded (A1 `:85`),
so the projection carries no second tamper authority.

### C.12 The moai-owned signing store — **A3's, and this SPEC does not presume it (R10)**

A1 `:94-98` places the append-only store at `$MOAI_HOME/db/<project-key>/contract/` and states
plainly that **"A1 does not create that store"** — it is A3's. No requirement here reads, writes, or
presumes that path; the guard reads a command line and the projection reads `show --json` output.
Recorded so that a later reader does not add a store-presence precondition to a criterion and make it
conditional on A3.

## §D — Requirements (GEARS)

### D.1 Push serializer

- **REQ-AP-001** (Capability gate + event) — Where `workflow.autonomy.mode` is `contract` and the
  resolved contract's `actions` contains `push-develop` with `push_requires_lease: true` as read
  from the `moai contract show --json` output A1 supplies for this purpose (`25283ebf8:…/spec.md:410`,
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

### D.2 Contract-sign and contract-decide guard

The three rules of §C.3 are three requirements, not one with three limbs, because they differ in the
predicate that activates them and a reader must be able to tell which one denied a call.

- **REQ-AP-003** (Ubiquitous) — The guard shall deny, regardless of `workflow.autonomy.mode`, in
  every session, and with no exemption of any kind, every Bash tool call that invokes
  `moai contract sign` on the **human signing path** — that is, with no `--signer` flag, or with
  `--signer human` (A1 `:329`) — with a reason prefixed `CONTRACT_SIGN_AGENT_VIOLATION:`, recognizing
  the invocation through shell quoting, leading `NAME=value` assignments, the prefix wrappers of
  REQ-AP-005, any path to an executable whose basename is `moai`, global flags preceding the verb,
  and one level of `sh -c` / `bash -c` / `zsh -c`. The predicate is the tool-call boundary itself:
  human signing happens at an operator terminal, so a human-path signature arriving as a tool call is
  by construction not the signature the contract model requires, and the boundary is the one signal an
  agent cannot unset. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AP-011** (Capability gate) — Where the calling session's environment sets
  `MOAI_FACTORY_ROLE` (REQ-AP-012) to the value `worker`, the guard shall deny every Bash tool call
  that invokes `moai contract sign` on the **non-interactive path** (`--signer llm` or
  `--signer llm+jev`, A1 `:462`) or that invokes `moai contract decide`, with a reason prefixed
  `CONTRACT_SIGN_AGENT_VIOLATION:` and recognized through the same shapes REQ-AP-003 enumerates; and
  where that variable is absent or holds any other value, the guard shall **allow** such a call,
  leaving its hook output unchanged. The allow direction is a requirement rather than an omission: it
  is the lead session's own `decide` path when the decider is `llm` or `llm+jev` (§C.3), and a guard
  that denied it would deny the caller the epic depends on.
- **REQ-AP-012** (Ubiquitous) — This SPEC shall define the role marker's name and value as exported
  constants in `internal/config/envkeys.go` — the variable name exactly `MOAI_FACTORY_ROLE` and the
  role value exactly `worker` — the canonical CLI role spelling; the retired alias `agent` is **not**
  accepted — and the guard and its tests shall read those constants rather
  than a repeated literal. No other card is a precondition of REQ-AP-011 being evaluable: a session
  that sets no such variable makes no role claim, and the tests set it themselves.
- **REQ-AP-013** (Ubiquitous) — The role value the guard expects shall not be an independent
  literal: the role-value constant REQ-AP-012 defines shall be **pinned to the canonical factory role
  vocabulary** — both of its two carriers, the role **token** the launcher parses from `-f <value>`
  (`internal/cli/factory.go`, the live token; its retired alias excluded) and the worker-label
  **prefix** the session-label producer emits (`internal/kanban/bootstrap.go`, the prefix of
  `FactoryLaneLabel`; its two retired aliases excluded) — by a mandatory equality assertion that
  **fails when either side is renamed without the other**. Both carriers are named because they are
  two constants, not one: an assertion naming only the token would pass while the prefix drifted, and
  the reverse. The guard, its tests, the value constant, and both carriers shall therefore carry one
  spelling or the build shall be RED, and no requirement, criterion, or test shall restate that
  spelling as a literal in place of the constant.
  - **Why an assertion and not a reference.** Deriving the value by referring to the carrier constant
    is **not available**: `internal/cli` imports `internal/config` (`factory.go:38`), so
    `internal/config` — where REQ-AP-012 puts the constant — cannot import `internal/cli` back without
    an import cycle, and both carriers are unexported besides. The pinning is therefore an assertion
    **inside each carrier's own package**, where the unexported constant is readable and `internal/config`
    is already importable: one assertion in `internal/cli`, one in `internal/kanban`, no new import
    edge, no export, no shared package. This is stated rather than left implicit because "derive from
    the same source" and "assert equal to the same source" differ in what they guarantee — the
    assertion guarantees the failure, not the derivation, and the failure is what this requirement is
    for. Whether a later card prefers a shared home for the role vocabulary is a design question, not
    a precondition of this requirement.

- **REQ-AP-004** (Unwanted) — When a command carries the word `contract` together with `sign` or
  `decide` but its structure cannot be classified — command substitution, a variable in program
  position, `eval`, an unenumerated wrapper, or nesting deeper than one `-c` level — the guard shall
  not allow it while a deny rule is in force for that call, and shall deny it with the reason marked
  unclassified (fail closed). **Which rule is in force decides the scope, and it is stated rather than
  left to a reader**: an unclassifiable command carrying `sign` is denied in every session, because
  REQ-AP-003 is always in force and an unclassifiable command cannot be shown to be off the human
  path; an unclassifiable command whose only deny-eligible verb is `decide` is denied only where
  REQ-AP-011's role gate is satisfied, and is otherwise allowed, because denying it in a session
  where no deny rule applies would deny the lead on the strength of a parse failure. The guard shall
  deny whether or not the invoked verb is implemented in the installed binary (§C.2, §C.3).
- **REQ-AP-005** (Ubiquitous; finding N4) — The guard's prefix-wrapper set shall be a closed,
  enumerated list containing at least `env`, `command`, `exec`, `nohup`, `script`, `timeout`,
  `sudo`, `stdbuf`, `nice`, and `xargs`, and shall skip each wrapper's own options before reading
  the program word; a wrapper not in the list, encountered in program position, shall be treated
  as unclassifiable under REQ-AP-004 rather than allowed.
- **REQ-AP-006** — **WITHDRAWN at v0.1.1** (receipt-path allowance). Receipt issuance is A3's
  (§C.4), so this SPEC recognizes no receipt path and grants no exemption on the human path. The id is
  retired and is **not** re-used, so a citation of REQ-AP-006 can only refer to the withdrawn meaning
  (§H). Note that REQ-AP-011 is not that allowance returning: it gates the *non-interactive signing
  path A1 defines*, on a role marker, and recognizes no receipt.
- **REQ-AP-009** (Where; §C.7) — Where the guard needs an agent-environment signal beyond the role
  marker of REQ-AP-012 — a harness-presence signal rather than a role claim — it shall read A1's
  closed marker set (`25283ebf8:…/design.md:416-426`: `CLAUDECODE`, `CLAUDE_CODE_SESSION_ID`) and
  shall not introduce a marker of its own beyond REQ-AP-012's. 「A1 plan-audit 통과본으로 재확인」
- **REQ-AP-010** (Ubiquitous; finding N8) — The guard shall be documented, in a `paths:`-scoped rule
  file this SPEC creates together with its `internal/template/templates/` mirror, as a
  mode-independent deny that is active under `guided`; that text shall state that the "nothing changes
  under `guided`" promise of SPEC-AUTONOMY-ESCALATION-001 REQ-AE-001 is scoped to the escalation
  detector rather than to contract-mode work as a whole, and shall carry no card id, SPEC id, internal
  date, or commit SHA (§E C5). The two surfaces this SPEC does not own — the template autonomy block's
  own comment, and the wording of REQ-AE-001 itself — are transferred rather than assumed (§C.10,
  §H.5).

### D.3 Mission-validator projection

- **REQ-AP-008** (Event-driven; O9 / N7 / A1 R8) — When a signed contract must be checked against a
  proposed action, the projection shall map the contract onto `mission.SealedContract` /
  `mission.MissionSnapshot` / `mission.Decision` and obtain the verdict from
  `mission.ValidateMissionDecision` (`internal/mission/policy.go:200`) rather than re-implementing
  its refusal ladder, and shall not widen the mission types to accommodate a contract field. The
  mapping is A1's drafted table (`25283ebf8:…/design.md:470-494`) with the corrections §C.11 measured,
  which this requirement states as obligations rather than notes:
  - it shall populate **every** field `mission.contractComplete` requires, including `Approved`
    (set from the contract's signature state) and a `ResourceLimits.MaxOperations` greater than zero,
    so a mappable contract reaches a verdict instead of being refused `incomplete_contract`;
  - it shall translate a scope glob ending `/**` to its prefix, so the exact-or-prefix containment of
    `targetInsideScope` (`internal/mission/policy.go:112-118`) matches the paths the glob names, and
    shall **neither widen nor narrow** the scope in doing so;
  - it shall carry an enumerated list of contract fields that are **deliberately not projected**,
    `card` among them (A1 REQ-CONTRACT-001), and shall fail closed naming the field for any field that
    is neither projected nor on that list — a glob with an inner wildcard being the case that fails
    closed rather than being approximated;
  - it shall require at least one mission-mappable action, because the sealer refuses an empty
    `AllowedActions` as `incomplete_contract`.

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
- **C4 — Required ordering, and what this SPEC's criteria therefore presume.** A1 `:142-148` states
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
- **C6 — No presumed store, and no second tamper authority.** The moai-owned signing store at
  `$MOAI_HOME/db/<project-key>/contract/` is A3's and A1 does not create it (§C.12), so no requirement
  here reads or presumes it. The contract digest is the sole tamper authority; the projection records
  no `mission_contract_sha256` (A1 `:85`).
- **C7 — The role gate is nominal in production, and that is recorded rather than compensated for.**
  Until card t1240 stamps `MOAI_FACTORY_ROLE`, REQ-AP-011 denies nothing in a real session; after it
  does, an agent can unset the variable (A1 `:155-156`). REQ-AP-011's criteria are green because the
  tests set the variable themselves (REQ-AP-012) — which is honest about what they measure: the
  guard's behavior given the marker, not the marker's presence in production. The requirement that
  does not depend on the marker at all is REQ-AP-003, which is the one protecting the signature.

## §F — Open items

Each entry states its **disposition** and, where anything remains open, what specifically would close
it. No entry gates a criterion's evaluability.

- **O1** [RESOLVED — `decide` is in scope, guarded by the role-scoped rule REQ-AP-011, and its owner
  is named.] The v0.1.1 disposition ("denied outright") is superseded by the three-way split. The
  verb's owner is **A3 (card t1236)**, which A1 `:90-92` gives the kickoff decision rules this verb
  would carry (§C.3) — so the earlier "owner unidentified" reading is closed by measurement, not left
  open. The verb is still unimplemented in this tree, which the shape-based deny does not need
  (AC-AP-008).
- **O2** [RESOLVED as to scope; one residual exposure recorded, not open.] The lead's three-way
  ruling is adopted in full (§C.3): the boundary for the human path, the `MOAI_FACTORY_ROLE=worker`
  role gate for the non-interactive path and `decide`, and an allow for a session carrying no marker.
  The role distinction the earlier O2 asked about is therefore named and owned — this SPEC defines the
  variable (REQ-AP-012). What remains is not a question but a **residual exposure** recorded in §E C7:
  the role gate is nominal in production until t1240 stamps the marker, and unset-able afterwards.
- **O5** [RESOLVED — operator ruling at the Kickoff gate: the value is `worker`.] What the question
  was: an earlier ruling named `agent`, and in `internal/cli/factory.go` `worker` (`:59`) is the live
  role token while `agent` (`:64`) is its **retired** pre-rename spelling, kept only as a parsing
  alias. Keying the guard on the retired spelling had a concrete failure mode — **if card t1240
  stamps the live spelling, REQ-AP-011 denies nothing and AC-AP-016's allow arm still passes**, so the
  guard would read as healthy while protecting nothing. The operator closed it on exactly that ground:
  the canonical spelling `worker` is the accepted value, and `agent` is not accepted. REQ-AP-012's
  value constant, AC-AP-015's fixture, and the value card t1240 must stamp (§G) all read `worker`.
- **O6** [DEFERRED to card **t1256**'s SPEC by design — deferred, not undecided.] The literal
  spelling of the role value. **Current state, measured:** `worker`, and it is carried independently by
  two constants at develop `553e224f3` — `internal/cli/factory.go:59` (`factoryWorkerRoleToken`, the
  role token) and `internal/kanban/bootstrap.go:247` (`factoryLaneRole`, the label prefix). Card t1256
  renames that vocabulary to `lane`, which is the **older** spelling returning rather than a new one:
  `origin/main` — an ancestor of develop, so simply behind it — still carries `factoryLaneRole = "lane"`
  and no role-token constant at all, so a reader who greps `main`, finds `lane`, and concludes the
  rename has already landed would be reading a stale tree rather than a completed change. This SPEC
  therefore fixes the **binding** (REQ-AP-013, AC-AP-018) and not the string: whatever t1256 makes
  canonical, the guard's constant follows it or the build is RED. What would close O6: t1256's SPEC
  naming the canonical value. No criterion here waits on it — every occurrence of `worker` in this
  document is the current measurement, not a ruling t1256 must honour.

- **O3** Whether the audit line of REQ-AP-002 shares a sink with the existing
  branch-guard audit log (`.moai/logs/branch-guard-audit.log`) or takes its own. Design default:
  its own, named in `design.md`; a shared sink is acceptable if the lead prefers one file.
- **O4** [RESOLVED — lead ruling 2026-09-26.] The SPEC body stays in **English**. The reason is
  operational rather than stylistic: this SPEC cites requirements and criteria across the AUTONOMY
  family (A1, A2, A3), those SPECs are English technical prose, and cross-SPEC comparison by reading
  and by grep needs one language. Korean appears only in the
  「A1 plan-audit 통과본으로 재확인」 dependency tag, as in the sibling SPEC.

## §G — Exclusions

### Out of Scope — Receipt issuance and receipt-path recognition (A3, card t1236)

- Issuing kickoff receipts from moai itself — A1 `:94-98` gives it to A3. This SPEC recognizes **no**
  receipt path and grants **no** exemption, because a criterion asserting a receipt invocation is
  allowed could only turn green after A3 (§C.4, finding N5).
- Validating a receipt's contents — A1's validator owns it (REQ-CONTRACT-023).
- The transferred criterion `AC-AE-025` — excluded, not carried (§H).

### Out of Scope — Verbs and surfaces this SPEC does not guard

- `moai contract revoke` — declared out of scope by A1 at `:89`, owned by A3.
- **Defining** the `decide` verb — A3's (§C.3, A1 `:90-92`). This SPEC guards a command-line shape; it
  does not create the command, its flags, or its decision rules.
- The **non-interactive signing path itself** — A1 (REQ-CONTRACT-024). REQ-AP-011 gates who may invoke
  it as a tool call; it does not define or alter the path.
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
  per A1 `:144-147` (C4).
- **Stamping `MOAI_FACTORY_ROLE` on a real session** — card t1240. This SPEC defines the constants
  (REQ-AP-012) so its criteria are evaluable now; it does not wire the variable into any launcher, and
  §E C7 records what that leaves un-protected in production. **The value t1240 must stamp is `worker`**
  (§F O5, operator ruling) — the canonical CLI role spelling; stamping the retired alias `agent` would
  leave REQ-AP-011 denying nothing while every criterion here still passes.
- The **documentation surfaces this SPEC does not own**: the template autonomy block's own comment
  (A1, REQ-CONTRACT-016) and the wording of SPEC-AUTONOMY-ESCALATION-001 REQ-AE-001 (A2). §C.10
  measures both absent and §H.5 records the transfer.
- The moai-owned signing store at `$MOAI_HOME/db/<project-key>/contract/` — A3 (§C.12, C6).
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

**How this section was produced, because its first version was wrong in a way its content hid.**
Every row below was read from the **pre-split blob** `d8926ff9a` and from a **named commit** on the
source branch, with `git show <sha>:<path>`. v0.1.1 reconstructed the table from the source SPEC's §K
summary, which lists ids without criterion bodies; the content came out right while three
attributions came out wrong, which is the worst combination — a reader who checks the ref finds no
source text and doubts the whole table, and a reader who finds the table plausible never checks.

### H.1 ID reuse record (source side, t1235)

**The criterion ids `AC-AE-022`..`AC-AE-025` were re-used after the split**, so a citation of them
means different things at different commits. Both sides are measured:

| Read at | `AC-AE-022` maps | `AC-AE-023` maps | `AC-AE-024` maps | `AC-AE-025` maps |
|---|---|---|---|---|
| `d8926ff9a` (v0.2.1, pre-split — the transferred meanings) | `REQ-AE-024` (push serializer) | `REQ-AE-024` (push serializer) | `REQ-AE-025` (sign deny) | `REQ-AE-025` (receipt path) |
| `78a95ee03` (v0.4.2, read 2026-09-26 — the current meanings) | `REQ-AE-018`, `REQ-AE-019` | `REQ-AE-019`, `REQ-AE-020` | `REQ-AE-022` (ownership decode) | `REQ-AE-023` (signed-invalid) |

Any citation of `AC-AE-022`..`AC-AE-025` in a plan-audit report predating t1235 v0.3.0 refers to the
**withdrawn** meanings of the first row. This SPEC therefore carries its own id namespace and never an
`AE` id outside this section.

**The current-meanings row is a moving claim, and has already moved twice.** Its subject is what that
branch *now* carries, so it is not reducible to a fixed SHA without changing what the sentence says —
but a value stated without its measuring command decays silently. Both are therefore given: the
command, and the value as a dated reference.

<!-- moving-ref-ok: the claim's subject is what the sibling's branch currently maps, which a pinned SHA would narrow to what one commit mapped; the deciding command is stated alongside, and the value is dated -->
Measure with `git show WT-escalation-detector:.moai/specs/SPEC-AUTONOMY-ESCALATION-001/acceptance.md | grep -E '^- \*\*AC-AE-02[2-5]'`.
Reference value, read 2026-09-26 at `78a95ee03` (`version: "0.4.2"`): the second row above. The prior
revision of this SPEC stated `AC-AE-022`→`REQ-AE-022` and `AC-AE-023`→`REQ-AE-023`, which was one row
out of step at the time and is two revisions stale now — which is the reason the command comes first.

### H.2 ID retirement record (this SPEC)

**`REQ-AP-006` was withdrawn at v0.1.1** (receipt-path allowance) and is **not** re-used; a citation
of it can only refer to the withdrawn meaning. The criteria `AC-AP-011` and `AC-AP-012`, which
mapped it, were removed at the same revision and their ids are likewise retired, not renumbered
into other criteria. No id retired here is re-used by v0.1.2's three new requirements
(REQ-AP-011, REQ-AP-012) or three new criteria (AC-AP-015, AC-AP-016, AC-AP-017), nor by v0.1.3's one
new requirement (REQ-AP-013) and one new criterion (AC-AP-018), all of which continue the numbering
instead of filling the gap.

### H.3 Provenance of carried items

Source ids are the **v0.2.1 meanings at `d8926ff9a`** (H.1, first row), read from that blob.

| This SPEC | Source id (v0.2.1, `d8926ff9a`) | Carried |
|---|---|---|
| REQ-AP-001, REQ-AP-002, REQ-AP-007 | REQ-AE-024 | text, split by fail direction and by record silence; the field renamed `push_requires_lease` and read from `show --json` (§C.5, §C.6) |
| REQ-AP-003, REQ-AP-004, REQ-AP-009 | REQ-AE-025 | text, split by recognition / fail-closed / marker; **receipt limb dropped** (H.4); at v0.1.2 REQ-AP-003 narrowed to the human signing path |
| REQ-AP-011, REQ-AP-012 | — (lead ruling 09-26 (3)) | new: the role-scoped deny of the non-interactive path and `decide`, and the marker constants that make it evaluable. Not a revival of REQ-AP-006 (H.2) |
| REQ-AP-005 | — (finding N4) | new requirement; the finding had no requirement |
| REQ-AP-010 | — (finding N8) | new requirement; narrowed at v0.1.2 to a surface this SPEC creates (§C.10) |
| REQ-AP-013, AC-AP-018 | — (lead ruling 09-26, rename-immunity) | new at v0.1.3: pin the role value to both carriers of the canonical vocabulary by an equality assertion, so the pending t1256 rename cannot leave the guard keyed on a dead spelling (§F O6) |
| REQ-AP-008 | O9 / N7, and A1 R8 (`25283ebf8:…/design.md:470-494`) | the projection item, which had no owner; A1's drafted mapping adopted with the three corrections of §C.11 |
| AC-AP-001, AC-AP-002 | AC-AE-022 | split: deny path, and the unaffected-first-push / off-condition controls |
| AC-AP-003, AC-AP-004 | AC-AE-023 | split: release, and stale/expired reclaim + fail-open |
| AC-AP-005, AC-AP-006 | AC-AE-024 | split: bypass shapes (human path only from v0.1.2), and positive controls |
| AC-AP-007, AC-AP-008 | — | new: the fail-closed unclassified path and the not-implemented-verb case |
| AC-AP-009, AC-AP-010 | — (finding N4) | new: the wrapper set, and the unknown-wrapper case |
| AC-AP-013 | — (finding N8) | new: the documentation surface. Retargeted at v0.1.2 (§C.10) |
| AC-AP-014 | — (O9 / N7, A1 R8) | new: the projection |
| AC-AP-015, AC-AP-016, AC-AP-017 | — (lead ruling 09-26 (3)) | new: the role-scoped deny arm, its allow arm with an armed control, and the marker constants |
| `AC-AP-011`, `AC-AP-012` | AC-AE-025 | **retired, not carried** — see H.2 and H.4. The gap at 011-012 is that retirement, not a numbering slip |
| Bypass-shape table | spec.md §J | extended with the N4 wrapper rows |
| Mechanics | design.md §G.1, §G.2 (`d8926ff9a`) | this SPEC's `design.md` §B, §C |
| Milestone | plan.md M6 (`d8926ff9a`) | split into M1-M4 |

### H.4 Transferred but **excluded** — not carried here

| Source id (`d8926ff9a`) | Disposition |
|---|---|
| `AC-AE-025` (receipt path allowed / forged receipt denied) | **Not transferred.** Owned by **A3 (card t1236)**, which owns receipt issuance (A1 `:94-98`). Reason: **N5** — its Given is "A1 has defined the moai-issued receipt path", so it could only turn green after A3, and carrying it would make t1245 uncompletable the same way t1235 was. Lead decision; not an open question. REQ-AP-011 does not revive it: that requirement gates *who may invoke* the non-interactive path and recognizes no receipt. |

### H.5 Transferred **out** at v0.1.2 — obligations this SPEC hands to their surface owners

The audit's D1 found `AC-AP-013` conditional on A1's template block. §C.10 measured that neither
surface v0.1.1 named exists, so the repair is a transfer of the two limbs whose surfaces are not this
SPEC's, not a rewording.

| Obligation | Transferred to | Why it cannot be evaluated here |
|---|---|---|
| Stating the mode-independent deny in the **template `workflow.yaml` autonomy block's comment** | **A1 (t1234)**, which creates the block (REQ-CONTRACT-016) | The block does not exist (§C.10); a criterion over its comment could only turn green after A1 — finding N5 |
| Re-scoping the sentence "nothing changes under `guided`" | **A2 (t1235)**, which wrote it (REQ-AE-001) | It is that SPEC's own requirement text; this SPEC cannot amend another SPEC's requirement, and a criterion asserting it had been amended would be conditional on A2 |

What remains with REQ-AP-010 is the rule file this SPEC creates and its template mirror — present in
the same change as the guard, so AC-AP-013 is evaluable at M3 with no dependency on any other track.

## §I — Cross-references

- SPEC-AUTONOMY-CONTRACT-001 (A1, card t1234) — contract schema, `verify`, the receipt format and
  validator, `push_requires_lease`, `show --json`, the two signing paths, the agent-environment marker
  set, the drafted mission mapping, and the required ordering. **Every citation in this document is
  pinned to commit `25283ebf8`** (that SPEC's v0.5.2, verified at that blob's line 4).
  <!-- moving-ref-ok: this sentence's subject is the citation rule itself — that a branch name moves while a SHA does not — so substituting a SHA for the branch name would make the sentence demonstrating the rule violate it -->
  The branch `WT-contract-schema` moves, a pinned SHA does not.
- SPEC-AUTONOMY-CONTRACT-001 §§ owned by A3 (card t1236) — receipt issuance, revocation, and the
  signing-event store; the receipt-path criterion this SPEC excludes (§H.4).
<!-- moving-ref-ok: an address saying which branch the sibling card's work lives on, so a reader can find it; it is not an evidence coordinate and nothing below is measured on the strength of it -->
- SPEC-AUTONOMY-ESCALATION-001 (A2, card t1235, on branch `WT-escalation-detector`) — the escalation
  detector and the contract resolver; §K of its spec.md is this SPEC's transfer record. Transferred
  **text** is cited at `d8926ff9a` and never at this branch name (§H).
- `.claude/rules/moai/workflow/resource-slot-lease.md` — the slot lease's liveness, bound, and
  takeover rules.
- `CLAUDE.local.md` §4.1 — why a cancelled CI run on `develop` is a lost verdict.
