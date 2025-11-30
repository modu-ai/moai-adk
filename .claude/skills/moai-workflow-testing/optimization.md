---
name: moai-workflow-testing/optimization
description: Performance optimization, parallel execution, and resource management for Playwright testing
---

# Playwright Testing Optimization (v5.0.0)

## Test Execution Optimization

### 1. Parallel Execution Strategy

```typescript
import { defineConfig } from '@playwright/test';

export default defineConfig({
 // Global test timeout
 timeout: 30000,

 // Parallel workers configuration
 workers: process.env.CI ? 1 : 4,

 // Retry failed tests
 retries: process.env.CI ? 2 : 0,

 // Reporter configuration
 reporter: [
 ['html', { outputFolder: 'playwright-report' }],
 ['json', { outputFile: 'test-results.json' }],
 ['junit', { outputFile: 'junit-results.xml' }]
 ],

 projects: [
 {
 name: 'chromium',
 use: { ...devices['Desktop Chrome'] },
 // Shard tests across multiple machines
 fullyParallel: true
 },
 {
 name: 'firefox',
 use: { ...devices['Desktop Firefox'] },
 fullyParallel: true
 }
 ]
});
```

### 2. Test Sharding Across Machines

```typescript
// Run on machine 1 (test shards 1-3)
// npx playwright test --shard=1/3

// Run on machine 2 (test shards 4-6)
// npx playwright test --shard=2/3

// Run on machine 3 (test shards 7-10)
// npx playwright test --shard=3/3

// Merge results locally
// npx playwright merge-reports ./blob-report

export default defineConfig({
 // Use blob reporting for merging across shards
 reporter: [['blob', { outputFile: 'blob-report.zip' }]],

 webServer: {
 command: 'npm run start',
 port: 3000,
 reuseExistingServer: !process.env.CI
 }
});
```

### 3. Resource Pool Management

```typescript
class BrowserPoolManager {
 private pool: Browser[] = [];
 private activeConnections = 0;
 private maxConnections = 5;
 private queue: (() => Promise<Browser>)[] = [];

 async getBrowser(): Promise<Browser> {
 if (this.activeConnections >= this.maxConnections) {
 // Wait for browser to become available
 return new Promise((resolve) => {
 this.queue.push(() => {
 this.activeConnections++;
 return this.getAvailableBrowser().then(browser => {
 resolve(browser);
 return browser;
 });
 });
 });
 }

 this.activeConnections++;
 return this.getAvailableBrowser();
 }

 private async getAvailableBrowser(): Promise<Browser> {
 if (this.pool.length > 0) {
 return this.pool.pop()!;
 }

 const playwright = require('playwright');
 return playwright.chromium.launch({
 args: ['--disable-extensions']
 });
 }

 async releaseBrowser(browser: Browser) {
 this.activeConnections--;

 if (this.queue.length > 0) {
 const fn = this.queue.shift()!;
 fn();
 } else {
 this.pool.push(browser);
 }
 }

 async cleanup() {
 for (const browser of this.pool) {
 await browser.close();
 }
 }
}
```

## Memory and Performance Optimization

### 1. Memory-Efficient Page Management

```typescript
class PagePoolManager {
 private pagePool: Map<string, Page> = new Map();
 private maxPages = 5;

 async getOrCreatePage(context: BrowserContext): Promise<Page> {
 // Reuse existing pages
 for (const [, page] of this.pagePool) {
 if (!page.isClosed()) {
 return page;
 }
 }

 if (this.pagePool.size >= this.maxPages) {
 // Close oldest page
 const oldest = Array.from(this.pagePool.entries())[0];
 await oldest[1].close();
 this.pagePool.delete(oldest[0]);
 }

 const page = await context.newPage();

 // Set memory limits
 await page.addInitScript(() => {
 // Clear unnecessary data
 (window as any).localStorage.clear();
 (window as any).sessionStorage.clear();
 });

 return page;
 }

 async cleanup() {
 for (const [, page] of this.pagePool) {
 await page.close();
 }
 this.pagePool.clear();
 }
}
```

### 2. Screenshot and Video Optimization

```typescript
class MediaOptimizer {
 async optimizeScreenshots(testResults: TestResult[]) {
 for (const result of testResults) {
 if (result.screenshotPath) {
 // Compress screenshots to 50% quality
 await sharp(result.screenshotPath)
 .png({ quality: 50, progressive: true })
 .toFile(result.screenshotPath);
 }
 }
 }

 async optimizeVideos(videoPath: string) {
 // Convert to webm format for smaller file size
 return new Promise((resolve, reject) => {
 ffmpeg(videoPath)
 .output(`${videoPath}.webm`)
 .videoCodec('libvpx-vp9')
 .audioBitrate('64k')
 .on('end', () => resolve(`${videoPath}.webm`))
 .on('error', reject)
 .run();
 });
 }

 configureRecording() {
 return {
 video: 'retain-on-failure',
 videoSize: { width: 1280, height: 720 },
 videoSnapshotInterval: 500, // 2fps instead of default
 };
 }
}
```

## Test Optimization Patterns

### 1. Smart Test Grouping

```typescript
// Group related tests to share setup
test.describe('User Authentication', () => {
 test.beforeAll(async () => {
 // Setup test data once for all tests in group
 await setupTestDatabase();
 await createTestUsers();
 });

 test('login with valid credentials', async ({ page }) => {
 // Uses shared test data
 });

 test('logout clears session', async ({ page }) => {
 // Uses shared test data
 });

 test.afterAll(async () => {
 // Cleanup once for all tests
 await cleanupTestDatabase();
 });
});
```

### 2. Fixture-Based Optimization

```typescript
import { test as base } from '@playwright/test';

// Create reusable fixtures
type TestFixtures = {
 authenticatedPage: Page;
 apiClient: APIClient;
 database: DatabaseConnection;
};

const test = base.extend<TestFixtures>({
 // Fixture with automatic cleanup
 authenticatedPage: async ({ page }, use) => {
 // Setup: Login user
 await page.goto('/login');
 await page.fill('[name="email"]', 'test@example.com');
 await page.fill('[name="password"]', 'password123');
 await page.click('button[type="submit"]');
 await page.waitForNavigation();

 // Use the page in tests
 await use(page);

 // Cleanup: Logout
 await page.click('[aria-label="User menu"]');
 await page.click('[aria-label="Logout"]');
 },

 apiClient: async ({}, use) => {
 const client = new APIClient('http://localhost:3000');
 await client.connect();

 await use(client);

 await client.disconnect();
 },

 database: async ({}, use) => {
 const db = new DatabaseConnection('postgresql://...');
 await db.connect();

 // Clear test data before each test
 await db.clearTestData();

 await use(db);

 await db.disconnect();
 }
});

export { test };

// Usage in tests
test('user can update profile', async ({ authenticatedPage, database }) => {
 // No setup needed - fixtures handle it
 await authenticatedPage.goto('/profile');
 await authenticatedPage.fill('[name="bio"]', 'Updated bio');
 await authenticatedPage.click('button[type="submit"]');

 // Verify in database
 const user = await database.query('SELECT * FROM users WHERE id = 1');
 expect(user.bio).toBe('Updated bio');
});
```

## CI/CD Optimization

### 1. Smart Test Selection

```typescript
// Run only tests affected by code changes
import { spawnSync } from 'child_process';

function getChangedFiles(): string[] {
 const result = spawnSync('git', ['diff', '--name-only', 'main...HEAD'], {
 encoding: 'utf-8'
 });
 return result.stdout.trim().split('\n');
}

function selectTestsForChanges(changedFiles: string[]): string[] {
 const testMap = {
 'src/auth': 'tests/auth.spec.ts',
 'src/profile': 'tests/profile.spec.ts',
 'src/payments': 'tests/payments.spec.ts'
 };

 const testsToRun = new Set<string>();

 for (const file of changedFiles) {
 for (const [srcPath, testFile] of Object.entries(testMap)) {
 if (file.includes(srcPath)) {
 testsToRun.add(testFile);
 }
 }
 }

 return Array.from(testsToRun);
}

// Usage in CI
if (process.env.CI) {
 const changedFiles = getChangedFiles();
 const testsToRun = selectTestsForChanges(changedFiles);
 console.log(`Running ${testsToRun.length} affected tests`);
}
```

### 2. Test Result Caching

```typescript
class TestResultCache {
 async getCachedResults(testHash: string): Promise<TestResult | null> {
 const cacheKey = `test-results-${testHash}`;
 const cached = await redis.get(cacheKey);

 if (cached) {
 return JSON.parse(cached);
 }

 return null;
 }

 async cacheResults(testHash: string, results: TestResult) {
 const cacheKey = `test-results-${testHash}`;

 // Cache for 24 hours
 await redis.setex(cacheKey, 86400, JSON.stringify(results));
 }

 generateTestHash(testCode: string, dependencies: string[]): string {
 const hash = crypto
 .createHash('sha256')
 .update(testCode)
 .update(dependencies.join(''))
 .digest('hex');

 return hash;
 }
}
```

### 3. Performance Budgets

```typescript
interface PerformanceBudget {
 maxTestDuration: number; // milliseconds
 maxMemoryUsage: number; // MB
 maxNetworkRequests: number;
 allowedRetries: number;
}

const budgets: Record<string, PerformanceBudget> = {
 'unit-test': {
 maxTestDuration: 500,
 maxMemoryUsage: 50,
 maxNetworkRequests: 0,
 allowedRetries: 0
 },
 'integration-test': {
 maxTestDuration: 5000,
 maxMemoryUsage: 200,
 maxNetworkRequests: 5,
 allowedRetries: 1
 },
 'e2e-test': {
 maxTestDuration: 30000,
 maxMemoryUsage: 500,
 maxNetworkRequests: 50,
 allowedRetries: 2
 }
};

async function validatePerformanceBudget(testName: string, metrics: TestMetrics) {
 const budget = budgets[testName] || budgets['e2e-test'];
 const violations: string[] = [];

 if (metrics.duration > budget.maxTestDuration) {
 violations.push(`Duration ${metrics.duration}ms exceeds ${budget.maxTestDuration}ms`);
 }

 if (metrics.memoryUsage > budget.maxMemoryUsage) {
 violations.push(`Memory ${metrics.memoryUsage}MB exceeds ${budget.maxMemoryUsage}MB`);
 }

 if (violations.length > 0) {
 throw new Error(`Performance budget exceeded:\n${violations.join('\n')}`);
 }
}
```

## Monitoring and Observability

### 1. Test Metrics Collection

```typescript
class TestMetricsCollector {
 async collectMetrics(testResult: TestResult) {
 const metrics = {
 testName: testResult.title,
 status: testResult.status,
 duration: testResult.duration,
 retries: testResult.retries,
 memoryUsage: process.memoryUsage().heapUsed / 1024 / 1024,
 browserMetrics: {
 domNodes: await testResult.page?.evaluate(() => document.querySelectorAll('*').length),
 eventListeners: await testResult.page?.evaluate(() => {
 let count = 0;
 for (const elem of document.querySelectorAll('*')) {
 count += getEventListeners(elem).length;
 }
 return count;
 })
 },
 timestamp: new Date().toISOString()
 };

 // Send to monitoring service
 await sendMetrics(metrics);
 return metrics;
 }
}
```

### 2. Test Report Generation

```typescript
import { reportPortalClient } from '@reportportal/client-javascript';

// Configure ReportPortal integration
const client = new reportPortalClient({
 endpoint: 'https://rp.example.com',
 project: 'my-project',
 token: process.env.RP_TOKEN
});

// Send test results
function reportTestResults(results: TestResult[]) {
 for (const result of results) {
 client.startTest({
 name: result.title,
 startTime: result.startTime
 });

 client.log({
 message: result.error || 'Test passed',
 level: result.status === 'passed' ? 'INFO' : 'ERROR',
 time: new Date()
 });

 client.finishTest({
 endTime: result.endTime,
 status: result.status === 'passed' ? 'PASSED' : 'FAILED'
 });
 }
}
```

---

Version: 5.0.0 | Last Updated: 2025-11-22 | Enterprise Ready:
