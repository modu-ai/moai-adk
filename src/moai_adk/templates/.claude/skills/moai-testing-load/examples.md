# Load Testing Examples

Comprehensive examples for k6 and Gatling load testing scenarios.

## Table of Contents

1. [k6 Complete Examples](#k6-complete-examples)
2. [Gatling Complete Examples](#gatling-complete-examples)
3. [Authentication Flows](#authentication-flows)
4. [Multi-Step Scenarios](#multi-step-scenarios)
5. [Data-Driven Testing](#data-driven-testing)
6. [Performance Metrics](#performance-metrics)

---

## k6 Complete Examples

### Example 1: Basic Load Test with Custom Metrics

```javascript
// load-test-basic.js
import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const apiDur

ation = new Trend('api_duration_ms');
const successfulRequests = new Counter('requests_success');
const activeConnections = new Gauge('active_connections');

const BASE_URL = __ENV.BASE_URL || 'https://httpbin.org';

export const options = {
  vus: 100,
  duration: '5m',

  thresholds: {
    errors: ['rate<0.1'],
    'http_req_duration': ['p(95)<200', 'p(99)<500'],
    'api_duration_ms': ['p(95)<300'],
    requests_success: ['count>4000'],
    checks: ['rate>0.95'],
  },
};

export default function () {
  group('API Test', function () {
    const response = http.get(`${BASE_URL}/get`);

    const success = check(response, {
      'status is 200': (r) => r.status === 200,
      'response time < 500ms': (r) => r.timings.duration < 500,
    });

    if (success) {
      successfulRequests.add(1);
    }

    errorRate.add(!success);
    apiDuration.add(response.timings.duration);
    activeConnections.add(1);

    sleep(1);
  });
}
```

### Example 2: Multi-Stage Load Test

```javascript
// load-test-stages.js
import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    // Ramp-up
    { duration: "2m", target: 50 },
    { duration: "2m", target: 100 },

    // Sustained load
    { duration: "10m", target: 100 },

    // Peak load
    { duration: "2m", target: 200 },
    { duration: "5m", target: 200 },

    // Ramp-down
    { duration: "2m", target: 50 },
    { duration: "1m", target: 0 },
  ],

  thresholds: {
    http_req_duration: ["p(95)<200"],
    http_req_failed: ["rate<0.01"],
  },
};

export default function () {
  const res = http.get("https://api.example.com/products");

  check(res, {
    "status 200": (r) => r.status === 200,
    "has products": (r) => r.json("products").length > 0,
  });

  sleep(Math.random() * 3 + 1); // Random think time 1-4s
}
```

### Example 3: Stress Test (Find Breaking Point)

```javascript
// stress-test.js
export const options = {
  scenarios: {
    stress: {
      executor: "ramping-arrival-rate",
      startRate: 50,
      timeUnit: "1s",
      preAllocatedVUs: 100,
      maxVUs: 1000,
      stages: [
        { duration: "2m", target: 100 }, // Start
        { duration: "5m", target: 100 }, // Baseline
        { duration: "2m", target: 200 }, // Increase
        { duration: "5m", target: 200 }, // Hold
        { duration: "2m", target: 500 }, // Stress
        { duration: "5m", target: 500 }, // Hold
        { duration: "2m", target: 1000 }, // Breaking
        { duration: "5m", target: 1000 }, // Hold
        { duration: "5m", target: 0 }, // Recovery
      ],
    },
  },
};

export default function () {
  const res = http.get("https://api.example.com/heavy-endpoint");

  check(res, {
    "status is 2xx or 429": (r) => r.status < 300 || r.status === 429,
    "response time acceptable": (r) => r.timings.duration < 5000,
  });
}
```

---

## Gatling Complete Examples

### Example 4: Gatling Basic Simulation

```scala
// BasicSimulation.scala
import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

class BasicSimulation extends Simulation {

  val httpProtocol = http
    .baseUrl("https://api.example.com")
    .acceptHeader("application/json")
    .userAgentHeader("Gatling/3.0")

  val scn = scenario("Basic Load Test")
    .exec(
      http("Get Products")
        .get("/products")
        .check(status.is(200))
        .check(jsonPath("$.products").exists)
    )
    .pause(1.second, 3.seconds)

  setUp(
    scn.inject(
      rampUsers(100).during(5.minutes)
    )
  ).protocols(httpProtocol)
    .assertions(
      global.responseTime.percentile3.lt(200),
      global.successfulRequests.percent.gt(99)
    )
}
```

### Example 5: Gatling Multi-Scenario

```scala
// MultiScenarioSimulation.scala
class MultiScenarioSimulation extends Simulation {

  val httpProtocol = http.baseUrl("https://api.example.com")

  val browsing = scenario("Browsing")
    .exec(http("Homepage").get("/"))
    .pause(2)
    .exec(http("Products").get("/products"))
    .pause(3)

  val purchasing = scenario("Purchase")
    .exec(http("Add to Cart").post("/cart"))
    .pause(1)
    .exec(http("Checkout").post("/checkout"))

  setUp(
    browsing.inject(rampUsers(80).during(10.minutes)),
    purchasing.inject(rampUsers(20).during(10.minutes))
  ).protocols(httpProtocol)
}
```

---

## Authentication Flows

### Example 6: OAuth2 Authentication

```javascript
// oauth-load-test.js
import http from "k6/http";
import { check } from "k6";

export function setup() {
  // Get access token once
  const loginRes = http.post("https://auth.example.com/oauth/token", {
    grant_type: "client_credentials",
    client_id: __ENV.CLIENT_ID,
    client_secret: __ENV.CLIENT_SECRET,
  });

  return { token: loginRes.json("access_token") };
}

export default function (data) {
  const params = {
    headers: {
      Authorization: `Bearer ${data.token}`,
      "Content-Type": "application/json",
    },
  };

  const res = http.get("https://api.example.com/protected", params);

  check(res, {
    authenticated: (r) => r.status === 200,
  });
}
```

### Example 7: Session-Based Auth

```javascript
// session-auth.js
export default function () {
  // Login
  const loginRes = http.post("https://api.example.com/login", {
    username: `user${__VU}@example.com`,
    password: "password123",
  });

  const sessionId = loginRes.cookies.sessionId[0].value;

  // Use session
  const params = {
    cookies: { sessionId: sessionId },
  };

  http.get("https://api.example.com/dashboard", params);
  http.get("https://api.example.com/profile", params);

  // Logout
  http.post("https://api.example.com/logout", null, params);
}
```

---

## Multi-Step Scenarios

### Example 8: E-Commerce User Journey

```javascript
// ecommerce-journey.js
import http from "k6/http";
import { check, group, sleep } from "k6";

export default function () {
  group("Browse Products", function () {
    const catalogRes = http.get("https://shop.example.com/api/products");
    check(catalogRes, { "products loaded": (r) => r.status === 200 });
    sleep(2);

    const productId = catalogRes.json("products[0].id");
    const detailRes = http.get(
      `https://shop.example.com/api/products/${productId}`
    );
    check(detailRes, { "product details": (r) => r.status === 200 });
    sleep(3);
  });

  group("Add to Cart", function () {
    const cartRes = http.post(
      "https://shop.example.com/api/cart",
      JSON.stringify({
        productId: 123,
        quantity: 1,
      }),
      {
        headers: { "Content-Type": "application/json" },
      }
    );
    check(cartRes, { "added to cart": (r) => r.status === 201 });
    sleep(1);
  });

  group("Checkout", function () {
    const checkoutRes = http.post(
      "https://shop.example.com/api/checkout",
      JSON.stringify({
        paymentMethod: "credit_card",
        shippingAddress: "123 Main St",
      }),
      {
        headers: { "Content-Type": "application/json" },
      }
    );
    check(checkoutRes, { "checkout success": (r) => r.status === 200 });
  });
}
```

---

## Data-Driven Testing

### Example 9: CSV Data Source

```javascript
// data-driven-test.js
import http from "k6/http";
import { SharedArray } from "k6/data";
import papaparse from "https://jslib.k6.io/papaparse/5.1.1/index.js";

const csvData = new SharedArray("users", function () {
  return papaparse.parse(open("./users.csv"), { header: true }).data;
});

export default function () {
  const user = csvData[__VU % csvData.length];

  const res = http.post(
    "https://api.example.com/login",
    JSON.stringify({
      email: user.email,
      password: user.password,
    }),
    {
      headers: { "Content-Type": "application/json" },
    }
  );

  check(res, { "login success": (r) => r.status === 200 });
}
```

---

## Performance Metrics

### Example 10: Comprehensive Metrics Collection

```javascript
// metrics-collection.js
import http from "k6/http";
import { Trend, Rate, Counter, Gauge } from "k6/metrics";

const apiLatency = new Trend("api_latency", true);
const dbLatency = new Trend("db_latency", true);
const cacheHitRate = new Rate("cache_hits");
const totalRequests = new Counter("total_requests");
const concurrentUsers = new Gauge("concurrent_users");

export default function () {
  concurrentUsers.add(1);
  totalRequests.add(1);

  const res = http.get("https://api.example.com/data", {
    tags: { endpoint: "data" },
  });

  // Extract custom timing headers
  const dbTime = parseFloat(res.headers["X-DB-Time"] || 0);
  const cacheHit = res.headers["X-Cache-Hit"] === "true";

  apiLatency.add(res.timings.duration);
  dbLatency.add(dbTime);
  cacheHitRate.add(cacheHit);

  concurrentUsers.add(-1);
}
```

---

## Distributed Testing

### Example 11: k6 Cloud Distributed Test

```bash
# Configure k6 Cloud
export K6_CLOUD_TOKEN="your-token"

# Run distributed test
k6 cloud --vus 10000 --duration 30m load-test.js
```

### Example 12: Gatling Enterprise Distribution

```scala
// DistributedSimulation.scala
class DistributedSimulation extends Simulation {
  setUp(
    scn.inject(
      rampUsers(10000).during(30.minutes)
    )
  ).protocols(httpProtocol)
   .deploymentInfo(
     numberOfNodes = 10,
     region = "us-east-1"
   )
}
```

---

**Note**: These examples demonstrate production-ready load testing patterns. Adjust VU counts, durations, and thresholds based on your specific requirements.
