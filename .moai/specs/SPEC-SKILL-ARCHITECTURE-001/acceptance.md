---
id: SPEC-SKILL-ARCHITECTURE-001-ACCEPTANCE
title: Acceptance Criteria - MoAI Skill Architecture Redesign
spec_id: SPEC-SKILL-ARCHITECTURE-001
created_at: 2025-11-24
updated_at: 2025-11-24
status: draft
---

# Acceptance Criteria: SPEC-SKILL-ARCHITECTURE-001

## 📋 Overview

This document defines comprehensive acceptance criteria for the MoAI Skill Architecture Redesign project. All criteria must pass before production deployment.

---

## ✅ Primary Acceptance Criteria

### AC-001: Skill Consolidation (Function Preservation)

**Requirement**: All 136 existing skills successfully consolidated into exactly 40 capability packages with zero function loss.

**Validation Method**:
1. Create function inventory of all 136 current skills
2. Map each function to one of 40 consolidated skills
3. Verify 100% function coverage
4. Test each function in new skill architecture

**Acceptance Test**:
```python
def test_all_136_functions_preserved():
    """
    GIVEN a comprehensive function inventory of 136 current skills
    WHEN functions are mapped to 40 consolidated skills
    THEN 100% of functions must be preserved
    AND no function is lost or duplicated
    """
    original_functions = load_136_skill_functions()
    consolidated_functions = load_40_skill_functions()

    assert len(original_functions) == len(consolidated_functions)
    for func in original_functions:
        assert func in consolidated_functions

    assert get_function_coverage() == 100.0  # 100% coverage
```

**Definition of Done**:
- [ ] Function inventory complete for 136 skills
- [ ] All functions mapped to 40 consolidated skills
- [ ] Test passes with 100% function preservation
- [ ] No duplicate or orphaned functions

---

### AC-002: Semantic Router Performance

**Requirement**: Semantic router returns top-3 skills with confidence ≥0.75 in <500ms for 95th percentile requests.

**Validation Method**:
1. Execute 1000 diverse test queries
2. Measure latency for each query
3. Calculate p50, p95, p99 latencies
4. Verify confidence scores ≥0.75 for auto-selection

**Acceptance Test**:
```python
def test_semantic_router_performance():
    """
    GIVEN 1000 diverse skill-related queries
    WHEN semantic router processes each query
    THEN p95 latency must be <500ms
    AND top-3 skills returned with confidence scores
    AND confidence ≥0.75 for auto-selection
    """
    test_queries = load_1000_test_queries()
    latencies = []

    for query in test_queries:
        start = time.time()
        results = semantic_router.search(query, top_k=3)
        latency = (time.time() - start) * 1000  # Convert to ms
        latencies.append(latency)

        assert len(results) == 3
        if results[0].confidence >= 0.75:
            assert results[0].skill_name in VALID_40_SKILLS

    assert percentile(latencies, 95) < 500  # p95 <500ms
    assert percentile(latencies, 50) < 200  # p50 <200ms
```

**Definition of Done**:
- [ ] 1000 test queries executed successfully
- [ ] p95 latency <500ms confirmed
- [ ] p50 latency <200ms (bonus)
- [ ] Confidence threshold (0.75) validated
- [ ] Top-3 results returned for all queries

---

### AC-003: Skill Invocation API Patterns

**Requirement**: Skill invocation API supports sequential, parallel, and conditional execution patterns with error handling.

**Validation Method**:
1. Test sequential skill chaining
2. Test parallel skill execution
3. Test conditional skill branching
4. Test error recovery patterns

**Acceptance Test**:
```python
@pytest.mark.asyncio
async def test_skill_invocation_api_patterns():
    """
    GIVEN the skill invocation API
    WHEN executing sequential, parallel, and conditional patterns
    THEN all patterns complete successfully
    AND error handling works correctly
    """
    api = SkillInvocationAPI()

    # Sequential execution
    result = await api.invoke_skills_sequential([
        ("moai-domain-backend", {"design": "REST API"}),
        ("moai-quality-security", {"validate": "OWASP"}),
        ("moai-quality-testing", {"generate": "test cases"})
    ], chain_context=True)
    assert len(result) == 3
    assert all(r.success for r in result)

    # Parallel execution
    results = await api.invoke_skills_parallel([
        ("moai-domain-backend", {"task": "API"}),
        ("moai-domain-frontend", {"task": "UI"}),
        ("moai-domain-database", {"task": "schema"})
    ])
    assert len(results) == 3
    assert all(r.success for r in results)

    # Conditional execution
    context = {"language": "python"}
    if context["language"] == "python":
        result = await api.invoke_skill("moai-lang-python", context)
    else:
        result = await api.invoke_skill("moai-lang-typescript", context)
    assert result.success
    assert result.skill_name == "moai-lang-python"

    # Error handling
    with pytest.raises(SkillNotFoundError):
        await api.invoke_skill("non-existent-skill", {})
```

**Definition of Done**:
- [ ] Sequential execution tested and passing
- [ ] Parallel execution tested and passing
- [ ] Conditional branching tested and passing
- [ ] Error handling tested and passing
- [ ] Context chaining works correctly

---

### AC-004: Token Efficiency Improvement

**Requirement**: Token efficiency improved by ≥70% (45-50K → 13-15K tokens per SPEC creation cycle).

**Validation Method**:
1. Measure token usage for 10 sample SPEC creation cycles (old system)
2. Measure token usage for same 10 SPECs (new system)
3. Calculate percentage reduction
4. Verify ≥70% improvement

**Acceptance Test**:
```python
def test_token_efficiency_improvement():
    """
    GIVEN 10 sample SPEC creation cycles
    WHEN comparing old (136 skills) vs new (40 skills) system
    THEN token usage reduction must be ≥70%
    """
    old_system_tokens = []
    new_system_tokens = []

    for spec in SAMPLE_SPECS:
        # Old system: Load all relevant skills from 136
        old_tokens = simulate_old_system_spec_creation(spec)
        old_system_tokens.append(old_tokens)

        # New system: Progressive disclosure + semantic router
        new_tokens = simulate_new_system_spec_creation(spec)
        new_system_tokens.append(new_tokens)

    avg_old = sum(old_system_tokens) / len(old_system_tokens)
    avg_new = sum(new_system_tokens) / len(new_system_tokens)
    reduction = ((avg_old - avg_new) / avg_old) * 100

    assert avg_old >= 45000  # Old system: 45-50K tokens
    assert avg_old <= 50000
    assert avg_new >= 13000  # New system: 13-15K tokens
    assert avg_new <= 15000
    assert reduction >= 70.0  # ≥70% reduction
```

**Definition of Done**:
- [ ] 10 SPEC creation cycles measured
- [ ] Old system average: 45-50K tokens confirmed
- [ ] New system average: 13-15K tokens confirmed
- [ ] Token reduction ≥70% validated

---

### AC-005: Agentic Execution Rate

**Requirement**: Agentic execution rate ≥85% (agents autonomously select correct skills without user intervention).

**Validation Method**:
1. Execute 1000 test requests through agent system
2. Measure autonomous skill selection accuracy
3. Track false positives and false negatives
4. Calculate success rate

**Acceptance Test**:
```python
def test_agentic_execution_rate():
    """
    GIVEN 1000 diverse test requests
    WHEN agents process requests using semantic router
    THEN autonomous skill selection rate must be ≥85%
    AND false positive rate <5%
    AND false negative rate <10%
    """
    test_requests = load_1000_test_requests()
    results = {
        "autonomous_success": 0,
        "manual_fallback": 0,
        "false_positive": 0,
        "false_negative": 0
    }

    for request in test_requests:
        outcome = agent_process_request(request)

        if outcome.autonomous and outcome.correct:
            results["autonomous_success"] += 1
        elif not outcome.autonomous:
            results["manual_fallback"] += 1
        elif outcome.autonomous and not outcome.correct:
            results["false_positive"] += 1
        else:
            results["false_negative"] += 1

    autonomous_rate = (results["autonomous_success"] / 1000) * 100
    false_positive_rate = (results["false_positive"] / 1000) * 100
    false_negative_rate = (results["false_negative"] / 1000) * 100

    assert autonomous_rate >= 85.0  # ≥85% autonomous
    assert false_positive_rate < 5.0  # <5% false positives
    assert false_negative_rate < 10.0  # <10% false negatives
```

**Definition of Done**:
- [ ] 1000 test requests executed
- [ ] Autonomous selection rate ≥85%
- [ ] False positive rate <5%
- [ ] False negative rate <10%

---

### AC-006: Progressive Disclosure Token Savings

**Requirement**: Progressive disclosure system reduces token loading by 80% for typical requests (Level 1 only).

**Validation Method**:
1. Measure tokens for full skill content (all 3 levels)
2. Measure tokens for Level 1 only
3. Calculate percentage reduction
4. Verify ≥80% savings

**Acceptance Test**:
```python
def test_progressive_disclosure_token_savings():
    """
    GIVEN all 40 skills with 3-level progressive disclosure
    WHEN loading Level 1 only (quick reference)
    THEN token usage reduction must be ≥80%
    COMPARED to loading full skill content
    """
    token_savings = []

    for skill_name in ALL_40_SKILLS:
        full_content = load_skill_full_content(skill_name)
        level1_content = load_skill_level1_only(skill_name)

        full_tokens = count_tokens(full_content)
        level1_tokens = count_tokens(level1_content)

        reduction = ((full_tokens - level1_tokens) / full_tokens) * 100
        token_savings.append(reduction)

        assert level1_tokens < 500  # Level 1 <500 chars ~150 tokens
        assert full_tokens >= 2000  # Full content ~600-1500 tokens

    avg_savings = sum(token_savings) / len(token_savings)
    assert avg_savings >= 80.0  # ≥80% average savings
```

**Definition of Done**:
- [ ] All 40 skills tested
- [ ] Level 1 content <500 characters confirmed
- [ ] Token savings ≥80% validated
- [ ] Progressive loading works correctly

---

### AC-007: Agent Migration Completion

**Requirement**: All 35 agents successfully migrated to reference new 40-skill architecture.

**Validation Method**:
1. Review all 35 agent files
2. Verify no references to old 136-skill system
3. Test agent functionality with new skills
4. Validate agent-skill integration

**Acceptance Test**:
```python
def test_all_agents_migrated():
    """
    GIVEN all 35 specialized agents
    WHEN reviewing agent skill references
    THEN all references must point to new 40-skill system
    AND no references to old 136-skill system
    AND all agents functional with new skills
    """
    agents = load_all_35_agents()

    for agent in agents:
        skill_refs = extract_skill_references(agent)

        # No old skill references
        old_skills = [ref for ref in skill_refs if ref not in NEW_40_SKILLS]
        assert len(old_skills) == 0, f"Agent {agent.name} has old references: {old_skills}"

        # All references valid
        for ref in skill_refs:
            assert ref in NEW_40_SKILLS

        # Agent functionality test
        test_result = test_agent_with_new_skills(agent)
        assert test_result.success
```

**Definition of Done**:
- [ ] All 35 agents reviewed
- [ ] No references to old 136-skill system
- [ ] All skill references valid (40-skill system)
- [ ] Agent functionality tests passing

---

### AC-008: Backward Compatibility Aliases

**Requirement**: Backward compatibility alias system functional for 136 → 40 redirects with deprecation warnings.

**Validation Method**:
1. Test all 136 legacy skill names
2. Verify correct redirection to 40 new skills
3. Confirm deprecation warnings displayed
4. Validate expiration dates (6 months)

**Acceptance Test**:
```python
def test_backward_compatibility_aliases():
    """
    GIVEN 136 legacy skill names from old system
    WHEN invoking legacy skill names
    THEN system redirects to correct consolidated skill (40 skills)
    AND deprecation warning is displayed
    AND expiration date is 2025-05-24 (6 months from 2024-11-24)
    """
    legacy_aliases = load_136_legacy_aliases()

    for legacy_name, expected_new_name in legacy_aliases.items():
        # Test redirection
        with warnings.catch_warnings(record=True) as w:
            result = invoke_skill(legacy_name, {})

            # Verify redirect
            assert result.actual_skill_name == expected_new_name

            # Verify deprecation warning
            assert len(w) == 1
            assert "deprecated" in str(w[0].message).lower()
            assert legacy_name in str(w[0].message)
            assert expected_new_name in str(w[0].message)

            # Verify expiration date
            expiration = get_alias_expiration(legacy_name)
            assert expiration == datetime(2025, 5, 24)
```

**Definition of Done**:
- [ ] All 136 aliases tested
- [ ] 100% redirection accuracy
- [ ] Deprecation warnings displayed correctly
- [ ] Expiration dates set to 2025-05-24

---

### AC-009: Test Coverage Compliance

**Requirement**: Test coverage ≥85% for all 40 skills per TRUST 5 framework.

**Validation Method**:
1. Run pytest with coverage plugin
2. Generate coverage report for all 40 skills
3. Verify each skill ≥85% coverage
4. Identify gaps and add tests

**Acceptance Test**:
```bash
pytest --cov=.claude/skills \
       --cov-report=term-missing \
       --cov-report=html \
       --cov-fail-under=85

# Verify each skill individually
for skill in $(ls .claude/skills/moai-*/); do
    coverage=$(pytest --cov=$skill --cov-report=term | grep "TOTAL" | awk '{print $4}')
    coverage_num=${coverage%\%}
    if [ "$coverage_num" -lt 85 ]; then
        echo "FAIL: $skill has only $coverage coverage (required: ≥85%)"
        exit 1
    fi
done

echo "PASS: All 40 skills have ≥85% test coverage"
```

**Definition of Done**:
- [ ] All 40 skills tested
- [ ] Each skill ≥85% coverage
- [ ] Coverage report generated
- [ ] No coverage gaps

---

### AC-010: Metadata Compliance

**Requirement**: Metadata compliance 100% (all required and recommended fields populated for 40 skills).

**Validation Method**:
1. Run metadata validation script
2. Check all 7 required fields present
3. Check all 9 recommended fields populated
4. Verify JSON Schema compliance

**Acceptance Test**:
```python
def test_metadata_100_percent_compliant():
    """
    GIVEN all 40 consolidated skills
    WHEN validating metadata against JSON Schema
    THEN all skills must have 100% metadata compliance
    AND all 7 required fields present
    AND all 9 recommended fields populated
    """
    skills = load_all_40_skills()
    schema = load_metadata_json_schema()

    for skill in skills:
        metadata = extract_yaml_frontmatter(skill)

        # Required fields (7)
        required_fields = [
            "name", "description", "version", "modularized",
            "last_updated", "allowed-tools", "compliance_score"
        ]
        for field in required_fields:
            assert field in metadata, f"{skill.name} missing required field: {field}"

        # Recommended fields (9)
        recommended_fields = [
            "modules", "dependencies", "deprecated", "successor",
            "category_tier", "auto_trigger_keywords", "agent_coverage",
            "context7_references", "invocation_api_version"
        ]
        for field in recommended_fields:
            assert field in metadata, f"{skill.name} missing recommended field: {field}"

        # JSON Schema validation
        validate(metadata, schema)

        # Compliance score must be 100%
        assert metadata["compliance_score"] == "100%"
```

**Definition of Done**:
- [ ] All 40 skills validated
- [ ] 7 required fields present for each skill
- [ ] 9 recommended fields populated for each skill
- [ ] JSON Schema validation passing
- [ ] Compliance scores all 100%

---

## 🎯 Scenario-Based Acceptance Tests

### Scenario 1: Semantic Router Auto-Selection

**Test Case**: Semantic router automatically selects correct skill for Python optimization request

**Given-When-Then**:
```
GIVEN a user request "optimize async Python code with FastAPI"
WHEN the semantic router processes the request
THEN it returns top-3 skills:
  1. moai-lang-python (confidence: 0.89)
  2. moai-domain-backend (confidence: 0.81)
  3. moai-quality-performance (confidence: 0.76)
AND auto-selects moai-lang-python (confidence ≥0.75)
AND completes in <500ms
```

**Implementation**:
```python
@pytest.mark.asyncio
async def test_scenario_semantic_router_auto_selection():
    query = "optimize async Python code with FastAPI"

    start = time.time()
    results = await semantic_router.search(query, top_k=3)
    latency = (time.time() - start) * 1000

    assert len(results) == 3
    assert results[0].skill_name == "moai-lang-python"
    assert results[0].confidence >= 0.75
    assert results[1].skill_name == "moai-domain-backend"
    assert results[2].skill_name == "moai-quality-performance"
    assert latency < 500  # <500ms
```

---

### Scenario 2: Skill Invocation API - Sequential Execution

**Test Case**: Backend API design workflow with sequential skill chaining

**Given-When-Then**:
```
GIVEN a backend API design task
WHEN agent invokes skills sequentially:
  1. moai-domain-backend (design REST API)
  2. moai-quality-security (validate OWASP compliance)
  3. moai-quality-testing (generate test cases)
THEN all skills execute sequentially with context chaining
AND each skill's output becomes input to next skill
AND total execution completes successfully
```

**Implementation**:
```python
@pytest.mark.asyncio
async def test_scenario_sequential_skill_execution():
    api = SkillInvocationAPI()

    results = await api.invoke_skills_sequential([
        ("moai-domain-backend", {"design": "REST API for user authentication"}),
        ("moai-quality-security", {"validate": "OWASP Top 10"}),
        ("moai-quality-testing", {"generate": "test cases"})
    ], chain_context=True)

    assert len(results) == 3
    assert all(r.success for r in results)

    # Verify context chaining
    assert "REST API" in results[1].context_received
    assert "OWASP" in results[2].context_received
```

---

### Scenario 3: Progressive Disclosure - Token Savings

**Test Case**: Simple Python coding task loads only Level 1 content

**Given-When-Then**:
```
GIVEN a simple Python coding task "write hello world function"
WHEN agent loads moai-lang-python skill
THEN only Level 1 content is loaded (<500 chars, ~150 tokens)
AND token savings ≥80% compared to loading full skill (3000 tokens)
AND agent completes task successfully with Level 1 only
```

**Implementation**:
```python
def test_scenario_progressive_disclosure_token_savings():
    task = "write hello world function in Python"

    # Load only Level 1
    level1_content = load_skill_level1("moai-lang-python")
    level1_tokens = count_tokens(level1_content)

    # Compare to full content
    full_content = load_skill_full_content("moai-lang-python")
    full_tokens = count_tokens(full_content)

    savings = ((full_tokens - level1_tokens) / full_tokens) * 100

    assert level1_tokens < 500  # ~150 tokens
    assert full_tokens >= 2000  # ~600-1500 tokens
    assert savings >= 80.0  # ≥80% savings

    # Agent completes task with Level 1 only
    result = agent_complete_task(task, level1_content)
    assert result.success
```

---

### Scenario 4: Backward Compatibility - Legacy Skill Redirect

**Test Case**: Legacy skill name redirects to consolidated skill with deprecation warning

**Given-When-Then**:
```
GIVEN a legacy skill reference "moai-lang-python-async"
WHEN system processes the invocation
THEN it redirects to consolidated skill "moai-lang-python"
AND displays deprecation warning:
  "moai-lang-python-async → moai-lang-python (expires 2025-05-24)"
AND skill executes successfully
```

**Implementation**:
```python
def test_scenario_backward_compatibility_redirect():
    legacy_skill = "moai-lang-python-async"
    expected_new_skill = "moai-lang-python"

    with warnings.catch_warnings(record=True) as w:
        result = invoke_skill(legacy_skill, {"task": "async optimization"})

        # Verify redirect
        assert result.actual_skill_name == expected_new_skill
        assert result.success

        # Verify deprecation warning
        assert len(w) == 1
        warning_msg = str(w[0].message)
        assert legacy_skill in warning_msg
        assert expected_new_skill in warning_msg
        assert "2025-05-24" in warning_msg
```

---

### Scenario 5: Agent Migration - Backend Expert

**Test Case**: Backend expert agent uses new skill references for REST API design

**Given-When-Then**:
```
GIVEN backend-expert agent
WHEN agent processes a REST API design request
THEN agent references new skills:
  - moai-domain-backend
  - moai-quality-security
  - moai-lang-python
AND no references to old 136-skill system
AND agent completes task successfully
```

**Implementation**:
```python
def test_scenario_agent_migration_backend_expert():
    agent = load_agent("backend-expert")
    request = {"task": "design REST API for user management"}

    # Execute agent workflow
    result = agent.process_request(request)

    # Verify skill references
    skill_refs = result.skills_invoked
    assert "moai-domain-backend" in skill_refs
    assert "moai-quality-security" in skill_refs
    assert "moai-lang-python" in skill_refs

    # No old skill references
    old_skills = [s for s in skill_refs if s not in NEW_40_SKILLS]
    assert len(old_skills) == 0

    # Task completed successfully
    assert result.success
```

---

## 📊 Performance Benchmarks

### Benchmark 1: Semantic Router Latency

**Metric**: p50, p95, p99 latency for 10,000 queries

**Target**:
- p50: <200ms
- p95: <500ms
- p99: <1000ms

**Test**:
```bash
# Run latency benchmark
python benchmarks/semantic_router_latency.py --queries=10000

# Expected output:
# p50 latency: 178ms ✅
# p95 latency: 467ms ✅
# p99 latency: 892ms ✅
```

---

### Benchmark 2: Token Efficiency

**Metric**: Average token usage per SPEC creation cycle

**Target**:
- Old system: 45-50K tokens
- New system: 13-15K tokens
- Reduction: ≥70%

**Test**:
```bash
# Run token efficiency benchmark
python benchmarks/token_efficiency.py --specs=100

# Expected output:
# Old system avg: 47,234 tokens
# New system avg: 14,123 tokens
# Reduction: 70.1% ✅
```

---

### Benchmark 3: Agentic Execution Rate

**Metric**: Autonomous skill selection accuracy over 1000 requests

**Target**:
- Autonomous success: ≥85%
- False positives: <5%
- False negatives: <10%

**Test**:
```bash
# Run agentic execution benchmark
python benchmarks/agentic_execution_rate.py --requests=1000

# Expected output:
# Autonomous success: 87.3% ✅
# False positives: 3.2% ✅
# False negatives: 9.5% ✅
```

---

## 🔍 Quality Gates

### Gate 1: Phase 1 Completion (End of Week 4)

**Criteria**:
- [ ] All design documents approved
- [ ] Metadata schema passes 100% validation
- [ ] API specification reviewed and approved
- [ ] Migration plan feasible within 12-week timeline

**Approval**: Project Manager, Technical Lead, QA Lead

---

### Gate 2: Phase 2 Completion (End of Week 8)

**Criteria**:
- [ ] All 40 skills implemented and documented
- [ ] Test coverage ≥85% for all skills
- [ ] Semantic router p95 latency <500ms
- [ ] Token efficiency ≥70% improvement
- [ ] No regressions in existing functionality

**Approval**: Technical Lead, QA Lead, Performance Engineer

---

### Gate 3: Phase 3 Completion (End of Week 12)

**Criteria**:
- [ ] All 35 agents migrated and tested
- [ ] Backward compatibility verified (6 months)
- [ ] Agentic execution rate ≥85%
- [ ] Token efficiency +70% confirmed
- [ ] Zero function loss from 136 skills
- [ ] Production monitoring active

**Approval**: Project Manager, Technical Lead, QA Lead, Product Owner, Security Expert

---

## 📚 Testing Strategy

### Unit Tests
- Individual skill functionality
- Metadata validation
- Semantic router components
- Skill invocation API methods

### Integration Tests
- Semantic router + skill invocation API
- Agent-skill integration
- Backward compatibility aliases
- Progressive disclosure system

### Performance Tests
- Semantic router latency benchmarks
- Token usage measurements
- Agentic execution rate tracking
- Scalability tests (1000+ queries)

### Regression Tests
- All 136 original skill functions
- Agent functionality with new skills
- Existing workflows (/moai:1-plan, 2-run, 3-sync)

### Security Tests
- OWASP Top 10 compliance
- Input validation and sanitization
- Injection attack prevention
- Information leakage prevention

---

## 🎯 Definition of Done

**Project Complete When**:
- ✅ All 10 primary acceptance criteria pass
- ✅ All 5 scenario tests pass
- ✅ All 3 performance benchmarks meet targets
- ✅ All 3 quality gates approved
- ✅ Production deployment successful
- ✅ Monitoring active and collecting metrics
- ✅ User migration guide published
- ✅ Stakeholder sign-off received

---

**Acceptance Criteria Status**: DRAFT ✅
**Next Step**: User approval to proceed with implementation
**Last Updated**: 2025-11-24
**Author**: spec-builder agent
