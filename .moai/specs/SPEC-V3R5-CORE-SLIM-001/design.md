# Design — SPEC-V3R5-CORE-SLIM-001

Implementation design rationale, prefix taxonomy, code-change pattern, and edge-case handling strategy for the V3R5 W2 LR-08 rule refinement + expert agent foundation skill symmetry.

## Design Rationale

### Why the LR-08 rule needs refinement (not just metadata edits)

The LR-08 rule (`internal/cli/agent_lint.go` `checkSkillPreloadDrift`, lines 892-977) enforces **strict uniform symmetry**: every skill that any agent in a category preloads MUST be preloaded by ALL agents in that category. The implementation is intentionally simple — see research.md §3.1 for the verbatim function source.

This design works perfectly for **universally relevant** skills:
- `moai-foundation-*` (TRUST 5 framework, project context, thinking primitives) — every agent benefits
- `moai-workflow-*` (DDD, TDD, testing, project orchestration) — every agent uses workflow primitives

This design fails for **domain-scoped** skills:
- `moai-domain-backend` (API design, persistence, observability) — relevant only to backend-adjacent agents
- `moai-domain-frontend` (component architecture, accessibility, state management) — relevant only to frontend-adjacent agents
- `moai-design-system` (design tokens, brand visual identity) — relevant only to design-aware agents
- `moai-library-*`, `moai-framework-*`, `moai-platform-*`, `moai-ref-*` — by definition agent-specific

The original v0.1.0 SPEC scope proposed adding domain skills to every expert agent in the category to satisfy the strict symmetry. This was rejected at plan-phase verification (research.md §4 EC-4 Discovery Log) on two grounds:

1. **Empirical**: All matrix domain skills were ALREADY present on the named source agents (`research.md` §3.2). The 4-file additive scope was a no-op.
2. **Semantic**: Adding `moai-domain-backend` to `expert-frontend` would bloat the frontend agent's context window with irrelevant backend domain knowledge. Adding `moai-design-system` to `expert-devops` would inject design-token vocabulary into a deployment-focused agent. The strict uniform symmetry is overzealous for domain-scoped skills.

The correct resolution is to **refine the rule** to exempt domain-scoped skill prefixes, while preserving strict enforcement for foundation and workflow skills. This is a one-time mechanism extension that handles both current and future domain-prefix introductions.

### Why foundation-quality symmetry addition (Track B) is still needed

After Track A exemption is applied, `moai-foundation-quality` is the only skill that legitimately triggers LR-08 in the expert category: 2 of 6 expert agents (performance + security) preload it, 4 do not. This is a real symmetry gap that should be resolved.

Three resolution paths exist:
- **(a) Add to the missing 4 agents** (this SPEC's Track B): Universal access to TRUST 5 framework, OWASP checklist, quality metrics, performance baselines. Aligns with the foundation skill's "universally relevant" classification.
- **(b) Remove from the 2 agents that have it**: Rejected. Performance and security audits inherently require quality vocabulary; removing it would regress their effectiveness.
- **(c) Leave the warning**: Rejected. The warning would persist as noise indefinitely, defeating the W2-deferred dissolution goal.

Path (a) is chosen because `moai-foundation-quality` IS a foundation skill — the LR-08 rule's enforcement is correct here. The agents that lack it are simply incomplete.

### Why Track A and Track B are both required

Together, Tracks A and B achieve full LR-08 dissolution:
- **Track A alone**: removes 10 of 12 LR-08 warnings (the domain-prefix ones). 2 foundation-quality warnings remain.
- **Track B alone**: removes 2 of 12 LR-08 warnings (the foundation-quality ones). 10 domain-skill warnings remain.
- **Track A + Track B**: removes all 12 warnings. Primary success metric AC-CSLM-006 satisfied.

Tracks are independent (no shared state, no execution dependency) but verification (AC-CSLM-006) requires both complete.

### Skill Prefix Taxonomy Mapping

The exemption decision aligns with the canonical skill taxonomy in `.claude/rules/moai/development/skill-authoring.md`:

| Prefix | Classification | LR-08 enforcement | Rationale |
|--------|---------------|-------------------|-----------|
| `moai-foundation-` | Universal foundation | ENFORCED | Every agent benefits (core principles, quality, thinking, CC integration) |
| `moai-workflow-` | Workflow primitives | ENFORCED | Every agent uses workflow primitives (DDD/TDD, testing, project, spec) |
| `moai-domain-` | Domain-scoped knowledge | **EXEMPTED** | Domain-specific by design (backend, frontend, database, copywriting, brand-design) |
| `moai-design-` | Design system / handoff | **EXEMPTED** | Design-scoped by design (design-system, design-handoff) |
| `moai-library-` | Third-party library | **EXEMPTED** | Library-specific by design (shadcn, react-query, prisma, etc.) |
| `moai-framework-` | Framework expertise | **EXEMPTED** | Framework-specific by design (electron, next.js, fastify, etc.) |
| `moai-platform-` | Deployment platforms | **EXEMPTED** | Platform-specific by design (auth, chrome-extension, deployment, vercel) |
| `moai-ref-` | Reference / pattern collections | **EXEMPTED** | Reference-specific by design (api-patterns, owasp-checklist, react-patterns) |

Verified against actual `.claude/skills/` tree (research.md §3.5): all 6 exempted prefixes exist in the catalog.

Note: `moai-meta-*` and `moai-harness-*` prefixes are intentionally NOT exempted (they should be uniform across categories) — see research.md §3.4 for taxonomy rationale.

## Code Change Pattern (Track A)

### Constant + Helper Definition

Add to `internal/cli/agent_lint.go` (after existing top-level declarations, before `checkSkillPreloadDrift`):

```go
// domainExemptPrefixes lists skill name prefixes that are intentionally
// agent-specific and exempt from LR-08 intra-category symmetry checking.
// Foundation skills (moai-foundation-*) and workflow skills (moai-workflow-*)
// are NOT exempted — they are universally relevant.
// See SPEC-V3R5-CORE-SLIM-001 §Design Rationale for the taxonomy mapping.
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
// Used by checkSkillPreloadDrift (LR-08) to skip symmetry enforcement
// on agent-specific domain skills.
func isDomainExemptSkill(skill string) bool {
    for _, prefix := range domainExemptPrefixes {
        if strings.HasPrefix(skill, prefix) {
            return true
        }
    }
    return false
}
```

### Skip Clause Insertion

Modify `checkSkillPreloadDrift` inner loop (current lines 959-973). The verbatim current loop:

```go
// Check for drift
for _, agent := range agents {
    for _, skill := range agent.skills {
        if skillCounts[skill] < len(agents) {
            // This skill is not shared by all agents in category
            lineNum := findFrontmatterLine(agent.file, "skills:")
            violations = append(violations, LintViolation{
                Rule:     "LR-08",
                Severity: SeverityWarning,
                File:     agent.file,
                Line:     lineNum,
                Message:  fmt.Sprintf("Skill preload drift in category '%s': %s is not preloaded by all agents (may cause inconsistent context)", category, skill),
            })
        }
    }
}
```

After the change:

```go
// Check for drift
for _, agent := range agents {
    for _, skill := range agent.skills {
        if isDomainExemptSkill(skill) {
            continue // Domain-scoped skills exempt from LR-08 — see SPEC-V3R5-CORE-SLIM-001
        }
        if skillCounts[skill] < len(agents) {
            // This skill is not shared by all agents in category
            lineNum := findFrontmatterLine(agent.file, "skills:")
            violations = append(violations, LintViolation{
                Rule:     "LR-08",
                Severity: SeverityWarning,
                File:     agent.file,
                Line:     lineNum,
                Message:  fmt.Sprintf("Skill preload drift in category '%s': %s is not preloaded by all agents (may cause inconsistent context)", category, skill),
            })
        }
    }
}
```

Only the 3-line `if isDomainExemptSkill(skill) { continue }` block is added. No other lines in `checkSkillPreloadDrift` are touched.

### Unit Test Pattern

Follow the existing `internal/cli/agent_lint_test.go` (47KB, ~1300 LOC) test conventions:
- Table-driven tests with `tests := []struct { ... }`
- Use `t.TempDir()` for fixture file creation
- Helper function for fixture writing if pattern exists in file
- Each test self-contained (no shared state)

Recommended test structure for the 4 new tests:

```go
func TestSkillPreloadDriftExemption_DomainSkills(t *testing.T) {
    // Fixture: 3 expert agents in tmpdir.
    // Agent 1 has moai-domain-backend; Agents 2 and 3 do not.
    // Setup fixture files (use existing helper if available)
    // Assert: checkSkillPreloadDrift returns 0 violations for moai-domain-backend
}

func TestSkillPreloadDriftExemption_FoundationSkills(t *testing.T) {
    // Fixture: 3 expert agents.
    // Agent 1 has moai-foundation-quality; Agents 2 and 3 do not.
    // Assert: checkSkillPreloadDrift returns ≥1 violation for moai-foundation-quality
}

func TestSkillPreloadDriftExemption_WorkflowSkills(t *testing.T) {
    // Fixture: 3 expert agents.
    // Agent 1 has moai-workflow-testing; Agents 2 and 3 do not.
    // Assert: checkSkillPreloadDrift returns ≥1 violation for moai-workflow-testing
}

func TestSkillPreloadDriftExemption_EdgeCases(t *testing.T) {
    // Table-driven sub-cases:
    // - empty skills array
    // - single-agent category (len(agents) < 2 early return)
    // - foundation-domain-lookalike (literal prefix match must not exempt)
    // - all 6 domain prefixes individually verified as exempted
}
```

Sub-case D in EdgeCases sweeps the 6 prefixes:

```go
prefixes := []string{
    "moai-domain-",
    "moai-design-",
    "moai-library-",
    "moai-framework-",
    "moai-platform-",
    "moai-ref-",
}
for _, prefix := range prefixes {
    t.Run("exempt_"+prefix, func(t *testing.T) {
        skill := prefix + "test"
        require.True(t, isDomainExemptSkill(skill))
    })
}
```

## Track B Skill Insertion Pattern

The `skills:` array in agent frontmatter is a YAML list under 2-space indentation. Example from `expert-refactoring.md` (current state):

```yaml
skills:
  - moai-foundation-core
  - moai-workflow-testing
```

After Track B insertion:

```yaml
skills:
  - moai-foundation-core
  - moai-foundation-quality
  - moai-workflow-testing
```

### Insertion Rule

- Add `- moai-foundation-quality` to the YAML `skills:` array.
- Position alphabetically within the existing array (if already sorted).
- If existing array is NOT strictly sorted (verify via Read first), insert at the position that preserves the existing pattern (typically: foundation-* skills grouped before workflow-* skills, end-anchored on `workflow-testing`).
- Preserve exact existing 2-space indentation and `- ` (dash + space) prefix.
- Tab characters are forbidden in YAML; never substitute.

### Per-Agent Insertion Details

| Agent | Before | After |
|-------|--------|-------|
| `expert-backend` | `core, domain-backend, domain-database, workflow-testing` | `core, domain-backend, domain-database, foundation-quality, workflow-testing` |
| `expert-frontend` | `core, domain-frontend, design-system, workflow-testing` | `core, domain-frontend, design-system, foundation-quality, workflow-testing` (preserve existing order, append before workflow-testing) |
| `expert-refactoring` | `core, workflow-testing` | `core, foundation-quality, workflow-testing` |
| `expert-devops` | `core, workflow-testing` | `core, foundation-quality, workflow-testing` |

For `expert-frontend`, the existing order is `core, domain-frontend, design-system, workflow-testing` — not strictly alphabetical (`design-system` would come before `domain-frontend` alphabetically). Decision per design.md §YAML Array Insertion Pattern: preserve existing free-form order and insert `moai-foundation-quality` between `design-system` and `workflow-testing` to maintain the `workflow-testing` end-anchor pattern visible across all 4 agents. Re-sorting the entire array is OUT OF SCOPE (Agent Core Behavior 5 — scope discipline).

### Edit Tool Invocation Example

For `expert-refactoring.md`:
- `old_string`: `  - moai-foundation-core\n  - moai-workflow-testing`
- `new_string`: `  - moai-foundation-core\n  - moai-foundation-quality\n  - moai-workflow-testing`

This pattern repeats for all 4 source files + 4 template mirrors.

## Edge Case Handling

### EC-1 — Future domain skill prefix introduced

A future SPEC introduces a new domain-scoped skill prefix (e.g., `moai-mobile-`, `moai-cloud-`). Mitigation: extend `domainExemptPrefixes` constant by one line. The helper `isDomainExemptSkill` iterates all entries; adding a prefix requires no algorithmic change. The 6-prefix list is a design baseline, not a hard maximum.

### EC-2 — Foundation skill named with domain-like pattern

Hypothetical skill `moai-foundation-frontend-test` exists. The exemption check uses `strings.HasPrefix(skill, "moai-domain-")` etc. — literal prefix matching, NOT substring matching. `moai-foundation-frontend-test` starts with `moai-foundation-`, NOT with any of the 6 exempted prefixes, so the exemption returns false. The skill is correctly subject to LR-08 enforcement.

Cross-check: `isDomainExemptSkill("moai-foundation-frontend-test")` returns false because none of the 6 exempted prefixes (`moai-domain-`, `moai-design-`, etc.) are prefixes of this skill name.

### EC-3 — Empty skills array

Agent file with `skills: []` (or missing the field entirely) is parsed by `parseSkillsList`. The function returns an empty slice. The inner drift loop has no entries to iterate, so no violations emitted. No edge handling required.

### EC-4 — Single-agent category

If a category has only 1 agent (none currently, possible in future), the existing `len(agents) < 2` early return at line 946 skips symmetry checking. No edge handling required.

### EC-5 — Track A regression on non-expert categories

The exemption logic is category-agnostic (operates inside the per-category drift loop with no category-specific code). Non-expert categories (manager, builder, evaluator) benefit from the same exemption naturally. AC-CSLM-007 verifies non-regression across all category lint counts. Track A unit tests use multi-category fixtures.

### EC-6 — Mirror missing at run-time

Pre-mirror-edit probe: `test -f internal/template/templates/.claude/agents/moai/expert-<name>.md`.
- If absent: halt with sentinel `MIRROR_MISSING_BLOCKER`, return blocker report to orchestrator.
- If present: proceed with mirror edit.

## Implementation Notes

### Template-First Discipline

Per `CLAUDE.local.md §2` ([HARD] Template-First Rule), the canonical source for agent definitions is `internal/template/templates/`, and `.claude/agents/moai/` is the sync target. The Track B run-phase implementation order is:

1. Edit `internal/template/templates/.claude/agents/moai/expert-{backend,frontend,refactoring,devops}.md` (source)
2. Run `make build` (regenerates `internal/template/embedded.go`)
3. Sync to live `.claude/agents/moai/expert-{backend,frontend,refactoring,devops}.md` (target)

Alternatively, edit both sets synchronously with `MultiEdit` and verify byte-identical via AC-CSLM-003. Both paths are semantically equivalent — the constraint is the post-edit invariant, not the edit order.

### Catalog Hash Refresh Order

The catalog hash table `internal/template/catalog.yaml` records SHA hashes for every template file. After editing template files, the catalog MUST be refreshed via `go run internal/template/scripts/gen-catalog-hashes.go --all`. If the catalog is NOT refreshed:
- `go test ./internal/template/...` fails (catalog hash audit mismatch).
- CI Tier 1 Test check fails.
- Branch protection blocks merge.

Refresh order for Track B:
1. All file edits complete (sources + mirrors)
2. `make build` (validates YAML)
3. `gen-catalog-hashes.go --all` (refreshes catalog)
4. `go test ./internal/template/...` (verifies catalog matches)
5. `./bin/moai agent lint --strict` (verifies LR-08 dissolution)

Track A is independent — Go code changes do NOT trigger catalog refresh requirement.

### Atomic Commit Boundary

The 11 files modified (1 Go src + 1 Go test + 4 source agents + 4 mirror agents + 1 catalog yaml) MUST be committed atomically in a single commit:

```
feat(SPEC-V3R5-CORE-SLIM-001): W2 — LR-08 rule refinement + foundation-quality symmetry (LR-08: 12 → 0)
```

Splitting into multiple commits risks intermediate states where:
- Track A merged alone leaves 2 foundation-quality LR-08 warnings (acceptable but not goal)
- Track B merged alone leaves 10 domain-skill LR-08 warnings (acceptable but not goal)
- Catalog hash refresh between commits could diverge from file SHAs

Single atomic commit avoids transient inconsistency in the lint output.

### LR rule pattern precedent

LR-09 (read-only worktree rejection) uses similar prefix-aware exemption logic — a design precedent confirming that adding prefix-based exemption to `checkSkillPreloadDrift` is consistent with existing rule patterns in `agent_lint.go`.

## Constitution Alignment

Per `.claude/rules/moai/development/coding-standards.md` and CLAUDE.md §1 HARD Rules:

- ✓ All edits are in English (skill names + YAML + Go code are English-only).
- ✓ Template-First Rule respected (Track B source edited first, mirrored to live).
- ✓ No new dependencies introduced (Track A uses existing `strings` package).
- ✓ No new commands added (uses existing `moai agent lint`, `make build`, `gen-catalog-hashes.go`).
- ✓ Track B is pure metadata edits; Track A is well-isolated logic refactor with bidirectional test coverage.
- ✓ Go test coverage extended (4 new test cases per AC-CSLM-005) per project's 85%+ coverage standard.

## Cross-References

- spec.md §3 (EARS Requirements) — REQ-CSLM-001 through 005
- spec.md §4 (Acceptance Criteria) — AC-CSLM-001 through 007
- spec.md §5 (Edge Cases) — EC-1 through EC-6
- plan.md §Track A + §Track B — step-by-step implementation
- research.md §3 (LR-08 semantics) and §4 (EC-4 Discovery Log)
- `internal/cli/agent_lint.go:892-977` — LR-08 rule authoritative implementation (Track A target)
- `internal/cli/agent_lint_test.go` — existing test scaffolding (Track A append target)
- `.claude/rules/moai/development/skill-authoring.md` — skill prefix taxonomy
- `.claude/rules/moai/development/agent-authoring.md` — agent frontmatter format rules
- `CLAUDE.local.md §2` — Template-First Rule
