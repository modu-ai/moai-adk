# SPEC-PROJECT-CONTINUATION-KEY-001 — Acceptance Criteria

Card **t191** · **v0.2.0** · 15 criteria (13 blocking) · every criterion is binary, names the command that decides it, and has a red reachable by an implementation the plan describes.
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
| AC-PCK-014 | REQ-PCK-008 (hygiene) | — | non-blocking |
| AC-PCK-015 | REQ-PCK-012 | M5 | blocking |

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
**Then** the `none` row instructs skipping issuance entirely; names `Create SPEC` — the pre-P1 wording measured from `e91def4ca` — as the recommended option; enumerates exactly the four pre-P1 options, **omitting `Create SPEC later`**; states that the four-option cap is met without routing any option to `Other`; and no `moai todo add` invocation is reachable on that path.

### AC-PCK-005 — `card` reproduces P1 and stops at `/moai plan`

**Given** the same file,
**When** the `card` row of the Step 4.2 table is read,
**Then** it names `Create the SPEC and start now` as the recommended option; its terminal instruction is `/moai plan` with **no** instruction to proceed to run-phase or to emit the Implementation Kickoff Approval gate; and Step 4.1.5's five standing-source steps (harness-spec read, empty-goal skip, duplicate-suppression read, single `[PROJECT] ` add, id reported) are present and unmodified from their P1 text.

### AC-PCK-006 — `pipeline` carries past `/moai plan`, and carries the gate clause with it

**Given** the same file,
**When** the `pipeline` row of the Step 4.2 table is read,
**Then** all three hold:

1. **The delta.** The row instructs the session to continue past `/moai plan` to **emitting** the Implementation Kickoff Approval gate in this same session. A row whose terminal instruction is `/moai plan` — i.e. a `card`-behaving implementation — fails this conjunct. This is the criterion that distinguishes `pipeline` from `card`; it is red for any implementation that ships `pipeline` as a wording change.
2. **The gate clause.** `grep -c "Implementation Kickoff Approval"` restricted to the `pipeline` row returns **≥ 1**, in the same terms as the `card` option.
3. **No answering.** The row instructs the orchestrator to emit the gate and stop for the operator's answer; it contains no instruction to select, answer, pre-fill, or default that gate.

Conjunct 3 replaces v0.1.0's untestable negative ("contains no instruction to …", which any silent row satisfied) with a positive assertion about what the row must say.

### AC-PCK-007 — the question survives every value

**Given** the same file,
**When** `grep -n "No branch is taken on the operator's behalf" doc-generation.md` and the new value-independence sentence are read,
**Then** both clauses are present, neither is scoped to a subset of values, and the new sentence states that the key changes only which option is recommended and how it is worded.

### AC-PCK-008 — every run-phase-offering branch carries the kickoff clause

**Given** `doc-generation.md` on the implementation branch,
**When** the Step 4.2 recommended-option table is read for all three values,
**Then** each recommended-option text that offers run-phase entry carries the Implementation Kickoff Approval sentence verbatim, and

```bash
grep -c "Implementation Kickoff Approval" .claude/skills/moai/workflows/project/doc-generation.md
```

returns a count **no lower than the number of such options** — pre-change value `1`, measured on `2660bcd09`; expected `≥ 2` after M1, since `card` and `pipeline` both offer run-phase entry and `none` does not.

**Reachable red, established.** The v0.1.0 criterion was a `git diff --stat origin/develop...HEAD` over three gate-owning files; run verbatim at plan time it returned empty with `rc=0`, so it was already satisfied before the SPEC existed, and none of the three files appears in any M1-M6 file list. This replacement goes red on the leak it is meant to catch: M1 rewrites the Step 4.2 option block containing the sole occurrence, so an implementer who writes the `pipeline` row without the clause drops the count below the option count.

### AC-PCK-009 — the shipped key passes the anti-rot guard

**Given** the template `workflow.yaml` carrying `continuation: card` and the inventory carrying the new row,
**When** `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders` runs,
**Then** the test passes, and the same test fails when the inventory row alone is removed (the mutant establishes the guard is not vacuous here).

### AC-PCK-010 — the wizard offers the closed set in four locales

**Given** the wizard package on the implementation branch,
**When** `go test ./internal/cli/wizard/...` runs,
**Then** all three hold:

1. The `project_continuation` question's `Type` is `QuestionTypeSelect`, its `Default` is `card`, and its option values equal `config.ValidProjectContinuations()`.
2. For each of `ko` / `ja` / `zh`, the `translations` entry carries a non-empty `Title`, a non-empty `Description`, and an `Options` slice of length **3**, every element having a non-empty `Label` **and** a non-empty `Desc` — the `audit_model` shape (`translations.go:137`).
3. **`TestWizardQuestionTranslationCompleteness` passes.** That existing test (`translations_completeness_test.go:95-134`) is the enforcer for conjunct 2, and `project_continuation` must NOT be added to `optionTranslationExemptIDs` to satisfy it.

**Reachable red.** `grep -rn "project_continuation" internal/cli/wizard/ | wc -l` → `0` today, so the question does not exist. Adding it as a non-exempt select with no option translations fails the completeness test with **one error per locale (3 across ko/ja/zh)** — the `len(trans.Options) != len(q.Options)` branch `continue`s at `:123` before the per-option loop runs, so the count is 3, not 9.

### AC-PCK-011 — the console field is closed-set and fully localized

**Given** the settings schema on the implementation branch,
**When** `go test ./internal/settings/... ./internal/web/...` runs,
**Then** a `FieldDef` exists for path `workflow → project → continuation` whose option values derive from `config.ValidProjectContinuations()` and which is wrapped in `withOptionDesc` (`schema_sections.go:143`), and `internal/web/assets/i18n.js` carries, in **each** of the four locale maps, all eight of: `f.workflow.project.continuation.title`, `.desc`, three `.opt.<token>` labels, and three `.option.<token>` descriptions — **8 keys × 4 locales = 32 entries, none empty**.

**Reachable red.** `grep -c "f.workflow.project" internal/web/assets/i18n.js` → `0` today. A bare `closedSeam(...)` without the `withOptionDesc` wrapper emits no `.option.` keys and fails the count; a `FieldDef` whose keys are missing from any locale map fails `TestI18nKeyCoverageForward` / `TestI18nKeyCoverageReverse` (`internal/web/i18n_governance_test.go:408,423`).

### AC-PCK-012 — mirror parity holds after `make build`

**Given** a completed `make build`,
**When** `cmp` is run over the three byte-identical pairs (`doc-generation.md`, `todo.md`, `tab_schema.json`) and over `workflow.yaml`,
**Then** the three return rc=0 and `workflow.yaml` returns rc=1, with `diff` showing only the intended repo-local content (`branch_guard`, `agent_stop_guard`, populated `audit` pins, `context_folding`) plus the new key on both sides.

### AC-PCK-013 — the unmatched value reaches the operator (non-blocking)

**Given** the Step 4.2 report contract in `doc-generation.md`,
**When** the report template is read,
**Then** it carries a line that prints the offending value together with the canonical domain when the resolver returned a non-empty unmatched string, and prints nothing when it did not.

### AC-PCK-014 — gate-owning rule files stay untouched (non-blocking hygiene)

**Given** the implementation branch,
**When** `git diff --stat <merge-base>...HEAD -- .claude/rules/moai/workflow/orchestration-mode-selection.md .claude/rules/moai/workflow/goal-directive.md .claude/skills/moai/workflows/goal.md` runs,
**Then** the output is empty.

**Filed non-blocking, deliberately.** This is the v0.1.0 `AC-PCK-008` first conjunct, retained as hygiene. It is **already satisfied at plan time** and none of the three files appears in any M1-M6 file list, so it carries no reachable red and MUST NOT be presented as the verification of `REQ-PCK-008` — `AC-PCK-008` is. Its value is regression signalling if a later iteration widens scope. The base is the merge-base rather than `origin/develop`, which has advanced 35 commits past this SPEC's declared baseline.

### AC-PCK-015 — the `.opt.` English-label guard is preserved

**Given** the implementation branch,
**When** `go test ./internal/web/...` runs,
**Then** the option-label keys use the `.opt.` prefix, the per-option-description keys use a prefix containing no `.opt.` substring (`.option.` satisfies this — the character after `opt` is `i`), and `applyI18n`'s `".opt."` guard in `app.js` is unmodified.

**Reachable red.** `internal/web/audit_option_desc_test.go:137-138` fails with "app.js lost the `.opt.` guard — enum option labels would follow the active locale, reversing the G1-2 decision" if the guard is removed, and `:78` fails if a description key contains `.opt.`. An implementer who names the description keys `f.workflow.project.continuation.opt.desc.<token>` trips the second.

---

## §D.1 Severity

Thirteen blocking, two non-blocking. `AC-PCK-013` is non-blocking because it is a report-wording obligation whose absence degrades diagnosis without changing behaviour (`AC-PCK-003` guarantees the resolution itself). `AC-PCK-014` is non-blocking because it has no reachable red — see its own note.

## §D.2 Traceability

Every `REQ-PCK-001..012` appears in at least one **blocking** AC row of §D. `REQ-PCK-007` and `REQ-PCK-008` are prohibitions verified positively rather than negatively: `AC-PCK-007` by clause presence and scope, `AC-PCK-008` by a per-branch kickoff-clause count against a measured pre-change value. The v0.1.0 diff-stat check is explicitly **struck from `REQ-PCK-008`'s verification path** and re-filed as `AC-PCK-014` hygiene.

## §D.3 Indirect verification

REQ-PCK-004/005/006 govern orchestrator prose, which has no runtime harness. They are verified by reading the shipped instruction text, which is the artifact the orchestrator actually consumes. This is indirect and is recorded as such.

## §D.4 Closure gates

- All thirteen blocking ACs pass.
- `go test ./internal/config/... ./internal/settings/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/web/...` is green, with each package's output cited.
- `golangci-lint run` on the touched packages is clean.
- `git status --short` shows no unintended file.

## §D.5 Forward-looking checks

- **The folding fallback is withdrawn.** v0.1.0 blessed resolving `REQ-PCK-010` by folding option descriptions into the question body; that path fails `TestWizardQuestionTranslationCompleteness` and is now forbidden by `AC-PCK-010` conjunct 3. `GetLocalizedQuestion` needs no extension — it already carries option translations (`translations.go:571-586`).
- If a later SPEC adds a fourth continuation token, `AC-PCK-001` must be updated deliberately rather than relaxed to a length check, and `AC-PCK-011`'s entry count rises from 32 to 40.
- If a later SPEC gives `pipeline` a mechanical carrier (for example a `progression_mode` default), `AC-PCK-006` conjunct 1 must be re-derived against that carrier rather than against the option text — see `spec.md` §3 D1.3 for why the progression-mode reading was rejected at this iteration.
