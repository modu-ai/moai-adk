# moai-lang-javascript - Working Examples

_Last updated: 2025-10-22_

## Example 1: Project Setup with npm & Jest

```bash
# Initialize new project
npm init -y

# Install Jest as dev dependency
npm install --save-dev jest @types/jest

# Install ESLint and Prettier
npm install --save-dev eslint prettier eslint-config-prettier eslint-plugin-jest

# Initialize ESLint configuration
npx eslint --init

# Add test script to package.json
npm pkg set scripts.test="jest"
npm pkg set scripts.lint="eslint ."
npm pkg set scripts.format="prettier --write ."
```

**package.json configuration**:
```json
{
  "name": "my-project",
  "scripts": {
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage",
    "lint": "eslint .",
    "format": "prettier --write ."
  },
  "devDependencies": {
    "jest": "^29.7.0",
    "eslint": "^9.16.0",
    "prettier": "^3.4.1",
    "eslint-plugin-jest": "^28.0.0"
  }
}
```

## Example 2: TDD Workflow with Jest

**RED: Write failing test**
```javascript
// src/calculator.test.js
describe('Calculator', () => {
  describe('add', () => {
    it('should add two positive numbers', () => {
      const calculator = new Calculator();
      expect(calculator.add(2, 3)).toBe(5);
    });

    it('should handle negative numbers', () => {
      const calculator = new Calculator();
      expect(calculator.add(-1, -2)).toBe(-3);
    });

    it('should handle zero', () => {
      const calculator = new Calculator();
      expect(calculator.add(0, 5)).toBe(5);
    });
  });
});
```

**GREEN: Implement feature**
```javascript
// src/calculator.js
class Calculator {
  add(a, b) {
    return a + b;
  }
}

module.exports = Calculator;
```

**REFACTOR: Improve code quality**
```javascript
// src/calculator.js
/**
 * Calculator class providing basic arithmetic operations
 * @class
 */
class Calculator {
  /**
   * Adds two numbers
   * @param {number} a - First operand
   * @param {number} b - Second operand
   * @returns {number} Sum of a and b
   * @throws {TypeError} If arguments are not numbers
   */
  add(a, b) {
    if (typeof a !== 'number' || typeof b !== 'number') {
      throw new TypeError('Both arguments must be numbers');
    }
    return a + b;
  }
}

module.exports = Calculator;
```

## Example 3: ESLint Configuration with Jest Support

**.eslintrc.json**:
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
    "no-unused-vars": "error",
    "no-console": "warn",
    "prefer-const": "error",
    "jest/expect-expect": "error",
    "jest/no-disabled-tests": "warn"
  }
}
```

**Run linting**:
```bash
# Check all files
npm run lint

# Fix auto-fixable issues
npx eslint . --fix
```

## Example 4: Jest Configuration with Coverage

**jest.config.js**:
```javascript
module.exports = {
  testEnvironment: 'node',
  coverageDirectory: 'coverage',
  collectCoverageFrom: [
    'src/**/*.js',
    '!src/**/*.test.js',
    '!src/**/*.spec.js'
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
  ]
};
```

**Run tests with coverage**:
```bash
# Run all tests
npm test

# Watch mode during development
npm run test:watch

# Generate coverage report
npm run test:coverage
```

**Expected output**:
```
PASS  src/calculator.test.js
  Calculator
    add
      ✓ should add two positive numbers (2 ms)
      ✓ should handle negative numbers (1 ms)
      ✓ should handle zero

----------|---------|----------|---------|---------|-------------------
File      | % Stmts | % Branch | % Funcs | % Lines | Uncovered Line #s
----------|---------|----------|---------|---------|-------------------
All files |     100 |      100 |     100 |     100 |
 calculator.js |     100 |      100 |     100 |     100 |
----------|---------|----------|---------|---------|-------------------

Test Suites: 1 passed, 1 total
Tests:       3 passed, 3 total
```

## Example 5: Async Testing with Jest

```javascript
// src/api.test.js
describe('API Client', () => {
  describe('fetchUser', () => {
    it('should fetch user data successfully', async () => {
      const api = new ApiClient();
      const user = await api.fetchUser(1);

      expect(user).toHaveProperty('id', 1);
      expect(user).toHaveProperty('name');
      expect(user).toHaveProperty('email');
    });

    it('should handle errors gracefully', async () => {
      const api = new ApiClient();

      await expect(api.fetchUser(-1))
        .rejects
        .toThrow('Invalid user ID');
    });

    it('should timeout after 5 seconds', async () => {
      const api = new ApiClient({ timeout: 5000 });

      jest.setTimeout(6000);
      await expect(api.fetchUser(999))
        .rejects
        .toThrow('Request timeout');
    });
  });
});
```

## Example 6: Mocking with Jest

```javascript
// src/userService.test.js
const UserService = require('./userService');
const database = require('./database');

// Mock the database module
jest.mock('./database');

describe('UserService', () => {
  beforeEach(() => {
    // Clear all mocks before each test
    jest.clearAllMocks();
  });

  it('should create user with hashed password', async () => {
    const mockUser = { id: 1, username: 'john', password: 'hashed123' };
    database.insert.mockResolvedValue(mockUser);

    const service = new UserService();
    const result = await service.createUser('john', 'plain123');

    expect(database.insert).toHaveBeenCalledTimes(1);
    expect(result.password).not.toBe('plain123');
    expect(result).toMatchObject({ id: 1, username: 'john' });
  });
});
```

---

_For complete API reference and configuration options, see reference.md_
