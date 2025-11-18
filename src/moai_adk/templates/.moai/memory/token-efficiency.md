# Token Efficiency & Context Management

**Advanced strategies for maximizing token efficiency in Claude Code and MoAI-ADK.**

> **See also**: CLAUDE.md → "Token Efficiency with Agent Delegation" for overview

---

## Understanding Token Consumption

### Token Budget Breakdown

Claude Code's 200,000-token context window:

```
Full codebase load:        50,000+ tokens
SPEC documents:            20,000 tokens
Conversation history:      30,000 tokens
Templates/skill guides:    20,000 tokens
─────────────────────────────────────
Already used:             120,000 tokens (60%!)
```

### Critical Insight

A single conversation can consume 120,000+ tokens before you write a single line of code!

---

## Session Initialization & `/clear`

### Why `/clear` is Essential

SPEC 생성 후 **반드시** `/clear`로 세션을 초기화:

```
Before /clear:
├─ SPEC 생성 대화: 40,000 tokens
├─ Compilation errors: 10,000 tokens
└─ Total context: 50,000 tokens (불필요!)

After /clear:
├─ SPEC 문서만 로드: 5,000 tokens
└─ TDD 구현 시작: 깨끗한 상태
```

### Token Savings Calculation

```
❌ Without /clear:
SPEC 작성: 40,000 tokens
구현: 50,000 tokens
총합: 90,000 tokens

✅ With /clear:
SPEC 로드: 5,000 tokens
구현: 40,000 tokens (최적화)
총합: 45,000 tokens (50% 절약!)
```

---

## Agent Delegation for Token Efficiency

### Strategy 1: Break Down Complex Tasks

```
Task: "Build full-stack app" (100K+ tokens)
              ↓
Break into:   ├─ Backend API design (15K)
              ├─ Frontend UI design (15K)
              ├─ Database schema (10K)
              ├─ Security audit (10K)
              └─ Deployment setup (10K)

Total: 50K tokens (50% savings!)
```

### Strategy 2: Model Selection Optimization

```
✅ Use Sonnet 4.5 for:
   - SPEC creation ($0.003/1K tokens)
   - Architecture decisions
   - Complex reasoning

✅ Use Haiku 4.5 for:
   - Codebase exploration ($0.0008/1K tokens)
   - Simple implementations
   - Code reviews

Result: 70% cheaper than all-Sonnet!
```

### Strategy 3: Context Pruning

```
Frontend task:
├─ Load: UI components only
└─ Exclude: Backend, database files

Backend task:
├─ Load: API and database files
└─ Exclude: Frontend, UI files

Savings: Don't load full codebase in every agent!
```

---

## Memory File Optimization

### What is Memory File?

Persistent context shared across sessions:

```bash
# .claude/memory.md (persistent)
- Project architecture overview
- Key patterns and conventions
- Critical deployment procedures
- Security considerations
```

### Using Memory Effectively

```yaml
# Good Memory File (concise)
- [ Backend Tech Stack ]
  - FastAPI 0.115+, Pydantic v2, SQLAlchemy 2.0
  - Pattern: API route → Service → Repository → DB

- [ Key Patterns ]
  - Async/await for I/O
  - Dependency injection in routes
  - Exception handling via middleware

# Bad Memory File (redundant)
- [Everything from README]
- [Full installation instructions]
- [All configuration examples]
```

---

## Context Management Commands

### Monitoring Context

```bash
/context          # Check current usage
/memory           # View persistent memory
/cost             # API usage and costs
/usage            # Plan usage limits
```

### Optimizing Context

```bash
/clear            # Start fresh (remove everything)
/compact          # Compress conversation
/add-dir src/     # Selectively add files
/remove-dir docs/ # Remove directory context
```

---

## Phase-Based Token Planning

### Phase 1: SPEC Creation
```
Token Budget: 50,000
├─ Initial context: 20,000
├─ SPEC writing: 25,000
└─ Feedback loops: 5,000
```

### Phase 2: Implementation (after /clear)
```
Token Budget: 120,000
├─ SPEC document: 5,000
├─ Relevant code: 30,000
├─ Implementation: 60,000
└─ Testing/fixes: 25,000
```

### Phase 3: Documentation (after /clear)
```
Token Budget: 80,000
├─ Code context: 20,000
├─ Doc generation: 40,000
└─ Review/polish: 20,000
```

**Total 3 phases: 250,000 tokens (vs. single session: 400,000+)**

---

## Best Practices Checklist

### ✅ Do's
- ✅ Use `/clear` after SPEC creation
- ✅ Break complex tasks into phases
- ✅ Use agent delegation for focused contexts
- ✅ Monitor `/context` regularly
- ✅ Keep memory file concise

### ❌ Don'ts
- ❌ Load full codebase at start
- ❌ Keep irrelevant conversation history
- ❌ Ignore context warnings
- ❌ Mix SPEC and implementation phases
- ❌ Use all-Sonnet for simple tasks

---

**Last Updated**: 2025-11-18
**Format**: Markdown | **Language**: English
