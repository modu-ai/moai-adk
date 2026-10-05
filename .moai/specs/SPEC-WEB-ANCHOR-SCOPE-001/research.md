# SPEC-WEB-ANCHOR-SCOPE-001 — research.md

Plan-phase research: corpus measurement, discriminator derivation, classification table.
All measurements below were taken in THIS run, on THIS tree, cited with command + verbatim output.

---

## 1. Corpus definition and raw-superset measurement

**Corpus**: every `internal/web/*_test.go` file (the web test package, 82 files —
measured on this tree by both `ls internal/web/*_test.go | wc -l` and
`find internal/web -name '*_test.go' | wc -l`, each printing 82).
**Tree**: `bf779ecf2` (= worktree HEAD `WT-anchor-scope-sweep` = develop tip; mirror
present — `git merge-base --is-ancestor 1aaf4951f HEAD` exit 0, verified by lane
orchestrator this session; 1aaf4951f = "docs(t509): closed — landed on develop").

Measurement (this run):

```
$ git rev-parse --short HEAD
bf779ecf2
$ grep -o 'strings\.Index(' internal/web/*_test.go | wc -l
      60
$ grep -l 'strings\.Index(' internal/web/*_test.go | wc -l
      19
```

**Raw superset = 60 sites / 19 files.** Context figures from other trees are NOT
restated as measurements: lead's 55/18 was measured @ 0b1e27877 (pre-t509, unfiltered
`strings.Index(` = 55/18 with the 34/15 being the body|html|page-receiver subset);
lane-1's 60/19 was its own post-t509 worktree. This SPEC's 60/19 matches lane-1's count
on an equivalent post-t509 tree — expected, since t509 added `panelHTML` call sites but
did not remove the body-wide `strings.Index(` it repaired (the helper itself contains 4).

## 2. Render-surface map — where the mirror can and cannot interpose

Evidence:

```
$ grep -rn 'func renderConsolePage' internal/web/schema_render_test.go   # → GET /settings via a.routes()
$ sed -n '34,44p' internal/web/restyle_test.go                            # renderIndexBody → serveGet(t, a.routes(), "/settings")
$ sed -n '205,213p' internal/web/agentfm_polish_test.go                   # renderAgentFMBody → GET /settings
```

- **EXPOSED receivers** (full GET /settings render — contains ALL tabpanels, mirror
  included): `renderConsolePage` (schema_render_test.go:13), `renderIndexBody`
  (restyle_test.go:34), `renderAgentFMBody` (agentfm_polish_test.go:199),
  `serveGet(..., "/settings")`, `renderSettingsGET`, and direct
  `GET /settings` handler calls (schema_sections_test.go:95).
- **NOT-EXPOSED receivers**: `readEmbeddedAsset("i18n.js" / app.js / css)` asset text;
  `icons.templ` source read (todo_route_test.go:142); the `/todo` page
  (todo_section_test.go:64 — mirror renders only on /settings); `panelHTML(...)` slices;
  and any slice whose start point lies AFTER the codex panel in tab order.

**Tab order** (`wantTabOrder`, tab_layout_test.go:23, asserted as a rendered SEQUENCE by
`TestTabPanelRenderOrderMatchesTabs`):

```
identity, language, launch, llm, workflow, git-worktree, audit, codex, agentfm, report, mcp, crosssession, feedback, gate
                                                          (7)    (8)                (11)
```

The load-bearing asymmetry: **audit (7) renders BEFORE codex (8); mcp (11) renders AFTER.**

## 3. Mirror emitted-text inventory (what the mirror duplicates)

From `fieldsets_codex_templ.go:152-215` (`codexMirrorRowView`) and `codexmirror.go`:

Per mirrored field F, the mirror emits:

1. `<code class="key">F</code>` (field__head chip)
2. `<span class="field__label" data-i18n="f.F.title">F</span>` (label — F appears a
   second time as baseline text)
3. a mirror-edit link `href="/settings?tab=audit|mcp"` (`fieldsets_codex_templ.go:276`)

where F ∈ mirrored families, derived by predicate (`isCodexMirrorField`):
`workflow.audit.gates.codex`, `workflow.audit.codex.*`, `workflow.codex.*`,
`workflow.audit.model` (declared exception), and `mcp.tools.codex_*.enabled`
(codex-prefixed MCP catalogue tools only).

The mirror emits **NO** `name=` attributes (structurally impossible —
fieldsets_codex_templ.go:150 comment), no `<form>`, no `data-panel=`, no `<script>`,
and no text of any non-mirrored field.

Note: `panelHTML(t, html, "mcp")` is safe against the mirror's `href="/settings?tab=mcp"`
links — the helper anchors on the literal `data-panel="mcp"`, which the mirror never emits.

## 4. The discriminator (durable — survives this card)

**A page-wide `strings.Index` anchor diverges under the codex mirror iff ALL THREE hold:**

1. **EXPOSURE** — the receiver string is a full GET /settings render (the mirror renders
   only there, at tab position 8). Asset text, other pages, and already-scoped slices
   cannot interpose.
2. **DUPLICATION** — the anchored needle is in the mirror's emitted-text inventory:
   `<code class="key">F</code>`, `data-i18n="f.F.title"`, or plain `F`, for F in the
   mirrored families of §3. Needles with `name=`, `<form`, `data-panel=`, `<script`,
   or any non-mirrored field's text are never duplicated.
3. **ORDER** — the mirror renders EARLIER in the receiver than the intended target.
   Given tab order this holds only for needles owned by panels AFTER position 8 (in
   practice the MCP tab at 11). An audit-tab needle keeps its true first occurrence —
   the audit panel precedes the mirror.

**(c) = TRUE ⇔ 1 ∧ 2 ∧ 3.** Repair for (c)=TRUE rows: scope narrowing only (§ REQ-WAS-003).
A test whose receiver is not exposed is class NOT-EXPOSED — the mirror cannot interpose
there; it is not forced into either side of the divergence dichotomy.

## 5. Classification table (one row per raw-superset site — 60 rows)

Legend — Exposure: EXPOSED (full /settings body) / ASSET (css/js/dict source) / SRC
(icon template source) / PAGE (/todo page) / SLICED (already-scoped receiver).
Verdict: (c)=TRUE / false / REPAIRED-BY-T509 / NOT-EXPOSED.

| # | Site | Function | Receiver | Needle | (a) intent | (b) mirror effect | (c) |
|---|------|----------|----------|--------|------------|-------------------|-----|
| 1 | agentfm_policy_test.go:235 | TestAgentFMPolicy_SelectorsRenderAtTopOfPanel | EXPOSED | `data-i18n="sec.agentfm.title"` | ordering: section marker precedes controls | not duplicated (sec.* not mirrored) | false |
| 2 | agentfm_policy_test.go:236 | 〃 | EXPOSED | `name="performance_tier"` | agentfm control anchor | mirror emits no name= | false |
| 3 | agentfm_policy_test.go:237 | 〃 | EXPOSED | `name="agentfm.manager-spec.model"` | 〃 | 〃 | false |
| 4 | agentfm_polish_test.go:162 | TestAgentFMTierSortOrder | EXPOSED | `agentfm.manager-git.model` | row ordering across agent rows | agentfm fields not mirrored | false |
| 5 | agentfm_polish_test.go:163 | 〃 | EXPOSED | `agentfm.manager-docs.model` | 〃 | 〃 | false |
| 6 | agentfm_polish_test.go:164 | 〃 | EXPOSED | `agentfm.manager-spec.model` | 〃 | 〃 | false |
| 7 | agentfm_polish_test.go:165 | 〃 | EXPOSED | `agentfm.manager-design.model` | 〃 | 〃 | false |
| 8 | agentfm_polish_test.go:166 | 〃 | EXPOSED | `agentfm.plan-auditor.model` | 〃 | 〃 | false |
| 9 | codex_panel_test.go:269 | codexRowHTML | SLICED (codex panel region) | `>F<` row marker | locate mirror row inside codex region | receiver IS the mirror, already scoped | NOT-EXPOSED |
| 10 | codex_panel_test.go:274 | codexRowHTML | SLICED | `class="key"` | row boundary | 〃 | NOT-EXPOSED |
| 11 | codex_panel_test.go:285 | codexRowValue | SLICED | `class="in in--ro"` | value span | 〃 | NOT-EXPOSED |
| 12 | codex_panel_test.go:290 | codexRowValue | SLICED | `>` | 〃 | 〃 | NOT-EXPOSED |
| 13 | codex_panel_test.go:295 | codexRowValue | SLICED | `<` | 〃 | 〃 | NOT-EXPOSED |
| 14 | console_ux_fix_test.go:28 | cssRuleBlock | ASSET (css) | sel+"{" | css rule block | assets carry no render | NOT-EXPOSED |
| 15 | console_ux_fix_test.go:31 | cssRuleBlock | ASSET | sel+" {" | 〃 | 〃 | NOT-EXPOSED |
| 16 | console_ux_fix_test.go:38 | cssRuleBlock | ASSET | "}" | rule end | 〃 | NOT-EXPOSED |
| 17 | console_ux_fix_test.go:76 | TestOptionLabelsStayEnglish | ASSET (js) | "function applyI18n" | js fn boundary | 〃 | NOT-EXPOSED |
| 18 | console_ux_fix_test.go:82 | 〃 | ASSET | "\n  function " | next fn | 〃 | NOT-EXPOSED |
| 19 | console_ux_fix_test.go:279 | agentFMSectionCount | SLICED (from LAST sec.agentfm.title — agentfm panel, position 9 > 8) | `class="panel__meta">` | panel heading meta | slice starts after codex panel | NOT-EXPOSED |
| 20 | htmx_test.go:52 | TestHtmxLinkedBeforeAppJS | EXPOSED | `src="/static/htmx.min.js"` | script load order | mirror emits no script tags | false |
| 21 | htmx_test.go:53 | 〃 | EXPOSED | `src="/static/app.js"` | 〃 | 〃 | false |
| 22 | i18n_test.go:326 | TestLangpickRendered | EXPOSED | "<form " | langpick precedes settings form | mirror emits no form | false |
| 23 | i18n_test.go:367 | TestLangpickNotFormField | EXPOSED | "uiLangSelect" | langpick select | not mirrored | false |
| 24 | i18n_test.go:368 | 〃 | EXPOSED | "<form " | 〃 | 〃 | false |
| 25 | i18n_test.go:404 | TestLangpickJSWiring | ASSET (js) | "function wireLangpick" | js fn boundary | 〃 | NOT-EXPOSED |
| 26 | i18n_test.go:409 | 〃 | ASSET | "\n  function " | 〃 | 〃 | NOT-EXPOSED |
| 27 | mcp_console_test.go:115 | TestMCPConsoleWriteCapableTextDistinction | SLICED (panelHTML "mcp" — t509 fix) | `<code class="key">mcp.tools.<tool>.enabled</code>` | per-tool row window in MCP panel | WAS: mirror (position 8) precedes mcp (11) and duplicates codex_* chips → window landed on mirror row | REPAIRED-BY-T509 (historical (c)=TRUE) |
| 28 | mx_rawview_test.go:58 | TestSPEC014I18nKeysFourLocale | ASSET (dict) | locale block header | i18n dict structure | 〃 | NOT-EXPOSED |
| 29 | i18n_governance_test.go:55 | parseI18nCatalogue | ASSET (dict) | dict marker | 〃 | 〃 | NOT-EXPOSED |
| 30 | schemaform_test.go:72 | TestConsoleRendersReportTab | EXPOSED | `data-panel="launch"` | panel sequencing (genuinely page-wide) | unique marker, mirror emits none | false |
| 31 | schemaform_test.go:73 | 〃 | EXPOSED | `data-panel="llm"` | 〃 | 〃 | false |
| 32 | schemaform_test.go:94 | TestReportFormatRendersAsRadio | EXPOSED | `data-panel="report"` | panel slice | 〃 | false |
| 33 | schema_select_preserve_test.go:88 | browserSelectedValue | EXPOSED | `name="<field>"` | locate named select | mirror emits no name= | false |
| 34 | schema_select_preserve_test.go:93 | 〃 | SLICED (seg from #33) | "</select>" | select end | 〃 | false (scoped) |
| 35 | validate_test.go:244 | TestRenderModelEffortPolicyAreSelects | EXPOSED | `name="model"` / `name="effort_level"` | locate named select | 〃 | false |
| 36 | profile_bar_test.go:332 | TestProfileBarMarksCurrentAndSelected | EXPOSED | `data-pop-panel="profile"` | popover slice | not mirrored | false |
| 37 | profile_bar_test.go:342 | 〃 | EXPOSED | `action="/profile/rename"` | rename form slice | 〃 | false |
| 38 | profile_bar_test.go:343 | 〃 | SLICED | "</form>" | form end | 〃 | false (scoped) |
| 39 | schema_sections_test.go:148 | TestSchemaSectionsRenderSmoke | EXPOSED | `<details class="rawview">` | iterate ALL rawview blocks (genuinely page-wide) | mirror emits no rawview details | false |
| 40 | schema_sections_test.go:152 | 〃 | SLICED (from #39) | "</details>" | block end | 〃 | false |
| 41 | tab_layout_test.go:55 | TestTabPanelRenderOrderMatchesTabs | EXPOSED | `data-panel="` | full panel sequence (genuinely page-wide by design) | 〃 | false |
| 42 | tab_layout_test.go:60 | 〃 | SLICED | `"` | attr end | 〃 | false |
| 43 | tab_layout_test.go:81 | panelHTML (the protected helper) | EXPOSED | `data-panel="<panel>"` | panel slice — THE repair shape | 〃 | false |
| 44 | tab_layout_test.go:86 | panelHTML | SLICED | `data-panel="` | next panel | 〃 | false |
| 45 | todo_section_test.go:157 | TestTodoSectionCarriesExistingKanbanMarker | PAGE (/todo) | `data-live="kanban"` | todo section marker | mirror not on /todo | NOT-EXPOSED |
| 46 | todo_section_test.go:161 | 〃 | PAGE | "data-todo-row" | 〃 | 〃 | NOT-EXPOSED |
| 47 | todo_route_test.go:142 | TestTodoIconCaseExists | SRC (icons.templ) | "templ iconAt" | template source | 〃 | NOT-EXPOSED |
| 48 | webux_haiku_effort_test.go:31 | agentEffortSelect | EXPOSED | `name="agentfm.<a>.effort"` | agent select slice | mirror emits no name=; agentfm not mirrored | false |
| 49 | webux_haiku_effort_test.go:35 | 〃 | SLICED | "</select>" | select end | 〃 | false |
| 50 | webux_haiku_effort_test.go:48 | agentRowMarkup | EXPOSED | `data-i18n="agentfm.<a>..."` marker | agent row slice | not mirrored | false |
| 51 | webux_haiku_effort_test.go:57 | 〃 | SLICED | `data-agent-row="` | row boundary | 〃 | false |
| 52 | webux_haiku_effort_test.go:147 | TestAppJSHaikuEffortLockWired | ASSET (js) | "function wireProfileMatrix" | js fn | 〃 | NOT-EXPOSED |
| 53 | webux_haiku_effort_test.go:151 | 〃 | ASSET | "\n  function " | 〃 | 〃 | NOT-EXPOSED |
| 54 | webux_followup_test.go:62 | agentModelSelect | EXPOSED | `name="agentfm.<a>.model"` | agent select slice | mirror emits no name= | false |
| 55 | webux_followup_test.go:66 | 〃 | SLICED | "</select>" | 〃 | 〃 | false |
| 56 | webux_followup_test.go:80 | agentDescSpan | EXPOSED | `data-i18n="agentdesc.<a>"` | desc span slice | not mirrored | false |
| 57 | webux_followup_test.go:89 | 〃 | SLICED | "</span>" | 〃 | 〃 | false |
| 58 | webux_followup_test.go:366 | TestD3MissingKeyKeepsBaseline | ASSET (js) | "function applyI18n" | js fn | 〃 | NOT-EXPOSED |
| 59 | webux_followup_test.go:370 | 〃 | ASSET | "\n  function " | 〃 | 〃 | NOT-EXPOSED |
| 60 | webux_followup_test.go:387 | localeBlocks | ASSET (dict) | locale block header | 〃 | 〃 | NOT-EXPOSED |

## 6. Classification summary

| Verdict | Sites |
|---|---|
| (c)=TRUE pending repair | **0** |
| REPAIRED-BY-T509 (historical (c)=TRUE) | 1 |
| Exposed, (c)=false (needle not in mirror inventory, or genuinely page-wide intent) | 36 |
| NOT-EXPOSED (receiver cannot contain the mirror) | 23 |
| **Total** | **60** |

Interpretation: the 34/15 (lead) and 60/19 (raw) figures were the provisional upper
bound the card predicted; the discriminator narrows the repair set to zero ON THIS TREE.
The one historical (c)=TRUE site is exactly the site t509 already repaired — convergent
evidence that the discriminator reproduces the known failure. The zero-repair outcome is
a finding (REQ-WAS-006), not an incomplete sweep — and the mutant check (AC-WAS-005)
exists precisely so a zero over an empty set is not a confident verdict without teeth.

## 7. Residual observations (non-mirror first-occurrence ambiguity — NOT repair targets)

- Rows 22/24 (`"<form "`): the first form on the page is the settings form today, but the
  intent is "THE settings form"; a future chrome form rendered earlier would diverge for
  non-mirror reasons. Out of scope (§E of spec.md); recorded for a future card.
- Row 47 (`icons.templ` source read at test time): reads the template SOURCE, not the
  binary — deliberate (source-contract test), noted only because it looks like a render
  receiver.

## 8. Green baseline (mirror present, this tree, this run)

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/web/
ok  	github.com/modu-ai/moai-adk/internal/web	3.902s
```

All 60 sites are green under the mirror today — which is exactly why a mirror-absent
tree cannot arbitrate this sweep (REQ-WAS-004) and why the mutant check is the only
demonstration that the discriminator has teeth.
