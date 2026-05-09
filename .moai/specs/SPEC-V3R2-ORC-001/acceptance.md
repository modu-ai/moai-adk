---
spec_id: SPEC-V3R2-ORC-001
phase: "1B — Acceptance Criteria"
created_at: 2026-05-09
author: manager-spec
ac_count: 17
---

# Acceptance Criteria: Agent roster consolidation (22 → 17)

Given-When-Then scenarios for every requirement in spec.md §5. Each AC
includes:

- A primary happy-path scenario.
- Edge cases where the requirement could silently degrade.
- Verification command(s) the auditor can run locally.

The numbering follows REQ-IDs from spec.md (REQ-ORC-001-001 → AC-ORC-001-01).

---

## AC-ORC-001-01 — Final roster contains exactly 17 R5-baseline active agents (REQ-001)

**Given** the ORC-001 PR has been merged to main
**When** I list active agents under `.claude/agents/moai/` (excluding retired
stubs and post-R5 additions: `expert-mobile`, `manager-brain`,
`claude-code-guide`)
**Then** I count exactly 17 agent files matching the REQ-001 list:
manager-spec, manager-strategy, manager-cycle, manager-quality, manager-docs,
manager-git, manager-project, expert-backend, expert-frontend,
expert-security, expert-devops, expert-performance, expert-refactoring,
builder-platform, evaluator-active, plan-auditor, researcher.

**Verification**:

```bash
# Active agents (no retired: true, exclude post-R5)
cd /Users/goos/.moai/worktrees/MoAI-ADK/orc-001-plan
ACTIVE=$(grep -L "retired: true" internal/template/templates/.claude/agents/moai/*.md \
  | grep -v -E "expert-mobile|manager-brain|claude-code-guide")
echo "$ACTIVE" | wc -l
# Expected: 17
```

**Edge cases**:
- If a stub accidentally drops `retired: true`, count goes to 18 → fail.
- If `expert-mobile` is misclassified, count goes to 18 → fail (but this is a documentation error, not roster).
- If `claude-code-guide` is in scope (OQ-5), count drifts; spec.md §10.1 destiny table is authoritative.

---

## AC-ORC-001-02 — manager-cycle preserves DDD + TDD parameterization (REQ-002)

**Given** `manager-cycle.md` exists in the template tree
**When** I read its body
**Then**:
1. Frontmatter declares `cycle_type` as a required input parameter (search for `cycle_type` in tools-list area or body header).
2. Body documents both DDD (ANALYZE-PRESERVE-IMPROVE) and TDD (RED-GREEN-REFACTOR) phase names with cycle_type-based dispatch.
3. Migration Notes table maps `manager-ddd` → `cycle_type=ddd` and `manager-tdd` → `cycle_type=tdd`.

**Verification**:

```bash
grep -n "cycle_type" internal/template/templates/.claude/agents/moai/manager-cycle.md | head -5
grep -n "ANALYZE-PRESERVE-IMPROVE\|RED-GREEN-REFACTOR" internal/template/templates/.claude/agents/moai/manager-cycle.md | head -5
grep -A2 "Migration Notes" internal/template/templates/.claude/agents/moai/manager-cycle.md | head -10
# Expected: cycle_type appears, both phase names appear, migration table present
```

**Edge cases**:
- cycle_type referenced only in description but not enforced in body precondition → partial pass; manager-spec must mark as observation.
- ANALYZE-PRESERVE-IMPROVE present but RED-GREEN-REFACTOR missing → fail (TDD path lost).

---

## AC-ORC-001-03 — builder-platform supports all 7 artifact types (REQ-003)

**Given** `builder-platform.md` is created in M2.1
**When** I inspect its frontmatter and body
**Then**:
1. Required parameter `artifact_type` enum: `agent | skill | plugin | command | hook | mcp-server | lsp-server`.
2. Body retains the 5-phase workflow (Requirements → Research → Architecture → Implementation → Validation).
3. Each artifact_type has a distinct dispatch entry (artifact-template differences captured).

**Verification**:

```bash
grep -n "artifact_type" internal/template/templates/.claude/agents/moai/builder-platform.md
grep -n "Requirements.*Research.*Architecture.*Implementation.*Validation\|Phase 1\|Phase 2\|Phase 3\|Phase 4\|Phase 5" internal/template/templates/.claude/agents/moai/builder-platform.md
# Expected: artifact_type with all 7 enum values; 5-phase workflow present
```

**Edge cases**:
- Only 3 artifact types supported (agent/skill/plugin) → fail (REQ requires 7).
- 5-phase structure collapsed into 3 phases → partial pass (less rigorous but acceptable if artifact-type dispatch covers same scope).

---

## AC-ORC-001-04 — manager-quality contains Diagnostic Sub-Mode (REQ-004)

**Given** M3.1 has run
**When** I read `manager-quality.md`
**Then**:
1. A "## Diagnostic Sub-Mode" section is present.
2. The section contains a delegation table (or routing list) covering at minimum: code defects → manager-cycle (refactor); architecture issues → expert-refactoring; git/branch issues → manager-git.
3. The expert-debug Common-Protocol scrub (the AskUserQuestion meta-warning) is preserved in the new section.

**Verification**:

```bash
grep -n "Diagnostic Sub-Mode" internal/template/templates/.claude/agents/moai/manager-quality.md
grep -A20 "Diagnostic Sub-Mode" internal/template/templates/.claude/agents/moai/manager-quality.md | grep -E "manager-cycle|expert-refactoring|manager-git"
# Expected: section header present; delegation entries present
```

**Edge cases**:
- Section header present but body empty → fail.
- Section refers users back to retired `expert-debug` → fail (defeats the purpose).

---

## AC-ORC-001-05 — All 7 retired stubs exist with status=retired (REQ-006)

**Given** M2 has run for all 5 new retirements
**When** I list retired stubs
**Then** all 7 retired stubs exist in template tree:
- `manager-ddd.md`, `manager-tdd.md` (already done by prior SPEC, verified by M1.1)
- `builder-agent.md`, `builder-skill.md`, `builder-plugin.md` (M2.2-M2.4)
- `expert-debug.md` (M2.5)
- `expert-testing.md` (M2.6)

Each stub has frontmatter: `retired: true`, `retired_replacement: <name>`,
optional `retired_param_hint`, `tools: []`, `skills: []`.

**Verification**:

```bash
grep -l "retired: true" internal/template/templates/.claude/agents/moai/*.md | sort
# Expected output (7 lines):
# .../builder-agent.md
# .../builder-plugin.md
# .../builder-skill.md
# .../expert-debug.md
# .../expert-testing.md
# .../manager-ddd.md
# .../manager-tdd.md
```

**Edge cases**:
- Stub frontmatter has `tools:` non-empty → CI rejects (REQ-015 advisor-only ban applies in reverse — stubs MUST be empty-tools).
- Body contains anything beyond redirect text → not a hard fail but flagged.

---

## AC-ORC-001-06 — Template and local trees are byte-identical (REQ-005)

**Given** M2.7, M3.7, and M4.5 have all run `make build`
**When** I diff template against local
**Then** `diff -r` returns exit code 0 with no output.

**Verification**:

```bash
diff -r internal/template/templates/.claude/agents/moai/ .claude/agents/moai/
echo "Exit code: $?"
# Expected: empty output, exit code 0
```

**Edge cases**:
- A single-byte newline difference → fail (the gate is byte-identical, not semantically identical).
- Local file missing entirely → `diff` reports "Only in"; fail.

---

## AC-ORC-001-07 — Legacy SPEC reference routes via MIG-001 (REQ-007)

**Given** SPEC-V3R2-MIG-001 ships in the same release
**When** a user invokes `/moai run` against a legacy SPEC body containing
`Use the manager-ddd subagent`
**Then** MIG-001 rewriter substitutes `Use the manager-cycle subagent with
cycle_type=ddd` before the spawn prompt is dispatched, and the routing
succeeds.

**Verification** (deferred to MIG-001 integration test; ORC-001 verifies the
target shape only):

```bash
# Verify the rewrite-target agent exists and is responsive to cycle_type
grep -n "cycle_type" internal/template/templates/.claude/agents/moai/manager-cycle.md
# Expected: at least one match
```

**Edge cases**:
- MIG-001 not yet shipped → manual fallback: stub redirects in retired
  agent files keep legacy invocations working with one-cycle deprecation.

---

## AC-ORC-001-08 — manager-project rejects out-of-scope writes (REQ-008)

**Given** M3.2 has scope-shrunk manager-project
**When** I spawn manager-project with a task writing to
`.moai/config/sections/language.yaml`
**Then** manager-project returns a blocker report citing
`moai cc` / `moai glm` / `moai update` as the correct CLI alternative; no
write operation is performed.

**Verification**:

```bash
grep -A5 "Scope Boundary\|blocker report" internal/template/templates/.claude/agents/moai/manager-project.md | head -15
grep -n "settings_modification\|glm_configuration\|template_update_optimization" internal/template/templates/.claude/agents/moai/manager-project.md
# Expected: Scope Boundary section present; old modes absent
```

**Edge cases**:
- Old modes still listed but body says "deprecated" → partial pass; manager-spec flags for review.
- Scope Boundary missing → fail.

---

## AC-ORC-001-09 — expert-backend trigger row deduped to 12-15 EN tokens (REQ-010)

**Given** M3.3 has run
**When** I parse `expert-backend.md` EN trigger row
**Then**:
1. Token count is between 12 and 15 (inclusive).
2. No duplicate tokens within the EN row.
3. KO/JA/ZH rows do not contain standalone `Oracle` if the localized form is also present.

**Verification**:

```bash
# Count EN trigger tokens (rough heuristic)
grep "EN:" internal/template/templates/.claude/agents/moai/expert-backend.md | head -1 | tr ',' '\n' | wc -l
# Expected: 12-15

# Confirm Oracle reduction
grep -c "Oracle" internal/template/templates/.claude/agents/moai/expert-backend.md
# Expected: ≤ 4 (one per language row, not 4 per row)
```

**Edge cases**:
- Token count exactly 11 → off-by-one; manager-spec accepts as soft-fail with documentation.
- Duplicate `database` and `데이터베이스` (cross-language) → not a duplicate; only intra-row dedup applies.

---

## AC-ORC-001-10 — Context7 absent from manager-git and manager-quality (REQ-011)

**Given** M3.1 has run for manager-quality (manager-git already clean per
research.md §3.7)
**When** I grep tools list lines
**Then** `mcp__context7__*` strings appear 0 times in both files.

**Verification**:

```bash
grep -c "context7" internal/template/templates/.claude/agents/moai/manager-git.md
grep -c "context7" internal/template/templates/.claude/agents/moai/manager-quality.md
# Expected: both return 0
```

**Edge cases**:
- Context7 present in body comment (not tools list) → still fails REQ-011 strictly; manager-spec decides.

---

## AC-ORC-001-11 — plan-auditor frontmatter contains memory: project (REQ-013)

**Given** M3.6 has run
**When** I read plan-auditor frontmatter (lines 1-20)
**Then** the line `memory: project` is present.

**Verification**:

```bash
sed -n '1,20p' internal/template/templates/.claude/agents/moai/plan-auditor.md | grep "memory: project"
# Expected: 1 match
```

**Edge cases**:
- `memory:` field present but value is `user` → fail (REQ-013 specifies `project`).
- Field present below frontmatter delimiter → not parsed by Claude Code; fail.

---

## AC-ORC-001-12 — Trigger-union preservation across all merges (REQ-009, REQ-017)

**Given** M2 (builder-platform creation) and the prior carry-over
(manager-cycle creation) are in place
**When** I compute the union of trigger keywords from all retired source
bodies and compare to the merged target's trigger row
**Then** every keyword from each source body appears in the target's trigger
row (within the same language tier EN/KO/JA/ZH).

**Verification** (per language, per merge):

```bash
# manager-cycle vs (manager-ddd retired stub source = research.md §3.1 historical citation)
# Note: source bodies are retired stubs; we use research.md baseline for this check.

# builder-platform vs builder-agent + builder-skill + builder-plugin (concrete here)
extract_triggers() {
  awk '/^  EN:/ {sub(/^  EN: /,""); print}' "$1" | tr ',' '\n' | sed 's/^ *//;s/ *$//' | sort -u
}
# Source union
SRC=$(extract_triggers internal/template/templates/.claude/agents/moai/builder-agent.md)
SRC+=$(extract_triggers internal/template/templates/.claude/agents/moai/builder-skill.md)
SRC+=$(extract_triggers internal/template/templates/.claude/agents/moai/builder-plugin.md)
# Target
TGT=$(extract_triggers internal/template/templates/.claude/agents/moai/builder-platform.md)
# Diff: any line in SRC not in TGT
comm -23 <(echo "$SRC" | sort -u) <(echo "$TGT" | sort -u)
# Expected: empty output (no missing keywords)
```

**Note**: After M2.2-M2.4 retire builder-{agent,skill,plugin} to stubs, the
source bodies will be unavailable for AC-12 verification at audit time. M1.1
captures a snapshot of source trigger rows to
`.moai/specs/SPEC-V3R2-ORC-001/trigger-source-snapshot.txt` BEFORE the
retirement runs, so AC-12 can be verified against the snapshot.

**Edge cases**:
- Trigger row in target uses different phrasing (e.g., `agent_creation` vs
  `create agent`) — REQ-017 lint considers these distinct → fail unless
  manager-spec documents the equivalence in spec.md HISTORY.

---

## AC-ORC-001-13 — Stub bodies properly redirect (REQ-006 follow-up)

**Given** M2.2-M2.6 stubs exist
**When** I read each stub body
**Then** body contains:
1. One-line "This agent has been retired; use {replacement} with {param}".
2. Migration Guide table mapping old → new invocation.
3. No body content beyond migration documentation.

**Verification**:

```bash
for stub in builder-agent builder-skill builder-plugin expert-debug expert-testing; do
  wc -l "internal/template/templates/.claude/agents/moai/${stub}.md"
done
# Expected: each ≤ 50 lines (matching manager-ddd.md / manager-tdd.md size)
```

---

## AC-ORC-001-14 — manager-project body retains only .moai/project/ writes (REQ-008 part 2)

**Given** M3.2 has run
**When** I grep manager-project body for write-target paths
**Then**:
1. References to `.moai/project/product.md`, `structure.md`, `tech.md` are present.
2. References to `.claude/settings.json`, `.moai/config/sections/`, GLM endpoints are absent.

**Verification**:

```bash
grep -E "\.moai/project/(product|structure|tech)\.md" internal/template/templates/.claude/agents/moai/manager-project.md
grep -E "settings\.json|sections/|glm" internal/template/templates/.claude/agents/moai/manager-project.md
# Expected: first command returns matches; second returns 0 (or only inside Scope Boundary deny-list)
```

---

## AC-ORC-001-15 — Advisor-only agent rejection (REQ-015)

**Given** REQ-015 declares CI must reject any future agent that has no
write-capable tool yet prescribes file-write side effects
**When** ORC-002 lint runs (deferred to that SPEC)
**Then** the rule is documented in this SPEC's spec.md §5.5 and traceability
table; no new agent introduced by ORC-001 violates the rule.

**Verification** (ORC-001 internal):

```bash
# Verify builder-platform tools list contains Write
grep -n "Write" internal/template/templates/.claude/agents/moai/builder-platform.md | head -1
# Expected: at least 1 match (in tools list)
```

**Edge cases**: Out of scope for ORC-001; full enforcement deferred to
ORC-002.

---

## AC-ORC-001-16 — No orphan agent deletions (REQ-016)

**Given** REQ-016 declares CI must fail on `ORC_AGENT_DELETE_WITHOUT_STUB`
**When** I diff `internal/template/templates/.claude/agents/moai/` from base
to PR head
**Then** every deleted file is replaced by a stub (same path, retired
frontmatter); no file is fully removed.

**Verification**:

```bash
git diff --name-status origin/main...HEAD -- 'internal/template/templates/.claude/agents/moai/*.md' | awk '$1=="D"'
# Expected: empty output (no deletions; only modifications)
```

---

## AC-ORC-001-17 — No silent trigger drops (REQ-017)

**Given** REQ-017 declares CI fails on dropped keywords
**When** the trigger-union test from AC-12 runs against snapshot
**Then** zero keywords are missing from any merged target.

**Verification**: same as AC-ORC-001-12 (acceptance check is shared).

---

## Traceability Matrix

### Forward (REQ → AC)

| REQ-ID | AC-ID(s) | Notes |
|--------|----------|-------|
| REQ-ORC-001-001 | AC-01 | Roster size |
| REQ-ORC-001-002 | AC-02, AC-12 | manager-cycle preservation |
| REQ-ORC-001-003 | AC-03, AC-12, AC-15 | builder-platform shape |
| REQ-ORC-001-004 | AC-04 | Diagnostic Sub-Mode |
| REQ-ORC-001-005 | AC-06 | Template-First |
| REQ-ORC-001-006 | AC-05, AC-13 | Retired stub schema + content |
| REQ-ORC-001-007 | AC-07 | MIG-001 dependency (verified at integration) |
| REQ-ORC-001-008 | AC-08, AC-14 | manager-project scope shrink |
| REQ-ORC-001-009 | AC-12 | Trigger union |
| REQ-ORC-001-010 | AC-09 | Trigger dedup |
| REQ-ORC-001-011 | AC-10 | Context7 audit |
| REQ-ORC-001-012 | AC-11 (informative — Optional) | manager-quality memory (also verified) |
| REQ-ORC-001-013 | AC-11 | plan-auditor memory |
| REQ-ORC-001-014 | AC-15 | expert-performance Write |
| REQ-ORC-001-015 | AC-15 | Advisor-only rejection (deferred to ORC-002) |
| REQ-ORC-001-016 | AC-16 | No orphan deletions |
| REQ-ORC-001-017 | AC-12, AC-17 | Trigger-drop CI |

### Reverse (AC → REQ)

| AC-ID | REQ-ID(s) | Coverage |
|-------|-----------|----------|
| AC-01 | REQ-001 | Direct |
| AC-02 | REQ-002 | Direct |
| AC-03 | REQ-003 | Direct |
| AC-04 | REQ-004 | Direct |
| AC-05 | REQ-006 | Direct |
| AC-06 | REQ-005 | Direct |
| AC-07 | REQ-007 | Direct (deferred to MIG-001) |
| AC-08 | REQ-008 | Primary |
| AC-09 | REQ-010 | Direct |
| AC-10 | REQ-011 | Direct |
| AC-11 | REQ-012, REQ-013 | Both |
| AC-12 | REQ-002, REQ-003, REQ-009, REQ-017 | Multi-coverage (union test) |
| AC-13 | REQ-006 | Stub content quality |
| AC-14 | REQ-008 | manager-project allowlist |
| AC-15 | REQ-014, REQ-015 | Optional Write + advisor-only |
| AC-16 | REQ-016 | Direct |
| AC-17 | REQ-017 | Direct |

**Coverage**: 17 REQs → 17 ACs; every REQ has at least 1 AC. Some ACs cover
multiple REQs (consolidating verification effort).

---

## Definition of Done

The SPEC is "done" (audit-ready) when all the following are true:

- [ ] All 17 ACs (AC-01 through AC-17) PASS per their verification commands.
- [ ] `diff -r` template↔local empty (AC-06).
- [ ] `make build` exit 0; `make test` exit 0; `golangci-lint run` exit 0.
- [ ] No remaining literal `manager-ddd|manager-tdd|builder-{agent,skill,plugin}|expert-debug|expert-testing` outside stub bodies and migration tables (M4 sweep complete).
- [ ] CLAUDE.md Agent Catalog block updated (M4.3).
- [ ] spec.md §10.1 destiny table extended for post-R5 additions (M5.2).
- [ ] All 5 mx_plan tags resolved on actual code lines (M5.1).
- [ ] PR opened with `plan-auditor` review request.

---

End of acceptance.md.
