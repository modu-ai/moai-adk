---
name: moai:custom-action
description: "Brief description of what this command does"
argument-hint: "required-arg [optional-arg] --option value"
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
model: "sonnet"  # Optional: haiku for simple tasks, sonnet for complex
skills:
  - relevant-skill-1
  - relevant-skill-2
---

# Custom Command Title

## 📋 Pre-execution Context

!git status --porcelain
!git branch --show-current
!git log --oneline -5

## 📁 Essential Files

@.moai/config/config.json
@.moai/project/structure.md
@CLAUDE.md

---

# Command Implementation

## 🎯 Purpose

Clear description of what this command accomplishes.

## 📝 Arguments

- **required-arg**: Description of required argument
- **optional-arg**: Description of optional argument (default: value)
- **--option**: Description of option flag

## 🔧 Implementation Steps

### Step 1: Initial Setup
- Validate arguments
- Load configuration
- Set up environment

### Step 2: Core Logic
- Main processing logic
- Error handling
- Progress tracking

### Step 3: Completion
- Generate output
- Update state
- Provide summary

## 📊 Output Format

Description of what the command outputs:
- Console output format
- File creation (if any)
- State changes
- Success indicators

## 🔗 Integration

- How this fits in the MoAI workflow
- Related commands
- Next steps for user

---

## 📝 Examples

### Basic Usage
```bash
/moai:custom-action arg1 --option value
```

### Advanced Usage
```bash
/moai:custom-action arg1 arg2 --option value --flag
```

---

## 🚀 Error Handling

Common error scenarios and resolutions:
- Error condition: Description and solution
- Validation failure: How to fix
- Permission issues: Resolution steps

---

**Generated with**: MoAI Command Template | **Last Updated**: 2025-11-24