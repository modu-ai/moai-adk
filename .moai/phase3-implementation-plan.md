# Phase 3: Documentation Alignment (2-3 Weeks)

## Objective
Update existing documentation to align with official patterns, removing hallucinated components and correcting skill references.

## Tasks

### Task 1: Update Agent Documentation
**Fix agent skill references to match official patterns**:

1. **spec-builder** (already aligned, minor corrections needed)
   ```python
   # Current pattern from spec-builder.md - already correct
   Skill("moai-foundation-ears")()  # Always required

   # Remove incorrect references if any:
   # ❌ Skill("moai-essentials-debug")() - Non-existent
   # ✅ Use debug-helper agent instead
   ```

2. **cc-manager** (already correct in official docs)
   ```python
   # Already correct pattern from cc-manager.md
   Skill("moai-foundation-specs")()  # Always loaded
   Skill("moai-alfred-workflow")()   # Always loaded

   # Conditional skills (from cc-manager.md)
   Skill("moai-alfred-language-detection")()
   Skill("moai-foundation-tags")()
   Skill("moai-foundation-trust")()
   ```

3. **Other Agents** - Remove non-existent skill references
   - Remove any references to `moai-essentials-debug`
   - Remove any references to hallucinated JIT skills
   - Keep only skills that exist in `.claude/skills/`

**Deliverable**: Corrected agent documentation files

### Task 2: Update Command Documentation
**Remove non-existent JIT skill references**:

1. **alfred:1-plan** (remove hallucinated JIT skills)
   ```python
   # Remove from documentation:
   # ❌ Skill("moai-session-info")()
   # ❌ Skill("moai-jit-docs-enhanced")()

   # Keep existing working skills:
   # ✅ Skill("moai-foundation-specs")()
   # ✅ Skill("moai-foundation-ears")()
   ```

2. **alfred:2-run** (remove hallucinated JIT skills)
   ```python
   # Remove from documentation:
   # ❌ Skill("moai-streaming-ui")()
   # ❌ Skill("moai-change-logger")()
   # ❌ Skill("moai-tag-policy-validator")()
   ```

3. **alfred:3-sync** (remove hallucinated JIT skills)
   ```python
   # Remove from documentation:
   # ❌ Skill("moai-learning-optimizer")()
   # ❌ Skill("moai-jit-docs-enhanced")()
   ```

**Deliverable**: Corrected command documentation files

### Task 3: Create Reference Documentation
**Document ONLY verified, existing patterns**:

```python
# Document: Verified Skill Reference
VERIFIED_SKILLS = {
    "foundation": [
        "moai-foundation-specs",
        "moai-foundation-ears",
        "moai-foundation-tags",
        "moai-foundation-trust",
        "moai-foundation-git",
        "moai-foundation-langs"
    ],
    "alfred": [
        "moai-alfred-workflow",
        "moai-alfred-language-detection",
        "moai-alfred-git-workflow",
        "moai-alfred-spec-metadata-validation",
        "moai-alfred-ask-user-questions",
        # ... other verified Alfred skills
    ],
    "cc": [
        "moai-cc-agents",
        "moai-cc-commands",
        "moai-cc-skills",
        "moai-cc-hooks",
        "moai-cc-settings",
        "moai-cc-memory"
    ]
}

NON_EXISTENT_SKILLS = [
    "moai-session-info",
    "moai-jit-docs-enhanced",
    "moai-streaming-ui",
    "moai-change-logger",
    "moai-tag-policy-validator",
    "moai-learning-optimizer",
    "moai-essentials-debug",
    "moai-project-config-manager",
    "moai-project-template-optimizer"
]
```

**Deliverable**: `verified-skill-reference.md`

## Implementation Strategy

### Week 1: Agent Documentation Review
- Review each agent file for skill references
- Remove non-existent skills
- Align with official patterns

### Week 2: Command Documentation Review
- Review each command file for JIT skill references
- Remove hallucinated JIT skills
- Keep only verified skills

### Week 3: Reference Documentation
- Create comprehensive verified skill reference
- Document correct usage patterns
- Create checklist for validation

## Expected Outcomes
- All agent documentation uses only existing skills
- All command documentation removes hallucinated JIT skills
- Clear reference documentation of verified patterns
- No performance optimizations or caching (not in official docs)

## Success Criteria
- Zero references to non-existent skills
- All skill references verified against `.claude/skills/` directory
- No suggested implementations not in official docs
- Documentation matches actual codebase

## Exclusions (What NOT to include)
- ❌ Performance optimizations (not in official docs)
- ❌ Caching mechanisms (don't exist)
- ❌ Shared utility functions (not in architecture)
- ❌ Suggested improvements beyond documentation
- ❌ Any skills not found in `.claude/skills/`