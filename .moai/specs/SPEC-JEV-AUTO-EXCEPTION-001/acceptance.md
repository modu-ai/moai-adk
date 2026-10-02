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
> contains `&&`, so the one ledger row that carries it (L7) is informational: its
> stdout also carries a wall-clock figure that differs between runs, and **no
> criterion cites L7 as its RED-now**. `-run` patterns are anchored (`^…$`); an
> unanchored pattern also selects longer names, which is a vacuous-pass shape.
> Commands that need the card's base use `CARD_BASE` from `git merge-base develop HEAD`
> evaluated at run time, never a literal SHA, and only before the card merges.
>
> **Classification.** **release-blocking**: a conforming RED-now ledger row exists and
> reproduces at the pinned tree `c50da9c2f` (11 criteria: AC-JAE-001..007, 011..014).
> **regression-guard**: no RED-now, GREEN from arrival (3 criteria: AC-JAE-008, 009,
> 010), each with a mutant probe instead.
>
> **RED-now / green-path pairing.** Every release-blocking criterion pairs the
> ledger row that shows why it is red now with the milestone that flips it
> (`plan.md` §E). Surfaces whose RED is "the literal is absent" are red for the right
> stated reason: the exception is unwritten there, and only this SPEC's work can
> write it. The behavior-bearing criteria are red because the test does not exist:
> AC-JAE-011 (L12, the arming constant is absent), AC-JAE-012 (L16, the wording test
> name is absent) and AC-JAE-013 (L17, the guard file is absent, so no ancestor
> relation can hold) — each a single-invocation `git grep` with its own stdout and
> exit code. The failing output of each falsifier before its implementation is
> captured verbatim at M1 (manager-develop §E8), not in this ledger.
>
> **Swept-set controls.** Every ledger row whose RED-now is an empty stdout is paired
> with a control row that runs a probe known to hit on the same pathspec (rows C1-C5
> and L8, L9, L18), so an empty stdout is a measured absence on a non-empty swept set
> and not a broken probe (`verification-completeness.md` §1.1).

## Evidence ledger (RED-now observations, pinned tree `c50da9c2f`)

Whole-ledger pin: tree `c50da9c2f`, branch `WT-jev-auto-exception`. The ledger was
re-executed at the revision tree `1eef55dd9`; `git diff --name-only c50da9c2f 1eef55dd9`
lists only this SPEC's own five files, none under any pathspec below, so every row
reproduces unchanged (re-run table: `research.md` §R5). **After the run these rows
are history, by design:** each RED-now row describes the pinned tree and is expected
to print something else on a tree that contains the linked commit `K` — that flip is
the criterion's green path. The reconciliation re-ran the flip-pending rows at
`e70578c24` (`research.md` §R7). Row ids are `L<n>` (RED-now and
context rows) and `C<n>` (swept-set controls) so they do not collide with
manager-develop's self-verification items E1-E8.

| Id | Command | Verbatim stdout | Exit | Why it is red |
|---|---|---|---|---|
| L1 | `git grep -c -F "auto-scoped ranking exception" -- internal/jev/jev.go internal/cli/mcp_jev.go` | (empty) | 1 | neither Go comment names the exception |
| L2 | `git grep -c -F "auto-scoped ranking exception" -- .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml` | (empty) | 1 | the jev comment, live and mirror, has no second exception |
| L3 | `git grep -c -F "auto-scoped ranking exception" -- .claude/rules/moai/core/moai-mcp-tools-catalogue.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md` | (empty) | 1 | both catalogue copies carry only the Kickoff exception |
| L4 | `git grep -c -F "auto-scoped ranking exception" -- .moai/specs/SPEC-JEV-CORE-001/spec.md .moai/specs/SPEC-MANAGER-TODO-001/spec.md` | (empty) | 1 | neither completed SPEC states the exception |
| L5 | `git grep -c -F "auto-scoped ranking exception" -- .moai/docs/jev-local-operations.md` | (empty) | 1 | the doctrine guide says "one place only" |
| L6 | `git grep -c -F "auto-scoped ranking exception" -- .claude/rules/moai/development/agent-authoring.md .claude/skills/moai/SKILL.md CLAUDE.md` | (empty) | 1 | the three live extension surfaces lack it |
| L6b | `git grep -c -F "auto-scoped ranking exception" -- internal/template/templates/.claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/skills/moai/SKILL.md internal/template/templates/CLAUDE.md` | (empty) | 1 | and so do their template mirrors |
| L7 | `<SCRUB> go test -count=1 -list '^(TestJevAutoExceptionLinkage\|TestJevAutoExceptionWording)$' ./internal/template/` | `ok  	github.com/modu-ai/moai-adk/internal/template	<elapsed>s` (no test name listed; `<elapsed>` is a wall-clock figure that differs per run — 0.292s and 0.169s were observed) | 0 | zero test names: neither guard exists. **Informational only** — the command carries the scrub prefix (`&&`), so it is outside the single-invocation form, and no criterion cites it as RED-now |
| L8 | `git grep -c -F "auto-scoped ranking exception" -- .claude/rules/moai/workflow/kanban-dispatch.md .claude/skills/moai/workflows/gtd.md .claude/agents/moai/manager-todo.md` | `.claude/agents/moai/manager-todo.md:2` · `.claude/rules/moai/workflow/kanban-dispatch.md:1` · `.claude/skills/moai/workflows/gtd.md:1` | 0 | **positive control**: the landed anchors do carry the token, so the zero rows above are about the Jev side, not about a broken probe |
| L9 | `git grep -n -F "The capability is display-only" -- .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml` | `.moai/config/sections/workflow.yaml:227:    # The capability is display-only: an answer is a labelled model-produced` · `internal/template/templates/.moai/config/sections/workflow.yaml:229:    # The capability is display-only: an answer is a labelled model-produced` | 0 | the not-yet-amended side, live and mirror (interim split evidence) |
| L10 | `git grep -n -F "as a display-only signal for dispatch order and priority, except for" -- .claude/agents/moai/manager-todo.md` | `.claude/agents/moai/manager-todo.md:7:  Jev as a display-only signal for dispatch order and priority, except for` | 0 | the landed side states the exception (interim split evidence) |
| L11 | `git grep -c -F "DISPLAY-ONLY: the answer is a labelled model signal a person reads" -- internal/cli/mcp_jev.go` | `internal/cli/mcp_jev.go:1` | 0 | baseline for the unchanged description string (AC-JAE-010) |
| L12 | `git grep -c -F "jevAutoExceptionAmended" -- internal/template` | (empty) | 1 | the arming constant does not exist |
| L13 | `git grep -c -F "v0.4.0" -- .moai/specs/SPEC-JEV-CORE-001/spec.md` | (empty) | 1 | no v0.4.0 marker |
| L14 | `git grep -n -F "version: \"0.3.0\"" -- .moai/specs/SPEC-JEV-CORE-001/spec.md` | `.moai/specs/SPEC-JEV-CORE-001/spec.md:4:version: "0.3.0"` | 0 | version still 0.3.0 |
| L15 | `git grep -c -F "doc_display_only_test" -- internal/jev/jev.go` | `internal/jev/jev.go:1` | 0 | `jev.go:27` names a test file that does not exist (`display_only_test.go` does) |
| L16 | `git grep -c -F "TestJevAutoExceptionWording" -- internal/template` | (empty) | 1 | the wording guard does not exist yet (AC-JAE-012) |
| L17 | `git grep -c -F "jev_auto_exception_test" -- internal/template` | (empty) | 1 | the guard file is not named anywhere in the package, so no commit `G` can be an ancestor of any `K` (AC-JAE-013). **What flips it:** the guard names its own path — the file-header comment (`jev_auto_exception_test.go:1`) and the `jaeGuardFile` constant (`:41`) that the registry's arming row uses as its path (plan.md D-3) — so L17 turns non-empty because the guard exists *and* names itself, as the registry requires; a guard that did not name its own path would leave L17 empty, so L17 is a probe of that, with AC-JAE-013's `git grep -c -F "jevAutoExceptionAmended = false" <G> -- <guard file>` (it prints `<G>:…:1` only if the file exists at `G`) as the existence probe |
| L18 | `git grep -c -F "TestJevDoctrineAmendment" -- internal/template` | `internal/template/contract_mode_blocks_test.go:2` | 0 | **positive control** for L12, L16, L17: the same pathspec and the same kind of probe hit a test that does exist, so the empty rows are measured absences, not a broken probe |
| L19 | `git grep -c -F "auto-scoped ranking exception" -- .claude/skills/moai-ref-jev-question-design/SKILL.md internal/template/templates/.claude/skills/moai-ref-jev-question-design/SKILL.md` | (empty) | 1 | the reference skill pair (X5) states the "queue mutation" sentence with no exception |
| C1 | `git grep -c -F "display-only" -- internal/jev/jev.go internal/cli/mcp_jev.go` | `internal/cli/mcp_jev.go:1` · `internal/jev/jev.go:1` | 0 | control for L1: the probe's sibling literal is live on this pathspec |
| C2 | `git grep -c -F "display-only" -- .claude/rules/moai/core/moai-mcp-tools-catalogue.md internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md .moai/specs/SPEC-JEV-CORE-001/spec.md .moai/specs/SPEC-MANAGER-TODO-001/spec.md` | `.claude/rules/moai/core/moai-mcp-tools-catalogue.md:2` · `.moai/specs/SPEC-JEV-CORE-001/spec.md:4` · `.moai/specs/SPEC-MANAGER-TODO-001/spec.md:1` · `internal/template/templates/.claude/rules/moai/core/moai-mcp-tools-catalogue.md:2` | 0 | control for L3 and L4 (L2 is controlled by L9) |
| C3 | `git grep -c -F "한 곳뿐" -- .moai/docs/jev-local-operations.md` | `.moai/docs/jev-local-operations.md:1` | 0 | control for L5 |
| C4 | `git grep -c -e "display-only" -e "never reorder by inferred priority" -- .claude/rules/moai/development/agent-authoring.md .claude/skills/moai/SKILL.md CLAUDE.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/skills/moai/SKILL.md internal/template/templates/CLAUDE.md` | `.claude/rules/moai/development/agent-authoring.md:1` · `.claude/skills/moai/SKILL.md:1` · `CLAUDE.md:1` · `internal/template/templates/.claude/rules/moai/development/agent-authoring.md:1` · `internal/template/templates/.claude/skills/moai/SKILL.md:1` · `internal/template/templates/CLAUDE.md:1` | 0 | control for L6 and L6b |
| C5 | `git grep -c -F "a labelled model signal a person reads" -- .claude/skills/moai-ref-jev-question-design/SKILL.md internal/template/templates/.claude/skills/moai-ref-jev-question-design/SKILL.md` | `.claude/skills/moai-ref-jev-question-design/SKILL.md:1` · `internal/template/templates/.claude/skills/moai-ref-jev-question-design/SKILL.md:1` | 0 | control for L19 |

## AC-JAE-001 — The two Go comments name the exception and stay comments

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-003, REQ-JAE-008

- **Class**: release-blocking (RED-now: L1, L15; control C1).
- **Given** the tree after M3, **When** the two files are read with `git grep` and the
  wording guard runs, **Then** `internal/jev/jev.go` and `internal/cli/mcp_jev.go` each
  carry `auto-scoped ranking exception`, `selection order only`, `workflow.jev.enabled`
  and `mechanical filters` in the comment block that says "display-only" (long-form),
  contain no claim phrasing and no over-reach phrasing; `mcp_jev.go` also says the
  exception does not reach "this tool"; `jev.go` no longer names
  `doc_display_only_test.go`; neither file contains `NearDuplicateMark`,
  `LaneQuestionRoute` or `SkillSuggest`.
- **Green** (M3):
  `git grep -c -F "auto-scoped ranking exception" -- internal/jev/jev.go internal/cli/mcp_jev.go`
  → `…jev.go:` and `…mcp_jev.go:` each ≥ 1, exit 0; the same with `"selection order only"`
  and with `"mechanical filters"`;
  `git grep -c -F "doc_display_only_test" -- internal/jev/jev.go` → empty, exit 1;
  `git grep -c -E "NearDuplicateMark|LaneQuestionRoute|SkillSuggest" -- internal/jev/jev.go internal/cli/mcp_jev.go`
  → empty, exit 1; and
  `<SCRUB> go test -count=1 -v -run '^TestJevAutoExceptionWording$' ./internal/template/`
  → `--- PASS: TestJevAutoExceptionWording/go-comments `.
- **Mutant probe.** A comment that carries both literals in a *different* paragraph from
  "display-only" passes the greps and fails the same-passage assertion. A comment that
  also drops "a queue mutation" from `mcp_jev.go` fails the closed-target assertion. A
  comment with both literals and no gate clause (`workflow.jev.enabled`) fails the
  required-literal assertion (`falsifier/long-form-literal-missing`'s shape). A comment
  ending "…selection order only, and the ordering is accurate" fails the claim-phrasing
  assertion.

## AC-JAE-002 — The `workflow.yaml` jev comment, live and mirror

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-004, REQ-JAE-005

- **Class**: release-blocking (RED-now: L2; context: L9).
- **Given** the tree after M3, **When** the grep and both guards run, **Then** both
  `workflow.yaml` copies carry the two literals, `workflow.jev.enabled` and
  `mechanical filters` inside the `# jev: ` comment run (long-form); the run
  still names `contract-mode Kickoff` exactly once, `llm+jev`, `never decides alone`,
  `completion verdict`, `merge` and `queue mutation`; it contains no claim or over-reach
  phrasing; and the added block is identical in the two copies.
- **Green** (M3): `git grep -c -F "auto-scoped ranking exception" -- .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml`
  → one line per file, each ≥ 1, exit 0; the same with `"mechanical filters"` → one line
  per file, each ≥ 1; `<SCRUB> go test -count=1 -v -run '^TestJevDoctrineAmendment$' ./internal/template/`
  → `--- PASS: TestJevDoctrineAmendment/rules-and-config `; and
  `<SCRUB> go test -count=1 -v -run '^TestJevAutoExceptionWording$' ./internal/template/`
  → `--- PASS` for `config-comment` and `mirror-parity`.
- **Mutant probe.** Repeating `contract-mode Kickoff` in the new block fails the
  existing "exactly once" assertion. Editing only the live copy fails `mirror-parity`.
  A new block without the gate or filter literal fails the `config-comment` required
  literals.

## AC-JAE-003 — The catalogue rows, both copies, byte-identical

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-003, REQ-JAE-004, REQ-JAE-005

- **Class**: release-blocking (RED-now: L3; control C2).
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

- **Class**: release-blocking (RED-now: L4, L13, L14; control C2).
- **Given** the tree after M2-M3, **When** the greps, the guards and the linter run,
  **Then** `spec.md` carries `version: "0.4.0"`, a v0.4.0 marker and an exception paragraph
  on REQ-JEVC-011 and REQ-JEVC-012 (both containing the two literals,
  `workflow.jev.enabled` and `mechanical filters`; the paragraph is long-form), both
  `Out of Scope — authority` bullets still two in number and each naming `contract-mode
  Kickoff`, `llm+jev` and the new exception, a HISTORY row for v0.4.0, `status: completed`,
  and its `progress.md` unchanged.
- **Green** (M2, verified at M3):
  `git grep -n -F "version: \"0.4.0\"" -- .moai/specs/SPEC-JEV-CORE-001/spec.md` → line `4`,
  exit 0; `git grep -c -F "v0.4.0" -- .moai/specs/SPEC-JEV-CORE-001/spec.md` → ≥ 3 (a marker
  line on each of the two requirements plus the HISTORY row), exit 0;
  `git grep -c -F "mechanical filters" -- .moai/specs/SPEC-JEV-CORE-001/spec.md` → ≥ 2
  (one in each requirement's exception paragraph), exit 0;
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

- **Class**: release-blocking (RED-now: L4; control C2).
- **Given** the tree after M2-M3, **When** the grep, the wording guard and the linter
  run, **Then** REQ-MT-014 and REQ-MT-015 each state that they govern the local Jev
  scripts (display-only unchanged) and that the `--auto` cycle's own ranking is the one
  `--auto` exception — the `auto-scoped ranking exception` with
  `selection order only` (short-form scope sentences); REQ-MT-015 still contains
  `queue mutation`; the HISTORY has a new row; `status: completed`; no requirement id
  removed or renumbered.
- **Green** (M2, verified at M3):
  `git grep -c -F "auto-scoped ranking exception" -- .moai/specs/SPEC-MANAGER-TODO-001/spec.md`
  → ≥ 2, exit 0; the wording guard's `spec-manager-todo` → `--- PASS`;
  `git grep -c -E "^- \*\*REQ-MT-0(14|15)\*\*" -- .moai/specs/SPEC-MANAGER-TODO-001/spec.md`
  → `2`; `<scratchpad>/moai spec lint SPEC-MANAGER-TODO-001` → `✓ No findings — all SPEC
  documents are valid`; and the version moved with the HISTORY row, 0.1.0 → 0.2.0:
  `git grep -n -F 'version: "0.1.0"' c50da9c2f -- .moai/specs/SPEC-MANAGER-TODO-001/spec.md`
  printed `c50da9c2f:.moai/specs/SPEC-MANAGER-TODO-001/spec.md:4:version: "0.1.0"` and
  `git grep -n -F 'version: "0.2.0"' -- .moai/specs/SPEC-MANAGER-TODO-001/spec.md` prints
  `.moai/specs/SPEC-MANAGER-TODO-001/spec.md:4:version: "0.2.0"` (exit 0; both
  measured at `e70578c24`). The plan (D-6) always expected 0.2.0; this criterion's
  original text asserted the HISTORY row and not the number, and no guard pins this
  SPEC's version (`progress.md` §E.2 finding 7).
- **Mutant probe.** A scope sentence on REQ-MT-014 only fails the per-requirement
  assertion on REQ-MT-015. A rewrite of REQ-MT-015 that deletes `queue mutation` fails
  the closed-target assertion.

## AC-JAE-006 — The doctrine guide is amended additively

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-005

- **Class**: release-blocking (RED-now: L5; control C3).
- **Given** the tree after M3, **When** the grep and the guards run, **Then**
  `.moai/docs/jev-local-operations.md` carries a new paragraph with the two literals,
  `workflow.jev.enabled` and the Korean filter literal `기계적 필터` (long-form) that
  scopes the earlier "한 곳뿐" sentence to the Kickoff cross-check and keeps the other
  grade-3 prohibitions, while the pinned paragraph is byte-for-byte unchanged and
  `AGENTS.local.md` is untouched.
- **Green** (M3): `git grep -c -F "auto-scoped ranking exception" -- .moai/docs/jev-local-operations.md`
  → `:1` or more, exit 0; `git grep -c -F "기계적 필터" -- .moai/docs/jev-local-operations.md`
  → `:1` or more, exit 0; `git grep -c -F "예외는 한 곳뿐이다" -- .moai/docs/jev-local-operations.md`
  → `:1`, exit 0 (the pinned sentence still present);
  `<SCRUB> go test -count=1 -v -run '^TestJevDoctrineAmendment$' ./internal/template/` →
  `--- PASS: TestJevDoctrineAmendment/local-guide `; the wording guard's `local-guide` →
  `--- PASS`; `git diff --stat <CARD_BASE> -- AGENTS.local.md` → empty.
- **Mutant probe.** Rewording the pinned paragraph ("한 곳" → "두 곳") fails
  `TestJevDoctrineAmendment/local-guide` ("lacks the design.md §11.1 amendment text").

## AC-JAE-007 — The confirmed extension surfaces (X2, X3, X4, X5), live and mirror

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-004, REQ-JAE-005

- **Class**: release-blocking (RED-now: L6, L6b, L19; controls C4, C5). Applies per
  confirmed surface; a surface the operator cuts is recorded in `progress.md` and its
  row is removed from this criterion and from the marker registry (plan.md D-3).
- **Given** the tree after M3, **When** the greps and the wording guard run, **Then**
  `agent-authoring.md:147`, `SKILL.md:180`, `CLAUDE.md:63` and the reference-skill
  sentence (X5) and their template mirrors carry both literals in the same sentence as
  the original wording (short-form); `CLAUDE.md` and its mirror are byte-identical and
  grew by at most 160 bytes; the two reference-skill copies are byte-identical and
  carry none of `internal/jev`, `mcp__moai__jev`, `jev_ask`, `moai jev`; the rest of
  each file is unchanged.
- **Green** (M3):
  `git grep -c -F "auto-scoped ranking exception" -- .claude/rules/moai/development/agent-authoring.md .claude/skills/moai/SKILL.md CLAUDE.md`
  → three lines, each `:1`, exit 0, and the same for the three template paths;
  `git grep -c -F "auto-scoped ranking exception" -- .claude/skills/moai-ref-jev-question-design/SKILL.md internal/template/templates/.claude/skills/moai-ref-jev-question-design/SKILL.md`
  → two lines, each `:1`, exit 0 (flipped from L19);
  `wc -c CLAUDE.md internal/template/templates/CLAUDE.md` → two equal sizes ≤ `15733`;
  `diff -q CLAUDE.md internal/template/templates/CLAUDE.md` → no output, exit 0;
  `diff -q .claude/skills/moai-ref-jev-question-design/SKILL.md internal/template/templates/.claude/skills/moai-ref-jev-question-design/SKILL.md`
  → no output, exit 0; the wording guard's `extension-rows` → `--- PASS`; V13 (the two
  reference-skill guards, two `--- PASS`) and V14 (`TestCatalogHashParity`, `0 drift`);
  if X3 or X5 survives, `git diff --stat <CARD_BASE> -- internal/template/catalog.yaml`
  shows the file changed and `make build` leaves no further diff.
- **Mutant probe.** Editing the live `SKILL.md` line only fails `mirror-parity`. A
  parenthetical that omits `selection order only` fails the bound-literal assertion.
  Naming the wrapper tool in the reference-skill sentence fails
  `TestJevQuestionDesignSkillCarriesNoCallPath`; editing one reference-skill copy only
  fails `TestJevQuestionDesignSkillCopiesStayIdentical`; editing both and not
  regenerating `catalog.yaml` fails `TestCatalogHashParity`.

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
- **Green** (M3): the commands V3-V7, V13 and V14 of `plan.md` §F —
  `TestJevDoctrineAmendment` (6 subtests PASS), `TestJevAmendmentLinkage` (6 subtests
  PASS), `TestMCPToolCatalogueDocsStayMirrorIdentical` and the five `--auto` doc guards
  (six `--- PASS`), `TestNoConsumerCallPathShips`, the two `internal/jev` import tests,
  the two reference-skill guards and `TestCatalogHashParity`;
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

- **Class**: release-blocking (RED-now: L12, with the positive control L18; L7 is
  informational context only).
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
  (one per registry row), `falsifier/arming-only`, `falsifier/self-match`,
  `falsifier/split-commits`, `falsifier/dangling-anchor`, `all-in-one-commit` and
  `tree`; after M3 the `tree` line logs `armed=true`; the `falsifier/*` logs contain
  `partial amendment`, `first appears in` and `dangling` respectively; and
  `git grep -c -F "jevAutoExceptionAmended = true" <G> -- internal/template/jev_auto_exception_test.go`,
  with `<G>` the guard's own commit (`6d012fd4d`), prints nothing, exit 1 (the token is
  built by parts, so it is absent from the file before the amendment — D1 control,
  measured at `e70578c24`: no output, `exit=1`; the working-tree form without `<G>`
  would only describe the tree it is run in, which after `K` holds the token once;
  the post-amendment value is AC-JAE-013's).
- **Mutant probe.** Delete the comparison of marker commits from the checker: `falsifier/
  split-commits` turns red. Delete the anchor check: `falsifier/dangling-anchor` turns
  red. Make the checker return nothing when the arming constant is absent from the
  registry: `falsifier/arming-only` turns red. Write the registry row as a literal
  `jevAutoExceptionAmended = true` inside the test file: the guard's own `tree` subtest
  turns red at `armed=false` (a `partial amendment` finding) and `falsifier/self-match`
  names the defect — the mutant the plan-audit constructed. Each deletion is made once
  at M1 and the red output is kept (manager-develop §E8).

## AC-JAE-012 — The wording guard bites on a missing literal, a dropped target, a stray bound, a missing long-form literal, a claim, an over-reach

**Covers**: maps REQ-JAE-001, REQ-JAE-002, REQ-JAE-010

- **Class**: release-blocking (RED-now: L16, with the positive control L18; L7 is
  informational context only).
- **Given** M1 landed, **When** the guard runs, **Then** `falsifier/literal-missing`,
  `falsifier/closed-target-dropped`, `falsifier/bound-in-other-paragraph`,
  `falsifier/long-form-literal-missing`, `falsifier/claim-phrasing` and
  `falsifier/over-reach-phrasing` each PASS because the checker rejected a planted
  passage; `disclaimer-not-flagged` PASSes because the checker accepted a passage
  carrying `no ordering accuracy is claimed` and `is not claimed to be accurate`; and
  the seven group subtests PASS after M3 (and skip with a logged reason before it).
- **Green** (M1, groups at M3):
  `<SCRUB> go test -count=1 -v -run '^TestJevAutoExceptionWording$' ./internal/template/`
  → `--- PASS: TestJevAutoExceptionWording ` with PASS lines for the six falsifiers,
  `disclaimer-not-flagged`, `mirror-parity`, and `go-comments`, `config-comment`,
  `catalogue-rows`, `spec-core`, `spec-manager-todo`, `local-guide`, `extension-rows`; a
  `--- SKIP` line for a group is acceptable only before M3 and never in the final run.
  The planted phrases are those of `plan.md` §E (flagged: `…and the ordering is
  accurate`, `…the ordering beats the fallback`, `…with a measured accuracy of 80%`,
  `…applies to every moai todo pick`; accepted: `no ordering accuracy is claimed`, `and
  is not claimed to be accurate`, `any other decision that is hard to undo`).
- **Mutant probe.** A checker that accepts any passage containing the literal anywhere in
  the file turns `falsifier/bound-in-other-paragraph` red. A checker that ignores closed
  targets turns `falsifier/closed-target-dropped` red. A checker that skips the
  long-form literals turns `falsifier/long-form-literal-missing` red. A checker with the
  claim and over-reach sets removed turns `falsifier/claim-phrasing` and
  `falsifier/over-reach-phrasing` red; one whose claim set is so broad that it matches
  the word `accurate` anywhere turns `disclaimer-not-flagged` red — the paired rule
  that catches the nothing-matching direction (`verification-completeness.md` §2).

## AC-JAE-013 — The guard lands first; the markers land together

**Covers**: maps REQ-JAE-011

- **Class**: release-blocking (RED-now: L17 — the guard file is named nowhere in the
  package, so no ancestor relation can hold; positive control L18).
- **Given** the run-phase commits, with the guard commit `G` and the linked commit `K`
  recorded in `progress.md` §E.2, **When** the commit graph is read, **Then** `G` is an
  ancestor of, and different from, `K`; the text `jevAutoExceptionAmended = true` is
  **absent** from the guard file at `G` and occurs exactly once at `K`; `G` contains
  `jevAutoExceptionAmended = false`; `K` is the only commit that adds the arming text
  and the first commit in which `auto-scoped ranking exception` appears in any registry
  file other than the anchors.
- **Green** (M4, SHAs from `progress.md`): `git merge-base --is-ancestor <G> <K>` →
  exit 0; `git rev-parse <G>` and `git rev-parse <K>` print two different SHAs;
  `git grep -c -F "jevAutoExceptionAmended = false" <G> -- internal/template/jev_auto_exception_test.go`
  → `<G>:internal/template/jev_auto_exception_test.go:1` (the `= false` text is the
  const declaration alone — plan.md D-5 constrains this spelling the way it constrains
  `= true`, so a comment or fixture repeating it would make the count 2 and is a
  defect, not a count to tolerate; measured with `<G>` = `6d012fd4d`: `…:1`, and at
  `e70578c24` the text is absent from the file, exit 1);
  `git grep -c -F "jevAutoExceptionAmended = true" <G> -- internal/template/jev_auto_exception_test.go`
  → empty, exit 1 (**D1 control: the token is absent at `G`**);
  `git grep -c -F "jevAutoExceptionAmended = true" <K> -- internal/template/jev_auto_exception_test.go`
  → `<K>:internal/template/jev_auto_exception_test.go:1`;
  `git log --reverse --format=%H -S"jevAutoExceptionAmended = true" -- internal/template/jev_auto_exception_test.go`
  → exactly one line, `<K>`; `<SCRUB> go test -count=1 -v -run '^TestJevAutoExceptionLinkage$' ./internal/template/`
  → `tree` PASS with `armed=true` (that subtest is the mechanical one-commit check over
  the whole registry); `git show --stat <K>` lists the files of M2 and M3 and no file
  outside them.
- **Mutant probe.** Committing the guard and the markers together makes `G` equal `K`, so
  the ancestor check is trivially true: the criterion therefore also requires `G` ≠ `K`,
  which the single-commit mutant fails. A registry row spelled as the literal in the
  test file makes the `<G>` token control print `…:1` instead of staying empty, and the
  `-S` log list two commits; either fails the criterion.

## AC-JAE-014 — No unamended class-(i) passage remains over the three stated sweeps

**Covers**: maps REQ-JAE-001, REQ-JAE-012

- **Class**: release-blocking (RED-now: L1-L6b and L19 — every class-(i) passage lacks
  the literal; controls C1-C5, L8, L9). The criterion claims closure **over the three
  stated patterns only** (spec.md §G R-8): the first two patterns missed X5, which the
  third found, so a fourth phrasing could still hide a passage.
- **Given** the tree after M3, **When** the three sweeps of `plan.md` §F V12 run,
  **Then** every hit's file appears in a `research.md` §R1 table with a class and a
  reason (the guard source as class iii); each class-(i) hit's passage carries both
  literals; each sweep's control hit is present in its output; and the hits and files
  beyond the baselines are exactly the classified delta below.
- **Green** (M4): each sweep prints its baseline plus **exactly the classified
  delta** — not "no new file": the guard source
  `internal/template/jev_auto_exception_test.go`, a new file, enters all three sweeps
  and is class (iii) (locators, closed-target tables and fixtures of the wording
  guard; `research.md` §R1.6). Primary —
  `git grep -n -i -E "display-only|display only" -- . ":!.moai/specs" ":!.moai/reports" ":!CHANGELOG.md"`
  → **76 hits in 44 files** = the baseline of research.md §R1 (69 hits in 43 files)
  + 1 hit in `internal/cli/mcp_jev.go` (the added line of the amended comment block,
  class i) + 6 hits in the guard file (class iii); the control line
  `internal/mcp/catalog.go:97:` present; synonym —
  `git grep -n -E "never reorder by inferred priority|판단 자료|모델 답을 입력으로도" -- . ":!.moai/specs" ":!.moai/reports" ":!CHANGELOG.md"`
  → **6 hits** = the baseline 5 + 1 in the guard file (`jev_auto_exception_test.go:188`,
  the locator `never reorder by inferred priority`), every hit classified; closed-target
  phrases —
  `git grep -n -i -E "a person reads|labelled model|queue mutation|hard to undo" -- . ":!.moai/specs" ":!.moai/reports" ":!CHANGELOG.md"`
  → **96 hits in 48 files** = the baseline of research.md §R1.5 (84 hits in 47 files)
  + 1 hit in `internal/jev/jev.go` (the added comment, class i) + 11 hits in the guard
  file (class iii), each hit's file classified, and the control line
  `internal/cli/todo_triage.go:14:` (class iii, untouched) present. The file sets
  differ from the baselines by that one file in each of the two file-listing sweeps
  (`git grep -l` at `c50da9c2f` against the tree: `diff` prints one added line,
  `internal/template/jev_auto_exception_test.go`, for each — measured at
  `e70578c24`); any other file entering a sweep is unclassified and fails the
  criterion. For every class-(i) file
  `git grep -c -F "auto-scoped ranking exception" -- <file>` is ≥ 1 (L1-L6b and L19
  flipped), the class-(i) list including the reference-skill pair (X5).
- **Mutant probe.** Deleting the literal from any one class-(i) file turns that file's
  `git grep -c` back to exit 1 and, independently, turns `TestJevAutoExceptionLinkage/tree`
  red as a partial amendment. The mutant "every surface amended except the reference-skill
  pair" is caught by the third sweep's class-(i) file list (the pair is in it) and by the
  `git grep -c` over X5; before the third sweep existed it passed this criterion. The
  relaxed clause stays falsifiable in both directions: a class-(i) passage left
  unamended fails the per-file `git grep -c` (exit 1), and a file entering a sweep
  beyond the guard source — a restatement added elsewhere — appears as a second line in
  the file-set `diff` and in the hit totals (76/44, 6, 96/48), so it fails as
  unclassified. A tolerance such as "plus any test files" would admit both and is not
  what the criterion says.

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
- **X3 or X5 survives.** `internal/template/catalog.yaml` changes in the linked commit;
  a catalog hash left unregenerated is caught by `TestCatalogHashParity` (V14), not by
  this SPEC's guard.
- **`jev.enabled` differs between live and template `workflow.yaml`** (live `true`,
  template `false`) by design; the mirror-parity check compares the exception block only.
- **A landed anchor loses the literal** (someone edits `kanban-dispatch.md`): the
  `dangling-anchor` check fails the tree subtest, which is its purpose.
- **The linked commit is reverted whole**: undetected (spec.md §G R-2).

## Quality gates

- The §F scoped tests (V1-V9, V13, V14) all pass; `go vet ./internal/template/` exits 0.
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
   satisfy AC-JAE-013. **Process attestation** (not observable in the commit graph, so
   not a criterion — REQ-JAE-011 was trimmed to the graph-checkable part): the two
   completed-SPEC edits were written by `manager-spec` and every other surface by
   `manager-develop`, one writer at a time, and `K` was staged by explicit pathspec.
   **The commit graph does not witness this for `K`.** `K` (`7983d9131`) carries no
   `Authored-By-Agent` trailer (`git log -1 --format=%b 7983d9131` ends at `🗿 MoAI`,
   measured at `e70578c24`); `G` (`6d012fd4d`) carries `Authored-By-Agent:
   manager-develop`. `K` is a joint commit — the `manager-spec` bodies of M2 left
   uncommitted and the `manager-develop` surfaces of M3 added to them, then staged
   once by the lane orchestrator — so a single-agent trailer would be wrong, and the
   split rests on the lane's attestation alone (`progress.md` §E.2, the M2-M3
   paragraph, and the Gaps entry that names item 4 as unreadable from the graph). An
   earlier draft of this item named the trailers as "the nearest readable witness";
   for `K` they are not, and the SPEC does not say what `K`'s message should carry.
5. The completion report states that this SPEC claims no Jev ordering accuracy
   (decision 6), that `REQ-JEVO-009` and the labelled accuracy set are out of scope with
   the reason, and which of X1-X5 the operator kept or cut.
6. `moai spec lint SPEC-JEV-CORE-001`, `SPEC-MANAGER-TODO-001` and
   `SPEC-JEV-AUTO-EXCEPTION-001` each report no findings with a build made from the tree
   under measurement, and the report names that build's tree HEAD next to each result.
   The `moai` binary carries no commit stamp (`moai version` prints
   `v3.1.3 none built unknown` when built without ldflags), so the build's provenance
   is "built from tree HEAD `<sha>` in the same session, by construction", stated beside
   the result, and not something the binary reports about itself
   (`verification-claim-integrity.md` §2.2).
7. The queue store and the live queue were not changed at any stage.
