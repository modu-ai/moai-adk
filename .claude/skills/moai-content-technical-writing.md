# moai-content-technical-writing

Writing technical content with clarity, accuracy, and engagement for engineering audiences.

## Quick Start

Technical writing requires precision, clarity, and understanding of audience expertise. Use this skill when writing API documentation, how-to guides, tutorials, or technical blog posts.

## Core Patterns

### Pattern 1: Structure & Organization

**Pattern**: Organize technical content using clear hierarchies and logical flow.

```markdown
# Main Topic
One-sentence summary of what readers will learn.

## Overview
- What is this?
- Why does it matter?
- Who should read this?

## Concepts
### Concept 1
Define and explain with examples.

### Concept 2
Build on previous concepts.

## Practical Implementation
### Step 1: Setup
Commands and configuration.

### Step 2: Usage
Real code examples.

### Step 3: Troubleshooting
Common issues and solutions.

## Best Practices
- Guideline 1
- Guideline 2

## Common Pitfalls
- Mistake 1: How to avoid
- Mistake 2: How to avoid

## References
- Official docs
- Related articles
```

**When to use**:
- Writing tutorials and guides
- Creating API documentation
- Writing blog posts
- Building knowledge bases

**Key benefits**:
- Easy to follow
- Scannable structure
- Complete coverage
- Good for learning

### Pattern 2: Code Examples

**Pattern**: Write clear, runnable code examples with explanations.

```typescript
// Good: Complete, runnable example with comments
// File: src/api-client.ts

import axios from 'axios';

const client = axios.create({
  baseURL: 'https://api.example.com',
  timeout: 5000,
});

/**
 * Fetch user by ID
 * @param userId - The unique user identifier
 * @returns Promise containing user data
 * @throws Error if user not found or network fails
 */
export async function getUser(userId: string) {
  try {
    const response = await client.get(`/users/${userId}`);
    return response.data;
  } catch (error) {
    // Handle specific error types
    if (error.response?.status === 404) {
      throw new Error(`User ${userId} not found`);
    }
    throw new Error(`Failed to fetch user: ${error.message}`);
  }
}

// Usage example
const user = await getUser('12345');
console.log(user.name); // "Alice Johnson"
```

**Key practices**:
- Complete and runnable
- Clear variable names
- Error handling
- Inline comments for complex logic
- Usage examples

### Pattern 3: Effective Explanations

**Pattern**: Explain concepts at multiple levels.

```markdown
## Asynchronous Programming

### Simple Explanation
Code that doesn't wait for operations to complete before moving to the next line.

### Intermediate Explanation
When a function calls an operation (like fetching data), JavaScript doesn't wait for
the result. Instead, it continues executing the next line. When the operation completes,
a callback or Promise handles the result.

### Technical Deep Dive

Async/await is syntactic sugar over Promises. When a function is declared with `async`,
it automatically returns a Promise. The `await` keyword pauses execution until the Promise
resolves.

\`\`\`typescript
// Without async/await (callback-based)
fetch('/data')
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error(error));

// With async/await (cleaner)
async function loadData() {
  try {
    const response = await fetch('/data');
    const data = await response.json();
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}
\`\`\`

### Why This Matters
Non-blocking operations allow applications to handle multiple requests simultaneously,
improving responsiveness and scalability.
```

**When to use**:
- Explaining complex concepts
- Writing for mixed audiences
- Creating reference documentation
- Building learning materials

**Key benefits**:
- Accessible to different levels
- Builds understanding gradually
- Comprehensive coverage
- Clear progression

## Progressive Disclosure

### Level 1: Basic
- Clear headings and organization
- Simple, direct language
- Basic code examples
- Essential information only

### Level 2: Advanced
- Multiple explanation levels
- Detailed code examples
- Error handling discussion
- Performance considerations

### Level 3: Expert
- Advanced patterns and trade-offs
- Edge cases and gotchas
- Performance benchmarks
- Production deployment

## Works Well With

- **Markdown**: Format for writing
- **MDX**: Add interactive components
- **Typedoc**: Auto-generate API docs
- **Code highlighting**: Syntax emphasis
- **Diagrams**: Visual explanations

## References

- **Google Technical Writing**: https://developers.google.com/tech-writing
- **Apple Style Guide**: https://help.apple.com/applestyleguide/
- **Microsoft Writing Style Guide**: https://docs.microsoft.com/en-us/style-guide/welcome/
- **Technical Writing Handbook**: https://jacobian.org/writing/technical-style/
