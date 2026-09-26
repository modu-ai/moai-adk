# design.md — SPEC-AUTONOMY-ESCALATION-001

Decisions most likely to change come first. Everything here is a proposal for the run
phase; spec.md states only the observable behavior.

## §A — Record layer: where needs-decision lives (highest change likelihood)

The queue store cannot carry a fourth state (research.md P10), and doctrine forbids a
machine acting on a card. Proposed shape:

- One escalation record per trip, stored under the card's evidence area
  (`.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md`), plus one machine-readable
  index line per record so a reader can list open escalations without parsing prose.
- "Needs-decision" = "the card has at least one open escalation record". It is derived by
  reading records; nothing in the queue changes.
- A1 and the factory redesign (F1) are said to share a record layer. If that layer lands
  first and offers a per-card record, the index moves there. Decision deferred: **Q1**.

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
| 1 acceptance-change | checkpoint | file hash + counter |
| 2a invariant command | PostToolUse Bash (exit status observed, command not re-run) | string compare |
| 2b frozen-file | PreToolUse write tools | path match |
| 3 ownership-move | PreToolUse write tools | glob match |
| 4 new-arch/API | checkpoint | git blob read + extraction |
| 5 contradictory-evidence | checkpoint (reads recorded verdict files) | file read |
| 6 irreversible-action | PreToolUse Bash | regex match |
| 7 budget | PostToolUse (counter) + checkpoint | counter |
| 8 same-diagnostic ×3 | PostToolUse Bash failure | fingerprint count |
| 9 audit cap | checkpoint (reads verdict files) | file read |

### C.1 Checkpoint

A checkpoint is (a) a commit on the card branch observed by the PostToolUse path, and
(b) an explicit on-demand check. The on-demand surface is a proposed CLI verb
(`moai escalation check`); adding a verb is itself a new-API event, which is expected and
is the reason this is flagged for the auditor rather than chosen silently.

### C.2 Invariant commands are observed, never executed

The detector never runs `go test` or any contract command. It compares the executed Bash
command string with each contract invariant string and reads the recorded exit status.
Consequence: an invariant that nobody runs is never checked — this is recorded as a known
limit rather than papered over.

### C.3 Class 6 pattern set

Built from: push target ref ∉ contract `actions`; `git tag` / `git push --tags` /
`git push <remote> refs/tags/*`; `gh release create`; force push; and the existing
denylist patterns (reused by reference, not copied).

### C.4 Class 4 before-side

`graph.FileAPI` reads the working tree. The before-side is obtained by materializing the
base-commit blob (`git show <base>:<path>`) to a temp file under the scratch area and
extracting from it with the same extractor. New package = directory with source files at
HEAD absent at base. New CLI verb / MCP tool / config key = new registration call, new
tool name literal, new yaml tag in the config types, compared base vs HEAD. Unsupported
language or CGO-less build → not-observed (REQ-AE-019).

### C.5 Class 5 inputs

Recorded verdict files under `.moai/reports/<card-id>/`; `audit_multi` JSON results when
persisted. `disagreement_flag` nil → not-observed.

### C.6 Counters

Turns: goal ceiling state when a goal is armed, else the detector's own PostToolUse turn
counter. Operations: detector-counted write/Bash tool calls. Audit retries: count of audit
verdict files for the card. Counters live beside the escalation index.

Option (not decided): the A1 draft projects a signed contract onto `mission.MissionContract`
(`ResourceLimits.MaxOperations` = `budget.operations`, `StopConditions` = `escalate_on`) so that
this detector could reuse `ValidateMissionDecision` (`internal/mission/policy.go:200`) for the
budget and action checks instead of its own counters.

### C.7 Contract input

Class detection reads the contract through A1's verify (state + reasons + measured acceptance
values), never by re-parsing the schema itself. Only the source-line mapping for the report's
`contract.yaml:<line>` reads the file text directly, because verify output carries no line
numbers (spec.md §F O6).

## §D — Report shape

```markdown
# Escalation — <class>
- spec: <SPEC-ID>   card: <card-id|unresolved>   head: <sha>
- tripped: contract.yaml:<line> (escalate_on: <token>) | config: <key>
- occurrences: <n>
## Observation
<command or event> → <verbatim evidence>
## Options
1. ...
2. ...
## Not observed
- <anything the detector could not check>
```

## §E — Alternatives considered

- **Add a `needs-decision` queue state** — rejected: schema freeze test and doctrine.
- **Reuse `scripts/ac-baseline` directly** — rejected: local-only, corpus comparison.
- **Deny on trip** — rejected: that is A3; this SPEC is detection only.
