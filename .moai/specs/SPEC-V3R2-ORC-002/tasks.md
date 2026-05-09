---
spec_id: SPEC-V3R2-ORC-002
phase: "1B — Tasks"
created_at: 2026-05-10
author: manager-spec
total_tasks: 22
---

# Tasks: Agent Common Protocol CI Lint (`moai agent lint`)

Discrete task breakdown for SPEC-V3R2-ORC-002 plan execution. Each task has
ID `T-ORC002-NN` (zero-padded 2-digit), milestone owner, dependency
declaration, and an AC/test handle.

Owner roles align with TDD methodology (`development_mode: tdd` in
`.moai/config/sections/quality.yaml`):

- **manager-tdd**: orchestrates RED → GREEN → REFACTOR cycle for the lint
  engine.
- **expert-backend**: authors Go source (lint engine, frontmatter parser,
  rule logic) and tests.
- **manager-spec**: rule-file amendment (extracting Skeptical block) and
  documentation.
- **expert-devops**: CI workflow integration.
- **manager-quality**: final verification and TRUST 5 audit.
- **manager-docs**: pre-commit hook documentation snippet.

---

## M1 — Baseline Capture + Delivery Contract (P0)

### T-ORC002-01 — Re-verify research.md violation baseline

- **Owner**: manager-spec
- **Depends on**: (none — entry task)
- **Action**: Run all `grep` invocations from research.md §1 (LR-01,
  LR-02, LR-07 violation locators). Confirm counts match the documented
  tables (9 unique LR-01 agents, 4-5 LR-02 hits, 2 Skeptical block
  occurrences).
- **Output**: notation in progress.md (M1.1 ✓).
- **Test**: pre-existing tree state (read-only).
- **Acceptance**: AC-V3R2-ORC-002-02 (baseline regression target).

### T-ORC002-02 — Capture baseline snapshot artefact

- **Owner**: expert-backend
- **Depends on**: T-ORC002-01
- **Action**: Save the violation list (file:line tuples) to
  `.moai/specs/SPEC-V3R2-ORC-002/baseline-snapshot.txt` for M3 RED test
  reference.
- **Files**: `.moai/specs/SPEC-V3R2-ORC-002/baseline-snapshot.txt` (NEW).
- **Test**: artefact existence; line count matches research.md §1.
- **Acceptance**: AC-V3R2-ORC-002-02 (snapshot used as test fixture).

### T-ORC002-03 — Resolve OQ-1..6 in HISTORY

- **Owner**: manager-spec
- **Depends on**: T-ORC002-02
- **Action**: Append HISTORY row v0.1.1 to spec.md with OQ-1..6
  resolutions per plan.md §3.1 step M1.3. (Author the row in the run-phase
  PR; this plan-phase PR does not modify spec.md.)
- **Files**: (deferred — plan PR documents the intent only.)
- **Test**: spec.md HISTORY table row at v0.1.1 in run-phase PR.
- **Acceptance**: traceability to all ACs (foundation).

---

## M2 — Cobra Subcommand Wiring + agent-common-protocol Amendment (P0)

### T-ORC002-04 — Author `internal/cli/agent_lint.go` skeleton

- **Owner**: expert-backend
- **Depends on**: T-ORC002-03
- **Action**: Create new file with `newAgentCmd()` factory returning
  `agent` parent + `agent lint` child cobra.Commands. Stub `RunE` returns
  `errors.New("lint engine not implemented")`. Define flag set: `--path`
  (StringSlice), `--format` (String, default "text"), `--strict` (Bool).
- **Files**: `internal/cli/agent_lint.go` (NEW, ~80 lines).
- **Test**: `go vet ./internal/cli/...` clean.
- **Acceptance**: AC-V3R2-ORC-002-01 (`--help` works after registration).

### T-ORC002-05 — Register subcommand in `internal/cli/root.go`

- **Owner**: expert-backend
- **Depends on**: T-ORC002-04
- **Action**: Add `rootCmd.AddCommand(newAgentCmd())` in `init()` after
  L81 (state command).
- **Files**: `internal/cli/root.go` (MODIFIED, +1 line + import if needed).
- **Test**: `moai agent lint --help` exits 0; `moai agent lint` returns
  the stub error.
- **Acceptance**: AC-V3R2-ORC-002-01.

### T-ORC002-06 — Smoke test for subcommand registration

- **Owner**: expert-backend (TDD pair: manager-tdd)
- **Depends on**: T-ORC002-05
- **Action**: Author `internal/cli/agent_lint_test.go` with two smoke
  tests: `TestAgentLint_HelpFlag` (passes), `TestAgentLint_StubFails`
  (passes against stub; flips to GREEN as M3 lands).
- **Files**: `internal/cli/agent_lint_test.go` (NEW, ~80 lines).
- **Test**: `go test ./internal/cli/... -run TestAgentLint_HelpFlag` exit
  0.
- **Acceptance**: AC-V3R2-ORC-002-01.

### T-ORC002-07 — Amend `agent-common-protocol.md` with §Skeptical Evaluation Stance

- **Owner**: manager-spec
- **Depends on**: T-ORC002-06
- **Action**: Insert new section (header + 6 canonical bullets) after the
  `## Output Format` section in
  `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`
  (Template-First).
- **Files**: `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` (MODIFIED).
- **Test**: `grep -c "^## Skeptical Evaluation Stance$"` on file = 1.
- **Acceptance**: AC-V3R2-ORC-002-12.

### T-ORC002-08 — Mirror amendment to local rule tree + zone-registry update

- **Owner**: manager-spec
- **Depends on**: T-ORC002-07
- **Action**:
  1. Run `make build` to mirror the rule-file change to
     `.claude/rules/moai/core/agent-common-protocol.md`.
  2. Update `internal/template/templates/.claude/rules/moai/core/zone-registry.md`
     with new entry (CONST-V3R2-049 or next available) marking the new
     section as EVOLVABLE.
  3. Run `make build` again to mirror zone-registry.
- **Files**: `internal/template/templates/.claude/rules/moai/core/zone-registry.md`
  (MODIFIED), `internal/template/embedded.go` (REGENERATED).
- **Test**: `diff -r` template↔local rules empty.
- **Acceptance**: AC-V3R2-ORC-002-12.

### T-ORC002-09 — Remove duplicate Skeptical block from manager-quality + evaluator-active

- **Owner**: expert-backend (file edit), manager-spec (review)
- **Depends on**: T-ORC002-08
- **Action**: In both `manager-quality.md` (template + local) and
  `evaluator-active.md` (template + local), delete the existing
  `## Skeptical Evaluation Mandate` section + 6 bullets. Replace with a
  one-line reference: `> See \`agent-common-protocol.md\` §Skeptical
  Evaluation Stance.`
- **Files**:
  `internal/template/templates/.claude/agents/moai/manager-quality.md`,
  `internal/template/templates/.claude/agents/moai/evaluator-active.md`,
  + local mirrors via `make build`.
- **Test**: `grep -c "Skeptical Evaluation Mandate"
  .claude/agents/moai/*.md` = 0; `grep -c "Skeptical Evaluation Stance"`
  on the rule file = 1.
- **Acceptance**: AC-V3R2-ORC-002-07, AC-V3R2-ORC-002-12.

### T-ORC002-10 — Drop dead `Agent` from expert-security tools list

- **Owner**: expert-backend
- **Depends on**: T-ORC002-09
- **Action**: Edit
  `internal/template/templates/.claude/agents/moai/expert-security.md` L13
  (`tools:` line) to remove `Agent` token. Run `make build` to mirror.
- **Files**:
  `internal/template/templates/.claude/agents/moai/expert-security.md`
  (MODIFIED), local mirror, `internal/template/embedded.go`.
- **Test**: `grep -E "^tools:.*\bAgent\b"
  .claude/agents/moai/expert-security.md` returns 0 lines.
- **Acceptance**: AC-V3R2-ORC-002-02 (LR-02 count reduces).

---

## M3 — TDD RED → GREEN: 8 Lint Rules (P0)

### T-ORC002-11 — RED: author 14-17 AC-driven test fixtures + assertions

- **Owner**: manager-tdd (orchestrate), expert-backend (write)
- **Depends on**: T-ORC002-10
- **Action**:
  1. Create 8 testdata fixture files in
     `internal/cli/testdata/agent_lint/`:
     - `fixture-clean.md` (passes all rules)
     - `fixture-lr01-violation.md` (literal AskUserQuestion outside fence)
     - `fixture-lr01-fence-ok.md` (AskUserQuestion only inside fence)
     - `fixture-lr01-inline-code.md` (inline-code, IS flagged per OQ-2)
     - `fixture-lr02-violation.md` (Agent in tools)
     - `fixture-lr04-dead-hook.md` (matcher refs absent tool)
     - `fixture-lr07-duplicate.md` (second copy of Skeptical block)
     - `fixture-malformed.md` (invalid YAML)
     - `fixture-orchestrator-allow.md` (manager-brain-style allowlist)
  2. Author 14 test functions in `agent_lint_test.go` per plan.md §3.3
     M3.1 list. Each function asserts AC-V3R2-ORC-002-NN.
- **Files**: `internal/cli/testdata/agent_lint/*.md` (NEW, 8-9 fixtures),
  `internal/cli/agent_lint_test.go` (EXTENDED, ~400 lines).
- **Test**: All test functions FAIL (RED gate; rule logic not implemented).
- **Acceptance**: All 14 ACs (test-skeleton).

### T-ORC002-12 — GREEN part 1: frontmatter parser + body scanner skeleton

- **Owner**: expert-backend
- **Depends on**: T-ORC002-11
- **Action**:
  1. Implement `parseFrontmatter([]byte) (AgentFrontmatter, body string,
     err error)` using `gopkg.in/yaml.v3`.
  2. Implement `bodyScanner` struct with fence-state tracking per
     research.md §5.1.
  3. Wire `RunE` to walk default paths (`.claude/agents/moai/` +
     `internal/template/templates/.claude/agents/moai/`) or `--path`
     overrides.
  4. Emit empty violation list (no rules implemented yet).
- **Files**: `internal/cli/agent_lint.go` (EXTENDED, +200 lines).
- **Test**: `TestAgentLint_AC01_HelpFlag` GREEN;
  `TestAgentLint_AC11_MalformedYAML` GREEN.
- **Acceptance**: AC-V3R2-ORC-002-01, AC-V3R2-ORC-002-11.

### T-ORC002-13 — GREEN part 2: implement LR-01, LR-02, LR-03

- **Owner**: expert-backend
- **Depends on**: T-ORC002-12
- **Action**:
  1. **LR-01**: body-line scan for `AskUserQuestion`; skip fences; exempt
     orchestrator-class agents (frontmatter `tools:` contains
     `AskUserQuestion`).
  2. **LR-02**: split `tools:` CSV; reject any token equal to `Agent`.
  3. **LR-03**: warn on `frontmatter.Effort == ""`; promote to error
     under `--strict`.
- **Files**: `internal/cli/agent_lint.go` (EXTENDED, +150 lines).
- **Test**: `TestAgentLint_AC02_BaselineV2_13_2`,
  `TestAgentLint_AC03_CleanRoster`, `TestAgentLint_AC08_MissingEffort_NonStrict`,
  `TestAgentLint_AC09_MissingEffort_Strict`,
  `TestAgentLint_AC10_FencedCodeExempt`, `TestAgentLint_AC13_Orchestrator`,
  `TestAgentLint_InlineCodeFlagged` GREEN.
- **Acceptance**: AC-V3R2-ORC-002-02, -03, -08, -09, -10, -13.

### T-ORC002-14 — GREEN part 3: implement LR-04, LR-05, LR-06

- **Owner**: expert-backend
- **Depends on**: T-ORC002-13
- **Action**:
  1. **LR-04**: parse `hooks:` array; for each hook, split `matcher:` by
     `\|`; emit `LR-04-COMPLEX-MATCHER` warning if metacharacters detected;
     compare each token to tools list; mismatch = LR-04 error.
  2. **LR-05**: detect role profile via `name`-prefix or explicit `role:`
     field; for write-heavy roles {implementer, tester, designer}, warn
     if `isolation: worktree` missing; error under `--strict`.
  3. **LR-06**: scan first 30 body lines (description-area heuristic) for
     `--deepthink` substring; warn (REQ-A19); error under `--strict`.
- **Files**: `internal/cli/agent_lint.go` (EXTENDED, +120 lines).
- **Test**: `TestAgentLint_AC06_DeadHook`,
  `TestAgentLint_ComplexMatcher` GREEN.
- **Acceptance**: AC-V3R2-ORC-002-06.

### T-ORC002-15 — GREEN part 4: implement LR-07, LR-08, JSON output, tree-drift

- **Owner**: expert-backend
- **Depends on**: T-ORC002-14
- **Action**:
  1. **LR-07**: first scan rule files (`.claude/rules/**/*.md`) then
     agent files; compute Skeptical-block fingerprint (header anchor +
     bullet count + SHA-256 sorted-lowercased bullets); first match
     allowed; subsequent = LR-07 error.
  2. **LR-08**: group agents by `name`-prefix family; compute
     family-skill-preload union; warn on agents missing skills that ≥50%
     of family preloads.
  3. **JSON output (REQ-010)**: emit canonical schema per plan.md §4.4
     when `--format=json`. Include `exemptions[]` entries for
     orchestrator carve-outs.
  4. **Tree drift (REQ-017)**: compare per-file violation tuples across
     trees; emit `LINT_TREE_DRIFT` (per-file violation diff) and
     `LINT_TREE_FILE_MISMATCH` (file presence diff) warnings.
- **Files**: `internal/cli/agent_lint.go` (EXTENDED, +200 lines).
- **Test**: `TestAgentLint_AC04_JSONOutput`,
  `TestAgentLint_AC07_DuplicateSkeptical`,
  `TestAgentLint_AC12_CanonicalSkeptical`,
  `TestAgentLint_AC14_TreeDrift` GREEN.
- **Acceptance**: AC-V3R2-ORC-002-04, -07, -12, -14.

### T-ORC002-16 — Verify M3 GREEN gate: full test pass + coverage

- **Owner**: manager-quality (verify), manager-tdd (orchestrate)
- **Depends on**: T-ORC002-15
- **Action**:
  1. Run `go test ./internal/cli/... -run TestAgentLint -v -race`.
  2. Run `go test -cover ./internal/cli/agent_lint*.go`.
  3. Run `go vet ./internal/cli/... && golangci-lint run`.
  4. Hand-run `moai agent lint` on live tree; compare to research.md §11
     expected output.
- **Files**: (verification only).
- **Test**: All 14-17 lint test functions GREEN; race-free; coverage ≥
  85% on `internal/cli/agent_lint*.go`; lint output matches expected.
- **Acceptance**: ALL 14 ACs pass at unit-test level.

---

## M4 — CI Integration + JSON Schema Freeze (P1)

### T-ORC002-17 — Add `moai agent lint` step to ci.yaml

- **Owner**: expert-devops
- **Depends on**: T-ORC002-16
- **Action**: Edit
  `.github/workflows/ci.yaml` to add a step after the existing Go-lint
  step in the **Lint** job:
  ```yaml
  - name: Run moai agent lint
    run: ./bin/moai agent lint
  ```
  Confirm `Build` job (or equivalent step that produces `./bin/moai`) is
  available.
- **Files**: `.github/workflows/ci.yaml` (MODIFIED, +3 lines).
- **Test**: open synthetic-violation PR; CI Lint job exits 1; revert →
  CI Lint job exits 0.
- **Acceptance**: AC-V3R2-ORC-002-05.

### T-ORC002-18 — Document JSON schema and pre-commit hook

- **Owner**: manager-docs
- **Depends on**: T-ORC002-17
- **Action**:
  1. Author `.moai/docs/agent-lint.md` (NEW) with:
     - Subcommand overview.
     - JSON schema (REQ-014 v1.0).
     - Pre-commit hook YAML snippet (REQ-013).
     - Exit-code matrix.
     - LR-01..08 rule descriptions (one paragraph each).
  2. Mirror under `internal/template/templates/.moai/docs/agent-lint.md`.
- **Files**: `.moai/docs/agent-lint.md` (NEW), template mirror, embedded
  regen.
- **Test**: `grep -A5 "pre-commit" .moai/docs/agent-lint.md` shows YAML
  snippet.
- **Acceptance**: AC-V3R2-ORC-002-04.b.

### T-ORC002-19 — Manual smoke test: synthetic CI violation

- **Owner**: manager-quality
- **Depends on**: T-ORC002-17
- **Action**: On a throwaway branch, append `Use AskUserQuestion` line to
  expert-backend.md, push, observe CI red status, revert.
- **Files**: (transient — no commit).
- **Test**: CI Lint job exits 1; status check turns red.
- **Acceptance**: AC-V3R2-ORC-002-05.

---

## M5 — REFACTOR + MX Tags + Completion Gate (P0)

### T-ORC002-20 — REFACTOR: table-driven rule dispatch + helper extraction

- **Owner**: expert-backend (refactor), manager-quality (review)
- **Depends on**: T-ORC002-19
- **Action**:
  1. Introduce `lintRuleSpec` struct
     `{ID string, Severity Severity, Run func(...) []Violation}`.
  2. Convert per-rule logic in `agent_lint.go` to entries in a `rules`
     slice; main loop iterates the slice.
  3. Extract `bodyScanner` fence-state machine into a typed helper
     struct.
  4. Re-run all M3 tests; assert no regression.
- **Files**: `internal/cli/agent_lint.go` (REFACTORED, ~500-700 lines
  final).
- **Test**: `go test ./internal/cli/... -race -v` zero failures;
  coverage unchanged.
- **Acceptance**: ALL 14 ACs preserved.

### T-ORC002-21 — Apply final @MX tags + resolve @MX:TODO

- **Owner**: expert-backend (apply), manager-quality (verify)
- **Depends on**: T-ORC002-20
- **Action**:
  1. `@MX:ANCHOR cmd-registration` on `internal/cli/root.go` new line.
  2. `@MX:ANCHOR lint-rule-table` on rule-spec slice in
     `agent_lint.go`.
  3. `@MX:WARN fenced-code-state` on body-scanner state machine.
  4. `@MX:NOTE skeptical-extracted` on rule-file new section header.
  5. `@MX:WARN skeptical-no-duplicates` on rule-file new section.
  6. `@MX:NOTE orchestrator-exempt` on manager-brain carve-out logic.
  7. `@MX:NOTE ci-integration` on new CI step.
  8. Resolve `@MX:TODO lr08-threshold-tune` (decision: keep 50% per M3.6
     observations).
  9. Run `moai mx scan` (or equivalent grep) to confirm zero TODOs.
- **Files**: `internal/cli/agent_lint.go`, `internal/cli/root.go`,
  `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`,
  `.github/workflows/ci.yaml`.
- **Test**: `grep -rn "@MX:TODO" internal/cli/agent_lint.go` returns
  empty; all other tags present.
- **Acceptance**: TRUST 5 Trackable.

### T-ORC002-22 — Final verification + push + PR

- **Owner**: manager-quality (final gate), manager-spec (sign-off)
- **Depends on**: T-ORC002-21
- **Action**:
  1. Run all 14 AC verifications from acceptance.md (manual or scripted).
  2. Run `make build && make test && go test -race ./...`.
  3. Run `golangci-lint run` on full project.
  4. Run `diff -r internal/template/templates/.claude/agents/moai/
     .claude/agents/moai/` and same for `.claude/rules/moai/core/`.
  5. Hand-run `moai agent lint --format=json | jq '.summary'` on live
     post-M2 tree; confirm matches the expected residual-violation list
     (until ORC-001 cleanup lands the rest).
  6. Update spec.md HISTORY with v0.1.1 row (post OQ-1..6 resolution
     note).
  7. Update progress.md with `plan_complete_at`, `plan_status:
     audit-ready`.
  8. Stage commit `feat(cli): SPEC-V3R2-ORC-002 — moai agent lint
     subcommand + agent-common-protocol Skeptical extraction`.
  9. Push branch.
  10. Open PR with plan-auditor request for review.
- **Files**: `progress.md` (final timestamp), `spec.md` (HISTORY only).
- **Test**: ALL acceptance criteria pass; CI green on PR; `diff -r`
  clean; `golangci-lint run` clean.
- **Acceptance**: ALL 14 ACs (final integration gate).

---

## Dependency Graph (textual)

```
T-01 → T-02 → T-03                           [M1 chain]
                ↓
T-04 → T-05 → T-06 → T-07 → T-08 → T-09 → T-10    [M2 chain]
                                                ↓
T-11 (RED) → T-12 → T-13 → T-14 → T-15 → T-16    [M3 chain]
                                              ↓
T-17 → T-18 → T-19    [M4 chain]
                  ↓
T-20 → T-21 → T-22    [M5 chain]
```

All milestone chains are sequential within each milestone (file edits
share state). Inter-milestone dependencies are explicit (T-10 must
complete before T-11; T-16 before T-17; T-19 before T-20).

---

## Owner Role Summary

| Owner | Tasks | Count |
|-------|-------|------:|
| manager-spec | T-01, T-03, T-07, T-08, T-22 (sign-off) | 5 |
| expert-backend | T-02, T-04, T-05, T-06, T-09, T-10, T-12, T-13, T-14, T-15, T-20, T-21 | 12 |
| manager-tdd | T-11 (orchestrate), T-16 (orchestrate) | 2 (overlap) |
| manager-quality | T-16, T-19, T-21 (verify), T-22 (final gate) | 4 |
| expert-devops | T-17 | 1 |
| manager-docs | T-18 | 1 |

**Total: 22 tasks across 6 owner roles** (manager-tdd overlaps as TDD
orchestrator on M3 RED+GREEN gates).

---

## Estimated Diff Size

| Milestone | Files touched | Approx lines added | Approx lines removed |
|-----------|--------------:|-------------------:|---------------------:|
| M1 | 1 (artefact) | 50 | 0 |
| M2 | ~8 (rule + 3 agents + root.go + 2 NEW Go files) | 200 | 50 (Skeptical block) |
| M3 | 2 Go files + 9 testdata | 800 | 0 |
| M4 | 2 (ci.yaml + agent-lint.md) | 80 | 0 |
| M5 | 2 (refactor only, line shuffle) | 50 | 50 |

Total: ~24 files changed; net ~+1,080 lines (700 in
`internal/cli/agent_lint.go` + tests + fixtures + docs); ~−100 lines
(Skeptical block removed × 2 files).

---

## Out-of-Scope Reminder

These items are NOT in this SPEC's task list:

- T-NN for ORC-001 cleanup of remaining 8 LR-01 violations (covered by
  ORC-001 own SPEC).
- T-NN for ORC-001 cleanup of LR-02 violations in
  builder-{agent,skill,plugin} (ORC-001 retires those agents to stubs).
- T-NN for expert-mobile LR-02 cleanup (post-R5 finding; future SPEC).
- T-NN for ORC-003 effort matrix population (LR-03 promotion to error).
- T-NN for ORC-004 isolation:worktree mandate (LR-05 promotion to error).
- T-NN for MIG-001 legacy SPEC body rewriter.

---

End of tasks.md.
