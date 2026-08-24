# SPEC-SESSION-TELEMETRY-001 — Implementation Plan

## §A Context

Producer: `internal/statusline`. Consumers touched: `internal/cli` (`moai tokens`),
`internal/spec` (one comment), two doctrine mirror pairs (four files), and twelve docs-site
pages across four locales.

Tier **L**, justified in §B.

Milestones are presented in **dependency order**, which is also close to reversibility order —
with one deliberate inversion recorded in §D.

## §B Tier justification — L

| Signal | Measurement |
|---|---|
| Files | **23** — statusline 4, `internal/cli` 2, doctrine 4, docs-site 12, `drift_cache.go` comment 1 |
| Packages touched | 3 (`internal/statusline`, `internal/cli`, `internal/spec`) |
| Cross-cutting schema change | 1 — an on-disk record whose shape and path both change, with a schema-version bump |
| Consumers outside this repo's code | 2 classes — an **always-loaded** doctrine read procedure consumed by agents and the orchestrator, and twelve published pages under a [HARD] four-locale synchronisation obligation |
| Milestones | 6, each independently shippable |

The file count alone exceeds the 15-file M ceiling. The always-loaded doctrine read procedure is
the second, independent L signal: changing a procedure every session loads is not an M-sized
change even when the diff is small.

## §C Producer / consumer split

This SPEC **is** the producer half of the split already taken. `SPEC-WEB-CONSOLE-015` consumes
the reader this SPEC exports and is deliberately not touched here (§D of `spec.md`).

The one carve-out considered and rejected: lifting the doctrine and docs-site sweep (M5, M6)
into a documentation SPEC. Rejected because the sweep's correctness is defined entirely by the
path this SPEC introduces — split out, its acceptance degrades to "the files mention a string",
which passes without observing that the string is the path anything actually writes. Keeping it
here is also what closes the parent SPEC's F5 defect, in which twelve enumerated files had no
requirement, no criterion, and no Definition-of-Done entry.

## §D Milestones

Each milestone leaves the tree green and adds no half-state.

**Review attention.** The decisions expensive to unwind are **M2** (the path and the key) and
**M3** (the payload shape and the schema-version bump). M1 precedes them for a mechanical safety
reason, not because it is more consequential — see the note under M1. M5 and M6 are
follow-the-shape sweeps.

### M1 — Consolidate onto one reader, before anything moves (prerequisite)

Export the reader (`context_usage.go:186` `readContextUsage`) and migrate
`internal/cli/tokens.go` onto it, deleting `tokensContextSnapshotFilename` (`:30`) and
`tokensContextSnapshot` (declaration `:81`, doc comment `:79`). Path unchanged in this
milestone.

**Why first, out of reversibility order.** `readTokensContextSnapshot` (`tokens.go:393-397`)
returns `nil` on any read error, so if the path moves while that reader still holds its own
filename constant, the context block silently vanishes from `moai tokens` output with no compile
error and no runtime error. Doing the consolidation first makes the path move (M2) a
single-place edit. The inversion is a hazard-avoidance decision, recorded here so it is not
mistaken for a claim that the reader shape matters more than the path.

Ships alone: one reader replaces two, behaviour unchanged.

### M2 — Per-session path and key (highest reversibility cost)

`context_usage.go:134` becomes `.moai/state/context-usage/<session-id>.json`, keyed by the
render payload's `session_id` — never by `.moai/state/current-session-id.txt`. Update
`builder.go:168`'s call site, and the statusline tests that assert the literal single-slot path.

`isFreshForSession` (`:236`) and its `writer_pid` discriminator become unreachable once the path
carries the identity; remove them rather than leave them as decoration. `sameSemanticPayload`
(`:203`) **stays** — it is the write throttle, not a validity guard.

Hard cut. No dual-write, no fallback read (D-3, `design.md` §1.2).

Ships alone: statusline and `moai tokens` both move together because M1 unified them.

### M3 — The record carries model and effort

Widen the record type (`context_usage.go:56`) to carry model and effort, bump
`contextUsageSchemaVersion` (`:27`), and rename the Go type to `sessionTelemetryRecord` — the
on-disk name does not change (D-1, `design.md` §1.1). Thread `input.Effort.Level` and the
resolved model into the `writeContextUsage` call at `builder.go:168`; the model value is
`resolveGLMModelName(displayName)`'s output (`metrics.go:197`), the same function the render
path already calls at `:51` (D-5, `design.md` §1.3).

Absent values are omitted, not defaulted; the reader reports them as not recorded.

Ships alone: additive fields with no console consumer are inert.

### M4 — Refuse a key that escapes the directory

The key is now a path component arriving from outside the process. Refuse — do not sanitise and
redirect — an empty value, a value containing a path separator, a parent-traversal value, or an
absolute-path value. The render must still complete: persistence is best-effort and never fails
a render, which is the package's existing contract.

Ships alone: a validation gate on a path introduced in M2.

### M5 — Doctrine sweep (four files, one change)

Update `.claude/rules/moai/workflow/context-window-management.md:100` and
`context-window-management-detail.md` §1-§2, plus both `internal/template/templates/…` mirrors,
to name the per-session path and to drop the single-slot validity guard the split makes
unreachable (the session-id equality check and the `writer_pid` discriminator, detail companion
§2). Run `make build`.

The pairs are byte-identical today, so a one-sided edit is mechanically detectable (AC-ST-009c).

Ships alone: documentation follows the behaviour landed in M2.

### M6 — Published documentation and the stray comment

Twelve docs-site pages — `content/{en,ko,ja,zh}/advanced/statusline.md`,
`advanced/token-budget.md`, `cli-reference/tokens.md` — updated in one change, three per locale,
with a warning-free site build. Fold in `internal/spec/drift_cache.go:24`, whose comment names
the old file as a sibling state file; NOTE-level, no behaviour depends on it.

Ships alone.

## §E Dependency graph

```
M1 ──► M2 ──► M3
        │
        ├──► M4
        ├──► M5
        └──► M6
```

- **M1 → M2**: the silent-break hazard above. Reversing this order is the one sequencing mistake
  that produces a green run with a broken command.
- **M2 → M3**: the payload widening lands in the record M2 relocated. The reverse order also
  works but touches the same struct twice.
- **M2 → M4/M5/M6**: nothing to validate, document, or publish until the path exists.
- M4, M5, M6 are mutually independent and may land in any order or together.

## §F Open verification items

Recorded as items to **measure during the run**, not as blockers. Neither changes the design if
it resolves the unfavourable way; both change what the run must do next.

1. **Does the widened payload degrade the write throttle?** `sameSemanticPayload`
   (`context_usage.go:203`) skips a write when the new record is semantically equal to the one on
   disk. Adding model and effort to the compared payload could, in principle, turn every render
   into a write. **Unmeasured.** The hypothesis is that it does not: both values change rarely
   within a session, so they are near-constant contributors to the comparison. They are not
   *fixed*, though — `.claude/rules/moai/workflow/cache-aware-execution.md` directive 10 records
   that model and effort are switchable mid-session — and that is what makes the fallback carry a
   cost of its own. Measure in M3 by counting writes across N renders with unchanged context
   values.

   **Both outcomes have a consequence; neither is free.** If the count rises, the fallback is to
   exclude the two new fields from the throttle comparison rather than from the record — but then
   a mid-session model or effort change is not persisted until an unrelated context value moves,
   so the record holds a value that is **present and wrong**. That state is worse than absence,
   because REQ-ST-003's "not recorded" path does not cover it: a reader cannot tell a stale value
   from a current one. So excluding the fields is only acceptable together with something that
   makes a changed value reach disk — comparing the two fields for inequality alone (not for
   throttle-payload equality), or forcing a write when either differs from the record on disk.
   Whichever the run picks, the run states it, because the choice is invisible in the output and
   only shows up as a wrong value much later.
2. **Does a GLM-backed session's payload carry `effort` at all?** **Unobserved** — no GLM session
   was running on this machine when the payload was captured, so the observation covers the
   Claude backend only. If GLM omits it, that session's effort is reported as not recorded through
   REQ-ST-003's own empty path; the design holds, but the console's effort cell would be
   permanently empty for GLM sessions, which is worth knowing before anyone reads it as a bug.
   Measure by capturing one render payload under `moai glm`.

## §G Resolved decisions

Ratified before authoring; recorded here as closed, with consequences in `design.md` §1.

- **D-1** — Model and effort ride the same per-session record as the context values, with a
  schema-version bump. The on-disk path name is unchanged; only the Go type name widens to
  `sessionTelemetryRecord`. *Rejected:* keeping them on `kanban.Record` (measured impossible —
  `spec.md` §A.2); a second per-session file (same write path, more files, one more thing to keep
  consistent).
- **D-3** — Hard cut, no dual-write window. The single slot **is** the defect; a compatibility
  window preserves it, and contradicts M5's own instruction to drop the single-slot validity
  guard.
- **D-5** — The recorded model is the backend-resolved `display_name`, one value.
  `resolveGLMModelName` (`metrics.go:197`) already takes `display_name`, already strips the `[1m]`
  suffix, and already substitutes the z.ai model. *Rejected:* `model.id` — Claude-shaped, so a
  GLM session has no such identifier and the suffix would need stripping again. *Rejected:*
  recording both — one console cell, two values, and the consumer left to re-choose.
- **A-001 keying** — The key is the identifier the session runtime delivered, never the
  project-wide sidecar (`spec.md` §A.3).

## §H Anti-patterns

- **Sanitising a hostile key instead of refusing it.** Rewriting `../escape` into `escape`
  produces a file that looks legitimate and belongs to no session. Refuse, and let the render
  complete without a record.
- **Leaving `isFreshForSession` in place "just in case".** Once the path carries the identity, its
  session-id equality check can never fail. Dead validation reads as live validation to the next
  person.
- **Grepping the doctrine files without their template mirrors.** The pairs are byte-identical
  today; updating one side is the failure mode AC-ST-009c exists to catch.
- **Treating the docs-site sweep as optional polish.** It is REQ-ST-009 with its own criterion,
  precisely because it was optional polish in the parent SPEC and would have shipped stale.
- **Moving the path before consolidating the readers.** See M1. The failure is silent.
- **Adding a reaper for stale per-session records in this change.** A disposal policy has its own
  liveness question and does not belong as a side effect of a path move (`acceptance.md` §E).

## §I Cross-references

- `.moai/reports/t207/spec-split-design.md` §2, §3, and its two appendices — the ratified split,
  the effort runtime observation, and the session-id investigation.
- `.moai/reports/t207/plan-audit-iter2-independent.md` F1, F5, F10, F13 — the defects this SPEC
  is written not to reproduce.
- `SPEC-WEB-CONSOLE-015` — the consumer of the reader this SPEC exports.
- `.claude/rules/moai/workflow/context-window-management.md` § Detection Heuristics and its
  `-detail.md` companion — the read procedure M5 must move, both mirrored under
  `internal/template/templates/`.
