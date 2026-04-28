---
id: SPEC-KARPATHY-001
version: "0.1.0"
status: planned
created_at: 2026-04-28
updated_at: 2026-04-28
author: manager-spec
priority: High
labels: [karpathy, acceptance, criteria]
issue_number: null
---

# SPEC-KARPATHY-001 -- Acceptance Criteria

## 1. Definition of Done

This SPEC is complete (`status: completed`) when:

- AC-001 through AC-006 all PASS
- TRUST 5 quality gate passed (Tested / Readable / Unified / Secured / Trackable)
- `go test ./internal/template/...` green
- `make build` succeeds (embedded.go regenerated)
- Constitution evolvable zone verified (no frozen content modified)
- 16-language neutrality verified across all new files

---

## 2. Acceptance Criteria

### AC-001 -- Anti-Pattern Reference Skill (REQ-KP-004)

**Given** Wave 1 completed,
**When** `.claude/skills/moai/references/anti-patterns.md` is created,
**Then** all of the following MUST hold:

- File exists at the specified path
- Contains exactly 8 anti-pattern categories:
  1. Hidden Assumptions
  2. Multiple Interpretations
  3. Over-Abstraction
  4. Speculative Features
  5. Drive-by Refactoring
  6. Style Drift
  7. Vague Goals
  8. Test-First Execution
- Each category includes:
  - Mapped Agent Core Behavior reference
  - WRONG code example (Go primary)
  - RIGHT code example (Go primary)
  - Detection heuristic
- Frontmatter includes:
  - `user_invocable: false`
  - `trigger_agents` listing at least: expert-backend, expert-frontend, manager-quality, evaluator-active
  - `trigger_phases: ["run", "review"]`
  - Progressive disclosure configuration (level1_tokens, level2_tokens)
- At least 3 categories include secondary examples in Python or TypeScript
- No language-specific bias in descriptions (16-language neutrality)

**Verification commands**:
```bash
# File exists
test -f .claude/skills/moai/references/anti-patterns.md

# 8 categories present
grep -c "^## [0-9]\\." .claude/skills/moai/references/anti-patterns.md
# Expected: 8

# WRONG/RIGHT pairs present
grep -c "WRONG\\|RIGHT\\|Anti-pattern\\|Correct" .claude/skills/moai/references/anti-patterns.md
# Expected: >= 16 (at least 2 per category)

# Frontmatter progressive disclosure
grep -q "progressive_disclosure" .claude/skills/moai/references/anti-patterns.md
grep -q "user_invocable: false" .claude/skills/moai/references/anti-patterns.md
```

---

### AC-002 -- Quick Reference Rule (REQ-KP-005)

**Given** Wave 1 completed,
**When** `.claude/rules/moai/development/karpathy-quickref.md` is created,
**Then** all of the following MUST hold:

- File exists at the specified path
- Paths frontmatter covers at least 6 languages: `**/*.go,**/*.py,**/*.ts,**/*.js,**/*.java,**/*.rs`
- Contains mapping for all 4 Karpathy principles:
  1. Think Before Coding
  2. Simplicity First
  3. Surgical Changes
  4. Goal-Driven Execution
- Each principle maps to at least one Agent Core Behavior
- Each principle includes 3-5 concrete checkpoint questions
- Cross-reference to anti-pattern skill for detailed examples
- File size under 200 lines (quick reference constraint)

**Verification commands**:
```bash
# File exists
test -f .claude/rules/moai/development/karpathy-quickref.md

# Paths frontmatter covers 6 languages
head -5 .claude/rules/moai/development/karpathy-quickref.md | grep -c "\\*\\."
# Expected: >= 6 glob patterns

# 4 principles mapped
grep -c "Think Before Coding\\|Simplicity First\\|Surgical Changes\\|Goal-Driven" \
  .claude/rules/moai/development/karpathy-quickref.md
# Expected: >= 4

# Checkpoint questions present
grep -c "?\\]" .claude/rules/moai/development/karpathy-quickref.md
# Expected: >= 12 (at least 3 per principle)

# Line count
wc -l .claude/rules/moai/development/karpathy-quickref.md
# Expected: < 200
```

---

### AC-003 -- Constitution Amendments (REQ-KP-001, REQ-KP-002, REQ-KP-003)

**Given** Wave 2 completed,
**When** `moai-constitution.md` Agent Core Behaviors section is amended,
**Then** all of the following MUST hold:

- Behavior 4 (Enforce Simplicity) contains quantitative LOC trigger text
- Behavior 5 (Maintain Scope Discipline) contains style-matching directive text
- Behavior 6 (Verify, Don't Assume) contains goal-to-test pattern text
- Total lines added to constitution <= 20 (constraint compliance)
- `moai:evolvable-start id="agent-core-behaviors"` marker preserved
- `moai:evolvable-end` marker preserved
- No content outside the evolvable zone is modified
- Frozen zone entries (CONST-V3R2-025..046 in Zone Registry) remain valid

**Verification commands**:
```bash
# Amendment A: quantitative trigger in Behavior 4
grep -A 15 "### 4\\. Enforce Simplicity" .claude/rules/moai/core/moai-constitution.md | \
  grep -i "3x\\|three times\\|LOC\\|rewrite"
# Expected: at least 1 match

# Amendment B: style-matching in Behavior 5
grep -A 15 "### 5\\. Maintain Scope Discipline" .claude/rules/moai/core/moai-constitution.md | \
  grep -i "style\\|convention\\|naming\\|consistency"
# Expected: at least 1 match

# Amendment C: goal-to-test in Behavior 6
grep -A 15 "### 6\\. Verify, Don't Assume" .claude/rules/moai/core/moai-constitution.md | \
  grep -i "goal\\|test case\\|ad-hoc\\|transform"
# Expected: at least 1 match

# Evolvable zone markers preserved
grep -c "moai:evolvable-start.*agent-core-behaviors" .claude/rules/moai/core/moai-constitution.md
# Expected: 1
grep -c "moai:evolvable-end" .claude/rules/moai/core/moai-constitution.md
# Expected: 1 (or more if other evolvable zones exist)

# Total line count delta (pre/post amendment)
# This must be verified manually during Wave 2
```

---

### AC-004 -- Template-First Synchronization (REQ-KP-006)

**Given** Wave 3 completed,
**When** all new/modified files are mirrored to templates,
**Then** all of the following MUST hold:

- `internal/template/templates/.claude/skills/moai/references/anti-patterns.md` exists and is byte-identical to local copy
- `internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md` exists and is byte-identical to local copy
- `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` is synced with local copy
- `make build` succeeds with exit code 0
- `go test ./internal/template/...` all tests green

**Verification commands**:
```bash
# Byte-identical checks
diff .claude/skills/moai/references/anti-patterns.md \
     internal/template/templates/.claude/skills/moai/references/anti-patterns.md
# Expected: exit 0 (no diff)

diff .claude/rules/moai/development/karpathy-quickref.md \
     internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md
# Expected: exit 0 (no diff)

diff .claude/rules/moai/core/moai-constitution.md \
     internal/template/templates/.claude/rules/moai/core/moai-constitution.md
# Expected: exit 0 (no diff)

# Build + test
make build && go test ./internal/template/...
# Expected: all green
```

---

### AC-005 -- Workflow Chaining Integration (REQ-KP-007)

**Given** all Waves completed,
**When** the MoAI orchestration pipeline is active,
**Then** all of the following MUST hold:

- Constitution amendments are always loaded (no trigger required) -- verified by constitution file being always-loaded
- Quick reference rule loads when code files matching paths frontmatter are edited
- Anti-pattern skill loads when trigger keywords match AND trigger agents are active AND trigger phases match
- No new agent definitions are required for the integration
- No new SPEC workflow phases are required
- Existing agent chaining architecture (plan -> run -> sync) is preserved unchanged

**Verification commands**:
```bash
# No new agent files created
find .claude/agents/ -newer .moai/specs/SPEC-KARPATHY-001/spec.md -name "*.md" | wc -l
# Expected: 0

# Quick reference paths frontmatter valid
head -5 .claude/rules/moai/development/karpathy-quickref.md | grep -q "paths:"
# Expected: success

# Anti-pattern skill triggers configured
grep -q "trigger_agents" .claude/skills/moai/references/anti-patterns.md
grep -q "trigger_phases" .claude/skills/moai/references/anti-patterns.md
# Expected: success for both
```

---

### AC-006 -- 16-Language Neutrality

**Given** all Waves completed,
**When** all new files are validated for language neutrality,
**Then** all of the following MUST hold:

- No new file contains language-specific tool commands as canonical examples (e.g., "run pip install" as the only option)
- Code examples in anti-pattern skill use Go as primary, Python/TypeScript as secondary, clearly labeled
- Quick reference rule contains no code examples (pure principle-behavior mapping)
- No file declares a single language as "PRIMARY" or "recommended"
- File paths in paths frontmatter cover at least 6 languages equally

**Verification commands**:
```bash
# No single-language bias in descriptions
for f in .claude/skills/moai/references/anti-patterns.md \
         .claude/rules/moai/development/karpathy-quickref.md; do
  # Check for language-specific tool commands without alternatives
  grep -n "pip install\\|npm install\\|go get" "$f" | grep -v "secondary\\|example\\|alternative" && \
    echo "POTENTIAL BIAS in $f" || true
done

# Anti-pattern skill has multi-language examples
grep -c "python\\|typescript\\|Python\\|TypeScript" .claude/skills/moai/references/anti-patterns.md
# Expected: >= 3 (secondary examples in at least 3 categories)
```

---

## 3. Edge Cases

### EC-001 -- Constitution line count exceeds budget

**Detection**: Wave 2 post-amendment line count shows > 20 lines added.

**Response**:
1. Reduce each amendment to minimum viable text.
2. Combine overlapping text between amendments.
3. If still over budget, defer lowest-priority amendment (Amendment C is lowest priority as it targets Optional pattern).

### EC-002 -- Anti-pattern skill Level 2 exceeds token budget

**Detection**: Skill body word count exceeds ~5000 tokens (~3500 words).

**Response**:
1. Reduce secondary examples (Python/TypeScript) to pseudocode comments.
2. Shorten detection heuristics to keyword lists.
3. If still over budget, split into 2 skills (4 categories each).

### EC-003 -- Template build failure after sync

**Detection**: Wave 3 `make build` returns non-zero exit code.

**Response**:
1. Check `internal/template/embedded.go` regeneration errors.
2. Verify file paths match expected template directory structure.
3. Check for encoding issues (UTF-8 BOM, line endings).
4. Fix and retry (max 3 attempts).

### EC-004 -- Evolvable zone marker accidentally moved

**Detection**: `moai:evolvable-start` or `moai:evolvable-end` marker count changes post-amendment.

**Response**:
1. Restore markers to original positions.
2. Ensure amendments are inserted BEFORE `moai:evolvable-end`, not after it.
3. Verify Zone Registry entries still resolve correctly.

### EC-005 -- Skill trigger false-positive loading

**Detection**: Anti-pattern skill loads during unrelated contexts (e.g., documentation editing).

**Response**:
1. Tighten trigger keyword specificity.
2. Add negative triggers or scope restrictions if supported.
3. Verify trigger_agents list is restrictive enough.

---

## 4. Test Strategy

### Phase 1 -- Content Verification (Wave 1-2)

```bash
# Anti-pattern skill completeness
grep -c "^## [0-9]\\." .claude/skills/moai/references/anti-patterns.md
# Expected: 8

# Quick reference principle count
grep -c "Think Before\\|Simplicity First\\|Surgical Changes\\|Goal-Driven" \
  .claude/rules/moai/development/karpathy-quickref.md
# Expected: >= 4

# Constitution amendments present
grep -c "3x\\|three times" .claude/rules/moai/core/moai-constitution.md
grep -c "match.*style\\|existing.*convention" .claude/rules/moai/core/moai-constitution.md
grep -c "goal.*test\\|ad-hoc.*test" .claude/rules/moai/core/moai-constitution.md
# Expected: >= 1 each
```

### Phase 2 -- Template Build (Wave 3)

```bash
make build
# Expected: exit 0

go test ./internal/template/...
# Expected: all green
```

### Phase 3 -- Mirror Verification (Wave 3)

```bash
# Byte-identical check for all 3 files
diff .claude/skills/moai/references/anti-patterns.md \
     internal/template/templates/.claude/skills/moai/references/anti-patterns.md

diff .claude/rules/moai/development/karpathy-quickref.md \
     internal/template/templates/.claude/rules/moai/development/karpathy-quickref.md

diff .claude/rules/moai/core/moai-constitution.md \
     internal/template/templates/.claude/rules/moai/core/moai-constitution.md
# All expected: exit 0 (no diff)
```

### Phase 4 -- Neutrality and Constraint Verification (Wave 3-4)

```bash
# 16-language neutrality
for f in .claude/skills/moai/references/anti-patterns.md \
         .claude/rules/moai/development/karpathy-quickref.md; do
  grep -n "PRIMARY\\|recommended language" "$f" && echo "NEUTRALITY VIOLATION: $f" || true
done

# Constitution line count budget
# Pre-amendment baseline should be captured before Wave 2
# Post-amendment: total added lines <= 20

# Evolvable zone markers intact
grep -c "moai:evolvable-start.*agent-core-behaviors" .claude/rules/moai/core/moai-constitution.md
# Expected: 1
```

---

## 5. Quality Gate Criteria (TRUST 5)

- **Tested**: Markdown artifacts -- regression testing via `internal/template/...` existing test suite. Constitution amendment verified by grep-based content checks.
- **Readable**: Clear section headers, progressive disclosure structure, cross-references between skill and rule.
- **Unified**: Consistent formatting across new files, aligned with existing MoAI rule/skill conventions.
- **Secured**: No secrets, credentials, or sensitive information in new files. Constitution frozen zone untouched.
- **Trackable**: SPEC-KARPATHY-001 ID in commit messages, history table in each file, version tracking in frontmatter.
