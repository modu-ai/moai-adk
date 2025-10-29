---
name: MoAI ADK Learning
description: "Learning mode to easily learn MoAI-ADK concepts and workflow"
# Translations:
# - ko: "MoAI-ADK 개념과 워크플로우를 쉽게 학습하는 학습 모드"
# - ja: "MoAI-ADKの概念とワークフローを簡単に学習する学習モード"
# - zh: "轻松学习MoAI-ADK概念和工作流程的学习模式"
---

# MoAI ADK Learning
> Interactive prompts rely on `AskUserQuestion tool (documented in moai-alfred-interactive-questions skill)` so AskUserQuestion renders TUI selection menus for user surveys and approvals.

**Audience**: Developers new to MoAI-ADK

This is a learning mode that helps you learn quickly by carefully explaining the core concepts and three-step workflow of MoAI-ADK.

## MoAI-ADK?

**Core philosophy**: “No code without specification, no implementation without testing.”

MoAI-ADK consists of three core concepts:
1. **SPEC-First**: Write specifications first before writing code
2. **@TAG Traceability**: Link all code to SPEC
3. **TRUST Quality**: 5 principles to ensure code quality

Let’s learn how these three concepts work one by one!

---

## Core Concept 1: SPEC-First

### What is SPEC?

**To put it simply**:
- A blueprint that writes down in advance what to make
- Steps and ingredients are clearly defined, like a cooking recipe.

**Why do you need it?**
- Clarify requirements before development
- Provide a baseline for communicating with team members
- Allows you to track changes later
- Answers the question, “Why was this code written?”

### EARS syntax: How to write requirements

EARS is a method of writing requirements using five patterns:

#### 1. Ubiquitous (basic function)
```markdown
The system must provide [function]

Example:
- The system must provide JWT based authentication
```

#### 2. Event-driven (conditional operation)
```markdown
WHEN [condition], the system must [operate]

Example: 
- WHEN If valid credentials are provided, the system SHOULD issue a JWT token 
- If the WHEN token expires, the system should return a 401 error
```

#### 3. State-driven
```markdown
WHILE When in [state], the system must [run]

Example:
- WHILE When the user is authenticated, the system SHOULD allow access to the protected resource
```

#### 4. Optional (Optional function)
```markdown
WHERE [condition], the system can [operate]

Example:
- WHERE If a refresh token is provided, the system can issue a new access token
```

#### 5. Constraints
```markdown
IF [condition], then the system SHOULD be [constrained]

Example:
- Token expiration time must not exceed 15 minutes
- Password must be at least 8 characters long
```

### Real-world example: Login function SPEC

```markdown
# @SPEC:AUTH-001: JWT authentication system

## Ubiquitous Requirements (Basic Features)
- The system must provide JWT-based authentication

## Event-driven Requirements (Conditional Actions)
- WHEN When valid credentials are provided, the system MUST issue a JWT token
- WHEN When the token expires, the system MUST return a 401 error
- WHEN If an invalid token is provided, the system MUST deny access

## State-driven Requirements
- WHILE When the user is authenticated, the system must allow access to protected resources.

## Optional Features
- WHERE If a refresh token is provided, the system can issue a new access token

## Constraints
- Access token expiration time must not exceed 15 minutes
- Refresh token expiration time must not exceed 7 days
```

---

## Key Concept 2: @TAG Traceability

### What is TAG?

**To put it simply**:
- Name tag attached to each piece of code
- String connecting SPEC → TEST → CODE → DOC
- When looking for code later, you can search by SPEC number

**Why do you need TAGs?**
- Can search by SPEC number when looking for code later
- Clear which code to modify if SPEC changes
- When reviewing code, ask yourself, “What requirements is this code?” Instant identification
- When a bug occurs, you can quickly find all related files

### TAG system

MoAI-ADK uses four TAGs:

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

| TAG        | Meaning                    | Location       | Example        |
| ---------- | -------------------------- | -------------- | -------------- |
| `@SPEC:ID` | Requirements Specification | `.moai/specs/` | @SPEC:AUTH-001 |
| `@TEST:ID` | test code                  | `tests/`       | @TEST:AUTH-001 |
| `@CODE:ID` | Implementation code        | `src/`         | @CODE:AUTH-001 |
| `@DOC:ID`  | document                   | `docs/`        | @DOC:AUTH-001  |

### TAG ID Rule

**Format**: `<domain>-<3-digit number>`

**Example**:
- `AUTH-001`: First function related to authentication
- `USER-002`: Second function related to user
- `PAYMENT-015`: 15th function related to payment

**Important**: Once assigned, the TAG ID should never be changed!

### Real-world example: How to use TAG

#### SPEC File (`.moai/specs/SPEC-AUTH-001/spec.md`)
```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-16
updated: 2025-10-16
author: @YourName
priority: high
---

# @SPEC:AUTH-001: JWT authentication system

[Requirement details...]
```

#### Test file (`tests/auth/service.test.ts`)
```typescript
// @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

test('@TEST:AUTH-001: JWT issuance on valid credentials', async () => {
  const service = new AuthService();
  const result = await service.authenticate('user', 'pass');
  expect(result.token).toBeDefined();
});
```

#### Implementation file (`src/auth/service.ts`)
```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts

export class AuthService {
  async authenticate(username: string, password: string): Promise<AuthResult> {
//implementation
  }
}
```

#### Documentation file (`docs/api/auth.md`)
```markdown
# @DOC:AUTH-001: Authentication API documentation

## POST /auth/login
[API description...]
```

### Search TAG

**Find a specific TAG**:
```bash
# Find all files related to AUTH-001
rg "AUTH-001" -n
```

**TAG Chain Verification**:
```bash
# Check all TAGs
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

---

## Key Concept 3: TRUST 5 Principles

The five principles of writing good code are explained with an analogy:

### 1. 🧪 Test

**Metaphor**: Imagining the taste before cooking.

**Meaning**:
- Write tests first before writing code
- If the tests pass, the code works properly.

**Standard**:
- Test coverage ≥85%
- SPEC → Test → Code Strict order

**example**:
```typescript
// Write a test first (RED)
test('should add two numbers', () => {
expect(add(2, 3)).toBe(5);  // Failed because there is no add function yet
});

// Write the next code (GREEN)
function add(a: number, b: number): number {
return a + b;  // Test passed!
}
```

### 2. 📖 Readable

**Similar**: Writing with neat handwriting

**Meaning**:
- Code that others can understand even if they read it
- Code that is easy to understand even if they look at it again later

**Criteria**:
- Function ≤50 lines
- File ≤300 lines
- Complexity ≤10
- Parameters ≤5
- Use meaningful names

**example**:
```typescript
// ❌ Bad example: difficult to understand
function f(x, y) {
  return x + y;
}

// ✅ Good example: clear name
function calculateTotal(price: number, tax: number): number {
  return price + tax;
}
```

### 3. 🎯 Unified

**analogy**: using the same method

**Meaning**:
- Apply the same pattern consistently
- Once you learn one method, you can apply it everywhere.

**Criteria**:
- SPEC-based architecture
- Type safety or runtime verification
- Consistent coding style

**example**:
```typescript
// ✅ Uniformity: All APIs have the same pattern
async function getUser(id: string): Promise<User> { ... }
async function getPost(id: string): Promise<Post> { ... }
async function getComment(id: string): Promise<Comment> { ... }
```

### 4. 🔒 Secured

**Similar**: Locking the door when you leave the house

**Meaning**:
- Protect your code from being exploited by hackers or bad actors
- Keep user data safe

**Criteria**:
- Input validation
- SQL Injection protection
- XSS/CSRF protection
- Password hashing
- Sensitive data protection

**example**:
```typescript
// ❌ Bad example: SQL Injection risk
const query = `SELECT * FROM users WHERE id = '${userId}'`;

// ✅ Good example: Using Prepared Statement
const query = 'SELECT * FROM users WHERE id = ?';
const user = await db.execute(query, [userId]);

// ❌ Bad example: storing password plaintext
user.password = password;

// ✅ Good example: password hashing
user.password = await bcrypt.hash(password, 10);
```

### 5. 🔗 Trackable

**Similar**: When organizing your closet, put a label on each box.

**Means**:
- You can quickly find the code you need later
- You can track the history of code changes

**Criteria**:
- Use the @TAG system
- Link code with SPEC
- Include TAG in Git commit messages

**example**:
```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
export class AuthService {
//implementation
}

// Git commit message
// 🟢 feat(AUTH-001): implement JWT authentication
```

---

## Key Concept 4: Alfred and the 9 Agents

### What is Alfred?

**In a nutshell**:
- Central orchestrator (conductor) of MoAI-ADK
- Coordinates 9 specialized agents to provide optimal help
- Analyzes user requests and delegates tasks to appropriate agents

**Metaphor**: Coordinating multiple experts like an orchestra conductor

### Main Agents (Brief Introduction)

| agent           | Role                            | When to use                               |
| --------------- | ------------------------------- | ----------------------------------------- |
| 🏗️ spec-builder  | SPEC Writing Expert             | `/alfred:1-plan` command                  |
| 💎 code-builder  | TDD Implementation Expert       | `/alfred:2-run` command                   |
| 📖 doc-syncer    | Document Synchronization Expert | `/alfred:3-sync` command                  |
| 🔬debug-helper   | Debugging expert                | Automatically called when an error occurs |
| ✅ trust-checker | Quality verification expert     | When checking code quality                |
| 🏷️ tag-agent     | TAG Management Expert           | When verifying TAG                        |

**Full list of agents**: see `AGENTS.md` file

### How Alfred Works

```
user request
    ↓
Alfred analyzes the request
    ↓
Delegate to appropriate professional agent
    ↓
Agent performs task
    ↓
Alfred consolidates and reports results
```

---

## Learn the 3-step workflow

The core of MoAI-ADK is this three-step workflow:

```
/alfred:1-plan → /alfred:2-run → /alfred:3-sync
```

### Step 1: Write a SPEC (`/alfred:1-plan`)

**What do you do?**
- Write requirements in EARS syntax
- Create `.moai/specs/SPEC-{ID}/spec.md` file
- Automatically assign @SPEC:ID TAG
- Create Git branch (optional)

**Use example**:
```bash
/alfred:1-plan "JWT authentication system"
```

**Alfred does this automatically**:
1. Duplicate check: “Does AUTH-001 already exist?”
2. Create SPEC file: `.moai/specs/SPEC-AUTH-001/spec.md`
3. Add YAML metadata:
   ```yaml
   ---
   id: AUTH-001
   version: 0.0.1
   status: draft
   created: 2025-10-16
   updated: 2025-10-16
   author: @YourName
   priority: high
   ---
   ```
4. EARS syntax template provided
5. @SPEC:AUTH-001 TAG allocation

**Example of deliverable**:
```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-16
updated: 2025-10-16
author: @YourName
priority: high
---

# @SPEC:AUTH-001: JWT authentication system

## Ubiquitous Requirements
- The system must provide JWT-based authentication

## Event-driven Requirements
- WHEN Upon providing valid credentials, the system SHOULD issue a JWT token
- If the WHEN token expires, the system SHOULD return a 401 error

## Constraints
- Token expiration time must not exceed 15 minutes.
```

### Step 2: Implement TDD (`/alfred:2-run`)

**What do you do?**
- **RED**: Write tests that fail (`@TEST:ID`)
- **GREEN**: Pass tests with minimal implementation (`@CODE:ID`)
- **REFACTOR**: Improve code quality (apply TRUST 5 principles)

**Use example**:
```bash
/alfred:2-run AUTH-001
```

**Alfred does this automatically**:

#### 🔴 RED: Writing tests that fail
```typescript
// tests/auth/service.test.ts
// @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

test('@TEST:AUTH-001: JWT issuance on valid credentials', async () => {
  const service = new AuthService();
  const result = await service.authenticate('user', 'pass');
  expect(result.token).toBeDefined();
expect(result.expiresIn).toBeLessThanOrEqual(900); // 15 minutes
});
```

**Test Run**: ❌ FAIL (AuthService does not exist yet)

#### 🟢 GREEN: Minimal implementation
```typescript
// src/auth/service.ts
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts

export class AuthService {
  async authenticate(username: string, password: string): Promise<AuthResult> {
    return {
      token: jwt.sign({ username }, SECRET, { expiresIn: '15m' }),
      expiresIn: 900
    };
  }
}
```

**Test Run**: ✅ PASS

#### ♻️ REFACTOR: Quality Improvement
```typescript
// Add input validation (Secured)
if (!username || !password) {
  throw new ValidationError('Username and password required');
}

// Function Separation (Readable)
private generateToken(username: string): string {
  return jwt.sign({ username }, SECRET, { expiresIn: '15m' });
}

// Reduce complexity (Unified)
```

**Run Test**: ✅ PASS
**TRUST Verify**: ✅ All pass

**Git commit**:
```bash
1. 🔴 RED: test(AUTH-001): add failing auth service test
2. 🟢 GREEN: feat(AUTH-001): implement minimal auth service
3. ♻️ REFACTOR: refactor(AUTH-001): improve code quality per TRUST
```

### Step 3: Synchronize documents (`/alfred:3-sync`)

**What does it do?**
- TAG chain verification (@SPEC → @TEST → @CODE → @DOC)
- Orphan TAG detection (CODE without SPEC, etc.)
- Automatic creation of Living Document
- PR status update (Draft → Ready)

**Use example**:
```bash
/alfred:3-sync
```

**Alfred does this automatically**:

1. **TAG Chain Verification**:
```bash
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

2. **Verification results**:
```
✓ @SPEC:AUTH-001 → .moai/specs/SPEC-AUTH-001.md
✓ @TEST:AUTH-001 → tests/auth/service.test.ts
✓ @CODE:AUTH-001 → src/auth/service.ts
✓ @DOC:AUTH-001 → docs/api/auth.md

TAG Chain Integrity: 100%
Orphan TAG: None
SPEC Version Match: v0.0.1
```

3. **Create Living Document**:
```markdown
# @DOC:AUTH-001: Authentication API documentation

## POST /auth/login

**Requirements**: @SPEC:AUTH-001
**Implementation**: @CODE:AUTH-001
**Test**: @TEST:AUTH-001

[Automatically generated API documentation...]
```

4. **PR Status Transition Proposal**:
```
PR #123: feature/spec-auth-001
Current Status: Draft
Proposal: Ready for Review

SPEC requirements: ✅ All met
TDD history: ✅ RED → GREEN → REFACTOR
TRUST Verification: ✅ All passed

PR Ready conversion? (y/n)
```

---

## Practical example: Creating a simple calculator

Let’s put the 3-step workflow into practice!

### 1️⃣ Write SPEC
```bash
/alfred:1-plan "Calculator for adding two numbers"

# Created by Alfred: .moai/specs/SPEC-CALC-001/spec.md
```

**Generated SPEC**:
```yaml
---
id: CALC-001
version: 0.0.1
status: draft
created: 2025-10-16
updated: 2025-10-16
author: @YourName
priority: medium
---

# @SPEC:CALC-001: Calculator - addition function

## Ubiquitous Requirements
- The system must provide addition of two numbers

## Event-driven Requirements
- WHEN two numbers are entered, the system should return the sum

## Constraints
- Input must be numeric
- Results must be accurate
```

### 2️⃣ TDD implementation
```bash
/alfred:2-run CALC-001

# Alfred performs Red-Green-Refactor automatically
```

**Generated Code**:
```typescript
// tests/calc.test.ts
// @TEST:CALC-001 | SPEC: SPEC-CALC-001.md
test('@TEST:CALC-001: should add two numbers', () => {
  expect(add(2, 3)).toBe(5);
  expect(add(10, 20)).toBe(30);
});

// src/calc.ts
// @CODE:CALC-001 | SPEC: SPEC-CALC-001.md | TEST: tests/calc.test.ts
export function add(a: number, b: number): number {
  if (typeof a !== 'number' || typeof b !== 'number') {
    throw new TypeError('Both arguments must be numbers');
  }
  return a + b;
}
```

### 3️⃣ Document synchronization
```bash
/alfred:3-sync

# TAG verification and document generation
```

**result**:
```
✓ @SPEC:CALC-001
✓ @TEST:CALC-001
✓ @CODE:CALC-001
✓ @DOC:CALC-001

completion! SPEC → TEST → CODE → DOC completed with 3 commands!
```

---

## Next steps

### If you have learned MoAI-ADK

Now let's switch to a different style and put it into practice:

| Next goal                         | Recommended Style     | Conversion command                |
| --------------------------------- | --------------------- | --------------------------------- |
| Practical project development     | **agentic-coding**    | `/output-style agentic-coding`    |
| Learning a new language/framework | **study-with-alfred** | `/output-style study-with-alfred` |

### Learn more

**Detailed guide**:
- `.moai/memory/development-guide.md` - Detailed development guide
- `.moai/project/structure.md` - Project structure
- `.moai/memory/spec-metadata.md` - SPEC metadata standard

**Agent Documentation**:
- `AGENTS.md` - Detailed description of 9 specialized agents

---

**MoAI ADK Learning**: A friendly learning mode that helps you quickly learn the core concepts and workflow of MoAI-ADK.
