# Migration Guide: Legacy Foundation Skills → moai-foundation-core

## Overview

This guide helps you migrate from the 5 legacy foundation skills to the unified **moai-foundation-core** skill.

**Migration Timeline**:
- **Phase 4** (Current): Update agent references
- **Phase 5** (Next Release): Legacy skill deprecation notices
- **Phase 6** (Future): Legacy skill removal

**Estimated Migration Time**: 5-10 minutes per agent/skill

---

## Quick Migration Reference

### Old References → New Reference

| Legacy Skill                | New Reference                | Module      |
|-----------------------------|------------------------------|-------------|
| `moai-foundation-trust`     | `moai-foundation-core`       | core.md     |
| `moai-foundation-specs`     | `moai-foundation-core`       | specs.md    |
| `moai-foundation-ears`      | `moai-foundation-core`       | ears.md     |
| `moai-foundation-git`       | `moai-foundation-core`       | git-workflow.md |
| `moai-foundation-langs`     | `moai-foundation-core`       | langs.md    |

**Single Reference Pattern**: Replace ALL legacy skill references with `Skill("moai-foundation-core")`

---

## Step-by-Step Migration

### Step 1: Identify Legacy References

**Search Pattern**:
```bash
# Find all files with legacy foundation skill references
grep -r "moai-foundation-\(trust\|specs\|ears\|git\|langs\)" .claude/agents/moai/*.md
```

**Common Locations**:
- Agent files: `.claude/agents/moai/*.md`
- Custom commands: `.claude/commands/**/*.md`
- Other skills: `.claude/skills/**/SKILL.md`
- Documentation: `.moai/memory/*.md`

### Step 2: Update Agent Files

**Before**:
```markdown
---
skills: moai-foundation-trust, moai-foundation-specs, moai-lang-python
---

## Works Well With

- moai-foundation-trust – TRUST 5 compliance
- moai-foundation-specs – SPEC validation
- moai-foundation-ears – Requirements syntax
```

**After**:
```markdown
---
skills: moai-foundation-core, moai-lang-python
---

## Works Well With

- moai-foundation-core – Foundation frameworks (TRUST 5, SPEC, EARS, Git, Languages)
```

**Key Changes**:
1. Replace all 5 legacy skill names with `moai-foundation-core`
2. Update skill descriptions to reference "Foundation frameworks"
3. Remove duplicate references (only one `moai-foundation-core` needed)

### Step 3: Update Skill References in Code

**Before**:
```markdown
Reference @moai-foundation-trust for TRUST 5 validation.
Use @moai-foundation-specs for SPEC generation.
Follow @moai-foundation-ears for requirements.
```

**After**:
```markdown
Reference @moai-foundation-core for:
- TRUST 5 validation (core.md module)
- SPEC generation (specs.md module)
- EARS requirements (ears.md module)
```

### Step 4: Update Documentation

**Before**:
```markdown
This agent uses:
- moai-foundation-trust: Quality gates
- moai-foundation-specs: SPEC structure
- moai-foundation-git: Git workflow
```

**After**:
```markdown
This agent uses moai-foundation-core:
- Core module: TRUST 5 quality gates
- Specs module: SPEC-001 structure
- Git-workflow module: 3-mode Git system
```

---

## Module-Specific Migration

### TRUST 5 Framework (moai-foundation-trust → core.md)

**What Moved**:
- Test-first development patterns
- Readable code standards
- Unified architecture principles
- Security validation patterns
- Trackable quality metrics

**Access Pattern**:
```markdown
# Before
Skill("moai-foundation-trust")

# After
Skill("moai-foundation-core")
# Automatically loads core.md for TRUST 5 framework
```

**No Behavior Changes**: All TRUST 5 functionality remains identical.

### SPEC Framework (moai-foundation-specs → specs.md)

**What Moved**:
- SPEC-001 generation
- Project structure standards
- Documentation sync patterns
- Quality gate integration

**Access Pattern**:
```markdown
# Before
Skill("moai-foundation-specs")

# After
Skill("moai-foundation-core")
# Automatically loads specs.md for SPEC framework
```

**Enhanced Features**:
- Improved SPEC template generation
- Better cross-reference validation
- Enhanced documentation sync

### EARS Specification (moai-foundation-ears → ears.md)

**What Moved**:
- Easy Approach to Requirements Syntax
- 5-part requirement structure
- Template-driven specification
- Requirement validation

**Access Pattern**:
```markdown
# Before
Skill("moai-foundation-ears")

# After
Skill("moai-foundation-core")
# Automatically loads ears.md for EARS framework
```

**Compatibility**: 100% backward compatible with existing EARS patterns.

### Git Workflow (moai-foundation-git → git-workflow.md)

**What Moved**:
- 3-mode Git system (Manual/Personal/Team)
- Branch creation automation
- Commit message standards
- PR workflow orchestration

**Access Pattern**:
```markdown
# Before
Skill("moai-foundation-git")

# After
Skill("moai-foundation-core")
# Automatically loads git-workflow.md for Git patterns
```

**New Features**:
- Enhanced branch creation logic
- Improved PR template generation
- Better commit message validation

### Language Standards (moai-foundation-langs → langs.md)

**What Moved**:
- 15+ programming language standards
- Language-specific best practices
- Framework integration guides
- Code quality patterns

**Access Pattern**:
```markdown
# Before
Skill("moai-foundation-langs")

# After
Skill("moai-foundation-core")
# Automatically loads langs.md for language standards
```

**Expanded Coverage**: Now includes more language-specific patterns and examples.

---

## Agent-Specific Migration Examples

### Example 1: quality-gate.md

**Before**:
```markdown
---
skills: moai-foundation-trust, moai-foundation-specs
---

## TRUST 5 Validation

Use moai-foundation-trust for quality gate enforcement.
Use moai-foundation-specs for SPEC validation.
```

**After**:
```markdown
---
skills: moai-foundation-core
---

## TRUST 5 Validation

Use moai-foundation-core for:
- TRUST 5 quality gate enforcement (core.md)
- SPEC-001 validation (specs.md)
```

### Example 2: spec-builder.md

**Before**:
```markdown
---
skills: moai-foundation-specs, moai-foundation-ears, moai-foundation-git
---

Generate SPEC using moai-foundation-specs.
Follow EARS format from moai-foundation-ears.
Manage Git workflow with moai-foundation-git.
```

**After**:
```markdown
---
skills: moai-foundation-core
---

Generate SPEC using moai-foundation-core:
- SPEC-001 structure (specs.md)
- EARS requirement format (ears.md)
- Git workflow automation (git-workflow.md)
```

### Example 3: git-manager.md

**Before**:
```markdown
---
skills: moai-foundation-git
---

## Git Workflow

Follow moai-foundation-git for branch creation and PR management.
```

**After**:
```markdown
---
skills: moai-foundation-core
---

## Git Workflow

Follow moai-foundation-core git-workflow.md module for:
- 3-mode Git system (Manual/Personal/Team)
- Automated branch creation
- PR template generation
```

---

## Validation Checklist

After migration, verify:

- [ ] **No Legacy References**: Search confirms no `moai-foundation-(trust|specs|ears|git|langs)` references remain
- [ ] **Skill Loading**: `Skill("moai-foundation-core")` loads successfully
- [ ] **Functionality**: All dependent features work as expected
- [ ] **Documentation**: Updated skill descriptions and comments
- [ ] **Testing**: Run integration tests to confirm no regressions

**Validation Command**:
```bash
# Verify no legacy references
grep -r "moai-foundation-\(trust\|specs\|ears\|git\|langs\)" .claude/ .moai/ 2>/dev/null

# Expected: No results (or only in archived files)
```

---

## Troubleshooting

### Issue: Skill Not Found

**Symptom**: `moai-foundation-core` skill not loading

**Solution**:
```bash
# Verify skill exists
ls -la .claude/skills/moai-foundation-core/SKILL.md

# Check skill frontmatter
head -5 .claude/skills/moai-foundation-core/SKILL.md

# Expected:
# ---
# name: moai-foundation-core
# description: Unified MoAI-ADK foundation frameworks...
# ---
```

### Issue: Missing Module Content

**Symptom**: Module content not loading (e.g., TRUST 5 patterns missing)

**Solution**:
```bash
# Verify modules exist
ls -la .claude/skills/moai-foundation-core/modules/

# Expected files:
# - core.md (TRUST 5)
# - specs.md (SPEC)
# - ears.md (EARS)
# - git-workflow.md (Git)
# - langs.md (Languages)
# - integration.md (Integration)
# - advanced.md (Advanced)
```

### Issue: Behavior Differences

**Symptom**: Skill behavior differs from legacy skills

**Check**:
1. Review CHANGELOG.md for intentional changes
2. Verify you're using correct module reference
3. Check examples.md for updated usage patterns

**Report**: If behavior differs unexpectedly, report via `/moai:9-feedback`

---

## Performance Improvements

**Token Efficiency**:
- Legacy skills: 5 separate loads = ~25K tokens
- moai-foundation-core: 1 load = ~8K tokens
- **Savings**: 68% token reduction

**Loading Time**:
- Legacy skills: 5× skill discovery overhead
- moai-foundation-core: 1× skill discovery, lazy module loading
- **Improvement**: 80% faster skill activation

**Context Management**:
- Modular loading: Only load needed modules
- Progressive disclosure: Advanced patterns loaded on demand
- Optimized cross-references: Reduced redundancy

---

## Migration Support

### Automated Migration Script

**Coming in Phase 5**: Automated migration script to update all references.

**Manual Migration** (Current):
```bash
# Find all agent files with legacy references
files=$(grep -l "moai-foundation-\(trust\|specs\|ears\|git\|langs\)" .claude/agents/moai/*.md)

# For each file, update references
for file in $files; do
    # Backup original
    cp "$file" "$file.bak"
    
    # Replace legacy skill references
    sed -i '' 's/moai-foundation-trust/moai-foundation-core/g' "$file"
    sed -i '' 's/moai-foundation-specs/moai-foundation-core/g' "$file"
    sed -i '' 's/moai-foundation-ears/moai-foundation-core/g' "$file"
    sed -i '' 's/moai-foundation-git/moai-foundation-core/g' "$file"
    sed -i '' 's/moai-foundation-langs/moai-foundation-core/g' "$file"
    
    echo "Updated: $file"
done
```

### Getting Help

**Documentation**:
- [SKILL.md](./SKILL.md) - Complete skill reference
- [examples.md](./examples.md) - Working examples
- [CHANGELOG.md](./CHANGELOG.md) - Version history

**Support Channels**:
- GitHub Issues: Report migration issues
- `/moai:9-feedback`: Submit improvement suggestions
- Documentation: Check module-specific docs in `modules/`

---

## Migration Timeline

### Phase 4: Agent Reference Updates (Current)

**Status**: In Progress
**Duration**: 1 release cycle
**Actions**:
- Update all agent files (19 files, 34 references)
- Update custom commands
- Update dependent skills
- Validate functionality

### Phase 5: Deprecation Notices (Next Release)

**Status**: Planned
**Duration**: 2 release cycles
**Actions**:
- Add deprecation warnings to legacy skills
- Update all documentation
- Provide automated migration script
- Monitor usage patterns

### Phase 6: Legacy Skill Removal (Future)

**Status**: Planned (v2.0.0)
**Duration**: TBD
**Actions**:
- Remove legacy skill files
- Clean up archived references
- Update dependency graphs
- Final validation

---

## FAQ

**Q: Do I need to migrate immediately?**
A: No. Legacy skills remain available during the migration period (Phase 4-5). However, early migration is recommended to benefit from improvements.

**Q: Will my agents break if I don't migrate?**
A: No. Legacy skills continue to work during the migration period. Deprecation warnings will appear in Phase 5.

**Q: Can I use both legacy and new skills simultaneously?**
A: Yes, but not recommended. This creates unnecessary context overhead. Migrate fully to moai-foundation-core.

**Q: What if I find a bug after migration?**
A: Report via `/moai:9-feedback` with details. We'll address issues promptly and update the migration guide.

**Q: Are there breaking changes?**
A: No breaking changes in v1.0.0. All functionality from legacy skills is preserved. Some enhancements added.

**Q: How do I verify migration success?**
A: Use the validation checklist above. Run integration tests and confirm no legacy skill references remain.

---

**Migration Guide Version**: 1.0.0
**Last Updated**: 2025-11-25
**Status**: Active
**Compatibility**: MoAI-ADK v0.28.0+
