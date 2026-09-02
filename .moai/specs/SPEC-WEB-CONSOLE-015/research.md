# SPEC-WEB-CONSOLE-015 — Research (reference material)

> **Not part of the required artifact set.** Reclassified from Tier L to Tier M in version 0.3.0;
> Tier M carries three artifacts. Retained rather than discarded because the measurements it records
> were paid for and remain the provenance for spec.md's claims. Note that its measurements predate
> the §A.4 correction in 0.3.0 — where the two disagree about how consistently a record is keyed to
> the wrong session, spec.md §A.4 is the current statement.

The measurement record for what this SPEC still asserts after the three-way carve-out. Everything
below was measured in worktree `.claude/worktrees/t207` at `dfbf828a6`, in the run that authored
version 0.2.0. Facts belonging to the sibling SPECs are not repeated here; each of them carries its
own record.

## §1 Live-refresh transport — already complete

| Fact | Location |
|---|---|
| 250ms debounce constant | `internal/web/events.go:22` |
| `text/event-stream` header | `internal/web/events.go:81` |
| 25s keepalive ticker | `internal/web/events.go:95` |
| `POLL_MS = 30000` | `internal/web/assets/app.js:638` |
| `startPolling()` definition | `internal/web/assets/app.js:700` |
| `new EventSource("/events")` | `internal/web/assets/app.js:721` |
| `startPolling()` called at `failures >= 3` | `internal/web/assets/app.js:743` |

Polling is the degraded mode of SSE, reachable only from `es.onerror` or a missing
`window.EventSource`. The card's "SSE vs polling — pick one" framing does not describe this tree.

## §2 The three placeholder cells and the marker that already renders them

`internal/web/viewmodel_ops.go:250-256` — `RoleVM` is built with `Model: ""`, `Effort: ""`,
`ContextPct: -1`, each with a comment naming the prerequisite.

`internal/web/screens.templ:165-175` — the model and effort cells already branch:
`if r.Model != "" { … } else { @missing() }`, and the same for effort.
`internal/web/widgets.templ:122-124` — `templ missing()` renders `—` with
`title="not recorded anywhere yet"` and an aria-label.

Consequence for version 0.1.0's REQ/AC-WC15-012: the "not recorded" marker it required is
implemented, wired, and rendering for every row today, because `viewmodel_ops.go:253` hardcodes the
empty string. Its second clause asserted "no empty `<td>`"; measured,
`grep -c "<td" internal/web/screens.templ` returns **0** — the view is `div`/`span`-based and emits
no table cell under any implementation. Both halves of the criterion were unobservable, which is
why it was deleted rather than reworded.

## §3 The note banner

`internal/web/screens.templ:192`:

```
@noteBanner("info", "Stage is estimated from heartbeat. Model, effort and context usage are not
recorded yet, so they are left blank — they fill in once kanban.Record is extended.", "")
```

`internal/web/widgets.templ:40-52` — `templ noteBanner(kind, text, key string)` emits
`<span data-i18n={ key }>` only when `key != ""`, otherwise a bare `<span>`. The call above passes
`""`, so this string is untranslated in a view whose other strings carry keys.

## §4 Lanes are invisible, and the console knows nothing about the factory

```
$ grep -rn "Factory\|factory" internal/web --include='*.go' --include='*.templ' | grep -v _test
(no output)

$ grep -rn "\"lane\"\|RoleLane" internal/web --include='*.go' --include='*.templ'
(no output)
```

`internal/web/viewmodel_ops.go:46` — `ChainRoles = []string{"lead", "plan", "run", "sync"}`; the
view iterates only these.

`internal/kanban/factory_slots.go` — `FactoryWorkerEntry` at `:37` (`PID`, `RegisteredAt`),
`FactoryRegistryPath` at `:47`, `LoadFactoryRegistry` at `:55` (fail-open),
`PruneFactoryDeadClaims` at `:84` (a separate call). Liveness only — no card, no spec, no stage.

`internal/web/viewmodel_ops.go:409-435` — `loadSessions` reads `active-sessions.json` once per
render, unmarshals into `[]session.Entry`, and maps each to
`SessionVM{ID: shortID(...), SpecID, State, Heartbeat, Cwd}` — **`PID` is dropped**, and `ID` is
shortened. The full session id survives only as the `byID` map key.

`internal/web/viewmodel_ops.go:266-275` — `estimateStage` maps `StateLive → StageActive` (estimated),
`StateStale → StageWait` (estimated), default → `StageBlocked` (not estimated, with the comment
"세션 없음은 추정이 아니라 사실이다"). `RoleVM.StageEstimated` (`:92`) carries the flag.

## §5 The record is not keyed by the session it describes — observed

Three live sessions in the primary checkout's registry:

```
$ python3 -c "…" .moai/state/active-sessions.json
2beac221-1716-459d-9dc0-8e8c5951b8b3  pid 15207  cwd …/.claude/worktrees/t219
c15d8434-be5f-4276-9e62-4758c1156368  pid 51045  cwd …/.claude/worktrees/t210
3db058e1-2692-44f5-9e0d-45b543bb3c1f  pid 36912  cwd …/.claude/worktrees/t207
```

None of the three has a kanban record:

```
$ ls .moai/state/kanban/{2beac221…,c15d8434…,3db058e1…}.json
ls: …/2beac221-….json: No such file or directory
ls: …/3db058e1-….json: No such file or directory
ls: …/c15d8434-….json: No such file or directory
```

The one record that names a live session names the **lead**, and carries a lane's role:

```
$ cat .moai/state/kanban/d281730e-a47e-4f82-878e-5fd0ddc4dcb9.json
{ "session_id": "d281730e-…", "spec_id": "", "role": "lane", "backend": "claude",
  "entered_at": "2026-08-23T17:47:22Z", "deepscan_dir": "", "verify_reentries": 0 }
```

So `workers.json[lane].PID → active-sessions.session_id → record` resolves to nothing, or to
another session's record. Version 0.1.0 §A.5's "closes on today's data" is withdrawn on this
evidence. **Not observed:** whether the deployed chain view visibly mis-attributes a row — the
console was not started, so §5 proves the on-disk join is broken and says nothing about the render.

## §6 The single-slot telemetry file — observed twice, minutes apart

```
$ cat .moai/state/context-usage.json
{ "schema_version": 1, "session_id": "d281730e-a47e-4f82-878e-5fd0ddc4dcb9",
  "writer_pid": 58721, "captured_at": "2026-08-24T13:22:50…", "context_window_size": 1000000,
  "tokens_used": 600000, "raw_pct": 60, "stage": "soft", "band": "large" }
```

A reading of the same file minutes earlier (recorded in the dispatch that commissioned this
revision) held `writer_pid 41575, raw_pct 55` for the same session id. One slot, rewritten by
whichever session renders; the other three live sessions are unreadable by construction. This is
the observation the dependency on `SPEC-SESSION-TELEMETRY-001` rests on — an observation in this
run, not a citation of the earlier investigation's May reading.

## §7 Read-only baseline for `internal/web`

```
$ grep -rnE 'os\.(WriteFile|MkdirAll|Rename|Create)\(|WriteBestEffort\(|acquireLock\(|SaveFactoryRegistry\(|\.Mutate\(' \
    internal/web --include='*.go' | grep -v '_test\.go'
internal/web/profile_crud.go:38:	return os.MkdirAll(dir, 0o755)
internal/web/profile_crud.go:64:	return os.Rename(src, dst)
```

Both derive their paths from `profile.GetProfileDir` (`profile_crud.go:33-64`) — the Claude profile
directory — and neither from the project state directory. Version 0.1.0's AC-WC15-002 asserted this
grep "returns zero", which is false as literally commanded; the corrected criterion asserts an
exhaustive inventory of exactly these two lines instead.

```
$ grep -rn "internal/statusline" internal/web --include='*.go'
rc=1   (no output)

$ grep -rn "context-usage\|ContextUsage" internal/web --include='*.go' --include='*.templ'
internal/web/viewmodel_ops.go:255:	ContextPct: -1, // 3단계: context-usage/<session-id>.json 분리 후
```

The console has no telemetry reader and no copy of the record's schema today — the baseline for
AC-WC15-021a.

## §8 The launcher does not know the session's model — carried, not re-measured

The measurement that closed audit finding F1 (`internal/cli/cc.go:36` is a help string, and
`internal/cli/glm.go:350-353` sets a four-slot map rather than one model) was taken by the split
design and is cited in spec.md §A.2 rather than re-run here, because the conclusion it supports —
that the producer is not the launcher — belongs to `SPEC-SESSION-TELEMETRY-001`. This SPEC asserts
only that its own producer is elsewhere.

## §9 Not consulted, not observed

- The two screenshots attached to card t207 were not opened. Nothing here derives from image
  content.
- `moai web` was not started. Every claim about the current render is derived from source, not from
  a running console (see §5).
- The GLM backend's statusline payload was not observed; whichever way that measurement falls, it
  belongs to `SPEC-SESSION-TELEMETRY-001`.
