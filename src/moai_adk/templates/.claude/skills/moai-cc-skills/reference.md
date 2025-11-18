# Claude Code Skills - Reference Documentation

## Official Documentation

### Core References
- **Claude Code Skills System**: [Built-in Documentation]
- **Skill Development Guidelines**: `.claude/skills/` directory structure
- **Progressive Disclosure**: Skill loading and activation patterns

### Skill Architecture

#### Knowledge Capsule Structure
```
skill-name/
├── SKILL.md          # Main skill documentation (≤500 words)
├── reference.md      # External links and deep documentation
├── examples.md       # Usage examples and patterns
├── scripts/          # Automation scripts (optional)
└── templates/        # Reusable templates (optional)
```

#### Metadata Standards
```yaml
---
name: skill-name                    # Unique identifier
version: X.Y.Z                      # Semantic versioning
created: YYYY-MM-DD                 # Creation date
updated: YYYY-MM-DD                 # Last modification
status: active|deprecated|experimental
description: Brief description     # Single sentence purpose
keywords: ['keyword1', 'keyword2']  # Search and discovery terms
allowed-tools: ['Tool1', 'Tool2']   # Required tool access
---
```

#### Progressive Disclosure System
- **Keyword Matching**: Skills load when keywords detected in context
- **Context Awareness**: Automatic skill activation based on conversation
- **Memory Efficiency**: <500-word capsules for optimal loading
- **Caching System**: Recently used skills remain in memory
- **Dependency Resolution**: Automatic skill dependency loading

### Skill Creation Patterns

#### Pattern 1: Domain Expertise Skill
- **Purpose**: Deep knowledge in specific technology domain
- **Structure**: What it does, When to use, Core patterns, Dependencies
- **Examples**: moai-lang-python, moai-domain-backend, moai-baas-foundation

#### Pattern 2: Process/Workflow Skill
- **Purpose**: Step-by-step guidance for complex processes
- **Structure**: Workflow phases, Decision trees, Checklists
- **Examples**: moai-core-agent-guide, moai-foundation-trust

#### Pattern 3: Tool/Integration Skill
- **Purpose**: Third-party service integration patterns
- **Structure**: Setup, Configuration, Common use cases, Troubleshooting
- **Examples**: moai-baas-auth0-ext, moai-context7-integration

#### Pattern 4: Foundation/Low-Level Skill
- **Purpose**: Core principles and fundamental patterns
- **Structure**: Principles, Standards, Validation, Governance
- **Examples**: moai-foundation-tags, moai-foundation-specs

### Quality Standards

#### Content Requirements
- **Length**: ≤500 words for main content (SKILL.md)
- **Structure**: Consistent sections and formatting
- **Clarity**: Clear, actionable guidance
- **Completeness**: All required metadata present
- **Accuracy**: Up-to-date information and examples

#### Validation Checklist
- [ ] Complete YAML metadata with all required fields
- [ ] Clear description of purpose and capabilities
- [ ] "When to Use" section with specific triggers
- [ ] Core patterns or implementation guidance
- [ ] Dependencies and integration information
- [ ] "Works Well With" skill cross-references
- [ ] Changelog with version history
- [ ] Reference documentation links
- [ ] Practical examples (examples.md)

### Skill Discovery and Loading

#### Keyword-Based Activation
```typescript
// Example skill activation logic
if (conversation.includes('authentication') || 
    conversation.includes('auth0') || 
    conversation.includes('user management')) {
  loadSkill('moai-baas-auth0-ext');
}
```

#### Contextual Loading
- **Project Context**: Skills relevant to current project type
- **Conversation History**: Previously discussed topics
- **User Expertise Level**: Appropriate complexity selection
- **Task Type**: Specific skill requirements based on current task

### Integration with Alfred Workflow

#### Skill Usage in 4-Step Process
1. **Intent Understanding**: Skills loaded for clarification
2. **Plan Creation**: Domain expertise skills for analysis
3. **Task Execution**: Implementation and process skills
4. **Report & Commit**: Documentation and validation skills

#### Agent-Skill Interaction
```python
# Example agent skill usage pattern
def handle_authentication_task(task):
    # Load relevant skill
    skill = load_skill('moai-baas-auth0-ext')
    
    # Get expert guidance
    guidance = skill.get_implementation_patterns()
    
    # Apply patterns to specific context
    return apply_patterns(task, guidance)
```

## External References

### Documentation Standards
- **Markdown Best Practices**: [CommonMark Specification](https://commonmark.org/)
- **Technical Writing**: [Google Developer Documentation Style Guide](https://developers.google.com/tech-writing)
- **API Documentation**: [OpenAPI Specification](https://swagger.io/specification/)

### Knowledge Management
- **Knowledge Engineering**: [Stanford Knowledge Systems Lab](https://ksl.stanford.edu/)
- **Information Architecture**: [Information Architecture Institute](https://iainstitute.org/)
- **Cognitive Load Theory**: [Cognitive Science Research](https://www.cogsci.ucsd.edu/)

---

**Last Updated**: 2025-11-11
**Related Skills**: moai-cc-memory, moai-cc-settings, moai-foundation-trust
