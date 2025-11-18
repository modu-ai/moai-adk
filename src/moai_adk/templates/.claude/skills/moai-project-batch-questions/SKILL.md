---
name: moai-project-batch-questions
version: 4.0.0
created: 2025-11-05
updated: '2025-11-18'
status: stable
description: Standardize AskUserQuestion patterns and provide reusable question templates
  for batch optimization
keywords:
- batch
- questions
- templates
- optimization
- ux
- interaction
freedom: medium
type: project
tags:
- project
- questions
- batch
- templates
- ux
- optimization
allowed-tools:
- AskUserQuestion
stability: stable
---


# Project Batch Questions - Skill Guide

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-project-batch-questions |
| **Version** | 1.0.0 (2025-11-05) |
| **Core Tool** | `AskUserQuestion` (Claude Code built-in) |
| **UX Goal** | 60% interaction reduction through batching |

---

## What It Does

**Purpose**: Standardize AskUserQuestion patterns with **reusable batch templates** that reduce user interactions while maintaining clarity.

**Key capabilities**:
- ✅ **Batch Templates**: Pre-designed question groups for common scenarios
- ✅ **UX Optimization**: 60% interaction reduction through strategic batching  
- ✅ **Multi-language Support**: Templates in Korean, English, Japanese, Chinese
- ✅ **Response Validation**: Built-in validation and processing patterns
- ✅ **Error Handling**: Graceful handling of invalid or missing responses

---

## Batch Design Philosophy

### Traditional vs Batch Approach

**Traditional**: Q1 → Answer → Q2 → Answer → Q3 → Answer (3 interactions)
**Batch**: Q1 + Q2 + Q3 → All answers at once (1 interaction, 66% reduction)

### Batching Rules

| Rule | Description | Example |
|------|-------------|---------|
| **Related Questions** | Group questions about same topic | Language settings |
| **Sequential Logic** | Q2 depends on Q1 answer | Team mode conditional questions |
| **Same Decision Context** | User thinking about same aspect | GitHub + Git workflow |

---

## Core Batch Templates

### Template 1: Language Selection Batch (3 questions)

**Purpose**: Set language preferences for project initialization
**Interaction Reduction**: 3 turns → 1 turn (66% improvement)

```typescript
const languageBatch = {
  questions: [
    {
      question: "Which language would you like to use for project initialization and documentation?",
      header: "Language",
      multiSelect: false,
      options: [
        { label: "English", description: "All dialogs and documentation in English" },
        { label: "한국어", description: "모든 대화와 문서를 한국어로" },
        { label: "日本語", description: "すべての対話と文書を日本語で" },
        { label: "中文", description: "所有对话和文档使用中文" }
      ]
    },
    {
      question: "In which language should Alfred's sub-agent prompts be written?",
      header: "Agent Prompt",
      multiSelect: false,
      options: [
        { label: "English (Global Standard)", description: "Reduces token usage by 15-20%" },
        { label: "Selected Language (Localized)", description: "Local efficiency with native language" }
      ]
    },
    {
      question: "How would you like to be called in our conversations? (max 20 chars)",
      header: "Nickname",
      multiSelect: false,
      options: [
        { label: "Enter custom nickname", description: "Type your preferred name using 'Other' option" }
      ]
    }
  ]
};
```

### Template 2: Team Mode Settings Batch (2 questions)

**Purpose**: Configure team-specific GitHub and Git settings
**Interaction Reduction**: 2 turns → 1 turn (50% improvement)
**Conditional**: Only shown when `mode: "team"` detected

```typescript
const teamModeBatch = {
  questions: [
    {
      question: "[Team Mode] Is 'Automatically delete head branches' enabled in your GitHub repository?",
      header: "GitHub Settings", 
      multiSelect: false,
      options: [
        { label: "Yes, already enabled", description: "PR merge 후 자동으로 원격 브랜치 삭제됨" },
        { label: "No, not enabled (Recommended)", description: "Settings → General에서 확인 필요" },
        { label: "Not sure / Need to check", description: "GitHub Settings 확인 후 다시 진행" }
      ]
    },
    {
      question: "[Team Mode] Which Git workflow should we use for SPEC documents?",
      header: "Git Workflow",
      multiSelect: false,
      options: [
        { label: "Feature Branch + PR", description: "매 SPEC마다 feature 브랜치 생성 → PR 리뷰 → develop 병합" },
        { label: "Direct Commit to Develop", description: "develop에 직접 커밋. 빠른 프로토타이핑에 최적" },
        { label: "Decide per SPEC", description: "SPEC 생성 시마다 매번 선택. 유연성 높지만 결정 필요" }
      ]
    }
  ]
};
```

### Template 3: Report Generation Batch (1 question)

**Purpose**: Configure report generation with token cost awareness

```typescript
const reportGenerationBatch = {
  questions: [
    {
      question: "Configure report generation:\n\n⚡ **Minimal (Recommended)**: Essential reports only (20-30 tokens)\n📊 **Enable**: Full analysis reports (50-60 tokens)\n🚫 **Disable**: No reports (0 tokens)\n\nAffects future /alfred:3-sync costs.",
      header: "Report Generation",
      multiSelect: false,
      options: [
        { label: "⚡ Minimal (Recommended)", description: "80% token reduction, faster sync" },
        { label: "Enable", description: "Complete reports, higher token usage" },
        { label: "🚫 Disable", description: "No automatic reports, zero cost" }
      ]
    }
  ]
};
```

### Template 4: Domain Selection Batch (Multi-select)

**Purpose**: Select project domains and technology areas

```typescript
const domainSelectionBatch = {
  questions: [
    {
      question: "Which domains and technology areas should be included in this project?",
      header: "Domains",
      multiSelect: true,
      options: [
        { label: "Backend API", description: "REST/GraphQL APIs, server-side logic, databases" },
        { label: "Frontend Web", description: "React/Vue/Angular, UI components, client-side" },
        { label: "Mobile App", description: "iOS/Android apps, React Native, Flutter" },
        { label: "DevOps/Infrastructure", description: "CI/CD, Docker, Kubernetes, cloud" },
        { label: "Data/Analytics", description: "Data processing, ML pipelines, analytics" }
      ]
    }
  ]
};
```

---

## Response Processing

### Validation Function

```typescript
function validateBatchResponse(responses: Record<string, string>, template: string): ValidationResult {
  const errors: string[] = [];
  
  switch (template) {
    case 'language-batch':
      const validLanguages = ['ko', 'en', 'ja', 'zh'];
      if (!validLanguages.includes(responses['Language'])) {
        errors.push('Invalid language selection');
      }
      if (responses['Nickname']?.length > 20) {
        errors.push('Nickname must be 20 characters or less');
      }
      break;
  }
  
  return { isValid: errors.length === 0, errors };
}
```

### Configuration Mapping

```typescript
function mapToConfig(responses: Record<string, string>, template: string): Partial<Config> {
  switch (template) {
    case 'language-batch':
      return {
        language: {
          conversation_language: responses['Language'],
          agent_prompt_language: responses['Agent Prompt'] === 'English (Global Standard)' ? 'english' : 'localized'
        },
        user: { nickname: responses['Nickname'], selected_at: new Date().toISOString() }
      };
      
    case 'team-mode-batch':
      return {
        github: {
          auto_delete_branches: responses['GitHub Settings'] === 'Yes, already enabled',
          spec_git_workflow: mapWorkflowToCode(responses['Git Workflow']),
          checked_at: new Date().toISOString()
        }
      };
  }
}
```

---

## Usage Integration

### Alfred Command Integration

```typescript
// In 0-project.md command
async function initializeProject() {
  // Step 1: Language selection batch
  const languageResponses = await executeBatchTemplate(LANGUAGE_BATCH_TEMPLATE);
  
  // Step 2: Check for team mode
  if (isTeamMode()) {
    const teamResponses = await executeBatchTemplate(TEAM_MODE_BATCH_TEMPLATE);
  }
  
  // Step 3: Report generation batch  
  const reportResponses = await executeBatchTemplate(REPORT_GENERATION_BATCH_TEMPLATE);
}
```

---

## Performance Metrics

### Interaction Reduction

| Template | Traditional | Batch | Reduction |
|----------|-------------|-------|-----------|
| **Language Selection** | 3 interactions | 1 interaction | 66% |
| **Team Mode Settings** | 2 interactions | 1 interaction | 50% |
| **Domain Selection** | 5+ questions | 1 interaction | 80%+ |

---

## Best Practices

### ✅ DO

- **Group related questions**: Same decision context
- **Show total question count**: "3 questions in this batch"
- **Use consistent headers**: Short, descriptive (≤12 chars)
- **Include progress indicators**: "Step 1 of 2"

### ❌ DON'T

- **Overload batches**: Max 4 questions per batch
- **Mix unrelated topics**: Keep thematic cohesion
- **Skip validation**: Always verify responses
- **Ignore cancellation**: Handle user gracefully

---

## Quick Reference

### Common Use Cases

| Use Case | Template | Questions | Interaction Reduction |
|----------|----------|-----------|----------------------|
| **Project initialization** | Language + Team batches | 5 questions total | 60% |
| **Settings modification** | Targeted batches | 1-3 questions | 50-80% |
| **Feature configuration** | Domain-specific batches | 2-4 questions | 75% |

### Integration Checklist

- [ ] Template selected for use case
- [ ] Response validation configured  
- [ ] Error handling implemented
- [ ] Configuration mapping tested
- [ ] Multi-language support if needed

---

**End of Skill** | Created 2025-11-05 | Optimized for batch interaction reduction
