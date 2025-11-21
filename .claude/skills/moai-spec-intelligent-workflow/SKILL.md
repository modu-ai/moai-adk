---
name: moai-spec-intelligent-workflow
description: Intelligent SPEC generation decision engine with 3-level templates and analytics for MoAI-ADK
---

# SPEC Intelligent Workflow Skill

## 🎯 Quick Reference (30 seconds)

**SPEC Intelligent Workflow** is Alfred's automated SPEC determination system.

Analyzing user requests with natural language processing:

- ✅ **0-1 conditions** → SPEC unnecessary (implement immediately)
- ✅ **2-3 conditions** → SPEC recommended (user choice)
- ✅ **4-5+ conditions** → SPEC strongly recommended (emphasized)

Automatically selects one of **3-level templates** and tracks effectiveness using **analytics system**.

---

## 📚 Core Features

### 1. Alfred's 5-Point Automatic Decision Criteria

```
① File modification scope (single file vs multiple files)
② Architecture impact (present/absent)
③ Component integration (single vs complex)
④ Implementation time (30-minute threshold)
⑤ Future maintenance (required/not required)
```

### 2. 3-Level SPEC Templates

```
Level 1 (Minimal)     → Simple tasks, 5-10 minutes writing
Level 2 (Standard)    → General features, 10-15 minutes writing
Level 3 (Comprehensive) → Complex tasks, 20-30 minutes writing
```

### 3. Analytics and Reporting

```
Session start: Auto-display SPEC statistics for last 30 days
Session end: Auto-collect SPEC-related data
Monthly report: Effectiveness analysis and improvement recommendations
```

---

## 🚀 Quick Start

### User Request → Automatic SPEC Decision

```
User: "Add user profile image upload functionality"
  ↓
Alfred Analysis: 4 conditions met → SPEC strongly recommended
  ↓
User Choice: "Yes, generate SPEC"
  ↓
Auto-run /moai:1-plan
  ↓
Level 2 template auto-selected
  ↓
SPEC-XXX generation complete
  ↓
/moai:2-run SPEC-XXX implementation
```

---

## 📁 Skill Structure

| File | Purpose | Size |
| --- | --- | --- |
| **README.md** | Overview and quick start | 5KB |
| **alfred-decision-logic.md** | Alfred decision algorithm detailed | 12KB |
| **templates.md** | 3-level SPEC templates and examples | 15KB |
| **analytics.md** | Analytics and reporting system design | 10KB |
| **examples.md** | 10+ real-world usage examples | 12KB |
| **FAQ.md** | Frequently asked questions | 5KB |

---

## 💡 Core Characteristics

- ✅ **Natural language processing**: No complex scoring calculations
- ✅ **Automatic selection**: 3-level templates auto-chosen
- ✅ **User autonomy**: All recommendations can be rejected
- ✅ **Data-driven**: Measure effectiveness via analytics
- ✅ **Balanced automation**: Prevent over-automation

---

## 🔗 Relationship with CLAUDE.md

```
CLAUDE.md (overview)
  ↓
This Skill (detailed implementation)
  ├── Alfred decision algorithm
  ├── 3-level templates complete definition
  ├── Analytics system design
  └── 10+ real-world examples
```

CLAUDE.md contains only an overview of this Skill; detailed content is referenced here.

---

## 📖 Documentation Guide

### Quick Understanding

→ **README.md** (5 minutes)

### Learning Alfred's Decision Criteria

→ **alfred-decision-logic.md** (10 minutes)

### SPEC Template Selection

→ **templates.md** (15 minutes)

### Understanding Analytics System

→ **analytics.md** (10 minutes)

### Real-world Usage Examples

→ **examples.md** (20 minutes)

### Resolving Questions

→ **FAQ.md** (5 minutes)

---

## 🎯 When to Use This Skill

This Skill is referenced in the following situations:

1. **When a user requests new work**

   - Alfred determines SPEC necessity
   - Review to understand 5-point criteria

2. **When deciding to generate a SPEC**

   - Confirm 3-level template selection criteria
   - Find the template matching your work

3. **When wanting to understand SPEC-First workflow effectiveness**

   - Understand analytics system
   - Analyze monthly reports

4. **When questions arise**
   - Check FAQ
   - Understand through real-world examples

---

## ✨ Expected Impact

| Metric | Expected Improvement |
| --- | --- |
| SPEC usage rate | 50% → 85% |
| Implementation time | 30% reduction |
| Code quality | 15% improvement |
| Test coverage | 80% → 90%+ |
| Bug reduction | 25% decrease |

---

## 🔗 Integration with MoAI-ADK

Works best with:
- `moai-core-spec-authoring` - SPEC generation and writing patterns
- `moai-core-ask-user-questions` - Interactive requirement clarification
- `moai-project-config-manager` - Project configuration management
- `moai-cc-hooks` - Automated workflow hooks

---

**Version**: 1.0.0
**Last Updated**: 2025-11-21
**Status**: Active - Production Ready
