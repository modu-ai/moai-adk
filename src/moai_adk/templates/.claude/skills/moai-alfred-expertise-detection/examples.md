# Alfred Expertise Detection - Real-World Examples

> **Main Skill**: [SKILL.md](SKILL.md)  
> **API Reference**: [reference.md](reference.md)

---

## Example 1: Beginner Detection (First-Time User)

### User Behavior

```
User: "What is MoAI-ADK?"
User: "How do I start?"
User: "What's a SPEC?"
User: (tries command) "/alfred project"
User: "It didn't work, what's the right syntax?"
```

### Signal Analysis

| Signal Category | Observation | Score |
|-----------------|-------------|-------|
| Command Usage | Incorrect syntax, asks for help | +2 |
| Questions | Fundamental "what is" questions | +2 |
| Error Recovery | Needs syntax guidance | +2 |
| Documentation | Unfamiliar with SPEC concept | +2 |
| Git Workflow | N/A (not yet used) | 0 |

**Total Score**: 8 → **Beginner**

### Alfred Adaptation

**Verbosity**: High (300-word explanations)
**Confirmations**: All operations confirmed
**Suggestions**: 4-5 proactive tips per session
**Role Bias**: Technical Mentor

**Response Example**:
```
MoAI-ADK is an Agentic Development Kit that follows SPEC-first TDD workflow.

Background:
SPECs are structured requirements documents following EARS format. They 
define what to build before writing code. Alfred helps you create SPECs, 
implement them via TDD, and keep documentation synchronized.

To start:
1. Run /alfred:0-project to initialize project
2. Then /alfred:1-plan to create your first SPEC
3. Then /alfred:2-run to implement via TDD

Related Skills to learn:
- Skill("moai-foundation-specs") - SPEC structure
- Skill("moai-foundation-trust") - TRUST 5 principles

Would you like me to guide you through /alfred:0-project step-by-step?
```

---

## Example 2: Intermediate Detection (Regular User)

### User Behavior

```
User: "/alfred:1-plan"
Alfred: "What's the SPEC title?"
User: "User authentication with JWT"
```

### Signal Analysis

| Signal Category | Observation | Score |
|-----------------|-------------|-------|
| Command Usage | Correct command syntax | +1 |
| Questions | Specific clarification ("Can I?") | +1 |
| Error Recovery | Notices mistake, asks fix | +1 |
| Git Workflow | Uses /alfred commands | +1 |

**Total Score**: 5 → **Intermediate**

### Alfred Adaptation

**Verbosity**: Moderate (150-word explanations)
**Confirmations**: Only for medium/high-risk
**Suggestions**: 2-3 optimizations per session
**Role Bias**: Project Manager

**Response Example**:
```

Edit .moai/specs/SPEC-AUTH-001/spec.md and insert:

Then run /alfred:3-sync to update documentation links.


Proceed with manual edit?
  [Yes, I'll edit now] [No, regenerate SPEC]
```

---

## Example 3: Expert Detection (Power User)

### User Behavior

```
User: "/alfred:2-run SPEC-AUTH-001 && git add . && git commit -m 'feat(auth): JWT impl'"
User: (test fails)
User: (fixes code directly without asking)
User: (reruns tests)
User: (creates PR with gh CLI)
```

### Signal Analysis

| Signal Category | Observation | Score |
|-----------------|-------------|-------|
| Command Usage | Advanced command chaining | 0 |
| Questions | Zero questions asked | 0 |
| Error Recovery | Self-debugs, fixes independently | 0 |
| Documentation | No documentation references | 0 |
| Git Workflow | Direct git + gh commands | 0 |

**Total Score**: 0 → **Expert**

### Alfred Adaptation

**Verbosity**: Minimal (50-word updates)
**Confirmations**: Skip all except high-risk
**Suggestions**: 1 advanced tip per session
**Role Bias**: Efficiency Coach

**Response Example**:
```
SPEC-AUTH-001 implemented.

Files: src/auth/jwt.py, tests/test_auth.py
Tests: 12/12 passing
Coverage: 95%
Commit: c9f82a3

PR #142 created.
```

---

## Example 4: Adaptive Learning (Beginner → Intermediate)

### Session Start (Beginner)

```
User: "How do I create a SPEC?" (Question keyword +2)
Alfred: (Verbose explanation with examples)

User: (follows steps, creates SPEC)
Alfred: "Great! Ready for /alfred:2-run?" (Confirmation)
```

**Score**: 8 → Beginner

### Session Middle (Still Beginner, but improving)

```
User: "/alfred:2-run SPEC-PROFILE-001" (Correct syntax +1)
Alfred: (Moderate explanation, less verbose)
```

**Score**: 6 → Beginner (borderline)

### Session End (Intermediate)

```
User: "Fix typo in line 42" (Direct command +1, no question)
User: (self-corrects minor error without asking +1)
```

**Score**: 4 → Intermediate

**Alfred adaptation**: Switches from Technical Mentor to Project Manager mid-session

---

## Example 5: Override Mechanism (Expert User Asks for Explanation)

### User Behavior

```
User: (Detected as Expert from previous commands)
User: "Explain why TRUST 5 requires test-first?"
```

### Signal Analysis

**Baseline**: Expert (score 0 from previous requests)

**Override detected**:
- Keyword: "Explain" → Force Beginner mode for this request

### Alfred Adaptation

**Normal Expert response**: "TRUST 5 = Test First, Readable, Unified, Secured, Trackable."

**Override to Beginner mode**:
```
TRUST 5 principle "Test First" ensures code correctness through TDD.

Why test-first matters:
- Write failing test (RED) → defines expected behavior
- Implement code (GREEN) → satisfies test requirements
- Refactor (REFACTOR) → improve code with safety net

Without test-first:
- Risk: Code written without clear requirements
- Risk: Tests written to match code (not requirements)
- Risk: "Works on my machine" syndrome

Example from your workflow:
  1. Write test: test_jwt_validation_rejects_expired_token()
  2. Run test: FAIL (function not implemented)
  3. Implement: def validate_jwt(token) → checks expiry
  4. Run test: PASS (behavior verified)
  5. Refactor: Optimize without breaking test

This cycle ensures every line of code has a verified purpose.

Related: Skill("moai-foundation-trust") for full TRUST 5 guide
```

**Characteristics**: Expert user received Beginner-level explanation due to "Explain" keyword

---

## Example 6: Signal Conflict Resolution

### User Behavior

```
User: "/alfred:1-plan" (Correct syntax +1)
User: "What's the difference between SPEC and plan?" (Basic question +2)
User: (Creates SPEC successfully) (+1)
```

### Signal Analysis

**Conflicting signals**:
- Command usage: Intermediate (+1)
- Questions: Beginner (+2, +2)
- SPEC creation: Intermediate (+1)

**Total Score**: 8 → **Beginner**

**Resolution**: Beginner dominates (questions outweigh command proficiency)

### Alfred Adaptation

**Behavior**: Technical Mentor mode

**Response Strategy**: Provide educational explanations while leveraging command proficiency

---

## Summary: Expertise Level Indicators

| Observable Behavior | Beginner | Intermediate | Expert |
|---------------------|----------|--------------|--------|
| **First request** | "What is...?" | Correct command | Command chain |
| **Error handling** | "Fix this for me" | "Why did this fail?" | (Fixes silently) |
| **Documentation** | Frequent Skill() | Occasional Skill() | No Skill() |
| **Confirmations** | Needs all | Needs some | Needs none |
| **Learning curve** | Steep | Moderate | Flat |

---

**End of Examples** | 2025-11-02
