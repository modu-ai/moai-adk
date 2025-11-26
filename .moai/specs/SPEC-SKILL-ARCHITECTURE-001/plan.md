---
id: SPEC-SKILL-ARCHITECTURE-001-PLAN
title: Implementation Plan - MoAI Skill Architecture Redesign
spec_id: SPEC-SKILL-ARCHITECTURE-001
created_at: 2025-11-24
updated_at: 2025-11-24
status: draft
---

# Implementation Plan: SPEC-SKILL-ARCHITECTURE-001

## 📋 Executive Summary

**Project**: Consolidate 136 fragmented skills into 40 autonomous, discoverable, and composable capability packages
**Timeline**: 12 weeks (3 phases × 4 weeks)
**Token Budget**: 120-150K tokens (within 250K limit)
**Team**: MoAI-ADK development team + 35 specialized agents
**Risk Level**: High (mitigated through phased approach + validation gates)

**Success Criteria**:
- Token efficiency: +70% improvement
- Agentic execution rate: ≥85%
- Discovery speed: p95 <500ms
- Zero orphaned skills: 100% utilization
- Test coverage: ≥85% per TRUST 5

---

## 🗓️ Timeline Overview

```
Week 1-4  (Phase 1): Design + Metadata Schema      [20K tokens]
Week 5-8  (Phase 2): Implementation                [60K tokens]
Week 9-12 (Phase 3): Migration + Validation        [40K tokens]
─────────────────────────────────────────────────────────────
Total: 12 weeks                                     [120K tokens]
Buffer: +2 weeks (contingency)                      [+30K tokens]
```

---

## 🎯 Phase 1: Design + Metadata Schema (Weeks 1-4)

### Week 1: Architecture Design

**Priority**: Critical
**Effort**: 8 story points
**Token Budget**: 5K tokens

**Milestones**:
1. ✅ Finalize 40-skill consolidation mapping (136 → 40)
2. ✅ Define 6-tier structure and naming conventions
3. ✅ Document skill composition patterns
4. ✅ Create migration strategy document

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Map 136 skills → 40 consolidated skills | spec-builder | 2 days | Current skill portfolio |
| Define 6-tier taxonomy | spec-builder | 1 day | Mapping complete |
| Document composition patterns | plan | 2 days | Taxonomy defined |
| Create migration strategy | plan | 1 day | All above tasks |

**Deliverables**:
- `docs/skill-consolidation-mapping.md` - 136 → 40 correspondence table
- `docs/6-tier-taxonomy.md` - Tier structure and naming rules
- `docs/composition-patterns.md` - Skill chaining patterns
- `docs/migration-strategy.md` - Phased rollout plan

**Validation Checklist**:
- [ ] All 136 skills mapped to 40 consolidated skills
- [ ] No function loss verified
- [ ] 6-tier naming conventions approved
- [ ] Composition patterns documented with examples
- [ ] Migration strategy reviewed by stakeholders

---

### Week 2: Metadata Schema Design

**Priority**: Critical
**Effort**: 8 story points
**Token Budget**: 5K tokens

**Milestones**:
1. ✅ Define extended metadata fields (16 total)
2. ✅ Create JSON Schema validation
3. ✅ Implement metadata compliance checker
4. ✅ Design backward compatibility alias system

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Define 16 metadata fields | spec-builder | 1 day | None |
| Create JSON Schema | backend-expert | 1 day | Field definitions |
| Implement compliance checker | backend-expert | 2 days | JSON Schema |
| Design alias system | plan | 2 days | Mapping from Week 1 |

**Deliverables**:
- `schemas/skill-metadata-v2.json` - Extended metadata JSON Schema
- `scripts/validate-skill-metadata.py` - Compliance checker script
- `docs/backward-compatibility-aliases.md` - Alias mapping 136 → 40
- `tests/test_metadata_compliance.py` - Automated validation tests

**Validation Checklist**:
- [ ] JSON Schema validates all 16 fields correctly
- [ ] Compliance checker passes on sample skills
- [ ] Alias system covers all 136 legacy skill names
- [ ] No conflicts in alias mappings

---

### Week 3: Semantic Router Design

**Priority**: High
**Effort**: 8 story points
**Token Budget**: 5K tokens

**Milestones**:
1. ✅ Select embedding model (text-embedding-3-small)
2. ✅ Design FAISS index structure
3. ✅ Define confidence threshold (≥0.75)
4. ✅ Implement fallback mechanism

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Evaluate embedding models | backend-expert | 1 day | None |
| Design FAISS index | backend-expert | 1 day | Model selected |
| Prototype semantic router | backend-expert | 2 days | FAISS design |
| Design fallback logic | plan | 1 day | Prototype tested |

**Deliverables**:
- `docs/semantic-router-architecture.md` - Router design document
- `prototypes/semantic-router-poc.py` - Proof of concept
- `docs/embedding-model-evaluation.md` - Model comparison
- `docs/fallback-strategy.md` - Manual selection fallback

**Validation Checklist**:
- [ ] Embedding model achieves ≥90% accuracy on test queries
- [ ] FAISS prototype completes queries in <500ms
- [ ] Confidence threshold (0.75) validated with sample data
- [ ] Fallback mechanism tested and approved

---

### Week 4: Skill Invocation API Design

**Priority**: High
**Effort**: 8 story points
**Token Budget**: 5K tokens

**Milestones**:
1. ✅ Define API interface (sequential, parallel, conditional)
2. ✅ Design error handling strategy
3. ✅ Create progressive disclosure framework
4. ✅ Validate with stakeholders

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Define API interface | api-designer | 2 days | None |
| Design error handling | backend-expert | 1 day | API interface |
| Create progressive disclosure | spec-builder | 2 days | None |
| Stakeholder review | plan | 1 day | All above tasks |

**Deliverables**:
- `docs/skill-invocation-api-spec.md` - API specification
- `docs/error-handling-strategy.md` - Error codes and recovery
- `docs/progressive-disclosure-framework.md` - 3-level structure
- `docs/api-design-review-notes.md` - Stakeholder feedback

**Validation Checklist**:
- [ ] API supports sequential, parallel, conditional execution
- [ ] Error handling covers all failure modes
- [ ] Progressive disclosure reduces tokens by ≥70%
- [ ] Stakeholder approval received

---

### Phase 1 Validation Gate (End of Week 4)

**Gate Criteria**:
- ✅ All design documents approved by stakeholders
- ✅ Metadata schema passes 100% validation
- ✅ API specification reviewed and approved
- ✅ Migration plan feasible within 12-week timeline

**Approval Required From**:
- Project Manager
- Technical Lead
- QA Lead
- Product Owner

**Next Phase**: Proceed to Phase 2 (Implementation) if gate passed

---

## 🛠️ Phase 2: Implementation (Weeks 5-8)

### Week 5: Foundation + Domain Tier Implementation

**Priority**: Critical
**Effort**: 13 story points
**Token Budget**: 15K tokens

**Milestones**:
1. ✅ Implement Tier 1 (Foundation): 3 skills
2. ✅ Implement Tier 2 (Domain): 12 skills
3. ✅ Total: 15 skills (37.5% of 40)

**Skills to Implement**:

**Tier 1 - Foundation (3)**:
- moai-foundation-ears
- moai-foundation-specs
- moai-foundation-trust

**Tier 2 - Domain (12)**:
- moai-domain-backend
- moai-domain-frontend
- moai-domain-database
- moai-domain-cloud
- moai-domain-cli
- moai-domain-mobile
- moai-domain-iot
- moai-domain-figma
- moai-domain-notion
- moai-domain-ml-ops
- moai-domain-monitoring
- moai-domain-devops

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Implement foundation-ears | spec-builder | 1 day | Metadata schema |
| Implement foundation-specs | spec-builder | 1 day | foundation-ears |
| Implement foundation-trust | quality-gate | 1 day | foundation-specs |
| Implement domain skills (12) | Multiple agents | 7 days | Foundation tier |
| Write tests for all 15 skills | test-engineer | 3 days | Implementation |

**Deliverables**:
- `.claude/skills/moai-foundation-ears/SKILL.md`
- `.claude/skills/moai-foundation-specs/SKILL.md`
- `.claude/skills/moai-foundation-trust/SKILL.md`
- `.claude/skills/moai-domain-*/SKILL.md` (12 files)
- `tests/test_tier1_foundation.py`
- `tests/test_tier2_domain.py`

**Validation Checklist**:
- [ ] All 15 skills implemented with metadata 100% compliant
- [ ] Test coverage ≥85% for each skill
- [ ] Progressive disclosure validated (Level 1 <500 chars)
- [ ] No function loss from original 136 skills

---

### Week 6: Quality + Language Tier Implementation

**Priority**: Critical
**Effort**: 13 story points
**Token Budget**: 15K tokens

**Milestones**:
1. ✅ Implement Tier 3 (Quality): 6 skills
2. ✅ Implement Tier 4 (Language): 8 skills
3. ✅ Total: 14 skills (35% of 40)

**Skills to Implement**:

**Tier 3 - Quality (6)**:
- moai-quality-testing
- moai-quality-security
- moai-quality-performance
- moai-quality-review
- moai-quality-debug
- moai-quality-refactor

**Tier 4 - Language (8)**:
- moai-lang-python
- moai-lang-typescript
- moai-lang-go
- moai-lang-rust
- moai-lang-kotlin
- moai-lang-java
- moai-lang-swift
- moai-lang-csharp

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Implement quality skills (6) | quality-gate | 3 days | Tier 1 complete |
| Implement language skills (8) | tdd-implementer | 4 days | Quality tier |
| Write tests for all 14 skills | test-engineer | 3 days | Implementation |

**Deliverables**:
- `.claude/skills/moai-quality-*/SKILL.md` (6 files)
- `.claude/skills/moai-lang-*/SKILL.md` (8 files)
- `tests/test_tier3_quality.py`
- `tests/test_tier4_language.py`

**Validation Checklist**:
- [ ] All 14 skills implemented with metadata 100% compliant
- [ ] Test coverage ≥85% for each skill
- [ ] Context7 integration working for language skills
- [ ] Quality skills reference TRUST 5 framework

---

### Week 7: BaaS + Specialized Tier Implementation

**Priority**: High
**Effort**: 10 story points
**Token Budget**: 15K tokens

**Milestones**:
1. ✅ Implement Tier 5 (BaaS): 6 skills
2. ✅ Implement Tier 6 (Specialized): 5 skills
3. ✅ Total: 11 skills (27.5% of 40)

**Skills to Implement**:

**Tier 5 - BaaS (6)**:
- moai-baas-vercel
- moai-baas-neon
- moai-baas-clerk
- moai-baas-supabase
- moai-baas-firebase
- moai-baas-cloudflare

**Tier 6 - Specialized (5)**:
- moai-specialized-context7
- moai-specialized-playwright
- moai-specialized-figma-api
- moai-specialized-notion-api
- moai-specialized-docs

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Implement BaaS skills (6) | backend-expert | 3 days | Domain tier |
| Implement specialized skills (5) | Multiple agents | 3 days | All tiers |
| Write tests for all 11 skills | test-engineer | 2 days | Implementation |

**Deliverables**:
- `.claude/skills/moai-baas-*/SKILL.md` (6 files)
- `.claude/skills/moai-specialized-*/SKILL.md` (5 files)
- `tests/test_tier5_baas.py`
- `tests/test_tier6_specialized.py`

**Validation Checklist**:
- [ ] All 11 skills implemented with metadata 100% compliant
- [ ] Test coverage ≥85% for each skill
- [ ] MCP integration tested for specialized skills
- [ ] BaaS skills include Context7 library references

---

### Week 8: Integration + Testing

**Priority**: Critical
**Effort**: 13 story points
**Token Budget**: 15K tokens

**Milestones**:
1. ✅ Implement semantic router (FAISS + embeddings)
2. ✅ Implement skill invocation API
3. ✅ Implement progressive disclosure system
4. ✅ Implement backward compatibility aliases

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Implement semantic router | backend-expert | 2 days | All 40 skills ready |
| Generate skill embeddings | backend-expert | 1 day | Router implemented |
| Implement invocation API | api-designer | 2 days | Router complete |
| Implement progressive disclosure | spec-builder | 1 day | All skills |
| Create alias mappings | plan | 1 day | 136 → 40 mapping |
| Integration testing | test-engineer | 3 days | All components |

**Deliverables**:
- `src/semantic_router.py` - Semantic router implementation
- `src/skill_invocation_api.py` - Invocation API
- `src/progressive_disclosure.py` - Disclosure system
- `data/skill-embeddings.pkl` - Pre-computed embeddings
- `data/legacy-aliases.json` - 136 → 40 alias mappings
- `tests/test_integration_semantic_router.py`
- `tests/test_integration_invocation_api.py`

**Validation Checklist**:
- [ ] Semantic router completes queries in <500ms (p95)
- [ ] Invocation API supports sequential, parallel, conditional
- [ ] Progressive disclosure reduces tokens by ≥70%
- [ ] All 136 aliases redirect correctly

---

### Phase 2 Validation Gate (End of Week 8)

**Gate Criteria**:
- ✅ All 40 skills implemented and documented
- ✅ Test coverage ≥85% for all skills
- ✅ Semantic router achieves <500ms p95 latency
- ✅ Token efficiency improved by ≥70%
- ✅ No regressions in existing functionality

**Testing Scope**:
- Unit tests: All 40 skills individually
- Integration tests: Semantic router + invocation API
- Performance tests: Latency and token usage benchmarks
- Regression tests: All 136 original functions preserved

**Approval Required From**:
- Technical Lead
- QA Lead
- Performance Engineer

**Next Phase**: Proceed to Phase 3 (Migration) if gate passed

---

## 🚀 Phase 3: Migration + Validation (Weeks 9-12)

### Week 9: Agent Migration

**Priority**: Critical
**Effort**: 10 story points
**Token Budget**: 10K tokens

**Milestones**:
1. ✅ Update all 35 agents to reference new 40-skill system
2. ✅ Test agent functionality with new skills
3. ✅ Validate agent-skill integration

**Agents to Migrate** (35 total):
- spec-builder
- plan
- tdd-implementer
- backend-expert
- frontend-expert
- database-expert
- security-expert
- quality-gate
- test-engineer
- debug-helper
- performance-engineer
- [... and 24 more agents]

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Migrate 5 agents (batch 1) | agent-factory | 1 day | Phase 2 complete |
| Test batch 1 agents | test-engineer | 1 day | Batch 1 migrated |
| Migrate 5 agents (batch 2-7) | agent-factory | 6 days | Previous batch tested |
| Final agent integration test | test-engineer | 2 days | All agents migrated |

**Deliverables**:
- Updated agent files (35 total)
- `tests/test_agent_migration.py` - Agent integration tests
- `docs/agent-migration-log.md` - Migration status tracker
- `docs/agent-skill-reference-matrix.md` - Agent → Skill mapping

**Validation Checklist**:
- [ ] All 35 agents migrated to new skill references
- [ ] No references to old 136-skill system
- [ ] Agent functionality tested and passing
- [ ] Agent-skill reference matrix validated

---

### Week 10: Backward Compatibility Testing

**Priority**: High
**Effort**: 8 story points
**Token Budget**: 10K tokens

**Milestones**:
1. ✅ Implement 136 → 40 alias mappings
2. ✅ Test legacy skill name redirects
3. ✅ Validate deprecation warnings
4. ✅ Set 6-month expiration for aliases

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Implement alias redirect system | backend-expert | 2 days | Alias mappings ready |
| Add deprecation warnings | backend-expert | 1 day | Redirect system |
| Test all 136 aliases | test-engineer | 3 days | Warnings added |
| Configure 6-month expiration | backend-expert | 1 day | Testing complete |

**Deliverables**:
- `src/legacy_skill_redirects.py` - Alias redirect implementation
- `data/alias-expiration-dates.json` - Expiration tracking
- `tests/test_legacy_aliases.py` - Comprehensive alias tests
- `docs/deprecation-timeline.md` - User communication plan

**Validation Checklist**:
- [ ] All 136 aliases redirect correctly to 40 skills
- [ ] Deprecation warnings displayed for all legacy names
- [ ] Expiration dates set to 2025-05-24 (6 months)
- [ ] User documentation updated

---

### Week 11: System Integration Testing

**Priority**: Critical
**Effort**: 13 story points
**Token Budget**: 10K tokens

**Milestones**:
1. ✅ End-to-end workflow testing (/moai:1-plan, 2-run, 3-sync)
2. ✅ Performance benchmarking (token usage, latency)
3. ✅ Agentic execution rate measurement (target: ≥85%)
4. ✅ Security and compliance validation

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| E2E workflow tests | test-engineer | 3 days | All components ready |
| Performance benchmarks | performance-engineer | 2 days | E2E tests passing |
| Agentic execution tests | quality-gate | 2 days | Performance done |
| Security audit | security-expert | 2 days | All tests complete |

**Deliverables**:
- `tests/test_e2e_workflows.py` - End-to-end test suite
- `benchmarks/token-efficiency-report.md` - Token usage analysis
- `benchmarks/latency-performance-report.md` - Latency analysis
- `benchmarks/agentic-execution-rate.md` - Autonomy metrics
- `audits/security-compliance-report.md` - Security findings

**Validation Checklist**:
- [ ] All /moai:1-plan, 2-run, 3-sync workflows passing
- [ ] Token efficiency ≥70% improvement validated
- [ ] Semantic router p95 latency <500ms confirmed
- [ ] Agentic execution rate ≥85% achieved
- [ ] No critical security vulnerabilities found

---

### Week 12: Documentation + Deployment

**Priority**: Critical
**Effort**: 10 story points
**Token Budget**: 10K tokens

**Milestones**:
1. ✅ Generate auto-documentation for all 40 skills
2. ✅ Create migration guide for users
3. ✅ Update CLAUDE.md and memory files
4. ✅ Deploy to production with monitoring

**Tasks**:

| Task | Owner | Duration | Dependencies |
|------|-------|----------|--------------|
| Generate skill documentation | docs-manager | 2 days | All skills finalized |
| Create migration guide | docs-manager | 1 day | Docs generated |
| Update CLAUDE.md | spec-builder | 1 day | Migration guide ready |
| Update memory files | spec-builder | 1 day | CLAUDE.md updated |
| Production deployment | devops-expert | 2 days | All docs complete |
| Setup monitoring | devops-expert | 1 day | Deployment done |

**Deliverables**:
- Auto-generated documentation for 40 skills
- `docs/migration-guide-v2.md` - User migration guide
- Updated `CLAUDE.md` with new skill system
- Updated `.moai/memory/*.md` files
- Production deployment artifacts
- Monitoring dashboard configuration

**Validation Checklist**:
- [ ] All 40 skills have auto-generated documentation
- [ ] Migration guide covers all user scenarios
- [ ] CLAUDE.md references new 40-skill system
- [ ] Memory files updated with new architecture
- [ ] Production deployment successful
- [ ] Monitoring active and collecting metrics

---

### Phase 3 Validation Gate (End of Week 12)

**Gate Criteria**:
- ✅ All 35 agents functional with new skill system
- ✅ Backward compatibility verified for 6-month period
- ✅ Agentic execution rate ≥85% measured
- ✅ Token efficiency +70% confirmed
- ✅ Zero function loss from 136 original skills
- ✅ Production monitoring active

**Final Acceptance**:
- Project Manager sign-off
- Technical Lead sign-off
- QA Lead sign-off
- Product Owner sign-off
- Security Expert sign-off

**Post-Deployment**:
- Monitor production metrics for 2 weeks
- Address any critical issues within 24 hours
- Collect user feedback
- Plan for future enhancements

---

## 📊 Resource Allocation

### Team Composition

| Role | Agent | Allocation | Weeks |
|------|-------|------------|-------|
| Project Lead | plan | 100% | 12 |
| SPEC Author | spec-builder | 100% | 12 |
| Backend Dev | backend-expert | 80% | 8 |
| API Designer | api-designer | 60% | 4 |
| Test Engineer | test-engineer | 100% | 12 |
| Quality Gate | quality-gate | 40% | 8 |
| DevOps | devops-expert | 20% | 4 |
| Docs Manager | docs-manager | 40% | 4 |
| Security Expert | security-expert | 20% | 2 |

### Token Budget Breakdown

| Phase | Weeks | Token Budget | Cumulative |
|-------|-------|--------------|------------|
| Phase 1 | 1-4 | 20K | 20K |
| Phase 2 | 5-8 | 60K | 80K |
| Phase 3 | 9-12 | 40K | 120K |
| Buffer | +2 | +30K | 150K |

### Risk Contingency

- **Timeline Buffer**: +2 weeks (extend to Week 14 if needed)
- **Token Buffer**: +30K tokens (up to 150K total)
- **Rollback Plan**: Maintain 136-skill system in parallel for 6 months

---

## 🎯 Success Metrics Tracking

### Weekly Progress Tracking

| Metric | Target | Week 4 | Week 8 | Week 12 |
|--------|--------|--------|--------|---------|
| Skills Implemented | 40 | 0 | 40 | 40 |
| Test Coverage | ≥85% | N/A | ≥85% | ≥85% |
| Token Efficiency | +70% | N/A | +70% | +70% |
| Agentic Execution | ≥85% | N/A | N/A | ≥85% |
| Agent Migration | 35/35 | 0 | 0 | 35 |
| Discovery Latency | <500ms | N/A | <500ms | <500ms |

### Daily Standup Focus

**Week 1-4**: Design completion, stakeholder alignment
**Week 5-8**: Skill implementation velocity, test coverage
**Week 9-12**: Agent migration progress, production readiness

---

## 🔗 Dependencies & Constraints

### External Dependencies

- **Context7 MCP**: Requires stable API access for library documentation
- **Sequential-Thinking MCP**: Requires active MCP server for complex reasoning
- **OpenAI API**: Requires text-embedding-3-small model access for semantic router
- **FAISS Library**: Requires FAISS ≥1.7.4 for vector search

### Internal Dependencies

- **SPEC-SKILL-PORTFOLIO-OPT-001**: Current skill portfolio optimization
- **Agent System**: 35 agents require coordination for migration
- **Git Strategy**: GitHub Flow 3-mode system must remain operational
- **Documentation System**: Auto-generation pipeline must be functional

### Technical Constraints

- **Token Budget**: 120-150K tokens (cannot exceed 250K hard limit)
- **Timeline**: 12 weeks (no extension beyond +2 week buffer)
- **Test Coverage**: ≥85% mandatory per TRUST 5 framework
- **Backward Compatibility**: 6-month alias expiration (non-negotiable)

---

## 📚 References

### Project Documents
- `SPEC-SKILL-ARCHITECTURE-001/spec.md` - Main specification
- `.moai/project/structure.md` - Current skill organization
- `.moai/project/tech.md` - Metadata standards
- `SPEC-SKILL-PORTFOLIO-OPT-001` - Portfolio optimization

### Technical References
- **FAISS Documentation**: https://github.com/facebookresearch/faiss
- **OpenAI Embeddings**: https://platform.openai.com/docs/guides/embeddings
- **EARS Methodology**: Easy Approach to Requirements Syntax
- **TRUST 5 Framework**: Test-first, Readable, Unified, Secured, Trackable

---

**Plan Status**: DRAFT ✅
**Next Step**: User approval to proceed with Phase 1 (Design)
**Last Updated**: 2025-11-24
**Author**: spec-builder agent
