# SPEC Intelligent Workflow Skill

## What is This Skill?

A core skill that realizes MoAI-ADK's **SPEC-First TDD workflow**.

**Alfred analyzes user requests** to automatically determine SPEC necessity,
**selects the appropriate one of 3-level templates**, and **tracks effectiveness via analytics**.

### Core Value

```
❌ Before: Users must always decide SPEC necessity → Burden
✅ After: Alfred automatically decides and proposes → Natural workflow
```

---

## 🎯 3 Core Features

### 1️⃣ Alfred's Intelligent Decision Making

Analyzes using **5 questions with natural language processing**:

```
① Modifying or creating multiple files?
② Architecture or data model changes?
③ Integration between multiple components required?
④ Expected implementation time over 30 minutes?
⑤ Future maintenance or expansion needed?
```

**Automatic Decision**:
- `0-1 "yes" answers` → SPEC unnecessary (implement immediately)
- `2-3 "yes" answers` → SPEC recommended (user choice)
- `4-5 "yes" answers` → SPEC strongly recommended (emphasized)

### 2️⃣ 3-Level SPEC Templates

Alfred automatically selects:

| Level | Target | Sections | Writing Time | Characteristics |
|-------|--------|----------|--------------|-----------------|
| **Level 1** | Simple modifications | 5 | 5-10 min | Fast and concise |
| **Level 2** | General features | 7 | 10-15 min | EARS format |
| **Level 3** | Complex tasks | 10+ | 20-30 min | Architecture design included |

### 3️⃣ Analytics and Reporting

**Automatically tracked metrics**:

```
Session start:
  📊 SPEC statistics for last 30 days
     • Number created
     • Average completion time
     • Code linkage rate
     • Test coverage

Session end:
  📈 Auto-collect data
     • Git commit linkage
     • Modified file tracking
     • Test results recording

Monthly:
  📋 Auto-generate report
     • Trend analysis
     • Improvement recommendations
```

---

## 📖 Quick Start

### Scenario A: Simple task

```
User: "Change login button color"

Alfred Analysis:
  ① File modification: 1 file only → No
  ② Architecture: No changes → No
  ③ Integration: Not needed → No
  ④ Time: 5 minutes → No
  ⑤ Maintenance: Not needed → No

Conclusion: 0 conditions met → SPEC unnecessary

→ Proceed with immediate implementation
```

### Scenario B: Medium complexity

```
User: "Add user profile image upload functionality"

Alfred Analysis:
  ① File modification: 4 files (Backend, Frontend, DB) → Yes
  ② Architecture: Add file upload flow → Yes
  ③ Integration: 3 components → Yes
  ④ Time: 2 hours → Yes
  ⑤ Maintenance: Required → Yes

Conclusion: 5 conditions met → SPEC strongly recommended

User choice: "Yes, generate SPEC"

→ Auto-run /moai:1-plan
→ Auto-select Level 2 (Standard) template
→ Generate SPEC-XXX
→ /moai:2-run SPEC-XXX implementation
```

### Scenario C: Prototype

```
User: "I want to quickly create a prototype"

Alfred Analysis: Detects "prototype" keyword

→ Skip SPEC, implement immediately
→ After completion, recommend SPEC for production transition
```

---

## 🔄 Alfred's SPEC Decision Flow

```
┌─────────────────┐
│  User Request   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  @agent-Plan    │    (optional)
│  Execute or     │
│  Analyze chat   │
└────────┬────────┘
         │
         ▼
┌──────────────────────┐
│ Alfred analyzes      │
│ 5 conditions via     │
│ natural language     │
└────────┬─────────────┘
         │
         ├─ 0-1 ──→ SPEC unnecessary ──→ Implement immediately
         │
         ├─ 2-3 ──→ SPEC recommended ──→ AskUserQuestion
         │                            │
         │                            ├─ User "Yes" ──→ /moai:1-plan
         │                            │
         │                            └─ User "No" ──→ Implement immediately
         │
         └─ 4-5 ──→ SPEC strongly recommended ──→ Emphasized proposal
                                              │
                                              ├─ "Yes" ──→ /moai:1-plan
                                              │
                                              └─ "No" ──→ Implement immediately
                                                       │
                                                       ▼
                                              If complexity increases
                                              during implementation,
                                              propose SPEC
```

---

## 📚 Documentation Guide

### 🔍 Reading by Understanding Level

#### Understand in 5 minutes (very fast)
→ Read this README.md

#### Fully understand in 15 minutes (fast)
→ Read **alfred-decision-logic.md**
   - Alfred's 5-point decision criteria detailed
   - 3 real-world examples

#### Know everything in 30 minutes (sufficient)
→ Above + read **templates.md**
   - Complete 3-level template understanding
   - Template selection criteria
   - 3 actual examples

#### In-depth understanding in 1 hour (very detailed)
→ Above + read **analytics.md**
   - Analytics system design
   - SessionStart/End Hook
   - Monthly report

#### Expert level in 2 hours (complete)
→ All documents + read **examples.md**
   - 10+ real-world use cases
   - Various scenarios

---

## ❓ Frequently Asked Questions

### Q: Is SPEC really necessary?
A: Alfred decides! Users only need to choose.

### Q: Do I need to use SPEC for every task?
A: No. Simple tasks are implemented directly without SPEC.

### Q: Doesn't creating SPEC take a long time?
A: AI auto-generates 80%, so it only takes 5-30 minutes.

### Q: Can I reject SPEC suggestions?
A: Yes, all suggestions can be rejected. It's not forced.

See **FAQ.md** for more questions

---

## 🎯 This Skill's Goals

### Problem Solving
```
❌ Before: "Must decide whether to write SPEC" (user burden)
✅ After: "Alfred decides and proposes" (natural workflow)
```

### Measure Effectiveness
```
❌ Before: "Can't know if SPEC really helps"
✅ After: "Confirm 30% time savings via analytics"
```

### Continuous Improvement
```
❌ Before: "One-time document"
✅ After: "Monthly report reveals improvements, then optimize"
```

---

## 🚀 Next Steps

### After Reading This Skill

1. **Read alfred-decision-logic.md**
   - Understand 5-point decision criteria

2. **Read templates.md**
   - Learn 3-level template selection criteria

3. **Start actual work**
   - Create SPEC based on Alfred's proposal
   - Or implement directly

4. **Check analytics** (1 week later)
   - Confirm effectiveness at SessionStart
   - Analyze monthly report

---

## 📊 Expected Impact

**Expected effectiveness through SPEC-First workflow**:

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Implementation time | 60 min | 45 min | 25% ↓ |
| Bug occurrence | 8 | 6 | 25% ↓ |
| Test coverage | 80% | 90% | 10% ↑ |
| Code review time | 20 min | 12 min | 40% ↓ |
| SPEC writing time | 30 min | 9 min | 70% ↓ |

---

## 🔗 Related Resources

| Document | Purpose |
|----------|---------|
| **CLAUDE.md** | Complete Alfred and MoAI-ADK structure (includes only overview of this Skill) |
| **alfred-decision-logic.md** | Alfred's decision algorithm detailed |
| **templates.md** | 3-level SPEC templates complete definition |
| **analytics.md** | Analytics and reporting system design |
| **examples.md** | 10+ real-world use cases |
| **FAQ.md** | Frequently asked questions and answers |

---

**Skill Version**: 1.0.0
**Last Updated**: 2025-11-21
**Status**: Active - In Use
