# Point-in-Time Recovery and Backups

## Overview

Neon provides automatic Point-in-Time Recovery (PITR) with up to 30 days of retention, enabling instant database restoration to any timestamp without manual backup management.

---

## PITR Fundamentals

### How PITR Works

Write-Ahead Logging: All database changes recorded to WAL
Continuous Backup: WAL continuously streamed to Neon storage
Instant Restore: Branch creation from any point in retention window
Copy-on-Write: Restore creates new branch without affecting original

### Retention Periods

Free Tier: 7 days of PITR history
Pro Tier: 7-30 days configurable
Enterprise: Custom retention periods

---

## Restore Operations

### Restore to Specific Timestamp

```typescript
async function restoreToPoint(timestamp: Date, branchName?: string) {
  const name = branchName || `restore-${timestamp.toISOString().replace(/[:.]/g, '-')}`

  const response = await fetch(
    `https://console.neon.tech/api/v2/projects/${process.env.NEON_PROJECT_ID}/branches`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${process.env.NEON_API_KEY}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        branch: {
          name,
          parent_id: 'main',
          parent_timestamp: timestamp.toISOString()
        }
      })
    }
  )

  if (!response.ok) {
    throw new Error(`PITR restore failed: ${response.statusText}`)
  }

  return response.json()
}
```

### Common Restore Scenarios

Restore to 1 Hour Ago:
```typescript
const oneHourAgo = new Date(Date.now() - 60 * 60 * 1000)
const restored = await restoreToPoint(oneHourAgo, 'restore-1-hour')
```

Restore to Yesterday:
```typescript
const yesterday = new Date()
yesterday.setDate(yesterday.getDate() - 1)
yesterday.setHours(0, 0, 0, 0)
const restored = await restoreToPoint(yesterday, 'restore-yesterday')
```

Restore to Specific Datetime:
```typescript
const specificTime = new Date('2024-01-15T14:30:00Z')
const restored = await restoreToPoint(specificTime, 'restore-jan15-1430')
```

---

## Restore Workflow

### Complete Restore Process

```typescript
interface RestoreResult {
  branchId: string
  branchName: string
  connectionString: string
  restoredTo: string
  createdAt: string
}

async function performRestore(
  targetTimestamp: Date,
  options: {
    branchName?: string
    sourceBranch?: string
    validateFirst?: boolean
  } = {}
): Promise<RestoreResult> {
  const {
    branchName = `restore-${Date.now()}`,
    sourceBranch = 'main',
    validateFirst = true
  } = options

  // Validate timestamp is within retention period
  if (validateFirst) {
    const retentionDays = 30
    const oldestAllowed = new Date(Date.now() - retentionDays * 24 * 60 * 60 * 1000)

    if (targetTimestamp < oldestAllowed) {
      throw new Error(`Timestamp ${targetTimestamp.toISOString()} is outside ${retentionDays}-day retention period`)
    }

    if (targetTimestamp > new Date()) {
      throw new Error('Cannot restore to future timestamp')
    }
  }

  // Create restore branch
  const response = await fetch(
    `https://console.neon.tech/api/v2/projects/${process.env.NEON_PROJECT_ID}/branches`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${process.env.NEON_API_KEY}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        branch: {
          name: branchName,
          parent_id: sourceBranch,
          parent_timestamp: targetTimestamp.toISOString()
        }
      })
    }
  )

  if (!response.ok) {
    const error = await response.text()
    throw new Error(`Restore failed: ${error}`)
  }

  const data = await response.json()

  // Get connection string for restored branch
  const endpointsResponse = await fetch(
    `https://console.neon.tech/api/v2/projects/${process.env.NEON_PROJECT_ID}/branches/${data.branch.id}/endpoints`,
    {
      headers: {
        'Authorization': `Bearer ${process.env.NEON_API_KEY}`
      }
    }
  )

  const endpointsData = await endpointsResponse.json()

  return {
    branchId: data.branch.id,
    branchName: data.branch.name,
    connectionString: endpointsData.endpoints[0]?.connection_uri,
    restoredTo: targetTimestamp.toISOString(),
    createdAt: new Date().toISOString()
  }
}
```

---

## Data Validation After Restore

### Verify Restored Data

```typescript
import { neon } from '@neondatabase/serverless'

async function validateRestore(
  originalConnectionString: string,
  restoredConnectionString: string,
  tableName: string
): Promise<boolean> {
  const originalDb = neon(originalConnectionString)
  const restoredDb = neon(restoredConnectionString)

  // Compare row counts
  const [originalCount] = await originalDb`SELECT COUNT(*) as count FROM ${tableName}`
  const [restoredCount] = await restoredDb`SELECT COUNT(*) as count FROM ${tableName}`

  console.log(`Original count: ${originalCount.count}, Restored count: ${restoredCount.count}`)

  // Compare checksums for critical tables
  const [originalChecksum] = await originalDb`
    SELECT md5(string_agg(id::text, '')) as checksum
    FROM ${tableName}
    ORDER BY id
  `
  const [restoredChecksum] = await restoredDb`
    SELECT md5(string_agg(id::text, '')) as checksum
    FROM ${tableName}
    ORDER BY id
  `

  const isValid = originalChecksum.checksum === restoredChecksum.checksum

  console.log(`Checksum validation: ${isValid ? 'PASS' : 'FAIL'}`)

  return isValid
}
```

### Restore Verification Checklist

```typescript
interface VerificationResult {
  tableChecks: Map<string, boolean>
  rowCounts: Map<string, number>
  overallStatus: 'success' | 'warning' | 'failure'
  issues: string[]
}

async function verifyRestore(connectionString: string): Promise<VerificationResult> {
  const sql = neon(connectionString)

  const criticalTables = ['users', 'orders', 'payments', 'products']
  const tableChecks = new Map<string, boolean>()
  const rowCounts = new Map<string, number>()
  const issues: string[] = []

  for (const table of criticalTables) {
    try {
      const [result] = await sql`
        SELECT COUNT(*) as count FROM ${sql(table)}
      `
      rowCounts.set(table, parseInt(result.count))
      tableChecks.set(table, true)
    } catch (error) {
      tableChecks.set(table, false)
      issues.push(`Table ${table}: ${error}`)
    }
  }

  const failedChecks = [...tableChecks.values()].filter(v => !v).length

  return {
    tableChecks,
    rowCounts,
    overallStatus: failedChecks === 0 ? 'success' : failedChecks < criticalTables.length / 2 ? 'warning' : 'failure',
    issues
  }
}
```

---

## Automated Backup Verification

### Scheduled PITR Testing

```typescript
interface BackupTestResult {
  testId: string
  timestamp: Date
  restoreSuccess: boolean
  validationSuccess: boolean
  duration: number
  cleanedUp: boolean
}

async function testPITRCapability(): Promise<BackupTestResult> {
  const testId = `pitr-test-${Date.now()}`
  const startTime = Date.now()
  let restoreSuccess = false
  let validationSuccess = false
  let cleanedUp = false

  try {
    // Test restore from 1 hour ago
    const testTimestamp = new Date(Date.now() - 60 * 60 * 1000)

    const restore = await performRestore(testTimestamp, {
      branchName: testId,
      validateFirst: true
    })

    restoreSuccess = true

    // Validate restored data
    const verification = await verifyRestore(restore.connectionString)
    validationSuccess = verification.overallStatus === 'success'

    // Cleanup test branch
    await fetch(
      `https://console.neon.tech/api/v2/projects/${process.env.NEON_PROJECT_ID}/branches/${restore.branchId}`,
      {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${process.env.NEON_API_KEY}`
        }
      }
    )
    cleanedUp = true

  } catch (error) {
    console.error('PITR test failed:', error)
  }

  return {
    testId,
    timestamp: new Date(),
    restoreSuccess,
    validationSuccess,
    duration: Date.now() - startTime,
    cleanedUp
  }
}
```

### GitHub Actions PITR Test

```yaml
name: PITR Verification

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM UTC
  workflow_dispatch:

jobs:
  test-pitr:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        run: npm ci

      - name: Run PITR Test
        env:
          NEON_API_KEY: ${{ secrets.NEON_API_KEY }}
          NEON_PROJECT_ID: ${{ secrets.NEON_PROJECT_ID }}
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: npm run test:pitr

      - name: Report Results
        if: failure()
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.create({
              owner: context.repo.owner,
              repo: context.repo.repo,
              title: 'PITR Verification Failed',
              body: 'Daily PITR verification test failed. Please investigate immediately.',
              labels: ['critical', 'infrastructure']
            })
```

---

## Disaster Recovery Procedures

### Emergency Restore Procedure

Step 1: Identify Recovery Point
- Determine exact timestamp before incident
- Account for timezone differences (use UTC)
- Consider data propagation delays

Step 2: Create Restore Branch
```typescript
const incidentTime = new Date('2024-01-15T14:30:00Z')
const safeTime = new Date(incidentTime.getTime() - 5 * 60 * 1000)  // 5 minutes before
const restored = await performRestore(safeTime, { branchName: 'emergency-restore' })
```

Step 3: Validate Restored Data
- Check critical table row counts
- Verify recent transactions
- Confirm data integrity

Step 4: Switch Production
- Update connection strings to restored branch
- Monitor for errors
- Keep original branch for investigation

Step 5: Post-Incident Cleanup
- Document timeline and decisions
- Archive investigation branch
- Update runbooks as needed

### Production Cutover

```typescript
async function cutoverToRestored(restoredBranchId: string) {
  // Get connection details for restored branch
  const endpoints = await fetch(
    `https://console.neon.tech/api/v2/projects/${process.env.NEON_PROJECT_ID}/branches/${restoredBranchId}/endpoints`,
    {
      headers: { 'Authorization': `Bearer ${process.env.NEON_API_KEY}` }
    }
  ).then(r => r.json())

  const newConnectionString = endpoints.endpoints[0]?.connection_uri

  // Update environment configuration
  // This varies by platform (Vercel, Railway, etc.)
  console.log('New connection string:', newConnectionString)
  console.log('Update DATABASE_URL_POOLED in your platform')

  return {
    newConnectionString,
    cutoverTime: new Date().toISOString(),
    instructions: [
      '1. Update DATABASE_URL_POOLED environment variable',
      '2. Trigger application redeployment',
      '3. Monitor for connection errors',
      '4. Verify data integrity in production'
    ]
  }
}
```

---

## Best Practices

### Retention Configuration

Align Retention with Compliance: Match retention period to regulatory requirements
Consider Recovery Objectives: Longer retention for lower RPO requirements
Balance Cost: Longer retention increases storage costs

### Testing Strategy

Regular PITR Tests: Weekly or daily automated restore tests
Full Recovery Drills: Quarterly complete disaster recovery exercises
Document Procedures: Maintain up-to-date runbooks

### Monitoring

Track PITR Health: Monitor WAL streaming status
Alert on Issues: Immediate notification of backup failures
Retention Warnings: Alert before data exits retention window

---

Version: 2.0.0
Last Updated: 2026-01-06
