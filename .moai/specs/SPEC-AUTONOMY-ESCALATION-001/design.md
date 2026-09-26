# design.md — SPEC-AUTONOMY-ESCALATION-001

Decisions most likely to change come first. Everything here is a proposal for the run
phase; spec.md states only the observable behavior.

## §A — Record layer (lead rulings 09-26 #4 and (2) #6)

The queue store cannot carry a fourth state (research.md P10), and doctrine forbids a
machine acting on a card. Settled shape:

- One Markdown file per class and fingerprint at
  `.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md`, YAML frontmatter per spec.md
  §I.1, body sections Observation / Options / Not observed. A re-trip after resolution gets a
  `-<n>` suffix so a resolved decision is never overwritten.
- Dedup is a file lookup: the same class and fingerprint with `status: open` means increment
  `occurrences` and `updated_at` in place.
- "Needs-decision" = "the card has at least one `contract` or `operational` record with
  `status: open`"; `revoke` records (A3) never count.
- The factory record layer (F1) ingests these files later; M1 does not wait for it.

## §B — Activation

- Config. `mode` and `escalation.budget_default` are defined by A1 (§ Configuration of the A1
  draft); this SPEC adds one key beside them:

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
| 10 contract-void (runs first) | every hook and checkpoint under `contract` mode, before resolution | PreToolUse: file stat + signature-block presence; commit checkpoint: A1 verify |
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
(REQ-AE-022, spec.md §F O10).

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
git strategy configuration names (REQ-AE-009). New package = directory with source files at HEAD
absent at base. CLI-verb / MCP-tool / config-key recognizers are per language; the first run
ships one for Go (command registration, tool-name literal, config yaml tag) and every other
language reports those sub-kinds as not-observed. Unsupported language or CGO-less build →
not-observed.

### C.5 Class 5 inputs

Recorded verdict files under `.moai/reports/<card-id>/`; `audit_multi` JSON results when
persisted. `disagreement_flag` nil → not-observed.

### C.6 Card state file — where "previously observed" lives (lead ruling 09-26 (2) #4)

`<worktree root>/.moai/state/escalation/<card-id>.json`, one per card, written by the detector
only. It holds:

- `contract` — the last contract observed signed-valid for this card: SPEC ID, path,
  `contract_sha256`, and the time it was observed; empty until the first signed-valid
  observation.
- `disarmed` — whether a `contract-void` record has already been written for that contract, so
  the loss escalates exactly once.
- counters for class 7 (operations, turns, audit retries per audit kind) and the failure
  fingerprint history for class 8.

Order on every hook and checkpoint under `contract` mode:

1. Read the card state file.
2. If it names a signed-valid contract and `disarmed` is false, check that contract (stat and
   signature-block presence on PreToolUse; A1 verify at the commit checkpoint). On absence,
   missing signature, or a non-acceptance `signed-invalid`, write the `contract-void` record,
   set `disarmed`, and append the audit line.
3. Only then run the resolver (§C.8). Its `not-armed` line, if any, follows the `contract-void`
   record, never replaces it.

Operations: PostToolUse write-capable and Bash calls. Turns: Stop hook events. Audit retries:
audit verdict files beyond the first, per audit kind.

### C.7 Contract input

Class detection reads the contract through A1's verify (state + reasons + measured acceptance
values), never by re-parsing the schema itself. Only the source-line mapping for the record's
`contract_ref` reads the file text directly, because verify output carries no line numbers
(spec.md §F O6).

### C.8 Contract resolver (lead rulings 09-26 #2 and (2) #1)

One function takes the tool call's working directory and returns armed (with the contract path)
or not-armed (with the layer and reason). Layer (a): find the worktree root, take its directory
base name as the card id, and read that card's `spec_id` from the queue store (read-only; the
queue file resolves against the primary checkout from every linked worktree). Layer (b): in that
SPEC's directory, accept `contract.yaml` only if it carries a `signature` block and the SPEC's
`spec.md` `status` is neither `completed` nor `archived`. Exactly one survivor arms. It never
reads the branch name. Keeping it one function is what lets the F1 worktree↔card record replace
layer (a) later without touching the class detectors.

### C.9 Exemption root sources (REQ-AE-013)

| Root | Source | When undeterminable |
|---|---|---|
| `.moai/reports/<card-id>/` | card id from the worktree directory name | never — the card id always comes from the path |
| `.moai/state/` | worktree root | never |
| OS temporary directory | `os.TempDir()` plus the resolved form of `$TMPDIR` | never |
| session scratchpad | hook input field or environment variable the runtime supplies, if any | outside-root writes are listed not-observed |
| auto-memory store | the store path `moai memory doctor` resolves (every candidate store) | outside-root writes are listed not-observed |
| `ownership.scratch` | resolved contract | not applicable without a contract |

## §D — Alternatives considered

- **Add a `needs-decision` queue state** — rejected: schema freeze test and doctrine.
- **Reuse `scripts/ac-baseline` directly** — rejected: local-only, corpus comparison.
- **Deny on trip** — rejected: that is A3; this SPEC is detection only.
- **Silently disarm on contract loss** — rejected by lead ruling 09-26 #1b.
- **Resolve the SPEC from the branch name** — rejected by lead ruling 09-26 #2; `WT-<slug>`
  branches do not carry the card or SPEC.
- **Count every signed contract in the tree** — rejected by lead ruling 09-26 (2) #1: closed
  SPECs keep their signed contracts, so the count converged to permanent not-armed.
- **One JSON record per trip under `escalations/`** — superseded by lead ruling 09-26 (2) #6.

(The former §G — push serializer and contract-sign guard — moved to card t1245; spec.md §K.)

## §F — Proposed package

Detector logic in `internal/escalation` (resolver, class detectors, record writer, state file);
hook wiring in `internal/hook`. Test names in acceptance.md bind to these packages; a rename is
recorded in progress.md §E.2.
