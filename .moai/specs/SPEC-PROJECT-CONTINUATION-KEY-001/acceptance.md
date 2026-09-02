# SPEC-PROJECT-CONTINUATION-KEY-001 — Acceptance Criteria

Card **t191** · 13 criteria · every criterion is binary and names the command that decides it.
Baseline tree `2660bcd09`.

---

## §D AC Matrix

| AC | REQ | Milestone | Severity |
|---|---|---|---|
| AC-PCK-001 | REQ-PCK-001, 002 | M2 | blocking |
| AC-PCK-002 | REQ-PCK-002 | M2 | blocking |
| AC-PCK-003 | REQ-PCK-003 | M2 | blocking |
| AC-PCK-004 | REQ-PCK-004 | M1 | blocking |
| AC-PCK-005 | REQ-PCK-005 | M1 | blocking |
| AC-PCK-006 | REQ-PCK-006 | M1 | blocking |
| AC-PCK-007 | REQ-PCK-007 | M1 | blocking |
| AC-PCK-008 | REQ-PCK-008 | M1 | blocking |
| AC-PCK-009 | REQ-PCK-009 | M3 | blocking |
| AC-PCK-010 | REQ-PCK-010 | M4 | blocking |
| AC-PCK-011 | REQ-PCK-011 | M5 | blocking |
| AC-PCK-012 | REQ-PCK-009, D6 | M6 | blocking |
| AC-PCK-013 | REQ-PCK-003, 004 | M1 | non-blocking |

---

### AC-PCK-001 — the closed set is exactly three tokens

**Given** the config package on the implementation branch,
**When** `go test ./internal/config/ -run TestValidProjectContinuations` runs a case asserting `config.ValidProjectContinuations()`,
**Then** the returned slice is exactly `["none", "card", "pipeline"]` and the test passes.

### AC-PCK-002 — absent resolves to card

**Given** a `workflow.yaml` carrying no `project:` block,
**When** the resolver is called on the loaded config,
**Then** it returns value `card` and an empty unmatched string, and the same holds for a nil `*Config` receiver.

### AC-PCK-003 — an unmatched value resolves to card and is reported

**Given** a `workflow.yaml` carrying `workflow.project.continuation: pipelien`,
**When** the resolver is called,
**Then** it returns value `card` **and** unmatched `pipelien`; the call does not error and does not panic.

### AC-PCK-004 — `none` issues no card and shows the pre-P1 recommended option

**Given** `doc-generation.md` on the implementation branch,
**When** `grep -n "none" -A 6` is read over the Step 4.1.5 resolution block and the Step 4.2 recommended-option table,
**Then** the `none` row instructs skipping issuance entirely and names `Create SPEC` — the pre-P1 wording measured from `e91def4ca` — as the recommended option, and no `moai todo add` invocation is reachable on that path.

### AC-PCK-005 — `card` reproduces P1

**Given** the same file,
**When** the `card` row of the Step 4.2 table is read,
**Then** it names `Create the SPEC and start now` as the recommended option, and Step 4.1.5's five standing-source steps (harness-spec read, empty-goal skip, duplicate-suppression read, single `[PROJECT] ` add, id reported) are present and unmodified from their P1 text.

### AC-PCK-006 — `pipeline` changes wording only

**Given** the same file,
**When** the `pipeline` row is read,
**Then** it names a recommended option whose text references continuation through run-phase implementation and tests, and the row contains no instruction to select, answer, skip, or default that option.

### AC-PCK-007 — the question survives every value

**Given** the same file,
**When** `grep -n "No branch is taken on the operator's behalf" doc-generation.md` and the new value-independence sentence are read,
**Then** both clauses are present, neither is scoped to a subset of values, and the new sentence states that the key changes only which option is recommended and how it is worded.

### AC-PCK-008 — the kickoff gate is untouched

**Given** the implementation branch,
**When** `git diff --stat origin/develop...HEAD -- .claude/rules/moai/workflow/orchestration-mode-selection.md .claude/rules/moai/workflow/goal-directive.md .claude/skills/moai/workflows/goal.md` runs,
**Then** the output is empty (zero files changed), and `grep -c "Implementation Kickoff Approval" doc-generation.md` returns a count no lower than its pre-change value.

### AC-PCK-009 — the shipped key passes the anti-rot guard

**Given** the template `workflow.yaml` carrying `continuation: card` and the inventory carrying the new row,
**When** `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders` runs,
**Then** the test passes, and the same test fails when the inventory row alone is removed (the mutant establishes the guard is not vacuous here).

### AC-PCK-010 — the wizard offers the closed set in four locales

**Given** the wizard package on the implementation branch,
**When** `go test ./internal/cli/wizard/...` runs a case that walks the `project_continuation` question,
**Then** its `Type` is `QuestionTypeSelect`, its `Default` is `card`, its option values equal `config.ValidProjectContinuations()`, and `GetLocalizedQuestion` returns a non-empty, mutually distinct Title for each of `en` / `ko` / `ja` / `zh`.

### AC-PCK-011 — the console field is closed-set and fully localized

**Given** the settings schema on the implementation branch,
**When** `go test ./internal/settings/... ./internal/web/...` runs,
**Then** a `FieldDef` exists for path `workflow → project → continuation` whose option values derive from `config.ValidProjectContinuations()`, and `internal/web/assets/i18n.js` carries `f.workflow.project.continuation.title`, `.desc`, and one `.opt.<value>` entry per token in each of the four locale maps — five keys × four locales = 20 entries, none empty.

### AC-PCK-012 — mirror parity holds after `make build`

**Given** a completed `make build`,
**When** `cmp` is run over the three byte-identical pairs (`doc-generation.md`, `todo.md`, `tab_schema.json`) and over `workflow.yaml`,
**Then** the three return rc=0 and `workflow.yaml` returns rc=1, with `diff` showing only the intended repo-local content (`branch_guard`, `agent_stop_guard`, populated `audit` pins, `context_folding`) plus the new key on both sides.

### AC-PCK-013 — the unmatched value reaches the operator (non-blocking)

**Given** the Step 4.2 report contract in `doc-generation.md`,
**When** the report template is read,
**Then** it carries a line that prints the offending value together with the canonical domain when the resolver returned a non-empty unmatched string, and prints nothing when it did not.

---

## §D.1 Severity

Twelve blocking, one non-blocking. AC-PCK-013 is non-blocking because it is a report-wording obligation whose absence degrades diagnosis without changing behaviour; AC-PCK-003 already guarantees the resolution itself.

## §D.2 Traceability

Every REQ-PCK-001..011 appears in at least one AC row of §D. REQ-PCK-007 and REQ-PCK-008 are prohibitions and are verified negatively — AC-PCK-007 by clause presence and scope, AC-PCK-008 by an empty diff over the three gate-owning files.

## §D.3 Indirect verification

REQ-PCK-004/005/006 govern orchestrator prose, which has no runtime harness. They are verified by reading the shipped instruction text, which is the artifact the orchestrator actually consumes. This is indirect and is recorded as such.

## §D.4 Closure gates

- All twelve blocking ACs pass.
- `go test ./internal/config/... ./internal/settings/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/web/...` is green, with each package's output cited.
- `golangci-lint run` on the touched packages is clean.
- `git status --short` shows no unintended file.

## §D.5 Forward-looking checks

- If M4 resolves §B item 1 by extending `GetLocalizedQuestion` with option translation, AC-PCK-010 gains a case asserting per-locale option descriptions; if it resolves by folding descriptions into the question body, AC-PCK-010 stands as written.
- If a later SPEC adds a fourth continuation token, AC-PCK-001 must be updated deliberately rather than relaxed to a length check.
