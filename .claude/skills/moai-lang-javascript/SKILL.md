---
name: moai-lang-javascript
description: JavaScript best practices with Jest, ESLint, Prettier, and npm package management
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# JavaScript Expert

## What it does

Provides JavaScript-specific expertise for TDD development, including Jest testing, ESLint linting, Prettier formatting, and npm package management.

## When to use

- "JavaScript 테스트 작성", "Jest 사용법", "ES6+ 문법", "웹 개발", "Node.js 백엔드", "프론트엔드"
- "DOM", "AJAX", "Promise", "async/await", "Fetch API"
- "jQuery", "Axios", "Lodash", "D3.js", "Three.js"
- Automatically invoked when working with JavaScript projects
- JavaScript SPEC implementation (`/alfred:2-build`)

## How it works

**TDD Framework**:
- **Jest**: Unit testing with mocking, snapshots
- **@testing-library**: DOM/React testing
- Test coverage ≥85% enforcement

**Code Quality**:
- **ESLint**: JavaScript linting with recommended rules
- **Prettier**: Code formatting (opinionated)
- **JSDoc**: Type hints via comments (for type safety)

**Package Management**:
- **npm**: Standard package manager
- **package.json** for dependencies and scripts
- Semantic versioning

**Modern JavaScript**:
- ES6+ features (arrow functions, destructuring, spread/rest)
- Async/await over callbacks
- Module imports (ESM) over CommonJS

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer `const` over `let`, avoid `var`
- Guard clauses for early returns
- Meaningful names, avoid abbreviations

## Modern JavaScript (ES2022+)

**Recommended Version**: Node.js 20+ LTS, 18+ for stable features

**Modern Features**:
- **Top-level await** (ES2022): Async imports
- **Class fields & private members** (ES2022): `#privateField`
- **Logical assignment** (ES2021): `??=`, `&&=`, `||=`
- **Optional chaining** (ES2020): `obj?.prop?.nested`
- **Nullish coalescing** (ES2020): `x ?? default`
- **BigInt** (ES2020): Arbitrary precision integers

**Version Check**:
```bash
node --version  # Check Node.js version
npm --version
```

## Package Management Commands

### Using npm
```bash
# Initialize
npm init -y
npm create vite@latest my-app

# Install
npm install
npm install express axios
npm install --save-dev jest @types/jest

# Install specific version
npm install react@18.2.0

# Update
npm update
npm update --latest

# Remove
npm uninstall express

# Run
npm start
npm run test
npm run dev
npm run build

# Check security
npm audit
npm audit fix
```

### Using yarn
```bash
yarn init
yarn add express axios
yarn add --dev jest

# Run
yarn start
yarn test
yarn dev
```

### Using pnpm
```bash
pnpm init
pnpm add express axios
pnpm add -D jest

pnpm run test
pnpm run dev
```

## Examples

### Example 1: TDD with Jest
User: "/alfred:2-build API-001"
Claude: (creates RED test with Jest, GREEN implementation, REFACTOR with JSDoc)

### Example 2: Linting
User: "ESLint 실행"
Claude: (runs eslint . and reports linting errors)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (JavaScript-specific review)
- alfred-debugger-pro (JavaScript debugging)
