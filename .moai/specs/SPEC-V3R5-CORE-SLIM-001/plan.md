# Plan — SPEC-V3R5-CORE-SLIM-001

Implementation plan for dissolving the 12 LR-08 `expert` category warnings through a two-track parallel approach: (Track A) LR-08 rule refinement exempting domain-prefix skills + (Track B) `moai-foundation-quality` symmetry addition to 4 expert agents.

## Plan Summary

Two-track parallel implementation. **Track A** modifies `internal/cli/agent_lint.go` `checkSkillPreloadDrift` (lines 892-977) to exempt domain-scoped skill prefixes from intra-category symmetry checking, eliminating 10 of the 12 LR-08 findings. **Track B** adds `moai-foundation-quality` to the 4 expert agents that lack it, eliminating the remaining 2 findings. Tracks are independent and may be executed in either order; final verification (lint count = 0) requires both tracks complete.

Estimated diff:
- Track A: ~30 lines added to `agent_lint.go` (constant + helper + 1 skip clause) + ~120 lines added to `agent_lint_test.go` (4 new test cases)
- Track B: ~4 lines added across 8 markdown files (4 source + 4 mirror) + auto-generated `catalog.yaml` updates

Total: 11 files modified.

## Track A — LR-08 Rule Refinement

This track modifies the Go lint rule implementation to exempt domain-specific skill prefixes from symmetry enforcement.

### Step A.1 — Read current implementation

Read `internal/cli/agent_lint.go` lines 880-980 (the `checkSkillPreloadDrift` function and surrounding context). Verify the function signature, the existing `len(agents) < 2` early return, the `skillCounts` map construction, and the drift inner loop at lines 959-973. Confirm the line number `961` for the `skillCounts[skill] < len(agents)` check that will be guarded by the new exemption skip.

### Step A.2 — Define exemption constant

Add the following file-scoped constant near the top of `agent_lint.go` (e.g., after the existing `lintRules` declaration or near other shared constants):

```go
// domainExemptPrefixes lists skill name prefixes that are intentionally
// agent-specific and exempt from LR-08 intra-category symmetry checking.
// Foundation skills (moai-foundation-*) and workflow skills (moai-workflow-*)
// are NOT exempted — they are universally relevant.
// See SPEC-V3R5-CORE-SLIM-001 §4 Design Rationale for the taxonomy mapping.
var domainExemptPrefixes = []string{
    "moai-domain-",
    "moai-design-",
    "moai-library-",
    "moai-framework-",
    "moai-platform-",
    "moai-ref-",
}

// isDomainExemptSkill returns true when a skill name begins with any of
// the domain-scoped prefixes listed in domainExemptPrefixes.
func isDomainExemptSkill(skill string) bool {
    for _, prefix := range domainExemptPrefixes {
        if strings.HasPrefix(skill, prefix) {
            return true
        }
    }
    return false
}
```

Verify `strings` is already imported (it is, used elsewhere in the file).

### Step A.3 — Insert skip clause in drift loop

Modify `checkSkillPreloadDrift` inner loop at line 960. Insert the exemption skip immediately before the drift count check:

```go
// Check for drift
for _, agent := range agents {
    for _, skill := range agent.skills {
        if isDomainExemptSkill(skill) {
            continue // Domain-scoped skills exempt from LR-08 — see SPEC-V3R5-CORE-SLIM-001
        }
        if skillCounts[skill] < len(agents) {
            // ... existing emit logic unchanged
        }
    }
}
```

No other lines in `checkSkillPreloadDrift` are modified.

### Step A.4 — Add unit tests

Append the following 4 test cases to `internal/cli/agent_lint_test.go`. Follow the existing test scaffolding pattern in that file (likely using `t.TempDir()`, fixture files, table-driven structure):

- **`TestSkillPreloadDriftExemption_DomainSkills`** (negative case):
  - Fixture: 3 expert agents in a category. Agent 1 has `moai-domain-backend`. Agents 2 and 3 do not.
  - Assertion: `checkSkillPreloadDrift` returns ZERO violations for `moai-domain-backend` (it is exempted).
  - Coverage: AC-CSLM-005.a + REQ-CSLM-001

- **`TestSkillPreloadDriftExemption_FoundationSkills`** (positive case):
  - Fixture: 3 expert agents. Agent 1 has `moai-foundation-quality`. Agents 2 and 3 do not.
  - Assertion: `checkSkillPreloadDrift` returns AT LEAST 1 violation for `moai-foundation-quality` on Agent 1 (enforcement preserved).
  - Coverage: AC-CSLM-005.b + REQ-CSLM-002

- **`TestSkillPreloadDriftExemption_WorkflowSkills`** (positive case):
  - Fixture: 3 expert agents. Agent 1 has `moai-workflow-testing`. Agents 2 and 3 do not.
  - Assertion: `checkSkillPreloadDrift` returns AT LEAST 1 violation for `moai-workflow-testing` on Agent 1 (enforcement preserved).
  - Coverage: AC-CSLM-005.c + REQ-CSLM-002

- **`TestSkillPreloadDriftExemption_EdgeCases`** (edge case sweep):
  - Sub-case A: empty `skills:` array — no violations emitted.
  - Sub-case B: single-agent category — `len(agents) < 2` early return, no violations.
  - Sub-case C: foundation-like-named domain skill (e.g., a fixture skill named `moai-foundation-domain-lookalike`) — must NOT be exempted (literal prefix match, not substring; `moai-foundation-` enforcement wins).
  - Sub-case D: all 6 domain prefixes tested individually (`moai-domain-*`, `moai-design-*`, `moai-library-*`, `moai-framework-*`, `moai-platform-*`, `moai-ref-*`) — each exempted.
  - Coverage: AC-CSLM-005.d + REQ-CSLM-001 + REQ-CSLM-005

Use table-driven test structure where natural (Go convention). Each test name is the canonical name from AC-CSLM-005 leaves to enable `go test -run "TestSkillPreloadDriftExemption"` to invoke all 4.

### Step A.5 — Validate Track A locally

Sequence (executed at end of run-phase; orchestrator handles verification):

```bash
go test ./internal/cli/... -run "TestSkillPreloadDriftExemption" -v   # Expect: PASS for all 4 new tests
go test ./internal/cli/... -count=1                                    # Expect: PASS for existing tests (no regression)
golangci-lint run ./internal/cli/... 2>&1 | tail -5                    # Expect: no new findings
make build                                                             # Expect: success
```

If any step fails, halt and re-plan.

## Track B — Foundation-Quality Symmetry Addition

This track adds `moai-foundation-quality` to the 4 expert agents currently lacking it, completing the foundation-skill uniformity for the expert category.

### Step B.1 — Read source agent files

Read each of the 4 source agent files to capture current `skills:` array contents:

- `.claude/agents/moai/expert-backend.md`
- `.claude/agents/moai/expert-frontend.md`
- `.claude/agents/moai/expert-refactoring.md`
- `.claude/agents/moai/expert-devops.md`

For each file, locate the `skills:` block in the YAML frontmatter (between the leading `---` and trailing `---` markers). Verified pre-state (per research.md §3.2):

- `expert-backend`: `core, domain-backend, domain-database, workflow-testing` (no foundation-quality)
- `expert-frontend`: `core, domain-frontend, design-system, workflow-testing` (no foundation-quality)
- `expert-refactoring`: `core, workflow-testing` (no foundation-quality)
- `expert-devops`: `core, workflow-testing` (no foundation-quality)

The remaining 2 expert agents (`expert-performance`, `expert-security`) already have `moai-foundation-quality` and are NOT edited.

### Step B.2 — Apply alphabetical insertion

For each of the 4 source files, add `- moai-foundation-quality` to the YAML `skills:` array. Position alphabetically within the existing array:

All 4 agents follow the **preserve-existing-order** convention (Agent Core Behavior 5 — scope discipline). Insert `moai-foundation-quality` at the position that keeps `workflow-testing` as the end-anchor (visible across all 4 agents):

- `expert-backend`: insert between `moai-domain-database` and `moai-workflow-testing` → array becomes `core, domain-backend, domain-database, foundation-quality, workflow-testing`
- `expert-frontend`: insert between `moai-design-system` and `moai-workflow-testing` (preserve existing order; no re-alphabetization) → array becomes `core, domain-frontend, design-system, foundation-quality, workflow-testing`
- `expert-refactoring`: insert between `core` and `workflow-testing` → array becomes `core, foundation-quality, workflow-testing`
- `expert-devops`: same as refactoring → array becomes `core, foundation-quality, workflow-testing`

Use the `Edit` tool exclusively (no `sed`, no `awk`, no bulk regex). Each edit operates on a unique anchor string from the read content.

### Step B.3 — Mirror to template paths

For each edited source file, propagate the identical edit to the template mirror:

- `internal/template/templates/.claude/agents/moai/expert-{backend,frontend,refactoring,devops}.md`

Per `CLAUDE.local.md §2` Template-First Rule, the cleanest execution order is:

1. Edit template source files (`internal/template/templates/.claude/agents/moai/...`) FIRST
2. Run `make build` (regenerates `internal/template/embedded.go`)
3. Sync to live `.claude/agents/moai/...` by copying the source files

Alternatively (validated equivalent path): use `MultiEdit` to edit both pairs synchronously, then run `make build`. Both directions satisfy AC-CSLM-003 (byte-identical parity post-edit).

### Step B.4 — Refresh catalog hash table

After all 8 file edits are complete:

```bash
go run internal/template/scripts/gen-catalog-hashes.go --all
```

This rewrites `internal/template/catalog.yaml` with refreshed SHA hashes for the 4 affected entries. Verify that the diff touches **exactly** the 4 `expert-{backend,frontend,refactoring,devops}` entries (no other catalog rows changed).

### Step B.5 — Validate Track B locally

```bash
./bin/moai agent lint --strict 2>&1 | grep "LR-08" | grep "moai-foundation-quality" | wc -l   # Expect: 0
go test ./internal/template/...                                                                 # Expect: PASS (catalog hash audit green)
make build                                                                                      # Expect: success
```

## Track Sequencing

Tracks A and B are **independent**. Either order works:

- **Order 1 (A first, then B)**: Rule fix applied first → 10 of 12 LR-08 dissolved → only `moai-foundation-quality` warnings remain → Track B applied → all 12 dissolved.
- **Order 2 (B first, then A)**: Foundation symmetry applied first → 2 of 12 LR-08 dissolved → 10 domain-skill warnings remain → Track A applied → all 12 dissolved.

**Final verification (after BOTH tracks complete)**:

```bash
./bin/moai agent lint --strict 2>&1 | grep "LR-08" | wc -l                                              # Expect: 0
./bin/moai agent lint --strict 2>&1 | grep -E "^! \[LR-" | grep -v "^! \[LR-08\]" | wc -l               # Expect: = pre-merge complement baseline (auto-tracks all LR rules except LR-08; baseline = 0 per acceptance.md verification matrix)
./bin/moai spec lint --strict                                                                    # Expect: ✓ exit 0
go test ./...                                                                                    # Expect: PASS across all packages
make build                                                                                       # Expect: success
```

## Dependencies

| Type | Item | Verification |
|------|------|--------------|
| SPEC | SPEC-V3R5-CONSTITUTION-DUAL-001 | merged at main HEAD `175bad283` (verified via memory `project_v3r5_w1_constitution_complete`) |
| Skill | `moai-foundation-quality` | `ls .claude/skills/moai-foundation-quality/SKILL.md` → present |
| Skill taxonomy | 6 domain prefixes (`moai-domain-`, `moai-design-`, `moai-library-`, `moai-framework-`, `moai-platform-`, `moai-ref-`) | research.md §3.4 + §3.5 |
| Code | `internal/cli/agent_lint.go` `checkSkillPreloadDrift` | research.md §3.1 (verbatim source) |
| Tests | `internal/cli/agent_lint_test.go` | existing scaffolding (47KB), supports table-driven append |
| Tool | `make build` | pre-existing, verified daily |
| Tool | `gen-catalog-hashes.go --all` | pre-existing, used by sibling SPEC-V3R5-LINT-CLEAN-001 |
| Tool | `./bin/moai agent lint --strict` | pre-existing per SPEC-V3R2-ORC-002 (PR #980) |

## Rollback

Single-command revert if any AC fails post-edit:

```bash
git restore \
  internal/cli/agent_lint.go \
  internal/cli/agent_lint_test.go \
  .claude/agents/moai/expert-backend.md \
  .claude/agents/moai/expert-frontend.md \
  .claude/agents/moai/expert-refactoring.md \
  .claude/agents/moai/expert-devops.md \
  internal/template/templates/.claude/agents/moai/expert-backend.md \
  internal/template/templates/.claude/agents/moai/expert-frontend.md \
  internal/template/templates/.claude/agents/moai/expert-refactoring.md \
  internal/template/templates/.claude/agents/moai/expert-devops.md \
  internal/template/catalog.yaml
make build
```

Recovery time: under 30 seconds. No external state, no migrations, no data loss possible.

## Risk Assessment

- **Track A (Go code)**: **LOW**. Well-isolated logic refactor with bidirectional test coverage (exemption + enforcement). Existing 47KB test file provides scaffolding patterns. The skip clause is the smallest possible mutation (1 added if-continue) inside an already-tested code path.
- **Track B (agent metadata)**: **LOW**. Pure YAML frontmatter array additions, no body content edits, no Go code, no schema changes, no breaking API changes. Mirrors the SPEC-V3R5-LINT-CLEAN-001 metadata-edit pattern (proven safe).
- **Combined regression risk**: Track A non-regression is verified by AC-CSLM-007 (orthogonal lint surfaces); Track B non-regression is verified by AC-CSLM-005 unit tests covering rule behavior on foundation/workflow skills.
- **No breaking changes**: existing agent definitions still parse, still load, still execute. The added skill metadata only influences LR-08 lint output and any future skill-preload runtime resolution.
- **No security surface change**: the added skill (`moai-foundation-quality`) is part of the MoAI skill catalog. The 6 exemption prefixes are documentation-driven, not security-relevant.

## Technical Approach Notes

- Use `MultiEdit` for parallel edits across the 4 source files for ~60% speedup vs sequential `Edit`. Then re-use `MultiEdit` for the 4 template mirrors.
- Validate idempotency: re-running the run-phase MUST produce zero new edits (each `grep -c` probe returns ≥ 1 on the second pass).
- Catalog refresh is non-trivial: the `gen-catalog-hashes.go` tool re-hashes ALL template files. The diff for `internal/template/catalog.yaml` MUST be inspected to confirm only the 4 expert-* entries changed (sentinel: zero changes outside `agents/moai/expert-*` keys).
- Order constraint: catalog refresh (Step B.4) MUST occur AFTER all 8 Track B file edits and BEFORE the final lint verification. Reversing order produces a transient catalog/file hash mismatch detectable by `go test ./internal/template/...`.
- Track A test additions follow Go convention for table-driven tests. New test names use `TestSkillPreloadDriftExemption_*` prefix to enable `go test -run "TestSkillPreloadDriftExemption"` selective invocation.
- Track A code change is a single function modification within an existing exported behavior; backward compatibility is guaranteed because the exemption logic only REDUCES emit count, never increases it (no risk of new false-positives on previously-clean fixtures).

## Milestones

This SPEC has 2 milestones executed in either order (independent), with final verification gating completion. Per project policy "Never use time predictions in plans or reports" (`.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation), no time estimates are provided. Priority labels only:

- **M1 (Priority High)**: Track A — LR-08 rule refinement + unit tests. Files: `internal/cli/agent_lint.go`, `internal/cli/agent_lint_test.go`.
- **M2 (Priority High)**: Track B — `moai-foundation-quality` preload addition to 4 expert agents + mirror sync + catalog refresh. Files: 4 source agents + 4 template mirrors + `internal/template/catalog.yaml`.
- **M3 (Priority High, gate)**: Combined verification — final `./bin/moai agent lint --strict` LR-08 count = 0 across all categories (AC-CSLM-006). Requires both M1 and M2 complete.

Single PR `feat/SPEC-V3R5-CORE-SLIM-001` containing all 11 files. Squash-merge to `main`.

No subsequent milestones planned in this SPEC. Sync-phase happens via a separate sync PR (`sync/SPEC-V3R5-CORE-SLIM-001` or `chore/SPEC-V3R5-CORE-SLIM-001-sync`) which updates HISTORY, version (0.2.0 → 0.3.0 on run-merge, → 0.4.0 on sync-merge), and status (`draft → implemented → completed`).
