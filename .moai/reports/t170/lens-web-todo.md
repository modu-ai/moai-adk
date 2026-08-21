# t170 lens — web settings surface + the todo feature's on/off switch

Read-only investigation. Every claim carries file:line. Gaps are named, not guessed.

## Summary

- The web console has a **fully declarative field registry**: `settings.FieldDef` records in `internal/settings/schema_sections.go`. Adding a new boolean toggle to an already-routed section is a ONE-LINE schema addition — parsing, rendering, validation, and persistence are all generic and derive from that record. No handler edit, no templ edit, no form wiring.
- The exemplar to copy is **`workflow.branch_guard.enabled`** (`internal/settings/schema_sections.go:334`) — a feature-enable bool in `workflow.yaml`, written through the yamlpatch seam, rendered on the Git & Worktree tab, with a typed runtime reader in `internal/config`.
- The `feedback` section is **NOT** currently on the web console. It has a surviving `FieldDef` (`feedback.repository`) but is `RouteExcluded`, has no tab, and no render meta — a feedback field is the *harder* of the two additions and needs 3 extra edits (route entry + tab + panel meta). There is also a latent inconsistency worth noting (§A.3).
- The todo feature spans **9 surfaces**, of which the two that would leak guidance while "off" are a shipped **always-loaded rule** (`kanban-dispatch.md`) and the **skill listing metadata** (`/moai todo` in the moai SKILL dispatch table). Neither reads config. This is the hard part of the card's "no guidance when off" requirement (§ Verdict).
- Precedent for a flag that gates *guidance* rather than code exists and is exact: **`handoff.guide`** (`.moai/config/sections/handoff.yaml:12-15`) gates a stderr hint only, and **`mcp.tools.<name>.enabled`** is the default-ON-when-absent read pattern the todo flag should copy verbatim.

---

# PART A — the web settings surface

## A.1 One boolean toggle traced end to end

Tracing `workflow.branch_guard.enabled` — a bool in a seam section, i.e. exactly the shape a `todo.enabled` toggle would take.

### Step 1 — schema/field declaration (the ONLY hand-written step)

`internal/settings/schema_sections.go:334`

```go
s(SectionWorkflow, "workflow", TypeBool, "workflow", "branch_guard", "enabled"),
```

`s` is `seamField`; the variadic tail is the yamlpatch key path from the document root. The record type is `FieldDef` (`internal/settings/schema.go:164-181`) carrying `Name / Section / Type / Options / EmptyLabel / Validate / Default / StoreOnly / EmptySubmits / I18nKey / Description / Persist`. `Persist` is a `PersistTarget` (`internal/settings/schema.go:145-152`) whose `Kind: PersistSeam` + `Section` (file base name) + `Path` ([]string key path) fully determine the write.

The comment at `schema_sections.go:330-333` is worth reading — it documents that the distributed template ships *without* the `branch_guard` block, so the key is absent until first edit, and "the seam writer (yamlpatch) upserts the nested mapping on first edit". That is the same absent-key situation a new `todo` block would be in.

### Step 2 — tab + panel placement (render meta)

The field is placed on a tab by name, not by section:

- `internal/web/schemaform.go:156` — `isWorktreeFieldName` returns true for `workflow.worktree.*` **and** the literal `workflow.branch_guard.enabled`.
- `internal/web/schemaform.go:180-193` — `partitionWorkflowFields()` splits the workflow section's fields across three panels (workflow / git-worktree / audit).
- `internal/web/schemaform.go:33-69` — `consoleTabs()` is the ordered tab list; `{ID: "git-worktree", LabelKey: "tab.git-worktree.title", ...}` at :47.
- `internal/web/schemaform.go:196+` — `schemaSectionMetas()` returns the per-panel `schemaSectionMeta{ID, PanelID, Icon, TitleKey, DescKey, Fields, Extras, ...}` (type at `schemaform.go:122-149`).

Note the documented invariant at `schemaform.go:115-121`: **`PanelID` is a render concern, `ID` is the persistence unit** — moving a field between tabs must not reclassify its section.

### Step 3 — HTML rendering (generic, no per-field code)

`internal/web/fieldsets.templ:247` `templ schemaFieldWidget(view pageView, f settings.FieldDef)` switches on `f.Type`. The bool branch emits the hidden companion at `internal/web/fieldsets.templ:366`:

```
<input type="hidden" name={ name + "__present" } value="1"/>
```

This companion is what distinguishes "unchecked → false" from "not submitted → preserve" (unchecked checkboxes are simply absent from a form POST).

### Step 4 — HTTP handler receiving the change

`internal/web/handlers.go:398`

```go
schemaEdits, schemaErrs := parseSchemaForm(r)
```

`parseSchemaForm` (`internal/web/schemaform.go:277-336`) iterates **`settings.AllFields()`** and skips anything that is not `PersistSeam`/`PersistTypedSection` (`schemaEditableField`, `schemaform.go:266-268`). The bool branch (`schemaform.go:286-294`):

```go
case settings.TypeBool:
    if r.PostFormValue(f.Name+"__present") == "" {
        continue // 미제출 → preserve
    }
    if r.PostFormValue(f.Name) != "" { edits[f.Name] = "true" } else { edits[f.Name] = "false" }
```

Per-field errors join an atomic-reject set; on reject the submitted values are echoed back over current values by `overlaySchemaEdits` (`handlers.go:450`, impl `schemaform.go:405-413`).

### Step 5 — the config write

`internal/web/handlers.go:508`

```go
if err := a.applySchemaEdits(a.cfg.ProjectRoot, schemaEdits); err != nil {
```

The injection point is `internal/web/app.go:139` (`applySchemaEdits: settings.ApplySchemaEdits`).

`settings.ApplySchemaEdits` (`internal/settings/sectionapply.go:30-81`) resolves each name back to its `FieldDef`, buckets by `Persist.Kind`, and for seam fields calls `WriteSectionViaSeam(projectRoot, section, edits)` — one `yamlpatch.PatchFile` per file, atomic per file (`sectionapply.go:70-79`).

`WriteSectionViaSeam` (`internal/settings/sectionwrite.go:57-70`) enforces two gates before touching disk:
1. `RouteForSection(section) != RouteSeam` → error;
2. every edit's `Path[0]` must be an allowed top-level key for that file (`sectionRootKeys`, `sectionwrite.go:31-53`).

Then patches `.moai/config/sections/<section>.yaml` via `yamlpatch.PatchFile` (node surgery — preserves comments and unmodeled keys; the rationale is the `@MX:WARN` at `sectionwrite.go:15-22`).

### Step 6 — read-back for render

`settings.SchemaCurrentValues` (`internal/settings/sectionvalues.go:71-95`) walks every field's `fieldYAMLPath` in the section doc; absent file or key → empty string, never an error.

### Step 7 — runtime consumer (typed)

`workflow.branch_guard.enabled` has a typed mirror: `internal/config/types.go:397` (`BranchGuard BranchGuardConfig`), struct at `types.go:563`, default at `internal/config/defaults.go:740`. **The seam write and the typed read are two independent surfaces over the same YAML** — the seam does not need the struct, but a runtime gate does.

## A.2 Which file is the write seam — and is there a field registry?

| File | Role |
|---|---|
| `internal/settings/schema_sections.go` | **The declarative field registry.** `seamSectionFields()` starts at :292; the workflow block is :303-360. A new key is one `s(...)` line here. |
| `internal/settings/sectionroute.go` | **Routing SSOT.** `sectionRoutes` map at :55-99 decides seam / typed / statusline / excluded. Unlisted name → `RouteExcluded` (zero value) → write refused. |
| `internal/settings/sectionapply.go` | **Dispatcher.** `ApplySchemaEdits` :30 buckets by `Persist.Kind` and calls the seam or typed writer. |
| `internal/settings/sectionwrite.go` | **The seam write itself.** `WriteSectionViaSeam` :57; per-file allowed top-level keys `sectionRootKeys` :31. |
| `internal/settings/sectionvalues.go` | Read side only (`SchemaCurrentValues` :71, `RawBlockValues` :99). No write path. |
| `internal/settings/nested.go` | Not on this path — I did not find it referenced by the schema seam. Named in the task brief; **gap**: I did not trace what consumes it (see Gaps). |

**Answer to "is there a declarative FieldDef list a new key must be added to":** yes — `internal/settings/schema_sections.go`, function `seamSectionFields()` (`:292`), plus `SchemaSectionIDs()` (`:630`) if the section is new.

## A.3 Is there an existing web settings entry for `feedback`?

**No — and the state is inconsistent, which matters for the card.**

What exists:
- A live `FieldDef`: `internal/settings/schema_sections.go` — `s(SectionFeedback, "feedback", TypeText, "feedback", "repository")` (in the `// feedback.` block of `seamSectionFields()`).
- The section id: `SectionFeedback` (`internal/settings/schema.go:38`).
- `SectionFeedback` is still listed in `SchemaSectionIDs()` (`internal/settings/schema_sections.go:637`).
- Allowed root key registered: `"feedback": {"feedback": true}` (`internal/settings/sectionwrite.go:37`).
- The config file exists and holds exactly one key: `.moai/config/sections/feedback.yaml` → `feedback.repository: modu-ai/moai-adk`.

What does **not** exist:
- **No route.** `"feedback"` is absent from `sectionRoutes` and explicitly listed in `ExcludedSections()` (`internal/settings/sectionroute.go:133`), pinned by `internal/settings/sectionroute_test.go:27` (`"feedback": RouteExcluded`). SPEC-WEBCONF-SIMPLIFY-001 M3 removed its tab and its write path (`sectionroute.go:91-99`).
- **No tab.** `consoleTabs()` (`internal/web/schemaform.go:33-69`) has no feedback entry.
- **No panel meta.** `grep -n "SectionFeedback\|feedback" internal/web/schemaform.go` returns **zero** matches.

**Latent inconsistency (flag it):** `parseSchemaForm` iterates `settings.AllFields()` and admits any `PersistSeam` field regardless of route (`internal/web/schemaform.go:281-284`). `feedback.repository` is `PersistSeam`. Nothing renders it, so it is never submitted in practice — but a hand-crafted POST carrying `feedback.repository` would reach `ApplySchemaEdits` → `WriteSectionViaSeam` → hard error `section "feedback" is not seam-writable`, which surfaces to the user as `"profile preferences saved, but section config write failed: ..."` (`internal/web/handlers.go:509-511`) — an atomic-partial outcome (the profile write already happened). This is a pre-existing condition, not something the card creates, but adding a *second* feedback field without re-routing the section would make it reachable through the UI.

**Cost of putting a new setting on the `feedback` section:** 4 edits (schema field + `sectionRoutes` entry + `consoleTabs` entry + `schemaSectionMetas` panel) plus i18n keys, and it must un-do a deliberate SPEC-WEBCONF-SIMPLIFY-001 decision. **Cost of putting it on `workflow`:** 1-2 edits (schema field + optionally a tab-partition predicate line). See the verdict.

---

# PART B — the todo feature and its on/off switch

## B.4 What the todo feature IS — every surface

Nine surfaces. Only #1 and #2 are code; the rest are guidance text or derived display.

| # | Surface | Location | Reads config? |
|---|---|---|---|
| 1 | **CLI** `moai todo` (add/list/done/next/edit/move/drop/undrop) | `internal/cli/todo.go` (cmd ctor `:163`, registered unconditionally at `:512` `rootCmd.AddCommand(newTodoCmd())`); siblings `todo_drop.go`, `todo_edit_move.go`; store `internal/kanban.BacklogStore` | **No** |
| 2 | **Queue store + path** | `internal/cli/todo.go:44` → `.moai/state/kanban/backlog.json`; non-git fallback resolved at `todo.go:98` | No |
| 3 | **Skill body** `/moai todo` | `.claude/skills/moai/workflows/todo.md` (mirrored at `internal/template/templates/.claude/skills/moai/workflows/todo.md`) — loaded on demand | No |
| 4 | **Skill dispatch table** (routing metadata) | `.claude/skills/moai/SKILL.md:6` (subcommand list), `:81` (verb table), `:105` (semantic routing exemplars: "add to the backlog", "remind me to"), `:166-172` (section) | No |
| 5 | **Slash command stub** | `.claude/commands/moai/todo.md` | No |
| 6 | **Always-loaded rule** | `.claude/rules/moai/workflow/kanban-dispatch.md` — `moai todo add` is the [HARD] sole-producer rule under "Entry into the board is an operator act"; 3 `todo` mentions; companion `kanban-dispatch-detail.md`. **This file is explicitly always-loaded** ("Intentionally always-loaded. A session learns it is the kanban lead from the SessionStart context, not from a file path") | No |
| 7 | **Foreman skill** (bare `/loop` driver) | `.claude/skills/moai-kanban-foreman/SKILL.md` — `allowed-tools:` includes `Bash(moai todo:*)` at `:17`; driver `.claude/loop.md:4,17` | No |
| 8 | **SessionStart hook notice** | `internal/hook/session_start_kanban.go:180` appends `backlogSummary`; 4-locale strings at `internal/hook/session_start_kanban_i18n.go:81/103/125/147` — e.g. ``"Kanban backlog: %d waiting — run `moai todo` to view the queue."``. Count via `queuedBacklogCount` (`session_start_kanban.go:187+`) | **No**, but gated on `MOAI_KANBAN*` env (kanban mode only) and on `source == "startup"` (`session_start_kanban.go:86-90`) |
| 9 | **Statusline segment** | `internal/statusline/renderer.go:188-190` → `🔄 TODO: %d/%d`; data at `internal/statusline/backlog.go:41-60`; segment key `SegmentBacklog = "backlog"` (`internal/statusline/types.go:349`) | **Partly** — `isSegmentEnabled` (`renderer.go:87-96`) returns **true when the key is absent**, and `backlog` is NOT in `.moai/config/sections/statusline.yaml` (16 segments listed, `backlog` absent). So it is on-by-default and already suppressible by adding `backlog: false`. It also self-suppresses when `Backlog.Available == false` (missing/corrupt file → fail-open empty) |

Surfaces #3-#7 are Markdown that a user turning the feature off would still see. #6 is the one that is *always* loaded.

## B.5 Where queue state lives, and which section should own an enable flag

**State:** `.moai/state/kanban/backlog.json` under the PRIMARY checkout (`internal/cli/todo.go:41-44`); a linked worktree resolves back to the primary queue (`resolveTodoQueueRoot`, `todo.go:81-99`). Non-git projects fall back to `~/.moai/todo/<project-key>/backlog.json` (documented at `.claude/skills/moai/workflows/todo.md:19-24`). `.moai/state/` is project-local and not committed.

**Measured, not assumed — what `workflow.yaml` actually holds.** Read in full; under a single `workflow:` root:
- scalar knobs: `default_mode`, `execution_mode`, `context_folding`
- **nested feature blocks with an `enabled` bool**: `auto_clear.enabled: true`, and — the direct precedent — `branch_guard.enabled: true` as the file's last block
- policy tables: `model_routing`, `model_routing_profiles`, `workflow_agents`, `token_budget`, `loop_prevention`, `agentic_loop`, `worktree.*`

So a `workflow.todo.enabled` block is **shape-consistent with what the file already contains** — sibling `*.enabled` feature gates already live there (`auto_clear.enabled`, `branch_guard.enabled`, plus schema-declared `codex.review_gate.enabled` which is absent from the shipped YAML by design).

Comparable feature-enable flags that already exist, by section:

| Flag | Section file | Route | On web console? |
|---|---|---|---|
| `workflow.branch_guard.enabled` | workflow.yaml | RouteSeam | **Yes** — Git & Worktree tab |
| `workflow.worktree.auto_create` (+ auto_cleanup/auto_merge/tmux_preferred) | workflow.yaml | RouteSeam | Yes |
| `workflow.codex.review_gate.enabled`, `workflow.codex.task.allow_write` | workflow.yaml | RouteSeam | Yes (rendered on the MCP tab, `internal/web/schemaform.go:165-170`) |
| `mcp.tools.<name>.enabled` × 17 | mcp.yaml | RouteSeam | Yes — dedicated MCP panel |
| `harness.auto_detection.enabled`, `learning.enabled` | harness.yaml | **RouteExcluded** | No (tab removed) |
| `observability.enabled` | observability.yaml | **RouteExcluded** | No |
| `handoff.guide` | handoff.yaml | **RouteExcluded** | No |
| `statusline.segments.<key>` × 16 | statusline.yaml | RouteStatusline | Yes |

**Only four sections are both seam-writable and web-rendered today: `workflow`, `mcp`, `crosssession`, `report`** (`internal/settings/sectionroute.go:55-99` crossed with `consoleTabs()`). Everything else is excluded.

## B.6 Precedent for a flag that gates GUIDANCE rather than code

Two, and both are useful.

**(a) `handoff.guide` — the closest match in kind.** `.moai/config/sections/handoff.yaml:12-15`:

```yaml
    # guide: when true, emit a best-effort stderr hint on non-/clear session
    #        starts (startup / resume / compact) that a saved handoff is waiting.
    #        Purely informational; never blocks the session.
    guide: false
```

Consumed at `internal/hook/handoff_inject.go:302` (`return mode, c.Handoff.Guide`). It gates **only a hint** — nothing functional. This is exactly "turning it off suppresses a prompt surface". Caveat: `handoff` is `RouteExcluded` (`internal/settings/sectionroute.go:133`), so it is not togglable from the web console today — it is a precedent for the *semantics*, not for the *console wiring*.

**(b) `mcp.tools.<name>.enabled` — the default-ON read pattern to copy.** `internal/cli/mcp_server.go:378-437`, `readMCPToolEnablement`. Its documented truth table is precisely what the card's "default is ON" needs:

```
//   - projectDir empty                      → all enabled
//   - file missing / unreadable             → all enabled
//   - YAML parse error                      → all enabled
//   - `mcp.tools` block absent              → all enabled
//   - `mcp.tools.<name>.enabled: false`     → that tool DISABLED
```

Implementation shape (`mcp_server.go:411-437`): seed the map to `true`, then a small hand-rolled `yaml.Unmarshal` into an anonymous struct with `Enabled *bool` — the pointer is what distinguishes "absent" from "explicitly false". The `@MX:NOTE` at `:410` records the fail-OPEN choice deliberately, contrasted against the codex gates' fail-CLOSED posture.

**(c) Weaker precedent, but real:** `statusline.segments.<key>` — display suppression, key-absent-means-enabled (`internal/statusline/renderer.go:87-96`). The `backlog` segment already rides this and is already suppressible.

**What I did NOT find:** any config flag that suppresses **skill listing metadata, a slash command, or an always-loaded rule file**. I grepped the hook package and the config types; guidance suppression today reaches stderr hints, statusline segments, and MCP tool registration — never Markdown that Claude Code loads. That gap is the card's hardest requirement (verdict below).

---

# Verdict

## Setting 1 — the todo enable flag

**Recommend `workflow.yaml` → `workflow.todo.enabled` (RouteSeam).** Why:

1. It is the **only** section that is simultaneously seam-writable, already web-rendered, and already the home of sibling `*.enabled` feature gates. Any other choice (`feedback`, a new `todo.yaml`) requires re-opening a route SPEC-WEBCONF-SIMPLIFY-001 deliberately closed, or registering a brand-new section in five or six places.
2. The change is genuinely one line in the registry: `s(SectionWorkflow, "workflow", TypeBool, "workflow", "todo", "enabled")` next to `internal/settings/schema_sections.go:334`. Everything downstream — parse, `__present` companion, validation, atomic reject, yamlpatch upsert of the absent nested mapping, read-back — is generic.
3. The runtime read should copy `readMCPToolEnablement`'s `*bool` default-ON shape, not a plain `bool` — a plain bool cannot tell "absent" from "false", and the shipped template will not carry the block.
4. Placement: it lands on the **workflow** tab by default (the `default:` arm of `partitionWorkflowFields`, `internal/web/schemaform.go:186-193`). If it should sit elsewhere, add one predicate line alongside `isWorktreeFieldName` at `:156`.

**Do not** invent a new `todo.yaml`: `sectionRootKeys` (`sectionwrite.go:31`), `sectionRoutes` (`sectionroute.go:55`), the `SectionID` const block, `SchemaSectionIDs()`, `consoleTabs()`, and `schemaSectionMetas()` would all need entries, plus i18n keys — six edits for a single bool.

## Setting 2 — the feedback-related setting

I do not know the card's exact second setting (**gap** — I was given the surfaces, not the card text). Two shapes:

- **If it belongs semantically to `feedback.yaml`:** it costs 4 edits and requires re-routing `"feedback"` from `RouteExcluded` to `RouteSeam` (`internal/settings/sectionroute.go`), which contradicts the recorded SPEC-WEBCONF-SIMPLIFY-001 M3 decision and will break `internal/settings/sectionroute_test.go:27` and `internal/web/scope_contract_test.go:79`, both of which pin `feedback` as excluded. Those tests are the enforcement of a decision, so this is a decision to reverse explicitly, not a test to update quietly.
- **Cheaper alternative:** if the setting is behavioural rather than about the feedback *repository slug*, host it under `workflow.yaml` alongside the todo flag and leave the `feedback` route closed. One tab, one panel, two toggles, zero reversals.

Either way, note §A.3's latent inconsistency: `parseSchemaForm` will pick up `feedback.repository` from a forged POST today and produce a hard write error. Re-routing `feedback` to `RouteSeam` incidentally fixes that; leaving it excluded leaves it as-is.

## Where "no todo guidance when off" is hard to satisfy

Ranked by difficulty:

1. **`.claude/rules/moai/workflow/kanban-dispatch.md` — always-loaded, unconditional, and it carries the `[HARD]` "the lead is the queue's sole producer / `moai todo add`" rule.** Claude Code loads rule files by path, not by config; there is no mechanism in this repo that suppresses a rule file based on a YAML flag. A user who turns todo off still has this rule in every session's context. **No existing pattern solves this.** Options, none free: (a) accept it as out of scope and say so in the SPEC; (b) move the todo-producer clause into the lazy companion `kanban-dispatch-detail.md` so the always-loaded stub only points at it; (c) have the SessionStart hook emit a counter-instruction when the flag is false — which adds noise rather than removing it.
2. **`.claude/skills/moai/SKILL.md` L1 metadata** (`:6`, `:81`, `:105`, `:166-172`). The skill listing is what makes `/moai todo` discoverable and routes natural-language "add to the backlog" phrasing. Skill listings are runtime-assembled from frontmatter; nothing in this repo gates a listed subcommand on config.
3. **`.claude/commands/moai/todo.md`** — a slash command stub is a file's presence. Same class of problem.
4. **`internal/cli/todo.go:512`** — the CLI command is registered unconditionally. Gating registration is straightforward Go (read the flag and hide the command), but consider whether hiding a CLI verb is desirable: the foreman skill's `allowed-tools` (`moai-kanban-foreman/SKILL.md:17`) and the queue file both still exist.

Genuinely easy to satisfy:

5. **SessionStart backlog notice** (`internal/hook/session_start_kanban.go:180`) — already conditional (kanban env + `source=="startup"`); adding a flag check is a one-line guard on an existing conditional.
6. **Statusline `🔄 TODO:` segment** (`internal/statusline/renderer.go:188`) — already gated by `isSegmentEnabled(SegmentBacklog)`; either document `backlog: false` in `statusline.yaml` or have the todo flag feed the same predicate.
7. **`.claude/skills/moai/workflows/todo.md`** — loaded on demand, so an off flag that stops the router from reaching it effectively suppresses it.

**Recommendation on scope:** state plainly in the SPEC that the flag suppresses the *runtime* surfaces (5, 6, 7, and optionally 4) and that the *always-loaded rule* and *skill-listing* surfaces (1, 2, 3) are a separate, harness-level problem with no precedent in this repo — rather than writing an acceptance criterion ("no todo guidance appears when off") that cannot be satisfied as literally worded.

---

## Gaps

- I did not trace `internal/settings/nested.go` — it was named in the brief but does not appear on the schema-seam write path I followed. Its likely consumers are `parseProjectNestedForm` / `writeProjectNestedConfig` (`internal/web/handlers.go:393`, `:497`), a separate curated-6-field seam; I did not read it.
- I did not read the t170 card text, so the identity of the second setting is inferred from the lens name (`feedback`) and the sibling lenses (`init`, `masking`).
- I ran no tests and no build. Every claim above is a file read, not an execution result.
- `internal/web/widget_policy_test.go:209` lists `"feedback.repository": true` in an allowlist; I did not read that test to determine which policy it encodes.
- I did not confirm which file owns the `f.<field>` / `fieldDesc.*` i18n keys, nor whether `i18n_governance_test.go` requires a new field to register them. Assume a new field needs i18n keys added.
- Surface count "nine" is what I found by grepping `.claude/`, `internal/`, and `internal/template/templates/` for `moai todo` / `backlog queue`; SPEC and report files under `.moai/specs/` and `.moai/reports/` also mention todo but are historical records, not live guidance surfaces.
