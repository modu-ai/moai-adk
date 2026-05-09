---
spec_id: SPEC-V3R2-ORC-002
phase: "1B — Acceptance Criteria"
created_at: 2026-05-10
author: manager-spec
ac_count: 14
ac_format: hierarchical
---

# Acceptance Criteria: Agent Common Protocol CI Lint (`moai agent lint`)

Given-When-Then scenarios for every requirement in spec.md §5. Hierarchical
format per SPEC-V3R2-SPC-001 schema: top-level `AC-V3R2-ORC-002-NN` with
optional `.a/.b/.c` sub-children when a single AC fans out into related
verifications.

Numbering aligns to spec.md §6 (AC-ORC-002-01 .. AC-ORC-002-12) extended
with 2 supplementary ACs (AC-13 .. AC-14) covering OQ-1 carve-out and tree
drift.

---

## Traceability Matrix (REQ ↔ AC)

### Forward (REQ → AC)

| REQ-ID | AC-ID(s) | Coverage notes |
|--------|----------|----------------|
| REQ-ORC-002-001 | AC-V3R2-ORC-002-01 | `moai agent lint --help` flag surface |
| REQ-ORC-002-002 | AC-V3R2-ORC-002-02, AC-V3R2-ORC-002-03 | Default scan path + dual-tree |
| REQ-ORC-002-003 | AC-V3R2-ORC-002-02..AC-V3R2-ORC-002-10, AC-V3R2-ORC-002-13, AC-V3R2-ORC-002-14 | All 8 LR rules verified individually |
| REQ-ORC-002-004 | AC-V3R2-ORC-002-08, AC-V3R2-ORC-002-09, AC-V3R2-ORC-002-11 | Exit code matrix (0/1/2/3 + strict promotion) |
| REQ-ORC-002-005 | AC-V3R2-ORC-002-12 | §Skeptical Evaluation Stance section added once |
| REQ-ORC-002-006 | AC-V3R2-ORC-002-02, AC-V3R2-ORC-002-05, AC-V3R2-ORC-002-10 | LR-01 detect + violation record fields |
| REQ-ORC-002-007 | AC-V3R2-ORC-002-02 | LR-02 `Agent` token detect |
| REQ-ORC-002-008 | AC-V3R2-ORC-002-06 | LR-04 dead-hook detect |
| REQ-ORC-002-009 | AC-V3R2-ORC-002-07, AC-V3R2-ORC-002-12 | LR-07 duplicate-block detect |
| REQ-ORC-002-010 | AC-V3R2-ORC-002-04 | JSON output schema + `summary.errors` parity |
| REQ-ORC-002-011 | AC-V3R2-ORC-002-05 | CI required-status integration |
| REQ-ORC-002-012 | AC-V3R2-ORC-002-08, AC-V3R2-ORC-002-09 | `--strict` mode promotion |
| REQ-ORC-002-013 | AC-V3R2-ORC-002-04 (.b) | Pre-commit hook (Optional, documentation only) |
| REQ-ORC-002-014 | AC-V3R2-ORC-002-04 (.a) | JSON `version` field stability |
| REQ-ORC-002-015 | AC-V3R2-ORC-002-10 | Fenced-code exemption |
| REQ-ORC-002-016 | AC-V3R2-ORC-002-11 | Malformed YAML → exit 2 |
| REQ-ORC-002-017 | AC-V3R2-ORC-002-14 | Two-tree drift LINT_TREE_DRIFT |

### Reverse (AC → REQ)

| AC-ID | REQ-ID(s) | Coverage |
|-------|-----------|----------|
| AC-V3R2-ORC-002-01 | REQ-001 | Direct |
| AC-V3R2-ORC-002-02 | REQ-002, REQ-003, REQ-006, REQ-007 | Multi-rule |
| AC-V3R2-ORC-002-03 | REQ-002, REQ-003 | Clean state |
| AC-V3R2-ORC-002-04 | REQ-010, REQ-013 (.b), REQ-014 (.a) | JSON + Optional |
| AC-V3R2-ORC-002-05 | REQ-006, REQ-011 | New violation + CI |
| AC-V3R2-ORC-002-06 | REQ-008 | LR-04 |
| AC-V3R2-ORC-002-07 | REQ-009 | LR-07 |
| AC-V3R2-ORC-002-08 | REQ-004 (warning), REQ-012 | LR-03 non-strict |
| AC-V3R2-ORC-002-09 | REQ-004 (error), REQ-012 | LR-03 strict |
| AC-V3R2-ORC-002-10 | REQ-006, REQ-015 | Fenced-code |
| AC-V3R2-ORC-002-11 | REQ-004 (exit 2), REQ-016 | Malformed YAML |
| AC-V3R2-ORC-002-12 | REQ-005, REQ-009 | Canonical Skeptical |
| AC-V3R2-ORC-002-13 | REQ-006 (carve-out) | Manager-brain orchestrator |
| AC-V3R2-ORC-002-14 | REQ-017 | Tree drift |

**Coverage**: 17 REQs → 14 ACs; every REQ covered by ≥ 1 AC.

---

## AC-V3R2-ORC-002-01 — `moai agent lint --help` prints usage with all flags (REQ-001)

**Given** the `moai` binary is built from this branch.
**When** I run `moai agent lint --help`.
**Then**:
1. Exit code is 0.
2. Output contains the subcommand description ("Lint agent files for
   common-protocol compliance" or similar).
3. Output lists all four flags: `--path`, `--format`, `--strict`, `--help`.
4. Each flag has a one-line description.

**Verification**:

```bash
./bin/moai agent lint --help
echo "Exit: $?"
# Expected: text contains "--path", "--format", "--strict"; exit 0
```

**Edge cases**:
- `moai agent --help` (parent without `lint`) lists `lint` as a sub-command.
- `moai help agent lint` (cobra alternate help syntax) returns same output.

---

## AC-V3R2-ORC-002-02 — Baseline v2.13.2 roster yields exactly the documented R5 violations (REQ-002, REQ-003, REQ-006, REQ-007)

**Given** the agent tree at base `ab0fc4dda` (template + local) before any
ORC-001 / ORC-002 cleanup.
**When** I run `moai agent lint`.
**Then**:
1. Exit code is 1 (errors present).
2. Output reports exactly 9 LR-01 violations (one per agent in the R5
   §Common Protocol compliance table: manager-strategy, builder-agent,
   manager-quality, expert-frontend, expert-backend, builder-skill,
   manager-spec, builder-plugin, expert-devops).
3. Output reports ≥ 4 LR-02 violations (builder-agent, builder-skill,
   builder-plugin, expert-security; possibly +expert-mobile post-R5
   addition).
4. Output reports 1 LR-07 violation (the duplicate Skeptical block in
   manager-quality.md OR evaluator-active.md, whichever comes second in
   path-sorted order).

**Verification**:

```bash
./bin/moai agent lint --format=json | jq '.summary.errors'
# Expected: ≥ 14 (9 LR-01 + 4-5 LR-02 + 1 LR-07)
./bin/moai agent lint --format=json | jq '[.violations[] | select(.rule=="LR-01")] | length'
# Expected: 9
./bin/moai agent lint --format=json | jq '[.violations[] | select(.rule=="LR-02")] | length'
# Expected: 4 or 5
./bin/moai agent lint --format=json | jq '[.violations[] | select(.rule=="LR-07")] | length'
# Expected: 1
```

**Edge cases**:
- manager-brain has 11 raw `AskUserQuestion` body matches but 0 LR-01
  violations because frontmatter `tools:` declares the API → carve-out
  applies (covered by AC-13).
- expert-mobile is post-R5 and may be present — accept LR-02 count of 4 OR
  5 in audit.

---

## AC-V3R2-ORC-002-03 — Post-cleanup roster yields zero violations (REQ-002, REQ-003)

**Given** ORC-001 has merged AND this SPEC's M2 surgery has run
(agent-common-protocol amended, manager-quality + evaluator-active
duplicates removed, expert-security `Agent` dropped).
**When** I run `moai agent lint`.
**Then**:
1. Exit code is 0.
2. JSON output reports `summary.errors: 0`.
3. Warnings may be non-zero (LR-03 missing-effort, LR-06 --deepthink
   boilerplate are pre-existing until ORC-003).

**Verification**:

```bash
./bin/moai agent lint
echo "Exit: $?"
# Expected: exit 0; possibly some warnings printed
./bin/moai agent lint --format=json | jq '.summary.errors'
# Expected: 0
```

**Edge cases**:
- If ORC-001 is partially merged (e.g., only manager-cycle but not
  builder-platform), LR-02 count may still be > 0; document in PR
  description.

---

## AC-V3R2-ORC-002-04 — JSON output schema valid + parity with text (REQ-010, REQ-013, REQ-014)

**Given** the lint produces violations on the v2.13.2 baseline.
**When** I run with `--format=json`.
**Then**:
1. Stdout is valid JSON parsable by `jq`.
2. Top-level fields present: `version`, `scanned_paths`, `summary`,
   `violations`.
3. `summary.errors` matches the count of `violations[].severity == "error"`.
4. `summary.warnings` matches the count of `violations[].severity ==
   "warning"`.

### AC-V3R2-ORC-002-04.a — JSON `version` field locked at "1.0" (REQ-014)

**When** I read the version field.
**Then** value is exactly `"1.0"` (string, not float).

```bash
./bin/moai agent lint --format=json | jq -r '.version'
# Expected: 1.0
```

### AC-V3R2-ORC-002-04.b — Pre-commit hook example present (REQ-013)

**When** I read `.moai/docs/agent-lint.md` (or README pre-commit section).
**Then** documentation includes a YAML snippet usable in
`.pre-commit-config.yaml` invoking `moai agent lint --path
.claude/agents/moai/`.

```bash
grep -A5 "pre-commit\|moai-agent-lint" .moai/docs/agent-lint.md
# Expected: snippet present
```

**Verification (root AC)**:

```bash
./bin/moai agent lint --format=json > /tmp/lint.json
jq -e '.version == "1.0" and (.summary.errors == ([.violations[] | select(.severity=="error")] | length))' /tmp/lint.json
# Expected: true
```

**Edge cases**:
- Trailing newline in stdout is acceptable; jq parses regardless.
- JSON output to non-tty must NOT emit ANSI colour codes.

---

## AC-V3R2-ORC-002-05 — New `AskUserQuestion` line fails CI (REQ-006, REQ-011)

**Given** a clean post-M2 tree with `summary.errors: 0`.
**When** a contributor adds the line `Use AskUserQuestion to confirm` (in
prose, outside fence) to any non-orchestrator agent and pushes the branch.
**Then**:
1. CI Lint job runs `moai agent lint` and exits 1.
2. CI status check turns red.
3. PR cannot be merged into a protected branch.

**Verification**:

```bash
# Synthetic injection (in run-phase smoke test, M4.4)
echo "" >> .claude/agents/moai/expert-backend.md
echo "Use AskUserQuestion to confirm parameters." >> .claude/agents/moai/expert-backend.md
./bin/moai agent lint
echo "Exit: $?"
# Expected: exit 1; output mentions expert-backend.md with LR-01

# Cleanup:
git checkout -- .claude/agents/moai/expert-backend.md
```

**Edge cases**:
- Adding the line inside a fenced code block must NOT fail (covered by
  AC-10).
- Adding the line to manager-brain must NOT fail (covered by AC-13).

---

## AC-V3R2-ORC-002-06 — Dead hook configuration triggers LR-04 (REQ-008)

**Given** a synthetic agent fixture (`testdata/agent_lint/fixture-lr04-dead-hook.md`)
declaring:
```yaml
hooks:
  - event: PreToolUse
    matcher: "Write|Edit"
    command: foo.sh
tools: Read, Grep
```
**When** I run `moai agent lint --path
internal/cli/testdata/agent_lint/fixture-lr04-dead-hook.md`.
**Then**:
1. Exit code is 1.
2. Output contains LR-04 violation listing matchers `Write` and `Edit` as
   absent from tools list.

**Verification**:

```bash
./bin/moai agent lint --path internal/cli/testdata/agent_lint/fixture-lr04-dead-hook.md
echo "Exit: $?"
# Expected: exit 1; output mentions LR-04, matcher Write|Edit
```

**Edge cases**:
- Matcher `Write` exists in tools list → no LR-04 violation.
- Matcher contains regex metacharacters (`^Write$`, `Write.*`) → soft
  warning `LR-04-COMPLEX-MATCHER` instead of hard error.

---

## AC-V3R2-ORC-002-07 — Duplicate Skeptical block triggers LR-07 (REQ-009)

**Given** the post-M2 tree has §Skeptical Evaluation Stance in
`agent-common-protocol.md` (canonical) AND a contributor pastes the same
6-bullet block into a new agent file (synthetic fixture
`testdata/agent_lint/fixture-lr07-duplicate.md`).
**When** I run `moai agent lint`.
**Then**:
1. Exit code is 1.
2. Output contains LR-07 violation pointing at the duplicate location.
3. The canonical location (`agent-common-protocol.md`) is NOT reported as
   a violation (first occurrence is allowed).

**Verification**:

```bash
./bin/moai agent lint --format=json | jq -r '.violations[] | select(.rule=="LR-07") | .file'
# Expected: only the duplicate file path; never agent-common-protocol.md
```

**Edge cases**:
- Block pasted with 5 bullets instead of 6 → fingerprint hash differs →
  no LR-07 (correctly, because semantic identity is not preserved).
- Block pasted with whitespace and casing changes → fingerprint
  normalisation neutralises → LR-07 still fires.

---

## AC-V3R2-ORC-002-08 — Missing `effort:` is warning in non-strict (REQ-004, REQ-012)

**Given** an agent file lacks the `effort:` frontmatter field (the live
state for all 26 agents at base `ab0fc4dda`).
**When** I run `moai agent lint` (no `--strict`).
**Then**:
1. Exit code is 0 (LR-03 is warning-only).
2. Output reports LR-03 violations under "WARNINGS:" section.
3. JSON `summary.warnings ≥ 1`; `summary.errors` excludes LR-03.

**Verification**:

```bash
./bin/moai agent lint --format=json | jq '.summary'
# Expected: errors: <some number>, warnings: ≥ 26 (one per file lacking effort)
./bin/moai agent lint --format=json | jq '[.violations[] | select(.rule=="LR-03")] | length'
# Expected: ≥ 26 in baseline
echo "Exit: $?"
# Expected: 0 (no errors triggered by LR-03 alone)
```

**Edge cases**:
- File with `effort: ""` (empty value) → still LR-03 warning (treated as
  missing).
- File with `effort: high` (or any non-empty value) → no LR-03 hit.

---

## AC-V3R2-ORC-002-09 — Missing `effort:` is error in `--strict` (REQ-004, REQ-012)

**Given** the same baseline as AC-08.
**When** I run `moai agent lint --strict`.
**Then**:
1. Exit code is 1 (LR-03 promoted to error).
2. JSON `summary.errors` includes LR-03 count.
3. Output reports the same LR-03 lines but under "ERRORS:" section.

**Verification**:

```bash
./bin/moai agent lint --strict
echo "Exit: $?"
# Expected: 1
./bin/moai agent lint --strict --format=json | jq '[.violations[] | select(.rule=="LR-03" and .severity=="error")] | length'
# Expected: ≥ 26
```

**Edge cases**:
- `--strict` also promotes LR-05, LR-06, LR-08 to error. AC-09 verifies
  LR-03 specifically; the others are covered by their respective rule
  fixtures.

---

## AC-V3R2-ORC-002-10 — Fenced-code blocks containing AskUserQuestion are NOT flagged (REQ-006, REQ-015)

**Given** a synthetic agent fixture (`testdata/agent_lint/fixture-lr01-fence-ok.md`)
whose body contains `AskUserQuestion` ONLY inside a triple-backtick fence:

````markdown
The orchestrator should:

```
Step 1: Call AskUserQuestion
Step 2: Process answer
```
````

**When** I run `moai agent lint --path <fixture>`.
**Then**:
1. Exit code is 0 (no LR-01 violation).
2. JSON `summary.errors == 0`.

**Verification**:

```bash
./bin/moai agent lint --path internal/cli/testdata/agent_lint/fixture-lr01-fence-ok.md
echo "Exit: $?"
# Expected: 0
./bin/moai agent lint --path internal/cli/testdata/agent_lint/fixture-lr01-fence-ok.md --format=json \
  | jq '.summary.errors'
# Expected: 0
```

**Edge cases**:
- Inline-code (single-backtick) `` `AskUserQuestion` `` IS flagged (per
  OQ-2 strict reading; covered by separate test fixture
  `fixture-lr01-inline-code.md`).
- Indented fence (4-space indent) is treated equivalently to backtick fence
  (skip the candidate).

---

## AC-V3R2-ORC-002-11 — Malformed YAML returns exit code 2 (REQ-004, REQ-016)

**Given** an agent file with broken YAML frontmatter (synthetic fixture
`testdata/agent_lint/fixture-malformed.md` with unbalanced quotes).
**When** I run `moai agent lint --path
internal/cli/testdata/agent_lint/`.
**Then**:
1. Exit code is 2 (parse error).
2. Output contains a single PARSE_ERROR record naming the offending file
   and the YAML parser error message.
3. Other files in the same scan continue to be linted (per REQ-016 — error
   isolation).

**Verification**:

```bash
./bin/moai agent lint --path internal/cli/testdata/agent_lint/
echo "Exit: $?"
# Expected: 2
./bin/moai agent lint --path internal/cli/testdata/agent_lint/ --format=json \
  | jq '.violations[] | select(.rule=="PARSE_ERROR")'
# Expected: 1 entry naming fixture-malformed.md
```

**Edge cases**:
- File with empty frontmatter (`---\n---\n`) → not malformed; treated as
  empty AgentFrontmatter; LR-03 fires (no effort).
- File with no frontmatter delimiter → treated as malformed; PARSE_ERROR.

---

## AC-V3R2-ORC-002-12 — `agent-common-protocol.md` contains exactly one Skeptical block (REQ-005, REQ-009)

**Given** M2.4 has run (extracted §Skeptical Evaluation Stance into the
rule file).
**When** I grep the rule file.
**Then**:
1. Exactly one occurrence of the section header
   `## Skeptical Evaluation Stance` exists.
2. The 6 canonical bullets follow the header.
3. No other rule file or agent file contains the 6-bullet pattern (verified
   by global grep).

**Verification**:

```bash
grep -c "^## Skeptical Evaluation Stance$" .claude/rules/moai/core/agent-common-protocol.md
# Expected: 1
grep -rln "^## Skeptical Evaluation Mandate$" .claude/agents/moai/ \
  internal/template/templates/.claude/agents/moai/
# Expected: empty (the old "Mandate" header should be gone)
grep -rln "^## Skeptical Evaluation Stance$" \
  .claude/rules/ internal/template/templates/.claude/rules/
# Expected: 2 lines (template + local mirror) of agent-common-protocol.md only
```

**Edge cases**:
- Mirror file in template tree may diverge if `make build` was skipped →
  caught by `diff -r` template↔local check.
- Section header lower-cased or with trailing whitespace → fingerprint
  algorithm detects regardless (LR-07 fires); explicit `^##` grep above
  catches the canonical case.

---

## AC-V3R2-ORC-002-13 — Manager-brain orchestrator carve-out exempts LR-01 (REQ-006 carve-out, OQ-1 resolution)

**Given** `.claude/agents/moai/manager-brain.md` declares
`AskUserQuestion` in its frontmatter `tools:` field (line 9).
**When** I run `moai agent lint`.
**Then**:
1. Zero LR-01 violations are reported for `manager-brain.md`, despite the
   body containing 11+ raw `AskUserQuestion` matches.
2. JSON output includes an exemption record:
   `exemptions: [{file: "...manager-brain.md", reason:
   "tools-asserts-aue"}]`.

**Verification**:

```bash
./bin/moai agent lint --format=json \
  | jq '[.violations[] | select(.file | contains("manager-brain.md") and .rule=="LR-01")] | length'
# Expected: 0
./bin/moai agent lint --format=json | jq '.exemptions[] | select(.file | contains("manager-brain.md"))'
# Expected: 1 entry with reason "tools-asserts-aue"
```

**Edge cases**:
- Future agent claims orchestrator role by adding `AskUserQuestion` to
  tools without the role context → carve-out applies (data-driven). This
  is intentional per OQ-1 Option A.
- Manager-brain frontmatter loses `AskUserQuestion` from `tools:` (e.g.,
  bug) → carve-out lifted; 11+ LR-01 violations fire. This is the
  intended fail-safe: misconfiguration is loud.

---

## AC-V3R2-ORC-002-14 — Two-tree drift produces LINT_TREE_DRIFT warning (REQ-017)

**Given** a synthetic divergence between `.claude/agents/moai/foo.md` and
`internal/template/templates/.claude/agents/moai/foo.md` where local has
an LR-01 violation but template does not (or vice-versa).
**When** I run `moai agent lint` (default = both trees).
**Then**:
1. Exit code respects the worse of the two trees.
2. JSON output includes `drift: {file: "foo.md", differences:
   [{rule: "LR-01", trees: ["local"]}]}`.
3. Text output prints `LINT_TREE_DRIFT` warning line referencing the
   diverging file.

**Verification** (synthetic test fixture in
`internal/cli/agent_lint_test.go`):

```bash
# In test function:
# 1. Set up two temp directories with intentional divergence.
# 2. Run lint with --path both.
# 3. Assert violations include drift warning.
```

**Edge cases**:
- File present in template only (not local, or vice-versa) → emits
  `LINT_TREE_FILE_MISMATCH` warning instead of `LINT_TREE_DRIFT`.
- Both trees identical (the post-`make build` steady state) → no drift
  warning.

---

## Definition of Done

The SPEC is "done" (audit-ready for plan-auditor) when:

- [ ] All 14 ACs (AC-V3R2-ORC-002-01 .. AC-V3R2-ORC-002-14) PASS per their
      verification commands.
- [ ] `go test ./internal/cli/... -run TestAgentLint -race` zero failures
      / zero races.
- [ ] Coverage on `internal/cli/agent_lint*.go` ≥ 85%.
- [ ] `make build` exit 0; `make test` exit 0; `golangci-lint run` exit 0.
- [ ] `diff -r` template↔local empty for both
      `.claude/agents/moai/` and `.claude/rules/moai/core/`.
- [ ] No remaining duplicate Skeptical block in any agent file (only
      canonical location in `agent-common-protocol.md`).
- [ ] CI workflow `ci.yaml` includes the `moai agent lint` step.
- [ ] All `@MX:TODO` from M2-M4 resolved on actual code lines.
- [ ] PR opened with plan-auditor review request.

---

End of acceptance.md.
