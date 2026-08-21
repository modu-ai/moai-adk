# Plan — SPEC-CODEX-HOOK-ADAPTER-001

> Ordered so the decisions most likely to change come first. The mapping table, the package
> placement, and the discard-diagnostic contract are the reviewable decisions; the rest is
> mechanical.

## §A Context

The adapter is a thin seam in front of the `moai hook` dispatcher. It exists because two
things differ between the harnesses — the event name the harness passes, and three output keys
Codex declares but does not act on. Everything else measured identical, including the payload
field names, so the seam stays outside `internal/hook`.

## §B Known Issues Going In

- The three inert keys were measured on two events only. Run-phase must widen that before the
  mapping's event coverage is fixed.
- Project-level hook discovery does not work in the measured build, and reports nothing when it
  fails. This SPEC installs nothing, so it is carried as a blocker for M4 rather than solved.
- The whole measurement basis is one Codex build. A version bump can invalidate REQ-2.
- Five of the eleven events are recognized but unadapted. They are refused rather than
  defaulted, which bounds the risk without removing it.

## §C Pre-Flight (Run-Phase Entry Checks)

1. `codex --version` matches the measured build, or the divergence is recorded before trusting
   §D of the SPEC.
2. `git ls-files .moai/specs/SPEC-CODEX-HOOK-ADAPTER-001/testdata/hook-payloads/` lists six or
   more payloads — the fixtures REQ-6 depends on, tracked so they resolve in a worktree and in
   CI. (The originals under `.moai/reports/t91/` are untracked and primary-only; do not depend
   on them.)
3. `go build ./...` green on the branch base.

## §D Constraints (Hard)

- The adapter package lives outside `internal/hook/`, and no file that existed there before
  this SPEC is modified (REQ-7).
- No claim of lossless translation for `systemMessage` (SPEC §E).
- No silent discard anywhere in the adapter (REQ-3).
- Template-First applies to any file mirrored under `internal/template/templates/`.

## §E Design Decisions

### D1 — The seam is a mapping layer in front of the dispatcher, not a fork of it

The payload key sets are identical between harnesses and `tool_name` already arrives normalized
to `"Bash"`, so the parsing layer is reusable as-is. Translating at the boundary keeps one
implementation of every gate. The alternative — a Codex-specific dispatcher — would double the
surface every future hook change has to land on.

### D2 — Recognize all eleven events; adapt six; refuse the rest explicitly

The dispatcher has a counterpart for every Codex event, so the table is complete by
construction. The six adapted are the ones with both a payload capture and observed behavior.
The other five are recognized-and-refused rather than omitted, because omission would make them
indistinguishable from a typo — and Codex's own handling of unknown names is to ignore them
silently, which is the failure shape this SPEC exists to avoid reproducing.

### D3 — Rewrite `continue:false` to `decision:block` rather than dropping it

`decision:block` is measured working on Stop, and its meaning ("do not end the turn; here is
why") covers what `continue:false` + `stopReason` expresses. The rewrite therefore preserves
intent. Codex rejects a `decision:block` with an empty reason, so the adapter substitutes a
default reason when `stopReason` is absent rather than emitting an invalid object.

### D4 — Discard is an event, and it does not travel on stderr

Undeliverable messages are logged to the adapter's own sink with event, key, and content
length. Length rather than content keeps the diagnostic from becoming an exfiltration path for
whatever the hook was reporting. The sink is deliberately not stderr: on an exit-2 path stderr
carries the blocking reason or continuation prompt, and appending to it would change what the
model receives.

### D5 — Per-event stderr classification is table-driven, with evidence tier recorded per row

Two classes were measured (PreToolUse blocking reason, Stop continuation prompt) and the binary
names others. Rows carry their evidence tier so a declared-but-unobserved classification cannot
be mistaken for a measured one at the point of use.

### D6 — The config constraint ships as a validator, not as prose

REQ-5's constraint is only checkable if something can be handed a config object and asked. A
validator gives the wiring-generator card a dependency to call and gives this SPEC an AC that
can fail. Prose in a scope section can do neither.

## §F Milestones

### M1 — Event-name mapping table + refusal paths (REQ-1)

The eleven-row table as data, with per-row adapted/not-adapted state. Two refusal paths:
recognized-but-unadapted, and unrecognized. Ships as data so the generator card can consume the
same source.

### M2 — Output mapping for the three inert keys (REQ-2)

`continue:false` (+`stopReason`) → `decision:block` (+`reason`); `systemMessage` →
`additionalContext` where the event supports it. Confined to the three measured-inert keys —
anything measured working passes through untouched.

### M3 — Discard diagnostics + sink (REQ-3)

The undeliverable path from M2, made visible, with the branch-count constant AC-REQ-3b asserts
against. Written after M2 so the discard cases are known rather than guessed.

### M4 — Per-event stderr classification (REQ-4)

The table from D5, with the two measured classes populated and evidence tier recorded per row.

### M5 — Config constraint validator (REQ-5)

The top-level and per-level key whitelists as a callable validator with negative samples drawn
from the measured failure (`version`).

### M6 — Golden-file tests (REQ-6)

Tests over the vendored `testdata/` payloads for the six adapted events. Each assertion cites
the file it reads.

### M7 — Verification + commit

Affected-package tests, `go vet`, `go build ./...`, the `internal/hook` placement check for
AC-REQ-7, and the template neutrality guard if any mirrored file moved. Full-suite verdict comes
from CI, not locally.

## §G Anti-Patterns (Avoid)

- Widening the mapping to keys that measured working — every added translation is a place for
  the two harnesses to drift apart.
- Treating a passing `codex exec` run as proof a key worked. The measured failures all produced
  `rc=0`; only the observable effect distinguishes them. Equally, treating `rc=0` as proof the
  config loaded — the `version` failure exits 0 and reports only inside the `--json` stream.
- Asserting the inert three are inert everywhere. Two events were tested.
- Reading fixtures from `.moai/reports/t91/` — untracked and primary-only, so the test passes
  on one machine and fails everywhere else.
- Silently defaulting an unadapted event to a handler.

## §H Cross-References

- `.moai/reports/t83/precondition-measurement-round3.md` §1, §2, §4 — the measurements M1-M5
  are built on
- `.moai/reports/t83/plan-audit.md` — iteration-1 audit; D1/D2/D5 there drove the changes to
  §E D2, §C pre-flight 2, and the anti-pattern list
- `.moai/reports/t91/README.md` §2 — the retired `SubagentStop` mapping
- `internal/cli/hook.go` — the dispatcher subcommand registrations
