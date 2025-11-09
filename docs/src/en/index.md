# MoAI-ADK: AI-Powered SPEC-First TDD Development Framework

> **Build reliable, maintainable software with AI assistance.** All artifacts from requirements to documentation are perfectly tracked, automatically tested, and always synchronized.

---

## :bullseye: Problems We Solve

### 6 Key Issues in Traditional AI-Powered Development

| Problem | Impact |
|---------|--------|
| **Ambiguous Requirements** | Developers spend 40% of time clarifying requirements |
| **Insufficient Testing** | Production bugs from untested code |
| **Out-of-Sync Documentation** | Docs don't match implementation |
| **Lost Context** | Repeated explanations across team members |
| **Impossible Impact Analysis** | Can't identify affected code when requirements change |
| **Inconsistent Quality** | Edge cases missed in manual QA |

### MoAI-ADK's Solution

✅ **SPEC-First**: Define clear requirements before coding
✅ **Guaranteed Testing**: Achieve 87%+ test coverage with automated TDD
✅ **Living Documentation**: Auto-synced, never out-of-date
✅ **Continuous Context**: Alfred remembers project history and patterns
✅ **Complete Traceability**: Connect all artifacts with `@TAG` system
✅ **Quality Automation**: Enforce TRUST 5 principles automatically

---

## ⚡ Core Features

### 1. SPEC-First Development
- **EARS Format Specifications**: Structured, unambiguous requirements
- **Pre-Implementation Clarity**: Prevents costly rework
- **Automatic Traceability**: Links requirements → code → tests

### 2. Automated TDD Workflow
- **RED → GREEN → REFACTOR** cycle management
- **Test-First Guarantee**: No code without tests
- **87%+ Coverage**: Achieved through systematic testing

### 3. Alfred SuperAgent
- **19 Specialized AI Agents** (spec-builder, tdd-implementer, doc-syncer, etc.)
- **93 Production-Ready Skills** (cover all development areas)
- **Adaptive Learning**: Learns from project patterns automatically
- **Smart Context Management**: Understands project structure and dependencies

### 4. @TAG System (Complete Traceability)

Traceability system connecting all artifacts:

```
@SPEC:AUTH-001 (Requirements)
    ↓
@TEST:AUTH-001 (Tests)
    ↓
@CODE:AUTH-001:SERVICE (Implementation)
    ↓
@DOC:AUTH-001 (Documentation)
```

### 5. Living Documentation
- **Real-Time Sync**: Code and docs always aligned
- **No Manual Updates**: Automatically generated
- **Multi-Language Support**: Python, TypeScript, Go, Rust, etc.
- **Auto Diagram Generation**: Automatically created from code structure

### 6. Quality Assurance
- **TRUST 5 Principles**: Test-first, Readable, Unified, Secured, Trackable
- **Automated Quality Gates** (linting, type checking, security scans)
- **Pre-Commit Validation**: Prevent violations before commit
- **Comprehensive Reporting**: Actionable metrics

---

## 🚀 Quick Start

### Installation (Recommended: uv tool)

```bash
# Install moai-adk globally using uv tool
uv tool install moai-adk

# Verify installation
moai-adk --version

# Initialize new project
moai-adk init my-awesome-project
cd my-awesome-project
```

### Project Configuration (Required)

After installation, you must configure the project:

```bash
# Initialize project metadata and environment
/alfred:0-project
```

### 5-Minute Quick Start

```bash
# 1. Plan new feature - SPEC auto-generated
/alfred:1-plan "User authentication with JWT token"

# 2. Run TDD - automatically test → implement → refactor
/alfred:2-run SPEC-AUTH-001

# 3. Sync documentation and validate quality
/alfred:3-sync
```

**Result**: Clear requirements → Test-first implementation → Auto-documentation → Quality assurance all complete!

---

## 📊 Project Statistics

| Item | Count |
|------|-------|
| **Test Coverage** | 87%+ |
| **Supported Languages** | 18 (Python, TypeScript, JavaScript, Go, Rust, Java, Kotlin, Swift, Dart, PHP, Ruby, C, C++, C#, Scala, R, SQL, Shell) |
| **AI Agents** | 19 specialist team |
| **Production-Ready Skills** | 93 |
| **Open Source License** | MIT |

---

## 7️⃣ BaaS Platform Ecosystem (v0.21.0+)

MoAI-ADK fully supports **10 modern cloud platforms**:

### Supported BaaS Platforms

| Platform | Features | Use Cases |
|----------|----------|-----------|
| **Supabase** | PostgreSQL + Auth + Edge | Full-stack SaaS |
| **Firebase** | Realtime DB + Auth | Mobile/Web apps |
| **Vercel** | Edge Functions + Postgres | Serverless APIs |
| **Cloudflare** | Workers + KV + D1 | Edge-first |
| **Auth0** | Enterprise auth | B2B SaaS |
| **Convex** | Real-time backend | Collaborative apps |
| **Railway** | All-in-one platform | Monolithic apps |
| **Neon** | Serverless Postgres | Database layer |
| **Clerk** | Modern auth | User management |

### 8 Architecture Patterns

- **Pattern A**: Multi-tenant SaaS (Supabase + Vercel)
- **Pattern B**: Serverless API (Vercel + Neon + Clerk)
- **Pattern C**: Monolithic Backend (Railway)
- **Pattern D**: Real-time Collaboration (Supabase + Firebase)
- **Pattern E**: Mobile Backend (Firebase + Convex)
- **Pattern F**: Real-time Backend (Convex)
- **Pattern G**: Edge Computing (Cloudflare + Vercel)
- **Pattern H**: Enterprise Security (Auth0 + Supabase)

### 10 Production-Ready Skills

Comprehensive guides with **11,500+ words and 60+ code examples**:

- ✅ **Foundation**: BaaS patterns and decision matrix
- ✅ **Firebase**: Security rules and performance optimization
- ✅ **Supabase**: Row-Level Security and production best practices
- ✅ **Vercel**: Deployment and Edge Functions
- ✅ **Cloudflare**: Durable Objects and Workers
- ✅ **Auth0**: Compliance and MAU management
- ✅ **Convex**: Advanced patterns and cost optimization
- ✅ **Railway**: Blue-green deployment and canary releases
- ✅ **Neon**: Database branching and serverless
- ✅ **Clerk**: Multi-tenancy and user management

---

## 🌟 Key Features

### Multi-Language Support
- **4 Language Support**: Korean, English, Japanese, Chinese
- **AI-Powered Translation**: High-quality translations
- **Real-Time Sync**: Always latest docs in all languages

### Supported Technology Stack
- **Frontend**: React, Vue, Angular (TypeScript)
- **Backend**: Node.js, Python, Go, Rust
- **Database**: SQL, NoSQL (MongoDB, PostgreSQL)
- **Deployment**: Docker, Kubernetes, AWS, Vercel

### Team Collaboration
- **Individual Mode**: Free-form local development
- **Team Mode**: Feature branches, auto PR management, auto-merge
- **Real-Time Context**: Perfect documentation sharing across team

---

## <span class="material-icons">library_books</span> Learning Resources

### Official Documentation
- **[Getting Started](getting-started/installation.md)**: Installation and basic setup
- **[Usage Guide](guides/alfred/index.md)**: Complete Alfred workflow guide
- **[API Reference](reference/cli/index.md)**: Commands and skill API
- **[Developer Guide](contributing/index.md)**: Contributing and extending

### Core Guides
- **[SPEC Writing](guides/specs/basics.md)**: SPEC-First methodology
- **[TDD Execution](guides/tdd/red.md)**: RED → GREEN → REFACTOR cycle
- **[TAG System](guides/specs/tags.md)**: Complete traceability management

---

## :sparkles: Community

- **GitHub**: [modu-ai/moai-adk](https://github.com/modu-ai/moai-adk)
- **Issues**: [Bug reports and feature requests](https://github.com/modu-ai/moai-adk/issues)
- **License**: MIT (commercial use allowed)

---

## 🎬 Next Steps

<div align="center">

### Start Now!

[Quick Start Guide](getting-started/installation.md) · [Alfred Workflow](guides/alfred/index.md) · [GitHub Repository](https://github.com/modu-ai/moai-adk)

---

**Experience the power of SPEC-First TDD development with MoAI-ADK!**

</div>
