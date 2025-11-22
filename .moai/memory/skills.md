# MoAI-ADK Skills Reference

## Overview

This document defines the 135 MoAI-ADK skills organized by category and usage patterns. Skills provide domain-specific knowledge and Context7 integration for latest API documentation.

## Skill Categories

### Foundation Skills (Core)

#### `moai-foundation-ears`
- **Purpose**: EARS (Easy Approach to Requirements Syntax) specification format
- **Use Cases**: Requirement writing, specification formatting
- **Integration**: Used by spec-builder for structured requirements

#### `moai-foundation-specs`
- **Purpose**: Specification management and lifecycle
- **Use Cases**: SPEC creation, maintenance, versioning
- **Integration**: Works with spec-builder and docs-manager

#### `moai-foundation-trust`
- **Purpose**: TRUST 5 quality principles implementation
- **Use Cases**: Quality gate enforcement, validation criteria
- **Integration**: Used by quality-gate and all implementation agents

### Language Skills (moai-lang-*)

#### Python Ecosystem
- `moai-lang-python`: Python development patterns and best practices
- `moai-lang-django`: Django framework specialization
- `moai-lang-fastapi`: FastAPI web framework
- `moai-lang-pandas`: Data manipulation and analysis
- `moai-lang-numpy`: Numerical computing
- `moai-lang-pytest`: Testing framework and patterns

#### JavaScript/TypeScript Ecosystem
- `moai-lang-typescript`: TypeScript development and type safety
- `moai-lang-react`: React component development
- `moai-lang-next`: Next.js full-stack framework
- `moai-lang-node`: Node.js server development
- `moai-lang-vue`: Vue.js framework
- `moai-lang-angular`: Angular framework

#### Other Languages
- `moai-lang-go`: Go language patterns and concurrency
- `moai-lang-rust`: Rust systems programming
- `moai-lang-java`: Java enterprise development
- `moai-lang-csharp`: C# and .NET development
- `moai-lang-swift`: Swift and iOS development

### Domain Skills (moai-domain-*)

#### Backend Development
- `moai-domain-backend`: Backend architecture patterns
- `moai-domain-api`: REST and GraphQL API design
- `moai-domain-microservices`: Microservices architecture
- `moai-domain-database`: Database design and optimization
- `moai-domain-authentication`: Authentication and authorization
- `moai-domain-queueing`: Message queuing systems

#### Frontend Development
- `moai-domain-frontend`: Frontend architecture patterns
- `moai-domain-state-management`: State management strategies
- `moai-domain-styling`: CSS and styling systems
- `moai-domain-testing`: Frontend testing strategies
- `moai-domain-performance`: Frontend optimization
- `moai-domain-accessibility`: Web accessibility implementation

#### DevOps & Infrastructure
- `moai-domain-devops`: DevOps practices and CI/CD
- `moai-domain-cloud`: Cloud platform integration
- `moai-domain-containerization`: Docker and Kubernetes
- `moai-domain-monitoring`: System monitoring and observability
- `moai-domain-security`: Security implementation patterns
- `moai-domain-networking`: Network configuration and protocols

#### Data & Analytics
- `moai-domain-data-engineering`: Data pipeline development
- `moai-domain-analytics`: Data analysis and insights
- `moai-domain-ml`: Machine learning implementation
- `moai-domain-etl`: ETL processes and data transformation
- `moai-domain-streaming`: Real-time data processing
- `moai-domain-visualization`: Data visualization techniques

#### Design & UX
- `moai-domain-design-systems`: Design system architecture
- `moai-domain-ui-components`: Component library development
- `moai-domain-ux-research`: User experience research methods
- `moai-domain-prototyping**: Prototyping techniques
- `moai-domain-wireframing`: Wireframe creation and layout
- `moai-domain-responsive-design`: Responsive design patterns

### Essential Skills (moai-essentials-*)

#### Code Quality & Review
- `moai-essentials-review`: Code review methodologies
- `moai-essentials-refactor`: Refactoring techniques
- `moai-essentials-testing`: Testing strategies and frameworks
- `moai-essentials-debugging`: Debugging methodologies
- `moai-essentials-profiling`: Performance profiling techniques
- `moai-essentials-cleanup`: Code cleanup and optimization

#### Development Practices
- `moai-essentials-agile`: Agile methodologies
- `moai-essentials-git`: Git best practices and workflows
- `moai-essentials-documentation`: Documentation standards
- `moai-essentials-versioning`: Version control and release management
- `moai-essentials-collaboration`: Team collaboration patterns
- `moai-essentials-communication`: Technical communication

#### Architecture Patterns
- `moai-essentials-patterns`: Design patterns implementation
- `moai-essentials-architecture`: Software architecture principles
- `moai-essentials-scalability`: System scalability patterns
- `moai-essentials-reliability`: System reliability and fault tolerance
- `moai-essentials-security`: Security best practices
- `moai-essentials-performance`: Performance optimization strategies

### Core System Skills (moai-core-*)

#### Agent Orchestration
- `moai-core-agent-factory`: Agent creation and configuration
- `moai-core-session-management`: Session handling and state
- `moai-core-context-optimization`: Context management and optimization
- `moai-core-task-delegation`: Task delegation patterns
- `moai-core-error-handling`: Error handling and recovery
- `moai-core-logging`: System logging and monitoring

#### Configuration Management
- `moai-core-config-schema`: Configuration schema validation
- `moai-core-environment`: Environment variable management
- `moai-core-settings`: Settings management and validation
- `moai-core-secrets`: Secrets management and security
- `moai-core-hooks`: System hooks and event handling
- `moai-core-permissions`: Permission and access control

#### Quality & Validation
- `moai-core-validation`: Input validation and sanitization
- `moai-core-quality-gates`: Quality gate implementation
- `moai-core-compliance`: Compliance checking and reporting
- `moai-core-auditing`: System auditing and tracking
- `moai-core-monitoring`: System health monitoring
- `moai-core-reporting`: Report generation and analysis

### Specialized Integration Skills

#### MCP Integration
- `mcp-context7-integration`: Context7 MCP server integration
- `mcp-playwright-integration`: Playwright browser automation
- `mcp-figma-integration`: Figma design system integration

#### Platform Integration
- `aws-integration`: Amazon Web Services integration
- `azure-integration`: Microsoft Azure integration
- `gcp-integration`: Google Cloud Platform integration
- `github-integration`: GitHub API integration
- `slack-integration`: Slack workspace integration
- `notion-integration`: Notion workspace integration

#### Tool Integration
- `docker-integration`: Docker containerization
- `kubernetes-integration`: Kubernetes orchestration
- `jenkins-integration`: Jenkins CI/CD integration
- `terraform-integration`: Infrastructure as code
- `ansible-integration`: Configuration management

## Skill Usage Patterns

### Single Skill Invocation
```python
# Load language-specific skill
Skill("moai-lang-python")
# Context: Python development task
# Provides: Python best practices, patterns, latest version info
```

### Multi-Skill Combination
```python
# Combine domain and language skills
Skill("moai-domain-backend") + Skill("moai-lang-python")
# Context: Backend API development in Python
# Provides: Architecture patterns + Python implementation
```

### Context7 Integration
```python
# Skill with Context7 for latest API documentation
Skill("moai-lang-react") + Context7("React", "19.0.0")
# Provides: Latest React patterns and API reference
```

### Foundation + Domain Combination
```python
# Foundation skill with domain specialization
Skill("moai-foundation-trust") + Skill("moai-domain-backend")
# Context: Quality backend implementation
# Provides: TRUST 5 principles + backend patterns
```

## Skill Discovery for Agents

### Finding the Right Skill

When delegated a task, agents can discover relevant skills using several strategies:

#### 1. By Category

**Skill naming conventions indicate category**:
- **moai-foundation-*** - Foundation patterns (TRUST, EARS, specs, git)
- **moai-core-*** - Core capabilities (agent-factory, workflow, config)
- **moai-domain-*** - Domain expertise (backend, frontend, database, security)
- **moai-lang-*** - Programming languages (Python, TypeScript, Go, SQL)
- **moai-essentials-*** - Essential utilities (debug, perf, review, refactor)
- **moai-security-*** - Security patterns (OWASP, auth, encryption, threat)
- **moai-docs-*** - Documentation tools (generation, validation, unified)
- **moai-cc-*** - Claude Code system (configuration, skills, agents, commands)

**Example Discovery**:
```
Task: Implement authentication API

Agent discovers by category:
- moai-domain-* → moai-domain-backend (backend patterns)
- moai-security-* → moai-security-api, moai-security-auth (security patterns)
- moai-lang-* → moai-lang-python (implementation language)
- moai-essentials-* → moai-essentials-perf (optimization)
```

#### 2. By Domain

**Domain-specific skill selection**:

**Backend Development**:
- Primary: `moai-domain-backend`, `moai-lang-python/typescript`
- Security: `moai-security-api`, `moai-security-auth`
- Database: `moai-domain-database`, `moai-lang-sql`
- Quality: `moai-essentials-perf`, `moai-essentials-debug`

**Frontend Development**:
- Primary: `moai-domain-frontend`, `moai-lang-typescript/javascript`
- Design: `moai-design-systems`, `moai-lib-shadcn-ui`
- Styling: `moai-lang-tailwind-css`, `moai-icons-vector`
- Performance: `moai-essentials-perf`, `moai-streaming-ui`

**Database Work**:
- Primary: `moai-domain-database`, `moai-lang-sql`
- Backend: `moai-domain-backend`, `moai-lang-python`
- Cloud: `moai-baas-supabase-ext`, `moai-baas-neon-ext`
- Security: `moai-security-encryption`

**DevOps & Infrastructure**:
- Primary: `moai-domain-devops`, `moai-domain-cloud`
- Advanced: `moai-cloud-aws-advanced`, `moai-cloud-gcp-advanced`
- Monitoring: `moai-domain-monitoring`, `moai-essentials-perf`
- Security: `moai-security-secrets`

**Security Work**:
- Primary: `moai-domain-security`, `moai-security-owasp`
- Specific: `moai-security-api`, `moai-security-auth`, `moai-security-encryption`
- Threat: `moai-security-threat`, `moai-security-identity`
- Compliance: `moai-security-compliance`, `moai-security-zero-trust`

#### 3. By Quality Level

**Testing & Quality**:
- Foundation: `moai-foundation-trust` (TRUST 5 principles)
- Testing: `moai-domain-testing`, `moai-essentials-review`
- TDD: `moai-core-dev-guide`, `moai-essentials-debug`
- Review: `moai-core-code-reviewer`, `moai-essentials-refactor`

**Performance Optimization**:
- Primary: `moai-essentials-perf`
- Monitoring: `moai-domain-monitoring`, `moai-essentials-debug`
- Domain-specific: `moai-domain-backend`, `moai-domain-frontend`, `moai-domain-database`

**Code Quality**:
- Review: `moai-essentials-review`, `moai-core-code-reviewer`
- Refactor: `moai-essentials-refactor`
- Standards: `moai-core-practices`, `moai-foundation-trust`

### Skill Loading Examples by Agent

**backend-expert** (Backend Implementation):
```python
# Pre-configured primary skills
skills = [
    "moai-domain-backend",         # Core backend patterns
    "moai-domain-database",        # Database integration
    "moai-security-api",           # API security
    "moai-security-auth",          # Authentication
    "moai-lang-python",            # Python implementation
    "moai-essentials-perf"         # Performance optimization
]

# Conditional skills (load based on project)
if using_typescript:
    skills.append("moai-lang-typescript")
if needs_cloud:
    skills.append("moai-domain-cloud")
```

**frontend-expert** (Frontend Implementation):
```python
# Pre-configured primary skills
skills = [
    "moai-domain-frontend",        # Frontend architecture
    "moai-design-systems",         # Design system integration
    "moai-lib-shadcn-ui",         # Component library
    "moai-lang-typescript",        # TypeScript
    "moai-lang-tailwind-css",      # Styling
    "moai-streaming-ui"            # Real-time UI
]

# Enhancement skills
skills.extend([
    "moai-essentials-perf",        # Frontend optimization
    "moai-icons-vector"            # Icon system
])
```

**security-expert** (Security Validation):
```python
# Pre-configured security skills
skills = [
    "moai-domain-security",        # Security patterns
    "moai-security-owasp",         # OWASP Top 10
    "moai-security-api",           # API security
    "moai-security-auth",          # Authentication
    "moai-security-encryption",    # Encryption
    "moai-security-threat",        # Threat modeling
    "moai-security-compliance"     # Compliance frameworks
]
```

**quality-gate** (Quality Validation):
```python
# Pre-configured quality skills
skills = [
    "moai-foundation-trust",       # TRUST 5 framework
    "moai-essentials-review",      # Code review
    "moai-core-code-reviewer",     # Review orchestration
    "moai-domain-testing",         # Testing strategies
    "moai-essentials-debug",       # Debugging
    "moai-essentials-perf",        # Performance
    "moai-domain-security"         # Security validation
]
```

### Advanced Skill Discovery

**By Task Complexity**:
- **Simple tasks**: 3-5 skills (foundation + domain + language)
- **Medium tasks**: 5-8 skills (+ essentials + security)
- **Complex tasks**: 8-12 skills (+ specialized + Context7)

**By Integration Needs**:
- **MCP Integration**: `moai-mcp-integration` + specific MCP skill
- **Context7 Docs**: `moai-context7-integration` + `moai-context7-lang-integration`
- **BaaS Platforms**: `moai-baas-[platform]-ext` (supabase, vercel, clerk, etc.)

### Skill Discovery Decision Tree

```
Task Analysis
    ↓
What is the primary domain?
    ├─ Backend → moai-domain-backend, moai-lang-python/typescript
    ├─ Frontend → moai-domain-frontend, moai-lang-typescript, moai-design-systems
    ├─ Database → moai-domain-database, moai-lang-sql
    ├─ Security → moai-domain-security, moai-security-*
    └─ Quality → moai-foundation-trust, moai-essentials-review
        ↓
What are the quality requirements?
    ├─ High → Add moai-foundation-trust, moai-domain-testing
    ├─ Security-critical → Add moai-security-*, moai-security-owasp
    └─ Performance-critical → Add moai-essentials-perf, moai-domain-monitoring
        ↓
What implementation languages?
    ├─ Python → moai-lang-python
    ├─ TypeScript → moai-lang-typescript
    ├─ SQL → moai-lang-sql
    └─ Multiple → Load all needed language skills
        ↓
What integrations needed?
    ├─ MCP → moai-mcp-integration + specific MCP skill
    ├─ Context7 → moai-context7-integration
    └─ BaaS → moai-baas-[platform]-ext
        ↓
Load Skills and Execute Task
```

---

## Skill Selection Guidelines

### By Development Phase

#### Planning Phase
- `moai-foundation-ears`: Requirement specification
- `moai-foundation-specs`: Specification management
- `moai-essentials-agile`: Agile planning

#### Design Phase
- `moai-domain-architecture`: System design
- `moai-domain-api`: API design
- `moai-domain-database`: Data modeling
- `moai-essentials-patterns`: Design patterns

#### Implementation Phase
- `moai-lang-{language}`: Language-specific patterns
- `moai-domain-{domain}`: Domain-specific implementation
- `moai-essentials-refactor`: Code quality
- `moai-core-validation`: Input validation

#### Testing Phase
- `moai-essentials-testing`: Testing strategies
- `moai-domain-testing`: Domain-specific testing
- `mcp-playwright-integration`: E2E testing

#### Deployment Phase
- `moai-domain-devops`: Deployment strategies
- `moai-domain-cloud`: Cloud deployment
- `moai-essentials-versioning`: Release management

### By Technology Stack

#### Web Development (MERN)
- `moai-lang-typescript` + `moai-lang-react` + `moai-lang-node`
- `moai-domain-api` + `moai-domain-database`
- `moai-domain-frontend` + `moai-domain-backend`

#### Data Science
- `moai-lang-python` + `moai-lang-pandas` + `moai-lang-numpy`
- `moai-domain-analytics` + `moai-domain-ml`
- `moai-domain-data-engineering`

#### Mobile Development
- `moai-lang-swift` (iOS) or `moai-lang-java` (Android)
- `moai-domain-mobile`: Mobile development patterns
- `moai-domain-api`: Backend integration

#### Enterprise Applications
- `moai-lang-java` + `moai-lang-spring`
- `moai-domain-microservices` + `moai-domain-authentication`
- `moai-domain-security` + `moai-essentials-compliance`

## Skill Maintenance

### Version Management
- Track skill versions and compatibility
- Update Context7 integration for latest APIs
- Maintain backward compatibility where possible

### Quality Assurance
- Regular skill validation and testing
- Performance optimization for skill loading
- Documentation updates for new features

### Community Contributions
- Skill development guidelines
- Contribution review process
- Skill library expansion

## Best Practices

1. **Load relevant skills**: Use domain-specific skills for better results
2. **Combine appropriately**: Match language and domain skills
3. **Use Context7**: Always integrate with latest API documentation
4. **Validate results**: Use foundation skills for quality assurance
5. **Monitor performance**: Track skill loading and execution performance
6. **Update regularly**: Keep skills current with latest technologies

## Skill Integration Examples

### API Development
```python
# Complete API development stack
skills = [
    "moai-foundation-ears",      # Requirements
    "moai-domain-api",           # API design
    "moai-lang-typescript",      # Implementation
    "moai-essentials-testing",    # Testing
    "moai-domain-documentation"  # Documentation
]
```

### Full-Stack Web Application
```python
# Full-stack development
skills = [
    "moai-domain-frontend",       # Frontend architecture
    "moai-lang-react",           # React implementation
    "moai-domain-backend",        # Backend architecture
    "moai-lang-python",          # Python backend
    "moai-domain-database",       # Database design
    "moai-essentials-security"     # Security
]
```

### Data Pipeline
```python
# Data engineering pipeline
skills = [
    "moai-domain-data-engineering",  # Pipeline architecture
    "moai-lang-python",              # Implementation
    "moai-domain-cloud",              # Cloud deployment
    "moai-domain-monitoring",         # Pipeline monitoring
    "moai-essentials-reliability"     # Reliability
]
```

## Context7 Integration Guide

### Library Resolution
```python
# Resolve library for latest documentation
library_id = Context7.resolve("React", "19.0.0")
documentation = Context7.get_docs(library_id)
```

### Skill + Context7 Pattern
```python
# Combine skill knowledge with latest API docs
skill = Skill("moai-lang-react")
api_docs = Context7.get_docs("/facebook/react/19.0.0")
# Provides: Best practices + latest API reference
```

### Multi-Library Integration
```python
# Multiple libraries with Context7
libraries = [
    Context7.get_docs("/facebook/react/19.0.0"),
    Context7.get_docs("/nodejs/node/v20.0.0"),
    Context7.get_docs("/expressjs/express/4.18.0")
]
```