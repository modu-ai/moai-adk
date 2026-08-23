# SPEC-WEB-CONSOLE-015 — Research

Ground truth measured in worktree `.claude/worktrees/t207` at base commit `28bde4022`. Every
row was read from the file named, not inferred.

## §1 Live-refresh transport — already complete

| Fact | Location |
|---|---|
| `Hub`, `Publish`, `ServeEvents` | `internal/web/events.go:42`, `:58`, `:70` |
| `text/event-stream` header | `internal/web/events.go:81` |
| 25s keepalive ticker | `internal/web/events.go:95`, `:108` |
| 250ms debounce constant | `internal/web/events.go:22` |
| `Hub.Watch` over fsnotify | `internal/web/events.go:117` |
| Watch skips absent dirs (fail-open) | `internal/web/events.go:130` |
| Watch errors are non-fatal, browser falls back | `internal/web/events.go:163` |
| Watch is wired at server start | `internal/web/server.go:251` |
| `/events` route | `internal/web/app.go:162` |
| `EVENTS` list, `POLL_MS = 30000` | `internal/web/assets/app.js` (~line 636) |
| `connect()` opens `new EventSource("/events")` | `internal/web/assets/app.js` |
| `es.onerror` → `startPolling()` at `failures >= 3` | `internal/web/assets/app.js` |
| `ready` listener → `stopPolling()` | `internal/web/assets/app.js` |
| Events carry no payload; each triggers `refresh(area)` re-fetch | `internal/web/assets/app.js` |
| htmx 2.0.4 core has no SSE extension — documented reason for hand-wired `EventSource` | `internal/web/assets/app.js:626-627` |

Conclusion: the card's "SSE vs polling — pick one" framing does not describe this tree.
Polling is the degraded mode of SSE, reachable only from `es.onerror` or a missing
`window.EventSource`.

## §2 Session telemetry gaps

`internal/web/viewmodel_ops.go:250-256` — three literal placeholders with comments naming the
prerequisite ("3단계"): `Model: ""`, `Effort: ""`, `ContextPct: -1`.

`internal/kanban/record.go:55-99` — `Record` fields: `SessionID`, `SpecID`,
`Role` (`omitempty`), `Backend`, `EnteredAt`, `DeepScanDir`, `VerifyRung` (`*Rung`,
`omitempty`), `VerifyReentries`.

`internal/kanban/record.go:45` — `@MX:ANCHOR` warning that the launcher, the orchestrator, and
the sync-phase dedup gate all bind to these JSON keys; a renamed key breaks unseen readers.

`internal/kanban/record.go:116-130` — `WithRole` lowercases and trims, accepts `RoleLead`,
`RoleLane`, or a companion role, and **silently discards** anything else.

`internal/cli/kanban.go:472` — `recordKanbanSession(specID, backend, role string)`; calls
`kanban.WriteBestEffort(projectRoot, kanban.NewRecord(...).WithRole(role))`. Eight callers:
`internal/cli/cc.go:161,175,192,208` and `internal/cli/glm.go:224,237,250,264`.

Model / effort producers that exist but are not threaded in: `internal/config/profile.go:73-76`
(`ModelEffort{Model, Effort}`, `EffectiveProfile`) and `internal/settings/agentfm`
(`AgentInfo{... Model, Effort ...}`).

## §3 Context-usage single-slot hazard

`internal/statusline/context_usage.go:134` — the write path is
`filepath.Join(stateDir, "context-usage.json")`, one file per project root.

`internal/statusline/context_usage.go:56-65` — `contextUsageRecord` is **unexported**, carrying
`schema_version`, `session_id`, `writer_pid`, `captured_at`, `context_window_size`,
`tokens_used`, `raw_pct`, `stage`, `band`.

`internal/statusline/context_usage.go:186` — `readContextUsage` is **unexported**. Any
non-statusline reader therefore needs either an exported reader or a duplicated struct;
the SPEC requires the former (REQ-WC15-021/022).

Single-slot support functions that the split makes partly unreachable:
`sameSemanticPayload` (`:203`), `isRealSessionID` (`:216`), `isFreshForSession` (`:236`),
`contextUsageFresh` (`:255`).

Observed race (prior investigation, `.moai/reports/webredesign/moai-web-menu-spec.md` §4.6):
read A showed session `368a2bd9…` at 260,000 tokens; read B showed session `e463a3c9…` at 0.
Last writer wins.

Consumers of the path, enumerated by grep across the tree:

- `internal/statusline/builder.go:157` (call site) and `context_usage.go` (writer/reader)
- `internal/statusline/builder_test.go:1819,1857,1860`,
  `internal/statusline/context_usage_test.go:16,19,52,86,347`
- `.claude/rules/moai/workflow/context-window-management.md:100` — names the file as the
  authoritative snapshot and defines the read procedure (session-id match, `writer_pid`
  discriminator, freshness check)
- `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`
  — confirmed present, the Template-First mirror
- docs-site: `content/{en,ko,ja,zh}/advanced/statusline.md`, `.../advanced/token-budget.md`,
  `.../cli-reference/tokens.md` — 12 files
- `.moai/README.md`

## §4 Backlog queue

`internal/kanban/backlog_store.go` — `BacklogItem{ID, Text, AddedAt, SpecID *string, State}`
at `:66-71`; `BacklogState` ∈ `queued|picked|dropped` at `:51-60`; `BacklogFinding` at `:123`;
`BacklogRecord{Version, LastSeq, Items, Findings}` at `:154`. Exported and read-safe:
`NewBacklogStore` (`:240`), `BacklogPathForRoot` (`:249`), `QueuedBacklogCountForRoot` (`:277`),
`Load` (`:299`). Mutating: `Mutate` (`:341`), `Add` (`:387`), `acquireLock` (`:419`).

`internal/cli/todo.go:66` — `resolveTodoQueueRoot()` is unexported. It resolves through
`gitcore.ResolveGitDirs(base).CommonDir`'s parent — the primary checkout from any worktree —
with a home-based fallback (`fallbackTodoQueueRoot`, `adoptLocalTodoQueue`) when git cannot
answer. Its comment records the measured failure: 30 queued cards on the primary, "queue is
empty" from a linked worktree, 2026-08-17.

`internal/web` contains no reference to the backlog. Confirmed by grep.

`SPEC-KANBAN-TODO-CLI-001` frontmatter reads `status: in-progress`, `module: internal/kanban`,
`tier: M` — it owns this store.

## §5 Factory lanes

`internal/kanban/factory_slots.go` — `FactoryWorkerEntry{PID int, RegisteredAt string}` at
`:37`; `FactoryRegistryPath` (`:47`), `LoadFactoryRegistry` (`:55`, fail-open),
`PruneFactoryDeadClaims` (`:84`), `FactoryFreeSlots` (`:99`). Liveness only — no card, no spec,
no stage.

`internal/session/registry.go:86-95` — `Entry{SessionID, SpecID, Phase, StartedAt,
LastHeartbeat, PID, Host, CWD}`, schema frozen per REQ-COORD-002 / REQ-COORD-024. `PID` is
present, so the `workers.json → registry → Record` join closes on existing data.

`internal/kanban/role.go:42` — `RoleLane = "lane"`, a bare constant with no lane number, and
deliberately not a `CompanionRoles` member.

`internal/config/envkeys.go:167-173` — `EnvMoaiKanbanID` is the **run** identifier
(e.g. `tk4ntu`), lead-owned, generated once per run. Set at `internal/cli/factory.go:255` and
`internal/cli/kanban.go:173`. It is not a card id, and no per-lane card id exists on disk.

`internal/web/viewmodel_ops.go:46` — `ChainRoles = []string{"lead","plan","run","sync"}`; the
view iterates only these, so lanes are invisible today.

`internal/web/viewmodel_ops.go` `estimateStage` — maps `StateLive → StageActive` (estimated),
`StateStale → StageWait` (estimated), default → `StageBlocked` (not estimated, "세션 없음은
추정이 아니라 사실이다"). `RoleVM.StageEstimated` carries the honesty flag.

## §6 Template-First applicability

`internal/template/templates/` contains `.claude/`, `.moai/`, `.codex/`, `.git_hooks/`,
`.github/`, `AGENTS.md`, `CLAUDE.md`, and root config — **no** `internal/`. Confirmed by `ls`.
So the Go changes carry no mirror obligation; only the doctrine-rule edit does, and its mirror
is confirmed present.

## §7 Web routes and i18n

Routes at `internal/web/app.go:157-192`: `/` (overview), `/kanban`, `/monitor`, `/settings`,
`/events`, `/save`, `/specs`, plus profile/GLM-key/shutdown/static.

`internal/web/assets/i18n.js` — 2,508 lines, four locale maps opening at `en:27`, `ko:646`,
`ja:1267`, `zh:1888`. Nav keys `nav.overview` / `nav.kanban` / `nav.specs` / `nav.monitor` /
`nav.settings` exist in each. Governance is enforced by
`internal/web/i18n_governance_test.go` and `i18n_untranslated_allowlist_test.go`.

## §8 Not consulted

The two screenshots attached to card t207 were not opened. `current-01~03.png` referenced by
the prior investigation's §9 was likewise not opened. Nothing in this SPEC derives from image
content.
