# Research — SPEC-V3R5-CORE-SLIM-001

Baseline data, LR-08 implementation analysis, EC-4 Discovery Log, and cross-reference index supporting the V3R5 W2 LR-08 rule refinement + expert agent foundation skill symmetry.

## 1. Pre-Plan Verification (2026-05-20)

### 1.1 Skill Availability

The required `moai-foundation-quality` skill exists on disk and is loadable. The 6 domain-prefix exemption groups all have at least one representative skill:

```bash
$ ls .claude/skills/moai-foundation-quality/SKILL.md
.claude/skills/moai-foundation-quality/SKILL.md  ✓ present

$ ls -d .claude/skills/moai-domain-*
.claude/skills/moai-domain-backend
.claude/skills/moai-domain-brand-design
.claude/skills/moai-domain-copywriting
.claude/skills/moai-domain-database
.claude/skills/moai-domain-design-handoff
.claude/skills/moai-domain-frontend
.claude/skills/moai-domain-ideation
.claude/skills/moai-domain-research

$ ls -d .claude/skills/moai-design-*
.claude/skills/moai-design-system

$ ls -d .claude/skills/moai-framework-*
.claude/skills/moai-framework-electron

$ ls -d .claude/skills/moai-platform-*
.claude/skills/moai-platform-auth
.claude/skills/moai-platform-chrome-extension
.claude/skills/moai-platform-deployment

$ ls -d .claude/skills/moai-ref-*
.claude/skills/moai-ref-api-patterns
.claude/skills/moai-ref-git-workflow
.claude/skills/moai-ref-owasp-checklist
.claude/skills/moai-ref-react-patterns
.claude/skills/moai-ref-testing-pyramid
```

Note: `moai-library-*` prefix has no current representatives in the catalog but is included in `domainExemptPrefixes` per the canonical taxonomy in `.claude/rules/moai/development/skill-authoring.md` and future-proofing.

### 1.2 Mirror File Availability

All 4 template mirror files (Track B targets) exist:

```bash
$ ls internal/template/templates/.claude/agents/moai/expert-{backend,frontend,refactoring,devops}.md
internal/template/templates/.claude/agents/moai/expert-backend.md      ✓ present
internal/template/templates/.claude/agents/moai/expert-frontend.md     ✓ present
internal/template/templates/.claude/agents/moai/expert-refactoring.md  ✓ present
internal/template/templates/.claude/agents/moai/expert-devops.md       ✓ present
```

EC-6 (mirror missing) is NOT triggered at plan-phase time.

### 1.3 Source File Availability

All 4 live source agent files (Track B targets) exist:

```bash
$ ls .claude/agents/moai/expert-{backend,frontend,refactoring,devops}.md
.claude/agents/moai/expert-backend.md      ✓ present
.claude/agents/moai/expert-frontend.md     ✓ present
.claude/agents/moai/expert-refactoring.md  ✓ present
.claude/agents/moai/expert-devops.md       ✓ present
```

### 1.4 Track A Code Availability

`internal/cli/agent_lint.go` and `internal/cli/agent_lint_test.go` both exist:

```bash
$ ls -la internal/cli/agent_lint*.go
-rw-r--r--@ 1 goos  staff  47073  internal/cli/agent_lint_test.go
-rw-r--r--@ 1 goos  staff  34426  internal/cli/agent_lint.go
```

Substantial test file (47KB, ~1300 LOC) provides scaffolding for Track A new test cases.

## 2. Baseline LR-08 Detection (verbatim, captured 2026-05-20)

The following `./bin/moai agent lint --strict` output (`grep "LR-08"` filtered) is the canonical baseline this SPEC targets. Captured on `plan/SPEC-V3R5-CORE-SLIM-001` branched from `origin/main` at `fedb67db4`:

```
! [LR-08] /Users/goos/moai/moai-adk-go/.claude/agents/moai/expert-backend.md:19: Skill preload drift in category 'expert': moai-domain-backend is not preloaded by all agents (may cause inconsistent context)
! [LR-08] /Users/goos/moai/moai-adk-go/.claude/agents/moai/expert-backend.md:19: Skill preload drift in category 'expert': moai-domain-database is not preloaded by all agents (may cause inconsistent context)
! [LR-08] /Users/goos/moai/moai-adk-go/.claude/agents/moai/expert-frontend.md:17: Skill preload drift in category 'expert': moai-domain-frontend is not preloaded by all agents (may cause inconsistent context)
! [LR-08] /Users/goos/moai/moai-adk-go/.claude/agents/moai/expert-frontend.md:17: Skill preload drift in category 'expert': moai-design-system is not preloaded by all agents (may cause inconsistent context)
! [LR-08] /Users/goos/moai/moai-adk-go/.claude/agents/moai/expert-performance.md:16: Skill preload drift in category 'expert': moai-foundation-quality is not preloaded by all agents (may cause inconsistent context)
! [LR-08] /Users/goos/moai/moai-adk-go/.claude/agents/moai/expert-security.md:15: Skill preload drift in category 'expert': moai-foundation-quality is not preloaded by all agents (may cause inconsistent context)
! [LR-08] /Users/goos/moai/moai-adk-go/internal/template/templates/.claude/agents/moai/expert-backend.md:19: Skill preload drift in category 'expert': moai-domain-backend is not preloaded by all agents (may cause inconsistent context)
! [LR-08] /Users/goos/moai/moai-adk-go/internal/template/templates/.claude/agents/moai/expert-backend.md:19: Skill preload drift in category 'expert': moai-domain-database is not preloaded by all agents (may cause inconsistent context)
! [LR-08] /Users/goos/moai/moai-adk-go/internal/template/templates/.claude/agents/moai/expert-frontend.md:17: Skill preload drift in category 'expert': moai-domain-frontend is not preloaded by all agents (may cause inconsistent context)
! [LR-08] /Users/goos/moai/moai-adk-go/internal/template/templates/.claude/agents/moai/expert-frontend.md:17: Skill preload drift in category 'expert': moai-design-system is not preloaded by all agents (may cause inconsistent context)
! [LR-08] /Users/goos/moai/moai-adk-go/internal/template/templates/.claude/agents/moai/expert-performance.md:16: Skill preload drift in category 'expert': moai-foundation-quality is not preloaded by all agents (may cause inconsistent context)
! [LR-08] /Users/goos/moai/moai-adk-go/internal/template/templates/.claude/agents/moai/expert-security.md:15: Skill preload drift in category 'expert': moai-foundation-quality is not preloaded by all agents (may cause inconsistent context)
```

**Total: 12 LR-08 warnings** (6 source + 6 mirror), matching the W1 documented `W2-deferred` baseline.

Breakdown by resolution track:
- **10 warnings** mention domain-prefix skills (`moai-domain-backend`, `moai-domain-database`, `moai-domain-frontend`, `moai-design-system`) → dissolved by Track A (rule fix)
- **2 unique warnings** (4 lines: 2 source + 2 mirror) mention `moai-foundation-quality` → dissolved by Track B (preload addition)

## 3. LR-08 Implementation Analysis

### 3.1 Rule Source (verbatim)

The LR-08 rule is implemented in `internal/cli/agent_lint.go` lines 892-977. Verbatim function:

```go
// checkSkillPreloadDrift checks for LR-08 across all agent files.
func checkSkillPreloadDrift(files []string) []LintViolation {
    var violations []LintViolation

    // Group agents by category
    type agentInfo struct {
        file   string
        skills []string
    }

    categories := map[string][]agentInfo{
        "manager":   {},
        "expert":    {},
        "builder":   {},
        "evaluator": {},
    }

    for _, file := range files {
        content, err := os.ReadFile(file)
        if err != nil {
            continue
        }

        parts := bytes.SplitN(content, []byte("---"), 3)
        if len(parts) < 3 {
            continue
        }

        // Parse skills list from YAML frontmatter
        skills := parseSkillsList(parts[1])
        name := parseFieldName(parts[1])

        // Determine category from agent name
        category := ""
        if strings.HasPrefix(name, "manager-") {
            category = "manager"
        } else if strings.HasPrefix(name, "expert-") {
            category = "expert"
        } else if strings.HasPrefix(name, "builder-") {
            category = "builder"
        } else if strings.HasPrefix(name, "evaluator-") {
            category = "evaluator"
        }

        if category != "" {
            categories[category] = append(categories[category], agentInfo{
                file:   file,
                skills: skills,
            })
        }
    }

    // Check for skill preload drift within each category
    for category, agents := range categories {
        if len(agents) < 2 {
            continue // Need at least 2 agents to compare
        }

        // Find baseline skill set (most common)
        skillCounts := make(map[string]int)
        for _, agent := range agents {
            for _, skill := range agent.skills {
                skillCounts[skill]++
            }
        }

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
    }

    return violations
}
```

The implementation enforces **strict uniform symmetry**: any skill that any agent in a category preloads MUST be preloaded by ALL agents in that category. There is NO domain-specificity exemption — the symmetry is enforced indiscriminately across all skill names.

The drift detection logic (inner loop at lines 959-973) fires the warning on the file that HAS the skill, not the one that lacks it. This is why the 4 expert agents listed in the original v0.1.0 scope appear in the LR-08 output — they have the skill, and OTHER agents in the category do not.

### 3.2 Per-Agent Pre-State Inspection (2026-05-20)

Captured via direct read of all 6 expert agent files:

#### expert-backend.md (skills array)

```yaml
skills:
  - moai-foundation-core
  - moai-domain-backend
  - moai-domain-database
  - moai-workflow-testing
```

#### expert-frontend.md (skills array)

```yaml
skills:
  - moai-foundation-core
  - moai-domain-frontend
  - moai-design-system
  - moai-workflow-testing
```

#### expert-performance.md (skills array)

```yaml
skills:
  - moai-foundation-core
  - moai-foundation-quality
  - moai-workflow-testing
```

#### expert-security.md (skills array)

```yaml
skills:
  - moai-foundation-core
  - moai-foundation-quality
  - moai-workflow-testing
```

#### expert-refactoring.md (skills array) — Track B target

```yaml
skills:
  - moai-foundation-core
  - moai-workflow-testing
```

#### expert-devops.md (skills array) — Track B target

```yaml
skills:
  - moai-foundation-core
  - moai-workflow-testing
```

### 3.3 Category-Wide Symmetry Analysis

Apply the LR-08 algorithm to the empirical state:

Category-wide skill counts in `expert` (6 agents total):
- `moai-foundation-core`: 6 / 6 (uniform → no drift)
- `moai-workflow-testing`: 6 / 6 (uniform → no drift)
- `moai-foundation-quality`: 2 / 6 (drift → 2 LR-08 fires, on performance + security)
- `moai-domain-backend`: 1 / 6 (drift → 1 LR-08 fires, on backend)
- `moai-domain-database`: 1 / 6 (drift → 1 LR-08 fires, on backend)
- `moai-domain-frontend`: 1 / 6 (drift → 1 LR-08 fires, on frontend)
- `moai-design-system`: 1 / 6 (drift → 1 LR-08 fires, on frontend)

Total: 6 unique (agent, skill) drift findings × 2 file paths (source + mirror) = 12 LR-08 lines, matching §2 verbatim baseline.

### 3.4 Skill Taxonomy Source

Per `.claude/rules/moai/development/skill-authoring.md` and verified against actual `.claude/skills/` tree:

| Prefix | Universal? | Examples |
|--------|-----------|----------|
| `moai-foundation-` | YES | core, cc, quality, thinking |
| `moai-workflow-` | YES | ddd, tdd, testing, spec, project, worktree, design-import, design-context, ci-autofix, ci-watch, gan-loop, loop |
| `moai-domain-` | NO (agent-specific) | backend, frontend, database, copywriting, brand-design, design-handoff, ideation, research |
| `moai-design-` | NO (design-specific) | system |
| `moai-library-` | NO (library-specific) | (none currently in catalog; reserved) |
| `moai-framework-` | NO (framework-specific) | electron |
| `moai-platform-` | NO (platform-specific) | auth, chrome-extension, deployment |
| `moai-ref-` | NO (reference-specific) | api-patterns, git-workflow, owasp-checklist, react-patterns, testing-pyramid |

The `moai-meta-*` and `moai-harness-*` prefixes also exist in the catalog (e.g., `moai-meta-harness`, `moai-harness-learner`) but are not used by expert agents and are out of scope for Track A's exemption list (they would be subject to LR-08 enforcement if any expert agent preloaded them; none do).

### 3.5 Domain Prefix List (Canonical 6)

Derived from §3.4 and codified in Track A's `domainExemptPrefixes`:

```
moai-domain-
moai-design-
moai-library-
moai-framework-
moai-platform-
moai-ref-
```

All 6 prefixes are documented in the canonical skill taxonomy and reflect the design intent that domain-scoped skills should NOT be uniformly preloaded across categorized agents.

## 4. EC-4 Discovery Log (v0.1.0 → v0.2.0 Scope Pivot Rationale)

This section documents the empirical investigation chain that led to the v0.2.0 scope pivot.

### 4.1 Original v0.1.0 Scope

The initial SPEC scope (v0.1.0, dated 2026-05-20) proposed:

- Add 6 missing skills to 4 expert agents (backend, frontend, performance, security)
- Mirror edits to 4 template files
- Total: 8 documentation file edits
- Estimated diff: ~12 lines added across 8 markdown files

### 4.2 Verification Chain

**Step 1 — Verbatim baseline capture** (`./bin/moai agent lint --strict | grep LR-08`):

Captured 12 LR-08 warnings as documented in §2 above. The warning text was understood at face value: "moai-domain-backend is not preloaded by all agents" → interpreted as "expert-backend lacks moai-domain-backend".

**Step 2 — Skill array empirical inspection** (`grep -A N "^skills:" <file>` across all 4 named source agents):

Result documented in §3.2. **All 6 matrix skills were ALREADY present on the named source agents**:
- expert-backend already has `moai-domain-backend` AND `moai-domain-database` (✓ matches matrix's 2)
- expert-frontend already has `moai-domain-frontend` AND `moai-design-system` (✓ matches matrix's 2)
- expert-performance already has `moai-foundation-quality` (✓ matches matrix's 1)
- expert-security already has `moai-foundation-quality` (✓ matches matrix's 1)

This was the first contradiction signal: the user prompt's matrix listed skills to ADD, but they were already present.

**Step 3 — LR-08 implementation read** (`internal/cli/agent_lint.go:892-977`, verbatim source in §3.1):

The rule implementation revealed the actual semantics:
- The rule emits the warning on the file that HAS the skill (line 967: `File: agent.file`), against the absent peers in the category.
- "X is not preloaded by all agents" means: X is preloaded by SOME agent here, but NOT by every agent in the category — so this agent (the one with the skill) is flagged.
- The strict uniform symmetry check has NO domain-specificity exemption.

**Step 4 — Re-analysis** (synthesis):

- The 4 expert agents named in the v0.1.0 matrix ARE the agents that already preload the listed skills.
- The LR-08 warnings fire because OTHER agents in the expert category (expert-refactoring, expert-devops, and the 2 agents missing foundation-quality) lack these skills.
- The v0.1.0 mechanical scope (add 6 skills to 4 named agents) would result in 0 net edits (idempotent no-ops), failing to dissolve LR-08.

**Step 5 — Inverse-interpretation alternatives** (preserved in this research as historical context):

- **(a) Additive on out-of-scope agents (expert-refactoring + expert-devops)**: add the 6 skills to the 2 out-of-scope agents. Direction satisfies AC-CSLM-002 (LR-08 = 0) but semantically wrong (bloats refactoring/devops contexts with backend/frontend/design domain knowledge).
- **(b) Subtractive on 4 matrix agents**: remove the 4 domain skills from the 4 named agents. Direction satisfies AC-CSLM-002 but semantically regressive (frontend agent loses design-system access).
- **(c) Rule refinement**: exempt domain-prefix skills from LR-08 symmetry check + add `moai-foundation-quality` to 4 agents missing it. Direction satisfies AC-CSLM-002 with minimum semantic distortion.

The user (2026-05-20) confirmed direction (c) as the correct path, leading to the v0.2.0 scope pivot.

### 4.3 v0.2.0 Scope Definition

The v0.2.0 scope (this SPEC) reflects direction (c):

- **Track A**: refine LR-08 rule to exempt domain-prefix skills → eliminates 10 domain-skill warnings
- **Track B**: add `moai-foundation-quality` to 4 expert agents missing it → eliminates 2 foundation-quality warnings
- Total: 12 → 0 LR-08 warnings

This scope is the **structural fix** rather than a mechanical patch: it addresses the rule's over-strict default while preserving its core invariant (foundation/workflow skill uniformity).

### 4.4 Lessons Captured

- **Verify scope empirically before planning**: the original scope assumed the matrix described missing skills; verification revealed they were present. Skip the verification step → waste planning effort.
- **Read the rule source before applying mechanical fixes**: the LR-08 warning text is ambiguous about which file is flagged. Direct source inspection (§3.1) resolved the ambiguity.
- **Distinguish "metadata patch" from "rule refinement"**: when the rule is over-strict for an intentional design pattern (domain-specific preloads), patch the rule, not the metadata.

## 5. Historical Reference

### 5.1 W1 Deferral Context

Source: project memory `project_v3r5_w1_constitution_complete` (2026-05-20 entry):

> AskUserQuestion 결과 → W2 CORE-SLIM-001 (권장) 선택: 12 W2-deferred LR-08 baseline (expert-{backend,frontend,performance,security} × moai-foundation-quality preload drift) → 0 해소 → chicken-and-egg admin override 종결.

The W1 SPEC (SPEC-V3R5-CONSTITUTION-DUAL-001) intentionally deferred these 12 findings as the W2-deferred baseline so that W1's plan + run + sync PRs (#1015, #1016, #1017) could pass CI via admin-squash-merge without resolving an unrelated lint regression. The deferral was an explicit chicken-and-egg break: W1 needed to land first to provide the constitution validator that future SPECs (including this one) consume.

Note: the W1 deferral memory described the deferred baseline as "expert-{backend,frontend,performance,security} × moai-foundation-quality preload drift", which is a simplified summary. The actual baseline (verified §2 above) covers 5 distinct skills across 4 distinct source agents + their 4 mirrors. The v0.2.0 scope (this SPEC) targets the full empirical baseline, not the simplified summary.

### 5.2 Sibling SPEC Precedent

`SPEC-V3R5-LINT-CLEAN-001` (LCLN-001, COMPLETE per memory `project_v3r5_lcln_001_complete`) established the delta-only D6 NEW=0 enforcement pattern this SPEC inherits:

- LCLN-001 cleared 164 of 176 baseline findings across 4 phases (60 + 4 + 30 + 70 = 164).
- LCLN-001 documented 12 residual findings (the same 12 this SPEC targets) as "W2-deferred" — 8 pure W2 (expert-{backend,frontend} × domain skills) + 4 post-W2 cleanup (moai-foundation-quality drift on expert-{performance,security}).
- LCLN-001 created a manifest at `.moai/state/lint-w2-deferred.json` enumerating these 12 deferred findings.
- LCLN-001 noted that **Pattern (a) experiment** (under-includer expansion — analogous to v0.1.0's mechanical-addition direction) was attempted and "confirmed worsened lint 12→14", and **Pattern (b) functionally regressive** — providing prior-art evidence that mechanical metadata patches CANNOT cleanly dissolve the W2-deferred manifest. The rule refinement direction (Track A) avoids both pitfalls.

This SPEC is the natural conclusion of LCLN-001's W2-deferred manifest with the corrected approach (rule refinement + minimal preload addition).

### 5.3 W2-Deferred Manifest Discrepancy

The on-disk W2-deferred manifest at `.moai/state/lint-w2-deferred.json` (authored by SPEC-V3R5-LINT-CLEAN-001 at LCLN-Phase 4) does NOT match the empirical lint output captured in §2. Both are on-disk artifacts; the discrepancy is between two persisted snapshots, NOT between memory and reality.

**(a) Manifest content** (`.moai/state/lint-w2-deferred.json`, captured 2026-05-19T23:30:00Z, main HEAD `917c95fef`):

The manifest enumerates 12 W2-deferred findings across 6 file paths (source + mirror for expert-backend / expert-frontend, 3 drift skills per file pair). Specifically:

- `expert-backend.md` (source + mirror): `moai-domain-backend`, `moai-domain-database`, `moai-workflow-testing` × 2 paths = 6 entries
- `expert-frontend.md` (source + mirror): `moai-design-system`, `moai-domain-frontend`, `moai-workflow-testing` × 2 paths = 6 entries

The manifest counts `moai-workflow-testing` as a drift skill (4 of 12 entries) on expert-backend + expert-frontend.

**(b) Empirical lint content** (verbatim §2 output, captured 2026-05-20 on `plan/SPEC-V3R5-CORE-SLIM-001` branched from `origin/main` at `fedb67db4`):

The empirical baseline shows 12 LR-08 lines, BUT the skill set is different:

- `moai-domain-backend`: 2 lines (source + mirror on expert-backend) — matches manifest
- `moai-domain-database`: 2 lines (source + mirror on expert-backend) — matches manifest
- `moai-domain-frontend`: 2 lines (source + mirror on expert-frontend) — matches manifest
- `moai-design-system`: 2 lines (source + mirror on expert-frontend) — matches manifest
- `moai-foundation-quality`: 4 lines (source + mirror on expert-performance AND expert-security) — NOT in manifest
- `moai-workflow-testing`: 0 lines — manifest claims 4 entries but empirical shows 0

Per §3.2 verified state, all 6 expert agents preload `moai-workflow-testing` (uniform 6/6), so LR-08 correctly emits ZERO warnings for it. The manifest's 4 `moai-workflow-testing` entries are stale or speculative; they do not match the actual rule output.

**(c) The disagreement is between two on-disk artifacts**, not between memory and reality. Both `.moai/state/lint-w2-deferred.json` and the empirical `./bin/moai agent lint --strict` output are persisted snapshots. The manifest was captured at LCLN-Phase 4 closure (2026-05-19); the empirical baseline was captured at this SPEC's plan-phase (2026-05-20).

**(d) The empirical lint output is the authoritative source of truth.** The manifest is stale. Specifically:

- The manifest's `moai-workflow-testing` entries (4 of 12) are not reproducible — the empirical rule output emits 0 such warnings, because all 6 expert agents already preload `moai-workflow-testing` (verified §3.2).
- The manifest omits the 4 `moai-foundation-quality` LR-08 warnings on expert-performance + expert-security (verified §2).
- Numerically the manifest and empirical both total 12, but the skill set composition differs.

**Resolution**: This SPEC dissolves the empirical 12-finding baseline (via Track A + Track B). The on-disk manifest is then stale-by-construction and MUST be deleted post-AC-CSLM-006 PASS, completing the chicken-and-egg dissolution. The deletion is enforced as an explicit DoD checkbox in `acceptance.md`. No partial-update or merge of manifest content is performed; the file is removed wholesale after the empirical baseline reaches 0.

## 6. Cross-References

### Rule Files

- `internal/cli/agent_lint.go:892-977` — LR-08 rule definition (authoritative source for Track A modification)
- `internal/cli/agent_lint_test.go` — existing test scaffolding (47KB, ~1300 LOC)
- `.claude/rules/moai/development/agent-authoring.md` — agent frontmatter rules including `skills:` array conventions
- `.claude/rules/moai/development/skill-authoring.md` — skill frontmatter and prefix taxonomy source
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — canonical 12-field SPEC frontmatter schema
- `.claude/rules/moai/development/coding-standards.md` — Template-First Pattern enforcement
- `CLAUDE.local.md §2` — [HARD] Template-First Rule for `internal/template/templates/` ↔ `.claude/` sync

### Related SPECs

- `SPEC-V3R5-CONSTITUTION-DUAL-001` (W1, COMPLETED, main HEAD `175bad283`) — origin of W2-deferred baseline
- `SPEC-V3R5-LINT-CLEAN-001` (LCLN-001, COMPLETED) — sibling lint-cleanup SPEC establishing delta-only D6 NEW=0 pattern; documents the same 12-finding W2-deferred manifest and Pattern (a)/(b) experiment failures
- `SPEC-V3R2-ORC-002` (COMPLETED, PR #980) — origin of `moai agent lint` CLI 8-rule lint engine (includes LR-08)
- `SPEC-V3R5-CLAUDE-REFRESH-001` (W0, COMPLETED) — predecessor in v3.5.0 Mega-Sprint sequence

### Project Memory References

- `project_v3r5_w1_constitution_complete` — W1 lifecycle completion + W2-deferred 12-finding documentation + next-session recommendation `/moai plan SPEC-V3R5-CORE-SLIM-001`
- `project_v3r5_lcln_001_complete` — LCLN-001 lifecycle completion + Pattern (a)/(b) experiment results
- `feedback_rules_no_dev_spec_metadata` — rules are user-deployed artifacts; do not embed dev SPEC metadata

### Tooling

- `make build` — invokes Go embed regeneration; validates YAML frontmatter syntax
- `go run internal/template/scripts/gen-catalog-hashes.go --all` — refreshes `internal/template/catalog.yaml`
- `go test ./internal/template/...` — catalog hash audit + other template tests
- `go test ./internal/cli/...` — agent_lint unit tests (Track A test target)
- `./bin/moai agent lint --strict` — LR-01 through LR-12 lint engine

### v3.5.0 Mega-Sprint Roadmap Context

Per `.moai/research/harness-autonomy-vision-2026-05-18.md` §5 (cited via memory):

| Wave | SPEC | Status |
|------|------|--------|
| W0 | SPEC-V3R5-CLAUDE-REFRESH-001 | COMPLETED |
| W1 | SPEC-V3R5-CONSTITUTION-DUAL-001 | COMPLETED |
| **W2** | **SPEC-V3R5-CORE-SLIM-001** | **draft (this SPEC, v0.2.0)** |
| W3 | SPEC-V3R5-HARNESS-AUTONOMY-001 | planned |
| W4 | SPEC-V3R5-PROJECT-MEGA-001 | planned |

This SPEC's completion is the gating condition for W3 entry — the v3.5.0 release timeline depends on the W2-deferred baseline dissolving to 0 here.
