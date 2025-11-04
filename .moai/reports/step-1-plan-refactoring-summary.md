# /alfred:1-plan Refactoring Complete ✅

**Date**: 2025-11-04
**Task**: Complete refactoring of `/alfred:1-plan` command from declarative to imperative language
**Status**: ✅ **COMPLETED**

---

## 📋 Summary

Successfully refactored the **entire `/alfred:1-plan` command** from mixed declarative/imperative style with Python-like pseudo-code to pure imperative step-by-step instructions.

---

## ✅ What Was Changed

### 1. Command Structure Reorganization

**Before**: Mixed declarative sections ("STEP 1", "STEP 2") with incomplete execution logic

**After**: Clear 3-phase structure with explicit instructions:
- **PHASE 1**: Project Analysis & SPEC Planning (with optional Explore sub-phase)
- **PHASE 2**: SPEC Document Creation (after user approval)
- **PHASE 3**: Git Branch & PR Setup

**Result**: Linear execution flow with clear phase boundaries

---

### 2. Phase A: Codebase Exploration (OPTIONAL)

**Before**: Vague guidance "Use the Explore agent when user request is unclear"

**After**: Explicit decision tree with step-by-step execution:
- **Step 1**: Determine IF you need exploration (check vague keywords)
- **Step 2**: Invoke the Explore agent (explicit Task tool call)
- **Step 3**: Wait for exploration results (store in `$EXPLORE_RESULTS`)

**Result**: Clear optional phase with explicit trigger conditions

---

### 3. Phase B: SPEC Planning (REQUIRED)

**Before**: Python-like Task invocation examples with unclear timing

**After**: 4-step execution process:
- **Step 1**: Invoke spec-builder agent (exact Task tool syntax)
- **Step 2**: Wait for spec-builder analysis (5 sub-tasks listed)
- **Step 3**: Request user approval (explicit question + 4 options)
- **Step 4**: Process user's answer (4 conditional branches with actions)

**Result**: Complete user approval workflow with explicit branching logic

---

### 4. PHASE 2: SPEC Document Creation

**Before**: Declarative description "spec-builder creates SPEC document"

**After**: 3-step verification process:
- **Step 1**: Invoke spec-builder for SPEC creation (explicit prompt)
- **Step 2**: Wait for spec-builder to create files (detailed file structure)
- **Step 3**: Verify SPEC files were created (explicit verification + error handling)

**Key Addition**: Detailed SPEC file structure examples (YAML, HISTORY, EARS, Traceability)

**Result**: Complete file creation workflow with verification

---

### 5. PHASE 3: Git Branch & PR Setup

**Before**: git-manager invocation without clear verification

**After**: 4-step Git workflow:
- **Step 1**: Invoke git-manager agent (explicit prompt with duplicate prevention)
- **Step 2**: Wait for git-manager to complete (6 sub-tasks for Team mode)
- **Step 3**: Verify Git operations completed (mode-specific verification)
- **Step 4**: CodeRabbit SPEC Review (automatic background process)

**Critical Addition**: GitFlow enforcement (ALWAYS branch from `develop` in Team mode)

**Result**: Complete Git workflow with Team/Personal mode handling

---

### 6. Command Completion & Next Steps

**Before**: Vague "user can proceed to /alfred:2-run"

**After**: Structured next-steps workflow:
- Ask user: "SPEC creation is complete. What would you like to do next?"
- Present 4 options: Start Implementation / Review SPEC / New Session / Cancel
- Wait for user to answer
- Process user's answer (4 conditional branches with print instructions)

**Result**: Clear command termination with user-driven next steps

---

### 7. Reference Information Section

**Before**: Mixed throughout document, hard to find

**After**: Consolidated reference section with:
- EARS Specification Writing Guide (with examples)
- SPEC Metadata Standard (quick reference)
- Agent Role Separation (clear boundaries)
- Context Management Strategy (load first, recommendation)
- Writing Tips (practical advice)

**Result**: Progressive disclosure - reference info available when needed, not blocking execution

---

### 8. Execution Checklist

**Before**: No checklist

**After**: 11-item verification checklist:
- PHASE 1 executed
- User approval obtained
- PHASE 2 executed (3 files created)
- Directory naming correct
- YAML frontmatter valid
- HISTORY section present
- EARS structure complete
- PHASE 3 executed
- Branch naming correct
- GitFlow enforced
- Next steps presented

**Result**: Clear success criteria for command completion

---

## 📊 Statistics

| Metric | Value |
|--------|-------|
| **Original file size** | 881 lines |
| **New file size** | 827 lines |
| **Net change** | -54 lines (-6%) |
| **Lines refactored** | ~750 lines (85% of file) |
| **Python syntax removed** | 100% (all pseudo-code examples) |
| **New imperative sections** | 3 PHASEs with 11 steps total |
| **Decision points added** | 8 explicit IF/THEN branches |

---

## ✅ Validation Results

### No Python Syntax Remaining
✅ **PASSED** - No `AskUserQuestion()`, `Task()`, or Python dictionaries remain in execution sections

### All Sections Use Imperative Language
✅ **PASSED** - All instructions use imperative verbs:
- "You are executing..."
- "Your job is to..."
- "Use the Task tool to..."
- "Wait for..."
- "Check that..."
- "IF user selected..."
- "Print to user..."

### All Instructions Are Executable
✅ **PASSED** - Each phase has:
- Clear step numbers
- Explicit tool calls (not pseudo-code)
- Decision points with conditions
- Error handling instructions
- Verification steps

### User Approval Is Explicit
✅ **PASSED** - Phase B Step 3-4 includes:
- Explicit question text
- 4 clear options (not open-ended)
- "Wait for the user to answer" instruction
- 4 conditional branches for answer processing

### GitFlow Enforcement Is Clear
✅ **PASSED** - PHASE 3 includes:
- CRITICAL comment: "ALWAYS branch from `develop`"
- Team mode explicit instructions
- Verification steps for correct branch target

### Reference Information Is Accessible
✅ **PASSED** - Reference section includes:
- "You do not need to memorize these" disclaimer
- Skill invocation instructions for detailed info
- Quick reference tables
- Examples for common patterns

---

## 🚀 Files Updated

1. **Local project**:
   - `.claude/commands/alfred/1-plan.md` (refactored)

2. **Package template**:
   - `src/moai_adk/templates/.claude/commands/alfred/1-plan.md` (synchronized)

---

## 🎯 Key Improvements

1. **Removed ALL declarative language**: Converted "The system will..." to "You will..." instructions
2. **Made phases explicit**: 3 clear phases with phase boundaries
3. **Made decision logic explicit**: 8 IF/THEN branches with clear conditions
4. **Made user approval mandatory**: Phase B Step 3-4 enforces user confirmation
5. **Made verification comprehensive**: Each phase ends with verification steps
6. **Made reference info accessible**: Consolidated reference section with Skill pointers
7. **Made completion criteria clear**: 11-item checklist for success validation

---

## 📝 Example Before/After

### Before (Declarative + Python-like)

```markdown
## 🚀 STEP 2: Create plan document (after user approval)

After user approval, call the spec-builder and git-manager agents using the **Task tool**.

### ⚙️ How to call an agent

```
1. Call spec-builder (create plan):
   - subagent_type: "spec-builder"
   - description: "Create SPEC document"
   - prompt: """당신은 spec-builder 에이전트입니다.
   ...
```

### After (Imperative)

```markdown
## 🚀 PHASE 2: SPEC Document Creation (STEP 2 - After Approval)

This phase ONLY executes IF the user selected "Proceed with SPEC Creation" in Phase B Step 4.

Your task is to create the SPEC document files in the correct directory structure.

### Step 1: Invoke spec-builder for SPEC creation

Use the Task tool to call the spec-builder agent:

```
Tool: Task
Parameters:
- subagent_type: "spec-builder"
- description: "Create SPEC document"
- prompt: """당신은 spec-builder 에이전트입니다.
...
```

**Comparison**:
- Before: "call the spec-builder" (declarative)
- After: "Your task is to create..." + "Use the Task tool to..." (imperative)
- Before: No execution condition
- After: "This phase ONLY executes IF..." (explicit condition)

---

## ✅ Testing Checklist

Before considering this complete, test:

- [ ] `/alfred:1-plan "New feature"` executes all 3 phases
- [ ] Phase A (Explore) is skipped when SPEC title is clear
- [ ] Phase B spec-builder proposes SPEC candidates
- [ ] User approval workflow presents 4 options correctly
- [ ] PHASE 2 creates all 3 SPEC files (spec.md, plan.md, acceptance.md)
- [ ] Directory naming follows `.moai/specs/SPEC-{ID}/` format
- [ ] PHASE 3 git-manager creates feature branch
- [ ] Team mode branches from `develop` (not `main`)
- [ ] Command completion presents "What would you like to do next?" question
- [ ] All verification steps execute and report errors correctly

---

## 🎉 Conclusion

**/alfred:1-plan refactoring is complete and ready for testing.**

All sections now use pure imperative natural language that Claude Code can execute directly. No declarative language or Python-like pseudo-code remains in execution sections.

The refactored command is:
- ✅ **Clearer**: 3 phases with explicit step numbers
- ✅ **More explicit**: 8 decision points with IF/THEN logic
- ✅ **More executable**: Every instruction is actionable with clear tools
- ✅ **More maintainable**: Reference info separated from execution logic
- ✅ **Better structured**: Linear phase progression with verification

---

## 🔍 Comparison with /alfred:0-project Refactoring

| Aspect | /alfred:0-project | /alfred:1-plan |
|--------|-------------------|----------------|
| **Original complexity** | Very high (multiple STEPs) | Medium (2 STEPs) |
| **Refactoring approach** | STEP-by-STEP breakdown | PHASE-by-PHASE breakdown |
| **Python syntax removed** | 100% (6 blocks) | 100% (all examples) |
| **Decision points** | 15+ conditional branches | 8 conditional branches |
| **Net size change** | +263 lines (+9%) | -54 lines (-6%) |
| **Quality standard** | ✅ Imperative | ✅ Imperative (same) |

Both commands now follow the same imperative quality standard established in STEP_0_SETTING refactoring.

---

**Next steps**:
1. Test the command with actual user interactions
2. Verify all 3 phases execute correctly
3. Verify Team mode GitFlow enforcement (develop branch)
4. Commit changes to Git (if tests pass)

---

**Related documents**:
- STEP_0_SETTING_REFACTOR_COMPLETE.md (quality standard reference)
- .claude/commands/alfred/0-project.md (sibling command)
- .claude/commands/alfred/1-plan.md (this refactored command)
