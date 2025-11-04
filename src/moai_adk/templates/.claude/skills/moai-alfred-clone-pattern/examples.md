# Clone Pattern Examples

## Basic Usage Examples

### Example 1: Large-Scale Migration

**Scenario**: Migrate project from v0.14.0 to v0.15.2

```python
# Alfred's analysis
task = UserRequest(
    type="migration",
    description="Migrate entire MoAI-ADK project from v0.14.0 to v0.15.2",
    scope="entire_project",
    steps=8,
    domains=["config", "hooks", "permissions", "documentation"],
    uncertainty=0.8  # High uncertainty - new structure
)

# Decision: Apply Clone pattern
if should_create_clone(task):
    clone = create_clone(
        task_description="Migrate v0.14.0 project structure to v0.15.2",
        context_scope="full",
        learning_enabled=True
    )
    
    # Clone executes autonomously:
    # 1. Analyzes current structure
    # 2. Creates migration plan
    # 3. Backs up critical files
    # 4. Applies migration incrementally
    # 5. Validates each step
    # 6. Handles conflicts automatically
    # 7. Creates PR for review
    # 8. Saves learnings to memory
```

**Clone's Execution Log**:
```
Phase 1: Structure Analysis
- Found 47 files needing updates
- Identified 3 breaking changes
- Created backup: .moai/backup/v0.14.0/

Phase 2: Migration Planning
- Planned 8-step incremental migration
- Estimated risk: medium
- Created rollback checkpoints

Phase 3: Execution
- Step 1: Updated config.json structure ✅
- Step 2: Migrated hook files ✅
- Step 3: Updated permissions ✅
- Step 4: Fixed import paths ✅
- Step 5: Updated documentation ✅
- Step 6: Validated all tests ✅
- Step 7: Created PR ✅
- Step 8: Saved learnings ✅

Result: Migration completed successfully
Time: 45 minutes (vs estimated 2 hours manual)
Issues: 2 minor conflicts resolved automatically
```

### Example 2: Parallel File Refactoring

**Scenario**: Update import statements across 150+ Python files

```python
# Alfred's analysis
task = UserRequest(
    type="refactoring",
    description="Update all import statements from old package structure",
    files_affected=150,
    pattern_change="from moai.adk import X → from moai_adk import X",
    complexity="high",
    parallelizable=True
)

# Decision: Apply Clone pattern with parallel processing
if should_create_clone(task):
    # Create 3 clones for parallel processing
    file_batches = split_files_into_batches("/src", batch_size=50)
    
    clones = [
        create_clone(f"Update imports in batch {i+1}: {batch}")
        for i, batch in enumerate(file_batches)
    ]
    
    # Execute in parallel
    results = parallel_execute(clones)
    
    # Merge results
    merge_refactoring_results(results)
```

**Parallel Execution Results**:
```
Clone 1: Files 1-50
- Updated 1,247 import statements
- Fixed 47 broken imports
- Time: 12 minutes

Clone 2: Files 51-100  
- Updated 1,156 import statements
- Fixed 38 broken imports
- Time: 11 minutes

Clone 3: Files 101-150
- Updated 1,089 import statements
- Fixed 42 broken imports
- Time: 13 minutes

Total Time: 13 minutes (vs 45 minutes sequential)
Success Rate: 100%
Tests: All passing
```

### Example 3: Architecture Exploration

**Scenario**: Evaluate different architecture approaches for new feature

```python
# Alfred's analysis
task = UserRequest(
    type="exploration",
    description="Evaluate 3 different architectures for new feature X",
    options=["microservices", "monolith", "modular_monolith"],
    evaluation_criteria=["performance", "maintainability", "scalability", "development_speed"],
    independence="high"  # Each architecture can be evaluated separately
)

# Decision: Apply Clone pattern for parallel exploration
if task.independence > 0.7:
    clones = []
    
    for architecture in task.options:
        clone = create_clone(
            f"Evaluate {architecture} architecture for feature X",
            context_scope="full"
        )
        clones.append(clone)
    
    # Execute all clones in parallel
    evaluation_results = parallel_execute(clones)
    
    # Synthesize results
    recommendation = synthesize_evaluations(evaluation_results)
```

**Evaluation Results**:
```
Clone 1: Microservices Architecture
- Performance: Excellent (9/10)
- Maintainability: Good (7/10)
- Scalability: Excellent (9/10)
- Development Speed: Poor (4/10)
- Recommendation: Best for large-scale, long-term projects

Clone 2: Monolith Architecture  
- Performance: Good (7/10)
- Maintainability: Poor (4/10)
- Scalability: Poor (3/10)
- Development Speed: Excellent (9/10)
- Recommendation: Good for small projects, quick MVP

Clone 3: Modular Monolith
- Performance: Good (8/10)
- Maintainability: Good (8/10)
- Scalability: Good (7/10)
- Development Speed: Good (7/10)
- Recommendation: Balanced approach, recommended for current project

Final Recommendation: Modular Monolith (best balance for current needs)
```

## Advanced Examples

### Example 4: Multi-Phase Project Restructuring

**Scenario**: Complete project restructuring with multiple phases

```python
# Complex multi-step task requiring strategic decisions
complex_task = UserRequest(
    type="restructuring",
    description="Restructure entire project for better maintainability",
    phases=[
        "directory_reorganization",
        "package_refactoring", 
        "documentation_update",
        "test_migration",
        "ci_cd_update",
        "deployment_changes"
    ],
    estimated_duration="2-3 days",
    risk_level="high"
)

# Create master clone for coordination
master_clone = create_clone(
    "Coordinate project restructuring across 6 phases",
    context_scope="full",
    learning_enabled=True
)

# Master clone creates specialist clones for each phase
phase_clones = []
for phase in complex_task.phases:
    phase_clone = master_clone.create_sub_clone(
        f"Execute {phase} phase",
        dependencies=get_phase_dependencies(phase)
    )
    phase_clones.append(phase_clone)

# Execute with dependency management
execution_plan = master_clone.create_execution_plan(phase_clones)
results = master_clone.execute_with_dependencies(execution_plan)
```

### Example 5: Learning-Based Optimization

**Scenario**: Use previous learnings to optimize current task

```python
# Check for similar previous tasks
previous_learnings = load_clone_learnings("migration")

current_task = UserRequest(
    type="migration",
    description="Migrate from v0.15.2 to v0.16.0",
    similarity_score=calculate_similarity(current_task, previous_learnings)
)

if current_task.similarity_score > 0.8:
    # Use optimized approach based on learnings
    optimized_approach = apply_learnings(
        previous_learnings,
        current_task
    )
    
    clone = create_clone(
        "Optimized migration based on previous experience",
        learning_strategy="incremental_improvement"
    )
    
    # Clone starts with proven approach
    clone.apply_pre_validated_strategy(optimized_approach)
```

## Error Handling Examples

### Example 6: Clone Recovery from Failure

**Scenario**: Clone encounters unexpected conflict

```python
# Clone encounters merge conflict
try:
    clone.execute_migration_step()
except MergeConflict as conflict:
    # Auto-recovery attempt 1: Analyze conflict
    conflict_analysis = clone.analyze_conflict(conflict)
    
    if conflict_analysis.resolvable_automatically:
        # Apply automatic resolution
        resolution = clone.resolve_conflict_automatically(conflict)
        clone.continue_execution()
    else:
        # Save state and escalate to main Alfred
        clone.save_failure_state(conflict)
        clone.report_to_main_alfred({
            "status": "blocked",
            "conflict": conflict_analysis,
            "suggestions": ["manual_resolution", "alternative_approach"]
        })
```

## Performance Optimization Examples

### Example 7: Resource-Efficient Clone Usage

**Scenario**: Optimize clone usage for resource management

```python
# Resource monitoring during clone execution
resource_monitor = CloneResourceMonitor()

def create_resource_aware_clone(task_description):
    """Create clone with resource constraints"""
    return create_clone(
        task_description,
        resource_limits={
            "max_tokens": 10000,
            "max_execution_time": 1800,  # 30 minutes
            "max_memory_usage": "512MB"
        },
        checkpoint_strategy="incremental"
    )

# Batch processing to minimize resource usage
tasks = get_pending_tasks()
optimized_batches = optimize_for_resources(tasks)

for batch in optimized_batches:
    clones = [create_resource_aware_clone(task) for task in batch]
    
    # Monitor resource usage
    with resource_monitor:
        results = parallel_execute(clones)
        
        # Adjust batch size based on performance
        if resource_monitor.average_usage() > 0.8:
            batch_size = max(1, batch_size - 1)
        elif resource_monitor.average_usage() < 0.5:
            batch_size = min(4, batch_size + 1)
```

## Integration Examples

### Example 8: Integration with Git Workflow

**Scenario**: Clone integrates with GitFlow for proper branch management

```python
# Clone creates proper GitFlow branches
def create_gitflow_clone(feature_spec):
    clone = create_clone(
        f"Implement {feature_spec.id}",
        git_integration=True
    )
    
    # Clone handles Git workflow automatically
    clone.setup_gitflow_environment({
        "feature_branch": f"feature/{feature_spec.id}",
        "target_branch": "develop",
        "pr_template": "standard_feature_pr"
    })
    
    return clone

# Clone execution includes Git operations
clone.execute_with_gitflow([
    "create_feature_branch",
    "implement_changes", 
    "run_tests",
    "create_commit",
    "push_to_remote",
    "create_pr",
    "wait_for_review",
    "merge_to_develop"
])
```
