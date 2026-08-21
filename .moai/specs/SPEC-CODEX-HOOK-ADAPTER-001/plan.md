# Plan — SPEC-CODEX-HOOK-ADAPTER-001

> Ordered so the decisions most likely to change come first. The mapping table and the
> discard-diagnostic contract are the reviewable decisions; the wiring is mechanical.

## §A Context

The adapter is a thin seam in front of the `moai hook` dispatcher. It exists because two
things differ between the harnesses — the event name the harness passes, and three output
keys Codex declares but does not act on. Everything else measured identical, including the
payload field names, so the seam stays outside `internal/hook`.

## §B Known Issues Going In

- The three inert keys were measured on two events only. Run-phase must widen that before
  the mapping's event coverage is fixed.
- Project-level hook discovery does not work in the measured build. This SPEC does not
  install anything, so it is carried as a blocker for M4 rather than solved here.
- The whole measurement basis is one Codex build. A version bump can invalidate REQ-2.

## §C Pre-Flight (Run-Phase Entry Checks)

1. `codex --version` matches the measured build, or the divergence is recorded before
   trusting §D of the SPEC.
2. `.moai/reports/t91/hook-payloads/` and `.moai/reports/t83/probe/` both resolve — the
   golden fixtures REQ-6 depends on.
3. `go build ./...` green on the branch base.

## §D Constraints (Hard)

- No edits to decision logic under `internal/hook` (REQ-7).
- No claim of lossless translation for `systemMessage` (SPEC §E).
- No silent discard anywhere in the adapter (REQ-3).
- Template-First applies to any file mirrored under `internal/template/templates/`.

## §E Design Decisions

### D1 — The seam is a mapping layer in front of the dispatcher, not a fork of it

The payload key sets are identical between harnesses and `tool_name` already arrives
normalized to `"Bash"`, so the parsing layer is reusable as-is. Translating at the boundary
keeps one implementation of every gate. The alternative — a Codex-specific dispatcher — would
double the surface that every future hook change has to land on.

### D2 — Rewrite `continue:false` to `decision:block` rather than dropping it

`decision:block` is measured working on Stop, and its meaning ("do not end the turn; here is
why") covers what `continue:false` + `stopReason` expresses. The rewrite therefore preserves
intent. Codex rejects a `decision:block` with an empty reason, so the adapter substitutes a
default reason when `stopReason` is absent rather than emitting an invalid object.

### D3 — Discard is an event, not a silence

The card's central finding is that Codex fails silently — unknown keys, unknown event names,
and a stray `version` all produce no error. An adapter that answered silence with silence
would reproduce the defect it exists to contain. Every undeliverable message is therefore
logged with event, key, and content length. Length rather than content keeps the diagnostic
from becoming an exfiltration path for whatever the hook was reporting.

### D4 — Per-event stderr classification is table-driven

Two classes were measured (blocking reason, continuation prompt) and the binary names a third
(`PermissionRequest` denial reason). A table keyed by event keeps the untested events explicit
rather than letting them inherit whichever branch was written first.

## §F Milestones

### M1 — Event-name mapping table + rejection path (REQ-1)

The eight-pair table, plus an explicit reject for anything unrecognized. Ships with the table
as data so M4 can consume the same source for its whitelist.

### M2 — Output mapping for the three inert keys (REQ-2)

`continue:false` (+`stopReason`) → `decision:block` (+`reason`); `systemMessage` →
`additionalContext` where the event supports it. Confined to the three measured-inert keys —
anything measured working passes through untouched.

### M3 — Discard diagnostics (REQ-3)

The undeliverable path from M2, made visible. Written after M2 so the discard cases are known
rather than guessed.

### M4 — Per-event stderr classification (REQ-4)

The table from D4, with the two measured classes populated and the unmeasured events marked
as such in the table itself.

### M5 — Config emission constraints (REQ-5)

The `version`-prohibition and the field whitelist, expressed where a config would be written.
No generator is built here; this milestone is the constraint other work reads.

### M6 — Golden-file tests (REQ-6)

Tests over the captured dumps for the six events with goldens. Each assertion cites the file
it reads.

### M7 — Verification + commit

Affected-package tests, `go vet`, `go build ./...`, and the template neutrality guard if any
mirrored file moved. Full-suite verdict comes from CI, not locally.

## §G Anti-Patterns (Avoid)

- Widening the mapping to keys that measured working — every added translation is a place for
  the two harnesses to drift apart.
- Treating a passing `codex exec` run as proof a key worked. The measured failures all
  produced `rc=0`; only the observable effect distinguishes them.
- Asserting the inert three are inert everywhere. Two events were tested.
- Hand-writing payload fixtures when captured dumps exist.

## §H Cross-References

- `.moai/reports/t83/precondition-measurement-round3.md` §1, §2, §4 — the measurements M1-M5
  are built on
- `.moai/reports/t91/README.md` §2 — the retired `SubagentStop` mapping
