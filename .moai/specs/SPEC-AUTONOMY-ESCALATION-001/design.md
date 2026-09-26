# design.md — SPEC-AUTONOMY-ESCALATION-001

Decisions most likely to change come first. Everything here is a proposal for the run
phase; spec.md states only the observable behavior.

## §A — Record layer (lead rulings 09-26 #4 and (2) #6)

The queue store cannot carry a fourth state (research.md P10), and doctrine forbids a
machine acting on a card. Settled shape:

- One Markdown file per class and fingerprint at
  `<worktree root>/.moai/reports/<card-id>/escalation/<class>-<fingerprint>.md`, YAML
  frontmatter per spec.md §I.1, body sections Observation / Options / Not observed. A re-trip
  after resolution gets a `-<n>` suffix so a resolved decision is never overwritten.
- The directory is gitignored and lives in the card worktree. The lead reads it before the
  worktree is disposed; the factory record layer (F1) ingests the files later. M1 does not wait
  for F1.
- Dedup is a file lookup: the same class and fingerprint with `status: open` means increment
  `occurrences` and `updated_at` in place.
- "Needs-decision" = "the card has at least one `contract` or `operational` record with
  `status: open`"; `revoke` records (A3) never count.

## §B — Activation

- Config. `mode` and `escalation.budget_default` are defined by A1 (A1 § Configuration); this
  SPEC adds one key beside them:

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
| 10 detection-disarmed (runs first) | every hook and checkpoint under `contract` mode, before resolution | state-file digest check; contract read; in-process A1 verify only when the contract bytes differ from the cached digest |
| 1 acceptance-change | commit checkpoint and any hook whose cached verify is invalidated (in-process A1 verify) | file hash + counter |
| 2a invariant command | PostToolUse Bash (exit status observed, command not re-run) | string compare |
| 2b frozen-file | PreToolUse write tools | glob match against cached `frozen_files` |
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
why it is flagged rather than chosen silently. On every hook path the HEAD commit a record needs
is read from `.git/HEAD` and the ref it names (or the worktree's `gitdir` file), with no
subprocess.

PreToolUse reads, in order: each `.moai/specs/*/contract.yaml` under the worktree root (only
for the `card` field and, when needed, the bytes to digest); the SPEC's `spec.md` frontmatter
`status` through A1's `terminal` derivation; the card state file and this card's audit log for the
chain and digest check; HEAD and ref files. A1 verify is called in-process — A1 states verify runs
without subprocesses or network access so a PreToolUse hook can call it — only at first arming
and when the contract digest changed (lead ruling 09-26 (3) #3).

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
a denylisted command is still recorded before the existing deny fires (AC-AE-013). Pattern set:
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

### C.6 Card state and card audit log — where "armed" lives (lead rulings (2) #4, (3) #2, (4) #2, (5))

Two files per card, both in the contract store `$MOAI_HOME/db/<project-key>/contract/escalation/`
and both written by the detector only:

| File | Exact name | Role |
|---|---|---|
| card audit log | `<card-id>.log.jsonl` | **authoritative**: append-only, hash-chained entries that decide whether the card is armed |
| card state file | `<card-id>.json` | cache: arming snapshot, verify cache, counters |

`<card-id>` is the worktree directory base name, byte for byte — no case folding, no escaping,
no hashing. It is already a valid file name on the host, because the worktree directory exists
under that name, and two worktrees of one project cannot share a directory name, so no two cards
share a file. The store is outside the repository and every worktree, in the queue database's
home layout, and shared with A3's signing-event store. `$MOAI_HOME` and `<project-key>` resolve
exactly as the queue database resolves them (research.md P17); because the key derives from the
project root, every worktree of one project shares one store. Where the store cannot be
resolved, the detector cannot record an arming: that is a REQ-AE-004 fault (`not-checked`), and
nothing arms. The store is not a project path, so no `ownership` glob matches it; a tool-call
write there trips `ownership-move` ahead of every exemption (REQ-AE-013, §C.9).

**Card audit log entries** (one JSON object per line, each carrying `prev` = SHA-256 of the
previous line of the same file, empty for the first line):

- `armed` — SPEC ID, contract path, `signature.contract_sha256`; starts an arming episode.
- `state` — SHA-256 of the card state file bytes just written.
- `disarmed` — the reason and the record fingerprint; ends the episode.
- `not-armed`, `not-checked`, `warning` — the audit lines REQ-AE-002/004/023 name.

The card is armed exactly while the log's most recent `armed` entry has no later `disarmed`
entry. Nothing in the card state file can change that answer.

**Card state file** holds:

- `armed` — the arming snapshot: SPEC ID, contract path, `signature.contract_sha256`, the digest
  of the contract file bytes, card id, the derived `frozen_files` (with the registry Frozen list
  obtained through the constitution registry loader at arming), `effective_never`, `scratch`, the
  budget, and the time armed.
- `verify_cache` — the last verify state and reasons, keyed by the contract byte digest.
- counters for class 7 (operations, turns, audit retries per audit kind) and the failure
  fingerprint history for class 8.

Order on every hook and checkpoint under `contract` mode:

1. Read this card's audit log. If the file is absent: when prior evidence of arming exists — a
   card state file carrying an `armed` value, or a record of kind `contract` or class
   `detection-disarmed` under `.moai/reports/<card-id>/escalation/` — write the `state-tamper`
   record and start a new log whose first entry is `disarmed`; otherwise treat the card as never
   armed. If the file is present, verify its chain (§C.11).
2. If the log shows the card armed, check the card state file against the log's latest `state`
   entry. A missing file, a missing `armed` value, a digest mismatch, or a broken chain is
   `state-tamper` — never a disarm by itself.
3. Otherwise evaluate the remaining disarm reasons against the arming: the armed contract file is
   gone (`contract-absent`); verify is no longer `signed-valid` (`signature-invalid`) — verify
   runs fresh at every commit checkpoint, so an `acceptance.md` edit that leaves the contract
   bytes unchanged is still seen there, and on PreToolUse only when the contract bytes differ
   from the cached digest; the SPEC's `terminal` is true (`terminal-status`); the contract's
   `card` no longer equals the card id, or another non-terminal contract claims the card
   (`card-mismatch`).
4. On the first reason found (steps 1-3), write the `detection-disarmed` record and append the
   `disarmed` log entry. Classes 7-9 keep counting against the budget recorded at arming.
5. Only then run the resolver (§C.8) for an unarmed card. Its `not-armed` line, if any, follows
   the disarm entry, never replaces it.

A new arming (a later signed-valid contract for the same card) appends a new `armed` entry and
starts a new episode; a later loss writes a new disarm record.

Operations: PostToolUse write-capable and Bash calls. Turns: Stop hook events. Audit retries:
audit verdict files beyond the first, per audit kind.

### C.7 Contract input

Class detection reads the contract through A1's verify (state, reasons, measured acceptance
values, derived lists), never by re-parsing the schema itself. Only the `card` field lookup and
the source-line mapping for `contract_ref` read the file text directly — the first because the
resolver must choose which contract to verify, the second because verify output carries no line
numbers (spec.md §F O6).

### C.8 Contract resolver (lead ruling 09-26 (3) #1)

One function takes the tool call's working directory and returns armed (with the contract path)
or not-armed (with the reason):

1. Find the worktree root; its directory base name is the card id.
2. Read every `.moai/specs/*/contract.yaml` under the worktree root and keep those whose `card`
   field equals the card id.
3. Drop those whose SPEC's A1 `terminal` is true (`completed` / `archived`, quoted or unquoted).
4. Zero survivors → `not-armed` (no claimant). Two or more → `not-armed` plus a warning naming
   every SPEC ID. Exactly one → the arming step (REQ-AE-023): in-process A1 verify; arm on
   `signed-valid`; otherwise `not-armed`, plus a warning with the reason codes on `signed-invalid`.

It never reads the branch name or the queue. Keeping the queue out is deliberate: `spec_id` was
filled on 0 of 119 live cards (plan-audit iteration 3, P1), and keeping it filled is F1's
concern. Keeping the resolver one function lets F1's worktree↔card record replace step 1 later
without touching the class detectors.

### C.9 Exemption root sources (REQ-AE-013)

| Root | Source | When undeterminable |
|---|---|---|
| `.moai/reports/<card-id>/` | card id from the worktree directory name | never — the card id always comes from the path |
| `.moai/state/` | worktree root | never |
| OS temporary directory | `os.TempDir()` plus the resolved form of `$TMPDIR` | never |
| session scratchpad | hook input field or environment variable the runtime supplies, if any | outside-root writes are listed not-observed |
| auto-memory store | the store path `moai memory doctor` resolves (every candidate store) | outside-root writes are listed not-observed |
| `ownership.scratch` | A1-derived `scratch` from the resolved contract | not applicable without a contract |

The contract store is deliberately absent from this table, and it is judged **before** the table
is consulted: a tool-call write into it — the only way an agent could edit the card state or log
without Bash — trips `ownership-move` even when the store itself lies under the OS temporary
directory or another root above (lead instruction on plan-audit iteration-4 Q3). It is never
listed as not-observed, because the store path comes from its own resolution rule.

### C.10 Fingerprint inputs per class

The fingerprint must stay stable across repeated observations of one decision and differ across
distinct decisions. Inputs, hashed together with the class name:

| Class | Normalized observation |
|---|---|
| 1 | the sorted set of acceptance reason codes |
| 2a | the invariant string |
| 2b, 3 | the repository-relative target path and the matching glob |
| 4 | addition kind and qualified name |
| 5 | the pair of disagreeing verdict sources |
| 6 | the normalized command (whitespace collapsed, remote and ref kept) |
| 7 | the exceeded dimension only — never the observed count, so a growing count increments one record |
| 8 | the failure fingerprint (normalized command plus diagnostic key) |
| 9 | the audit kind |
| 10 | the arming's `signature.contract_sha256` — so one arming yields one record whatever the reason |

### C.11 Card audit log chain and tamper evidence (lead ruling (5))

Every line of `<card-id>.log.jsonl` carries `prev`, the SHA-256 of the previous line of the same
file. Each write of the card state file is followed by one `state` entry carrying the SHA-256 of
the new bytes. A chain is broken when any `prev` does not match its predecessor or a line does
not parse. Because each card has its own file, another card's appends never touch this card's
chain (no shared-log serialization is needed, and a legitimate append by one card cannot trip
`state-tamper` on another).

Tamper evidence, and its limits:

- A write-capable tool call to either file trips `ownership-move` before any exemption applies
  (REQ-AE-013), even when the store sits under the OS temporary directory.
- A Bash edit is seen only at the next hook, and only if it did not recompute the chain: the chain
  is unkeyed, so rewriting the state file and appending a matching `state` entry is not caught.
- Deleting the log alone is caught while prior evidence of arming survives (§C.6 step 1).
  Deleting the log, the state file, and every `contract`-kind or `detection-disarmed` record of
  the card leaves no trace; the card reads as never armed and the next arming restarts the class
  7 counters. The store is local and single-user: the goal is tamper evidence, not prevention.

## §D — Alternatives considered

- **Add a `needs-decision` queue state** — rejected: schema freeze test and doctrine.
- **Reuse `scripts/ac-baseline` directly** — rejected: local-only, corpus comparison.
- **Deny on trip** — rejected: that is A3; this SPEC is detection only.
- **Silently disarm on contract loss** — rejected by lead ruling 09-26 #1b; generalized to every
  disarm reason by (3) #2.
- **Resolve the SPEC from the branch name** — rejected by lead ruling 09-26 #2; `WT-<slug>`
  branches do not carry the card or SPEC.
- **Count every signed contract in the tree** — rejected by lead ruling 09-26 (2) #1: closed
  SPECs keep their signed contracts, so the count converged to permanent not-armed.
- **Resolve through the queue `spec_id`** — rejected by lead ruling 09-26 (3) #1: the field is
  almost never filled, so nearly every card would stay not-armed.
- **Keep the state file inside the `.moai/state/` exemption** — rejected by lead ruling 09-26
  (3) #4: an agent could rewrite "previously armed" and disarm silently.
- **Keep the state file in the worktree with a carve-out** — superseded by lead ruling 09-26
  (4) #2: the moai-owned store keeps it out of every project path and out of worktree disposal.
- **One JSON record per trip under `escalations/`** — superseded by lead ruling 09-26 (2) #6.

(The former §G — push serializer and contract-sign guard — moved to card t1245; spec.md §K.)

## §F — Proposed package

Detector logic in `internal/escalation` (resolver, class detectors, record writer, state file,
audit chain); hook wiring in `internal/hook`. Test names in acceptance.md bind to these
packages; a rename is recorded in progress.md §E.2.
