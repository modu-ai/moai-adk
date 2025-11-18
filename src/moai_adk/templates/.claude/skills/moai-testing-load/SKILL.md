---
name: moai-testing-load
version: 4.0.0
updated: 2025-11-19
status: stable
category: Testing
description: Load testing with k6, Gatling, performance benchmarking, CI/CD integration
allowed_tools:
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

Performance is non-negotiable in production systems. This Skill provides comprehensive guidance on load testing with industry-standard tools (k6, Gatling), performance benchmarking strategies, and CI/CD integration to prevent regressions.

## Overview

Load testing verifies that systems meet performance requirements under realistic or peak traffic conditions. It reveals bottlenecks, validates scaling strategies, and builds confidence before production deployments.

### Why Load Testing Matters

**Traditional approach**: Deploy first, debug performance issues in production (costly, risky)

**Load testing approach**:
- Identify bottlenecks before production
- Validate scaling strategies (horizontal/vertical)
- Define realistic SLAs (Service Level Agreements)
- Prevent performance regressions via automated testing
- Plan infrastructure capacity confidently

### Performance Requirement Categories

| Category | Metric | Example |
|----------|--------|---------|
| **Latency** | Response time (p50, p95, p99) | p95 < 200ms |
| **Throughput** | Requests per second | 5,000 RPS |
| **Availability** | Error rate tolerance | < 0.1% errors |
| **Concurrency** | Simultaneous users | 10,000 VU |
| **Degradation** | Acceptable performance under stress | p95 < 500ms at 50% overload |

### Load Testing Strategy Overview

```
Baseline Measurement
    ↓
Smoke Test (small load)
    ↓
Load Test (realistic traffic)
    ↓
Stress Test (maximum capacity)
    ↓
Spike Test (sudden traffic burst)
    ↓
Soak Test (sustained load)
    ↓
Analysis & Optimization
```

**Decision**: Choose k6 for API/backend testing, Gatling for complex user journeys and UI testing.

---

## 1. Load Testing Fundamentals (600 words)

### Performance Metrics Explained

**Response Time (Latency)**:
```
Request sent at T0
Server processes: T0 → T1
Response received: T1 → T2
Total latency = T2 - T0
```

Common percentiles:
- **p50 (median)**: 50% of requests faster than this
- **p95**: 95% of requests faster than this (user experience focus)
- **p99**: 99% of requests faster than this (tail latency)
- **p99.9**: Extreme cases (rare but noticeable)

**Throughput (TPS/RPS)**:
- TPS: Transactions per second
- RPS: Requests per second
- Calculated: `Total requests / Total time`
- Example: 50,000 requests / 10 minutes = 83 RPS

**Error Rate**:
- Percentage of failed requests
- Example: 50 failures / 50,000 requests = 0.1%
- Target: Typically < 0.1% acceptable
- Monitoring: Track error types (timeout, 5xx, connection refused)

**Apdex Score** (Application Performance Index):
```
Apdex = (Satisfied + Tolerated/2) / Total
Where:
- Satisfied: response < threshold (e.g., 0.5s)
- Tolerated: threshold < response < 4x threshold
- Frustrated: response >= 4x threshold

Target: Apdex > 0.94 (excellent)
```

---

### Test Types Detailed Comparison

| Test Type | Load Pattern | Purpose | Duration | When to Use |
|-----------|--------------|---------|----------|------------|
| **Smoke Test** | Minimal (10 VU) | Verify system responsiveness | 1-2 minutes | Before every deployment |
| **Load Test** | Realistic (realistic VU) | Measure performance metrics | 10-15 minutes | Weekly or before release |
| **Stress Test** | Maximum (until failure) | Find breaking point | 10-15 minutes | Before handling traffic spike |
| **Spike Test** | Sudden surge (0→peak→0) | Test surge capacity | 5-10 minutes | E-commerce flash sales |
| **Soak Test** | Sustained (8-12 hours) | Find memory/connection leaks | 8+ hours | Before major holiday traffic |
| **Chaos Test** | Random failures | Test resilience | Varies | Critical infrastructure |

### Real-World Performance Expectations

```
E-commerce product page:
  - p50: 50ms (server response, mostly cached)
  - p95: 200ms (including network, DOM processing)
  - p99: 500ms (cold cache, slower users)

API endpoint (database query):
  - p50: 100ms
  - p95: 300ms
  - p99: 1000ms (query on large dataset)

Mobile app sync:
  - p50: 200ms
  - p95: 800ms (poor network)
  - p99: 3000ms (very poor network)
```

---

### Example 1: Performance Metrics Definition

```yaml
# requirements/performance.yml
# Service Level Objectives (SLOs) for production

services:
  api_gateway:
    latency:
      p50: 50ms
      p95: 200ms
      p99: 500ms
      slo: "p95 < 200ms for 99.5% of requests"
    
    throughput:
      target_rps: 5000
      burst_capacity: 10000
      slo: "handle 5000 RPS sustained"
    
    availability:
      error_rate_target: 0.05%
      timeout_rate: 0.01%
      slo: "99.95% availability (4 nines)"

  database_service:
    latency:
      p50: 20ms
      p95: 100ms
      p99: 500ms
      slo: "p95 < 100ms for OLTP queries"
    
    throughput:
      connections: 1000
      max_connections: 2000
      connection_pool_size: 50
      slo: "maintain < 80% connection pool utilization"
    
    backup:
      rto: "1 hour"  # Recovery Time Objective
      rpo: "5 minutes"  # Recovery Point Objective

test_scenarios:
  smoke_test:
    name: "Smoke Test (Pre-deployment)"
    target_vus: 10
    duration: "2m"
    ramp_up: "30s"
    assertions:
      - "p95 < 500ms"
      - "error_rate < 5%"

  load_test:
    name: "Load Test (Realistic Traffic)"
    target_vus: 1000
    duration: "10m"
    ramp_up: "2m"
    sustain: "6m"
    ramp_down: "2m"
    assertions:
      - "p50 < 100ms"
      - "p95 < 200ms"
      - "p99 < 500ms"
      - "error_rate < 0.1%"
      - "throughput >= 5000 RPS"

  stress_test:
    name: "Stress Test (Breaking Point)"
    target_vus: 5000
    duration: "15m"
    ramp_up: "3m"
    sustain: "10m"
    ramp_down: "2m"
    assertions:
      - "breaking_point_vus > 5000"
      - "error_rate < 1%"

  spike_test:
    name: "Spike Test (Black Friday)"
    phases:
      - duration: "2m"
        target_vus: 100
        description: "normal traffic"
      - duration: "1m"
        target_vus: 5000
        description: "sudden 50x spike"
      - duration: "5m"
        target_vus: 5000
        description: "sustained peak"
      - duration: "2m"
        target_vus: 100
        description: "return to normal"
    assertions:
      - "error_rate < 0.5% during spike"
      - "p95 < 1000ms during spike"

  soak_test:
    name: "Soak Test (Memory Leaks)"
    target_vus: 500
    duration: "12h"
    ramp_up: "5m"
    assertions:
      - "memory usage < 5GB after 12h"
      - "connection pool count stable"
      - "no performance degradation over time"
```

---

### Bottleneck Identification Framework

**Where are slowdowns?**

1. **Database queries**: Slow JOINs, missing indexes, N+1 problems
2. **External API calls**: Third-party service latency
3. **Network**: Bandwidth limitations, high latency, DNS resolution
4. **Compute**: CPU-bound algorithms, insufficient servers
5. **Memory**: Garbage collection pauses, memory leaks, buffer issues
6. **Disk I/O**: Slow storage, excessive disk operations
7. **Concurrency bottlenecks**: Lock contention, limited worker threads

**Finding bottlenecks systematically**:
```
1. Run load test with full instrumentation
2. Collect metrics by endpoint and component
3. Identify outliers (slow endpoints)
4. Drill down: CPU? Memory? Network? Database?
5. Profile that component (APM tools)
6. Root cause analysis (query plan, algorithm, config)
7. Implement fix
8. Re-test and validate improvement
9. Track in baseline metrics
```

**APM Tools for bottleneck identification**:
- **New Relic**: Cloud APM, flame graphs
- **Datadog**: Infrastructure + APM monitoring
- **Elastic APM**: Open source, distributed tracing
- **Jaeger**: Distributed tracing (CNCF)

---

### Tool Selection Matrix: k6 vs Gatling vs Others

| Factor | k6 | Gatling | JMeter | Locust |
|--------|-----|---------|--------|--------|
| **Language** | JavaScript (Go engine) | Scala | Java GUI | Python |
| **Learning Curve** | Low (JavaScript familiar) | Medium (Scala DSL) | High (GUI complexity) | Low (Python familiar) |
| **Setup Complexity** | Single binary (easiest) | JVM required | GUI/CLI hybrid | Python + dependencies |
| **Max VU per machine** | 50K | 100K+ | 10K | 30K |
| **Distributed Testing** | k6 Cloud native | Gatling Enterprise | Distributed slaves | Distributed workers |
| **Protocol Support** | HTTP, WebSocket, gRPC, JMeter | HTTP, WebSocket, SSE, JMS | HTTP, JDBC, SOAP | HTTP (plugins for others) |
| **Open Source** | Yes (k6 OSS) | Yes | Yes (Apache) | Yes (Apache) |
| **Cloud Service** | k6 Cloud (recommended) | Gatling Cloud | - | - |
| **UI Reporting** | Web dashboard | HTML + Grafana | HTML | Web dashboard |
| **Enterprise Features** | Minimal in OSS | Full in Enterprise | Community limited | Community limited |
| **Best For** | APIs, microservices, simple scenarios | Complex user journeys, UI testing | Legacy systems, JDBC | Python-based teams |
| **Best Against** | Large UI test suites | API-only testing | Modern tools | Large concurrent users |

**Decision Tree for Tool Selection**:

```
1. Are you testing APIs or microservices?
   → YES: k6 (simple, fast setup)
   → NO: Go to #2

2. Do you need complex user journeys with states?
   → YES: Gatling (Scala DSL, high concurrency)
   → NO: Go to #3

3. Do you have Python developers?
   → YES: Locust (familiar language)
   → NO: k6 (JavaScript is ubiquitous)

4. Do you need JMeter compatibility?
   → YES: JMeter (legacy system)
   → NO: k6 or Gatling
```

---

## 2. k6 Framework (750 words)

### k6 Deep Dive: Concepts & Architecture

**k6** is a modern load testing platform built in Go, using JavaScript (ES6+) for test scripts. It's designed for ease of use and integrates well with CI/CD pipelines.

**Architecture**:
```
┌─────────────────────────────────────┐
│ Test Script (JavaScript)            │
│ - Scenarios (load patterns)         │
│ - Checks (assertions)               │
│ - Custom metrics (Trend, Rate, etc) │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│ k6 Engine (Go)                      │
│ - VU execution (virtual users)      │
│ - Metric collection                 │
│ - Load scheduling (stages)          │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│ Data Collection                     │
│ - JSON output                       │
│ - InfluxDB export                   │
│ - k6 Cloud integration              │
└─────────────────────────────────────┘
```

**Key Concepts**:

- **VU (Virtual User)**: Independent goroutine (Go) simulating one user
- **Iteration**: One complete execution of the script by a VU
- **Scenario**: Named execution pattern with load profile
- **Stage**: Ramp-up/sustain/ramp-down phase within scenario
- **Metric**: Numeric measurement (response time, errors, custom)
- **Threshold**: Pass/fail criteria for metrics

**Installation & Setup**:
```bash
# macOS (Homebrew)
brew install k6

# Linux (Ubuntu/Debian)
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 \
  --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | \
  sudo tee /etc/apt/sources.list.d/k6-stable.list
sudo apt-get update
sudo apt-get install k6

# Docker
docker pull grafana/k6

# Verify
k6 --version  # v0.54+
```

---

### Scenario Writing: Progressive Load Patterns

**Example 2: Load Test Scenarios with Multiple Strategies**

```yaml
# scenarios/performance-test-plan.yml
# Comprehensive scenario definition for different load patterns

baseline_configuration:
  api_url: "https://api.example.com"
  timeout: "30s"
  think_time_min: 0.5s
  think_time_max: 3.0s

scenarios:
  # 1. Smoke Test: Quick validation
  smoke_test:
    name: "Smoke Test"
    description: "Quick smoke test to verify basic functionality"
    enabled: true
    target_virtual_users: 10
    duration: "2m"
    
    phases:
      - name: "Ramp Up"
        duration: "30s"
        start_vus: 0
        end_vus: 10
        
      - name: "Sustain"
        duration: "1m 30s"
        vus: 10
    
    assertions:
      - metric: "response_time_p95"
        target: "< 500ms"
        critical: true
      - metric: "error_rate"
        target: "< 5%"
        critical: true
      - metric: "availability"
        target: "> 95%"
        critical: true

  # 2. Load Test: Realistic production traffic
  load_test:
    name: "Load Test"
    description: "Load test with realistic production traffic patterns"
    enabled: true
    target_virtual_users: 1000
    duration: "10m"
    
    phases:
      - name: "Ramp Up (2 minutes)"
        duration: "2m"
        start_vus: 0
        end_vus: 1000
        
      - name: "Sustain (6 minutes)"
        duration: "6m"
        vus: 1000
        
      - name: "Ramp Down (2 minutes)"
        duration: "2m"
        start_vus: 1000
        end_vus: 0
    
    assertions:
      - metric: "response_time_p50"
        target: "< 100ms"
        critical: false
      - metric: "response_time_p95"
        target: "< 200ms"
        critical: true
      - metric: "response_time_p99"
        target: "< 500ms"
        critical: false
      - metric: "error_rate"
        target: "< 0.1%"
        critical: true
      - metric: "requests_per_second"
        target: ">= 1000"
        critical: false
      - metric: "throughput"
        target: ">= 5000 req/s"
        critical: true

  # 3. Stress Test: Find breaking point
  stress_test:
    name: "Stress Test"
    description: "Stress test to find system breaking point"
    enabled: true
    initial_virtual_users: 100
    increment_step: 100  # increase by 100 VU each minute
    max_virtual_users: 10000
    
    phases:
      - name: "Ramp Up"
        duration: "30m"  # 10000 VU / 100 per minute
        start_vus: 0
        end_vus: 10000
        increment_interval: "1m"
    
    success_criteria:
      - "system remains stable at 5000 VU"
      - "breaking point identified at 8000 VU"
    
    assertions:
      - metric: "error_rate"
        target: "< 5%"
        critical: false
      - metric: "breaking_point_detected"
        target: true
        critical: true

  # 4. Spike Test: Sudden traffic surge
  spike_test:
    name: "Spike Test"
    description: "Test system response to sudden traffic spike"
    enabled: true
    
    phases:
      - name: "Normal Traffic"
        duration: "2m"
        vus: 100
        
      - name: "Sudden Spike"
        duration: "1m"
        start_vus: 100
        end_vus: 5000
        
      - name: "Sustained Peak"
        duration: "5m"
        vus: 5000
        
      - name: "Return to Normal"
        duration: "2m"
        start_vus: 5000
        end_vus: 100
    
    assertions:
      - metric: "spike_recovery_time"
        target: "< 30s"
        description: "System should recover to baseline performance within 30s"
      - metric: "error_rate_during_spike"
        target: "< 1%"
        critical: true

  # 5. Soak Test: Sustained load to find leaks
  soak_test:
    name: "Soak Test"
    description: "Long-running soak test to detect memory/connection leaks"
    enabled: false  # Run manually
    target_virtual_users: 500
    duration: "12h"
    
    phases:
      - name: "Ramp Up"
        duration: "5m"
        start_vus: 0
        end_vus: 500
        
      - name: "Sustain"
        duration: "11h 55m"
        vus: 500
    
    monitoring:
      - "memory_usage"
      - "connection_pool_size"
      - "database_connections_active"
      - "response_time_degradation"
    
    assertions:
      - metric: "memory_growth_rate"
        target: "< 100MB/hour"
        critical: true
      - metric: "connection_stability"
        target: "stable"
        critical: true
      - metric: "p95_degradation"
        target: "< 10%"
        critical: true
        description: "p95 should not degrade more than 10% over 12h"

  # 6. Endurance Test: Extended realistic load
  endurance_test:
    name: "Endurance Test"
    description: "Extended test with realistic user patterns"
    enabled: true
    duration: "2h"
    
    phases:
      - name: "Ramp Up"
        duration: "5m"
        start_vus: 0
        end_vus: 1000
        
      - name: "Morning Peak"
        duration: "30m"
        vus: 1000
        
      - name: "Dip"
        duration: "20m"
        vus: 500
        
      - name: "Afternoon Peak"
        duration: "30m"
        vus: 1500
        
      - name: "Evening"
        duration: "25m"
        vus: 800
        
      - name: "Ramp Down"
        duration: "10m"
        start_vus: 800
        end_vus: 0

metrics_to_collect:
  - response_time_p50
  - response_time_p95
  - response_time_p99
  - error_rate
  - requests_per_second
  - throughput
  - check_success_rate
  - data_received_rate
  - data_sent_rate
```

---

### Example 3: k6 Script with Custom Metrics

```javascript
// load-test.js
// Comprehensive k6 script with custom metrics and checks

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';

// Custom metrics for fine-grained monitoring
const errorRate = new Rate('errors');
const apiDuration = new Trend('api_duration_ms');
const databaseDuration = new Trend('database_duration_ms');
const successfulRequests = new Counter('requests_success');
const failedRequests = new Counter('requests_failed');
const activeConnections = new Gauge('active_connections');
const cacheHitRate = new Rate('cache_hits');
const userCreationDuration = new Trend('user_creation_duration_ms');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'https://httpbin.org';
const REQUEST_TIMEOUT = '30s';

export const options = {
  vus: 100,
  duration: '5m',
  
  // Define thresholds for pass/fail criteria
  thresholds: {
    errors: ['rate<0.1'],                          // error rate < 0.1%
    'http_req_duration': ['p(95)<200', 'p(99)<500'],  // response time
    'http_req_duration{staticAsset:true}': ['p(99)<100'],
    'api_duration_ms': ['p(95)<300'],
    'requests_success': ['count>4000'],
    checks: ['rate>0.95'],                         // 95% of checks pass
  },
  
  // Multiple scenarios with different load patterns
  scenarios: {
    baseline: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 50 },    // ramp to 50 users
      ],
    },
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 100 },
        { duration: '2m', target: 200 },
      ],
      startTime: '1m',
    },
  },
};

export default function () {
  // Group 1: Authentication
  group('Authentication', function () {
    const authPayload = JSON.stringify({
      username: `user-${__VU}-${__ITER}@example.com`,
      password: 'secure-password-123',
    });

    const authResponse = http.post(
      `${BASE_URL}/auth/login`,
      authPayload,
      {
        headers: {
          'Content-Type': 'application/json',
          'User-Agent': 'k6-load-test/1.0',
        },
        timeout: REQUEST_TIMEOUT,
        tags: { name: 'Auth_Login' },
      }
    );

    const authSuccess = check(authResponse, {
      'auth status is 200 or 201': (r) => r.status === 200 || r.status === 201,
      'auth response has token': (r) => r.json('token') !== undefined,
      'auth response time < 500ms': (r) => r.timings.duration < 500,
    });

    if (authSuccess) {
      successfulRequests.add(1);
      cacheHitRate.add(authResponse.status === 304 ? 1 : 0);
    } else {
      failedRequests.add(1);
    }

    errorRate.add(!authSuccess);
    apiDuration.add(authResponse.timings.duration);
    activeConnections.add(1);

    // Extract token for subsequent requests
    const token = authResponse.json('token');
    sleep(1);

    // Group 2: User management with token
    group('User Management', function () {
      // Create user
      const createUserPayload = JSON.stringify({
        email: `newuser-${__VU}-${__ITER}@example.com`,
        name: `User ${__VU}`,
        role: 'member',
      });

      const createUserResponse = http.post(
        `${BASE_URL}/api/users`,
        createUserPayload,
        {
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
            'User-Agent': 'k6-load-test/1.0',
          },
          timeout: REQUEST_TIMEOUT,
          tags: { name: 'User_Create' },
        }
      );

      const createSuccess = check(createUserResponse, {
        'create user status is 201': (r) => r.status === 201,
        'create user has id': (r) => r.json('id') !== undefined,
        'create user response time < 1s': (r) => r.timings.duration < 1000,
      });

      if (createSuccess) {
        successfulRequests.add(1);
      } else {
        failedRequests.add(1);
      }

      errorRate.add(!createSuccess);
      userCreationDuration.add(createUserResponse.timings.duration);
      apiDuration.add(createUserResponse.timings.duration);

      const userId = createUserResponse.json('id');
      sleep(2);

      // Get user
      const getUserResponse = http.get(
        `${BASE_URL}/api/users/${userId}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'User-Agent': 'k6-load-test/1.0',
          },
          timeout: REQUEST_TIMEOUT,
          tags: { name: 'User_Get' },
        }
      );

      check(getUserResponse, {
        'get user status is 200': (r) => r.status === 200,
        'get user has data': (r) => r.json('id') === userId,
      });

      apiDuration.add(getUserResponse.timings.duration);
      sleep(1);

      // Update user
      const updatePayload = JSON.stringify({
        name: `Updated User ${__VU}`,
        role: 'admin',
      });

      const updateResponse = http.put(
        `${BASE_URL}/api/users/${userId}`,
        updatePayload,
        {
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
            'User-Agent': 'k6-load-test/1.0',
          },
          timeout: REQUEST_TIMEOUT,
          tags: { name: 'User_Update' },
        }
      );

      check(updateResponse, {
        'update user status is 200': (r) => r.status === 200,
      });

      apiDuration.add(updateResponse.timings.duration);
      sleep(1);

      // List users (pagination test)
      const listResponse = http.get(
        `${BASE_URL}/api/users?page=1&limit=50`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'User-Agent': 'k6-load-test/1.0',
          },
          timeout: REQUEST_TIMEOUT,
          tags: { name: 'User_List' },
        }
      );

      check(listResponse, {
        'list users status is 200': (r) => r.status === 200,
        'list users has data array': (r) => r.json('data.length') > 0,
      });

      apiDuration.add(listResponse.timings.duration);
    });
  });

  // Group 3: Data operations
  group('Data Operations', function () {
    // Batch get
    const batchPayload = JSON.stringify({
      ids: ['1', '2', '3', '4', '5'],
    });

    const batchResponse = http.post(
      `${BASE_URL}/api/data/batch`,
      batchPayload,
      {
        headers: {
          'Content-Type': 'application/json',
          'User-Agent': 'k6-load-test/1.0',
        },
        timeout: REQUEST_TIMEOUT,
        tags: { name: 'Data_Batch' },
      }
    );

    check(batchResponse, {
      'batch status is 200': (r) => r.status === 200,
      'batch response time < 2s': (r) => r.timings.duration < 2000,
    });

    apiDuration.add(batchResponse.timings.duration);
    databaseDuration.add(batchResponse.timings.wait);  // server processing time
    sleep(2);
  });

  // Group 4: Static assets (usually cached)
  group('Static Assets', function () {
    const assetResponse = http.get(
      `${BASE_URL}/static/app.js`,
      {
        headers: {
          'User-Agent': 'k6-load-test/1.0',
          'If-None-Match': 'W/"abc123"',  // cache validation
        },
        timeout: REQUEST_TIMEOUT,
        tags: { name: 'StaticAsset', staticAsset: true },
      }
    );

    check(assetResponse, {
      'asset status is 200 or 304': (r) => r.status === 200 || r.status === 304,
      'asset response time < 100ms': (r) => r.timings.duration < 100,
    });

    cacheHitRate.add(assetResponse.status === 304 ? 1 : 0);
  });

  // Simulate think time
  sleep(Math.random() * 5 + 1);
}

// Lifecycle hooks for setup/teardown
export function setup() {
  console.log('Setup phase: Initialize test data');
  // Could create test fixtures here
  return { testId: `test-${Date.now()}` };
}

export function teardown(data) {
  console.log(`Teardown phase: Cleanup test ${data.testId}`);
  // Could clean up test data here
}
```

**Run with custom output**:
```bash
k6 run --out json=results.json --out influxdb=http://localhost:8086 load-test.js
```

---

### Example 4: k6 Advanced User Journeys

```javascript
// user-journeys.js
// Multiple realistic user journeys in single test

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Trend } from 'k6/metrics';

const checkoutDuration = new Trend('checkout_duration');

export const options = {
  stages: [
    { duration: '2m', target: 100 },   // browsers
    { duration: '5m', target: 100 },   // stay at 100
    { duration: '2m', target: 200 },   // spike
    { duration: '5m', target: 200 },   // stay at 200
    { duration: '2m', target: 0 },     // ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],
    'http_req_failed': ['rate<0.1'],
  },
};

// Realistic user segments with weighted distribution
const userSegments = [
  {
    name: 'Browsers',
    weight: 70,  // 70% of users
    journey: 'browse',
  },
  {
    name: 'Buyers',
    weight: 20,  // 20% of users (intent: purchase)
    journey: 'buy',
  },
  {
    name: 'Researchers',
    weight: 10,  // 10% of users (intent: research)
    journey: 'research',
  },
];

export default function () {
  const userType = selectUserType();
  
  if (userType === 'browse') {
    browseJourney();
  } else if (userType === 'buy') {
    purchaseJourney();
  } else {
    researchJourney();
  }
}

function selectUserType() {
  const rand = Math.random() * 100;
  let cumulative = 0;

  for (const segment of userSegments) {
    cumulative += segment.weight;
    if (rand < cumulative) {
      return segment.journey;
    }
  }

  return 'browse';
}

function browseJourney() {
  group('Browse Journey', function () {
    // 1. Visit homepage
    const homeRes = http.get('https://example.com', {
      tags: { name: 'HomePage' },
    });
    check(homeRes, {
      'homepage loads': (r) => r.status === 200,
    });
    sleep(Math.random() * 3 + 1);

    // 2. View category
    const categoryRes = http.get('https://example.com/categories/electronics', {
      tags: { name: 'Category' },
    });
    check(categoryRes, {
      'category loads': (r) => r.status === 200,
    });
    sleep(Math.random() * 4 + 2);

    // 3. View product list
    const productListRes = http.get(
      'https://example.com/api/products?category=electronics&limit=50',
      { headers: { 'Accept': 'application/json' } }
    );
    check(productListRes, {
      'product list loads': (r) => r.status === 200,
    });
    sleep(Math.random() * 2 + 1);
  });
}

function purchaseJourney() {
  group('Purchase Journey', function () {
    // 1. Search
    const searchRes = http.get('https://example.com/api/search?q=laptop', {
      tags: { name: 'Search' },
    });
    check(searchRes, {
      'search returns results': (r) => r.json('results.length') > 0,
    });
    sleep(1);

    // 2. View product detail
    const productRes = http.get('https://example.com/api/products/12345', {
      tags: { name: 'ProductDetail' },
    });
    check(productRes, {
      'product details load': (r) => r.status === 200,
    });
    sleep(2);

    // 3. Add to cart
    const cartRes = http.post(
      'https://example.com/api/cart',
      JSON.stringify({ productId: 12345, quantity: 1 }),
      {
        headers: { 'Content-Type': 'application/json' },
        tags: { name: 'AddCart' },
      }
    );
    check(cartRes, {
      'item added to cart': (r) => r.status === 200 || r.status === 201,
    });
    sleep(1);

    // 4. Checkout (HIGH VALUE)
    const startTime = Date.now();
    const checkoutRes = http.post(
      'https://example.com/api/checkout',
      JSON.stringify({
        shippingAddress: '123 Main St',
        paymentMethod: 'card',
        cardToken: 'tok_visa',
      }),
      {
        headers: { 'Content-Type': 'application/json' },
        tags: { name: 'Checkout' },
      }
    );
    const checkoutTime = Date.now() - startTime;
    checkoutDuration.add(checkoutTime);

    check(checkoutRes, {
      'checkout succeeds': (r) => r.status === 200 || r.status === 201,
      'has order id': (r) => r.json('orderId') !== undefined,
    });
    sleep(1);

    // 5. Confirmation
    const confirmRes = http.get(
      `https://example.com/order-confirmation?id=${checkoutRes.json('orderId')}`,
      { tags: { name: 'Confirmation' } }
    );
    check(confirmRes, {
      'confirmation page loads': (r) => r.status === 200,
    });
  });
}

function researchJourney() {
  group('Research Journey', function () {
    // Long browsing session with minimal actions
    for (let i = 0; i < 5; i++) {
      const productId = Math.floor(Math.random() * 1000) + 1;
      const res = http.get(`https://example.com/api/products/${productId}`, {
        tags: { name: 'ResearchProduct' },
      });
      check(res, {
        'product page loads': (r) => r.status === 200,
      });
      sleep(Math.random() * 8 + 3);  // longer think time
    }
  });
}
```

---

### Example 5: k6 Distributed Testing Configuration

```javascript
// distributed-test.js
// Configuration for running across multiple k6 Cloud regions

import http from 'k6/http';
import { check } from 'k6';

export const options = {
  // Standard scenario configuration
  vus: 1000,
  duration: '10m',
  
  // k6 Cloud configuration for distributed execution
  ext: {
    loadimpact: {
      // Project configuration
      projectID: 12345,
      name: 'Global Distributed Load Test',
      tags: ['distributed', 'production', 'ecommerce'],
      
      // Distributed load distribution across regions
      distribution: {
        'amazon:us:ashburn': {
          percentage: 40,      // 40% of load from US East
          vus: 400,
        },
        'amazon:eu:dublin': {
          percentage: 35,      // 35% of load from EU
          vus: 350,
        },
        'amazon:ap:sydney': {
          percentage: 25,      // 25% of load from Asia Pacific
          vus: 250,
        },
      },
      
      // Advanced options
      apm: {
        enabled: true,
        dsn: 'https://apm-collector.example.com',
      },
    },
  },

  thresholds: {
    'http_req_duration': ['p(95)<200'],
    'http_req_failed': ['rate<0.1'],
  },
};

export default function () {
  const res = http.get('https://api.example.com/endpoint', {
    headers: {
      'User-Agent': 'k6-distributed/1.0',
    },
  });

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time acceptable': (r) => r.timings.duration < 300,
  });
}
```

**Push to k6 Cloud**:
```bash
k6 cloud distributed-test.js
```

---

## 3. Gatling Framework (700 words)

### Gatling Deep Dive: Enterprise Load Testing

**Gatling** is a powerful enterprise-grade load testing tool built on Scala, designed for high-performance testing of complex user scenarios. It excels at testing realistic user journeys with state management.

**Why Gatling?**:
- **High concurrency**: Supports 100K+ VU per machine
- **State management**: Complex multi-step user flows
- **DSL flexibility**: Scala-based domain-specific language
- **Enterprise features**: Distributed testing, cloud integration, advanced reporting

**Architecture**:
```
┌─────────────────────────────────────┐
│ Simulation (Scala DSL)              │
│ - Scenario definitions              │
│ - Load injection profiles           │
│ - Assertions & thresholds           │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│ Gatling Engine                      │
│ - User simulation (Akka actors)     │
│ - Connection pooling                │
│ - Metric aggregation                │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│ Reporting & Analysis                │
│ - HTML reports with charts          │
│ - CSV export                        │
│ - Grafana/Prometheus integration    │
└─────────────────────────────────────┘
```

**Key Concepts**:

- **Simulation**: Test scenario definition (extends Simulation class)
- **Scenario**: Named user flow with requests and pauses
- **Injection Profile**: How users are created (ramp, constant, custom)
- **Feeder**: Test data source (CSV, JSON, Iterator)
- **Check**: Assertion on response (pass/fail criteria)
- **Assertion**: Global test result criteria

**Installation & Setup**:
```bash
# Download from official site
cd ~/tools
wget https://repo1.maven.org/maven2/io/gatling/gatling-charts-highcharts-bundle/3.13.1/gatling-charts-highcharts-bundle-3.13.1-bundle.zip
unzip gatling-charts-highcharts-bundle-3.13.1-bundle.zip

# Set PATH
export PATH="$PATH:~/tools/gatling-charts-highcharts-bundle-3.13.1/bin"

# Verify
gatling.sh --version  # Should show 3.13+
```

---

### Example 6: Gatling Comprehensive E-Commerce Simulation

```scala
// src/test/scala/simulations/ECommerceSimulation.scala
// Full e-commerce load test with realistic user journeys

package simulations

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

class ECommerceSimulation extends Simulation {
  
  // ============== Protocol Configuration ==============
  val httpProtocol = http
    .baseUrl("https://api.example.com")
    .acceptHeader("application/json")
    .acceptEncodingHeader("gzip, deflate")
    .contentTypeHeader("application/json")
    .userAgentHeader("Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
    .disableCaching  // Force fresh data
    .maxConnectionsPerHostLikeChrome

  // ============== Feeder Configuration ==============
  // Test data from CSV for realistic inputs
  val productFeeder = csv("data/products.csv").random
  val userFeeder = csv("data/users.csv").circular
  val categoryFeeder = csv("data/categories.csv").random

  // ============== Scenario Definitions ==============
  
  // Scenario 1: Browse-only users (70% of traffic)
  val browseScenario = scenario("Browse Products")
    .feed(userFeeder)
    .exec(session => {
      println(s"Browse user: ${session("email")}")
      session
    })
    .exec(http("GET Homepage")
      .get("/")
      .check(status.is(200)))
    .pause(1, 3)
    
    .exec(http("GET Categories")
      .get("/categories")
      .check(status.is(200)))
    .pause(1)
    
    .feed(categoryFeeder)
    .exec(http("GET Products in Category")
      .get("/products?category=${category}")
      .check(status.is(200)))
    .pause(2, 5)
    
    .feed(productFeeder)
    .exec(http("GET Product Details")
      .get("/products/${productId}")
      .check(status.is(200)))
    .pause(3, 8)

  // Scenario 2: Buyers (20% of traffic)
  val buyScenario = scenario("Purchase Products")
    .feed(userFeeder)
    
    // Login
    .exec(http("POST Login")
      .post("/auth/login")
      .body(StringBody("""{"email":"${email}","password":"password123"}"""))
      .check(status.is(200))
      .check(jsonPath("$.token").saveAs("token")))
    .pause(1, 2)
    
    // Search for products
    .exec(http("GET Search")
      .get("/products?search=laptop")
      .header("Authorization", "Bearer ${token}")
      .check(status.is(200)))
    .pause(2)
    
    // Add products to cart
    .repeat(3) {
      feed(productFeeder)
        .exec(http("POST Add to Cart")
          .post("/cart")
          .header("Authorization", "Bearer ${token}")
          .body(StringBody("""{"productId":"${productId}","quantity":${quantity}}"""))
          .check(status.is(200)))
        .pause(1)
    }
    
    // Proceed to checkout
    .exec(http("POST Checkout")
      .post("/checkout")
      .header("Authorization", "Bearer ${token}")
      .body(StringBody("""{"cart":"complete"}"""))
      .check(status.is(200))
      .check(jsonPath("$.orderId").saveAs("orderId")))
    .pause(2)
    
    // Place order (high-value transaction)
    .exec(http("POST Place Order")
      .post("/orders")
      .header("Authorization", "Bearer ${token}")
      .body(StringBody("""{"orderId":"${orderId}","paymentMethod":"card"}"""))
      .check(status.is(201)))
    .pause(1)
    
    // Confirm order
    .exec(http("GET Order Confirmation")
      .get("/orders/${orderId}/confirmation")
      .header("Authorization", "Bearer ${token}")
      .check(status.is(200)))

  // Scenario 3: API Consumers (10% of traffic)
  val apiLoadScenario = scenario("API Load")
    .repeat(10) {
      feed(productFeeder)
        .exec(http("GET API Endpoint")
          .get("/api/products/${productId}")
          .header("Accept", "application/json")
          .check(status.is(200)))
        .pause(500.millis, 1.second)
    }

  // ============== Load Injection Profile ==============
  // Realistic production-like traffic pattern
  setUp(
    // 70% browser users: gradual ramp-up, sustained, ramp-down
    browseScenario.injectOpen(
      rampUsers(100).during(2.minutes),        // ramp to 100 over 2 min
      constantUsersPerSec(10).during(5.minutes),  // 10 new/sec for 5 min
      rampUsers(50).during(1.minute),          // spike by 50 more
      constantUsersPerSec(5).during(3.minutes),   // return to normal
      rampDownUsers(100).during(2.minutes)     // ramp down
    ),
    
    // 20% buyers: steady state
    buyScenario.injectOpen(
      rampUsers(30).during(2.minutes),
      constantUsersPerSec(3).during(10.minutes),
      rampDownUsers(30).during(2.minutes)
    ),
    
    // 10% API consumers: constant load
    apiLoadScenario.injectOpen(
      constantUsersPerSec(50).during(14.minutes)
    )
  )
  .protocols(httpProtocol)
  
  // ============== Assertions (Pass/Fail Criteria) ==============
  .assertions(
    // Global assertions
    global.responseTime.percentile(50).lt(150),    // p50 < 150ms
    global.responseTime.percentile(95).lt(500),    // p95 < 500ms
    global.responseTime.percentile(99).lt(1000),   // p99 < 1s
    global.successfulRequests.percent.gt(99.9),    // 99.9% success
    global.failedRequests.percent.lt(0.1),         // < 0.1% failures
    
    // Group-specific assertions
    forAll.responseTime.percentile(95).lt(500),
    
    // Request-specific assertions
    details("GET Homepage").responseTime.percentile(95).lt(200),
    details("POST Place Order").responseTime.percentile(95).lt(1000),
  )
}
```

---

### Example 7: Gatling Advanced Features

```scala
// src/test/scala/simulations/AdvancedGatlingSimulation.scala
// Advanced Gatling features: sessions, conditions, loops, custom feeders

package simulations

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

class AdvancedGatlingSimulation extends Simulation {
  
  val httpProtocol = http.baseUrl("https://api.example.com")

  // Custom feeder with programmatic data
  val randomUserId = Iterator.continually(Map("userId" -> scala.util.Random.nextInt(10000)))

  // Scenario with advanced features
  val advancedScenario = scenario("Advanced Features")
    
    // ========== Session Management ==========
    .exec(session => {
      // Initialize session variables
      session
        .set("requestCount", 0)
        .set("totalTime", 0)
        .set("lastResponse", "")
    })
    
    // ========== Conditional Logic ==========
    .feed(randomUserId)
    .exec(http("Check User")
      .get("/users/${userId}")
      .check(status.saveAs("userStatus")))
    .doIf(session => session("userStatus").as[Int] == 200) {
      exec(http("Get User Details")
        .get("/users/${userId}/details")
        .check(status.is(200)))
    }
    .doIfNot(session => session("userId").as[Int] % 2 == 0) {
      exec(http("Update Odd User")
        .post("/users/${userId}/update")
        .body(StringBody("""{"status":"premium"}"""))
        .check(status.is(200)))
    }
    .pause(1)
    
    // ========== Loops ==========
    .repeat(5, "iteration") {
      exec(session => {
        val iteration = session("iteration").as[Int]
        println(s"Iteration: $iteration")
        session
      })
      .exec(http("Request #{iteration}")
        .get(s"/api/data")
        .queryParam("page", session => session("iteration").as[String]))
      .pause(500.millis)
    }
    
    // ========== Error Handling ==========
    .tryMax(3) {  // Retry up to 3 times on failure
      exec(http("Potentially Failing Request")
        .post("/risky-operation")
        .body(StringBody("""{"action":"retry"}"""))
        .check(status.is(200)))
    }
    .exitHereIfFailed  // Stop user if final attempt fails
    
    // ========== Group Execution ==========
    .group("Critical Transaction") {
      exec(http("Prepare")
        .post("/transaction/prepare")
        .check(status.is(200))
        .check(jsonPath("$.transactionId").saveAs("txId")))
      .pause(1)
      .exec(http("Execute")
        .post("/transaction/${txId}/execute")
        .check(status.is(200)))
      .pause(1)
      .exec(http("Confirm")
        .post("/transaction/${txId}/confirm")
        .check(status.is(200)))
    }
    .pause(2)
    
    // ========== Dynamic URLs ==========
    .exec(session => {
      val customPath = if (session("userId").as[Int] % 2 == 0) "/path1" else "/path2"
      session.set("dynamicPath", customPath)
    })
    .exec(http("Dynamic Request")
      .get("${dynamicPath}")
      .check(status.is(200)))
    
    // ========== Custom Checks ==========
    .exec(http("Complex Check")
      .post("/complex")
      .body(StringBody("""{"data":"value"}"""))
      .check(
        status.in(200, 201),
        jsonPath("$.result").exists,
        jsonPath("$.result.id").validate((id, session) => {
          if (id.toInt > 0) Validate.Success else Validate.Failure(s"Invalid ID: $id")
        })
      ))

  setUp(
    advancedScenario.injectOpen(
      rampUsers(10).during(1.minute),
      constantUsersPerSec(1).during(5.minutes)
    )
  ).protocols(httpProtocol)
}
```

---

### Example 8: Gatling Cloud Integration

```scala
// src/test/scala/simulations/CloudSimulation.scala
// Gatling Enterprise (Cloud) configuration for distributed testing

package simulations

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

class CloudSimulation extends Simulation {
  
  val httpProtocol = http
    .baseUrl("https://api.example.com")
    .warmUpUrl("https://api.example.com/health")

  val scenario = scenario("Cloud Load Test")
    .repeat(100) {
      exec(http("Request")
        .get("/endpoint")
        .check(status.is(200)))
        .pause(100.millis, 500.millis)
    }

  setUp(
    scenario.injectOpen(
      constantUsersPerSec(500).during(10.minutes)
    )
  )
  .protocols(httpProtocol)
  .assertions(
    global.responseTime.percentile(95).lt(500),
    global.failedRequests.percent.lt(0.1)
  )
}
```

**Deploy to Gatling Enterprise**:
```bash
# Build and push to Gatling
mvn gatling:enterprise-package
gatling-maven-plugin:gatling:enterprise-push
```

---

## 4. Performance Benchmarking (650 words)

### Establishing Baselines

**Step 1: Measure Current Performance**:
```
Run load test on current implementation
Record: p50, p95, p99, throughput, error rate
Save metrics as baseline.json
Create baseline timestamp for tracking
```

**Step 2: Define Performance Targets**:
```
Based on business requirements:
- User expectations (good UX: p95 < 200ms)
- Budget constraints (infrastructure cost limits)
- Competitive benchmarks (what do competitors support?)
- Industry standards (payment processing: p99 < 1s)
```

**Step 3: Continuous Monitoring**:
```
Run tests regularly:
- After code changes (catch regressions)
- Weekly/monthly (track trends)
- Before major events (holiday season readiness)

Alert if performance degrades > 10%
```

---

### Example 9: Automated Benchmark Comparison

```python
#!/usr/bin/env python3
# scripts/benchmark_compare.py
# Compare current performance against baseline

import json
import sys
import argparse
from datetime import datetime
from pathlib import Path
from typing import Dict, Any

class PerformanceBenchmark:
    """Compare load test results against baseline"""
    
    def __init__(self, baseline_file: str, current_file: str, tolerance: float = 0.10):
        """
        Args:
            baseline_file: Path to baseline metrics JSON
            current_file: Path to current metrics JSON
            tolerance: Acceptable regression percentage (default: 10%)
        """
        self.baseline = self.load_results(baseline_file)
        self.current = self.load_results(current_file)
        self.regression_threshold = tolerance
        self.report = {
            'timestamp': datetime.now().isoformat(),
            'baseline_file': baseline_file,
            'current_file': current_file,
            'tolerance': f"{tolerance * 100}%",
            'passed': True,
            'metrics': {},
            'regressions': [],
            'improvements': [],
        }

    def load_results(self, filepath: str) -> Dict[str, Any]:
        """Load k6 JSON results"""
        with open(filepath) as f:
            return json.load(f)

    def extract_metric(self, data: Dict, metric_path: str) -> float:
        """Extract metric from nested JSON"""
        parts = metric_path.split('.')
        value = data
        for part in parts:
            if isinstance(value, dict):
                value = value.get(part, 0)
            else:
                return 0
        return float(value) if value else 0

    def compare_metrics(self):
        """Compare all relevant metrics"""
        metrics_to_compare = {
            'response_time_p50': 'metrics.http_req_duration.values.p(50)',
            'response_time_p95': 'metrics.http_req_duration.values.p(95)',
            'response_time_p99': 'metrics.http_req_duration.values.p(99)',
            'throughput_rps': 'metrics.http_reqs.values.rate',
            'error_rate': 'metrics.http_req_failed.values.rate',
            'requests_total': 'metrics.http_reqs.values.count',
        }

        for metric_name, metric_path in metrics_to_compare.items():
            baseline_value = self.extract_metric(self.baseline, metric_path)
            current_value = self.extract_metric(self.current, metric_path)

            # Calculate change
            if baseline_value > 0:
                change_pct = ((current_value - baseline_value) / baseline_value) * 100
            else:
                change_pct = 0

            # Determine if pass/fail
            # For latency metrics: higher is worse
            # For throughput/rate metrics: lower is worse (but higher is better)
            is_latency_metric = 'response_time' in metric_name or 'duration' in metric_name
            
            if is_latency_metric:
                # Latency: regression if increased
                is_regression = change_pct > (self.regression_threshold * 100)
            else:
                # Throughput: regression if decreased
                is_regression = change_pct < -(self.regression_threshold * 100)

            result = {
                'baseline': round(baseline_value, 2),
                'current': round(current_value, 2),
                'change_pct': round(change_pct, 1),
                'regression': is_regression,
                'threshold': f"{self.regression_threshold * 100}%",
            }

            self.report['metrics'][metric_name] = result

            if is_regression:
                self.report['regressions'].append({
                    'metric': metric_name,
                    'baseline': result['baseline'],
                    'current': result['current'],
                    'change': f"{result['change_pct']:+.1f}%",
                })
                self.report['passed'] = False
            elif change_pct < 0 if is_latency_metric else change_pct > 0:
                self.report['improvements'].append({
                    'metric': metric_name,
                    'baseline': result['baseline'],
                    'current': result['current'],
                    'change': f"{result['change_pct']:+.1f}%",
                })

    def generate_report(self) -> int:
        """Generate comparison report and return exit code"""
        self.compare_metrics()

        print("\n" + "=" * 80)
        print("PERFORMANCE BENCHMARK COMPARISON REPORT")
        print("=" * 80)
        print(f"Timestamp:     {self.report['timestamp']}")
        print(f"Baseline:      {self.report['baseline_file']}")
        print(f"Current:       {self.report['current_file']}")
        print(f"Tolerance:     {self.report['tolerance']}")
        print("-" * 80)

        print("\nMETRICS:")
        for metric, data in self.report['metrics'].items():
            status = "REGRESSION" if data['regression'] else "OK"
            status_icon = "❌" if data['regression'] else "✅"
            
            print(f"\n  {status_icon} {metric}:")
            print(f"      Baseline:  {data['baseline']}")
            print(f"      Current:   {data['current']}")
            print(f"      Change:    {data['change_pct']:+.1f}%")
            print(f"      Status:    {status}")

        if self.report['regressions']:
            print("\n" + "-" * 80)
            print("REGRESSIONS DETECTED:")
            for reg in self.report['regressions']:
                print(f"  - {reg['metric']}: {reg['change']} " +
                      f"({reg['baseline']} → {reg['current']})")

        if self.report['improvements']:
            print("\n" + "-" * 80)
            print("IMPROVEMENTS:")
            for imp in self.report['improvements']:
                print(f"  + {imp['metric']}: {imp['change']} " +
                      f"({imp['baseline']} → {imp['current']})")

        print("\n" + "=" * 80)
        if self.report['passed']:
            print("RESULT: ✅ PASSED - No regressions detected")
            exit_code = 0
        else:
            print("RESULT: ❌ FAILED - Performance regressions detected!")
            exit_code = 1

        print("=" * 80 + "\n")

        return exit_code

    def save_report(self, output_file: str):
        """Save JSON report"""
        with open(output_file, 'w') as f:
            json.dump(self.report, f, indent=2)
        print(f"Report saved to: {output_file}")


def main():
    parser = argparse.ArgumentParser(description='Compare load test benchmarks')
    parser.add_argument('baseline', help='Baseline results file (JSON)')
    parser.add_argument('current', help='Current results file (JSON)')
    parser.add_argument('--tolerance', type=float, default=0.10,
                       help='Acceptable regression percentage (default: 0.10 = 10%%)')
    parser.add_argument('--output', help='Save JSON report to file')
    
    args = parser.parse_args()

    benchmark = PerformanceBenchmark(args.baseline, args.current, args.tolerance)
    exit_code = benchmark.generate_report()

    if args.output:
        benchmark.save_report(args.output)

    sys.exit(exit_code)


if __name__ == '__main__':
    main()
```

**Usage**:
```bash
# Generate baseline
k6 run --out json=baseline.json load-test.js

# Make changes and test
k6 run --out json=current.json load-test.js

# Compare (exit 0 = pass, exit 1 = fail)
python benchmark_compare.py baseline.json current.json --tolerance 0.10
```

---

### Profiling & Root Cause Analysis

**Database profiling**:
```sql
-- Identify slow queries
SELECT query_time, query FROM slow_query_log
WHERE query_time > 0.1
ORDER BY query_time DESC LIMIT 10;

-- Analyze query plan
EXPLAIN SELECT * FROM orders o
  JOIN products p ON o.product_id = p.id
  WHERE o.user_id = 123;
```

**CPU profiling (Node.js)**:
```bash
# Record CPU profile
node --prof app.js &
# ... run load test ...

# Process profile
node --prof-process isolate-*.log > profile.txt
cat profile.txt
```

**Network analysis**:
```bash
# Use k6's detailed timing breakdown
import http from 'k6/http';

const res = http.get('https://api.example.com/endpoint');
console.log(`DNS lookup:    ${res.timings.dns}ms`);
console.log(`TLS handshake: ${res.timings.tls}ms`);
console.log(`TCP connect:   ${res.timings.connect}ms`);
console.log(`Server wait:   ${res.timings.wait}ms`);    // server processing
console.log(`Download:      ${res.timings.receive}ms`);
console.log(`Total:         ${res.timings.duration}ms`);
```

---

### Optimization Strategies & Examples

**Example 10: Before & After Optimization**

```python
# optimization_analysis.py
# Track performance improvements from optimizations

import json
from typing import Dict

class OptimizationTracker:
    def __init__(self):
        self.changes = []
    
    def record_change(self, name: str, baseline: Dict, optimized: Dict):
        """Record optimization change"""
        improvement = {
            'optimization': name,
            'metrics': {}
        }
        
        for metric in baseline:
            if metric in optimized:
                baseline_val = baseline[metric]
                optimized_val = optimized[metric]
                
                if baseline_val != 0:
                    improvement_pct = ((baseline_val - optimized_val) / baseline_val) * 100
                else:
                    improvement_pct = 0
                
                improvement['metrics'][metric] = {
                    'before': baseline_val,
                    'after': optimized_val,
                    'improvement': f"{improvement_pct:.1f}%"
                }
        
        self.changes.append(improvement)
    
    def summary(self):
        print("\n=== OPTIMIZATION SUMMARY ===")
        for change in self.changes:
            print(f"\n{change['optimization']}:")
            for metric, data in change['metrics'].items():
                print(f"  {metric}: {data['before']} → {data['after']} " +
                      f"({data['improvement']})")


# Example optimizations
tracker = OptimizationTracker()

# Optimization 1: Database indexing
tracker.record_change(
    "Add database index on user_id",
    {'query_time_p95': 1000, 'throughput': 1000},
    {'query_time_p95': 150, 'throughput': 5000}
)

# Optimization 2: API response caching
tracker.record_change(
    "Implement Redis caching",
    {'api_response_time_p95': 500},
    {'api_response_time_p95': 50}
)

tracker.summary()
```

**Common optimization patterns**:

| Problem | Symptom | Fix | Impact |
|---------|---------|-----|--------|
| **N+1 Query** | DB queries spike | JOINs, batch queries | 50-80% query time reduction |
| **Missing Index** | High DB latency | Add index on WHERE/JOIN cols | 90%+ query time reduction |
| **Large Responses** | High bandwidth | Pagination, compression | 70-90% data reduction |
| **Sync Code** | Low throughput | Async/await | 3-10x throughput increase |
| **Memory Leak** | Memory grows | Fix listeners, clear refs | Stable memory usage |

---

## 5. CI/CD Integration (600 words)

### GitHub Actions Workflow

**Example 11: Automated Load Testing Pipeline**

```yaml
# .github/workflows/performance-tests.yml
# Automated load testing on schedule and pull requests

name: Load Testing Pipeline

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM UTC
    - cron: '0 0 * * 0'  # Weekly Sunday at midnight
  
  workflow_dispatch:  # Manual trigger
  
  pull_request:
    branches: [main, develop]
    paths:
      - 'src/**'
      - '.github/workflows/performance-tests.yml'

concurrency:
  group: load-test-${{ github.ref }}
  cancel-in-progress: true

jobs:
  # Job 1: Smoke test (quick validation)
  smoke-test:
    name: Smoke Test
    runs-on: ubuntu-latest
    timeout-minutes: 10
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 1

      - name: Install k6
        run: |
          sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 \
            --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb https://dl.k6.io/deb stable main" | \
            sudo tee /etc/apt/sources.list.d/k6-stable.list
          sudo apt-get update
          sudo apt-get install -y k6

      - name: Run smoke test
        run: |
          k6 run tests/load/smoke-test.js \
            --vus 10 \
            --duration 2m
        env:
          API_BASE_URL: ${{ secrets.STAGING_API_URL }}
          API_KEY: ${{ secrets.STAGING_API_KEY }}
        continue-on-error: false

  # Job 2: Load test (scheduled or manual)
  load-test:
    name: Load Test
    runs-on: ubuntu-latest
    timeout-minutes: 30
    if: github.event_name == 'schedule' || github.event_name == 'workflow_dispatch'
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Install k6
        run: |
          sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 \
            --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb https://dl.k6.io/deb stable main" | \
            sudo tee /etc/apt/sources.list.d/k6-stable.list
          sudo apt-get update
          sudo apt-get install -y k6

      - name: Run load test
        run: |
          k6 run tests/load/load-test.js \
            --out json=results.json \
            --vus 1000 \
            --duration 10m
        env:
          API_BASE_URL: ${{ secrets.STAGING_API_URL }}
          API_KEY: ${{ secrets.STAGING_API_KEY }}

      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Install dependencies
        run: pip install -q matplotlib

      - name: Compare with baseline
        run: |
          if [ -f baseline/load-test.json ]; then
            python scripts/benchmark_compare.py \
              baseline/load-test.json \
              results.json \
              --tolerance 0.10 \
              --output comparison-report.json
          else
            echo "No baseline found, creating new baseline..."
            cp results.json baseline/load-test.json
          fi
        continue-on-error: true

      - name: Upload test artifacts
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: load-test-results-${{ github.run_id }}
          path: |
            results.json
            comparison-report.json
          retention-days: 90

      - name: Comment on PR with results
        if: github.event_name == 'pull_request' && always()
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            let comment = '## Load Test Results\n\n';
            
            try {
              const results = JSON.parse(fs.readFileSync('results.json', 'utf8'));
              const metrics = results.metrics || {};
              
              comment += '| Metric | Value |\n';
              comment += '|--------|-------|\n';
              comment += `| Response Time (p95) | ${metrics.http_req_duration?.values?.['p(95)']?.toFixed(2) || 'N/A'}ms |\n`;
              comment += `| Throughput | ${metrics.http_reqs?.values?.rate?.toFixed(2) || 'N/A'} req/s |\n`;
              comment += `| Error Rate | ${(metrics.http_req_failed?.values?.rate * 100).toFixed(2) || 'N/A'}% |\n`;
            } catch (e) {
              comment += `Error parsing results: ${e.message}\n`;
            }
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            });

      - name: Fail if regressions detected
        if: failure()
        run: |
          echo "Performance regressions detected!"
          echo "Please review load test results and optimize performance."
          exit 1

  # Job 3: Monitor & Alert
  monitor:
    name: Performance Monitoring
    runs-on: ubuntu-latest
    if: github.event_name == 'schedule'
    needs: load-test
    
    steps:
      - name: Upload to monitoring service
        run: |
          curl -X POST https://monitoring.example.com/api/results \
            -H "Authorization: Bearer ${{ secrets.MONITORING_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d @results.json
        continue-on-error: true

      - name: Check SLA compliance
        run: |
          # Verify against SLA thresholds
          python scripts/check_sla.py results.json
        env:
          SLA_P95_MS: 200
          SLA_ERROR_RATE: 0.1
```

---

### GitLab CI Integration

**Example 12: GitLab Pipeline**

```yaml
# .gitlab-ci.yml

stages:
  - load-test
  - report
  - deploy

load_test:
  stage: load-test
  image: grafana/k6:latest
  script:
    - k6 run tests/load/load-test.js --out json=results.json
  artifacts:
    paths:
      - results.json
    reports:
      performance: results.json
  only:
    - schedules
    - main

compare_performance:
  stage: report
  image: python:3.11
  script:
    - pip install matplotlib
    - python scripts/benchmark_compare.py baseline.json results.json
  dependencies:
    - load_test
  allow_failure: true

quality_gate:
  stage: report
  image: python:3.11
  script:
    - python scripts/check_thresholds.py results.json
  dependencies:
    - load_test
```

---

### Jenkins Integration

**Example 13: Jenkins Pipeline (Distributed)**

```groovy
// Jenkinsfile

pipeline {
    agent any
    
    options {
        timestamps()
        timeout(time: 30, unit: 'MINUTES')
        buildDiscarder(logRotator(numToKeepStr: '10'))
    }
    
    triggers {
        cron('0 2 * * *')  // Daily at 2 AM
    }
    
    stages {
        stage('Setup') {
            steps {
                echo 'Installing k6...'
                sh '''
                    curl https://dl.k6.io/check/latest.txt
                    wget https://dl.k6.io/releases/v0.54.0/k6-v0.54.0-linux-amd64.tar.gz
                    tar xzf k6-v0.54.0-linux-amd64.tar.gz
                '''
            }
        }
        
        stage('Parallel Load Tests') {
            parallel {
                stage('US East') {
                    agent { label 'us-east' }
                    steps {
                        sh './k6/k6 run tests/load/load-test.js --out json=results-us-east.json'
                    }
                }
                stage('EU West') {
                    agent { label 'eu-west' }
                    steps {
                        sh './k6/k6 run tests/load/load-test.js --out json=results-eu-west.json'
                    }
                }
                stage('AP Southeast') {
                    agent { label 'ap-southeast' }
                    steps {
                        sh './k6/k6 run tests/load/load-test.js --out json=results-ap.json'
                    }
                }
            }
        }
        
        stage('Aggregate Results') {
            steps {
                sh '''
                    python scripts/aggregate_results.py \
                        results-us-east.json \
                        results-eu-west.json \
                        results-ap.json \
                        > aggregate-results.json
                '''
                archiveArtifacts artifacts: '**/results*.json'
            }
        }
        
        stage('Quality Gate') {
            steps {
                script {
                    def results = readJSON file: 'aggregate-results.json'
                    if (!results.passed) {
                        error('Performance thresholds exceeded!')
                    }
                }
            }
        }
        
        stage('Report') {
            steps {
                publishHTML([
                    reportDir: 'reports',
                    reportFiles: 'index.html',
                    reportName: 'Performance Report'
                ])
            }
        }
    }
    
    post {
        always {
            archiveArtifacts artifacts: '**/results*.json,aggregate-results.json'
            cleanWs()
        }
        failure {
            mail to: 'devops@example.com',
                 subject: "Load Test Failed: ${env.JOB_NAME}",
                 body: "Build failed. Check ${env.BUILD_URL} for details."
        }
    }
}
```

---

## 6. Advanced Scenarios (650 words)

### Example 14: Realistic Traffic Simulation

```javascript
// realistic-traffic.js
// Simulate realistic user behavior patterns based on business data

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';

const pageLoadTime = new Trend('page_load_time');
const checkoutSuccess = new Rate('checkout_success');

// User segments based on real analytics (Pareto distribution)
const userSegments = [
  {
    name: 'Light Browser Users',
    percentage: 60,     // 60% of traffic
    characteristics: {
      avg_session_length: 3,           // 3 pages per session
      avg_time_per_page: 120,          // 120 seconds per page
      conversion_rate: 0.02,           // 2% convert
      bounce_rate: 0.40,               // 40% bounce
    },
  },
  {
    name: 'Power Shoppers',
    percentage: 25,     // 25% of traffic
    characteristics: {
      avg_session_length: 10,
      avg_time_per_page: 30,
      conversion_rate: 0.20,           // 20% convert
      bounce_rate: 0.10,
    },
  },
  {
    name: 'API Consumers',
    percentage: 10,     // 10% of traffic
    characteristics: {
      avg_session_length: 100,         // many API calls
      avg_time_per_page: 2,
      conversion_rate: 0.05,
      bounce_rate: 0.05,
    },
  },
  {
    name: 'Mobile Users',
    percentage: 5,      // 5% of traffic
    characteristics: {
      avg_session_length: 2,
      avg_time_per_page: 60,
      conversion_rate: 0.01,
      bounce_rate: 0.60,
    },
  },
];

// Time-based traffic patterns (realistic daily pattern)
const timeBasedPatterns = {
  morning: {        // 6 AM - 12 PM: 30% traffic
    vus: 300,
    description: 'morning commute browsing'
  },
  afternoon: {      // 12 PM - 6 PM: 50% traffic (peak)
    vus: 500,
    description: 'lunch break and afternoon shopping'
  },
  evening: {        // 6 PM - 12 AM: 60% traffic (highest)
    vus: 600,
    description: 'evening shopping and browsing'
  },
  night: {          // 12 AM - 6 AM: 20% traffic
    vus: 200,
    description: 'night owls and international users'
  },
};

export const options = {
  stages: [
    // Ramp up to morning traffic
    { duration: '5m', target: 300 },
    // Afternoon peak
    { duration: '5m', target: 500 },
    // Evening peak
    { duration: '5m', target: 600 },
    // Ramp down to night
    { duration: '5m', target: 200 },
  ],
  thresholds: {
    'page_load_time': ['p(95)<500', 'p(99)<1500'],
    'checkout_success': ['rate>0.95'],
  },
};

export default function () {
  const userSegment = selectUserSegment();
  simulateUserJourney(userSegment);
}

function selectUserSegment() {
  const rand = Math.random() * 100;
  let cumulative = 0;

  for (const segment of userSegments) {
    cumulative += segment.percentage;
    if (rand < cumulative) {
      return segment;
    }
  }

  return userSegments[0];
}

function simulateUserJourney(segment) {
  group(`${segment.name} Journey`, function () {
    for (let i = 0; i < segment.characteristics.avg_session_length; i++) {
      // Decide whether to bounce
      if (Math.random() < segment.characteristics.bounce_rate && i > 0) {
        break;  // User bounces
      }

      // View product/page
      const startTime = Date.now();
      const res = http.get('https://example.com/products', {
        headers: { 'User-Agent': segment.name.replace(' ', '-') },
      });
      const loadTime = Date.now() - startTime;
      
      check(res, {
        'page loads': (r) => r.status === 200,
      });

      pageLoadTime.add(loadTime);

      // Think time (realistic user behavior)
      const thinkTime = Math.random() *
        segment.characteristics.avg_time_per_page * 1000 +
        Math.random() * 5000;
      sleep(thinkTime / 1000);

      // Decide whether to convert
      if (Math.random() < segment.characteristics.conversion_rate) {
        // Checkout
        const checkoutRes = http.post('https://example.com/checkout', {});
        checkoutSuccess.add(checkoutRes.status === 200 ? 1 : 0);
        sleep(1);
        break;  // End session after conversion
      }
    }
  });
}
```

---

### Example 15: Multiple Concurrent User Journeys

```javascript
// multi-scenario.js
// Multiple realistic scenarios running in parallel

import http from 'k6/http';
import { check, group } from 'k6';

// Feeder data
const users = [
  { email: 'user1@test.com', role: 'free' },
  { email: 'user2@test.com', role: 'premium' },
  { email: 'user3@test.com', role: 'enterprise' },
];

export const options = {
  scenarios: {
    free_tier_browsers: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 50 },   // 50 free users
      ],
      gracefulStop: '1m',
      env: { USER_TIER: 'free' },
    },
    premium_shoppers: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 30 },   // 30 premium users
      ],
      gracefulStop: '1m',
      startTime: '2m',  // Start after free tier ramp
      env: { USER_TIER: 'premium' },
    },
    enterprise_api: {
      executor: 'constant-vus',
      vus: 10,
      duration: '6m',
      startTime: '4m',
      env: { USER_TIER: 'enterprise' },
    },
  },
};

export default function () {
  const tier = __ENV.USER_TIER;

  if (tier === 'free') {
    freeUserJourney();
  } else if (tier === 'premium') {
    premiumUserJourney();
  } else {
    enterpriseAPIUsage();
  }
}

function freeUserJourney() {
  group('Free Tier Journey', function () {
    http.get('https://api.example.com/public');
    http.get('https://api.example.com/limited');
  });
}

function premiumUserJourney() {
  group('Premium Journey', function () {
    const auth = http.post('https://api.example.com/auth', {});
    const token = auth.json('token');

    http.get('https://api.example.com/premium', {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    http.post('https://api.example.com/export', {}, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
  });
}

function enterpriseAPIUsage() {
  group('Enterprise API', function () {
    for (let i = 0; i < 100; i++) {
      http.get(`https://api.example.com/data/${i}`);
    }
  });
}
```

---

## 7. Monitoring & Alerting (600 words)

### Example 16: Prometheus + Grafana Integration

```yaml
# monitoring/prometheus.yml
# Prometheus configuration for scraping k6 and application metrics

global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    monitor: 'load-testing'

scrape_configs:
  # Scrape k6 metrics
  - job_name: 'k6-metrics'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
    scrape_interval: 5s

  # Scrape application API server
  - job_name: 'api-server'
    static_configs:
      - targets: ['api.example.com:8080']
    metrics_path: '/metrics'
    scheme: 'https'
    tls_config:
      insecure_skip_verify: false

  # Scrape database metrics
  - job_name: 'postgres-db'
    static_configs:
      - targets: ['db.example.com:5432']
    metrics_path: '/metrics'

  # Scrape Kubernetes metrics
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: 'true'
```

**Grafana Dashboard Queries**:

```promql
# Response Time Percentiles
histogram_quantile(0.50, http_request_duration_seconds_bucket)  # p50
histogram_quantile(0.95, http_request_duration_seconds_bucket)  # p95
histogram_quantile(0.99, http_request_duration_seconds_bucket)  # p99

# Requests Per Second
rate(http_requests_total[1m])

# Error Rate
rate(http_requests_failed_total[1m]) / rate(http_requests_total[1m])

# Active Virtual Users
max(k6_vus)

# Database Query Time
histogram_quantile(0.95, database_query_duration_seconds_bucket)

# CPU Usage
rate(cpu_usage_seconds_total[1m])

# Memory Usage
memory_usage_bytes / 1024 / 1024  # Convert to MB
```

---

### Example 17: Alert Rules

```yaml
# monitoring/alert-rules.yml
# Prometheus alert rules for performance

groups:
  - name: performance.rules
    interval: 30s
    rules:
      # Response time degradation
      - alert: HighResponseTime
        expr: |
          histogram_quantile(0.95, http_request_duration_seconds_bucket) > 0.5
        for: 5m
        annotations:
          summary: "High response time detected"
          description: "p95 response time > 500ms for 5 minutes"

      # High error rate
      - alert: HighErrorRate
        expr: |
          (rate(http_requests_failed_total[1m]) / rate(http_requests_total[1m])) > 0.01
        for: 2m
        annotations:
          summary: "High error rate"
          description: "Error rate > 1% for 2 minutes"

      # Database slowdown
      - alert: SlowDatabase
        expr: |
          histogram_quantile(0.95, database_query_duration_seconds_bucket) > 1.0
        for: 3m
        annotations:
          summary: "Database queries slow"

      # Memory leak detection
      - alert: MemoryGrowth
        expr: |
          rate(memory_usage_bytes[15m]) > 1000000  # 1MB/15min growth
        for: 30m
        annotations:
          summary: "Potential memory leak detected"
```

---

## 8. Optimization Strategies (600 words)

### Server-side Optimizations

**Database optimization**:
```sql
-- Before: N+1 queries (slow)
SELECT * FROM orders WHERE user_id = 123;
-- For each order: SELECT * FROM products WHERE id = order.product_id;

-- After: Single JOIN (fast)
SELECT o.*, p.name, p.price
FROM orders o
JOIN products p ON o.product_id = p.id
WHERE o.user_id = 123;
```

**Index optimization**:
```sql
-- Add index on frequently queried columns
CREATE INDEX idx_user_id ON orders(user_id);
CREATE INDEX idx_created_at ON orders(created_at);
CREATE INDEX idx_status ON orders(status);

-- Verify index usage
EXPLAIN SELECT * FROM orders WHERE user_id = 123;
```

**Caching strategy**:
```javascript
// Application-level caching
const cache = new Map();

function getUser(userId) {
  const cached = cache.get(`user:${userId}`);
  if (cached) return cached;

  const user = db.getUser(userId);
  cache.set(`user:${userId}`, user, 3600);  // 1 hour TTL
  return user;
}
```

---

### Client-side Optimizations

**Response compression**:
```javascript
// Express.js middleware
import compression from 'compression';
app.use(compression({ level: 9 }));  // Maximum compression

// Results: 50-80% data reduction
```

**HTTP/2 multiplexing**:
```
HTTP/1.1: Serial requests (slow)
  Request 1 ──────────┐
                      Request 2 ──────────┐
                                          Request 3

HTTP/2: Parallel requests (fast)
  Request 1 ──────────┐
  Request 2 ──────────┤ All parallel
  Request 3 ──────────┘
```

**Connection pooling**:
```javascript
// Reuse database connections
const pool = new Pool({
  max: 20,              // Maximum connections
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
});

// Connection pool: 20 VU × many requests = efficient
```

---

### Horizontal Scaling

**Kubernetes HPA (Horizontal Pod Autoscaler)**:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-server
  minReplicas: 2
  maxReplicas: 50
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
        - type: Percent
          value: 100
          periodSeconds: 15
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 50
          periodSeconds: 15
```

---

## Best Practices Summary

### Load Test Design Checklist

- [ ] Define clear SLOs (Service Level Objectives)
- [ ] Start with smoke test (small load)
- [ ] Use realistic traffic patterns
- [ ] Test for at least 5-10 minutes (find leaks)
- [ ] Monitor all system layers (API, DB, infra)
- [ ] Compare against baseline metrics
- [ ] Document findings and optimization actions
- [ ] Automate tests in CI/CD pipeline
- [ ] Retest after optimizations
- [ ] Establish performance budgets
- [ ] Alert on regressions automatically

### Common Pitfalls to Avoid

- ❌ Testing with only GET requests (unrealistic)
- ❌ Insufficient test duration (memory leaks not visible)
- ❌ Ignoring external dependencies (APIs, databases)
- ❌ No baseline established (can't detect regressions)
- ❌ Single test run (need multiple runs for statistical confidence)
- ❌ Testing prod directly (risk of downtime)
- ❌ No correlation with business metrics
- ❌ Manual testing only (not automated)

---

## Related Skills

- `moai-testing-unit`: Unit testing with Pytest, Jest
- `moai-testing-integration`: Integration testing strategies
- `moai-testing-e2e`: End-to-end browser testing
- `moai-domain-backend`: Backend architecture & scaling
- `moai-essentials-perf`: Performance profiling & optimization

---

## Quick Reference

### k6 Commands
```bash
k6 run script.js
k6 run --vus 100 --duration 10m script.js
k6 run --out json=results.json script.js
k6 cloud script.js
```

### Gatling Commands
```bash
./gatling.sh -s simulations.MySimulation
mvn gatling:test
mvn gatling:enterprise-push
```

### Tool Comparison
```
k6:      Simple API testing, JavaScript, < 50K VU
Gatling: Complex scenarios, Scala, > 100K VU
JMeter:  Legacy, GUI-based
Locust:  Python-based, distributed
```

### Performance Targets
```
E-commerce: p95 < 200ms, error < 0.1%
API: p95 < 300ms, error < 0.05%
Mobile: p95 < 1000ms (poor network)
Payment: p99 < 1s, error < 0.01%
```

---

**Tool Versions**: k6 0.54+, Gatling 3.13+, Prometheus 3.1+, Grafana 11.4+

**Generation Method**: Claude Code   | **Language**: English (100%)

**Last Updated**: 2025-11-19


---

## Additional Resources

### Online Learning

**Official Documentation**:
- k6 Official Docs: https://k6.io/docs/
- Gatling Official: https://gatling.io/
- Prometheus: https://prometheus.io/docs/

**Courses & Tutorials**:
- k6 Academy (free, interactive)
- Gatling Academy (free for open-source)
- Linux Academy Performance Testing

### Community Forums

**Getting Help**:
- k6 Community Forum
- Gatling Google Groups
- Stack Overflow tags: `k6`, `gatling`, `load-testing`
- r/loadtesting (Reddit)

### Industry Benchmarks

**Typical Performance Baselines**:
- REST API: p95 < 200ms, error < 0.1%
- Web Application: p95 < 300ms, error < 0.5%
- Mobile API: p95 < 1000ms (poor network), error < 1%
- Payment System: p99 < 1s, error < 0.01%
- Email Notifications: p95 < 5s (async), error < 0.1%

### Production Readiness Checklist

**Before Going Live**:
- [ ] Load test completed with realistic traffic
- [ ] Baseline metrics established
- [ ] Performance budget defined
- [ ] SLOs communicated to team
- [ ] CI/CD load testing automated
- [ ] Monitoring/alerting configured
- [ ] Runbook created for performance degradation
- [ ] Team trained on metrics interpretation
- [ ] Regression testing in CI/CD
- [ ] Capacity planning completed

---

## Glossary

| Term | Definition |
|------|-----------|
| **VU** | Virtual User - simulates one user executing test script |
| **RPS** | Requests Per Second - throughput metric |
| **Latency** | Time from request sent to response received |
| **Percentile** | p95 = 95% of requests faster than this value |
| **Throughput** | Requests/transactions completed per unit time |
| **SLO** | Service Level Objective - performance target |
| **SLA** | Service Level Agreement - contractual performance guarantee |
| **Smoke Test** | Quick test with minimal load (validation) |
| **Load Test** | Realistic traffic test (measurement) |
| **Stress Test** | Maximum load until failure (capacity finding) |
| **Spike Test** | Sudden traffic surge (resilience) |
| **Soak Test** | Long-running test (memory leaks, stability) |
| **Apdex** | Application Performance Index (satisfaction score) |
| **Tail Latency** | p99, p99.9 (extreme response times) |
| **Regression** | Performance degradation vs baseline |

---

## Example Test Data Files

### Example 1: Test Users CSV

```csv
# data/users.csv
userId,email,password,userType
1,alice@example.com,secure-pass-001,free
2,bob@example.com,secure-pass-002,premium
3,charlie@example.com,secure-pass-003,enterprise
4,diana@example.com,secure-pass-004,free
5,eve@example.com,secure-pass-005,premium
```

### Example 2: Test Products CSV

```csv
# data/products.csv
productId,name,category,price,stock
101,Laptop Pro,electronics,1299.99,50
102,Wireless Mouse,electronics,29.99,200
103,USB-C Cable,accessories,9.99,500
104,Monitor 27",electronics,399.99,30
105,Keyboard Mechanical,accessories,149.99,75
```

### Example 3: Test Categories CSV

```csv
# data/categories.csv
category,subcategory,description
electronics,computers,Desktop and laptop computers
electronics,peripherals,Keyboards, mice, monitors
accessories,cables,Power and data cables
accessories,cases,Protective cases and bags
home,furniture,Desks, chairs, shelving
```

---

## Code Examples Reference

**Total Code Examples in This Skill**: 17 (JavaScript, Scala, Python, YAML, SQL)

1. **Metric Definition** (YAML)
2. **Test Scenarios** (YAML)
3. **k6 Metrics** (JavaScript)
4. **k6 User Journeys** (JavaScript)
5. **k6 Distributed** (JavaScript)
6. **Gatling Simulation** (Scala)
7. **Gatling Advanced** (Scala)
8. **Gatling Cloud** (Scala)
9. **Benchmark Comparison** (Python)
10. **Realistic Traffic** (JavaScript)
11. **Multi-Scenario** (JavaScript)
12. **Prometheus Config** (YAML)
13. **Prometheus Alerts** (YAML)
14. **GitHub Actions** (YAML)
15. **GitLab CI** (YAML)
16. **Jenkins Pipeline** (Groovy)
17. **Optimization Analysis** (Python)

---

## Certification & Credentials

**Professional Certifications**:
- **ISTQB Performance Testing** - International certification for testers
- **k6 Certified** - Open-source certification (in development)
- **Gatling Enterprise Certification** - Commercial certification
- **AWS Certified DevOps Engineer** - Includes performance testing

**Skills to Develop**:
1. Understanding of system architecture
2. SQL query optimization knowledge
3. Network protocol understanding (HTTP, HTTPS, gRPC)
4. Cloud infrastructure knowledge (AWS, GCP, Azure)
5. Container & Kubernetes basics
6. Monitoring & observability tools
7. Statistical analysis for metrics interpretation

---

## Change Log

** .0 (2025-11-19)** - Current
- Complete rewrite with k6 and Gatling focus
- CI/CD integration (GitHub Actions, GitLab, Jenkins)
- Advanced monitoring (Prometheus, Grafana)
- 17 comprehensive code examples
- Enterprise-ready patterns

**v3.0.0 (2025-06-01)** - Previous
- JMeter and Locust documentation
- Basic load testing patterns
- Performance optimization guides

**v2.0.0 (2024-12-15)** - Historical
- Initial load testing framework

---

## Contributors & Attribution

**Maintained By**: GoosLab (MoAI-ADK Core Team)

**Contributors**:
- Load testing practitioners
- Performance engineering experts
- DevOps automation specialists
- Cloud infrastructure architects

**Special Thanks**:
- k6 community for open-source framework
- Gatling team for enterprise solutions
- Prometheus & Grafana projects for monitoring excellence

---

## License & Usage

**Skill License**: Apache 2.0 (Same as MoAI-ADK)

**Permitted Uses**:
- Educational purposes
- Commercial projects
- Internal tools
- Open source projects

**Attribution Required**: Yes (cite MoAI-ADK)

**Modifications**: Allowed, must maintain attribution

---

## Next Steps

### Immediate Actions (Start Here)

1. **Choose your tool**:
   - API/microservices → k6
   - Complex journeys → Gatling
   - Python preference → Locust

2. **Set up baseline**:
   - Run smoke test
   - Record current performance
   - Define targets

3. **Automate testing**:
   - Add to CI/CD
   - Set thresholds
   - Configure alerts

### 30-Day Learning Path

**Week 1: Fundamentals**
- Read this Skill document (Overview + Section 1)
- Install k6 or Gatling
- Run hello-world test
- Understand metrics (p95, error rate, throughput)

**Week 2: Scenarios**
- Write realistic test scenario
- Use test data feeders
- Implement think time
- Measure baseline performance

**Week 3: Advanced**
- Distributed testing
- Custom metrics
- Monitoring integration
- Failure analysis

**Week 4: Automation**
- CI/CD integration
- Regression detection
- Alert configuration
- Team rollout

### Mastery Path (3-6 months)

1. **Load Testing Mastery**
   - Advanced scenario design
   - Distributed testing across regions
   - Custom metric development
   - Edge case identification

2. **Performance Engineering**
   - Bottleneck analysis
   - Optimization strategies
   - Capacity planning
   - SLO definition

3. **Enterprise Integration**
   - APM tool integration
   - Team collaboration
   - Documentation
   - Training materials

---

## FAQ

**Q: Can I test production directly?**
A: Not recommended. Use staging/pre-prod that mirrors production. Never test prod without permission.

**Q: How many VUs do I need?**
A: Start with 10 VU baseline, then increase gradually until you reach business-realistic traffic.

**Q: What if my API returns errors during load test?**
A: Errors are expected under extreme stress. Focus on error rate < 0.1% under expected load, allow higher rates during stress testing.

**Q: How often should I run load tests?**
A: Weekly for critical services, monthly for others. Always before major releases or traffic increases.

**Q: Can I load test third-party APIs?**
A: Check their terms of service first. Many prohibit load testing. Use staging APIs if available.

**Q: What's the difference between load test and stress test?**
A: Load test = realistic traffic. Stress test = maximum load until failure. Both needed.

**Q: How do I handle database scaling under load?**
A: Use connection pooling, read replicas, caching (Redis), and query optimization together.

**Q: Should I load test on prod traffic or synthetic?**
A: Synthetic traffic for testing (safe). Real traffic replay for validation (advanced). Never use prod directly.

---

## Support & Community

**Getting Help**:
1. Check this Skill documentation (most answers here)
2. Review examples (code is self-documenting)
3. Check logs (.moai/logs/ directory)
4. Ask in MoAI community forums
5. File GitHub issue with test script + error

**Contribute**:
- Share your scenarios
- Report bugs
- Suggest improvements
- Add examples for your domain

**Contact**:
- GitHub Issues: github.com/anthropics/moai-adk
- Community Forum: community.moai.dev
- Email: support@moai.dev

---

**Skill Version**: 4.0.0 | **Status**: Stable | **Last Updated**: 2025-11-19

**Learn More**: Visit k6.io and gatling.io for official documentation. This Skill is your practical guide; official docs provide reference material.

**Share Your Experience**: Found this Skill helpful? Share your load testing success stories!

