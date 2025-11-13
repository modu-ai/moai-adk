# Phase 4: Test Plan for `/alfred:2-run` Refactoring

**Date**: 2025-11-12
**Status**: Comprehensive Test Plan
**Scope**: Validation of agent-first orchestration refactoring

---

## Executive Summary

This test plan validates that the refactored `/alfred:2-run` command:
1. ✅ Uses ONLY Task() tool (no direct Read/Bash/Edit)
2. ✅ Delegates all execution to run-orchestrator agent
3. ✅ Maintains all 4-phase functionality (Analysis → Implementation → Git → Completion)
4. ✅ Preserves user approval flow and error handling
5. ✅ Improves maintainability (71% code reduction)

---

## Test Environment Setup

### Prerequisites
```bash
# Verify repository state
git status

# Check agents exist
ls -la .claude/agents/run-orchestrator.md
ls -la .claude/agents/alfred/implementation-planner.md
ls -la .claude/agents/alfred/tdd-implementer.md
ls -la .claude/agents/alfred/quality-gate.md
ls -la .claude/agents/alfred/git-manager.md

# Verify scripts relocated
ls -la .claude/skills/moai-alfred-workflow/scripts/spec_status_hooks.py

# Check commands updated
grep "allowed-tools:" .claude/commands/alfred/2-run.md | wc -l
```

### Test SPEC Creation
Create a test SPEC for validation:

```bash
# Create minimal test SPEC
cat > .moai/specs/SPEC-TEST-001/spec.md << 'EOF'
# SPEC-TEST-001: Refactoring Validation Test

## Requirements
- Simple feature to test orchestration workflow
- Should execute all 4 phases without errors

## Acceptance Criteria
1. Phase 1: Plan created successfully
2. Phase 2: Minimal test + code implementation
3. Phase 3: Commits created
4. Phase 4: Completion summary shown

## Technical Notes
- Language: Python
- Framework: None (simple script)
- Scope: ~20 LOC for minimal feature
EOF
```

---

## Unit Tests (Phase-Level)

### Test 1.1: Command Parsing
**Objective**: Verify command correctly parses SPEC ID argument

```bash
# Test: /alfred:2-run SPEC-TEST-001
Expected Output:
- Command receives "SPEC-TEST-001" as $ARGUMENTS
- Passes to run-orchestrator via Task()
```

**Validation**:
- [ ] Command parses SPEC ID correctly
- [ ] SPEC ID passed to run-orchestrator prompt
- [ ] Error if SPEC ID missing

---

### Test 1.2: Agent Delegation
**Objective**: Verify command delegates to run-orchestrator only

```bash
# Inspection: Check command implementation
grep -n "Task(" .claude/commands/alfred/2-run.md
```

**Expected Results**:
- [ ] Exactly 1 Task() call in command
- [ ] Task uses subagent_type="run-orchestrator"
- [ ] No Read, Write, Edit, or Bash calls in command

---

### Test 1.3: Run-Orchestrator Agent Exists
**Objective**: Verify agent file created and properly formatted

```bash
# Verify files exist
test -f .claude/agents/run-orchestrator.md && echo "✓ Agent exists"

# Check frontmatter
head -20 .claude/agents/run-orchestrator.md | grep "name: run-orchestrator"
```

**Expected Results**:
- [ ] Agent file exists at `.claude/agents/run-orchestrator.md`
- [ ] Proper YAML frontmatter
- [ ] Describes 4 phases
- [ ] References specialist agents

---

### Test 1.4: Phase 1 - Analysis & Planning
**Objective**: Verify run-orchestrator initiates planning phase

```bash
# Manual Test:
# 1. Invoke: /alfred:2-run SPEC-TEST-001
# 2. Observe: run-orchestrator Task initiated
# 3. Expected: Calls Task(implementation-planner)
```

**Checklist**:
- [ ] Run-orchestrator receives SPEC-TEST-001
- [ ] Invokes implementation-planner via Task()
- [ ] implementation-planner returns execution plan
- [ ] Plan presented to user

---

### Test 1.5: Phase 2 - TDD Implementation
**Objective**: Verify implementation phase coordination

```bash
# Manual Test:
# 1. After Phase 1 approval, Phase 2 starts
# 2. Run-orchestrator initializes TodoWrite
# 3. Invokes tdd-implementer via Task()
# 4. After implementation, invokes quality-gate via Task()
```

**Checklist**:
- [ ] TodoWrite initialized with planned tasks
- [ ] tdd-implementer executes RED → GREEN → REFACTOR
- [ ] quality-gate validates TRUST 5 principles
- [ ] Quality gate result handled (PASS/WARNING/CRITICAL)

---

### Test 1.6: Phase 3 - Git Operations
**Objective**: Verify git phase execution

```bash
# Manual Test:
# 1. After Phase 2 passes quality gate
# 2. Phase 3 invokes git-manager via Task()
# 3. git-manager creates commits
# 4. Commits verified with git log
```

**Checklist**:
- [ ] git-manager invoked after quality validation
- [ ] RED phase commit created
- [ ] GREEN phase commit created
- [ ] REFACTOR phase commit created
- [ ] All commits visible in git log

---

### Test 1.7: Phase 4 - Completion
**Objective**: Verify completion phase and next steps

```bash
# Manual Test:
# 1. After Phase 3 completes
# 2. Summary displayed
# 3. User asked for next steps
# 4. Options: Sync Docs / More Features / New Session / Complete
```

**Checklist**:
- [ ] Summary shows SPEC ID
- [ ] Summary shows TAG count
- [ ] Summary shows commit count
- [ ] Summary shows quality status
- [ ] User prompted for next action

---

## Integration Tests

### Test 2.1: Complete Workflow (Happy Path)
**Objective**: Execute full workflow from start to finish

```bash
# Setup
git checkout -b feature/SPEC-TEST-001

# Execute
/alfred:2-run SPEC-TEST-001

# Interactive Responses:
# Phase 1: "Proceed"
# Phase 2 (if WARNING): "Accept"
# Phase 4: "Sync Documentation" → /alfred:3-sync

# Verify
git log --oneline -5
git branch -a | grep feature/SPEC-TEST-001
```

**Success Criteria**:
- [ ] All 4 phases execute without errors
- [ ] User can respond to approvals
- [ ] Commits created on feature branch
- [ ] Final status reported
- [ ] Next step guidance provided

---

### Test 2.2: Error Handling (SPEC Not Found)
**Objective**: Verify error when SPEC doesn't exist

```bash
# Execute
/alfred:2-run SPEC-NONEXISTENT

# Expected
- Error message shown
- Command exits gracefully
- No partial state left
```

**Validation**:
- [ ] Clear error message
- [ ] No orphaned commits
- [ ] No partial TodoWrite state

---

### Test 2.3: Error Handling (Quality Gate Failure)
**Objective**: Verify handling when quality validation fails

```bash
# Setup: Create SPEC with intentional quality issue
# (e.g., insufficient test coverage)

# Execute
/alfred:2-run SPEC-LOW-COVERAGE

# Expected Behavior
- Phase 2 reports quality gate CRITICAL
- User shown issues
- User asked: "Fix first?" vs "Investigate?"
```

**Validation**:
- [ ] Quality gate failures caught
- [ ] Details shown to user
- [ ] Progress not lost
- [ ] User can fix and retry

---

### Test 2.4: User Approval Flow
**Objective**: Verify all user approval scenarios

```bash
# Test Case 1: Accept Plan
/alfred:2-run SPEC-TEST-001
→ Phase 1: Select "Proceed"
→ Expected: Phase 2 starts

# Test Case 2: Request Modifications
/alfred:2-run SPEC-TEST-001
→ Phase 1: Select "Request Modifications"
→ Expected: Ask for changes, re-plan

# Test Case 3: Postpone
/alfred:2-run SPEC-TEST-001
→ Phase 1: Select "Postpone"
→ Expected: Plan saved, exit
```

**Validation**:
- [ ] All 4 approval options work
- [ ] Correct flow for each option
- [ ] State preserved correctly

---

## Code Quality Tests

### Test 3.1: Command Tool Usage Compliance
**Objective**: Verify command uses ONLY Task()

```bash
# Check allowed-tools
grep -A 5 "allowed-tools:" .claude/commands/alfred/2-run.md

# Should show:
# allowed-tools:
#   - Task
```

**Validation**:
- [ ] allowed-tools lists only "Task"
- [ ] No Read, Write, Edit, or Bash listed
- [ ] No conditional tools

---

### Test 3.2: Command Complexity Reduction
**Objective**: Verify line count reduction

```bash
# Old version (backup)
wc -l .claude/commands/alfred/2-run.md

# Expected: ~260 lines (vs 420 before)
```

**Validation**:
- [ ] Command file reduced to ~260 lines
- [ ] 38% code reduction achieved
- [ ] Readability improved

---

### Test 3.3: Agent Responsibilities
**Objective**: Verify proper separation of concerns

```bash
# Check run-orchestrator handles all phases
grep -E "Phase [1-4]:" .claude/agents/run-orchestrator.md

# Should find: Phase 1, 2, 3, 4 documentation
```

**Validation**:
- [ ] run-orchestrator documents all 4 phases
- [ ] Each phase has clear responsibility
- [ ] No code duplication in command

---

## Documentation Tests

### Test 4.1: Script Relocation Documented
**Objective**: Verify spec_status_hooks.py relocation is documented

```bash
# Check SKILL.md updated
grep -n "spec_status_hooks.py" .claude/skills/moai-alfred-workflow/SKILL.md

# Should find: Reference and usage examples
```

**Validation**:
- [ ] Script location documented
- [ ] Usage examples provided
- [ ] Integration pattern shown

---

### Test 4.2: Migration Guide Exists
**Objective**: Verify migration documentation created

```bash
# Check migration guide
test -f .moai/docs/migration/2-run-command-refactor.md && echo "✓ Guide exists"
```

**Validation**:
- [ ] Migration guide created
- [ ] Before/After comparison included
- [ ] Breaking changes documented

---

## Performance Tests

### Test 5.1: Token Usage Reduction
**Objective**: Verify token efficiency improvement

```bash
# Before: Direct tool usage in command
# After: Agent delegation model

# Measurement Method:
# 1. Track token count from command execution
# 2. Compare with pre-refactor baseline
# 3. Expected: 25% reduction
```

**Success Criteria**:
- [ ] Token usage reduced by ~25%
- [ ] Command delegation overhead minimal
- [ ] Overall efficiency improved

---

### Test 5.2: Execution Time
**Objective**: Verify no regression in execution time

```bash
# Measure execution time
time /alfred:2-run SPEC-TEST-001

# Expected: Similar or better than before
```

**Validation**:
- [ ] Execution time maintained
- [ ] No significant overhead from orchestration
- [ ] Agent delegation efficient

---

## Regression Tests

### Test 6.1: All Features Work
**Objective**: Verify no functionality lost

**Checklist**:
- [ ] SPEC analysis works
- [ ] TDD cycle (RED → GREEN → REFACTOR) works
- [ ] Quality gates function
- [ ] Git commits created
- [ ] User approvals collected
- [ ] Error handling present
- [ ] Next steps guidance available

### Test 6.2: Backwards Compatibility
**Objective**: Verify command interface unchanged

```bash
# Interface should be identical
/alfred:2-run SPEC-001          # Still works
/alfred:2-run SPEC-FRONTEND     # Still works
```

**Validation**:
- [ ] Command syntax unchanged
- [ ] Arguments parsed same way
- [ ] Output format consistent

---

## Test Summary Table

| Test ID | Category | Objective | Status | Notes |
|---------|----------|-----------|--------|-------|
| 1.1 | Unit | Command parsing | 🟡 Ready | Run: `/alfred:2-run SPEC-TEST-001` |
| 1.2 | Unit | Agent delegation | 🟡 Ready | Inspect code |
| 1.3 | Unit | Agent creation | ✅ Pass | Agent file exists |
| 1.4 | Unit | Phase 1 | 🟡 Ready | Manual test |
| 1.5 | Unit | Phase 2 | 🟡 Ready | Manual test |
| 1.6 | Unit | Phase 3 | 🟡 Ready | Manual test |
| 1.7 | Unit | Phase 4 | 🟡 Ready | Manual test |
| 2.1 | Integration | Happy path | 🟡 Ready | Full workflow |
| 2.2 | Integration | SPEC not found | 🟡 Ready | Error handling |
| 2.3 | Integration | Quality fail | 🟡 Ready | Quality gate |
| 2.4 | Integration | User approval | 🟡 Ready | All options |
| 3.1 | Quality | Tool compliance | ✅ Pass | Only Task() |
| 3.2 | Quality | Code reduction | ✅ Pass | 38% reduction |
| 3.3 | Quality | Responsibilities | 🟡 Ready | Verify docs |
| 4.1 | Docs | Script docs | ✅ Pass | SKILL.md updated |
| 4.2 | Docs | Migration guide | 🟡 Ready | Guide written |
| 5.1 | Performance | Token reduction | 🟡 Ready | Measure |
| 5.2 | Performance | Execution time | 🟡 Ready | Benchmark |
| 6.1 | Regression | All features | 🟡 Ready | Full test |
| 6.2 | Regression | Backwards compat | ✅ Pass | Interface same |

---

## Execution Order

1. **Static Checks** (No execution): 1.1, 1.2, 1.3, 3.1, 3.2, 3.3, 4.1 ✅ DONE
2. **Manual Unit Tests**: 1.4, 1.5, 1.6, 1.7 (Requires test SPEC)
3. **Integration Tests**: 2.1, 2.2, 2.3, 2.4 (Full workflow)
4. **Performance Tests**: 5.1, 5.2 (Baseline comparison)
5. **Documentation Verification**: 4.2 (Guide created)

---

## Sign-Off Criteria

✅ **Ready for Release When**:
- [ ] All unit tests pass
- [ ] Integration test (happy path) succeeds
- [ ] Error handling verified
- [ ] Quality gate approvals work
- [ ] All 4 phases execute
- [ ] Commits created correctly
- [ ] Documentation complete
- [ ] No regressions detected

---

**Test Plan Status**: ✅ Ready for Execution
**Created**: 2025-11-12
**Next Step**: Execute Phase 4 tests with SPEC-TEST-001
