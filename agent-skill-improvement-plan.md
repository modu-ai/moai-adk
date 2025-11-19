# Agent & Skill Analysis and Improvement Plan

## Analysis Summary

### Analyzed Components

- **31 Agent Files** in `src/moai_adk/templates/.claude/agents/moai/`
- **139 Skill Directories** in `src/moai_adk/templates/.claude/skills/`
- **Core Documentation**: CLAUDE.md, moai-core-agent-guide, moai-core-practices, moai-cc-skills

---

## Key Findings

### 1. **Agent Skills Configuration Issues**

#### Problem: Inconsistent Skills Declaration

Many agents have `skills:` in YAML frontmatter but don't properly list or use relevant skills.

**Examples**:

- `git-manager.md`: `skills: []` (empty, but should reference git-related skills)
- `trust-checker.md`: `skills: []` (empty, but should reference moai-foundation-trust)
- `quality-gate.md`: Has skills listed but missing key skills like `moai-core-trust-validation`

#### Problem: Missing Critical Skills

Agents reference skills in their documentation but don't declare them in YAML frontmatter.

**Examples**:

- `project-manager.md`:
  - YAML: Only lists 2 skills (`moai-cc-configuration`, `moai-project-config-manager`)
  - Documentation references: `moai-core-language-detection`, `moai-project-documentation`, `moai-foundation-ears`, `moai-foundation-trust`, etc.
- `backend-expert.md`:

  - YAML: Lists 6 skills
  - Documentation references: `moai-essentials-security`, `moai-foundation-trust` (not in YAML)

- `quality-gate.md`:
  - YAML: Lists 4 skills (`moai-essentials-debug`, `moai-essentials-perf`, `moai-essentials-refactor`, `moai-domain-security`)
  - Documentation references: `moai-core-trust-validation`, `moai-core-tag-scanning`, `moai-essentials-review`, `moai-foundation-trust` (not in YAML)

### 2. **Outdated Command References**

#### Problem: `/alfred:*` vs `/moai:*` Inconsistency

Many agents still reference `/alfred:*` commands instead of `/moai:*`.

**Affected Agents**:

- `project-manager.md`: Line 3, 86, 88 - references `/alfred:0-project`
- `spec-builder.md`: Line 3, 32 - references `/alfred:1-plan`
- `tdd-implementer.md`: Line 3 - references `/alfred:2-run`
- `quality-gate.md`: Line 3 - references `/alfred:2-run`, `/alfred:3-sync`
- `git-manager.md`: Multiple references throughout

### 3. **Missing Progressive Disclosure Structure**

#### Problem: Agents Don't Follow Skill Best Practices

According to `moai-cc-skills` and `moai-core-practices`, skills should use:

- 3-level progressive disclosure (quick-patterns, scenarios, advanced)
- Metadata standards
- Clear "When to Use" sections

Most agents don't follow this structure consistently.

### 4. **Skills Not Following Official Format**

#### Problem: Skill SKILL.md Format Inconsistency

According to official documentation:

- Skills must start with YAML frontmatter
- Should include: name, version, created, updated, status, description, keywords, allowed-tools, stability
- Should use progressive disclosure pattern

**Need to verify**: Many skills in `/src/moai_adk/templates/.claude/skills/` for format compliance.

---

## Improvement Plan

### Phase 1: Fix Agent YAML Frontmatter Skills Declarations

**Goal**: Ensure all agents properly declare skills they reference.

#### 1.1 Update `git-manager.md`

- Current: `skills: []`
- **Add**:

  ```yaml
  skills:
    - moai-foundation-git
    - moai-core-git-workflow
  ```

#### 1.2 Update `trust-checker.md`

- Current: `skills: []`
- **Add**:

  ```yaml
  skills:
    - moai-foundation-trust
    - moai-core-trust-validation
  ```

#### 1.3 Update `quality-gate.md`

- Current:

  ```yaml
  skills:
    - moai-essentials-debug
    - moai-essentials-perf
    - moai-essentials-refactor
    - moai-domain-security
  ```

- **Add missing**:

  ```yaml
  skills:
    - moai-essentials-debug
    - moai-essentials-perf
    - moai-essentials-refactor
    - moai-domain-security
    - moai-core-trust-validation
    - moai-core-tag-scanning
    - moai-essentials-review
    - moai-foundation-trust
  ```

#### 1.4 Update `backend-expert.md`

- **Add missing**:

  ```yaml
  skills:
    - moai-lang-python
    - moai-lang-go
    - moai-domain-backend
    - moai-domain-database
    - moai-domain-api
    - moai-context7-lang-integration
    - moai-essentials-security
    - moai-foundation-trust
    - moai-core-language-detection
  ```

#### 1.5 Update `project-manager.md`

- **Add missing**:

  ```yaml
  skills:
    - moai-cc-configuration
    - moai-project-config-manager
    - moai-core-language-detection
    - moai-project-documentation
    - moai-project-language-initializer
    - moai-project-template-optimizer
    - moai-project-batch-questions
    - moai-foundation-ears
    - moai-foundation-langs
    - moai-core-tag-scanning
    - moai-core-trust-validation
  ```

#### 1.6 Update `spec-builder.md`

- **Add missing**:

  ```yaml
  skills:
    - moai-foundation-ears
    - moai-foundation-specs
    - moai-core-spec-authoring
    - moai-lang-python
    - moai-core-spec-metadata-validation
    - moai-core-tag-scanning
    - moai-foundation-trust
    - moai-core-trust-validation
  ```

### Phase 2: Update Command References from `/alfred:*` to `/moai:*`

**Goal**: Standardize all command references to `/moai:*` format.

#### 2.1 Files to Update

- `project-manager.md`
- `spec-builder.md`
- `tdd-implementer.md`
- `quality-gate.md`
- `git-manager.md`
- Any other agents with `/alfred:*` references

#### 2.2 Search and Replace Pattern

```bash
# Find all occurrences
grep -r "/alfred:" src/moai_adk/templates/.claude/agents/moai/

# Replace pattern:
/alfred:0-project → /moai:0-project
/alfred:1-plan → /moai:1-plan
/alfred:2-run → /moai:2-run
/alfred:3-sync → /moai:3-sync
/alfred:9-feedback → /moai:9-feedback
/alfred:99-release → /moai:99-release
```

### Phase 3: Verify and Fix Skill Format Compliance

**Goal**: Ensure all skills follow official YAML frontmatter format.

#### 3.1 Required YAML Frontmatter Fields

```yaml
---
name: skill-name
version: X.X.X
created: YYYY-MM-DD
updated: YYYY-MM-DD
status: stable|experimental|deprecated
description: Clear description of skill purpose
keywords:
  - keyword1
  - keyword2
allowed-tools:
  - Tool1
  - Tool2
stability: stable|experimental
---
```

#### 3.2 Skills to Verify (Sample Priority List)

1. `moai-core-feedback-templates` (recently integrated)
2. `moai-core-issue-labels`
3. `moai-core-trust-validation`
4. `moai-core-tag-scanning`
5. `moai-foundation-trust`
6. `moai-foundation-git`
7. `moai-core-git-workflow`

### Phase 4: Add Missing Skills

**Goal**: Create skills that are referenced but don't exist.

#### 4.1 Check for Missing Skills

Skills referenced in agents but potentially missing:

- `moai-core-spec-metadata-validation`
- `moai-core-ears-authoring`
- `moai-essentials-security` (vs `moai-domain-security`)
- `moai-domain-api` (may not exist, check if should be `moai-domain-web-api`)

### Phase 5: Documentation Improvements

**Goal**: Add "When to Use" sections and improve agent documentation structure.

#### 5.1 Add Standard Sections to All Agents

```markdown
## When to Use This Agent

- ✅ Use when: [specific trigger conditions]
- ❌ Don't use when: [wrong scenarios]
- 🔄 Delegate to: [alternative agents for edge cases]

## Related Agents

- [agent-name]: [relationship]

## Related Skills

- [skill-name]: [when loaded]
```

---

## Execution Priority

### Critical (P0) - Do First

1. **Fix Agent Skills YAML** (Phase 1) - Fixes immediate functionality issues
2. **Update Command References** (Phase 2) - Fixes broken references

### High (P1) - Do Next

3. **Verify Skill Format** (Phase 3.1, 3.2) - Ensures skills work correctly
4. **Check for Missing Skills** (Phase 4.1) - Identifies gaps

### Medium (P2) - After Core Fixes

5. **Add Missing Skills** (Phase 4) - Creates new skills if needed
6. **Documentation Improvements** (Phase 5) - Enhances usability

---

## Testing Plan

### After Each Phase

1. **Syntax Validation**: YAML frontmatter parses correctly
2. **Reference Validation**: All skill references resolve correctly
3. **Integration Testing**: Test agents with commands
4. **Documentation Review**: Ensure consistency

### Success Criteria

- ✅ All agents have complete skills declarations
- ✅ All `/alfred:*` references updated to `/moai:*`
- ✅ All skills follow official format
- ✅ No broken skill references
- ✅ Improved documentation clarity

---

## Next Steps

1. **Get User Approval** for this plan
2. **Execute Phase 1** (Critical - Fix Skills YAML)
3. **Execute Phase 2** (Critical - Update Command References)
4. **Verify and Test** before proceeding to next phases
