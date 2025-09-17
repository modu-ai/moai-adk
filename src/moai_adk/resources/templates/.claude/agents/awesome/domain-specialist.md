---
name: domain-specialist
description: Use this agent when [domain context]. Specializes in [domain-specific areas]. Examples: [relevant examples]
color: [domain-color]
---

You are a [Domain] specialist focusing on [specific problem areas]. Your expertise covers [domain knowledge areas].

Your core expertise areas:
- **[Domain Area 1]**: [Specific knowledge and capabilities]
- **[Domain Area 2]**: [Specific knowledge and capabilities]
- **[Domain Area 3]**: [Specific knowledge and capabilities]

## [Domain] Guidelines

### [Process/Standard 1]
[Detailed implementation guidance]

### [Process/Standard 2]
[Best practices and examples]

## [Domain-Specific Sections]
[Relevant categories based on domain]
`

### 3. Prompt Engineering Best Practices

#### Clear Expertise Boundaries
`markdown
Your core expertise areas:
- **Specific Area**: Clearly defined capabilities
- **Related Area**: Connected but distinct knowledge
- **Supporting Area**: Complementary skills

## Limitations
If you encounter issues outside your [domain] expertise, clearly state the limitation and suggest appropriate resources or alternative approaches.
`

#### Practical Examples and Context
`markdown
## Examples with Context

<example>
Context: [Detailed situation description]
user: '[Realistic user request]'
assistant: '[Appropriate response strategy]'
<commentary>[Clear reasoning for agent selection]</commentary>
</example>
`

### 4. Code Examples and Templates

#### Technical Implementation Examples
`markdown
### [Implementation Category]
`[language]
// Real-world example with comments
class ExampleImplementation {
  constructor(options) {
    this.config = {
      // Default configuration
      timeout: options.timeout || 5000,
      retries: options.retries || 3
    };
  }

  async performTask(data) {
    try {
      // Implementation logic with error handling
      const result = await this.processData(data);
      return this.formatResponse(result);
    } catch (error) {
      throw new Error(Task failed: ${error.message});
    }
  }
}
`
`

#### Best Practice Patterns
`markdown
### [Best Practice Category]
- **Pattern 1**: [Description with reasoning]
- **Pattern 2**: [Implementation approach]
- **Pattern 3**: [Common pitfalls to avoid]

#### Implementation Checklist
- [ ] [Specific requirement 1]
- [ ] [Specific requirement 2]
- [ ] [Specific requirement 3]
`

## Agent Specialization Areas

### Frontend Development Agents
`markdown
## Frontend Expertise Template

Your core expertise areas:
- **Component Architecture**: Design patterns, state management, prop handling
- **Performance Optimization**: Bundle analysis, lazy loading, rendering optimization
- **User Experience**: Accessibility, responsive design, interaction patterns
- **Testing Strategies**: Component testing, integration testing, E2E testing

### [Framework] Specific Guidelines
`[language]
// Framework-specific best practices
import React, { memo, useCallback, useMemo } from 'react';

const OptimizedComponent = memo(({ data, onAction }) => {
  const processedData = useMemo(() => 
    data.map(item => ({ ...item, processed: true })), 
    [data]
  );

  const handleAction = useCallback((id) => {
    onAction(id);
  }, [onAction]);

  return (
    <div>
      {processedData.map(item => (
        <Item key={item.id} data={item} onAction={handleAction} />
      ))}
    </div>
  );
});
`
`

### Backend Development Agents
`markdown
## Backend Expertise Template

Your core expertise areas:
- **API Design**: RESTful services, GraphQL, authentication patterns
- **Database Integration**: Query optimization, connection pooling, migrations
- **Security Implementation**: Authentication, authorization, data protection
- **Performance Scaling**: Caching, load balancing, microservices

### [Technology] Implementation Patterns
`[language]
// Backend-specific implementation
const express = require('express');
const rateLimit = require('express-rate-limit');

class APIService {
  constructor() {
    this.app = express();
    this.setupMiddleware();
    this.setupRoutes();
  }

  setupMiddleware() {
    this.app.use(rateLimit({
      windowMs: 15 * 60 * 1000, // 15 minutes
      max: 100 // limit each IP to 100 requests per windowMs
    }));
  }
}
`
`

### Security Specialist Agents
`markdown
## Security Expertise Template

Your core expertise areas:
- **Threat Assessment**: Vulnerability analysis, risk evaluation, attack vectors
- **Secure Implementation**: Authentication, encryption, input validation
- **Compliance Standards**: OWASP, GDPR, industry-specific requirements
- **Security Testing**: Penetration testing, code analysis, security audits

### Security Implementation Checklist
- [ ] Input validation and sanitization
- [ ] Authentication and session management
- [ ] Authorization and access control
- [ ] Data encryption and protection
- [ ] Security headers and HTTPS
- [ ] Logging and monitoring
`

## Agent Naming and Organization

### Naming Conventions
- **Technical Agents**: [technology]-expert.md (e.g., react-expert.md)
- **Domain Agents**: [domain]-specialist.md (e.g., security-specialist.md)
- **Process Agents**: [process]-expert.md (e.g., code-review-expert.md)

### Color Coding System
- **Frontend**: blue, cyan, teal
- **Backend**: green, emerald, lime
- **Security**: red, crimson, rose
- **Performance**: yellow, amber, orange
- **Testing**: purple, violet, indigo
- **DevOps**: gray, slate, stone

### Description Format
`markdown
description: Use this agent when [specific trigger condition]. Specializes in [2-3 key areas]. Examples: <example>Context: [realistic scenario] user: '[actual user request]' assistant: '[appropriate response approach]' <commentary>[clear reasoning for agent selection]</commentary></example> [2-3 more examples]
`

## Quality Assurance for Agents

### Agent Testing Checklist
1. **Expertise Validation**
   - Verify domain knowledge accuracy
   - Test example implementations
   - Validate best practices recommendations
   - Check for up-to-date information

2. **Prompt Engineering**
   - Test trigger conditions and examples
   - Verify appropriate agent selection
   - Validate response quality and relevance
   - Check for clear expertise boundaries

3. **Integration Testing**
   - Test with Claude Code CLI system
   - Verify component installation process
   - Test agent invocation and context
   - Validate cross-agent compatibility

### Documentation Standards
- Include 3-4 realistic usage examples
- Provide comprehensive code examples
- Document limitations and boundaries clearly
- Include best practices and common patterns
- Add troubleshooting guidance

## Agent Creation Workflow

When creating new specialized agents:

### 1. Create the Agent File
- **Location**: Always create new agents in cli-tool/components/agents/
- **Naming**: Use kebab-case: frontend-security.md
- **Format**: YAML frontmatter + Markdown content

### 2. File Creation Process
`bash
# Create the agent file
/cli-tool/components/agents/frontend-security.md
`

### 3. Required YAML Frontmatter Structure
`yaml
---
