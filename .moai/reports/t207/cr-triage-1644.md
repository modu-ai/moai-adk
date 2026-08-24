# CodeRabbit triage — PR #1644 (card t207)

- PR: modu-ai/moai-adk#1644 · head `035c4ab00`
- Findings: **29 inline** (`gh api .../pulls/1644/comments`) + a walkthrough body. Enumerated all 29.
- Triaged by the orchestrator directly: the delegated triage agent died on an account spend limit
  after writing only its header, so this file is a first-hand triage, not a relayed one.

## On staleness — the review's pin versus the comments' anchor

The review body's `Merge Risk: 🟡 Moderate · up to cd8e9` names the **pre-merge** head, so CodeRabbit
analysed `cd8e9e9e0` and did **not** review the merge commit `035c4ab00`. That is what the lead
measured, and it is correct.

But every one of the 29 inline comments carries `commit_id == 035c4ab00`:

```
$ gh api "repos/modu-ai/moai-adk/pulls/1644/comments" --paginate --jq '[.[] | .commit_id] | unique'
["035c4ab004eb52334c7378be892904a04906ac61"]
```

GitHub re-anchored them onto the current head, meaning each cited line still exists there. The two
facts are compatible and mean different things: **what was analysed** is `cd8e9`, **where the
comments point** is the current head. The practical consequence is that the findings are actionable
as written — I verified each against the current tree rather than against the diff CodeRabbit saw.

## Blocking — must fix before merge (4)

### B1. `internal/kanban/todo_root.go` — adoption writes to a path no consumer reads

**PROVEN by experiment**, not by reading. Every consumer resolves the store through
`BacklogPathForRoot(root)` = `<root>/.moai/state/kanban/backlog.json`
(`internal/cli/todo.go:49`, `internal/web/todo_queue_read.go:33`). Adoption writes
`filepath.Join(fallbackRoot, "backlog.json")` = `<root>/backlog.json`
(`todo_root.go:155`), and `fallbackTodoQueueRoot`'s "populated fallback wins" check tests that same
wrong path (`:133`).

A throwaway test seeded a project-local queue at the consumer path, called
`ResolveTodoQueueRootAdopting`, and then looked for the queue where consumers read:

```
resolved root = …/001/.moai/todo/002-3cd2aa91
CONSUMER READ PATH MISSING: …/002-3cd2aa91/.moai/state/kanban/backlog.json (no such file or directory)
ADOPTION WROTE HERE INSTEAD:  …/002-3cd2aa91/backlog.json
```

So in the home-fallback branch the operator's cards are moved to a path `moai todo` never reads:
the queue reads empty afterwards while the file sits one directory tree away. The existing tests
did not catch it because they assert against the same wrong convention.

Severity: **blocking.** This is the data-integrity half of the SPEC's own premise.

### B2. `internal/kanban/todo_root.go:86` — the adopting path does not read through when adoption fails

`adoptLocalTodoQueue` is best-effort and returns early on `MkdirAll` / `Rename` + `ReadFile`
failure. `ResolveTodoQueueRootAdopting` returns the home root regardless, so `moai todo` reports an
empty queue — while the pure `ResolveTodoQueueRoot` reads through to the project-local root under
the same preconditions and reports the real cards.

That is exactly the console-versus-CLI divergence ratified decision D-2 exists to eliminate
(`SPEC-WEB-TODO-QUEUE-001` REQ-WTQ-005), reappearing in the adoption-failure branch, which no
criterion exercises.

Severity: **blocking.** A ratified requirement is violated on a reachable branch.

### B3. `internal/web/factory_lanes.go:141` — the record's lane number overrides the registry's

`if rec.Lane > 0 { row.Lane = rec.Lane }` lets a joined record relabel the row. A mismatched record
— precisely what the join's own guards exist to catch — renders `lane-5` as "lane 3", and two lanes
joining records that both claim lane 3 render two rows numbered 3 while the registered lane
disappears from the display.

The registry lane is the row's identity; a `rec.Lane` that disagrees with it is evidence of a bad
join, which `REQ-WC15-047` already says must render unresolved.

Severity: **blocking.** It defeats the misattribution guard the SPEC was written around.

### B4. `internal/web/viewmodel_ops.go:730` — factory records make the chain read as present

`buildChain` sets `Present: len(records) > 0` (`:281`) over **all** records, but renders only
`ChainRoles`. `buildKanban` passes the complete record set, factory lanes included (`:722`). A
project running only factory lanes therefore shows a chain that is "present" with four idle roles
and `IdleRole = "lead"` — a false "the chain stopped at lead" reading for a project that has no
chain at all.

Severity: **blocking.** It is a wrong answer on a supported configuration, which is the failure
class this SPEC's §A.4 exists to prevent.

## Non-blocking — fixing in the same pass (6)

| # | Where | What | Why fix now |
|---|---|---|---|
| N1 | `internal/web/app.go:162`, `viewmodel_ops.go:83-87`, `widgets.templ` | Korean comments in Go source | HARD repo rule: code, comments, and godoc are English (`CLAUDE.local.md` §3). Regenerating `widgets_templ.go` follows. |
| N2 | `internal/web/screens.templ:335` | `todoStateBadge` renders `queued`/`picked`/`dropped` untranslated | `REQ-WTQ-008` requires every user-visible string in all four locale maps; a state badge is user-visible. |
| N3 | `internal/kanban/todo_root_test.go:37`, `internal/web/todo_section_test.go:273` | literal `/dev/null` for git config overrides | Unix-only; on Windows the fixtures inherit the runner's global git config. `os.DevNull` is portable. |
| N4 | `internal/statusline/session_telemetry_key_test.go:73` | snapshot compared by length, not content | A refused key that both removes and adds one file passes. The snapshots are sorted, so content comparison is exact. |
| N5 | `internal/web/app.go:163` | the `/todo` 405 test covers only POST and accepts a 403 | It does not prove the route returns 405. Cover POST/PUT/PATCH/DELETE and assert `Allow: GET`. |
| N6 | `.moai/reports/t207/*.md` (6 files, 38 occurrences) | machine-absolute home-directory paths in a **public** repository | `gh repo view --json visibility` → `PUBLIC`. Cheap to sanitize and permanent once merged. Session ids and PIDs are left: they are ephemeral and they are the evidence the audits rest on. |

## Deferred — not this agent's to edit (SPEC bodies, manager-spec's ownership)

Seven findings land on `spec.md` / `plan.md` / `acceptance.md` / `design.md` bodies, which the
ownership matrix reserves to `manager-spec` (`spec-frontmatter-schema.md` § Forbidden ownership
crossings). Editing them here would repeat the boundary violation this card already had to repair
once. They are listed for a follow-up delegation:

- `SPEC-SESSION-TELEMETRY-001/design.md:15` — the design describes the record type in its
  unexported spelling while the contract and implementation use `SessionTelemetryRecord`.
- `SPEC-SESSION-TELEMETRY-001/acceptance.md:101` — fenced blocks without info strings (MD040).
- `SPEC-KANBAN-RECORD-SESSION-KEY-001/spec.md:16` + `acceptance.md:20,65` — frontmatter field
  ordering, non-portable runtime-state paths in criteria, and a criterion that should also assert
  `T.json` was neither created nor changed.
- `SPEC-WEB-CONSOLE-015/plan.md:124` — the tier item still reads as unresolved after the operator
  ruled M; `acceptance.md:206` — a Definition-of-Done item left unchecked at close.
- `SPEC-WEB-TODO-QUEUE-001/spec.md:160` — the read-only contract wording versus the actual console
  path.
- `SPEC-FACTORY-MODE-001/acceptance.md:284` — CodeRabbit reads the supersede annotation as leaving a
  contradictory active assertion. **Judgement: this one is a REJECT**, and it is recorded here rather
  than acted on. The annotation is the ratified outcome of an operator ruling: it deliberately keeps
  the criterion's text and marks the record half superseded, precisely so the invalidation is not
  silent. Removing the assertion is what the ruling declined to do.

## Documentation findings — accepted, deferred to a docs pass

`docs-site/content/en/advanced/statusline.md:163` is the substantive one: it describes a snapshot
lifecycle the readers do not implement (no `SessionStart` handler reads a previous snapshot; no
stale/mismatch rejection). `en/cli-reference/web.md:45` (405 wording), `ko/advanced/token-budget.md:127`
(evidence-path divergence), and `ja/advanced/moai-web-console.md:28` (a Japanese subject-phrasing
correction) follow. All four-locale, so a docs pass rather than a one-line edit.

## What I could not verify

- The walkthrough body's two outside-diff notes were read but are summaries of the inline set rather
  than distinct findings; I did not treat them as separate rows.
- The docs findings are accepted on CodeRabbit's reading of the prose; I did not re-derive the
  statusline lifecycle claim from the code, so their severity is provisional.
