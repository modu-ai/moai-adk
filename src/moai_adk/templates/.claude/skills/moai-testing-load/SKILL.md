---
name: moai-testing-load
version: 4.0.0
updated: 2025-11-19
status: stable
category: Testing
description: Load testing with k6, Gatling, performance benchmarking, and CI/CD integration. Use when testing performance, capacity, or implementing load tests.
allowed-tools:
  - Read
  - Bash
  - WebFetch
  - WebSearch
tags:
  - performance-testing
  - load-testing
  - k6
  - gatling
  - benchmarking
  - ci-cd
---

# Load Testing & Performance Benchmarking

Production-grade load testing with k6 and Gatling for enterprise systems.

## Quick Start

**Choose Your Tool**:

- **k6**: API/microservices testing (JavaScript, simple setup)
- **Gatling**: Complex user journeys, UI testing (Scala DSL)

**5-Minute k6 Example**:

```bash
# Install k6
brew install k6  # macOS
# or: apt install k6  # Linux

# Create test
cat > load-test.js << 'EOF'
import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: 100,
  duration: '30s',
  thresholds: {
    'http_req_duration': ['p(95)<200'],
    'errors': ['rate<0.01'],
  },
};

export default function () {
  const response = http.get('https://httpbin.org/get');
  check(response, {
    'status is 200': (r) => r.status === 200,
  });
}
EOF

# Run test
k6 run load-test.js
```

**Performance Metrics**:

- **p50/p95/p99**: Response time percentiles
- **RPS**: Requests per second
- **Error Rate**: Failed requests percentage

---

## Core Concepts

### Test Types

| Type       | Load            | Purpose                    | Duration  | When                 |
| ---------- | --------------- | -------------------------- | --------- | -------------------- |
| **Smoke**  | Minimal (10 VU) | Verify basic functionality | 1-2 min   | Before deployment    |
| **Load**   | Realistic       | Measure normal performance | 10-15 min | Weekly/release       |
| **Stress** | Maximum         | Find breaking point        | 10-15 min | Before traffic spike |
| **Spike**  | Sudden surge    | Test surge capacity        | 5-10 min  | Flash sales prep     |
| **Soak**   | Sustained       | Find memory leaks          | 8+ hours  | Before holidays      |

### Performance Requirements

```yaml
# Define your SLOs (Service Level Objectives)
api_gateway:
  latency:
    p95: 200ms
    p99: 500ms
  throughput:
    target: 5000 RPS
  availability:
    error_rate: < 0.1%
```

---

## Tool Selection: k6 vs Gatling

| Factor             | k6                  | Gatling              |
| ------------------ | ------------------- | -------------------- |
| **Language**       | JavaScript          | Scala                |
| **Learning Curve** | Low                 | Medium               |
| **Setup**          | Single binary       | JVM required         |
| **Best For**       | APIs, microservices | Complex journeys, UI |
| **Max VU**         | 50K/machine         | 100K+/machine        |
| **Open Source**    | Yes                 | Yes                  |
| **Cloud Service**  | k6 Cloud            | Gatling Enterprise   |

**Decision Tree**:

```
Testing APIs/microservices?
  → YES: Use k6
  → NO: Complex user flows?
    → YES: Use Gatling
    → NO: Use k6 (JavaScript is common)
```

---

## k6 Essentials

### Basic Script Structure

```javascript
import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "2m", target: 100 }, // Ramp up
    { duration: "5m", target: 100 }, // Stay at 100
    { duration: "2m", target: 0 }, // Ramp down
  ],
  thresholds: {
    http_req_duration: ["p(95)<200"],
  },
};

export default function () {
  const res = http.get("https://api.example.com/users");

  check(res, {
    "status 200": (r) => r.status === 200,
    "response time < 200ms": (r) => r.timings.duration < 200,
  });

  sleep(1);
}
```

### Custom Metrics

```javascript
import { Trend, Rate, Counter } from "k6/metrics";

const apiDuration = new Trend("api_duration_ms");
const errorRate = new Rate("errors");
const requests = new Counter("total_requests");

export default function () {
  const res = http.get("https://api.example.com");

  apiDuration.add(res.timings.duration);
  errorRate.add(res.status !== 200);
  requests.add(1);
}
```

### Scenarios

```javascript
export const options = {
  scenarios: {
    baseline: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "1m", target: 50 },
        { duration: "3m", target: 50 },
        { duration: "1m", target: 0 },
      ],
    },
    stress: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "2m", target: 200 },
        { duration: "5m", target: 200 },
      ],
      startTime: "5m", // Start after baseline
    },
  },
};
```

---

## Gatling Essentials

### Basic Simulation

```scala
import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

class BasicSimulation extends Simulation {
  val httpProtocol = http
    .baseUrl("https://api.example.com")
    .acceptHeader("application/json")

  val scn = scenario("Load Test")
    .exec(
      http("Get Users")
        .get("/users")
        .check(status.is(200))
    )
    .pause(1.second)

  setUp(
    scn.inject(
      rampUsers(100).during(5.minutes)
    )
  ).protocols(httpProtocol)
    .assertions(
      global.responseTime.percentile3.lt(200),
      global.failedRequests.percent.lt(0.1)
    )
}
```

---

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/load-test.yml
name: Load Tests

on:
  schedule:
    - cron: "0 2 * * *" # Daily at 2 AM
  workflow_dispatch:

jobs:
  k6-load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run k6
        uses: grafana/k6-action@v0.3.0
        with:
          filename: tests/load-test.js
          flags: --out json=results.json

      - name: Check thresholds
        run: |
          if grep -q '"passed":false' results.json; then
            echo "Performance thresholds failed"
            exit 1
          fi
```

---

## Bottleneck Identification

**Common Bottlenecks**:

1. Database queries (slow joins, missing indexes)
2. External API calls (third-party latency)
3. Network (bandwidth, DNS resolution)
4. CPU (insufficient servers, inefficient algorithms)
5. Memory (GC pauses, memory leaks)

**Debugging Process**:

```
1. Run load test with monitoring
2. Identify slow endpoints (>p95)
3. Profile that component (APM tools)
4. Root cause analysis
5. Fix and re-test
6. Track in baseline
```

**APM Tools**:

- New Relic, Datadog, Elastic APM, Jaeger

---

## Performance Benchmarks

**Typical Response Times**:

```
API endpoint (cached):
  p50: 50ms
  p95: 200ms
  p99: 500ms

Database query:
  p50: 100ms
  p95: 300ms
  p99: 1000ms

E-commerce page:
  p50: 100ms
  p95: 500ms
  p99: 2000ms
```

---

## Best Practices

✅ **DO**:

- Start with smoke tests
- Define SLOs before testing
- Run tests regularly (CI/CD)
- Monitor real user metrics (RUM)
- Test at realistic scale
- Document baseline performance

❌ **DON'T**:

- Test production without permission
- Ignore tail latency (p99)
- Skip ramp-up periods
- Test without monitoring
- Assume performance = functionality

---

## Advanced Topics

For advanced patterns, see:

- **[examples.md](examples.md)**: Complete test scenarios, multi-step flows, authentication
- **[reference.md](reference.md)**: k6/Gatling APIs, distributed testing, performance optimization

**Related Skills**:

- `moai-essentials-perf`: Performance optimization
- `moai-observability-advanced`: APM and monitoring
- `moai-testing-integration`: Integration testing patterns

---

## Troubleshooting

| Issue                | Solution                                    |
| -------------------- | ------------------------------------------- |
| High error rate      | Check server capacity, validate test script |
| Inconsistent results | Ensure stable test environment, use ramp-up |
| Slow test execution  | Increase VU incrementally, check network    |
| Memory issues        | Use distributed testing, optimize script    |

---

**Version**: 4.0.0  
**Last Updated**: 2025-11-19  
**Status**: Production Ready

For detailed examples and API references, see `examples.md` and `reference.md`.
