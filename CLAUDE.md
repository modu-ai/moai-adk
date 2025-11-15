# MoAI-ADK

**SPEC-First TDD Development with Alfred SuperAgent - Claude Code v4.0 Integration**

> **Document Language**: Korean > **Project Owner**: GoosLab > **Config**: `.moai/config/config.json` > **Version**: 0.25.6 (from .moai/config.json)
> **Current Conversation Language**: Korean (conversation_language: "ko")
> **Claude Code Compatibility**: Latest v4.0+ Features Integrated

**🌐 Check My Conversation Language**: `cat .moai/config.json | jq '.language.conversation_language'`

---

## 📚 Documentation Index

This documentation is split into modular files for better maintainability:

### Quick Start
- **[Getting Started](.moai/learning/01-quick-start.md)** - 5-minute SPEC-First + TRUST 5 workflow
- **[MoAI Workflow](.moai/learning/02-moai-workflow.md)** - Step-by-step commands and phases

### Core Philosophy
- **[SPEC-First Philosophy](.moai/learning/03-spec-first-philosophy.md)** - Why requirements-first prevents 80% of bugs
- **[TRUST 5 Principles](.moai/learning/04-trust-5-principles.md)** - Automated quality enforcement

### Alfred SuperAgent
- **[Alfred Workflow Protocol](.moai/learning/05-alfred-workflow-protocol.md)** - 5-phase intelligent execution
- **[How Alfred Thinks](.moai/learning/06-how-alfred-thinks.md)** - Senior developer intelligence model
- **[Persona System](.moai/learning/07-persona-system.md)** - 5 personas for different learning styles

### Advanced Topics
- **[Agent Delegation](.moai/learning/08-agent-delegation.md)** - Task() and parallel execution patterns
- **[MCP Integration](.moai/learning/09-mcp-integration.md)** - External service orchestration
- **[Claude Code v4.0](.moai/learning/10-claude-code-v4.md)** - Plan Mode, Explore, MCP setup

### Configuration
- **[Language Architecture](.moai/learning/11-language-architecture.md)** - Multi-language support
- **[Settings Configuration](.moai/learning/12-settings-configuration.md)** - Claude Code settings

---

## 🚀 Quick Start (First 5 Minutes)

### What You'll Accomplish

In just 5 minutes, you'll:
1. ✅ Create a clear SPEC (requirements with traceability)
2. ✅ Implement with TDD (tests-first, production-ready)
3. ✅ Auto-generate documentation (zero manual docs)
4. ✅ Validate TRUST 5 quality (automated checks)

**Result**: Fully functional, tested, documented, production-ready feature.

### The 3-Step Workflow

```
Step 1: /alfred:1-plan "feature description"
   → SPEC-XXX created with EARS format requirements

Step 2: /alfred:2-run SPEC-XXX
   → Red-Green-Refactor cycle with TRUST 5 validation

Step 3: /alfred:3-sync auto SPEC-XXX
   → Documentation auto-generated from code
```

### Why SPEC-First + TRUST 5?

| Traditional | SPEC-First + TRUST 5 |
|------------|-------------------|
| Vague requirements | Crystal clear EARS format SPEC |
| Code-first (guessing) | SPEC-first (certainty) |
| Tests afterward | Tests before code |
| Bugs in production | Zero bugs with TRUST 5 validation |
| Manual documentation | Auto-generated from code |
| Code reviews (3-5 hours) | Automated checks (seconds) |
| **Timeline**: 2+ weeks | **Timeline**: 3-5 days |

---

## 🎩 Alfred SuperAgent Personas

Alfred adapts to **5 different personas** based on your needs:

1. **🎩 Alfred** - Step-by-step guidance (starting new project)
2. **🧙 Yoda** - Deep learning + documentation generation
3. **🤖 R2-D2** - Fast tactical support (production issues)
4. **🤖 R2-D2 Partner** - Pair programming + code review
5. **🧑‍🏫 Keating** - Personalized tutoring (skill mastery)

### How to Use

**Method 1: Natural Language**
```
"Yoda, explain SPEC-First philosophy"
"R2-D2, quick help with this bug"
"Keating, teach me TDD from fundamentals"
```

**Method 2: Commands**
```
/alfred:0-project    # Alfred persona (beginner-friendly)
/alfred:1-plan       # Plan mode with deep analysis
/alfred:2-run        # Implementation with agents
/alfred:3-sync       # Documentation sync
```

---

## 🛡️ TRUST 5 Quality Model

Every feature automatically validates **5 quality principles**:

| Principle | Meaning | Enforcement |
|-----------|---------|------------|
| **T**est-first | No code without tests | TDD mandatory |
| **R**eadable | Clear, maintainable code | Linting + formatting |
| **U**nified | Consistent patterns & style | Style guides |
| **S**ecured | Security-first approach | OWASP checks |
| **T**rackable | Full requirements traceability | SPEC linking |

**Result**: Production-ready code from day 1, zero manual code review.

---

## 📋 Key SPEC-First Concepts

### EARS Format (Easy Approach to Requirements Syntax)

All SPECs use **5 EARS patterns**:

```
Ubiquitous:    The system SHALL [always]
Event-Driven:  WHEN [trigger], The system SHALL [action]
Unwanted:      IF [bad], THEN [prevent]
State-Driven:  WHILE [state], The system SHALL [maintain]
Optional:      WHERE [user choice], The system SHALL [feature]
```

**Example SPEC-LOGIN-001**:
```
Ubiquitous:
> The system SHALL hash passwords using bcrypt with 10+ rounds
> The system SHALL validate email format before submission

Event-Driven:
> WHEN user submits valid email/password
> The system SHALL authenticate and create session

Unwanted Behavior:
> IF credentials invalid
> THEN the system SHALL reject and log attempt
> The system SHALL lock account after 3 failures

Optional:
> WHERE user enables "remember me"
> The system SHALL set persistent cookie for 30 days
```

---

## 🔄 Typical Project Timeline

```
Day 1: Planning
  /alfred:1-plan "user authentication feature"
  → SPEC-AUTH-001 created (1 hour)

Day 2-3: Development
  /alfred:2-run SPEC-AUTH-001
  → Red phase: 10 tests written
  → Green phase: Implementation
  → Refactor: Code quality improvement
  → TRUST 5 validation passes ✅

Day 4: Documentation & Deployment
  /alfred:3-sync auto SPEC-AUTH-001
  → Docs auto-generated from code
  → Ready for production
```

**Total: 3-4 days vs 2 weeks traditional = 75% faster**

---

## 🎯 Key Features of MoAI-ADK

1. **SPEC-First**: Requirements before code (prevents 80% of bugs)
2. **TDD Enforced**: Red-Green-Refactor with 85%+ coverage requirement
3. **Automated Quality**: TRUST 5 validation (no manual code review)
4. **19 Specialized Agents**: Parallel execution for speed
5. **Living Documentation**: Auto-generated, always in sync
6. **Full Traceability**: SPEC → Code → Tests → Docs linked
7. **Production-Ready Day 1**: No surprises, no surprises

---

## 🚀 Next Steps

### Want to Learn More?

- 📖 **[SPEC-First Philosophy](.moai/learning/03-spec-first-philosophy.md)** - Deep dive into requirements-first
- 🛡️ **[TRUST 5 Principles](.moai/learning/04-trust-5-principles.md)** - Quality enforcement model
- 🧠 **[How Alfred Thinks](.moai/learning/06-how-alfred-thinks.md)** - Intelligence model & reasoning

### Ready to Start?

```bash
# Initialize your project
/alfred:0-project

# Create your first SPEC
/alfred:1-plan "Your feature here"

# Implement with TDD
/alfred:2-run SPEC-001

# Generate documentation
/alfred:3-sync auto SPEC-001
```

### Need Help?

- **Learning**: "Yoda, explain [topic]" (generates .moai/learning/ docs)
- **Production Issue**: "R2-D2, [urgent problem]" (fast tactical help)
- **Pair Programming**: "R2-D2 Partner, let's [task]" (collaborative coding)
- **Skill Mastery**: "Keating, teach me [skill]" (personalized tutoring)

---

## 📁 Project Structure

```
.
├── CLAUDE.md                    # This file (quick reference)
├── .moai/
│   ├── config/
│   │   └── config.json         # Project configuration
│   ├── specs/                  # SPEC documents (SPEC-XXX.md)
│   ├── reports/                # Generated reports
│   ├── learning/               # Detailed learning materials
│   │   ├── 01-quick-start.md
│   │   ├── 03-spec-first.md
│   │   ├── 04-trust-5.md
│   │   └── ...
│   └── ...
├── .claude/
│   ├── agents/                 # Agent definitions
│   ├── skills/                 # Skill implementations
│   └── hooks/                  # Claude Code hooks
└── src/                        # Your codebase
```

---

## 📞 Support & Community

- **Issues**: GitHub Issues (with SPEC reference)
- **Discussions**: GitHub Discussions
- **Documentation**: .moai/learning/ directory
- **Examples**: .moai/examples/ directory

---

**Last Updated**: 2025-11-16
**Version**: 0.25.6
**Claude Code**: v4.0+ ready
**Status**: Production-ready
