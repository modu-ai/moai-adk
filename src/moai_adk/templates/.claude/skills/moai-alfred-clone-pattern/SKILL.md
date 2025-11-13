---
name: "moai-alfred-clone-pattern"
version: "4.0.0"
created: 2025-11-05
updated: 2025-11-12
status: stable
description: Enterprise Master-Clone pattern implementation guide for complex multi-step tasks with full project context, autonomous delegation, parallel processing, and intelligent task distribution; activates for large-scale migrations, complex refactoring, parallel exploration, architecture restructuring, and multi-file transformations
keywords: ['clone-pattern', 'master-clone', 'delegation', 'multi-step', 'parallel-processing', 'autonomous-agents', 'task-distribution', 'project-context', 'complex-workflows', 'enterprise-delegation']
allowed-tools: 
  - Read
  - Bash
  - Task
---

# Enterprise Master-Clone Pattern Skill v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-clone-pattern |
| **Version** | 4.0.0 Enterprise (2025-11-12) |
| **Allowed tools** | Read, Bash, Task |
| **Auto-load** | On demand for complex multi-step tasks |
| **Tier** | Alfred (Orchestration) |
| **Lines of Content** | 900+ with 12+ enterprise examples |
| **Progressive Disclosure** | 3-level (quick-start, patterns, advanced) |

---

## What It Does

Provides comprehensive guidance for Alfred's **Master-Clone pattern** - a delegation mechanism where Alfred creates autonomous clones (Task-delegated agents) to handle complex multi-step tasks that don't require domain-specific expertise but benefit from:
- Full project context and codebase understanding
- Parallel processing capabilities
- Independent decision-making
- Comprehensive state tracking

---

## When to Use (Decision Framework)

**Use Clone Pattern when:**
- Task requires 5+ sequential steps OR affects 100+ files
- No domain-specific expertise needed (not UI, Backend, DB, Security, ML)
- Task is complex with high uncertainty
- Parallel processing would be beneficial
- Full project context is required for optimal decisions
- Task can run autonomously without continuous user input

**Examples**:
- Large-scale migrations (v0.14.0 → v0.15.2 affecting 200 files)
- Refactoring across many files (100+ imports, API changes)
- Parallel exploration/evaluation tasks
- Complex architecture restructuring
- Bulk file transformations with context-aware logic
- Schema migrations affecting multiple services
- Dependency upgrade cascades

**DON'T use Clone Pattern when:**
- Domain expertise needed (use specialist agents instead)
- Task < 5 steps (direct execution is faster)
- Quick yes/no decision (use AskUserQuestion)
- Single file modification (use tdd-implementer)

---

## Key Concepts

### Master-Clone Architecture

```
Master Agent (Alfred)
    ↓ Creates with Task()
Clone Agent #1          Clone Agent #2          Clone Agent #3
(Parallel execution with shared context)
    ↓                       ↓                       ↓
[Exploration]        [Analysis]             [Implementation]
    ↓                       ↓                       ↓
Results aggregation & synthesis by Master
    ↓
User presentation + next steps
```

**Master Responsibilities**:
- Analyze task scope and complexity
- Decompose into independent parallel work
- Create clones with appropriate context
- Aggregate and synthesize results
- Present findings to user

**Clone Responsibilities**:
- Execute assigned sub-task autonomously
- Use full project context for intelligent decisions
- Track state and progress
- Report findings with evidence
- Handle errors gracefully

---

## 3-Level Architecture

### Level 1: Simple Parallel Task

**Scenario**: Explore multiple implementation approaches in parallel

```typescript
// Master agent: Create parallel clones for exploration
const clones = [
  Task({
    description: "Explore PostgreSQL implementation for user persistence",
    prompt: "Analyze PostgreSQL libraries, schema design, migration strategy. Provide pros/cons and code examples."
  }),
  Task({
    description: "Explore MongoDB implementation for user persistence",
    prompt: "Analyze MongoDB libraries, document schema, migration strategy. Provide pros/cons and code examples."
  }),
  Task({
    description: "Explore Supabase implementation for user persistence",
    prompt: "Analyze Supabase SDK, schema design, migration strategy. Provide pros/cons and code examples."
  })
];

// Wait for all parallel clones to complete
const [postgresAnalysis, mongoAnalysis, supabaseAnalysis] = await Promise.all(clones);

// Master synthesizes results
const comparison = {
  options: [postgresAnalysis, mongoAnalysis, supabaseAnalysis],
  recommendation: selectBestOption(clones),
  tradeoffs: analyzeTradeoffs(clones)
};
```

### Level 2: Sequential Dependent Tasks

**Scenario**: Complex migration where later steps depend on earlier analysis

```typescript
// Step 1: Analyze current state
const analysisResult = await Task({
  description: "Analyze v0.14.0 codebase structure",
  prompt: "Scan project for all imports of 'old-api'. Document usage patterns, edge cases, and dependencies. Provide summary with file-by-file breakdown."
});

// Step 2: Plan migration strategy (depends on analysis)
const planResult = await Task({
  description: "Plan migration strategy from v0.14.0 to v0.15.2",
  prompt: `Using this analysis: ${analysisResult}\n\nCreate a step-by-step migration plan with:\n- Phased approach (phase 1, 2, 3)\n- Risk mitigation\n- Testing strategy\n- Rollback procedure`
});

// Step 3: Execute migration (depends on plan)
const migrationResult = await Task({
  description: "Execute v0.14.0 → v0.15.2 migration",
  prompt: `Using this plan: ${planResult}\n\nExecute the migration:\n- Update imports\n- Modify APIs\n- Update tests\n- Verify compatibility`
});

// Step 4: Validate results (depends on migration)
const validationResult = await Task({
  description: "Validate migration completeness",
  prompt: `Verify migration:\n- All imports updated\n- No breaking changes\n- Tests passing\n- Performance metrics maintained`
});

// Master reports final state
return {
  analysis: analysisResult,
  plan: planResult,
  migration: migrationResult,
  validation: validationResult,
  status: validationResult.passed ? "SUCCESS" : "NEEDS_REVIEW"
};
```

### Level 3: Hybrid Parallel + Sequential (Advanced)

**Scenario**: Large refactoring with parallel analysis, synchronized implementation

```typescript
// Phase 1: Parallel analysis clones
const [apiAnalysis, dbAnalysis, authAnalysis] = await Promise.all([
  Task({ description: "Analyze API layer usage...", prompt: "..." }),
  Task({ description: "Analyze DB layer usage...", prompt: "..." }),
  Task({ description: "Analyze Auth layer usage...", prompt: "..." })
]);

// Phase 2: Synchronized implementation (waits for all analyses)
const [apiRefactor, dbRefactor, authRefactor] = await Promise.all([
  Task({
    description: "Refactor API layer",
    prompt: `Based on analysis:\n${apiAnalysis}\n\nRefactor API with:\n- New patterns\n- Tests\n- Documentation`
  }),
  Task({
    description: "Refactor DB layer",
    prompt: `Based on analysis:\n${dbAnalysis}\n\nRefactor DB with:\n- Schema updates\n- Migration scripts\n- Tests`
  }),
  Task({
    description: "Refactor Auth layer",
    prompt: `Based on analysis:\n${authAnalysis}\n\nRefactor Auth with:\n- New strategy\n- Migration\n- Tests`
  })
]);

// Phase 3: Integration validation
const integrationResult = await Task({
  description: "Validate refactored layer integration",
  prompt: `Verify all refactored layers work together:\n- API uses new DB patterns\n- Auth integrates with API\n- No breaking changes\n- All tests passing`
});

return {
  phase1: { apiAnalysis, dbAnalysis, authAnalysis },
  phase2: { apiRefactor, dbRefactor, authRefactor },
  phase3: integrationResult,
  status: "COMPLETE"
};
```

---

## Best Practices

### DO
- **Define clear sub-tasks**: Each clone should have a specific, measurable goal
- **Provide full context**: Include project structure, existing patterns, constraints
- **Use sequential when needed**: Depend on earlier results when necessary
- **Validate results**: Always verify clone outputs before proceeding
- **Document findings**: Track what each clone discovered
- **Handle failures gracefully**: Plan for individual clone failures
- **Aggregate intelligently**: Synthesize parallel results into coherent analysis

### DON'T
- **Over-parallelize**: Creating 20 clones is overkill (use 2-5)
- **Under-specify tasks**: Vague descriptions lead to mediocre results
- **Ignore dependencies**: Force sequential when tasks actually depend
- **Skip validation**: Trust but verify clone outputs
- **Lose context**: Always include relevant project information
- **Create circular dependencies**: Avoid Task A waiting on Task B waiting on Task A

---

## Implementation Patterns

### Pattern 1: Exploration with Synthesis
```typescript
// Create 3 clones exploring different approaches
const explorations = await Promise.all(approaches.map(approach =>
  Task({
    description: `Explore ${approach.name} approach`,
    prompt: `Research ${approach.name}...\nProvide: pros, cons, code example, learning curve`
  })
));

// Master synthesizes into comparison table
return {
  comparison: createComparisonTable(explorations),
  recommendation: selectBestApproach(explorations),
  decisionRationale: explainDecision(explorations)
};
```

### Pattern 2: Phased Migration
```typescript
// Phase 1: Analyze
const analysis = await analyzeCurrentState();

// Phase 2: Plan (depends on analysis)
const plan = await planMigration(analysis);

// Phase 3: Implement (depends on plan)
const implementation = await implementMigration(plan);

// Phase 4: Validate (depends on implementation)
const validation = await validateMigration(implementation);

return { analysis, plan, implementation, validation };
```

### Pattern 3: Distributed Refactoring
```typescript
// Split files into groups, refactor each group in parallel
const fileGroups = splitFilesIntoGroups(files, 5);
const refactorResults = await Promise.all(
  fileGroups.map(group =>
    Task({
      description: `Refactor files: ${group.join(", ")}`,
      prompt: `Refactor these files using new patterns:\n${group.join("\n")}`
    })
  )
);

// Master validates all refactored files work together
return validateIntegration(refactorResults);
```

---

## When NOT to Use (Anti-Patterns)

| Scenario | Why | Use Instead |
|----------|-----|------------|
| Single file change | Too much overhead | Direct tdd-implementer |
| 2-3 quick steps | Sequential simpler | Direct execution |
| Domain expertise required | Needs specialist | Specialist agent (security, DB, etc.) |
| Real-time interaction | Clones run independently | Interactive agent |
| Simple query | Overkill complexity | Direct lookup |

---

## Related Skills

- `moai-alfred-agent-guide` (Agent architecture & delegation)
- `moai-alfred-task-decomposition` (Breaking down complex tasks)
- `moai-essentials-refactor` (Refactoring patterns & examples)

---

**For detailed API specifications**: [reference.md](reference.md)  
**For real-world examples**: [examples.md](examples.md)  
**Last Updated**: 2025-11-12  
**Status**: Production Ready (Enterprise v4.0.0)
