# SPEC-AGENT-FACTORY-001: Agent-Factory Optimization Plan

## 🏷️ TAG BLOCK

**TAG**: SPEC-AGENT-FACTORY-001
**Title**: Agent-Factory Optimization and Skill Reallocation
**Created**: 2025-11-20
**Status**: DRAFT
**Priority**: HIGH
**Domain**: System Architecture & Performance Optimization
**Complexity**: HIGH
**Estimated Duration**: Multi-phase implementation

---

## 🌍 Environment

**Current MoAI-ADK Context**:
- **Version**: 0.26.0 (release/0.26.0 branch)
- **Agents**: 31 total agents in production
- **Skills**: 144 available skills in library
- **Agent-Factory**: Currently configured with 24 skills
- **Target**: Optimize agent-factory and systematically update all 31 agents
- **User Configuration**: Korean language (conversation), English prompts (agent reasoning)

**Technical Context**:
- Claude Code integration with sub-agent delegation via Task()
- MCP (Model Context Protocol) integration with Context7
- Hybrid skill allocation: 20 core + domain-specific skills per agent
- Performance optimization through token management and skill loading

---

## 🎯 Assumptions

1. **Agent-Factory Role**: Primary agent for creating new Claude Code sub-agents with automated skill optimization
2. **Skill Library**: All 144 skills are functional and optimized for their respective domains
3. **Agent Performance**: Current skill allocation may be causing inefficiencies or redundancies
4. **Claude Code Standards**: Generated agents must comply with official Claude Code sub-agent patterns
5. **Model Selection**: Optimal model assignment (Sonnet/Haiku/Inherit) affects skill requirements
6. **Context7 Integration**: MCP tools provide up-to-date documentation for agent creation

---

## 📋 Requirements

### 🎯 Primary Requirements

#### R1: Agent-Factory Skill Optimization
- **Priority**: CRITICAL
- **Description**: Optimize agent-factory's 24 skills for maximum efficiency in agent creation
- **Acceptance Criteria**:
  - Reduce agent-factory skills to optimal range (15-20 skills)
  - Remove redundant or overlapping skills
  - Add missing critical skills for agent creation
  - Maintain full agent generation capabilities
  - Improve agent generation quality and speed

#### R2: Skill Allocation Framework
- **Priority**: HIGH
- **Description**: Create systematic approach for skill allocation across all 31 agents
- **Acceptance Criteria**:
  - Define skill categories: Essential, Important, Optional
  - Create skill-to-agent mapping matrix
  - Establish skill loading patterns (auto-load vs conditional)
  - Implement skill redundancy elimination
  - Balance skill distribution across agents

#### R3: Agent Update Strategy
- **Priority**: HIGH
- **Description**: Develop systematic approach to update all 31 agents with optimized skill configurations
- **Acceptance Criteria**:
  - Create agent prioritization framework
  - Develop batch update process
  - Maintain backward compatibility
  - Validate agent functionality post-update
  - Document all changes and rationale

#### R4: Performance Optimization
- **Priority**: MEDIUM
- **Description**: Optimize agent performance through intelligent skill loading and model selection
- **Acceptance Criteria**:
  - Implement dynamic skill loading based on context
  - Optimize model selection per agent type
  - Reduce token consumption in agent operations
  - Improve agent response times
  - Monitor and measure performance improvements

#### R5: Quality Assurance
- **Priority**: CRITICAL
- **Description**: Ensure all optimized agents meet quality standards and maintain functionality
- **Acceptance Criteria**:
  - Implement automated testing for agent functionality
  - Validate skill configurations per agent
  - Ensure TRUST 5 compliance maintained
  - Perform regression testing on all agents
  - Document quality metrics and improvements

### 🔧 Secondary Requirements

#### R6: Claude Code Integration Compliance
- **Priority**: HIGH
- **Description**: Ensure all optimized agents comply with Claude Code sub-agent standards
- **Acceptance Criteria**:
  - Follow official agent definition patterns
  - Validate tool permissions and MCP integration
  - Ensure proper Task() delegation patterns
  - Maintain compatibility with Claude Code features
  - Document compliance verification

#### R7: Documentation and Guidelines
- **Priority**: MEDIUM
- **Description**: Create comprehensive documentation for skill optimization and agent configuration
- **Acceptance Criteria**:
  - Document skill optimization methodology
  - Create agent configuration guidelines
  - Provide skill selection framework
  - Update agent development best practices
  - Create troubleshooting guides

---

## 📜 Specifications

### S1: Agent-Factory Skill Analysis and Optimization

#### S1.1 Current State Assessment
**Current Agent-Factory Skills (24 total)**:
```yaml
Core Skills (8):
  - moai-foundation-ears
  - moai-foundation-specs
  - moai-foundation-trust
  - moai-foundation-git
  - moai-foundation-langs
  - moai-core-personas
  - moai-core-workflow
  - moai-core-language-detection

Language Skills (7):
  - moai-lang-python
  - moai-lang-typescript
  - moai-lang-javascript
  - moai-lang-go
  - moai-lang-shell
  - moai-lang-sql

Essential Skills (5):
  - moai-essentials-debug
  - moai-essentials-perf
  - moai-essentials-refactor
  - moai-essentials-review
  - moai-core-code-reviewer

Domain Skills (3):
  - moai-domain-security
  - moai-core-agent-factory
  - moai-cc-configuration

Integration Skills (1):
  - moai-cc-skills
  - moai-core-dev-guide
```

#### S1.2 Skill Optimization Strategy

**Category Analysis**:

**ESSENTIAL (Must Keep - 8 skills)**:
```yaml
Agent Generation Core:
  - moai-core-agent-factory      # MASTER SKILL for agent creation
  - moai-foundation-ears         # EARS methodology for specs
  - moai-foundation-specs        # SPEC documentation patterns
  - moai-core-language-detection # Multilingual support
  - moai-core-workflow           # Agent workflow patterns
  - moai-core-personas           # Agent persona development
  - moai-cc-configuration        # Claude Code compliance
  - moai-cc-skills               # Skill management patterns
```

**IMPORTANT (Should Keep - 7 skills)**:
```yaml
Agent Creation Support:
  - moai-foundation-trust        # Quality assurance
  - moai-foundation-git          # Version control patterns
  - moai-foundation-langs        # Multi-language support
  - moai-essentials-debug        # Agent debugging capabilities
  - moai-essentials-review       # Code review for generated agents
  - moai-core-code-reviewer      # Automated code review
  - moai-domain-security         # Security patterns for agents
```

**OPTIONAL (Consider Removing - 6 skills)**:
```yaml
Context-Dependent:
  - moai-lang-python             # Conditional: Only for Python agents
  - moai-lang-typescript         # Conditional: Only for TypeScript agents
  - moai-lang-javascript         # Conditional: Only for JavaScript agents
  - moai-lang-go                 # Conditional: Only for Go agents
  - moai-lang-shell              # Conditional: Only for DevOps agents
  - moai-lang-sql                # Conditional: Only for Database agents
```

**REPLACE WITH** (New Critical Skills - 2 skills):
```yaml
Missing Critical Skills:
  - moai-context7-lang-integration  # Latest documentation research
  - moai-core-dev-guide             # Agent development best practices
```

#### S1.3 Optimized Agent-actory Configuration (Target: 17 skills)

```yaml
Essential Core (8):
  - moai-core-agent-factory
  - moai-foundation-ears
  - moai-foundation-specs
  - moai-core-language-detection
  - moai-core-workflow
  - moai-core-personas
  - moai-cc-configuration
  - moai-cc-skills

Important Support (7):
  - moai-foundation-trust
  - moai-foundation-git
  - moai-foundation-langs
  - moai-essentials-debug
  - moai-essentials-review
  - moai-core-code-reviewer
  - moai-domain-security

Critical Integration (2):
  - moai-context7-lang-integration
  - moai-core-dev-guide
```

### S2: Systematic Agent Optimization Framework

#### S2.1 Agent Classification System

**Agent Categories**:
```yaml
Category A - Planning & Architecture:
  - spec-builder
  - api-designer
  - implementation-planner
  - agent-factory

Category B - Implementation & Development:
  - tdd-implementer
  - backend-expert
  - frontend-expert
  - database-expert
  - component-designer

Category C - Quality & Assurance:
  - quality-gate
  - security-expert
  - performance-engineer
  - trust-checker
  - testing-expert

Category D - Integration & Operations:
  - devops-expert
  - monitoring-expert
  - migration-expert
  - cc-manager

Category E - Documentation & Management:
  - docs-manager
  - git-manager
  - project-manager
  - skill-factory
```

#### S2.2 Skill Allocation Matrix

**Skill Distribution Strategy**:
```yaml
Universal Core Skills (All Agents - 6 skills):
  - moai-foundation-ears       # EARS methodology
  - moai-foundation-trust      # Quality assurance
  - moai-core-language-detection # Multilingual support
  - moai-core-workflow         # Workflow patterns
  - moai-core-personas         # Agent behavior
  - moai-core-dev-guide        # Development best practices

Category-Specific Skills (Per Category - 8-12 skills):
  Planning: + moai-foundation-specs, moai-foundation-git, moai-cc-configuration
  Implementation: + moai-essentials-debug, moai-essentials-refactor, moai-core-code-reviewer
  Quality: + moai-essentials-review, moai-domain-security, moai-essentials-perf
  Operations: + moai-domain-devops, moai-domain-cloud, moai-ml-ops
  Documentation: + moai-docs-generation, moai-docs-validation

Domain-Specific Skills (Per Agent - 2-4 skills):
  - Language skills (conditional loading)
  - Framework-specific skills
  - Domain knowledge bases
  - Specialized tool integration
```

#### S2.3 Dynamic Skill Loading System

**Load Pattern Implementation**:
```yaml
Auto-Load Skills (Always Available):
  - Universal Core Skills (6)
  - Category-Specific Skills (8-12)

Conditional Skills (Load on Demand):
  - Language Detection → Load specific language skills
  - Domain Analysis → Load domain-specific skills
  - Complexity Assessment → Load advanced skills
  - User Requirements → Load specialized skills

Just-in-Time Skills:
  - MCP integration tools
  - Framework-specific documentation
  - Advanced analysis capabilities
```

### S3: Implementation Strategy

#### S3.1 Phase 1: Agent-Factory Optimization (Week 1)

**Deliverables**:
1. Reduce agent-factory skills from 24 to 17 skills
2. Implement dynamic skill loading for language skills
3. Add missing critical skills (Context7 integration, dev guide)
4. Test agent generation quality and performance
5. Document optimization rationale

**Validation Criteria**:
- Agent generation time reduced by 20%
- Generated agent quality maintained or improved
- All agent generation scenarios supported
- Memory usage optimized
- No functionality lost

#### S3.2 Phase 2: Category A Agents Optimization (Week 2)

**Target Agents**: spec-builder, api-designer, implementation-planner

**Optimization Tasks**:
1. Apply skill allocation matrix
2. Remove redundant language skills
3. Add category-specific skills
4. Implement dynamic loading
5. Validate functionality

#### S3.3 Phase 3: Category B & C Agents Optimization (Week 3-4)

**Target Agents**: Implementation and Quality agents (8 agents)

**Optimization Focus**:
1. Domain-specific skill optimization
2. Performance tuning for execution agents
3. Quality assurance skill integration
4. MCP tool optimization

#### S3.4 Phase 4: Remaining Agents Optimization (Week 5-6)

**Target Agents**: Categories D & E (13 agents)

**Optimization Priorities**:
1. Integration and operations agents
2. Documentation and management agents
3. Specialized domain agents
4. Cross-agent coordination skills

### S4: Quality Assurance and Testing

#### S4.1 Automated Testing Framework

**Test Categories**:
```yaml
Functionality Tests:
  - Agent generation success rate
  - Skill loading correctness
  - Agent response accuracy
  - Tool permission validation

Performance Tests:
  - Agent response times
  - Memory usage optimization
  - Token consumption reduction
  - Skill loading efficiency

Integration Tests:
  - Claude Code compatibility
  - MCP tool functionality
  - Task() delegation patterns
  - Cross-agent collaboration
```

#### S4.2 Quality Metrics

**Success Indicators**:
```yaml
Performance Metrics:
  - Agent response time: < 15 seconds (current: < 20 seconds)
  - Memory usage: 20% reduction
  - Token consumption: 15% reduction
  - Skill loading time: < 2 seconds

Quality Metrics:
  - Agent generation success rate: > 98%
  - Generated agent quality score: > 90%
  - Skill configuration accuracy: > 95%
  - Backward compatibility: 100%
```

### S5: Documentation and Guidelines

#### S5.1 Skill Optimization Documentation

**Contents**:
1. Skill categorization framework
2. Agent optimization methodology
3. Dynamic skill loading patterns
4. Performance optimization techniques
5. Troubleshooting and maintenance guides

#### S5.2 Agent Configuration Guidelines

**Documentation Structure**:
```yaml
Agent Development:
  - Skill selection guidelines
  - Configuration best practices
  - Performance optimization tips
  - Quality assurance checklists

Maintenance:
  - Skill update procedures
  - Agent versioning strategy
  - Compatibility testing protocols
  - Rollback procedures
```

---

## 🔗 Traceability

**Requirements Traceability Matrix**:

| Requirement | Spec Section | Test Case | Implementation Phase |
|-------------|--------------|-----------|---------------------|
| R1: Agent-Factory Optimization | S1.1-S1.3 | TC-AF-001 | Phase 1 |
| R2: Skill Allocation Framework | S2.1-S2.3 | TC-SAF-001 | All Phases |
| R3: Agent Update Strategy | S3.1-S3.4 | TC-AUS-001 | All Phases |
| R4: Performance Optimization | S2.3, S4.2 | TC-PO-001 | All Phases |
| R5: Quality Assurance | S4.1-S4.2 | TC-QA-001 | All Phases |
| R6: Claude Code Compliance | S1.3, S5.2 | TC-CCC-001 | All Phases |
| R7: Documentation | S5.1-S5.2 | TC-DOC-001 | All Phases |

**Dependencies**:
- **D1**: Claude Code sub-agent documentation (Context7 MCP)
- **D2**: Current agent functionality validation
- **D3**: Performance benchmarking baseline
- **D4**: Quality assurance framework

---

## 📅 Implementation Timeline

**Phase 1 (Week 1)**: Agent-Factory Core Optimization
- Day 1-2: Skill analysis and optimization design
- Day 3-4: Implementation and testing
- Day 5: Validation and documentation

**Phase 2 (Week 2)**: Category A Agents (Planning & Architecture)
- Day 1-3: spec-builder optimization
- Day 4-5: api-designer and implementation-planner optimization

**Phase 3 (Weeks 3-4)**: Categories B & C (Implementation & Quality)
- Week 3: Implementation agents (4 agents)
- Week 4: Quality agents (4 agents)

**Phase 4 (Weeks 5-6)**: Categories D & E (Operations & Documentation)
- Week 5: Operations agents (7 agents)
- Week 6: Documentation agents (6 agents)

**Phase 5 (Week 7)**: Integration Testing and Validation
- Comprehensive testing across all agents
- Performance validation
- Documentation finalization
- Rollout preparation

---

**Status**: Ready for Implementation Review
**Next Steps**:
1. Stakeholder approval of optimization strategy
2. Phase 1 implementation kick-off
3. Performance baseline establishment
4. Quality assurance framework setup