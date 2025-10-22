# moai-lang-javascript - CLI Reference

_Last updated: 2025-10-22_

## Official Documentation Links

- **MDN Web Docs**: https://developer.mozilla.org/en-US/docs/Web/JavaScript
- **Node.js**: https://nodejs.org/docs/latest/api/
- **Jest**: https://jestjs.io/docs/getting-started
- **ESLint**: https://eslint.org/docs/latest/
- **Prettier**: https://prettier.io/docs/en/
- **npm**: https://docs.npmjs.com/

## Tool Versions (2025-10-22)

| Tool | Version | Release Date | Status |
|------|---------|--------------|--------|
| **Node.js** | 22.11.0 | 2024-10 | ✅ LTS |
| **Jest** | 29.7.0 | 2023-09 | ✅ Stable |
| **ESLint** | 9.16.0 | 2024-12 | ✅ Latest |
| **Prettier** | 3.4.1 | 2024-12 | ✅ Latest |
| **npm** | 10.9.0 | 2024-09 | ✅ Stable |

## Installation

### Prerequisites

```bash
# Check Node.js installation
node --version  # Should be v22.x or higher

# Check npm installation
npm --version   # Should be v10.x or higher
```

### Project Initialization

```bash
# Create new project
mkdir my-project && cd my-project
npm init -y

# Install testing dependencies
npm install --save-dev jest @types/jest

# Install linting and formatting
npm install --save-dev eslint prettier eslint-config-prettier eslint-plugin-jest

# Install specific packages for your project type
npm install --save-dev @babel/preset-env  # For modern JS/ES modules
```

## Common Commands

### Testing with Jest

```bash
# Run all tests
npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with coverage
npm test -- --coverage

# Run specific test file
npm test -- calculator.test.js

# Run tests matching pattern
npm test -- --testNamePattern="add"

# Update snapshots
npm test -- -u

# Run tests in CI mode
npm test -- --ci --coverage --maxWorkers=2
```

### Linting with ESLint

```bash
# Lint all files
npx eslint .

# Lint specific file
npx eslint src/calculator.js

# Lint with auto-fix
npx eslint . --fix

# Lint only changed files (with git)
npx eslint $(git diff --name-only --diff-filter=ACM | grep '\.js$')

# Check for security issues
npm audit
npm audit fix
```

### Formatting with Prettier

```bash
# Check formatting
npx prettier --check .

# Format all files
npx prettier --write .

# Format specific file
npx prettier --write src/calculator.js

# Format only staged files (with git)
npx prettier --write $(git diff --cached --name-only --diff-filter=ACM)
```

### Package Management

```bash
# Install all dependencies
npm install

# Install specific package
npm install lodash

# Install dev dependency
npm install --save-dev @types/node

# Update dependencies
npm update

# Check for outdated packages
npm outdated

# Remove unused dependencies
npm prune

# View dependency tree
npm list

# Search for packages
npm search <keyword>
```

## Configuration Files

### package.json Scripts

```json
{
  "scripts": {
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage",
    "lint": "eslint .",
    "lint:fix": "eslint . --fix",
    "format": "prettier --write .",
    "format:check": "prettier --check .",
    "validate": "npm run lint && npm run format:check && npm test"
  }
}
```

### .eslintrc.json

```json
{
  "env": {
    "node": true,
    "es2024": true,
    "jest": true
  },
  "extends": [
    "eslint:recommended",
    "plugin:jest/recommended",
    "prettier"
  ],
  "plugins": ["jest"],
  "parserOptions": {
    "ecmaVersion": "latest",
    "sourceType": "module"
  },
  "rules": {
    "no-unused-vars": ["error", { "argsIgnorePattern": "^_" }],
    "no-console": ["warn", { "allow": ["warn", "error"] }],
    "prefer-const": "error",
    "no-var": "error",
    "eqeqeq": ["error", "always"],
    "curly": ["error", "all"]
  }
}
```

### .prettierrc.json

```json
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 100,
  "tabWidth": 2,
  "useTabs": false,
  "arrowParens": "always",
  "endOfLine": "lf"
}
```

### jest.config.js

```javascript
module.exports = {
  testEnvironment: 'node',
  coverageDirectory: 'coverage',
  collectCoverageFrom: [
    'src/**/*.js',
    '!src/**/*.test.js',
    '!src/**/*.spec.js',
    '!src/**/index.js'
  ],
  coverageThreshold: {
    global: {
      branches: 85,
      functions: 85,
      lines: 85,
      statements: 85
    }
  },
  testMatch: [
    '**/__tests__/**/*.js',
    '**/?(*.)+(spec|test).js'
  ],
  moduleFileExtensions: ['js', 'json', 'node'],
  verbose: true
};
```

## Jest Matchers Quick Reference

### Basic Matchers
- `expect(value).toBe(expected)` - Strict equality (===)
- `expect(value).toEqual(expected)` - Deep equality
- `expect(value).toBeNull()` - Value is null
- `expect(value).toBeUndefined()` - Value is undefined
- `expect(value).toBeDefined()` - Value is not undefined
- `expect(value).toBeTruthy()` - Value is truthy
- `expect(value).toBeFalsy()` - Value is falsy

### Number Matchers
- `expect(value).toBeGreaterThan(number)`
- `expect(value).toBeGreaterThanOrEqual(number)`
- `expect(value).toBeLessThan(number)`
- `expect(value).toBeLessThanOrEqual(number)`
- `expect(value).toBeCloseTo(number, precision)`

### String Matchers
- `expect(string).toMatch(regex)`
- `expect(string).toContain(substring)`

### Array/Object Matchers
- `expect(array).toContain(item)`
- `expect(array).toHaveLength(number)`
- `expect(object).toHaveProperty(keyPath, value)`
- `expect(object).toMatchObject(partial)`

### Exception Matchers
- `expect(fn).toThrow()`
- `expect(fn).toThrow(Error)`
- `expect(fn).toThrow('error message')`

### Async Matchers
- `await expect(promise).resolves.toBe(value)`
- `await expect(promise).rejects.toThrow()`

## Best Practices

### Code Organization
✅ Use ES6+ features (const, let, arrow functions, destructuring)
✅ Follow single responsibility principle
✅ Keep functions small (<50 lines)
✅ Use meaningful variable and function names
✅ Avoid deeply nested code (max 3 levels)

### Testing
✅ Write tests before implementation (TDD)
✅ Use descriptive test names
✅ Follow AAA pattern (Arrange, Act, Assert)
✅ Test edge cases and error conditions
✅ Maintain ≥85% code coverage
✅ Use `beforeEach`/`afterEach` for setup/cleanup
✅ Mock external dependencies

### Error Handling
✅ Use try-catch for synchronous code
✅ Use async/await with try-catch for async code
✅ Throw specific error types
✅ Provide meaningful error messages
✅ Validate inputs early

### Security
✅ Run `npm audit` regularly
✅ Keep dependencies updated
✅ Use environment variables for secrets
✅ Validate and sanitize user inputs
✅ Use HTTPS for external requests

## Common Issues & Solutions

### Issue: Tests failing in CI but passing locally
**Solution**: Ensure consistent Node.js version, run `npm ci` instead of `npm install`

### Issue: ESLint conflicts with Prettier
**Solution**: Install `eslint-config-prettier` and add "prettier" to extends array

### Issue: Jest not finding modules
**Solution**: Check `moduleNameMapper` in jest.config.js, ensure correct paths

### Issue: Coverage threshold not met
**Solution**: Add more tests, check `collectCoverageFrom` configuration

### Issue: Slow test execution
**Solution**: Use `--maxWorkers` flag, avoid unnecessary mocks, use `--onlyChanged`

---

_For working examples and use cases, see examples.md_
