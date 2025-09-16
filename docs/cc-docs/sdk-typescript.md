# Claude Code TypeScript SDK

## Overview

The Claude Code TypeScript SDK provides a powerful, type-safe interface for building AI agents with advanced tool access and conversation management. Perfect for Node.js applications, web development, and TypeScript-based automation.

## Installation

### Prerequisites

- Node.js 18 or newer
- TypeScript 4.5+ (optional but recommended)

### Install

```bash
npm install -g @anthropic-ai/claude-code
```

### Verify Installation

```bash
claude --version
```

## Authentication

Set up authentication before using the SDK:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

**Alternative Providers:**

```bash
export CLAUDE_CODE_USE_BEDROCK=1    # For AWS Bedrock
export CLAUDE_CODE_USE_VERTEX=1     # For Google Vertex AI
```

## Quick Start

### Simple Query

```typescript
import { query } from '@anthropic-ai/claude-code';

// Basic query
for await (const message of query(
  'Analyze this TypeScript file for improvements'
)) {
  if (message.type === 'result') {
    console.log(message.result);
  }
}
```

### With Configuration

```typescript
import { query, ClaudeCodeOptions } from '@anthropic-ai/claude-code';

const options: ClaudeCodeOptions = {
  systemPrompt: 'You are a TypeScript expert',
  maxTurns: 3,
  allowedTools: ['Read', 'Edit', 'Grep'],
  permissionMode: 'acceptEdits',
};

for await (const message of query({
  prompt: 'Refactor this component for better performance',
  options,
})) {
  if (message.type === 'result') {
    console.log(message.result);
  }
}
```

## Core Interfaces

### ClaudeCodeOptions

```typescript
interface ClaudeCodeOptions {
  // Core configuration
  systemPrompt?: string; // Custom system instructions
  maxTurns?: number; // Limit conversation rounds

  // Tool control
  allowedTools?: string[]; // Restrict tool access
  permissionMode?: PermissionMode; // Control approval flow

  // Model settings
  model?: string; // Override default model

  // Session management
  resumeSessionId?: string; // Resume specific session
  workingDirectories?: string[]; // Additional work directories

  // Output control
  verbose?: boolean; // Enable debug output
  stream?: boolean; // Enable streaming responses

  // Advanced options
  environmentVariables?: Record<string, string>;
  timeout?: number; // Operation timeout in seconds
  abortController?: AbortController; // Cancel operations
}

type PermissionMode = 'ask' | 'acceptEdits' | 'acceptAll' | 'bypassPermissions';
```

### Message Types

```typescript
interface MessageResult {
  type: 'result';
  result: string;
  timestamp: Date;
  sessionId: string;
}

interface MessageThinking {
  type: 'thinking';
  content: string;
}

interface MessageToolUse {
  type: 'tool_use';
  toolName: string;
  input: Record<string, any>;
  status: 'started' | 'completed' | 'failed';
}

interface MessageError {
  type: 'error';
  error: string;
  code?: string;
}

type ClaudeMessage =
  | MessageResult
  | MessageThinking
  | MessageToolUse
  | MessageError;
```

## Advanced Usage

### Multi-Turn Conversations

```typescript
import { query } from '@anthropic-ai/claude-code';

async function codeReviewSession() {
  const options = {
    systemPrompt: 'You are a senior code reviewer',
    maxTurns: 10,
    allowedTools: ['Read', 'Grep', 'Bash'],
    permissionMode: 'acceptEdits' as const,
  };

  // Start the review session
  for await (const message of query({
    prompt: 'Review the authentication module for security issues',
    options,
  })) {
    switch (message.type) {
      case 'thinking':
        console.log('🤔', message.content);
        break;

      case 'tool_use':
        console.log(`🔧 Using ${message.toolName}:`, message.input);
        break;

      case 'result':
        console.log('✅ Result:', message.result);
        break;

      case 'error':
        console.error('❌ Error:', message.error);
        break;
    }
  }
}

codeReviewSession();
```

### Session Management

```typescript
import { query } from "@anthropic-ai/claude-code";

class DebuggingSession {
  private sessionId?: string;

  async startDebugging(issue: string) {
    const options = {
      systemPrompt: "You are a debugging expert",
      maxTurns: 15,
      allowedTools: ["Read", "Bash", "Edit"],
      permissionMode: "acceptAll" as const
    };

    for await (const message of query({
      prompt: `Help me debug: ${issue}`,
      options
    })) {
      if (message.type === "result") {
        this.sessionId = message.sessionId;
        console.log("Debugging started:", message.result);
        break;
      }
    }
  }

  async continueDebugging(additionalInfo: string) {
    if (!this.sessionId) {
      throw new Error("No active debugging session");
    }

    const options = {
      resumeSessionId: this.sessionId,
      systemPrompt: "You are a debugging expert",
      maxTurns: 10
    };

    for await (const message of query({
      prompt: additionalInfo,
      options
    })) {
      if (message.type === "result") {
        console.log("Continue debugging:", message.result);
      }
    }
  }
}

// Usage
const debugger = new DebuggingSession();
await debugger.startDebugging("API calls are timing out");
await debugger.continueDebugging("The timeout happens only in production");
```

### Streaming Input Mode

```typescript
import { query } from '@anthropic-ai/claude-code';

async function* generatePrompts() {
  yield 'Start analyzing the codebase';
  yield 'Focus on performance bottlenecks';
  yield 'Suggest specific optimizations';
}

const options = {
  systemPrompt: 'You are a performance optimization expert',
  allowedTools: ['Read', 'Grep', 'Glob'],
  stream: true,
};

for await (const message of query({
  prompt: generatePrompts(),
  options,
})) {
  if (message.type === 'result') {
    console.log(message.result);
  }
}
```

### Image Attachments

```typescript
import { query } from '@anthropic-ai/claude-code';
import { readFileSync } from 'fs';

// Attach image for analysis
const imageBuffer = readFileSync('screenshot.png');

for await (const message of query({
  prompt: 'Analyze this UI screenshot and suggest improvements',
  images: [imageBuffer],
  options: {
    systemPrompt: 'You are a UI/UX expert',
    maxTurns: 3,
  },
})) {
  if (message.type === 'result') {
    console.log(message.result);
  }
}
```

## Practical Examples

### Automated Code Review

```typescript
import { query } from '@anthropic-ai/claude-code';
import { glob } from 'glob';

interface ReviewResult {
  file: string;
  issues: string[];
  suggestions: string[];
}

class CodeReviewer {
  private options = {
    systemPrompt: `
      You are a senior code reviewer. For each file, identify:
      1. Security vulnerabilities
      2. Performance issues  
      3. Code quality problems
      4. Best practice violations
      
      Return findings in JSON format.
    `,
    allowedTools: ['Read', 'Grep'],
    maxTurns: 3,
    permissionMode: 'acceptEdits' as const,
  };

  async reviewFiles(pattern: string): Promise<ReviewResult[]> {
    const files = await glob(pattern);
    const results: ReviewResult[] = [];

    for (const file of files) {
      console.log(`Reviewing ${file}...`);

      for await (const message of query({
        prompt: `Review this file: ${file}`,
        options: this.options,
      })) {
        if (message.type === 'result') {
          try {
            const review = JSON.parse(message.result);
            results.push({
              file,
              issues: review.issues || [],
              suggestions: review.suggestions || [],
            });
          } catch (error) {
            console.error(`Failed to parse review for ${file}:`, error);
          }
          break;
        }
      }
    }

    return results;
  }

  async generateReport(results: ReviewResult[]): Promise<string> {
    let report = '# Code Review Report\n\n';

    for (const result of results) {
      report += `## ${result.file}\n\n`;

      if (result.issues.length > 0) {
        report += '### Issues\n';
        result.issues.forEach((issue) => {
          report += `- ${issue}\n`;
        });
        report += '\n';
      }

      if (result.suggestions.length > 0) {
        report += '### Suggestions\n';
        result.suggestions.forEach((suggestion) => {
          report += `- ${suggestion}\n`;
        });
        report += '\n';
      }
    }

    return report;
  }
}

// Usage
const reviewer = new CodeReviewer();
const results = await reviewer.reviewFiles('src/**/*.ts');
const report = await reviewer.generateReport(results);
console.log(report);
```

### Smart Testing Assistant

```typescript
import { query } from '@anthropic-ai/claude-code';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

class TestingAssistant {
  private options = {
    systemPrompt:
      'You are a testing expert who writes comprehensive tests and fixes failures',
    allowedTools: ['Read', 'Write', 'Edit', 'Bash'],
    permissionMode: 'acceptAll' as const,
    maxTurns: 20,
  };

  async generateAndRunTests(modulePath: string): Promise<void> {
    console.log(`Generating tests for ${modulePath}`);

    for await (const message of query({
      prompt: `
        Analyze the module at ${modulePath} and:
        1. Write comprehensive test cases using Jest
        2. Cover edge cases and error conditions
        3. Run the tests and fix any failures
        4. Ensure good test coverage
      `,
      options: this.options,
    })) {
      if (message.type === 'tool_use' && message.toolName === 'Bash') {
        if (
          message.input.command?.includes('npm test') ||
          message.input.command?.includes('jest')
        ) {
          console.log('Running tests:', message.input.command);
        }
      }

      if (message.type === 'result') {
        console.log('Testing complete:', message.result);
        break;
      }
    }
  }

  async fixFailingTests(): Promise<void> {
    // Check for failing tests
    try {
      await execAsync('npm test');
      console.log('All tests passing!');
    } catch (error) {
      console.log('Tests failing, asking Claude to fix...');

      for await (const message of query({
        prompt:
          'The tests are failing. Please analyze the failures and fix them.',
        options: this.options,
      })) {
        if (message.type === 'result') {
          console.log('Test fixes applied:', message.result);
          break;
        }
      }
    }
  }
}

// Usage
const testAssistant = new TestingAssistant();
await testAssistant.generateAndRunTests('src/auth.ts');
await testAssistant.fixFailingTests();
```

### Documentation Generator

```typescript
import { query } from '@anthropic-ai/claude-code';
import { writeFileSync } from 'fs';

interface DocSection {
  title: string;
  content: string;
  level: number;
}

class DocumentationGenerator {
  private options = {
    systemPrompt: `
      You are a technical writer who creates clear, comprehensive documentation.
      Generate documentation that includes:
      1. Overview and purpose
      2. Installation instructions
      3. API reference
      4. Usage examples
      5. Best practices
    `,
    allowedTools: ['Read', 'Glob', 'Grep'],
    maxTurns: 8,
  };

  async generateProjectDocs(): Promise<string> {
    let documentation = '';

    for await (const message of query({
      prompt: 'Generate comprehensive documentation for this project',
      options: this.options,
    })) {
      if (message.type === 'result') {
        documentation = message.result;
        break;
      }
    }

    return documentation;
  }

  async generateAPIReference(srcPattern: string): Promise<DocSection[]> {
    const sections: DocSection[] = [];

    for await (const message of query({
      prompt: `Generate API reference documentation for files matching: ${srcPattern}`,
      options: {
        ...this.options,
        systemPrompt:
          'Generate structured API documentation with TypeScript interfaces and examples',
      },
    })) {
      if (message.type === 'result') {
        // Parse the result into sections
        const lines = message.result.split('\n');
        let currentSection: DocSection | null = null;

        for (const line of lines) {
          const headerMatch = line.match(/^(#{1,6})\s+(.+)$/);
          if (headerMatch) {
            if (currentSection) {
              sections.push(currentSection);
            }
            currentSection = {
              title: headerMatch[2],
              content: '',
              level: headerMatch[1].length,
            };
          } else if (currentSection) {
            currentSection.content += line + '\n';
          }
        }

        if (currentSection) {
          sections.push(currentSection);
        }
        break;
      }
    }

    return sections;
  }

  async saveDocs(content: string, filename: string): Promise<void> {
    writeFileSync(filename, content, 'utf-8');
    console.log(`Documentation saved to ${filename}`);
  }
}

// Usage
const docGen = new DocumentationGenerator();
const projectDocs = await docGen.generateProjectDocs();
await docGen.saveDocs(projectDocs, 'README.md');

const apiSections = await docGen.generateAPIReference('src/**/*.ts');
const apiDocs = apiSections
  .map(
    (section) =>
      `${'#'.repeat(section.level)} ${section.title}\n\n${section.content}`
  )
  .join('\n');
await docGen.saveDocs(apiDocs, 'API.md');
```

### CI/CD Integration

```typescript
import { query } from '@anthropic-ai/claude-code';

interface CIResult {
  success: boolean;
  summary: string;
  details: string[];
}

class CIAssistant {
  private options = {
    systemPrompt:
      'You are a CI/CD expert who ensures code quality and deployment readiness',
    allowedTools: ['Read', 'Bash', 'Grep'],
    maxTurns: 10,
    permissionMode: 'acceptAll' as const,
  };

  async runPreCommitChecks(): Promise<CIResult> {
    const details: string[] = [];
    let success = true;

    for await (const message of query({
      prompt: `
        Run pre-commit checks:
        1. Check code formatting (prettier/eslint)
        2. Run type checking (tsc)
        3. Run unit tests
        4. Check for security vulnerabilities
        5. Verify build succeeds
      `,
      options: this.options,
    })) {
      if (message.type === 'tool_use' && message.toolName === 'Bash') {
        details.push(`Running: ${message.input.command}`);
      }

      if (message.type === 'error') {
        success = false;
        details.push(`Error: ${message.error}`);
      }

      if (message.type === 'result') {
        const summary = message.result;
        return {
          success,
          summary,
          details,
        };
      }
    }

    return {
      success: false,
      summary: 'Pre-commit checks failed to complete',
      details,
    };
  }

  async analyzeDeploymentReadiness(): Promise<boolean> {
    for await (const message of query({
      prompt: `
        Analyze if this codebase is ready for deployment:
        1. All tests passing
        2. No critical security issues
        3. Performance benchmarks met
        4. Documentation up to date
      `,
      options: this.options,
    })) {
      if (message.type === 'result') {
        return message.result.toLowerCase().includes('ready for deployment');
      }
    }

    return false;
  }
}

// Usage in CI pipeline
const ciAssistant = new CIAssistant();

// Pre-commit hooks
const checkResult = await ciAssistant.runPreCommitChecks();
if (!checkResult.success) {
  console.error('Pre-commit checks failed:', checkResult.summary);
  process.exit(1);
}

// Deployment readiness
const isReady = await ciAssistant.analyzeDeploymentReadiness();
if (!isReady) {
  console.error('Code not ready for deployment');
  process.exit(1);
}

console.log('All checks passed - ready for deployment!');
```

## Error Handling

### Basic Error Handling

```typescript
import { query } from '@anthropic-ai/claude-code';

async function robustQuery(prompt: string) {
  try {
    for await (const message of query(prompt)) {
      if (message.type === 'error') {
        throw new Error(`Claude error: ${message.error}`);
      }

      if (message.type === 'result') {
        return message.result;
      }
    }
  } catch (error) {
    console.error('Query failed:', error);
    throw error;
  }
}
```

### Timeout and Cancellation

```typescript
import { query } from '@anthropic-ai/claude-code';

async function queryWithTimeout(prompt: string, timeoutMs: number) {
  const abortController = new AbortController();

  // Set timeout
  const timeout = setTimeout(() => {
    abortController.abort();
  }, timeoutMs);

  try {
    for await (const message of query({
      prompt,
      options: {
        abortController,
        timeout: timeoutMs / 1000, // Convert to seconds
      },
    })) {
      if (message.type === 'result') {
        clearTimeout(timeout);
        return message.result;
      }
    }
  } catch (error) {
    clearTimeout(timeout);
    if (abortController.signal.aborted) {
      throw new Error(`Query timed out after ${timeoutMs}ms`);
    }
    throw error;
  }
}

// Usage
try {
  const result = await queryWithTimeout('Analyze large codebase', 30000);
  console.log(result);
} catch (error) {
  console.error('Query failed or timed out:', error);
}
```

### Retry Logic

```typescript
import { query } from '@anthropic-ai/claude-code';

async function retryQuery(prompt: string, maxRetries = 3): Promise<string> {
  let lastError: Error;

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      for await (const message of query(prompt)) {
        if (message.type === 'result') {
          return message.result;
        }

        if (message.type === 'error') {
          throw new Error(message.error);
        }
      }
    } catch (error) {
      lastError = error as Error;
      console.warn(`Attempt ${attempt} failed:`, error);

      if (attempt < maxRetries) {
        // Exponential backoff
        await new Promise((resolve) =>
          setTimeout(resolve, 1000 * Math.pow(2, attempt - 1))
        );
      }
    }
  }

  throw new Error(
    `All ${maxRetries} attempts failed. Last error: ${lastError!.message}`
  );
}
```

## Integration Examples

### Express.js API

```typescript
import express from 'express';
import { query } from '@anthropic-ai/claude-code';

const app = express();
app.use(express.json());

app.post('/api/analyze', async (req, res) => {
  try {
    const { prompt, options } = req.body;

    for await (const message of query({ prompt, options })) {
      if (message.type === 'result') {
        res.json({
          success: true,
          result: message.result,
          sessionId: message.sessionId,
        });
        return;
      }
    }

    res.status(500).json({ success: false, error: 'No result received' });
  } catch (error) {
    res.status(500).json({ success: false, error: error.message });
  }
});

app.listen(3000, () => {
  console.log('API server running on port 3000');
});
```

### Next.js API Route

```typescript
// pages/api/code-review.ts
import { NextApiRequest, NextApiResponse } from 'next';
import { query } from '@anthropic-ai/claude-code';

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  if (req.method !== 'POST') {
    return res.status(405).json({ error: 'Method not allowed' });
  }

  const { codeFile } = req.body;

  try {
    for await (const message of query({
      prompt: `Review this code file for issues: ${codeFile}`,
      options: {
        allowedTools: ['Read', 'Grep'],
        maxTurns: 3,
      },
    })) {
      if (message.type === 'result') {
        return res.json({ review: message.result });
      }
    }
  } catch (error) {
    return res.status(500).json({ error: error.message });
  }
}
```

### WebSocket Real-time Streaming

```typescript
import WebSocket from 'ws';
import { query } from '@anthropic-ai/claude-code';

const wss = new WebSocket.Server({ port: 8080 });

wss.on('connection', (ws) => {
  ws.on('message', async (data) => {
    try {
      const { prompt, options } = JSON.parse(data.toString());

      for await (const message of query({
        prompt,
        options: { ...options, stream: true },
      })) {
        ws.send(JSON.stringify(message));

        if (message.type === 'result') {
          break;
        }
      }
    } catch (error) {
      ws.send(JSON.stringify({ type: 'error', error: error.message }));
    }
  });
});

console.log('WebSocket server running on port 8080');
```

## Best Practices

### Type Safety

```typescript
import { query, ClaudeMessage } from '@anthropic-ai/claude-code';

// Type-safe message handling
function handleMessage(message: ClaudeMessage): void {
  switch (message.type) {
    case 'result':
      console.log('Result:', message.result);
      console.log('Session:', message.sessionId);
      break;

    case 'thinking':
      console.log('Thinking:', message.content);
      break;

    case 'tool_use':
      console.log(`Tool: ${message.toolName}`, message.input);
      break;

    case 'error':
      console.error('Error:', message.error);
      if (message.code) {
        console.error('Code:', message.code);
      }
      break;

    default:
      // TypeScript will catch unhandled message types
      const _exhaustive: never = message;
  }
}
```

### Resource Management

```typescript
import { query } from '@anthropic-ai/claude-code';

class ClaudeService {
  private activeQueries = new Set<AbortController>();

  async performQuery(prompt: string): Promise<string> {
    const abortController = new AbortController();
    this.activeQueries.add(abortController);

    try {
      for await (const message of query({
        prompt,
        options: { abortController },
      })) {
        if (message.type === 'result') {
          return message.result;
        }
      }
      throw new Error('No result received');
    } finally {
      this.activeQueries.delete(abortController);
    }
  }

  cancelAllQueries(): void {
    for (const controller of this.activeQueries) {
      controller.abort();
    }
    this.activeQueries.clear();
  }
}

// Usage with proper cleanup
const claudeService = new ClaudeService();

process.on('SIGINT', () => {
  claudeService.cancelAllQueries();
  process.exit(0);
});
```

### Performance Optimization

```typescript
import { query } from '@anthropic-ai/claude-code';

// Batch operations for efficiency
async function batchAnalysis(files: string[]): Promise<string[]> {
  const results: string[] = [];

  // Process files in smaller batches to avoid overwhelming the system
  const batchSize = 5;
  for (let i = 0; i < files.length; i += batchSize) {
    const batch = files.slice(i, i + batchSize);

    const batchPromises = batch.map(async (file) => {
      for await (const message of query({
        prompt: `Analyze file: ${file}`,
        options: {
          allowedTools: ['Read'],
          maxTurns: 2,
        },
      })) {
        if (message.type === 'result') {
          return message.result;
        }
      }
      return '';
    });

    const batchResults = await Promise.all(batchPromises);
    results.push(...batchResults);
  }

  return results;
}
```

This comprehensive guide covers the TypeScript SDK's capabilities for building sophisticated AI agents with full type safety and modern JavaScript/TypeScript patterns.
