# Reference: Enterprise Testing Frameworks & Official Documentation

## Unit Testing Frameworks

### pytest 8.4.x — Python

**Official Documentation**:
- [pytest Official Docs](https://docs.pytest.org/en/stable/)
- [pytest Fixtures](https://docs.pytest.org/en/stable/how-to/fixtures.html)
- [pytest Parametrization](https://docs.pytest.org/en/stable/how-to/parametrize.html)
- [pytest Good Practices](https://docs.pytest.org/en/stable/explanation/goodpractices.html)

**Key Libraries**:
- pytest-asyncio: [PyPI](https://pypi.org/project/pytest-asyncio/)
- pytest-cov: [PyPI](https://pypi.org/project/pytest-cov/)
- pytest-mock: [PyPI](https://pypi.org/project/pytest-mock/)
- pytest-xdist: Parallel execution

**Latest Release**: 8.4.2 (November 2025)

---

### Vitest 4.x — JavaScript/TypeScript

**Official Documentation**:
- [Vitest Official](https://vitest.dev/)
- [Vitest Blog   Release](https://vitest.dev/blog/vitest-4)
- [Browser Mode Documentation](https://vitest.dev/guide/browser.html)
- [Visual Regression Testing](https://vitest.dev/guide/visual-regression.html)

**Key Features**:
- Browser mode: Stable (graduated from experimental)
- Type-aware hooks: Full TypeScript support
- Visual regression: Screenshot-based testing
- Playwright integration: Native support

**Latest Release**: 4.0.0 (October 2025)

---

### Jest 30.x — JavaScript Testing

**Official Documentation**:
- [Jest Docs](https://jestjs.io/)
- [Jest Configuration](https://jestjs.io/docs/configuration)
- [Jest Snapshot Testing](https://jestjs.io/docs/snapshot-testing)

**Latest Release**: 30.x (2025)

---

### Mocha 10.x — Node.js Testing

**Official Documentation**:
- [Mocha Docs](https://mochajs.org/)
- [Mocha API](https://mochajs.org/api/mocha)
- [Mocha Examples](https://github.com/mochajs/mocha/tree/master/example)

**Latest Release**: 10.x (Stable)

---

## E2E & Integration Testing

### Playwright 1.48.x

**Official Documentation**:
- [Playwright Official](https://playwright.dev/)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Playwright Multi-Browser Testing](https://playwright.dev/docs/browsers)
- [Playwright Trace Viewer](https://playwright.dev/docs/trace-viewer)
- [Playwright CI/CD Integration](https://playwright.dev/docs/ci)

**Configuration Files**:
- [playwright.config.ts Guide](https://playwright.dev/docs/test-configuration)
- [Reporters](https://playwright.dev/docs/test-reporters)

**Latest Release**: 1.48.x (November 2025)

---

### Cypress 14.x

**Official Documentation**:
- [Cypress Docs](https://docs.cypress.io/)
- [Cypress Best Practices](https://docs.cypress.io/guides/references/best-practices)
- [Cypress Configuration](https://docs.cypress.io/guides/references/configuration)
- [Cypress CI/CD](https://docs.cypress.io/guides/continuous-integration/introduction)

**Latest Release**: 14.x (2025)

---

### Selenium 4.x

**Official Documentation**:
- [Selenium Official](https://www.selenium.dev/)
- [Selenium Python Docs](https://selenium-python.readthedocs.io/)
- [WebDriver API](https://www.selenium.dev/webdriver/)

---

## API Testing

### httpx 0.28.x — Python

**Official Documentation**:
- [HTTPX Official](https://www.python-httpx.org/)
- [HTTPX Async Support](https://www.python-httpx.org/async/)
- [HTTPX Authentication](https://www.python-httpx.org/authentication/)

**Testing Integration**:
- [pytest-httpx](https://colin-b.github.io/pytest_httpx/)
- [respx for mocking](https://lundberg.github.io/respx/)

**Latest Release**: 0.28.1 (November 2025)

---

### Supertest — Node.js

**Official Documentation**:
- [Supertest GitHub](https://github.com/visionmedia/supertest)
- [Supertest NPM](https://www.npmjs.com/package/supertest)

---

## Component & DOM Testing

### Testing Library 15.x

**Official Documentation**:
- [Testing Library Official](https://testing-library.com/)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
- [Vue Testing Library](https://testing-library.com/docs/vue-testing-library/intro/)
- [Svelte Testing Library](https://testing-library.com/docs/svelte-testing-library/intro/)
- [Common Mistakes](https://testing-library.com/docs/queries/about#priority)

**DOM Query Priorities**:
1. getByRole (most accessible)
2. getByLabelText (form fields)
3. getByPlaceholderText
4. getByText
5. getByTestId (last resort)

**Latest Release**: 15.x (2025)

---

### User Event Library

**Official Documentation**:
- [user-event Official](https://testing-library.com/docs/user-event/intro/)
- [user-event API](https://testing-library.com/docs/user-event/intro/)

---

## Test Data & Fixtures

### polyfactory 2.x

**Official Documentation**:
- [polyfactory GitHub](https://github.com/litestar-org/polyfactory)
- [polyfactory Docs](https://docs.litestar.dev/latest/reference/factories.html)
- [PyPI Package](https://pypi.org/project/polyfactory/)

**Features**:
- Pydantic v2 support
- Dataclass generation
- TypedDict support
- Custom handlers

**Latest Release**: 2.x (October 2025)

---

### pytest-factoryboy

**Official Documentation**:
- [pytest-factoryboy](https://pytest-factoryboy.readthedocs.io/)
- [GitHub Repository](https://github.com/pytest-dev/pytest-factoryboy)

---

### factory-boy 3.x

**Official Documentation**:
- [factory-boy Official](https://factory-boy.readthedocs.io/)
- [GitHub Repository](https://github.com/FactoryBoy/factory_boy)

**Supported Frameworks**:
- Django ORM
- SQLAlchemy
- Mongoengine
- Custom support

---

## Test Coverage & Quality

### Coverage.py 7.11.x

**Official Documentation**:
- [Coverage.py Official](https://coverage.readthedocs.io/)
- [Configuration](https://coverage.readthedocs.io/en/latest/config.html)
- [Coverage Reports](https://coverage.readthedocs.io/en/latest/reports.html)

**Latest Release**: 7.11.3 (November 2025)

---

### pytest-cov

**Official Documentation**:
- [pytest-cov](https://pytest-cov.readthedocs.io/)
- [GitHub](https://github.com/pytest-dev/pytest-cov)

---

### c8 — JavaScript Coverage

**Official Documentation**:
- [c8 GitHub](https://github.com/bcoe/c8)
- [NPM Package](https://www.npmjs.com/package/c8)

---

### Istanbul

**Official Documentation**:
- [Istanbul Official](https://istanbul.js.org/)
- [GitHub](https://github.com/istanbuljs/istanbuljs)

---

## Performance & Load Testing

### k6 1.0+ — Load Testing

**Official Documentation**:
- [k6 Official](https://k6.io/)
- [k6 Documentation](https://k6.io/docs/)
- [k6 JavaScript API](https://k6.io/docs/using-k6/javascript-api/)
- [k6 v1.0 Release](https://grafana.com/blog/2025/05/07/grafana-k6-1.0-release/)

**Key Features**:
- TypeScript support (v1.0+)
- Grafana Cloud integration
- k6 Extensions API
- Distributed testing

**Latest Release**: 1.0+ (May 2025)

---

### Grafana k6 Operator

**Official Documentation**:
- [k6 Operator Docs](https://grafana.com/docs/k6/latest/extensions/operator/)
- [GitHub Repository](https://github.com/grafana/k6-operator)

---

### Locust — Python Load Testing

**Official Documentation**:
- [Locust Official](https://locust.io/)
- [Locust Docs](https://docs.locust.io/)

---

### Apache Bench (ab)

**Official Documentation**:
- [Apache Bench](https://httpd.apache.org/docs/current/programs/ab.html)

---

### Artillery

**Official Documentation**:
- [Artillery Official](https://artillery.io/)
- [Artillery Docs](https://artillery.io/docs/)

---

## Accessibility Testing

### axe-core 4.8.x

**Official Documentation**:
- [axe-core GitHub](https://github.com/dequelabs/axe-core)
- [axe-core Rules](https://github.com/dequelabs/axe-core/blob/develop/doc/rule-descriptions.md)
- [WCAG 2.1 Support](https://www.deque.com/blog/support-for-wcag-2-1-in-axe-core/)
- [WCAG 2.2 Support](https://www.deque.com/blog/wcag-2-2-axe-core/)

**Standards Supported**:
- WCAG 2.0 (Level A, AA, AAA)
- WCAG 2.1 (Level A, AA, AAA)
- WCAG 2.2 (Level A, AA, AAA)
- Section 508
- ACT Rules

**Latest Release**: 4.8.x (2025)

---

### @axe-core/playwright

**Official Documentation**:
- [axe-core Playwright](https://github.com/dequelabs/axe-core-npm/tree/develop/packages/playwright)

---

### pa11y 6.x

**Official Documentation**:
- [pa11y Official](https://pa11y.org/)
- [pa11y Docs](https://pa11y.org/document/)

---

### Deque axe DevTools

**Official Documentation**:
- [axe DevTools Browser Extension](https://www.deque.com/axe/devtools/)
- [axe DevTools Documentation](https://www.deque.com/documentation/devtools/)

---

## Mocking & Stubbing

### pytest-mock

**Official Documentation**:
- [pytest-mock Docs](https://pytest-mock.readthedocs.io/)
- [GitHub](https://github.com/pytest-dev/pytest-mock)

---

### unittest.mock — Python Standard Library

**Official Documentation**:
- [unittest.mock Docs](https://docs.python.org/3/library/unittest.mock.html)

---

### sinon.js 18.x

**Official Documentation**:
- [Sinon.JS Official](https://sinonjs.org/)
- [Sinon API Docs](https://sinonjs.org/releases/latest/api/)

---

### Mock Service Worker (MSW) 2.x

**Official Documentation**:
- [MSW Official](https://mswjs.io/)
- [MSW Docs](https://mswjs.io/docs/)
- [GitHub](https://github.com/mswjs/msw)

---

## Testing Best Practices & Resources

### Official Standards & Guidelines

- [WCAG 2.1 Official](https://www.w3.org/WAI/WCAG21/quickref/)
- [WCAG 2.2 Official](https://www.w3.org/WAI/WCAG22/quickref/)
- [Web Content Accessibility Guidelines](https://www.w3.org/WAI/standards-guidelines/)

### Accessibility Testing Resources

- [Deque University](https://dequeuniversity.com/)
- [A11ycasts with Google Chrome](https://www.youtube.com/playlist?list=PLNYkxOF6rcICWx0C9Xc-RgEzwLvePng7V)
- [MDN Accessibility](https://developer.mozilla.org/en-US/docs/Web/Accessibility)

### Performance Testing Resources

- [Web Performance Working Group](https://www.w3.org/webperf/)
- [Performance Best Practices](https://web.dev/performance/)

### Testing Architecture Resources

- [Testing Pyramid (Martin Fowler)](https://martinfowler.com/bliki/TestPyramid.html)
- [Test Automation Strategy](https://martinfowler.com/articles/testing-strategies.html)

---

## Version Matrix (Stable as of November 2025)

| Technology | Version | Status | Python | Node.js | Browsers |
|-----------|---------|--------|--------|---------|----------|
| **pytest** | 8.4.2 | Stable | 3.8-3.13 | N/A | N/A |
| **Vitest** | 4.0+ | Stable | N/A | 18+ | Chromium, Firefox, WebKit |
| **Jest** | 30.x | Stable | N/A | 18+ | Node.js |
| **Playwright** | 1.48.x | Stable | N/A | 18+ | Chrome, Firefox, Safari, Edge |
| **Cypress** | 14.x | Stable | N/A | 18+ | Chrome, Edge, Firefox |
| **Testing Library** | 15.x | Stable | N/A | 18+ | React, Vue, Svelte |
| **httpx** | 0.28.1 | Stable | 3.8+ | N/A | N/A |
| **Coverage.py** | 7.11.3 | Stable | 3.8-3.15 | N/A | N/A |
| **k6** | 1.0+ | Stable | N/A | N/A | N/A |
| **axe-core** | 4.8.x | Stable | N/A | 18+ | All modern |

