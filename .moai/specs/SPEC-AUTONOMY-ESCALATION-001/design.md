# design.md — SPEC-AUTONOMY-ESCALATION-001

Decisions most likely to change come first. Everything here is a proposal for the run
phase; spec.md states only the observable behavior.

## §A — Record layer: where needs-decision lives (lead ruling 09-26 #4)

The queue store cannot carry a fourth state (research.md P10), and doctrine forbids a
machine acting on a card. Settled shape:

- One JSON record per trip at `.moai/reports/<card-id>/escalations/<timestamp>.json`, schema in
  spec.md §I. With the card id unresolved (plan.md Q9), the record lands under
  `.moai/reports/<SPEC-ID>/escalations/` with `card_id: "unresolved"`.
- "Needs-decision" = "the card has at least one record with `status: open`", derived by reading
  records; nothing in the queue changes.
- The factory record layer (F1) ingests these files later; M1 does not wait for it.

## §B — Activation

- Config. `mode` and `escalation.budget_default` are defined by A1 (draft `8f77d9a33`
  § Configuration); this SPEC adds one key beside them:

  ```yaml
  workflow:
    autonomy:
      escalation:
        new_api_detector: graph   # graph | off   (added by this SPEC)
  ```

- The mode check is the first statement on every detector path; `guided` returns before
  any file read beyond the already-loaded config (REQ-AE-001, C4).

## §C — Detector placement per class

| Class | Hook point | Cost class |
|---|---|---|
| 1 acceptance-change | commit checkpoint (PostToolUse, in-process A1 verify) | file hash + counter |
| 2a invariant command | PostToolUse Bash (exit status observed, command not re-run) | string compare |
| 2b frozen-file | PreToolUse write tools | path match |
| 3 ownership-move | PreToolUse write tools | glob match |
| 4 new-arch/API | on-demand checkpoint only | git blob read + extraction |
| 5 contradictory-evidence | commit checkpoint (reads recorded verdict files) | file read |
| 6 irreversible-action | PreToolUse Bash, above the denylist | regex match |
| 7 budget | PostToolUse (operations), Stop (turns), checkpoint (audit retries) | counter |
| 8 same-diagnostic ×3 | PostToolUse Bash failure | fingerprint count |
| 9 audit cap | commit checkpoint (reads verdict files) | file read |
| 10 contract-void | every armed hook and checkpoint (compare with state file) | file stat + A1 verify |

### C.1 Checkpoints and hook budget

A **commit checkpoint** runs inside the PostToolUse hook of a commit on the card branch and
does only in-process work within the hook's configured timeout (spec.md C4). The **on-demand
checkpoint** is the only place class 4 runs; its surface is a proposed CLI verb
(`moai escalation check`, open question Q2) — adding a verb is itself a class-4 event, which is
why it is flagged rather than chosen silently. On the PreToolUse path the HEAD commit a record
needs is read from `.git/HEAD` and the ref it names (or the worktree's `gitdir` file), with no
subprocess.

### C.2 Invariant commands are observed, never executed

The detector never runs `go test` or any contract command. It compares the executed Bash
command string with each contract invariant string and reads the recorded exit status. An
invariant no tool call executed since the previous checkpoint is listed as not-observed
(REQ-AE-019, spec.md §F O10).

### C.3 Class 6 placement against the denylist anchor

`internal/hook/pre_tool.go:505-517` runs `checkBashCommand` and returns early on deny/ask, and
the `@MX:ANCHOR` there forbids any conditional return above it. The class 6 detector is called
**immediately before** `checkBashCommand` as a function with no return value on the hook path:
it records and returns nothing the hook branches on, so no conditional return is introduced and
a denylisted command is still recorded before the existing deny fires (AC-AE-014). Pattern set:
push target not authorized by `push-develop`; `git tag` / `git push --tags` /
`git push <remote> refs/tags/*`; `gh release create`; force push; and the existing denylist
patterns (reused by reference, not copied).

### C.4 Class 4 before-side and languages

`graph.FileAPI` reads the working tree. The before-side is obtained by materializing the
base-commit blob (`git show <base>:<path>`) to a temp file under the scratch area and extracting
from it with the same extractor. Card base = merge-base of HEAD with the integration branch the
git strategy configuration names (spec.md REQ-AE-009). New package = directory with source files
at HEAD absent at base. CLI-verb / MCP-tool / config-key recognizers are per language; the
first run ships one for Go (command registration, tool-name literal, config yaml tag) and every
other language reports those sub-kinds as not-observed. Unsupported language or CGO-less build →
not-observed.

### C.5 Class 5 inputs

Recorded verdict files under `.moai/reports/<card-id>/`; `audit_multi` JSON results when
persisted. `disagreement_flag` nil → not-observed.

### C.6 Counters and state file

Operations: PostToolUse write-capable and Bash calls. Turns: Stop hook events. Audit retries:
audit verdict files beyond the first, per audit kind. The state file (beside the escalation
directory) holds these counters, the failure-fingerprint history, and the last observed contract
state per SPEC (needed by class 10).

Option (not decided): the A1 draft projects a signed contract onto `mission.MissionContract`
(`ResourceLimits.MaxOperations` = `budget.operations`, `StopConditions` = `escalate_on`) so that
this detector could reuse `ValidateMissionDecision` (`internal/mission/policy.go:200`) for the
budget and action checks instead of its own counters.

### C.7 Contract input

Class detection reads the contract through A1's verify (state + reasons + measured acceptance
values), never by re-parsing the schema itself. Only the source-line mapping for the record's
`tripped.line` reads the file text directly, because verify output carries no line numbers
(spec.md §F O6).

### C.8 Contract resolver (lead ruling 09-26 #2)

One function takes the tool call's working directory and returns one of: armed (with the
contract path), not-armed (zero signed candidates), or ambiguous (candidate list). It finds the
worktree root, globs `.moai/specs/*/contract.yaml`, and counts files carrying a `signature`
block. It never reads the branch name. Keeping it one function is what lets the F1
worktree↔card record replace it later without touching the class detectors.

## §D — Alternatives considered

- **Add a `needs-decision` queue state** — rejected: schema freeze test and doctrine.
- **Reuse `scripts/ac-baseline` directly** — rejected: local-only, corpus comparison.
- **Deny on trip** — rejected: that is A3; this SPEC is detection only.
- **Silently disarm on contract loss** — rejected by lead ruling 09-26 #1b.
- **Resolve the SPEC from the branch name** — rejected by lead ruling 09-26 #2; `WT-<slug>`
  branches do not carry the card or SPEC.

## §G — A3 preconditions (lead assignment 09-26; separate from the detector)

### G.1 Push serializer (REQ-AE-024)

- **Resource.** The existing slot lease, resource name `push-develop` (a valid slot resource
  name), record at `.moai/state/slot-leases/push-develop.json` in the primary checkout, root
  resolved by `kanban.ResolveSlotLeaseRoot` — the same record `moai slot status --resource
  push-develop` reads. No new record format, lock, or verb.
- **Activation.** Only under `autonomy.mode: contract` with `push-develop` in the resolved
  contract's `actions`. It does not depend on `workflow.slot_lease.enabled`, which keeps gating
  the generic slot guard only.
- **Matcher.** A Bash command whose program is `git`, subcommand `push`, and whose refspec
  targets `develop` (with the same quote handling as `checkSlotLease`).
- **Admit.** Free, expired (declared bound elapsed), stale (owning session gone), or held by the
  calling session → admit and write the lease for the calling session with bound
  `workflow.slot_lease.default_max_duration`; a takeover names the displaced holder (the slot
  lease's existing rule).
- **Deny.** Held by a different live session within its bound → deny with
  `PUSH_SERIALIZATION_VIOLATION: push-develop held by <holder>`. The agent waits and retries;
  no escalation record is written, because serialization is expected traffic, not a contract
  breach.
- **Release.** PostToolUse on the admitted push: a non-zero exit releases immediately (nothing is
  in flight). Otherwise the holder releases with `moai slot release --resource push-develop`
  after reading the CI result for the pushed head; a forgotten release costs at most the bound.
- **Fail direction.** Unreadable record, unresolvable root, unknown caller → allow plus an audit
  line (the slot guard's fail-open), because an overlapping push costs one cancelled CI run
  while a stuck deny halts the lane.

### G.2 Contract-sign guard (REQ-AE-025)

- **Placement.** A PreToolUse Bash check independent of `workflow.autonomy.mode`, placed with the
  other deny guards and called before the class 6 detector.
- **Parse.** Split the command into shell words with quote removal (not quoted-span scrubbing,
  which would erase `'moai'`); strip leading `NAME=value` assignments and the `env`, `command`,
  `exec`, `nohup` prefixes; take the program word's basename; skip global flags; match
  `moai` + `contract` + `sign`. For `sh|bash|zsh -c <string>`, parse `<string>` once more.
- **Unclassifiable.** Command substitution, a variable in program position, `eval`, or nesting
  beyond one `-c` level, when the words `contract` and `sign` both occur → deny (fail closed),
  reason marked unclassified. A classified invocation whose program is not `moai` (e.g.
  `echo "moai contract sign"`) passes.
- **Receipt path.** Allowed only in the form A1 defines (R5). Until A1 defines it, no
  `moai contract sign` form is allowed.
- **Why fail closed here and fail open in G.1.** A wrongly allowed signature voids the contract
  model; a wrongly denied sign costs the operator one terminal command, which is where signing
  belongs anyway.

## §F — Proposed package

Detector logic in `internal/escalation` (resolver, class detectors, record writer, state file);
hook wiring in `internal/hook`. Test names in acceptance.md bind to these packages; a rename is
recorded in progress.md §E.2.
