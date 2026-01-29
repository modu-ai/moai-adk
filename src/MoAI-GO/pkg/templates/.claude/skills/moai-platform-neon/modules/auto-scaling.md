# Auto-Scaling and Compute Management

## Overview

Neon auto-scaling automatically adjusts compute resources based on workload demand, with scale-to-zero capability for cost optimization during idle periods.

---

## Compute Unit Specifications

### Available Compute Sizes

0.25 CU Specifications:
- vCPU: 0.25
- RAM: 1 GB
- Use Case: Development, testing, low-traffic applications

0.5 CU Specifications:
- vCPU: 0.5
- RAM: 2 GB
- Use Case: Light production workloads, staging environments

1 CU Specifications:
- vCPU: 1
- RAM: 4 GB
- Use Case: Standard production applications

2 CU Specifications:
- vCPU: 2
- RAM: 8 GB
- Use Case: Medium workloads, moderate traffic

4 CU Specifications:
- vCPU: 4
- RAM: 16 GB
- Use Case: Heavy workloads, high traffic applications

8 CU Specifications:
- vCPU: 8
- RAM: 32 GB
- Use Case: High-performance requirements, data processing

---

## Auto-Scaling Configuration

### Configuration via API

```typescript
interface AutoScalingConfig {
  minCu: number      // Minimum compute units (0.25 for scale-to-zero)
  maxCu: number      // Maximum compute units
  suspendTimeout: number  // Seconds before suspension (0 to disable)
}

async function configureAutoScaling(endpointId: string, config: AutoScalingConfig) {
  const response = await fetch(
    `https://console.neon.tech/api/v2/projects/${process.env.NEON_PROJECT_ID}/endpoints/${endpointId}`,
    {
      method: 'PATCH',
      headers: {
        'Authorization': `Bearer ${process.env.NEON_API_KEY}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        endpoint: {
          autoscaling_limit_min_cu: config.minCu,
          autoscaling_limit_max_cu: config.maxCu,
          suspend_timeout_seconds: config.suspendTimeout
        }
      })
    }
  )

  if (!response.ok) {
    throw new Error(`Failed to configure auto-scaling: ${response.statusText}`)
  }

  return response.json()
}
```

### Environment-Specific Configurations

Development Configuration:
```typescript
const devConfig: AutoScalingConfig = {
  minCu: 0.25,        // Scale to zero when idle
  maxCu: 0.5,         // Limited compute for development
  suspendTimeout: 300  // Suspend after 5 minutes idle
}
```

Staging Configuration:
```typescript
const stagingConfig: AutoScalingConfig = {
  minCu: 0.25,        // Scale to zero during off-hours
  maxCu: 1,           // Moderate compute for testing
  suspendTimeout: 600  // Suspend after 10 minutes idle
}
```

Production Configuration:
```typescript
const productionConfig: AutoScalingConfig = {
  minCu: 0.5,         // Always-on minimum for faster response
  maxCu: 4,           // Scale up for peak traffic
  suspendTimeout: 3600 // Suspend after 1 hour idle (or 0 to disable)
}
```

High-Traffic Production Configuration:
```typescript
const highTrafficConfig: AutoScalingConfig = {
  minCu: 1,           // Higher baseline for consistent performance
  maxCu: 8,           // Maximum scale for peak loads
  suspendTimeout: 0   // Never suspend (always-on)
}
```

---

## Scale-to-Zero Behavior

### How Scale-to-Zero Works

Idle Detection: Neon monitors connection activity and query execution
Suspension: After suspend_timeout seconds of inactivity, compute suspends
Wake-Up: First connection request triggers automatic wake-up
Cold Start: Typical wake-up time is 500ms to 2 seconds

### Cold Start Optimization

Connection Pooling Impact:
- Use pooled connections to reduce cold start frequency
- Pooler maintains connection state during suspension
- First query after wake-up experiences latency

Warming Strategies:
```typescript
// Health check endpoint to keep database warm
async function warmDatabase() {
  const sql = neon(process.env.DATABASE_URL_POOLED!)

  // Simple query to prevent suspension
  await sql`SELECT 1`
}

// Schedule periodic warming (every 4 minutes for 5-minute timeout)
setInterval(warmDatabase, 4 * 60 * 1000)
```

### Cost Implications

Scale-to-Zero Benefits:
- Zero compute charges during idle periods
- Ideal for development and low-traffic applications
- Automatic cost optimization without manual intervention

Trade-offs:
- Cold start latency on first request after suspension
- Not suitable for latency-sensitive applications
- Consider always-on minimum for production workloads

---

## Endpoint Management

### Get Current Endpoint Configuration

```typescript
async function getEndpointConfig(endpointId: string) {
  const response = await fetch(
    `https://console.neon.tech/api/v2/projects/${process.env.NEON_PROJECT_ID}/endpoints/${endpointId}`,
    {
      headers: {
        'Authorization': `Bearer ${process.env.NEON_API_KEY}`
      }
    }
  )

  return response.json()
}
```

### List All Endpoints

```typescript
async function listEndpoints() {
  const response = await fetch(
    `https://console.neon.tech/api/v2/projects/${process.env.NEON_PROJECT_ID}/endpoints`,
    {
      headers: {
        'Authorization': `Bearer ${process.env.NEON_API_KEY}`
      }
    }
  )

  return response.json()
}
```

### Monitor Endpoint Status

```typescript
interface EndpointStatus {
  id: string
  state: 'active' | 'idle' | 'suspended'
  currentCu: number
  lastActiveAt: string
}

async function getEndpointStatus(endpointId: string): Promise<EndpointStatus> {
  const { endpoint } = await getEndpointConfig(endpointId)

  return {
    id: endpoint.id,
    state: endpoint.current_state,
    currentCu: endpoint.autoscaling_limit_min_cu,
    lastActiveAt: endpoint.last_active
  }
}
```

---

## Cost Optimization Strategies

### Development Environments

Strategy: Aggressive scale-to-zero with low maximum compute

```typescript
// Minimize costs for development databases
await configureAutoScaling(devEndpointId, {
  minCu: 0.25,
  maxCu: 0.5,
  suspendTimeout: 180  // 3 minutes - quick suspension
})
```

### Staging Environments

Strategy: Balance between cost and performance for testing

```typescript
// Cost-effective staging with reasonable performance
await configureAutoScaling(stagingEndpointId, {
  minCu: 0.25,
  maxCu: 2,
  suspendTimeout: 600  // 10 minutes
})
```

### Production Environments

Strategy: Prioritize performance with cost awareness

```typescript
// Production with always-on minimum
await configureAutoScaling(prodEndpointId, {
  minCu: 0.5,         // Avoid cold starts
  maxCu: 4,           // Handle traffic spikes
  suspendTimeout: 0   // Never suspend
})
```

### Off-Hours Optimization

```typescript
// Reduce compute during off-peak hours
async function adjustForOffHours(endpointId: string, isOffHours: boolean) {
  const config = isOffHours
    ? { minCu: 0.25, maxCu: 1, suspendTimeout: 300 }
    : { minCu: 0.5, maxCu: 4, suspendTimeout: 0 }

  await configureAutoScaling(endpointId, config)
}

// Schedule-based adjustment
const hour = new Date().getUTCHours()
const isOffHours = hour >= 2 && hour < 8  // 2 AM - 8 AM UTC
await adjustForOffHours(productionEndpointId, isOffHours)
```

---

## Monitoring and Alerts

### Compute Usage Tracking

```typescript
interface ComputeMetrics {
  currentCu: number
  avgCu: number
  peakCu: number
  suspendedMinutes: number
  activeMinutes: number
}

async function getComputeMetrics(endpointId: string): Promise<ComputeMetrics> {
  // Fetch from Neon console API or monitoring integration
  const { endpoint } = await getEndpointConfig(endpointId)

  return {
    currentCu: endpoint.current_state === 'active' ? endpoint.autoscaling_limit_min_cu : 0,
    avgCu: 0,  // Calculate from historical data
    peakCu: endpoint.autoscaling_limit_max_cu,
    suspendedMinutes: 0,  // Calculate from suspension logs
    activeMinutes: 0  // Calculate from activity logs
  }
}
```

### Alert Thresholds

High Compute Alert: Notify when consistently at max CU
Frequent Suspensions: Alert if cold starts affecting performance
Cost Threshold: Alert when monthly compute exceeds budget

---

## Best Practices

### Configuration Guidelines

Development: Use 0.25-0.5 CU range with short suspension timeout
Staging: Use 0.25-2 CU range with moderate suspension timeout
Production: Use 0.5-4 CU range, consider disabling suspension for critical apps
High-Traffic: Use 1-8 CU range with suspension disabled

### Performance Considerations

Connection Pooling: Always use pooled connections with auto-scaling
Query Optimization: Optimize queries to reduce compute time
Indexing: Proper indexes reduce CPU usage and allow smaller compute
Caching: Implement application-level caching to reduce database load

### Cost Management

Monitor Usage: Track compute hours and identify optimization opportunities
Right-Size: Adjust max CU based on actual peak usage
Schedule Scaling: Reduce compute during known low-traffic periods
Branch Cleanup: Delete unused branches to avoid dormant compute costs

---

Version: 2.0.0
Last Updated: 2026-01-06
