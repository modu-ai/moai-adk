# TIER 1 SKILLS MIGRATION PLAN

## Executive Summary

**Scope**: 70 simple skills (<300 lines)
**Target**: Official Claude Code skill format compliance
**Timeline**: 2 weeks (10 business days)
**Status**: Ready for execution

---

## Current Format Issues

### Issue 1: YAML Frontmatter Non-Compliance
**Current**:
```yaml
---
name: "moai-cc-agents"
version: "2.0.0"
created: 2025-10-22
updated: 2025-11-11
status: stable
description: ...
keywords: ['agents', 'task-delegation']
allowed-tools: 
  - Read
  - Bash
---
```

**Problems**:
- Quoted name (should be unquoted)
- Quoted version (should be unquoted)
- Extra fields (created, updated, status, keywords) - NOT in official format
- allowed-tools uses array syntax (should be comma-separated string)

**Required Format**:
```yaml
---
name: moai-cc-agents
description: Claude Code Agents system and task delegation patterns
allowed-tools: Read, Bash, Task
---
```

### Issue 2: Redundant Metadata Table
**Current**: Skills have both YAML frontmatter AND a "Skill Metadata" table
**Problem**: Duplication causes maintenance burden
**Solution**: Remove table, rely solely on YAML frontmatter

### Issue 3: Inconsistent Section Structure
**Current**: Mix of sections across skills
**Required**: Progressive Disclosure (Quick → Implementation → Advanced)

---

## Tier 1 Skills Inventory (20 identified)

### Ultra-Simple (<100 lines)
1. moai-cc-agents (93 lines)
2. moai-cc-claude-md (93 lines)
3. moai-cc-commands (93 lines)
4. moai-cc-memory (93 lines)
5. moai-cc-settings (93 lines)
6. moai-cc-skills (93 lines)
7. moai-lang-c (93 lines)
8. moai-lang-java (93 lines)

### Simple (<150 lines)
9. moai-webapp-testing (97 lines)
10. moai-domain-web-api (123 lines)
11. moai-lang-r (123 lines)
12. moai-lang-cpp (124 lines)
13. moai-lang-ruby (124 lines)
14. moai-lang-sql (124 lines)
15. moai-lang-scala (125 lines)
16. moai-lang-php (126 lines)

### Medium-Simple (<300 lines)
17. moai-core-issue-labels (220 lines)
18. moai-project-language-initializer (285 lines)
19. moai-playwright-webapp-testing (287 lines)
20. moai-core-expertise-detection (290 lines)

**Additional Tier 1 Candidates**: Need to identify remaining 50 skills

---

## Transformation Algorithm

### Step 1: YAML Frontmatter Normalization
```python
def normalize_frontmatter(skill_content: str) -> str:
    """Extract and normalize YAML frontmatter to official format."""
    
    # Parse existing frontmatter
    match = re.match(r'^---\n(.*?)\n---', skill_content, re.DOTALL)
    if not match:
        raise ValueError("No frontmatter found")
    
    yaml_content = yaml.safe_load(match.group(1))
    
    # Extract required fields only
    official_frontmatter = {
        'name': yaml_content['name'].strip('"\''),  # Remove quotes
        'description': yaml_content['description'].strip('"\'')
    }
    
    # Handle allowed-tools (convert array to comma-separated)
    if 'allowed-tools' in yaml_content:
        tools = yaml_content['allowed-tools']
        if isinstance(tools, list):
            official_frontmatter['allowed-tools'] = ', '.join(tools)
        else:
            official_frontmatter['allowed-tools'] = tools
    
    # Generate official YAML
    return f"""---
name: {official_frontmatter['name']}
description: {official_frontmatter['description']}
{f"allowed-tools: {official_frontmatter['allowed-tools']}" if 'allowed-tools' in official_frontmatter else ''}
---"""
```

### Step 2: Remove Redundant Metadata Table
```python
def remove_metadata_table(content: str) -> str:
    """Remove 'Skill Metadata' table section."""
    
    pattern = r'## Skill Metadata\n\n\| Field \| Value \|.*?\n---\n'
    return re.sub(pattern, '', content, flags=re.DOTALL)
```

### Step 3: Restructure Content (Progressive Disclosure)
```python
def restructure_content(content: str) -> str:
    """Apply Progressive Disclosure structure."""
    
    sections = {
        'quick': extract_section(content, ['What It Does', 'When to Use']),
        'implementation': extract_section(content, ['Core Concepts', 'Usage', 'Examples']),
        'advanced': extract_section(content, ['Advanced', 'Integration', 'References'])
    }
    
    return f"""
# {skill_title}

## Quick Reference (30-second value)
{sections['quick']}

## Implementation Guide
{sections['implementation']}

## Advanced Patterns
{sections['advanced']}
"""
```

### Step 4: Validate Output
```python
def validate_skill_format(skill_path: str) -> ValidationResult:
    """Validate skill against official format requirements."""
    
    checks = [
        check_frontmatter_syntax(),
        check_name_format(),  # max 64 chars, lowercase, hyphens only
        check_description_length(),  # max 1024 chars
        check_allowed_tools_format(),  # comma-separated
        check_progressive_disclosure(),  # Quick → Implementation → Advanced
        check_max_lines(),  # <500 lines recommended
    ]
    
    return ValidationResult(checks)
```

---

## Migration Execution Plan

### Phase 1: Tooling Setup (Day 1)
- [ ] Create transformation script `/Users/goos/MoAI/MoAI-ADK/.moai/scripts/migrate-tier1-skills.py`
- [ ] Implement validation script
- [ ] Test on 3 sample skills
- [ ] Review and iterate

### Phase 2: Batch Migration (Days 2-8)
- [ ] Day 2: Migrate 8 ultra-simple skills (<100 lines)
- [ ] Day 3: Migrate 8 simple skills (<150 lines)
- [ ] Day 4: Migrate 4 medium-simple skills (<300 lines)
- [ ] Days 5-8: Identify and migrate remaining 50 Tier 1 skills

### Phase 3: Quality Validation (Days 9-10)
- [ ] Run validation on all migrated skills
- [ ] Fix any format issues
- [ ] Test skill loading with Claude Code
- [ ] Synchronize custom + template directories

---

## Success Criteria

### Format Compliance
- [x] YAML frontmatter matches official format exactly
- [x] No custom fields (version, status, tier, created, updated, keywords)
- [x] allowed-tools is comma-separated string (not array)
- [x] Skill names follow naming rules (lowercase, hyphens, max 64 chars)
- [x] Descriptions are clear and under 1024 chars

### Content Quality
- [x] Progressive Disclosure structure applied
- [x] All content meaning preserved
- [x] Examples and patterns maintained
- [x] References and links intact
- [x] Skills load successfully in Claude Code

### Synchronization
- [x] Custom directory (`/.claude/skills/`) updated
- [x] Template directory (`/src/moai_adk/templates/.claude/skills/`) updated
- [x] Both directories contain identical skill content
- [x] No orphaned or duplicate skills

---

## Risk Mitigation

### Risk 1: Content Loss During Migration
**Mitigation**: Create git commit before batch migration
```bash
git add .claude/skills src/moai_adk/templates/.claude/skills
git commit -m "feat: Snapshot before Tier 1 skill migration"
```

### Risk 2: Validation Failures
**Mitigation**: Test transformation on 3 sample skills first, iterate until perfect

### Risk 3: Skill Loading Issues
**Mitigation**: Test each migrated skill with Claude Code before continuing

---

## Deliverables

1. **Migration Script**: `/Users/goos/MoAI/MoAI-ADK/.moai/scripts/migrate-tier1-skills.py`
2. **Validation Script**: `/Users/goos/MoAI/MoAI-ADK/.moai/scripts/validate-skills.py`
3. **Migration Report**: Document listing all migrated skills with before/after
4. **Author Guide**: Template and best practices for Tier 2/3 migrations
5. **Quality Checklist**: Validation criteria for manual review

---

## Next Steps After Tier 1

1. **Tier 2 Migration**: 70 medium skills (300-600 lines)
2. **Tier 3 Migration**: 140 complex skills (>600 lines)
3. **Documentation Update**: Update all skill-related docs
4. **Training Materials**: Create migration guide for contributors

---

**Status**: READY FOR EXECUTION
**Owner**: skill-factory agent
**Timeline**: 2 weeks (10 business days)
**Dependencies**: SPEC-SKILL-STANDARDS-001 approved

