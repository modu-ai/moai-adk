# SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 — design

> Authored at v0.2.0 by card t1259, when the Tier was raised to L. It is deliberately thin: it
> carries the two design decisions **this** SPEC takes, and cross-references
> `SPEC-INSTRUCTION-FILES-UNIFY-001`'s `design.md` §C for the read-order analysis both SPECs
> share rather than duplicating it.

## §A The `REQ-IFU-010` split — resolution and refusal are two situations

### The contradiction this resolves

The carve transferred a requirement and a criterion that could not both be satisfied.
`REQ-IFU-010` commanded **resolution** — "the migration verb shall leave exactly one of the two in
place" — and `AC-IFU-014` commanded **refusal** — "exits non-zero without modifying either file".
In the case that matters, both files present, refusing leaves **both** in place, which is what the
requirement forbids; resolving satisfies the requirement and fails the criterion. A correct
implementation of either violated the other.

This was observed, not taken on the parent audit's word. Both bodies were re-read in this tree and
the contradiction reproduces on their current text (`progress.md` §E.1).

### The decision, and why refusal is the half that survives

**The verb refuses when both files exist, and the requirement is the thing that changes.**

The auditor proposed this and the reasoning holds on its own: `AGENTS.local.md` and
`CLAUDE.local.md` are both **user-authored**. Choosing between them means either discarding
content the user wrote or concatenating two documents whose relationship the tool cannot know. A
tool that silently picks is a data-integrity hazard, and the failure is quiet — the user discovers
it when instructions stop taking effect, not when the command runs.

Refusing costs the user one command and one decision. Choosing wrong costs them a file. The
asymmetry is not close.

### The shape

`REQ-IFU-010` becomes two clauses under one id rather than two requirements, because they are two
faces of one invariant — no coexistence — and splitting the id would break the cross-references
the carve exists to preserve:

- **`REQ-IFU-010a`** — where exactly one file is present, leave exactly one in place (resolution).
- **`REQ-IFU-010b`** — where both are present, refuse: exit non-zero, modify neither, name the
  coexistence (refusal).

**No new acceptance criterion is needed, and this was checked rather than assumed.** The concern
is real — a newly separated clause with no criterion is a coverage gap a plan-audit finds — but
both clauses already had one. `AC-IFU-013` fixtures `CLAUDE.local.md` present and
`AGENTS.local.md` absent, which is exactly `010a`'s unambiguous case; `AC-IFU-014` fixtures both
present, which is `010b`. The traceability table now records them on separate rows so the mapping
is visible rather than inferred.

What did change in `AC-IFU-014`: it now captures a sha256 of each file before and after. "Without
modifying either file" was previously unasserted — a refusal that exits non-zero *after* writing
would have passed.

### What is not decided here

Whether the verb should offer a `--force` or `--pick` escape from the refusal. Nothing in this
SPEC requires one, adding one would re-open the data-integrity question, and a flag is cheap to
add later and expensive to remove. Left out deliberately, not overlooked.

## §B Where the deprecation advisory goes

### The question (parent `research.md` Q4)

Whether the Codex launcher's fallback branch has a diagnostic surface on which to emit the
advisory **without polluting `developer_instructions`**. The concern is precise: everything the
launcher writes into that payload becomes model context, so an advisory addressed to the operator
would arrive as an instruction to the model.

### The answer: yes, at the caller

Measured in this tree, 2026-09-26 (`research.md` §A carries the readings):

`codexLocalDeveloperInstructionArgs(projectRoot string) ([]string, error)` —
`internal/cli/codex_launcher.go` — is a pure producer. It has no writer, no logger, and no
`cobra.Command`; everything it builds is JSON-encoded into the `developer_instructions` override.
Emitting from inside it would land in the payload, which is the outcome Q4 asks to avoid.

Its **sole** caller, `runCodexLaunch`, holds `cmd.ErrOrStderr()` and already writes operator
diagnostics to it — the install hint and the worktree-writer error both go there. That is the
diagnostic surface, it already exists, and it is already the established channel for exactly this
kind of message.

### The consequence for the implementation

The producer must report *which* files it read so the caller can decide whether the fallback
branch was taken. Two shapes, both with precedent in this file:

1. **Widen the return** — the producer returns the names it read alongside the args, and the
   caller emits when the set contains `CLAUDE.local.md`. Keeps the producer pure.
2. **Pass the writer in** — the producer takes an `io.Writer` and emits directly. Fewer moving
   parts, but it makes a pure function do I/O.

**Shape 1 is preferred.** The function's `@MX:ANCHOR` comment declares it the "single
local-instruction producer for every launch form", and purity is what lets the existing tests
assemble the payload without a command. Shape 2 would force every caller of a producer to own a
writer. The decision is recorded as a preference, not a mandate: `AC-IFU-011` asserts the
advisory reaches the diagnostic stream and not the payload, and either shape can satisfy it.

### Ordering

This work builds on the parent SPEC's M2, which changes the same loop's **iteration order**. The
two edits are compatible but not independent, which is why plan.md §B holds the dependency and
why the advisory milestone is M2 here rather than M1.

## §C Cross-references

- `SPEC-INSTRUCTION-FILES-UNIFY-001/design.md` §C — the read-order analysis both SPECs rest on.
  Not duplicated here.
- `internal/cli/codex_launcher.go` — `codexLocalDeveloperInstructionArgs` (the producer) and
  `runCodexLaunch` (the caller holding the diagnostic surface). Cited by symbol; `origin/develop`
  keeps moving line numbers.
- `internal/cli/codex_contract.go` — `codexLocalInstructionName` / `codexClaudeLocalName`, the two
  filename constants the no-coexistence invariant is about.
- `internal/cli/migrate_agency.go` — the move-plus-backup precedent M1 models the verb on.
