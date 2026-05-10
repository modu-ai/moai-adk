# SPEC-V3R2-ORC-004 Acceptance Criteria

> Detailed Given-When-Then scenarios for **Worktree MUST Rule for write-heavy role profiles**.
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `research.md` v0.1.0, `tasks.md` v0.1.0.
> Each AC corresponds to one of the 10 acceptance criteria summarized in `spec.md` §6.

## HISTORY

| Version | Date       | Author        | Description |
|---------|------------|---------------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec  | Initial Given-When-Then scenarios. Each AC has primary + edge case + sentinel-key verification scenarios. |

---

## Acceptance Overview

10 acceptance criteria mapped from spec.md §6:

| AC | Title | Primary REQ |
|----|-------|-------------|
| AC-V3R2-ORC-004-01 | Rule text uplift to MUST | REQ-001 |
| AC-V3R2-ORC-004-02 | 5 v3r2 agents declare `isolation: worktree` | REQ-002 |
| AC-V3R2-ORC-004-03 | Read-only agents do NOT declare `isolation: worktree` | REQ-003 |
| AC-V3R2-ORC-004-04 | workflow.yaml role_profiles preserved | REQ-004 |
| AC-V3R2-ORC-004-05 | Lint clean on post-edit roster | REQ-006, REQ-009 |
| AC-V3R2-ORC-004-06 | LR-05 fires with `ORC_WORKTREE_MISSING` sentinel | REQ-013 |
| AC-V3R2-ORC-004-07 | LR-09 fires with `ORC_WORKTREE_ON_READONLY` sentinel | REQ-014 |
| AC-V3R2-ORC-004-08 | Researcher body line definitive | REQ-011 |
| AC-V3R2-ORC-004-09 | `moai workflow lint` rejects implementer.isolation=none with `ORC_WORKTREE_REQUIRED` | REQ-008 |
| AC-V3R2-ORC-004-10 | Skill body absolute-path discipline preserved | REQ-010, REQ-015 |

Definition of Done at the bottom.

---

## AC-V3R2-ORC-004-01: Rule text uplift to MUST

**Source**: spec.md §6 AC-ORC-004-01, REQ-001.

### Primary Scenario

**Given** the file `.claude/rules/moai/workflow/worktree-integration.md` exists in the post-merge `main` HEAD.

**When** an automated reader greps for the §HARD Rules section content.

**Then** the file MUST contain a clause matching the following normative text (modulo whitespace):

> Implementation agents that write files across 3 or more paths per invocation MUST use `isolation: worktree` in their frontmatter. This includes v3r2 agents manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher, and team-mode role profiles implementer, tester, designer.

**And** the previous SHOULD-style line at line 135 MUST no longer be present.

### Verification Command

```bash
grep -n "MUST use \`isolation: worktree\` in their frontmatter" \
  .claude/rules/moai/workflow/worktree-integration.md
# Expected: at least one match
grep -c "SHOULD use \`isolation: \"worktree\"\` when making cross-file changes" \
  .claude/rules/moai/workflow/worktree-integration.md
# Expected: 0
```

### Sentinel Key Documentation

The file MUST also document the three `ORC_WORKTREE_*` sentinel keys in a "Sentinel Key Glossary" subsection:

```bash
grep -E "ORC_WORKTREE_(REQUIRED|MISSING|ON_READONLY)" \
  .claude/rules/moai/workflow/worktree-integration.md
# Expected: at least 3 matches
```

### Edge Case: Template Mirror

The same content must exist in `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` (template-first parity). Verify with:

```bash
diff .claude/rules/moai/workflow/worktree-integration.md \
     internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
# Expected: no diff
```

### Pass Conditions

1. Greps above succeed with expected counts.
2. Template parity diff is empty.
3. The ALL-CAPS keyword `MUST` (not `should`) appears in the new clause.

---

## AC-V3R2-ORC-004-02: 5 v3r2 agents declare `isolation: worktree`

**Source**: spec.md §6 AC-ORC-004-02, REQ-002.

### Primary Scenario

**Given** the post-edit `.claude/agents/moai/` directory.

**When** an automated reader inspects the YAML frontmatter of each of the 5 v3r2 write-heavy agents.

**Then** each MUST declare `isolation: worktree` as a top-level key.

### Verification Command

```bash
for agent in manager-cycle expert-backend expert-frontend expert-refactoring researcher; do
  if grep -qE "^isolation: worktree$" ".claude/agents/moai/${agent}.md"; then
    echo "OK: $agent"
  else
    echo "MISSING: $agent"
  fi
done
# Expected: 5 OK lines
```

### Conditional Path: manager-cycle dependency

If SPEC-V3R2-ORC-001 has not yet merged at run-phase start, the manager-cycle file does not exist. In that case, the verification reports 4 OK + 1 MISSING and the implementer pauses with a structured blocker report `ORC_DEPENDENCY_MISSING`.

The blocker resolution is to:

1. Wait for ORC-001 to merge.
2. Pull origin/main into the SPEC worktree (`git pull --rebase origin main`).
3. Re-run T-ORC004-09 (manager-cycle frontmatter add).
4. Re-verify AC-02.

### Edge Case: YAML Position

The `isolation: worktree` key MUST appear inside the YAML frontmatter (between the two `---` delimiters), not in the body. Verify by checking line range:

```bash
for agent in manager-cycle expert-backend expert-frontend expert-refactoring researcher; do
  python3 -c "
import sys
content = open('.claude/agents/moai/${agent}.md').read()
parts = content.split('---', 2)
if len(parts) < 3: sys.exit(1)
fm = parts[1]
if 'isolation: worktree' not in fm:
    print('NOT IN FRONTMATTER: ${agent}'); sys.exit(1)
print('OK: ${agent}')
"
done
```

### Pass Conditions

1. All 5 agents declare `isolation: worktree` (or 4 + conditional 1 with documented blocker).
2. The declaration is in YAML frontmatter, not body text.
3. Template-side parity holds (same key in `internal/template/templates/.claude/agents/moai/*.md`).

---

## AC-V3R2-ORC-004-03: Read-only agents do NOT declare `isolation: worktree`

**Source**: spec.md §6 AC-ORC-004-03, REQ-003.

### Primary Scenario

**Given** the post-edit `.claude/agents/moai/` directory.

**When** an automated reader inspects the frontmatter of read-only agents.

**Then** none of {manager-strategy, manager-quality, evaluator-active, plan-auditor} MUST declare `isolation: worktree`.

### Verification Command

```bash
for agent in manager-strategy manager-quality evaluator-active plan-auditor; do
  if grep -qE "^isolation: worktree$" ".claude/agents/moai/${agent}.md"; then
    echo "UNEXPECTED: $agent has isolation: worktree"
  else
    echo "OK: $agent"
  fi
done
# Expected: 4 OK lines, 0 UNEXPECTED
```

### Edge Case: Lint Cross-Verification

`moai agent lint` must NOT report LR-05 errors for these 4 agents. (LR-05's substring match would not match these names anyway, but verification is mandatory.)

```bash
moai agent lint --format json | jq '.violations[] | select(.rule == "LR-05") | .file' | grep -E "(manager-strategy|manager-quality|evaluator-active|plan-auditor)"
# Expected: no output
```

### Pass Conditions

1. All 4 read-only agents lack `isolation: worktree` in frontmatter.
2. `moai agent lint` does not flag any of them with LR-05 or LR-09.

---

## AC-V3R2-ORC-004-04: workflow.yaml role_profiles preserved

**Source**: spec.md §6 AC-ORC-004-04, REQ-004.

### Primary Scenario

**Given** the file `.moai/config/sections/workflow.yaml` in post-edit state.

**When** an automated reader parses the `workflow.team.role_profiles` map.

**Then** exactly 3 role profiles MUST have `isolation: worktree`: `implementer`, `tester`, `designer`. The other 4 (researcher, analyst, architect, reviewer) MUST have `isolation: none`.

### Verification Command

```bash
yq '.workflow.team.role_profiles | to_entries[] | "\(.key)=\(.value.isolation)"' \
  .moai/config/sections/workflow.yaml | sort
# Expected output (sorted alphabetically by role name):
# analyst=none
# architect=none
# designer=worktree
# implementer=worktree
# researcher=none
# reviewer=none
# tester=worktree
```

### Edge Case: Template Parity

The template-side `internal/template/templates/.moai/config/sections/workflow.yaml` MUST match the same role_profile assignments. Verify:

```bash
diff <(yq '.workflow.team.role_profiles' .moai/config/sections/workflow.yaml) \
     <(yq '.workflow.team.role_profiles' internal/template/templates/.moai/config/sections/workflow.yaml)
# Expected: no diff (or whitespace-only diff)
```

### Pass Conditions

1. Exactly 3 of 7 role profiles have `isolation: worktree`.
2. The 3 are precisely `implementer`, `tester`, `designer`.
3. No additional role profiles introduced.
4. Template parity holds.

---

## AC-V3R2-ORC-004-05: Lint clean on post-edit roster

**Source**: spec.md §6 AC-ORC-004-05, REQ-006, REQ-009.

### Primary Scenario

**Given** the full post-edit agent roster + workflow.yaml state.

**When** the operator runs `moai agent lint` from the project root.

**Then** the command MUST exit with code 0 and produce no LR-05 / LR-09 violations.

### Verification Command

```bash
moai agent lint --format json | jq '.summary'
# Expected:
# {
#   "total": 0,
#   "errors": 0,
#   "warnings": <some_count>  # LR-06 deepthink boilerplate may still produce warnings
# }
echo "Exit: $?"
# Expected: 0
```

If LR-06 deepthink-boilerplate warnings exist (pre-existing in some agents), they remain at warning severity and do not affect exit code in non-strict mode.

### Edge Case: Strict Mode

```bash
moai agent lint --strict --format json | jq '.summary.errors'
# In strict mode, LR-06 warnings may promote to errors. AC-05 only requires
# zero LR-05 / LR-09 violations; LR-06 is independently tracked by ORC-002.
```

The strict-mode exit code is OUT OF SCOPE for AC-05 — only LR-05 / LR-09 cleanliness is required.

### Pass Conditions

1. `moai agent lint` exits 0 (default mode).
2. JSON output `.violations[] | select(.rule == "LR-05")` returns empty.
3. JSON output `.violations[] | select(.rule == "LR-09")` returns empty.

---

## AC-V3R2-ORC-004-06: LR-05 fires with `ORC_WORKTREE_MISSING` sentinel

**Source**: spec.md §6 AC-ORC-004-06, REQ-013.

### Primary Scenario

**Given** the post-edit agent roster.

**When** an operator stashes the changes in `expert-backend.md` (removing `isolation: worktree`) and re-runs `moai agent lint`.

**Then** the command MUST exit with code 1 and emit a violation containing the sentinel key `ORC_WORKTREE_MISSING`.

### Verification Command

```bash
# Setup: temporarily revert isolation: worktree from expert-backend
sed -i.bak '/^isolation: worktree$/d' .claude/agents/moai/expert-backend.md
# Verify: LR-05 fires with sentinel
moai agent lint --format json | jq '.violations[] | select(.rule == "LR-05") | .message'
# Expected: at least one string containing "ORC_WORKTREE_MISSING"
echo "Exit: $?"
# Expected: 1

# Cleanup: restore isolation: worktree from backup
mv .claude/agents/moai/expert-backend.md.bak .claude/agents/moai/expert-backend.md
moai agent lint --format json | jq '.summary.errors'
# Expected: 0 (back to clean)
```

### Edge Case: Sentinel Key Format

The sentinel MUST be the exact string `ORC_WORKTREE_MISSING` (uppercase, underscore-separated). Verify:

```bash
moai agent lint --format json | jq -r '.violations[] | select(.rule == "LR-05") | .message' | grep -oE 'ORC_WORKTREE_MISSING' | head -1
# Expected: ORC_WORKTREE_MISSING (exact match)
```

### Edge Case: Test Fixture

The unit test `TestLintLR05_OrcWorktreeMissingSentinel` (added in T-ORC004-10) MUST pass:

```bash
go test -run TestLintLR05_OrcWorktreeMissingSentinel ./internal/cli/...
# Expected: PASS
```

### Pass Conditions

1. After removing `isolation: worktree` from expert-backend, `moai agent lint` exits 1.
2. The LR-05 violation message contains the literal substring `ORC_WORKTREE_MISSING`.
3. The unit test fixture passes.
4. After restoring, exit returns to 0.

---

## AC-V3R2-ORC-004-07: LR-09 fires with `ORC_WORKTREE_ON_READONLY` sentinel

**Source**: spec.md §6 AC-ORC-004-07, REQ-014.

### Primary Scenario

**Given** the post-edit agent roster.

**When** an operator manually adds `isolation: worktree` to `evaluator-active.md` (which has `permissionMode: plan`) and re-runs `moai agent lint`.

**Then** the command MUST exit with code 1 and emit an LR-09 violation containing the sentinel key `ORC_WORKTREE_ON_READONLY`.

### Verification Command

```bash
# Setup: inject isolation: worktree into evaluator-active (a read-only agent)
python3 - <<'EOF'
import re
with open('.claude/agents/moai/evaluator-active.md') as f:
    content = f.read()
# Insert after first --- line
new = re.sub(r'(---\n)', r'\1isolation: worktree\n', content, count=1)
with open('.claude/agents/moai/evaluator-active.md.test', 'w') as f:
    f.write(new)
EOF
mv .claude/agents/moai/evaluator-active.md .claude/agents/moai/evaluator-active.md.bak
mv .claude/agents/moai/evaluator-active.md.test .claude/agents/moai/evaluator-active.md

# Verify: LR-09 fires with sentinel
moai agent lint --format json | jq '.violations[] | select(.rule == "LR-09") | .message'
# Expected: at least one string containing "ORC_WORKTREE_ON_READONLY"
echo "Exit: $?"
# Expected: 1

# Cleanup
mv .claude/agents/moai/evaluator-active.md.bak .claude/agents/moai/evaluator-active.md
```

### Edge Case: Sentinel Key Format

```bash
moai agent lint --format json | jq -r '.violations[] | select(.rule == "LR-09") | .message' | grep -oE 'ORC_WORKTREE_ON_READONLY' | head -1
# Expected: ORC_WORKTREE_ON_READONLY (exact match)
```

### Edge Case: Test Fixture

The unit test `TestLintLR09_OrcWorktreeOnReadonlySentinel` (added in T-ORC004-11) MUST pass:

```bash
go test -run TestLintLR09_OrcWorktreeOnReadonlySentinel ./internal/cli/...
# Expected: PASS
```

### Pass Conditions

1. After injecting `isolation: worktree` into evaluator-active, `moai agent lint` exits 1.
2. The LR-09 violation message contains the literal substring `ORC_WORKTREE_ON_READONLY`.
3. The unit test fixture passes.

---

## AC-V3R2-ORC-004-08: Researcher body line definitive

**Source**: spec.md §6 AC-ORC-004-08, REQ-011.

### Primary Scenario

**Given** the post-edit `researcher.md` file.

**When** an automated reader inspects the body section for the line previously reading "All experiments in worktree isolation when possible".

**Then** the file MUST NO LONGER contain "when possible" qualifier; the line MUST read substantively "All experiments in worktree isolation (mandatory per SPEC-V3R2-ORC-004)" or equivalent definitive phrasing.

### Verification Command

```bash
grep -E "All experiments in worktree isolation" .claude/agents/moai/researcher.md
# Expected: 1+ matches with definitive phrasing (no "when possible")

grep "when possible" .claude/agents/moai/researcher.md
# Expected: 0 matches (in the worktree-isolation context — other unrelated "when possible" lines may exist)
```

### Edge Case: P-A22 Reconciliation

The frontmatter (`isolation: worktree`) and the body line MUST be consistent. Pre-edit state had:

- Frontmatter: NO `isolation:` key
- Body: "when possible"

Post-edit state MUST have:

- Frontmatter: `isolation: worktree`
- Body: definitive statement aligning with frontmatter

Verify both:

```bash
grep -E "^isolation: worktree$" .claude/agents/moai/researcher.md && \
grep -E "All experiments in worktree isolation .* (mandatory|always|invariant)" .claude/agents/moai/researcher.md
# Expected: both matches succeed
```

### Pass Conditions

1. `researcher.md` body no longer contains "when possible" in the worktree-isolation paragraph.
2. The replacement phrasing references SPEC-V3R2-ORC-004 (or equivalent strong language).
3. Frontmatter and body are mutually consistent.

---

## AC-V3R2-ORC-004-09: `moai workflow lint` rejects implementer.isolation=none with `ORC_WORKTREE_REQUIRED`

**Source**: spec.md §6 AC-ORC-004-09, REQ-008.

### Primary Scenario

**Given** the new `moai workflow lint` CLI subcommand and an injected workflow.yaml with `implementer.isolation = none`.

**When** the operator runs `moai workflow lint` against the modified file.

**Then** the command MUST exit with code 1 and emit a violation containing the sentinel key `ORC_WORKTREE_REQUIRED`.

### Verification Command

```bash
# Setup: temporarily mutate workflow.yaml
cp .moai/config/sections/workflow.yaml /tmp/workflow.yaml.bak
yq -i '.workflow.team.role_profiles.implementer.isolation = "none"' .moai/config/sections/workflow.yaml

# Verify: ORC_WORKTREE_REQUIRED sentinel fires
moai workflow lint --format json | jq '.violations[] | select(.rule == "ORC_WORKTREE_REQUIRED")'
# Expected: at least one entry with the rule field exactly equal to "ORC_WORKTREE_REQUIRED"
echo "Exit: $?"
# Expected: 1

# Cleanup
cp /tmp/workflow.yaml.bak .moai/config/sections/workflow.yaml
moai workflow lint --format json | jq '.summary.errors'
# Expected: 0 (back to clean)
```

### Edge Case: tester.isolation = none

```bash
yq -i '.workflow.team.role_profiles.tester.isolation = "none"' .moai/config/sections/workflow.yaml
moai workflow lint --format json | jq '.violations[] | select(.path == "workflow.team.role_profiles.tester.isolation")'
# Expected: violation entry with rule = ORC_WORKTREE_REQUIRED
```

### Edge Case: designer.isolation = none

```bash
yq -i '.workflow.team.role_profiles.designer.isolation = "none"' .moai/config/sections/workflow.yaml
moai workflow lint --format json | jq '.violations[] | select(.path == "workflow.team.role_profiles.designer.isolation")'
# Expected: violation entry with rule = ORC_WORKTREE_REQUIRED
```

### Edge Case: Read-only roles unaffected

Setting `researcher.isolation` from `none` to `worktree` MUST NOT trigger `ORC_WORKTREE_REQUIRED` (it's a different concern; `moai workflow lint` does not enforce `isolation: none` on read-only roles in the current ORC-004 scope. That enforcement could be a future LR for workflow.yaml).

```bash
yq -i '.workflow.team.role_profiles.researcher.isolation = "worktree"' .moai/config/sections/workflow.yaml
moai workflow lint --format json | jq '.violations[] | select(.rule == "ORC_WORKTREE_REQUIRED")'
# Expected: no entries (researcher is not in the enforcement set for ORC_WORKTREE_REQUIRED)
```

### Edge Case: Test Fixtures

```bash
go test -run "TestWorkflowLint_OrcWorktreeRequiredOn(Implementer|Tester|Designer)" ./internal/cli/...
# Expected: 3 PASS lines
go test -run "TestWorkflowLint_NoViolationsOnCorrectConfig" ./internal/cli/...
# Expected: 1 PASS line
```

### Pass Conditions

1. `moai workflow lint` exits 1 when any of {implementer, tester, designer} has `isolation != "worktree"`.
2. The violation `rule` field is exactly `ORC_WORKTREE_REQUIRED`.
3. The violation `path` field correctly identifies the YAML key.
4. All 4 unit test fixtures pass.

---

## AC-V3R2-ORC-004-10: Skill body absolute-path discipline preserved

**Source**: spec.md §6 AC-ORC-004-10, REQ-010, REQ-015.

### Primary Scenario

**Given** the existing `.claude/skills/moai/team/run.md` (orchestrator skill body for team mode).

**When** a code reviewer inspects the body for guidance about spawning worktree-isolated agents.

**Then** the body MUST contain language reinforcing relative-path discipline (no `cd /absolute/path && ...` prefixes in spawn prompts) AND reference `worktree-integration.md` §Prompt Path Rules.

### Verification Command

```bash
grep -E "(absolute path|relative path|/absolute/path|cd .*&&|Prompt Path Rules)" \
  .claude/skills/moai/team/run.md
# Expected: at least 1 match referencing path discipline

grep "Prompt Path Rules" .claude/skills/moai/team/run.md || \
  grep "worktree-integration" .claude/skills/moai/team/run.md
# Expected: at least one cross-reference
```

### Edge Case: Cross-reference Existence

The cross-reference target `.claude/rules/moai/workflow/worktree-integration.md` §Prompt Path Rules MUST exist:

```bash
grep -i "prompt path rules" .claude/rules/moai/workflow/worktree-integration.md
# Expected: at least one heading or section reference
```

### Edge Case: No New Rule Text

ORC-004 does NOT introduce new path-discipline rule text. The acceptance is verifying that the existing rule remains intact and the cross-reference is functional.

### Pass Conditions

1. `team/run.md` references absolute-path discipline OR cross-references worktree-integration.md §Prompt Path Rules.
2. The cross-referenced section exists in worktree-integration.md.
3. No regression: pre-existing guidance is preserved.

---

## Definition of Done (DoD)

The SPEC implementation is **DONE** when ALL of the following criteria are simultaneously satisfied:

### Code-Level DoD

- [ ] All 5 v3r2 write-heavy agents declare `isolation: worktree` in YAML frontmatter (or 4 + documented blocker).
- [ ] `worktree-integration.md` §HARD Rules contains the new MUST clause from REQ-001.
- [ ] `worktree-integration.md` documents `ORC_WORKTREE_*` sentinel keys.
- [ ] `agent_lint.go` LR-05 violation message includes `ORC_WORKTREE_MISSING` sentinel.
- [ ] `agent_lint.go` LR-09 violation message includes `ORC_WORKTREE_ON_READONLY` sentinel.
- [ ] `internal/cli/workflow_lint.go` exists and compiles.
- [ ] `internal/cli/workflow_lint_test.go` exists with ≥4 passing tests.
- [ ] `agent_lint_test.go` extended with 2 new sentinel-verification tests.
- [ ] `internal/cli/sentinels.go` (or equivalent) hosts shared sentinel constants.
- [ ] `researcher.md` body line "All experiments in worktree isolation when possible" replaced with definitive phrasing.

### Lint & Test Gate DoD

- [ ] `moai agent lint` exits 0 against post-edit roster (AC-05).
- [ ] `moai workflow lint` exits 0 against current workflow.yaml (AC-09 baseline).
- [ ] `make test ./...` exits 0 (full Go test suite green).
- [ ] `go vet ./...` exits 0.
- [ ] `golangci-lint run` exits 0.
- [ ] LR-05 sentinel verification fixture (T-ORC004-10) passes.
- [ ] LR-09 sentinel verification fixture (T-ORC004-11) passes.
- [ ] All 4 workflow_lint_test fixtures (T-ORC004-14) pass.
- [ ] AC-06 manual injection scenario reproduces sentinel (verified once during sync).
- [ ] AC-07 manual injection scenario reproduces sentinel (verified once during sync).
- [ ] AC-09 manual injection scenarios for implementer/tester/designer reproduce sentinel.

### Template-First Mirror DoD

- [ ] `make build` succeeds; `internal/template/embedded.go` regenerates without error.
- [ ] `make install` succeeds; new `moai` binary at `~/go/bin/moai`.
- [ ] `moai update` (run from worktree) mirrors templates to `.claude/`.
- [ ] `diff -r .claude/agents/moai/ internal/template/templates/.claude/agents/moai/` returns zero diff (excluding allowlist).
- [ ] `diff -u .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` returns zero diff.

### Documentation DoD

- [ ] `CHANGELOG.md` Unreleased entry added.
- [ ] MX tags added per plan.md §6 (5 annotations across 4 files).
- [ ] sync phase produces release notes draft with migration guidance.
- [ ] PR body references SPEC-V3R2-ORC-004 issue (if created) and lists modified files.

### Pre-Submission Self-Review DoD

- [ ] Full diff reviewed against acceptance.md scenarios.
- [ ] "Is there a simpler approach?" — answer documented as: NO. Frontmatter additions are mechanical; new CLI is minimal.
- [ ] "Would removing any changes still satisfy the SPEC?" — answer documented as: NO. Each delivered change maps to a specific REQ/AC.
- [ ] Drift check: actual modified file count vs plan.md §1.3 deliverables — drift ≤ 30%.

### Plan-Auditor Gate DoD

- [ ] plan-auditor verdict: PASS (target score ≥ 0.85 first iteration).
- [ ] If REVISE: defects addressed via in-place plan amendment (not new SPEC).
- [ ] Final plan-auditor PASS recorded in `.moai/reports/plan-audit/SPEC-V3R2-ORC-004-<DATE>.md`.

### Cross-SPEC DoD

- [ ] No regression on SPEC-V3R2-ORC-001 (manager-cycle) functionality.
- [ ] No regression on SPEC-V3R2-ORC-002 (LR-05 origin) tests.
- [ ] SPEC-V3R2-ORC-005 (dynamic team) impact: documented if workflow_lint type collides with future workflow.yaml schema changes.

### Post-Merge DoD

- [ ] sync PR opens with codemap regeneration if applicable.
- [ ] Memory entry written: `project_orc004_complete.md` per session-handoff.md auto-memory protocol.
- [ ] Worktree disposed via `moai worktree done SPEC-V3R2-ORC-004` after BOTH run AND sync PRs merged.

---

End of acceptance.
