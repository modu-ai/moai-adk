# SPEC-AUTONOMY-PRECONDITION-001 — Design

> Two deny components — one keyed on the tool-call boundary, one on a role marker — plus one reuse
> projection. They share the contract resolver (SPEC-AUTONOMY-ESCALATION-001 REQ-AE-002) and nothing
> else. Mechanics carried from
> `d8926ff9a:.moai/specs/SPEC-AUTONOMY-ESCALATION-001/design.md` §G — the pre-split commit, cited by
<!-- moving-ref-ok: the ref is the SUBJECT of a recorded correction, not an address anything is read from; substituting a SHA would erase which citation was wrong -->
> SHA rather than by the relative ref `WT-escalation-detector~1` that v0.1.1 used and that has since
> slid past the split — with the N4 wrapper rows folded in, R6 applied (`push_requires_lease`, read
> from `show --json`), and R5 applied as a **scope removal** — no receipt exemption on the human path
> (§C.5).
>
> Every A1 citation here is pinned to commit **`25283ebf8`** (A1 SPEC v0.5.2), not to a branch.

## §A — Why this is a design decision and not a mechanical change

Two questions have more than one defensible answer, which is what makes this card Class C:

1. **How the PreToolUse deny recognizes bypass shapes.** The existing guards scrub quoted spans
   (`substituteQuotedArguments`, `internal/hook/branch_guard.go:197`), which is correct when the
   goal is to *ignore* quoted text. Here the goal is the opposite — `'moai' contract "sign"` must be
   recognized — so scrubbing would defeat the matcher. The decision is to **remove quotes and split
   into words**, then classify, and to treat anything unclassifiable as denied.
2. **How the contract projects onto the mission validator.** `internal/mission` is a sealed
   fail-closed validator with its own types. The decision is a **one-way projection that fails
   closed on an unmapped field**, rather than widening the mission types.

## §B — Push serializer (REQ-AP-001, REQ-AP-002, REQ-AP-007)

- **Resource.** The existing slot lease, resource name `push-develop` — a valid slot resource name
  under `kanban.ValidateSlotResourceName` (`internal/cli/slot.go:114`). Record read with
  `kanban.ReadSlotLease(root, "push-develop")`, root resolved the way `internal/cli/slot.go`
  resolves it. No new record format, lock, or verb. `moai slot status --resource push-develop`
  reads the same record a lane would read by hand.
- **Activation.** Only under `workflow.autonomy.mode: contract` **and** `push-develop` present in
  `actions` **and** `push_requires_lease: true`, all three read from **`moai contract show --json`**
  — A1's declared contact point for exactly this purpose (`25283ebf8:…/spec.md:410`: "the verifier
  shall expose this as a derived `push_requires_lease: true` field in `show --json` output for A2b
  to enforce"). Reading the JSON rather than parsing `contract.yaml` keeps the derivation A1's, so a
  schema change reaches this guard through A1's projection instead of through a second parser.
  Measured: that command does not exist yet (spec.md §C.6) — A1 supplies it. Absent field or absent
  command → inactive, which is AC-AP-002 condition (d). It does **not** depend on
  `workflow.slot_lease.enabled`, which keeps gating the generic slot guard only.
- **Matcher.** A Bash command whose program is `git`, whose subcommand is `push`, and whose refspec
  targets `develop`, with the same quote handling `checkSlotLease` already applies
  (`internal/hook/slot_lease_guard.go`).
- **Admit.** Free, expired (declared bound elapsed), stale (owning session gone), or held by the
  calling session → admit, and write the lease for the calling session with bound
  `workflow.slot_lease.default_max_duration`. A takeover names the displaced holder, which is the
  slot lease's existing rule, not a new one.
- **Deny.** Held by a different live session within its bound → deny with
  `PUSH_SERIALIZATION_VIOLATION: push-develop held by <holder>`. The lane waits and retries. No
  escalation record: serialization is expected traffic, not a contract breach (REQ-AP-007).
- **Release.** PostToolUse on the admitted push. A non-zero exit releases immediately — nothing is
  in flight. Otherwise the holder releases with `moai slot release --resource push-develop` after
  reading the CI result for the pushed head; a forgotten release costs at most the bound.
- **Fail direction: open.** Unreadable record, unresolvable root, or unknown caller → allow plus one
  audit line. An overlapping push costs one cancelled CI run; a stuck deny halts the lane.

## §C — Contract-sign and contract-decide guard (REQ-AP-003, REQ-AP-004, REQ-AP-005, REQ-AP-009 .. REQ-AP-012)

### C.1 Placement, and the one decision that changed at v0.1.2

A PreToolUse Bash check independent of `workflow.autonomy.mode`, placed with the other deny guards.

**The predicate is not one thing.** v0.1.1 used the tool-call boundary for everything, reasoning that
every Bash tool call is agent-issued and that a boundary cannot be unset while a variable can. The
first half is still true and is why REQ-AP-003 keys on the boundary: human signing happens at an
operator terminal, so a human-path `sign` arriving as a tool call is wrong by construction. The second
half made the conclusion too strong — **`moai contract decide` has a legitimate tool-call caller**,
the lead session's own LLM when the decider is `llm` or `llm+jev` (spec.md §C.3), so a boundary-keyed
deny of `decide` is not "stronger", it is a false deny of the path the epic depends on. The guard
therefore reads a role marker for the non-interactive sign path and for `decide` (REQ-AP-011), and
this SPEC defines that marker's constants itself so the branch is reachable (REQ-AP-012). Where a
harness-presence signal rather than a role claim is needed, A1's closed set (`CLAUDECODE`,
`CLAUDE_CODE_SESSION_ID`) is read (REQ-AP-009).

### C.2 Parse

1. Split the command into shell words with **quote removal** — not quoted-span scrubbing, which
   would erase `'moai'`.
2. Strip leading `NAME=value` assignments.
3. Strip prefix wrappers from the closed list of §C.3, skipping each wrapper's own options.
4. Take the program word's **basename**; skip global flags before the verb; match
   `moai` + `contract` + (`sign` | `decide`).
5. For `sh|bash|zsh -c <string>`, parse `<string>` once more. Exactly one level.
6. **Select the rule** (this step is new at v0.1.2 and is the only place the three rules diverge):
   `sign` with no `--signer` or `--signer human` → REQ-AP-003, deny unconditionally. `sign` with
   `--signer llm` / `--signer llm+jev`, or `decide` → REQ-AP-011, deny iff the calling session's
   `MOAI_FACTORY_ROLE` equals the value constant, otherwise allow. The `--signer` value is read from
   the same word list, so no separate parser exists for it.

### C.3 Prefix-wrapper set (finding N4)

A closed, enumerated list. Each row states what the wrapper's own options look like, because
skipping them wrongly is how a wrapper bypass survives a matcher that claims to handle it.

| Wrapper | Own options to skip before the program word |
|---|---|
| `env` | `NAME=value` pairs, `-i`, `-u NAME` |
| `command`, `exec` | `-p`, `-v`, `-V` (`command`); `-a NAME`, `-c`, `-l` (`exec`) |
| `nohup` | none |
| `script` | `-q`, `-c CMD`, and the typescript file operand |
| `timeout` | `-k DURATION`, `-s SIG`, `--preserve-status`, and the leading duration operand |
| `sudo` | `-n`, `-u USER`, `-E`, `-H`, `--` |
| `stdbuf` | `-i`, `-o`, `-e` with their MODE arguments |
| `nice` | `-n N` |
| `xargs` | `-n N`, `-0`, `-I STR`, `-P N` |

A wrapper **not** in this list, found in program position, is unclassifiable → denied (REQ-AP-004,
AC-AP-010). `script -c 'moai contract sign'` and `timeout`'s duration operand are the two rows most
likely to be mis-skipped, which is why the operand column is explicit.

### C.4 Unclassifiable → deny (fail closed)

Command substitution, a variable in program position, `eval`, an unknown wrapper, or nesting beyond
one `-c` level — when `contract` occurs together with `sign` or `decide` → deny, reason marked
`unclassified`. A *classified* invocation whose program is not `moai` (`echo "moai contract sign"`,
`git commit -m "sign the contract"`) is allowed.

The deny does not depend on the verb existing in the installed binary (spec.md §C.2, §C.3): the
point is that the verb is agent-unreachable from the moment any track lands it. This matters
doubly for `decide`, which **no track currently defines** — the guard is forward-looking by
construction. AC-AP-008 asserts the unimplemented precondition first so the case cannot pass merely
because the command would have failed anyway.

### C.4a Which predicate applies to which invocation, and why fail-closed is rule-scoped

| Invocation | Predicate | Rule | Rationale |
|---|---|---|---|
| `sign`, human path (no `--signer`, `--signer human`) | the tool-call boundary | REQ-AP-003 | human signing happens at a terminal; the boundary is unset-able by nobody |
| `sign --signer llm` / `llm+jev` | `MOAI_FACTORY_ROLE == agent` | REQ-AP-011 | A1 leaves this path ungated by any marker (`25283ebf8:…/design.md:419-421`), and it has an automated caller by design |
| `decide` | `MOAI_FACTORY_ROLE == agent` | REQ-AP-011 | the lead's own LLM invokes it as a tool call; a boundary deny would deny the lead |

`MOAI_FACTORY_ROLE` does not exist in this tree, which is why v0.1.1 avoided it. That is solved by
**defining it here** (REQ-AP-012) rather than by substituting a predicate: the tests set the variable
and the branch is reachable without card t1240.

**Fail-closed is scoped to the rule in force**, which the parse cannot ignore. An unclassifiable
command carrying `sign` is denied everywhere, because REQ-AP-003 is always in force and an
unclassifiable command cannot be shown to be off the human path. An unclassifiable command whose only
deny-eligible verb is `decide` is denied only under the role marker — denying it in a session where no
rule applies would deny the lead on the strength of a parse failure, trading a real path for a
hypothetical one.

**What the role marker does not buy, stated here so it is not read as protection it is not.** An agent
can unset a variable (A1 `:155-156`), and until t1240 stamps it nothing carries it. REQ-AP-011 is
therefore nominal in production in both directions; the protection that is not nominal is REQ-AP-003,
which is keyed on the boundary. AC-AP-016's armed human-path limb is what keeps the allow arm from
passing on an absent guard.

### C.5 No receipt exemption (R5, resolved by removing scope)

The guard recognizes **no** receipt path and grants **no** exemption: it never reads `--receipt`.
Every **human-path** `sign` invocation is denied whatever a `--receipt` argument says (REQ-AP-003).
The non-interactive path and `decide` are gated on the role marker, not on a receipt (REQ-AP-011) —
the `--signer` value selects which rule applies, and `--receipt` travels through as an opaque
argument the guard does not inspect.

This is a deliberate narrowing, decided by the lead, and it makes the guard *simpler and more
testable* rather than less capable:

- Receipt **issuance** is A3's (`25283ebf8:…/spec.md:94-98`). A criterion asserting a receipt
  invocation is allowed could only turn green after A3 — the shape of finding **N5**, which is why
  this card exists at all. An exemption here would re-create the uncompletable SPEC.
- With no exemption, the guard has **no dependency on any A1 or A3 interface**: it reads a command
  line plus one environment variable this SPEC defines itself (REQ-AP-012), and answers. Nothing in §C
  needs A1 or card t1240 to land.
- The absence is asserted by AC-AP-005's fixture set carrying no exempt form, so an implementation
  that quietly adds a receipt branch fails that criterion rather than passing unnoticed.

A1's own §C.8 records that **while A3's amendment of the display-only principle has not landed**, any
receipt whose effective decider is `llm+jev` is refused `receipt_requires_human`; once A3 lands that
amendment, A3 removes the refusal (A1 v0.5.2, `25283ebf8:…/spec.md:231-237`, which restated the rule
conditionally for exactly this reason). The decider set is `human | llm | llm+jev`, so a plain `llm`
receipt is **not** covered by that interim refusal — which is why REQ-AP-011 gates the non-interactive
path here rather than relying on A1's refusal to cover it.

### C.6 Fail directions are opposite on purpose

§B fails open; §C fails closed. A wrongly allowed signature voids the contract model; a wrongly
denied sign costs the operator one terminal command, which is where signing belongs anyway. A
wrongly denied push halts a lane; a wrongly allowed one costs a cancelled CI run.

### C.6a Where the deny is documented, and why the surface moved (finding N8, audit D1)

v0.1.1 pointed the documentation obligation at A1's template `workflow.yaml` autonomy block. Measured
at `eee5f635e`, that block does not exist — and neither does any rule text stating contract-mode
behavior (spec.md §C.10). The obligation had no surface this SPEC could write, which made `AC-AP-013`
green only after A1: a recurrence of finding N5.

The surface is therefore one this change creates: a `paths:`-scoped rule file documenting this guard,
plus its `internal/template/templates/` mirror (Template-First). `paths:`-scoped rather than
always-loaded — a guard's documentation does not earn a place in every session's always-loaded budget.
The two limbs whose surfaces belong elsewhere are transferred rather than assumed: the template block's
own comment to A1, the wording of REQ-AE-001 to A2 (spec.md §H.5).

### C.7 The `guided` promise (finding N8)

SPEC-AUTONOMY-ESCALATION-001 REQ-AE-001 promises that nothing changes under `guided`. This guard
denies under `guided` too, because a contract signed by an agent is worthless in any mode. The
contradiction is resolved in **wording, not behavior**: the promise is scoped to the escalation
detector, and both the template comment and the rule text say the sign deny is mode-independent
(REQ-AP-010, AC-AP-013) — on the surfaces §C.6a names. Weakening the guard to preserve the sentence
would trade a real protection for a phrasing.

## §D — Mission-validator projection (REQ-AP-008)

- **Target.** `mission.ValidateMissionDecision` (`internal/mission/policy.go:200`), verified present
  with the doc comment "treats every Decision field as untrusted data. It never calls a tool or
  performs a state change".
- **Direction.** One way: contract → `mission.SealedContract` / `mission.MissionSnapshot` /
  `mission.Decision`. The verdict comes back from mission. Its five refusal reasons
  (`mission_not_running`, `mission_mismatch`, `policy_mismatch`, `stale_snapshot`,
  `expired_decision`) are **not** re-implemented.
- **The mapping, adopted from A1's forward note with three corrections.** A1 drafted the table at
  `25283ebf8:…/design.md:470-494` and left it for this SPEC "to adopt or revise". It is adopted; the
  corrections are the three §C.11 measured. The 14th row and the two correction columns are this
  SPEC's:

  | `mission.MissionContract` field | Source | Correction |
  |---|---|---|
  | `MissionID` | `spec_id` | — |
  | `Goal` | `approach` | — |
  | `CompletionEvidence` | `["acceptance:" + acceptance.sha256, "ac_count:" + N]` | — |
  | `Scope` | `ownership.write` | **trailing `/**` → prefix**; an inner-wildcard glob is unmappable → fail closed (`targetInsideScope` is exact-or-prefix, `internal/mission/policy.go:112-118`) |
  | `AllowedActions` | mapped allowed actions (A1 Action Vocabulary) | **at least one** mission-mappable action required (`contractComplete` refuses an empty list as `incomplete_contract`) |
  | `MergeTarget` | `develop` | — |
  | `ResourceLimits` | `{MaxOperations: budget.operations, MaxRetries: budget.audit_retries}` | **`MaxOperations` must be > 0** (`internal/mission/contract.go:58`); zero or absent is unmappable, never mapped to zero |
  | `ProhibitedActions` | `main_merge, force_push, release_branch, release_pr` | — |
  | `StopConditions` | `escalate_on` | — |
  | `RecoveryConditions` | `["sign --resign after acceptance change"]` | — |
  | `RevocationBehavior` | `"stop-before-next-action"` | — |
  | `PolicyVersion` | `"contract-v1"` | — |
  | **`Approved`** | the contract's signature state (`signed-valid` → true) | **absent from A1's draft.** `contractComplete` requires it (`internal/mission/contract.go:60`), so a projection built from the drafted 13 rows alone is refused `incomplete_contract` and every verdict fails closed |
  | *(not projected)* | `card` (A1 REQ-CONTRACT-001) | enumerated as **deliberately not projected** — a bare unmapped-field rule would fail every contract on it, since `card` is required and has no mission counterpart |

  `mission_contract_sha256` is not recorded — A1 `:85` keeps the contract digest as the sole tamper
  authority, so the projection does not add a second one.
- **Unmapped field → fail closed.** A contract field with no mission counterpart returns an error
  naming the field. The alternative — widening mission's exported types — is rejected: mission is a
  sealed validator, and a type widened for a second caller stops being one thing validated one way.
  AC-AP-014 asserts mission's exported surface is unchanged.
- **Why here.** The item (O9 / N7) had no owner; it is a *reuse* decision about the same
  fail-closed logic these two guards rely on, so it belongs with them rather than in a track that
  only consumes it.

## §E — Packages

Both guards in `internal/hook`, beside the branch and integration-lock guards whose parsing helpers
they reuse. The projection in `internal/contract` (A1's package) or a sibling — the choice is
run-phase's, recorded in progress.md §E.2 if it differs, because the test name in AC-AP-014 binds to
the package.

## §F — Residual risks

| Risk | Statement |
|---|---|
| A script file, alias, or shell function hides the sign invocation | Out of scope (spec.md §G) — the hook sees a path, not contents. `go run ./cmd/moai contract sign` is caught by the fail-closed rule only because both words occur |
| A wrapper outside the §C.3 list is added to a lane's habits | It is denied as unclassified, so the failure mode is a false deny, not a false allow — the direction this guard should err in |
| The holder never releases the `push-develop` lease | Bounded by `workflow.slot_lease.default_max_duration`; the cost is one wait, not a permanent stall |
| Non-interactive signing has no path while this guard denies unconditionally | Accepted: signing belongs at an operator terminal, and A3 owns receipt issuance. If A3 needs an exemption it amends this SPEC's REQ-AP-003 rather than adding an unowned branch here |
| `moai contract show --json` may not exist or may rename the field when M1 is reached | Absent field or command → inactive (AC-AP-002 (d)); pre-flight re-measures and stops on a rename (plan.md §C step 3) |
| No track defines `decide`, so its deny is untested against a real verb | The deny is shape-based and evaluable now (AC-AP-008); when a track defines the verb, its own criteria confirm reachability |
| A full keyless re-seal of a contract | A1's residual risk, restated here only because this guard does not close it either |
| `MOAI_FACTORY_ROLE` is never stamped, or is stamped with a different value | REQ-AP-011 denies nothing in production. Recorded, not compensated for: the protection that does not depend on the marker is REQ-AP-003 (the human path, keyed on the boundary). The value-spelling hazard is concrete — `agent` is the **retired** CLI role spelling, `worker` is the live one (`internal/cli/factory.go:59`, `:64`) — and sits with the lead as spec.md §F O5 |
| An agent unsets `MOAI_FACTORY_ROLE` before invoking `decide` | Accepted and recorded (A1 `:155-156` says an agent can unset a variable). The boundary-keyed human-path deny is unaffected; a role-keyed rule cannot be made unset-proof, and pretending otherwise would be the claim this table exists to avoid |
| A harness populates the hook process's environment differently from the session's | **Unmeasured.** The guard reads the calling session's environment at PreToolUse, so a wrapper's own env stripping (`sudo`, `env -i`) is irrelevant — but whether every harness propagates identically was not measured. Recorded as a Gap in progress.md §E.1 rather than assumed either way |
| A1's drafted mission mapping changes after this SPEC adopted it | The three corrections of §C.11 are measured against `internal/mission` source, not against A1's prose, so they survive a prose change. A changed *field source* is a pre-flight re-measure (plan.md §C) |
