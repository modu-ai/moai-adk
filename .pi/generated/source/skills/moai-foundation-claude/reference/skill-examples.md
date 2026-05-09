# Claude Code Skills Examples Collection

Comprehensive collection of real-world skill examples covering various domains and complexity levels, all following official Claude Code standards.

Purpose: Practical examples and templates for skill creation
Target: Skill developers and Claude Code users
Last Updated: 2025-11-25
Version: 2.0.0

---

## Quick Reference (30 seconds)

Examples Cover: Documentation skills, language-specific patterns, domain expertise, integration patterns. Complexity Levels: Simple utilities, intermediate workflows, advanced orchestration. All Examples: Follow official formatting standards with proper frontmatter and progressive disclosure.

---

## Example Categories

### 1. Documentation and Analysis Skills

#### Example 1: API Documentation Generator

```yaml
---
name: moai-docs-api-generator
description: Generate comprehensive API documentation from OpenAPI specifications and code comments. Use when you need to create, update, or analyze API documentation for REST/GraphQL services.
allowed-tools: Read, Write, Edit, Grep, Glob, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
version: 1.2.0
tags: [documentation, api, openapi, graphql]
updated: 2025-11-25
status: active
---

# API Documentation Generator

Automated API documentation generation from OpenAPI specs, code comments, and GraphQL schemas with markdown output and interactive examples.

## Quick Reference (30 seconds)

Parse OpenAPI/GraphQL specifications and generate comprehensive documentation with examples, authentication guides, and interactive testing instructions.

## Implementation Guide

### Core Capabilities
- OpenAPI Processing: Parse and validate OpenAPI 3.0+ specifications
- GraphQL Schema Analysis: Extract documentation from GraphQL schemas
- Code Comment Extraction: Generate docs from JSDoc and docstring comments
- Interactive Examples: Create runnable code examples for each endpoint

### When to Use
- New API Projects: Generate initial documentation structure from specifications
- Documentation Updates: Sync existing docs with API changes
- API Reviews: Analyze API completeness and consistency
- Developer Portals: Create comprehensive API reference sites

### Essential Patterns
```python
# Parse OpenAPI specification
def parse_openapi_spec(spec_path):
 """Parse and validate OpenAPI 3.0+ specification."""
 with open(spec_path, 'r') as f:
 spec = yaml.safe_load(f)

 validate_openapi(spec)
 return spec

# Generate endpoint documentation
def generate_endpoint_docs(endpoint_spec):
 """Generate markdown documentation for API endpoint."""
 return f"""
 ## {endpoint_spec['method'].upper()} {endpoint_spec['path']}

 Description: {endpoint_spec.get('summary', 'No description')}

 Parameters:
 {format_parameters(endpoint_spec.get('parameters', []))}

 Response:
 {format_response(endpoint_spec.get('responses', {}))}
 """
```

```bash
# Generate documentation from codebase
find ./src -name "*.py" -exec grep -l "def.*api_" {} \; | \
xargs python extract_docs.py --output ./docs/api/
```

## Best Practices

 DO:
- Include authentication examples for all security schemes
- Provide curl and client library examples for each endpoint
- Validate all generated examples against actual API
- Include error response documentation

 DON'T:
- Generate documentation without example responses
- Skip authentication and authorization details
- Use deprecated OpenAPI specification versions
- Forget to document rate limiting and quotas

## Works Well With

- [`moai-docs-toolkit`](../moai-docs-toolkit/SKILL.md) - General documentation patterns
- [`moai-domain-backend`](../moai-domain-backend/SKILL.md) - Backend API expertise
- [`moai-context7-integration`](../moai-context7-integration/SKILL.md) - Latest framework docs

## Advanced Features

### Interactive Documentation
Generate interactive API documentation with embedded testing tools:
```html
<!-- Interactive API tester -->
<div class="api-tester">
 <input type="text" id="endpoint-url" placeholder="/api/users">
 <select id="http-method">
 <option value="GET">GET</option>
 <option value="POST">POST</option>
 </select>
 <button onclick="testEndpoint()">Test</button>
 <pre id="response"></pre>
</div>
```

### Multi-language Client Examples
Automatically generate client library examples in multiple languages:
```javascript
// JavaScript/Node.js Example
const response = await fetch('/api/users', {
 method: 'GET',
 headers: {
 'Authorization': `Bearer ${token}`,
 'Content-Type': 'application/json'
 }
});
const users = await response.json();
```

```python
# Python Example
import requests

response = requests.get('/api/users', headers={
 'Authorization': f'Bearer {token}',
 'Content-Type': 'application/json'
})
users = response.json()
```
```

#### Example 2: Code Comment Analyzer

```yaml
---
name: moai-code-comment-analyzer
description: Extract and analyze code comments, documentation, and annotations from source code across multiple programming languages. Use when you need to audit code documentation quality or generate documentation from code.
allowed-tools: Read, Grep, Glob, Write, Edit
version: 1.0.0
tags: [documentation, code-analysis, quality]
updated: 2025-11-25
status: active
---

# Code Comment Analyzer

Extract and analyze code comments, docstrings, and documentation patterns to assess documentation quality and generate structured documentation.

## Quick Reference (30 seconds)

Parse source code files to extract comments, docstrings, and annotations, then analyze documentation coverage and quality across multiple programming languages.

## Implementation Guide

### Core Capabilities
- Multi-language Parsing: Support for Python, JavaScript, Java, Go, Rust
- Documentation Coverage: Calculate percentage of documented functions/classes
- Quality Assessment: Analyze comment quality and completeness
- Missing Documentation: Identify undocumented code elements

### When to Use
- Code Reviews: Assess documentation quality before merging
- Documentation Audits: Comprehensive analysis of project documentation
- Onboarding: Generate documentation summaries for new team members
- Compliance: Ensure documentation meets organizational standards

### Essential Patterns
```python
# Extract docstrings from Python code
def extract_python_docstrings(file_path):
 """Extract all docstrings from Python source file."""
 with open(file_path, 'r') as f:
 tree = ast.parse(f.read())

 docstrings = []
 for node in ast.walk(tree):
 if isinstance(node, (ast.FunctionDef, ast.ClassDef, ast.Module)):
 if ast.get_docstring(node):
 docstrings.append({
 'type': type(node).__name__,
 'name': node.name,
 'docstring': ast.get_docstring(node),
 'line': node.lineno
 })

 return docstrings

# Calculate documentation coverage
def calculate_coverage(docstrings, total_elements):
 """Calculate percentage of documented code elements."""
 documented = len(docstrings)
 coverage = (documented / total_elements) * 100
 return {
 'documented': documented,
 'total': total_elements,
 'coverage_percentage': round(coverage, 2)
 }
```

## Best Practices

 DO:
- Analyze documentation completeness for all public APIs
- Check for outdated or incorrect documentation
- Consider comment quality, not just quantity
- Generate reports with actionable recommendations

 DON'T:
- Count comments without assessing their quality
 Ignore different documentation styles across languages
- Focus only on function-level documentation
- Assume all comments are accurate or current

## Works Well With

- [`moai-lang-python`](../moai-lang-python/SKILL.md) - Python-specific patterns
- [`moai-code-quality`](../moai-code-quality/SKILL.md) - General code quality assessment
- [`moai-cc-claude-md`](../moai-cc-claude-md/SKILL.md) - Documentation generation

## Advanced Features

### Documentation Quality Scoring
Implement sophisticated quality assessment:
```python
def assess_docstring_quality(docstring, context):
 """Assess docstring quality on multiple dimensions."""
 score = 0
 factors = []

 # Check for description
 if docstring.strip():
 score += 20
 factors.append("Has description")

 # Check for parameters documentation
 if "Args:" in docstring or "Parameters:" in docstring:
 score += 25
 factors.append("Documents parameters")

 # Check for return value documentation
 if "Returns:" in docstring or "Return:" in docstring:
 score += 20
 factors.append("Documents return value")

 # Check for examples
 if "Example:" in docstring or "Usage:" in docstring:
 score += 20
 factors.append("Includes examples")

 # Check for error documentation
 if "Raises:" in docstring or "Exceptions:" in docstring:
 score += 15
 factors.append("Documents exceptions")

 return score, factors
```

### Multi-language Standardization
Normalize documentation patterns across languages:
```javascript
// JavaScript JSDoc standardization
function standardizeJSDoc(comment) {
 // Ensure consistent JSDoc format
 return comment
 .replace(/\/\*\*?\s*\n/g, '/\n * ')
 .replace(/\*\s*@\w+/g, ' * @')
 .replace(/\s*\*\//g, ' */');
}
```
```

### 2. Language-Specific Skills

#### Example 3: Python Testing Expert

```yaml
---
name: moai-python-testing-expert
description: Comprehensive Python testing expertise covering pytest, unittest, mocking, and test-driven development patterns. Use when writing tests, setting up test infrastructure, or improving test coverage and quality.
allowed-tools: Read, Write, Edit, Bash, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
version: 1.1.0
tags: [python, testing, pytest, ddd, quality]
updated: 2025-11-25
status: active
---

# Python Testing Expert

Complete Python testing solution with pytest expertise, mocking strategies, test automation, and DDD testing patterns for production-ready code quality.

## Quick Reference (30 seconds)

Design and implement comprehensive Python test suites using pytest, unittest, mocking frameworks, and comprehensive testing methodologies for reliable, maintainable code.

## Implementation Guide

### Core Capabilities
- Pytest Mastery: Advanced pytest features, fixtures, and plugins
- Mocking Strategies: unittest.mock and pytest-mock best practices
- Test Organization: Structure tests for maintainability and scalability
- Coverage Analysis: Achieve and maintain high test coverage

### When to Use
- New Projects: Set up comprehensive testing infrastructure from scratch
- Test Improvement: Enhance existing test suites with better patterns
- Code Reviews: Validate test quality and coverage
- CI/CD Integration: Implement automated testing pipelines

### Essential Patterns
```python
# Advanced pytest fixture with factory pattern
@pytest.fixture
def user_factory():
 """Factory fixture for creating test users with different attributes."""
 created_users = []

 def create_user(kwargs):
 defaults = {
 'username': 'testuser',
 'email': 'test@example.com',
 'is_active': True
 }
 defaults.update(kwargs)

 user = User(defaults)
 created_users.append(user)
 return user

 yield create_user

 # Cleanup
 User.objects.filter(id__in=[u.id for u in created_users]).delete()

# Parametrized test with multiple scenarios
@pytest.mark.parametrize("input_data,expected_status", [
 ({"username": "valid"}, 201),
 ({"username": ""}, 400),
 ({"email": "invalid"}, 400),
])
def test_user_creation_validation(client, user_factory, input_data, expected_status):
 """Test user creation with various input validation scenarios."""
 response = client.post('/api/users', json=input_data)
 assert response.status_code == expected_status

# Mock external service with realistic behavior
@patch('requests.get')
def test_external_api_integration(mock_get, sample_responses):
 """Test integration with external API service."""
 mock_get.return_value.json.return_value = sample_responses['success']
 mock_get.return_value.status_code = 200

 result = external_service.get_data()

 assert result['status'] == 'success'
 mock_get.assert_called_once_with(
 'https://api.example.com/data',
 headers={'Authorization': 'Bearer token123'}
 )
```

```bash
# pytest configuration and execution
# pytest.ini
[tool:pytest]
testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*
addopts =
 --strict-markers
 --strict-config
 --cov=src
 --cov-report=html
 --cov-report=term-missing
 --cov-fail-under=85
markers =
 slow: marks tests as slow (deselect with '-m "not slow"')
 integration: marks tests as integration tests
 unit: marks tests as unit tests

# Run specific test categories
pytest -m unit # Unit tests only
pytest -m integration # Integration tests only
pytest -m "not slow" # Skip slow tests
pytest --cov=src --cov-report=html # With coverage report
```

## Best Practices

 DO:
- Use descriptive test names that explain the scenario
- Write independent tests that don't rely on execution order
- Create realistic test data with factories or fixtures
- Mock external dependencies but test integration points
- Aim for 85%+ coverage with meaningful tests

 DON'T:
- Write tests that depend on external services or real databases
- Use hardcoded test data that makes tests brittle
- Skip error handling and edge case testing
- Write tests that are too complex or test multiple things
- Ignore test performance and execution time

## Works Well With

- [`moai-lang-python`](../moai-lang-python/SKILL.md) - Python language patterns
- [`moai-workflow-ddd`](../moai-workflow-ddd/SKILL.md) - DDD methodology
- [`moai-quality-gate`](../moai-quality-gate/SKILL.md) - Quality validation

## Advanced Features

### Property-Based Testing
Use Hypothesis for sophisticated testing:
```python
from hypothesis import given, strategies as st

@given(st.text(min_size=1), st.text(min_size=1))
def test_string_concatenation_properties(str1, str2):
 """Test properties of string concatenation."""
 result = str1 + str2

 # Property: Length is sum of lengths
 assert len(result) == len(str1) + len(str2)

 # Property: Contains both strings
 assert str1 in result
 assert str2 in result

 # Property: Order is preserved
 assert result.index(str1) < result.index(str2) if str1 and str2 else True
```

### Performance Testing
Integrate performance testing with pytest:
```python
import time
import pytest

@pytest.mark.performance
def test_api_response_time(client):
 """Test API response time meets requirements."""
 start_time = time.time()
 response = client.get('/api/users')
 end_time = time.time()

 response_time = end_time - start_time

 assert response.status_code == 200
 assert response_time < 0.5 # Should respond in under 500ms
```

### Database Transaction Testing
Test database transaction behavior:
```python
@pytest.mark.django_db
class TestUserCreationTransaction:
 """Test user creation with database transactions."""

 def test_successful_creation(self):
 """Test successful user creation commits transaction."""
 user_count_before = User.objects.count()

 user = User.objects.create_user(
 username='testuser',
 email='test@example.com'
 )

 user_count_after = User.objects.count()
 assert user_count_after == user_count_before + 1
 assert User.objects.filter(username='testuser').exists()

 def test_rollback_on_error(self):
 """Test transaction rollback on validation error."""
 user_count_before = User.objects.count()

 with pytest.raises(ValidationError):
 User.objects.create_user(
 username='', # Invalid: empty username
 email='test@example.com'
 )

 user_count_after = User.objects.count()
 assert user_count_after == user_count_before
```
```

#### Example 4: JavaScript/TypeScript Modern Patterns

```yaml
---
name: moai-modern-javascript-patterns
description: Modern JavaScript and TypeScript patterns including ES2023+, async programming, functional programming, and type-safe development. Use when implementing modern web applications or libraries.
allowed-tools: Read, Write, Edit, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
version: 1.3.0
tags: [javascript, typescript, es2023, patterns, web]
updated: 2025-11-25
status: active
---

# Modern JavaScript & TypeScript Patterns

Contemporary JavaScript and TypeScript development patterns with ES2023+ features, type-safe programming, and modern web development best practices.

## Quick Reference (30 seconds)

Implement modern JavaScript applications using TypeScript, ES2023+ features, async/await patterns, functional programming, and type-safe development for scalable web applications.

## Implementation Guide

### Core Capabilities
- TypeScript Mastery: Advanced types, generics, and utility types
- Modern JavaScript: ES2023+ features and best practices
- Async Patterns: Promises, async/await, and concurrent programming
- Functional Programming: Immutability, pure functions, and composition

### When to Use
- Web Applications: Modern frontend and full-stack development
- Node.js Services: Backend services with type safety
- Library Development: Reusable components and utilities
- API Integration: Type-safe client-server communication

### Essential Patterns
```typescript
// Advanced TypeScript utility types
type DeepPartial<T> = {
 [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

type OptionalKeys<T> = {
 [K in keyof T]-?: {} extends Pick<T, K> ? K : never
}[keyof T];

type RequiredKeys<T> = Exclude<keyof T, OptionalKeys<T>>;

// Type-safe API client with generics
interface ApiResponse<T> {
 data: T;
 status: number;
 message?: string;
}

class ApiClient {
 async get<T>(url: string): Promise<ApiResponse<T>> {
 const response = await fetch(url);
 const data = await response.json();

 return {
 data,
 status: response.status,
 message: response.statusText
 };
 }

 async post<T>(url: string, payload: unknown): Promise<ApiResponse<T>> {
 const response = await fetch(url, {
 method: 'POST',
 headers: {
 'Content-Type': 'application/json',
 },
 body: JSON.stringify(payload),
 });

 const data = await response.json();

 return {
 data,
 status: response.status,
 message: response.statusText
 };
 }
}

// Modern async patterns with error handling
class AsyncResourceLoader<T> {
 private cache = new Map<string, Promise<T>>();
 private loading = new Map<string, boolean>();

 async load(id: string, loader: () => Promise<T>): Promise<T> {
 // Return cached promise if loading
 if (this.cache.has(id)) {
 return this.cache.get(id)!;
 }

 // Prevent duplicate loads
 if (this.loading.get(id)) {
 throw new Error(`Resource ${id} is already being loaded`);
 }

 this.loading.set(id, true);

 const promise = loader()
 .then(result => {
 this.loading.delete(id);
 return result;
 })
 .catch(error => {
 this.loading.delete(id);
 this.cache.delete(id);
 throw error;
 });

 this.cache.set(id, promise);
 return promise;
 }

 isLoaded(id: string): boolean {
 return this.cache.has(id) && !this.loading.get(id);
 }
}

// Functional programming patterns
type Predicate<T> = (value: T) => boolean;
type Mapper<T, U> = (value: T) => U;
type Reducer<T, U> = (accumulator: U, value: T) => U;

class FunctionalArray<T> extends Array<T> {
 static from<T>(array: T[]): FunctionalArray<T> {
 return Object.setPrototypeOf(array, FunctionalArray.prototype);
 }

 filter(predicate: Predicate<T>): FunctionalArray<T> {
 return FunctionalArray.from(Array.prototype.filter.call(this, predicate));
 }

 map<U>(mapper: Mapper<T, U>): FunctionalArray<U> {
 return FunctionalArray.from(Array.prototype.map.call(this, mapper));
 }

 reduce<U>(reducer: Reducer<T, U>, initialValue: U): U {
 return Array.prototype.reduce.call(this, reducer, initialValue);
 }

 // Custom functional methods
 partition(predicate: Predicate<T>): [FunctionalArray<T>, FunctionalArray<T>] {
 const truthy: T[] = [];
 const falsy: T[] = [];

 for (const item of this) {
 if (predicate(item)) {
 truthy.push(item);
 } else {
 falsy.push(item);
 }
 }

 return [FunctionalArray.from(truthy), FunctionalArray.from(falsy)];
 }

 async mapAsync<U>(mapper: Mapper<T, Promise<U>>): Promise<FunctionalArray<U>> {
 const promises = this.map(mapper);
 const results = await Promise.all(promises);
 return FunctionalArray.from(results);
 }
}

// Usage examples
const numbers = FunctionalArray.from([1, 2, 3, 4, 5, 6]);
const [even, odd] = numbers.partition(n => n % 2 === 0);

const doubled = even.map(n => n * 2); // [4, 8, 12]
const sum = doubled.reduce((acc, n) => acc + n, 0); // 24
```

## Best Practices

 DO:
- Use strict TypeScript configuration for better type safety
- Leverage utility types for type transformations
- Implement proper error handling with typed exceptions
- Use async/await consistently for asynchronous operations
- Write pure functions when possible for better testability

 DON'T:
- Use `any` type without justification
- Mix callbacks and promises in the same codebase
- Ignore TypeScript compilation errors or warnings
- Create deeply nested callback structures
- Skip proper error boundaries in React applications

## Works Well With

- [`moai-domain-frontend`](../moai-domain-frontend/SKILL.md) - Frontend development patterns
- [`moai-context7-integration`](../moai-context7-integration/SKILL.md) - Latest framework docs
- [`moai-web-performance`](../moai-web-performance/SKILL.md) - Performance optimization

## Advanced Features

### Advanced Type Manipulation
```typescript
// Type-safe event emitter
interface EventEmitterEvents {
 'user:login': (user: User) => void;
 'user:logout': () => void;
 'error': (error: Error) => void;
}

type EventHandler<T> = (payload: T) => void;

class TypedEventEmitter<T extends Record<string, any>> {
 private listeners = {} as Record<keyof T, Set<EventHandler<any>>>;

 on<K extends keyof T>(event: K, handler: EventHandler<T[K]>): void {
 if (!this.listeners[event]) {
 this.listeners[event] = new Set();
 }
 this.listeners[event].add(handler);
 }

 off<K extends keyof T>(event: K, handler: EventHandler<T[K]>): void {
 this.listeners[event]?.delete(handler);
 }

 emit<K extends keyof T>(event: K, payload: T[K]): void {
 this.listeners[event]?.forEach(handler => {
 try {
 handler(payload);
 } catch (error) {
 console.error(`Error in event handler for ${String(event)}:`, error);
 }
 });
 }
}

// Usage
const emitter = new TypedEventEmitter<EventEmitterEvents>();
emitter.on('user:login', (user) => {
 console.log(`User logged in: ${user.name}`);
});
emitter.emit('user:login', { id: 1, name: 'John' });
```

### Concurrent Programming Patterns
```typescript
// Concurrent execution with error handling
class ConcurrentExecutor {
 async executeAll<T, U>(
 tasks: Array<() => Promise<T>>,
 concurrency: number = 3
 ): Promise<Array<T | U>> {
 const results: Array<T | U> = [];
 const executing: Promise<void>[] = [];

 for (const task of tasks) {
 const promise = task()
 .then(result => {
 results.push(result);
 })
 .catch(error => {
 results.push(error as U);
 })
 .finally(() => {
 executing.splice(executing.indexOf(promise), 1);
 });

 executing.push(promise);

 if (executing.length >= concurrency) {
 await Promise.race(executing);
 }
 }

 await Promise.all(executing);
 return results;
 }
}

// Rate-limited API calls
class RateLimitedApi {
 private queue: Array<() => Promise<any>> = [];
 private processing = false;
 private lastExecution = 0;
 private readonly minInterval: number;

 constructor(requestsPerSecond: number) {
 this.minInterval = 1000 / requestsPerSecond;
 }

 async execute<T>(task: () => Promise<T>): Promise<T> {
 return new Promise((resolve, reject) => {
 this.queue.push(async () => {
 try {
 const result = await task();
 resolve(result);
 } catch (error) {
 reject(error);
 }
 });

 this.processQueue();
 });
 }

 private async processQueue(): Promise<void> {
 if (this.processing || this.queue.length === 0) {
 return;
 }

 this.processing = true;

 while (this.queue.length > 0) {
 const now = Date.now();
 const elapsed = now - this.lastExecution;

 if (elapsed < this.minInterval) {
 await new Promise(resolve =>
 setTimeout(resolve, this.minInterval - elapsed)
 );
 }

 const task = this.queue.shift()!;
 await task();
 this.lastExecution = Date.now();
 }

 this.processing = false;
 }
}
```
```

### 3. Domain-Specific Skills

#### Example 5: Security Analysis Expert

```yaml
---
name: moai-security-analysis-expert
description: Comprehensive security analysis expertise covering OWASP Top 10, vulnerability assessment, secure coding practices, and compliance validation. Use when conducting security audits, implementing security controls, or validating security measures.
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
version: 1.2.0
tags: [security, owasp, vulnerability, compliance, audit]
updated: 2025-11-25
status: active
---

# Security Analysis Expert

Complete security analysis solution with OWASP Top 10 expertise, vulnerability assessment, secure coding practices, and compliance validation for production-ready security.

## Quick Reference (30 seconds)

Conduct comprehensive security analysis using OWASP Top 10 standards, vulnerability scanning, secure code review, and compliance validation for robust application security.

## Implementation Guide

### Core Capabilities
- OWASP Top 10 Analysis: Comprehensive coverage of current security threats
- Vulnerability Assessment: Automated and manual security testing
- Secure Code Review: Static and dynamic analysis for security issues
- Compliance Validation: SOC 2, ISO 27001, and regulatory compliance

### When to Use
- Security Audits: Comprehensive security assessment of applications
- Code Reviews: Security-focused analysis of new code
- Compliance Checks: Validate against security standards and regulations
- Incident Response: Security breach analysis and remediation

### Essential Patterns
```python
# OWASP Top 10 vulnerability detection
class OWASPVulnerabilityScanner:
 def __init__(self):
 self.vulnerability_patterns = {
 'A01_2021_Broken_Access_Control': [
 r'authorization.*==.*None',
 r'@login_required.*decorator.*missing',
 r'if.*user\.is_admin.*else.*pass'
 ],
 'A02_2021_Cryptographic_Failures': [
 r'hash.*without.*salt',
 r'MD5\(',
 r'SHA1\(',
 r'password.*==.*text'
 ],
 'A03_2021_Injection': [
 r'execute\(',
 r'eval\(',
 r'\.format\(',
 r'\%.*s.*format',
 r'SQL.*string.*concatenation'
 ]
 }

 def scan_file(self, file_path: str) -> List[Dict]:
 """Scan source file for security vulnerabilities."""
 vulnerabilities = []

 with open(file_path, 'r') as f:
 content = f.read()
 lines = content.split('\n')

 for category, patterns in self.vulnerability_patterns.items():
 for pattern in patterns:
 for line_num, line in enumerate(lines, 1):
 if re.search(pattern, line, re.IGNORECASE):
 vulnerabilities.append({
 'category': category,
 'severity': self._assess_severity(category),
 'line': line_num,
 'code': line.strip(),
 'pattern': pattern,
 'recommendation': self._get_recommendation(category)
 })

 return vulnerabilities

 def _assess_severity(self, category: str) -> str:
 """Assess vulnerability severity based on category."""
 high_severity = [
 'A01_2021_Broken_Access_Control',
 'A02_2021_Cryptographic_Failures',
 'A03_2021_Injection'
 ]
 return 'HIGH' if category in high_severity else 'MEDIUM'

 def _get_recommendation(self, category: str) -> str:
 """Get security recommendation for vulnerability category."""
 recommendations = {
 'A01_2021_Broken_Access_Control':
 'Implement proper authorization checks and validate user permissions on all sensitive operations.',
 'A02_2021_Cryptographic_Failures':
 'Use strong cryptographic algorithms with proper key management and salt for password hashing.',
 'A03_2021_Injection':
 'Use parameterized queries, prepared statements, or ORMs to prevent injection attacks.'
 }
 return recommendations.get(category, 'Review OWASP guidelines for this vulnerability category.')

# Security compliance validator
class SecurityComplianceValidator:
 def __init__(self, framework: str = 'OWASP'):
 self.framework = framework
 self.compliance_rules = self._load_compliance_rules()

 def validate_application(self, app_path: str) -> Dict:
 """Validate application security compliance."""
 results = {
 'compliant': True,
 'violations': [],
 'score': 0,
 'total_checks': 0
 }

 for rule in self.compliance_rules:
 results['total_checks'] += 1

 if not self._check_rule(app_path, rule):
 results['compliant'] = False
 results['violations'].append({
 'rule': rule['name'],
 'description': rule['description'],
 'severity': rule['severity']
 })
 else:
 results['score'] += rule['weight']

 results['compliance_percentage'] = (results['score'] / results['total_checks']) * 100
 return results

 def _check_rule(self, app_path: str, rule: Dict) -> bool:
 """Check individual compliance rule."""
 if rule['type'] == 'file_exists':
 return os.path.exists(os.path.join(app_path, rule['path']))
 elif rule['type'] == 'code_scan':
 scanner = OWASPVulnerabilityScanner()
 vulnerabilities = scanner.scan_file(os.path.join(app_path, rule['file']))
 return len(vulnerabilities) == 0
 elif rule['type'] == 'configuration':
 return self._check_configuration(app_path, rule)

 return False

# Secure coding patterns generator
class SecureCodingGenerator:
 def generate_secure_code(self, template: str, context: Dict) -> str:
 """Generate secure code from templates with security best practices."""
 secure_patterns = {
 'database_access': self._generate_secure_db_access,
 'authentication': self._generate_secure_auth,
 'input_validation': self._generate_input_validation,
 'error_handling': self._generate_secure_error_handling
 }

 if template in secure_patterns:
 return secure_patterns[template](context)

 return self._apply_security_measures(template, context)

 def _generate_secure_db_access(self, context: Dict) -> str:
 """Generate secure database access code."""
 return f"""
# Secure database access with parameterized queries
def get_user_by_id(user_id: int) -> Optional[User]:
 \"\"\"Get user by ID with secure database access.\"\"\"
 try:
 with get_db_connection() as conn:
 cursor = conn.cursor()

 # Use parameterized query to prevent SQL injection
 cursor.execute(
 "SELECT id, username, email, created_at FROM users WHERE id = %s",
 (user_id,)
 )

 result = cursor.fetchone()
 if result:
 return User(
 id=result[0],
 username=result[1],
 email=result[2],
 created_at=result[3]
 )
 return None

 except DatabaseError as e:
 logger.error(f"Database error when fetching user {{user_id}}: {{e}}")
 return None
 """
```

## Best Practices

 DO:
- Follow defense-in-depth principle with multiple security layers
- Implement proper logging and monitoring for security events
- Use secure coding frameworks and libraries
- Regularly update dependencies and security patches
- Conduct periodic security assessments and penetration testing

 DON'T:
- Roll your own cryptography or security implementations
- Store sensitive data in plaintext or weak encryption
- Trust client-side input without proper validation
- Ignore security warnings from automated tools
 Assume security through obscurity is sufficient

## Works Well With

- [`moai-cc-security`](../moai-cc-security/SKILL.md) - General security patterns
- [`moai-quality-gate`](../moai-quality-gate/SKILL.md) - Quality validation
- [`moai-domain-backend`](../moai-domain-backend/SKILL.md) - Backend security

## Advanced Features

### Threat Modeling Integration
```python
# Automated threat modeling
class ThreatModelAnalyzer:
 def __init__(self):
 self.threat_categories = {
 'spoofing': self._analyze_spoofing_threats,
 'tampering': self._analyze_tampering_threats,
 'repudiation': self._analyze_repudiation_threats,
 'information_disclosure': self._analyze_information_disclosure,
 'denial_of_service': self._analyze_dos_threats,
 'elevation_of_privilege': self._analyze_elevation_threats
 }

 def analyze_application(self, app_spec: Dict) -> Dict:
 """Analyze application using STRIDE threat model."""
 threats = {}

 for category, analyzer in self.threat_categories.items():
 threats[category] = analyzer(app_spec)

 return {
 'threats': threats,
 'risk_score': self._calculate_risk_score(threats),
 'mitigations': self._generate_mitigations(threats)
 }

 def _analyze_spoofing_threats(self, app_spec: Dict) -> List[Dict]:
 """Analyze spoofing threats."""
 threats = []

 # Check authentication mechanisms
 if 'authentication' in app_spec:
 auth_spec = app_spec['authentication']
 if auth_spec.get('method') == 'password_only':
 threats.append({
 'threat': 'Credential spoofing',
 'likelihood': 'MEDIUM',
 'impact': 'HIGH',
 'description': 'Password-only authentication is vulnerable to credential spoofing'
 })

 return threats
```

### Continuous Security Monitoring
```python
# Real-time security monitoring
class SecurityMonitor:
 def __init__(self):
 self.alert_thresholds = {
 'failed_login_attempts': 5,
 'unusual_access_patterns': 10,
 'data_access_anomalies': 3
 }

 async def monitor_security_events(self):
 """Monitor security events and detect anomalies."""
 while True:
 events = await self.collect_security_events()
 anomalies = self.detect_anomalies(events)

 if anomalies:
 await self.handle_security_anomalies(anomalies)

 await asyncio.sleep(60) # Check every minute

 def detect_anomalies(self, events: List[Dict]) -> List[Dict]:
 """Detect security anomalies using pattern analysis."""
 anomalies = []

 # Check for brute force attacks
 failed_logins = [e for e in events if e['type'] == 'failed_login']
 if len(failed_logins) > self.alert_thresholds['failed_login_attempts']:
 anomalies.append({
 'type': 'brute_force_attack',
 'severity': 'HIGH',
 'events': failed_logins,
 'recommendation': 'Implement rate limiting and account lockout'
 })

 return anomalies
```
```

### 4. Integration and Workflow Skills

#### Example 6: Workflow Automation Expert

```yaml
---
name: moai-workflow-automation-expert
description: Workflow automation expertise covering CI/CD pipelines, DevOps automation, infrastructure as code, and deployment strategies. Use when setting up automated workflows, CI/CD pipelines, or infrastructure management.
allowed-tools: Read, Write, Edit, Bash, Grep, Glob, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
version: 1.1.0
tags: [automation, cicd, devops, infrastructure, workflow]
updated: 2025-11-25
status: active
---

# Workflow Automation Expert

Complete workflow automation solution with CI/CD pipeline expertise, DevOps automation, infrastructure as code, and deployment strategies for modern software development.

## Quick Reference (30 seconds)

Design and implement automated workflows using CI/CD pipelines, infrastructure as code, deployment automation, and monitoring for efficient software development and deployment.

## Implementation Guide

### Core Capabilities
- CI/CD Pipeline Design: GitHub Actions, GitLab CI, Jenkins automation
- Infrastructure as Code: Terraform, CloudFormation, Ansible expertise
- Deployment Automation: Blue-green, canary, and rolling deployments
- Monitoring Integration: Automated testing, quality gates, and alerting

### When to Use
- New Projects: Set up comprehensive CI/CD from scratch
- Pipeline Optimization: Improve existing automation workflows
- Infrastructure Management: Automate infrastructure provisioning and management
- Deployment Strategies: Implement advanced deployment patterns

### Essential Patterns
```yaml
# GitHub Actions workflow template
name: CI/CD Pipeline
on:
 push:
 branches: [main, develop]
 pull_request:
 branches: [main]

env:
 NODE_VERSION: '18'
 PYTHON_VERSION: '3.11'

jobs:
 test:
 name: Test Suite
 runs-on: ubuntu-latest
 strategy:
 matrix:
 node-version: [16, 18, 20]

 steps:
 - name: Checkout code
 uses: actions/checkout@v4

 - name: Setup Node.js
 uses: actions/setup-node@v4
 with:
 node-version: ${{ matrix.node-version }}
 cache: 'npm'

 - name: Install dependencies
 run: npm ci

 - name: Run linting
 run: npm run lint

 - name: Run tests
 run: npm run test:coverage

 - name: Upload coverage reports
 uses: codecov/codecov-action@v3
 with:
 file: ./coverage/lcov.info

 security:
 name: Security Scan
 runs-on: ubuntu-latest
 steps:
 - name: Checkout code
 uses: actions/checkout@v4

 - name: Run security audit
 run: npm audit --audit-level high

 - name: Run dependency check
 uses: snyk/actions/node@master
 env:
 SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}

 build:
 name: Build Application
 runs-on: ubuntu-latest
 needs: [test, security]
 steps:
 - name: Checkout code
 uses: actions/checkout@v4

 - name: Setup Node.js
 uses: actions/setup-node@v4
 with:
 node-version: ${{ env.NODE_VERSION }}
 cache: 'npm'

 - name: Install dependencies
 run: npm ci

 - name: Build application
 run: npm run build

 - name: Build Docker image
 run: |
 docker build -t myapp:${{ github.sha }} .
 docker tag myapp:${{ github.sha }} myapp:latest

 - name: Push to registry
 if: github.ref == 'refs/heads/main'
 run: |
 echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
 docker push myapp:${{ github.sha }}
 docker push myapp:latest

 deploy:
 name: Deploy to Production
 runs-on: ubuntu-latest
 needs: [build]
 if: github.ref == 'refs/heads/main'
 environment: production

 steps:
 - name: Deploy to Kubernetes
 run: |
 kubectl set image deployment/myapp myapp=myapp:${{ github.sha }}
 kubectl rollout status deployment/myapp

 - name: Run smoke tests
 run: |
 npm run test:smoke
```

```python
# Terraform infrastructure as code
# main.tf
terraform {
 required_version = ">= 1.0"
 required_providers {
 aws = {
 source = "hashicorp/aws"
 version = "~> 5.0"
 }
 }

 backend "s3" {
 bucket = "my-terraform-state"
 key = "production/terraform.tfstate"
 region = "us-west-2"
 }
}

provider "aws" {
 region = var.aws_region
}

# VPC configuration
resource "aws_vpc" "main" {
 cidr_block = var.vpc_cidr
 enable_dns_hostnames = true
 enable_dns_support = true

 tags = {
 Name = "main-vpc"
 Environment = var.environment
 }
}

# Security groups
resource "aws_security_group" "web" {
 name_prefix = "web-sg"
 vpc_id = aws_vpc.main.id

 ingress {
 from_port = 80
 to_port = 80
 protocol = "tcp"
 cidr_blocks = ["0.0.0.0/0"]
 }

 ingress {
 from_port = 443
 to_port = 443
 protocol = "tcp"
 cidr_blocks = ["0.0.0.0/0"]
 }

 egress {
 from_port = 0
 to_port = 0
 protocol = "-1"
 cidr_blocks = ["0.0.0.0/0"]
 }

 tags = {
 Name = "web-sg"
 Environment = var.environment
 }
}

# ECS cluster and service
resource "aws_ecs_cluster" "main" {
 name = "${var.project_name}-cluster"

 setting {
 name = "containerInsights"
 value = "enabled"
 }
}

resource "aws_ecs_task_definition" "app" {
 family = "${var.project_name}-app"
 network_mode = "awsvpc"
 requires_compatibilities = ["FARGATE"]
 cpu = "256"
 memory = "512"

 container_definitions = jsonencode([
 {
 name = "app"
 image = "${var.aws_account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/${var.project_name}:${var.image_tag}"

 port_mappings = [
 {
 containerPort = 80
 protocol = "tcp"
 }
 ]

 environment = [
 {
 name = "NODE_ENV"
 value = var.environment
 }
 ]

 log_configuration = {
 log_driver = "awslogs"
 options = {
 "awslogs-group" = "/ecs/${var.project_name}"
 "awslogs-region" = var.aws_region
 "awslogs-stream-prefix" = "ecs"
 }
 }
 }
 ])
}

# Auto Scaling
resource "aws_appautoscaling_target" "ecs_target" {
 max_capacity = 10
 min_capacity = 2
 resource_id = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.app.name}"
 scalable_dimension = "ecs:service:DesiredCount"
 service_namespace = "ecs"
}

resource "aws_appautoscaling_policy" "ecs_policy_cpu" {
 name = "cpu-autoscaling"
 policy_type = "TargetTrackingScaling"
 resource_id = aws_appautoscaling_target.ecs_target.resource_id
 scalable_dimension = aws_appautoscaling_target.ecs_target.scalable_dimension
 service_namespace = aws_appautoscaling_target.ecs_target.service_namespace

 target_tracking_scaling_policy_configuration {
 predefined_metric_specification {
 predefined_metric_type = "ECSServiceAverageCPUUtilization"
 }

 target_value = 70.0
 }
}
```

## Best Practices

 DO:
- Implement proper secrets management using environment variables or secret stores
- Use infrastructure as code for reproducible deployments
- Implement proper monitoring and alerting for all services
- Use blue-green or canary deployments for zero-downtime releases
- Implement proper rollback mechanisms for failed deployments

 DON'T:
- Hardcode credentials or sensitive information in configuration
 Skip proper testing and validation in CI/CD pipelines
- Deploy directly to production without staging validation
- Ignore security scanning and vulnerability assessment
- Use manual processes for repetitive deployment tasks

## Works Well With

- [`moai-devops-expert`](../moai-devops-expert/SKILL.md) - DevOps best practices
- [`moai-monitoring-expert`](../moai-monitoring-expert/SKILL.md) - Monitoring strategies
- [`moai-security-expert`](../moai-security-expert/SKILL.md) - Security automation

## Advanced Features

### Multi-Environment Deployment
```python
# Environment-specific configuration management
class DeploymentManager:
 def __init__(self, config_file: str):
 self.config = self._load_config(config_file)
 self.environments = self.config['environments']

 def deploy_to_environment(self, environment: str, version: str):
 """Deploy application to specific environment with validation."""
 if environment not in self.environments:
 raise ValueError(f"Unknown environment: {environment}")

 env_config = self.environments[environment]

 # Pre-deployment validation
 self._validate_environment(environment)

 # Deploy with environment-specific configuration
 self._apply_configuration(environment)
 self._deploy_application(version, env_config)

 # Post-deployment validation
 self._validate_deployment(environment, version)

 def _validate_environment(self, environment: str):
 """Validate environment is ready for deployment."""
 env_config = self.environments[environment]

 # Check required services are running
 for service in env_config.get('required_services', []):
 if not self._check_service_health(service):
 raise RuntimeError(f"Required service {service} is not healthy")

 # Check resource availability
 if not self._check_resource_availability(environment):
 raise RuntimeError(f"Insufficient resources in {environment}")

 def _validate_deployment(self, environment: str, version: str):
 """Validate deployment was successful."""
 env_config = self.environments[environment]

 # Run health checks
 for health_check in env_config.get('health_checks', []):
 if not self._run_health_check(health_check):
 raise RuntimeError(f"Health check failed: {health_check}")

 # Run smoke tests
 if 'smoke_tests' in env_config:
 self._run_smoke_tests(env_config['smoke_tests'])
```

### Automated Rollback
```yaml
# Deployment with automatic rollback
name: Deploy with Rollback
on:
 push:
 branches: [main]

jobs:
 deploy:
 runs-on: ubuntu-latest
 steps:
 - name: Deploy application
 id: deploy
 run: |
 # Deploy and get new version
 NEW_VERSION=$(deploy.sh)
 echo "new_version=$NEW_VERSION" >> $GITHUB_OUTPUT

 - name: Health check
 run: |
 # Wait for deployment to be ready
 sleep 30

 # Run health checks
 if ! health-check.sh; then
 echo "Health check failed, initiating rollback"
 echo "needs_rollback=true" >> $GITHUB_ENV
 fi

 - name: Rollback on failure
 if: env.needs_rollback == 'true'
 run: |
 # Get previous version
 PREVIOUS_VERSION=$(get-previous-version.sh)

 # Rollback to previous version
 rollback.sh $PREVIOUS_VERSION

 # Notify team
 notify-rollback.sh ${{ steps.deploy.outputs.new_version }} $PREVIOUS_VERSION
```
```

---

## Skill Creation Process

### 1. Planning Phase

Identify Need:
- What specific problem does this skill solve?
- Who are the target users?
- What are the trigger scenarios?

Define Scope:
- Single responsibility principle
- Clear boundaries and limitations
- Integration points with other skills

### 2. Design Phase

Architecture Design:
- Progressive disclosure structure
- Tool permission requirements
- Error handling strategies

Content Planning:
- Quick Reference (30-second value)
- Implementation Guide structure
- Best Practices and examples

### 3. Implementation Phase

Frontmatter Creation:
```yaml
---
name: skill-name
description: Specific description with trigger scenarios
allowed-tools: minimal, specific, tools
version: 1.0.0
tags: [relevant, tags]
updated: 2025-11-25
status: active
---
```

Content Development:
- Start with Quick Reference
- Build Implementation Guide with examples
- Add Best Practices with DO/DON'T
- Include Works Well With section

### 4. Validation Phase

Technical Validation:
- YAML syntax validation
- Code example testing
- Link verification
- Line count compliance

Quality Validation:
- Content clarity and specificity
- Technical accuracy
- User experience optimization
- Standards compliance

### 5. Publication Phase

File Structure:
```
skill-name/
 SKILL.md (≤500 lines)
 reference.md (if needed)
 examples.md (if needed)
 scripts/ (if needed)
```

Version Control:
- Semantic versioning
- Change documentation
- Update tracking
- Compatibility notes

---

## Maintenance and Updates

### Regular Review Schedule

Monthly Reviews:
- Check for official standards updates
- Review example code for currency
- Validate external links and references
- Update best practices based on community feedback

Quarterly Updates:
- Major version compatibility checks
- Performance optimization reviews
- Integration testing with other skills
- User feedback incorporation

### Update Process

1. Assessment: Determine update scope and impact
2. Planning: Plan changes with backward compatibility
3. Implementation: Update content and examples
4. Testing: Validate all functionality and examples
5. Documentation: Update changelog and version info
6. Publication: Deploy with proper version bumping

---

Version: 2.0.0
Compliance: Claude Code Official Standards
Last Updated: 2025-11-25
Examples Count: 6 comprehensive examples
Skill Categories: Documentation, Language, Domain, Integration

Generated with Claude Code using official documentation and best practices.
