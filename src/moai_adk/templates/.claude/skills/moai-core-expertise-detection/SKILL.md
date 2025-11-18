---
name: moai-core-expertise-detection
version: 4.0.0
created: 2025-11-11
updated: '2025-11-18'
status: stable
description: Enterprise AI-powered user expertise detection with behavioral analysis,
  communication pattern recognition, code complexity assessment, Context7 integration,
  and adaptive response calibration; activates for personalized guidance generation,
  complexity adjustment, tutorial depth selection, and communication style matching
keywords:
- expertise-detection
- behavioral-analysis
- adaptive-learning
- ai-expertise-modeling
- communication-patterns
- skill-assessment
- personalization
- context7-integration
- user-profiling
- enterprise-adaptation
allowed-tools:
- Read
- AskUserQuestion
- mcp__context7__resolve-library-id
- mcp__context7__get-library-docs
- WebFetch
stability: stable
---


# Enterprise AI-Powered Expertise Detection 

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-core-expertise-detection |
| **Version** | 4.0.0 Enterprise (2025-11-18) |
| **AI Integration** | ✅ Context7 MCP, behavioral analysis, pattern recognition |
| **Auto-load** | Continuous during user interactions |
| **Expertise Levels** | Beginner, Intermediate, Advanced, Expert |
| **Lines of Content** | 920+ with 14+ production examples |
| **Progressive Disclosure** | 3-level (detection, patterns, advanced) |

---

## What It Does

Continuously detects and adapts to user expertise level based on behavioral signals, communication patterns, code examples, and interaction history. Enables Alfred to calibrate complexity, example selection, and communication style dynamically.

---

## Expertise Level Framework

### Level 1: Beginner
**Characteristics**: 
- First interaction with tool/framework
- Asks about basic concepts
- Needs step-by-step guidance
- Values examples and analogies

**Detection Signals**:
- Questions about fundamentals ("What is a hook?")
- Copy-paste behavior (not understanding code)
- Struggles with terminology
- Prefers verbose, explicit examples
- First project in technology

**Adaptation**:
```
✓ Provide high-level analogies
✓ Explain concepts before showing code
✓ Include step-by-step examples
✓ Link to beginner tutorials
✓ Avoid advanced jargon
✓ Provide working examples
```

### Level 2: Intermediate
**Characteristics**:
- Understands fundamentals
- Works with code confidently
- Asks about patterns and best practices
- Wants to improve and learn

**Detection Signals**:
- Questions about patterns ("How should I structure this?")
- Can read and understand code
- Familiar with core concepts
- Interested in optimization
- Has completed several projects

**Adaptation**:
```
✓ Provide architectural guidance
✓ Explain trade-offs and decisions
✓ Show best practices and patterns
✓ Include edge cases
✓ Use technical terminology correctly
✓ Suggest optimizations
```

### Level 3: Advanced
**Characteristics**:
- Deep framework/language knowledge
- Works with complex architectures
- Asks about edge cases and optimization
- Contributes to open source

**Detection Signals**:
- Questions about performance ("How to optimize this?")
- Writes complex code confidently
- Familiar with framework internals
- Interested in implementation details
- Contributing to frameworks/libraries

**Adaptation**:
```
✓ Focus on nuanced details
✓ Discuss implementation trade-offs
✓ Reference RFC documents
✓ Show performance implications
✓ Discuss advanced patterns
✓ Link to source code
```

### Level 4: Expert
**Characteristics**:
- Framework/language maintainer or deep contributor
- Authors best practices and patterns
- Advises others on architecture
- Contributes to language/framework development

**Detection Signals**:
- Questions about specific implementation choices
- Discusses language/framework internals
- Suggests optimizations based on IL/bytecode
- Participates in language design discussions

**Adaptation**:
```
✓ Assume deep knowledge
✓ Focus on bleeding-edge details
✓ Link to implementation source
✓ Discuss language design decisions
✓ No hand-holding required
```

---

## Continuous Detection Signals

### Signal 1: Terminology Usage
```
"How do I use useState?" 
  → Beginner (basic concept)

"What's the best pattern for managing state?"
  → Intermediate (pattern awareness)

"How does React's useState closure capture work?"
  → Advanced (implementation detail)

"Can we optimize useState with useMemo patterns in concurrent rendering?"
  → Expert (deep architectural knowledge)
```

### Signal 2: Code Complexity
```python
# Beginner: Simple, linear logic
def greet(name):
    return f"Hello, {name}!"

# Intermediate: Using patterns
class UserService:
    def __init__(self, repo):
        self.repo = repo
    
    def get_user(self, user_id):
        return self.repo.find(user_id)

# Advanced: Complex architecture
async def get_user_with_cache(user_id, cache, repo):
    try:
        cached = await cache.get(f"user:{user_id}")
        if cached:
            return json.loads(cached)
    except CacheError:
        pass
    
    user = await repo.find(user_id)
    await cache.set(f"user:{user_id}", json.dumps(user))
    return user

# Expert: Framework internals
async def get_user_with_prefetch(user_id, cache, repo, query_planner):
    # Uses query optimization, connection pooling, prefetch logic
    plan = query_planner.optimize(f"SELECT * FROM users WHERE id={user_id}")
    # Custom execution with monitoring and fallback strategies
```

### Signal 3: Question Type Patterns
```
"How do I...?"           → Beginner
"What's the best way to...?" → Intermediate
"Why does X work like Y?" → Advanced
"How does the implementation..." → Expert
```

### Signal 4: Error Patterns
```
Beginner: Syntax errors, missing imports, undefined variables
Intermediate: Logic errors, wrong patterns, performance issues
Advanced: Race conditions, memory leaks, optimization gaps
Expert: Language semantics, compiler optimizations, platform-specific bugs
```

---

## Dynamic Response Calibration

### Content Depth Adjustment

```
Beginner:   Explain → Example → Hands-on (3 steps)
Intermediate: Pattern → Example → Optimization (3 steps)
Advanced:   Theory → Implementation → Edge cases (3 steps)
Expert:     Design decision → Trade-offs → Implications (3 steps)
```

### Example Complexity Matching

```
# Beginner: Simple, complete example
def add(a, b):
    return a + b

# Intermediate: Pattern-based example
class Calculator:
    def add(self, a, b):
        return a + b

# Advanced: Complex but realistic
class MonoidCalculator(Generic[T]):
    def __init__(self, empty: T, combine: Callable[[T, T], T]):
        self.empty = empty
        self.combine = combine
    
    def fold(self, values: List[T]) -> T:
        return reduce(self.combine, values, self.empty)

# Expert: Framework-level implementation
async def compute_with_cache(
    key: str,
    fn: Callable[[], Awaitable[T]],
    cache: CacheLayer,
    options: ComputeOptions
) -> T:
    # Implementation with error handling, observability, etc.
```

### Terminology Usage

```
Beginner:   "function", "loop", "variable"
Intermediate: "closure", "prototype", "event loop"
Advanced:   "hoisting", "temporal dead zone", "thunk"
Expert:     "reification", "lifting", "unfold semantics"
```

---

## Best Practices for Detection

### DO
- ✅ Update assessment based on each interaction
- ✅ Assume growth (someone at intermediate → advanced)
- ✅ Ask clarifying questions if uncertain
- ✅ Provide escape hatches ("Want more detail?")
- ✅ Err toward more detail (better to simplify)
- ✅ Learn from code examples provided
- ✅ Adapt communication style smoothly

### DON'T
- ❌ Lock expertise level (reassess continuously)
- ❌ Assume based on single signal
- ❌ Be patronizing to beginners
- ❌ Be overly technical without context
- ❌ Skip basics for advanced users (sometimes needed)
- ❌ Assume all users in same domain at same level

---

## Related Skills

- `moai-core-personas` (Communication style adaptation)
- `moai-core-practices` (Pattern examples at all levels)

---

**For detailed detection patterns**: [reference.md](reference.md)  
**For real-world examples**: [examples.md](examples.md)  
**Last Updated**: 2025-11-18  
**Status**: Production Ready (Enterprise )
