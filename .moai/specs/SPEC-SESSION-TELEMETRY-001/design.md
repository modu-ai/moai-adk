# SPEC-SESSION-TELEMETRY-001 — Design

Scoped to what `spec.md` and `plan.md` do not already carry: the ratified decisions and their
consequences (§1), the migration sequence (§2), and the record-schema widening (§3). It does
**not** restate the blast radius (`spec.md` §A.4, §C.1) or the milestone breakdown (`plan.md`
§D).

## §1 Ratified decisions

### 1.1 D-1 — model and effort ride the same record

**Decision.** Model and effort are recorded on the per-session telemetry record this SPEC is
already creating, with a schema-version bump. The on-disk path name stays
`.moai/state/context-usage/<session-id>.json`; only the Go type name widens, to
`sessionTelemetryRecord`.

**Why this shape.** The values arrive on the same render input as the context values, and the
snapshot is written from the one function where the session id and the collected data are
simultaneously in scope (`builder.go:168`). One writer, one write, one file. No new producer, no
launcher wiring, no eight call sites.

**Why the path name does not change.** Sixteen files already point at that name — four doctrine
files and twelve published pages. Renaming it would move two things in one sweep, and each
additional thing moving in a sweep is another place a locale or a mirror can be left behind.
The type name is internal and costs nothing to widen; the path name is a public contract with
its own consumers and costs sixteen files.

**Consequences.**

- The record's meaning widens from "context usage" to "session telemetry", and the name no
  longer describes it. This is accepted, and recorded here so a later reader does not read the
  mismatch as an oversight. A rename is available later as its own change, with its own sweep.
- Because the payload widens, `contextUsageSchemaVersion` (`context_usage.go:27`) is bumped and
  the reader must tolerate records at the previous version — §3.
- The values are stable for a session's lifetime, which is why they are expected not to disturb
  the write throttle. Expected, not measured: `plan.md` §F item 1.

**Rejected.** *(a) Keeping model and effort on `kanban.Record`, written by the launcher.*
Measured impossible on both backends: `moai cc` never parses or sets a model, and `moai glm`
sets a four-slot map rather than one model (`spec.md` §A.2). This was the parent SPEC's design
and the measurement is what moved it here. *(b) A second per-session file for the two new
values.* Same write path, same key, twice the files, and a new way for the two halves of one
session's telemetry to disagree.

### 1.2 D-3 — hard cut, and what that forecloses

**Decision.** `.moai/state/context-usage.json` → `.moai/state/context-usage/<session-id>.json`,
with no dual-write window and no fallback read of the old path.

**Why this shape.** Three reasons, in the order that decided it.

1. **There is nothing to migrate.** The file is render-ephemeral: regenerated on every statusline
   render, and observed changing twice within eight minutes during this SPEC's authoring
   (`raw_pct` 55 → 56 → 57). A cut loses at most one render's snapshot.
2. **Every reader is in-tree and enumerated** (`spec.md` §A.4). No plugin, no external API, no
   persisted consumer.
3. **A dual-write window would keep the defect alive.** The single slot is not incidental to the
   bug; it *is* the bug. Writing to it during a compatibility window means the window's function
   is to preserve a known last-writer-wins race — and it would contradict M5's own instruction
   to drop the single-slot validity guard, leaving `isFreshForSession` and the `writer_pid`
   discriminator alive as decoration.

**Consequences.**

- `AC-ST-001`'s second half ("the single-slot path does not exist") is a legitimate assertion
  only under this decision. Under a dual-write window it would fail by construction. The
  criterion is written this way by decision, not by presupposition.
- A build straddling the change reads nothing rather than reading something stale: an old reader
  looking for the single slot finds no file and takes its existing fail-open path. That is the
  correct failure — absent, not wrong.
- The parent SPEC's G-3 reached the same decision on the same reasoning; it is restated here
  because this SPEC now owns the path move.

### 1.3 D-5 — one model value, backend-resolved

**Decision.** The recorded model is `resolveGLMModelName(model.display_name)` — one value.

**Why this shape.** The render payload delivers both `model.id` (`claude-opus-5[1m]`) and
`model.display_name` (`Opus 5 (1M context)`). `resolveGLMModelName` (`metrics.go:197`) already
takes `display_name`, already strips the `[1m]` suffix, and already substitutes the z.ai model
when the `ANTHROPIC_DEFAULT_*` slots carry non-Claude names — and the render path already calls
it at `:51`. Its output is, by construction, "the model this session actually runs".

**Consequences.**

- The recorded value is a display name, not an API identifier. A consumer wanting to key on a
  model programmatically has a human-facing string; that is the trade accepted for having one
  value that is correct on both backends.
- REQ-ST-004's `Where` clause is satisfied by an existing function rather than by new logic, so
  the GLM half of the requirement carries essentially no implementation risk — only test
  coverage.

**Rejected.** *(a) `model.id`.* Claude-shaped. A GLM session has no such identifier — what it
actually runs arrived through `ANTHROPIC_DEFAULT_*_MODEL` — and the `[1m]` suffix would need
stripping again, duplicating logic that already exists. *(b) Recording both.* One console cell,
two values, and the consumer left to re-choose — which relocates the decision rather than making
it.

### 1.4 The key is the session's own identifier

**Decision.** The record is keyed by the render payload's `session_id`. The statusline does not
read `.moai/state/current-session-id.txt`.

**Why this shape.** The payload's `session_id` is, structurally, the id of the session doing the
rendering — it cannot name another session. The sidecar can: it is a single project-wide slot
written by whichever SessionStart ran last, which is the *same* last-writer-wins defect this
SPEC exists to remove. Sourcing the key from it would put the bug inside the fix.

**Consequences.**

- A consumer needing a session's id gets it from the session registry
  (`.moai/state/active-sessions.json`), which SessionStart writes with the same runtime id.
- `AC-ST-002`'s two-id fixture is the pin: payload id `S`, sidecar id `T`, and the file must be
  named for `S`.
- This SPEC therefore takes no position on repairing the sidecar, and takes no dependency on it
  being repaired.

## §2 Migration sequence

The order below is the one thing in this SPEC where getting the sequence wrong produces a green
run and a broken command.

1. **Export one reader and migrate `moai tokens` onto it** — path unchanged. After this step,
   exactly one place in the tree knows the filename.
2. **Move the path** to `.moai/state/context-usage/<session-id>.json`, keyed by the payload id.
   Update `builder.go:168`'s call site and the tests asserting the literal old path. Delete
   `isFreshForSession` and the `writer_pid` discriminator; keep `sameSemanticPayload`.
3. **Widen the payload** — model and effort, schema-version bump, type rename.
4. **Gate the key** — refuse an empty, separator-bearing, traversing, or absolute key; the render
   still completes.
5. **Move the doctrine** — four files in one change, then `make build`.
6. **Move the published documentation** — twelve pages, three per locale; fold in the
   `drift_cache.go:24` comment.

Step 1 before step 2 is the load-bearing constraint. `readTokensContextSnapshot` returns `nil`
on any read error (`tokens.go:393-397`), so a path move ahead of the consolidation removes the
context block from `moai tokens` output with no compile error, no runtime error, and no failing
test — the failure mode is a feature quietly disappearing. Steps 4-6 may be reordered freely
among themselves.

## §3 Record-schema widening

The record gains two string fields, both omitted when empty, and the schema version is bumped.

**Absent-field tolerance is a reader obligation, not a writer one.** Three populations of record
will exist on disk simultaneously during any rollout:

| Population | Shape | Reader behaviour |
|---|---|---|
| Written by the pre-change build | previous schema version, no model, no effort, at the single-slot path | Not read at all — the path is gone. No fallback (D-3) |
| Written by the post-change build, full payload | current version, both fields present | Both values reported |
| Written by the post-change build, partial payload | current version, one or neither field present | The present value reported; the absent one reported as **not recorded** |

The second and third rows are the ones the reader must actually handle, and the third is the one
REQ-ST-003 exists for: a payload that omits `effort` or `model` yields a valid record with an
honest gap, never a defaulted value and never an error. `AC-ST-010` pins the tolerance using a
fixture produced by marshalling the pre-change struct, so the bytes are the marshaller's own
rather than hand-authored.

**Why omitted rather than empty-string.** An absent key and an empty string are
distinguishable on the wire and should stay so: the record says "this render did not carry a
model", not "this session's model is the empty string". The distinction is what lets a consumer
render "not recorded" without guessing.

**Why the version is bumped even though the change is additive.** The version's job is to let a
future reader know which fields it may expect. An additive change that leaves the version
unchanged makes "field absent" ambiguous between "this writer omitted it" and "this writer
predates it" — and REQ-ST-003 requires both to be reported the same way, which is only safe if
the reader is not asked to tell them apart. Bumping it keeps that option open for a later reader
that does need to.
