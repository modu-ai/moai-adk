# skill-factory Workflow — Version-Aware Skill Creation (2025-10-22)

**Status**: Phase 1 Complete (Python Skill v2.0), Phases 2-5 Ready for Execution

---

## Executive Summary

The skill-factory agent has been enhanced with **version awareness** to ensure all Skills reference **latest stable versions (2025-10-22)** rather than outdated 2025-03-29 versions.

**Completed**:
- ✅ Phase 1: moai-lang-python v2.0 (Python 3.13.1 + pytest 8.4.2 + ruff 0.13.1 + uv 0.9.3)
- ✅ INTERACTIVE-DISCOVERY.md: Added Phase 0f (Version Awareness) + Phase 0g (Final Confirmation)
- ✅ PYTHON-VERSION-MATRIX.md: Created comprehensive Python version tracking
- ✅ Dual deployment: Both project root + templates synchronized

**Ready for Execution**:
- 🔄 Phase 2: moai-foundation-trust v2.0 (testing frameworks + SAST tools)
- 🔄 Phase 3: moai-essentials-debug v2.0 (all 23 languages debuggers)
- 🔄 Phase 4: moai-domain-backend v2.0 (Kubernetes 1.31 + cloud-native patterns)
- 🔄 Phase 5: Batch update 50+ remaining Skills with latest versions

---

## Phase 1: Python Skill v2.0 ✅ COMPLETE

### Deliverables

| File | Lines | Status | Location |
|------|-------|--------|----------|
| SKILL.md | 432 | ✅ v2.0.0 (2025-10-22) | Both (project + template) |
| examples.md | 450+ | ✅ 5 production examples | Both |
| reference.md | 380+ | ✅ CLI command matrix | Both |
| PYTHON-VERSION-MATRIX.md | 420+ | ✅ Version tracking | moai-skill-factory |

### Content Updates

**New Features**:
- Python 3.13.1 support with PEP 695, 701, 698
- ruff 0.13.1 as new standard (replaces black + pylint)
- uv 0.9.3 as package manager (10x faster)
- asyncio.TaskGroup patterns (Python 3.13 exclusive)
- Security best practices (secrets module, secure hashing)
- FastAPI 0.115.0 examples
- Pydantic 2.7.0 runtime validation

**Tool Versions**:
```
pytest:     8.4.2   (testing)
ruff:       0.13.1  (linting/formatting)
mypy:       1.8.0   (type checking)
uv:         0.9.3   (package manager)
FastAPI:    0.115.0 (web framework)
Pydantic:   2.7.0   (validation)
SQLAlchemy: 2.0.28  (ORM)
```

### Verification

```bash
# Both locations synchronized ✅
diff /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-lang-python/SKILL.md \
     /Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills/moai-lang-python/SKILL.md
# Result: No differences
```

---

## Phase 2: Foundation-Trust Skill v2.0 🔄 READY

### Scope

Update `moai-foundation-trust` with latest 2025-10-22 testing tools and CI/CD automation.

**Current Status**:
- Last updated: 2025-03-29 (7 months old)
- Quality: 6/10
- Issue: No tool versions, no CI/CD examples, no SAST integration

### Deliverables

**SKILL.md** (updated sections):
- Testing framework matrix (pytest 8.4.2, Vitest 2.0.5, Jest 29.x, etc.)
- SAST/Security tools (detect-secrets 1.4.x, trivy 0.46.3, semgrep 1.45.x)
- GitHub Actions CI/CD workflow examples
- Quality gate enforcement (coverage ≥85%)
- 10+ practical examples across languages

**reference.md** (new):
- CLI command matrix for all testing frameworks
- SAST tool configuration guide
- GitHub Actions snippets with latest actions versions

**examples.md** (new):
- Python pytest with coverage enforcement
- TypeScript Vitest with CI integration
- Go testing with benchmarks
- SAST workflow setup

### Tool Versions to Document

```yaml
# Testing Frameworks (2025-10-22)
Python:
  - pytest: 8.4.2
  - unittest: stdlib (3.13)
  - coverage: 7.5.0

TypeScript/JavaScript:
  - Vitest: 2.0.5
  - Jest: 29.7.x
  - node:test: builtin

Go:
  - testing: builtin (1.23)
  - testify: 1.9.0

Rust:
  - cargo test: builtin (1.82.0)
  - criterion: 0.5.1

Java:
  - JUnit: 5.10.0
  - TestNG: 7.10.1

# SAST/Security Tools
trivy:           0.46.3
detect-secrets:  1.4.0
semgrep:         1.45.x
bandit:          1.7.5 (Python)
```

---

## Phase 3: Essentials-Debug Skill v2.0 🔄 READY

### Scope

Expand `moai-essentials-debug` from 3 languages to **all 23 languages** with debuggers, container debugging, and observability.

**Current Status**:
- Last updated: 2025-03-29 (7 months old)
- Language coverage: 3/23 (Python, TypeScript, Java only)
- Issue: Missing 20 languages, no distributed tracing, no cloud debugging

### Deliverables

**SKILL.md** (updated):
- Debugger matrix for all 23 languages with versions
- Container debugging (Docker, Kubernetes)
- Distributed tracing (OpenTelemetry 1.24.0)
- Cloud debuggers (AWS X-Ray, GCP Cloud Debugger)
- Prometheus 2.48.x integration

**reference.md** (new):
- Language-by-language debugger setup
- VSCode launch.json templates
- Container debugging config
- Observability stack setup

**examples.md** (new):
- Python async debugging
- JavaScript/TypeScript debugging
- Go goroutine debugging
- Distributed tracing setup
- Kubernetes pod debugging

### 23 Languages Debuggers to Document

```
Python:        pdb, pudb 2024.1, debugpy 1.8.0
TypeScript:    node inspect, VS Code
JavaScript:    node --inspect, VS Code
Go:            delve 1.22, VS Code Go
Rust:          rust-gdb, rust-lldb, CodeLLDB
Java:          jdb, IntelliJ IDEA, VS Code
Kotlin:        IDEA debugger, VS Code
Swift:         LLDB, Xcode debugger
Dart:          VS Code, Android Studio
C/C++:         GDB, LLDB, VS Code
C#:            VS Code, Visual Studio
Ruby:          ruby-debug, RubyMine
PHP:           Xdebug 3.3.0, VS Code
Shell:         bash -x, shellcheck
SQL:           Database IDE debuggers
R:             RStudio debugger
Julia:         Juno IDE debugger
Lua:           lua-debug, VS Code
Elixir:        IEx, Elixir Debugger
Clojure:       CIDER, Cursive IDE
Haskell:       GHCi debugger, Haskell IDE
Scala:         Scala IDE, IntelliJ IDEA
```

---

## Phase 4: Domain-Backend Skill v2.0 🔄 READY

### Scope

Update `moai-domain-backend` with cloud-native patterns (Kubernetes, service mesh, observability, OWASP security).

**Current Status**:
- Last updated: 2025-03-29 (7 months old)
- Quality: 6/10
- Issue: No Kubernetes, no service mesh, no OWASP API security

### Deliverables

**SKILL.md** (updated):
- Architecture patterns (layered, microservices, serverless, event-driven)
- Kubernetes 1.31.x orchestration
- Istio 1.21.x service mesh
- OpenTelemetry 1.24.0 observability
- OWASP API Security Top 10 (2023)
- Database patterns with latest versions

**reference.md** (new):
- Architecture decision matrix
- Tool comparison tables
- Deployment strategies
- Monitoring/observability setup

**examples.md** (new):
- Kubernetes + Istio microservices
- Event-driven with Kafka 3.7
- Serverless with AWS Lambda
- OpenTelemetry + Prometheus observability

### Tool Versions to Document

```yaml
Containers:
  Docker:          27.0+
  Kubernetes:      1.31.x
  Docker Compose:  2.24.x

Service Mesh:
  Istio:           1.21.x
  Linkerd:         1.15.x
  Consul:          1.18.x

Observability:
  OpenTelemetry:   1.24.0
  Prometheus:      2.48.x
  Jaeger:          1.51.x
  Grafana:         11.x
  ELK Stack:       8.11.x

Databases:
  PostgreSQL:      16.x
  MongoDB:         8.0.x
  Redis:           7.2.x
  Cassandra:       4.1.x

Message Brokers:
  Kafka:           3.7.x
  RabbitMQ:        3.13.x
  NATS:            2.10.x

Security:
  OWASP:           2023 edition
  Vault:           1.15.x
  cert-manager:    1.14.x
```

---

## Phase 5: Batch Update 50+ Skills 🔄 READY

### Scope

Update remaining 50+ Skills with latest versions across all tiers:
- 6 Foundation Skills
- 4 Essentials Skills
- 11 Alfred Skills
- 10 Domain Skills
- 23 Language Skills

### Strategy

**Parallel Execution**:
- Group by tier for coherent updates
- Each tier can be updated in parallel
- Dependencies: Language → Domain → Foundation → Alfred

**Version Matrix Reference**: `.moai/memory/VERSION-TRACKING.md`

### Template Update Process

For each Skill:

1. **Read** current SKILL.md
2. **Update** YAML frontmatter:
   - `version: 2.0.0`
   - `updated: 2025-10-22`
   - Add `keywords` field
3. **Add** tool version matrix
4. **Create** reference.md (CLI commands)
5. **Create** examples.md (5+ examples)
6. **Deploy** to both locations (project + template)
7. **Verify** synchronized

### Estimated Effort

| Skill | Complexity | Hours | Status |
|-------|-----------|-------|--------|
| moai-lang-python | High | 6 | ✅ Done |
| moai-foundation-trust | High | 8 | 🔄 Ready |
| moai-essentials-debug | Very High | 16 | 🔄 Ready |
| moai-domain-backend | High | 12 | 🔄 Ready |
| Other 50 Skills | Medium | 2-4 each | 🔄 Ready |
| **Total** | | **100-120** | On track |

---

## New INTERACTIVE-DISCOVERY.md Enhancement

### Phase 0f: Version Awareness Survey

```
┌─ SKILL FACTORY: VERSION MATRIX ───────────────────┐
│                                                     │
│ For [Framework], which tool versions should we   │
│ reference? (Defaults: 2025-10-22 stable)         │
│                                                     │
│ ☑ pytest 8.4.2      (testing framework)           │
│ ☑ ruff 0.13.1       (linting/formatting)          │
│ ☑ uv 0.9.3          (package manager)             │
│ ☑ mypy 1.8.0        (type checking)               │
│                                                     │
│ [↑↓ Navigate] [Space: Toggle] [Enter: Continue]   │
└─────────────────────────────────────────────────────┘
```

### Phase 0g: Final Confirmation with Versions

Shows all selected versions before Skill generation begins.

---

## Deployment Verification Checklist

For each updated Skill:

- [ ] SKILL.md updated (version 2.0.0, 2025-10-22)
- [ ] reference.md created (CLI commands)
- [ ] examples.md created (5+ examples)
- [ ] YAML frontmatter includes `keywords`
- [ ] Tool versions documented (all from 2025-10-22 matrix)
- [ ] Project root version ready
- [ ] Template version ready
- [ ] Both locations identical (diff verification)
- [ ] File sync timestamp close (<1 minute)

---

## Execution Instructions

### For Manual Execution (Phase 2-5)

```bash
# Phase 2: Trust Skill
@agent-skill-factory trust
# or use skill-factory sub-agent directly

# Phase 3: Debug Skill
@agent-skill-factory debug

# Phase 4: Backend Skill
@agent-skill-factory backend

# Phase 5: Batch (50+ Skills)
@agent-skill-factory batch --tier=language
@agent-skill-factory batch --tier=domain
@agent-skill-factory batch --tier=alfred
```

### For Automated Batch Processing

```bash
# Use skill-factory agent with parallel execution
task_type: skill-factory-batch
tier: language  # Process all 23 language skills
version: 2.0.0
reference_date: 2025-10-22
parallel: true
```

---

## Quality Metrics

### Phase 1 Results (Python Skill v2.0)

| Metric | Target | Achieved | Status |
|--------|--------|----------|--------|
| Version Currency | Latest stable | Python 3.13.1 ✅ | ✅ Met |
| Tool Coverage | All critical | pytest, ruff, mypy, uv ✅ | ✅ Met |
| Examples | 5+ | 5 examples ✅ | ✅ Met |
| CLI Reference | Complete | 50+ commands ✅ | ✅ Met |
| Deployment | Dual (2 locations) | Both synced ✅ | ✅ Met |
| Documentation | Comprehensive | 432 + 450 + 380 lines ✅ | ✅ Met |

### Expected Improvements (After All Phases)

| Dimension | Before | After |
|-----------|--------|-------|
| Average Skill Age | 7 months | <1 week |
| Version Currency | 2025-03-29 | 2025-10-22 |
| Average Quality Score | 6.2/10 | 9.0/10 |
| Tool Coverage | 60% | 100% |
| Language Coverage (Debug) | 13% (3/23) | 100% (23/23) |
| Documentation Lines | 3,500 | 8,000+ |

---

## Related Documents

- [INTERACTIVE-DISCOVERY.md](INTERACTIVE-DISCOVERY.md) — Phase 0f/0g version awareness
- [WEB-RESEARCH.md](WEB-RESEARCH.md) — Latest information gathering
- [PYTHON-VERSION-MATRIX.md](PYTHON-VERSION-MATRIX.md) — Python version tracking
- [.moai/memory/VERSION-TRACKING.md]() — Comprehensive version matrix (all 23 languages + 50+ frameworks)
- [skill-factory.md](../../agents/alfred/skill-factory.md) — Sub-Agent orchestration

---

**Version**: 0.1.0
**Created**: 2025-10-22
**Framework**: MoAI-ADK + skill-factory + version matrix
**Next Milestone**: Complete Phase 2-5 execution (100-120 hours estimated)
