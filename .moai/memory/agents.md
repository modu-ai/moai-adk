# MoAI-ADK Agents Reference

Alfred's agent delegation reference. Each agent is optimized for specific tasks.

## Planning & Specification

- `spec-builder`: SPEC generation in EARS format
- `plan`: Decompose complex tasks step-by-step

## Implementation

- `tdd-implementer`: Execute TDD cycle (RED-GREEN-REFACTOR)
- `backend-expert`: Backend architecture and API development
- `frontend-expert`: Frontend UI component development
- `database-expert`: Database schema design and optimization

## Quality & Testing

- `security-expert`: Security analysis and OWASP validation
- `quality-gate`: Code quality validation (TRUST 5)
- `test-engineer`: Test strategy and implementation

## Architecture & Design

- `api-designer`: REST/GraphQL API design
- `component-designer`: Reusable component design
- `ui-ux-expert`: User experience and interface design

## DevOps & Infrastructure

- `devops-expert`: CI/CD pipeline and deployment
- `monitoring-expert`: Monitoring and observability
- `performance-engineer`: Performance optimization and analysis

## Data & Integration

- `migration-expert`: Database migration
- `data-engineer`: Data pipeline development

## Documentation & Process

- `docs-manager`: Technical and API documentation generation
- `git-manager`: Git workflow and version management
- `project-manager`: Project coordination and planning

## Specialized Services

- `accessibility-expert`: WCAG accessibility validation
- `debug-helper`: Error analysis and solution suggestions
- `agent-factory`: New agent creation and configuration
- `skill-factory`: Skill definition creation and management
- `command-factory`: Custom slash command creation and optimization
- `format-expert`: Code formatting and style consistency

## System Agents

- `Explore`: Codebase exploration and file system analysis
- `Plan`: Strategy decomposition and planning

---

**Delegation Principles**:
1. Alfred always delegates to specialized agents via Task().
2. Analyze request complexity and dependencies to determine sequential or parallel execution.
3. Pass each agent's results as context to the next agent.

**Agent Selection Criteria**:
- Simple tasks (1 file): 1-2 agents sequential execution
- Medium tasks (3-5 files): 2-3 agents sequential execution
- Complex tasks (10+ files): 5+ agents mixed parallel/sequential execution

---

Refer to CLAUDE.md for detailed agent descriptions.
