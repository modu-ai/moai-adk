# SPEC-WEB-CONSOLE-015 — Implementation Plan

## §A Context

One package: `internal/web`. No producer is touched — every producer obligation version 0.1.0
carried moved to `SPEC-SESSION-TELEMETRY-001` or `SPEC-KANBAN-RECORD-SESSION-KEY-001`, and the
todo axis to `SPEC-WEB-TODO-QUEUE-001` (spec.md §D).

Both dependency SPECs must land before this one's criteria can be satisfied; the work below is
implementable against test fixtures before they land, but not verifiable end-to-end (spec.md §A.5).

Milestones are ordered by **decision reversibility**: M1 fixes the view-model shape and the join
policy, which are expensive to unwind once tests and templates bind to them; M3 is
follow-the-shape work.

## §B Tier — declared L, and the honest measurement says it is near the M/L boundary

The frontmatter keeps `tier: L`, carried over from version 0.1.0. Measured against what actually
remains, that is generous, and the reclassification is flagged rather than taken here — the
operator holds it.

| Signal | Version 0.1.0 | This revision |
|---|---|---|
| Packages touched | 4 (`web`, `kanban`, `cli`, `statusline`) | **1** (`internal/web`) |
| Cross-cutting schema change | 2 (record fields, state-file path) | **0** |
| Always-loaded doctrine surface | 4 files (two mirror pairs) | **0** |
| Published documentation | 12 docs-site pages | **0** |
| Source files | > 10 | **7** — `viewmodel_ops.go`, `screens.templ`, `widgets.templ`, `assets/i18n.js`, and three test files |
| Generated files following | 2 | **2** — `screens_templ.go`, `widgets_templ.go` |
| Milestones | 6 | **3** |

**What still argues for L:** the SPEC has two hard dependencies and its correctness is a join
across three state files whose keying is currently broken (spec.md §A.4), so the failure mode is
silent mis-attribution rather than a compile error. That is a Tier-L-shaped risk on a
Tier-M-shaped diff.

**What argues for M:** one package, no schema change, no doctrine surface, seven files.

A Tier M reclassification is defensible and would drop `design.md` and `research.md` from the
required artifact set and the audit threshold from 0.85 to 0.80. Both files exist and are current,
so reclassifying costs nothing already spent. **Recommendation: reclassify to M** if the operator
agrees the dependency risk is carried by the dependency SPECs' own audits rather than by this
one's tier.

## §C Milestones

Each milestone leaves the tree green and adds no half-state.

### M1 — lane view model and the registry join (highest reversibility cost)

Three decisions land here, and all three are cheaper to get right than to change later.

1. **Where lanes live in the view model.** A separate lane collection beside the chain-role
   iteration, not a widening of `ChainRoles` (`viewmodel_ops.go:46`). Lanes are not chain roles;
   the four-role chain is a fixed dispatch vocabulary and widening it would make every chain
   consumer defend against a variable-length role list.
2. **How the join reaches the process identifier.** `loadSessions` (`viewmodel_ops.go:409-435`)
   currently maps `session.Entry` to `SessionVM` and **drops `PID`** — it keeps id, spec, state,
   heartbeat, and cwd. The join needs it, so either `SessionVM` gains the field or the lane
   builder reads the registry entries directly. Prefer widening `SessionVM`: one read of
   `active-sessions.json` per render stays one read.
3. **The duplicate-identifier policy (REQ-WC15-047).** A plain `map[pid]session` lookup attributes
   one session to two lanes silently. Build the lookup so a duplicate collapses to "ambiguous"
   rather than to a winner — a count-then-resolve pass, not a last-write-wins map.

Ships alone: a lane section that renders unresolved rows until the dependency SPECs land is
correct output, not a half-state (spec.md §C.4).

Closes REQ-WC15-043 / -044 / -045 / -046 / -047 and their criteria.

### M2 — telemetry cells

Fill the three `RoleVM` placeholders (`viewmodel_ops.go:250-256`) from the reader
`SPEC-SESSION-TELEMETRY-001` exports, keyed per session. Import that reader; declare nothing that
restates its record's fields (REQ-WC15-021). Keep `@missing()` (`widgets.templ:122-124`) for the
absent case — the honest rendering already exists and is already wired (spec.md §A.3).

Depends on `SPEC-SESSION-TELEMETRY-001` for the reader symbol. Closes REQ-WC15-021 / -023 / -051.

### M3 — note banner and locale keys

Rewrite the kanban note banner (`screens.templ:192`) so it asserts neither that the values are
unrecorded nor that the kanban record produces them, and give it a translation key — its third
argument is currently the empty string, which `noteBanner` (`widgets.templ:40-52`) renders as a
bare `<span>` with no `data-i18n` attribute. Add every key this SPEC introduces to the four locale
maps.

What the banner should still say: the stage is estimated from the heartbeat. That half stays true
and is the honesty flag REQ-WC15-045 requires.

Depends on M1 and M2 only in the sense that the sentence must match what the view then shows.
Closes REQ-WC15-050 / -052.

## §D Dependency graph

```
SPEC-SESSION-TELEMETRY-001 ─────────► M2 ─┐
SPEC-KANBAN-RECORD-SESSION-KEY-001 ─► M1 ─┴─► M3
```

## §E Anti-patterns to avoid

- **Re-deciding the transport.** The card asks for a choice that was already made and built
  (spec.md §A.1).
- **Declaring a second copy of the telemetry record's schema** in `internal/web` instead of
  importing the exported reader. Two declarations of one on-disk schema is how a format forks —
  and version 0.1.0 measured that exact failure already present between the statusline and
  `moai tokens`.
- **Widening `ChainRoles` to carry lanes.** M1 decision 1.
- **A plain `map[pid]session` lookup.** M1 decision 3; it fails REQ-WC15-047 silently.
- **Adding a "not recorded" marker.** It exists (`screens.templ:165-175`, `widgets.templ:122-124`)
  and renders today. Version 0.1.0's REQ/AC-WC15-012 required building it, which is why that
  criterion passed on the untouched tree and observed nothing.
- **Reaching into a producer to fix a join.** If the join comes up empty, the fix is in
  `SPEC-KANBAN-RECORD-SESSION-KEY-001`, not here.
- **Running the full Go suite locally.** Target `internal/web` and read CI for the full-suite
  verdict.

## §F Open items for the operator

1. **Tier** — §B recommends M; the declared value stays L until the operator rules.
2. **Landing order** — the two dependency SPECs are independent of each other and can run in
   parallel; this SPEC is last regardless.

## §G Cross-references

- `SPEC-SESSION-TELEMETRY-001` — the per-session telemetry record and its one exported reader.
- `SPEC-KANBAN-RECORD-SESSION-KEY-001` — the record's session keying, lane number, and card
  identifier.
- `SPEC-WEB-TODO-QUEUE-001` — the `/todo` route and the queue-root resolution, carved out of this
  SPEC.
- `.moai/reports/t207/spec-split-design.md` — the ratified three-way carve-out this revision
  implements.
- `.moai/reports/t207/plan-audit-iter2-independent.md` — findings F2, F7, F9, F11, F12, F13, closed
  here; F1, F3, F4, F5, F6, F10 left with the carve-outs.
- `.moai/reports/webredesign/moai-web-menu-spec.md` — the prior console investigation; check its
  claims against code before reusing them.
