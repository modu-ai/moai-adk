# SPEC-JEV-AUTO-EXCEPTION-001 — Implementation Plan

Companion to `spec.md`. Priority labels and phase ordering only; no time estimates.
Verbatim evidence above 50 lines lives in `research.md`.

## §Findings (read-only investigation, pinned tree `c50da9c2f`; revision re-run at HEAD `1eef55dd9`, branch `WT-jev-auto-exception`)

**(a) Environment.** `git rev-parse --show-toplevel` →
`/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1403`; `git branch --show-current` →
`WT-jev-auto-exception`; `git rev-parse --short HEAD` → `1eef55dd9` at the revision
(`c50da9c2f` at plan start); `git diff --name-only c50da9c2f 1eef55dd9` lists only the
five files of this SPEC's directory, none under a ledger pathspec, so the pinned tree's
measurements hold for the revision tree. Tools built from the pinned tree for the
first run: `go build -o <scratchpad>/moai ./cmd/moai` (the tree-local binary reports
`v3.1.3 none built unknown` because no ldflags were passed; it is the judging build
for every `moai spec lint` result cited, per `verification-claim-integrity.md` §2.2;
the revision lint result names the tree HEAD of the build it used).

**(b) Mirror state of the in-scope pairs.**

| Pair | State today | Consequence |
|---|---|---|
| `moai-mcp-tools-catalogue.md` live ↔ template | byte-identical (`diff` silent) | edit both identically; `TestMCPToolCatalogueDocsStayMirrorIdentical` enforces it |
| `CLAUDE.md` live ↔ template | byte-identical (15,573 bytes each) | edit both identically |
| `agent-authoring.md` live ↔ template | differ at `:130` and `:184` only (pre-existing sanitization), line `:147` identical | edit `:147` in both; not in `sanitizedPairPaths` |
| `skills/moai/SKILL.md` live ↔ template | differ (80 diff lines, pre-existing), line `:180` identical | edit `:180` in both; the skill directory is hashed in `internal/template/catalog.yaml` |
| `workflow.yaml` live ↔ template | differ (405 diff lines — local config vs shipped defaults) | the jev comment run must carry the same wording in both; `jev.enabled` differs by design (live `true`, template `false`) and is not touched |
| `moai-ref-jev-question-design/SKILL.md` live ↔ template (X5) | byte-identical (`diff` silent, exit 0) | edit both identically; `TestJevQuestionDesignSkillCopiesStayIdentical` enforces it; the directory is hashed in `internal/template/catalog.yaml` (entry at `:61-65`) |

**(c) Guards that read the passages this SPEC edits** — the pinned-text table is
`spec.md` §B.8 (P1-P12; P10 is env-gated and excluded). Measured PASS at the pinned
tree (research.md §R3): the
`TestJevAmendmentLinkage` run (6 subtests), `TestJevDoctrineAmendment` (6 subtests),
`TestMCPToolCatalogueDocsStayMirrorIdentical`, `TestNoConsumerCallPathShips`,
`TestPackageImports_AreStandardLibraryOnly`, and the five `--auto` doc/mirror
guards (`TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`,
`TestAutoRankMarkerDisclosure`, `TestAutoHelpAndRefusalDoNotAssertPickOrder`,
`TestAutoRankAgentDoctrine`). Eight ranking behavior tests also PASS. The two X5
guards (`TestJevQuestionDesignSkillCarriesNoCallPath`,
`TestJevQuestionDesignSkillCopiesStayIdentical`) and `TestCatalogHashParity` were run
at the revision tree `1eef55dd9` (no Go or template file differs from the pinned tree)
and PASS — two `--- PASS` lines, and `verified 49 catalog entries … 0 drift`
(research.md §R3, V13/V14).

**(d) Regeneration facts.** `make build` runs `agents-emit-check`,
`commands-emit-check`, `tool-policy-drift-check`, `templ-generate`, then
`gen-catalog-hashes --all`. This SPEC edits no agent definition and no command, so no
`make agents-emit` or `make commands-emit` is owed. Editing
`internal/template/templates/.claude/skills/moai/SKILL.md` (X3) changes the `moai`
skill-directory hash in `internal/template/catalog.yaml` (49 `hash:` entries; the
`moai` entry sits at its top), and editing
`internal/template/templates/.claude/skills/moai-ref-jev-question-design/SKILL.md`
(X5) changes that directory's hash (entry `:61-65`), so each surviving one brings
`catalog.yaml` into the same commit; `TestCatalogHashParity`
(`internal/spec/catalog_hash_test.go:112`) fails on a hash left stale. The other
template files in scope (rules, `workflow.yaml`, `CLAUDE.md`) are not
catalog-hashed.

**(e) What the first amendment looked like**, as the style to follow
(commit `185569ef3`, 10 files): the catalogue rows and the `workflow.yaml` comment
gained one clause each; `SPEC-JEV-CORE-001` gained markers, an exception paragraph
per requirement, edited authority bullets and a HISTORY row; `CLAUDE.local.md` gained
one line; `kickoff.go` flipped a constant; `contract_mode_blocks_test.go` gained the
content guard (231 lines); one commit. Its assembly order (design.md §11.2):
`manager-spec` writes the SPEC body, `manager-develop` writes code, rules, config and
the guide line, the lane orchestrator stages by explicit path into one commit.

## §A Context

Card t1403 closes the split state `spec.md` §A.1 describes. The deliverable is
documentation, two completed SPEC bodies, code comments, and one new guard test. No
runtime behavior changes (REQ-JAE-008). The work is small in lines and wide in
files, and its risk is not in the edits but in the pinned text around them
(`spec.md` §B.8) and in the commit-graph discipline the guard imposes.

## §B Decisions and open items

Ordered by decision-reversibility: the wording and surface decisions first (most
likely to be changed by the operator), the guard design next, the mechanical
assembly last.

**Design decisions (D):**

- **D-1 — The canonical clause.** One clause, adapted per surface, so every passage
  says the same thing. Reference wording (English; the run phase may refine it as
  long as the requirements and the pins hold):

  > A second exception, separate from the Kickoff cross-check, is the
  > `moai todo --auto` cycle's own candidate ranking — the
  > auto-scoped ranking exception. Behind the default-off `workflow.jev.enabled`
  > gate, the in-process capability may supply the key that orders the queued
  > candidates the cycle is about to accept, which sets its
  > selection order only. The candidate set is fixed by
  > mechanical filters before any answer is read, and an answer that cannot be
  > used as a whole falls back to recorded priority over the same set; the answer
  > never adds, removes or edits a card, is never the basis of a completion verdict,
  > a merge approval or an operator gate, and is not claimed to be accurate.
  > (Keep each of the literals `auto-scoped ranking exception`,
  > `selection order only`, `mechanical filters` and `workflow.jev.enabled` on one
  > line when copying the clause into a file.)

  It deliberately does **not** contain the exact phrase `contract-mode Kickoff`
  (P2 counts that phrase and requires exactly one occurrence per row and per YAML
  comment run).

- **D-2 — Per-surface adaptation.**

  | Surface | Form | Notes |
  |---|---|---|
  | S1 `jev.go:24-28` (long-form) | extend the third bullet: name the two in-process consumers, the exception literals, the gate literal `workflow.jev.enabled`, the `mechanical filters` clause, "they live outside this package and never reach the MCP tool"; correct `:27` to `display_only_test.go` (the file that exists) | Go comment; none of the three P6 literals; imports untouched |
  | S3 `mcp_jev.go:8-10` (long-form) | add: "the `todo --auto` exception (auto-scoped ranking exception, selection order only), behind the default-off workflow.jev.enabled gate with the candidate set fixed by mechanical filters first, is an in-process consumer and does not reach this tool" | description string `:48` untouched |
  | S2 `workflow.yaml` | a second comment block after the existing "One exception:" block, still inside the `# jev:` run | live `:230-232`, template `:232-234` anchor the insertion; keep "contract-mode Kickoff" once |
  | S4 catalogue rows `:139`, `:233` | append one clause to each row: `and except as the ordering key of the \`todo --auto\` cycle's own candidate ranking (the auto-scoped ranking exception — selection order only, in-process, never through this tool)` | short-form; both copies byte-identical |
  | S5 `SPEC-JEV-CORE-001` (long-form) | see D-6 | the second-exception paragraphs carry `workflow.jev.enabled` and `mechanical filters` |
  | S6 `SPEC-MANAGER-TODO-001` (short-form) | see D-6 | scope sentences, not the long-form clause (REQ-JAE-007 fixes their content) |
  | X1 guide (long-form) | one **new** paragraph after `:32`, Korean; the pinned `:32` paragraph untouched | scopes the old "한 곳뿐" to the Kickoff cross-check; carries `workflow.jev.enabled` and `기계적 필터` |
  | X2 `agent-authoring.md:147` | "consults Jev as a display-only signal (the `--auto` cycle's ranking is the one auto-scoped ranking exception — selection order only)" | |
  | X3 `SKILL.md:180` | parenthetical after "never reorder by inferred priority" naming the `--auto` cycle's own ranking as the one auto-scoped ranking exception, selection order only | |
  | X4 `CLAUDE.md:63` | "Jev display-only consultation; the `--auto` ranking is the one auto-scoped ranking exception, selection order only" | growth bound in §D |
  | X5 `moai-ref-jev-question-design/SKILL.md:25-29` (short-form), live and mirror | append one sentence after "…or any other decision that is hard to undo.": "The one exception is the `todo --auto` cycle's own candidate ranking — the auto-scoped ranking exception, selection order only — an in-process consumer behind the default-off gate." | byte-identical in both copies (P11); none of `internal/jev`, `mcp__moai__jev`, `jev_ask`, `moai jev` (the sentence names the `todo --auto` cycle, not the wrapper tool); `catalog.yaml` hash regenerated (P12) |

- **D-3 — The marker registry is derived from the confirmed surface list.** One row
  per file: `{path, token}` with `token = auto-scoped ranking exception` for every
  file except the arming row, whose path is the test file itself and whose token is
  built at run time as `"jevAutoExceptionAmended" + " = true"` — **never written as one
  literal anywhere in the test file** (D-5). Cutting an extension removes its rows;
  nothing else changes. The registry lists live and mirror files separately (so a
  missing mirror is a "partial amendment"): `jev.go`, `mcp_jev.go`, `workflow.yaml` ×2,
  catalogue ×2, `SPEC-JEV-CORE-001`, `SPEC-MANAGER-TODO-001`, the guide,
  `agent-authoring.md` ×2, `SKILL.md` ×2, `CLAUDE.md` ×2, the reference skill ×2, plus
  the arming row — 18 rows with every extension.

- **D-4 — The guard lives in `internal/template/jev_auto_exception_test.go`**
  (package `template_test`). Rejected: `internal/contract/kickoff` (about the
  Kickoff; adding unrelated tests there blurs the file's meaning), `internal/jev`
  (standard-library import set), `internal/cli` (compile cost of the largest test
  binary and an unrelated package for a repository-level guard).

- **D-5 — Arming constant, and how its token stays out of the guard's own source.**
  `const jevAutoExceptionAmended = false` in the test file at M1; `true` in the linked
  commit. The tree subtest reads it: when `false`, every marker must be absent
  (all-or-none still applies); when `true`, every marker must be present and
  first-appear in one commit. The constant is itself a marker, so flipping it alone or
  landing the other markers without it both fail. **The registry token is assembled by
  parts** (`"jevAutoExceptionAmended" + " = true"`), because the predecessor's presence
  check is `strings.Contains` over the file bytes and its first-commit check is
  `git log -S<token> -- <path>` (`internal/contract/kickoff/activation_test.go:179`,
  `:44-50`): a registry row spelling the literal in the file under test would make
  the marker present — and first-appearing — in the guard's own commit `G`. Rules the
  run phase keeps: the constant is a standalone `const` line (inside a `const (…)`
  block `gofmt` would align the `=` and break the literal); the text
  `jevAutoExceptionAmended = true` appears nowhere else in the file — not in a comment,
  not in a fixture, not in a message string (fixtures build it from the same two
  pieces); and the constant name alone, without ` = true`, may appear freely. The same
  holds for the `= false` text: it appears once at `G`, as the const declaration
  (`git grep -c -F "jevAutoExceptionAmended = false" 6d012fd4d -- <guard>` printed
  `…:1`), and a comment or fixture spelling it would make AC-JAE-013's count 2 — so
  the rule covers both spellings, and after `K` neither the `= false` text nor a second
  `= true` text is in the file (`git grep -n "jevAutoExceptionAmended"` lists the
  name only in the name constant, comments and uses; read at `e70578c24`).
  Alternatives rejected: the constant in a non-test Go file (adds a non-comment Go
  line, contradicting REQ-JAE-008) and detecting the flip by a regular expression
  instead of `Contains` (departs from the predecessor's mechanics for no gain).
  Measured on a throwaway draft re-implementing the predecessor's presence and
  first-commit checks (research.md §R6): by-parts — no findings at `G`, no findings
  at `K`, `git grep -c -F` of the literal exits 1 at `G` and prints `:1` at `K`; the
  self-match variant — a `partial amendment` finding at `G`, `first appears in …, not
  …` findings and an occurrence count of 2 at `K`.

- **D-6 — The two completed SPECs.** Shape of the S5 edit, mirroring v0.3.0:
  (1) frontmatter `version: "0.4.0"`, `updated:` the landing date; (2) HISTORY row
  v0.4.0 above the 0.3.0 row, in that table's four-column format; (3) on REQ-JEVC-011
  and REQ-JEVC-012 an `[AMENDED <date> — v0.4.0; see HISTORY]` marker after the
  existing two, the sentence "…with exactly one exception" reworded to name both
  exceptions, and a bold paragraph "**The second exception (v0.4.0).**" placed after
  the v0.3.0 exception paragraph and before the next `**REQ-` or heading (so
  `grReqBody` still sees it as part of the requirement); (4) the two
  `Out of Scope — authority` bullets each gain the second exception inside the
  existing bullet — **no third bullet** (P1 counts two) — and each keeps
  `contract-mode Kickoff` and `llm+jev`; (5) "Still excluded after the v0.2.0 and
  v0.3.0 amendments" is extended to v0.4.0. The S6 edit: version `"0.2.0"`, a HISTORY
  row in the Version/Date/Changes/Author format, and a scope sentence appended to
  REQ-MT-014 and REQ-MT-015.

- **D-7 — Comment-only verification.** A changed-line filter over `git diff -U0`
  for the two Go files proves every changed line is a `//` comment line; it runs
  once at M4, it is not a permanent test.

- **D-8 — Tier M (kept after the run).** Evidence at plan time: no constitutional
  clause names the principle (`git grep -c -i -E 'jev|display-only'` over
  `zone-registry.md` and `moai-constitution.md` prints nothing, exit 1); no new
  package and no behavior change; 12 requirements and 14 criteria against the Tier M
  ceilings 16/16; 13 distinct edits (the plan-audit iteration 2's count, each
  live/mirror pair and the generated hash line counted once). The plan estimated "the
  new test about 250-350 lines — under 1000 on any reading". **The run measured
  otherwise:** the guard is **846 lines** (`wc -l internal/template/jev_auto_exception_test.go`),
  about 2.4 times the top of the estimate (the run record counts 32 passing and 7
  skipping subtests at G, plus the locator, fixture and claim tables;
  `progress.md` §E.2 finding 1). Measured
  sizes: G (`6d012fd4d`) 3 files, +869/−2, 846 of the insertions the guard; K
  (`7983d9131`) 19 files, +80/−28 (17 non-SPEC files plus the two completed SPEC
  bodies); G and K together +949; `git diff --shortstat 5f8c6e051 e70578c24` counts
  21 files, +1061/−30 for the whole run including its own evidence record, and
  17 files, +909/−17 outside `.moai/specs`. "Under 1000" therefore holds for G plus K
  (949) and for the deliverables outside the SPEC directory (909) and does **not**
  hold if the run's evidence record in `progress.md` is counted (1061); the 300-1000
  LOC band of Tier M contains the first two. **File count is the one column that
  points at Tier L:** 17 distinct non-SPEC files (19 with the two completed SPEC
  bodies) against Tier M's 5-15, with six of them mirror copies of one edit and one a
  generated hash line. Tier M is kept on the evidence above, **not** on the cost of a
  `design.md` — the iteration-1 text gave that reason, and the tier table names it as
  its anti-pattern. The classification is the independent auditor's, not the
  author's: `progress.md` §G OD-6 deferred it, and plan-audit iteration 2 judged Tier
  M (PASS 0.94 against 0.80; under a Tier L reading the same measurements score 0.88
  against 0.85, with an absent `design.md` as non-blocking debt). The run does not
  reopen it.

**Assumptions flagged for confirmation (A):**

- **A-1 — Scope extensions X1-X5** are included. X1-X4 were confirmed by the
  operator (`progress.md` §G OD-1) and X5 by the operator after plan-audit
  iteration 1 (OD-3). X1 is the guide the first guard names; X2-X4 state the
  principle for `manager-todo`; X5 is the question-design reference skill. Cut
  candidates in order: X3 (catalog-hash regeneration), X5 (catalog-hash
  regeneration; the operator's own addition), X4 (always-loaded bytes), X2.
  Evidence: `spec.md` §B.3.
- **A-2 — Amendment style** for the two completed SPECs is the v0.3.0 in-place
  style with `status: completed` unchanged, not the `completed → in-progress
  (amendment)` transition (§B.7). **Decided** — `progress.md` §G OD-4
  (orchestrator, 2026-10-02) applies it to both SPECs, on the precedent commit
  `185569ef3` and on what the audits observed — iteration 1: both lint-clean before
  the edit; iteration 2: `SPEC-JEV-CORE-001` already carries two in-place amendments
  and lints clean, and a scratch copy of `SPEC-MANAGER-TODO-001` with the planned edit
  added no finding (spec.md §B.7 states it at that strength). The in-tree result after
  the real edit is now observed: `moai spec lint` prints `No findings` for both.
  `SPEC-JEV-CORE-001` went 0.3.0 → 0.4.0 and `SPEC-MANAGER-TODO-001` 0.1.0 → 0.2.0
  (D-6 planned both). The `status: completed` pin of `TestJevDoctrineAmendment` (P1)
  exists for `SPEC-JEV-CORE-001` only.
- **A-3 — Tool-level description string unchanged** (§B.4). OD-4 does not decide this
  one; it stays the author's decision: no test pins the string, but changing it
  would tell agents about a path the tool refuses, so it is left alone and listed in
  the report for the orchestrator.
- **A-4 — The marker date** in `[AMENDED <date> — v0.4.0; see HISTORY]` is the landing
  date of the linked commit, written by `manager-spec` at M2; the guard's token for
  the SPEC files is the universal literal, not the date.
- **A-5 — `phase: "v3.2.0 target"`** follows the predecessor SPEC; the tree reports
  `v3.2.0-rc.23`.

## §C Pre-flight (the run inherits)

Every Go command in this SPEC runs with all eleven lane variables scrubbed, in one
compound invocation (a separate `unset` does not carry):

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_WORKER MOAI_FACTORY_ROLE MOAI_FACTORY_WORKERS MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_AUTO_DISPATCH MOAI_KANBAN_BACKEND && go test -count=1 -v -run '^<TestName>$' ./internal/<pkg>/
```

All `-run` patterns are anchored (`^…$`). A green `-run` with an empty swept set is a
vacuous pass, so the swept count is read from `go test -list` first.

1. `git rev-parse --short HEAD` and `git branch --show-current`; re-read before every
   commit (`AGENTS.md` §2).
2. Baseline: re-run the guard set of §Findings (c); every one must PASS before any
   edit (research.md §R3 is the plan-time reading at `c50da9c2f`).
3. `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...`.
4. Sweep control for the new tests: `go test -list '^(TestJevAutoExceptionLinkage|TestJevAutoExceptionWording)$' ./internal/template/`
   prints no test name today (acceptance.md ledger row L7, informational). After M1 it
   prints exactly the two names and `ok`.
5. A tree-local `moai` for lint: `go build -o <scratchpad>/moai ./cmd/moai`; invoke it
   by path, and cite its tree HEAD next to every lint result. The binary carries no
   commit stamp — `moai version` prints `v3.1.3 none built unknown` when built without
   ldflags (observed at `e70578c24`) — so the judging build's commit cannot be read
   from the binary: the provenance is "built from tree HEAD `<sha>` by construction,
   in the same session as the lint", stated beside the result, not read off the binary
   (`verification-claim-integrity.md` §2.2).
6. Verification cost of the linkage guard. `TestJevAutoExceptionLinkage` shells out to
   `git`: with `armed=true` the `tree` subtest runs `git log -S<token>` once per
   registry row (18 rows) plus the presence reads, and with `armed=false` it only
   asserts absence. Wall time therefore tracks machine load, not the tree: the run
   recorded 110-210 s at a load average of about 300 (M1, `progress.md` §E.2) and 8.26 s
   at load averages 40.5 43.8 117.4 (M4); the reconciliation measured 6.10 s
   (`--- PASS: TestJevAutoExceptionLinkage (6.10s)`, `ok … 6.550s`) at 20.53 29.39
   78.90. The cost is git subprocess launches (the run's reading, `progress.md` §E.2
   finding 4); it was not profiled here.
7. Worktree guard: use plain, separate commands. Two compound forms were refused in
   this plan run (a `for` loop over `git grep`; a `git grep` redirected to a
   `$VAR/…` path) — see §I.

## §D Constraints

- **Template-First.** For each pair: edit `internal/template/templates/<path>` first,
  run `make build`, then bring the live copy to the same wording. Template files
  carry no SPEC id, requirement token, ISO date or commit hash (P9). Citations such as
  `SPEC-TODO-AUTO-PRIORITY-001` live only in SPEC files and commit messages.
- **The pinned text of `spec.md` §B.8 is not edited.** In particular: exactly two
  authority bullets; `contract-mode Kickoff` exactly once per catalogue row and per
  `workflow.yaml` jev comment; the guide's `:32` paragraph verbatim; `AGENTS.local.md`
  untouched; `kickoff.go`, `activation_test.go` and `contract_mode_blocks_test.go`
  untouched.
- **Do not edit** `kanban-dispatch.md`, `workflows/gtd.md`, `manager-todo.md` or its
  emissions (P8), `todo_auto_rank.go`, `todo_auto.go`, `defaults.go`, `catalog.go`,
  the `jev_ask` description string, `auto-semantics.md`.
- **`CLAUDE.md` growth (X4).** Net growth of at most 160 bytes in each copy, both
  copies byte-identical, total far below the 40,000-character ceiling.
- **Comments in Go files** respect the `code_comments` setting (English), contain no
  `NearDuplicateMark`, `LaneQuestionRoute` or `SkillSuggest`, and leave every import
  unchanged.
- **The X5 skill carries no call path.** The amended sentence contains none of
  `internal/jev`, `mcp__moai__jev`, `jev_ask`, `moai jev` (P11), and the two copies are
  byte-identical.
- **The arming literal stays out of the guard's source** (D-5): the test file spells
  `jevAutoExceptionAmended = true` only as the standalone const declaration written in
  the linked commit.
- **Scoped tests only.** No `go test ./...` locally (`AGENTS.local.md` §4); CI gives
  the full-suite verdict after the leader's push.
- **One writer per tree.** M2 (`manager-spec`) and M3 (`manager-develop`) write
  sequentially in the same worktree, never concurrently.
- **Stage by explicit pathspec**; never `git add -A`, `git add .` or `git commit -a`.

## §E Milestones

Milestones are the delegation unit; the **commit** unit differs for M2-M3 because the
guard's first-commit rule makes the marker-bearing edits one commit (D-5, REQ-JAE-011).

### M1 (High) — the guard, first (`manager-develop`, `cycle_type: tdd`)

Create `internal/template/jev_auto_exception_test.go`: the marker registry and anchors
of D-3 (the arming token built by parts, D-5), `autoExceptionLinkageFindings`, the
wording tuples and the claim and over-reach sets of the tables below, and the two tests
with their subtests. RED first: write the falsifier subtests against a stub that
returns no findings and capture the failing output verbatim (manager-develop §E8);
then implement until green. Subtests —
`TestJevAutoExceptionLinkage`: `falsifier/partial/<n>` (one per registry row, all rows
but one present), `falsifier/arming-only`, `falsifier/self-match` (a fixture whose file
under test spells the full arming literal beside `= false`: the checker must report it
present — the mutant the plan-audit named), `falsifier/split-commits`,
`falsifier/dangling-anchor`, `all-in-one-commit`, `tree`;
`TestJevAutoExceptionWording`: `falsifier/literal-missing`,
`falsifier/closed-target-dropped`, `falsifier/bound-in-other-paragraph`,
`falsifier/long-form-literal-missing` (a long-form passage without
`workflow.jev.enabled` or the filter literal), `falsifier/claim-phrasing`,
`falsifier/over-reach-phrasing`, `disclaimer-not-flagged` (a passage carrying
`no ordering accuracy is claimed` and `is not claimed to be accurate` is accepted),
`mirror-parity` (the added block is identical in every live/mirror pair), one subtest
per surface group (`go-comments`, `config-comment`, `catalogue-rows`, `spec-core`,
`spec-manager-todo`, `local-guide`, `extension-rows`). With
`jevAutoExceptionAmended = false` the group subtests assert nothing about the
surfaces yet (they skip with a logged reason) and the tree subtest asserts all
markers absent. Fixtures follow `TestJevAmendmentLinkage`: a temp repo, one commit
per scenario, `grGit` for `git`. Commit alone:
`test(SPEC-JEV-AUTO-EXCEPTION-001): M1 linkage and wording guards (armed=false)`.
This is the first run-phase commit, so `manager-develop` also flips `status:
draft → in-progress`.

Exit: AC-JAE-011, AC-JAE-012 in their armed=false form — the falsifier subtests PASS, the group subtests skip with a logged reason, both tests PASS, `-list` shows both names, the five neighbouring guards of §Findings (c) still PASS; both criteria complete at M3, where `armed=true` and the groups run.

Wording tuples the group subtests check. "Required" lists the literals every passage
of the group must carry in the same passage; the long-form groups (S1, S2, S3, S5, X1)
also carry `workflow.jev.enabled` and the filter literal.

| Group | Passage locator | Required in the passage | Closed targets retained |
|---|---|---|---|
| `go-comments` (long-form) | the comment block containing "display-only" in `jev.go` and `mcp_jev.go` | `auto-scoped ranking exception`, `selection order only`, `workflow.jev.enabled`, `mechanical filters`; `mcp_jev.go` also `this tool` | `mcp_jev.go`: `completion verdict`, `merge approval`, `queue mutation` |
| `config-comment` (long-form) | `# jev: ` … `\n    jev:` in both `workflow.yaml` copies | both literals, `workflow.jev.enabled`, `mechanical filters` | `completion verdict`, `merge`, `queue mutation`, `never decides alone` |
| `catalogue-rows` (short-form) | the two row lines in both catalogue copies | both literals, `never through this tool` | `completion predicate`, `merge approval`, `queue mutation`, `never decides alone` |
| `spec-core` (long-form) | REQ-JEVC-011 and REQ-JEVC-012 bodies; the two authority bullets (the bullets are held to both literals and the closed targets only) | both literals, `workflow.jev.enabled`, `mechanical filters` in the two requirement bodies; `v0.4.0` marker | `contract-mode Kickoff`, `llm+jev` (P1) |
| `spec-manager-todo` (short-form) | REQ-MT-014 and REQ-MT-015 lines | both literals | `display-only`, `never as authority` / `queue mutation` |
| `local-guide` (long-form, Korean) | the paragraph following the pinned one in the guide | both literals, `workflow.jev.enabled`, `기계적 필터` | the pinned `:32` text verbatim (P3) |
| `extension-rows` (short-form) | `agent-authoring.md:147`, `SKILL.md:180`, `CLAUDE.md:63`, the reference-skill sentence, live and mirror | both literals | the original sentence's own clauses; the reference skill keeps `completion predicate`, `merge approval`, `queue mutation` and none of the P11 tokens |

Claim set and over-reach set, applied to every amended passage of every group (English
patterns; case-insensitive). They were exercised on eight phrase cases before being
written here (research.md §R6): flagged — `…selection order only, and the ordering is
accurate`, `…the ordering beats the fallback`, `…with a measured accuracy of 80%`, `…it
applies to every moai todo pick`; not flagged — `no ordering accuracy is claimed`, `and
is not claimed to be accurate`, `…never the basis of a completion verdict, a merge
approval or any other decision that is hard to undo`, `any other decision that is hard
to undo`. The reference expressions the run phase may refine as long as those eight
cases keep their outcomes:

- claim set: `\b(is|are|was|were) (accurate|measured|reliable|validated)\b`, `\bbeats\b`,
  `\boutperform`, `measured accuracy`, `accuracy of \d`
- over-reach set: `\b(every|any|all) (` + optional backtick + `moai todo|picks?)\b`

### M2 (High) — the two completed SPEC bodies (`manager-spec`, orchestrator re-delegation)

Edit `SPEC-JEV-CORE-001/spec.md` and `SPEC-MANAGER-TODO-001/spec.md`
per D-6. Left **uncommitted** in the working tree. Verification (not a commit gate):
`moai spec lint SPEC-JEV-CORE-001` and `SPEC-MANAGER-TODO-001` with the tree-local
binary report `No findings`; the pre-edit baseline is `No findings` for both
(research.md §R3).

Exit: AC-JAE-004, AC-JAE-005 — the two completed SPEC bodies carry the v0.4.0 and scope text, lint-clean (verified again at M3, where the tests that read them run).

### M3 (High) — the rest of the linked set, then one commit (`manager-develop`; the lane orchestrator stages)

In Template-First order: template mirrors (`workflow.yaml`, catalogue, and the
confirmed extensions including the reference skill), `make build` (regenerates the
`catalog.yaml` hashes for X3 and X5), live copies, the Go comments (S1, S3), the guide
paragraph (X1), then flip `jevAutoExceptionAmended` to `true` as a standalone `const`
line. Run the §F scoped tests. The lane orchestrator then re-reads `git rev-parse
--short HEAD` and `git branch --show-current`, runs `git status --short`, and stages by
explicit pathspec every file of M2 and M3 (plus `internal/template/catalog.yaml` if X3
or X5 survives) into **one** commit:
`docs(SPEC-JEV-AUTO-EXCEPTION-001): M2-M3 linked Jev auto-exception amendment (t1403)`.

Exit: AC-JAE-001, AC-JAE-002, AC-JAE-003, AC-JAE-006, AC-JAE-007, AC-JAE-008, AC-JAE-009, AC-JAE-011, AC-JAE-012 — the guard's `tree` subtest PASSes with `armed=true`, the seven group subtests PASS, and every pinned guard of §F still PASSes.

### M4 (Medium) — evidence and closure (`manager-develop`, then `manager-docs`)

The AC matrix with verbatim output (manager-develop §E1-E8 self-verification), the
changed-line filter of D-7, the three inventory sweeps of §F, the commit-order
evidence of AC-JAE-013 (both SHAs recorded in `progress.md`), `progress.md`
§E.2/§E.3. The sync phase (`manager-docs`) carries the CHANGELOG entry and the
`completed` transition; the sync-audit re-reads the decision record.

Exit: AC-JAE-010, AC-JAE-013, AC-JAE-014 — no behavior change shown, the commit order witnessed by the graph, and the inventory closed over the three stated sweeps.

Files changed (planned): `internal/jev/jev.go`, `internal/cli/mcp_jev.go`,
`internal/template/jev_auto_exception_test.go` (new), `.moai/config/sections/workflow.yaml`
and its template, the catalogue and its template, `SPEC-JEV-CORE-001/spec.md`,
`SPEC-MANAGER-TODO-001/spec.md`, `.moai/docs/jev-local-operations.md` — 10 files; plus,
if kept, `agent-authoring.md` ×2, `SKILL.md` ×2, `CLAUDE.md` ×2, the reference skill
×2 and `catalog.yaml` — 19. `progress.md` of this SPEC in addition.

## §F Verification commands (all Go commands carry the §C scrub prefix)

| # | Command | Expected |
|---|---|---|
| V1 | `go test -count=1 -list '^(TestJevAutoExceptionLinkage\|TestJevAutoExceptionWording)$' ./internal/template/` | exactly the two names, then `ok` |
| V2 | `go test -count=1 -v -run '^(TestJevAutoExceptionLinkage\|TestJevAutoExceptionWording)$' ./internal/template/` | both `--- PASS`, every subtest PASS, `tree` logs `armed=true` after M3 |
| V3 | `go test -count=1 -v -run '^TestJevDoctrineAmendment$' ./internal/template/` | `--- PASS` with `spec`, `rules-and-config`, `local-guide` PASS |
| V4 | `go test -count=1 -v -run '^TestJevAmendmentLinkage$' ./internal/contract/kickoff/` | `--- PASS` (6 subtests) |
| V5 | `go test -count=1 -v -run '^(TestMCPToolCatalogueDocsStayMirrorIdentical\|TestAutoRankDoctrineAmendment\|TestAutoRankMirrorParity\|TestAutoRankMarkerDisclosure\|TestAutoHelpAndRefusalDoNotAssertPickOrder\|TestAutoRankAgentDoctrine)$' ./internal/cli/` | six `--- PASS` |
| V6 | `go test -count=1 -v -run '^TestNoConsumerCallPathShips$' ./internal/jevmeasure/` | `--- PASS` |
| V7 | `go test -count=1 -v -run '^(TestPackageImports_AreStandardLibraryOnly\|TestImportClassifierPositiveControl)$' ./internal/jev/` | two `--- PASS` |
| V8 | `go test -count=1 -v -run '^TestTemplateNoInternalContentLeak$' ./internal/template/` | `--- PASS` |
| V9 | `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 both |
| V10 | `<scratchpad>/moai spec lint SPEC-JEV-CORE-001` and `… SPEC-MANAGER-TODO-001` and `… SPEC-JEV-AUTO-EXCEPTION-001` | `No findings` each |
| V11 | `git diff -U0 <CARD_BASE> -- internal/jev/jev.go internal/cli/mcp_jev.go`, `CARD_BASE` from `git merge-base develop HEAD` taken at run time | every `+`/`-` content line is a `//` comment line; no hunk touches the `mcp.WithDescription` line |
| V12 | the three inventory sweeps, each excluding `.moai/specs`, `.moai/reports`, `CHANGELOG.md` — primary: `git grep -n -i -E 'display-only\|display only' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md'`; synonym: `git grep -n -E 'never reorder by inferred priority\|판단 자료\|모델 답을 입력으로도' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md'`; closed-target phrases: `git grep -n -i -E 'a person reads\|labelled model\|queue mutation\|hard to undo' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md'` | every hit's file is in the `research.md` §R1 tables — the guard source `internal/template/jev_auto_exception_test.go` (class iii) is the one file the sweeps gain over their baselines, and at `e70578c24` they print primary 76 hits / 44 files, synonym 6 hits, closed-target 96 hits / 48 files; each class-(i) passage carries both literals; controls: `internal/mcp/catalog.go:97` (class iii, untouched) is present in the primary sweep, and `internal/cli/todo_triage.go` (class iii, untouched) is present in the closed-target-phrase sweep |
| V13 | `go test -count=1 -v -run '^(TestJevQuestionDesignSkillCarriesNoCallPath\|TestJevQuestionDesignSkillCopiesStayIdentical)$' ./internal/cli/` | two `--- PASS` (P11, X5) |
| V14 | `go test -count=1 -v -run '^TestCatalogHashParity$' ./internal/spec/` | `--- PASS`, `0 drift` (P12, X3 and X5 hash regeneration) |

The closure each sweep supports is over its own stated pattern only (spec.md §G R-8).

`CARD_BASE` is taken from `git merge-base develop HEAD` at evaluation time and never
pinned: the local rule forbids a literal base SHA for "what did this card change"
(`.claude/rules/local/gitflow-lane-protocol.md` §8), and the range is meaningful
only before the card merges.

## §G Anti-patterns (named for the run to refuse)

- Appending rows to `linkageMarkers` in `activation_test.go` — fails on first-commit
  (spec §A.4); and any edit to that file.
- Editing the guide's pinned `:32` paragraph, or `design.md` §11.1 — breaks P3.
- Adding a third bullet under `### Out of Scope — authority`, or a `- ` line anywhere
  between that heading and the next `## ` — breaks P1's count of two.
- Writing `contract-mode Kickoff` a second time in a catalogue row or the `workflow.yaml`
  jev comment — breaks P2's "exactly once".
- Putting the exception into the `jev_ask` description string or the `catalog.go`
  comment — widens the tool (REQ-JAE-003).
- Landing the guard and the markers in one commit — the ordering would be unwitnessable
  (`verification-claim-integrity.md` §2.3).
- Splitting the marker-bearing edits across commits "because the milestones are
  separate" — the guard then fails by design.
- A wording that says or implies the Jev ordering is accurate, measured, or shipped
  as default — decision 6.
- A SPEC id, `REQ-` token, date or hash in any template mirror.
- Spelling `jevAutoExceptionAmended = true` as one literal anywhere in the guard's source
  (registry row, comment, fixture, message) — it makes the arming marker first appear in
  the guard's own commit (D-5).
- Putting the arming constant inside a `const (…)` block, where `gofmt` aligns the `=`
  and breaks the literal.
- Naming the wrapper tool or its call path in the reference skill (`internal/jev`,
  `mcp__moai__jev`, `jev_ask`, `moai jev`) — P11 fails.
- `go test ./...` locally.

## §H Cross-references

`spec.md`; `acceptance.md`; `research.md`; `SPEC-TODO-AUTO-PRIORITY-001` (spec §B.5,
§D, plan §B "Deferred follow-up"); `SPEC-JEV-CORE-001`; `SPEC-MANAGER-TODO-001`;
`SPEC-AUTONOMY-GATE-REWIRE-001/design.md` §11.1-§11.2 (the first amendment's text and
assembly); `internal/contract/kickoff/activation_test.go` (the first guard);
`internal/template/contract_mode_blocks_test.go` (the first content guard);
`.claude/rules/moai/core/verification-claim-integrity.md` §2.2, §2.3;
`.claude/rules/moai/development/verification-completeness.md` §1-§2.

## §I Gaps (unobserved in this run)

- **G-1 — Guard-refused commands.** Two compound forms were refused by the worktree
  guard during this run and re-run as plain separate commands: a `for` loop over
  `git grep` (re-run as five single `git grep` calls) and a `git grep … >
  $VAR/file` redirect (re-run with a literal path). Neither result was inferred.
- **G-2 — Run-phase commands are unrun.** V1-V2, V11 and the sweep V12 as a closure
  check are run-phase commands; only their RED-now counterparts and the baseline
  guards were run now. `go test -list` for the new names was run (empty).
  *Status after the run (0.1.2):* closed — the run exercised V1-V12 (`progress.md`
  §E.2 E1 and the sweeps); the reconciliation re-ran V1-V2, V3 and the three sweeps
  at `e70578c24`.
- **G-3 — The 15-name ranking suite** was not re-run in full; eight names were
  (research.md §R3).
- **G-4 — Draft wording is untested.** D-1 and D-2 are reference wording; whether
  each rewritten passage still satisfies its pin is established only at M3.
  *Status after the run (0.1.2):* closed — at `e70578c24` `TestJevDoctrineAmendment`
  passes 6/6 and the seven group subtests of `TestJevAutoExceptionWording` run
  (`armed=true`) and pass over the amended passages.
- **G-5 — `grReqBody`'s paragraph boundary** for the new "second exception"
  paragraph was reasoned from reading the helper (`contract_mode_blocks_test.go:833-846`),
  not exercised. *Status after the run (0.1.2):* closed — `TestJevDoctrineAmendment/spec`
  reads the amended REQ-JEVC-011/-012 bodies with the new paragraph and passes.
- **G-6 — `moai spec lint` on the amended completed SPECs** is a post-M2 check; only
  the pre-edit baseline was run. *Status after the run (0.1.2):* closed — `moai spec
  lint SPEC-JEV-CORE-001` and `SPEC-MANAGER-TODO-001` each print `✓ No findings — all
  SPEC documents are valid` (tree-local build from tree HEAD `e70578c24`).
- **G-7 — The D1 draft re-implements the predecessor's mechanics; it is not the
  predecessor's test.** It reproduces the presence check (`strings.Contains`), the
  first-commit check (`git log --reverse -S`) and the tree-versus-commit behavior in a
  throwaway program outside the repository (research.md §R6). That the real
  `TestJevAmendmentLinkage` helpers behave identically is established by reading
  `activation_test.go:44-50,179`, not by running the draft against them; the real guard
  is first exercised at M1. *Status after the run (0.1.2):* closed — the real guard
  ran RED against stubs and GREEN at G (`falsifier/self-match` included), and
  `TestJevAutoExceptionLinkage` passes with `armed=true` at `e70578c24`; the draft
  remains a re-implementation, and the guard's own `git log -S` is the real check.
- **G-8 — The claim and over-reach expressions were tried on eight phrase cases**
  (research.md §R6), not on the amended passages, which do not exist yet. Whether the
  final wording trips them is established at M3. *Status after the run (0.1.2):* the
  final passages pass the groups; the expressions were refined by two narrowing
  exemptions, one false positive remains and several miss classes are measured
  (spec.md §G R-5, research.md §R7).
- **G-9 — The third sweep is itself a pattern** (spec.md §G R-8): it found X5, and a
  restatement phrased without `a person reads`, `labelled model`, `queue mutation` or
  `hard to undo` would still be missed.
- **G-10 — Revision-phase measurements ran at tree `1eef55dd9` plus this revision's
  uncommitted SPEC edits.** No Go or template file differs from the pinned tree
  (`git diff --name-only c50da9c2f 1eef55dd9` lists only this SPEC's directory).
