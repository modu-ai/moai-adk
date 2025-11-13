# Project Batch Questions - API Reference

> **Main Skill**: [SKILL.md](SKILL.md)  
> **Examples**: [examples.md](examples.md)

---

## Core API

### BatchTemplate Interface

```typescript
interface BatchTemplate {
  name: string;
  description: string;
  questions: Question[];
  conditional?: ConditionalLogic;
  validation?: ValidationRules;
  mapping?: ResponseMapping;
}
```

### Question Interface (Extended)

```typescript
interface Question {
  question: string;              // The question text
  header: string;               // Column header (max 12 chars)
  multiSelect: boolean;         // true = multiple selections
  options: Option[];            // 2-4 options recommended
  conditional?: QuestionCondition;  // Show/hide logic
  validation?: QuestionValidation;   // Input validation rules
}
```

### Option Interface (Extended)

```typescript
interface Option {
  label: string;               // 1-5 words, displayed in TUI
  description: string;         // Rationale and context
  value?: string;              // Internal value (optional)
  icon?: string;               // Icon for visual distinction
  warning?: string;            // Warning text for risky options
  recommended?: boolean;       // Mark as recommended choice
}
```

---

## Pre-built Batch Templates

### LANGUAGE_BATCH_TEMPLATE

```typescript
export const LANGUAGE_BATCH_TEMPLATE: BatchTemplate = {
  name: 'language-batch',
  description: 'Set language preferences for project initialization',
  questions: [
    {
      question: "Which language would you like to use for project initialization and documentation?",
      header: "Language",
      multiSelect: false,
      options: [
        {
          label: "English",
          description: "All dialogs and documentation in English",
          value: "en"
        },
        {
          label: "한국어", 
          description: "All dialogs and documentation in Korean",
          value: "ko"
        },
        {
          label: "日本語",
          description: "All dialogs and documentation in Japanese",
          value: "ja"
        },
        {
          label: "中文",
          description: "All dialogs and documentation in Chinese",
          value: "zh"
        }
      ]
    },
    {
      question: "In which language should Alfred's sub-agent prompts be written?",
      header: "Agent Prompt",
      multiSelect: false,
      options: [
        {
          label: "English (Global Standard)",
          description: "All sub-agent prompts in English. Reduces token usage by 15-20%",
          value: "english",
          recommended: true
        },
        {
          label: "Selected Language (Localized)",
          description: "All sub-agent prompts in your selected language for local efficiency",
          value: "localized"
        }
      ]
    },
    {
      question: "How would you like to be called in our conversations? (max 20 chars)",
      header: "Nickname",
      multiSelect: false,
      options: [
        {
          label: "Enter custom nickname",
          description: "Type your preferred name using the 'Other' option below",
          value: "custom"
        }
      ],
      validation: {
        maxLength: 20,
        required: true,
        pattern: /^[a-zA-Z0-9가-힣ぁ-ゟ一-龯\s]{1,20}$/
      }
    }
  ],
  mapping: {
    'Language': 'language.conversation_language',
    'Agent Prompt': 'language.agent_prompt_language', 
    'Nickname': 'user.nickname'
  }
};
```

### TEAM_MODE_BATCH_TEMPLATE

```typescript
export const TEAM_MODE_BATCH_TEMPLATE: BatchTemplate = {
  name: 'team-mode-batch',
  description: 'Configure team-specific GitHub and Git settings',
  conditional: {
    field: 'mode',
    operator: 'equals',
    value: 'team'
  },
  questions: [
    {
      question: "[Team Mode] Is 'Automatically delete head branches' enabled in your GitHub repository settings?",
      header: "GitHub Settings", 
      multiSelect: false,
      options: [
        {
          label: "Yes, already enabled",
          description: "PR merge 후 자동으로 원격 브랜치 삭제됨",
          value: "true"
        },
        {
          label: "No, not enabled (Recommended: Enable)",
          description: "Settings → General → '자동 삭제' 체크박스 확인 필요",
          value: "false",
          warning: "Recommended to enable for better repository hygiene"
        },
        {
          label: "Not sure / Need to check",
          description: "GitHub Settings → General 확인 후 다시 진행",
          value: "null"
        }
      ]
    },
    {
      question: "[Team Mode] Which Git workflow should we use when creating SPEC documents?",
      header: "Git Workflow",
      multiSelect: false,
      options: [
        {
          label: "Feature Branch + PR",
          description: "매 SPEC마다 feature 브랜치 생성 → PR 리뷰 → develop 병합",
          value: "feature_branch",
          recommended: true
        },
        {
          label: "Direct Commit to Develop", 
          description: "브랜치 생성 없이 develop에 직접 커밋. 빠른 프로토타이핑에 최적",
          value: "develop_direct"
        },
        {
          label: "Decide per SPEC",
          description: "SPEC 생성 시마다 매번 선택. 유연성이 높지만 결정 필요",
          value: "per_spec"
        }
      ]
    }
  ],
  mapping: {
    'GitHub Settings': 'github.auto_delete_branches',
    'Git Workflow': 'github.spec_git_workflow'
  }
};
```

### REPORT_GENERATION_BATCH_TEMPLATE

```typescript
export const REPORT_GENERATION_BATCH_TEMPLATE: BatchTemplate = {
  name: 'report-generation-batch',
  description: 'Configure report generation with token cost awareness',
  questions: [
    {
      question: `Configure report generation for better performance and cost management:

⚡ **Minimal (Recommended)**: Essential reports only (20-30 tokens/report)
📊 **Enable**: Full analysis reports (50-60 tokens/report)  
🚫 **Disable**: No automatic reports (0 tokens)

Choice affects future /alfred:3-sync execution time and costs.`,
      header: "Report Generation",
      multiSelect: false,
      options: [
        {
          label: "Minimal (Recommended)",
          description: "80% token reduction, faster sync times",
          value: "minimal",
          recommended: true
        },
        {
          label: "Enable",
          description: "Complete analysis reports, higher token usage",
          value: "enable"
        },
        {
          label: "Disable", 
          description: "No automatic reports, zero token cost",
          value: "disable",
          warning: "You'll miss important project insights and metrics"
        }
      ]
    }
  ],
  mapping: {
    'Report Generation': 'report_generation.enabled'
  }
};
```

### DOMAIN_SELECTION_BATCH_TEMPLATE

```typescript
export const DOMAIN_SELECTION_BATCH_TEMPLATE: BatchTemplate = {
  name: 'domain-selection-batch',
  description: 'Select project domains and technology areas',
  questions: [
    {
      question: "Which domains and technology areas should be included in this project?",
      header: "Domains",
      multiSelect: true,
      options: [
        {
          label: "Backend API",
          description: "REST/GraphQL APIs, server-side logic, databases",
          value: "backend"
        },
        {
          label: "Frontend Web", 
          description: "React/Vue/Angular, UI components, client-side logic",
          value: "frontend"
        },
        {
          label: "Mobile App",
          description: "iOS/Android apps, React Native, Flutter",
          value: "mobile"
        },
        {
          label: "DevOps/Infrastructure",
          description: "CI/CD, Docker, Kubernetes, cloud deployment",
          value: "devops"
        },
        {
          label: "Data/Analytics",
          description: "Data processing, ML pipelines, analytics dashboards",
          value: "data"
        }
      ]
    }
  ],
  mapping: {
    'Domains': 'project.domains'
  },
  validation: {
    minSelections: 1,
    maxSelections: 5
  }
};
```

---

## Template Execution API

### executeBatchTemplate()

```typescript
async function executeBatchTemplate(
  template: BatchTemplate,
  context?: ExecutionContext
): Promise<BatchResult>

interface ExecutionContext {
  currentConfig?: Partial<Config>;
  language?: string;
  mode?: string;
  skipValidation?: boolean;
}

interface BatchResult {
  success: boolean;
  responses: Record<string, string | string[]>;
  config: Partial<Config>;
  errors?: string[];
  warnings?: string[];
  metrics?: {
    executionTime: number;
    questionCount: number;
    tokenUsage: number;
  };
}
```

### Usage Example

```typescript
// Execute language batch with context
const result = await executeBatchTemplate(LANGUAGE_BATCH_TEMPLATE, {
  language: 'ko',  // Show questions in Korean
  currentConfig: existingConfig
});

if (result.success) {
  // Merge into existing config
  const updatedConfig = { ...existingConfig, ...result.config };
  await saveConfig('.moai/config.json', updatedConfig);
  
  console.log(`✅ Batch completed in ${result.metrics.executionTime}ms`);
} else {
  console.error('❌ Batch failed:', result.errors);
}
```

---

## Template Builder API

### createCustomBatch()

```typescript
function createCustomBatch(config: CustomBatchConfig): BatchTemplate

interface CustomBatchConfig {
  name: string;
  description: string;
  questions: QuestionConfig[];
  conditional?: ConditionalConfig;
  validation?: ValidationConfig;
}

interface QuestionConfig {
  question: string;
  header: string;
  multiSelect: boolean;
  options: OptionConfig[];
  conditional?: string;
  validation?: ValidationRuleConfig;
}
```

### Example

```typescript
const customBatch = createCustomBatch({
  name: 'database-selection',
  description: 'Select database and migration strategy',
  questions: [
    {
      question: "Which database should we use for this project?",
      header: "Database",
      multiSelect: false,
      options: [
        { label: "PostgreSQL", description: "Relational, ACID-compliant" },
        { label: "MongoDB", description: "Document store, flexible schema" },
        { label: "SQLite", description: "File-based, simple setup" }
      ]
    },
    {
      question: "How should we handle database migrations?",
      header: "Migrations",
      multiSelect: false,
      conditional: "Database !== SQLite",  // Only show if not SQLite
      options: [
        { label: "Alembic (Python)", description: "Version control for schema" },
        { label: "Prisma Migrate", description: "Type-safe migrations" },
        { label: "Manual SQL", description: "Custom migration scripts" }
      ]
    }
  ]
});
```

---

## Validation API

### validateBatchResponses()

```typescript
function validateBatchResponses(
  responses: Record<string, string | string[]>,
  template: BatchTemplate
): ValidationResult

interface ValidationResult {
  isValid: boolean;
  errors: ValidationError[];
  warnings: ValidationWarning[];
}

interface ValidationError {
  field: string;
  message: string;
  code: string;
}

interface ValidationWarning {
  field: string;
  message: string;
  code: string;
}
```

### Built-in Validation Rules

```typescript
export const VALIDATION_RULES = {
  REQUIRED: 'required',
  MAX_LENGTH: 'max_length',
  MIN_LENGTH: 'min_length',
  PATTERN: 'pattern',
  MIN_SELECTIONS: 'min_selections',
  MAX_SELECTIONS: 'max_selections',
  UNIQUE_VALUES: 'unique_values'
};
```

### Custom Validation Example

```typescript
const templateWithValidation: BatchTemplate = {
  name: 'project-settings',
  questions: [
    {
      question: "Enter project name (3-50 chars, alphanumeric + spaces)",
      header: "Project Name",
      multiSelect: false,
      options: [{ label: "Enter name", description: "Use 'Other' option" }],
      validation: {
        required: true,
        minLength: 3,
        maxLength: 50,
        pattern: /^[a-zA-Z0-9\s]{3,50}$/,
        errorMessage: "Project name must be 3-50 characters, alphanumeric and spaces only"
      }
    }
  ]
};
```

---

## Configuration Mapping API

### mapResponsesToConfig()

```typescript
function mapResponsesToConfig(
  responses: Record<string, string | string[]>,
  template: BatchTemplate
): Partial<Config>

interface ConfigMapping {
  [header: string]: string;  // Maps question header to config path
}
```

### Advanced Mapping Example

```typescript
const advancedTemplate: BatchTemplate = {
  name: 'advanced-setup',
  questions: [
    {
      question: "Select project deployment environments",
      header: "Environments",
      multiSelect: true,
      options: [
        { label: "Development", value: "dev" },
        { label: "Staging", value: "staging" }, 
        { label: "Production", value: "prod" }
      ]
    }
  ],
  mapping: {
    'Environments': 'deployment.environments'
  },
  postProcessors: {
    'Environments': (value: string[]) => ({
      enabled: value,
      default: value.includes('prod') ? 'staging' : value[0]
    })
  }
};
```

---

## Multi-Language Support API

### getLocalizedTemplate()

```typescript
function getLocalizedTemplate(
  templateName: string,
  language: string
): BatchTemplate
```

### Supported Languages

```typescript
export const SUPPORTED_LANGUAGES = {
  KO: 'ko',    // Korean
  EN: 'en',    // English  
  JA: 'ja',    // Japanese
  ZH: 'zh'     // Chinese
};
```

### Example

```typescript
// Get Korean version of language batch
const koreanTemplate = getLocalizedTemplate('language-batch', 'ko');

// The template will have:
// - Korean question text
// - Korean option labels
// - Korean descriptions
// - Same validation rules
// - Same mapping logic
```

---

## Error Handling API

### BatchError Classes

```typescript
class BatchValidationError extends Error {
  constructor(
    public validationErrors: ValidationError[],
    message?: string
  ) {
    super(message || 'Batch validation failed');
  }
}

class BatchExecutionError extends Error {
  constructor(
    public originalError: Error,
    public templateName: string,
    message?: string
  ) {
    super(message || `Batch execution failed: ${templateName}`);
  }
}

class UserCancelledError extends Error {
  constructor(public templateName: string) {
    super(`User cancelled batch: ${templateName}`);
  }
}
```

### Error Handling Pattern

```typescript
try {
  const result = await executeBatchTemplate(template, context);
  // Process result...
} catch (error) {
  if (error instanceof BatchValidationError) {
    // Handle validation errors
    console.error('Validation failed:', error.validationErrors);
    await retryWithCorrections(error.validationErrors);
  } else if (error instanceof UserCancelledError) {
    // Handle user cancellation
    console.log('User cancelled:', error.templateName);
    await offerPartialSave();
  } else {
    // Handle other errors
    console.error('Unexpected error:', error);
    await fallbackToManualInput();
  }
}
```

---

## Performance Monitoring API

### BatchMetrics Class

```typescript
class BatchMetrics {
  private metrics: MetricEntry[] = [];
  
  recordExecution(entry: MetricEntry): void;
  getMetrics(templateName?: string): MetricEntry[];
  generateReport(): PerformanceReport;
  exportMetrics(): string;  // JSON export
}

interface MetricEntry {
  templateName: string;
  timestamp: Date;
  executionTime: number;
  questionCount: number;
  interactionCount: number;
  tokenUsage: number;
  success: boolean;
  errorType?: string;
}
```

### Usage Example

```typescript
const metrics = new BatchMetrics();

// Track execution
const result = await metrics.trackExecution('language-batch', async () => {
  return await executeBatchTemplate(LANGUAGE_BATCH_TEMPLATE);
});

// Generate report
const report = metrics.generateReport();
console.log(`Total interactions saved: ${report.interactionsSaved}`);
console.log(`Average execution time: ${report.avgExecutionTime}ms`);
```

---

## Integration Points

### Alfred Command Integration

```typescript
// In 0-project.md
async function handleProjectInitialization() {
  const context = { 
    language: detectUserLanguage(),
    mode: detectProjectMode() 
  };
  
  // Execute language batch
  const langResult = await executeBatchTemplate(LANGUAGE_BATCH_TEMPLATE, context);
  
  // Conditionally execute team batch
  if (context.mode === 'team') {
    const teamResult = await executeBatchTemplate(TEAM_MODE_BATCH_TEMPLATE, context);
  }
  
  // Continue with initialization...
}
```

### Sub-agent Integration

```typescript
// In project-manager agent
async function conductProjectInterview() {
  const batches = [
    DOMAIN_SELECTION_BATCH_TEMPLATE,
    LANGUAGE_BATCH_TEMPLATE,
    REPORT_GENERATION_BATCH_TEMPLATE
  ];
  
  const results = {};
  for (const batch of batches) {
    const result = await executeBatchTemplate(batch);
    Object.assign(results, result.config);
  }
  
  return results;
}
```

---

**End of Reference** | Created 2025-11-05 | Complete API documentation
