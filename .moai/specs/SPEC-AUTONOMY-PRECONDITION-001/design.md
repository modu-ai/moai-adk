# SPEC-AUTONOMY-PRECONDITION-001 — Design

> Two independent deny components plus one reuse projection. They share the contract resolver
> (SPEC-AUTONOMY-ESCALATION-001 REQ-AE-002) and nothing else. Mechanics carried from
> `WT-escalation-detector~1:.moai/specs/SPEC-AUTONOMY-ESCALATION-001/design.md` §G, with the N4
> wrapper rows folded in, R6 applied (`push_requires_lease`, read from `show --json`), and R5
> applied as a **scope removal** — no receipt exemption (§C.5).
>
> Every A1 citation here is pinned to commit **`67a2f55cb`** (A1 SPEC v0.5.0), not to the moving
> branch.

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
  — A1's declared contact point for exactly this purpose (`67a2f55cb:…/spec.md:382`: "the verifier
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

## §C — Contract-sign guard (REQ-AP-003 .. REQ-AP-006, REQ-AP-009, REQ-AP-010)

### C.1 Placement

A PreToolUse Bash check independent of `workflow.autonomy.mode`, placed with the other deny guards.
Every Bash tool call is by construction agent-issued, so the boundary itself carries the
agent-vs-human distinction — which is why the guard needs no role variable, and why
`MOAI_FACTORY_ROLE` (which does not exist, spec.md §C.6) is not consulted. Where a marker is
genuinely needed, A1's closed set (`CLAUDECODE`, `CLAUDE_CODE_SESSION_ID`) is read (REQ-AP-009).

### C.2 Parse

1. Split the command into shell words with **quote removal** — not quoted-span scrubbing, which
   would erase `'moai'`.
2. Strip leading `NAME=value` assignments.
3. Strip prefix wrappers from the closed list of §C.3, skipping each wrapper's own options.
4. Take the program word's **basename**; skip global flags before the verb; match
   `moai` + `contract` + (`sign` | `decide`).
5. For `sh|bash|zsh -c <string>`, parse `<string>` once more. Exactly one level.

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

### C.4a The predicate is the boundary, not a role variable

The lead's ruling is an outright deny for agent-role sessions. `MOAI_FACTORY_ROLE` does not exist
(spec.md §C.7), so the guard does not read it — it denies on **every** Bash tool call, which covers
the ruling completely rather than partially: an agent can unset a variable (A1 says so at
`67a2f55cb:…/spec.md:149-150`) and cannot unset the tool-call boundary. AC-AP-005's Given pins the
variable as absent, so an implementation that gated on it fails rather than passes empty.

### C.5 No receipt exemption (R5, resolved by removing scope)

The guard recognizes **no** receipt path and grants **no** exemption. Every `sign` / `decide`
invocation is denied, whatever `--signer` or `--receipt` say.

This is a deliberate narrowing, decided by the lead, and it makes the guard *simpler and more
testable* rather than less capable:

- Receipt **issuance** is A3's (`67a2f55cb:…/spec.md:85`). A criterion asserting a receipt
  invocation is allowed could only turn green after A3 — the shape of finding **N5**, which is why
  this card exists at all. An exemption here would re-create the uncompletable SPEC.
- With no exemption, the guard has **no dependency on any A1 or A3 interface**: it reads a command
  line and answers. Nothing in §C needs A1 to land.
- The absence is asserted by AC-AP-005's fixture set carrying no exempt form, so an implementation
  that quietly adds a receipt branch fails that criterion rather than passing unnoticed.

A1's own §C.8 records that any Jev decision routes to a human (`receipt_requires_human`), so no
`jev` / `llm+jev` receipt can sign in A1 either. Nothing is lost by the narrowing today.

### C.6 Fail directions are opposite on purpose

§B fails open; §C fails closed. A wrongly allowed signature voids the contract model; a wrongly
denied sign costs the operator one terminal command, which is where signing belongs anyway. A
wrongly denied push halts a lane; a wrongly allowed one costs a cancelled CI run.

### C.7 The `guided` promise (finding N8)

SPEC-AUTONOMY-ESCALATION-001 REQ-AE-001 promises that nothing changes under `guided`. This guard
denies under `guided` too, because a contract signed by an agent is worthless in any mode. The
contradiction is resolved in **wording, not behavior**: the promise is scoped to the escalation
detector, and both the template comment and the rule text say the sign deny is mode-independent
(REQ-AP-010, AC-AP-013). Weakening the guard to preserve the sentence would trade a real protection
for a phrasing.

## §D — Mission-validator projection (REQ-AP-008)

- **Target.** `mission.ValidateMissionDecision` (`internal/mission/policy.go:200`), verified present
  with the doc comment "treats every Decision field as untrusted data. It never calls a tool or
  performs a state change".
- **Direction.** One way: contract → `mission.SealedContract` / `mission.MissionSnapshot` /
  `mission.Decision`. The verdict comes back from mission. Its five refusal reasons
  (`mission_not_running`, `mission_mismatch`, `policy_mismatch`, `stale_snapshot`,
  `expired_decision`) are **not** re-implemented.
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
