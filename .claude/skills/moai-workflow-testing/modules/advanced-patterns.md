---
name: moai-workflow-testing/advanced-patterns
description: Advanced Playwright patterns for enterprise web testing, visual regression, E2E orchestration
---

# Advanced Playwright Testing Patterns (v5.0.0)

## Visual Regression Testing Strategies

### 1. Screenshot-Based Regression Testing

```typescript
import { test, expect } from '@playwright/test';

test.describe('Visual Regression Tests', () => {
 test('homepage visual consistency', async ({ page }) => {
 await page.goto('https://example.com');

 // Take full page screenshot
 await expect(page).toHaveScreenshot('homepage.png', {
 fullPage: true,
 timeout: 5000,
 animations: 'disabled'
 });
 });

 test('component visual regression - header', async ({ page }) => {
 await page.goto('https://example.com');

 const header = page.locator('header');

 // Component-specific screenshot
 await expect(header).toHaveScreenshot('header-component.png', {
 mask: [page.locator('nav a')] // Mask dynamic elements
 });
 });

 test('responsive visual regression', async ({ page }) => {
 const breakpoints = [
 { name: 'mobile', width: 375, height: 812 },
 { name: 'tablet', width: 768, height: 1024 },
 { name: 'desktop', width: 1920, height: 1080 }
 ];

 for (const breakpoint of breakpoints) {
 await page.setViewportSize(breakpoint);
 await page.goto('https://example.com');

 await expect(page).toHaveScreenshot(
 `homepage-${breakpoint.name}.png`,
 { fullPage: true }
 );
 }
 });

 test('dark mode visual consistency', async ({ page }) => {
 // Set dark mode preference
 await page.emulateMedia({ colorScheme: 'dark' });
 await page.goto('https://example.com');

 await expect(page).toHaveScreenshot('homepage-dark-mode.png');
 });
});
```

### 2. Pixel-Perfect Diff Analysis

```typescript
class VisualRegressionAnalyzer {
 async analyzeScreenshotDiff(
 baseline: Buffer,
 current: Buffer
 ): Promise<DiffAnalysis> {
 const pixelsChanged = this.calculatePixelDifference(baseline, current);

 return {
 totalPixels: baseline.length,
 changedPixels: pixelsChanged,
 percentageChange: (pixelsChanged / baseline.length) * 100,
 severity: this.classifyDifference(pixelsChanged),
 regions: this.identifyChangedRegions(baseline, current),
 recommendations: this.generateRecommendations(pixelsChanged)
 };
 }

 private classifyDifference(changedPixels: number): string {
 if (changedPixels === 0) return 'IDENTICAL';
 if (changedPixels < 1000) return 'MINOR';
 if (changedPixels < 10000) return 'MODERATE';
 return 'MAJOR';
 }

 private identifyChangedRegions(
 baseline: Buffer,
 current: Buffer
 ): ChangedRegion[] {
 // Use image processing to identify changed regions
 const regions: ChangedRegion[] = [];
 const diff = this.pixelDiff(baseline, current);

 // Find contiguous regions of differences
 for (const region of this.findContiguousRegions(diff)) {
 regions.push({
 x: region.x,
 y: region.y,
 width: region.width,
 height: region.height,
 confidence: region.confidence
 });
 }

 return regions;
 }

 private generateRecommendations(changedPixels: number): string[] {
 const recommendations: string[] = [];

 if (changedPixels > 50000) {
 recommendations.push('REJECT: Significant visual changes detected');
 recommendations.push('Recommendation: Verify design changes with design team');
 } else if (changedPixels > 10000) {
 recommendations.push('REVIEW: Moderate visual changes detected');
 recommendations.push('Recommendation: Manual review recommended before merge');
 } else {
 recommendations.push('APPROVE: Minor visual changes detected');
 recommendations.push('Recommendation: OK to auto-approve if expected');
 }

 return recommendations;
 }
}
```

## Advanced Playwright Patterns

### 1. Page Object Model (POM) with Advanced Features

```typescript
// Base page object class
class BasePage {
 constructor(protected page: Page) {}

 async goto(url: string) {
 await this.page.goto(url, { waitUntil: 'networkidle' });
 }

 async waitForElement(selector: string, timeout = 5000) {
 await this.page.waitForSelector(selector, { timeout });
 }

 async fillForm(formData: Record<string, string>) {
 for (const [selector, value] of Object.entries(formData)) {
 await this.page.fill(selector, value);
 }
 }

 async clickAndWaitForNavigation(selector: string) {
 await Promise.all([
 this.page.waitForNavigation({ waitUntil: 'networkidle' }),
 this.page.click(selector)
 ]);
 }

 async handleDialog(action: 'accept' | 'dismiss', text?: string) {
 this.page.on('dialog', async (dialog) => {
 if (action === 'accept') {
 await dialog.accept(text);
 } else {
 await dialog.dismiss();
 }
 });
 }
}

// Specific page object
class LoginPage extends BasePage {
 // Selectors
 private emailInput = 'input[type="email"]';
 private passwordInput = 'input[type="password"]';
 private loginButton = 'button:has-text("Login")';
 private errorMessage = '.error-message';

 async login(email: string, password: string) {
 await this.page.fill(this.emailInput, email);
 await this.page.fill(this.passwordInput, password);
 await this.page.click(this.loginButton);
 await this.page.waitForNavigation({ waitUntil: 'networkidle' });
 }

 async getErrorMessage(): Promise<string> {
 return this.page.locator(this.errorMessage).textContent();
 }

 async isErrorMessageVisible(): Promise<boolean> {
 return this.page.locator(this.errorMessage).isVisible();
 }
}

// Usage in test
test('login with valid credentials', async ({ page }) => {
 const loginPage = new LoginPage(page);
 await loginPage.goto('https://example.com/login');
 await loginPage.login('user@example.com', 'password123');

 // Navigate to dashboard
 expect(page.url()).toContain('/dashboard');
});
```

### 2. Intelligent Wait Strategies

```typescript
class SmartWaitManager {
 constructor(private page: Page) {}

 async waitForAPI(method: 'GET' | 'POST' | 'PUT', urlPattern: RegExp) {
 const response = await this.page.waitForResponse(
 (response) => response.request().method() === method &&
 urlPattern.test(response.url())
 );
 return response.json();
 }

 async waitForMultipleRequests(patterns: { method: string; url: RegExp }[]) {
 const responses = await Promise.all(
 patterns.map(pattern =>
 this.page.waitForResponse(response =>
 response.request().method() === pattern.method &&
 pattern.url.test(response.url())
 )
 )
 );
 return responses;
 }

 async waitForNetworkIdle(timeout = 5000) {
 // Wait for network to be completely idle
 let activeRequests = 0;
 let lastRequestTime = Date.now();

 this.page.on('request', () => {
 activeRequests++;
 lastRequestTime = Date.now();
 });

 this.page.on('response', () => {
 activeRequests--;
 });

 while (Date.now() - lastRequestTime < timeout || activeRequests > 0) {
 await this.page.waitForTimeout(100);
 }
 }

 async waitForStateChange(selector: string, expectedState: string) {
 await this.page.waitForFunction(
 (selector, state) => {
 const element = document.querySelector(selector);
 return element?.getAttribute('data-state') === state;
 },
 [selector, expectedState],
 { timeout: 5000 }
 );
 }
}
```

### 3. API Request/Response Interception

```typescript
class APIInterceptor {
 constructor(private page: Page) {}

 async mockAPI(
 urlPattern: RegExp,
 responseData: any,
 statusCode = 200
 ) {
 await this.page.route(urlPattern, (route) => {
 route.abort('aborted');
 route.continue();
 });

 // Intercept and mock response
 await this.page.route(urlPattern, (route) => {
 route.fulfill({
 status: statusCode,
 contentType: 'application/json',
 body: JSON.stringify(responseData)
 });
 });
 }

 async recordAPITraffic(): Promise<APICall[]> {
 const apiCalls: APICall[] = [];

 this.page.on('response', async (response) => {
 if (response.request().resourceType() === 'xhr' ||
 response.request().resourceType() === 'fetch') {
 apiCalls.push({
 method: response.request().method(),
 url: response.url(),
 statusCode: response.status(),
 headers: response.headers(),
 body: await response.json().catch(() => null),
 timestamp: new Date()
 });
 }
 });

 return apiCalls;
 }

 async validateAPIContract(
 endpoint: string,
 expectedSchema: object
 ): Promise<boolean> {
 const apiCalls = await this.recordAPITraffic();
 const call = apiCalls.find(c => c.url.includes(endpoint));

 if (!call) return false;

 return this.validateAgainstSchema(call.body, expectedSchema);
 }
}
```

## Cross-Browser Testing Orchestration

### 1. Distributed Cross-Browser Testing

```typescript
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
 testDir: './tests',

 // Parallel execution across browsers
 workers: 4,

 use: {
 baseURL: 'http://localhost:3000',
 trace: 'on-first-retry',
 screenshot: 'only-on-failure',
 video: 'retain-on-failure'
 },

 projects: [
 {
 name: 'chromium',
 use: { ...devices['Desktop Chrome'] }
 },
 {
 name: 'firefox',
 use: { ...devices['Desktop Firefox'] }
 },
 {
 name: 'webkit',
 use: { ...devices['Desktop Safari'] }
 },
 {
 name: 'mobile-chrome',
 use: { ...devices['Pixel 5'] }
 },
 {
 name: 'mobile-safari',
 use: { ...devices['iPhone 12'] }
 },
 {
 name: 'tablet',
 use: { ...devices['iPad Pro'] }
 }
 ]
});
```

### 2. Browser-Specific Test Adjustment

```typescript
test.describe('Cross-browser compatibility', () => {
 test.beforeEach(async ({ browser, page }) => {
 const browserName = browser.browserType().name();

 // Browser-specific setup
 if (browserName === 'webkit') {
 // Safari-specific handling
 await page.addInitScript(() => {
 // Polyfills for Safari
 if (!window.Promise) {
 console.warn('Promise not supported');
 }
 });
 } else if (browserName === 'firefox') {
 // Firefox-specific setup
 }
 });

 test('form submission works in all browsers', async ({ page, browserName }) => {
 await page.goto('/form');
 await page.fill('input[name="email"]', 'test@example.com');
 await page.fill('input[name="message"]', 'Test message');

 // Browser-specific selector adjustments
 const submitSelector = browserName === 'webkit'
 ? 'button:visible:text("Submit")'
 : 'button[type="submit"]';

 await page.click(submitSelector);

 // Verify success
 if (browserName === 'firefox') {
 // Firefox-specific verification
 await page.waitForNavigation({ timeout: 3000 });
 } else {
 await page.waitForSelector('.success-message');
 }
 });
});
```

## Performance Testing with Playwright

### 1. Lighthouse Integration

```typescript
import { playAudit } from 'playwright-lighthouse';

test('performance metrics', async ({ page }) => {
 await playAudit({
 page: page,
 port: 9222,
 thresholds: {
 performance: 75,
 accessibility: 80,
 'best-practices': 85,
 seo: 90
 },
 reports: {
 formats: {
 json: './lighthouse-report.json',
 html: './lighthouse-report.html'
 },
 directory: './lighthouse'
 }
 });
});
```

### 2. Custom Performance Metrics

```typescript
class PerformanceMonitor {
 async collectMetrics(page: Page): Promise<PerformanceMetrics> {
 const metrics = await page.evaluate(() => {
 const navigation = performance.getEntriesByType('navigation')[0];
 const paintEntries = performance.getEntriesByType('paint');

 return {
 // Core Web Vitals
 FCP: paintEntries.find(e => e.name === 'first-contentful-paint')?.startTime || 0,
 LCP: performance.getEntriesByType('largest-contentful-paint').pop()?.startTime || 0,
 CLS: Math.round(performance.getEntriesByType('layout-shift')
 .reduce((sum, entry) => sum + (entry as any).value, 0) * 100) / 100,

 // Navigation timing
 domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
 loadComplete: navigation.loadEventEnd - navigation.loadEventStart,

 // Resource timing
 totalResourceTime: performance.getEntriesByType('resource')
 .reduce((sum, entry) => sum + (entry as any).duration, 0)
 };
 });

 return metrics;
 }

 async compareMetrics(baseline: PerformanceMetrics, current: PerformanceMetrics) {
 const comparison = {
 FCP: this.calculateChange(baseline.FCP, current.FCP),
 LCP: this.calculateChange(baseline.LCP, current.LCP),
 CLS: this.calculateChange(baseline.CLS, current.CLS)
 };

 // Alert if degradation > 10%
 if (comparison.LCP.percentageChange > 10) {
 console.warn('LCP degradation detected:', comparison.LCP);
 }

 return comparison;
 }

 private calculateChange(baseline: number, current: number) {
 return {
 baseline,
 current,
 absoluteChange: current - baseline,
 percentageChange: ((current - baseline) / baseline) * 100
 };
 }
}
```

## Accessibility Testing

### 1. Automated Accessibility Checks

```typescript
import { injectAxe, checkA11y } from 'axe-playwright';

test('accessibility compliance', async ({ page }) => {
 await page.goto('https://example.com');

 // Inject axe accessibility testing library
 await injectAxe(page);

 // Check for violations
 await checkA11y(
 page,
 null, // Target all elements
 {
 detailedReport: true,
 detailedReportOptions: {
 html: true
 }
 }
 );
});

test('WCAG 2.1 AA compliance', async ({ page }) => {
 await page.goto('https://example.com');
 await injectAxe(page);

 const accessibility = await page.evaluate(() => {
 return (window as any).axe.run({
 standards: 'wcag21aa',
 resultTypes: ['violations', 'incomplete']
 });
 });

 // Assert no violations
 expect(accessibility.violations).toHaveLength(0);

 // Review incomplete items
 for (const incomplete of accessibility.incomplete) {
 console.log(`Needs manual review: ${incomplete.id}`);
 }
});
```

### 2. Keyboard Navigation Testing

```typescript
test('keyboard navigation', async ({ page }) => {
 await page.goto('https://example.com');

 // Navigate using Tab key
 await page.keyboard.press('Tab');
 let activeElement = await page.evaluate(() => document.activeElement?.tagName);
 expect(['A', 'BUTTON', 'INPUT']).toContain(activeElement);

 // Test keyboard shortcuts
 await page.keyboard.press('Control+K'); // Search shortcut
 const searchBox = page.locator('input[aria-label="Search"]');
 expect(await searchBox.isVisible()).toBe(true);

 // Test Escape key closes modals
 await page.keyboard.press('Escape');
 const modal = page.locator('[role="dialog"]');
 expect(await modal.isVisible()).toBe(false);
});
```

---

Version: 5.0.0 | Last Updated: 2025-11-22 | Enterprise Ready:
