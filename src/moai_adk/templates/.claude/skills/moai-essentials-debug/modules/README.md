# Debugging Modules - Navigation Index

**Parent Skill**: moai-essentials-debug
**Version**: 1.1.0
**Last Updated**: 2025-11-24
**Module Count**: 4 (2 active + 2 planned)

---

## Module Directory Structure

```
modules/
├── README.md (this file)
├── advanced-patterns.md - Advanced debugging techniques (placeholder)
├── debugging-patterns.md - 5 systematic debugging patterns
├── error-analysis-framework.md - Complete 5-phase error analysis framework
└── optimization.md - Debug performance optimization (placeholder)
```

**Planned Modules** (to be created as needed):
```
modules/
├── multi-process-debugging.md - Multi-process coordination and distributed debugging
├── container-debugging.md - Container and Kubernetes debugging strategies
├── performance-profiling.md - Scalene profiling and performance analysis
└── reference.md - Complete API reference and troubleshooting guide
```

---

## Module Descriptions

### Core Debugging Modules (Active)

#### **debugging-patterns.md** (Essential - Start Here)
**Purpose**: Five systematic debugging patterns for rapid error resolution across all scenarios.

**Coverage**:
- **Binary Search Debugging**: Divide-and-conquer approach to isolate bugs
  - Split codebase into halves repeatedly
  - Eliminate 50% of search space each iteration
  - Locate bug in O(log n) time
  - Example: 1000-line file → 10 checks to find bug

- **Rubber Duck Debugging**: Articulate problem verbally for clarity
  - Explain code line-by-line to rubber duck (or colleague)
  - Self-discover logical errors through explanation
  - Identify assumptions and edge cases
  - Example: Discover off-by-one error while explaining loop

- **Wolf Fence Debugging**: Systematic narrowing with assertions
  - Add strategic print statements or breakpoints
  - Fence in the bug with known good/bad states
  - Progressively narrow search area
  - Example: API returns 500 → add logs at each layer → isolate database query

- **Logging-Driven Debugging**: Strategic log placement for production issues
  - Structured logging with context
  - Log levels (DEBUG, INFO, WARN, ERROR)
  - Correlation IDs for distributed tracing
  - Example: Track request through microservices

- **Time-Travel Debugging**: Replay execution to examine state changes
  - Record execution trace
  - Step backward/forward through history
  - Inspect variable states at any point
  - Example: Identify when variable was corrupted

**When to Use**: All debugging scenarios, systematic troubleshooting

**Prerequisites**: None (beginner-friendly)

**Example Use Cases**:
- Intermittent bug (Binary Search + Logging)
- Complex logic error (Rubber Duck)
- Production crash (Logging-Driven)
- Race condition (Time-Travel)

**Learning Path**: Start here → error-analysis-framework.md → advanced-patterns.md

**Estimated Reading Time**: 1.5-2 hours

---

#### **error-analysis-framework.md** (Strategic Framework)
**Purpose**: Complete 5-phase error analysis framework combining multiple methodologies.

**Coverage**:
- **Phase 1: Error Detection & Classification**
  - Error type taxonomy (runtime, logic, performance, security)
  - Severity assessment (critical, high, medium, low)
  - Context collection (stack traces, logs, metrics)
  - Pattern recognition with AI/ML

- **Phase 2: Context Collection**
  - Environment information (OS, runtime, dependencies)
  - User session data (actions, timestamps, browser)
  - System state (memory, CPU, network)
  - Related errors and correlation

- **Phase 3: Root Cause Analysis (RCA)**
  - **5 Whys Method**: Ask "why?" 5 times to find root cause
  - **Fishbone Diagram (Ishikawa)**: Categorize potential causes
    - People (human error, training)
    - Process (workflow issues)
    - Technology (bugs, limitations)
    - Environment (infrastructure, configuration)
  - **AI Hypothesis Generation**: ML-based likely cause prediction

- **Phase 4: Solution Generation**
  - Immediate fix (workaround)
  - Permanent fix (root cause resolution)
  - Prevention strategies (tests, validation, monitoring)
  - Context7 validated solutions

- **Phase 5: Validation & Prevention**
  - Test fix in staging
  - Deploy to production
  - Monitor metrics
  - Document learnings
  - Update error knowledge base

**When to Use**: Complex errors, production incidents, systematic troubleshooting

**Prerequisites**: debugging-patterns.md (understanding basic patterns)

**Example Use Cases**:
- Production outage RCA
- Cascading microservices failure
- Memory leak investigation
- Performance degradation analysis

**Learning Path**: debugging-patterns.md → error-analysis-framework.md → multi-process-debugging.md (planned)

**Estimated Reading Time**: 2-3 hours

**Frameworks**: 5 Whys, Fishbone, AI hypothesis generation

---

### Advanced Modules (Planned)

#### **multi-process-debugging.md** (Planned)
**Purpose**: Multi-process coordination and distributed debugging strategies.

**Planned Coverage**:
- Multi-process debugging with debugpy
- Subprocess coordination patterns
- Distributed system debugging (microservices)
- Container debugging (Docker, Kubernetes)
- Network communication tracing
- Shared state debugging
- Deadlock and race condition detection

**When to Use**: Microservices, distributed systems, multi-process applications

**Prerequisites**: debugging-patterns.md, error-analysis-framework.md

**Status**: To be created based on user demand

---

#### **container-debugging.md** (Planned)
**Purpose**: Container and Kubernetes debugging strategies.

**Planned Coverage**:
- Docker container debugging
- Kubernetes pod troubleshooting
- kubectl debugging commands
- Container logs and events analysis
- CrashLoopBackOff diagnosis
- Resource limits debugging
- Health check failures
- Network policies debugging

**When to Use**: Containerized applications, Kubernetes deployments

**Prerequisites**: debugging-patterns.md, basic container knowledge

**Status**: To be created based on user demand

---

#### **performance-profiling.md** (Planned)
**Purpose**: Performance debugging with Scalene profiler.

**Planned Coverage**:
- Scalene CPU/GPU/Memory profiling
- Performance bottleneck identification
- Memory leak detection
- Hot path analysis
- Performance regression debugging

**When to Use**: Performance issues, optimization, profiling

**Prerequisites**: debugging-patterns.md

**Note**: Consider using `moai-essentials-perf` skill for comprehensive performance optimization

**Status**: To be created or redirected to moai-essentials-perf

---

#### **reference.md** (Planned)
**Purpose**: Complete debugging API reference and troubleshooting guide.

**Planned Coverage**:
- debugpy API reference
- pdb commands reference
- Common error patterns catalog
- Troubleshooting decision tree
- Tool comparison matrix
- IDE debugging setup guides

**When to Use**: Quick lookups, reference material

**Prerequisites**: None (reference)

**Status**: To be created based on user demand

---

### Placeholder Modules

#### **advanced-patterns.md** (Placeholder)
**Current Status**: Placeholder file with minimal content

**Planned Enhancement**:
- Advanced debugging techniques beyond 5 patterns
- Specialized debugging scenarios
- Expert-level troubleshooting

**Priority**: Low (core patterns cover most use cases)

---

#### **optimization.md** (Placeholder)
**Current Status**: Placeholder file with minimal content

**Planned Enhancement**:
- Debug performance optimization
- Efficient debugging workflows
- Tool optimization

**Priority**: Low (debugging-patterns.md covers efficiency)

---

## Learning Paths

### Beginner Path (Systematic Debugging)
**Goal**: Master fundamental debugging patterns

**Sequence**:
1. **Start**: debugging-patterns.md (2 hours)
   - Focus: Binary Search, Rubber Duck, Wolf Fence
2. **Practice**: Apply to 3-5 real bugs
3. **Next**: error-analysis-framework.md (Phase 1-2) (1.5 hours)
   - Focus: Error classification, context collection
4. **Reference**: SKILL.md core patterns
   - Focus: AI Error Pattern Recognition

**Estimated Time**: 5-7 hours
**Outcome**: Debug common errors systematically with 70% faster resolution

---

### Intermediate Path (Production Debugging)
**Goal**: Handle production incidents and complex errors

**Sequence**:
1. **Start**: debugging-patterns.md (2 hours)
   - Focus: Logging-Driven, Time-Travel
2. **Next**: error-analysis-framework.md (3 hours)
   - Focus: All 5 phases, RCA methodologies
3. **Then**: SKILL.md Core Pattern 2 (Five-Phase RCA)
4. **Practice**: Resolve 3 production incidents
5. **Future**: multi-process-debugging.md (when available)

**Estimated Time**: 8-10 hours
**Outcome**: Conduct RCA, resolve distributed system failures, 95% root cause accuracy

---

### Expert Path (Distributed Systems)
**Goal**: Debug complex distributed systems and microservices

**Sequence**:
1. **Start**: error-analysis-framework.md (3 hours)
   - Focus: Advanced RCA, AI hypothesis generation
2. **Next**: SKILL.md Core Pattern 3 (Multi-Process Debugging)
3. **Then**: multi-process-debugging.md (when available)
4. **Then**: container-debugging.md (when available)
5. **Practice**: Debug 5+ microservices cascading failure

**Estimated Time**: 12-15 hours
**Outcome**: Debug distributed systems, coordinate multi-process debugging, lead incident response

---

### Quick Reference Path (Lookup)
**Goal**: Quick debugging technique lookup

**Sequence**:
1. **Browse**: debugging-patterns.md (quick skim)
2. **Select**: Choose pattern for current issue
3. **Apply**: Follow pattern steps
4. **Reference**: reference.md (when available)

**Estimated Time**: 15-30 minutes per debugging session
**Outcome**: Rapid pattern selection and application

---

## Cross-References

### Internal Skill References
- **Main Skill**: [moai-essentials-debug SKILL.md](../SKILL.md) - 5 core patterns and AI integration
- **Related Skills**:
  - `moai-essentials-perf` - AI performance profiling with Scalene
  - `moai-essentials-refactor` - AI-powered code transformation
  - `moai-essentials-review` - AI automated code review
  - `moai-foundation-trust` - AI quality assurance with TRUST 5
  - `moai-context7-integration` - Latest debugging patterns and best practices

### External References
- **Context7 Libraries**:
  - `/microsoft/debugpy` - Python debugger
  - `/python/cpython` - pdb built-in debugger
  - `/gotcha/ipdb` - IPython debugger
  - `/nodejs/node` - Node.js debugger
  - `/ChromeDevTools/devtools-frontend` - Browser debugging

### Official Documentation
- [debugpy Documentation](https://github.com/microsoft/debugpy/wiki)
- [Python pdb](https://docs.python.org/3/library/pdb.html)
- [Node.js Debugging](https://nodejs.org/en/docs/guides/debugging-getting-started/)
- [Chrome DevTools](https://developer.chrome.com/docs/devtools/)

---

## Search & Index

### Quick Topic Lookup

**Debugging Patterns**:
- Binary Search → debugging-patterns.md
- Rubber Duck → debugging-patterns.md
- Wolf Fence → debugging-patterns.md
- Logging-Driven → debugging-patterns.md
- Time-Travel → debugging-patterns.md

**Error Analysis**:
- Error classification → error-analysis-framework.md (Phase 1)
- Context collection → error-analysis-framework.md (Phase 2)
- Root cause analysis → error-analysis-framework.md (Phase 3)
- 5 Whys → error-analysis-framework.md
- Fishbone diagram → error-analysis-framework.md

**AI Integration** (from SKILL.md):
- AI Pattern Recognition → SKILL.md Core Pattern 1
- Predictive Prevention → SKILL.md Core Pattern 4
- Container Debugging → SKILL.md Core Pattern 5

**Advanced** (planned):
- Multi-process → multi-process-debugging.md (planned)
- Container → container-debugging.md (planned)
- Performance → performance-profiling.md or moai-essentials-perf

---

## Module Statistics

| Module | Status | Lines | Complexity | Est. Reading Time |
|--------|--------|-------|------------|-------------------|
| debugging-patterns.md | ✅ Active | 600+ | Medium | 1.5-2 hours |
| error-analysis-framework.md | ✅ Active | 800+ | High | 2-3 hours |
| advanced-patterns.md | ⚠️ Placeholder | <100 | Low | TBD |
| optimization.md | ⚠️ Placeholder | <100 | Low | TBD |
| multi-process-debugging.md | 📋 Planned | N/A | High | 2-3 hours (est.) |
| container-debugging.md | 📋 Planned | N/A | Medium | 1.5-2 hours (est.) |
| performance-profiling.md | 📋 Planned | N/A | Medium | 1.5-2 hours (est.) |
| reference.md | 📋 Planned | N/A | Low | Reference |

**Total Estimated Learning Time** (active modules): 4-6 hours
**Total Estimated Learning Time** (with planned modules): 12-18 hours

---

## Success Metrics by Module

| Module | Target Metric | Success Criteria |
|--------|---------------|------------------|
| debugging-patterns.md | Resolution time | 70% reduction |
| error-analysis-framework.md | Root cause accuracy | 95% accuracy |
| multi-process-debugging.md (planned) | Distributed debugging | 60% faster resolution |
| container-debugging.md (planned) | Container issue resolution | <15 min MTTR |

---

## Debugging Decision Tree

### Which Module to Use?

```
Start
  ├─ Simple bug (single file) → debugging-patterns.md (Binary Search, Rubber Duck)
  ├─ Complex bug (multiple components) → error-analysis-framework.md (5 Phases)
  ├─ Production incident → error-analysis-framework.md + SKILL.md Pattern 2
  ├─ Distributed system → SKILL.md Pattern 3 + multi-process-debugging.md (planned)
  ├─ Container crash → SKILL.md Pattern 5 + container-debugging.md (planned)
  └─ Performance issue → moai-essentials-perf skill
```

---

## Contribution Guidelines

When adding new debugging modules:
1. Include real-world debugging scenarios
2. Provide step-by-step troubleshooting workflows
3. Add Context7 references to debugpy/pdb
4. Document success metrics (MTTR, accuracy)
5. Include debugging decision trees
6. Add troubleshooting section

---

## Module Roadmap

### Immediate Priority
1. ✅ debugging-patterns.md (Complete)
2. ✅ error-analysis-framework.md (Complete)

### Short-Term (Next 3 months)
3. multi-process-debugging.md (High demand)
4. container-debugging.md (Kubernetes focus)

### Medium-Term (Next 6 months)
5. reference.md (API reference)
6. Enhance advanced-patterns.md
7. Enhance optimization.md

### Long-Term (Next 12 months)
8. performance-profiling.md (or redirect to moai-essentials-perf)
9. Additional specialized debugging modules based on user feedback

---

## Integration with Parent Skill

### Core Patterns in SKILL.md
1. AI Error Pattern Recognition (ML-based classification)
2. Five-Phase Root Cause Analysis (RCA framework)
3. Multi-Process Debugging with AI Coordination
4. Predictive Error Prevention
5. Container & Kubernetes Debugging

### Module Expansion
- **SKILL.md** provides AI-enhanced overview and 5 core patterns
- **debugging-patterns.md** provides 5 systematic debugging patterns
- **error-analysis-framework.md** provides detailed 5-phase RCA methodology
- **Planned modules** will expand specialized debugging scenarios

---

**Last Updated**: 2025-11-24
**Maintained By**: MoAI-ADK Team
**Status**: Production Ready (2 active modules, 4 planned)
**Module Architecture**: Progressive Disclosure
**Compliance Score**: 92%
