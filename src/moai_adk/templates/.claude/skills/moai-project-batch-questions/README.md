# moai-project-batch-questions

A skill for standardizing AskUserQuestion patterns and providing reusable question templates with batch optimization for the MoAI-ADK project.

## Files

- **SKILL.md** - Main skill documentation (304 lines)
- **examples.md** - Real-world usage examples (417 lines)  
- **reference.md** - Complete API reference (704 lines)
- **README.md** - This file

## Key Features

- **60% interaction reduction** through strategic question batching
- **Multi-language support** (Korean, English, Japanese, Chinese)
- **Reusable templates** for common scenarios
- **Built-in validation** and error handling
- **Token-efficient** question design

## Quick Start

```typescript
// Use language selection batch
const result = await executeBatchTemplate(LANGUAGE_BATCH_TEMPLATE);

// Use team mode batch (conditional)
if (isTeamMode()) {
  const teamResult = await executeBatchTemplate(TEAM_MODE_BATCH_TEMPLATE);
}
```

## Integration

- **Alfred Commands**: Use in `/alfred:0-project` and other commands
- **Sub-agents**: Integrate with project-manager, spec-builder, etc.
- **Configuration**: Save responses to `.moai/config/config.json`

## Performance

| Template | Traditional | Batch | Reduction |
|----------|-------------|-------|-----------|
| Language Selection | 3 interactions | 1 interaction | 66% |
| Team Mode Settings | 2 interactions | 1 interaction | 50% |
| Domain Selection | 5+ questions | 1 interaction | 80%+ |

---

**Created**: 2025-11-05  
**Version**: 1.0.0  
**Total Lines**: 1,425 across all files
