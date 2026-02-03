# Database Branching Workflows

## Overview

Neon database branching creates instant copy-on-write clones of your database, enabling isolated development environments, preview deployments, and safe testing with production data.

---

## Core Concepts

### Copy-on-Write Architecture

Branch Creation: Instant creation with no data copying
Storage Efficiency: Branches share unchanged data with parent
Write Isolation: Changes in branch do not affect parent
Inheritance: New data in parent does not propagate to existing branches

### Branch Types

Main Branch: Primary production database
Development Branch: Long-lived branch for development work
Feature Branch: Short-lived branch for specific features
Preview Branch: Ephemeral branch for PR preview environments
Restore Branch: Branch created for point-in-time recovery

---

## Branch Management API

### NeonBranchManager Implementation

```typescript
class NeonBranchManager {
  private apiKey: string
  private projectId: string
  private baseUrl = 'https://console.neon.tech/api/v2'

  constructor(apiKey: string, projectId: string) {
    this.apiKey = apiKey
    this.projectId = projectId
  }

  private async request(path: string, options: RequestInit = {}) {
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
        ...options.headers
      }
    })
    if (!response.ok) throw new Error(`Neon API error: ${response.statusText}`)
    return response.json()
  }

  async createBranch(name: string, parentId: string = 'main') {
    return this.request(`/projects/${this.projectId}/branches`, {
      method: 'POST',
      body: JSON.stringify({
        branch: { name, parent_id: parentId }
      })
    })
  }

  async deleteBranch(branchId: string) {
    return this.request(`/projects/${this.projectId}/branches/${branchId}`, {
      method: 'DELETE'
    })
  }

  async listBranches() {
    return this.request(`/projects/${this.projectId}/branches`)
  }

  async getBranchConnectionString(branchId: string) {
    const endpoints = await this.request(
      `/projects/${this.projectId}/branches/${branchId}/endpoints`
    )
    return endpoints.endpoints[0]?.connection_uri
  }
}
```

---

## Preview Environment Pattern

### Per-PR Database Branches

```typescript
async function createPreviewEnvironment(prNumber: number) {
  const branchManager = new NeonBranchManager(
    process.env.NEON_API_KEY!,
    process.env.NEON_PROJECT_ID!
  )

  // Create branch from main with PR identifier
  const branch = await branchManager.createBranch(`pr-${prNumber}`, 'main')

  // Get connection string for the new branch
  const connectionString = await branchManager.getBranchConnectionString(branch.branch.id)

  return {
    branchId: branch.branch.id,
    branchName: branch.branch.name,
    connectionString,
    createdAt: new Date().toISOString()
  }
}

async function cleanupPreviewEnvironment(prNumber: number) {
  const branchManager = new NeonBranchManager(
    process.env.NEON_API_KEY!,
    process.env.NEON_PROJECT_ID!
  )

  // Find and delete the PR branch
  const { branches } = await branchManager.listBranches()
  const prBranch = branches.find(b => b.name === `pr-${prNumber}`)

  if (prBranch) {
    await branchManager.deleteBranch(prBranch.id)
  }
}
```

---

## GitHub Actions Integration

### Preview Environment Workflow

```yaml
name: Preview Environment

on:
  pull_request:
    types: [opened, synchronize, closed]

env:
  NEON_API_KEY: ${{ secrets.NEON_API_KEY }}
  NEON_PROJECT_ID: ${{ secrets.NEON_PROJECT_ID }}

jobs:
  create-preview:
    if: github.event.action != 'closed'
    runs-on: ubuntu-latest
    outputs:
      branch_id: ${{ steps.create-branch.outputs.branch_id }}
      database_url: ${{ steps.create-branch.outputs.database_url }}
    steps:
      - name: Create Neon Branch
        id: create-branch
        run: |
          RESPONSE=$(curl -s -X POST \
            -H "Authorization: Bearer $NEON_API_KEY" \
            -H "Content-Type: application/json" \
            -d '{"branch":{"name":"pr-${{ github.event.number }}"}}' \
            "https://console.neon.tech/api/v2/projects/$NEON_PROJECT_ID/branches")

          BRANCH_ID=$(echo $RESPONSE | jq -r '.branch.id')
          echo "branch_id=$BRANCH_ID" >> $GITHUB_OUTPUT

          # Get connection string
          ENDPOINTS=$(curl -s \
            -H "Authorization: Bearer $NEON_API_KEY" \
            "https://console.neon.tech/api/v2/projects/$NEON_PROJECT_ID/branches/$BRANCH_ID/endpoints")

          DATABASE_URL=$(echo $ENDPOINTS | jq -r '.endpoints[0].connection_uri')
          echo "database_url=$DATABASE_URL" >> $GITHUB_OUTPUT

      - name: Comment PR with Database URL
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `Preview database created: \`pr-${{ github.event.number }}\`\n\nBranch ID: \`${{ steps.create-branch.outputs.branch_id }}\``
            })

  deploy-preview:
    needs: create-preview
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Migrations
        env:
          DATABASE_URL: ${{ needs.create-preview.outputs.database_url }}
        run: |
          npm ci
          npm run db:migrate

      - name: Deploy Preview
        env:
          DATABASE_URL: ${{ needs.create-preview.outputs.database_url }}
        run: |
          # Deploy to preview environment (Vercel, Netlify, etc.)
          echo "Deploying with preview database..."

  cleanup-preview:
    if: github.event.action == 'closed'
    runs-on: ubuntu-latest
    steps:
      - name: Find Branch ID
        id: find-branch
        run: |
          BRANCHES=$(curl -s \
            -H "Authorization: Bearer $NEON_API_KEY" \
            "https://console.neon.tech/api/v2/projects/$NEON_PROJECT_ID/branches")

          BRANCH_ID=$(echo $BRANCHES | jq -r '.branches[] | select(.name == "pr-${{ github.event.number }}") | .id')
          echo "branch_id=$BRANCH_ID" >> $GITHUB_OUTPUT

      - name: Delete Neon Branch
        if: steps.find-branch.outputs.branch_id != ''
        run: |
          curl -X DELETE \
            -H "Authorization: Bearer $NEON_API_KEY" \
            "https://console.neon.tech/api/v2/projects/$NEON_PROJECT_ID/branches/${{ steps.find-branch.outputs.branch_id }}"
```

---

## Development Workflow Pattern

### Feature Branch Database

```typescript
interface FeatureBranchConfig {
  featureName: string
  baseBranch?: string
  autoCleanupDays?: number
}

async function createFeatureBranch(config: FeatureBranchConfig) {
  const { featureName, baseBranch = 'main', autoCleanupDays = 7 } = config

  const branchManager = new NeonBranchManager(
    process.env.NEON_API_KEY!,
    process.env.NEON_PROJECT_ID!
  )

  // Create branch with feature prefix
  const branchName = `feature-${featureName}-${Date.now()}`
  const branch = await branchManager.createBranch(branchName, baseBranch)

  // Store cleanup metadata
  const metadata = {
    branchId: branch.branch.id,
    branchName,
    featureName,
    createdAt: new Date().toISOString(),
    expiresAt: new Date(Date.now() + autoCleanupDays * 24 * 60 * 60 * 1000).toISOString()
  }

  return {
    ...metadata,
    connectionString: await branchManager.getBranchConnectionString(branch.branch.id)
  }
}

async function cleanupExpiredBranches() {
  const branchManager = new NeonBranchManager(
    process.env.NEON_API_KEY!,
    process.env.NEON_PROJECT_ID!
  )

  const { branches } = await branchManager.listBranches()
  const now = new Date()

  for (const branch of branches) {
    // Check if branch is a feature branch and expired
    if (branch.name.startsWith('feature-')) {
      const createdAt = new Date(branch.created_at)
      const ageInDays = (now.getTime() - createdAt.getTime()) / (24 * 60 * 60 * 1000)

      if (ageInDays > 7) {
        console.log(`Cleaning up expired branch: ${branch.name}`)
        await branchManager.deleteBranch(branch.id)
      }
    }
  }
}
```

---

## Branch Reset Pattern

### Reset Branch to Parent State

```typescript
async function resetBranchToParent(branchName: string) {
  const branchManager = new NeonBranchManager(
    process.env.NEON_API_KEY!,
    process.env.NEON_PROJECT_ID!
  )

  // Find the current branch
  const { branches } = await branchManager.listBranches()
  const currentBranch = branches.find(b => b.name === branchName)

  if (!currentBranch) {
    throw new Error(`Branch ${branchName} not found`)
  }

  const parentId = currentBranch.parent_id

  // Delete current branch
  await branchManager.deleteBranch(currentBranch.id)

  // Recreate with same name from parent
  const newBranch = await branchManager.createBranch(branchName, parentId)

  return {
    branchId: newBranch.branch.id,
    connectionString: await branchManager.getBranchConnectionString(newBranch.branch.id),
    resetAt: new Date().toISOString()
  }
}
```

---

## Best Practices

### Branch Naming Conventions

Preview Branches: pr-{number} for pull request previews
Feature Branches: feature-{name}-{timestamp} for development
Staging Branches: staging or staging-{version} for staging environments
Restore Branches: restore-{timestamp} for point-in-time recovery

### Lifecycle Management

Automatic Cleanup: Configure scheduled cleanup for expired branches
Branch Limits: Monitor branch count against project limits
Connection Management: Use pooled connections for branch databases
Cost Awareness: Branches consume compute when active

### Security Considerations

Sensitive Data: Consider data masking for non-production branches
Access Control: Limit API key permissions for branch operations
Audit Trail: Log branch creation and deletion operations
Credential Rotation: Rotate branch credentials on schedule

---

Version: 2.0.0
Last Updated: 2026-01-06
