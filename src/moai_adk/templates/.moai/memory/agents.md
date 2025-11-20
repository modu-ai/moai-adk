# MoAI-ADK Agents Reference

## Overview

This document defines 35 specialized MoAI-ADK agents available for delegation. Each agent has specific capabilities and optimal use cases.

## Core Principles

1. **Delegation Priority**: Always delegate to specialized agents first
2. **Context Transfer**: Pass relevant results between agents using context parameter
3. **Specialization**: Each agent handles specific domain tasks
4. **Quality Gates**: All agents follow TRUST 5 principles

## Agent Categories

### Specification & Planning

#### `spec-builder`
- **Purpose**: Create EARS-format specifications from user requirements
- **Use Cases**: New features, system requirements, API specifications
- **Delegation Trigger**: User requests for planning, requirements, specifications
- **Output**: Structured EARS specification document

#### `plan`
- **Purpose**: Decompose complex tasks into executable steps
- **Use Cases**: Multi-step implementations, project planning
- **Delegation Trigger**: Complex user requests requiring breakdown
- **Output**: Step-by-step execution plan

### Implementation

#### `tdd-implementer`
- **Purpose**: Execute Red-Green-Refactor TDD cycle
- **Use Cases**: Feature implementation, code development
- **Delegation Trigger**: "implement X", "build Y", "create Z"
- **Process**:
  1. Write failing tests (Red)
  2. Implement minimum code (Green)
  3. Refactor for quality (Refactor)

#### `backend-expert`
- **Purpose**: Server-side architecture and implementation
- **Use Cases**: API development, database integration, microservices
- **Delegation Trigger**: Backend-related tasks, server development
- **Specialties**: REST APIs, GraphQL, database design

#### `frontend-expert`
- **Purpose**: Client-side development and UI implementation
- **Use Cases**: Web applications, React components, user interfaces
- **Delegation Trigger**: Frontend development, UI implementation
- **Specialties**: React, Vue, Angular, state management

#### `database-expert`
- **Purpose**: Database design, optimization, and migrations
- **Use Cases**: Schema design, query optimization, data modeling
- **Delegation Trigger**: Database-related tasks, performance issues
- **Specialties**: SQL, NoSQL, migrations, performance tuning

### Quality & Security

#### `security-expert`
- **Purpose**: Security analysis, vulnerability assessment
- **Use Cases**: Security audits, penetration testing, secure coding
- **Delegation Trigger**: Security-related concerns, OWASP compliance
- **Standards**: OWASP Top 10, security best practices

#### `quality-gate`
- **Purpose**: Code quality validation and TRUST 5 enforcement
- **Use Cases**: Code reviews, quality checks, compliance validation
- **Delegation Trigger**: Quality assurance, code validation
- **Criteria**: Test coverage, code standards, security compliance

#### `test-engineer`
- **Purpose**: Comprehensive testing strategy and implementation
- **Use Cases**: Unit tests, integration tests, E2E testing
- **Delegation Trigger**: Testing requirements, quality assurance
- **Specialties**: Test automation, coverage analysis

### Architecture & Design

#### `api-designer`
- **Purpose**: API architecture and endpoint design
- **Use Cases**: REST API design, GraphQL schemas, API documentation
- **Delegation Trigger**: API-related design tasks
- **Standards**: REST principles, OpenAPI specifications

#### `component-designer`
- **Purpose**: Reusable component architecture
- **Use Cases**: Design systems, component libraries, UI kits
- **Delegation Trigger**: Component design, system architecture
- **Principles**: Atomic design, reusability patterns

#### `ui-ux-expert`
- **Purpose**: User experience and interface design
- **Use Cases**: UX analysis, interface design, usability testing
- **Delegation Trigger**: UX-related tasks, user experience optimization
- **Standards**: WCAG 2.1, usability principles

### DevOps & Infrastructure

#### `devops-expert`
- **Purpose**: Deployment pipelines, infrastructure management
- **Use Cases**: CI/CD setup, cloud deployment, infrastructure as code
- **Delegation Trigger**: Deployment, infrastructure tasks
- **Specialties**: Docker, Kubernetes, cloud platforms

#### `monitoring-expert`
- **Purpose**: System monitoring, alerting, observability
- **Use Cases**: Monitoring setup, alerting systems, performance monitoring
- **Delegation Trigger**: Monitoring requirements, system observability
- **Specialties**: Metrics, logging, distributed tracing

#### `performance-engineer`
- **Purpose**: Performance optimization and analysis
- **Use Cases**: Performance tuning, bottleneck analysis, optimization
- **Delegation Trigger**: Performance issues, optimization requirements
- **Focus**: Application performance, database optimization

### Data & Integration

#### `migration-expert`
- **Purpose**: Database migrations and data transformations
- **Use Cases**: Schema migrations, data transfers, legacy system migration
- **Delegation Trigger**: Migration-related tasks, data transformations
- **Specialties**: Zero-downtime migrations, data integrity

#### `data-engineer`
- **Purpose**: Data pipeline development and ETL processes
- **Use Cases**: Data processing, ETL pipelines, data warehouse
- **Delegation Trigger**: Data engineering tasks, pipeline development
- **Specialties**: Big data, real-time processing

### Documentation & Communication

#### `docs-manager`
- **Purpose**: Technical documentation and knowledge management
- **Use Cases**: API documentation, user guides, knowledge bases
- **Delegation Trigger**: Documentation requirements, knowledge management
- **Standards**: Technical writing standards, documentation architecture

#### `git-manager`
- **Purpose**: Git workflow management and version control
- **Use Cases**: Branch management, pull requests, release management
- **Delegation Trigger**: Git-related tasks, version control
- **Specialties**: GitHub Flow, GitOps, release automation

### Project Management

#### `project-manager`
- **Purpose**: Project coordination and workflow management
- **Use Cases**: Project planning, milestone tracking, team coordination
- **Delegation Trigger**: Project management tasks, workflow optimization
- **Focus**: Agile methodologies, project delivery

#### `accessibility-expert`
- **Purpose**: Accessibility compliance and inclusive design
- **Use Cases**: WCAG compliance, accessibility testing, inclusive design
- **Delegation Trigger**: Accessibility requirements, compliance checks
- **Standards**: WCAG 2.1 AA/AAA, Section 508

### Specialized Tools

#### `agent-factory`
- **Purpose**: Create and configure new Claude Code agents
- **Use Cases**: Custom agent development, agent configuration
- **Delegation Trigger**: Agent creation, customization requirements
- **Output**: Configured agent definitions

#### `skill-factory`
- **Purpose**: Create and manage MoAI skill definitions
- **Use Cases**: Skill development, skill library management
- **Delegation Trigger**: Skill creation, skill optimization
- **Output**: Skill definition files

#### `format-expert`
- **Purpose**: Code formatting and style enforcement
- **Use Cases**: Code formatting, style guide compliance
- **Delegation Trigger**: Formatting requirements, style consistency
- **Specialties**: Language-specific formatters, linting

#### `debug-helper`
- **Purpose**: Error analysis and debugging assistance
- **Use Cases**: Runtime errors, debugging complex issues
- **Delegation Trigger**: Error analysis, debugging requirements
- **Process**: Error analysis → Root cause identification → Solution proposal

### System Agents (Claude Code Native)

#### `Explore`
- **Purpose**: Codebase discovery and file system exploration
- **Use Cases**: Code navigation, file search, project structure analysis
- **Delegation Trigger**: Exploration tasks, file system operations
- **Scope**: Read-only operations, code analysis

#### `Plan`
- **Purpose**: Strategic decomposition and planning
- **Use Cases**: Complex task breakdown, strategic planning
- **Delegation Trigger**: High-level planning, strategy development
- **Output**: Structured execution plans

## Delegation Patterns

### Simple Delegation
```
User: "Implement user authentication"
→ Delegate: tdd-implementer
→ Context: SPEC-001, security requirements
```

### Complex Workflow
```
User: "Build e-commerce platform"
→ 1. spec-builder (requirements)
→ 2. api-designer (API design)
→ 3. backend-expert (implementation)
→ 4. security-expert (security)
→ 5. docs-manager (documentation)
```

### Error Handling
```
Error: "Implementation failed"
→ 1. debug-helper (analyze error)
→ 2. relevant expert (fix issue)
→ 3. quality-gate (validate fix)
```

## Best Practices

1. **Always provide context**: Pass relevant information between agents
2. **Use appropriate agents**: Match task complexity to agent specialization
3. **Validate results**: Use quality-gate for critical implementations
4. **Handle errors gracefully**: Use debug-helper for troubleshooting
5. **Document outcomes**: Use docs-manager for knowledge capture

## Agent Selection Matrix

| Task Type | Primary Agent | Secondary Agents |
|------------|---------------|------------------|
| Requirements | spec-builder | docs-manager |
| Implementation | tdd-implementer | domain-expert, security-expert |
| API Design | api-designer | backend-expert |
| UI Development | frontend-expert | component-designer, ui-ux-expert |
| Database | database-expert | migration-expert |
| Security | security-expert | quality-gate |
| Performance | performance-engineer | monitoring-expert |
| Deployment | devops-expert | monitoring-expert |
| Documentation | docs-manager | domain-expert |
| Testing | test-engineer | quality-gate |
| Debugging | debug-helper | relevant domain expert |