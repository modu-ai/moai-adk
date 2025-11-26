---
id: SPEC-SKILL-ARCHITECTURE-001
title: MoAI Skill Architecture Redesign - Consolidate 136 Skills into 40 Autonomous Packages
status: draft
created_at: 2025-11-24
updated_at: 2025-11-24
priority: critical
effort: 13
version: 1.0.0
epic: SKILL-REDESIGN
domains:
  - architecture
  - skill-system
  - semantic-routing
  - api-design
dependencies:
  - SPEC-SKILL-PORTFOLIO-OPT-001
acceptance_difficulty: high
rollback_risk: high
risks: |
  - Function loss during skill consolidation (136 → 40)
  - Semantic router accuracy and false positive/negative matching
  - Agent integration breakage during migration
  - Token budget overrun (requires 120-150K tokens for full implementation)
  - Backward compatibility requirements for legacy skill references
tags:
  - skill-architecture
  - semantic-routing
  - autonomous-agents
  - skill-discovery
  - consolidation
---

# SPEC-SKILL-ARCHITECTURE-001: MoAI Skill Architecture Redesign

## 📋 Executive Summary

**Problem Statement**: Current MoAI-ADK skill portfolio contains 136 fragmented skills (127 tiered + 9 special) with inefficient discovery, high token consumption, and low agentic autonomy. Skills are manually invoked, lack composition patterns, and exhibit poor discoverability.

**Solution**: Consolidate 136 skills into 40 autonomous, discoverable, and composable capability packages organized in 6 tiers. Implement semantic routing for automatic skill discovery and skill invocation API for direct execution.

**Business Value**:
- **Token Efficiency**: +70% reduction in context loading (45-50K → 13-15K tokens per SPEC cycle)
- **Agentic Execution**: 85%+ autonomous skill selection (vs. current 0% autonomy)
- **Discovery Speed**: <500ms skill matching (vs. manual selection)
- **Zero Orphaned Skills**: 100% skill utilization through unified architecture
- **Developer Experience**: Clear skill boundaries, composition patterns, progressive disclosure

---

## 🌍 Environment (Context & Assumptions)

### Current System State

**Skill Portfolio Statistics (as of 2025-11-24)**:
- Total Skills: 136 (127 tiered + 9 special)
- Skill Organization: 10-tier categorization system
- Metadata Compliance: 100% (all skills have frontmatter)
- Agent References: 33/35 agents reference skills (94%)
- Total Keywords: 1,270 across all skills
- Average Skill File Size: 250 lines (within 500-line limit)

**Current Tier System (10 Tiers)**:
1. Tier 1 (moai-lang-*): 13 language-specific skills
2. Tier 2 (moai-domain-*): 13 domain-specific skills
3. Tier 3 (moai-security-*): 8 security skills
4. Tier 4 (moai-core-*): 8 core development skills
5. Tier 5 (moai-foundation-*): 5 foundation skills
6. Tier 6 (moai-cc-*): 7 Claude Code skills
7. Tier 7 (moai-baas-*): 10 BaaS integration skills
8. Tier 8 (moai-essentials-*): 6 essential development skills
9. Tier 9 (moai-project-*): 4 project management skills
10. Tier 10 (moai-lib-*): 1 library integration skill

**Special Skills**: 20 (documentation, design systems, specialized tools)

### Technical Constraints

**Platform**: Claude Code with MCP integration (Context7, Sequential-Thinking)
**Token Budget**: 250K tokens per feature implementation (SPEC + TDD + Docs)
**Test Coverage**: ≥85% per TRUST 5 framework
**Skill File Size**: ≤500 lines per file (progressive disclosure)
**Semantic Embedding Model**: text-embedding-3-small (OpenAI, 1536 dimensions)
**Agent System**: 35 specialized agents requiring skill references

### Assumptions

1. **Backward Compatibility**: Legacy skill names must work via alias system during 6-month transition period
2. **No Function Loss**: All 136 current skill functions must be preserved in 40 consolidated skills
3. **Agent Integration**: All 35 agents must be updated to reference new skill architecture
4. **Documentation**: Auto-generated documentation from metadata (no manual docs)
5. **Migration Path**: Phased rollout with validation gates at each stage

---

## ⚙️ Requirements (EARS Format)

### Functional Requirements (System Behavior)

#### REQ-001: Skill Consolidation (Universal)
**SPEC**: The skill system SHALL consolidate 136 existing skills into exactly 40 capability packages organized in 6 semantic tiers.

**Tier Structure (40 Total Skills)**:
- **Tier 1 - Foundation (3 skills)**: moai-foundation-{ears, specs, trust}
- **Tier 2 - Domain (12 skills)**: moai-domain-{backend, frontend, database, cloud, cli, mobile, iot, figma, notion, ml-ops, monitoring, devops}
- **Tier 3 - Quality (6 skills)**: moai-quality-{testing, security, performance, review, debug, refactor}
- **Tier 4 - Language (8 skills)**: moai-lang-{python, typescript, go, rust, kotlin, java, swift, csharp}
- **Tier 5 - BaaS (6 skills)**: moai-baas-{vercel, neon, clerk, supabase, firebase, cloudflare}
- **Tier 6 - Specialized (5 skills)**: moai-specialized-{context7, playwright, figma-api, notion-api, docs}

**Related Tests**:
- `test_skill_count_exactly_40()`
- `test_all_tier_skills_present()`
- `test_no_duplicate_skill_ids()`
- `test_all_136_functions_preserved()`

#### REQ-002: Semantic Router (Conditional)
**SPEC**: If a user request contains skill-related keywords, then the semantic router SHALL automatically identify the top-3 relevant skills with confidence scores ≥0.75 within 500ms.

**Router Components**:
1. **Embedding Model**: text-embedding-3-small (1536-dim vectors)
2. **Vector Database**: In-memory FAISS index (40 skill embeddings)
3. **Threshold**: Confidence ≥0.75 for auto-selection
4. **Fallback**: If confidence <0.75, prompt user for manual selection

**Related Tests**:
- `test_semantic_router_top3_skills()`
- `test_router_latency_under_500ms()`
- `test_confidence_threshold_filtering()`
- `test_fallback_to_manual_selection()`

#### REQ-003: Skill Invocation API (Universal)
**SPEC**: The skill system SHALL provide a standardized invocation API supporting sequential execution, parallel execution, and conditional branching patterns.

**API Interface**:
```python
# Sequential execution
result = await invoke_skill("moai-lang-python", context={"task": "optimize async code"})

# Parallel execution
results = await invoke_skills_parallel([
    ("moai-domain-backend", {"design": "REST API"}),
    ("moai-domain-database", {"schema": "PostgreSQL"}),
    ("moai-quality-security", {"validate": "OWASP Top 10"})
])

# Conditional execution
if context["language"] == "python":
    result = await invoke_skill("moai-lang-python", context)
else:
    result = await invoke_skill("moai-lang-typescript", context)
```

**Related Tests**:
- `test_sequential_invocation()`
- `test_parallel_invocation()`
- `test_conditional_invocation()`
- `test_invocation_error_handling()`

#### REQ-004: Progressive Disclosure (Boundary Condition)
**SPEC**: Each skill MUST provide Level 1 quick reference (<500 characters), Level 2 implementation guide (<2000 characters), and Level 3 advanced patterns (<5000 characters) for token-efficient context loading.

**Disclosure Levels**:
- **Level 1** (Quick Reference): 30-second scan, <500 chars
- **Level 2** (Implementation Guide): 5-minute read, <2000 chars
- **Level 3** (Advanced Patterns): 10+ minute deep dive, <5000 chars

**Related Tests**:
- `test_level1_under_500_chars()`
- `test_level2_under_2000_chars()`
- `test_level3_under_5000_chars()`
- `test_progressive_loading_saves_tokens()`

#### REQ-005: Autonomous Composition (Stakeholder)
**SPEC**: As an agentic system, the skill architecture SHOULD enable autonomous skill composition patterns where agents can automatically chain multiple skills (e.g., domain-backend → quality-security → quality-testing) without explicit user instruction.

**Composition Patterns**:
1. **Sequential Chaining**: backend → security → testing
2. **Parallel Execution**: frontend + backend + database (concurrent)
3. **Conditional Branching**: if python then lang-python else lang-typescript
4. **Error Recovery**: if skill fails, fallback to alternative skill

**Related Tests**:
- `test_sequential_skill_chaining()`
- `test_parallel_skill_execution()`
- `test_conditional_skill_branching()`
- `test_skill_error_recovery()`

---

### Non-Functional Requirements (Quality Attributes)

#### REQ-006: Token Efficiency (Performance)
**SPEC**: The new skill architecture SHALL reduce token consumption by at least 70% compared to current system (45-50K tokens → 13-15K tokens per SPEC creation cycle).

**Token Reduction Strategies**:
- Progressive disclosure (load Level 1 only by default)
- Consolidated skills (40 vs. 136 eliminates duplication)
- Semantic routing (auto-select relevant skills only)
- Metadata-driven discovery (avoid loading full skill content)

**Related Tests**:
- `test_token_usage_under_15k_per_spec()`
- `test_progressive_disclosure_saves_tokens()`
- `test_semantic_routing_reduces_tokens()`

#### REQ-007: Discovery Speed (Performance)
**SPEC**: The semantic router SHALL complete skill discovery and return top-3 matches in <500ms for 95th percentile requests.

**Performance Targets**:
- p50 latency: <200ms
- p95 latency: <500ms
- p99 latency: <1000ms
- Throughput: ≥100 requests/second

**Related Tests**:
- `test_router_p50_latency_under_200ms()`
- `test_router_p95_latency_under_500ms()`
- `test_router_throughput_100_rps()`

#### REQ-008: Agentic Execution Rate (Reliability)
**SPEC**: The skill system SHALL achieve ≥85% autonomous skill selection rate, meaning agents can correctly identify required skills without user intervention in 85%+ of requests.

**Measurement Method**:
- Track skill selection accuracy over 1000 test requests
- Success = correct skill selected + confidence ≥0.75
- Failure = incorrect skill or confidence <0.75 (manual fallback)

**Related Tests**:
- `test_autonomous_selection_accuracy_85_percent()`
- `test_false_positive_rate_under_5_percent()`
- `test_false_negative_rate_under_10_percent()`

#### REQ-009: Zero Orphaned Skills (Quality)
**SPEC**: The new architecture SHALL eliminate all orphaned skills by ensuring 100% of the 40 consolidated skills are referenced by at least one agent and have documented use cases.

**Verification**:
- Each skill must have ≥1 agent reference
- Each skill must have ≥3 documented use cases
- Each skill must have ≥5 auto-trigger keywords
- Each skill must have test coverage ≥85%

**Related Tests**:
- `test_all_skills_have_agent_references()`
- `test_all_skills_have_use_cases()`
- `test_all_skills_have_keywords()`
- `test_all_skills_test_coverage_85_percent()`

---

### Interface Requirements (Integration Points)

#### REQ-010: Agent Integration (Universal)
**SPEC**: All 35 specialized agents SHALL update their skill references to point to the new 40-skill architecture within the migration period.

**Agent Update Checklist** (per agent):
- [ ] Replace old skill names with new consolidated skill names
- [ ] Update skill invocation calls to use new API
- [ ] Test agent functionality with new skills
- [ ] Update agent documentation

**Related Tests**:
- `test_all_35_agents_updated()`
- `test_agent_skill_references_valid()`
- `test_agent_invocation_api_usage()`

#### REQ-011: Backward Compatibility Aliases (Conditional)
**SPEC**: If a legacy skill name (from 136 skills) is invoked, then the system SHALL automatically redirect to the corresponding consolidated skill (from 40 skills) with a deprecation warning for 6 months.

**Alias Mapping Example**:
```yaml
# Legacy skill → New skill
moai-lang-python-async: moai-lang-python  # Consolidated
moai-domain-rest-api: moai-domain-backend  # Merged
moai-security-auth: moai-quality-security  # Reorganized
```

**Related Tests**:
- `test_legacy_skill_redirect()`
- `test_deprecation_warning_shown()`
- `test_alias_expires_after_6_months()`

#### REQ-012: MCP Integration (Universal)
**SPEC**: The skill system SHALL integrate with Context7 MCP for real-time library documentation and Sequential-Thinking MCP for complex reasoning tasks.

**MCP Usage Patterns**:
- **Context7**: Fetch latest API docs for language/framework skills
- **Sequential-Thinking**: Decompose complex multi-skill workflows
- **Error Handling**: Graceful degradation if MCP unavailable

**Related Tests**:
- `test_context7_integration()`
- `test_sequential_thinking_integration()`
- `test_mcp_graceful_degradation()`

---

### Design Constraints (Technical Limitations)

#### REQ-013: Skill File Size (Universal)
**SPEC**: Each skill file SHALL NOT exceed 500 lines (progressive disclosure enforcement) with modularization for skills requiring >500 lines.

**Enforcement**:
- Pre-commit hook validates file size
- CI/CD pipeline blocks merge if size >500 lines
- Modularization required if base skill >300 lines

**Related Tests**:
- `test_all_skills_under_500_lines()`
- `test_modularized_skills_marked_correctly()`
- `test_pre_commit_hook_blocks_oversized_files()`

#### REQ-014: Metadata Compliance (Universal)
**SPEC**: All 40 skills SHALL achieve 100% metadata compliance including all 7 required fields (name, description, version, modularized, last_updated, allowed-tools, compliance_score) and 9 recommended fields.

**Required Fields** (7):
1. name
2. description (100-200 chars)
3. version (semantic versioning)
4. modularized (boolean)
5. last_updated (ISO 8601)
6. allowed-tools (array)
7. compliance_score (percentage)

**Recommended Fields** (9):
8. modules (if modularized=true)
9. dependencies (skill dependencies)
10. deprecated (boolean)
11. successor (if deprecated)
12. category_tier (1-6 for new system)
13. auto_trigger_keywords (8-15 keywords)
14. agent_coverage (agent references)
15. context7_references (external docs)
16. invocation_api_version (API compatibility)

**Related Tests**:
- `test_all_skills_100_percent_compliant()`
- `test_required_fields_present()`
- `test_recommended_fields_populated()`

#### REQ-015: Migration Timeline (Boundary Condition)
**SPEC**: The entire skill architecture redesign SHALL complete within 12 weeks organized in 3 phases of 4 weeks each, with validation gates at each phase boundary.

**Phase Breakdown**:
- **Phase 1 (Weeks 1-4)**: Design + metadata schema (20K tokens)
- **Phase 2 (Weeks 5-8)**: 40 skills implementation (60K tokens)
- **Phase 3 (Weeks 9-12)**: Migration + validation (40K tokens)

**Validation Gates**:
- Gate 1 (Week 4): Schema design approved, 100% metadata compliance
- Gate 2 (Week 8): All 40 skills implemented, tests passing ≥85%
- Gate 3 (Week 12): All agents migrated, backward compatibility verified

**Related Tests**:
- `test_phase1_deliverables_complete()`
- `test_phase2_deliverables_complete()`
- `test_phase3_deliverables_complete()`

---

## 🚫 Unwanted Behaviors (Constraints & Security)

### Security Constraints

**UWB-001**: The skill system SHALL NOT expose internal skill implementation details through the invocation API (information leakage prevention).

**UWB-002**: The semantic router SHALL NOT cache embeddings beyond 24 hours to prevent stale skill discovery (cache invalidation).

**UWB-003**: The skill system SHALL NOT allow arbitrary code execution through skill parameters (injection attack prevention).

**Related Tests**:
- `test_no_internal_details_exposed()`
- `test_cache_expires_after_24_hours()`
- `test_skill_parameters_sanitized()`

### Performance Constraints

**UWB-004**: The skill invocation API SHALL NOT block synchronously for >1 second during skill loading (async pattern enforcement).

**UWB-005**: The semantic router SHALL NOT consume >100MB memory for embedding storage (memory efficiency).

**UWB-006**: The skill system SHALL NOT exceed 150K token budget during Phase 2 implementation (token management).

**Related Tests**:
- `test_no_blocking_over_1_second()`
- `test_embedding_memory_under_100mb()`
- `test_phase2_tokens_under_150k()`

### Reliability Constraints

**UWB-007**: The skill migration process SHALL NOT lose any existing skill functionality from the 136 current skills (function preservation).

**UWB-008**: The backward compatibility alias system SHALL NOT redirect to incorrect skills (mapping accuracy).

**UWB-009**: The semantic router SHALL NOT return zero results for valid skill queries (discovery completeness).

**Related Tests**:
- `test_no_function_loss_during_migration()`
- `test_alias_mapping_100_percent_accurate()`
- `test_router_always_returns_results()`

### Data Integrity Constraints

**UWB-010**: The skill metadata SHALL NOT contain conflicting or inconsistent information (consistency validation).

**UWB-011**: The skill dependency graph SHALL NOT contain circular dependencies (acyclic enforcement).

**UWB-012**: The agent references SHALL NOT point to non-existent skills (referential integrity).

**Related Tests**:
- `test_metadata_consistency_validated()`
- `test_no_circular_dependencies()`
- `test_agent_references_valid_skills()`

---

## 📊 Technical Approach

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    User Request (Natural Language)           │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              Semantic Router (text-embedding-3-small)        │
│  • Embedding Generation (1536-dim vectors)                   │
│  • FAISS Vector Search (40 skill embeddings)                 │
│  • Top-3 Skills with Confidence Scores                       │
│  • Threshold: ≥0.75 for auto-selection                       │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  Skill Invocation API                        │
│  • Sequential Execution: invoke_skill(name, context)         │
│  • Parallel Execution: invoke_skills_parallel([...])         │
│  • Conditional Branching: if-then-else patterns              │
│  • Error Handling: Retry + Fallback                          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│         40 Consolidated Skills (6 Tiers)                     │
│  Tier 1: Foundation (3) │ Tier 4: Language (8)              │
│  Tier 2: Domain (12)    │ Tier 5: BaaS (6)                  │
│  Tier 3: Quality (6)    │ Tier 6: Specialized (5)           │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              Progressive Disclosure System                   │
│  Level 1: Quick Reference (<500 chars, 30-sec scan)         │
│  Level 2: Implementation Guide (<2000 chars, 5-min read)    │
│  Level 3: Advanced Patterns (<5000 chars, 10+ min dive)     │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│           Agent Execution (35 Specialized Agents)            │
│  • backend-expert → domain-backend, quality-security         │
│  • frontend-expert → domain-frontend, lang-typescript        │
│  • tdd-implementer → quality-testing, foundation-trust       │
└─────────────────────────────────────────────────────────────┘
```

### Semantic Router Implementation

**Technology Stack**:
- **Embedding Model**: OpenAI text-embedding-3-small (1536 dimensions)
- **Vector Database**: FAISS (Facebook AI Similarity Search)
- **Indexing**: Pre-compute embeddings for all 40 skills at system startup
- **Query Processing**: Real-time embedding generation for user requests
- **Similarity Metric**: Cosine similarity (range: 0-1)
- **Threshold**: ≥0.75 for high-confidence auto-selection

**Workflow**:
1. User request received (natural language)
2. Generate embedding for request (1536-dim vector)
3. FAISS search against 40 skill embeddings
4. Return top-3 skills with confidence scores
5. If max confidence ≥0.75 → auto-select skill
6. If max confidence <0.75 → prompt user for manual selection

**Latency Optimization**:
- Pre-compute all skill embeddings (one-time cost)
- Use FAISS IndexFlatIP for fast cosine similarity
- Cache embeddings in memory (< 100MB total)
- Target p95 latency: <500ms

### Skill Invocation API Design

**Core Interface**:
```python
from typing import Dict, List, Tuple, Optional

class SkillInvocationAPI:
    async def invoke_skill(
        self,
        skill_name: str,
        context: Dict[str, Any],
        level: int = 1  # Progressive disclosure level (1-3)
    ) -> Dict[str, Any]:
        """
        Invoke a single skill with context.

        Args:
            skill_name: Skill identifier (e.g., "moai-lang-python")
            context: Request context and parameters
            level: Disclosure level (1=quick, 2=impl, 3=advanced)

        Returns:
            Skill execution result with metadata
        """
        pass

    async def invoke_skills_parallel(
        self,
        skills: List[Tuple[str, Dict[str, Any]]]
    ) -> List[Dict[str, Any]]:
        """
        Invoke multiple skills concurrently.

        Args:
            skills: List of (skill_name, context) tuples

        Returns:
            List of execution results (same order as input)
        """
        pass

    async def invoke_skills_sequential(
        self,
        skills: List[Tuple[str, Dict[str, Any]]],
        chain_context: bool = True
    ) -> List[Dict[str, Any]]:
        """
        Invoke skills sequentially with optional context chaining.

        Args:
            skills: List of (skill_name, context) tuples
            chain_context: Pass output of skill N as context to skill N+1

        Returns:
            List of execution results
        """
        pass
```

**Error Handling**:
```python
class SkillNotFoundError(Exception):
    """Raised when skill name doesn't exist in 40-skill registry."""
    pass

class SkillInvocationError(Exception):
    """Raised when skill execution fails."""
    pass

class ConfidenceBelowThresholdError(Exception):
    """Raised when semantic router confidence <0.75."""
    pass
```

### Progressive Disclosure Implementation

**File Structure** (per skill):
```markdown
---
# YAML frontmatter metadata (required 7 + recommended 9 fields)
---

## Level 1: Quick Reference (30-sec, <500 chars)
One-paragraph summary of skill purpose, use cases, and value proposition.

## Level 2: Implementation Guide (5-min, <2000 chars)
- Core capabilities (bullet list)
- Usage patterns (code examples)
- Common scenarios
- Related skills

## Level 3: Advanced Patterns (10+ min, <5000 chars)
- Complex workflows
- Performance optimization
- Edge cases and gotchas
- Integration examples
- Context7 library references
```

**Token Efficiency**:
- **Level 1 only**: ~150 tokens (90% use cases)
- **Level 1 + 2**: ~600 tokens (complex tasks)
- **Level 1 + 2 + 3**: ~1500 tokens (advanced scenarios)
- **Current system**: ~3000 tokens (load all content)

**Savings**: 80% token reduction for typical requests (150 tokens vs. 3000 tokens)

---

## 🗺️ Implementation Roadmap

### Phase 1: Design + Metadata Schema (Weeks 1-4, 20K tokens)

**Week 1: Architecture Design**
- Finalize 40-skill consolidation mapping (136 → 40)
- Define 6-tier structure and naming conventions
- Document skill composition patterns
- Create migration strategy document

**Week 2: Metadata Schema**
- Define extended metadata fields (16 total)
- Create JSON Schema validation
- Implement metadata compliance checker
- Design backward compatibility alias system

**Week 3: Semantic Router Design**
- Select embedding model (text-embedding-3-small)
- Design FAISS index structure
- Define confidence threshold (≥0.75)
- Implement fallback mechanism

**Week 4: Skill Invocation API Design**
- Define API interface (sequential, parallel, conditional)
- Design error handling strategy
- Create progressive disclosure framework
- Validate with stakeholders

**Deliverables**:
- [ ] 40-skill mapping document (136 → 40 correspondence)
- [ ] 6-tier taxonomy definition
- [ ] Metadata JSON Schema (16 fields)
- [ ] Semantic router architecture diagram
- [ ] Skill invocation API specification
- [ ] Migration plan with validation gates

**Validation Gate 1** (End of Week 4):
- All design documents approved by stakeholders
- Metadata schema passes 100% validation
- API specification reviewed and approved
- Migration plan feasible within 12-week timeline

---

### Phase 2: Implementation (Weeks 5-8, 60K tokens)

**Week 5: Foundation + Domain Tier**
- Implement Tier 1 (Foundation): 3 skills
- Implement Tier 2 (Domain): 12 skills
- Total: 15 skills (37.5% of 40)

**Week 6: Quality + Language Tier**
- Implement Tier 3 (Quality): 6 skills
- Implement Tier 4 (Language): 8 skills
- Total: 14 skills (35% of 40)

**Week 7: BaaS + Specialized Tier**
- Implement Tier 5 (BaaS): 6 skills
- Implement Tier 6 (Specialized): 5 skills
- Total: 11 skills (27.5% of 40)

**Week 8: Integration + Testing**
- Implement semantic router (FAISS + embeddings)
- Implement skill invocation API
- Implement progressive disclosure system
- Implement backward compatibility aliases

**Deliverables**:
- [ ] 40 consolidated skills implemented
- [ ] All skills pass metadata compliance (100%)
- [ ] Semantic router functional (latency <500ms)
- [ ] Skill invocation API tested (sequential, parallel, conditional)
- [ ] Progressive disclosure validated (token savings ≥70%)
- [ ] Test coverage ≥85% per TRUST 5

**Validation Gate 2** (End of Week 8):
- All 40 skills implemented and documented
- Test coverage ≥85% for all skills
- Semantic router achieves <500ms p95 latency
- Token efficiency improved by ≥70%
- No regressions in existing functionality

---

### Phase 3: Migration + Validation (Weeks 9-12, 40K tokens)

**Week 9: Agent Migration**
- Update all 35 agents to reference new 40-skill system
- Test agent functionality with new skills
- Validate agent-skill integration

**Week 10: Backward Compatibility Testing**
- Implement 136 → 40 alias mappings
- Test legacy skill name redirects
- Validate deprecation warnings
- Set 6-month expiration for aliases

**Week 11: System Integration Testing**
- End-to-end workflow testing (/moai:1-plan, 2-run, 3-sync)
- Performance benchmarking (token usage, latency)
- Agentic execution rate measurement (target: ≥85%)
- Security and compliance validation

**Week 12: Documentation + Deployment**
- Generate auto-documentation for all 40 skills
- Create migration guide for users
- Update CLAUDE.md and memory files
- Deploy to production with monitoring

**Deliverables**:
- [ ] All 35 agents migrated and tested
- [ ] 136 → 40 alias mappings functional
- [ ] Agentic execution rate ≥85%
- [ ] Token efficiency +70% validated
- [ ] Zero orphaned skills (100% utilization)
- [ ] Migration guide published
- [ ] Production deployment complete

**Validation Gate 3** (End of Week 12):
- All agents functional with new skill system
- Backward compatibility verified for 6-month period
- Agentic execution rate ≥85% measured
- Token efficiency +70% confirmed
- Zero function loss from 136 original skills
- Production monitoring active

---

## 📏 Success Metrics

### Primary KPIs

1. **Token Efficiency**: ≥70% reduction (45-50K → 13-15K tokens per SPEC)
2. **Agentic Execution Rate**: ≥85% autonomous skill selection
3. **Discovery Speed**: p95 latency <500ms
4. **Zero Orphaned Skills**: 100% skill utilization (40/40 skills referenced)
5. **Test Coverage**: ≥85% per TRUST 5 framework
6. **Migration Success**: 100% function preservation (136 → 40 skills)

### Secondary Metrics

7. **Metadata Compliance**: 100% (all 40 skills)
8. **Agent Integration**: 35/35 agents migrated
9. **Documentation Quality**: Auto-generated docs for all 40 skills
10. **Backward Compatibility**: 136 aliases functional for 6 months
11. **Semantic Router Accuracy**: ≥90% correct skill matches
12. **Developer Satisfaction**: Positive feedback from skill users

---

## 🔗 Dependencies

### Technical Dependencies

- **Context7 MCP**: Latest library documentation integration
- **Sequential-Thinking MCP**: Complex reasoning support
- **FAISS Library**: Vector similarity search (≥1.7.4)
- **OpenAI API**: text-embedding-3-small model access
- **Python 3.13+**: Async/await support
- **pytest 8.3+**: Test framework with async support

### Process Dependencies

- **SPEC-SKILL-PORTFOLIO-OPT-001**: Current skill portfolio optimization work
- **Agent System**: 35 specialized agents requiring updates
- **Git Strategy**: GitHub Flow 3-mode system (manual, personal, team)
- **Token Budget**: 120-150K tokens allocated for 12-week implementation

---

## 🎯 Acceptance Criteria

### Primary Acceptance Criteria

✅ **AC-001**: All 136 existing skills successfully consolidated into exactly 40 capability packages with zero function loss.

✅ **AC-002**: Semantic router returns top-3 skills with confidence ≥0.75 in <500ms for 95th percentile requests.

✅ **AC-003**: Skill invocation API supports sequential, parallel, and conditional execution patterns with error handling.

✅ **AC-004**: Token efficiency improved by ≥70% (45-50K → 13-15K tokens per SPEC creation cycle).

✅ **AC-005**: Agentic execution rate ≥85% (agents autonomously select correct skills without user intervention).

✅ **AC-006**: Progressive disclosure system reduces token loading by 80% for typical requests (Level 1 only).

✅ **AC-007**: All 35 agents successfully migrated to reference new 40-skill architecture.

✅ **AC-008**: Backward compatibility alias system functional for 136 → 40 redirects with deprecation warnings.

✅ **AC-009**: Test coverage ≥85% for all 40 skills per TRUST 5 framework.

✅ **AC-010**: Metadata compliance 100% (all required and recommended fields populated for 40 skills).

### Validation Scenarios

**Scenario 1: Semantic Router Auto-Selection**
```
GIVEN a user request "optimize async Python code with FastAPI"
WHEN the semantic router processes the request
THEN it returns top-3 skills: moai-lang-python (0.89), moai-domain-backend (0.81), moai-quality-performance (0.76)
AND auto-selects moai-lang-python due to confidence ≥0.75
AND completes in <500ms
```

**Scenario 2: Skill Invocation API - Sequential Execution**
```
GIVEN a backend API design task
WHEN agent invokes:
  1. moai-domain-backend (design REST API)
  2. moai-quality-security (validate OWASP compliance)
  3. moai-quality-testing (generate test cases)
THEN all skills execute sequentially with context chaining
AND each skill's output becomes input to next skill
AND total execution completes successfully
```

**Scenario 3: Progressive Disclosure - Token Savings**
```
GIVEN a simple Python coding task
WHEN agent loads moai-lang-python skill
THEN only Level 1 content is loaded (<500 chars, ~150 tokens)
AND token savings ≥80% compared to loading full skill (3000 tokens)
AND agent completes task successfully with Level 1 only
```

**Scenario 4: Backward Compatibility - Legacy Skill Redirect**
```
GIVEN a legacy skill reference "moai-lang-python-async"
WHEN system processes the invocation
THEN it redirects to consolidated skill "moai-lang-python"
AND displays deprecation warning: "moai-lang-python-async → moai-lang-python (expires 2025-05-24)"
AND skill executes successfully
```

**Scenario 5: Agent Migration - Backend Expert**
```
GIVEN backend-expert agent
WHEN agent processes a REST API design request
THEN agent references new skills: moai-domain-backend, moai-quality-security, moai-lang-python
AND no references to old 136-skill system
AND agent completes task successfully
```

---

## 🛡️ Risk Mitigation

### Risk 1: Function Loss During Consolidation
**Probability**: Medium | **Impact**: Critical

**Mitigation**:
- Create comprehensive mapping document (136 → 40 correspondence)
- Validate every function from 136 skills is preserved in 40 skills
- Implement automated test suite to verify function coverage
- Manual review by 2+ stakeholders before migration

**Rollback Plan**: Maintain 136-skill system in parallel for 6 months

---

### Risk 2: Semantic Router Accuracy
**Probability**: Medium | **Impact**: High

**Mitigation**:
- Test router with 1000+ diverse requests
- Tune confidence threshold dynamically (start at 0.75, adjust based on data)
- Implement fallback to manual selection if confidence <0.75
- Monitor false positive/negative rates in production

**Rollback Plan**: Disable semantic router, use manual skill selection

---

### Risk 3: Agent Integration Breakage
**Probability**: Low | **Impact**: High

**Mitigation**:
- Test each of 35 agents individually after migration
- Implement backward compatibility aliases for 6 months
- Gradual rollout (5 agents per week over 7 weeks)
- Rollback capability per agent

**Rollback Plan**: Revert agent to old skill references, re-test

---

### Risk 4: Token Budget Overrun
**Probability**: Low | **Impact**: Medium

**Mitigation**:
- Strict phase-wise token allocation (Phase 1: 20K, Phase 2: 60K, Phase 3: 40K)
- Monitor token usage daily with automated reports
- Implement /clear between phases to reset context
- Buffer allocation (30K tokens) for contingencies

**Rollback Plan**: Extend timeline by 2-4 weeks if needed

---

### Risk 5: Backward Compatibility Issues
**Probability**: Medium | **Impact**: Medium

**Mitigation**:
- Comprehensive alias testing (all 136 → 40 mappings)
- 6-month transition period with deprecation warnings
- User documentation and migration guide
- Automated tests for all alias redirects

**Rollback Plan**: Extend alias expiration by 6 more months

---

## 📚 References

### Internal Documents
- `.moai/project/structure.md` - Current 10-tier skill organization
- `.moai/project/tech.md` - Metadata standards and validation
- `SPEC-SKILL-PORTFOLIO-OPT-001` - Skill portfolio optimization
- `.moai/memory/execution-rules.md` - TRUST 5 framework
- `.moai/memory/agents.md` - 35 agent specifications

### External Standards
- **Semantic Versioning**: https://semver.org/
- **EARS Methodology**: Easy Approach to Requirements Syntax
- **TRUST 5 Framework**: Test-first, Readable, Unified, Secured, Trackable
- **FAISS Documentation**: https://github.com/facebookresearch/faiss
- **OpenAI Embeddings**: https://platform.openai.com/docs/guides/embeddings

---

**SPEC Status**: DRAFT ✅
**Next Step**: User approval required to proceed with implementation
**Estimated Timeline**: 12 weeks (3 phases × 4 weeks)
**Token Budget**: 120-150K tokens (within 250K limit)
**Risk Level**: High (mitigated through phased approach + validation gates)

---

**Created**: 2025-11-24
**Author**: spec-builder agent
**Reviewers**: TBD
**Approval**: PENDING
