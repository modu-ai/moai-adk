# moai-foundation-core Reference

Progressive Disclosure Level 2: Extended documentation for foundational principles and architectural patterns.

---

## API Reference

### Core Concepts

TRUST 5 Framework:
- Purpose: Automated quality gate system for code quality assurance
- Components: Test-first, Readable, Unified, Secured, Trackable
- Integration: Pre-commit hooks, CI/CD pipelines, agent workflows
- Validation: Automated checks with configurable thresholds

SPEC-First DDD:
- Purpose: Specification-driven development workflow
- Phases: SPEC (Plan), DDD (Run), Docs (Sync)
- Format: EARS (Event-Action-Response-State) specifications
- Token Budget: 30K (SPEC) + 180K (DDD) + 40K (Docs) = 250K total

Delegation Patterns:
- Purpose: Task orchestration via specialized agents
- Core Principle: MoAI delegates all work through Task() calls
- Patterns: Sequential, Parallel, Conditional delegation
- Agent Selection: Complexity-based agent matching

Token Optimization:
- Purpose: Efficient 200K token budget management
- Strategies: Phase separation, selective loading, context reset
- Savings: /clear after phases, model selection optimization
- Monitoring: /context command for budget tracking

Progressive Disclosure:
- Purpose: Three-tier knowledge delivery architecture
- Levels: Quick (30s), Implementation (5min), Advanced (10+min)
- Token Allocation: 1K, 3K, 5K tokens respectively
- Structure: SKILL.md core with modules/ overflow

Modular System:
- Purpose: Scalable file organization for unlimited content
- Structure: SKILL.md (500 lines) + modules/ (unlimited)
- Principles: Topic-focused, self-contained modules
- Discovery: Progressive navigation from core to deep dive

---

## Configuration Options

### TRUST 5 Thresholds

```yaml
trust_5:
  test_first:
    min_coverage: 0.85           # Minimum 85% test coverage
    coverage_tool: "pytest-cov"  # Coverage measurement tool
    failure_action: "block"      # block, warn, ignore

  readable:
    linter: "ruff"               # Code linter tool
    max_complexity: 10           # Maximum cyclomatic complexity
    naming_convention: "snake_case"
    failure_action: "warn"

  unified:
    formatter: "black"           # Code formatter
    import_sorter: "isort"       # Import organization
    line_length: 88              # Maximum line length
    failure_action: "auto_fix"

  secured:
    scanner: "bandit"            # Security scanner
    owasp_compliance: true       # OWASP Top 10 check
    secret_detection: true       # Credential scanning
    failure_action: "block"

  trackable:
    commit_format: "conventional"  # Commit message format
    branch_naming: "feature/*"     # Branch naming pattern
    changelog_required: true       # Changelog requirement
    failure_action: "warn"
```

### Token Budget Allocation

```yaml
token_budget:
  total: 250000                  # Total available tokens

  phases:
    spec:
      budget: 30000              # 30K for SPEC phase
      strategy: "minimal_context"
      clear_after: true          # Execute /clear after phase

    ddd:
      budget: 180000             # 180K for DDD phase
      strategy: "selective_loading"
      file_priority: ["tests", "src", "config"]

    docs:
      budget: 40000              # 40K for docs phase
      strategy: "cached_results"
      template_reuse: true

  optimization:
    clear_threshold: 150000      # Suggest /clear at this level
    message_limit: 50            # Suggest /clear after messages
    model_selection:
      quality: "sonnet"          # For quality-critical tasks
      speed: "haiku"             # For speed/cost optimization
```

### Agent Selection Matrix

```yaml
agent_selection:
  by_complexity:
    simple:                      # 1 file
      agents: 1-2
      pattern: "sequential"

    medium:                      # 3-5 files
      agents: 2-3
      pattern: "sequential"

    complex:                     # 10+ files
      agents: 5+
      pattern: "mixed"

  by_domain:
    backend:
      primary: "backend-expert"
      supporting: ["database-expert", "api-designer"]

    frontend:
      primary: "frontend-expert"
      supporting: ["ui-specialist", "accessibility-checker"]

    security:
      primary: "security-expert"
      supporting: ["compliance-checker", "penetration-tester"]

    infrastructure:
      primary: "devops-expert"
      supporting: ["cloud-architect", "monitoring-specialist"]
```

---

## Integration Patterns

### Pattern 1: TRUST 5 Pre-Commit Hook

```python
# Pre-commit hook integration for TRUST 5 validation
from moai_foundation_core import TRUST5Validator

def pre_commit_validation(staged_files: List[str]) -> bool:
    """Run TRUST 5 validation on staged files."""

    validator = TRUST5Validator()

    # Run all pillars
    results = {
        "test_first": validator.check_test_coverage(staged_files),
        "readable": validator.check_readability(staged_files),
        "unified": validator.check_formatting(staged_files),
        "secured": validator.check_security(staged_files),
        "trackable": validator.check_commit_message()
    }

    # Evaluate results
    all_passed = all(r.passed for r in results.values())

    if not all_passed:
        for pillar, result in results.items():
            if not result.passed:
                print(f"[TRUST 5] {pillar.upper()}: {result.message}")

    return all_passed
```

### Pattern 2: SPEC-First DDD Workflow

```python
# Complete SPEC-First DDD cycle implementation
from moai_foundation_core import SPECManager, DDDExecutor, DocsGenerator

async def spec_first_workflow(requirements: str):
    """Execute complete SPEC-First DDD workflow."""

    # Phase 1: SPEC Generation (30K tokens)
    spec_manager = SPECManager()
    spec = await spec_manager.generate_spec(
        requirements=requirements,
        format="EARS"
    )
    print(f"Generated: {spec.id}")

    # Critical: Clear context before Phase 2
    await execute_clear()  # Saves 45-50K tokens

    # Phase 2: DDD Implementation (180K tokens)
    ddd = DDDExecutor(spec.id)

    # ANALYZE: Understand requirements and existing behavior
    analysis = await ddd.analyze()
    assert analysis.requirements_complete == True

    # PRESERVE: Ensure existing behavior is protected
    characterization = await ddd.preserve()
    assert await ddd.run_tests() == "PASS"

    # IMPROVE: Implement improvements incrementally
    improved = await ddd.improve()
    assert await ddd.run_tests() == "PASS"

    # Validate coverage
    coverage = await ddd.get_coverage()
    assert coverage >= 0.85

    # Phase 3: Documentation (40K tokens)
    docs = DocsGenerator(spec.id)
    await docs.generate_api_docs()
    await docs.generate_architecture_diagrams()
    await docs.update_project_readme()

    return {"spec": spec, "coverage": coverage}
```

### Pattern 3: Intelligent Agent Delegation

```python
# Complexity-based agent delegation
from moai_foundation_core import TaskAnalyzer, AgentRouter

async def delegate_task(task_description: str, context: dict):
    """Delegate task to appropriate agents based on complexity."""

    # Analyze task complexity
    analyzer = TaskAnalyzer()
    analysis = analyzer.analyze(task_description, context)

    print(f"Complexity: {analysis.complexity}")
    print(f"Affected files: {analysis.file_count}")
    print(f"Domains: {analysis.domains}")

    # Route to appropriate agents
    router = AgentRouter()

    if analysis.complexity == "simple":
        # Sequential single agent
        result = await Task(
            subagent_type=router.get_primary_agent(analysis.domains[0]),
            prompt=task_description,
            context=context
        )

    elif analysis.complexity == "medium":
        # Sequential multiple agents
        results = []
        for domain in analysis.domains:
            result = await Task(
                subagent_type=router.get_primary_agent(domain),
                prompt=f"Handle {domain} aspects: {task_description}",
                context={**context, "previous_results": results}
            )
            results.append(result)

    else:  # complex
        # Parallel then sequential integration
        parallel_results = await Promise.all([
            Task(subagent_type=router.get_primary_agent(d), prompt=f"{d}: {task_description}")
            for d in analysis.domains
        ])

        # Integration phase
        result = await Task(
            subagent_type="integration-specialist",
            prompt="Integrate all components",
            context={"results": parallel_results}
        )

    return result
```

### Pattern 4: Token-Optimized Context Management

```python
# Proactive token budget management
from moai_foundation_core import TokenMonitor, ContextOptimizer

class TokenAwareWorkflow:
    """Workflow with automatic token optimization."""

    def __init__(self, total_budget: int = 200000):
        self.monitor = TokenMonitor(total_budget)
        self.optimizer = ContextOptimizer()

    async def execute_with_optimization(self, tasks: List[dict]):
        """Execute tasks with automatic context management."""

        results = []

        for task in tasks:
            # Check current usage
            usage = self.monitor.get_current_usage()

            if usage > 0.75:  # 75% threshold
                print(f"Token usage: {usage*100:.1f}%")

                # Apply optimization strategies
                if usage > 0.9:
                    print("Executing /clear to reset context")
                    await execute_clear()
                    self.monitor.reset()
                else:
                    # Selective context pruning
                    self.optimizer.prune_least_relevant()

            # Execute task
            result = await self.execute_task(task)
            results.append(result)

            # Update token tracking
            self.monitor.add_usage(result.tokens_used)

        return results
```

---

## Troubleshooting

### Common Issues

Issue: TRUST 5 validation fails inconsistently:
- Cause: Different tool versions between local and CI
- Solution: Pin tool versions in requirements.txt/pyproject.toml
- Prevention: Use locked dependency files

Issue: SPEC phase consumes too many tokens:
- Cause: Excessive context loading during spec generation
- Solution: Use minimal context mode for SPEC phase
- Prevention: Configure `spec.strategy: "minimal_context"`

Issue: Agent delegation creates infinite loops:
- Cause: Circular dependencies between agents
- Solution: Implement task ID tracking and loop detection
- Prevention: Design clear agent responsibilities with no overlap

Issue: Token budget exhausted mid-workflow:
- Cause: Missing /clear between phases
- Solution: Always execute /clear after Phase 1
- Prevention: Enable automatic phase boundary detection

Issue: Progressive disclosure not loading modules:
- Cause: Incorrect cross-reference paths
- Solution: Verify paths in SKILL.md match actual file locations
- Prevention: Use relative paths from SKILL.md location

### Diagnostic Commands

```bash
# TRUST 5 status check
moai-trust check --all --verbose

# Token budget analysis
moai-context analyze --show-breakdown

# Agent routing debug
moai-agent route --task "description" --dry-run

# SPEC validation
moai-spec validate SPEC-001 --format EARS
```

### Validation Utilities

```python
from moai_foundation_core import diagnose

# Full system diagnostics
report = diagnose.run_full_check()

# Individual component checks
diagnose.check_trust5_tools()
diagnose.check_agent_availability()
diagnose.check_token_tracking()
diagnose.verify_module_structure()
```

### Log Locations

- TRUST 5 logs: `.moai/logs/trust5/`
- SPEC artifacts: `.moai/specs/`
- Agent execution logs: `.moai/logs/agents/`
- Token usage history: `.moai/logs/tokens/`

---

## External Resources

### Official Documentation

- MoAI-ADK Documentation: See project README
- Claude Code Skills Guide: https://docs.anthropic.com/claude-code/skills
- EARS Specification Format: See `modules/spec-first-ddd.md`

### Module References

- TRUST 5 Framework: `modules/trust-5-framework.md`
- SPEC-First DDD: `modules/spec-first-ddd.md`
- Delegation Patterns: `modules/delegation-patterns.md`
- Token Optimization: `modules/token-optimization.md`
- Progressive Disclosure: `modules/progressive-disclosure.md`
- Modular System: `modules/modular-system.md`
- Agents Reference: `modules/agents-reference.md`
- Commands Reference: `modules/commands-reference.md`
- Execution Rules: `modules/execution-rules.md`

### Related Skills

- moai-foundation-claude - Claude Code integration patterns
- moai-workflow-project - Project management with core principles
- moai-workflow-testing - Testing workflows with TRUST 5
- moai-workflow-templates - Template management integration

### Tool References

- pytest: https://docs.pytest.org/
- ruff: https://docs.astral.sh/ruff/
- black: https://black.readthedocs.io/
- bandit: https://bandit.readthedocs.io/

### Best Practices

TRUST 5 Implementation:
- Run all validations before commits
- Configure appropriate failure actions
- Document exceptions and waivers
- Track quality metrics over time

SPEC-First Development:
- Write clear EARS specifications
- Execute /clear between phases
- Maintain 85%+ test coverage
- Generate documentation with each feature

Agent Delegation:
- Match complexity to agent count
- Use parallel execution for independent tasks
- Implement proper error handling
- Track delegation outcomes

Token Management:
- Monitor usage proactively
- Clear context at phase boundaries
- Use selective file loading
- Choose appropriate model for task

### Version History

| Version | Date       | Changes                                           |
|---------|------------|---------------------------------------------------|
| 2.3.0   | 2025-12-03 | Added agents, commands, execution rules modules   |
| 2.2.0   | 2025-11-28 | Optimized to 500-line SKILL.md                    |
| 2.1.0   | 2025-11-25 | Added modular architecture patterns               |
| 2.0.0   | 2025-11-20 | Unified six foundational principles               |
| 1.0.0   | 2025-11-15 | Initial release with TRUST 5 and SPEC-First       |

---

Status: Reference Documentation Complete
Last Updated: 2025-12-06
Skill Version: 2.3.0
