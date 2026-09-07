# SPEC Review Report: SPEC-WEB-CODEX-PANEL-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.69** (Tier M PASS threshold 0.80 — `spec-workflow.md:141`)

Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t509`, branch `WT-codex-model-config`, HEAD `a732bc6d5`.
Reasoning context ignored per M1 Context Isolation. The dispatch's framing was used only to select
attack surfaces; every judgement below is against the artifacts and the source in this tree.

Note: the SPEC cites HEAD `8a6e21d98`; the audited tree is `a732bc6d5`. Every cited line number in
spec.md that I checked still resolves in this tree (drift ≤ 3 lines on two citations), so the
citation set is treated as sound.

---

## What I verified in this tree vs what I took on the documents' word

**Verified by execution or by reading source in this tree** (all findings below rest on these):

| Claim | How verified |
|---|---|
| 11 codex-named editable fields + `workflow.audit.model` in `settings.AllFields()` | temporary `go test` enumerating `AllFields()` (written, run, deleted — tree left clean) |
| `mcpcat.MoaiMCPTools()` exists and is enumerable | `internal/mcp/catalog.go:77`, six `codex_*` entries at `:50-55` |
| Panel region is delimitable | `internal/web/root.templ:64-88` (`data-panel` / `id="panel-…"`), and the existing `panelHTML` slicer at `internal/web/tab_layout_test.go:76-88` |
| 54 of 125 distinct form `name` values already repeat on the pre-change page | temporary `go test` over `renderConsolePage` (written, run, deleted) |
| Every bool renders as a radio PAIR sharing one `name` | `internal/web/fieldsets.templ:365-375` (`boolSegment`) |
| Probe sentinels already render on the MCP panel | `internal/web/fieldsets.templ:739-762` (`codexAuthBlock`), called from `:693` |
| A missing git ref makes AC-WCP-011's grep return `rc=1` | ran `git diff --name-only origin/nosuchref...HEAD 2>/dev/null \| grep … ; echo rc=$?` → `rc=1` |
| Four locale blocks; a tab key appears exactly 4× | `grep -c 'tab.audit.title' internal/web/assets/i18n.js` → `4` |
| Referenced SPECs exist, all `status: completed` | `grep '^status:'` per directory |
| No `syscall`, no `[NEEDS CLARIFICATION]` | `grep` over the SPEC artifacts |
| New-tab couplings | `wantTabOrder` (`tab_layout_test.go:21`), `settingsTabFieldNames` / `settingsTabFieldCount` (`internal/web/settings_shell.go:120-157`), `schemaPanelMeta` empty-meta fallback (`schemaform.go:283-290`) |

**Taken on the documents' word (NOT verified here)**: the contents of `.moai/reports/t509/verdict.md`
§9/§11/§13/§14/§15 as an investigation record (I read the SPEC's use of it, not the verdict itself,
except where the SPEC's claim was independently re-measurable — and every such claim I re-measured is
listed above); the operator ruling of 2026-09-07 selecting the read-only axis; the scope boundary of
card t517.

**Explicitly NOT checked**: whether the tab-switching mechanism is CSS or JS (the SPEC declares this
unmeasured and out of scope — I did not measure it either); the `hx-boost` submit path; whether
`moai web` rewrites config files on GET (B-3); any run-phase behaviour; the full `internal/web` test
suite (I ran only two throwaway tests, both deleted).

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-WCP-001…011, sequential, no gaps, no duplicates, uniform 3-digit padding (`spec.md:143-184`).
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md §C`), never against ACs. All 11 match a GEARS pattern: Ubiquitous (001, 004, 005, 007, 008), Unwanted canonical `shall not` (002, `spec.md:146`), Event-driven (003, 009, 010), Where capability-gate (006, `spec.md:161`), While state-driven (011, `spec.md:182`). Given-When-Then in `acceptance.md` is the correct verification-layer format and is graded in Group 4 only.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (`spec.md:2-13`); `created`/`updated` ISO, `version` quoted, no rejected snake_case alias. Extra `era`/`tier`/`related_specs` are permitted optionals.
- **[N/A] MP-4 language neutrality** — single-language (Go/templ) SPEC scoped to `internal/web`. Auto-pass.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — referenced IDs `SPEC-MCP-CONSOLE-001`, `SPEC-V3R6-AUDIT-MODEL-PIN-001`, `SPEC-WEB-CONSOLE-014` all exist under `.moai/specs/` with `status: completed`; none retired/superseded/archived. No BLOCKING.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. Auto-pass per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' spec.md plan.md acceptance.md` → no match (rc=1). No `research.md` (correct for Tier M).

No must-pass failure. **The FAIL is driven by the rubric scores** — specifically Testability, where
three of the twelve deciding commands do not decide what they claim.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75–1.0 | Requirements are unambiguous and the §B premise re-measurement is exemplary. Two ambiguities: AC-WCP-008 does not say the sentinel search is panel-scoped (`acceptance.md:84-94`), and AC-WCP-006's "at least the twelve fields" (`acceptance.md:70`) inherits an unresolved decision |
| Completeness | 0.75 | 0.75 band | All sections present, frontmatter complete, four `### Out of Scope — <topic>` H3s with specific bullets (`spec.md:205-233`). Sparse in one place: the new-tab coupling to `settingsTabFieldNames`/`settingsTabFieldCount` is named nowhere (D6) |
| Testability | 0.50 | 0.50 band | Several ACs cannot be evaluated as written: AC-WCP-005 is RED before the change (D1), AC-WCP-011's non-vacuity claim is false (D2), AC-WCP-008 passes without the panel existing (D3), AC-WCP-012's command cannot decide its claim (D7) |
| Traceability | 0.75 | 0.75 band | Every REQ has ≥1 AC and every AC maps to an existing REQ (`acceptance.md:166-180`). One indirect mapping is load-bearing and does not hold: REQ-WCP-009 → AC-WCP-006 (D5) |

Aggregate (harmonic mean, per the skeptical-evaluation stance): **0.69** < 0.80.

---

## Answers to the eight attack points

**1. Is the guard's mutant split load-bearing? — Partly; the stated rationale is wrong.**
Under AC-WCP-003 **as written** ("zero occurrences of `name=`" in the region, `acceptance.md:30`),
MU-2 emits `name="…__present"`, which contains `name=`. So MU-2 fails AC-WCP-003 too, and AC-WCP-004
is strictly *implied* by AC-WCP-003 — no conforming implementation can pass MU-2 while failing MU-1.
A guard that passes MU-2 while failing MU-1 is reachable only if the guard is implemented *narrower
than specified* (e.g. matching `<input type="text"` / `<select`), and then only if MU-1 mutates a
string field — `workflow.audit.codex.model` is `type=text` (measured), so such a pair does exist.
Conclusion: keep both mutants (they guard against a mis-implemented guard, which is legitimate), but
`acceptance.md:158,163-164` asserts a discrimination that does not hold against the AC as written.
Recorded as D8 (minor).

**2. Is the panel region delimitable? — Yes, confirmed.**
`root.templ:64-72` emits `data-panel={ tab.ID }` and `id={ "panel-" + tab.ID }` per panel, and the
test helper `panelHTML` (`tab_layout_test.go:76-88`) already slices marker-to-next-marker. One
dependency the AC does not state: `panelHTML`'s fallback slices to end-of-document for the **last**
panel, so the assertion silently depends on codex not being last. Plan M2's baseline (after `audit`)
satisfies it, and the failure mode is loud rather than vacuous, so this is minor (D9).

**3. Is REQ-WCP-005's predicate derivable? — Yes for 11 of the 12 fields, no for the 12th.**
`mcpcat.MoaiMCPTools()` (`internal/mcp/catalog.go:77`) is enumerable in Go and carries the six
`codex_*` entries (`:50-55`), and `mcpFields()` (`internal/settings/schema_sections.go:636-644`)
derives `mcp.tools.<name>.enabled` from it. A predicate over `AllFields()` + that catalogue reaches
exactly 11 fields. `workflow.audit.model` — the twelfth — contains no `codex` token and belongs to no
codex catalogue, so it can enter the mirrored set **only by being hand-named**, which is precisely
what REQ-WCP-005 forbids. See D4.

**4. Does any plan step touch a save path? — No.**
Read literally, plan.md's steps modify: `consoleTabs()` (`schemaform.go:34`), the `root.templ` switch,
a new templ component, `icons`/i18n, and `tab_layout_test.go`. None of them touch `parseSchemaForm`
(`schemaform.go:310`), `handleSave`, `ApplySchemaEdits`, or the config write path, and §D:37 states the
prohibition. The plan is clean here. The *guards* meant to prove it are not — see D7.

**5. Counts. — 11 codex-named + 1 shared; the "≥12" floor is wrong-by-construction.**
Measured `AllFields()`: `workflow.audit.gates.codex` (radio), `workflow.audit.codex.model` (text),
`workflow.audit.codex.effort` (select), `workflow.codex.review_gate.enabled` (bool),
`workflow.codex.task.allow_write` (bool), and six `mcp.tools.codex_*.enabled` (bool) = **11**.
`workflow.audit.model` exists (radio) but is not codex-named. The SPEC's claim that
`codex.{auth_provider,binary,version}` are probe state, not registry fields, is **confirmed** —
`CodexStateView` (`internal/web/codex_state.go:19-26`) carries them and no `FieldDef` does. The floor
cannot pass vacuously *given a correct predicate*, but it is coupled to an open decision (D4).

**6. i18n coupling. — The named file is right; one coupling is missing.**
`internal/web/assets/i18n.js` holds all four blocks (`en` at :26ff, `ko` ~:1018, `ja` ~:1723,
`zh` ~:2428); `grep -c 'tab.audit.title'` returns exactly `4`, so AC-WCP-010's shape is correct, and
`TestI18nKeyCoverageForward/Reverse` (`i18n_governance_test.go:408,423`) enforce it. The SPEC/plan name
`wantTabOrder` (plan M5) and the icon case (plan M2). **Missing**: `settingsTabFieldNames` /
`settingsTabFieldCount` (`settings_shell.go:120-157`) — see D6. The icon step is itself unactionable as
written — see D10. Neither breaks the build (an unmatched icon name renders nothing, precedent
`shield-check` at `schemaform.go:270` with no case in `icons.templ`; an unknown panel id returns an
empty meta, `schemaform.go:283-290`), so no build/test breakage is missed.

**7. Tier. — Tier M is correct.** 11 REQ / 12 AC are inside the Tier M budget of 16/16
(`spec-workflow.md:146-150`); the plan touches ~7 files, inside the 5–15 band (`:141`). The 3-artifact
set (spec + plan + acceptance) matches. No defect.

**8. An AC that passes without the feature working. — Two of them.** D2 (AC-WCP-011 passes on a
broken ref) and D3 (AC-WCP-008 passes on markup that already exists on another tab). Plus D1, which
is the inverse failure: an AC that fails before the feature is even attempted.

---

## Defects Found

**D1 — AC-WCP-005 is RED on the pre-change tree, and its stated premise is false**
`acceptance.md:50-59` — Severity: **critical** — Class: **blocking**
The AC asserts "no `name` value appears more than once" inside `id="settings-form"`, and line 58-59
asserts "It must be green on the pre-change tree as well". Measured on this tree: **125 distinct names,
54 of them duplicated** (e.g. `workflow.audit.gates.codex` ×3, `mcp.tools.codex_audit.enabled` ×2,
`performance_tier` ×4). The cause is structural and permanent: `boolSegment`
(`internal/web/fieldsets.templ:365-375`) renders every bool as an on/off **radio pair sharing one
`name`**, and radio groups repeat a name by definition (127 radio inputs in the form). The invariant as
written is unimplementable, and its mechanism claim is also wrong — `PostFormValue`'s first-value
semantics is not defeated by name repetition per se, but by the same field being submitted from more
than one panel.
*Required fix*: restate AC-WCP-005 as the invariant that actually bites — **no field name renders in
more than one panel** — which the codebase already implements as a pattern (`assertPanelFields`,
`tab_layout_test.go:90-103`, comparing `strings.Count(html, marker)` against `strings.Count(body,
marker)`). If a form-wide count check is still wanted, it must group same-name controls by field group
and exempt a single radio group, and the *measured* pre-change baseline (54 duplicates) must be
recorded rather than an assertion of greenness.

**D2 — AC-WCP-011's non-vacuity claim is false; the AC passes on a broken base ref**
`acceptance.md:125-128` — Severity: **critical** — Class: **blocking**
The AC runs `git diff --name-only origin/develop...HEAD | grep -E '…' ; echo "rc=$?"` and expects
`rc=1`, stating that "`grep` returning 1 is the positive signal that the sweep ran and matched
nothing". Measured: `git diff --name-only origin/nosuchref...HEAD 2>/dev/null | grep -E '\.moai/config/'
; echo rc=$?` prints **`rc=1`**. An unfetched, renamed, or mistyped base ref produces the identical
pass signal — exactly the failure the AC claims to have closed. (`origin/develop` does resolve here,
at `ace1c5440`, so the AC would pass today for the right reason by luck, not by construction.)
*Required fix*: gate the sweep on the ref resolving and on the diff being non-empty, e.g.
`git rev-parse --verify -q origin/develop` first, then assert `git diff --name-only origin/develop...HEAD
| wc -l` > 0 before interpreting the filtered grep's `rc=1`.

**D3 — AC-WCP-008 is satisfiable without the codex panel rendering anything**
`acceptance.md:84-94` — Severity: **major** — Class: **blocking**
The AC injects binary/version/auth-provider sentinels and asserts "all three sentinels appear
verbatim", explicitly modelling the test on `TestCodexCard_SentinelPropagationCrossSurface` and
`renderAppBody` (`internal/web/mcp_codex_surface_test.go:24,39`) — both **whole-body** renders. Those
three sentinels **already render today** on the MCP panel via `codexAuthBlock`
(`internal/web/fieldsets.templ:739-762`, called from `:693`). A test written as specified is green
before a single line of the codex panel exists.
*Required fix*: scope the assertion to `panelHTML(t, html, "codex")` and assert the sentinels appear
*inside that region*; keep the whole-body count only to confirm the MCP surface is unchanged.

**D4 — AC-WCP-006's "at least twelve" floor is coupled to an unresolved plan decision, and the twelfth field is not derivable**
`acceptance.md:70`, `spec.md:70`, `plan.md:58-61` — Severity: **major** — Class: **blocking**
The floor of 12 holds only if `workflow.audit.model` is mirrored — which plan.md M1 explicitly leaves
open ("Decide whether `workflow.audit.model` … is mirrored. Baseline: yes … the one row whose inclusion
is a judgement rather than a derivation"). Measured derivable set = **11**. So an implementer taking the
other branch of a decision the plan invites them to take fails a blocking AC; and taking the baseline
branch requires hand-naming a field, in tension with REQ-WCP-005 (`spec.md:157-159`) and with plan.md
§D:38-39 ("Mirror rows are derived by predicate, never hand-listed").
*Required fix*: close the decision in spec.md rather than in plan.md, and split the AC — assert the
**derived** set is exactly the 11 predicate matches, and, if `workflow.audit.model` is included, declare
it in the SPEC as a single named exception with its own assertion, so the derivation invariant and the
judgement row are tested separately.

**D5 — REQ-WCP-009 is not covered by the AC it maps to**
`spec.md:174-176`, `acceptance.md:61-70,178` — Severity: **major** — Class: **blocking**
REQ-WCP-009 requires that a codex field present in the registry but absent from the panel fails a test.
AC-WCP-006's coverage half is **self-referential**: the predicate defines both the selected set and the
expected rows, so a predicate narrowed to miss a field cannot fail it. The only non-self-referential
content is the hardcoded floor, and MU-3 (`acceptance.md:159`) bites solely through that floor. A codex
field added to the registry that the predicate does not match therefore satisfies AC-WCP-006 while
violating REQ-WCP-009.
*Required fix*: give the coverage test an **independent** enumeration to compare the predicate against —
e.g. every `settings.AllFields()` entry whose name contains `codex`, unioned with every
`mcpcat.MoaiMCPTools()` entry whose `Name` has the `codex_` prefix — and assert set equality with the
predicate's output, not with itself.

**D6 — the new-tab coupling to the rail field-count is named nowhere**
`spec.md §C/§F`, `plan.md §F M2` (omission); source `internal/web/settings_shell.go:120-157` —
Severity: **major** — Class: **blocking**
A new tab id falls to the `default` branch of `settingsTabFieldNames` (`:154-155`), which calls
`schemaPanelMeta("codex")`; that returns an **empty meta** for an unregistered panel
(`schemaform.go:283-290`), so `settingsTabFieldCount` returns **0** and the rail subnav renders "0"
beside a panel showing ~12 rows. The in-code contract at `settings_shell.go:113-118` states this
explicitly: the rail number "must equal the number the panel puts in its own header — if they differ
there is no way to know which is right". Plan M2 forbids adding a `schemaSectionMeta` entry (correctly,
since the panel owns no fields), which means this needs a deliberate `case "codex"` in
`settingsTabFieldNames` or an explicit decision to render no count. No AC covers it, and nothing fails
if it is missed.
*Required fix*: name `settings_shell.go` in spec.md §F and in a plan milestone, decide what the rail
shows for a field-owning-nothing mirror panel, and add an AC pinning rail count == rendered mirror-row
count (or == 0 by explicit decision).

**D7 — AC-WCP-012's deciding command cannot decide its claim**
`acceptance.md:130-137` — Severity: **major** — Class: **blocking**
The claim is "`parseSchemaForm`, `ApplySchemaEdits`, and `handleSave` are unmodified". The command
greps **added** lines for the literal tokens `PostFormValue|PostForm\[|func parseSchemaForm|func
handleSave|func ApplySchemaEdits`. Any edit inside those function bodies that does not add one of those
literals passes; any deletion passes; and an empty or mis-pathed diff passes vacuously (same shape as
D2). The related command in AC-WCP-009 (`acceptance.md:104`) is worse: `git diff --stat
internal/web/schemaform.go` is asked to show "no change to `partitionWorkflowFields` or
`isCodexToggleFieldName`", but `--stat` reports file-level line counts only — and plan M2 *does* modify
that file (`consoleTabs()`), so the command will show a non-zero stat that decides nothing either way.
*Required fix*: decide function-body invariance by extracting the function bodies at both revisions and
comparing them (`awk`-bounded extraction or `git show <base>:file` + a function-range diff), and replace
the `--stat` check with a targeted `git diff -U0` hunk-header assertion naming the two functions.

**D8 — the MU-1/MU-2 rationale asserts a discrimination that does not hold**
`acceptance.md:158,163-164` — Severity: **minor** — Class: **optional**
Under AC-WCP-003 as written, MU-2's emitted `name="…__present"` also trips the `name=` sweep, and
AC-WCP-004 is strictly implied by AC-WCP-003. The split is defensible as insurance against a
narrower-than-specified guard implementation, but the text claims it is the only thing separating two
failure modes, which is not true of the AC it is written against.
*Required fix*: reword to "MU-2 guards against a guard implemented narrower than AC-WCP-003 specifies
(e.g. one matching only `<input type="text"` / `<select`); against the AC as written it is redundant
with MU-1, and it is retained deliberately."

**D9 — AC-WCP-003's `name=` substring sweep and its unstated placement dependency**
`acceptance.md:27-36` — Severity: **minor** — Class: **optional**
Two small things: the bare substring `name=` also matches any attribute ending in `-name=`
(`data-…-name=`), so a future markup choice could false-fail; and the region slice depends on codex not
being the last panel (`panelHTML`'s fallback runs to end-of-document, `tab_layout_test.go:83-86`).
*Required fix*: assert on `name="` and pair it with an explicit control-tag sweep; state the
"codex is not the last panel" dependency alongside plan M2's placement decision, or slice on the
`id="panel-codex"` … next `id="panel-` boundary instead.

**D10 — plan M2's icon step is unactionable as written**
`plan.md:74` — Severity: **minor** — Class: **optional**
"Pick an icon from the existing `icons.templ` cases" has no landing place: `consoleTab`
(`schemaform.go:26-30`) carries no `Icon` field, and `Icon` lives on `schemaSectionMeta`, which the very
next bullet (`plan.md:75-76`) forbids adding for this panel. The icon can only be hardcoded inside the
dedicated component.
*Required fix*: say so — "hardcode an existing `iconAt` case inside `fieldsetCodex`; a name with no
`case` renders nothing (precedent: `Icon: "shield-check"` at `schemaform.go:270` has no case in
`icons.templ`)".

---

## What is good, and should survive the revision

Stated so the fix does not throw it away: §B (premise re-measurement, `spec.md:86-139`) is exactly the
right instinct and two of its three corrections are verified sound in this tree; §A.1's split between
registry fields and probe state is correct and is the distinction the earlier rounds got wrong; §D's
record of *why* read-only was chosen (`spec.md:186-201`) is what stops a later reader "fixing" it; and
the Out-of-Scope sections correctly hand the save-path surface to t517. The defects above are all in the
verification layer and in two omitted couplings — none of them is a reason to re-open the mechanism.

## Recommendation

FAIL. Route the six blocking defects (D1, D2, D3, D4, D5, D6) plus D7 back to manager-spec; the three
optional findings (D8, D9, D10) are the orchestrator's call. Concretely:

1. Replace AC-WCP-005 with the per-panel uniqueness invariant and record the measured 54-duplicate
   baseline as a fact rather than asserting greenness (D1).
2. Add a ref-resolution and non-empty-diff precondition to AC-WCP-011, and drop the false `rc=1`
   rationale (D2).
3. Scope AC-WCP-008's sentinel assertion to the codex panel region (D3).
4. Close the `workflow.audit.model` decision in spec.md, split AC-WCP-006 into a derived-set equality
   assertion plus a named-exception assertion, and drop the "≥12" literal (D4, D5).
5. Name `internal/web/settings_shell.go` as a coupled surface, decide the rail count, and add an AC for
   it (D6).
6. Replace AC-WCP-012's token grep and AC-WCP-009's `--stat` check with function-body comparisons (D7).

Iteration 2 will be scoped to this enumerated delta plus a regression check over it, per the Tier M
ceiling of 2.

---

## Residual risk

- I did not run the existing `internal/web` suite; a pre-existing RED there would change the meaning of
  plan §C's baseline step, and I cannot rule it out.
- The two throwaway tests I wrote measured the page rendered by `renderConsolePage` / `newTestApp`; a
  differently-seeded view could render a different name population, though not a different
  `boolSegment` structure (the duplicate cause is structural, not data-dependent).
- The tree is left unmodified apart from this report: `git status --short` was checked after deleting
  both temporary test files.
