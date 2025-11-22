# Skill Consolidation Implementation Checklist

## Pre-Implementation (Day 0)

### Preparation
- [ ] Review full consolidation report
- [ ] Backup current `.claude/skills/` directory
- [ ] Create archive directory: `.claude/skills/_archived/`
- [ ] Set up testing environment
- [ ] Document all current skill references in codebase

---

## Week 1: High Priority Consolidations

### Consolidation 1: Claude Code Skills (Days 1-2)

#### Day 1: Setup & Migration
- [ ] Create `moai-cc-skills/` directory structure
- [ ] Create `modules/` subdirectory
- [ ] Write unified `SKILL.md` (350-400 lines)
  - [ ] Quick Reference section
  - [ ] Skill Architecture Overview
  - [ ] When to Use
  - [ ] Core Concepts
  - [ ] Best Practices
- [ ] Extract factory patterns → `modules/factory.md`
- [ ] Extract usage guidance → `modules/guide.md`
- [ ] Create `examples.md` (10+ examples)
- [ ] Create `reference.md` (official docs)

#### Day 2: Testing & Archive
- [ ] Test skill activation: `Skill("moai-cc-skills")`
- [ ] Verify all examples work
- [ ] Update cross-references in:
  - [ ] `.moai/memory/skills.md`
  - [ ] Other skills referencing factory/guide
- [ ] Move originals to `_archived/`:
  - [ ] `moai-cc-skill-factory/`
  - [ ] `moai-cc-skills-guide/`
- [ ] Git commit: "feat: Consolidate Claude Code skills management"

---

### Consolidation 2: MCP Integration (Days 3-4)

#### Day 3: Setup & Migration
- [ ] Enhance `moai-mcp-integration/` directory
- [ ] Create `modules/` subdirectory
- [ ] Update `SKILL.md` (400 lines)
  - [ ] MCP Protocol Overview
  - [ ] Tool Integration Patterns
  - [ ] Context7 Quick Start
  - [ ] Best Practices
- [ ] Create `modules/context7.md` (Context7 patterns, 350 lines)
- [ ] Create `modules/custom-tools.md` (custom MCP tools, 200 lines)
- [ ] Create unified `examples.md`
- [ ] Create unified `reference.md`

#### Day 4: Testing & Archive
- [ ] Test MCP tool integration
- [ ] Test Context7 specific patterns
- [ ] Update cross-references in:
  - [ ] All skills using Context7
  - [ ] `.moai/memory/mcp-integration.md`
- [ ] Move original to `_archived/`:
  - [ ] `moai-context7-integration/`
- [ ] Git commit: "feat: Consolidate MCP and Context7 integration"

---

### Consolidation 3: Agent Management (Days 5-7)

#### Day 5-6: Setup & Migration
- [ ] Create `moai-cc-agents/` directory structure
- [ ] Create `modules/` subdirectory
- [ ] Write unified `SKILL.md` (400 lines)
  - [ ] Agent Architecture
  - [ ] Lifecycle Management
  - [ ] Orchestration Patterns
  - [ ] Best Practices
- [ ] Create `modules/factory.md` (agent creation, 300 lines)
- [ ] Create `modules/lifecycle.md` (subagent patterns, 300 lines)
- [ ] Create `modules/coordination.md` (multi-agent, 200 lines)
- [ ] Create `examples.md` (agent workflows)
- [ ] Create `reference.md` (agent API)

#### Day 7: Testing & Archive
- [ ] Test agent creation workflows
- [ ] Test subagent lifecycle
- [ ] Test multi-agent coordination
- [ ] Update cross-references in:
  - [ ] `.moai/memory/agents.md`
  - [ ] Other skills referencing agents
- [ ] Move originals to `_archived/`:
  - [ ] `moai-cc-subagents-guide/`
  - [ ] `moai-core-agent-factory/`
- [ ] Git commit: "feat: Consolidate agent management system"
- [ ] **Week 1 Review**: Test all 3 consolidations together

---

## Week 2: Medium Priority Consolidations

### Consolidation 4: Security Web Application (Days 8-9)

#### Day 8: Setup & Migration
- [ ] Create `moai-security-webapp/` directory structure
- [ ] Create `modules/` subdirectory
- [ ] Write `SKILL.md` (400 lines)
  - [ ] Web Security Foundations
  - [ ] OWASP Integration
  - [ ] Authentication Overview
  - [ ] API Security Overview
- [ ] Create `modules/api.md` (API security, 300 lines)
- [ ] Create `modules/auth.md` (authentication, 350 lines)
- [ ] Create `modules/owasp.md` (OWASP Top 10, 400 lines)

#### Day 9: Testing & Archive
- [ ] Test web security patterns
- [ ] Test API security examples
- [ ] Test authentication workflows
- [ ] Update cross-references
- [ ] Move originals to `_archived/`:
  - [ ] `moai-security-api/`
  - [ ] `moai-security-auth/`
  - [ ] `moai-security-owasp/`
- [ ] Git commit: "feat: Consolidate web security skills"

---

### Consolidation 5: Security Identity (Day 10)

#### Day 10: Setup & Migration
- [ ] Enhance `moai-security-identity/` directory
- [ ] Create `modules/` subdirectory
- [ ] Update `SKILL.md` (450 lines)
- [ ] Create `modules/zero-trust.md` (400 lines)
- [ ] Create unified `examples.md`
- [ ] Test identity workflows
- [ ] Move original to `_archived/`:
  - [ ] `moai-security-zero-trust/`
- [ ] Git commit: "feat: Consolidate identity management"

---

### Consolidation 6: Documentation Toolkit (Day 11)

#### Day 11: Setup & Migration
- [ ] Enhance existing `moai-docs-toolkit/`
- [ ] Create `modules/` subdirectory
- [ ] Update `SKILL.md` (400 lines)
- [ ] Create `modules/generation.md` (300 lines)
- [ ] Create `modules/linting.md` (300 lines)
- [ ] Create `modules/validation.md` (300 lines)
- [ ] Test complete docs workflow
- [ ] Move originals to `_archived/`:
  - [ ] `moai-docs-generation/`
  - [ ] `moai-docs-linting/`
  - [ ] `moai-docs-validation/`
- [ ] Git commit: "feat: Consolidate documentation lifecycle"

---

### Consolidation 7: Project Configuration (Day 12)

#### Day 12: Setup & Migration
- [ ] Enhance `moai-project-config-manager/`
- [ ] Create `modules/` subdirectory
- [ ] Update `SKILL.md` (400 lines)
- [ ] Create `modules/schema.md` (350 lines)
- [ ] Test configuration workflows
- [ ] Move original to `_archived/`:
  - [ ] `moai-core-config-schema/`
- [ ] Git commit: "feat: Consolidate project configuration"

---

## Final Testing & Validation (Days 13-14)

### Day 13: Comprehensive Testing
- [ ] Test all 7 consolidated skills
- [ ] Verify skill activation
- [ ] Run all examples
- [ ] Check cross-references
- [ ] Validate modular structure
- [ ] Test progressive disclosure
- [ ] Performance check (token usage)

### Day 14: Documentation & Cleanup
- [ ] Update `.moai/memory/skills.md` with new structure
- [ ] Update `README.md` skill count
- [ ] Create migration guide for users
- [ ] Document backward compatibility notes
- [ ] Final git commit: "docs: Update skill documentation post-consolidation"
- [ ] Tag release: `v2.0.0-skill-consolidation`

---

## Post-Implementation (Week 3+)

### Monitoring Period (1 month)
- [ ] Monitor skill usage patterns
- [ ] Track user feedback
- [ ] Fix any broken references
- [ ] Update examples as needed
- [ ] Maintain archived skills for backward compatibility

### Final Cleanup (After 1 month)
- [ ] Review archived skills usage (should be 0)
- [ ] Permanently remove archived skills (or keep for history)
- [ ] Final documentation update
- [ ] Celebrate consolidation success!

---

## Success Validation

### Metrics to Track
- [ ] Skill count: 126 → 116 ✓
- [ ] Zero broken skill references ✓
- [ ] All tests passing ✓
- [ ] Token efficiency improved by 10-15% ✓
- [ ] User confusion reduced (feedback) ✓

### Quality Gates
- [ ] Each consolidated skill has complete Quick Reference
- [ ] Modular structure (SKILL.md + modules/) implemented
- [ ] 10+ working examples per skill
- [ ] Cross-references updated
- [ ] Archive directory created

---

**Status**: Ready for Implementation  
**Estimated Time**: 2-3 weeks (1 developer)  
**Risk Level**: Low (gradual migration with archives)  
**Review Date**: 2025-11-22
