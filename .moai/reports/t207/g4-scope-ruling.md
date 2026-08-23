# t207 — G-4 scope ruling: the event vocabulary stays unchanged

Companion to `plan-audit.md` (iteration 1). Records why SPEC-WEB-CONSOLE-015 keeps an exclusion the audit warned would have to move, so a later reader does not read the exclusion as the warning being ignored.

## What the audit warned

The plan-auditor recommended a panel on `/kanban` over a separate `/todo` route, on a concrete ground: AC-WC15-034 keys the todo section on the existing `data-live="kanban"` refresh area, and a separate route "requires an event-vocabulary change that §D Out of Scope explicitly forbids".

## What the operator decided

A separate `/todo` route, against that recommendation. A legitimate product call: the backlog is the surface the lead reads to pick the next card, and a route makes it addressable; a panel subordinates the queue to the board it actually feeds.

## What the dispatch then inferred, and why that was withdrawn

Iteration 2 was dispatched with an instruction to move the vocabulary change **into** scope — reasoning that the route decision made the auditor's named cost unavoidable rather than avoided, and that leaving a route-based AC contradicting an Out-of-Scope entry would reproduce the D3/D4 self-contradiction shape.

That inference was wrong, and measurement rather than argument is what withdrew it:

1. **The existing event already covers the route.** `watchMap["kanban"]` watches `.moai/state/kanban` (`internal/web/events.go:29`), and the backlog file is `.moai/state/kanban/backlog.json` — inside it.

2. **The refresh marker binds a region, not a route.** `refresh(area)` gates on `document.querySelector('[data-live="' + area + '"]')` and then re-fetches `window.location.href`. A `/todo` page carrying `data-live="kanban"` is refreshed by the existing event with no producer change.

3. **A new event would be actively harmful.** It would have to watch the same directory, and `eventFor` (`events.go:171-183`) attributes a change to the longest registered watch path with a strict comparison:

   ```go
   if len(p) > bestLen {
       best, bestLen = name, len(p)
   }
   ```

   Two entries of equal length therefore tie, and the winner is decided by Go's randomized map iteration order. That is precisely the nondeterminism the file's own header comment (`events.go:6-9`) records as a previously fixed bug — a change under one watched directory matching a different event name on different runs, which made the debounce test fail nondeterministically.

Adding a known race to remove a bookkeeping contradiction would make the code worse to make the document tidier.

## The ruling

The operator's route decision **stands**. The consequential scope the dispatch had inferred from it is **retracted**. The §D exclusion is not deleted but narrowed: it now states why it survives the route decision, and `AC-WC15-034` pins it mechanically with an unchanged-diff assertion on `watchMap` and `EVENTS` rather than leaving it to assumption.

Approved by the lead after review; not taken unilaterally.

## What the route did pull into scope

Enumerated in spec.md §C.6 and asserted by AC-WC15-035, each verified in the t207 worktree:

| Surface | Change |
|---|---|
| `internal/web/app.go` `routes()` | the `/todo` route |
| `internal/web/shell.templ` `rail()` | a sixth `@navRow` (measured: five today — overview, kanban, specs, monitor, settings) |
| `internal/web/icons.templ` `iconAt` | a `todo` case — `navRow` passes the nav id straight to `@iconAt`, so a missing case renders a blank glyph rather than failing |
| `internal/web/screens.go` | the `Area` value |
| `internal/web/assets/i18n.js` | `nav.todo` in all four locale maps |
| generated | `shell_templ.go`, `icons_templ.go` |

## Residual limitation, recorded not fixed

`Hub.Watch(root, …)` watches the *served* project root, while REQ-WC15-031 resolves the backlog to the *primary checkout*. A console served from a linked worktree renders the primary's queue correctly but receives no live event when that queue changes — and the 30-second fallback poll does not engage, because SSE is healthy. The section is correct on load and on any other `kanban` event, and stale in between.

Widening the watched **paths** is a producer change distinct from the vocabulary question above, and this SPEC declines it (spec.md §D). The lead's ruling: a stale render is a limitation, not data corruption, and it does not justify breaking the Tier L requirement ceiling of 25 to fit it into M3-M6. It becomes a separate card if the operator wants it.
