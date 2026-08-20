## Follow-up audit — four stale claims corrected, one of them mine

A full sweep of every v3.1.1 surface against `README × 4` and `docs-site × 4` turned up 9 gaps and **6 stale claims**. Stale claims outrank gaps — a gap leaves a reader with nothing, a stale claim sends them the wrong way — so those were fixed first, in `023562c3e` and `ea76a118b`.

### The one this PR introduced

The v3.1.1 highlights added earlier in this PR asserted that `cache.yaml` **governs** cache-breakpoint lifetimes. It does not. `internal/config/cache_config.go` records that the `cache_control` injector was never reachable from production and was removed, that `LoadCacheConfig` has zero callers (`grep -rn 'LoadCacheConfig' --include='*.go' . | grep -v _test` returns only the defining file), and that prompt caching is performed by Claude Code itself — with an `@MX:CEILING` warning that the file *"misleads as soon as a user expects the values to change behaviour"*. The paragraph did exactly that, in four locales.

Removed from the highlights. The config-table row stays — the file ships and shows up in the `moai web` settings editor — but now says plainly that nothing reads the values. Its old placement claim (`session_ttl` on session-start context, `spec_ttl` on SPEC bodies) came from the same removed injector and went with it.

Root cause worth recording: the change inventory found a new config file and I classified it as a new capability without checking whether anything consumed it. File existence read as feature existence.

### The other three

| Claim | Reality |
|---|---|
| `kanban-mode.md:237` — the factory lead "polls the backlog queue" | `internal/hook/session_start_factory.go:98` states in capitals that queue polling is **deliberately not taught** to the factory lead; doing so hands it a second, conflicting polling protocol. The sentence also contradicted itself — lead polls, *and* routes what the operator picked. `launchers.md:36` and `manager-lead.md:35` were already correct, so the docs disagreed with each other too. |
| `moai-todo.md:77` — `moai todo done` is the only way to remove an item, and nothing sets `dropped` | `drop`, `undrop`, `edit`, `move` all exist (`internal/cli/todo_drop.go:66,149`, `todo_edit_move.go:45,110`). The page was actively telling users a command they need does not exist. |
| `skill-guide.md` — 32 skills, 6 domain | 34 and 7. The READMEs were corrected earlier in this PR, so the two surfaces had started disagreeing. Derived token arithmetic updated with the totals; the core/optional split was also wrong (19 → 21). |

`config-sections.md` additionally documented `statusline.yaml` without its new `forge` key and had no `crosssession.yaml` section at all. Both added — with `isolate_machines: false` (no approval before a message leaves this machine) given its own row rather than buried in prose.

### Deliberately not done

- **No `cache.yaml` docs-site page.** Nothing reads those values; a page would multiply the misleading claim across four more locales. `cost-optimization/prompt-caching.md` correctly describes Claude-Code-side caching and was left alone.
- **`MOAI_SESSION_PID` is not a gap.** `internal/cli/launch_session_pid.go:12` says a hook must never set it — it is stamped by the launcher. No user configures it, so it does not belong in user docs.

### Still open

- **`moai migrate profiles` is undocumented** in all 8 files. `internal/cli/migrate_profiles.go:362` describes the failure it repairs: anyone who has run `moai cc -p <name>` has project memory split in two, and *"a session launched one way cannot recall what a session launched the other way learned"*. Users hitting this have no way to discover the fix. Highest-value remaining gap.
- **Cross-session messaging has no docs-site page** in any locale, only doctrine in `.claude/rules/`.
- **`internal/template/templates/.moai/config/sections/cache.yaml` contradicts itself** — line 12 comments "Ships disabled; flip to true to opt in", line 13 is `enabled: true`, and the web console string matches the comment. A code/template question rather than a docs one, so untouched here.
- Five lower-priority Factory Mode facts (`/clear` between cards, class collapse in-lane, release-branch integration discipline, the free-slot notice line, the lane join line) remain undocumented on both surfaces.

Verification unchanged and re-run: READMEs 816 lines × 4 with identical H2 offsets, `hugo --printPathWarnings` warning-free, 161 files per locale with empty path diffs.

🗿 MoAI
