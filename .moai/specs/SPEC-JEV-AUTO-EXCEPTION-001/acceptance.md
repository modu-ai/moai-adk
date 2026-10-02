# SPEC-JEV-AUTO-EXCEPTION-001 — Acceptance Criteria

> This is the **verification layer**: each criterion is an `AC-JAE-XXX` written as
> Given-When-Then and is binary-testable. The requirement layer (GEARS) is
> `spec.md` §C. Every criterion opens with **Covers**, the `REQ-JAE-XXX` it covers,
> and its verdict is the real output of the named command. 14 criteria (Tier M
> ceiling 16).
>
> **Command convention.** RED-now cells cite a ledger row: a plain single-invocation
> command (no pipe, redirection, `&&`, `;` or subshell — `verification-completeness.md`
> §2.1), its verbatim stdout, and its exit code as a separate field. Green-path Go
> commands carry the scrub prefix `<SCRUB>`, which unsets the eleven lane variables of
> `plan.md` §C inside one compound invocation (`unset … && go test …`); that prefix
> contains `&&`, so a ledger row that carries it (L7) is informational and no
> criterion's RED-now rests on it alone. `-run` patterns are anchored (`^…$`); an
> unanchored pattern also selects longer names, which is a vacuous-pass shape.
> Commands that need the card's base use `CARD_BASE` from `git merge-base develop HEAD`
> evaluated at run time, never a literal SHA, and only before the card merges.
>
> **Classification.** **release-blocking**: a RED-now ledger row exists and reproduces
> at tree `c50da9c2f` (11 criteria: AC-JAE-001..007, 011..014). **regression-guard**:
> no RED-now, GREEN from arrival (3 criteria: AC-JAE-008, 009, 010), each with a
> mutant probe instead.
>
> **RED-now / green-path pairing.** Every release-blocking criterion pairs the
> ledger row that shows why it is red now with the milestone that flips it
> (`plan.md` §E). Surfaces whose RED is "the literal is absent" are red for the right
> stated reason: the exception is unwritten there, and only this SPEC's work can
> write it. The behavior-bearing criteria (AC-JAE-011, 012) are red because the test
> does not exist (L7); the failing output of each falsifier before its implementation
> is captured verbatim at M1 (manager-develop §E8), not in this ledger.

## Evidence ledger (RED-now observations, tree `c50da9c2f`)

Whole-ledger pin: tree `c50da9c2f` (HEAD, branch `WT-jev-auto-exception`). Row ids are
`L<n>` so they do not collide with manager-develop's self-verification items E1-E8.

| Id | Command | Verbatim stdout | Exit | Why it is red |
|---|---|---|---|---|
| L1 | `git grep -c -F "auto-scoped ranking exception" -- internal/jev/jev.go internal/cli/mcp_jev.go` | (empty) | 1 | neither Go comment names the exception |
| L2 | `git grep -c -F "auto-scoped ranking exception" -- .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml` | (empty) | 1 | the jev comment, live and mirror, has no second exception |
| L3 | `git grep -c -F "auto-scoped ranking exception" -- .claude/rules/moai/core/moai-mcp-tools-catalogue.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | (empty) | 1 | both catalogue copies carry only the Kickoff exception |
| L4 | `git grep -c -F "auto-scoped ranking exception" -- .moai/specs/SPEC-JEV-CORE-001/spec.md .moai/specs/SPEC-MANAGER-TODO-001/spec.md` | (empty) | 1 | neither completed SPEC states the exception |
| L5 | `git grep -c -F "auto-scoped ranking exception" -- .moai/docs/jev-local-operations.md` | (empty) | 1 | the doctrine guide says "one place only" |
| L6 | `git grep -c -F "auto-scoped ranking exception" -- .claude/rules/moai/development/agent-authoring.md .claude/skills/moai/SKILL.md CLAUDE.md` | (empty) | 1 | the three live extension surfaces lack it |
| L6b | `git grep -c -F "auto-scoped ranking exception" -- internal/template/templates/.claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/skills/moai/SKILL.md internal/template/templates/CLAUDE.md` | (empty) | 1 | and so do their template mirrors |
| L7 | `<SCRUB> go test -count=1 -list '^(TestJevAutoExceptionLinkage\|TestJevAutoExceptionWording)$' ./internal/template/` | `ok  	github.com/modu-ai/moai-adk/internal/template	0.292s` | 0 | zero test names: neither guard exists (informational row) |
| L8 | `git grep -c -F "auto-scoped ranking exception" -- .claude/rules/moai/workflow/kanban-dispatch.md .claude/skills/moai/workflows/gtd.md .claude/agents/moai/manager-todo.md` | `.claude/agents/moai/manager-todo.md:2` · `.claude/rules/moai/workflow/kanban-dispatch.md:1` · `.claude/skills/moai/workflows/gtd.md:1` | 0 | **positive control**: the landed anchors do carry the token, so the zero rows above are about the Jev side, not about a broken probe |
| L9 | `git grep -n -F "The capability is display-only" -- .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml` | `.moai/config/sections/workflow.yaml:227:    # The capability is display-only: an answer is a labelled model-produced` · `internal/template/templates/.moai/config/sections/workflow.yaml:229:    # The capability is display-only: an answer is a labelled model-produced` | 0 | the not-yet-amended side, live and mirror (interim split evidence) |
| L10 | `git grep -n -F "as a display-only signal for dispatch order and priority, except for" -- .claude/agents/moai/manager-todo.md` | `.claude/agents/moai/manager-todo.md:7:  Jev as a display-only signal for dispatch order and priority, except for` | 0 | the landed side states the exception (interim split evidence) |
| L11 | `git grep -c -F "DISPLAY-ONLY: the answer is a labelled model signal a person reads" -- internal/cli/mcp_jev.go` | `internal/cli/mcp_jev.go:1` | 0 | baseline for the unchanged description string (AC-JAE-010) |
| L12 | `git grep -c -F "jevAutoExceptionAmended" -- internal/template` | (empty) | 1 | the arming constant does not exist |
| L13 | `git grep -c -F "v0.4.0" -- .moai/specs/SPEC-JEV-CORE-001/spec.md` | (empty) | 1 | no v0.4.0 marker |
| L14 | `git grep -n -F "version: \"0.3.0\"" -- .moai/specs/SPEC-JEV-CORE-001/spec.md` | `.moai/specs/SPEC-JEV-CORE-001/spec.md:4:version: "0.3.0"` | 0 | version still 0.3.0 |
| L15 | `git grep -c -F "doc_display_only_test" -- internal/jev/jev.go` | `internal/jev/jev.go:1` | 0 | `jev.go:27` names a test file that does not exist (`display_only_test.go` does) |

## AC-JAE-001 — The two Go comments name the exception and stay comments

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-003, REQ-JAE-008

- **Class**: release-blocking (RED-now: L1, L15).
- **Given** the tree after M3, **When** the two files are read with `git grep` and the
  wording guard runs, **Then** `internal/jev/jev.go` and `internal/cli/mcp_jev.go` each
  carry `auto-scoped ranking exception` and `selection order only` in the comment block
  that says "display-only"; `mcp_jev.go` also says the exception does not reach "this
  tool"; `jev.go` no longer names `doc_display_only_test.go`; neither file contains
  `NearDuplicateMark`, `LaneQuestionRoute` or `SkillSuggest`.
- **Green** (M3):
  `git grep -c -F "auto-scoped ranking exception" -- internal/jev/jev.go internal/cli/mcp_jev.go`
  → `…jev.go:` and `…mcp_jev.go:` each ≥ 1, exit 0; the same with `"selection order only"`;
  `git grep -c -F "doc_display_only_test" -- internal/jev/jev.go` → empty, exit 1;
  `git grep -c -E "NearDuplicateMark|LaneQuestionRoute|SkillSuggest" -- internal/jev/jev.go internal/cli/mcp_jev.go`
  → empty, exit 1; and
  `<SCRUB> go test -count=1 -v -run '^TestJevAutoExceptionWording$' ./internal/template/`
  → `--- PASS: TestJevAutoExceptionWording/go-comments `.
- **Mutant probe.** A comment that carries both literals in a *different* paragraph from
  "display-only" passes the greps and fails the same-passage assertion. A comment that
  also drops "a queue mutation" from `mcp_jev.go` fails the closed-target assertion.

## AC-JAE-002 — The `workflow.yaml` jev comment, live and mirror

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-004, REQ-JAE-005

- **Class**: release-blocking (RED-now: L2; context: L9).
- **Given** the tree after M3, **When** the grep and both guards run, **Then** both
  `workflow.yaml` copies carry the two literals inside the `# jev: ` comment run; the run
  still names `contract-mode Kickoff` exactly once, `llm+jev`, `never decides alone`,
  `completion verdict`, `merge` and `queue mutation`; and the added block is identical in
  the two copies.
- **Green** (M3): `git grep -c -F "auto-scoped ranking exception" -- .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml`
  → one line per file, each ≥ 1, exit 0; `<SCRUB> go test -count=1 -v -run '^TestJevDoctrineAmendment$' ./internal/template/`
  → `--- PASS: TestJevDoctrineAmendment/rules-and-config `; and
  `<SCRUB> go test -count=1 -v -run '^TestJevAutoExceptionWording$' ./internal/template/`
  → `--- PASS` for `config-comment` and `mirror-parity`.
- **Mutant probe.** Repeating `contract-mode Kickoff` in the new block fails the
  existing "exactly once" assertion. Editing only the live copy fails `mirror-parity`.

## AC-JAE-003 — The catalogue rows, both copies, byte-identical

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-003, REQ-JAE-004, REQ-JAE-005

- **Class**: release-blocking (RED-now: L3).
- **Given** the tree after M3, **When** the grep, the mirror test and the guards run,
  **Then** the `| \`mcp__moai__jev_ask\` |` row and the `| Judgment (gated) |` row of each
  copy carry the two literals and the words `never through this tool`, keep
  `contract-mode Kickoff` once, `llm+jev`, `never decides alone`, `completion predicate`,
  `merge approval` and `queue mutation`, and the two copies are byte-identical.
- **Green** (M3): `git grep -c -F "auto-scoped ranking exception" -- .claude/rules/moai/core/moai-mcp-tools-catalogue.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md`
  → both `:2`, exit 0; the same with `"never through this tool"` → both `:2`;
  `<SCRUB> go test -count=1 -v -run '^TestMCPToolCatalogueDocsStayMirrorIdentical$' ./internal/cli/`
  → `--- PASS`; `TestJevDoctrineAmendment/rules-and-config` → `--- PASS`; the wording guard's
  `catalogue-rows` → `--- PASS`.
- **Mutant probe.** A clause appended to the template copy only fails the mirror test. A
  clause that says the exception applies "through `jev_ask`" fails the
  `never through this tool` assertion.

## AC-JAE-004 — `SPEC-JEV-CORE-001` v0.4.0

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-005, REQ-JAE-006

- **Class**: release-blocking (RED-now: L4, L13, L14).
- **Given** the tree after M2-M3, **When** the greps, the guards and the linter run,
  **Then** `spec.md` carries `version: "0.4.0"`, a v0.4.0 marker and an exception paragraph
  on REQ-JEVC-011 and REQ-JEVC-012 (both containing the two literals), both
  `Out of Scope — authority` bullets still two in number and each naming `contract-mode
  Kickoff`, `llm+jev` and the new exception, a HISTORY row for v0.4.0, `status: completed`,
  and its `progress.md` unchanged.
- **Green** (M2, verified at M3):
  `git grep -n -F "version: \"0.4.0\"" -- .moai/specs/SPEC-JEV-CORE-001/spec.md` → line `4`,
  exit 0; `git grep -c -F "v0.4.0" -- .moai/specs/SPEC-JEV-CORE-001/spec.md` → ≥ 3 (a marker
  line on each of the two requirements plus the HISTORY row), exit 0;
  `<SCRUB> go test -count=1 -v -run '^TestJevDoctrineAmendment$' ./internal/template/` →
  `--- PASS: TestJevDoctrineAmendment/spec `; the wording guard's `spec-core` → `--- PASS`;
  `<scratchpad>/moai spec lint SPEC-JEV-CORE-001` → `✓ No findings — all SPEC documents are
  valid` (baseline: the same line, plan.md §Findings / research.md §R3); and
  `git diff --stat <CARD_BASE> -- .moai/specs/SPEC-JEV-CORE-001/progress.md` → empty.
- **Mutant probe.** A third bullet under `Out of Scope — authority` fails
  `TestJevDoctrineAmendment/spec` ("authority section has 3 items, want 2"). Changing
  `status:` away from `completed` fails the same subtest. Dropping `contract-mode Kickoff`
  from a bullet fails it.

## AC-JAE-005 — `SPEC-MANAGER-TODO-001` scope sentences

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-007

- **Class**: release-blocking (RED-now: L4).
- **Given** the tree after M2-M3, **When** the grep, the wording guard and the linter
  run, **Then** REQ-MT-014 and REQ-MT-015 each state that they govern the local Jev
  scripts (display-only unchanged) and that the `--auto` cycle's own ranking is the one
  `auto-scoped ranking exception` with `selection order only`; REQ-MT-015 still contains
  `queue mutation`; the HISTORY has a new row; `status: completed`; no requirement id
  removed or renumbered.
- **Green** (M2, verified at M3):
  `git grep -c -F "auto-scoped ranking exception" -- .moai/specs/SPEC-MANAGER-TODO-001/spec.md`
  → ≥ 2, exit 0; the wording guard's `spec-manager-todo` → `--- PASS`;
  `git grep -c -E "^- \*\*REQ-MT-0(14|15)\*\*" -- .moai/specs/SPEC-MANAGER-TODO-001/spec.md`
  → `2`; `<scratchpad>/moai spec lint SPEC-MANAGER-TODO-001` → `✓ No findings — all SPEC
  documents are valid`.
- **Mutant probe.** A scope sentence on REQ-MT-014 only fails the per-requirement
  assertion on REQ-MT-015. A rewrite of REQ-MT-015 that deletes `queue mutation` fails
  the closed-target assertion.

## AC-JAE-006 — The doctrine guide is amended additively

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-005

- **Class**: release-blocking (RED-now: L5).
- **Given** the tree after M3, **When** the grep and the guards run, **Then**
  `.moai/docs/jev-local-operations.md` carries a new paragraph with the two literals that
  scopes the earlier "한 곳뿐" sentence to the Kickoff cross-check and keeps the other
  grade-3 prohibitions, while the pinned paragraph is byte-for-byte unchanged and
  `AGENTS.local.md` is untouched.
- **Green** (M3): `git grep -c -F "auto-scoped ranking exception" -- .moai/docs/jev-local-operations.md`
  → `:1` or more, exit 0; `git grep -c -F "예외는 한 곳뿐이다" -- .moai/docs/jev-local-operations.md`
  → `:1`, exit 0 (the pinned sentence still present);
  `<SCRUB> go test -count=1 -v -run '^TestJevDoctrineAmendment$' ./internal/template/` →
  `--- PASS: TestJevDoctrineAmendment/local-guide `; the wording guard's `local-guide` →
  `--- PASS`; `git diff --stat <CARD_BASE> -- AGENTS.local.md` → empty.
- **Mutant probe.** Rewording the pinned paragraph ("한 곳" → "두 곳") fails
  `TestJevDoctrineAmendment/local-guide` ("lacks the design.md §11.1 amendment text").

## AC-JAE-007 — The confirmed extension surfaces (X2, X3, X4), live and mirror

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-004

- **Class**: release-blocking (RED-now: L6, L6b). Applies per confirmed surface; a surface
  the operator cuts is recorded in `progress.md` and its row is removed from this
  criterion and from the marker registry (plan.md D-3).
- **Given** the tree after M3, **When** the greps and the wording guard run, **Then**
  `agent-authoring.md:147`, `SKILL.md:180` and `CLAUDE.md:63` and their template mirrors
  carry both literals in the same sentence as the original wording; `CLAUDE.md` and its
  mirror are byte-identical and grew by at most 160 bytes; the rest of each file is
  unchanged.
- **Green** (M3):
  `git grep -c -F "auto-scoped ranking exception" -- .claude/rules/moai/development/agent-authoring.md .claude/skills/moai/SKILL.md CLAUDE.md`
  → three lines, each `:1`, exit 0, and the same for the three template paths;
  `wc -c CLAUDE.md internal/template/templates/CLAUDE.md` → two equal sizes ≤ `15733`;
  `diff -q CLAUDE.md internal/template/templates/CLAUDE.md` → no output, exit 0; the
  wording guard's `extension-rows` → `--- PASS`; if X3 survives,
  `git diff --stat <CARD_BASE> -- internal/template/catalog.yaml` shows the file changed
  and `make build` leaves no further diff.
- **Mutant probe.** Editing the live `SKILL.md` line only fails `mirror-parity`. A
  parenthetical that omits `selection order only` fails the bound-literal assertion.

## AC-JAE-008 — Mirrors agree and carry nothing internal

**Covers**: maps REQ-JAE-004

- **Class**: regression-guard (no RED-now; the mirrors are consistent today, which is
  the baseline). Green from arrival.
- **Given** the tree after M3, **When** the mirror, neutrality and parity checks run,
  **Then** every live/mirror pair carries the same amended passage and no template file
  carries a SPEC id, requirement token, ISO date or commit hash.
- **Green** (M3): `<SCRUB> go test -count=1 -v -run '^TestMCPToolCatalogueDocsStayMirrorIdentical$' ./internal/cli/`
  → `--- PASS`; `<SCRUB> go test -count=1 -v -run '^TestTemplateNoInternalContentLeak$' ./internal/template/`
  → `--- PASS`; the wording guard's `mirror-parity` → `--- PASS`.
- **Mutant probe.** Planting `SPEC-JEV-AUTO-EXCEPTION-001` in a template mirror fails
  `TestTemplateNoInternalContentLeak`; planting a differing clause in one copy fails
  `mirror-parity`.

## AC-JAE-009 — The guards that pin the first amendment still pass, unmodified

**Covers**: maps REQ-JAE-005

- **Class**: regression-guard (baseline PASS at `c50da9c2f`, research.md §R3).
- **Given** the tree after M3, **When** the pinned guards of `spec.md` §B.8 run,
  **Then** all PASS and none of their source files differs from `CARD_BASE`.
- **Green** (M3): the commands V3-V7 of `plan.md` §F — `TestJevDoctrineAmendment`
  (6 subtests PASS), `TestJevAmendmentLinkage` (6 subtests PASS),
  `TestMCPToolCatalogueDocsStayMirrorIdentical` and the five `--auto` doc guards (six
  `--- PASS`), `TestNoConsumerCallPathShips`, and the two `internal/jev` import tests;
  and `git diff --stat <CARD_BASE> -- internal/contract/kickoff/activation_test.go internal/template/contract_mode_blocks_test.go internal/contract/kickoff/kickoff.go AGENTS.local.md`
  → empty.
- **Mutant probe.** Removing `contract-mode Kickoff` from any `workflow.yaml` copy turns
  `TestJevAmendmentLinkage/tree` and `TestJevDoctrineAmendment/rules-and-config` red,
  which is why this criterion exists next to the edits.

## AC-JAE-010 — No behavior change

**Covers**: maps REQ-JAE-003, REQ-JAE-008

- **Class**: regression-guard (no RED-now; the mutant probe is a code change).
- **Given** the tree after M3, **When** the changed-line filter and the build run,
  **Then** every changed content line of `internal/jev/jev.go` and
  `internal/cli/mcp_jev.go` is a `//` comment line; the `mcp.WithDescription` line is
  unchanged; `todo_auto_rank.go`, `defaults.go` and `internal/mcp/catalog.go` have no
  diff; both builds exit 0.
- **Green** (M4): `git diff -U0 <CARD_BASE> -- internal/jev/jev.go internal/cli/mcp_jev.go`
  piped through `grep -E '^[+-]'`, then `grep -v -E '^(\+\+\+|---)'`, then
  `grep -v -E '^[+-][[:space:]]*//'` → no output (every changed content line is a `//`
  comment line). The filter was given a positive control at plan time: the same pipeline
  over `git diff -U0 7d8a9bdbc c50da9c2f -- internal/cli/todo_auto.go`, a diff that does
  change code, prints code lines (`+	landed  autoLandedLookup // …`), so an empty result
  is not a broken filter. Then `git grep -c -F "DISPLAY-ONLY: the answer is a labelled model signal a person reads" -- internal/cli/mcp_jev.go`
  → `internal/cli/mcp_jev.go:1` (the baseline L11);
  `git diff --stat <CARD_BASE> -- internal/cli/todo_auto_rank.go internal/config/defaults.go internal/mcp/catalog.go`
  → empty; `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` → exit 0;
  `<SCRUB> go test -count=1 -v -run '^(TestPackageImports_AreStandardLibraryOnly|TestImportClassifierPositiveControl)$' ./internal/jev/`
  → two `--- PASS`.
- **Mutant probe.** A one-token change to a non-comment line of `mcp_jev.go` (for example
  the description string) leaves the grep for L11's phrase at `:0` or the changed-line
  filter non-empty; either fails the criterion.

## AC-JAE-011 — The linkage guard bites on partial, split and dangling amendments

**Covers**: maps REQ-JAE-009

- **Class**: release-blocking (RED-now: L7, L12).
- **Given** M1 landed, **When** the guard runs, **Then** every `falsifier/*` subtest
  PASSes because the checker returned a finding for its fixture, `all-in-one-commit`
  PASSes because the checker returned none, and `tree` PASSes with the logged state
  consistent with `jevAutoExceptionAmended`.
- **Green** (M1; `tree` re-checked at M3):
  `<SCRUB> go test -count=1 -list '^(TestJevAutoExceptionLinkage|TestJevAutoExceptionWording)$' ./internal/template/`
  → exactly `TestJevAutoExceptionLinkage`, `TestJevAutoExceptionWording`, then `ok`
  (sweep control; RED-now L7 printed no name); then
  `<SCRUB> go test -count=1 -v -run '^TestJevAutoExceptionLinkage$' ./internal/template/`
  → `--- PASS: TestJevAutoExceptionLinkage ` with PASS lines for `falsifier/partial/*`
  (one per registry row), `falsifier/arming-only`, `falsifier/split-commits`,
  `falsifier/dangling-anchor`, `all-in-one-commit` and `tree`; after M3 the `tree` line
  logs `armed=true`; the `falsifier/*` logs contain `partial amendment`, `first appears
  in` and `dangling` respectively.
- **Mutant probe.** Delete the first-commit comparison from the checker: `falsifier/
  split-commits` turns red. Delete the anchor check: `falsifier/dangling-anchor` turns
  red. Make the checker return nothing when the arming constant is absent from the
  registry: `falsifier/arming-only` turns red. Each deletion is made once at M1 and the
  red output is kept (manager-develop §E8).

## AC-JAE-012 — The wording guard bites on a missing literal, a dropped target, a stray bound

**Covers**: maps REQ-JAE-010

- **Class**: release-blocking (RED-now: L7).
- **Given** M1 landed, **When** the guard runs, **Then** `falsifier/literal-missing`,
  `falsifier/closed-target-dropped` and `falsifier/bound-in-other-paragraph` each PASS
  because the checker rejected a planted passage, and the seven group subtests PASS
  after M3 (and skip with a logged reason before it).
- **Green** (M1, groups at M3):
  `<SCRUB> go test -count=1 -v -run '^TestJevAutoExceptionWording$' ./internal/template/`
  → `--- PASS: TestJevAutoExceptionWording ` with PASS lines for the three falsifiers,
  `mirror-parity`, and `go-comments`, `config-comment`, `catalogue-rows`, `spec-core`,
  `spec-manager-todo`, `local-guide`, `extension-rows`; a `--- SKIP` line for a group is
  acceptable only before M3 and never in the final run.
- **Mutant probe.** A checker that accepts any passage containing the literal anywhere in
  the file turns `falsifier/bound-in-other-paragraph` red. A checker that ignores closed
  targets turns `falsifier/closed-target-dropped` red.

## AC-JAE-013 — The guard lands first; the markers land together

**Covers**: maps REQ-JAE-011

- **Class**: release-blocking (RED-now: L7 — no guard exists, so no ancestor relation can
  hold).
- **Given** the run-phase commits, with the guard commit `G` and the linked commit `K`
  recorded in `progress.md` §E.2, **When** the commit graph is read, **Then** `G` is an
  ancestor of `K`; `G` contains `jevAutoExceptionAmended = false`; `K` is the first
  commit in which the arming constant reads `true` and the first commit in which
  `auto-scoped ranking exception` appears in any registry file other than the anchors;
  and `K` was staged by explicit pathspec.
- **Green** (M4, SHAs from `progress.md`): `git merge-base --is-ancestor <G> <K>` →
  exit 0; `git grep -c -F "jevAutoExceptionAmended = false" <G> -- internal/template/jev_auto_exception_test.go`
  → `…:1`; `git grep -c -F "jevAutoExceptionAmended = true" <K> -- internal/template/jev_auto_exception_test.go`
  → `…:1`; `<SCRUB> go test -count=1 -v -run '^TestJevAutoExceptionLinkage$' ./internal/template/`
  → `tree` PASS with `armed=true` (that subtest is the mechanical first-commit check over
  the whole registry); `git show --stat <K>` lists the M2 and M3 files and no file
  outside them.
- **Mutant probe.** Committing the guard and the markers together makes `G` equal `K`, so
  the ancestor check is trivially true: the criterion therefore also requires `G` ≠ `K`
  (`git rev-parse <G>` differs from `git rev-parse <K>`), which the single-commit mutant
  fails.

## AC-JAE-014 — The inventory is closed: no unamended class-(i) passage remains

**Covers**: maps REQ-JAE-001, REQ-JAE-012

- **Class**: release-blocking (RED-now: L1-L6b — every class-(i) passage lacks the literal).
- **Given** the tree after M3, **When** the two sweeps of `plan.md` §F V12 run, **Then**
  every hit's file appears in the `research.md` §R1 table with a class and a reason; each
  class-(i) hit's passage carries both literals; and a planted fixture hits the same
  command's output (positive control: a known unchanged hit, `internal/mcp/catalog.go:97`,
  is present in the primary sweep).
- **Green** (M4): `git grep -n -i -E "display-only|display only" -- . ":!.moai/specs" ":!.moai/reports" ":!CHANGELOG.md"`
  → the baseline of research.md §R1 (69 hits in 43 files) plus the added literals and no
  new file; `git grep -n -E "never reorder by inferred priority|판단 자료|모델 답을 입력으로도" -- . ":!.moai/specs" ":!.moai/reports" ":!CHANGELOG.md"`
  → each hit classified; for every class-(i) file
  `git grep -c -F "auto-scoped ranking exception" -- <file>` is ≥ 1 (L1-L6b flipped); and
  the control: the primary sweep's output contains the line
  `internal/mcp/catalog.go:97:` (a class-(iii) hit this SPEC does not touch), so a sweep
  that returns nothing cannot be read as a closed inventory.
- **Mutant probe.** Deleting the literal from any one class-(i) file turns that file's
  `git grep -c` back to exit 1 and, independently, turns `TestJevAutoExceptionLinkage/tree`
  red as a partial amendment.

## Sweep control (precondition for every green reading of the new tests)

Before reading any `PASS`, count what ran. After M1,
`<SCRUB> go test -count=1 -list '^(TestJevAutoExceptionLinkage|TestJevAutoExceptionWording)$' ./internal/template/`
must print exactly two test names and `ok`; RED-now (L7) printed none. Then the `-v` run
of both names, written to a file and counted with `grep -c "^--- PASS: "`, must report
`2` top-level PASS lines (subtest lines are indented and not counted) and
`grep -c "^--- FAIL"` must report `0`. A green with an empty swept set asserts nothing.

## Edge cases

- **An extension is cut.** The surface's rows leave the registry and this file's
  AC-JAE-007 row; the decision is recorded in `progress.md`. If X1 (the guide) is cut,
  AC-JAE-006 and the `local-guide` group go with it, and the first guard's guide marker
  is unaffected.
- **The arming constant flips without the markers, or the markers land without it.**
  Either is a partial amendment and fails `tree`.
- **The guard lands with `armed=false` and the markers are absent.** The group
  subtests skip with a logged reason, the `tree` subtest passes; this is the intended
  M1 state and the only one in which a skip is acceptable.
- **X3 survives.** `internal/template/catalog.yaml` changes in the linked commit; a
  catalog hash left unregenerated is caught by the catalog hash test, not by this SPEC's
  guard.
- **`jev.enabled` differs between live and template `workflow.yaml`** (live `true`,
  template `false`) by design; the mirror-parity check compares the exception block only.
- **A landed anchor loses the literal** (someone edits `kanban-dispatch.md`): the
  `dangling-anchor` check fails the tree subtest, which is its purpose.
- **The linked commit is reverted whole**: undetected (spec.md §G R-2).

## Quality gates

- The §F scoped tests all pass; `go vet ./internal/template/` exits 0.
- `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- `golangci-lint run` over the changed packages (`./internal/jev/...`, `./internal/cli/...`,
  `./internal/template/...`) reports no new finding against the CI version of the tool.
- The new test file uses named constants for the registry tokens and no inline
  environment-variable names; it calls no `AskUserQuestion`.
- Every template file passes the neutrality checks (AC-JAE-008).

## Definition of Done

1. AC-JAE-001..014 are PASS on the real output of the green commands above, and the
   sweep control's two names were executed.
2. AC-JAE-011 and AC-JAE-012 each kept a verbatim RED (the falsifier failing against its
   stub or its deleted check) before GREEN (manager-develop §E8).
3. None of the files named in AC-JAE-009's `git diff --stat` differs from `CARD_BASE`.
4. The guard commit `G` and the linked commit `K` are recorded in `progress.md` and
   satisfy AC-JAE-013; the linked commit was assembled `manager-spec` first, then
   `manager-develop`, staged by explicit pathspec.
5. The completion report states that this SPEC claims no Jev ordering accuracy
   (decision 6), that `REQ-JEVO-009` and the labelled accuracy set are out of scope with
   the reason, and which of X1-X4 the operator kept or cut.
6. `moai spec lint SPEC-JEV-CORE-001`, `SPEC-MANAGER-TODO-001` and
   `SPEC-JEV-AUTO-EXCEPTION-001` each report no findings with a build made from the tree
   under measurement, and the report names that build's tree HEAD next to each result.
7. The queue store and the live queue were not changed at any stage.
